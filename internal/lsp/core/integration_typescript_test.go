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

// tsConfig는 typescript-language-server 통합 테스트용 ServerConfig를 반환합니다.
func tsConfig() config.ServerConfig {
	return config.ServerConfig{
		Language:       "typescript",
		Command:        "typescript-language-server",
		Args:           []string{"--stdio"},
		FileExtensions: []string{".ts", ".tsx"},
		RootMarkers:    []string{"tsconfig.json", "package.json"},
	}
}

// TestIntegration_TypeScript_InitializeAndShutdown은 typescript-language-server를
// 실제로 기동하고 StateReady에 도달한 후 정상적으로 종료되는지 확인합니다.
//
// typescript-language-server가 PATH에 없으면 Skip됩니다.
func TestIntegration_TypeScript_InitializeAndShutdown(t *testing.T) {
	skipIfBinaryMissing(t, "typescript-language-server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(tsConfig())

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

// TestIntegration_TypeScript_OpenFileAndDiagnostics는 typescript-language-server가
// TypeScript 소스 파일의 타입 오류에 대해 진단을 push하는지 확인합니다.
//
// typescript-language-server가 PATH에 없으면 Skip됩니다.
func TestIntegration_TypeScript_OpenFileAndDiagnostics(t *testing.T) {
	skipIfBinaryMissing(t, "typescript-language-server")

	tmpDir := t.TempDir()

	// 타입 오류가 있는 TypeScript 파일 생성
	tsContent := "const x: number = \"string\";\n"
	tsPath := writeTempFile(t, tmpDir, "main.ts", tsContent)

	// tsconfig.json 생성 (root marker)
	writeTempFile(t, tmpDir, "tsconfig.json", `{"compilerOptions":{"strict":true}}`+"\n")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(tsConfig())

	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, tsPath, tsContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// 진단 폴링: 최대 10s 대기
	diags := waitForDiagnostics(ctx, cl, tsPath, 1, 200*time.Millisecond, 10*time.Second)
	if len(diags) == 0 {
		t.Fatal("expected at least 1 diagnostic from typescript-language-server for type mismatch, got 0")
	}

	// 진단 메시지에 타입 오류 관련 키워드가 포함되어 있는지 확인 (case-insensitive)
	typeErrorKeywords := []string{"type", "string", "number", "assignable"}
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
		t.Errorf("expected type-error diagnostic, got: %v", msgs)
	}
}
