package factory

// bootstrap.go holds the vocabulary of a factory run's multi-session bootstrap:
// the run identifier and the label that names a leader session — the bare role
// name, with an optional numeric suffix the launcher appends on collision.
//
// It lives here rather than in internal/cli because both sides of the bootstrap
// need it and they cannot import each other: the launcher (internal/cli) mints
// the id and recognizes the label, while the SessionStart hook (internal/hook)
// announces them — and internal/cli already imports internal/hook, so the
// dependency can only run in that direction.
//
// Naming policy (card t56, operator decision 2026-08-17; extended to the leader
// by card t133): NO session name carries a run id. Every session is named by its
// role alone, and a second live session claiming the same role takes the next
// free number. The premise this accepts is one run per machine: the run id was
// the only distinguisher of two concurrent runs, and the operator has decided
// that case out of scope.
//
// t133 finished what t56 began. The leader kept its run id in its name one card
// longer because the name looked like the only path by which a relaunched leader
// could recover its own id. It is not: the per-session records are keyed by the
// Claude session id (record.go), so the id survives only as the notice header
// and the conventional leader-socket path, both display. The launcher recovers
// it from a still-set run-id marker instead, and mints a fresh one when there
// is none.

import (
	"strconv"
	"strings"
	"time"
)

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
// Two leaders launched within the same second collide. That residual is left
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

// LeaderLabel returns the label a leader session is launched under: the bare role
// name. This is the form the notice announces and the form an unobstructed launch keeps; a collision with a live
// claim appends a number (LeaderNumberLabel), which the launcher — not this
// composer — resolves.
//
// The run id is deliberately NOT in the name any more. It was carried there so
// a relaunched leader could read its own id back out, but nothing functional
// depends on that continuity: no session name carries a run id under the
// one-machine-one-run policy, and the session records are keyed by the Claude
// session id rather than the run id. What remains — the notice
// header and the conventional leader-socket path — is display, and the launcher
// adopts a still-set MOAI_KANBAN_ID rather than reading the name (see
// internal/cli/factory_launch_helpers.go leaderRunID).
func LeaderLabel() string {
	return RoleLeader
}

// LeaderNumberLabel joins the leader role and a collision number into the
// bumped label a leader launches under when the bare name is held by a live
// session (`leader-1`, `leader-2`, ...).
func LeaderNumberLabel(n int) string {
	return RoleLeader + "-" + strconv.Itoa(n)
}

// SplitLeaderLabel splits a leader label into its optional suffix and reports
// whether the value has the leader shape at all. Two forms parse:
//
//   - the bare role (`leader`) — the form the notice announces and the common
//     case under the one-machine-one-run policy, returning an empty suffix;
//   - `leader-<suffix>` — a collision number the launcher appended
//     (`leader-2`), or a run id an operator is still pasting
//     (`leader-abc123`).
//
// The legacy `lead` / `lead-<suffix>` spellings do NOT parse (they are
// detection values, IsLegacyLeaderSpelling) — a legacy label maps to no role
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
//
// The suffix is returned for the caller to interpret; it is NOT itself a run
// id. Whether a suffix is adopted as one is the launcher's decision
// (internal/cli/factory_launch_helpers.go leaderRunID), which is where a bump number must not be
// mistaken for an id.
func SplitLeaderLabel(label string) (suffix string, ok bool) {
	if label == RoleLeader {
		return "", true
	}
	role, suffix, found := strings.Cut(label, "-")
	if !found || role != RoleLeader || !isRunIDShape(suffix) {
		return "", false
	}
	return suffix, true
}

// factoryLaneRole is the label prefix of a factory run's numbered lanes
// (`lane-<n>`, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-004/-010).
//
// COMPATIBILITY (label rename, rejection not aliasing): factory lane sessions
// were previously labelled `worker-<n>` and, before that, `agent-<n>`. The
// legacy spellings are NOT readable aliases any more — a legacy label maps to
// no role and holds no lane number (SplitFactoryLaneLabel admits `lane-<n>`
// only). They survive solely as DETECTION values for the stale-record rule
// (SplitFactoryLegacyLabel, IsLegacyFactoryLabel, IsLegacyFactoryRoleValue):
// a live legacy record of the same run refuses the join (REQ-RNC-022), a dead
// one ages out through the existing stale paths, and nothing on disk is
// rewritten.
const factoryLaneRole = "lane"

// factoryLegacyWorkerRole and factoryLegacyAgentRole are the legacy lane
// label prefixes (`worker-<n>`, `agent-<n>`). They exist only to be detected
// and refused or reported stale — never written, never mapped to a role
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
const (
	factoryLegacyWorkerRole = "worker"
	factoryLegacyAgentRole  = "agent"
)

// FactoryLaneLabel joins the lane prefix and a number into the label a
// factory lane session is launched under (`lane-3`).
//
// The label deliberately never satisfies SplitLeaderLabel: a factory lane does
// not occupy the leader position, so a lane name is never reclassified by the
// leader shape discriminator. Like the leader label it carries no run id
// — every lane is addressed by name alone (the run's leader dispatches cards
// over cross-session messages), which is why the launcher maintains a
// liveness-checked registry to keep the numbered names unique.
func FactoryLaneLabel(n int) string {
	return factoryLaneRole + "-" + strconv.Itoa(n)
}

