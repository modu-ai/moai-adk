//go:build !windows

package cli

// migrate_ownership_ancestorlink_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, card-review round 4, finding A; REQ-OWR-005, AC-OWR-08 family): the
// guard against reading outside the project must look at the WHOLE path, not
// only its last element.
//
// Chunk 9 kept an unregistered file unread when the file itself was a symlink
// out of the project. A symlinked ANCESTOR (the project's `.codex`, `.claude` or
// `.agents` pointing at an external directory, e.g. a profile tree) made every
// regular file below it pass that check — the walk yields them as ordinary files
// and os.ReadFile follows the ancestor. The expected behaviour, pinned here: an
// unregistered file whose fully resolved path is outside the project root is kept
// unread (`symlink-escape`, class unregistered, `v.data` nil, a row in the
// remaining list), while a symlinked ancestor that stays INSIDE the project, a
// dangling link and an ordinary file are judged as before.
//
// How the walk reaches these files (the walk under test is
// migrateProjectCommonAssets over projectRoot/<root>): filepath.WalkDir takes
// os.Lstat of the walk root, which follows symlinked ANCESTORS, so a symlinked
// `.codex` makes the walk descend the external `agents/moai` tree and yield each
// file below it under its project-relative name. If the walk root itself is the
// symlink, Lstat reports a link, WalkDir does not descend, and it yields the link
// as one non-directory entry. Both shapes are driven below.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

const (
	acRoot       = ".codex/agents/moai"
	acTop        = acRoot + "/profile-agent.md" // below a symlinked ancestor
	acDeep       = acRoot + "/sub/deep.md"      // one level further down
	acProfileKey = "synthetic-profile-private-content\n"
)

