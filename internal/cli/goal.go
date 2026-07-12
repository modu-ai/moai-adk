// @MX:ANCHOR: [AUTO] moai goal CLI — the arming surface for the goal engine
// @MX:REASON: fan_in via the goal-engine loop — `goal arm` writes .moai/state/goal/<session-id>.json
// that the `moai hook stop-goal` evaluator loads on every turn-end. The arm↔eval shared-file keying
// (the SAME session id on both sides) is the load-bearing contract; a pid-keyed arm file would be
// unreachable by the hook (different PID), leaving the armed goal inert.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// trailingExitClause matches an optional trailing "exits <N>" / "exit <N>" suffix
// on a mechanical condition string (e.g. "go test ./... exits 0" → cmd="go test
// ./...", expect_exit=0). The clause is stripped so the shell never runs the
// literal words "exits 0" as command arguments.
var trailingExitClause = regexp.MustCompile(`(?i)^(.*\S)\s+exits?\s+(\d+)\s*$`)

// parseCondition classifies a single condition string by the EXPLICIT rule of
// REQ-GLE-032: a claim that references the conversation transcript becomes a
// model condition; any other string is a runnable shell command (mechanical).
//
// The discriminator is the literal token "transcript" (case-insensitive) — the
// canonical model form is "all AC rows show PASS in the transcript". A mechanical
// condition may carry a trailing "exits <N>" clause setting the expected exit
// code; absent it, expect_exit defaults to 0.
func parseCondition(s string) goal.Condition {
	s = strings.TrimSpace(s)
	if strings.Contains(strings.ToLower(s), "transcript") {
		return goal.Condition{Type: goal.ConditionModel, Claim: s}
	}
	cmd := s
	expect := 0
	if m := trailingExitClause.FindStringSubmatch(s); m != nil {
		cmd = strings.TrimSpace(m[1])
		expect, _ = strconv.Atoi(m[2])
	}
	return goal.Condition{Type: goal.ConditionMechanical, Cmd: cmd, ExpectExit: expect}
}

// newGoalCmd builds the `moai goal` cobra command tree: arm / status / clear.
// It REUSES the internal/goal engine (NewGoal / SaveGoal / LoadGoal / ClearGoal)
// and does NOT reimplement state / schema / prune logic. The `resume` verb is
// out of scope for this amendment (§D.6) and is deliberately NOT registered.
//
// IMPORTANT: this CLI surface MUST NOT invoke AskUserQuestion (subagent boundary,
// C-HRA-008). It returns exit codes + structured stdout only; the orchestrator
// owns all user interaction.
func newGoalCmd() *cobra.Command {
	var sessionFlag string
	var jsonOutput bool
	var showAll bool

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

Condition parsing: a runnable shell command (optionally suffixed "exits <N>")
becomes a mechanical condition; a claim that references the conversation
transcript becomes a model condition the orchestrator evaluates.`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			// Bare form: `goal "<condition>"` aliases `goal arm "<condition>"`.
			if len(args) == 0 {
				return c.Help()
			}
			return runGoalArm(c, args, sessionFlag, jsonOutput)
		},
	}
	cmd.PersistentFlags().StringVar(&sessionFlag, "session", "", "override the session id (default: resolve via 'moai session current')")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")

	armCmd := &cobra.Command{
		Use:          "arm <condition>",
		Short:        "Register and arm a goal for the active session",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runGoalArm(c, args, sessionFlag, jsonOutput)
		},
	}
	statusCmd := &cobra.Command{
		Use:          "status",
		Short:        "Print the active session's goal state",
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
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runGoalClear(c, sessionFlag, jsonOutput)
		},
	}

	cmd.AddCommand(armCmd, statusCmd, clearCmd)
	return cmd
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
	if uuid, _, ok := resolveCurrentSessionID(); ok {
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

	sessionID, warn := resolveArmSessionID(sessionFlag)
	if warn != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warn)
	}

	cond := parseCondition(conditionText)
	g := goal.NewGoal(sessionID, conditionText, []goal.Condition{cond})
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

func init() {
	// SPEC amendment reachability: register the `moai goal` arming command under
	// rootCmd so it appears in `moai --help` and the arm→eval linkage is reachable.
	rootCmd.AddCommand(newGoalCmd())
}
