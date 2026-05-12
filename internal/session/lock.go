//go:build !windows

package session

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// fileLock는 cross-platform advisory lock 인터페이스입니다.
// SPEC-V3R2-RT-004 REQ-040: concurrent checkpoint write 방지를 위한 lock primitive.
// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-004 REQ-040 cross-platform lock contract
// @MX:REASON: lock_unix.go and lock_windows.go are the two implementations. New platforms add a new file with the same interface.
type fileLock interface {
	acquire(path string, retries int, backoff time.Duration) error
	release() error
}

// newFileLock는 platform-specific lock 구현을 반환합니다.
// SPEC-V3R2-RT-004 plan.md §3.2: Unix flock 사용 (Windows는 이후 구현).
func newFileLock() fileLock {
	return &flockLock{}
}

// flockLock는 Unix flock advisory lock 구현입니다.
// SPEC-V3R2-RT-004 REQ-040: syscall.Flock(LOCK_EX|LOCK_NB) 사용.
type flockLock struct {
	mu     sync.Mutex
	fd     int
	active bool
}

// acquire는 flock(LOCK_EX|LOCK_NB)로 exclusive lock을 획득합니다.
// Lock companion file path: <checkpoint-path>.lock
func (l *flockLock) acquire(path string, retries int, backoff time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Lock companion 파일 경로
	lockPath := path + ".lock"

	// 파일 열기 (생성 포함)
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0644)
	if err != nil {
		return fmt.Errorf("open lock file %s: %w", lockPath, err)
	}

	l.fd = fd

	// Non-blocking exclusive lock 시도
	err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		_ = unix.Close(fd)
		l.fd = 0
		return fmt.Errorf("flock %s: %w", lockPath, err)
	}

	l.active = true
	return nil
}

// release는 flock을 해제하고 lock 파일을 닫습니다.
func (l *flockLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.active {
		return nil // 이미 해제됨
	}

	if l.fd != 0 {
		// Flock unlock (implicit on close)
		err := unix.Close(l.fd)
		l.fd = 0
		l.active = false
		return err
	}

	return nil
}

// SPEC-V3R2-RT-004: acquireWithRetry는 3-retry / 10ms-backoff 정책으로 lock을 획득합니다.
// REQ-040: repeated loss 시 ErrCheckpointConcurrent 반환.
func acquireWithRetry(lock fileLock, path string, retries int, backoff time.Duration) error {
	var lastErr error
	for i := 0; i < retries; i++ {
		err := lock.acquire(path, retries, backoff)
		if err == nil {
			return nil // 성공
		}
		lastErr = err
		// 백오프 대기
		time.Sleep(backoff)
	}
	return fmt.Errorf("%w: after %d retries: %v", ErrCheckpointConcurrent, retries, lastErr)
}
