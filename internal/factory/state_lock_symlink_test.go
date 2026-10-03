//go:build !windows

// state_lock_symlink_test.go — the shared state-lock opener must refuse a
// lock path that would route its owner-record write outside the project
// (symlinked lock file, symlinked parent directory, hardlinked or non-regular
// artifact). The shared opener serves the queue lock and the factory
// worktree-step lock alike, so both entry points are driven here (card t1458,
// P1 repair, ported onto the state-lock names by the develop absorb).
package factory

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// outsideBytes is the known content planted in a file outside the project; a
// lock acquisition that follows a link to it would replace it with the owner
// record.
var outsideBytes = []byte("OUTSIDE-FILE-MUST-SURVIVE\n")

// lockEntryPoint drives one real acquisition path.
type lockEntryPoint struct {
	name string
	// relDir is the lock's parent directory beneath the project root.
	relDir string
	// file is the lock artifact's base name.
	file string
	// acquire takes the lock; the release function is nil on error.
	acquire func(root string) (func() error, error)
	// guardsDirCreation is true when the entry point runs the pre-create
	// ancestor check (ensureStateLockDir) before making the lock's parent
	// directories, so a symlinked ancestor is refused without any write
	// through the link. The queue lock creates its directory itself before
	// the opener runs, so it is excluded from the no-write tests.
	guardsDirCreation bool
}

func lockEntryPoints() []lockEntryPoint {
	return []lockEntryPoint{
		{
			name:   "queue-lock",
			relDir: filepath.Join(".moai", "state"),
			file:   backlogLockFileName,
			acquire: func(root string) (func() error, error) {
				store := NewBacklogStore(filepath.Join(root, ".moai", "state", backlogFileName))
				l, err := store.acquireLock()
				if err != nil {
					return nil, err
				}
				return l.Release, nil
			},
		},
		{
			name:              "factory-step-lock",
			relDir:            filepath.Join(".moai", "state"),
			file:              filepath.Base(factoryStepLockRelPath),
			guardsDirCreation: true,
			acquire: func(root string) (func() error, error) {
				rel, err := AcquireFactoryStepLock(root, 50*time.Millisecond)
				if err != nil {
					return nil, err
				}
				return rel, nil
			},
		},
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertBytesUnchanged(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("outside file %s was modified: got %q want %q", path, got, want)
	}
}

func assertDirEntryCount(t *testing.T, dir string, want int) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	if len(entries) != want {
		t.Fatalf("outside dir %s has %d entries, want %d", dir, len(entries), want)
	}
}

func assertRefusedUnsafe(t *testing.T, release func() error, err error) {
	t.Helper()
	if err == nil {
		if release != nil {
			_ = release()
		}
		t.Fatal("acquisition succeeded; want refusal of the unsafe lock path")
	}
	if !errors.Is(err, ErrStateLockUnsafePath) {
		t.Fatalf("error %v does not wrap ErrStateLockUnsafePath", err)
	}
	if IsStateLockHeld(err) {
		t.Fatalf("unsafe-path refusal must not read as contention: %v", err)
	}
}

// (a) the lock FILE is a symlink to an existing outside file.
func TestStateLockSymlinkedLockFileRefused(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			outside := filepath.Join(t.TempDir(), "victim.txt")
			mustWrite(t, outside, outsideBytes)

			dir := filepath.Join(root, ep.relDir)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, filepath.Join(dir, ep.file)); err != nil {
				t.Fatal(err)
			}

			release, err := ep.acquire(root)
			assertBytesUnchanged(t, outside, outsideBytes)
			assertDirEntryCount(t, filepath.Dir(outside), 1)
			assertRefusedUnsafe(t, release, err)
		})
	}
}

// (b) the lock's PARENT directory is a symlink to an outside directory that
// already holds a regular file of the lock's name.
func TestStateLockSymlinkedParentDirRefused(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			outsideDir := t.TempDir()
			victim := filepath.Join(outsideDir, ep.file)
			mustWrite(t, victim, outsideBytes)

			link := filepath.Join(root, ep.relDir)
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outsideDir, link); err != nil {
				t.Fatal(err)
			}

			release, err := ep.acquire(root)
			assertBytesUnchanged(t, victim, outsideBytes)
			assertDirEntryCount(t, outsideDir, 1)
			assertRefusedUnsafe(t, release, err)
		})
	}
}

// A symlinked `.moai` directory (the topmost ancestor the opener checks) is
// refused too.
func TestStateLockSymlinkedMoaiDirRefused(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			outsideMoai := t.TempDir()
			sub, err := filepath.Rel(".moai", ep.relDir)
			if err != nil {
				t.Fatal(err)
			}
			victim := filepath.Join(outsideMoai, sub, ep.file)
			mustWrite(t, victim, outsideBytes)

			if err := os.Symlink(outsideMoai, filepath.Join(root, ".moai")); err != nil {
				t.Fatal(err)
			}

			release, aerr := ep.acquire(root)
			assertBytesUnchanged(t, victim, outsideBytes)
			assertRefusedUnsafe(t, release, aerr)
		})
	}
}

// Positive control: a normal acquisition still succeeds and records the owner.
func TestStateLockNormalAcquisitionStillSucceeds(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			release, err := ep.acquire(root)
			if err != nil {
				t.Fatalf("normal acquisition failed: %v", err)
			}
			defer func() { _ = release() }()
			data, err := os.ReadFile(filepath.Join(root, ep.relDir, ep.file))
			if err != nil || len(data) == 0 {
				t.Fatalf("owner record missing: %v %q", err, data)
			}
		})
	}
}

// Positive control: a project root reached through a symlink (macOS /var ->
// /private/var is the everyday case) is not an attack; only the project's own
// `.moai` subtree is checked.
func TestStateLockSymlinkedProjectRootAllowed(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "rootlink")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			release, err := ep.acquire(link)
			if err != nil {
				t.Fatalf("symlinked project root must be accepted: %v", err)
			}
			_ = release()
		})
	}
}

// A hardlinked lock file (nlink > 1) is refused: the write would reach every
// other name of the inode.
func TestStateLockHardlinkedLockFileRefused(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			outside := filepath.Join(t.TempDir(), "victim.txt")
			mustWrite(t, outside, outsideBytes)

			dir := filepath.Join(root, ep.relDir)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(outside, filepath.Join(dir, ep.file)); err != nil {
				t.Skipf("hardlink unsupported here: %v", err)
			}

			release, err := ep.acquire(root)
			assertBytesUnchanged(t, outside, outsideBytes)
			assertRefusedUnsafe(t, release, err)
		})
	}
}

// A non-regular artifact (FIFO) is refused without blocking.
func TestStateLockFIFOLockFileRefusedWithoutBlocking(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, ep.relDir)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := unix.Mkfifo(filepath.Join(dir, ep.file), 0o644); err != nil {
				t.Skipf("mkfifo unsupported: %v", err)
			}

			type result struct {
				release func() error
				err     error
			}
			done := make(chan result, 1)
			go func() {
				rel, err := ep.acquire(root)
				done <- result{rel, err}
			}()
			select {
			case r := <-done:
				assertRefusedUnsafe(t, r.release, r.err)
			case <-time.After(5 * time.Second):
				t.Fatal("acquisition blocked on a FIFO lock path")
			}
		})
	}
}

// A directory at the lock path is refused.
func TestStateLockDirectoryAtLockPathRefused(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ep.relDir, ep.file), 0o755); err != nil {
				t.Fatal(err)
			}
			release, err := ep.acquire(root)
			assertRefusedUnsafe(t, release, err)
		})
	}
}
