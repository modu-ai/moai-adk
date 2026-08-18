package cli

// @MX:NOTE: [AUTO] cc command switches LLM backend to Claude-only mode
// @MX:NOTE: [AUTO] Removes GLM env vars and resets team mode before launching Claude Code
// @MX:NOTE: [AUTO] Supports profile switching via CLAUDE_CONFIG_DIR
// @MX:NOTE: [AUTO] M6-S1 DDD: cc is a thin delegate-only entry point; print sites live in launcher.go::launchClaudeDefault

import (
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// findProjectRootFn is the function used to locate the project root.
// Tests may override this to return a temporary directory, preventing
// any file mutations from reaching the real project or home directory.
var findProjectRootFn = findProjectRoot

var ccCmd = &cobra.Command{
	Use:   "cc [-p profile] [-k [SPEC-ID] | -k --name <role> | -f [N] | -f worker-<n>] [-- claude-args...]",
	Short: "Launch Claude Code with Claude backend",
	Long: `Launch Claude Code with Claude backend.

This command:
  1. Removes GLM-specific environment variables from .claude/settings.local.json
  2. Resets team mode if it was enabled (glm or cg)
  3. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  4. Reads DO_CLAUDE_* settings and converts them to CLI flags
  5. Launches Claude Code via exec (replaces current process)

Flags:
  -p, --profile <name>          Use a named Claude profile (~/.moai/claude-profiles/<name>/)
  --permission-mode <mode>      Set permission mode (default, acceptEdits, plan, auto, bypassPermissions, dontAsk)
  -b, --bypass                  Shorthand for --permission-mode bypassPermissions
  -c, --continue                Continue previous session
  -m, --model <model>           Override model selection
  -w, --worktree [name]         Launch in an isolated git worktree (.claude/worktrees/<name>/);
                                name omitted = auto-generated (same as claude --worktree)
      --spawn                   Run this command in a new tmux window instead of
                                replacing the current session (requires tmux)
  --chrome / --no-chrome        Toggle Chrome MCP

Kanban Mode:
  -k, --kanban [SPEC-ID]       Enter as the LEAD of a kanban run. Seeds a
                                plan -> run -> sync chain in this
                                session. The optional SPEC-ID ties the run to a
                                SPEC. The lead drives the whole chain; three
                                companion sessions are launched by hand.
  -k --name <role>             Enter as a COMPANION of an existing kanban run.
                                Joins the run without seeding a chain. The three
                                roles are: planner, runner, syncer. A role
                                name held by a live session is bumped to
                                the next free number (planner-1,
                                planner-2, ...).

Factory Mode (dedicated -f entry):
  -f, --factory [N]            Enter as the LEAD of a factory run with N
                                numbered workers; N omitted = one worker
                                (worker-1), grown afterwards with the
                                incremental form below. The lead routes
                                operator-picked cards to free workers over
                                cross-session messages — A/B-class cards
                                wholesale to one worker, C-class cards a
                                serial 3-stage path.
  -f worker-<n>                Launch exactly one additional worker — worker
                                n — and connect it to the leader socket of the
                                running factory. A number whose label is held by a
                                live session is bumped to the next free number.
  -k <N> / -k <N> --name worker-<i>
                                The v1.2.0 unified -k factory shapes, still
                                valid: -k N is the lead of an N-worker run,
                                -k N --name worker-<i> is worker i of it (a
                                bare -k --name worker-<i> defaults to 8).
                                One entry token per launch: -k and -f
                                together is an error.

  Genealogy: the pre-3.1 "factory" flag (-f/--factory) was RENAMED to
  -k/--kanban in #1513 (7f61332ef) and now drives the three-role kanban chain
  above. -f briefly returned as the factory fan-out flag and was RETIRED
  (v1.2.0) in favor of '-k <N>'; t118 (v3.1.1) revived it as the dedicated
  factory entry — the kanban chain keeps -k, the factory gets -f.

Permission Modes:
  default            Ask permissions for file edits and commands
  acceptEdits        Auto-accept file edits, ask for commands (project default)
  plan               Read-only exploration and planning
  auto               Background classifier checks actions (requires Team plan + Sonnet/Opus 4.6)
  bypassPermissions  Skip all checks (isolated environments only)
  dontAsk            Only pre-approved tools

Examples:
  moai cc                              # Default profile, launch Claude
  moai cc -p work                      # Use 'work' profile
  moai cc --permission-mode auto       # Launch with auto mode
  moai cc -p work -- --print           # Profile + pass-through args to Claude
  moai cc -w feat-login                # Launch in isolated worktree 'feat-login'
  moai cc -w                           # Launch in auto-named isolated worktree
  moai cc -w feat-login --spawn        # Teammate session in a new tmux window
  moai cc -k                           # Kanban lead: seeds the plan->run->sync chain
  moai cc -k SPEC-AUTH-001             # Kanban lead tied to SPEC-AUTH-001
  moai cc -k --name planner           # Kanban companion: joins as the planner
  moai cc -f                           # Factory lead: one worker (worker-1)
  moai cc -f 4                         # Factory lead: announces worker-1..worker-4
  moai cc -f worker-2                  # Add worker 2 to the running factory
  moai glm -f worker-3                 # Same lane on the GLM backend`,
	GroupID:            "launch",
	DisableFlagParsing: true,
	RunE:               runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude, then launches Claude Code.
func runCC(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	// --spawn re-issues this same command in a new tmux window instead of
	// replacing the current process. It runs before any settings mutation so a
	// failed spawn leaves the environment untouched; the spawned `moai cc`
	// performs the mutations itself.
	if spawnArgs, spawn := stripSpawnFlag(args); spawn {
		return spawnLaunch(cmd.OutOrStdout(), "cc", spawnArgs)
	}

	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		return err
	}
	// The unified entry parse (t118): parseLauncherEntry covers BOTH entry
	// tokens — the -k shapes of SPEC-FACTORY-BOOTSTRAP-001 (kanban membership,
	// role disambiguated by --name, per the §A.2 truth table REQ-FB-001 /
	// REQ-FB-002) plus the v1.2.0 factory shapes (-k N), and the revived
	// dedicated -f surface (bare -f / -f N / -f worker-<n>). Parsed after
	// --spawn is stripped (a spawned session re-issues this command and must
	// carry the token through) and before worktree handling (so an entry token
	// can never be mistaken for a -w value). The environment mutation is
	// restored on every return path, including error.
	entry, err := parseLauncherEntry(filteredArgs)
	if err != nil {
		return err
	}
	filteredArgs = entry.Rest
	label, isCompanion := parseCompanionLabel(filteredArgs)
	factoryLabel, isFactoryWorker := parseFactoryWorkerLabel(filteredArgs)
	switch resolveFactoryBranch(entry.FactoryEnabled, isFactoryWorker) {
	case factoryBranchLead:
		// The operator-supplied `leader-<run-id>` name is adopted on the same
		// terms as the kanban lead (see kanban.go leadRunID) — the run id, the
		// session name, and the worker commands the notice prints stay on one
		// run.
		leadLabel, _ := parseLeadLabel(filteredArgs)
		defer enterFactoryLeadMode(entry.FactoryWorkers, leadLabel)()
		recordKanbanSession(entry.Spec, kanban.BackendClaude, kanban.RoleLead)
		filteredArgs = append(filteredArgs, leadNameArgs(filteredArgs)...)
		settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
		if len(settingsFlag) > 0 {
			filteredArgs = append(filteredArgs, settingsFlag...)
		}
		defer settingsCleanup()
	case factoryBranchWorker:
		// A number held by a live session is bumped to the next free one, and
		// the bumped value must reach the backend argv — the session name is
		// the address the lead dispatches to.
		finalLabel := resolveFactoryWorkerName(launchProjectRoot(), factoryLabel, cmd.ErrOrStderr())
		filteredArgs = replaceNamedLabel(filteredArgs, factoryLabel, finalLabel)
		defer enterFactoryWorkerMode(finalLabel, entry.FactoryWorkers)()
		recordKanbanSession(entry.Spec, kanban.BackendClaude, kanban.RoleWorker)
		settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
		if len(settingsFlag) > 0 {
			filteredArgs = append(filteredArgs, settingsFlag...)
		}
		defer settingsCleanup()
	}
	if !entry.FactoryEnabled {
		switch resolveKanbanBranch(entry.KanbanEnabled, isCompanion) {
		case kanbanBranchLead:
			// An operator-supplied `leader-<run-id>` name carries the run id this
			// session already belongs to — a relaunch of an existing lead, or a
			// board whose companions are already open. Adopting it is what keeps
			// the session name, MOAI_KANBAN_ID, the leader socket path, and the
			// companion commands the SessionStart notice prints on one run.
			leadLabel, _ := parseLeadLabel(filteredArgs)
			defer enterKanbanMode(entry.Spec, leadLabel)()
			recordKanbanSession(entry.Spec, kanban.BackendClaude, kanban.RoleLead)
			// A lead launched bare has only an AI-generated title, which claude
			// discards on /clear; naming it explicitly is what survives.
			filteredArgs = append(filteredArgs, leadNameArgs(filteredArgs)...)
			settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
			if len(settingsFlag) > 0 {
				filteredArgs = append(filteredArgs, settingsFlag...)
			}
			defer settingsCleanup()
		case kanbanBranchCompanion:
			// A label held by a live session is bumped to the next free number
			// for the role, and the bumped value must reach the backend argv —
			// the session name is the address the lead dispatches to.
			finalLabel := resolveCompanionName(launchProjectRoot(), label, cmd.ErrOrStderr())
			filteredArgs = replaceNamedLabel(filteredArgs, label, finalLabel)
			defer enterKanbanCompanionMode(finalLabel)()
			recordKanbanSession(entry.Spec, kanban.BackendClaude, companionRole(finalLabel))
			settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
			if len(settingsFlag) > 0 {
				filteredArgs = append(filteredArgs, settingsFlag...)
			}
			defer settingsCleanup()
		}
	}
	// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a: validate absolute-path -w values
	// BEFORE normalizeWorktreeFlag so out-of-prefix paths are rejected with a
	// clear error (AC-WES-010c) and L2 (~/.moai/worktrees/) paths are accepted
	// (AC-WES-010a). normalizeWorktreeFlag remains the owner of short-name
	// token normalization (AC-WES-010b).
	if err := resolveWorktreeL2Path(filteredArgs); err != nil {
		return err
	}
	filteredArgs = normalizeWorktreeFlag(filteredArgs)
	return unifiedLaunch(profileName, "claude", filteredArgs)
}
