// workflow_split_test.go: Audit suite for SPEC-V3R4-WORKFLOW-SPLIT-001
//
// Three test functions enforce structural invariants for the wave-by-wave
// workflow skill split (Bundle F). These tests serve as CI-permanent fixtures
// that prevent future regressions once the split is complete.
//
// Test functions:
//   - TestSubSkillLOCCeiling: Each workflows/{name}/*.md file MUST be ≤500 LOC
//   - TestEntryRouterLOCCeiling: Each of the 4 entry router .md files MUST be ≤200 LOC
//     (currently SKIPPED at baseline-RED — see comment on the skip)
//   - TestTemplateMirrorParity: local and template/templates directories MUST contain
//     identical .md file paths under .claude/skills/moai/workflows/
package skills_test

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// findProjectRoot walks upward from the test working directory until it finds go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}

// countLines returns the number of lines in a file.
func countLines(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			t.Errorf("close %s: %v", path, cerr)
		}
	}()

	count := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		count++
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return count
}

// TestSubSkillLOCCeiling asserts that every .md file under workflows/{name}/*.md
// (i.e., at depth >= 2 from the workflows directory) is ≤500 LOC.
//
// At Wave 0 baseline (before Wave 1 split), no sub-skill files exist yet — the
// walk finds zero files and the test PASSES trivially. Wave 1-4 will populate
// sub-skill files, and any violation will cause CI failure.
//
// REQ: REQ-WFSP-001a, AC-WFSP-001
func TestSubSkillLOCCeiling(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	workflowsDir := filepath.Join(root, ".claude", "skills", "moai", "workflows")

	const maxLOC = 500

	violations := 0
	err := filepath.WalkDir(workflowsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Ignore walk errors (directory may not exist for some sub-dirs yet)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		// Only check files at depth >= 2 from workflowsDir (i.e., in sub-directories)
		rel, _ := filepath.Rel(workflowsDir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) < 2 {
			// Top-level .md file (entry router) — not a sub-skill, skip
			return nil
		}

		loc := countLines(t, path)
		if loc > maxLOC {
			t.Errorf("SUB_SKILL_LOC_VIOLATION: %s has %d LOC (ceiling: %d)", rel, loc, maxLOC)
			violations++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir %s: %v", workflowsDir, err)
	}

	if violations == 0 {
		// No sub-skill files exist yet or all within ceiling — PASS
		t.Logf("TestSubSkillLOCCeiling: PASS (0 violations)")
	}
}

// TestEntryRouterLOCCeiling asserts that each of the 4 workflow entry router
// .md files is ≤200 LOC.
//
// Wave 4 complete: all 4 entry routers (run/sync/project/plan) have been
// refactored to ≤200 LOC thin routers. This test enforces the 200 LOC ceiling
// permanently, preventing future regressions.
//
// REQ: REQ-WFSP-002a, AC-WFSP-002
func TestEntryRouterLOCCeiling(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	workflowsDir := filepath.Join(root, ".claude", "skills", "moai", "workflows")

	entryRouters := []string{"run.md", "sync.md", "project.md", "plan.md"}
	const maxLOC = 200

	for _, name := range entryRouters {
		path := filepath.Join(workflowsDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("ENTRY_ROUTER_MISSING: %s does not exist", name)
			continue
		}
		loc := countLines(t, path)
		if loc > maxLOC {
			t.Errorf("ENTRY_ROUTER_LOC_VIOLATION: %s has %d LOC (ceiling: %d)", name, loc, maxLOC)
		}
	}
}

// TestTemplateMirrorParity asserts that the set of .md files under
// .claude/skills/moai/workflows/ in the local tree and under
// internal/template/templates/.claude/skills/moai/workflows/ are identical.
//
// At Wave 0 baseline, both directories contain only the 4 entry router .md
// files at top level (plus any files added by Wave 0 scaffolding). This test
// PASSES as long as both trees have the same relative paths.
//
// Dev-only files (CLAUDE.local.md §21): certain .md files are intentionally
// present in local but excluded from the template distribution (e.g.,
// release-update.md, github.md). These are listed in devOnlyLocalFiles and
// excluded from parity checking.
//
// REQ: REQ-WFSP-004d, AC-WFSP-008
func TestTemplateMirrorParity(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	localDir := filepath.Join(root, ".claude", "skills", "moai", "workflows")
	templateDir := filepath.Join(root, "internal", "template", "templates", ".claude", "skills", "moai", "workflows")

	// Dev-only files: present in local but intentionally excluded from template
	// distribution (CLAUDE.local.md §21 — dev-only commands isolation).
	devOnlyLocalFiles := map[string]bool{
		"release-update.md": true,
		"github.md":         true,
	}

	localFiles := collectMDFiles(t, localDir)
	templateFiles := collectMDFiles(t, templateDir)

	// Exclude dev-only files from local before comparison
	for rel := range devOnlyLocalFiles {
		delete(localFiles, rel)
	}

	// Find missing in template (local has it but template does not)
	for rel := range localFiles {
		if !templateFiles[rel] {
			t.Errorf("TEMPLATE_MIRROR_MISSING: %s exists in local but not in template mirror", rel)
		}
	}

	// Find extra in template (template has it but local does not)
	for rel := range templateFiles {
		if !localFiles[rel] {
			t.Errorf("TEMPLATE_MIRROR_EXTRA: %s exists in template mirror but not in local", rel)
		}
	}

	if !t.Failed() {
		sortedPaths := make([]string, 0, len(localFiles))
		for rel := range localFiles {
			sortedPaths = append(sortedPaths, rel)
		}
		sort.Strings(sortedPaths)
		t.Logf("TestTemplateMirrorParity: PASS (%d files in sync)", len(sortedPaths))
	}
}

// collectMDFiles returns a set of relative .md file paths under dir.
// Returns an empty map if dir does not exist (vacuously pass for missing dirs).
func collectMDFiles(t *testing.T, dir string) map[string]bool {
	t.Helper()
	result := make(map[string]bool)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory does not exist yet; vacuously OK
		return result
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		result[rel] = true
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir %s: %v", dir, err)
	}
	return result
}
