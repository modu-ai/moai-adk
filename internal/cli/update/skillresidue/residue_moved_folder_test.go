package skillresidue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Card t1733, repair round 6 (sync-audit iteration 6): the renames are relative to the folders
// the step holds open, so a folder that is renamed away after it was opened still receives the
// move. These tests drive that through the same seams as round 5 and require that
//
//   - a held stage folder renamed out of the project (codex P1, F20) gets the user's skill.md
//     back, through the held descriptors, and the error never names a path with nothing at it
//     (F22);
//   - a skill folder renamed away after the first rename, with a decoy at its old path (codex
//     P2, F21), is reported as an error that says where the content is, and nothing is removed.

const contentClaim = "content is at "

// readIfThere returns the file's text and whether it could be read.
func readIfThere(path string) (string, bool) {
	data, err := os.ReadFile(path)
	return string(data), err == nil
}

// moveStageOut renames the one staging folder under dir to outside/moved-stage, the way a
// same-user process would after the step opened it, and returns the staging folder's name.
func moveStageOut(t *testing.T, dir, outside string) string {
	t.Helper()
	stage := stageFolder(t, dir)
	if err := os.Rename(filepath.Join(dir, stage), filepath.Join(outside, "moved-stage")); err != nil {
		t.Fatalf("move the stage folder out of the project: %v", err)
	}
	return stage
}

// requireBackInPlace fails unless skill.md holds the user's bytes at the requested path, the
// folder outside the project holds none of them, and the error exists without claiming the
// content sits at a path (it was put back, so there is nothing to point at).
func requireBackInPlace(t *testing.T, dir, outside string, err error) {
	t.Helper()
	if got, ok := readIfThere(filepath.Join(dir, "skill.md")); !ok || got != "inside body" {
		t.Fatalf("skill.md = %q (present %v), want the user's bytes back at the requested path (project %v, outside %v)",
			got, ok, treeSnapshot(t, filepath.Dir(dir)), treeSnapshot(t, outside))
	}
	if fileWithBodyUnder(t, outside, "inside body") {
		t.Fatalf("the user's bytes are in the folder outside the project: %v", treeSnapshot(t, outside))
	}
	if err == nil {
		t.Fatal("the rename reported success although the stage folder was moved away and the step did not finish")
	}
	if strings.Contains(err.Error(), contentClaim) {
		t.Fatalf("error = %v; it claims the content sits at a path although it was put back", err)
	}
}

// codex P1 / F20: the stage folder is renamed out of the project after it was opened and before
// the first rename. The first rename lands in the moved folder; the put-back must reach the
// skill folder through the held descriptors, because the path look at the old stage path can
// only fail.
func TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject(t *testing.T) {
	for _, tc := range []struct {
		name         string
		newAtOldPath bool
	}{
		{"nothing at the old path", false},
		{"a new folder at the old path", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, dir, outside := handleSetup(t)
			moved := false
			afterNthRename(t, nil, func(n int) error {
				if n == 1 {
					moved = true
					stage := moveStageOut(t, dir, outside)
					if tc.newAtOldPath {
						if err := os.Mkdir(filepath.Join(dir, stage), 0o755); err != nil {
							t.Fatalf("create a new folder at the old stage path: %v", err)
						}
					}
				}
				return nil
			})

			err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

			if !moved {
				t.Fatal("the stage folder was never moved")
			}
			requireBackInPlace(t, dir, outside, err)
		})
	}
}

// F20, the same move one step later: the stage folder leaves the project after the first rename
// succeeded. The identity check runs on the held folder and passes, the look before the second
// rename fails, and the put-back has to be made through the held descriptors.
func TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutBetweenTheRenames(t *testing.T) {
	root, dir, outside := handleSetup(t)
	afterNthRename(t, func(n int) {
		if n == 1 {
			moveStageOut(t, dir, outside)
		}
	}, nil)

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	requireBackInPlace(t, dir, outside, err)
}

