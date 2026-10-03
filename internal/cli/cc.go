package cli

// @MX:NOTE: [AUTO] cc command switches LLM backend to Claude-only mode
// @MX:NOTE: [AUTO] Removes GLM env vars and resets team mode before launching Claude Code
// @MX:NOTE: [AUTO] Supports profile switching via CLAUDE_CONFIG_DIR
// @MX:NOTE: [AUTO] M6-S1 DDD: cc is a thin delegate-only entry point; print sites live in launcher.go::launchClaudeDefault

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// findProjectRootFn is the function used to locate the project root.
// Tests may override this to return a temporary directory, preventing
// any file mutations from reaching the real project or home directory.
var findProjectRootFn = findProjectRoot

var ccCmd = &cobra.Command{
	Use:   "cc [-p profile] [-f | -l] [-- claude-args...]",
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
      --branch <existing>         With -w <name>: create the worktree checked out at an
                                EXISTING branch (e.g. moai cc -w develop --branch develop
                                for the gitflow integration worktree) instead of a new
                                branch. The branch must already exist — this flag never
                                creates one. The tree is registered in
                                .moai/state/worktrees.json for worktree tooling.
      --spawn                   Run this command in a new tmux window instead of
                                replacing the current session (requires tmux)
  --chrome / --no-chrome        Passed through to Claude Code unchanged; the
                                launcher adds neither (Chrome stays attachable
                                via /chrome unless you pass --no-chrome)

Factory Mode (dedicated -f and -l entries):
  -f, --factory                Enter as the LEADER of a factory run. The flag
                                takes no argument: lanes join one at a time
                                with -l. The leader routes operator-picked
                                cards to free lanes over cross-session
                                messages — each card goes WHOLE to one lane,
                                which carries it through plan -> run -> sync
                                in-session.
  -l, --lane                   Join the running factory as a LANE: the next
                                free lane-<n> label is claimed for this
                                session. The flag takes no argument. If the
                                run's record is missing or retired while a
                                live leader session exists, the join verifies
                                that leader (pid + process-start) and
                                restores its run, so the lane still lands on
                                the live factory.
  --leader <name>              With -l: which leader session the
                                record-absence verification targets
                                (default: leader). A legacy spelling of the
                                leader name is refused.
  One entry token per launch: -f beside -l is an error. A lane number
  held by a live session is bumped to the next free number.
  Legacy role and label spellings (the pre-rename nouns, any letter case)
  are refused — the error names the canonical lane-<n> label.

  Genealogy: the pre-3.1 "factory" flag (-f/--factory) was RENAMED to -k
  in #1513 (7f61332ef). -f briefly returned as the factory fan-out flag and
  was RETIRED (v1.2.0) in favor of '-k <N>'; t118 (v3.1.1) revived it as the
  dedicated factory entry, and -k itself is now retired: a launch carrying
  it is refused with one line naming -f and -l.

Permission Modes:
  default            Ask permissions for file edits and commands
  acceptEdits        Auto-accept file edits, ask for commands (moai init default)
  plan               Read-only exploration and planning
  auto               Background classifier checks actions (requires a supported model and plan; see Claude Code permission-modes docs)
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
  moai cc -w develop --branch develop  # Integration worktree on the existing develop branch
  moai cc -f                          # Factory leader: one lane (lane-1)
  moai cc -l                           # Join the running factory as the next free lane
  moai glm -l                          # Same lane on the GLM backend`,
	GroupID:            "launch",
	DisableFlagParsing: true,
	RunE:               runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude, then launches Claude Code.
func runCC(cmd *cobra.Command, args []string) error {
	return runClaudeEntry(cmd, args, "cc", "claude", factory.BackendClaude, unifiedLaunch)
}

func runClaudeEntry(cmd *cobra.Command, args []string, commandName, mode, backend string, launch func(string, string, []string) error) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	// SPEC-CODEX-DEBUG-MODE-001 REQ-004: observe-only debug detection — the
	// token STAYS in the child arguments untouched (the child's native -d
	// handling is preserved); only the launcher-side trace activates. The
	// scan mirrors the --help scan above: post--- tokens are the child's and
	// are never inspected (REQ-002's scoping).
	debugRequested := launcherDebugRequested(args)
	var debugTiming *factoryLaunchTiming
	if debugRequested {
		debugTiming = &factoryLaunchTiming{debug: true}
	}
	endEntry := debugTiming.beginDebug(launchStepEntryParse, "")

	if err := guardCGLaunchMode(mode); err != nil {
		endEntry()
		return err
	}

	// --spawn re-issues this same command in a new tmux window instead of
	// replacing the current process. It runs before any settings mutation so a
	// failed spawn leaves the environment untouched; the spawned `moai cc`
	// performs the mutations itself.
	if spawnArgs, spawn := stripSpawnFlag(args); spawn {
		endEntry()
		if err := refuseBadEntryBeforeSpawn(spawnArgs); err != nil {
			return err
		}
		return spawnLaunch(cmd.OutOrStdout(), commandName, spawnArgs)
	}

	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		endEntry()
		return err
	}
	// The entry parse (t118): parseLauncherEntry refuses the retired -k
	// spelling, then covers the factory surface (-f for the leader, -l / --lane
	// for a lane). Parsed after --spawn is stripped (a spawned session re-issues
	// this command and must carry the token through) and before worktree
	// handling (so an entry token can never be mistaken for a -w value). The
	// environment mutation is restored on every return path, including error.
	entry, err := parseLauncherEntry(filteredArgs)
	if err != nil {
		endEntry()
		return err
	}
	filteredArgs = entry.Rest
	endEntry()
	factoryLabel, isFactoryLane := parseFactoryLaneLabel(filteredArgs)
	switch resolveFactoryBranch(entry.FactoryEnabled, isFactoryLane) {
	case factoryBranchLeader:
		// The operator-supplied `leader-<run-id>` name is adopted (see
		// factory_launch_helpers.go leaderRunID) — the run id, the session name, and the lane
		// commands the notice prints stay on one run.
		leaderLabel, _ := parseLeaderLabel(filteredArgs)
		restoreFactory := enterFactoryLeaderMode(entry.FactoryLanes, leaderLabel)
		defer restoreFactory()
		restoreRun, runErr := enterSelectedFactoryRun(launchProjectRoot(), entry.FactoryRun, false, nil)
		if runErr != nil {
			return runErr
		}
		defer restoreRun()
		if err := recordFactoryRunStart(launchProjectRoot(), os.Getenv(config.EnvFactoryRunID), backend, entry.Spec, factoryDeclaredLanes(entry)); err != nil {
			return fmt.Errorf("record factory run: %w", err)
		}
		defer exportFactoryLaunchFacts(entry.Spec, backend)()
		var leaderName string
		filteredArgs, leaderName = appendLeaderName(filteredArgs, launchProjectRoot(), cmd.ErrOrStderr())
		defer exportLeaderSessionName(leaderName)()
		endSettings := debugTiming.beginDebug(launchStepSettingsPrep, "")
		settingsFlag, settingsCleanup := prepareFactorySettings(profileName, filteredArgs)
		endSettings()
		if len(settingsFlag) > 0 {
			filteredArgs = append(filteredArgs, settingsFlag...)
		}
		defer settingsCleanup()
	case factoryBranchLane:
		// The shared lane join: record-absence refusal falls back to
		// verified leader discovery + resume, then re-enters the gate
		// (SPEC-FACTORY-LANE-JOIN-SOCKET-001). One implementation for cc,
		// glm, and the codex twin (REQ-010). Under debug mode the join gate
		// and active-run resolution record through the launch's collector —
		// the same t1378 steps the codex lane launch names (REQ-013's
		// shared vocabulary).
		restoreRun, runErr := enterFactoryLaneRun(launchProjectRoot(), entry.FactoryRun, entry.FactoryLead, debugTiming)
		if runErr != nil {
			return runErr
		}
		defer restoreRun()
		// A number held by a live session is bumped to the next free one, and
		// the bumped value must reach the backend argv — the session name is
		// the address the leader dispatches to.
		endClaim := debugTiming.beginDebug(factoryStepLaneClaim, "")
		finalLabel, claimErr := resolveFactoryLaneName(launchProjectRoot(), factoryLabel, backend, entry.FactoryAutoNumber, cmd.ErrOrStderr())
		endClaim()
		if claimErr == nil {
			debugTiming.annotateDetail("label=" + finalLabel)
		}
		if claimErr != nil {
			return claimErr
		}
		filteredArgs = replaceNamedLabel(filteredArgs, factoryLabel, finalLabel)
		defer enterFactoryLaneMode(finalLabel, entry.FactoryLanes, entry.ClearPolicy, laneDispatchSelection(entry))()
		defer exportFactoryLaunchFacts(entry.Spec, backend)()
		// REQ-SD-020: the relaunch policy turns this launcher into the
		// supervising loop — it stays the parent, leases the next card,
		// starts one interactive session per card, and never exec's in
		// place (design.md §6). The stamps above are live for the loop's
		// own `next` calls and reach every child through the environment.
		if entry.ClearPolicy == config.FactoryClearPolicyRelaunch {
			endSettings := debugTiming.beginDebug(launchStepSettingsPrep, "")
			settingsFlag, settingsCleanup := prepareFactorySettings(profileName, filteredArgs)
			endSettings()
			defer settingsCleanup()
			if len(settingsFlag) > 0 {
				filteredArgs = append(filteredArgs, settingsFlag...)
			}
			if debugRequested {
				// The relaunch loop replaces the one-shot launch: the dump is
				// this launcher's pre-exec trace, printed before the loop's
				// first session handoff (REQ-009's cc/glm form).
				debugTiming.debugDump(cmd.ErrOrStderr())
			}
			return runFactoryLaneRelaunch(cmd, finalLabel, filteredArgs, entry.FactoryRun, entry.FactoryLead)
		}
		endSettings := debugTiming.beginDebug(launchStepSettingsPrep, "")
		settingsFlag, settingsCleanup := prepareFactorySettings(profileName, filteredArgs)
		endSettings()
		if len(settingsFlag) > 0 {
			filteredArgs = append(filteredArgs, settingsFlag...)
		}
		defer settingsCleanup()
	}
	// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a: validate absolute-path -w values
	// BEFORE normalizeWorktreeFlag so out-of-prefix paths are rejected with a
	// clear error (AC-WES-010c) and L2 (~/.moai/worktrees/) paths are accepted
	// (AC-WES-010a). normalizeWorktreeFlag remains the owner of short-name
	// token normalization (AC-WES-010b).
	endWt := debugTiming.beginDebug(launchStepWorktree, "")
	if err := resolveWorktreeL2Path(filteredArgs, cmd.ErrOrStderr()); err != nil {
		endWt()
		return err
	}
	// Card t295: `-w <name> --branch <existing>` materializes the worktree at
	// the existing branch before launch; the flag tokens are stripped so the
	// backend re-enters the tree that now exists. No-op without --branch.
	filteredArgs, err = resolveWorktreeExistingBranch(filteredArgs, cmd.ErrOrStderr())
	endWt()
	if err != nil {
		return err
	}
	// A tree another live session is anchored in gets no second writer. The
	// check only reads; Claude Code writes its own lock on entry.
	if err := ccWorktreeWriterPrecheck(filteredArgs); err != nil {
		return err
	}
	// SPEC-HANDOFF-NEUTRAL-001 REQ-HN-006: launcher-entry backfill — applied
	// only now that the admission above succeeded, so a refused launch leaves
	// the tree untouched (audit F0, sync-audit-opus.md).
	seedAdmittedWorktreeHooks(filteredArgs, cmd.ErrOrStderr())
	endHandoff := debugTiming.beginDebug(launchStepLaunchHandoff, "")
	filteredArgs = normalizeWorktreeFlag(filteredArgs)
	endHandoff()
	if debugRequested {
		// REQ-009's cc/glm form: the backend launch replaces this process
		// (exec), so the dump is the launcher's last output, printed before
		// the launch call.
		debugTiming.debugDump(cmd.ErrOrStderr())
	}
	return launch(profileName, mode, filteredArgs)
}
