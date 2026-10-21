package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/skillresidue"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// stateResidueMinAge is how long a candidate must have sat untouched before it
// is removed. A session that has only just created its database may still hold
// it open, and unlinking an open file would discard that session's first write.
const stateResidueMinAge = 10 * time.Minute

// removeStateResidue deletes the one residue file earlier state layouts left under
// .moai/state that is safe to delete: the zero-byte backlog.json.lock, the retired
// lock name that no current reader opens. Rollback snapshots such as backlog.json,
// live locks such as backlog.lock, every database (empty or not: see
// emptyStateDBs), and anything written within stateResidueMinAge are never
// touched. A symbolic .moai refuses the whole pass: isMoAIProject follows links,
// so the walk would otherwise reach a tree outside the project. It returns the
// removed paths relative to projectRoot so the update summary can list each one.
// The walk is best-effort: an entry it cannot read or remove is skipped, since a
// leftover file is harmless. A second pass finds nothing new.
func removeStateResidue(projectRoot string) []string {
	var removed []string
	walkStateResidue(projectRoot, func(path, name string) {
		if name != "backlog.json.lock" || os.Remove(path) != nil {
			return
		}
		if rel, relErr := filepath.Rel(projectRoot, path); relErr == nil {
			removed = append(removed, filepath.ToSlash(rel))
		}
	})
	return removed
}

// emptyStateDBs lists the zero-byte *.db files under .moai/state that are older
// than stateResidueMinAge and have no write-ahead log or shared-memory sidecar,
// relative to projectRoot (REQ-UBR-018). Nothing is removed: a process that opened
// such a file without writing to it takes an I/O error on its next write if the
// file is unlinked, so the update only reports it. The same refusals as
// removeStateResidue apply, including the symbolic .moai.
func emptyStateDBs(projectRoot string) []string {
	var empty []string
	walkStateResidue(projectRoot, func(path, name string) {
		if !strings.HasSuffix(name, ".db") || walSidecarPresent(path) {
			return
		}
		if rel, relErr := filepath.Rel(projectRoot, path); relErr == nil {
			empty = append(empty, filepath.ToSlash(rel))
		}
	})
	return empty
}

// walkStateResidue calls visit for every zero-byte regular backlog.json.lock or
// *.db under .moai/state that has sat untouched for stateResidueMinAge. A symbolic
// .moai ends the walk before it starts, and an entry it cannot read is skipped.
func walkStateResidue(projectRoot string, visit func(path, name string)) {
	// Lstat, not Stat: a symbolic .moai is refused, not followed.
	if info, err := os.Lstat(filepath.Join(projectRoot, defs.MoAIDir)); err != nil || info.Mode()&fs.ModeSymlink != 0 {
		return
	}
	stateDir := filepath.Join(projectRoot, defs.MoAIDir, "state")
	cutoff := time.Now().Add(-stateResidueMinAge)
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
		visit(path, name)
		return nil
	})
}

// walSidecarPresent reports whether the SQLite write-ahead log or shared-memory
// file beside the database at path exists. In WAL mode the committed pages stay in
// the -wal file until a checkpoint, so a zero-byte main file with a sidecar may
// hold data. A lookup that fails for any reason other than "does not exist" counts
// as present, so the stub is not reported as empty on an unknown answer.
func walSidecarPresent(path string) bool {
	for _, suffix := range []string{"-wal", "-shm"} {
		_, err := os.Lstat(path + suffix)
		if err == nil || !os.IsNotExist(err) {
			return true
		}
	}
	return false
}

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

// runUpdateHousekeeping removes state residue, repairs skill-folder residue, and
// applies the held-back skill.md rename, for one moai update run that has no template
// sync to confirm (the binary-only run and the config-edit path). It does nothing
// outside a genuine MoAI project (checkProjectMarker), and each change is recorded as
// a terminal reference row.
func runUpdateHousekeeping(projectRoot string) {
	runUpdateHousekeepingDeferRenames(projectRoot)
	applyDeferredSkillRenames(projectRoot)
}

// runUpdateHousekeepingDeferRenames is runUpdateHousekeeping for the normal update flow,
// which can still be cancelled at the template-sync confirmation. The lowercase skill.md
// rename is held back here and applied by applyDeferredSkillRenames after the user
// confirms, so a cancelled update leaves every skill file where it was.
func runUpdateHousekeepingDeferRenames(projectRoot string) {
	if checkProjectMarker(projectRoot) != nil {
		return
	}
	if removed := removeStateResidue(projectRoot); len(removed) > 0 {
		updateLedger.referencef("removed %d leftover state file(s): %s", len(removed), strings.Join(removed, ", "))
	}
	// REQ-UBR-018: an empty database is reported and left in place.
	for _, rel := range emptyStateDBs(projectRoot) {
		updateLedger.referencef("empty DB file present: %s (zero bytes, left in place; nothing was removed)", rel)
	}
	// Skill-folder residue: empty folders of shipped skills, a lowercase skill.md
	// (held back, see above), and retired skill folders, which are reported and left in
	// place. Each action is reported; see skillresidue.Clean.
	skillActions, skillErr := skillresidue.Clean(projectRoot, skillresidue.Options{
		Shipped:     embeddedShips,
		SkipRenames: true,
	})
	if skillErr != nil {
		updateLedger.requiref(sevWarn, "skill folder cleanup failed: %v", skillErr)
	}
	for _, a := range skillActions {
		updateLedger.referencef("%s", skillActionMessage(a))
	}
}

// applyDeferredSkillRenames performs the lowercase skill.md renames that
// runUpdateHousekeepingDeferRenames held back. Call it only once the update is
// confirmed or needs no confirmation.
func applyDeferredSkillRenames(projectRoot string) {
	if checkProjectMarker(projectRoot) != nil {
		return
	}
	actions, err := skillresidue.ApplyDeferredRenames(projectRoot, skillresidue.Options{
		Shipped: embeddedShips,
	})
	if err != nil {
		updateLedger.requiref(sevWarn, "skill folder rename failed: %v", err)
	}
	for _, a := range actions {
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
	case "reported-retired-skill":
		return fmt.Sprintf("left retired skill %s in place: its successor ships, and this update does not archive it", a.Path)
	case "kept-retired-skill-with-user-files":
		return fmt.Sprintf("kept retired skill %s: it holds files besides SKILL.md", a.Path)
	case "reported-retired-link":
		return fmt.Sprintf("left link %s in place: it points at a retired skill folder that this update does not remove", a.Path)
	case "kept-non-file-skill-md":
		return fmt.Sprintf("kept %s: it is not a regular file, so it is left as the user's and not renamed", a.Path)
	case "refused-symlinked-root":
		return fmt.Sprintf("refused to clean skill folders: %s is a symbolic link; nothing under it was changed", a.Path)
	}
	return fmt.Sprintf("%s: %s", a.Kind, a.Path)
}
