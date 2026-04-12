//go:build integration

package core_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
)

// pyrightConfig는 pyright-langserver 통합 테스트용 ServerConfig를 반환합니다.
func pyrightConfig() config.ServerConfig {
	return config.ServerConfig{
		Language:       "python",
		Command:        "pyright-langserver",
		Args:           []string{"--stdio"},
		FileExtensions: []string{".py"},
		RootMarkers:    []string{"pyproject.toml", "setup.py", "setup.cfg"},
	}
}

// pyrightConfigWithRoot는 rootDir이 설정된 pyright ServerConfig를 반환합니다.
func pyrightConfigWithRoot(rootDir string) config.ServerConfig {
	cfg := pyrightConfig()
	cfg.RootDir = rootDir
	return cfg
}

// TestIntegration_Pyright_InitializeAndShutdown은 pyright-langserver를 실제로
// 기동하고 StateReady에 도달한 후 정상적으로 종료되는지 확인합니다.
func TestIntegration_Pyright_InitializeAndShutdown(t *testing.T) {
	skipIfBinaryMissing(t, "pyright-langserver")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(pyrightConfig())

	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	assertStateEquals(t, cl, core.StateReady)

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := cl.Shutdown(shutCtx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
	assertStateEquals(t, cl, core.StateShutdown)
}

// TestIntegration_Pyright_OpenFileAndGetDiagnostics는 pyright-langserver가
// Python 소스 파일의 타입 오류에 대해 진단을 push하는지 확인합니다.
//
// pyright는 초기 분석 시간이 길어 최대 15s 폴링 윈도우를 사용합니다.
func TestIntegration_Pyright_OpenFileAndGetDiagnostics(t *testing.T) {
	skipIfBinaryMissing(t, "pyright-langserver")

	tmpDir := t.TempDir()

	// 타입 오류가 포함된 Python 파일 생성
	// pyright는 타입 불일치를 정적으로 감지함
	pyContent := `x: int = "this is a string"
`
	pyPath := writeTempFile(t, tmpDir, "main.py", pyContent)

	// pyproject.toml 생성 (pyright root marker)
	writeTempFile(t, tmpDir, "pyproject.toml", "[tool.pyright]\n")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cl core.Client
	var startErr error

	// pyright 기동 재시도 (최대 2회): 첫 번째 기동이 실패하는 경우 대비
	// RootDir을 tmpDir로 설정하여 pyright에 workspace root 전달
	for attempt := 0; attempt < 2; attempt++ {
		cl = core.NewClient(pyrightConfigWithRoot(tmpDir))
		startErr = cl.Start(ctx)
		if startErr == nil {
			break
		}
		t.Logf("Pyright Start attempt %d failed: %v — retrying", attempt+1, startErr)
		time.Sleep(500 * time.Millisecond)
	}
	if startErr != nil {
		t.Fatalf("Start (after 2 attempts): %v", startErr)
	}

	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, pyPath, pyContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// 진단 폴링: pyright는 gopls보다 느린 분석기 — 최대 15s 대기
	diags := waitForDiagnostics(ctx, cl, pyPath, 1, 200*time.Millisecond, 15*time.Second)
	if len(diags) == 0 {
		t.Fatal("expected at least 1 diagnostic from pyright for type mismatch, got 0")
	}

	// 진단 메시지에 타입 오류 관련 키워드가 포함되어 있는지 확인 (case-insensitive)
	// pyright는 "Expression of type" 혹은 "Cannot assign" 등을 출력함
	typeErrorKeywords := []string{"type", "assign", "str", "int", "expression"}
	found := false
outer:
	for _, d := range diags {
		lower := strings.ToLower(d.Message)
		for _, kw := range typeErrorKeywords {
			if strings.Contains(lower, kw) {
				found = true
				break outer
			}
		}
	}
	if !found {
		msgs := make([]string, len(diags))
		for i, d := range diags {
			msgs[i] = d.Message
		}
		t.Errorf("expected type-error diagnostic from pyright, got: %v", msgs)
	}
}
