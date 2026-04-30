package mx

import (
	"context"
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
