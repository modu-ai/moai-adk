package cli

// factory.go is the Factory Mode entry surface on the two single-backend
// launchers. Factory Mode adds no runtime: it seeds a session whose
// orchestrator drives a plan -> run -> verify -> sync chain, and everything in
// this file exists to get that signal into the session.
//
// @MX:NOTE: [AUTO] the factory signal travels through the PROCESS environment, not a threaded parameter
// The one production consumer of the signal is the block-cap inject, five hops
// below runCC and reached with the factory token already stripped from args.
// Threading a parameter there would change four signatures plus two test seams
// to carry something the environment already delivers to the same line — and
// the child process needs the variables anyway.

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// The entry tokens. `-f` is unbound on cc / glm / cg; the commands that do bind
// it (`doctor config dump`, `state`) are distinct and unaffected.
const (
	factoryFlagLong  = "--factory"
	factoryFlagShort = "-f"
)

// factoryUnsupportedBackendSentinel is the machine-greppable marker on the
// `moai cg` rejection. A mixed leader/teammate backend contradicts the
// one-session / one-backend / one-chain premise, so the invocation is rejected
// rather than adapted.
const factoryUnsupportedBackendSentinel = "FACTORY_MODE_UNSUPPORTED_BACKEND"

// parseFactoryFlag extracts --factory / -f and its optional SPEC identifier
// from args, returning the identifier, whether Factory Mode was requested, and
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
func parseFactoryFlag(args []string) (spec string, enabled bool, rest []string) {
	filtered := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			filtered = append(filtered, args[i:]...)
			break
		}
		if arg != factoryFlagLong && arg != factoryFlagShort {
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
// @MX:REASON: os.Setenv is process-global; an unrestored mutation leaves MOAI_FACTORY set for every later test in the internal/cli binary (making the block-cap negative control pass or fail by execution order) and lets a later production re-exec inherit factory semantics it was never given
//
// enterFactoryMode publishes the factory signal into the process environment
// and returns the function that puts the environment back.
//
// The restore returns each variable to its PRIOR PRESENCE, not merely its prior
// value: a variable that was unset is unset again rather than set to "". The
// distinction is observable — os.Getenv cannot see it but os.LookupEnv can, and
// downstream readers treat presence as the signal.
//
// Callers must defer the returned function so it also runs on the error path; a
// restore that only runs on success is the same leak with a narrower trigger.
func enterFactoryMode(specID string) func() {
	restoreFactory := captureEnvState(config.EnvMoaiFactory)
	restoreSpec := captureEnvState(config.EnvMoaiFactorySpec)

	_ = os.Setenv(config.EnvMoaiFactory, "1")
	if specID != "" {
		_ = os.Setenv(config.EnvMoaiFactorySpec, specID)
	}

	return func() {
		restoreSpec()
		restoreFactory()
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

// recordFactorySession persists the session-scoped factory state record the
// orchestrator later fills in as the chain progresses.
//
// Best-effort throughout: an unresolvable session, an unwritable state
// directory, or an encoding failure all degrade to a session with no record.
// That is a launch the chain drives without stored state, and downstream the
// missing record reads as "no evidence" — which resolves in the safe direction,
// running the sync-phase check rather than skipping it.
func recordFactorySession(specID, backend string) {
	projectRoot := launchProjectRoot()
	sessionID := resolveLaunchSessionID("")
	if projectRoot == "" || sessionID == "" {
		return
	}
	factory.WriteBestEffort(projectRoot, factory.NewRecord(sessionID, specID, backend))
}

// rejectFactoryOnCG returns the sentinel-bearing error when a factory token
// appears in a `moai cg` invocation, and nil otherwise.
func rejectFactoryOnCG(args []string) error {
	if _, enabled, _ := parseFactoryFlag(args); !enabled {
		return nil
	}
	return fmt.Errorf("%s: moai cg runs a mixed backend (leader Claude, teammates GLM), "+
		"which contradicts Factory Mode's one-session / one-backend / one-chain premise; "+
		"use 'moai cc --factory' or 'moai glm --factory' instead", factoryUnsupportedBackendSentinel)
}
