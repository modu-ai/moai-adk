package manifest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// setupProject creates a temporary project directory with .moai/ subdirectory.
func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	return dir
}

// writeManifest writes a manifest JSON file to the project's .moai directory.
func writeManifest(t *testing.T, projectRoot string, data []byte) {
	t.Helper()
	path := filepath.Join(projectRoot, ".moai", manifestFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
}

// writeProjectFile creates a file in the project root with the given content.
func writeProjectFile(t *testing.T, projectRoot, relPath string, content []byte) {
	t.Helper()
	absPath := filepath.Join(projectRoot, relPath)
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	if err := os.WriteFile(absPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
}

func TestManagerLoad(t *testing.T) {
	t.Run("valid_manifest", func(t *testing.T) {
		root := setupProject(t)

		mf := &Manifest{
			Version:    "1.0.0",
			DeployedAt: "2026-02-03T10:00:00Z",
			Files: map[string]FileEntry{
				"CLAUDE.md": {
					Provenance:   TemplateManaged,
					TemplateHash: "sha256:abc",
					DeployedHash: "sha256:abc",
					CurrentHash:  "sha256:abc",
				},
			},
		}
		data, err := json.MarshalIndent(mf, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent error: %v", err)
		}
		writeManifest(t, root, data)

		mgr := NewManager()
		loaded, err := mgr.Load(root)
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if loaded.Version != "1.0.0" {
			t.Errorf("Version = %q, want %q", loaded.Version, "1.0.0")
		}
		if loaded.DeployedAt != "2026-02-03T10:00:00Z" {
			t.Errorf("DeployedAt = %q, want %q", loaded.DeployedAt, "2026-02-03T10:00:00Z")
		}
		if len(loaded.Files) != 1 {
			t.Fatalf("Files count = %d, want 1", len(loaded.Files))
		}

		entry, ok := loaded.Files["CLAUDE.md"]
		if !ok {
			t.Fatal("missing entry for CLAUDE.md")
		}
		if entry.Provenance != TemplateManaged {
			t.Errorf("Provenance = %v, want %v", entry.Provenance, TemplateManaged)
		}
	})

	t.Run("file_not_exist_returns_empty_manifest", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		loaded, err := mgr.Load(root)
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if loaded == nil {
			t.Fatal("Load returned nil manifest")
		}
		if len(loaded.Files) != 0 {
			t.Errorf("Files count = %d, want 0", len(loaded.Files))
		}
	})

	t.Run("corrupt_json_recovery", func(t *testing.T) {
		root := setupProject(t)
		writeManifest(t, root, []byte("{invalid json!!!"))

		mgr := NewManager()
		loaded, err := mgr.Load(root)

		// Should return ErrManifestCorrupt
		if !errors.Is(err, ErrManifestCorrupt) {
			t.Fatalf("expected ErrManifestCorrupt, got: %v", err)
		}

		// Should still return a usable empty manifest
		if loaded == nil {
			t.Fatal("Load returned nil manifest on corrupt")
		}
		if len(loaded.Files) != 0 {
			t.Errorf("Files count = %d, want 0", len(loaded.Files))
		}

		// Corrupt file should be backed up
		corruptPath := filepath.Join(root, ".moai", manifestFileName+".corrupt")
		if _, err := os.Stat(corruptPath); errors.Is(err, os.ErrNotExist) {
			t.Error("corrupt backup file was not created")
		}

		// Original should no longer exist (was renamed to .corrupt)
		origPath := filepath.Join(root, ".moai", manifestFileName)
		if _, err := os.Stat(origPath); !errors.Is(err, os.ErrNotExist) {
			t.Error("original corrupt file should have been renamed")
		}
	})
}

func TestManagerSave(t *testing.T) {
	t.Run("save_and_reload", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		loaded, err := mgr.Load(root)
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}

		loaded.Version = "1.14.0"
		loaded.DeployedAt = "2026-02-03T12:00:00Z"
		loaded.Files["test.md"] = FileEntry{
			Provenance:   TemplateManaged,
			TemplateHash: "sha256:aaa",
			DeployedHash: "sha256:aaa",
			CurrentHash:  "sha256:aaa",
		}

		if err := mgr.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		// Verify file exists and is valid JSON
		manifestPath := filepath.Join(root, ".moai", manifestFileName)
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}
		if !json.Valid(data) {
			t.Fatal("saved manifest is not valid JSON")
		}

		// Reload and verify
		mgr2 := NewManager()
		reloaded, err := mgr2.Load(root)
		if err != nil {
			t.Fatalf("Load after Save error: %v", err)
		}

		if reloaded.Version != "1.14.0" {
			t.Errorf("Version = %q, want %q", reloaded.Version, "1.14.0")
		}
		if len(reloaded.Files) != 1 {
			t.Fatalf("Files count = %d, want 1", len(reloaded.Files))
		}
	})

	t.Run("save_creates_directory", func(t *testing.T) {
		root := t.TempDir()
		// Do NOT create .moai/ directory

		mgr := &manifestManager{
			projectRoot:  root,
			manifestPath: filepath.Join(root, ".moai", manifestFileName),
			manifest:     NewManifest(),
		}
		mgr.manifest.Version = "1.0.0"

		if err := mgr.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		// Verify directory and file exist
		path := filepath.Join(root, ".moai", manifestFileName)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("manifest file should exist: %v", err)
		}
	})

	t.Run("save_without_load_returns_error", func(t *testing.T) {
		mgr := NewManager()
		err := mgr.Save()
		if err == nil {
			t.Fatal("expected error from Save without Load")
		}
		if !errors.Is(err, ErrManifestNotFound) {
			t.Errorf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("atomic_write_no_partial_content", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		mf, err := mgr.Load(root)
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}

		mf.Version = "1.0.0"
		mf.Files["file.txt"] = FileEntry{
			Provenance:   TemplateManaged,
			TemplateHash: "sha256:test",
			DeployedHash: "sha256:test",
			CurrentHash:  "sha256:test",
		}

		if err := mgr.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		// Read and verify complete JSON
		path := filepath.Join(root, ".moai", manifestFileName)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile error: %v", err)
		}

		var restored Manifest
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("saved data is not valid JSON: %v", err)
		}

		if restored.Version != "1.0.0" {
			t.Errorf("Version = %q, want %q", restored.Version, "1.0.0")
		}
	})
}

