package cli

// init_resume_checkpoint_order_test.go — card t1619, codex review P2 (init.go, the
// pre-executor user-asset ensure). The resume route of an interrupted init must be decided
// before the ensure moves any state: the ensure reconciles and clears the pending-install
// journal, and it fails on a corrupt user manifest. A checkpoint read after it would miss
// the interruption it exists to detect, and the resume's own quarantine and journal
// recovery would never be reached.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestInitResumeJournalSurvivesIntoResume: the pending-install journal is the only
// interruption evidence left (the init marker is retired), on a deployed project whose
// post-steps did not finish. The resume must read the journal as its checkpoint.
func TestInitResumeJournalSurvivesIntoResume(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "journal-resume")
	resumeSeedRun(t, projectDir, "claude")

	clearInitPending(projectDir)
	journal := userassets.JournalPath(userHomeDirOrEmpty())
	if err := userassets.WriteJournal(journal, &userassets.PendingJournal{SchemaVersion: userassets.SchemaVersion}); err != nil {
		t.Fatalf("stage the pending-install journal: %v", err)
	}

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("resume with the pending-install journal as its only checkpoint failed: %v", err)
	}
}

// TestInitResumeCorruptManifestReachesQuarantine: a corrupt user manifest on a deployed,
// interrupted project. The resume route quarantines the manifest, renaming it aside and never
// deleting it, and installs from scratch. The pre-executor ensure must not fail on the
// corrupt manifest before the resume gets there.
func TestInitResumeCorruptManifestReachesQuarantine(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "corrupt-manifest-resume")
	resumeSeedRun(t, projectDir, "claude")

	manifest := userassets.ManifestPath(userHomeDirOrEmpty())
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte("{not a manifest"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("resume over a corrupt user manifest failed before reaching the quarantine: %v", err)
	}
	quarantined, err := filepath.Glob(manifest + ".corrupt-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(quarantined) == 0 {
		t.Fatal("the resume completed without quarantining the corrupt manifest: the quarantine path was never reached")
	}
}
