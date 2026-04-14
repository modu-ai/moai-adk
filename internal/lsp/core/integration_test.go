//go:build integration

package core_test

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
)

// goplsConfig는 gopls 통합 테스트에 사용되는 표준 ServerConfig를 반환합니다.
// rootDir이 비어 있으면 rootUri를 생략합니다.
func goplsConfig() config.ServerConfig {
	return config.ServerConfig{
		Language:       "go",
		Command:        "gopls",
		Args:           []string{"serve"},
		FileExtensions: []string{".go"},
		RootMarkers:    []string{"go.mod"},
	}
}

// goplsConfigWithRoot는 rootDir이 설정된 gopls ServerConfig를 반환합니다.
// rootDir은 go.mod가 위치한 임시 디렉토리여야 합니다.
func goplsConfigWithRoot(rootDir string) config.ServerConfig {
	cfg := goplsConfig()
	cfg.RootDir = rootDir
	return cfg
}

// TestIntegration_Gopls_InitializeAndShutdown은 gopls를 실제로 기동하고
// StateReady에 도달한 후 정상적으로 종료되는지 확인합니다 (REQ-LC-010).
func TestIntegration_Gopls_InitializeAndShutdown(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(goplsConfig())

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

	// 고루틴 누수 확인: 종료 후 최대 2개 이내 증가 허용
	// (runtime 내부 GC 고루틴 등 일시적 변동 허용)
	after := runtime.NumGoroutine()
	if delta := after - before; delta > 2 {
		t.Errorf("goroutine leak suspected: before=%d after=%d delta=%d", before, after, delta)
	}
}

// TestIntegration_Gopls_OpenFileAndGetDiagnostics는 gopls가 Go 소스 파일의
// 컴파일 오류에 대해 진단을 push하는지 확인합니다 (REQ-LC-010).
func TestIntegration_Gopls_OpenFileAndGetDiagnostics(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	// makeGoModuleDir: /tmp 아래에 생성하여 macOS system temp root 제한 우회
	tmpDir := makeGoModuleDir(t)

	// undefined 변수를 참조하는 소스: gopls가 진단을 push해야 함
	mainContent := `package main

func main() {
	_ = undefinedVariable
}
`
	_, mainGoPath := writeGoModule(t, tmpDir, "testmod", mainContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// RootDir을 tmpDir로 설정: gopls에 rootUri+workspaceFolders를 전달
	cl := core.NewClient(goplsConfigWithRoot(tmpDir))

	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	// didOpen 전송
	if err := cl.OpenFile(ctx, mainGoPath, mainContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// 진단 폴링: gopls는 비동기로 publishDiagnostics를 push함
	diags := waitForDiagnostics(ctx, cl, mainGoPath, 1, 100*time.Millisecond, 10*time.Second)
	if len(diags) == 0 {
		t.Fatal("expected at least 1 diagnostic from gopls for undefined variable, got 0")
	}

	// 진단 메시지에 "undefined" 포함 여부 확인 (case-insensitive)
	found := false
	for _, d := range diags {
		if strings.Contains(strings.ToLower(d.Message), "undefined") {
			found = true
			break
		}
	}
	if !found {
		msgs := make([]string, len(diags))
		for i, d := range diags {
			msgs[i] = d.Message
		}
		t.Errorf("expected diagnostic with 'undefined' in message, got: %v", msgs)
	}
}

// TestIntegration_Gopls_FindReferences는 gopls가 심볼 참조 위치를 반환하는지
// 확인합니다 (REQ-LC-010, REQ-LC-002b).
func TestIntegration_Gopls_FindReferences(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	// makeGoModuleDir: /tmp 아래에 생성하여 macOS system temp root 제한 우회
	tmpDir := makeGoModuleDir(t)
	// foo() 정의 (line 2, 0-indexed) + foo() 호출 (line 5, 0-indexed)
	mainContent := `package main

func foo() {}

func main() {
	foo()
}
`
	_, mainGoPath := writeGoModule(t, tmpDir, "testmod", mainContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// RootDir을 tmpDir로 설정: gopls에 rootUri+workspaceFolders를 전달
	cl := core.NewClient(goplsConfigWithRoot(tmpDir))
	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, mainGoPath, mainContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// gopls 인덱싱 대기 (didOpen 후 초기 분석 시간 필요)
	time.Sleep(500 * time.Millisecond)

	// foo 정의 위치 (line=2, character=5): "func foo()" 에서 'f' 위치
	locs, err := cl.FindReferences(ctx, mainGoPath, lspPosition(2, 5))
	if err != nil {
		// ErrCapabilityUnsupported면 gopls 버전 문제 — skip
		if strings.Contains(err.Error(), "capability") {
			t.Skipf("gopls does not advertise references capability: %v", err)
		}
		t.Fatalf("FindReferences: %v", err)
	}

	if len(locs) < 2 {
		t.Errorf("FindReferences: expected >= 2 locations (definition + call site), got %d: %v", len(locs), locs)
	}
}

// TestIntegration_Gopls_GotoDefinition는 gopls가 심볼 정의 위치를 반환하는지
// 확인합니다 (REQ-LC-010, REQ-LC-002b).
func TestIntegration_Gopls_GotoDefinition(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	// makeGoModuleDir: /tmp 아래에 생성하여 macOS system temp root 제한 우회
	tmpDir := makeGoModuleDir(t)
	// foo() 정의 (line 2) + foo() 호출 (line 5)
	mainContent := `package main

func foo() {}

func main() {
	foo()
}
`
	_, mainGoPath := writeGoModule(t, tmpDir, "testmod", mainContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// RootDir을 tmpDir로 설정: gopls에 rootUri+workspaceFolders를 전달
	cl := core.NewClient(goplsConfigWithRoot(tmpDir))
	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, mainGoPath, mainContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// gopls 인덱싱 대기
	time.Sleep(500 * time.Millisecond)

	// foo() 호출 위치 (line=5, character=1): "\tfoo()" 에서 'f' 위치
	locs, err := cl.GotoDefinition(ctx, mainGoPath, lspPosition(5, 1))
	if err != nil {
		if strings.Contains(err.Error(), "capability") {
			t.Skipf("gopls does not advertise definition capability: %v", err)
		}
		t.Fatalf("GotoDefinition: %v", err)
	}

	if len(locs) == 0 {
		t.Fatal("GotoDefinition: expected at least 1 location, got 0")
	}

	// 반환된 위치는 정의 라인(line=2)을 포함해야 함
	found := false
	for _, loc := range locs {
		if loc.Range.Start.Line == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GotoDefinition: expected location on line 2 (foo definition), got: %v", locs)
	}
}
