package template

import (
	"os"
	"path/filepath"
	"testing"
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
