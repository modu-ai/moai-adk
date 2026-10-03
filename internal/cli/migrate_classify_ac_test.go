package cli

// migrate_classify_ac_test.go — the AC-named migration-classification
// criteria (SPEC-INIT-SHRINK-001 acceptance.md AC-010/AC-013, plan M1),
// measured against the real embedded template render. The classifier lives
// in internal/cli/update (plan M1); these tests drive it through the
// production render adapter.

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// embeddedForTest returns the embedded template tree for fixture reads.
func embeddedForTest(t *testing.T) fs.FS {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	return embedded
}

// migrationTestRender returns the production TemplateRender over the real
// embedded template tree (raw reads; the fixture avoids .tmpl paths so raw
// bytes equal the render).
func migrationTestRender(t *testing.T) update.TemplateRender {
	t.Helper()
	return templateRenderCarriage(embeddedForTest(t), nil, nil)
}

// TestMigrationClassification is AC-010: a fixture old-project tree with one
// file per class classifies identical / modified / foreign / modified, and
// the counts match.
func TestMigrationClassification(t *testing.T) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	identicalPath := ".claude/skills/moai-foundation-core/SKILL.md"
	modifiedPath := ".claude/skills/moai-workflow-tdd/SKILL.md"
	absentRecordPath := ".claude/commands/moai/todo.md"
	foreignPath := ".claude/skills/moai-custom/SKILL.md"

	identicalContent, err := fs.ReadFile(embedded, identicalPath)
	if err != nil {
		t.Fatalf("embedded read %s: %v", identicalPath, err)
	}
	modifiedContent, err := fs.ReadFile(embedded, modifiedPath)
	if err != nil {
		t.Fatalf("embedded read %s: %v", modifiedPath, err)
	}
	todoContent, err := fs.ReadFile(embedded, absentRecordPath)
	if err != nil {
		t.Fatalf("embedded read %s: %v", absentRecordPath, err)
	}

	root := t.TempDir()
	writeMigrationFixtureFile(t, root, identicalPath, string(identicalContent))
	// The modified file is deployed from the render, then edited by the user.
	writeMigrationFixtureFile(t, root, modifiedPath, string(modifiedContent))
	writeMigrationFixtureFile(t, root, absentRecordPath, string(todoContent))
	writeMigrationFixtureFile(t, root, foreignPath, "# moai-custom: the user's own skill\n")

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	// Track both skill copies as template_managed: the identical one's record
	// stays healthy; the modified one's record goes stale when the user edit
	// lands after tracking.
	if err := mgr.Track(identicalPath, manifest.TemplateManaged, manifest.HashBytes(identicalContent)); err != nil {
		t.Fatalf("track identical: %v", err)
	}
	if err := mgr.Track(modifiedPath, manifest.TemplateManaged, manifest.HashBytes(modifiedContent)); err != nil {
		t.Fatalf("track modified: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(modifiedPath)),
		[]byte(string(modifiedContent)+"\n<!-- user edit -->\n"), 0o644); err != nil {
		t.Fatalf("apply user edit: %v", err)
	}

	plan, err := update.ClassifyMigration(root, migrationTestRender(t), mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyMigration: %v", err)
	}

	type want struct {
		path  string
		class update.MigrateClass
	}
	for _, w := range []want{
		{identicalPath, update.ClassIdentical},
		{modifiedPath, update.ClassModified},
		{absentRecordPath, update.ClassModified}, // absent record → conservative archive-then-remove
		{foreignPath, update.ClassForeign},       // the render does not carry it — carriage is the gate
	} {
		if got := plan.ClassOf(w.path); got != w.class {
			t.Errorf("%s classified %q, want %q", w.path, got, w.class)
		}
	}
	if plan.CountIdentical() != 1 || plan.CountModified() != 2 || plan.CountForeign() != 1 {
		t.Errorf("counts = %d/%d/%d, want 1/2/1",
			plan.CountIdentical(), plan.CountModified(), plan.CountForeign())
	}
}

// TestMigrationLeavesForeignFilesUntouched is AC-013 (M1 classification
// half): a foreign user skill at a managed-glob-matching name the render
// does not carry classifies foreign — never identical or modified, so it
// never enters a removal or archive set — and stays byte-unchanged on disk;
// a symlink under a dropped root is never classified at all.
func TestMigrationLeavesForeignFilesUntouched(t *testing.T) {
	foreignPath := ".claude/skills/moai-custom/SKILL.md"
	foreignBody := "# moai-custom\n\nThe user's own skill at a managed-glob name.\n"
	root := t.TempDir()
	writeMigrationFixtureFile(t, root, foreignPath, foreignBody)
	realPath := ".claude/skills/moai-foundation-core/SKILL.md"
	realBody, rerr := fs.ReadFile(embeddedForTest(t), realPath)
	if rerr != nil {
		t.Fatalf("embedded read %s: %v", realPath, rerr)
	}
	writeMigrationFixtureFile(t, root, realPath, string(realBody))

	linkPath := filepath.Join(root, ".claude", "skills", "moai-linkdir")
	if err := os.Symlink(filepath.Join(root, ".claude", "skills", "moai-foundation-core"), linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	plan, err := update.ClassifyMigration(root, migrationTestRender(t), nil)
	if err != nil {
		t.Fatalf("ClassifyMigration: %v", err)
	}

	if got := plan.ClassOf(foreignPath); got != update.ClassForeign {
		t.Errorf("foreign skill classified %q, want %q", got, update.ClassForeign)
	}
	// The classes are exclusive: classified foreign means the file sits in
	// none of the removal/archive sets. The real skill copy carries no
	// manifest record in this fixture, so it routes to the conservative
	// modified class (RK-9) — the foreign file is the only foreign entry.
	if plan.CountIdentical() != 0 || plan.CountModified() != 1 || plan.CountForeign() != 1 {
		t.Errorf("counts = %d/%d/%d, want 0/1/1",
			plan.CountIdentical(), plan.CountModified(), plan.CountForeign())
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(foreignPath)))
	if err != nil {
		t.Fatalf("foreign file disappeared: %v", err)
	}
	if string(got) != foreignBody {
		t.Errorf("foreign file bytes changed:\n%s", string(got))
	}
	if len(plan.Symlinks) != 1 {
		t.Errorf("symlinks = %v, want the one link", plan.Symlinks)
	}
}

func writeMigrationFixtureFile(t *testing.T, root, relSlashPath, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(relSlashPath))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relSlashPath, err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relSlashPath, err)
	}
}
