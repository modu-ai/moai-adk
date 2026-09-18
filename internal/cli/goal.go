// @MX:ANCHOR: [AUTO] moai goal CLI — the arming surface for the goal engine
// @MX:REASON: fan_in via the goal-engine loop — `goal arm` writes .moai/state/goal/<session-id>.json
// that the `moai hook stop-goal` evaluator loads on every turn-end. The arm↔eval shared-file keying
// (the SAME session id on both sides) is the load-bearing contract; a pid-keyed arm file would be
// unreachable by the hook (different PID), leaving the armed goal inert.
package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/mission"
	"github.com/modu-ai/moai-adk/internal/session"
)

// trailingExitClause matches an optional trailing "exits <N>" / "exit <N>" suffix
// on a mechanical condition string (e.g. "go test ./... exits 0" → cmd="go test
// ./...", expect_exit=0). The clause is stripped so the shell never runs the
// literal words "exits 0" as command arguments.
var trailingExitClause = regexp.MustCompile(`(?i)^(.*\S)\s+exits?\s+(\d+)\s*$`)

// modelConditionReferents are the literal tokens (case-insensitive) that mark a
// condition string as a claim about the conversation transcript, per REQ-GLE-032
// ("a natural-language claim that references the conversation transcript").
// BOTH referents in that phrase count: the canonical CLI form says "in the
// transcript", while the canonical ac_converge condition in run.md § Run-phase
// Autonomy says "surfaced in the conversation". Keying on "transcript" alone
// routed the whole ac_converge paragraph into the mechanical path, where it ran
// as a shell command, exited 2, and blocked every turn-end to the ceiling
// (issue #1660). Both the CLI arm path and the MCP goal_arm wrapper classify
// through this one function, so the miss was shared by both.
var modelConditionReferents = []string{"transcript", "conversation"}

// conditionDeclarationPrefix matches a leading `model:` / `cmd:` declaration
// prefix (case-insensitive, surrounding whitespace tolerated) on a condition
// string. Only those two exact words followed by a colon count — "modelling:"
// and "cmdline:" are ordinary text, not prefixes.
var conditionDeclarationPrefix = regexp.MustCompile(`(?is)^\s*(model|cmd)\s*:\s*(.*)$`)

// parseCondition classifies a single condition string by the EXPLICIT rule of
// REQ-GLE-032: a claim that references the conversation transcript becomes a
// model condition; any other string is a runnable shell command (mechanical).
//
// Classification runs in two stages, most explicit first:
//
//  1. An explicit `model:` / `cmd:` declaration prefix wins outright. REQ-GLE-032
//     asks for classification by an EXPLICIT rule rather than an implicit
//     heuristic, and only the prefix actually delivers that: the author says
//     which tier they meant instead of hoping the discriminator guesses it.
//  2. Absent a prefix, the modelConditionReferents substring fallback applies
//     unchanged (back-compat for every already-armed and already-documented
//     condition).
//
// The fallback is retained but demoted deliberately. Its discriminator is an
// ENGLISH substring allowlist, and an allowlist fails silently on what it omits:
// prose in any other language — and English prose phrased without those two
// words — classifies mechanical, reaches `sh -c`, and exits 127 on every turn
// until the ceiling. The prefix is the escape hatch that does not depend on the
// allowlist being complete.
//
// A mechanical condition may carry a trailing "exits <N>" clause setting the
// expected exit code; absent it, expect_exit defaults to 0. The clause is parsed
// on the prefixed form too, so `cmd: grep -q X f exits 1` behaves as expected.
func parseCondition(s string) goal.Condition {
	s = strings.TrimSpace(s)
	if m := conditionDeclarationPrefix.FindStringSubmatch(s); m != nil {
		body := strings.TrimSpace(m[2])
		if strings.EqualFold(m[1], "model") {
			return goal.Condition{Type: goal.ConditionModel, Claim: body}
		}
		return mechanicalCondition(body)
	}
	lower := strings.ToLower(s)
	for _, referent := range modelConditionReferents {
		if strings.Contains(lower, referent) {
			return goal.Condition{Type: goal.ConditionModel, Claim: s}
		}
	}
	return mechanicalCondition(s)
}

// mechanicalCondition builds a Tier-1 condition from a command string, peeling
// the optional trailing "exits <N>" clause so the shell never runs those literal
// words as arguments.
func mechanicalCondition(s string) goal.Condition {
	cmd := s
	expect := 0
	if m := trailingExitClause.FindStringSubmatch(s); m != nil {
		cmd = strings.TrimSpace(m[1])
		expect, _ = strconv.Atoi(m[2])
	}
	return goal.Condition{Type: goal.ConditionMechanical, Cmd: cmd, ExpectExit: expect}
}

