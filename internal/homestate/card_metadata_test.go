package homestate

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SPEC-CARD-METADATA-VERBS-001 M1: the card-keyed metadata audit surface and
// the set-spec record-layer mutation function. The RED scope is plan.md §F
// M1's four named tests plus the entry points the M2/M3 verbs consume.

var mdNow = time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)

// mdOpen opens a fresh factory database in a hermetic sandbox.
func mdOpen(t *testing.T) *FactoryDB {
	t.Helper()
	db, err := OpenFactory(factorySandbox(t))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// mdPlaceCard inserts one card row directly at the named state and version —
// a fixture placement, so the sweep can park a row at every F1 state without
// walking the transition graph.
func mdPlaceCard(t *testing.T, db *FactoryDB, runID, cardID, state string, version int64) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,updated_at,spec_id) VALUES(?,?,?,?,?,'')`,
		runID, cardID, state, version, "2026-10-21T00:00:00Z"); err != nil {
		t.Fatalf("place card %s at %s: %v", cardID, state, err)
	}
}

func TestCardMetadataEvent_RowShape(t *testing.T) {
	db := mdOpen(t)
	err := db.RecordCardMetadataEvent(context.Background(), CardMetadataEvent{
		CardID: "t1", Kind: "set-spec", Actor: "operator", Session: "sess-1",
		Field: "spec_id", Old: "", New: "SPEC-FOO-001", Reason: "link the SPEC after pick",
	}, mdNow)
	if err != nil {
		t.Fatalf("append metadata event: %v", err)
	}
	var (
		kind, payload, created string
		cardID                 string
	)
	if err := db.DB.QueryRow(`SELECT card_id,kind,payload_json,created_at FROM card_metadata_events ORDER BY seq DESC LIMIT 1`).
		Scan(&cardID, &kind, &payload, &created); err != nil {
		t.Fatalf("read metadata event row: %v", err)
	}
	if cardID != "t1" || kind != "set-spec" {
		t.Fatalf("row keys = (%s, %s), want (t1, set-spec)", cardID, kind)
	}
	if created != mdNow.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("created_at = %q, want the append time", created)
	}
	var decoded struct {
		Actor        string `json:"actor"`
		Session      string `json:"session"`
		Field        string `json:"field"`
		Old          string `json:"old"`
		New          string `json:"new"`
		Reason       string `json:"reason"`
		ClassifiedAt string `json:"classified_at"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if decoded.Actor != "operator" || decoded.Session != "sess-1" || decoded.Field != "spec_id" ||
		decoded.Old != "" || decoded.New != "SPEC-FOO-001" || decoded.Reason != "link the SPEC after pick" {
		t.Fatalf("payload = %+v, want the recorded mutation verbatim", decoded)
	}

	// A second append gets its own autoincrement row (REQ-CMV-010: exactly one
	// row per completed mutation).
	if err := db.RecordCardMetadataEvent(context.Background(), CardMetadataEvent{
		CardID: "t1", Kind: "reclassify", Actor: "lane-9", Field: "classification",
		Old: `{"blocked":false,"mode":"serial","priority":"normal"}`, New: `{"blocked":false,"mode":"serial","priority":"high"}`,
		Reason: "raise priority", ClassifiedAt: mdNow.UTC().Format(time.RFC3339Nano),
	}, mdNow); err != nil {
		t.Fatalf("append second event: %v", err)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM card_metadata_events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("card_metadata_events rows = %d, want 2", n)
	}
}

func TestCardMetadataCatchUp_AppendsExactlyOneRow(t *testing.T) {
	db := mdOpen(t)
	// The direction-6 catch-up append is a separate entry point from the
	// after-both-commits append; both land rows in the same surface.
	err := db.RecordCardMetadataCatchUp(context.Background(), CardMetadataEvent{
		CardID: "t1", Kind: "set-spec", Actor: "operator", Field: "spec_id",
		Old: "SPEC-OLD-001", New: "SPEC-NEW-001", Reason: "crash recovery attestation",
	}, mdNow)
	if err != nil {
		t.Fatalf("append catch-up row: %v", err)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM card_metadata_events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("rows = %d, want exactly the one owed row", n)
	}
}

