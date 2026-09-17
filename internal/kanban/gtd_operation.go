package kanban

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type GTDOperationState string

const (
	GTDOperationPrepared   GTDOperationState = "prepared"
	GTDOperationInvoking   GTDOperationState = "invoking"
	GTDOperationReconciled GTDOperationState = "reconciled"
	GTDOperationBlocked    GTDOperationState = "blocked"
)

type GTDOperation struct {
	OperationID  string            `json:"operation_id"`
	MissionID    string            `json:"mission_id"`
	Action       string            `json:"action"`
	Target       string            `json:"target"`
	State        GTDOperationState `json:"state"`
	ReceiptJSON  []byte            `json:"receipt"`
	SnapshotHash string            `json:"snapshot_hash"`
	UpdatedAt    string            `json:"updated_at"`
}

type GTDOperationOwner interface {
	Readback(context.Context, GTDOperation) (bool, error)
	Apply(context.Context, GTDOperation) error
}

func validGTDOperation(op GTDOperation) bool {
	return strings.TrimSpace(op.OperationID) != "" && strings.TrimSpace(op.MissionID) != "" && strings.TrimSpace(op.Action) != "" && strings.TrimSpace(op.Target) != "" && strings.TrimSpace(op.SnapshotHash) != "" && len(op.ReceiptJSON) > 0
}

func scanGTDOperation(row interface{ Scan(...any) error }) (GTDOperation, error) {
	var op GTDOperation
	var state string
	err := row.Scan(&op.OperationID, &op.MissionID, &op.Action, &op.Target, &state, &op.ReceiptJSON, &op.SnapshotHash, &op.UpdatedAt)
	op.State = GTDOperationState(state)
	return op, err
}

func LoadGTDOperation(ctx context.Context, store *BacklogStore, operationID string) (GTDOperation, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return GTDOperation{}, err
	}
	defer func() { _ = db.Close() }()
	op, err := scanGTDOperation(db.QueryRowContext(ctx, `SELECT operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at FROM gtd_operations WHERE operation_id=?`, operationID))
	if err != nil {
		return GTDOperation{}, mapBacklogEngineError("load gtd operation", err)
	}
	return op, nil
}

func PrepareGTDOperation(ctx context.Context, store *BacklogStore, candidate GTDOperation) (GTDOperation, bool, error) {
	if !validGTDOperation(candidate) {
		return GTDOperation{}, false, errors.New("gtd operation: invalid")
	}
	db, err := openGTDDB(store)
	if err != nil {
		return GTDOperation{}, false, err
	}
	defer func() { _ = db.Close() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO gtd_operations(operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at) VALUES(?,?,?,?,?,?,?,?)`, candidate.OperationID, candidate.MissionID, candidate.Action, candidate.Target, string(GTDOperationPrepared), candidate.ReceiptJSON, candidate.SnapshotHash, now)
	if err != nil {
		return GTDOperation{}, false, mapBacklogEngineError("prepare gtd operation", err)
	}
	rows, _ := res.RowsAffected()
	stored, err := scanGTDOperation(db.QueryRowContext(ctx, `SELECT operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at FROM gtd_operations WHERE operation_id=?`, candidate.OperationID))
	if err != nil {
		return GTDOperation{}, false, mapBacklogEngineError("read prepared gtd operation", err)
	}
	if stored.MissionID != candidate.MissionID || stored.Action != candidate.Action || stored.Target != candidate.Target || stored.SnapshotHash != candidate.SnapshotHash || !bytes.Equal(stored.ReceiptJSON, candidate.ReceiptJSON) {
		return GTDOperation{}, false, errors.New("gtd operation: identity_collision")
	}
	return stored, rows == 1, nil
}

func markGTDOperationState(ctx context.Context, store *BacklogStore, operationID string, from, to GTDOperationState) error {
	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	res, err := db.ExecContext(ctx, `UPDATE gtd_operations SET state=?,updated_at=? WHERE operation_id=? AND state=?`, string(to), time.Now().UTC().Format(time.RFC3339Nano), operationID, string(from))
	if err != nil {
		return mapBacklogEngineError("transition gtd operation", err)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("gtd operation: transition_conflict")
	}
	return nil
}

// ExecuteGTDOperation makes the SQLite receipt the effect gate. An uncertain
// invoking record is never retried from a timeout: only authoritative owner
// readback may reconcile it, preventing a second external effect after crash.
func ExecuteGTDOperation(ctx context.Context, store *BacklogStore, candidate GTDOperation, owner GTDOperationOwner) (GTDOperation, error) {
	if owner == nil {
		return GTDOperation{}, errors.New("gtd operation: nil owner")
	}
	stored, _, err := PrepareGTDOperation(ctx, store, candidate)
	if err != nil {
		return GTDOperation{}, err
	}
	applied, err := owner.Readback(ctx, stored)
	if err != nil {
		return stored, err
	}
	if applied {
		if stored.State == GTDOperationReconciled {
			return stored, nil
		}
		if err := markGTDOperationState(ctx, store, stored.OperationID, stored.State, GTDOperationReconciled); err != nil {
			return LoadGTDOperation(ctx, store, stored.OperationID)
		}
		return LoadGTDOperation(ctx, store, stored.OperationID)
	}
	if stored.State != GTDOperationPrepared {
		return stored, nil
	}
	if err := markGTDOperationState(ctx, store, stored.OperationID, GTDOperationPrepared, GTDOperationInvoking); err != nil {
		return LoadGTDOperation(ctx, store, stored.OperationID)
	}
	stored.State = GTDOperationInvoking
	if err := owner.Apply(ctx, stored); err != nil {
		return stored, err
	}
	applied, err = owner.Readback(ctx, stored)
	if err != nil {
		return stored, err
	}
	if !applied {
		return stored, errors.New("gtd operation: authoritative_readback_missing")
	}
	if err := markGTDOperationState(ctx, store, stored.OperationID, GTDOperationInvoking, GTDOperationReconciled); err != nil {
		return stored, err
	}
	return LoadGTDOperation(ctx, store, stored.OperationID)
}
