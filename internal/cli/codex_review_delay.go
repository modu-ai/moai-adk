package cli

// codex_review_delay.go — the delayed-verdict axis of the turn-end review gate
// (SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 / AC-GBN-004; the entry half amended by
// SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-001 and REQ-RG-002). The Stop gate no
// longer runs the synchronous codex RPC (the measured bottleneck: 11 of 12
// Stops blocked, 14m13s): a cache-miss Stop KICKS the same review the receipt
// producer runs (`moai verify codex-review`) as a detached background process
// and allows the turn. HandleCodexReviewEntry enforces the verdict at the NEXT
// turn entry, but it never blocks a prompt: a fresh FAIL's finding detail goes
// to the model as additional context once per review key, and a per-key
// delivery record keeps later prompts of that key from repeating it. Every
// other state allows (fail-open). The Stop side keeps its synchronous block (its
// per-key limit is SPEC-REVIEW-GATE-REPAIR-001 REQ-RG-003, a later milestone);
// the entry hook's timeout is outside this change (F2).
//
// Both halves share the receipt store, the scope resolver, and the receipt
// producer with the Stop gate (codex_review_gate.go / codex_review_receipt.go)
// — one store, both harnesses, no second judging path.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
// binds (HEAD + porcelain digest), under the runtime-managed state dir, with
// the review budget BUCKET in the name. Deduplication works within one
// bucket: the same tree state in the same budget window maps to the same
// path, so the exclusive create is the only mutation and the dedup cannot
// race. When the bucket rolls over (the budget elapsed) the path name
// changes and the create naturally succeeds — a dead review re-kicks within
// one budget without any take-over-and-recreate protocol, whose empty-path
// window let two callers own the kick (the turn-end gate's r5, then the
// marker-never-expires regression its removal caused, r7).
func codexReviewKickMarkerPath(dir, head, digest string, now time.Time) string {
	bucket := now.Unix() / int64(config.DefaultCodexReviewGateTimeout.Seconds())
	return filepath.Join(dir, ".moai", "state", "verify", "codex-review",
		fmt.Sprintf("%s-%s.kick-%d", head, digest, bucket))
}

// kickInFlight reports whether a kick for this exact tree state is already in
// flight in the CURRENT budget bucket, and otherwise records the marker for
// the kick the caller is about to start. Acquisition is a single exclusive
// create — the only mutation the marker ever takes; the previous bucket's
// marker is garbage-collected best-effort (a different path, so the removal
// cannot race the current bucket's create). An unkeyed state and an
// unwritable marker are fail-open toward reviewing: the kick proceeds
// undeduplicated.
func kickInFlight(dir string, state verify.ReceiptState, now time.Time) (bool, string) {
	if state.Head == "" || state.TreeDigest == "" {
		return false, "" // unkeyed ⇒ no marker; the kick proceeds undeduplicated
	}
	path := codexReviewKickMarkerPath(dir, state.Head, state.TreeDigest, now)
	// Garbage-collect the aged-out bucket's marker: a different path, so this
	// removal never races the create below (best-effort; a missed sweep only
	// leaves state-dir litter).
	prev := codexReviewKickMarkerPath(dir, state.Head, state.TreeDigest,
		now.Add(-config.DefaultCodexReviewGateTimeout))
	_ = os.Remove(prev)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, ""
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if !os.IsExist(err) {
			return false, "" // unwritable ⇒ fail-open toward reviewing
		}
		return true, path // a kick for this tree state is in flight this bucket
	}
	_, _ = f.WriteString(time.Now().UTC().Format(time.RFC3339))
	_ = f.Close()
	return false, path // we own the kick
}

// codexReviewDeliveryRecord is the per-key delivery record: the review key a
// fail detail was delivered for, and when (SPEC-REVIEW-GATE-REPAIR-001
// REQ-RG-002). The five receipt fields are kept verbatim, so the record reads
// back against the receipt it came from.
type codexReviewDeliveryRecord struct {
	Head         string    `json:"head"`
	TreeDigest   string    `json:"tree_digest"`
	ConfigDigest string    `json:"config_digest"`
	Command      string    `json:"command"`
	ToolVersion  string    `json:"tool_version"`
	DeliveredAt  time.Time `json:"delivered_at"`
}

