package cli

// codex_review_delay_test.go — SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 /
// AC-GBN-004: the delayed block. A cache-miss Stop kicks the receipt
// producer's review in the background and allows the turn; the NEXT-turn-entry
// handler (HandleCodexReviewEntry) enforces the verdict that review recorded.
// Every failure path here ends in an allow (fail-open).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// withKickRecorder swaps the background-kick seam for a recorder and returns
// the kicked dirs. Every test that can reach a cache miss MUST inject this —
// the production default execs the real moai binary.
func withKickRecorder(t *testing.T) *[]string {
	t.Helper()
	var kicked []string
	prev := codexReviewBackgroundKick
	codexReviewBackgroundKick = func(dir string) error {
		kicked = append(kicked, dir)
		return nil
	}
	t.Cleanup(func() { codexReviewBackgroundKick = prev })
	return &kicked
}

// withKickFailure swaps the kick seam for one that always fails (AC-GBN-004's
// red input).
func withKickFailure(t *testing.T) {
	t.Helper()
	prev := codexReviewBackgroundKick
	codexReviewBackgroundKick = func(string) error { return errors.New("spawn failed") }
	t.Cleanup(func() { codexReviewBackgroundKick = prev })
}

// entrySeams installs the stubs every entry-handler test needs: a reviewable
// change, a fake reviewer path, and the fixed version probe the receipt key
// carries. It returns nothing — tests add exactly the receipts they need.
func entrySeams(t *testing.T) {
	t.Helper()
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withFixedVersionProbe(t)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
}

// recordEntryReceipt records a receipt for the fixture's current tree state —
// what the background producer would have written for the previous turn.
func recordEntryReceipt(t *testing.T, root, verdict string, exit int, digestSuffix string) {
	t.Helper()
	state, err := codexReviewReceiptStateForScope(context.Background(), reviewScopeResolver(root), "/fake/codex")
	if err != nil {
		t.Fatalf("receipt state: %v", err)
	}
	if digestSuffix != "" {
		state.TreeDigest += digestSuffix
	}
	if err := verify.RecordReceipt(root, verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		Verdict:      verdict,
		ExitCode:     exit,
		RecordedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("record the %s receipt: %v", verdict, err)
	}
}

// --- the Stop side: cache miss ⇒ kick + ALLOW ------------------------------

// TestReviewGate_CacheMissKicksBackgroundReview — REQ-GBN-002: a cache-miss
// Stop allows the turn and kicks exactly one background review for the
// session's scope dir, with ZERO codex RPCs in-hook.
func TestReviewGate_CacheMissKicksBackgroundReview(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	kicked := withKickRecorder(t)
	runner := &fakeCodexRunner{}
	withCodexRunner(t, runner)

	out, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("kick-path gate must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a cache-miss Stop must ALLOW (the review runs in the background), got %+v", out)
	}
	// M2 anchoring: the kick carries the git toplevel (macOS temp dirs sit
	// behind a /var → /private/var symlink — resolve before comparing).
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval fixture root: %v", err)
	}
	if len(*kicked) != 1 || (*kicked)[0] != resolvedRoot {
		t.Fatalf("the gate must kick exactly one background review for the git root, got %v", *kicked)
	}
	if runner.calls != 0 {
		t.Errorf("the Stop must not run a codex RPC in-hook; got %d calls", runner.calls)
	}
}

// TestReviewGate_BackgroundKickFailureAllows — AC-GBN-004: a failed kick is
// fail-open; the allow stands and the next Stop re-kicks.
func TestReviewGate_BackgroundKickFailureAllows(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	withKickFailure(t)

	out, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("a failed kick must not error the gate; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a failed kick must ALLOW (fail-open), got %+v", out)
	}
}

// TestReviewGate_MissingCodexDoesNotKick — a missing reviewer allows AND kicks
// nothing: no reviewer means no receipt can ever arrive, so a kick would be a
// process spawn that only ever fails.
func TestReviewGate_MissingCodexDoesNotKick(t *testing.T) {
	root := cacheTestRoot(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
	kicked := withKickRecorder(t)

	out, _ := HandleCodexReviewGate(gateInput(false), true, root)
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("missing codex must ALLOW (fail-open), got %+v", out)
	}
	if len(*kicked) != 0 {
		t.Errorf("a missing reviewer must not kick a background review, got %v", *kicked)
	}
}

