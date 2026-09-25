package kanban

// bootstrap.go holds the vocabulary of a kanban run's multi-session bootstrap:
// the run identifier, the companion roles, and the label that names a companion
// session — the bare role name, with an optional numeric suffix the launcher
// appends on collision.
//
// It lives here rather than in internal/cli because both sides of the bootstrap
// need it and they cannot import each other: the launcher (internal/cli) mints
// the id and recognizes the label, while the SessionStart hook (internal/hook)
// announces them — and internal/cli already imports internal/hook, so the
// dependency can only run in that direction.
//
// Naming policy (card t56, operator decision 2026-08-17; extended to the lead
// by card t133): NO session name carries a run id. Every session — the lead
// and its companions alike — is named by its role alone, and a second live
// session claiming the same role takes the next free number. The premise this
// accepts is one kanban run per machine: the run id was the only distinguisher
// of two concurrent runs, and the operator has decided that case out of scope.
//
// t133 finished what t56 began. The lead kept its run id in its name one card
// longer because the name looked like the only path by which a relaunched lead
// could recover its own id. It is not: the board and the per-session records
// are keyed by the Claude session id (board_store.go, record.go), and companion
// names stopped carrying the run id at t56 — so the id survives only as the
// notice header and the conventional lead-socket path, both display. The
// launcher recovers it from a still-set MOAI_KANBAN_ID instead, and mints a
// fresh one when there is none.

import (
	"strconv"
	"strings"
	"time"
)

// CompanionRoles are the three companion roles of a kanban run, in the order
// the lead announces them: plan / run / sync — the bare phase names (operator
// final design 2026-08-18; full chain lead > plan > run > sync). D1 (card
// t97) retired the review role: the chain is the three phases
// plan -> run -> sync, and reviewing gates integration from the hub instead
// of occupying a companion session.
//
// COMPATIBILITY (naming break, deliberate): an intermediate t118 working-tree
// state renamed these to planner / runner / syncer, but that value set was
// NEVER committed, tagged, or released (`git log --all -S` finds it in no
// commit) — and plan / run / sync are the very values v3.1.0 shipped, so
// this final naming RESTORES released-session compatibility rather than
// breaking it. No alias table maps the transient names: the shape check below
// is the single discriminator, and carrying permanent normalization for a
// name set no binary ever launched would complicate it for nothing.
//
// The lead is deliberately absent from this list. It is the only session that
// carries the kanban token, because that token seeds a session whose
// orchestrator drives the whole plan -> run -> sync chain; giving it
// to a companion would produce three sessions each driving the whole chain.
var CompanionRoles = []string{"plan", "run", "sync"}

// companionLaunchers maps each companion role to the launcher verb the
// bootstrap notice prints for it: "cc" runs the Claude backend, "glm" the GLM
// backend. The mapping IS the recommended default — judgment and review work
// (plan, sync) on Claude, implementation (run) on GLM, the always-on
// lead on GLM — so the notice's copyable launch lines and its per-locale
// recommendation table render one recommendation, not two that can drift.
var companionLaunchers = map[string]string{
	"plan": "cc",
	"run":  "glm",
	"sync": "cc",
}

// CompanionLauncher returns the entry-point verb the bootstrap notice prints
// for a companion role — "cc" or "glm". An unknown role falls back to "cc",
// the pre-t118 shape of the launch lines (every companion on Claude), so a
// display path degrades to the conservative default rather than printing a
// malformed command.
func CompanionLauncher(role string) string {
	if verb, ok := companionLaunchers[role]; ok {
		return verb
	}
	return "cc"
}

// base36Digits is the alphabet of NewRunID.
const base36Digits = "0123456789abcdefghijklmnopqrstuvwxyz"

// NewRunID returns the identifier for one kanban run: the current Unix second
// in lowercase base36, unpadded — six characters at the present epoch.
//
// Monotonic by construction, so a later run sorts after an earlier one, and it
// needs no state file, no counter, and no lock. The session registry is
// deliberately NOT consulted for uniqueness: it has been observed holding a
// dead PID as live, so it cannot answer that question.
//
// Two leads launched within the same second collide. That residual is left
// standing — the window is one second of wall clock on a manual, two-terminal
// operation, and every mechanism that would close it costs more than the case
// is worth.
func NewRunID() string {
	return base36(time.Now().Unix())
}

