//go:build !windows

package migration

import "golang.org/x/sys/unix"

// lockHandle은 Unix advisory lock의 파일 디스크립터를 보유합니다.
type lockHandle struct{ fd int }

// acquireLock은 lockPath에 대한 exclusive advisory lock을 획득합니다.
// unix.Flock(LOCK_EX)를 사용하므로 다른 프로세스가 lock을 보유 중이면
// 해제될 때까지 blocking됩니다 (REQ-V3R2-RT-007-031).
func acquireLock(lockPath string) (*lockHandle, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}

	return &lockHandle{fd: fd}, nil
}

// releaseLock은 advisory lock을 해제하고 파일 디스크립터를 닫습니다.
func releaseLock(h *lockHandle) error {
	if err := unix.Flock(h.fd, unix.LOCK_UN); err != nil {
		_ = unix.Close(h.fd)
		return err
	}
	return unix.Close(h.fd)
}
