package mx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockFanInCounter is a test mock implementation of FanInCounter.
type mockFanInCounter struct {
	// counts maps AnchorID -> fan-in count.
	counts map[string]int
	// method is the fan-in calculation method to return.
	method string
	// err is the error to return.
	err error
}

// Count returns the preconfigured fan-in value.
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

// TestTextualFanInCounter_Count tests text-based fan-in counting.
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

	// Method must be textual
	if method != "textual" {
		t.Errorf("fan_in_method: 기대 'textual', 실제 '%s'", method)
	}

	// RED phase: count returns 0 (stub)
	if count < 0 {
		t.Errorf("fan-in 수는 음수일 수 없음: %d", count)
	}
}

// TestMockFanInCounter verifies that the mock implementation conforms to the interface.
func TestMockFanInCounter(t *testing.T) {
	var _ FanInCounter = &mockFanInCounter{}
}

// TestFanInCounter_InterfaceCompliance verifies FanInCounter interface compliance.
func TestFanInCounter_InterfaceCompliance(t *testing.T) {
	var _ FanInCounter = &TextualFanInCounter{}
}

// TestIsTestFile tests the test-file discriminator function.
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

// TestTextualFanInCounter_CountWithRealFiles tests reference search across real files.
func TestTextualFanInCounter_CountWithRealFiles(t *testing.T) {
	projectRoot := t.TempDir()

	// Create a file containing the anchor-auth symbol
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

	// The tag's own file (must be excluded from the reference count)
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

	// 2 references in callerFile (including comments)
	if count < 1 {
		t.Errorf("count: 최소 1 기대, 실제 %d", count)
	}
}

// TestTextualFanInCounter_CountEmptyAnchorID verifies that an empty AnchorID returns 0.
func TestTextualFanInCounter_CountEmptyAnchorID(t *testing.T) {
	counter := &TextualFanInCounter{}
	tag := Tag{
		Kind:      MXAnchor,
		File:      "internal/auth.go",
		AnchorID:  "", // empty AnchorID
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

// TestTextualFanInCounter_CountEmptyProjectRoot verifies that an empty projectRoot returns 0.
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

// TestIsTestFile_UserPattern_IntegrationDir uses a user glob pattern to classify an integration directory as a test file.
// AC-SPC-004-11: user-defined TestPaths glob support (G-05).
func TestIsTestFile_UserPattern_IntegrationDir(t *testing.T) {
	// isTestFileWithPatterns("internal/foo/integration/bar.go", []string{"**/integration/**"}) -> true
	got := isTestFileWithPatterns("internal/foo/integration/bar.go", []string{"**/integration/**"})
	if !got {
		t.Errorf("isTestFileWithPatterns: **/integration/** 패턴으로 integration 디렉토리 파일을 테스트 파일로 판별해야 함")
	}
}

// TestIsTestFile_UserPattern_NoMatch_FallbackHardcoded tests the hardcoded fallback when a user pattern does not match.
// AC-SPC-004-11: hardcoded _test.go fallback when the user pattern does not match (G-05).
func TestIsTestFile_UserPattern_NoMatch_FallbackHardcoded(t *testing.T) {
	// isTestFileWithPatterns("internal/foo/foo_test.go", []string{"**/integration/**"}) -> true (hardcoded _test.go fallback)
	got := isTestFileWithPatterns("internal/foo/foo_test.go", []string{"**/integration/**"})
	if !got {
		t.Errorf("isTestFileWithPatterns: 사용자 패턴 불일치 시 _test.go 하드코딩 폴백으로 true 반환해야 함")
	}
}

// TestTextualFanInCounter_RespectsUserTestPaths tests the TextualFanInCounter.TestPaths field.
// AC-SPC-004-11: inject user glob patterns via the TestPaths field (G-06).
func TestTextualFanInCounter_RespectsUserTestPaths(t *testing.T) {
	projectRoot := t.TempDir()

	// Create an anchor-referencing file in the integration directory (must be treated as a test file)
	integrationFile := filepath.Join(projectRoot, "internal", "myfeature", "integration", "anchor_caller.go")
	if err := os.MkdirAll(filepath.Dir(integrationFile), 0755); err != nil {
		t.Fatalf("디렉토리 생성 실패: %v", err)
	}
	integrationContent := "package integration\n// anchor-test-userpath 참조\n"
	if err := os.WriteFile(integrationFile, []byte(integrationContent), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	// Create a regular file (not a test file)
	regularFile := filepath.Join(projectRoot, "internal", "myfeature", "foo.go")
	if err := os.MkdirAll(filepath.Dir(regularFile), 0755); err != nil {
		t.Fatalf("디렉토리 생성 실패: %v", err)
	}
	regularContent := "package myfeature\n// anchor-test-userpath 참조\n"
	if err := os.WriteFile(regularFile, []byte(regularContent), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	// Anchor tag file (excluded from the reference count)
	tagFile := filepath.Join(projectRoot, "internal", "myfeature", "anchor.go")
	if err := os.WriteFile(tagFile, []byte("package myfeature\n"), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	tag := Tag{
		Kind:       MXAnchor,
		File:       tagFile,
		AnchorID:   "anchor-test-userpath",
		CreatedBy:  "test",
		LastSeenAt: time.Now(),
	}

	// With TestPaths: integration file excluded -> count=1 (only foo.go)
	counterWithPaths := &TextualFanInCounter{
		TestPaths: []string{"**/integration/**"},
	}
	countWithPaths, _, err := counterWithPaths.Count(context.Background(), tag, projectRoot, true)
	if err != nil {
		t.Fatalf("TestPaths 있을 때 오류: %v", err)
	}

	// Without TestPaths: integration file included -> count=2 (foo.go + integration/anchor_caller.go)
	counterNoPaths := &TextualFanInCounter{}
	countNoPaths, _, err := counterNoPaths.Count(context.Background(), tag, projectRoot, false)
	if err != nil {
		t.Fatalf("TestPaths 없을 때 오류: %v", err)
	}

	if countWithPaths >= countNoPaths {
		t.Errorf("TestPaths 있을 때(%d) < 없을 때(%d) 기대 (integration 파일 제외 효과)",
			countWithPaths, countNoPaths)
	}

	if countWithPaths != 1 {
		t.Errorf("TestPaths 있을 때 count=1 기대 (foo.go만), 실제 %d", countWithPaths)
	}

	if countNoPaths != 2 {
		t.Errorf("TestPaths 없을 때 count=2 기대 (foo.go + integration/anchor_caller.go), 실제 %d", countNoPaths)
	}
}

// TestTextualFanInCounter_ExcludeTests verifies exclusion of test files.
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

	// Both cases must use the textual method
	if methodWithTests != "textual" {
		t.Errorf("include-tests=true: fan_in_method 기대 textual, 실제 %s", methodWithTests)
	}
	if methodWithoutTests != "textual" {
		t.Errorf("include-tests=false: fan_in_method 기대 textual, 실제 %s", methodWithoutTests)
	}

	// When tests are included, the count must be greater than or equal to when excluded (verified in GREEN phase)
	if countWithTests < countWithoutTests {
		t.Errorf("include-tests=true 시 count(%d)이 false 시(%d)보다 작을 수 없음",
			countWithTests, countWithoutTests)
	}
}
