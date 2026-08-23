package template

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"testing/fstest"
)

// mirrorEntryNames lists the first-level entry names under .agents/skills/.
// It deliberately does NOT filter on d.IsDir(): fs.DirEntry is Lstat-based, so
// a symlink pointing at a directory reports IsDir()==false and would silently
// vanish from the collected set.
func mirrorEntryNames(t *testing.T, root string) []string {
	t.Helper()
	return firstLevelNames(t, filepath.Join(root, ".agents", "skills"))
}

// canonicalSkillNames lists the first-level entry names under .claude/skills/.
func canonicalSkillNames(t *testing.T, root string) []string {
	t.Helper()
	return firstLevelNames(t, filepath.Join(root, ".claude", "skills"))
}

func firstLevelNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read dir %q: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSkillMirror_SetIsDerivedNotConstant covers AC-CSC-014: the mirror set is
// derived from what this run actually deployed, never from a constant or from
// the real catalog.
func TestSkillMirror_SetIsDerivedNotConstant(t *testing.T) {
	root, mgr := setupDeployProject(t)
	synthetic := fstest.MapFS{
		".claude/skills/moai-alpha/SKILL.md": &fstest.MapFile{Data: []byte("alpha")},
		".claude/skills/moai-beta/SKILL.md":  &fstest.MapFile{Data: []byte("beta")},
		".claude/rules/unrelated.md":         &fstest.MapFile{Data: []byte("rule")},
	}

	d := NewDeployer(synthetic)
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	got := mirrorEntryNames(t, root)
	want := []string{"moai-alpha", "moai-beta"}
	if !sameStringSlice(got, want) {
		t.Errorf("mirror set = %v, want exactly %v (a larger set means the code consults a constant or the real catalog)", got, want)
	}
}

// TestSkillMirror_SlimSetEqualsCanonicalAndIsSmaller covers AC-CSC-003.
func TestSkillMirror_SlimSetEqualsCanonicalAndIsSmaller(t *testing.T) {
	cat, err := LoadCatalog(embeddedRaw)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	slim, err := SlimFS(embeddedRaw, cat)
	if err != nil {
		t.Fatalf("slim fs: %v", err)
	}
	full, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}

	slimRoot, slimMgr := setupDeployProject(t)
	if err := NewDeployer(slim).Deploy(context.Background(), slimRoot, slimMgr, nil); err != nil {
		t.Fatalf("slim Deploy: %v", err)
	}
	fullRoot, fullMgr := setupDeployProject(t)
	if err := NewDeployer(full).Deploy(context.Background(), fullRoot, fullMgr, nil); err != nil {
		t.Fatalf("full Deploy: %v", err)
	}

	slimMirror := mirrorEntryNames(t, slimRoot)
	slimCanonical := canonicalSkillNames(t, slimRoot)
	fullMirror := mirrorEntryNames(t, fullRoot)

	// 1. slim mirror set == slim canonical set
	if !sameStringSlice(slimMirror, slimCanonical) {
		t.Errorf("slim mirror set %v != slim canonical set %v", slimMirror, slimCanonical)
	}

	// 2. no non-core name leaked into the slim mirror (zero dangling links)
	canonical := make(map[string]struct{}, len(slimCanonical))
	for _, n := range slimCanonical {
		canonical[n] = struct{}{}
	}
	for _, n := range slimMirror {
		if _, ok := canonical[n]; !ok {
			t.Errorf("slim mirror contains %q which is not deployed under .claude/skills (dangling link)", n)
		}
	}

	// 3. the slim set is strictly smaller than the full set — this is what
	//    proves the tier filter was actually traversed; assertion 1 alone is
	//    true by construction for any derive-from-target implementation.
	if len(slimMirror) >= len(fullMirror) {
		t.Errorf("slim mirror count %d must be < full mirror count %d", len(slimMirror), len(fullMirror))
	}
}

// TestSkillMirror_BothPathsReadableAfterFullDeploy covers AC-CSC-002.
func TestSkillMirror_BothPathsReadableAfterFullDeploy(t *testing.T) {
	full, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	root, mgr := setupDeployProject(t)
	if err := NewDeployer(full).Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	names := canonicalSkillNames(t, root)
	if len(names) == 0 {
		t.Fatal("no skills deployed — fixture assumption broken")
	}
	for _, name := range names {
		canonicalPath := filepath.Join(root, ".claude", "skills", name, "SKILL.md")
		mirrorPath := filepath.Join(root, ".agents", "skills", name, "SKILL.md")
		want, readErr := os.ReadFile(canonicalPath)
		if readErr != nil {
			t.Errorf("read canonical %q: %v", canonicalPath, readErr)
			continue
		}
		got, readErr := os.ReadFile(mirrorPath)
		if readErr != nil {
			t.Errorf("read mirror %q: %v", mirrorPath, readErr)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("mirror content for %q differs from canonical", name)
		}
	}
}
