package cli

// kanban.go is the Kanban Mode entry surface on the two single-backend
// launchers. Kanban Mode adds no runtime: it seeds a session whose
// orchestrator drives a plan -> run -> verify -> sync chain, and everything in
// this file exists to get that signal into the session.
//
// @MX:NOTE: [AUTO] the kanban signal travels through the PROCESS environment, not a threaded parameter
// The one production consumer of the signal is the block-cap inject, five hops
// below runCC and reached with the kanban token already stripped from args.
// Threading a parameter there would change four signatures plus two test seams
// to carry something the environment already delivers to the same line — and
// the child process needs the variables anyway.

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// The entry tokens. `-k` is unbound on cc / glm / cg; the commands that do bind
// it (`doctor config dump`, `state`) are distinct and unaffected.
const (
	kanbanFlagLong  = "--kanban"
	kanbanFlagShort = "-k"
)

// kanbanUnsupportedBackendSentinel is the machine-greppable marker on the
// `moai cg` rejection. A mixed leader/teammate backend contradicts the
// one-session / one-backend / one-chain premise, so the invocation is rejected
// rather than adapted.
const kanbanUnsupportedBackendSentinel = "KANBAN_MODE_UNSUPPORTED_BACKEND"

// parseKanbanFlag extracts --kanban / -k and its optional SPEC identifier
// from args, returning the identifier, whether Kanban Mode was requested, and
// the remaining args with the consumed tokens removed.
//
// The SPEC identifier is optional: its absence means the chain begins at
// plan-phase from the operator's first prompt, which is a valid entry, not an
// error. A token that looks like a flag is never consumed as the identifier.
//
// The `--` discipline matches stripSpawnFlag, parseProfileFlag, and
// normalizeWorktreeFlag exactly: iterate, break at the pass-through marker, and
// forward everything from that marker onward verbatim. Both commands set
// DisableFlagParsing, so a cobra flag registration would be silently inert —
// this manual parser is the only mechanism available.
func parseKanbanFlag(args []string) (spec string, enabled bool, rest []string) {
	filtered := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			filtered = append(filtered, args[i:]...)
			break
		}
		if arg != kanbanFlagLong && arg != kanbanFlagShort {
			filtered = append(filtered, arg)
			continue
		}

		enabled = true
		// Consume a following positional token as the SPEC identifier. A flag
		// or the pass-through marker belongs to someone else.
		if next := i + 1; next < len(args) && args[next] != "--" && !strings.HasPrefix(args[next], "-") {
			spec = args[next]
			i = next
		}
	}

	return spec, enabled, filtered
}

// @MX:ANCHOR: [AUTO] the deferred restore is a correctness requirement, not hygiene
// @MX:REASON: os.Setenv is process-global; an unrestored mutation leaves MOAI_KANBAN set for every later test in the internal/cli binary (making the block-cap negative control pass or fail by execution order) and lets a later production re-exec inherit kanban semantics it was never given
//
// enterKanbanMode publishes the kanban signal into the process environment
// and returns the function that puts the environment back.
//
// The restore returns each variable to its PRIOR PRESENCE, not merely its prior
// value: a variable that was unset is unset again rather than set to "". The
// distinction is observable — os.Getenv cannot see it but os.LookupEnv can, and
// downstream readers treat presence as the signal.
//
// Callers must defer the returned function so it also runs on the error path; a
// restore that only runs on success is the same leak with a narrower trigger.
func enterKanbanMode(specID string) func() {
	restoreKanban := captureEnvState(config.EnvMoaiKanban)
	restoreSpec := captureEnvState(config.EnvMoaiKanbanSpec)
	restoreID := captureEnvState(config.EnvMoaiKanbanID)
	restoreAddr := captureEnvState(config.EnvMoaiKanbanLeadAddr)
	restoreTier := seedAutonomyTier()

	_ = os.Setenv(config.EnvMoaiKanban, "1")
	runID := kanban.NewRunID()
	_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	if specID != "" {
		_ = os.Setenv(config.EnvMoaiKanbanSpec, specID)
	}
	// SPEC-FACTORY-BOOTSTRAP-001 M3: surface a leader socket path for the
	// SessionStart hook to print. The actual messaging-substrate address is a
	// run-phase concern; this conventional path-shaped value gives the notice
	// a non-empty, grep-friendly address line.
	_ = os.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-"+runID)

	return func() {
		restoreTier()
		restoreAddr()
		restoreID()
		restoreSpec()
		restoreKanban()
	}
}

