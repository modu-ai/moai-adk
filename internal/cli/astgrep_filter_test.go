package cli

// F1 reproduction test: filterByLang unimplemented pass-through
// Tests the Finding.Language field and filterByLang function directly.
// Written as a package-internal test (package cli) to access unexported functions.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestFilterByLang_KeepsMatchingLanguage: verifies that the language filter returns only matching items and items with no language info
func TestFilterByLang_KeepsMatchingLanguage(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Language: "go"},
		{RuleID: "r2", Language: "python"},
		{RuleID: "r3", Language: ""},
	}
	got := filterByLang(findings, "python")
	// python + empty language = 2 returned
	if len(got) != 2 {
		t.Errorf("filterByLang(python) len = %d, want 2 (python item + item with no language info)", len(got))
	}
	for _, f := range got {
		if strings.ToLower(f.Language) != "python" && f.Language != "" {
			t.Errorf("filterByLang(python) contains unexpected language: %q", f.Language)
		}
	}
}

// TestFilterByLang_EmptyLang_ReturnsAll: verifies that an empty language filter returns all items
func TestFilterByLang_EmptyLang_ReturnsAll(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Language: "go"},
		{RuleID: "r2", Language: "python"},
		{RuleID: "r3", Language: ""},
	}
	got := filterByLang(findings, "")
	if len(got) != len(findings) {
		t.Errorf("filterByLang(\"\") len = %d, want %d (return all)", len(got), len(findings))
	}
}

// TestFilterByLang_CaseInsensitive: verifies that the language filter is case-insensitive
func TestFilterByLang_CaseInsensitive(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Language: "Go"},
		{RuleID: "r2", Language: "PYTHON"},
		{RuleID: "r3", Language: ""},
	}
	got := filterByLang(findings, "go")
	// Go (uppercase) + empty language = 2
	if len(got) != 2 {
		t.Errorf("filterByLang(go) case-insensitive test: len = %d, want 2", len(got))
	}
}
