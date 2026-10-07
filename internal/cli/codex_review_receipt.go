package cli

// codex_review_receipt.go — the out-of-hook codex review runner for the Codex
// Stop chain's member 6 (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d, R1; design
// §D3.3 "Member 6 on the receipt method"). `moai verify codex-review` makes the
// same review call the Claude gate makes in-hook (codex_review_gate.go) and
// records the verdict as a receipt. On Codex the review cannot run inside the
// Stop handler: it may take up to config.DefaultCodexReviewGateTimeout, far
// above the handler timeout (REQ-HPR-018).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// codexReviewCheckID names the review gate's receipt entry.
const codexReviewCheckID = "codex-review"

// Receipt verdicts, mirroring the Claude gate's three outcomes: fail blocks,
// pass and inconclusive (including a review call that errored,
// codex_review_gate.go step 5) allow.
const (
	codexReviewVerdictPass         = "pass"
	codexReviewVerdictFail         = "fail"
	codexReviewVerdictInconclusive = "inconclusive"
)

// errCodexReviewerMissing: the codex binary is not on PATH, so there is no
// reviewer and no receipt is written.
var errCodexReviewerMissing = errors.New("codex binary not found on PATH: no reviewer, no receipt recorded")

// codexVersionProbe returns `codex --version` for the resolved binary. It is
// the receipt's tool_version on both sides — the runner that writes the
// receipt and the Stop chain that compares it. Tests pin it.
var codexVersionProbe = func(ctx context.Context, binaryPath string) (string, error) {
	out, err := codexRunner.run(ctx, binaryPath, []string{"--version"}, "")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// codexReviewReceiptState is the current value of the review receipt's bound
// fields (design §D3.6) for a TREE-scope key: HEAD and the working-tree digest
// from key, a config digest over what selects the review, the runner command,
// and the codex version. A probe failure leaves tool_version empty, which
// CheckReceipt reads as unbound — not run, never passed. The config digest is
// byte-identical to its pre-SPEC form: the tree-scope binding shape does not
// change (REQ-CGS-003).
func codexReviewReceiptState(ctx context.Context, key, binaryPath string) verify.ReceiptState {
	head, digest, _ := strings.Cut(key, ":")
	return reviewReceiptStateFrom(ctx, head, digest, map[string]string{
		"gate":   codexReviewCheckID,
		"method": codexMethodReviewStart,
		"target": codexTargetUncommitted,
	}, binaryPath)
}

// reviewReceiptStateFrom is the shared builder over the five bound fields.
func reviewReceiptStateFrom(ctx context.Context, head, digest string, cfg map[string]string, binaryPath string) verify.ReceiptState {
	version, _ := codexVersionProbe(ctx, binaryPath)
	return verify.ReceiptState{
		Head:         head,
		TreeDigest:   digest,
		ConfigDigest: verify.ConfigDigest(cfg),
		Command:      codexwiring.CodexReviewReceiptCommand,
		ToolVersion:  version,
	}
}

// codexReviewReceiptStateForScope builds the receipt state for a RESOLVED
// scope: the card class binds to the card diff (cardReviewReceiptState), the
// tree class keeps verify.Key over the scope's tree. Both execution paths —
// the producer below and the Codex Stop chain's member 6 — call THIS so the
// binding cannot drift between them (REQ-CGS-009).
func codexReviewReceiptStateForScope(ctx context.Context, scope reviewScope, binaryPath string) (verify.ReceiptState, error) {
	if scope.Class == reviewScopeCard {
		return cardReviewReceiptState(ctx, scope, binaryPath)
	}
	key, err := verify.Key(ctx, scope.Dir)
	if err != nil {
		return verify.ReceiptState{}, err
	}
	return codexReviewReceiptState(ctx, key, binaryPath), nil
}

// cardReviewReceiptState is the card-scope receipt binding (REQ-CGS-007):
// head = the card branch HEAD, TreeDigest = the card-scope digest over the
// recomputed merge base + the card diff + the non-runtime untracked files
// (cardScopeKeyParts), and the config digest carries the scope class and the
// card target so a card receipt can never satisfy a tree-scope state, nor a
// state measured before an absorption moved the base (AC-CGS-009/012).
func cardReviewReceiptState(ctx context.Context, scope reviewScope, binaryPath string) (verify.ReceiptState, error) {
	head, digest, err := cardScopeKeyParts(ctx, scope)
	if err != nil {
		return verify.ReceiptState{}, fmt.Errorf("codex review receipt: %w", err)
	}
	return reviewReceiptStateFrom(ctx, head, digest, map[string]string{
		"gate":   codexReviewCheckID,
		"method": codexMethodReviewStart,
		"target": codexTargetBaseBranch,
		"scope":  reviewScopeCard,
	}, binaryPath), nil
}

// @MX:ANCHOR: [AUTO] codex review receipt producer (R1) — the only writer of the receipt the Codex Stop chain's member 6 compares
// @MX:REASON: the Stop chain's member 6, `moai verify codex-review`, and the AC-HPR-002 goldens call it; a verdict mapping that differs from HandleCodexReviewGate breaks the parity the goldens assert

// produceCodexReviewReceipt runs the codex review for the session tree's
// RESOLVED scope (the same reviewScopeResolver the turn-end path uses,
// REQ-CGS-009) and records the verdict. The key is measured before the review,
// so a tree (or card diff) that changes while the review runs leaves a receipt
// the Stop chain reads as stale. It returns errCodexReviewerMissing, and
// records nothing, when the codex binary is absent — the Stop chain then
// allows on its own, as Claude does (codex_review_gate.go step 5).
//
// A TREE-scope fail whose every finding targets only the runtime-config
// surfaces is reclassified like the Claude gate's step 7-pre (REQ-CGSC-008,
// card-review repair R2 — both automatic paths take the one decision): the
// receipt records the gate's outcome (pass) and the reclassification row rides
// stderr (REQ-CGSC-011). Card scope keeps the fail.
func produceCodexReviewReceipt(ctx context.Context, root string) (verify.Receipt, error) {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return verify.Receipt{}, errCodexReviewerMissing
	}
	// One discriminator for both paths (plan §C [HARD]): root IS the session
	// tree here — the command runs inside the session's working tree, so the
	// scope resolves from it exactly as the Stop chain resolves from c.root.
	scope := reviewScopeResolver(root)
	state, err := codexReviewReceiptStateForScope(ctx, scope, binaryPath)
	if err != nil {
		return verify.Receipt{}, fmt.Errorf("codex review receipt: %w", err)
	}

	rctx, cancel := context.WithTimeout(ctx, config.DefaultCodexReviewGateTimeout)
	defer cancel()
	out, rpcErr := runCodexReviewRPC(rctx, binaryPath, codexMethodReviewStart, reviewRequestParams(scope))
	verdict, exit := codexReviewVerdictPass, 0
	drift := false
	switch {
	case rpcErr != nil:
		// Inconclusive records a non-zero exit (2): HasLocalPass keys on
		// ExitCode == 0, so the zero value would read the failed call as
		// local-pass evidence (the same class the gate's recorder fixes).
		verdict, exit = codexReviewVerdictInconclusive, 2
	case isBlockVerdict(out.Verdict):
		// REQ-CGSC-008 (card-review repair R2): BOTH automatic paths take the
		// one reclassification decision. On a TREE-scope review whose every
		// finding targets only the runtime-config surfaces the Claude gate
		// ALLOWs (codex_review_gate.go step 7-pre); the receipt mirrors that
		// outcome — the receipt verdict is the GATE's disposition of this tree
		// state, so a fail the gate reclassifies must not be recorded as one
		// the Codex Stop chain's member 6 would DENY on. The reclassification
		// row on the diagnostic channel (REQ-CGSC-011) keeps it from reading
		// as a silent reviewer pass. Card scope keeps the fail: the
		// reclassification is tree-only, as on Claude.
		if targets, ok := runtimeConfigOnlyFindings(out.Findings, scope.Dir); ok && scope.Class == reviewScopeTree {
			logRuntimeDriftReclassification(scope, targets)
			drift = true
		} else {
			verdict, exit = codexReviewVerdictFail, 1
		}
	case !strings.HasPrefix(strings.ToLower(strings.TrimSpace(out.Verdict)), codexReviewVerdictPass):
		// Same non-zero rule as the error arm above: a review that ran but
		// produced no verdict is not local-pass evidence.
		verdict, exit = codexReviewVerdictInconclusive, 2
	}
	r := verify.Receipt{
		CheckID:      codexReviewCheckID,
		Head:         state.Head,
		TreeDigest:   state.TreeDigest,
		ConfigDigest: state.ConfigDigest,
		Command:      state.Command,
		ToolVersion:  state.ToolVersion,
		ExitCode:     exit,
		Verdict:      verdict,
		RecordedAt:   time.Now(),
	}
	if err := verify.RecordReceipt(root, r); err != nil {
		return r, fmt.Errorf("codex review receipt: %w", err)
	}
	// SPEC-GATE-BOTTLENECK-001 REQ-GBN-002: the receipt store carries no free
	// text, so a fail's summary and findings ride the local detail file — the
	// delayed verdict (the next-turn entry hook) and the next Stop's cached
	// block must still say WHAT to fix. Mirrors the gate's former recorder;
	// a preservation failure loses only the detail text (fail-open).
	if verdict == codexReviewVerdictFail {
		codexReviewPreserveFindings(scope.Dir, state, out)
	}
	// The receipt store carries no free text, so the findings reach the
	// working agent here, on the runner's stderr.
	switch {
	case rpcErr != nil:
		_, _ = fmt.Fprintf(os.Stderr, "codex review: %s (review call failed: %v)\n", verdict, rpcErr)
	case drift:
		// logRuntimeDriftReclassification already wrote the REQ-CGSC-011 row;
		// this runner line keeps the recorded pass from reading as a silent
		// reviewer pass on the producer's own channel.
		_, _ = fmt.Fprintln(os.Stderr, "codex review: reviewer fail recorded as pass (runtime-managed drift only — see the reclassification row)")
	default:
		_, _ = fmt.Fprintf(os.Stderr, "codex review: %s: %s\n", verdict, strings.TrimSpace(out.Summary))
	}
	return r, nil
}

