package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"
)

func inspect(ctx context.Context, api application.API, area domain.Area, kind, id string, out io.Writer) error {
	record, err := api.Inspect(ctx, area, kind, id)
	if err != nil {
		return err
	}
	if len(record.CanonicalJSON) > 0 {
		_, err = fmt.Fprintf(out, "%s\n", record.CanonicalJSON)
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: %s\n", kind)
	table := tabwriter.NewWriter(&rendered, 0, 4, 2, ' ', 0)
	for _, row := range record.Rows {
		for column, value := range row {
			if column > 0 {
				_, _ = io.WriteString(table, "\t")
			}
			_, _ = io.WriteString(table, value)
		}
		_, _ = io.WriteString(table, "\n")
	}
	if err = table.Flush(); err != nil {
		return err
	}
	_, err = io.Copy(out, &rendered)
	return err
}
