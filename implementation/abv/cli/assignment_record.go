package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"fmt"
	"io"
)

// assignmentLines renders one assignment. The binding is printed first and the
// id after it, because the binding is what identifies the record — the id is a
// handle, and printing it first would suggest otherwise.
func assignmentLines(a domain.Assignment) string {
	return fmt.Sprintf("internal projection: assignment\ngrant  %s\nrecipient  %s %s\nid  %s\nadopted revision  %d\nstatus  %s\n",
		a.GrantID, a.Recipient.Type, a.Recipient.ID, a.ID, a.GrantRevision, a.Status)
}

func renderAssignmentRecord(out, diag io.Writer, a domain.Assignment) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	_, err := fmt.Fprint(out, assignmentLines(a))
	return err
}

func renderAssignmentPage(out, diag io.Writer, page domain.AssignmentPage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: assignments\ncount  %d\ntotal  %d\n", len(page.Assignments), page.Total)
	for _, a := range page.Assignments {
		fmt.Fprintf(&rendered, "%-14s %-6s %-14s rev %-4d %-9s %s\n",
			a.GrantID, a.Recipient.Type, a.Recipient.ID, a.GrantRevision, a.Status, a.ID)
	}
	_, err := out.Write(rendered.Bytes())
	return err
}
