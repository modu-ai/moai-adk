package backup

// run_retention.go — the runs that backup retention counts (SPEC-UPDATE-REPAIR-
// RC30-001 D9, REQ-RC30-010). A run is the set of backup directories that carry
// one run id in their .moai-run marker, across both backup roots. A directory
// without a readable marker is a run of its own, but only when its name has a
// shape a moai writer produces under its root.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// namespaceBackupPattern matches a namespace backup a moai writer creates under
// .moai/backups/: update-<stamp> or v2-to-v3-<stamp>, where the stamp is
// YYYY-MM-DDTHH-MM-SSZ and an optional -N suffix marks a same-second collision.
var namespaceBackupPattern = regexp.MustCompile(`^(update|v2-to-v3)-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z(-\d+)?$`)

// backupRun is one run in the retention window: the backup directories that
// carry its run id, or the one unmarked directory it is made of.
type backupRun struct {
	// when is the run id's time, or the modification time of an unmarked directory.
	when time.Time
	// dirs holds the run's directory paths in ascending order.
	dirs []string
}

// newerThan orders runs newest first. Equal times break by the first path, so
// the order is deterministic.
func (r backupRun) newerThan(other backupRun) bool {
	if !r.when.Equal(other.when) {
		return r.when.After(other.when)
	}
	return r.dirs[0] < other.dirs[0]
}

// backupRoot is a directory that holds backup runs, with the name shape a moai
// writer gives an unmarked directory under it.
type backupRoot struct {
	dir   string
	shape func(name string) bool
}

// backupRoots returns the two backup roots of projectRoot: the config root and
// the namespace root.
func backupRoots(projectRoot string) []backupRoot {
	return []backupRoot{
		{dir: filepath.Join(projectRoot, defs.BackupsDir), shape: isConfigBackupName},
		{dir: filepath.Join(projectRoot, defs.MoAIDir, defs.NamespaceBackupsSubdir), shape: isNamespaceBackupName},
	}
}

// collectBackupRuns groups the backup directories of both roots into runs. A
// directory with a readable marker joins the run of its run id. An unmarked
// directory whose name has its root's shape is a run of its own. Any other
// unmarked directory is left out, so it is neither counted nor deleted.
func collectBackupRuns(projectRoot string) []backupRun {
	var runs []backupRun
	runIndex := make(map[string]int) // run id -> index into runs
	for _, root := range backupRoots(projectRoot) {
		entries, err := os.ReadDir(root.dir)
		if err != nil {
			continue // a missing or unreadable root holds no runs to prune
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dir := filepath.Join(root.dir, entry.Name())
			if id, when, ok := readRunMarker(dir); ok {
				if i, seen := runIndex[id]; seen {
					runs[i].dirs = append(runs[i].dirs, dir)
					continue
				}
				runIndex[id] = len(runs)
				runs = append(runs, backupRun{when: when, dirs: []string{dir}})
				continue
			}
			if !root.shape(entry.Name()) {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			runs = append(runs, backupRun{when: info.ModTime(), dirs: []string{dir}})
		}
	}
	for i := range runs {
		sort.Strings(runs[i].dirs)
	}
	return runs
}

// isConfigBackupName reports whether name has the shape BackupMoaiConfig gives a
// config backup: YYYYMMDD_HHMMSS, 15 characters with an underscore between an
// 8-character and a 6-character part.
func isConfigBackupName(name string) bool {
	parts := strings.SplitN(name, "_", 2)
	return len(name) == 15 && len(parts) == 2 && len(parts[0]) == 8 && len(parts[1]) == 6
}

// isNamespaceBackupName reports whether name has the shape of a namespace backup
// a moai writer creates under .moai/backups/.
func isNamespaceBackupName(name string) bool {
	return namespaceBackupPattern.MatchString(name)
}
