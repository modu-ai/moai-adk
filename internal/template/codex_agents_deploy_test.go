// codex_agents_deploy_test.go — SPEC-CODEX-DUAL-AGENTS-001 MS3 AC-010 deploy
// fixture: the .codex/ root rides the existing template deployment untouched
// — Deploy(embedded FS) into a t.TempDir() project must land
// .codex/agents/moai/ with the 11 agent TOMLs byte-equal to the committed
// sources. This verifies the deploy premise mechanically instead of assuming
// it (the deployer walks every embedded file; the test proves the dot-dir
// root survives the whole path).
package template_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

func TestCodexAgentsDeployFixture(t *testing.T) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}

	projectRoot := t.TempDir()
	// Manifest setup mirrors deployer_test.go setupDeployProject: the
	// manifest store lives at <root>/.moai/manifest.json and must Load once.
	if err := os.MkdirAll(filepath.Join(projectRoot, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		t.Fatalf("manifest Load: %v", err)
	}
	deployer := template.NewDeployer(embedded)
	if err := deployer.Deploy(context.Background(), projectRoot, mgr, nil); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	deployedDir := filepath.Join(projectRoot, ".codex", "agents", "moai")
	entries, err := os.ReadDir(deployedDir)
	if err != nil {
		t.Fatalf("deployed .codex/agents/moai missing: %v", err)
	}

	// Committed sources for byte comparison.
	srcDir := filepath.Join("templates", ".codex", "agents", "moai")
	srcEntries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("committed .codex/agents/moai missing: %v", err)
	}
	srcSet := map[string]bool{}
	for _, e := range srcEntries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".toml") {
			srcSet[e.Name()] = true
		}
	}
	if len(srcSet) != 11 {
		t.Fatalf("committed sources: %d TOMLs, want 11", len(srcSet))
	}

	deployed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		deployed++
		if !srcSet[e.Name()] {
			t.Errorf("deployed %s has no committed source", e.Name())
			continue
		}
		got, err := os.ReadFile(filepath.Join(deployedDir, e.Name()))
		if err != nil {
			t.Fatalf("read deployed %s: %v", e.Name(), err)
		}
		want, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
		if err != nil {
			t.Fatalf("read committed %s: %v", e.Name(), err)
		}
		if string(got) != string(want) {
			t.Errorf("%s: deployed bytes differ from committed source", e.Name())
		}
	}
	if deployed != 11 {
		t.Errorf("deployed %d TOMLs into .codex/agents/moai, want 11", deployed)
	}
}
