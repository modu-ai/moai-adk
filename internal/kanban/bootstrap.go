package kanban

// bootstrap.go holds the vocabulary of a kanban run's multi-session bootstrap:
// the run identifier, the companion roles, and the `<role>-<run-id>` label that
// joins the two.
//
// It lives here rather than in internal/cli because both sides of the bootstrap
// need it and they cannot import each other: the launcher (internal/cli) mints
// the id and recognizes the label, while the SessionStart hook (internal/hook)
// announces them — and internal/cli already imports internal/hook, so the
// dependency can only run in that direction.

import (
	"strconv"
	"strings"
	"time"
)

// CompanionRoles are the four companion roles of a kanban run, in the order
// the lead announces them.
//
// The lead is deliberately absent from this list. It is the only session that
// carries the kanban token, because that token seeds a session whose
// orchestrator drives the whole plan -> run -> verify -> sync chain; giving it
// to a companion would produce four sessions each driving the whole chain.
var CompanionRoles = []string{"plan", "run", "review", "sync"}

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

// CompanionLabel joins a role and a run id into the label a companion session
// is launched under.
func CompanionLabel(role, runID string) string {
	return role + "-" + runID
}

// LeadLabel joins the lead role and a run id into the label a lead session is
// launched under, in the same `<role>-<run-id>` form CompanionLabel produces.
//
// The label deliberately never satisfies SplitCompanionLabel: RoleLead is
// absent from CompanionRoles, so a session launched under this name is never
// reclassified as a companion by the shape discriminator.
func LeadLabel(runID string) string {
	return RoleLead + "-" + runID
}

// SplitCompanionLabel splits a `<role>-<run-id>` label into its parts and
// reports whether the value has the companion shape at all.
//
// The shape IS the discriminator for companion recognition, so this check is
// load-bearing rather than cosmetic: matching every named session instead would
// silently change launch behavior for unrelated work. A role name carries no
// hyphen, so the first hyphen is the boundary.
func SplitCompanionLabel(label string) (role, runID string, ok bool) {
	role, runID, found := strings.Cut(label, "-")
	if !found || !isCompanionRole(role) || !isRunIDShape(runID) {
		return "", "", false
	}
	return role, runID, true
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
// SplitLeadLabel: factory workers neither occupy the four-role kanban chain
// nor the lead position, so a factory name is never reclassified by the
// kanban shape discriminators. Unlike the kanban labels it carries no run id
// — a factory worker is addressed by name alone (the run's lead dispatches
// cards over cross-session messages), which is why the launcher maintains a
// separate liveness-checked registry to keep the numbered names unique.
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
// shape NewRunID produces.
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
