package kanban

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

const identityTestDDL = `CREATE TABLE todo_identities(entity_kind TEXT NOT NULL,local_id TEXT NOT NULL,uuid TEXT NOT NULL UNIQUE,PRIMARY KEY(entity_kind,local_id)); INSERT INTO meta(key,value) VALUES('identity_schema_version','1');`

func identityJSON(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func identityUUIDv7(t *testing.T, m map[string]json.RawMessage, key, at string) string {
	t.Helper()
	raw, ok := m[key]
	if !ok {
		t.Errorf("%s: missing %s after successful writer", at, key)
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Errorf("%s: %s is not a string: %s", at, key, raw)
		return ""
	}
	u, err := uuid.Parse(s)
	if err != nil || u == uuid.Nil || u.String() != s || strings.ToLower(s) != s || u.Version() != 7 {
		t.Errorf("%s: %s=%q, want nonzero canonical lowercase RFC UUIDv7", at, key, s)
		return ""
	}
	return s
}

func identityNull(t *testing.T, m map[string]json.RawMessage, key, at string) {
	t.Helper()
	raw, ok := m[key]
	if !ok {
		t.Errorf("%s: %s key absent, want key-present literal null", at, key)
		return
	}
	if string(raw) != "null" {
		t.Errorf("%s: %s=%s, want literal null", at, key, raw)
	}
}

func identityDB(t *testing.T, s *BacklogStore) *sql.DB {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, backlogDSN(s.EnginePath()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func identityRemove(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`DROP TABLE IF EXISTS todo_identities; DELETE FROM meta WHERE key='identity_schema_version'`); err != nil {
		t.Fatal(err)
	}
}

func identityCreate(t *testing.T, db *sql.DB, version string) {
	t.Helper()
	if _, err := db.Exec(identityTestDDL); err != nil {
		t.Fatal(err)
	}
	if version != "1" {
		if _, err := db.Exec(`UPDATE meta SET value=? WHERE key='identity_schema_version'`, version); err != nil {
			t.Fatal(err)
		}
	}
}

func identitySchema(t *testing.T, db *sql.DB) string {
	t.Helper()
	rows, err := db.Query(`SELECT type,name,coalesce(sql,'') FROM sqlite_master ORDER BY type,name`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var a, n, d string
		if err := rows.Scan(&a, &n, &d); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&b, "%s|%s|%s\n", a, n, d)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func identityPersistentState(t *testing.T, db *sql.DB) string {
	t.Helper()
	var b strings.Builder
	for _, query := range []string{
		`SELECT key,value FROM meta ORDER BY key`,
		`SELECT CAST(seq AS TEXT),id,text,added_at,coalesce(spec_id,'<null>'),state FROM items ORDER BY seq`,
		`SELECT entity_kind,local_id,uuid FROM todo_identities ORDER BY entity_kind,local_id`,
	} {
		rows, err := db.Query(query)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]string, len(columns))
			scan := make([]any, len(columns))
			for i := range values {
				scan[i] = &values[i]
			}
			if err := rows.Scan(scan...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "%q\n", values)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		_ = rows.Close()
	}
	return identitySchema(t, db) + b.String()
}

func identityBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func identityLegacySQLite(t *testing.T, archive bool) (*BacklogStore, *sql.DB) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	s := NewBacklogStore(BacklogPathForRoot(t.TempDir()))
	card, _, err := s.Add("legacy")
	if err != nil {
		t.Fatal(err)
	}
	if archive {
		if err := s.Mutate(func(r *BacklogRecord) error { return r.ArchiveCard(card.ID) }); err != nil {
			t.Fatal(err)
		}
		if _, _, err := s.Add("live"); err != nil {
			t.Fatal(err)
		}
	}
	db := identityDB(t, s)
	identityRemove(t, db)
	return s, db
}

func TestTodoIdentityAC001GitAndNonGitUUIDv7(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	executed := 0
	var projectUUIDs, cardUUIDs []string
	for _, tc := range []struct {
		name string
		git  bool
	}{{"git", true}, {"non-git", false}} {
		t.Run(tc.name, func(t *testing.T) {
			executed++
			root := t.TempDir()
			if tc.git {
				if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
					t.Fatalf("git init: %v: %s", err, out)
				}
			}
			s := NewBacklogStore(BacklogPathForRoot(root))
			a, _, err := s.Add("a")
			if err != nil {
				t.Fatal(err)
			}
			b, _, err := s.Add("b")
			if err != nil {
				t.Fatal(err)
			}
			r, err := s.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			if a.ID != "t1" || b.ID != "t2" || r.LastSeq != 2 || len(r.Items) != 2 {
				t.Fatalf("tNN/last_seq regression: a=%+v b=%+v record=%+v", a, b, r)
			}
			project := identityUUIDv7(t, identityJSON(t, r), "project_uuid", tc.name+" project")
			ar := identityUUIDv7(t, identityJSON(t, a), "card_uuid", tc.name+" first Add")
			br := identityUUIDv7(t, identityJSON(t, b), "card_uuid", tc.name+" second Add")
			al := identityUUIDv7(t, identityJSON(t, r.Items[0]), "card_uuid", tc.name+" first LoadPure")
			bl := identityUUIDv7(t, identityJSON(t, r.Items[1]), "card_uuid", tc.name+" second LoadPure")
			if ar != "" && (ar != al || br != bl || al == bl) {
				t.Errorf("Add equality/uniqueness failed: add=%q,%q load=%q,%q", ar, br, al, bl)
			}
			if project != "" {
				projectUUIDs = append(projectUUIDs, project)
			}
			for _, cardUUID := range []string{al, bl} {
				if cardUUID != "" {
					cardUUIDs = append(cardUUIDs, cardUUID)
				}
			}
		})
	}
	all := append(append([]string{}, projectUUIDs...), cardUUIDs...)
	unique := make(map[string]struct{}, len(all))
	for _, value := range all {
		unique[value] = struct{}{}
	}
	if len(projectUUIDs) != 2 || len(cardUUIDs) != 4 || len(unique) != 6 {
		t.Errorf("cross-project UUID cardinality: projects=%d cards=%d globally_unique=%d, want 2/4/6", len(projectUUIDs), len(cardUUIDs), len(unique))
	}
	t.Logf("AC-TID-001 executed fixtures=%d git=1 non_git=1 cards=4", executed)
}

func TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backlog.json")
	raw := []byte(`{"version":1,"last_seq":2,"items":[{"id":"t1","text":"live","added_at":"2026-09-12T00:00:00Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[{"item":{"id":"t2","text":"archive","added_at":"2026-09-12T00:00:01Z","spec_id":null,"state":"picked"},"position":1,"findings":[]}]}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	identityNull(t, identityJSON(t, r), "project_uuid", "JSON project")
	identityNull(t, identityJSON(t, r.Items[0]), "card_uuid", "JSON live")
	identityNull(t, identityJSON(t, r.Archived[0].Item), "card_uuid", "JSON archive")
	if !bytes.Equal(raw, identityBytes(t, path)) {
		t.Error("legacy JSON LoadPure changed bytes")
	}
	if _, err := os.Stat(backlogSQLitePath(path)); !os.IsNotExist(err) {
		t.Errorf("legacy JSON LoadPure changed DB existence: %v", err)
	}
	t.Log("AC-TID-002 executed JSON=1 live=1 archive=1")
}

func TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation(t *testing.T) {
	s, db := identityLegacySQLite(t, true)
	before := identityBytes(t, s.EnginePath())
	schema := identitySchema(t, db)
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	identityNull(t, identityJSON(t, r), "project_uuid", "SQLite project")
	identityNull(t, identityJSON(t, r.Items[0]), "card_uuid", "SQLite live")
	identityNull(t, identityJSON(t, r.Archived[0].Item), "card_uuid", "SQLite archive")
	if !bytes.Equal(before, identityBytes(t, s.EnginePath())) {
		t.Error("SQLite LoadPure changed main DB bytes")
	}
	if schema != identitySchema(t, db) {
		t.Error("SQLite LoadPure changed schema")
	}
	t.Log("AC-TID-002 executed SQLite=1 live=1 archive=1")
}

func TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation(t *testing.T) {
	s, _ := identityLegacySQLite(t, true)
	deferred := openDeferringConn(t, s.EnginePath())
	if _, err := deferred.Exec(`INSERT INTO meta(key,value) VALUES('identity_wal_probe','committed')`); err != nil {
		t.Fatal(err)
	}
	wal := s.EnginePath() + "-wal"
	if info, err := os.Stat(wal); err != nil || info.Size() == 0 {
		t.Fatalf("GAP: active WAL fixture not reached: %v", err)
	}
	dbBefore, walBefore := identityBytes(t, s.EnginePath()), identityBytes(t, wal)
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	identityNull(t, identityJSON(t, r), "project_uuid", "WAL project")
	identityNull(t, identityJSON(t, r.Items[0]), "card_uuid", "WAL live")
	identityNull(t, identityJSON(t, r.Archived[0].Item), "card_uuid", "WAL archive")
	if !bytes.Equal(dbBefore, identityBytes(t, s.EnginePath())) || !bytes.Equal(walBefore, identityBytes(t, wal)) {
		t.Error("active-WAL LoadPure changed DB/WAL bytes")
	}
	t.Log("AC-TID-002 executed active_WAL=1 committed_snapshot=1 live=1 archive=1")
}

type identityFailReader struct{ reached int }

func (r *identityFailReader) Read([]byte) (int, error) { r.reached++; return 0, io.ErrUnexpectedEOF }

func TestTodoIdentityAC003UUIDEntropyFailureRollsBack(t *testing.T) {
	s, _ := identityLegacySQLite(t, false)
	before, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	beforeRaw, _ := json.Marshal(before)
	fault := &identityFailReader{}
	uuid.SetRand(fault)
	defer uuid.SetRand(nil) // selector has no parallel tests; always restore global seam.
	_, _, writeErr := s.Add("rollback")
	uuid.SetRand(nil)
	after, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	afterRaw, _ := json.Marshal(after)
	t.Logf("AC-TID-003 entropy fault reached=%d writer_error=%v", fault.reached, writeErr)
	if fault.reached != 1 {
		t.Errorf("UUIDv7 entropy seam reached=%d want=1", fault.reached)
	}
	if writeErr == nil {
		t.Error("UUIDv7 entropy failure ignored")
	}
	if !bytes.Equal(beforeRaw, afterRaw) {
		t.Error("entropy failure left partial change")
	}
	retried, _, retryErr := s.Add("successful retry")
	if retryErr != nil {
		t.Fatalf("retry after entropy fault: %v", retryErr)
	}
	retryRecord, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	project := identityUUIDv7(t, identityJSON(t, retryRecord), "project_uuid", "entropy retry project")
	card := identityUUIDv7(t, identityJSON(t, retried), "card_uuid", "entropy retry Add")
	if err := s.Mutate(func(r *BacklogRecord) error { r.Items[len(r.Items)-1].Text = "repeat writer"; return nil }); err != nil {
		t.Fatal(err)
	}
	repeated, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	repeatedProject := identityUUIDv7(t, identityJSON(t, repeated), "project_uuid", "entropy repeated writer project")
	repeatedCard := identityUUIDv7(t, identityJSON(t, repeated.Items[len(repeated.Items)-1]), "card_uuid", "entropy repeated writer card")
	if project != "" && (project != repeatedProject || card != repeatedCard) {
		t.Errorf("entropy retry identity changed on repeated writer: project=%q/%q card=%q/%q", project, repeatedProject, card, repeatedCard)
	}
	t.Log("AC-TID-003 entropy fault retry=1 repeated_writer=1")
}

func TestTodoIdentityAC003BackfillFaultRollsBack(t *testing.T) {
	s, db := identityLegacySQLite(t, true)
	identityCreate(t, db, "1")
	if _, err := db.Exec(`CREATE TRIGGER identity_fault BEFORE INSERT ON todo_identities BEGIN SELECT RAISE(ABORT,'identity-fault-reached'); END`); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	beforeRaw, _ := json.Marshal(before)
	_, _, writeErr := s.Add("backfill")
	reached := 0
	if writeErr != nil && strings.Contains(writeErr.Error(), "identity-fault-reached") {
		reached = 1
	}
	after, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	afterRaw, _ := json.Marshal(after)
	var rows int
	if err := db.QueryRow(`SELECT count(*) FROM todo_identities`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	t.Logf("AC-TID-003 backfill fault reached=%d writer_error=%v identity_rows=%d", reached, writeErr, rows)
	if reached != 1 || writeErr == nil {
		t.Errorf("writer missed backfill fault: reached=%d err=%v", reached, writeErr)
	}
	if !bytes.Equal(beforeRaw, afterRaw) || rows != 0 {
		t.Errorf("rollback failed: record_equal=%v identity_rows=%d", bytes.Equal(beforeRaw, afterRaw), rows)
	}
	if _, err := db.Exec(`DROP TRIGGER identity_fault`); err != nil {
		t.Fatal(err)
	}
	retried, _, retryErr := s.Add("backfill retry")
	if retryErr != nil {
		t.Fatalf("retry after backfill fault: %v", retryErr)
	}
	retryRecord, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	project := identityUUIDv7(t, identityJSON(t, retryRecord), "project_uuid", "backfill retry project")
	card := identityUUIDv7(t, identityJSON(t, retried), "card_uuid", "backfill retry Add")
	for i, item := range retryRecord.Items {
		identityUUIDv7(t, identityJSON(t, item), "card_uuid", fmt.Sprintf("backfilled live %d", i))
	}
	for i, entry := range retryRecord.Archived {
		identityUUIDv7(t, identityJSON(t, entry.Item), "card_uuid", fmt.Sprintf("backfilled archive %d", i))
	}
	if err := s.Mutate(func(r *BacklogRecord) error { r.Items[len(r.Items)-1].Text = "repeat writer"; return nil }); err != nil {
		t.Fatal(err)
	}
	repeated, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	repeatedProject := identityUUIDv7(t, identityJSON(t, repeated), "project_uuid", "backfill repeated project")
	repeatedCard := identityUUIDv7(t, identityJSON(t, repeated.Items[len(repeated.Items)-1]), "card_uuid", "backfill repeated card")
	if project != "" && (project != repeatedProject || card != repeatedCard) {
		t.Errorf("backfill retry identity changed on repeated writer")
	}
	t.Log("AC-TID-003 backfill fault retry=1 repeated_writer=1")
}

func TestTodoIdentityAC003SchemaCreationFaultRollsBack(t *testing.T) {
	s, db := identityLegacySQLite(t, false)
	if _, err := db.Exec(`CREATE TRIGGER identity_stamp_fault BEFORE INSERT ON meta WHEN NEW.key='identity_schema_version' BEGIN SELECT RAISE(ABORT,'identity-stamp-fault-reached'); END`); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	beforeRaw, _ := json.Marshal(before)
	_, _, writeErr := s.Add("schema creation fault")
	reached := 0
	if writeErr != nil && strings.Contains(writeErr.Error(), "identity-stamp-fault-reached") {
		reached = 1
	}
	after, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	afterRaw, _ := json.Marshal(after)
	var tableCount, stampCount, identityRows int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='todo_identities'`).Scan(&tableCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM meta WHERE key='identity_schema_version'`).Scan(&stampCount); err != nil {
		t.Fatal(err)
	}
	if tableCount != 0 {
		if err := db.QueryRow(`SELECT count(*) FROM todo_identities`).Scan(&identityRows); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("AC-TID-003 schema fault reached=%d writer_error=%v table=%d stamp=%d rows=%d", reached, writeErr, tableCount, stampCount, identityRows)
	if reached != 1 || writeErr == nil || tableCount != 0 || stampCount != 0 || identityRows != 0 || !bytes.Equal(beforeRaw, afterRaw) {
		t.Errorf("schema/stamp failure was not atomic: reached=%d err=%v table=%d stamp=%d rows=%d record_equal=%v", reached, writeErr, tableCount, stampCount, identityRows, bytes.Equal(beforeRaw, afterRaw))
	}
	if _, err := db.Exec(`DROP TRIGGER identity_stamp_fault`); err != nil {
		t.Fatal(err)
	}
	retried, _, retryErr := s.Add("schema creation retry")
	if retryErr != nil {
		t.Fatalf("retry after schema creation fault: %v", retryErr)
	}
	retryRecord, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	identityUUIDv7(t, identityJSON(t, retryRecord), "project_uuid", "schema retry project")
	identityUUIDv7(t, identityJSON(t, retried), "card_uuid", "schema retry card")
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='todo_identities'`).Scan(&tableCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM meta WHERE key='identity_schema_version'`).Scan(&stampCount); err != nil {
		t.Fatal(err)
	}
	if tableCount != 1 || stampCount != 1 {
		t.Errorf("successful retry did not create identity schema: table=%d stamp=%d", tableCount, stampCount)
	}
	var ddl, coreVersion, identityVersion string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='todo_identities'`).Scan(&ddl); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&coreVersion); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT value FROM meta WHERE key='identity_schema_version'`).Scan(&identityVersion); err != nil {
		t.Fatal(err)
	}
	runtimeVersion := "absent"
	if err := db.QueryRow(`SELECT value FROM meta WHERE key='runtime_schema_version'`).Scan(&runtimeVersion); err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	rows, err := db.Query(`PRAGMA index_list(todo_identities)`)
	if err != nil {
		t.Fatal(err)
	}
	var indexes []string
	for rows.Next() {
		var seq, unique, partial int
		var name, origin string
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		indexes = append(indexes, fmt.Sprintf("%s:%s:unique=%d", name, origin, unique))
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		t.Fatal(err)
	}
	_ = rows.Close()
	if ddl != identityDDL || coreVersion != "1" || identityVersion != "1" || (runtimeVersion != "absent" && runtimeVersion != "1") || len(indexes) != 2 {
		t.Errorf("identity schema readback ddl=%q core=%q runtime=%q identity=%q indexes=%v", ddl, coreVersion, runtimeVersion, identityVersion, indexes)
	}
	t.Logf("AC-TID-003 schema fault retry=1 ddl=%q meta=schema:%s/runtime:%s/identity:%s indexes=%v", ddl, coreVersion, runtimeVersion, identityVersion, indexes)
}

func TestTodoIdentityAC004LifecycleStability(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	s := NewBacklogStore(BacklogPathForRoot(t.TempDir()))
	a, _, err := s.Add("before")
	if err != nil {
		t.Fatal(err)
	}
	r0, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	p0 := identityUUIDv7(t, identityJSON(t, r0), "project_uuid", "initial project")
	c0 := identityUUIDv7(t, identityJSON(t, a), "card_uuid", "Add return")
	if err := s.Mutate(func(r *BacklogRecord) error {
		r.Items[0].Text = "after"
		r.Items[0].State = BacklogStatePicked
		spec := "SPEC-IDENTITY-001"
		r.Items[0].SpecID = &spec
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Mutate(func(r *BacklogRecord) error { return r.ArchiveCard(a.ID) }); err != nil {
		t.Fatal(err)
	}
	ra, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	ca := identityUUIDv7(t, identityJSON(t, ra.Archived[0].Item), "card_uuid", "archive")
	if err := s.Mutate(func(r *BacklogRecord) error { return r.RestoreCard(a.ID) }); err != nil {
		t.Fatal(err)
	}
	rr, err := NewBacklogStore(s.Path()).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	cr := identityUUIDv7(t, identityJSON(t, rr.Items[0]), "card_uuid", "restore")
	pr := identityUUIDv7(t, identityJSON(t, rr), "project_uuid", "reopen project")
	if rr.Items[0].Text != "after" || rr.Items[0].State != BacklogStatePicked || rr.Items[0].SpecID == nil {
		t.Fatalf("lifecycle fields regressed: %+v", rr.Items[0])
	}
	if c0 != "" && (c0 != ca || c0 != cr || p0 != pr) {
		t.Errorf("identity unstable project=%q/%q card=%q/%q/%q", p0, pr, c0, ca, cr)
	}
	t.Log("AC-TID-004 executed mutate=1 archive=1 restore=1 reopen=1")
}

func TestTodoIdentityAC005RuntimeLinksLiveAndArchived(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	s := NewBacklogStore(BacklogPathForRoot(root))
	live, _, err := s.Add("live")
	if err != nil {
		t.Fatal(err)
	}
	archived, _, err := s.Add("archived")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Mutate(func(r *BacklogRecord) error { return r.ArchiveCard(archived.ID) }); err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryRunStart(root, "run", BackendClaude, ""); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{live.ID, archived.ID} {
		if err := RecordFactoryCardAssignment(root, "run", id, "worker", ""); err != nil {
			t.Fatal(err)
		}
	}
	if err := RecordFactoryCardState(root, "run", live.ID, "worker", "", "completed", "card.completed"); err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	project := identityUUIDv7(t, identityJSON(t, r), "project_uuid", "project")
	liveUUID := identityUUIDv7(t, identityJSON(t, r.Items[0]), "card_uuid", "live")
	archiveUUID := identityUUIDv7(t, identityJSON(t, r.Archived[0].Item), "card_uuid", "archive")
	var runs, assignments []map[string]json.RawMessage
	runtime := identityJSON(t, r.Runtime)
	if err := json.Unmarshal(runtime["runs"], &runs); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(runtime["assignments"], &assignments); err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || len(assignments) != 2 {
		t.Fatalf("runtime regression runs=%d assignments=%d", len(runs), len(assignments))
	}
	runProject := identityUUIDv7(t, runs[0], "project_uuid", "run")
	seen := map[string]string{}
	for _, a := range assignments {
		var id string
		if err := json.Unmarshal(a["card_id"], &id); err != nil {
			t.Fatal(err)
		}
		ap := identityUUIDv7(t, a, "project_uuid", "assignment "+id)
		seen[id] = identityUUIDv7(t, a, "card_uuid", "assignment "+id)
		if project != "" && ap != project {
			t.Errorf("assignment project mismatch")
		}
		if id == live.ID {
			var reported string
			if err := json.Unmarshal(a["reported_state"], &reported); err != nil {
				t.Fatal(err)
			}
			if reported != "completed" || r.Items[0].State != BacklogStateQueued {
				t.Errorf("reported_state became card lifecycle authority: reported=%q card_state=%q", reported, r.Items[0].State)
			}
		}
	}
	if project != "" && (runProject != project || seen[live.ID] != liveUUID || seen[archived.ID] != archiveUUID) {
		t.Errorf("runtime linkage mismatch")
	}
	t.Log("AC-TID-005 executed runs=1 assignments=2 live=1 archive=1 state_updates=1")
}

func TestTodoIdentityAC005MissingRuntimeCardMutationZero(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	s := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := s.Add("survivor"); err != nil {
		t.Fatal(err)
	}
	before, _ := s.LoadPure()
	b, _ := json.Marshal(before)
	err := RecordFactoryCardAssignment(root, "missing", "t999", "worker", "")
	after, e := s.LoadPure()
	if e != nil {
		t.Fatal(e)
	}
	a, _ := json.Marshal(after)
	if err == nil || !bytes.Equal(b, a) {
		t.Errorf("missing card must error/change-zero err=%v equal=%v", err, bytes.Equal(b, a))
	}
	t.Log("AC-TID-005 regression guard executed missing=1 mutation_zero=1")
}

func TestTodoIdentityAC005AmbiguousLiveArchiveMutationZero(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	s := NewBacklogStore(BacklogPathForRoot(root))
	card, _, err := s.Add("ambiguous")
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryRunStart(root, "ambiguous-run", BackendClaude, ""); err != nil {
		t.Fatal(err)
	}
	db := identityDB(t, s)
	if _, err := db.Exec(`INSERT INTO archived_items(seq,id,text,added_at,spec_id,state,position) SELECT seq,id,text,added_at,spec_id,state,0 FROM items WHERE id=?`, card.ID); err != nil {
		t.Fatal(err)
	}
	var beforeRuns, beforeAssignments, beforeLive, beforeArchive, beforeIdentities int
	if err := db.QueryRow(`SELECT count(*) FROM todo_runtime_runs`).Scan(&beforeRuns); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM todo_runtime_assignments`).Scan(&beforeAssignments); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM items WHERE id=?`, card.ID).Scan(&beforeLive); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM archived_items WHERE id=?`, card.ID).Scan(&beforeArchive); err != nil {
		t.Fatal(err)
	}
	var identityTable int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='todo_identities'`).Scan(&identityTable); err != nil {
		t.Fatal(err)
	}
	if identityTable == 1 {
		if err := db.QueryRow(`SELECT count(*) FROM todo_identities`).Scan(&beforeIdentities); err != nil {
			t.Fatal(err)
		}
	}
	writeErr := RecordFactoryCardAssignment(root, "ambiguous-run", card.ID, "worker", "")
	var afterRuns, afterAssignments, afterLive, afterArchive, afterIdentities int
	if err := db.QueryRow(`SELECT count(*) FROM todo_runtime_runs`).Scan(&afterRuns); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM todo_runtime_assignments`).Scan(&afterAssignments); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM items WHERE id=?`, card.ID).Scan(&afterLive); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM archived_items WHERE id=?`, card.ID).Scan(&afterArchive); err != nil {
		t.Fatal(err)
	}
	if identityTable == 1 {
		if err := db.QueryRow(`SELECT count(*) FROM todo_identities`).Scan(&afterIdentities); err != nil {
			t.Fatal(err)
		}
	}
	unchanged := beforeRuns == afterRuns && beforeAssignments == afterAssignments && beforeLive == afterLive && beforeArchive == afterArchive && beforeIdentities == afterIdentities
	t.Logf("AC-TID-005 ambiguous candidates=%d writer_error=%v mutation_zero=%v", beforeLive+beforeArchive, writeErr, unchanged)
	if beforeLive+beforeArchive != 2 {
		t.Fatalf("GAP: ambiguous fixture candidates=%d want=2", beforeLive+beforeArchive)
	}
	if writeErr == nil || !unchanged {
		t.Errorf("ambiguous live/archive card must error/change-zero: err=%v unchanged=%v assignments=%d/%d", writeErr, unchanged, beforeAssignments, afterAssignments)
	}
}

