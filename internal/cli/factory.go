package cli

// factory.go is the Factory Mode machinery behind the dedicated entry
// surface (`moai cc -f` starts the leader, `moai cc -l` joins as a lane, and
// likewise on glm; t118 launcher axis, v3.1.1, with the lane entry added by
// SPEC-LAUNCHER-ENTRY-FLAGS-001) and the factory shapes of the unified -k token
// (v1.2.0, kept for compat until its removal).
// A factory run is a leader session plus numbered lanes: the leader routes
// cards to the lanes over cross-session messages, and everything in this
// file exists to get that signal into the session and to keep the lane
// names unique. The -k shapes of the entry parse live in parseKanbanFlag
// (kanban.go); this file owns the -f/-l parse (parseFactoryFlag), the merge of
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
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryUnsupportedBackendSentinel is the machine-greppable marker on the
// `moai cg` rejection, mirroring kanbanUnsupportedBackendSentinel: a mixed
// leader/teammate backend contradicts factory mode's one-session /
// one-backend premise, so the invocation is rejected rather than adapted.
const factoryUnsupportedBackendSentinel = "FACTORY_MODE_UNSUPPORTED_BACKEND"

// The entry tokens. `-f` / `--factory` start the factory leader; `-l` /
// `--lane` join the running factory as the next free lane (SPEC-LAUNCHER-
// ENTRY-FLAGS-001). `--leader` names the leader session a lane join's
// discovery targets (SPEC-FACTORY-LANE-JOIN-SOCKET-001 REQ-008) and composes
// with the lane entry only. The former `-l` short of `--leader` is retired on
// cc and glm; leadFlagShort survives only because the Codex entry parse
// (codex_factory.go) still reads it until M3 replaces that parse.
const (
	factoryFlagLong  = "--factory"
	factoryFlagShort = "-f"
	laneFlagLong     = "--lane"
	laneFlagShort    = "-l"
	leadFlagLong     = "--leader"
	leadFlagShort    = "-l"

	// factoryLaneRoleToken is the canonical lane role value
	// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-002), kept as the lane role's name
	// for the Codex parse and the role pin; cc and glm no longer accept it
	// as an `-f` value. The legacy `worker` / `agent` tokens survive only as
	// detection values (kanban.IsLegacyFactoryRoleValue).
	factoryLaneRoleToken = "lane"

	// clearPolicyFlag is the lane clear-policy selection token (REQ-SD-020):
	// `--clear-policy <value>` stamps the selected policy into the lane
	// session's environment through config.EnvFactoryClearPolicy.
	clearPolicyFlag = "--clear-policy"

	// noAutoDispatchFlag is the lane auto-dispatch opt-out token
	// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011): a boolean flag, no
	// value. Absent, a lane launch is a SELF-DISPATCH lane — the default is
	// recorded in code, and the flag is the only way to select the manual
	// mode.
	noAutoDispatchFlag = "--no-auto-dispatch"
)

// factoryFlagUsageError is the refusal for any value after -f/--factory: the
// flag takes none (bare -f starts the factory leader) and a lane joins with
// -l / --lane. It is the reference the help texts paraphrase; the supplied
// value is appended by the caller.
const factoryFlagUsageError = "-f/--factory takes no argument (bare -f starts the factory leader); a lane joins with -l or --lane"

// laneFlagArgumentError, laneFlagNameError and the entry-token error are the
// one-line refusals of the lane entry (REQ-002): an argument after -l/--lane,
// an operator --name beside it, and a second entry token.
const (
	laneFlagArgumentError = "-l/--lane takes no argument (it joins the running factory as the next free lane); --leader <name> selects the leader session"
	laneFlagNameError     = "-l/--lane already names the role; drop the --name/-n flag"
	entryTokenConflict    = "%s and %s are two entry tokens; a launch carries at most one"
	laneNameUnderFactory  = "--name lane-<n> does not select a lane under -f/--factory (-f starts the factory leader); the lane entry is -l or --lane"
	leaderNeedsLaneEntry  = "--leader names the leader session a lane join targets; it composes with -l or --lane only"
)

