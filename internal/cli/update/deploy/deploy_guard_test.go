package deploy

// deploy_guard_test.go — SPEC-UPDATE-MIGRATION-001 (card t1547): the
// REQ-UPM-015 guard on the retained wholesale path (AC-UPM-033, legacy-arm
// half; the default-path half is the reconciliation pipeline's own contract
// and is proven by the hazard tests in internal/cli/update).

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestCleanGuarded_RefusesProtectedFiles — AC-UPM-013/015 legacy arm: the
// guarded wholesale walk clears template-carried and unprotected entries but
// refuses to delete a user-owned file and an unresolved user-modified file,
// each with its own skip-with-report progress line.
func TestCleanGuarded_RefusesProtectedFiles(t *testing.T) {
	const (
		templateOwned = "# template render (current)\n"
		userOwned     = "# operator's local rule — exists in no template\n"
		userModified  = "# template render — with the operator's local edits\n"
		staleUnprot   = "# not carried, not protected — clears\n"
	)
	root := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), defs.DirPerm); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), defs.FilePerm); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write(".claude/rules/moai/update-note.md", templateOwned)
	write(".claude/rules/moai/dev-only-rule.md", userOwned)
	write(".claude/rules/moai/tuned.md", userModified)
	write(".claude/rules/moai/old-unprotected.md", staleUnprot)

	tmplFS := fstest.MapFS{
		".claude/rules/moai/update-note.md": &fstest.MapFile{Data: []byte(templateOwned)},
	}

	// The production predicate's shape (the composition itself — namespace
	// predicate ∪ unresolved manifest modification — is asserted in the
	// internal/cli/update package tests over update.ProtectFuncFor).
	protected := map[string]bool{
		".claude/rules/moai/dev-only-rule.md": true,
		".claude/rules/moai/tuned.md":         true,
	}
	protect := func(rel string) bool { return protected[rel] }

	targets := []CleanTarget{{
		DisplayPath: ".claude/rules/moai",
		FullPath:    filepath.Join(root, ".claude", "rules", "moai"),
	}}
	if err := CleanMoaiManagedPathsWithTargetsGuarded(root, io.Discard, tmplFS, targets, protect); err != nil {
		t.Fatalf("CleanMoaiManagedPathsWithTargetsGuarded: %v", err)
	}

	// Protected files survive byte-for-byte.
	for rel, want := range map[string]string{
		".claude/rules/moai/dev-only-rule.md": userOwned,
		".claude/rules/moai/tuned.md":         userModified,
	} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("protected file %s was deleted: %v", rel, err)
		}
		if string(data) != want {
			t.Errorf("protected file %s content = %q, want untouched %q", rel, data, want)
		}
	}
	// Template-carried and unprotected entries cleared.
	for _, rel := range []string{".claude/rules/moai/update-note.md", ".claude/rules/moai/old-unprotected.md"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("%s still exists — the guard must not block the clear", rel)
		}
	}
	// The emptied directory was pruned; the one holding protected entries stays.
	if _, err := os.Stat(filepath.Join(root, ".claude", "rules", "moai")); err != nil {
		t.Fatalf("directory holding protected entries must stay: %v", err)
	}
}

// TestCleanGuarded_NilProtectMatchesUnguardedBehavior documents the guard's
// composition neutrality: a nil ProtectFunc degrades to the unguarded walk's
// behavior (the historic legacy form), so the guard adds a constraint
// without silently changing the machinery.
func TestCleanGuarded_NilProtectMatchesUnguardedBehavior(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "rules", "moai")
	if err := os.MkdirAll(dir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x\n"), defs.FilePerm); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	tmplFS := fstest.MapFS{}
	targets := []CleanTarget{{DisplayPath: ".claude/rules/moai", FullPath: dir}}
	if err := CleanMoaiManagedPathsWithTargetsGuarded(root, io.Discard, tmplFS, targets, nil); err != nil {
		t.Fatalf("guarded walk: %v", err)
	}
	for _, name := range []string{"a.md", "b.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("%s still exists with nil protect — expected the unguarded behavior", name)
		}
	}
}
