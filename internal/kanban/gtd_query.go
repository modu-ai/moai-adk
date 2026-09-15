package kanban

import (
	"context"
	"database/sql"
	"strconv"
)

type GTDProjectionRelation struct{ From, To, Kind, Sensitivity string }

func GTDProjectionSource(ctx context.Context, store *BacklogStore) (int64, []GTDProjectionRelation, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var raw string
	if err := tx.QueryRowContext(ctx, `SELECT value FROM gtd_meta WHERE key='logical_revision'`).Scan(&raw); err != nil {
		return 0, nil, err
	}
	revision, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT r.subject_id,r.object_id,r.kind,i.sensitivity FROM gtd_relations r JOIN gtd_items i ON i.item_id=r.subject_id ORDER BY r.subject_id,r.object_id,r.kind`)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []GTDProjectionRelation
	for rows.Next() {
		var r GTDProjectionRelation
		if err := rows.Scan(&r.From, &r.To, &r.Kind, &r.Sensitivity); err != nil {
			return 0, nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}
	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}
	return revision, out, nil
}

type GTDReflection struct {
	Items  []GTDItem     `json:"items"`
	Result ReflectResult `json:"result"`
}

type gtdQuerier interface {
	gtdQueryRower
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func listGTDItems(ctx context.Context, db gtdQuerier) ([]GTDItem, error) {
	rows, err := db.QueryContext(ctx, `SELECT item_id FROM gtd_items ORDER BY item_id`)
	if err != nil {
		return nil, mapBacklogEngineError("list gtd items", err)
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	items := make([]GTDItem, 0, len(ids))
	for _, id := range ids {
		item, err := loadGTDItem(ctx, db, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func listGTDRelations(ctx context.Context, db gtdQuerier) ([]GTDRelation, error) {
	rows, err := db.QueryContext(ctx, `SELECT subject_id,object_id,kind,note,source,assertion_status,source_revision,policy_version FROM gtd_relations ORDER BY subject_id,object_id,kind`)
	if err != nil {
		return nil, mapBacklogEngineError("list gtd relations", err)
	}
	defer func() { _ = rows.Close() }()
	var relations []GTDRelation
	for rows.Next() {
		var r GTDRelation
		var kind string
		if err := rows.Scan(&r.SubjectID, &r.ObjectID, &kind, &r.Note, &r.Source, &r.AssertionStatus, &r.SourceRevision, &r.PolicyVersion); err != nil {
			return nil, err
		}
		r.Kind = GTDRelationKind(kind)
		relations = append(relations, r)
	}
	return relations, rows.Err()
}

func ReflectGTDStore(ctx context.Context, store *BacklogStore) (GTDReflection, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return GTDReflection{}, err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return GTDReflection{}, err
	}
	defer func() { _ = tx.Rollback() }()
	items, err := listGTDItems(ctx, tx)
	if err != nil {
		return GTDReflection{}, err
	}
	relations, err := listGTDRelations(ctx, tx)
	if err != nil {
		return GTDReflection{}, err
	}
	reflectItems := make([]ReflectItem, 0, len(items))
	for _, item := range items {
		completed := false
		if item.CardID != "" {
			var archived int
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM archived_items WHERE id=?)`, item.CardID).Scan(&archived); err != nil {
				return GTDReflection{}, err
			}
			completed = archived == 1
		}
		reflectItems = append(reflectItems, ReflectItem{ItemID: item.ItemID, Class: item.Class, Cancelled: item.Cancelled, Completed: completed, EvidenceRevision: item.SourceRevision, CurrentEvidenceRevision: item.SourceRevision})
	}
	for _, r := range relations {
		for i := range reflectItems {
			if reflectItems[i].ItemID == r.SubjectID && r.SourceRevision != reflectItems[i].CurrentEvidenceRevision {
				reflectItems[i].EvidenceRevision = r.SourceRevision
			}
		}
		if r.Kind == RelationDependsOn {
			for i := range reflectItems {
				if reflectItems[i].ItemID == r.SubjectID {
					reflectItems[i].DependsOn = append(reflectItems[i].DependsOn, r.ObjectID)
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return GTDReflection{}, err
	}
	return GTDReflection{Items: items, Result: ReflectGTDState(ReflectInput{Items: reflectItems, Relations: relations})}, nil
}