// factoryFlagParse is the entry parse of the factory surface. Two entry
// tokens, no value on either:
//
//	-f / --factory   → the factory leader, one lane (DefaultFactoryLeaderLanes)
//	-l / --lane      → join the running factory as the next free lane-<n>
//
// Everything else a launch could carry after -f (a number, a role token, a
// lane label) is REFUSED with one line naming -l (REQ-005/-007), and -l
// itself refuses any argument or a second entry token (REQ-002). The legacy
// role spellings (`worker`, `agent`, any letter case) keep their own refusal
// line, which also names -l (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-003/-005).
//
// The factory carries no SPEC identifier (that is the kanban chain's -k
// SPEC-ID shape), so there is no second interpretation for a supplied value
// to silently fall into, and hiding a typo would be worse than naming it.
type factoryFlagParse struct {
	Enabled     bool   // -f or -l present (either entry token)
	Lanes       int    // always 0 post-N-removal; kept for the merge contract
	LaneNumber  int    // set by the Codex entry parse only (`-f lane-<n>`); 0 on cc/glm
	LaneLabel   string // set by the Codex entry parse only; empty on cc/glm
	LaneRole    bool   // -l/--lane: join as the next free lane
	RunID       string // explicit --factory-run selector (MoAI-owned, pre--- only)
	ClearPolicy string // --clear-policy value; a lane-only selection (REQ-SD-020)
	// NoAutoDispatch marks the --no-auto-dispatch opt-out
	// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011); a lane-only boolean.
	NoAutoDispatch bool
	Lead           string   // `--leader <name>`: which leader session discovery targets (lane joins only)
	Rest           []string // args with the entry token and its consumed value removed
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
	sawFactoryFlag := false

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

		// --leader names WHICH leader session a lane join's discovery
		// targets (SPEC-FACTORY-LANE-JOIN-SOCKET-001 REQ-008) — a target,
		// not this session's own name. The legacy spelling is refused here
		// with the same canonical-form shape refuseLegacyEntryNames applies
		// to --name (AC-010): nothing parsed further, nothing launched,
		// nothing written. Its former short `-l` is the lane entry now.
		if arg == leadFlagLong || strings.HasPrefix(arg, leadFlagLong+"=") {
			switch {
			case strings.HasPrefix(arg, leadFlagLong+"="):
				p.Lead = strings.TrimPrefix(arg, leadFlagLong+"=")
			case i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-"):
				i++
				p.Lead = args[i]
			default:
				return p, fmt.Errorf("%s requires a leader label", leadFlagLong)
			}
			if kanban.IsLegacyLeaderSpelling(p.Lead) {
				return p, fmt.Errorf("%s %q is the legacy leader spelling; use %q (leader label forms: leader, leader-<n>, leader-<run-id>)",
					leadFlagLong, p.Lead, kanban.LeaderLabel()+strings.TrimPrefix(p.Lead, "lead"))
			}
			continue
		}

		if arg == noAutoDispatchFlag || strings.HasPrefix(arg, noAutoDispatchFlag+"=") {
			if strings.HasPrefix(arg, noAutoDispatchFlag+"=") {
				return p, fmt.Errorf("%s takes no value; it is a boolean opt-out", noAutoDispatchFlag)
			}
			p.NoAutoDispatch = true
			continue
		}

		if arg == clearPolicyFlag || strings.HasPrefix(arg, clearPolicyFlag+"=") {
			var value string
			if strings.HasPrefix(arg, clearPolicyFlag+"=") {
				value = strings.TrimPrefix(arg, clearPolicyFlag+"=")
			} else if i+1 < len(args) && args[i+1] != "--" {
				i++
				value = args[i]
			} else {
				return p, fmt.Errorf("%s requires a clear policy value (%s, %s, or %s)",
					clearPolicyFlag, config.FactoryClearPolicyEach, config.FactoryClearPolicyWhenFull, config.FactoryClearPolicyRelaunch)
			}
			switch value {
			case config.FactoryClearPolicyEach, config.FactoryClearPolicyWhenFull, config.FactoryClearPolicyRelaunch:
				p.ClearPolicy = value
			default:
				return p, fmt.Errorf("%s takes %s, %s, or %s (got %q)",
					clearPolicyFlag, config.FactoryClearPolicyEach, config.FactoryClearPolicyWhenFull, config.FactoryClearPolicyRelaunch, value)
			}
			continue
		}

		// -l / --lane: the lane entry. It takes no argument — a following
		// positional token or an `=`-joined value is refused (REQ-002), so a
		// parser that read `--lane <x>` as `--lane` plus a forwarded value
		// cannot pass. A following flag or the pass-through marker belongs to
		// someone else and is left in place.
		if arg == laneFlagShort || arg == laneFlagLong ||
			strings.HasPrefix(arg, laneFlagShort+"=") || strings.HasPrefix(arg, laneFlagLong+"=") {
			if strings.Contains(arg, "=") || (i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-")) {
				return p, errors.New(laneFlagArgumentError)
			}
			p.Enabled = true
			p.LaneRole = true
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

		sawFactoryFlag = true
		p.Enabled = true
		if hasValue {
			// No `-f <value>` form is left. A legacy spelling (any letter
			// case, REQ-RNC-003/-005) keeps its own line naming the legacy
			// kind and the lane entry; every other value (a number, `lane`,
			// a lane label, anything else) takes the one usage refusal.
			lowered := strings.ToLower(value)
			if kanban.IsLegacyFactoryRoleValue(lowered) {
				if n, isLabel := kanban.SplitFactoryLegacyLabel(lowered); isLabel {
					return p, fmt.Errorf("%q is the legacy lane label; lanes are named %q now and join with -l", value, kanban.FactoryLaneLabel(n))
				}
				return p, fmt.Errorf("%q is the legacy role token; a lane joins with -l", value)
			}
			return p, fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
		}
	}

	// One entry token per launch: -f and -l each name a different role, and a
	// launch carrying both would have to drop one silently.
	if sawFactoryFlag && p.LaneRole {
		return p, fmt.Errorf(entryTokenConflict, "-f/--factory", "-l/--lane")
	}

	// The --leader surface gates (REQ-004, REQ-008): the flag names a TARGET
	// for a lane join's discovery, so it composes with the lane entry only —
	// a leader entry carrying it names two things at once, and carrying it
	// beside --factory-run names two different selectors. Both refuse before
	// anything launches.
	if p.Lead != "" {
		if p.RunID != "" {
			return p, fmt.Errorf("%s and --factory-run name two different selectors; carry one", leadFlagLong)
		}
		if !p.LaneRole {
			return p, errors.New(leaderNeedsLaneEntry)
		}
	}
	// The clear policy is a LANE selection (REQ-SD-020: a Claude-harness
	// lane launch). Carrying it on a non-factory launch or on the leader
	// shape has no defined meaning, so the parser refuses it rather than
	// silently swallowing the flag — the same typo-honesty as the -f
	// value refusal above.
	if p.ClearPolicy != "" {
		if !p.Enabled {
			return p, fmt.Errorf("%s selects a factory lane's clear policy; it composes with -l or --lane only", clearPolicyFlag)
		}
		if !p.LaneRole {
			return p, fmt.Errorf("%s selects a lane's clear policy; the factory leader takes none", clearPolicyFlag)
		}
	}
	// The auto-dispatch opt-out is a LANE selection too (REQ-TCD-011), on
	// the same typo-honesty terms as the clear policy: a leader shape or a
	// non-factory launch refuses it rather than silently swallowing the flag.
	if p.NoAutoDispatch {
		if !p.Enabled {
			return p, fmt.Errorf("%s selects a factory lane's dispatch mode; it composes with -l or --lane only", noAutoDispatchFlag)
		}
		if !p.LaneRole {
			return p, fmt.Errorf("%s selects a lane's dispatch mode; the factory leader takes none", noAutoDispatchFlag)
		}
	}

	return p, nil
}

