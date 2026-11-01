package project

// Card t2010 / PJ-14 — `moai init` validates every template before it writes
// any (all-or-nothing pre-validation, as `moai update` already does via
// ValidateAll). The invalid template sorts AFTER the valid one so that a
// deployer which renders and writes in file order leaves the valid file behind;
// a test whose invalid file sorts first passes by ordering alone.
//
// Scope of the guarantee: a template that fails to parse or render writes zero
// files. It is not a rollback — an I/O failure part-way through the write phase
// can still leave earlier files on disk.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

func orderSensitiveDeployer() template.Deployer {
	fsys := fstest.MapFS{
		"a-valid.tmpl":   &fstest.MapFile{Data: []byte("ok {{.Version}}")},
		"z-invalid.tmpl": &fstest.MapFile{Data: []byte("broken {{.Version")},
	}
	return template.NewDeployerWithRenderer(fsys, template.NewRenderer(fsys))
}

func TestInit_InvalidTemplateWritesNoTemplateFiles(t *testing.T) {
	root := t.TempDir()

	_, err := NewInitializer(orderSensitiveDeployer(), manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root))
	if err == nil {
		t.Fatal("Init() returned nil although one template cannot be rendered")
	}

	for _, name := range []string{"a-valid", "z-invalid"} {
		if _, statErr := os.Stat(filepath.Join(root, name)); statErr == nil {
			t.Errorf("Init() left %q behind after a template validation failure", name)
		}
	}
}

func TestInit_ValidTemplatesStillDeployAfterPreValidation(t *testing.T) {
	root := t.TempDir()
	fsys := fstest.MapFS{
		"a-valid.tmpl": &fstest.MapFile{Data: []byte("ok {{.Version}}")},
		"b-valid.tmpl": &fstest.MapFile{Data: []byte("ok too {{.Version}}")},
	}
	dep := template.NewDeployerWithRenderer(fsys, template.NewRenderer(fsys))

	if _, err := NewInitializer(dep, manifest.NewManager(), nil).
		Init(context.Background(), mirrorNoticeInitOptions(root)); err != nil {
		t.Fatalf("Init() with valid templates = %v", err)
	}
	for _, name := range []string{"a-valid", "b-valid"} {
		if _, statErr := os.Stat(filepath.Join(root, name)); statErr != nil {
			t.Errorf("valid template %q was not deployed: %v", name, statErr)
		}
	}
}
