// integration_candidate.go — `moai integration candidate --card <id>`
// (card t1478, SPEC-CANDIDATE-CI-001 REQ-CCI-001..003, 017): build the
// exact would-be merge commit with `git merge-tree --write-tree` +
// `git commit-tree`, push it to `ci/<card>`, and record it keyed
// (card, pinned SHA) with a `pending` verdict. The candidate verb is
// window-free (design.md D4): it acquires nothing, mutates no window
// record, and merges nothing — the CI run on `ci/<card>` is the verdict
// surface, and the merge step's landing check (REQ-CCI-011) consumes it.
//
// The integration target is the CONFIG-RESOLVED branch (design.md D5 —
// the resolver, never a literal). The record carries the branch and tip it
// was built against; the landing check binds them to the merge's ACTUAL
// destination (design.md D10), so a window retargeted by `--branch` after
// the candidate was built refuses at merge time instead of merging an
// unverified tree.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// integrationCandidateInput names the verb's operands. Root is the PRIMARY
// checkout (config + candidate record store, the same visibility contract
// the integration lock has); CardWorktree is the tree whose checked-out
// WT- branch is the card's work — the lane runs the verb from its own card
// worktree, and the REQ-CCI-004 resolver validates it.
type integrationCandidateInput struct {
	Root              string
	CardID            string
	CardWorktree      string
	IntegrationBranch string
}

// integrationCandidateSeams carries the verb's test seams. nil Git is the
// real git runner; nil Now is the wall clock.
type integrationCandidateSeams struct {
	Git func(dir string, args ...string) (string, error)
	Now func() time.Time
}

func (s *integrationCandidateSeams) gitRunner() func(dir string, args ...string) (string, error) {
	if s.Git != nil {
		return s.Git
	}
	return func(dir string, args ...string) (string, error) {
		runner := factorylane.ExecGitRunner{Dir: dir}
		return runner.Git(args...)
	}
}

func (s *integrationCandidateSeams) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

func newIntegrationCandidateCmd() *cobra.Command {
	var cardFlag string
	cmd := &cobra.Command{
		Use:   "candidate --card <id>",
		Short: "Build and push this card's pre-landing candidate commit to ci/<card> (the landing gate consumes its verdict)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(cardFlag) == "" {
				return fmt.Errorf("integration candidate: --card <id> is required (a candidate is a per-card artifact)")
			}
			root := integrationLockRoot()
			// The card worktree is the CALLER's tree: the verb is the
			// lane's act, run from the card worktree it stands in.
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("integration candidate: resolve the card worktree: %w", err)
			}
			integBranch := configuredIntegrationBranch()
			if integBranch == "" {
				return fmt.Errorf("integration candidate: no integration branch is configured — resolve git_strategy's flow-scoped integration target first")
			}
			rec, err := runIntegrationCandidate(integrationCandidateInput{
				Root:              root,
				CardID:            cardFlag,
				CardWorktree:      cwd,
				IntegrationBranch: integBranch,
			}, integrationCandidateSeams{})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "integration candidate: card %s candidate %s pushed to %s (built from %s at %s; the CI run there is the verdict surface)\n",
				rec.CardID, rec.CandidateSHA[:12], rec.CandidateBranch, rec.PinnedSHA[:12], rec.IntegrationBranch)
			return nil
		},
	}
	cmd.Flags().StringVar(&cardFlag, "card", "", "The card id the candidate is built for")
	return cmd
}