// laneDispatchSelection resolves the dispatch value a lane launch stamps
// into config.EnvFactoryAutoDispatch (REQ-TCD-011): the opt-out selects the
// manual mode; the code default is auto-dispatch.
func laneDispatchSelection(entry kanbanEntryParse) string {
	if entry.AutoDispatchManual {
		return config.FactoryDispatchManual
	}
	return config.FactoryDispatchAuto
}

// refuseLegacyEntryNames refuses the legacy spellings on the operator-name
// input path, before any branch resolves (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-004/-005/-007): a `lead` / `lead-<suffix>` name errors naming the
// canonical `leader[-…]` form; a `worker-<n>` / `agent-<n>` (or bare legacy
// role) name errors naming the canonical `lane-<n>` / `-l`. Nothing is
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
		return fmt.Errorf("--name %q is a legacy factory spelling; the leader launches as %q and lanes join with -l",
			name, kanban.LeaderLabel())
	}
	return nil
}

// parseLauncherEntry is the unified -k + -f/-l entry parse (t118):
// parseKanbanFlag for the -k shapes, parseFactoryFlag on what remains, then
// merge. The merge enforces the one-entry-token rule (-k beside -f or -l is an
// error — each names a different mode, and a launch that carried both would
// have to silently drop one) and resolves the factory shapes into the same
// kanbanEntryParse the dispatch branches already read:
//
//   - bare `-f` sets FactoryEnabled with the leader's one-lane default
//     (DefaultFactoryLeaderLanes); a lane-shaped --name beside it is refused
//     naming -l, because the lane branch has no -f trigger any more;
//   - `-l` / `--lane` sets FactoryEnabled and appends the next free
//     `--name lane-<n>` to Rest, desugaring into the existing lane branch
//     (registry bump, replaceNamedLabel, settings injection, per-lane agent
//     cap) so the lane entry and the `-k N --name lane-<i>` form share one
//     implementation. The run's count is NOT fabricated — a lane joined
//     incrementally carries 0 ("unknown"), which the lane notice degrades
//     from. -l plus an operator-supplied --name is a conflict error: the
//     entry token already names the role.
//
// Legacy spellings of an operator --name (`worker-<n>`, `agent-<n>`, bare
// legacy roles, and `lead` / `lead-<suffix>` leader names) are refused before
// any branch resolves (REQ-RNC-004/-005/-007).
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
		token := "-f/--factory"
		if fp.LaneRole {
			token = "-l/--lane"
		}
		return entry, fmt.Errorf(entryTokenConflict, "-k/--kanban", token)
	}

	entry.FactoryEnabled = true
	entry.FactoryRun = fp.RunID
	entry.FactoryLead = fp.Lead
	entry.ClearPolicy = fp.ClearPolicy
	entry.AutoDispatchManual = fp.NoAutoDispatch
	// The stripped args always become the launch args — for every entry
	// shape, not only the lane form below. (The lane form appends its
	// desugared --name on top of these.)
	entry.Rest = fp.Rest
	if fp.LaneRole {
		if operatorSuppliedName(fp.Rest) {
			return entry, errors.New(laneFlagNameError)
		}
		// Desugar into the next free lane label so the bump/liveness rules
		// and the leader's dispatch address stay one implementation. The
		// label is always the canonical lane-<n> — every writer emits the
		// new vocabulary (REQ-RNC-010).
		next := kanban.NextFactoryLaneNumber(loadFactoryRegistry(factoryRegistryPath(launchProjectRoot())), factoryProcessAlive)
		label := kanban.FactoryLaneLabel(next)
		entry.Rest = insertBeforePassthrough(entry.Rest, nameFlagLong, label)
		entry.FactoryAutoNumber = true
		return entry, nil
	}
	// Bare -f: the factory leader. A lane-shaped --name beside it used to
	// select the lane branch; the lane's only entry is -l now, so the name is
	// refused rather than read as a lane.
	if _, isLane := parseFactoryLaneLabel(fp.Rest); isLane {
		return entry, errors.New(laneNameUnderFactory)
	}
	entry.FactoryLanes = config.DefaultFactoryLeaderLanes
	return entry, nil
}

