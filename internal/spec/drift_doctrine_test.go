package spec

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// findRepoRoot는 cwd에서 위로 올라가며 go.mod를 포함한 디렉터리(repo root)를 찾는다.
// Go 테스트는 package 디렉터리(internal/spec)를 cwd로 실행하므로 doctrine 파일
// (repo root 기준 .claude/rules/...)에 도달하려면 root를 찾아야 한다.
func findRepoRoot(t *testing.T) string {
	t.Helper()
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
			t.Fatalf("repo root (go.mod) not found from cwd")
		}
		dir = parent
	}
}

// TestCloseSubjectDoctrineAmendment는 AC-DLC-011 검증 (REQ-DLC-011, D-NEW-2 portable oracle).
// 두 doctrine SSOT 파일이 close-subject full-ID mandate amendment를 포함하는지 검증한다 —
// combined/abbreviated scope 문구와 prohibition verb가 동일 amendment block 내 ≤400자 이내에
// co-located되어야 한다 (incidental "full SPEC-ID" 언급은 불충분).
//
// grep -Pzo (GNU 전용) 대신 portable Go regexp로 구현 (D-NEW-2 MINOR).
func TestCloseSubjectDoctrineAmendment(t *testing.T) {
	root := findRepoRoot(t)

	files := []string{
		filepath.Join(root, ".claude", "rules", "moai", "workflow", "lifecycle-sync-gate.md"),
		filepath.Join(root, ".claude", "rules", "moai", "development", "spec-frontmatter-schema.md"),
	}

	// combined/abbreviated scope 문구와 prohibition을 ≤400자 이내 co-located로 매칭한다 (양방향).
	// `(?s)`는 `.`이 개행도 매칭하게 한다 (multi-line amendment block).
	prohibitionFirst := regexp.MustCompile(`(?s)(MUST use individual full-ID|prohibited|disallowed).{0,400}?(combined|abbreviated)[^\n]*scope`)
	scopeFirst := regexp.MustCompile(`(?s)(combined|abbreviated)[^\n]*scope.{0,400}?(MUST use individual full-ID|prohibited|disallowed)`)

	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			content, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			s := string(content)
			if !prohibitionFirst.MatchString(s) && !scopeFirst.MatchString(s) {
				t.Errorf("%s: close-subject full-ID mandate amendment 누락 — combined/abbreviated scope 문구와 prohibition(MUST use individual full-ID / prohibited / disallowed)이 ≤400자 이내 co-located되어야 함 (AC-DLC-011)", f)
			}
		})
	}
}
