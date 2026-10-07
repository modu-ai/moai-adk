package cli

// codex_review_delay.go — the delayed-block axis of the turn-end review gate
// (SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 / AC-GBN-004). The Stop gate no
// longer runs the synchronous codex RPC (the measured bottleneck: 11 of 12
// Stops blocked, 14m13s): a cache-miss Stop KICKS the same review the receipt
// producer runs (`moai verify codex-review`) as a detached background process
// and allows the turn, and the verdict is enforced at the NEXT turn entry by
// HandleCodexReviewEntry — a fresh FAIL blocks the prompt with the preserved
// finding detail, every other state allows (fail-open). `async:true` is
// deliberately not adopted: an async hook can only deliver additionalContext
// and would surrender the block capability (REQ-GBN-002).
//
// Both halves share the receipt store, the scope resolver, and the receipt
// producer with the Stop gate (codex_review_gate.go / codex_review_receipt.go)
// — one store, both harnesses, no second judging path.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// codexReviewBackgroundKick is the injectable "start the review in the
// background" seam. The production default launches the detached
// `moai verify codex-review`; tests record the call instead. A kick error is
// fail-open: the caller allows and the next Stop re-kicks (AC-GBN-004).
var codexReviewBackgroundKick = kickCodexBackgroundReview

// kickCodexBackgroundReview launches `moai verify codex-review
// --project-root <dir>` as a detached child. The hook process IS the moai
// binary (os.Executable), so no binary resolution is needed and the kicked
// review runs the same build that judged the tree. Output appends to
// .moai/logs/codex-review-bg.log (best-effort: a log failure does not stop
// the kick). cmd.Start returns once the process is launched — the child
// outlives the short-lived hook process, which is the point: the Stop is
// never waited on. A launch failure returns the error; the gate allows on it.
func kickCodexBackgroundReview(dir string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("codex review kick: resolve self: %w", err)
	}
	cmd := exec.Command(self, "verify", "codex-review", "--project-root", dir)
	_ = os.MkdirAll(filepath.Join(dir, ".moai", "logs"), 0o755)
	if log, logErr := os.OpenFile(filepath.Join(dir, ".moai", "logs", "codex-review-bg.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); logErr == nil {
		// The child dups the handle at Start; the parent's copy is closed by
		// the defer after Start returns.
		cmd.Stdout = log
		cmd.Stderr = log
		defer func() { _ = log.Close() }()
	}
	return cmd.Start()
}

// codexReviewKickMarkerPath is the per-tree-key in-flight marker a Stop
// writes before launching the background review — the same key the receipt
// binds (HEAD + porcelain digest), under the runtime-managed state dir. Its
// presence within the review budget means a kick for THIS exact tree state is
// already running, and a second Stop over the unchanged tree must not start a
// second review (the turn-end gate's own P2: duplicated RPCs on consecutive
// Stops). The marker is advisory state: it expires by mtime at the review
// budget, a failed start removes it (retryable), and the fresh receipt makes
// it irrelevant — the cache-hit path answers before the marker is consulted.
func codexReviewKickMarkerPath(dir, head, digest string) string {
	return filepath.Join(dir, ".moai", "state", "verify", "codex-review",
		fmt.Sprintf("%s-%s.kick", head, digest))
}

// kickInFlight reports whether a kick for this exact tree state is already in
// flight, and otherwise records the marker for the kick the caller is about
// to start. Acquisition is ATOMIC (exclusive create): two overlapping Stops
// on the same tree race on O_EXCL, and exactly one of them wins the marker —
// the loser reads it as in-flight and does not start a second review. A
// marker older than the review budget reads as a dead review (the kick it
// stood for never recorded) and is replaced for the re-kick; a replace lost
// to a concurrent re-creator reads as in-flight. An unkeyed state (the key
// itself unmeasurable) and an unwritable marker are both fail-open toward
// reviewing: the kick proceeds undeduplicated.
func kickInFlight(dir string, state verify.ReceiptState) (bool, string) {
	if state.Head == "" || state.TreeDigest == "" {
		return false, "" // unkeyed ⇒ no marker; the kick proceeds undeduplicated
	}
	path := codexReviewKickMarkerPath(dir, state.Head, state.TreeDigest)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, ""
	}
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = f.WriteString(time.Now().UTC().Format(time.RFC3339))
			_ = f.Close()
			return false, path // we own the kick
		}
		if !os.IsExist(err) {
			return false, "" // unwritable ⇒ fail-open toward reviewing
		}
		// The marker exists: fresh means in-flight; stale means dead.
		if fi, statErr := os.Stat(path); statErr == nil && time.Since(fi.ModTime()) < config.DefaultCodexReviewGateTimeout {
			return true, path
		}
		_ = os.Remove(path) // dead — race the exclusive create once more
	}
	// The replace race was lost to a concurrent re-creator: treat as in-flight.
	return true, path
}