// insertBeforePassthrough adds tokens to args ahead of the first `--` marker,
// so a launcher-desugared flag never lands in the child's pass-through region
// (where parseFactoryLaneLabel stops reading and the session would launch as a
// leader instead of the lane the entry token named).
func insertBeforePassthrough(args []string, tokens ...string) []string {
	for i, arg := range args {
		if arg == "--" {
			out := make([]string, 0, len(args)+len(tokens))
			out = append(out, args[:i]...)
			out = append(out, tokens...)
			return append(out, args[i:]...)
		}
	}
	return append(args, tokens...)
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

// enterSelectedFactoryRun resolves the run a lane joins or a leader starts.
// timing, when non-nil, records the active-run resolution step of the codex
// lane launch's pre-exec phase (SPEC-CODEX-LANE-SLOTS-001 REQ-012); the
// cc/glm twins pass nil and measure nothing.
func enterSelectedFactoryRun(root, explicit string, requireActive bool, timing *factoryLaunchTiming) (func(), error) {
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
	endResolve := timing.begin(factoryStepRunResolve)
	runID, err := factorymsg.ResolveActiveRun(context.Background(), root, explicit)
	endResolve()
	if err != nil {
		return func() {}, err
	}
	if !requireActive {
		// An existing run may be joined across backends, but its leader
		// remains its owner until the run is retired.
		if err := refuseCodexLeaderRun(root, runID); err != nil {
			return func() {}, err
		}
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

// discoverFactoryLeader is the leader-discovery seam; tests override it to
// stage verified leaders without spawning processes (the
// installFactoryLaunchSeam pattern).
var discoverFactoryLeader = discovery.DiscoverLeader

// ambiguousFactoryLeaderSentinel is the multi-leader fail-closed refusal
// (REQ-005): two or more verified live leaders is genuine ambiguity, and —
// the same prohibition REQ-014 applies to picking among active records —
// selection among live leaders is forbidden. The refusal names every
// candidate with its run id and mutates nothing.
const ambiguousFactoryLeaderSentinel = "AMBIGUOUS_FACTORY_LEADER"

// noActiveFactorySentinel is the resolver's record-absence answer; the lane
// gate matches it to decide whether discovery may run.
const noActiveFactorySentinel = "NO_ACTIVE_FACTORY"

// enterFactoryLaneRun joins a factory lane to the run selected by explicit,
// falling back to verified leader discovery on the record-absence refusal
// (SPEC-FACTORY-LANE-JOIN-SOCKET-001 REQ-001). leadTarget names which leader
// session discovery aims at; empty means the canonical leader label.
//
// @MX:ANCHOR: [AUTO] enterFactoryLaneRun — the single shared lane-join gate every launcher (cc/glm/codex twin) passes (REQ-010)
// @MX:REASON: the discovery+resume path lives HERE and nowhere else; a launcher-private copy of this seam is the AC-013 defect, and the re-entry into enterSelectedFactoryRun is what keeps a discovered run indistinguishable from a lead-recorded one
//
// The path is the single join point every lane entry (cc / glm / codex twin)
// passes (REQ-010): on NO_ACTIVE_FACTORY with no explicit selection it probes
// for a live leader of this project; zero verified → the original refusal
// stands; two or more → fail closed naming them; exactly one → the dedicated
// resume writer restores that leader's run (REQ-004) and the gate is
// RE-ENTERED, so a discovered run resolves exactly as one the lead recorded
// itself — every downstream invariant (run id into MOAI_KANBAN_ID, lane
// claim, hook ValidateActiveRun) is exercised, not bypassed. On the discovery
// path the verified leader's own name is exported to the child as
// MOAI_KANBAN_LEAD_NAME (REQ-009); ordinary joins export no leader name and
// are unchanged.
//
// An explicit --factory-run is a decision, not an absence (REQ-006): the
// resolver's answer stands and discovery never runs on it.
// enterFactoryLaneRun joins a factory lane to the run selected by explicit,
// falling back to verified leader discovery on the record-absence refusal
// (SPEC-FACTORY-LANE-JOIN-SOCKET-001 REQ-001). leadTarget names which leader
// session discovery aims at; empty means the canonical leader label. timing,
// when non-nil, records the join-gate and active-run-resolution steps of the
// codex lane launch's pre-exec phase (SPEC-CODEX-LANE-SLOTS-001 REQ-012);
// the cc/glm twins pass nil and measure nothing.
func enterFactoryLaneRun(root, explicit, leadTarget string, timing *factoryLaunchTiming) (func(), error) {
	restore, err := enterSelectedFactoryRun(root, explicit, true, timing)
	if err == nil {
		return restore, nil
	}
	if explicit != "" || err.Error() != noActiveFactorySentinel {
		return restore, err
	}
	target := leadTarget
	if target == "" {
		target = kanban.LeaderLabel()
	}
	verified, derr := discoverFactoryLeader(context.Background(), root, target)
	if derr != nil {
		return restore, derr
	}
	switch {
	case len(verified) == 0:
		// Discovery verified no live leader for this project — the refusal
		// stands, now backed by a probe instead of a record (REQ-001).
		return restore, err
	case len(verified) > 1:
		return restore, fmt.Errorf("%s: %s", ambiguousFactoryLeaderSentinel, discovery.DescribeVerifiedLeaders(verified))
	}
	leader := verified[0]
	if rerr := resumeDiscoveredRun(root, leader); rerr != nil {
		return restore, fmt.Errorf("resume discovered run %s: %w", leader.RunID, rerr)
	}
	restoreName := captureEnvState(config.EnvMoaiKanbanLeadName)
	_ = os.Setenv(config.EnvMoaiKanbanLeadName, leader.Name)
	reentered, jerr := enterSelectedFactoryRun(root, "", true, timing)
	if jerr != nil {
		restoreName()
		return restore, jerr
	}
	return func() { reentered(); restoreName() }, nil
}

// resumeDiscoveredRun is the join path's ONLY write: the dedicated resume
// writer stamps the verified leader's probe-measured identity (REQ-004).
// recordFactoryRunStart is deliberately NOT reused — it stamps the calling
// lane, and the lane's exit would then retire a live lead's run.
//
// @MX:NOTE: [AUTO] the run id comes from the VERIFIED LEADER's own MOAI_KANBAN_ID env read by the probe — never minted or defaulted; an unreadable env declines the candidate (fail-closed), because the run id is the address of the broker the lead already speaks on
func resumeDiscoveredRun(root string, leader discovery.VerifiedLeader) error {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	return db.ResumeRun(context.Background(), leader.RunID, leader.PID, leader.ProcessStart, leader.Basis)
}

// refuseCodexLeaderRun prevents a different leader from adopting an active
// Codex-led run. Lane joins are handled separately and may cross backends.
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
	return fmt.Errorf("factory run %s is led by %s and cannot be adopted by another leader; retire it with 'moai factory runs --retire %s' before starting a new run",
		runID, BackendCodex, runID)
}

// recordFactoryRunStart records a run's start. declaredLanes is the run's
// declared lane capacity (SPEC-CODEX-LANE-SLOTS-001 REQ-004): the
// operator-supplied count, or homestate.LaneCapacityDerived when the leader
// start carried none — the marker the join reads as capacity-open.
func recordFactoryRunStart(root, runID, backend, specID string, declaredLanes int) (err error) {
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
		LaneCapacity: declaredLanes,
	})
}

