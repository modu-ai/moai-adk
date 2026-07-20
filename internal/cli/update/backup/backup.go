// Package backup contains the config backup, restore, restore-rotation,
// path-containment guard, and YAML merge functions extracted from the root
// update command during M3d-A decomposition (SPEC-CLI-TUX-V3-003). Behavior is
// byte-identical to the pre-decomposition implementation; only the package
// location changed.
//
// The merge-history noise-suppression ledger (recordMergeFallback + the
// merge-history.json counter) stays in package cli and is invoked from
// RestoreMoaiConfig via a MergeFallbackRecorder callback seam, so that the
// ledger subsystem is not dragged across the package boundary.
package backup

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

func BackupMoaiConfig(projectRoot string) (string, error) {
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)

	// Check if config directory exists
	info, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No config to backup
		}
		return "", fmt.Errorf("stat config directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("config path is not a directory")
	}

	timestamp := time.Now().Format(defs.BackupTimestampFormat)
	backupDir := filepath.Join(projectRoot, defs.BackupsDir, timestamp)

	// Create backup directory
	if err := os.MkdirAll(backupDir, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	// All config files are backed up (including sections/) for full restore.
	// The clean step will delete everything, and the restore step will
	// merge backed-up values back into the freshly deployed templates.
	excludedDirs := []string{}

	// Track backed up items and excluded items for metadata
	backedUpItems := []string{}
	excludedItems := []string{}

	// Copy all files from config to backup, excluding sections directory
	err = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(configDir, path)
		if err != nil {
			return err
		}

		// Check for exclusion first - both directory and file level
		for _, excludedDir := range excludedDirs {
			if relPath == excludedDir || strings.HasPrefix(relPath, excludedDir+string(filepath.Separator)) {
				// Track excluded item
				excludedItems = append(excludedItems, relPath)
				// Skip this file or directory
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip directories that are not excluded
		if info.IsDir() {
			return nil
		}

		// Get relative path from backup directory
		// Use forward slashes for consistent metadata across platforms
		backupRelPath := filepath.ToSlash(filepath.Join(defs.MoAIDir, defs.ConfigSubdir, relPath))
		backedUpItems = append(backedUpItems, backupRelPath)

		backupPath := filepath.Join(backupDir, relPath)
		if err := os.MkdirAll(filepath.Dir(backupPath), defs.DirPerm); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(backupPath, data, defs.FilePerm)
	})

	if err != nil {
		_ = os.RemoveAll(backupDir)
		return "", fmt.Errorf("copy config files: %w", err)
	}

	// Save template defaults from embedded FS for 3-way merge.
	// This allows the restore step to distinguish user-modified values
	// from unchanged template defaults.
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults")
	if err := SaveTemplateDefaults(templateDefaultsDir); err != nil {
		// Non-fatal: if template defaults can't be saved, restore falls back to 2-way merge
		_, _ = fmt.Fprintf(os.Stderr, "Warning: could not save template defaults: %v\n", err)
	}

	// Create backup metadata file
	metadata := BackupMetadata{
		Timestamp:           timestamp,
		Description:         "config_backup",
		BackedUpItems:       backedUpItems,
		ExcludedItems:       excludedItems,
		ExcludedDirs:        excludedDirs,
		ProjectRoot:         projectRoot,
		BackupType:          "config",
		TemplateDefaultsDir: ".template-defaults",
	}

	metadataPath := filepath.Join(backupDir, "backup_metadata.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		_ = os.RemoveAll(backupDir)
		return "", fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, defs.FilePerm); err != nil {
		_ = os.RemoveAll(backupDir)
		return "", fmt.Errorf("write metadata: %w", err)
	}

	return backupDir, nil
}