func TestTodoIdentityAC006ConcurrentDistinctHandles(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	path := BacklogPathForRoot(t.TempDir())
	stores := []*BacklogStore{NewBacklogStore(path), NewBacklogStore(path)}
	start := make(chan struct{})
	type res struct {
		item *BacklogItem
		err  error
	}
	ch := make(chan res, 2)
	var wg sync.WaitGroup
	for i, s := range stores {
		wg.Add(1)
		go func(i int, s *BacklogStore) {
			defer wg.Done()
			<-start
			it, _, err := s.Add(fmt.Sprintf("writer-%d", i))
			ch <- res{it, err}
		}(i, s)
	}
	close(start)
	wg.Wait()
	close(ch)
	issued := map[string]bool{}
	success := 0
	for x := range ch {
		if x.err != nil {
			t.Errorf("writer failed: %v", x.err)
			continue
		}
		success++
		v := identityUUIDv7(t, identityJSON(t, x.item), "card_uuid", "concurrent Add")
		if v != "" {
			if issued[v] {
				t.Errorf("duplicate UUID %s", v)
			}
			issued[v] = true
		}
	}
	r, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	p := identityUUIDv7(t, identityJSON(t, r), "project_uuid", "concurrent project")
	for _, it := range r.Items {
		identityUUIDv7(t, identityJSON(t, it), "card_uuid", "stored "+it.ID)
	}
	if success != 2 || len(r.Items) != 2 || r.LastSeq != 2 || p == "" || len(issued) != 2 {
		t.Errorf("conservation success=%d items=%d last_seq=%d project=%q uuids=%d", success, len(r.Items), r.LastSeq, p, len(issued))
	}
	t.Logf("AC-TID-006 executed handles=2 barrier=1 successes=%d cards=%d UUIDs=%d", success, len(r.Items), len(issued))
}

