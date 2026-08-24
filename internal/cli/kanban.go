package cli

// kanban.go is the Kanban Mode entry surface on the two single-backend
// launchers. Kanban Mode adds no runtime: it seeds a session whose
// orchestrator drives a plan -> run -> sync chain, and everything in
// this file exists to get that signal into the session.
//
// @MX:NOTE: [AUTO] the kanban signal travels through the PROCESS environment, not a threaded parameter
// The one production consumer of the signal is the block-cap inject, five hops
// below runCC and reached with the kanban token already stripped from args.
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
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// The entry tokens. `-k` is unbound on cc / glm / cg; the commands that do bind
// it (`doctor config dump`, `state`) are distinct and unaffected.
const (
	kanbanFlagLong  = "--kanban"
	kanbanFlagShort = "-k"
)

// kanbanFlagUsageError names every accepted -k shape. It is the error text for
// an invalid SUPPLIED value and the reference the help texts paraphrase.
const kanbanFlagUsageError = "-k/--kanban takes a SPEC identifier (e.g. -k SPEC-X-001), " +
	"a lane count of 1 or more (e.g. -k 4) for Factory Mode, " +
	"or no argument for the plain kanban lead"

// kanbanUnsupportedBackendSentinel is the machine-greppable marker on the
// `moai cg` rejection. A mixed leader/teammate backend contradicts the
// one-session / one-backend / one-chain premise, so the invocation is rejected
// rather than adapted.
const kanbanUnsupportedBackendSentinel = "KANBAN_MODE_UNSUPPORTED_BACKEND"

// kanbanEntryParse is the unified -k/--kanban entry parse (v1.2.0): ONE flag
// token selects one of two session shapes —
//
//	-k                  → the three-role kanban chain (lead branch)
//	-k SPEC-ID          → the kanban chain tied to a SPEC
//	-k N (N ≥ 1)        → Factory Mode with N numbered lanes
//	-k --name lane-<n>  → Factory Mode as lane n; the count defaults to
//	                       config.DefaultFactoryWorkers because the
//	                       lane-shape name selected the factory with no
//	                       count supplied (a bare -k alone is the kanban lead,
//	                       so a count-less FACTORY lead does not exist)
//
// The numeric-positional discriminator is unambiguous in both directions: a
// SPEC identifier is never a bare integer, and a bare integer is never a
// meaningful SPEC identifier. FactoryEnabled implies KanbanEnabled (the -k
// token was present); the launcher branches on FactoryEnabled first.
type kanbanEntryParse struct {
	Spec           string   // non-numeric positional — the kanban SPEC identifier
	KanbanEnabled  bool     // -k present (any shape)
	FactoryEnabled bool     // -k selected the factory (numeric count or lane-shape name)
	FactoryWorkers int      // the factory count (explicit or the default)
	Rest           []string // args with -k and its consumed value removed
}

