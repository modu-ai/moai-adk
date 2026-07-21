//go:build windows

package atomicfile

import (
	"os"
	"time"
)

// Retry budget for a transiently-blocked read. Deliberately smaller than the
// replace budget: a blocked replace must eventually win or the write is lost,
// whereas a blocked read has a caller waiting on it, and the window it absorbs
// (one concurrent rename) is short.
const (
	readMaxAttempts = 20
	readRetryDelay  = 5 * time.Millisecond
)

// readFile reads path, retrying while Windows reports a transient sharing
// conflict.
//
// Opening a file for reading fails with ERROR_ACCESS_DENIED or
// ERROR_SHARING_VIOLATION while another process is replacing it, because the
// destination sits briefly in a delete-pending state that MoveFileEx creates.
// This is the mirror image of the conflict Replace absorbs: there the reader
// blocks the writer, here the writer blocks the reader. Both are transient, so
// the same bounded-retry policy converges.
//
// A missing file is NOT transient and returns immediately, so callers keep
// matching on os.IsNotExist as with os.ReadFile.
func readFile(path string) ([]byte, error) {
	var (
		data []byte
		err  error
	)
	for attempt := 0; attempt < readMaxAttempts; attempt++ {
		data, err = os.ReadFile(path)
		if err == nil || !isTransientReplaceError(err) {
			return data, err
		}
		time.Sleep(readRetryDelay)
	}
	return data, err
}
