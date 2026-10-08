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
	var cardFlag, runFlag, sessionFlag string
	cmd := &cobra.Command{
		Use:   "merge --card <id> [--session <id>]",
		Short: "Run the in-window merge step for a card (holder only; thirteen causes, one path)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// REQ-SD-025: the Codex merge edge is refused on EVERY path that
			// reaches a merge — the verb is one more path, not an exception,
			// and the refusal precedes every other check (card-review r1
			// P1-3): a Codex lane stops at merge-ready whatever else is
			// wrong with the invocation.
			if err := factoryRefuseCodexMergeEdge("merge"); err != nil {
				return err
			}
			if strings.TrimSpace(cardFlag) == "" {
				return fmt.Errorf("integration merge: --card <id> is required (the step gates on the card, so an unnamed merge merges nothing)")
			}
			// t1576 review round 3: the flag the holder-refusal message
			// advises is now registered — a session whose environment cannot
			// carry the id passes it explicitly.
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" {
				return fmt.Errorf("integration merge: cannot resolve this session's id; pass --session <id> (the holder decision needs an address)")
			}
			root := integrationLockRoot()
			initWindowLeaseOverride(root)
			// P1-5 (card-review r1): the merge target is the branch the
			// WINDOW RECORD names — the record is what the lane acquired —
			// and the config value is only the fallback for a record that
			// predates targets. Reading config first let a window taken with
			// --branch <other> merge into the wrong branch.
			integBranch := ""
			if lock, lockErr := factory.ReadIntegrationLock(root); lockErr == nil && lock.Held() && lock.Branch != "" {
				integBranch = lock.Branch
			}
			if integBranch == "" {
				integBranch = configuredIntegrationBranch()
			}
			if integBranch == "" {
				integBranch = "develop"
			}
			// The integration worktree resolves at the RECORD's repository —
			// the lock root (the primary checkout), never the process cwd: a
			// verb invoked from a card worktree would otherwise read that
			// worktree's own repository and find a different develop.
			integ, integErr := integrationMergeWorktree(root, integBranch)
			if integErr != nil {
				return integErr
			}
			lane, laneErr := factoryLaneLabelFromEnv("merge")
			if laneErr != nil {
				return laneErr
			}
			// The card-gate read (O4): the SAME predicate complete's step 1
			// runs, over the factory db — merge-ready, the caller's own
			// unexpired lease, version as read.
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			runID, runErr := resolveFactoryCardRun(ctx, root, runFlag)
			if runErr != nil {
				return runErr
			}
			seams := factory.MergeStepSeams{
				ReadCard: func(cardID string) (factory.MergeCardState, error) {
					return integrationReadMergeCardForRun(ctx, root, runID, cardID, lane)
				},
			}
			// The landing check rides the candidate-CI key (spec.md §F):
			// absent/false is the absent no-op seam. While t1478 is
			// unlanded the key cannot read true — but if a future
			// configuration flips it early, the step refuses LOUDLY rather
			// than silently skipping a gate the project asked for.
			if candidateCIEnabled(root) {
				seams.LandingCheck = func(string, string) error {
					return fmt.Errorf("the shared landing check is not wired until SPEC-CANDIDATE-CI-001 lands")
				}
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
	cmd.Flags().StringVar(&sessionFlag, "session", "", "This session's id when the environment cannot carry it (the holder decision needs an address)")
	return cmd
}

// integrationMergeWorktree resolves the worktree the merge step merges INTO,
// and refuses the PRIMARY checkout (card t1479 r3 F3): the resolver below
// returns ANY tree holding the branch — the primary included, with no
// primary exclusion on the path — and a merge whose target is the shared
// primary checkout is exactly what the branch-guard doctrine forbids (the
// primary never changes branch for anyone). complete's own resolution
// refuses the same shape; the verb owns the identical test, by the same
// identity helper factorySameTree.
func integrationMergeWorktree(root, integBranch string) (string, error) {
	integ := factoryWorktreeForBranchIn(root, integBranch)
	if integ == "" {
		return "", fmt.Errorf("integration merge: no worktree holds the integration branch %q — the leader provisions the integration worktree", integBranch)
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return "", fmt.Errorf("integration merge: cannot identify the primary checkout of %s: %w", root, err)
	}
	if factorySameTree(integ, primary) {
		return "", fmt.Errorf("integration merge: refused — the only tree holding %q is the primary checkout %s (the primary never changes branch; the leader provisions the integration worktree)", integBranch, primary)
	}
	return integ, nil
}

// integrationReadMergeCardForRun is the card-gate read O4 pins: one
// predicate shared by the merge verb's gate and factory complete's step 1.
// It reads the card's stage, the caller's lease, and the version AS READ —
// the callers never bump the version between the read and the gate.
func integrationReadMergeCardForRun(ctx context.Context, root, runID, cardID, lane string) (factory.MergeCardState, error) {
	if ctx == nil {
		ctx = context.Background()
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
		State:          card.State,
		LeaseUnexpired: leaseUnexpired,
		Version:        int(card.Version),
		WorktreePath:   card.WorktreePath,
	}, nil
}

// candidateCIEnabled reports whether the shared landing check is wired.
// SPEC-CANDIDATE-CI-001 (card t1478) owns the workflow.candidate_ci.enabled
// key and has not landed: absent reads FALSE, so the landing check is the
// absent no-op seam (spec.md §F), and whichever card lands second wires
// this read to the real key and assigns the LandingCheck seam its real
// implementation.
//
// @MX:DEBT: constant-false placeholder for the landing-check gate
// @MX:CEILING: only until SPEC-CANDIDATE-CI-001 lands its key
// @MX:UPGRADE: t1478's landing — replace with the real config read and
// wire factory.MergeStepSeams.LandingCheck
func candidateCIEnabled(root string) bool {
	_ = root // the key arrives with t1478; absent reads false today
	return false
}
