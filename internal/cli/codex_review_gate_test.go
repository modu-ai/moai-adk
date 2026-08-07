package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// SPEC-MOAI-MCP-SERVER-001 M2 (AC-MCP-009 / AC-MCP-010).
//
// These tests pin the `moai hook codex-review-gate` Stop-hook logic BEFORE the
// implementation exists (RED). They assert the mandatory self-gate (no false
// block on a no-edit / loop / disabled turn), the ALLOW/BLOCK contract, the
// opt-in default-off behavior, and fail-open on a missing codex. The handler
// is pure logic (HandleCodexReviewGate); the CLI subcommand wires stdin/stdout.

// gateInput builds a Stop-hook HookInput for the tests.
func gateInput(stopHookActive bool) *hook.HookInput {
	return &hook.HookInput{StopHookActive: stopHookActive}
}

// withChangeDetector swaps the review-gate change detector seam.
func withChangeDetector(t *testing.T, hasChanges bool) {
	t.Helper()
	prev := reviewGateChangeDetector
	reviewGateChangeDetector = func(string) bool { return hasChanges }
	t.Cleanup(func() { reviewGateChangeDetector = prev })
}

// --- AC-MCP-010: opt-in default-off + ALLOW/BLOCK contract ---

// TestReviewGate_DisabledAllows proves the gate is opt-in: when the config
// toggle is off, the hook ALLOWs immediately without invoking codex (the
// distributed default is OFF per C6 / AC-MCP-010).
func TestReviewGate_DisabledAllows(t *testing.T) {
	withChangeDetector(t, true) // even with changes present...
	withCodexLookPath(t, func(string) (string, error) { t.Fatal("codex must not be consulted when gate disabled"); return "", nil })
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, err := HandleCodexReviewGate(gateInput(false), false /* enabled */, "/proj")
	if err != nil {
		t.Fatalf("disabled gate must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("disabled gate must ALLOW (empty output), got %+v", out)
	}
	if runner.calls != 0 {
		t.Errorf("disabled gate must NOT invoke codex; got %d calls", runner.calls)
	}
}

// --- AC-MCP-009: self-gate prevents false blocks ---

// TestReviewGate_LoopPreventionAllows proves stop_hook_active short-circuits to
// ALLOW (mandatory Claude Code loop-prevention protocol; mirrors stopHandler).
func TestReviewGate_LoopPreventionAllows(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { t.Fatal("codex must not be consulted on loop-prevention ALLOW"); return "", nil })
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, _ := HandleCodexReviewGate(gateInput(true) /* stop_hook_active */, true /* enabled */, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("stop_hook_active must ALLOW (loop prevention), got %+v", out)
	}
	if runner.calls != 0 {
		t.Errorf("loop-prevention ALLOW must NOT invoke codex; got %d calls", runner.calls)
	}
}

// TestReviewGate_NoEditTurnAllows proves the self-gate: a turn that produced no
// reviewable code change (clean working tree ⇒ status report / review-result /
// no-op) ALLOWs immediately without invoking codex. This is the AC-MCP-009
// named case — no false block on a non-editing turn.
func TestReviewGate_NoEditTurnAllows(t *testing.T) {
	withChangeDetector(t, false) // clean working tree = no reviewable change
	withCodexLookPath(t, func(string) (string, error) { t.Fatal("codex must not be consulted on a no-edit turn"); return "", nil })
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, _ := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("no-edit turn must ALLOW (self-gate), got %+v", out)
	}
	if runner.calls != 0 {
		t.Errorf("self-gate ALLOW must NOT invoke codex; got %d calls", runner.calls)
	}
}

// TestReviewGate_CodexPassAllows proves an edit turn that codex approves ALLOWs
// (the gate reviews the uncommitted change, codex returns pass → ALLOW).
func TestReviewGate_CodexPassAllows(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdout: rpcResponse(t, ReviewOutput{Verdict: "pass", Summary: "approved"})})

	out, _ := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("codex pass must ALLOW, got %+v", out)
	}
}

