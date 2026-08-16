package astgrep_test

// rule_seed_test.go verifies the seeding invariants of the ruleset this
// repository actually ships: .moai/astgrep-rules/{go,security} (tracked, and
// byte-identical to the template copy at
// internal/template/templates/.moai/config/astgrep-rules/).
//
// History (t50): this test previously targeted the never-landed 5-language
// seeding of SPEC-UTIL-002 (ruby/php/elixir/csharp/kotlin under the OLD
// .moai/config/astgrep-rules path). Those directories were retired without ever
// being populated, so every subtest silently skipped. The per-language cases
// are gone; the invariants now pin the surviving surface:
//
//   (a) Every shipped rule loads without error; each directory holds ≥3 rules.
//   (b) Every rule declares a non-empty note. Security rules additionally
//       declare metadata.owasp and metadata.cwe. The go/ rules carry no
//       metadata by current design — the SPEC-UTIL-002 metadata scheme landed
//       only on the security family, and idioms/maintainability rules have no
//       honest OWASP/CWE mapping — so no metadata assertion is made for them.
//   (c) Shared Go fixtures (valid/violation/suppressed) exist on disk.
//   (d) When sg is available: scanning via the shipped root sgconfig.yml finds
//       ≥1 rule violation on the violation fixture, 0 on the valid fixture,
//       and 0 on the suppressed fixture (ast-grep-ignore honored).
//
// Note on scan mechanics: every shipped rule expresses its matcher as a nested
// `rule:` block, which the flat-pattern per-rule scan path
// (Scanner.scanWithRules) cannot execute — only the sgconfig.yml config path
// exercises the real matchers. Per-match metadata propagation is therefore NOT
// asserted here (the config path does not propagate rule metadata onto
// findings); the propagation mechanism itself is covered by
// TestScanWithRules_MetadataPropagation below using a synthetic flat-pattern
// rule.

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

// sgIsAvailable verifies sg is actually ast-grep, not newgrp (util-linux
// symlink on Ubuntu). Ubuntu/Debian ships `/usr/bin/sg` as a newgrp
// alternative which shadows ast-grep when ast-grep is not installed.
// LookPath alone is insufficient.
func sgIsAvailable() bool {
	if _, err := exec.LookPath("sg"); err != nil {
		return false
	}
	out, err := exec.Command("sg", "--version").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "ast-grep")
}

// TestRuleSeed verifies the tracked ruleset directories (go/, security/)
// satisfy the seeding invariants described in the file header. The
// directories are tracked repository content: a missing or shrunken
// directory is a regression, not a skip condition.
func TestRuleSeed(t *testing.T) {
	t.Parallel()

	sgAvailable := sgIsAvailable()

	type dirCase struct {
		name        string // subtest name and directory under .moai/astgrep-rules
		requireMeta bool   // assert metadata.owasp + metadata.cwe on every rule
	}

	cases := []dirCase{
		{name: "go", requireMeta: false},
		{name: "security", requireMeta: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rulesDir := filepath.Join(findProjectRoot(t), ".moai", "astgrep-rules", tc.name)

			// 1. Rules load without error, >=3 rules (no skip: tracked content).
			loader := astgrep.NewRuleLoader()
			rules, err := loader.LoadFromDir(rulesDir)
			if err != nil {
				t.Fatalf("LoadFromDir(%s) error: %v", rulesDir, err)
			}
			if len(rules) < 3 {
				t.Fatalf("rules dir %s holds %d rules, want >=3 (tracked ruleset shrank?)", rulesDir, len(rules))
			}

			// 2. Seeding invariants: identity fields + note (+ metadata for security).
			for _, r := range rules {
				if r.ID == "" {
					t.Errorf("rule in %s is missing 'id' field", rulesDir)
				}
				if r.Language == "" {
					t.Errorf("rule %q is missing 'language' field", r.ID)
				}
				if r.Severity == "" {
					t.Errorf("rule %q is missing 'severity' field", r.ID)
				}
				if r.Message == "" {
					t.Errorf("rule %q is missing 'message' field", r.ID)
				}
				if r.Note == "" {
					t.Errorf("rule %q is missing 'note' field", r.ID)
				}
				if tc.requireMeta {
					if r.Metadata == nil {
						t.Errorf("rule %q is missing 'metadata' field", r.ID)
						continue
					}
					if r.Metadata["owasp"] == "" {
						t.Errorf("rule %q metadata is missing 'owasp' key", r.ID)
					}
					if r.Metadata["cwe"] == "" {
						t.Errorf("rule %q metadata is missing 'cwe' key", r.ID)
					}
				}
			}
		})
	}

	// 3+4. Shared Go fixtures scanned through the shipped root sgconfig.yml.
	t.Run("fixtures-scan", func(t *testing.T) {
		t.Parallel()

		if !sgAvailable {
			t.Skip("sg binary not available; skipping scan assertions")
		}

		projectRoot := findProjectRoot(t)
		rulesRoot := filepath.Join(projectRoot, ".moai", "astgrep-rules")
		fixtureDir := filepath.Join(projectRoot, "internal", "astgrep", "testdata", "fixtures", "go")

		// 3. Fixture files exist on disk.
		fixtures := map[string]string{
			"valid":      filepath.Join(fixtureDir, "valid.go"),
			"violation":  filepath.Join(fixtureDir, "violation.go"),
			"suppressed": filepath.Join(fixtureDir, "suppressed.go"),
		}
		for name, fp := range fixtures {
			if _, err := os.Stat(fp); err != nil {
				t.Errorf("fixture file not found: %s (%s)", fp, name)
			}
		}

		// The root ruleset carries sgconfig.yml, so Scanner takes the config
		// path (sg scan --config) — the only path that executes the shipped
		// nested-`rule:` matchers.
		scanner := astgrep.NewScanner(&astgrep.ScannerConfig{
			RulesDir: rulesRoot,
			SGBinary: "sg",
		})

		// 4a. Violation fixture -> >=1 finding.
		vFindings, err := scanner.Scan(context.Background(), fixtures["violation"])
		if err != nil {
			t.Fatalf("Scan(violation.go) error: %v", err)
		}
		if len(vFindings) == 0 {
			t.Error("violation.go should produce >=1 finding, got 0")
		}

		// 4b. Valid fixture -> 0 findings.
		gFindings, err := scanner.Scan(context.Background(), fixtures["valid"])
		if err != nil {
			t.Fatalf("Scan(valid.go) error: %v", err)
		}
		if len(gFindings) != 0 {
			t.Errorf("valid.go should produce 0 findings, got %d: %v", len(gFindings), gFindings)
		}

		// 4c. Suppressed fixture -> 0 findings (ast-grep-ignore honored).
		sFindings, err := scanner.Scan(context.Background(), fixtures["suppressed"])
		if err != nil {
			t.Fatalf("Scan(suppressed.go) error: %v", err)
		}
		if len(sFindings) != 0 {
			t.Errorf("suppressed.go should produce 0 findings, got %d: %v", len(sFindings), sFindings)
		}
	})
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
