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
	case "role":
		switch positional[0] {
		case "publish":
			revision, _ := strconv.ParseInt(flags["--revision"], 10, 64)
			proposed := domain.RoleContent{ID: positional[1], Name: flags["--name"], Revision: revision, Permissions: strings.Split(flags["--permissions"], ",")}
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
