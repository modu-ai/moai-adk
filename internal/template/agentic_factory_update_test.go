// agentic_factory_update_test.go: install and update behaviour of the Agentic
// Factory templates, asserted on the files a project ends up with.
//
// Scope. Fresh install (init mode, no force) and update (force-update mode)
// both drive the real Deployer over the real embedded files in t.TempDir()
// projects; nothing touches the network or a real repository. The assertions
// are on resulting files and manifest provenance only, so they do not depend
// on how the `moai update` command orchestrates the Deployer. The command-level
// update path (internal/cli/update.go) is not driven here and must be re-run
// after the PR that reworks it lands.
//
// User-file preservation under force-update is asserted by
// agentic_factory_update_preserve_test.go in this package and
// update_agentic_factory_preserve_test.go in internal/cli; the deployer
// protects Agentic Factory provider templates through isUserOwnedTemplatePath
// (user_owned_templates.go).
//
// Sentinels on failure: AF_INSTALL / AF_UPDATE.
package template_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

func afReadFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func afWriteFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestAgenticFactoryInstallThenUpdate: a new project receives every feature
// file byte-for-byte from the embedded FS, and an update over an older
// template-managed copy refreshes it to the shipped bytes.
func TestAgenticFactoryInstallThenUpdate(t *testing.T) {
	t.Parallel()
	sub := afFeatureSubset(t)

	// fresh install
	fresh := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(fresh); err != nil {
		t.Fatalf("AF_INSTALL: manifest load: %v", err)
	}
	if err := template.NewDeployer(sub).Deploy(context.Background(), fresh, mgr, nil); err != nil {
		t.Fatalf("AF_INSTALL: Deploy: %v", err)
	}
	for _, rel := range afExpectedFiles() {
		want, err := sub.ReadFile(rel)
		if err != nil {
			t.Fatalf("AF_INSTALL: read shipped %s: %v", rel, err)
		}
		if got := afReadFile(t, fresh, rel); got != string(want) {
			t.Errorf("AF_INSTALL: %s differs from the shipped bytes", rel)
		}
	}

	// existing project: an older template-managed copy of one file, then update
	old := t.TempDir()
	rel := afExpectedFiles()[0]
	afWriteFile(t, old, rel, "an older shipped revision\n")
	oldMgr := manifest.NewManager()
	if _, err := oldMgr.Load(old); err != nil {
		t.Fatalf("AF_UPDATE: manifest load: %v", err)
	}
	if err := oldMgr.Track(rel, manifest.TemplateManaged, manifest.HashBytes([]byte("an older shipped revision\n"))); err != nil {
		t.Fatalf("AF_UPDATE: track: %v", err)
	}
	if err := template.NewDeployerWithForceUpdate(sub, true).Deploy(context.Background(), old, oldMgr, nil); err != nil {
		t.Fatalf("AF_UPDATE: Deploy(force): %v", err)
	}
	want, _ := sub.ReadFile(rel)
	if got := afReadFile(t, old, rel); got != string(want) {
		t.Errorf("AF_UPDATE: %s was not refreshed to the shipped bytes", rel)
	}
}
