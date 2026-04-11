package loop

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// mockBridge는 테스트용 gopls.Bridge 대역이다.
// GetDiagnostics 호출 결과를 미리 설정할 수 있다.
type mockBridge struct {
	diagnostics []gopls.Diagnostic
	err         error
	called      bool
}

func (m *mockBridge) GetDiagnostics(ctx context.Context, path string) ([]gopls.Diagnostic, error) {
	m.called = true
	return m.diagnostics, m.err
}

// TestGoFeedbackGenerator_NilBridge는 bridge가 nil일 때 기존 동작이
// 그대로 유지되는지 검증한다 (하위 호환성).
func TestGoFeedbackGenerator_NilBridge(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	// bridge=nil로 생성
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, nil)

	if gen == nil {
		t.Fatal("NewGoFeedbackGeneratorWithBridge()는 nil을 반환해선 안 된다")
	}
}

// TestGoFeedbackGenerator_NilBridgeCollect는 bridge가 nil일 때
// Collect()가 Diagnostics 필드를 비워두고 정상 반환하는지 검증한다.
func TestGoFeedbackGenerator_NilBridgeCollect(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() 오류 = %v, 기대값: nil", err)
	}
	if fb == nil {
		t.Fatal("Collect()가 nil Feedback을 반환했다")
	}
	// bridge가 nil이면 Diagnostics는 nil이어야 한다.
	if fb.Diagnostics != nil {
		t.Errorf("Diagnostics = %v, 기대값: nil (bridge가 nil이므로)", fb.Diagnostics)
	}
}

// TestGoFeedbackGenerator_WithBridge_PopulatesDiagnostics는
// bridge가 있을 때 GetDiagnostics 결과가 Feedback.Diagnostics에 채워지는지 검증한다.
func TestGoFeedbackGenerator_WithBridge_PopulatesDiagnostics(t *testing.T) {
	t.Parallel()

	expectedDiags := []gopls.Diagnostic{
		{
			Severity: gopls.SeverityError,
			Message:  "undeclared name: foo",
			Source:   "compiler",
		},
		{
			Severity: gopls.SeverityWarning,
			Message:  "unused variable: bar",
			Source:   "staticcheck",
		},
	}

	mock := &mockBridge{diagnostics: expectedDiags}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() 오류 = %v, 기대값: nil", err)
	}
	if fb == nil {
		t.Fatal("Collect()가 nil Feedback을 반환했다")
	}

	if !mock.called {
		t.Error("bridge.GetDiagnostics()가 호출되지 않았다")
	}

	if len(fb.Diagnostics) != len(expectedDiags) {
		t.Errorf("Diagnostics 수 = %d, 기대값 %d", len(fb.Diagnostics), len(expectedDiags))
	}

	for i, d := range fb.Diagnostics {
		if d.Severity != expectedDiags[i].Severity {
			t.Errorf("[%d] Severity = %d, 기대값 %d", i, d.Severity, expectedDiags[i].Severity)
		}
		if d.Message != expectedDiags[i].Message {
			t.Errorf("[%d] Message = %q, 기대값 %q", i, d.Message, expectedDiags[i].Message)
		}
	}
}

// TestGoFeedbackGenerator_WithBridge_ErrorIsSilent는
// bridge.GetDiagnostics가 오류를 반환해도 Collect()가 오류 없이
// 빈 Diagnostics를 반환하는지 검증한다.
func TestGoFeedbackGenerator_WithBridge_ErrorIsSilent(t *testing.T) {
	t.Parallel()

	mock := &mockBridge{err: context.DeadlineExceeded}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect()는 bridge 오류를 상위로 전파해선 안 된다. 오류 = %v", err)
	}
	if fb == nil {
		t.Fatal("Collect()가 nil Feedback을 반환했다")
	}
	// 오류 시 Diagnostics는 nil이어야 한다 (부분 결과 아님).
	if len(fb.Diagnostics) != 0 {
		t.Errorf("bridge 오류 시 Diagnostics는 비어야 함, got %v", fb.Diagnostics)
	}
}

// TestGoFeedbackGenerator_BackwardCompat은 NewGoFeedbackGenerator(기존 시그니처)가
// 여전히 동작하는지 검증한다 (하위 호환성).
func TestGoFeedbackGenerator_BackwardCompat(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	// 기존 생성자가 여전히 컴파일되어야 한다.
	gen := NewGoFeedbackGenerator(projectRoot)
	if gen == nil {
		t.Fatal("NewGoFeedbackGenerator()는 nil을 반환해선 안 된다")
	}
}

// TestGoFeedbackGenerator_WithBridge_EmptyDiagnostics는
// bridge가 빈 슬라이스를 반환할 때 Diagnostics가 비어있는지 검증한다.
func TestGoFeedbackGenerator_WithBridge_EmptyDiagnostics(t *testing.T) {
	t.Parallel()

	mock := &mockBridge{diagnostics: []gopls.Diagnostic{}}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() 오류 = %v, 기대값: nil", err)
	}
	if len(fb.Diagnostics) != 0 {
		t.Errorf("Diagnostics = %v, 기대값: 빈 슬라이스", fb.Diagnostics)
	}
}
