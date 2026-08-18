package cli

// factory.go is the Factory Mode machinery behind the dedicated -f entry
// surface (`moai cc -f [N]` / `moai glm -f [N]`, t118 launcher axis, v3.1.1)
// and the factory shapes of the unified -k token (v1.2.0, kept for compat).
// A factory run is a lead session plus numbered workers: the lead routes
// cards to the workers over cross-session messages, and everything in this
// file exists to get that signal into the session and to keep the worker
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
const factoryFlagUsageError = "-f/--factory takes a worker count of 1 or more (e.g. -f 4), " +
	"a worker label (e.g. -f worker-2) that launches exactly that one worker, " +
	"or no argument for the one-worker factory default"

// factoryFlagParse is the -f/--factory entry parse (t118). ONE flag token,
// three shapes:
//
//	-f                → the factory lead, one worker (DefaultFactoryLeadWorkers)
//	-f N (N ≥ 1)      → the factory lead with N numbered workers
//	-f worker-<n>     → exactly one additional worker, worker n — the
//	                    incremental form; the run's count is not carried, and
//	                    a count of 0 ("unknown") flows into the worker env
//
// The factory carries no SPEC identifier (that is the kanban chain's -k
// SPEC-ID shape), so a SUPPLIED value that is neither a bare positive integer
// nor a worker label is an error — there is no second interpretation to
// silently fall into, and hiding the typo would be worse than naming it.
type factoryFlagParse struct {
	Enabled      bool     // -f present (any shape)
	Workers      int      // the explicit count; 0 when omitted or worker-form
	WorkerNumber int      // n of `-f worker-<n>`; 0 unless the worker form
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
		if n, ok := kanban.SplitFactoryWorkerLabel(value); ok {
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
//   - `-f worker-<n>` sets FactoryEnabled and appends `--name worker-<n>` to
//     Rest, desugaring into the existing worker branch (registry bump,
//     replaceNamedLabel, settings injection, per-lane agent cap) so the new
//     form and the `-k N --name worker-<i>` form share one implementation.
//     The run's count is NOT fabricated — a worker joined incrementally
//     carries 0 ("unknown"), which the worker notice degrades from. A -f
//     worker form plus an operator-supplied --name is a conflict error: the
//     flag value already named the worker.
//   - `-f N --name worker-<i>` keeps N (mirroring `-k N --name worker-<i>`).
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
	// not only the worker form below. (The worker form appends its desugared
	// --name on top of these.)
	entry.Rest = fp.Rest
	switch {
	case fp.WorkerNumber > 0:
		if operatorSuppliedName(fp.Rest) {
			return entry, fmt.Errorf("-f worker-<n> already names the worker; drop the --name/-n flag (got args %v)", fp.Rest)
		}
		entry.Rest = append(entry.Rest, nameFlagLong, kanban.FactoryWorkerLabel(fp.WorkerNumber))
	case fp.Workers > 0:
		entry.FactoryWorkers = fp.Workers
	default:
		// Bare -f with no worker-shape --name: the one-worker lead default.
		// With a worker-shape --name and no count (the -k-style combo), the
		// count is 0 (unknown) — same honesty as the -f worker-<n> form.
		if _, isWorker := parseFactoryWorkerLabel(fp.Rest); !isWorker {
			entry.FactoryWorkers = config.DefaultFactoryLeadWorkers
		}
	}
	return entry, nil
}

// factoryBranch enumerates the dispatch outcomes, mirroring kanbanBranch.
type factoryBranch int

const (
	factoryBranchNone   factoryBranch = iota // no-op — -f absent (regardless of --name shape)
	factoryBranchLead                        // -f N present, --name is NOT worker-shape
	factoryBranchWorker                      // -f N present, --name IS worker-shape
)

// resolveFactoryBranch selects the dispatch branch from -f present and
// worker-shape --name present — the factory counterpart of
// resolveKanbanBranch's truth table:
//
//	factoryEnabled | isWorker || branch
//	----------------++--------------
//	      true      |   false   || lead     (-f N alone, or -f N --name <non-worker>)
//	      true      |   true    || worker   (-f N --name worker-<n>)
//	      false     |   any     || no-op    (--name worker-3 alone does NOT join a run)
func resolveFactoryBranch(factoryEnabled, isWorker bool) factoryBranch {
	switch {
	case factoryEnabled && isWorker:
		return factoryBranchWorker
	case factoryEnabled && !isWorker:
		return factoryBranchLead
	default:
		return factoryBranchNone
	}
}

// parseFactoryWorkerLabel reports the `worker-<n>` label in args, if any.
// It matches only the worker SHAPE (kanban.SplitFactoryWorkerLabel), for the
// same reason parseCompanionLabel matches only the companion shape: treating
// every named session as a worker would silently change launch behavior for
// unrelated work.
func parseFactoryWorkerLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		_, isWorker := kanban.SplitFactoryWorkerLabel(candidate)
		return isWorker
	})
}

