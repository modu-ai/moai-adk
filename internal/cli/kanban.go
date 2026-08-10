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

// The entry tokens. `-k` is unbound everywhere else in the CLI, so no other
// command's flag set is affected. (The pre-rename `-f` short form was bound by
// `doctor config dump` and `state` as their --format shorthand; those two are
// distinct commands and were unaffected then as well.)
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

	_ = os.Setenv(config.EnvMoaiKanban, "1")
	if specID != "" {
		_ = os.Setenv(config.EnvMoaiKanbanSpec, specID)
	}

	return func() {
		restoreSpec()
		restoreKanban()
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
