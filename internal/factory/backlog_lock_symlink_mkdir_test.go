//go:build !windows

// backlog_lock_symlink_mkdir_test.go — the queue-lock entry (BacklogStore.Mutate
// and WithLock, both through acquireLock) must refuse a symlinked `.moai` or
// `.moai/state` BEFORE creating the queue directory: a bare os.MkdirAll
// follows the link and creates the queue directory OUTSIDE the project (card
// t1458, P2 follow-up after develop retired the board-lock entry point).
package factory

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type queueEntryPoint struct {
	name string
	run  func(root string) error
}

// queueEntryPoints drives the real queue mutation entries on a store rooted at
// root, using the real layout `<root>/.moai/state/backlog.json`.
func queueEntryPoints() []queueEntryPoint {
	storeAt := func(root string) *BacklogStore {
		return NewBacklogStore(filepath.Join(root, ".moai", "state", backlogFileName))
	}
	return []queueEntryPoint{
		{"Mutate", func(root string) error {
			return storeAt(root).Mutate(func(*BacklogRecord) error { return nil })
		}},
		{"WithLock", func(root string) error {
			return storeAt(root).WithLock(func(*LockedBacklog) error { return nil })
		}},
	}
}

func assertQueueRefused(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("queue mutation succeeded; want refusal of the symlinked path")
	}
	if !errors.Is(err, ErrStateLockUnsafePath) {
		t.Fatalf("error %v does not wrap ErrStateLockUnsafePath", err)
	}
}

func TestQueueLockSymlinkedMoaiCreatesNothing(t *testing.T) {
	for _, ep := range queueEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			external := t.TempDir()
			if err := os.Symlink(external, filepath.Join(root, ".moai")); err != nil {
				t.Fatal(err)
			}
			err := ep.run(root)
			assertDirEmpty(t, external)
			assertQueueRefused(t, err)
		})
	}
}

func TestQueueLockSymlinkedStateDirCreatesNothing(t *testing.T) {
	for _, ep := range queueEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			external := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, filepath.Join(root, ".moai", "state")); err != nil {
				t.Fatal(err)
			}
			err := ep.run(root)
			assertDirEmpty(t, external)
			assertQueueRefused(t, err)
		})
	}
}

// Positive control: a normal root still mutates and creates its directories.
func TestQueueLockNormalRootStillCreatesItsDirs(t *testing.T) {
	for _, ep := range queueEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			if err := ep.run(root); err != nil {
				t.Fatalf("normal queue mutation failed: %v", err)
			}
			info, err := os.Stat(filepath.Join(root, ".moai", "state"))
			if err != nil || !info.IsDir() {
				t.Fatalf("queue dir not created: %v", err)
			}
		})
	}
}