// base36 renders n in lowercase base36 without padding. Non-positive input
// yields "0" — unreachable for a Unix timestamp, but it keeps the function
// total rather than yielding an empty id on a clock anomaly.
func base36(n int64) string {
	if n <= 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{base36Digits[n%36]}, b...)
		n /= 36
	}
	return string(b)
}

// CompanionLabel returns the label a companion session is launched under: the
// bare role name. This is the form the lead announces and the form an
// unobstructed launch keeps; a collision with a live claim appends a number
// (CompanionNumberLabel), which the launcher — not this composer — resolves.
func CompanionLabel(role string) string {
	return role
}

// CompanionNumberLabel joins a role and a collision number into the bumped
// label a companion launches under when its bare name is held by a live
// session (`plan-1`, `plan-2`, ...). It is the companion sibling of
// FactoryLaneLabel's shape.
func CompanionNumberLabel(role string, n int) string {
	return role + "-" + strconv.Itoa(n)
}

// LeadLabel returns the label a lead session is launched under: the bare role
// name, the same shape companions already use. This is the form the notice
// announces and the form an unobstructed launch keeps; a collision with a live
// claim appends a number (LeadNumberLabel), which the launcher — not this
// composer — resolves.
//
// The run id is deliberately NOT in the name any more. It was carried there so
// a relaunched lead could read its own id back out, but nothing functional
// depends on that continuity: companion names carry no run id under the
// one-machine-one-run policy, and the board and session records are keyed by
// the Claude session id rather than the run id. What remains — the notice
// header and the conventional lead-socket path — is display, and the launcher
// adopts a still-set MOAI_KANBAN_ID rather than reading the name (see
// internal/cli/kanban.go leadRunID).
//
// The label deliberately never satisfies SplitCompanionLabel: RoleLead is
// absent from CompanionRoles, so a session launched under this name is never
// reclassified as a companion by the shape discriminator.
func LeadLabel() string {
	return RoleLead
}

// LeadNumberLabel joins the lead role and a collision number into the bumped
// label a lead launches under when the bare name is held by a live session
// (`lead-1`, `lead-2`, ...). It is the lead sibling of CompanionNumberLabel
// and shares its shape.
func LeadNumberLabel(n int) string {
	return RoleLead + "-" + strconv.Itoa(n)
}

// SplitCompanionLabel splits a companion label into its role and optional
// suffix and reports whether the value has the companion shape at all.
//
// The shape IS the discriminator for companion recognition, so this check is
// load-bearing rather than cosmetic: matching every named session instead would
// silently change launch behavior for unrelated work.
//
// Two forms parse:
//
//   - the bare role (`plan`) — the form the lead announces and the common
//     case under the one-machine-one-run policy;
//   - `<role>-<suffix>` — the suffix is a collision number the launcher
//     appended (`plan-1`), or any run-id-shaped value in that position.
//
// The suffixed form is MIGRATED, not rejected: a rejected suffix would fail
// the shape check, and `-k --name plan-abc123` then reroutes down the
// LEAD branch of the launcher's truth table (a `-k` launch whose name is not
// companion-shaped seeds a whole second chain) — a silent misroute far worse
// than joining under a stale suffix. A role name carries no hyphen, so the
// first hyphen is the boundary and a second hyphen never parses.
func SplitCompanionLabel(label string) (role, suffix string, ok bool) {
	if isCompanionRole(label) {
		return label, "", true
	}
	role, suffix, found := strings.Cut(label, "-")
	if !found || !isCompanionRole(role) || !isRunIDShape(suffix) {
		return "", "", false
	}
	return role, suffix, true
}

