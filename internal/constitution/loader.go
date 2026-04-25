package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// designMirrorStart는 design subsystem mirror 엔트리 ID 범위 시작 (051).
const designMirrorStart = 51

// designMirrorEnd는 design subsystem mirror 엔트리 ID 범위 끝 (099, 포함).
const designMirrorEnd = 99

// designOverflowEnd는 design mirror overflow 범위 끝 (149, 포함).
const designOverflowEnd = 149

// Registry는 로드된 zone registry를 나타낸다.
// Entries는 전체 엔트리 목록이며, lookup은 ID→인덱스 O(1) 조회를 위한 내부 맵이다.
type Registry struct {
	// Entries는 로드된 Rule 목록이다.
	Entries []Rule
	// Warnings는 로드 중 발생한 경고 메시지 목록이다 (orphan, overflow 등).
	Warnings []string
	// lookup은 ID→인덱스 내부 맵 (O(1) Get 지원).
	lookup map[string]int
}

// Get은 ID로 Rule을 O(1)에 조회한다.
func (r *Registry) Get(id string) (Rule, bool) {
	if r.lookup == nil {
		return Rule{}, false
	}
	idx, ok := r.lookup[id]
	if !ok {
		return Rule{}, false
	}
	return r.Entries[idx], true
}

// FilterByZone은 지정된 zone의 Rule 목록을 반환한다.
// REQ-CON-001-012 지원.
func (r *Registry) FilterByZone(z Zone) []Rule {
	var result []Rule
	for _, entry := range r.Entries {
		if entry.Zone == z {
			result = append(result, entry)
		}
	}
	return result
}

// rawEntry는 YAML unmarshal용 내부 구조체이다.
// Zone 필드를 문자열로 먼저 파싱하기 위해 별도 정의.
type rawEntry struct {
	ID         string `yaml:"id"`
	Zone       string `yaml:"zone"`
	File       string `yaml:"file"`
	Anchor     string `yaml:"anchor"`
	Clause     string `yaml:"clause"`
	CanaryGate bool   `yaml:"canary_gate"`
}

