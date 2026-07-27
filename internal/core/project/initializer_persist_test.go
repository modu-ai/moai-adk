package project

// initializer_persist_test.go — Deployer-path persistence integration test
// (SPEC-CLI-WIZARD-RESTRUCTURE-001 C39 / AC-WIZ-010 + AC-WIZ-010a).
//
// This is the artifact that proves the C32 Step-3d relocation actually
// persists. Every other test in this package exercises the FALLBACK path
// (deployer nil), which `moai init` never reaches: the CLI assigns a deployer
// in both branches of its deploy-mode selection and returns early on error, so
// generateConfigsFallback — the former home of the Page-3 writes — is
// structurally unreachable in production. A test that runs with a nil deployer
// therefore cannot detect the defect this SPEC exists to fix.
//
// The initializer here is built the same way internal/cli/init.go builds it:
// embedded FS -> renderer -> real Deployer -> NewInitializer.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// newDeployerPathInitializer mirrors the wiring in internal/cli/init.go: it
// loads the embedded catalog and template FS, builds a renderer, and returns an
// initializer backed by a REAL, non-nil template.Deployer. distributeAll picks
// between the two branches the CLI itself chooses between.
func newDeployerPathInitializer(t *testing.T, distributeAll bool) Initializer {
	t.Helper()

	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("LoadEmbeddedCatalog: %v", err)
	}
	embeddedFS, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	renderer := template.NewRenderer(embeddedFS)

	var deployer template.Deployer
	if distributeAll {
		deployer = template.NewDeployerWithRenderer(embeddedFS, renderer)
	} else {
		deployer, err = template.NewSlimDeployerWithRenderer(cat, renderer)
		if err != nil {
			t.Fatalf("NewSlimDeployerWithRenderer: %v", err)
		}
	}
	// Guard the whole premise of this file: a nil deployer would silently route
	// Init down the fallback path and make every assertion below meaningless.
	if deployer == nil {
		t.Fatal("deployer is nil — this test MUST exercise the deployer path")
	}

	return NewInitializer(deployer, manifest.NewManager(), nil)
}

// sectionPath returns the on-disk path of a .moai/config/sections/ file.
func sectionPath(root, name string) string {
	return filepath.Join(root, ".moai", "config", "sections", name)
}

// yamlLookup parses the file with a real YAML parser and walks a dotted key
// path. Using an independent parser (rather than the production patcher) means
// a bug in patchYAMLPathValue cannot mask itself, and it additionally proves
// the patched document is still structurally valid YAML.
func yamlLookup(t *testing.T, path, dotted string) any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatalf("%s is not valid YAML after patching: %v", path, err)
	}

	var cur any = root
	segments := strings.Split(dotted, ".")
	for i, seg := range segments {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("%s: %q is not a mapping while resolving %q",
				path, strings.Join(segments[:i], "."), dotted)
		}
		cur, ok = m[seg]
		if !ok {
			t.Fatalf("%s: key %q absent while resolving %q", path, seg, dotted)
		}
	}
	return cur
}

// runDeployerInit runs a full deployer-path Init and fails loudly if template
// deployment was skipped — Init records deployment failures as non-fatal
// warnings, which would otherwise turn this test vacuous.
func runDeployerInit(t *testing.T, opts InitOptions, distributeAll bool) *InitResult {
	t.Helper()

	initializer := newDeployerPathInitializer(t, distributeAll)
	result, err := initializer.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, w := range result.Warnings {
		if strings.Contains(w, "template deployment") || strings.Contains(w, "page-3 config") {
			t.Fatalf("deployer path did not complete cleanly: %s", w)
		}
	}
	return result
}

// baseOpts returns InitOptions with everything except the Page-3 answers filled
// in, so each scenario only states what it is actually exercising.
func baseOpts(root, name string) InitOptions {
	return InitOptions{
		ProjectRoot:     root,
		ProjectName:     name,
		Language:        "Go",
		UserName:        "tester",
		ConvLang:        "en",
		DevelopmentMode: "tdd",
	}
}

// TestDeployerPath_Page3AnswersPersist_ScenarioA is AC-WIZ-010 Scenario A:
// project_mode=team, lsp_enabled=true, enforce_quality=false, design_enabled=true,
// claude_design_enabled=false — each read back from its on-disk yaml.
//
// Non-vacuity (acceptance.md AC-WIZ-010, corrected at v0.2.1): against the
// pre-change tree rows 1, 2, 3 and 5 MUST fail, because the Page-3 writes were
// unreachable and each of those four answers differs from its shipped default.
// Row 4 is default-coincident (design.yaml already ships enabled: true) and is
// retained only as a template-default regression guard.
func TestDeployerPath_Page3AnswersPersist_ScenarioA(t *testing.T) {
	root := t.TempDir()

	opts := baseOpts(root, "persist-scenario-a")
	opts.ProjectMode = "team"        // row 1 — shipped default: personal
	opts.LSPEnabled = true           // row 2 — shipped default: false
	opts.EnforceQuality = false      // row 3 — shipped default renders true
	opts.DesignEnabled = true        // row 4 — shipped default: true (coincident)
	opts.ClaudeDesignEnabled = false // row 5 — shipped default: true

	runDeployerInit(t, opts, true)

	cases := []struct {
		row  string
		file string
		path string
		want any
	}{
		{"row 1 project_mode=team", "project.yaml", "project.mode", "team"},
		{"row 2 lsp_enabled=true", "lsp.yaml", "lsp.enabled", true},
		{"row 3 enforce_quality=false", "quality.yaml", "constitution.enforce_quality", false},
		{"row 4 design_enabled=true (default-coincident)", "design.yaml", "design.enabled", true},
		{"row 5 claude_design_enabled=false", "design.yaml", "design.claude_design.enabled", false},
	}
	for _, c := range cases {
		got := yamlLookup(t, sectionPath(root, c.file), c.path)
		if got != c.want {
			t.Errorf("%s: %s %s = %v (%T), want %v", c.row, c.file, c.path, got, got, c.want)
		}
	}
}

