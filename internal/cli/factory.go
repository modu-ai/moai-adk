package cli

// factory.go is the Factory Mode machinery behind the dedicated -f entry
// surface (`moai cc -f [N]` / `moai glm -f [N]`, t118 launcher axis, v3.1.1)
// and the factory shapes of the unified -k token (v1.2.0, kept for compat).
// A factory run is a lead session plus numbered lanes: the lead routes
// cards to the lanes over cross-session messages, and everything in this
// file exists to get that signal into the session and to keep the lane
// names unique. The -k shapes of the entry parse live in parseKanbanFlag
// (kanban.go); this file owns the -f parse (parseFactoryFlag), the merge of
// the two (parseLauncherEntry), and everything after the factory shape is
// selected.
//
// GENEALOGY (binding): the pre-3.1 "factory" flag (-f/--factory) was RENAMED
// to -k/--kanban in #1513 (7f61332ef) and drove the three-role kanban chain.
// -f briefly returned as the factory worker fan-out flag (v1.0.0, 2026-08-17)
// and was RETIRED the same day (v1.2.0) in favor of `-k <N>`. t118 (v3.1.1)
// REVIVED it as the dedicated factory entry — one entry flag per mode: the
// kanban chain keeps -k, the factory gets -f. This is where the retirement
// error (rejectRetiredFactoryFlag, v1.2.0) used to live; the #1513 rename
// history stays recorded here and in both launchers' help text.
//
// Like the kanban signal, the factory signal travels through the PROCESS
// environment rather than a threaded parameter, for the same reason: the
// consumers (block-cap inject, SessionStart hook) are reached with the flag
// already stripped from args, and the child process needs the variables
// anyway.

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryUnsupportedBackendSentinel is the machine-greppable marker on the
// `moai cg` rejection, mirroring kanbanUnsupportedBackendSentinel: a mixed
// leader/teammate backend contradicts factory mode's one-session /
// one-backend premise, so the invocation is rejected rather than adapted.
const factoryUnsupportedBackendSentinel = "FACTORY_MODE_UNSUPPORTED_BACKEND"

// The entry tokens. `-f` is unbound on cc / glm / cg outside this file's
// parse; `--factory` is its long form.
const (
	factoryFlagLong  = "--factory"
	factoryFlagShort = "-f"
)

// factoryFlagUsageError names every accepted -f shape. It is the error text
// for an invalid SUPPLIED value and the reference the help texts paraphrase.
const factoryFlagUsageError = "-f/--factory takes a lane count of 1 or more (e.g. -f 4), " +
	"a lane label (e.g. -f lane-2) that launches exactly that one lane, " +
	"or no argument for the one-lane factory default"

// factoryFlagParse is the -f/--factory entry parse (t118). ONE flag token,
// three shapes:
//
//	-f                → the factory lead, one lane (DefaultFactoryLeadWorkers)
//	-f N (N ≥ 1)      → the factory lead with N numbered lanes
//	-f lane-<n>       → exactly one additional lane, lane n — the
//	                    incremental form; the run's count is not carried, and
//	                    a count of 0 ("unknown") flows into the lane env
//
// The factory carries no SPEC identifier (that is the kanban chain's -k
// SPEC-ID shape), so a SUPPLIED value that is neither a bare positive integer
// nor a lane label is an error — there is no second interpretation to
// silently fall into, and hiding the typo would be worse than naming it.
type factoryFlagParse struct {
	Enabled      bool     // -f present (any shape)
	Workers      int      // the explicit count; 0 when omitted or lane-form
	WorkerNumber int      // n of `-f lane-<n>`; 0 unless the lane form
	Rest         []string // args with -f and its consumed value removed
}

// parseFactoryFlag extracts --factory / -f and its optional value from args.
//
// The value-consumption and `--` discipline match parseKanbanFlag exactly:
// a following token that looks like a flag is never consumed as the value,
// everything from the pass-through marker onward is forwarded verbatim, and
// both commands set DisableFlagParsing so this manual parser is the only
// mechanism available.
func parseFactoryFlag(args []string) (p factoryFlagParse, err error) {
	p.Rest = make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			p.Rest = append(p.Rest, args[i:]...)
			break
		}

		var value string
		hasValue := false
		switch {
		case arg == factoryFlagLong || arg == factoryFlagShort:
			if next := i + 1; next < len(args) && args[next] != "--" && !strings.HasPrefix(args[next], "-") {
				value, hasValue = args[next], true
				i = next
			}
		case strings.HasPrefix(arg, factoryFlagLong+"="), strings.HasPrefix(arg, factoryFlagShort+"="):
			value = strings.TrimPrefix(strings.TrimPrefix(arg, factoryFlagShort+"="), factoryFlagLong+"=")
			hasValue = true
		default:
			p.Rest = append(p.Rest, arg)
			continue
		}

		p.Enabled = true
		if !hasValue {
			continue
		}
		if n, perr := strconv.Atoi(value); perr == nil {
			if n < 1 {
				return p, fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
			}
			p.Workers = n
			continue
		}
		if n, ok := kanban.SplitFactoryLaneLabel(value); ok {
			p.WorkerNumber = n
			continue
		}
		return p, fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
	}

	return p, nil
}

