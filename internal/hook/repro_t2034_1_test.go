//go:build repro_t2034

package hook

// repro_t2034_1_test.go — card t2034 claim 1 (reproduce only, never default-run).
//
// Claim: `cd /lane && cd /primary && git commit` passes the BranchGuard because
// the command holds one git call and the guard adopts only the first `cd`. The
// tests assert the EXPECTED-CORRECT behavior (a protected-branch commit that
// lands in the primary checkout is denied), so a FAIL means REPRODUCED and a
// PASS means REFUTED for that scenario. Fixtures are real temporary git
// repositories (primary on main plus one linked worktree); the hook's decision
// function and the real handler are called in-process.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repro1Guard(t *testing.T, cwd, command string, fx affGuardFixture) (decision, reason string) {
	t.Helper()
	toolInput, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatal(err)
	}
	in := &HookInput{
		SessionID: "sess-t2034-1", HookEventName: "PreToolUse", ToolName: "Bash",
		CWD: cwd, ToolInput: json.RawMessage(toolInput),
	}
	return checkProtectedCommit(in, fx.primary, []string{"main"})
}

func repro1Denies(decision, reason string) bool {
	return decision == DecisionDeny && strings.HasPrefix(reason, branchGuardViolationPrefix+":")
}

// TestRepro1ClaimedChain is the claim as stated: hook cwd is the lane worktree,
// the command cds into the worktree then into the primary and commits.
func TestRepro1ClaimedChain(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	t.Setenv("HOME", t.TempDir())
	fx := newAFFGuardFixture(t)

	cases := []struct {
		name    string
		cwd     string
		command string
		want    bool // want deny
		control bool
	}{
		// the claim: expected-correct is deny (the commit really lands in primary)
		{"claim: cwd=worktree cd wt && cd primary && git commit", fx.release, "cd " + fx.release + " && cd " + fx.primary + " && git commit -m x", true, false},
		{"claim: cwd=primary cd wt && cd primary && git commit", fx.primary, "cd " + fx.release + " && cd " + fx.primary + " && git commit -m x", true, false},
		// positive controls: forms that must already behave as intended
		{"control: cwd=worktree cd primary && git commit (must deny)", fx.release, "cd " + fx.primary + " && git commit -m x", true, true},
		{"control: cwd=worktree cd wt && git commit (must allow)", fx.release, "cd " + fx.release + " && git commit -m x", false, true},
		{"control: cwd=worktree git commit (must allow, lands in worktree)", fx.release, "git commit -m x", false, true},
		{"control: cwd=primary git commit (must deny)", fx.primary, "git commit -m x", true, true},
	}
	swept := 0
	for _, c := range cases {
		decision, reason := repro1Guard(t, c.cwd, c.command, fx)
		got := repro1Denies(decision, reason)
		swept++
		t.Logf("%s -> decision=%q deny-with-sentinel=%t", c.name, decision, got)
		if got != c.want {
			t.Errorf("%s: deny=%t, want %t (decision=%q reason=%q)", c.name, got, c.want, decision, reason)
		}
	}
	if swept != len(cases) || swept == 0 {
		t.Fatalf("swept %d of %d cases", swept, len(cases))
	}
}

// TestRepro1RedObservedOnKnownBadInput proves the assertion can go red: a commit
// that unambiguously lands in the primary (cwd=primary, no cd) is checked with
// the DENY list emptied, which must allow. If this did not allow, the harness
// above could not distinguish allow from deny.
func TestRepro1RedObservedOnKnownBadInput(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	t.Setenv("HOME", t.TempDir())
	fx := newAFFGuardFixture(t)
	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -m x"})
	in := &HookInput{SessionID: "s", HookEventName: "PreToolUse", ToolName: "Bash", CWD: fx.primary, ToolInput: json.RawMessage(toolInput)}
	if d, _ := checkProtectedCommit(in, fx.primary, nil); d != "" {
		t.Fatalf("empty deny list should allow, got %q", d)
	}
	if d, r := checkProtectedCommit(in, fx.primary, []string{"main"}); !repro1Denies(d, r) {
		t.Fatalf("known-bad input (commit in primary on main) was not denied: %q %q", d, r)
	}
}

// TestRepro1ThroughHandler repeats the claimed chain through the whole
// preToolHandler (what the user sees), cwd = lane worktree.
func TestRepro1ThroughHandler(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	t.Setenv("HOME", t.TempDir())
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)
	command := "cd " + fx.release + " && cd " + fx.primary + " && git commit -m x"
	decision, reason := affObserve(t, handler, fx.release, command)
	t.Logf("handler decision=%q reason=%q", decision, reason)
	if decision != DecisionDeny {
		t.Errorf("handler decision = %q, want deny", decision)
	}
}

// TestRepro1ExtraVariants are NOT the stated claim; they probe neighbouring
// spellings of "redirect the target inside a compound command" so the verdict
// is not read wider than measured. Each failing row is a separate observation.
func TestRepro1ExtraVariants(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	t.Setenv("HOME", t.TempDir())
	fx := newAFFGuardFixture(t)

	rel, err := filepath.Rel(fx.release, fx.primary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(fx.release, rel)); err != nil {
		t.Fatalf("relative path does not resolve: %v", err)
	}
	cases := []struct{ name, command string }{
		{"semicolon: cd wt; cd primary; git commit", "cd " + fx.release + "; cd " + fx.primary + "; git commit -m x"},
		{"subshell: cd wt && (cd primary && git commit)", "cd " + fx.release + " && (cd " + fx.primary + " && git commit -m x)"},
		{"relative: cd wt && cd <rel to primary> && git commit", "cd " + fx.release + " && cd " + rel + " && git commit -m x"},
		{"pushd: cd wt && pushd primary && git commit", "cd " + fx.release + " && pushd " + fx.primary + " && git commit -m x"},
		{"newline-separated", "cd " + fx.release + "\ncd " + fx.primary + "\ngit commit -m x"},
	}
	for _, c := range cases {
		decision, reason := repro1Guard(t, fx.release, c.command, fx)
		t.Logf("EXTRA %s -> decision=%q deny-with-sentinel=%t", c.name, decision, repro1Denies(decision, reason))
		if !repro1Denies(decision, reason) {
			t.Errorf("EXTRA %s: not denied (decision=%q), command lands in primary", c.name, decision)
		}
	}
}
