//go:build !race

package search_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/search"
)

// writeJSONL은 테스트용 JSONL 파일을 생성한다.
func writeJSONL(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("JSONL 파일 생성 실패: %v", err)
	}
	return path
}

// TestParseJSONL_ValidRecords는 정상적인 JSONL 레코드를 파싱하는지 확인한다.
func TestParseJSONL_ValidRecords(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	content := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"abc123","message":{"role":"user","content":[{"type":"text","text":"안녕하세요 인증 시스템을 구현하고 싶습니다"}]}}
{"type":"assistant","timestamp":"2026-03-06T10:01:00Z","sessionId":"abc123","message":{"role":"assistant","content":[{"type":"text","text":"JWT 토큰 기반 인증을 구현하는 방법을 안내해드리겠습니다"}]}}
`
	path := writeJSONL(t, tmpDir, "abc123.jsonl", content)

	msgs, err := search.ParseJSONL(path, "main", "/project")
	if err != nil {
		t.Fatalf("ParseJSONL 실패: %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("메시지 수 불일치: 원함=2, 실제=%d", len(msgs))
	}

	// user 메시지 확인
	if msgs[0].Role != "user" {
		t.Errorf("첫 번째 역할 불일치: %q", msgs[0].Role)
	}
	if msgs[0].SessionID != "abc123" {
		t.Errorf("세션 ID 불일치: %q", msgs[0].SessionID)
	}
	if msgs[0].GitBranch != "main" {
		t.Errorf("브랜치 불일치: %q", msgs[0].GitBranch)
	}
	if msgs[0].ProjectPath != "/project" {
		t.Errorf("프로젝트 경로 불일치: %q", msgs[0].ProjectPath)
	}
}

// TestParseJSONL_NoiseFiltering은 노이즈 레코드(XML 태그, 짧은 텍스트)를 필터링하는지 확인한다.
func TestParseJSONL_NoiseFiltering(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	content := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"sess1","message":{"role":"user","content":[{"type":"text","text":"<system-reminder>시스템 주의사항</system-reminder>"}]}}
{"type":"user","timestamp":"2026-03-06T10:01:00Z","sessionId":"sess1","message":{"role":"user","content":[{"type":"text","text":"짧음"}]}}
{"type":"user","timestamp":"2026-03-06T10:02:00Z","sessionId":"sess1","message":{"role":"user","content":[{"type":"text","text":"<function_calls>도구 호출 내용 포함</function_calls>"}]}}
{"type":"user","timestamp":"2026-03-06T10:03:00Z","sessionId":"sess1","message":{"role":"user","content":[{"type":"text","text":"이것은 충분히 긴 텍스트입니다 실제 대화 내용"}]}}
`
	path := writeJSONL(t, tmpDir, "sess1.jsonl", content)

	msgs, err := search.ParseJSONL(path, "main", "/project")
	if err != nil {
		t.Fatalf("ParseJSONL 실패: %v", err)
	}

	// 노이즈 3개 필터링, 유효 메시지 1개만 남아야 함
	if len(msgs) != 1 {
		t.Fatalf("메시지 수 불일치: 원함=1, 실제=%d (노이즈 필터링 실패)", len(msgs))
	}
}

