package wiring

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// MaxRequestBytes bounds one question. A service that reads any length is one a
// caller can exhaust.
//
// It is exported because anything wrapping this handler has to read to the same
// limit: a wrapper that reads less and hands on what it read would make the
// service reject a question it would otherwise have answered.
const MaxRequestBytes = 1 << 16

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
func (s *Service) Handler(agents AgentIdentity, observers ...Observer) (http.Handler, error) {
	if agents == nil {
		return nil, errors.New("an agent identity source is required")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/{tenant}/abv/applications/{application}/authority.resolve",
		func(w http.ResponseWriter, r *http.Request) { s.resolve(agents, observers, w, r) })
	return mux, nil
}

// Observer is told about a question this service accepted and answered, and is
// how a demonstration shows what actually crossed the boundary.
//
// It is a hook here rather than a wrapper around the handler because a wrapper
// has to read the body to see the question, and reading it outside this function
// puts that read in front of the authentication below — which would let an
// unauthenticated caller make the service buffer 64 KiB per request, and put
// text of its choosing into the record a demonstration reads. The service calls
// this only for a question it authenticated, decoded and answered.
type Observer func(method, path string, question []byte)

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
func (s *Service) resolve(agents AgentIdentity, observers []Observer, w http.ResponseWriter, r *http.Request) {
	area, err := domain.NewArea(r.PathValue("tenant"), r.PathValue("application"))
	if err != nil {
		fail(w, http.StatusBadRequest, "MALFORMED_AREA")
		return
	}
	// Authenticated before the body is parsed. An unauthenticated caller should
	// not get schema feedback, nor make the service decode 64 KiB per request
	// before being turned away.
	caller, err := agents.Establish(r)
	if err != nil {
		fail(w, http.StatusUnauthorized, "UNAUTHENTICATED")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxRequestBytes+1))
	if err != nil || len(body) > MaxRequestBytes {
		fail(w, http.StatusBadRequest, "MALFORMED_REQUEST")
		return
	}
	if err := rejectDuplicateKeys(body); err != nil {
		fail(w, http.StatusBadRequest, "MALFORMED_REQUEST")
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var asked wireRequest
	if err := decoder.Decode(&asked); err != nil {
		fail(w, http.StatusBadRequest, "MALFORMED_REQUEST")
		return
	}
	if decoder.More() {
		fail(w, http.StatusBadRequest, "MALFORMED_REQUEST")
		return
	}
	// Unsupported rather than malformed: the document says so, and the two are
	// different answers — "we do not speak that" against "that is not a request".
	if asked.Version != "1" {
		fail(w, http.StatusNotImplemented, "UNSUPPORTED_VERSION")
		return
	}
	// The identity block states its own contract version, so that the identity
	// contract can evolve independently of the transport one
	// (identity-context.md). It was decoded and discarded, which meant a v2
	// block — one where human_id became a tenant-qualified reference, say — was
	// read by a v1 server as a bare human id and answered about whoever that
	// spelling happened to name. CONTRACT-010 is explicit that a consumer
	// rejects an unsupported version rather than guessing a default.
	if asked.Identity.Version != "1" {
		fail(w, http.StatusNotImplemented, "UNSUPPORTED_VERSION")
		return
	}
	// The subject comes from the body; the actor comes from the credential. That
	// is the whole shape: an application asks as itself about somebody else.
	//
	// What the body *claims* about the actor is read only to refuse a claim that
	// contradicts the credential. It is never adopted: a body has no way to
	// prove who is asking, and the day somebody wires delegation evidence in is
	// the day that distinction stops being free.
	if claimed := asked.Identity.Actor; claimed.Type != "" || claimed.ID != "" {
		if claimed.Type != caller.Actor.Type || claimed.ID != caller.Actor.ID {
			fail(w, http.StatusForbidden, "NOT_ENTITLED_TO_ASK")
			return
		}
	}
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
	for _, observe := range observers {
		observe(r.Method, r.URL.Path, body)
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
		// Not found and not entitled are one answer to a caller who is not
		// established in this area. Splitting them let one valid credential
		// enumerate which tenants exist and which applications each has
		// installed — invisible in one process, a scanner on the wire.
		return http.StatusForbidden
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
		return "NOT_ENTITLED_TO_ASK"
	case errors.Is(err, domain.ErrUnsupported):
		return "UNSUPPORTED"
	default:
		return "AUTHORITY_UNAVAILABLE"
	}
}

// rejectDuplicateKeys refuses a body that names a field twice, or that carries a
// second document after the first.
//
// encoding/json takes the last occurrence silently, so a request naming human_id
// twice was answered about the second while any log, proxy or audit reading the
// first recorded the other. authmiddleware rejects both for its own two wire
// contracts; this one sits beside them and was held to a weaker standard.
func rejectDuplicateKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := scanValue(decoder); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("trailing JSON")
	}
	return nil
}

func scanValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return errors.New("malformed object key")
			}
			if seen[name] {
				return errors.New("duplicate JSON field " + name)
			}
			seen[name] = true
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	}
	_, err = decoder.Token()
	return err
}

func fail(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"version": "1", "error_code": code})
}