// LoadRegistry는 지정된 경로의 zone registry 마크다운 파일을 로드한다.
//
// 파싱 전략 (plan.md §7 OQ1 Decision):
//  1. 파일에서 첫 번째 ```yaml ... ``` code fence 추출
//  2. gopkg.in/yaml.v3로 []rawEntry unmarshal
//  3. Rule 검증: 중복 ID는 fatal error, orphan 파일은 warning + Orphan=true
//
// projectDir은 filepath.Clean + scope 제한을 위한 기준 디렉토리.
// 경로 traversal 방지를 위해 registry 파일 경로를 projectDir 내부로 제한한다.
func LoadRegistry(path, projectDir string) (*Registry, error) {
	// 경로 traversal 방지: path를 clean하고 projectDir 범위 내인지 확인
	cleanPath := filepath.Clean(path)
	cleanProjectDir := filepath.Clean(projectDir)
	if filepath.IsAbs(cleanPath) {
		// 절대 경로인 경우 projectDir 기준으로 상대 경로로 변환하여 scope 확인
		rel, err := filepath.Rel(cleanProjectDir, cleanPath)
		if err == nil && strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("registry 경로 %q가 project dir %q를 벗어난다", path, projectDir)
		}
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("registry 파일 읽기 오류 %q: %w", path, err)
	}

	yamlBlock, err := extractYAMLFence(string(data))
	if err != nil {
		return nil, fmt.Errorf("YAML code fence 추출 오류: %w", err)
	}

	var raw []rawEntry
	if err := yaml.Unmarshal([]byte(yamlBlock), &raw); err != nil {
		return nil, fmt.Errorf("YAML 파싱 오류: %w", err)
	}

	reg := &Registry{
		Entries:  make([]Rule, 0, len(raw)),
		Warnings: nil,
		lookup:   make(map[string]int, len(raw)),
	}

	var mirrorCount int

	for i, re := range raw {
		z, parseErr := ParseZone(re.Zone)
		if parseErr != nil {
			return nil, fmt.Errorf("엔트리 %d (id=%q) zone 파싱 오류: %w", i, re.ID, parseErr)
		}

		rule := Rule{
			ID:         re.ID,
			Zone:       z,
			File:       re.File,
			Anchor:     re.Anchor,
			Clause:     re.Clause,
			CanaryGate: re.CanaryGate,
		}

		// ID 유효성 검증
		if validErr := rule.Validate(); validErr != nil {
			return nil, fmt.Errorf("엔트리 %d 유효성 오류: %w", i, validErr)
		}

		// 중복 ID 탐지 (fatal)
		if _, exists := reg.lookup[rule.ID]; exists {
			return nil, fmt.Errorf("중복 ID 탐지: %q", rule.ID)
		}

		// orphan 파일 확인: file 경로가 실제로 존재하는지 확인
		// projectDir 기준으로 상대 경로 해석
		rulePath := rule.File
		if !filepath.IsAbs(rulePath) {
			rulePath = filepath.Join(cleanProjectDir, rulePath)
		}
		if _, statErr := os.Stat(filepath.Clean(rulePath)); statErr != nil {
			rule = rule.withOrphan(true)
			reg.Warnings = append(reg.Warnings,
				fmt.Sprintf("orphan: 파일 %q를 찾을 수 없다 (rule %s)", rule.File, rule.ID))
		}

		// design mirror overflow 탐지 (051-099 범위 초과)
		idNum := extractIDNumber(rule.ID)
		if idNum >= designMirrorStart && idNum <= designMirrorEnd {
			mirrorCount++
		}
		if idNum > designMirrorEnd && idNum <= designOverflowEnd {
			// overflow range 사용 중
			reg.Warnings = append(reg.Warnings,
				fmt.Sprintf("design mirror overflow: ID %s가 확장 범위(100-149)를 사용하고 있다", rule.ID))
		}

		reg.lookup[rule.ID] = len(reg.Entries)
		reg.Entries = append(reg.Entries, rule)
	}

	// design mirror 49 slot 초과 감지
	if mirrorCount > designMirrorEnd-designMirrorStart+1 {
		reg.Warnings = append(reg.Warnings,
			fmt.Sprintf("design mirror overflow: mirror 엔트리 수 %d가 허용 슬롯 49를 초과한다", mirrorCount))
	}

	// 경고는 reg.Warnings에 저장되어 반환된다.
	// orphan, overflow 경고는 caller가 reg.Warnings를 통해 확인한다.
	// REQ-CON-001-040: orphan은 error를 반환하지 않고 Orphan=true 플래그만 설정.
	return reg, nil
}

// extractYAMLFence는 마크다운 문서에서 첫 번째 ```yaml ... ``` 블록을 추출한다.
func extractYAMLFence(content string) (string, error) {
	const openFence = "```yaml"
	const closeFence = "```"

	start := strings.Index(content, openFence)
	if start == -1 {
		return "", fmt.Errorf("```yaml code fence를 찾을 수 없다")
	}

	// openFence 다음 위치
	afterOpen := start + len(openFence)
	// 닫는 fence 찾기 (openFence 이후에서)
	end := strings.Index(content[afterOpen:], closeFence)
	if end == -1 {
		return "", fmt.Errorf("닫는 ``` fence를 찾을 수 없다")
	}

	return content[afterOpen : afterOpen+end], nil
}

// extractIDNumber는 CONST-V3R2-NNN 형식에서 숫자 부분을 추출한다.
// 파싱 실패 시 -1 반환.
func extractIDNumber(id string) int {
	const prefix = "CONST-V3R2-"
	if !strings.HasPrefix(id, prefix) {
		return -1
	}
	numStr := id[len(prefix):]
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return -1
	}
	return n
}
