package cli

// codex_review_cache_test.go — SPEC-GATE-BOTTLENECK-001 REQ-GBN-001: the
// turn-end gate reuses a FRESH receipt for the resolved scope's tree key
// instead of re-running the codex RPC, and falls through to a live review
// when the key is stale. The cache is an accelerator, never an authority:
// every failure path here must end in the pre-cache behavior (live review).

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// cacheTestRoot builds a real git repo (the cache consult runs git for the
// tree key, so a fake path cannot reach it) and returns it as the session
// tree for the gate input. The tree-key digest counts untracked files, so
// the fixture mirrors the real repo's ignore of `.moai/state/` — otherwise
// the receipt write itself would move the key between the record and the
// consult (the real repo's tracked .gitignore excludes it).
func cacheTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".moai/state/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitIn(t, root, "add", ".gitignore")
	runGitIn(t, root, "commit", "-q", "-m", "ignore the runtime state")
	return root
}

// withPrimaryScopeReview stubs the primary/tree-scope policy readers so a
// standalone fixture repo is not skipped before the cache consult: the real
// readers read the repo's config, which a t.TempDir fixture does not carry.
func withPrimaryScopeReview(t *testing.T) {
	t.Helper()
	prevPrimary := primaryScopeReader
	prevTree := reviewGateTreeScopeReader
	primaryScopeReader = func(string) string { return config.CodexReviewGatePrimaryScopeReview }
	reviewGateTreeScopeReader = func(string) string { return config.CodexReviewGatePrimaryScopeReview }
	t.Cleanup(func() {
		primaryScopeReader = prevPrimary
		reviewGateTreeScopeReader = prevTree
	})
}

// withFixedVersionProbe pins the receipt's tool_version probe: the cache key
// carries the reviewer version, so both the recorded receipt and the gate's
// own state computation must see the same non-empty answer.
func withFixedVersionProbe(t *testing.T) {
	t.Helper()
	prev := codexVersionProbe
	codexVersionProbe = func(context.Context, string) (string, error) { return "9.9.9", nil }
	t.Cleanup(func() { codexVersionProbe = prev })
}

// TestReviewGate_CacheHitReusesFailVerdict — AC-GBN-001/002: a stored FAIL
// receipt for the current tree key blocks the gate with ZERO codex RPCs; a
// second invocation over the unchanged tree hits the same cached verdict.
func TestReviewGate_CacheHitReusesFailVerdict(t *testing.T) {
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	// The --version probe must answer, or the receipt's tool_version is
	// unbound on both sides and CheckReceipt reads the pair as a mismatch.
	runner := &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}}
	withCodexRunner(t, runner)

	ctx := context.Background()
	scope := reviewScopeResolver(root)
	state, err := codexReviewReceiptStateForScope(ctx, scope, "/fake/codex")
	if err != nil {
		t.Fatalf("receipt state: %v", err)
	}
	// Prerequisite: the probe answered — an unbound tool_version would make
	// this test vacuous (an unbound field is never evidence).
	if state.ToolVersion == "" {
		t.Fatal("premise: receipt state carries no tool_version — the version probe must answer through the fake runner")
	}
	if err := verify.RecordReceipt(root, verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		Verdict:      codexReviewVerdictFail,
		RecordedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("record the cached fail: %v", err)
	}

	skipsBefore := codexReviewCacheSkips.Load()
	out, err := HandleCodexReviewGate(gateInput(false), true /* enabled */, root)
	if err != nil {
		t.Fatalf("cache-hit gate must not error; got %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a cached FAIL verdict must BLOCK, got %+v", out)
	}
	if runner.calls != 0 {
		t.Errorf("cache hit must NOT invoke codex; got %d calls", runner.calls)
	}
	if codexReviewCacheSkips.Load() != skipsBefore+1 {
		t.Errorf("cache hit did not record the skip: %d → %d", skipsBefore, codexReviewCacheSkips.Load())
	}
}

// TestReviewGate_StaleKeyRunsLiveReview — AC-GBN-001's miss arm: an edit that
// changes the tree digest makes the cached receipt stale, and the gate runs a
// live review (here: one that fails, blocking again and refreshing the
// receipt).
func TestReviewGate_StaleKeyRunsLiveReview(t *testing.T) {
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	runner := &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}}
	withCodexRunner(t, runner)

	ctx := context.Background()
	scope := reviewScopeResolver(root)
	state, err := codexReviewReceiptStateForScope(ctx, scope, "/fake/codex")
	if err != nil {
		t.Fatalf("receipt state: %v", err)
	}
	if err := verify.RecordReceipt(root, verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest + "-stale",
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		Verdict:      codexReviewVerdictPass,
		RecordedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("record the stale pass: %v", err)
	}
	// The live review is scripted to FAIL: a stale PASS receipt reused would
	// ALLOW (the cache bug this guards against), while a genuine live run of
	// the failing fixture BLOCKs — the sharp discriminator for "the stale key
	// actually fell through".
	withCodexSession(t, codexSessionScript("Verdict: fail\n\nblocking finding."))

	out, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("stale-key gate must not error; got %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a stale receipt must fall through to the live review (fixture verdict fail ⇒ BLOCK), got %+v", out)
	}
}
