// Package measure tests — pure parser leaf package (SPEC-HARNESS-REGRESSION-GATE-001 M1).
// These tests assert the three exported parsers (ParseGoTestJSON / ParseCoverageFile /
// CountNonEmptyLines) produce correct results, and that the transitive helpers
// (parseIntFromString / mustParseInt / mustParseIntErr) behave as before extraction.
package measure

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseGoTestJSON covers go test -json pass/fail counting (REQ-RG-002).
func TestParseGoTestJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantPassed int
		wantFailed int
	}{
		{
			name:       "empty input yields zero",
			input:      "",
			wantPassed: 0,
			wantFailed: 0,
		},
		{
			name: "two pass one fail (top-level tests only)",
			input: `{"Action":"pass","Test":"TestA"}
{"Action":"pass","Test":"TestB"}
{"Action":"fail","Test":"TestC"}`,
			wantPassed: 2,
			wantFailed: 1,
		},
		{
			name: "package-level events (empty Test) are ignored",
			input: `{"Action":"pass","Package":"pkg"}
{"Action":"pass","Test":"TestA"}
{"Action":"fail","Package":"pkg2"}`,
			wantPassed: 1,
			wantFailed: 0,
		},
		{
			name: "malformed JSON lines are skipped",
			input: `not-json
{"Action":"pass","Test":"TestA"}
{bad`,
			wantPassed: 1,
			wantFailed: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			passed, failed := ParseGoTestJSON([]byte(tt.input))
			if passed != tt.wantPassed {
				t.Errorf("ParseGoTestJSON() passed = %d, want %d", passed, tt.wantPassed)
			}
			if failed != tt.wantFailed {
				t.Errorf("ParseGoTestJSON() failed = %d, want %d", failed, tt.wantFailed)
			}
		})
	}
}

// TestParseCoverageFile covers coverage-profile percentage computation (REQ-RG-002).
func TestParseCoverageFile(t *testing.T) {
	t.Parallel()

	t.Run("absent file returns zero", func(t *testing.T) {
		t.Parallel()
		got := ParseCoverageFile(filepath.Join(t.TempDir(), "nonexistent.out"))
		if got != 0 {
			t.Errorf("ParseCoverageFile(absent) = %f, want 0", got)
		}
	})

	t.Run("half covered yields 50 percent", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "cover.out")
		// Two blocks, 10 statements each. First covered (count>0), second not.
		content := "mode: set\n" +
			"pkg/file.go:1.1,2.2 10 1\n" +
			"pkg/file.go:3.1,4.2 10 0\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write coverage fixture: %v", err)
		}
		got := ParseCoverageFile(path)
		if got != 50.0 {
			t.Errorf("ParseCoverageFile(half) = %f, want 50.0", got)
		}
	})

	t.Run("zero statements yields zero", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "cover.out")
		if err := os.WriteFile(path, []byte("mode: set\n"), 0o644); err != nil {
			t.Fatalf("write coverage fixture: %v", err)
		}
		got := ParseCoverageFile(path)
		if got != 0 {
			t.Errorf("ParseCoverageFile(empty) = %f, want 0", got)
		}
	})
}

// TestCountNonEmptyLines covers go vet lint-line counting (REQ-RG-002).
func TestCountNonEmptyLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"only whitespace lines", "  \n\t\n   \n", 0},
		{"three content lines", "a\nb\nc\n", 3},
		{"mixed content and blank", "a\n\n  \nb\n", 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CountNonEmptyLines([]byte(tt.input))
			if got != tt.want {
				t.Errorf("CountNonEmptyLines(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseIntFromString covers the transitive helper moved from internal/loop
// (parseIntFromString / mustParseInt / mustParseIntErr) — kept unexported here.
func TestParseIntFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"42", 42, false},
		{"100", 100, false},
		{"", 0, false}, // empty string: no digit chars, loop skips, returns 0
		{"abc", 0, true},
		{"1a", 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := parseIntFromString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseIntFromString(%q) = %d, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseIntFromString(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseIntFromString(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestMustParseInt covers mustParseInt for valid + error inputs.
func TestMustParseInt(t *testing.T) {
	t.Parallel()

	if got := mustParseInt("7"); got != 7 {
		t.Errorf("mustParseInt(\"7\") = %d, want 7", got)
	}
	if got := mustParseInt("abc"); got != 0 {
		t.Errorf("mustParseInt(\"abc\") = %d, want 0 (error case)", got)
	}
}