// NewAutoMissionCommand builds both preserved condition-goal commands and the
// separately persisted auto-mission lifecycle: approve, run, revoke, and resume.
// Condition goals still reuse the internal/goal engine without changing its
// state, schema, pruning, or clear semantics.
//
// IMPORTANT: this CLI surface MUST NOT invoke AskUserQuestion (subagent boundary,
// C-HRA-008). It returns exit codes + structured stdout only; the orchestrator
// owns all user interaction.
func NewAutoMissionCommand() *cobra.Command {
	var sessionFlag string
	var jsonOutput bool
	var showAll bool
	var autoMission bool
	var approvalScope, approvalActions, approvalEvidence []string
	var approvalMaxOperations int
	var runAction, runTarget string
	var governorRecommend, supervise bool
	var runRepo, runCardWorktree, runDevelopWorktree, runBranch, runMessage, runCardSHA, runBaseSHA, runLease, runTestsReceipt, runCompletionReceipt string
	var runGovernorReceipt, runAuditReceipt, runLane, runID string
	var runPaths []string

	cmd := &cobra.Command{
		Use:   "goal",
		Short: "Arm/inspect/clear a condition-declared agentic goal loop for this session",
		Long: `Arm a condition-declared goal loop for the active session.

The MoAI goal engine keeps the session working across turns until the declared
conditions hold or a turn ceiling is reached. State lives at
.moai/state/goal/<session-id>.json (one file per session). The Stop hook
'moai hook stop-goal' evaluates the goal each turn-end.

Verbs:
  goal arm "<condition>"   register + arm a goal (bare 'goal "<condition>"' aliases arm)
  goal status              print the active session's goal state
  goal clear               clear the active session's goal
  goal render              render the live goal's dashboard to a self-contained HTML file

Auto missions:
  goal --auto "<mission>"  create a natural-language autonomous mission draft
  goal approve             seal explicit scope, actions, evidence, and limits
  goal run                 validate one receipt-backed owner operation
  goal run --supervise     run the bounded sealed action plan until complete or blocked
  goal revoke              stop new effects and retain reconciliation state
  goal resume              resume only a persisted blocked approved mission

Approval required before effects. After approval, deterministic validators and
authoritative readback gate every operation. Runtime mode remains
active-session-only unless durable provider capabilities are mechanically proven.

Condition parsing: a runnable shell command (optionally suffixed "exits <N>")
becomes a mechanical condition; a claim that references the conversation
transcript becomes a model condition the orchestrator evaluates.`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if autoMission {
				return runGoalAutoMission(c, args, sessionFlag, jsonOutput)
			}
			// Bare form: `goal "<condition>"` aliases `goal arm "<condition>"`.
			if len(args) == 0 {
				return c.Help()
			}
			// A single word a user would type as a verb is not a condition.
			// Only the bare form is checked: `goal arm <word>` states the
			// intent explicitly, and a multi-word or prefixed condition never
			// matches a listed word.
			if len(args) == 1 {
				word := strings.TrimSpace(args[0])
				if strings.EqualFold(word, "help") {
					return c.Help()
				}
				if verb, ok := misreadGoalVerb(c, word); ok {
					return misreadGoalVerbError(word, verb)
				}
				if bareWordNeedsDeclaration(word) {
					return bareWordDeclarationError(word)
				}
			}
			return runGoalArm(c, args, sessionFlag, jsonOutput)
		},
	}
	cmd.PersistentFlags().StringVar(&sessionFlag, "session", "", "override the session id (default: resolve via 'moai session current')")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
	cmd.PersistentFlags().BoolVar(&autoMission, "auto", false, "create a natural-language autonomous mission without condition or shell parsing")
	// SPEC-INFINITE-GOAL-001 REQ-1/REQ-4 arm-time bound flags. --max-turns N
	// (0 = infinite, the C2 finding's entry point); --max-duration <seconds> and
	// --cost-cap <N> are the REAL bounds required when --max-turns 0 is supplied
	// (AC-011 fail-closed reject). Sentinel -1 means "omitted" so the default
	// (30) from NewGoal is preserved when the flag is absent (AC-002).
	cmd.PersistentFlags().Int("max-turns", -1, "turn ceiling (0 = infinite; default 30 when omitted)")
	cmd.PersistentFlags().Int("max-duration", 0, "wall-clock bound in seconds since arm time (real bound for --max-turns 0)")
	cmd.PersistentFlags().Int("cost-cap", 0, "cost bound (max invocations); recorded, enforcement is a follow-up")

	armCmd := &cobra.Command{
		Use:          "arm <condition>",
		Short:        "Register and arm a goal for the active session",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if autoMission {
				return runGoalAutoMission(c, args, sessionFlag, jsonOutput)
			}
			return runGoalArm(c, args, sessionFlag, jsonOutput)
		},
	}
	statusCmd := &cobra.Command{
		Use:          "status",
		Short:        "Print the active session's goal state",
		SuggestFor:   []string{"show", "list", "info", "stat"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runGoalStatus(c, sessionFlag, jsonOutput, showAll)
		},
	}
	statusCmd.Flags().BoolVar(&showAll, "all", false, "list goals for every session, not just the active one")

	clearCmd := &cobra.Command{
		Use:          "clear",
		Short:        "Clear the active session's goal",
		SuggestFor:   []string{"cancel", "reset", "stop", "done"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runGoalClear(c, sessionFlag, jsonOutput)
		},
	}

	// renderCmd (SPEC-GOAL-HTML-FLOW-001 REQ-GHF-004) renders the live goal's
	// dashboard to a self-contained .html file beside the .json state. It reuses
	// statusSessionID + goal.LoadGoal (the read path), calls goal.RenderDashboard,
	// and writes to goal.HTMLPath. With no armed goal it exits non-zero, names the
	// session id on stderr, and writes NO html (AC-GHF-010).
	renderCmd := &cobra.Command{
		Use:          "render",
		Short:        "Render the active session's goal dashboard to a self-contained HTML file",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runGoalRender(c, sessionFlag, jsonOutput)
		},
	}
	approveCmd := &cobra.Command{Use: "approve", Short: "Approve and seal an autonomous mission", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		return runGoalMissionApprove(c, sessionFlag, jsonOutput, approvalScope, approvalActions, approvalEvidence, approvalMaxOperations)
	}}
	approveCmd.Flags().StringSliceVar(&approvalScope, "scope", nil, "approved target scope")
	approveCmd.Flags().StringSliceVar(&approvalActions, "action", nil, "approved autonomous action")
	approveCmd.Flags().StringSliceVar(&approvalEvidence, "completion-evidence", nil, "mission completion evidence")
	approveCmd.Flags().IntVar(&approvalMaxOperations, "max-operations", 20, "maximum autonomous operations")
	runCmd := &cobra.Command{Use: "run", Short: "Advance one validated autonomous mission operation", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		opts := missionGitRunOptions{Repository: runRepo, CardWorktree: runCardWorktree, DevelopWorktree: runDevelopWorktree, Branch: runBranch, Message: runMessage, Paths: runPaths, CardSHA: runCardSHA, BaseSHA: runBaseSHA, LeasePath: runLease, TestsReceipt: runTestsReceipt, CompletionReceipt: runCompletionReceipt, GovernorReceipt: runGovernorReceipt, AuditReceipt: runAuditReceipt, Lane: runLane, RunID: runID}
		if supervise {
			return runGoalMissionSupervisor(c, sessionFlag, jsonOutput, runTarget, opts)
		}
		if runAction == "" || runTarget == "" {
			return errors.New("auto mission run: --action and --target are required unless --supervise is set")
		}
		return runGoalMissionOperation(c, sessionFlag, jsonOutput, runAction, runTarget, governorRecommend, opts)
	}}
	runCmd.Flags().StringVar(&runAction, "action", "", "proposed action")
	runCmd.Flags().StringVar(&runTarget, "target", "", "proposed target")
	runCmd.Flags().BoolVar(&governorRecommend, "recommend", false, "compatibility flag only; never grants authority")
	runCmd.Flags().StringVar(&runGovernorReceipt, "governor-receipt", "", "0600 mission-governor decision receipt")
	runCmd.Flags().StringVar(&runAuditReceipt, "audit-receipt", "", "0600 independent audit receipt")
	runCmd.Flags().StringVar(&runLane, "lane", "", "leased lane for dispatch")
	runCmd.Flags().StringVar(&runID, "run-id", "", "stable disk dispatch run identity")
	runCmd.Flags().BoolVar(&supervise, "supervise", false, "run the bounded sealed action plan until complete or durably blocked")
	runCmd.Flags().StringVar(&runRepo, "repo", "", "absolute repository path for a Git owner action")
	runCmd.Flags().StringVar(&runCardWorktree, "card-worktree", "", "absolute WT-* card worktree path")
	runCmd.Flags().StringVar(&runDevelopWorktree, "develop-worktree", "", "absolute integration develop worktree path")
	runCmd.Flags().StringVar(&runBranch, "worktree-branch", "", "WT-* card branch")
	runCmd.Flags().StringVar(&runMessage, "message", "", "commit message")
	runCmd.Flags().StringSliceVar(&runPaths, "path", nil, "explicit path to stage")
	runCmd.Flags().StringVar(&runCardSHA, "card-sha", "", "card commit SHA for local merge")
	runCmd.Flags().StringVar(&runBaseSHA, "base-sha", "", "leased local develop base SHA")
	runCmd.Flags().StringVar(&runLease, "integration-lease", "", "0600 integration lease receipt under .git")
	runCmd.Flags().StringVar(&runTestsReceipt, "tests-receipt", "", "0600 JSON test receipt inside the repository")
	runCmd.Flags().StringVar(&runCompletionReceipt, "completion-receipt", "", "0600 sealed authoritative completion receipt")
	revokeCmd := &cobra.Command{Use: "revoke", Short: "Revoke an autonomous mission", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error { return runGoalMissionRevoke(c, sessionFlag, jsonOutput) }}
	resumeCmd := &cobra.Command{Use: "resume", Short: "Resume a policy-blocked autonomous mission", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error { return runGoalMissionResume(c, sessionFlag, jsonOutput) }}

	cmd.AddCommand(armCmd, statusCmd, clearCmd, renderCmd, approveCmd, runCmd, revokeCmd, resumeCmd)
	return cmd
}

func loadRequiredAutoMission(root, sessionID string) (*mission.AutoMission, error) {
	if err := mission.ValidateMissionSessionID(sessionID); err != nil {
		return nil, err
	}
	state, err := mission.LoadAutoMission(root, sessionID)
	if err != nil {
		if state != nil {
			state.State = mission.StateBlocked
			state.LastBlocker = err.Error()
			_ = mission.SaveAutoMission(root, *state)
		}
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("auto mission not found")
	}
	return state, nil
}

func runGoalMissionApprove(cmd *cobra.Command, sessionID string, jsonOutput bool, scope, actions, evidence []string, maxOperations int) error {
	root := goalProjectRoot()
	sessionID = statusSessionID(sessionID)
	state, err := loadRequiredAutoMission(root, sessionID)
	if err != nil {
		return err
	}
	allowed := make([]mission.Action, 0, len(actions))
	for _, a := range actions {
		allowed = append(allowed, mission.Action(a))
	}
	contract := mission.MissionContract{MissionID: sessionID, Goal: state.Text, CompletionEvidence: evidence, Scope: scope, AllowedActions: allowed, MergeTarget: "develop", ResourceLimits: mission.ResourceLimits{MaxOperations: maxOperations, MaxRetries: 2}, ProhibitedActions: []mission.Action{mission.ActionForcePush}, StopConditions: []string{"revoked", "policy_denied"}, RecoveryConditions: []string{"authoritative_readback"}, RevocationBehavior: "stop_new_and_reconcile", PolicyVersion: "gtd-auto-v1", Approved: true}
	sealed, err := mission.SealMissionContract(contract)
	if err != nil {
		return err
	}
	state.Contract = &sealed.Contract
	state.ContractHash = sealed.Hash
	state.State = mission.StateApproved
	state.LastBlocker = ""
	if err := mission.SaveAutoMission(root, *state); err != nil {
		return err
	}
	if jsonOutput {
		return printGTD(cmd, state, true)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "approved auto mission %s contract %s\n", sessionID, sealed.Hash)
	return err
}

