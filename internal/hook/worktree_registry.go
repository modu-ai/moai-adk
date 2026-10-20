package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// WorktreeEntry records a currently active isolated worktree.
type WorktreeEntry struct {
	Path      string    `json:"path"`
	Branch    string    `json:"branch"`
	AgentName string    `json:"agent_name"`
	CreatedAt time.Time `json:"created_at"`
}

// StateLockLapseBudget is the number of state-file lock waits a migration
// run may spend before it ends: one lapse ends the run, the remaining trees
// are skipped with the reason and retried by the next run
// (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-016/017, design.md D-12). The M6
// engine consumes this constant beside the wait bound below.
const StateLockLapseBudget = 1

// stateLockWaitBudget bounds one writer's wait for the state-file lock. The
// value is the 5-second default; the test seam lowers it for the lapse
// tests (AC-WHR-017b step (7)).
var stateLockWaitBudget = 5 * time.Second

// ErrStateLockWaitLapsed is the distinct error a writer returns when the
// state-file lock stayed held past the wait budget — never conflated with
// the lock's other failures (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-017).
var ErrStateLockWaitLapsed = errors.New("state-file lock wait budget spent")

// errM5Injected is the fault the atomic-write tests inject between the temp
// write and the rename (AC-WHR-017b step (7)'s fault-injection seam,
// worktreeAtomicFault).
var errM5Injected = errors.New("injected fault between the temp write and the rename")

// worktreeAtomicFault, when non-nil, runs between the temp-file write and
// the rename — the fault-injection seam of AC-WHR-017b step (7)'s atomicity
// clause. Production leaves it nil.
var worktreeAtomicFault func() error

// @MX:WARN: [AUTO] Package-level mutex guards shared worktree state file; all reads and writes must acquire this lock
// @MX:REASON: global state mutation — concurrent hook handlers (WorktreeCreate/WorktreeRemove) share this lock; missing lock acquisition causes data races on the JSON state file
var worktreeMu sync.Mutex

// registerWorktree records a WorktreeEntry in the persistent state file,
// replacing any existing entry with the same path (idempotent — the
// WorktreeCreate handler re-registers when it reuses an existing worktree
// directory, issue #1570). If the state directory or file does not exist it
// is created.
//
// SPEC-WORKTREE-HOME-ROOT-001 M5 (REQ-WHR-017): the read-modify-write runs
// under the OS-held state-file lock with a bounded wait, the load is strict
// (an unparseable file fails the step instead of rewriting from an empty
// slice), and the save is a temp-file write plus atomic rename. The error
// is returned; the WorktreeCreate handler logs it.
func registerWorktree(projectDir, path, branch, agentName string) error {
	worktreeMu.Lock()
	defer worktreeMu.Unlock()

	stateFile := worktreeStateFile(projectDir)
	var entries []WorktreeEntry
	if err := withStateFileLock(stateFile, func() error {
		var err error
		entries, err = loadWorktreeEntriesStrict(stateFile)
		if err != nil {
			return err
		}
		replaced := false
		for i := range entries {
			if entries[i].Path == path {
				entries[i] = WorktreeEntry{
					Path:      path,
					Branch:    branch,
					AgentName: agentName,
					CreatedAt: time.Now(),
				}
				replaced = true
				break
			}
		}
		if !replaced {
			entries = append(entries, WorktreeEntry{
				Path:      path,
				Branch:    branch,
				AgentName: agentName,
				CreatedAt: time.Now(),
			})
		}
		return saveWorktreeEntriesAtomic(stateFile, entries)
	}); err != nil {
		return err
	}
	return nil
}

// RegisterLauncherWorktree records a worktree created by the moai launcher's
// `-w <name> --branch <existing>` path (card t295) in the same persistent
// state file the WorktreeCreate hook maintains, so `moai worktree clean`'s
// anchor check sees launcher-created trees too. Non-blocking on error.
func RegisterLauncherWorktree(projectDir, path, branch, agentName string) {
	_ = registerWorktree(projectDir, path, branch, agentName)
}

// RegisterLauncherWorktreeE is RegisterLauncherWorktree with the step's
// error surfaced — the form the M5 state-file tests and the M6 engine call
// (AC-WHR-017b: a writer whose lock wait lapses fails its step).
func RegisterLauncherWorktreeE(projectDir, path, branch, agentName string) error {
	return registerWorktree(projectDir, path, branch, agentName)
}

// unregisterWorktree removes the entry matching the given path from the
// persistent state file.
//
// SPEC-WORKTREE-HOME-ROOT-001 M5 (REQ-WHR-017): locked, strict-load and
// atomically saved like every writer; the error is returned.
func unregisterWorktree(projectDir, path string) error {
	worktreeMu.Lock()
	defer worktreeMu.Unlock()

	stateFile := worktreeStateFile(projectDir)
	var filtered []WorktreeEntry
	if err := withStateFileLock(stateFile, func() error {
		entries, err := loadWorktreeEntriesStrict(stateFile)
		if err != nil {
			return err
		}
		filtered = make([]WorktreeEntry, 0, len(entries))
		for _, e := range entries {
			if e.Path != path {
				filtered = append(filtered, e)
			}
		}
		return saveWorktreeEntriesAtomic(stateFile, filtered)
	}); err != nil {
		return err
	}
	return nil
}

