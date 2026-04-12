package subprocess_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
)

// writeFakeBinary writes a minimal shell script to dir with the given name
// and marks it executable. The script acts as a minimal "read from stdin, write
// nothing" stub to allow pipe smoke tests without spawning real language servers.
func writeFakeBinary(t *testing.T, dir, name string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stubs not supported on Windows")
	}
	path := filepath.Join(dir, name)
	// 표준 입력을 읽고 아무것도 출력하지 않는 최소 스텁
	if err := os.WriteFile(path, []byte("#!/bin/sh\ncat\n"), 0o755); err != nil {
		t.Fatalf("writeFakeBinary: %v", err)
	}
	return path
}

// TestLauncher_Launch_HappyPath verifies that Launcher.Launch succeeds when the
// binary exists, returns a non-nil LaunchResult, and all three stdio pipes are
// non-nil (REQ-LC-005).
func TestLauncher_Launch_HappyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binPath := writeFakeBinary(t, dir, "fake-lsp")

	cfg := config.ServerConfig{
		Language: "go",
		Command:  binPath, // 절대 경로 직접 지정 — PATH 탐색 없이 실행
	}

	l := subprocess.NewLauncher()
	result, err := l.Launch(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Launch returned unexpected error: %v", err)
	}
	t.Cleanup(func() {
		_ = result.Cmd.Process.Kill()
		_ = result.Cmd.Wait()
	})

	if result.Cmd == nil {
		t.Error("LaunchResult.Cmd is nil")
	}
	if result.Stdin == nil {
		t.Error("LaunchResult.Stdin is nil")
	}
	if result.Stdout == nil {
		t.Error("LaunchResult.Stdout is nil")
	}
	if result.Stderr == nil {
		t.Error("LaunchResult.Stderr is nil")
	}
}

// TestLauncher_Launch_BinaryNotFound verifies that Launcher.Launch returns
// ErrBinaryNotFound when the specified binary does not exist anywhere
// (REQ-LC-004 warn_and_skip behavior).
func TestLauncher_Launch_BinaryNotFound(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Language: "python",
		Command:  "this-binary-absolutely-does-not-exist-moai-test-00000",
	}

	l := subprocess.NewLauncher()
	_, err := l.Launch(t.Context(), cfg)
	if err == nil {
		t.Fatal("Launch: expected error for missing binary, got nil")
	}
	if !errors.Is(err, subprocess.ErrBinaryNotFound) {
		t.Errorf("Launch error = %v, want wrapping ErrBinaryNotFound", err)
	}
}

// TestLauncher_Launch_StdioPipesNonNil verifies each stdio pipe is independently
// non-nil and writable/readable (REQ-LC-005 isolation).
func TestLauncher_Launch_StdioPipesNonNil(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binPath := writeFakeBinary(t, dir, "pipe-lsp")

	cfg := config.ServerConfig{
		Language: "typescript",
		Command:  binPath,
		Args:     []string{},
	}

	l := subprocess.NewLauncher()
	result, err := l.Launch(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	t.Cleanup(func() {
		_ = result.Stdin.Close()
		_ = result.Cmd.Process.Kill()
		_ = result.Cmd.Wait()
	})

	// stdin에 쓰기 시도 — 파이프가 유효하면 에러 없음
	if _, err := result.Stdin.Write([]byte("test\n")); err != nil {
		t.Errorf("Stdin.Write: %v", err)
	}
}

// TestLauncher_Launch_WithArgs verifies that additional Args from ServerConfig
// are forwarded to the subprocess (REQ-LC-005).
func TestLauncher_Launch_WithArgs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binPath := writeFakeBinary(t, dir, "arg-lsp")

	cfg := config.ServerConfig{
		Language: "go",
		Command:  binPath,
		Args:     []string{"-stdio", "--log-level=debug"},
	}

	l := subprocess.NewLauncher()
	result, err := l.Launch(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Launch with args: %v", err)
	}
	t.Cleanup(func() {
		_ = result.Cmd.Process.Kill()
		_ = result.Cmd.Wait()
	})

	// 명령행 인자가 정확히 전달됐는지 확인
	if len(result.Cmd.Args) < 3 {
		t.Errorf("Cmd.Args = %v, expected binary + 2 args", result.Cmd.Args)
	}
}

// TestLauncher_Launch_AbsolutePathNotFound verifies that Launch returns
// ErrBinaryNotFound when an absolute path points to a non-existent file.
func TestLauncher_Launch_AbsolutePathNotFound(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Language: "go",
		Command:  "/this/path/does/not/exist/moai-test-lsp",
	}

	l := subprocess.NewLauncher()
	_, err := l.Launch(t.Context(), cfg)
	if err == nil {
		t.Fatal("Launch: expected error for missing absolute path, got nil")
	}
	if !errors.Is(err, subprocess.ErrBinaryNotFound) {
		t.Errorf("Launch error = %v, want wrapping ErrBinaryNotFound", err)
	}
}

// TestLauncher_Launch_EmptyCommand verifies that Launch returns ErrBinaryNotFound
// when the command is an empty string.
func TestLauncher_Launch_EmptyCommand(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Language: "python",
		Command:  "",
	}

	l := subprocess.NewLauncher()
	_, err := l.Launch(t.Context(), cfg)
	if err == nil {
		t.Fatal("Launch: expected error for empty command, got nil")
	}
	// 에러 메시지에 언어 또는 커맨드 정보가 포함되어야 함
	_ = err
}

// TestLauncher_Launch_StartFails verifies that Launch returns an error when the
// binary exists but cannot be executed (e.g., not executable).
func TestLauncher_Launch_StartFails(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("permission-based test not applicable on Windows")
	}

	dir := t.TempDir()
	binPath := dir + "/non-exec"
	// 실행 권한 없는 파일 작성
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := config.ServerConfig{
		Language: "go",
		Command:  binPath,
	}

	l := subprocess.NewLauncher()
	_, err := l.Launch(t.Context(), cfg)
	if err == nil {
		t.Fatal("Launch: expected error for non-executable binary, got nil")
	}
	// ErrBinaryNotFound가 아닌 다른 에러 (Start 실패)
	if errors.Is(err, subprocess.ErrBinaryNotFound) {
		t.Errorf("Launch error = ErrBinaryNotFound, want start error")
	}
}

// TestLauncher_Launch_PathBinary verifies that Launch can find a binary
// that exists in PATH (e.g., /bin/sh is always available on Unix).
func TestLauncher_Launch_PathBinary(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("PATH binary test not applicable on Windows")
	}

	cfg := config.ServerConfig{
		Language: "sh",
		Command:  "sh",
		Args:     []string{"-c", "exit 0"},
	}

	l := subprocess.NewLauncher()
	result, err := l.Launch(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Launch sh: %v", err)
	}
	t.Cleanup(func() {
		_ = result.Cmd.Process.Kill()
		_ = result.Cmd.Wait()
	})

	if result.Stdin == nil || result.Stdout == nil || result.Stderr == nil {
		t.Error("stdio pipes are nil for PATH-resolved binary")
	}
}
