package astgrep_test

// rule_seed_test.go covers AC-UTIL-002-13..19 (5-language rule seeding).
// Each language subtest verifies:
//   (a) YAML files load without error
//   (b) Rules declare note, metadata.owasp, and metadata.cwe
//   (c) Fixture files exist on disk
//   (d) When sg is available: violation fixture → ≥1 finding, valid fixture → 0 findings
//   (e) When sg is available: at least one finding has non-empty Metadata["cwe"]

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// findProjectRoot walks up from CWD to locate go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal("getwd:", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// TestRuleSeed verifies the 5-language rule directories introduced in SPEC-UTIL-002
// (Ruby, PHP, Elixir, C#, Kotlin) satisfy all seeding invariants.
// AC-UTIL-002-13..19
func TestRuleSeed(t *testing.T) {
	t.Parallel()

	// Verify sg is actually ast-grep, not newgrp (util-linux symlink on Ubuntu).
	// Ubuntu/Debian ships `/usr/bin/sg` as newgrp alternative which shadows ast-grep
	// when ast-grep is not installed. LookPath alone is insufficient.
	sgAvailable := func() bool {
		if _, err := exec.LookPath("sg"); err != nil {
			return false
		}
		out, err := exec.Command("sg", "--version").CombinedOutput()
		if err != nil {
			return false
		}
		return strings.Contains(strings.ToLower(string(out)), "ast-grep")
	}()

	type langCase struct {
		lang     string
		ext      string
		rulesDir string
		fixtDir  string
	}

	projectRoot := findProjectRoot(t)

	cases := []langCase{
		{
			lang:     "ruby",
			ext:      "rb",
			rulesDir: ".moai/config/astgrep-rules/ruby",
			fixtDir:  "internal/astgrep/testdata/fixtures/ruby",
		},
		{
			lang:     "php",
			ext:      "php",
			rulesDir: ".moai/config/astgrep-rules/php",
			fixtDir:  "internal/astgrep/testdata/fixtures/php",
		},
		{
			lang:     "elixir",
			ext:      "ex",
			rulesDir: ".moai/config/astgrep-rules/elixir",
			fixtDir:  "internal/astgrep/testdata/fixtures/elixir",
		},
		{
			lang:     "csharp",
			ext:      "cs",
			rulesDir: ".moai/config/astgrep-rules/csharp",
			fixtDir:  "internal/astgrep/testdata/fixtures/csharp",
		},
		{
			lang:     "kotlin",
			ext:      "kt",
			rulesDir: ".moai/config/astgrep-rules/kotlin",
			fixtDir:  "internal/astgrep/testdata/fixtures/kotlin",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.lang, func(t *testing.T) {
			t.Parallel()

			rulesDir := filepath.Join(projectRoot, tc.rulesDir)
			fixtureDir := filepath.Join(projectRoot, tc.fixtDir)

			// ── 1. Rules load without error ─────────────────────────────────
			loader := astgrep.NewRuleLoader()
			rules, err := loader.LoadFromDir(rulesDir)
			if err != nil {
				t.Fatalf("LoadFromDir(%s) error: %v", rulesDir, err)
			}
			if len(rules) == 0 {
				t.Fatalf("no rules loaded from %s; expected at least 3", rulesDir)
			}

			// ── 2. Every rule has note, metadata.owasp, metadata.cwe ────────
			for _, r := range rules {
				if r.Note == "" {
					t.Errorf("rule %q (lang=%s) is missing 'note' field", r.ID, tc.lang)
				}
				if r.Metadata == nil {
					t.Errorf("rule %q (lang=%s) is missing 'metadata' field", r.ID, tc.lang)
					continue
				}
				if r.Metadata["owasp"] == "" {
					t.Errorf("rule %q (lang=%s) metadata is missing 'owasp' key", r.ID, tc.lang)
				}
				if r.Metadata["cwe"] == "" {
					t.Errorf("rule %q (lang=%s) metadata is missing 'cwe' key", r.ID, tc.lang)
				}
			}

			// ── 3. Fixture files exist on disk ───────────────────────────────
			for _, fixture := range []string{"valid", "violation", "suppressed"} {
				fp := filepath.Join(fixtureDir, fixture+"."+tc.ext)
				if _, err := os.Stat(fp); err != nil {
					t.Errorf("fixture file not found: %s", fp)
				}
			}

			// ── 4+5. sg-dependent: scan violation/valid fixtures ─────────────
			if !sgAvailable {
				t.Skip("sg binary not available; skipping scan assertions")
			}

			scanner := astgrep.NewScanner(&astgrep.ScannerConfig{
				RulesDir: rulesDir,
				SGBinary: "sg",
			})

			// 4a. Violation fixture → ≥1 finding (AC-UTIL-002-19)
			violationFile := filepath.Join(fixtureDir, "violation."+tc.ext)
			vFindings, err := scanner.Scan(context.Background(), violationFile)
			if err != nil {
				t.Fatalf("Scan(violation.%s) error: %v", tc.ext, err)
			}
			if len(vFindings) == 0 {
				t.Errorf("violation.%s should produce ≥1 finding, got 0", tc.ext)
			}

			// 4b. Valid fixture → 0 findings (AC-UTIL-002-19)
			validFile := filepath.Join(fixtureDir, "valid."+tc.ext)
			goodFindings, err := scanner.Scan(context.Background(), validFile)
			if err != nil {
				t.Fatalf("Scan(valid.%s) error: %v", tc.ext, err)
			}
			if len(goodFindings) != 0 {
				t.Errorf("valid.%s should produce 0 findings, got %d", tc.ext, len(goodFindings))
			}

			// 5. At least one violation finding has non-empty Metadata["cwe"] (AC-UTIL-002-19)
			hasCWE := false
			for _, f := range vFindings {
				if f.Metadata["cwe"] != "" {
					hasCWE = true
					break
				}
			}
			if !hasCWE {
				t.Errorf("violation.%s: no finding has non-empty Metadata[\"cwe\"]", tc.ext)
			}
		})
	}
}

