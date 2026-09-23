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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
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

	// factoryWorkerRoleToken is the `-f worker` role value: join the running
	// factory as the next free worker-<n> — the symmetric form across the
	// cc/glm/codex launchers.
	factoryWorkerRoleToken = "worker"

	// factoryLegacyAgentRoleToken is the pre-rename spelling of the role
	// token. It still parses (keep-alias) and behaves exactly like `-f
	// worker`; the launch prints a deprecation hint naming the new form.
	factoryLegacyAgentRoleToken = "agent"
)

// factoryFlagUsageError names every accepted -f shape. It is the error text
// for an invalid SUPPLIED value and the reference the help texts paraphrase.
// The numeric count form was REMOVED (N is unnecessary — workers join via the
// role token). Only the worker-axis spellings are taught here; the legacy
// `agent` / `lane-<n>` spellings still parse and print their own hint.
const factoryFlagUsageError = "-f/--factory takes no argument (the factory lead), " +
	"the worker role token (e.g. -f worker) that joins this session to a running factory as the next free worker, " +
	"or a worker label (e.g. -f worker-2) that launches exactly that one worker"

// factoryFlagParse is the -f/--factory entry parse (t118). ONE flag token,
// three shapes:
//
//	-f                → the factory lead, one worker (DefaultFactoryLeadWorkers)
//	-f worker         → join the running factory as the next free worker-<n>
//	-f worker-<n>     → exactly one additional worker, worker n — the
//	                    incremental form; the run's count is not carried, and
//	                    a count of 0 ("unknown") flows into the worker env
//
// Legacy spellings (keep-alias): `-f agent` parses as `-f worker`, and `-f
// lane-<n>` as `-f worker-<n>`; the launch prints a deprecation hint.
//
// The factory carries no SPEC identifier (that is the kanban chain's -k
// SPEC-ID shape), so a SUPPLIED value that is neither the role token nor a
// worker label is an error — there is no second interpretation to silently
// fall into, and hiding the typo would be worse than naming it.
type factoryFlagParse struct {
	Enabled          bool     // -f present (any shape)
	Workers          int      // always 0 post-N-removal; kept for the merge contract
	WorkerNumber     int      // n of `-f worker-<n>` (or legacy `-f lane-<n>`); 0 otherwise
	WorkerLabel      string   // the worker label exactly as typed (`worker-3`, or legacy `lane-3`)
	WorkerRole       bool     // `-f worker` (or legacy `-f agent`): join as the next free worker
	LegacyAgentToken bool     // the role token was typed as the legacy `agent`
	RunID            string   // explicit --factory-run selector (MoAI-owned, pre--- only)
	Rest             []string // args with -f and its consumed value removed
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

		if arg == "--factory-run" || strings.HasPrefix(arg, "--factory-run=") {
			if strings.HasPrefix(arg, "--factory-run=") {
				p.RunID = strings.TrimPrefix(arg, "--factory-run=")
			} else if i+1 < len(args) && args[i+1] != "--" {
				i++
				p.RunID = args[i]
			} else {
				return p, fmt.Errorf("--factory-run requires a run id")
			}
			if strings.TrimSpace(p.RunID) == "" {
				return p, fmt.Errorf("--factory-run requires a run id")
			}
			continue
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
		if value == factoryWorkerRoleToken || value == factoryLegacyAgentRoleToken {
			p.WorkerRole = true
			p.LegacyAgentToken = value == factoryLegacyAgentRoleToken
			continue
		}
		if n, ok := kanban.SplitFactoryLaneLabel(value); ok {
			p.WorkerNumber = n
			p.WorkerLabel = value
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
//   - `-f worker-<n>` sets FactoryEnabled and appends `--name worker-<n>`
//     to Rest, desugaring into the existing worker branch (registry bump,
//     replaceNamedLabel, settings injection, per-worker agent cap) so the
//     new form and the `-k N --name worker-<i>` form share one
//     implementation. The run's count is NOT fabricated — a worker joined
//     incrementally carries 0 ("unknown"), which the worker notice degrades
//     from. A -f worker-label form plus an operator-supplied --name is a
//     conflict error: the flag value already named the worker.
//   - `-f worker` appends the next free worker label the same way.
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
	entry.FactoryRun = fp.RunID
	// The stripped args always become the launch args — for every -f shape,
	// not only the lane form below. (The lane and agent forms append their
	// desugared --name on top of these.)
	entry.Rest = fp.Rest
	switch {
	case fp.WorkerRole:
		if operatorSuppliedName(fp.Rest) {
			return entry, fmt.Errorf("-f worker already names the role; drop the --name/-n flag (got args %v)", fp.Rest)
		}
		// Desugar into the next free worker label so the bump/liveness rules
		// and the lead's dispatch address stay one implementation. The
		// legacy `-f agent` spelling desugars into the legacy agent label,
		// which the claim canonicalizes to worker-<n> and reports as
		// deprecated (resolveFactoryWorkerName is the one hint site).
		next := kanban.NextFactoryWorkerNumber(loadFactoryRegistry(factoryRegistryPath(launchProjectRoot())), factoryProcessAlive)
		label := kanban.FactoryLaneLabel(next)
		if fp.LegacyAgentToken {
			label = kanban.FactoryAgentLabel(next)
		}
		entry.Rest = append(entry.Rest, nameFlagLong, label)
		entry.FactoryAutoNumber = true
	case fp.WorkerNumber > 0:
		if operatorSuppliedName(fp.Rest) {
			return entry, fmt.Errorf("-f worker-<n> already names the worker; drop the --name/-n flag (got args %v)", fp.Rest)
		}
		// The label travels as typed: a legacy `lane-<n>` is canonicalized
		// (with its deprecation hint) at the claim, exactly like a typed
		// --name lane-<n>.
		entry.Rest = append(entry.Rest, nameFlagLong, fp.WorkerLabel)
	default:
		// Bare -f: the factory lead. Workers join via -f worker / -f
		// worker-<n>; the numeric count form is retired.
		entry.FactoryWorkers = config.DefaultFactoryLeadWorkers
	}
	return entry, nil
}

func enterSelectedFactoryRun(root, explicit string, requireActive bool) (func(), error) {
	if explicit == "" && !requireActive {
		return func() {}, nil
	}
	runID, err := factorymsg.ResolveActiveRun(context.Background(), root, explicit)
	if err != nil {
		return func() {}, err
	}
	restore := captureEnvState(config.EnvMoaiKanbanID)
	_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	return restore, nil
}

func recordFactoryRunStart(root, runID, backend, specID string) (err error) {
	if err := kanban.RecordFactoryRunStart(root, runID, backend, specID); err != nil {
		return err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	return db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: backend, ManifestJSON: "{}",
	})
}

func stripFactoryRunFlag(head []string) ([]string, string, error) {
	rest := make([]string, 0, len(head))
	runID := ""
	for i := 0; i < len(head); i++ {
		a := head[i]
		switch {
		case a == "--factory-run":
			if i+1 >= len(head) {
				return nil, "", fmt.Errorf("--factory-run requires a run id")
			}
			i++
			runID = head[i]
		case strings.HasPrefix(a, "--factory-run="):
			runID = strings.TrimPrefix(a, "--factory-run=")
		default:
			rest = append(rest, a)
		}
	}
	if strings.TrimSpace(runID) == "" && len(rest) != len(head) {
		return nil, "", fmt.Errorf("--factory-run requires a run id")
	}
	return rest, runID, nil
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
//	factoryEnabled | isLane   || branch
//	----------------++--------------
//	      true      |   false   || lead     (-f N alone, or -f N --name <non-worker>)
//	      true      |   true    || worker   (-f N --name worker-<n>)
//	      false     |   any     || no-op    (--name worker-3 alone does NOT join a run)
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

// parseFactoryLaneLabel reports the worker label in args, if any — the
// canonical `worker-<n>` or a legacy `lane-<n>` / `agent-<n>` spelling (the
// `-f worker` desugared name flows through the same --name channel the
// incremental worker form uses).
// It matches only the worker SHAPES (kanban.SplitFactoryLaneLabel /
// kanban.SplitFactoryAgentLabel), for the same reason parseCompanionLabel
// matches only the companion shape: treating every named session as a
// worker would silently change launch behavior for unrelated work.
func parseFactoryLaneLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		if _, isLane := kanban.SplitFactoryLaneLabel(candidate); isLane {
			return true
		}
		_, isAgent := kanban.SplitFactoryAgentLabel(candidate)
		return isAgent
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

// enterFactoryWorkerMode publishes the factory lane signal for a `worker-<n>`
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
// worker-<n>` form carries 0 ("unknown"), which the lane notice degrades
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
// label MUST have a worker shape, canonical or legacy; the claim records the
// canonical worker-<n>. notes, when non-nil, receives the operator-visible
// lines: the deprecation hint for a legacy spelling, and the bump line.
func resolveFactoryWorkerName(root, label string, auto bool, notes io.Writer) (string, error) {
	claim, err := kanban.ClaimFactoryWorker(root, label, auto, os.Getpid(), factoryProcessAlive)
	var collision *kanban.FactoryLegacyCollisionError
	if errors.As(err, &collision) {
		// Legacy agent-/lane- labels share the worker number space; an
		// operator-typed number held by one is refused by name, never moved.
		return "", fmt.Errorf("claim factory worker %s: %s (a live session launched under the legacy spelling; "+
			"legacy agent-<n> and lane-<n> labels share the worker number space) — "+
			"pick another number with -f worker-<n>, or use -f worker to take the next free one", label, collision)
	}
	if err != nil {
		return "", fmt.Errorf("claim factory worker %s: %w", label, err)
	}
	final := claim.Label
	if notes == nil {
		return final, nil
	}
	if len(claim.SkippedLegacy) > 0 {
		_, _ = fmt.Fprintf(notes, "factory: skipped number(s) held by legacy label(s) %s — legacy agent-<n> and lane-<n> labels share the worker number space; launching as %s\n",
			strings.Join(claim.SkippedLegacy, ", "), final)
	}
	// Both notes are best-effort operator guidance; the SessionStart worker
	// notice is the reliable surface for the final name.
	if kanban.IsLegacyFactoryLabel(label) {
		_, _ = fmt.Fprintf(notes, "factory: %s; launching as %s\n", legacyFactorySpellingHint(label), final)
	}
	if canonical, _ := kanban.CanonicalFactoryLabel(label); canonical != final {
		_, _ = fmt.Fprintf(notes, "factory: %s is held by a live session; launching as %s\n", canonical, final)
	}
	return final, nil
}

// legacyFactorySpellingHint names the deprecated spelling a legacy worker
// label came from and the worker-axis form that replaces it. The legacy
// spellings keep working (keep-alias); the hint is how they retire without a
// silent break.
func legacyFactorySpellingHint(label string) string {
	if _, isAgent := kanban.SplitFactoryAgentLabel(label); isAgent {
		return "`-f agent` (and the agent-<n> label) is a deprecated spelling — use `-f worker`"
	}
	return "`-f lane-<n>` / `--name lane-<n>` is a deprecated spelling — use `-f worker-<n>`"
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
