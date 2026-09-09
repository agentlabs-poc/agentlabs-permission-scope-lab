package authmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type IdentitySource interface {
	Establish(context.Context, *http.Request) (RequestContext, error)
}

type InputValues map[string]json.RawMessage

type BoundOperation struct {
	Material Material
	Execute  func(context.Context, http.ResponseWriter)
}

type Binder func(context.Context, RequestContext, InputValues, map[string]json.RawMessage) (BoundOperation, error)

type FailureHandler func(http.ResponseWriter, *http.Request, Result, error)

func Wrap(policy Policy, identities IdentitySource, evaluator *Evaluator, bind Binder, fail FailureHandler) (http.Handler, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if nilInterface(identities) || evaluator == nil || nilInterface(evaluator.source) || nilInterface(evaluator.clock) || nilInterface(bind) || nilInterface(fail) {
		return nil, errors.New("identity source, initialized evaluator, binder, and failure handler are required")
	}
	policy.Inputs = cloneInputs(policy.Inputs)

	mux := http.NewServeMux()
	handler := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		handleHTTP(policy, identities, evaluator, bind, fail, w, request)
	})
	if err := registerRoute(mux, policy.Method+" "+policy.Path, handler); err != nil {
		return nil, err
	}
	return mux, nil
}

func cloneInputs(inputs map[string]Input) map[string]Input {
	cloned := make(map[string]Input, len(inputs))
	for name, input := range inputs {
		cloned[name] = input
	}
	return cloned
}

func registerRoute(mux *http.ServeMux, pattern string, handler http.Handler) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("invalid policy route: %v", recovered)
		}
	}()
	mux.Handle(pattern, handler)
	return nil
}

func handleHTTP(policy Policy, identities IdentitySource, evaluator *Evaluator, bind Binder, fail FailureHandler, w http.ResponseWriter, request *http.Request) {
	failure := func(result Result, err error) { fail(w, request, result, err) }
	ctx := request.Context()
	if err := ctx.Err(); err != nil {
		failure(Result{}, err)
		return
	}
	if request.Method != policy.Method {
		failure(Result{}, errors.New("request method does not match policy"))
		return
	}
	identityRequest := request.Clone(ctx)
	identityRequest.Body = http.NoBody
	requestContext, err := identities.Establish(ctx, identityRequest)
	if err != nil {
		failure(Result{}, err)
		return
	}
	if err := validateRequest(ctx, Request{Context: requestContext, Permission: policy.Permission, Material: Material{}}); err != nil {
		failure(Result{}, err)
		return
	}
	if tenant := request.PathValue("tenant"); tenant != "" && tenant != requestContext.Area.TenantID {
		failure(Result{}, errors.New("path tenant does not match trusted area"))
		return
	}
	if application := request.PathValue("application"); application != "" && application != requestContext.Area.ApplicationID {
		failure(Result{}, errors.New("path application does not match trusted area"))
		return
	}
	body, err := decodeBusinessBody(request.Body)
	if err != nil {
		failure(Result{}, err)
		return
	}
	values, err := selectInputs(policy.Inputs, request, body)
	if err != nil {
		failure(Result{}, err)
		return
	}
	operation, err := bind(ctx, requestContext, values, body)
	if err != nil {
		failure(Result{}, err)
		return
	}
	if operation.Execute == nil {
		failure(Result{}, errors.New("binder returned no effect"))
		return
	}
	result, err := evaluator.Evaluate(ctx, Request{Context: requestContext, Permission: policy.Permission, Material: operation.Material})
	if err != nil {
		failure(Result{}, err)
		return
	}
	if err := result.Validate(); err != nil {
		failure(Result{}, err)
		return
	}
	if result.Decision != Allow {
		failure(result, nil)
		return
	}
	if err := ctx.Err(); err != nil {
		failure(Result{}, err)
		return
	}
	operation.Execute(ctx, w)
}

func decodeBusinessBody(body io.Reader) (map[string]json.RawMessage, error) {
	if body == nil {
		return nil, nil
	}
	raw, err := io.ReadAll(io.LimitReader(body, maxJSONBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw) > maxJSONBytes {
		return nil, errors.New("JSON exceeds 1 MiB")
	}
	if err := validateJSONAllowNull(raw); err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, errors.New("request body must be a JSON object")
	}
	return fields, nil
}

func selectInputs(inputs map[string]Input, request *http.Request, body map[string]json.RawMessage) (InputValues, error) {
	values := make(InputValues, len(inputs))
	for local, input := range inputs {
		switch input.Source {
		case SourcePath:
			value := request.PathValue(input.Name)
			if value == "" {
				return nil, fmt.Errorf("missing path input %q", input.Name)
			}
			values[local], _ = json.Marshal(value)
		case SourceBody:
			value, ok := body[input.Name]
			if !ok {
				return nil, fmt.Errorf("missing body input %q", input.Name)
			}
			values[local] = value
		}
	}
	return values, nil
}