func autoSnapshotHash(contractHash, target string, revision int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d", contractHash, target, revision)))
	return fmt.Sprintf("%x", sum[:])
}

func autoGitSnapshotHash(contractHash, target string, action mission.Action, repository, head, cardSHA string) string {
	resolved, _ := filepath.EvalSymlinks(repository)
	sum := sha256.Sum256([]byte(strings.Join([]string{contractHash, target, string(action), resolved, head, cardSHA}, "\x00")))
	return fmt.Sprintf("%x", sum[:])
}

func actionRepository(opts missionGitRunOptions, action mission.Action) string {
	if action == mission.ActionCommit && opts.CardWorktree != "" {
		return opts.CardWorktree
	}
	if action == mission.ActionLocalMerge && opts.DevelopWorktree != "" {
		return opts.DevelopWorktree
	}
	return opts.Repository
}

func persistMissionBlock(root string, state *mission.AutoMission, reason string) error {
	state.State = mission.StateBlocked
	state.LastBlocker = reason
	return mission.SaveAutoMission(root, *state)
}

type missionGitRunOptions struct {
	Repository, CardWorktree, DevelopWorktree                  string
	Branch, Message, CardSHA, BaseSHA, LeasePath, TestsReceipt string
	CompletionReceipt                                          string
	GovernorReceipt, AuditReceipt, Lane, RunID                 string
	Paths                                                      []string
}

type cliSupervisorStore struct {
	root, session string
}

func (s cliSupervisorStore) Load(_ context.Context, _ string) (mission.SupervisorState, error) {
	state, err := loadRequiredAutoMission(s.root, s.session)
	if err != nil {
		return mission.SupervisorState{}, err
	}
	return mission.SupervisorState{MissionID: state.SessionID, ContractHash: state.ContractHash, State: state.State, NextStep: len(state.OperationIDs), OperationIDs: append([]string(nil), state.OperationIDs...), Blocker: state.LastBlocker}, nil
}

func (s cliSupervisorStore) Save(_ context.Context, value mission.SupervisorState) error {
	state, err := mission.LoadAutoMission(s.root, s.session)
	if err != nil || state == nil {
		return err
	}
	state.State = value.State
	state.LastBlocker = value.Blocker
	state.OperationIDs = append([]string(nil), value.OperationIDs...)
	return mission.SaveAutoMission(s.root, *state)
}

type cliSupervisorEngine struct {
	cmd                *cobra.Command
	root, session      string
	jsonOutput         bool
	governor, audit    string
	gitOpts            missionGitRunOptions
	latestSnapshotHash string
	latestHeadSHA      string
}

func receiptPathForAction(pattern string, action mission.Action) string {
	return strings.ReplaceAll(pattern, "{action}", string(action))
}

func (e *cliSupervisorEngine) Snapshot(ctx context.Context, step mission.SupervisionStep) (mission.SupervisionSnapshot, error) {
	state, err := loadRequiredAutoMission(e.root, e.session)
	if err != nil {
		return mission.SupervisionSnapshot{}, err
	}
	revision := int64(1)
	if strings.HasPrefix(step.Target, "gtd:") {
		item, err := kanban.LoadGTDItem(ctx, newTodoStore(), strings.TrimPrefix(step.Target, "gtd:"))
		if err != nil {
			return mission.SupervisionSnapshot{}, err
		}
		revision = item.SourceRevision
	}
	if step.Action == mission.ActionCommit || step.Action == mission.ActionLocalMerge {
		repo := actionRepository(e.gitOpts, step.Action)
		head, err := currentMissionHead(repo)
		if err != nil {
			return mission.SupervisionSnapshot{}, err
		}
		cardSHA := ""
		if step.Action == mission.ActionLocalMerge {
			cardSHA, err = currentMissionHead(e.gitOpts.CardWorktree)
			if err != nil || (e.gitOpts.CardSHA != "" && cardSHA != e.gitOpts.CardSHA) {
				return mission.SupervisionSnapshot{}, errors.New("auto mission supervisor: card_sha_mismatch")
			}
			e.gitOpts.CardSHA = cardSHA
		}
		e.latestHeadSHA = head
		e.latestSnapshotHash = autoGitSnapshotHash(state.ContractHash, step.Target, step.Action, repo, head, cardSHA)
	} else {
		e.latestHeadSHA, err = currentMissionHead(e.root)
		if err != nil {
			return mission.SupervisionSnapshot{}, err
		}
		e.latestSnapshotHash = autoSnapshotHash(state.ContractHash, step.Target, revision)
	}
	return mission.SupervisionSnapshot{Hash: e.latestSnapshotHash}, nil
}

func (e *cliSupervisorEngine) Governance(_ context.Context, step mission.SupervisionStep, snapshot mission.SupervisionSnapshot) error {
	state, err := loadRequiredAutoMission(e.root, e.session)
	if err != nil {
		return err
	}
	_, err = mission.LoadGovernanceReceipts(e.root, receiptPathForAction(e.governor, step.Action), receiptPathForAction(e.audit, step.Action), mission.GovernanceExpectation{MissionID: e.session, ContractHash: state.ContractHash, SnapshotHash: snapshot.Hash, Action: step.Action, Targets: []string{step.Target}, HeadSHA: e.latestHeadSHA, Now: time.Now()})
	return err
}

func (e *cliSupervisorEngine) Validate(_ context.Context, step mission.SupervisionStep, snapshot mission.SupervisionSnapshot) error {
	state, err := loadRequiredAutoMission(e.root, e.session)
	if err != nil {
		return err
	}
	if snapshot.Hash != e.latestSnapshotHash || state.Contract == nil || !slices.Contains(state.Contract.AllowedActions, step.Action) {
		return errors.New("auto mission supervisor: validation_failed")
	}
	return nil
}

func (e *cliSupervisorEngine) Execute(ctx context.Context, step mission.SupervisionStep, _ mission.SupervisionSnapshot) (mission.OperationReceipt, error) {
	opts := e.gitOpts
	opts.GovernorReceipt = receiptPathForAction(e.governor, step.Action)
	opts.AuditReceipt = receiptPathForAction(e.audit, step.Action)
	if err := runGoalMissionOperation(e.cmd, e.session, e.jsonOutput, string(step.Action), step.Target, false, opts); err != nil {
		return mission.OperationReceipt{}, err
	}
	state, err := loadRequiredAutoMission(e.root, e.session)
	if err != nil || len(state.OperationIDs) == 0 {
		return mission.OperationReceipt{}, errors.New("auto mission supervisor: operation_receipt_missing")
	}
	op, err := kanban.LoadGTDOperation(ctx, newTodoStore(), state.OperationIDs[len(state.OperationIDs)-1])
	if err != nil {
		return mission.OperationReceipt{}, err
	}
	return mission.OperationReceipt{OperationID: op.OperationID, MissionID: op.MissionID, Action: mission.Action(op.Action), SnapshotHash: op.SnapshotHash, State: mission.ReceiptReconciled}, nil
}

func (e *cliSupervisorEngine) Readback(ctx context.Context, _ mission.SupervisionStep, receipt mission.OperationReceipt) error {
	op, err := kanban.LoadGTDOperation(ctx, newTodoStore(), receipt.OperationID)
	if err != nil {
		return err
	}
	if op.State != kanban.GTDOperationReconciled || op.SnapshotHash != receipt.SnapshotHash {
		return errors.New("auto mission supervisor: authoritative_readback_missing")
	}
	return nil
}

