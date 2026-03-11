package workspace

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceCommand_GoWorkspace_RendersModuleEdges(t *testing.T) {
	repoDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(repoDir, "go.work"), []byte("go 1.25\n\nuse (\n    ./mod-a\n    ./mod-b\n)\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	modADir := filepath.Join(repoDir, "mod-a")
	modBDir := filepath.Join(repoDir, "mod-b")
	if err := os.MkdirAll(modADir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(modBDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(modADir, "go.mod"), []byte("module example.com/mod-a\n\ngo 1.25\n\nrequire example.com/mod-b v0.0.0\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(modBDir, "go.mod"), []byte("module example.com/mod-b\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"-r", repoDir, "--language", "go", "-f", "dot"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "[go] example.com/mod-a") || !strings.Contains(output, "[go] example.com/mod-b") {
		t.Fatalf("expected output to include go module nodes, got:\n%s", output)
	}
	if !strings.Contains(output, `"[go] example.com/mod-a" -> "[go] example.com/mod-b"`) {
		t.Fatalf("expected output to include go module edge mod-a -> mod-b, got:\n%s", output)
	}
}

func TestWorkspaceCommand_RustWorkspace_RendersCrateEdges(t *testing.T) {
	repoDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(repoDir, "Cargo.toml"), []byte("[workspace]\nmembers = [\"crate-a\", \"crate-b\", \"crate-c\"]\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	crateADir := filepath.Join(repoDir, "crate-a")
	crateBDir := filepath.Join(repoDir, "crate-b")
	crateCDir := filepath.Join(repoDir, "crate-c")
	if err := os.MkdirAll(crateADir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(crateBDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(crateCDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(crateADir, "Cargo.toml"), []byte("[package]\nname = \"crate_a\"\nversion = \"0.1.0\"\n\n[dependencies]\ncrate_b = { path = \"../crate-b\", package = \"crate_b\" }\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(crateBDir, "Cargo.toml"), []byte("[package]\nname = \"crate_b\"\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(crateCDir, "Cargo.toml"), []byte("[package]\nname = \"crate_c\"\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{"-r", repoDir, "--language", "rust", "-f", "dot"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "[rust] crate_a") || !strings.Contains(output, "[rust] crate_b") {
		t.Fatalf("expected output to include rust crate nodes, got:\n%s", output)
	}
	if !strings.Contains(output, `"[rust] crate_a" -> "[rust] crate_b"`) {
		t.Fatalf("expected output to include rust crate edge crate_a -> crate_b, got:\n%s", output)
	}
}

func TestWorkspaceCommand_RustWorkspace_ManifestPrunesToComponent(t *testing.T) {
	repoDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(repoDir, "Cargo.toml"), []byte("[workspace]\nmembers = [\"crate-a\", \"crate-b\", \"crate-c\"]\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	crateADir := filepath.Join(repoDir, "crate-a")
	crateBDir := filepath.Join(repoDir, "crate-b")
	crateCDir := filepath.Join(repoDir, "crate-c")
	if err := os.MkdirAll(crateADir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(crateBDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(crateCDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(crateADir, "Cargo.toml"), []byte("[package]\nname = \"crate_a\"\nversion = \"0.1.0\"\n\n[dependencies]\ncrate_b = { path = \"../crate-b\", package = \"crate_b\" }\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(crateBDir, "Cargo.toml"), []byte("[package]\nname = \"crate_b\"\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(crateCDir, "Cargo.toml"), []byte("[package]\nname = \"crate_c\"\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"-r",
		repoDir,
		"--language",
		"rust",
		"--manifest",
		"crate-a/Cargo.toml",
		"-f",
		"dot",
	})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "[rust] crate_a") || !strings.Contains(output, "[rust] crate_b") {
		t.Fatalf("expected pruned output to include crate_a and crate_b, got:\n%s", output)
	}
	if strings.Contains(output, "[rust] crate_c") {
		t.Fatalf("expected pruned output to exclude disconnected crate_c, got:\n%s", output)
	}
}

func TestWorkspaceCommand_ManifestNotFound_ReturnsError(t *testing.T) {
	repoDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(repoDir, "Cargo.toml"), []byte("[workspace]\nmembers = [\"crate-a\"]\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	crateADir := filepath.Join(repoDir, "crate-a")
	if err := os.MkdirAll(crateADir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(crateADir, "Cargo.toml"), []byte("[package]\nname = \"crate_a\"\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"-r",
		repoDir,
		"--language",
		"rust",
		"--manifest",
		"crate-b/Cargo.toml",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing manifest")
	}
	if !strings.Contains(err.Error(), "manifest not found in workspace graph") {
		t.Fatalf("expected manifest not found error, got: %v", err)
	}
}

func TestWorkspaceCommand_UnknownLanguage_ReturnsError(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"--language", "python"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown language")
	}
	if !strings.Contains(err.Error(), "unknown language") {
		t.Fatalf("expected unknown language error, got: %v", err)
	}
}
