// classification_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M1 acceptance
// tests: the classification value sets (AC-TCD-004's closed-set contract at
// the type layer), the additive per-item field's byte-identical round trip
// (AC-TCD-002), the SQLite mirror column (AC-TCD-002's NULL half), and the
// read-side default derivation (REQ-TCD-014).
package factory

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

func TestCardClassificationValidateAcceptsClosedSets(t *testing.T) {
	valid := []CardClassification{
		{Priority: ClassPriorityHigh, Blocked: false, Mode: ClassModeSerial, Decider: DeciderIdentityLLM},
		{Priority: ClassPriorityNormal, Blocked: true, Mode: ClassModeParallelizable, Decider: DeciderIdentityHuman},
		{Priority: ClassPriorityLow, Blocked: false, Mode: ClassModeSerial, Decider: DeciderIdentityDefault},
	}
	for i, c := range valid {
		if err := ValidateCardClassification(c); err != nil {
			t.Errorf("valid[%d] %+v refused: %v", i, c, err)
		}
	}
}

func TestCardClassificationValidateRefusesOutOfSet(t *testing.T) {
	cases := []struct {
		name string
		c    CardClassification
	}{
		{"priority urgent", CardClassification{Priority: "urgent", Mode: ClassModeSerial, Decider: DeciderIdentityLLM}},
		{"priority empty", CardClassification{Priority: "", Mode: ClassModeSerial, Decider: DeciderIdentityLLM}},
		{"mode serially", CardClassification{Priority: ClassPriorityNormal, Mode: "serially", Decider: DeciderIdentityLLM}},
		{"mode empty", CardClassification{Priority: ClassPriorityNormal, Mode: "", Decider: DeciderIdentityLLM}},
		{"decider jev", CardClassification{Priority: ClassPriorityNormal, Mode: ClassModeSerial, Decider: "jev"}},
		{"decider llm+jev", CardClassification{Priority: ClassPriorityNormal, Mode: ClassModeSerial, Decider: DeciderIdentityLLMJev}},
		{"decider unknown", CardClassification{Priority: ClassPriorityNormal, Mode: ClassModeSerial, Decider: "gpt"}},
		{"decider empty", CardClassification{Priority: ClassPriorityNormal, Mode: ClassModeSerial, Decider: ""}},
	}
	for _, tc := range cases {
		err := ValidateCardClassification(tc.c)
		if err == nil {
			t.Errorf("%s: refused nothing — %+v validated", tc.name, tc.c)
			continue
		}
		if strings.Contains(tc.name, "jev") && !strings.Contains(err.Error(), "jev") {
			t.Errorf("%s: refusal %q does not name the refused identity", tc.name, err)
		}
	}
}

func TestParseCardClassificationJSON(t *testing.T) {
	good := []byte(`{"priority":"high","blocked":true,"mode":"parallelizable","decider":"llm","reason":"urgency"}`)
	c, err := ParseCardClassificationJSON(good)
	if err != nil {
		t.Fatalf("parse valid: %v", err)
	}
	if c.Priority != ClassPriorityHigh || !c.Blocked || c.Mode != ClassModeParallelizable || c.Decider != DeciderIdentityLLM || c.Reason != "urgency" {
		t.Errorf("parsed %+v, want the supplied values", c)
	}
	for _, bad := range []string{
		`{"priority":"urgent","mode":"serial","decider":"llm"}`,
		`{"priority":"normal","mode":"serially","decider":"llm"}`,
		`{"priority":"normal","mode":"serial","decider":"jev"}`,
		`{"priority":"normal","mode":"serial","decider":"llm"`, // malformed JSON
		`not json at all`,
	} {
		if _, err := ParseCardClassificationJSON([]byte(bad)); err == nil {
			t.Errorf("parse %q: refused nothing", bad)
		}
	}
}