func (e *cliSupervisorEngine) Finalize(ctx context.Context, plan mission.SupervisionPlan, state mission.SupervisorState) error {
	auto, err := loadRequiredAutoMission(e.root, e.session)
	if err != nil || auto.Snapshot == nil {
		return errors.New("auto mission supervisor: completion_snapshot_missing")
	}
	if len(state.OperationIDs) != len(plan.Steps) {
		return errors.New("auto mission supervisor: completion_operations_missing")
	}
	store := newTodoStore()
	for i, id := range state.OperationIDs {
		op, loadErr := kanban.LoadGTDOperation(ctx, store, id)
		if loadErr != nil || op.State != kanban.GTDOperationReconciled || op.MissionID != plan.MissionID || op.Action != string(plan.Steps[i].Action) || op.Target != plan.Steps[i].Target {
			return errors.New("auto mission supervisor: completion_operation_lineage")
		}
	}
	headRoot := e.root
	if plan.RequireLandedAncestry {
		headRoot = e.gitOpts.DevelopWorktree
	} else if e.gitOpts.CardWorktree != "" {
		headRoot = e.gitOpts.CardWorktree
	}
	head, err := currentMissionHead(headRoot)
	if err != nil {
		return err
	}
	_, err = mission.LoadCompletionReceipt(e.root, e.gitOpts.CompletionReceipt, mission.CompletionExpectation{MissionID: plan.MissionID, ContractHash: plan.ContractHash, SnapshotHash: auto.Snapshot.SnapshotHash, HeadSHA: head, RequiredEvidence: plan.CompletionEvidence, RequireLandedAncestry: plan.RequireLandedAncestry, Now: time.Now()})
	return err
}

func runGoalMissionSupervisor(cmd *cobra.Command, sessionID string, jsonOutput bool, target string, opts missionGitRunOptions) error {
	root := goalProjectRoot()
	sessionID = statusSessionID(sessionID)
	state, err := loadRequiredAutoMission(root, sessionID)
	if err != nil {
		return err
	}
	if state.Contract == nil || state.ContractHash == "" {
		return errors.New("auto mission supervisor: sealed contract required")
	}
	if target == "" {
		for _, candidate := range state.Contract.Scope {
			if strings.HasPrefix(candidate, "gtd:") {
				target = candidate
				break
			}
		}
	}
	if target == "" {
		return errors.New("auto mission supervisor: target required")
	}
	gitCommit := slices.Contains(state.Contract.AllowedActions, mission.ActionCommit)
	gitMerge := slices.Contains(state.Contract.AllowedActions, mission.ActionLocalMerge)
	if (gitCommit || gitMerge) && (opts.Repository != "" || !filepath.IsAbs(opts.CardWorktree) || (gitMerge && !filepath.IsAbs(opts.DevelopWorktree)) || opts.CardWorktree == opts.DevelopWorktree) {
		return errors.New("auto mission supervisor: split_worktree_required")
	}
	steps := make([]mission.SupervisionStep, 0, len(state.Contract.AllowedActions))
	ordered := []mission.Action{mission.ActionPublish, mission.ActionPick, mission.ActionDispatch, mission.ActionCommit, mission.ActionLocalMerge, mission.ActionBatchPush, mission.ActionReleaseBranch, mission.ActionReleasePR, mission.ActionMainMerge}
	for _, action := range ordered {
		if slices.Contains(state.Contract.AllowedActions, action) {
			steps = append(steps, mission.SupervisionStep{Action: action, Target: target})
		}
	}
	engine := &cliSupervisorEngine{cmd: cmd, root: root, session: sessionID, jsonOutput: jsonOutput, governor: opts.GovernorReceipt, audit: opts.AuditReceipt, gitOpts: opts}
	result, err := mission.SuperviseAutoMission(cmd.Context(), mission.SupervisionPlan{MissionID: sessionID, ContractHash: state.ContractHash, MaxOperations: state.Contract.ResourceLimits.MaxOperations, Steps: steps, CompletionEvidence: state.Contract.CompletionEvidence, RequireLandedAncestry: gitMerge}, cliSupervisorStore{root: root, session: sessionID}, engine)
	if err != nil {
		return err
	}
	return printGTD(cmd, result, jsonOutput)
}

type missionTestsReceipt struct {
	HeadSHA string `json:"head_sha"`
	Status  string `json:"status"`
}

func validateMissionTestsReceipt(repository, receiptPath, headSHA string) error {
	if receiptPath == "" || !filepath.IsAbs(receiptPath) {
		return fmt.Errorf("auto mission blocked: tests_receipt_required")
	}
	repo, err := filepath.EvalSymlinks(repository)
	if err != nil {
		return err
	}
	info, err := os.Lstat(receiptPath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("auto mission blocked: tests_receipt_unsafe")
	}
	resolved, err := filepath.EvalSymlinks(receiptPath)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(repo, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return fmt.Errorf("auto mission blocked: tests_receipt_outside_scope")
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return err
	}
	var receipt missionTestsReceipt
	if json.Unmarshal(raw, &receipt) != nil || receipt.Status != "passed" || receipt.HeadSHA == "" || receipt.HeadSHA != headSHA {
		return fmt.Errorf("auto mission blocked: tests_receipt_stale_or_failed")
	}
	return nil
}

