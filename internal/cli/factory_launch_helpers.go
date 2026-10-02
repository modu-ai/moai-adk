package cli

// factory_launch_helpers.go holds the launcher helpers the factory entry shares across cc, glm
// and codex: the entry parse type, the leader run id and name resolution, the
// lane autonomy seeds, the launch facts, and the session-name parsers. The
// retired `-k` entry lives in launcher_retired_entries.go.
//
// @MX:NOTE: [AUTO] the factory signal travels through the PROCESS environment, not a threaded parameter
// The one production consumer of the signal is the block-cap inject, five hops
// below runCC and reached with the entry token already stripped from args.
// Threading a parameter there would change four signatures plus two test seams
// to carry something the environment already delivers to the same line — and
// the child process needs the variables anyway.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// launcherEntryParse is the launcher entry parse: the factory entry the `-f`
// leader and `-l` lane tokens select, plus the arguments left for the backend.
// The retired `-k` entry never reaches it — parseLauncherEntry refuses that
// spelling first (launcher_retired_entries.go).
type launcherEntryParse struct {
	Spec           string // the SPEC identifier a factory run records; empty on every current entry
	FactoryEnabled bool   // -f selected the factory leader
	FactoryLanes   int    // the factory count (explicit or the default)
	// FactoryLanesDeclared records that the count was operator-supplied, not a
	// parse default (SPEC-CODEX-LANE-SLOTS-001 REQ-004): the leader start
	// records the count as the run's declared capacity only when this is true,
	// and the derived-capacity marker otherwise.
	FactoryLanesDeclared bool
	FactoryRun           string // explicit --factory-run selector for mixed factory joins
	// FactoryLead is the raw `--leader <name>` target a lane join's leader
	// discovery aims at (SPEC-FACTORY-LANE-JOIN-SOCKET-001 REQ-008); empty
	// means the canonical leader label, resolved by the join gate. The field
	// carries the RAW value: the default resolution lives in one place, and
	// the parse-level refusals (legacy spelling, surface gates) have already
	// run by the time a non-empty value reaches the gate.
	FactoryLead string
	// FactoryAutoNumber marks a lane number the launcher chose itself
	// (`-l`), as opposed to one the operator typed; the claim reports legacy
	// collisions differently.
	FactoryAutoNumber bool
	// ClearPolicy carries the lane's --clear-policy selection (REQ-SD-020,
	// factory entry only): "" when none was given, which the lane reads as
	// the default clear-each.
	ClearPolicy string
	// AutoDispatchManual marks the lane's --no-auto-dispatch opt-out
	// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011, factory entry only):
	// false (the code default) launches a self-dispatch lane, true launches
	// a manual-mode lane.
	AutoDispatchManual bool
	Rest               []string // args with the entry tokens and their consumed values removed
}

// leaderRunID resolves the run id a leader session publishes, in three steps: a
// `leader-<run-id>` name the operator pasted, then a run id already standing in
// the environment, then a fresh mint.
//
// The name is no longer where a run id lives (factory.LeaderLabel is the bare
// role now), so the first step only serves an operator naming the run
// explicitly at launch. A bump number is not a run id and
// must not be adopted as one, which is why an all-digit suffix is skipped —
// `leader-1` is the second live leader on this machine, not run 1.
//
// The environment step is what replaces the name round-trip. Nothing
// functional was ever downstream of that round-trip — the per-session records
// key on the Claude session id — so the id now survives a relaunch exactly as
// far as the run-id marker does, and no new state file is introduced to carry a
// value that is only ever displayed. It is read BEFORE enterFactoryLeaderMode
// publishes this launch's id, so what it sees is the prior value or nothing.
func leaderRunID(leaderLabel string) string {
	if suffix, ok := factory.SplitLeaderLabel(leaderLabel); ok && suffix != "" && !allDigits(suffix) {
		return suffix
	}
	if runID := os.Getenv(config.EnvFactoryRunID); runID != "" {
		return runID
	}
	return factory.NewRunID()
}