// parseKanbanFlag extracts --kanban / -k and its optional value from args.
//
// The SPEC identifier is optional: its absence means the chain begins at
// plan-phase from the operator's first prompt, which is a valid entry, not an
// error. A token that looks like a flag is never consumed as the identifier.
// A SUPPLIED value that is numeric but invalid (zero/negative) is an error —
// it names a count the operator clearly intended, and silently treating it as
// a kanban SPEC identifier would hide the typo.
//
// The `--` discipline matches stripSpawnFlag, parseProfileFlag, and
// normalizeWorktreeFlag exactly: iterate, break at the pass-through marker, and
// forward everything from that marker onward verbatim. Both commands set
// DisableFlagParsing, so a cobra flag registration would be silently inert —
// this manual parser is the only mechanism available.
func parseKanbanFlag(args []string) (p kanbanEntryParse, err error) {
	p.Rest = make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			p.Rest = append(p.Rest, args[i:]...)
			break
		}

		var value string
		hasValue := false
		joined := false
		switch {
		case arg == kanbanFlagLong || arg == kanbanFlagShort:
			// Consume a following positional token as the value. A flag or
			// the pass-through marker belongs to someone else.
			if next := i + 1; next < len(args) && args[next] != "--" && !strings.HasPrefix(args[next], "-") {
				value, hasValue = args[next], true
				i = next
			}
		case strings.HasPrefix(arg, kanbanFlagLong+"="), strings.HasPrefix(arg, kanbanFlagShort+"="):
			// The `=`-joined form exists for the FACTORY COUNT only (v1.2.0);
			// the kanban SPEC identifier never had one.
			value = strings.TrimPrefix(strings.TrimPrefix(arg, kanbanFlagShort+"="), kanbanFlagLong+"=")
			hasValue, joined = true, true
		default:
			p.Rest = append(p.Rest, arg)
			continue
		}

		p.KanbanEnabled = true
		if !hasValue {
			continue
		}
		if n, perr := strconv.Atoi(value); perr == nil {
			if n < 1 {
				return p, fmt.Errorf("%s, got %q", kanbanFlagUsageError, value)
			}
			p.FactoryEnabled = true
			p.FactoryWorkers = n
			continue
		}
		if joined {
			return p, fmt.Errorf("%s, got %q", kanbanFlagUsageError, value)
		}
		p.Spec = value
	}

	// A bare -k plus a lane-shape --name selected the factory with no count
	// supplied — the count-less factory entry, resolved to the default here.
	// parseFactoryLaneLabel stops at the pass-through marker like this
	// parser does, so a `-- --name lane-1` passthrough never selects it.
	if p.KanbanEnabled && !p.FactoryEnabled {
		if _, isLane := parseFactoryLaneLabel(args); isLane {
			p.FactoryEnabled = true
			p.FactoryWorkers = config.DefaultFactoryWorkers
		}
	}

	return p, nil
}

// @MX:ANCHOR: [AUTO] the deferred restore is a correctness requirement, not hygiene
// @MX:REASON: os.Setenv is process-global; an unrestored mutation leaves MOAI_KANBAN set for every later test in the internal/cli binary (making the block-cap negative control pass or fail by execution order) and lets a later production re-exec inherit kanban semantics it was never given
//
// enterKanbanMode publishes the kanban signal into the process environment
// and returns the function that puts the environment back.
//
// The restore returns each variable to its PRIOR PRESENCE, not merely its prior
// value: a variable that was unset is unset again rather than set to "". The
// distinction is observable — os.Getenv cannot see it but os.LookupEnv can, and
// downstream readers treat presence as the signal.
//
// Callers must defer the returned function so it also runs on the error path; a
// restore that only runs on success is the same leak with a narrower trigger.
// leadLabel is the operator-supplied `lead-<run-id>` name for this session when
// there is one, and "" otherwise. Its run id is ADOPTED rather than replaced —
// see leadRunID.
func enterKanbanMode(specID, leadLabel string) func() {
	restoreKanban := captureEnvState(config.EnvMoaiKanban)
	restoreSpec := captureEnvState(config.EnvMoaiKanbanSpec)
	restoreID := captureEnvState(config.EnvMoaiKanbanID)
	restoreAddr := captureEnvState(config.EnvMoaiKanbanLeadAddr)
	restoreTier := seedAutonomyTier()

	_ = os.Setenv(config.EnvMoaiKanban, "1")
	runID := leadRunID(leadLabel)
	_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	if specID != "" {
		_ = os.Setenv(config.EnvMoaiKanbanSpec, specID)
	}
	// SPEC-FACTORY-BOOTSTRAP-001 M3: surface a leader socket path for the
	// SessionStart hook to print. The actual messaging-substrate address is a
	// run-phase concern; this conventional path-shaped value gives the notice
	// a non-empty, grep-friendly address line. t118 socket scheme: the kanban
	// run's address lives under its own directory, never the factory's.
	_ = os.Setenv(config.EnvMoaiKanbanLeadAddr, kanban.LeaderSocketPath(runID))

	return func() {
		restoreTier()
		restoreAddr()
		restoreID()
		restoreSpec()
		restoreKanban()
	}
}

