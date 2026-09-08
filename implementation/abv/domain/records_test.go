package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Wrong JSON tags or adding recipient/context to content breaks the approved core shape.
func TestCoreRecordJSON(t *testing.T) {
	cases := []struct {
		value any
		want  string
	}{
		{GrantControl{Version: "1", ID: "G1", Status: "enabled"}, `{"version":"1","id":"G1","status":"enabled"}`},
		{GrantContent{Version: "1", GrantID: "G1", Revision: 2, ParentGrantID: "G0", Permissions: []string{"hrms:employee:certificate::read"}, Scope: map[string]string{}}, `{"version":"1","grant_id":"G1","revision":2,"parent_grant_id":"G0","permissions":["hrms:employee:certificate::read"],"scope":{}}`},
		{GrantContent{Version: "1", GrantID: "G2", Revision: 1, ParentGrantID: "G1", RoleID: "reader", RoleRevision: 1, Scope: map[string]string{"dept": "FIN"}}, `{"version":"1","grant_id":"G2","revision":1,"parent_grant_id":"G1","role_id":"reader","role_revision":1,"scope":{"dept":"FIN"}}`},
		{Assignment{Version: "1", ID: "A1", GrantID: "G1", GrantRevision: 2, Recipient: Recipient{Type: "group", ID: "Team1"}, Status: "enabled"}, `{"version":"1","id":"A1","grant_id":"G1","grant_revision":2,"recipient":{"type":"group","id":"Team1"},"status":"enabled"}`},
		{Identity{Version: "1", Actor: Actor{Type: "user", ID: "maya"}, HumanID: "maya"}, `{"version":"1","actor":{"type":"user","id":"maya"},"human_id":"maya"}`},
	}
	for _, tc := range cases {
		got, err := json.Marshal(tc.value)
		if err != nil {
			t.Fatal(err)
		}
		var actual, expected any
		if err := json.Unmarshal(got, &actual); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(tc.want), &expected); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("wrong core shape: %s; want %s", got, tc.want)
		}
	}
}
