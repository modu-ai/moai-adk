package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/spf13/cobra"
)

var gtdIDPattern = regexp.MustCompile(`^gtd-[0-9a-f]{16}$`)

type gtdCLIOwner struct {
	readback func() (bool, error)
	apply    func() error
	// The dispatch reconciliation identifiers (review round-21): the
	// assigned card and its run, when this owner performs a dispatch whose
	// operation Target/MissionID name another vocabulary (the goal
	// mission's gtd item ref and session). Empty leaves the engine on the
	// op's own identifiers.
	reconcileCardID string
	reconcileRunID  string
}

func (o gtdCLIOwner) Readback(_ context.Context, _ factory.GTDOperation) (bool, error) {
	return o.readback()
}
func (o gtdCLIOwner) Apply(_ context.Context, _ factory.GTDOperation) error { return o.apply() }

// DispatchReconcileIdentifiers is factory.DispatchIdentifiers (review
// round-21): the assigned card and its run for a dispatch owner.
func (o gtdCLIOwner) DispatchReconcileIdentifiers() (string, string) {
	return o.reconcileCardID, o.reconcileRunID
}

func printGTD(cmd *cobra.Command, value any, jsonOutput bool) error {
	if jsonOutput {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), value)
	return err
}

func requireGTDID(id string) error {
	if !gtdIDPattern.MatchString(id) {
		return errors.New("gtd: invalid item id")
	}
	return nil
}

// NewGTDCommand returns the canonical user-facing name for the established
// queue command. The entire verb tree is built by newTodoCmd so gtd and the
// compatibility todo spelling cannot acquire different handlers or flags.
// Internal package names and the on-disk todo/backlog.db location deliberately
// remain unchanged.
func NewGTDCommand() *cobra.Command {
	cmd := newTodoCmd()
	cmd.Use = "gtd"
	cmd.Short = "Operate the GTD-managed backlog queue"
	cmd.Long = `Manage captured work through Capture, Clarify, Organize, Reflect, and Engage.

Captured GTD items stay separate from the established development queue. Only
an explicitly approved Engage operation may publish into the same backlog.db
used by the todo compatibility command; the queue's existing states, IDs,
ordering, archive, and restore behavior remain unchanged.`
	cmd.AddCommand(newGTDCaptureCmd(), newGTDClarifyCmd(), newGTDOrganizeCmd(), newGTDReflectCmd(), newGTDEngageCmd(), newGTDAnswerCmd())
	return cmd
}

func newGTDCaptureCmd() *cobra.Command {
	var source, sensitivity, eventID string
	var sourceAuthorized, jsonOutput bool
	cmd := &cobra.Command{Use: "capture <text>", Short: "Capture an inbox item without publishing a card", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		item, err := factory.CaptureGTDItem(cmd.Context(), newTodoStore(), factory.CaptureInput{Content: args[0], Source: source, SourceAllowed: source == "user" || sourceAuthorized, Sensitivity: factory.GTDSensitivity(sensitivity), EventID: eventID})
		if err != nil {
			return err
		}
		return printGTD(cmd, item, jsonOutput)
	}}
	cmd.Flags().StringVar(&source, "source", "user", "capture source")
	cmd.Flags().StringVar(&sensitivity, "sensitivity", "private", "public, private, or secret")
	cmd.Flags().StringVar(&eventID, "event", "", "stable capture event identity")
	cmd.Flags().BoolVar(&sourceAuthorized, "source-authorized", false, "confirm a non-user source is authorized")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	_ = cmd.MarkFlagRequired("event")
	return cmd
}

func newGTDClarifyCmd() *cobra.Command {
	var disposition, outcome, evidence, authority string
	var trusted, jsonOutput bool
	cmd := &cobra.Command{Use: "clarify <gtd-id>", Short: "Clarify outcome, evidence, trust, and authority", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireGTDID(args[0]); err != nil {
			return err
		}
		result, err := factory.ClarifyGTDItem(cmd.Context(), newTodoStore(), factory.ClarifyInput{ItemID: args[0], Disposition: factory.GTDDisposition(disposition), DesiredOutcome: outcome, CompletionEvidence: evidence, Authority: authority, SourceTrusted: trusted})
		if err != nil {
			return err
		}
		return printGTD(cmd, result, jsonOutput)
	}}
	cmd.Flags().StringVar(&disposition, "disposition", "", "action, reference, someday, waiting, or trash")
	cmd.Flags().StringVar(&outcome, "outcome", "", "desired outcome")
	cmd.Flags().StringVar(&evidence, "evidence", "", "completion evidence")
	cmd.Flags().StringVar(&authority, "authority", "", "delegated authority")
	cmd.Flags().BoolVar(&trusted, "trusted", false, "confirm the source is trusted")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	_ = cmd.MarkFlagRequired("disposition")
	return cmd
}

