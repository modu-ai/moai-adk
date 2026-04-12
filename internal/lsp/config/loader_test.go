package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// sampleLSPYAML is a minimal lsp.yaml fixture with go, python, typescript servers.
const sampleLSPYAML = `
lsp:
  servers:
    go:
      command: gopls
      args: []
      init_options:
        staticcheck: true
      idle_shutdown_seconds: 600
      root_markers:
        - go.mod
        - go.sum
      file_extensions:
        - .go
    python:
      command: pylsp
      args: []
      idle_shutdown_seconds: 300
      root_markers:
        - pyproject.toml
        - setup.py
      file_extensions:
        - .py
    typescript:
      command: typescript-language-server
      args:
        - --stdio
      idle_shutdown_seconds: 600
      root_markers:
        - tsconfig.json
        - package.json
      file_extensions:
        - .ts
        - .tsx
`

// sampleLSPYAMLWithExtensions is a fixture including the REQ-LM-004/008/001 fields.
const sampleLSPYAMLWithExtensions = `
lsp:
  servers:
    python:
      command: pylsp
      args: []
      install_hint: "pip install python-lsp-server"
      fallback_binaries:
        - pyright-langserver
        - basedpyright-langserver
      project_markers:
        - pyproject.toml
        - requirements.txt
      root_markers:
        - pyproject.toml
      file_extensions:
        - .py
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
`

// writeTempYAML writes content to a temp file and returns the path.
func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTempYAML: %v", err)
	}
	return path
}

// TestLoad_HappyPath verifies YAML parsing of well-formed lsp.yaml.
func TestLoad_HappyPath(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleLSPYAML)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load(%q) error = %v", path, err)
	}
	if cfg == nil {
		t.Fatal("Load returned nil config")
	}

	if len(cfg.Servers) != 3 {
		t.Errorf("Servers count = %d, want 3", len(cfg.Servers))
	}

	go_, ok := cfg.Servers["go"]
	if !ok {
		t.Fatal("Servers missing key 'go'")
	}
	if go_.Command != "gopls" {
		t.Errorf("go.Command = %q, want 'gopls'", go_.Command)
	}
	if go_.IdleShutdownSeconds != 600 {
		t.Errorf("go.IdleShutdownSeconds = %d, want 600", go_.IdleShutdownSeconds)
	}
	if len(go_.RootMarkers) != 2 {
		t.Errorf("go.RootMarkers length = %d, want 2", len(go_.RootMarkers))
	}
	if len(go_.FileExtensions) != 1 {
		t.Errorf("go.FileExtensions length = %d, want 1", len(go_.FileExtensions))
	}

	ts, ok := cfg.Servers["typescript"]
	if !ok {
		t.Fatal("Servers missing key 'typescript'")
	}
	if len(ts.Args) != 1 || ts.Args[0] != "--stdio" {
		t.Errorf("typescript.Args = %v, want [--stdio]", ts.Args)
	}
}

// TestLoad_InitOptions verifies init_options are parsed into map[string]any.
func TestLoad_InitOptions(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleLSPYAML)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	go_, ok := cfg.Servers["go"]
	if !ok {
		t.Fatal("Servers missing key 'go'")
	}
	if go_.InitOptions == nil {
		t.Fatal("go.InitOptions is nil, want map with staticcheck")
	}
	v, ok := go_.InitOptions["staticcheck"]
	if !ok {
		t.Error("InitOptions missing key 'staticcheck'")
	}
	if v != true {
		t.Errorf("InitOptions[staticcheck] = %v (%T), want true", v, v)
	}
}

// TestLoad_MissingFile verifies Load returns an error when file does not exist.
func TestLoad_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := config.Load("/nonexistent/path/lsp.yaml")
	if err == nil {
		t.Error("Load on missing file: expected error, got nil")
	}
}

// TestLoad_MalformedYAML verifies Load returns an error for invalid YAML.
func TestLoad_MalformedYAML(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, `this: is: not: valid: yaml: :::`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("Load on malformed YAML: expected error, got nil")
	}
}

// TestLoad_EmptyServersSection verifies Load handles missing servers gracefully.
func TestLoad_EmptyServersSection(t *testing.T) {
	t.Parallel()

	const emptyServersYAML = `
lsp:
  servers: {}
`
	path := writeTempYAML(t, emptyServersYAML)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load with empty servers: error = %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("Servers count = %d, want 0", len(cfg.Servers))
	}
}

