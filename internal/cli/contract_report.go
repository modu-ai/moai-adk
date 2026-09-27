package cli

// contract_report.go — `moai contract report <card-id>`, the A4 closure
// report command (SPEC-AUTONOMY-CLOSURE-001 REQ-CLOSURE-001/024).
//
// The command is a thin adapter over internal/closure: it resolves the card's
// SPEC ID from the queue store, the card evidence home by the §C.7 rule,
// verifies the contract through A1, and hands plain values to closure.Build.
// Exit codes: 0 report written, 2 refusal (unknown card, no SPEC, contract
// mismatch, I/O).
//
// The three A4 subcommands (report, verdict, push-check) attach to A1's
// `moai contract` command in THIS file's init: Go runs package-level init()
// in file-name order, and "contract.go" sorts before "contract_report.go",
// so the command tree exists when registration runs. A1's own files stay
// untouched (plan.md §D).

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// contractQueueRootFn resolves the queue root (the primary checkout — the
// queue is a property of one repository, never of a worktree). Tests replace
// it.
var contractQueueRootFn = func() string {
	return kanban.ResolveTodoQueueRoot(resolveProjectDir())
}

// contractRunDirFn is the tree the command runs in; the card evidence home
// is resolved against it. Tests replace it.
var contractRunDirFn = func() (string, error) { return findProjectRootFn() }

// cardToSpecID resolves a card's SPEC ID from the queue store, "" with a
// cause when the card is unknown or carries no SPEC.
func cardToSpecID(card string) (string, string) {
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(contractQueueRootFn()))
	rec, err := store.LoadPure()
	if err != nil {
		return "", "queue store unreadable: " + err.Error()
	}
	for _, item := range rec.Items {
		if item.ID != card {
			continue
		}
		if item.SpecID == nil || *item.SpecID == "" {
			return "", "card " + card + " has no SPEC ID"
		}
		return *item.SpecID, ""
	}
	return "", "card " + card + " is not in the queue"
}

// detectorArmedFor applies the armed-for-the-contract-in-force rule: the
// card audit log shows an arming in force and the card state's armed
// snapshot carries the contract's recorded digest.
func detectorArmedFor(home, card, recordedDigest string) bool {
	files, err := escalation.CardFilesFor(home, card)
	if err != nil {
		return false
	}
	lg, err := escalation.ReadCardLog(files.Log)
	if err != nil || !lg.Armed() {
		return false
	}
	st, ok, err := escalation.ReadCardState(files.State)
	if err != nil || !ok {
		return false
	}
	return st.Armed != nil && st.Armed.ContractSHA256 == recordedDigest
}

// runContractReport builds and writes the closure report pair.
func runContractReport(cmd *cobra.Command, cardID string) error {
	specID, cause := cardToSpecID(cardID)
	if specID == "" {
		return contractUsageError(cmd, "%s", cause)
	}
	runDir, err := contractRunDirFn()
	if err != nil {
		return contractUsageError(cmd, "%v", err)
	}
	home, err := closure.ResolveEvidenceHome(runDir, cardID)
	if err != nil {
		return contractUsageError(cmd, "card evidence home: %v", err)
	}
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}

	// Contract facts from the card evidence home's own SPEC directory.
	rep, progress, acceptance, ac, acOK, err := closure.LoadBuildSideFiles(
		home, specID, env.autonomy, env.ruleIDs, env.frozenFiles)
	if err != nil {
		return contractUsageError(cmd, "%v", err)
	}
	if rep.Contract == nil || rep.Card != cardID {
		named := "(none)"
		if rep.Contract != nil {
			named = rep.Card
		}
		return contractUsageError(cmd, "the contract of %s names card %q, not %q", specID, named, cardID)
	}

	ev := closure.EvidenceFor(home, cardID)
	records, unreadable, err := closure.LoadA2Records(home, cardID)
	if err != nil {
		return contractUsageError(cmd, "escalation records: %v", err)
	}
	secondReviews, secondSkipped, _ := closure.LoadSecondReviews(ev.SecondReview)
	verdicts, _, _ := closure.LoadVerdictRecords(ev.ClosureVerdict)
	receipt, receiptPresent := readSpecReceipt(filepath.Join(home, ".moai", "specs", specID))

	head, err := gitio.Head(home)
	if err != nil {
		return contractUsageError(cmd, "resolve report HEAD: %v", err)
	}

	in := closure.BuildInput{
		Card: cardID, SpecID: specID, Home: home,
		Mode:                env.autonomy.Mode,
		SecondReviewPolicy:  env.autonomy.SecondReview,
		Verify:              rep,
		Receipt:             receipt,
		ReceiptPresent:      receiptPresent,
		ProgressMD:          progress,
		AcceptanceMD:        acceptance,
		MeasuredAC:          ac,
		ACAvailable:         acOK,
		SecondReviews:       secondReviews,
		SecondReviewSkipped: secondSkipped,
		Verdicts:            verdicts,
		Records:             records,
		UnreadableRecords:   unreadable,
		DetectorArmed:       detectorArmedFor(home, cardID, rep.RecordedContractSHA256),
		Facts: closure.GitFacts{
			IsAncestor:      func(a, b string) (bool, error) { return gitio.IsAncestor(home, a, b) },
			NonMergeCommits: func(from, to string) ([]closure.CommitPaths, error) { return gitio.NonMergeCommits(home, from, to) },
		},
		EvalCommit:         head,
		PreviousReportHash: closure.PreviousReportHashOf(ev.ReportJSON),
	}
	report, err := closure.Build(in)
	if err != nil {
		return contractUsageError(cmd, "build report: %v", err)
	}
	mdPath, err := closure.WriteReportPair(ev, report)
	if err != nil {
		return contractUsageError(cmd, "write report: %v", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), mdPath)
	return nil
}

// readSpecReceipt reads a SPEC directory's kickoff receipt (nil when absent).
func readSpecReceipt(specDir string) ([]byte, bool) {
	data, err := os.ReadFile(filepath.Join(specDir, contract.ReceiptFile))
	if err != nil {
		return nil, false
	}
	return data, true
}

// newContractReportCmd builds the `report` subcommand.
func newContractReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "report <card-id>",
		Short:         "Write the card's closure report pair (exit 0 written, 2 refused)",
		Args:          contractArgs(1, 1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractReport(c, args[0])
		},
	}
}

// registerClosureCommands attaches the three A4 subcommands to the existing
// `moai contract` command tree without touching A1's file.
func registerClosureCommands() {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "contract" {
			c.AddCommand(newContractReportCmd(), newContractVerdictCmd(), newContractPushCheckCmd())
			return
		}
	}
	// The contract command is registered by contract.go's init in the same
	// package; reaching this line would mean that file changed shape.
	panic("moai contract: command tree not found for A4 subcommands")
}

func init() {
	registerClosureCommands()
}
