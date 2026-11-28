package aif

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
)

// Structural diagnostics of §3.5. Each reason identifies one cause so the
// consumer can act mechanically (R-02).
const (
	malformedEmpty         = "la entrada es vacía"
	malformedTooLarge      = "supera el tamaño máximo de 256 KiB"
	malformedTooDeep       = "supera la profundidad máxima"
	malformedSyntax        = "no es JSON válido"
	malformedNotObject     = "no es un objeto"
	malformedNoAIF         = "carece de aif"
	malformedAIFType       = "aif no es una cadena"
	malformedOperationType = "operation no es una cadena"
	malformedInputType     = "input no es un objeto"
	malformedDuplicate     = "claves duplicadas"
)

var (
	errDuplicate = errors.New("duplicate keys")
	errTooDeep   = errors.New("maximum depth exceeded")
	errSyntax    = errors.New("invalid JSON")
)

// Handle parses one request and returns its response. It never panics on
// malformed input: every structural defect becomes malformed_request.
func (e *Executor) Handle(raw []byte) *Response {
	if len(bytes.TrimSpace(raw)) == 0 {
		return malformed(malformedEmpty)
	}
	if len(raw) > MaxInputBytes {
		return malformed(malformedTooLarge)
	}
	top, err := scanRequest(raw)
	if err != nil {
		if errors.Is(err, errDuplicate) {
			return malformed(malformedDuplicate)
		}
		if errors.Is(err, errTooDeep) {
			return malformed(malformedTooDeep)
		}
		return malformed(malformedSyntax)
	}
	if top != '{' {
		return malformed(malformedNotObject)
	}
	var envelope struct {
		AIF       json.RawMessage `json:"aif"`
		Operation json.RawMessage `json:"operation"`
		Input     json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return malformed(malformedSyntax)
	}
	if isNull(envelope.AIF) {
		return malformed(malformedNoAIF)
	}
	var version string
	if err := json.Unmarshal(envelope.AIF, &version); err != nil {
		return malformed(malformedAIFType)
	}
	if version != Version {
		return &Response{AIF: Version, Result: ResultUnsupportedVersion, Supported: []string{Version}}
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return malformed(malformedSyntax)
	}
	if key, ok := unknownRootMember(members); ok {
		return malformed("miembro desconocido en la raíz: " + key +
			"; los campos de la operación van dentro de input")
	}
	if isNull(envelope.Operation) {
		return &Response{AIF: Version, Discovery: e.discovery()}
	}
	var operationID string
	if err := json.Unmarshal(envelope.Operation, &operationID); err != nil {
		return malformed(malformedOperationType)
	}
	op := e.find(operationID)
	if op == nil {
		return &Response{
			AIF:             Version,
			Operation:       operationID,
			Result:          ResultUnknownOperation,
			KnownOperations: e.knownOperations(),
		}
	}
	input := map[string]json.RawMessage{}
	if !isNull(envelope.Input) {
		if err := json.Unmarshal(envelope.Input, &input); err != nil {
			return malformed(malformedInputType)
		}
	}
	return e.evaluate(op, input)
}

