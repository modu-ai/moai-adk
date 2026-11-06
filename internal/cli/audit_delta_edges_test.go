package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M3 — failure modes of the delta codex leg and
// of the default delta proposal beyond the AC rows (REQ-AEC-001, REQ-AEC-003;
// card t2017). Every test works in a t.TempDir() tree.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// A diff the server cannot collect, or one that is empty, makes the leg
// inconclusive: it never reviews something the delta did not contain, and never
// falls back to asking codex to compute the diff itself.
func TestCodexDeltaLegPrompt_UncollectableOrEmptyDiff_IsAnError(t *testing.T) {
	d := newDeltaRepo(t)
	if _, err := codexDeltaLegPrompt(t.TempDir(), "", d.shaB); err == nil {
		t.Errorf("a tree with no git history must be an error")
	}
	head := reviewTargetGitOut(t, d.root, "rev-parse", "HEAD")
	if _, err := codexDeltaLegPrompt(d.root, "", head); err == nil {
		t.Errorf("an empty delta (base == HEAD) must be an error")
	}
}

// Through the handler: a delta whose diff is empty sends nothing to codex and the
// leg reports inconclusive, naming why.
func TestAEC001a_EmptyDelta_CodexLegIsInconclusive_AndSendsNothing(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, adversarialScript())
	prev := backendCall
	backendCall = defaultBackendCaller
	t.Cleanup(func() { backendCall = prev })

	head := reviewTargetGitOut(t, d.root, "rev-parse", "HEAD")
	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"target": codexTargetBaseBranch, "project_root": d.root, "delta_base": head,
		"gates": map[string]any{"claude": "off", "codex": "required", "glm": "off"},
	}}})
	if err != nil || res.IsError {
		t.Fatalf("handleAuditMulti: err=%v isError=%v %v", err, res.IsError, res.Content)
	}
	var raw string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			raw += tc.Text
		}
	}
	var result ConvergenceResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(sess.sent) != 0 {
		t.Errorf("an empty delta must send nothing to codex; sent=%v", sess.sent)
	}
	for _, v := range result.PerBackendVerdicts {
		if v.Backend == BackendCodex && (v.Verdict != VerdictInconclusive || !strings.Contains(v.Summary, "delta diff not collected")) {
			t.Errorf("codex leg = %q (%s), want inconclusive naming the uncollected diff", v.Verdict, v.Summary)
		}
	}
}

// A call that names no project_root resolves the tree the way the GLM leg does, so
// the codex leg still carries the delta diff the server collected.
func TestAEC001a_NoProjectRoot_CodexLegResolvesTheTreeAndStillCarriesTheDiff(t *testing.T) {
	d := newDeltaRepo(t)
	t.Setenv("CLAUDE_PROJECT_DIR", d.root)
	sess := withCodexSession(t, adversarialScript())
	prev := backendCall
	backendCall = defaultBackendCaller
	t.Cleanup(func() { backendCall = prev })

	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"target": codexTargetBaseBranch, "delta_base": d.shaB,
		"gates": map[string]any{"claude": "off", "codex": "required", "glm": "off"},
	}}})
	if err != nil || res.IsError {
		t.Fatalf("handleAuditMulti: err=%v isError=%v %v", err, res.IsError, res.Content)
	}
	if prompt := sentTurnPrompt(t, sess); !strings.Contains(prompt, "canary_c_delta") {
		t.Errorf("the turn does not carry the delta content; prompt:\n%s", prompt)
	}
}

// A delta larger than the review size limit is cut, and the turn says so.
func TestCodexDeltaLegPrompt_OversizedDelta_IsCutAndSaysSo(t *testing.T) {
	d := newDeltaRepo(t)
	big := strings.Repeat("line of delta content\n", reviewDiffMaxBytes/20)
	if err := os.WriteFile(filepath.Join(d.root, "big.txt"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	reviewTargetGit(t, d.root, "add", "big.txt")
	reviewTargetGit(t, d.root, "commit", "-m", "big delta")

	prompt, err := codexDeltaLegPrompt(d.root, "focus text", d.shaC)
	if err != nil {
		t.Fatalf("codexDeltaLegPrompt: %v", err)
	}
	if !strings.Contains(prompt, "cut at the review size limit") {
		t.Errorf("a cut diff must be declared in the turn")
	}
	if !strings.Contains(prompt, "focus text") {
		t.Errorf("the caller's focus must stay in the turn")
	}
	if len(prompt) > reviewDiffMaxBytes+10_000 {
		t.Errorf("the turn is not bounded: %d bytes", len(prompt))
	}
}

// The default proposal proposes nothing, and says why, for the inputs below.
func TestDefaultDeltaProposal_UnusableInputs_ProposeNothing(t *testing.T) {
	t.Run("an invalid card id", func(t *testing.T) {
		sha, note := defaultDeltaProposal(t.TempDir(), "../escape")
		if sha != "" || !strings.Contains(note, dpPrefix+" "+dpNone) || !strings.Contains(note, "not a valid card id") {
			t.Errorf("got sha %q note %q", sha, note)
		}
	})
	t.Run("a tree whose evidence home cannot be resolved", func(t *testing.T) {
		sha, note := defaultDeltaProposal(t.TempDir(), "tnone")
		if sha != "" || !strings.Contains(note, dpPrefix+" "+dpNone) {
			t.Errorf("got sha %q note %q", sha, note)
		}
	})
	t.Run("a newest verdict file that cannot be read in full", func(t *testing.T) {
		w := newAECWorld(t)
		w.verdictRaw("sync-audit-iter1.md", failBody(w.head(), 0)+strings.Repeat("x", maxVerdictFileBytes+8))
		sha, note := defaultDeltaProposal(w.tree, aecCard)
		if sha != "" || !strings.Contains(note, "cannot be read in full") {
			t.Errorf("got sha %q note %q", sha, note)
		}
	})
	t.Run("conflicting verdict lines", func(t *testing.T) {
		w := newAECWorld(t)
		w.verdictRaw("sync-audit-iter1.md", "verdict: FAIL\nverdict: PASS\naudited_sha: "+w.head()+"\n")
		sha, note := defaultDeltaProposal(w.tree, aecCard)
		if sha != "" || !strings.Contains(note, "conflicting verdict lines") {
			t.Errorf("got sha %q note %q", sha, note)
		}
	})
	t.Run("an audited_sha that is no commit of the tree", func(t *testing.T) {
		w := newAECWorld(t)
		w.verdictRaw("sync-audit-iter1.md", failBody("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", 0))
		sha, note := defaultDeltaProposal(w.tree, aecCard)
		if sha != "" || !strings.Contains(note, "not a commit of the audited tree") || !strings.Contains(note, dpNotAncestor) {
			t.Errorf("got sha %q note %q", sha, note)
		}
	})
}