// allDigits reports whether s is a non-empty run of ASCII digits — the shape a
// launcher-appended bump number has, and the one suffix leaderRunID must not read
// as a run id.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// seedAutonomyTier publishes the autonomy tier a factory session runs at, and
// returns the function that puts it back on the same prior-presence contract as
// the other launch variables.
//
// A factory run exists to run unattended: the operator launches the sessions
// once and the cards advance without being asked at every step. The tier is what
// makes that true. At fully-autonomous the synchronous vet+lint+test commit gate
// and the SubagentStop / TeammateIdle / TaskCompleted lifecycle hooks stand
// down, so a card advances without stopping for a verification tax the operator
// already accepted when they launched the factory. Leaving the variable
// unset resolves to semi-auto — config.AutonomyTier fails safe — which is the
// most-interrupted tier and the opposite of what a factory entry asks for.
//
// What the tier does NOT reach: the destructive-pattern denylist in
// internal/hook/pre_tool.go is tier-invariant, and Implementation Kickoff
// Approval is a gate the tier has no bearing on. Autonomy here buys fewer
// interruptions, never a weaker refusal.
//
// An operator who sets the variable themselves keeps it. Only an absent or
// blank value is filled in, and blank counts as absent because that is how
// config.AutonomyTier already reads it — a wrapper that exports the name with
// no value has not made a choice.
func seedAutonomyTier() func() {
	restore := captureEnvState(config.EnvAutonomyTier)
	if strings.TrimSpace(os.Getenv(config.EnvAutonomyTier)) == "" {
		_ = os.Setenv(config.EnvAutonomyTier, config.AutonomyTierFullyAutonomous)
	}
	return restore
}

// seedLaneAgentCap publishes the per-lane concurrent-subagent cap and returns
// the function that puts it back, on the same prior-presence contract as
// seedAutonomyTier (t118 launcher axis, operator-confirmed architecture: each
// factory lane runs up to DefaultLaneMaxConcurrentSubagents agents in
// parallel).
//
// A lane, not the leader: the cap is seeded on the lane branch only, where
// dispatched cards are implemented and fanned out to subagents.
// The leader keeps the runtime default, which already bounds its own turns.
//
// An operator who set the variable themselves keeps it, on the same contract
// as seedAutonomyTier: only an absent or blank value is filled, and blank
// counts as absent because that is how the sibling seed reads it too.
func seedLaneAgentCap() func() {
	restore := captureEnvState(config.EnvClaudeCodeMaxConcurrentSubagents)
	if strings.TrimSpace(os.Getenv(config.EnvClaudeCodeMaxConcurrentSubagents)) == "" {
		_ = os.Setenv(config.EnvClaudeCodeMaxConcurrentSubagents,
			strconv.Itoa(config.DefaultLaneMaxConcurrentSubagents))
	}
	return restore
}

// leaderRegistryPath returns the liveness-checked leader-name registry's home.
//
// It is a SEPARATE file from the factory lane registry because the two
// namespaces are separate: a leader name and a lane label can never collide,
// and sharing one file would make each role's launches contend on the other's
// writes for no benefit. The shape and the machinery are identical.
func leaderRegistryPath(root string) string {
	return filepath.Join(factory.RuntimeStateDirForRoot(root), "leads.json")
}

// resolveLeaderName returns the label this leader session should launch under:
// label itself when it is free, or the next free number (`leader-1`,
// `leader-2`, ...) when a live session already holds it — the leader sibling
// of resolveFactoryLaneName, on the same registry machinery.
//
// The bump is what the one-machine-one-run policy costs and what makes it
// survivable. Two runs on one machine both want the name `leader`, and the
// peer listing distinguishes same-named sessions only by an opaque reference
// — so without the bump a dispatch addressed to `leader` is ambiguous exactly
// when a second run is what made it ambiguous. Numbering the second leader
// keeps every session addressable by name alone, which is the property the
// whole naming policy is for.
//
// A LIVE registry entry named `lead` / `lead-<suffix>` (a pre-rename leader)
// never blocks: one notice names the entry and the relaunch step, and the
// launch proceeds under `leader` (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-025). A
// dead legacy entry is pruned by the claim like any dead claim.
//
// Best-effort throughout, like the lane claim: an unreadable or unwritable
// registry degrades to using the label as supplied. The launch must never block
// on a name claim.
func resolveLeaderName(root, label string, notes io.Writer) string {
	noteLegacyLeaderRegistryEntries(root, notes)
	return claimName(leaderRegistryPath(root), label, factory.LeaderNumberLabel, notes)
}

// noteLegacyLeaderRegistryEntries writes one notice per LIVE legacy leader
// entry in the leader-name registry (leads.json), naming the entry and the
// end-and-relaunch
// step. Best-effort: notes may be nil, and the registry read is fail-open.
func noteLegacyLeaderRegistryEntries(root string, notes io.Writer) {
	if notes == nil {
		return
	}
	reg := loadFactoryRegistry(leaderRegistryPath(root))
	seen := map[string]bool{}
	for name, entry := range reg {
		if !factory.IsLegacyLeaderSpelling(name) || seen[name] {
			continue
		}
		if entry.PID <= 0 || !factoryProcessAlive(entry.PID) {
			continue // dead legacy entries are pruned by the claim, not noticed
		}
		seen[name] = true
		_, _ = fmt.Fprintf(notes, "kanban: registry entry %q is a leader session from before the leader/lane rename; end that session and relaunch it — launching as %s\n", name, factory.LeaderLabel())
	}
}

