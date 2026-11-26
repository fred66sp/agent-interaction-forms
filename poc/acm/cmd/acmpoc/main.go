// Command acmpoc is the AIF v0.2 CLI profile (§7) for one real ACM operation.
// The learning operation mirrors EVENT_FORMS["learning"] of the ACM adapter,
// which is run as a child process (ACM_POC_TOOL): AIF decides the envelope,
// the field structure and the closed confidence set, while ACM keeps its own
// validation, deduplication by content_hash, form_id idempotency, ETag and
// transactions. No domain rule and no guarantee is reimplemented here.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/fred66sp/agent-interaction-forms"
)

// confidences is the closed set of EVENT_FORMS["learning"] in acm.py.
var confidences = []string{"confirmed", "inference", "hypothesis", "proposal", "unknown"}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Getenv))
}

// run is the whole CLI, testable without spawning the process.
func run(stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string) int {
	raw, err := io.ReadAll(io.LimitReader(stdin, aif.MaxInputBytes+1))
	if err != nil {
		fmt.Fprintf(stderr, "acmpoc: no se pudo leer la entrada: %v\n", err)
		return write(stdout, stderr, &aif.Response{
			AIF:     aif.Version,
			Result:  aif.ResultFailed,
			Message: "no se pudo leer la petición",
			Code:    "stdin_unavailable",
		})
	}
	executor := aif.Executor{Operations: []*aif.Operation{learning(getenv)}}
	return write(stdout, stderr, executor.Handle(raw))
}

// write applies the CLI profile of §7: one JSON object, a trailing newline and
// the exit code of the result.
func write(stdout, stderr io.Writer, response *aif.Response) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		if errors.Is(err, syscall.EPIPE) {
			return 0
		}
		fmt.Fprintf(stderr, "acmpoc: no se pudo escribir la respuesta: %v\n", err)
		return 1
	}
	return exitCode(response)
}

func exitCode(response *aif.Response) int {
	if response.Discovery != nil {
		return 0
	}
	switch response.Result {
	case aif.ResultExecuted:
		return 0
	case aif.ResultInputNeeded, aif.ResultInvalid:
		return 2
	case aif.ResultDenied:
		return 4
	case aif.ResultFailed:
		return 1
	default:
		return 3
	}
}

func learning(getenv func(string) string) *aif.Operation {
	return &aif.Operation{
		ID:          "learning",
		Description: "Registra un aprendizaje reutilizable en la memoria ACM del proyecto.",
		Fields: []aif.Field{
			{
				Name:        "title",
				Required:    true,
				Type:        "string",
				Description: "Título breve del aprendizaje",
				// ACM treats a blank required value as absent; declare the same
				// rule so AIF answers with a repair instead of calling ACM.
				Schema:   json.RawMessage(`{"minLength":1}`),
				Validate: notBlank,
			},
			{
				Name:        "body",
				Required:    true,
				Type:        "string",
				Description: "Texto completo del aprendizaje",
				Schema:      json.RawMessage(`{"minLength":1}`),
				Validate:    notBlank,
			},
			{
				Name:        "topic",
				Type:        "string",
				Description: "Clave de tema libre; ACM la normaliza",
			},
			{
				Name:        "confidence",
				Type:        "string",
				Allowed:     confidences,
				Description: "Si se omite, ACM registra confirmed",
			},
		},
		Execute: func(input map[string]json.RawMessage) (json.RawMessage, error) {
			return record(input, getenv)
		},
	}
}

func notBlank(v any) (string, bool) {
	s, _ := v.(string)
	if strings.TrimSpace(s) == "" {
		return "obligatorio", false
	}
	return "", true
}

// record delegates one learning to the real ACM adapter and returns its JSON
// answer as the opaque outcome (§3.1). No sensitive value exists in this
// domain, so nothing is hidden from the outcome.
func record(input map[string]json.RawMessage, getenv func(string) string) (json.RawMessage, error) {
	tool := getenv("ACM_POC_TOOL")
	if tool == "" {
		return nil, aif.DomainError{
			Message: "ACM_POC_TOOL no está definido; no hay ejecutor de dominio",
			Code:    "acm_unavailable",
		}
	}
	args := []string{tool, "learning", "--title", text(input, "title"), "--body", text(input, "body")}
	if topic := text(input, "topic"); topic != "" {
		args = append(args, "--topic", topic)
	}
	if confidence := text(input, "confidence"); confidence != "" {
		args = append(args, "--confidence", confidence)
	}
	cmd := exec.Command("python3", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	exit := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exit = exitErr.ExitCode()
	} else if err != nil {
		return nil, aif.DomainError{
			Message: "no se pudo invocar ACM: " + err.Error(),
			Code:    "acm_unavailable",
		}
	}
	var reply acmReply
	if json.Unmarshal(stdout.Bytes(), &reply) != nil {
		return nil, aif.DomainError{
			Message: fmt.Sprintf("ACM devolvió salida no JSON (exit %d)", exit),
			Code:    "acm_contract",
		}
	}
	if !reply.OK {
		return nil, acmFailure(reply)
	}
	switch reply.Result {
	case "created", "consolidated", "already_submitted":
		// The full ACM answer, unchanged, is the outcome.
		return json.RawMessage(bytes.TrimSpace(stdout.Bytes())), nil
	}
	return nil, aif.DomainError{
		Message: fmt.Sprintf("ACM devolvió un resultado inesperado (%q); la escritura no se da por hecha", reply.Result),
		Code:    "acm_contract",
	}
}

// acmReply is the one-line JSON contract of acm.py (success and error).
type acmReply struct {
	OK      bool              `json:"ok"`
	Result  string            `json:"result"`
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Invalid map[string]string `json:"invalid"`
}

func acmFailure(reply acmReply) error {
	if reply.Error == "E_DENIED" {
		return aif.DeniedError{Reason: reply.Message}
	}
	if len(reply.Invalid) > 0 {
		// ACM validates before writing; AIF rebuilds the repairable response
		// from these reasons and the request, and never echoes secrets.
		return aif.InvalidError{Fields: reply.Invalid}
	}
	return aif.DomainError{Message: reply.Message, Code: reply.Error}
}

// text reads a value the library already validated as a string.
func text(input map[string]json.RawMessage, name string) string {
	var s string
	_ = json.Unmarshal(input[name], &s)
	return s
}
