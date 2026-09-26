package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestHealManifestFromBackups pins the repair for projects an earlier
// `init --force` already damaged: every file it found outside .moai/ was
// recorded user_created, and the manifest that knew better was moved into
// .moai-backups/<ts>/. The heal reads that backup back.
func TestHealManifestFromBackups(t *testing.T) {
	const (
		untouched = ".claude/rules/moai/a.md"
		edited    = ".claude/rules/moai/b.md"
		owned     = ".claude/rules/moai/c.md"
		absent    = ".claude/rules/moai/d.md"
	)
	oldBody, newTmpl := "older deploy\n", "newer template\n"
	oldH, newH := manifest.HashBytes([]byte(oldBody)), manifest.HashBytes([]byte(newTmpl))
	managed := manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: oldH, DeployedHash: oldH, CurrentHash: oldH}
	damaged := manifest.FileEntry{Provenance: manifest.UserCreated, TemplateHash: newH, DeployedHash: newH, CurrentHash: newH}

	saveManifest := func(t *testing.T, dir string, files map[string]manifest.FileEntry) {
		t.Helper()
		mgr := manifest.NewManager()
		if _, err := mgr.Load(dir); err != nil {
			t.Fatal(err)
		}
		for k, v := range files {
			mgr.Manifest().Files[k] = v
		}
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
	}
	// writeBackup places a manifest at .moai-backups/<ts>/manifest.json, the
	// layout BackupExistingProject leaves behind.
	writeBackup := func(t *testing.T, root, ts string, files map[string]manifest.FileEntry) {
		t.Helper()
		tmp := t.TempDir()
		saveManifest(t, tmp, files)
		data := readFile(t, filepath.Join(tmp, defs.MoAIDir, defs.ManifestJSON))
		writeFile(t, root, filepath.Join(defs.BackupsDir, ts, defs.ManifestJSON), data)
	}

	root := t.TempDir()
	writeFile(t, root, untouched, oldBody)
	writeFile(t, root, edited, "the user's edit\n")
	writeFile(t, root, owned, "a file the user owns\n")
	ownedH := manifest.HashBytes([]byte("a file the user owns\n"))
	// The pre-damage manifest, then a second, already-damaged backup left by
	// a later init --force: the heal must see through the newer one.
	writeBackup(t, root, "20260901_100000", map[string]manifest.FileEntry{
		untouched: managed, edited: managed, absent: managed,
		owned: {Provenance: manifest.UserCreated, TemplateHash: newH, DeployedHash: ownedH, CurrentHash: ownedH},
	})
	writeBackup(t, root, "20260902_100000", map[string]manifest.FileEntry{
		untouched: damaged, edited: damaged, owned: damaged,
	})
	saveManifest(t, root, map[string]manifest.FileEntry{untouched: damaged, edited: damaged, owned: damaged})

	healed, err := HealManifestFromBackups(root)
	if err != nil {
		t.Fatal(err)
	}
	if healed != 2 {
		t.Errorf("healed = %d, want 2", healed)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	got := func(rel string) manifest.FileEntry {
		e, _ := mgr.GetEntry(rel)
		if e == nil {
			return manifest.FileEntry{}
		}
		return *e
	}
	if e := got(untouched); e.Provenance != manifest.TemplateManaged || e.CurrentHash != oldH {
		t.Errorf("%s = %+v, want template_managed at the backed-up hash", untouched, e)
	}
	if e := got(edited); e.Provenance != manifest.UserModified {
		t.Errorf("%s provenance = %q, want user_modified", edited, e.Provenance)
	}
	if e := got(owned); e.Provenance != manifest.UserCreated {
		t.Errorf("%s provenance = %q, want user_created kept", owned, e.Provenance)
	}
	if _, ok := mgr.GetEntry(absent); ok {
		t.Errorf("%s: heal added an entry the live manifest never had", absent)
	}
	if n, err := HealManifestFromBackups(root); err != nil || n != 0 {
		t.Errorf("second heal = %d, %v; want 0, nil", n, err)
	}
	if file := readFile(t, filepath.Join(root, edited)); file != "the user's edit\n" {
		t.Errorf("heal touched a file on disk: %q", file)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestHealManifestFromBackupsWithoutBackupsIsNoop(t *testing.T) {
	root := t.TempDir()
	if n, err := HealManifestFromBackups(root); err != nil || n != 0 {
		t.Errorf("no manifest, no backups: %d, %v", n, err)
	}
	writeFile(t, root, filepath.Join(defs.BackupsDir, "20260901_100000", defs.ManifestJSON), "{not json")
	if n, err := HealManifestFromBackups(root); err != nil || n != 0 {
		t.Errorf("corrupt backup: %d, %v", n, err)
	}
	// A live manifest that is valid JSON but not a manifest is the update's
	// concern: the heal must leave it where it is, not move it aside.
	for _, body := range []string{"[]", `{"files":[]}`, `{"version":3,"files":{}}`} {
		live := filepath.Join(defs.MoAIDir, defs.ManifestJSON)
		writeFile(t, root, live, body)
		if n, err := HealManifestFromBackups(root); err != nil || n != 0 {
			t.Errorf("%s: %d, %v", body, n, err)
		}
		if got := readFile(t, filepath.Join(root, live)); got != body {
			t.Errorf("%s: live manifest changed to %q", body, got)
		}
	}
}
