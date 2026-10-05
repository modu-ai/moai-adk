//go:build !windows

package decision

import "golang.org/x/sys/unix"

// lockBoard takes an exclusive advisory lock on the board's companion lock
// file and returns its release.
func lockBoard(path string) (func(), error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	return func() { _ = unix.Close(fd) }, nil
}
