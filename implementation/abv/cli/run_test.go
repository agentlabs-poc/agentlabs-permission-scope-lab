package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

// Parsing must fail before any adapter call. Nil adapters turn an accidental dispatch into a failure.
func TestInvalidCommandNeverDispatches(t *testing.T) {
	cases := [][]string{
		{}, {"unknown"}, {"inspect", "grant", "G1", "--app", "hrms"},
		{"inspect", "grant", "G1", "--tenant", "acme"},
		{"inspect", "grant", "G1", "--tenant"},
		{"inspect", "grant", "G1", "--tenant", "*", "--app", "hrms"},
		{"inspect", "grant", "G1", "--tenant", "acme", "--app", "hrms", "--skip-abv"},
		{"inspect", "grant", "G1", "--tenant", "acme", "--tenant", "other", "--app", "hrms"},
		{"inspect", "unknown", "G1", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "G1", "--db=", "--tenant", "acme", "--app", "hrms"},
		{"check", "assignment", "--file=", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"assign", "--file", "a.json", "--db", "", "--fixture-context", "maya", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "G1", "--file", "a.json", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"check", "assignment", "--file", "a.json", "--db", "x", "--case", "bad", "--tenant", "acme", "--app", "hrms"},
		{"assign", "--file", "a.json", "--db", "x", "--fixture-context", "maya", "--case", "bad", "--tenant", "acme", "--app", "hrms"},
		{"scenario", "seed", "team-fin-c17", "--file", "a.json", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"scenario", "run", "team-fin-c17", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"scenario", "seed", "team-fin-c17", "--case", "bad", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "G1", "--db", "x", "--db", "y", "--tenant", "acme", "--app", "hrms"},
	}
	for _, args := range cases {
		var out, diag bytes.Buffer
		if got := Run(context.Background(), args, strings.NewReader(""), &out, &diag, nil, nil); got != 2 {
			t.Fatalf("%q: exit %d, want malformed; %s", args, got, diag.String())
		}
		if out.Len() != 0 || diag.Len() == 0 {
			t.Fatalf("wrong output streams for %q", args)
		}
	}
}
func TestUnimplementedCommandsAreNeverSuccess(t *testing.T) {
	for _, args := range [][]string{
		{"inspect", "grant", "G1", "--db", "lab.db"},
		{"inspect", "assignment", "A1", "--db", "lab.db"},
		{"check", "assignment", "--file", "a2.json", "--db", "lab.db"},
		{"assign", "--file", "a2.json", "--fixture-context", "maya-team1", "--db", "lab.db"},
		{"scenario", "seed", "team-fin-c17", "--db", "lab.db"},
		{"scenario", "run", "team-fin-c17", "--case", "unsupported-permission", "--db", "lab.db"},
	} {
		args = append(args, "--tenant", "acme", "--app", "hrms")
		var out, diag bytes.Buffer
		if got := Run(context.Background(), args, strings.NewReader(""), &out, &diag, nil, nil); got != 5 {
			t.Fatalf("%q: %d; %s", args, got, diag.String())
		}
		if out.Len() != 0 || diag.Len() == 0 {
			t.Fatal("unsupported operation looked successful")
		}
	}
}
func TestHelpIsReusableWithoutAuthorityContext(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		var out, diag bytes.Buffer
		if got := Run(context.Background(), []string{arg}, strings.NewReader(""), &out, &diag, nil, nil); got != 0 {
			t.Fatal("help failed", got)
		}
		if !strings.Contains(out.String(), "--tenant") || !strings.Contains(out.String(), "--app") || diag.Len() != 0 {
			t.Fatal("help must explain both boundaries")
		}
	}
}
func TestCancelledCommandDoesNotDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, diag bytes.Buffer
	if got := Run(ctx, []string{"assign", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), &out, &diag, nil, nil); got != 4 {
		t.Fatal("cancelled command did not report evaluation failure", got)
	}
}

func TestExpiredDeadlineDoesNotDispatch(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	var out, diag bytes.Buffer
	if got := Run(ctx, []string{"assign", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), &out, &diag, nil, nil); got != 4 {
		t.Fatal("expired deadline did not report evaluation failure", got)
	}
	if out.Len() != 0 || diag.Len() == 0 {
		t.Fatal("deadline result used wrong output stream")
	}
}
