package handoff

// pending.go implements the reverse-handoff half (SPEC-HANDOFF-AUTORESUME-001).
// Resume rows and SessionEnd memory rows share factory.db but remain distinct
// tables and state machines, so neither flow consumes the other's records.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// PendingSchemaVersion is the current resume handoff payload version.
const PendingSchemaVersion = 1

// Directives carries the mode-change directive metadata recorded at save time.
// These are NEVER activated on injection — the SessionStart handler renders them
// as manual-paste restoration guidance only (verification-claim-integrity §1.1).
type Directives struct {
	Ultrathink bool   `json:"ultrathink"`
	Ultracode  bool   `json:"ultracode"`
	Goal       string `json:"goal"`
}

// EmbeddedGoal carries a live armed goal's verbatim condition + arm-time Ceiling
// so the SessionStart /clear handler can re-arm it under the NEW session-id
// (SPEC-INFINITE-GOAL-001 REQ-6, Option A session-id keying). The Ceiling fields
// mirror goal.Ceiling without importing internal/goal (keeps the handoff package
// decoupled). nil when no goal was armed at save time → no rearm on /clear.
type EmbeddedGoal struct {
	Condition   string `json:"condition"`
	MaxTurns    int    `json:"max_turns"`
	MaxDuration int    `json:"max_duration,omitempty"`
	CostCap     int    `json:"cost_cap,omitempty"`
}

// IsUnbounded reports whether the embedded goal is an infinite arm (MaxTurns==0)
// with NO real bound (neither MaxDuration nor CostCap). Used by the D8 defense-
// in-depth re-validation: an unbounded embedded record is rejected at rearm so a
// corrupt pending.json cannot re-open the unbounded hole.
func (e *EmbeddedGoal) IsUnbounded() bool {
	return e != nil && e.MaxTurns == 0 && e.MaxDuration <= 0 && e.CostCap <= 0
}

// PendingRecord is the resume handoff payload stored in factory.db.
// REQ-AUTORESUME-006 mandates at least schema_version, body, directives,
// conversation_language, and saved_at.
type PendingRecord struct {
	ID                   int64      `json:"-"`
	SchemaVersion        int        `json:"schema_version"`
	SpecID               string     `json:"spec_id,omitempty"`
	Phase                string     `json:"phase,omitempty"`
	SavedAt              time.Time  `json:"saved_at"`
	SavedBySession       string     `json:"saved_by_session,omitempty"`
	ConversationLanguage string     `json:"conversation_language,omitempty"`
	Directives           Directives `json:"directives"`
	// EmbeddedGoal, when non-nil, carries a live armed goal for /clear re-arm
	// (SPEC-INFINITE-GOAL-001 REQ-6). Populated by `moai handoff save` when a
	// goal is armed for the saving session.
	EmbeddedGoal *EmbeddedGoal `json:"embedded_goal,omitempty"`
	// Body is the verbatim paste-ready resume (6-block, cut-line markers included).
	Body string `json:"body"`
}

// handoffStateDir identifies the legacy pre-SQLite compatibility tree.
func handoffStateDir(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", "handoff")
}

// PendingPath returns the project-scoped factory.db path.
func PendingPath(projectDir string) string {
	path, err := homestate.FactoryDBPath(projectDir)
	if err != nil {
		return filepath.Join(handoffStateDir(projectDir), "pending.json")
	}
	return path
}

// SavePending writes a private resume row and never touches the SessionEnd
// memory-handoff table or its legacy pending.md compatibility file.
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

	row, err := pendingRow(rec)
	if err != nil {
		return err
	}
	db, err := homestate.OpenFactory(projectDir)
	if err != nil {
		return fmt.Errorf("open factory handoff store: %w", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.SaveResume(context.Background(), row); err != nil {
		return err
	}
	legacy := filepath.Join(handoffStateDir(projectDir), "pending.json")
	if err := os.Remove(legacy); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("retire legacy pending.json: %w", err)
	}
	return nil
}

func pendingRow(rec *PendingRecord) (homestate.ResumeHandoff, error) {
	directives, err := json.Marshal(rec.Directives)
	if err != nil {
		return homestate.ResumeHandoff{}, fmt.Errorf("marshal pending record: %w", err)
	}
	var embedded *string
	if rec.EmbeddedGoal != nil {
		raw, err := json.Marshal(rec.EmbeddedGoal)
		if err != nil {
			return homestate.ResumeHandoff{}, fmt.Errorf("marshal embedded goal: %w", err)
		}
		value := string(raw)
		embedded = &value
	}
	return homestate.ResumeHandoff{SchemaVersion: rec.SchemaVersion, SpecID: rec.SpecID,
		Phase: rec.Phase, SavedAt: rec.SavedAt, SavedBySession: rec.SavedBySession,
		ConversationLanguage: rec.ConversationLanguage, DirectivesJSON: string(directives),
		EmbeddedGoalJSON: embedded, Body: rec.Body}, nil
}

