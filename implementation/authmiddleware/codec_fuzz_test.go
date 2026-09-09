package authmiddleware

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// C14: accepted policies must survive serialization without changing the
// permission or input bindings; failed parsing must never return a partial policy.
func FuzzDecodePolicyPreservesBindings(f *testing.F) {
	for _, seed := range []string{
		`{"version":"1","method":"GET","path":"/api/{tenant}/{cert}","permission":"hrms:certificate::read","inputs":{"cert":{"source":"path","name":"cert"}}}`,
		`{"version":"1","method":"PUT","path":"/api/{cert}","permission":"hrms:certificate::write","inputs":{"dept":{"source":"body","name":"department_id"}}}`,
		`{"version":"1","version":"2"}`, `null`, `{"inputs":{"x":null}}`, `{"version":"\ud800"}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		policy, err := DecodePolicy(raw)
		if err != nil {
			if !reflect.DeepEqual(policy, Policy{}) {
				t.Fatal("failed decode returned partial policy")
			}
			return
		}
		encoded, err := json.Marshal(policy)
		if err != nil {
			t.Fatal(err)
		}
		again, err := DecodePolicy(encoded)
		if err != nil || !reflect.DeepEqual(again, policy) {
			t.Fatalf("policy bindings changed on round trip: %v", err)
		}
	})
}

// C14: errors cannot leak a decision; accepted result variants must preserve
// grant references and both messages through a wire round trip.
func FuzzDecodeResultPreservesVariant(f *testing.F) {
	for _, seed := range []string{
		`{"version":"1","decision":"allow","grant_ids":["G0","G1"]}`,
		`{"version":"1","decision":"deny","error_code":"DENIED","error_message":"No access","error_message_reason":"Outside boundary"}`,
		`{"version":"1","error_code":"UNAVAILABLE","error_message":"Try later","error_message_reason":"Authority unavailable"}`,
		`{"version":"1","decision":"allow","error_code":"DENIED"}`, `[]`, `{"decision":null}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		result, err := DecodeResult(raw)
		if err != nil {
			if !reflect.DeepEqual(result, Result{}) {
				t.Fatal("error returned a partial authorization result")
			}
			var evaluationError *EvaluationError
			if errors.As(err, &evaluationError) {
				encoded, marshalErr := json.Marshal(evaluationError)
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				_, nextErr := DecodeResult(encoded)
				var again *EvaluationError
				if !errors.As(nextErr, &again) || !reflect.DeepEqual(again, evaluationError) {
					t.Fatal("evaluation error variant changed")
				}
			}
			return
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		again, err := DecodeResult(encoded)
		if err != nil || !reflect.DeepEqual(again, result) {
			t.Fatalf("authorization result changed on round trip: %v", err)
		}
	})
}
