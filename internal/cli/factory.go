package cli

// factory.go is the Factory Mode machinery behind the unified -k entry surface
// (`moai cc -k <N>` / `moai glm -k <N>`, SPEC-FACTORY-WORKER-FANOUT-001
// v1.2.0). A factory run is a lead session plus N numbered workers: the lead
// routes cards to the workers over cross-session messages, and everything in
// this file exists to get that signal into the session and to keep the worker
// names unique. The ENTRY PARSE itself lives in parseKanbanFlag (kanban.go) —
// one -k token selects either shape there; this file owns what happens after
// the factory shape is selected.
//
// GENEALOGY (binding): the pre-3.1 "factory" flag (-f/--factory) was RENAMED
// to -k/--kanban in #1513 (7f61332ef) and now drives the four-role kanban
// chain. -f briefly returned as the factory worker fan-out flag (v1.0.0,
// 2026-08-17) and was RETIRED the same day (v1.2.0): the factory is entered
// with the same -k carrying a worker count. Any -f/--factory token is an
// explicit error today (rejectRetiredFactoryFlag), never a silent pass-
// through.
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

// rejectRetiredFactoryFlag errors on any -f / --factory token (v1.2.0): the
// flag was retired when the factory entry unified on -k <N>, and a retired
// flag that silently falls through to the backend would either misfire there
// or — worse — look like it worked. The error names where the entry went.
// Exact-token matching only: `-foo` and `--factory-reset` belong to whoever
// reads them, and a prefix match would steal those tokens.
func rejectRetiredFactoryFlag(args []string) error {
	for _, arg := range args {
		if arg == "--" {
			return nil
		}
		if arg == "-f" || arg == "--factory" ||
			strings.HasPrefix(arg, "-f=") || strings.HasPrefix(arg, "--factory=") {
			return fmt.Errorf("the -f/--factory flag was retired; enter Factory Mode with -k <N> (e.g. moai cc -k 4), or use -k [SPEC-ID] for the kanban chain")
		}
	}
	return nil
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
// It reuses the kanban lead's run-id and leader-socket surfaces — the run id
// names the run (and the injected `lead-<run-id>` session name), and the
// socket path is the notice's address line — while deliberately NOT setting
// EnvMoaiKanban or EnvMoaiKanbanLabel: those seed the four-role kanban chain,
// which a factory lead never drives. The factory discriminator the hook and
// the block-cap inject read is EnvMoaiFactoryWorkers.
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
	// Same conventional path-shaped address as the kanban lead: the actual
	// messaging-substrate address is a run concern, and this value gives the
	// notice a non-empty, grep-friendly address line.
	_ = os.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-"+runID)

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
// the session through EnvMoaiFactoryWorkers, and the autonomy tier is seeded
// because the worker is where dispatched cards are actually implemented.
//
// workers is the run's fan-out size from the worker's own `-f <N>` token —
// the count travels in the launch command precisely so the worker knows it.
func enterFactoryWorkerMode(label string, workers int) func() {
	restoreLabel := captureEnvState(config.EnvMoaiFactoryWorker)
	restoreWorkers := captureEnvState(config.EnvMoaiFactoryWorkers)
	restoreTier := seedAutonomyTier()

	_ = os.Setenv(config.EnvMoaiFactoryWorker, label)
	_ = os.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(workers))

	return func() {
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
// of the unified -k surface appears in a `moai cg` invocation (and the parse
// error when the count is malformed), and nil otherwise. The plain kanban
// shapes of -k are rejectKanbanOnCG's to answer — the two rejections split
// the unified flag's shapes between them.
func rejectFactoryOnCG(args []string) error {
	p, err := parseKanbanFlag(args)
	if err != nil {
		return err
	}
	if !p.FactoryEnabled {
		return nil
	}
	return fmt.Errorf("%s: moai cg runs a mixed backend (leader Claude, teammates GLM), "+
		"which contradicts Factory Mode's one-session / one-backend premise; "+
		"use 'moai cc -k <N>' or 'moai glm -k <N>' instead", factoryUnsupportedBackendSentinel)
}
