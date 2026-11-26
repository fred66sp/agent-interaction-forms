// Package aif implements the AIF v0.2 interaction contract (docs/SPECIFICATION.md
// §§1–6): request envelopes, the eight results, field descriptions, sensitive
// values and stateless continuation. The CLI profile of §7 is demonstrated by
// cmd/aifdemo.
package aif

import "encoding/json"

// Version is the AIF contract version implemented by this package. It is the
// only version accepted: any other declared version is unsupported.
const Version = "0.2"

// Result values defined by §2.
const (
	ResultExecuted           = "executed"
	ResultInputNeeded        = "input_needed"
	ResultInvalid            = "invalid"
	ResultDenied             = "denied"
	ResultFailed             = "failed"
	ResultUnsupportedVersion = "unsupported_version"
	ResultMalformedRequest   = "malformed_request"
	ResultUnknownOperation   = "unknown_operation"
)

// MaxInputBytes is the maximum request size the executor accepts (L-04).
const MaxInputBytes = 256 << 10

// MaxDepth is the maximum JSON nesting depth the executor accepts (L-04).
const MaxDepth = 32

// Field describes one input member of an operation (§4).
type Field struct {
	Name        string
	Required    bool
	Type        string
	Allowed     []string
	Description string
	Schema      json.RawMessage
	// Sensitive values are never returned in a response (L-03).
	Sensitive bool
	// Validate applies a domain rule that Type and Allowed cannot express,
	// such as maxLength. It returns the deterministic reason when rejected.
	Validate func(v any) (reason string, ok bool)
}

// Operation is one operation served by an Executor.
type Operation struct {
	ID          string
	Description string
	Fields      []Field
	// Execute receives the validated raw input and returns the opaque outcome.
	// Return DeniedError to deny, DomainError to report a failure, or any other
	// error to report a failure without a code.
	Execute func(input map[string]json.RawMessage) (json.RawMessage, error)
}

// Executor serves a fixed set of operations.
type Executor struct {
	Operations []*Operation
}

// DeniedError reports a domain denial (§3.3).
type DeniedError struct{ Reason string }

func (e DeniedError) Error() string { return e.Reason }

// InvalidError reports a domain rule that rejected input fields (§3.2). Fields
// maps each field name to its deterministic reason. Return it from Execute at
// the start, before any effect: the executor rebuilds the repairable response
// from these reasons, the request and the operation metadata, so the domain
// never composes `invalid` or `collected` itself.
type InvalidError struct{ Fields map[string]string }

func (e InvalidError) Error() string { return "invalid input" }

// DomainError reports a domain or infrastructure failure (§3.4).
type DomainError struct {
	Message string
	Code    string
}

func (e DomainError) Error() string { return e.Message }