// SaveTemplateDefaults extracts config section files from embedded templates
// and saves them to the given directory as baseline for 3-way merge.
func SaveTemplateDefaults(destDir string) error {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Walk embedded FS to find config section files
	prefix := ".moai/config/sections/"
	entries, err := fs.ReadDir(embedded, ".moai/config/sections")
	if err != nil {
		return fmt.Errorf("read embedded config sections: %w", err)
	}

	sectionsDestDir := filepath.Join(destDir, "sections")
	if err := os.MkdirAll(sectionsDestDir, defs.DirPerm); err != nil {
		return fmt.Errorf("create template defaults directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Read the raw file from embedded FS
		data, err := fs.ReadFile(embedded, prefix+name)
		if err != nil {
			continue // Skip files that can't be read
		}

		// For .tmpl files, save the raw template (not rendered) - the keys
		// and structure are what matter for 3-way comparison, template vars
		// will have placeholder values like {{.Version}} which won't match
		// user values, so they'll be treated as "user changed" = correct behavior.
		// Strip .tmpl extension for the output filename.
		outputName := strings.TrimSuffix(name, ".tmpl")
		if err := os.WriteFile(filepath.Join(sectionsDestDir, outputName), data, defs.FilePerm); err != nil {
			continue
		}
	}

	return nil
}

// BackupMetadata represents the structure of backup_metadata.json
type BackupMetadata struct {
	Timestamp           string   `json:"timestamp"`
	Description         string   `json:"description"`
	BackedUpItems       []string `json:"backed_up_items"`
	ExcludedItems       []string `json:"excluded_items"`
	ExcludedDirs        []string `json:"excluded_dirs"`
	ProjectRoot         string   `json:"project_root"`
	BackupType          string   `json:"backup_type"`
	TemplateDefaultsDir string   `json:"template_defaults_dir,omitempty"`
}

func CleanupOldBackups(projectRoot string, keepCount int) int {
	backupDir := filepath.Join(projectRoot, defs.BackupsDir)

	// Check if backup directory exists
	info, err := os.Stat(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0 // No backups to clean up
		}
		// Return 0 on stat errors (ignore for cleanup)
		return 0
	}
	if !info.IsDir() {
		return 0
	}

	// Get all subdirectories in backup directory
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return 0
	}

	// Filter for directories matching YYYYMMDD_HHMMSS pattern
	// Pattern: 8 digits + underscore + 6 digits = 15 characters
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 15 {
			// Check if it matches the timestamp pattern (digits + underscore + digits)
			parts := strings.SplitN(entry.Name(), "_", 2)
			if len(parts) == 2 {
				if len(parts[0]) == 8 && len(parts[1]) == 6 {
					backups = append(backups, entry.Name())
				}
			}
		}
	}

	// If we have fewer backups than keepCount, no cleanup needed
	if len(backups) <= keepCount {
		return 0
	}

	// Sort backups by name (timestamp) ascending (oldest first)
	sort.Strings(backups)

	// Delete backups exceeding the keep limit
	deletedCount := 0
	for _, backupName := range backups[keepCount:] {
		backupPath := filepath.Join(backupDir, backupName)
		if err := os.RemoveAll(backupPath); err != nil {
			// Log error but continue with other backups
			fmt.Fprintf(os.Stderr, "Warning: failed to delete backup %s: %v\n", backupName, err)
		} else {
			deletedCount++
		}
	}

	return deletedCount
}

// IsSymlinkEntry reports whether path is a symbolic link (via os.Lstat, which
// does NOT follow the link). C-F2 backup-entry symlink guard for
// SPEC-SEC-HARDEN-003 (REQ-SEC3-005/006). Fail-closed: a stat error other than
// not-exist is treated as a symlink (reject) so a racing replacement cannot
// slip a link past the guard; a clean not-exist returns false (regular walk).
func IsSymlinkEntry(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		// Not-exist is benign (nothing to follow); any other error → fail-closed.
		return !os.IsNotExist(err)
	}
	return info.Mode()&os.ModeSymlink != 0
}

// RestoreTargetContained reports whether targetPath is a safe write destination
// inside configDir. C-F2 traversal guard for SPEC-SEC-HARDEN-003 (REQ-SEC3-007):
// it rejects (1) a target whose cleaned relative path escapes configDir
// (filepath.Rel yields a `..` prefix or an outside-absolute path) and (2) a
// target that already exists as a symlink (which os.WriteFile would follow out
// of configDir). Cross-platform via filepath (NFR-SEC3-005); fail-closed on any
// resolution error (NFR-SEC3-004). configDir itself counts as contained.
//
// SPEC-SEC-HARDEN-004 F1 (REQ-SEC4-001/002/003): additionally resolves the
// targetPath PARENT chain with filepath.EvalSymlinks and re-checks that the
// resolved parent stays inside configDir. The leaf guards above only inspect the
// leaf, so a pre-existing symlinked intermediate directory
// (configDir/linkdir → /outside) would let os.MkdirAll/os.WriteFile escape
// configDir via the symlinked parent (CWE-22). This guard is shared by both the
// modern walk (restoreMoaiConfig) and the legacy walk (restoreMoaiConfigLegacy),
// so the single addition closes both paths at once.
//
// TOCTOU note (SPEC-SEC-HARDEN-005 §F.3, OPT-SEC5-001 — non-gating): containment
// is checked at decision time; a concurrent adversarial process could in
// principle race the check against the subsequent os.WriteFile/os.MkdirAll write
// (check-vs-use window). This TOCTOU window is out of scope under the offline
// single-process threat model — `moai update` is a single process on the user's
// own machine — per SEC-HARDEN-003/004 §F.1 precedent. No code-behavior change.
func RestoreTargetContained(configDir, targetPath string) bool {
	if configDir == "" || targetPath == "" {
		return false
	}
	absConfig, err := filepath.Abs(configDir)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absConfig, absTarget)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	// Reject an existing symlink at the target path: os.WriteFile follows it and
	// can escape configDir even when the literal path looks contained.
	if IsSymlinkEntry(absTarget) {
		return false
	}
	// SPEC-SEC-HARDEN-004 F1: parent-chain symlink containment.
	if !ParentChainContained(absConfig, absTarget) {
		return false
	}
	return true
}

