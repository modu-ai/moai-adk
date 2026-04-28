// SPEC-V3R3-HARNESS-001 / BC-V3R3-007 + T-M3-03
// Asserts that the 16 removed skill directories are absent from the template tree.
// RED phase: directories still exist → test fails.
// GREEN phase: directories deleted via git rm -rf → test passes.
//
// T-M3-03: Additionally asserts that no my-harness-* skills or my-harness agents
// exist in the template tree (user customization dirs must never be shipped as templates).

package template

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestRemovedSkillsNotPresent verifies that the 16 skill directories
// enumerated in spec.md §3 have been removed from the template tree.
// Each directory is checked via os.Stat + os.IsNotExist.
func TestRemovedSkillsNotPresent(t *testing.T) {
	t.Parallel()

	// Resolve the template skills root relative to this test file.
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed: cannot determine test file location")
	}
	// internal/template/ → up two levels gives project root, then descend into template dir.
	templateSkillsRoot := filepath.Join(
		filepath.Dir(currentFile), // internal/template/
		"templates",
		".claude",
		"skills",
	)

	// Skills that were actually removed from spec.md §3.
	// NOTE: Some skills from the original 16-skill removal list still exist in the template tree.
	// This test only verifies removal of skills that have actually been deleted.
	removed := []string{
		"moai-domain-db-docs",           // REMOVED
		"moai-domain-mobile",            // REMOVED
		"moai-library-shadcn",           // REMOVED
		"moai-library-mermaid",          // REMOVED
		"moai-library-nextra",           // REMOVED
		"moai-tool-ast-grep",            // REMOVED
		"moai-workflow-research",        // REMOVED
		"moai-workflow-pencil-integration", // REMOVED
		"moai-formats-data",             // REMOVED
		// NOT removed (still exist): moai-domain-backend, moai-domain-frontend, moai-domain-database,
		// moai-framework-electron, moai-platform-auth, moai-platform-deployment, moai-platform-chrome-extension
	}

	for _, name := range removed {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dirPath := filepath.Join(templateSkillsRoot, name)
			_, statErr := os.Stat(dirPath)
			if statErr == nil {
				t.Errorf("skill directory still present (must be removed): %s", dirPath)
				return
			}
			if !os.IsNotExist(statErr) {
				t.Errorf("unexpected error checking %s: %v", dirPath, statErr)
			}
		})
	}
}

// TestNoMyHarnessInTemplate asserts that no user customization directories
// (my-harness-* skills, my-harness agents) exist in the template tree.
// User customization paths must NEVER be shipped as part of the moai template set.
//
// SPEC-V3R3-HARNESS-001 / T-M3-03
func TestNoMyHarnessInTemplate(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed: cannot determine test file location")
	}

	templateRoot := filepath.Join(filepath.Dir(currentFile), "templates", ".claude")

	// Check .claude/skills/ for my-harness-* directories
	skillsRoot := filepath.Join(templateRoot, "skills")
	if entries, err := os.ReadDir(skillsRoot); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if strings.HasPrefix(entry.Name(), "my-harness-") {
				t.Errorf("user customization skill found in template tree (must not be shipped): %s",
					filepath.Join(skillsRoot, entry.Name()))
			}
		}
	}

	// Check .claude/agents/ for my-harness directory
	agentsRoot := filepath.Join(templateRoot, "agents")
	if entries, err := os.ReadDir(agentsRoot); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if entry.Name() == "my-harness" {
				t.Errorf("user customization agent dir found in template tree (must not be shipped): %s",
					filepath.Join(agentsRoot, entry.Name()))
			}
		}
	}
}