// ClearPending marks a pending resume row cleared and removes only the legacy
// resume file. It never touches memory handoffs.
func ClearPending(projectDir string) error {
	db, err := homestate.OpenFactory(projectDir)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if err := db.ClearPendingResume(context.Background()); err != nil {
		return err
	}
	legacy := filepath.Join(handoffStateDir(projectDir), "pending.json")
	if err := os.Remove(legacy); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy pending.json: %w", err)
	}
	return nil
}

// ReadPending reads the pending SQLite row, with read-only legacy compatibility.
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
	db, err := homestate.OpenFactory(projectDir)
	if err == nil {
		defer func() { _ = db.Close() }()
		ctx := context.Background()
		row, present, readErr := db.ReadPendingResume(ctx)
		if readErr != nil || present {
			if readErr != nil {
				return nil, false, readErr
			}
			return resumeRecord(row)
		}
		retired, retiredErr := db.LegacyResumeRetired(ctx)
		if retiredErr != nil {
			return nil, false, retiredErr
		}
		if retired {
			return nil, false, nil
		}
	} else if path, pathErr := homestate.FactoryDBPath(projectDir); pathErr == nil {
		if _, statErr := os.Stat(path); statErr == nil {
			return nil, true, fmt.Errorf("read factory handoff database: %w", err)
		}
	}
	// Read-only compatibility for a pre-migration pending.json. The next save or
	// explicit home migration moves the flow into factory.db.
	legacy := filepath.Join(handoffStateDir(projectDir), "pending.json")
	data, err := os.ReadFile(legacy)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read legacy pending.json: %w", err)
	}
	var rec PendingRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, true, fmt.Errorf("parse pending.json: %w", err)
	}
	return &rec, true, nil
}

func resumeRecord(row *homestate.ResumeHandoff) (*PendingRecord, bool, error) {
	if row == nil {
		return nil, false, nil
	}
	rec := &PendingRecord{ID: row.ID, SchemaVersion: row.SchemaVersion, SpecID: row.SpecID, Phase: row.Phase,
		SavedAt: row.SavedAt, SavedBySession: row.SavedBySession,
		ConversationLanguage: row.ConversationLanguage, Body: row.Body}
	if err := json.Unmarshal([]byte(row.DirectivesJSON), &rec.Directives); err != nil {
		return nil, true, fmt.Errorf("parse directives: %w", err)
	}
	if row.EmbeddedGoalJSON != nil {
		var goal EmbeddedGoal
		if err := json.Unmarshal([]byte(*row.EmbeddedGoalJSON), &goal); err != nil {
			return nil, true, fmt.Errorf("parse embedded goal: %w", err)
		}
		rec.EmbeddedGoal = &goal
	}
	return rec, true, nil
}

// ClaimPending atomically claims one pending resume row. A legacy pending.json
// is imported once before claiming so upgrades do not silently skip a resume.
func ClaimPending(projectDir, token string) (*PendingRecord, int64, bool, error) {
	db, err := homestate.OpenFactory(projectDir)
	if err != nil {
		return nil, 0, false, err
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	row, present, err := db.ClaimPendingResume(ctx, token)
	if err == nil && !present {
		legacy := filepath.Join(handoffStateDir(projectDir), "pending.json")
		data, readErr := os.ReadFile(legacy)
		if readErr == nil {
			var rec PendingRecord
			if err := json.Unmarshal(data, &rec); err != nil {
				return nil, 0, true, fmt.Errorf("parse pending.json: %w", err)
			}
			if rec.SchemaVersion == 0 {
				rec.SchemaVersion = PendingSchemaVersion
			}
			if rec.SavedAt.IsZero() {
				rec.SavedAt = time.Now()
			}
			legacyRow, err := pendingRow(&rec)
			if err != nil {
				return nil, 0, true, err
			}
			imported, err := db.ImportLegacyResume(ctx, legacyRow)
			if err != nil {
				return nil, 0, true, err
			}
			if imported {
				if err := os.Remove(legacy); err != nil && !os.IsNotExist(err) {
					return nil, 0, true, fmt.Errorf("retire legacy pending.json: %w", err)
				}
			}
			row, present, err = db.ClaimPendingResume(ctx, token)
		} else if !os.IsNotExist(readErr) {
			return nil, 0, true, fmt.Errorf("read legacy pending.json: %w", readErr)
		}
	}
	if err != nil || !present {
		return nil, 0, present, err
	}
	rec, _, err := resumeRecord(row)
	if err != nil {
		_ = db.FinishResume(ctx, row.ID, row.ClaimToken, "failed", err.Error())
		return nil, row.ID, true, err
	}
	return rec, row.ID, true, err
}

func ExpirePending(projectDir string, id int64) error {
	db, err := homestate.OpenFactory(projectDir)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	_, err = db.ExpireResumeIfPending(context.Background(), id)
	return err
}

func FinishClaim(projectDir string, id int64, success bool, detail, claimToken string) error {
	db, err := homestate.OpenFactory(projectDir)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	status := "failed"
	if success {
		status = "consumed"
	}
	if claimToken == "" {
		return fmt.Errorf("resume claim token is required")
	}
	return db.FinishResume(context.Background(), id, claimToken, status, detail)
}
