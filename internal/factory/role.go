// role.go — the role vocabulary the factory and the session records share:
// RoleLeader, RoleLane, and the detection of the retired leader spelling.
// (The role-declaration carrier that used to live here went with the kanban
// board: SPEC-LAUNCHER-ENTRY-FLAGS-001 M6. Declaration files a past board wrote
// under .moai/state/kanban-board/roles are left on disk, unread.)
package factory

import "strings"

// RoleLeader names the leader role. The role SET and the election that
// occupies it are the bootstrap sibling's (REQ-KS-004) — this constant names a
// role value, not an election rule. The value is the operator's leader noun
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-006/-010): the managing session is
// `leader`, its bumped and run-id forms `leader-<n>` / `leader-<run-id>` — the
// same value the label composer (LeaderLabel), the collision bumper
// (LeaderNumberLabel), the shape parser (SplitLeaderLabel), and the session
// records all use. The pre-rename value `lead` survives only as a detection value
// (IsLegacyLeaderSpelling): readers refuse it, writers never write it, and no
// existing record is rewritten (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
const RoleLeader = "leader"

// legacyLeaderSpelling is the pre-rename leader value. It exists only to be
// detected (stale-run notices, registry notices) — never written, never
// accepted as a role (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
const legacyLeaderSpelling = "lead"

// IsLegacyLeaderSpelling reports whether label is a pre-rename leader spelling —
// the bare `lead` or any `lead-<suffix>` form. Detection only: a true result
// never maps the value to the leader role, it triggers a refuse-or-notice
// path (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009, REQ-RNC-025).
func IsLegacyLeaderSpelling(label string) bool {
	return label == legacyLeaderSpelling || strings.HasPrefix(label, legacyLeaderSpelling+"-")
}

// RoleLane names a factory run's numbered lane session (the `lane-<n>` labels,
// SPEC-ROLE-NAMING-CODE-001 REQ-RNC-004). The persisted value stays "lane":
// it is the record's role key on disk, not user-facing notation, so renaming
// it would change the record format existing readers parse.
const RoleLane = "lane"
