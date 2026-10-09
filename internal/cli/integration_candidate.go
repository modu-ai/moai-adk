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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/gitenv"
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
	return candidateScrubbedGit
}

// candidateScrubbedGit runs one git command with the repo-scoping environment
// variables removed (card t1478 M2 repair, P1 — a data-destruction path):
// the git environment outranks cmd.Dir, so an inherited GIT_DIR /
// GIT_WORK_TREE had the verb build the candidate from ANOTHER repository's
// branch and force-push it to THAT repository's origin. Scope follows
// gitenv.RepoScopingVars exactly — identity and behavior vars stay, only
// repository LOCATION is removed. The failure semantics mirror
// factorylane.ExecGitRunner: a non-zero exit is a *factorylane.GitExitError
// carrying both streams (the conflict path reads exitErr.Stdout).
func candidateScrubbedGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitenv.Env()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		code := 1
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		}
		return stdout.String(), &factorylane.GitExitError{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}
	}
	return stdout.String(), nil
}

func (s *integrationCandidateSeams) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

func newIntegrationCandidateCmd() *cobra.Command {
	var cardFlag, runFlag string
	cmd := &cobra.Command{
		Use:   "candidate --card <id>",
		Short: "Build and push this card's pre-landing candidate commit to ci/<card> (the landing gate consumes its verdict)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(cardFlag) == "" {
				return fmt.Errorf("integration candidate: --card <id> is required (a candidate is a per-card artifact)")
			}
			root := integrationLockRoot()
			// The card worktree is the CALLER's tree: the verb is the
			// lane's act, run from the card worktree it stands in. The
			// guard (card t1478 M2 repair) verifies the caller's tree IS
			// the card's recorded worktree before anything is built — a
			// candidate constructed from a foreign tree would overwrite
			// ci/<card> in place.
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("integration candidate: resolve the card worktree: %w", err)
			}
			integBranch := configuredIntegrationBranch()
			if integBranch == "" {
				return fmt.Errorf("integration candidate: no integration branch is configured — resolve git_strategy's flow-scoped integration target first")
			}
			if gateOn := candidateCIEnabled(root); gateOn {
				callerTree, treeErr := candidateScrubbedGit(cwd, "rev-parse", "--show-toplevel")
				if treeErr != nil {
					return fmt.Errorf("integration candidate: resolve the caller's tree: %v", treeErr)
				}
				runID, runSelErr := candidateRunSelection(root, runFlag)
				if runSelErr != nil {
					return runSelErr
				}
				if guardErr := candidateCallerTreeGuard(root, cardFlag, strings.TrimSpace(callerTree), runID); guardErr != nil {
					return guardErr
				}
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
	cmd.Flags().StringVar(&runFlag, "run", "", "Factory run id (default: MOAI_KANBAN_ID, then the single active run) — read for the card record the tree guard verifies against")
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

	// The tip-equality refusal (card t1478 M2 repair): when the card branch
	// IS the integration tip, commit-tree dedupes the identical parents and
	// the "candidate" carries ONE parent — a shape AC-CCI-002-1's two-parent
	// witness rejects, pushed and recorded pending by the former code. The
	// card is already merged; refuse before any construction, push, or
	// record (the merge step's own cause-10 twin).
	if pinned == tip {
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: card %s's branch is at %s — the %s tip itself, nothing to candidate (the would-be commit would carry a single parent); advance the card branch or skip the candidate for an already-merged card", in.CardID, shortSHA(pinned), in.IntegrationBranch)
	}

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

	// The push + record section is ONE per-card critical section (card
	// t1478 M2 repair): two invocations of the same card interleaving as
	// A.push → B.push+record → A.record left the remote naming B's commit
	// while the record named A's. Inside the section the verdict decision
	// reads the record the previous section published — never a read taken
	// before the wait (the integration mutation lock's own placement rule).
	// The lock is the candidate store's scope over the shared state-lock
	// substrate, never the integration window (design.md D4).
	candidateBranch := "ci/" + in.CardID
	var rec factory.CandidateRecord
	if lockErr := factory.WithCandidateMutation(in.Root, in.CardID, func() error {
		// The identical-SHA re-candidate (card t1478 M2 repair): fixed git
		// dates (or any deterministic construction) reproduce the SAME
		// candidate commit, whose re-push does not move the remote ref and
		// fires no CI event — the only verdict it will ever have is the one
		// already recorded. Preserve verdict, run id, and observation time;
		// refresh only PushedAt. A DIFFERENT candidate SHA is a genuine
		// re-candidate: fresh pending (AC-CCI-003-1's supersede).
		verdict, runID, observedAt := factory.CandidateVerdictPending, "", ""
		if existing, readErr := factory.ReadCandidateRecord(in.Root, in.CardID, pinned); readErr == nil && existing.CandidateSHA == candidateSHA {
			verdict, runID, observedAt = existing.Verdict, existing.RunID, existing.ObservedAt
		}
		if _, err := git(in.CardWorktree, "push", "--force", "origin", candidateSHA+":refs/heads/"+candidateBranch); err != nil {
			// REQ-CCI-017: a push failure reports the stage reached and
			// leaves every recorded verdict untouched — the write below is
			// the only mutation, and it is not reached.
			return fmt.Errorf("push failed (stage: push to %s): %v — no candidate record was written and no existing verdict was touched", candidateBranch, err)
		}
		rec = factory.CandidateRecord{
			CardID:            in.CardID,
			PinnedSHA:         pinned,
			CandidateSHA:      candidateSHA,
			IntegrationBranch: in.IntegrationBranch,
			IntegrationTip:    tip,
			CandidateBranch:   candidateBranch,
			Verdict:           verdict,
			RunID:             runID,
			PushedAt:          now,
			ObservedAt:        observedAt,
		}
		if err := factory.WriteCandidateRecord(in.Root, rec); err != nil {
			return fmt.Errorf("record the candidate: %v", err)
		}
		return nil
	}); lockErr != nil {
		// %w: the busy sentinel must travel — a caller distinguishes
		// transient contention (retry-me) from a real refusal.
		return factory.CandidateRecord{}, fmt.Errorf("integration candidate: card %s: %w", in.CardID, lockErr)
	}
	return rec, nil
}

