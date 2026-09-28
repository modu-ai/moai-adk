package closure

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestReportSectionOrder pins REQ-CLOSURE-002: the JSON top-level keys appear
// in the fixed section order, and schema_version is 1.
func TestReportSectionOrder(t *testing.T) {
	r := NewReport("c1", "SPEC-EXAMPLE-001", "abc1234", "contract", "required")
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var keys []string
	// json.Unmarshal cannot give key order; decode raw.
	dec := json.NewDecoder(strings.NewReader(string(data)))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		t.Fatalf("expected object start, got %v %v", tok, err)
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("token: %v", err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("expected string key, got %v", tok)
		}
		keys = append(keys, key)
		// Skip the value.
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("decode value: %v", err)
		}
	}

	want := []string{
		"schema_version", "card", "spec_id", "head_sha", "generated_at",
		"mode", "second_review_policy",
		"summary", "kickoff", "reconciliation", "invariants", "ownership",
		"new_apis", "escalations", "first_verdict", "second_verdict",
		"plan_audit_binding", "not_performed", "residual_risk",
		"human_verdict", "sources",
	}
	if len(keys) != len(want) {
		t.Fatalf("top-level key count = %d, want %d (keys: %v)", len(keys), len(want), keys)
	}
	for i, k := range want {
		if keys[i] != k {
			t.Fatalf("key[%d] = %q, want %q (full order: %v)", i, keys[i], k, keys)
		}
	}
	if r.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", r.SchemaVersion)
	}
}

// TestNotPerformedCatalogueClosed pins design.md §B: the token set is closed
// and every token is a non-empty identifier.
func TestNotPerformedCatalogueClosed(t *testing.T) {
	want := []string{
		"progress-missing",
		"ac-not-reported",
		"first-verdict-missing",
		"second-review-not-performed",
		"second-review-stale",
		"receipt-missing",
		"receipt-field-unrecognized",
		"plan-audit-self-reported",
		"plan-audit-mismatch",
		"escalation-unreadable",
		"escalation-detector-not-armed",
		"invariant-not-observed",
		"new-api-comparison-unavailable",
		"human-verdict-none",
		"human-verdict-stale",
	}
	if got := NotPerformedTokens(); len(got) != len(want) {
		t.Fatalf("NotPerformedTokens length = %d, want %d: %v", len(got), len(want), got)
	}
	seen := map[string]bool{}
	for _, tok := range NotPerformedTokens() {
		if tok == "" || strings.ContainsAny(tok, " \t\n") {
			t.Fatalf("token %q is empty or contains whitespace", tok)
		}
		if seen[tok] {
			t.Fatalf("duplicate token %q", tok)
		}
		seen[tok] = true
	}
	for _, w := range want {
		if !seen[w] {
			t.Fatalf("token %q missing from catalogue: %v", w, seen)
		}
	}
	if !IsNotPerformedToken("progress-missing") {
		t.Fatalf("IsNotPerformedToken(progress-missing) = false, want true")
	}
	if IsNotPerformedToken("made-up-token") {
		t.Fatalf("IsNotPerformedToken(made-up-token) = true, want false")
	}
}

// TestAddNotPerformed pins that AddNotPerformed appends {item, detail} pairs
// and that only catalogue tokens are accepted (an unknown token is a builder
// bug and must fail loudly at build time, not render).
func TestAddNotPerformed(t *testing.T) {
	r := NewReport("c1", "SPEC-EXAMPLE-001", "abc", "guided", "required")
	r.AddNotPerformed(NotPerformedProgressMissing, "no progress.md")
	if len(r.NotPerformed) != 1 {
		t.Fatalf("NotPerformed length = %d, want 1", len(r.NotPerformed))
	}
	if r.NotPerformed[0].Item != NotPerformedProgressMissing || r.NotPerformed[0].Detail != "no progress.md" {
		t.Fatalf("entry = %+v, want {progress-missing, no progress.md}", r.NotPerformed[0])
	}
	if r.NotPerformed == nil {
		t.Fatalf("NotPerformed must be non-nil so it marshals as []")
	}
}
