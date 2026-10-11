package deploy

// deploy_characterization_test.go — SPEC-UPDATE-MIGRATION-001 M1 (card t1547,
// DDD PRESERVE phase): pins the CURRENT CleanMoaiManagedPaths behavior before
// the preservation pipeline changes it, so the fix is provable against the
// deletion hazard it exists to close. These tests document what the code
// DOES today — including the destructive parts — not what it should do.
//
// Characterized behaviors (one test each):
//  1. a local-only file under a managed root is copied into the run's
//     pre-clean backup AND then deleted (the t750-documented destruction —
//     recovery is a manual backup restore);
//  2. a template-carried but user-modified file is deleted with NO backup
//     and silently replaced by the next deploy's render (the overwrite
//     hazard — no merge, no report);
//  3. the .moai/config directory is removed wholesale (the wipe the
//     preservation pipeline replaces), with template-absent files reaching
//     the pre-clean backup;
//  4. a clean tree's managed roots are fully emptied by the clean step
//     alone (wipe-first: the end state is produced by the subsequent
//     deployment, not by the clean).

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// charFixtureRoot builds a project tree under t.TempDir with the given map of
// relative slash paths to contents.
func charFixtureRoot(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), defs.DirPerm); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), defs.FilePerm); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}

// charTemplateFS builds an embedded-template FS stand-in carrying exactly the
// given relative slash paths.
func charTemplateFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for rel, content := range files {
		fsys[rel] = &fstest.MapFile{Data: []byte(content)}
	}
	return fsys
}

// findPreCleanBackups returns every file under the fixture's pre-clean backup
// trees, as absolute paths.
func findPreCleanBackups(t *testing.T, root string) []string {
	t.Helper()
	pattern := filepath.Join(root, defs.BackupsDir, "*", preCleanBackupSubdir, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob pre-clean backups: %v", err)
	}
	var found []string
	for _, m := range matches {
		_ = filepath.Walk(m, func(path string, info os.FileInfo, err error) error {
			if err == nil && info != nil && !info.IsDir() {
				found = append(found, path)
			}
			return nil
		})
	}
	return found
}

// TestCharacterizeClean_LocalOnlyFileBackedUpAndDeleted documents the current
// destruction of a local-only file under a managed root: the t111 pre-clean
// backup catches a copy, then the file itself is deleted. Recovery is a
// manual backup restore — the destruction is still the design.
func TestCharacterizeClean_LocalOnlyFileBackedUpAndDeleted(t *testing.T) {
	const localOnly = "# operator's local rule — exists in no template\n"
	root := charFixtureRoot(t, map[string]string{
		".claude/rules/moai/dev-only-rule.md": localOnly,
	})
	tmplFS := charTemplateFS(map[string]string{
		".claude/rules/moai/update-note.md": "# template render\n",
	})

	if err := CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	// Characterized: the file is GONE from the project.
	if _, err := os.Stat(filepath.Join(root, ".claude", "rules", "moai", "dev-only-rule.md")); !os.IsNotExist(err) {
		t.Errorf("local-only file still exists after clean (err=%v) — characterization mismatch", err)
	}

	// Characterized: exactly one copy reached the pre-clean backup, byte-identical.
	backups := findPreCleanBackups(t, root)
	if len(backups) != 1 {
		t.Fatalf("pre-clean backup files = %v, want exactly 1", backups)
	}
	data, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != localOnly {
		t.Errorf("backup content = %q, want %q", data, localOnly)
	}
	if filepath.Base(backups[0]) != "dev-only-rule.md" {
		t.Errorf("backup name = %s, want dev-only-rule.md (layout preserved)", backups[0])
	}
}

