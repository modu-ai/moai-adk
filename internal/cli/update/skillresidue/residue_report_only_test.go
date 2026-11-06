package skillresidue

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// SPEC-UPDATE-BACKUP-RETENTION-001 M3, AC-UBR-04, AC-UBR-05 and AC-UBR-15 (REQ-UBR-015,
// REQ-UBR-016). The retired-skill and link handling reports and changes nothing: the
// operator ruled (Q3, 2026-10-16) that archive and link removal are not reintroduced.
// These are regression guards. They are green on the tree they were written against, so
// their mutant is the defect they keep out: a cleanup that moves or removes what it reports.

// treeSnapshot lists every entry under dir as "<relative path> <kind>", where a symbolic
// link carries its target text. It never follows a link, so a link into another project
// is recorded, not walked.
func treeSnapshot(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(dir, p)
		if relErr != nil {
			return relErr
		}
		kind := "file"
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			target, linkErr := os.Readlink(p)
			if linkErr != nil {
				return linkErr
			}
			kind = "link->" + target
		case d.IsDir():
			kind = "dir"
		}
		out = append(out, filepath.ToSlash(rel)+" "+kind)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	sort.Strings(out)
	return out
}

func sameSnapshot(a, b []string) bool {
	return strings.Join(a, "\n") == strings.Join(b, "\n")
}

// onlyReports fails the test when an action is anything but a report row.
func onlyReports(t *testing.T, actions []Action) {
	t.Helper()
	for _, a := range actions {
		switch a.Kind {
		case "reported-retired-skill", "reported-retired-link":
		default:
			t.Errorf("action %+v is not a report row; the retired-skill pass reports and changes nothing", a)
		}
	}
}

func shipsSuccessor(rel string) bool {
	return rel == ".claude/skills/moai-factory-foreman/SKILL.md"
}

// AC-UBR-04 (P4a): project A's .agents/skills is a symbolic link to project B's
// .agents/skills, and B holds a live link of the retired name. The cleanup on A must
// leave B untouched and A's retired folder in place.
func TestAC_UBR_04_ContentBehindSymbolicAgentsSkillsIsUntouched(t *testing.T) {
	base := t.TempDir()
	a, b := filepath.Join(base, "A"), filepath.Join(base, "B")
	for _, p := range []string{a, b} {
		mustWriteFile(t, filepath.Join(p, ".claude", "skills", "moai-kanban-foreman", "SKILL.md"), "retired\n")
	}
	mustMkdirAll(t, filepath.Join(b, ".agents", "skills"))
	bLink := filepath.Join(b, ".agents", "skills", "moai-kanban-foreman")
	if err := os.Symlink("../../.claude/skills/moai-kanban-foreman", bLink); err != nil {
		t.Skip(err)
	}
	mustMkdirAll(t, filepath.Join(a, ".agents"))
	if err := os.Symlink(filepath.Join(b, ".agents", "skills"), filepath.Join(a, ".agents", "skills")); err != nil {
		t.Skip(err)
	}
	beforeA, beforeB := treeSnapshot(t, a), treeSnapshot(t, b)

	actions, err := Clean(a, Options{Shipped: shipsSuccessor})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}

	onlyReports(t, actions)
	if !hasAction(actions, "reported-retired-skill", ".claude/skills/moai-kanban-foreman") {
		t.Errorf("the retired folder was not reported: %+v", actions)
	}
	if _, err := os.Lstat(bLink); err != nil {
		t.Errorf("project B's link did not survive: %v", err)
	}
	if got := treeSnapshot(t, b); !sameSnapshot(beforeB, got) {
		t.Errorf("project B changed:\nbefore %q\nafter  %q", beforeB, got)
	}
	if got := treeSnapshot(t, a); !sameSnapshot(beforeA, got) {
		t.Errorf("project A changed:\nbefore %q\nafter  %q", beforeA, got)
	}
	if _, err := os.Lstat(filepath.Join(a, ".claude", "skills", "moai-kanban-foreman", "SKILL.md")); err != nil {
		t.Errorf("project A's retired folder did not stay in .claude/skills: %v", err)
	}
}

// AC-UBR-05 (P4b): project A's .moai/archive is a symbolic link to a directory outside
// the project. Nothing is written outside, and the retired folder stays in
// .claude/skills.
func TestAC_UBR_05_NothingIsWrittenOutsideTheProjectThroughArchiveLink(t *testing.T) {
	base := t.TempDir()
	a, outside := filepath.Join(base, "A"), filepath.Join(base, "outside")
	mustWriteFile(t, filepath.Join(a, ".claude", "skills", "moai-kanban-foreman", "SKILL.md"), "retired\n")
	mustMkdirAll(t, filepath.Join(a, ".moai"))
	mustMkdirAll(t, outside)
	if err := os.Symlink(outside, filepath.Join(a, ".moai", "archive")); err != nil {
		t.Skip(err)
	}
	beforeA := treeSnapshot(t, a)

	actions, err := Clean(a, Options{Shipped: shipsSuccessor})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}

	onlyReports(t, actions)
	if got := treeSnapshot(t, outside); len(got) != 1 {
		t.Errorf("content appeared outside the project through .moai/archive: %q", got)
	}
	if got := treeSnapshot(t, a); !sameSnapshot(beforeA, got) {
		t.Errorf("project A changed:\nbefore %q\nafter  %q", beforeA, got)
	}
}

// AC-UBR-15: a retired folder holding only SKILL.md, its shipped successor, a regular
// .moai directory and a link under .agents/skills to the retired folder. The cleanup runs
// twice; both runs report the folder and the link, and neither moves nor removes
// anything, nor creates .moai/archive.
func TestAC_UBR_15_RetiredFoldersAndLinksAreReportedNeverMovedOrRemoved(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".claude", "skills", "moai-kanban-foreman", "SKILL.md"), "retired\n")
	mustMkdirAll(t, filepath.Join(root, ".moai"))
	mustMkdirAll(t, filepath.Join(root, ".agents", "skills"))
	link := filepath.Join(root, ".agents", "skills", "moai-kanban-foreman")
	if err := os.Symlink("../../.claude/skills/moai-kanban-foreman", link); err != nil {
		t.Skip(err)
	}
	before := treeSnapshot(t, root)

	for run := 1; run <= 2; run++ {
		actions, err := Clean(root, Options{Shipped: shipsSuccessor})
		if err != nil {
			t.Fatalf("run %d: Clean: %v", run, err)
		}
		onlyReports(t, actions)
		if !hasAction(actions, "reported-retired-skill", ".claude/skills/moai-kanban-foreman") {
			t.Errorf("run %d: the retired folder was not reported: %+v", run, actions)
		}
		if !hasAction(actions, "reported-retired-link", ".agents/skills/moai-kanban-foreman") {
			t.Errorf("run %d: the retired link was not reported: %+v", run, actions)
		}
		if got := treeSnapshot(t, root); !sameSnapshot(before, got) {
			t.Fatalf("run %d changed the tree:\nbefore %q\nafter  %q", run, before, got)
		}
		if _, err := os.Lstat(filepath.Join(root, ".moai", "archive")); !os.IsNotExist(err) {
			t.Errorf("run %d: .moai/archive exists (stat err: %v)", run, err)
		}
	}
}
