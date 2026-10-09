package cli

// factory_merge.go — the lane-direct merge surface of
// SPEC-FACTORY-LANE-AUTONOMY-001 fragment 3 (design.md D3): the check
// sequence a lane executes before entering the integration window, and the
// window gate the merge act itself sits behind. It consumes the EXISTING
// `moai integration` window (factory.AcquireIntegrationLock and the recorded
// hold) — no new serialization mechanism (REQ-FLA-010/011) and no F3
// controller machinery (spec.md §F exclusion): no write-ahead start events,
// no trial merges, no tick loop. This file never performs a merge.
//
// Every refusal here is a verdict, not a failure: the command exits 0 on
// refused / waiting exactly like the probe's channel-unavailable verdict,
// because what a lane does with the answer is the lane's doctrine move.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// factoryMergeVerdict is the machine-readable shape of one merge-readiness
// outcome. Verdict is cleared | refused | waiting.
type factoryMergeVerdict struct {
	Verdict         string                   `json:"verdict"`
	Lane            string                   `json:"lane"`
	Card            string                   `json:"card"`
	FailedCondition string                   `json:"failed_condition,omitempty"`
	Holder          string                   `json:"holder,omitempty"`
	Detail          string                   `json:"detail,omitempty"`
	Checks          []factorylane.MergeCheck `json:"checks,omitempty"`
}

func newFactoryMergeCommand() *cobra.Command {
	merge := &cobra.Command{
		Use:   "merge",
		Short: "Lane-direct merge readiness: the condition triple and the window gate (AC-FLA-009/010/011)",
	}
	merge.AddCommand(newFactoryMergeReadyCommand(), newFactoryMergeGateCommand())
	return merge
}

