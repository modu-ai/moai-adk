package skillresidue

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Card t1733, repair round 4 (sync-audit iteration 4, codex P1 on claimStage and dropOurs):
// the held-back rename moves the file with atomic no-replace renames, so no user-visible
// name is unlinked behind a check that a second actor could outrun. These tests cover the
// primitive, the two windows the audit reproduced, every branch of the rename pair, and the
// link-based fallback that stays for filesystems without a no-replace rename.

// requireNoReplace skips a test when this platform or filesystem has no no-replace rename;
// the fallback tests cover that case.
func requireNoReplace(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	mustWriteFile(t, filepath.Join(d, "a"), "x")
	if err := renameInDir(t, d, "a", "b"); errors.Is(err, errNoReplaceUnsupported) {
		t.Skip("no no-replace rename on this filesystem; the link fallback is tested instead")
	}
}

// renameInDir calls the platform primitive on two names inside dir, held open as one folder.
// Round 5 moved the primitive from absolute paths to open folders, so the primitive tests go
// through this wrapper and keep their assertions.
func renameInDir(t *testing.T, dir, from, to string) error {
	t.Helper()
	f, err := os.Open(dir)
	if err != nil {
		t.Fatalf("open %s: %v", dir, err)
	}
	defer func() { _ = f.Close() }()
	return noReplaceRenameAt(f, from, f, to)
}

// legacyStageNames makes the staging folder names deterministic (skill.md.tmp-rename,
// skill.md.tmp-rename-1, ...), which the tests of the link fallback address by literal path.
func legacyStageNames(t *testing.T) {
	t.Helper()
	orig := stageSuffix
	t.Cleanup(func() { stageSuffix = orig })
	stageSuffix = func() string { return "" }
}

// forceFallback makes the no-replace rename report "unsupported", so the rename runs the
// link-based path, and keeps the staging names deterministic.
func forceFallback(t *testing.T) {
	t.Helper()
	legacyStageNames(t)
	orig := renameNoReplace
	t.Cleanup(func() { renameNoReplace = orig })
	renameNoReplace = func(*os.File, string, *os.File, string) error { return errNoReplaceUnsupported }
}

// afterNthRename wraps the no-replace seam: fail(n), when it returns an error for the n-th
// call (1-based), replaces the real rename; act(n) runs after a real rename that succeeded.
// The returned counter holds the number of calls.
func afterNthRename(t *testing.T, act func(n int), fail func(n int) error) *int {
	t.Helper()
	orig := renameNoReplace
	t.Cleanup(func() { renameNoReplace = orig })
	calls := new(int)
	renameNoReplace = func(srcDir *os.File, src string, dstDir *os.File, dst string) error {
		*calls++
		if fail != nil {
			if err := fail(*calls); err != nil {
				return err
			}
		}
		if err := orig(srcDir, src, dstDir, dst); err != nil {
			return err
		}
		if act != nil {
			act(*calls)
		}
		return nil
	}
	return calls
}

// stageFolder returns the name of the one staging folder under dir, or fails.
func stageFolder(t *testing.T, dir string) string {
	t.Helper()
	var found []string
	for _, name := range exactNames(t, dir) {
		if strings.HasPrefix(name, "skill.md.tmp-rename") {
			found = append(found, name)
		}
	}
	if len(found) != 1 {
		t.Fatalf("want exactly one staging folder under %s, have %v (tree %v)", dir, found, treeSnapshot(t, dir))
	}
	return found[0]
}

func nrSetup(t *testing.T) string {
	t.Helper()
	requireNoReplace(t)
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")
	return dir
}

// --- the primitive, on the real filesystem -------------------------------------------

func TestNoReplaceRename_MovesWhenTheDestinationIsFree(t *testing.T) {
	requireNoReplace(t)
	d := t.TempDir()
	mustWriteFile(t, filepath.Join(d, "a"), "body")

	if err := renameInDir(t, d, "a", "b"); err != nil {
		t.Fatalf("rename onto a free name: %v", err)
	}
	if got := readOrFail(t, filepath.Join(d, "b")); got != "body" {
		t.Fatalf("b = %q, want the moved content", got)
	}
	if _, err := os.Lstat(filepath.Join(d, "a")); !os.IsNotExist(err) {
		t.Fatalf("a still exists after the move (stat err %v)", err)
	}
}

