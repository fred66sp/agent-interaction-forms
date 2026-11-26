package main

import (
	"encoding/json"
	"testing"

	"github.com/fred66sp/agent-interaction-forms"
)

// A required value that is only whitespace is absent for ACM (require_text);
// AIF must answer with a repairable result and never call the domain.
func TestBlankRequiredIsRepairable(t *testing.T) {
	executor := aif.Executor{Operations: []*aif.Operation{learning(func(string) string { return "" })}}
	response := executor.Handle([]byte(
		`{"aif":"0.2","operation":"learning","input":{"title":"  ","body":"texto"}}`))
	problem, ok := response.Invalid["title"]
	if response.Result != aif.ResultInvalid || !ok || problem.Reason != "obligatorio" {
		t.Fatalf("result=%q invalid=%+v", response.Result, response.Invalid)
	}
}

// ACM validates before writing; its field reasons become a repairable AIF
// response without echoing hidden values or the fields it rejected.
func TestDomainInvalidIsRepairable(t *testing.T) {
	operation := learning(func(string) string { return "" })
	operation.Execute = func(map[string]json.RawMessage) (json.RawMessage, error) {
		return nil, acmFailure(acmReply{
			Error:   "E_INVALID",
			Message: "ACM rechazó los campos",
			Invalid: map[string]string{"title": "ya existe"},
		})
	}
	executor := aif.Executor{Operations: []*aif.Operation{operation}}
	response := executor.Handle([]byte(
		`{"aif":"0.2","operation":"learning","input":{"title":"T","body":"texto"}}`))

	if response.Result != aif.ResultInvalid {
		t.Fatalf("result = %q, want %q", response.Result, aif.ResultInvalid)
	}
	problem, ok := response.Invalid["title"]
	if !ok || problem.Reason != "ya existe" {
		t.Errorf("invalid = %+v, want title con el motivo de ACM", response.Invalid)
	}
	if _, ok := response.Collected["title"]; ok {
		t.Errorf("collected no debe incluir el campo rechazado: %v", response.Collected)
	}
	if _, ok := response.Collected["body"]; !ok {
		t.Errorf("collected debe conservar body: %v", response.Collected)
	}
}
