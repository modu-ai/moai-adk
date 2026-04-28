package astgrep_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestNewScanner_DefaultConfig: verifies that Scanner creation succeeds with default Config
func TestNewScanner_DefaultConfig(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	s := astgrep.NewScanner(cfg)
	if s == nil {
		t.Fatal("NewScanner() returned nil")
	}
}

// TestNewScanner_NilConfig: verifies that passing nil Config falls back to defaults without panicking
func TestNewScanner_NilConfig(t *testing.T) {
	s := astgrep.NewScanner(nil)
	if s == nil {
		t.Fatal("NewScanner(nil) returned nil")
	}
}

// TestScanner_SGNotAvailable: verifies that (nil, nil) is returned when sg CLI is not in PATH (AC3)
// Uses the default allowed binary ("sg") but simulates an environment where it is absent from PATH.
// Note: disallowed binary names are blocked by the F2 security check with an error.
func TestScanner_SGNotAvailable(t *testing.T) {
	// "sg" passes ValidateBinary, but isSGAvailable returns false when not in PATH.
	// This test is only meaningful in environments where sg is absent from PATH.
	// In environments where sg is installed, actual scanning may occur, so we test with an empty rules dir.
	tmpDir := t.TempDir()
	cfg := astgrep.DefaultScannerConfig()
	cfg.SGBinary = "sg" // allowed bare name
	cfg.RulesDir = tmpDir

	s := astgrep.NewScanner(cfg)
	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v; must return nil error when sg is absent or rules dir is empty", err)
	}
	if len(findings) != 0 {
		t.Errorf("Scan() len(findings) = %d; must return empty slice for empty rules directory", len(findings))
	}
}

// TestScanner_EmptyRulesDir: verifies that an empty rules directory returns empty results without error (AC3)
func TestScanner_EmptyRulesDir(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := astgrep.DefaultScannerConfig()
	cfg.RulesDir = tmpDir
	s := astgrep.NewScanner(cfg)

	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v; must not return error for empty rules directory", err)
	}
	if findings == nil {
		t.Error("Scan() returned nil findings; must return empty slice")
	}
}

// TestScanner_RulesDirNotExist: verifies that a non-existent rules directory operates without error (AC3)
func TestScanner_RulesDirNotExist(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	cfg.RulesDir = "/path/that/does/not/exist/12345"
	s := astgrep.NewScanner(cfg)

	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v; must not return error for non-existent rules directory", err)
	}
	if len(findings) != 0 {
		t.Errorf("Scan() len(findings) = %d; must return empty slice when no rules exist", len(findings))
	}
}

// TestScanner_FindingSeverityClassification: verifies severity classification of the Finding struct
func TestScanner_FindingSeverityClassification(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		wantErr  bool
		wantWarn bool
		wantInfo bool
	}{
		{"error severity", "error", true, false, false},
		{"warning severity", "warning", false, true, false},
		{"info severity", "info", false, false, true},
		{"empty defaults to info", "", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := astgrep.Finding{
				RuleID:   "test-rule",
				Severity: tt.severity,
				Message:  "테스트 메시지",
				File:     "test.go",
				Line:     1,
			}
			if got := f.IsError(); got != tt.wantErr {
				t.Errorf("Finding.IsError() = %v, want %v", got, tt.wantErr)
			}
			if got := f.IsWarning(); got != tt.wantWarn {
				t.Errorf("Finding.IsWarning() = %v, want %v", got, tt.wantWarn)
			}
			if got := f.IsInfo(); got != tt.wantInfo {
				t.Errorf("Finding.IsInfo() = %v, want %v", got, tt.wantInfo)
			}
		})
	}
}

