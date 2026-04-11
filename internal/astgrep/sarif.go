package astgrep

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SARIF 2.1.0 스키마 URL
const sarifSchema = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

// sarifDocument는 SARIF 2.1.0 최상위 문서입니다.
// https://docs.oasis-open.org/sarif/sarif/v2.1.0/
type sarifDocument struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun은 단일 도구 실행 결과입니다.
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool은 스캔 도구 정보입니다.
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver는 도구 드라이버 메타데이터입니다.
type sarifDriver struct {
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Rules   []sarifRule  `json:"rules,omitempty"`
}

// sarifRule은 SARIF 규칙 정의입니다.
type sarifRule struct {
	ID               string            `json:"id"`
	ShortDescription *sarifMessage     `json:"shortDescription,omitempty"`
	Properties       map[string]string `json:"properties,omitempty"`
}

// sarifResult는 단일 발견 결과입니다.
type sarifResult struct {
	RuleID    string             `json:"ruleId"`
	Level     string             `json:"level"`
	Message   sarifMessage       `json:"message"`
	Locations []sarifLocation    `json:"locations"`
	Properties map[string]string `json:"properties,omitempty"`
}

// sarifMessage는 SARIF 메시지 텍스트입니다.
type sarifMessage struct {
	Text string `json:"text"`
}

// sarifLocation은 코드 위치 정보입니다.
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

// sarifPhysicalLocation은 파일 및 영역 정보입니다.
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

// sarifArtifactLocation은 파일 URI입니다.
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifRegion은 코드 영역(줄/컬럼)입니다.
type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// ToSARIF는 Finding 슬라이스를 SARIF 2.1.0 형식의 JSON으로 변환합니다.
// REQ-ASTG-UPG-022: SARIF 2.1.0 출력 지원
//
// sgVersion은 ast-grep CLI 버전 문자열입니다 (tool.driver.version에 반영).
// findings가 nil이면 빈 results 배열을 포함한 유효한 SARIF 문서를 반환합니다.
func ToSARIF(findings []Finding, sgVersion string) ([]byte, error) {
	if sgVersion == "" {
		sgVersion = "unknown"
	}

	// 규칙 목록 수집 (dedup)
	ruleSet := make(map[string]sarifRule)
	for _, f := range findings {
		if f.RuleID == "" {
			continue
		}
		if _, exists := ruleSet[f.RuleID]; !exists {
			rule := sarifRule{
				ID: f.RuleID,
				ShortDescription: &sarifMessage{Text: f.Message},
			}
			if len(f.Metadata) > 0 {
				rule.Properties = f.Metadata
			}
			ruleSet[f.RuleID] = rule
		}
	}

	rules := make([]sarifRule, 0, len(ruleSet))
	for _, r := range ruleSet {
		rules = append(rules, r)
	}

	// results 변환
	results := make([]sarifResult, 0, len(findings))
	for _, f := range findings {
		result := sarifResult{
			RuleID:  f.RuleID,
			Level:   severityToSARIFLevel(f.Severity),
			Message: sarifMessage{Text: f.Message},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: toFileURI(f.File),
						},
						Region: sarifRegion{
							StartLine:   maxInt(f.Line, 1),
							StartColumn: f.Column,
							EndLine:     f.EndLine,
							EndColumn:   f.EndColumn,
						},
					},
				},
			},
		}

		// 메타데이터를 SARIF properties로 전달 (CWE/OWASP 보존)
		if len(f.Metadata) > 0 {
			result.Properties = f.Metadata
		}

		results = append(results, result)
	}

	doc := sarifDocument{
		Schema:  sarifSchema,
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "moai-ast-grep",
						Version: sgVersion,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	output, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("SARIF 직렬화: %w", err)
	}

	return output, nil
}

// severityToSARIFLevel은 ast-grep severity를 SARIF level로 변환합니다.
// SARIF 2.1.0 지원 level: "error", "warning", "note", "none"
// REQ-ASTG-UPG-022: severity 매핑 (error→error, warning→warning, info→note)
func severityToSARIFLevel(severity string) string {
	switch strings.ToLower(severity) {
	case "error":
		return "error"
	case "warning", "warn":
		return "warning"
	default:
		return "note" // info, hint, "" 모두 note로 매핑
	}
}

// toFileURI는 파일 경로를 SARIF URI 형식으로 변환합니다.
// 절대 경로는 file:// 스킴을 붙이지 않고 상대 경로를 그대로 사용합니다.
func toFileURI(path string) string {
	if path == "" {
		return ""
	}
	// 이미 URI 형식이면 그대로 반환
	if strings.HasPrefix(path, "file://") {
		return path
	}
	return path
}

// maxInt는 두 정수 중 큰 값을 반환합니다.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
