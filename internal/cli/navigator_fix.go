package cli

// navigator_fix CLI: BAS Fix layer-1 + layer-3 entry point
// (SPEC-NAVIGATOR-SYNC-005, M3.2 + M3.5). Sibling to M0's navigator-sync,
// M2's navigator-route, and M4's navigator-tiers Hidden subcommands.
// On-demand ONLY — the default invocation (no --apply) is the deterministic
// draft-request producer (layer 1). It does NOT generate the AI draft (layer 2,
// orchestrator-mediated via manager-develop) and does NOT apply it.
//
// The --apply <draft-id> invocation (M3.5) is the deterministic post-approval
// consumer (layer 3). It validates the approval.json token (AC-NS5-008d) and,
// on a valid token, atomic-renames the approved + scope-conformant draft
// subtrees into the live doc surfaces (AC-NS5-008c). On a missing/invalid
// token it REFUSES (exit non-zero + a message naming the missing/invalid
// approval.json token) — the hard guard, distinct from the fail-open degraded
// paths (REQ-NS5-009, M3.6) which exit 0.
//
// The CLI loads the four read-only inputs (work-items.json + detect JSONL +
// git-diff + nav-graph.json), computes the diff-scope, and emits
// .moai/project/navigator/fix-drafts/<draft-id>/request.json plus a stdout JSON
// signal ({"draft_request_path":..., "status":..., "draft_id":...}) — the
// design.md §A.4 handoff contract the orchestrator consumes to spawn the
// AI-draft delegation.
//
// On-demand ONLY (REQ-NS5-001 / AC-NS5-001b): there is NO PostToolUse hook and
// NO handle-navigator-fix.sh wrapper. The sole entry point is this Hidden cobra
// subcommand. Fail-open (REQ-NS5-009): exit 0 always for the layer-1 producer.
// Provenance is git-sourced (no wall-clock) so two runs on the same HEAD are
// byte-identical.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/navigator/fix"
)

// newNavigatorFixCmd creates the navigator-fix subcommand. Hidden from the
// top-level `moai --help` surface (mirrors navigator-sync / navigator-route /
// navigator-tiers).
func newNavigatorFixCmd() *cobra.Command {
	var (
		projectRoot string
		compareTo   string
		applyDraft  string
	)

	cmd := &cobra.Command{
		Use:    "navigator-fix",
		Short:  "Produce the draft-request manifest for stale doc subtrees",
		Hidden: true,
		Long: `BAS Fix layer-1 producer + layer-3 consumer (sibling of navigator-sync/route/tiers).

Without --apply: loads the four read-only inputs (M2 work-items.json + M1
detect JSONL + git-diff + M0 nav-graph.json), computes the stale doc-subtree
diff-scope, and emits .moai/project/navigator/fix-drafts/<draft-id>/request.json
plus a stdout JSON signal — the handoff contract the orchestrator consumes to
spawn the AI-draft delegation (layer 2). It does NOT generate the draft and
does NOT apply it.

With --apply <draft-id>: validates the approval.json token (AC-NS5-008d) and,
on a valid token, atomic-renames the approved + scope-conformant draft
subtrees into their target live doc surfaces (AC-NS5-008c). On a
missing/invalid token it REFUSES (exit non-zero) — the hard guard.

On-demand ONLY (REQ-NS5-001): no PostToolUse hook, no handle wrapper. Fail-open
(REQ-NS5-009): exit 0 always for the layer-1 producer. Provenance is git-sourced
(no wall-clock) — two runs on the same HEAD are byte-identical.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveNavigatorFixRoot(projectRoot)
			if applyDraft != "" {
				return runNavigatorFixApply(root, applyDraft)
			}
			return runNavigatorFix(root, compareTo)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"project root (defaults to $CLAUDE_PROJECT_DIR then $PWD)")
	cmd.Flags().StringVar(&compareTo, "compare-to", "",
		"baseline commit override (default: nav-graph provenance, then HEAD~1)")
	cmd.Flags().StringVar(&applyDraft, "apply", "",
		"apply draft <draft-id> post-approval (validates approval.json token; AC-NS5-008c/008d)")

	return cmd
}

// resolveNavigatorFixRoot resolves the project root: explicit flag >
// $CLAUDE_PROJECT_DIR > CWD (B7 path resolution — prefer $CLAUDE_PROJECT_DIR,
// CWD fallback OK).
func resolveNavigatorFixRoot(projectRoot string) string {
	root := projectRoot
	if root == "" {
		root = os.Getenv("CLAUDE_PROJECT_DIR")
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			root = "."
		}
	}
	if absRoot, err := filepath.Abs(root); err == nil {
		root = absRoot
	}
	return root
}

// runNavigatorFix is the fail-open layer-1 producer core. It never returns a
// non-nil error to the cobra RunE (REQ-NS5-009): every failure mode degrades
// inside fix.Run to an exit-0 stdout signal. The Result signal is printed to
// stdout so the orchestrator consumes the §A.4 handoff contract.
func runNavigatorFix(projectRoot, compareTo string) error {
	res := fix.Run(fix.Options{ProjectRoot: projectRoot, CompareTo: compareTo})
	if line, err := res.SignalJSON(); err == nil {
		_, _ = fmt.Fprint(os.Stdout, string(line))
	}
	return nil
}

// runNavigatorFixApply is the layer-3 post-approval consumer core. Unlike the
// layer-1 producer (fail-open, exit 0 always), the apply path returns a
// non-nil error on the token-refusal path (AC-NS5-008d — the hard guard, NOT
// fail-open). The cobra RunE propagates the error to a non-zero process exit
// so a bare-shell invocation without a valid approval token is REFUSED.
//
// Exit-code scope (plan.md §F M3.5 last bullet): fail-open (REQ-NS5-009,
// AC-NS5-009, M3.6) is a *degraded* path → exit 0 + advisory. Token-refusal
// (REQ-NS5-008 c4, AC-NS5-008d, THIS path) is a *hard guard* → exit non-zero.
// The two MUST NOT share an exit-code contract; only the token-refusal lives
// here.
func runNavigatorFixApply(projectRoot, draftID string) error {
	_, err := fix.Apply(fix.ApplyOptions{ProjectRoot: projectRoot, DraftID: draftID})
	if err == nil {
		// On a successful apply (including the idempotent no-op resume), emit a
		// stdout JSON signal mirroring the §A.4 handoff contract shape so the
		// orchestrator / a shell caller can consume the outcome.
		_, _ = fmt.Fprintln(os.Stdout, `{"status":"applied","draft_id":"`+draftID+`"}`)
		return nil
	}
	// Token-refusal path: print a message naming the missing/invalid
	// approval.json token to stderr (AC-NS5-008d) + return the error so cobra
	// exits non-zero.
	if errors.Is(err, fix.ErrApprovalTokenMissing) || errors.Is(err, fix.ErrApprovalTokenInvalid) {
		fmt.Fprintf(os.Stderr, "navigator-fix --apply %s: refused — %v\n", draftID, err)
	}
	return err
}
