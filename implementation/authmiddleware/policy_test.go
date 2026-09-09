package authmiddleware

import (
	"reflect"
	"strings"
	"testing"
)

const policyFixture = `{
  "version": "1",
  "method": "GET",
  "path": "/api/v1/{tenant}/{dept}/{cert}",
  "permission": "hrms:employee:certificate::read",
  "inputs": {
    "tenant": {"source": "path", "name": "tenant"},
    "dept": {"source": "path", "name": "dept"},
    "cert": {"source": "path", "name": "cert"}
  }
}`

func TestDecodePolicyCanonicalFixture(t *testing.T) {
	got, err := DecodePolicy([]byte(policyFixture))
	want := Policy{
		Version: "1", Method: "GET", Path: "/api/v1/{tenant}/{dept}/{cert}",
		Permission: "hrms:employee:certificate::read",
		Inputs: map[string]Input{
			"tenant": {Source: SourcePath, Name: "tenant"},
			"dept":   {Source: SourcePath, Name: "dept"},
			"cert":   {Source: SourcePath, Name: "cert"},
		},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("DecodePolicy() = %#v, %v", got, err)
	}
}

func TestDecodePolicyAcceptsExplicitEmptyInputs(t *testing.T) {
	got, err := DecodePolicy([]byte(`{"version":"1","method":"GET","path":"/health","permission":"health::read","inputs":{}}`))
	if err != nil || got.Inputs == nil || len(got.Inputs) != 0 {
		t.Fatalf("explicit empty inputs lost: %#v, %v", got, err)
	}
}

func TestDecodePolicyRejectsUnsupportedOrAmbiguousJSON(t *testing.T) {
	cases := map[string]string{
		"missing version":     strings.Replace(policyFixture, "  \"version\": \"1\",\n", "", 1),
		"unsupported version": strings.Replace(policyFixture, `"version": "1"`, `"version": "2"`, 1),
		"missing policy":      `{}`,
		"null":                `null`,
		"missing inputs":      `{"version":"1","method":"GET","path":"/x","permission":"x::read"}`,
		"null inputs":         `{"version":"1","method":"GET","path":"/x","permission":"x::read","inputs":null}`,
		"unknown root":        strings.Replace(policyFixture, `"version": "1"`, `"version": "1", "relationship": {}`, 1),
		"capitalized root":    strings.Replace(policyFixture, `"version": "1"`, `"Version": "1"`, 1),
		"mixed case root":     strings.Replace(policyFixture, `"version": "1"`, `"version": "1", "Version": "1"`, 1),
		"duplicate root":      strings.Replace(policyFixture, `"version": "1"`, `"version": "1", "version": "1"`, 1),
		"unknown input":       strings.Replace(policyFixture, `"source": "path"`, `"source": "path", "type": "string"`, 1),
		"capitalized source":  strings.Replace(policyFixture, `"source": "path"`, `"Source": "path"`, 1),
		"capitalized name":    strings.Replace(policyFixture, `"name": "tenant"`, `"Name": "tenant"`, 1),
		"duplicate input":     strings.Replace(policyFixture, `"source": "path"`, `"source": "path", "source": "path"`, 1),
		"escaped duplicate":   strings.Replace(policyFixture, `"source": "path"`, `"source": "path", "\u0073ource": "path"`, 1),
		"wrong type":          strings.Replace(policyFixture, `"method": "GET"`, `"method": 1`, 1),
		"wrong input type":    strings.Replace(policyFixture, `{"source": "path", "name": "tenant"}`, `"path.tenant"`, 1),
		"trailing JSON":       policyFixture + `{}`,
		"truncated":           policyFixture[:40],
		"over size limit":     strings.Repeat(" ", (1<<20)+1),
		"over nesting limit":  `{"version":"1","method":"GET","path":"/x","permission":"x::read","inputs":{` + strings.Repeat(`"x":{"source":"body","name":`, 65) + `"x"` + strings.Repeat(`}`, 65) + `}}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := DecodePolicy([]byte(raw))
			if err == nil || !reflect.DeepEqual(got, Policy{}) {
				t.Fatalf("accepted invalid policy: %#v, %v", got, err)
			}
		})
	}
}

func TestPolicyJSONValidationRemainsStrictWhileBusinessJSONAllowsNull(t *testing.T) {
	if err := validateJSON([]byte(`{"x":null}`)); err == nil {
		t.Fatal("canonical JSON accepted null")
	}
	if err := validateJSONAllowNull([]byte(`{"x":null}`)); err != nil {
		t.Fatalf("business JSON rejected null: %v", err)
	}
	if err := validateJSONAllowNull([]byte(`{"x":null,"x":1}`)); err == nil {
		t.Fatal("business JSON accepted a duplicate key")
	}
}

func TestPolicyValidateRejectsUnsupportedDeclarations(t *testing.T) {
	valid := Policy{Version: "1", Method: "PUT", Path: "/api/{tenant}/certificates/{cert}", Permission: "certificate::write", Inputs: map[string]Input{
		"tenant": {Source: SourcePath, Name: "tenant"},
		"cert":   {Source: SourcePath, Name: "cert"},
		"dept":   {Source: SourceBody, Name: "department_id"},
	}}
	cases := map[string]func(*Policy){
		"missing version":       func(p *Policy) { p.Version = "" },
		"unsupported version":   func(p *Policy) { p.Version = "2" },
		"lowercase method":      func(p *Policy) { p.Method = "put" },
		"relative path":         func(p *Policy) { p.Path = "api/{tenant}" },
		"query in path":         func(p *Policy) { p.Path += "?dept=FIN" },
		"empty permission":      func(p *Policy) { p.Permission = "" },
		"missing separator":     func(p *Policy) { p.Permission = "certificate.read" },
		"empty permission verb": func(p *Policy) { p.Permission = "certificate::" },
		"multiple separators":   func(p *Policy) { p.Permission = "hrms::certificate::read" },
		"wildcard permission":   func(p *Policy) { p.Permission = "certificate::*" },
		"multiple permissions":  func(p *Policy) { p.Permission = "read,write" },
		"nil inputs":            func(p *Policy) { p.Inputs = nil },
		"empty local name":      func(p *Policy) { p.Inputs[""] = Input{Source: SourceBody, Name: "x"} },
		"unsupported source":    func(p *Policy) { p.Inputs["dept"] = Input{Source: "query", Name: "dept"} },
		"undeclared path name":  func(p *Policy) { p.Inputs["dept"] = Input{Source: SourcePath, Name: "dept"} },
		"nested body selector":  func(p *Policy) { p.Inputs["dept"] = Input{Source: SourceBody, Name: "employee.department_id"} },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			p := valid
			p.Inputs = make(map[string]Input, len(valid.Inputs))
			for k, v := range valid.Inputs {
				p.Inputs[k] = v
			}
			edit(&p)
			if err := p.Validate(); err == nil {
				t.Fatalf("accepted invalid policy: %#v", p)
			}
		})
	}
}
