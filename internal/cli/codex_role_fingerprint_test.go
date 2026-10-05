// Package cli — codex_role_fingerprint_test.go
//
// AC-RLP-001 (parse) and AC-RLP-002 (expectation-table eligibility) for
// SPEC-ROLE-LOAD-PREDICATE-001.
package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"
)

// TestCodexRoleBodyFingerprintParse — AC-RLP-001 (REQ-RLP-003).
func TestCodexRoleBodyFingerprintParse(t *testing.T) {
	const bodyText = "line one\nline two\n"

	fixtures := map[string]string{
		"leading_newline":    "model_reasoning_effort = \"high\"\ndeveloper_instructions = '''\n" + bodyText + "'''\n",
		"no_leading_newline": "model_reasoning_effort = \"high\"\ndeveloper_instructions = '''" + bodyText + "'''\n",
		"inner_quotes":       "model_reasoning_effort = \"high\"\ndeveloper_instructions = '''line one\ncan''t stop\nline two\n'''\n",
		"missing_key":        "model_reasoning_effort = \"high\"\nsandbox_mode = \"read-only\"\n",
	}

	regexExtract := regexp.MustCompile(`(?s)developer_instructions = '''(.*)'''`)

	var (
		leadingDelta     int
		leadingDeltaByte byte
		deltaMeasured    bool
	)

	t.Run("leading_newline", func(t *testing.T) {
		got, err := codexRoleBodyExtractTOML(fixtures["leading_newline"])
		if err != nil {
			t.Fatalf("extract: %v", err)
		}
		if got != bodyText {
			t.Fatalf("got %q, want %q", got, bodyText)
		}

		// Naive regex-style extraction: captures everything between the
		// opening and closing ''' delimiters literally, WITHOUT the TOML
		// leading-newline-trim rule. Constructed here only for comparison —
		// the implementation never uses regexp (REQ-RLP-003).
		m := regexExtract.FindStringSubmatch(fixtures["leading_newline"])
		if m == nil {
			t.Fatal("regex extraction found no match")
		}
		regexVal := m[1]
		if len(regexVal) != len(got)+1 {
			t.Fatalf("delta is not exactly one byte: regex len=%d toml len=%d", len(regexVal), len(got))
		}
		if regexVal[1:] != got {
			t.Fatalf("regex extraction differs from TOML-parsed value beyond the leading byte")
		}
		leadingDelta = len(regexVal) - len(got)
		leadingDeltaByte = regexVal[0]
		deltaMeasured = true
	})

	t.Run("no_leading_newline", func(t *testing.T) {
		got, err := codexRoleBodyExtractTOML(fixtures["no_leading_newline"])
		if err != nil {
			t.Fatalf("extract: %v", err)
		}
		if got != bodyText {
			t.Fatalf("got %q, want %q", got, bodyText)
		}
	})

	t.Run("inner_quotes", func(t *testing.T) {
		want := "line one\ncan''t stop\nline two\n"
		got, err := codexRoleBodyExtractTOML(fixtures["inner_quotes"])
		if err != nil {
			t.Fatalf("extract: %v", err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("missing_key", func(t *testing.T) {
		_, err := codexRoleBodyExtractTOML(fixtures["missing_key"])
		if err == nil {
			t.Fatal("expected extraction failure for a TOML source with no developer_instructions key, got success (silent empty string)")
		}
	})

	t.Run("regex_mutant_rejected", func(t *testing.T) {
		m := regexExtract.FindStringSubmatch(fixtures["leading_newline"])
		if m == nil {
			t.Fatal("regex extraction found no match")
		}
		regexVal := m[1]
		tomlVal, err := codexRoleBodyExtractTOML(fixtures["leading_newline"])
		if err != nil {
			t.Fatalf("extract: %v", err)
		}
		if codexRoleBodySHA256Hex(regexVal) == codexRoleBodySHA256Hex(tomlVal) {
			t.Fatal("a regex-extraction mutant must NOT fingerprint-match the TOML-spec extraction")
		}
	})

	if !deltaMeasured {
		t.Fatal("leading-newline delta was not measured")
	}
	if leadingDelta != 1 || leadingDeltaByte != '\n' {
		t.Fatalf("unexpected delta: %d bytes, byte=0x%02x", leadingDelta, leadingDeltaByte)
	}
	fmt.Printf("PARSE_LEADING_DELTA %d 0x%02x\n", leadingDelta, leadingDeltaByte)
}

// TestCodexRoleBodyExpectationTable — AC-RLP-002 (REQ-RLP-005, REQ-RLP-007).
func TestCodexRoleBodyExpectationTable(t *testing.T) {
	reasons := map[string]string{}

	t.Run("real_tree_wellformed", func(t *testing.T) {
		root := repoRoot(t)
		dir := filepath.Join(root, "internal/template/templates/.codex/agents/moai")
		table, reason, err := codexBuildRoleExpectationTable(dir)
		if err != nil {
			t.Fatalf("build table: %v", err)
		}
		if reason != "" {
			t.Fatalf("real tree reported ineligible: %s", reason)
		}
		if len(table.ByRole) == 0 {
			t.Fatal("real tree produced an empty table")
		}
		if len(table.ByRole) != len(table.ByHash) {
			t.Fatalf("role count %d != distinct hash count %d", len(table.ByRole), len(table.ByHash))
		}
	})

	t.Run("colliding_bodies", func(t *testing.T) {
		dir := t.TempDir()
		body := "model_reasoning_effort = \"high\"\ndeveloper_instructions = '''\nsame body for both roles\n'''\n"
		writeFile(t, filepath.Join(dir, "role-a.toml"), body)
		writeFile(t, filepath.Join(dir, "role-b.toml"), body)
		_, reason, err := codexBuildRoleExpectationTable(dir)
		if err != nil {
			t.Fatalf("build table: %v", err)
		}
		if reason != codexRoleTableReasonCollision {
			t.Fatalf("got reason %q, want %q", reason, codexRoleTableReasonCollision)
		}
		reasons["colliding_bodies"] = reason
	})

	t.Run("empty_body", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "role-empty.toml"),
			"model_reasoning_effort = \"high\"\ndeveloper_instructions = ''''''\n")
		_, reason, err := codexBuildRoleExpectationTable(dir)
		if err != nil {
			t.Fatalf("build table: %v", err)
		}
		if reason != codexRoleTableReasonEmptyBody {
			t.Fatalf("got reason %q, want %q", reason, codexRoleTableReasonEmptyBody)
		}
		reasons["empty_body"] = reason
	})

	t.Run("empty_dir", func(t *testing.T) {
		dir := t.TempDir()
		_, reason, err := codexBuildRoleExpectationTable(dir)
		if err != nil {
			t.Fatalf("build table: %v", err)
		}
		if reason != codexRoleTableReasonEmptyDir {
			t.Fatalf("got reason %q, want %q", reason, codexRoleTableReasonEmptyDir)
		}
		reasons["empty_dir"] = reason
	})

	distinct := map[string]bool{}
	for _, r := range reasons {
		distinct[r] = true
	}
	if len(distinct) != 3 {
		t.Fatalf("expected 3 distinct reason codes, got %d: %v", len(distinct), reasons)
	}
	fmt.Printf("TABLE_DISTINCT_REASONS %d\n", len(distinct))
}
