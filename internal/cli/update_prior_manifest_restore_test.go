package cli

// update_prior_manifest_restore_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, card-review finding 1), flow level: a project whose pre-existing config
// has no sections/ directory makes the update's Restore step walk the run backup
// root. manifest.prior.json lives at that root as backup bookkeeping and must not
// be copied into .moai/config.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateDoesNotRestorePriorManifestIntoConfig(t *testing.T) {
	ownProfileSandbox(t)
	root := buildMigrationFixture(t)
	// A config tree without sections/: the backup will have none either, so the
	// restore falls through to the walk over the backup root.
	if err := os.RemoveAll(filepath.Join(root, ".moai", "config", "sections")); err != nil {
		t.Fatal(err)
	}
	const noteRel = ".moai/config/user-note.yaml"
	ownWriteFile(t, root, noteRel, []byte("note: kept\n"))

	run := runUpdateFlowForced(t, root)
	// Reach: the Backup stage wrote its copy, and the legacy restore ran (it is what
	// brings the config file back).
	if run.NewPriorManifest == "" {
		t.Fatalf("the run left no new %s: the Backup stage was not reached", priorManifestName)
	}
	if got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(noteRel))); err != nil || string(got) != "note: kept\n" {
		t.Fatalf("control: the config file was not restored, the Restore step did not walk the backup root (err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "config", "manifest.prior.json")); !os.IsNotExist(err) {
		t.Errorf("manifest.prior.json was copied into .moai/config by the update (stat err: %v)", err)
	}
}
