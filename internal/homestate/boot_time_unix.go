//go:build !windows && !darwin

package homestate

import (
	"os"
	"time"
)

// platformBootTime reads the kernel's `btime` record from /proc/stat through
// the build-tag-free procStatBootTime seam. A host without procfs, a failed
// read, or a missing or malformed record reports no boot time, which disables
// the boot proof there; the seam, not this reader, carries the cause.
func platformBootTime() (time.Time, bool) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = f.Close() }()
	boot, err := procStatBootTime(f)
	if err != nil {
		return time.Time{}, false
	}
	return boot, true
}
