// convert-nextra-to-hextra: Bulk Nextra MDX → Hextra Markdown converter
// SPEC-DOCS-SITE-001 Phase 3
//
// 변환 규격:
//   T1 — Callout JSX → Hextra shortcode (735건)
//   T2 — _meta.ts → _meta.yaml (38개)
//   T3 — YAML frontmatter 주입 (219페이지)
//   T4 — .mdx → .md 확장자 변경
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MetaEntry: _meta.ts 한 항목
type MetaEntry struct {
	Key     string
	Title   string
	Type    string
	Display string
	Weight  int
}

// Stats: 변환 통계
type Stats struct {
	FilesProcessed  int
	CalloutsChanged int
	MetaConverted   int
	FrontmatterAdded int
	FilesRenamed    int
	Errors          []string
}

var (
	// nextra import 패턴 (여러 변형)
	reImportCallout = regexp.MustCompile(`(?m)^import\s*\{[^}]*Callout[^}]*\}\s*from\s*["']nextra/components["'];?\s*\n?`)
	reImportMeta    = regexp.MustCompile(`(?m)^import\s+type\s*\{[^}]*MetaRecord[^}]*\}\s*from\s*["']nextra["'];?\s*\n?`)

	// <Callout type="..."> 패턴 — 태그 시작 매핑
	reCalloutOpen  = regexp.MustCompile(`(?m)<Callout(\s+type\s*=\s*"([^"]*)")?\s*/?>`)
	reCalloutClose = regexp.MustCompile(`</Callout>`)

	// H1 추출
	reH1 = regexp.MustCompile(`(?m)^#\s+(.+)$`)

	// _meta.ts 파싱 — 라인 단위 파서에서 사용
	reMetaKeyStr = regexp.MustCompile(`^\s*"([^"]+)"\s*:\s*"([^"]+)"\s*,?\s*$`)
	reMetaKeyBare = regexp.MustCompile(`^\s*([a-zA-Z][a-zA-Z0-9_-]*)\s*:\s*"([^"]+)"\s*,?\s*$`)
	reObjTitle   = regexp.MustCompile(`title\s*:\s*"([^"]+)"`)
	reObjDisplay = regexp.MustCompile(`display\s*:\s*"([^"]+)"`)
	reObjType    = regexp.MustCompile(`\btype\s*:\s*"([^"]+)"`)
	// 객체 시작 라인: "key": { 또는 key: {
	reObjStart = regexp.MustCompile(`^\s*"?([a-zA-Z][a-zA-Z0-9_-]*)"?\s*:\s*\{`)
)

