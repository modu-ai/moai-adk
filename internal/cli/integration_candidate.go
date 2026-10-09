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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	var observeFlag bool
	cmd := &cobra.Command{
		Use:   "candidate --card <id> [--observe]",
		Short: "Build and push this card's pre-landing candidate commit to ci/<card> (the landing gate consumes its verdict); --observe records the CI run's verdict instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(cardFlag) == "" {
				return fmt.Errorf("integration candidate: --card <id> is required (a candidate is a per-card artifact)")
			}
			root := integrationLockRoot()
			// The observation path (REQ-CCI-010, design.md D3): an explicit
			// read of the CI runs from the lane/leader side — nothing writes
			// back from CI, and the record stays pending until this runs. It
			// mutates only the record store, so the caller-tree guard does
			// not apply here (nothing is built from the caller's tree).
			if observeFlag {
				if !candidateCIEnabled(root) {
					return fmt.Errorf("integration candidate: refused — workflow.candidate_ci.enabled is not true (set it in .moai/config/sections/workflow.yaml to enable the candidate path)")
				}
				rec, wrote, err := runCandidateObservation(root, cardFlag)
				if err != nil {
					return err
				}
				if wrote {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "integration candidate: card %s candidate %s observed run %s — verdict %s (observed %s)\n",
						cardFlag, shortSHA(rec.CandidateSHA), rec.RunID, rec.Verdict, rec.ObservedAt)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "integration candidate: card %s candidate %s — no binding run found (head SHA + ref must both match); record unchanged, verdict %s\n",
						cardFlag, shortSHA(rec.CandidateSHA), rec.Verdict)
				}
				return nil
			}
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
	cmd.Flags().BoolVar(&observeFlag, "observe", false, "Observe the CI runs for this card's latest candidate and record the first binding run's verdict (REQ-CCI-010) instead of pushing a new candidate")
	return cmd
}

// candidateGhRunsFn is the gh read seam: a package var so tests script the
// run states without network (the autoLaneRunResolveFn precedent).
var candidateGhRunsFn = ghRunStates

// candidateGhCommandFn and candidateGhRunsListFn are the gh command seams
// behind the required-check verdict (card t1478 M4 repair): the raw gh
// invocations are swappable so tests script the protection read and the
// job list without network.
var (
	candidateGhCommandFn  = candidateGhRunner
	candidateGhRunsListFn = candidateGhRunner
)

// candidateCIWorkflowFile names the workflow the candidate push runs —
// the run query is filtered to it so a same-SHA same-ref success from
// ANOTHER workflow can never satisfy the candidate verdict.
const candidateCIWorkflowFile = "ci.yml"

// candidateGuardBundleCheckName is the check name the guard-bundle job
// publishes (the `name:` in ci.yml) — what the candidate verdict reads
// when guard_bundle_required is true.
const candidateGuardBundleCheckName = "Guard Bundle"

// candidateGuardBundleRequired reads workflow.candidate_ci.guard_bundle_required
// (default true — D6 preserves today's gating where the guards ride the
// ordinary suite). An absent key and an unreadable config both read true:
// the bundle never silently loses its gate.
func candidateGuardBundleRequired(root string) bool {
	cfg, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil || cfg == nil {
		return true
	}
	return cfg.Workflow.CandidateCI.GuardBundleRequired
}

