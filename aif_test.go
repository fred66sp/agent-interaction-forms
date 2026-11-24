package aif

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func testExecutor(t *testing.T, calls *int, execute func(map[string]json.RawMessage) (json.RawMessage, error)) *Executor {
	t.Helper()
	if execute == nil {
		execute = func(map[string]json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"id":"obs_1"}`), nil
		}
	}
	wrapped := func(input map[string]json.RawMessage) (json.RawMessage, error) {
		*calls++
		return execute(input)
	}
	return &Executor{Operations: []*Operation{{
		ID:          "test",
		Description: "operación de prueba",
		Fields: []Field{
			{Name: "name", Required: true, Type: "string", Description: "nombre"},
			{Name: "mode", Required: true, Type: "string", Allowed: []string{"a", "b"}},
			{Name: "note", Required: true, Type: "string"},
			{Name: "flag", Type: "boolean"},
			{Name: "count", Type: "number"},
			{Name: "tags", Type: "array"},
			{Name: "secret", Type: "string", Sensitive: true},
		},
		Execute: wrapped,
	}}}
}

func handleJSON(t *testing.T, ex *Executor, request string) map[string]json.RawMessage {
	t.Helper()
	data, err := json.Marshal(ex.Handle([]byte(request)))
	if err != nil {
		t.Fatalf("no se pudo serializar la respuesta: %v", err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("la respuesta no es un objeto JSON: %v", err)
	}
	return doc
}

func member(t *testing.T, doc map[string]json.RawMessage, name string) json.RawMessage {
	t.Helper()
	raw, ok := doc[name]
	if !ok {
		t.Fatalf("falta el miembro %q en %v", name, doc)
	}
	return raw
}

func stringMember(t *testing.T, doc map[string]json.RawMessage, name string) string {
	t.Helper()
	var s string
	if err := json.Unmarshal(member(t, doc, name), &s); err != nil {
		t.Fatalf("%q no es una cadena: %v", name, err)
	}
	return s
}

func assertResult(t *testing.T, doc map[string]json.RawMessage, want string) {
	t.Helper()
	if got := stringMember(t, doc, "result"); got != want {
		t.Errorf("result = %q, want %q", got, want)
	}
}

func absent(t *testing.T, doc map[string]json.RawMessage, name string) {
	t.Helper()
	if _, ok := doc[name]; ok {
		t.Errorf("el miembro %q no debe aparecer: %v", name, doc)
	}
}

func TestExecuted(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)

	assertResult(t, doc, ResultExecuted)
	if got := stringMember(t, doc, "operation"); got != "test" {
		t.Errorf("operation = %q, want %q", got, "test")
	}
	var outcome map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, "outcome"), &outcome); err != nil {
		t.Fatalf("outcome no es un objeto: %v", err)
	}
	if got := stringMember(t, outcome, "id"); got != "obs_1" {
		t.Errorf("outcome.id = %q, want %q", got, "obs_1")
	}
	absent(t, doc, "missing")
	absent(t, doc, "invalid")
	if calls != 1 {
		t.Errorf("Execute llamadas = %d, want 1", calls)
	}
}

func TestInputNeededKeepsFalsyValues(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"name":"","flag":false,"count":0,"tags":[]}}`)

	assertResult(t, doc, ResultInputNeeded)
	var missing []string
	if err := json.Unmarshal(member(t, doc, "missing"), &missing); err != nil {
		t.Fatalf("missing no es una matriz: %v", err)
	}
	if want := []string{"mode", "note"}; !reflect.DeepEqual(missing, want) {
		t.Errorf("missing = %v, want %v", missing, want)
	}
	var collected map[string]any
	if err := json.Unmarshal(member(t, doc, "collected"), &collected); err != nil {
		t.Fatalf("collected no es un objeto: %v", err)
	}
	want := map[string]any{"name": "", "flag": false, "count": float64(0), "tags": []any{}}
	if !reflect.DeepEqual(collected, want) {
		t.Errorf("collected = %v, want %v", collected, want)
	}
	var fields map[string]FieldDescriptor
	if err := json.Unmarshal(member(t, doc, "fields"), &fields); err != nil {
		t.Fatalf("fields no es un objeto: %v", err)
	}
	for _, name := range []string{"mode", "note"} {
		if !fields[name].Required {
			t.Errorf("fields[%q].required = false, want true", name)
		}
	}
	if calls != 0 {
		t.Errorf("Execute llamadas = %d, want 0", calls)
	}
}

func TestSensitiveValueNeverReturned(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	data, err := json.Marshal(ex.Handle([]byte(`{"aif":"0.1","operation":"test","input":{"name":"n","mode":"a","secret":"s3cr3t"}}`)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "s3cr3t") {
		t.Errorf("el valor sensible aparece en la respuesta: %s", data)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	assertResult(t, doc, ResultInputNeeded)
	var collected map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, "collected"), &collected); err != nil {
		t.Fatal(err)
	}
	if _, ok := collected["secret"]; ok {
		t.Errorf("collected no debe incluir valores sensibles: %v", collected)
	}
}

