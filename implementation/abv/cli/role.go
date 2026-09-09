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
	fmt.Fprintf(&rendered, "internal projection: role\nid  %s\nrevision  %d\npermissions  %s\n", role.ID, role.Revision, strings.Join(role.Permissions, ","))
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; publication does not grant business access"); err != nil {
		return err
	}
	if _, err := out.Write(rendered.Bytes()); err != nil {
		return err
	}
	return nil
}
