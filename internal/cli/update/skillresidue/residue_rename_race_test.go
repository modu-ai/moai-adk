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

// AC-UBR-17 (F25): a user file that appears at the temporary name just before the first
// move of the rename must keep its bytes. The step may finish under another temporary name
// or put the content back under its own name and report; it may not overwrite.
func TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", raceSkill)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")

	orig := moveFile
	t.Cleanup(func() { moveFile = orig })
	calls := 0
	moveFile = func(from, to string) error {
		calls++
		if calls == 1 {
			// The user's file lands at the destination name of the first move.
			if err := os.WriteFile(to, []byte("user data"), 0o644); err != nil {
				t.Fatalf("create the user file: %v", err)
			}
		}
		return orig(from, to)
	}

	_, err := Clean(root, Options{Shipped: raceShipped})

	if calls == 0 {
		t.Fatal("no move ran: the pass did not reach the rename")
	}
	if got := namesHolding(t, dir, "user data"); len(got) == 0 {
		t.Fatalf("the user file's bytes are gone; names now %v (Clean err: %v)", exactNames(t, dir), err)
	}
	if got := namesHolding(t, dir, "lower body"); len(got) == 0 {
		t.Fatalf("the skill file's content is gone; names now %v (Clean err: %v)", exactNames(t, dir), err)
	}
}