// TestLoad_NoLSPSection verifies Load handles YAML with no lsp key.
func TestLoad_NoLSPSection(t *testing.T) {
	t.Parallel()

	const noLSPYAML = `
other_key: value
`
	path := writeTempYAML(t, noLSPYAML)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load with no lsp section: error = %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("Servers count = %d, want 0", len(cfg.Servers))
	}
}

// TestMergeInitOptions_NilOptions verifies MergeInitOptions handles nil gracefully.
func TestMergeInitOptions_NilOptions(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"existing": "value"}
	result := config.MergeInitOptions(dst, nil)
	if len(result) != 1 {
		t.Errorf("result length = %d, want 1", len(result))
	}
	if result["existing"] != "value" {
		t.Errorf("result[existing] = %v, want 'value'", result["existing"])
	}
}

// TestMergeInitOptions_OverridesExisting verifies src values override dst values.
func TestMergeInitOptions_OverridesExisting(t *testing.T) {
	t.Parallel()

	dst := map[string]any{
		"key1": "old",
		"key2": false,
	}
	src := map[string]any{
		"key1": "new",
		"key3": 42,
	}
	result := config.MergeInitOptions(dst, src)

	if result["key1"] != "new" {
		t.Errorf("result[key1] = %v, want 'new' (src overrides dst)", result["key1"])
	}
	if result["key2"] != false {
		t.Errorf("result[key2] = %v, want false (dst preserved)", result["key2"])
	}
	if result["key3"] != 42 {
		t.Errorf("result[key3] = %v, want 42 (src added)", result["key3"])
	}
}

// TestMergeInitOptions_NilDst verifies MergeInitOptions handles nil dst.
func TestMergeInitOptions_NilDst(t *testing.T) {
	t.Parallel()

	src := map[string]any{"a": 1}
	result := config.MergeInitOptions(nil, src)
	if result["a"] != 1 {
		t.Errorf("result[a] = %v, want 1", result["a"])
	}
}

// TestMergeInitOptions_BothNil verifies MergeInitOptions with both nil returns empty map.
func TestMergeInitOptions_BothNil(t *testing.T) {
	t.Parallel()

	result := config.MergeInitOptions(nil, nil)
	if result == nil {
		t.Error("MergeInitOptions(nil, nil) = nil, want non-nil empty map")
	}
	if len(result) != 0 {
		t.Errorf("result length = %d, want 0", len(result))
	}
}

// TestLoad_InstallHint verifies install_hint is parsed from lsp.yaml (REQ-LM-004).
func TestLoad_InstallHint(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleLSPYAMLWithExtensions)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	py, ok := cfg.Servers["python"]
	if !ok {
		t.Fatal("Servers missing 'python'")
	}
	if py.InstallHint != "pip install python-lsp-server" {
		t.Errorf("python.InstallHint = %q, want 'pip install python-lsp-server'", py.InstallHint)
	}

	go_, ok := cfg.Servers["go"]
	if !ok {
		t.Fatal("Servers missing 'go'")
	}
	if go_.InstallHint != "go install golang.org/x/tools/gopls@latest" {
		t.Errorf("go.InstallHint = %q, want install command", go_.InstallHint)
	}
}

// TestLoad_FallbackBinaries verifies fallback_binaries are parsed from lsp.yaml (REQ-LM-008).
func TestLoad_FallbackBinaries(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleLSPYAMLWithExtensions)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	py, ok := cfg.Servers["python"]
	if !ok {
		t.Fatal("Servers missing 'python'")
	}
	if len(py.FallbackBinaries) != 2 {
		t.Errorf("python.FallbackBinaries length = %d, want 2", len(py.FallbackBinaries))
	}
	if len(py.FallbackBinaries) > 0 && py.FallbackBinaries[0] != "pyright-langserver" {
		t.Errorf("python.FallbackBinaries[0] = %q, want 'pyright-langserver'", py.FallbackBinaries[0])
	}
}

// TestLoad_ProjectMarkers verifies project_markers are parsed from lsp.yaml (REQ-LM-001).
func TestLoad_ProjectMarkers(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleLSPYAMLWithExtensions)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	py, ok := cfg.Servers["python"]
	if !ok {
		t.Fatal("Servers missing 'python'")
	}
	if len(py.ProjectMarkers) != 2 {
		t.Errorf("python.ProjectMarkers length = %d, want 2", len(py.ProjectMarkers))
	}
	if len(py.ProjectMarkers) > 0 && py.ProjectMarkers[0] != "pyproject.toml" {
		t.Errorf("python.ProjectMarkers[0] = %q, want 'pyproject.toml'", py.ProjectMarkers[0])
	}
}
