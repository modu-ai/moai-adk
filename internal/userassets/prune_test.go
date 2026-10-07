// prune_test.go — the selection-based prune (SPEC-USER-ASSET-INSTALL-001
// REQ-009; the sole removal rule iter4 D25; the AC-018 D28 flip-criterion
// arm).
package userassets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPruneRemovesUnselectedTracked pins AC-018's D28 arm: a file of a
// shipped-but-DESELECTED bundle is pruned under the selection-based
// criterion — not kept.
func TestPruneRemovesUnselectedTracked(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	// Deselect: the recorded selection drops extras.
	m, _ := Load(ManifestPath(f.home))
	m.Bundles = nil
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}

	in := f.installer(t)
	m2, _ := Load(ManifestPath(f.home))
	res, err := in.PruneUnselected(m2)
	if err != nil {
		t.Fatal(err)
	}
	if res.Removed == 0 {
		t.Fatal("nothing pruned — the deselected bundle's files survived")
	}
	if _, err := os.Stat(filepath.Join(f.home, ".claude/skills/moai-beta/SKILL.md")); !os.IsNotExist(err) {
		t.Errorf("deselected skill survived the prune: %v", err)
	}
	if _, tracked := m2.Files["claude-skills/moai-beta/SKILL.md"]; tracked {
		t.Error("manifest entry not dropped by the prune")
	}
}

// TestPruneKeepsSelectedAndDivergent pins the REQ-023 removal arm inside the
// prune: a tracked file matching neither its manifest hash nor the shipped
// bytes is preserved (kept + reported), and the L0 set is never a prune
// candidate.
func TestPruneKeepsSelectedAndDivergent(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	// Diverge a selected file (an L0 skill).
	target := filepath.Join(f.home, ".claude/skills/moai-alpha/SKILL.md")
	edited := []byte("user edit\n")
	if err := os.WriteFile(target, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := Load(ManifestPath(f.home))
	in := f.installer(t)
	res, err := in.PruneUnselected(m)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(target); string(got) != string(edited) {
		t.Error("L0 file pruned or rewritten")
	}
	if res.DivergencePreserved != 0 {
		t.Errorf("L0 file entered the prune: %+v", res)
	}
}

// TestPruneMissingFileDropsEntry pins the missing-on-disk arm: a tracked file
// absent from disk at prune time drops its manifest entry and counts removed.
func TestPruneMissingFileDropsEntry(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(f.home, ".claude/skills/moai-beta/SKILL.md")); err != nil {
		t.Fatal(err)
	}
	// Deselect extras so the file becomes a prune candidate.
	m, _ := Load(ManifestPath(f.home))
	m.Bundles = nil
	if err := m.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}
	in := f.installer(t)
	m2, _ := Load(ManifestPath(f.home))
	res, err := in.PruneUnselected(m2)
	if err != nil {
		t.Fatal(err)
	}
	if res.Removed == 0 {
		t.Error("missing tracked file not counted removed")
	}
	if _, tracked := m2.Files["claude-skills/moai-beta/SKILL.md"]; tracked {
		t.Error("manifest entry for the missing file not dropped")
	}
	_ = time.Now
}