func main() {
	dryRun := flag.Bool("dry-run", false, "시뮬레이션 모드 — 파일 수정 없이 통계만 출력")
	contentDir := flag.String("content", "docs-site/content", "변환 대상 content 디렉토리 경로")
	flag.Parse()

	// 작업 루트 감지 (go run ./scripts/... 실행 위치)
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "프로젝트 루트 감지 실패: %v\n", err)
		os.Exit(1)
	}
	absContentDir := filepath.Join(root, *contentDir)

	fmt.Printf("변환 시작: %s\n", absContentDir)
	if *dryRun {
		fmt.Println("[DRY-RUN 모드] 파일을 수정하지 않습니다.")
	}

	stats := &Stats{}

	// 1단계: _meta.ts 파싱 (weight 맵 구성) + T2 변환
	// locale별로 처리 (각 locale 루트의 _meta.ts 먼저, 그 다음 섹션별)
	locales := []string{"ko", "en", "ja", "zh"}

	// (locale + dir) → []MetaEntry 순서 보존 맵
	weightMaps := make(map[string]map[string]int) // key: locale+"/"+relDir, val: filename→weight

	for _, locale := range locales {
		localeDir := filepath.Join(absContentDir, locale)
		err := filepath.Walk(localeDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || info.Name() != "_meta.ts" {
				return nil
			}
			dir := filepath.Dir(path)
			relDir, _ := filepath.Rel(absContentDir, dir)

			entries, parseErr := parseMetaTS(path)
			if parseErr != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("_meta.ts 파싱 실패 [%s]: %v", path, parseErr))
				return nil
			}

			// weight 맵 등록
			wmap := make(map[string]int)
			for _, e := range entries {
				wmap[e.Key] = e.Weight
			}
			weightMaps[relDir] = wmap

			// T2: _meta.yaml 생성
			if !*dryRun {
				yamlPath := filepath.Join(dir, "_meta.yaml")
				if writeErr := writeMetaYAML(entries, yamlPath); writeErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("_meta.yaml 쓰기 실패 [%s]: %v", yamlPath, writeErr))
					return nil
				}
				// 원본 _meta.ts 삭제
				if rmErr := os.Remove(path); rmErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("_meta.ts 삭제 실패 [%s]: %v", path, rmErr))
				}
			}
			stats.MetaConverted++
			return nil
		})
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("locale Walk 오류 [%s]: %v", locale, err))
		}
	}

	// 2단계: .mdx 파일 처리 (T1, T3, T4)
	for _, locale := range locales {
		localeDir := filepath.Join(absContentDir, locale)
		err := filepath.Walk(localeDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || filepath.Ext(path) != ".mdx" {
				return nil
			}

			// 이 파일의 디렉토리 기준 weight 맵 조회
			dir := filepath.Dir(path)
			relDir, _ := filepath.Rel(absContentDir, dir)
			wmap := weightMaps[relDir]

			// 파일명 → weight 계산
			baseName := strings.TrimSuffix(filepath.Base(path), ".mdx")
			weight := 99
			if wmap != nil {
				if w, ok := wmap[baseName]; ok {
					weight = w
				}
			}

			// 파일 읽기
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("읽기 실패 [%s]: %v", path, readErr))
				return nil
			}
			original := string(content)

			// T1: import 제거 + Callout 변환
			converted, calloutCount := convertContent(original)
			stats.CalloutsChanged += calloutCount

			// T3: frontmatter 주입
			title := extractH1Title(converted)
			if title == "" {
				title = toTitleCase(baseName)
			}
			if !hasFrontmatter(converted) {
				converted = buildFrontmatter(title, weight) + converted
				stats.FrontmatterAdded++
			}

			stats.FilesProcessed++

			if !*dryRun {
				// .mdx 파일에 변환된 내용 저장
				if writeErr := os.WriteFile(path, []byte(converted), 0644); writeErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("쓰기 실패 [%s]: %v", path, writeErr))
					return nil
				}

				// T4: .mdx → .md 이름 변경
				newPath := strings.TrimSuffix(path, ".mdx") + ".md"
				if renameErr := os.Rename(path, newPath); renameErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("이름 변경 실패 [%s]: %v", path, renameErr))
					return nil
				}
				stats.FilesRenamed++
			}

			return nil
		})
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("mdx Walk 오류 [%s]: %v", locale, err))
		}
	}

	// 결과 출력
	printReport(stats, *dryRun)

	if len(stats.Errors) > 0 {
		os.Exit(1)
	}
}

// findProjectRoot: go.mod 파일이 있는 루트 디렉토리를 찾는다
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// fallback: cwd 사용
	return cwd, nil
}

// parseMetaTS: _meta.ts 파일을 파싱하여 MetaEntry 슬라이스 반환
// brace-depth 카운팅으로 최상위 객체 본문을 추출한다.
func parseMetaTS(path string) ([]MetaEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)

	// 최상위 export 객체 위치 찾기: const meta = { ... }; 또는 export default { ... };
	// brace-depth 카운팅으로 중첩 객체 안전하게 처리
	body, extractErr := extractTopLevelObjectBody(content)
	if extractErr != nil {
		return nil, fmt.Errorf("메타 객체를 찾을 수 없음 [%s]: %w", path, extractErr)
	}

	return parseMetaBody(body), nil
}

// extractTopLevelObjectBody: TypeScript 소스에서 최상위 export 객체의 본문(내용)을 추출한다.
// brace-depth 카운팅으로 중첩 객체 안전하게 처리.
func extractTopLevelObjectBody(content string) (string, error) {
	// "const meta" 또는 "export default" 뒤에 오는 첫 번째 { 위치 찾기
	reObjEntryPoint := regexp.MustCompile(`(?:const\s+meta\s*(?::\s*\w+)?\s*=\s*|export\s+default\s+)\{`)
	loc := reObjEntryPoint.FindStringIndex(content)
	if loc == nil {
		return "", fmt.Errorf("export 객체를 찾을 수 없음")
	}

	// {부터 시작하여 brace-depth 카운팅
	startIdx := loc[1] - 1 // { 위치
	depth := 0
	for i := startIdx; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				// 최상위 } 찾음 — 내부 본문 반환
				return content[startIdx+1 : i], nil
			}
		}
	}
	return "", fmt.Errorf("객체 닫는 brace를 찾을 수 없음")
}

