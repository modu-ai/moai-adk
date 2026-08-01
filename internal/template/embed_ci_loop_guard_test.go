package template

import (
	"io/fs"
	"strings"
	"testing"
)

// ciLoopSkillPrefix is the deployment-relative directory the retired ci-loop
// skill used to occupy. Paths returned by EmbeddedTemplates() already have the
// "templates/" prefix stripped, so this is the exact prefix to match against.
const ciLoopSkillPrefix = ".claude/skills/moai-workflow-ci-loop/"

// TestEmbeddedTemplatesExcludeCILoopSkill asserts that the compiled embed FS —
// the actual deployment artifact, not the source tree — carries zero files
// under the retired ci-loop skill directory.
//
// Why the embed FS and not a source-tree grep: the skill is compiled into the
// binary via //go:embed all:templates, so a source-tree grep alone cannot
// prove the skill stopped shipping. Walking the FS that EmbeddedTemplates()
// returns reads the deployment artifact directly.
//
// KNOWN LIMITATION — this test CANNOT detect a missing `make build`.
// `go test` compiles the package, so //go:embed always reflects the current
// source tree; a stale binary on disk is invisible here. Rebuild-staleness is
// caught by the catalog hash-freshness gate
// (internal/template/scripts/gen-catalog-hashes.go --all + a clean
// `git diff` on catalog.yaml), not by this test.
func TestEmbeddedTemplatesExcludeCILoopSkill(t *testing.T) {
	tmplFS, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error = %v, want nil", err)
	}

	var found []string
	err = fs.WalkDir(tmplFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, ciLoopSkillPrefix) {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir over embedded templates: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("embedded templates still ship %d file(s) under %s: %v",
			len(found), ciLoopSkillPrefix, found)
	}
}
