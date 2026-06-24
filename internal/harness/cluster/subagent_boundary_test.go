package cluster

import (
	"os"
	"strings"
	"testing"
)

// TestCluster_NoAskUserQuestion 은 정적 검사다 — cluster 패키지 소스는 AskUserQuestion /
// mcp__askuser 를 참조하면 안 된다(C-HRA-008 / REQ-OBL-015). 클러스터러는 subagent
// 컨텍스트에서 동작하며 사용자 상호작용은 orchestrator 가 소유한다. 이 가드는
// internal/cli/worktree/new_test.go 의 TestNew_NoAskUserQuestion 패턴을 미러링한다.
//
// AC-OBL-006: grep 게이트의 Go-test 대응물.
func TestCluster_NoAskUserQuestion(t *testing.T) {
	// 패키지의 모든 비-테스트 소스 파일을 스캔한다.
	files := []string{
		"cluster.go",
		"report.go",
	}
	forbidden := []string{"AskUserQuestion", "mcp__askuser"}

	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(src)
		for _, token := range forbidden {
			if strings.Contains(content, token) {
				t.Errorf("internal/harness/cluster/%s 는 %q 를 참조하면 안 됨 (subagent boundary, orchestrator-only HARD)", file, token)
			}
		}
	}
}
