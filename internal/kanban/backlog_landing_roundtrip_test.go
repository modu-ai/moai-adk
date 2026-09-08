// backlog_landing_roundtrip_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card
// t359) M5: the landing evidence survives the legacy-JSON round trip on BOTH
// card-bearing tables, and the migration's own parity verification is the
// thing that notices when it does not.
//
// AC-TLE-017's risk is NOT "the evidence is lost". It is "the evidence is lost
// and the migration reports success" — the cutover flips authority onto the
// database and quarantines the legacy file, so a parity check that does not
// compare the column turns a recoverable loss into a permanent one. That is
// why the mutant this criterion mandates must red the PARITY CHECK and not
// merely an equality assertion in a test.
package kanban

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// landingRoundTripFixture is the record both tests below start from: one live
// card and one archived card, each carrying evidence of a DIFFERENT shape.
//
// The shapes differ deliberately. The live card carries an operator assertion
// (sha + sha_source), the archived card carries only an observation — so a
// carry that preserved one field group and dropped the other cannot pass both
// halves, and a comparison keyed on a single field cannot pass either.
func landingRoundTripFixture() *BacklogRecord {
	live := LandingEvidence{
		Ref:        "origin/main",
		RefHead:    "1111111111111111111111111111111111111111",
		ObservedAt: "2026-01-01T00:00:00Z",
		SHA:        "2222222222222222222222222222222222222222",
		SHASource:  LandingSHASourceOperator,
		SpecStatus: "completed",
	}
	archived := LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    "3333333333333333333333333333333333333333",
		ObservedAt: "2026-01-02T00:00:00Z",
	}
	return &BacklogRecord{
		Version: backlogVersion,
		LastSeq: 2,
		Items: []BacklogItem{
			{ID: "t1", Text: "live card", AddedAt: "2026-01-01T00:00:00Z",
				State: BacklogStateQueued, Landing: &live},
		},
		Findings: []BacklogFinding{},
		Archived: []BacklogArchiveEntry{
			{
				Item: BacklogItem{ID: "t2", Text: "archived card", AddedAt: "2026-01-02T00:00:00Z",
					State: BacklogStateQueued, Landing: &archived},
				Position: 0,
				Findings: []BacklogArchivedFinding{},
			},
		},
	}
}

// seedLegacyLandingJSON writes the fixture as a legacy backlog.json and returns
// a store over it. The next Load is the cutover: migrate, re-read, verify
// parity, and only then flip authority.
func seedLegacyLandingJSON(t *testing.T, rec *BacklogRecord) *BacklogStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "backlog.json")
	raw, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal legacy fixture: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write legacy fixture: %v", err)
	}
	return NewBacklogStore(path)
}

// AC-TLE-017 — the round trip preserves the evidence on both tables, and the
// migration's parity verification passes.
func TestBacklogLanding_RoundTripPreservesEvidence(t *testing.T) {
	source := landingRoundTripFixture()
	store := seedLegacyLandingJSON(t, source)

	got, err := store.Load()
	if err != nil {
		// A parity failure surfaces HERE, because migrateLegacyBacklog
		// refuses the cutover on one. Naming it separately from a decode or
		// open failure is what makes the mutant's red readable.
		t.Fatalf("Load (migration + parity) failed: %v", err)
	}

	if len(got.Items) != 1 || len(got.Archived) != 1 {
		t.Fatalf("migrated record has %d live / %d archived cards, want 1 / 1",
			len(got.Items), len(got.Archived))
	}
	assertLandingEqual(t, "live card t1", source.Items[0].Landing, got.Items[0].Landing)
	assertLandingEqual(t, "archived card t2",
		source.Archived[0].Item.Landing, got.Archived[0].Item.Landing)
}