// parseLauncherEntry is the unified -k + -f entry parse (t118): parseKanbanFlag
// for the -k shapes, parseFactoryFlag on what remains, then merge. The merge
// enforces the one-entry-token rule (-f and -k together is an error — each
// names a different mode, and a launch that carried both would have to
// silently drop one) and resolves the -f shapes into the same kanbanEntryParse
// the dispatch branches already read:
//
//   - `-f [N]` sets FactoryEnabled with Workers (N, or
//     DefaultFactoryLeadWorkers when omitted);
//   - `-f lane-<n>` sets FactoryEnabled and appends `--name lane-<n>` to
//     Rest, desugaring into the existing lane branch (registry bump,
//     replaceNamedLabel, settings injection, per-lane agent cap) so the new
//     form and the `-k N --name lane-<i>` form share one implementation.
//     The run's count is NOT fabricated — a lane joined incrementally
//     carries 0 ("unknown"), which the lane notice degrades from. A -f
//     lane form plus an operator-supplied --name is a conflict error: the
//     flag value already named the lane.
//   - `-f N --name lane-<i>` keeps N (mirroring `-k N --name lane-<i>`).
func parseLauncherEntry(args []string) (kanbanEntryParse, error) {
	entry, err := parseKanbanFlag(args)
	if err != nil {
		return entry, err
	}
	fp, err := parseFactoryFlag(entry.Rest)
	if err != nil {
		return entry, err
	}
	if !fp.Enabled {
		return entry, nil
	}
	if entry.KanbanEnabled {
		return entry, fmt.Errorf("-k/--kanban and -f/--factory are two entry tokens; a launch carries at most one — " +
			"use -k [SPEC-ID] for the kanban chain or -f [N] for the factory")
	}

	entry.FactoryEnabled = true
	// The stripped args always become the launch args — for every -f shape,
	// not only the lane form below. (The lane form appends its desugared
	// --name on top of these.)
	entry.Rest = fp.Rest
	switch {
	case fp.WorkerNumber > 0:
		if operatorSuppliedName(fp.Rest) {
			return entry, fmt.Errorf("-f lane-<n> already names the lane; drop the --name/-n flag (got args %v)", fp.Rest)
		}
		entry.Rest = append(entry.Rest, nameFlagLong, kanban.FactoryLaneLabel(fp.WorkerNumber))
	case fp.Workers > 0:
		entry.FactoryWorkers = fp.Workers
	default:
		// Bare -f with no lane-shape --name: the one-lane lead default.
		// With a lane-shape --name and no count (the -k-style combo), the
		// count is 0 (unknown) — same honesty as the -f lane-<n> form.
		if _, isLane := parseFactoryLaneLabel(fp.Rest); !isLane {
			entry.FactoryWorkers = config.DefaultFactoryLeadWorkers
		}
	}
	return entry, nil
}

// factoryBranch enumerates the dispatch outcomes, mirroring kanbanBranch.
type factoryBranch int

const (
	factoryBranchNone   factoryBranch = iota // no-op — -f absent (regardless of --name shape)
	factoryBranchLead                        // -f N present, --name is NOT lane-shape
	factoryBranchWorker                      // -f N present, --name IS lane-shape
)

// resolveFactoryBranch selects the dispatch branch from -f present and
// lane-shape --name present — the factory counterpart of
// resolveKanbanBranch's truth table:
//
//	factoryEnabled | isLane   || branch
//	----------------++--------------
//	      true      |   false   || lead     (-f N alone, or -f N --name <non-lane>)
//	      true      |   true    || worker   (-f N --name lane-<n>)
//	      false     |   any     || no-op    (--name lane-3 alone does NOT join a run)
func resolveFactoryBranch(factoryEnabled, isLane bool) factoryBranch {
	switch {
	case factoryEnabled && isLane:
		return factoryBranchWorker
	case factoryEnabled && !isLane:
		return factoryBranchLead
	default:
		return factoryBranchNone
	}
}

