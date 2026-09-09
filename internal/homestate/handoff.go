package homestate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type ResumeHandoff struct {
	ID                   int64
	SchemaVersion        int
	SpecID               string
	Phase                string
	SavedAt              time.Time
	SavedBySession       string
	ConversationLanguage string
	DirectivesJSON       string
	EmbeddedGoalJSON     *string
	Body                 string
}

func (f *FactoryDB) LegacyResumeRetired(ctx context.Context) (bool, error) {
	var marker string
	err := f.DB.QueryRowContext(ctx, `SELECT value FROM meta WHERE key='legacy_resume_retired'`).Scan(&marker)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (f *FactoryDB) SaveResume(ctx context.Context, row ResumeHandoff) error {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at)
SELECT 'resume',id,'pending','cleared','superseded by newer save',? FROM resume_handoffs WHERE status='pending'`, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status='cleared' WHERE status='pending'`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO meta(key,value) VALUES('legacy_resume_retired','1')
ON CONFLICT(key) DO UPDATE SET value=excluded.value`); err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(row.Body))
	result, err := tx.ExecContext(ctx, `INSERT INTO resume_handoffs(
status,schema_version,spec_id,phase,saved_at,saved_by_session,conversation_language,directives_json,embedded_goal_json,body,body_sha256)
VALUES('pending',?,?,?,?,?,?,?,?,?,?)`, row.SchemaVersion, row.SpecID, row.Phase,
		row.SavedAt.UTC().Format(time.RFC3339Nano), row.SavedBySession, row.ConversationLanguage,
		row.DirectivesJSON, row.EmbeddedGoalJSON, row.Body, hex.EncodeToString(sum[:]))
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,to_status,created_at) VALUES('resume',?,'pending',?)`, id, now); err != nil {
		return err
	}
	return tx.Commit()
}

