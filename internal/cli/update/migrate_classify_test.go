package update

// migrate_classify_test.go — unit tests for the migration classifier
// (SPEC-INIT-SHRINK-001 REQ-010/REQ-013, plan M1). The AC-named criteria
// (TestMigrationClassification, TestMigrationLeavesForeignFilesUntouched)
// live in the internal/cli package per acceptance.md's command form; these
// unit tests cover the classifier's own table of behaviors against a fake
// render so no embedded-tree fixture is needed here.

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// fakeRender is a TemplateRender over an in-memory FS: a path is carried iff
// the FS holds a regular file at it.
type fakeRender struct {
	fsys fstest.MapFS
}

func (f fakeRender) Carries(relPath string) ([]byte, bool) {
	data, err := f.fsys.ReadFile(relPath)
	if err != nil {
		return nil, false
	}
	return data, true
}

// newClassifyFixture builds a project tree under t.TempDir with the given
// map of relative slash paths to file contents, returning the root.
func newClassifyFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}

// renderWith carries exactly the paths in carried, returning their bytes as
// the render content.
func renderWith(carried map[string]string) TemplateRender {
	fsys := fstest.MapFS{}
	for rel, content := range carried {
		fsys[rel] = &fstest.MapFile{Data: []byte(content)}
	}
	return fakeRender{fsys: fsys}
}

// TestClassifyMigrationThreeClasses walks a fixture holding one file per
// class and asserts the routing (REQ-010: template carriage is the class
// gate; the manifest is the fast path between the two removable classes).
func TestClassifyMigrationThreeClasses(t *testing.T) {
	const (
		renderedSkill  = "# skill rendered\n"
		renderedOther  = "# other rendered\n"
		foreignContent = "# user's own skill\n"
	)
	root := newClassifyFixture(t, map[string]string{
		".claude/skills/moai-a/SKILL.md":  renderedSkill,
		".claude/skills/moai-b/SKILL.md":  "# user changed this\n",
		".claude/skills/moai-custom/note": foreignContent,
		".claude/commands/moai/goal.md":   renderedOther,
	})
	render := renderWith(map[string]string{
		".claude/skills/moai-a/SKILL.md": renderedSkill,
		".claude/skills/moai-b/SKILL.md": renderedSkill,
		".claude/commands/moai/goal.md":  renderedOther,
	})

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	// moai-a: healthy managed record whose content equals the render.
	if err := mgr.Track(".claude/skills/moai-a/SKILL.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track moai-a: %v", err)
	}
	// moai-b: healthy managed record whose content was edited after tracking
	// (stale current hash + content differs from the render).
	if err := mgr.Track(".claude/skills/moai-b/SKILL.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track moai-b: %v", err)
	}
	// goal.md: carried by the render but NO manifest record (RK-9).

	plan, err := ClassifyMigration(root, render, mgr.Manifest())
	if err != nil {
		t.Fatalf("ClassifyMigration: %v", err)
	}

	if got := len(plan.Identical); got != 1 {
		t.Fatalf("identical count = %d (%v), want 1", got, plan.Identical)
	}
	if plan.Identical[0].RelPath != ".claude/skills/moai-a/SKILL.md" {
		t.Errorf("identical file = %s, want the moai-a skill", plan.Identical[0].RelPath)
	}
	if got := len(plan.Modified); got != 2 {
		t.Fatalf("modified count = %d (%v), want 2", got, plan.Modified)
	}
	modifiedPaths := map[string]bool{}
	for _, f := range plan.Modified {
		modifiedPaths[f.RelPath] = true
	}
	if !modifiedPaths[".claude/skills/moai-b/SKILL.md"] {
		t.Errorf("moai-b (content differs) not classified modified: %v", plan.Modified)
	}
	if !modifiedPaths[".claude/commands/moai/goal.md"] {
		t.Errorf("goal.md (absent manifest record) not classified modified: %v", plan.Modified)
	}
	if got := len(plan.Foreign); got != 1 {
		t.Fatalf("foreign count = %d (%v), want 1", got, plan.Foreign)
	}
	if plan.Foreign[0].RelPath != ".claude/skills/moai-custom/note" {
		t.Errorf("foreign file = %s, want moai-custom", plan.Foreign[0].RelPath)
	}
	if plan.CountIdentical() != 1 || plan.CountModified() != 2 || plan.CountForeign() != 1 {
		t.Errorf("count accessors disagree with the sets: %d/%d/%d",
			plan.CountIdentical(), plan.CountModified(), plan.CountForeign())
	}
}