// TestCharacterizeClean_UserModifiedTemplateFileOverwrittenWithoutMerge
// documents the silent-overwrite hazard: a template-carried file the user
// edited is deleted WITHOUT any backup (backupThenRemove skips
// template-carried files — "deployment rewrites them moments later") and the
// next deploy replaces it with the fresh render. No merge, no report.
func TestCharacterizeClean_UserModifiedTemplateFileOverwrittenWithoutMerge(t *testing.T) {
	const userEdit = "# template render — with the operator's local edits\n"
	root := charFixtureRoot(t, map[string]string{
		".claude/rules/moai/update-note.md": userEdit,
	})
	tmplFS := charTemplateFS(map[string]string{
		".claude/rules/moai/update-note.md": "# template render (current)\n",
	})

	if err := CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	// Characterized: the user's edited copy is gone.
	if _, err := os.Stat(filepath.Join(root, ".claude", "rules", "moai", "update-note.md")); !os.IsNotExist(err) {
		t.Errorf("user-modified file still exists after clean (err=%v) — characterization mismatch", err)
	}
	// Characterized: and it was NOT backed up — the edit's only copy was the
	// deleted one (this is the data-loss shape the merge disposition closes).
	if backups := findPreCleanBackups(t, root); len(backups) != 0 {
		t.Errorf("pre-clean backup files = %v, want 0 (template-carried files are not backed up)", backups)
	}
}

// TestCharacterizeClean_ConfigDirWiped documents the .moai/config wholesale
// removal: the whole directory is deleted (deploy.go's inline config branch),
// and files the template does not carry reach the pre-clean backup first.
func TestCharacterizeClean_ConfigDirWiped(t *testing.T) {
	root := charFixtureRoot(t, map[string]string{
		".moai/config/sections/custom.yaml":   "custom: true\n",
		".moai/config/sections/system.yaml":   "moai:\n  template_version: 0.0.0\n",
		".moai/config/astgrep-rules/item.yml": "id: x\n",
	})
	tmplFS := charTemplateFS(map[string]string{
		".moai/config/sections/system.yaml": "moai:\n  template_version: 9.9.9\n",
	})

	if err := CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	// Characterized: the entire .moai/config tree is gone, template-carried
	// files included (the Backup step's copy and the post-deploy restore are
	// what bring values back — outside this step's contract).
	if _, err := os.Stat(filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)); !os.IsNotExist(err) {
		t.Errorf(".moai/config still exists after clean (err=%v) — characterization mismatch", err)
	}

	// Characterized: the two template-absent files reached the pre-clean
	// backup; the template-carried system.yaml did not.
	backups := findPreCleanBackups(t, root)
	names := map[string]bool{}
	for _, b := range backups {
		names[filepath.Base(b)] = true
	}
	if !names["custom.yaml"] || !names["item.yml"] {
		t.Errorf("pre-clean backups = %v, want custom.yaml and item.yml", backups)
	}
	if names["system.yaml"] {
		t.Errorf("template-carried system.yaml unexpectedly backed up: %v", backups)
	}
}

// TestCharacterizeClean_ManagedRootsFullyEmptied documents the wipe-first end
// state on a clean tree: after the clean step ALONE, every managed root is
// fully empty — the update's end state is produced entirely by the subsequent
// deployment, which is why the pipeline that replaces this step must classify
// before it removes.
func TestCharacterizeClean_ManagedRootsFullyEmptied(t *testing.T) {
	root := charFixtureRoot(t, map[string]string{
		".claude/settings.json":                   "{}\n",
		".claude/commands/moai/goal.md":           "cmd\n",
		".claude/skills/moai-core/SKILL.md":       "skill\n",
		".claude/rules/moai/rule.md":              "rule\n",
		".claude/output-styles/moai/style.md":     "style\n",
		".claude/hooks/moai/hook.sh":              "hook\n",
		".moai/config/sections/git-strategy.yaml": "git_strategy:\n  worktree_base_branch: develop\n",
	})
	tmplFS := charTemplateFS(map[string]string{
		".claude/settings.json":                   "{}\n",
		".claude/commands/moai/goal.md":           "cmd\n",
		".claude/skills/moai-core/SKILL.md":       "skill\n",
		".claude/rules/moai/rule.md":              "rule\n",
		".claude/output-styles/moai/style.md":     "style\n",
		".claude/hooks/moai/hook.sh":              "hook\n",
		".moai/config/sections/git-strategy.yaml": "git_strategy:\n  worktree_base_branch: \"\"\n",
	})

	if err := CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	for _, rel := range []string{
		".claude/settings.json",
		".claude/commands/moai/goal.md",
		".claude/skills/moai-core/SKILL.md",
		".claude/rules/moai/rule.md",
		".claude/output-styles/moai/style.md",
		".claude/hooks/moai/hook.sh",
		".moai/config/sections/git-strategy.yaml",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("%s still exists after clean (err=%v) — wipe-first characterization mismatch", rel, err)
		}
	}
}
