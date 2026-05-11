//go:build windows

package migration

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// lockHandle은 Windows file mutex의 lock 파일 경로를 보유합니다.
// Windows에는 unix.Flock에 해당하는 advisory lock이 없으므로
// O_CREATE|O_EXCL (atomic create-or-fail) 기반 file mutex로 대체합니다.
// REQ-V3R2-RT-007-031 advisory lock 의미론은 단일 사용자 가정 하에 단순화됩니다.
type lockHandle struct{ path string }

const (
	lockMaxRetries = 100             // 최대 재시도 횟수 (~1s @ 10ms/retry)
	lockRetryDelay = 10 * time.Millisecond
)

// acquireLock은 O_CREATE|O_EXCL 으로 lock 파일을 생성해 mutex를 획득합니다.
// 파일이 이미 존재하면 10ms 대기 후 재시도하며, 100회(~1s) 후 timeout합니다.
func acquireLock(lockPath string) (*lockHandle, error) {
	for i := 0; i < lockMaxRetries; i++ {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			_ = f.Close()
			return &lockHandle{path: lockPath}, nil
		}
		if errors.Is(err, fs.ErrExist) {
			time.Sleep(lockRetryDelay)
			continue
		}
		return nil, err
	}
	return nil, errors.New("lock acquire timeout (windows)")
}

// releaseLock은 lock 파일을 삭제해 mutex를 해제합니다.
// 파일이 이미 존재하지 않으면 성공으로 처리합니다 (idempotent).
func releaseLock(h *lockHandle) error {
	err := os.Remove(h.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