// runCandidateObservation reads the card's latest candidate record, walks
// its candidate-branch CI runs newest-first, and applies the first run
// that binds (REQ-CCI-010). No binding run leaves the record untouched —
// the caller reads wrote to say so.
func runCandidateObservation(root, cardID string) (factory.CandidateRecord, bool, error) {
	latest, err := factory.LatestCandidateRecord(root, cardID)
	if errors.Is(err, factory.ErrCandidateRecordAbsent) {
		return factory.CandidateRecord{}, false, fmt.Errorf("integration candidate: card %s has no candidate record — push a candidate first (moai integration candidate --card %s)", cardID, cardID)
	}
	if err != nil {
		return factory.CandidateRecord{}, false, fmt.Errorf("integration candidate: read card %s's candidate record: %w", cardID, err)
	}
	runs, err := candidateGhRunsFn(root, latest.CandidateBranch)
	if err != nil {
		return factory.CandidateRecord{}, false, fmt.Errorf("integration candidate: card %s: %w", cardID, err)
	}
	// The verdict set key is the branch pattern the candidate ACTUALLY ran
	// on — ci/**, where the observed run published its checks. The record's
	// IntegrationBranch is the MERGE target and names a different set
	// (main's contexts include checks a candidate push never publishes —
	// judging a candidate by them held every candidate red).
	return observeCandidateRuns(root, cardID, latest.PinnedSHA, "ci/**", runs, time.Now().UTC())
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
		// The push sequence (card t1478 M4 repair): allocated inside this
		// critical section, it orders same-second pushes where PushedAt's
		// second granularity ties — the latest-candidate scan must never
		// read filename order as recency.
		seq, seqErr := factory.NextCandidateSequence(in.Root, in.CardID)
		if seqErr != nil {
			return fmt.Errorf("allocate the push sequence: %v", seqErr)
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
			Seq:               seq,
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

// candidateAcquirePrecondition refuses the OWNING card's acquire while its
// latest candidate reads red (REQ-CCI-012, design.md D11): the refusal
// names card, verdict, and pinned SHA, and happens BEFORE any window-record
// mutation — so the record and the policy bytes are untouched by
// construction. Only RED holds: a card with no candidate, a pending one,
// or a green one acquires exactly as before. The hold clears the moment a
// re-candidate's record reads green — no separate state to reset.
func candidateAcquirePrecondition(root, cardID string) error {
	latest, err := factory.LatestCandidateRecord(root, cardID)
	if errors.Is(err, factory.ErrCandidateRecordAbsent) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("integration acquire: read card %s's candidate record: %w", cardID, err)
	}
	if latest.Verdict != factory.CandidateVerdictRed {
		return nil
	}
	return fmt.Errorf("integration acquire: refused — card %s's candidate %s is red (run %s, observed %s, pinned %s); fix and re-candidate with moai integration candidate --card %s (the hold is this card's record — other cards are unaffected)", cardID, shortSHA(latest.CandidateSHA), orUnsetStr(latest.RunID), orUnsetStr(latest.ObservedAt), shortSHA(latest.PinnedSHA), cardID)
}

func orUnsetStr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(unset)"
	}
	return s
}

// observeCandidateRuns is the explicit verdict observation (REQ-CCI-010,
// design.md D3): a gh read on the lane/leader side — nothing writes back
// from CI. Runs arrive newest-first; the FIRST run that binds the record
// (head SHA == candidate SHA, ref == candidate branch) is applied and the
// walk stops. A run that binds nothing leaves the record untouched — the
// caller reads what changed from the returned record.
func observeCandidateRuns(root, cardID, pinnedSHA, targetBranch string, runs []factory.CandidateRunState, now time.Time) (factory.CandidateRecord, bool, error) {
	for _, run := range runs {
		// The job-level verdict (card t1478 M4 repair): the run-level
		// conclusion conflates advisory jobs — a Race Test failure reds
		// the candidate while every required check is green. The verdict
		// reflects the REQUIRED checks for the record's TARGET BRANCH; a
		// read failure never mints green (the fallback keeps pending).
		if run.Status == "completed" {
			if adjusted := candidateRunVerdict(root, run.RunID, run.Status, run.Conclusion, targetBranch); adjusted.authoritative || adjusted.conclusion == "" {
				run.Conclusion = adjusted.conclusion
			}
		}
		rec, wrote, err := factory.ObserveCandidateVerdict(root, cardID, pinnedSHA, run, now)
		if err != nil {
			return factory.CandidateRecord{}, false, err
		}
		if wrote {
			return rec, true, nil
		}
	}
	rec, err := factory.ReadCandidateRecord(root, cardID, pinnedSHA)
	if err != nil {
		return factory.CandidateRecord{}, false, err
	}
	return *rec, false, nil
}

