package cli

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
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

// --- SPEC-REVIEW-GATE-REPAIR-001 M3: turn-end block count and state messages ---
//
// REQ-RG-003 / REQ-RG-004: a fresh fail blocks at turn end, C = 1 block per review
// key in total, a fresh pass clears the key's count, and a new key starts at 0.
// REQ-RG-007: every turn end allowed without a fresh passing receipt names its
// state in systemMessage. Each test reads the outcome through the handler's own
// output (decision and systemMessage); the per-key block count is read back from
// the record path the turn-end handler keeps beside the key's delivery record.

// m3GateFixture builds the turn-end fixture the M3 tests share: a fixture repo as
// the session tree, the primary-scope and version readers pinned, the change
// detector on, codex found, and the --version probe answered. It returns the root.
func m3GateFixture(t *testing.T) string {
	t.Helper()
	root := cacheTestRoot(t)
	withFixedVersionProbe(t)
	withPrimaryScopeReview(t)
	withChangeDetector(t, true)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
	return root
}

// m3ReceiptState is the receipt key the turn-end gate computes for root right now.
func m3ReceiptState(t *testing.T, root string) verify.ReceiptState {
	t.Helper()
	state, err := codexReviewReceiptStateForScope(context.Background(), reviewScopeResolver(root), "/fake/codex")
	if err != nil {
		t.Fatalf("receipt state: %v", err)
	}
	return state
}

// m3ProduceFail runs the receipt producer over the scripted review text, so the
// stored fail carries the same detail file a real review leaves behind.
func m3ProduceFail(t *testing.T, root, reviewText string) {
	t.Helper()
	withCodexSession(t, codexSessionScript(reviewText))
	if _, err := produceCodexReviewReceipt(context.Background(), root); err != nil {
		t.Fatalf("receipt producer: %v", err)
	}
}

// m3RecordReceipt stores a receipt for state through the receipt store, with the
// exit code the producer records for the verdict.
func m3RecordReceipt(t *testing.T, root string, state verify.ReceiptState, verdict string, recordedAt time.Time) {
	t.Helper()
	exit := map[string]int{codexReviewVerdictPass: 0, codexReviewVerdictFail: 1, codexReviewVerdictInconclusive: 2}[verdict]
	if err := verify.RecordReceipt(root, verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		ExitCode:     exit,
		Verdict:      verdict,
		RecordedAt:   recordedAt,
	}); err != nil {
		t.Fatalf("record %s receipt: %v", verdict, err)
	}
}

// m3GateTurnEnd runs the turn-end handler for a turn end that is not a
// loop-prevention continuation (stop_hook_active false).
func m3GateTurnEnd(t *testing.T, root string) *hook.HookOutput {
	t.Helper()
	out, err := HandleCodexReviewGate(gateInput(false), true, root)
	if err != nil {
		t.Fatalf("turn-end gate must not error; got %v", err)
	}
	if out == nil {
		t.Fatal("turn-end gate returned no output")
	}
	return out
}

// m3WantAllowWithState asserts that the turn end is allowed (no block decision)
// and that its systemMessage names state.
func m3WantAllowWithState(t *testing.T, out *hook.HookOutput, state string) {
	t.Helper()
	if out.Decision == hook.DecisionBlock {
		t.Fatalf("want allow naming state %q, got block: %+v", state, out)
	}
	if !strings.Contains(out.SystemMessage, state) {
		t.Errorf("systemMessage must name state %q, got %q", state, out.SystemMessage)
	}
}

// m3BlockRecordPath is the per-key block-count record: the key's delivery-record
// name with the "blocked-" prefix, so the two records name one review key.
func m3BlockRecordPath(root string, state verify.ReceiptState) string {
	delivery := codexReviewDeliveryPath(root, state)
	return filepath.Join(filepath.Dir(delivery), strings.Replace(filepath.Base(delivery), "delivered-", "blocked-", 1))
}

// m3BlockCount reads the key's recorded block count; no record means zero blocks.
func m3BlockCount(t *testing.T, root string, state verify.ReceiptState) int {
	t.Helper()
	raw, err := os.ReadFile(m3BlockRecordPath(root, state))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("read block record: %v", err)
	}
	var rec struct {
		Blocks int `json:"blocks"`
	}
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("block record is not JSON: %v (raw %q)", err, raw)
	}
	return rec.Blocks
}