// SplitLeadLabel splits a lead label into its optional suffix and reports
// whether the value has the lead shape at all. It is the lead-side counterpart
// of SplitCompanionLabel, and admits the same two forms:
//
//   - the bare role (`lead`) — the form the notice announces and the common
//     case under the one-machine-one-run policy, returning an empty suffix;
//   - `lead-<suffix>` — a collision number the launcher appended (`lead-1`),
//     or a legacy run id an operator is still pasting (`lead-abc123`).
//
// The suffixed form is MIGRATED, not rejected, for the same reason the
// companion side migrates its own: a rejected suffix fails the shape check,
// and `-k --name lead-abc123` then falls through to the branch that treats an
// unrecognized name as no lead name at all — a silent misroute worse than
// joining under a stale suffix. A role name carries no hyphen, so the first
// hyphen is the boundary and a second hyphen never parses.
//
// The suffix is returned for the caller to interpret; it is NOT itself a run
// id. Whether a legacy suffix is adopted as one is the launcher's decision
// (internal/cli/kanban.go leadRunID), which is where a bump number must not be
// mistaken for an id.
func SplitLeadLabel(label string) (suffix string, ok bool) {
	if label == RoleLead {
		return "", true
	}
	role, suffix, found := strings.Cut(label, "-")
	if !found || role != RoleLead || !isRunIDShape(suffix) {
		return "", false
	}
	return suffix, true
}

func isCompanionRole(role string) bool {
	for _, r := range CompanionRoles {
		if r == role {
			return true
		}
	}
	return false
}

// factoryLaneRole is the label prefix of a factory run's numbered workers
// (`worker-<n>`). It is deliberately absent from CompanionRoles: a factory
// worker is not a kanban companion, so the companion shape discriminator must
// not admit the label (and by the same construction a kanban role is never
// mistaken for a worker number).
//
// COMPATIBILITY (label rename with read aliases): factory worker sessions were
// previously labelled `lane-<n>` (numbered form) and `agent-<n>` (the
// role-token join form). Both legacy prefixes stay READABLE — a registry row,
// broker slot, or `--name` value carrying one parses to the same number — but
// every label this package PRODUCES is `worker-<n>`, and the two legacy
// shapes share the one worker numbering (a live `lane-3` or `agent-3` claim
// takes number 3). See SplitFactoryLaneLabel, IsLegacyFactoryLabel, and
// CanonicalFactoryLabel.
const factoryLaneRole = "agent"

// factoryLegacyLaneRole and factoryLegacyAgentRole are the read-only legacy
// prefixes of factory worker labels.
const (
	factoryLegacyLaneRole  = "lane"
	factoryLegacyAgentRole = "worker"
)

// FactoryLaneLabel joins the worker prefix and a number into the label a
// factory worker session is launched under (`worker-3`).
//
// The label deliberately never satisfies SplitCompanionLabel or
// SplitLeadLabel: factory workers neither occupy the three-role kanban chain
// nor the lead position, so a factory name is never reclassified by the
// kanban shape discriminators. Like the companion labels it carries no run id
// — every worker is addressed by name alone (the run's lead dispatches cards
// over cross-session messages), which is why the launcher maintains a
// liveness-checked registry to keep the numbered names unique.
func FactoryLaneLabel(n int) string {
	return factoryLaneRole + "-" + strconv.Itoa(n)
}

// FactoryAgentLabel renders the LEGACY `agent-<n>` label. Nothing launches
// under it any more; it exists so the launcher can route a legacy `-f agent`
// join through the one deprecation-hint site (the claim canonicalizes it to
// `worker-<n>`).
func FactoryAgentLabel(n int) string {
	return FactoryLaneLabel(n)
}

// splitFactoryPrefixedLabel parses `<prefix>-<n>` (n >= 1) and reports the
// prefix it carried. A suffix that does not parse as such a number (`worker-`,
// `worker-a`, `worker-3-extra`, `worker--3`) reads as "not a worker", never
// as an error.
func splitFactoryPrefixedLabel(label string) (prefix string, n int, ok bool) {
	role, suffix, found := strings.Cut(label, "-")
	if !found {
		return "", 0, false
	}
	n, err := strconv.Atoi(suffix)
	if err != nil || n < 1 {
		return "", 0, false
	}
	return role, n, true
}

// SplitFactoryAgentLabel splits a legacy `agent-<n>` label into its number
// and reports whether the value has that shape at all.
func SplitFactoryAgentLabel(label string) (n int, ok bool) {
	prefix, n, ok := splitFactoryPrefixedLabel(label)
	if !ok || prefix != factoryLaneRole {
		return 0, false
	}
	return n, true
}

