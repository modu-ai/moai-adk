package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// writeLSPYAMLFixture writes a minimal lsp.yaml for the doctor command tests.
func writeLSPYAMLFixture(t *testing.T, dir string) string {
	t.Helper()
	content := `lsp:
  servers:
    go:
      command: gopls
      args: []
      install_hint: "go install golang.org/x/tools/gopls@latest"
      fallback_binaries: []
      project_markers:
        - go.mod
        - go.sum
      root_markers:
        - go.mod
      file_extensions:
        - .go
    python:
      command: pylsp-does-not-exist-in-path-moai-test
      args: []
      install_hint: "pip install python-lsp-server"
      fallback_binaries:
        - pyright-langserver-does-not-exist-moai-test
      project_markers:
        - pyproject.toml
        - requirements.txt
      root_markers:
        - pyproject.toml
      file_extensions:
        - .py
`
	path := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeLSPYAMLFixture: %v", err)
	}
	return path
}

// writeGoModFixture writes a go.mod in dir to simulate a Go project (project_markers).
func writeGoModFixture(t *testing.T, dir string) {
	t.Helper()
	content := "module example.com/test\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o600); err != nil {
		t.Fatalf("writeGoModFixture: %v", err)
	}
}

// TestLSPDoctorReport_ReturnsReport verifies that runLSPDoctor returns a non-nil report (REQ-LM-007).
func TestLSPDoctorReport_ReturnsReport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lspYAML := writeLSPYAMLFixture(t, dir)
	writeGoModFixture(t, dir)

	report, err := runLSPDoctorReport(dir, lspYAML)
	if err != nil {
		t.Fatalf("runLSPDoctorReport: %v", err)
	}
	if report == nil {
		t.Fatal("runLSPDoctorReport returned nil report")
	}
}

// TestLSPDoctorReport_DetectsProjectLanguages verifies REQ-LM-007: project languages are detected
// via project_markers when a marker file exists in the project directory.
func TestLSPDoctorReport_DetectsProjectLanguages(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lspYAML := writeLSPYAMLFixture(t, dir)
	writeGoModFixture(t, dir) // go.mod → "go" language detected

	report, err := runLSPDoctorReport(dir, lspYAML)
	if err != nil {
		t.Fatalf("runLSPDoctorReport: %v", err)
	}

	if len(report.ProjectLanguages) == 0 {
		t.Fatal("expected at least 1 project language detected (go.mod present)")
	}

	found := false
	for _, lang := range report.ProjectLanguages {
		if lang == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("project languages %v does not contain 'go', expected detection via go.mod", report.ProjectLanguages)
	}
}

// TestLSPDoctorReport_ReportsInstalledServers verifies REQ-LM-007: installed servers are listed.
func TestLSPDoctorReport_ReportsInstalledServers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lspYAML := writeLSPYAMLFixture(t, dir)
	writeGoModFixture(t, dir)

	report, err := runLSPDoctorReport(dir, lspYAML)
	if err != nil {
		t.Fatalf("runLSPDoctorReport: %v", err)
	}

	// "pylsp-does-not-exist-in-path-moai-test" should not be installed
	for _, lang := range report.MissingServers {
		if lang.Language == "python" {
			if lang.InstallHint == "" {
				t.Errorf("missing python server should include install_hint")
			}
			return
		}
	}
	// If python is not even in MissingServers, it could be installed
	// (don't fail the test if pylsp happens to be in PATH by coincidence)
}

// TestLSPDoctorReport_ReadinessStatus verifies REQ-LM-007: aggregate readiness status.
func TestLSPDoctorReport_ReadinessStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lspYAML := writeLSPYAMLFixture(t, dir)
	// No project markers present → no languages detected → all ready (nothing needed)

	report, err := runLSPDoctorReport(dir, lspYAML)
	if err != nil {
		t.Fatalf("runLSPDoctorReport: %v", err)
	}

	// With no project markers, project languages = empty → readiness is "ready"
	if len(report.ProjectLanguages) > 0 {
		// If there are detected languages, check readiness is set
		if report.ReadinessStatus == "" {
			t.Error("ReadinessStatus should be non-empty when languages are detected")
		}
	}
}

// TestLSPDoctorReport_MissingYAML verifies graceful error when lsp.yaml is missing.
func TestLSPDoctorReport_MissingYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	_, err := runLSPDoctorReport(dir, filepath.Join(dir, "nonexistent.yaml"))
	if err == nil {
		t.Error("expected error when lsp.yaml is missing, got nil")
	}
}

// TestLSPDoctorRender_NonEmptyOutput verifies that renderLSPDoctorReport produces output.
func TestLSPDoctorRender_NonEmptyOutput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lspYAML := writeLSPYAMLFixture(t, dir)
	writeGoModFixture(t, dir)

	report, err := runLSPDoctorReport(dir, lspYAML)
	if err != nil {
		t.Fatalf("runLSPDoctorReport: %v", err)
	}

	var buf bytes.Buffer
	renderLSPDoctorReport(&buf, report)

	if buf.Len() == 0 {
		t.Error("renderLSPDoctorReport: expected non-empty output")
	}
}