// RekeyWorktreeEntry replaces the entry keyed to oldPath with an entry keyed
// to newPath, keeping the entry's other fields (SPEC-WORKTREE-HOME-ROOT-001
// M5, REQ-WHR-012a — the state-file re-key the migration performs after
// `git worktree move`). A re-key naming a path with no entry fails without
// writing. Locked, strict-load and atomically saved like every writer.
func RekeyWorktreeEntry(projectDir, oldPath, newPath string) error {
	return rekeyWorktreeEntry(projectDir, oldPath, newPath)
}

// RekeyWorktreeEntryInverse is the re-key's inverse: it replaces the entry
// keyed to newPath with an entry keyed to oldPath — the reverse move's
// re-key (REQ-WHR-012, the migration's rollback direction).
func RekeyWorktreeEntryInverse(projectDir, newPath, oldPath string) error {
	return rekeyWorktreeEntry(projectDir, newPath, oldPath)
}

func rekeyWorktreeEntry(projectDir, from, to string) error {
	worktreeMu.Lock()
	defer worktreeMu.Unlock()

	stateFile := worktreeStateFile(projectDir)
	if err := withStateFileLock(stateFile, func() error {
		entries, err := loadWorktreeEntriesStrict(stateFile)
		if err != nil {
			return err
		}
		found := false
		for i := range entries {
			if entries[i].Path == from {
				entries[i].Path = to
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("worktree registry: no entry keyed to %s", from)
		}
		return saveWorktreeEntriesAtomic(stateFile, entries)
	}); err != nil {
		return err
	}
	return nil
}

// worktreeStateFile returns the canonical path to the worktrees state file.
func worktreeStateFile(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", "worktrees.json")
}

// stateLockPath returns the lock file of the state file at path: the same
// name with a .lock suffix, created on first acquisition. An OS-held flock
// (Unix) or process-local mutex (Windows) on it is the state-file lock of
// design.md D-12.
func stateLockPath(stateFile string) string {
	return stateFile + ".lock"
}

// withStateFileLock runs fn while holding the state-file lock of the given
// state file. The lock is acquired with lockfile.TryLock (never blocking)
// and the wait is bounded by stateLockWaitBudget: a budget spent is the
// distinct ErrStateLockWaitLapsed — the caller fails its step and never
// rewrites the file (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-017, design.md
// D-12). The lock is OS-held and leaves no blocking file behind: a holder
// killed by SIGKILL releases it with its process.
func withStateFileLock(stateFile string, fn func() error) error {
	lockPath := stateLockPath(stateFile)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return fmt.Errorf("state-file lock: %w", err)
	}
	deadline := time.Now().Add(stateLockWaitBudget)
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return fmt.Errorf("state-file lock: open %s: %w", lockPath, err)
		}
		if terr := lockfile.TryLock(f); terr == nil {
			err := fn()
			_ = lockfile.Unlock(f)
			_ = f.Close()
			return err
		}
		_ = f.Close()
		if !time.Now().Before(deadline) {
			return fmt.Errorf("%w: %s stayed held for %s", ErrStateLockWaitLapsed, lockPath, stateLockWaitBudget)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// loadWorktreeEntriesStrict reads entries from the state file for a writer:
// a missing file yields an empty slice (the first writer creates it), but a
// file that does not parse is an error — the writer fails its step and
// never rewrites the file from an empty slice
// (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-017).
func loadWorktreeEntriesStrict(path string) ([]WorktreeEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []WorktreeEntry{}, nil
		}
		return nil, fmt.Errorf("worktree registry: read %s: %w", path, err)
	}
	var entries []WorktreeEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("worktree registry: parse %s: %w", path, err)
	}
	return entries, nil
}

// loadWorktreeEntries is the lenient read for consumers that only inspect
// the file: a missing or unparseable file yields an empty slice. The four
// state-file writers never use this form (REQ-WHR-017).
func loadWorktreeEntries(path string) []WorktreeEntry {
	entries, _ := loadWorktreeEntriesStrict(path)
	return entries
}

// marshalWorktreeEntries renders the entries for a state-file write.
func marshalWorktreeEntries(entries []WorktreeEntry) ([]byte, error) {
	return json.MarshalIndent(entries, "", "  ")
}

// saveWorktreeEntriesAtomic replaces the state file through a temp file and
// an atomic rename: a fault or crash between the write and the rename
// leaves the previous file intact and parseable — the file is never
// truncated in place (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-017,
// AC-WHR-017b step (7)). The caller holds the state-file lock.
func saveWorktreeEntriesAtomic(path string, entries []WorktreeEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("worktree registry: create state directory %s: %w", filepath.Dir(path), err)
	}
	data, err := marshalWorktreeEntries(entries)
	if err != nil {
		return fmt.Errorf("worktree registry: marshal entries: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("worktree registry: write %s: %w", tmp, err)
	}
	if worktreeAtomicFault != nil {
		if ferr := worktreeAtomicFault(); ferr != nil {
			_ = os.Remove(tmp)
			return ferr
		}
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("worktree registry: rename into %s: %w", path, err)
	}
	return nil
}