// leadRunID resolves the run id a lead session publishes, in three steps: a
// legacy `lead-<run-id>` name the operator is still pasting, then a run id
// already standing in the environment, then a fresh mint.
//
// The name is no longer where a run id lives (kanban.LeadLabel is the bare
// role now), so the first step is a MIGRATION path and nothing else: it keeps
// an operator who copied an old launch line landing on the run they meant
// rather than silently opening a second one. A bump number is not a run id and
// must not be adopted as one, which is why an all-digit suffix is skipped —
// `lead-1` is the second live lead on this machine, not run 1.
//
// The environment step is what replaces the name round-trip. Nothing
// functional was ever downstream of that round-trip — the board and the
// per-session records key on the Claude session id, and companion names stopped
// carrying the run id at t56 — so the id now survives a relaunch exactly as far
// as MOAI_KANBAN_ID does, and no new state file is introduced to carry a value
// that is only ever displayed. It is read BEFORE enterKanbanMode publishes this
// launch's id, so what it sees is the prior value or nothing.
func leadRunID(leadLabel string) string {
	if suffix, ok := kanban.SplitLeadLabel(leadLabel); ok && suffix != "" && !allDigits(suffix) {
		return suffix
	}
	if runID := os.Getenv(config.EnvMoaiKanbanID); runID != "" {
		return runID
	}
	return kanban.NewRunID()
}

// allDigits reports whether s is a non-empty run of ASCII digits — the shape a
// launcher-appended bump number has, and the one suffix leadRunID must not read
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

// seedAutonomyTier publishes the autonomy tier Kanban Mode runs at, and returns
// the function that puts it back on the same prior-presence contract as the
// other kanban variables.
//
// Kanban Mode exists to run unattended: the operator launches the sessions once
// and the board advances without being asked at every step. The tier is what
// makes that true. At fully-autonomous the synchronous vet+lint+test commit gate
// and the SubagentStop / TeammateIdle / TaskCompleted lifecycle hooks stand
// down, so a card crosses a column without stopping for a verification tax the
// operator already accepted when they launched the board. Leaving the variable
// unset resolves to semi-auto — config.AutonomyTier fails safe — which is the
// most-interrupted tier and the opposite of what -k asks for.
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
// lane — kanban companion or factory worker — runs up to
// DefaultLaneMaxConcurrentSubagents agents in parallel).
//
// A lane, not the lead: the cap is seeded on the companion / worker branches
// only, where dispatched cards are implemented and fanned out to subagents.
// The lead keeps the runtime default, which already bounds its own turns.
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

// enterKanbanCompanionMode publishes the companion signal for a label — the
// bare role under the naming policy, or a bumped `<role>-<n>` — and returns
// the function that puts the environment back, on the same prior-presence
// contract as enterKanbanMode.
//
// It deliberately does NOT set config.EnvMoaiKanban: that variable seeds the
// chain, and only the lead drives the chain. What the companion shares with the
// lead is the raised Stop-hook block cap, which the inject reads from either
// variable.
//
// It also deliberately does NOT set config.EnvMoaiKanbanID: the run id is
// lead-owned state, and publishing one on the companion side — historically
// derived from the label suffix — is the root of the t21 incident class (a
// lead's MOAI_KANBAN_ID disagreeing with a live companion's suffix produced
// announcements for a run id no live session carried). Under the bare-role
// policy no companion surface carries a run id at all, so the disagreement has
// nothing to disagree about. Nothing on the companion path reads the variable.
func enterKanbanCompanionMode(label string) func() {
	restoreLabel := captureEnvState(config.EnvMoaiKanbanLabel)
	// A companion is where the work actually lands — it plans, implements,
	// reviews, and commits. Seeding the tier on the lead alone would leave every
	// interruption exactly where it already was, because the lead does not
	// commit code. The same reasoning seeds the per-lane agent cap: the
	// companion's subagent fan-out is what the cap bounds.
	restoreTier := seedAutonomyTier()
	restoreCap := seedLaneAgentCap()

	_ = os.Setenv(config.EnvMoaiKanbanLabel, label)

	return func() {
		restoreCap()
		restoreTier()
		restoreLabel()
	}
}

