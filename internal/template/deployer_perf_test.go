package template

import (
	"context"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// countingRenderer wraps a Renderer and counts Render calls (REQ-PERF-006-A test double).
type countingRenderer struct {
	mu       sync.Mutex
	count    int
	failPath string // if set, Render returns error for this path
}

func (c *countingRenderer) Render(path string, ctx any) ([]byte, error) {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
	if path == c.failPath {
		return nil, &renderError{path: path}
	}
	return []byte("rendered:" + path), nil
}

func (c *countingRenderer) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

type renderError struct{ path string }

func (e *renderError) Error() string { return "render error: " + e.path }

// TestDeployerSingleRender (AC-PERF-006a) verifies that ValidateAll → Deploy
// on the same deployer instance renders each .tmpl exactly once (not twice).
func TestDeployerSingleRender(t *testing.T) {
	testFS := fstest.MapFS{
		"a.tmpl":    {Data: []byte("template a {{.GoBinPath}}")},
		"b.tmpl":    {Data: []byte("template b {{.HomeDir}}")},
		"c.tmpl":    {Data: []byte("template c")},
		"plain.txt": {Data: []byte("not a template")},
	}

	cr := &countingRenderer{}
	dep := NewDeployerWithRendererAndForceUpdate(testFS, cr, true)

	ctx := context.Background()
	tmplCtx := &TemplateContext{GoBinPath: "/test/bin", HomeDir: "/test/home"}

	// ValidateAll renders 3 templates and caches results
	if err := dep.ValidateAll(ctx, tmplCtx); err != nil {
		t.Fatalf("ValidateAll failed: %v", err)
	}
	validateCount := cr.Count()
	if validateCount != 3 {
		t.Errorf("ValidateAll should render 3 templates, got %d", validateCount)
	}

	// Deploy should reuse cached renders — 0 additional Render calls
	tmpDir := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(tmpDir); err != nil {
		t.Fatalf("manifest Load failed: %v", err)
	}
	if err := dep.Deploy(ctx, tmpDir, mgr, tmplCtx); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	totalCount := cr.Count()
	if totalCount != 3 {
		t.Errorf("ValidateAll+Deploy should render 3 total (cached reuse), got %d (baseline would be 6)", totalCount)
	}
}

// TestDeployerValidateErrorPreventsDeploy (AC-PERF-006b) verifies that when
// ValidateAll encounters a render error, it returns the error (pre-deploy gate).
func TestDeployerValidateErrorPreventsDeploy(t *testing.T) {
	testFS := fstest.MapFS{
		"good.tmpl": {Data: []byte("good")},
		"bad.tmpl":  {Data: []byte("bad")},
	}

	cr := &countingRenderer{failPath: "bad.tmpl"}
	dep := NewDeployerWithRendererAndForceUpdate(testFS, cr, true)

	ctx := context.Background()
	tmplCtx := &TemplateContext{GoBinPath: "/test/bin", HomeDir: "/test/home"}

	err := dep.ValidateAll(ctx, tmplCtx)
	if err == nil {
		t.Fatal("ValidateAll should fail for bad.tmpl")
	}
}
