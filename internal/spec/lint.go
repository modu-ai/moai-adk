// Package spec은 SPEC 문서 파싱 및 검증 기능을 제공한다.
// lint.go는 moai spec lint CLI의 핵심 엔진으로, Rule 인터페이스와
// Linter 구조체를 통해 SPEC 문서의 EARS 준수성, 커버리지, DAG 등을 검증한다.
package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// Severity는 finding의 심각도를 나타낸다.
type Severity string

const (
	// SeverityError는 linter가 비정상 종료를 유발하는 치명적 오류이다.
	SeverityError Severity = "error"
	// SeverityWarning은 기본 모드에서는 종료 코드에 영향을 주지 않는 경고이다.
	// --strict 플래그 사용 시 error로 승격된다.
	SeverityWarning Severity = "warning"
	// SeverityInfo는 정보성 메시지이다.
	SeverityInfo Severity = "info"
)

// Finding은 linter가 발견한 단일 문제를 나타낸다.
// JSON 직렬화 시 file, line, severity, code, message 필드가 포함된다.
type Finding struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
}

// Report는 lint 실행 결과를 나타낸다.
type Report struct {
	// Findings는 모든 finding의 목록이다.
	Findings []Finding
	// Strict는 --strict 플래그 상태이다. HasErrors() 계산에 영향을 준다.
	Strict bool
}