// AC-TLE-017's parity half, fired directly: the check must compare the landing
// column on both tables, so a migration that dropped it cannot report success.
//
// This is the sibling of TestMigrationParityCatchesTamperedRecord's per-axis
// table, added rather than folded into it so the landing axis fails with its
// own name.
func TestBacklogLanding_ParityComparesEvidence(t *testing.T) {
	t.Parallel()

	if err := assertBacklogParity(landingRoundTripFixture(), landingRoundTripFixture()); err != nil {
		t.Fatalf("identical records reported a mismatch: %v", err)
	}

	other := LandingEvidence{
		Ref:        "origin/main",
		RefHead:    "4444444444444444444444444444444444444444",
		ObservedAt: "2026-01-03T00:00:00Z",
	}
	cases := map[string]func(*BacklogRecord){
		"live evidence dropped":     func(r *BacklogRecord) { r.Items[0].Landing = nil },
		"live evidence altered":     func(r *BacklogRecord) { r.Items[0].Landing = &other },
		"archived evidence dropped": func(r *BacklogRecord) { r.Archived[0].Item.Landing = nil },
		"archived evidence altered": func(r *BacklogRecord) { r.Archived[0].Item.Landing = &other },
	}
	for name, tamper := range cases {
		t.Run(name, func(t *testing.T) {
			tampered := landingRoundTripFixture()
			tamper(tampered)
			if err := assertBacklogParity(landingRoundTripFixture(), tampered); err == nil {
				t.Fatalf("parity check passed a %s — the cutover would have flipped authority onto a record missing the operator's evidence", name)
			}
		})
	}
}

// The gap M3 measured and left for M5: `landed -> done -> undone` silently
// lost the record, because writeArchive carried no landing column.
//
// M3's probe was a t.Logf that passed either way and was deleted for it. This
// is the assertion that discriminates: it fails while the archived write drops
// the column and passes once it carries it.
func TestBacklogLanding_SurvivesDoneAndUndone(t *testing.T) {
	store := archiveFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	recorded := LandingEvidence{
		Ref:        "origin/main",
		RefHead:    "5555555555555555555555555555555555555555",
		ObservedAt: "2026-01-04T00:00:00Z",
		SHA:        "6666666666666666666666666666666666666666",
		SHASource:  LandingSHASourceOperator,
		SpecStatus: "implemented",
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Items[0].Landing = &recorded
		return nil
	}); err != nil {
		t.Fatalf("record landing: %v", err)
	}

	// done
	if err := store.Mutate(func(rec *BacklogRecord) error {
		return rec.ArchiveCard("t1")
	}); err != nil {
		t.Fatalf("done: %v", err)
	}
	afterDone, err := store.Load()
	if err != nil {
		t.Fatalf("load after done: %v", err)
	}
	if len(afterDone.Archived) != 1 {
		t.Fatalf("archive holds %d entries after done, want 1", len(afterDone.Archived))
	}
	assertLandingEqual(t, "archived t1", &recorded, afterDone.Archived[0].Item.Landing)

	// undone
	if err := store.Mutate(func(rec *BacklogRecord) error {
		return rec.RestoreCard("t1")
	}); err != nil {
		t.Fatalf("undone: %v", err)
	}
	afterUndone, err := store.Load()
	if err != nil {
		t.Fatalf("load after undone: %v", err)
	}
	if len(afterUndone.Items) != 1 {
		t.Fatalf("queue holds %d live cards after undone, want 1", len(afterUndone.Items))
	}
	assertLandingEqual(t, "restored t1", &recorded, afterUndone.Items[0].Landing)
}

// assertLandingEqual compares by null-shape AND value, the same two-part test
// equalSpecID applies: a nil and a pointer to a zero record are different
// facts, and reporting only "not equal" would hide which of the two happened.
func assertLandingEqual(t *testing.T, what string, want, got *LandingEvidence) {
	t.Helper()
	switch {
	case want == nil && got == nil:
		return
	case want == nil:
		t.Errorf("%s: evidence invented — want none, got %+v", what, *got)
	case got == nil:
		t.Errorf("%s: evidence LOST — want %+v, got none", what, *want)
	case *want != *got:
		t.Errorf("%s: evidence altered — want %+v, got %+v", what, *want, *got)
	}
}