// F20, S7: the skill folder itself is renamed (inside the project) before the first rename. The
// staging folder travels with it. The put-back goes into the held skill folder, so the content
// is restored as skill.md inside the moved folder, and no staging path is named as its place.
func TestAC_UBR_16_RenamePutsTheFileBackIntoASkillFolderThatWasMovedAside(t *testing.T) {
	root, dir, _ := handleSetup(t)
	afterNthRename(t, nil, func(n int) error {
		if n == 1 {
			if err := os.Rename(dir, dir+".moved"); err != nil {
				t.Fatalf("move the skill folder aside: %v", err)
			}
		}
		return nil
	})

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	if got, ok := readIfThere(filepath.Join(dir+".moved", "skill.md")); !ok || got != "inside body" {
		t.Fatalf("moved skill folder's skill.md = %q (present %v), want the user's bytes put back (tree %v)", got, ok, treeSnapshot(t, filepath.Dir(dir)))
	}
	if err == nil {
		t.Fatal("the rename reported success although the skill folder was moved before the first rename")
	}
	if strings.Contains(err.Error(), contentClaim) {
		t.Fatalf("error = %v; it claims the content sits at a path although it was put back", err)
	}
}

// codex P2 / F21: after the first rename the skill folder is renamed away and a decoy staging
// folder (with a skill.md of its own) takes the old path. The look at the old path passes and
// the second rename runs in the held, moved folder, so SKILL.md appears there and nowhere at the
// requested path. The step must not report success; it removes nothing and says where the
// content is.
func TestAC_UBR_17_RenameDoesNotReportSuccessWhenTheSkillFolderWasMovedAndADecoyTookItsPath(t *testing.T) {
	root, dir, _ := handleSetup(t)
	var stage string
	afterNthRename(t, func(n int) {
		if n == 1 {
			stage = stageFolder(t, dir)
			if err := os.Rename(dir, dir+".moved"); err != nil {
				t.Fatalf("move the skill folder aside: %v", err)
			}
			mustWriteFile(t, filepath.Join(dir, stage, "skill.md"), "decoy body")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	if err == nil {
		t.Fatalf("the rename reported success although %s has no SKILL.md (tree %v)",
			dir, treeSnapshot(t, filepath.Dir(dir)))
	}
	if !strings.Contains(err.Error(), "moved") {
		t.Fatalf("error = %v; want one that says the skill folder was moved", err)
	}
	if got := readOrFail(t, filepath.Join(dir+".moved", "SKILL.md")); got != "inside body" {
		t.Fatalf("moved skill folder's SKILL.md = %q, want the renamed content kept", got)
	}
	if got := readOrFail(t, filepath.Join(dir, stage, "skill.md")); got != "decoy body" {
		t.Fatalf("the decoy = %q, want it left alone", got)
	}
	if _, statErr := os.Lstat(filepath.Join(dir, "SKILL.md")); !os.IsNotExist(statErr) {
		t.Fatalf("a SKILL.md exists at the requested path (stat err %v)", statErr)
	}
	found := false
	for _, name := range exactNames(t, dir+".moved") {
		found = found || name == stage
	}
	if !found {
		t.Fatalf("the staging folder was removed from the moved skill folder; tree %v", treeSnapshot(t, filepath.Dir(dir)))
	}
}

// F22: the stage folder was moved out between the renames and a file of the user's took
// skill.md, so the put-back is refused too. The content stays in the moved folder; the error
// must say so instead of pointing at the old staging path, where nothing is.
func TestAC_UBR_17_RenameNamesTheMovedFolderWhenThePutBackIsRefused(t *testing.T) {
	root, dir, outside := handleSetup(t)
	afterNthRename(t, func(n int) {
		if n == 1 {
			moveStageOut(t, dir, outside)
			mustWriteFile(t, filepath.Join(dir, "skill.md"), "USER NEW FILE")
		}
	}, nil)

	err := renameSkillFile(openRootT(t, root), ".claude/skills/"+raceSkill)

	if err == nil {
		t.Fatal("the rename reported success although it could not finish")
	}
	if strings.Contains(err.Error(), contentClaim) {
		t.Fatalf("error = %v; it points at a staging path that holds nothing", err)
	}
	if !strings.Contains(err.Error(), "was moved") {
		t.Fatalf("error = %v; want one that says the staging folder was moved", err)
	}
	if got := readOrFail(t, filepath.Join(outside, "moved-stage", "skill.md")); got != "inside body" {
		t.Fatalf("the content in the moved folder = %q, want it kept", got)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "USER NEW FILE" {
		t.Fatalf("skill.md = %q, want the user's file kept", got)
	}
}