// newFactoryMergeReadyCommand is the full lane-direct sequence: evaluate the
// t1241 condition triple, record it, take the integration window through the
// existing acquire behavior — refusing with the failing condition named
// BEFORE any window is taken, and with the holder named when the window is
// already held (AC-FLA-009/010).
//
// @MX:ANCHOR: [AUTO] the lane-direct merge entry point — triple, record, window in one sequence
// @MX:REASON: every self-served integration in the factory funnels through it; a wrong verdict here either blocks a ready merge or clears an unready one past the recorded serialization point.
// @MX:SPEC: SPEC-FACTORY-LANE-AUTONOMY-001
func newFactoryMergeReadyCommand() *cobra.Command {
	var card, specID, branch, developRef, mergeCommit, sessionFlag string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "ready",
		Short: "Run the condition triple, record it, and take the integration window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("merge ready")
			if err != nil {
				return err
			}
			if strings.TrimSpace(card) == "" {
				return fmt.Errorf("factory merge ready: --card is required")
			}
			if strings.TrimSpace(specID) == "" {
				return fmt.Errorf("factory merge ready: --spec is required (the card's SPEC id — its sync phase record is condition (a))")
			}
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if strings.TrimSpace(branch) == "" {
				branch = currentBranch()
			}
			if strings.TrimSpace(developRef) == "" {
				// The target is resolved exactly as the integration window
				// resolves it (the configured integration target), but
				// with NO caller fallback: a merge-readiness check must not
				// guess its target — the caller's own card branch is never a
				// merge target.
				targetRoot := integrationLockRoot()
				targetCfg := config.LoadGitFlowIntegrationConfig(targetRoot)
				developRef = targetCfg.IntegrationTarget
				if developRef == "" {
					return fmt.Errorf("factory merge ready: no integration branch configured — the merge target would be a guess: %s", targetCfg.EmptyTargetGuidance(targetRoot, "pass --develop <branch>"))
				}
			}

			run, err := factorylane.EvaluateMergeTriple(factorylane.MergeTripleInput{
				Lane:        lane,
				Card:        card,
				SpecDir:     filepath.Join(root, ".moai", "specs", specID),
				Branch:      branch,
				Develop:     developRef,
				RepoDir:     root,
				MergeCommit: mergeCommit,
				// REQ-MWQ-021's fourth condition: the re-measure record keyed
				// to the candidate tree, with the recorded command printed
				// verbatim in the condition's detail.
				VerifyRemeasure: func(treeSHA string) (string, error) {
					rec, err := factory.ReadRemeasureRecord(integrationLockRoot(), treeSHA)
					if err != nil {
						return "", err
					}
					if err := factory.ValidateRemeasureRecord(rec); err != nil {
						return "", err
					}
					return rec.Command, nil
				},
			}, factorylane.ExecGitRunner{Dir: root})
			if err != nil {
				return err
			}
			recorded, err := factorylane.NewStore(root, nil).RecordMergeCheckRun(run)
			if err != nil {
				return err
			}

			// Human lines ride stderr on the JSON path so a --json
			// consumer's stdout stays one document (the release-verb
			// precedent).
			human := cmd.OutOrStdout()
			if asJSON {
				human = cmd.ErrOrStderr()
			}

			if !recorded.AllPassed {
				// Refused BEFORE the window: the failing condition is named,
				// the run is recorded, and no acquire happens.
				_, _ = fmt.Fprintf(human, "merge-readiness: REFUSED — failing condition: %s\nno window was taken\n", recorded.FailedCondition)
				printMergeChecks(human, recorded.Checks)
				refusal := factoryMergeVerdict{
					Verdict: "refused", Lane: lane, Card: card,
					FailedCondition: recorded.FailedCondition,
					Detail:          "the condition triple failed — no window was taken",
					Checks:          recorded.Checks,
				}
				if err := emitFactoryMergeVerdict(cmd, asJSON, refusal); err != nil {
					return err
				}
				// REQ-MWQ2-006 (card t1582 item ③): a measurement failure — the
				// re-measure record absent or invalid — must not read as a pass
				// to a calling script. The verdict above is printed either way;
				// the exit code carries the failure, under cause 1 (record
				// invalid), the code the merge verb gives the same record state.
				// The window-contest refusal (waiting) and the other refused
				// conditions keep the zero-exit verdict (REQ-MWQ2-010).
				if recorded.FailedCondition == factorylane.CheckRemeasureRecord {
					return &exitCodeError{code: factory.MergeExitRecordInvalid, msg: "factory merge ready: the measurement failed — the re-measure record for the candidate tree is absent or invalid; run moai integration remeasure, then merge ready again"}
				}
				return nil
			}

			// github-flow delivers by pull request: the integration window is
			// no prerequisite (REQ-GFD-006), so a cleared triple takes none.
			if factoryGitHubFlow(integrationLockRoot()) {
				const detail = "github-flow takes no integration window — run moai factory complete to push the card branch and open the pull request"
				_, _ = fmt.Fprintf(human, "merge-readiness: CLEARED — three checks recorded; %s\n", detail)
				return emitFactoryMergeVerdict(cmd, asJSON, factoryMergeVerdict{
					Verdict: "cleared", Lane: lane, Card: card, Detail: detail,
				})
			}
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" {
				return fmt.Errorf("factory merge ready: cannot resolve this session's id; pass --session <id> (a window with an invented holder can be neither released by its holder nor recognized by the guard)")
			}
			lockRoot := integrationLockRoot()
			ownerPID, _ := session.ResolveOwnerPID()
			// AcquireIntegrationLock's first return is the lock it REPLACED,
			// not the one it wrote — the acquired record is read back from
			// the store, which is also what carries the acquire stamp the
			// pre-check proof below is verified against.
			if _, err := factory.AcquireIntegrationLock(lockRoot, factory.IntegrationLock{
				SessionID:    sessionID,
				SessionName:  lane,
				PID:          ownerPID,
				PIDSource:    factory.PIDSourceSessionOwner,
				Branch:       developRef,
				BranchSource: factory.BranchSourceConfig,
				Worktree:     worktreeForBranch(developRef),
				Card:         card,
			}, false); err != nil {
				if !factory.IsIntegrationLockHeld(err) {
					return err
				}
				// AC-FLA-010: refused/waiting WITH THE HOLDER NAMED — the
				// existing acquire behavior consumed, surfaced in this path's
				// output. The window stays with its holder.
				current, readErr := factory.ReadIntegrationLock(lockRoot)
				if readErr != nil {
					return readErr
				}
				holder := fmt.Sprintf("%s (session %s, pid %d)", orDash(current.SessionName), current.SessionID, current.PID)
				if current.Stale() {
					holder += " — the holder's session is gone (reclaimable)"
				}
				_, _ = fmt.Fprintf(human, "merge-readiness: WAITING — the integration window is held by %s\nthis lane's checks are recorded; re-run ready once the window frees (release by the holder, or a recorded takeover)\n", holder)
				return emitFactoryMergeVerdict(cmd, asJSON, factoryMergeVerdict{
					Verdict: "waiting", Lane: lane, Card: card, Holder: holder,
					Detail: "the integration window is held by another lane (REQ-FLA-010)",
					Checks: recorded.Checks,
				})
			}

			// AC-FLA-009 second half: the recorded check output for all three
			// exists BEFORE the integration acquire timestamp. The window is
			// already ours here, so a failed proof is released again — a
			// cleared verdict never rides an unproven record.
			acquired, err := factory.ReadIntegrationLock(lockRoot)
			if err != nil {
				return err
			}
			if !acquired.Held() || acquired.SessionID != sessionID {
				return fmt.Errorf("factory merge ready: the window record after acquire does not name this session (%s) — refusing on the read-back", sessionID)
			}
			acquireAt, parseErr := time.Parse(time.RFC3339, acquired.AcquiredAt)
			if parseErr != nil {
				return fmt.Errorf("factory merge ready: the recorded acquired_at %q is unreadable: %w", acquired.AcquiredAt, parseErr)
			}
			proofOK, proofWhy := factorylane.VerifyRunBeforeAcquire(&recorded, acquireAt)
			// One case the strict record-only verifier cannot see: the
			// window's acquire stamp carries RFC3339 SECOND precision (the
			// factory record format) while the check record carries
			// nanoseconds, so a record written within the acquire's own
			// second is unorderable from the records alone. sameSecondProof
			// admits exactly that: a passing record falling at or before the
			// end of the acquire's second, written by THIS call's immediately
			// preceding statement. A failed run, a missing run, or a record
			// later than the acquire's second all still refuse, and the
			// window is released again.
			sameSecondProof := recorded.AllPassed && recorded.CheckedAt.Before(acquireAt.Add(time.Second))
			if !proofOK && !sameSecondProof {
				if _, releaseErr := factory.ReleaseIntegrationLock(lockRoot, sessionID, ownerPID, false); releaseErr != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "factory merge ready: the pre-acquire proof failed (%s) and the window release also failed: %v\n", proofWhy, releaseErr)
				}
				_, _ = fmt.Fprintf(human, "merge-readiness: REFUSED — %s\nthe window was released again; no cleared verdict rides an unproven record\n", proofWhy)
				return emitFactoryMergeVerdict(cmd, asJSON, factoryMergeVerdict{
					Verdict: "refused", Lane: lane, Card: card, Detail: proofWhy,
				})
			}

			_, _ = fmt.Fprintf(human, "merge-readiness: CLEARED — three checks recorded before the window acquire\n%s\nthe window is held by this lane; merge in the integration worktree, then release\n", factorylane.VerifyWhyText)
			return emitFactoryMergeVerdict(cmd, asJSON, factoryMergeVerdict{
				Verdict: "cleared", Lane: lane, Card: card, Detail: factorylane.VerifyWhyText,
			})
		},
	}
	cmd.Flags().StringVar(&card, "card", "", "The card being merged")
	_ = cmd.MarkFlagRequired("card")
	cmd.Flags().StringVar(&specID, "spec", "", "The card's SPEC id — its sync phase record is condition (a)")
	_ = cmd.MarkFlagRequired("spec")
	cmd.Flags().StringVar(&branch, "branch", "", "The card branch (merge source; default: the current branch)")
	cmd.Flags().StringVar(&developRef, "develop", "", "The integration branch (merge target; default: the configured git-flow develop branch)")
	cmd.Flags().StringVar(&mergeCommit, "merge-commit", "", "A prepared merge commit — switches tree identity to the literal HEAD^{tree} == HEAD^2^{tree} form")
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id to record as the window holder (default: this session)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit machine-readable JSON")
	return cmd
}