// HasErrors는 비정상 종료가 필요한 상태인지 반환한다.
// strict 모드에서는 warning도 error로 간주한다.
func (r *Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityError {
			return true
		}
		if r.Strict && f.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// ToJSON은 findings를 JSON 바이트 슬라이스로 직렬화한다.
// findings가 nil인 경우 빈 JSON 배열([])을 반환한다.
func (r *Report) ToJSON() ([]byte, error) {
	findings := r.Findings
	if findings == nil {
		findings = []Finding{}
	}
	return json.Marshal(findings)
}

// ToSARIF는 findings를 SARIF 2.1.0 형식의 JSON 바이트 슬라이스로 직렬화한다.
func (r *Report) ToSARIF() ([]byte, error) {
	return marshalSARIF(r.Findings)
}

// LinterOptions는 Linter 생성 옵션이다.
type LinterOptions struct {
	// RegistryPath는 zone registry 마크다운 파일 경로이다.
	// 비어 있으면 DanglingRuleReference 검사를 건너뛴다.
	RegistryPath string
	// BaseDir는 no-args 실행 시 SPEC 파일을 탐색하는 기준 디렉토리이다.
	// 의존성 SPEC 존재 확인에도 사용된다.
	BaseDir string
	// Strict는 --strict 플래그 상태이다.
	Strict bool
}

// Linter는 SPEC 문서를 검증하는 엔진이다.
//
// @MX:ANCHOR: [AUTO] Linter is the central lint engine; all lint rules are dispatched through it.
// @MX:REASON: Fan-in hub — CLI, tests, and future integrations all call Linter.Lint.
type Linter struct {
	opts     LinterOptions
	registry *constitution.Registry
	rules    []Rule
}

// NewLinter는 새로운 Linter 인스턴스를 생성한다.
// options.RegistryPath가 지정된 경우 zone registry를 로드한다.
func NewLinter(opts LinterOptions) *Linter {
	l := &Linter{opts: opts}

	// zone registry 로드
	if opts.RegistryPath != "" {
		projectDir := opts.BaseDir
		if projectDir == "" {
			projectDir = "."
		}
		reg, err := constitution.LoadRegistry(opts.RegistryPath, projectDir)
		if err == nil {
			l.registry = reg
		}
		// registry 로드 실패는 silent — DanglingRuleReference 체크를 건너뜀
	}

	// 규칙 등록
	l.rules = []Rule{
		&EARSModalityRule{},
		&REQIDUniquenessRule{},
		&CoverageRule{},
		&FrontmatterSchemaRule{},
		&DependencyExistsRule{},
		&OutOfScopeRule{},
		&BreakingChangeIDRule{},
		// cross-SPEC 규칙
		&DependencyCycleRule{},
		&DuplicateSPECIDRule{},
		// registry 필요
		&ZoneRegistryRule{registry: l.registry},
	}

	return l
}

// Lint는 지정된 SPEC 파일 경로들을 검증한다.
// paths가 nil이거나 비어 있으면 opts.BaseDir 아래의 spec.md 파일들을 자동 탐색한다.
//
// @MX:ANCHOR: [AUTO] Lint is the primary entry point; orchestrates rule execution across all SPECs.
// @MX:REASON: Fan-in hub — all callers (CLI, tests) go through this method.
func (l *Linter) Lint(paths []string) (*Report, error) {
	// 경로 미지정 시 BaseDir에서 자동 탐색
	if len(paths) == 0 {
		discovered, err := discoverSPECs(l.opts.BaseDir)
		if err != nil {
			return nil, fmt.Errorf("SPEC 탐색 실패: %w", err)
		}
		paths = discovered
	}

	// SPEC 문서 파싱
	var docs []*SPECDoc
	var findings []Finding

	for _, path := range paths {
		doc := parseSPECDoc(path)
		docs = append(docs, doc)
		if doc.ParseError != nil {
			findings = append(findings, Finding{
				File:     path,
				Line:     1,
				Severity: SeverityError,
				Code:     "ParseFailure",
				Message:  fmt.Sprintf("SPEC 파싱 실패: %v", doc.ParseError),
			})
		}
	}

	// 단일 SPEC 규칙 실행
	for _, doc := range docs {
		if doc.ParseError != nil {
			continue // 파싱 실패 SPEC은 규칙 건너뜀
		}
		for _, rule := range l.rules {
			// cross-SPEC 규칙은 나중에 처리
			if _, ok := rule.(crossSPECRule); ok {
				continue
			}
			ruleFindings := rule.Check(doc, docs)
			ruleFindings = applylintSkip(ruleFindings, doc.LintSkip)
			findings = append(findings, ruleFindings...)
		}
	}

	// cross-SPEC 규칙 실행 (모든 문서를 한번에 검사)
	for _, rule := range l.rules {
		if cr, ok := rule.(crossSPECRule); ok {
			crossFindings := cr.CheckAll(docs)
			findings = append(findings, crossFindings...)
		}
	}

	return &Report{
		Findings: findings,
		Strict:   l.opts.Strict,
	}, nil
}

// crossSPECRule은 모든 SPEC 문서를 함께 검사하는 규칙 인터페이스이다.
type crossSPECRule interface {
	CheckAll(docs []*SPECDoc) []Finding
}

// Rule은 단일 lint 규칙 인터페이스이다.
//
// @MX:NOTE: [AUTO] Rule 인터페이스는 단일 SPEC 문서를 검사한다.
// crossSPECRule은 모든 SPEC을 함께 검사하는 별도 인터페이스이다.
type Rule interface {
	// Code는 규칙의 고유 코드를 반환한다.
	Code() string
	// Check는 단일 SPEC 문서를 검사하고 findings를 반환한다.
	// all은 다른 SPEC 문서들이며 규칙이 참조 가능하다.
	Check(doc *SPECDoc, all []*SPECDoc) []Finding
}

// applylintSkip은 doc의 lint.skip 코드 목록에 해당하는 findings를 제거한다.
func applylintSkip(findings []Finding, skipCodes []string) []Finding {
	if len(skipCodes) == 0 {
		return findings
	}
	skipSet := make(map[string]bool, len(skipCodes))
	for _, code := range skipCodes {
		skipSet[code] = true
	}
	var result []Finding
	for _, f := range findings {
		if !skipSet[f.Code] {
			result = append(result, f)
		}
	}
	return result
}

// discoverSPECs는 baseDir/.moai/specs/SPEC-*/spec.md 또는 baseDir/SPEC-*/spec.md 패턴으로
// SPEC 파일들을 탐색한다.
func discoverSPECs(baseDir string) ([]string, error) {
	if baseDir == "" {
		baseDir = "."
	}

	var paths []string

	// baseDir 바로 아래 SPEC-*/spec.md 패턴
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("디렉토리 읽기 실패 %q: %w", baseDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		candidate := filepath.Join(baseDir, entry.Name(), "spec.md")
		if _, err := os.Stat(candidate); err == nil {
			paths = append(paths, candidate)
		}
	}

	return paths, nil
}

// --- SPECDoc ---

// SPECFrontmatter는 SPEC 문서의 YAML 프론트매터를 나타낸다.
type SPECFrontmatter struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Version      string   `yaml:"version"`
	Status       string   `yaml:"status"`
	Created      string   `yaml:"created"`
	Updated      string   `yaml:"updated"`
	Author       string   `yaml:"author"`
	Priority     string   `yaml:"priority"`
	Phase        string   `yaml:"phase"`
	Module       string   `yaml:"module"`
	Dependencies []string `yaml:"dependencies"`
	BcID         []string `yaml:"bc_id"`
	Lifecycle    string   `yaml:"lifecycle"`
	Tags         string   `yaml:"tags"`
	Breaking     bool     `yaml:"breaking"`
	RelatedRule  []string `yaml:"related_rule"`
	// LintConfig는 lint.skip 코드 목록을 담는 중첩 구조이다.
	LintConfig struct {
		Skip []string `yaml:"skip"`
	} `yaml:"lint"`
}

