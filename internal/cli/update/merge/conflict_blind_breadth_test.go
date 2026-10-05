package merge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/defs"
	mrg "github.com/modu-ai/moai-adk/internal/merge"
)

// SPEC-UPDATE-MERGE-CONFLICT-BLIND-001 M2.0 — the precondition measurement.
//
// M1 (conflict_blind_repro_test.go) measured a flat JSON document, so spec.md
// §A.6 is established for top-level JSON keys and for nothing else (spec.md
// §A.3). This file measures the two breadths M1 never entered, each on its own
// terms:
//
//	(A) the recursive pruneToShared path (base.go:124-126), down to a nested
//	    leaf — the real shape being permissions.ask inside permissions;
//	(B) the YAML path, which deriveTemplateBase dispatches through a different
//	    codec (base.go:75-78) and which the engine routes to mergeYAML.
//
// Neither breadth is inferred from the other. Every cell fixes its prediction
// before the run, every reading is taken from the MergeResult the engine
// returned, and each breadth carries its own discriminator cell — a cell
// predicted to come out the OTHER way — because the cells that predict "the
// user's side stands" are all satisfied by a harness that merely copied the
// user's file.
//
// No production code is changed here. The milestone's exit is the recorded
// evidence.

// breadthFixtureBase is the fixture's stem. Only the extension is load-bearing:
// deriveTemplateBase and the engine's strategy selector both key on the
// extension and never on the name.
const breadthFixtureBase = "fixture"

// writeBreadthFixture writes the two fixture documents into an isolated
// temporary directory under the requested extension and reads them back, so the
// merge runs over bytes that made a round-trip through disk exactly as the
// update path's backup and redeploy do.
func writeBreadthFixture(t *testing.T, ext, current, updated string) (currentData, updatedData []byte, path string) {
	t.Helper()

	dir := t.TempDir()
	name := breadthFixtureBase + ext
	currentPath := filepath.Join(dir, "current-"+name)
	updatedPath := filepath.Join(dir, "updated-"+name)

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

	return currentData, updatedData, filepath.Join(dir, name)
}

// decodeBreadth parses merged output with the codec the extension implies. The
// merged document is read back through the same format the strategy wrote it
// in, so a reading is never taken from a shape the merge did not produce.
func decodeBreadth(t *testing.T, ext string, data []byte) map[string]any {
	t.Helper()

	var out map[string]any
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("parse merged json: %v", err)
		}
	case ".yaml":
		if err := yaml.Unmarshal(data, &out); err != nil {
			t.Fatalf("parse merged yaml: %v", err)
		}
	default:
		t.Fatalf("decodeBreadth: unsupported extension %q", ext)
	}
	return out
}

// mergeBreadthFixture derives the base the way the update path derives it and
// runs the engine, returning the whole MergeResult. Conflict readings are taken
// from this return value; re-deriving the base and comparing it to what it was
// derived from would assert nothing.
func mergeBreadthFixture(t *testing.T, ext, current, updated string) (*mrg.MergeResult, map[string]any) {
	t.Helper()

	currentData, updatedData, path := writeBreadthFixture(t, ext, current, updated)

	base, ok := deriveTemplateBase(path, currentData, updatedData)
	if !ok {
		t.Fatalf("no base derived for %s", path)
	}
	t.Logf("derived base for %s:\n%s", filepath.Base(path), string(base))

	result, err := mrg.NewEngine().MergeFile(context.Background(), path, base, currentData, updatedData)
	if err != nil {
		t.Fatalf("MergeFile: %v", err)
	}

	return result, decodeBreadth(t, ext, result.Content)
}

// nestedMap reads a container out of a merged document, failing rather than
// silently reporting an absent container as an empty one.
func nestedMap(t *testing.T, doc map[string]any, key string) map[string]any {
	t.Helper()

	raw, present := doc[key]
	if !present {
		t.Fatalf("container %q absent from merged output", key)
	}
	child, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("container %q is %T, not a map", key, raw)
	}
	return child
}

// breadthCell is one control cell: a key, the JSON form of the value predicted
// for it, and the reason the prediction is what it is.
type breadthCell struct {
	name       string
	key        string
	prediction string
	rationale  string
}

