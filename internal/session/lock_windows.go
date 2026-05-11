//go:build windows

package session

import (
	"sync"
	"time"
)

// fileLock는 cross-platform advisory lock 인터페이스입니다 (Windows variant).
// SPEC-V3R2-RT-004 REQ-040: concurrent checkpoint write 방지를 위한 lock primitive.
type fileLock interface {
	acquire(path string, retries int, backoff time.Duration) error
	release() error
}

// newFileLock는 platform-specific lock 구현을 반환합니다.
// Windows에서는 mutex-only same-process fallback (cross-process는 추후 LockFileEx로 대체).
// SPEC-V3R2-RT-004 WIP: native Windows advisory lock은 M4-M5 또는 별도 SPEC.
func newFileLock() fileLock {
	return &mutexLock{}
}

// mutexLock는 Windows에서 사용하는 process-level mutex-only fallback입니다.
// SPEC-V3R2-RT-004 WIP M3 단계: 동일 프로세스 내 goroutine 간 race만 직렬화.
// cross-process race는 본 단계에서 보장하지 않으며, M4-M5 또는 후속 SPEC에서
// LockFileEx 기반 native 구현으로 교체 예정.
type mutexLock struct {
	mu     sync.Mutex
	locked bool
}

// acquire는 same-process 차원의 exclusive lock을 즉시 획득합니다.
// retries / backoff 인자는 Unix flock과의 인터페이스 일치를 위해 보존하지만 사용하지 않습니다.
func (l *mutexLock) acquire(path string, retries int, backoff time.Duration) error {
	l.mu.Lock()
	l.locked = true
	return nil
}

// release는 mutex를 해제합니다.
func (l *mutexLock) release() error {
	if !l.locked {
		return nil
	}
	l.locked = false
	l.mu.Unlock()
	return nil
}

// acquireWithRetry는 Windows fallback에서 mutex 즉시 획득으로 위임합니다.
// Unix 버전의 retry 정책 (3-retry / 10ms-backoff)은 native 구현이 도입되는 시점에 적용 예정.
func acquireWithRetry(lock fileLock, path string, retries int, backoff time.Duration) error {
	return lock.acquire(path, retries, backoff)
}
