package authclient

import (
	"agentlabs.local/authmiddleware"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// maxResponseBytes bounds what one answer may be. A client that will read any
// length is a client an unreachable Auth can exhaust.
const maxResponseBytes = 1 << 20

// defaultTimeout bounds one question, and loadTimeout bounds it again inside
// Load so that a caller-supplied Doer without one cannot hang a request
// forever. Each hung authorization pins a goroutine and a connection; an Auth
// that accepts and stalls would otherwise take the application down with it.
const (
	defaultTimeout = 5 * time.Second
	loadTimeout    = 10 * time.Second
)

// refuseRedirect stops the client following a Location header.
//
// An authorization answer must come from the host that was dialled. Go's default
// client follows up to ten redirects, and this client hands whatever comes back
// to the gate as authority — so a single Location header turned the first
// network-crossing decision in this system into an attacker-authored one, with
// the bearer credential forwarded along with it. Demonstrated, not theorised.
func refuseRedirect(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

// Error is a failure to establish authority, which is never a denial.
//
// Q-128 is explicit that the two must stay apart: "Failure to establish evidence
// remains an evaluation failure, distinct from a policy denial." A gate that
// turned an unreachable Auth into a deny would fail closed in the reassuring
// direction and lie about why.
type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

// evaluation renders a failure as the canonical block, so it reaches the
// application as an evaluation error rather than as anything else.
//
// Returning a bare error was not enough. The application's own failure handler
// switches on *authmiddleware.EvaluationError to answer 503, and a plain error
// fell through to its default — so an Auth outage was reported to the caller as
// *their* malformed request: no availability signal, and the 503 branch the
// codebase already had went unreached.
func (e *Error) evaluation() *authmiddleware.EvaluationError {
	return &authmiddleware.EvaluationError{
		Version: "1", Code: e.Code,
		Message:       "We could not check your access.",
		MessageReason: e.Message,
		Cause:         e,
	}
}

// Doer is the HTTP client, taken as an interface so a deployment supplies its
// own timeouts, transport and instrumentation rather than inheriting ours.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Credential is what this application authenticates as. It is not the person the
// question is about.
type Credential struct {
	// Type is a Q-086 actor type — `service_account` for an application.
	Type string
	// ID identifies the credential the Auth service issued.
	ID string
	// Bearer is sent as the Authorization header when set. The lab has no
	// issuer, so it is optional here and will not be once one exists.
	Bearer string
}

// Source answers the gate's authority question over HTTP.
type Source struct {
	base       *url.URL
	credential Credential
	doer       Doer
}

// New builds a Source against one Auth service.
//
// It takes a base URL where the other implementation takes a database path, and
// that is the whole difference an application sees.
func New(baseURL string, credential Credential, doer Doer) (*Source, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, &Error{Code: "MISCONFIGURED", Message: "a base URL is required"}
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" {
		return nil, &Error{Code: "MISCONFIGURED", Message: "malformed base URL " + baseURL, Cause: err}
	}
	// Cleartext carries an authorization answer that anything on the path may
	// rewrite into a blanket grant. Allowed only when a deployment says so out
	// loud, which a lab does and a deployment should not.
	if scheme := strings.ToLower(parsed.Scheme); scheme != "https" && !(scheme == "http" && allowCleartext) {
		return nil, &Error{Code: "MISCONFIGURED", Message: "the authority service must be reached over https: " + baseURL}
	}
	if credential.Type == "" || credential.ID == "" {
		return nil, &Error{Code: "MISCONFIGURED", Message: "a credential is required: an application asks as itself, not as the human"}
	}
	if doer == nil {
		doer = &http.Client{Timeout: defaultTimeout, CheckRedirect: refuseRedirect}
	}
	return &Source{base: parsed, credential: credential, doer: doer}, nil
}

// Load resolves what one human holds in one area — all of it.
func (s *Source) Load(ctx context.Context, query authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	if ctx == nil {
		return authmiddleware.Authority{}, (&Error{Code: "MISCONFIGURED", Message: "context is required"}).evaluation()
	}
	// Bounded regardless of the Doer. A caller that supplies its own client
	// without a timeout — which is exactly what a test server's client does —
	// must not be able to block a decision indefinitely.
	ctx, cancel := context.WithTimeout(ctx, loadTimeout)
	defer cancel()
	body, err := json.Marshal(resolveRequest{
		Version: Version,
		Identity: identity{
			Version: Version,
			Actor:   actor{Type: s.credential.Type, ID: s.credential.ID},
			HumanID: query.Context.Identity.HumanID,
		},
		// No permission filter. The complete answer is the cacheable one, and an
		// answer narrowed to the permission in hand could only ever have served
		// the request that asked for it. Nothing caches yet; the shape is what
		// makes caching possible later without changing the contract.
		Options: options{},
	})
	if err != nil {
		return authmiddleware.Authority{}, (&Error{Code: "MISCONFIGURED", Message: "could not encode the request", Cause: err}).evaluation()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.route(query), bytes.NewReader(body))
	if err != nil {
		return authmiddleware.Authority{}, (&Error{Code: "MISCONFIGURED", Message: "could not build the request", Cause: err}).evaluation()
	}
	request.Header.Set("Content-Type", "application/json")
	if s.credential.Bearer != "" {
		request.Header.Set("Authorization", "Bearer "+s.credential.Bearer)
	}
	response, err := s.doer.Do(request)
	if err != nil {
		return authmiddleware.Authority{}, (&Error{Code: "AUTH_UNREACHABLE", Message: "the authority service did not answer", Cause: err}).evaluation()
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// Drained before closing so the connection returns to the pool. During
		// an incident every retry would otherwise pay a fresh handshake against
		// an already-struggling service.
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return authmiddleware.Authority{}, (&Error{
			Code:    "AUTH_REFUSED",
			Message: fmt.Sprintf("the authority service answered %d", response.StatusCode),
		}).evaluation()
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return authmiddleware.Authority{}, (&Error{Code: "AUTH_UNREACHABLE", Message: "the answer could not be read", Cause: err}).evaluation()
	}
	if len(raw) > maxResponseBytes {
		return authmiddleware.Authority{}, (&Error{Code: "AUTH_OVERSIZED", Message: "the answer exceeded the accepted size"}).evaluation()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var answer resolveResponse
	if err := decoder.Decode(&answer); err != nil {
		return authmiddleware.Authority{}, (&Error{Code: "AUTH_MALFORMED", Message: "the answer did not match the contract", Cause: err}).evaluation()
	}
	// After the decode, so that text which is not the contract at all is
	// reported as malformed rather than as ambiguous — but before the answer is
	// used for anything, because an answer that says two things has not been
	// read yet.
	if err := rejectAmbiguousJSON(raw); err != nil {
		return authmiddleware.Authority{}, (&Error{Code: "AUTH_AMBIGUOUS", Message: "the answer did not mean exactly one thing", Cause: err}).evaluation()
	}
	return answer.decode(query)
}

// AllowCleartext permits http:// base URLs process-wide. It exists for the lab
// and for tests; a deployment that calls it is choosing to let anything on the
// path author its authorization answers.
func AllowCleartext() { allowCleartext = true }

var allowCleartext bool

// route is the pinned tenant base plus the application segment. The path is this
// package's, not a deployment's: a client needs a base URL and nothing else, and
// no deployment can move a route and silently break every agent.
func (s *Source) route(query authmiddleware.AuthorityQuery) string {
	// Escaped, because an area id is not guaranteed to be a single safe path
	// segment: domain.NewArea admits "..", "a/b" and "?", and JoinPath cleans
	// rather than escapes — so a crafted area could send this question, with the
	// subject and the credential, to a different tenant's route entirely.
	return s.base.JoinPath(
		"api", "v1", url.PathEscape(query.Context.Area.TenantID), "abv",
		"applications", url.PathEscape(query.Context.Area.ApplicationID), "authority.resolve",
	).String()
}