// enterFactoryLeadMode publishes the factory lead signal into the process
// environment and returns the function that puts the environment back, on the
// same prior-presence contract as enterKanbanMode.
//
// It reuses the kanban lead's run-id and leader-socket env surfaces — the run
// id names the run (and the injected `leader-<run-id>` session name) — while
// deliberately NOT setting EnvMoaiKanban or EnvMoaiKanbanLabel: those seed the
// three-role kanban chain, which a factory lead never drives. The factory
// discriminator the hook and the block-cap inject read is
// EnvMoaiFactoryWorkers.
//
// leadLabel is the operator-supplied `leader-<run-id>` name for this session
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

// enterFactoryWorkerMode publishes the factory worker signal for a
// `worker-<n>` label and returns the function that puts the environment back,
// on the same prior-presence contract. It is enterKanbanCompanionMode's
// factory counterpart: no chain is seeded (no factory analogue of
// EnvMoaiKanban exists to seed one), the raised Stop-hook block cap reaches
// the session through EnvMoaiFactoryWorkers, and the autonomy tier and the
// per-lane agent cap are seeded because the worker is where dispatched cards
// are actually implemented.
//
// workers is the run's fan-out size from the worker's own entry token —
// `-f <N>` / `-k <N>` carry it explicitly, and the incremental `-f
// worker-<n>` form carries 0 ("unknown"), which the worker notice degrades
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

// factoryWorkerEntry is one registered worker: the pid of the process that
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

// resolveFactoryWorkerName returns the label this worker session should
// launch under: label itself when its number is free, or the next incremented
// number whose label is free (the "bump a conflicting number up" rule).
//
// A number is taken when the registry maps its label to a pid that is alive
// right now — a crashed or exited worker leaves a dead pid behind, and a dead
// claim frees the name so a relaunch reuses it instead of counting up
// forever. Dead entries are pruned on the way through. The final label is
// registered to this process's pid before returning.
//
// label MUST have the worker shape (kanban.SplitFactoryWorkerLabel); the
// caller has already branched on it. Best-effort throughout: an unreadable or
// unwritable registry degrades to using the label as supplied, exactly like
// every other launch-path state write — the launch must never block on it.
// notes, when non-nil, receives the operator-visible bump line.
func resolveFactoryWorkerName(root, label string, notes io.Writer) string {
	path := factoryRegistryPath(root)
	reg := kanban.PruneFactoryDeadClaims(loadFactoryRegistry(path), factoryProcessAlive)

	final := label
	if n, ok := kanban.SplitFactoryWorkerLabel(label); ok {
		for {
			claim, taken := reg[final]
			if !taken || claim.PID <= 0 || !factoryProcessAlive(claim.PID) {
				break
			}
			n++
			final = kanban.FactoryWorkerLabel(n)
		}
		if final != label && notes != nil {
			// The note is best-effort operator guidance; the SessionStart
			// worker notice is the reliable surface for the final name.
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
// appears in a `moai cg` invocation — a worker count or worker label on
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
