package gopls

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDefaultConfig verifies that DefaultConfig returns reasonable default values.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() must not return nil")
	}
	if cfg.Binary != "gopls" {
		t.Errorf("default Binary = %q, want %q", cfg.Binary, "gopls")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("default Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
	if cfg.Enabled {
		t.Error("default Enabled must be false")
	}
}

// TestLoadConfig_ValidYAML verifies that config is loaded from a valid YAML file.
func TestLoadConfig_ValidYAML(t *testing.T) {
	yaml := `lsp:
    gopls_bridge:
        enabled: true
        binary: /usr/local/bin/gopls
        args:
            - "-remote=auto"
        init_options:
            staticcheck: true
        timeouts:
            initialize_seconds: 60
            request_seconds: 10
            shutdown_seconds: 10
            diagnostics_debounce_ms: 200
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, want true")
	}
	if cfg.Binary != "/usr/local/bin/gopls" {
		t.Errorf("Binary = %q, want %q", cfg.Binary, "/usr/local/bin/gopls")
	}
	if len(cfg.Args) != 1 || cfg.Args[0] != "-remote=auto" {
		t.Errorf("Args = %v, want [-remote=auto]", cfg.Args)
	}
	// validate timeouts
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout (request) = %v, want %v", cfg.Timeout, 10*time.Second)
	}
	if cfg.InitTimeout != 60*time.Second {
		t.Errorf("InitTimeout = %v, want %v", cfg.InitTimeout, 60*time.Second)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, 10*time.Second)
	}
	if cfg.DebounceWindow != 200*time.Millisecond {
		t.Errorf("DebounceWindow = %v, want %v", cfg.DebounceWindow, 200*time.Millisecond)
	}
	// validate init_options
	if v, ok := cfg.InitOptions["staticcheck"]; !ok || v != true {
		t.Errorf("InitOptions[staticcheck] = %v, want true", v)
	}
}

// TestLoadConfig_MissingFile verifies that default values are returned when the file is absent.
// REQ-GB-002: missing file is not an error; return defaults instead.
func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/lsp.yaml")
	if err != nil {
		t.Fatalf("missing file returned error: %v (wanted default values)", err)
	}
	if cfg == nil {
		t.Fatal("missing file returned nil (wanted default values)")
	}
	// verify defaults
	if cfg.Binary != "gopls" {
		t.Errorf("default Binary = %q, want gopls", cfg.Binary)
	}
}

// TestLoadConfig_InvalidYAML verifies that an error is returned for an invalid YAML file.
func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(configPath, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("invalid YAML did not return an error")
	}
}

// TestLoadConfig_PartialYAML verifies that default values fill in unspecified fields in a partial YAML.
func TestLoadConfig_PartialYAML(t *testing.T) {
	yaml := `lsp:
    gopls_bridge:
        enabled: true
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, want true")
	}
	// unspecified fields use defaults
	if cfg.Binary != "gopls" {
		t.Errorf("partial YAML: Binary = %q, want default gopls", cfg.Binary)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("partial YAML: Timeout = %v, want default 30s", cfg.Timeout)
	}
}

// ─── F5: validateBinary / validateArgs tests ─────────────────────────────────

// TestValidateBinary_AllowsGopls verifies that the bare name "gopls" is allowed.
func TestValidateBinary_AllowsGopls(t *testing.T) {
	if err := validateBinary("gopls"); err != nil {
		t.Errorf("validateBinary(gopls) = %v, want nil", err)
	}
}

// TestValidateBinary_RejectsUntrustedPath verifies that untrusted absolute paths are rejected.
func TestValidateBinary_RejectsUntrustedPath(t *testing.T) {
	err := validateBinary("/tmp/evil/gopls")
	if err == nil {
		t.Error("validateBinary(/tmp/evil/gopls) did not return an error")
	}
}

// TestValidateBinary_AllowsTrustedPrefix verifies that trusted prefix paths are allowed.
func TestValidateBinary_AllowsTrustedPrefix(t *testing.T) {
	trustedPaths := []string{
		"/usr/bin/gopls",
		"/usr/local/bin/gopls",
		"/opt/homebrew/bin/gopls",
	}
	// $HOME/go/bin/gopls
	if home := os.Getenv("HOME"); home != "" {
		trustedPaths = append(trustedPaths, filepath.Join(home, "go", "bin", "gopls"))
		trustedPaths = append(trustedPaths, filepath.Join(home, ".local", "bin", "gopls"))
	}

	for _, p := range trustedPaths {
		if err := validateBinary(p); err != nil {
			t.Errorf("validateBinary(%q) = %v, want nil", p, err)
		}
	}
}

// TestValidateBinary_RejectsTraversal verifies that path traversal attempts are rejected.
func TestValidateBinary_RejectsTraversal(t *testing.T) {
	err := validateBinary("/usr/local/bin/../../../etc/passwd")
	if err == nil {
		t.Error("validateBinary(path traversal) did not return an error")
	}
}

// TestValidateArgs_RejectsShellMetachars verifies that args containing shell metacharacters are rejected.
func TestValidateArgs_RejectsShellMetachars(t *testing.T) {
	dangerous := []string{
		"; rm -rf /",
		"| cat /etc/passwd",
		"& evil",
		"`whoami`",
		"$(id)",
		"> /tmp/out",
		"< /etc/passwd",
		"arg\nwith-newline",
	}
	for _, arg := range dangerous {
		if err := validateArgs([]string{arg}); err == nil {
			t.Errorf("validateArgs(%q) did not return an error", arg)
		}
	}
}

// TestValidateArgs_AllowsSafeArgs verifies that safe args are allowed.
func TestValidateArgs_AllowsSafeArgs(t *testing.T) {
	safe := [][]string{
		{},
		{"-remote=auto"},
		{"-rpc.trace"},
		{"-logfile=/tmp/gopls.log"},
	}
	for _, args := range safe {
		if err := validateArgs(args); err != nil {
			t.Errorf("validateArgs(%v) = %v, want nil", args, err)
		}
	}
}

// TestLoadConfig_RealLSPYaml verifies that the real .moai/config/sections/lsp.yaml can be loaded.
// Skips if the file is not present.
func TestLoadConfig_RealLSPYaml(t *testing.T) {
	// find the real file path from project root.
	// tests run from the package directory, so navigate up.
	candidates := []string{
		"../../../.moai/config/sections/lsp.yaml",
		"../../../../.moai/config/sections/lsp.yaml",
	}
	var found string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			found = c
			break
		}
	}
	if found == "" {
		t.Skip("real lsp.yaml not found, skipping test")
	}

	cfg, err := LoadConfig(found)
	if err != nil {
		t.Fatalf("loading real lsp.yaml error: %v", err)
	}
	if cfg == nil {
		t.Fatal("loading real lsp.yaml returned nil")
	}
	// real file should have enabled=false (default)
	if cfg.Binary != "gopls" {
		t.Errorf("Binary = %q, want gopls", cfg.Binary)
	}
}