// assertCells measures every cell against one merged container and logs the
// three readings per cell (value written, HasConflict, len(Conflicts)).
func assertCells(t *testing.T, label string, container map[string]any, result *mrg.MergeResult, cells []breadthCell) {
	t.Helper()

	for _, cell := range cells {
		t.Run(cell.name, func(t *testing.T) {
			written, present := container[cell.key]
			if !present {
				t.Fatalf("%s cell %s: key %q absent from merged output; predicted %s (%s)",
					label, cell.name, cell.key, cell.prediction, cell.rationale)
			}

			observed := jsonOf(t, written)
			t.Logf("%s cell %s: key=%s written=%s HasConflict=%t len(Conflicts)=%d strategy=%s predicted=%s (%s)",
				label, cell.name, cell.key, observed, result.HasConflict, len(result.Conflicts),
				result.Strategy, cell.prediction, cell.rationale)

			if observed != cell.prediction {
				t.Errorf("%s cell %s: merge wrote %s for %q, prediction was %s (%s)",
					label, cell.name, observed, cell.key, cell.prediction, cell.rationale)
			}
		})
	}
}

// nestedCells are the four REQ-UMC-005 cells restated one level down, plus the
// user-only leaf. Cell "omitted_leaf" is the discriminator: it is the only one
// predicted to end with the template's value, so a harness measuring nothing
// fails it.
func nestedCells() []breadthCell {
	return []breadthCell{
		{
			name:       "untouched_shared",
			key:        "untouched_shared",
			prediction: `["template-old"]`,
			rationale:  "shared nested leaf: the recursion puts the template's value in the base, so the template's change reads as no change and the user's value stands",
		},
		{
			name:       "emptied_shared",
			key:        "emptied_shared",
			prediction: `[]`,
			rationale:  "shared nested leaf emptied by the user: the empty array is the user's change and is preserved",
		},
		{
			name:       "changed_shared",
			key:        "changed_shared",
			prediction: `["user-choice"]`,
			rationale:  "shared nested leaf changed by the user: the user's value is preserved",
		},
		{
			name:       "omitted_leaf",
			key:        "omitted_leaf",
			prediction: `["template-only"]`,
			rationale:  "DISCRIMINATOR — leaf absent from the user's side is not shared, so the recursion leaves it out of the base and the template reads as introducing it",
		},
		{
			name:       "user_only_leaf",
			key:        "user_only",
			prediction: `["user-addition"]`,
			rationale:  "nested leaf only the user carries: preserved as the user's addition",
		},
	}
}

const (
	nestedCurrentJSON = `{
  "container": {
    "untouched_shared": ["template-old"],
    "emptied_shared": [],
    "changed_shared": ["user-choice"],
    "user_only": ["user-addition"]
  }
}`

	nestedUpdatedJSON = `{
  "container": {
    "untouched_shared": ["template-new"],
    "emptied_shared": ["template-a", "template-b"],
    "changed_shared": ["template-choice"],
    "omitted_leaf": ["template-only"]
  }
}`

	nestedCurrentYAML = `container:
  untouched_shared:
    - template-old
  emptied_shared: []
  changed_shared:
    - user-choice
  user_only:
    - user-addition
`

	nestedUpdatedYAML = `container:
  untouched_shared:
    - template-new
  emptied_shared:
    - template-a
    - template-b
  changed_shared:
    - template-choice
  omitted_leaf:
    - template-only
`

	flatCurrentYAML = `untouched_shared:
  - template-old
emptied_shared: []
changed_shared:
  - user-choice
user_only:
  - user-addition
`

	flatUpdatedYAML = `untouched_shared:
  - template-new
emptied_shared:
  - template-a
  - template-b
changed_shared:
  - template-choice
omitted_key:
  - template-only
`
)

// TestBreadthRecursiveJSONCells measures breadth (A): the four control cells one
// level down, inside a container both sides carry, driven through the JSON
// strategy. This is the shape permissions.ask has.
func TestBreadthRecursiveJSONCells(t *testing.T) {
	result, merged := mergeBreadthFixture(t, ".json", nestedCurrentJSON, nestedUpdatedJSON)

	container := nestedMap(t, merged, "container")
	assertCells(t, "recursive-json", container, result, nestedCells())

	// The container's own value is a separate reading from its leaves'. Logged
	// rather than asserted as a cell, because it is an observation about
	// granularity: a shared container key's value CAN change even where every
	// shared leaf inside it keeps the user's value.
	t.Logf("recursive-json container reading: written=%s HasConflict=%t len(Conflicts)=%d strategy=%s",
		jsonOf(t, merged["container"]), result.HasConflict, len(result.Conflicts), result.Strategy)
}

