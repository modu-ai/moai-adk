package userassets

// retired_bundle_remove_test.go — codex review finding P2 (card t1591). A bundle
// an older release installed, and the current catalog no longer ships, has no
// entries in collectEntries, so RemoveBundle's before/after closure difference
// cannot see its files. The manifest's own Bundle attribution must cover them.

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveBundlePrunesRetiredBundleRecordedInManifest(t *testing.T) {
	f := newFixture(t)
	inst := f.installer(t)
	if _, err := inst.Install([]string{"extras"}); err != nil {
		t.Fatalf("install: %v", err)
	}
	manifest, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatal(err)
	}

	// An older release installed "retired-pack". The current catalog has no such
	// bundle, so no catalog walk reaches the file it left behind.
	retiredPath := filepath.Join(f.home, ".claude", "skills", "retired-skill", "SKILL.md")
	body := []byte("retired skill\n")
	if err := os.MkdirAll(filepath.Dir(retiredPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(retiredPath, body, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	manifest.Files["claude-skills/retired-skill/SKILL.md"] = FileEntry{
		SHA256: hex.EncodeToString(sum[:]), Bundle: "retired-pack",
		InstalledAt: "t0", MoaiVersion: "vOld",
	}
	manifest.Bundles = append(manifest.Bundles, "retired-pack")

	if _, err := inst.RemoveBundle(manifest, "extras", nil); err != nil {
		t.Fatalf("RemoveBundle: %v", err)
	}
	if _, err := os.Stat(retiredPath); !os.IsNotExist(err) {
		t.Fatalf("the retired bundle's file survived the removal (stat err = %v) — prune targets came from the current catalog only", err)
	}
}
