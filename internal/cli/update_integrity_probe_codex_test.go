package cli

// update_integrity_probe_codex_test.go — SPEC-UPDATE-MIGRATION-FIX-001 F2
// (REQ-UMF-001): the managed-surface integrity probe must not name the
// Claude-only .claude/settings.json member for a codex-only project. A gpt
// project hides .claude/** by construction (internal/template/harness_fs.go
// hideClaude), so the path is not managed there, and the version-matched probe
// must stay silent about it.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow is the real-path
// witness: an intact codex-only project, updated again at the same version, takes
// the version-match skip path and must print no integrity row.
func TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow(t *testing.T) {
	projectDir := runCodexOnlyProjectThenUpdate(t)
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err == nil {
		t.Fatal("precondition: a codex-only project must carry no .claude/ tree")
	}

	out, err := runUpdateFixtureErr(t, projectDir, true)
	if err != nil {
		t.Fatalf("an intact codex-only update must succeed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Up to date") {
		t.Fatalf("precondition: the version-match skip must run on the codex-only update:\n%s", out)
	}
	if rows := integrityRows(out); len(rows) != 0 {
		t.Fatalf("F2: an intact codex-only project printed integrity rows %q\n%s", rows, out)
	}
}

// TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness pins the same condition at the
// probe level for every harness value: the .claude/settings.json row is printed
// whenever the harness deploys the Claude surfaces (no llm.yaml, claude, both) and
// withheld for a codex-only (gpt) project.
func TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness(t *testing.T) {
	cases := []struct {
		name    string
		harness string // "" writes no llm.yaml, which reads as the claude default
		wantRow bool
	}{
		{"no_llm_yaml_defaults_to_claude", "", true},
		{"claude", "claude", true},
		{"both", "both", true},
		{"gpt_codex_only", "gpt", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			// The manifest is written intact; the assertion below reads only the settings row.
			writeProbeFile(t, root, ".moai/manifest.json", "{}")
			if tc.harness != "" {
				writeProbeFile(t, root, ".moai/config/sections/llm.yaml", "llm:\n  harness: "+tc.harness+"\n")
			}
			var out bytes.Buffer
			runManagedSurfaceIntegrityProbe(&out, root)
			if got := strings.Contains(out.String(), ".claude/settings.json"); got != tc.wantRow {
				t.Fatalf("settings row printed = %v, want %v (harness %q)\n%s", got, tc.wantRow, tc.harness, out.String())
			}
		})
	}
}
