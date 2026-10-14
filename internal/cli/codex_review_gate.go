// Package cli — `moai hook codex-review-gate` Stop-hook logic
// (SPEC-MOAI-MCP-SERVER-001 M2, REQ-MCP-008 / AC-MCP-009 / AC-MCP-010).
//
// codex_review_gate.go holds the PURE gate logic. The CLI subcommand wiring
// (stdin read → project-root resolve → config-gate read → this handler →
// stdout) lives in hook.go runCodexReviewGate. Keeping the logic pure + in the
// cli package lets it reuse the codex backend (mcp_codex.go, same package) and
// the readCodexReviewGateEnabled config reader without an import cycle
// (internal/cli → internal/hook already exists for the HookInput/HookOutput
// types). It returns the standard ALLOW/BLOCK HookOutput and NEVER invokes
// AskUserQuestion (subagent boundary — the orchestrator translates a BLOCK).
//
// SPEC-GATE-BOTTLENECK-001 rebalances this gate onto the receipt store: the
// live verdict is reused from the tree-keyed cache (REQ-GBN-001), a cache
// miss kicks the review in the background and allows the turn (REQ-GBN-002),
// and the verdict is enforced at the NEXT turn entry by HandleCodexReviewEntry
// (codex_review_delay.go). The gate itself never runs an RPC anymore.
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
// @MX:SPEC: SPEC-GATE-BOTTLENECK-001
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"

	"github.com/spf13/cobra"
)

// reviewGateRuntimePrefixes are the working-tree paths the change detector
// ignores: they are hook/session-written state that drifts on every turn, so
// counting them as "reviewable" would make the self-gate fire on every Stop
// and defeat AC-MCP-009 (no false block on a non-editing turn).
var reviewGateRuntimePrefixes = []string{
	".moai/state/",
	".moai/cache/",
	".moai/reports/",
	".moai/logs/",
	".moai/harness/",
	".claude/agent-memory/",
}

// reviewGateTreeConfigExactPaths and reviewGateTreeConfigDirPrefixes are the
// runtime-managed CONFIGURATION surfaces the TREE-scope self-gate additionally
// ignores (SPEC-CODEX-GATE-SCOPING-001 REQ-CGSC-007): the session's local
// Claude settings file and the MoAI managed config tree are known local state,
// not reviewable work. Deliberately SEPARATE from reviewGateRuntimePrefixes
// AND from the shared parser below (card-review repair R1): the exclusion is
// consulted only by the tree-scope self-gate (treeConfigOnlyFromPorcelain) —
// the card scope's path filter shares only the runtime list, a card's own
// commits under .moai/config/ keep counting as card work (REQ-CGSC-005 /
// AC-CGSC-009), and the multi-review gates keep the shared baseline detector.
// The settings file matches EXACTLY (card-review repair R4): a sibling name
// that merely extends it (.claude/settings.json.template) is a source-shaped
// path and stays reviewable; only the directory matches by prefix.
var (
	reviewGateTreeConfigExactPaths  = []string{".claude/settings.json"}
	reviewGateTreeConfigDirPrefixes = []string{".moai/config/"}
)

// reviewGateChangeDetector is the injectable "is there reviewable uncommitted
// work?" seam. The production default runs `git status --porcelain` and filters
// runtime-managed paths; tests swap it to drive the self-gate deterministically
// (true = reviewable change present, false = clean / no-edit turn).
var reviewGateChangeDetector = hasReviewableChanges

// codexReviewBlockLimit is C, the number of turn-end blocks one review key may
// take in total (SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-003 and REQ-RG-004; operator
// verdict Q2 = 1). An allowance never gives a block back, and a fresh pass for
// the key clears its count.
// C = 1, operator ruling Q2 (REQ-RG-004).
//
// @MX:NOTE: [AUTO] C = 1 is an operator decision (Q2), not a tuned threshold; the count is per review key, not per turn
const codexReviewBlockLimit = 1

// codexReviewBlockLockWait bounds how long one turn end waits for its review key's
// block lock (SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-004). The critical section is one
// count read and one count write, so the bound is generous. Past it the turn end
// allows and names codexStateBlockRecordFailed: the gate never blocks without a
// recorded count, and never waits without a bound.
const codexReviewBlockLockWait = time.Second

// codexReviewBlockLockRetry is the pause between lock attempts within that bound.
const codexReviewBlockLockRetry = 5 * time.Millisecond

// codexReviewBlockLockStaleAfter is the age past which a block-lock artifact that
// names no owner is abandoned: its holder died between creating the artifact and
// recording its identity, which a live holder does in microseconds.
const codexReviewBlockLockStaleAfter = 30 * time.Second

