package cli

// factory.go is the Factory Mode machinery behind the dedicated -f entry
// surface (`moai cc -f [N]` / `moai glm -f [N]`, t118 launcher axis, v3.1.1)
// and the factory shapes of the unified -k token (v1.2.0, kept for compat).
// A factory run is a leader session plus numbered lanes: the leader routes
// cards to the lanes over cross-session messages, and everything in this
// file exists to get that signal into the session and to keep the lane
// names unique. The -k shapes of the entry parse live in parseKanbanFlag
// (kanban.go); this file owns the -f parse (parseFactoryFlag), the merge of
// the two (parseLauncherEntry), and everything after the factory shape is
// selected.
//
// GENEALOGY (binding): the pre-3.1 "factory" flag (-f/--factory) was RENAMED
// to -k/--kanban in #1513 (7f61332ef) and drove the three-role kanban chain.
// -f briefly returned as the factory lane fan-out flag (v1.0.0, 2026-08-17)
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
	"os/exec"
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

	// factoryLaneRoleToken is the canonical `-f lane` role value
	// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-002): join the running factory as
	// the next free lane-<n>. The legacy `worker` / `agent` tokens are
	// REFUSED (REQ-RNC-003) — they survive only as detection values
	// (kanban.IsLegacyFactoryRoleValue).
	factoryLaneRoleToken = "lane"
)

// factoryFlagUsageError names every accepted -f shape. It is the error text
// for an invalid SUPPLIED value and the reference the help texts paraphrase.
// The numeric count form was REMOVED (N is unnecessary — lanes join via the
// role token). Only the lane-axis forms are taught here; the legacy
// `worker` / `agent` spellings are refused naming this text's canonical
// forms (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-002/-003).
const factoryFlagUsageError = "-f/--factory takes no argument (the factory leader), " +
	"the role token -f lane, which joins this session to a running factory as the next free lane, " +
	"or a lane label (e.g. -f lane-2) that launches exactly that one lane"

// factoryFlagParse is the -f/--factory entry parse (t118). ONE flag token,
// three shapes:
//
//	-f                → the factory leader, one lane (DefaultFactoryLeaderLanes)
//	-f lane           → join the running factory as the next free lane-<n>
//	-f lane-<n>       → exactly one additional lane, lane n — the
//	                    incremental form; the run's count is not carried, and
//	                    a count of 0 ("unknown") flows into the lane env
//
// Legacy spellings are REFUSED, any letter case (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-003/-005): `-f worker` / `-f agent` error naming `-f lane`; a
// `worker-<n>` / `agent-<n>` label errors naming the canonical `lane-<n>`.
//
// The factory carries no SPEC identifier (that is the kanban chain's -k
// SPEC-ID shape), so a SUPPLIED value that is neither the role token nor a
// lane label is an error — there is no second interpretation to silently
// fall into, and hiding the typo would be worse than naming it.
type factoryFlagParse struct {
	Enabled    bool     // -f present (any shape)
	Lanes      int      // always 0 post-N-removal; kept for the merge contract
	LaneNumber int      // n of `-f lane-<n>`; 0 otherwise
	LaneLabel  string   // the lane label exactly as typed (`lane-3`)
	LaneRole   bool     // `-f lane`: join as the next free lane
	RunID      string   // explicit --factory-run selector (MoAI-owned, pre--- only)
	Rest       []string // args with -f and its consumed value removed
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
		if value == factoryLaneRoleToken {
			p.LaneRole = true
			continue
		}
		// Legacy spellings are refused on any letter case (REQ-RNC-003,
		// REQ-RNC-005): one error line naming the canonical form, nothing
		// parsed, nothing launched.
		lowered := strings.ToLower(value)
		if kanban.IsLegacyFactoryRoleValue(lowered) {
			if n, isLabel := kanban.SplitFactoryLegacyLabel(lowered); isLabel {
				return p, fmt.Errorf("%q is the legacy lane label; use %q (or -f lane to join as the next free lane)",
					value, kanban.FactoryLaneLabel(n))
			}
			return p, fmt.Errorf("%q is the legacy role token; use -f lane to join the running factory as the next free lane", value)
		}
		if n, ok := kanban.SplitFactoryLaneLabel(value); ok {
			p.LaneNumber = n
			p.LaneLabel = value
			continue
		}
		return p, fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
	}

	return p, nil
}