// seedAutonomyTier publishes the autonomy tier Kanban Mode runs at, and returns
// the function that puts it back on the same prior-presence contract as the
// other kanban variables.
//
// Kanban Mode exists to run unattended: the operator launches the sessions once
// and the board advances without being asked at every step. The tier is what
// makes that true. At fully-autonomous the synchronous vet+lint+test commit gate
// and the SubagentStop / TeammateIdle / TaskCompleted lifecycle hooks stand
// down, so a card crosses a column without stopping for a verification tax the
// operator already accepted when they launched the board. Leaving the variable
// unset resolves to semi-auto — config.AutonomyTier fails safe — which is the
// most-interrupted tier and the opposite of what -k asks for.
//
// What the tier does NOT reach: the destructive-pattern denylist in
// internal/hook/pre_tool.go is tier-invariant, and Implementation Kickoff
// Approval is a gate the tier has no bearing on. Autonomy here buys fewer
// interruptions, never a weaker refusal.
//
// An operator who sets the variable themselves keeps it. Only an absent or
// blank value is filled in, and blank counts as absent because that is how
// config.AutonomyTier already reads it — a wrapper that exports the name with
// no value has not made a choice.
func seedAutonomyTier() func() {
	restore := captureEnvState(config.EnvAutonomyTier)
	if strings.TrimSpace(os.Getenv(config.EnvAutonomyTier)) == "" {
		_ = os.Setenv(config.EnvAutonomyTier, config.AutonomyTierFullyAutonomous)
	}
	return restore
}

// enterKanbanCompanionMode publishes the companion signal for a `<role>-<id>`
// label and returns the function that puts the environment back, on the same
// prior-presence contract as enterKanbanMode.
//
// It deliberately does NOT set config.EnvMoaiKanban: that variable seeds the
// chain, and only the lead drives the chain. What the companion shares with the
// lead is the raised Stop-hook block cap, which the inject reads from either
// variable.
//
// The run id is derived from the label rather than carried separately, so the
// two can never disagree.
func enterKanbanCompanionMode(label string) func() {
	restoreLabel := captureEnvState(config.EnvMoaiKanbanLabel)
	restoreID := captureEnvState(config.EnvMoaiKanbanID)
	// A companion is where the work actually lands — it plans, implements,
	// reviews, and commits. Seeding the tier on the lead alone would leave every
	// interruption exactly where it already was, because the lead does not
	// commit code.
	restoreTier := seedAutonomyTier()

	_ = os.Setenv(config.EnvMoaiKanbanLabel, label)
	if _, runID, ok := kanban.SplitCompanionLabel(label); ok {
		_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	}

	return func() {
		restoreTier()
		restoreID()
		restoreLabel()
	}
}

// captureEnvState records a variable's value AND its presence, returning the
// function that restores both.
func captureEnvState(key string) func() {
	prev, had := os.LookupEnv(key)
	return func() {
		if had {
			_ = os.Setenv(key, prev)
			return
		}
		_ = os.Unsetenv(key)
	}
}

// recordKanbanSession persists the session-scoped kanban state record the
// orchestrator later fills in as the chain progresses.
//
// Best-effort throughout: an unresolvable session, an unwritable state
// directory, or an encoding failure all degrade to a session with no record.
// That is a launch the chain drives without stored state, and downstream the
// missing record reads as "no evidence" — which resolves in the safe direction,
// running the sync-phase check rather than skipping it.
func recordKanbanSession(specID, backend string) {
	projectRoot := launchProjectRoot()
	sessionID := resolveLaunchSessionID("")
	if projectRoot == "" || sessionID == "" {
		return
	}
	kanban.WriteBestEffort(projectRoot, kanban.NewRecord(sessionID, specID, backend))
}