func TestManagerTrack(t *testing.T) {
	t.Run("track_new_file", func(t *testing.T) {
		root := setupProject(t)
		content := []byte("hello template")
		writeProjectFile(t, root, ".claude/settings.json", content)

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		templateHash := HashBytes(content)
		if err := mgr.Track(".claude/settings.json", TemplateManaged, templateHash); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		entry, ok := mgr.GetEntry(".claude/settings.json")
		if !ok {
			t.Fatal("entry not found after Track")
		}

		if entry.Provenance != TemplateManaged {
			t.Errorf("Provenance = %v, want %v", entry.Provenance, TemplateManaged)
		}
		if entry.TemplateHash != templateHash {
			t.Errorf("TemplateHash = %q, want %q", entry.TemplateHash, templateHash)
		}
		if entry.DeployedHash == "" {
			t.Error("DeployedHash should not be empty")
		}
		if entry.CurrentHash == "" {
			t.Error("CurrentHash should not be empty")
		}
		if entry.DeployedHash != entry.CurrentHash {
			t.Error("DeployedHash and CurrentHash should be equal at track time")
		}
	})

	t.Run("track_updates_existing", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "CLAUDE.md", []byte("version 1"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("CLAUDE.md", TemplateManaged, "sha256:tmpl1"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		// Update file and re-track
		writeProjectFile(t, root, "CLAUDE.md", []byte("version 2"))

		if err := mgr.Track("CLAUDE.md", TemplateManaged, "sha256:tmpl2"); err != nil {
			t.Fatalf("Re-Track error: %v", err)
		}

		entry, ok := mgr.GetEntry("CLAUDE.md")
		if !ok {
			t.Fatal("entry not found after re-Track")
		}
		if entry.TemplateHash != "sha256:tmpl2" {
			t.Errorf("TemplateHash = %q, want %q", entry.TemplateHash, "sha256:tmpl2")
		}
	})

	t.Run("track_nonexistent_file_returns_error", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		err := mgr.Track("nonexistent.txt", TemplateManaged, "sha256:xxx")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestManagerGetEntry(t *testing.T) {
	t.Run("existing_entry", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "file.md", []byte("content"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("file.md", TemplateManaged, "sha256:hash"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		entry, ok := mgr.GetEntry("file.md")
		if !ok {
			t.Fatal("expected entry to exist")
		}
		if entry == nil {
			t.Fatal("entry is nil")
		}
	})

	t.Run("nonexistent_entry", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		entry, ok := mgr.GetEntry("nonexistent.txt")
		if ok {
			t.Error("expected ok to be false")
		}
		if entry != nil {
			t.Error("expected nil entry")
		}
	})
}

func TestManagerRemove(t *testing.T) {
	t.Run("remove_existing", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "file.md", []byte("content"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("file.md", TemplateManaged, "sha256:hash"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		if err := mgr.Remove("file.md"); err != nil {
			t.Fatalf("Remove error: %v", err)
		}

		_, ok := mgr.GetEntry("file.md")
		if ok {
			t.Error("entry should not exist after Remove")
		}
	})

	t.Run("remove_nonexistent_is_noop", func(t *testing.T) {
		root := setupProject(t)

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		// Removing a nonexistent entry should not error
		if err := mgr.Remove("nonexistent.txt"); err != nil {
			t.Fatalf("Remove error: %v", err)
		}
	})
}

func TestManagerDetectChanges(t *testing.T) {
	t.Run("no_changes", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "file.md", []byte("content"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("file.md", TemplateManaged, "sha256:tmpl"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		changes, err := mgr.DetectChanges()
		if err != nil {
			t.Fatalf("DetectChanges error: %v", err)
		}
		if len(changes) != 0 {
			t.Errorf("expected 0 changes, got %d", len(changes))
		}
	})

	t.Run("modified_file", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "file.md", []byte("original"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("file.md", TemplateManaged, "sha256:tmpl"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		// Modify the file
		writeProjectFile(t, root, "file.md", []byte("modified"))

		changes, err := mgr.DetectChanges()
		if err != nil {
			t.Fatalf("DetectChanges error: %v", err)
		}
		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d", len(changes))
		}

		if changes[0].Path != "file.md" {
			t.Errorf("Path = %q, want %q", changes[0].Path, "file.md")
		}
		if changes[0].OldHash == changes[0].NewHash {
			t.Error("OldHash and NewHash should differ")
		}
		if changes[0].NewHash == "" {
			t.Error("NewHash should not be empty for modified file")
		}
	})

	t.Run("deleted_file", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "file.md", []byte("content"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("file.md", TemplateManaged, "sha256:tmpl"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		// Delete the file
		_ = os.Remove(filepath.Join(root, "file.md"))

		changes, err := mgr.DetectChanges()
		if err != nil {
			t.Fatalf("DetectChanges error: %v", err)
		}
		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d", len(changes))
		}

		if changes[0].Path != "file.md" {
			t.Errorf("Path = %q, want %q", changes[0].Path, "file.md")
		}
		if changes[0].NewHash != "" {
			t.Errorf("NewHash = %q, want empty for deleted file", changes[0].NewHash)
		}
	})

	t.Run("mixed_changes", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "unchanged.md", []byte("same"))
		writeProjectFile(t, root, "modified.md", []byte("original"))
		writeProjectFile(t, root, "deleted.md", []byte("will be deleted"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		for _, name := range []string{"unchanged.md", "modified.md", "deleted.md"} {
			if err := mgr.Track(name, TemplateManaged, "sha256:tmpl"); err != nil {
				t.Fatalf("Track(%s) error: %v", name, err)
			}
		}

		// Modify one, delete one, leave one unchanged
		writeProjectFile(t, root, "modified.md", []byte("changed content"))
		_ = os.Remove(filepath.Join(root, "deleted.md"))

		changes, err := mgr.DetectChanges()
		if err != nil {
			t.Fatalf("DetectChanges error: %v", err)
		}
		if len(changes) != 2 {
			t.Fatalf("expected 2 changes, got %d", len(changes))
		}

		pathSet := make(map[string]bool)
		for _, c := range changes {
			pathSet[c.Path] = true
		}

		if !pathSet["modified.md"] {
			t.Error("modified.md should be in changes")
		}
		if !pathSet["deleted.md"] {
			t.Error("deleted.md should be in changes")
		}
		if pathSet["unchanged.md"] {
			t.Error("unchanged.md should not be in changes")
		}
	})
}

func TestManagerIntegration(t *testing.T) {
	t.Run("full_lifecycle", func(t *testing.T) {
		root := setupProject(t)

		// Create test files
		writeProjectFile(t, root, ".claude/settings.json", []byte(`{"hooks":{}}`))
		writeProjectFile(t, root, "CLAUDE.md", []byte("# MoAI"))

		// 1. Load (fresh manifest)
		mgr := NewManager()
		mf, err := mgr.Load(root)
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
		if len(mf.Files) != 0 {
			t.Fatalf("expected empty manifest")
		}

		// 2. Track files
		for _, name := range []string{".claude/settings.json", "CLAUDE.md"} {
			if err := mgr.Track(name, TemplateManaged, "sha256:tmpl"); err != nil {
				t.Fatalf("Track(%s) error: %v", name, err)
			}
		}
		if len(mf.Files) != 2 {
			t.Fatalf("expected 2 files tracked, got %d", len(mf.Files))
		}

		// 3. Save
		mf.Version = "1.0.0"
		if err := mgr.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		// 4. Reload from disk
		mgr2 := NewManager()
		mf2, err := mgr2.Load(root)
		if err != nil {
			t.Fatalf("Reload error: %v", err)
		}
		if len(mf2.Files) != 2 {
			t.Fatalf("expected 2 files after reload, got %d", len(mf2.Files))
		}

		// 5. DetectChanges (no changes)
		changes, err := mgr2.DetectChanges()
		if err != nil {
			t.Fatalf("DetectChanges error: %v", err)
		}
		if len(changes) != 0 {
			t.Errorf("expected 0 changes, got %d", len(changes))
		}

		// 6. Modify file and detect
		writeProjectFile(t, root, "CLAUDE.md", []byte("# Modified MoAI"))
		changes, err = mgr2.DetectChanges()
		if err != nil {
			t.Fatalf("DetectChanges error: %v", err)
		}
		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d", len(changes))
		}
		if changes[0].Path != "CLAUDE.md" {
			t.Errorf("changed path = %q, want %q", changes[0].Path, "CLAUDE.md")
		}

		// 7. Remove entry
		if err := mgr2.Remove(".claude/settings.json"); err != nil {
			t.Fatalf("Remove error: %v", err)
		}

		_, ok := mgr2.GetEntry(".claude/settings.json")
		if ok {
			t.Error("entry should not exist after Remove")
		}
	})
}

func TestManagerManifest(t *testing.T) {
	t.Run("nil_before_load", func(t *testing.T) {
		mgr := NewManager()
		if mf := mgr.Manifest(); mf != nil {
			t.Errorf("Manifest() = %v, want nil before Load", mf)
		}
	})

	t.Run("returns_loaded_manifest", func(t *testing.T) {
		root := setupProject(t)
		mgr := NewManager()
		loaded, err := mgr.Load(root)
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}

		got := mgr.Manifest()
		if got != loaded {
			t.Error("Manifest() should return the same pointer as Load()")
		}
	})

	t.Run("preserves_tracked_entries", func(t *testing.T) {
		root := setupProject(t)
		writeProjectFile(t, root, "test.md", []byte("content"))

		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatalf("Load error: %v", err)
		}

		if err := mgr.Track("test.md", TemplateManaged, "sha256:tmpl"); err != nil {
			t.Fatalf("Track error: %v", err)
		}

		// Manifest() should return the manifest with the tracked entry
		mf := mgr.Manifest()
		if mf == nil {
			t.Fatal("Manifest() returned nil after Load+Track")
		}
		if len(mf.Files) != 1 {
			t.Errorf("Files count = %d, want 1", len(mf.Files))
		}
	})
}

func TestManagerNilManifestGuards(t *testing.T) {
	t.Run("getentry_nil_manifest", func(t *testing.T) {
		mgr := NewManager()
		entry, ok := mgr.GetEntry("any.txt")
		if ok {
			t.Error("expected ok=false for uninitialized manager")
		}
		if entry != nil {
			t.Error("expected nil entry for uninitialized manager")
		}
	})

	t.Run("remove_nil_manifest", func(t *testing.T) {
		mgr := NewManager()
		err := mgr.Remove("any.txt")
		if err == nil {
			t.Fatal("expected error for uninitialized manager")
		}
		if !errors.Is(err, ErrManifestNotFound) {
			t.Errorf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("track_nil_manifest", func(t *testing.T) {
		mgr := NewManager()
		err := mgr.Track("any.txt", TemplateManaged, "sha256:xxx")
		if err == nil {
			t.Fatal("expected error for uninitialized manager")
		}
		if !errors.Is(err, ErrManifestNotFound) {
			t.Errorf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("detect_changes_nil_manifest", func(t *testing.T) {
		mgr := NewManager()
		_, err := mgr.DetectChanges()
		if err == nil {
			t.Fatal("expected error for uninitialized manager")
		}
		if !errors.Is(err, ErrManifestNotFound) {
			t.Errorf("expected ErrManifestNotFound, got: %v", err)
		}
	})
}

func TestManagerSaveReadOnlyDir(t *testing.T) {
	t.Run("save_to_readonly_parent", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows (Unix permissions not supported)")
		}

		root := t.TempDir()
		readonlyDir := filepath.Join(root, "readonly")
		if err := os.MkdirAll(readonlyDir, 0o555); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}

		mgr := &manifestManager{
			projectRoot:  readonlyDir,
			manifestPath: filepath.Join(readonlyDir, ".moai", manifestFileName),
			manifest:     NewManifest(),
		}
		mgr.manifest.Version = "1.0.0"

		err := mgr.Save()
		if err == nil {
			t.Fatal("expected error saving to readonly directory")
		}
	})
}

func TestManagerLoadPermissionError(t *testing.T) {
	t.Run("unreadable_manifest", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping permission test on Windows (Unix permissions not supported)")
		}

		root := setupProject(t)
		manifestPath := filepath.Join(root, ".moai", manifestFileName)

		// Write a manifest then make it unreadable
		if err := os.WriteFile(manifestPath, []byte(`{"version":"1.0.0","files":{}}`), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}
		if err := os.Chmod(manifestPath, 0o000); err != nil {
			t.Fatalf("Chmod error: %v", err)
		}
		t.Cleanup(func() {
			// Restore permissions for cleanup
			_ = os.Chmod(manifestPath, 0o644)
		})

		mgr := NewManager()
		_, err := mgr.Load(root)
		if err == nil {
			t.Fatal("expected error for unreadable manifest")
		}
	})
}
