package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func check(ctx context.Context, api application.API, area domain.Area, raw []byte, out io.Writer) error {
	diagnostic, err := api.CheckAssignment(ctx, area, raw)
	if err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "READ-ONLY DIAGNOSIS: %s\nThis does not authorize a later write.\n", diagnostic.Summary)
	if diagnostic.Route != nil {
		_, _ = fmt.Fprintln(&rendered, "internal projection: proposed route")
		table := tabwriter.NewWriter(&rendered, 0, 4, 2, ' ', 0)
		_, _ = fmt.Fprintln(table, "field\tvalue\tsource")
		_, _ = fmt.Fprintf(table, "grant_id\t%s\t\n", diagnostic.Route.GrantID)
		_, _ = fmt.Fprintf(table, "permissions\t%s\t\n", strings.Join(diagnostic.Route.Permissions, ","))
		for _, predicate := range diagnostic.Route.Predicates {
			_, _ = fmt.Fprintf(table, "scope.%s\t%s\t%s\n", predicate.Key, predicate.Value, predicate.SourceGrantID)
		}
		_, _ = fmt.Fprintf(table, "assignment_ids\t%s\t\n", strings.Join(diagnostic.Route.AssignmentIDs, ","))
		if err = table.Flush(); err != nil {
			return err
		}
	}
	_, err = io.Copy(out, &rendered)
	return err
}
