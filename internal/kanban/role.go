// role.go — the role-declaration carrier (SPEC-KANBAN-BOARD-001 REQ-KB-025,
// M1).
//
// The CONTRACT this carrier satisfies is SPEC-KANBAN-BOOTSTRAP-001 REQ-KS-006,
// consumed WHOLE and by reference — read that requirement for what a
// declaration is. Nothing here restates or enumerates its properties: a
// partial enumeration is how the borrow forks (spec.md §A.8). What this file
// defines is the CARRIER alone — where the declaration lives — because
// REQ-KS-006 deliberately fixes no carrier and the sibling defining the
// contract has not landed (REQ-KB-025's establish branch). Where that sibling
// later lands, it adopts this carrier and defines no second one.
//
// The carrier is anchored at the board root — the same single origin the
// board state uses — so any session, in any checkout, resolves any session's
// declaration. This file defines no part of role election, dispatch routing,
// or quorum accounting (REQ-KS-004 / REQ-KS-019 / REQ-KS-012, the bootstrap
// sibling's); it carries the key those mechanisms read.
package kanban

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RoleLead names the one role the board's write guard admits (REQ-KB-017).
// The role SET and the election that occupies it are the bootstrap sibling's
// (REQ-KS-004) — this constant names a role value, not an election rule, and
// a declaration may carry any role string the topology defines. The value is
// the operator final design's session noun (2026-08-18): the lead session is
// named `lead-<run-id>` — the same value v3.1.0 shipped — and
// LeadLabel/SplitLeadLabel compose and parse that shape from this constant.
const RoleLead = "lead"

// RoleLane names a factory run's numbered lane (the `lane-<n>` lanes of
// `moai cc -f <N>`, SPEC-FACTORY-WORKER-FANOUT-001). It is deliberately NOT a
// CompanionRoles member: factory lanes are dispatched cards by their lead
// over cross-session messages and never occupy the three-role kanban chain, so
// the companion shape discriminators must not admit their labels.
const RoleLane = "lane"

// RoleDeclaration is the persisted declaration artifact: the role a session
// occupies, recorded by that session at launch.
//
// Label is the launch label recorded ALONGSIDE the role as a separate datum
// — a role does not determine a label and a label does not determine a role
// (one role corresponds to two-or-more possible labels chosen by the operator
// at launch). No code path computes Role from Label or the reverse; the field
// exists so the two data stay observably distinct in the artifact.
type RoleDeclaration struct {
	SessionID  string `json:"session_id"`
	Role       string `json:"role"`
	Label      string `json:"label,omitempty"`
	DeclaredAt string `json:"declared_at"`
}

// roleDeclarationDir is the carrier's home beneath the board directory. It is
// a sibling of the state file, not part of it, so the board's bounded
// recovery (which modifies the state file alone) never touches declarations.
func roleDeclarationDir(root string) string {
	return filepath.Join(BoardDir(root), "roles")
}

// DeclareRole records that sessionID occupies role, launched under label.
//
// The declaration is written by the declaring session itself and read by
// every other session through the shared origin — both cross-session
// directions (a worker resolving the lead, the lead resolving a worker) are
// the same read against the same path. Re-declaring overwrites, which is how
// a relaunched session refreshes its declaration.
func DeclareRole(root, sessionID, role, label string) error {
	if err := validateSessionID(sessionID); err != nil {
		return fmt.Errorf("declare role: %w", err)
	}
	if role == "" {
		return fmt.Errorf("declare role: role is empty")
	}

	dir := roleDeclarationDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("declare role: creating declaration dir: %w", err)
	}

	rec := RoleDeclaration{
		SessionID:  sessionID,
		Role:       role,
		Label:      label,
		DeclaredAt: time.Now().UTC().Format(time.RFC3339),
	}
	encoded, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("declare role: encoding: %w", err)
	}
	path := filepath.Join(dir, sessionID+".json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return fmt.Errorf("declare role: %w", err)
	}
	return nil
}

// @MX:ANCHOR: [AUTO] ResolveDeclaredRole — the single-origin role-declaration read (REQ-KB-025 carrier)
// @MX:REASON: expected fan_in >= 3 (requireLeadRole for board writes and recovery, cross-session consumers in the worktree sibling's gates, bootstrap's dispatch/quorum key); a parallel declaration datum is the fork AP-31 names
//
// ResolveDeclaredRole returns the role sessionID declared. It is callable by
// any session — the lead resolving a worker's declaration (the dispatch and
// quorum key) and a worker resolving the lead's (the sibling's gates) are the
// same operation against the same single-origin artifact.
//
// An absent or unreadable declaration is an error, never a default role: the
// board's write guard treats an unreadable role as a refusal, because a
// refusal with nothing to read does not refuse — it admits every write.
func ResolveDeclaredRole(root, sessionID string) (string, error) {
	if err := validateSessionID(sessionID); err != nil {
		return "", fmt.Errorf("resolve declared role: %w", err)
	}

	raw, err := os.ReadFile(filepath.Join(roleDeclarationDir(root), sessionID+".json"))
	if err != nil {
		return "", fmt.Errorf("resolve declared role: %w", err)
	}

	var rec RoleDeclaration
	if err := json.Unmarshal(raw, &rec); err != nil {
		return "", fmt.Errorf("resolve declared role: decoding: %w", err)
	}
	if rec.Role == "" {
		return "", fmt.Errorf("resolve declared role: declaration for %q carries no role", sessionID)
	}
	return rec.Role, nil
}
