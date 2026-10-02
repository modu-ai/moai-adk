//go:build !windows

// state_lock_errno_test.go — bidirectional contract for the Unix state-lock
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

// TestStateFlockErrnoContentionRemainsHeld covers AC-BLE-001a (REQ-BLE-001).
// It runs the REAL acquisition path — the wiring detector: under the M-narrow
// mutant this reddens only if acquireStateLockImpl returns the classifier's
// result rather than a hardcoded sentinel.
func TestStateFlockErrnoContentionRemainsHeld(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "state.lock")

	held, err := acquireStateLockImpl(lockPath)
	if err != nil {
		t.Fatalf("first acquireStateLockImpl: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = held.release() })

	second, err := acquireStateLockImpl(lockPath)
	if err == nil {
		_ = second.release()
		t.Fatal("second acquireStateLockImpl: expected contention, got nil error")
	}
	if !IsStateLockHeld(err) {
		t.Fatalf("IsStateLockHeld(%v) = false, want true", err)
	}
	if !errors.Is(err, ErrStateLockHeld) {
		t.Fatalf("errors.Is(%v, ErrStateLockHeld) = false, want true", err)
	}
}

// TestStateFlockErrnoNonContentionIsNotHeld covers AC-BLE-001b (REQ-BLE-002).
// It feeds synthetic errnos to the classification predicate directly. It
// asserts NOTHING about the contention direction — that belongs to
// TestStateFlockErrnoContentionRemainsHeld, so the two mutants redden
// different tests.
func TestStateFlockErrnoNonContentionIsNotHeld(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "state.lock")

	for _, errno := range nonContentionErrnos {
		t.Run(errno.Error(), func(t *testing.T) {
			got := classifyStateFlockErr(errno, lockPath)
			if got == nil {
				t.Fatal("classifyStateFlockErr returned nil; the error must not be swallowed")
			}
			if IsStateLockHeld(got) {
				t.Fatalf("IsStateLockHeld(%v) = true, want false", got)
			}
		})
	}
}

// TestStateFlockErrnoPreservesErrnoAndPath covers AC-BLE-002 (REQ-BLE-003).
// Preserving the errno for errors.Is inspection is the substance of this
// SPEC: an implementation that keeps only the errno TEXT in the message while
// breaking errors.Is fails here.
func TestStateFlockErrnoPreservesErrnoAndPath(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "state.lock")

	for _, errno := range nonContentionErrnos {
		t.Run(errno.Error(), func(t *testing.T) {
			got := classifyStateFlockErr(errno, lockPath)
			if got == nil {
				t.Fatal("classifyStateFlockErr returned nil")
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

// TestStateFlockErrnoFailurePathClosesDescriptor covers AC-BLE-003
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
// The single unix.Close(fd) in acquireStateLockImpl serves BOTH the
// contention and the non-contention return path, so removing it (the M-leak
// mutant, re-laid against the shape actually built) reddens this test even
// though only contention is inducible here.
func TestStateFlockErrnoFailurePathClosesDescriptor(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "state.lock")

	held, err := acquireStateLockImpl(lockPath)
	if err != nil {
		t.Fatalf("first acquireStateLockImpl: unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = held.release() })

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
		contender, err := acquireStateLockImpl(lockPath)
		if err == nil {
			_ = contender.release()
			t.Fatalf("attempt %d: expected contention, got nil error", i)
		}
		if !IsStateLockHeld(err) {
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
