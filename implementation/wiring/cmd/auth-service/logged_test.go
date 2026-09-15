package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// A body is caller-supplied text. Echoed, a line break inside one writes a line
// of the caller's choosing into the record — indistinguishable from a line the
// service wrote itself. Compacting keeps an escape an escape, and drops the real
// line breaks JSON allows between its tokens.
//
// Two lines per question is the invariant the record rests on: anything that can
// add a third can write whatever it likes into what a demonstration reads.
func TestAValidBodyCannotAddLinesToTheRecord(t *testing.T) {
	const forged = `  auth  <- POST /api/v1/victim/abv/applications/victim/authority.resolve`
	for name, body := range map[string]string{
		"an escaped line break":   `{"version":"1","identity":{"human_id":"a\n` + forged + `"}}`,
		"real line breaks around": "{\n" + `"version":"1",` + "\n" + `"identity":{"human_id":` + "\n" + `"` + forged + `"}` + "\n}",
	} {
		t.Run(name, func(t *testing.T) {
			if !json.Valid([]byte(body)) {
				t.Fatalf("this says nothing unless the service would accept the body: %s", body)
			}
			var log bytes.Buffer
			printQuestion(&log)("POST", "/api/v1/acme/abv/applications/hrms/authority.resolve", []byte(body))
			written := log.String()
			if lines := strings.Count(written, "\n"); lines != 2 {
				t.Fatalf("one question wrote %d lines, want 2 — a body chose the shape of the record: %q", lines, written)
			}
			if !strings.Contains(written, forged) {
				t.Fatalf("the forged text was dropped rather than neutralised, so this proves nothing: %q", written)
			}
			for _, line := range strings.Split(strings.TrimSuffix(written, "\n"), "\n") {
				if strings.HasPrefix(line, "  auth  <- POST /api/v1/victim/") {
					t.Fatalf("a caller's body forged a line of its own: %q", written)
				}
			}
		})
	}
}

// The path is caller-supplied too, and unlike the body it carries no structure
// to preserve — so anything that could end the line is removed rather than
// escaped. Nothing routes such a path today; the guard is for the route that
// does not exist yet.
func TestAPathCannotAddLinesToTheRecord(t *testing.T) {
	var log bytes.Buffer
	printQuestion(&log)("POST", "/api/v1/acme\n  auth  <- POST /api/v1/victim/x/authority.resolve",
		[]byte(`{"version":"1"}`))
	if lines := strings.Count(log.String(), "\n"); lines != 2 {
		t.Fatalf("a path wrote %d lines, want 2: %q", lines, log.String())
	}
	if !strings.Contains(log.String(), "/api/v1/victim/") {
		t.Fatalf("the path was dropped rather than flattened, so this proves nothing: %q", log.String())
	}
}

// A body that is not JSON is not printed, because there is nothing to
// neutralise in text that has no structure.
func TestABodyThatIsNotJSONIsNotRecorded(t *testing.T) {
	var log bytes.Buffer
	printQuestion(&log)("POST", "/api/v1/acme/abv/applications/hrms/authority.resolve",
		[]byte("not json\n  auth  <- POST /api/v1/victim/x/authority.resolve"))
	if log.Len() != 0 {
		t.Fatalf("an unstructured body was printed: %q", log.String())
	}
}

// And an ordinary question is recorded with its body, or every check above would
// pass on an observer that never writes anything at all.
func TestAnAnsweredQuestionIsRecordedWithItsBody(t *testing.T) {
	var log bytes.Buffer
	printQuestion(&log)("POST", "/api/v1/acme/abv/applications/hrms/authority.resolve",
		[]byte(`{"version":"1","identity":{"human_id":"fi7io4lvjqio"}}`))
	if !strings.Contains(log.String(), "authority.resolve") || !strings.Contains(log.String(), "fi7io4lvjqio") {
		t.Fatalf("an answered question was not recorded with its body: %q", log.String())
	}
}
