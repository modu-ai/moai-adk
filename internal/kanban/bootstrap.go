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
// Naming policy (card t56, operator decision 2026-08-17): companion names
// carry no run id. The run id remains the LEAD's identifier (its session name,
// its leader socket path, its MOAI_KANBAN_ID); a companion is named by its
// role alone, and a second live session claiming the same role takes the next
// free number. The premise this accepts is one kanban run per machine — the
// run id was the only distinguisher of two concurrent runs, and the operator
// has decided that case out of scope.

import (
	"strconv"
	"strings"
	"time"
)

// CompanionRoles are the three companion roles of a kanban run, in the order
// the lead announces them. D1 (card t97) retired the review role: the chain
// is the three phases plan -> run -> sync, and reviewing gates integration
// from the hub instead of occupying a companion session.
//
// The lead is deliberately absent from this list. It is the only session that
// carries the kanban token, because that token seeds a session whose
// orchestrator drives the whole plan -> run -> sync chain; giving it
// to a companion would produce three sessions each driving the whole chain.
var CompanionRoles = []string{"plan", "run", "sync"}

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
// FactoryWorkerLabel's shape.
func CompanionNumberLabel(role string, n int) string {
	return role + "-" + strconv.Itoa(n)
}

// LeadLabel joins the lead role and a run id into the label a lead session is
// launched under. The lead is the one session that keeps its run id in its
// name — it survives /clear and anchors the leader socket — while companions
// are named by their bare roles.
//
// The label deliberately never satisfies SplitCompanionLabel: RoleLead is
// absent from CompanionRoles, so a session launched under this name is never
// reclassified as a companion by the shape discriminator.
func LeadLabel(runID string) string {
	return RoleLead + "-" + runID
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
//   - the bare role (`plan`) — the form the lead announces and the common case
//     under the one-machine-one-run policy;
//   - `<role>-<suffix>` — the suffix is a collision number the launcher
//     appended (`plan-1`), or a run id from a pre-policy launch (`plan-abc123`).
//
// The legacy run-id form is MIGRATED, not rejected: a rejected suffix would
// fail the shape check, and `-k --name plan-abc123` then reroutes down the
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

// SplitLeadLabel splits a `lead-<run-id>` label into its run id and reports
// whether the value has the lead shape at all. It is the lead-side counterpart
// of SplitCompanionLabel, and admits exactly the same run-id shape.
//
// It exists so the launcher can ADOPT the run id an operator embedded in a name
// it was handed, rather than minting a second one beside it. A lead has two
// id-bearing surfaces — the session name and MOAI_KANBAN_ID — and without this
// splitter nothing joins them: the SessionStart notice composes its companion
// launch commands from the environment id, so a session named `lead-X` can
// print commands belonging to run Y, and anyone who copies one opens an orphan.
// The companion branch already derives its id from its label and states that it
// does so precisely so the two can never disagree; this restores the symmetry.
func SplitLeadLabel(label string) (runID string, ok bool) {
	role, runID, found := strings.Cut(label, "-")
	if !found || role != RoleLead || !isRunIDShape(runID) {
		return "", false
	}
	return runID, true
}

func isCompanionRole(role string) bool {
	for _, r := range CompanionRoles {
		if r == role {
			return true
		}
	}
	return false
}

// factoryWorkerRole is the label prefix of a factory run's numbered workers.
// It is deliberately absent from CompanionRoles: a factory worker is not a
// kanban companion, so the companion shape discriminator must not admit the
// `worker-<n>` label (and by the same construction a kanban role is never
// mistaken for a worker number).
const factoryWorkerRole = "worker"

// FactoryWorkerLabel joins the worker prefix and a worker number into the
// label a factory worker session is launched under (`worker-3`).
//
// The label deliberately never satisfies SplitCompanionLabel or
// SplitLeadLabel: factory workers neither occupy the three-role kanban chain
// nor the lead position, so a factory name is never reclassified by the
// kanban shape discriminators. Like the companion labels it carries no run id
// — every lane is addressed by name alone (the run's lead dispatches cards
// over cross-session messages), which is why the launcher maintains a
// liveness-checked registry to keep the numbered names unique.
func FactoryWorkerLabel(n int) string {
	return factoryWorkerRole + "-" + strconv.Itoa(n)
}

// SplitFactoryWorkerLabel splits a `worker-<n>` label into its number and
// reports whether the value has the worker shape at all. The shape is the
// discriminator for factory worker recognition, exactly as
// SplitCompanionLabel is for kanban companions: a number of 1 or more
// following the `worker-` prefix, nothing else. A suffix that does not parse
// as such a number (`worker-`, `worker-a`, `worker-3-extra`, `worker--3`)
// reads as "not a worker", never as an error.
func SplitFactoryWorkerLabel(label string) (n int, ok bool) {
	role, suffix, found := strings.Cut(label, "-")
	if !found || role != factoryWorkerRole {
		return 0, false
	}
	n, err := strconv.Atoi(suffix)
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
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
