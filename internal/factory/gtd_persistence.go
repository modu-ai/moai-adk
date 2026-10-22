package factory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/privatefile"
)

type gtdExport struct {
	Version         int                 `json:"version"`
	LogicalRevision int64               `json:"logical_revision"`
	Items           []GTDItem           `json:"items"`
	Relations       []GTDRelation       `json:"relations"`
	Contracts       []gtdContractExport `json:"contracts,omitempty"`
	Missions        []gtdMissionExport  `json:"missions,omitempty"`
	Events          []gtdEventExport    `json:"events,omitempty"`
	Operations      []GTDOperation      `json:"operations,omitempty"`
}

type gtdContractExport struct {
	MissionID, PolicyVersion, ContractHash string
	ContractJSON                           []byte
	Approved                               int
}
type gtdMissionExport struct {
	MissionID, MissionMode, State, ContractHash, SnapshotHash, OwnerID string
	LeaseVersion                                                       int64
}
type gtdEventExport struct{ EventID, MissionID, Kind, PayloadHash, CreatedAt string }

func exportGovernance(ctx context.Context, db gtdQuerier) ([]gtdContractExport, []gtdMissionExport, []gtdEventExport, []GTDOperation, error) {
	var contracts []gtdContractExport
	rows, err := db.QueryContext(ctx, `SELECT mission_id,policy_version,contract_hash,contract_json,approved FROM gtd_contracts ORDER BY mission_id`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for rows.Next() {
		var v gtdContractExport
		if err := rows.Scan(&v.MissionID, &v.PolicyVersion, &v.ContractHash, &v.ContractJSON, &v.Approved); err != nil {
			_ = rows.Close()
			return nil, nil, nil, nil, err
		}
		contracts = append(contracts, v)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, nil, nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, nil, nil, err
	}
	var missions []gtdMissionExport
	rows, err = db.QueryContext(ctx, `SELECT mission_id,mission_mode,state,contract_hash,snapshot_hash,owner_id,lease_version FROM gtd_missions ORDER BY mission_id`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for rows.Next() {
		var v gtdMissionExport
		if err := rows.Scan(&v.MissionID, &v.MissionMode, &v.State, &v.ContractHash, &v.SnapshotHash, &v.OwnerID, &v.LeaseVersion); err != nil {
			_ = rows.Close()
			return nil, nil, nil, nil, err
		}
		missions = append(missions, v)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, nil, nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, nil, nil, err
	}
	var events []gtdEventExport
	rows, err = db.QueryContext(ctx, `SELECT event_id,mission_id,kind,payload_hash,created_at FROM gtd_events ORDER BY event_id`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for rows.Next() {
		var v gtdEventExport
		if err := rows.Scan(&v.EventID, &v.MissionID, &v.Kind, &v.PayloadHash, &v.CreatedAt); err != nil {
			_ = rows.Close()
			return nil, nil, nil, nil, err
		}
		events = append(events, v)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, nil, nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, nil, nil, err
	}
	var operations []GTDOperation
	rows, err = db.QueryContext(ctx, `SELECT operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at FROM gtd_operations ORDER BY operation_id`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for rows.Next() {
		op, err := scanGTDOperation(rows)
		if err != nil {
			_ = rows.Close()
			return nil, nil, nil, nil, err
		}
		operations = append(operations, op)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, nil, nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, nil, nil, err
	}
	return contracts, missions, events, operations, nil
}

func ExportGTD(ctx context.Context, store *BacklogStore, explicitOptIn bool) ([]byte, error) {
	if !explicitOptIn {
		return nil, errors.New("gtd export: explicit opt-in required")
	}
	db, err := openGTDDB(store)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()
	items, err := listGTDItems(ctx, db)
	if err != nil {
		return nil, err
	}
	relations, err := listGTDRelations(ctx, db)
	if err != nil {
		return nil, err
	}
	var raw string
	if err := db.QueryRowContext(ctx, `SELECT value FROM gtd_meta WHERE key='logical_revision'`).Scan(&raw); err != nil {
		return nil, err
	}
	revision, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	contracts, missions, events, operations, err := exportGovernance(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(gtdExport{Version: 2, LogicalRevision: revision, Items: items, Relations: relations, Contracts: contracts, Missions: missions, Events: events, Operations: operations})
}

func BackupGTDStore(ctx context.Context, store *BacklogStore, destination string, explicitOptIn bool) (result error) {
	if !explicitOptIn {
		return errors.New("gtd backup: explicit opt-in required")
	}
	if !filepath.IsAbs(destination) {
		return errors.New("gtd backup: absolute destination required")
	}
	if _, err := os.Lstat(destination); err == nil {
		return errors.New("gtd backup: destination exists")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, db.Close()) }()
	return vacuumGTDInto(ctx, db, destination)
}

type gtdVacuumExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// The SQL executor is entered only after exclusive native-private creation.
// Exec nil and the function-owned file Close are distinct observations; neither
// proves that SQLite's internal VFS has closed every kernel handle.
func vacuumGTDInto(ctx context.Context, db gtdVacuumExecutor, destination string) (result error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	var output *os.File
	defer func() {
		if output != nil {
			result = errors.Join(result, output.Close())
		}
	}()
	if err := privatefile.CreateExclusive(destination, func(f *os.File) { output = f }); err != nil {
		return err
	}
	before, err := output.Stat()
	if err != nil {
		return err
	}
	if !before.Mode().IsRegular() || before.Size() != 0 {
		return errors.New("gtd backup: new output is not an empty regular file")
	}
	quoted := `'` + strings.ReplaceAll(destination, `'`, `''`) + `'`
	if _, err := db.ExecContext(ctx, `VACUUM INTO `+quoted); err != nil {
		return mapBacklogEngineError("vacuum gtd backup", err)
	}
	if err := output.Sync(); err != nil {
		return err
	}
	after, err := os.Lstat(destination)
	if err != nil {
		return err
	}
	if !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return errors.New("gtd backup: output identity changed")
	}
	return privatefile.Check(destination)
}

func RestoreGTDStore(ctx context.Context, backup string, store *BacklogStore, explicitOptIn bool) (result error) {
	if !explicitOptIn {
		return errors.New("gtd restore: explicit opt-in required")
	}
	if !filepath.IsAbs(backup) {
		return errors.New("gtd restore: absolute backup required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	source, err := os.Open(backup)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, source.Close()) }()
	if err := ensureBacklogDirectory(filepath.Dir(store.EnginePath())); err != nil {
		return err
	}
	if _, err := os.Lstat(store.EnginePath()); err == nil {
		return errors.New("gtd restore: destination exists")
	}
	if err := restoreGTDSnapshot(ctx, source, store.EnginePath()); err != nil {
		return err
	}

	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, db.Close()) }()
	var integrity string
	if err := db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		return fmt.Errorf("gtd restore: integrity check failed")
	}
	return nil
}

// restoreGTDSnapshot keeps the staging/publication operation separate from
// schema validation without changing its full-file snapshot semantics.
func restoreGTDSnapshot(ctx context.Context, source io.Reader, destination string) (result error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	tmp, err := privatefile.CreateTemp(filepath.Dir(destination), ".gtd-restore-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	closed := false
	defer func() {
		if !closed {
			result = errors.Join(result, tmp.Close())
		}
		if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
			result = errors.Join(result, err)
		}
	}()
	if _, err := io.Copy(tmp, gtdSnapshotReader{ctx: ctx, reader: source}); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	closeErr := tmp.Close()
	closed = true // Do not retry a Close whose native outcome is uncertain.
	if closeErr != nil {
		return closeErr
	}
	if err := privatefile.Check(name); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return atomicfile.RenameNoReplace(name, destination)
}

type gtdSnapshotReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r gtdSnapshotReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func ImportGTD(ctx context.Context, store *BacklogStore, data []byte, explicitOptIn bool) error {
	if !explicitOptIn {
		return errors.New("gtd import: explicit opt-in required")
	}
	var payload gtdExport
	if json.Unmarshal(data, &payload) != nil || (payload.Version != 1 && payload.Version != 2) || payload.LogicalRevision < 0 {
		return errors.New("gtd import: invalid payload")
	}
	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM gtd_items)+(SELECT count(*) FROM gtd_relations)`).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return errors.New("gtd import: destination not empty")
	}
	for _, i := range payload.Items {
		trusted, cancelled := 0, 0
		if i.SourceTrusted {
			trusted = 1
		}
		if i.Cancelled {
			cancelled = 1
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_items(item_id,content,source,sensitivity,event_id,outcome,completion_evidence,disposition,class,action_context,review_at,authority,source_trust,status,source_revision,card_id,cancelled) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, i.ItemID, i.Content, i.Source, string(i.Sensitivity), i.EventID, i.DesiredOutcome, i.CompletionEvidence, string(i.Disposition), string(i.Class), i.Context, i.ReviewAt, i.Authority, trusted, string(i.Status), i.SourceRevision, nilIfEmpty(i.CardID), cancelled); err != nil {
			return err
		}
	}
	for _, r := range payload.Relations {
		note := r.Note
		if note == nil {
			note = []byte{}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_relations(subject_id,object_id,kind,note,source,assertion_status,source_revision,policy_version) VALUES(?,?,?,?,?,?,?,?)`, r.SubjectID, r.ObjectID, string(r.Kind), note, r.Source, r.AssertionStatus, r.SourceRevision, r.PolicyVersion); err != nil {
			return err
		}
	}
	for _, v := range payload.Contracts {
		if v.MissionID == "" || v.PolicyVersion == "" || v.ContractHash == "" || len(v.ContractJSON) == 0 || (v.Approved != 0 && v.Approved != 1) {
			return errors.New("gtd import: invalid contract")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_contracts VALUES(?,?,?,?,?)`, v.MissionID, v.PolicyVersion, v.ContractHash, v.ContractJSON, v.Approved); err != nil {
			return err
		}
	}
	for _, v := range payload.Missions {
		if v.MissionID == "" || v.MissionMode != "auto" || v.ContractHash == "" || v.SnapshotHash == "" || v.LeaseVersion < 0 {
			return errors.New("gtd import: invalid mission")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_missions VALUES(?,?,?,?,?,?,?)`, v.MissionID, v.MissionMode, v.State, v.ContractHash, v.SnapshotHash, v.OwnerID, v.LeaseVersion); err != nil {
			return err
		}
	}
	for _, v := range payload.Events {
		if v.EventID == "" || v.Kind == "" || v.PayloadHash == "" || v.CreatedAt == "" {
			return errors.New("gtd import: invalid event")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_events VALUES(?,?,?,?,?)`, v.EventID, v.MissionID, v.Kind, v.PayloadHash, v.CreatedAt); err != nil {
			return err
		}
	}
	for _, v := range payload.Operations {
		if !validGTDOperation(v) {
			return errors.New("gtd import: invalid operation")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_operations VALUES(?,?,?,?,?,?,?,?)`, v.OperationID, v.MissionID, v.Action, v.Target, string(v.State), v.ReceiptJSON, v.SnapshotHash, v.UpdatedAt); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE gtd_meta SET value=? WHERE key='logical_revision'`, strconv.FormatInt(payload.LogicalRevision, 10)); err != nil {
		return err
	}
	return tx.Commit()
}
func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func SetGTDItemCancelled(ctx context.Context, store *BacklogStore, itemID string, cancelled bool) error {
	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	value := 0
	if cancelled {
		value = 1
	}
	res, err := tx.ExecContext(ctx, `UPDATE gtd_items SET cancelled=?,source_revision=source_revision+1 WHERE item_id=?`, value, itemID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("gtd lifecycle: item not found")
	}
	if err := bumpGTDRevision(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}
func DeleteGTDItem(ctx context.Context, store *BacklogStore, itemID string) error {
	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM gtd_relations WHERE subject_id=? OR object_id=?`, itemID, itemID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM gtd_items WHERE item_id=?`, itemID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("gtd lifecycle: item not found")
	}
	if err := bumpGTDRevision(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}
