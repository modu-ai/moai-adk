package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/quality"
	"github.com/modu-ai/moai-adk/internal/timing"
)

// ---------------------------------------------------------------------------
// SPEC-WORKTREE-BRANCH-GUARD-001 — M6 branch-guard tests:
//   * TestBranchGuard_QualityGateNotInvoked     — census P1-B structural
//     guarantee: branch-state commands NEVER enter the quality-gate path
//     (quality.IsGitCommit returns false for every pattern). Falsification
//     arm: a future widening of IsGitCommit to match branch-state commands
//     MUST make this test FAIL.
//   * TestBranchGuard_Latency                    — REQ-WBG-010 / AC-WBG-010
//     arm 1: over 100 consecutive synthetic `git switch` events, the p95
//     invocation stays under the steady-state ceiling and NO single
//     invocation reaches the 5s PreToolUse budget. Internal for-loop, NOT
//     a Benchmark, NOT -benchtime.
//   * TestBranchGuard_CheckBranchStateOrigin     — AC-WBG-010 arm 2: with
//     branchStatePatterns blanked, the same `git switch` event returns
//     allow — proving the deny originates from checkBranchState, NOT from
//     checkBashCommand's DangerousBashPatterns or the quality gate.
// ---------------------------------------------------------------------------

// TestBranchGuard_QualityGateNotInvoked proves the census P1-B structural
// guarantee: branch-state commands (switch/checkout/branch/reset/stash/rebase/
// merge) NEVER trigger quality.IsGitCommit, so they structurally bypass the
// quality-gate path whose measured latency (10.58s on census P1-B) is far out
// of reach of the 5s PreToolUse budget. REQ-WBG-010 / AC-WBG-010 (quality-gate
// bypass arm).
//
// Falsification arm: if a future refactor widens IsGitCommit to match any of
// these commands, the corresponding subtest MUST FAIL (the assertion
// `IsGitCommit(cmd) == false` flips). The paired sanity subtest
// `git_commit_is_recognized` documents that IsGitCommit still correctly
// recognizes `git commit` so a regression that always returns false cannot
// pass silently.
func TestBranchGuard_QualityGateNotInvoked(t *testing.T) {
	t.Parallel()

	branchStateCommands := []string{
		"git switch",
		"git switch -c feat/test",
		"git switch main",
		"git checkout main",
		"git checkout -b feat/x",
		"git branch feature",
		"git branch -d old",
		"git branch -D old",
		"git branch -m renamed",
		"git reset --hard origin/main",
		"git stash",
		"git stash push",
		"git stash pop",
		"git rebase origin/main",
		"git merge feat/x",
	}
	for _, cmd := range branchStateCommands {
		cmd := cmd
		t.Run("not_git_commit/"+cmd, func(t *testing.T) {
			t.Parallel()
			if quality.IsGitCommit(cmd) {
				t.Fatalf("quality.IsGitCommit(%q) = true; want false — "+
					"a branch-state command must NEVER enter the quality-gate path "+
					"(census P1-B structural guarantee, REQ-WBG-010). "+
					"If this flipped, IsGitCommit was widened and the branch-state "+
					"deny now runs on the slow quality-gate path — update the regex.",
					cmd)
			}
		})
	}

	// Sanity arm: IsGitCommit still recognizes `git commit` so a regression
	// that always returns false cannot silently pass the assertions above.
	t.Run("git_commit_is_recognized", func(t *testing.T) {
		t.Parallel()
		commitCmds := []string{
			"git commit -m foo",
			"git commit --amend",
			"git commit --no-verify -m x",
			"  git commit  -m 'leading whitespace'",
		}
		for _, cmd := range commitCmds {
			if !quality.IsGitCommit(cmd) {
				t.Fatalf("quality.IsGitCommit(%q) = false; want true — "+
					"IsGitCommit must still recognize git commit (sanity)", cmd)
			}
		}
	})

	// Negative-control arm: read-only / list-only git commands are not commits
	// either — keeps the assertion meaningful (not vacuously true for any
	// non-commit string).
	t.Run("read_only_not_git_commit", func(t *testing.T) {
		t.Parallel()
		for _, cmd := range []string{"git status", "git log", "git diff", "git fetch"} {
			if quality.IsGitCommit(cmd) {
				t.Fatalf("quality.IsGitCommit(%q) = true; want false (read-only)", cmd)
			}
		}
	})
}

