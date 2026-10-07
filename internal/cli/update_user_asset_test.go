// update_user_asset_test.go — `moai update`'s user-asset phase
// (SPEC-USER-ASSET-INSTALL-001 M3; AC-005's behavior verdict basis + the
// REQ-009 selection-based prune + the REQ-004 selection honoring).
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

func shaHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// TestUpdatePhaseRefreshesManifestMatch pins AC-005's verdict basis: a
// manifest-tracked user file whose current hash equals its manifest hash
// while differing from the shipped bytes is refreshed to the shipped bytes,
// hash and version re-recorded.
func TestUpdatePhaseRefreshesManifestMatch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	home, _ := os.UserHomeDir()

	if err := ensureUserAssetsLocked(home, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	target := filepath.Join(home, ".claude", "skills", "moai-workflow-tdd", "SKILL.md")
	edited := []byte("user edits, then moai refreshes\n")
	if err := os.WriteFile(target, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	// The manifest records the EDITED hash (current == manifest ≠ shipped).
	manifestPath := userassets.ManifestPath(home)
	m, _ := userassets.Load(manifestPath)
	fe := m.Files["claude-skills/moai-workflow-tdd/SKILL.md"]
	fe.SHA256 = shaHex(edited)
	m.Files["claude-skills/moai-workflow-tdd/SKILL.md"] = fe
	if err := m.Save(manifestPath); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runUserAssetUpdatePhase(home, &out); err != nil {
		t.Fatalf("update phase: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == string(edited) {
		t.Error("refresh did not rewrite the file to the shipped bytes")
	}
	m2, _ := userassets.Load(manifestPath)
	if m2.Files["claude-skills/moai-workflow-tdd/SKILL.md"].SHA256 == shaHex(edited) {
		t.Error("manifest hash not refreshed")
	}
	if !strings.Contains(out.String(), "refreshed") {
		t.Errorf("summary missing the refreshed count: %q", out.String())
	}
}

// TestUpdatePhaseHonorsRecordedSelectionAndPrunes pins the REQ-004
// update-honoring arm + the D28 prune: a bundle added then deselected at the
// manifest (the artifact of a selection change) is pruned by the phase, and
// the recorded selection itself survives the run.
func TestUpdatePhaseHonorsRecordedSelectionAndPrunes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	home, _ := os.UserHomeDir()

	cmd := newBundleTestCmd()
	if err := runBundleAdd(cmd, []string{"ops-tools"}); err != nil {
		t.Fatal(err)
	}
	// Deselect at the manifest (the recorded selection is the state the
	// phase reads and honors).
	manifestPath := userassets.ManifestPath(home)
	m, _ := userassets.Load(manifestPath)
	m.Bundles = nil
	if err := m.Save(manifestPath); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runUserAssetUpdatePhase(home, &out); err != nil {
		t.Fatalf("update phase: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "moai-workflow-loop", "SKILL.md")); !os.IsNotExist(err) {
		t.Error("deselected bundle artifact survived the update phase (the D28 flip arm)")
	}
	m2, _ := userassets.Load(manifestPath)
	for _, b := range m2.Bundles {
		if b == "ops-tools" {
			t.Error("deselected bundle back in the recorded selection")
		}
	}
	// The L0 set is untouched.
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "moai-workflow-tdd", "SKILL.md")); err != nil {
		t.Errorf("L0 skill pruned by mistake: %v", err)
	}
}
