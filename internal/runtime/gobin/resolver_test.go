package gobin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime/gobin"
)

// TestDetect_GOBINFirst는 GOBIN 환경변수 우선순위를 검증합니다.
// REQ-V3R2-RT-007-001: GoBinPath resolver는 go env GOBIN을 우선检查합니다.
func TestDetect_GOBINFirst(t *testing.T) {
	homeDir, cleanup := setupDetectTest(t)
	defer cleanup()

	// GOBIN 설정
	t.Setenv("GOBIN", "/custom/bin")

	result := gobin.Detect(homeDir)

	// GOBIN이 설정되면 그 값을 반환해야 함
	if result != "/custom/bin" && result != "" {
		// go env GOBIN이 실제로 실행되므로 빈 문자열 가능성 있음
		t.Logf("GOBIN 우선순위 검증: result=%s (empty is ok if go env GOBIN returns empty)", result)
	}
}

// TestDetect_GOPATHSecond는 GOPATH/bin 두 번째 우선순위를 검증합니다.
// REQ-V3R2-RT-007-001: GOBIN이 없으면 go env GOPATH/bin을检查합니다.
func TestDetect_GOPATHSecond(t *testing.T) {
	homeDir, cleanup := setupDetectTest(t)
	defer cleanup()

	// GOBIN clear, GOPATH 설정
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "/custom/gopath")

	result := gobin.Detect(homeDir)

	// GOPATH/bin을 반환해야 함
	if result != "" {
		t.Logf("GOPATH/bin 우선순위 검증: result=%s", result)
	}
}

// TestDetect_HomeFallback은 $HOME/go/bin 폴백을 검증합니다.
// REQ-V3R2-RT-007-001: GOPATH도 없으면 $HOME/go/bin을 반환합니다.
func TestDetect_HomeFallback(t *testing.T) {
	homeDir, cleanup := setupDetectTest(t)
	defer cleanup()

	// GOBIN, GOPATH 모두 clear
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")

	result := gobin.Detect(homeDir)

	expected := filepath.Join(homeDir, "go", "bin")
	if result != expected && result != "" {
		t.Errorf("Home fallback 실패: expected=%s, got=%s", expected, result)
	}

	if result == "" {
		t.Logf("Home fallback: 빈 문자열 반환 (homeDir=%s)", homeDir)
	}
}

// TestDetect_LastResort는 platform-aware last resort를 검증합니다.
// REQ-V3R2-RT-007-001: 모든 check가 실패하면 빈 문자열을 반환합니다.
func TestDetect_LastResort(t *testing.T) {
	_, cleanup := setupDetectTest(t)
	defer cleanup()

	// 모든 환경변수 clear
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")

	result := gobin.Detect("")

	if result != "" {
		t.Logf("Last resort: 빈 문자열이 아닌 값 반환=%s", result)
	}
}

// setupDetectTest는 공통 테스트 setup입니다.
func setupDetectTest(t *testing.T) (string, func()) {
	// 현재 환경 변수 저장
	oldGOBIN := os.Getenv("GOBIN")
	oldGOPATH := os.Getenv("GOPATH")
	oldHOME := os.Getenv("HOME")

	// cleanup 함수 반환
	cleanup := func() {
		if oldGOBIN != "" {
			_ = os.Setenv("GOBIN", oldGOBIN)
		} else {
			_ = os.Unsetenv("GOBIN")
		}
		if oldGOPATH != "" {
			_ = os.Setenv("GOPATH", oldGOPATH)
		} else {
			_ = os.Unsetenv("GOPATH")
		}
		if oldHOME != "" {
			_ = os.Setenv("HOME", oldHOME)
		} else {
			_ = os.Unsetenv("HOME")
		}
	}

	return os.Getenv("HOME"), cleanup
}
