package security

import (
	"testing"
)

// enumeratedClasses are the vulnerability classes REQ-SG-012 requires the
// guardian to cover. Kept here (not imported from patterns.go) so the test is
// an independent specification of the required coverage.
var enumeratedClasses = []string{
	"unsafe-deserialization",
	"dom-injection-xss",
	"hardcoded-secret",
	"code-injection-eval",
	"sql-injection",
	"command-injection",
	"path-traversal",
	"ssrf",
	"weak-crypto",
	"insecure-random",
}

// TestPatternClassCoverage asserts every enumerated class is present and the
// total pattern count falls in the bounded range 20 <= N <= 30 (AC-SG-003,
// REQ-SG-012). The range is asserted, not a soft "approximately 25".
func TestPatternClassCoverage(t *testing.T) {
	classes := GuardianPatterns()

	present := make(map[string]bool, len(classes))
	total := 0
	for _, c := range classes {
		present[c.Name] = true
		if len(c.Patterns) == 0 {
			t.Errorf("class %q has zero patterns", c.Name)
		}
		total += len(c.Patterns)
	}

	for _, want := range enumeratedClasses {
		if !present[want] {
			t.Errorf("required vulnerability class %q missing from the pattern table", want)
		}
	}

	if total < 20 || total > 30 {
		t.Errorf("total pattern count = %d, want within bounded range 20 <= N <= 30", total)
	}
}

// TestPatternsLanguageNeutral asserts a class fires equally across the 16
// supported languages with no PRIMARY language (AC-SG-002, REQ-SG-011). The
// hardcoded-secret class is exercised because a `password = "literal"` idiom is
// legible in every one of the 16 languages.
func TestPatternsLanguageNeutral(t *testing.T) {
	// 16 fixtures: the same dangerous idiom (a hardcoded secret) written in the
	// idiomatic form of each of the 16 supported languages.
	fixtures := map[string]string{
		"go":         `apiKey := "sk-live-abc123def456"`,
		"python":     `api_key = "sk-live-abc123def456"`,
		"typescript": `const apiKey = "sk-live-abc123def456";`,
		"javascript": `const api_key = "sk-live-abc123def456";`,
		"rust":       `let api_key = "sk-live-abc123def456";`,
		"java":       `String apiKey = "sk-live-abc123def456";`,
		"kotlin":     `val apiKey = "sk-live-abc123def456"`,
		"csharp":     `var apiKey = "sk-live-abc123def456";`,
		"ruby":       `api_key = "sk-live-abc123def456"`,
		"php":        `$api_key = "sk-live-abc123def456";`,
		"elixir":     `api_key = "sk-live-abc123def456"`,
		"cpp":        `std::string apiKey = "sk-live-abc123def456";`,
		"scala":      `val apiKey = "sk-live-abc123def456"`,
		"r":          `api_key <- "sk-live-abc123def456"`,
		"flutter":    `final apiKey = "sk-live-abc123def456";`,
		"swift":      `let apiKey = "sk-live-abc123def456"`,
	}

	if len(fixtures) != 16 {
		t.Fatalf("expected 16 language fixtures, got %d", len(fixtures))
	}

	for lang, content := range fixtures {
		findings := ScanBuffer(content)
		hit := false
		for _, f := range findings {
			if f.Class == "hardcoded-secret" {
				hit = true
				break
			}
		}
		if !hit {
			t.Errorf("hardcoded-secret class did not fire for language %q (content=%q)", lang, content)
		}
	}
}

// TestPatternsNoPrimaryLanguage asserts the table is organized by class, never
// by a single PRIMARY language: no class is scoped to exactly one language
// (REQ-SG-011). A class is either language-agnostic (empty Langs) or a genuine
// multi-language subset.
func TestPatternsNoPrimaryLanguage(t *testing.T) {
	for _, c := range GuardianPatterns() {
		if len(c.Langs) == 1 {
			t.Errorf("class %q is scoped to a single language %v — violates 16-language neutrality (REQ-SG-011)", c.Name, c.Langs)
		}
	}
}

// TestPatternsFalsePositiveBaseline pins the false-positive baseline: benign
// code must not raise findings (REQ-SG-014 advisory-first posture — a false
// positive costs one advisory line, so the baseline is measured, not guessed).
func TestPatternsFalsePositiveBaseline(t *testing.T) {
	benign := []string{
		`func Add(a, b int) int { return a + b }`,
		`const greeting = "hello world";`,
		`result := computeTotal(items)`,
		`# a plain comment about passwords in general`,
		`let count = numbers.length;`,
		`db.Query("SELECT id FROM users WHERE active = ?", true)`, // parameterized — no concat
	}
	for _, code := range benign {
		if findings := ScanBuffer(code); len(findings) != 0 {
			t.Errorf("benign code raised %d finding(s) (false positive): %q -> %+v", len(findings), code, findings)
		}
	}
}
