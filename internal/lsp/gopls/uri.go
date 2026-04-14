package gopls

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// pathToURI는 로컬 파일 경로를 RFC 3986 준수 file:// URI로 변환한다.
//
// LSP 사양에 따라 파일 URI는 경로의 공백, 유니코드 문자를 퍼센트 인코딩해야 한다.
// 직접 "file://" + path 문자열 결합은 다음 케이스에서 실패한다:
//   - 공백 포함 경로 (/Users/goos/My Project/) — LSP 사양 위반
//   - 유니코드 경로 (/내 프로젝트/main.go) — didOpen 무음 실패
//   - Windows 드라이브 경로 (C:\...) — 슬래시 개수 오류
//
// 변환 규칙:
//   - 상대 경로는 filepath.Abs로 절대 경로로 변환한다.
//   - Windows: 백슬래시를 슬래시로 변환하고 /C:/... 형식으로 정규화한다.
//   - Unix: /path/to/file → file:///path/to/file
//
// @MX:ANCHOR: [AUTO] LSP URI 변환 핵심 헬퍼 — bridge.go의 initialize, GetDiagnostics에서 호출된다
// @MX:REASON: fan_in >= 3 (initialize, GetDiagnostics, 테스트)
func pathToURI(absPath string) (string, error) {
	if absPath == "" {
		return "", fmt.Errorf("gopls: uri: 빈 경로는 허용되지 않는다")
	}

	// 상대 경로를 절대 경로로 변환한다.
	abs, err := filepath.Abs(absPath)
	if err != nil {
		return "", fmt.Errorf("gopls: uri: 절대 경로 변환 실패 %q: %w", absPath, err)
	}

	// filepath.ToSlash로 OS 경로 구분자를 슬래시로 통일한다.
	slashed := filepath.ToSlash(abs)

	// Windows 드라이브 경로(C:/...) 처리:
	// url.URL.String()은 Path가 /C:/...로 시작해야 file:///C:/... 형식을 생성한다.
	// Unix에서는 abs가 이미 /로 시작하므로 조정 불필요.
	if runtime.GOOS == "windows" && len(slashed) >= 2 && slashed[1] == ':' {
		// "C:/..." → "/C:/..."
		slashed = "/" + slashed
	}

	// url.URL을 사용하여 경로를 퍼센트 인코딩한다.
	// Scheme만 지정하고 Path는 url.URL이 자동으로 인코딩한다.
	u := &url.URL{
		Scheme: "file",
		// Host를 빈 문자열로 두면 file:///path 형식이 생성된다.
		Path: slashed,
	}

	// url.URL.String()은 "file:///path%20with%20space/main.go" 형식으로 반환한다.
	result := u.String()

	// 슬래시 트리플 검증: file:// + /path 는 file:///path가 된다.
	// url.URL은 Host가 비어있으면 authority를 생략하므로 file:/path 가 될 수 있다.
	// 명시적으로 file:///를 보장한다.
	if !strings.HasPrefix(result, "file:///") {
		// file://path → file:///path로 수정
		if strings.HasPrefix(result, "file://") {
			result = "file:///" + strings.TrimPrefix(result, "file://")
		}
	}

	return result, nil
}
