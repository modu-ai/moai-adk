package subprocess_test

// @MX:NOTE: [AUTO] TestMain — 패키지 전역 공유 fake binary 초기화 진입점
// @MX:SPEC: SPEC-LSP-FLAKY-001
//
// 이 파일은 패키지 테스트 프로세스의 진입점인 TestMain을 정의한다.
// TestMain에서 공유 fake binary를 단 1회만 생성하고 os.RemoveAll로 정리한다.
// 이로써 t.Parallel() 테스트들이 동시에 실행되는 중에는 어떤 fake binary 파일
// 쓰기도 발생하지 않아 Linux fork-exec ETXTBSY race가 근본적으로 제거된다.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// pkgTempDir는 TestMain이 생성한 패키지 전역 임시 디렉터리다.
// 테스트 프로세스 종료 시 os.RemoveAll로 정리된다.
var pkgTempDir string

// sharedBinaryPath는 sync.OnceValues를 통해 패키지 전역 fake binary를
// 정확히 1회만 생성하고 그 경로를 반환한다.
//
// @MX:NOTE: [AUTO] sharedBinaryPath — ETXTBSY race 제거를 위한 공유 바이너리 경로
// @MX:SPEC: SPEC-LSP-FLAKY-001 REQ-LSP-FLAKY-001-001, REQ-LSP-FLAKY-001-002
//
// 핵심 불변: 이 함수가 반환한 경로의 파일은 테스트 실행 중 절대 재작성되지 않는다.
// 모든 호출자는 read+exec 전용으로 사용해야 하며, 파일 내용을 변경해서는 안 된다.
var sharedBinaryPath = sync.OnceValues(func() (string, error) {
	if runtime.GOOS == "windows" {
		// Windows는 shell script stub을 지원하지 않으므로 빈 경로를 반환.
		// 호출자(sharedFakeBinaryPath)에서 t.Skip()으로 처리한다.
		return "", nil
	}
	path := filepath.Join(pkgTempDir, "shared-fake-lsp")

	// ETXTBSY mitigation 시퀀스: Create → Write → Sync → Close → Chmod 순서 준수.
	// fork/exec 전에 모든 writer fd가 닫혀 있음을 보장한다.
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("sharedBinaryPath create: %w", err)
	}
	if _, err := f.Write([]byte("#!/bin/sh\ncat\n")); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("sharedBinaryPath write: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("sharedBinaryPath sync: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("sharedBinaryPath close: %w", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return "", fmt.Errorf("sharedBinaryPath chmod: %w", err)
	}
	return path, nil
})

// sharedFakeBinaryPath는 패키지 전역 공유 fake LSP stub 바이너리 경로를 반환한다.
//
// 사용 조건:
//   - 바이너리를 read+exec 전용으로 사용하는 테스트에서만 호출한다.
//   - 바이너리 내용이나 권한을 수정하는 테스트는 반드시 writeFakeBinary를 사용한다.
//
// @MX:SPEC: SPEC-LSP-FLAKY-001 REQ-LSP-FLAKY-001-001
func sharedFakeBinaryPath(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stub은 Windows에서 지원되지 않음")
	}
	path, err := sharedBinaryPath()
	if err != nil {
		t.Fatalf("공유 fake binary 초기화 실패: %v", err)
	}
	return path
}

// TestMain은 패키지 테스트 프로세스의 진입점이다.
// 패키지 전역 임시 디렉터리를 생성하고 모든 테스트 완료 후 정리한다.
// 공유 fake binary는 sharedBinaryPath (sync.OnceValues)에 의해 최초 요청 시 생성된다.
func TestMain(m *testing.M) {
	// 패키지 전역 임시 디렉터리 생성.
	// t.TempDir()은 *testing.T가 필요하므로 TestMain에서는 os.MkdirTemp를 사용한다.
	dir, err := os.MkdirTemp("", "moai-lsp-subprocess-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: 임시 디렉터리 생성 실패: %v\n", err)
		os.Exit(1)
	}
	pkgTempDir = dir

	// 모든 테스트 실행.
	code := m.Run()

	// 임시 디렉터리 정리 (공유 fake binary 포함).
	_ = os.RemoveAll(pkgTempDir)

	os.Exit(code)
}