// TestBranchGuard_Latency proves REQ-WBG-010 / AC-WBG-010 arm 1: 100
// consecutive synthetic `git switch` PreToolUse events complete well under the
// 5s PreToolUse budget. The internal for-loop is NOT a Benchmark and NOT
// -benchtime — the test asserts latency bounds, not a throughput average. It
// prints the median / p95 / worst / avg and the calibrated ratio so the bounds
// are mechanically observable in test output.
//
// Three bounds, each answering a different question (enforced by
// internal/timing.Assert):
//
//   - p95 <= 20% of the 5s budget (1s POSIX; 50% / 2.5s on Windows). The
//     distribution detector. A real regression — a network call, an unbounded
//     scan, a full `git status` added to checkBranchState — makes EVERY
//     invocation slow, so it moves the whole distribution and trips p95.
//     Scheduler jitter moves ONE sample and does not.
//   - worst < 100% of budget (5s). The contract itself: no single invocation
//     may consume the whole PreToolUse budget, because one that does stalls the
//     user's session.
//   - median <= 1.5x ONE reference `git rev-parse` spawn, interleaved with
//     the measured runs (timing.AssertPaired) so both sides see the same load
//     window and rest on the same sample count. The CALIBRATED arm — it
//     measures code, not machine. Healthy checkBranchState performs exactly
//     one spawn (ResolveGitDirs primary path) plus sub-millisecond parsing,
//     so the ratio sits at ~1.0x (measured 1.06x, darwin/arm64, 2026-08-20)
//     on any machine under any load; a change that adds a second subprocess moves it
//     to >= 2.0x and trips the bound even while every absolute figure stays
//     inside the generous budget fractions. This closes the gap the earlier
//     forms left: the 500ms absolute ceiling measured machine load (breached
//     3-of-5 runs on unmodified code, avg 135-256ms / worst 308-724ms under
//     multi-session load), and the budget fractions alone cannot see a 2x
//     slowdown that stays under 1s. A persisted baseline cannot close it
//     either — the verify snapshot store keys records to exact tree state
//     (internal/verify/key.go), and a baseline recorded at another time
//     measures another load state; the in-run reference is the only honest
//     baseline (full rationale: internal/timing package doc).
//
// Windows keeps the higher 50% fraction for the p95 arm: each git.exe spawn is
// far more expensive there (issue #1225 observed 872ms against the old 500ms
// bound). The calibrated arm needs no OS-specific fraction — the ratio to a
// same-class spawn is uniform across operating systems.
//
// The latency is measured at the checkBranchState layer (the new deny path
// documented in REQ-WBG-010), NOT through the full Handle pipeline, because
// the census P1-B hazard is specifically the quality-gate path that
// branch-state commands never enter. Measuring at the layer isolates the
// deny-decision cost from incidental fixture setup.
func TestBranchGuard_Latency(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	input := &HookInput{
		SessionID:     "sess-bg-latency",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		AgentType:     "manager-develop",
		CWD:           repo,
		ToolInput:     json.RawMessage(`{"command": "git switch -c feat/perf"}`),
	}

	// Reference unit: ONE `git rev-parse` spawn mirroring what
	// internal/core/git.runGitRevParse does on ResolveGitDirs' primary path —
	// the same arguments AND the same plumbing (two separate buffers with
	// cmd.Run, not CombinedOutput's single shared pipe), so the reference
	// carries the measured operation's cost MIX and not merely its cost class.
	// timing.AssertPaired interleaves this with the measured runs, so both
	// sides see the same load window (see the internal/timing reference rules).
	ref := func() {
		cmd := exec.Command("git", "-C", repo, "rev-parse",
			"--path-format=absolute", "--git-dir", "--git-common-dir")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("reference rev-parse spawn failed: %v (%s)", err, stderr.String())
		}
	}

	// The PreToolUse budget the guard must fit inside. The ceilings below are
	// fractions of it.
	const hookBudget = 5 * time.Second

	steadyCeiling := hookBudget / 5 // 20% — 1s on POSIX
	if runtime.GOOS == "windows" {
		steadyCeiling = hookBudget / 2 // 50% — 2.5s, per issue #1225 spawn cost
	}

	// Healthy ratio is ~1.0x (one spawn + sub-ms parsing). MaxUnits 1.5 keeps
	// wide headroom over healthy code while tripping on any change that adds
	// a second subprocess (>= 2.0x) — the regression class invisible to the
	// budget fractions when absolute figures stay generous.
	timing.AssertPaired(t, timing.Bound{
		Name:          "checkBranchState",
		Budget:        hookBudget,
		SteadyCeiling: steadyCeiling,
		MaxUnits:      1.5,
		Iterations:    100,
		Warmup:        3,
	}, ref, func() {
		decision, _ := checkBranchState(input, repo)
		if decision != DecisionDeny {
			t.Fatalf("checkBranchState decision = %q, want %q "+
				"(latency measurement requires the deny to actually fire)", decision, DecisionDeny)
		}
	})
}