// TestReviewGate_BlockCountIsOneAfterFirstBlock — AC-RG-004 (REQ-RG-003): a fresh
// fail at turn end blocks with the preserved detail, and the block is counted
// against the key (count 0 -> 1).
func TestReviewGate_BlockCountIsOneAfterFirstBlock(t *testing.T) {
	root := m3GateFixture(t)
	m3ProduceFail(t, root, "Verdict: fail\n\nblocking finding on shared file.")
	state := m3ReceiptState(t, root)

	out := m3GateTurnEnd(t, root)
	if out.Decision != hook.DecisionBlock {
		t.Fatalf("a fresh fail at turn end must BLOCK, got %+v", out)
	}
	if !strings.Contains(out.Reason, "blocking finding on shared file.") {
		t.Errorf("the block must give the fail detail as its reason, got %q", out.Reason)
	}
	if got := m3BlockCount(t, root, state); got != 1 {
		t.Errorf("block count for the key after one block = %d, want 1", got)
	}
}

// TestReviewGate_BlockCapAllowsNextTurnWithMessage — AC-RG-005 (REQ-RG-004,
// REQ-RG-007): once the key has taken its C = 1 block, the next turn end on the
// same fail key is allowed, and its systemMessage names the exhausted count.
func TestReviewGate_BlockCapAllowsNextTurnWithMessage(t *testing.T) {
	root := m3GateFixture(t)
	m3ProduceFail(t, root, "Verdict: fail\n\nblocking finding on shared file.")
	if out := m3GateTurnEnd(t, root); out.Decision != hook.DecisionBlock {
		t.Fatalf("premise: the first fail turn end must BLOCK, got %+v", out)
	}

	m3WantAllowWithState(t, m3GateTurnEnd(t, root), "차단 횟수 소진")
}

// TestReviewGate_AllowedAfterCapStaysAllowedAndCountKept — AC-RG-017 (REQ-RG-004):
// an allowance never gives the block count back. A turn end after the allowance
// is allowed and named again, and the count is still C.
func TestReviewGate_AllowedAfterCapStaysAllowedAndCountKept(t *testing.T) {
	root := m3GateFixture(t)
	m3ProduceFail(t, root, "Verdict: fail\n\nblocking finding on shared file.")
	state := m3ReceiptState(t, root)
	if out := m3GateTurnEnd(t, root); out.Decision != hook.DecisionBlock {
		t.Fatalf("premise: the first fail turn end must BLOCK, got %+v", out)
	}
	m3WantAllowWithState(t, m3GateTurnEnd(t, root), "차단 횟수 소진")

	m3WantAllowWithState(t, m3GateTurnEnd(t, root), "차단 횟수 소진")
	if got := m3BlockCount(t, root, state); got != 1 {
		t.Errorf("block count after the allowance = %d, want 1 (an allowance never gives the count back)", got)
	}
}

// TestReviewGate_FreshPassClearsBlockCount — REQ-RG-004 clearing rule (no AC id):
// a fresh pass recorded for the key removes its block count, so a later fail for
// the same key blocks again. The fresh pass itself is allowed silently (REQ-RG-005).
func TestReviewGate_FreshPassClearsBlockCount(t *testing.T) {
	root := m3GateFixture(t)
	m3ProduceFail(t, root, "Verdict: fail\n\nblocking finding on shared file.")
	state := m3ReceiptState(t, root)
	if out := m3GateTurnEnd(t, root); out.Decision != hook.DecisionBlock {
		t.Fatalf("premise: the first fail turn end must BLOCK, got %+v", out)
	}
	m3WantAllowWithState(t, m3GateTurnEnd(t, root), "차단 횟수 소진")

	m3RecordReceipt(t, root, state, codexReviewVerdictPass, time.Now())
	if out := m3GateTurnEnd(t, root); out.Decision == hook.DecisionBlock || out.SystemMessage != "" {
		t.Fatalf("a fresh pass must allow silently, got %+v", out)
	}

	m3RecordReceipt(t, root, state, codexReviewVerdictFail, time.Now())
	if out := m3GateTurnEnd(t, root); out.Decision != hook.DecisionBlock {
		t.Fatalf("after a fresh pass cleared the count, a new fail for the key must BLOCK again, got %+v", out)
	}
	if got := m3BlockCount(t, root, state); got != 1 {
		t.Errorf("block count after the cleared history = %d, want 1", got)
	}
}

