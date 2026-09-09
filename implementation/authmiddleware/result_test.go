package authmiddleware

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeResultCanonicalVariants(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Result
		err  *EvaluationError
	}{
		{"allow", `{"version":"1","decision":"allow","grant_ids":["G-17"]}`, Result{Version: "1", Decision: Allow, GrantIDs: []string{"G-17"}}, nil},
		{"deny", `{"version":"1","decision":"deny","error_code":"NO_AUTHORIZING_GRANT","error_message":"You do not have access to this certificate.","error_message_reason":"No grant authorizes this certificate read within Finance."}`, Result{Version: "1", Decision: Deny, ErrorCode: "NO_AUTHORIZING_GRANT", ErrorMessage: "You do not have access to this certificate.", ErrorMessageReason: "No grant authorizes this certificate read within Finance."}, nil},
		{"evaluation error", `{"version":"1","error_code":"AUTH_SERVICE_TIMEOUT","error_message":"We could not check your access.","error_message_reason":"The authorization service did not respond in time."}`, Result{}, &EvaluationError{Version: "1", Code: "AUTH_SERVICE_TIMEOUT", Message: "We could not check your access.", MessageReason: "The authorization service did not respond in time."}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeResult([]byte(tc.raw))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("result = %#v", got)
			}
			if tc.err == nil {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var gotErr *EvaluationError
			if !errors.As(err, &gotErr) || !reflect.DeepEqual(gotErr, tc.err) {
				t.Fatalf("error = %#v", err)
			}
			if gotErr.Error() != tc.err.Message || gotErr.MessageReason != tc.err.MessageReason {
				t.Fatalf("messages unavailable: %#v", gotErr)
			}
		})
	}
}

func TestDecodeResultRejectsMixedMalformedAndUnsupportedVariants(t *testing.T) {
	allow := `{"version":"1","decision":"allow","grant_ids":["G-17"]}`
	cases := map[string]string{
		"missing version":     strings.Replace(allow, `"version":"1",`, "", 1),
		"future version":      strings.Replace(allow, `"version":"1"`, `"version":"2"`, 1),
		"unknown field":       strings.Replace(allow, `}`, `,"debug_note":"checked"}`, 1),
		"duplicate field":     strings.Replace(allow, `"decision":"allow"`, `"decision":"allow","decision":"deny"`, 1),
		"wrong decision type": strings.Replace(allow, `"decision":"allow"`, `"decision":true`, 1),
		"wrong grant ID type": strings.Replace(allow, `["G-17"]`, `[17]`, 1),
		"mixed allow error":   strings.Replace(allow, `}`, `,"error_code":"AUTH_SERVICE_TIMEOUT"}`, 1),
		"mixed deny grants":   `{"version":"1","decision":"deny","grant_ids":[],"error_code":"X","error_message":"m","error_message_reason":"r"}`,
		"error with decision": `{"version":"1","decision":"error","error_code":"X","error_message":"m","error_message_reason":"r"}`,
		"empty allow grants":  `{"version":"1","decision":"allow","grant_ids":[]}`,
		"null allow grants":   `{"version":"1","decision":"allow","grant_ids":null}`,
		"empty grant ID":      `{"version":"1","decision":"allow","grant_ids":[""]}`,
		"truncated deny":      `{"version":"1","decision":"deny","error_code":"X"}`,
		"truncated error":     `{"version":"1","error_code":"X","error_message":"m"}`,
		"trailing JSON":       allow + `{}`,
		"null":                `null`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeResult([]byte(raw))
			if err == nil || !reflect.DeepEqual(got, Result{}) {
				t.Fatalf("accepted invalid result: %#v, %v", got, err)
			}
			var evaluationErr *EvaluationError
			if errors.As(err, &evaluationErr) {
				t.Fatalf("malformed result became evaluation error: %#v", evaluationErr)
			}
		})
	}
}

func TestResultValidateRejectsMixedTypedVariants(t *testing.T) {
	validAllow := Result{Version: "1", Decision: Allow, GrantIDs: []string{"G-17"}}
	validDeny := Result{Version: "1", Decision: Deny, ErrorCode: "X", ErrorMessage: "message", ErrorMessageReason: "reason"}
	for name, result := range map[string]Result{
		"allow error field": {Version: "1", Decision: Allow, GrantIDs: []string{"G-17"}, ErrorCode: "X"},
		"deny grants":       {Version: "1", Decision: Deny, GrantIDs: []string{"G-17"}, ErrorCode: "X", ErrorMessage: "m", ErrorMessageReason: "r"},
		"unknown decision":  {Version: "1", Decision: "error"},
		"zero":              {},
	} {
		t.Run(name, func(t *testing.T) {
			if err := result.Validate(); err == nil {
				t.Fatalf("accepted %#v", result)
			}
		})
	}
	if err := validAllow.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := validDeny.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluationErrorUnwrapsCause(t *testing.T) {
	cause := errors.New("timeout")
	err := &EvaluationError{Version: "1", Code: "X", Message: "message", MessageReason: "reason", Cause: cause}
	if !errors.Is(err, cause) {
		t.Fatal("cause unavailable")
	}
}
