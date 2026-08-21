package template

import (
	"context"
	"strings"
	"testing"
)

// TestSkillMirror_NotTrackedInManifest covers AC-CSC-012: mirror entries are
// never registered with the manifest manager.
//
// This is not a stylistic preference. manifest.Track hashes the path through
// os.Open + io.Copy, and a directory symlink opens fine and then fails EISDIR
// on the copy — so tracking a mirror would surface an error from Deploy and
// collide head-on with the fail-open contract (AC-CSC-013). The mirror is a
// derivative of the canonical tree, which is already tracked, so nothing is
// lost by leaving it out.
func TestSkillMirror_NotTrackedInManifest(t *testing.T) {
	root, mgr := setupDeployProject(t)
	dep, ok := NewDeployer(threeSkillFS()).(ResultDeployer)
	if !ok {
		t.Fatal("deployer does not implement ResultDeployer")
	}
	if _, err := dep.DeployWithResult(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	m := mgr.Manifest()
	if m == nil {
		t.Fatal("manifest not loaded")
	}

	// 1. no key under the mirror root.
	for path := range m.Files {
		if strings.HasPrefix(path, ".agents/") || strings.HasPrefix(path, ".agents"+string('\\')) {
			t.Errorf("manifest tracks mirror path %q — hashing a directory symlink fails EISDIR", path)
		}
	}

	// 2. the canonical skill entries are recorded as before.
	for _, canonical := range []string{
		".claude/skills/moai-alpha/SKILL.md",
		".claude/skills/moai-beta/SKILL.md",
		".claude/skills/moai-beta/modules/extra.md",
		".claude/skills/moai-gamma/SKILL.md",
	} {
		entry, found := mgr.GetEntry(canonical)
		if !found {
			t.Errorf("canonical entry %q missing from the manifest", canonical)
			continue
		}
		if entry.TemplateHash == "" {
			t.Errorf("canonical entry %q has an empty template hash", canonical)
		}
	}
}
