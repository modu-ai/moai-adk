package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/skillresidue"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// stateResidueMinAge is how long a candidate must have sat untouched before it
// is removed. A session that has only just created its database may still hold
// it open, and unlinking an open file would discard that session's first write.
const stateResidueMinAge = 10 * time.Minute

// removeStateResidue deletes what earlier state layouts left under .moai/state:
// zero-byte *.db stubs (an empty SQLite file, no queue data) and the zero-byte
// backlog.json.lock, the retired lock name that no current reader opens. Rollback
// snapshots such as backlog.json, live locks such as backlog.lock, populated
// databases, and anything written within stateResidueMinAge are never touched.
// It returns the removed paths relative to projectRoot so the update summary can
// list each one. The walk is best-effort: an entry it cannot read or remove is
// skipped, since a leftover stub is harmless. A second pass finds nothing new.
func removeStateResidue(projectRoot string) []string {
	stateDir := filepath.Join(projectRoot, defs.MoAIDir, "state")
	cutoff := time.Now().Add(-stateResidueMinAge)
	var removed []string
	_ = filepath.WalkDir(stateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		name := d.Name()
		if name != "backlog.json.lock" && !strings.HasSuffix(name, ".db") {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil || info.Size() != 0 || info.ModTime().After(cutoff) {
			return nil
		}
		if os.Remove(path) != nil {
			return nil
		}
		if rel, relErr := filepath.Rel(projectRoot, path); relErr == nil {
			removed = append(removed, filepath.ToSlash(rel))
		}
		return nil
	})
	return removed
}

// namespaceBackupRetainCount is how many namespace backups one update run keeps.
const namespaceBackupRetainCount = 5

// updateProjectRoot resolves the project root the way the update lock does: the
// discovered project root, falling back to the working directory.
func updateProjectRoot() string {
	root, _ := findProjectRoot()
	if root == "" {
		root, _ = os.Getwd()
	}
	return root
}

// runUpdateHousekeepingLocked runs the housekeeping for an entry point that
// returns before the main flow takes the update lock (-c). The marker check comes
// first, because acquireUpdateLock creates .moai/ in whatever directory it is
// given. A busy lock means another update is running and will housekeep, so this
// run skips rather than racing it.
func runUpdateHousekeepingLocked(projectRoot string) {
	if checkProjectMarker(projectRoot) != nil {
		return
	}
	release, err := acquireUpdateLock(projectRoot)
	if err != nil {
		return
	}
	defer release()
	runUpdateHousekeeping(projectRoot)
}

// runUpdateHousekeeping removes state residue and caps the namespace backup root
// for one moai update run. It does nothing outside a genuine MoAI project
// (checkProjectMarker), and each change is recorded as a terminal reference row.
func runUpdateHousekeeping(projectRoot string) {
	if checkProjectMarker(projectRoot) != nil {
		return
	}
	if removed := removeStateResidue(projectRoot); len(removed) > 0 {
		updateLedger.referencef("removed %d leftover state file(s): %s", len(removed), strings.Join(removed, ", "))
	}
	if backup.NamespaceBackupRootRefused(projectRoot) {
		updateLedger.requiref(sevWarn, "namespace backup pruning skipped: the backup root is a symbolic link or leads outside the project")
	} else {
		reportRetention("namespace backup", backup.PruneNamespaceBackups(projectRoot, namespaceBackupRetainCount, backupPastVersionsKeep(projectRoot)))
	}
	// Skill-folder residue: empty folders of shipped skills, a retired skill name,
	// and a lowercase skill.md. Each change is reported; see skillresidue.Clean.
	skillActions, skillErr := skillresidue.Clean(projectRoot, skillresidue.Options{
		Shipped: embeddedShips,
		Stamp:   time.Now().UTC().Format("20060102T150405Z"),
	})
	if skillErr != nil {
		updateLedger.requiref(sevWarn, "skill folder cleanup failed: %v", skillErr)
	}
	for _, a := range skillActions {
		updateLedger.referencef("%s", skillActionMessage(a))
	}
}

// embeddedShips reports whether the embedded template ships a project-relative
// path. A template that cannot be opened ships nothing, so nothing is repaired.
func embeddedShips(rel string) bool {
	tmpl, err := template.EmbeddedTemplates()
	if err != nil {
		return false
	}
	_, err = fs.Stat(tmpl, rel)
	return err == nil
}

// skillActionMessage renders one skill-folder action as a terminal reference row.
func skillActionMessage(a skillresidue.Action) string {
	switch a.Kind {
	case "removed-empty-dir":
		return fmt.Sprintf("removed empty skill folder %s (the template ships this skill; the deploy refills it)", a.Path)
	case "renamed-skill-file":
		return fmt.Sprintf("renamed %s to SKILL.md (the template ships SKILL.md and none existed)", a.Path)
	case "reported-both-skill-files":
		return fmt.Sprintf("%s holds both skill.md and SKILL.md; nothing changed, resolve by hand", a.Path)
	case "archived-retired-skill":
		return fmt.Sprintf("archived retired skill %s under .moai/archive/skills/ (its successor ships)", a.Path)
	case "kept-retired-skill-with-user-files":
		return fmt.Sprintf("kept retired skill %s: it holds files besides SKILL.md", a.Path)
	case "removed-stale-link":
		return fmt.Sprintf("removed link %s (it pointed at the archived skill)", a.Path)
	case "refused-symlinked-root":
		return fmt.Sprintf("refused to clean skill folders: %s is a symbolic link; nothing under it was changed", a.Path)
	}
	return fmt.Sprintf("%s: %s", a.Kind, a.Path)
}

// backupPastVersionsKeep is the per-file budget of superseded backup versions from
// state.yaml (backup_past_versions_keep). It falls back to the default when the
// configuration cannot be loaded or holds a negative value.
func backupPastVersionsKeep(projectRoot string) int {
	cfg, err := config.NewLoader().Load(filepath.Join(projectRoot, defs.MoAIDir))
	if err != nil || cfg.State.BackupPastVersionsKeep < 0 {
		return config.DefaultBackupPastVersionsKeep
	}
	return cfg.State.BackupPastVersionsKeep
}

// reportRetention writes one ledger row per backup of the given kind that a retention
// pass removed, named, and one row for the entries it kept, oldest first.
func reportRetention(kind string, r backup.Retention) {
	for _, name := range r.Removed {
		updateLedger.referencef("removed %s %s: it holds no superseded version within the per-file budget", kind, name)
	}
	if len(r.Kept) > 0 {
		updateLedger.referencef("kept %d older %s(s) holding a superseded version within the per-file budget (oldest: %s)", len(r.Kept), kind, r.Kept[0])
	}
}

// reportUnmanagedBackupEntries reports the entries under .moai-backups that carry no
// timestamp. The cap never prunes them, so the update shows their count and size.
func reportUnmanagedBackupEntries(projectRoot string) {
	count, size := backup.UnmanagedEntries(projectRoot)
	if count == 0 {
		return
	}
	updateLedger.referencef("%d entr(ies) under .moai-backups carry no timestamp and are never pruned (%s); review them by hand", count, humanByteCount(size))
}

// humanByteCount formats a byte count for the update report.
func humanByteCount(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