// codexReviewBlockLockBreakable says whether a block-lock artifact whose holder
// died may be broken. Only the atomic-create substrate (Windows) leaves an artifact
// that outlives its holder: on Unix the kernel releases the flock when the process
// exits, so a held artifact is never stale there. This is the platform gate
// ClearStaleGateLock states. It is a variable so the breaking path can be exercised
// on any platform.
var codexReviewBlockLockBreakable = runtime.GOOS == "windows"

// The turn-end state names (SPEC-REVIEW-GATE-REPAIR-001 design.md §7). Each names
// why a turn end was allowed without a fresh passing receipt (REQ-RG-007). They
// are the systemMessage vocabulary, so the names are the contract readers match on.
const (
	codexStateAbsent            = "absent"
	codexStateStale             = "stale"
	codexStateKeyMismatch       = "key-mismatch"
	codexStateInconclusive      = "inconclusive"
	codexStateBlockExhausted    = "차단 횟수 소진"
	codexStateBlockRecordFailed = "차단 횟수 기록 실패"
	codexStateSwitchOff         = "스위치 꺼짐"
	codexStateNoChange          = "검토할 변경 없음"
	codexStateStopHookRelease   = "stop-hook 해제"
	codexStateReviewerAbsent    = "reviewer 없음"
	codexStateStaleBinarySkip   = "stale 바이너리 건너뜀"
	codexStateTreeScopeSkip     = "tree_scope 건너뜀"
)

