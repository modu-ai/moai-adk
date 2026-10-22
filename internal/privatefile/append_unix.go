//go:build unix

package privatefile

import (
	"os"
	"syscall"
)

func openAppendNative(path string) (*os.File, error) {
	// NOFOLLOW closes the link-swap window; NONBLOCK prevents a raced-in FIFO
	// from parking the open. Neither changes regular-file append semantics.
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
}
