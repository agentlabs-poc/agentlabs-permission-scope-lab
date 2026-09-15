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
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, &Error{Code: "MISCONFIGURED", Message: "malformed base URL " + baseURL, Cause: err}
	}
	if credential.Type == "" || credential.ID == "" {
		return nil, &Error{Code: "MISCONFIGURED", Message: "a credential is required: an application asks as itself, not as the human"}
	}
	if doer == nil {
		doer = &http.Client{Timeout: 5 * time.Second}
	}
	return &Source{base: parsed, credential: credential, doer: doer}, nil
}

// Load resolves what one human holds, for one permission, in one area.
func (s *Source) Load(ctx context.Context, query authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	if ctx == nil {
		return authmiddleware.Authority{}, &Error{Code: "MISCONFIGURED", Message: "context is required"}
	}
	body, err := json.Marshal(resolveRequest{
		Version: Version,
		Identity: identity{
			Version: Version,
			Actor:   actor{Type: s.credential.Type, ID: s.credential.ID},
			HumanID: query.Context.Identity.HumanID,
		},
		// One permission, because the gate is deciding one request. The complete
		// answer is the cacheable one, and nothing caches yet.
		Options: options{Permissions: []string{query.Permission}},
	})
	if err != nil {
		return authmiddleware.Authority{}, &Error{Code: "MISCONFIGURED", Message: "could not encode the request", Cause: err}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.route(query), bytes.NewReader(body))
	if err != nil {
		return authmiddleware.Authority{}, &Error{Code: "MISCONFIGURED", Message: "could not build the request", Cause: err}
	}
	request.Header.Set("Content-Type", "application/json")
	if s.credential.Bearer != "" {
		request.Header.Set("Authorization", "Bearer "+s.credential.Bearer)
	}
	response, err := s.doer.Do(request)
	if err != nil {
		return authmiddleware.Authority{}, &Error{Code: "AUTH_UNREACHABLE", Message: "the authority service did not answer", Cause: err}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return authmiddleware.Authority{}, &Error{
			Code:    "AUTH_REFUSED",
			Message: fmt.Sprintf("the authority service answered %d", response.StatusCode),
		}
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return authmiddleware.Authority{}, &Error{Code: "AUTH_UNREACHABLE", Message: "the answer could not be read", Cause: err}
	}
	if len(raw) > maxResponseBytes {
		return authmiddleware.Authority{}, &Error{Code: "AUTH_OVERSIZED", Message: "the answer exceeded the accepted size"}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var answer resolveResponse
	if err := decoder.Decode(&answer); err != nil {
		return authmiddleware.Authority{}, &Error{Code: "AUTH_MALFORMED", Message: "the answer did not match the contract", Cause: err}
	}
	return answer.decode(query)
}

// route is the pinned tenant base plus the application segment. The path is this
// package's, not a deployment's: a client needs a base URL and nothing else, and
// no deployment can move a route and silently break every agent.
func (s *Source) route(query authmiddleware.AuthorityQuery) string {
	return s.base.JoinPath(
		"api", "v1", query.Context.Area.TenantID, "abv",
		"applications", query.Context.Area.ApplicationID, "authority.resolve",
	).String()
}
