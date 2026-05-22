// Package cli — full SubagentStop hook handler test matrix (T-C2).
// REQ-HRN-OBS-005: verifies subagent_stop events are recorded.
// REQ-HRN-FND-009: isHarnessLearningEnabled gate — full no-op when disabled.
// REQ-HRN-FND-010: preserves the 4-field existing schema plus SubagentStop-specific fields.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunHarnessObserveSubagentStop_NoOpWhenLearningDisabled verifies that the
// SubagentStop handler does not create usage-log.jsonl when learning.enabled=false.
// REQ-HRN-FND-009: reuses the isHarnessLearningEnabled gate.
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

// TestRunHarnessObserveSubagentStop_PreservesExistingLogWhenDisabled verifies that,
// in the disabled state, an existing log file is not modified.
// REQ-HRN-FND-009: when the gate is disabled, existing data must remain unchanged.
func TestRunHarnessObserveSubagentStop_PreservesExistingLogWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// Pre-create an existing log file
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

	// File content must remain identical
	after, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}
	if string(after) != existingContent {
		t.Errorf("learning.enabled=false 시 기존 로그가 수정됨\n기대: %q\n실제: %q", existingContent, string(after))
	}
}

// TestRunHarnessObserveSubagentStop_RecordsAllFields verifies that the SubagentStop
// handler records the subagent_stop event correctly when learning.enabled=true.
// - event_type = "subagent_stop"
// - subject = agent_name (from stdin)
// - agent_name, agent_type, agent_id, parent_session_id fields
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

	// Verify event_type
	if entry["event_type"] != "subagent_stop" {
		t.Errorf("event_type: got=%v, want=%q", entry["event_type"], "subagent_stop")
	}

	// subject must be agent_name (REQ-HRN-OBS-005)
	if entry["subject"] != "manager-develop" {
		t.Errorf("subject: got=%v, want=%q", entry["subject"], "manager-develop")
	}

	// Verify presence of the 4 base fields
	for _, field := range []string{"timestamp", "event_type", "subject", "schema_version"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("기본 필드 %q 누락 — schema additivity 위반", field)
		}
	}

	// Verify SubagentStop-specific fields (REQ-HRN-OBS-005)
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

	// unknown subject fallback: covered by a separate test when agent_name is empty,
	// resulting in subject "unknown". This test covers only the agent_name-present case.
}

// TestRunHarnessObserveSubagentStop_LogErrorPathDoesNotReturn verifies that the
// handler does not return an error on log-write failure (non-blocking) and instead
// writes to stderr.
func TestRunHarnessObserveSubagentStop_LogErrorPathDoesNotReturn(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// Pre-create a directory at the usage-log.jsonl path to induce a file-write failure
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
	// The handler returns nil; errors are written to stderr (non-blocking)
}

// TestRunHarnessObserveSubagentStop_UnknownSubjectFallback verifies that, when
// agent_name is absent, subject falls back to "unknown".
func TestRunHarnessObserveSubagentStop_UnknownSubjectFallback(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	// T-A4 spec: omitting agentName → "unknown" fallback
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
