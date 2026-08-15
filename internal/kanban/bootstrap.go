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

func isCompanionRole(role string) bool {
	for _, r := range CompanionRoles {
		if role == r {
			return true
		}
	}
	return false
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