// claimName returns the label to launch under after claiming it in the
// liveness-checked registry at path: label itself when free, or the first
// bump(n) that is, counting up from 1.
//
// A label is taken when the registry maps it to a pid that is alive right now
// — a crashed or exited session leaves a dead pid behind, and a dead claim
// frees the name so a relaunch reuses it instead of counting up forever. Dead
// entries are pruned on the way through. The final label is registered to this
// process's pid before returning.
//
// bump may be nil, which means this label has no numbered form to fall back to;
// the label is then claimed as supplied whether or not it is held. That is the
// pre-existing behavior for a label that fails its own shape check, preserved
// rather than tightened: the launch path is not where a malformed name should
// start failing.
//
// The loop is a single copy so that its callers differ only in which registry
// they consult and how they render a number. Keeping one copy is how they avoid
// drifting apart — the same reasoning parseNamedLabel records for its own side.
func claimName(path, label string, bump func(int) string, notes io.Writer) string {
	reg := loadFactoryRegistry(path)

	// Prune dead claims first so they neither block a name nor accumulate.
	for l, e := range reg {
		if e.PID <= 0 || !factoryProcessAlive(e.PID) {
			delete(reg, l)
		}
	}

	final := label
	if bump != nil {
		for n := 1; ; n++ {
			claim, taken := reg[final]
			if !taken || claim.PID <= 0 || !factoryProcessAlive(claim.PID) {
				break
			}
			final = bump(n)
		}
		if final != label && notes != nil {
			// The note is best-effort operator guidance; the SessionStart
			// notice is the reliable surface for the final name.
			_, _ = fmt.Fprintf(notes, "kanban: %s is held by a live session; launching as %s\n", label, final)
		}
	}

	reg[final] = factoryLaneEntry{
		PID:          os.Getpid(),
		RegisteredAt: time.Now().UTC().Format(time.RFC3339),
	}
	_ = saveFactoryRegistry(path, reg)
	return final
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

// exportFactoryLaunchFacts publishes the launch facts a launched session cannot
// observe for itself, and returns the function that puts the environment back
// on the same prior-presence contract the other enter*Mode helpers use.
//
// It does NOT write a session record, and that absence is the point. The
// launcher cannot key a record correctly under any implementation: the
// identifier it would need belongs to a process that does not exist yet, so
// the pre-change write landed under whichever session last wrote the
// project-wide single-identifier sidecar slot (card t221's surface, untouched
// by this change) — in practice the
// LAUNCHING session, and never the launched one by design. The record is now
// written by the session's own SessionStart, the first actor that holds the
// described session's identifier (SPEC-KANBAN-RECORD-SESSION-KEY-001
// REQ-KRS-001/002, decision D-6). There is exactly one writer, and it is the
// session.
//
// One fact travels here: the BACKEND, because nothing in the session's
// environment names it and inferring it from ANTHROPIC_BASE_URL would be a
// guess dressed as a measurement — that variable is set by the GLM path but is
// settable by anyone. The first parameter is the SPEC identifier the entry
// parse carries; it is no longer exported (no launcher publishes it), and the
// parameter stays only so the Claude, GLM and Codex call sites keep their shape.
// This is the one launch-facts function: the Claude and GLM launchers and both
// Codex entries all call it.
//
// The card-identifier override is deliberately NOT exported here. It is the
// operator's or the leader's to set, and the launch environment carries it
// through unchanged (config.EnvFactoryCard); the session reads it directly.
//
// Callers must defer the returned function so it also runs on the error path.
func exportFactoryLaunchFacts(_, backend string) func() {
	restoreBackend := captureEnvState(config.EnvFactoryBackend)

	if backend != "" {
		_ = os.Setenv(config.EnvFactoryBackend, backend)
	}

	return restoreBackend
}

// The tokens claude uses to name a session. moai RECOGNIZES them; it never
// consumes them — the value has to reach claude unchanged.
const (
	nameFlagLong  = "--name"
	nameFlagShort = "-n"
)

// parseLeaderLabel reports the `leader-<run-id>` label in args, if any.
//
// It exists for the launcher to read back a run id the operator embedded in the
// session name, so the leader branch can adopt it instead of minting a second
// one (see leaderRunID).
func parseLeaderLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		_, isLeader := factory.SplitLeaderLabel(candidate)
		return isLeader
	})
}

