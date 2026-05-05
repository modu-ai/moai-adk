package mx

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

// FanInCounter는 @MX:ANCHOR 태그의 코드 참조 수를 계산하는 인터페이스입니다.
// LSP 기반 구현과 텍스트 폴백 구현을 동일한 인터페이스로 추상화합니다 (REQ-SPC-004-003).
type FanInCounter interface {
	// Count는 주어진 태그의 fan-in(코드 참조 수)을 계산합니다.
	// excludeTests가 true이면 테스트 파일의 참조는 제외합니다 (REQ-SPC-004-040).
	// 반환값: count (참조 수), method ("lsp" 또는 "textual"), err
	Count(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (count int, method string, err error)
}

// TextualFanInCounter는 텍스트 기반 grep 방식으로 fan-in을 계산하는 구현체입니다.
// LSP 서버가 없는 언어에 대한 폴백으로 사용됩니다 (REQ-SPC-004-020).
//
// @MX:WARN: [AUTO] TextualFanInCounter — 텍스트 검색 방식은 문자열/주석의 오탐(false positive) 위험이 있습니다
// @MX:REASON: 소스 코드가 아닌 문자열 리터럴이나 주석에서도 심볼 이름이 발견될 수 있어 fan-in이 과대 계산될 수 있습니다
type TextualFanInCounter struct {
	// ProjectRoot는 프로젝트 루트 디렉토리 경로입니다.
	ProjectRoot string
}

// isTestFile는 파일 경로가 테스트 파일인지 확인합니다 (REQ-SPC-004-040).
// 테스트 파일 판별 기준: _test.go 접미사 또는 tests/, fixtures/ 디렉토리 하위.
func isTestFile(filePath string) bool {
	base := filepath.Base(filePath)
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	// 경로에 tests/ 또는 fixtures/ 디렉토리가 포함되면 테스트 파일로 간주
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for _, part := range parts {
		if part == "tests" || part == "fixtures" || part == "testdata" {
			return true
		}
	}
	return false
}

// Count는 텍스트 검색을 통해 AnchorID의 참조 수를 계산합니다.
// 결과의 fan_in_method는 항상 "textual"입니다.
func (c *TextualFanInCounter) Count(_ context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	if tag.AnchorID == "" {
		return 0, "textual", nil
	}

	if projectRoot == "" {
		projectRoot = c.ProjectRoot
	}

	if projectRoot == "" {
		return 0, "textual", nil
	}

	count := 0

	// 프로젝트 루트를 재귀적으로 스캔하여 AnchorID 참조를 검색
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 오류 무시하고 계속
		}

		if info.IsDir() {
			// vendor, node_modules 등 제외
			base := filepath.Base(path)
			if base == "vendor" || base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// 태그 자체 파일 제외 (자기 참조)
		if path == tag.File {
			return nil
		}

		// 테스트 파일 제외 (excludeTests=true인 경우)
		if excludeTests && isTestFile(path) {
			return nil
		}

		// 파일에서 AnchorID 참조 검색
		refs := countReferencesInFile(path, tag.AnchorID)
		count += refs
		return nil
	})

	if err != nil {
		return 0, "textual", nil
	}

	return count, "textual", nil
}

// countReferencesInFile는 파일에서 심볼 이름의 등장 횟수를 셉니다.
// 단순 문자열 검색이므로 주석/문자열 내 오탐 가능성이 있습니다.
func countReferencesInFile(filePath, symbol string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, symbol) {
			count++
		}
	}
	return count
}