func currentMissionHead(root string) (string, error) {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("auto mission blocked: head_readback_failed: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func publishDependenciesReady(ctx context.Context, store *kanban.BacklogStore, itemID string) (bool, error) {
	reflection, err := kanban.ReflectGTDStore(ctx, store)
	if err != nil {
		return false, err
	}
	return !reflection.Result.Blocked[itemID] && !reflection.Result.Stale[itemID], nil
}

func authoritativeDispatchEvidence(ctx context.Context, store *kanban.BacklogStore, root, sessionID, itemID, cardID, lane, runID string, expectedRevision int64) (map[string]string, error) {
	item, err := kanban.LoadGTDItem(ctx, store, itemID)
	if err != nil {
		return nil, err
	}
	if item.SourceRevision != expectedRevision || item.CardID != cardID || cardID == "" || strings.TrimSpace(lane) == "" || strings.TrimSpace(runID) == "" {
		return nil, errors.New("auto mission blocked: stale_dispatch_snapshot")
	}
	record, err := store.LoadPure()
	if err != nil {
		return nil, err
	}
	picked, laneOwnerFree := false, true
	for _, card := range record.Items {
		if card.ID == cardID {
			picked = card.State == kanban.BacklogStatePicked
		}
	}
	for _, assignment := range record.Runtime.Assignments {
		if assignment.CardID == cardID && (assignment.RunID != runID || assignment.OwnerLabel != lane) {
			laneOwnerFree = false
		}
		if assignment.OwnerLabel == lane && assignment.CardID != cardID {
			laneOwnerFree = false
		}
	}
	lease, leaseErr := kanban.ReadSlotLease(root, lane)
	leaseHeld := leaseErr == nil && lease.Held() && lease.SessionID == sessionID && !lease.Expired(time.Now())
	values := map[string]string{
		"fresh_snapshot":  strconv.FormatInt(expectedRevision, 10),
		"lane_available":  strconv.FormatBool(laneOwnerFree),
		"lane_owner_free": strconv.FormatBool(laneOwnerFree),
		"lease":           strconv.FormatBool(leaseHeld),
		"picked":          strconv.FormatBool(picked),
	}
	if !picked || !laneOwnerFree || !leaseHeld {
		return values, errors.New("auto mission blocked: dispatch_authority_missing")
	}
	return values, nil
}

func runGoalMissionOperation(cmd *cobra.Command, sessionID string, jsonOutput bool, actionText, target string, recommend bool, gitOpts missionGitRunOptions) error {
	_ = recommend // compatibility-only: authority comes exclusively from receipts below.
	root := goalProjectRoot()
	sessionID = statusSessionID(sessionID)
	state, err := loadRequiredAutoMission(root, sessionID)
	if err != nil {
		return err
	}
	if state.State != mission.StateApproved && state.State != mission.StateRunning {
		return fmt.Errorf("auto mission is not approved")
	}
	if state.Contract == nil || state.ContractHash == "" {
		return fmt.Errorf("auto mission contract missing")
	}
	sealed := mission.SealedContract{Contract: *state.Contract, Version: state.Contract.PolicyVersion, Hash: state.ContractHash}
	action := mission.Action(actionText)
	evidence := map[string]string{}
	var revision int64
	store := newTodoStore()
	itemID := strings.TrimPrefix(target, "gtd:")
	var linkedCardID string
	dependenciesReady := true
	if (action == mission.ActionPublish || action == mission.ActionPick || action == mission.ActionDispatch) && strings.HasPrefix(target, "gtd:") {
		item, loadErr := kanban.LoadGTDItem(cmd.Context(), store, itemID)
		if loadErr != nil {
			return loadErr
		}
		revision = item.SourceRevision
		linkedCardID = item.CardID
		switch action {
		case mission.ActionPublish:
			dependenciesReady, err = publishDependenciesReady(cmd.Context(), store, itemID)
			if err != nil {
				return err
			}
			if !dependenciesReady {
				_ = persistMissionBlock(root, state, "dependencies_blocked")
				return errors.New("auto mission blocked: dependencies_blocked")
			}
			evidence = map[string]string{"approval": "explicit", "clarified": strconv.FormatBool(item.Disposition == kanban.DispositionAction && item.SourceTrusted), "fresh_snapshot": strconv.FormatInt(revision, 10), "organized": strconv.FormatBool(item.Status == kanban.GTDStatusOrganized)}
		case mission.ActionPick:
			if linkedCardID == "" {
				_ = persistMissionBlock(root, state, "published_card_missing")
				return fmt.Errorf("auto mission blocked: published_card_missing")
			}
			evidence = map[string]string{"approval": "explicit", "fresh_snapshot": strconv.FormatInt(revision, 10), "published": linkedCardID}
		case mission.ActionDispatch:
			if linkedCardID == "" || strings.TrimSpace(gitOpts.Lane) == "" || strings.TrimSpace(gitOpts.RunID) == "" {
				_ = persistMissionBlock(root, state, "dispatch_input_missing")
				return fmt.Errorf("auto mission blocked: dispatch_input_missing")
			}
			evidence, err = authoritativeDispatchEvidence(cmd.Context(), store, root, sessionID, itemID, linkedCardID, gitOpts.Lane, gitOpts.RunID, revision)
			if err != nil {
				_ = persistMissionBlock(root, state, err.Error())
				return err
			}
		}
	}
	var gitOwner mission.GitOwnerAdapter
	var gitHead string
	if action == mission.ActionCommit || action == mission.ActionLocalMerge {
		repository := actionRepository(gitOpts, action)
		integration := ""
		if action == mission.ActionLocalMerge {
			integration = repository
			if gitOpts.CardWorktree != "" {
				actualCardSHA, cardErr := currentMissionHead(gitOpts.CardWorktree)
				if cardErr != nil || gitOpts.CardSHA == "" || actualCardSHA != gitOpts.CardSHA {
					_ = persistMissionBlock(root, state, "card_sha_mismatch")
					return errors.New("auto mission blocked: card_sha_mismatch")
				}
			}
		}
		gitOwner = mission.GitOwnerAdapter{Effect: mission.GitEffect{Action: action, Repository: repository, IntegrationWorktree: integration, WorktreeBranch: gitOpts.Branch, ExplicitPaths: gitOpts.Paths, CommitMessage: gitOpts.Message, CardSHA: gitOpts.CardSHA, BaseSHA: gitOpts.BaseSHA, LeasePath: gitOpts.LeasePath, SessionID: sessionID}}
		head, branch, snapErr := gitOwner.Snapshot(cmd.Context())
		if snapErr != nil {
			return snapErr
		}
		revision = 1
		gitHead = head
		if action == mission.ActionCommit {
			if branch != gitOpts.Branch || len(gitOpts.Paths) == 0 {
				_ = persistMissionBlock(root, state, "git_scope_invalid")
				return fmt.Errorf("auto mission blocked: git_scope_invalid")
			}
			if err := validateMissionTestsReceipt(repository, gitOpts.TestsReceipt, head); err != nil {
				_ = persistMissionBlock(root, state, err.Error())
				return err
			}
			evidence = map[string]string{"fresh_snapshot": head, "scope": strings.Join(gitOpts.Paths, ","), "tests": head}
		}
		if action == mission.ActionLocalMerge {
			if err := gitOwner.ValidateIntegrationLease(); err != nil {
				_ = persistMissionBlock(root, state, err.Error())
				return err
			}
			evidence = map[string]string{"commit": gitOpts.CardSHA, "fresh_snapshot": head, "integration_lease": gitOpts.LeasePath}
		}
	}
	if action == mission.ActionBatchPush || action == mission.ActionReleaseBranch || action == mission.ActionReleasePR || action == mission.ActionMainMerge {
		delivery := mission.CapabilityDeliveryOwner{Provider: mission.UnsupportedDeliveryProvider{}, Action: action, Target: target}
		if _, deliveryErr := delivery.Snapshot(cmd.Context()); deliveryErr != nil {
			_ = persistMissionBlock(root, state, deliveryErr.Error())
			return deliveryErr
		}
	}
	snapshotHash := autoSnapshotHash(sealed.Hash, target, revision)
	if (action == mission.ActionCommit || action == mission.ActionLocalMerge) && (gitOpts.CardWorktree != "" || gitOpts.DevelopWorktree != "") {
		cardSHA := ""
		if action == mission.ActionLocalMerge {
			cardSHA = gitOpts.CardSHA
		}
		snapshotHash = autoGitSnapshotHash(sealed.Hash, target, action, actionRepository(gitOpts, action), gitHead, cardSHA)
	}
	snapshot := mission.MissionSnapshot{MissionID: sessionID, ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: snapshotHash, EvidenceRevision: revision, Evidence: evidence, State: mission.StateRunning, OperationsUsed: len(state.OperationIDs)}
	decision := mission.Decision{DecisionID: fmt.Sprintf("%s:%s:%d", action, target, revision), MissionID: sessionID, ContractHash: sealed.Hash, PolicyVersion: sealed.Version, SnapshotHash: snapshot.SnapshotHash, EvidenceRevision: revision, Action: action, Targets: []string{target}, RequiredEvidence: mission.RequiredEvidenceForAction(action), Evidence: evidence, ExpiresAt: time.Now().Add(time.Minute)}
	headRoot := root
	if action == mission.ActionCommit || action == mission.ActionLocalMerge {
		headRoot = actionRepository(gitOpts, action)
	}
	headSHA, err := currentMissionHead(headRoot)
	if err != nil {
		_ = persistMissionBlock(root, state, err.Error())
		return err
	}
	_, err = mission.LoadGovernanceReceipts(root, gitOpts.GovernorReceipt, gitOpts.AuditReceipt, mission.GovernanceExpectation{MissionID: sessionID, ContractHash: sealed.Hash, SnapshotHash: snapshot.SnapshotHash, Action: action, Targets: []string{target}, HeadSHA: headSHA, Now: time.Now()})
	if err != nil {
		_ = persistMissionBlock(root, state, err.Error())
		return err
	}
	receipt, err := mission.ValidateMissionDecision(sealed, snapshot, decision, time.Now())
	if err != nil {
		_ = persistMissionBlock(root, state, err.Error())
		return err
	}
	var owner gtdCLIOwner
	if action == mission.ActionPublish && strings.HasPrefix(target, "gtd:") {
		owner = gtdCLIOwner{readback: func() (bool, error) {
			item, err := kanban.LoadGTDItem(cmd.Context(), store, itemID)
			return item.CardID != "", err
		}, apply: func() error {
			ready, readyErr := publishDependenciesReady(cmd.Context(), store, itemID)
			if readyErr != nil || !ready {
				if readyErr != nil {
					return readyErr
				}
				return errors.New("auto mission blocked: dependencies_blocked")
			}
			result, err := kanban.EngageGTDItem(cmd.Context(), store, kanban.EngageInput{ItemID: itemID, Authorized: true, EvidenceFresh: true, DependenciesReady: dependenciesReady && ready, LaneAvailable: true, ResourcesAvailable: true})
			if err == nil && result.CardID == "" {
				return fmt.Errorf("publish blocked: %s", strings.Join(result.Reasons, ","))
			}
			return err
		}}
	} else if action == mission.ActionPick && linkedCardID != "" {
		owner = gtdCLIOwner{readback: func() (bool, error) {
			record, err := store.LoadPure()
			if err != nil {
				return false, err
			}
			for _, card := range record.Items {
				if card.ID == linkedCardID {
					return card.State == kanban.BacklogStatePicked, nil
				}
			}
			return false, nil
		}, apply: func() error {
			return store.Mutate(func(record *kanban.BacklogRecord) error {
				for i := range record.Items {
					if record.Items[i].ID == linkedCardID {
						if record.Items[i].State != kanban.BacklogStateQueued {
							return errors.New("auto mission: card_not_queued")
						}
						record.Items[i].State = kanban.BacklogStatePicked
						return nil
					}
				}
				return errors.New("auto mission: card_not_live")
			})
		}}
	} else if action == mission.ActionDispatch && linkedCardID != "" {
		owner = gtdCLIOwner{readback: func() (bool, error) {
			record, err := store.LoadPure()
			if err != nil {
				return false, err
			}
			for _, a := range record.Runtime.Assignments {
				if a.RunID == gitOpts.RunID && a.CardID == linkedCardID && a.OwnerLabel == gitOpts.Lane {
					return true, nil
				}
			}
			return false, nil
		}, apply: func() error {
			if _, err := authoritativeDispatchEvidence(cmd.Context(), store, root, sessionID, itemID, linkedCardID, gitOpts.Lane, gitOpts.RunID, revision); err != nil {
				return err
			}
			return kanban.RecordFactoryCardAssignment(root, gitOpts.RunID, linkedCardID, gitOpts.Lane, "")
		}}
	} else if action == mission.ActionCommit || action == mission.ActionLocalMerge {
		owner = gtdCLIOwner{readback: func() (bool, error) { return gitOwner.Readback(cmd.Context(), receipt.OperationID) }, apply: func() error { return gitOwner.Apply(cmd.Context(), receipt.OperationID) }}
	} else {
		_ = persistMissionBlock(root, state, "owner_adapter_unsupported")
		return fmt.Errorf("auto mission blocked: owner_adapter_unsupported")
	}
	receiptJSON, _ := json.Marshal(receipt)
	op := kanban.GTDOperation{OperationID: receipt.OperationID, MissionID: sessionID, Action: string(action), Target: target, SnapshotHash: snapshot.SnapshotHash, ReceiptJSON: receiptJSON}
	stored, err := kanban.ExecuteGTDOperation(cmd.Context(), store, op, owner)
	if err != nil {
		_ = persistMissionBlock(root, state, err.Error())
		return err
	}
	state.State = mission.StateRunning
	state.Snapshot = &snapshot
	state.LastBlocker = ""
	if !slices.Contains(state.OperationIDs, stored.OperationID) {
		state.OperationIDs = append(state.OperationIDs, stored.OperationID)
	}
	if err := mission.SaveAutoMission(root, *state); err != nil {
		return err
	}
	return printGTD(cmd, stored, jsonOutput)
}

func runGoalMissionRevoke(cmd *cobra.Command, sessionID string, jsonOutput bool) error {
	root := goalProjectRoot()
	state, err := loadRequiredAutoMission(root, statusSessionID(sessionID))
	if err != nil {
		return err
	}
	state.State = mission.StateRevoking
	state.LastBlocker = "revoked"
	if err := mission.SaveAutoMission(root, *state); err != nil {
		return err
	}
	return printGTD(cmd, state, jsonOutput)
}
func runGoalMissionResume(cmd *cobra.Command, sessionID string, jsonOutput bool) error {
	root := goalProjectRoot()
	state, err := loadRequiredAutoMission(root, statusSessionID(sessionID))
	if err != nil {
		return err
	}
	if state.State != mission.StateBlocked || state.Contract == nil {
		return fmt.Errorf("auto mission is not resumable")
	}
	state.State = mission.StateApproved
	state.LastBlocker = ""
	if err := mission.SaveAutoMission(root, *state); err != nil {
		return err
	}
	return printGTD(cmd, state, jsonOutput)
}

func newGoalCmd() *cobra.Command { return NewAutoMissionCommand() }

func runGoalAutoMission(cmd *cobra.Command, args []string, sessionFlag string, jsonOutput bool) error {
	root := goalProjectRoot()
	if root == "" {
		return fmt.Errorf("goal --auto: cannot resolve project root")
	}
	text := strings.TrimSpace(strings.Join(args, " "))
	if text == "" {
		return fmt.Errorf("goal --auto: mission must not be empty")
	}
	sessionID, warn := resolveArmSessionID(sessionFlag)
	if warn != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warn)
	}
	state := mission.AutoMission{SessionID: sessionID, Text: text, MissionMode: mission.ModeAuto, ProgressionMode: goal.DefaultProgressionMode, State: mission.StateDraft}
	if err := mission.SaveAutoMission(root, state); err != nil {
		return fmt.Errorf("goal --auto: %w", err)
	}
	if jsonOutput {
		return emitOK(cmd, true, map[string]any{"action": "mission-create", "session_id": sessionID, "mission_mode": mission.ModeAuto, "state": mission.StateDraft, "mission": text})
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "created auto mission for session %s (approval required): %s\n", sessionID, text)
	return nil
}