// parseMetaBody: 메타 객체 본문을 파싱하여 순서가 보존된 MetaEntry 슬라이스 반환
func parseMetaBody(body string) []MetaEntry {
	var entries []MetaEntry
	weight := 10

	scanner := bufio.NewScanner(strings.NewReader(body))
	var objLines []string
	var objKey string
	inObj := false
	braceDepth := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 주석 스킵
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		if inObj {
			objLines = append(objLines, line)
			for _, ch := range line {
				switch ch {
				case '{':
					braceDepth++
				case '}':
					braceDepth--
				}
			}
			if braceDepth <= 0 {
				// 객체 종료
				objContent := strings.Join(objLines, "\n")
				entry := parseObjectEntry(objKey, objContent)
				if entry != nil {
					entry.Weight = weight
					weight += 10
					entries = append(entries, *entry)
				}
				inObj = false
				objLines = nil
				objKey = ""
			}
			continue
		}

		// 단순 문자열 값 패턴: "key": "value"
		if m := reMetaKeyStr.FindStringSubmatch(trimmed); m != nil {
			entries = append(entries, MetaEntry{Key: m[1], Title: m[2], Weight: weight})
			weight += 10
			continue
		}
		// bare key 패턴: key: "value"
		if m := reMetaKeyBare.FindStringSubmatch(trimmed); m != nil {
			entries = append(entries, MetaEntry{Key: m[1], Title: m[2], Weight: weight})
			weight += 10
			continue
		}
		// 객체 시작 감지 (단일 라인 완결: "key": { ... } 또는 멀티라인 시작: "key": {)
		if mo := reObjStart.FindStringSubmatch(trimmed); mo != nil {
			objKey = mo[1]
			inObj = true
			// 초기 depth: 이 줄에서 { } 카운트
			for _, ch := range trimmed {
				switch ch {
				case '{':
					braceDepth++
				case '}':
					braceDepth--
				}
			}
			objLines = []string{trimmed}
			if braceDepth <= 0 {
				// 단일 라인 완결 객체
				objContent := strings.Join(objLines, "\n")
				entry := parseObjectEntry(objKey, objContent)
				if entry != nil {
					entry.Weight = weight
					weight += 10
					entries = append(entries, *entry)
				}
				inObj = false
				objLines = nil
				objKey = ""
				braceDepth = 0
			}
			continue
		}
	}

	return entries
}

// parseObjectEntry: 객체 값에서 title/display/type을 추출
func parseObjectEntry(key, objContent string) *MetaEntry {
	key = strings.Trim(key, `"`)
	e := &MetaEntry{Key: key}
	if m := reObjTitle.FindStringSubmatch(objContent); m != nil {
		e.Title = m[1]
	} else {
		e.Title = toTitleCase(key)
	}
	if m := reObjDisplay.FindStringSubmatch(objContent); m != nil {
		e.Display = m[1]
	}
	if m := reObjType.FindStringSubmatch(objContent); m != nil {
		e.Type = m[1]
	}
	return e
}

