// Package httpdemo is a bounded in-process example of application handlers
// protected by authmiddleware. It is not a production identity adapter.
package hrms

import (
	"agentlabs.local/authmiddleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

var errBadRequest = errors.New("bad request")

const (
	payslipRead  = "hrms:payroll:payslip::read"
	payslipWrite = "hrms:payroll:payslip::write"
)

type Record struct {
	TenantID      string `json:"tenant_id"`
	DepartmentID  string `json:"department_id"`
	CertificateID string `json:"certificate_id"`
	EmployeeID    string `json:"employee_id"`
	OwnerID       string `json:"owner_id"`
	Title         string `json:"title"`
}

type Store struct {
	mu      sync.Mutex
	records []Record
	writes  []WriteEvidence
}

// WriteEvidence is what the write path records beside the change it made. An
// allow is not a bare yes — it names the grants that authorized the request —
// and the effect is the only place that can write those down next to what they
// permitted. The endpoint can only record evidence the gate handed it.
type WriteEvidence struct {
	TenantID      string
	CertificateID string
	GrantIDs      []string
}

// Writes returns the evidence recorded so far, oldest first.
func (s *Store) Writes() []WriteEvidence {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]WriteEvidence(nil), s.writes...)
}

func (s *Store) record(evidence WriteEvidence) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writes = append(s.writes, evidence)
}

func NewStore(records []Record) *Store {
	return &Store{records: append([]Record(nil), records...)}
}

func DefaultRecords() []Record {
	return []Record{
		{TenantID: "acme", DepartmentID: "FIN", CertificateID: "C17", EmployeeID: "fi7io4lvjqio", OwnerID: "fi7io4lvjqio", Title: "FIN annual"},
		{TenantID: "acme", DepartmentID: "FIN", CertificateID: "C19", EmployeeID: "fi7io4lvjwu8", OwnerID: "fi7io4lvjqio", Title: "FIN supplemental"},
		{TenantID: "acme", DepartmentID: "ENG", CertificateID: "C18", EmployeeID: "fi7io4lvk35s", OwnerID: "fi7io4lvk35s", Title: "ENG confidential"},
	}
}

func (s *Store) Get(tenant, department, certificate string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range s.records {
		if record.TenantID == tenant && record.DepartmentID == department && record.CertificateID == certificate {
			return record, true
		}
	}
	return Record{}, false
}

func (s *Store) list(tenant, department string) []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	var records []Record
	for _, record := range s.records {
		if record.TenantID == tenant && (department == "" || record.DepartmentID == department) {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].CertificateID < records[j].CertificateID })
	return records
}

// Move relocates a certificate to another department, which is the one change
// that moves a record across an authorization boundary.
//
// No endpoint exposes it. ENFORCEMENT-004 requires authority covering *both* the
// current and the proposed boundary for such a move, and that rule is not
// implemented here — so exposing this would be shipping the half of the shape
// that is easy. It exists because the boundary change is what Q-074 is about,
// and a rule about a record leaving its authorized boundary between the decision
// and the effect cannot be tested without something that moves one.
func (s *Store) Move(tenant, from, to, certificate string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.records {
		record := &s.records[i]
		if record.TenantID == tenant && record.DepartmentID == from && record.CertificateID == certificate {
			record.DepartmentID = to
			return true
		}
	}
	return false
}

func (s *Store) update(tenant, department, certificate, title string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.records {
		record := &s.records[i]
		if record.TenantID == tenant && record.DepartmentID == department && record.CertificateID == certificate {
			record.Title = title
			return *record, true
		}
	}
	return Record{}, false
}

type fixedIdentity struct{ requestContext authmiddleware.RequestContext }

func TrustedIdentity(tenant, application, human string) authmiddleware.IdentitySource {
	return fixedIdentity{requestContext: authmiddleware.RequestContext{
		Area:     authmiddleware.Area{TenantID: tenant, ApplicationID: application},
		Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: human}, HumanID: human},
	}}
}

func (i fixedIdentity) Establish(context.Context, *http.Request) (authmiddleware.RequestContext, error) {
	return i.requestContext, nil
}