func TestTodoIdentityAC006FutureIdentityVersionFailClosed(t *testing.T) {
	for _, op := range []string{"LoadPure", "writer"} {
		t.Run(op, func(t *testing.T) {
			s, db := identityLegacySQLite(t, false)
			identityCreate(t, db, "999")
			if op == "LoadPure" {
				before, err := s.LoadPure()
				if err == nil {
					t.Errorf("LoadPure accepted future identity version: %+v", before)
				}
				return
			}
			var beforeItems, beforeLastSeq, beforeIDs int
			if err := db.QueryRow(`SELECT count(*) FROM items`).Scan(&beforeItems); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT CAST(value AS INTEGER) FROM meta WHERE key='last_seq'`).Scan(&beforeLastSeq); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT count(*) FROM todo_identities`).Scan(&beforeIDs); err != nil {
				t.Fatal(err)
			}
			_, _, werr := s.Add("reject")
			var afterItems, afterLastSeq, afterIDs int
			if err := db.QueryRow(`SELECT count(*) FROM items`).Scan(&afterItems); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT CAST(value AS INTEGER) FROM meta WHERE key='last_seq'`).Scan(&afterLastSeq); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT count(*) FROM todo_identities`).Scan(&afterIDs); err != nil {
				t.Fatal(err)
			}
			unchanged := beforeItems == afterItems && beforeLastSeq == afterLastSeq && beforeIDs == afterIDs
			if werr == nil || !unchanged {
				t.Errorf("writer accepted/mutated future identity version err=%v unchanged=%v before=%d/%d/%d after=%d/%d/%d", werr, unchanged, beforeItems, beforeLastSeq, beforeIDs, afterItems, afterLastSeq, afterIDs)
			}
		})
	}
	for _, tc := range []struct {
		name string
		seed string
	}{
		{"table-without-stamp", `CREATE TABLE todo_identities(entity_kind TEXT NOT NULL,local_id TEXT NOT NULL,uuid TEXT NOT NULL UNIQUE,PRIMARY KEY(entity_kind,local_id))`},
		{"stamp-without-table", `INSERT INTO meta(key,value) VALUES('identity_schema_version','1')`},
		{"wrong-column-type", `CREATE TABLE todo_identities(entity_kind INTEGER NOT NULL,local_id TEXT NOT NULL,uuid TEXT NOT NULL UNIQUE,PRIMARY KEY(entity_kind,local_id)); INSERT INTO meta(key,value) VALUES('identity_schema_version','1')`},
		{"missing-column", `CREATE TABLE todo_identities(entity_kind TEXT NOT NULL,local_id TEXT NOT NULL,PRIMARY KEY(entity_kind,local_id)); INSERT INTO meta(key,value) VALUES('identity_schema_version','1')`},
		{"wrong-column-order", `CREATE TABLE todo_identities(local_id TEXT NOT NULL,entity_kind TEXT NOT NULL,uuid TEXT NOT NULL UNIQUE,PRIMARY KEY(local_id,entity_kind)); INSERT INTO meta(key,value) VALUES('identity_schema_version','1')`},
		{"uuid-without-unique", `CREATE TABLE todo_identities(entity_kind TEXT NOT NULL,local_id TEXT NOT NULL,uuid TEXT NOT NULL,PRIMARY KEY(entity_kind,local_id)); INSERT INTO meta(key,value) VALUES('identity_schema_version','1')`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, db := identityLegacySQLite(t, false)
			if _, err := db.Exec(tc.seed); err != nil {
				t.Fatal(err)
			}
			before := identitySchema(t, db)
			if _, err := s.LoadPure(); err == nil {
				t.Fatal("LoadPure accepted partial or malformed identity schema")
			}
			if after := identitySchema(t, db); after != before {
				t.Fatal("failed pure read mutated partial or malformed identity schema")
			}
		})
	}
	for _, op := range []string{"LoadPure", "writer"} {
		t.Run("partial-uuid-unique-"+op, func(t *testing.T) {
			s, db := identityLegacySQLite(t, false)
			const duplicate = "01890f3e-8b00-7000-8000-000000000001"
			if _, err := db.Exec(`
CREATE TABLE todo_identities(entity_kind TEXT NOT NULL,local_id TEXT NOT NULL,uuid TEXT NOT NULL,PRIMARY KEY(entity_kind,local_id));
CREATE UNIQUE INDEX uuid_project_only ON todo_identities(uuid) WHERE entity_kind='project';
INSERT INTO meta(key,value) VALUES('identity_schema_version','1');
INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES('card','t1',?),('card','t2',?);`, duplicate, duplicate); err != nil {
				t.Fatal(err)
			}
			var duplicateCount int
			if err := db.QueryRow(`SELECT count(*) FROM todo_identities WHERE uuid=?`, duplicate).Scan(&duplicateCount); err != nil {
				t.Fatal(err)
			}
			if duplicateCount != 2 {
				t.Fatalf("partial UNIQUE counterexample not reached: duplicate rows=%d want=2", duplicateCount)
			}
			before := identityPersistentState(t, db)
			var err error
			if op == "LoadPure" {
				_, err = s.LoadPure()
			} else {
				_, _, err = s.Add("must be rejected")
			}
			after := identityPersistentState(t, db)
			if err == nil || before != after {
				t.Errorf("%s accepted/mutated partial UNIQUE(uuid): err=%v mutation_zero=%v duplicates=%d", op, err, before == after, duplicateCount)
			}
		})
	}
	const validV7 = "01890f3e-8b00-7000-8000-000000000001"
	for _, tc := range []struct {
		name string
		row  string
	}{
		{"malformed-uuid", `INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES('project','project','not-a-uuid')`},
		{"unknown-kind", `INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES('unknown','x','` + validV7 + `')`},
		{"invalid-project-local-id", `INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES('project','other','` + validV7 + `')`},
		{"empty-card-local-id", `INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES('card','','` + validV7 + `')`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, db := identityLegacySQLite(t, false)
			identityCreate(t, db, "1")
			if _, err := db.Exec(tc.row); err != nil {
				t.Fatal(err)
			}
			before := identitySchema(t, db)
			if _, err := s.LoadPure(); err == nil {
				t.Fatal("LoadPure accepted malformed identity row")
			}
			if after := identitySchema(t, db); after != before {
				t.Fatal("failed pure read mutated malformed identity row")
			}
		})
	}
	for _, tc := range []struct {
		name  string
		value string
	}{
		{"malformed-preferred", "not-a-uuid"},
		{"mismatched-preferred", "01890f3e-8b01-7000-8000-000000000002"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewBacklogStore(BacklogPathForRoot(t.TempDir()))
			if _, _, err := s.Add("stable"); err != nil {
				t.Fatal(err)
			}
			before, err := s.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			beforeJSON, _ := json.Marshal(before)
			err = s.Mutate(func(rec *BacklogRecord) error {
				value := tc.value
				rec.Items[0].CardUUID = &value
				return nil
			})
			if err == nil {
				t.Fatal("writer accepted malformed or mismatched preferred identity")
			}
			after, loadErr := s.LoadPure()
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			afterJSON, _ := json.Marshal(after)
			if !bytes.Equal(beforeJSON, afterJSON) {
				t.Fatal("rejected preferred identity mutation changed the record")
			}
		})
	}
	t.Run("duplicate-preferred-on-new-local-id", func(t *testing.T) {
		s := NewBacklogStore(BacklogPathForRoot(t.TempDir()))
		if _, _, err := s.Add("stable"); err != nil {
			t.Fatal(err)
		}
		before, err := s.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		beforeJSON, _ := json.Marshal(before)
		err = s.Mutate(func(rec *BacklogRecord) error {
			rec.Items = append(rec.Items, BacklogItem{
				ID:       "t999",
				Text:     "duplicate preferred UUID",
				AddedAt:  rec.Items[0].AddedAt,
				State:    BacklogStateQueued,
				CardUUID: cloneIdentity(rec.Items[0].CardUUID),
			})
			return nil
		})
		if err == nil {
			t.Fatal("writer accepted one UUID for two card local IDs")
		}
		after, loadErr := s.LoadPure()
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		afterJSON, _ := json.Marshal(after)
		if !bytes.Equal(beforeJSON, afterJSON) {
			t.Fatal("rejected duplicate preferred UUID changed the record")
		}
	})
	t.Log("AC-TID-006 executed future_version operations=2 schema_guards=6 partial_unique_operations=2 row_guards=4 preferred_guards=3")
}

