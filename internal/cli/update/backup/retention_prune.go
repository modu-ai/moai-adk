package backup

// retention_prune.go — the pruner of SPEC-UPDATE-BACKUP-RETENTION-001 (design D1):
// containment check, run collection, census, selection, removal, report. The
// selection itself lives in retention_select.go and takes no filesystem handle;
// this file owns the two filesystem steps around it, the census and the removal.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// PruneItem is the report row of one excess run a prune decided on: the run's
// directories, what happened to it, and, for a retained run, why.
type PruneItem struct {
	// Run is the path of the run's first directory, the name the report uses.
	Run string
	// Dirs are the run's directories across both backup roots, ascending.
	Dirs []string
	// Action is RetentionActionRemoved or RetentionActionRetained.
	Action string
	// Reason is one of the RetentionReason constants for a retained run.
	Reason string
	// Protects counts the protected versions a run retained for them holds.
	Protects int
	// Detail names the unclassifiable entry, or the removal error.
	Detail string
}

// UnmanagedEntries summarizes the entries under the config backup root whose
// names no backup writer produces (REQ-UBR-012). They are reported, never deleted.
type UnmanagedEntries struct {
	Count int
	Bytes int64
}

// PruneReport is the outcome of one PruneBackups call.
type PruneReport struct {
	// Runs is the number of runs the pruner counted across both roots.
	Runs int
	// Window is the number of runs kept whole as the retention window.
	Window int
	// Items has one row per excess run, newest first.
	Items []PruneItem
	// Removed counts the backup directories the prune removed.
	Removed int
	// Protected is the number of distinct protected versions (REQ-UBR-010).
	Protected int
	// RootRefused is true when a backup root is a symbolic link or leads outside
	// the project; nothing was pruned (REQ-UBR-013).
	RootRefused bool
	// Unmanaged summarizes the entries no backup writer names.
	Unmanaged UnmanagedEntries
}

// PruneBackups prunes old backup runs and reports every decision. Retention counts
// runs across the config root (.moai-backups/) and the namespace root
// (.moai/backups/): the newest keepCount runs are the window and are kept whole;
// among the older runs it keeps the ones the selection step requires (a version
// no window run holds, within the per-file budget pastKeep) and removes the rest
// whole. A pastKeep below 1 reads as the shipped default. keepCount 0 removes
// every run and a negative keepCount removes nothing.
//
// A backup root that is a symbolic link or leads outside projectRoot is refused:
// nothing is removed from either root and the report says so. Entries under
// .moai-backups that no backup writer names are counted, never removed.
//
// @MX:ANCHOR: [AUTO] The single removal path of backup rotation, called by
// CleanupOldBackups and through it by both update flows.
// @MX:REASON: A change here decides which restore points survive an update; the
// destructive-target registry pins its one os.RemoveAll site (AC-UBR-19).
// @MX:SPEC: SPEC-UPDATE-BACKUP-RETENTION-001
func PruneBackups(projectRoot string, keepCount, pastKeep int) PruneReport {
	var rep PruneReport
	if OldBackupRootRefused(projectRoot) || NamespaceBackupRootRefused(projectRoot) {
		rep.RootRefused = true
		return rep
	}
	runs := collectBackupRuns(projectRoot)
	rep.Runs = len(runs)
	rep.Unmanaged = scanUnmanagedConfigEntries(projectRoot, runs)
	if keepCount < 0 || len(runs) <= keepCount {
		rep.Window = len(runs)
		return rep
	}
	rep.Window = keepCount

	// The census is the only step that reads backed-up files, and keepCount 0
	// decides without it: an explicit wipe applies no protection (REQ-UBR-014).
	census := make([]censusRun, len(runs))
	for i, run := range runs {
		if keepCount == 0 {
			census[i] = censusRun{id: run.dirs[0]}
			continue
		}
		census[i] = takeCensus(run)
	}
	plan := selectRetention(census, keepCount, pastKeep)
	rep.Protected = plan.protected

	removing := make(map[string]struct{}, len(plan.remove))
	for _, id := range plan.remove {
		removing[id] = struct{}{}
	}
	retained := make(map[string]retainedRun, len(plan.retain))
	for _, r := range plan.retain {
		retained[r.id] = r
	}

	for _, run := range runs[keepCount:] {
		id := run.dirs[0]
		item := PruneItem{Run: id, Dirs: run.dirs}
		if _, ok := removing[id]; ok {
			item.Action = RetentionActionRemoved
			for _, dir := range run.dirs {
				if err := os.RemoveAll(dir); err != nil {
					// The run could not be removed whole; it is reported as
					// retained and the next update evaluates it again (design D10).
					item.Action = RetentionActionRetained
					item.Reason = RetentionReasonRemovalFailed
					item.Detail = err.Error()
					continue
				}
				rep.Removed++
			}
		} else if r, ok := retained[id]; ok {
			item.Action = RetentionActionRetained
			item.Reason = r.reason
			item.Protects = r.protects
			item.Detail = r.detail
		}
		rep.Items = append(rep.Items, item)
	}
	return rep
}

