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
	"strings"
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

// TestReviewGate_StaleKeyKicksBackgroundReview — AC-GBN-001's miss arm, M2
// shape (REQ-GBN-002): an edit that changes the tree digest makes the cached
// receipt stale, and the gate KICKS the background review instead of running
// the RPC in-hook. The session stub stays unused (zero RPCs) — the sharp
// discriminator for "the stale key fell through to the kick, not the cache"
// (a stale receipt reused would allow silently with no kick at all).
func TestReviewGate_StaleKeyKicksBackgroundReview(t *testing.T) {
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	runner := &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}}
	withCodexRunner(t, runner)
	kicked := withKickRecorder(t)

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

	out, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("stale-key gate must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a stale receipt must fall through to the background kick (ALLOW), got %+v", out)
	}
	if len(*kicked) != 1 {
		t.Fatalf("the stale key must kick exactly one background review, got %d", len(*kicked))
	}
	if runner.calls != 0 {
		t.Errorf("the stale-key arm must not run a codex RPC in-hook; got %d calls", runner.calls)
	}
}

// TestProducerFailReceiptPreservesDetailForDelayedBlock — the record-side
// repairs at their M2 home: the receipt PRODUCER (which now runs the review,
// in the background) records the fail exit code (1, not a pass-shaped 0) and
// preserves the summary+findings to the detail file, so the next Stop's cached
// block AND the next-turn entry hook both still say WHAT to fix.
func TestProducerFailReceiptPreservesDetailForDelayedBlock(t *testing.T) {
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	runner := &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}}
	withCodexRunner(t, runner)
	withCodexSession(t, codexSessionScript("Verdict: fail\n\nblocking finding on shared file."))

	// The background producer records the fail receipt with exit 1 and the
	// preserved detail (REQ-GBN-002: the producer now runs the review).
	ctx := context.Background()
	if _, err := produceCodexReviewReceipt(ctx, root); err != nil {
		t.Fatalf("receipt producer: %v", err)
	}
	scope := reviewScopeResolver(root)
	state, err := codexReviewReceiptStateForScope(ctx, scope, "/fake/codex")
	if err != nil {
		t.Fatalf("receipt state: %v", err)
	}
	rec := verify.LoadReceipt(root, state)
	if rec == nil {
		t.Fatal("the producer must record a receipt")
	}
	if rec.Verdict != codexReviewVerdictFail || rec.ExitCode != 1 {
		t.Fatalf("fail receipt must carry verdict=fail exit=1, got verdict=%q exit=%d", rec.Verdict, rec.ExitCode)
	}

	// The next Stop over the unchanged tree: the cache hits and the block
	// message carries the preserved finding text.
	second, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("cache-hit gate must not error; got %v", err)
	}
	if second == nil || second.Decision != hook.DecisionBlock {
		t.Fatalf("a cached FAIL must BLOCK, got %+v", second)
	}
	if !strings.Contains(second.Reason, "blocking finding on shared file.") {
		t.Fatalf("the cached block must carry the preserved finding text, got %q", second.Reason)
	}

	// The NEXT TURN ENTRY: the delayed enforcement point blocks the prompt
	// with the same detail (REQ-GBN-002).
	entry, err := HandleCodexReviewEntry(&hook.HookInput{CWD: root}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if entry == nil || entry.Decision != hook.DecisionBlock {
		t.Fatalf("a fresh FAIL receipt must BLOCK at turn entry, got %+v", entry)
	}
	if !strings.Contains(entry.Reason, "blocking finding on shared file.") {
		t.Fatalf("the delayed block must carry the preserved finding text, got %q", entry.Reason)
	}
}

