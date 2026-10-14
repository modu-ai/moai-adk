package skillresidue

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readOrFail(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func hasAction(actions []Action, kind, path string) bool {
	for _, a := range actions {
		if a.Kind == kind && a.Path == path {
			return true
		}
	}
	return false
}

// exactNames lists a folder's entry names as stored. On a case-insensitive
// filesystem, os.Stat cannot tell skill.md from SKILL.md, so these tests compare
// the stored names instead.
func exactNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func containsName(names []string, want string) bool {
	return slices.Contains(names, want)
}

// caseSensitiveDir skips the test when the directory's filesystem folds case,
// because skill.md and SKILL.md would then be one file.
func caseSensitiveDir(t *testing.T, dir string) {
	t.Helper()
	mustMkdirAll(t, dir)
	probe := filepath.Join(dir, "case-probe.txt")
	mustWriteFile(t, probe, "x")
	if _, err := os.Stat(filepath.Join(dir, "CASE-PROBE.TXT")); err == nil {
		_ = os.Remove(probe)
		t.Skip("case-insensitive filesystem: skill.md and SKILL.md cannot coexist")
	}
	_ = os.Remove(probe)
}

// Condition 1: an empty folder is removed only when moai ships that skill. A
// folder of a name the template does not ship is the user's, and it stays.
func TestClean_RemovesEmptyFolderOnlyForMoaiSkillNames(t *testing.T) {
	root := t.TempDir()
	shippedEmpty := filepath.Join(root, ".claude", "skills", "moai-lane-watchdog")
	userEmpty := filepath.Join(root, ".claude", "skills", "my-private-notes")
	mustMkdirAll(t, shippedEmpty)
	mustMkdirAll(t, userEmpty)
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-lane-watchdog/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped, Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if _, err := os.Stat(shippedEmpty); !os.IsNotExist(err) {
		t.Fatalf("empty folder of a shipped skill still present: err=%v", err)
	}
	if !hasAction(actions, "removed-empty-dir", ".claude/skills/moai-lane-watchdog") {
		t.Fatalf("actions = %+v, want removed-empty-dir for the shipped skill", actions)
	}
	if _, err := os.Stat(userEmpty); err != nil {
		t.Fatalf("user-created empty folder was touched: %v", err)
	}
	if hasAction(actions, "removed-empty-dir", ".claude/skills/my-private-notes") {
		t.Fatalf("user-created empty folder reported as removed: %+v", actions)
	}
}

// Condition 2 (rename): a lowercase skill.md becomes SKILL.md only when SKILL.md
// is absent and the template ships it, and the content moves unchanged.
func TestClean_RenamesLowercaseSkillFileWhenCanonicalAbsent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-ref-owasp-checklist")
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-ref-owasp-checklist/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped, Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	names := exactNames(t, dir)
	if !containsName(names, "SKILL.md") || containsName(names, "skill.md") {
		t.Fatalf("entry names = %v, want exactly SKILL.md", names)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "lower body" {
		t.Fatalf("renamed content = %q, want the lowercase body", got)
	}
	if !hasAction(actions, "renamed-skill-file", ".claude/skills/moai-ref-owasp-checklist/skill.md") {
		t.Fatalf("actions = %+v, want renamed-skill-file", actions)
	}
}

// Condition 2 (both present): when both names exist, nothing is renamed or
// overwritten; the pair is reported for a human to resolve.
func TestClean_ReportsButKeepsBothSkillFilesPresent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-ref-owasp-checklist")
	caseSensitiveDir(t, dir)
	mustWriteFile(t, filepath.Join(dir, "skill.md"), "lower body")
	mustWriteFile(t, filepath.Join(dir, "SKILL.md"), "canonical body")
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-ref-owasp-checklist/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped, Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md")); got != "lower body" {
		t.Fatalf("lowercase file changed: %q", got)
	}
	if got := readOrFail(t, filepath.Join(dir, "SKILL.md")); got != "canonical body" {
		t.Fatalf("canonical file changed: %q", got)
	}
	if !hasAction(actions, "reported-both-skill-files", ".claude/skills/moai-ref-owasp-checklist") {
		t.Fatalf("actions = %+v, want reported-both-skill-files", actions)
	}
	if hasAction(actions, "renamed-skill-file", ".claude/skills/moai-ref-owasp-checklist/skill.md") {
		t.Fatalf("renamed although both names exist: %+v", actions)
	}
}

