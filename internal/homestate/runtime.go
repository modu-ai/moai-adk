package homestate

import (
	"context"
	"fmt"
	"time"
)

type FactoryRun struct {
	RunID, LeadSessionID, Backend, ManifestJSON string
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
	if _, err := tx.ExecContext(ctx, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at)
VALUES(?,?,?,'active',?,?,?) ON CONFLICT(run_id) DO UPDATE SET
lead_session_id=excluded.lead_session_id,lead_backend=excluded.lead_backend,status='active',manifest_json=excluded.manifest_json,updated_at=excluded.updated_at`,
		row.RunID, row.LeadSessionID, row.Backend, row.ManifestJSON, now, now); err != nil {
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
