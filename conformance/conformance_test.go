package conformance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "aifdemo-conformance-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo crear el directorio temporal: %v\n", err)
		os.Exit(1)
	}
	binary = filepath.Join(dir, "aifdemo")
	build := exec.Command("go", "build", "-o", binary, "./cmd/aifdemo")
	build.Dir = repoRoot()
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo compilar aifdemo: %v\n%s", err, output)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file))
}

type fixture struct {
	ID            string        `json:"id"`
	Description   string        `json:"description"`
	Request       string        `json:"request"`
	Requests      []requestCase `json:"requests"`
	Preconditions preconditions `json:"preconditions"`
	Assertions    expectations  `json:"assertions"`
	FollowUp      *followUp     `json:"follow_up"`
}

type requestCase struct {
	Text   string `json:"text"`
	Reason string `json:"reason"`
}

type preconditions struct {
	Env   map[string]string `json:"env"`
	Store []json.RawMessage `json:"store"`
}

type followUp struct {
	Request    string       `json:"request"`
	Assertions expectations `json:"assertions"`
}

type expectations struct {
	ExitCode         *int                      `json:"exit_code"`
	Result           string                    `json:"result"`
	Discovery        bool                      `json:"discovery"`
	EchoOperation    *bool                     `json:"echo_operation"`
	Operation        string                    `json:"operation"`
	Reason           string                    `json:"reason"`
	Message          string                    `json:"message"`
	Code             string                    `json:"code"`
	Missing          []string                  `json:"missing"`
	Invalid          map[string]map[string]any `json:"invalid"`
	Fields           map[string]map[string]any `json:"fields"`
	Collected        map[string]any            `json:"collected"`
	Supported        []string                  `json:"supported"`
	KnownOperations  []string                  `json:"known_operations"`
	OutcomeMembers   []string                  `json:"outcome_members"`
	OperationsMustID []string                  `json:"operations_must_id"`
	Must             []string                  `json:"must"`
	MustNot          []string                  `json:"must_not"`
	AbsentSubstrings []string                  `json:"absent_substrings"`
	Writes           int                       `json:"writes"`
	Repeat           bool                      `json:"repeat"`
}

var requiredIDs = []string{
	"CJ-01", "CJ-02", "CJ-03", "CJ-04", "CJ-05", "CJ-06a", "CJ-06b",
	"CJ-07", "CJ-08", "CJ-09", "CJ-10", "CJ-11", "CJ-12", "CJ-13", "CJ-14",
}

func TestConformance(t *testing.T) {
	paths, err := filepath.Glob("fixtures/CJ-*.json")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	seen := map[string]bool{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", path, err)
		}
		var f fixture
		if err := json.Unmarshal(raw, &f); err != nil {
			t.Fatalf("%s no es un fixture válido: %v", path, err)
		}
		if f.ID == "" {
			t.Fatalf("%s no declara id", path)
		}
		if seen[f.ID] {
			t.Errorf("el caso %s aparece en más de un fixture", f.ID)
		}
		seen[f.ID] = true
		t.Run(f.ID, func(t *testing.T) {
			runFixture(t, f)
		})
	}
	for _, id := range requiredIDs {
		if !seen[id] {
			t.Errorf("falta el fixture del caso %s", id)
		}
	}
}

func runFixture(t *testing.T, f fixture) {
	t.Helper()
	store := filepath.Join(t.TempDir(), "store.jsonl")
	seedStore(t, store, f.Preconditions.Store)
	env := baseEnv(store, f.Preconditions.Env)

	cases := f.Requests
	if len(cases) == 0 {
		cases = []requestCase{{Text: f.Request}}
	}
	for i, c := range cases {
		if c.Text == "" {
			t.Fatalf("la petición %d está vacía", i)
		}
		reason := f.Assertions.Reason
		if c.Reason != "" {
			reason = c.Reason
		}
		first := assertRun(t, f.Assertions, c.Text, reason, store, env)
		if f.Assertions.Repeat {
			second := assertRun(t, f.Assertions, c.Text, reason, store, env)
			firstDecision := decisionMembers(first)
			secondDecision := decisionMembers(second)
			if !reflect.DeepEqual(firstDecision, secondDecision) {
				t.Errorf("la repetición no es determinista:\nprimera: %v\nsegunda: %v", firstDecision, secondDecision)
			}
		}
		if f.FollowUp != nil {
			assertRun(t, f.FollowUp.Assertions, f.FollowUp.Request, f.FollowUp.Assertions.Reason, store, env)
		}
	}
}

