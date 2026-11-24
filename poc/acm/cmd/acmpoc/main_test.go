package main

import (
	"testing"

	"github.com/fred66sp/agent-interaction-forms"
)

// A required value that is only whitespace is absent for ACM (require_text);
// AIF must answer with a repairable result and never call the domain.
func TestBlankRequiredIsRepairable(t *testing.T) {
	executor := aif.Executor{Operations: []*aif.Operation{learning(func(string) string { return "" })}}
	response := executor.Handle([]byte(
		`{"aif":"0.1","operation":"learning","input":{"title":"  ","body":"texto"}}`))
	problem, ok := response.Invalid["title"]
	if response.Result != aif.ResultInvalid || !ok || problem.Reason != "obligatorio" {
		t.Fatalf("result=%q invalid=%+v", response.Result, response.Invalid)
	}
}