func NewHandler(store *Store, evaluator *authmiddleware.Evaluator, identity authmiddleware.IdentitySource) (http.Handler, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}
	type route struct {
		policy authmiddleware.Policy
		bind   authmiddleware.Binder
	}
	routes := []route{
		{policy(http.MethodGet, "/api/v1/{tenant}/{dept}/{cert}", payslipRead, map[string]authmiddleware.Input{
			"tenant": {Source: authmiddleware.SourcePath, Name: "tenant"}, "dept": {Source: authmiddleware.SourcePath, Name: "dept"}, "cert": {Source: authmiddleware.SourcePath, Name: "cert"},
		}, "tenant"), store.bindGet},
		{policy(http.MethodPut, "/api/v1/{tenant}/certificates/{cert}", payslipWrite, map[string]authmiddleware.Input{
			"tenant": {Source: authmiddleware.SourcePath, Name: "tenant"}, "cert": {Source: authmiddleware.SourcePath, Name: "cert"},
			"proposed_dept": {Source: authmiddleware.SourceBody, Name: "department_id"}, "title": {Source: authmiddleware.SourceBody, Name: "title"},
		}, "tenant"), store.bindPut},
		{policy(http.MethodGet, "/api/v1/{tenant}/departments/{dept}/certificates", payslipRead, map[string]authmiddleware.Input{
			"tenant": {Source: authmiddleware.SourcePath, Name: "tenant"}, "dept": {Source: authmiddleware.SourcePath, Name: "dept"},
		}, "tenant"), store.bindDepartment},
		{policy(http.MethodGet, "/api/v1/{tenant}/certificates", payslipRead, map[string]authmiddleware.Input{
			"tenant": {Source: authmiddleware.SourcePath, Name: "tenant"},
		}, "tenant"), store.bindAll},
	}
	mux := http.NewServeMux()
	for _, route := range routes {
		wrapped, err := authmiddleware.Wrap(route.policy, identity, evaluator, route.bind, renderFailure)
		if err != nil {
			return nil, err
		}
		mux.Handle(route.policy.Method+" "+route.policy.Path, wrapped)
	}
	return mux, nil
}

// policy names, as its last argument, the input the gate must hold against the
// trusted tenant. Every route here is about one tenant's certificates, so every
// one of them declares it; the gate refuses to mount a policy that does not.
func policy(method, path, permission string, inputs map[string]authmiddleware.Input, tenantInput string) authmiddleware.Policy {
	return authmiddleware.Policy{
		Version: "1", Method: method, Path: path, Permission: permission, Inputs: inputs,
		Trusted: map[string]string{authmiddleware.TrustedTenant: tenantInput},
	}
}

func (s *Store) bindGet(_ context.Context, _ authmiddleware.RequestContext, values authmiddleware.InputValues, _ map[string]json.RawMessage) (authmiddleware.BoundOperation, error) {
	tenant, dept, cert, err := strings3(values, "tenant", "dept", "cert")
	if err != nil {
		return authmiddleware.BoundOperation{}, err
	}
	record, found := s.Get(tenant, dept, cert)
	material := authmiddleware.Material{
		"dept": {Kind: authmiddleware.SelectionExact, Value: dept}, "cert": {Kind: authmiddleware.SelectionExact, Value: cert},
	}
	if found {
		material["user"] = authmiddleware.Selection{Kind: authmiddleware.SelectionExact, Value: record.EmployeeID}
	}
	return authmiddleware.BoundOperation{Material: material, Execute: func(_ context.Context, w http.ResponseWriter, _ authmiddleware.Result) {
		record, ok := s.Get(tenant, dept, cert)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, record)
	}}, nil
}

