package kanban

import (
	"context"
	"errors"
	"strings"
)

type GTDDisposition string

const (
	DispositionAction    GTDDisposition = "action"
	DispositionReference GTDDisposition = "reference"
	DispositionSomeday   GTDDisposition = "someday"
	DispositionWaiting   GTDDisposition = "waiting"
	DispositionTrash     GTDDisposition = "trash"
)

type ClarifyInput struct {
	ItemID             string
	Disposition        GTDDisposition
	DesiredOutcome     string
	CompletionEvidence string
	Authority          string
	SourceTrusted      bool
}

type ClarifyResult struct {
	Item        GTDItem
	Publishable bool
	HoldReasons []string
}

func knownGTDDisposition(disposition GTDDisposition) bool {
	switch disposition {
	case DispositionAction, DispositionReference, DispositionSomeday, DispositionWaiting, DispositionTrash:
		return true
	default:
		return false
	}
}

func ClarifyGTDItem(ctx context.Context, store *BacklogStore, in ClarifyInput) (ClarifyResult, error) {
	if strings.TrimSpace(in.ItemID) == "" {
		return ClarifyResult{}, errors.New("gtd clarify: missing_item")
	}
	if !knownGTDDisposition(in.Disposition) {
		return ClarifyResult{}, errors.New("gtd clarify: invalid_disposition")
	}
	db, err := openGTDDB(store)
	if err != nil {
		return ClarifyResult{}, err
	}
	defer func() { _ = db.Close() }()
	if _, err := loadGTDItem(ctx, db, in.ItemID); err != nil {
		return ClarifyResult{}, err
	}
	var holds []string
	if in.Disposition == DispositionAction {
		if strings.TrimSpace(in.DesiredOutcome) == "" {
			holds = append(holds, "goal_unresolved")
		}
		if strings.TrimSpace(in.CompletionEvidence) == "" {
			holds = append(holds, "completion_evidence_unresolved")
		}
		if strings.TrimSpace(in.Authority) == "" {
			holds = append(holds, "authority_unresolved")
		}
		if !in.SourceTrusted {
			holds = append(holds, "source_untrusted")
		}
	}
	status := GTDStatusClarified
	if len(holds) > 0 {
		status = GTDStatusHeld
	}
	trusted := 0
	if in.SourceTrusted {
		trusted = 1
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ClarifyResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE gtd_items SET outcome=?,completion_evidence=?,disposition=?,authority=?,source_trust=?,status=?,source_revision=source_revision+1 WHERE item_id=?`, in.DesiredOutcome, in.CompletionEvidence, string(in.Disposition), in.Authority, trusted, string(status), in.ItemID); err != nil {
		return ClarifyResult{}, mapBacklogEngineError("clarify gtd item", err)
	}
	if err := bumpGTDRevision(ctx, tx); err != nil {
		return ClarifyResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ClarifyResult{}, err
	}
	item, err := loadGTDItem(ctx, db, in.ItemID)
	return ClarifyResult{Item: item, Publishable: in.Disposition == DispositionAction && len(holds) == 0, HoldReasons: holds}, err
}
