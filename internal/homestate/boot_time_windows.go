//go:build windows

package homestate

import (
	"time"

	"golang.org/x/sys/windows"
)

func platformBootTime() (time.Time, bool) {
	up := windows.DurationSinceBoot()
	if up <= 0 {
		return time.Time{}, false
	}
	return time.Now().Add(-up), true
}
