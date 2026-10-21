package skillresidue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Card t1733, repair round 5 (sync-audit iteration 5): the held-back rename must not let a
// swapped path redirect a move out of the project (codex P1, F15), must not report success
// when the staged file was replaced between the two renames (codex P2, F16), and its
// root-bound look at the source must be pinned by a test that fails without it (F17).

// handleSetup builds a project root holding one skill folder with a lowercase skill.md and an
// empty folder outside the project, and returns (project root, skill folder, outside folder).
func handleSetup(t *testing.T) (root, dir, outside string) {
	t.Helper()
	requireNoReplace(t)
	root = t.TempDir()
	outside = t.TempDir()
	dir = filepath.Join(root, ".claude", "skills", raceSkill)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "inside body")
	return root, dir, outside
}

// requireOutsideEmpty fails when the folder outside the project received anything.
func requireOutsideEmpty(t *testing.T, outside string) {
	t.Helper()
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatalf("the folder outside the project received %d entr(ies): %v", len(entries), treeSnapshot(t, outside))
	}
}

// codex P1 / F15: the freshly made stage folder is swapped for a link to a folder outside the
// project before the first move. The user's file must not leave the project, the outside
// folder must stay empty, and the step reports an error.
func TestAC_UBR_16_RenameRefusesAStageFolderSwappedForALinkOutsideTheProject(t *testing.T) {
	root, dir, outside := handleSetup(t)
	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	swapped := false
	mkdirFile = func(r *os.Root, name string) error {
		if err := orig(r, name); err != nil {
			return err
		}
		if !swapped {
			swapped = true
			p := filepath.Join(root, filepath.FromSlash(name))
			if err := os.Remove(p); err != nil {
				t.Fatalf("remove the new stage folder: %v", err)
			}
			if err := os.Symlink(outside, p); err != nil {
				t.Skip(err)
			}
		}
		return nil
	}

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	if !swapped {
		t.Fatal("the swap never ran")
	}
	requireOutsideEmpty(t, outside)
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "inside body" {
		t.Fatalf("skill.md = %q, want it untouched inside the project (tree %v)", got, treeSnapshot(t, dir))
	}
	if err == nil {
		t.Fatal("the rename reported success although its stage folder was swapped for an outside link")
	}
}

// F15, the descriptor-bound control: the stage path is swapped for a link to an outside
// folder AFTER the step has opened the stage folder and before the first rename. The rename
// is relative to the opened folders, so it still lands in the original stage folder (moved
// aside, inside the project); the outside folder stays empty and the user's bytes stay in the
// project. A path-based rename would carry the file out.
func TestAC_UBR_16_RenameIsBoundToTheOpenedStageFolderNotToItsPath(t *testing.T) {
	root, dir, outside := handleSetup(t)
	afterNthRename(t, nil, func(n int) error {
		if n == 1 {
			stage := filepath.Join(dir, stageFolder(t, dir))
			if err := os.Rename(stage, stage+".held"); err != nil {
				t.Fatalf("move the stage folder aside: %v", err)
			}
			if err := os.Symlink(outside, stage); err != nil {
				t.Skip(err)
			}
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	requireOutsideEmpty(t, outside)
	if !fileWithBodyUnder(t, root, "inside body") {
		t.Fatalf("the skill file's content left the project; tree %v", treeSnapshot(t, root))
	}
	if err == nil {
		t.Fatal("the rename reported success although the stage path no longer names the folder it moved into")
	}
}

// F17: a skill folder swapped for a link to an outside folder AFTER the first rename. The
// root-bound look at the second rename's source refuses; without it the rename would run
// against the held folder and report success while the project path names something else.
func TestAC_UBR_16_RenameRefusesWhenTheSkillFolderIsSwappedBetweenTheTwoRenames(t *testing.T) {
	root, dir, outside := handleSetup(t)
	afterNthRename(t, func(n int) {
		if n == 1 {
			if err := os.Rename(dir, dir+".gone"); err != nil {
				t.Fatalf("move the skill folder away: %v", err)
			}
			if err := os.Symlink(outside, dir); err != nil {
				t.Skip(err)
			}
		}
	}, nil)

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	requireOutsideEmpty(t, outside)
	if err == nil {
		t.Fatal("the rename reported success although the skill folder path was swapped for an outside link")
	}
	if !fileWithBodyUnder(t, root, "inside body") {
		t.Fatalf("the skill file's content left the project; tree %v", treeSnapshot(t, root))
	}
}

// codex P2 / F16: the staged file is replaced by a foreign file between the two renames. The
// second rename must not move it: the step reports an error naming the staged path, leaves
// the foreign file where it is and does not publish SKILL.md.
func TestAC_UBR_17_RenameRefusesAStagedFileThatWasReplacedBetweenTheRenames(t *testing.T) {
	dir := nrSetup(t)
	var staged string
	calls := afterNthRename(t, func(n int) {
		if n == 1 {
			staged = filepath.Join(dir, stageFolder(t, dir), "skill.md")
			replaceByRename(t, staged, "FOREIGN STAGED")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, dir), ".")

	if err == nil || !strings.Contains(err.Error(), "skill.md.tmp-rename") {
		t.Fatalf("error = %v; want one naming the staged path", err)
	}
	if *calls != 1 {
		t.Fatalf("no-replace seam calls = %d, want 1 (the second rename must not run)", *calls)
	}
	if got := readOrFail(t, staged); got != "FOREIGN STAGED" {
		t.Fatalf("the foreign staged file = %q, want it left where it is", got)
	}
	for _, name := range exactNames(t, dir) {
		if name == "SKILL.md" {
			t.Fatalf("SKILL.md was published from a file that is not the one staged; tree %v", treeSnapshot(t, dir))
		}
	}
}

// F16, the window between the identity check and the rename: the staged name is replaced right
// before the second rename, so the foreign file is moved. The step cannot prevent that, but it
// must not report success: the file now at SKILL.md is not the one staged.
func TestAC_UBR_17_RenameReportsAFileThatReplacedTheStagedNameDuringTheSecondRename(t *testing.T) {
	dir := nrSetup(t)
	calls := afterNthRename(t, nil, func(n int) error {
		if n == 2 {
			replaceByRename(t, filepath.Join(dir, stageFolder(t, dir), "skill.md"), "FOREIGN LATE")
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, dir), ".")

	if *calls != 2 {
		t.Fatalf("no-replace seam calls = %d, want 2", *calls)
	}
	if err == nil || !strings.Contains(err.Error(), "SKILL.md") {
		t.Fatalf("error = %v; want one naming SKILL.md", err)
	}
	if !fileWithBodyUnder(t, dir, "FOREIGN LATE") {
		t.Fatalf("the foreign file was destroyed; tree %v", treeSnapshot(t, dir))
	}
}