// TestClassifyMigrationGlobHitForeignIsNeverRemovable pins the D-15 loss
// path at the classification layer (leader condition 5a): files whose names
// match the managed clean glob (.claude/skills/moai*) but which the template
// render does not carry are foreign whatever their manifest state — a
// template_managed record included — and never enter a removal or archive
// class. The executor that consumes only the classified sets is M3's
// TestMigrationRemovesIdenticalDroppedComponents.
func TestClassifyMigrationGlobHitForeignIsNeverRemovable(t *testing.T) {
	const userContent = "# moai-custom: the user's own skill, never remove\n"
	for _, tc := range []struct {
		name     string
		manifest func(t *testing.T, root string) *manifest.Manifest
	}{
		{
			name: "missing manifest record",
			manifest: func(t *testing.T, root string) *manifest.Manifest {
				return nil // no manifest at all
			},
		},
		{
			name: "template_managed record present",
			manifest: func(t *testing.T, root string) *manifest.Manifest {
				mgr := manifest.NewManager()
				if _, err := mgr.Load(root); err != nil {
					t.Fatalf("load manifest: %v", err)
				}
				// A healthy template_managed record for the user file — the
				// provenance that would make it removable if managed-name
				// matching were the gate. Carriage is the gate, so the
				// record must not matter: the file is foreign anyway.
				if err := mgr.Track(".claude/skills/moai-custom/SKILL.md", manifest.TemplateManaged, ""); err != nil {
					t.Fatalf("track: %v", err)
				}
				return mgr.Manifest()
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newClassifyFixture(t, map[string]string{
				".claude/skills/moai-custom/SKILL.md":   userContent, // managed-glob hit, user-created
				".claude/skills/moai-usermade/SKILL.md": userContent, // managed-glob hit, user-created
				".claude/commands/moai-userfile.md":     userContent, // user command at a moai- name
			})
			render := renderWith(map[string]string{
				// The render carries unrelated skills only — none of the
				// three fixture files.
				".claude/skills/moai-unrelated/SKILL.md": "# x\n",
			})

			plan, err := ClassifyMigration(root, render, tc.manifest(t, root))
			if err != nil {
				t.Fatalf("ClassifyMigration: %v", err)
			}
			if got := len(plan.Foreign); got != 3 {
				t.Fatalf("foreign count = %d (%v), want 3", got, plan.Foreign)
			}
			if len(plan.Identical) != 0 || len(plan.Modified) != 0 {
				t.Fatalf("foreign files entered a removal class: identical=%v modified=%v",
					plan.Identical, plan.Modified)
			}
			seen := map[string]bool{}
			for _, f := range plan.Foreign {
				seen[f.RelPath] = true
			}
			for _, want := range []string{
				".claude/skills/moai-custom/SKILL.md",
				".claude/skills/moai-usermade/SKILL.md",
				".claude/commands/moai-userfile.md",
			} {
				if !seen[want] {
					t.Errorf("%s not classified foreign: %v", want, plan.Foreign)
				}
			}
		})
	}
}

// TestClassifyMigrationSymlinkNeverClassified pins REQ-013's symlink refusal
// at the classification layer: a link under a dropped root is recorded in
// Symlinks and never lands in any classified set.
func TestClassifyMigrationSymlinkNeverClassified(t *testing.T) {
	root := newClassifyFixture(t, map[string]string{
		".claude/skills/moai-a/SKILL.md": "# real skill\n",
	})
	// A link INSIDE a dropped root pointing at the real skill.
	link := filepath.Join(root, ".claude", "skills", "moai-link")
	if err := os.Symlink(filepath.Join(root, ".claude", "skills", "moai-a"), link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	plan, err := ClassifyMigration(root, renderWith(map[string]string{
		".claude/skills/moai-a/SKILL.md": "# real skill\n",
	}), nil)
	if err != nil {
		t.Fatalf("ClassifyMigration: %v", err)
	}
	if len(plan.Symlinks) != 1 {
		t.Fatalf("symlinks = %v, want the one link", plan.Symlinks)
	}
	if len(plan.Identical)+len(plan.Modified)+len(plan.Foreign) != 1 {
		t.Fatalf("classified sets hold %d files, want only the real one",
			len(plan.Identical)+len(plan.Modified)+len(plan.Foreign))
	}
}

// TestMigrationRootsAreDroppedComponentRoots pins the classified scope
// (REQ-010): the three dropped component roots.
func TestMigrationRootsAreDroppedComponentRoots(t *testing.T) {
	roots := MigrationRoots()
	want := []string{".claude/skills", ".claude/commands", ".agents/skills"}
	if len(roots) != len(want) {
		t.Fatalf("roots = %v, want %v", roots, want)
	}
	for i := range want {
		if roots[i] != want[i] {
			t.Errorf("roots[%d] = %s, want %s", i, roots[i], want[i])
		}
	}
}

// TestMigrationArchiveLayout pins the archive layout constants (REQ-012,
// plan M1): a distinct tag from the legacy v2.16 archive, the skill-file
// root preserving the skill-directory layout, and the standalone-file
// variant carrying the original relative path.
func TestMigrationArchiveLayout(t *testing.T) {
	if MigrationArchiveTag == "v2.16" {
		t.Errorf("migration tag %q collides with the legacy archive tag", MigrationArchiveTag)
	}
	if got := ArchiveSkillFileRoot(); got != filepath.Join(".moai", "archive", "skills", MigrationArchiveTag) {
		t.Errorf("ArchiveSkillFileRoot = %q", got)
	}
	if got := ArchiveFilesRoot(); got != filepath.Join(".moai", "archive", "files", MigrationArchiveTag) {
		t.Errorf("ArchiveFilesRoot = %q", got)
	}
}