// HandleCodexReviewGate is the Stop-hook gate logic. ALLOW/BLOCK contract:
//   - empty HookOutput ({})                            = ALLOW (Claude may stop)
//   - {Decision: "block", Reason: "..."}               = BLOCK (keep working)
//
// An ALLOW may carry a systemMessage that names the state it allowed under
// (SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-007). The decision itself is unchanged by it.
//
// Decision order (AC-MCP-009 self-gate + AC-MCP-010 opt-in + REQ-MCP-012
// fail-open; scope resolution per SPEC-CODEX-GATE-SCOPE-001):
//  1. gate disabled (config off)            → ALLOW (opt-in default-off, C6)
//  2. stop_hook_active (loop prevention)    → ALLOW (mandatory CC protocol)
//  3. scope resolution (REQ-CGS-001)        → card | tree, from the session tree
//     (REQ-CGS-005); class + basis logged (REQ-CGS-010), env context only
//     3a. tree class, no WT- branch, tree_scope skip → ALLOW before the self-gate
//     (SPEC-CODEX-REVIEW-OWNERSHIP-001 REQ-CRO-002; one policy row logged)
//  4. no reviewable change IN THE SCOPE     → ALLOW (self-gate; no false block)
//  5. codex missing                         → ALLOW (fail-open; no reviewer to kick)
//  6. fresh receipt for the tree key        → reuse its verdict (REQ-GBN-001):
//     FAIL blocks carrying the preserved detail while the key's block count is
//     below C (REQ-RG-003) and otherwise ALLOWs naming why (REQ-RG-004): the
//     exhausted count, or the record failure when the count cannot be read or
//     recorded, or the key's block lock cannot be taken within its bound; PASS
//     clears the key's block count; inconclusive ALLOWs naming inconclusive
//  7. cache miss                            → kick the review in the background
//     and ALLOW (REQ-GBN-002); the verdict is enforced at the NEXT turn entry
//     (HandleCodexReviewEntry). A failed kick allows (AC-GBN-004). The allow
//     names the key's state: absent, stale, or key-mismatch (REQ-RG-007).
//
// `enabled` is read by the caller (runCodexReviewGate via
// readCodexReviewGateEnabled) and passed in; it is re-checked here as
// defense-in-depth. The only config I/O in this function is step 3a's read of
// tree_scope, and only for a tree-class session: the root it reads is the one
// the caller read `enabled` from (reviewGateConfigRoot(projectDir)), through an
// injectable reader, so the logic stays testable without a real config tree.
func HandleCodexReviewGate(input *hook.HookInput, enabled bool, projectDir string) (*hook.HookOutput, error) {
	allow := &hook.HookOutput{}
	if !enabled {
		return codexGateAllowed(codexStateSwitchOff), nil // (1) opt-in default-off
	}
	if input != nil && input.StopHookActive {
		return codexGateAllowed(codexStateStopHookRelease), nil // (2) loop prevention — never re-block an already-continuing turn
	}
	// (3) The scope is determined BEFORE any review or consult (REQ-CGS-001),
	// from the SESSION's working-directory tree — never from a spawn-frozen
	// CLAUDE_PROJECT_DIR naming a different tree (REQ-CGS-005). One resolver
	// serves both execution paths (REQ-CGS-009).
	scope := reviewScopeResolver(reviewScopeSessionDir(input, projectDir))
	reviewGateScopeLogger(scope, reviewGateEnvContext())
	// The state root anchors on the git toplevel for a tree-scope session: a
	// subdirectory CWD would root .moai/state inside the subdirectory, where
	// the repo's .gitignore does not cover it — the receipt write itself would
	// then move the tree key and the next turn entry could never match the
	// verdict. Run, store, and consult all use the same git root.
	if scope.Class == reviewScopeTree {
		scope.Dir = reviewExclusionRoot(scope.Dir)
	}
	// (3a-0) The binary-age policy (card t1528): a gate binary whose build
	// commit predates the session tree judges the tree with policy older than
	// the code under review — skip before any scope-dependent policy runs.
	// Shared with the Codex Stop-chain path (REQ-CRO-006).
	if staleBinarySkipApplies(scope.Dir) {
		return codexGateAllowed(codexStateStaleBinarySkip), nil
	}
	// (3a) The tree_scope policy: a tree-class session with no WT- evidence has
	// no card to attribute its tree to. The read root is the one `enabled` came
	// from (reviewGateConfigRoot), resolved only when the class is tree.
	if treeScopeSkipApplies(scope, func() string { return reviewGateConfigRoot(projectDir) }) {
		return codexGateAllowed(codexStateTreeScopeSkip), nil
	}
	if !reviewGateScopedChangeDetector(scope) {
		return codexGateAllowed(codexStateNoChange), nil // (4) scoped self-gate — nothing reviewable in the session's scope ⇒ no false block
	}

	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return codexGateAllowed(codexStateReviewerAbsent), nil // (5) fail-open: a missing reviewer must not trap the session
	}

	// (4a) SPEC-GATE-BOTTLENECK-001 REQ-GBN-001 — the tree-keyed reuse cache:
	// a FRESH receipt for the resolved scope's key (HEAD + tree digest or the
	// card-scope binding) is the same reviewer judging the same code, so its
	// verdict is reused without the RPC. The key covers HEAD, the porcelain
	// digest, the review selection config, and the reviewer version — an
	// untouched tree cannot shed a fail, and any edit (or a reviewer upgrade)
	// makes the key stale. The consult needs no review-length budget: nothing
	// below runs an RPC anymore (the 900s context left with the live arm,
	// REQ-GBN-002).
	state, stateErr := codexReviewReceiptStateForScope(context.Background(), scope, binaryPath)
	// keyState names the key's receipt state for the allows below (REQ-RG-007).
	// With no key there is no receipt for it, which reads as absent.
	keyState := codexStateAbsent
	if stateErr == nil {
		now := time.Now()
		stored := verify.LoadReceipt(scope.Dir, state)
		if chk := verify.CheckReceipt(stored, state, now, codexReviewCacheTTL); chk.Run {
			codexReviewCacheSkips.Add(1)
			switch {
			case isBlockVerdict(chk.Receipt.Verdict):
				// REQ-RG-003 / REQ-RG-004: a fail blocks while the key is under C
				// and otherwise allows, naming the exhausted count.
				return codexReviewGateFail(scope.Dir, state, *chk.Receipt, now), nil
			case isPassVerdict(chk.Receipt.Verdict):
				codexReviewClearBlocks(scope.Dir, state) // REQ-RG-004: a fresh pass ends the key's block history
				_, _ = fmt.Fprintf(os.Stderr, "codex review gate: reusing the cached %s verdict for the unchanged tree (skip #%d)\n",
					chk.Receipt.Verdict, codexReviewCacheSkips.Load())
				return allow, nil
			default:
				// inconclusive, and any verdict that is neither pass nor fail, allows
				// naming the state (REQ-RG-007); it is never read as a pass.
				_, _ = fmt.Fprintf(os.Stderr, "codex review gate: reusing the cached %s verdict for the unchanged tree (skip #%d)\n",
					chk.Receipt.Verdict, codexReviewCacheSkips.Load())
				return codexGateAllowed(codexStateInconclusive), nil
			}
		}
		keyState = codexReviewKeyState(stored, state, now)
	}
	// A state or cache-read failure falls through to the kick — the cache is
	// an accelerator, never an authority (fail-open, REQ-GBN-004).

	// (4b) SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 — the delayed block: a cache
	// miss no longer runs the synchronous review (the measured bottleneck:
	// 11 of 12 Stops blocked, 14m13s). The gate kicks the SAME review the
	// receipt producer runs (`moai verify codex-review`) as a detached
	// background process and ALLOWS the turn. The verdict is enforced at the
	// NEXT turn entry (HandleCodexReviewEntry) and reused from this cache at
	// the next Stop. A kick for THIS tree state already in flight is not
	// repeated (the turn-end gate's P2: consecutive Stops over the unchanged
	// tree deduplicated by the per-key marker). A failed kick is fail-open
	// (AC-GBN-004): the marker is removed — retryable — the allow stands.
	// `async:true` is deliberately not adopted — an async hook can only
	// deliver additionalContext and would surrender the block capability
	// (REQ-GBN-002).
	inFlight, markerPath := kickInFlight(scope.Dir, state, time.Now())
	if inFlight {
		_, _ = fmt.Fprintln(os.Stderr, "codex review gate: a background review for this tree state is already in flight; not re-kicking")
		return codexGateAllowed(keyState), nil
	}
	if kickErr := codexReviewBackgroundKick(scope.Dir); kickErr != nil {
		if markerPath != "" {
			_ = os.Remove(markerPath) // a failed start is retryable on the next Stop
		}
		_, _ = fmt.Fprintf(os.Stderr, "codex review gate: background review not started (%v) — allowing; the next Stop re-kicks\n", kickErr)
		return codexGateAllowed(keyState), nil
	}
	_, _ = fmt.Fprintln(os.Stderr, "codex review gate: background review started for this tree state; the verdict is enforced at the next turn entry (codex-review-entry)")
	return codexGateAllowed(keyState), nil
}

