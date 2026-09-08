package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"encoding/json"
	"io"
	"reflect"
)

func grantStatus(ctx context.Context, api application.API, area domain.Area, fixture, verb, id string, out io.Writer) error {
	statusAPI, ok := api.(application.GrantStatusAPI)
	if !ok || nilCapability(statusAPI) {
		return domain.ErrUnsupported
	}
	status := "enabled"
	if verb == "disable" {
		status = "disabled"
	}
	control, err := statusAPI.SetGrantStatus(ctx, area, domain.FixtureContext{Name: fixture}, domain.GrantControl{Version: "1", ID: id, Status: status})
	if err != nil {
		return err
	}
	raw, err := json.Marshal(control)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	_, err = out.Write(raw)
	return err
}

func nilCapability(value any) bool {
	if value == nil {
		return true
	}
	kind := reflect.ValueOf(value).Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflect.ValueOf(value).IsNil()
}
