//go:build !windows

package userassets

import (
	"io"
	"os"
	"syscall"
	"time"
)

// The guard is a SECOND file held with flock(2) for the entire acquire
// attempt, so the reclaim-then-create sequence runs strictly serially per
// user (review-fix round 2 F5/A3/B2). Item 3 (fix round 3): the handle is
// returned to ITS OWN call — the former package-global was overwritten by
// concurrent calls from different HOMEs, releasing another call's guard and
// leaking the original (a real data race under -race).

func acquireGuard(path string, timeout time.Duration) (func(), error) {
	// Item 4: ONE deadline covers the whole acquire — the flock is
	// NON-BLOCKING with a retry loop bounded by the same deadline that
	// bounds the O_EXCL retries; a held guard can no longer overshoot it.
	deadline := time.Now().Add(timeout)
	fd, err := os.OpenFile(path+".guard", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	for {
		err = syscall.Flock(int(fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = fd.Close()
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
	}
	release := func() {
		_ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN)
		_ = fd.Close()
	}
	return release, nil
}

// classifyPidlessMarker decides a pid-less guard file's state on unix
// (SPEC-USERASSET-DEPLOY-GUARD-001 M2, gate rounds 18/19): the flock IS the
// truth here — the guard file survives a clean release by design, so a
// free file is a clean leftover with no recovery meaning (absent, to the
// doctor), while a flock-held file belongs to a live process. A probe
// error fails closed to held.
func classifyPidlessMarker(markerPath string) (GuardMarkerState, int) {
	// Gate round 51-4: the SECOND open carries the same discipline as
	// readLockRecord's first — O_NONBLOCK (a FIFO swap before this open
	// must not park on the writer wait) and the type check BOUND TO THIS
	// HANDLE. A non-regular handle fails closed to held.
	fd, err := os.OpenFile(markerPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return GuardMarkerOwnerAlive, 0
	}
	defer func() { _ = fd.Close() }()
	var st syscall.Stat_t
	if fstatErr := syscall.Fstat(int(fd.Fd()), &st); fstatErr != nil || st.Mode&syscall.S_IFMT != syscall.S_IFREG {
		return GuardMarkerOwnerAlive, 0 // not a regular file — fail closed
	}
	err = syscall.Flock(int(fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return GuardMarkerOwnerAlive, 0 // held by a live process
	}
	_ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN)
	return GuardMarkerAbsent, 0 // a cleanly-released leftover, not a recovery target
}

// lockRecord statuses for readLockRecord (mirrored in lock.go).
const (
	recordAbsent = iota
	recordIrregular
	recordOK
)

// readLockRecord reads one ownership record with the TYPE CHECK BOUND TO
// THE OPEN HANDLE (gate round 23 #1): the file is opened NON-BLOCKING (an
// O_NONBLOCK open of a FIFO returns immediately instead of parking on the
// writer wait), the handle is fstat'd — a regular file only — and the
// bytes are read from the same handle. A path swapped to a FIFO after any
// earlier check therefore surfaces as recordIrregular instead of hanging
// the read, and the bytes read are the inode the type was verified on.
func readLockRecord(path string) lockRecord {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return lockRecord{status: recordAbsent}
		}
		return lockRecord{status: recordIrregular}
	}
	defer func() { _ = f.Close() }()
	info, statErr := f.Stat()
	if statErr != nil || !info.Mode().IsRegular() {
		return lockRecord{status: recordIrregular}
	}
	data, readErr := io.ReadAll(f)
	if readErr != nil {
		return lockRecord{status: recordIrregular}
	}
	return lockRecord{data: data, status: recordOK}
}