// takeCensus reads every regular file of run once and returns its census view
// (REQ-UBR-002). A key is the file's path inside its backup directory with the
// SHA-256 of its bytes. The .template-defaults render and the bookkeeping files
// of REQ-UBR-009 are not keys. An entry that is not a regular file, a file that
// cannot be read, and a directory that cannot be walked mark the run exempt
// (REQ-UBR-008); a non-regular entry is never opened.
func takeCensus(run *backupRun) censusRun {
	c := censusRun{id: run.dirs[0]}
	markExempt := func(detail string) {
		if !c.exempt {
			c.exempt = true
			c.exemptDetail = detail
		}
	}
	for _, dir := range run.dirs {
		walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			rel, relErr := filepath.Rel(dir, path)
			if relErr != nil {
				rel = path
			}
			rel = filepath.ToSlash(rel)
			if err != nil {
				markExempt(rel)
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				if rel == templateDefaultsDirName {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				markExempt(rel)
				return nil
			}
			if !strings.Contains(rel, "/") && isBookkeepingFile(path, rel) {
				return nil
			}
			sum, hashErr := hashRegularFile(path)
			if hashErr != nil {
				markExempt(rel)
				return nil
			}
			c.keys = append(c.keys, versionKey{path: rel, sum: sum})
			return nil
		})
		if walkErr != nil {
			markExempt(filepath.ToSlash(dir))
		}
	}
	return c
}

// templateDefaultsDirName is the per-backup render of the embedded template
// defaults; it is not user content and the budget never protects it (REQ-UBR-006).
const templateDefaultsDirName = ".template-defaults"

// bookkeepingReadLimit bounds the read that decides whether a root-level file is
// writer bookkeeping. A file beyond it is not the file a writer produced.
const bookkeepingReadLimit = 16 << 20

// isBookkeepingFile reports whether the regular file at path, found at the root
// of a backup directory under the name name, is what a backup writer produced:
// the run marker (one run-id line), the namespace or inventory completion marker
// (key=value lines opening with stamp= or inventory=), or the backup metadata
// record (a BackupMetadata with its type set). A file that merely carries one of
// these names is user content and enters the census (REQ-UBR-009).
func isBookkeepingFile(path, name string) bool {
	switch name {
	case RunMarkerName, ".complete", "backup_metadata.json":
	default:
		return false
	}
	f, err := openRunMarker(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	if info, statErr := f.Stat(); statErr != nil || !info.Mode().IsRegular() {
		return false
	}
	data, err := io.ReadAll(io.LimitReader(f, bookkeepingReadLimit+1))
	if err != nil || len(data) > bookkeepingReadLimit {
		return false
	}
	text := string(data)
	switch name {
	case RunMarkerName:
		line := strings.TrimSuffix(text, "\n")
		if line == "" || strings.Contains(line, "\n") {
			return false
		}
		_, perr := time.Parse(runIDLayout, strings.TrimSpace(line))
		return perr == nil
	case ".complete":
		return isCompleteMarkerShape(text)
	default:
		var meta BackupMetadata
		dec := json.NewDecoder(strings.NewReader(text))
		dec.DisallowUnknownFields()
		if dec.Decode(&meta) != nil {
			return false
		}
		return meta.BackupType != "" && meta.Timestamp != ""
	}
}

// isCompleteMarkerShape reports whether text is a .complete marker as the
// namespace writer (stamp=, files=, timestamp=) or the preserve-inventory writer
// (inventory=) produces it: a few key=value lines, the first stamp= or inventory=.
func isCompleteMarkerShape(text string) bool {
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) == 0 || len(lines) > 4 {
		return false
	}
	for i, line := range lines {
		key, _, ok := strings.Cut(line, "=")
		if !ok {
			return false
		}
		switch key {
		case "stamp", "files", "timestamp", "inventory":
		default:
			return false
		}
		if i == 0 && key != "stamp" && key != "inventory" {
			return false
		}
	}
	return true
}

// hashRegularFile streams the bytes of the regular file at path through SHA-256.
// The file is opened without blocking and checked for regularity on the handle it
// is read from, so a FIFO swapped in after the walk is not read.
func hashRegularFile(path string) (string, error) {
	f, err := openRunMarker(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fs.ErrInvalid
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// scanUnmanagedConfigEntries counts the entries directly under the config backup
// root whose names no backup writer produces and that belong to no run, with
// their total size (REQ-UBR-012). It reads names and sizes only and never removes.
func scanUnmanagedConfigEntries(projectRoot string, runs []*backupRun) UnmanagedEntries {
	root := filepath.Join(projectRoot, defs.BackupsDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		return UnmanagedEntries{}
	}
	inRun := make(map[string]struct{})
	for _, run := range runs {
		for _, dir := range run.dirs {
			inRun[dir] = struct{}{}
		}
	}
	var out UnmanagedEntries
	for _, e := range entries {
		path := filepath.Join(root, e.Name())
		if _, ok := inRun[path]; ok || configBackupName.MatchString(e.Name()) {
			continue
		}
		out.Count++
		out.Bytes += entrySize(path)
	}
	return out
}

// entrySize returns the total size in bytes of the regular files at or under
// path. It does not follow symbolic links; a link counts as its own size.
func entrySize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if info, infoErr := d.Info(); infoErr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}
