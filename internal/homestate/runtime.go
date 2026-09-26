package homestate

import (
	"context"
	"fmt"
	"time"
)

// FactoryRun is a factory run record. LeadPID and LeadProcessStart carry the
// run's owner identity — the process id and process-start fingerprint of the
// process the caller is in a position to name at record time. On a
// replace-shaped launch door that process IS the session; on the spawn and
// pane shapes the launcher restamps the row once the session identity exists
// (REQ-002b).
type FactoryRun struct {
	RunID, LeadSessionID, Backend, ManifestJSON string
	LeadPID                                     int
	LeadProcessStart                            string
}

func (f *FactoryDB) RecordRun(ctx context.Context, row FactoryRun) error {
	if row.RunID == "" {
		return fmt.Errorf("factory run id is empty")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// The owner stamp joins the row's own transaction rather than a follow-up
	// write: a stamp lost against the row it describes leaves an active run
	// with an empty owner column, which stays indeterminate — never
	// auto-retired — until the REQ-006b boot proof holds for it after the next
	// host reboot (REQ-006): a second generator of the defect this fixes
	// (REQ-002).
	if _, err := tx.ExecContext(ctx, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,created_at,updated_at)
VALUES(?,?,?,'active',?,?,?,?,?) ON CONFLICT(run_id) DO UPDATE SET
lead_session_id=excluded.lead_session_id,lead_backend=excluded.lead_backend,status='active',manifest_json=excluded.manifest_json,lead_pid=excluded.lead_pid,lead_process_start=excluded.lead_process_start,updated_at=excluded.updated_at`,
		row.RunID, row.LeadSessionID, row.Backend, row.ManifestJSON, row.LeadPID, row.LeadProcessStart, now, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.started',?,?)`, row.RunID, row.ManifestJSON, now); err != nil {
		return err
	}
	return tx.Commit()
}

type FactoryCard struct {
	RunID, CardID, OwnerLabel, State, EvidencePath, EventKind, PayloadJSON string
}

func (f *FactoryDB) RecordCard(ctx context.Context, row FactoryCard) error {
	if row.RunID == "" || row.CardID == "" {
		return fmt.Errorf("factory run id and card id are required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO runs(run_id,status,manifest_json,created_at,updated_at) VALUES(?,'active','{}',?,?)`, row.RunID, now, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO cards(run_id,card_id,owner_label,state,version,evidence_path,updated_at)
VALUES(?,?,?,?,1,?,?) ON CONFLICT(run_id,card_id) DO UPDATE SET
owner_label=excluded.owner_label,state=excluded.state,version=cards.version+1,evidence_path=excluded.evidence_path,updated_at=excluded.updated_at`,
		row.RunID, row.CardID, row.OwnerLabel, row.State, row.EvidencePath, now); err != nil {
		return err
	}
	kind := row.EventKind
	if kind == "" {
		kind = "card.updated"
	}
	payload := row.PayloadJSON
	if payload == "" {
		payload = "{}"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,?,?,?)`, row.RunID, kind, payload, now); err != nil {
		return err
	}
	return tx.Commit()
}