// REQEntry는 파싱된 단일 요구사항이다.
type REQEntry struct {
	ID   string
	Text string
	Line int
}

// SPECDoc은 파싱된 SPEC 문서를 나타낸다.
type SPECDoc struct {
	Path        string
	Frontmatter SPECFrontmatter
	Body        string
	Criteria    []Acceptance
	REQs        []REQEntry
	ParseError  error
	LintSkip    []string
}

// reqIDPattern는 REQ-<DOMAIN>-<NNN>-<NNN> 형식을 검증하는 정규표현식이다.
var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z]{2,5}-\d{3}-\d{3}$`)

// reqLinePattern은 마크다운 REQ 라인을 파싱하는 정규표현식이다.
// 예: "- REQ-SPC-003-001: The system SHALL do X."
var reqLinePattern = regexp.MustCompile(`-\s+(REQ-[A-Z]{2,5}-\d{3}-\d{3})\s*:\s*(.+)`)

// parseSPECDoc은 주어진 경로의 SPEC 문서를 파싱한다.
// 파싱 실패 시 ParseError 필드에 오류를 설정한다.
func parseSPECDoc(path string) *SPECDoc {
	doc := &SPECDoc{Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		doc.ParseError = fmt.Errorf("파일 읽기 실패: %w", err)
		return doc
	}

	content := string(data)

	// YAML 프론트매터 추출 및 파싱
	fm, body, err := extractFrontmatter(content)
	if err != nil {
		doc.ParseError = fmt.Errorf("프론트매터 파싱 오류: %w", err)
		return doc
	}

	doc.Frontmatter = fm
	doc.Body = body
	doc.LintSkip = fm.LintConfig.Skip

	// REQ 목록 파싱
	doc.REQs = parseREQs(body)

	// Acceptance Criteria 파싱
	criteria, _ := ParseAcceptanceCriteria(body, false)
	doc.Criteria = criteria

	return doc
}

// extractFrontmatter는 마크다운 문서에서 YAML 프론트매터를 추출하고 파싱한다.
func extractFrontmatter(content string) (SPECFrontmatter, string, error) {
	var fm SPECFrontmatter

	if !strings.HasPrefix(content, "---") {
		return fm, content, fmt.Errorf("YAML 프론트매터가 없거나 '---'로 시작하지 않음")
	}

	// 두 번째 --- 찾기
	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return fm, content, fmt.Errorf("프론트매터 닫는 '---'를 찾을 수 없음")
	}

	yamlPart := rest[:endIdx]
	body := rest[endIdx+4:] // "\n---" 이후

	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return fm, body, fmt.Errorf("YAML 파싱 오류: %w", err)
	}

	return fm, body, nil
}

// parseREQs는 마크다운 본문에서 REQ 항목들을 파싱한다.
func parseREQs(body string) []REQEntry {
	var reqs []REQEntry
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		matches := reqLinePattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			reqs = append(reqs, REQEntry{
				ID:   matches[1],
				Text: strings.TrimSpace(matches[2]),
				Line: i + 1,
			})
		}
	}
	return reqs
}

// collectAllREQIDs는 Acceptance 트리의 모든 노드(리프+비리프)에서 REQ ID를 수집한다.
func collectAllREQIDs(criteria []Acceptance) map[string]bool {
	covered := make(map[string]bool)
	var visit func(ac *Acceptance)
	visit = func(ac *Acceptance) {
		for _, reqID := range ac.RequirementIDs {
			covered["REQ-"+reqID] = true
		}
		for i := range ac.Children {
			visit(&ac.Children[i])
		}
	}
	for i := range criteria {
		visit(&criteria[i])
	}
	return covered
}

// --- Rule 구현체들 ---

// EARSModalityRule은 REQ 텍스트의 EARS 모달리티 준수성을 검사한다.
// REQ-SPC-003-003, REQ-SPC-003-050 구현.
type EARSModalityRule struct{}

func (r *EARSModalityRule) Code() string { return "ModalityMalformed" }

func (r *EARSModalityRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	for _, req := range doc.REQs {
		if isModalityMalformed(req.Text) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "ModalityMalformed",
				Message:  fmt.Sprintf("REQ %s: EARS 모달리티 위반 — SHALL이 없거나 형식 불일치: %q", req.ID, req.Text),
			})
		}
	}
	return findings
}

// isModalityMalformed는 REQ 텍스트가 EARS 모달리티를 위반하는지 검사한다.
// EARS 형식: 모든 요구사항은 정확히 하나의 모달리티 키워드를 사용하고 SHALL을 포함해야 한다.
func isModalityMalformed(text string) bool {
	// EARS 모달리티 키워드로 시작하는지 확인
	upper := strings.ToUpper(text)

	// WHEN으로 시작하지만 SHALL이 없는 경우
	if strings.HasPrefix(upper, "WHEN ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	// WHILE로 시작하지만 SHALL이 없는 경우
	if strings.HasPrefix(upper, "WHILE ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	// WHERE로 시작하지만 SHALL이 없는 경우
	if strings.HasPrefix(upper, "WHERE ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	// IF로 시작하지만 SHALL이 없는 경우
	if strings.HasPrefix(upper, "IF ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	// Ubiquitous 형식: "The [system] SHALL"으로 시작해야 함
	// 대문자 THE로 시작하면서 SHALL이 없는 경우
	if strings.HasPrefix(upper, "THE ") && !strings.Contains(upper, " SHALL") {
		return true
	}
	return false
}

// REQIDUniquenessRule은 SPEC 내에서 REQ ID 유일성을 검사한다.
// REQ-SPC-003-004 구현.
type REQIDUniquenessRule struct{}

func (r *REQIDUniquenessRule) Code() string { return "DuplicateREQID" }

func (r *REQIDUniquenessRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	seen := make(map[string]int) // ID → 첫 번째 발견 라인

	for _, req := range doc.REQs {
		if !reqIDPattern.MatchString(req.ID) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "InvalidREQID",
				Message:  fmt.Sprintf("REQ ID %q가 패턴 REQ-[A-Z]{{2,5}}-NNN-NNN에 맞지 않음", req.ID),
			})
			continue
		}
		if firstLine, exists := seen[req.ID]; exists {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "DuplicateREQID",
				Message:  fmt.Sprintf("REQ ID %q가 중복됨 (첫 번째 등장: 라인 %d)", req.ID, firstLine),
			})
		} else {
			seen[req.ID] = req.Line
		}
	}
	return findings
}

// CoverageRule은 AC→REQ 커버리지를 검사한다.
// 모든 REQ는 최소 하나의 AC 리프 노드에서 참조되어야 한다.
// REQ-SPC-003-005 구현.
type CoverageRule struct{}

func (r *CoverageRule) Code() string { return "CoverageIncomplete" }

func (r *CoverageRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if len(doc.REQs) == 0 {
		return nil
	}

	// 모든 AC 노드(리프+비리프)에서 참조된 REQ ID 수집
	covered := collectAllREQIDs(doc.Criteria)

	var findings []Finding
	for _, req := range doc.REQs {
		if !covered[req.ID] {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     req.Line,
				Severity: SeverityError,
				Code:     "CoverageIncomplete",
				Message:  fmt.Sprintf("REQ %s가 어떤 AC에도 참조되지 않음", req.ID),
			})
		}
	}
	return findings
}

// FrontmatterSchemaRule은 SPEC 프론트매터 스키마를 검사한다.
// REQ-SPC-003-006 구현.
type FrontmatterSchemaRule struct{}

func (r *FrontmatterSchemaRule) Code() string { return "FrontmatterInvalid" }

// specIDPattern은 SPEC ID 형식을 검증하는 정규표현식이다.
var specIDPattern = regexp.MustCompile(`^SPEC-[A-Z][A-Z0-9]+-[A-Z]{2,5}-\d{3}$`)

// semverPattern은 시맨틱 버전 형식을 검증하는 정규표현식이다.
var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+`)