// HandleCodexReviewEntry is the NEXT-TURN-ENTRY enforcement half of the
// delayed block (REQ-GBN-002). The Stop allowed the previous turn and kicked
// a background review; this handler reads the receipt that review recorded
// and enforces it against the CURRENT tree state. ALLOW/BLOCK contract:
//   - empty HookOutput ({})                            = ALLOW (prompt proceeds)
//   - {Decision: "block", Reason: "..."}               = BLOCK (prompt held)
//
// Decision order (fail-open on every read path, AC-GBN-004):
//  1. gate disabled (same flag as the Stop gate) → ALLOW (opt-in default-off)
//  2. scope resolution                           → card | tree, same resolver
//     as the Stop gate, from the session tree (REQ-CGS-005); stale-binary and
//     tree_scope skips mirror the Stop gate so the two halves skip together
//  3. no reviewable change in the scope          → ALLOW (nothing was kicked)
//  4. codex missing                              → ALLOW (no reviewer, no
//     receipt can ever arrive)
//  5. receipt state unreadable                   → ALLOW (unkeyed verdict)
//  6. fresh FAIL receipt for the current key     → BLOCK with the preserved
//     finding detail (the gate's delayed enforcement point)
//  7. anything else (no receipt yet, stale key, pass, inconclusive) → ALLOW;
//     the background review is still running or the tree already moved, and
//     the next Stop re-kicks for the new key
//
// The entry hook never kicks: enforcement and kick-start stay on separate
// hooks so a blocked prompt cannot also be waiting on a review it started.
func HandleCodexReviewEntry(input *hook.HookInput, enabled bool, projectDir string) (*hook.HookOutput, error) {
	allow := &hook.HookOutput{}
	if !enabled {
		return allow, nil // (1) opt-in default-off
	}
	scope := reviewScopeResolver(reviewScopeSessionDir(input, projectDir))
	reviewGateScopeLogger(scope, reviewGateEnvContext())
	// Same git-root anchoring as the Stop gate: the entry hook must read the
	// receipt from the state root the producer wrote it to.
	if scope.Class == reviewScopeTree {
		scope.Dir = reviewExclusionRoot(scope.Dir)
	}
	if staleBinarySkipApplies(scope.Dir) {
		return allow, nil
	}
	if treeScopeSkipApplies(scope, func() string { return reviewGateConfigRoot(projectDir) }) {
		return allow, nil
	}
	if !reviewGateScopedChangeDetector(scope) {
		return allow, nil // (3) nothing reviewable ⇒ nothing was kicked to enforce
	}
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return allow, nil // (4) no reviewer ⇒ no receipt can ever arrive
	}
	state, stateErr := codexReviewReceiptStateForScope(context.Background(), scope, binaryPath)
	if stateErr != nil {
		return allow, nil // (5) unkeyed ⇒ nothing enforceable
	}
	if chk := verify.CheckReceipt(verify.LoadReceipt(scope.Dir, state), state, time.Now(), codexReviewCacheTTL); chk.Run {
		if isBlockVerdict(chk.Receipt.Verdict) {
			// (6) the delayed enforcement point: the fresh FAIL the previous
			// turn's background review recorded. The detail file says WHAT to
			// fix (preserved by the receipt producer at record time); an
			// unreadable detail keeps the bare verdict — fail-open, never
			// invented.
			reason := "codex review gate (delayed verdict — the background review of the previous turn failed): " + chk.Receipt.Verdict
			if detail := codexReviewCachedDetail(scope.Dir, *chk.Receipt); detail != "" {
				reason += "\n\n" + detail
			}
			return &hook.HookOutput{
				Decision: hook.DecisionBlock,
				Reason:   reason,
			}, nil
		}
		_, _ = fmt.Fprintf(os.Stderr, "codex review entry: background verdict %s for the unchanged tree — allowing\n", chk.Receipt.Verdict)
	}
	// (7) no fresh receipt yet / stale key / pass / inconclusive ⇒ the prompt
	// proceeds; the next Stop re-kicks if the tree moved.
	return allow, nil
}

// runCodexReviewEntry is the cobra RunE for `moai hook codex-review-entry`
// (SPEC-GATE-BOTTLENECK-001 REQ-GBN-002). It reads the UserPromptSubmit stdin
// JSON, resolves the project root, reads the same workflow.codex.review_gate.
// enabled opt-in flag as the Stop gate, and forwards to HandleCodexReviewEntry.
// The output is emitted as JSON on stdout (the hook contract — the entry hook
// blocks via the top-level decision field, the same shape the Stop gate uses).
// Fail-open: any error logs to stderr and exits 0 so the prompt pipeline is
// never broken.
func runCodexReviewEntry(cmd *cobra.Command, _ []string) error {
	input, err := readHookInput(cmd.InOrStdin())
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-entry: invalid stdin JSON (%v); emitting default output\n", err)
		// Fail-open: emit an empty ALLOW.
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	projectDir := resolveProjectDirFromInput(input)
	enabled := readCodexReviewGateEnabled(reviewGateConfigRoot(projectDir))
	out, entryErr := HandleCodexReviewEntry(input, enabled, projectDir)
	if entryErr != nil {
		// Fail-open: a handler error MUST NOT trap the prompt pipeline.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-entry: error:", entryErr)
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	if out == nil {
		out = &hook.HookOutput{}
	}
	return emitHookOutput(cmd.OutOrStdout(), out)
}
