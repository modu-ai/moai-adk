package cli

// migrate_ownership_previewclose_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, chunk 13, leader disposition B): the `--dry-run` preview assembles the same
// verdict environment as the migration, so it opens the project root once, like the
// migration, and closes the handle when the preview ends.

import (
	"os"
	"testing"
)

func TestPreviewVerdictsCloseTheRootTheyOpened(t *testing.T) {
	root, home := ownershipFixtureBase(t)
	prevHome := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = prevHome })
	userHomeDirFn = func() (string, error) { return home, nil }
	var opened []*os.Root
	prevOpen := openProjectRoot
	t.Cleanup(func() { openProjectRoot = prevOpen })
	openProjectRoot = func(name string) (*os.Root, error) {
		r, err := prevOpen(name)
		if err == nil {
			opened = append(opened, r)
		}
		return r, err
	}
	var p previewVerdicts
	p.removalEligible(root, nil, ".agents/skills/moai-preview-close/a.md")
	p.removalEligible(root, nil, ".agents/skills/moai-preview-close/b.md")
	if len(opened) != 1 {
		t.Fatalf("the preview opened the project root %d times, want exactly once", len(opened))
	}
	if f, err := opened[0].Open("."); err != nil {
		t.Fatalf("the root is closed before the preview ended: %v", err)
	} else {
		_ = f.Close()
	}
	p.close()
	if f, err := opened[0].Open("."); err == nil {
		_ = f.Close()
		t.Errorf("the preview's root is still open after close")
	}
	p.close() // closing twice, or a preview that never assembled anything, is harmless
	var unused previewVerdicts
	unused.close()
}
