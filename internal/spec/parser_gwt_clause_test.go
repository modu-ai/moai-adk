package spec

import "testing"

// TestParseAcceptanceCriteria_GWTClausesSurvive is the card t808 reproduction:
// `moai spec view` rendered only the Given clause of an EARS/GWT acceptance
// line, dropping When and Then. The renderer joins whatever the parser filled,
// so the clause loss is measured here at the parser boundary.
func TestParseAcceptanceCriteria_GWTClausesSurvive(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantGiven string
		wantWhen  string
		wantThen  string
	}{
		{
			name:      "full Given/When/Then",
			line:      "- AC-GWT-001-01: Given a SPEC, When it is parsed, Then the declaration is kept (maps REQ-GWT-001-001)",
			wantGiven: "Given a SPEC",
			wantWhen:  "When it is parsed",
			wantThen:  "Then the declaration is kept",
		},
		{
			name:      "Given/Then with no When",
			line:      "- AC-GWT-001-02: Given a SPEC, Then the declaration is kept (maps REQ-GWT-001-001)",
			wantGiven: "Given a SPEC",
			wantWhen:  "",
			wantThen:  "Then the declaration is kept",
		},
		{
			name:      "When/Then with no Given",
			line:      "- AC-GWT-001-03: When it is parsed, Then the declaration is kept (maps REQ-GWT-001-001)",
			wantGiven: "",
			wantWhen:  "When it is parsed",
			wantThen:  "Then the declaration is kept",
		},
		{
			name:      "Given only",
			line:      "- AC-GWT-001-04: Given a SPEC (maps REQ-GWT-001-001)",
			wantGiven: "Given a SPEC",
			wantWhen:  "",
			wantThen:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown := "## 5. Requirements (EARS)\n\n" +
				"- REQ-GWT-001-001: The system SHALL keep every clause.\n\n" +
				"## 6. Acceptance Criteria\n\n" + tt.line + "\n"

			criteria, errs := ParseAcceptanceCriteria(markdown, false)
			if len(errs) > 0 {
				t.Fatalf("unexpected parse errors: %v", errs)
			}
			if len(criteria) != 1 {
				t.Fatalf("want 1 acceptance criterion, got %d", len(criteria))
			}
			// A childless root is auto-wrapped (autoWrapSingle): the clauses move
			// to the ".a" child and the wrapper is left empty. The clauses the
			// renderer prints are the child's, so they are what is asserted here.
			if len(criteria[0].Children) != 1 {
				t.Fatalf("want 1 auto-wrapped child, got %d", len(criteria[0].Children))
			}

			got := criteria[0].Children[0]
			if got.Given != tt.wantGiven {
				t.Errorf("Given = %q, want %q", got.Given, tt.wantGiven)
			}
			if got.When != tt.wantWhen {
				t.Errorf("When = %q, want %q", got.When, tt.wantWhen)
			}
			if got.Then != tt.wantThen {
				t.Errorf("Then = %q, want %q", got.Then, tt.wantThen)
			}
		})
	}
}