// TestDeployerPath_Page3AnswersPersist_ScenarioB is AC-WIZ-010 Scenario B row 6.
// It exists because the REQ-WIZ-006 nesting makes it impossible for both design
// keys to diverge from their shipped defaults in one run: Scenario A can only
// assert design.enabled at its default value, so a writeDesignYAML that dropped
// design.enabled entirely while handling claude_design.enabled correctly would
// still pass Scenario A. This scenario is what catches that.
func TestDeployerPath_Page3AnswersPersist_ScenarioB(t *testing.T) {
	root := t.TempDir()

	opts := baseOpts(root, "persist-scenario-b")
	opts.DesignEnabled = false // row 6 — shipped default: true
	// claude_design_enabled is hidden by the design nesting and is NOT asserted.

	runDeployerInit(t, opts, true)

	got := yamlLookup(t, sectionPath(root, "design.yaml"), "design.enabled")
	if got != false {
		t.Errorf("row 6 design_enabled=false: design.yaml design.enabled = %v, want false", got)
	}
}

// TestDeployerPath_PersistenceIsNonDestructive is AC-WIZ-010a: the Page-3 writes
// patch one key each, they do not replace the document. The sentinel keys are
// drawn from deep inside each deployed file and the size floors are far above
// what the old wholesale writers produced (2-4 line documents, < 100 bytes), so
// a regression to wholesale writing fails every assertion here.
func TestDeployerPath_PersistenceIsNonDestructive(t *testing.T) {
	root := t.TempDir()

	opts := baseOpts(root, "persist-nondestructive")
	opts.ProjectMode = "team"
	opts.LSPEnabled = true
	opts.EnforceQuality = false
	opts.DesignEnabled = true
	opts.ClaudeDesignEnabled = false

	runDeployerInit(t, opts, true)

	checks := []struct {
		file      string
		sentinels []string
		minBytes  int
	}{
		{"lsp.yaml", []string{"delegate_to_astgrep:", "circuit_breaker:"}, 8000},
		{"design.yaml", []string{"figma:", "brand_context:"}, 2000},
		// harness.yaml is untouched entirely (C36) — it must survive verbatim.
		{"harness.yaml", []string{"default_profile:"}, 7000},
	}
	for _, c := range checks {
		path := sectionPath(root, c.file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", c.file, err)
			continue
		}
		if len(data) < c.minBytes {
			t.Errorf("%s is %d bytes, want > %d — the document was replaced, not patched",
				c.file, len(data), c.minBytes)
		}
		for _, sentinel := range c.sentinels {
			if !strings.Contains(string(data), sentinel) {
				t.Errorf("%s lost deep key %q — the document was replaced, not patched",
					c.file, sentinel)
			}
		}
	}
}

// TestDeployerPath_NestedSameNamedKeysSurvive is AC-WIZ-017 asserted on the real
// post-init files rather than on a fixture: after patching, the indentation
// multiset of every `enabled:` key must be unchanged. A depth-blind patch yields
// {2sp: 5} for design.yaml and {2sp: 2} for lsp.yaml.
func TestDeployerPath_NestedSameNamedKeysSurvive(t *testing.T) {
	root := t.TempDir()

	opts := baseOpts(root, "persist-indentation")
	opts.LSPEnabled = true
	opts.DesignEnabled = true
	opts.ClaudeDesignEnabled = false

	runDeployerInit(t, opts, true)

	for _, c := range []struct {
		file string
		want map[int]int
	}{
		{"design.yaml", map[int]int{2: 1, 4: 3, 6: 1}},
		{"lsp.yaml", map[int]int{2: 1, 4: 1}},
	} {
		data, err := os.ReadFile(sectionPath(root, c.file))
		if err != nil {
			t.Errorf("read %s: %v", c.file, err)
			continue
		}
		assertMultiset(t, "post-init "+c.file, string(data), "enabled", c.want)
	}
}

// TestDeployerPath_SlimDeployerAlsoPersists covers the OTHER real CLI branch.
// internal/cli/init.go picks the slim deployer unless --distribute-all is set,
// so the default `moai init` takes this path; Step 3d must persist on it too.
func TestDeployerPath_SlimDeployerAlsoPersists(t *testing.T) {
	root := t.TempDir()

	opts := baseOpts(root, "persist-slim")
	opts.ProjectMode = "team"
	opts.LSPEnabled = true
	opts.EnforceQuality = false

	runDeployerInit(t, opts, false)

	if got := yamlLookup(t, sectionPath(root, "project.yaml"), "project.mode"); got != "team" {
		t.Errorf("slim deployer: project.mode = %v, want team", got)
	}
	if got := yamlLookup(t, sectionPath(root, "lsp.yaml"), "lsp.enabled"); got != true {
		t.Errorf("slim deployer: lsp.enabled = %v, want true", got)
	}
	if got := yamlLookup(t, sectionPath(root, "quality.yaml"), "constitution.enforce_quality"); got != false {
		t.Errorf("slim deployer: constitution.enforce_quality = %v, want false", got)
	}
}