// companionRegistryPath returns the liveness-checked companion-name registry's
// home, the sibling of the factory worker registry it shares its machinery
// with. A project's kanban companions claim names here so the bump has
// something to consult — Claude Code owns the session-name namespace but
// offers moai no query into it, so the claim set is moai's own state.
func companionRegistryPath(root string) string {
	return filepath.Join(root, ".moai", "state", "kanban", "companions.json")
}

// resolveCompanionName returns the label this companion session should launch
// under: label itself when it is free, or the next free number for its role
// (the "bump a conflicting name up" rule, the companion sibling of
// resolveFactoryWorkerName).
//
// A label is taken when the registry maps it to a pid that is alive right now
// — a crashed or exited companion leaves a dead pid behind, and a dead claim
// frees the name so a relaunch reuses it instead of counting up forever. Dead
// entries are pruned on the way through. The bumped candidates are
// `<role>-<n>` regardless of the label's own suffix shape, so a held legacy
// `plan-abc123` bumps to `plan-1` (never a second hyphen). The final label is
// registered to this process's pid before returning.
//
// label MUST have the companion shape (kanban.SplitCompanionLabel); the caller
// has already branched on it. Best-effort throughout: an unreadable or
// unwritable registry degrades to using the label as supplied, exactly like
// every other launch-path state write — the launch must never block on it.
// notes, when non-nil, receives the operator-visible bump line.
func resolveCompanionName(root, label string, notes io.Writer) string {
	role, _, ok := kanban.SplitCompanionLabel(label)
	if !ok {
		return claimName(companionRegistryPath(root), label, nil, notes)
	}
	return claimName(companionRegistryPath(root), label, func(n int) string {
		return kanban.CompanionNumberLabel(role, n)
	}, notes)
}

// leadRegistryPath returns the liveness-checked lead-name registry's home.
//
// It is a SEPARATE file from the companion registry because the two namespaces
// are separate: a lead named `lead` and a companion named `plan` can never
// collide, and sharing one file would make each mode's launches contend on the
// other's writes for no benefit. The shape and the machinery are identical.
func leadRegistryPath(root string) string {
	return filepath.Join(root, ".moai", "state", "kanban", "leads.json")
}

