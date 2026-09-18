package kanban

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type GTDSensitivity string

const (
	SensitivityUnknown GTDSensitivity = "unknown"
	SensitivityPublic  GTDSensitivity = "public"
	SensitivityPrivate GTDSensitivity = "private"
	SensitivitySecret  GTDSensitivity = "secret"
)

type GTDStatus string

const (
	GTDStatusCaptured  GTDStatus = "captured"
	GTDStatusClarified GTDStatus = "clarified"
	GTDStatusHeld      GTDStatus = "held"
	GTDStatusOrganized GTDStatus = "organized"
	GTDStatusPublished GTDStatus = "published"
)

type GTDItem struct {
	ItemID             string         `json:"item_id"`
	Content            string         `json:"content"`
	Source             string         `json:"source"`
	Sensitivity        GTDSensitivity `json:"sensitivity"`
	EventID            string         `json:"event_id"`
	DesiredOutcome     string         `json:"desired_outcome"`
	CompletionEvidence string         `json:"completion_evidence"`
	Disposition        GTDDisposition `json:"disposition"`
	Class              GTDClass       `json:"class"`
	Context            string         `json:"context"`
	ReviewAt           string         `json:"review_at"`
	Authority          string         `json:"authority"`
	SourceTrusted      bool           `json:"source_trusted"`
	Status             GTDStatus      `json:"status"`
	SourceRevision     int64          `json:"source_revision"`
	CardID             string         `json:"card_id"`
	Cancelled          bool           `json:"cancelled"`
}

type CaptureInput struct {
	Content       string
	Source        string
	SourceAllowed bool
	Sensitivity   GTDSensitivity
	EventID       string
}

func openGTDDB(store *BacklogStore) (*sql.DB, error) {
	if err := MigrateGTDSchema(store); err != nil {
		return nil, err
	}
	db, err := sql.Open(sqliteDriverName, backlogDSN(store.EnginePath()))
	if err != nil {
		return nil, mapBacklogEngineError("open gtd store", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// CaptureGTDItem records an inbox item without creating a queue card. Source
// authorization and privacy classification are deterministic inputs, never
// inferred from the captured text.
func CaptureGTDItem(ctx context.Context, store *BacklogStore, in CaptureInput) (GTDItem, error) {
	if !in.SourceAllowed {
		return GTDItem{}, errors.New("gtd capture: source_not_allowed")
	}
	if in.Sensitivity != SensitivityPublic && in.Sensitivity != SensitivityPrivate && in.Sensitivity != SensitivitySecret {
		return GTDItem{}, errors.New("gtd capture: sensitivity_unclassified")
	}
	if strings.TrimSpace(in.Content) == "" || strings.TrimSpace(in.Source) == "" || strings.TrimSpace(in.EventID) == "" {
		return GTDItem{}, errors.New("gtd capture: incomplete_input")
	}
	sum := sha256.Sum256([]byte("gtd-capture\x00" + in.EventID))
	id := "gtd-" + hex.EncodeToString(sum[:8])
	db, err := openGTDDB(store)
	if err != nil {
		return GTDItem{}, err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return GTDItem{}, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO gtd_items(item_id,content,source,sensitivity,event_id,status,source_revision) VALUES(?,?,?,?,?,?,1)`, id, in.Content, in.Source, string(in.Sensitivity), in.EventID, string(GTDStatusCaptured))
	if err != nil {
		return GTDItem{}, mapBacklogEngineError("capture gtd item", err)
	}
	if n, _ := res.RowsAffected(); n == 1 {
		if err := bumpGTDRevision(ctx, tx); err != nil {
			return GTDItem{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return GTDItem{}, err
	}
	return loadGTDItem(ctx, db, id)
}

type gtdQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadGTDItem(ctx context.Context, db gtdQueryRower, id string) (GTDItem, error) {
	var item GTDItem
	var sensitivity, disposition, class, status string
	var trusted string
	var cancelled int
	err := db.QueryRowContext(ctx, `SELECT item_id,content,source,sensitivity,event_id,outcome,completion_evidence,disposition,class,action_context,review_at,authority,source_trust,status,source_revision,COALESCE(card_id,''),cancelled FROM gtd_items WHERE item_id=?`, id).Scan(
		&item.ItemID, &item.Content, &item.Source, &sensitivity, &item.EventID, &item.DesiredOutcome, &item.CompletionEvidence, &disposition, &class, &item.Context, &item.ReviewAt, &item.Authority, &trusted, &status, &item.SourceRevision, &item.CardID, &cancelled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GTDItem{}, fmt.Errorf("gtd item %s: not_found", id)
		}
		return GTDItem{}, mapBacklogEngineError("load gtd item", err)
	}
	item.Sensitivity = GTDSensitivity(sensitivity)
	item.Disposition = GTDDisposition(disposition)
	item.Class = GTDClass(class)
	item.Status = GTDStatus(status)
	item.SourceTrusted = trusted == "1"
	item.Cancelled = cancelled == 1
	return item, nil
}

func LoadGTDItem(ctx context.Context, store *BacklogStore, id string) (GTDItem, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return GTDItem{}, err
	}
	defer func() { _ = db.Close() }()
	return loadGTDItem(ctx, db, id)
}