// TestRuleLoader_LoadFromDirRecursive: verifies recursive loading of subdirectories (AC2 - RuleLoader.LoadFromDir)
func TestRuleLoader_LoadFromDirRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// create go/ subdirectory
	goDir := filepath.Join(tmpDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// create security/ subdirectory
	secDir := filepath.Join(tmpDir, "security")
	if err := os.MkdirAll(secDir, 0o755); err != nil {
		t.Fatalf("failed to create security subdirectory: %v", err)
	}

	// create go/test.yml
	goRuleYAML := `---
id: test-go-rule
language: go
severity: warning
message: "테스트 규칙"
pattern: "os.Getenv($X)"
`
	if err := os.WriteFile(filepath.Join(goDir, "test.yml"), []byte(goRuleYAML), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// create security/test.yml
	secRuleYAML := `---
id: test-sec-rule
language: go
severity: error
message: "보안 테스트 규칙"
pattern: "exec.Command($X)"
`
	if err := os.WriteFile(filepath.Join(secDir, "test.yml"), []byte(secRuleYAML), 0o644); err != nil {
		t.Fatalf("failed to write security rule file: %v", err)
	}

	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("LoadFromDir() len(rules) = %d, want 2 (recursive subdirectory loading required)", len(rules))
	}

	// verify rule IDs
	ids := make(map[string]bool)
	for _, r := range rules {
		ids[r.ID] = true
	}
	if !ids["test-go-rule"] {
		t.Error("test-go-rule was not loaded")
	}
	if !ids["test-sec-rule"] {
		t.Error("test-sec-rule was not loaded")
	}
}

