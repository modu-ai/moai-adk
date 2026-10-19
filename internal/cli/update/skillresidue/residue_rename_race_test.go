package skillresidue

import (
	"os"
	"path/filepath"
	"testing"
)

// SPEC-UPDATE-BACKUP-RETENTION-001 M3 (REQ-UBR-017, findings F24 and F25): the held-back
// lowercase skill.md rename is bound to the project root and never overwrites a name.

const raceSkill = "moai-ref-owasp-checklist"

func raceShipped(rel string) bool { return rel == ".claude/skills/"+raceSkill+"/SKILL.md" }

// namesHolding returns the stored names under dir whose content equals body.
func namesHolding(t *testing.T, dir, body string) []string {
	t.Helper()
	var out []string
	for _, name := range exactNames(t, dir) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil && string(data) == body {
			out = append(out, name)
		}
	}
	return out
}

// AC-UBR-16 (F24): a skill folder that the Shipped callback replaces with a symbolic link
// to an outside directory, just before the rename, must not have the outside skill.md
// renamed. The callback runs after the pass's own checks and before the first move, which
// is where a swap would land.
func TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename(t *testing.T) {
	passes := map[string]func(root string, opts Options) error{
		"Clean": func(root string, opts Options) error {
			_, err := Clean(root, opts)
			return err
		},
		"ApplyDeferredRenames": func(root string, opts Options) error {
			_, err := ApplyDeferredRenames(root, opts)
			return err
		},
	}
	for name, run := range passes {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			outside := t.TempDir()
			dir := filepath.Join(root, ".claude", "skills", raceSkill)
			mustWriteFile(t, filepath.Join(dir, "skill.md"), "inside body")
			mustWriteFile(t, filepath.Join(outside, "skill.md"), "outside body")

			swapped := false
			shipped := func(rel string) bool {
				if !swapped && rel == ".claude/skills/"+raceSkill+"/SKILL.md" {
					swapped = true
					if err := os.RemoveAll(dir); err != nil {
						t.Fatalf("remove the skill folder: %v", err)
					}
					if err := os.Symlink(outside, dir); err != nil {
						t.Skip(err)
					}
				}
				return raceShipped(rel)
			}

			// The pass may report the refusal as an error; what matters is the outside file.
			_ = run(root, Options{Shipped: shipped})

			if !swapped {
				t.Fatal("the swap never ran: the pass did not reach the rename")
			}
			names := exactNames(t, outside)
			if !containsName(names, "skill.md") || containsName(names, "SKILL.md") || len(names) != 1 {
				t.Fatalf("the outside folder was changed through the swapped component: %v", names)
			}
			if got := readOrFail(t, filepath.Join(outside, "skill.md")); got != "outside body" {
				t.Fatalf("outside skill.md = %q", got)
			}
		})
	}
}

// AC-UBR-17 (F25): a user file that appears at the temporary name just before the rename's
// first step must keep its bytes. The step may finish under another temporary name or put
// the content back under its own name and report; it may not overwrite. (On the starting
// tree the same hook sat on the first move, whose Rename replaced the user file; the
// rename now claims its temporary name with an exclusive Mkdir, so the hook moved to it.)
func TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", raceSkill)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")

	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	calls := 0
	mkdirFile = func(r *os.Root, name string) error {
		calls++
		if calls == 1 {
			// The user's file lands at the temporary name the rename is about to take.
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte("user data"), 0o644); err != nil {
				t.Fatalf("create the user file: %v", err)
			}
		}
		return orig(r, name)
	}

	_, err := Clean(root, Options{Shipped: raceShipped})

	if calls == 0 {
		t.Fatal("no temporary name was claimed: the pass did not reach the rename")
	}
	if got := namesHolding(t, dir, "user data"); len(got) == 0 {
		t.Fatalf("the user file's bytes are gone; names now %v (Clean err: %v)", exactNames(t, dir), err)
	}
	if got := namesHolding(t, dir, "lower body"); len(got) == 0 {
		t.Fatalf("the skill file's content is gone; names now %v (Clean err: %v)", exactNames(t, dir), err)
	}
}

// AC-UBR-17 (F25, inner name; audit finding F1 of card t1733): a file that appears INSIDE
// the freshly claimed staging folder, at the name the first move is about to take, keeps
// its bytes. The step reports the refusal and the skill file stays under its own name.
// The claim of the folder (Mkdir) does not protect this inner name; only a move that
// cannot replace does.
func TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsInsideTheStagingFolder(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", raceSkill)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")

	orig := mkdirFile
	t.Cleanup(func() { mkdirFile = orig })
	var stageFile string
	mkdirFile = func(r *os.Root, name string) error {
		if err := orig(r, name); err != nil {
			return err
		}
		// The stage folder is ours now; the user's file lands inside it before the move.
		if stageFile == "" {
			stageFile = filepath.Join(root, filepath.FromSlash(name), "skill.md")
			if err := os.WriteFile(stageFile, []byte("user data"), 0o644); err != nil {
				t.Fatalf("create the user file inside the stage: %v", err)
			}
		}
		return nil
	}

	_, err := Clean(root, Options{Shipped: raceShipped})

	if stageFile == "" {
		t.Fatal("no staging folder was claimed: the pass did not reach the rename")
	}
	if !fileWithBodyUnder(t, dir, "user data") {
		t.Fatalf("the user's file inside the staging folder was overwritten; tree now %v (Clean err: %v)", treeSnapshot(t, dir), err)
	}
	if err == nil {
		t.Fatal("the rename reported success although its first move was refused")
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "lower body" {
		t.Fatalf("skill.md = %q, want it untouched under its own name", got)
	}
}

// AC-UBR-17 (F25, destination): a SKILL.md that appears after the check, before the final
// step, keeps its bytes, and the step reports the refusal.
func TestAC_UBR_17_RenameDoesNotOverwriteASkillFileThatAppearsAtTheDestination(t *testing.T) {
	forceFallback(t) // drives the link seam; the no-replace path has its own test
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", raceSkill)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")

	orig := linkFile
	t.Cleanup(func() { linkFile = orig })
	linkFile = func(r *os.Root, from, to string) error {
		if filepath.Base(to) == "SKILL.md" {
			// The staged link and the drop of skill.md are done; the user now creates
			// SKILL.md just before the final link.
			if werr := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("user data"), 0o644); werr != nil {
				t.Fatalf("create SKILL.md: %v", werr)
			}
		}
		return orig(r, from, to)
	}

	_, err := Clean(root, Options{Shipped: raceShipped})

	if err == nil {
		t.Fatal("the rename reported success although SKILL.md appeared at its destination")
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "user data" {
		t.Fatalf("SKILL.md = %q, want the user's bytes", got)
	}
	// Where the filesystem keeps skill.md and SKILL.md apart, the lowercase file is back
	// under its own name; where it folds case the name is taken by the user's file and the
	// content stays in the staging folder. Either way no byte is lost.
	if !fileWithBodyUnder(t, dir, "lower body") {
		t.Fatalf("the skill file's content is gone; tree now %v", treeSnapshot(t, dir))
	}
}

// fileWithBodyUnder reports whether any regular file under dir holds body.
func fileWithBodyUnder(t *testing.T, dir, body string) bool {
	t.Helper()
	found := false
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if data, readErr := os.ReadFile(p); readErr == nil && string(data) == body {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return found
}
