package userassets

// m7_gate51_remove_test.go — gate round 51-5: RemoveBundle prunes the
// BEFORE/AFTER DEPENDENCY-CLOSURE DIFFERENCE. A parent-only bundle
// (deployment → extras, no own assets) records its dependency assets under
// the DEPENDENCY's label, so the former label-only enumeration left the
// whole dependency subtree (files + records) behind.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

func TestRemoveBundlePrunesDependencyClosure(t *testing.T) {
	f := newFixture(t)
	// The parent-only pack: no own assets, depends on the extras pack.
	f.cat.Catalog.OptionalPacks["deployment"] = &template.Pack{
		Description: "parent-only bundle",
		DependsOn:   []string{"extras"},
	}

	inst := f.installer(t)
	if _, err := inst.Install([]string{"deployment"}); err != nil {
		t.Fatalf("install: %v", err)
	}
	betaPath := filepath.Join(f.home, ".claude", "skills", "moai-beta", "SKILL.md")
	if _, err := os.Stat(betaPath); err != nil {
		t.Fatalf("the dependency asset did not install: %v", err)
	}
	manifest, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatal(err)
	}
	if _, tracked := manifest.Files["claude-skills/moai-beta/SKILL.md"]; !tracked {
		t.Fatal("precondition: the dependency asset is not tracked under its OWN label")
	}

	// Removing the parent-only bundle must prune the dependency closure:
	// the beta file AND its record go.
	res, err := inst.RemoveBundle(manifest, "deployment", nil)
	if err != nil {
		t.Fatalf("RemoveBundle: %v", err)
	}
	if res.Removed == 0 {
		t.Fatalf("the parent-only removal removed nothing — the dependency closure stayed behind (gate 51-5): %+v", res)
	}
	if _, err := os.Stat(betaPath); !os.IsNotExist(err) {
		t.Fatalf("the dependency asset survived the parent bundle removal (stat err = %v)", err)
	}
	// The caller-save contract: the command surface saves the mutated
	// manifest under the lock (bundle.go's flow).
	if err := manifest.Save(ManifestPath(f.home)); err != nil {
		t.Fatal(err)
	}
	manifest2, err := Load(ManifestPath(f.home))
	if err != nil {
		t.Fatal(err)
	}
	if _, tracked := manifest2.Files["claude-skills/moai-beta/SKILL.md"]; tracked {
		t.Fatal("the dependency record survived the parent bundle removal (gate 51-5)")
	}
	// The L0 core stays.
	if _, err := os.Stat(filepath.Join(f.home, ".claude", "skills", "moai-alpha", "SKILL.md")); err != nil {
		t.Fatalf("the L0 core was pruned by the bundle removal: %v", err)
	}
}
