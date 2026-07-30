package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config/toolpolicy"
	"github.com/modu-ai/moai-adk/internal/constitution"
)

// rune_truncate_test.go exercises SPEC-CLIFIX-HYGIENE-001 AC-HYG-001-006:
// every user-facing string truncation site in the CLI tree MUST yield
// utf8.ValidString(output) == true when fed multi-byte (CJK) content longer
// than the truncation limit. The pre-fix code byte-slices (s[:N]) which splits
// multi-byte runes mid-rune and produces invalid UTF-8.
//
// The complete site set derived by grepping the CLI tree for byte-slice
// truncation on user-content strings (ASCII-only slices — hex hashes, API-key
// masking, session IDs, SHAs — are excluded because they cannot carry CJK):
//
//  1. internal/cli/constitution.go renderConstitutionTable clause  (prefix)
//  2. internal/cli/constitution.go renderConstitutionTable file    (suffix, "...")
//  3. internal/cli/tool_policy.go   renderList audit               (prefix)
//  4. internal/cli/tool_policy.go   truncateArg                    (prefix)
//  5. internal/cli/github.go        runParseIssue body             (prefix)
//
// Site 2 (the file-path suffix slice) is the "+1 remaining site" the audit
// undercounted — it is co-located with site 1 and splits CJK file paths at the
// left boundary of the suffix.

// cjkRepeat builds a string of n 3-byte Hangul syllables (가, U+AC00).
// Each rune is exactly 3 UTF-8 bytes, so any byte boundary that is not a
// multiple of 3 lands inside a rune and produces invalid UTF-8.
func cjkRepeat(n int) string {
	return strings.Repeat("가", n)
}

// TestRuneTruncateConstitutionClause reproduces the clause-truncation defect:
// renderConstitutionTable byte-slices the CJK clause at clauseWidth-3 (37 bytes),
// splitting the 13th Hangul rune (byte 37 = 12*3+1). Pre-fix the rendered line
// contains an incomplete UTF-8 sequence → utf8.ValidString == false.
func TestRuneTruncateConstitutionClause(t *testing.T) {
	// 14 Hangul syllables = 42 bytes > clauseWidth(40); truncated at byte 37.
	rules := []constitution.Rule{{
		ID:     "CONST-V3R6-TEST-001",
		Clause: cjkRepeat(14),
	}}
	var buf bytes.Buffer
	renderConstitutionTable(&buf, rules)
	out := buf.String()
	if !utf8.ValidString(out) {
		t.Errorf("renderConstitutionTable produced invalid UTF-8 from CJK clause: %v", identifyInvalidRunes(out))
	}
}

// TestRuneTruncateConstitutionFile reproduces the file-path suffix-slice defect
// (the "+1 remaining site"): a CJK file path longer than fileWidth(50) is
// suffix-sliced at [len-(fileWidth-3):], starting the suffix mid-rune.
func TestRuneTruncateConstitutionFile(t *testing.T) {
	rules := []constitution.Rule{{
		ID:   "CONST-V3R6-TEST-002",
		File: cjkRepeat(20), // 60 bytes > fileWidth(50)
	}}
	var buf bytes.Buffer
	renderConstitutionTable(&buf, rules)
	out := buf.String()
	if !utf8.ValidString(out) {
		t.Errorf("renderConstitutionTable produced invalid UTF-8 from CJK file path: %v", identifyInvalidRunes(out))
	}
}

// TestRuneTruncateToolPolicyAudit reproduces the audit-column truncation in
// renderList: audit string byte-sliced at 77.
func TestRuneTruncateToolPolicyAudit(t *testing.T) {
	entries := []toolpolicy.PolicyEntry{{Tool: "Bash", Audit: cjkRepeat(30)}} // 90 bytes > 80
	var buf bytes.Buffer
	if err := renderList(&buf, entries, "text"); err != nil {
		t.Fatalf("renderList: %v", err)
	}
	out := buf.String()
	if !utf8.ValidString(out) {
		t.Errorf("renderList produced invalid UTF-8 from CJK audit: %v", identifyInvalidRunes(out))
	}
}

// TestRuneTruncateToolPolicyArg reproduces the args_pattern truncation in
// truncateArg: string byte-sliced at 37.
func TestRuneTruncateToolPolicyArg(t *testing.T) {
	got := truncateArg(cjkRepeat(15)) // 45 bytes > 40
	if !utf8.ValidString(got) {
		t.Errorf("truncateArg produced invalid UTF-8 from CJK arg: %q (%v)", got, identifyInvalidRunes(got))
	}
}

// TestRuneTruncateGithubBody reproduces the issue-body truncation defect.
// runParseIssue is not unit-testable without a live GitHub API, so this test
// asserts the source-level defect: github.go MUST NOT byte-slice the body
// (body[:200]). Pre-fix the literal is present → the assertion fails.
func TestRuneTruncateGithubBody(t *testing.T) {
	src, err := os.ReadFile("github.go")
	if err != nil {
		t.Fatalf("read github.go: %v", err)
	}
	if strings.Contains(string(src), "body[:200]") {
		t.Errorf("github.go still byte-slices issue body (body[:200]) — splits CJK runes; route through the rune-safe helper")
	}
}

// identifyInvalidRunes returns the byte offsets where a string is not valid
// UTF-8, to make RED failure output diagnostic.
func identifyInvalidRunes(s string) []int {
	var bad []int
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			bad = append(bad, i)
		}
		i += size
	}
	return bad
}
