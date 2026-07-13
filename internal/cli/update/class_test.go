package update

import "testing"

// TestClassificationExhaustive asserts every input combination maps to exactly
// one ChangeClass (REQ-TUX3-001). The priority order is: user-owned > conflict
// > add(!exists) > update. No input produces "unknown".
func TestClassificationExhaustive(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		exists   bool
		conflict bool
		userOwn  bool
		want     ChangeClass
	}{
		{"user-owned new file preserved", "skills/hns-mine/SKILL.md", false, false, true, ClassPreserveUserOwned},
		{"user-owned existing preserved", "agents/harness/custom.md", true, false, true, ClassPreserveUserOwned},
		{"user-owned wins over conflict", "skills/my-tool/SKILL.md", true, true, true, ClassPreserveUserOwned},
		{"moai-managed new file added", ".claude/rules/moai/core/x.md", false, false, false, ClassAdd},
		{"moai-managed existing updated", ".claude/rules/moai/core/x.md", true, false, false, ClassUpdate},
		{"moai-managed conflict", ".claude/settings.json", true, true, false, ClassConflict},
		{"unmanaged non-existing added", "config/new-section.yaml", false, false, false, ClassAdd},
		{"nil predicate: existing update", ".claude/rules/x.md", true, false, false, ClassUpdate},
		{"nil predicate: non-existing add", ".claude/rules/x.md", false, false, false, ClassAdd},
		{"nil predicate: conflict", ".claude/rules/x.md", true, true, false, ClassConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pred UserOwnedPredicate
			if tt.userOwn || tt.name[:4] != "nil " {
				pred = func(rel string) bool { return tt.userOwn }
			}
			got := Classify(tt.relPath, tt.exists, tt.conflict, pred)
			if got != tt.want {
				t.Errorf("Classify(%q, exists=%v, conflict=%v, userOwn=%v) = %v (%s), want %v (%s)",
					tt.relPath, tt.exists, tt.conflict, tt.userOwn, got, got, tt.want, tt.want)
			}
			if got.String() == "unknown" {
				t.Errorf("Classify returned an unlabelled class for %q", tt.relPath)
			}
		})
	}
}

// TestChangeClassString pins the human-readable labels, including the
// "preserved (user-owned)" label surfaced in both the TUI table and the text
// fallback (REQ-TUX3-014).
func TestChangeClassString(t *testing.T) {
	want := map[ChangeClass]string{
		ClassAdd:               "add",
		ClassUpdate:            "update",
		ClassPreserveUserOwned: "preserved (user-owned)",
		ClassConflict:          "conflict",
	}
	for class, label := range want {
		if got := class.String(); got != label {
			t.Errorf("ChangeClass(%d).String() = %q, want %q", class, got, label)
		}
	}
}

// TestPreserveClassSource asserts the preserve classification is derived from
// the injected namespace-protection predicate (REQ-TUX3-002) — toggling the
// predicate result flips the classification for the SAME path. No parallel
// heuristic exists inside the classifier.
func TestPreserveClassSource(t *testing.T) {
	rel := "skills/hns-mine/SKILL.md"
	// Predicate says user-owned → preserve, even for a file that would otherwise
	// be a content update of an existing file.
	if got := Classify(rel, true, false, func(string) bool { return true }); got != ClassPreserveUserOwned {
		t.Errorf("predicate=true: Classify = %v (%s), want ClassPreserveUserOwned", got, got)
	}
	// Same path, predicate says NOT user-owned → ordinary update branch.
	if got := Classify(rel, true, false, func(string) bool { return false }); got != ClassUpdate {
		t.Errorf("predicate=false: Classify = %v (%s), want ClassUpdate", got, got)
	}
}

// TestPredicateShared documents the shared-source-of-truth contract: the
// classifier holds NO internal user-owned path list — it delegates entirely to
// the injected predicate. Two different predicates produce different
// classifications for the same path, proving no parallel heuristic.
func TestPredicateShared(t *testing.T) {
	rel := ".claude/agents/harness/my-specialist.md"
	conservative := func(r string) bool { return true }  // protect everything
	standard := func(r string) bool { return false }      // protect nothing
	if got := Classify(rel, true, false, conservative); got != ClassPreserveUserOwned {
		t.Errorf("conservative predicate: got %v, want ClassPreserveUserOwned", got)
	}
	if got := Classify(rel, true, false, standard); got != ClassUpdate {
		t.Errorf("standard predicate: got %v, want ClassUpdate", got)
	}
}
