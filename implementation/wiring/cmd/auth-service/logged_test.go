package main

import (
	"agentlabs.local/wiring"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// echoing stands in for the service: it reads the body the wrapper handed on,
// and answers with whatever status the test asks for. What it received is what
// the real handler would have received.
func echoing(t *testing.T, status int, received *[]byte) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, wiring.MaxRequestBytes+1))
		if err != nil {
			t.Errorf("the service could not read the body it was handed: %v", err)
		}
		*received = body
		w.WriteHeader(status)
	})
}

// A logger that reads less than the service reads is a logger that breaks the
// service. The first version read 8 KiB and handed on the truncation, so every
// legitimate question between 8 KiB and the service's own 64 KiB limit arrived
// as invalid JSON and came back a bad request.
func TestTheLoggerHandsOnEveryByteTheServiceWouldRead(t *testing.T) {
	for name, size := range map[string]int{
		"a small question":           64,
		"past the old 8 KiB read":    12 * 1024,
		"just under the service cap": wiring.MaxRequestBytes - 64,
		"past the service's own cap": wiring.MaxRequestBytes + 512,
	} {
		t.Run(name, func(t *testing.T) {
			body := `{"version":"1","filler":"` + strings.Repeat("x", size) + `"}`
			var received []byte
			recorder := httptest.NewRecorder()
			logged(echoing(t, http.StatusOK, &received), io.Discard).ServeHTTP(recorder,
				httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve", strings.NewReader(body)))
			// Capped at what the service itself reads, and never less: the
			// service applies its own limit to the same bytes and refuses them
			// on its own terms.
			want := min(len(body), wiring.MaxRequestBytes+1)
			if len(received) != want {
				t.Fatalf("the service was handed %d bytes of a %d byte question, want %d", len(received), len(body), want)
			}
			if !bytes.HasPrefix([]byte(body), received) {
				t.Fatal("the service was handed bytes that were not the ones sent")
			}
		})
	}
}

// The log is what a demonstration reads as evidence of what crossed the
// boundary. An unauthenticated caller who can write lines into it can write
// whatever it likes into the record — so nothing is written for a question the
// service did not answer.
func TestNothingIsLoggedForAQuestionTheServiceRefused(t *testing.T) {
	for name, status := range map[string]int{
		"unauthenticated": http.StatusUnauthorized,
		"not entitled":    http.StatusForbidden,
		"malformed":       http.StatusBadRequest,
		"another version": http.StatusNotImplemented,
		"a service fault": http.StatusInternalServerError,
	} {
		t.Run(name, func(t *testing.T) {
			var log bytes.Buffer
			var received []byte
			logged(echoing(t, status, &received), &log).ServeHTTP(httptest.NewRecorder(),
				httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve",
					strings.NewReader(`{"version":"1"}`)))
			if log.Len() != 0 {
				t.Fatalf("a refused question wrote to the log: %q", log.String())
			}
		})
	}
}

// And an answered one is logged, or the check above would pass on a logger that
// never writes anything at all.
func TestAnAnsweredQuestionIsLoggedWithItsBody(t *testing.T) {
	var log bytes.Buffer
	var received []byte
	logged(echoing(t, http.StatusOK, &received), &log).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve",
			strings.NewReader(`{"version":"1","identity":{"human_id":"fi7io4lvjqio"}}`)))
	if !strings.Contains(log.String(), "authority.resolve") || !strings.Contains(log.String(), "fi7io4lvjqio") {
		t.Fatalf("an answered question was not logged with its body: %q", log.String())
	}
}

// A body is caller-supplied text. Printed raw, a newline inside one writes a
// line of the caller's choosing into the record — indistinguishable from a line
// the service wrote itself. Compacting rather than echoing keeps the escape an
// escape: the two characters backslash-n, on the one line the service wrote.
func TestALoggedBodyCannotForgeALine(t *testing.T) {
	const forged = `  auth  <- POST /api/v1/victim/abv/applications/victim/authority.resolve`
	var log bytes.Buffer
	var received []byte
	logged(echoing(t, http.StatusOK, &received), &log).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve",
			strings.NewReader(`{"version":"1","identity":{"human_id":"a\n`+forged+`"}}`)))
	written := log.String()
	if strings.Count(written, "auth  <-") != 2 {
		t.Fatalf("the forged text was not carried through verbatim, so this proves nothing: %q", written)
	}
	// Two mentions, one line: the second is inside the body, escaped, where it
	// cannot be mistaken for a line the service wrote.
	for _, line := range strings.Split(strings.TrimSuffix(written, "\n"), "\n") {
		if strings.HasPrefix(line, "  auth  <- POST /api/v1/victim/") {
			t.Fatalf("a caller's body forged a line of its own: %q", written)
		}
	}
	if !strings.Contains(written, `\n`) {
		t.Fatalf("the newline was not kept as an escape: %q", written)
	}
}

// The escape above is the easy half. JSON also admits real line breaks between
// its tokens, and those are not an escape to preserve — they are newlines in a
// body the service will happily accept. Echoing the bytes would put them in the
// record, and a question would no longer be two lines.
//
// Two lines per question is the invariant the record rests on: anything that can
// add a third can write whatever it likes into what a demonstration reads.
func TestAValidBodyCannotAddLinesToTheRecord(t *testing.T) {
	const forged = `  auth  <- POST /api/v1/victim/abv/applications/victim/authority.resolve`
	// Valid JSON: every line break is whitespace between tokens, and the forged
	// text is an ordinary string value.
	body := "{\n" + `"version":"1",` + "\n" + `"identity":{"human_id":` + "\n" + `"` + forged + `"}` + "\n}"
	if !json.Valid([]byte(body)) {
		t.Fatalf("this test only says something if the service would accept the body: %s", body)
	}
	var log bytes.Buffer
	var received []byte
	logged(echoing(t, http.StatusOK, &received), &log).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve",
			strings.NewReader(body)))
	written := log.String()
	if lines := strings.Count(written, "\n"); lines != 2 {
		t.Fatalf("one question wrote %d lines, want 2 — a body chose the shape of the record: %q", lines, written)
	}
	if !strings.Contains(written, forged) {
		t.Fatalf("the forged text was dropped rather than neutralised, so this proves nothing: %q", written)
	}
}