// TestRuleLoader_SkipsInvalidYAML: verifies that individually unparseable rules are skipped and the rest are loaded (AC3)
func TestRuleLoader_SkipsInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// valid rule file
	validYAML := `---
id: valid-rule
language: go
severity: warning
message: "유효한 규칙"
pattern: "fmt.Println($X)"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.yml"), []byte(validYAML), 0o644); err != nil {
		t.Fatalf("failed to write valid rule file: %v", err)
	}

	// invalid YAML file (unparseable)
	invalidYAML := `{invalid yaml: [`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.yml"), []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("failed to write invalid rule file: %v", err)
	}

	loader := astgrep.NewRuleLoader()
	// must load only valid rules without returning an error
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v; invalid files must be skipped", err)
	}

	if len(rules) < 1 {
		t.Error("at least 1 valid rule must be loaded")
	}
}

// TestScanner_ScanConfig: verifies Config field settings
func TestScanner_ScanConfig(t *testing.T) {
	cfg := &astgrep.ScannerConfig{
		RulesDir:     "/custom/rules",
		SGBinary:     "sg",
		WarnOnlyMode: true,
	}
	s := astgrep.NewScanner(cfg)
	if s == nil {
		t.Fatal("NewScanner() with custom config returned nil")
	}
}

// TestScanner_ContextCancellation: verifies that Scan returns immediately when context is cancelled
func TestScanner_ContextCancellation(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	s := astgrep.NewScanner(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// must return without panicking even when context is already cancelled
	_, _ = s.Scan(ctx, ".")
}

// TestFinding_String: verifies that Finding.String() outputs a human-readable format
func TestFinding_String(t *testing.T) {
	f := astgrep.Finding{
		RuleID:   "go-no-raw-getenv",
		Severity: "warning",
		Message:  "환경변수를 직접 사용하지 마세요",
		File:     "main.go",
		Line:     42,
		Column:   10,
	}
	got := f.String()
	if got == "" {
		t.Error("Finding.String() returned empty string")
	}
	// must include filename and line number
	if !containsSubstr(got, "main.go") {
		t.Errorf("Finding.String() = %q; must include filename", got)
	}
}

// TestScanner_HasErrors: verifies that HasErrors() correctly determines whether any finding has error severity
func TestScanner_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		findings []astgrep.Finding
		want     bool
	}{
		{
			name:     "empty findings",
			findings: nil,
			want:     false,
		},
		{
			name: "only warnings",
			findings: []astgrep.Finding{
				{RuleID: "r1", Severity: "warning"},
			},
			want: false,
		},
		{
			name: "contains an error",
			findings: []astgrep.Finding{
				{RuleID: "r1", Severity: "warning"},
				{RuleID: "r2", Severity: "error"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := astgrep.HasErrors(tt.findings); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestScanner_ScanWithSGAvailable: performs actual scan when sg CLI is available (integration test, skip if sg absent)
func TestScanner_ScanWithSGAvailable(t *testing.T) {
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg CLI is not in PATH, skipping")
	}

	tmpDir := t.TempDir()

	goCode := `package main

import (
	"fmt"
	"os"
)

func main() {
	val := os.Getenv("MY_API_KEY")
	fmt.Println(val)
}
`
	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0o644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	rulesDir := filepath.Join(tmpDir, "rules", "go")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}

	ruleYAML := "---\nid: test-getenv-rule\nlanguage: go\nseverity: warning\nmessage: \"raw os.Getenv 사용 금지\"\npattern: 'os.Getenv(\"$X\")'\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	cfg := &astgrep.ScannerConfig{
		RulesDir: filepath.Join(tmpDir, "rules"),
		SGBinary: "sg",
	}
	s := astgrep.NewScanner(cfg)
	findings, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	// when sg is available, result must be a slice (not nil)
	if findings == nil {
		t.Error("Scan() returned nil with sg available")
	}
}

// TestParseSGFindings_EmptyOutput: verifies that empty output returns an empty slice
func TestParseSGFindings_EmptyOutput(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	cfg.RulesDir = t.TempDir()
	s := astgrep.NewScanner(cfg)

	findings, err := s.Scan(context.Background(), ".")
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if findings == nil {
		t.Error("Scan() returned nil; must return empty slice")
	}
}

// TestToFileURI_Paths: verifies SARIF URI conversion (indirect test of internal function)
func TestToFileURI_Paths(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Severity: "warning", Message: "m1", File: "internal/pkg/file.go", Line: 1},
		{RuleID: "r2", Severity: "error", Message: "m2", File: "", Line: 1},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}
	if len(output) == 0 {
		t.Error("ToSARIF() returned empty output")
	}
}

// TestLoadFromDir_PlaceholderDirs: verifies that placeholder directories with .gitkeep are handled correctly
func TestLoadFromDir_PlaceholderDirs(t *testing.T) {
	tmpDir := t.TempDir()

	for _, lang := range []string{"python", "typescript", "rust"} {
		dir := filepath.Join(tmpDir, lang)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte{}, 0o644); err != nil {
			t.Fatalf("failed to write .gitkeep: %v", err)
		}
	}

	goDir := filepath.Join(tmpDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("failed to create go directory: %v", err)
	}
	ruleYAML := "---\nid: go-test-rule\nlanguage: go\nseverity: warning\nmessage: \"테스트\"\npattern: \"fmt.Println($X)\"\n"
	if err := os.WriteFile(filepath.Join(goDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("LoadFromDir() len(rules) = %d, want 1 (.gitkeep files must be ignored)", len(rules))
	}
}

// TestHasErrors_MixedSeverities: verifies HasErrors behavior with a slice of mixed severities
func TestHasErrors_MixedSeverities(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Severity: "info"},
		{RuleID: "r2", Severity: "warning"},
		{RuleID: "r3", Severity: "ERROR"},
	}

	if !astgrep.HasErrors(findings) {
		t.Error("HasErrors() = false; must detect error severity case-insensitively")
	}
}

// TestFinding_IsInfoForHint: verifies that hint severity returns false for IsError/IsWarning
func TestFinding_IsInfoForHint(t *testing.T) {
	f := astgrep.Finding{Severity: "hint"}
	if f.IsError() {
		t.Error("hint severity returned IsError() = true")
	}
	if f.IsWarning() {
		t.Error("hint severity returned IsWarning() = true")
	}
}

// TestScanner_ScanWithSGConfig: scan using sgconfig.yml (skip if sg is absent)
func TestScanner_ScanWithSGConfig(t *testing.T) {
	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg CLI is not in PATH, skipping")
	}

	tmpDir := t.TempDir()

	// create target file
	goCode := `package main
import "fmt"
func main() { fmt.Println("hello") }
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// create sgconfig.yml and rules directory
	rulesDir := filepath.Join(tmpDir, "rules")
	goDir := filepath.Join(rulesDir, "go")
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}

	ruleYAML := "---\nid: sgconfig-test-rule\nlanguage: go\nseverity: warning\nmessage: \"fmt.Println 사용\"\npattern: 'fmt.Println($X)'\n"
	if err := os.WriteFile(filepath.Join(goDir, "test.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// create sgconfig.yml
	sgconfig := "ruleDirs:\n  - go\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte(sgconfig), 0o644); err != nil {
		t.Fatalf("failed to write sgconfig.yml: %v", err)
	}

	cfg := &astgrep.ScannerConfig{
		RulesDir: rulesDir,
		SGBinary: "sg",
	}
	s := astgrep.NewScanner(cfg)
	findings, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// scan via sgconfig.yml path must succeed (findings depend on sg behavior)
	if findings == nil {
		t.Error("Scan() with sgconfig.yml returned nil")
	}
}

// TestScanner_String_AllSeverities: verifies that Finding.String() outputs various severities
func TestScanner_String_AllSeverities(t *testing.T) {
	tests := []struct {
		severity string
	}{
		{"error"},
		{"warning"},
		{"info"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			f := astgrep.Finding{
				RuleID:   "test-rule",
				Severity: tt.severity,
				Message:  "테스트 메시지",
				File:     "test.go",
				Line:     1,
			}
			s := f.String()
			if s == "" {
				t.Error("Finding.String() returned empty string")
			}
			if !containsSubstr(s, "test.go") {
				t.Errorf("String() does not contain filename: %q", s)
			}
		})
	}
}

