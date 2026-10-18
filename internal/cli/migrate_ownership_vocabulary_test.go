package cli

// migrate_ownership_vocabulary_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865): characterization of the closed vocabularies and of the small
// outcomes the larger criteria only touch in passing. A remaining-file list is
// only worth tallying by reason if the reason set stays closed.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestKeepReasonVocabularyIsClosed pins the closed reason and class sets:
// every defined reason is known, a parameterised reason is known only for a real
// evidence item, and anything else is not.
func TestKeepReasonVocabularyIsClosed(t *testing.T) {
	known := []keepReason{
		keepUserModified, keepUserCreated, keepNoProvenanceRecord, keepRecordHashMismatch,
		keepProfileReach, keepCounterpartUnconfirmed, keepNonCatalog, keepRemovalGateOff,
		keepSymlinkEscape, keepReadFailed, keepNoTemplateHash, keepEmptyDirNoProvenance, keepBackupFailed,
		// Added deliberately by the second card-review fix: a path that is not a regular
		// file is kept without being read.
		keepReason("non-regular-file"),
	}
	for _, item := range allEvidenceItems {
		known = append(known, evidenceIncompleteReason(item))
	}
	for _, r := range known {
		if !isKnownKeepReason(r) {
			t.Errorf("defined reason %q is not in the closed set", r)
		}
	}
	for _, r := range []keepReason{"", "something-new", evidenceIncompletePrefix, evidenceIncompletePrefix + "made-up-item", "Removal-Gate-Off"} {
		if isKnownKeepReason(r) {
			t.Errorf("reason %q must not be in the closed set", r)
		}
	}
	for _, c := range []ownershipClass{ownershipManaged, ownershipUserModified, ownershipUserCreated, ownershipUnregistered} {
		if !c.valid() {
			t.Errorf("class %q should be valid", c)
		}
	}
	for _, c := range []ownershipClass{"", "owned", "template-managed", "Managed"} {
		if c.valid() {
			t.Errorf("class %q must not be valid", c)
		}
	}
}

// TestPreservePriorManifestOutcomes covers the four outcomes of keeping the
// start-of-run manifest: nothing to keep, kept, the writer failing, and the run
// directory itself being impossible to create.
func TestPreservePriorManifestOutcomes(t *testing.T) {
	data := []byte(`{"version":"v","deployed_at":"x","files":{}}` + "\n")

	t.Run("no manifest at load time is not a failure", func(t *testing.T) {
		project := t.TempDir()
		runDir, state, err := preservePriorManifest(project, "", nil)
		if state != priorManifestAbsent || err != nil || runDir != "" {
			t.Errorf("got (%q, %v, %v), want an absent outcome without an error", runDir, state, err)
		}
		if entries, _ := os.ReadDir(filepath.Join(project, ".moai-backups")); len(entries) != 0 {
			t.Errorf("a run directory was created although there was nothing to keep")
		}
	})

	t.Run("kept in the run's own backup directory", func(t *testing.T) {
		project := t.TempDir()
		own := filepath.Join(project, ".moai-backups", "20260101_000001")
		runDir, state, err := preservePriorManifest(project, own, data)
		if state != priorManifestWritten || err != nil || runDir != own {
			t.Fatalf("got (%q, %v, %v)", runDir, state, err)
		}
		got, readErr := os.ReadFile(filepath.Join(own, priorManifestName))
		if readErr != nil || string(got) != string(data) {
			t.Errorf("the kept copy differs from the input: %v", readErr)
		}
	})

	t.Run("a fresh run directory when the config backup made none", func(t *testing.T) {
		project := t.TempDir()
		runDir, state, err := preservePriorManifest(project, "", data)
		if state != priorManifestWritten || err != nil || runDir == "" {
			t.Fatalf("got (%q, %v, %v)", runDir, state, err)
		}
		if _, statErr := os.Stat(filepath.Join(runDir, priorManifestName)); statErr != nil {
			t.Errorf("no copy in the fresh run directory: %v", statErr)
		}
	})

	t.Run("a failing writer is reported as failed", func(t *testing.T) {
		project := t.TempDir()
		boom := errors.New("injected")
		swapPriorManifestWriter(t, func(string, []byte) error { return boom })
		_, state, err := preservePriorManifest(project, filepath.Join(project, ".moai-backups", "20260101_000001"), data)
		if state != priorManifestFailed || !errors.Is(err, boom) {
			t.Errorf("got (%v, %v), want failed wrapping the injected error", state, err)
		}
	})

	t.Run("an impossible run directory is reported as failed", func(t *testing.T) {
		project := t.TempDir()
		blocker := filepath.Join(project, "blocker")
		if err := os.WriteFile(blocker, []byte("a file where a directory must go"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, state, err := preservePriorManifest(project, filepath.Join(blocker, "run"), data)
		if state != priorManifestFailed || err == nil {
			t.Errorf("got (%v, %v), want failed with an error", state, err)
		}
	})

	t.Run("the temporary write fails cleanly in a missing directory", func(t *testing.T) {
		tmp, err := writePriorManifestTemp(filepath.Join(t.TempDir(), "missing"), data)
		if err == nil || tmp != "" {
			t.Errorf("got (%q, %v), want an error and no temporary path", tmp, err)
		}
	})
}
