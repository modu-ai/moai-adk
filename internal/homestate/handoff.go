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
	ClaimToken           string
	ClaimExpiresAt       *time.Time
	ClaimOwnerPID        int
	ClaimOwnerSession    string
}

type ResumeClaim struct {
	Token            string
	OwnerPID         int
	OwnerSession     string
	OwnerFingerprint string
	TTL              time.Duration
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
	return f.ClaimResume(ctx, ResumeClaim{Token: token, TTL: 5 * time.Minute})
}

// ClaimResume provides at-least-once delivery: pending work has priority and,
// only when none exists, an expired non-legacy claim can be atomically reclaimed.
func (f *FactoryDB) ClaimResume(ctx context.Context, claim ResumeClaim) (*ResumeHandoff, bool, error) {
	if claim.Token == "" {
		return nil, false, fmt.Errorf("claim token is required")
	}
	if claim.TTL == 0 {
		claim.TTL = 5 * time.Minute
	}
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()
	row := &ResumeHandoff{}
	var saved string
	var prior string
	err = tx.QueryRowContext(ctx, `SELECT id,schema_version,spec_id,phase,saved_at,saved_by_session,
conversation_language,directives_json,embedded_goal_json,body,status FROM resume_handoffs
WHERE status='pending' OR (status='claimed' AND legacy_recovery=0 AND claim_expires_at IS NOT NULL AND claim_expires_at<=?)
ORDER BY CASE status WHEN 'pending' THEN 0 ELSE 1 END,id DESC LIMIT 1`, time.Now().UTC().Format(time.RFC3339Nano)).Scan(&row.ID, &row.SchemaVersion, &row.SpecID,
		&row.Phase, &saved, &row.SavedBySession, &row.ConversationLanguage, &row.DirectivesJSON,
		&row.EmbeddedGoalJSON, &row.Body, &prior)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339Nano)
	expiry := nowTime.Add(claim.TTL).Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status='claimed',claim_token=?,claimed_at=?,claim_expires_at=?,claim_owner_pid=?,claim_owner_session=?,claim_owner_fingerprint=? WHERE id=? AND status=? AND (status='pending' OR (legacy_recovery=0 AND claim_expires_at<=?))`, claim.Token, now, expiry, claim.OwnerPID, claim.OwnerSession, claim.OwnerFingerprint, row.ID, prior, now)
	if err != nil {
		return nil, false, err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return nil, false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,?,'claimed',?,?)`, row.ID, prior, claim.Token, now); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	row.SavedAt, err = time.Parse(time.RFC3339Nano, saved)
	row.ClaimToken = claim.Token
	row.ClaimOwnerPID = claim.OwnerPID
	row.ClaimOwnerSession = claim.OwnerSession
	if parsed, parseErr := time.Parse(time.RFC3339Nano, expiry); parseErr == nil {
		row.ClaimExpiresAt = &parsed
	}
	return row, true, err
}

func (f *FactoryDB) FinishResume(ctx context.Context, id int64, token, to, detail string) error {
	if to != "consumed" && to != "failed" {
		return fmt.Errorf("invalid finish status %q", to)
	}
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status=?,consumed_at=CASE WHEN ?='consumed' THEN ? ELSE consumed_at END,error=CASE WHEN ?='failed' THEN ? ELSE error END WHERE id=? AND status='claimed' AND claim_token=?`, to, to, now, to, detail, id, token)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return fmt.Errorf("resume handoff %d stale claim token", id)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,'claimed',?,?,?)`, id, to, detail, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (f *FactoryDB) RecoverLegacyResume(ctx context.Context, id int64, expectedToken, decision string, probe func(int) (string, ProcessIdentityState), verifyZeroActive func() error) error {
	if decision != "requeue" && decision != "fail" {
		return fmt.Errorf("invalid recovery decision %q", decision)
	}
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var pid sql.NullInt64
	var fingerprint string
	if err := tx.QueryRowContext(ctx, `SELECT claim_owner_pid,claim_owner_fingerprint FROM resume_handoffs WHERE id=? AND status='claimed' AND claim_token=? AND legacy_recovery=1`, id, expectedToken).Scan(&pid, &fingerprint); err != nil {
		return fmt.Errorf("legacy resume recovery CAS precondition: %w", err)
	}
	if pid.Valid {
		got, state := probe(int(pid.Int64))
		if state != ProcessIdentityDead {
			if state == ProcessIdentityLive && got != fingerprint && fingerprint != "" { /* PID reuse is stale */
			} else {
				return fmt.Errorf("legacy resume owner is %s", state)
			}
		}
	} else {
		if verifyZeroActive == nil {
			return fmt.Errorf("legacy resume unknown owner requires zero-active census")
		}
		if err := verifyZeroActive(); err != nil {
			return fmt.Errorf("legacy resume unknown owner census: %w", err)
		}
	}
	status := "failed"
	if decision == "requeue" {
		status = "pending"
	}
	res, err := tx.ExecContext(ctx, `UPDATE resume_handoffs SET status=?,claim_token='',claimed_at=NULL,claim_expires_at=NULL,claim_owner_pid=NULL,claim_owner_session='',claim_owner_fingerprint='',legacy_recovery=0,legacy_recovery_reason='',error=CASE WHEN ?='failed' THEN 'operator legacy recovery' ELSE error END WHERE id=? AND status='claimed' AND claim_token=? AND legacy_recovery=1`, status, status, id, expectedToken)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return fmt.Errorf("legacy resume recovery stale CAS")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,detail,created_at) VALUES('resume',?,'claimed',?,'operator legacy recovery',?)`, id, status, now); err != nil {
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
