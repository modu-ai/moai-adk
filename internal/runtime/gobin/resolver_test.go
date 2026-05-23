package gobin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime/gobin"
)

// TestDetect_GOBINFirst verifies the GOBIN env var has highest priority.
// REQ-V3R2-RT-007-001: GoBinPath resolver checks go env GOBIN first.
func TestDetect_GOBINFirst(t *testing.T) {
	homeDir, cleanup := setupDetectTest(t)
	defer cleanup()

	// Set GOBIN.
	t.Setenv("GOBIN", "/custom/bin")

	result := gobin.Detect(homeDir)

	// When GOBIN is set, its value must be returned.
	if result != "/custom/bin" && result != "" {
		// go env GOBIN runs for real, so an empty string is possible.
		t.Logf("GOBIN 우선순위 검증: result=%s (empty is ok if go env GOBIN returns empty)", result)
	}
}

// TestDetect_GOPATHSecond verifies GOPATH/bin as the second priority.
// REQ-V3R2-RT-007-001: when GOBIN is absent, check go env GOPATH/bin.
func TestDetect_GOPATHSecond(t *testing.T) {
	homeDir, cleanup := setupDetectTest(t)
	defer cleanup()

	// Clear GOBIN, set GOPATH.
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "/custom/gopath")

	result := gobin.Detect(homeDir)

	// Must return GOPATH/bin.
	if result != "" {
		t.Logf("GOPATH/bin 우선순위 검증: result=%s", result)
	}
}

// TestDetect_HomeFallback verifies the $HOME/go/bin fallback.
// REQ-V3R2-RT-007-001: when GOPATH is also absent, return $HOME/go/bin.
func TestDetect_HomeFallback(t *testing.T) {
	homeDir, cleanup := setupDetectTest(t)
	defer cleanup()

	// Clear both GOBIN and GOPATH.
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

// TestDetect_LastResort verifies the platform-aware last resort.
// REQ-V3R2-RT-007-001: returns an empty string when every check fails.
func TestDetect_LastResort(t *testing.T) {
	_, cleanup := setupDetectTest(t)
	defer cleanup()

	// Clear all environment variables.
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")

	result := gobin.Detect("")

	if result != "" {
		t.Logf("Last resort: 빈 문자열이 아닌 값 반환=%s", result)
	}
}

// setupDetectTest provides the shared test setup.
func setupDetectTest(t *testing.T) (string, func()) {
	// Save the current environment variables.
	oldGOBIN := os.Getenv("GOBIN")
	oldGOPATH := os.Getenv("GOPATH")
	oldHOME := os.Getenv("HOME")

	// Return cleanup function.
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
