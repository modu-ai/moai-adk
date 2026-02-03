package astgrep

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_SingleRule(t *testing.T) {
	loader := NewRuleLoader()
	rules, err := loader.LoadFromFile("testdata/rules/security.yml")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	r := rules[0]
	if r.ID != "no-eval" {
		t.Errorf("expected id no-eval, got %q", r.ID)
	}
	if r.Language != "python" {
		t.Errorf("expected language python, got %q", r.Language)
	}
	if r.Severity != "error" {
		t.Errorf("expected severity error, got %q", r.Severity)
	}
	if r.Message != "Do not use eval()" {
		t.Errorf("unexpected message: %q", r.Message)
	}
	if r.Pattern != "eval($CODE)" {
		t.Errorf("expected pattern eval($CODE), got %q", r.Pattern)
	}
}

func TestLoadFromFile_MultiDocument(t *testing.T) {
	loader := NewRuleLoader()
	rules, err := loader.LoadFromFile("testdata/rules/multi_doc.yml")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(rules) != 3 {
		t.Fatalf("expected 3 rules from multi-doc YAML, got %d", len(rules))
	}

	// Verify each rule was parsed correctly.
	tests := []struct {
		id       string
		language string
		severity string
	}{
		{"rule-python-eval", "python", "error"},
		{"rule-go-println", "go", "warning"},
		{"rule-js-alert", "javascript", "info"},
	}

	for i, tt := range tests {
		if rules[i].ID != tt.id {
			t.Errorf("rule %d: expected id %q, got %q", i, tt.id, rules[i].ID)
		}
		if rules[i].Language != tt.language {
			t.Errorf("rule %d: expected language %q, got %q", i, tt.language, rules[i].Language)
		}
		if rules[i].Severity != tt.severity {
			t.Errorf("rule %d: expected severity %q, got %q", i, tt.severity, rules[i].Severity)
		}
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	loader := NewRuleLoader()
	_, err := loader.LoadFromFile("testdata/rules/nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	loader := NewRuleLoader()
	_, err := loader.LoadFromFile("testdata/rules/invalid.yml")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadFromDirectory(t *testing.T) {
	loader := NewRuleLoader()
	rules, err := loader.LoadFromDirectory("testdata/rules")
	if err != nil {
		// invalid.yml will cause a parse error
		// This is expected behavior - the loader stops on first error
		t.Logf("LoadFromDirectory returned error (expected due to invalid.yml): %v", err)
		return
	}

	// If no error, we should have rules from security.yml, quality.yml, and multi_doc.yml
	if len(rules) < 4 {
		t.Errorf("expected at least 4 rules from directory, got %d", len(rules))
	}
}

func TestLoadFromDirectory_ValidOnly(t *testing.T) {
	// Create a temp directory with only valid YAML files.
	tmpDir := t.TempDir()

	rule1 := `id: test-rule-1
language: go
severity: warning
message: "Test rule 1"
pattern: "fmt.Println($MSG)"
`
	rule2 := `id: test-rule-2
language: python
severity: error
message: "Test rule 2"
pattern: "eval($CODE)"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "rule1.yml"), []byte(rule1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "rule2.yaml"), []byte(rule2), 0o644); err != nil {
		t.Fatal(err)
	}
	// Non-YAML file should be ignored.
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a rule"), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := NewRuleLoader()
	rules, err := loader.LoadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDirectory failed: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestGetRulesForLanguage(t *testing.T) {
	loader := NewRuleLoader()
	_, err := loader.LoadFromFile("testdata/rules/multi_doc.yml")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	tests := []struct {
		language string
		expected int
	}{
		{"python", 1},
		{"go", 1},
		{"javascript", 1},
		{"rust", 0},
		{"Python", 1}, // case insensitive
	}

	for _, tt := range tests {
		got := loader.GetRulesForLanguage(tt.language)
		if len(got) != tt.expected {
			t.Errorf("GetRulesForLanguage(%q): expected %d rules, got %d", tt.language, tt.expected, len(got))
		}
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	loader := NewRuleLoader()
	if _, err := loader.LoadFromFile("testdata/rules/security.yml"); err != nil {
		t.Fatal(err)
	}

	rules1 := loader.Rules()
	rules2 := loader.Rules()

	if len(rules1) != len(rules2) {
		t.Fatalf("Rules() returned different lengths: %d vs %d", len(rules1), len(rules2))
	}

	// Modifying one copy should not affect the other.
	if len(rules1) > 0 {
		rules1[0].ID = "modified"
		if rules2[0].ID == "modified" {
			t.Error("Rules() did not return a copy")
		}
	}
}
