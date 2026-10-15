package cli

// t1619_codex_skills_disable_coverage_test.go — card t1619, bundle B1: the unexecuted
// lines of codex_skills_disable.go (the user-install skill faces and the fallback in
// resolveCodexSkillMirrorPath). The existing tests for this file sit outside every
// narrowed family, so these tests carry the new family prefix TestCodexSkillDisable.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// t1619SkillFile writes <root>/<rel>/<skill>/SKILL.md as a regular file.
func t1619SkillFile(t *testing.T, root, rel, skill string) string {
	t.Helper()
	file := filepath.Join(root, rel, skill, "SKILL.md")
	t1619MustWrite(t, file, "# skill\n")
	return file
}

// TestCodexSkillDisableWithoutHomeHasNoUserFaces — with no home there are no user-install
// faces to fall back to, so a name resolves to nothing when the project mirror is absent.
func TestCodexSkillDisableWithoutHomeHasNoUserFaces(t *testing.T) {
	if faces := userInstallSkillFaces(""); len(faces) != 0 {
		t.Fatalf("an empty home must name no user-install faces, got %q", faces)
	}
	res := resolveCodexSkillMirrorPath(t.TempDir(), "", "demo-skill")
	if res.Outcome != codexSkillMirrorAbsent {
		t.Fatalf("with no project mirror and no home the name must not resolve: outcome %v (reason %q)",
			res.Outcome, res.Reason)
	}
}

// TestCodexSkillDisableUserFacesListCodexFaceFirst — the user-install faces are the Codex
// face first, then the Claude face, so a two-copies install disables the Codex copy.
func TestCodexSkillDisableUserFacesListCodexFaceFirst(t *testing.T) {
	home := t.TempDir()
	want := []string{
		filepath.Join(home, template.MirrorSkillsRelDir),
		filepath.Join(home, ".claude", "skills"),
	}
	if got := userInstallSkillFaces(home); !slices.Equal(got, want) {
		t.Fatalf("user-install faces = %q, want %q", got, want)
	}
}

// TestCodexSkillDisableProjectMirrorSuppressesUserFaces — while the project mirror exists
// the user-install faces do not apply: a skill present only under the user face is not
// resolved through the project mirror path.
func TestCodexSkillDisableProjectMirrorSuppressesUserFaces(t *testing.T) {
	projectRoot := t.TempDir()
	home := t.TempDir()
	t1619MustMkdir(t, filepath.Join(projectRoot, template.MirrorSkillsRelDir))
	t1619SkillFile(t, home, template.MirrorSkillsRelDir, "demo")

	res := resolveCodexSkillMirrorPath(projectRoot, home, "demo")
	if res.Outcome == codexSkillResolved {
		t.Fatalf("a skill carried only by the user face must not resolve while the project mirror exists: %q", res.Path)
	}
	if res.Outcome != codexSkillUnresolved {
		t.Fatalf("with the project mirror present and no project SKILL.md the name is unresolved, got outcome %v (reason %q)",
			res.Outcome, res.Reason)
	}
}

// TestCodexSkillDisableUserCodexFaceResolvesRegularSkillFile — with no project mirror, a
// regular SKILL.md under the Codex user face resolves to that file.
func TestCodexSkillDisableUserCodexFaceResolvesRegularSkillFile(t *testing.T) {
	projectRoot := t.TempDir()
	home := t.TempDir()
	file := t1619SkillFile(t, home, template.MirrorSkillsRelDir, "demo")

	res := resolveCodexSkillMirrorPath(projectRoot, home, "demo")
	if res.Outcome != codexSkillResolved || res.Path != file {
		t.Fatalf("a user-installed Codex-face skill must resolve to its file: outcome %v, path %q (want %q)",
			res.Outcome, res.Path, file)
	}
}

// TestCodexSkillDisableUserClaudeFaceResolvesWhenCodexFaceAbsent — the Claude user face is
// consulted when the Codex face does not carry the skill.
func TestCodexSkillDisableUserClaudeFaceResolvesWhenCodexFaceAbsent(t *testing.T) {
	projectRoot := t.TempDir()
	home := t.TempDir()
	file := t1619SkillFile(t, home, filepath.Join(".claude", "skills"), "demo")

	res := resolveCodexSkillMirrorPath(projectRoot, home, "demo")
	if res.Outcome != codexSkillResolved || res.Path != file {
		t.Fatalf("a skill carried only by the Claude face must resolve to it: outcome %v, path %q (want %q)",
			res.Outcome, res.Path, file)
	}
}

// TestCodexSkillDisableDirectoryShapedSkillIsNotResolved — a SKILL.md that is a directory
// is not a skill file, so the user faces do not resolve the name.
func TestCodexSkillDisableDirectoryShapedSkillIsNotResolved(t *testing.T) {
	projectRoot := t.TempDir()
	home := t.TempDir()
	t1619MustMkdir(t, filepath.Join(home, template.MirrorSkillsRelDir, "demo", "SKILL.md"))

	res := resolveCodexSkillMirrorPath(projectRoot, home, "demo")
	if res.Outcome == codexSkillResolved {
		t.Fatalf("a directory-shaped SKILL.md must not resolve: path %q", res.Path)
	}
	if _, err := os.Stat(filepath.Join(home, template.MirrorSkillsRelDir, "demo", "SKILL.md")); err != nil {
		t.Fatalf("the directory-shaped entry vanished: %v", err)
	}
}