// TestBreadthRecursiveJSONOmittedContainer measures the second case the
// recursion makes possible and a flat document cannot: the container itself is
// absent from the user's side. Predicted to land whole, with its contents.
func TestBreadthRecursiveJSONOmittedContainer(t *testing.T) {
	const current = `{
  "kept": {"leaf": ["user-value"]}
}`
	const updated = `{
  "kept": {"leaf": ["user-value"]},
  "omitted_container": {
    "leaf_a": ["template-a"],
    "leaf_b": ["template-b"]
  }
}`

	result, merged := mergeBreadthFixture(t, ".json", current, updated)

	const prediction = `{"leaf_a":["template-a"],"leaf_b":["template-b"]}`
	raw, present := merged["omitted_container"]
	if !present {
		t.Fatalf("omitted_container absent from merged output; predicted %s", prediction)
	}
	observed := jsonOf(t, raw)
	t.Logf("recursive-json cell omitted_container: written=%s HasConflict=%t len(Conflicts)=%d strategy=%s predicted=%s (DISCRIMINATOR — container absent from the user's side is not shared, so the whole container reads as a template addition)",
		observed, result.HasConflict, len(result.Conflicts), result.Strategy, prediction)

	if observed != prediction {
		t.Errorf("recursive-json cell omitted_container: merge wrote %s, prediction was %s", observed, prediction)
	}
}

// TestBreadthRecursiveJSONConflictSurface measures the conflict surface one
// level down, on a nested shared leaf whose two sides genuinely disagree.
//
// Paired with a control, because "no conflict was reported" and "this harness
// never reads a conflict" produce the same two readings. The control runs the
// same engine over the same fixture with a base that differs from both sides at
// the nested leaf, so a conflict is reachable there.
func TestBreadthRecursiveJSONConflictSurface(t *testing.T) {
	const current = `{
  "container": {"divergent_leaf": "user-value"}
}`
	const updated = `{
  "container": {"divergent_leaf": "template-value"}
}`

	t.Run("derived_base", func(t *testing.T) {
		result, merged := mergeBreadthFixture(t, ".json", current, updated)
		container := nestedMap(t, merged, "container")

		t.Logf("recursive-json divergent nested leaf, base derived as the update path derives it: written=%s HasConflict=%t len(Conflicts)=%d strategy=%s (prediction: false, 0)",
			jsonOf(t, container["divergent_leaf"]), result.HasConflict, len(result.Conflicts), result.Strategy)

		if result.HasConflict {
			t.Errorf("MergeResult.HasConflict = true on a divergent nested shared leaf; prediction was false")
		}
		if len(result.Conflicts) != 0 {
			t.Errorf("len(MergeResult.Conflicts) = %d on a divergent nested shared leaf; prediction was 0", len(result.Conflicts))
		}
	})

	t.Run("genuine_base_control", func(t *testing.T) {
		const base = `{
  "container": {"divergent_leaf": "base-value"}
}`

		result, err := mrg.NewEngine().MergeFile(
			context.Background(),
			breadthFixtureBase+".json",
			[]byte(base),
			[]byte(current),
			[]byte(updated),
		)
		if err != nil {
			t.Fatalf("MergeFile: %v", err)
		}

		t.Logf("recursive-json control, same nested key and same engine, base differing from both sides at the leaf: HasConflict=%t len(Conflicts)=%d strategy=%s (prediction: true, 1)",
			result.HasConflict, len(result.Conflicts), result.Strategy)

		if !result.HasConflict || len(result.Conflicts) == 0 {
			t.Fatalf("control reported no conflict (HasConflict=%t len(Conflicts)=%d): the nested conflict surface is unread by this harness, so the derived_base readings measure nothing",
				result.HasConflict, len(result.Conflicts))
		}
	})
}

