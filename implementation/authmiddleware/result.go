package authmiddleware

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
)

type Result struct {
	Version            string   `json:"version"`
	Decision           Decision `json:"decision"`
	GrantIDs           []string `json:"grant_ids,omitempty"`
	ErrorCode          string   `json:"error_code,omitempty"`
	ErrorMessage       string   `json:"error_message,omitempty"`
	ErrorMessageReason string   `json:"error_message_reason,omitempty"`
}

type EvaluationError struct {
	Version       string `json:"version"`
	Code          string `json:"error_code"`
	Message       string `json:"error_message"`
	MessageReason string `json:"error_message_reason"`
	Cause         error  `json:"-"`
}

func (e *EvaluationError) Error() string { return e.Message }
func (e *EvaluationError) Unwrap() error { return e.Cause }

func (r Result) Validate() error {
	if r.Version != "1" {
		return fmt.Errorf("unsupported result version %q", r.Version)
	}
	switch r.Decision {
	case Allow:
		if len(r.GrantIDs) == 0 || r.ErrorCode != "" || r.ErrorMessage != "" || r.ErrorMessageReason != "" {
			return errors.New("invalid allow result")
		}
		for _, id := range r.GrantIDs {
			if invalidText(id) {
				return errors.New("invalid allow grant ID")
			}
		}
	case Deny:
		if len(r.GrantIDs) != 0 || invalidText(r.ErrorCode) || invalidText(r.ErrorMessage) || invalidText(r.ErrorMessageReason) {
			return errors.New("invalid deny result")
		}
	default:
		return fmt.Errorf("unsupported decision %q", r.Decision)
	}
	return nil
}

func DecodeResult(raw []byte) (Result, error) {
	if err := validateJSON(raw); err != nil {
		return Result{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return Result{}, errors.New("result must be a JSON object")
	}
	if _, hasDecision := fields["decision"]; !hasDecision {
		if len(fields) != 4 || !hasOnly(fields, "version", "error_code", "error_message", "error_message_reason") {
			return Result{}, errors.New("invalid evaluation error variant")
		}
		var evaluationError EvaluationError
		if err := strictDecode(raw, &evaluationError); err != nil {
			return Result{}, err
		}
		if evaluationError.Version != "1" || invalidText(evaluationError.Code) || invalidText(evaluationError.Message) || invalidText(evaluationError.MessageReason) {
			return Result{}, errors.New("invalid evaluation error")
		}
		return Result{}, &evaluationError
	}

	var result Result
	if err := strictDecode(raw, &result); err != nil {
		return Result{}, err
	}
	switch result.Decision {
	case Allow:
		if len(fields) != 3 || !hasOnly(fields, "version", "decision", "grant_ids") {
			return Result{}, errors.New("invalid allow fields")
		}
	case Deny:
		if len(fields) != 5 || !hasOnly(fields, "version", "decision", "error_code", "error_message", "error_message_reason") {
			return Result{}, errors.New("invalid deny fields")
		}
	}
	if err := result.Validate(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func hasOnly(fields map[string]json.RawMessage, names ...string) bool {
	if len(fields) != len(names) {
		return false
	}
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return false
		}
	}
	return true
}
