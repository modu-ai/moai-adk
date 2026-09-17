package kanban

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"
)

type EngageInput struct {
	ItemID             string
	DryRun             bool
	Authorized         bool
	EvidenceFresh      bool
	DependenciesReady  bool
	LaneAvailable      bool
	ResourcesAvailable bool
}

type EngageResult struct {
	Actionable bool     `json:"actionable"`
	CardID     string   `json:"card_id"`
	Reasons    []string `json:"reasons"`
}

func gtdCardUUID(itemID string) string {
	sum := sha256.Sum256([]byte("moai:gtd-card:" + itemID))
	var identity uuid.UUID
	copy(identity[:], sum[:16])
	identity[6] = (identity[6] & 0x0f) | 0x70
	identity[8] = (identity[8] & 0x3f) | 0x80
	return identity.String()
}

func findCardByUUID(store *BacklogStore, identity string) (*BacklogItem, error) {
	record, err := store.LoadPure()
	if err != nil {
		return nil, err
	}
	for i := range record.Items {
		if record.Items[i].CardUUID != nil && *record.Items[i].CardUUID == identity {
			card := record.Items[i]
			return &card, nil
		}
	}
	for i := range record.Archived {
		card := record.Archived[i].Item
		if card.CardUUID != nil && *card.CardUUID == identity {
			return &card, nil
		}
	}
	return nil, nil
}

func linkGTDCard(ctx context.Context, store *BacklogStore, itemID, cardID string) (string, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return "", err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `UPDATE gtd_items SET card_id=?,status=?,source_revision=source_revision+1 WHERE item_id=? AND card_id IS NULL`, cardID, string(GTDStatusPublished), itemID)
	if err != nil {
		return "", mapBacklogEngineError("publish gtd item", err)
	}
	n, _ := res.RowsAffected()
	if n == 1 {
		if err := bumpGTDRevision(ctx, tx); err != nil {
			return "", err
		}
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return cardID, nil
	}
	_ = tx.Rollback()
	item, err := loadGTDItem(ctx, db, itemID)
	if err != nil {
		return "", err
	}
	if item.CardID == cardID {
		return cardID, nil
	}
	return "", errors.New("gtd engage: concurrent_publish")
}

func EngageGTDItem(ctx context.Context, store *BacklogStore, in EngageInput) (EngageResult, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return EngageResult{}, err
	}
	item, err := loadGTDItem(ctx, db, in.ItemID)
	_ = db.Close()
	if err != nil {
		return EngageResult{}, err
	}
	if item.CardID != "" {
		return EngageResult{Actionable: true, CardID: item.CardID}, nil
	}
	var reasons []string
	if item.Status != GTDStatusOrganized || item.Disposition != DispositionAction {
		reasons = append(reasons, "not_organized_action")
	}
	if !in.Authorized {
		reasons = append(reasons, "not_authorized")
	}
	if !in.EvidenceFresh {
		reasons = append(reasons, "stale_evidence")
	}
	if !in.DependenciesReady {
		reasons = append(reasons, "dependencies_blocked")
	}
	if !in.LaneAvailable {
		reasons = append(reasons, "lane_unavailable")
	}
	if !in.ResourcesAvailable {
		reasons = append(reasons, "resource_limit")
	}
	if len(reasons) > 0 {
		return EngageResult{Reasons: reasons}, nil
	}
	if in.DryRun {
		return EngageResult{Actionable: true}, nil
	}
	identity := gtdCardUUID(item.ItemID)
	card, err := findCardByUUID(store, identity)
	if err != nil {
		return EngageResult{}, err
	}
	if card == nil {
		card, _, err = store.addWithCardUUID(item.Content, &identity)
		if err != nil {
			// A concurrent publisher may have committed the deterministic
			// identity before its GTD link. Read back before deciding retry.
			card, _ = findCardByUUID(store, identity)
			if card == nil {
				return EngageResult{}, err
			}
		}
	}
	cardID, err := linkGTDCard(ctx, store, item.ItemID, card.ID)
	if err != nil {
		return EngageResult{}, err
	}
	return EngageResult{Actionable: true, CardID: cardID}, nil
}
