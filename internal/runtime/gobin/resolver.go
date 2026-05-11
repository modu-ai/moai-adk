package gobin

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// @MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-007 REQ-001 GoBinPath 단일 진실 소스.
// internal/core/project/initializer.go:286 (init 경로)와
// internal/cli/update.go:2553 (update 경로)에서 호출됩니다.
// 새로운 호출자를 추가할 때 폴백 체인 일관성을 유지해야 합니다.

// Detect는 사용자의 Go bin directory 경로를 반환합니다.
// 폴백 체인 순서 (REQ-V3R2-RT-007-001):
// 1. go env GOBIN (설정된 경우 우선)
// 2. go env GOPATH/bin (GOBIN이 없는 경우)
// 3. $HOME/go/bin (기본값)
// 4. platform-aware last resort (모두 실패한 경우)
func Detect(homeDir string) string {
	// 1. GOBIN 환경변수 또는 go env GOBIN 확인
	if gobin := goEnvGOBIN(); gobin != "" {
		return gobin
	}

	// 2. GOPATH/bin 확인
	if gopathBin := goEnvGOPATHBin(); gopathBin != "" {
		return gopathBin
	}

	// 3. $HOME/go/bin 기본값
	if homeDir != "" {
		return filepath.Join(homeDir, "go", "bin")
	}

	// 4. Last resort: 빈 문자열 (호출자가 처리)
	return ""
}

// goEnvGOBIN은 go env GOBIN 값을 반환합니다.
func goEnvGOBIN() string {
	// go env GOBIN 실행
	cmd := exec.Command("go", "env", "GOBIN")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// goEnvGOPATHBin은 go env GOPATH/bin 값을 반환합니다.
func goEnvGOPATHBin() string {
	// go env GOPATH 실행
	cmd := exec.Command("go", "env", "GOPATH")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	gopath := strings.TrimSpace(string(output))
	if gopath == "" {
		return ""
	}

	return filepath.Join(gopath, "bin")
}
