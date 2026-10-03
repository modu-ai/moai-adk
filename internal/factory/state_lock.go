// state_lock.go — the shared cross-process file-lock substrate (the type, its
// contention sentinels, and the owner-identity record) that the todo queue
// (backlog_store.go), the integration lock (integration_lock_mutation.go), and
// the slot lease (slot_lease.go) each take over their own lock path
// (SPEC-KANBAN-BOARD-001 REQ-KB-019/023, kept after the board went:
// SPEC-LAUNCHER-ENTRY-FLAGS-001 M6).
//
// The substrate reuses the repository's existing cross-process per-scope lock
// PATTERN (internal/spec/lock.go and its platform counterparts: flock on Unix,
// atomic-create on Windows); internal/lockfile's in-process mutex is neither
// used nor upgraded.
package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrStateLockHeld is returned by a state-lock acquisition when another
// process holds the lock at that path.
var ErrStateLockHeld = errors.New("factory state lock held")

// IsStateLockHeld reports whether err is the contention sentinel.
func IsStateLockHeld(err error) bool {
	return errors.Is(err, ErrStateLockHeld)
}

// ErrStateLockUnsafePath is returned when the lock path would route the
// owner-record write outside the project or onto a shared inode: a symlinked
// lock file or ancestor directory, a hardlinked artifact, or a non-regular
// file. It is a refusal, never contention.
var ErrStateLockUnsafePath = errors.New("factory state lock path is unsafe")

// stateLockProjectDirName is the project-state directory that bounds the
// ancestor check: a lock path beneath a `.moai` directory is validated from its
// immediate parent up to and including that `.moai`, and never above it (the OS
// temp root or the project root may legitimately sit behind a symlink, e.g.
// macOS /var -> /private/var).
const stateLockProjectDirName = ".moai"

// unsafeStateLockPath builds the refusal error for lockPath.
func unsafeStateLockPath(lockPath, reason string) error {
	return fmt.Errorf("%w: %s: %s", ErrStateLockUnsafePath, lockPath, reason)
}

// isLinkMode reports whether mode names a symlink or a Windows reparse-point
// style irregular entry.
func isLinkMode(mode os.FileMode) bool {
	return mode&(os.ModeSymlink|os.ModeIrregular) != 0
}

// checkStateLockAncestors refuses a lock whose parent directory (or, when the
// path lies beneath a `.moai` directory, any directory from the parent up to
// and including `.moai`) is a symlink: such a link would redirect the lock
// artifact — and its owner-record write — outside the project.
//
// Depth: when no `.moai` ancestor exists only the immediate parent is checked.
// Ancestors above `.moai` are deliberately NOT checked (see
// stateLockProjectDirName); a symlink there is a residual this opener cannot
// distinguish from a legitimate one given only lockPath.
//
// A missing directory is not refused here; the subsequent open reports it.
func checkStateLockAncestors(lockPath string) error {
	parent := filepath.Dir(lockPath)
	dirs := []string{parent}
	for dir := parent; ; {
		next := filepath.Dir(dir)
		if next == dir {
			// No `.moai` ancestor: immediate parent only.
			dirs = dirs[:1]
			break
		}
		if filepath.Base(dir) == stateLockProjectDirName {
			break
		}
		dirs = append(dirs, next)
		dir = next
	}
	for _, dir := range dirs {
		info, err := os.Lstat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("inspect board lock directory %s: %w", dir, err)
		}
		if isLinkMode(info.Mode()) {
			return unsafeStateLockPath(lockPath, "directory "+dir+" is a symlink")
		}
	}
	return nil
}

