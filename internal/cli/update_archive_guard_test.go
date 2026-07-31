// SPEC-UPDATE-LEGACY-SKILL-LIST-001 / M2
// update_archive_guard_test.go — cross-check guard asserting that no entry of
// legacySkillIDs is also a live skill in the embedded template tree.
//
// Why this guard exists: every other test that touches legacySkillIDs is
// self-referential — it seeds synthetic skill directories FROM the list and
// then asserts against the list, so it passes whatever the list contains. That
// blind spot let three revived template skills (moai-domain-backend /
// -frontend / -database) sit on the "removed in BC-V3R3-007" list from
// 2026-04-28 onward, which made the v2.16 archive drift-check compare a
// freshly-redeployed skill against a frozen snapshot on every `moai update`.
// This guard is the first check that compares the list against the real
// embedded manifest instead of against itself.

package cli

import (
	"errors"
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// manifestVerdict is the classification of a (legacyIDs, embedded, err) triple.
type manifestVerdict int

const (
	// verdictCompare — the manifest is usable; the overlap is meaningful.
	verdictCompare manifestVerdict = iota
	// verdictSkipError — the manifest read failed; degrade rather than judge.
	verdictSkipError
	// verdictSkipEmpty — the manifest read succeeded but derived no names.
	// EmbeddedMoaiSkillNames' contract requires treating this as
	// "manifest unavailable" rather than as "no skills exist"; classifying it
	// as the latter would report every legacy ID as absent from the template
	// tree, i.e. a vacuous green.
	verdictSkipEmpty
)

// classifyManifest is a pure comparison over the manifest triple, factored out
// so the degradation branches are reachable from a test. The real embedded FS
// always reads successfully in a compiled binary, so the error and empty paths
// cannot be exercised through EmbeddedMoaiSkillNames itself.
//
// Returns the verdict and, for verdictCompare, the sorted intersection.
func classifyManifest(legacyIDs, embedded []string, err error) (manifestVerdict, []string) {
	if err != nil {
		return verdictSkipError, nil
	}
	if len(embedded) == 0 {
		return verdictSkipEmpty, nil
	}

	live := make(map[string]struct{}, len(embedded))
	for _, name := range embedded {
		live[name] = struct{}{}
	}

	var overlap []string
	for _, id := range legacyIDs {
		if _, ok := live[id]; ok {
			overlap = append(overlap, id)
		}
	}
	sort.Strings(overlap)

	return verdictCompare, overlap
}

// TestLegacySkillIDsNotEmbedded asserts legacySkillIDs is disjoint from the
// embedded template skill set.
//
// Test shape is pinned by SPEC-UPDATE-LEGACY-SKILL-LIST-001 plan.md §E M2: the
// production assertion lives directly in this parent body (NOT in a
// "production" subtest), because Go's -v output prints a --- PASS line for the
// parent and for every passing subtest, and the acceptance criteria count
// exactly one parent PASS line.
func TestLegacySkillIDsNotEmbedded(t *testing.T) {
	embedded, err := template.EmbeddedMoaiSkillNames()

	verdict, overlap := classifyManifest(legacySkillIDs, embedded, err)
	switch verdict {
	case verdictSkipError:
		t.Skipf("embedded skill manifest unavailable (read error), declining to judge: %v", err)
	case verdictSkipEmpty:
		t.Skip("embedded skill manifest derived zero names, declining to judge: an empty set means unavailable, not 'no skills exist'")
	}

	if len(overlap) > 0 {
		t.Errorf(
			"legacySkillIDs contains %d skill(s) that are still shipped in the embedded template tree: %v\n"+
				"An entry that is redeployed by every `moai update` is not a removed legacy skill. "+
				"Because its source directory is recreated on each update, archiveSkill never takes its "+
				"source-absent short-circuit, so the v2.16 drift-check re-fires forever. "+
				"Remove these from legacySkillIDs in update_archive.go.",
			len(overlap), overlap,
		)
	}

	// Degradation coverage. Both subtests drive classifyManifest with synthetic
	// inputs and end in t.Skip, so the SKIP marker is the observable evidence
	// that the branch declined to judge rather than silently passing.
	t.Run("manifest_error", func(t *testing.T) {
		got, ids := classifyManifest([]string{"fixture-alpha"}, nil, errors.New("synthetic manifest read failure"))
		if got != verdictSkipError {
			t.Fatalf("a non-nil manifest error must classify as verdictSkipError, got verdict %d (ids %v)", got, ids)
		}
		t.Skip("degradation verified: a manifest read error declines to judge")
	})

	t.Run("manifest_empty", func(t *testing.T) {
		got, ids := classifyManifest([]string{"fixture-alpha"}, []string{}, nil)
		if got != verdictSkipEmpty {
			t.Fatalf("an empty manifest must classify as verdictSkipEmpty, got verdict %d (ids %v)", got, ids)
		}
		t.Skip("degradation verified: an empty manifest declines to judge")
	})
}
