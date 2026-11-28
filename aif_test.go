package aif

import (
	"encoding/json"
	"fmt"
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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)

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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"","flag":false,"count":0,"tags":[]}}`)

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
	data, err := json.Marshal(ex.Handle([]byte(`{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","secret":"s3cr3t"}}`)))
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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"c"}}`)

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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x","flag":"yes","bogus":1}}`)

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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"code":"abcd"}}`)
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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"code":"x"}}`)
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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
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
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
	assertResult(t, doc, ResultFailed)
	if got := stringMember(t, doc, "message"); got != "no se pudo escribir" {
		t.Errorf("message = %q", got)
	}
	if got := stringMember(t, doc, "code"); got != "storage_unavailable" {
		t.Errorf("code = %q", got)
	}
	absent(t, doc, "outcome")
}

func TestDomainInvalidIsRepairable(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
		return nil, InvalidError{Fields: map[string]string{"mode": "combinación no permitida"}}
	})
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x","secret":"s3cr3t"}}`)

	assertResult(t, doc, ResultInvalid)
	absent(t, doc, "missing")
	absent(t, doc, "outcome")
	var invalid map[string]FieldProblem
	if err := json.Unmarshal(member(t, doc, "invalid"), &invalid); err != nil {
		t.Fatal(err)
	}
	problem, ok := invalid["mode"]
	if !ok {
		t.Fatalf("invalid no describe mode: %v", invalid)
	}
	if problem.Reason != "combinación no permitida" {
		t.Errorf("reason = %q, want %q", problem.Reason, "combinación no permitida")
	}
	var value string
	if err := json.Unmarshal(problem.Value, &value); err != nil || value != "a" {
		t.Errorf("value = %s, want \"a\"", problem.Value)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(problem.Allowed, want) {
		t.Errorf("allowed = %v, want %v", problem.Allowed, want)
	}
	var collected map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, "collected"), &collected); err != nil {
		t.Fatal(err)
	}
	want := map[string]json.RawMessage{"name": json.RawMessage(`"n"`), "note": json.RawMessage(`"x"`)}
	if !reflect.DeepEqual(collected, want) {
		t.Errorf("collected = %v, want %v", collected, want)
	}
	if calls != 1 {
		t.Errorf("Execute llamadas = %d, want 1", calls)
	}
}

func TestDomainInvalidReportsEveryProblemAndHidesSecrets(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
		return nil, InvalidError{Fields: map[string]string{
			"secret": "formato no válido",
			"mode":   "no disponible",
		}}
	})
	data, err := json.Marshal(ex.Handle([]byte(`{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x","secret":"s3cr3t"}}`)))
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
	assertResult(t, doc, ResultInvalid)
	var invalid map[string]FieldProblem
	if err := json.Unmarshal(member(t, doc, "invalid"), &invalid); err != nil {
		t.Fatal(err)
	}
	if len(invalid) != 2 {
		t.Errorf("invalid = %v, want los dos campos rechazados", invalid)
	}
	if _, ok := invalid["secret"]; !ok {
		t.Errorf("invalid no describe secret: %v", invalid)
	}
	if invalid["secret"].Value != nil {
		t.Errorf("invalid[secret].value = %s, want ausente", invalid["secret"].Value)
	}
	var collected map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, "collected"), &collected); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"mode", "secret"} {
		if _, ok := collected[name]; ok {
			t.Errorf("collected no debe incluir %q: %v", name, collected)
		}
	}
	if calls != 1 {
		t.Errorf("Execute llamadas = %d, want 1", calls)
	}
}

// A domain InvalidError that cannot be rebuilt is a domain defect: the
// response is failed, never a reparable invalid with invented content.
func TestUnusableInvalidErrorFails(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		fields map[string]string
	}{
		{"mapa nulo", `{"name":"n","mode":"a","note":"x"}`, nil},
		{"mapa vacío", `{"name":"n","mode":"a","note":"x"}`, map[string]string{}},
		{"campo desconocido", `{"name":"n","mode":"a","note":"x"}`, map[string]string{"bogus": "no existe"}},
		{"motivo vacío", `{"name":"n","mode":"a","note":"x"}`, map[string]string{"mode": ""}},
		{"motivo en blanco", `{"name":"n","mode":"a","note":"x"}`, map[string]string{"mode": "   "}},
		{"campo ausente", `{"name":"n","mode":"a","note":"x"}`, map[string]string{"flag": "no puede estar ausente"}},
		{"campo nulo", `{"name":"n","mode":"a","note":"x","flag":null}`, map[string]string{"flag": "no puede ser nulo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
				return nil, InvalidError{Fields: tt.fields}
			})
			doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":`+tt.input+`}`)
			assertResult(t, doc, ResultFailed)
			if got := stringMember(t, doc, "message"); got == "" {
				t.Error("failed sin message")
			}
			for _, name := range []string{"invalid", "collected", "missing", "outcome"} {
				absent(t, doc, name)
			}
			if calls != 1 {
				t.Errorf("Execute llamadas = %d, want 1", calls)
			}
		})
	}
}

// A typed-nil error pointer (the classic `var e *DeniedError; return e` bug)
// reaches the executor as a non-nil error with no data: the response is
// failed, and the host process must not panic dereferencing the receiver.
func TestNilDomainErrorPointerFails(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"denied nulo", (*DeniedError)(nil)},
		{"invalid nulo", (*InvalidError)(nil)},
		{"domain nulo", (*DomainError)(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
				return nil, tt.err
			})
			doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
			assertResult(t, doc, ResultFailed)
			if got := stringMember(t, doc, "message"); got == "" {
				t.Error("failed sin message")
			}
			for _, name := range []string{"invalid", "collected", "missing", "outcome"} {
				absent(t, doc, name)
			}
			if calls != 1 {
				t.Errorf("Execute llamadas = %d, want 1", calls)
			}
		})
	}
}

func TestWrappedDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		result string
		err    error
	}{
		{"denied", ResultDenied, fmt.Errorf("dominio: %w", DeniedError{Reason: "sin permiso"})},
		{"failed", ResultFailed, fmt.Errorf("dominio: %w", DomainError{Message: "no se pudo", Code: "storage"})},
		{"invalid", ResultInvalid, fmt.Errorf("dominio: %w", InvalidError{Fields: map[string]string{"mode": "no disponible"}})},
		{"invalid puntero", ResultInvalid, &InvalidError{Fields: map[string]string{"mode": "no disponible"}}},
		{"invalid puntero envuelto", ResultInvalid, wrapPtr{&InvalidError{Fields: map[string]string{"mode": "no disponible"}}}},
		{"denied puntero", ResultDenied, &DeniedError{Reason: "sin permiso"}},
		{"failed puntero", ResultFailed, &DomainError{Message: "no se pudo", Code: "storage"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			ex := testExecutor(t, &calls, func(map[string]json.RawMessage) (json.RawMessage, error) {
				return nil, tt.err
			})
			doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
			assertResult(t, doc, tt.result)
			absent(t, doc, "outcome")
			if calls != 1 {
				t.Errorf("Execute llamadas = %d, want 1", calls)
			}
		})
	}
}

// wrapPtr mimics an adapter that returns a pointer-form library error wrapped
// in its own error, without tripping vet's %w operand rule.
type wrapPtr struct{ err error }

func (w wrapPtr) Error() string { return "dominio: " + w.err.Error() }
func (w wrapPtr) Unwrap() error { return w.err }

// Exhaustive validation: one exchange reports every detected problem, both
// absent and invalid fields, not just the first one.
func TestReportsEveryDetectedProblem(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"test","input":{"name":"n","mode":"c","note":null,"flag":"yes","bogus":1}}`)

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
	for _, name := range []string{"mode", "flag", "bogus"} {
		if _, ok := invalid[name]; !ok {
			t.Errorf("invalid no describe %q: %v", name, invalid)
		}
	}
	if len(invalid) != 3 {
		t.Errorf("invalid = %v, want 3 problemas", invalid)
	}
	var collected map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, "collected"), &collected); err != nil {
		t.Fatal(err)
	}
	want := map[string]json.RawMessage{"name": json.RawMessage(`"n"`)}
	if !reflect.DeepEqual(collected, want) {
		t.Errorf("collected = %v, want %v", collected, want)
	}
	if calls != 0 {
		t.Errorf("Execute llamadas = %d, want 0", calls)
	}
}

func TestUnsupportedVersionDoesNotProcessInput(t *testing.T) {
	for _, version := range []string{"0.1", "1.0"} {
		t.Run(version, func(t *testing.T) {
			calls := 0
			ex := testExecutor(t, &calls, nil)
			doc := handleJSON(t, ex, `{"aif":"`+version+`","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`)
			assertResult(t, doc, ResultUnsupportedVersion)
			if got := stringMember(t, doc, "aif"); got != Version {
				t.Errorf("aif = %q, want %q", got, Version)
			}
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
		})
	}
}

func TestUnknownOperationEchoesIt(t *testing.T) {
	calls := 0
	ex := testExecutor(t, &calls, nil)
	doc := handleJSON(t, ex, `{"aif":"0.2","operation":"nope"}`)
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
	response := ex.Handle([]byte(`{"aif":"0.2"}`))
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
	var request RequestTemplate
	if err := json.Unmarshal(member(t, doc, "request"), &request); err != nil {
		t.Fatal(err)
	}
	if request.AIF != Version || request.Operation != "<id>" || len(request.Input) != 0 {
		t.Errorf("request = %+v, want la plantilla de emisión", request)
	}
	var operations []OperationInfo
	if err := json.Unmarshal(member(t, doc, "operations"), &operations); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 1 || operations[0].ID != "test" {
		t.Errorf("operations = %+v", operations)
	}
}

func TestMalformedRequests(t *testing.T) {
	deep := `{"aif":"0.2","operation":"test","input":{"tags":` +
		strings.Repeat("[", MaxDepth+1) + strings.Repeat("]", MaxDepth+1) + `}}`
	large := strings.Repeat(" ", MaxInputBytes+1) + `{"aif":"0.2","operation":"test"}`

	tests := []struct {
		name    string
		request string
		reason  string
	}{
		{"vacía", "", malformedEmpty},
		{"solo espacios", "   ", malformedEmpty},
		{"json inválido", `{"aif":`, malformedSyntax},
		{"contenido tras el objeto", `{"aif":"0.2","operation":"test"}{}`, malformedSyntax},
		{"no es objeto", `[1,2,3]`, malformedNotObject},
		{"sin aif", `{"operation":"test"}`, malformedNoAIF},
		{"aif nula", `{"aif":null,"operation":"test"}`, malformedNoAIF},
		{"aif no textual", `{"aif":1,"operation":"test"}`, malformedAIFType},
		{"claves duplicadas", `{"aif":"0.2","aif":"0.2","operation":"test"}`, malformedDuplicate},
		{"input no objeto", `{"aif":"0.2","operation":"test","input":[]}`, malformedInputType},
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
		`{"aif":"0.2","operation":"test","input":{"name":"n","mode":"a","note":"x"}}`,
		`{"aif":"0.2","operation":"test","input":{"mode":"a"}}`,
		`{"aif":"0.2","operation":"nope"}`,
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
