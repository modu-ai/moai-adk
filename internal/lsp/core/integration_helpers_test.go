//go:build integration

package core_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
)

// skipIfBinaryMissing는 cmd 바이너리가 PATH에 없으면 테스트를 건너뜁니다.
func skipIfBinaryMissing(t *testing.T, cmd string) {
	t.Helper()
	if _, err := exec.LookPath(cmd); err != nil {
		t.Skipf("integration: binary %q not found in PATH — skipping (install to run this test)", cmd)
	}
}

// waitForDiagnostics는 pollInterval 간격으로 GetDiagnostics를 폴링하여
// minCount 이상의 진단이 반환될 때까지 최대 totalTimeout 동안 대기합니다.
// totalTimeout 내에 조건이 충족되지 않으면 빈 슬라이스를 반환합니다.
func waitForDiagnostics(
	ctx context.Context,
	cl core.Client,
	path string,
	minCount int,
	pollInterval, totalTimeout time.Duration,
) []lsp.Diagnostic {
	deadline := time.Now().Add(totalTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		diags, err := cl.GetDiagnostics(ctx, path)
		if err == nil && len(diags) >= minCount {
			return diags
		}
		time.Sleep(pollInterval)
	}
	// 마지막으로 한 번 더 시도
	diags, err := cl.GetDiagnostics(ctx, path)
	if err == nil {
		return diags
	}
	return nil
}

// writeTempFile은 dir/name 경로에 content를 쓰고 절대 경로를 반환합니다.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: failed to write %s: %v", path, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("writeTempFile: failed to resolve absolute path for %s: %v", path, err)
	}
	return abs
}

// assertStateEquals는 Client.State()가 expected와 같은지 확인합니다.
func assertStateEquals(t *testing.T, cl core.Client, expected core.ClientState) {
	t.Helper()
	got := cl.State()
	if got != expected {
		t.Errorf("client state: got %q, want %q", got, expected)
	}
}

// lspPosition은 0-based line, character를 lsp.Position으로 변환합니다.
func lspPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

// writeGoModule은 tmpDir에 go.mod와 main.go를 생성하고 절대 경로를 반환합니다.
// mainContent가 비어 있으면 기본 main.go를 사용합니다.
func writeGoModule(t *testing.T, tmpDir, moduleName, mainContent string) (goModPath, mainGoPath string) {
	t.Helper()

	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n", moduleName)
	goModPath = writeTempFile(t, tmpDir, "go.mod", goModContent)

	if mainContent == "" {
		mainContent = "package main\n\nfunc main() {}\n"
	}
	mainGoPath = writeTempFile(t, tmpDir, "main.go", mainContent)
	return goModPath, mainGoPath
}

// makeGoModuleDir는 gopls가 인식할 수 있는 Go 모듈 임시 디렉토리를 생성합니다.
//
// macOS에서 os.TempDir()은 /var/folders/...를 반환하며, Go 도구는 system temp root
// 내의 go.mod를 무시합니다. 이 함수는 명시적으로 /tmp 아래에 디렉토리를 생성하여
// gopls가 Go 모듈을 올바르게 인식하도록 합니다.
//
// t.Cleanup으로 자동 삭제됩니다.
func makeGoModuleDir(t *testing.T) string {
	t.Helper()
	// /tmp 아래에 생성: macOS system temp root(/var/folders) 우회
	dir, err := os.MkdirTemp("/tmp", "lsp_integration_*")
	if err != nil {
		t.Fatalf("makeGoModuleDir: MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	// symlink 해제: gopls가 실제 경로로 URI를 구성하기 때문에 일관성 유지
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return dir
	}
	return realDir
}
