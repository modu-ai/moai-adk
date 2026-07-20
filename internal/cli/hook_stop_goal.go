package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// stopGoalHookTimeout bounds a single mechanical-condition command. Goal cmds
// SHOULD be fast (prefer `go test -run <pattern>` over the full suite) because
// the eval runs at turn-end; this ceiling keeps a runaway cmd from stalling the
// Stop hook past its registered timeout (120s per settings.json.tmpl).
const stopGoalHookTimeout = 90 * time.Second

// realCmdRunner executes mechanical conditions via `sh -c`. It is the
// production CmdRunner for the goal evaluator; tests inject a fake.
type realCmdRunner struct{}

func (realCmdRunner) Run(ctx context.Context, cmd string) (int, string, error) {
	cctx, cancel := context.WithTimeout(ctx, stopGoalHookTimeout)
	defer cancel()
	cmdCtx := exec.CommandContext(cctx, "sh", "-c", cmd)
	out, err := cmdCtx.CombinedOutput()
	exit := 0
	if cmdCtx.ProcessState != nil {
		exit = cmdCtx.ProcessState.ExitCode()
	}
	if err != nil && exit == 0 {
		// A non-nil err with exit 0 is unusual; surface it but keep the exit.
		exit = 1
	}
	return exit, string(out), nil
}

// stopGoalCmd is the `moai hook stop-goal` verb. Registered under hookCmd in
// init() below. It evaluates the active session's goal on turn-end and emits a
// block decision until the conditions hold or the ceiling is reached.
//
// The hook MUST NOT call AskUserQuestion (REQ-GLE-014) — it emits structured
// JSON only; the orchestrator translates semi-autonomous checkpoint signals into
// AskUserQuestion rounds (agent-common-protocol.md § User Interaction Boundary).
var stopGoalCmd = &cobra.Command{
	Use:          "stop-goal",
	Short:        "Evaluate the active session goal on turn-end",
	Long:         "Reads Stop hook stdin JSON, loads the session's goal state, evaluates its conditions, and emits a block decision on stdout. Called from handle-stop-goal.sh.",
	SilenceUsage: true,
	RunE:         runStopGoalHook,
}

func init() {
	hookCmd.AddCommand(stopGoalCmd)
}

func runStopGoalHook(cmd *cobra.Command, _ []string) error {
	hookInput := readNormalizedHookInput()
	root := resolveHookProjectRoot()
	if root == "" {
		// No resolvable project root → nothing to evaluate; exit 0 silently.
		return nil
	}
	sessionID := hookInput.SessionID
	g, err := goal.LoadGoal(root, sessionID)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "stop-goal load: %v\n", err)
		return nil // fail-open: never block on a load error
	}
	if g == nil {
		// No armed goal → no block.
		return nil
	}
	// Snapshot source: a fresh shared-diagnostic-snapshot entry exactly matching
	// a Tier-1 condition command reuses the recorded exit code instead of
	// re-executing. The lookup is memoized (one key computation per turn-end)
	// and time-boxed per the Advisory-Check Discipline — on deadline exceed or
	// any error it degrades to command re-execution; correctness never depends
	// on the optimization.
	e := &goal.Eval{Runner: realCmdRunner{}, Snapshot: &verify.Source{ProjectRoot: root}}
	verdict, block := e.Evaluate(context.Background(), g)
	// Persist the updated goal (turns incremented, status set, progress appended).
	if err := goal.SaveGoal(root, g); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "stop-goal save: %v\n", err)
		// Continue: the verdict is still emitted; a save failure is best-effort.
	}
	if !block && !verdict.CeilingExit && !verdict.Stagnation && !verdict.Yielded {
		// All conditions satisfied — nothing to emit.
		return nil
	}
	out, err := json.Marshal(verdict)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "stop-goal marshal: %v\n", err)
		return nil
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
	// A block decision is emitted on stdout; the hook always exits 0 (per Claude
	// Code Stop semantics, the runtime honors stdout JSON `decision:"block"` on
	// exit 0; exit 2 is reserved for sync-phase-quality-gate-style blocking).
	return nil
}
