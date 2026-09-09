package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func publishGrantRevision(ctx context.Context, api application.API, area domain.Area, fixture, source string, raw []byte, out, diag io.Writer) error {
	publication, ok := api.(application.GrantRevisionAPI)
	if !ok || nilCapability(publication) {
		return domain.ErrUnsupported
	}
	grant, err := publication.PublishGrantRevision(ctx, area, domain.FixtureContext{Name: fixture}, source, raw)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(grant)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; support assignment is transient publication evidence"); err != nil {
		return err
	}
	_, err = out.Write(append(encoded, '\n'))
	return err
}
