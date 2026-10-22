package toolpolicy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// acquireProjectRecordLock keeps the original exclusive primary record and
// retry/stale policy. A separate persistent kernel guard serializes all primary
// create, observed-instance removal and release mutations across processes.
func acquireProjectRecordLock(settings string) (release func(), result error) {
	path := managedAllowLockPath(settings)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create record lock dir: %w", err)
	}
	deadline := time.Now().Add(recordLockRetryBudget)
	g, err := openRecordGuard(path + ".guard")
	if err != nil {
		return nil, fmt.Errorf("open persistent record guard: %w", err)
	}
	transferred := false
	defer func() {
		if !transferred {
			result = errors.Join(result, g.file.Close())
		}
	}()
	for {
		if err := g.lockUntil(deadline); err != nil {
			return nil, fmt.Errorf("acquire persistent record guard: %w", err)
		}
		f, createErr := createGuardedRecord(g, path)
		unlockErr := unlockRecordGuard(g.file)
		if unlockErr != nil {
			if f != nil {
				unlockErr = errors.Join(unlockErr, f.Close())
			}
			return nil, errors.Join(createErr, unlockErr)
		}
		if createErr == nil {
			transferred = true
			return releaseGuardedRecord(g, f, path), nil
		}
		if !os.IsExist(createErr) {
			return nil, fmt.Errorf("create record lock %s: %w", path, createErr)
		}
		if !time.Now().Before(deadline) {
			if err := takeoverObservedRecord(g, path, nil, deadline); err == nil {
				continue
			} else {
				return nil, fmt.Errorf("record lock %s held by another process (retry budget exhausted; holder alive, lock fresh): %w", path, err)
			}
		}
		time.Sleep(recordLockRetryInterval)
	}
}

func createGuardedRecord(g *recordGuard, path string) (*os.File, error) {
	f, err := createRecordPrimary(path)
	if err != nil {
		return nil, err
	}
	_, stampErr := fmt.Fprintf(f, "pid=%d guard=%s created=%s\n", os.Getpid(), g.id, time.Now().UTC().Format(time.RFC3339Nano))
	if err := errors.Join(stampErr, g.verify()); err != nil {
		info, statErr := f.Stat()
		if statErr == nil {
			statErr = removeObservedRecord(g, path, info)
		}
		return nil, errors.Join(err, statErr, f.Close())
	}
	return f, nil
}

func releaseGuardedRecord(g *recordGuard, f *os.File, path string) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			err := g.lockUntil(time.Now().Add(recordLockRetryBudget))
			if err == nil {
				info, statErr := f.Stat()
				if statErr == nil {
					statErr = removeObservedRecord(g, path, info)
				}
				err = errors.Join(statErr, unlockRecordGuard(g.file))
			}
			err = errors.Join(err, f.Close(), g.file.Close())
			if err != nil {
				slog.Warn("tool-policy record lock release refused or failed", "path", path, "error", err)
			}
		})
	}
}

// removeObservedRecord is called only while the persistent guard is held. All
// cooperating creators/releases share that guard, so the compare/remove span
// cannot dispose of another cooperating writer's newly created record.
func removeObservedRecord(g *recordGuard, path string, observed os.FileInfo) (result error) {
	if err := g.verify(); err != nil {
		return err
	}
	current, err := openRecordObservation(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, current.Close()) }()
	info, err := current.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(observed, info) {
		return errors.New("record lock was replaced; preserving successor")
	}
	body, err := io.ReadAll(current)
	if err != nil {
		return err
	}
	if err := checkRecordGuardBinding(body, g.id); err != nil {
		return err
	}
	return os.Remove(path)
}

func takeoverStaleRecordLock(path string) error {
	return takeoverStaleRecordLockObserved(path, nil)
}

// observed is per-call, after the held descriptor's actual stale observation
// and before serialization. Tests can pause two observers without global hooks.
func takeoverStaleRecordLockObserved(path string, observed func()) (result error) {
	g, err := openRecordGuard(path + ".guard")
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, g.file.Close()) }()
	return takeoverObservedRecord(g, path, observed, time.Now().Add(recordLockRetryBudget))
}

func takeoverObservedRecord(g *recordGuard, path string, observed func(), deadline time.Time) (result error) {
	f, err := openRecordObservation(path)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, f.Close()) }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	body, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err := checkRecordGuardBinding(body, g.id); err != nil {
		return err
	}
	if time.Since(info.ModTime()) < recordLockStaleAge {
		pid, ok := parseRecordLockPid(body)
		if !ok || !pidProvablyDead(pid) {
			return errors.New("lockfile holder not provably dead and lockfile age under the stale bound")
		}
	}
	if observed != nil {
		observed()
	}
	if err := g.lockUntil(deadline); err != nil {
		return err
	}
	defer func() { result = errors.Join(result, unlockRecordGuard(g.file)) }()
	return removeObservedRecord(g, path, info)
}