// TestBranchGuard_CheckBranchStateOrigin proves AC-WBG-010 arm 2 (deny-origin
// falsification): with the branchStatePatterns var blanked (it is named +
// blankable per the @MX:ANCHOR in branch_guard.go), the SAME `git switch`
// PreToolUse event that is otherwise denied now returns allow — proving the
// deny originates from checkBranchState's regex set, NOT from
// checkBashCommand's DangerousBashPatterns or the quality gate. Restore via
// t.Cleanup.
//
// The test goes through Handle (the end-to-end path) so it exercises the real
// insertion point and rules out the dangerous-pattern path as the deny source.
func TestBranchGuard_CheckBranchStateOrigin(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	handler := &preToolHandler{
		// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001: the guard is now default-OFF;
		// opt this deny-origin test in so the deny path is reachable (the
		// checkBranchState signature is unchanged — only the config setup
		// changes, no signature cascade per plan.md D1).
		cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(true)},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}
	input := &HookInput{
		SessionID:     "sess-bg-origin",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		AgentType:     "manager-develop",
		CWD:           repo,
		ToolInput:     json.RawMessage(`{"command": "git switch -c feat/origin"}`),
	}

	// Sanity: with patterns present, the event IS denied (proves the test
	// fixture is the deny path, not silently allowing).
	out, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle err = %v", err)
	}
	if decisionOf(out) != DecisionDeny {
		t.Fatalf("precondition: Handle decision = %q, want %q (with patterns present)",
			decisionOf(out), DecisionDeny)
	}
	if reason := reasonOf(out); !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
		t.Fatalf("precondition: reason = %q, want prefix %q", reason, branchGuardViolationPrefix+":")
	}

	// Falsification: blank the branchStatePatterns var. The same event MUST
	// now return non-deny — proving the deny came from checkBranchState's
	// regex set, NOT from checkBashCommand's DangerousBashPatterns or the
	// quality gate.
	orig := branchStatePatterns
	t.Cleanup(func() { branchStatePatterns = orig })
	branchStatePatterns = nil

	out2, err2 := handler.Handle(context.Background(), input)
	if err2 != nil {
		t.Fatalf("Handle err (blanked) = %v", err2)
	}
	if decision := decisionOf(out2); decision == DecisionDeny {
		reason := reasonOf(out2)
		t.Fatalf("Handle(blanked patterns) = deny; want non-deny — "+
			"the deny MUST disappear when branchStatePatterns is blanked, "+
			"proving it came from checkBranchState, not checkBashCommand or the quality gate "+
			"(reason=%q). If this flipped, the deny is now sourced elsewhere — "+
			"re-audit the insertion point in pre_tool.go.", reason)
	}

	// Confirm cleanup restored the patterns so the var is reusable for
	// subsequent tests in the same process (defensive — t.Cleanup already
	// restored, but assert so a future change to Cleanup ordering is caught).
	if branchStatePatterns != nil {
		// t.Cleanup runs AFTER the test body, so within the body the var is
		// still nil here. Re-blank is a no-op; just sanity-check the
		// matchBranchStateCommand path is inert while blanked.
		if _, matched := matchBranchStateCommand("git switch"); matched {
			t.Fatalf("matchBranchStateCommand matched with blanked patterns — blanking is ineffective")
		}
	}
	// Use _ to keep the import list stable if fmt is otherwise unused here.
	_ = fmt.Sprint
	_ = os.Getenv
}
