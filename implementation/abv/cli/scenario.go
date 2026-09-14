package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

const maxAssignmentBytes = 1 << 20

func dispatch(ctx context.Context, command string, positional []string, flags map[string]string, in io.Reader, out, diag io.Writer, api application.API, area domain.Area) int {
	switch command {
	case "inspect":
		if err := inspect(ctx, api, area, positional[0], positional[1], out); err != nil {
			return report(diag, err)
		}
		return 0
	case "check", "assign":
		raw, err := readInput(flags["--file"], in)
		if err != nil {
			return report(diag, err)
		}
		if command == "check" {
			err = check(ctx, api, area, raw, out)
		} else {
			err = assign(ctx, api, area, flags["--fixture-context"], raw, out, diag)
		}
		if err != nil {
			return report(diag, err)
		}
		return 0
	case "grant":
		if positional[0] == "publish" {
			raw, err := readInput(flags["--file"], in)
			if err != nil {
				return report(diag, err)
			}
			if err = publishGrantRevision(ctx, api, area, flags["--fixture-context"], flags["--support-assignment"], raw, out, diag); err != nil {
				return report(diag, err)
			}
			return 0
		}
		if err := grantStatus(ctx, api, area, flags["--fixture-context"], positional[0], positional[1], out); err != nil {
			return report(diag, err)
		}
		return 0
	case "assignment":
		if err := assignmentStatus(ctx, api, area, flags["--fixture-context"], positional[0], positional[1], out); err != nil {
			return report(diag, err)
		}
		return 0
	case "team":
		teamAPI, ok := api.(application.TeamAPI)
		if !ok || nilCapability(teamAPI) {
			return report(diag, domain.ErrUnsupported)
		}
		fixture := domain.FixtureContext{Name: flags["--fixture-context"]}
		switch positional[0] {
		case "get":
			team, err := teamAPI.GetTeam(ctx, area, fixture, positional[1])
			if err != nil {
				return report(diag, err)
			}
			if err := renderTeam(out, diag, team); err != nil {
				return 4
			}
		case "list":
			filter := domain.TeamFilter{Name: flags["--name"]}
			if has(flags, "--roots") {
				root := ""
				filter.ParentID = &root
			} else if has(flags, "--parent") {
				parent := flags["--parent"]
				filter.ParentID = &parent
			}
			if code := teamBounds(flags, &filter.Offset, &filter.Limit, diag); code != 0 {
				return code
			}
			page, err := teamAPI.ListTeams(ctx, area, fixture, filter)
			if err != nil {
				return report(diag, err)
			}
			if err := renderTeamPage(out, diag, page); err != nil {
				return 4
			}
		case "create":
			team, err := teamAPI.CreateTeam(ctx, area, fixture, flags["--name"], flags["--parent"])
			if err != nil {
				return report(diag, err)
			}
			if err := renderTeam(out, diag, team); err != nil {
				return 4
			}
		case "reparent":
			parent := flags["--parent"]
			if has(flags, "--roots") {
				parent = ""
			}
			team, err := teamAPI.SetTeamParent(ctx, area, fixture, positional[1], parent)
			if err != nil {
				return report(diag, err)
			}
			if err := renderTeam(out, diag, team); err != nil {
				return 4
			}
		case "delete":
			if err := teamAPI.DeleteTeam(ctx, area, fixture, positional[1]); err != nil {
				return report(diag, err)
			}
			if err := renderRemoved(out, diag, "team", positional[1]); err != nil {
				return 4
			}
		case "add-member":
			if err := teamAPI.AddMember(ctx, area, fixture, flags["--id"], flags["--human"]); err != nil {
				return report(diag, err)
			}
			if err := renderMembershipChange(out, diag, "added", flags["--id"], flags["--human"]); err != nil {
				return 4
			}
		case "remove-member":
			if err := teamAPI.RemoveMember(ctx, area, fixture, flags["--id"], flags["--human"]); err != nil {
				return report(diag, err)
			}
			if err := renderMembershipChange(out, diag, "removed", flags["--id"], flags["--human"]); err != nil {
				return 4
			}
		case "members":
			filter := domain.MemberFilter{TeamID: flags["--id"], HumanID: flags["--human"]}
			if code := teamBounds(flags, &filter.Offset, &filter.Limit, diag); code != 0 {
				return code
			}
			page, err := teamAPI.ListMembers(ctx, area, fixture, filter)
			if err != nil {
				return report(diag, err)
			}
			if err := renderMemberPage(out, diag, page); err != nil {
				return 4
			}
		}
		return 0
	case "role":
		switch positional[0] {
		case "publish":
			revision, _ := strconv.ParseInt(flags["--revision"], 10, 64)
			proposed := domain.RoleContent{ID: flags["--id"], Name: flags["--name"], Revision: revision, Permissions: strings.Split(flags["--permissions"], ",")}
			if has(flags, "--application") {
				roleAPI, ok := api.(application.RoleAPI)
				if !ok || nilCapability(roleAPI) {
					return report(diag, domain.ErrUnsupported)
				}
				app, appErr := domain.NewApplication(area.ApplicationID())
				if appErr != nil {
					return report(diag, appErr)
				}
				role, err := roleAPI.PublishApplicationRole(ctx, app, domain.FixtureContext{Name: flags["--fixture-context"]}, proposed)
				if err != nil {
					return report(diag, err)
				}
				if err := renderRole(out, diag, role); err != nil {
					return 4
				}
				break
			}
			if err := publishRole(ctx, api, area, flags["--fixture-context"], proposed, out, diag); err != nil {
				return report(diag, err)
			}
		case "get":
			roleAPI, ok := api.(application.RoleAPI)
			if !ok || nilCapability(roleAPI) {
				return report(diag, domain.ErrUnsupported)
			}
			revision, _ := strconv.ParseInt(flags["--revision"], 10, 64)
			role, err := roleAPI.GetRole(ctx, area, domain.FixtureContext{Name: flags["--fixture-context"]}, positional[1], revision)
			if err != nil {
				return report(diag, err)
			}
			if err := renderRole(out, diag, role); err != nil {
				return 4
			}
		case "list":
			roleAPI, ok := api.(application.RoleAPI)
			if !ok || nilCapability(roleAPI) {
				return report(diag, domain.ErrUnsupported)
			}
			filter := domain.RoleFilter{ID: flags["--id"], Name: flags["--name"]}
			if has(flags, "--managed") {
				managed := domain.TenantManaged
				if flags["--managed"] == "application" {
					managed = domain.ApplicationManaged
				}
				filter.Managed = &managed
			}
			if has(flags, "--latest") {
				filter.Revisions = domain.LatestRevision
			}
			for flag, target := range map[string]*int{"--offset": &filter.Offset, "--limit": &filter.Limit} {
				if has(flags, flag) {
					value, convErr := strconv.Atoi(flags[flag])
					if convErr != nil {
						return report(diag, domain.ErrMalformed)
					}
					*target = value
				}
			}
			page, err := roleAPI.ListRoles(ctx, area, domain.FixtureContext{Name: flags["--fixture-context"]}, filter)
			if err != nil {
				return report(diag, err)
			}
			if err := renderRolePage(out, diag, page); err != nil {
				return 4
			}
		}
		return 0
	}
	return report(diag, domain.ErrUnsupported)
}

func readInput(path string, in io.Reader) (raw []byte, err error) {
	reader := in
	var file *os.File
	if path != "-" {
		file, err = os.Open(path)
		if err != nil {
			return nil, errors.Join(domain.ErrUnavailable, err)
		}
		defer func() { err = errors.Join(err, file.Close()) }()
		reader = file
	}
	raw, err = io.ReadAll(io.LimitReader(reader, maxAssignmentBytes+1))
	if err != nil {
		return nil, errors.Join(domain.ErrUnavailable, err)
	}
	if len(raw) > maxAssignmentBytes {
		return nil, domain.ErrMalformed
	}
	return raw, nil
}
