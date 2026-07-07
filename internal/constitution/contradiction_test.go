package constitution

import (
	"testing"
)

// TestNewContradictionDetector verifies the constructor.
func TestNewContradictionDetector(t *testing.T) {
	d := NewContradictionDetector()
	if d == nil {
		t.Fatal("NewContradictionDetector returned nil")
	}
	var _ ContradictionDetector = d
}

// TestContradictionDetector_Scan_NoConflicts verifies the clean path.
func TestContradictionDetector_Scan_NoConflicts(t *testing.T) {
	d := NewContradictionDetector()
	registry := &Registry{
		Entries: []Rule{
			{ID: "CONST-V3R2-001", Zone: ZoneFrozen, Clause: "unrelated rule"},
		},
	}
	result, err := d.Scan(&AmendmentProposal{
		RuleID: "CONST-V3R2-099",
		Before: "MUST do A",
		After:  "MUST do B",
	}, registry)
	if err != nil {
		t.Fatalf("Scan clean: %v", err)
	}
	if result.HasBlockingContradiction {
		t.Error("should have no blocking contradiction")
	}
}

// TestContradictionDetector_Scan_ZoneRelaxation verifies Frozen constraint relaxation blocks.
func TestContradictionDetector_Scan_ZoneRelaxation(t *testing.T) {
	d := NewContradictionDetector()
	registry := &Registry{
		Entries: []Rule{
			{ID: "CONST-V3R2-001", Zone: ZoneFrozen, Clause: "MUST enforce X"},
		},
	}
	_, err := d.Scan(&AmendmentProposal{
		RuleID: "CONST-V3R2-099",
		Before: "MUST enforce X",
		After:  "MAY optionally enforce X",
	}, registry)
	if err == nil {
		t.Error("Scan should block on Frozen constraint relaxation")
	}
}

// TestContradictionDetector_Scan_ClauseMustNot verifies MUST NOT vs MUST blocks.
func TestContradictionDetector_Scan_ClauseMustNot(t *testing.T) {
	d := NewContradictionDetector()
	registry := &Registry{
		Entries: []Rule{
			{ID: "CONST-V3R2-001", Zone: ZoneEvolvable, Clause: "MUST allow feature X"},
		},
	}
	_, err := d.Scan(&AmendmentProposal{
		RuleID: "CONST-V3R2-099",
		Before: "MUST allow feature X",
		After:  "MUST NOT allow feature X",
	}, registry)
	if err == nil {
		t.Error("Scan should block on MUST NOT vs MUST contradiction")
	}
}

// TestContainsSequence table-drives the consecutive-word-sequence finder.
func TestContainsSequence(t *testing.T) {
	tests := []struct {
		name  string
		words []string
		seq   []string
		want  bool
	}{
		{"match at start", []string{"MUST", "NOT", "do"}, []string{"MUST", "NOT"}, true},
		{"match in middle", []string{"a", "MUST", "NOT", "b"}, []string{"MUST", "NOT"}, true},
		{"no match", []string{"MUST", "do"}, []string{"MUST", "NOT"}, false},
		{"single word match", []string{"MUST", "X"}, []string{"MUST"}, true},
		{"seq longer than words", []string{"MUST"}, []string{"MUST", "NOT"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsSequence(tt.words, tt.seq...); got != tt.want {
				t.Errorf("containsSequence(%v, %v) = %v, want %v", tt.words, tt.seq, got, tt.want)
			}
		})
	}
}

// TestContainsWord table-drives the word-in-slice finder.
func TestContainsWord(t *testing.T) {
	words := []string{"MUST", "SHALL", "REQUIRED"}
	if !containsWord(words, "MUST") {
		t.Error("containsWord MUST: want true")
	}
	if containsWord(words, "NEVER") {
		t.Error("containsWord NEVER: want false")
	}
}

// TestExtractAction verifies quoted-content extraction.
func TestExtractAction(t *testing.T) {
	tests := []struct {
		clause string
		want   string
	}{
		{`always "feature X"`, "feature X"},
		{`no quotes here`, ""},
		{`"quoted"`, "quoted"},
	}
	for _, tt := range tests {
		if got := extractAction(tt.clause); got != tt.want {
			t.Errorf("extractAction(%q) = %q, want %q", tt.clause, got, tt.want)
		}
	}
}

// TestExtractModifier verifies modifier extraction.
func TestExtractModifier(t *testing.T) {
	tests := []struct {
		clause string
		want   string
	}{
		{"MUST do X", "MUST"},
		{"MUST NOT do X", "MUST NOT"},
		{"SHALL NOT do X", "SHALL NOT"},
		{"SHALL do X", "SHALL"},
		{"MAY do X", "MAY"},
		{"no modifier here", ""},
	}
	for _, tt := range tests {
		if got := extractModifier(tt.clause); got != tt.want {
			t.Errorf("extractModifier(%q) = %q, want %q", tt.clause, got, tt.want)
		}
	}
}

// TestIsOppositeModifier table-drives the opposite-modifier checker.
func TestIsOppositeModifier(t *testing.T) {
	tests := []struct {
		mod1, mod2 string
		want       bool
	}{
		{"MUST", "MUST NOT", true},
		{"MUST NOT", "MUST", true},
		{"MUST", "SHALL", false},
		{"NEVER", "ALWAYS", true},
		{"ALWAYS", "NEVER", true},
		{"MAY", "MUST", false},
	}
	for _, tt := range tests {
		if got := isOppositeModifier(tt.mod1, tt.mod2); got != tt.want {
			t.Errorf("isOppositeModifier(%q, %q) = %v, want %v", tt.mod1, tt.mod2, got, tt.want)
		}
	}
}
