package kanban

import (
	"context"
	"database/sql"
	"errors"
)

type GTDRelationKind string

const (
	RelationDependsOn   GTDRelationKind = "depends_on"
	RelationPartOf      GTDRelationKind = "part_of"
	RelationSupportedBy GTDRelationKind = "supported_by"
	RelationRelatedTo   GTDRelationKind = "related_to"
	RelationSupersedes  GTDRelationKind = "supersedes"
	RelationContains    GTDRelationKind = "contains"
	RelationAbsorbs     GTDRelationKind = "absorbs"
	RelationReplaces    GTDRelationKind = "replaces"
	RelationConflicts   GTDRelationKind = "conflicts"
)

type GTDRelation struct {
	SubjectID       string
	ObjectID        string
	Kind            GTDRelationKind
	Note            []byte
	Source          string
	AssertionStatus string
	SourceRevision  int64
	PolicyVersion   string
}

func knownGTDRelationKind(kind GTDRelationKind) bool {
	switch kind {
	case RelationDependsOn, RelationPartOf, RelationSupportedBy, RelationRelatedTo, RelationSupersedes,
		RelationContains, RelationAbsorbs, RelationReplaces, RelationConflicts:
		return true
	default:
		return false
	}
}

// ValidateGTDRelation validates new assertions without translating or
// normalizing the four legacy relation values. depends_on is directed from
// successor to prerequisite; adding an edge that can reach its subject is a
// cycle and is rejected.
func ValidateGTDRelation(candidate GTDRelation, existing []GTDRelation, targets map[string]bool) error {
	if candidate.SubjectID == "" || candidate.ObjectID == "" || candidate.SubjectID == candidate.ObjectID || candidate.SourceRevision < 0 || !knownGTDRelationKind(candidate.Kind) {
		return errors.New("gtd relation: invalid_relation")
	}
	if !targets[candidate.SubjectID] || !targets[candidate.ObjectID] {
		return errors.New("gtd relation: missing_target")
	}
	if candidate.Kind != RelationDependsOn {
		return nil
	}
	graph := map[string][]string{}
	for _, relation := range existing {
		if relation.Kind == RelationDependsOn {
			graph[relation.SubjectID] = append(graph[relation.SubjectID], relation.ObjectID)
		}
	}
	graph[candidate.SubjectID] = append(graph[candidate.SubjectID], candidate.ObjectID)
	seen := map[string]bool{}
	var reaches func(string) bool
	reaches = func(node string) bool {
		if node == candidate.SubjectID {
			return true
		}
		if seen[node] {
			return false
		}
		seen[node] = true
		for _, next := range graph[node] {
			if reaches(next) {
				return true
			}
		}
		return false
	}
	if reaches(candidate.ObjectID) {
		return errors.New("gtd relation: dependency_cycle")
	}
	return nil
}

func GTDRelationBlocks(kind GTDRelationKind, targetCompleted bool) bool {
	return kind == RelationDependsOn && !targetCompleted
}

func loadGTDRelationValidationState(ctx context.Context, tx *sql.Tx) ([]GTDRelation, map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT subject_id,object_id,kind,note,source,assertion_status,source_revision,policy_version FROM gtd_relations`)
	if err != nil {
		return nil, nil, err
	}
	var persisted []GTDRelation
	for rows.Next() {
		var r GTDRelation
		var kind string
		if err := rows.Scan(&r.SubjectID, &r.ObjectID, &kind, &r.Note, &r.Source, &r.AssertionStatus, &r.SourceRevision, &r.PolicyVersion); err != nil {
			_ = rows.Close()
			return nil, nil, err
		}
		r.Kind = GTDRelationKind(kind)
		persisted = append(persisted, r)
	}
	if err := rows.Close(); err != nil {
		return nil, nil, err
	}
	targets := map[string]bool{}
	itemRows, err := tx.QueryContext(ctx, `SELECT item_id FROM gtd_items`)
	if err != nil {
		return nil, nil, err
	}
	for itemRows.Next() {
		var id string
		if err := itemRows.Scan(&id); err != nil {
			_ = itemRows.Close()
			return nil, nil, err
		}
		targets[id] = true
	}
	if err := itemRows.Close(); err != nil {
		return nil, nil, err
	}
	return persisted, targets, nil
}

func PutGTDRelation(ctx context.Context, store *BacklogStore, relation GTDRelation, existing []GTDRelation, targets map[string]bool) error {
	_ = existing
	_ = targets
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
	persisted, targetSet, err := loadGTDRelationValidationState(ctx, tx)
	if err != nil {
		return err
	}
	if err := ValidateGTDRelation(relation, persisted, targetSet); err != nil {
		return err
	}
	note := relation.Note
	if note == nil {
		note = []byte{}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO gtd_relations(subject_id,object_id,kind,note,source,assertion_status,source_revision,policy_version) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(subject_id,object_id,kind) DO UPDATE SET note=excluded.note,source=excluded.source,assertion_status=excluded.assertion_status,source_revision=excluded.source_revision,policy_version=excluded.policy_version`, relation.SubjectID, relation.ObjectID, string(relation.Kind), note, relation.Source, relation.AssertionStatus, relation.SourceRevision, relation.PolicyVersion)
	if err != nil {
		return mapBacklogEngineError("put gtd relation", err)
	}
	if err := bumpGTDRevision(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}
