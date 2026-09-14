// Package domain holds the application registry's canonical types.
//
// This is a separate 123 domain from Auth-AL. It owns which applications exist
// and which tenants have them installed; it knows nothing about permissions,
// scopes, roles or grants, and nothing in it imports Auth-AL.
package domain

import "errors"

// The registry's errors are its own. They mirror Auth-AL's vocabulary because
// the vocabulary is good, not because the domains share anything: an error
// crossing between them is translated at the seam, never passed through.
var (
	ErrMalformed   = errors.New("malformed input")
	ErrRejected    = errors.New("registry operation rejected")
	ErrConflict    = errors.New("registry write conflict")
	ErrNotFound    = errors.New("registry record not found")
	ErrUnsupported = errors.New("unsupported operation or representation")
	ErrUnavailable = errors.New("registry evidence unavailable")
)
