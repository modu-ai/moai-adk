// factory_close_landed.go — `moai factory close-landed <card> --sha <merge>
// --run <run> [--ref <target>] [--reason <text>]`: the leader's one verb for
// closing a card whose delivering commit already sits on the landing target
// (af-flow: LOCAL main, reached by the audited batch). It performs, in order,
// the chain that .moai/docs/af-flow-landed-card-close.md lists by hand, by
// calling the commands behind those verbs rather than re-implementing them:
//
//  1. git merge-base --is-ancestor <sha> <ref>   (refuse otherwise)
//  2. moai todo landed <card> --sha <sha> --ref <ref>
//  3. moai factory rebind-evidence <card> --run <run> --reason <reason>
//  4. moai factory approve <card> --run <run>
//  5. moai todo done <card>
//
// A step that is already satisfied is skipped with a note; the first real
// refusal stops the chain and names the step that refused.
package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
)

// newFactoryCloseLandedCommand builds `moai factory close-landed`.
func newFactoryCloseLandedCommand() *cobra.Command {
	var sha, run, ref, reason string
	cmd := &cobra.Command{
		Use:          "close-landed <card> --sha <merge-sha> --run <run>",
		Short:        "Close a card whose delivering commit already landed on the target: landed, rebind-evidence, approve, done (leader path only)",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if factoryLaneRefusal() {
				return fmt.Errorf("factory close-landed: refused — %s: closing a landed card issues a receipt, which is the leader path's act (%s=%s marks a lane)",
					factoryLaneBoundarySentinel, config.EnvFactoryRole, config.FactoryRoleLane)
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return runFactoryCloseLanded(ctx, cmd, strings.TrimSpace(args[0]), strings.TrimSpace(sha),
				strings.TrimSpace(run), strings.TrimSpace(ref), strings.TrimSpace(reason))
		},
	}
	cmd.Flags().StringVar(&sha, "sha", "", "the delivering (merge) commit; must be reachable from --ref (required)")
	cmd.Flags().StringVar(&run, "run", "", "factory run id the receipt binds (required)")
	cmd.Flags().StringVar(&ref, "ref", "", "the landing target the commit must be on (default: the project's landed ref; local main under af-flow)")
	cmd.Flags().StringVar(&reason, "reason", "", "why the evidence is re-bound (default: a line naming the card, commit and ref)")
	return cmd
}

// closeLandedStep runs one command of the chain with its own output buffers and
// copies its stdout to out, so every step's line reaches the leader.
func closeLandedStep(ctx context.Context, out *bytes.Buffer, c *cobra.Command, args ...string) error {
	var stdout, stderr bytes.Buffer
	c.SetOut(&stdout)
	c.SetErr(&stderr)
	c.SetArgs(args)
	c.SilenceUsage = true
	c.SilenceErrors = true
	err := c.ExecuteContext(ctx)
	out.Write(stdout.Bytes())
	return err
}

func runFactoryCloseLanded(ctx context.Context, cmd *cobra.Command, cardID, sha, run, ref, reason string) error {
	if cardID == "" || sha == "" || run == "" {
		return fmt.Errorf("factory close-landed: a card id, --sha and --run are required")
	}
	if ref == "" {
		ref = todoLandedRef()
	}
	if reason == "" {
		reason = fmt.Sprintf("close-landed: %s delivered by %s on %s", cardID, sha, ref)
	}
	out := cmd.OutOrStdout()
	var buf bytes.Buffer
	flush := func() { _, _ = out.Write(buf.Bytes()) }
	refuse := func(step int, name string, err error) error {
		flush()
		return fmt.Errorf("factory close-landed: step %d (%s) refused: %w", step, name, err)
	}

	// (1) the commit must be on the target.
	onTarget, err := gitio.IsAncestor(factoryCardRoot(), sha, ref)
	if err != nil {
		return fmt.Errorf("factory close-landed: step 1 (ancestry) could not be answered for %s against %s: %w", sha, ref, err)
	}
	if !onTarget {
		return fmt.Errorf("factory close-landed: step 1 (ancestry) refused: %s is not an ancestor of %s — the card has not landed there", sha, ref)
	}
	_, _ = fmt.Fprintf(&buf, "close-landed: %s is on %s\n", sha, ref)

	rec, err := newTodoStore().LoadPure()
	if err != nil {
		return refuse(2, "todo landed", fmt.Errorf("the queue could not be read: %w", err))
	}
	live := false
	var landing struct{ sha, ref string }
	for i := range rec.Items {
		if rec.Items[i].ID == cardID {
			live = true
			if l := rec.Items[i].Landing; l != nil {
				landing.sha, landing.ref = strings.TrimSpace(l.SHA), strings.TrimSpace(l.Ref)
			}
		}
	}
	if !live {
		if rec.ArchivedIndex(cardID) >= 0 {
			_, _ = fmt.Fprintf(&buf, "close-landed: %s is already done; nothing to do\n", cardID)
			flush()
			return nil
		}
		return fmt.Errorf("factory close-landed: no backlog item %s", cardID)
	}

	// (2) the landing record.
	if landing.sha != "" && landing.ref == ref && (strings.HasPrefix(landing.sha, sha) || strings.HasPrefix(sha, landing.sha)) {
		_, _ = fmt.Fprintf(&buf, "close-landed: step 2 (todo landed) skipped — %s already records %s on %s\n", cardID, shortSHA(landing.sha), ref)
	} else if err := closeLandedStep(ctx, &buf, newTodoLandedCmd(), cardID, "--sha", sha, "--ref", ref); err != nil {
		return refuse(2, "todo landed", err)
	}

	// (3) the evidence re-bind (idempotent: an already-bound record is a note).
	if err := runFactoryRebindEvidence(withBuffer(cmd, &buf), cardID, reason, run); err != nil {
		return refuse(3, "rebind-evidence", err)
	}

	// (4) the receipt (re-issuance for the same card and run replaces).
	if err := closeLandedStep(ctx, &buf, newFactoryApproveCommand(), cardID, "--run", run); err != nil {
		return refuse(4, "approve", err)
	}

	// (5) the close.
	if err := closeLandedStep(ctx, &buf, newTodoDoneCmd(), cardID); err != nil {
		return refuse(5, "todo done", err)
	}
	flush()
	return nil
}

// withBuffer returns a throwaway command that shares cmd's context and writes
// to buf, so a step called as a function prints into the verb's ordered output.
func withBuffer(cmd *cobra.Command, buf *bytes.Buffer) *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(cmd.Context())
	c.SetOut(buf)
	c.SetErr(buf)
	return c
}