func (r *FrontmatterSchemaRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	required := []struct {
		name  string
		value string
	}{
		{"id", fm.ID},
		{"title", fm.Title},
		{"version", fm.Version},
		{"status", fm.Status},
		{"created", fm.Created},
		{"updated", fm.Updated},
		{"author", fm.Author},
		{"priority", fm.Priority},
		{"phase", fm.Phase},
		{"module", fm.Module},
		{"lifecycle", fm.Lifecycle},
		{"tags", fm.Tags},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "FrontmatterInvalid",
				Message:  fmt.Sprintf("프론트매터 필수 필드 누락: %s", field.name),
			})
		}
	}

	// id 형식 검증
	if fm.ID != "" && !specIDPattern.MatchString(fm.ID) {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "FrontmatterInvalid",
			Message:  fmt.Sprintf("id %q가 SPEC-<PREFIX>-<DOMAIN>-<NNN> 형식에 맞지 않음", fm.ID),
		})
	}

	// version 시맨틱 버전 검증
	if fm.Version != "" && !semverPattern.MatchString(fm.Version) {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "FrontmatterInvalid",
			Message:  fmt.Sprintf("version %q이 시맨틱 버전 형식(X.Y.Z)에 맞지 않음", fm.Version),
		})
	}

	// bc_id는 항상 배열이어야 함 (nil이 아닌 빈 배열 허용)
	// 타입 검사는 YAML 파싱 시 이미 처리됨

	return findings
}