// unknownRootMember returns the lexicographically first root member outside
// the envelope of §1, so a member placed outside input is never dropped in
// silence. Duplicates are rejected earlier by scanRequest.
func unknownRootMember(members map[string]json.RawMessage) (string, bool) {
	var unknown []string
	for name := range members {
		if name != "aif" && name != "operation" && name != "input" {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) == 0 {
		return "", false
	}
	sort.Strings(unknown)
	return unknown[0], true
}

func (e *Executor) evaluate(op *Operation, input map[string]json.RawMessage) *Response {
	known := make(map[string]struct{}, len(op.Fields))
	for _, f := range op.Fields {
		known[f.Name] = struct{}{}
	}
	var missing []string
	invalid := map[string]FieldProblem{}
	collected := map[string]json.RawMessage{}
	for _, f := range op.Fields {
		raw, present := input[f.Name]
		if !present || isNull(raw) {
			if f.Required {
				missing = append(missing, f.Name)
			}
			continue
		}
		value, err := decodeValue(raw)
		if err != nil {
			invalid[f.Name] = problemFor(f, raw, "no es un valor válido")
			continue
		}
		if reason, ok := validateField(f, value); !ok {
			// §3.2 requires a reason; an integrator hook returning ok=false
			// without one still yields a deterministic diagnostic.
			if reason == "" {
				reason = "regla del dominio incumplida"
			}
			invalid[f.Name] = problemFor(f, raw, reason)
			continue
		}
		if !f.Sensitive {
			collected[f.Name] = raw
		}
	}
	for name, raw := range input {
		if _, ok := known[name]; ok {
			continue
		}
		invalid[name] = FieldProblem{Value: raw, Reason: "miembro desconocido: " + name}
	}
	if len(invalid) > 0 {
		return reparable(op, ResultInvalid, missing, invalid, collected)
	}
	if len(missing) > 0 {
		return reparable(op, ResultInputNeeded, missing, invalid, collected)
	}
	return execute(op, input)
}

func reparable(op *Operation, result string, missing []string, invalid map[string]FieldProblem, collected map[string]json.RawMessage) *Response {
	return &Response{
		AIF:       Version,
		Operation: op.ID,
		Result:    result,
		Missing:   missing,
		Invalid:   invalid,
		Fields:    descriptors(op.Fields),
		Collected: collected,
	}
}

func validateField(f Field, value any) (string, bool) {
	if f.Type != "" && !matchesType(value, f.Type) {
		return "tipo no válido: se esperaba " + f.Type + ", llegó " + typeName(value), false
	}
	if len(f.Allowed) > 0 {
		s, ok := value.(string)
		if !ok || !contains(f.Allowed, s) {
			return "valor no permitido", false
		}
	}
	if f.Validate != nil {
		return f.Validate(value)
	}
	return "", true
}

func problemFor(f Field, raw json.RawMessage, reason string) FieldProblem {
	p := FieldProblem{Reason: reason}
	if !f.Sensitive {
		p.Value = raw
	}
	if len(f.Allowed) > 0 {
		p.Allowed = f.Allowed
	}
	if len(f.Schema) > 0 {
		p.Schema = f.Schema
	}
	return p
}

// execute maps a domain error to its response. Each of the three library
// error types is accepted in the value and in the pointer form, matched in
// that order: the Error methods have value receivers, so either form may be
// returned, and errors.As only matches the exact dynamic type, so the pointer
// check goes first.
func execute(op *Operation, input map[string]json.RawMessage) *Response {
	if op.Execute == nil {
		return &Response{
			AIF:       Version,
			Operation: op.ID,
			Result:    ResultFailed,
			Message:   "la operación no tiene implementación",
			Code:      "not_implemented",
		}
	}
	outcome, err := op.Execute(input)
	if err == nil {
		if len(outcome) == 0 {
			outcome = json.RawMessage("null")
		}
		return &Response{AIF: Version, Operation: op.ID, Result: ResultExecuted, Outcome: outcome}
	}
	var deniedPtr *DeniedError
	if errors.As(err, &deniedPtr) {
		if deniedPtr == nil {
			return brokenDomainError(op)
		}
		return &Response{AIF: Version, Operation: op.ID, Result: ResultDenied, Reason: deniedPtr.Reason}
	}
	var denied DeniedError
	if errors.As(err, &denied) {
		return &Response{AIF: Version, Operation: op.ID, Result: ResultDenied, Reason: denied.Reason}
	}
	var invalidPtr *InvalidError
	if errors.As(err, &invalidPtr) {
		if invalidPtr == nil {
			return brokenDomainError(op)
		}
		return reparableFromDomain(op, input, invalidPtr.Fields)
	}
	var invalid InvalidError
	if errors.As(err, &invalid) {
		return reparableFromDomain(op, input, invalid.Fields)
	}
	var domainPtr *DomainError
	if errors.As(err, &domainPtr) {
		if domainPtr == nil {
			return brokenDomainError(op)
		}
		return &Response{
			AIF:       Version,
			Operation: op.ID,
			Result:    ResultFailed,
			Message:   domainPtr.Message,
			Code:      domainPtr.Code,
		}
	}
	var domain DomainError
	if errors.As(err, &domain) {
		return &Response{
			AIF:       Version,
			Operation: op.ID,
			Result:    ResultFailed,
			Message:   domain.Message,
			Code:      domain.Code,
		}
	}
	return &Response{AIF: Version, Operation: op.ID, Result: ResultFailed, Message: err.Error()}
}

// reparableFromDomain rebuilds an `invalid` response from a domain rule. The
// domain must name known, present, non-null fields with a non-empty reason;
// anything else is a defect it cannot report as repairable, so it is failed.
// Collected is rebuilt from the request, never from the domain, and sensitive
// values stay out of both members.
func reparableFromDomain(op *Operation, input map[string]json.RawMessage, problems map[string]string) *Response {
	if len(problems) == 0 {
		return invalidErrorFailed(op)
	}
	known := make(map[string]Field, len(op.Fields))
	for _, f := range op.Fields {
		known[f.Name] = f
	}
	invalid := make(map[string]FieldProblem, len(problems))
	for name, reason := range problems {
		f, ok := known[name]
		raw, present := input[name]
		if !ok || !present || isNull(raw) || strings.TrimSpace(reason) == "" {
			return invalidErrorFailed(op)
		}
		invalid[name] = problemFor(f, raw, reason)
	}
	collected := make(map[string]json.RawMessage, len(op.Fields))
	for _, f := range op.Fields {
		raw, present := input[f.Name]
		if !present || isNull(raw) || f.Sensitive {
			continue
		}
		if _, rejected := invalid[f.Name]; rejected {
			continue
		}
		collected[f.Name] = raw
	}
	return reparable(op, ResultInvalid, nil, invalid, collected)
}

func invalidErrorFailed(op *Operation) *Response {
	return &Response{
		AIF:       Version,
		Operation: op.ID,
		Result:    ResultFailed,
		Message:   "el dominio informó de campos inválidos que la operación no puede describir",
	}
}

// brokenDomainError is the response for an unusable domain error: a typed-nil
// error pointer (the classic `var e *DeniedError; return e` bug) is a non-nil
// error interface with no data, a domain defect — never a denial or a repair,
// and never dereferenced.
func brokenDomainError(op *Operation) *Response {
	return &Response{
		AIF:       Version,
		Operation: op.ID,
		Result:    ResultFailed,
		Message:   "el dominio devolvió un error que el ejecutor no puede interpretar",
	}
}

func malformed(reason string) *Response {
	return &Response{
		AIF:       Version,
		Result:    ResultMalformedRequest,
		Reason:    reason,
		Supported: []string{Version},
	}
}

func (e *Executor) find(id string) *Operation {
	for _, op := range e.Operations {
		if op.ID == id {
			return op
		}
	}
	return nil
}

func (e *Executor) knownOperations() []string {
	ids := make([]string, 0, len(e.Operations))
	for _, op := range e.Operations {
		ids = append(ids, op.ID)
	}
	return ids
}

func (e *Executor) discovery() *Discovery {
	ops := make([]OperationInfo, 0, len(e.Operations))
	for _, op := range e.Operations {
		ops = append(ops, OperationInfo{ID: op.ID, Description: op.Description})
	}
	return &Discovery{
		AIF:        Version,
		Request:    RequestTemplate{AIF: Version, Operation: "<id>", Input: map[string]any{}},
		Operations: ops,
	}
}

func descriptors(fields []Field) map[string]FieldDescriptor {
	m := make(map[string]FieldDescriptor, len(fields))
	for _, f := range fields {
		m[f.Name] = FieldDescriptor{
			Required:    f.Required,
			Type:        f.Type,
			Allowed:     f.Allowed,
			Description: f.Description,
			Schema:      f.Schema,
		}
	}
	return m
}

func decodeValue(raw json.RawMessage) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func matchesType(value any, want string) bool {
	switch want {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(json.Number)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	}
	return false
}

// typeName names a decoded JSON value for a diagnostic reason, which is prose
// and not contract (§3.2, R-03).
func typeName(value any) string {
	switch value.(type) {
	case string:
		return "cadena"
	case json.Number:
		return "número"
	case bool:
		return "booleano"
	case map[string]any:
		return "objeto"
	case []any:
		return "array"
	}
	return "valor de tipo desconocido"
}

func contains(allowed []string, value string) bool {
	for _, candidate := range allowed {
		if candidate == value {
			return true
		}
	}
	return false
}

func isNull(raw json.RawMessage) bool {
	return len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

// scanRequest validates the JSON grammar, rejects duplicate object keys and
// enforces the declared depth, which encoding/json alone does not do.
func scanRequest(raw []byte) (json.Delim, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	token, err := dec.Token()
	if err != nil {
		return 0, err
	}
	if err := scanValue(dec, token, 1); err != nil {
		return 0, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return 0, errSyntax
	}
	delim, _ := token.(json.Delim)
	return delim, nil
}

func scanValue(dec *json.Decoder, token json.Token, depth int) error {
	if depth > MaxDepth {
		return errTooDeep
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errSyntax
			}
			if _, duplicate := seen[key]; duplicate {
				return errDuplicate
			}
			seen[key] = struct{}{}
			valueToken, err := dec.Token()
			if err != nil {
				return err
			}
			if err := scanValue(dec, valueToken, depth+1); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	case '[':
		for dec.More() {
			valueToken, err := dec.Token()
			if err != nil {
				return err
			}
			if err := scanValue(dec, valueToken, depth+1); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	}
	return nil
}