// misreadGoalVerb reports the registered goal verb a single word was most
// likely meant as, using each subcommand's SuggestFor list. The list is
// explicit on purpose: it names the word the refusal should SUGGEST, which a
// verb-shape rule could not.
//
// It no longer carries the whole burden of the single-word form. Card t890
// measured what the list alone lets through — `false`, `date`, `ls` — and
// bareWordNeedsDeclaration now demands a declaration for every undeclared bare
// word. A one-word condition such as `true` or `make` is not thrown away by
// that: it is written `cmd: true`, and the refusal says so.
func misreadGoalVerb(goalCmd *cobra.Command, word string) (string, bool) {
	for _, sub := range goalCmd.Commands() {
		for _, alias := range sub.SuggestFor {
			if strings.EqualFold(word, alias) {
				return sub.Name(), true
			}
		}
	}
	return "", false
}

// misreadGoalVerbError renders the refusal for a word read as a verb. It names
// the verb that was probably meant and the cmd: prefix for the rare case where
// the word really is the intended condition.
func misreadGoalVerbError(word, verb string) error {
	return fmt.Errorf(
		"goal: %q is not a goal verb, so it would be armed as a condition. "+
			"Did you mean \"moai goal %s\"? If %q really is the condition, "+
			"declare it: moai goal \"cmd: %s\"",
		word, verb, word, word)
}

// bareWordNeedsDeclaration reports whether a single bare argument must carry a
// `cmd:` / `model:` declaration before it can be armed as a condition.
//
// The two existing gates leave a gap between them (card t890). misreadGoalVerb
// only knows the SuggestFor aliases, and unrunnableCommandToken only refuses a
// first word that resolves to NOTHING — so a bare word that IS a real command
// passes both and arms silently. `false` then blocks every turn-end to the
// ceiling; `date` is satisfied at once and the goal ends without having meant
// anything. Neither is what one typed word intends.
//
// The rule is a declaration requirement, not a ban: the word still arms as
// `cmd: <word>`. That keeps it consistent with the adjacent refusal, which
// already demands the same prefix rather than guessing.
//
// It reads the bare single-word form ONLY — `goal arm <word>` states the intent
// explicitly, a prefixed word is already declared, and a multi-word condition
// was never the ambiguous shape. An empty argument falls through to the
// existing empty-condition refusal, which owns that case.
func bareWordNeedsDeclaration(word string) bool {
	if word == "" || len(strings.Fields(word)) != 1 {
		return false
	}
	return !conditionDeclarationPrefix.MatchString(word)
}

// bareWordDeclarationError renders the refusal for an undeclared bare word. It
// names both escapes because the word alone does not say which tier was meant:
// a command belongs behind cmd:, a claim about the conversation behind model:.
func bareWordDeclarationError(word string) error {
	return fmt.Errorf(
		"goal: %q is a single bare word, so it is ambiguous — it would be armed "+
			"as a shell command whether or not that is what you meant, and a "+
			"one-word command is as likely to be satisfied instantly as it is "+
			"to never exit 0. Declare which you mean: moai goal \"cmd: %s\" to "+
			"run it as a command, or moai goal \"model: %s\" for a claim about "+
			"the conversation",
		word, word, word)
}

// goalProjectRoot resolves the project root for goal state I/O (CLAUDE_PROJECT_DIR
// then cwd), reusing the same resolver the session CLI uses so the arm path and
// the stop-goal hook anchor to the same .moai/state/goal/ directory.
func goalProjectRoot() string {
	return resolveProjectDir()
}

