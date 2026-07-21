//go:build windows

package atomicfile

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// Retry budget for a transiently-blocked replace. Mirrors the ~1s bounded
// budget already used by the migration version lock (internal/migration/
// version_windows.go): 100 attempts at 10ms.
const (
	replaceMaxAttempts = 100
	replaceRetryDelay  = 10 * time.Millisecond
)

// replace renames oldpath onto newpath with a bounded retry.
//
// Unlike POSIX rename(2), Windows MoveFileEx fails when any process still
// holds an open handle on the destination — even with MOVEFILE_REPLACE_EXISTING
// — reporting ERROR_ACCESS_DENIED or ERROR_SHARING_VIOLATION. The common
// source is a concurrent reader (os.ReadFile opens without FILE_SHARE_DELETE)
// or a virus scanner touching the freshly written file; both hold the handle
// only transiently, so a bounded retry converges.
//
// A caller that holds the destination open itself is NOT recoverable here —
// that is a lock-design defect and must use a sidecar .lock file instead
// (see internal/harness/tier, internal/session, internal/migration).
//
// Non-transient errors return immediately so callers still observe a genuine
// failure (missing source, cross-device link) without paying the retry budget.
func replace(oldpath, newpath string) error {
	var err error
	for attempt := 0; attempt < replaceMaxAttempts; attempt++ {
		err = os.Rename(oldpath, newpath)
		if err == nil || !isTransientReplaceError(err) {
			return err
		}
		time.Sleep(replaceRetryDelay)
	}
	return err
}

// isTransientReplaceError reports whether err is a Windows sharing conflict
// that a later retry may clear.
func isTransientReplaceError(err error) bool {
	return errors.Is(err, windows.ERROR_ACCESS_DENIED) ||
		errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}