// ensureStateLockDir creates lockPath's parent directory, but only after the
// same ancestor Lstat chain the opener runs has proven no checked directory is
// a symlink. os.MkdirAll follows links, so creating first and refusing in the
// opener would already have created the lock directory OUTSIDE the project
// through a symlinked `.moai`. The chain tolerates directories that do not
// exist yet (checkStateLockAncestors skips ENOENT) and keeps its depth rule:
// up to and including `.moai`, never above it. The opener repeats the check
// (defense in depth); a swap between this check and the open stays the
// accepted residual (card t1458).
//
// @MX:ANCHOR: [AUTO] Single pre-create guard for every caller that makes a state-lock directory (AcquireFactoryStepLock).
// @MX:REASON: a MkdirAll that runs before this check mutates the directory a symlinked ancestor points at, whatever the opener later refuses (card t1458 P2).
// @MX:SPEC: SPEC-FACTORY-ATOMIC-LEASE-001
func ensureStateLockDir(lockPath string) error {
	if err := checkStateLockAncestors(lockPath); err != nil {
		return err
	}
	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating lock directory %s: %w", dir, err)
	}
	return nil
}

// checkStateLockArtifact refuses a pre-existing lock artifact that is a
// symlink or any other non-regular file. An absent artifact is fine.
func checkStateLockArtifact(lockPath string) error {
	info, err := os.Lstat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("inspect board lock %s: %w", lockPath, err)
	}
	if isLinkMode(info.Mode()) {
		return unsafeStateLockPath(lockPath, "lock file is a symlink")
	}
	if !info.Mode().IsRegular() {
		return unsafeStateLockPath(lockPath, "lock file is not a regular file")
	}
	return nil
}

// ErrStateLockChangedHands is returned by the Windows stale-lock clear when the
// pre-removal re-read observes a different recorded identity than the
// inspection did — the artifact was released and re-acquired inside the
// window, and the clear aborts rather than unlinking a valid lock.
var ErrStateLockChangedHands = errors.New("factory state lock changed hands between inspection and removal")

// IsStateLockChangedHands reports whether err is the changed-hands abort.
func IsStateLockChangedHands(err error) bool {
	return errors.Is(err, ErrStateLockChangedHands)
}

// StateLockOwner is the creating process's identity recorded IN the lock
// artifact (REQ-KB-023). The identity is what makes a stale artifact
// distinguishable from a live holder's: without it, "the holder is gone" is a
// guess, and clearing on a guess unlinks a lock a live process may hold.
type StateLockOwner struct {
	PID       int    `json:"pid"`
	CreatedAt string `json:"created_at"`
}

// StateLock represents an acquired state lock. Callers MUST call Release when
// the guarded read-modify-write completes.
type StateLock struct {
	path string
	impl stateLockImpl
}

// stateLockImpl is the platform-specific lock implementation (flock on Unix,
// atomic-create on Windows — mirroring internal/spec's substrate split).
type stateLockImpl interface {
	release() error
}

// Release releases the state lock. Safe to call multiple times.
func (l *StateLock) Release() error {
	if l == nil || l.impl == nil {
		return nil
	}
	err := l.impl.release()
	l.impl = nil
	return err
}

// Path returns the lock artifact's path (diagnostics).
func (l *StateLock) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

// ClearStaleReport is what a clear operation observed and did — the operation
// is explicit and operator-visible, so it REPORTS rather than acting silently
// (REQ-KB-023).
type ClearStaleReport struct {
	// Removed is true only when the artifact was unlinked by this call.
	Removed bool
	// PID names the recorded owner the decision was made about.
	PID int
	// Reason names the observation that decided the outcome.
	Reason string
}

// newLockOwnerRecord builds the owner identity block written at acquisition.
func newLockOwnerRecord() []byte {
	owner := StateLockOwner{
		PID:       os.Getpid(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	encoded, err := json.MarshalIndent(owner, "", "  ")
	if err != nil {
		// Marshalling a two-field struct cannot fail; fall back to the
		// minimal record rather than failing the acquisition.
		return []byte(fmt.Sprintf("{\"pid\":%d}\n", owner.PID))
	}
	return append(encoded, '\n')
}
