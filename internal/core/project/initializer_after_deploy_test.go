package project

// Card t1139 follow-up (sync-audit F6): InitOptions.AfterTemplateDeploy is the
// seam the CLI uses to record the pure template render as the next update's
// merge BASE. Its contract is positional — after a successful deploy, before
// any section-writing step (report.yaml, the Page-3 wizard patches, workflow
// toggles) — so the test observes the section files from INSIDE the hook.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// templateLSPRender is what the deployer double writes for lsp.yaml: the
// template's default, which the wizard patch later turns to true.
const templateLSPRender = "lsp:\n  enabled: false\n"

// sectionDeployer is a Deployer double that writes lsp.yaml at its template
// value, the way the real deployer renders the section files.
type sectionDeployer struct{}

func (sectionDeployer) Deploy(_ context.Context, root string, _ manifest.Manager, _ *template.TemplateContext) error {
	dir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(dir, defs.DirPerm); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, defs.LSPYAML), []byte(templateLSPRender), defs.FilePerm)
}
func (sectionDeployer) ExtractTemplate(string) ([]byte, error) { return nil, nil }
func (sectionDeployer) ListTemplates() []string                { return nil }
func (sectionDeployer) ValidateAll(context.Context, *template.TemplateContext) error {
	return nil
}

func afterDeployOpts(root string, hook func(string)) InitOptions {
	return InitOptions{
		ProjectRoot:         root,
		ProjectName:         "after-deploy",
		Language:            "Go",
		Framework:           "none",
		UserName:            "tester",
		ConvLang:            "en",
		DevelopmentMode:     "tdd",
		LSPEnabled:          true,
		AfterTemplateDeploy: hook,
	}
}

// TestInit_AfterTemplateDeployRunsBeforeSectionWrites pins the ordering: inside
// the hook lsp.yaml still holds the deployed template value and report.yaml
// (Step 3c) does not exist yet; after Init both section writers have run.
func TestInit_AfterTemplateDeployRunsBeforeSectionWrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sections := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)

	calls := 0
	var lspInHook string
	var reportExistedInHook bool
	hook := func(hookRoot string) {
		calls++
		if hookRoot != filepath.Clean(root) {
			t.Errorf("hook root = %q, want %q", hookRoot, filepath.Clean(root))
		}
		data, err := os.ReadFile(filepath.Join(sections, defs.LSPYAML))
		if err != nil {
			t.Errorf("inside hook: lsp.yaml unreadable (the deploy has not run?): %v", err)
		}
		lspInHook = string(data)
		_, statErr := os.Stat(filepath.Join(sections, defs.ReportYAML))
		reportExistedInHook = statErr == nil
	}

	init := NewInitializer(sectionDeployer{}, manifest.NewManager(), nil)
	if _, err := init.Init(context.Background(), afterDeployOpts(root, hook)); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if calls != 1 {
		t.Fatalf("AfterTemplateDeploy calls = %d, want exactly 1", calls)
	}
	if lspInHook != templateLSPRender {
		t.Errorf("inside hook lsp.yaml = %q, want the deployed template render %q (a section patch ran before the hook)", lspInHook, templateLSPRender)
	}
	if reportExistedInHook {
		t.Errorf("inside hook report.yaml already existed; Step 3c ran before the hook")
	}

	// Positive controls: both section writers did run after the hook, so the
	// in-hook observations above are about ordering, not about writers that
	// never touch these files.
	after, err := os.ReadFile(filepath.Join(sections, defs.LSPYAML))
	if err != nil {
		t.Fatalf("read lsp.yaml after Init: %v", err)
	}
	if !strings.Contains(string(after), "enabled: true") {
		t.Errorf("after Init lsp.yaml = %q, want the wizard patch enabled: true", after)
	}
	if _, err := os.Stat(filepath.Join(sections, defs.ReportYAML)); err != nil {
		t.Errorf("after Init report.yaml missing: %v", err)
	}
}

// TestInit_AfterTemplateDeploySkippedOnFallback pins the other half of the
// contract: with no deployer there is no template render to record, and the
// hook must not fire.
func TestInit_AfterTemplateDeploySkippedOnFallback(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	calls := 0
	init := NewInitializer(nil, nil, nil)
	if _, err := init.Init(context.Background(), afterDeployOpts(root, func(string) { calls++ })); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if calls != 0 {
		t.Errorf("AfterTemplateDeploy calls on the no-deployer fallback = %d, want 0", calls)
	}
	// Positive control: the fallback path did write section files, so "no
	// call" is not "Init returned before reaching Step 3".
	if _, err := os.Stat(filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.LSPYAML)); err != nil {
		t.Errorf("fallback path wrote no lsp.yaml: %v", err)
	}
}

// TestInit_AfterTemplateDeploySkippedOnDeployFailure: a failed deploy leaves no
// render to record, and Init returns the deploy error without firing the hook.
func TestInit_AfterTemplateDeploySkippedOnDeployFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	calls := 0
	init := NewInitializer(&mockDeployer{deployErr: os.ErrPermission}, manifest.NewManager(), nil)
	if _, err := init.Init(context.Background(), afterDeployOpts(root, func(string) { calls++ })); err == nil {
		t.Fatalf("Init returned nil, want the deploy error")
	}
	if calls != 0 {
		t.Errorf("AfterTemplateDeploy calls after a failed deploy = %d, want 0", calls)
	}
}
