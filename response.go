package aif

import "encoding/json"

// FieldProblem is the descriptor of an invalid field (§3.2).
type FieldProblem struct {
	Value   json.RawMessage `json:"value,omitempty"`
	Reason  string          `json:"reason"`
	Allowed []string        `json:"allowed,omitempty"`
	Schema  json.RawMessage `json:"schema,omitempty"`
}

// FieldDescriptor is the description of a field (§4).
type FieldDescriptor struct {
	Required    bool            `json:"required"`
	Type        string          `json:"type,omitempty"`
	Allowed     []string        `json:"allowed,omitempty"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
}

// Response is one AIF response. Discovery is set instead of Result when the
// request omits the operation member (CL-05).
type Response struct {
	AIF       string
	Operation string
	Result    string
	Outcome   json.RawMessage
	Missing   []string
	Invalid   map[string]FieldProblem
	Fields    map[string]FieldDescriptor
	// Collected echoes the known, present, non-null, non-sensitive values the
	// request supplied, minus the fields reported in Invalid. It is not a
	// validity claim: Fields and Invalid carry the descriptors.
	Collected       map[string]json.RawMessage
	Reason          string
	Message         string
	Code            string
	Supported       []string
	KnownOperations []string
	Discovery       *Discovery
}

// MarshalJSON emits exactly the members §3 requires for the result. The
// encoding/json package sorts map keys, so the output is deterministic (R-02).
func (r *Response) MarshalJSON() ([]byte, error) {
	if r.Discovery != nil {
		return json.Marshal(r.Discovery)
	}
	m := map[string]any{
		"aif":    r.AIF,
		"result": r.Result,
	}
	if r.Result != ResultUnsupportedVersion && r.Result != ResultMalformedRequest {
		m["operation"] = r.Operation
	}
	switch r.Result {
	case ResultExecuted:
		m["outcome"] = r.Outcome
	case ResultInputNeeded, ResultInvalid:
		if len(r.Missing) > 0 {
			m["missing"] = r.Missing
		}
		if len(r.Invalid) > 0 {
			m["invalid"] = r.Invalid
		}
		fields := r.Fields
		if fields == nil {
			fields = map[string]FieldDescriptor{}
		}
		m["fields"] = fields
		collected := r.Collected
		if collected == nil {
			collected = map[string]json.RawMessage{}
		}
		m["collected"] = collected
	case ResultDenied:
		m["reason"] = r.Reason
	case ResultFailed:
		m["message"] = r.Message
		if r.Code != "" {
			m["code"] = r.Code
		}
	case ResultUnsupportedVersion:
		m["supported"] = r.Supported
	case ResultMalformedRequest:
		m["reason"] = r.Reason
		if len(r.Supported) > 0 {
			m["supported"] = r.Supported
		}
	case ResultUnknownOperation:
		m["known_operations"] = r.KnownOperations
	}
	return json.Marshal(m)
}

// Discovery is the CL-05 response: it is not an operation response and carries
// neither result nor operation.
type Discovery struct {
	AIF        string          `json:"aif"`
	Operations []OperationInfo `json:"operations"`
}

// OperationInfo is one line of the discovery response.
type OperationInfo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}