// resolveLeadName returns the label this lead session should launch under:
// label itself when it is free, or the next free number (`lead-1`, `lead-2`,
// ...) when a live session already holds it — the lead sibling of
// resolveCompanionName, on the same registry machinery.
//
// The bump is what the one-machine-one-run policy costs and what makes it
// survivable. Two runs on one machine both want the name `lead`, and the peer
// listing distinguishes same-named sessions only by an opaque reference — so
// without the bump a dispatch addressed to `lead` is ambiguous exactly when a
// second run is what made it ambiguous. Numbering the second lead keeps every
// session addressable by name alone, which is the property the whole naming
// policy is for.
//
// Best-effort throughout, like its companion sibling: an unreadable or
// unwritable registry degrades to using the label as supplied. The launch must
// never block on a name claim.
func resolveLeadName(root, label string, notes io.Writer) string {
	return claimName(leadRegistryPath(root), label, kanban.LeadNumberLabel, notes)
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
// The loop is shared by the companion and lead resolvers because they differ
// only in which registry they consult and how they render a number. Keeping one
// copy is how the two avoid drifting apart — the same reasoning parseNamedLabel
// records for its own side.
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

	reg[final] = factoryWorkerEntry{
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

// exportKanbanLaunchFacts publishes the launch facts a launched session cannot
// observe for itself, and returns the function that puts the environment back
// on the same prior-presence contract the other enter*Mode helpers use.
//
// It does NOT write the kanban record, and that absence is the point. The
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
// Two facts travel here. The BACKEND, because nothing in the session's
// environment names it and inferring it from ANTHROPIC_BASE_URL would be a
// guess dressed as a measurement — that variable is set by the GLM path but is
// settable by anyone. The SPEC identifier, because enterKanbanMode publishes it
// for the kanban lead alone, so a companion or a factory lane had no way to
// learn it. Re-exporting it on the lead path is harmless: the value is the same
// one enterKanbanMode set, and the restores unwind in reverse order.
//
// The card-identifier override is deliberately NOT exported here. It is the
// operator's or the lead's to set, and the launch environment carries it
// through unchanged (config.EnvMoaiKanbanCard); the session reads it directly.
//
// Callers must defer the returned function so it also runs on the error path.
func exportKanbanLaunchFacts(specID, backend string) func() {
	restoreBackend := captureEnvState(config.EnvMoaiKanbanBackend)
	restoreSpec := captureEnvState(config.EnvMoaiKanbanSpec)

	if backend != "" {
		_ = os.Setenv(config.EnvMoaiKanbanBackend, backend)
	}
	// An absent SPEC is not exported as an empty value: the session reads
	// PRESENCE, and an empty export would announce a SPEC that is not there.
	if specID != "" {
		_ = os.Setenv(config.EnvMoaiKanbanSpec, specID)
	}

	return func() {
		restoreSpec()
		restoreBackend()
	}
}

// The tokens claude uses to name a session. moai RECOGNIZES them; it never
// consumes them — the value has to reach claude unchanged.
const (
	nameFlagLong  = "--name"
	nameFlagShort = "-n"
)

// parseCompanionLabel reports the companion label in args, if any.
//
// It matches only the companion SHAPE (kanban.SplitCompanionLabel) because the
// alternative discriminators are worse. Treating every named session as a
// companion would
// silently raise the Stop-hook block cap from 8 to 200 for unrelated work, and a
// state file the lead writes and companions read buys nothing here beyond one
// more file to keep consistent.
//
// The `--` discipline matches parseKanbanFlag, stripSpawnFlag, parseProfileFlag
// and normalizeWorktreeFlag: iterate, break at the pass-through marker, and read
// nothing beyond it. args is returned to the caller untouched.
func parseCompanionLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		_, _, isCompanion := kanban.SplitCompanionLabel(candidate)
		return isCompanion
	})
}

// parseLeadLabel reports the `lead-<run-id>` label in args, if any.
//
// It is parseCompanionLabel's counterpart and exists for the launcher to read
// back a run id the operator embedded in the session name, so the lead branch
// can adopt it instead of minting a second one (see leadRunID).
//
// A lead label never satisfies the companion discriminator — RoleLead is absent
// from kanban.CompanionRoles — so the two parsers can never both match the same
// name, and recognizing a lead name here cannot reroute the session down the
// companion branch.
func parseLeadLabel(args []string) (label string, ok bool) {
	return parseNamedLabel(args, func(candidate string) bool {
		_, isLead := kanban.SplitLeadLabel(candidate)
		return isLead
	})
}

// parseNamedLabel returns the first `--name` / `-n` value before the
// pass-through marker that accept admits.
//
// The scan is shared by parseCompanionLabel and parseLeadLabel because the two
// differ only in which shape they admit. The four name forms claude accepts and
// the `--` discipline are identical for both, and keeping one copy per role is
// how the two would drift — which is the class of defect this whole change is
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
// It is deliberately NOT parseCompanionLabel: that function matches the
// companion SHAPE, so a lead the operator named `board-watch` reads as "no
// name present" there. The question here is only whether a name exists at all,
// because the operator's choice wins over any moai-supplied one.
//
// The `--` discipline matches operatorSuppliedSettings and parseKanbanFlag:
// nothing past the marker is read — a `--name` after `--` is the backend's
// argument, already destined for claude unchanged.
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