// codexReviewDeliveryPath is the delivery record of one review key, in the same
// runtime-managed state directory as the detail file. The key is the five
// combined receipt fields. Each field is length-prefixed before hashing, so
// distinct keys never share a byte sequence, and a change to any one field
// names a new record: deliveries are per key, never global.
func codexReviewDeliveryPath(dir string, state verify.ReceiptState) string {
	h := sha256.New()
	for _, field := range []string{state.Head, state.TreeDigest, state.ConfigDigest, state.Command, state.ToolVersion} {
		_, _ = fmt.Fprintf(h, "%d:%s;", len(field), field)
	}
	return filepath.Join(dir, ".moai", "state", "verify", "codex-review",
		"delivered-"+hex.EncodeToString(h.Sum(nil))+".json")
}

// codexReviewClaimDelivery reports whether the fail detail may be delivered for
// this review key now, and claims the delivery when it may. The exclusive create
// of the key's record is the claim: the first prompt of a key wins, and a later
// prompt finds the record and delivers nothing. Fail-open toward delivering: a
// record that cannot be written leaves the detail advisory, and it never holds
// the prompt.
func codexReviewClaimDelivery(dir string, state verify.ReceiptState, now time.Time) bool {
	path := codexReviewDeliveryPath(dir, state)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return true
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return !os.IsExist(err) // a record already exists ⇒ this key was delivered before
	}
	rec := codexReviewDeliveryRecord{
		Head: state.Head, TreeDigest: state.TreeDigest, ConfigDigest: state.ConfigDigest,
		Command: state.Command, ToolVersion: state.ToolVersion, DeliveredAt: now.UTC(),
	}
	if b, mErr := json.Marshal(rec); mErr == nil {
		_, _ = f.Write(append(b, '\n'))
	}
	_ = f.Close()
	return true
}

// HandleCodexReviewEntry is the NEXT-TURN-ENTRY half of the delayed verdict
// (REQ-GBN-002, amended by REQ-RG-001 and REQ-RG-002). The Stop allowed the
// previous turn and kicked a background review; this handler reads the receipt
// that review recorded and reports it against the CURRENT tree state. It never
// blocks a prompt (REQ-RG-001). Output contract:
//   - empty HookOutput ({}) = ALLOW: the prompt proceeds with nothing added
//   - HookSpecificOutput.AdditionalContext = "..." = ALLOW with the fail detail,
//     delivered once for the review key (REQ-RG-002)
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
//  6. fresh FAIL receipt for the current key     → ALLOW. The finding detail is
//     added as context on the key's first prompt, and the delivery record makes
//     later prompts of the same key silent
//  7. anything else (no receipt yet, stale key, pass, inconclusive) → ALLOW;
//     the background review is still running or the tree already moved, and
//     the next Stop re-kicks for the new key
//
// The entry hook never kicks: enforcement and kick-start stay on separate
// hooks, so the prompt path never starts a review it would then have to wait on.
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
			// (6) the fresh FAIL the previous turn's background review recorded.
			// It never blocks (REQ-RG-001): the detail goes to the model as
			// additional context, once per review key (REQ-RG-002). The claim is
			// the delivery record; a key that already has one stays silent. The
			// detail file says WHAT to fix (preserved by the receipt producer at
			// record time); an unreadable detail keeps the bare verdict —
			// fail-open, never invented.
			if !codexReviewClaimDelivery(scope.Dir, state, time.Now()) {
				return allow, nil
			}
			ctx := "codex review (delayed verdict — the background review of the previous turn failed): " +
				chk.Receipt.Verdict + ". This prompt is not blocked."
			if detail := codexReviewCachedDetail(scope.Dir, *chk.Receipt); detail != "" {
				ctx += "\n\n" + detail
			}
			return &hook.HookOutput{
				HookSpecificOutput: &hook.HookSpecificOutput{
					HookEventName:     string(hook.EventUserPromptSubmit),
					AdditionalContext: ctx,
				},
			}, nil
		}
		_, _ = fmt.Fprintf(os.Stderr, "codex review entry: background verdict %s for the unchanged tree — allowing\n", chk.Receipt.Verdict)
	}
	// (7) no fresh receipt yet / stale key / pass / inconclusive ⇒ the prompt
	// proceeds; the next Stop re-kicks if the tree moved.
	return allow, nil
}

// runCodexReviewEntry is the cobra RunE for `moai hook codex-review-entry`
// (SPEC-GATE-BOTTLENECK-001 REQ-GBN-002; no-block amendment REQ-RG-001). It
// reads the UserPromptSubmit stdin JSON, resolves the project root, reads the
// same workflow.codex.review_gate.enabled opt-in flag as the Stop gate, and
// forwards to HandleCodexReviewEntry. The output is emitted as JSON on stdout.
// The entry hook never sets the top-level decision; a fresh FAIL's detail rides
// hookSpecificOutput.additionalContext (REQ-RG-002). Fail-open: any error logs
// to stderr and exits 0 so the prompt pipeline is never broken.
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