func newVerifyCodexReviewCmd(projectRoot *string) *cobra.Command {
	return &cobra.Command{
		Use:   "codex-review",
		Short: "Run the codex review of uncommitted changes and record a receipt (Codex Stop chain)",
		Long: `Run the same codex review the Claude codex review gate runs in-hook, for the
current tree's uncommitted changes, and record the verdict as a receipt. The
Codex Stop chain reads this receipt when the codex review gate is enabled and
codex is installed; it never runs the review itself. Run it after the last
edit of a turn, and again whenever a Stop continuation names it.

Prints the receipt as JSON with the review summary. The verdict is "pass",
"fail", or "inconclusive" (the review call failed). Fails, recording nothing,
when the codex binary is not on PATH.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			root, err := verifyResolveRoot(*projectRoot)
			if err != nil {
				return err
			}
			r, err := produceCodexReviewReceipt(verifyReceiptContext(c), root)
			if err != nil {
				return err
			}
			return verifyEmitJSON(c, map[string]any{
				"check_id":     r.CheckID,
				"verdict":      r.Verdict,
				"exit_code":    r.ExitCode,
				"head":         r.Head,
				"tree_digest":  r.TreeDigest,
				"command":      r.Command,
				"tool_version": r.ToolVersion,
			})
		},
	}
}