// isBlockVerdict reports whether a codex verdict should BLOCK the session. A
// BLOCK fires ONLY on a clear failure; the inconclusive fail-open value and any
// non-fail verdict ALLOW. Matched case-insensitively on a "fail" prefix so
// "fail" / "FAIL" / "failed" all block while "inconclusive" / "pass" do not.
func isBlockVerdict(verdict string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(verdict)), "fail")
}

// isPassVerdict reports whether a stored verdict is a pass, matched the way
// isBlockVerdict matches a fail: case-insensitively, ignoring surrounding space.
func isPassVerdict(verdict string) bool {
	return strings.EqualFold(strings.TrimSpace(verdict), codexReviewVerdictPass)
}

// codexGateAllowed is an ALLOW that names the state it allowed under
// (REQ-RG-007). The decision stays an empty allow (REQ-RG-006): the message is
// the only addition. A fresh pass never reaches this helper, so it stays silent.
func codexGateAllowed(state string) *hook.HookOutput {
	return &hook.HookOutput{
		SystemMessage: "codex review gate: turn allowed without a fresh pass (state: " + state + ")",
	}
}

// codexReviewBlockLockPath is the lock artifact that guards one review key's block
// count. It sits beside the count record, with a ".lock" suffix.
func codexReviewBlockLockPath(dir string, state verify.ReceiptState) string {
	return codexReviewBlockPath(dir, state) + ".lock"
}

// acquireCodexBlockLock takes the review key's block lock, waiting at most
// codexReviewBlockLockWait. The try primitive is acquireGateLockNoFollowImpl, the
// gate-run lock's open that refuses a symbolic-link path (flock on Unix, exclusive
// create on Windows); this helper adds the bounded wait and the stale-owner break.
// ok is false when the lock is not taken within the bound or the lock machinery
// fails (a refused path included), and the caller then allows (REQ-RG-004): a
// turn end never blocks without a recorded count.
//
// @MX:WARN: [AUTO] cross-process lock on the per-key block count; a holder that is alive but stalled past the bound makes contending turn ends allow instead of block
// @MX:REASON: the check and the record of one block must be a single step across processes (REQ-RG-004); the bound and the dead-owner break keep a stuck or crashed holder from wedging the gate
func acquireCodexBlockLock(dir string, state verify.ReceiptState) (release func(), ok bool) {
	path := codexReviewBlockLockPath(dir, state)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, false
	}
	deadline := time.Now().Add(codexReviewBlockLockWait)
	for {
		impl, err := acquireGateLockNoFollowImpl(path)
		if err == nil {
			return func() { _ = impl.release() }, true
		}
		if !errors.Is(err, ErrGateLockHeld) {
			return nil, false // the lock machinery failed; this is not contention
		}
		if time.Now().After(deadline) {
			return nil, false
		}
		if !codexReviewBreakStaleBlockLock(path, time.Now()) {
			time.Sleep(codexReviewBlockLockRetry)
		}
	}
}

// codexReviewBreakStaleBlockLock removes a block-lock artifact whose holder died,
// and reports whether it did (REQ-RG-004: a lock left by a crashed process must
// eventually be broken). An artifact is broken only when the owner it names is
// positively observed dead, or when it names no owner and is older than
// codexReviewBlockLockStaleAfter. A live owner is never broken.
//
// The removal follows a re-read that must find the artifact unchanged, so a lock
// that changed hands during the inspection is kept. The re-read narrows the
// inspection-to-removal window without closing it, because removal takes a path.
// That residual is the one ClearStaleGateLock states.
func codexReviewBreakStaleBlockLock(path string, now time.Time) bool {
	if !codexReviewBlockLockBreakable {
		return false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var stale bool
	if owner, perr := parseGateLockOwner(raw); perr == nil {
		stale = !factory.FactoryProcessAlive(owner.PID)
	} else if info, serr := os.Stat(path); serr == nil {
		stale = now.Sub(info.ModTime()) > codexReviewBlockLockStaleAfter
	}
	if !stale {
		return false
	}
	again, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, again) {
		return false
	}
	return os.Remove(path) == nil
}