// TestParseJSONL_InvalidJSON은 잘못된 JSON 라인을 건너뛰는지 확인한다.
func TestParseJSONL_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	content := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"sess2","message":{"role":"user","content":[{"type":"text","text":"유효한 첫 번째 메시지입니다 길이 확인용"}]}}
{invalid json line that should be skipped}
{"type":"assistant","timestamp":"2026-03-06T10:01:00Z","sessionId":"sess2","message":{"role":"assistant","content":[{"type":"text","text":"유효한 두 번째 응답입니다 길이 확인용"}]}}
`
	path := writeJSONL(t, tmpDir, "sess2.jsonl", content)

	msgs, err := search.ParseJSONL(path, "main", "/project")
	if err != nil {
		t.Fatalf("ParseJSONL 실패: %v", err)
	}

	// 잘못된 JSON은 건너뛰고 유효한 2개만 반환
	if len(msgs) != 2 {
		t.Fatalf("메시지 수 불일치: 원함=2, 실제=%d", len(msgs))
	}
}

// TestParseJSONL_EmptyFile은 빈 파일을 처리할 때 오류 없이 빈 슬라이스를 반환하는지 확인한다.
func TestParseJSONL_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := writeJSONL(t, tmpDir, "empty.jsonl", "")

	msgs, err := search.ParseJSONL(path, "main", "/project")
	if err != nil {
		t.Fatalf("ParseJSONL 실패: %v", err)
	}

	if msgs == nil {
		t.Error("빈 파일에 대해 nil이 반환됨 (빈 슬라이스여야 함)")
	}
	if len(msgs) != 0 {
		t.Fatalf("빈 파일에 메시지가 반환됨: %d개", len(msgs))
	}
}

// TestParseJSONL_CJKSafeFiltering은 노이즈 필터링이 CJK 문자를 제거하지 않는지 확인한다.
// noisePatterns는 ASCII XML 태그만 필터링하므로 한국어/일본어/중국어 텍스트는 보존된다.
func TestParseJSONL_CJKSafeFiltering(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// 한국어, 일본어, 중국어 메시지 - 노이즈 패턴 없음
	content := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"cjk1","message":{"role":"user","content":[{"type":"text","text":"한국어 인증 구현: JWT 토큰 기반 시스템 설계"}]}}
{"type":"user","timestamp":"2026-03-06T10:01:00Z","sessionId":"cjk1","message":{"role":"user","content":[{"type":"text","text":"日本語認証実装: JWTトークンシステムの設計方法"}]}}
{"type":"user","timestamp":"2026-03-06T10:02:00Z","sessionId":"cjk1","message":{"role":"user","content":[{"type":"text","text":"中文认证实现: 基于JWT令牌的系统设计方案"}]}}
`
	path := writeJSONL(t, tmpDir, "cjk1.jsonl", content)

	msgs, err := search.ParseJSONL(path, "main", "/project")
	if err != nil {
		t.Fatalf("ParseJSONL 실패: %v", err)
	}

	// CJK 메시지 3개 모두 보존되어야 함
	if len(msgs) != 3 {
		t.Fatalf("CJK 메시지 수 불일치: 원함=3, 실제=%d (노이즈 필터가 CJK를 제거)", len(msgs))
	}

	// 한국어 포함 확인
	if !strings.Contains(msgs[0].Text, "인증") {
		t.Error("한국어 텍스트가 손실됨")
	}
	// 일본어 포함 확인
	if !strings.Contains(msgs[1].Text, "認証") {
		t.Error("일본어 텍스트가 손실됨")
	}
	// 중국어 포함 확인
	if !strings.Contains(msgs[2].Text, "认证") {
		t.Error("중국어 텍스트가 손실됨")
	}
}

// TestParseJSONL_RoleFiltering은 user/assistant 이외의 역할을 필터링하는지 확인한다.
func TestParseJSONL_RoleFiltering(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	content := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"sess3","message":{"role":"user","content":[{"type":"text","text":"이것은 유효한 사용자 메시지입니다 길이 충분"}]}}
{"type":"system","timestamp":"2026-03-06T10:01:00Z","sessionId":"sess3","message":{"role":"system","content":[{"type":"text","text":"시스템 메시지는 필터링되어야 합니다 길이 충분"}]}}
{"type":"assistant","timestamp":"2026-03-06T10:02:00Z","sessionId":"sess3","message":{"role":"assistant","content":[{"type":"text","text":"이것은 유효한 어시스턴트 응답입니다 길이 충분"}]}}
{"type":"tool","timestamp":"2026-03-06T10:03:00Z","sessionId":"sess3","message":{"role":"tool","content":[{"type":"text","text":"도구 결과는 필터링되어야 합니다 길이 충분"}]}}
`
	path := writeJSONL(t, tmpDir, "sess3.jsonl", content)

	msgs, err := search.ParseJSONL(path, "main", "/project")
	if err != nil {
		t.Fatalf("ParseJSONL 실패: %v", err)
	}

	// user와 assistant 2개만 반환
	if len(msgs) != 2 {
		t.Fatalf("메시지 수 불일치: 원함=2, 실제=%d", len(msgs))
	}
	for _, m := range msgs {
		if m.Role != "user" && m.Role != "assistant" {
			t.Errorf("허용되지 않는 역할 반환: %q", m.Role)
		}
	}
}
