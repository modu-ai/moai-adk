package skillresidue

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A user file that already holds the temporary name the lowercase rename uses
// must survive the rename. os.Rename replaces an existing target, so the rename
// may not take a name that a user file holds.
func TestClean_TempRenameNeverOverwritesUserFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-ref-owasp-checklist")
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	mustWriteFile(t, filepath.Join(dir, "skill.md.tmp-rename"), "user data")
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-ref-owasp-checklist/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped})
	if err != nil {
		t.Fatal(err)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md.tmp-rename")); got != "user data" {
		t.Fatalf("user file overwritten: got %q", got)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "lower body" {
		t.Fatalf("SKILL.md = %q, want the lowercase body", got)
	}
	if !hasAction(actions, "renamed-skill-file", ".claude/skills/moai-ref-owasp-checklist/skill.md") {
		t.Fatalf("rename not reported: %+v", actions)
	}
}

// A .claude/skills that is itself a symbolic link leads out of the project. Clean
// must not move a folder out of the shared directory it points at, and must say
// that it refused.
func TestClean_RefusesSymlinkedSkillsRoot(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(t.TempDir(), "shared-skills")
	mustWriteFile(t, filepath.Join(shared, "moai-kanban-foreman", "SKILL.md"), "shared body")
	mustMkdirAll(t, filepath.Join(root, ".claude"))
	if err := os.Symlink(shared, filepath.Join(root, ".claude", "skills")); err != nil {
		t.Skip(err)
	}
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(shared, "moai-kanban-foreman", "SKILL.md")); err != nil {
		t.Fatalf("shared folder moved out of the shared directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "archive")); !os.IsNotExist(err) {
		t.Fatalf("archive created through a symlinked skills root (stat err: %v)", err)
	}
	if !hasAction(actions, "refused-symlinked-root", ".claude/skills") {
		t.Fatalf("refusal not reported: %+v", actions)
	}
}

// A link under .agents/skills whose target is another project's folder of the
// same name is not residue of this project. Clean reports only the link to the
// retired folder it found here, and reports nothing about the foreign one.
func TestClean_KeepsLinkToAnotherProjectsFolder(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".claude", "skills", "moai-kanban-foreman", "SKILL.md"), "old body")
	foreign := filepath.Join(other, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(foreign, "SKILL.md"), "other project live body")
	link := filepath.Join(root, ".agents", "skills", "moai-kanban-foreman")
	mustMkdirAll(t, filepath.Dir(link))
	if err := os.Symlink(foreign, link); err != nil {
		t.Skip(err)
	}
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("link to another project's folder removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(foreign, "SKILL.md")); err != nil {
		t.Fatalf("another project's folder changed: %v", err)
	}
	if hasAction(actions, "reported-retired-link", ".agents/skills/moai-kanban-foreman") {
		t.Fatalf("foreign link reported as this project's retired link: %+v", actions)
	}
}

// A link to SKILL.md that fails must put the lowercase file back under its own name.
// Otherwise its content stays under the staging folder, where nothing reads it.
func TestRenameSkillFile_RestoresLowercaseNameWhenTheLinkIsRefused(t *testing.T) {
	forceFallback(t) // injects the refusal through the link seam
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	orig := linkFile
	t.Cleanup(func() { linkFile = orig })
	linkFile = func(*os.Root, string, string) error {
		return errors.New("injected failure on the link")
	}

	if err := renameSkillFile(openRootT(t, dir), "."); err == nil {
		t.Fatal("renameSkillFile reported success although the link failed")
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "lower body" {
		t.Fatalf("skill.md = %q after the failed rename, want the original body", got)
	}
	if _, err := os.Lstat(filepath.Join(dir, "skill.md.tmp-rename")); !os.IsNotExist(err) {
		t.Fatalf("staging folder left behind after the failed rename (stat err: %v)", err)
	}
}