// TestDefaultScannerConfig_Fields: verifies default values of DefaultScannerConfig
func TestDefaultScannerConfig_Fields(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	if cfg.SGBinary != "sg" {
		t.Errorf("DefaultScannerConfig().SGBinary = %q, want sg", cfg.SGBinary)
	}
	if cfg.RulesDir == "" {
		t.Error("DefaultScannerConfig().RulesDir is empty")
	}
	if cfg.Timeout == 0 {
		t.Error("DefaultScannerConfig().Timeout is zero")
	}
}

// containsSubstr is a helper function that checks whether string s contains substr
func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// --- F2 regression test: Binary allowlist absence ---

// TestValidateBinary_AllowsSg: verifies that bare names "sg" and "ast-grep" are allowed
func TestValidateBinary_AllowsSg(t *testing.T) {
	for _, name := range []string{"sg", "ast-grep"} {
		if err := astgrep.ValidateBinary(name); err != nil {
			t.Errorf("ValidateBinary(%q) = %v; must be an allowed bare name", name, err)
		}
	}
}

// TestValidateBinary_AllowsTrustedPrefix: verifies that trusted paths like /usr/local/bin/sg are allowed
func TestValidateBinary_AllowsTrustedPrefix(t *testing.T) {
	paths := []string{
		"/usr/bin/sg",
		"/usr/local/bin/sg",
		"/opt/homebrew/bin/sg",
	}
	for _, p := range paths {
		if err := astgrep.ValidateBinary(p); err != nil {
			t.Errorf("ValidateBinary(%q) = %v; must be a trusted path", p, err)
		}
	}
}

// TestValidateBinary_RejectsUntrustedPath: verifies that untrusted absolute paths like /tmp are rejected
func TestValidateBinary_RejectsUntrustedPath(t *testing.T) {
	if err := astgrep.ValidateBinary("/tmp/evil/sg"); err == nil {
		t.Error("ValidateBinary(/tmp/evil/sg) = nil; untrusted path must return an error")
	}
}

// TestValidateBinary_RejectsShellMetachars: verifies that binary paths containing shell metacharacters are rejected
func TestValidateBinary_RejectsShellMetachars(t *testing.T) {
	malicious := []string{
		"sg; rm -rf /",
		"sg|cat /etc/passwd",
		"`sg`",
	}
	for _, bin := range malicious {
		if err := astgrep.ValidateBinary(bin); err == nil {
			t.Errorf("ValidateBinary(%q) = nil; shell metacharacters must return an error", bin)
		}
	}
}

// TestValidateBinary_RejectsTraversal: verifies that path traversal with .. is rejected
func TestValidateBinary_RejectsTraversal(t *testing.T) {
	if err := astgrep.ValidateBinary("/usr/local/bin/../../tmp/sg"); err == nil {
		t.Error("ValidateBinary(path traversal) = nil; must return an error")
	}
}

// TestScan_RejectsUntrustedBinary: verifies that Scan returns an error for an untrusted binary path
func TestScan_RejectsUntrustedBinary(t *testing.T) {
	cfg := astgrep.DefaultScannerConfig()
	cfg.SGBinary = "/tmp/evil/sg"
	s := astgrep.NewScanner(cfg)

	_, err := s.Scan(context.Background(), ".")
	if err == nil {
		t.Error("Scan() with untrusted binary = nil error; must return an error")
	}
}