// resolveArmSessionID resolves the session id the arm path keys on. When
// --session is provided it wins (deterministic testing + programmatic arming).
// Otherwise it resolves the real session id via resolveCurrentSessionID (the same
// side channel the stop-goal hook reads). Only when NO real session id resolves
// does it degrade to the writer_pid discriminator — and it returns a non-empty
// warning so the degrade is surfaced, never silent (REQ-GLE-033 / no silent
// pid-fallback). The WriterPidKey fallback stays valid ONLY here (REQ-GLE-008).
func resolveArmSessionID(sessionFlag string) (id, warn string) {
	if sessionFlag != "" {
		return sessionFlag, ""
	}
	if uuid, source, ok := resolveCurrentSessionID(); ok {
		// An id resolved from the per-process env var names THIS session by
		// construction, so the multi-session guard below does not apply to it:
		// concurrency cannot make it foreign. Only the side-channel path needs
		// the guard.
		if sessionIDSourceIsAuthoritative(source) {
			return uuid, ""
		}
		// Multi-session guard. The side-channel file
		// (.moai/state/current-session-id.txt) is per-project and unconditionally
		// overwritten on every SessionStart (internal/hook/session_start.go), so
		// under >=2 concurrent sessions in the same project dir it holds only the
		// most-recently-started session's id -- which may belong to a FOREIGN
		// session. Arming under a foreign id lands the goal under the wrong
		// session's state file and silently breaks the arm<->eval keying contract
		// (the authoritative id Claude Code passes to the stop-goal hook via stdin
		// would differ). Count registry entries whose CWD matches this project dir;
		// >=2 means the side-channel id is unreliable -- surface a warning so the
		// orchestrator re-arms with --session <authoritative-id>. Fail-open: a
		// registry read error degrades to the resolved id with no warning (an
		// advisory query must not block arming). Single-session (<=1 matching
		// entry) is the unchanged happy path -- no warning.
		if projectDir := resolveProjectDir(); projectDir != "" {
			if entries, qerr := session.QueryActiveWork(""); qerr == nil {
				matches := 0
				for _, e := range entries {
					if e.CWD == projectDir {
						matches++
					}
				}
				if matches >= 2 {
					return uuid, fmt.Sprintf(
						"%d sessions are active in this project directory (%s); the "+
							"side-channel id %q may belong to a concurrent session rather "+
							"than this one. Re-run with --session <authoritative-id> (the "+
							"source_session_id from the SessionStart context) to arm the "+
							"goal deterministically.", matches, projectDir, uuid)
				}
			}
		}
		return uuid, ""
	}
	pid := goal.WriterPidKey()
	return pid, fmt.Sprintf(
		"session.id not available from the runtime; goal armed under %q. The "+
			"'moai hook stop-goal' evaluator runs in a different process and may not "+
			"find a pid-keyed goal — register a session (moai session register) or "+
			"re-run once a session id is available.", pid)
}

// statusSessionID resolves the session id for the read/idempotent verbs (status,
// clear). It never warns (a status/clear against a pid key simply reports "no
// armed goal"); the no-silent-pid-fallback discipline is an ARM-path property.
func statusSessionID(sessionFlag string) string {
	if sessionFlag != "" {
		return sessionFlag
	}
	if uuid, _, ok := resolveCurrentSessionID(); ok {
		return uuid
	}
	return goal.WriterPidKey()
}

func runGoalArm(cmd *cobra.Command, args []string, sessionFlag string, jsonOutput bool) error {
	root := goalProjectRoot()
	if root == "" {
		return fmt.Errorf("goal arm: cannot resolve project root (set CLAUDE_PROJECT_DIR or run from the project directory)")
	}
	conditionText := strings.TrimSpace(strings.Join(args, " "))
	if conditionText == "" {
		return fmt.Errorf("goal arm: condition must not be empty")
	}

	// SPEC-INFINITE-GOAL-001 REQ-1/REQ-4: read the arm-time bound flags. The
	// sentinel -1 (default) means "omitted" → NewGoal's DefaultMaxTurns (30)
	// is preserved (AC-002 backward compat). 0 is the infinite entry point.
	maxTurnsFlag := cmd.Flags().Lookup("max-turns")
	maxDuration, _ := cmd.Flags().GetInt("max-duration")
	costCap, _ := cmd.Flags().GetInt("cost-cap")
	maxTurns := -1
	if maxTurnsFlag != nil {
		// Changed() distinguishes "user passed --max-turns 0" from "flag absent".
		if maxTurnsFlag.Changed {
			v, _ := strconv.Atoi(maxTurnsFlag.Value.String())
			maxTurns = v
		}
	}

	// AC-011 fail-closed (SPEC-INFINITE-GOAL-001 REQ-004 / D1): --max-turns 0
	// (infinite) REQUIRES --max-duration <seconds> as the real wall-clock bound.
	// --cost-cap is RECORDED-ONLY (enforcement is a follow-up, per flag help and
	// SPEC §D.5) and does NOT satisfy the real-bound requirement — so
	// cost-cap-alone is REJECTED. An infinite goal without a real bound would
	// run unbounded (the stagnation guard only fires on N consecutive no-change
	// turns with >=1 mechanical condition; real progress never triggers it, and
	// a model-only goal skips stagnation entirely per D2). No goal state file
	// is written on reject.
	if maxTurns == 0 && maxDuration <= 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(),
			"goal arm: --max-turns 0 (infinite) requires --max-duration <seconds> "+
				"as the real wall-clock bound. --cost-cap is recorded-only and does "+
				"NOT satisfy this requirement (it does not bound the goal). Re-run "+
				"with --max-duration <seconds>.")
		return fmt.Errorf("goal arm: --max-turns 0 requires --max-duration <seconds> as the real bound (cost-cap is recorded-only)")
	}

	sessionID, warn := resolveArmSessionID(sessionFlag)
	if warn != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warn)
	}

	cond := parseCondition(conditionText)
	// Arm-time gate: a mechanical condition that can only ever fail buys a goal
	// that blocks every turn-end to the ceiling. Refuse on positive evidence
	// only (both probes fail open) and write NO state file on refusal. The gate
	// is shared with the goal_arm MCP wrapper — see armTimeConditionGate.
	//
	// Returned, not also printed: the root command renders the error, and
	// printing it here too would show the user the same paragraph twice.
	if err := armTimeConditionGate(cmd.Context(), "goal arm", conditionText, cond); err != nil {
		return err
	}
	g := goal.NewGoal(sessionID, conditionText, []goal.Condition{cond})
	if maxTurns >= 0 {
		g.Ceiling.MaxTurns = maxTurns // 0 = infinite entry point (AC-001)
	}
	g.Ceiling.MaxDuration = maxDuration
	g.Ceiling.CostCap = costCap
	g.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := goal.SaveGoal(root, g); err != nil {
		return fmt.Errorf("goal arm: %w", err)
	}

	if jsonOutput {
		return emitOK(cmd, true, map[string]any{
			"action":         "arm",
			"session_id":     sessionID,
			"condition_type": string(cond.Type),
			"ceiling":        g.Ceiling.MaxTurns,
			"goal":           conditionText,
		})
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"armed goal for session %s (%s condition, ceiling %d turns): %s\n",
		sessionID, cond.Type, g.Ceiling.MaxTurns, conditionText)
	return nil
}

