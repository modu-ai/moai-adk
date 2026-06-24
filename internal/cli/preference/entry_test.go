package preference

import (
	"testing"
	"time"
)

// TestEntry_SevenFieldsPresent (AC-ADM-003) verifies every stored entry carries
// the canonical 7-field schema: fact, source_citation, valid_time, last_used,
// scope, domain, confidence (plus the lookup key decision_key).
func TestEntry_SevenFieldsPresent(t *testing.T) {
	t.Parallel()

	e := Entry{
		Fact:           "prefers Go backend over Python",
		SourceCitation: "session:abc123",
		ValidTime:      mustParseRFC3339(t, "2026-06-20T10:00:00Z"),
		LastUsed:       mustParseRFC3339(t, "2026-06-24T10:00:00Z"),
		Scope:          ScopeStable,
		Domain:         "tech_stack",
		DecisionKey:    "backend_language",
		Confidence:     ConfidenceObserved,
	}

	if err := e.Validate(); err != nil {
		t.Fatalf("Entry.Validate() on fully-populated entry returned error: %v", err)
	}
}

// TestEntry_Validate_RejectsMissingFields (AC-ADM-003) verifies the schema
// validator rejects entries missing any of the 7 required fields or carrying
// an invalid enum value.
func TestEntry_Validate_RejectsMissingFields(t *testing.T) {
	t.Parallel()

	validTime := mustParseRFC3339(t, "2026-06-20T10:00:00Z")
	lastUsed := mustParseRFC3339(t, "2026-06-24T10:00:00Z")

	cases := []struct {
		name       string
		mutate     func(*Entry)
		wantSubstr string
	}{
		{"missing fact", func(e *Entry) { e.Fact = "" }, "fact"},
		{"missing source_citation", func(e *Entry) { e.SourceCitation = "" }, "source_citation"},
		{"missing valid_time (zero)", func(e *Entry) { e.ValidTime = time.Time{} }, "valid_time"},
		{"missing last_used (zero)", func(e *Entry) { e.LastUsed = time.Time{} }, "last_used"},
		{"missing domain", func(e *Entry) { e.Domain = "" }, "domain"},
		{"missing decision_key", func(e *Entry) { e.DecisionKey = "" }, "decision_key"},
		{"invalid scope", func(e *Entry) { e.Scope = Scope("bogus") }, "scope"},
		{"invalid confidence", func(e *Entry) { e.Confidence = Confidence("bogus") }, "confidence"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := Entry{
				Fact:           "prefers Go backend over Python",
				SourceCitation: "session:abc123",
				ValidTime:      validTime,
				LastUsed:       lastUsed,
				Scope:          ScopeStable,
				Domain:         "tech_stack",
				DecisionKey:    "backend_language",
				Confidence:     ConfidenceObserved,
			}
			tc.mutate(&e)
			err := e.Validate()
			if err == nil {
				t.Fatalf("Entry.Validate() with %s returned nil; want error containing %q", tc.name, tc.wantSubstr)
			}
		})
	}
}

// TestScope_Constants verifies the typed Scope enum has the two stable
// canonical values.
func TestScope_Constants(t *testing.T) {
	t.Parallel()

	if ScopeStable != "stable" {
		t.Errorf("ScopeStable = %q, want %q", ScopeStable, "stable")
	}
	if ScopeTransient != "transient" {
		t.Errorf("ScopeTransient = %q, want %q", ScopeTransient, "transient")
	}
}

// TestConfidence_Constants verifies the typed Confidence enum has the two
// stable canonical values (REQ-ADM-003; unverified inferences MUST tag
// `confidence: inferred`).
func TestConfidence_Constants(t *testing.T) {
	t.Parallel()

	if ConfidenceObserved != "observed" {
		t.Errorf("ConfidenceObserved = %q, want %q", ConfidenceObserved, "observed")
	}
	if ConfidenceInferred != "inferred" {
		t.Errorf("ConfidenceInferred = %q, want %q", ConfidenceInferred, "inferred")
	}
}

func mustParseRFC3339(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("failed to parse %q as RFC3339: %v", s, err)
	}
	return ts
}
