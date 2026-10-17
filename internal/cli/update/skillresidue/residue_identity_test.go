package skillresidue

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-UBR-17 (audit findings F11 and F12 of card t1733): the held-back rename compares each
// of its names with the identity of the file it started from, and removes a name only while
// that name is still that file. These tests drive the states listed in the table above
// renameSkillFile by running a second actor right after the first link or after a later
// step, through the same seams the atomic-save test uses.

const (
	idLower  = "skill.md"
	idStaged = "skill.md.tmp-rename/skill.md"
)

// afterNthLink installs a linkFile that runs act(n) after the n-th call (1-based) when the
// real link succeeded. fail, when non-nil for n, replaces the real link by that error.
func afterNthLink(t *testing.T, act func(n int), fail func(n int) error) {
	t.Helper()
	orig := linkFile
	t.Cleanup(func() { linkFile = orig })
	n := 0
	linkFile = func(r *os.Root, from, to string) error {
		n++
		if fail != nil {
			if err := fail(n); err != nil {
				return err
			}
		}
		if err := orig(r, from, to); err != nil {
			return err
		}
		if act != nil {
			act(n)
		}
		return nil
	}
}

// replaceByRename puts a new file with body at path the way an atomic writer does: a
// temporary file in the same folder renamed over the name.
func replaceByRename(t *testing.T, path, body string) {
	t.Helper()
	tmp := filepath.Join(filepath.Dir(path), ".foreign-writer-tmp")
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatalf("write the foreign writer's temporary file: %v", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatalf("atomic replace of %s: %v", path, err)
	}
}

// idSetup runs the test through the link-based fallback with deterministic stage names:
// the round-3 identity tests address the staging folder by literal path and drive the link
// and drop seams (card t1733 round 4 moved the default path to no-replace renames).
func idSetup(t *testing.T) string {
	t.Helper()
	forceFallback(t)
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, idLower), "old body")
	return dir
}

// namesBoth reports whether err names the staged path and, apart from it, the skill.md name.
func namesBoth(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, idStaged) && strings.Contains(strings.ReplaceAll(msg, idStaged, ""), idLower)
}

