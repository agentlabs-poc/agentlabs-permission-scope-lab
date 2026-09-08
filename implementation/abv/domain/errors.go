package domain

import "errors"

// Internal error categories, not canonical public error codes.
var (
	ErrMalformed   = errors.New("malformed input")
	ErrRejected    = errors.New("authority rejected")
	ErrUnsupported = errors.New("unsupported operation or representation")
	ErrNotFound    = errors.New("record not found")
	ErrConflict    = errors.New("authority write conflict")
	ErrUnavailable = errors.New("authority evidence unavailable")
)
