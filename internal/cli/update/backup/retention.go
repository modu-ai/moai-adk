package backup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// DefaultPastVersionsKeep is the per-file budget of superseded versions used when no
// budget is passed. The config default (state.yaml backup_past_versions_keep) carries
// the same value.
const DefaultPastVersionsKeep = 3

// templateOwnedPrefixes are backup paths the unique-content guard does not protect:
// template renders and their pre-clean copies, which are rebuilt from the template,
// so their versions are not user content.
var templateOwnedPrefixes = []string{".template-defaults/", "pre-clean/"}

// Retention reports one retention pass over a backup root, entry by entry.
type Retention struct {
	// Removed names the excess entries the pass deleted, oldest first.
	Removed []string
	// Kept names the excess entries the pass left in place, oldest first. An entry
	// stays when it holds a superseded version within the per-file budget, content
	// the guard cannot classify, or when its removal failed.
	Kept []string
}

// PruneOldBackups applies the retention of CleanupOldBackups with an explicit
// per-file budget and reports the outcome entry by entry. A refused root reports
// nothing.
func PruneOldBackups(projectRoot string, keepCount, pastKeep int) Retention {
	return pruneOldBackups(projectRoot, keepCount, pastKeep)
}

// PruneNamespaceBackups is PruneOldBackups for the namespace backup root.
func PruneNamespaceBackups(projectRoot string, keepCount, pastKeep int) Retention {
	return pruneNamespaceBackups(projectRoot, keepCount, pastKeep)
}

// UnmanagedEntries counts the entries under .moai-backups whose names lack the
// timestamp shape the cap prunes, and their total size in bytes. The cap never
// removes them; the update reports them so a person can review them. A refused root
// reports nothing.
func UnmanagedEntries(projectRoot string) (int, int64) {
	backupDir := filepath.Join(projectRoot, defs.BackupsDir)
	if !backupRootContained(projectRoot, backupDir) {
		return 0, 0
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return 0, 0
	}
	count := 0
	var size int64
	for _, entry := range entries {
		if entry.IsDir() && isTimestampBackupName(entry.Name()) {
			continue
		}
		count++
		size += pathSize(filepath.Join(backupDir, entry.Name()))
	}
	return count, size
}

// pathSize returns the total size of the regular files under path.
func pathSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		if info, infoErr := d.Info(); infoErr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// pruneOldBackups is the retention of CleanupOldBackups with an explicit per-file
// budget: the newest keepCount timestamped entries are retained and the excess is
// pruned by pruneBackups.
func pruneOldBackups(projectRoot string, keepCount, pastKeep int) Retention {
	backupDir := filepath.Join(projectRoot, defs.BackupsDir)
	if !backupRootContained(projectRoot, backupDir) {
		return Retention{}
	}
	info, err := os.Stat(backupDir)
	if err != nil || !info.IsDir() {
		return Retention{}
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return Retention{}
	}
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && isTimestampBackupName(entry.Name()) {
			backups = append(backups, entry.Name())
		}
	}
	if len(backups) <= keepCount {
		return Retention{}
	}
	sort.Strings(backups)
	return pruneBackups(backupDir, backups, keepCount, pastKeep)
}

// pruneNamespaceBackups is the retention of CleanupNamespaceBackups with an explicit
// per-file budget.
func pruneNamespaceBackups(projectRoot string, keepCount, pastKeep int) Retention {
	root := filepath.Join(projectRoot, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	if !backupRootContained(projectRoot, root) {
		return Retention{}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return Retention{}
	}
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "update-") {
			backups = append(backups, entry.Name())
		}
	}
	if len(backups) <= keepCount {
		return Retention{}
	}
	sort.Strings(backups)
	return pruneBackups(root, backups, keepCount, pastKeep)
}

// pruneBackups removes the oldest excess entries of sorted and keeps the newest
// keepCount under backupDir. keepCount 0 is an explicit wipe and removes every entry.
// Otherwise an excess entry stays only when it holds content the guard protects (see
// entryMustStay). Each outcome is reported by name, oldest first.
func pruneBackups(backupDir string, sorted []string, keepCount, pastKeep int) Retention {
	var res Retention
	if keepCount == 0 {
		for _, name := range sorted {
			if err := os.RemoveAll(filepath.Join(backupDir, name)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete backup %s: %v\n", name, err)
				res.Kept = append(res.Kept, name)
				continue
			}
			res.Removed = append(res.Removed, name)
		}
		return res
	}

	retained := make(map[string]bool)
	for _, name := range sorted[len(sorted)-keepCount:] {
		addRetainedBackupKeys(filepath.Join(backupDir, name), retained)
	}
	excess := sorted[:len(sorted)-keepCount]
	budget := withinBudgetKeys(backupDir, excess, pastKeep)
	for _, name := range excess {
		path := filepath.Join(backupDir, name)
		if entryMustStay(path, retained, budget) {
			res.Kept = append(res.Kept, name)
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete backup %s: %v\n", name, err)
			res.Kept = append(res.Kept, name)
			continue
		}
		res.Removed = append(res.Removed, name)
	}
	return res
}

// withinBudgetKeys returns the backupFileKey of every superseded version the per-file
// budget keeps. For each user path, the excess entries are the only holders of its
// superseded versions; the pastKeep most recent distinct versions of that path are
// kept, where a version's recency is the newest excess entry that holds it.
func withinBudgetKeys(backupDir string, excess []string, pastKeep int) map[string]bool {
	newest := make(map[string]int)
	for i, name := range excess {
		entry := filepath.Join(backupDir, name)
		_ = filepath.WalkDir(entry, func(path string, d fs.DirEntry, err error) error {
			if err != nil || !d.Type().IsRegular() || isBackupBookkeeping(entry, path) || isTemplateOwned(entry, path) {
				return nil
			}
			key, keyErr := backupFileKey(entry, path)
			if keyErr != nil {
				return nil
			}
			newest[key] = i // excess is oldest first, so the last index is the newest
			return nil
		})
	}

	versionsOf := make(map[string][]string)
	for key := range newest {
		userPath, _, _ := strings.Cut(key, "\x00")
		versionsOf[userPath] = append(versionsOf[userPath], key)
	}
	keep := make(map[string]bool)
	for _, keys := range versionsOf {
		sort.Slice(keys, func(a, b int) bool { return newest[keys[a]] > newest[keys[b]] })
		for i := 0; i < len(keys) && i < pastKeep; i++ {
			keep[keys[i]] = true
		}
	}
	return keep
}

// entryMustStay reports whether an excess entry holds content the cap may not remove:
// a regular file that no retained entry has at its path and whose version is within
// the per-file budget, or any entry the guard cannot read or classify. A redundant
// file, a file beyond the budget, and a template-owned file do not hold it.
func entryMustStay(entry string, retained, budget map[string]bool) bool {
	stay := false
	_ = filepath.WalkDir(entry, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			stay = true
		case d.IsDir():
		case !d.Type().IsRegular():
			stay = true
		case isBackupBookkeeping(entry, path), isTemplateOwned(entry, path):
		default:
			key, keyErr := backupFileKey(entry, path)
			if keyErr != nil || (!retained[key] && budget[key]) {
				stay = true
			}
		}
		if stay {
			return fs.SkipAll
		}
		return nil
	})
	return stay
}

// isTemplateOwned reports whether path, a file under entry, lies in a template-owned
// prefix of the backup.
func isTemplateOwned(entry, path string) bool {
	rel, err := filepath.Rel(entry, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	for _, prefix := range templateOwnedPrefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}