// FactoryLegacyAgentLabel renders the LEGACY `agent-<n>` label shape. Nothing
// launches under it and no reader maps it to a lane; it exists for detection
// and test fixtures only (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
func FactoryLegacyAgentLabel(n int) string {
	return factoryLegacyAgentRole + "-" + strconv.Itoa(n)
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

// SplitFactoryLegacyAgentLabel splits a legacy `agent-<n>` label into its number
// and reports whether the value has that shape at all. Detection only: a
// match never maps the label to a lane (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-009).
func SplitFactoryLegacyAgentLabel(label string) (n int, ok bool) {
	prefix, n, ok := splitFactoryPrefixedLabel(label)
	if !ok || prefix != factoryLegacyAgentRole {
		return 0, false
	}
	return n, true
}

// SplitFactoryLaneLabel splits a factory lane label into its number and
// reports whether the value has the canonical `lane-<n>` shape at all. It is
// the discriminator for factory lane recognition, exactly as
// SplitLeaderLabel is for the leader. The legacy `worker-<n>` /
// `agent-<n>` shapes are DETECTION values only (SplitFactoryLegacyLabel,
// IsLegacyFactoryLabel) — they never parse as a lane.
func SplitFactoryLaneLabel(label string) (n int, ok bool) {
	prefix, n, ok := splitFactoryPrefixedLabel(label)
	if !ok || prefix != factoryLaneRole {
		return 0, false
	}
	return n, true
}

// SplitFactoryLegacyLabel splits a legacy factory label (`worker-<n>` /
// `agent-<n>`) into its number and reports whether the value has a legacy
// shape at all. Detection only: the number is for refusal/stale messaging,
// never a lane number (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
func SplitFactoryLegacyLabel(label string) (n int, ok bool) {
	prefix, n, ok := splitFactoryPrefixedLabel(label)
	if !ok || (prefix != factoryLegacyWorkerRole && prefix != factoryLegacyAgentRole) {
		return 0, false
	}
	return n, true
}

// factoryLabelNumber parses a CANONICAL factory lane label (`lane-<n>`) into
// its number. Legacy shapes are deliberately not accepted — a legacy claim
// holds no number in the new regime; the same-run refusal replaced the shared
// number space (SPEC-ROLE-NAMING-CODE-001 design §4).
func factoryLabelNumber(label string) (n int, ok bool) {
	return SplitFactoryLaneLabel(label)
}

// IsLegacyFactoryLabel reports whether label is one of the pre-rename lane
// shapes (`worker-<n>`, `agent-<n>`) — the stale-record detection trigger
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-022).
func IsLegacyFactoryLabel(label string) bool {
	prefix, _, ok := splitFactoryPrefixedLabel(label)
	return ok && (prefix == factoryLegacyWorkerRole || prefix == factoryLegacyAgentRole)
}

// IsLegacyFactoryRoleValue reports whether value is ANY legacy factory role
// spelling — the bare role tokens (`worker`, `agent`) as well as their
// numbered labels. Detection only: a true result triggers a refuse-or-notice
// path, never a role mapping (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
func IsLegacyFactoryRoleValue(value string) bool {
	if value == factoryLegacyWorkerRole || value == factoryLegacyAgentRole {
		return true
	}
	return IsLegacyFactoryLabel(value)
}

// CanonicalFactoryLabel maps a canonical factory lane label onto its
// canonical `lane-<n>` form; ok is false for anything that is not a lane
// label.
func CanonicalFactoryLabel(label string) (string, bool) {
	n, ok := factoryLabelNumber(label)
	if !ok {
		return "", false
	}
	return FactoryLaneLabel(n), true
}

// NextFactoryLaneNumber returns the number a `-l` lane join takes: one
// past the highest LIVE canonical claim in the pruned registry (1 when
// nothing is claimed). Legacy claims hold no number — a live legacy record
// refuses the join instead (REQ-RNC-022); dead claims are pruned first so a
// crashed lane frees its number for reuse.
func NextFactoryLaneNumber(reg map[string]FactoryLaneEntry, alive func(int) bool) int {
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
// shape NewRunID produces, and the shape a leader label's suffix admits
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

// The factory leader-socket root (t118 socket scheme, v3.1.1:
// `/tmp/moai-socket-factory/<run-id>`). The value is a conventional address
// line the SessionStart notice prints (EnvFactoryLeadAddr), not a filesystem
// contract the messaging substrate is bound to — the actual transport is
// runtime-owned — which is why the /tmp literal is acceptable here rather than
// os.TempDir().
const factorySocketDir = "/tmp/moai-socket-factory"

// FactoryLeaderSocketPath returns the conventional leader-socket address a
// FACTORY leader publishes for runID: <factorySocketDir>/<run-id>. The value is
// printed by the SessionStart notice (EnvFactoryLeadAddr carries it).
func FactoryLeaderSocketPath(runID string) string {
	return factorySocketDir + "/" + runID
}