// candidateCallerTreeGuard refuses a caller standing outside the card's
// recorded worktree (card t1478 M2 repair): the verb builds the candidate
// from the CALLER's tree, so running it from another WT-* tree would
// force-overwrite ci/<card> with a foreign candidate. The factory card
// record's WorktreePath is what the merge step itself trusts for branch
// resolution, so it is what the guard verifies against. An unknown card
// refuses fail-closed: an unverifiable caller is exactly the shape the
// guard exists to stop. The comparison reads DIRECTORIES (factorySameTree),
// never strings — macOS presents /var/... and /private/var/... for one
// directory.
func candidateCallerTreeGuard(root, cardID, callerTree, runID string) error {
	ctx := context.Background()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("integration candidate: card %s: cannot verify the caller's tree: %w", cardID, err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(ctx, runID, cardID)
	if err != nil {
		return fmt.Errorf("integration candidate: card %s: no factory card record in run %s to verify the caller's tree against — run the card through the factory first, or run the candidate verb from the card's own worktree (%w)", cardID, runID, err)
	}
	if !factorySameTree(callerTree, card.WorktreePath) {
		return fmt.Errorf("integration candidate: refused — the caller's tree %s is not card %s's recorded worktree %s; a candidate built here would overwrite ci/%s with a foreign candidate (run the verb from the card's own worktree)", callerTree, cardID, card.WorktreePath, cardID)
	}
	return nil
}

// candidateRunSelection resolves the factory run the card record is read
// from (card t1478 M2 repair): the explicit --run flag first, then the
// launcher-selected run (MOAI_KANBAN_ID — the autoLaneRunResolveFn
// precedent, whose ResolveActiveRun validates the selection against the
// active-run table so a stale selection fails closed), then the
// single-active-run discovery.
func candidateRunSelection(root, explicit string) (string, error) {
	if run := strings.TrimSpace(explicit); run != "" {
		return resolveFactoryCardRun(context.Background(), root, run)
	}
	if run := strings.TrimSpace(os.Getenv(config.EnvFactoryRunID)); run != "" {
		return factorymsg.ResolveActiveRun(context.Background(), root, run)
	}
	return resolveFactoryCardRun(context.Background(), root, "")
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
