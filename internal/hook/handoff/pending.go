package handoff

// pending.go implements the reverse-handoff half (SPEC-HANDOFF-AUTORESUME-001):
// the `moai handoff save` / `clear` CLI writer and the SessionStart reader operate
// on a SEPARATE tree — `<projectDir>/.moai/state/handoff/pending.json` (JSON) —
// that is fully decoupled from the SessionEnd → memory flow in persist.go
// (`session-handoff/pending.md`, Markdown). The two flows never read or write
// each other's paths, so the SessionEnd-vs-SessionStart consume race is
// statically unreachable (design.md §B).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PendingSchemaVersion is the current schema version of pending.json.
const PendingSchemaVersion = 1

// Directives carries the mode-change directive metadata recorded at save time.
// These are NEVER activated on injection — the SessionStart handler renders them
// as manual-paste restoration guidance only (verification-claim-integrity §1.1).
type Directives struct {
	Ultrathink bool   `json:"ultrathink"`
	Ultracode  bool   `json:"ultracode"`
	Goal       string `json:"goal"`
}

// PendingRecord is the JSON schema of handoff/pending.json (design.md §B.1).
// REQ-AUTORESUME-006 mandates at least schema_version, body, directives,
// conversation_language, and saved_at.
type PendingRecord struct {
	SchemaVersion        int        `json:"schema_version"`
	SpecID               string     `json:"spec_id,omitempty"`
	Phase                string     `json:"phase,omitempty"`
	SavedAt              time.Time  `json:"saved_at"`
	SavedBySession       string     `json:"saved_by_session,omitempty"`
	ConversationLanguage string     `json:"conversation_language,omitempty"`
	Directives           Directives `json:"directives"`
	// Body is the verbatim paste-ready resume (6-block, cut-line markers included).
	Body string `json:"body"`
}

// handoffStateDir returns <projectDir>/.moai/state/handoff — the reverse-handoff
// tree, kept strictly separate from session-handoff/ (persist.go).
func handoffStateDir(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", "handoff")
}

// PendingPath returns <projectDir>/.moai/state/handoff/pending.json.
func PendingPath(projectDir string) string {
	return filepath.Join(handoffStateDir(projectDir), "pending.json")
}

// ConsumedDir returns <projectDir>/.moai/state/handoff/consumed — the audit-trail
// directory the SessionStart handler renames a claimed pending.json into.
func ConsumedDir(projectDir string) string {
	return filepath.Join(handoffStateDir(projectDir), "consumed")
}

// SavePending writes rec to handoff/pending.json using an atomic temp-file +
// rename (reusing the persist.go atomicWriteFile contract). REQ-AUTORESUME-005:
// it writes ONLY the handoff/ tree and NEVER touches session-handoff/pending.md.
// The record is written 0o600 because the resume body may carry session context.
func SavePending(projectDir string, rec *PendingRecord) error {
	if rec == nil {
		return fmt.Errorf("save pending: nil record")
	}
	if rec.SchemaVersion == 0 {
		rec.SchemaVersion = PendingSchemaVersion
	}
	if rec.SavedAt.IsZero() {
		rec.SavedAt = time.Now()
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pending record: %w", err)
	}

	dir := handoffStateDir(projectDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create handoff state dir: %w", err)
	}
	if err := atomicWriteFile(dir, "pending.json", data, 0o600); err != nil {
		return fmt.Errorf("atomic write pending.json: %w", err)
	}
	return nil
}

// ClearPending removes handoff/pending.json. An absent file is a no-op (nil).
// REQ-AUTORESUME-007: it touches ONLY the handoff/ tree, never
// session-handoff/pending.md.
func ClearPending(projectDir string) error {
	if err := os.Remove(PendingPath(projectDir)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove pending.json: %w", err)
	}
	return nil
}

// ReadPending reads and parses handoff/pending.json.
//
//   - absent file        → (nil, false, nil)   — the caller treats this as a no-op.
//   - present + valid     → (&rec, true, nil)
//   - present + corrupt    → (nil, true, err)   — the caller preserves the file and
//     logs a best-effort warning (never blocks the session).
//
// The `present` bool lets the SessionStart handler distinguish "nothing to do"
// (absent, silent) from "corrupt pending" (present, slog.Warn + preserve) per
// REQ-AUTORESUME-017.
func ReadPending(projectDir string) (*PendingRecord, bool, error) {
	data, err := os.ReadFile(PendingPath(projectDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read pending.json: %w", err)
	}
	var rec PendingRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, true, fmt.Errorf("parse pending.json: %w", err)
	}
	return &rec, true, nil
}