// candidateVerdictAdjustment is the required-check verdict for one run:
// the conclusion to record, and whether it came from the required-check
// read (authoritative) or is the run-level fallback.
type candidateVerdictAdjustment struct {
	conclusion    string
	authoritative bool
	why           string
}

// candidateRunVerdict judges one completed run against the FULL required
// check set for the candidate's TARGET BRANCH (card t1478 M4 repair): the
// set comes from the required-checks SSoT (.github/required-checks.yml —
// the file branch protection itself is rendered from) keyed by the
// branch pattern the candidate push ran on (ci/**) — main's set names
// checks a candidate push never publishes and held every candidate red.
//
// THE JUDGED SURFACE IS THE OBSERVED RUN'S OWN JOBS (card t1478 M4
// repair): a SHA can carry check runs from MULTIPLE runs (re-runs, a
// later candidate re-push of the same commit) — a cross-run rollup let a
// failed required check hide behind another run's success. The ci/** set
// is entirely ci.yml's own jobs, so the observed run's jobs are the exact
// surface the set publishes on; nothing outside the run can satisfy it.
// GREEN requires every required context present in THIS run's jobs AND
// successful; a context absent from the run names itself in `why`
// (fail-closed).
//
// Skipped companions do not fail their name (the matrix-skip pair): the
// `test`/`test-skip-marker` jobs live in the SAME run and publish the
// SAME context name with exactly one of them executed — a skipped
// instance counts as failure only when no successful execution of that
// name exists in the run.
//
// READ UNCERTAINTY IS NEVER GREEN: any gh read failure falls back to the
// run-level conclusion, but a fallback conclusion of success records
// NOTHING (empty conclusion — the observation leaves the record pending).
// A failed read cannot mint the green that fails open; a failure may
// still fall back to red (the safe direction).
//
// The SSoT absent or the branch unkeyed → the fallback set is this run's
// own job names (named in `why`), minus the Guard Bundle check when the
// key excludes it — the same admission policy as the SSoT set.
func candidateRunVerdict(root, runID, status, conclusion, targetBranch string) candidateVerdictAdjustment {
	// Never-green fallback: read uncertainty keeps the record pending
	// rather than minting green; a failure falls back to red (safe).
	fallback := candidateVerdictAdjustment{conclusion: conclusion, authoritative: false}
	if fallback.conclusion == "success" {
		fallback.conclusion = ""
	}
	bundleRequired := candidateGuardBundleRequired(root)

	// The required set: the SSoT keyed by the candidate-run branch pattern
	// first, the run's own jobs as the named fallback.
	required := []string{}
	source := "required-checks.yml[" + targetBranch + "]"
	setFound := false
	if checks, err := config.LoadRequiredChecks(root); err == nil {
		if branchSet, ok := checks.Branches[targetBranch]; ok {
			required = branchSet.Contexts
			setFound = true
		}
	}
	if !setFound {
		source = "run jobs (no required-checks.yml set for " + targetBranch + ")"
	}
	// The observed run's jobs — the judged surface.
	jobsRaw, err := candidateGhRunsListFn(root, "run", "view", runID, "--json", "jobs")
	if err != nil {
		return fallback
	}
	var jobsShape struct {
		Jobs []struct {
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal([]byte(jobsRaw), &jobsShape); err != nil {
		return fallback
	}
	if !setFound {
		// The fallback set: this run's own job names, under the same
		// Guard Bundle admission policy as the SSoT set.
		seenJob := map[string]bool{}
		for _, job := range jobsShape.Jobs {
			if job.Name == candidateGuardBundleCheckName && !bundleRequired {
				continue
			}
			if !seenJob[job.Name] {
				required = append(required, job.Name)
				seenJob[job.Name] = true
			}
		}
	}
	// The guard bundle's admission (SPEC-CANDIDATE-CI-001 M5,
	// workflow.candidate_ci.guard_bundle_required, default true): the key
	// decides whether the Guard Bundle check is required for the CANDIDATE
	// verdict — enforced here at the verdict collector, never by branch
	// protection (a required-check name is static there). Key false: a red
	// bundle stays visible in the run's checks but does not red the
	// candidate verdict.
	if bundleRequired {
		required = append(required, candidateGuardBundleCheckName)
	}
	if len(required) == 0 {
		return fallback
	}
	// Roll THIS run's jobs up per context name: a name's real verdict is
	// failure when any of its instances failed and NONE succeeded — a
	// skipped companion never fails a name a successful execution answers.
	contextState := map[string]string{} // name → "success" | "failure" | ""
	for _, job := range jobsShape.Jobs {
		current, seen := contextState[job.Name]
		switch {
		case job.Conclusion == "success":
			contextState[job.Name] = "success"
		case job.Conclusion == "" && !seen:
			// still running — not a verdict yet
			contextState[job.Name] = ""
		case job.Conclusion != "" && job.Conclusion != "success" && current != "success":
			contextState[job.Name] = "failure"
		}
	}
	missing := []string{}
	failed := []string{}
	for _, name := range required {
		state, published := contextState[name]
		switch {
		case !published:
			missing = append(missing, name)
		case state == "":
			// A required check still running: no verdict yet.
			return candidateVerdictAdjustment{conclusion: "", authoritative: true, why: source}
		case state == "failure":
			failed = append(failed, name)
		}
	}
	if len(failed) > 0 {
		return candidateVerdictAdjustment{conclusion: "failure", authoritative: true, why: source + " (failed: " + strings.Join(failed, ", ") + ")"}
	}
	if len(missing) > 0 {
		return candidateVerdictAdjustment{conclusion: "failure", authoritative: true, why: source + " (not in the observed run: " + strings.Join(missing, ", ") + ")"}
	}
	return candidateVerdictAdjustment{conclusion: "success", authoritative: true, why: source}
}

// ghRunStates reads the card's candidate-branch CI runs from gh, newest
// first. The child runs with the repo-scoping environment removed (gh
// resolves the repository through git — an inherited GIT_DIR would point
// the read at another repository) from the ROOT, whose origin names the
// repository the candidate was pushed to.
func ghRunStates(root, candidateBranch string) ([]factory.CandidateRunState, error) {
	out, err := candidateGhCommandFn(root, "run", "list",
		// The candidate verdict is the run of the CANDIDATE WORKFLOW —
		// without this filter, a same-SHA same-ref success from any other
		// workflow could satisfy the verdict (card t1478 M4 repair).
		"--workflow", candidateCIWorkflowFile,
		"--branch", candidateBranch, "--limit", "20",
		"--json", "databaseId,headSha,headBranch,status,conclusion")
	if err != nil {
		return nil, fmt.Errorf("gh run list for %s: %v", candidateBranch, err)
	}
	var rows []struct {
		DatabaseID int64  `json:"databaseId"`
		HeadSha    string `json:"headSha"`
		HeadBranch string `json:"headBranch"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
	}
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, fmt.Errorf("parse gh run list output for %s: %v", candidateBranch, err)
	}
	runs := make([]factory.CandidateRunState, 0, len(rows))
	for _, row := range rows {
		runs = append(runs, factory.CandidateRunState{
			RunID:      fmt.Sprintf("%d", row.DatabaseID),
			HeadSHA:    row.HeadSha,
			Ref:        row.HeadBranch,
			Status:     row.Status,
			Conclusion: row.Conclusion,
		})
	}
	return runs, nil
}

// candidateGhRunner runs gh with the repo-scoping environment removed (the same
// scrub the verb's git children take).
func candidateGhRunner(dir string, args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	cmd.Env = gitenv.Env()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
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
