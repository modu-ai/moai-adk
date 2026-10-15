package cli

// init_fresh_corrupt_manifest_test.go — card t1619, leader ruling d-20261014T041318Z-dcea (D2).
// A fresh init runs its user-asset ensure BEFORE the executor (card t1587), and the re-run promise
// that card makes must hold for a damaged user manifest too. The ensure quarantines a damaged
// manifest under the user lock: the file is renamed aside, never deleted or rewritten, and the
// install runs from a clean manifest. Without that step the ensure fails on the damaged file and
// every re-run fails the same way.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

func TestInitQuarantinesCorruptUserManifestBeforeEnsure(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "fresh-corrupt-manifest")
	manifest := userassets.ManifestPath(userHomeDirOrEmpty())
	damaged := []byte("{not a manifest")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, damaged, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := resumeRun(t, projectDir, "claude"); err != nil {
		t.Fatalf("a fresh init over a damaged user manifest must quarantine it under the ensure lock and complete (the t1587 re-run promise): %v", err)
	}

	quarantined, err := filepath.Glob(manifest + ".corrupt-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(quarantined) != 1 {
		t.Fatalf("want exactly one quarantined copy of the damaged manifest, got %d", len(quarantined))
	}
	got, err := os.ReadFile(quarantined[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, damaged) {
		t.Fatal("the quarantined copy differs from the damaged manifest: quarantine renames the file aside and must never rewrite or delete it")
	}
}
