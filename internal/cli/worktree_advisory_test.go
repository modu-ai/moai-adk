package cli

// worktree_advisory_test.go — SPEC-WORKTREE-KEY-WIRING-001 M3 (REQ-WKW-011 /
// AC-WKW-011): the auto_create=true advisory wording claims nothing no code
// performs, keeps the AC-WBG-009 observability regex match, and points at the
// key that actually enables automatic isolation. The false wording and the
// config-failure degradation are unchanged.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// advisoryWBG009Regex is the AC-WBG-009 observability pattern, verbatim.
var advisoryWBG009Regex = regexp.MustCompile(`worktree.*isolation|use a worktree|moai \(cc\|cg\) -w|claude --worktree`)

// advisoryFalseWording is the exact recommendation wording the false branch
// (and every degradation path) must keep emitting byte-identically. The flag
// list carries `moai glm -w` (not the retired `cg` token) per the t649
// gateway rename absorbed into this branch — the byte-identity contract is
// unchanged, only the retired token was replaced.
const advisoryFalseWording = "Tip: this checkout is shared across concurrent sessions; " +
	"for branch-changing work (switch/reset/rebase), use a worktree for isolation — " +
	"`moai cc -w` / `moai glm -w`, or `claude --worktree`. " +
	"See .claude/rules/moai/workflow/main-checkout-branch-guard.md.\n"

// advisoryFixture writes a minimal project workflow.yaml carrying the given
// auto_create value and returns the project root.
func advisoryFixture(t *testing.T, autoCreate bool) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("workflow:\n    worktree:\n        auto_create: %v\n", autoCreate)
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestWorktreeAdvisoryTruthful is AC-WKW-011.
func TestWorktreeAdvisoryTruthful(t *testing.T) {
	t.Run("auto_create_true_no_false_claim", func(t *testing.T) {
		out := &bytes.Buffer{}
		emitWorktreeAdvisory(out, advisoryFixture(t, true))
		s := out.String()

		if !advisoryWBG009Regex.MatchString(s) {
			t.Errorf("true wording must keep the AC-WBG-009 observability regex match, got: %q", s)
		}
		for _, neg := range []string{"is auto-creating", "will be created"} {
			if strings.Contains(s, neg) {
				t.Errorf("true wording must claim no automatic creation (%q found): %q", neg, s)
			}
		}
		if !strings.Contains(s, "workflow.session_worktree.enabled") {
			t.Errorf("true wording must point at the key that enables real automatic isolation: %q", s)
		}
		if !strings.Contains(s, "main-checkout-branch-guard.md") {
			t.Errorf("true wording must keep the rule pointer: %q", s)
		}
	})

	t.Run("auto_create_false_wording_unchanged", func(t *testing.T) {
		out := &bytes.Buffer{}
		emitWorktreeAdvisory(out, advisoryFixture(t, false))
		if got := out.String(); got != advisoryFalseWording {
			t.Errorf("false wording must stay byte-identical:\n got %q\nwant %q", got, advisoryFalseWording)
		}
	})

	t.Run("config_failure_degrades_to_false_wording", func(t *testing.T) {
		out := &bytes.Buffer{}
		// A root with no .moai directory at all — the loader fails, the
		// advisory degrades to the default (false) wording.
		emitWorktreeAdvisory(out, filepath.Join(t.TempDir(), "nonexistent"))
		if got := out.String(); got != advisoryFalseWording {
			t.Errorf("degradation wording must match the false wording byte-identically:\n got %q\nwant %q", got, advisoryFalseWording)
		}
	})
}
