package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/tui"
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
//
// The live .claude/settings.json is never written by this path (SPEC-UPDATE-REPAIR-RC30-001
// D9). The restore names it as not restored and prints the manual copy
// instruction (reportUncoveredRestorePaths).
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
	// user_modified on the next init --force. Card t1527 repair round 3: the
	// failure returns; the ! line is this path's surface.
	if retrackErr := retrackSectionFiles(projectRoot, out); retrackErr != nil {
		emitSeverityLine(out, sevWarn, resolveTheme(), "manifest retrack (config sections) failed: %v", retrackErr)
	}

	// M2: evaluator-profiles/ and astgrep-rules/ restore beside the section merge;
	// the explicit restore reports each file it acted on.
	unmanagedRefs, err := backup.RestoreUnmanagedConfigDirs(projectRoot, absBackup)
	if err != nil {
		return err
	}
	for _, msg := range unmanagedRestoreMessages(unmanagedRefs, absBackup) {
		_, _ = fmt.Fprintln(out, msg)
	}
	_, _ = fmt.Fprintf(out, "Restored .moai/config from %s\n", absBackup)
	reportUncoveredRestorePaths(out, projectRoot, absBackup)
	return nil
}

// restoreSettingsRel is the live settings file this restore never writes.
// SPEC-UPDATE-REPAIR-RC30-001 D9 (operator verdict d-20261010T053033Z-7012): the
// restore names it as not restored and leaves the copy to the user.
const restoreSettingsRel = ".claude/settings.json"

// reportUncoveredRestorePaths prints what this restore leaves alone, so the user
// can finish the recovery by hand. It writes no file.
//
// The live .claude/settings.json is always named as not restored. When the run
// backup holds a copy, the line also carries the manual copy instruction with
// the backup path and the target path. The other run-backup copies outside the
// config tree are listed as backed up but not restored; inMemoryOnlyBackupTargets
// is their single source of truth.
func reportUncoveredRestorePaths(out io.Writer, projectRoot, absBackup string) {
	// Read after the restore, so a restored language.yaml sets the language.
	lang := config.LoadUpdateRenderValues(projectRoot).ConversationLanguage
	if lang == "" {
		lang = "en"
	}
	copyRoot := filepath.Join(absBackup, diskBackupSubdir)

	_, _ = fmt.Fprintln(out, fmt.Sprintf(tui.Translate("update.restore.settings_not_restored", lang), restoreSettingsRel))
	settingsCopy := filepath.Join(copyRoot, filepath.FromSlash(restoreSettingsRel))
	if _, err := os.Stat(settingsCopy); err == nil {
		target := filepath.Join(projectRoot, filepath.FromSlash(restoreSettingsRel))
		_, _ = fmt.Fprintln(out, fmt.Sprintf(tui.Translate("update.restore.copy_by_hand", lang), settingsCopy, target))
	}

	var items []string
	for _, rel := range inMemoryOnlyBackupTargets() {
		// The run marker is backup bookkeeping, not a backed-up user file: it is
		// never listed as uncovered.
		if rel == restoreSettingsRel || rel == backup.RunMarkerName {
			continue
		}
		copyPath := filepath.Join(copyRoot, filepath.FromSlash(rel))
		if _, err := os.Stat(copyPath); err != nil {
			continue
		}
		items = append(items, fmt.Sprintf(tui.Translate("update.restore.not_covered_item", lang), rel, copyPath))
	}
	if len(items) == 0 {
		return
	}
	_, _ = fmt.Fprintln(out, tui.Translate("update.restore.not_covered_heading", lang))
	for _, item := range items {
		_, _ = fmt.Fprintln(out, item)
	}
}