// assertRun executes one request, applies the preconditions already present in
// the environment and checks the fixture expectations.
func assertRun(t *testing.T, want expectations, request, reason, store string, env []string) map[string]json.RawMessage {
	t.Helper()
	before := countLines(t, store)
	stdout, stderr, code := execute(t, request, env)
	after := countLines(t, store)

	if want.ExitCode != nil && code != *want.ExitCode {
		t.Errorf("código de salida = %d, want %d\nstdout: %s\nstderr: %s", code, *want.ExitCode, stdout, stderr)
	}
	if got := after - before; got != want.Writes {
		t.Errorf("escrituras en el registro = %d, want %d", got, want.Writes)
	}
	resp := parseSingleObject(t, stdout)

	if want.Result != "" {
		assertString(t, resp, "result", want.Result)
	}
	if want.Discovery {
		if _, ok := resp["result"]; ok {
			t.Errorf("el descubrimiento no debe llevar result: %v", resp)
		}
		assertDiscovery(t, resp, want.OperationsMustID)
	}
	if want.EchoOperation != nil {
		raw, ok := resp["operation"]
		if *want.EchoOperation && !ok {
			t.Errorf("falta el eco de operation")
		}
		if !*want.EchoOperation && ok {
			t.Errorf("operation no debe aparecer en la respuesta: %s", raw)
		}
		if ok && want.Operation != "" {
			assertRawString(t, raw, want.Operation, "operation")
		}
	}
	if reason != "" {
		assertString(t, resp, "reason", reason)
	}
	if want.Message != "" {
		assertString(t, resp, "message", want.Message)
	}
	if want.Code != "" {
		assertString(t, resp, "code", want.Code)
	}
	if want.Missing != nil {
		assertStringSlice(t, resp, "missing", want.Missing)
	}
	if want.Supported != nil {
		assertStringSlice(t, resp, "supported", want.Supported)
	}
	if want.KnownOperations != nil {
		assertStringSlice(t, resp, "known_operations", want.KnownOperations)
	}
	if want.Invalid != nil {
		assertSubsetObject(t, resp, "invalid", want.Invalid)
	}
	if want.Fields != nil {
		assertSubsetObject(t, resp, "fields", want.Fields)
	}
	if want.Collected != nil {
		assertExactObject(t, resp, "collected", want.Collected)
	}
	if want.OutcomeMembers != nil {
		var outcome map[string]json.RawMessage
		if err := json.Unmarshal(member(t, resp, "outcome"), &outcome); err != nil {
			t.Fatalf("outcome no es un objeto: %v", err)
		}
		for _, name := range want.OutcomeMembers {
			if _, ok := outcome[name]; !ok {
				t.Errorf("falta outcome.%s en %v", name, outcome)
			}
		}
	}
	for _, name := range want.Must {
		if _, ok := resp[name]; !ok {
			t.Errorf("falta el miembro obligatorio %q", name)
		}
	}
	for _, name := range want.MustNot {
		if _, ok := resp[name]; ok {
			t.Errorf("el miembro %q no debe aparecer", name)
		}
	}
	for _, secret := range want.AbsentSubstrings {
		if strings.Contains(stdout, secret) || strings.Contains(stderr, secret) {
			t.Errorf("el valor %q aparece en la salida", secret)
		}
	}
	return resp
}

func execute(t *testing.T, request string, env []string) (string, string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = env
	cmd.Stdin = strings.NewReader(request)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("aifdemo excedió el tiempo límite")
	}
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("no se pudo ejecutar aifdemo: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

func parseSingleObject(t *testing.T, stdout string) map[string]json.RawMessage {
	t.Helper()
	if !strings.HasSuffix(stdout, "\n") {
		t.Errorf("stdout no termina en salto de línea: %q", stdout)
	}
	dec := json.NewDecoder(strings.NewReader(stdout))
	var resp map[string]json.RawMessage
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("stdout no es un objeto JSON: %v\nstdout: %q", err, stdout)
	}
	if resp == nil {
		t.Fatalf("stdout no es un objeto JSON: %q", stdout)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		t.Fatalf("stdout no contiene exactamente un objeto JSON: %q", stdout)
	}
	return resp
}

