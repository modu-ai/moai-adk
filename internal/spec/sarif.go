package spec

import "encoding/json"

// sarif.go는 SARIF 2.1.0 형식의 출력을 생성한다.
// SPEC-V3R2-SPC-003 REQ-SPC-003-021 구현.
//
// SARIF 2.1.0 스키마: https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json

const (
	sarifVersion = "2.1.0"
	sarifSchema  = "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json"
	toolName     = "moai-spec-lint"
	toolVersion  = "0.1.0"
)

// sarifLog는 SARIF 2.1.0 최상위 구조체이다.
type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun은 단일 lint 실행을 나타낸다.
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool은 lint 도구 정보이다.
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver는 lint 도구 드라이버 정보이다.
type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules"`
}

// sarifRule은 lint 규칙 정보이다.
type sarifRule struct {
	ID               string           `json:"id"`
	ShortDescription sarifMessage     `json:"shortDescription"`
	DefaultConfig    sarifRuleDefault `json:"defaultConfiguration"`
}

// sarifRuleDefault는 규칙의 기본 설정이다.
type sarifRuleDefault struct {
	Level string `json:"level"`
}

// sarifResult는 단일 finding을 나타낸다.
type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

// sarifMessage는 SARIF 메시지 구조체이다.
type sarifMessage struct {
	Text string `json:"text"`
}

// sarifLocation은 finding의 위치 정보이다.
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

// sarifPhysicalLocation은 파일 + 라인 위치이다.
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

// sarifArtifactLocation은 파일 URI이다.
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifRegion은 파일 내 위치 범위이다.
type sarifRegion struct {
	StartLine int `json:"startLine"`
}

// severityToSARIFLevel은 Severity를 SARIF level 문자열로 변환한다.
func severityToSARIFLevel(sev Severity) string {
	switch sev {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}

// marshalSARIF는 findings를 SARIF 2.1.0 형식의 JSON 바이트로 변환한다.
func marshalSARIF(findings []Finding) ([]byte, error) {
	// 규칙 목록 수집 (중복 제거)
	rulesSeen := make(map[string]bool)
	var rules []sarifRule
	for _, f := range findings {
		if !rulesSeen[f.Code] {
			rulesSeen[f.Code] = true
			rules = append(rules, sarifRule{
				ID: f.Code,
				ShortDescription: sarifMessage{
					Text: f.Code,
				},
				DefaultConfig: sarifRuleDefault{
					Level: severityToSARIFLevel(f.Severity),
				},
			})
		}
	}

	// results 구성
	results := make([]sarifResult, 0, len(findings))
	for _, f := range findings {
		result := sarifResult{
			RuleID: f.Code,
			Level:  severityToSARIFLevel(f.Severity),
			Message: sarifMessage{
				Text: f.Message,
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: f.File,
						},
						Region: sarifRegion{
							StartLine: positiveLineNum(f.Line),
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	log := sarifLog{
		Schema:  sarifSchema,
		Version: sarifVersion,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    toolName,
						Version: toolVersion,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	return json.MarshalIndent(log, "", "  ")
}

// positiveLineNum은 라인 번호가 최소 1이 되도록 보장한다.
func positiveLineNum(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
