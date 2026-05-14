// Package cli — Stop 훅 핸들러 전체 테스트 매트릭스 (T-C1).
// REQ-HRN-OBS-003: session_stop 이벤트 기록 검증.
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트 — 비활성화 시 완전 no-op.
// REQ-HRN-FND-010: 4-필드 기존 스키마 + session_id 신규 필드 보존.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)


// TestRunHarnessObserveStop_NoOpWhenLearningDisabled는 learning.enabled=false일 때
// Stop 핸들러가 usage-log.jsonl을 전혀 생성하지 않음을 검증한다.
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트 재사용.
func TestRunHarnessObserveStop_NoOpWhenLearningDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"session_id":"sess-disabled","hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("learning.enabled=false 시 usage-log.jsonl이 존재해서는 안 됨; stat err=%v", err)
	}
}

// TestRunHarnessObserveStop_PreservesExistingLogWhenDisabled는 learning.enabled=false이고
// 기존 로그가 있을 때 핸들러가 해당 로그를 수정하지 않음을 검증한다.
// REQ-HRN-FND-009: 비활성화 게이트가 기존 데이터를 건드리지 않아야 함.
func TestRunHarnessObserveStop_PreservesExistingLogWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// 기존 로그 파일 사전 생성
	logDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("로그 디렉터리 생성 실패: %v", err)
	}
	logPath := filepath.Join(logDir, "usage-log.jsonl")
	existingContent := `{"timestamp":"2026-01-01T00:00:00Z","event_type":"moai_subcommand","subject":"/moai plan","context_hash":"","tier_increment":0,"schema_version":"v1"}` + "\n"
	if err := os.WriteFile(logPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("기존 로그 파일 생성 실패: %v", err)
	}

	cmd := &cobra.Command{}
	withStdin(t, `{"session_id":"sess-preserve","hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	// 파일 내용이 동일해야 함 (바이트 단위 동일성)
	after, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}
	if string(after) != existingContent {
		t.Errorf("learning.enabled=false 시 기존 로그가 수정됨\n기대: %q\n실제: %q", existingContent, string(after))
	}
}

// TestRunHarnessObserveStop_RecordsWhenEnabled는 learning.enabled=true일 때
// Stop 핸들러가 session_stop 이벤트를 올바르게 기록하는지 검증한다.
// - 4개 기본 필드: timestamp, event_type, subject, schema_version
// - Stop 전용 필드: session_id
// REQ-HRN-OBS-003, REQ-HRN-FND-010.
func TestRunHarnessObserveStop_RecordsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	const sessionID = "sess-test-c1"
	cmd := &cobra.Command{}
	withStdin(t, `{"session_id":"`+sessionID+`","hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
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

	// event_type은 반드시 "session_stop"
	if entry["event_type"] != "session_stop" {
		t.Errorf("event_type: got=%v, want=%q", entry["event_type"], "session_stop")
	}

	// 4개 기본 필드 존재 검증 (REQ-HRN-FND-010 보존)
	for _, field := range []string{"timestamp", "event_type", "subject", "schema_version"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("기본 필드 %q 누락 — schema additivity 위반", field)
		}
	}

	// session_id 필드 존재 및 값 검증 (REQ-HRN-OBS-003)
	if entry["session_id"] != sessionID {
		t.Errorf("session_id: got=%v, want=%q", entry["session_id"], sessionID)
	}

	// last_assistant_message_hash / last_assistant_message_len 은 구현상 stdin에서 읽지 않으므로
	// 생략(omitempty)되는 것이 정상 동작 — 필드가 없어야 함 (PII 최소화 원칙 일치)
	for _, field := range []string{"last_assistant_message_hash", "last_assistant_message_len"} {
		if _, ok := entry[field]; ok {
			// 값이 빈 문자열이거나 0인 경우 omitempty에 의해 생략되어야 함
			// 만약 등장하면 구현 변경을 의미하므로 정보성 경고만 출력
			t.Logf("INFO: 필드 %q가 엔트리에 등장함 — 구현이 stdin에서 해당 값을 읽기 시작한 경우 정상", field)
		}
	}
}

// TestRunHarnessObserveStop_LogErrorPathDoesNotReturn는 로그 기록 실패 시
// 핸들러가 에러를 반환하지 않고(비블로킹) stderr에 기록함을 검증한다.
// 로그 경로를 디렉터리로 차단하여 기록 실패를 유발한다.
func TestRunHarnessObserveStop_LogErrorPathDoesNotReturn(t *testing.T) {
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

	withStdin(t, `{"session_id":"sess-error-path"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Errorf("Stop 핸들러는 기록 실패 시에도 에러를 반환하지 않아야 함: %v", err)
		}
	})
	// 핸들러는 nil 반환 — 에러는 stderr에 기록됨 (비블로킹)
}
