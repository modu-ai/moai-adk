package guardliveness

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// The additive Expectation field on Entry is a CROSS-SPEC write: this package
// belongs to SPEC-GUARD-LIVENESS-001 (card t333, completed) and the field is
// added by SPEC-GUARD-STATE-MODEL-001 M3 (card t347) to discharge AC-GSM-007
// clause (d) — a row-4 STALE entry must carry the expectation it missed.
//
// The tests below are the questions an additive field on a landed type has to
// answer, and none of them is the compile question. Every existing literal in
// this package is KEYED (measured: 22 of 22), so the field compiles without
// touching a single site — that is the easy half and it establishes nothing
// about behaviour.

// TestEntryExpectationZeroValueIsAbsentNotEmpty pins the meaning of the zero
// value.
//
// An entry that is not row 4 carries no missed expectation, and "no expectation
// was missed" must not be readable as "the expectation was empty". A value
// struct cannot express that difference: its zero value is
// {Window:"", Measure:""}, which is indistinguishable from a carried
// expectation whose two fields happen to be blank. A nil POINTER can, and it is
// the representation this file already uses one type over for exactly this
// distinction (Result.Clean: "a nil pointer is an absent designation — distinct
// from a present-but-null one").
func TestEntryExpectationZeroValueIsAbsentNotEmpty(t *testing.T) {
	t.Parallel()

	var zero Entry
	if zero.Expectation != nil {
		t.Fatalf("a zero Entry must carry no expectation, got %+v", zero.Expectation)
	}

	// The distinction the pointer buys: a carried-but-blank expectation is a
	// DIFFERENT state from an absent one, and both are representable.
	blank := Entry{Expectation: &Expectation{}}
	if blank.Expectation == nil {
		t.Fatal("a carried expectation with blank fields must not read as absent")
	}
}

// TestEntryExpectationSurvivesTheStoreRoundTrip is the JSON-tag test, and it is
// stated over the STORE rather than over json.Marshal because that is the path
// the field actually travels: the refresh that produces a result and the LATER
// activation that renders it are different activations, and store.go is what
// carries it between them. A field that serializes wrongly arrives at the
// consumer as a contract violation that never happened.
//
// AC-GSM-015's Given was widened at v0.9.0 to bind exactly this.
func TestEntryExpectationSurvivesTheStoreRoundTrip(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	const root = "/some/tree"

	original := Result{
		Clean: &Designation{Values: []string{"alpha"}},
		Entries: []Entry{
			{
				Subject:         "subject-carrying",
				Classifications: []string{"beta"},
				Surface:         "unsettled",
				Expectation:     &Expectation{Window: "14d", Measure: "fired-at-all"},
			},
			{
				Subject:         "subject-not-carrying",
				Classifications: []string{"alpha"},
				Surface:         "settled",
			},
		},
	}

	if err := store.Save(root, original, time.Now()); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := store.Load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	got := loaded.Result.Entries
	if len(got) != 2 {
		t.Fatalf("round trip returned %d entries, want 2", len(got))
	}

	if got[0].Expectation == nil {
		t.Fatal("the carried expectation did not survive the round trip: nil after load")
	}
	if got[0].Expectation.Window != "14d" || got[0].Expectation.Measure != "fired-at-all" {
		t.Fatalf("the carried expectation changed across the round trip: %+v", *got[0].Expectation)
	}
	if got[1].Expectation != nil {
		t.Fatalf("an absent expectation became present across the round trip: %+v", *got[1].Expectation)
	}
}

// TestEntryExpectationAbsentFromAnOlderPersistedResult is the unfilled-path
// question, and it is the one an additive field is most likely to get wrong in
// production rather than in a test.
//
// The store holds results written BEFORE the field existed. Their JSON carries
// no `expectation` key at all. Decoding one must leave the field nil — "this
// entry carries no missed expectation" — rather than failing to decode or
// producing a present-but-blank expectation a reader would take for a declared
// one.
func TestEntryExpectationAbsentFromAnOlderPersistedResult(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	const root = "/some/tree"

	// Written by hand in the pre-field shape rather than by marshalling the
	// current type, which would silently acquire whatever the current tags say
	// and test nothing.
	legacy := `{"taken_at":"2026-08-01T00:00:00Z","result":{` +
		`"entries":[{"subject":"legacy-subject","classifications":["alpha"],"surface":"settled"}],` +
		`"clean":{"values":["alpha"]}}}`
	if err := os.WriteFile(store.pathFor(root, resultSuffix), []byte(legacy), 0o600); err != nil {
		t.Fatalf("plant the legacy snapshot: %v", err)
	}

	loaded, err := store.Load(root)
	if err != nil {
		t.Fatalf("a persisted result written before the field existed must still decode: %v", err)
	}
	if len(loaded.Result.Entries) != 1 {
		t.Fatalf("decoded %d entries, want 1", len(loaded.Result.Entries))
	}
	if loaded.Result.Entries[0].Expectation != nil {
		t.Fatalf("an older result acquired an expectation it never carried: %+v",
			*loaded.Result.Entries[0].Expectation)
	}

	// And it still partitions: the additive field changes no contract clause.
	if _, err := loaded.Result.Partition(); err != nil {
		t.Fatalf("partitioning an older persisted result: %v", err)
	}
}

// TestExpectationJSONShapeIsExplicit pins the wire shape itself. The field is
// carried as an explicit null rather than omitted when absent, so a reader of
// the persisted file sees "this entry was considered and carries none" instead
// of having to infer it from a missing key.
func TestExpectationJSONShapeIsExplicit(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(Entry{Subject: "s", Classifications: []string{"alpha"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	v, present := decoded["expectation"]
	if !present {
		t.Fatalf("the expectation key is absent from the wire shape: %s", payload)
	}
	if v != nil {
		t.Fatalf("an absent expectation must serialize as null, got %v", v)
	}
}
