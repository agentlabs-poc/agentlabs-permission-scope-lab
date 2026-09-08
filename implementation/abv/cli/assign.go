package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"fmt"
	"io"
)

func assign(ctx context.Context, api application.API, area domain.Area, fixture string, raw []byte, out, diag io.Writer) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixture-context is not authenticated identity; both authority gates are rechecked"); err != nil {
		return err
	}
	receipt, err := api.Assign(ctx, area, domain.FixtureContext{Name: fixture}, raw)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "assignment %s created\n", receipt.AssignmentID)
	return err
}
