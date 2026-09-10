// landing_evidence_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card t359) M2:
// the record type and the attribution boundary.
//
// AC-TLE-005 asserts all six facts survive the round trip, each compared to a
// distinct supplied or observed value, so an encoder that emits a subset
// cannot satisfy it. AC-TLE-006 asserts absence is SQL NULL at the column, not
// an empty object. AC-TLE-013 asserts the observed ref position and the
// operator-asserted delivering SHA occupy DIFFERENT keys and that the
// delivering key is absent rather than aliased to the head.
//
// What these tests do NOT assert is stated in the M2 report and in
// progress.md §E.2: the Givens of AC-TLE-012 and AC-TLE-013 are written in the
// verb's language ("the operator records a landing without --sha"), and the
// verb does not exist until M3. The records below are constructed directly,
// which is the strongest assertion M2 can carry and is weaker than the
// criterion as written.
package kanban

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// AC-TLE-005 — a decoded record carries all six facts, each equal to the
// supplied or observed value.
//
// Every field gets a value distinguishable from every other, so a field
// dropped by the encoder cannot be masked by a neighbour's value.
func TestLandingEvidence_CarriesAllSixFacts(t *testing.T) {
	want := LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    "e50964ad3f0000000000000000000000000000aa",
		ObservedAt: "2026-09-03T10:14:22Z",
		SHA:        "c9f712232a0000000000000000000000000000bb",
		SHASource:  LandingSHASourceOperator,
		SpecStatus: "completed",
	}

	encoded, err := EncodeLandingEvidence(want)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := DecodeLandingEvidence(encoded)
	if err != nil {
		t.Fatalf("decode %q: %v", encoded, err)
	}

	// Asserted field by field rather than by struct equality, so a failure
	// names WHICH fact the encoder dropped.
	for _, f := range []struct{ name, got, want string }{
		{"ref", got.Ref, want.Ref},
		{"ref_head", got.RefHead, want.RefHead},
		{"observed_at", got.ObservedAt, want.ObservedAt},
		{"sha", got.SHA, want.SHA},
		{"sha_source", got.SHASource, want.SHASource},
		{"spec_status", got.SpecStatus, want.SpecStatus},
	} {
		if f.got != f.want {
			t.Errorf("%s = %q, want %q", f.name, f.got, f.want)
		}
	}

	// The six facts must also be six KEYS on the wire. A struct-level round
	// trip alone would pass an encoder that folded two facts into one key and
	// split them again on decode.
	raw := unmarshalObject(t, encoded)
	for _, key := range []string{"ref", "ref_head", "observed_at", "sha", "sha_source", "spec_status"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("encoded record has no %q key: %s", key, encoded)
		}
	}
}

// AC-TLE-006 (column half) — a card that has never had a landing recorded
// stores SQL NULL, not an empty object and not an empty string.
//
// The render half of the criterion ("renders as absent") belongs to M4; this
// asserts the storage half only, and the M2 report says so.
func TestLandingEvidence_AbsenceIsSQLNull(t *testing.T) {
	store := archiveFixture(t)
	item, _, err := store.Add("a card with no landing evidence")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	t.Cleanup(func() {
		// Close failures on a temp-directory fixture are not the property
		// under test; the directory goes away with the test either way.
		if closeErr := eng.close(); closeErr != nil {
			t.Logf("close engine: %v", closeErr)
		}
	})

	var isNull int
	if err := eng.db.QueryRowContext(context.Background(),
		`SELECT landing IS NULL FROM items WHERE id = ?`, item.ID).Scan(&isNull); err != nil {
		t.Fatalf("read landing null-ness for %s: %v", item.ID, err)
	}
	if isNull != 1 {
		var stored string
		if scanErr := eng.db.QueryRowContext(context.Background(),
			`SELECT ifnull(landing, '<null>') FROM items WHERE id = ?`, item.ID).Scan(&stored); scanErr != nil {
			stored = "<unreadable: " + scanErr.Error() + ">"
		}
		t.Errorf("SELECT landing IS NULL = %d for %s, want 1; stored value is %q", isNull, item.ID, stored)
	}

	// The seam that keeps it NULL: a nil record must convert to a nil driver
	// value, never to "{}" and never to "". Asserted directly, because a
	// caller that encoded a zero-valued record would produce exactly the
	// stored shape REQ-TLE-006 forbids.
	value, err := LandingEvidenceValue(nil)
	if err != nil {
		t.Fatalf("value(nil): %v", err)
	}
	if value != nil {
		t.Errorf("LandingEvidenceValue(nil) = %#v, want nil (SQL NULL)", value)
	}
}

