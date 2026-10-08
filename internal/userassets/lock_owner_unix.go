//go:build !windows

package userassets

import (
	"errors"
	"golang.org/x/sys/unix"
)

func lockProcessGone(pid int) bool {
	return errors.Is(unix.Kill(pid, 0), unix.ESRCH)
}
