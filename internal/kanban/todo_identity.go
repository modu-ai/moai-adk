package kanban

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	identitySchemaVersion = "1"
	identityMetaKey       = "identity_schema_version"
	identityProjectKind   = "project"
	identityCardKind      = "card"
	identityProjectLocal  = "project"
)

const identityDDL = `CREATE TABLE todo_identities(
  entity_kind TEXT NOT NULL,
  local_id TEXT NOT NULL,
  uuid TEXT NOT NULL UNIQUE,
  PRIMARY KEY(entity_kind,local_id)
)`

type todoIdentitySnapshot struct {
	project *string
	cards   map[string]string
}

type identityColumn struct {
	name    string
	notNull int
	pk      int
}

// identitySchemaPresent distinguishes a genuinely legacy database (neither
// table nor stamp) from every partial/future shape, which fails closed.
func (e *backlogEngine) identitySchemaPresent(ctx context.Context) (bool, error) {
	var tables int
	if err := e.queryDB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='todo_identities'`).Scan(&tables); err != nil {
		return false, err
	}
	var version string
	err := e.queryDB().QueryRowContext(ctx, `SELECT value FROM meta WHERE key=?`, identityMetaKey).Scan(&version)
	stamp := true
	if errors.Is(err, sql.ErrNoRows) {
		stamp = false
	} else if err != nil {
		return false, err
	}
	if tables == 0 && !stamp {
		return false, nil
	}
	if tables != 1 || !stamp {
		return false, fmt.Errorf("partial identity schema: table=%d stamp=%t: %w", tables, stamp, ErrBacklogCorrupt)
	}
	if version != identitySchemaVersion {
		return false, fmt.Errorf("unsupported identity_schema_version %q: %w", version, ErrBacklogCorrupt)
	}
	if err := e.validateIdentityTable(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (e *backlogEngine) validateIdentityTable(ctx context.Context) error {
	rows, err := e.queryDB().QueryContext(ctx, `PRAGMA table_info(todo_identities)`)
	if err != nil {
		return err
	}
	var columns []identityColumn
	for rows.Next() {
		var cid, notNull, pk int
		var name, typ string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		_ = cid
		if strings.ToUpper(typ) != "TEXT" {
			_ = rows.Close()
			return fmt.Errorf("identity column %s type %q: %w", name, typ, ErrBacklogCorrupt)
		}
		columns = append(columns, identityColumn{name: name, notNull: notNull, pk: pk})
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	want := []identityColumn{{"entity_kind", 1, 1}, {"local_id", 1, 2}, {"uuid", 1, 0}}
	if len(columns) != len(want) {
		return fmt.Errorf("identity column count %d: %w", len(columns), ErrBacklogCorrupt)
	}
	for i := range want {
		if columns[i] != want[i] {
			return fmt.Errorf("identity column %d=%+v want %+v: %w", i, columns[i], want[i], ErrBacklogCorrupt)
		}
	}
	uniqueUUID, err := e.identityUUIDUnique(ctx)
	if err != nil {
		return err
	}
	if !uniqueUUID {
		return fmt.Errorf("identity uuid lacks a single-column UNIQUE constraint: %w", ErrBacklogCorrupt)
	}
	return nil
}

func (e *backlogEngine) identityUUIDUnique(ctx context.Context) (bool, error) {
	rows, err := e.queryDB().QueryContext(ctx, `PRAGMA index_list(todo_identities)`)
	if err != nil {
		return false, err
	}
	var indexes []string
	for rows.Next() {
		var seq, unique, partial int
		var name, origin string
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			_ = rows.Close()
			return false, err
		}
		if unique == 1 && partial == 0 {
			indexes = append(indexes, name)
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return false, err
	}
	for _, index := range indexes {
		indexRows, err := e.queryDB().QueryContext(ctx, `SELECT name FROM pragma_index_info(?) ORDER BY seqno`, index)
		if err != nil {
			return false, err
		}
		var names []string
		for indexRows.Next() {
			var name string
			if err := indexRows.Scan(&name); err != nil {
				_ = indexRows.Close()
				return false, err
			}
			names = append(names, name)
		}
		err = indexRows.Err()
		_ = indexRows.Close()
		if err != nil {
			return false, err
		}
		if len(names) == 1 && names[0] == "uuid" {
			return true, nil
		}
	}
	return false, nil
}

func validIdentityUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.Version() == 7 && parsed.String() == value && strings.ToLower(value) == value
}

func (e *backlogEngine) readIdentitySnapshot(ctx context.Context) (todoIdentitySnapshot, error) {
	snapshot := todoIdentitySnapshot{cards: map[string]string{}}
	present, err := e.identitySchemaPresent(ctx)
	if err != nil || !present {
		return snapshot, err
	}
	rows, err := e.queryDB().QueryContext(ctx, `SELECT entity_kind,local_id,uuid FROM todo_identities ORDER BY entity_kind,local_id`)
	if err != nil {
		return snapshot, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var kind, localID, value string
		if err := rows.Scan(&kind, &localID, &value); err != nil {
			return snapshot, err
		}
		if !validIdentityUUID(value) {
			return snapshot, fmt.Errorf("malformed %s identity %q: %w", kind, value, ErrBacklogCorrupt)
		}
		switch kind {
		case identityProjectKind:
			if localID != identityProjectLocal || snapshot.project != nil {
				return snapshot, fmt.Errorf("invalid project identity local_id %q: %w", localID, ErrBacklogCorrupt)
			}
			v := value
			snapshot.project = &v
		case identityCardKind:
			if localID == "" {
				return snapshot, fmt.Errorf("empty card identity local_id: %w", ErrBacklogCorrupt)
			}
			snapshot.cards[localID] = value
		default:
			return snapshot, fmt.Errorf("unknown identity kind %q: %w", kind, ErrBacklogCorrupt)
		}
	}
	return snapshot, rows.Err()
}

func (snapshot todoIdentitySnapshot) apply(rec *BacklogRecord) {
	rec.ProjectUUID = cloneIdentity(snapshot.project)
	for i := range rec.Items {
		rec.Items[i].CardUUID = identityCard(snapshot.cards, rec.Items[i].ID)
	}
	for i := range rec.Archived {
		rec.Archived[i].Item.CardUUID = identityCard(snapshot.cards, rec.Archived[i].Item.ID)
	}
	for i := range rec.Runtime.Runs {
		rec.Runtime.Runs[i].ProjectUUID = cloneIdentity(snapshot.project)
	}
	for i := range rec.Runtime.Assignments {
		rec.Runtime.Assignments[i].ProjectUUID = cloneIdentity(snapshot.project)
		rec.Runtime.Assignments[i].CardUUID = identityCard(snapshot.cards, rec.Runtime.Assignments[i].CardID)
	}
}

func cloneIdentity(value *string) *string {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func identityCard(cards map[string]string, localID string) *string {
	value, ok := cards[localID]
	if !ok {
		return nil
	}
	v := value
	return &v
}

func ensureIdentitySchema(ctx context.Context, tx *sql.Tx, dbPath string) error {
	snapshot := &backlogEngine{dbPath: dbPath, reader: tx}
	present, err := snapshot.identitySchemaPresent(ctx)
	if err != nil || present {
		return err
	}
	if _, err := tx.ExecContext(ctx, identityDDL); err != nil {
		return mapBacklogEngineError("create identity schema", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO meta(key,value) VALUES(?,?)`, identityMetaKey, identitySchemaVersion); err != nil {
		return mapBacklogEngineError("stamp identity schema", err)
	}
	return nil
}

