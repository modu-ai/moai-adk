package skillresidue

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// requireUnprivileged skips a permission test when the process can read and write
// any file regardless of mode bits, which is the case for root. Windows does not
// enforce POSIX mode bits at all, so the permission-denied paths are covered on unix.
func requireUnprivileged(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforced on Windows; permission-denied path covered on unix")
	}
	if os.Geteuid() == 0 {
		t.Skip("permission-denied paths need a non-root user")
	}
}

// openRootT opens dir as the handle the rename runs through, closed at cleanup. The rename
// takes the project-root handle and the skill folder's path relative to it; tests that
// exercise the rename alone pass "." for a root that is the skill folder.
func openRootT(t *testing.T, dir string) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("open root %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return root
}

// restoreMode sets path back to a usable mode at cleanup, so the temporary directory
// can be removed after a test made it unreadable or unwritable.
func restoreMode(t *testing.T, path string) {
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
}

// A skills root that is a regular file is not a folder to clean. The read error is
// returned, not swallowed.
func TestClean_SkillsRootThatIsAFileIsAnError(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".claude", "skills"), "not a folder")

	if _, err := Clean(root, Options{}); err == nil {
		t.Fatal("Clean on a skills root that is a file returned no error")
	}
}

// A regular file directly under the skills root is not a skill folder and is left alone.
func TestClean_SkipsRegularFilesInSkillsRoot(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".claude", "skills", "README.md"), "notes")

	actions, err := Clean(root, Options{})
	if err != nil || len(actions) != 0 {
		t.Fatalf("Clean over a regular file = %+v, %v; want no action", actions, err)
	}
	if got := readOrFail(t, filepath.Join(root, ".claude", "skills", "README.md")); got != "notes" {
		t.Fatalf("README.md changed: %q", got)
	}
}

// A retired-link path that cannot be inspected is an error, not an assumption.
func TestRetiredLinkPresent_UninspectableLinkPathIsAnError(t *testing.T) {
	requireUnprivileged(t)
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(dir, "SKILL.md"), "old body")
	mustMkdirAll(t, filepath.Join(root, ".agents", "skills"))
	if err := os.Chmod(filepath.Join(root, ".agents"), 0o000); err != nil {
		t.Fatal(err)
	}
	restoreMode(t, filepath.Join(root, ".agents"))
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	if _, err := Clean(root, Options{Shipped: shipped}); err == nil {
		t.Fatal("Clean with an uninspectable link path returned no error")
	}
}

// A real folder at the retired-link path is not a link and is not reported.
func TestRetiredLinkPresent_RealFolderAtLinkPathIsNotReported(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(dir, "SKILL.md"), "old body")
	realFolder := filepath.Join(root, ".agents", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(realFolder, "SKILL.md"), "agent-side copy")
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped})
	if err != nil {
		t.Fatal(err)
	}
	if hasAction(actions, "reported-retired-link", ".agents/skills/moai-kanban-foreman") {
		t.Fatalf("a real folder was reported as a retired link: %+v", actions)
	}
	if got := readOrFail(t, filepath.Join(realFolder, "SKILL.md")); got != "agent-side copy" {
		t.Fatalf("real folder changed: %q", got)
	}
}

// A shipped skill's empty folder that cannot be removed is an error from the cleanup.
func TestClean_EmptyFolderRemovalFailureIsReturned(t *testing.T) {
	requireUnprivileged(t)
	root := t.TempDir()
	skills := filepath.Join(root, ".claude", "skills")
	mustMkdirAll(t, filepath.Join(skills, "moai-factory-foreman"))
	if err := os.Chmod(skills, 0o555); err != nil {
		t.Fatal(err)
	}
	restoreMode(t, skills)
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	if _, err := Clean(root, Options{Shipped: shipped}); err == nil {
		t.Fatal("Clean with an unremovable empty folder returned no error")
	}
}