// runIntegrationCandidate builds, pushes, and records one candidate
// (REQ-CCI-002/003/005). The order is deliberate: card-id validation and
// the config gate refuse BEFORE any git — and therefore any remote — call
// (REQ-CCI-017/006); the record is written only AFTER the push succeeded,
// so a failed push never reads as a pending-or-green candidate and never
// touches a previously recorded verdict (REQ-CCI-017).
func runIntegrationCandidate(in integrationCandidateInput, seams integrationCandidateSeams) (factory.CandidateRecord, error) {
	// Card-id validation precedes everything: the id becomes a remote
	// branch segment and a record path segment, so an invalid shape must
	// not reach either (REQ-CCI-017).
	if !homestate.ValidCardID(in.CardID) {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: %q is not a valid card id — refusing before any git or remote call", in.CardID)
	}
	// The capability gate (REQ-CCI-006): an absent or false key refuses
	// naming it, before any git call.
	if !candidateCIEnabled(in.Root) {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: refused — workflow.candidate_ci.enabled is not true (set it in .moai/config/sections/workflow.yaml to enable the candidate path)")
	}
	git := seams.gitRunner()
	now := seams.now().Format(time.RFC3339)

	// REQ-CCI-004: the card's worktree resolves to its WT- branch via the
	// SAME resolver the merge step uses, and ONE SHA is pinned from the
	// branch ref (never HEAD — a detached tree cannot sneak in).
	cardBranch, err := factory.ResolveCardBranch(in.CardWorktree, in.CardID)
	if err != nil {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: %v", err)
	}
	pinned, err := git(in.CardWorktree, "rev-parse", "refs/heads/"+cardBranch)
	if err != nil {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: pin %s: %v", cardBranch, err)
	}
	pinned = strings.TrimSpace(pinned)
	tip, err := git(in.CardWorktree, "rev-parse", "refs/heads/"+in.IntegrationBranch)
	if err != nil {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: read the %s tip: %v", in.IntegrationBranch, err)
	}
	tip = strings.TrimSpace(tip)

	// The two-parent construction (REQ-CCI-002, design.md D1):
	// merge-tree --write-tree produces the would-be merge tree; a conflict
	// is git's exit 1 — a RESULT the verb reads, refusing with the card
	// and the conflicted paths named, writing no branch and no record.
	tree, err := git(in.CardWorktree, "merge-tree", "--write-tree", tip, pinned)
	if err != nil {
		var exitErr *factorylane.GitExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode == 1 {
			conflicted := strings.TrimSpace(exitErr.Stdout)
			return factory.CandidateRecord{}, fmt.Errorf("integration candidate: card %s conflicts with %s — resolve the conflict in the card tree and re-candidate (conflict resolution remains the owning lane's duty):\n%s", in.CardID, in.IntegrationBranch, conflicted)
		}
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: merge-tree could not run: %v", err)
	}
	tree = strings.TrimSpace(tree)
	tree = gitFirstLine(tree)
	candidateMsg := fmt.Sprintf("Candidate: merge %s (%s) into %s for card %s (pre-landing candidate; disposable ci/ branch)", cardBranch, shortSHA(pinned), in.IntegrationBranch, in.CardID)
	candidateSHA, err := git(in.CardWorktree, "commit-tree", tree, "-p", tip, "-p", pinned, "-m", candidateMsg)
	if err != nil {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: commit-tree: %v", err)
	}
	candidateSHA = strings.TrimSpace(candidateSHA)

	// The push (REQ-CCI-003): replace-in-place on the disposable ci/
	// branch. The --force scope is the guard, not a habit: the refspec is
	// exactly ci/<validated-card-id>, and candidate branches are never
	// protected refs. A push failure reports the stage reached and leaves
	// every recorded verdict untouched (REQ-CCI-017) — the record below
	// is written only after this succeeds.
	candidateBranch := "ci/" + in.CardID
	if _, err := git(in.CardWorktree, "push", "--force", "origin", candidateSHA+":refs/heads/"+candidateBranch); err != nil {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: push failed (stage: push to %s): %v — no candidate record was written and no existing verdict was touched", candidateBranch, err)
	}

	rec := factory.CandidateRecord{
		CardID:            in.CardID,
		PinnedSHA:         pinned,
		CandidateSHA:      candidateSHA,
		IntegrationBranch: in.IntegrationBranch,
		IntegrationTip:    tip,
		CandidateBranch:   candidateBranch,
		Verdict:           factory.CandidateVerdictPending,
		PushedAt:          now,
	}
	if err := factory.WriteCandidateRecord(in.Root, rec); err != nil {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: record the candidate: %v", err)
	}
	return rec, nil
}

// firstLine returns the first line of a git output blob — merge-tree
// prints the result tree OID first, whatever follows.
func gitFirstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

// shortSHA renders the 12-char form the verb family's messages use.
func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
