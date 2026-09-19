package cli

// integration_known_target_test.go — card t886.
//
// Measured on this tree BEFORE the repair: a github-flow project whose
// git_strategy answers "main" to IntegrationTarget records the CALLER's card
// branch and says nothing on standard error. acquire reads DevelopBranch
// alone, and DevelopBranch is git-flow-gated by contract, so every valid
// non-git-flow flow falls through to the caller silently even though the
// interpretation table (WorkflowIntegrationTarget, REQ-GWS-004) already
// answered the question.
//
// That silence is not an oversight of card t637 — its warning is scoped to
// the git-flow cell on purpose — but of REQ-GWS-008, which left acquire
// unwired and named the follow-up. This is that follow-up, kept warn-only:
// the resolution, the record, stdout and the exit code stay byte-identical,
// exactly as the t637 and t656 warnings did.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// The defect cell: the config names a target, the fallback recorded the
// caller's branch instead, and the two disagree.
func TestIntegrationAcquire_KnownTargetCallerFallbackWarns(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyBody(t, repo, "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n        main_branch: main\n")
	gitRun(t, repo, "checkout", "-q", "-b", "WT-some-card")
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lines := warningLines(stderr)
	if len(lines) != 1 {
		t.Fatalf("want exactly one warning line, got %d: %q (stderr %q)", len(lines), lines, stderr)
	}
	// The warning earns its line only if it names all three: the target the
	// config answered, the branch actually recorded, and the remedy.
	for _, token := range []string{"main", "WT-some-card", "--branch"} {
		if !strings.Contains(lines[0], token) {
			t.Errorf("warning does not name %q: %q", token, lines[0])
		}
	}
	// Warn-only: the record is exactly the pre-change record.
	lock := readLock(t, repo)
	if lock.BranchSource != kanban.BranchSourceCaller || lock.Branch != "WT-some-card" {
		t.Errorf("record = (%q, source %q), want (WT-some-card, source %q) — the warning must not move the record",
			lock.Branch, lock.BranchSource, kanban.BranchSourceCaller)
	}
}

// No-warn negative — the caller is standing in the named target. Nothing is
// wrong there, so a warning would be pure noise. This cell separates "the
// config was never consulted" from "the config's answer and the caller's
// branch coincide", which the pre-existing github-flow negatives cannot tell
// apart: they all run with the caller on main.
func TestIntegrationAcquire_KnownTargetMatchingCallerDoesNotWarn(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyBody(t, repo, "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n        main_branch: main\n")
	chdirRepo(t, repo) // scratchRepo leaves the caller on main

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got := warningLines(stderr); len(got) != 0 {
		t.Errorf("a caller standing in the named target warned: %q", got)
	}
}

// No-warn negative — an explicit --branch decided the target, so no fallback
// happened at all and there is nothing to diagnose.
func TestIntegrationAcquire_KnownTargetExplicitBranchDoesNotWarn(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyBody(t, repo, "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n        main_branch: main\n")
	gitRun(t, repo, "checkout", "-q", "-b", "WT-some-card")
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--branch", "fixture-integration")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got := warningLines(stderr); len(got) != 0 {
		t.Errorf("an explicit --branch warned: %q", got)
	}
}

// No-warn negative — a git-flow project with an empty develop branch answers
// "" to IntegrationTarget, so this warning stays silent and the t637 warning
// remains the single line that cell emits. Two warnings for one fallback
// would be worse than none.
func TestIntegrationAcquire_KnownTargetDoesNotDoubleWarnOnGitFlow(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "git-flow", `""`)
	gitRun(t, repo, "checkout", "-q", "-b", "WT-some-card")
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lines := warningLines(stderr)
	if len(lines) != 1 {
		t.Fatalf("want exactly one warning line, got %d: %q", len(lines), lines)
	}
	if !strings.Contains(lines[0], "develop_branch") {
		t.Errorf("the git-flow cell emitted the wrong warning: %q", lines[0])
	}
}
