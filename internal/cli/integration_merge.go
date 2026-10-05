// integration_merge.go — `moai integration merge --card <id>` (card t1479,
// REQ-MWQ-017): the one in-window merge verb. The verb resolves the
// session, the integration worktree, and the card-gate read, then hands
// everything to factory.RunMergeStep — the step owns the gate order, the
// collision check, the merge, and the release. A MergeStepError reaches
// the CLI root through the ExitCoder chain, so each cause's exit code is
// the process exit code without any verb-level plumbing.
package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

func newIntegrationMergeCmd() *cobra.Command {
	var cardFlag, runFlag string
	cmd := &cobra.Command{
		Use:   "merge --card <id>",
		Short: "Run the in-window merge step for a card (holder only; thirteen causes, one path)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(cardFlag) == "" {
				return fmt.Errorf("integration merge: --card <id> is required (the step gates on the card, so an unnamed merge merges nothing)")
			}
			sessionID := integrationSessionID("")
			if sessionID == "" {
				return fmt.Errorf("integration merge: cannot resolve this session's id; pass --session <id> (the holder decision needs an address)")
			}
			root := integrationLockRoot()
			integBranch := configuredIntegrationBranch()
			if integBranch == "" {
				integBranch = "develop"
			}
			integ := worktreeForBranch(integBranch)
			if integ == "" {
				return fmt.Errorf("integration merge: no worktree holds the integration branch %q — the leader provisions the integration worktree", integBranch)
			}
			lane, laneErr := factoryLaneLabelFromEnv("merge")
			if laneErr != nil {
				return laneErr
			}
			// The card-gate read (O4): the SAME predicate complete's step 1
			// runs, over the factory db — merge-ready, the caller's own
			// unexpired lease, version as read.
			seams := factory.MergeStepSeams{
				ReadCard: func(cardID string) (factory.MergeCardState, error) {
					return integrationReadMergeCard(cmd.Context(), root, cardID, lane)
				},
			}
			mergeSHA, err := factory.RunMergeStep(factory.MergeStepInput{
				Root:                root,
				IntegrationWorktree: integ,
				IntegrationBranch:   integBranch,
				CardID:              cardFlag,
				CallerSessionID:     sessionID,
			}, seams)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "integration merge: card %s merged as %s into %s (window released; the next queued ticket holds it)\n", cardFlag, mergeSHA[:12], integBranch)
			return nil
		},
	}
	cmd.Flags().StringVar(&cardFlag, "card", "", "The card id whose merge the step runs")
	cmd.Flags().StringVar(&runFlag, "run", "", "Factory run id (default: the single active run) — read for the card record")
	return cmd
}

// integrationReadMergeCard is the card-gate read O4 pins: one predicate
// shared by the merge verb's gate and factory complete's step 1. It reads
// the card's stage, the caller's lease, and the version AS READ — the
// callers never bump the version between the read and the gate.
func integrationReadMergeCard(ctx context.Context, root, cardID, lane string) (factory.MergeCardState, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	runID, err := resolveFactoryCardRun(ctx, root, "")
	if err != nil {
		return factory.MergeCardState{}, err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return factory.MergeCardState{}, err
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(ctx, runID, cardID)
	if err != nil {
		return factory.MergeCardState{}, err
	}
	leaseUnexpired := false
	if card.LeaseHolder == lane && card.LeaseExpiresAt != "" {
		if expires, parseErr := time.Parse(time.RFC3339, card.LeaseExpiresAt); parseErr == nil {
			leaseUnexpired = time.Now().UTC().Before(expires)
		}
	}
	return factory.MergeCardState{
		Stage:          card.Stage,
		LeaseUnexpired: leaseUnexpired,
		Version:        int(card.Version),
		WorktreePath:   card.WorktreePath,
	}, nil
}

// candidateCIEnabled reports whether the shared landing check is wired.
// SPEC-CANDIDATE-CI-001 (card t1478) owns the workflow.candidate_ci.enabled
// key and has not landed: absent reads FALSE, so the landing check is the
// absent no-op seam (spec.md §F), and whichever card lands second wires
// this read to the real key.
func candidateCIEnabled(root string) bool {
	_ = root // the key arrives with t1478; absent reads false today
	return false
}
