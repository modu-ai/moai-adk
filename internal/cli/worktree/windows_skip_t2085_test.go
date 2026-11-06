package worktree

// windows_skip_t2085_test.go — card t2085.
//
// Some t2053 tests build their failure from a precondition that does not exist
// on Windows. They fail there for that reason, not because the product is
// wrong, so they skip with the reason stated (the explicit-reason precedent of
// card t2063). Each reason below was read from the test's own setup, not
// inferred from the failure text:
//
//	posix-mode   a chmod 000 / 0o555 path is the unreadable-input fixture; NTFS
//	             ignores POSIX mode bits, so the path stays readable
//	shell-shim   a "#!/bin/sh" script put first on PATH is the fault injector
//	             (blocking git, hanging lsof); Windows cannot execute it, so the
//	             real binary runs and the fault is never injected
//	same-file    os.SameFile on two FileInfo values compares file identity
//	             lazily by re-opening the path on Windows, so a replaced file
//	             always reads as "the same file"

import (
	"runtime"
	"testing"
)

// t2085SkipOnWindows skips the calling test on Windows, naming the reason.
func t2085SkipOnWindows(t *testing.T, reason string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skipf("skipped on Windows: %s", reason)
	}
}

const (
	t2085ReasonPosixMode = "the fixture is an unreadable path made with chmod; Windows ignores POSIX mode bits, so the input stays readable"
	t2085ReasonShim      = "the fault is injected by a #!/bin/sh script placed first on PATH; Windows cannot execute it, so the fault never happens"
	t2085ReasonSameFile  = "os.SameFile on two FileInfo values re-opens the path on Windows, so a replaced file always reads as the same file"
)