// parseNamedLabel returns the first `--name` / `-n` value before the
// pass-through marker that accept admits.
//
// The scan is shared by parseLeaderLabel and parseFactoryLaneLabel because the
// two differ only in which shape they admit. The four name forms claude accepts
// and the `--` discipline are identical for both, and keeping one copy per role
// is how the two would drift — which is the class of defect this whole change is
// repairing.
func parseNamedLabel(args []string, accept func(candidate string) bool) (label string, ok bool) {
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

		if accept(candidate) {
			return candidate, true
		}
	}
	return "", false
}

// operatorSuppliedName reports whether the operator named this session
// themselves, in any of the four forms claude accepts (`--name v`, `--name=v`,
// `-n v`, `-n=v`), before the pass-through marker.
//
// It is deliberately NOT a shape match: a leader the operator named
// `board-watch` must read as "a name is present". The question here is only
// whether a name exists at all, because the operator's choice wins over any
// moai-supplied one.
//
// The `--` discipline matches operatorSuppliedSettings and the launcher entry
// parse: nothing past the marker is read — a `--name` after `--` is the
// backend's argument, already destined for claude unchanged.
func operatorSuppliedName(args []string) bool {
	for _, arg := range args {
		switch {
		case arg == "--":
			return false
		case arg == nameFlagLong || arg == nameFlagShort:
			return true
		case strings.HasPrefix(arg, nameFlagLong+"="):
			return true
		case strings.HasPrefix(arg, nameFlagShort+"="):
			return true
		}
	}
	return false
}

// leaderNameArgs returns the `--name leader` pair to append to a leader session's
// argv, or nil when the operator named the session themselves.
//
// The injection exists because claude keeps an EXPLICIT name across /clear and
// discards an AI-generated title. A lane is already explicitly named — the
// launcher claims its `lane-<n>` label — so only the leader, launched as a bare
// `moai cc -f`, would lose its identity on every clear and have to be renamed
// by hand.
//
// The name is the bare role (factory.LeaderLabel), so unlike its prior form this
// needs nothing from the environment and has one self-gate rather than two: the
// operator's own name wins. The absent-run-id gate is gone with the run id it
// guarded — there is no longer a `lead-` degenerate form to avoid emitting (the
// legacy spelling is refused outright, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-007).
// A name held by a live session is bumped by the CALLER (appendLeaderName), not
// here, so this stays a pure function of args and the bump can consult the
// registry the launch path already owns.
func leaderNameArgs(args []string) []string {
	if operatorSuppliedName(args) {
		return nil
	}
	return []string{nameFlagLong, factory.LeaderLabel()}
}

// appendLeaderName appends the leader session's `--name` pair to args, bumped past
// any live claim, and returns the result unchanged when the operator supplied a
// name of their own.
//
// It is the leader's counterpart to the lane branch's
// resolveFactoryLaneName + replaceNamedLabel pair, and exists as one helper
// because both leader branches (cc and glm) need exactly this and would
// otherwise carry two copies of it — the class of drift the shared
// parseNamedLabel already exists to prevent.
//
// The bumped value must reach the backend argv: the session name is the address
// the operator and the peers dispatch to, so a name resolved but not injected
// leaves everyone addressing a session that answers to something else.
func appendLeaderName(args []string, root string, notes io.Writer) ([]string, string) {
	nameArgs := leaderNameArgs(args)
	if len(nameArgs) == 0 {
		// The operator named the session themselves; that name is already in
		// argv and is the one peers address. Report it so the title registered
		// downstream is the name the session actually answers to, rather than
		// the role moai would have chosen.
		name, _ := parseNamedLabel(args, func(string) bool { return true })
		return args, name
	}
	nameArgs[1] = resolveLeaderName(root, nameArgs[1], notes)
	return append(args, nameArgs...), nameArgs[1]
}

// exportLeaderSessionName publishes the leader's resolved session name to the child
// session's environment and returns the restore func its callers defer.
//
// The name and the TITLE are two separate registrations in Claude Code: `--name`
// makes the session addressable, and a title has to be set on its own or a
// generated one takes the slot (issue #1596). Only the launcher knows the
// resolved name and only a hook can set a title, so the value crosses that gap
// through the environment the child already inherits.
//
// An empty name is a no-op that still returns a usable restore func, so a caller
// can defer the result unconditionally.
func exportLeaderSessionName(name string) func() {
	if name == "" {
		return func() {}
	}
	restore := captureEnvState(config.EnvFactoryLeadName)
	_ = os.Setenv(config.EnvFactoryLeadName, name)
	return restore
}