// newFactoryMergeGateCommand is the REQ-FLA-011 gate: refuse unless a live
// acquire record for this lane covers the moment. The lane runs it before
// performing the merge in the integration worktree — the negative case is
// the property: no record, no merge on this path.
func newFactoryMergeGateCommand() *cobra.Command {
	var laneFlag, card string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Refuse unless a live acquire record for this lane covers the moment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane := strings.TrimSpace(laneFlag)
			if lane == "" {
				var err error
				if lane, err = factoryLaneLabelFromEnv("merge gate"); err != nil {
					return err
				}
			}
			lock, err := factory.ReadIntegrationLock(integrationLockRoot())
			if err != nil {
				return err
			}
			snapshot := factorylane.WindowSnapshot{
				Held:       lock.Held(),
				Live:       !lock.Stale(),
				HolderName: lock.SessionName,
			}
			if at, parseErr := time.Parse(time.RFC3339, lock.AcquiredAt); parseErr == nil {
				snapshot.AcquiredAt, snapshot.KnownAt = at, true
			}
			ok, why := factorylane.WindowCoversMerge(snapshot, lane, time.Now().UTC())
			verdict := "REFUSED"
			if ok {
				verdict = "PROCEED"
			}
			human := cmd.OutOrStdout()
			if asJSON {
				human = cmd.ErrOrStderr()
			}
			_, _ = fmt.Fprintf(human, "merge gate: %s — %s\n", verdict, why)
			if asJSON {
				data, err := json.Marshal(factoryMergeVerdict{
					Verdict: strings.ToLower(verdict), Lane: lane, Card: card, Holder: lock.SessionName, Detail: why,
				})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&laneFlag, "lane", "", "The lane the hold must name (default: the lane-identity environment variable)")
	cmd.Flags().StringVar(&card, "card", "", "The card being merged (carried on the JSON verdict)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit machine-readable JSON")
	return cmd
}

// printMergeChecks renders the recorded checks for the human channel.
func printMergeChecks(w io.Writer, checks []factorylane.MergeCheck) {
	for _, chk := range checks {
		state := "PASS"
		if !chk.Passed {
			state = "FAIL"
		}
		_, _ = fmt.Fprintf(w, "  %-14s %s  %s\n", chk.Name, state, strings.ReplaceAll(chk.Detail, "\n", "\n    "))
	}
}

// emitFactoryMergeVerdict writes the machine-readable verdict on the JSON
// path; stdout stays one parseable document per the JSON contract.
func emitFactoryMergeVerdict(cmd *cobra.Command, asJSON bool, v factoryMergeVerdict) error {
	if !asJSON {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}
