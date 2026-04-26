// SPEC-V3R3-HARNESS-001 / BC-V3R3-007
// Asserts that the 16 removed skill directories are absent from the template tree.
// RED phase: directories still exist → test fails.
// GREEN phase: directories deleted via git rm -rf → test passes.

package template

import (
	"os"
	"path/filepath"
	"runtime"
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

	// 16-skill removal list from spec.md §3, in canonical order.
	removed := []string{
		"moai-domain-backend",
		"moai-domain-frontend",
		"moai-domain-database",
		"moai-domain-db-docs",
		"moai-domain-mobile",
		"moai-framework-electron",
		"moai-library-shadcn",
		"moai-library-mermaid",
		"moai-library-nextra",
		"moai-tool-ast-grep",
		"moai-platform-auth",
		"moai-platform-deployment",
		"moai-platform-chrome-extension",
		"moai-workflow-research",
		"moai-workflow-pencil-integration",
		"moai-formats-data",
	}

	for _, name := range removed {
		name := name // capture loop variable
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
