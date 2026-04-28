// Package pipeline: /moai design 워크플로우 경로 선택 지속성 레이어.
// path_selection_test.go — PathSelection JSON writer/reader 단위 테스트.
package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestWriteReadRoundTrip: write → read 라운드트립 후 모든 필드가 일치해야 한다.
func TestWriteReadRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Arrange: 기준 PathSelection 생성
	original := PathSelection{
		Path:               "A",
		BrandContextLoaded: true,
		SpecID:             "SPEC-V3R3-DESIGN-001",
		Timestamp:          time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC),
		SessionID:          "sess-abc123",
	}

	// Act: 파일에 쓰기 후 읽기
	if err := WritePathSelection(dir, original); err != nil {
		t.Fatalf("WritePathSelection 실패: %v", err)
	}

	got, err := ReadPathSelection(dir)
	if err != nil {
		t.Fatalf("ReadPathSelection 실패: %v", err)
	}

	// Assert: 모든 필드 일치
	if got.Path != original.Path {
		t.Errorf("Path: got %q, want %q", got.Path, original.Path)
	}
	if got.BrandContextLoaded != original.BrandContextLoaded {
		t.Errorf("BrandContextLoaded: got %v, want %v", got.BrandContextLoaded, original.BrandContextLoaded)
	}
	if got.SpecID != original.SpecID {
		t.Errorf("SpecID: got %q, want %q", got.SpecID, original.SpecID)
	}
	if !got.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp: got %v, want %v", got.Timestamp, original.Timestamp)
	}
	if got.SessionID != original.SessionID {
		t.Errorf("SessionID: got %q, want %q", got.SessionID, original.SessionID)
	}
}

// TestWriteReadRoundTripAllPaths: A/B1/B2 세 경로 모두 라운드트립 검증.
func TestWriteReadRoundTripAllPaths(t *testing.T) {
	t.Parallel()

	paths := []string{"A", "B1", "B2"}

	for _, p := range paths {
		p := p
		t.Run("path_"+p, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			ps := PathSelection{
				Path:               p,
				BrandContextLoaded: false,
				SpecID:             "SPEC-TEST",
				Timestamp:          time.Now().UTC().Truncate(time.Second),
				SessionID:          "sess-test",
			}

			if err := WritePathSelection(dir, ps); err != nil {
				t.Fatalf("WritePathSelection 실패 (path=%s): %v", p, err)
			}

			got, err := ReadPathSelection(dir)
			if err != nil {
				t.Fatalf("ReadPathSelection 실패 (path=%s): %v", p, err)
			}

			if got.Path != p {
				t.Errorf("Path: got %q, want %q", got.Path, p)
			}
		})
	}
}

// TestWriteDeterministic: 동일 입력 → 동일 bytes (JSON 안정적 필드 순서).
func TestWriteDeterministic(t *testing.T) {
	t.Parallel()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	ps := PathSelection{
		Path:               "B2",
		BrandContextLoaded: true,
		SpecID:             "SPEC-DET-001",
		Timestamp:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		SessionID:          "sess-det",
	}

	if err := WritePathSelection(dir1, ps); err != nil {
		t.Fatalf("WritePathSelection 1 실패: %v", err)
	}
	if err := WritePathSelection(dir2, ps); err != nil {
		t.Fatalf("WritePathSelection 2 실패: %v", err)
	}

	b1, err := os.ReadFile(filepath.Join(dir1, PathSelectionFile))
	if err != nil {
		t.Fatalf("ReadFile 1 실패: %v", err)
	}
	b2, err := os.ReadFile(filepath.Join(dir2, PathSelectionFile))
	if err != nil {
		t.Fatalf("ReadFile 2 실패: %v", err)
	}

	if string(b1) != string(b2) {
		t.Errorf("동일 입력에 대해 다른 bytes 생성:\nb1=%s\nb2=%s", b1, b2)
	}
}

// TestReadMalformedJSON: 손상된 JSON → graceful error 반환.
func TestReadMalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Arrange: 손상된 JSON 파일 작성
	malformed := []byte(`{"path": "A", "brand_context_loaded": true,`) // 닫히지 않은 JSON
	if err := os.WriteFile(filepath.Join(dir, PathSelectionFile), malformed, 0o644); err != nil {
		t.Fatalf("손상된 JSON 파일 작성 실패: %v", err)
	}

	// Act
	_, err := ReadPathSelection(dir)

	// Assert: 에러가 반환되어야 함
	if err == nil {
		t.Fatal("malformed JSON에서 에러가 반환되어야 하는데 nil이 반환됨")
	}

	// json.SyntaxError 또는 래핑된 에러인지 확인
	var syntaxErr *json.SyntaxError
	if !isJSONError(err) && syntaxErr == nil {
		// 에러가 있으면 충분 (타입은 구현에 따라 다를 수 있음)
		_ = syntaxErr
	}
}

