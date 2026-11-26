// Command aifdemo demonstrates the AIF v0.2 CLI profile (§7) with one
// operation, record_observation, backed by the JSON Lines file named by
// AIF_DEMO_STORE. It reads one request from stdin (limits: 256 KiB and depth
// 32), writes one JSON object to stdout and diagnostics to stderr. Domain
// preconditions: AIF_DEMO_DENY_WRITE=1 denies, AIF_DEMO_STORAGE_FAIL=1 fails.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/fred66sp/agent-interaction-forms"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Getenv, time.Now))
}

// run is the whole CLI, testable without spawning the process.
func run(stdin io.Reader, stdout, stderr io.Writer, getenv func(string) string, now func() time.Time) int {
	raw, err := io.ReadAll(io.LimitReader(stdin, aif.MaxInputBytes+1))
	if err != nil {
		fmt.Fprintf(stderr, "aifdemo: no se pudo leer la entrada: %v\n", err)
		return write(stdout, stderr, &aif.Response{
			AIF:     aif.Version,
			Result:  aif.ResultFailed,
			Message: "no se pudo leer la petición",
			Code:    "stdin_unavailable",
		})
	}
	executor := aif.Executor{Operations: []*aif.Operation{recordObservation(getenv, now)}}
	return write(stdout, stderr, executor.Handle(raw))
}

func write(stdout, stderr io.Writer, response *aif.Response) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		if errors.Is(err, syscall.EPIPE) {
			return 0
		}
		fmt.Fprintf(stderr, "aifdemo: no se pudo escribir la respuesta: %v\n", err)
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

func recordObservation(getenv func(string) string, now func() time.Time) *aif.Operation {
	return &aif.Operation{
		ID:          "record_observation",
		Description: "Registra una observación con título, tipo y confianza.",
		Fields: []aif.Field{
			{
				Name:        "title",
				Required:    true,
				Type:        "string",
				Description: "Título breve de la observación",
				Schema:      json.RawMessage(`{"maxLength":40}`),
				Validate: func(v any) (string, bool) {
					s, ok := v.(string)
					if ok && utf8.RuneCountInString(s) > 40 {
						return "supera la longitud máxima", false
					}
					return "", true
				},
			},
			{
				Name:     "kind",
				Required: true,
				Type:     "string",
				Allowed:  []string{"learning", "decision", "discovery", "bugfix"},
			},
			{
				Name:     "confidence",
				Required: true,
				Type:     "string",
				Allowed:  []string{"hypothesis", "confirmed"},
			},
			{
				Name:        "contact",
				Type:        "string",
				Sensitive:   true,
				Description: "Dato de contacto del autor (sensible)",
			},
		},
		Execute: func(input map[string]json.RawMessage) (json.RawMessage, error) {
			return record(input, getenv, now)
		},
	}
}

func record(input map[string]json.RawMessage, getenv func(string) string, now func() time.Time) (json.RawMessage, error) {
	// Domain rule before any effect: metadata validates each field, but only
	// the domain can decide that a combination is unacceptable.
	if stringValue(input["kind"]) == "decision" && stringValue(input["confidence"]) == "hypothesis" {
		return nil, aif.InvalidError{Fields: map[string]string{
			"confidence": "una decisión no puede ser una hipótesis",
		}}
	}
	if getenv("AIF_DEMO_DENY_WRITE") == "1" {
		return nil, aif.DeniedError{Reason: "el consumidor no tiene permiso de escritura sobre el dominio"}
	}
	if getenv("AIF_DEMO_STORAGE_FAIL") == "1" {
		return nil, storageFailure()
	}
	path := getenv("AIF_DEMO_STORE")
	if path == "" {
		return nil, aif.DomainError{
			Message: "AIF_DEMO_STORE no está definido; la observación no queda registrada",
			Code:    "storage_unavailable",
		}
	}
	count, err := countRecords(path)
	if err != nil {
		return nil, storageFailure()
	}
	entry := recordEntry{
		ID:         fmt.Sprintf("obs_%04d", count+1),
		RecordedAt: now().UTC().Format(time.RFC3339),
		Title:      stringValue(input["title"]),
		Kind:       stringValue(input["kind"]),
		Confidence: stringValue(input["confidence"]),
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return nil, storageFailure()
	}
	if err := appendLine(path, line); err != nil {
		return nil, storageFailure()
	}
	outcome, err := json.Marshal(map[string]string{
		"id":          entry.ID,
		"recorded_at": entry.RecordedAt,
	})
	if err != nil {
		return nil, storageFailure()
	}
	return outcome, nil
}

type recordEntry struct {
	ID         string `json:"id"`
	RecordedAt string `json:"recorded_at"`
	Title      string `json:"title"`
	Kind       string `json:"kind"`
	Confidence string `json:"confidence"`
}

// storageFailure reports the unknown state §3.4 requires: the caller cannot
// tell whether the write completed, so it is never reported as executed.
func storageFailure() error {
	return aif.DomainError{
		Message: "el almacenamiento rechazó la escritura; estado desconocido, la observación no queda registrada",
		Code:    "storage_unavailable",
	}
}

// stringValue reads a value the library already validated as a string.
func stringValue(raw json.RawMessage) string {
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

func countRecords(path string) (int, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

func appendLine(path string, line []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(append(line, '\n'))
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