// refuseLegacyEntryNames refuses the legacy spellings on the operator-name
// input path, before any branch resolves (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-004/-005/-007): a `lead` / `lead-<suffix>` name errors naming the
// canonical `leader[-…]` form; a `worker-<n>` / `agent-<n>` (or bare legacy
// role) name errors naming the canonical `lane-<n>` / `-f lane`. Nothing is
// parsed further, nothing launches, nothing is written.
func refuseLegacyEntryNames(args []string) error {
	if name, ok := parseNamedLabel(args, kanban.IsLegacyLeaderSpelling); ok {
		return fmt.Errorf("--name %q is the legacy leader spelling; use %q (leader label forms: leader, leader-<n>, leader-<run-id>)",
			name, kanban.LeaderLabel()+strings.TrimPrefix(name, "lead"))
	}
	if name, ok := parseNamedLabel(args, kanban.IsLegacyFactoryRoleValue); ok {
		if n, isLabel := kanban.SplitFactoryLegacyLabel(name); isLabel {
			return fmt.Errorf("--name %q is the legacy lane label; use %q", name, kanban.FactoryLaneLabel(n))
		}
		return fmt.Errorf("--name %q is a legacy factory spelling; the leader launches as %q and lanes join with -f lane",
			name, kanban.LeaderLabel())
	}
	return nil
}