// factoryDeclaredLanes resolves a leader start's declared lane capacity from
// the entry parse (SPEC-CODEX-LANE-SLOTS-001 REQ-004): the operator-supplied
// count when one was typed (`-k N`), the derived-capacity marker when the
// count is a parse default (`-f` bare — the count-less leader). The
// distinction is the whole policy: a defaulted count never becomes a declared
// bound.
func factoryDeclaredLanes(entry kanbanEntryParse) int {
	if entry.FactoryLanesDeclared && entry.FactoryLanes >= 1 {
		return entry.FactoryLanes
	}
	return homestate.LaneCapacityDerived
}

// factoryBranch enumerates the dispatch outcomes, mirroring kanbanBranch.
type factoryBranch int

const (
	factoryBranchNone   factoryBranch = iota // no-op — no factory entry (regardless of --name shape)
	factoryBranchLeader                      // factory entry present, --name is NOT lane-shape
	factoryBranchLane                        // factory entry present, --name IS lane-shape
)

// resolveFactoryBranch selects the dispatch branch from a factory entry
// present and lane-shape --name present — the factory counterpart of
// resolveKanbanBranch's truth table. The lane-shape name is not typed by the
// operator on the -f/-l surface: -l desugars it from the claim, and a
// lane-shaped --name beside -f is refused at the parse, so the lane row is
// reached by -l (and, until M5a removes it, by `-k N --name lane-<i>`):
//
//	factoryEnabled | isLane   || branch
//	----------------++--------------
//	      true      |   false   || leader   (-f alone, or -f --name <non-lane>)
//	      true      |   true    || lane     (-l, desugared to --name lane-<n>)
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
// `-k <N>` carries it explicitly, and the incremental `-l` lane entry
// carries 0 ("unknown"), which the lane notice degrades from rather than
// fabricating a count.
//
// clearPolicy is the lane's --clear-policy selection (REQ-SD-020): the
// value stamped into config.EnvFactoryClearPolicy for the SessionStart rule
// and the `complete` output to read. "" (no selection) is stamped too — the
// stamp always overwrites, so a policy inherited from an outer session can
// never leak into a lane launched without one. Codex-harness lanes take no
// policy and pass "" (REQ-SD-020).
//
// dispatch is the lane's auto-dispatch selection
// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011): config.FactoryDispatchAuto
// (the code default) or config.FactoryDispatchManual (--no-auto-dispatch).
// The stamp always overwrites, like the policy's — nothing leaks from an
// outer session.
func enterFactoryLaneMode(label string, lanes int, clearPolicy string, dispatch string) func() {
	restoreLabel := captureEnvState(config.EnvMoaiFactoryWorker)
	restoreMarker := captureEnvState(config.EnvFactoryRole)
	restoreLaneCount := captureEnvState(config.EnvMoaiFactoryWorkers)
	restorePolicy := captureEnvState(config.EnvFactoryClearPolicy)
	restoreDispatch := captureEnvState(config.EnvFactoryAutoDispatch)
	restoreTier := seedAutonomyTier()
	restoreCap := seedLaneAgentCap()

	_ = os.Setenv(config.EnvMoaiFactoryWorker, label)
	_ = os.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	_ = os.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(lanes))
	_ = os.Setenv(config.EnvFactoryClearPolicy, clearPolicy)
	_ = os.Setenv(config.EnvFactoryAutoDispatch, dispatch)

	return func() {
		restoreCap()
		restoreTier()
		restoreDispatch()
		restorePolicy()
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
// the -k token (v1.2.0 shapes), the -f leader, or the -l lane entry — and nil
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
		"use 'moai cc -f' / 'moai glm -f' for the leader or 'moai cc -l' / 'moai glm -l' for a lane instead", factoryUnsupportedBackendSentinel)
}
