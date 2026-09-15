package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"fmt"
	"io"
)

// renderOwnerPage prints owners as the pairs they are. There is no owner id to
// print, because an ownership has none — the pair is the record.
func renderOwnerPage(out, diag io.Writer, page domain.OwnerPage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: owners\ncount  %d\ntotal  %d\n", len(page.Owners), page.Total)
	for _, o := range page.Owners {
		fmt.Fprintf(&rendered, "team=%-14s human=%s\n", o.TeamID, o.HumanID)
	}
	_, err := out.Write(rendered.Bytes())
	return err
}