// parseLauncherEntry is the unified -k + -f entry parse (t118): parseKanbanFlag
// for the -k shapes, parseFactoryFlag on what remains, then merge. The merge
// enforces the one-entry-token rule (-f and -k together is an error — each
// names a different mode, and a launch that carried both would have to
// silently drop one) and resolves the -f shapes into the same kanbanEntryParse
// the dispatch branches already read:
//
//   - `-f [N]` sets FactoryEnabled with Lanes (N, or
//     DefaultFactoryLeaderLanes when omitted);
//   - `-f lane-<n>` sets FactoryEnabled and appends `--name lane-<n>`
//     to Rest, desugaring into the existing lane branch (registry bump,
//     replaceNamedLabel, settings injection, per-lane agent cap) so the
//     new form and the `-k N --name lane-<i>` form share one
//     implementation. The run's count is NOT fabricated — a lane joined
//     incrementally carries 0 ("unknown"), which the lane notice degrades
//     from. A -f lane-label form plus an operator-supplied --name is a
//     conflict error: the flag value already named the lane.
//   - `-f lane` appends the next free lane label the same way.
//
// Legacy spellings (`-f worker`, `-f agent`, `worker-<n>`, `agent-<n>` as a
// -f value, and any of them as an operator --name, plus `lead` /
// `lead-<suffix>` leader names) are refused before any branch resolves
// (REQ-RNC-003/-004/-005/-007).
func parseLauncherEntry(args []string) (kanbanEntryParse, error) {
	entry, err := parseKanbanFlag(args)
	if err != nil {
		return entry, err
	}
	if err := refuseLegacyEntryNames(args); err != nil {
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
	// not only the lane form below. (The lane forms append their desugared
	// --name on top of these.)
	entry.Rest = fp.Rest
	switch {
	case fp.LaneRole:
		if operatorSuppliedName(fp.Rest) {
			return entry, fmt.Errorf("-f lane already names the role; drop the --name/-n flag (got args %v)", fp.Rest)
		}
		// Desugar into the next free lane label so the bump/liveness rules
		// and the leader's dispatch address stay one implementation. The
		// label is always the canonical lane-<n> — every writer emits the
		// new vocabulary (REQ-RNC-010).
		next := kanban.NextFactoryLaneNumber(loadFactoryRegistry(factoryRegistryPath(launchProjectRoot())), factoryProcessAlive)
		label := kanban.FactoryLaneLabel(next)
		entry.Rest = append(entry.Rest, nameFlagLong, label)
		entry.FactoryAutoNumber = true
	case fp.LaneNumber > 0:
		if operatorSuppliedName(fp.Rest) {
			return entry, fmt.Errorf("-f lane-<n> already names the lane; drop the --name/-n flag (got args %v)", fp.Rest)
		}
		entry.Rest = append(entry.Rest, nameFlagLong, fp.LaneLabel)
	default:
		// Bare -f: the factory leader. Lanes join via -f lane / -f
		// lane-<n>; the numeric count form is retired.
		entry.FactoryLanes = config.DefaultFactoryLeaderLanes
	}
	return entry, nil
}

// factoryLaneRequiresGitTree is the REQ-SD-005 lane-launch precondition: the
// session a lane joins from must be a git working tree — the repository the
// lane will carry cards for. The refusal is one line naming the git
// requirement, and it fires before any registry, factory, or queue record
// can be written.
func factoryLaneRequiresGitTree(root string) error {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "true" {
		return nil
	}
	return fmt.Errorf("factory lane launch: refused — %s is not a git working tree, and a lane session needs the git repository it will carry cards for (run the lane inside the project checkout)", root)
}

func enterSelectedFactoryRun(root, explicit string, requireActive bool) (func(), error) {
	// REQ-SD-005: a LANE join needs a git working tree, refused before any
	// write. requireActive is the lane discriminator here — the leader passes
	// false and is unaffected.
	if requireActive {
		if err := factoryLaneRequiresGitTree(root); err != nil {
			return func() {}, err
		}
	}
	if explicit == "" && !requireActive {
		return func() {}, nil
	}
	runID, err := factorymsg.ResolveActiveRun(context.Background(), root, explicit)
	if err != nil {
		return func() {}, err
	}
	if err := refuseCodexLeaderRun(root, runID); err != nil {
		return func() {}, err
	}
	// REQ-RNC-022 run boundary: a LIVE legacy record in the run's broker
	// database refuses the launch/join with the retire-and-relaunch message.
	// A dead legacy row is stale and blocks nothing.
	if value, live, err := factorymsg.LiveLegacyPeer(context.Background(), root, runID); err != nil {
		return func() {}, err
	} else if live {
		return func() {}, fmt.Errorf("factory run %s holds live legacy record %q from a binary before the leader/lane rename — "+
			"end its sessions, retire with 'moai factory runs --retire %s', then relaunch", runID, value, runID)
	}
	restore := captureEnvState(config.EnvMoaiKanbanID)
	_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	return restore, nil
}

// refuseCodexLeaderRun refuses a selected run whose recorded leader backend is
// codex (SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-010). The codex factory path is
// retired, so such a run is left over from before the retirement: joining it
// would make a lane of a leader that no longer exists, and adopting it with a
// claude leader would re-own it silently. It runs before any registry claim or
// run-row write; every other backend passes unchanged (REQ-CFR-011).
func refuseCodexLeaderRun(root, runID string) (err error) {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	var backend string
	if err := db.DB.QueryRowContext(context.Background(), `SELECT lead_backend FROM runs WHERE run_id=?`, runID).Scan(&backend); err != nil {
		return fmt.Errorf("read factory run %s leader backend: %w", runID, err)
	}
	if backend != BackendCodex {
		return nil
	}
	return fmt.Errorf("%s: factory run %s is led by %s, and the codex factory path is retired; "+
		"it cannot be joined or adopted — retire it with 'moai factory runs --retire %s', then start a new run with 'moai cc -f' or 'moai glm -f'",
		factoryUnsupportedBackendSentinel, runID, BackendCodex, runID)
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
	// The owner stamp written here is the RECORDING process, and it is correct
	// only for the instant it describes. On a replace-shaped door (syscall.Exec)
	// this process IS the session and the stamp stays correct; on the spawn and
	// pane shapes the launcher restamps to the session identity via
	// stampFactoryRunOwner once that identity exists (REQ-002b).
	return db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: backend, ManifestJSON: "{}",
		LeadPID: os.Getpid(), LeadProcessStart: homestate.CurrentProcessFingerprint(),
	})
}

