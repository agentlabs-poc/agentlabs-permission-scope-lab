package wiring

import (
	"agentlabs.local/abv/domain"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// maxRequestBytes bounds one question. A service that reads any length is one a
// caller can exhaust.
const maxRequestBytes = 1 << 16

// AgentIdentity establishes which application is asking, from its own
// credentials on the request.
//
// It is a port because issuing and verifying credentials is the Auth service's
// business and not this lab's: the shape is the workload client — an id and a
// secret bound to one tenant application — and this repository models it rather
// than issuing one. What arrives here is an identity a deployment has already
// established.
type AgentIdentity interface {
	Establish(*http.Request) (domain.Identity, error)
}

// Handler is the standard endpoint set: the routes are this package's, not a
// deployment's, so a client needs a base URL and nothing else.
func (s *Service) Handler(agents AgentIdentity) (http.Handler, error) {
	if agents == nil {
		return nil, errors.New("an agent identity source is required")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/{tenant}/abv/applications/{application}/authority.resolve",
		func(w http.ResponseWriter, r *http.Request) { s.resolve(agents, w, r) })
	return mux, nil
}

type wireIdentity struct {
	Version string `json:"version"`
	Actor   struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"actor"`
	HumanID string `json:"human_id"`
}

type wireRequest struct {
	Version  string       `json:"version"`
	Identity wireIdentity `json:"identity"`
	Options  struct {
		Permissions []string `json:"permissions,omitempty"`
		OmitSource  bool     `json:"omit_source,omitempty"`
	} `json:"options"`
}

// resolve answers one question: what does this human hold here.
//
// The path selects the area and does not prove it — the caller must be
// established as entitled to ask, which is CheckAuthorityRead's job inside the
// facade, and the tenant's installation of the application is checked before any
// record is read.
func (s *Service) resolve(agents AgentIdentity, w http.ResponseWriter, r *http.Request) {
	area, err := domain.NewArea(r.PathValue("tenant"), r.PathValue("application"))
	if err != nil {
		fail(w, http.StatusBadRequest, "MALFORMED_AREA")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	if err != nil || len(body) > maxRequestBytes {
		fail(w, http.StatusBadRequest, "MALFORMED_REQUEST")
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	var asked wireRequest
	if err := decoder.Decode(&asked); err != nil || asked.Version != "1" {
		fail(w, http.StatusBadRequest, "MALFORMED_REQUEST")
		return
	}
	// The caller is whoever a deployment established from the credentials on the
	// request, never whoever the body claims. A submitted identity block is not
	// proof of anything.
	caller, err := agents.Establish(r)
	if err != nil {
		fail(w, http.StatusUnauthorized, "UNAUTHENTICATED")
		return
	}
	// The subject comes from the body; the actor comes from the credential. That
	// is the whole shape: an application asks as itself about somebody else.
	caller.HumanID = asked.Identity.HumanID
	resolved, err := s.authority.ResolveAuthority(r.Context(), area, caller, domain.ResolveOptions{
		Permissions: asked.Options.Permissions, OmitSource: asked.Options.OmitSource,
	})
	if err != nil {
		fail(w, statusFor(err), codeFor(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resolved); err != nil {
		// The status is already written; nothing useful can be said now.
		return
	}
}

// statusFor keeps the kinds apart on the wire. An empty answer is a success and
// never appears here: resolving nothing is a completed answer, not a failure.
func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrMalformed):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrRejected):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnsupported):
		return http.StatusNotImplemented
	default:
		return http.StatusServiceUnavailable
	}
}

func codeFor(err error) string {
	switch {
	case errors.Is(err, domain.ErrMalformed):
		return "MALFORMED_REQUEST"
	case errors.Is(err, domain.ErrRejected):
		return "NOT_ENTITLED_TO_ASK"
	case errors.Is(err, domain.ErrNotFound):
		return "AREA_NOT_FOUND"
	case errors.Is(err, domain.ErrUnsupported):
		return "UNSUPPORTED"
	default:
		return "AUTHORITY_UNAVAILABLE"
	}
}

func fail(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"version": "1", "error_code": code})
}
