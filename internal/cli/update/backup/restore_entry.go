package backup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// BackupMarkerFile is the metadata file BackupMoaiConfig writes at the root of
// every backup directory it creates. Its presence is what makes a directory a
// recognisable moai update backup.
const BackupMarkerFile = "backup_metadata.json"

// ErrNotAnUpdateBackup is returned when a directory handed to
// RestoreFromBackupDir does not carry BackupMarkerFile.
var ErrNotAnUpdateBackup = errors.New("not a moai update backup directory")

// RestoreFromBackupDir applies a run-scoped update backup to a project tree.
//
// It EXTENDS RestoreMoaiConfig rather than duplicating it: the copy/merge
// semantics stay in the single existing owner, and this function adds only the
// preconditions the mid-run caller does not need — a marker-file check that
// refuses an arbitrary directory (REQ-UDS-024) instead of copying its contents
// into the project.
//
// It deliberately does NOT consult the project-marker gate. The restore entry
// point exists to repair a tree whose .moai/config/sections/system.yaml was
// destroyed, so requiring that marker would lock the user out of the only
// command that can restore it (REQ-UDS-022). The ordinary update path keeps the
// gate unchanged (REQ-UDS-025).
//
// Applying the same backup twice leaves the same tree state as applying it once
// (REQ-UDS-023); the entry point only writes, it never deletes.
//
// Idempotency needs one extra pass to hold at the byte level. RestoreMoaiConfig
// raw-copies a section whose target is ABSENT but re-serialises a section whose
// target EXISTS through the YAML merge, which canonicalises key order and
// indentation. A single pass over a wiped config directory therefore leaves
// raw bytes, and a second call would rewrite them in canonical form — the same
// tree state semantically, but not byte-identical. Running the merge to its
// fixed point here (the merge is stable from the second pass onward: it
// re-marshals an already-canonical document to itself) makes every invocation
// of this entry point leave the tree in the same state. The extra pass is
// scoped to the user-invoked restore; the mid-run callers of RestoreMoaiConfig
// keep their existing single-pass behaviour unchanged (NFR-UDS-005).
func RestoreFromBackupDir(projectRoot, backupDir string, recordFallback MergeFallbackRecorder) error {
	if projectRoot == "" {
		return errors.New("restore: empty project root")
	}
	if backupDir == "" {
		return errors.New("restore: empty backup directory")
	}

	info, err := os.Stat(backupDir)
	if err != nil {
		return fmt.Errorf("restore: stat backup directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("restore: %w: %s is not a directory", ErrNotAnUpdateBackup, backupDir)
	}

	marker := filepath.Join(backupDir, BackupMarkerFile)
	if _, err := os.Stat(marker); err != nil {
		return fmt.Errorf("restore: %w: %s missing in %s",
			ErrNotAnUpdateBackup, BackupMarkerFile, backupDir)
	}

	// restoreFixedPointPasses: the first pass applies the backup, the second
	// drives any raw-copied section through the merge so the tree lands on the
	// merge fixed point. See the idempotency note above.
	const restoreFixedPointPasses = 2
	for i := 0; i < restoreFixedPointPasses; i++ {
		if err := RestoreMoaiConfig(projectRoot, backupDir, recordFallback); err != nil {
			return fmt.Errorf("restore: apply backup %s: %w", backupDir, err)
		}
	}
	return nil
}
