package mx

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// specFrontmatter는 spec.md YAML frontmatter의 파싱 대상 필드들입니다.
type specFrontmatter struct {
	ID     string      `yaml:"id"`
	Module interface{} `yaml:"module"`
}

// LoadSpecModules는 projectRoot/.moai/specs/*/spec.md 를 순회하여
// SPEC ID → []modulePath 맵을 반환합니다 (REQ-SPC-004-005).
//
// @MX:ANCHOR: [AUTO] LoadSpecModules — SPEC 모듈 경로 로더; CLI, Resolver, SpecAssociator 모두 이 함수를 통해 경로 기반 연결을 설정
// @MX:REASON: fan_in >= 3 — CLI mx_query.go, Resolver 초기화 경로, 향후 codemaps 생성기 모두 호출
//
// module 필드 지원 형식:
//   - 문자열: "internal/mx/, cmd/moai/" → 쉼표 분리 + TrimSpace
//   - YAML 시퀀스: [internal/foo/, internal/bar/] → as-is
//   - 빈 문자열: "" → 빈 슬라이스
//
// .moai/specs/ 디렉터리가 없으면 에러 없이 빈 맵을 반환합니다.
func LoadSpecModules(projectRoot string) (map[string][]string, error) {
	specsDir := filepath.Join(projectRoot, ".moai", "specs")

	if _, err := os.Stat(specsDir); errors.Is(err, os.ErrNotExist) {
		return map[string][]string{}, nil
	}

	pattern := filepath.Join(specsDir, "*", "spec.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// 결정적 순서를 위해 정렬
	sort.Strings(matches)

	result := make(map[string][]string)

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		fm, err := parseFrontmatter(data)
		if err != nil || fm.ID == "" {
			continue
		}

		result[fm.ID] = parseModuleField(fm.Module)
	}

	return result, nil
}

// parseFrontmatter는 spec.md 파일의 YAML frontmatter(--- ... --- 사이)를 파싱합니다.
func parseFrontmatter(data []byte) (specFrontmatter, error) {
	content := string(data)

	// --- 로 시작하는 frontmatter 추출
	if !strings.HasPrefix(content, "---") {
		return specFrontmatter{}, nil
	}

	// 첫 번째 --- 이후의 내용에서 두 번째 --- 를 찾음
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		// 닫는 --- 없음: 파일 끝까지 frontmatter로 처리
		end = len(rest)
	}

	yamlContent := rest[:end]

	var fm specFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return specFrontmatter{}, err
	}

	return fm, nil
}

// parseModuleField는 module 필드 값을 []string 으로 변환합니다.
// 지원 타입:
//   - string: 쉼표 구분 → split + TrimSpace, 빈 문자열 항목 제거
//   - []interface{}: 각 요소를 string으로 캐스팅
func parseModuleField(v interface{}) []string {
	if v == nil {
		return []string{}
	}

	switch val := v.(type) {
	case string:
		if val == "" {
			return []string{}
		}
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result

	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result

	default:
		return []string{}
	}
}
