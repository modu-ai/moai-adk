package merge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	mrg "github.com/modu-ai/moai-adk/internal/merge"
)

// SPEC-UPDATE-MERGE-CONFLICT-BLIND-001 M1 — the reproduction.
//
// These tests measure the shared-key merge outcome rather than repairing it. The
// prediction for each cell is fixed here, before the measurement, so the
// measurement is free to disagree; a disagreement is the finding, not a harness
// bug.
//
// The fixture lives in a directory the test creates, and the base is derived the
// way the update path derives it (deriveTemplateBase over the user's content and
// the freshly deployed template) rather than hand-constructed, so what is
// measured is the real derivation.

// reproFixtureName is the fixture's file name. Only the extension is
// load-bearing — the strategy selector keys on ".json" and never on the name —
// so a neutral name keeps the fixture from reading as a real project file.
const reproFixtureName = "fixture.json"

// writeReproFixture writes the two fixture documents into an isolated temporary
// directory and reads them back, so the merge runs over bytes that made a
// round-trip through disk exactly as the update path's backup and redeploy do.
func writeReproFixture(t *testing.T, current, updated string) (currentData, updatedData []byte, path string) {
	t.Helper()

	dir := t.TempDir()
	currentPath := filepath.Join(dir, "current-"+reproFixtureName)
	updatedPath := filepath.Join(dir, "updated-"+reproFixtureName)

	if err := os.WriteFile(currentPath, []byte(current), defs.FilePerm); err != nil {
		t.Fatalf("write current fixture: %v", err)
	}
	if err := os.WriteFile(updatedPath, []byte(updated), defs.FilePerm); err != nil {
		t.Fatalf("write updated fixture: %v", err)
	}

	currentData, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("read current fixture: %v", err)
	}
	updatedData, err = os.ReadFile(updatedPath)
	if err != nil {
		t.Fatalf("read updated fixture: %v", err)
	}

	return currentData, updatedData, filepath.Join(dir, reproFixtureName)
}

// mergeReproFixture derives the base and runs the engine, returning the whole
// MergeResult. The conflict readings are taken from this return value; deriving
// the base again and comparing it to what it was derived from would assert
// nothing.
func mergeReproFixture(t *testing.T, current, updated string) (*mrg.MergeResult, map[string]any) {
	t.Helper()

	currentData, updatedData, path := writeReproFixture(t, current, updated)

	base, ok := deriveTemplateBase(path, currentData, updatedData)
	if !ok {
		t.Fatalf("no base derived for %s", path)
	}

	result, err := mrg.NewEngine().MergeFile(context.Background(), path, base, currentData, updatedData)
	if err != nil {
		t.Fatalf("MergeFile: %v", err)
	}

	var merged map[string]any
	if err := json.Unmarshal(result.Content, &merged); err != nil {
		t.Fatalf("parse merged content: %v", err)
	}

	return result, merged
}

// jsonOf renders a value in its JSON form, which is how the readings are
// recorded and compared: two values are the same reading when their JSON forms
// are byte-identical.
func jsonOf(t *testing.T, v any) string {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal reading: %v", err)
	}
	return string(data)
}

