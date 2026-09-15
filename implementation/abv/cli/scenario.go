package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"errors"
	"fmt"
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
	case "owners":
		ownerAPI, ok := api.(application.OwnerAPI)
		if !ok || nilCapability(ownerAPI) {
			return report(diag, domain.ErrUnsupported)
		}
		fixture := domain.FixtureContext{Name: flags["--fixture-context"]}
		switch positional[0] {
		case "add", "remove":
			act, verb := ownerAPI.AddOwner, "owns"
			if positional[0] == "remove" {
				act, verb = ownerAPI.RemoveOwner, "no longer owns"
			}
			if err := act(ctx, area, fixture, flags["--team"], flags["--human"]); err != nil {
				return report(diag, err)
			}
			if _, err := fmt.Fprintf(out, "internal projection: ownership\n%s %s %s\n", flags["--human"], verb, flags["--team"]); err != nil {
				return 4
			}
		case "list":
			filter := domain.OwnerFilter{TeamID: flags["--team"], HumanID: flags["--human"]}
			if code := teamBounds(flags, &filter.Offset, &filter.Limit, diag); code != 0 {
				return code
			}
			page, err := ownerAPI.ListOwners(ctx, area, fixture, filter)
			if err != nil {
				return report(diag, err)
			}
			if err := renderOwnerPage(out, diag, page); err != nil {
				return 4
			}
		}
		return 0
	case "assignments":
		assignmentAPI, ok := api.(application.AssignmentAPI)
		if !ok || nilCapability(assignmentAPI) {
			return report(diag, domain.ErrUnsupported)
		}
		fixture := domain.FixtureContext{Name: flags["--fixture-context"]}
		switch positional[0] {
		case "get":
			found, err := assignmentAPI.GetAssignment(ctx, area, fixture, positional[1])
			if err != nil {
				return report(diag, err)
			}
			if err := renderAssignmentRecord(out, diag, found); err != nil {
				return 4
			}
		case "list":
			filter := domain.AssignmentFilter{GrantID: flags["--grant"], Status: flags["--status"]}
			if !empty(flags["--recipient"]) {
				recipient := domain.Recipient{Type: flags["--recipient-type"], ID: flags["--recipient"]}
				filter.Recipient = &recipient
			}
			if code := teamBounds(flags, &filter.Offset, &filter.Limit, diag); code != 0 {
				return code
			}
			page, err := assignmentAPI.ListAssignments(ctx, area, fixture, filter)
			if err != nil {
				return report(diag, err)
			}
			if err := renderAssignmentPage(out, diag, page); err != nil {
				return 4
			}
		case "delete":
			if err := assignmentAPI.DeleteAssignment(ctx, area, fixture, positional[1]); err != nil {
				return report(diag, err)
			}
			if _, err := fmt.Fprintf(out, "internal projection: assignment\ndeleted  %s\n", positional[1]); err != nil {
				return 4
			}
		case "upgrade":
			after, err := assignmentAPI.UpgradeAssignment(ctx, area, fixture, positional[1])
			if err != nil {
				return report(diag, err)
			}
			if err := renderAssignmentRecord(out, diag, after); err != nil {
				return 4
			}
		}
		return 0
	case "resolve":
		authorityAPI, ok := api.(application.AuthorityAPI)
		if !ok || nilCapability(authorityAPI) {
			return report(diag, domain.ErrUnsupported)
		}
		human := flags["--human"]
		// Without --client the caller is the subject: a human asking about their
		// own authority. With it the caller is an application's own credential,
		// which is how an enforcing client asks about somebody else.
		actor := domain.Actor{Type: "user", ID: human}
		if !empty(flags["--client"]) {
			actor = domain.Actor{Type: "service_account", ID: flags["--client"]}
		}
		identity := domain.Identity{Version: "1", Actor: actor, HumanID: human}
		opts := domain.ResolveOptions{OmitSource: has(flags, "--no-source")}
		if !empty(flags["--permissions"]) {
			opts.Permissions = strings.Split(flags["--permissions"], ",")
		}
		resolved, err := authorityAPI.ResolveAuthority(ctx, area, domain.FixtureContext{Name: flags["--fixture-context"]}, identity, opts)
		if err != nil {
			return report(diag, err)
		}
		if err := renderResolved(out, diag, resolved); err != nil {
			return 4
		}
		return 0
	case "root":
		rootAPI, ok := api.(application.RootAPI)
		if !ok || nilCapability(rootAPI) {
			return report(diag, domain.ErrUnsupported)
		}
		fixture := domain.FixtureContext{Name: flags["--fixture-context"]}
		establish := rootAPI.EstablishRoot
		if positional[0] == "establish-auth" {
			establish = rootAPI.EstablishAuthRoot
		}
		grant, content, err := establish(ctx, area, fixture, flags["--team"])
		if err != nil {
			return report(diag, err)
		}
		if err := renderGrant(out, diag, grant, content); err != nil {
			return 4
		}
		return 0
	case "grants":
		grantAPI, ok := api.(application.GrantAPI)
		if !ok || nilCapability(grantAPI) {
			return report(diag, domain.ErrUnsupported)
		}
		fixture := domain.FixtureContext{Name: flags["--fixture-context"]}
		switch positional[0] {
		case "get":
			var revision int64
			if has(flags, "--revision") {
				parsed, err := strconv.ParseInt(flags["--revision"], 10, 64)
				if err != nil {
					return report(diag, domain.ErrMalformed)
				}
				revision = parsed
			}
			grant, content, err := grantAPI.GetGrant(ctx, area, fixture, positional[1], revision)
			if err != nil {
				return report(diag, err)
			}
			if err := renderGrant(out, diag, grant, content); err != nil {
				return 4
			}
		case "list":
			filter := domain.GrantFilter{Status: flags["--status"]}
			if has(flags, "--roots") {
				root := true
				filter.Root = &root
			} else if has(flags, "--children") {
				child := false
				filter.Root = &child
			}
			if code := teamBounds(flags, &filter.Offset, &filter.Limit, diag); code != 0 {
				return code
			}
			page, err := grantAPI.ListGrants(ctx, area, fixture, filter)
			if err != nil {
				return report(diag, err)
			}
			if err := renderGrantPage(out, diag, page); err != nil {
				return 4
			}
		case "revisions":
			offset, limit := 0, 0
			if code := teamBounds(flags, &offset, &limit, diag); code != 0 {
				return code
			}
			page, err := grantAPI.ListGrantRevisions(ctx, area, fixture, positional[1], offset, limit)
			if err != nil {
				return report(diag, err)
			}
			if err := renderRevisionPage(out, diag, positional[1], page); err != nil {
				return 4
			}
		case "create":
			scope, err := grantScope(flags["--scope"])
			if err != nil {
				return report(diag, err)
			}
			proposed := domain.GrantContent{Scope: scope}
			if flags["--permissions"] != "" {
				proposed.Permissions = strings.Split(flags["--permissions"], ",")
			} else {
				proposed.RoleID = flags["--role"]
				parsed, parseErr := strconv.ParseInt(flags["--role-revision"], 10, 64)
				if parseErr != nil {
					return report(diag, domain.ErrMalformed)
				}
				proposed.RoleRevision = parsed
			}
			grant, content, err := grantAPI.CreateGrant(ctx, area, fixture, flags["--parent"], proposed)
			if err != nil {
				return report(diag, err)
			}
			if err := renderGrant(out, diag, grant, content); err != nil {
				return 4
			}
		case "delete":
			if err := grantAPI.DeleteGrant(ctx, area, fixture, positional[1]); err != nil {
				return report(diag, err)
			}
			if _, err := fmt.Fprintf(out, "internal projection: grant\ndeleted  %s\n", positional[1]); err != nil {
				return 4
			}
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