// TestReviewGate_InconclusiveReceiptIsNotPassEvidence — an inconclusive review
// must allow (fail-open, unchanged) but its receipt must NOT read as local-pass
// evidence: HasLocalPass keys on ExitCode == 0, so the zero value would feed
// the escalation detector's local-pass|ci-failure contradiction limb a pass
// the reviewer never produced. The producer now records it (M2: the producer
// runs the review).
func TestReviewGate_InconclusiveReceiptIsNotPassEvidence(t *testing.T) {
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	runner := &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}}
	withCodexRunner(t, runner)
	// Post-#1718 parser: prose approval carries no pinned Verdict line, so the
	// producer records inconclusive.
	withCodexSession(t, codexSessionScript("Looks good to me — no blocking findings."))

	ctx := context.Background()
	r, err := produceCodexReviewReceipt(ctx, root)
	if err != nil {
		t.Fatalf("receipt producer: %v", err)
	}
	if r.Verdict != codexReviewVerdictInconclusive {
		t.Fatalf("premise: receipt verdict = %q, want inconclusive", r.Verdict)
	}
	if r.ExitCode == 0 {
		t.Fatal("an inconclusive receipt must not carry exit 0 — HasLocalPass would read it as a local pass the reviewer never produced")
	}

	// Fail-open unchanged: the Stop over the recorded inconclusive receipt
	// allows.
	out, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("inconclusive gate must not error; got %v", err)
	}
	// Fail-open unchanged: an empty HookOutput IS the allow.
	if out == nil || out.Decision != "" {
		t.Fatalf("an inconclusive review must allow (fail-open), got %+v", out)
	}
}

// TestProduceCodexReviewReceipt_FailedTurnIsInconclusiveNotPass — the card-t52
// contract at its M2 home: a review turn ending in a NON-completed terminal
// state (failed / interrupted) must record inconclusive (exit 2), never a
// synthesized pass — the real-world shape being the placeholder review codex
// emits when the reviewer itself died ("Reviewer failed to output a
// response.").
func TestProduceCodexReviewReceipt_FailedTurnIsInconclusiveNotPass(t *testing.T) {
	for _, tc := range []struct {
		name       string
		turnStatus string
		turnErrMsg string
	}{
		{"failed with usage-limit error", "failed", "You've hit your usage limit. Try again later."},
		{"interrupted without error object", "interrupted", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := cacheTestRoot(t)
			withFixedVersionProbe(t)
			withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
			withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
			turn := `{"id":"trn","status":` + jsonString(tc.turnStatus)
			if tc.turnErrMsg != "" {
				turn += `,"error":{"message":` + jsonString(tc.turnErrMsg) + `}`
			}
			turn += `}`
			withCodexSession(t, []string{
				`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
				`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
				`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
				// The placeholder codex emits when the reviewer itself died —
				// indistinguishable from a real review by bullet-shape alone.
				`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","item":{"type":"exitedReviewMode","id":"e1","review":"Reviewer failed to output a response."}}}`,
				`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":` + turn + `}}`,
			})

			r, err := produceCodexReviewReceipt(context.Background(), root)
			if err != nil {
				t.Fatalf("a %s turn must not error the producer (fail-open); got %v", tc.turnStatus, err)
			}
			if r.Verdict != codexReviewVerdictInconclusive {
				t.Fatalf("a %s turn must record %q, got %q (a synthesized pass would launder a dead reviewer)", tc.turnStatus, codexReviewVerdictInconclusive, r.Verdict)
			}
			if r.ExitCode != 2 {
				t.Fatalf("an inconclusive receipt must carry exit 2, got %d", r.ExitCode)
			}
		})
	}
}

// TestProduceCodexReviewReceipt_RetriedErrorThenBulletsIsFail — the
// complementary direction: an `error` notification alone (willRetry semantics,
// codex retries internally) whose turn LATER completes with severity-tagged
// finding bullets is a real review — the producer records fail (exit 1) and
// preserves the findings detail.
func TestProduceCodexReviewReceipt_RetriedErrorThenBulletsIsFail(t *testing.T) {
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
	withCodexSession(t, []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
		`{"method":"error","params":{"error":{"message":"transient stream error"},"willRetry":true,"threadId":"tid-fake","turnId":"trn"}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","item":{"type":"exitedReviewMode","id":"e1","review":"- [P1] injection sink found"}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"completed"}}}`,
	})

	r, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("a retried-then-completed turn is a real review; got err=%v", err)
	}
	if r.Verdict != codexReviewVerdictFail || r.ExitCode != 1 {
		t.Fatalf("bullet-carrying findings after a retried error must record fail/1, got %q/%d", r.Verdict, r.ExitCode)
	}
}