// AC-TLE-013 (key half) — the observed ref position is keyed distinctly from
// the operator-asserted delivering SHA, and the delivering key is ABSENT
// rather than aliased to the head when no SHA was supplied.
//
// The Given of the criterion as written ("a record produced without --sha")
// and the render half ("the rendered form labels it as a ref position") are
// verb-shaped and render-shaped respectively; both belong to M3/M4. This is
// the type-level assertion.
func TestLandingEvidence_RefHeadIsNotADeliveringSHA(t *testing.T) {
	const head = "e50964ad3f0000000000000000000000000000aa"
	rec := LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    head,
		ObservedAt: "2026-09-03T10:14:22Z",
	}
	encoded, err := EncodeLandingEvidence(rec)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	raw := unmarshalObject(t, encoded)

	// Two distinct keys, named by the implementation's own constants rather
	// than by transcription, so collapsing them in the source is what the
	// test sees.
	if LandingKeyRefHead == LandingKeyDeliveringSHA {
		t.Fatalf("the ref-position key and the delivering-SHA key are the same key (%q)", LandingKeyRefHead)
	}
	if _, ok := raw[LandingKeyRefHead]; !ok {
		t.Errorf("encoded record carries no %q key: %s", LandingKeyRefHead, encoded)
	}
	if _, ok := raw[LandingKeyDeliveringSHA]; ok {
		t.Errorf("no SHA was supplied, yet the record carries a %q key: %s", LandingKeyDeliveringSHA, encoded)
	}
	if _, ok := raw[LandingKeySHASource]; ok {
		t.Errorf("no SHA was supplied, yet the record carries a %q key: %s", LandingKeySHASource, encoded)
	}

	// And the head is not aliased into the delivering field on decode.
	got, err := DecodeLandingEvidence(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.SHA != "" {
		t.Errorf("decoded delivering sha = %q, want empty; the ref head is %q", got.SHA, head)
	}
	if got.RefHead != head {
		t.Errorf("decoded ref head = %q, want %q", got.RefHead, head)
	}

	// The label a reader keys on: with no operator SHA the record marks
	// itself as a ref position, not as a delivery.
	if marker := got.Marker(); marker != LandingMarkerRefHead {
		t.Errorf("marker = %q, want %q", marker, LandingMarkerRefHead)
	}
	withSHA := rec
	withSHA.SHA = "c9f712232a0000000000000000000000000000bb"
	withSHA.SHASource = LandingSHASourceOperator
	if marker := withSHA.Marker(); marker != LandingSHASourceOperator {
		t.Errorf("marker with an operator sha = %q, want %q", marker, LandingSHASourceOperator)
	}
}

// The provenance pairing is structural, not conventional: a delivering SHA
// with no operator provenance, or a provenance with no SHA, is refused by the
// encoder rather than stored. REQ-TLE-012 says a stored SHA is
// operator-supplied or absent; this is the seam that makes "or absent" the
// only other reachable state.
func TestLandingEvidence_RefusesUnpairedProvenance(t *testing.T) {
	base := LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    "e50964ad3f0000000000000000000000000000aa",
		ObservedAt: "2026-09-03T10:14:22Z",
	}
	cases := map[string]func(LandingEvidence) LandingEvidence{
		"sha without provenance": func(r LandingEvidence) LandingEvidence {
			r.SHA = "c9f712232a0000000000000000000000000000bb"
			return r
		},
		"provenance without sha": func(r LandingEvidence) LandingEvidence {
			r.SHASource = LandingSHASourceOperator
			return r
		},
		"sha with a non-operator provenance": func(r LandingEvidence) LandingEvidence {
			r.SHA = "c9f712232a0000000000000000000000000000bb"
			r.SHASource = "grep"
			return r
		},
		"no ref": func(r LandingEvidence) LandingEvidence {
			r.Ref = ""
			return r
		},
		"no ref head": func(r LandingEvidence) LandingEvidence {
			r.RefHead = ""
			return r
		},
		"no observation instant": func(r LandingEvidence) LandingEvidence {
			r.ObservedAt = ""
			return r
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := EncodeLandingEvidence(mutate(base)); err == nil {
				t.Errorf("EncodeLandingEvidence(%s) = nil error, want a refusal", name)
			}
		})
	}
}

// An unreadable SPEC status is an explicit marker, never an omitted key: the
// two carry different facts (read and found nothing / never asked), and an
// omitted key cannot distinguish them (REQ-TLE-010, design.md §1).
func TestLandingEvidence_UnknownSpecStatusIsExplicit(t *testing.T) {
	rec := LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    "e50964ad3f0000000000000000000000000000aa",
		ObservedAt: "2026-09-03T10:14:22Z",
		SpecStatus: LandingSpecStatusUnknown,
	}
	encoded, err := EncodeLandingEvidence(rec)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, ok := unmarshalObject(t, encoded)[LandingKeySpecStatus]
	if !ok {
		t.Fatalf("the unknown marker was omitted rather than stored: %s", encoded)
	}
	if strings.TrimSpace(string(got)) != `"`+LandingSpecStatusUnknown+`"` {
		t.Errorf("%s = %s, want %q", LandingKeySpecStatus, got, LandingSpecStatusUnknown)
	}

	// A card with no spec_id asks nothing, and omits the key. That is the
	// other side of the distinction the explicit marker exists to preserve.
	unasked := rec
	unasked.SpecStatus = ""
	encodedUnasked, err := EncodeLandingEvidence(unasked)
	if err != nil {
		t.Fatalf("encode without a spec status: %v", err)
	}
	if _, present := unmarshalObject(t, encodedUnasked)[LandingKeySpecStatus]; present {
		t.Errorf("a card that was never asked carries a %q key: %s", LandingKeySpecStatus, encodedUnasked)
	}
}

// A stored value that does not decode is an ERROR, never a zero-valued record
// silently standing in for one. A record that reads as empty would render as
// "observed nothing" and be indistinguishable from a genuine observation with
// blank fields.
func TestLandingEvidence_DecodeRefusesGarbage(t *testing.T) {
	for _, bad := range []string{"", "   ", "{", "null", `"origin/develop"`, "{}"} {
		if _, err := DecodeLandingEvidence(bad); err == nil {
			t.Errorf("DecodeLandingEvidence(%q) = nil error, want a refusal", bad)
		}
	}
}

// unmarshalObject reads an encoded record as a raw key set, so a test can ask
// which KEYS are present rather than which struct fields decoded — the two are
// different questions and only the first catches a collapsed key.
func unmarshalObject(t *testing.T, encoded string) map[string]json.RawMessage {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(encoded), &raw); err != nil {
		t.Fatalf("unmarshal %q: %v", encoded, err)
	}
	return raw
}
