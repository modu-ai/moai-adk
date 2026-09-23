package codexwiring

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// Interruption points of one wiring change (design §A.3 / §A.4). A pass
// stopped at a point leaves the state that point names; recovery classifies
// it from the journal.
const (
	pointJournaled       = "P1" // journal appended, temp not yet written
	pointStaged          = "P2" // temp written, target not yet re-checked
	pointRechecked       = "P3" // target re-checked, not yet renamed
	pointRenamed         = "P4" // renamed, not yet read back
	pointReadBack        = "P5" // read back, provenance not yet applied
	pointProvenance      = "P6" // provenance applied, entry not yet complete
	pointDeleteJournaled = "D1" // delete journaled, target not yet removed
	pointDeleted         = "D2" // target removed, provenance not yet removed
)

// RefusalReason names why a wiring step left a file untouched (REQ-DHR-006).
type RefusalReason string

const (
	ReasonSymlinkBoundary RefusalReason = "symlink-boundary"
	ReasonConflict        RefusalReason = "conflict"
	ReasonDiverged        RefusalReason = "diverged"
)

// Refusal is one file a wiring step left untouched, with the reason.
type Refusal struct {
	Path   string
	Reason RefusalReason
	Detail string
}

// Conflict is a change refused because the target changed under it
// (REQ-DHR-003): Expected is the pre-change hash the pass read, Observed the
// hash found immediately before rename.
type Conflict struct {
	Path, Expected, Observed string
}

// ErrWiringConflict is the non-zero outcome of a pass in which a wiring file
// changed under the write (conflict before rename, or divergence after it).
var ErrWiringConflict = errors.New("codex wiring: a wiring file changed during the write")

// writeGuards are the safety checks of a wiring change. Production always
// runs all three; tests switch one off to show it is load-bearing.
type writeGuards struct {
	recheck  bool // re-read the target immediately before rename
	readback bool // read the target back after rename
	lstat    bool // refuse symbolic links on the path (Lstat)
}

// passOptions configures one pass. fault is a test seam called at every
// interruption point; a non-nil return stops the pass on the spot, as if the
// process died there.
type passOptions struct {
	guards writeGuards
	fault  func(point, rel string) error
}

func defaultPassOptions() passOptions {
	return passOptions{guards: writeGuards{recheck: true, readback: true, lstat: true}}
}

// pass is one wiring pass under the wiring lock.
type pass struct {
	root      string
	out, warn io.Writer
	opts      passOptions
	evidence  bool
	crashed   bool
	res       Result
}

func newPass(root string, out, warn io.Writer, opts passOptions, evidence bool) *pass {
	return &pass{root: root, out: out, warn: warn, opts: opts, evidence: evidence}
}

func (p *pass) at(point, rel string) error {
	if p.opts.fault == nil {
		return nil
	}
	err := p.opts.fault(point, rel)
	if err != nil {
		p.crashed = true
	}
	return err
}

func (p *pass) refuse(rel string, reason RefusalReason, detail string) {
	p.res.Refusals = append(p.res.Refusals, Refusal{Path: rel, Reason: reason, Detail: detail})
	warnf(p.warn, "%s left untouched (%s): %s", rel, reason, detail)
}

// fileState reads the target: its hash, or "" when absent.
func fileState(path string) (string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return sha256Hex(b), nil
}

// boundaryViolation reports why rel may not be written through: the target
// itself is a symbolic link, or a directory between the root and the target
// is a symbolic link resolving outside the project root. The link itself is
// examined (Lstat), never its destination.
func boundaryViolation(root, rel string) string {
	target := filepath.Join(root, filepath.FromSlash(rel))
	if info, err := os.Lstat(target); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "target is a symbolic link"
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "project root does not resolve: " + err.Error()
	}
	dir := root
	for _, seg := range strings.Split(filepath.Dir(filepath.FromSlash(rel)), string(filepath.Separator)) {
		if seg == "" || seg == "." {
			continue
		}
		dir = filepath.Join(dir, seg)
		info, err := os.Lstat(dir)
		if err != nil {
			return "" // absent directories are created as real directories
		}
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		dest, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return "directory link " + seg + " does not resolve"
		}
		if dest != resolvedRoot && !strings.HasPrefix(dest, resolvedRoot+string(filepath.Separator)) {
			return "directory " + seg + " links outside the project root"
		}
	}
	return ""
}

