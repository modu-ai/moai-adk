package template

// SPEC-INIT-HARNESS-001 M2 — harnessFS unit tests: the codex-only deployer's
// template listing hides claude-only surfaces, keeps the universal set, and
// re-homes the catalog; a real deployment materializes .agents/skills as
// REAL directories with zero symlinks (AC-IH-004 remap integrity).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestCodexOnlyRelocationRespectsCatalogFilter pins the P2 fix
// (SPEC-USER-ASSET-INSTALL-001 leader mid-run finding), re-baselined by the
// t1547 repair round (gate r4 finding 1): the listing now applies the SAME
// common-asset exclusion as the deploy walk (isCommonAssetRoot — REQ-005,
// in any mode), so NO .agents/skills path appears in the listing — the
// unselected-bundle leak assertion holds, and the former L0-presence
// assertions retired with the listing-superset contract they pinned (a real
// deployment writes no .agents/skills — REQ-005; pinned on disk by
// TestCodexOnlyForceUpdateVariant). The catalog remap integrity itself stays
// pinned at the harnessFS Open/ReadDir layer.
func TestCodexOnlyRelocationRespectsCatalogFilter(t *testing.T) {
	d := newCodexOnlyTestDeployer(t)

	seen := make(map[string]bool)
	for _, p := range d.ListTemplates() {
		seen[p] = true
		if strings.HasPrefix(p, ".agents/skills/") {
			t.Errorf("common-asset path %q visible in the listing — the deploy walk never writes it (deploy parity)", p)
		}
	}
	if seen[".agents/skills/moai-workflow-loop/SKILL.md"] {
		t.Error("CATALOG_FILTER_LEAK: .agents/skills/moai-workflow-loop/SKILL.md visible in codex-only deployment — the relocation path bypassed the catalog filter (unselected-bundle skill deployed)")
	}
}

// newCodexOnlyTestDeployer builds the codex-only deployer against the real
// embedded tree (the production path under test).
func newCodexOnlyTestDeployer(t *testing.T) Deployer {
	t.Helper()
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	d, err := NewCodexOnlyDeployerWithRenderer(cat, NewRenderer(embedded))
	if err != nil {
		t.Fatalf("new codex-only deployer: %v", err)
	}
	return d
}

// TestCodexOnlyDeployerWalkIntegrity asserts the template listing: hidden set
// absent, universal set present, catalog re-homed under .agents/skills.
func TestCodexOnlyDeployerWalkIntegrity(t *testing.T) {
	d := newCodexOnlyTestDeployer(t)

	list := d.ListTemplates()
	if len(list) == 0 {
		t.Fatal("codex-only deployer lists zero templates")
	}
	seen := make(map[string]bool, len(list))
	for _, p := range list {
		seen[p] = true
		if p == ".claude" || strings.HasPrefix(p, ".claude/") || p == "CLAUDE.md" || p == ".mcp.json" ||
			p == ".claudeignore" || p == ".moai/status_line.sh" || p == ".moai/status_line.sh.tmpl" {
			t.Errorf("hidden path %q visible in codex-only template listing (REQ-IH-005)", p)
		}
	}
	for _, want := range []string{
		".moai/config/sections/llm.yaml",
		// DEPLOY-TARGET name, not the template-source name. `ListTemplates`
		// strips the `.tmpl` suffix (deployer.go), so the mirror — which ships
		// as `AGENTS.md.tmpl` to stay out of Codex's filename-keyed discovery
		// inside this repo (card t925) — appears here as `AGENTS.md`.
		// The two layers are easy to confuse: `harnessFS.Stat` in
		// apply_harness_test.go sees template-SOURCE names and must be given
		// `AGENTS.md.tmpl`, while this listing sees deploy targets.
		"AGENTS.md",
		".gitignore",
	} {
		if !seen[want] {
			t.Errorf("%q missing from codex-only template listing", want)
		}
	}
}

// TestCodexOnlyForceUpdateVariant deploys through the update-path constructor
// (force-update semantics) and re-checks the codex-only file set.
func TestCodexOnlyForceUpdateVariant(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded: %v", err)
	}
	d, err := NewCodexOnlyDeployerWithRendererAndForceUpdate(cat, NewRenderer(embedded))
	if err != nil {
		t.Fatalf("new codex-only force-update deployer: %v", err)
	}
	rd, ok := d.(ResultDeployer)
	if !ok {
		t.Fatal("force-update codex deployer must satisfy ResultDeployer")
	}

	root := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	// A template context is REQUIRED, not incidental: the deployer only strips
	// the `.tmpl` suffix on the render branch, which it takes when a renderer
	// AND a context are both present. Passing nil here would deploy the mirror
	// verbatim as `AGENTS.md.tmpl` and the assertion below would report a
	// regression that production does not have — both production deploy sites
	// (initializer.go, mirror_notice.go) pass a context (card t925).
	if err := rd.Deploy(context.Background(), root, mgr, NewTemplateContext()); err != nil {
		t.Fatalf("deploy: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md missing from force-update codex deploy: %v", err)
	}
	// ...and the template-source name must NOT survive into the project.
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md.tmpl")); err == nil {
		t.Error("AGENTS.md.tmpl leaked into the deployed project — the .tmpl suffix must be stripped")
	}
	if _, err := os.Stat(filepath.Join(root, ".claude")); err == nil {
		t.Error(".claude/ deployed by the force-update codex deployer")
	}
}

