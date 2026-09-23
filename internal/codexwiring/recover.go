package codexwiring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RecoveryClass is how recovery classified an interrupted change
// (REQ-DHR-004, design §A.4).
type RecoveryClass string

const (
	// RecoveryCompleted: the target is in the intended post-change state.
	RecoveryCompleted RecoveryClass = "completed"
	// RecoveryNotApplied: the target is in the pre-change state.
	RecoveryNotApplied RecoveryClass = "not-applied"
	// RecoveryDiverged: the target is in neither state and is left untouched.
	RecoveryDiverged RecoveryClass = "diverged"
	// RecoveryOrphanTemp: a temp file no journal entry references.
	RecoveryOrphanTemp RecoveryClass = "orphan-temp"
)

// RecoveryOutcome is one recovered journal entry or orphan temp file.
type RecoveryOutcome struct {
	Path  string
	Op    string
	Class RecoveryClass
	Temp  string
}

// RecoverCommand is the command that runs recovery, named by doctor.
const RecoverCommand = "moai tool enable codex"

// wiringTargetDirs are the directories wiring temp files are staged in.
var wiringTargetDirs = []string{".codex"}

// Recover classifies and resolves every interrupted wiring change of the
// project, under the wiring lock. It is the recovery step the enable,
// disable, and update-path entry points run before any new write.
func Recover(projectRoot string, warn io.Writer) ([]RecoveryOutcome, error) {
	release, err := acquireWiringLock(projectRoot)
	if err != nil {
		return nil, err
	}
	defer release()
	return recoverLocked(projectRoot, warn)
}

// recoverLocked is Recover for a caller already holding the wiring lock.
//
// @MX:WARN: [AUTO] deletes files in the project's wiring directories
// @MX:REASON: only temp files carrying the wiring prefix are deleted, never a target; a target in neither the pre nor the post state is reported and left untouched
func recoverLocked(root string, warn io.Writer) ([]RecoveryOutcome, error) {
	entries, err := LoadJournal(root)
	if err != nil {
		return nil, err
	}
	var outcomes []RecoveryOutcome
	for i := range entries {
		e := &entries[i]
		if e.State != JournalStaged {
			continue
		}
		target := filepath.Join(root, filepath.FromSlash(e.Path))
		cur, err := fileState(target)
		if err != nil {
			return outcomes, fmt.Errorf("recover %s: %w", e.Path, err)
		}
		o := RecoveryOutcome{Path: e.Path, Op: e.Op, Temp: e.Temp}
		switch cur {
		case e.PostHash:
			if err := applyProvenance(root, e.Path, e.Provenance, warn); err != nil {
				return outcomes, err
			}
			o.Class, e.State = RecoveryCompleted, JournalComplete
		case e.PreHash:
			o.Class, e.State = RecoveryNotApplied, JournalDiscarded
		default:
			o.Class, e.State, e.ObservedHash = RecoveryDiverged, JournalDiverged, hashOrAbsent(cur)
			warnf(warn, "%s changed during an interrupted wiring change and was left untouched (expected %s or %s, found %s)",
				e.Path, hashOrAbsent(e.PreHash), hashOrAbsent(e.PostHash), hashOrAbsent(cur))
		}
		if e.Temp != "" {
			tmp := filepath.Join(filepath.Dir(target), e.Temp)
			if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
				return outcomes, fmt.Errorf("remove %s: %w", tmp, err)
			}
		}
		outcomes = append(outcomes, o)
	}
	// Every staged entry is now resolved, so no temp file is referenced any
	// more; what remains is an orphan (an older binary's writeAtomic, or a
	// crash before the journal existed).
	orphans, err := orphanTemps(root, nil)
	if err != nil {
		return outcomes, err
	}
	for _, rel := range orphans {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil && !os.IsNotExist(err) {
			return outcomes, fmt.Errorf("remove orphan %s: %w", rel, err)
		}
		outcomes = append(outcomes, RecoveryOutcome{Path: filepath.ToSlash(filepath.Dir(rel)), Class: RecoveryOrphanTemp, Temp: filepath.Base(rel)})
	}
	// Resolved entries were reported by the pass or recovery that resolved
	// them; the journal keeps only what is still open.
	var open []JournalEntry
	for _, e := range entries {
		if e.State == JournalStaged {
			open = append(open, e)
		}
	}
	if len(entries) > 0 {
		if err := saveJournal(root, open); err != nil {
			return outcomes, err
		}
	}
	return outcomes, nil
}

// orphanTemps lists wiring temp files (project-relative, slash form) that no
// entry in referenced names. A directory that is a symbolic link out of the
// project is not scanned.
func orphanTemps(root string, referenced map[string]bool) ([]string, error) {
	var out []string
	for _, dir := range wiringTargetDirs {
		if why := boundaryViolation(root, dir+"/x"); why != "" {
			continue
		}
		items, err := os.ReadDir(filepath.Join(root, dir))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			if it.IsDir() || !strings.HasPrefix(it.Name(), tempPrefix) || referenced[it.Name()] {
				continue
			}
			out = append(out, dir+"/"+it.Name())
		}
	}
	return out, nil
}

// JournalReport is the read-only view doctor prints (REQ-DHR-004): the
// incomplete entries and the unreferenced temp files, with the command that
// recovers them.
type JournalReport struct {
	Incomplete  []JournalEntry
	OrphanTemps []string
	Command     string
	Err         error
}

// InspectJournal reports incomplete wiring changes without writing, renaming,
// or deleting anything and without taking the lock.
func InspectJournal(projectRoot string) JournalReport {
	r := JournalReport{Command: RecoverCommand}
	entries, err := LoadJournal(projectRoot)
	if err != nil {
		r.Err = err
		return r
	}
	referenced := map[string]bool{}
	for _, e := range entries {
		if e.State == JournalStaged {
			r.Incomplete = append(r.Incomplete, e)
			if e.Temp != "" {
				referenced[e.Temp] = true
			}
		}
	}
	r.OrphanTemps, r.Err = orphanTemps(projectRoot, referenced)
	return r
}
