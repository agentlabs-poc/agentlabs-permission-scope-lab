package main

import (
	"strings"
	"testing"
)

// This package had no test file at all, and the argument parsing changed under
// it — a flag whose value is missing now exits rather than defaulting in
// silence. Shipping that with nothing pinning it is the pattern this repository
// keeps repeating.
func TestArgumentsThatDoNotNameAConfiguration(t *testing.T) {
	t.Setenv("HRMS_AUTH_TOKEN", "shared_secret_fyd2k7x0q4nb")
	// Cleartext, and not asked for — so if parsing ever succeeds, authclient
	// refuses the URL and returns before anything binds a port. Without that, a
	// case whose flag was wrongly accepted starts a listening server and the test
	// hangs instead of failing, which is a worse signal than none.
	complete := []string{
		"--auth", "http://127.0.0.1:1", "--tenant", "acme", "--app", "hrms",
		"--human", "fi7io4lvjqio", "--client", "agent_hrms", "--listen", "127.0.0.1:8081",
	}
	without := func(flag string) []string {
		out := []string{}
		for i := 0; i < len(complete); i++ {
			if complete[i] == flag {
				i++
				continue
			}
			out = append(out, complete[i])
		}
		return out
	}

	for name, tc := range map[string]struct {
		args []string
		says string
	}{
		"nothing at all":  {nil, "usage"},
		"no auth service": {without("--auth"), "usage"},
		"no human":        {without("--human"), "usage"},
		// The hazard the change exists for: an operator who said exactly where to
		// listen, and whose value went missing. It used to take the default port.
		"a listen flag with no value": {append(without("--listen"), "--listen"), "--listen needs a value"},
		"an auth flag with no value":  {append(without("--auth"), "--auth"), "--auth needs a value"},
	} {
		t.Run(name, func(t *testing.T) {
			var out, diag strings.Builder
			if code := run(tc.args, &out, &diag); code != 2 {
				t.Fatalf("exit = %d, want 2 — %s", code, diag.String())
			}
			if !strings.Contains(diag.String(), tc.says) {
				t.Fatalf("diagnostic %q does not say %q", diag.String(), tc.says)
			}
		})
	}

	// A switch is not a flag with a missing value, including in final position.
	t.Run("a switch in final position", func(t *testing.T) {
		var out, diag strings.Builder
		// Cleartext is asked for, so the http URL is accepted; the listen address
		// is not one, so it fails after parsing rather than during it.
		// The switch is last, and asking for cleartext is what lets the http URL
		// through — so reaching the later failure at all proves the switch was
		// read as a switch.
		// An address nothing can bind, so the run ends the moment parsing and
		// construction are done rather than serving.
		args := append(without("--listen"), "--listen", "256.256.256.256:1", "--allow-cleartext")
		if code := run(args, &out, &diag); strings.Contains(diag.String(), "needs a value") {
			t.Fatalf("exit %d: a switch was read as a flag needing a value: %s", code, diag.String())
		}
		if strings.Contains(diag.String(), "over https") {
			t.Fatalf("the switch did not take effect: %s", diag.String())
		}
	})

	// And the secret is not a flag: it arrives in the environment or not at all.
	t.Run("no token in the environment", func(t *testing.T) {
		t.Setenv("HRMS_AUTH_TOKEN", "")
		var out, diag strings.Builder
		if code := run(complete, &out, &diag); code != 2 {
			t.Fatalf("exit = %d, want 2", code)
		}
		if !strings.Contains(diag.String(), "HRMS_AUTH_TOKEN") {
			t.Fatalf("diagnostic %q does not name the token", diag.String())
		}
	})
}
