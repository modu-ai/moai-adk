package manifest

// manifest_deployed_version_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M1b
// (card t1865), AC-OWR-06 (a): a manifest entry records the moai binary version
// that DEPLOYED it, only the deploy path stamps it, no other path invents,
// back-fills or clears it, and a manifest written before the field existed
// loads and saves byte for byte.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTrackedProject(t *testing.T, files ...string) (string, Manager) {
	t.Helper()
	root := t.TempDir()
	for _, rel := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("content of "+rel+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mgr := NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load: %v", err)
	}
	return root, mgr
}

func entryOf(t *testing.T, mgr Manager, rel string) FileEntry {
	t.Helper()
	e, ok := mgr.GetEntry(rel)
	if !ok {
		t.Fatalf("no manifest entry for %s", rel)
	}
	return *e
}

// TestManifestRecordsDeployedVersion is AC-OWR-06 (a).
func TestManifestRecordsDeployedVersion(t *testing.T) {
	t.Run("DeployPathStampsTemplateManagedEntries", func(t *testing.T) {
		_, mgr := newTrackedProject(t, "a.md", "b.md")
		mgr.SetDeployingVersion("v-deploy-1")
		if err := mgr.Track("a.md", TemplateManaged, "tpl-a"); err != nil {
			t.Fatal(err)
		}
		if got := entryOf(t, mgr, "a.md").DeployedVersion; got != "v-deploy-1" {
			t.Errorf("a deployed template_managed entry carries DeployedVersion %q, want v-deploy-1", got)
		}
		// The deployer records a file it did NOT write (an existing untracked
		// file) as user_created and skips it: that entry is not stamped.
		if err := mgr.Track("b.md", UserCreated, "tpl-b"); err != nil {
			t.Fatal(err)
		}
		if got := entryOf(t, mgr, "b.md").DeployedVersion; got != "" {
			t.Errorf("a skipped user_created entry carries DeployedVersion %q, want none (the binary did not write it)", got)
		}
	})

	t.Run("NonDeployPathsNeitherSetNorClear", func(t *testing.T) {
		_, mgr := newTrackedProject(t, "kept.md", "legacy.md", "fresh.md")
		mgr.SetDeployingVersion("v-old")
		if err := mgr.Track("kept.md", TemplateManaged, "tpl"); err != nil {
			t.Fatal(err)
		}
		if err := mgr.Track("legacy.md", UserCreated, "tpl"); err != nil { // never stamped
			t.Fatal(err)
		}
		mgr.SetDeployingVersion("") // the deploy window is closed

		// Re-tracking (the post-merge retrack path) rewrites hashes but must keep
		// what is already recorded and must not stamp anything new.
		for _, rel := range []string{"kept.md", "legacy.md"} {
			if err := mgr.Track(rel, TemplateManaged, "tpl"); err != nil {
				t.Fatal(err)
			}
		}
		if got := entryOf(t, mgr, "kept.md").DeployedVersion; got != "v-old" {
			t.Errorf("re-tracking cleared or changed DeployedVersion: %q, want v-old", got)
		}
		if got := entryOf(t, mgr, "legacy.md").DeployedVersion; got != "" {
			t.Errorf("re-tracking back-filled DeployedVersion %q on an entry the run did not deploy", got)
		}
		if err := mgr.Track("fresh.md", TemplateManaged, "tpl"); err != nil {
			t.Fatal(err)
		}
		if got := entryOf(t, mgr, "fresh.md").DeployedVersion; got != "" {
			t.Errorf("a track outside the deploy window stamped DeployedVersion %q", got)
		}
	})

	t.Run("PersistsUnderTheDeployedVersionKeyAndOmitsEmpty", func(t *testing.T) {
		root, mgr := newTrackedProject(t, "stamped.md", "plain.md")
		mgr.SetDeployingVersion("v-persist")
		if err := mgr.Track("stamped.md", TemplateManaged, "tpl"); err != nil {
			t.Fatal(err)
		}
		mgr.SetDeployingVersion("")
		if err := mgr.Track("plain.md", TemplateManaged, "tpl"); err != nil {
			t.Fatal(err)
		}
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(root, ".moai", "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"deployed_version": "v-persist"`) {
			t.Errorf("the stamped entry is not stored under the deployed_version key:\n%s", raw)
		}
		if strings.Count(string(raw), `"deployed_version"`) != 1 {
			t.Errorf("an unstamped entry carries an empty deployed_version key (it must be omitted):\n%s", raw)
		}
		reload := NewManager()
		if _, err := reload.Load(root); err != nil {
			t.Fatal(err)
		}
		if got := entryOf(t, reload, "stamped.md").DeployedVersion; got != "v-persist" {
			t.Errorf("reloaded DeployedVersion %q, want v-persist", got)
		}
	})

	t.Run("LegacyJSONRoundTripsByteForByte", func(t *testing.T) {
		root := t.TempDir()
		legacy := &Manifest{
			Version:    "v3.0.0",
			DeployedAt: "2026-01-02T03:04:05Z",
			Files: map[string]FileEntry{
				".claude/commands/moai/x.md": {Provenance: TemplateManaged, TemplateHash: "sha256:aa", DeployedHash: "sha256:aa", CurrentHash: "sha256:aa"},
				".claude/skills/y/SKILL.md":  {Provenance: UserModified, TemplateHash: "sha256:bb", DeployedHash: "sha256:bb", CurrentHash: "sha256:cc", KeptBy: KeptByProfileReach},
			},
			TemplateTreeHash: "sha256:tree",
		}
		want, err := json.MarshalIndent(legacy, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		want = append(want, '\n')
		if bytes.Contains(want, []byte("deployed_version")) {
			t.Fatalf("the legacy fixture must not contain the new key")
		}
		path := filepath.Join(root, ".moai", "manifest.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, want, 0o644); err != nil {
			t.Fatal(err)
		}
		mgr := NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("a manifest written before the field existed did not round-trip byte for byte:\n--- want\n%s\n--- got\n%s", want, got)
		}
	})
}