func newGTDOrganizeCmd() *cobra.Command {
	var class, contextName, reviewAt, partOf, dependsOn string
	var jsonOutput bool
	cmd := &cobra.Command{Use: "organize <gtd-id>", Short: "Organize a clarified item and its relationships", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireGTDID(args[0]); err != nil {
			return err
		}
		var relations []factory.GTDRelation
		for _, pair := range []struct {
			kind   factory.GTDRelationKind
			target string
		}{{factory.RelationPartOf, partOf}, {factory.RelationDependsOn, dependsOn}} {
			kind, target := pair.kind, pair.target
			if target == "" {
				continue
			}
			if err := requireGTDID(target); err != nil {
				return err
			}
			relations = append(relations, factory.GTDRelation{SubjectID: args[0], ObjectID: target, Kind: kind, Source: "operator", AssertionStatus: "confirmed", PolicyVersion: "gtd-v1"})
		}
		item, err := factory.OrganizeGTDItemWithRelations(cmd.Context(), newTodoStore(), factory.OrganizeInput{ItemID: args[0], Class: factory.GTDClass(class), Context: contextName, ReviewAt: reviewAt}, relations)
		if err != nil {
			return err
		}
		return printGTD(cmd, item, jsonOutput)
	}}
	cmd.Flags().StringVar(&class, "class", "", "action, project, reference, waiting, scheduled, or someday")
	cmd.Flags().StringVar(&contextName, "context", "", "action context")
	cmd.Flags().StringVar(&reviewAt, "review-at", "", "review time")
	cmd.Flags().StringVar(&partOf, "part-of", "", "project GTD item id")
	cmd.Flags().StringVar(&dependsOn, "depends-on", "", "prerequisite GTD item id")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	_ = cmd.MarkFlagRequired("class")
	return cmd
}

func newGTDReflectCmd() *cobra.Command {
	var jsonOutput, rebuild bool
	cmd := &cobra.Command{Use: "reflect", Short: "Review blockers, stale evidence, and missing next actions", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		view, err := factory.ReflectGTDStore(cmd.Context(), newTodoReadStore())
		if err != nil {
			return err
		}
		if rebuild {
			projection, err := graph.BuildPrivateGTDProjectionFromStore(cmd.Context(), newTodoReadStore())
			if err != nil {
				return err
			}
			return printGTD(cmd, map[string]any{"reflection": view, "projection": projection}, jsonOutput)
		}
		return printGTD(cmd, view, jsonOutput)
	}}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	cmd.Flags().BoolVar(&rebuild, "rebuild-projection", false, "rebuild the private GTD relation projection")
	return cmd
}

