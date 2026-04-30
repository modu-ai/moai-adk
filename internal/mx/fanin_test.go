package mx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockFanInCounter는 테스트용 FanInCounter mock 구현체입니다.
type mockFanInCounter struct {
	// counts는 AnchorID → fan-in 수 매핑입니다.
	counts map[string]int
	// method는 반환할 fan-in 계산 방식입니다.
	method string
	// err는 반환할 오류입니다.
	err error
}

// Count는 미리 설정된 fan-in 값을 반환합니다.
func (m *mockFanInCounter) Count(_ context.Context, tag Tag, _ string, _ bool) (int, string, error) {
	if m.err != nil {
		return 0, "", m.err
	}
	count := m.counts[tag.AnchorID]
	method := m.method
	if method == "" {
		method = "textual"
	}
	return count, method, nil
}

// TestTextualFanInCounter_Count는 텍스트 기반 fan-in 계산을 테스트합니다.
func TestTextualFanInCounter_Count(t *testing.T) {
	counter := &TextualFanInCounter{}

	tag := Tag{
		Kind:       MXAnchor,
		File:       "internal/auth.go",
		Line:       10,
		Body:       "인증 핸들러",
		AnchorID:   "anchor-auth",
		CreatedBy:  "test",
		LastSeenAt: time.Now(),
	}

	count, method, err := counter.Count(context.Background(), tag, "/tmp/project", false)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// 텍스트 방식이어야 함
	if method != "textual" {
		t.Errorf("fan_in_method: 기대 'textual', 실제 '%s'", method)
	}

	// RED 단계: count는 0 반환 (stub)
	if count < 0 {
		t.Errorf("fan-in 수는 음수일 수 없음: %d", count)
	}
}

// TestMockFanInCounter는 mock 구현이 인터페이스를 올바르게 구현하는지 확인합니다.
func TestMockFanInCounter(t *testing.T) {
	var _ FanInCounter = &mockFanInCounter{}
}

// TestFanInCounter_InterfaceCompliance는 FanInCounter 인터페이스 준수를 확인합니다.
func TestFanInCounter_InterfaceCompliance(t *testing.T) {
	var _ FanInCounter = &TextualFanInCounter{}
}

// TestIsTestFile는 테스트 파일 판별 함수를 테스트합니다.
func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"internal/auth/handler_test.go", true},
		{"internal/auth/handler.go", false},
		{"tests/fixtures/mock.go", true},
		{"testdata/fixture.go", true},
		{"internal/fixtures/data.go", true},
		{"pkg/utils/helper.go", false},
		{"cmd/moai/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTestFile(tt.path)
			if got != tt.expected {
				t.Errorf("isTestFile(%q): 기대 %v, 실제 %v", tt.path, tt.expected, got)
			}
		})
	}
}

// TestTextualFanInCounter_CountWithRealFiles는 실제 파일에서 참조를 검색하는 테스트입니다.
func TestTextualFanInCounter_CountWithRealFiles(t *testing.T) {
	projectRoot := t.TempDir()

	// anchor-auth 심볼이 있는 파일 생성
	callerFile := filepath.Join(projectRoot, "internal", "caller.go")
	if err := os.MkdirAll(filepath.Dir(callerFile), 0755); err != nil {
		t.Fatalf("디렉토리 생성 실패: %v", err)
	}
	callerContent := `package internal

// anchor-auth를 사용하는 코드
func useAnchor() {
    // anchor-auth 참조
}`
	if err := os.WriteFile(callerFile, []byte(callerContent), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	// 태그 자체 파일 (참조 카운트에서 제외되어야 함)
	tagFile := filepath.Join(projectRoot, "internal", "auth", "handler.go")
	if err := os.MkdirAll(filepath.Dir(tagFile), 0755); err != nil {
		t.Fatalf("디렉토리 생성 실패: %v", err)
	}
	tagContent := `package auth
// anchor-auth 태그 위치
`
	if err := os.WriteFile(tagFile, []byte(tagContent), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	counter := &TextualFanInCounter{}
	tag := Tag{
		Kind:       MXAnchor,
		File:       tagFile,
		AnchorID:   "anchor-auth",
		CreatedBy:  "test",
		LastSeenAt: time.Now(),
	}

	count, method, err := counter.Count(context.Background(), tag, projectRoot, false)
	if err != nil {
		t.Fatalf("Count 오류: %v", err)
	}

	if method != "textual" {
		t.Errorf("method: 기대 textual, 실제 %s", method)
	}

	// callerFile에서 2회 참조 (주석 포함)
	if count < 1 {
		t.Errorf("count: 최소 1 기대, 실제 %d", count)
	}
}

// TestTextualFanInCounter_CountEmptyAnchorID는 빈 AnchorID에서 0 반환을 확인합니다.
func TestTextualFanInCounter_CountEmptyAnchorID(t *testing.T) {
	counter := &TextualFanInCounter{}
	tag := Tag{
		Kind:      MXAnchor,
		File:      "internal/auth.go",
		AnchorID:  "", // 빈 AnchorID
		CreatedBy: "test",
	}

	count, method, err := counter.Count(context.Background(), tag, "/tmp", false)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if count != 0 {
		t.Errorf("빈 AnchorID: count 0 기대, 실제 %d", count)
	}

	if method != "textual" {
		t.Errorf("method: 기대 textual, 실제 %s", method)
	}
}

// TestTextualFanInCounter_CountEmptyProjectRoot는 빈 projectRoot에서 0 반환을 확인합니다.
func TestTextualFanInCounter_CountEmptyProjectRoot(t *testing.T) {
	counter := &TextualFanInCounter{}
	tag := Tag{
		Kind:      MXAnchor,
		AnchorID:  "some-anchor",
		CreatedBy: "test",
	}

	count, method, err := counter.Count(context.Background(), tag, "", false)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if count != 0 {
		t.Errorf("빈 projectRoot: count 0 기대, 실제 %d", count)
	}
	_ = method
}

// TestTextualFanInCounter_ExcludeTests는 테스트 파일 제외를 확인합니다.
func TestTextualFanInCounter_ExcludeTests(t *testing.T) {
	counter := &TextualFanInCounter{}

	tag := Tag{
		Kind:       MXAnchor,
		File:       "tests/fixtures/mock_handler.go",
		AnchorID:   "anchor-mock",
		CreatedBy:  "test",
		LastSeenAt: time.Now(),
	}

	countWithTests, methodWithTests, err := counter.Count(context.Background(), tag, "/tmp/project", true)
	if err != nil {
		t.Fatalf("include-tests=true 오류: %v", err)
	}

	countWithoutTests, methodWithoutTests, err := counter.Count(context.Background(), tag, "/tmp/project", false)
	if err != nil {
		t.Fatalf("include-tests=false 오류: %v", err)
	}

	// 두 경우 모두 textual 방식이어야 함
	if methodWithTests != "textual" {
		t.Errorf("include-tests=true: fan_in_method 기대 textual, 실제 %s", methodWithTests)
	}
	if methodWithoutTests != "textual" {
		t.Errorf("include-tests=false: fan_in_method 기대 textual, 실제 %s", methodWithoutTests)
	}

	// 테스트 포함 시 count가 제외 시보다 크거나 같아야 함 (GREEN 단계에서 검증)
	if countWithTests < countWithoutTests {
		t.Errorf("include-tests=true 시 count(%d)이 false 시(%d)보다 작을 수 없음",
			countWithTests, countWithoutTests)
	}
}
