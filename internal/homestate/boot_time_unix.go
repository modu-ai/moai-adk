//go:build !windows && !darwin

package homestate

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// platformBootTime reads the kernel's `btime` line from /proc/stat. A host
// without procfs reports no boot time, which disables the boot proof there.
func platformBootTime() (time.Time, bool) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && fields[0] == "btime" {
			sec, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil || sec <= 0 {
				return time.Time{}, false
			}
			return time.Unix(sec, 0), true
		}
	}
	return time.Time{}, false
}