// factoryBranch enumerates the dispatch outcomes, mirroring kanbanBranch.
type factoryBranch int

const (
	factoryBranchNone   factoryBranch = iota // no-op — -f absent (regardless of --name shape)
	factoryBranchLeader                      // -f N present, --name is NOT lane-shape
	factoryBranchLane                        // -f N present, --name IS lane-shape
)

// resolveFactoryBranch selects the dispatch branch from -f present and
// lane-shape --name present — the factory counterpart of
// resolveKanbanBranch's truth table:
//
//	factoryEnabled | isLane   || branch
//	----------------++--------------
//	      true      |   false   || leader   (-f N alone, or -f N --name <non-lane>)
//	      true      |   true    || lane     (-f N --name lane-<n>)
//	      false     |   any     || no-op    (--name lane-3 alone does NOT join a run)
func resolveFactoryBranch(factoryEnabled, isLane bool) factoryBranch {
	switch {
	case factoryEnabled && isLane:
		return factoryBranchLane
	case factoryEnabled && !isLane:
		return factoryBranchLeader
	default:
		return factoryBranchNone
	}
}

// parseFactoryLaneLabel reports the canonical `lane-<n>` label in args, if
// any. It matches only the lane SHAPE (kanban.SplitFactoryLaneLabel) for the
// same reason parseCompanionLabel matches only the companion shape: treating
// every named session as a lane would silently change launch behavior for
// unrelated work. A legacy `worker-<n>` / `agent-<n>` spelling never reaches
// this parser — refuseLegacyEntryNames refuses it at the entry parse
// (REQ-RNC-005); the claim keeps its own legacy refusal as the library-level
// defense for direct callers.
func parseFactoryLaneLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		_, isLane := kanban.SplitFactoryLaneLabel(candidate)
		return isLane
	})
}

// enterFactoryLeaderMode publishes the factory leader signal into the process
// environment and returns the function that puts the environment back, on the
// same prior-presence contract as enterKanbanMode.
//
// It reuses the kanban leader's run-id and leader-socket env surfaces — the run
// id names the run (and the injected `leader-<run-id>` session name) — while
// deliberately NOT setting EnvMoaiKanban or EnvMoaiKanbanLabel: those seed the
// three-role kanban chain, which a factory leader never drives. The factory
// discriminator the hook and the block-cap inject read is
// EnvMoaiFactoryWorkers.
//
// leaderLabel is the operator-supplied `leader-<run-id>` name for this session
// when there is one, and "" otherwise; its run id is adopted rather than
// replaced (see leaderRunID).
func enterFactoryLeaderMode(lanes int, leaderLabel string) func() {
	restoreLaneCount := captureEnvState(config.EnvMoaiFactoryWorkers)
	restoreID := captureEnvState(config.EnvMoaiKanbanID)
	restoreAddr := captureEnvState(config.EnvMoaiKanbanLeadAddr)
	restoreTier := seedAutonomyTier()

	_ = os.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(lanes))
	runID := leaderRunID(leaderLabel)
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
		restoreLaneCount()
	}
}

// enterFactoryLaneMode publishes the factory lane signal for a `lane-<n>`
// label and returns the function that puts the environment back, on the same
// prior-presence contract. It is enterKanbanCompanionMode's factory
// counterpart: no chain is seeded (no factory analogue of EnvMoaiKanban
// exists to seed one), the raised Stop-hook block cap reaches the session
// through EnvMoaiFactoryWorkers, and the autonomy tier and the per-lane
// agent cap are seeded because the lane is where dispatched cards are
// actually implemented. It also stamps the factory ROLE MARKER
// (config.EnvFactoryRole = config.FactoryRoleLane, name and value through
// the internal/config constants only — REQ-SD-002/-017): the stamp is what
// arms lane admission and the contract guard's widened role gate in the
// launched session, and it is restored on return like the other keys.
//
// lanes is the run's fan-out size from the lane's own entry token —
// `-f <N>` / `-k <N>` carry it explicitly, and the incremental `-f
// lane-<n>` form carries 0 ("unknown"), which the lane notice degrades
// from rather than fabricating a count.
func enterFactoryLaneMode(label string, lanes int) func() {
	restoreLabel := captureEnvState(config.EnvMoaiFactoryWorker)
	restoreMarker := captureEnvState(config.EnvFactoryRole)
	restoreLaneCount := captureEnvState(config.EnvMoaiFactoryWorkers)
	restoreTier := seedAutonomyTier()
	restoreCap := seedLaneAgentCap()

	_ = os.Setenv(config.EnvMoaiFactoryWorker, label)
	_ = os.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	_ = os.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(lanes))

	return func() {
		restoreCap()
		restoreTier()
		restoreLaneCount()
		restoreMarker()
		restoreLabel()
	}
}