// parseFactoryLaneLabel reports the `lane-<n>` label in args, if any.
// It matches only the lane SHAPE (kanban.SplitFactoryLaneLabel), for the
// same reason parseCompanionLabel matches only the companion shape: treating
// every named session as a lane would silently change launch behavior for
// unrelated work.
func parseFactoryLaneLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		_, isLane := kanban.SplitFactoryLaneLabel(candidate)
		return isLane
	})
}

// enterFactoryLeadMode publishes the factory lead signal into the process
// environment and returns the function that puts the environment back, on the
// same prior-presence contract as enterKanbanMode.
//
// It reuses the kanban lead's run-id and leader-socket env surfaces — the run
// id names the run (and the injected `lead-<run-id>` session name) — while
// deliberately NOT setting EnvMoaiKanban or EnvMoaiKanbanLabel: those seed the
// three-role kanban chain, which a factory lead never drives. The factory
// discriminator the hook and the block-cap inject read is
// EnvMoaiFactoryWorkers.
//
// leadLabel is the operator-supplied `lead-<run-id>` name for this session
// when there is one, and "" otherwise; its run id is adopted rather than
// replaced (see leadRunID).
func enterFactoryLeadMode(workers int, leadLabel string) func() {
	restoreWorkers := captureEnvState(config.EnvMoaiFactoryWorkers)
	restoreID := captureEnvState(config.EnvMoaiKanbanID)
	restoreAddr := captureEnvState(config.EnvMoaiKanbanLeadAddr)
	restoreTier := seedAutonomyTier()

	_ = os.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(workers))
	runID := leadRunID(leadLabel)
	_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	// The conventional path-shaped address, from the factory's own socket
	// directory (t118 scheme): the actual messaging-substrate address is a run
	// concern, and this value gives the notice a non-empty, grep-friendly
	// address line that never collides with a concurrent kanban run's.
	_ = os.Setenv(config.EnvMoaiKanbanLeadAddr, kanban.FactoryLeaderSocketPath(runID))

	return func() {
		restoreTier()
		restoreAddr()
		restoreID()
		restoreWorkers()
	}
}

// enterFactoryWorkerMode publishes the factory lane signal for a `lane-<n>`
// label and returns the function that puts the environment back, on the same
// prior-presence contract. It is enterKanbanCompanionMode's factory
// counterpart: no chain is seeded (no factory analogue of EnvMoaiKanban
// exists to seed one), the raised Stop-hook block cap reaches the session
// through EnvMoaiFactoryWorkers, and the autonomy tier and the per-lane
// agent cap are seeded because the lane is where dispatched cards are
// actually implemented.
//
// workers is the run's fan-out size from the lane's own entry token —
// `-f <N>` / `-k <N>` carry it explicitly, and the incremental `-f
// lane-<n>` form carries 0 ("unknown"), which the lane notice degrades
// from rather than fabricating a count.
func enterFactoryWorkerMode(label string, workers int) func() {
	restoreLabel := captureEnvState(config.EnvMoaiFactoryWorker)
	restoreWorkers := captureEnvState(config.EnvMoaiFactoryWorkers)
	restoreTier := seedAutonomyTier()
	restoreCap := seedLaneAgentCap()

	_ = os.Setenv(config.EnvMoaiFactoryWorker, label)
	_ = os.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(workers))

	return func() {
		restoreCap()
		restoreTier()
		restoreWorkers()
		restoreLabel()
	}
}

// factoryWorkerEntry is one registered lane: the pid of the process that
// claimed the label. The type lives in internal/kanban (factory_slots.go)
// since the t85 lead loop — the alias keeps this package's call sites and
// tests on their historical name.
type factoryWorkerEntry = kanban.FactoryWorkerEntry

// factoryRegistryPath / loadFactoryRegistry / saveFactoryRegistry delegate to
// the kanban registry cluster (factory_slots.go). The cluster moved out of
// this file because the SessionStart hook needs the same reads and cannot
// import this package; these delegates keep the cli surface stable.
func factoryRegistryPath(root string) string { return kanban.FactoryRegistryPath(root) }

func loadFactoryRegistry(path string) map[string]factoryWorkerEntry {
	return kanban.LoadFactoryRegistry(path)
}