// write changes rel to content through the journaled path (REQ-DHR-002):
// journal, temp, re-check, rename, read back, provenance, complete.
//
// @MX:ANCHOR: [AUTO] the single write path for every Codex wiring file change
// @MX:REASON: wiring, recovery-driven completion, and unwire all rely on this ordering; journal-before-temp and provenance-before-complete are what make every interruption point recoverable
func (p *pass) write(rel string, content []byte, entry manifest.FileEntry) (bool, error) {
	target := filepath.Join(p.root, filepath.FromSlash(rel))
	if p.opts.guards.lstat {
		if why := boundaryViolation(p.root, rel); why != "" {
			p.refuse(rel, ReasonSymlinkBoundary, why)
			return false, nil
		}
	}
	pre, err := fileState(target)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", rel, err)
	}
	post := sha256Hex(content)
	e := JournalEntry{ID: randomToken(), Op: OpWrite, Path: rel, PreHash: pre, PostHash: post, Temp: tempPrefix + randomToken(), Provenance: &entry, State: JournalStaged}
	if err := appendJournal(p.root, e); err != nil {
		return false, err
	}
	if err := p.at(pointJournaled, rel); err != nil {
		return false, err
	}
	tmp := filepath.Join(filepath.Dir(target), e.Temp)
	if err := stageTemp(tmp, content); err != nil {
		_ = os.Remove(tmp)
		_ = setJournalState(p.root, e.ID, JournalDiscarded, "")
		return false, fmt.Errorf("stage %s: %w", rel, err)
	}
	if err := p.at(pointStaged, rel); err != nil {
		return false, err
	}
	if p.opts.guards.recheck {
		cur, err := fileState(target)
		if err != nil {
			return false, fmt.Errorf("re-read %s: %w", rel, err)
		}
		if cur != pre {
			_ = os.Remove(tmp)
			if err := setJournalState(p.root, e.ID, JournalConflict, hashOrAbsent(cur)); err != nil {
				return false, err
			}
			p.res.Conflicts = append(p.res.Conflicts, Conflict{Path: rel, Expected: pre, Observed: hashOrAbsent(cur)})
			p.refuse(rel, ReasonConflict, fmt.Sprintf("changed during the write (expected %s, found %s)", hashOrAbsent(pre), hashOrAbsent(cur)))
			return false, nil
		}
	}
	if err := p.at(pointRechecked, rel); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		_ = setJournalState(p.root, e.ID, JournalDiscarded, "")
		return false, fmt.Errorf("rename %s: %w", rel, err)
	}
	if err := p.at(pointRenamed, rel); err != nil {
		return false, err
	}
	if p.opts.guards.readback {
		got, err := fileState(target)
		if err != nil {
			return false, fmt.Errorf("read back %s: %w", rel, err)
		}
		if got != post {
			if err := setJournalState(p.root, e.ID, JournalDiverged, hashOrAbsent(got)); err != nil {
				return false, err
			}
			p.res.Conflicts = append(p.res.Conflicts, Conflict{Path: rel, Expected: post, Observed: hashOrAbsent(got)})
			p.refuse(rel, ReasonDiverged, fmt.Sprintf("read back %s, intended %s", hashOrAbsent(got), post))
			return false, nil
		}
	}
	if err := p.at(pointReadBack, rel); err != nil {
		return false, err
	}
	if err := applyProvenance(p.root, rel, &entry, p.warn); err != nil {
		return false, err
	}
	if err := p.at(pointProvenance, rel); err != nil {
		return false, err
	}
	return true, setJournalState(p.root, e.ID, JournalComplete, "")
}

// remove deletes rel through the journaled path and removes its part record.
// Unwire (moai tool disable codex) is its caller; recovery completes it.
func (p *pass) remove(rel string) error {
	target := filepath.Join(p.root, filepath.FromSlash(rel))
	if p.opts.guards.lstat {
		if why := boundaryViolation(p.root, rel); why != "" {
			p.refuse(rel, ReasonSymlinkBoundary, why)
			return nil
		}
	}
	pre, err := fileState(target)
	if err != nil || pre == "" {
		return err
	}
	e := JournalEntry{ID: randomToken(), Op: OpDelete, Path: rel, PreHash: pre, State: JournalStaged}
	if err := appendJournal(p.root, e); err != nil {
		return err
	}
	if err := p.at(pointDeleteJournaled, rel); err != nil {
		return err
	}
	if p.opts.guards.recheck {
		cur, err := fileState(target)
		if err != nil {
			return err
		}
		if cur != pre {
			if err := setJournalState(p.root, e.ID, JournalConflict, hashOrAbsent(cur)); err != nil {
				return err
			}
			p.res.Conflicts = append(p.res.Conflicts, Conflict{Path: rel, Expected: pre, Observed: hashOrAbsent(cur)})
			p.refuse(rel, ReasonConflict, "changed before delete")
			return nil
		}
	}
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("delete %s: %w", rel, err)
	}
	if err := p.at(pointDeleted, rel); err != nil {
		return err
	}
	if err := applyProvenance(p.root, rel, nil, p.warn); err != nil {
		return err
	}
	return setJournalState(p.root, e.ID, JournalComplete, "")
}

// hashOrAbsent renders an absent file's empty hash visibly.
func hashOrAbsent(h string) string {
	if h == "" {
		return "absent"
	}
	return h
}

// stageTemp creates the temp file exclusively and fills it.
func stageTemp(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// applyProvenance sets (entry != nil) or removes (entry == nil) the manifest
// record of rel. Setting replaces the whole record, so applying the same
// journaled record twice leaves one copy.
func applyProvenance(root, rel string, entry *manifest.FileEntry, warn io.Writer) error {
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		if !errors.Is(err, manifest.ErrManifestCorrupt) {
			return fmt.Errorf("load manifest: %w", err)
		}
		warnf(warn, "manifest was unreadable and was reset (%v); wiring part records restart from this pass", err)
	}
	mf := mgr.Manifest()
	if entry == nil {
		delete(mf.Files, rel)
	} else {
		mf.Files[rel] = *entry
	}
	if err := mgr.Save(); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	return nil
}
