package cli

// migrate_ownership_env_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M7 (card
// t1865): the fail-closed edges of the preview's verdict environment
// (migrate_ownership_env.go) and the guards of the nested empty-directory prune.
// These are behavior the AC-OWR-16 test relies on but does not exercise: when the
// preview cannot assemble what the verdict needs it announces NOTHING — it never
// guesses a removal — and the nested prune takes only an empty directory inside
// the project root.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestPreviewVerdictsFailClosed: with the removal switch on, an eligible file is
// announced when every input is readable, and none is announced when an input is
// not.
func TestPreviewVerdictsFailClosed(t *testing.T) {
	enableRemovalForTest(t)
	root, home := previewFixture(t, false)
	mf := manifest.NewManager()
	if _, err := mf.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	loaded := mf.Manifest()

	ask := func(t *testing.T, m *manifest.Manifest, rel string) bool {
		t.Helper()
		var v previewVerdicts
		return v.removalEligible(root, m, rel)
	}

	t.Run("ReachControlAnnouncesWhenEverythingReads", func(t *testing.T) {
		if !ask(t, loaded, migIdenticalSkill) {
			t.Fatalf("%s is removal-eligible in this fixture, yet the preview verdict says no: the cases below would prove nothing", migIdenticalSkill)
		}
	})

	t.Run("NoHomeMeansNoCandidate", func(t *testing.T) {
		prev := userHomeDirFn
		t.Cleanup(func() { userHomeDirFn = prev })
		userHomeDirFn = func() (string, error) { return "", errors.New("no home") }
		if ask(t, loaded, migIdenticalSkill) {
			t.Errorf("the preview announced %s although it could not resolve the home directory", migIdenticalSkill)
		}
	})

	t.Run("CorruptUserManifestMeansNoCandidate", func(t *testing.T) {
		path := filepath.Join(home, ".moai", "user-assets.json")
		orig, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read user manifest: %v", err)
		}
		t.Cleanup(func() { _ = os.WriteFile(path, orig, 0o644) })
		if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		if ask(t, loaded, migIdenticalSkill) {
			t.Errorf("the preview announced %s although the user manifest is corrupt", migIdenticalSkill)
		}
	})

	t.Run("UnresolvableProjectRootMeansNoCandidate", func(t *testing.T) {
		var v previewVerdicts
		if v.removalEligible(filepath.Join(root, "no-such-directory"), loaded, migIdenticalSkill) {
			t.Errorf("the preview announced a file of a project root that does not resolve")
		}
	})

	t.Run("NoManifestMeansNoCandidate", func(t *testing.T) {
		// An absent or corrupt manifest reads as nil: nothing is registered, so
		// nothing is promoted and nothing is announced.
		if ask(t, nil, migIdenticalSkill) {
			t.Errorf("the preview announced %s with no manifest record at all", migIdenticalSkill)
		}
	})

	t.Run("EnvironmentIsBuiltOnceAndOnlyWhenAsked", func(t *testing.T) {
		var v previewVerdicts
		if v.built {
			t.Fatalf("a fresh previewVerdicts has already built its environment")
		}
		_ = v.removalEligible(root, loaded, migIdenticalSkill)
		if !v.built || !v.ok {
			t.Fatalf("after one question built=%v ok=%v, want both true", v.built, v.ok)
		}
		prev := userHomeDirFn
		t.Cleanup(func() { userHomeDirFn = prev })
		userHomeDirFn = func() (string, error) { return "", errors.New("no home") }
		if !v.removalEligible(root, loaded, migIdenticalSkill) {
			t.Errorf("a second question rebuilt the environment instead of reusing the first")
		}
	})
}

// TestViewOfManifestIsAReadOnlyCopy: the preview's manifest view hands out copies
// of entries, so judging a file can never change the manifest it was loaded from.
func TestViewOfManifestIsAReadOnlyCopy(t *testing.T) {
	mf := &manifest.Manifest{Files: map[string]manifest.FileEntry{
		".claude/skills/a/SKILL.md": {Provenance: manifest.TemplateManaged, TemplateHash: "h"},
		".claude/skills/a/Extra.md": {Provenance: manifest.TemplateManaged, TemplateHash: "h2"},
	}}
	v := viewOfManifest(mf)
	e, ok := v.entry(".claude/skills/a/SKILL.md")
	if !ok || e.TemplateHash != "h" {
		t.Fatalf("entry lookup = %+v, %v", e, ok)
	}
	e.TemplateHash = "changed"
	if mf.Files[".claude/skills/a/SKILL.md"].TemplateHash != "h" {
		t.Errorf("mutating the returned entry changed the manifest it came from")
	}
	if _, ok := v.entry(".claude/skills/a/none.md"); ok {
		t.Errorf("an absent path was found")
	}
	if got := len(v.keysByDir[".claude/skills/a"]); got != 2 {
		t.Errorf("keysByDir holds %d keys for the directory, want 2", got)
	}
	empty := viewOfManifest(nil)
	if _, ok := empty.entry("x"); ok || len(empty.keysByDir) != 0 {
		t.Errorf("a nil manifest must read as empty")
	}
}

// TestPruneNestedEmptyDirGuards: the deeper-level prune takes only an empty
// directory that stays inside the resolved project root.
func TestPruneNestedEmptyDirGuards(t *testing.T) {
	root := t.TempDir()
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	mk := func(rel string) string {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("RemovesAnEmptyDirectory", func(t *testing.T) {
		d := mk(".claude/skills/s/refs")
		pruned, err := pruneNestedEmptyDir(rootResolved, d)
		if err != nil || !pruned {
			t.Fatalf("pruned=%v err=%v, want the empty directory removed", pruned, err)
		}
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("the directory is still there (%v)", err)
		}
	})

	t.Run("KeepsADirectoryWithContent", func(t *testing.T) {
		d := mk(".claude/skills/s2/refs")
		if err := os.WriteFile(filepath.Join(d, "keep.md"), []byte("mine"), 0o644); err != nil {
			t.Fatal(err)
		}
		pruned, err := pruneNestedEmptyDir(rootResolved, d)
		if err != nil || pruned {
			t.Fatalf("pruned=%v err=%v, want the directory with content kept", pruned, err)
		}
		if _, err := os.Stat(filepath.Join(d, "keep.md")); err != nil {
			t.Errorf("the content vanished: %v", err)
		}
	})

	t.Run("RefusesAnEmptyDirectoryOutsideTheRoot", func(t *testing.T) {
		// The boundary guard holds without a link: an empty directory that is not
		// under the resolved project root is not taken, whatever path reached it.
		outside := t.TempDir()
		pruned, err := pruneNestedEmptyDir(rootResolved, outside)
		if err != nil || pruned {
			t.Fatalf("pruned=%v err=%v, want the directory outside the root refused", pruned, err)
		}
		if _, err := os.Stat(outside); err != nil {
			t.Errorf("the directory outside the root was removed: %v", err)
		}
	})

	t.Run("AMissingDirectoryIsNothingToDo", func(t *testing.T) {
		pruned, err := pruneNestedEmptyDir(rootResolved, filepath.Join(root, "gone"))
		if err != nil || pruned {
			t.Fatalf("pruned=%v err=%v for a directory that does not exist", pruned, err)
		}
	})
}
