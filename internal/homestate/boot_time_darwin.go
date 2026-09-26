//go:build darwin

package homestate

import (
	"time"

	"golang.org/x/sys/unix"
)

func platformBootTime() (time.Time, bool) {
	tv, err := unix.SysctlTimeval("kern.boottime")
	if err != nil || tv == nil || tv.Sec <= 0 {
		return time.Time{}, false
	}
	return time.Unix(tv.Sec, int64(tv.Usec)*int64(time.Microsecond)), true
}