// ParentChainContained reports whether the parent directory chain of absTarget,
// once its symbolic links are resolved, stays inside absConfig. F1 guard for
// SPEC-SEC-HARDEN-004 (REQ-SEC4-001). Both paths are absolute and cleaned.
//
// Strategy (REQ-SEC4-003 + run-phase watch-item 3 + SEC-HARDEN-004 deep-variant
// fast-follow): the configDir base is itself normalized with EvalSymlinks before
// the containment compare, so a legitimately symlinked configDir base does NOT
// yield a false ".." rejection.
//
// The parent chain is then resolved with a DEEPEST-EXISTING-ANCESTOR walk rather
// than a single EvalSymlinks(filepath.Dir): walk UP from filepath.Dir(absTarget)
// until a component that EXISTS is found, then EvalSymlinks that deepest-existing
// ancestor and require it to stay inside the normalized configDir.
//
// Why the walk (and why a single-shot EvalSymlinks(filepath.Dir) was wrong):
// when the leaf's immediate parent does not exist yet, EvalSymlinks fails with
// os.IsNotExist — but a symlink CAN still be in place at a SHALLOWER component.
// For a deep target like configDir/sections/linkdir/sub/evil.yaml where linkdir
// is an existing symlink to /outside but sub does not exist, the old "not-exist
// parent → no symlink can be in place → allow" reasoning was false: the restore
// loop's os.MkdirAll(filepath.Dir(target)) creates /outside/sub THROUGH linkdir
// and os.WriteFile escapes to /outside/sub/evil.yaml. The walk catches this by
// resolving linkdir (the deepest existing ancestor) and rejecting its escape.
//
// Resolution outcomes:
//   - deepest-existing-ancestor resolves inside configDir → allow. The
//     not-yet-existing components below it are created by os.MkdirAll as REAL
//     dirs (no symlink to follow), so a legit deep first-restore is permitted.
//   - deepest-existing-ancestor resolves OUTSIDE configDir (an existing ancestor
//     is a symlink escape) → reject. This is the F1 close, now covering the deep
//     variant.
//   - the walk reaches configDir itself (the floor) and configDir is contained →
//     allow (the entire parent chain is fresh and in-config).
//   - any EvalSymlinks error other than os.IsNotExist → fail-closed (reject); a
//     coarse "any error → allow" would RE-OPEN the hole.
//
// TOCTOU note (SPEC-SEC-HARDEN-005 §F.3, OPT-SEC5-001 — non-gating): the parent
// chain is resolved at decision time; a concurrent adversarial process could in
// principle swap an ancestor between this check and the subsequent write
// (check-vs-use window). This TOCTOU window is out of scope under the offline
// single-process threat model per SEC-HARDEN-003/004 §F.1 precedent. No
// code-behavior change.
func ParentChainContained(absConfig, absTarget string) bool {
	// Normalize the configDir base. EvalSymlinks requires the path to exist; if
	// it does not yet exist (or fails for a non-not-exist reason), fall back to
	// the unresolved absConfig — the leaf/lexical guards still apply, and a
	// not-yet-existing configDir cannot host a symlinked parent.
	normConfig := absConfig
	if resolved, err := filepath.EvalSymlinks(absConfig); err == nil {
		normConfig = resolved
	} else if !os.IsNotExist(err) {
		// configDir present but unresolvable for a non-not-exist reason →
		// fail-closed.
		return false
	}

	// Deepest-existing-ancestor walk: start at the leaf's parent and climb until
	// a component exists (EvalSymlinks succeeds). configDir is the FLOOR — never
	// inspect an ancestor at or above it. The caller (RestoreTargetContained) has
	// already established lexically that absTarget is inside absConfig, so the
	// parent chain from absTarget up to absConfig is in-config; only the part of
	// that chain BELOW configDir can host an attacker-planted symlink escape.
	// Climbing to-or-above configDir would compare a fresh, not-yet-created
	// in-config chain against the (possibly non-existent) configDir base and
	// produce a false ".." rejection — the floor check prevents that.
	ancestor := filepath.Dir(absTarget)
	for {
		// Floor check: if ancestor is no longer strictly below absConfig (it has
		// reached configDir itself or climbed above it), the chain below configDir
		// is entirely fresh / in-config — no symlink to follow → allow.
		floorRel, floorErr := filepath.Rel(absConfig, ancestor)
		if floorErr != nil {
			return false
		}
		if floorRel == "." || floorRel == ".." || strings.HasPrefix(floorRel, ".."+string(os.PathSeparator)) {
			// "." → ancestor == configDir; ".."/"../…" → above configDir.
			return true
		}

		resolved, err := filepath.EvalSymlinks(ancestor)
		if err == nil {
			// Found the deepest existing ancestor strictly below configDir.
			// Require its resolved form to stay inside the normalized configDir.
			rel, relErr := filepath.Rel(normConfig, resolved)
			if relErr != nil {
				return false
			}
			if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				return false
			}
			return true
		}
		if !os.IsNotExist(err) {
			// Any other resolution error → fail-closed.
			return false
		}

		// ancestor does not exist yet — climb one level toward configDir.
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			// Reached the filesystem root without finding an existing ancestor
			// (filepath.Dir of root returns root). No symlink can be in place on
			// a fully non-existent chain → allow.
			return true
		}
		ancestor = parent
	}
}
