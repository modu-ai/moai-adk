package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// Card t1126 — the remaining `status:` readers must normalize a YAML-quoted
// value (`status: "completed"`) the same way the audit reader does (t1119),
// so every reader agrees on one value for one SPEC. A pair of mismatched
// quotes is not a quoted scalar and is left exactly as written.

var quotedStatusCases = []struct {
	name, raw, want string
}{
	{"bare", `completed`, "completed"},
	{"double-quoted", `"completed"`, "completed"},
	{"single-quoted", `'in-progress'`, "in-progress"},
	{"mismatched quotes are not stripped", `"completed'`, `"completed'`},
	{"lone leading quote is not stripped", `"completed`, `"completed`},
}

// ParseStatus reads the frontmatter `status:` line of spec.md.
func TestParseStatus_QuotedYAMLValue(t *testing.T) {
	t.Parallel()
	for _, tc := range quotedStatusCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			specDir := filepath.Join(t.TempDir(), ".moai", "specs", "SPEC-QUOTED-001")
			if err := os.MkdirAll(specDir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			content := "---\nid: SPEC-QUOTED-001\nstatus: " + tc.raw + "\n---\n# Test\n"
			if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
			got, err := ParseStatus(specDir)
			if err != nil {
				t.Fatalf("ParseStatus: %v", err)
			}
			if got != tc.want {
				t.Errorf("ParseStatus(status: %s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// parseStatusDiffLine reads `status:` from an added or removed diff line.
func TestParseStatusDiffLine_QuotedValue(t *testing.T) {
	t.Parallel()
	for _, tc := range quotedStatusCases {
		for _, sign := range []string{"+", "-"} {
			line := sign + "status: " + tc.raw
			got, ok := parseStatusDiffLine(line, sign)
			if !ok {
				t.Fatalf("parseStatusDiffLine(%q) found no status", line)
			}
			if got != tc.want {
				t.Errorf("parseStatusDiffLine(%q) = %q, want %q", line, got, tc.want)
			}
		}
	}
}

// A quoted value on one side of a transition must not read as a transition
// between two different values when only the quoting changed.
func TestExtractStatusDelta_QuotingOnlyChangeIsSameValue(t *testing.T) {
	t.Parallel()
	diff := "-status: completed\n+status: \"completed\"\n"
	oldS, newS, found := extractStatusDelta(diff)
	if !found {
		t.Fatal("extractStatusDelta found no status delta")
	}
	if oldS != "completed" || newS != "completed" {
		t.Errorf("extractStatusDelta = (%q, %q), want (completed, completed)", oldS, newS)
	}
}