func newGTDEngageCmd() *cobra.Command {
	var approved, fresh, dependenciesReady, resources, pick, dispatch, jsonOutput bool
	var lane, runID string
	cmd := &cobra.Command{Use: "engage <gtd-id>", Short: "Publish, pick, and dispatch an approved actionable item", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireGTDID(args[0]); err != nil {
			return err
		}
		if dispatch && (!pick || strings.TrimSpace(lane) == "" || strings.TrimSpace(runID) == "") {
			return errors.New("gtd engage: dispatch requires --pick, --lane, and --run-id")
		}
		store := newTodoStore()
		item, err := factory.LoadGTDItem(cmd.Context(), store, args[0])
		if err != nil {
			return err
		}
		var result factory.EngageResult
		publishOwner := gtdCLIOwner{
			readback: func() (bool, error) {
				current, err := factory.LoadGTDItem(cmd.Context(), store, args[0])
				return current.CardID != "", err
			},
			apply: func() error {
				var applyErr error
				result, applyErr = factory.EngageGTDItem(cmd.Context(), store, factory.EngageInput{ItemID: args[0], Authorized: approved, EvidenceFresh: fresh, DependenciesReady: dependenciesReady, LaneAvailable: lane != "", ResourcesAvailable: resources})
				if applyErr == nil && result.CardID == "" {
					return fmt.Errorf("gtd engage: blocked: %s", strings.Join(result.Reasons, ","))
				}
				return applyErr
			},
		}
		op := factory.GTDOperation{OperationID: "gtd-publish:" + args[0], MissionID: runIDOrManual(runID), Action: "publish", Target: args[0], SnapshotHash: fmt.Sprintf("gtd-revision:%d", item.SourceRevision), ReceiptJSON: []byte(`{"source":"gtd_cli"}`)}
		if _, err = factory.ExecuteGTDOperation(cmd.Context(), store, op, publishOwner); err != nil {
			return err
		}
		item, err = factory.LoadGTDItem(cmd.Context(), store, args[0])
		if err != nil {
			return err
		}
		result = factory.EngageResult{Actionable: item.CardID != "", CardID: item.CardID}
		if result.CardID != "" && pick {
			pickOwner := gtdCLIOwner{readback: func() (bool, error) {
				record, err := store.LoadPure()
				if err != nil {
					return false, err
				}
				for _, card := range record.Items {
					if card.ID == result.CardID {
						return card.State == factory.BacklogStatePicked, nil
					}
				}
				return false, nil
			}, apply: func() error {
				return store.Mutate(func(record *factory.BacklogRecord) error {
					for i := range record.Items {
						if record.Items[i].ID == result.CardID {
							if record.Items[i].State == factory.BacklogStateQueued {
								record.Items[i].State = factory.BacklogStatePicked
								// REQ-TST-004: this is a picked transition too,
								// so the stamp rides the same locked write.
								record.Items[i].PickedAt = todoStampNow()
							}
							return nil
						}
					}
					return errors.New("gtd engage: published card not live")
				})
			}}
			pickOp := factory.GTDOperation{OperationID: "gtd-pick:" + args[0], MissionID: runIDOrManual(runID), Action: "pick", Target: result.CardID, SnapshotHash: op.SnapshotHash, ReceiptJSON: []byte(`{"source":"gtd_cli"}`)}
			if _, err = factory.ExecuteGTDOperation(cmd.Context(), store, pickOp, pickOwner); err != nil {
				return err
			}
		}
		if result.CardID != "" && dispatch {
			dispatchOwner := gtdCLIOwner{readback: func() (bool, error) {
				record, err := store.LoadPure()
				if err != nil {
					return false, err
				}
				// card t1349 integration repair: assignments are stored through
				// the REQ-TSP-052 canonical write normalization, so the readback
				// maps the caller's lane label onto the same canonical form
				// before comparing — a legacy-spelled lane keeps its authority.
				want := factory.NormalizeOwnerLabel(lane)
				for _, a := range record.Runtime.Assignments {
					if a.RunID == runID && a.CardID == result.CardID && a.OwnerLabel == want {
						return true, nil
					}
				}
				return false, nil
			}, apply: func() error {
				root := resolveTodoQueueRoot()
				// The assignment primitive carries the dispatch binding in
				// its own queue-lock critical section (review round-20 P1):
				// the completion path cannot interleave between the save
				// and the re-point, and even when the mirror fails
				// (FACTORY_RECORD_UNAVAILABLE) the old run's approval is
				// inert — the completion gate resolves the NEW run and
				// refuses.
				if err := factory.RecordFactoryCardAssignment(root, runID, result.CardID, lane, ""); err != nil {
					return err
				}
				mirrorFactoryAssignment(cmd.Context(), cmd.ErrOrStderr(), root, store, runID, result.CardID, lane)
				return nil
			}}
			dispatchOp := factory.GTDOperation{OperationID: "gtd-dispatch:" + runID + ":" + args[0], MissionID: runID, Action: "dispatch", Target: result.CardID, SnapshotHash: op.SnapshotHash, ReceiptJSON: []byte(`{"source":"gtd_cli"}`)}
			if _, err = factory.ExecuteGTDOperation(cmd.Context(), store, dispatchOp, dispatchOwner); err != nil {
				return err
			}
		}
		return printGTD(cmd, result, jsonOutput)
	}}
	cmd.Flags().BoolVar(&approved, "approve", false, "explicitly approve queue effects")
	cmd.Flags().BoolVar(&fresh, "fresh", false, "confirm current evidence revision")
	cmd.Flags().BoolVar(&dependenciesReady, "dependencies-ready", false, "confirm dependencies are complete")
	cmd.Flags().BoolVar(&resources, "resources", false, "confirm resource budget")
	cmd.Flags().BoolVar(&pick, "pick", false, "pick the published card")
	cmd.Flags().BoolVar(&dispatch, "dispatch", false, "record dispatch after picking")
	cmd.Flags().StringVar(&lane, "lane", "", "available lane owner")
	cmd.Flags().StringVar(&runID, "run-id", "", "stable mission run identity")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func runIDOrManual(runID string) string {
	if strings.TrimSpace(runID) != "" {
		return runID
	}
	return "manual-gtd"
}

func init() {
	rootCmd.AddCommand(NewGTDCommand())
}
