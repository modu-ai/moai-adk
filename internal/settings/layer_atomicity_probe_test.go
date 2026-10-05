package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// Card t1049, deliverable (2) — per-layer atomicity of the three persistence
// layers handleSave writes through.
//
// THE QUESTION. handleSave's nine steps stop at the first failure, so a
// mid-failure leaves the earlier steps on disk (measured in
// internal/web/partial_apply_repro_test.go). This file asks the same question
// one level down: WITHIN a single layer, does a mid-failure leave part of that
// layer's own work on disk?
//
// HOW A FAILURE IS INJECTED. The target path is replaced by a NON-EMPTY
// DIRECTORY. A read-only file is NOT sufficient: two of the three layers write
// via temp-file + rename, and renaming over a read-only file succeeds as long
// as the containing directory is writable — the probe would silently measure a
// success. A directory blocks both the read (EISDIR) and the rename.
//
// WHAT IS NOT MEASURED. Crash- and power-loss atomicity. These probes inject a
// filesystem error, which is the failure mode handleSave actually surfaces;
// durability under abrupt termination is a different question and is not
// answered here.

// blockPath replaces path with a non-empty directory so any write to it fails.
func blockPath(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("clear %s: %v", path, err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("block %s: %v", path, err)
	}
	// Non-empty: an empty directory can be replaced by rename on some systems.
	if err := os.WriteFile(filepath.Join(path, "occupant"), []byte("x"), 0o644); err != nil {
		t.Fatalf("occupy %s: %v", path, err)
	}
}

// TestLayer1ProfileStoreAtomicity — layer 1 (profile store, handleSave step 1).
//
// WritePreferences performs a single os.WriteFile (preferences.go), so there is
// no second file for a failure to land between: across-files atomicity is
// vacuous here. The open question is the other one — does a FAILED write
// destroy what was already on disk?
func TestLayer1ProfileStoreAtomicity(t *testing.T) {
	// Not parallel: BaseDirOverride is package state.
	base := t.TempDir()
	prev := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = prev })

	first := profile.ProfilePreferences{UserName: "FIRST-WRITE"}
	if err := profile.WritePreferences("default", first); err != nil {
		t.Fatalf("seed write: %v", err)
	}
	path := profile.GetPreferencesPath("default")
	seeded, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read seeded: %v", err)
	}

	// Positive control: an unblocked second write must actually land, otherwise
	// the "preserved" result below could come from a probe that never writes.
	if err := profile.WritePreferences("default", profile.ProfilePreferences{UserName: "SECOND-WRITE"}); err != nil {
		t.Fatalf("control write: %v", err)
	}
	if after, _ := os.ReadFile(path); string(after) == string(seeded) {
		t.Fatal("control: second write did not change the file — probe cannot discriminate")
	}

	// Re-seed, then block the target and write again.
	if err := profile.WritePreferences("default", first); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
	seeded, _ = os.ReadFile(path)
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	writeErr := profile.WritePreferences("default", profile.ProfilePreferences{UserName: "BLOCKED-WRITE"})
	if writeErr == nil {
		t.Skip("write to a 0444 file succeeded — running as root or on a permissive filesystem; probe inapplicable")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after blocked write: %v", err)
	}
	if string(after) != string(seeded) {
		t.Errorf("failed write damaged prior content:\nbefore=%q\nafter =%q", seeded, after)
	}
	t.Logf("L1 profile store: single-file write; a failed write left prior content intact (err=%v)", writeErr)
}