// DependencyExistsRule은 dependencies 필드의 SPEC이 실제로 존재하는지 검사한다.
// REQ-SPC-003-007 구현.
type DependencyExistsRule struct{}

func (r *DependencyExistsRule) Code() string { return "MissingDependency" }

func (r *DependencyExistsRule) Check(doc *SPECDoc, all []*SPECDoc) []Finding {
	if len(doc.Frontmatter.Dependencies) == 0 {
		return nil
	}

	// 알려진 SPEC ID 집합 구성
	knownIDs := make(map[string]bool, len(all))
	for _, d := range all {
		if d.Frontmatter.ID != "" {
			knownIDs[d.Frontmatter.ID] = true
		}
	}

	var findings []Finding
	for _, dep := range doc.Frontmatter.Dependencies {
		// 알려진 SPECs에 없으면 디스크에서 확인
		if knownIDs[dep] {
			continue
		}

		// dep 디렉토리가 실제로 존재하는지 확인
		// doc.Path의 부모 디렉토리를 기준으로 탐색
		docDir := filepath.Dir(filepath.Dir(doc.Path))
		depDir := filepath.Join(docDir, dep)
		depSpec := filepath.Join(depDir, "spec.md")
		if _, err := os.Stat(depSpec); os.IsNotExist(err) {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "MissingDependency",
				Message:  fmt.Sprintf("의존 SPEC %q를 찾을 수 없음", dep),
			})
		}
	}
	return findings
}

// OutOfScopeRule은 "Out of Scope" 섹션의 존재를 검사한다.
// REQ-SPC-003-009 구현.
type OutOfScopeRule struct{}

func (r *OutOfScopeRule) Code() string { return "MissingExclusions" }

func (r *OutOfScopeRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	body := strings.ToLower(doc.Body)
	// "out of scope" 또는 "2.2 out of scope" 패턴
	hasOutOfScope := strings.Contains(body, "out of scope")
	if !hasOutOfScope {
		return []Finding{{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "MissingExclusions",
			Message:  "'Out of Scope' 섹션이 없음 — 최소 하나의 항목이 있는 Out of Scope 서브섹션 필수",
		}}
	}

	// Out of Scope 섹션이 있지만 내용이 없는 경우 확인
	lines := strings.Split(doc.Body, "\n")
	inOutOfScope := false
	hasContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lowerTrimmed := strings.ToLower(trimmed)

		if strings.HasPrefix(lowerTrimmed, "###") && strings.Contains(lowerTrimmed, "out of scope") {
			inOutOfScope = true
			continue
		}
		if strings.HasPrefix(lowerTrimmed, "##") && !strings.Contains(lowerTrimmed, "out of scope") && inOutOfScope {
			break // 다른 섹션 시작
		}
		if inOutOfScope && strings.HasPrefix(trimmed, "-") && len(strings.TrimPrefix(trimmed, "-")) > 0 {
			hasContent = true
			break
		}
	}

	if !hasContent {
		return []Finding{{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "MissingExclusions",
			Message:  "'Out of Scope' 섹션에 항목이 없음 — 최소 하나의 항목 필수",
		}}
	}

	return nil
}

// BreakingChangeIDRule은 breaking:true이면서 bc_id가 비어 있을 때 오류를 보고한다.
// REQ-SPC-003-052 구현.
type BreakingChangeIDRule struct{}

