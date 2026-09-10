//go:build !windows

// board_lock_errno_test.go — bidirectional contract for the Unix board-lock
// flock(2) failure classification (SPEC-BOARDLOCK-ERRNO-001, card t379).
//
// The pair is load-bearing: the positive direction alone admits an
// always-contention predicate (the pre-repair defect), and the negative
// direction alone admits an always-hard-error predicate (the rule switched
// off). Neither test asserts the other's direction, so a mutant that inverts
// the classification reddens exactly one of them.
package kanban

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// nonContentionErrnos are the synthetic inputs for the negative direction.
// EBADF and EOPNOTSUPP are NOT reachable through the real acquisition path
// (the call site flocks a descriptor a just-succeeded unix.Open returned, and
// passes compile-time-constant flags); they are valid inputs to the
// classification predicate regardless. Unreachable and out-of-scope-for-
// classification are different propositions. EINTR is included deliberately:
// SPEC §1.3.1 records its reclassification from contention-equivalent to hard
// error as an ACCEPTED, UNMEASURED behaviour change, and this row is what
// stops that decision being silently reverted.
var nonContentionErrnos = []unix.Errno{
	unix.ENOLCK,
	unix.EBADF,
	unix.EOPNOTSUPP,
	unix.EINTR,
}

// TestBoardFlockErrnoContentionRemainsHeld covers AC-BLE-001a (REQ-BLE-001).
// It runs the REAL acquisition path — the wiring detector: under the M-narrow
// mutant this reddens only if acquireBoardLockImpl returns the classifier's
// result rather than a hardcoded sentinel.
func TestBoardFlockErrnoContentionRemainsHeld(t *testing.T) {
	root := t.TempDir()

	held, err := AcquireBoardLock(root)
	if err != nil {
		t.Fatalf("first AcquireBoardLock: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })

	second, err := AcquireBoardLock(root)
	if err == nil {
		_ = second.Release()
		t.Fatal("second AcquireBoardLock: expected contention, got nil error")
	}
	if !IsBoardLockHeld(err) {
		t.Fatalf("IsBoardLockHeld(%v) = false, want true", err)
	}
	if !errors.Is(err, ErrBoardLockHeld) {
		t.Fatalf("errors.Is(%v, ErrBoardLockHeld) = false, want true", err)
	}
}

// TestBoardFlockErrnoNonContentionIsNotHeld covers AC-BLE-001b (REQ-BLE-002).
// It feeds synthetic errnos to the classification predicate directly. It
// asserts NOTHING about the contention direction — that belongs to
// TestBoardFlockErrnoContentionRemainsHeld, so the two mutants redden
// different tests.
func TestBoardFlockErrnoNonContentionIsNotHeld(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "board.lock")

	for _, errno := range nonContentionErrnos {
		t.Run(errno.Error(), func(t *testing.T) {
			got := classifyBoardFlockErr(errno, lockPath)
			if got == nil {
				t.Fatal("classifyBoardFlockErr returned nil; the error must not be swallowed")
			}
			if IsBoardLockHeld(got) {
				t.Fatalf("IsBoardLockHeld(%v) = true, want false", got)
			}
		})
	}
}

// TestBoardFlockErrnoPreservesErrnoAndPath covers AC-BLE-002 (REQ-BLE-003).
// Preserving the errno for errors.Is inspection is the substance of this
// SPEC: an implementation that keeps only the errno TEXT in the message while
// breaking errors.Is fails here.
func TestBoardFlockErrnoPreservesErrnoAndPath(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "board.lock")

	for _, errno := range nonContentionErrnos {
		t.Run(errno.Error(), func(t *testing.T) {
			got := classifyBoardFlockErr(errno, lockPath)
			if got == nil {
				t.Fatal("classifyBoardFlockErr returned nil")
			}
			if !errors.Is(got, errno) {
				t.Fatalf("errors.Is(%v, %v) = false, want true", got, errno)
			}
			if !strings.Contains(got.Error(), lockPath) {
				t.Fatalf("error message %q does not name the lock path %q", got.Error(), lockPath)
			}
		})
	}
}

// TestBoardFlockErrnoFailurePathClosesDescriptor covers AC-BLE-003
// (REQ-BLE-004).
//
// Mechanism: POSIX open(2) guarantees "the lowest-numbered file descriptor
// not currently open for the process". A probe descriptor opened before and
// after N failed acquisitions therefore lands on the same number when nothing
// leaked, and on a number ~N higher when every attempt leaked one. This
// replaces the plan's /dev/fd entry-count mechanism, whose availability was
// asserted for darwin and linux without being measured on either and which is
// /proc-dependent on the ubuntu CI runner. The guarantee used here is a
// documented syscall contract, is identical on both platforms, and needs no
// filesystem to be readable.
//
// The check cannot pass vacuously: a probe that fails to open is fatal, and
// the induced-failure count is asserted equal to N, so a sweep that induced
// nothing fails rather than reporting ok.
//
// The single unix.Close(fd) in acquireBoardLockImpl serves BOTH the
// contention and the non-contention return path, so removing it (the M-leak
// mutant, re-laid against the shape actually built) reddens this test even
// though only contention is inducible here.
func TestBoardFlockErrnoFailurePathClosesDescriptor(t *testing.T) {
	root := t.TempDir()

	held, err := AcquireBoardLock(root)
	if err != nil {
		t.Fatalf("first AcquireBoardLock: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })

	probeFD := func(label string) int {
		fd, err := unix.Open(os.DevNull, unix.O_RDONLY|unix.O_CLOEXEC, 0)
		if err != nil {
			t.Fatalf("%s probe: unix.Open(%s): %v", label, os.DevNull, err)
		}
		if cerr := unix.Close(fd); cerr != nil {
			t.Fatalf("%s probe: unix.Close(%d): %v", label, fd, cerr)
		}
		return fd
	}

	// attempts is large enough that a per-attempt leak dwarfs slack.
	const attempts = 200
	// slack absorbs descriptors the Go runtime may open during the loop
	// (netpoll, /dev/urandom). It is 8% of attempts, so a leak on even a
	// tenth of the attempts is still caught.
	const slack = 16

	before := probeFD("before")

	induced := 0
	for i := 0; i < attempts; i++ {
		contender, err := AcquireBoardLock(root)
		if err == nil {
			_ = contender.Release()
			t.Fatalf("attempt %d: expected contention, got nil error", i)
		}
		if !IsBoardLockHeld(err) {
			t.Fatalf("attempt %d: expected contention sentinel, got %v", i, err)
		}
		induced++
	}
	if induced != attempts {
		t.Fatalf("induced %d failed acquisitions, want %d — an empty sweep asserts nothing", induced, attempts)
	}

	after := probeFD("after")

	if after > before+slack {
		t.Fatalf("descriptor leak: probe fd %d before, %d after %d failed acquisitions (slack %d)",
			before, after, attempts, slack)
	}
}
