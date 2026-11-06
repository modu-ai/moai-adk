package cli

// Card t1702, audit iter5 P3: the Lstat check and the removal of
// <config-root>/hooks/moai were two steps, so a swap of `hooks` for a link
// between them made RemoveAll delete the link's external target. The codex leg
// of the audit reproduced it; the lead auditor did not. This test makes the
// window deterministic through the seam between the two steps.

import (
	"os"
	"path/filepath"
	"testing"
)

func writeHooksFixtureFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveConfigRootMoaiHooks_ASwapToALinkAfterTheCheckNeverReachesTheExternalTree(t *testing.T) {
	base := t.TempDir()
	configRoot := filepath.Join(base, "profile")
	external := filepath.Join(base, "external")
	writeHooksFixtureFile(t, filepath.Join(configRoot, "hooks", "moai", "hook.sh"), "managed\n")
	precious := filepath.Join(external, "moai", "precious.txt")
	writeHooksFixtureFile(t, precious, "not ours\n")

	orig := afterMoaiHooksCheck
	afterMoaiHooksCheck = func(root string) {
		// Between the check and the removal, `hooks` becomes a link to a tree
		// outside the config root.
		if err := os.Rename(filepath.Join(root, "hooks"), filepath.Join(root, "hooks.real")); err != nil {
			t.Errorf("swap, step 1: %v", err)
			return
		}
		if err := os.Symlink(external, filepath.Join(root, "hooks")); err != nil {
			t.Errorf("swap, step 2: %v", err)
		}
	}
	t.Cleanup(func() { afterMoaiHooksCheck = orig })

	removeConfigRootMoaiHooks(configRoot)

	if _, err := os.Stat(precious); err != nil {
		t.Fatalf("the removal followed the swapped link and deleted a file outside the config root: %v", err)
	}
}

// Control: with no swap, the managed hooks directory is still removed, and a
// symlinked `hooks` is still refused up front (the F1 contract).
func TestRemoveConfigRootMoaiHooks_StillRemovesARealDirectoryAndRefusesALink(t *testing.T) {
	base := t.TempDir()

	real := filepath.Join(base, "real-root")
	writeHooksFixtureFile(t, filepath.Join(real, "hooks", "moai", "hook.sh"), "managed\n")
	removeConfigRootMoaiHooks(real)
	if _, err := os.Stat(filepath.Join(real, "hooks", "moai")); !os.IsNotExist(err) {
		t.Errorf("a real hooks/moai directory was not removed (stat err %v)", err)
	}

	linked := filepath.Join(base, "linked-root")
	external := filepath.Join(base, "external")
	precious := filepath.Join(external, "moai", "precious.txt")
	writeHooksFixtureFile(t, precious, "not ours\n")
	if err := os.MkdirAll(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(linked, "hooks")); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	removeConfigRootMoaiHooks(linked)
	if _, err := os.Stat(precious); err != nil {
		t.Errorf("a symlinked hooks directory was not refused: %v", err)
	}
}