// Condition 3: the retired folder moves to the archive, and the .agents link
// that pointed at it is removed in the same pass, so no link dangles.
func TestClean_ArchivesRetiredFolderAndRemovesItsLinkWhenSuccessorShipped(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(old, "SKILL.md"), "old body")
	link := filepath.Join(root, ".agents", "skills", "moai-kanban-foreman")
	mustMkdirAll(t, filepath.Dir(link))
	if err := os.Symlink("../../.claude/skills/moai-kanban-foreman", link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped, Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("retired folder still in place: err=%v", err)
	}
	archived := filepath.Join(root, ".moai", "archive", "skills", "retired-20260101T000000Z", "moai-kanban-foreman", "SKILL.md")
	if got := readOrFail(t, archived); got != "old body" {
		t.Fatalf("archived copy = %q, want the retired body", got)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("link to the archived folder still present: err=%v", err)
	}
	if !hasAction(actions, "archived-retired-skill", ".claude/skills/moai-kanban-foreman") ||
		!hasAction(actions, "removed-stale-link", ".agents/skills/moai-kanban-foreman") {
		t.Fatalf("actions = %+v, want the archive and link actions", actions)
	}
}

// Condition 3 (user files): a retired folder holding anything besides its
// SKILL.md is not moved, and the link to it stays, so the user's files are never
// relocated or removed.
func TestClean_KeepsRetiredFolderWithUserFilesAndItsLink(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(old, "SKILL.md"), "old body")
	mustWriteFile(t, filepath.Join(old, "notes.md"), "my own notes")
	link := filepath.Join(root, ".agents", "skills", "moai-kanban-foreman")
	mustMkdirAll(t, filepath.Dir(link))
	if err := os.Symlink("../../.claude/skills/moai-kanban-foreman", link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-factory-foreman/SKILL.md" }

	actions, err := Clean(root, Options{Shipped: shipped, Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if got := readOrFail(t, filepath.Join(old, "notes.md")); got != "my own notes" {
		t.Fatalf("user file changed or removed: %q", got)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("link to a folder with user files was removed: %v", err)
	}
	if !hasAction(actions, "kept-retired-skill-with-user-files", ".claude/skills/moai-kanban-foreman") {
		t.Fatalf("actions = %+v, want kept-retired-skill-with-user-files", actions)
	}
	if hasAction(actions, "archived-retired-skill", ".claude/skills/moai-kanban-foreman") {
		t.Fatalf("folder with user files was archived: %+v", actions)
	}
}

// Without the successor in the template, the retired folder is the only copy of
// that skill, so it stays.
func TestClean_KeepsRetiredSkillWhenSuccessorNotShipped(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, ".claude", "skills", "moai-kanban-foreman")
	mustWriteFile(t, filepath.Join(old, "SKILL.md"), "old body")

	actions, err := Clean(root, Options{Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if _, err := os.Stat(filepath.Join(old, "SKILL.md")); err != nil {
		t.Fatalf("retired skill removed without a successor: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("actions = %+v, want none", actions)
	}
}

// A second pass over a repaired tree finds nothing left to do.
func TestClean_SecondRunIsIdempotent(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, ".claude", "skills", "moai-lane-watchdog"))
	mustWriteFile(t, filepath.Join(root, ".claude", "skills", "moai-kanban-foreman", "SKILL.md"), "old body")
	mustWriteFile(t, filepath.Join(root, ".claude", "skills", "moai-ref-owasp-checklist", "skill.md"), "lower body")
	shipped := func(rel string) bool {
		return rel == ".claude/skills/moai-lane-watchdog/SKILL.md" ||
			rel == ".claude/skills/moai-factory-foreman/SKILL.md" ||
			rel == ".claude/skills/moai-ref-owasp-checklist/SKILL.md"
	}
	opts := Options{Shipped: shipped, Stamp: "20260101T000000Z"}

	if _, err := Clean(root, opts); err != nil {
		t.Fatalf("first Clean: %v", err)
	}
	actions, err := Clean(root, opts)
	if err != nil {
		t.Fatalf("second Clean: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("second pass acted again: %+v", actions)
	}
}

// A project without a skills folder is left exactly as it is.
func TestClean_NoSkillsFolderIsNoOp(t *testing.T) {
	root := t.TempDir()
	actions, err := Clean(root, Options{Stamp: "20260101T000000Z"})
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("actions = %+v, want none", actions)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude")); !os.IsNotExist(err) {
		t.Fatalf(".claude created in a project without skills: err=%v", err)
	}
}