// ImportLegacyResume records one pre-SQLite pending.json exactly once. The
// meta marker makes the import idempotent even when several SessionStart hooks
// race after one caller has already claimed the imported row.
func (f *FactoryDB) ImportLegacyResume(ctx context.Context, row ResumeHandoff) (bool, error) {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	var marker string
	err = tx.QueryRowContext(ctx, `SELECT value FROM meta WHERE key='legacy_resume_retired'`).Scan(&marker)
	if err == nil {
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	var pending int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM resume_handoffs WHERE status='pending'`).Scan(&pending); err != nil {
		return false, err
	}
	if pending != 0 {
		return false, nil
	}
	sum := sha256.Sum256([]byte(row.Body))
	result, err := tx.ExecContext(ctx, `INSERT INTO resume_handoffs(
status,schema_version,spec_id,phase,saved_at,saved_by_session,conversation_language,directives_json,embedded_goal_json,body,body_sha256)
VALUES('pending',?,?,?,?,?,?,?,?,?,?)`, row.SchemaVersion, row.SpecID, row.Phase,
		row.SavedAt.UTC().Format(time.RFC3339Nano), row.SavedBySession, row.ConversationLanguage,
		row.DirectivesJSON, row.EmbeddedGoalJSON, row.Body, hex.EncodeToString(sum[:]))
	if err != nil {
		return false, err
	}
	id, _ := result.LastInsertId()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,to_status,detail,created_at) VALUES('resume',?,'pending','legacy pending.json import',?)`, id, now); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO meta(key,value) VALUES('legacy_resume_retired','1')`); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (f *FactoryDB) ReadPendingResume(ctx context.Context) (*ResumeHandoff, bool, error) {
	row := &ResumeHandoff{}
	var saved string
	err := f.DB.QueryRowContext(ctx, `SELECT id,schema_version,spec_id,phase,saved_at,saved_by_session,
conversation_language,directives_json,embedded_goal_json,body FROM resume_handoffs
WHERE status='pending' ORDER BY id DESC LIMIT 1`).Scan(&row.ID, &row.SchemaVersion, &row.SpecID,
		&row.Phase, &saved, &row.SavedBySession, &row.ConversationLanguage, &row.DirectivesJSON,
		&row.EmbeddedGoalJSON, &row.Body)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	row.SavedAt, err = time.Parse(time.RFC3339Nano, saved)
	if err != nil {
		return nil, true, fmt.Errorf("parse saved_at: %w", err)
	}
	return row, true, nil
}

// ClaimPendingResume atomically elects one consumer. The update predicate is
// the compare-and-swap gate; exactly one concurrent caller changes one row.
func (f *FactoryDB) ClaimPendingResume(ctx context.Context, token string) (*ResumeHandoff, bool, error) {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()
	row := &ResumeHandoff{}
	var saved string
	err = tx.QueryRowContext(ctx, `SELECT id,schema_version,spec_id,phase,saved_at,saved_by_session,
conversation_language,directives_json,embedded_goal_json,body FROM resume_handoffs
WHERE status='pending' ORDER BY id DESC LIMIT 1`).Scan(&row.ID, &row.SchemaVersion, &row.SpecID,
		&row.Phase, &saved, &row.SavedBySession, &row.ConversationLanguage, &row.DirectivesJSON,
		&row.EmbeddedGoalJSON, &row.Body)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status='claimed',claim_token=?,claimed_at=? WHERE id=? AND status='pending'`, token, now, row.ID)
	if err != nil {
		return nil, false, err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return nil, false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,'pending','claimed',?,?)`, row.ID, token, now); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	row.SavedAt, err = time.Parse(time.RFC3339Nano, saved)
	return row, true, err
}

func (f *FactoryDB) SetResumeStatus(ctx context.Context, id int64, from, to, detail string) error {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status=?,consumed_at=CASE WHEN ?='consumed' THEN ? ELSE consumed_at END,error=CASE WHEN ?='failed' THEN ? ELSE error END WHERE id=? AND status=?`, to, to, now, to, detail, id, from)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return fmt.Errorf("resume handoff %d is not %s", id, from)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,?,?,?,?)`, id, from, to, detail, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (f *FactoryDB) ClearPendingResume(ctx context.Context) error {
	return f.setPendingResumeStatus(ctx, "cleared", "explicit clear")
}

func (f *FactoryDB) ExpirePendingResume(ctx context.Context) error {
	return f.setPendingResumeStatus(ctx, "expired", "stale TTL")
}

func (f *FactoryDB) ExpireResumeIfPending(ctx context.Context, id int64) (bool, error) {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status='expired',consumed_at=? WHERE id=? AND status='pending'`, now, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil || n == 0 {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,'pending','expired','stale TTL',?)`, id, now); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (f *FactoryDB) setPendingResumeStatus(ctx context.Context, status, detail string) error {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT id FROM resume_handoffs WHERE status='pending'`)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status=?,consumed_at=? WHERE status='pending'`, status, now); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,'pending',?,?,?)`, id, status, detail, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type MemoryHandoff struct {
	ID         int64
	Sprint     string
	Spec       string
	Status     string
	Body       string
	IndexLine  string
	Supersedes string
}

func (f *FactoryDB) SaveMemory(ctx context.Context, row MemoryHandoff) error {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at)
SELECT 'memory',id,'pending','expired','superseded by newer save',? FROM memory_handoffs WHERE status='pending'`, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE memory_handoffs SET status='expired' WHERE status='pending'`); err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(row.Body))
	result, err := tx.ExecContext(ctx, `INSERT INTO memory_handoffs(status,sprint,spec,result_status,body,index_line,supersedes,body_sha256,created_at) VALUES('pending',?,?,?,?,?,?,?,?)`, row.Sprint, row.Spec, row.Status, row.Body, row.IndexLine, row.Supersedes, hex.EncodeToString(sum[:]), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,to_status,created_at) VALUES('memory',?,'pending',?)`, id, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (f *FactoryDB) ReadPendingMemory(ctx context.Context) (*MemoryHandoff, bool, error) {
	row := &MemoryHandoff{}
	err := f.DB.QueryRowContext(ctx, `SELECT id,sprint,spec,result_status,body,index_line,supersedes FROM memory_handoffs WHERE status='pending' ORDER BY id DESC LIMIT 1`).Scan(&row.ID, &row.Sprint, &row.Spec, &row.Status, &row.Body, &row.IndexLine, &row.Supersedes)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return row, true, nil
}

func (f *FactoryDB) SetMemoryStatus(ctx context.Context, id int64, to, detail string) error {
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE memory_handoffs SET status=?,persisted_at=CASE WHEN ?='persisted' THEN ? ELSE persisted_at END,error=? WHERE id=? AND status='pending'`, to, to, now, detail, id)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return fmt.Errorf("memory handoff %d is not pending", id)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('memory',?,'pending',?,?,?)`, id, to, detail, now); err != nil {
		return err
	}
	return tx.Commit()
}
