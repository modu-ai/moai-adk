package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// Card t2011: `todo relate <a> <b> --relation absorbs|contains` records the
// caller's direction. The pair is still deduplicated symmetrically
// (REQ-TCI-013: the opposite-order re-record maps onto the first record), but
// the stored subject/related are what the caller typed — a directional claim
// ("a absorbs b") is never stored reversed because of the card ids' order.
func TestTodoRelate_DirectionalRelationsKeepTheCallersDirection(t *testing.T) {
	for _, relation := range []string{"absorbs", "contains"} {
		t.Run(relation, func(t *testing.T) {
			_, store := todoFixture(t)
			seedItems(t, store, "Alpha card", "Beta card")

			// The larger id is the subject: the old normalization stored (t1, t2).
			if _, _, err := runTodo(t, "relate", "t2", "t1", "--relation", relation); err != nil {
				t.Fatalf("relate: %v", err)
			}
			got := loadFindings(t, store)
			if len(got) != 1 || got[0].SubjectID != "t2" || got[0].RelatedID != "t1" || got[0].Relation != relation {
				t.Fatalf("recorded %+v, want {subject t2, related t1, %s} exactly as typed", got, relation)
			}

			// The opposite order is still the same pair for dedup.
			_, _, err := runTodo(t, "relate", "t1", "t2", "--relation", relation)
			if err == nil || !strings.Contains(err.Error(), "already recorded") {
				t.Fatalf("opposite-order re-record: err = %v, want an already-recorded refusal", err)
			}
			if n := len(loadFindings(t, store)); n != 1 {
				t.Fatalf("findings = %d after the refused re-record, want 1", n)
			}
		})
	}
}

// Control: the truly symmetric kinds still normalize smaller-id-first.
func TestTodoRelate_SymmetricKindsStillNormalize(t *testing.T) {
	_, store := todoFixture(t)
	seedItems(t, store, "Alpha card", "Beta card")
	if _, _, err := runTodo(t, "relate", "t2", "t1", "--relation", factory.BacklogRelationConflicts); err != nil {
		t.Fatalf("relate: %v", err)
	}
	got := loadFindings(t, store)
	if len(got) != 1 || got[0].SubjectID != "t1" || got[0].RelatedID != "t2" {
		t.Fatalf("recorded %+v, want the conflicts pair normalized to (t1, t2)", got)
	}
}
