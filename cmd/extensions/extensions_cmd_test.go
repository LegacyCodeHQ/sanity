package extensions

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestExtensionsCommand_DefaultJSON(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	var payload struct {
		Extensions []struct {
			Extension string `json:"extension"`
			Language  string `json:"language"`
			Maturity  string `json:"maturity"`
		} `json:"extensions"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}

	if len(payload.Extensions) == 0 {
		t.Fatalf("expected at least one extension entry")
	}

	foundRS := false
	for _, entry := range payload.Extensions {
		if entry.Extension == ".rs" {
			foundRS = true
			if entry.Language != "Rust" {
				t.Fatalf("extension .rs language = %q, want %q", entry.Language, "Rust")
			}
		}
	}
	if !foundRS {
		t.Fatalf("expected .rs extension in output")
	}
}

func TestExtensionsCommand_TextFormat(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--format", "text"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "EXTENSION") || !strings.Contains(got, "LANGUAGE") {
		t.Fatalf("expected table header, got:\n%s", got)
	}
	if !strings.Contains(got, ".rs") {
		t.Fatalf("expected .rs in text output, got:\n%s", got)
	}
}

func TestExtensionsCommand_UnknownFormat(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--format", "yaml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("unexpected error: %v", err)
	}
}
