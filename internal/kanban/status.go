// status.go — the status vocabulary the board READS
// (SPEC-KANBAN-BOARD-001 REQ-KB-007/024, M2).
//
// The board never writes a status and never extends the enum; it reads the
// canonical 8-value set that .claude/rules/moai/development/
// spec-frontmatter-schema.md § Status Enum owns. The Go-side constants below
// mirror that schema so no string literal is invented — the schema is the
// single source of truth, and a value added there requires a mirror update
// here, not a parallel vocabulary.
package kanban

// Status values — the canonical 8-value enum, mirrored from
// spec-frontmatter-schema.md § Status Enum (the SSOT; REQ-KB-007 forbids
// extending it and the board writes none of them).
const (
	StatusDraft       = "draft"
	StatusPlanned     = "planned"
	StatusInProgress  = "in-progress"
	StatusImplemented = "implemented"
	StatusCompleted   = "completed"
	StatusSuperseded  = "superseded"
	StatusArchived    = "archived"
	StatusRejected    = "rejected"
)

// StatusUnresolved is NOT a member of the canonical enum (REQ-KB-024): it is
// the board's report that a card's status source could not be resolved to
// exactly one tree — a worktree existing but reporting no branch, or a
// single-branch search refused on more than one match. No member of the enum
// is ever substituted for it.
const StatusUnresolved = "unresolved"

// canonicalStatuses is the closed 8-value set, for membership checks.
var canonicalStatuses = map[string]bool{
	StatusDraft:       true,
	StatusPlanned:     true,
	StatusInProgress:  true,
	StatusImplemented: true,
	StatusCompleted:   true,
	StatusSuperseded:  true,
	StatusArchived:    true,
	StatusRejected:    true,
}

// IsCanonicalStatus reports whether s is a member of the canonical 8-value
// enum. StatusUnresolved deliberately is not.
func IsCanonicalStatus(s string) bool {
	return canonicalStatuses[s]
}
