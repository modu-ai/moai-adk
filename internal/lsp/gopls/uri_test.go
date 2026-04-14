package gopls

import (
	"runtime"
	"strings"
	"testing"
)

// TestPathToURI_EncodesSpaces는 경로에 공백이 있을 때 퍼센트 인코딩하는지 검증한다.
func TestPathToURI_EncodesSpaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix 경로 테스트: Windows에서 건너뜁니다")
	}
	uri, err := pathToURI("/a b/main.go")
	if err != nil {
		t.Fatalf("pathToURI 오류: %v", err)
	}
	// 공백은 %20으로 인코딩되어야 한다.
	if !strings.Contains(uri, "%20") {
		t.Errorf("pathToURI(%q) = %q, %%20 인코딩을 기대했다", "/a b/main.go", uri)
	}
	if !strings.HasPrefix(uri, "file:///") {
		t.Errorf("pathToURI(%q) = %q, file:/// 접두사를 기대했다", "/a b/main.go", uri)
	}
}

// TestPathToURI_EncodesUnicode는 유니코드 경로를 퍼센트 인코딩하는지 검증한다.
func TestPathToURI_EncodesUnicode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix 경로 테스트: Windows에서 건너뜁니다")
	}
	uri, err := pathToURI("/내/main.go")
	if err != nil {
		t.Fatalf("pathToURI 오류: %v", err)
	}
	// 유니코드 문자는 퍼센트 인코딩되어야 한다.
	if !strings.Contains(uri, "%") {
		t.Errorf("pathToURI(%q) = %q, 퍼센트 인코딩을 기대했다", "/내/main.go", uri)
	}
	if !strings.HasPrefix(uri, "file:///") {
		t.Errorf("pathToURI(%q) = %q, file:/// 접두사를 기대했다", "/내/main.go", uri)
	}
}

// TestPathToURI_WindowsDrive는 Windows 드라이브 경로를 올바르게 처리하는지 검증한다.
// runtime.GOOS가 windows가 아니어도 로직이 동작해야 하므로 slash 변환 결과를 검증한다.
func TestPathToURI_WindowsDrive(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows 드라이브 경로 테스트: Windows에서만 실행합니다")
	}
	uri, err := pathToURI(`C:\x\main.go`)
	if err != nil {
		t.Fatalf("pathToURI 오류: %v", err)
	}
	// Windows: file:///C:/x/main.go 형식이어야 한다.
	want := "file:///C:/x/main.go"
	if uri != want {
		t.Errorf("pathToURI = %q, %q를 기대했다", uri, want)
	}
}

// TestPathToURI_Idempotent는 동일한 절대 경로를 두 번 변환해도 결과가 같은지 검증한다.
func TestPathToURI_Idempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix 경로 테스트: Windows에서 건너뜁니다")
	}
	path := "/usr/local/src/main.go"
	uri1, err := pathToURI(path)
	if err != nil {
		t.Fatalf("pathToURI 오류: %v", err)
	}
	uri2, err := pathToURI(path)
	if err != nil {
		t.Fatalf("pathToURI(2) 오류: %v", err)
	}
	if uri1 != uri2 {
		t.Errorf("pathToURI 멱등성 위반: %q != %q", uri1, uri2)
	}
}

// TestPathToURI_RejectsEmpty는 빈 경로를 거부하는지 검증한다.
func TestPathToURI_RejectsEmpty(t *testing.T) {
	_, err := pathToURI("")
	if err == nil {
		t.Error("빈 경로에 오류가 반환되지 않았다")
	}
}