func saveFactoryRegistry(path string, reg map[string]factoryWorkerEntry) error {
	return kanban.SaveFactoryRegistry(path, reg)
}

// factoryProcessAlive is the liveness seam; tests override it to simulate
// live and dead claims without spawning processes. The default probe itself
// lives in internal/kanban (factory_alive_*.go) since the same move.
var factoryProcessAlive = kanban.FactoryProcessAlive

// resolveFactoryWorkerName returns the label this lane session should
// launch under: label itself when its number is free, or the next incremented
// number whose label is free (the "bump a conflicting number up" rule).
//
// A number is taken when the registry maps its label to a pid that is alive
// right now — a crashed or exited lane leaves a dead pid behind, and a dead
// claim frees the name so a relaunch reuses it instead of counting up
// forever. Dead entries are pruned on the way through. The final label is
// registered to this process's pid before returning.
//
// label MUST have the lane shape (kanban.SplitFactoryLaneLabel); the
// caller has already branched on it. Best-effort throughout: an unreadable or
// unwritable registry degrades to using the label as supplied, exactly like
// every other launch-path state write — the launch must never block on it.
// notes, when non-nil, receives the operator-visible bump line.
func resolveFactoryWorkerName(root, label string, notes io.Writer) string {
	path := factoryRegistryPath(root)
	reg := kanban.PruneFactoryDeadClaims(loadFactoryRegistry(path), factoryProcessAlive)

	final := label
	if n, ok := kanban.SplitFactoryLaneLabel(label); ok {
		for {
			claim, taken := reg[final]
			if !taken || claim.PID <= 0 || !factoryProcessAlive(claim.PID) {
				break
			}
			n++
			final = kanban.FactoryLaneLabel(n)
		}
		if final != label && notes != nil {
			// The note is best-effort operator guidance; the SessionStart
			// lane notice is the reliable surface for the final name.
			_, _ = fmt.Fprintf(notes, "factory: %s is held by a live session; launching as %s\n", label, final)
		}
	}

	reg[final] = kanban.NewFactoryWorkerEntry()
	_ = saveFactoryRegistry(path, reg)
	return final
}

// replaceNamedLabel returns args with the first `--name` / `-n` value equal
// to oldLabel (in any of the four forms claude accepts, before the
// pass-through marker) replaced by newLabel. It is what carries a bumped
// worker number into the session's actual name — the name is the address the
// lead dispatches to, so the bump must reach the backend argv, not just the
// environment. args is never mutated.
func replaceNamedLabel(args []string, oldLabel, newLabel string) []string {
	if oldLabel == newLabel {
		return args
	}
	out := make([]string, len(args))
	copy(out, args)
	for i := 0; i < len(out); i++ {
		arg := out[i]
		if arg == "--" {
			break
		}
		switch {
		case (arg == nameFlagLong || arg == nameFlagShort) && i+1 < len(out) && out[i+1] == oldLabel:
			out[i+1] = newLabel
			return out
		case strings.HasPrefix(arg, nameFlagLong+"=") && strings.TrimPrefix(arg, nameFlagLong+"=") == oldLabel:
			out[i] = nameFlagLong + "=" + newLabel
			return out
		case strings.HasPrefix(arg, nameFlagShort+"=") && strings.TrimPrefix(arg, nameFlagShort+"=") == oldLabel:
			out[i] = nameFlagShort + "=" + newLabel
			return out
		}
	}
	return out
}

// rejectFactoryOnCG returns the sentinel-bearing error when a FACTORY shape
// appears in a `moai cg` invocation — a lane count or lane label on
// either entry token, -k (v1.2.0 shapes) or -f (t118 shapes) — and nil
// otherwise. Parse errors (an invalid count, an invalid -f value, the -f+-k
// conflict) surface here too, because the unified parse runs in this
// function. The plain kanban shapes of -k are rejectKanbanOnCG's to answer —
// the two rejections split the entry surface's shapes between them.
func rejectFactoryOnCG(args []string) error {
	p, err := parseLauncherEntry(args)
	if err != nil {
		return err
	}
	if !p.FactoryEnabled {
		return nil
	}
	return fmt.Errorf("%s: moai cg runs a mixed backend (leader Claude, teammates GLM), "+
		"which contradicts Factory Mode's one-session / one-backend premise; "+
		"use 'moai cc -f <N>' or 'moai glm -f <N>' instead", factoryUnsupportedBackendSentinel)
}
