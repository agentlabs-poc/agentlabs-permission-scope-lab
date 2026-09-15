package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

type ownerRecordAPI struct {
	apiSpy
	filter      domain.OwnerFilter
	team, human string
	verb        string
	calls       int
}

func (a *ownerRecordAPI) AddOwner(_ context.Context, area domain.Area, fixture domain.FixtureContext, teamID, humanID string) error {
	a.area, a.fixture, a.team, a.human, a.verb, a.calls = area, fixture, teamID, humanID, "add", a.calls+1
	return nil
}

func (a *ownerRecordAPI) RemoveOwner(_ context.Context, area domain.Area, fixture domain.FixtureContext, teamID, humanID string) error {
	a.area, a.fixture, a.team, a.human, a.verb, a.calls = area, fixture, teamID, humanID, "remove", a.calls+1
	return nil
}

func (a *ownerRecordAPI) ListOwners(_ context.Context, area domain.Area, fixture domain.FixtureContext, filter domain.OwnerFilter) (domain.OwnerPage, error) {
	a.area, a.fixture, a.filter, a.calls = area, fixture, filter, a.calls+1
	return domain.OwnerPage{Owners: []domain.Ownership{{TeamID: "fibggi2juubk", HumanID: "fi7io4lvjqio"}}, Total: 1}, nil
}

func runOwners(t *testing.T, api *ownerRecordAPI, args ...string) (int, string, string) {
	t.Helper()
	connector := &connectorSpy{api: api}
	var out, diag bytes.Buffer
	code := Run(t.Context(), append(args, grantArea...), strings.NewReader(""), &out, &diag, connector.connect, nil)
	return code, out.String(), diag.String()
}

func TestOwnersAddAndRemoveForwardThePair(t *testing.T) {
	api := &ownerRecordAPI{}
	code, out, diag := runOwners(t, api, "owners", "add", "--team", "fibggi2juubk", "--human", "fi7io4lvjqio")
	if code != 0 || api.verb != "add" || api.team != "fibggi2juubk" || api.human != "fi7io4lvjqio" {
		t.Fatalf("exit=%d verb=%q team=%q human=%q stderr=%q", code, api.verb, api.team, api.human, diag)
	}
	// The line reads as the relationship, because that is what the record is.
	if !strings.Contains(out, "fi7io4lvjqio owns fibggi2juubk") {
		t.Fatalf("stdout=%q", out)
	}
	if code, out, _ := runOwners(t, api, "owners", "remove", "--team", "fibggi2juubk", "--human", "fi7io4lvjqio"); code != 0 || api.verb != "remove" || !strings.Contains(out, "no longer owns") {
		t.Fatalf("exit=%d verb=%q stdout=%q", code, api.verb, out)
	}
}

func TestOwnersListForwardsEitherDirection(t *testing.T) {
	api := &ownerRecordAPI{}
	if code, out, _ := runOwners(t, api, "owners", "list", "--team", "fibggi2juubk"); code != 0 || !strings.Contains(out, "total  1") {
		t.Fatalf("exit=%d stdout=%q", code, out)
	}
	if api.filter.TeamID != "fibggi2juubk" || api.filter.HumanID != "" {
		t.Fatalf("by team gave filter %#v", api.filter)
	}
	if code, _, _ := runOwners(t, api, "owners", "list", "--human", "fi7io4lvjqio"); code != 0 {
		t.Fatal("by human refused")
	}
	if api.filter.HumanID != "fi7io4lvjqio" || api.filter.TeamID != "" {
		t.Fatalf("by human gave filter %#v", api.filter)
	}
}

func TestOwnersRefusesMalformedInvocationsBeforeConnecting(t *testing.T) {
	cases := map[string][]string{
		"no verb":           {"owners"},
		"unknown verb":      {"owners", "transfer"},
		"add without team":  {"owners", "add", "--human", "fi7io4lvjqio"},
		"add without human": {"owners", "add", "--team", "fibggi2juubk"},
		"list unfiltered":   {"owners", "list"},
		"list both filters": {"owners", "list", "--team", "fibggi2juubk", "--human", "fi7io4lvjqio"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			api := &ownerRecordAPI{}
			code, out, _ := runOwners(t, api, args...)
			if code != 2 || api.calls != 0 || out != "" {
				t.Fatalf("exit=%d calls=%d stdout=%q, want a refusal before connecting", code, api.calls, out)
			}
		})
	}
}
