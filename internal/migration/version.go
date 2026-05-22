package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// @MX:NOTE - SPEC-V3R2-RT-007 .moai/state/migration-version is the single source of truth for "applied migrations".
// Protected by atomic rename + advisory lock. Version absent = 0 (fresh install or first encounter from v2.x).
// External callers outside this package must not write to the version-file.

const versionFileName = "migration-version"
const versionTmpFileName = "migration-version.tmp"

// readVersion reads the current migration version.
// REQ-V3R2-RT-007-030: returns 0 if the version-file is absent.
func readVersion(projectRoot string) (int, error) {
	versionFile := filepath.Join(projectRoot, ".moai", "state", versionFileName)

	data, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File absent = 0 (fresh install or first v2.x encounter).
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

// writeVersion records the migration version.
// REQ-V3R2-RT-007-013: uses atomic write (*.tmp + os.Rename).
// REQ-V3R2-RT-007-031: protects concurrent writes via an advisory lock.
//
// Lock semantics:
//   - Unix: unix.Flock(LOCK_EX) — blocking until acquired (kernel serializes contention).
//   - Windows: file mutex based on O_CREATE|O_EXCL — bounded retry up to ~1s.
//     Under the single-user assumption, a lock acquire timeout indicates an abnormal state.
func writeVersion(projectRoot string, version int) error {
	stateDir := filepath.Join(projectRoot, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("state 디렉터리 생성 실패: %w", err)
	}

	versionTmpFile := filepath.Join(stateDir, versionTmpFileName)
	versionFile := filepath.Join(stateDir, versionFileName)

	// Acquire advisory lock (platform-specific impl: version_unix.go / version_windows.go).
	lockPath := filepath.Join(stateDir, versionFileName+".lock")
	handle, err := acquireLock(lockPath)
	if err != nil {
		return fmt.Errorf("lock 획득 실패: %w", err)
	}
	defer func() { _ = releaseLock(handle) }()

	// Write the version to the temporary file.
	if err := os.WriteFile(versionTmpFile, []byte(strconv.Itoa(version)), 0644); err != nil {
		return fmt.Errorf("version-tmp 파일 쓰기 실패: %w", err)
	}

	// Atomic rename (REQ-V3R2-RT-007-013).
	if err := os.Rename(versionTmpFile, versionFile); err != nil {
		return fmt.Errorf("version-file atomic rename 실패: %w", err)
	}

	return nil
}

// detectInFlightState detects an in-flight state (presence of a *.tmp file).
// REQ-V3R2-RT-007-031: recovers a partial write on restart after a crash.
func detectInFlightState(projectRoot string) bool {
	stateDir := filepath.Join(projectRoot, ".moai", "state")
	versionTmpFile := filepath.Join(stateDir, versionTmpFileName)

	_, err := os.Stat(versionTmpFile)
	return err == nil
}

// cleanupInFlightState clears the in-flight state.
// REQ-V3R2-RT-007-031: removes the temporary file before reapplying.
func cleanupInFlightState(projectRoot string) error {
	stateDir := filepath.Join(projectRoot, ".moai", "state")
	versionTmpFile := filepath.Join(stateDir, versionTmpFileName)

	if err := os.Remove(versionTmpFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("in-flight state 정리 실패: %w", err)
	}

	return nil
}
