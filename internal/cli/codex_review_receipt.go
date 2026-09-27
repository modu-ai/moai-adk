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
// fields (design §D3.6): HEAD and the working-tree digest from key, a config
// digest over what selects the review, the runner command, and the codex
// version. A probe failure leaves tool_version empty, which CheckReceipt reads
// as unbound — not run, never passed.
func codexReviewReceiptState(ctx context.Context, key, binaryPath string) verify.ReceiptState {
	head, digest, _ := strings.Cut(key, ":")
	version, _ := codexVersionProbe(ctx, binaryPath)
	return verify.ReceiptState{
		Head:       head,
		TreeDigest: digest,
		ConfigDigest: verify.ConfigDigest(map[string]string{
			"gate":   codexReviewCheckID,
			"method": codexMethodReviewStart,
			"target": codexTargetUncommitted,
		}),
		Command:     codexwiring.CodexReviewReceiptCommand,
		ToolVersion: version,
	}
}

// @MX:ANCHOR: [AUTO] codex review receipt producer (R1) — the only writer of the receipt the Codex Stop chain's member 6 compares
// @MX:REASON: the Stop chain's member 6, `moai verify codex-review`, and the AC-HPR-002 goldens call it; a verdict mapping that differs from HandleCodexReviewGate breaks the parity the goldens assert

// produceCodexReviewReceipt runs the codex review for the current tree and
// records the verdict. The key is measured before the review, so a tree that
// changes while the review runs leaves a receipt the Stop chain reads as
// stale. It returns errCodexReviewerMissing, and records nothing, when the
// codex binary is absent — the Stop chain then allows on its own, as Claude
// does (codex_review_gate.go step 4).
func produceCodexReviewReceipt(ctx context.Context, root string) (verify.Receipt, error) {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return verify.Receipt{}, errCodexReviewerMissing
	}
	key, err := verify.Key(ctx, root)
	if err != nil {
		return verify.Receipt{}, fmt.Errorf("codex review receipt: %w", err)
	}
	state := codexReviewReceiptState(ctx, key, binaryPath)

	rctx, cancel := context.WithTimeout(ctx, config.DefaultCodexReviewGateTimeout)
	defer cancel()
	out, rpcErr := runCodexReviewRPC(rctx, binaryPath, codexMethodReviewStart, map[string]any{
		"target": codexTargetUncommitted,
		"cwd":    root,
	})
	verdict, exit := codexReviewVerdictPass, 0
	switch {
	case rpcErr != nil:
		verdict = codexReviewVerdictInconclusive
	case isBlockVerdict(out.Verdict):
		verdict, exit = codexReviewVerdictFail, 1
	case !strings.HasPrefix(strings.ToLower(strings.TrimSpace(out.Verdict)), codexReviewVerdictPass):
		verdict = codexReviewVerdictInconclusive
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
	// The receipt store carries no free text, so the findings reach the
	// working agent here, on the runner's stderr.
	if rpcErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex review: %s (review call failed: %v)\n", verdict, rpcErr)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "codex review: %s: %s\n", verdict, strings.TrimSpace(out.Summary))
	}
	return r, nil
}

func init() {
	verifyExtraCommands = append(verifyExtraCommands, newVerifyCodexReviewCmd)
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