func TestLatestCardMetadataNewValue(t *testing.T) {
	db := mdOpen(t)
	// No row yet: found=false, empty value.
	if v, found, err := db.LatestCardMetadataNewValue(context.Background(), "t1", "set-spec", "spec_id"); err != nil || found || v != "" {
		t.Fatalf("latest with no rows = (%q, %v, %v), want (\"\", false, nil)", v, found, err)
	}
	// Two rows: X→A then A→B. The LATEST row's new value is B — an older row
	// whose new value equals a commanded value does not make a duplicate.
	for _, ev := range []CardMetadataEvent{
		{CardID: "t1", Kind: "set-spec", Actor: "operator", Field: "spec_id", Old: "", New: "SPEC-A-001", Reason: "r1"},
		{CardID: "t1", Kind: "set-spec", Actor: "operator", Field: "spec_id", Old: "SPEC-A-001", New: "SPEC-B-001", Reason: "r2"},
	} {
		if err := db.RecordCardMetadataEvent(context.Background(), ev, mdNow); err != nil {
			t.Fatalf("append %v: %v", ev.New, err)
		}
	}
	v, found, err := db.LatestCardMetadataNewValue(context.Background(), "t1", "set-spec", "spec_id")
	if err != nil || !found {
		t.Fatalf("latest = (%q, %v, %v), want found", v, found, err)
	}
	if v != "SPEC-B-001" {
		t.Fatalf("latest new value = %q, want SPEC-B-001 (the LATEST row, not any prior row)", v)
	}
	// A different card and a different kind read as no rows.
	if _, found, _ := db.LatestCardMetadataNewValue(context.Background(), "t2", "set-spec", "spec_id"); found {
		t.Fatal("card t2 reported a row it does not have")
	}
	if _, found, _ := db.LatestCardMetadataNewValue(context.Background(), "t1", "reclassify", "classification"); found {
		t.Fatal("kind reclassify reported a row it does not have")
	}
}