// TestReviewGate_BlockCountIsPerReviewKey — REQ-RG-004 per-key rule (no AC id): the
// block count belongs to the review key. A key that has reached C does not spend
// the count of a later key, which starts at zero (a global counter would fail here).
func TestReviewGate_BlockCountIsPerReviewKey(t *testing.T) {
	root := m3GateFixture(t)
	m3ProduceFail(t, root, "Verdict: fail\n\nblocking finding on shared file.")
	firstKey := m3ReceiptState(t, root)
	if out := m3GateTurnEnd(t, root); out.Decision != hook.DecisionBlock {
		t.Fatalf("premise: the first fail turn end must BLOCK, got %+v", out)
	}
	m3WantAllowWithState(t, m3GateTurnEnd(t, root), "차단 횟수 소진")

	// An edit moves the tree digest, so the review key changes.
	if err := os.WriteFile(filepath.Join(root, "m3_next_key.txt"), []byte("next review key\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	secondKey := m3ReceiptState(t, root)
	if secondKey.TreeDigest == firstKey.TreeDigest {
		t.Fatal("premise: the edit must move the tree digest")
	}
	m3RecordReceipt(t, root, secondKey, codexReviewVerdictFail, time.Now())

	if out := m3GateTurnEnd(t, root); out.Decision != hook.DecisionBlock {
		t.Fatalf("a new review key starts at zero blocks and must BLOCK, got %+v", out)
	}
	if got := m3BlockCount(t, root, secondKey); got != 1 {
		t.Errorf("block count for the new key = %d, want 1", got)
	}
}

// TestReviewGate_NonFreshStateNamedInMessage — AC-RG-018 (REQ-RG-007): with no fresh
// receipt for the current key — absent, stale, or key-mismatch — the turn end is
// allowed, and its systemMessage names that state.
func TestReviewGate_NonFreshStateNamedInMessage(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed func(t *testing.T, root string, state verify.ReceiptState)
	}{
		{name: "absent", seed: func(*testing.T, string, verify.ReceiptState) {}},
		{name: "stale", seed: func(t *testing.T, root string, state verify.ReceiptState) {
			m3RecordReceipt(t, root, state, codexReviewVerdictPass, time.Now().Add(-25*time.Hour))
		}},
		{name: "key-mismatch", seed: func(t *testing.T, root string, state verify.ReceiptState) {
			other := state
			other.ToolVersion = "0.0.1" // a receipt recorded by a different reviewer build
			m3RecordReceipt(t, root, other, codexReviewVerdictFail, time.Now())
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := m3GateFixture(t)
			withKickRecorder(t)
			tc.seed(t, root, m3ReceiptState(t, root))

			m3WantAllowWithState(t, m3GateTurnEnd(t, root), tc.name)
		})
	}
}

// TestReviewGate_NoReviewAllowsCarryStateMessage — AC-RG-019 (REQ-RG-006, REQ-RG-007):
// each turn end allowed without a review keeps its allow decision and names its
// state in systemMessage — the switch off, no reviewable change, the stop-hook
// release, and an absent reviewer.
func TestReviewGate_NoReviewAllowsCarryStateMessage(t *testing.T) {
	t.Run("switch off", func(t *testing.T) {
		root := m3GateFixture(t)
		withCodexLookPath(t, func(string) (string, error) {
			t.Fatal("codex must not be consulted with the switch off")
			return "", nil
		})
		out, err := HandleCodexReviewGate(gateInput(false), false /* enabled */, root)
		if err != nil || out == nil {
			t.Fatalf("switch-off turn end must allow without error; got out=%+v err=%v", out, err)
		}
		m3WantAllowWithState(t, out, "스위치 꺼짐")
	})
	t.Run("no reviewable change", func(t *testing.T) {
		root := m3GateFixture(t)
		withChangeDetector(t, false)
		m3WantAllowWithState(t, m3GateTurnEnd(t, root), "검토할 변경 없음")
	})
	t.Run("stop-hook release", func(t *testing.T) {
		root := m3GateFixture(t)
		out, err := HandleCodexReviewGate(gateInput(true), true, root)
		if err != nil || out == nil {
			t.Fatalf("stop-hook release must allow without error; got out=%+v err=%v", out, err)
		}
		m3WantAllowWithState(t, out, "stop-hook 해제")
	})
	t.Run("reviewer absent", func(t *testing.T) {
		root := m3GateFixture(t)
		withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
		m3WantAllowWithState(t, m3GateTurnEnd(t, root), "reviewer 없음")
	})
}

// TestReviewGate_FreshInconclusiveAllowsWithMessage — AC-RG-025 (REQ-RG-007): a fresh
// inconclusive receipt for the current key allows the turn end, and its
// systemMessage names inconclusive.
func TestReviewGate_FreshInconclusiveAllowsWithMessage(t *testing.T) {
	root := m3GateFixture(t)
	m3RecordReceipt(t, root, m3ReceiptState(t, root), codexReviewVerdictInconclusive, time.Now())

	m3WantAllowWithState(t, m3GateTurnEnd(t, root), "inconclusive")
}
