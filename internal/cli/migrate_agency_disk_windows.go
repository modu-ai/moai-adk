//go:build windows

// migrate_agency_disk_windows.go: Windows 디스크 공간 확인 로직.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-011
package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// availableDiskBytes는 지정된 경로가 위치한 파일시스템의 가용 디스크 공간(바이트)을 반환한다.
// GetDiskFreeSpaceExW Win32 API를 사용한다.
func availableDiskBytes(path string) (uint64, error) {
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.MustFindProc("GetDiskFreeSpaceExW")

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("utf16ptr %s: %w", path, err)
	}

	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	ret, _, callErr := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	if ret == 0 {
		return 0, fmt.Errorf("GetDiskFreeSpaceExW: %w", callErr)
	}
	return freeBytesAvailable, nil
}

// dirSizeBytes는 지정된 디렉터리 하위 모든 파일의 총 크기(바이트)를 반환한다.
// 심볼릭 링크는 건너뛴다.
//
// @MX:NOTE: [AUTO] 마이그레이션 사전 검사 전용 헬퍼 — 디스크 공간 계산에만 사용
func dirSizeBytes(root string) (uint64, error) {
	var total uint64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return nil //nolint:nilerr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		//nolint:gosec
		total += uint64(info.Size())
		return nil
	})
	return total, err
}

// checkDiskSpaceFn은 테스트에서 checkDiskSpace를 주입하기 위한 함수 변수다.
//
// @MX:ANCHOR: [AUTO] 디스크 공간 사전 검사 진입점 — 테스트 모킹 지원
// @MX:REASON: [AUTO] runFull, 테스트(복수), 향후 resume 경로에서 호출; fan_in >= 3
var checkDiskSpaceFn = checkDiskSpace

// checkDiskSpace는 sourcePath 하위 파일 총 크기의 2배 이상의 디스크 공간이
// sourcePath가 위치한 파일시스템에 가용한지 검증한다.
// 가용 공간이 부족하면 ErrMigrateDiskFull 코드를 포함한 *MigrateError를 반환한다.
//
// REQ-MIGRATE-011: 최소 .agency/ 크기의 2배 확보 필요.
func checkDiskSpace(sourcePath string) error {
	sourceSize, err := dirSizeBytes(sourcePath)
	if err != nil {
		return nil
	}

	available, err := availableDiskBytes(sourcePath)
	if err != nil {
		return nil
	}

	required := sourceSize * 2
	if available < required {
		return &MigrateError{
			Code: ErrMigrateDiskFull,
			Message: fmt.Sprintf(
				"가용 디스크 공간 부족: 필요 %d 바이트 (소스 크기 %d × 2), 가용 %d 바이트",
				required, sourceSize, available,
			),
		}
	}
	return nil
}
