package cli

import (
	"context"
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
	withCodexLookPath(t, func(string) (string, error) {
		t.Fatal("codex must not be consulted when gate disabled")
		return "", nil
	})
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
	withCodexLookPath(t, func(string) (string, error) {
		t.Fatal("codex must not be consulted on loop-prevention ALLOW")
		return "", nil
	})
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

// SPEC-GATE-BOTTLENECK-001 (REQ-GBN-002): the in-hook LIVE review tests
// (CodexPassAllows / CodexFailBlocks / FailOpenOnCodexError /
// InconclusiveAllows / FailedTurnSurfacesErrorNotPass /
// ErrorNotificationWithRetryingTurnStillReviews) moved with the review
// itself: the verdict-producing contracts now live at the receipt PRODUCER
// (codex_review_cache_test.go), the kick contracts at the delayed-block tests
// (codex_review_delay_test.go), and the enforcement contracts at the entry
// handler. The tests below pin what still belongs to the Stop gate: the
// self-gates, the fail-open on a missing reviewer, and the cached-verdict
// reuse.

// --- fail-open (REQ-MCP-012 preview) ---

// TestReviewGate_FailOpenOnMissingCodex proves a missing codex cannot trap the
// session: the gate ALLOWs rather than blocking on an unavailable reviewer
// (and kicks nothing — the kick contract is pinned at
// TestReviewGate_MissingCodexDoesNotKick).
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

// TestRunCodexReviewRPC_FailedTurnIsInconclusiveNotPass pins the same contract
// one level down: the RPC layer's ReviewOutput for a failed turn must be the
// inconclusive fail-open verdict — never "pass", which would read downstream
// as "reviewed and found nothing wrong".
func TestRunCodexReviewRPC_FailedTurnIsInconclusiveNotPass(t *testing.T) {
	withCodexSession(t, []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","item":{"type":"exitedReviewMode","id":"e1","review":"Reviewer failed to output a response."}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"failed","error":{"message":"usage limit exceeded"}}}}`,
	})

	out, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart, map[string]any{
		"target": codexTargetUncommitted,
	})
	if err == nil {
		t.Fatalf("failed turn must return an error; got nil with verdict=%q", out.Verdict)
	}
	if out.Verdict != VerdictInconclusive {
		t.Errorf("failed turn must synthesize %q, got %q", VerdictInconclusive, out.Verdict)
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

// TestCodexReviewGateNonGitDirNoPrimarySkip pins REQ-CGSC-006: a session tree
// whose primary-versus-linked status cannot be established (a non-git
// directory) does not take the primary-checkout skip — the gate keeps the
// pre-SPEC behavior of that state (the self-gate runs, the reviewer path is
// consulted) and no policy skip row is logged. M2 (REQ-GBN-002): the consult
// is the background kick now — the gate still fails open to ALLOW.
func TestCodexReviewGateNonGitDirNoPrimarySkip(t *testing.T) {
	withChangeDetector(t, true)
	p := newOwnershipProbe(t)
	skips := captureTreeScopeSkips(t)

	out := gatePath(t, "/proj", "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("an undecidable tree must keep the pre-SPEC consult behavior (fail-open ALLOW; the kick replaces the in-hook review), got %+v", out)
	}
	if p.detects != 1 {
		t.Errorf("the primary skip must not fire for an undecidable tree: the self-gate must run exactly once, got %d detector calls", p.detects)
	}
	if len(*skips) != 0 {
		t.Errorf("no policy skip row expected for an undecidable tree, got %d", len(*skips))
	}
}

// TestHasReviewableChanges_RuntimePathsExcluded proves the change detector
// ignores runtime-managed paths (.moai/state, .moai/cache, agent-memory, etc.)
// so a session whose only working-tree drift is hook-written state does NOT
// trip the gate (the self-gate stays honest under normal session churn).
func TestHasReviewableChanges_RuntimePathsExcluded(t *testing.T) {
	for _, tc := range []struct {
		name      string
		porcelain string
		want      bool
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