// A retired folder whose contents cannot be read is kept in place.
func TestClean_UnreadableRetiredFolderIsKept(t *testing.T) {
	requireUnprivileged(t)
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(dir, "SKILL.md"), "old body")
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	restoreMode(t, dir)
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped})
	if err != nil {
		t.Fatal(err)
	}
	if !hasAction(actions, "kept-retired-skill-with-user-files", ".claude/skills/moai-kanban-foreman") {
		t.Fatalf("unreadable retired folder not kept: %+v", actions)
	}
}

// An unreadable skill folder is an error from the cleanup, not a silent skip.
func TestClean_UnreadableFolderIsAnError(t *testing.T) {
	requireUnprivileged(t)
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-ref-api-patterns")
	mustWriteFile(t, filepath.Join(dir, "SKILL.md"), "body")
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	restoreMode(t, dir)

	if _, err := Clean(root, Options{}); err == nil {
		t.Fatal("Clean over an unreadable folder returned no error")
	}
}

// A failed first move leaves the lowercase file where it was, returns the error, and
// leaves no staging folder behind.
func TestClean_RenameFailureIsReturnedAndNothingMoves(t *testing.T) {
	forceFallback(t) // injects the failure through the link seam
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-ref-owasp-checklist")
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-ref-owasp-checklist/SKILL.md" }
	orig := linkFile
	t.Cleanup(func() { linkFile = orig })
	linkFile = func(_ *os.Root, from, to string) error { return errors.New("injected failure on the first move") }

	if _, err := Clean(root, Options{Shipped: shipped}); err == nil {
		t.Fatal("Clean with a failing rename returned no error")
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "lower body" {
		t.Fatalf("skill.md changed by a failed rename: %q", got)
	}
	if names := exactNames(t, dir); len(names) != 1 {
		t.Fatalf("a failed first move left extra entries behind: %v", names)
	}
}

// When no temporary name is free, the rename is refused and nothing changes.
func TestRenameSkillFile_NoFreeTemporaryNameIsRefused(t *testing.T) {
	legacyStageNames(t) // the deterministic names are the ones the test fills
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	for n := range 100 {
		name := "skill.md.tmp-rename"
		if n > 0 {
			name = fmt.Sprintf("skill.md.tmp-rename-%d", n)
		}
		mustWriteFile(t, filepath.Join(dir, name), "user data")
	}

	if err := renameSkillFile(openRootT(t, dir), "."); err == nil {
		t.Fatal("renameSkillFile succeeded with no free temporary name")
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "lower body" {
		t.Fatalf("skill.md changed: %q", got)
	}
}

// When the link to SKILL.md and the restore both fail, the error names both and the
// staging folder, so a person knows where the content went.
func TestRenameSkillFile_RestoreFailureIsReported(t *testing.T) {
	forceFallback(t) // drives the link seam and addresses the stage folder by literal path
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	origLink := linkFile
	t.Cleanup(func() { linkFile = origLink })
	calls := 0
	// The first link (into the staging folder) works; every later link fails, so the
	// restore has no way back.
	linkFile = func(root *os.Root, from, to string) error {
		calls++
		if calls == 1 {
			return origLink(root, from, to)
		}
		return errors.New("injected link refusal")
	}

	err := renameSkillFile(openRootT(t, dir), ".")
	if err == nil || !strings.Contains(err.Error(), "restoring") || !strings.Contains(err.Error(), "skill.md.tmp-rename") {
		t.Fatalf("renameSkillFile error = %v; want one that reports the failed restore and where the content is", err)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md.tmp-rename", "skill.md")); got != "lower body" {
		t.Fatalf("the content is not where the error says: %q", got)
	}
}

// A folder with no search permission cannot be inspected, so the rename fails with that
// error before it moves anything.
func TestRenameSkillFile_UninspectableDirectoryIsAnError(t *testing.T) {
	requireUnprivileged(t)
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	if err := os.Chmod(dir, 0o400); err != nil {
		t.Fatal(err)
	}
	restoreMode(t, dir)

	if err := renameSkillFile(openRootT(t, dir), "."); err == nil {
		t.Fatal("renameSkillFile succeeded in an uninspectable folder")
	}
}