// The tokens claude uses to name a session. moai RECOGNIZES them; it never
// consumes them — the value has to reach claude unchanged.
const (
	nameFlagLong  = "--name"
	nameFlagShort = "-n"
)

// parseCompanionLabel reports the companion label in args, if any.
//
// It matches only the companion SHAPE (kanban.SplitCompanionLabel) because the
// alternative discriminators are worse. Treating every named session as a
// companion would
// silently raise the Stop-hook block cap from 8 to 200 for unrelated work, and a
// state file the lead writes and companions read buys nothing here beyond one
// more file to keep consistent.
//
// The `--` discipline matches parseKanbanFlag, stripSpawnFlag, parseProfileFlag
// and normalizeWorktreeFlag: iterate, break at the pass-through marker, and read
// nothing beyond it. args is returned to the caller untouched.
func parseCompanionLabel(args []string) (label string, ok bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}

		var candidate string
		switch {
		case arg == nameFlagLong || arg == nameFlagShort:
			if next := i + 1; next < len(args) {
				candidate = args[next]
				i = next
			}
		case strings.HasPrefix(arg, nameFlagLong+"="):
			candidate = strings.TrimPrefix(arg, nameFlagLong+"=")
		case strings.HasPrefix(arg, nameFlagShort+"="):
			candidate = strings.TrimPrefix(arg, nameFlagShort+"=")
		default:
			continue
		}

		if _, _, isCompanion := kanban.SplitCompanionLabel(candidate); isCompanion {
			return candidate, true
		}
	}
	return "", false
}

// rejectKanbanOnCG returns the sentinel-bearing error when a kanban token
// appears in a `moai cg` invocation, and nil otherwise.
func rejectKanbanOnCG(args []string) error {
	if _, enabled, _ := parseKanbanFlag(args); !enabled {
		return nil
	}
	return fmt.Errorf("%s: moai cg runs a mixed backend (leader Claude, teammates GLM), "+
		"which contradicts Kanban Mode's one-session / one-backend / one-chain premise; "+
		"use 'moai cc --kanban' or 'moai glm --kanban' instead", kanbanUnsupportedBackendSentinel)
}

// kanbanBranch enumerates the three dispatch outcomes of the §A.2 truth table.
type kanbanBranch int

const (
	kanbanBranchNone     kanbanBranch = iota // no-op — -k absent (regardless of --name shape)
	kanbanBranchLead                          // -k present, --name is NOT companion-shape
	kanbanBranchCompanion                     // -k present, --name IS companion-shape
)

// resolveKanbanBranch selects the dispatch branch from the combination of -k
// present and companion-shape --name present.
//
// This is the four-row truth table at spec.md §A.2 (REQ-FB-001, REQ-FB-002):
//
//	kanbanEnabled | isCompanion || branch
//	---------------++--------------
//	      true      |    false     || lead       (-k alone, or -k --name <non-companion>)
//	      true      |    true      || companion  (-k --name <role>-<run-id>)
//	      false     |    false     || no-op      (--name <non-companion>, or no --name)
//	      false     |    true      || no-op      (--name <companion-shape> alone — BREAKING from 94025ce0a)
//
// The two !kanbanEnabled rows collapse to no-op because `isCompanion` is
// consulted only when -k is present (spec.md §A.2.1 / AC-FB-027): a companion-
// shape --name alone, which entered companion mode under 94025ce0a, is
// reclassified as a no-op by REQ-FB-001's no-`-k` clause.
func resolveKanbanBranch(kanbanEnabled, isCompanion bool) kanbanBranch {
	switch {
	case kanbanEnabled && isCompanion:
		return kanbanBranchCompanion
	case kanbanEnabled && !isCompanion:
		return kanbanBranchLead
	default:
		return kanbanBranchNone
	}
}