// writeMetaYAML: _meta.yaml 파일 생성
func writeMetaYAML(entries []MetaEntry, path string) error {
	var sb strings.Builder
	for _, e := range entries {
		// key 인용 여부 판단 (- 포함 시 인용)
		key := e.Key
		if strings.Contains(key, "-") {
			key = fmt.Sprintf(`"%s"`, key)
		}
		sb.WriteString(fmt.Sprintf("%s:\n", key))
		sb.WriteString(fmt.Sprintf("  title: %q\n", e.Title))
		if e.Display != "" {
			sb.WriteString(fmt.Sprintf("  display: %q\n", e.Display))
		}
		if e.Type != "" {
			sb.WriteString(fmt.Sprintf("  type: %q\n", e.Type))
		}
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// convertContent: Nextra import 제거 + Callout JSX → Hextra shortcode 변환
func convertContent(content string) (string, int) {
	// T1-1: import 라인 제거
	content = reImportCallout.ReplaceAllString(content, "")
	content = reImportMeta.ReplaceAllString(content, "")

	// T1-2: <Callout type="..."> 단일 태그 (self-closing or 여는 태그)
	calloutCount := 0

	// 여는 태그 변환: <Callout type="info"> → {{< callout type="info" >}}
	// <Callout> (타입 없음) → {{< callout >}}
	content = reCalloutOpen.ReplaceAllStringFunc(content, func(match string) string {
		calloutCount++
		sub := reCalloutOpen.FindStringSubmatch(match)
		typeVal := ""
		if len(sub) >= 3 && sub[2] != "" {
			rawType := sub[2]
			typeVal = normalizeCalloutType(rawType)
		}
		if typeVal == "" {
			return "{{< callout >}}"
		}
		return fmt.Sprintf(`{{< callout type="%s" >}}`, typeVal)
	})

	// 닫는 태그 변환: </Callout> → {{< /callout >}}
	content = reCalloutClose.ReplaceAllString(content, "{{< /callout >}}")

	// 연속된 빈 줄 정리 (import 제거 후 빈 줄 3개 이상 → 2개로)
	reTripleBlank := regexp.MustCompile(`\n{3,}`)
	content = reTripleBlank.ReplaceAllString(content, "\n\n")

	return content, calloutCount
}

// normalizeCalloutType: Hextra 지원 타입으로 정규화
// Hextra 기본 지원: info, warning, error (default 타입은 type 없이)
func normalizeCalloutType(t string) string {
	switch t {
	case "info":
		return "info"
	case "warning":
		return "warning"
	case "error":
		return "error"
	case "tip":
		// tip → info (Hextra에서 tip은 별도 타입이 아님, info로 매핑)
		return "info"
	case "success":
		// success → info fallback
		return "info"
	default:
		// 알 수 없는 타입 → default (타입 속성 없이)
		return ""
	}
}

// extractH1Title: 콘텐츠에서 첫 번째 H1 헤더 텍스트 추출
func extractH1Title(content string) string {
	m := reH1.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	// 인라인 마크다운 제거 (볼드, 이탤릭, 코드)
	title := m[1]
	title = regexp.MustCompile(`\*+`).ReplaceAllString(title, "")
	title = regexp.MustCompile("`"+"[^`]+`").ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	return title
}

// hasFrontmatter: 파일이 이미 frontmatter를 가지고 있는지 확인
func hasFrontmatter(content string) bool {
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "---")
}

// buildFrontmatter: YAML frontmatter 문자열 생성
func buildFrontmatter(title string, weight int) string {
	// 제목의 특수문자 이스케이프 (: 포함 시 인용)
	titleStr := title
	if strings.ContainsAny(titleStr, `:"{}[]|>&*!`) {
		titleStr = fmt.Sprintf(`"%s"`, strings.ReplaceAll(titleStr, `"`, `\"`))
	}
	return fmt.Sprintf("---\ntitle: %s\nweight: %d\ndraft: false\n---\n", titleStr, weight)
}

// toTitleCase: kebab-case → Title Case 변환
func toTitleCase(s string) string {
	words := strings.Split(s, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	result := strings.Join(words, " ")
	// 인덱스 파일 처리
	if result == "Index" {
		result = "Overview"
	}
	return result
}

// printReport: 변환 결과 리포트 출력
func printReport(stats *Stats, dryRun bool) {
	mode := "실행"
	if dryRun {
		mode = "DRY-RUN"
	}
	fmt.Printf("\n========== 변환 결과 [%s] ==========\n", mode)
	fmt.Printf("처리된 MDX 파일:    %d\n", stats.FilesProcessed)
	fmt.Printf("Callout 변환:       %d\n", stats.CalloutsChanged)
	fmt.Printf("_meta.ts 변환:      %d\n", stats.MetaConverted)
	fmt.Printf("Frontmatter 주입:   %d\n", stats.FrontmatterAdded)
	if !dryRun {
		fmt.Printf("파일명 변경 (.md):  %d\n", stats.FilesRenamed)
	}

	// 검증 힌트
	if stats.CalloutsChanged != 735 {
		fmt.Printf("\n[경고] Callout 변환 수가 예상(735)과 다릅니다: %d\n", stats.CalloutsChanged)
	} else {
		fmt.Printf("\n[OK] Callout 735건 정확히 변환\n")
	}
	if stats.MetaConverted != 38 {
		fmt.Printf("[경고] _meta.ts 변환 수가 예상(38)과 다릅니다: %d\n", stats.MetaConverted)
	} else {
		fmt.Printf("[OK] _meta.ts 38개 변환\n")
	}
	if stats.FilesProcessed != 219 {
		fmt.Printf("[경고] 처리 파일 수가 예상(219)과 다릅니다: %d\n", stats.FilesProcessed)
	} else {
		fmt.Printf("[OK] MDX 219페이지 처리\n")
	}

	if len(stats.Errors) > 0 {
		fmt.Printf("\n오류 목록 (%d건):\n", len(stats.Errors))
		sort.Strings(stats.Errors)
		for _, e := range stats.Errors {
			fmt.Printf("  - %s\n", e)
		}
	} else {
		fmt.Println("\n[OK] 오류 0건")
	}
}