// TestReviewGate_CodexFailBlocks proves the BLOCK contract: an edit turn that
// codex rejects BLOCKS the session end ({decision: block, reason: ...}).
func TestReviewGate_CodexFailBlocks(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdout: rpcResponse(t, ReviewOutput{Verdict: "fail", Summary: "found issues", Findings: []Finding{{Severity: "high", Title: "x"}}})})

	out, _ := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if out == nil {
		t.Fatalf("nil output")
	}
	if out.Decision != hook.DecisionBlock {
		t.Errorf("codex fail must BLOCK (decision=%q), got %+v", hook.DecisionBlock, out)
	}
	if out.Reason == "" {
		t.Errorf("BLOCK must carry a reason")
	}
}

// --- fail-open (REQ-MCP-012 preview) ---

// TestReviewGate_FailOpenOnMissingCodex proves a missing codex cannot trap the
// session: the gate ALLOWs rather than blocking on an unavailable reviewer.
func TestReviewGate_FailOpenOnMissingCodex(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, _ := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("missing codex must ALLOW (fail-open), got %+v", out)
	}
}

// TestReviewGate_FailOpenOnCodexError proves a codex error / malformed response
// degrades to ALLOW (the gate never hard-blocks on an inconclusive reviewer).
func TestReviewGate_FailOpenOnCodexError(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{err: errFakeCodexCrash})

	out, _ := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("codex error must ALLOW (fail-open), got %+v", out)
	}
}

// TestReviewGate_InconclusiveAllows proves a codex inconclusive verdict (the
// fail-open ReviewOutput shape) does NOT block.
func TestReviewGate_InconclusiveAllows(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdout: rpcResponse(t, inconclusiveReview("codex timed out"))})

	out, _ := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("inconclusive codex must ALLOW (fail-open), got %+v", out)
	}
}

// TestHasReviewableChanges_RealGitRepo exercises the PRODUCTION git-shellout
// path (not the injected seam) against a real temp git repo, so the 0%-by-seam
// function is honestly covered: modified source ⇒ true, clean tree ⇒ false,
// and a non-git directory ⇒ false (fail-open, no false block outside a repo).
func TestHasReviewableChanges_RealGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	for _, c := range []string{"init -q", "add a.go", "commit -q -m init"} {
		args := append([]string{"-C", dir}, strings.Fields(c)...)
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Skipf("git %s failed: %v", c, err)
		}
	}
	// clean tree right after commit ⇒ nothing reviewable
	if hasReviewableChanges(dir) {
		t.Errorf("clean tree: want false (nothing reviewable)")
	}
	// modify a source file ⇒ reviewable
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package y\n"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if !hasReviewableChanges(dir) {
		t.Errorf("modified source: want true (reviewable)")
	}
	// non-git dir ⇒ fail-open false
	if hasReviewableChanges(t.TempDir()) {
		t.Errorf("non-git dir: want false (fail-open)")
	}
	if hasReviewableChanges("") {
		t.Errorf("empty projectDir: want false")
	}
}

// TestCodexReviewGate_SubcommandRegistered proves the `moai hook
// codex-review-gate` subcommand is wired into the hook command tree (the
// settings.json TEMPLATE registration is M4; M2 ships + tests the subcommand +
// handler via direct invocation).
func TestCodexReviewGate_SubcommandRegistered(t *testing.T) {
	for _, c := range hookCmd.Commands() {
		if c.Name() == "codex-review-gate" {
			return // found
		}
	}
	t.Errorf("subcommand 'codex-review-gate' not registered under `moai hook`")
}

// TestHasReviewableChanges_RuntimePathsExcluded proves the change detector
// ignores runtime-managed paths (.moai/state, .moai/cache, agent-memory, etc.)
// so a session whose only working-tree drift is hook-written state does NOT
// trip the gate (the self-gate stays honest under normal session churn).
func TestHasReviewableChanges_RuntimePathsExcluded(t *testing.T) {
	for _, tc := range []struct {
		name    string
		porcelain string
		want    bool
	}{
		{"empty", "", false},
		{"only runtime", " M .moai/state/active-sessions.json\n?? .moai/cache/x\n M .claude/agent-memory/manager-develop/foo.md\n", false},
		{"source change", " M internal/cli/mcp_codex.go\n", true},
		{"mixed", " M .moai/state/x.json\n M internal/hook/codex_review_gate.go\n", true},
		{"untracked source", "?? internal/cli/new_file.go\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := reviewableFromPorcelain(tc.porcelain); got != tc.want {
				t.Errorf("reviewableFromPorcelain(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
