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
  -w, --worktree [name]         Launch in an isolated git worktree (.claude/worktrees/<name>/
                                when created new; an existing tree of the same name under
                                <MoaiHome>/worktrees/projects/<project-key> is re-entered);
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

	// REQ-SCV-012 (SPEC-SESSION-CC-VERSION-002): the option model the resume
	// interpreters read is derived once per process from the claude binary on
	// PATH. Every derivation failure degrades silently to the compile-time
	// snapshot and changes no exit status; the scan below stays a pure argv
	// walk over package state.
	refreshActiveClaudeOptionModel()

	// REQ-SCV-009 (SPEC-SESSION-CC-VERSION-001): a --resume token with no
	// session id is a broken launch — refuse before any launch side effect,
	// including the --spawn window. Pure argv scan: no environment, no
	// filesystem, no lane claim.
	if err := validateResumeArgs(args); err != nil {
		endEntry()
		return err
	}

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
		// SPEC-LAUNCHER-ARGV-001 REQ-LA-004: the anchor guard judges the launch
		// view BEFORE any session mutation — the role marker, the run record, the
		// settings preparation and the worktree materialization all wait behind
		// it. The override flag is the launcher's own option, consumed here; the
		// assignment (never :=) keeps the function-level filteredArgs the one the
		// rest of the entry sees.
		var anchorOverride bool
		filteredArgs, anchorOverride = consumeAllowCardAnchor(filteredArgs)
		if err := enforceLeaderAnchorGuard(backend, leaderLaunchDir(), launchProjectRoot(), resolveLaunchView(filteredArgs), anchorOverride, cmd.ErrOrStderr()); err != nil {
			return err
		}
		// The operator-supplied `leader-<run-id>` name is adopted (see
		// factory_launch_helpers.go leaderRunID) — the run id, the session name, and the lane
		// commands the notice prints stay on one run.
		priorRunID := os.Getenv(config.EnvFactoryRunID)
		leaderLabel, _ := parseLeaderLabel(filteredArgs)
		restoreFactory := enterFactoryLeaderMode(entry.FactoryLanes, leaderLabel)
		defer restoreFactory()
		restoreRun, runErr := enterSelectedFactoryRun(launchProjectRoot(), entry.FactoryRun, false, nil)
		if runErr != nil {
			return runErr
		}
		defer restoreRun()
		// A resume-shaped leader start whose id was freshly minted re-adopts
		// the prior run instead (card t1707): the run id survives only in the
		// launch environment, so the resume launch would otherwise mint a new
		// run over the live one and the dead-owner reconciliation would retire
		// the prior run's records behind it. Shared with the glm leader, which
		// mirrors this entry exactly.
		adopted, adoptErr := leaderEntryRunAdoption(launchProjectRoot(), filteredArgs, leaderLabel, priorRunID, entry.FactoryRun, cmd.ErrOrStderr())
		if adoptErr != nil {
			return fmt.Errorf("re-adopt factory run: %w", adoptErr)
		}
		if err := recordLeaderRunStart(launchProjectRoot(), adopted, backend, entry.Spec, factoryDeclaredLanes(entry)); err != nil {
			return err
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
		// Card t1554 (operator decision): the relaunch policy's lease chain
		// is removed — card consumption is the unified `moai todo --auto`
		// engine's alone. The flag stays accepted and degrades to the
		// one-shot lane session below, which consumes the queue itself.
		if entry.ClearPolicy == config.FactoryClearPolicyRelaunch {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), factoryRelaunchSupersededNote)
		}
		endSettings := debugTiming.beginDebug(launchStepSettingsPrep, "")
		settingsFlag, settingsCleanup := prepareFactorySettings(profileName, filteredArgs)
		endSettings()
		// REQ-SCV-008: the lane-join child argv is assembled by the one pure
		// assembler (lane_resume.go) — name pair, settings, and pass-through
		// tokens, in a deterministic order.
		filteredArgs = laneJoinChildArgv(filteredArgs, finalLabel, settingsFlag)
		defer settingsCleanup()
	}
	// SPEC-LAUNCHER-ARGV-001: the launch argv is resolved ONCE here (launch_view.go).
	// The readers below receive the stage values the view carries — none of them
	// re-reads raw arguments — and the view's LaunchArgs is the argv handed to the
	// launch seam.
	view := resolveLaunchView(filteredArgs)
	entryDir, err := existingWorktreeEntryDir(view.LaunchArgs)
	if err != nil {
		return err
	}
	// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a: validate absolute-path -w values
	// BEFORE the token normalization so out-of-prefix paths are rejected with a
	// clear error (AC-WES-010c) and L2 (~/.moai/worktrees/) paths are accepted
	// (AC-WES-010a). The normalization stays the owner of short-name token
	// rewriting (AC-WES-010b).
	endWt := debugTiming.beginDebug(launchStepWorktree, "")
	if err := validateWorktreeL2Value(view.StageA, cmd.ErrOrStderr()); err != nil {
		endWt()
		return err
	}
	// Card t295: `-w <name> --branch <existing>` materializes the worktree at
	// the existing branch before launch; the flag tokens are already stripped
	// from the view's LaunchArgs so the backend re-enters the tree that now
	// exists. No-op without --branch.
	err = view.materializeBranch(cmd.ErrOrStderr())
	endWt()
	if err != nil {
		return err
	}
	if view.Branch.Present {
		// Materialization may have created the home-root tree to enter.
		entryDir, err = existingWorktreeEntryDir(view.LaunchArgs)
		if err != nil {
			return err
		}
	}
	// A tree another live session is anchored in gets no second writer. The
	// check only reads; Claude Code writes its own lock on entry.
	if err := ccWorktreeWriterPrecheck(view.StageB); err != nil {
		return err
	}
	// SPEC-HANDOFF-NEUTRAL-001 REQ-HN-006: launcher-entry backfill — applied
	// only now that the admission above succeeded, so a refused launch leaves
	// the tree untouched (audit F0, sync-audit-opus.md).
	seedAdmittedWorktreeHooksFor(view.StageB, cmd.ErrOrStderr())
	endHandoff := debugTiming.beginDebug(launchStepLaunchHandoff, "")
	// SPEC-WORKTREE-HOME-ROOT-001 M2 (REQ-WHR-006), entry form repaired by
	// M8R (card t1640 review P1-2): a short name naming an existing home-root
	// tree ENTERS that tree — the launcher starts the backend inside it and
	// hands it no worktree selection. The argv rewrite this used to be is
	// measured-rejected: PM-3 (operator ruling bd72) recorded claude 2.1.303
	// exiting 1 on absolute --worktree values at argument validation. Every
	// other short name passes through unchanged (branch Y — the launcher
	// creates nothing).
	// The view owns the launch argv (SPEC-LAUNCHER-ARGV-001): what the exec
	// boundary receives is the view's normalized LaunchArgs, exactly as the
	// old unconditional rewrite assignment delivered.
	filteredArgs = view.LaunchArgs
	if entryDir != "" {
		filteredArgs = stripWorktreeSelection(view.LaunchArgs)
		if cherr := os.Chdir(entryDir); cherr != nil {
			endHandoff()
			return fmt.Errorf("enter the existing worktree %s: %w", entryDir, cherr)
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: entering the existing worktree %s\n", entryDir)
	}
	endHandoff()
	if debugRequested {
		// REQ-009's cc/glm form: the backend launch replaces this process
		// (exec), so the dump is the launcher's last output, printed before
		// the launch call.
		debugTiming.debugDump(cmd.ErrOrStderr())
	}
	return launch(profileName, mode, filteredArgs)
}