// TestEffectiveCardClassificationDerivesAbsentDefaults — REQ-TCD-014: the
// absent-field READ default is positive derivation (normal / false / serial /
// default), never a hidden fourth priority or third mode. AC-TCD-002's read
// half rests on the same derivation.
func TestEffectiveCardClassificationDerivesAbsentDefaults(t *testing.T) {
	got := EffectiveCardClassification(BacklogItem{ID: "t1"})
	want := CardClassification{
		Priority: ClassPriorityNormal,
		Blocked:  false,
		Mode:     ClassModeSerial,
		Decider:  DeciderIdentityDefault,
	}
	if got.Priority != want.Priority || got.Blocked != want.Blocked || got.Mode != want.Mode || got.Decider != want.Decider {
		t.Errorf("effective classification of an absent field = %+v, want %+v", got, want)
	}
	// A present classification round-trips through the same read seam
	// untouched — derivation applies only to absence.
	present := CardClassification{Priority: ClassPriorityHigh, Blocked: true, Mode: ClassModeParallelizable, Decider: DeciderIdentityLLM}
	got = EffectiveCardClassification(BacklogItem{ID: "t2", Classification: &present})
	if got != present {
		t.Errorf("effective classification of a present field = %+v, want the record's own %+v", got, present)
	}
}

