package cli

// factory.go is the Factory Mode entry surface on the two single-backend
// launchers (`moai cc -f <N>` / `moai glm -f <N>`), the kanban entry's sibling
// (SPEC-FACTORY-WORKER-FANOUT-001). A factory run is a lead session plus N
// numbered workers: the lead dispatches cards to the workers over
// cross-session messages, and everything in this file exists to get that
// signal into the session and to keep the worker names unique.
//
// GENEALOGY (binding): the pre-3.1 "factory" flag (-f/--factory) was RENAMED
// to -k/--kanban in #1513 (7f61332ef) and now drives the four-role kanban
// chain. Today's -f is a NEW feature — a numbered worker fan-out — and shares
// nothing with that predecessor beyond the recycled letter. Any surface that
// documents one must not describe it as the other.
//
// Like the kanban signal, the factory signal travels through the PROCESS
// environment rather than a threaded parameter, for the same reason: the
// consumers (block-cap inject, SessionStart hook) are reached with the flag
// already stripped from args, and the child process needs the variables
// anyway.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// The entry tokens. `-f` is unbound on cc / glm; the long form exists because
// a bare `-f` reads as "force" to terminal convention and the confusion is
// cheaper to prevent than to debug.
const (
	factoryFlagLong  = "--factory"
	factoryFlagShort = "-f"
)

// factoryUnsupportedBackendSentinel is the machine-greppable marker on the
// `moai cg` rejection, mirroring kanbanUnsupportedBackendSentinel: a mixed
// leader/teammate backend contradicts factory mode's one-session /
// one-backend premise, so the invocation is rejected rather than adapted.
const factoryUnsupportedBackendSentinel = "FACTORY_MODE_UNSUPPORTED_BACKEND"

// parseFactoryFlag extracts --factory / -f and its OPTIONAL worker count from
// args, returning the count, whether Factory Mode was requested, and the
// remaining args with the consumed tokens removed.
//
// The count is optional (t85 lead loop, REQ-FF-001 amendment): a bare `-f` /
// `--factory` takes the operator-decided default fan-out
// (config.DefaultFactoryWorkers, 8), because "spin up the usual factory" is
// an entry, not a typo. `-f 4`, `-f=4`, `--factory 4`, and `--factory=4` are
// accepted as before; a SUPPLIED count that is non-numeric or zero/negative
// is an error naming the expected forms and the bare-form default. There is
// deliberately no upper bound (v1): the runtime's own concurrency cap governs
// how many workers are useful.
//
// The `--` discipline matches parseKanbanFlag, stripSpawnFlag, parseProfileFlag,
// and normalizeWorktreeFlag exactly: iterate, break at the pass-through marker,
// and forward everything from that marker onward verbatim.
func parseFactoryFlag(args []string) (workers int, enabled bool, rest []string, err error) {
	filtered := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			filtered = append(filtered, args[i:]...)
			break
		}

		var value string
		hasValue := false
		switch {
		case arg == factoryFlagLong || arg == factoryFlagShort:
			// Consume a following positional token as N. A flag or the
			// pass-through marker belongs to someone else, and reads as the
			// bare form (default count) below — not as a malformed count.
			if next := i + 1; next < len(args) && args[next] != "--" && !strings.HasPrefix(args[next], "-") {
				value, hasValue = args[next], true
				i = next
			}
		case strings.HasPrefix(arg, factoryFlagLong+"="):
			value, hasValue = strings.TrimPrefix(arg, factoryFlagLong+"="), true
		case strings.HasPrefix(arg, factoryFlagShort+"="):
			value, hasValue = strings.TrimPrefix(arg, factoryFlagShort+"="), true
		default:
			filtered = append(filtered, arg)
			continue
		}

		enabled = true
		if !hasValue {
			workers = config.DefaultFactoryWorkers
			continue
		}
		n, perr := strconv.Atoi(value)
		if perr != nil || n < 1 {
			return 0, false, nil, fmt.Errorf("%s requires a worker count of 1 or more (e.g. %s 4), or no count at all to use the default of %d workers",
				factoryFlagLong, factoryFlagLong, config.DefaultFactoryWorkers)
		}
		workers = n
	}

	return workers, enabled, filtered, nil
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

// factoryFreeSlots returns the FREE slot numbers among 1..workers — the lead
// loop's picker input (t85). A slot is free when its worker-<n> label carries
// no live claim; dead claims are pruned on the way through, and every
// registry failure fails open to "all free". The hook's lead notice renders
// the same list through kanban.FactoryFreeSlots directly.
func factoryFreeSlots(root string, workers int) []int {
	return kanban.FactoryFreeSlots(root, workers, factoryProcessAlive)
}

// factoryQueuedCards returns the number of cards waiting in the backlog queue
// under root — the lead loop's poll input. The count goes through the shared
// kanban shape (BacklogStore.QueuedCount via
// kanban.QueuedBacklogCountForRoot) so the lead notice, the kanban notice,
// and `moai todo` cannot disagree about what "waiting" means.
func factoryQueuedCards(root string) int {
	return kanban.QueuedBacklogCountForRoot(root)
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

// rejectFactoryOnCG returns the sentinel-bearing error when a factory token
// appears in a `moai cg` invocation (and the parse error when the token is
// malformed), and nil otherwise. Same premise as rejectKanbanOnCG.
func rejectFactoryOnCG(args []string) error {
	_, enabled, _, err := parseFactoryFlag(args)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	return fmt.Errorf("%s: moai cg runs a mixed backend (leader Claude, teammates GLM), "+
		"which contradicts Factory Mode's one-session / one-backend premise; "+
		"use 'moai cc --factory <N>' or 'moai glm --factory <N>' instead", factoryUnsupportedBackendSentinel)
}

// rejectConflictingModes returns an error when -k and -f appear together: the
// two tokens seed different session shapes (a four-role chain vs a numbered
// fan-out) and honoring both is not a meaningful launch.
func rejectConflictingModes(kanbanEnabled, factoryEnabled bool) error {
	if !kanbanEnabled || !factoryEnabled {
		return nil
	}
	return errors.New("-k/--kanban and -f/--factory are mutually exclusive: " +
		"-k seeds the four-role kanban chain, -f a numbered worker fan-out")
}
