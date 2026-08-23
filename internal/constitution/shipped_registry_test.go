package constitution_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestShippedRegistriesLoad guards the registry files this repository actually
// ships against a parse break.
//
// The loader takes the FIRST yaml-tagged code fence in the document as the entry
// list, so any yaml-tagged example added above "## Entries" — or that fence
// marker written out literally in prose — silently becomes the registry and the
// file stops loading. Prose edits to these documents are otherwise untested, and
// the CI job that would notice is advisory (continue-on-error).
func TestShippedRegistriesLoad(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	paths := map[string]string{
		"local":    filepath.Join(repoRoot, ".claude", "rules", "moai", "core", "zone-registry.md"),
		"template": filepath.Join(repoRoot, "internal", "template", "templates", ".claude", "rules", "moai", "core", "zone-registry.md"),
	}

	for name, path := range paths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := os.Stat(path); err != nil {
				t.Skipf("registry not present at %s: %v", path, err)
			}

			reg, err := constitution.LoadRegistry(path, repoRoot)
			if err != nil {
				t.Fatalf("LoadRegistry(%s): %v", path, err)
			}
			if len(reg.Entries) == 0 {
				t.Errorf("LoadRegistry(%s) returned 0 entries", path)
			}
		})
	}
}