func TestTodoIdentityAC006RetiredRegression(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	s := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := s.Add("before"); err != nil {
		t.Fatal(err)
	}
	before, _ := s.LoadPure()
	b, _ := json.Marshal(before)
	if err := os.WriteFile(filepath.Join(filepath.Dir(s.Path()), backlogRetiredFileName), []byte("target"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := s.Add("reject")
	runtimeErr := RecordFactoryRunStart(root, "retired", BackendClaude, "")
	after, e := s.LoadPure()
	if e != nil {
		t.Fatal(e)
	}
	a, _ := json.Marshal(after)
	if !errors.Is(err, ErrBacklogRelocated) || !errors.Is(runtimeErr, ErrBacklogRelocated) || !bytes.Equal(b, a) {
		t.Errorf("retired regression add=%v runtime=%v equal=%v", err, runtimeErr, bytes.Equal(b, a))
	}
	t.Log("AC-TID-006 retired regression executed writers=2 mutation_zero=1")
}

func TestTodoIdentityAC007TimestampNonAuthority(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	s := NewBacklogStore(BacklogPathForRoot(root))
	first, _, err := s.Add("same")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := s.Add("same")
	if err != nil {
		t.Fatal(err)
	}
	specID := "SPEC-TIMESTAMP-001"
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# timestamp authority fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RecordFactoryCardState(root, "time-run", first.ID, "owner-before-uuid", specID, "completed", "card.completed"); err != nil {
		t.Fatal(err)
	}
	beforeUUID, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(beforeUUID.Runtime.Assignments) != 1 {
		t.Fatalf("runtime assignment precondition=%d want=1", len(beforeUUID.Runtime.Assignments))
	}
	beforeAssignment := beforeUUID.Runtime.Assignments[0]
	var beforeProvenance factoryProvenance
	if err := json.Unmarshal([]byte(beforeAssignment.ProvenanceJSON), &beforeProvenance); err != nil {
		t.Fatal(err)
	}
	if beforeAssignment.OwnerLabel != "owner-before-uuid" || beforeProvenance.SpecID != specID || beforeProvenance.SpecPath == "" || beforeProvenance.SpecSHA256 == "" || beforeProvenance.CapturedAt == "" {
		t.Fatalf("owner/provenance fixture not reached: assignment=%+v provenance=%+v", beforeAssignment, beforeProvenance)
	}
	db := identityDB(t, s)
	// GREEN writers already create identities. Replace only that extension so
	// this fixture can seed deliberately reversed UUIDv7 timestamps.
	identityRemove(t, db)
	identityCreate(t, db, "1")
	early, late := "01890f3e-8b00-7000-8000-000000000001", "01890f3e-8b01-7000-8000-000000000002"
	for _, v := range []string{early, late} {
		u, e := uuid.Parse(v)
		if e != nil || u.Version() != 7 {
			t.Fatalf("bad fixture UUIDv7 %q: %v", v, e)
		}
	}
	if _, err := db.Exec(`INSERT INTO todo_identities(entity_kind,local_id,uuid) VALUES('project','project',?),('card',?,?),('card',?,?)`, "01890f3e-8aff-7000-8000-000000000000", first.ID, late, second.ID, early); err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Items) != 2 || r.Items[0].State != BacklogStateQueued || r.Items[1].State != BacklogStateQueued || r.Items[0].AddedAt == "" || r.Items[1].AddedAt == "" {
		t.Fatalf("business fields changed: %+v", r.Items)
	}
	beforeBusiness, afterBusiness := beforeUUID.Runtime, r.Runtime
	for i := range beforeBusiness.Runs {
		beforeBusiness.Runs[i].ProjectUUID = nil
	}
	for i := range afterBusiness.Runs {
		afterBusiness.Runs[i].ProjectUUID = nil
	}
	for i := range beforeBusiness.Assignments {
		beforeBusiness.Assignments[i].ProjectUUID = nil
		beforeBusiness.Assignments[i].CardUUID = nil
	}
	for i := range afterBusiness.Assignments {
		afterBusiness.Assignments[i].ProjectUUID = nil
		afterBusiness.Assignments[i].CardUUID = nil
	}
	if !reflect.DeepEqual(beforeBusiness, afterBusiness) {
		t.Errorf("UUID-only insertion changed runtime owner/provenance: before=%+v after=%+v", beforeBusiness, afterBusiness)
	}
	if beforeUUID.Items[0].AddedAt != r.Items[0].AddedAt || beforeUUID.Items[1].AddedAt != r.Items[1].AddedAt || beforeUUID.Items[0].State != r.Items[0].State || beforeUUID.Items[1].State != r.Items[1].State {
		t.Error("UUID-only insertion changed AddedAt/State")
	}
	a := identityUUIDv7(t, identityJSON(t, r.Items[0]), "card_uuid", "later UUID card")
	b := identityUUIDv7(t, identityJSON(t, r.Items[1]), "card_uuid", "earlier UUID card")
	if a != late || b != early {
		t.Errorf("projection got=%q,%q want=%q,%q", a, b, late, early)
	}
	if len(r.Runtime.Assignments) != 1 || r.Runtime.Assignments[0].ReportedState != "completed" || r.Runtime.Assignments[0].OwnerLabel != "owner-before-uuid" || r.Runtime.Assignments[0].ProvenanceJSON != beforeAssignment.ProvenanceJSON || r.Items[0].State != BacklogStateQueued {
		t.Errorf("reported_state became authority")
	}
	t.Log("AC-TID-007 executed UUIDv7_timestamps=2 authority_fields=added_at,state,reported_state,owner_label,provenance_json")
}

// Historical RED selector names remain as stable audit carriers. The stricter
// AC-specific tests above extend their contracts without erasing the original
// evidence names used by the first baseline report.
func TestTodoIdentityREDAddReturnAndProjectIsolation(t *testing.T) {
	TestTodoIdentityAC001GitAndNonGitUUIDv7(t)
}

func TestTodoIdentityREDLifetimeAcrossArchiveRestore(t *testing.T) {
	TestTodoIdentityAC004LifecycleStability(t)
}

func TestTodoIdentityREDRuntimeLinksActualCard(t *testing.T) {
	TestTodoIdentityAC005RuntimeLinksLiveAndArchived(t)
}

func TestTodoIdentityLegacyPureReadDoesNotIssueOrWrite(t *testing.T) {
	TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation(t)
}
