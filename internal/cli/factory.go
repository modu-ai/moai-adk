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
	restoreID := captureEnvState(config.EnvMoaiFactoryID)
	restoreAddr := captureEnvState(config.EnvMoaiFactoryLeadAddr)

	_ = os.Setenv(config.EnvMoaiFactory, "1")
	runID := factory.NewRunID()
	_ = os.Setenv(config.EnvMoaiFactoryID, runID)
	if specID != "" {
		_ = os.Setenv(config.EnvMoaiFactorySpec, specID)
	}
	// SPEC-FACTORY-BOOTSTRAP-001 M3: surface a leader socket path for the
	// SessionStart hook to print. The actual messaging-substrate address is a
	// run-phase concern; this conventional path-shaped value gives the notice
	// a non-empty, grep-friendly address line.
	_ = os.Setenv(config.EnvMoaiFactoryLeadAddr, "/tmp/moai-factory-"+runID)

	return func() {
		restoreAddr()
		restoreID()
		restoreSpec()
		restoreFactory()
	}
}

// enterFactoryCompanionMode publishes the companion signal for a `<role>-<id>`
// label and returns the function that puts the environment back, on the same
// prior-presence contract as enterFactoryMode.
//
// It deliberately does NOT set config.EnvMoaiFactory: that variable seeds the
// chain, and only the lead drives the chain. What the companion shares with the
// lead is the raised Stop-hook block cap, which the inject reads from either
// variable.
//
// The run id is derived from the label rather than carried separately, so the
// two can never disagree.
func enterFactoryCompanionMode(label string) func() {
	restoreLabel := captureEnvState(config.EnvMoaiFactoryLabel)
	restoreID := captureEnvState(config.EnvMoaiFactoryID)

	_ = os.Setenv(config.EnvMoaiFactoryLabel, label)
	if _, runID, ok := factory.SplitCompanionLabel(label); ok {
		_ = os.Setenv(config.EnvMoaiFactoryID, runID)
	}

	return func() {
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

// The tokens claude uses to name a session. moai RECOGNIZES them; it never
// consumes them — the value has to reach claude unchanged.
const (
	nameFlagLong  = "--name"
	nameFlagShort = "-n"
)

// parseCompanionLabel reports the companion label in args, if any.
//
// It matches only the companion SHAPE (factory.SplitCompanionLabel) because the
// alternative discriminators are worse. Treating every named session as a
// companion would
// silently raise the Stop-hook block cap from 8 to 200 for unrelated work, and a
// state file the lead writes and companions read buys nothing here beyond one
// more file to keep consistent.
//
// The `--` discipline matches parseFactoryFlag, stripSpawnFlag, parseProfileFlag
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

		if _, _, isCompanion := factory.SplitCompanionLabel(candidate); isCompanion {
			return candidate, true
		}
	}
	return "", false
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

// factoryBranch enumerates the three dispatch outcomes of the §A.2 truth table.
type factoryBranch int

const (
	factoryBranchNone     factoryBranch = iota // no-op — -f absent (regardless of --name shape)
	factoryBranchLead                          // -f present, --name is NOT companion-shape
	factoryBranchCompanion                     // -f present, --name IS companion-shape
)

// resolveFactoryBranch selects the dispatch branch from the combination of -f
// present and companion-shape --name present.
//
// This is the four-row truth table at spec.md §A.2 (REQ-FB-001, REQ-FB-002):
//
//	factoryEnabled | isCompanion || branch
//	---------------++--------------
//	      true      |    false     || lead       (-f alone, or -f --name <non-companion>)
//	      true      |    true      || companion  (-f --name <role>-<run-id>)
//	      false     |    false     || no-op      (--name <non-companion>, or no --name)
//	      false     |    true      || no-op      (--name <companion-shape> alone — BREAKING from 94025ce0a)
//
// The two !factoryEnabled rows collapse to no-op because `isCompanion` is
// consulted only when -f is present (spec.md §A.2.1 / AC-FB-027): a companion-
// shape --name alone, which entered companion mode under 94025ce0a, is
// reclassified as a no-op by REQ-FB-001's no-`-f` clause.
func resolveFactoryBranch(factoryEnabled, isCompanion bool) factoryBranch {
	switch {
	case factoryEnabled && isCompanion:
		return factoryBranchCompanion
	case factoryEnabled && !isCompanion:
		return factoryBranchLead
	default:
		return factoryBranchNone
	}
}