func ensureIdentity(ctx context.Context, tx *sql.Tx, kind, localID string, preferred *string) (string, error) {
	if preferred != nil && !validIdentityUUID(*preferred) {
		return "", fmt.Errorf("malformed preferred %s identity %q: %w", kind, *preferred, ErrBacklogCorrupt)
	}
	var existing string
	err := tx.QueryRowContext(ctx, `SELECT uuid FROM todo_identities WHERE entity_kind=? AND local_id=?`, kind, localID).Scan(&existing)
	if err == nil {
		if !validIdentityUUID(existing) {
			return "", fmt.Errorf("malformed %s identity %q: %w", kind, existing, ErrBacklogCorrupt)
		}
		if preferred != nil && existing != *preferred {
			return "", fmt.Errorf("%s identity mismatch for %q: stored %q, preferred %q: %w", kind, localID, existing, *preferred, ErrBacklogCorrupt)
		}
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	var value string
	if preferred != nil {
		value = *preferred
	} else {
		issued, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("issue %s UUIDv7: %w", kind, err)
		}
		value = issued.String()
		if !validIdentityUUID(value) {
			return "", fmt.Errorf("issued invalid %s UUIDv7: %w", kind, ErrBacklogCorrupt)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES(?,?,?)`, kind, localID, value); err != nil {
		return "", mapBacklogEngineError("insert identity", err)
	}
	return value, nil
}

// ensureRecordIdentities issues/backfills every identity in the same
// transaction that writes the whole record.
func ensureRecordIdentities(ctx context.Context, tx *sql.Tx, dbPath string, rec *BacklogRecord) error {
	if err := ensureIdentitySchema(ctx, tx, dbPath); err != nil {
		return err
	}
	project, err := ensureIdentity(ctx, tx, identityProjectKind, identityProjectLocal, rec.ProjectUUID)
	if err != nil {
		return err
	}
	ids := make(map[string]*string, len(rec.Items)+len(rec.Archived))
	for _, item := range rec.Items {
		if _, duplicate := ids[item.ID]; duplicate {
			return fmt.Errorf("duplicate card identity %q: %w", item.ID, ErrBacklogIDConflict)
		}
		ids[item.ID] = item.CardUUID
	}
	for _, entry := range rec.Archived {
		if _, duplicate := ids[entry.Item.ID]; duplicate {
			return fmt.Errorf("duplicate card identity %q: %w", entry.Item.ID, ErrBacklogIDConflict)
		}
		ids[entry.Item.ID] = entry.Item.CardUUID
	}
	cards := make(map[string]string, len(ids))
	for id, preferred := range ids {
		value, err := ensureIdentity(ctx, tx, identityCardKind, id, preferred)
		if err != nil {
			return err
		}
		cards[id] = value
	}
	projectValue := project
	todoIdentitySnapshot{project: &projectValue, cards: cards}.apply(rec)
	return nil
}

func ensureStoredIdentities(ctx context.Context, tx *sql.Tx, dbPath string) error {
	if err := ensureIdentitySchema(ctx, tx, dbPath); err != nil {
		return err
	}
	if _, err := ensureIdentity(ctx, tx, identityProjectKind, identityProjectLocal, nil); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM items UNION ALL SELECT id FROM archived_items ORDER BY id`)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	for i, id := range ids {
		if i > 0 && ids[i-1] == id {
			return fmt.Errorf("runtime card %q is ambiguous", id)
		}
		if _, err := ensureIdentity(ctx, tx, identityCardKind, id, nil); err != nil {
			return err
		}
	}
	return nil
}
