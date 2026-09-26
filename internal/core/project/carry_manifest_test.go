package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestCarryManifestForward(t *testing.T) {
	manifestPath := func(root string) string { return filepath.Join(root, ".moai", "manifest.json") }

	t.Run("no backup carries nothing", func(t *testing.T) {
		root := t.TempDir()
		if err := CarryManifestForward(root, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(manifestPath(root)); !os.IsNotExist(err) {
			t.Errorf("manifest written without a backup: %v", err)
		}
	})

	t.Run("backup without a manifest carries nothing", func(t *testing.T) {
		root := t.TempDir()
		if err := CarryManifestForward(root, t.TempDir()); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(manifestPath(root)); !os.IsNotExist(err) {
			t.Errorf("manifest written from an empty backup: %v", err)
		}
	})

	t.Run("invalid backed-up manifest carries nothing", func(t *testing.T) {
		root, backup := t.TempDir(), t.TempDir()
		writeFile(t, backup, "manifest.json", "{not json")
		if err := CarryManifestForward(root, backup); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(manifestPath(root)); !os.IsNotExist(err) {
			t.Errorf("invalid manifest carried forward: %v", err)
		}
	})

	t.Run("structurally invalid backed-up manifest carries nothing", func(t *testing.T) {
		for _, body := range []string{"[]", `"x"`, `{"files": 5}`} {
			root, backup := t.TempDir(), t.TempDir()
			writeFile(t, backup, "manifest.json", body)
			if err := CarryManifestForward(root, backup); err != nil {
				t.Fatalf("%s: %v", body, err)
			}
			entries, _ := os.ReadDir(filepath.Join(root, ".moai"))
			if len(entries) != 0 {
				t.Errorf("%s: left behind in .moai/: %v", body, entries)
			}
		}
	})

	t.Run("unreadable and out-of-root entries never abort the carry", func(t *testing.T) {
		parent := t.TempDir()
		root := filepath.Join(parent, "proj")
		mkDir(t, root, "dir-at-file-path")
		writeFile(t, parent, "outside.md", "outside, changed\n")

		deployed := manifest.HashBytes([]byte("as deployed\n"))
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		files := mgr.Manifest().Files
		entry := manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: deployed, DeployedHash: deployed, CurrentHash: deployed}
		files["dir-at-file-path"] = entry
		files["../outside.md"] = entry
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
		backup, err := BackupExistingProject(root)
		if err != nil {
			t.Fatal(err)
		}

		if err := CarryManifestForward(root, backup); err != nil {
			t.Fatalf("carry aborted: %v", err)
		}
		after := manifest.NewManager()
		if _, err := after.Load(root); err != nil {
			t.Fatal(err)
		}
		// A path that is not a regular file cannot be proven unedited, so it
		// is never overwritten.
		if e, _ := after.GetEntry("dir-at-file-path"); e == nil || e.Provenance != manifest.UserModified {
			t.Errorf("dir-at-file-path = %+v, want user_modified", e)
		}
		// A key outside the project root is not read at all: its provenance
		// is carried unchanged even though the file behind it differs.
		if e, _ := after.GetEntry("../outside.md"); e == nil || e.Provenance != manifest.TemplateManaged {
			t.Errorf("../outside.md = %+v, want carried unchanged", e)
		}
	})

	t.Run("reclassifies only edited template files", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "kept.md", "as deployed\n")
		writeFile(t, root, "edited.md", "user edit\n")
		writeFile(t, root, "owned.md", "mine, changed since\n")

		// Record the old deployment in a manifest, then move .moai/ aside the
		// way --force does.
		deployed := manifest.HashBytes([]byte("as deployed\n"))
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		files := mgr.Manifest().Files
		files["kept.md"] = manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: deployed, DeployedHash: deployed, CurrentHash: deployed}
		files["edited.md"] = manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: deployed, DeployedHash: deployed, CurrentHash: deployed}
		files["deleted.md"] = manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: deployed, DeployedHash: deployed, CurrentHash: deployed}
		files["owned.md"] = manifest.FileEntry{Provenance: manifest.UserCreated, CurrentHash: deployed}
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
		backup, err := BackupExistingProject(root)
		if err != nil {
			t.Fatal(err)
		}

		if err := CarryManifestForward(root, backup); err != nil {
			t.Fatal(err)
		}

		after := manifest.NewManager()
		if _, err := after.Load(root); err != nil {
			t.Fatal(err)
		}
		want := map[string]manifest.Provenance{
			"kept.md":    manifest.TemplateManaged,
			"edited.md":  manifest.UserModified,
			"deleted.md": manifest.TemplateManaged,
			"owned.md":   manifest.UserCreated,
		}
		for rel, prov := range want {
			e, ok := after.GetEntry(rel)
			if !ok || e == nil {
				t.Errorf("%s: entry not carried", rel)
				continue
			}
			if e.Provenance != prov {
				t.Errorf("%s provenance = %q, want %q", rel, e.Provenance, prov)
			}
		}
	})
}
