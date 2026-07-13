package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// MergeFallbackRecorder records a 3-way merge fallback event for a given
// relPath. The production implementation (recordMergeFallback in
// internal/cli/update_noise.go) updates the per-project merge-history ledger
// and emits the noise-suppression advisory or legacy warning per REQ-UN-007 /
// REQ-UN-008 / REQ-UN-010. success=true means the 3-way merge succeeded and
// the counter should be reset; success=false means it failed and the caller is
// about to fall back to 2-way. errOut MUST NOT be nil in production (os.Stderr).
//
// RestoreMoaiConfig accepts this as a callback (nil-safe) so that the
// noise-suppression ledger subsystem can remain in package cli and is not
// dragged into the backup subpackage by the restore-path move. The root
// update command wires the real recorder; tests pass nil to skip recording.
type MergeFallbackRecorder func(projectRoot, relPath string, success bool, errOut io.Writer)

// RestoreMoaiConfig restores the user's backed-up configuration sections into
// the config directory, preferring a 3-way merge (user + template + template
// defaults) and falling back to a 2-way merge when base data is unavailable.
//
// recordFallback (nil-safe) is invoked for each 3-way merge success/failure so
// the caller's merge-history ledger stays consistent. Pass nil to skip
// recording (test path).
//
// Moved from internal/cli/update.go during M3d-A decomposition
// (SPEC-CLI-TUX-V3-003). Behavior is byte-identical to the pre-decomposition
// implementation; only the package location and the recordFallback callback
// seam (formerly a direct recordMergeFallback call reading the package-level
// updateVerboseMode flag) changed.
func RestoreMoaiConfig(projectRoot, backupDir string, recordFallback MergeFallbackRecorder) error {
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults")

	// Check if template defaults are available for 3-way merge
	has3Way := false
	if info, err := os.Stat(templateDefaultsDir); err == nil && info.IsDir() {
		has3Way = true
	}

	// Walk through backup files (only sections/*.yaml)
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	if info, err := os.Stat(sectionsBackupDir); err != nil || !info.IsDir() {
		// No sections in backup, try walking from backup root
		return RestoreMoaiConfigLegacy(projectRoot, backupDir, configDir)
	}

	return filepath.Walk(sectionsBackupDir, func(backupPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// SPEC-SEC-HARDEN-003 C-F2 봉쇄 (in-scope sibling, REQ-SEC3-006): 심볼릭 링크
		// 백업 엔트리는 os.ReadFile로 따라가지 않고 스킵한다(CWE-61 fail-closed).
		if IsSymlinkEntry(backupPath) {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping symlink backup entry %s (containment)\n", backupPath)
			return nil
		}

		// Skip non-YAML files (e.g., backup_metadata.json)
		if filepath.Ext(backupPath) != ".yaml" && filepath.Ext(backupPath) != ".yml" {
			return nil
		}

		relPath, err := filepath.Rel(sectionsBackupDir, backupPath)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(configDir, "sections", relPath)

		// SPEC-SEC-HARDEN-003 C-F2 봉쇄 (REQ-SEC3-007): 쓰기 대상이 configDir를
		// 탈출(filepath.Rel `..`)하거나 심볼릭 링크로 configDir 밖을 가리키면 거부한다.
		if !RestoreTargetContained(configDir, targetPath) {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping restore target %s escaping config dir (containment)\n", targetPath)
			return nil
		}

		// Read backup (old user) data
		oldData, err := os.ReadFile(backupPath)
		if err != nil {
			return err
		}

		// Check if target file exists (new template)
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				// User's custom config section not in new template - restore as-is
				destDir := filepath.Dir(targetPath)
				if mkErr := os.MkdirAll(destDir, defs.DirPerm); mkErr != nil {
					return mkErr
				}
				return os.WriteFile(targetPath, oldData, defs.FilePerm)
			}
			return err
		}

		// Read new template data
		newData, err := os.ReadFile(targetPath)
		if err != nil {
			return err
		}

		// Try 3-way merge if template defaults are available
		if has3Way {
			basePath := filepath.Join(templateDefaultsDir, "sections", relPath)
			baseData, err := os.ReadFile(basePath)
			if err == nil {
				merged, mergeErr := MergeYAML3Way(newData, oldData, baseData)
				if mergeErr == nil {
					// REQ-UN-009: reset the merge-history counter on success so the
					// next failure starts a fresh 3-strike count.
					if recordFallback != nil {
						recordFallback(projectRoot, relPath, true, os.Stderr)
					}
					return os.WriteFile(targetPath, merged, defs.FilePerm)
				}
				// 3-way merge failed, fall through to 2-way.
				// REQ-UN-007/008/010: emit advisory or legacy warning per noise-suppression policy.
				if recordFallback != nil {
					recordFallback(projectRoot, relPath, false, os.Stderr)
				}
			}
		}

		// Fallback: 2-way merge (old behavior)
		merged, err := MergeYAMLDeep(newData, oldData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
			return os.WriteFile(targetPath, oldData, defs.FilePerm)
		}

		return os.WriteFile(targetPath, merged, defs.FilePerm)
	})
}

// RestoreMoaiConfigLegacy handles restore from legacy backup format
// (pre-3-way merge) where files might be at the backup root level.
func RestoreMoaiConfigLegacy(projectRoot, backupDir, configDir string) error {
	return filepath.Walk(backupDir, func(backupPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// SPEC-SEC-HARDEN-003 C-F2 봉쇄 (REQ-SEC3-005): 심볼릭 링크 백업 엔트리는
		// os.ReadFile로 따라가지 않고 스킵한다(CWE-61 fail-closed).
		if IsSymlinkEntry(backupPath) {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping symlink backup entry %s (containment)\n", backupPath)
			return nil
		}

		relPath, err := filepath.Rel(backupDir, backupPath)
		if err != nil {
			return err
		}

		// Skip metadata and template defaults
		if filepath.Base(relPath) == "backup_metadata.json" ||
			strings.HasPrefix(relPath, ".template-defaults") {
			return nil
		}

		targetPath := filepath.Join(configDir, relPath)

		// SPEC-SEC-HARDEN-003 C-F2 봉쇄 (REQ-SEC3-007): 쓰기 대상이 configDir를
		// 탈출(filepath.Rel `..`)하거나 심볼릭 링크로 configDir 밖을 가리키면 거부한다.
		if !RestoreTargetContained(configDir, targetPath) {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping restore target %s escaping config dir (containment)\n", targetPath)
			return nil
		}

		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			return err
		}

		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(targetPath), defs.DirPerm); err != nil {
					return fmt.Errorf("create parent directory for %s: %w", relPath, err)
				}
				return os.WriteFile(targetPath, backupData, defs.FilePerm)
			}
			return err
		}

		targetData, err := os.ReadFile(targetPath)
		if err != nil {
			return err
		}

		merged, err := MergeYAMLDeep(targetData, backupData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
			return os.WriteFile(targetPath, backupData, defs.FilePerm)
		}

		return os.WriteFile(targetPath, merged, defs.FilePerm)
	})
}