func (r *BreakingChangeIDRule) Code() string { return "BreakingChangeMissingID" }

func (r *BreakingChangeIDRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	var findings []Finding

	if fm.Breaking && len(fm.BcID) == 0 {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "BreakingChangeMissingID",
			Message:  "breaking: true이지만 bc_id가 비어 있음 — breaking change에는 bc_id가 필요함",
		})
	}

	if !fm.Breaking && len(fm.BcID) > 0 {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityWarning,
			Code:     "OrphanBCID",
			Message:  "breaking: false이지만 bc_id가 비어 있지 않음 — orphan breaking change ID",
		})
	}

	return findings
}

// ZoneRegistryRule은 related_rule 필드의 CONST-V3R2-NNN 참조가 zone registry에 존재하는지 검사한다.
// REQ-SPC-003-010 구현.
type ZoneRegistryRule struct {
	registry *constitution.Registry
}

func (r *ZoneRegistryRule) Code() string { return "DanglingRuleReference" }

func (r *ZoneRegistryRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	if r.registry == nil || len(doc.Frontmatter.RelatedRule) == 0 {
		return nil
	}

	var findings []Finding
	for _, ruleID := range doc.Frontmatter.RelatedRule {
		if _, ok := r.registry.Get(ruleID); !ok {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityWarning,
				Code:     "DanglingRuleReference",
				Message:  fmt.Sprintf("related_rule %q가 zone registry에 없음", ruleID),
			})
		}
	}
	return findings
}

// DependencyCycleRule은 SPEC 의존성 DAG에서 사이클을 탐지한다.
// REQ-SPC-003-008 구현.
// crossSPECRule 인터페이스를 구현하여 Linter.Lint에서 cross-SPEC 단계에 실행된다.
type DependencyCycleRule struct{}

func (r *DependencyCycleRule) Code() string { return "DependencyCycle" }

func (r *DependencyCycleRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding {
	// single-spec check는 사용하지 않음; CheckAll에서 처리
	return nil
}

func (r *DependencyCycleRule) CheckAll(docs []*SPECDoc) []Finding {
	// ID → 인덱스 맵
	idToIdx := make(map[string]int, len(docs))
	for i, doc := range docs {
		if doc.Frontmatter.ID != "" {
			idToIdx[doc.Frontmatter.ID] = i
		}
	}

	// 인접 리스트 구성 (인덱스 기반)
	adj := make([][]int, len(docs))
	for i, doc := range docs {
		for _, dep := range doc.Frontmatter.Dependencies {
			if j, ok := idToIdx[dep]; ok {
				adj[i] = append(adj[i], j)
			}
		}
	}

	// Tarjan SCC로 사이클 탐지
	cycles := findCyclesTarjan(adj, len(docs))

	if len(cycles) == 0 {
		return nil
	}

	var findings []Finding
	for _, cycle := range cycles {
		names := make([]string, 0, len(cycle))
		for _, idx := range cycle {
			names = append(names, docs[idx].Frontmatter.ID)
			if names[len(names)-1] == "" {
				names[len(names)-1] = docs[idx].Path
			}
		}
		findings = append(findings, Finding{
			File:     docs[cycle[0]].Path,
			Line:     1,
			Severity: SeverityError,
			Code:     "DependencyCycle",
			Message:  fmt.Sprintf("의존성 사이클 탐지: %s", strings.Join(names, " → ")),
		})
	}
	return findings
}

// DuplicateSPECIDRule은 여러 SPEC이 동일한 id를 선언하는지 검사한다.
// REQ-SPC-003-031 구현.
// crossSPECRule 인터페이스를 구현한다.
type DuplicateSPECIDRule struct{}

func (r *DuplicateSPECIDRule) Code() string { return "DuplicateSPECID" }

func (r *DuplicateSPECIDRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding {
	return nil
}

func (r *DuplicateSPECIDRule) CheckAll(docs []*SPECDoc) []Finding {
	seen := make(map[string]string) // ID → 첫 번째 파일 경로
	var findings []Finding

	for _, doc := range docs {
		id := doc.Frontmatter.ID
		if id == "" {
			continue
		}
		if firstPath, exists := seen[id]; exists {
			findings = append(findings, Finding{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityError,
				Code:     "DuplicateSPECID",
				Message:  fmt.Sprintf("SPEC ID %q 중복 선언 (첫 번째 위치: %s)", id, firstPath),
			})
		} else {
			seen[id] = doc.Path
		}
	}
	return findings
}
