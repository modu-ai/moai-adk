//go:build windows

package factory

import (
	"fmt"
	"os"
)

// openCandidateRecordFile opens a candidate record and verifies the OPENED
// file is regular before any read (card t1478 M2 repair — the Windows half
// of the TOCTOU tail). The O_NONBLOCK|O_NOFOLLOW form is unix-only; on
// Windows the stat-after-open check is the achievable equivalent — a
// non-regular file system object (a named pipe) is refused on the opened
// handle, closing the same verify-what-you-read gap.
//
// @MX:ANCHOR: [AUTO] windows record opener: the stat-after-open check stands in for the unix open flags.
// @MX:REASON: Windows has no O_NONBLOCK or O_NOFOLLOW equivalent on this path, so refusing a non-regular handle after open is what keeps a named pipe from parking the candidate lock.
// @MX:SPEC: SPEC-CANDIDATE-CI-001
func openCandidateRecordFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("candidate store: stat %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, fmt.Errorf("candidate record at %s is not a regular file — refusing to open it", path)
	}
	return f, nil
}