// TestLayer2ConfigManagerAtomicity — layer 2 (config manager, handleSave steps
// 3/4/5/6/8 all converge here).
//
// Save() writes six section files in sequence, each via temp+rename, returning
// on the first error. The question is whether the sections written before the
// failure stay on disk.
func TestLayer2ConfigManagerAtomicity(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}

	mgr := config.NewConfigManager()
	if _, err := mgr.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	// user.yaml is written before language.yaml in Save's sequence.
	if err := mgr.SetSection("user", models.UserConfig{Name: "PROBE-USER"}); err != nil {
		t.Fatalf("SetSection(user): %v", err)
	}

	// Positive control: an unblocked Save must produce user.yaml.
	if err := mgr.Save(); err != nil {
		t.Fatalf("control Save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sections, "user.yaml")); err != nil {
		t.Fatalf("control: user.yaml absent after a clean Save — probe cannot discriminate: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(sections, "user.yaml")); err != nil {
		t.Fatal(err)
	}

	// Block a section that Save reaches AFTER user.yaml.
	blockPath(t, filepath.Join(sections, "language.yaml"))

	saveErr := mgr.Save()
	if saveErr == nil {
		t.Skip("Save succeeded with a blocked section file — probe inapplicable on this filesystem")
	}
	_, statErr := os.Stat(filepath.Join(sections, "user.yaml"))
	partial := statErr == nil
	t.Logf("L2 config manager: Save failed (%v); an earlier section was already on disk: %v", saveErr, partial)
	if !partial {
		t.Error("expected the pre-failure section to be on disk (Save is documented as sequential per-file writes)")
	}
}

// TestLayer3YamlpatchAtomicity — layer 3 (yamlpatch seam, handleSave step 6's
// seam half).
//
// PatchFile writes one file via temp+rename, so a single call is atomic.
// ApplySchemaEdits loops over sections and returns on the first error, so the
// question is whether the earlier sections in that loop stay patched.
func TestLayer3YamlpatchAtomicity(t *testing.T) {
	t.Parallel()

	// Pick two seam-persisted bool fields in DIFFERENT sections, so the loop
	// has two files to walk. Selected at runtime from the schema rather than
	// hardcoded, so a schema rename does not silently empty this probe.
	var firstField, secondField FieldDef
	seen := map[SectionID]bool{}
	for _, f := range AllFields() {
		// PersistSeam in the schema is NOT the same as seam-writable: some
		// sections keep their seam FieldDefs while RouteForSection excludes
		// them (measured: "harness" is PersistSeam yet rejects the write with
		// `section "harness" is not seam-writable`). Filter on the route, which
		// is what ApplySchemaEdits actually consults.
		if f.Persist.Kind != PersistSeam || f.Type != TypeBool || seen[f.Section] {
			continue
		}
		if RouteForSection(f.Persist.Section) != RouteSeam {
			continue
		}
		seen[f.Section] = true
		if firstField.Name == "" {
			firstField = f
			continue
		}
		secondField = f
		break
	}
	if firstField.Name == "" || secondField.Name == "" {
		t.Fatalf("schema yielded fewer than two seam bool sections (%q / %q) — probe would be vacuous",
			firstField.Name, secondField.Name)
	}
	// The loop walks sections in sorted order; block whichever runs second.
	blocked, survivor := secondField, firstField
	if secondField.Persist.Section < firstField.Persist.Section {
		blocked, survivor = firstField, secondField
	}

	root := t.TempDir()
	seedSectionFixture(t, root, survivor.Persist.Section)
	seedSectionFixture(t, root, blocked.Persist.Section)
	before := readSection(t, root, survivor.Persist.Section)

	// The survivor's edit must actually flip its stored value, or the control
	// below cannot tell "patched" from "untouched". Try both bool values and
	// keep whichever changes the file.
	var edits map[string]string
	for _, v := range []string{"true", "false"} {
		probe := t.TempDir()
		seedSectionFixture(t, probe, survivor.Persist.Section)
		seedSectionFixture(t, probe, blocked.Persist.Section)
		candidate := map[string]string{survivor.Name: v, blocked.Name: v}
		if err := ApplySchemaEdits(probe, candidate); err != nil {
			t.Fatalf("control ApplySchemaEdits(%s=%s): %v", survivor.Name, v, err)
		}
		if readSection(t, probe, survivor.Persist.Section) != before {
			edits = candidate
			t.Logf("control: %q=%s changes %s.yaml — probe discriminates",
				survivor.Name, v, survivor.Persist.Section)
			break
		}
	}
	if edits == nil {
		t.Skipf("neither bool value changes %s.yaml for %q — probe cannot discriminate",
			survivor.Persist.Section, survivor.Name)
	}

	// Re-seed and block the section the loop reaches second.
	root2 := t.TempDir()
	seedSectionFixture(t, root2, survivor.Persist.Section)
	seedSectionFixture(t, root2, blocked.Persist.Section)
	pre := readSection(t, root2, survivor.Persist.Section)
	blockPath(t, filepath.Join(root2, ".moai", "config", "sections", blocked.Persist.Section+".yaml"))

	applyErr := ApplySchemaEdits(root2, edits)
	if applyErr == nil {
		t.Skip("ApplySchemaEdits succeeded with a blocked section file — probe inapplicable")
	}
	post := readSection(t, root2, survivor.Persist.Section)
	partial := post != pre
	t.Logf("L3 yamlpatch: ApplySchemaEdits failed on %q (%v); the earlier section %q was already patched: %v",
		blocked.Persist.Section, applyErr, survivor.Persist.Section, partial)
	if !partial {
		t.Errorf("expected %s.yaml to be patched before the loop hit the blocked section", survivor.Persist.Section)
	}
}
