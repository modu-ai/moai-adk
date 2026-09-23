package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestReleaseOwnMirrorLink(t *testing.T) {
	const skill = "moai"
	rel := filepath.ToSlash(filepath.Join(MirrorSkillsRelDir, skill, "SKILL.md"))

	cases := []struct {
		name       string
		setup      func(t *testing.T, link string)
		wantExists bool
	}{
		{"own dangling link removed", func(t *testing.T, link string) {
			mustSymlink(t, MirrorLinkTarget(skill), link)
		}, false},
		{"foreign link kept", func(t *testing.T, link string) {
			mustSymlink(t, "../../elsewhere/"+skill, link)
		}, true},
		{"real directory kept", func(t *testing.T, link string) {
			if err := os.MkdirAll(link, 0o755); err != nil {
				t.Fatal(err)
			}
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			link := filepath.Join(root, MirrorSkillsRelDir, skill)
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				t.Fatal(err)
			}
			tc.setup(t, link)
			if err := releaseOwnMirrorLink(root, rel); err != nil {
				t.Fatalf("releaseOwnMirrorLink: %v", err)
			}
			_, err := os.Lstat(link)
			if exists := err == nil; exists != tc.wantExists {
				t.Fatalf("entry exists=%v, want %v", exists, tc.wantExists)
			}
		})
	}

	t.Run("paths outside the mirror are ignored", func(t *testing.T) {
		root := t.TempDir()
		link := filepath.Join(root, MirrorSkillsRelDir, skill)
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			t.Fatal(err)
		}
		mustSymlink(t, MirrorLinkTarget(skill), link)
		for _, p := range []string{".claude/skills/moai/SKILL.md", filepath.ToSlash(filepath.Join(MirrorSkillsRelDir, skill))} {
			if err := releaseOwnMirrorLink(root, p); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := os.Lstat(link); err != nil {
			t.Fatalf("link removed for a path outside a mirror entry: %v", err)
		}
	})
}

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
}

// TestDeployReleasesOwnMirrorLinkBeforeExistenceCheck reproduces the
// init --force --llm both path over a claude deployment: the non-force
// existence check stats .agents/skills/<skill>/SKILL.md through MoAI's own
// mirror link, finds the canonical file behind it, and without the release
// running first it records the path user_created and leaves the link in place.
func TestDeployReleasesOwnMirrorLinkBeforeExistenceCheck(t *testing.T) {
	const skill = "moai"
	rel := filepath.ToSlash(filepath.Join(MirrorSkillsRelDir, skill, "SKILL.md"))
	root, mgr := setupDeployProject(t)

	canonical := filepath.Join(root, CanonicalSkillsRelDir, skill, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(canonical), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(canonical, []byte("canonical"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, MirrorSkillsRelDir, skill)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	mustSymlink(t, MirrorLinkTarget(skill), link)

	fsys := fstest.MapFS{rel: &fstest.MapFile{Data: []byte("mirror")}}
	if err := NewDeployer(fsys).Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat %s: %v", link, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%s is still MoAI's mirror link; want a real directory", link)
	}
	if got, _ := os.ReadFile(filepath.Join(link, "SKILL.md")); string(got) != "mirror" {
		t.Fatalf("mirror SKILL.md = %q, want %q", got, "mirror")
	}
	if got, _ := os.ReadFile(canonical); string(got) != "canonical" {
		t.Fatalf("canonical SKILL.md = %q, want it untouched", got)
	}
	entry, found := mgr.GetEntry(rel)
	if !found || entry.Provenance != manifest.TemplateManaged {
		t.Fatalf("manifest entry for %s = %+v (found=%v), want template_managed", rel, entry, found)
	}
}
