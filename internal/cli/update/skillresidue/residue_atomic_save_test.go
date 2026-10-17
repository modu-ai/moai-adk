package skillresidue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-UBR-17 (source name; audit finding F7 of card t1733): the held-back rename links
// skill.md into the staging folder and then drops the original name. A user's editor that
// atomic-saves (writes a temporary file, renames it over skill.md) between those two steps
// gives skill.md a NEW inode, and dropping the name would delete the user's new file. The
// rename must notice the replacement, leave the user's file alone, report an error, and
// take back its own staging names.
func TestAC_UBR_17_RenameDoesNotDropASkillFileThatWasAtomicallyReplaced(t *testing.T) {
	forceFallback(t) // drives the link seam; the default path is no-replace renames
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "old body")

	orig := linkFile
	t.Cleanup(func() { linkFile = orig })
	saved := false
	linkFile = func(r *os.Root, from, to string) error {
		if err := orig(r, from, to); err != nil {
			return err
		}
		if !saved && filepath.Base(to) == "skill.md" {
			// The first link, into the staging folder, has succeeded. The user's editor now
			// saves atomically: a temporary file is renamed over skill.md.
			saved = true
			tmp := filepath.Join(dir, ".skill.md.editor-tmp")
			if werr := os.WriteFile(tmp, []byte("USER NEW EDIT"), 0o644); werr != nil {
				t.Fatalf("write the editor's temporary file: %v", werr)
			}
			if rerr := os.Rename(tmp, filepath.Join(dir, "skill.md")); rerr != nil {
				t.Fatalf("atomic save over skill.md: %v", rerr)
			}
		}
		return nil
	}

	err := renameSkillFile(openRootT(t, dir), ".")

	if !saved {
		t.Fatal("the atomic save never ran: the rename did not reach its first link")
	}
	if err == nil {
		t.Fatalf("the rename reported success although skill.md was replaced mid-step; tree now %v", treeSnapshot(t, dir))
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "USER NEW EDIT" {
		t.Fatalf("skill.md = %q, want the user's atomically saved bytes (tree %v)", got, treeSnapshot(t, dir))
	}
	for _, name := range exactNames(t, dir) {
		if strings.HasPrefix(name, "skill.md.tmp-rename") {
			t.Fatalf("the staging folder %q was left behind; tree now %v", name, treeSnapshot(t, dir))
		}
		if name == "SKILL.md" {
			t.Fatalf("SKILL.md was created from the replaced file; tree now %v", treeSnapshot(t, dir))
		}
	}
}
