package template_test

// TestNoAskUserQuestionInSubagents는 .claude/agents/**/*.md 파일에서
// 본문에 AskUserQuestion( 리터럴이 포함되지 않음을 검증합니다.
//
// SPEC-V3R2-RT-004 AC-11: 서브에이전트에서 AskUserQuestion 직접 호출 금지.
// 서브에이전트는 BlockerReport를 통해 오케스트레이터에 보고해야 합니다.
// 위반 시 sentinel: ASKUSERQUESTION_IN_SUBAGENT
//
// 근거: .claude/rules/moai/core/agent-common-protocol.md §User Interaction Boundary.
// 서브에이전트는 고립된 상태 비저장 컨텍스트에서 실행되므로 AskUserQuestion이
// 작동하지 않습니다. 오케스트레이터만 AskUserQuestion을 사용해야 합니다.

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoAskUserQuestionInSubagents walks .claude/agents/**/*.md (NOT skills)
// and asserts body contains no literal AskUserQuestion(.
func TestNoAskUserQuestionInSubagents(t *testing.T) {
	t.Parallel()

	// 프로젝트 루트 탐색 (go.mod 기준)
	projectRoot := findProjectRoot(t)
	agentsDir := filepath.Join(projectRoot, ".claude", "agents")

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Skipf(".claude/agents/ not found at %s; skipping audit", agentsDir)
	}

	// .claude/agents/**/*.md 순회
	err := filepath.WalkDir(agentsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// 파일 읽기 + body 검사 (frontmatter 이후 시작)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		body := extractBody(data)
		scanner := bufio.NewScanner(bytes.NewReader(body))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.Contains(line, "AskUserQuestion(") {
				// sentinel 출력 + 테스트 실패
				rel, _ := filepath.Rel(projectRoot, path)
				t.Errorf(
					"ASKUSERQUESTION_IN_SUBAGENT: %s body line %d contains AskUserQuestion(; "+
						"subagents must use BlockerReport, not AskUserQuestion. "+
						"See agent-common-protocol.md §User Interaction Boundary.",
					rel, lineNum,
				)
			}
		}
		return scanner.Err()
	})

	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}
}

// extractBody는 YAML frontmatter(---...--- 블록) 이후의 본문을 반환합니다.
// frontmatter가 없으면 전체 내용을 반환합니다.
func extractBody(data []byte) []byte {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return data
	}

	// 두 번째 --- 찾기
	rest := content[3:] // 첫 번째 --- 이후
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return data
	}

	// 두 번째 --- 이후 콘텐츠
	body := rest[idx+3:]
	return []byte(body)
}

// findProjectRoot는 go.mod가 있는 디렉토리를 탐색합니다.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	// 현재 워킹 디렉토리부터 상위로 탐색
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}
