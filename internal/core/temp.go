package core

import (
	"os"
	"runtime"
)

// @MX:ANCHOR: [AUTO] 임시 디렉토리 해석 유틸리티. Windows 환경에서 비-ASCII 사용자명 문제를 처리하는 단일 진입점.
// @MX:REASON: fan_in=9, 파일 생성·훅 처리·템플릿 배포 등 다수 경로에서 호출됨
// TempDir returns the appropriate temp directory.
// On Windows, allows override to avoid 8.3 short filename issues with non-ASCII usernames.
func TempDir() string {
	if runtime.GOOS == "windows" {
		// Check if user has configured custom temp dir
		if custom := os.Getenv("MOAI_TEMP_DIR"); custom != "" {
			return custom
		}
	}
	return os.TempDir()
}

// CreateTempFile wraps os.CreateTemp with path normalization for Windows
func CreateTempFile(dir, pattern string) (*os.File, error) {
	if dir == "" {
		dir = TempDir()
	}
	return os.CreateTemp(dir, pattern)
}