func acWriteExternal(t *testing.T, ext string) {
	t.Helper()
	for rel, data := range map[string]string{
		"agents/moai/profile-agent.md": acProfileKey,
		"agents/moai/sub/deep.md":      acProfileKey + "deeper\n",
	} {
		p := filepath.Join(ext, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func acEnv(t *testing.T, root string, removalEnabled bool) ownershipEnv {
	t.Helper()
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return ownershipEnv{
		entry:                func(string) (*manifest.FileEntry, bool) { return nil, false },
		evidence:             fakeProvenance{},
		counterpartConfirmed: func(string) bool { return true },
		notInCatalog:         func(string) bool { return false },
		removalEnabled:       removalEnabled,
		rootResolved:         rootResolved,
	}
}

func TestUnregisteredFileBelowASymlinkedAncestorIsNotRead(t *testing.T) {
	// Verdict level.
	t.Run("AncestorLeavesTheProject", func(t *testing.T) {
		root, _ := ownershipFixtureBase(t)
		ext := filepath.Join(t.TempDir(), "profiles", "work")
		acWriteExternal(t, ext)
		if _, err := os.Lstat(filepath.Join(root, ".codex")); !os.IsNotExist(err) {
			t.Fatalf("premise: the fixture already has a .codex (%v)", err)
		}
		if err := os.Symlink(ext, filepath.Join(root, ".codex")); err != nil {
			t.Fatalf("os.Symlink failed: %v", err)
		}
		for _, enabled := range []bool{false, true} {
			env := acEnv(t, root, enabled)
			for _, rel := range []string{acTop, acDeep} {
				v := classifyOwnership(env, rel, filepath.Join(root, filepath.FromSlash(rel)))
				if len(v.data) != 0 {
					t.Errorf("switch=%v %s: the verdict function read %d bytes from outside the project", enabled, rel, len(v.data))
				}
				if v.class != ownershipUnregistered || v.keep != keepSymlinkEscape || v.removalEligible {
					t.Errorf("switch=%v %s: class=%q keep=%q eligible=%v, want unregistered/%q/false", enabled, rel, v.class, v.keep, v.removalEligible, keepSymlinkEscape)
				}
			}
		}
	})

	// Controls. A symlinked ancestor that stays INSIDE the project is kept unread too
	// (chunk 13, changed baseline); an ordinary file is judged by content; a path
	// below a dangling ancestor or to a file that is absent fails its read.
	t.Run("ControlsAreJudgedAsDecided", func(t *testing.T) {
		root := t.TempDir()
		abs := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
		write := func(rel, data string) {
			if err := os.MkdirAll(filepath.Dir(abs(rel)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(abs(rel), []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		write("real/agents/moai/in.md", "# inside, through a symlinked ancestor\n")
		write("plain/agents/moai/plain.md", "# ordinary\n")
		if err := os.Symlink(abs("real"), abs("inlink")); err != nil {
			t.Fatalf("os.Symlink failed: %v", err)
		}
		if err := os.Symlink(abs("nowhere"), abs("deadlink")); err != nil {
			t.Fatalf("os.Symlink failed: %v", err)
		}
		for _, enabled := range []bool{false, true} {
			env := acEnv(t, root, enabled)
			for _, c := range []struct {
				name, path, data string
				keep             keepReason
			}{
				{"in-project symlinked ancestor", "inlink/agents/moai/in.md", "", keepSymlinkEscape},
				{"ordinary file", "plain/agents/moai/plain.md", "# ordinary\n", keepNoProvenanceRecord},
				{"file below a dangling ancestor", "deadlink/agents/moai/x.md", "", keepReadFailed},
				{"file that does not exist", "plain/agents/moai/absent.md", "", keepReadFailed},
			} {
				v := classifyOwnership(env, c.path, abs(c.path))
				if v.class != ownershipUnregistered || v.keep != c.keep || string(v.data) != c.data {
					t.Errorf("switch=%v %s: class=%q keep=%q data=%q, want unregistered/%q/%q", enabled, c.name, v.class, v.keep, v.data, c.keep, c.data)
				}
			}
		}
	})

	// Flow level: the real walk, with a symlinked ancestor of the walk root and
	// with the walk root itself a link.
	for _, shape := range []string{"AncestorOfTheWalkRoot", "TheWalkRootItself"} {
		t.Run("Flow/"+shape, func(t *testing.T) {
			root, home := ownershipFixtureBase(t)
			ext := filepath.Join(t.TempDir(), "profiles", "work")
			acWriteExternal(t, ext)
			if shape == "AncestorOfTheWalkRoot" {
				if err := os.Symlink(ext, filepath.Join(root, ".codex")); err != nil {
					t.Fatalf("os.Symlink failed: %v", err)
				}
			} else {
				if err := os.MkdirAll(filepath.Join(root, ".codex", "agents"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(ext, "agents", "moai"), filepath.Join(root, filepath.FromSlash(acRoot))); err != nil {
					t.Fatalf("os.Symlink failed: %v", err)
				}
			}
			carried := map[string]int{}
			swapClassifier(t, func(env ownershipEnv, rel, p string) ownershipVerdict {
				v := classifyOwnership(env, rel, p)
				carried[rel] = len(v.data)
				return v
			})
			rc := &migrationRunContext{userAssetsInstalled: true}
			removed, err := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}, rc)
			if err != nil {
				t.Fatalf("migration: %v", err)
			}
			if !slices.Equal(removed, []string{migIdenticalSkill}) {
				t.Errorf("removed %v, want exactly [%s] (the reach control)", removed, migIdenticalSkill)
			}
			// Every file the walk yielded under the external tree is kept unread.
			want := []string{acTop, acDeep}
			if shape == "TheWalkRootItself" {
				want = []string{acRoot} // the walk yields the link, not what is below it
			}
			for _, rel := range want {
				r, ok := findRemaining(rc, rel)
				switch {
				case !ok:
					t.Errorf("%s has no row in the remaining list", rel)
				case r.Class != ownershipUnregistered || r.Reason != keepSymlinkEscape:
					t.Errorf("%s: row class=%q reason=%q, want unregistered/%q", rel, r.Class, r.Reason, keepSymlinkEscape)
				}
				if n, seen := carried[rel]; !seen {
					t.Errorf("%s was never classified: the no-read check would prove nothing", rel)
				} else if n != 0 {
					t.Errorf("%s: the verdict carried %d bytes read from outside the project", rel, n)
				}
			}
			for rel, n := range carried {
				if n != 0 && (rel == acTop || rel == acDeep || rel == acRoot) {
					t.Errorf("%s carried %d bytes", rel, n)
				}
			}
			// Nothing outside the project was touched.
			if got, err := os.ReadFile(filepath.Join(ext, "agents", "moai", "profile-agent.md")); err != nil || string(got) != acProfileKey {
				t.Errorf("the external file changed (%v)", err)
			}
		})
	}
}
