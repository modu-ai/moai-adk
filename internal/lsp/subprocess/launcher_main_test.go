package subprocess_test

// @MX:NOTE: [AUTO] TestMain — 패키지 전역 공유 fake binary 초기화 진입점
// @MX:SPEC: SPEC-LSP-FLAKY-001, SPEC-LSP-FLAKY-002
//
// 이 파일은 패키지 테스트 프로세스의 진입점인 TestMain을 정의한다.
// TestMain에서 공유 fake binary를 m.Run() 호출 *전*에 작성하고 모든 writer fd가
// 닫힌 후에만 t.Parallel() 테스트들이 시작된다. 이로써 Linux fork-exec ETXTBSY
// race가 근본 제거된다.
//
// History:
// - SPEC-LSP-FLAKY-001 (1차 시도): sync.OnceValues lazy 초기화. supervisor_test
//   goroutine 이 t.Parallel() 중 fork 할 때 launcher OnceValues 의 writer fd 를
//   잠시 inherit 하는 race 가 잔존하여 Ubuntu CI 에서 여전히 flake 발생.
// - SPEC-LSP-FLAKY-002 (현재): TestMain 에서 eager 작성. m.Run() 진입 시점에
//   writer fd 가 이미 닫혀 있으므로 어떤 t.Parallel() goroutine 의 fork 도
//   shared binary 의 writer fd 를 inherit 할 수 없다.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// pkgTempDir는 TestMain이 생성한 패키지 전역 임시 디렉터리다.
// 테스트 프로세스 종료 시 os.RemoveAll로 정리된다.
var pkgTempDir string

// pkgSharedBinaryPath는 TestMain이 m.Run() 호출 전에 작성한 공유 fake LSP stub
// 바이너리의 절대 경로이다. Windows 에서는 빈 문자열 (호출자가 t.Skip 처리).
//
// @MX:ANCHOR: [AUTO] pkgSharedBinaryPath — ETXTBSY race 제거를 위한 공유 바이너리 경로
// @MX:REASON: fan_in >= 3 — TestLauncher_Launch_HappyPath, _StdioPipesNonNil, _WithArgs 가 모두 참조
// @MX:SPEC: SPEC-LSP-FLAKY-002 REQ-001
//
// 핵심 불변: TestMain 종료 후 이 경로의 파일은 절대 재작성되지 않는다.
// 모든 호출자는 read+exec 전용으로 사용해야 하며, 파일 내용을 변경해서는 안 된다.
var pkgSharedBinaryPath string

// sharedFakeBinaryPath는 패키지 전역 공유 fake LSP stub 바이너리 경로를 반환한다.
//
// 사용 조건:
//   - 바이너리를 read+exec 전용으로 사용하는 테스트에서만 호출한다.
//   - 바이너리 내용이나 권한을 수정하는 테스트는 반드시 writeFakeBinary를 사용한다.
//
// @MX:SPEC: SPEC-LSP-FLAKY-002 REQ-001
func sharedFakeBinaryPath(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stub은 Windows에서 지원되지 않음")
	}
	if pkgSharedBinaryPath == "" {
		t.Fatal("pkgSharedBinaryPath 가 초기화되지 않음 (TestMain 에서 setup 실패)")
	}
	return pkgSharedBinaryPath
}

// buildSharedBinary는 TestMain 단계에서 fake LSP stub 바이너리를 작성한다.
// 작성 순서: Create → Write → Sync → Close → Chmod.
// 반환 시점에 모든 writer fd 가 닫혀 있음을 보장한다.
//
// 이 함수는 m.Run() 호출 *전* 에만 호출되어야 하며, 어떤 t.Parallel() goroutine
// 도 시작되지 않은 상태에서 실행된다. 따라서 fork-exec ETXTBSY race 가 발생
// 할 수 없다.
func buildSharedBinary(dir string) (string, error) {
	if runtime.GOOS == "windows" {
		// Windows 에서는 shell script stub 을 지원하지 않음. 호출자가 t.Skip 처리.
		return "", nil
	}
	path := filepath.Join(dir, "shared-fake-lsp")

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("buildSharedBinary create: %w", err)
	}
	if _, err := f.Write([]byte("#!/bin/sh\ncat\n")); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("buildSharedBinary write: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("buildSharedBinary sync: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("buildSharedBinary close: %w", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return "", fmt.Errorf("buildSharedBinary chmod: %w", err)
	}
	return path, nil
}

// TestMain은 패키지 테스트 프로세스의 진입점이다.
//
// 책임:
//  1. 패키지 전역 임시 디렉터리 생성
//  2. 공유 fake binary 작성 (m.Run() 진입 전, t.Parallel() 시작 전)
//  3. m.Run() 으로 모든 테스트 실행 (writer fd 가 이미 닫힌 상태)
//  4. 임시 디렉터리 정리
//
// 핵심 설계 원칙: m.Run() 호출 전에 모든 binary 작성이 완료되어 writer fd 가
// 닫혀 있어야 한다. 그래야 t.Parallel() 테스트들이 동시에 fork-exec 를 호출해도
// 어떤 자식 프로세스도 shared-fake-lsp 의 writer fd 를 inherit 할 수 없다.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "moai-lsp-subprocess-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: 임시 디렉터리 생성 실패: %v\n", err)
		os.Exit(1)
	}
	pkgTempDir = dir

	// CRITICAL: m.Run() 호출 *전* 에 binary 를 eager 하게 작성한다.
	// 이 시점에는 t.Parallel() goroutine 이 아직 시작되지 않았으므로 fork-exec
	// ETXTBSY race 가 발생할 수 없다.
	binPath, err := buildSharedBinary(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: 공유 fake binary 작성 실패: %v\n", err)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}
	pkgSharedBinaryPath = binPath

	code := m.Run()

	_ = os.RemoveAll(pkgTempDir)
	os.Exit(code)
}
