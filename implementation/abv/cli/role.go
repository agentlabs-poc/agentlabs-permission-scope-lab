package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

func publishRole(ctx context.Context, api application.API, area domain.Area, fixture string, proposed domain.RoleContent, out, diag io.Writer) error {
	roleAPI, ok := api.(application.RoleAPI)
	if !ok || nilCapability(roleAPI) {
		return domain.ErrUnsupported
	}
	role, err := roleAPI.PublishRole(ctx, area, domain.FixtureContext{Name: fixture}, proposed)
	if err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprint(&rendered, roleLines(role))
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; publication does not grant business access"); err != nil {
		return err
	}
	if _, err := out.Write(rendered.Bytes()); err != nil {
		return err
	}
	return nil
}

// roleLines renders one role revision. The revision prints as the int64 the
// contract carries, never the zero-padded slot: padding is storage's business
// and a reader never sees it.
func roleLines(role domain.RoleContent) string {
	return fmt.Sprintf("internal projection: role\nid  %s\nname  %s\nmanaged  %s\nrevision  %d\npermissions  %s\n",
		role.ID, role.Name, role.Managed, role.Revision, strings.Join(role.Permissions, ","))
}

func renderRole(out, diag io.Writer, role domain.RoleContent) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	_, err := fmt.Fprint(out, roleLines(role))
	return err
}

// renderRolePage prints the page and the two numbers that make it navigable.
// Total follows the selector the caller used, so it counts rows under
// AllRevisions and roles under LatestRevision.
func renderRolePage(out, diag io.Writer, page domain.RolePage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: roles\ncount  %d\ntotal  %d\ngeneration  %d\n",
		len(page.Roles), page.Total, page.Generation)
	for _, role := range page.Roles {
		fmt.Fprintf(&rendered, "%s  rev=%d  %-11s  %s  %s\n",
			role.ID, role.Revision, role.Managed, role.Name, strings.Join(role.Permissions, ","))
	}
	_, err := out.Write(rendered.Bytes())
	return err
}