func TestInvalidTakesPrecedenceOverMissing(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"name":"n","mode":"c"}}`)

	assertResult(t, doc, ResultInvalid)
	var missing []string
	if err := json.Unmarshal(member(t, doc, "missing"), &missing); err != nil {
		t.Fatal(err)
	}
	if want := []string{"note"}; !reflect.DeepEqual(missing, want) {
		t.Errorf("missing = %v, want %v", missing, want)
	}
	var invalid map[string]FieldProblem
	if err := json.Unmarshal(member(t, doc, "invalid"), &invalid); err != nil {
		t.Fatal(err)
	}
	problem, ok := invalid["mode"]
	if !ok {
		t.Fatalf("invalid no describe mode: %v", invalid)
	}
	if problem.Reason != "valor no permitido" {
		t.Errorf("reason = %q, want %q", problem.Reason, "valor no permitido")
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(problem.Allowed, want) {
		t.Errorf("allowed = %v, want %v", problem.Allowed, want)
	}
	var value string
	if err := json.Unmarshal(problem.Value, &value); err != nil || value != "c" {
		t.Errorf("value = %s, want \"c\"", problem.Value)
	}
	if calls != 0 {
		t.Errorf("Execute llamadas = %d, want 0", calls)
	}
}

func TestOptionalInvalidAndUnknownMember(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"name":"n","mode":"a","note":"x","flag":"yes","bogus":1}}`)

	assertResult(t, doc, ResultInvalid)
	var invalid map[string]FieldProblem
	if err := json.Unmarshal(member(t, doc, "invalid"), &invalid); err != nil {
		t.Fatal(err)
	}
	if problem, ok := invalid["flag"]; !ok || !strings.Contains(problem.Reason, "boolean") {
		t.Errorf("invalid[flag] = %+v, want a type reason", problem)
	}
	if problem, ok := invalid["bogus"]; !ok || !strings.Contains(problem.Reason, "bogus") {
		t.Errorf("invalid[bogus] = %+v, want a reason naming the member", problem)
	}
	var collected map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, "collected"), &collected); err != nil {
		t.Fatal(err)
	}
	if _, ok := collected["bogus"]; ok {
		t.Errorf("collected no debe incluir el miembro desconocido: %v", collected)
	}
	if calls != 0 {
		t.Errorf("Execute llamadas = %d, want 0", calls)
	}
}

func TestCustomRulePublishesSchema(t *testing.T) {
	ex := &Executor{Operations: []*Operation{{
		ID: "test",
		Fields: []Field{{
			Name:     "code",
			Required: true,
			Type:     "string",
			Schema:   json.RawMessage(`{"maxLength":3}`),
			Validate: func(v any) (string, bool) {
				if s, ok := v.(string); ok && len(s) > 3 {
					return "supera la longitud máxima", false
				}
				return "", true
			},
		}},
		Execute: func(map[string]json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	}}}
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"code":"abcd"}}`)
	assertResult(t, doc, ResultInvalid)
	var invalid map[string]FieldProblem
	if err := json.Unmarshal(member(t, doc, "invalid"), &invalid); err != nil {
		t.Fatal(err)
	}
	problem := invalid["code"]
	if problem.Reason != "supera la longitud máxima" {
		t.Errorf("reason = %q", problem.Reason)
	}
	if string(problem.Schema) != `{"maxLength":3}` {
		t.Errorf("schema = %s", problem.Schema)
	}
	var fields map[string]FieldDescriptor
	if err := json.Unmarshal(member(t, doc, "fields"), &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["code"].Schema) != `{"maxLength":3}` {
		t.Errorf("fields[code].schema = %s", fields["code"].Schema)
	}
}

// An integrator hook that rejects without a reason still yields a
// deterministic diagnostic: §3.2 requires reason in every invalid.
func TestHookWithoutReasonGetsDefaultOne(t *testing.T) {
	ex := &Executor{Operations: []*Operation{{
		ID: "test",
		Fields: []Field{{
			Name:     "code",
			Required: true,
			Validate: func(v any) (string, bool) { return "", false },
		}},
		Execute: func(map[string]json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	}}}
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"code":"x"}}`)
	assertResult(t, doc, ResultInvalid)
	var invalid map[string]FieldProblem
	if err := json.Unmarshal(member(t, doc, "invalid"), &invalid); err != nil {
		t.Fatal(err)
	}
	if invalid["code"].Reason == "" {
		t.Error("reason vacía en invalid, 3.2 exige una")
	}
}

func TestDenied(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
		return nil, DeniedError{Reason: "sin permiso"}
	})
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
	assertResult(t, doc, ResultDenied)
	if got := stringMember(t, doc, "reason"); got != "sin permiso" {
		t.Errorf("reason = %q", got)
	}
	for _, name := range []string{"missing", "invalid", "fields", "collected"} {
		absent(t, doc, name)
	}
}