// --- the in-flight kick marker (the turn-end gate's P2) ---------------------

// TestReviewEntry_SubdirSessionReadsRootReceipt — the state root anchors on
// the git toplevel: a session sitting in a SUBDIRECTORY must consult (and the
// producer must write) the receipt at <gitroot>/.moai/state, or the receipt
// write itself would move the tree key and the delayed verdict would never
// match (the turn-end gate's P1, overlay-reproduced pre-fix).
func TestReviewEntry_SubdirSessionReadsRootReceipt(t *testing.T) {
	root := cacheTestRoot(t)
	sub := filepath.Join(root, "internal", "cli")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	entrySeams(t)
	// The receipt is recorded the way the producer would: bound to the
	// git-root scope state, stored under the git root.
	rootScope := reviewScopeResolver(root)
	if rootScope.Class != reviewScopeTree {
		t.Fatalf("premise: the fixture must resolve tree-scope, got %+v", rootScope)
	}
	state, err := codexReviewReceiptStateForScope(context.Background(), rootScope, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	if err := verify.RecordReceipt(root, verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		Verdict:      codexReviewVerdictFail,
		ExitCode:     1,
		RecordedAt:   time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	out, err := HandleCodexReviewEntry(&hook.HookInput{CWD: sub}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a subdir session must read the root-anchored FAIL receipt and BLOCK, got %+v", out)
	}
}

// TestKickInFlight_ExclusiveAcquisition — the marker takes exactly ONE
// mutation (the exclusive create): the first caller owns the kick, every
// later caller on the same marker — fresh OR stale — reads in-flight, and the
// stale one ages out by mtime within the review budget (the next Stop then
// re-kicks through the expired marker). No takeover path exists to race.
func TestKickInFlight_ExclusiveAcquisition(t *testing.T) {
	root := cacheTestRoot(t)
	state, err := codexReviewReceiptStateForScope(context.Background(), reviewScopeResolver(root), "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}

	inFlight, markerPath := kickInFlight(root, state)
	if inFlight || markerPath == "" {
		t.Fatalf("the first acquisition must own the kick, got inFlight=%v path=%q", inFlight, markerPath)
	}
	inFlight, _ = kickInFlight(root, state)
	if !inFlight {
		t.Fatal("a concurrent second acquisition on the fresh marker must read in-flight")
	}

	// A stale marker ALSO reads in-flight — it ages out by mtime; there is no
	// takeover to race on.
	stale := time.Now().Add(-2 * config.DefaultCodexReviewGateTimeout)
	if err := os.Chtimes(markerPath, stale, stale); err != nil {
		t.Fatal(err)
	}
	inFlight, _ = kickInFlight(root, state)
	if !inFlight {
		t.Fatal("a stale marker must read in-flight (it ages out by mtime; no takeover exists)")
	}
}

// TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot — the producer's
// storage root follows the anchored git toplevel: a caller that named a
// subdirectory gets its receipt stored where the entry hook reads it (the
// turn-end gate's P2 overlay repro).
func TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot(t *testing.T) {
	root := cacheTestRoot(t)
	sub := filepath.Join(root, "internal", "cli")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	entrySeams(t)
	withCodexSession(t, codexSessionScript("Verdict: fail\n\nblocking finding on shared file."))

	if _, err := produceCodexReviewReceipt(context.Background(), sub); err != nil {
		t.Fatalf("receipt producer: %v", err)
	}
	// The root-anchored state — the same binding the entry hook consults —
	// must hold the recorded receipt.
	rootScope := reviewScopeResolver(root)
	state, err := codexReviewReceiptStateForScope(context.Background(), rootScope, "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	rec := verify.LoadReceipt(root, state)
	if rec == nil {
		t.Fatal("the receipt must be stored at the git root's state, where the entry hook reads it")
	}
	if rec.Verdict != codexReviewVerdictFail {
		t.Fatalf("receipt verdict = %q, want fail", rec.Verdict)
	}
	// And the delayed enforcement blocks the subdir session end to end.
	entry, err := HandleCodexReviewEntry(&hook.HookInput{CWD: sub}, true, root)
	if err != nil {
		t.Fatalf("entry error: %v", err)
	}
	if entry == nil || entry.Decision != hook.DecisionBlock {
		t.Fatalf("the subdir session's entry must BLOCK on the root-stored fail, got %+v", entry)
	}
}

// TestReviewGate_InFlightKickNotRepeated — consecutive Stops over the SAME
// tree state must not start a second background review while the first is in
// flight: the per-key marker deduplicates them.
func TestReviewGate_InFlightKickNotRepeated(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	kicked := withKickRecorder(t)

	for i := 1; i <= 2; i++ {
		out, err := HandleCodexReviewGate(gateInput(false), true, root)
		if err != nil {
			t.Fatalf("Stop %d must not error; got %v", i, err)
		}
		if out == nil || out.Decision == hook.DecisionBlock {
			t.Fatalf("Stop %d must ALLOW, got %+v", i, out)
		}
	}
	if len(*kicked) != 1 {
		t.Fatalf("the in-flight kick must not be repeated over the unchanged tree; got %v", *kicked)
	}
}

// TestReviewGate_StaleMarkerAgesOut — a marker older than the review budget
// reads as a dead review that ages out by mtime: the next Stop does NOT
// re-kick while it sits there (no takeover, no race), and the marker's
// expiry restores the kick within one budget.
func TestReviewGate_StaleMarkerAgesOut(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	kicked := withKickRecorder(t)

	if _, err := HandleCodexReviewGate(gateInput(false), true, root); err != nil {
		t.Fatalf("first Stop error: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".moai", "state", "verify", "codex-review", "*.kick"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("premise: exactly one kick marker expected, got %v (%v)", matches, err)
	}
	stale := time.Now().Add(-2 * config.DefaultCodexReviewGateTimeout)
	if err := os.Chtimes(matches[0], stale, stale); err != nil {
		t.Fatal(err)
	}

	if _, err := HandleCodexReviewGate(gateInput(false), true, root); err != nil {
		t.Fatalf("second Stop error: %v", err)
	}
	if len(*kicked) != 1 {
		t.Fatalf("a stale marker must NOT re-kick (it ages out; no takeover exists); got %v", *kicked)
	}
}

// TestReviewGate_FailedKickIsRetryable — a failed start removes its marker, so
// the next Stop over the same tree retries the kick (only finished or failed
// work is retried; in-flight work never is).
func TestReviewGate_FailedKickIsRetryable(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)

	var kicks []string
	shouldFail := true
	prev := codexReviewBackgroundKick
	codexReviewBackgroundKick = func(dir string) error {
		if shouldFail {
			return errors.New("spawn failed")
		}
		kicks = append(kicks, dir)
		return nil
	}
	t.Cleanup(func() { codexReviewBackgroundKick = prev })

	if out, err := HandleCodexReviewGate(gateInput(false), true, root); err != nil || out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a failed kick must ALLOW (fail-open); out=%+v err=%v", out, err)
	}
	shouldFail = false
	if _, err := HandleCodexReviewGate(gateInput(false), true, root); err != nil {
		t.Fatalf("second Stop error: %v", err)
	}
	if len(kicks) != 1 {
		t.Fatalf("the failed kick must be retried on the next Stop; got %v", kicks)
	}
}

// --- the entry side: the delayed enforcement point -------------------------

// TestReviewEntry_FreshFailBlocks — REQ-GBN-002's enforcement point: the
// fresh FAIL receipt the previous turn's background review recorded blocks
// the prompt, carrying the preserved finding detail.
func TestReviewEntry_FreshFailBlocks(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	recordEntryReceipt(t, root, codexReviewVerdictFail, 1, "")
	codexReviewPreserveFindings(root, mustEntryState(t, root), ReviewOutput{
		Verdict: "fail",
		Summary: "reviewer failed the turn",
		Findings: []Finding{{
			Severity: "P1", Title: "delayed finding",
			Body: "blocking finding on shared file.", File: "main.go", Line: 3,
		}},
	})

	out, err := HandleCodexReviewEntry(&hook.HookInput{CWD: root}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a fresh FAIL receipt must BLOCK at turn entry, got %+v", out)
	}
	if !strings.Contains(out.Reason, "delayed verdict") {
		t.Fatalf("the delayed block must name the delayed verdict, got %q", out.Reason)
	}
	if !strings.Contains(out.Reason, "blocking finding on shared file.") {
		t.Fatalf("the delayed block must carry the preserved finding text, got %q", out.Reason)
	}
}

// TestReviewEntry_FreshPassAllows — a fresh pass receipt allows; the entry
// hook never blocks on a clean background verdict.
func TestReviewEntry_FreshPassAllows(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	recordEntryReceipt(t, root, codexReviewVerdictPass, 0, "")

	out, err := HandleCodexReviewEntry(&hook.HookInput{CWD: root}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a fresh pass receipt must ALLOW, got %+v", out)
	}
}

// TestReviewEntry_NoReceiptAllows — the background review is still running (or
// its kick failed): no receipt yet ⇒ the prompt proceeds (fail-open; the next
// Stop re-kicks if the tree moved).
func TestReviewEntry_NoReceiptAllows(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)

	out, err := HandleCodexReviewEntry(&hook.HookInput{CWD: root}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a missing receipt must ALLOW (background still in flight), got %+v", out)
	}
}

// TestReviewEntry_StaleReceiptAllows — the tree moved between the kick and
// the entry (another actor edited): the receipt no longer binds the current
// key, so it cannot block. The next Stop kicks for the new key.
func TestReviewEntry_StaleReceiptAllows(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	recordEntryReceipt(t, root, codexReviewVerdictFail, 1, "-stale")

	out, err := HandleCodexReviewEntry(&hook.HookInput{CWD: root}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("a stale receipt must ALLOW (the key moved), got %+v", out)
	}
}

// TestReviewEntry_NoEditTurnAllows — the self-gate: nothing reviewable means
// nothing was kicked for this tree, and the entry hook must not even resolve
// the reviewer.
func TestReviewEntry_NoEditTurnAllows(t *testing.T) {
	withPrimaryScopeReview(t)
	withChangeDetector(t, false)
	withCodexLookPath(t, func(string) (string, error) { t.Fatal("codex must not be consulted on a no-edit turn"); return "", nil })

	out, _ := HandleCodexReviewEntry(&hook.HookInput{CWD: "/proj"}, true, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("no-edit turn must ALLOW (self-gate), got %+v", out)
	}
}

// TestReviewEntry_DisabledAllows — the entry hook shares the Stop gate's
// opt-in flag: disabled ⇒ allow with zero work.
func TestReviewEntry_DisabledAllows(t *testing.T) {
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) {
		t.Fatal("codex must not be consulted when gate disabled")
		return "", nil
	})

	out, _ := HandleCodexReviewEntry(&hook.HookInput{CWD: "/proj"}, false, "/proj")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Fatalf("disabled entry hook must ALLOW (empty output), got %+v", out)
	}
}

// TestCodexReviewEntry_SubcommandRegistered proves the `moai hook
// codex-review-entry` subcommand is wired into the hook command tree (the
// settings registration is pinned by the template-side tests).
func TestCodexReviewEntry_SubcommandRegistered(t *testing.T) {
	for _, c := range hookCmd.Commands() {
		if c.Name() == "codex-review-entry" {
			return // found
		}
	}
	t.Errorf("subcommand 'codex-review-entry' not registered under `moai hook`")
}

// mustEntryState is a test convenience: the receipt state for the fixture's
// current tree state, failing the test when the key cannot be measured.
func mustEntryState(t *testing.T, root string) verify.ReceiptState {
	t.Helper()
	state, err := codexReviewReceiptStateForScope(context.Background(), reviewScopeResolver(root), "/fake/codex")
	if err != nil {
		t.Fatalf("receipt state: %v", err)
	}
	return state
}
