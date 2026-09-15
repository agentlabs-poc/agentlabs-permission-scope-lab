package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// renderResolved prints the canonical document a client consumes. It is JSON
// rather than a projection because this one is a contract: what an application
// receives over the wire is what prints here.
func renderResolved(out, diag io.Writer, resolved domain.ResolvedAuthority) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	encoder := json.NewEncoder(&rendered)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resolved); err != nil {
		return err
	}
	_, err := out.Write(rendered.Bytes())
	return err
}