// factoryLaneEntry is one registered lane: the pid of the process that
// claimed the label. The type lives in internal/kanban (factory_slots.go)
// since the t85 leader loop — the alias keeps this package's call sites and
// tests on their historical name.
type factoryLaneEntry = kanban.FactoryLaneEntry

// factoryRegistryPath / loadFactoryRegistry / saveFactoryRegistry delegate to
// the kanban registry cluster (factory_slots.go). The cluster moved out of
// this file because the SessionStart hook needs the same reads and cannot
// import this package; these delegates keep the cli surface stable.
func factoryRegistryPath(root string) string { return kanban.FactoryRegistryPath(root) }

func loadFactoryRegistry(path string) map[string]factoryLaneEntry {
	return kanban.LoadFactoryRegistry(path)
}

func saveFactoryRegistry(path string, reg map[string]factoryLaneEntry) error {
	return kanban.SaveFactoryRegistry(path, reg)
}

// factoryProcessAlive is the liveness seam; tests override it to simulate
// live and dead claims without spawning processes. The default probe itself
// lives in internal/kanban (factory_alive_*.go) since the same move.
var factoryProcessAlive = kanban.FactoryProcessAlive

// resolveFactoryLaneName returns the label this lane session should
// launch under: label itself when its number is free, or the next incremented
// number whose label is free (the "bump a conflicting number up" rule).
//
// A number is taken when the registry maps its label to a pid that is alive
// right now — a crashed or exited lane leaves a dead pid behind, and a dead
// claim frees the name so a relaunch reuses it instead of counting up
// forever. Dead entries are pruned on the way through. The final label is
// registered to this process's pid before returning.
//
// label MUST have a lane shape (canonical, or legacy — which the claim
// refuses naming the canonical lane-<n>). notes, when non-nil, receives the
// operator-visible bump line.
func resolveFactoryLaneName(root, label string, auto bool, notes io.Writer) (string, error) {
	runID := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanID))
	claim, err := kanban.ClaimFactoryLane(root, label, auto, os.Getpid(), runID, factoryProcessAlive)
	var legacyRun *kanban.FactoryLegacyRunError
	if errors.As(err, &legacyRun) {
		// REQ-RNC-022: a live legacy record of the same run refuses the join.
		return "", fmt.Errorf("claim factory lane: %s", legacyRun)
	}
	if err != nil {
		return "", fmt.Errorf("claim factory lane %s: %w", label, err)
	}
	final := claim.Label
	if notes == nil {
		return final, nil
	}
	if canonical, _ := kanban.CanonicalFactoryLabel(label); canonical != "" && canonical != final {
		_, _ = fmt.Fprintf(notes, "factory: %s is held by a live session; launching as %s\n", canonical, final)
	}
	return final, nil
}

// replaceNamedLabel returns args with the first `--name` / `-n` value equal
// to oldLabel (in any of the four forms claude accepts, before the
// pass-through marker) replaced by newLabel. It is what carries a bumped
// lane number into the session's actual name — the name is the address the
// leader dispatches to, so the bump must reach the backend argv, not just the
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
	return fmt.Errorf("%s: moai cg runs a mixed backend (CG leader Claude, CG teammates GLM), "+
		"which contradicts Factory Mode's one-session / one-backend premise; "+
		"use 'moai cc -f <N>' or 'moai glm -f <N>' instead", factoryUnsupportedBackendSentinel)
}