// TestBreadthYAMLFlatCells measures breadth (B): M1's four control cells
// restated as a YAML document and driven through the YAML strategy. The
// Strategy field is asserted, so the reading cannot be mistaken for a JSON one
// taken through a .yaml file name.
func TestBreadthYAMLFlatCells(t *testing.T) {
	result, merged := mergeBreadthFixture(t, ".yaml", flatCurrentYAML, flatUpdatedYAML)

	if result.Strategy != mrg.YAMLDeep {
		t.Fatalf("strategy = %s, want %s: the YAML path was not entered, so these readings say nothing about it", result.Strategy, mrg.YAMLDeep)
	}

	assertCells(t, "yaml-flat", merged, result, []breadthCell{
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
			rationale:  "shared key emptied by the user: the empty sequence is the user's change and is preserved",
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
			rationale:  "DISCRIMINATOR — key absent from the user's side is not shared, so it stays out of the base and the template reads as introducing it",
		},
		{
			name:       "user_only",
			key:        "user_only",
			prediction: `["user-addition"]`,
			rationale:  "key only the user carries: preserved as the user's addition",
		},
	})
}

// TestBreadthYAMLNestedCells measures the intersection of the two breadths: a
// nested leaf reached through the YAML codec. This is the shape
// .moai/config/sections/*.yaml actually has, so neither breadth alone covers it.
func TestBreadthYAMLNestedCells(t *testing.T) {
	result, merged := mergeBreadthFixture(t, ".yaml", nestedCurrentYAML, nestedUpdatedYAML)

	if result.Strategy != mrg.YAMLDeep {
		t.Fatalf("strategy = %s, want %s", result.Strategy, mrg.YAMLDeep)
	}

	container := nestedMap(t, merged, "container")
	assertCells(t, "yaml-nested", container, result, nestedCells())

	t.Logf("yaml-nested container reading: written=%s HasConflict=%t len(Conflicts)=%d strategy=%s",
		jsonOf(t, merged["container"]), result.HasConflict, len(result.Conflicts), result.Strategy)
}

// TestBreadthYAMLConflictSurface measures the YAML conflict surface on a shared
// key whose two sides genuinely disagree, paired with its own control on the
// same codec. The control is not inherited from the JSON breadth: a control
// inside one codec says nothing about the other.
func TestBreadthYAMLConflictSurface(t *testing.T) {
	const current = "divergent_shared: user-value\n"
	const updated = "divergent_shared: template-value\n"

	t.Run("derived_base", func(t *testing.T) {
		result, merged := mergeBreadthFixture(t, ".yaml", current, updated)

		if result.Strategy != mrg.YAMLDeep {
			t.Fatalf("strategy = %s, want %s", result.Strategy, mrg.YAMLDeep)
		}

		t.Logf("yaml divergent shared key, base derived as the update path derives it: written=%s HasConflict=%t len(Conflicts)=%d strategy=%s (prediction: false, 0)",
			jsonOf(t, merged["divergent_shared"]), result.HasConflict, len(result.Conflicts), result.Strategy)

		if result.HasConflict {
			t.Errorf("MergeResult.HasConflict = true on a divergent shared YAML key; prediction was false")
		}
		if len(result.Conflicts) != 0 {
			t.Errorf("len(MergeResult.Conflicts) = %d on a divergent shared YAML key; prediction was 0", len(result.Conflicts))
		}
	})

	t.Run("genuine_base_control", func(t *testing.T) {
		const base = "divergent_shared: base-value\n"

		result, err := mrg.NewEngine().MergeFile(
			context.Background(),
			breadthFixtureBase+".yaml",
			[]byte(base),
			[]byte(current),
			[]byte(updated),
		)
		if err != nil {
			t.Fatalf("MergeFile: %v", err)
		}

		t.Logf("yaml control, same key and same engine, base differing from both sides: HasConflict=%t len(Conflicts)=%d strategy=%s (prediction: true, 1)",
			result.HasConflict, len(result.Conflicts), result.Strategy)

		if !result.HasConflict || len(result.Conflicts) == 0 {
			t.Fatalf("control reported no conflict (HasConflict=%t len(Conflicts)=%d): the YAML conflict surface is unread by this harness, so the derived_base readings measure nothing",
				result.HasConflict, len(result.Conflicts))
		}
	})
}