// TestBacklogClassificationAdditiveMarshalByteIdentical — AC-TCD-002: a
// pre-SPEC record (items without the classification key) unmarshals and
// re-marshals byte-identically. The field is `omitempty` after the Landing
// precedent, so absence stays absence in the bytes.
func TestBacklogClassificationAdditiveMarshalByteIdentical(t *testing.T) {
	pre := []byte(`{"project_uuid":null,"version":1,"last_seq":2,"items":[` +
		`{"id":"t1","text":"alpha","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued","card_uuid":null},` +
		`{"id":"t2","text":"beta","added_at":"2026-01-02T00:00:00Z","spec_id":null,"state":"picked","card_uuid":null,"picked_at":"2026-01-02T01:00:00Z"}` +
		`],"findings":[],"archived":[],"runtime":{"runs":[],"assignments":[]}}`)
	var rec BacklogRecord
	if err := json.Unmarshal(pre, &rec); err != nil {
		t.Fatalf("unmarshal pre-SPEC record: %v", err)
	}
	for _, it := range rec.Items {
		if it.Classification != nil {
			t.Errorf("item %s parsed a classification from bytes that carry none", it.ID)
		}
	}
	post, err := json.Marshal(&rec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(post) != string(pre) {
		t.Errorf("round trip is not byte-identical:\n got: %s\nwant: %s", post, pre)
	}
}

// TestBacklogClassificationSQLiteMirror — AC-TCD-002's storage half: one add
// writes classification to exactly the new card; the pre-existing card's
// classification stays absent (nil in the record, NULL in the column).
func TestBacklogClassificationSQLiteMirror(t *testing.T) {
	store := NewBacklogStore(BacklogPathForRoot(t.TempDir()))
	if _, _, err := store.Add("legacy card without classification"); err != nil {
		t.Fatalf("seed add t1: %v", err)
	}
	if _, _, err := store.Add("second card"); err != nil {
		t.Fatalf("seed add t2: %v", err)
	}
	cls := CardClassification{
		Priority: ClassPriorityHigh, Blocked: false, Mode: ClassModeParallelizable,
		Decider: DeciderIdentityLLM, ClassifiedAt: "2026-09-29T00:00:00Z", Reason: "test",
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == "t2" {
				c := cls
				rec.Items[i].Classification = &c
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("classify t2: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0].Classification != nil {
		t.Errorf("t1 classification = %+v, want absent", rec.Items[0].Classification)
	}
	if rec.Items[1].Classification == nil {
		t.Fatalf("t2 carries no classification after the write")
	}
	if *rec.Items[1].Classification != cls {
		t.Errorf("t2 classification = %+v, want %+v", *rec.Items[1].Classification, cls)
	}

	// The column is physically present and stores the JSON TEXT for the
	// classified row and NULL for the unclassified one.
	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	defer func() { _ = eng.close() }()
	rows, err := eng.db.QueryContext(context.Background(), `SELECT id, classification FROM items ORDER BY seq`)
	if err != nil {
		t.Fatalf("read classification column: %v", err)
	}
	defer func() { _ = rows.Close() }()
	got := map[string]*string{}
	for rows.Next() {
		var id string
		var col sql.NullString
		if err := rows.Scan(&id, &col); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if col.Valid {
			v := col.String
			got[id] = &v
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate: %v", err)
	}
	if got["t1"] != nil {
		t.Errorf("t1 classification column = %q, want NULL", *got["t1"])
	}
	if got["t2"] == nil || !strings.Contains(*got["t2"], `"priority":"high"`) {
		t.Errorf("t2 classification column = %v, want JSON text carrying the classification", got["t2"])
	}
}

// TestSchemaFreezeCarriesClassificationColumn — M1's recorded schema decision:
// the classification TEXT column lands additively on both card-bearing
// tables through the pragma_table_info-gated ADD COLUMN path (the
// landing/stamp precedent), followed later by the claim-lease pair
// picked_by/lease_expires_at, so a fresh and an upgraded database converge
// on one tuple.
// This is the classification sibling of AC-TST-011's recorded check.
func TestSchemaFreezeCarriesClassificationColumn(t *testing.T) {
	store := archiveFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	defer func() { _ = eng.close() }()
	wantItems := "seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL " +
		"spec_id:TEXT:0:NULL state:TEXT:1:NULL landing:TEXT:0:NULL " +
		"picked_at:TEXT:0:NULL dropped_at:TEXT:0:NULL classification:TEXT:0:NULL " +
		"picked_by:TEXT:0:NULL lease_expires_at:TEXT:0:NULL " +
		// SPEC-TODO-CARD-ISSUANCE-001: the issuance attributes are the next
		// additive pair on the same ensure path (card t1454).
		"issuance:TEXT:0:NULL"
	if got := columnTupleSequence(t, eng, "items"); got != wantItems {
		t.Errorf("items column tuples =\n %s\nwant\n %s", got, wantItems)
	}
	wantArchived := "seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL " +
		"spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL landing:TEXT:0:NULL " +
		"picked_at:TEXT:0:NULL dropped_at:TEXT:0:NULL archived_at:TEXT:0:NULL " +
		"landing_verdict:TEXT:0:NULL classification:TEXT:0:NULL " +
		"picked_by:TEXT:0:NULL lease_expires_at:TEXT:0:NULL issuance:TEXT:0:NULL"
	if got := columnTupleSequence(t, eng, "archived_items"); got != wantArchived {
		t.Errorf("archived_items column tuples =\n %s\nwant\n %s", got, wantArchived)
	}
}

// TestPriorityRank pins the exported rank accessor (SPEC-TODO-AUTO-PRIORITY-001
// M1): the closed priority set ranks high above normal above low, and a value
// outside the set ranks below every member instead of tying with one. The
// accessor is the only spelling of the ranking outside this package.
func TestPriorityRank(t *testing.T) {
	cases := []struct {
		name     string
		priority string
		want     int
	}{
		{"high", ClassPriorityHigh, 2},
		{"normal", ClassPriorityNormal, 1},
		{"low", ClassPriorityLow, 0},
		{"unrecognized value ranks below the closed set", "urgent", -1},
		{"empty value ranks below the closed set", "", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PriorityRank(tc.priority); got != tc.want {
				t.Errorf("PriorityRank(%q) = %d, want %d", tc.priority, got, tc.want)
			}
		})
	}
	if PriorityRank(ClassPriorityHigh) <= PriorityRank(ClassPriorityNormal) ||
		PriorityRank(ClassPriorityNormal) <= PriorityRank(ClassPriorityLow) {
		t.Errorf("rank order is not high > normal > low")
	}
}