// codexReviewReserveBlock is the check and the record of one turn-end block, made as
// ONE step under the review key's block lock (REQ-RG-003, REQ-RG-004). It reports
// whether the turn end blocks; when it does, the block is already counted. When it
// does not, allowState names why the stop is allowed (REQ-RG-007). A count that
// cannot be read or recorded, and a lock that cannot be taken within the bound, all
// name the record failure, never the exhausted count.
func codexReviewReserveBlock(dir string, state verify.ReceiptState, now time.Time) (blocked bool, allowState string) {
	release, ok := acquireCodexBlockLock(dir, state)
	if !ok {
		return false, codexStateBlockRecordFailed
	}
	defer release()
	blocks, ok := codexReviewBlockCount(dir, state)
	if !ok {
		return false, codexStateBlockRecordFailed
	}
	if blocks >= codexReviewBlockLimit {
		return false, codexStateBlockExhausted
	}
	if err := codexReviewRecordBlock(dir, state, blocks+1, now); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex review gate: block count not recorded (%v) — allowing (fail-open)\n", err)
		return false, codexStateBlockRecordFailed
	}
	return true, ""
}

// codexReviewGateFail decides a fresh FAIL at turn end (REQ-RG-003, REQ-RG-004).
// While the key's block count is below C it blocks, carrying the preserved fail
// detail. The check and the record of the block are one step under the key's block
// lock (codexReviewReserveBlock), so simultaneous turn ends cannot each read the
// same count and each block. Every allow here is fail-open and names its own state
// (REQ-RG-007): at C the allow names the exhausted count; a count that cannot be
// read or recorded, and a lock that cannot be taken in time, name the record
// failure instead. An unreadable, unwritable, or contended count is never reported
// as exhausted, and a count the gate cannot prove is below C never blocks.
func codexReviewGateFail(dir string, state verify.ReceiptState, r verify.Receipt, now time.Time) *hook.HookOutput {
	blocked, allowState := codexReviewReserveBlock(dir, state, now)
	if !blocked {
		return codexGateAllowed(allowState)
	}
	// The cached block still says WHAT to fix: the fail's summary and findings
	// were preserved at record time by the receipt producer (the receipt store
	// carries no free text). An unreadable detail file keeps the bare verdict —
	// fail-open, never invented.
	reason := "codex review gate (cached verdict): " + r.Verdict
	if detail := codexReviewCachedDetail(dir, r); detail != "" {
		reason += "\n\n" + detail
	}
	return &hook.HookOutput{
		Decision: hook.DecisionBlock,
		Reason:   reason,
	}
}

// codexReviewKeyState names why the current review key has no fresh receipt
// (design.md §1): absent when the store holds no receipt for the key, stale when
// every bound field matches and only the TTL has run out, and key-mismatch when a
// receipt is held but a bound field differs. The store is keyed by head and tree
// digest, so a receipt recorded for an earlier tree reads as absent here.
func codexReviewKeyState(stored *verify.Receipt, state verify.ReceiptState, now time.Time) string {
	if stored == nil {
		return codexStateAbsent
	}
	// An unbounded TTL isolates the bound-field comparison from the freshness leg:
	// the check passes exactly when only the TTL failed.
	if verify.CheckReceipt(stored, state, now, math.MaxInt64).Run {
		return codexStateStale
	}
	return codexStateKeyMismatch
}