// F11: the STAGED name is replaced by a foreign writer inside the private folder. The
// foreign file, the folder and the original skill.md all stay, and the error names both.
func TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedName(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, func(n int) {
		if n == 1 {
			replaceByRename(t, filepath.Join(dir, idStaged), "FOREIGN STAGED")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatalf("the rename reported success although the staged name was replaced; tree %v", treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "FOREIGN STAGED" {
		t.Fatalf("the foreign staged file = %q, want it left in place (tree %v)", got, treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "old body" {
		t.Fatalf("skill.md = %q, want the original left in place", got)
	}
	if !namesBoth(err) {
		t.Fatalf("error %q does not name both %s and %s", err, idLower, idStaged)
	}
}

// Both names replaced: nothing is proven ours, so nothing is removed.
func TestAC_UBR_17_RenameLeavesEverythingWhenBothNamesWereReplaced(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, func(n int) {
		if n == 1 {
			replaceByRename(t, filepath.Join(dir, idStaged), "FOREIGN STAGED")
			replaceByRename(t, filepath.Join(dir, idLower), "USER NEW EDIT")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatal("the rename reported success although both names were replaced")
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "FOREIGN STAGED" {
		t.Fatalf("the foreign staged file = %q, want it left in place", got)
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "USER NEW EDIT" {
		t.Fatalf("skill.md = %q, want the user's file left in place", got)
	}
	if !namesBoth(err) {
		t.Fatalf("error %q does not name both %s and %s", err, idLower, idStaged)
	}
}

// F12: skill.md vanishes after the link (or cannot be looked at). The staged name may be
// the only one left, so the content stays staged and the error says where.
func TestAC_UBR_17_RenameKeepsTheContentStagedWhenSkillMdIsGone(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, func(n int) {
		if n == 1 {
			if err := os.Remove(filepath.Join(dir, idLower)); err != nil {
				t.Fatalf("remove skill.md: %v", err)
			}
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !strings.Contains(err.Error(), idStaged) {
		t.Fatalf("error = %v; want one naming %s", err, idStaged)
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "old body" {
		t.Fatalf("the staged content = %q, want it kept (tree %v)", got, treeSnapshot(t, dir))
	}
}

// The staged name disappears and skill.md is still the original: nothing is removed.
func TestAC_UBR_17_RenameLeavesSkillMdWhenTheStagedNameIsGone(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, func(n int) {
		if n == 1 {
			if err := os.Remove(filepath.Join(dir, idStaged)); err != nil {
				t.Fatalf("remove the staged name: %v", err)
			}
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !namesBoth(err) {
		t.Fatalf("error = %v; want one naming both names", err)
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "old body" {
		t.Fatalf("skill.md = %q, want the original", got)
	}
}

// The staged name cannot be looked at (the folder lost its search permission). skill.md is
// still readable; nothing is removed and the error names both.
func TestAC_UBR_17_RenameLeavesEverythingWhenTheStagedNameIsUnreadable(t *testing.T) {
	requireUnprivileged(t)
	dir := idSetup(t)
	stage := filepath.Join(dir, "skill.md.tmp-rename")
	afterNthLink(t, func(n int) {
		if n == 1 {
			if err := os.Chmod(stage, 0o000); err != nil {
				t.Fatalf("chmod the stage folder: %v", err)
			}
		}
	}, nil)
	restoreMode(t, stage)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !namesBoth(err) {
		t.Fatalf("error = %v; want one naming both names", err)
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "old body" {
		t.Fatalf("skill.md = %q, want the original", got)
	}
	if err := os.Chmod(stage, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "old body" {
		t.Fatalf("the staged link = %q, want it kept", got)
	}
}

// The drop of skill.md is refused: skill.md keeps its name, our extra staged link and the
// folder are taken back.
func TestAC_UBR_17_RenameTakesBackItsLinkWhenTheDropOfSkillMdFails(t *testing.T) {
	dir := idSetup(t)
	orig := dropFile
	t.Cleanup(func() { dropFile = orig })
	dropFile = func(r *os.Root, name string) error {
		if name == idLower {
			return errors.New("injected refusal to drop skill.md")
		}
		return orig(r, name)
	}

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatal("the rename reported success although skill.md could not be dropped")
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "old body" {
		t.Fatalf("skill.md = %q, want the original", got)
	}
	if _, statErr := os.Lstat(filepath.Join(dir, "skill.md.tmp-rename")); !os.IsNotExist(statErr) {
		t.Fatalf("our staging folder was left behind (stat err %v); tree %v", statErr, treeSnapshot(t, dir))
	}
}

// The drop of skill.md and the undo of the staged link are both refused: the content keeps
// both names and the error reports it.
func TestAC_UBR_17_RenameReportsWhenTheUndoOfTheStagedLinkFails(t *testing.T) {
	dir := idSetup(t)
	orig := dropFile
	t.Cleanup(func() { dropFile = orig })
	dropFile = func(*os.Root, string) error { return errors.New("injected refusal to drop") }

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatal("the rename reported success although every drop was refused")
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "old body" {
		t.Fatalf("skill.md = %q, want the original", got)
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "old body" {
		t.Fatalf("the staged link = %q, want it kept", got)
	}
}

// The staged name is replaced after SKILL.md was linked, before our staged link is dropped:
// the foreign file survives, SKILL.md holds the content, and the error says so.
func TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedNameBeforeTheFinalDrop(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, func(n int) {
		if n == 2 {
			replaceByRename(t, filepath.Join(dir, idStaged), "FOREIGN STAGED")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatal("the rename reported success although the staged name was replaced before its drop")
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "FOREIGN STAGED" {
		t.Fatalf("the foreign staged file = %q, want it left in place (tree %v)", got, treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "old body" {
		t.Fatalf("SKILL.md = %q, want the renamed content", got)
	}
}

// The link to SKILL.md is refused, the restore link works, and the staged name is replaced
// before it is dropped: the foreign file survives and skill.md is back.
func TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedNameBeforeTheRestoreDrop(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, func(n int) {
		if n == 3 {
			replaceByRename(t, filepath.Join(dir, idStaged), "FOREIGN STAGED")
		}
	}, func(n int) error {
		if n == 2 {
			return errors.New("injected refusal of the SKILL.md link")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatal("the rename reported success although the SKILL.md link was refused")
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "FOREIGN STAGED" {
		t.Fatalf("the foreign staged file = %q, want it left in place (tree %v)", got, treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "old body" {
		t.Fatalf("skill.md = %q, want the restored original", got)
	}
}

// F2: the SKILL.md link and the restore link are both refused while skill.md is free. The
// step has no way left that cannot replace a name, so the content stays staged and the
// error names where; no fallback move puts it back.
func TestAC_UBR_17_RenameLeavesTheContentStagedWhenTheRestoreLinkIsRefused(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, nil, func(n int) error {
		if n >= 2 {
			return errors.New("injected link refusal")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !strings.Contains(err.Error(), idStaged) {
		t.Fatalf("error = %v; want one naming %s", err, idStaged)
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "old body" {
		t.Fatalf("the staged content = %q, want it kept", got)
	}
	if _, statErr := os.Lstat(filepath.Join(dir, idLower)); !os.IsNotExist(statErr) {
		t.Fatalf("skill.md was put back by a move that can replace a name (stat err %v); tree %v", statErr, treeSnapshot(t, dir))
	}
}

// F2 race: the user creates skill.md between the drop of the original and the restore. The
// restore link fails on the taken name; the user's file keeps its bytes and the content
// stays staged.
func TestAC_UBR_17_RenameKeepsAFileTheUserCreatedAtSkillMdDuringTheRestore(t *testing.T) {
	dir := idSetup(t)
	afterNthLink(t, nil, func(n int) error {
		if n == 2 {
			mustWriteFile(t, filepath.Join(dir, idLower), "USER NEW FILE")
			return errors.New("injected refusal of the SKILL.md link")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !strings.Contains(err.Error(), idStaged) {
		t.Fatalf("error = %v; want one naming %s", err, idStaged)
	}
	if got := readOrFail(t, filepath.Join(dir, idLower)); got != "USER NEW FILE" {
		t.Fatalf("skill.md = %q, want the user's file", got)
	}
	if got := readOrFail(t, filepath.Join(dir, idStaged)); got != "old body" {
		t.Fatalf("the staged content = %q, want it kept", got)
	}
}