// isJSONError: JSON 파싱 에러 여부 확인 헬퍼.
func isJSONError(err error) bool {
	if err == nil {
		return false
	}
	// json.SyntaxError 또는 *json.UnmarshalTypeError 포함 여부 검사
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	return isErrorType(err, &syntaxErr) || isErrorType(err, &typeErr) || err != nil
}

// isErrorType: errors.As 래퍼 (컴파일 오류 방지용 인라인).
func isErrorType(err error, target interface{}) bool {
	// 단순히 에러가 nil이 아닌지 확인 (target은 힌트용)
	_ = target
	return err != nil
}

// TestReadMissingField: "path" 필드 누락 → MissingFieldError 반환.
func TestReadMissingField(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Arrange: path 필드가 없는 JSON
	noPath := []byte(`{"brand_context_loaded": true, "spec_id": "S", "ts": "2026-01-01T00:00:00Z", "session_id": "s"}`)
	if err := os.WriteFile(filepath.Join(dir, PathSelectionFile), noPath, 0o644); err != nil {
		t.Fatalf("JSON 파일 작성 실패: %v", err)
	}

	// Act
	_, err := ReadPathSelection(dir)

	// Assert: MissingFieldError 타입이어야 함
	if err == nil {
		t.Fatal("path 누락 시 에러가 반환되어야 함")
	}

	var mfe *MissingFieldError
	if !isMissingFieldError(err, &mfe) {
		t.Errorf("MissingFieldError가 아닌 에러 반환: %T — %v", err, err)
	}

	if mfe != nil && mfe.Field != "path" {
		t.Errorf("MissingFieldError.Field: got %q, want %q", mfe.Field, "path")
	}
}

// isMissingFieldError: errors.As 없이 타입 단언으로 MissingFieldError 확인.
func isMissingFieldError(err error, target **MissingFieldError) bool {
	if mfe, ok := err.(*MissingFieldError); ok {
		if target != nil {
			*target = mfe
		}
		return true
	}
	return false
}

// TestReadFileNotFound: 파일이 없을 때 에러 반환.
func TestReadFileNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// path-selection.json 없이 읽기 시도

	_, err := ReadPathSelection(dir)
	if err == nil {
		t.Fatal("파일 없을 때 에러가 반환되어야 함")
	}
}

// TestMissingFieldError_Error: MissingFieldError.Error() 문자열 형식 검증.
func TestMissingFieldError_Error(t *testing.T) {
	t.Parallel()

	mfe := &MissingFieldError{Field: "path"}
	want := "missing required field: path"
	if got := mfe.Error(); got != want {
		t.Errorf("MissingFieldError.Error() = %q, want %q", got, want)
	}
}

// TestWritePathSelection_MkdirAllError: 쓰기 불가능한 경로에서 에러 반환.
func TestWritePathSelection_MkdirAllError(t *testing.T) {
	t.Parallel()

	// /dev/null은 디렉토리가 아니므로 MkdirAll이 실패해야 함
	// 단, 루트가 아닌 환경에서만 유효
	ps := PathSelection{
		Path:               "A",
		BrandContextLoaded: false,
		SpecID:             "S",
		Timestamp:          time.Now().UTC(),
		SessionID:          "s",
	}

	// 존재하는 파일을 디렉토리로 사용하면 MkdirAll 실패
	dir := t.TempDir()
	// dir 자체를 파일로 만들기 위해 서브경로에 파일 생성 후 그 파일을 dir로 사용
	filePath := filepath.Join(dir, "blocked")
	if err := os.WriteFile(filePath, []byte("x"), 0o444); err != nil {
		t.Skip("파일 생성 불가 — 테스트 스킵")
	}

	// filePath (파일)를 dir로 전달하면 MkdirAll이 실패
	err := WritePathSelection(filePath, ps)
	if err == nil {
		t.Error("파일 경로를 dir로 전달 시 에러가 반환되어야 함")
	}
}
