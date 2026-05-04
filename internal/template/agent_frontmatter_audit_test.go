// agent_frontmatter_audit_test.go: 은퇴 에이전트 frontmatter 표준화 검증.
// REQ-RA-002, REQ-RA-013, REQ-RA-016 매핑.
//
// M1 RED phase: 이 파일의 테스트들은 아직 구현되지 않은 상태를 가정하여
// 의도적으로 실패하도록 설계되었다.
//
// 예상 RED 상태:
//   - TestAgentFrontmatterAudit: manager-tdd.md에 retired:true가 없으므로 FAIL (M2에서 GREEN)
//   - TestRetirementCompletenessAssertion: manager-cycle.md가 embedded FS에 없으므로 FAIL (M2에서 GREEN)
//   - TestNoOrphanedManagerTDDReference: 여러 파일에 manager-tdd 참조가 남아 있으므로 FAIL (M5에서 GREEN)
package template

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// retiredFrontmatter는 파싱된 에이전트 파일의 retired 관련 필드를 담는다.
type retiredFrontmatter struct {
	retired            bool
	retiredReplacement string
	retiredParamHint   string
	tools              string // 원시 값 (빈 배열은 "[]" 로 파싱됨)
	skills             string // 원시 값 (빈 배열은 "[]" 로 파싱됨)
	hasStatusRetired   bool   // legacy status: retired 필드 존재 여부
}

// parseRetiredFields는 frontmatter map에서 retired 관련 필드를 추출한다.
func parseRetiredFields(fm map[string]string) retiredFrontmatter {
	result := retiredFrontmatter{}

	if val, ok := fm["retired"]; ok {
		// YAML boolean: "true" 문자열로 파싱됨 (parseFrontmatterAndBody는 따옴표 제거 후 raw 반환)
		result.retired = strings.TrimSpace(val) == "true"
	}
	if val, ok := fm["retired_replacement"]; ok {
		result.retiredReplacement = strings.TrimSpace(val)
	}
	if val, ok := fm["retired_param_hint"]; ok {
		result.retiredParamHint = strings.TrimSpace(val)
	}
	if val, ok := fm["tools"]; ok {
		result.tools = strings.TrimSpace(val)
	}
	if val, ok := fm["skills"]; ok {
		result.skills = strings.TrimSpace(val)
	}
	// legacy status: retired 필드 감지
	if val, ok := fm["status"]; ok && strings.TrimSpace(val) == "retired" {
		result.hasStatusRetired = true
	}

	return result
}