func TestFailed(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
		return nil, DomainError{Message: "no se pudo escribir", Code: "storage_unavailable"}
	})
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
	assertResult(t, doc, ResultFailed)
	if got := stringMember(t, doc, "message"); got != "no se pudo escribir" {
		t.Errorf("message = %q", got)
	}
	if got := stringMember(t, doc, "code"); got != "storage_unavailable" {
		t.Errorf("code = %q", got)
	}
	absent(t, doc, "outcome")
}

func TestUnsupportedVersionDoesNotProcessInput(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"1.0","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
	assertResult(t, doc, ResultUnsupportedVersion)
	var supported []string
	if err := json.Unmarshal(member(t, doc, "supported"), &supported); err != nil {
		t.Fatal(err)
	}
	if want := []string{Version}; !reflect.DeepEqual(supported, want) {
		t.Errorf("supported = %v, want %v", supported, want)
	}
	absent(t, doc, "operation")
	absent(t, doc, "outcome")
	if calls != 0 {
		t.Errorf("Execute llamadas = %d, want 0", calls)
	}
}

func TestUnknownOperationEchoesIt(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.1","operation":"nope"}`)
	assertResult(t, doc, ResultUnknownOperation)
	if got := stringMember(t, doc, "operation"); got != "nope" {
		t.Errorf("operation = %q, want %q", got, "nope")
	}
	var known []string
	if err := json.Unmarshal(member(t, doc, "known_operations"), &known); err != nil {
		t.Fatal(err)
	}
	if want := []string{"test"}; !reflect.DeepEqual(known, want) {
		t.Errorf("known_operations = %v, want %v", known, want)
	}
	if calls != 0 {
		t.Errorf("Execute llamadas = %d, want 0", calls)
	}
}

func TestDiscovery(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	response := ex.Handle([]byte(`{"aif":"0.1"}`))
	if response.Discovery == nil {
		t.Fatal("se esperaba una respuesta de descubrimiento")
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	absent(t, doc, "result")
	absent(t, doc, "operation")
	var operations []OperationInfo
	if err := json.Unmarshal(member(t, doc, "operations"), &operations); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 1 || operations[0].ID != "test" {
		t.Errorf("operations = %+v", operations)
	}
}

func TestMalformedRequests(t *testing.T) {
	deep := `{"aif":"0.1","operation":"test","input":{"tags":` +
		strings.Repeat("[", MaxDepth+1) + strings.Repeat("]", MaxDepth+1) + `}}`
	large := strings.Repeat(" ", MaxInputBytes+1) + `{"aif":"0.1","operation":"test"}`

	tests := []struct {
		name    string
		request string
		reason  string
	}{
		{"vacía", "", malformedEmpty},
		{"solo espacios", "   ", malformedEmpty},
		{"json inválido", `{"aif":`, malformedSyntax},
		{"contenido tras el objeto", `{"aif":"0.1","operation":"test"}{}`, malformedSyntax},
		{"no es objeto", `[1,2,3]`, malformedNotObject},
		{"sin aif", `{"operation":"test"}`, malformedNoAIF},
		{"aif nula", `{"aif":null,"operation":"test"}`, malformedNoAIF},
		{"aif no textual", `{"aif":1,"operation":"test"}`, malformedAIFType},
		{"claves duplicadas", `{"aif":"0.1","aif":"0.1","operation":"test"}`, malformedDuplicate},
		{"input no objeto", `{"aif":"0.1","operation":"test","input":[]}`, malformedInputType},
		{"demasiado profundo", deep, malformedTooDeep},
		{"demasiado grande", large, malformedTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			ex := testExecutor(t, &calls, nil)
			doc := handleJSON(t, ex, tt.request)
			assertResult(t, doc, ResultMalformedRequest)
			if got := stringMember(t, doc, "reason"); got != tt.reason {
				t.Errorf("reason = %q, want %q", got, tt.reason)
			}
			absent(t, doc, "operation")
			absent(t, doc, "outcome")
			if calls != 0 {
				t.Errorf("Execute llamadas = %d, want 0", calls)
			}
		})
	}
}

func TestDeterministic(t *testing.T) {
	requests := []string{
		`{"aif":"0.1","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`,
		`{"aif":"0.1","operation":"test","input":{"mode":"a"}}`,
		`{"aif":"0.1","operation":"nope"}`,
	}
	for _, request := range requests {
		t.Run(request, func(t *testing.T) {
			calls := 0
			ex := testExecutor(t, &calls, nil)
			first, err := json.Marshal(ex.Handle([]byte(request)))
			if err != nil {
				t.Fatal(err)
			}
			second, err := json.Marshal(ex.Handle([]byte(request)))
			if err != nil {
				t.Fatal(err)
			}
			if string(first) != string(second) {
				t.Errorf("la respuesta cambió entre ejecuciones:\nprimera: %s\nsegunda: %s", first, second)
			}
		})
	}
}
