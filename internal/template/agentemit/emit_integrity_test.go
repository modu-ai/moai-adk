// emit_integrity_test.go — P2 review repairs on PR #1843 (card t1649): the
// emitter never follows a symlinked agent file, the read-only check refuses a
// symlink at a generated path of a partial set (full sets were already covered
// by the orphan walk), and Update removes the rendered copies of deleted
// agent files without touching partial sets or their hand-managed neighbours.
package agentemit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// TestEmitRejectsSymlinkedAgent pins that a source entry that is not a regular
// file is refused with a named error instead of being read through the link.
func TestEmitRejectsSymlinkedAgent(t *testing.T) {
	man := mustManifest(t)

	t.Run("regular_sources_still_render", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update over regular sources: %v", err)
		}
		if sum, err := agentemit.Check(root, lay, man); err != nil || sum.String() != "specs=2 copies=4 diff=0" {
			t.Errorf("regular sources: %q, %v", sum.String(), err)
		}
	})

	t.Run("internal_symlink_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		link := filepath.Join(root, "src", "gamma.md")
		if err := os.Symlink(filepath.Join(root, "src", "alpha.md"), link); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "gamma.md") || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("a symlinked agent file must be refused naming its path, got %v", err)
		}
	})

	t.Run("external_symlink_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		twin := filepath.Join(t.TempDir(), "twin.md")
		writeFile(t, twin, []byte(twoSources()["beta.md"]))
		if err := os.Remove(filepath.Join(root, "src", "beta.md")); err != nil {
			t.Fatalf("remove: %v", err)
		}
		if err := os.Symlink(twin, filepath.Join(root, "src", "beta.md")); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "beta.md") || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("an externally symlinked agent file must be refused naming its path, got %v", err)
		}
		if got, rerr := os.ReadFile(twin); rerr != nil || len(got) == 0 {
			t.Errorf("the file behind the source symlink was disturbed: %v", rerr)
		}
	})
}

// TestCheckRefusesSymlinkAtGeneratedPath pins that the read-only check never
// accepts a symlink sitting at a generated path: reading through the link
// would compare bytes the tree does not actually hold.
func TestCheckRefusesSymlinkAtGeneratedPath(t *testing.T) {
	man := mustManifest(t)
	// twin writes the freshly rendered bytes of rel outside the tree and
	// returns the outside path, so a link at the generated path points at a
	// byte-identical file — accepted today, refused by the fix.
	twin := func(t *testing.T, root, rel string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read rendered %s: %v", rel, err)
		}
		p := filepath.Join(t.TempDir(), "twin")
		writeFile(t, p, data)
		return p
	}

	t.Run("partial_set_symlink_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		target := filepath.Join(root, "proj", "c1", "alpha.md")
		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(twin(t, root, "out/c2/alpha.md"), target); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		sum, err := agentemit.Check(root, lay, man)
		if err == nil {
			t.Fatal("a symlink at a partial-set generated path must fail the check")
		}
		if !strings.Contains(strings.Join(sum.Diffs, "\n"), "proj/c1/alpha.md") {
			t.Errorf("the failure must name proj/c1/alpha.md, got %v", sum.Diffs)
		}
	})

	t.Run("full_set_symlink_at_generated_path_named", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		target := filepath.Join(root, "out", "c2", "alpha.md")
		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(twin(t, root, "out/c2/beta.md"), target); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		sum, err := agentemit.Check(root, lay, man)
		if err == nil {
			t.Fatal("a symlink at a generated path must fail the check")
		}
		if !strings.Contains(strings.Join(sum.Diffs, "\n"), "out/c2/alpha.md") {
			t.Errorf("the failure must name out/c2/alpha.md, got %v", sum.Diffs)
		}
	})

	t.Run("hand_managed_neighbors_still_pass", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		writeFile(t, filepath.Join(root, "proj/c1/hand-managed.md"), []byte("hand\n"))
		writeFile(t, filepath.Join(root, "proj/c1/nested/hand.md"), []byte("hand\n"))
		if sum, err := agentemit.Check(root, lay, man); err != nil {
			t.Errorf("hand-managed files beside a partial set must pass: %q, %v", sum.String(), err)
		}
	})
}

// TestUpdateRemovesObsoleteOutputs pins that regenerating a full set removes
// the rendered copies of deleted agent files, without touching partial sets or
// their hand-managed neighbours, and refusing to remove through a link or any
// non-regular entry.
func TestUpdateRemovesObsoleteOutputs(t *testing.T) {
	man := mustManifest(t)
	generated := func(t *testing.T) (string, agentemit.Layout) {
		t.Helper()
		root, lay := fixtureRoot(t, twoSources())
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		return root, lay
	}
	deleteBeta := func(t *testing.T, root string) {
		t.Helper()
		if err := os.Remove(filepath.Join(root, "src", "beta.md")); err != nil {
			t.Fatalf("delete agent file: %v", err)
		}
	}

	t.Run("deleted_agent_leaves_no_rendered_copy", func(t *testing.T) {
		root, lay := generated(t)
		deleteBeta(t, root)
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update after deleting an agent file: %v", err)
		}
		if got := listNames(t, filepath.Join(root, "out/c2"), ".md"); strings.Join(got, ",") != "alpha" {
			t.Errorf("C2 after deletion = %v, want only alpha", got)
		}
		if got := listNames(t, filepath.Join(root, "out/c3"), ".toml"); strings.Join(got, ",") != "alpha" {
			t.Errorf("C3 after deletion = %v, want only alpha", got)
		}
		if sum, err := agentemit.Check(root, lay, man); err != nil {
			t.Errorf("Check after stale removal: %q, %v", sum.String(), err)
		}
	})

	t.Run("partial_sets_and_hand_managed_files_untouched", func(t *testing.T) {
		root, lay := generated(t)
		writeFile(t, filepath.Join(root, "proj/c1/hand-managed.md"), []byte("hand\n"))
		deleteBeta(t, root)
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if got := listNames(t, filepath.Join(root, "proj/c1"), ".md"); strings.Join(got, ",") != "alpha,hand-managed" {
			t.Errorf("project set after deletion = %v, want hand-managed files preserved", got)
		}
	})

	t.Run("symlink_at_stale_path_refused", func(t *testing.T) {
		root, lay := generated(t)
		victim := filepath.Join(t.TempDir(), "victim.md")
		writeFile(t, victim, []byte("outside file: must stay untouched\n"))
		stale := filepath.Join(root, "out/c2/beta.md")
		if err := os.Remove(stale); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(victim, stale); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		deleteBeta(t, root)
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "beta.md") || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("Update must refuse to remove through a symlink naming the path, got %v", err)
		}
		if got, rerr := os.ReadFile(victim); rerr != nil || string(got) != "outside file: must stay untouched\n" {
			t.Errorf("the file behind the stale symlink changed: %q, %v", got, rerr)
		}
		if fi, lerr := os.Lstat(stale); lerr != nil || fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("the refused symlink must stay in place, got %v, %v", fi, lerr)
		}
	})

	t.Run("directory_at_stale_path_refused", func(t *testing.T) {
		root, lay := generated(t)
		stale := filepath.Join(root, "out/c3/beta.toml")
		if err := os.Remove(stale); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(stale, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		deleteBeta(t, root)
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "beta.toml") {
			t.Fatalf("Update must refuse a non-regular stale entry naming it, got %v", err)
		}
		if fi, lerr := os.Lstat(stale); lerr != nil || !fi.IsDir() {
			t.Errorf("the refused directory must stay in place, got %v, %v", fi, lerr)
		}
	})
}
