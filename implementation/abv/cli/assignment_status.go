package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"encoding/json"
	"io"
)

func assignmentStatus(ctx context.Context, api application.API, area domain.Area, fixture, verb, id string, out io.Writer) error {
	statusAPI, ok := api.(application.AssignmentStatusAPI)
	if !ok || nilCapability(statusAPI) {
		return domain.ErrUnsupported
	}
	status := "enabled"
	if verb == "disable" {
		status = "disabled"
	}
	assignment, err := statusAPI.SetAssignmentStatus(ctx, area, domain.FixtureContext{Name: fixture}, id, status)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(assignment)
	if err != nil {
		return err
	}
	_, err = out.Write(append(raw, '\n'))
	return err
}
