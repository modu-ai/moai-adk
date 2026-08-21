package deploy

// Order-independence test for linked clean roots (SPEC-CLI-CLEAN-SYMLINK-001,
// AC-CSL-008 / REQ-CSL-008). Under the removal-only dispositions, processing
// two clean entries linked to each other in either order must converge to
// the same final tree: whichever entry goes first, the other degrades to a
// form (live link → dangling link) whose disposition is still removal.
// Evaluated at the backupThenRemove unit boundary (plan M3 — no new seam).

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// linkedOrderResult is the observable final state of one ordering run.
type linkedOrderResult struct {
	aIsRealDir   bool // A redeployed as a real (non-link) directory
	bGone        bool // B absent from the tree
	backupIntact bool // A's unmanaged file reached the pre-clean backup
	redeployed   bool // template content landed under A
}

// runLinkedOrder builds the linked pair (real directory A carrying an
// unmanaged file + live directory link B → A, both glob-match names),
// processes them through backupThenRemove in the given order, runs the
// deploy-side MkdirAll simulation, and returns the observable final state.
func runLinkedOrder(t *testing.T, linkFirst bool) linkedOrderResult {
	var res linkedOrderResult
	t.Helper()
	root := t.TempDir()
	skillsDir := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir)
	aDir := filepath.Join(skillsDir, "moai")
	bLink := filepath.Join(skillsDir, "moai-order-link")
	if err := os.MkdirAll(aDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aDir, "user-extra.md"), []byte("ORDER-UNMANAGED-v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, aDir, bLink)

	backupBase := filepath.Join(root, defs.BackupsDir, "order-run", preCleanBackupSubdir)
	tmplFS := preCleanTestFS(".claude/skills/moai/SKILL.md")

	first, second := aDir, bLink
	if linkFirst {
		first, second = bLink, aDir
	}
	rel := func(p string) string {
		r, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	for _, entry := range []string{first, second} {
		if _, _, err := backupThenRemove(entry, rel(entry), backupBase, tmplFS); err != nil {
			t.Fatalf("backupThenRemove(%s): %v", entry, err)
		}
	}

	// Deploy-side simulation for the template-carried name A.
	deploySimMkdirAll(t, aDir)
	if err := os.WriteFile(filepath.Join(aDir, "SKILL.md"), []byte("template:skill"), 0o644); err != nil {
		t.Errorf("redeploy write failed: %v", err)
	}

	if fi, serr := os.Lstat(aDir); serr == nil && fi.Mode().IsDir() && fi.Mode()&os.ModeSymlink == 0 {
		res.aIsRealDir = true
	}
	if _, lerr := os.Lstat(bLink); os.IsNotExist(lerr) {
		res.bGone = true
	}
	if data, rerr := os.ReadFile(filepath.Join(backupBase, rel(aDir), "user-extra.md")); rerr == nil && string(data) == "ORDER-UNMANAGED-v1" {
		res.backupIntact = true
	}
	if data, rerr := os.ReadFile(filepath.Join(aDir, "SKILL.md")); rerr == nil && string(data) == "template:skill" {
		res.redeployed = true
	}
	return res
}

// TestBackupThenRemove_LinkedRootsOrderIndependence is AC-CSL-008: B→A and
// A→B orders produce the identical final tree — A redeployed as a real
// directory with template content, B absent (template-carried-free name),
// and A's unmanaged file in the pre-clean backup regardless of order.
func TestBackupThenRemove_LinkedRootsOrderIndependence(t *testing.T) {
	linkFirst := runLinkedOrder(t, true) // B (link) → A (dir)
	dirFirst := runLinkedOrder(t, false) // A (dir) → B (now dangling)

	if !linkFirst.aIsRealDir || !dirFirst.aIsRealDir {
		t.Errorf("root A not a real redeployed directory in both orders: linkFirst=%v dirFirst=%v",
			linkFirst.aIsRealDir, dirFirst.aIsRealDir)
	}
	if !linkFirst.bGone || !dirFirst.bGone {
		t.Errorf("link B not absent in both orders: linkFirst=%v dirFirst=%v",
			linkFirst.bGone, dirFirst.bGone)
	}
	if !linkFirst.backupIntact || !dirFirst.backupIntact {
		t.Errorf("A's unmanaged file not backed up in both orders: linkFirst=%v dirFirst=%v",
			linkFirst.backupIntact, dirFirst.backupIntact)
	}
	if !linkFirst.redeployed || !dirFirst.redeployed {
		t.Errorf("template content not redeployed under A in both orders: linkFirst=%v dirFirst=%v",
			linkFirst.redeployed, dirFirst.redeployed)
	}
}