func runGoalStatus(cmd *cobra.Command, sessionFlag string, jsonOutput, showAll bool) error {
	root := goalProjectRoot()
	if root == "" {
		return fmt.Errorf("goal status: cannot resolve project root")
	}
	if showAll {
		return runGoalStatusAll(cmd, root, jsonOutput)
	}

	sessionID := statusSessionID(sessionFlag)
	if filepath.IsAbs(sessionID) || sessionID == "." || sessionID == ".." || strings.ContainsAny(sessionID, `/\`) {
		return fmt.Errorf("goal status: unsafe session id")
	}
	if mission.ValidateMissionSessionID(sessionID) == nil {
		auto, autoErr := mission.LoadAutoMission(root, sessionID)
		if autoErr != nil {
			return fmt.Errorf("goal status: %w", autoErr)
		}
		if auto != nil {
			if jsonOutput {
				out, err := json.Marshal(auto)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "session: %s\nmission_mode: auto\nstate: %s\nmission: %s\n", auto.SessionID, auto.State, auto.Text)
			return err
		}
	}
	g, err := goal.LoadGoal(root, sessionID)
	if err != nil {
		return fmt.Errorf("goal status: %w", err)
	}
	if g == nil {
		if jsonOutput {
			return emitOK(cmd, true, map[string]any{"action": "status", "session_id": sessionID, "armed": false})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "no armed goal for session %s\n", sessionID)
		return nil
	}
	if jsonOutput {
		out, marshalErr := json.MarshalIndent(g, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("goal status marshal: %w", marshalErr)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	}
	printGoalHuman(cmd, g)
	return nil
}

func runGoalStatusAll(cmd *cobra.Command, root string, jsonOutput bool) error {
	dir := filepath.Join(root, goal.StateDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no armed goals")
			return nil
		}
		return fmt.Errorf("goal status --all: %w", err)
	}
	var goals []*goal.Goal
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		g, loadErr := goal.LoadGoal(root, id)
		if loadErr != nil || g == nil {
			continue
		}
		goals = append(goals, g)
	}
	if jsonOutput {
		out, marshalErr := json.MarshalIndent(goals, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("goal status --all marshal: %w", marshalErr)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	}
	if len(goals) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no armed goals")
		return nil
	}
	for _, g := range goals {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "- %s [%s] %s\n", g.SessionID, g.Status, g.Goal)
	}
	return nil
}

func runGoalClear(cmd *cobra.Command, sessionFlag string, jsonOutput bool) error {
	root := goalProjectRoot()
	if root == "" {
		return fmt.Errorf("goal clear: cannot resolve project root")
	}
	sessionID := statusSessionID(sessionFlag)
	if filepath.IsAbs(sessionID) || sessionID == "." || sessionID == ".." || strings.ContainsAny(sessionID, `/\`) {
		return fmt.Errorf("goal clear: unsafe session id")
	}
	if mission.ValidateMissionSessionID(sessionID) == nil {
		if auto, _ := mission.LoadAutoMission(root, sessionID); auto != nil {
			if err := mission.ClearAutoMission(root, sessionID); err != nil {
				return fmt.Errorf("goal clear: %w", err)
			}
			if jsonOutput {
				return emitOK(cmd, true, map[string]any{"action": "clear", "session_id": sessionID, "cleared": true, "mission_mode": "auto"})
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "cleared auto mission for session %s\n", sessionID)
			return err
		}
	}

	existed := false
	if g, _ := goal.LoadGoal(root, sessionID); g != nil {
		existed = true
	}
	if err := goal.ClearGoal(root, sessionID); err != nil {
		return fmt.Errorf("goal clear: %w", err)
	}
	if jsonOutput {
		return emitOK(cmd, true, map[string]any{"action": "clear", "session_id": sessionID, "cleared": existed})
	}
	if existed {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cleared goal for session %s\n", sessionID)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "no armed goal for session %s\n", sessionID)
	}
	return nil
}

// printGoalHuman renders a goal's state as a readable block on stdout.
func printGoalHuman(cmd *cobra.Command, g *goal.Goal) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "session:    %s\n", g.SessionID)
	_, _ = fmt.Fprintf(out, "status:     %s\n", g.Status)
	_, _ = fmt.Fprintf(out, "goal:       %s\n", g.Goal)
	_, _ = fmt.Fprintf(out, "turns:      %d / %d\n", g.TurnsUsed, g.Ceiling.MaxTurns)
	_, _ = fmt.Fprintf(out, "mode:       %s\n", g.ProgressionMode)
	_, _ = fmt.Fprintf(out, "conditions: %d\n", len(g.Conditions))
	for i, c := range g.Conditions {
		if c.Type == goal.ConditionMechanical {
			_, _ = fmt.Fprintf(out, "  [%d] mechanical: %s (expect exit %d)\n", i, c.Cmd, c.ExpectExit)
		} else {
			_, _ = fmt.Fprintf(out, "  [%d] model: %s\n", i, c.Claim)
		}
	}
	if len(g.Progress) > 0 {
		_, _ = fmt.Fprintf(out, "progress:   %d entries (last: %s)\n", len(g.Progress), g.Progress[len(g.Progress)-1].Note)
	}
}

// runGoalRender (SPEC-GOAL-HTML-FLOW-001 REQ-GHF-004 / AC-GHF-001 / AC-GHF-010)
// resolves the live goal, renders the dashboard, and writes a self-contained
// HTML file to goal.HTMLPath (<root>/.moai/state/goal/<session>.html). It reuses
// statusSessionID (the read/idempotent resolver) + goal.LoadGoal. With no armed
// goal it exits non-zero, names the session id on stderr, and writes NO html.
func runGoalRender(cmd *cobra.Command, sessionFlag string, jsonOutput bool) error {
	root := goalProjectRoot()
	if root == "" {
		return fmt.Errorf("goal render: cannot resolve project root")
	}
	sessionID := statusSessionID(sessionFlag)
	g, err := goal.LoadGoal(root, sessionID)
	if err != nil {
		return fmt.Errorf("goal render: %w", err)
	}
	if g == nil {
		// AC-GHF-010: no armed goal → non-zero exit + stderr names session + NO html.
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"goal render: no armed goal for session %s; nothing to render\n", sessionID)
		return fmt.Errorf("goal render: no armed goal for session %s", sessionID)
	}

	// SPEC-GOAL-HTML-WIRING-001 REQ-WIRE-001 / AC-WIRE-001: load the last-produced
	// verdict sidecar (at-ceiling-only write by the stop-goal evaluator) and pass
	// it to RenderDashboard so the 5 CeilingVerdict sections render instead of
	// the "no verdict yet" placeholder. Fail-open to nil on missing/corrupt
	// sidecar (REQ-WIRE-003 / AC-WIRE-002): the placeholder path is preserved
	// byte-identical.
	v, _ := goal.LoadVerdict(root, sessionID)
	// SPEC-GOAL-HTML-WIRING-001 REQ-WIRE-009 / AC-WIRE-007: construct the
	// render-only ReArmContext from the already-landed SPEC-INFINITE-GOAL-001
	// state (factory.db resume row + post-/clear new-session goal file).
	// nil reArm → byte-identical base view per AC-GHF-007 / AC-WIRE-009.
	reArm := buildReArmContext(root, sessionID)
	raw, err := goal.RenderDashboardReArm(g, v, reArm)
	if err != nil {
		return fmt.Errorf("goal render: %w", err)
	}
	htmlPath := goal.HTMLPath(root, sessionID)
	if err := os.MkdirAll(filepath.Dir(htmlPath), 0o755); err != nil {
		return fmt.Errorf("goal render mkdir: %w", err)
	}
	if err := os.WriteFile(htmlPath, raw, 0o644); err != nil {
		return fmt.Errorf("goal render write %s: %w", htmlPath, err)
	}

	if jsonOutput {
		return emitOK(cmd, true, map[string]any{
			"action":     "render",
			"session_id": sessionID,
			"path":       htmlPath,
			"bytes":      len(raw),
		})
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"rendered goal dashboard for session %s to %s\n", sessionID, htmlPath)
	return nil
}

func init() {
	// SPEC amendment reachability: register the `moai goal` arming command under
	// rootCmd so it appears in `moai --help` and the arm→eval linkage is reachable.
	rootCmd.AddCommand(newGoalCmd())
}

// buildReArmContext constructs the render-only re-arm UI context from the
// already-landed SPEC-INFINITE-GOAL-001 state (SPEC-GOAL-HTML-WIRING-001
// REQ-WIRE-009 / AC-WIRE-007). It reads the pending resume row in factory.db
// (consume-only — the SPEC-INFINITE-GOAL-001 shape is untouched) and, when an
// EmbeddedGoal is present, scans the per-session goal files for a post-/clear
// new-session goal whose `Goal` text matches the embedded condition (the
// `rearmEmbeddedGoal` write signature). Returns nil when no EmbeddedGoal is
// present → the base view renders byte-identically (AC-WIRE-009). All steps are
// best-effort / fail-open: a missing/corrupt database or a scan miss leaves
// the corresponding ReArmContext field empty.
func buildReArmContext(root, sessionID string) *goal.ReArmContext {
	rec, present, _ := handoff.ReadPending(root)
	if !present || rec == nil || rec.EmbeddedGoal == nil {
		return nil
	}
	eg := rec.EmbeddedGoal
	ra := &goal.ReArmContext{
		EmbeddedCondition:   eg.Condition,
		EmbeddedMaxTurns:    eg.MaxTurns,
		EmbeddedMaxDuration: eg.MaxDuration,
		EmbeddedCostCap:     eg.CostCap,
		EmbeddedUnbounded:   eg.IsUnbounded(),
	}
	if eg.Condition != "" {
		ra.NewSessionID = findReArmSession(root, sessionID, eg.Condition)
	}
	return ra
}

// findReArmSession scans the per-session goal-state files for a session other
// than currentSessionID whose goal text matches embeddedCondition (the
// `rearmEmbeddedGoal` write signature: it reconstructs the goal with
// `Goal: eg.Condition`). Best-effort — returns "" on miss / error.
func findReArmSession(root, currentSessionID, embeddedCondition string) string {
	dir := filepath.Join(root, goal.StateDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if id == currentSessionID {
			continue
		}
		g, err := goal.LoadGoal(root, id)
		if err != nil || g == nil {
			continue
		}
		if g.Goal == embeddedCondition {
			return id
		}
	}
	return ""
}