// TestAgentFrontmatterAudit는 .claude/agents/moai/*.md 파일을 순회하며
// retired:true frontmatter의 5개 표준 필드 존재를 검증한다.
//
// REQ-RA-002: 표준 retired frontmatter 필드 검증
// 예상 RED: manager-tdd.md에 아직 retired:true가 없으므로 FAIL
func TestAgentFrontmatterAudit(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	var agentFiles []string
	walkErr := fs.WalkDir(fsys, ".claude/agents/moai", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			agentFiles = append(agentFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir 오류: %v", walkErr)
	}
	if len(agentFiles) == 0 {
		t.Fatal(".claude/agents/moai/ 하위 에이전트 파일이 없음")
	}

	// 검증 규칙:
	// 1. retired:true 에이전트: 5개 표준 필드 모두 필수
	// 2. 비-retired 에이전트: retired: 키 자체 없어야 함
	// 3. legacy status:retired 필드 금지
	for _, path := range agentFiles {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) 오류: %v", path, readErr)
			}

			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				t.Fatalf("frontmatter 파싱 실패: %s", parseErr)
			}

			rf := parseRetiredFields(fm)

			// legacy status:retired 필드 금지 (REQ-RA-002: 'retired: true' boolean 사용)
			if rf.hasStatusRetired {
				t.Errorf("RETIREMENT_INCOMPLETE: legacy 'status: retired' 필드 감지. 'retired: true' boolean 필드로 교체 필요 (REQ-RA-002)")
			}

			if rf.retired {
				// retired:true 에이전트: 5개 표준 필드 검증
				if rf.retiredReplacement == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'retired_replacement' 필드 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
				if rf.retiredParamHint == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'retired_param_hint' 필드 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
				// tools: [] 빈 배열 명시적 필요
				if rf.tools == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'tools: []' 명시적 빈 배열 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
				// skills: [] 빈 배열 명시적 필요
				if rf.skills == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'skills: []' 명시적 빈 배열 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
			}
		})
	}

	// RED 트리거: manager-tdd.md가 retired:true를 가져야 한다는 명시적 단언
	// M2에서 manager-tdd.md가 표준화되면 GREEN이 됨
	t.Run("manager-tdd must be retired", func(t *testing.T) {
		t.Parallel()

		const tddPath = ".claude/agents/moai/manager-tdd.md"
		data, readErr := fs.ReadFile(fsys, tddPath)
		if readErr != nil {
			t.Fatalf("manager-tdd.md 읽기 실패 (파일 존재해야 함): %v", readErr)
		}

		fm, _, parseErr := parseFrontmatterAndBody(string(data))
		if parseErr != "" {
			t.Fatalf("manager-tdd.md frontmatter 파싱 실패: %s", parseErr)
		}

		rf := parseRetiredFields(fm)

		// SPEC-V3R3-RETIRED-AGENT-001 REQ-RA-002: manager-tdd는 retired:true여야 함
		// 현재는 full definition이므로 이 테스트가 FAIL (예상 RED)
		if !rf.retired {
			t.Errorf("RETIREMENT_INCOMPLETE_manager-tdd: manager-tdd.md에 'retired: true' 없음. " +
				"SPEC-V3R3-RETIRED-AGENT-001 M2에서 retired stub으로 교체 필요 (REQ-RA-002)")
		}
	})
}

// TestRetirementCompletenessAssertion은 retired:true 에이전트 각각에 대해
// 교체 에이전트 파일이 embedded FS에 존재하는지 검증한다.
//
// REQ-RA-016: CI에서 RETIREMENT_INCOMPLETE_<agent> 검사
// 예상 RED: manager-cycle.md가 embedded FS에 없으므로 FAIL
func TestRetirementCompletenessAssertion(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	// manager-tdd → manager-cycle 대체 관계 명시적 단언 (M2 이전 RED 트리거)
	// manager-tdd.md가 retired:true를 가질 때, manager-cycle.md가 embedded FS에 있어야 함
	t.Run("manager-tdd replacement manager-cycle must exist", func(t *testing.T) {
		t.Parallel()

		const replacementPath = ".claude/agents/moai/manager-cycle.md"
		_, statErr := fs.Stat(fsys, replacementPath)
		if statErr != nil {
			t.Errorf("RETIREMENT_INCOMPLETE_manager-tdd: 교체 에이전트 '%s'가 embedded FS에 없음. "+
				"SPEC-V3R3-RETIRED-AGENT-001 M2에서 manager-cycle.md 추가 필요 (REQ-RA-016)", replacementPath)
		}
	})

	// 범용 검증: embedded FS의 모든 retired:true 에이전트에 대해 교체 파일 존재 확인
	t.Run("all retired agents have replacement in embedded FS", func(t *testing.T) {
		t.Parallel()

		var agentFiles []string
		_ = fs.WalkDir(fsys, ".claude/agents/moai", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".md") {
				agentFiles = append(agentFiles, path)
			}
			return nil
		})

		for _, path := range agentFiles {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				continue
			}
			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				continue
			}
			rf := parseRetiredFields(fm)
			if !rf.retired || rf.retiredReplacement == "" {
				continue
			}

			// 교체 파일 경로 탐색: .claude/agents/moai/<replacement>.md
			replacementPath := fmt.Sprintf(".claude/agents/moai/%s.md", rf.retiredReplacement)
			_, statErr := fs.Stat(fsys, replacementPath)
			if statErr != nil {
				t.Errorf("RETIREMENT_INCOMPLETE_%s: retired_replacement '%s' 파일이 embedded FS에 없음 (%s)",
					agentNameFromPath(path), rf.retiredReplacement, replacementPath)
			}
		}
	})
}