// codexReviewBlockRecord is the per-key block count (REQ-RG-003, REQ-RG-004): how
// many turn-end blocks the review key has taken. The five receipt fields are kept
// verbatim, as the key's delivery record keeps them, so the record reads back
// against the key it counts.
type codexReviewBlockRecord struct {
	Head         string    `json:"head"`
	TreeDigest   string    `json:"tree_digest"`
	ConfigDigest string    `json:"config_digest"`
	Command      string    `json:"command"`
	ToolVersion  string    `json:"tool_version"`
	Blocks       int       `json:"blocks"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// codexReviewBlockPath is the block-count record of one review key. It sits in
// the same state directory as the key's delivery record and takes that record's
// name with the "blocked-" prefix, so the two records cannot disagree on what one
// key is.
func codexReviewBlockPath(dir string, state verify.ReceiptState) string {
	delivery := codexReviewDeliveryPath(dir, state)
	return filepath.Join(filepath.Dir(delivery), strings.Replace(filepath.Base(delivery), "delivered-", "blocked-", 1))
}

// codexReviewBlockCount reports how many turn-end blocks the key has taken. A
// missing record is zero. An unreadable or malformed record reports ok=false: the
// gate cannot prove the count is below C, so the caller allows rather than block.
// The read is non-blocking and regular-file only (readManifestRecord), so a FIFO at
// the record path reads as unreadable instead of stalling the Stop hook (card t1619,
// codex review round 2).
func codexReviewBlockCount(dir string, state verify.ReceiptState) (int, bool) {
	path := codexReviewBlockPath(dir, state)
	raw, ok := readManifestRecord(path)
	if !ok {
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			return 0, true
		}
		return 0, false
	}
	var rec codexReviewBlockRecord
	if err := json.Unmarshal(raw, &rec); err != nil || rec.Blocks < 0 {
		return 0, false
	}
	return rec.Blocks, true
}

// codexReviewRecordBlock stores the key's block count through the shared atomic
// writer, so a reader never sees a torn record. The count it writes is only a
// correct count when the caller holds the key's block lock (codexReviewReserveBlock):
// the read that produced blocks and this write must be one step.
func codexReviewRecordBlock(dir string, state verify.ReceiptState, blocks int, now time.Time) error {
	rec := codexReviewBlockRecord{
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		Blocks:       blocks,
		UpdatedAt:    now.UTC(),
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return writeFileAtomic(codexReviewBlockPath(dir, state), append(b, '\n'), 0o644)
}

// codexReviewClearBlocks removes the key's block count (REQ-RG-004): a fresh pass for
// the key ends its block history. The removal takes the block lock when it can, so
// it cannot interleave with a block's read and record and put the count back after
// the pass. When the lock is not taken within the bound, the record is removed
// anyway: a pass is never held back by lock contention. A missing record is already
// clear.
func codexReviewClearBlocks(dir string, state verify.ReceiptState) {
	if release, ok := acquireCodexBlockLock(dir, state); ok {
		defer release()
	}
	_ = os.Remove(codexReviewBlockPath(dir, state))
}

// hasReviewableChanges reports whether the working tree at projectDir carries
// any uncommitted change to a NON-runtime-managed path. It shells out to
// `git status --porcelain` (the SAME signal internal/hook/security turn-review
// uses). Fail-open: a non-git dir, missing git, or git error ⇒ false (nothing
// reviewable ⇒ ALLOW), so the gate never false-blocks outside a git repo.
func hasReviewableChanges(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	out, err := exec.Command("git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return reviewableFromPorcelain(string(out))
}

// reviewableFromPorcelain is the pure parser over a `git status --porcelain`
// payload: it returns true iff at least one changed path is NOT under a
// runtime-managed prefix. Split out so the filtering rule is unit-testable
// without a git repo. Porcelain line shape is fixed-width: columns 0-1 are the
// XY status (either may be a space), column 2 is a space, and the path begins
// at column 3 — so the path MUST be sliced from the raw line, NOT from a
// TrimSpace'd copy (TrimSpace would strip a leading-space status and shift the
// path off by one, dropping its leading "." and defeating the prefix filter).
// Renames use "XY <old> -> <new>"; the prefix check against <old> is sufficient.
//
// This is the SHARED baseline parser (card-review repair R1): it consults only
// the runtime-managed state prefixes, never the tree-only config surfaces —
// the multi-review gate (HandleMultiReviewGate) and Codex Stop-chain member 7
// consume it through the reviewGateChangeDetector seam, and a config-only
// exclusion here let them silently allow over a stored required FAIL. The
// tree-scope counterpart is treeConfigOnlyFromPorcelain below.
func reviewableFromPorcelain(porcelain string) bool {
	for _, raw := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if len(raw) <= 3 {
			continue // malformed — no path column
		}
		path := strings.TrimSpace(raw[3:]) // path column; trim trailing space / CR only
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = path[:idx] // rename source for the prefix check
		}
		if path == "" || isRuntimeManagedPath(path) {
			continue
		}
		return true
	}
	return false
}

// treeExcludedPath reports whether one porcelain path is excluded on the TREE
// path: the shared runtime-managed state prefixes plus the tree-only config
// surfaces.
func treeExcludedPath(path string) bool {
	return isRuntimeManagedPath(path) || isTreeRuntimeConfigPath(path)
}

// treeConfigOnlyFromPorcelain is the pure TREE-scope counterpart of
// reviewableFromPorcelain (card-review repair R1): it reports whether a
// `git status --porcelain` payload carries at least one change and EVERY one
// of them is excluded on the tree path — the config-only turn the tree
// self-gate must not review (REQ-CGSC-007). An empty payload and any
// non-excluded record read false, so the probe can only narrow the shared
// detector's answer, never widen it. A rename record is config-only only when
// BOTH sides are excluded (card-review repair R5):
// `.claude/settings.json -> main.go` is a real source change — the
// destination must not inherit the source's exclusion.
func treeConfigOnlyFromPorcelain(porcelain string) bool {
	any := false
	for _, raw := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if len(raw) <= 3 {
			continue // malformed — no path column
		}
		path := strings.TrimSpace(raw[3:]) // path column; trim trailing space / CR only
		excluded := false
		if idx := strings.Index(path, " -> "); idx >= 0 {
			excluded = treeExcludedPath(strings.TrimSpace(path[:idx])) &&
				treeExcludedPath(strings.TrimSpace(path[idx+4:]))
		} else {
			excluded = treeExcludedPath(path)
		}
		if !excluded {
			return false
		}
		any = true
	}
	return any
}

// treeConfigOnlyChanges reports whether the working tree at dir carries ONLY
// tree-excluded changes (treeConfigOnlyFromPorcelain over `git status
// --porcelain`). Fail-open in the PRESERVE direction: a measurement failure
// reads false — the exclusion never widens on an unreadable tree, so the
// scoped self-gate keeps the shared detector's answer there.
//
// The untracked leg is collected at FILE level (--untracked-files=all,
// card-review repair round 2, N5): default porcelain collapses a fully
// untracked .moai/ tree to `?? .moai/`, an entry the exclusion sets cannot
// see .moai/config/ inside, so a config-only untracked change counted as
// reviewable. Individual file paths keep the probe honest — the
// cardChangedPaths precedent.
func treeConfigOnlyChanges(dir string) bool {
	if dir == "" {
		return false
	}
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain", "--untracked-files=all").Output()
	if err != nil {
		return false
	}
	return treeConfigOnlyFromPorcelain(string(out))
}

// isRuntimeManagedPath reports whether path falls under a hook/session-written
// prefix the gate must ignore (otherwise every turn's .moai/state drift would
// trip the self-gate).
func isRuntimeManagedPath(path string) bool {
	for _, p := range reviewGateRuntimePrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// isTreeRuntimeConfigPath reports whether path falls under a runtime-managed
// configuration surface the TREE path ignores (REQ-CGSC-007). The card path's
// filter never consults this: a card commit under .moai/config/ stays card
// work (REQ-CGSC-005). The settings FILE matches exactly (card-review repair
// R4); the managed config DIRECTORY matches by prefix.
func isTreeRuntimeConfigPath(path string) bool {
	for _, p := range reviewGateTreeConfigExactPaths {
		if path == p {
			return true
		}
	}
	for _, p := range reviewGateTreeConfigDirPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// normalizeFindingPath brings a review finding's file anchor into the
// repo-relative slash form the exclusion sets are written in (card-review
// repair R3): a "./"-prefixed relative is stripped, and an absolute path is
// relativized against the reviewed scope's tree. A path OUTSIDE the scope
// relativizes to a "../" form that matches no prefix — the fail-closed block.
func normalizeFindingPath(file, dir string) string {
	p := strings.TrimSpace(file)
	if p == "" {
		return ""
	}
	p = strings.TrimPrefix(p, "./")
	if filepath.IsAbs(p) && dir != "" {
		if rel, err := filepath.Rel(dir, filepath.FromSlash(p)); err == nil {
			p = filepath.ToSlash(rel)
		}
	}
	return p
}

// reviewExclusionRoot returns the tree the finding-path exclusion comparison
// anchors on (card-review repair round 2, N3): the GIT REPOSITORY ROOT of the
// reviewed scope's tree, never the scope dir itself. A session sitting in a
// subdirectory made an absolute repo-root config path relativize to a "../"
// form that escapes the exclusion sets, and a config-only FAIL stayed BLOCK.
// Fail-open to dir when the root cannot be resolved (a non-git dir), so the
// anchor never moves on an unreadable tree.
func reviewExclusionRoot(dir string) string {
	if dir == "" {
		return dir
	}
	if root, err := reviewScopeGit(dir, "rev-parse", "--show-toplevel"); err == nil && root != "" {
		return root
	}
	return dir
}

// codexReviewStateRoot is the ONE root the receipt producer and the Stop
// chain's receipt reader share (SPEC-MCP-NEUTRALITY-001 REQ-MN-025): the git
// toplevel of the session tree, falling back to the tree itself when git
// cannot resolve one. A session started in a subdirectory reads the receipt
// the producer wrote for the same state instead of a false "receipt absent".
func codexReviewStateRoot(dir string) string {
	return reviewExclusionRoot(dir)
}

// runtimeConfigOnlyFindings reports whether EVERY finding of a review targets
// only the runtime-managed configuration surfaces, and returns the distinct
// targets when so (REQ-CGSC-008). A review with no findings at all is NOT
// config-only: a fail verdict behind an unparseable findings list is the
// contradiction state, never a licence to allow. Findings without a file
// anchor, and anchors outside the surfaces, keep the review's block. The
// anchor is normalized against the reviewed scope's tree dir before the
// comparison (card-review repair R3), so an absolute anchor reclassifies
// exactly as its relative twin. The anchor must also be UNAMBIGUOUS
// (card-review repair round 2, N1): codexFindingsOf leaves File empty when a
// finding message carries several distinct path candidates, and an
// anchor-less finding keeps the strict disposition below.
func runtimeConfigOnlyFindings(findings []Finding, dir string) ([]string, bool) {
	if len(findings) == 0 {
		return nil, false
	}
	anchor := reviewExclusionRoot(dir)
	var targets []string
	seen := make(map[string]bool)
	for _, f := range findings {
		file := normalizeFindingPath(f.File, anchor)
		if file == "" || !isTreeRuntimeConfigPath(file) {
			return nil, false
		}
		if !seen[file] {
			seen[file] = true
			targets = append(targets, file)
		}
	}
	return targets, true
}

// logRuntimeDriftReclassification writes the reclassification row to stderr,
// the gate's diagnostic channel (REQ-CGSC-011): the reason in the SPEC's own
// words and the targeted paths, distinguishable from a skip row. stdout stays
// the pure HookOutput contract.
func logRuntimeDriftReclassification(scope reviewScope, targets []string) {
	row := map[string]any{
		"gate":         "codex-review-gate",
		"scope":        scope.Class,
		"reclassified": "runtime-managed drift",
		"targets":      targets,
	}
	b, err := json.Marshal(row)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}

// runCodexReviewGate is the cobra RunE for `moai hook codex-review-gate`
// (SPEC-MOAI-MCP-SERVER-001 M2 REQ-MCP-008). It reads the Stop-hook stdin JSON,
// resolves the project root, reads the workflow.codex.review_gate.enabled
// opt-in flag, and forwards to HandleCodexReviewGate. The output is emitted as
// JSON on stdout (the Stop-hook contract). Fail-open: any error logs to stderr
// and exits 0 so the Stop pipeline is never broken (REQ-MCP-012).
func runCodexReviewGate(cmd *cobra.Command, _ []string) error {
	input, err := readHookInput(cmd.InOrStdin())
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-gate: invalid stdin JSON (%v); emitting default output\n", err)
		// Fail-open: emit an empty ALLOW.
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	projectDir := resolveProjectDirFromInput(input)
	enabled := readCodexReviewGateEnabled(reviewGateConfigRoot(projectDir))
	out, gateErr := HandleCodexReviewGate(input, enabled, projectDir)
	if gateErr != nil {
		// Fail-open: a handler error MUST NOT trap the Stop pipeline.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-gate: error:", gateErr)
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	if out == nil {
		out = &hook.HookOutput{}
	}
	return emitHookOutput(cmd.OutOrStdout(), out)
}

// reviewGateConfigRoot returns the root whose workflow config carries the Stop
// review gates' opt-in flags for a resolved projectDir: the primary checkout
// for a config-orphaned linked worktree, projectDir itself otherwise, and ""
// (gate disabled, the fail-open direction of these gates) when that primary
// cannot be identified (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-008). The order
// in which projectDir is picked, and the tree the codex review targets, do
// not change.
func reviewGateConfigRoot(projectDir string) string {
	root, err := auditreceipt.StoreRoot(projectDir)
	if err != nil {
		return ""
	}
	return root
}

// readHookInput reads and parses the hook stdin JSON into a HookInput. It is
// broken out so the fail-open path can distinguish "could not read stdin" from
// "handler errored".
func readHookInput(r io.Reader) (*hook.HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	var input hook.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parse stdin: %w", err)
	}
	return &input, nil
}

// emitHookOutput serializes out as JSON to w. A nil-safe helper used by the
// fail-open path and the success path alike.
func emitHookOutput(w io.Writer, out *hook.HookOutput) error {
	enc := json.NewEncoder(w)
	return enc.Encode(out)
}

// resolveProjectDirFromInput picks the best available project dir from the hook
// input: ProjectDir first, then CWD, then the CLAUDE_PROJECT_DIR environment
// variable, then empty (the gate self-gates on empty).
//
// The CWD arm is load-bearing: Claude Code sends `cwd` in the Stop payload and
// treats `project_dir` as a legacy field (internal/hook/types.go), so a
// ProjectDir-only resolution returned "" on every real invocation and both gate
// readers fail-CLOSED to false no matter how the project was configured. The
// env arm mirrors how the shell wrappers resolve the project root, so the two
// layers agree on which workflow.yaml they are reading.
func resolveProjectDirFromInput(input *hook.HookInput) string {
	if input != nil {
		if input.ProjectDir != "" {
			return input.ProjectDir
		}
		if input.CWD != "" {
			return input.CWD
		}
	}
	return os.Getenv("CLAUDE_PROJECT_DIR")
}
