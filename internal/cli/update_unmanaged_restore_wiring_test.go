package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// F37 (iter6): the template-sync path must not restore an evaluator profile that the
// same run archived and removed. This drives the real forced template-sync flow
// (backup, reconcile, deploy, restore) over an initialized project, so removing the
// production skip wiring in update_template_sync.go fails this test. The retired
// profile is tracked as template-managed in the manifest and absent from the
// template, which makes the reconciliation archive and remove it.
func TestTemplateSync_UnmanagedRestoreKeepsArchivedRemovedProfile(t *testing.T) {
	projectDir := initIdentityProject(t)
	const rel = ".moai/config/evaluator-profiles/retired.md"
	profile := filepath.Join(projectDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(profile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profile, []byte("# profile retired by the current template\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectDir); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track retired profile: %v", err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	runForcedTemplateSyncAt(t, projectDir)

	if _, err := os.Stat(profile); !os.IsNotExist(err) {
		t.Errorf("archived-removed profile %s came back after the template sync (stat err=%v)", rel, err)
	}
}