// SplitFactoryLaneLabel splits a factory worker label into its number and
// reports whether the value has the worker shape at all: the canonical
// `worker-<n>` or the legacy `lane-<n>`. The shape is the discriminator for
// factory worker recognition, exactly as SplitCompanionLabel is for kanban
// companions. The legacy `agent-<n>` shape is recognised separately by
// SplitFactoryAgentLabel (and by CanonicalFactoryLabel, which accepts all
// three).
func SplitFactoryLaneLabel(label string) (n int, ok bool) {
	prefix, n, ok := splitFactoryPrefixedLabel(label)
	if !ok || (prefix != factoryLegacyAgentRole && prefix != factoryLegacyLaneRole) {
		return 0, false
	}
	return n, true
}

// factoryLabelNumber parses any factory worker label shape — canonical
// `worker-<n>` or legacy `lane-<n>` / `agent-<n>` — into its number.
func factoryLabelNumber(label string) (n int, ok bool) {
	if n, ok := SplitFactoryAgentLabel(label); ok {
		return n, true
	}
	return SplitFactoryLaneLabel(label)
}

// IsLegacyFactoryLabel reports whether label is one of the pre-rename worker
// shapes (`lane-<n>`, `agent-<n>`) — the launcher's deprecation-hint trigger.
func IsLegacyFactoryLabel(label string) bool {
	prefix, _, ok := splitFactoryPrefixedLabel(label)
	return ok && (prefix == factoryLegacyLaneRole || prefix == factoryLegacyAgentRole)
}

// CanonicalFactoryLabel maps any factory worker label shape onto its
// canonical `worker-<n>` form; ok is false for anything that is not a worker
// label.
func CanonicalFactoryLabel(label string) (string, bool) {
	n, ok := factoryLabelNumber(label)
	if !ok {
		return "", false
	}
	return FactoryLaneLabel(n), true
}

// NextFactoryWorkerNumber returns the number a `-f worker` join takes: one
// past the highest LIVE claim in the pruned registry, across the canonical
// and both legacy shapes (1 when nothing is claimed). Dead claims are pruned
// first so a crashed worker frees its number for reuse.
func NextFactoryWorkerNumber(reg map[string]FactoryWorkerEntry, alive func(int) bool) int {
	reg = PruneFactoryDeadClaims(reg, alive)
	highest := 0
	for label := range reg {
		if n, ok := factoryLabelNumber(label); ok && n > highest {
			highest = n
		}
	}
	return highest + 1
}

// isRunIDShape reports whether s is one or more lowercase alphanumerics — the
// shape NewRunID produces, and the shape a companion label's suffix admits
// (a collision number, or a migrated legacy run id).
func isRunIDShape(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// The leader-socket roots (t118 socket scheme, v3.1.1). The two modes keep
// separate directories — `/tmp/moai-socket-kanban/<run-id>` and
// `/tmp/moai-socket-factory/<run-id>` — replacing the pre-t118 flat
// single-directory shape, which named one mode and left the other squatting
// in it. The value is a conventional address line the SessionStart notice
// prints (EnvMoaiKanbanLeadAddr), not a filesystem contract the messaging
// substrate is bound to — the actual transport is runtime-owned — which is
// why the /tmp literal is acceptable here rather than os.TempDir().
const (
	kanbanSocketDir  = "/tmp/moai-socket-kanban"
	factorySocketDir = "/tmp/moai-socket-factory"
)

// LeaderSocketPath returns the conventional leader-socket address a KANBAN
// lead publishes for runID: <kanbanSocketDir>/<run-id>. The launcher sets it
// into EnvMoaiKanbanLeadAddr and the SessionStart notice prints it verbatim;
// the factory lead uses FactoryLeaderSocketPath, its factory sibling.
func LeaderSocketPath(runID string) string {
	return kanbanSocketDir + "/" + runID
}

// FactoryLeaderSocketPath returns the conventional leader-socket address a
// FACTORY lead publishes for runID: <factorySocketDir>/<run-id>. Same carrier
// and same print surface as LeaderSocketPath; a separate directory so a kanban
// run and a factory run sharing one machine never address the same line.
func FactoryLeaderSocketPath(runID string) string {
	return factorySocketDir + "/" + runID
}
