package manifest

import (
	"errors"
	"testing"
)

func TestManagerSetKeptBy(t *testing.T) {
	t.Run("marks_entry_and_keeps_other_fields", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, ".claude/settings.json", []byte("{}"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track(".claude/settings.json", UserModified, "sha256:tmpl"); err != nil {
			t.Fatalf("Track error: %v", err)
		}
		before, ok := mgr.GetEntry(".claude/settings.json")
		if !ok {
			t.Fatal("entry not found after Track")
		}

		if err := mgr.SetKeptBy(".claude/settings.json", KeptByProfileReach); err != nil {
			t.Fatalf("SetKeptBy error: %v", err)
		}

		after, ok := mgr.GetEntry(".claude/settings.json")
		if !ok {
			t.Fatal("entry not found after SetKeptBy")
		}
		if after.KeptBy != KeptByProfileReach {
			t.Errorf("KeptBy = %q, want %q", after.KeptBy, KeptByProfileReach)
		}
		if after.Provenance != before.Provenance || after.TemplateHash != before.TemplateHash ||
			after.DeployedHash != before.DeployedHash || after.CurrentHash != before.CurrentHash {
			t.Errorf("SetKeptBy changed fields other than KeptBy: before=%+v after=%+v", *before, *after)
		}
	})

	t.Run("mark_survives_save_and_load", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "CLAUDE.md", []byte("project guide"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if err := mgr.Track("CLAUDE.md", TemplateManaged, "sha256:tmpl"); err != nil {
			t.Fatalf("Track error: %v", err)
		}
		if err := mgr.SetKeptBy("CLAUDE.md", KeptByProfileReach); err != nil {
			t.Fatalf("SetKeptBy error: %v", err)
		}
		if err := mgr.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		reloaded := NewManager()
		if _, err := reloaded.Load(root); err != nil {
			t.Fatalf("Load after Save error: %v", err)
		}
		entry, ok := reloaded.GetEntry("CLAUDE.md")
		if !ok {
			t.Fatal("entry not found after reload")
		}
		if entry.KeptBy != KeptByProfileReach {
			t.Errorf("KeptBy after reload = %q, want %q", entry.KeptBy, KeptByProfileReach)
		}
	})

	t.Run("unknown_path_returns_entry_not_found", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		err := mgr.SetKeptBy("missing.md", KeptByProfileReach)
		if !errors.Is(err, ErrEntryNotFound) {
			t.Fatalf("expected ErrEntryNotFound, got: %v", err)
		}
		if _, ok := mgr.GetEntry("missing.md"); ok {
			t.Error("SetKeptBy must not create an entry for an unknown path")
		}
	})

	t.Run("uninitialized_manager_returns_manifest_not_found", func(t *testing.T) {
		mgr := NewManager()

		err := mgr.SetKeptBy("any.md", KeptByProfileReach)
		if !errors.Is(err, ErrManifestNotFound) {
			t.Fatalf("expected ErrManifestNotFound, got: %v", err)
		}
	})
}
