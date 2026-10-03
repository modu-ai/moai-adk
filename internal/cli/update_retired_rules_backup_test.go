package cli

// Update fixture for the retired factory-dispatch rule paths (SPEC-LAUNCHER-
// ENTRY-FLAGS-001 AC-020, REQ-025). The update's managed clean removes
// `.claude/rules/moai` as a root and backs up, into the run's pre-clean
// directory, every file the embedded template does not carry at the same
// relative path. Renaming the three dispatch rules makes the old paths
// "not carried", so a copy a user kept (modified or not) must reach the backup
// before the root is removed; a path the template still carries is rewritten by
// deployment and is not backed up. The test builds a project with the real
// embedded template and runs the clean twice.

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// retiredRuleFixtureDir is the managed-root directory the three old rule files
// live in, relative to the project root.
func retiredRuleFixtureDir() string {
	return filepath.Join(defs.ClaudeDir, defs.RulesMoaiSubdir, "workflow")
}

// retiredRuleSnapshot maps every regular file under root to its bytes, so two
// snapshots compare the whole project tree.
func retiredRuleSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		snap[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return snap
}

func TestUpdateRemovesRetiredRuleFilesWithBackup(t *testing.T) {
	root := t.TempDir()

	const (
		shippedName    = "kanban-dispatch.md"
		modifiedName   = "kanban-dispatch-detail.md"
		absentName     = "kanban-dispatch-mechanics.md"
		shippedBody    = "the rule text as a past release shipped it\n"
		modifiedBody   = "the rule text as a past release shipped it\nplus a line the user added\n"
		bystanderSkill = "user skill that no managed root owns\n"
		bystanderRule  = "user rule under the local namespace\n"
	)

	workflowDir := filepath.Join(root, retiredRuleFixtureDir())
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(workflowDir, shippedName), shippedBody)
	write(filepath.Join(workflowDir, modifiedName), modifiedBody)
	// absentName is never written: the third case is a project that already
	// lost the file.
	skillPath := filepath.Join(root, defs.ClaudeDir, defs.SkillsSubdir, "my-own-skill", "SKILL.md")
	localRulePath := filepath.Join(root, defs.ClaudeDir, "rules", "local", "note.md")
	write(skillPath, bystanderSkill)
	write(localRulePath, bystanderRule)

	tmplFS, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}

	var firstOut bytes.Buffer
	if err := deploy.CleanMoaiManagedPaths(root, &firstOut, tmplFS); err != nil {
		t.Fatalf("first clean: %v", err)
	}
	afterFirst := retiredRuleSnapshot(t, root)

	var secondOut bytes.Buffer
	if err := deploy.CleanMoaiManagedPaths(root, &secondOut, tmplFS); err != nil {
		t.Fatalf("second clean: %v", err)
	}
	afterSecond := retiredRuleSnapshot(t, root)

	// backupCopy returns the bytes of a retired file in the run's pre-clean
	// backup, or ok=false when the backup holds no copy.
	backupCopy := func(name string) (string, bool) {
		matches, gerr := filepath.Glob(filepath.Join(root, defs.BackupsDir, "*", "pre-clean", retiredRuleFixtureDir(), name))
		if gerr != nil {
			t.Fatalf("glob backup: %v", gerr)
		}
		if len(matches) == 0 {
			return "", false
		}
		if len(matches) > 1 {
			t.Fatalf("backup holds %d copies of %s: %v", len(matches), name, matches)
		}
		data, rerr := os.ReadFile(matches[0])
		if rerr != nil {
			t.Fatalf("read backup copy: %v", rerr)
		}
		return string(data), true
	}

	assertRemoved := func(t *testing.T, name string) {
		t.Helper()
		if _, serr := os.Stat(filepath.Join(workflowDir, name)); !os.IsNotExist(serr) {
			t.Errorf("%s is still under the live tree after the clean: %v", name, serr)
		}
	}

	t.Run("unmodified", func(t *testing.T) {
		assertRemoved(t, shippedName)
		got, ok := backupCopy(shippedName)
		if !ok {
			t.Fatalf("%s has no copy in the pre-clean backup; the clean removed it without backing it up", shippedName)
		}
		if got != shippedBody {
			t.Errorf("backup copy of %s = %q, want the original bytes %q", shippedName, got, shippedBody)
		}
	})

	t.Run("user-modified", func(t *testing.T) {
		assertRemoved(t, modifiedName)
		got, ok := backupCopy(modifiedName)
		if !ok {
			t.Fatalf("%s has no copy in the pre-clean backup; a user-modified file was removed without a backup", modifiedName)
		}
		if got != modifiedBody {
			t.Errorf("backup copy of %s = %q, want the user's bytes %q", modifiedName, got, modifiedBody)
		}
	})

	t.Run("absent", func(t *testing.T) {
		assertRemoved(t, absentName)
		if got, ok := backupCopy(absentName); ok {
			t.Errorf("%s was absent from the project but the backup holds %q", absentName, got)
		}
	})

	t.Run("bystanders-and-reruns", func(t *testing.T) {
		if got := afterFirst[mustRel(t, root, skillPath)]; got != bystanderSkill {
			t.Errorf("user skill outside the managed roots = %q, want %q", got, bystanderSkill)
		}
		if got := afterFirst[mustRel(t, root, localRulePath)]; got != bystanderRule {
			t.Errorf("user rule under rules/local = %q, want %q", got, bystanderRule)
		}
		if !strings.Contains(firstOut.String(), "backed up") {
			t.Errorf("the progress output does not report the backup:\n%s", firstOut.String())
		}
		if len(afterFirst) != len(afterSecond) {
			t.Errorf("the second clean changed the project: %d files before, %d after", len(afterFirst), len(afterSecond))
		}
		for rel, body := range afterFirst {
			if afterSecond[rel] != body {
				t.Errorf("the second clean changed %s", rel)
			}
		}
	})
}

// mustRel returns path relative to root, failing the test when it cannot.
func mustRel(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("rel %s: %v", path, err)
	}
	return rel
}