func TestNoReplaceRename_RefusesAnExistingDestination(t *testing.T) {
	requireNoReplace(t)
	d := t.TempDir()
	mustWriteFile(t, filepath.Join(d, "a"), "mine")
	mustWriteFile(t, filepath.Join(d, "b"), "theirs")

	err := renameInDir(t, d, "a", "b")

	if err == nil {
		t.Fatal("rename onto an existing name succeeded")
	}
	if errors.Is(err, errNoReplaceUnsupported) || !errors.Is(err, fs.ErrExist) {
		t.Fatalf("error %q must be a destination-exists error and not the unsupported sentinel", err)
	}
	if readOrFail(t, filepath.Join(d, "a")) != "mine" || readOrFail(t, filepath.Join(d, "b")) != "theirs" {
		t.Fatalf("a refused rename changed a file; tree %v", treeSnapshot(t, d))
	}
}

// --- window 1 (codex P1, claimStage): the stage folder is swapped after it was created ---

// The freshly created stage folder is replaced by a regular file before the first move. The
// move fails (a file is not a folder), the foreign file is not removed, and the error names
// the staged path.
func TestAC_UBR_17_RenameLeavesAFileThatReplacedTheStageFolderBeforeTheFirstMove(t *testing.T) {
	requireNoReplace(t)
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")
	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	var swapped string
	mkdirFile = func(r *os.Root, name string) error {
		if err := orig(r, name); err != nil {
			return err
		}
		if swapped == "" {
			swapped = name
			p := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.Remove(p); err != nil {
				t.Fatalf("remove the new stage folder: %v", err)
			}
			mustWriteFile(t, p, "FOREIGN FILE AT THE STAGE NAME")
		}
		return nil
	}

	err := renameSkillFile(openRootT(t, dir), ".")

	if swapped == "" {
		t.Fatal("the swap never ran: no stage folder was created")
	}
	if got := readOrFail(t, filepath.Join(dir, swapped)); got != "FOREIGN FILE AT THE STAGE NAME" {
		t.Fatalf("the foreign file at the stage name = %q, want it left in place (tree %v)", got, treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "old body" {
		t.Fatalf("skill.md = %q, want it untouched", got)
	}
	if err == nil || !strings.Contains(err.Error(), swapped) {
		t.Fatalf("error = %v; want one naming %s", err, swapped)
	}
}

// The stage folder is replaced by a foreign folder that holds its own file. The user's file
// lands in it, the rename completes, the foreign folder keeps its file because the final
// removal is not recursive, and the error names the folder it could not remove. This held on
// the link-based code too: it is a pin.
func TestAC_UBR_17_RenameKeepsTheContentsOfAFolderThatReplacedTheStageFolder(t *testing.T) {
	requireNoReplace(t)
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")
	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	var swapped string
	mkdirFile = func(r *os.Root, name string) error {
		if err := orig(r, name); err != nil {
			return err
		}
		if swapped == "" {
			swapped = name
			p := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.Remove(p); err != nil {
				t.Fatalf("remove the new stage folder: %v", err)
			}
			mustWriteFile(t, filepath.Join(p, "foreign.txt"), "FOREIGN CONTENT")
		}
		return nil
	}

	err := renameSkillFile(openRootT(t, dir), ".")

	if got := readOrFail(t, filepath.Join(dir, swapped, "foreign.txt")); got != "FOREIGN CONTENT" {
		t.Fatalf("the foreign folder's file = %q, want it kept (tree %v)", got, treeSnapshot(t, dir))
	}
	if !fileWithBodyUnder(t, dir, "old body") {
		t.Fatalf("the skill file's content is gone; tree %v", treeSnapshot(t, dir))
	}
	if err == nil || !strings.Contains(err.Error(), swapped) {
		t.Fatalf("error = %v; want one naming the folder it could not remove (%s)", err, swapped)
	}
}

// The staging folder name carries a random component, so a foreign actor has to guess it.
func TestAC_UBR_17_RenameStageFolderNameIsUnpredictable(t *testing.T) {
	requireNoReplace(t)
	shape := regexp.MustCompile(`^skill\.md\.tmp-rename-[0-9a-f]{12}$`)
	seen := map[string]bool{}
	for range 3 {
		dir := t.TempDir()
		mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")
		orig := mkdirFile
		var name string
		mkdirFile = func(r *os.Root, n string) error {
			name = n
			return orig(r, n)
		}
		err := renameSkillFile(openRootT(t, dir), ".")
		mkdirFile = orig
		if err != nil {
			t.Fatalf("rename: %v", err)
		}
		if !shape.MatchString(name) {
			t.Fatalf("stage folder name %q does not match %s", name, shape)
		}
		seen[name] = true
	}
	if len(seen) != 3 {
		t.Fatalf("three renames produced %d distinct stage names, want 3: %v", len(seen), seen)
	}
}

// --- window 2 (codex P1, dropOurs): no name is unlinked behind a check ------------------

// On a filesystem with a no-replace rename the rename pair removes no user-visible name, so
// there is no check-to-unlink window left. Any call of the drop seam fails the test.
func TestAC_UBR_17_RenameUnlinksNoNameWhenTheFilesystemCanRenameWithoutReplacing(t *testing.T) {
	dir := nrSetup(t)
	orig := dropFile
	t.Cleanup(func() { dropFile = orig })
	dropFile = func(r *os.Root, name string) error {
		t.Errorf("dropFile(%q): the no-replace path must not unlink a name behind a check", name)
		return orig(r, name)
	}

	if err := renameSkillFile(openRootT(t, dir), "."); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if names := exactNames(t, dir); len(names) != 1 || names[0] != "SKILL.md" {
		t.Fatalf("names after the rename = %v, want only SKILL.md", names)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "old body" {
		t.Fatalf("SKILL.md = %q, want the renamed content", got)
	}
}

// A SKILL.md the user creates between the two renames makes the second rename refuse. The
// user's file keeps its bytes, the content survives (back under skill.md, or staged where the
// filesystem folds case), and the error names SKILL.md.
func TestAC_UBR_17_RenameKeepsAFileThatAppearsAtSkillMdBetweenTheTwoRenames(t *testing.T) {
	dir := nrSetup(t)
	calls := afterNthRename(t, func(n int) {
		if n == 1 {
			mustWriteFile(t, filepath.Join(dir, "SKILL.md"), "USER SKILL")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if *calls < 2 {
		t.Fatalf("the rename made %d calls of the no-replace seam, want at least 2", *calls)
	}
	if err == nil || !strings.Contains(err.Error(), "SKILL.md") {
		t.Fatalf("error = %v; want one naming SKILL.md", err)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "USER SKILL" {
		t.Fatalf("SKILL.md = %q, want the user's file", got)
	}
	if !fileWithBodyUnder(t, dir, "old body") {
		t.Fatalf("the skill file's content is gone; tree %v", treeSnapshot(t, dir))
	}
}

// The user creates skill.md between the two renames (the original name is free once the first
// rename has run). Where the filesystem folds case this also blocks SKILL.md; where it does
// not, SKILL.md is free and the rename completes. Either way the user's file keeps its bytes
// and the content survives.
func TestAC_UBR_17_RenameKeepsAFileThatAppearsAtSkillMdLowerBetweenTheTwoRenames(t *testing.T) {
	dir := nrSetup(t)
	calls := afterNthRename(t, func(n int) {
		if n == 1 {
			mustWriteFile(t, filepath.Join(dir, "skill.md"), "USER LOWER")
		}
	}, nil)

	_ = renameSkillFile(openRootT(t, dir), ".")

	if *calls < 1 {
		t.Fatal("the rename never called the no-replace seam: the hook did not run")
	}
	if !fileWithBodyUnder(t, dir, "USER LOWER") {
		t.Fatalf("the user's file is gone; tree %v", treeSnapshot(t, dir))
	}
	if !fileWithBodyUnder(t, dir, "old body") {
		t.Fatalf("the skill file's content is gone; tree %v", treeSnapshot(t, dir))
	}
}

// The first rename fails for a reason other than a taken name: skill.md is untouched and the
// empty stage folder is taken back.
func TestAC_UBR_17_RenameLeavesSkillMdUntouchedWhenTheFirstRenameFails(t *testing.T) {
	dir := nrSetup(t)
	calls := afterNthRename(t, nil, func(n int) error {
		if n == 1 {
			return errors.New("injected failure of the first rename")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if *calls != 1 {
		t.Fatalf("no-replace seam calls = %d, want 1", *calls)
	}
	if err == nil || !strings.Contains(err.Error(), "injected failure") {
		t.Fatalf("error = %v; want the injected failure", err)
	}
	if names := exactNames(t, dir); len(names) != 1 || names[0] != "skill.md" {
		t.Fatalf("names = %v, want only skill.md (no stage folder left behind)", names)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "old body" {
		t.Fatalf("skill.md = %q, want the original", got)
	}
}

// The second rename is refused: the file is moved back to skill.md, the stage folder goes,
// and the error says why. (The rollback is what this test pins.)
func TestAC_UBR_17_RenamePutsTheFileBackWhenTheSecondRenameIsRefused(t *testing.T) {
	dir := nrSetup(t)
	calls := afterNthRename(t, nil, func(n int) error {
		if n == 2 {
			return errors.New("injected refusal of the SKILL.md rename")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if *calls != 3 {
		t.Fatalf("no-replace seam calls = %d, want 3 (move, refused move, put back)", *calls)
	}
	if err == nil || !strings.Contains(err.Error(), "injected refusal") {
		t.Fatalf("error = %v; want the injected refusal", err)
	}
	if names := exactNames(t, dir); len(names) != 1 || names[0] != "skill.md" {
		t.Fatalf("names = %v, want only skill.md (put back, no stage folder)", names)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "old body" {
		t.Fatalf("skill.md = %q, want the put-back content", got)
	}
}

// The second rename is refused and the user has created skill.md meanwhile: the put-back is
// refused too, the user's skill.md keeps its bytes, the content stays staged, and the error
// names where it is.
func TestAC_UBR_17_RenameLeavesTheContentStagedWhenThePutBackIsRefused(t *testing.T) {
	dir := nrSetup(t)
	afterNthRename(t, nil, func(n int) error {
		if n == 2 {
			mustWriteFile(t, filepath.Join(dir, "skill.md"), "USER NEW FILE")
			return errors.New("injected refusal of the SKILL.md rename")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !strings.Contains(err.Error(), "skill.md.tmp-rename") {
		t.Fatalf("error = %v; want one naming the staged path", err)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "USER NEW FILE" {
		t.Fatalf("skill.md = %q, want the user's file", got)
	}
	if got := readOrFail(t, filepath.Join(dir, stageFolder(t, dir), "skill.md")); got != "old body" {
		t.Fatalf("the staged content = %q, want it kept", got)
	}
}

// Both the second rename and the put-back are refused while skill.md is free: the content
// stays staged and no other move puts it back.
func TestAC_UBR_17_RenameLeavesTheContentStagedWhenBothLaterRenamesFail(t *testing.T) {
	dir := nrSetup(t)
	afterNthRename(t, nil, func(n int) error {
		if n >= 2 {
			return errors.New("injected rename refusal")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !strings.Contains(err.Error(), "skill.md.tmp-rename") {
		t.Fatalf("error = %v; want one naming the staged path", err)
	}
	if got := readOrFail(t, filepath.Join(dir, stageFolder(t, dir), "skill.md")); got != "old body" {
		t.Fatalf("the staged content = %q, want it kept", got)
	}
	if _, statErr := os.Lstat(filepath.Join(dir, "skill.md")); !os.IsNotExist(statErr) {
		t.Fatalf("skill.md exists (stat err %v); the content must stay staged", statErr)
	}
}

// The stage folder, empty after the second rename, is replaced by a regular file before it is
// removed: the removal cannot take a file, so the foreign file stays and the error names it.
func TestAC_UBR_17_RenameLeavesAFileThatReplacedTheStageFolderBeforeItsRemoval(t *testing.T) {
	dir := nrSetup(t)
	var stage string
	calls := afterNthRename(t, func(n int) {
		if n == 2 {
			stage = stageFolder(t, dir)
			p := filepath.Join(dir, stage)
			if err := os.Remove(p); err != nil {
				t.Fatalf("remove the empty stage folder: %v", err)
			}
			mustWriteFile(t, p, "FOREIGN FILE AT THE STAGE NAME")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if *calls != 2 {
		t.Fatalf("no-replace seam calls = %d, want 2", *calls)
	}
	if got := readOrFail(t, filepath.Join(dir, stage)); got != "FOREIGN FILE AT THE STAGE NAME" {
		t.Fatalf("the foreign file at the stage name = %q, want it kept (tree %v)", got, treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "old body" {
		t.Fatalf("SKILL.md = %q, want the renamed content", got)
	}
	if err == nil || !strings.Contains(err.Error(), stage) {
		t.Fatalf("error = %v; want one naming %s", err, stage)
	}
}

// The user's editor saves skill.md atomically after the stage was claimed and before the first
// rename. The step moves the file that is at skill.md now, so the user's new bytes become
// SKILL.md and nothing is lost.
func TestAC_UBR_17_RenameMovesTheUsersNewFileWhenSkillMdWasSavedBeforeTheFirstRename(t *testing.T) {
	dir := nrSetup(t)
	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	saved := false
	mkdirFile = func(r *os.Root, name string) error {
		if err := orig(r, name); err != nil {
			return err
		}
		if !saved {
			saved = true
			replaceByRename(t, filepath.Join(dir, "skill.md"), "USER NEW EDIT")
		}
		return nil
	}

	err := renameSkillFile(openRootT(t, dir), ".")

	if !saved {
		t.Fatal("the save never ran")
	}
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "USER NEW EDIT" {
		t.Fatalf("SKILL.md = %q, want the user's atomically saved bytes", got)
	}
	if names := exactNames(t, dir); len(names) != 1 {
		t.Fatalf("names = %v, want only SKILL.md", names)
	}
}

// --- the fallback -------------------------------------------------------------------

// The no-replace rename reports "unsupported" on its first use: nothing moved, and the
// link-based path finishes the rename. Only this path keeps a microsecond window.
func TestAC_UBR_17_RenameFallsBackToLinksWhenNoReplaceRenameIsUnsupported(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")
	calls := 0
	orig := renameNoReplace
	t.Cleanup(func() { renameNoReplace = orig })
	renameNoReplace = func(*os.File, string, *os.File, string) error {
		calls++
		return errNoReplaceUnsupported
	}

	if err := renameSkillFile(openRootT(t, dir), "."); err != nil {
		t.Fatalf("rename through the fallback: %v", err)
	}

	if calls != 1 {
		t.Fatalf("no-replace seam calls = %d, want 1 (one probe, then the fallback)", calls)
	}
	if names := exactNames(t, dir); len(names) != 1 || names[0] != "SKILL.md" {
		t.Fatalf("names = %v, want only SKILL.md", names)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "old body" {
		t.Fatalf("SKILL.md = %q, want the renamed content", got)
	}
}

// The fallback keeps the round-3 identity rule with the random stage folder name: a foreign
// file that replaced the staged name is not removed.
func TestAC_UBR_17_FallbackKeepsAForeignStagedFileUnderARandomStageName(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")
	orig := renameNoReplace
	t.Cleanup(func() { renameNoReplace = orig })
	renameNoReplace = func(*os.File, string, *os.File, string) error { return errNoReplaceUnsupported }
	afterNthLink(t, func(n int) {
		if n == 1 {
			matches, err := filepath.Glob(filepath.Join(dir, "skill.md.tmp-rename-*", "skill.md"))
			if err != nil || len(matches) != 1 {
				t.Fatalf("staged file not found: %v %v", matches, err)
			}
			replaceByRename(t, matches[0], "FOREIGN STAGED")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil {
		t.Fatal("the rename reported success although the staged name was replaced")
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "skill.md.tmp-rename-*", "skill.md"))
	if len(matches) != 1 || readOrFail(t, matches[0]) != "FOREIGN STAGED" {
		t.Fatalf("the foreign staged file was not left in place; tree %v", treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "old body" {
		t.Fatalf("skill.md = %q, want the original", got)
	}
}

// --- containment (AC-UBR-16 after the rename left the root handle) ---------------------

// The renames are path-based, so each is preceded by a root-bound look at its source: a skill
// folder swapped for a link to an outside folder after the stage was claimed is refused and
// the outside file is not moved.
func TestAC_UBR_16_RenameDoesNotFollowASkillFolderSwappedAfterTheStageWasClaimed(t *testing.T) {
	requireNoReplace(t)
	root := t.TempDir()
	outside := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", raceSkill)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "inside body")
	mustWriteFile(t, filepath.Join(outside, "skill.md"), "outside body")
	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	swapped := false
	mkdirFile = func(r *os.Root, name string) error {
		if err := orig(r, name); err != nil {
			return err
		}
		if !swapped {
			swapped = true
			if err := os.Rename(dir, dir+".gone"); err != nil {
				t.Fatalf("move the skill folder away: %v", err)
			}
			if err := os.Symlink(outside, dir); err != nil {
				t.Skip(err)
			}
			mustMkdirAll(t, filepath.Join(outside, filepath.Base(filepath.FromSlash(name))))
		}
		return nil
	}

	_ = renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	if !swapped {
		t.Fatal("the swap never ran")
	}
	if got := readOrFail(t, filepath.Join(outside, "skill.md")); got != "outside body" {
		t.Fatalf("outside skill.md = %q", got)
	}
	for _, name := range exactNames(t, outside) {
		if name == "SKILL.md" || name == "skill.md" {
			continue
		}
		if entries, _ := os.ReadDir(filepath.Join(outside, name)); len(entries) != 0 {
			t.Fatalf("the outside folder received %s/%v through the swapped component", name, entries)
		}
	}
}
