package convention

import (
	"regexp"
	"strings"
	"testing"
)

func TestValidate_NilConvention(t *testing.T) {
	result := Validate("any message", nil)
	if !result.Valid {
		t.Error("Validate with nil convention should be valid")
	}
	if result.Message != "any message" {
		t.Errorf("Message = %q, want %q", result.Message, "any message")
	}
}

func TestValidate_EmptyMessage(t *testing.T) {
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^.+`),
		MaxLength: 100,
	}

	tests := []struct {
		name    string
		message string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"newline only", "\n"},
		{"whitespace and newline", "  \n  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.message, conv)
			if result.Valid {
				t.Error("Validate with empty message should be invalid")
			}
			if len(result.Violations) == 0 {
				t.Error("expected at least one violation")
			}
			if result.Violations[0].Type != ViolationRequired {
				t.Errorf("violation type = %q, want %q", result.Violations[0].Type, ViolationRequired)
			}
		})
	}
}

func TestValidate_ValidConventionalCommits(t *testing.T) {
	conv, err := ParseBuiltin("conventional-commits")
	if err != nil {
		t.Fatalf("ParseBuiltin: %v", err)
	}

	tests := []struct {
		name    string
		message string
	}{
		{"feat with scope", "feat(auth): add JWT token validation"},
		{"fix without scope", "fix: resolve null pointer in user service"},
		{"docs with scope", "docs(readme): update installation guide"},
		{"chore no scope", "chore: update dependencies"},
		{"feat with breaking", "feat(api)!: change response format"},
		{"refactor", "refactor: simplify error handling"},
		{"test with scope", "test(unit): add coverage for parser"},
		{"build", "build: update CI pipeline"},
		{"ci", "ci: add deployment step"},
		{"perf", "perf: optimize query execution"},
		{"style", "style: format code with gofmt"},
		{"revert", "revert: undo last commit"},
		{"multiline body", "feat(auth): add JWT\n\nDetailed description here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.message, conv)
			if !result.Valid {
				t.Errorf("Validate(%q) should be valid, violations: %v", tt.message, result.Violations)
			}
		})
	}
}

func TestValidate_InvalidType(t *testing.T) {
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^(feat|fix)(\(.+\))?: .+`),
		Types:     []string{"feat", "fix"},
		MaxLength: 100,
	}

	// "docs" matches the general pattern but is not in the Types list.
	// We need a pattern that allows "docs" to match but Types list restricts it.
	conv.Pattern = regexp.MustCompile(`^(\w+)(\(.+\))?: .+`)

	result := Validate("docs: update readme", conv)
	if result.Valid {
		t.Error("expected invalid result for unrecognized type")
	}

	found := false
	for _, v := range result.Violations {
		if v.Type == ViolationInvalidType {
			found = true
			if v.Actual != "docs" {
				t.Errorf("violation Actual = %q, want %q", v.Actual, "docs")
			}
		}
	}
	if !found {
		t.Error("expected ViolationInvalidType violation")
	}
}

func TestValidate_InvalidScope(t *testing.T) {
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^(feat|fix)(\(.+\))?: .+`),
		Types:     []string{"feat", "fix"},
		Scopes:    []string{"api", "cli"},
		MaxLength: 100,
	}

	result := Validate("feat(unknown): add feature", conv)
	if result.Valid {
		t.Error("expected invalid result for unrecognized scope")
	}

	found := false
	for _, v := range result.Violations {
		if v.Type == ViolationInvalidScope {
			found = true
			if v.Actual != "unknown" {
				t.Errorf("violation Actual = %q, want %q", v.Actual, "unknown")
			}
		}
	}
	if !found {
		t.Error("expected ViolationInvalidScope violation")
	}
}

func TestValidate_ValidScope(t *testing.T) {
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^(feat|fix)(\(.+\))?: .+`),
		Types:     []string{"feat", "fix"},
		Scopes:    []string{"api", "cli"},
		MaxLength: 100,
	}

	result := Validate("feat(api): add endpoint", conv)
	if !result.Valid {
		t.Errorf("expected valid result, violations: %v", result.Violations)
	}
}