func TestValidSetSpecID_Grammar(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"SPEC-CARD-METADATA-VERBS-001", true},
		{"SPEC-FOO-001", true},
		{"SPEC-A1B2-001", true},
		{"spec-auth-001", false},  // lowercase
		{"SPEC-2FA-001", false},   // digit-leading middle segment
		{"SPEC-AUTH-1", false},    // tail not three digits
		{"AUTH-001", false},       // SPEC prefix missing
		{"", false},               // empty
		{"SPEC-AUTH-001 ", false}, // trailing space
		{"SPEC-AUTH-0012", false}, // four-digit tail
		{"SPEC--001", false},      // empty middle segment
		{"SPEC-AUTH-00A", false},  // letter in the tail
	}
	for _, tc := range cases {
		if got := ValidSetSpecID(tc.value); got != tc.want {
			t.Errorf("ValidSetSpecID(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestSetCardSpec_AnyState_Lands(t *testing.T) {
	states := CardStates()
	if len(states) != 21 {
		t.Fatalf("CardStates() = %d states, want the 21 the source enumerates", len(states))
	}
	for i, state := range states {
		db := mdOpen(t)
		cardID := "c" + strings.ReplaceAll(state, "-", "")
		mdPlaceCard(t, db, "run-md", cardID, state, 3)
		got, err := db.SetCardSpec(context.Background(), "run-md", cardID, "SPEC-LAND-001", 3, mdNow)
		if err != nil {
			t.Fatalf("SetCardSpec at state %s: %v", state, err)
		}
		if got.SpecID != "SPEC-LAND-001" {
			t.Fatalf("state %s: spec_id = %q, want SPEC-LAND-001", state, got.SpecID)
		}
		if got.Version != 4 {
			t.Fatalf("state %s: version = %d, want 4 (exactly +1)", state, got.Version)
		}
		if got.UpdatedAt != mdNow.UTC().Format(time.RFC3339Nano) {
			t.Fatalf("state %s: updated_at = %q, want the mutation time", state, got.UpdatedAt)
		}
		// The stored row carries the value too (not just the returned struct).
		if c := mdLoadCard(t, db, "run-md", cardID); c.SpecID != "SPEC-LAND-001" || c.Version != 4 {
			t.Fatalf("state %s: stored row = (spec %q, v%d)", state, c.SpecID, c.Version)
		}
		_ = i
		_ = db.Close()
	}
}

// mdLoadCard reads one row through a fresh connection.
func mdLoadCard(t *testing.T, db *FactoryDB, runID, cardID string) Card {
	t.Helper()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s: %v", cardID, err)
	}
	return c
}

// The contrast case: the while-picked refusal of RecordPicked still stands
// for a post-picked row — the new all-state write is this function alone, it
// never widens RecordPicked (plan.md §F M1, §G anti-pattern 1).
func TestSetCardSpec_RecordPickedContrast_StandsStill(t *testing.T) {
	db := mdOpen(t)
	mdPlaceCard(t, db, "run-md", "t1", CardRun, 2)
	spec := "SPEC-LATE-001"
	if _, err := db.RecordPicked(context.Background(), "run-md", "t1", CardFields{SpecID: &spec}, "assign", mdNow); err == nil {
		t.Fatal("RecordPicked accepted a field change on a run-state card; the while-picked rule must not widen")
	}
	if c := mdLoadCard(t, db, "run-md", "t1"); c.SpecID != "" || c.Version != 2 {
		t.Fatalf("refused RecordPicked left the row at (spec %q, v%d), want (\"\", v2)", c.SpecID, c.Version)
	}
}

func TestSetCardSpec_StaleVersion_Refused(t *testing.T) {
	db := mdOpen(t)
	mdPlaceCard(t, db, "run-md", "t1", CardRun, 2)
	// Advance the row under the caller's feet: N → N+1.
	if _, err := db.SetCardSpec(context.Background(), "run-md", "t1", "SPEC-NEW-001", 2, mdNow); err != nil {
		t.Fatalf("first write: %v", err)
	}
	// The stale presentation of N refuses and writes nothing.
	if _, err := db.SetCardSpec(context.Background(), "run-md", "t1", "SPEC-LATER-001", 2, mdNow); err == nil {
		t.Fatal("stale expected version was accepted")
	}
	if !strings.Contains(func() string {
		_, err := db.SetCardSpec(context.Background(), "run-md", "t1", "SPEC-LATER-001", 2, mdNow)
		return err.Error()
	}(), "stale card version") {
		t.Fatal("refusal does not carry ErrStaleVersion's text")
	}
	if c := mdLoadCard(t, db, "run-md", "t1"); c.SpecID != "SPEC-NEW-001" || c.Version != 3 {
		t.Fatalf("stale write left the row at (spec %q, v%d), want (SPEC-NEW-001, v3)", c.SpecID, c.Version)
	}
}

func TestRestoreCardSpec_CompensatingEntry(t *testing.T) {
	db := mdOpen(t)
	mdPlaceCard(t, db, "run-md", "t1", CardRun, 2)
	mdPlaceCard(t, db, "run-md", "t2", CardRun, 2)
	// The direction-2 shape: the record write moved t1 to v3 (spec NEW), the
	// compensation restores the prior value against that post-write version.
	if _, err := db.SetCardSpec(context.Background(), "run-md", "t1", "SPEC-NEW-001", 2, mdNow); err != nil {
		t.Fatalf("advance t1: %v", err)
	}
	// A restore through the same version-fenced mutation puts the prior
	// value back and bumps the version once.
	got, err := db.RestoreCardSpec(context.Background(), "run-md", "t1", "SPEC-PRIOR-001", 3, mdNow)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got.SpecID != "SPEC-PRIOR-001" || got.Version != 4 {
		t.Fatalf("restored = (spec %q, v%d), want (SPEC-PRIOR-001, v4)", got.SpecID, got.Version)
	}
	// A stale expected version refuses the restore too (same fencing).
	if _, err := db.RestoreCardSpec(context.Background(), "run-md", "t2", "SPEC-PRIOR-001", 9, mdNow); err == nil {
		t.Fatal("stale restore was accepted")
	}
	// A missing card row is ErrCardNotFound, never a created row (REQ-CMV-003
	// gate (b): the record layer creates nothing here).
	if _, err := db.RestoreCardSpec(context.Background(), "run-md", "tX", "SPEC-PRIOR-001", 1, mdNow); err == nil {
		t.Fatal("restore invented a card row")
	}
}

func TestSetCardSpec_RefusesInput_andLegacyRows(t *testing.T) {
	db := mdOpen(t)
	mdPlaceCard(t, db, "run-md", "t1", CardDone, 1)
	if _, err := db.SetCardSpec(context.Background(), "run-md", "t1", "", 1, mdNow); err == nil {
		t.Fatal("empty spec id accepted")
	}
	if _, err := db.SetCardSpec(context.Background(), "run-md", "missing", "SPEC-A-001", 1, mdNow); err != nil {
		if err != ErrCardNotFound {
			t.Fatalf("missing card error = %v, want ErrCardNotFound", err)
		}
	} else {
		t.Fatal("missing card row was accepted")
	}
}

// TestCardMetadataSurface_RowPreserving — REQ-CMV-014 / AC-CMV-015: adding
// the card_metadata_events surface rewrites, migrates, or deletes no existing
// row of any existing table, and the migration is idempotent on a second run.
func TestCardMetadataSurface_RowPreserving(t *testing.T) {
	root := factorySandbox(t)
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	seed := map[string]string{
		"meta":             `INSERT OR IGNORE INTO meta(key,value) VALUES('schema_probe','1')`,
		"workers":          `INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES('w1',1,'2026-10-21T00:00:00Z','2026-10-21T00:00:00Z')`,
		"runs":             `INSERT INTO runs(run_id,status,created_at,updated_at) VALUES('r1','active','2026-10-21T00:00:00Z','2026-10-21T00:00:00Z')`,
		"cards":            `INSERT INTO cards(run_id,card_id,state,version,updated_at) VALUES('r1','t1','picked',1,'2026-10-21T00:00:00Z')`,
		"events":           `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('r1','card.transition','{}','2026-10-21T00:00:00Z')`,
		"dead_letters":     `INSERT INTO dead_letters(envelope_json,error,created_at) VALUES('{}','boom','2026-10-21T00:00:00Z')`,
		"resume_handoffs":  `INSERT INTO resume_handoffs(status,schema_version,saved_at,body,body_sha256) VALUES('pending',1,'2026-10-21T00:00:00Z','body','abc')`,
		"memory_handoffs":  `INSERT INTO memory_handoffs(status,sprint,spec,result_status,body,index_line,body_sha256,created_at) VALUES('pending','s','SPEC-X-001','ok','b','i','abc','2026-10-21T00:00:00Z')`,
		"handoff_events":   `INSERT INTO handoff_events(flow,handoff_id,from_status,to_status,created_at) VALUES('resume',1,'','pending','2026-10-21T00:00:00Z')`,
		"leader_approvals": `INSERT INTO leader_approvals(card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at) VALUES('u1','r1','t1',1,'h','lead','leader','2026-10-21T00:00:00Z')`,
		"card_dispatch":    `INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES('t1','r1','2026-10-21T00:00:00Z')`,
	}
	for table, stmt := range seed {
		if _, err := db.DB.Exec(stmt); err != nil {
			t.Fatalf("seed %s: %v", table, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close after seed: %v", err)
	}

	raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(factoryPathOf(t, root)))
	if err != nil {
		t.Fatalf("snapshot open: %v", err)
	}
	before := mdDumpAll(t, raw, seed)
	_ = raw.Close()

	// Re-open: the DDL (the migration) runs again over a populated store.
	db2, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("re-open with migration: %v", err)
	}
	after := mdDumpAll(t, db2.DB, seed)
	for table := range seed {
		if strings.Join(after[table], "\n") != strings.Join(before[table], "\n") {
			t.Fatalf("%s rows changed by the migration:\n got %v\nwant %v", table, after[table], before[table])
		}
	}
	// The new surface exists.
	var present int
	if err := db2.DB.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='card_metadata_events'`).Scan(&present); err != nil || present != 1 {
		t.Fatalf("card_metadata_events present = %d (%v), want 1", present, err)
	}
	if err := db2.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	// Idempotent on a second run.
	db3, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("second re-open: %v", err)
	}
	defer func() { _ = db3.Close() }()
	again := mdDumpAll(t, db3.DB, seed)
	for table := range seed {
		if strings.Join(again[table], "\n") != strings.Join(before[table], "\n") {
			t.Fatalf("%s rows changed by the second migration run: %v", table, again[table])
		}
	}
}

// mdDumpAll renders every row of every seeded table as joined column
// strings — the row-level content comparison REQ-CMV-014 judges (file-byte
// identity is not the criterion, acceptance.md §A).
func mdDumpAll(t *testing.T, db *sql.DB, seed map[string]string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for table := range seed {
		rows, err := db.Query(`SELECT * FROM ` + table)
		if err != nil {
			t.Fatalf("dump %s: %v", table, err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]sql.NullString, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatalf("scan %s: %v", table, err)
			}
			parts := make([]string, len(vals))
			for i := range vals {
				parts[i] = vals[i].String
			}
			out[table] = append(out[table], strings.Join(parts, "|"))
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows %s: %v", table, err)
		}
		_ = rows.Close()
	}
	return out
}

// factoryPathOf resolves the sandbox's factory database path (the same
// layout OpenFactory created it under).
func factoryPathOf(t *testing.T, root string) string {
	t.Helper()
	path, err := FactoryDBPath(root)
	if err != nil {
		t.Fatalf("factory path: %v", err)
	}
	return path
}