// TestCodexOnlyHiddenPathsErrNotExist pins the Open/Stat hiding contract
// directly: hidden paths answer fs.ErrNotExist through both entry points.
func TestCodexOnlyHiddenPathsErrNotExist(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded: %v", err)
	}
	d, err := NewCodexOnlyDeployerWithRenderer(cat, NewRenderer(embedded))
	if err != nil {
		t.Fatalf("new deployer: %v", err)
	}
	for _, p := range []string{".claude/skills/moai/SKILL.md", "CLAUDE.md", ".mcp.json", ".claudeignore", ".moai/status_line.sh"} {
		// ExtractTemplate wraps ErrTemplateNotFound — the hidden contract is
		// "reads as absent", whichever error shape the entry point wraps.
		if _, err := d.ExtractTemplate(p); err == nil {
			t.Errorf("hidden path %q readable via ExtractTemplate — not hidden", p)
		} else if !errors.Is(err, ErrTemplateNotFound) && !os.IsNotExist(err) {
			t.Errorf("hidden path %q: unexpected error shape: %v", p, err)
		}
		entries := d.ListTemplates()
		for _, listed := range entries {
			if listed == p {
				t.Errorf("hidden path %q visible in listing", p)
			}
		}
	}
}

func TestHarnessProfilesResolveSharedReferences(t *testing.T) {
	cat, err := LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	renderer := NewRenderer(embedded)
	profiles := []struct {
		name        string
		newDeployer func() (Deployer, error)
		wantCodex   bool
	}{
		{"claude", func() (Deployer, error) { return NewClaudeHarnessDeployerWithRenderer(cat, renderer) }, false},
		{"gpt", func() (Deployer, error) { return NewCodexOnlyDeployerWithRenderer(cat, renderer) }, true},
		{"both", func() (Deployer, error) { return NewDualHarnessDeployerWithRenderer(cat, renderer) }, true},
	}
	for _, tc := range profiles {
		t.Run(tc.name, func(t *testing.T) {
			d, err := tc.newDeployer()
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			owned := filepath.Join(root, ".moai", "policies", "README.md")
			if err := os.MkdirAll(filepath.Dir(owned), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(owned, []byte("user policy note\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			mgr := manifest.NewManager()
			if _, err := mgr.Load(root); err != nil {
				t.Fatal(err)
			}
			if err := d.Deploy(context.Background(), root, mgr, NewTemplateContext()); err != nil {
				t.Fatal(err)
			}
			if got, err := os.ReadFile(owned); err != nil || string(got) != "user policy note\n" {
				t.Errorf("user policy note changed: bytes=%q error=%v", got, err)
			}
			for _, rel := range []string{".moai/policies/core/moai-constitution.md", ".moai/workflows/plan.md", "AGENTS.md"} {
				if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
					t.Errorf("required shared reference %s: %v", rel, err)
				}
			}
			// AGENTS.md-primary product: AGENTS.md (deployed from
			// AGENTS.md.tmpl) is the sole instruction file for every harness
			// value; CLAUDE.md is never deployed on any profile.
			if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
				t.Errorf("AGENTS.md presence = %v, want true on every profile", err == nil)
			}
			if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); err == nil {
				t.Error("CLAUDE.md presence = true, want false on every profile")
			}
			// SPEC-USER-ASSET-INSTALL-001 (REQ-005): NO common agent or
			// skill lands project-side in any profile — the user installer
			// owns both (its collision/confinement semantics are covered in
			// internal/userassets). The profile deploy only owes the shared
			// references and the harness-axis files.
			if _, err := os.Stat(filepath.Join(root, ".codex", "agents", "moai")); !os.IsNotExist(err) {
				t.Error("project .codex/agents/moai placement exists — REQ-005 forbids it")
			}
			if _, err := os.Stat(filepath.Join(root, ".agents", "skills")); !os.IsNotExist(err) {
				t.Error("project .agents/skills placement exists — REQ-005 forbids it")
			}
		})
	}
}