func TestValidate_MaxLengthExceeded(t *testing.T) {
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^.+`),
		MaxLength: 20,
	}

	long := "feat: " + strings.Repeat("a", 20)
	result := Validate(long, conv)
	if result.Valid {
		t.Error("expected invalid result for max length exceeded")
	}

	found := false
	for _, v := range result.Violations {
		if v.Type == ViolationMaxLength {
			found = true
		}
	}
	if !found {
		t.Error("expected ViolationMaxLength violation")
	}
}

func TestValidate_PatternMismatch(t *testing.T) {
	conv, err := ParseBuiltin("conventional-commits")
	if err != nil {
		t.Fatalf("ParseBuiltin: %v", err)
	}

	tests := []struct {
		name    string
		message string
	}{
		{"no type", "add something new"},
		{"wrong format", "FEAT: uppercase type"},
		{"missing colon space", "feat:no space after colon"},
		{"random text", "just some random text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.message, conv)
			if result.Valid {
				t.Errorf("Validate(%q) should be invalid", tt.message)
			}
			found := false
			for _, v := range result.Violations {
				if v.Type == ViolationPattern {
					found = true
					if v.Suggestion == "" {
						t.Error("expected non-empty suggestion")
					}
				}
			}
			if !found {
				t.Error("expected ViolationPattern violation")
			}
		})
	}
}

func TestValidate_ScopeNotCheckedWhenEmpty(t *testing.T) {
	// When convention has no Scopes defined, any scope should be valid.
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^(feat|fix)(\(.+\))?: .+`),
		Types:     []string{"feat", "fix"},
		Scopes:    nil, // no scope restriction
		MaxLength: 100,
	}

	result := Validate("feat(anything): add feature", conv)
	if !result.Valid {
		t.Errorf("expected valid result when scopes not defined, violations: %v", result.Violations)
	}
}

func TestValidate_MultilineOnlyChecksHeader(t *testing.T) {
	conv := &Convention{
		Name:      "test",
		Pattern:   regexp.MustCompile(`^(feat|fix): .+`),
		Types:     []string{"feat", "fix"},
		MaxLength: 50,
	}

	msg := "feat: short header\n\n" + strings.Repeat("a", 200)
	result := Validate(msg, conv)
	if !result.Valid {
		t.Errorf("expected valid result, body length should not matter, violations: %v", result.Violations)
	}
}

func TestExtractType(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{"feat(auth): add JWT", "feat"},
		{"fix: resolve bug", "fix"},
		{"feat!: breaking change", "feat"},
		{"docs(readme): update", "docs"},
		{"no delimiter", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := extractType(tt.header)
			if got != tt.want {
				t.Errorf("extractType(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestExtractScope(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{"feat(auth): add JWT", "auth"},
		{"fix: no scope", ""},
		{"docs(readme): update", "readme"},
		{"feat(multi-word): something", "multi-word"},
		{"no parens at all", ""},
		{"unclosed(paren", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := extractScope(tt.header)
			if got != tt.want {
				t.Errorf("extractScope(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestSuggestFix(t *testing.T) {
	conv := &Convention{
		Name:    "test",
		Pattern: regexp.MustCompile(`^(feat|fix|docs|test|refactor|chore): .+`),
		Types:   []string{"feat", "fix", "docs", "test", "refactor", "chore"},
	}

	tests := []struct {
		header     string
		wantPrefix string
	}{
		{"Fix the login bug", "fix: "},
		{"Add new feature", "feat: "},
		{"Update documentation", "docs: "},
		{"Improve test coverage", "test: "},
		{"Refactor auth module", "refactor: "},
		{"Clean up old code", "refactor: "},
		{"Some random change", "chore: "},
		{"New user endpoint", "feat: "},
		{"Bug in parser", "fix: "},
		{"Update readme file", "docs: "},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := suggestFix(tt.header, conv)
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("suggestFix(%q) = %q, want prefix %q", tt.header, got, tt.wantPrefix)
			}
		})
	}
}
