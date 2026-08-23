package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Card t178 — two independent defects surfaced by the SPEC-MCP-WORKTREE-ROOT-001
// M5 post-repair run, both of which let a backend report a verdict that does not
// describe the code under review.
//
//  (a) codex's ADVERSARIAL path returns free prose whose own first line reads
//      "Verdict: fail — merge blocked."  synthesizeReviewOutput read only the
//      review-mode "- [P1]" bullet form, so that fail was synthesized as PASS
//      and a required gate inverted silently.
//
//  (b) the GLM backend was never sent any code. It was asked to "Review the
//      proposed change" with no diff, no tree, and no file — so a confident
//      `fail` it returned was about nothing at all.

// ── (a) codex adversarial verdict ──

// TestSynthesizeReviewOutput_AdversarialVerdictLine pins the adversarial-mode
// prose shape. The text is the shape codex actually returned in the M5 run: a
// leading "Verdict: fail" line and findings written as prose, carrying none of
// the "- [P1]" severity bullets the review-mode heuristic keys on.
func TestSynthesizeReviewOutput_AdversarialVerdictLine(t *testing.T) {
	const adversarial = `Verdict: fail — merge blocked.

1. audit_multi persists its convergence state to the wrong tree
   internal/cli/mcp_convergence.go:520 — persistConvergenceResult never
   consults cfg.ProjectRoot.

2. The shared parameter description promised the wrong fallback.`

	if got := synthesizeReviewOutput(adversarial).Verdict; got != "fail" {
		t.Errorf("adversarial prose opening with %q synthesized Verdict = %q, want \"fail\"\nreview text:\n%s",
			"Verdict: fail", got, adversarial)
	}
}

// TestSynthesizeReviewOutput_VerdictLineDirections covers both directions plus
// the conflict case. A stated pass must NOT override finding bullets: the two
// signals disagreeing resolves toward fail, never toward the clean verdict.
func TestSynthesizeReviewOutput_VerdictLineDirections(t *testing.T) {
	cases := []struct {
		name, text, want string
	}{
		{"stated fail, no bullets", "Verdict: fail — merge blocked.", "fail"},
		{"stated FAIL uppercase", "VERDICT: FAIL\nreasons follow", "fail"},
		{"stated fail, markdown bold", "**Verdict:** fail\n\nfindings follow", "fail"},
		{"stated pass, no bullets", "Verdict: pass — no blocking findings.", "pass"},
		{"stated pass but bullets present", "Verdict: pass\n- [P1] secret at vuln.go:7", "fail"},
		{"no verdict line, bullets", "- [P2] minor style issue", "fail"},
		{"no verdict line, clean", "The change introduces no blocking issues.", "pass"},
		{"the word verdict in prose only", "I could not reach a verdict on the caching layer.", "pass"},
	}
	for _, c := range cases {
		if got := synthesizeReviewOutput(c.text).Verdict; got != c.want {
			t.Errorf("%s: Verdict = %q, want %q (text: %q)", c.name, got, c.want, c.text)
		}
	}
}

// ── (b) GLM reviews nothing ──

// TestGLMAudit_RequestCarriesTheDiff asserts the z.ai request body actually
// contains the change under review. Before the repair the body carried only the
// generic instruction, so the model had no material and answered from
// imagination.
func TestGLMAudit_RequestCarriesTheDiff(t *testing.T) {
	root := gitTreeWithChange(t)
	doer := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "ok"})}
	withGLMSeams(t, "test-key", doer)

	out := performGLMAudit(context.Background(), codexTargetUncommitted, "", root)
	if out.Verdict == VerdictInconclusive {
		t.Fatalf("expected a real verdict with a diff available; got inconclusive: %s", out.Summary)
	}
	if !strings.Contains(doer.gotBody, "canary_symbol_t178") {
		t.Errorf("z.ai request body does not carry the change under review.\nbody:\n%s", doer.gotBody)
	}
}

// TestGLMAudit_NoMaterialIsInconclusiveNotAVerdict pins the fail-open direction
// for the case the repair cannot fix: when no diff can be produced, GLM has
// nothing to review and MUST NOT be asked for a verdict. A confident answer
// about nothing is worse than an inconclusive one.
func TestGLMAudit_NoMaterialIsInconclusiveNotAVerdict(t *testing.T) {
	root := t.TempDir() // not a git tree ⇒ no diff obtainable
	doer := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "fail", Summary: "hallucinated"})}
	withGLMSeams(t, "test-key", doer)

	out := performGLMAudit(context.Background(), codexTargetUncommitted, "", root)
	if out.Verdict != VerdictInconclusive {
		t.Errorf("no review material available: Verdict = %q, want %q", out.Verdict, VerdictInconclusive)
	}
	if doer.gotBody != "" {
		t.Errorf("z.ai must not be called with no material to review; body sent:\n%s", doer.gotBody)
	}
}

// gitTreeWithChange builds a throwaway git repository holding one committed file
// and one uncommitted edit carrying a canary symbol, and returns its path.
func gitTreeWithChange(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	f := filepath.Join(root, "a.go")
	if err := os.WriteFile(f, []byte("package a\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run("add", "a.go")
	run("commit", "-q", "-m", "base")
	if err := os.WriteFile(f, []byte("package a\n\nfunc canary_symbol_t178() {}\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return root
}