func decisionMembers(resp map[string]json.RawMessage) map[string]json.RawMessage {
	decision := make(map[string]json.RawMessage, len(resp))
	for name, raw := range resp {
		if name == "outcome" {
			continue
		}
		decision[name] = raw
	}
	return decision
}

func member(t *testing.T, doc map[string]json.RawMessage, name string) json.RawMessage {
	t.Helper()
	raw, ok := doc[name]
	if !ok {
		t.Fatalf("falta el miembro %q en %v", name, doc)
	}
	return raw
}

func assertString(t *testing.T, doc map[string]json.RawMessage, name, want string) {
	t.Helper()
	assertRawString(t, member(t, doc, name), want, name)
}

func assertRawString(t *testing.T, raw json.RawMessage, want, name string) {
	t.Helper()
	var got string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Errorf("%s no es una cadena: %v", name, err)
		return
	}
	if got != want {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}

func assertStringSlice(t *testing.T, doc map[string]json.RawMessage, name string, want []string) {
	t.Helper()
	var got []string
	if err := json.Unmarshal(member(t, doc, name), &got); err != nil {
		t.Errorf("%s no es una matriz de cadenas: %v", name, err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// assertSubsetObject checks that each expected field and member is present with
// the exact value, while allowing extra members and fields.
func assertSubsetObject(t *testing.T, doc map[string]json.RawMessage, name string, want map[string]map[string]any) {
	t.Helper()
	var actual map[string]map[string]json.RawMessage
	if err := json.Unmarshal(member(t, doc, name), &actual); err != nil {
		t.Errorf("%s no es un objeto de objetos: %v", name, err)
		return
	}
	for field, members := range want {
		gotMembers, ok := actual[field]
		if !ok {
			t.Errorf("%s no describe el campo %q: %v", name, field, actual)
			continue
		}
		for memberName, wantValue := range members {
			raw, ok := gotMembers[memberName]
			if !ok {
				t.Errorf("%s[%s] no tiene %q", name, field, memberName)
				continue
			}
			var gotValue any
			if err := json.Unmarshal(raw, &gotValue); err != nil {
				t.Errorf("%s[%s].%s no es JSON: %v", name, field, memberName, err)
				continue
			}
			if !reflect.DeepEqual(gotValue, wantValue) {
				t.Errorf("%s[%s].%s = %v, want %v", name, field, memberName, gotValue, wantValue)
			}
		}
	}
}

func assertExactObject(t *testing.T, doc map[string]json.RawMessage, name string, want map[string]any) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(member(t, doc, name), &got); err != nil {
		t.Errorf("%s no es un objeto: %v", name, err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func assertDiscovery(t *testing.T, resp map[string]json.RawMessage, required []string) {
	t.Helper()
	var operations []map[string]json.RawMessage
	if err := json.Unmarshal(member(t, resp, "operations"), &operations); err != nil {
		t.Errorf("operations no es una matriz: %v", err)
		return
	}
	seen := map[string]bool{}
	for _, operation := range operations {
		var id, description string
		if err := json.Unmarshal(operation["id"], &id); err != nil || id == "" {
			t.Errorf("operación sin id: %v", operation)
			continue
		}
		if err := json.Unmarshal(operation["description"], &description); err != nil || description == "" {
			t.Errorf("la operación %q no tiene descripción de una línea", id)
		}
		seen[id] = true
	}
	for _, id := range required {
		if !seen[id] {
			t.Errorf("el descubrimiento no incluye la operación %q: %v", id, operations)
		}
	}
}

func baseEnv(store string, extra map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(extra)+1)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "AIF_DEMO_") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "AIF_DEMO_STORE="+store)
	for key, value := range extra {
		env = append(env, key+"="+value)
	}
	return env
}

func seedStore(t *testing.T, path string, entries []json.RawMessage) {
	t.Helper()
	if len(entries) == 0 {
		return
	}
	var buffer bytes.Buffer
	for _, entry := range entries {
		var compact bytes.Buffer
		if err := json.Compact(&compact, entry); err != nil {
			t.Fatalf("entrada inicial del registro inválida: %v", err)
		}
		buffer.Write(compact.Bytes())
		buffer.WriteByte('\n')
	}
	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("no se pudo preparar el registro: %v", err)
	}
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0
	}
	if err != nil {
		t.Fatalf("no se pudo leer el registro: %v", err)
	}
	return bytes.Count(data, []byte("\n"))
}