// TestNoOrphanedManagerTDDReference는 특정 핵심 파일들에서 manager-tdd 참조가
// 남아 있지 않은지 검증한다.
//
// REQ-RA-013: manager-cycle이 활성 통합 에이전트일 때 모든 문서 참조가 갱신되어야 함
// 예상 RED: 여러 파일에 manager-tdd 참조가 아직 남아 있으므로 FAIL
func TestNoOrphanedManagerTDDReference(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	// manager-tdd 참조가 없어야 하는 파일들 (REQ-RA-013 §M5 substitution scope)
	// 각 파일에서 "manager-tdd" 문자열이 발견되면 FAIL
	// 예외: manager-tdd.md 파일 자체 (frontmatter name: 라인, 마이그레이션 노트)
	checkFiles := []struct {
		path        string
		description string
	}{
		{
			path:        "CLAUDE.md",
			description: "CLAUDE.md §4 Manager Agents 및 §5 Agent Chain",
		},
		{
			path:        ".claude/rules/moai/development/agent-authoring.md",
			description: "agent-authoring.md Manager Agents 목록",
		},
		{
			path:        ".claude/rules/moai/core/agent-hooks.md",
			description: "agent-hooks.md Agent Hook Actions 테이블",
		},
		{
			path:        ".claude/rules/moai/workflow/spec-workflow.md",
			description: "spec-workflow.md Phase Overview 테이블",
		},
		{
			path:        ".claude/agents/moai/manager-strategy.md",
			description: "manager-strategy.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/manager-ddd.md",
			description: "manager-ddd.md 인라인 참조 (2개 라인)",
		},
	}

	for _, cf := range checkFiles {
		cf := cf
		t.Run(cf.path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, cf.path)
			if readErr != nil {
				// 파일이 없으면 테스트 스킵 (M5에서 파일 존재 확인)
				t.Skipf("파일 %q 읽기 실패 (make build 필요): %v", cf.path, readErr)
				return
			}

			content := string(data)
			// manager-tdd 참조 검색 (대소문자 구분)
			// 단순 포함 검사: 정확한 단어 경계를 위해 공통 패턴 검색
			orphanedRefs := findManagerTDDReferences(content)
			if len(orphanedRefs) > 0 {
				t.Errorf("ORPHANED_MANAGER_TDD_REFERENCE in %s (%s): %d개 참조 발견. "+
					"SPEC-V3R3-RETIRED-AGENT-001 M5에서 'manager-cycle'로 교체 필요 (REQ-RA-013):\n%s",
					cf.path, cf.description, len(orphanedRefs), strings.Join(orphanedRefs, "\n"))
			}
		})
	}
}

// agentNameFromPath는 파일 경로에서 에이전트 이름을 추출한다.
// ".claude/agents/moai/manager-tdd.md" → "manager-tdd"
func agentNameFromPath(path string) string {
	base := path
	// 마지막 / 이후 부분
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		base = path[idx+1:]
	}
	// .md 제거
	return strings.TrimSuffix(base, ".md")
}

// findManagerTDDReferences는 콘텐츠에서 manager-tdd 관련 참조를 찾아 반환한다.
// frontmatter의 name: 라인 및 마이그레이션 노트는 허용한다.
func findManagerTDDReferences(content string) []string {
	var refs []string
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// manager-tdd 참조 검색
		if !strings.Contains(line, "manager-tdd") {
			continue
		}
		// 허용 예외: frontmatter name 필드 (manager-tdd.md 자체의 name: manager-tdd)
		// 허용 예외: 마이그레이션 노트 (# deprecated, # 이전 이름, <!-- , [DEPRECATED 등)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name: manager-tdd") {
			continue
		}
		if strings.HasPrefix(trimmed, "#") && strings.Contains(strings.ToLower(trimmed), "deprecated") {
			continue
		}
		if strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		refs = append(refs, fmt.Sprintf("  L%d: %s", i+1, trimmed))
	}
	return refs
}
