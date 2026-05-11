package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// @MX:NOTE - SPEC-V3R2-RT-007 .moai/state/migration-version은 "적용된 마이그레이션"의 단일 진실 소스입니다.
// Atomic rename + advisory lock로 보호됩니다. Version 부재 = 0 (신규 설치 또는 v2.x 첫 만남).
// 이 패키지 외부에서 version-file을 쓰지 않도록 해야 합니다.

const versionFileName = "migration-version"
const versionTmpFileName = "migration-version.tmp"

// readVersion은 현재 마이그레이션 버전을 읽습니다.
// REQ-V3R2-RT-007-030: version-file이 부재하면 0을 반환합니다.
func readVersion(projectRoot string) (int, error) {
	versionFile := filepath.Join(projectRoot, ".moai", "state", versionFileName)

	data, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일 부재 = 0 (신규 설치 또는 v2.x 첫 만남)
			return 0, nil
		}
		return 0, fmt.Errorf("version-file 읽기 실패: %w", err)
	}

	version, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("version-file 파싱 실패: %w", err)
	}

	return version, nil
}

// writeVersion은 마이그레이션 버전을 기록합니다.
// REQ-V3R2-RT-007-013: atomic write (*.tmp + os.Rename)를 사용합니다.
// REQ-V3R2-RT-007-031: advisory lock으로 동시 쓰기를 보호합니다.
//
// Lock semantics:
//   - Unix: unix.Flock(LOCK_EX) — blocking until acquired (커널이 경쟁을 직렬화함).
//   - Windows: O_CREATE|O_EXCL 기반 file mutex — bounded retry 최대 ~1s.
//     단일 사용자 가정 하에 lock acquire timeout은 비정상 상태를 의미합니다.
func writeVersion(projectRoot string, version int) error {
	stateDir := filepath.Join(projectRoot, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("state 디렉터리 생성 실패: %w", err)
	}

	versionTmpFile := filepath.Join(stateDir, versionTmpFileName)
	versionFile := filepath.Join(stateDir, versionFileName)

	// Advisory lock 획득 (플랫폼별 구현: version_unix.go / version_windows.go)
	lockPath := filepath.Join(stateDir, versionFileName+".lock")
	handle, err := acquireLock(lockPath)
	if err != nil {
		return fmt.Errorf("lock 획득 실패: %w", err)
	}
	defer func() { _ = releaseLock(handle) }()

	// 임시 파일에 버전 기록
	if err := os.WriteFile(versionTmpFile, []byte(strconv.Itoa(version)), 0644); err != nil {
		return fmt.Errorf("version-tmp 파일 쓰기 실패: %w", err)
	}

	// Atomic rename (REQ-V3R2-RT-007-013)
	if err := os.Rename(versionTmpFile, versionFile); err != nil {
		return fmt.Errorf("version-file atomic rename 실패: %w", err)
	}

	return nil
}

// detectInFlightState는 in-flight 상태 (*.tmp 파일 존재)를 감지합니다.
// REQ-V3R2-RT-007-031: crash 후 재시작 시 partial write를 복구합니다.
func detectInFlightState(projectRoot string) bool {
	stateDir := filepath.Join(projectRoot, ".moai", "state")
	versionTmpFile := filepath.Join(stateDir, versionTmpFileName)

	_, err := os.Stat(versionTmpFile)
	return err == nil
}

// cleanupInFlightState는 in-flight 상태를 정리합니다.
// REQ-V3R2-RT-007-031: 재적용 전에 임시 파일을 제거합니다.
func cleanupInFlightState(projectRoot string) error {
	stateDir := filepath.Join(projectRoot, ".moai", "state")
	versionTmpFile := filepath.Join(stateDir, versionTmpFileName)

	if err := os.Remove(versionTmpFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("in-flight state 정리 실패: %w", err)
	}

	return nil
}
