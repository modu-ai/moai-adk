package kanban

import (
	"context"
	"errors"
	"strings"
)

type GTDClass string

const (
	ClassAction    GTDClass = "action"
	ClassProject   GTDClass = "project"
	ClassReference GTDClass = "reference"
	ClassWaiting   GTDClass = "waiting"
	ClassScheduled GTDClass = "scheduled"
	ClassSomeday   GTDClass = "someday"
)

type OrganizeInput struct {
	ItemID   string
	Class    GTDClass
	Context  string
	ReviewAt string
}

func OrganizeGTDItem(ctx context.Context, store *BacklogStore, in OrganizeInput) (GTDItem, error) {
	return OrganizeGTDItemWithRelations(ctx, store, in, nil)
}

func knownGTDClass(class GTDClass) bool {
	switch class {
	case ClassAction, ClassProject, ClassReference, ClassWaiting, ClassScheduled, ClassSomeday:
		return true
	default:
		return false
	}
}

// OrganizeGTDItemWithRelations commits the classification and every requested
// relationship as one logical revision. Validation completes before any write.
func OrganizeGTDItemWithRelations(ctx context.Context, store *BacklogStore, in OrganizeInput, requested []GTDRelation) (GTDItem, error) {
	if strings.TrimSpace(in.ItemID) == "" {
		return GTDItem{}, errors.New("gtd organize: incomplete_input")
	}
	if !knownGTDClass(in.Class) {
		return GTDItem{}, errors.New("gtd organize: invalid_class")
	}
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
	item, err := loadGTDItem(ctx, tx, in.ItemID)
	if err != nil {
		return GTDItem{}, err
	}
	if item.Status != GTDStatusClarified {
		return GTDItem{}, errors.New("gtd organize: clarification_required")
	}
	persisted, targets, err := loadGTDRelationValidationState(ctx, tx)
	if err != nil {
		return GTDItem{}, err
	}
	for i := range requested {
		relation := &requested[i]
		if relation.SubjectID == "" {
			relation.SubjectID = in.ItemID
		}
		if relation.SourceRevision == 0 {
			relation.SourceRevision = item.SourceRevision + 1
		}
		if err := ValidateGTDRelation(*relation, persisted, targets); err != nil {
			return GTDItem{}, err
		}
		persisted = append(persisted, *relation)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE gtd_items SET class=?,action_context=?,review_at=?,status=?,source_revision=source_revision+1 WHERE item_id=?`, string(in.Class), in.Context, in.ReviewAt, string(GTDStatusOrganized), in.ItemID); err != nil {
		return GTDItem{}, mapBacklogEngineError("organize gtd item", err)
	}
	for _, relation := range requested {
		note := relation.Note
		if note == nil {
			note = []byte{}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gtd_relations(subject_id,object_id,kind,note,source,assertion_status,source_revision,policy_version) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(subject_id,object_id,kind) DO UPDATE SET note=excluded.note,source=excluded.source,assertion_status=excluded.assertion_status,source_revision=excluded.source_revision,policy_version=excluded.policy_version`, relation.SubjectID, relation.ObjectID, string(relation.Kind), note, relation.Source, relation.AssertionStatus, relation.SourceRevision, relation.PolicyVersion); err != nil {
			return GTDItem{}, mapBacklogEngineError("organize gtd relation", err)
		}
	}
	if err := bumpGTDRevision(ctx, tx); err != nil {
		return GTDItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return GTDItem{}, err
	}
	return loadGTDItem(ctx, db, in.ItemID)
}