func (s *Store) bindPut(_ context.Context, _ authmiddleware.RequestContext, values authmiddleware.InputValues, body map[string]json.RawMessage) (authmiddleware.BoundOperation, error) {
	if len(body) != 2 {
		return authmiddleware.BoundOperation{}, errBadRequest
	}
	tenant, cert, dept, err := strings3(values, "tenant", "cert", "proposed_dept")
	if err != nil {
		return authmiddleware.BoundOperation{}, err
	}
	title, err := stringValue(values, "title")
	if err != nil || len(title) > 200 {
		return authmiddleware.BoundOperation{}, errBadRequest
	}
	return authmiddleware.BoundOperation{
		Material: authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: dept}},
		Execute: func(_ context.Context, w http.ResponseWriter, result authmiddleware.Result) {
			record, ok := s.update(tenant, dept, cert, title)
			if !ok {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			// Recorded only for a change that actually happened, and only from
			// the result the gate passed in.
			s.record(WriteEvidence{TenantID: tenant, CertificateID: cert, GrantIDs: append([]string(nil), result.GrantIDs...)})
			writeJSON(w, http.StatusOK, record)
		},
	}, nil
}

func (s *Store) bindDepartment(_ context.Context, _ authmiddleware.RequestContext, values authmiddleware.InputValues, _ map[string]json.RawMessage) (authmiddleware.BoundOperation, error) {
	tenant, dept, err := strings2(values, "tenant", "dept")
	if err != nil {
		return authmiddleware.BoundOperation{}, err
	}
	return authmiddleware.BoundOperation{
		Material: authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: dept}},
		Execute: func(_ context.Context, w http.ResponseWriter, _ authmiddleware.Result) {
			writeJSON(w, http.StatusOK, s.list(tenant, dept))
		},
	}, nil
}

func (s *Store) bindAll(_ context.Context, _ authmiddleware.RequestContext, values authmiddleware.InputValues, _ map[string]json.RawMessage) (authmiddleware.BoundOperation, error) {
	tenant, err := stringValue(values, "tenant")
	if err != nil {
		return authmiddleware.BoundOperation{}, err
	}
	return authmiddleware.BoundOperation{
		Material: authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionAll}},
		Execute: func(_ context.Context, w http.ResponseWriter, _ authmiddleware.Result) {
			writeJSON(w, http.StatusOK, s.list(tenant, ""))
		},
	}, nil
}

func stringValue(values authmiddleware.InputValues, name string) (string, error) {
	var value string
	if err := json.Unmarshal(values[name], &value); err != nil || value == "" || value != strings.TrimSpace(value) || !utf8.ValidString(value) || strings.Contains(value, "*") {
		return "", errBadRequest
	}
	return value, nil
}

func strings2(values authmiddleware.InputValues, a, b string) (string, string, error) {
	first, err := stringValue(values, a)
	if err != nil {
		return "", "", err
	}
	second, err := stringValue(values, b)
	return first, second, err
}

func strings3(values authmiddleware.InputValues, a, b, c string) (string, string, string, error) {
	first, second, err := strings2(values, a, b)
	if err != nil {
		return "", "", "", err
	}
	third, err := stringValue(values, c)
	return first, second, third, err
}

func renderFailure(w http.ResponseWriter, _ *http.Request, result authmiddleware.Result, err error) {
	if result.Decision == authmiddleware.Deny {
		writeJSON(w, http.StatusForbidden, result)
		return
	}
	var evaluationError *authmiddleware.EvaluationError
	if errors.As(err, &evaluationError) {
		writeJSON(w, http.StatusServiceUnavailable, evaluationError)
		return
	}
	// Everything left is either the caller's input or a source that failed, and
	// the two are told apart by type rather than guessed at: an AuthoritySource
	// reports failure as an EvaluationError, which the branch above catches.
	//
	// Making every error 503 was tempting and wrong — an existing test caught it
	// immediately, because a body with a numeric title is a bad request however
	// healthy Auth is. The fix belongs in the sources.
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		// The same canonical block every other evaluation failure carries. It
		// used to answer {"error":"request failed"} with no code and neither
		// message, so the one outcome a person most needs told apart from a
		// denial arrived in a shape no consumer could read — while its sibling
		// branch above carried all four fields.
		writeJSON(w, http.StatusServiceUnavailable, &authmiddleware.EvaluationError{
			Version:       "1",
			Code:          "AUTHORITY_TIMEOUT",
			Message:       "We could not check your access.",
			MessageReason: "the authority question did not finish",
		})
	default:
		writeError(w, http.StatusBadRequest, "request failed")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