// TestSharedKeyControlCells measures the four control cells of REQ-UMC-005 in
// one run against one fixture. Cell "omitted" is the discriminator: the other
// three all predict the user's side standing, so a harness that merely copied
// the user's file would pass them.
func TestSharedKeyControlCells(t *testing.T) {
	const current = `{
  "untouched_shared": ["template-old"],
  "emptied_shared": [],
  "changed_shared": ["user-choice"],
  "user_only": ["user-addition"]
}`
	const updated = `{
  "untouched_shared": ["template-new"],
  "emptied_shared": ["template-a", "template-b"],
  "changed_shared": ["template-choice"],
  "omitted_key": ["template-only"]
}`

	result, merged := mergeReproFixture(t, current, updated)

	cells := []struct {
		// name is fixed by the SPEC's acceptance commands, which select on it.
		name string
		key  string
		// prediction is the JSON form of the value the merge is predicted to
		// write, fixed before the measurement.
		prediction string
		// rationale states why that value is predicted, so a disagreement is
		// legible without re-reading the SPEC.
		rationale string
	}{
		{
			name:       "untouched_shared",
			key:        "untouched_shared",
			prediction: `["template-old"]`,
			rationale:  "shared key: derived base carries the template's value, so the template's change reads as no change and the user's value stands",
		},
		{
			name:       "emptied_shared",
			key:        "emptied_shared",
			prediction: `[]`,
			rationale:  "shared key emptied by the user: the empty array is the user's change and is preserved",
		},
		{
			name:       "changed_shared",
			key:        "changed_shared",
			prediction: `["user-choice"]`,
			rationale:  "shared key changed by the user: the user's value is preserved",
		},
		{
			name:       "omitted",
			key:        "omitted_key",
			prediction: `["template-only"]`,
			rationale:  "key absent from the user's side is not shared, so it stays out of the base and the template reads as introducing it",
		},
	}

	for _, cell := range cells {
		t.Run(cell.name, func(t *testing.T) {
			written, present := merged[cell.key]
			if !present {
				t.Fatalf("key %q absent from merged output; predicted %s (%s)", cell.key, cell.prediction, cell.rationale)
			}

			observed := jsonOf(t, written)
			t.Logf("cell %s: key=%s written=%s HasConflict=%t len(Conflicts)=%d predicted=%s (%s)",
				cell.name, cell.key, observed, result.HasConflict, len(result.Conflicts), cell.prediction, cell.rationale)

			if observed != cell.prediction {
				t.Errorf("cell %s: merge wrote %s for %q, prediction was %s (%s)",
					cell.name, observed, cell.key, cell.prediction, cell.rationale)
			}
		})
	}
}

// TestConflictSurfaceReachability measures the conflict surface on a shared key
// whose two sides genuinely disagree. Both readings come from the merge's own
// return value.
//
// The measurement is paired with a control, because "no conflict was reported"
// and "this harness never reads a conflict" produce the same two readings. The
// control runs the same engine over the same fixture with a base that is not the
// derived one, so a conflict is reachable; a control that also reports no
// conflict would mean the readings are an instrument artifact and both are void.
func TestConflictSurfaceReachability(t *testing.T) {
	const current = `{
  "divergent_shared": "user-value"
}`
	const updated = `{
  "divergent_shared": "template-value"
}`

	t.Run("derived_base", func(t *testing.T) {
		result, merged := mergeReproFixture(t, current, updated)

		t.Logf("divergent shared key, base derived as the update path derives it: written=%s HasConflict=%t len(Conflicts)=%d (prediction: false, 0)",
			jsonOf(t, merged["divergent_shared"]), result.HasConflict, len(result.Conflicts))

		if result.HasConflict {
			t.Errorf("MergeResult.HasConflict = true on a divergent shared key; prediction was false")
		}
		if len(result.Conflicts) != 0 {
			t.Errorf("len(MergeResult.Conflicts) = %d on a divergent shared key; prediction was 0", len(result.Conflicts))
		}
	})

	t.Run("genuine_base_control", func(t *testing.T) {
		// A base that agrees with neither side is what the derived base cannot
		// be for a shared key, and it is the input under which the both-changed
		// arm can execute.
		const base = `{
  "divergent_shared": "base-value"
}`

		result, err := mrg.NewEngine().MergeFile(
			context.Background(),
			reproFixtureName,
			[]byte(base),
			[]byte(current),
			[]byte(updated),
		)
		if err != nil {
			t.Fatalf("MergeFile: %v", err)
		}

		t.Logf("same key and same engine, base differing from both sides: HasConflict=%t len(Conflicts)=%d (prediction: true, 1)",
			result.HasConflict, len(result.Conflicts))

		if !result.HasConflict || len(result.Conflicts) == 0 {
			t.Fatalf("control reported no conflict (HasConflict=%t len(Conflicts)=%d): the conflict surface is unread by this harness, so the derived_base readings measure nothing",
				result.HasConflict, len(result.Conflicts))
		}
	})
}