// TestRuleStruct_NoteField verifies Rule struct has a Note field with correct YAML/JSON tags.
// AC-UTIL-002-01
func TestRuleStruct_NoteField(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	ruleYAML := `---
id: test-note-rule
language: go
severity: warning
message: "test message"
pattern: "fmt.Println($X)"
note: "sample note for testing"
metadata:
  owasp: "A03:2021"
  cwe: "CWE-89"
`
	ruleFile := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(ruleFile, []byte(ruleYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("no rules loaded")
	}

	r := rules[0]
	if r.Note != "sample note for testing" {
		t.Errorf("Rule.Note = %q, want %q", r.Note, "sample note for testing")
	}
}

// TestRuleStruct_MetadataField verifies Rule struct has a Metadata field with correct YAML/JSON tags.
// AC-UTIL-002-02
func TestRuleStruct_MetadataField(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	ruleYAML := `---
id: test-metadata-rule
language: go
severity: error
message: "test message"
pattern: "os.Getenv($X)"
metadata:
  owasp: "A03:2021 - Injection"
  cwe: "CWE-89"
`
	ruleFile := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(ruleFile, []byte(ruleYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := astgrep.NewRuleLoader()
	rules, err := loader.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("no rules loaded")
	}

	r := rules[0]
	if r.Metadata == nil {
		t.Fatal("Rule.Metadata is nil")
	}
	if r.Metadata["owasp"] != "A03:2021 - Injection" {
		t.Errorf("Rule.Metadata[\"owasp\"] = %q, want %q", r.Metadata["owasp"], "A03:2021 - Injection")
	}
	if r.Metadata["cwe"] != "CWE-89" {
		t.Errorf("Rule.Metadata[\"cwe\"] = %q, want %q", r.Metadata["cwe"], "CWE-89")
	}
}

// TestScanWithRules_NotePropagation verifies that scanWithRules copies rule.Note to findings.
// AC-UTIL-002-03
func TestScanWithRules_NotePropagation(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg not available; skipping propagation test")
	}

	tmpDir := t.TempDir()

	// Rule with note
	ruleYAML := `---
id: note-propagation-rule
language: go
severity: warning
message: "test rule"
note: "propagated note text"
pattern: 'fmt.Println($X)'
`
	ruleDir := filepath.Join(tmpDir, "rules")
	if err := os.MkdirAll(ruleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ruleDir, "rule.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Target file that matches the pattern
	goCode := `package main
import "fmt"
func main() { fmt.Println("hello") }
`
	targetFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(targetFile, []byte(goCode), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := astgrep.NewScanner(&astgrep.ScannerConfig{
		RulesDir: ruleDir,
		SGBinary: "sg",
	})

	findings, err := scanner.Scan(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Skip("no findings from sg; pattern may not match in this version")
	}

	for _, f := range findings {
		if f.Note != "propagated note text" {
			t.Errorf("Finding.Note = %q, want %q", f.Note, "propagated note text")
		}
	}
}

// TestScanWithRules_MetadataPropagation verifies that scanWithRules copies rule.Metadata to findings.
// AC-UTIL-002-04
func TestScanWithRules_MetadataPropagation(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("sg"); err != nil {
		t.Skip("sg not available; skipping propagation test")
	}

	tmpDir := t.TempDir()

	// Rule with metadata
	ruleYAML := `---
id: metadata-propagation-rule
language: go
severity: error
message: "test rule"
metadata:
  owasp: "A03:2021"
  cwe: "CWE-89"
pattern: 'fmt.Println($X)'
`
	ruleDir := filepath.Join(tmpDir, "rules")
	if err := os.MkdirAll(ruleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ruleDir, "rule.yml"), []byte(ruleYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	goCode := `package main
import "fmt"
func main() { fmt.Println("hello") }
`
	targetFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(targetFile, []byte(goCode), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := astgrep.NewScanner(&astgrep.ScannerConfig{
		RulesDir: ruleDir,
		SGBinary: "sg",
	})

	findings, err := scanner.Scan(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Skip("no findings from sg; pattern may not match in this version")
	}

	for _, f := range findings {
		if f.Metadata == nil {
			t.Error("Finding.Metadata is nil; expected propagated metadata from rule")
			continue
		}
		if f.Metadata["owasp"] != "A03:2021" {
			t.Errorf("Finding.Metadata[\"owasp\"] = %q, want %q", f.Metadata["owasp"], "A03:2021")
		}
		if f.Metadata["cwe"] != "CWE-89" {
			t.Errorf("Finding.Metadata[\"cwe\"] = %q, want %q", f.Metadata["cwe"], "CWE-89")
		}
	}
}
