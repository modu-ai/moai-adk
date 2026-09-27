package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// checkProjectMarker refuses a directory that is not a moai project.
//
// The positive marker is `.moai/config/sections/system.yaml`; when it is absent
// the directory is not a moai project and `moai update` MUST NOT write anything
// into it. Extracted from runUpdate so the gate is drivable by a test without
// mutating the process working directory.
//
// The error names the missing marker relative to the directory and directs the
// user to `moai init`. It deliberately does NOT echo the absolute path.
func checkProjectMarker(dir string) error {
	if isMoAIProject(dir) {
		return nil
	}
	marker := filepath.ToSlash(filepath.Join(defs.MoAIDir, "config", "sections", "system.yaml"))
	return fmt.Errorf(
		"not a moai project: %s not found in the current directory\n\n"+
			"Run `moai init` to initialize a project here, or change to an "+
			"existing project directory and retry", marker)
}

// runUpdateRestore is the user-invocable restore entry point: it applies a
// run-scoped update backup to the project tree.
//
// It is the lockout escape named in plan.md §B.3. Callers reach it BEFORE the
// project-marker gate, because the marker's absence is exactly the damage this
// entry point repairs (REQ-UDS-022). Scoping the exemption here — and nowhere
// else — keeps the gate in force for the ordinary update path (REQ-UDS-025).
//
// The backup directory is validated before anything is written: a directory
// without the backup marker file is refused rather than copied into the project
// (REQ-UDS-024). Applying the same backup twice is safe (REQ-UDS-023).
func runUpdateRestore(projectRoot, backupDir string, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}

	absBackup, err := filepath.Abs(backupDir)
	if err != nil {
		return fmt.Errorf("resolve backup directory %s: %w", backupDir, err)
	}

	if err := backup.RestoreFromBackupDir(projectRoot, absBackup,
		func(pr, relPath string, success bool, errOut io.Writer) {
			recordMergeFallback(pr, relPath, success, updateVerboseMode, errOut)
		}); err != nil {
		return err
	}

	// Card t1139: this path deliberately does NOT write the template snapshot.
	// The snapshot is the merge BASE and must record a template render; the
	// restored config holds the user's values, and recording those as BASE
	// makes the next update read every customization as "unchanged" and drop
	// it. No template is deployed here, so the snapshot the last deploy wrote
	// stays the most recent render and is left as it is.

	// Card t1276 F3: the restore rewrote .moai/config/sections/*.yaml from the
	// backup, while the failed run's deploy step may already have saved a
	// manifest carrying the NEW hashes — re-record the restored content so the
	// manifest matches the tree and the restored files are not frozen
	// user_modified on the next init --force.
	retrackSectionFiles(projectRoot, out)

	_, _ = fmt.Fprintf(out, "Restored .moai/config from %s\n", absBackup)
	return nil
}
