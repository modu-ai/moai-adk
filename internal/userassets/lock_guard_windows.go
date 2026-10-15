//go:build windows

package userassets

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// On Windows, CreateFile-based byte-range locks need a handle dance the
// stdlib does not expose cleanly; the guard degrades to a short-lived
// marker file whose removal the next acquirer retries around. The O_EXCL
// create of the lock itself is already atomic on Windows, so the guard only
// needs to bound the reclaim window, which the retry loop tolerates.
// Item 3 (fix round 3): the handle is returned to its own call (no package
// global); item 4: the marker acquisition respects the same deadline.
//
// M2 (SPEC-USERASSET-DEPLOY-GUARD-001, REQ-LOCK-001/002): the marker
// records its owner's PID — the same ownership posture the .lock file
// carries — and a marker whose owner is PROVEN dead (pid record + death
// check, lockOwnerGone) is reclaimed by the next acquirer. An ownerless
// marker (no pid record) is NEVER auto-reclaimed: it surfaces through the
// user-visible recovery path (doctor report + explicit confirmed removal),
// because a suspended process's marker has exactly this shape and age is
// no proof of death (design §3, gate-measured release-deletes-B repro).

func acquireGuard(path string, timeout time.Duration) (func(), error) {
	marker := path + ".guard"
	serializePath := marker + ".serialize"
	deadline := time.Now().Add(timeout)
	var fd *os.File
	var err error
	for {
		// Gate round 38-1: the WHOLE verify→reclaim→create span is
		// serialized under an OS mutex (LockFileEx on a dedicated sibling)
		// — the byte-compare+restore identity check left a displacement
		// window where a third caller acquired and the stale reclaim
		// overlapped into TWO simultaneous lock holders (the deterministic
		// cross-run repro). Inside the span the dead-marker removal is a
		// plain remove: no other acquirer can interleave.
		ser, serErr := os.OpenFile(serializePath, os.O_CREATE|os.O_RDWR, 0o644)
		if serErr != nil {
			return nil, serErr
		}
		lockErr := lockWindowsExclusive(ser, 200*time.Millisecond)
		if lockErr != nil {
			_ = ser.Close()
			if time.Now().After(deadline) {
				return nil, lockErr
			}
			time.Sleep(5 * time.Millisecond)
			continue
		}
		if guardMarkerDead(marker) {
			_ = os.Remove(marker)
		}
		fd, err = os.OpenFile(marker, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			// REQ-LOCK-001: stamp the caller's PID — the .lock posture.
			_, _ = fmt.Fprintf(fd, "pid=%d acquired=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
			_ = unlockWindows(ser)
			_ = ser.Close()
			break
		}
		_ = unlockWindows(ser)
		_ = ser.Close()
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
	}
	release := func() {
		_ = fd.Close()
		_ = os.Remove(marker)
	}
	return release, nil
}

// lockWindowsExclusive takes an exclusive LockFileEx on the file's handle,
// retrying within the timeout (a held serialization lock releases when its
// holder finishes the span).
func lockWindowsExclusive(f *os.File, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		err := windows.LockFileEx(windows.Handle(f.Fd()),
			windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
			0, 1, 0, &windows.Overlapped{})
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// unlockWindows releases the exclusive LockFileEx.
func unlockWindows(f *os.File) error {
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &windows.Overlapped{})
}

// lockRecord statuses for readLockRecord (mirrored in lock.go).
const (
	recordAbsent = iota
	recordIrregular
	recordOK
)

// readLockRecord reads one ownership record on windows: the unix FIFO
// open-hang hazard has no windows-path equivalent (named pipes live under
// \\.\\pipe\\ spellings a literal lock path never reaches), so the Lstat +
// ReadFile form suffices and the handle-binding refinement stays unix-
// only. The type judgment still fails closed on non-regular files.
func readLockRecord(path string) lockRecord {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return lockRecord{status: recordAbsent}
		}
		return lockRecord{status: recordIrregular}
	}
	if !info.Mode().IsRegular() {
		return lockRecord{status: recordIrregular}
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return lockRecord{status: recordIrregular}
	}
	return lockRecord{data: data, status: recordOK}
}

// classifyPidlessMarker decides a pid-less guard file's state on windows
// (SPEC-USERASSET-DEPLOY-GUARD-001 M2, gate rounds 18/19): on windows the
// marker's EXISTENCE is the held evidence — there is no flock fallback —
// so a pid-less marker is ownerless: never auto-reclaimed, resolved through
// the doctor's visible-recovery row.
func classifyPidlessMarker(markerPath string) (GuardMarkerState, int) {
	return GuardMarkerOwnerless, 0
}
