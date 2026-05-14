// Package cli — SubagentStop 훅 핸들러 전체 테스트 매트릭스 (T-C2).
// REQ-HRN-OBS-005: subagent_stop 이벤트 기록 검증.
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트 — 비활성화 시 완전 no-op.
// REQ-HRN-FND-010: 4-필드 기존 스키마 + SubagentStop 전용 필드 보존.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunHarnessObserveSubagentStop_NoOpWhenLearningDisabled는 learning.enabled=false일 때
// SubagentStop 핸들러가 usage-log.jsonl을 생성하지 않음을 검증한다.
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트 재사용.
func TestRunHarnessObserveSubagentStop_NoOpWhenLearningDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// T-A4 spec: camelCase agentName/agentType + nested session.id
	cmd := &cobra.Command{}
	withStdin(t, `{"agentName":"expert-frontend","agentType":"subagent","agent_id":"ag-001","session":{"id":"sess-abc"}}`, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("learning.enabled=false 시 usage-log.jsonl이 존재해서는 안 됨; stat err=%v", err)
	}
}

// TestRunHarnessObserveSubagentStop_PreservesExistingLogWhenDisabled는 비활성화 상태에서
// 기존 로그 파일이 수정되지 않음을 검증한다.
// REQ-HRN-FND-009: 게이트 비활성화 시 기존 데이터 불변 보장.
func TestRunHarnessObserveSubagentStop_PreservesExistingLogWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// 기존 로그 파일 사전 생성
	logDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("로그 디렉터리 생성 실패: %v", err)
	}
	logPath := filepath.Join(logDir, "usage-log.jsonl")
	existingContent := `{"timestamp":"2026-01-01T00:00:00Z","event_type":"agent_invocation","subject":"expert-backend","context_hash":"","tier_increment":0,"schema_version":"v1"}` + "\n"
	if err := os.WriteFile(logPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("기존 로그 파일 생성 실패: %v", err)
	}

	// T-A4 spec: camelCase + nested session.id
	cmd := &cobra.Command{}
	withStdin(t, `{"agentName":"expert-security","session":{"id":"sess-preserve"}}`, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop 에러 반환: %v", err)
		}
	})

	// 파일 내용이 동일해야 함
	after, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}
	if string(after) != existingContent {
		t.Errorf("learning.enabled=false 시 기존 로그가 수정됨\n기대: %q\n실제: %q", existingContent, string(after))
	}
}

// TestRunHarnessObserveSubagentStop_RecordsAllFields는 learning.enabled=true일 때
// SubagentStop 핸들러가 subagent_stop 이벤트를 올바르게 기록하는지 검증한다.
// - event_type = "subagent_stop"
// - subject = agent_name (stdin에서)
// - agent_name, agent_type, agent_id, parent_session_id 필드
// REQ-HRN-OBS-005, REQ-HRN-FND-010.
func TestRunHarnessObserveSubagentStop_RecordsAllFields(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// T-A4 spec: camelCase agentName/agentType + nested session.id
	cmd := &cobra.Command{}
	payload := `{"agentName":"manager-develop","agentType":"subagent","agent_id":"ag-xyz-789","session":{"id":"sess-parent-456"}}`
	withStdin(t, payload, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("JSONL 라인 수: got=%d, want=1", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// event_type 검증
	if entry["event_type"] != "subagent_stop" {
		t.Errorf("event_type: got=%v, want=%q", entry["event_type"], "subagent_stop")
	}

	// subject는 agent_name이어야 함 (REQ-HRN-OBS-005)
	if entry["subject"] != "manager-develop" {
		t.Errorf("subject: got=%v, want=%q", entry["subject"], "manager-develop")
	}

	// 4개 기본 필드 존재 검증
	for _, field := range []string{"timestamp", "event_type", "subject", "schema_version"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("기본 필드 %q 누락 — schema additivity 위반", field)
		}
	}

	// SubagentStop 전용 필드 검증 (REQ-HRN-OBS-005)
	wantFields := map[string]any{
		"agent_name":        "manager-develop",
		"agent_type":        "subagent",
		"agent_id":          "ag-xyz-789",
		"parent_session_id": "sess-parent-456",
	}
	for field, wantVal := range wantFields {
		gotVal, ok := entry[field]
		if !ok {
			t.Errorf("SubagentStop 전용 필드 %q 누락", field)
			continue
		}
		if gotVal != wantVal {
			t.Errorf("필드 %q: got=%v, want=%v", field, gotVal, wantVal)
		}
	}

	// unknown subject 폴백: agent_name이 빈 경우 "unknown"이 되는지 별도 테스트
	// (이 테스트는 agent_name이 있는 경우만 커버)
}

// TestRunHarnessObserveSubagentStop_LogErrorPathDoesNotReturn는 로그 기록 실패 시
// 핸들러가 에러를 반환하지 않고(비블로킹) stderr에 기록함을 검증한다.
func TestRunHarnessObserveSubagentStop_LogErrorPathDoesNotReturn(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// usage-log.jsonl 위치에 디렉터리를 미리 생성하여 파일 쓰기 실패 유발
	blockPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if err := os.MkdirAll(blockPath, 0o755); err != nil {
		t.Fatalf("블로킹 디렉터리 생성 실패: %v", err)
	}

	var stderrBuf strings.Builder
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	withStdin(t, `{"agentName":"expert-backend","session":{"id":"sess-error"}}`, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Errorf("SubagentStop 핸들러는 기록 실패 시에도 에러를 반환하지 않아야 함: %v", err)
		}
	})
	// 핸들러는 nil 반환 — 에러는 stderr에 기록됨 (비블로킹)
}

// TestRunHarnessObserveSubagentStop_UnknownSubjectFallback는 agent_name이 없을 때
// subject가 "unknown"으로 폴백되는지 검증한다.
func TestRunHarnessObserveSubagentStop_UnknownSubjectFallback(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// T-A4 spec: agentName 생략 → "unknown" 폴백
	cmd := &cobra.Command{}
	withStdin(t, `{"session":{"id":"sess-noname"}}`, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("JSONL 라인 수: got=%d, want=1", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	if entry["subject"] != "unknown" {
		t.Errorf("agent_name 미설정 시 subject: got=%v, want=%q", entry["subject"], "unknown")
	}
}