// leadNameArgs returns the `--name lead` pair to append to a lead session's
// argv, or nil when the operator named the session themselves.
//
// The injection exists because claude keeps an EXPLICIT name across /clear and
// discards an AI-generated title. A companion is already explicitly named — the
// SessionStart notice prints `--name <role>` and the operator pastes it — so
// only the lead, launched as a bare `moai cc -k`, loses its identity on every
// clear and has to be renamed by hand.
//
// The name is the bare role (kanban.LeadLabel), so unlike its prior form this
// needs nothing from the environment and has one self-gate rather than two: the
// operator's own name wins. The absent-run-id gate is gone with the run id it
// guarded — there is no longer a `lead-` degenerate form to avoid emitting.
//
// A name held by a live session is bumped by the CALLER (appendLeadName), not
// here, so this stays a pure function of args and the bump can consult the
// registry the launch path already owns.
func leadNameArgs(args []string) []string {
	if operatorSuppliedName(args) {
		return nil
	}
	return []string{nameFlagLong, kanban.LeadLabel()}
}

// appendLeadName appends the lead session's `--name` pair to args, bumped past
// any live claim, and returns the result unchanged when the operator supplied a
// name of their own.
//
// It is the lead's counterpart to the companion branch's
// resolveCompanionName + replaceNamedLabel pair, and exists as one helper
// because all four lead branches (kanban and factory, on cc and on glm) need
// exactly this and would otherwise carry four copies of it — the class of
// drift the shared parseNamedLabel already exists to prevent.
//
// The bumped value must reach the backend argv: the session name is the address
// the operator and the peers dispatch to, so a name resolved but not injected
// leaves everyone addressing a session that answers to something else.
func appendLeadName(args []string, root string, notes io.Writer) []string {
	nameArgs := leadNameArgs(args)
	if len(nameArgs) == 0 {
		return args
	}
	nameArgs[1] = resolveLeadName(root, nameArgs[1], notes)
	return append(args, nameArgs...)
}

// rejectKanbanOnCG returns the sentinel-bearing error when a kanban token
// appears in a `moai cg` invocation, and nil otherwise.
func rejectKanbanOnCG(args []string) error {
	p, err := parseKanbanFlag(args)
	if err != nil {
		return err
	}
	// v1.2.0: the FACTORY shapes of -k (a numeric count, or a worker-shape
	// name) are the factory rejection's to answer (rejectFactoryOnCG fires
	// right after this one); only the plain kanban shapes carry the kanban
	// sentinel here.
	if !p.KanbanEnabled || p.FactoryEnabled {
		return nil
	}
	return fmt.Errorf("%s: moai cg runs a mixed backend (leader Claude, teammates GLM), "+
		"which contradicts Kanban Mode's one-session / one-backend / one-chain premise; "+
		"use 'moai cc --kanban' or 'moai glm --kanban' instead", kanbanUnsupportedBackendSentinel)
}

// kanbanBranch enumerates the three dispatch outcomes of the §A.2 truth table.
type kanbanBranch int

const (
	kanbanBranchNone      kanbanBranch = iota // no-op — -k absent (regardless of --name shape)
	kanbanBranchLead                          // -k present, --name is NOT companion-shape
	kanbanBranchCompanion                     // -k present, --name IS companion-shape
)

// resolveKanbanBranch selects the dispatch branch from the combination of -k
// present and companion-shape --name present.
//
// This is the four-row truth table at spec.md §A.2 (REQ-FB-001, REQ-FB-002):
//
//	kanbanEnabled | isCompanion || branch
//	---------------++--------------
//	      true      |    false     || lead       (-k alone, or -k --name <non-companion>)
//	      true      |    true      || companion  (-k --name <role>, or a bumped <role>-<n>)
//	      false     |    false     || no-op      (--name <non-companion>, or no --name)
//	      false     |    true      || no-op      (--name <companion-shape> alone — BREAKING from 94025ce0a)
//
// The two !kanbanEnabled rows collapse to no-op because `isCompanion` is
// consulted only when -k is present (spec.md §A.2.1 / AC-FB-027): a companion-
// shape --name alone, which entered companion mode under 94025ce0a, is
// reclassified as a no-op by REQ-FB-001's no-`-k` clause.
func resolveKanbanBranch(kanbanEnabled, isCompanion bool) kanbanBranch {
	switch {
	case kanbanEnabled && isCompanion:
		return kanbanBranchCompanion
	case kanbanEnabled && !isCompanion:
		return kanbanBranchLead
	default:
		return kanbanBranchNone
	}
}
