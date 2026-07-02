// Package cli — Stop-path auto-classify chain tests.
// SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-003: the Stop hook path auto-runs the
// classifier (D3-(a) chain), removing the dependency on a manual
// `/moai:harness status` invocation. Fail-open (AP-3): classify errors never
// block session end.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// seedUsageLog pre-populates usage-log.jsonl with n identical user_prompt events
// so the classifier finds at least one aggregatable pattern.
func seedUsageLog(t *testing.T, dir string, n int) {
	t.Helper()
	logDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("usage-log 디렉터리 생성 실패: %v", err)
	}
	line := `{"timestamp":"2026-06-17T10:00:00Z","event_type":"user_prompt","subject":"","context_hash":"","tier_increment":0,"schema_version":"v2.1"}` + "\n"
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(line)
	}
	logPath := filepath.Join(logDir, "usage-log.jsonl")
	if err := os.WriteFile(logPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("usage-log 작성 실패: %v", err)
	}
}

// TestRunHarnessObserveStop_AutoClassifyChain verifies AC-HEP-003a: the Stop
// handler chains the classifier, producing tier-promotions.jsonl automatically.
func TestRunHarnessObserveStop_AutoClassifyChain(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	seedUsageLog(t, dir, 3)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"last_assistant_message":"done","session":{"id":"sess-classify-chain"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("tier-promotions.jsonl 미생성 — Stop 경로 classify 미실행: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) < 1 || lines[0] == "" {
		t.Fatalf("tier-promotions.jsonl 라인 수 = %d, want >= 1 (classify 흔적)", len(lines))
	}
	// The aggregated pattern includes the pre-seeded user_prompt events + the
	// session_stop this handler just recorded — at least one promotion expected.
}

// TestRunHarnessObserveStop_ClassifyFailOpen verifies AC-HEP-003b: when the
// classify path is degraded (usage-log path blocked by a directory), the Stop
// handler still returns nil (exit 0, non-blocking) and logs to stderr.
func TestRunHarnessObserveStop_ClassifyFailOpen(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	// Block the usage-log path with a directory: both the observe record and the
	// classify aggregate fail — the handler must NOT propagate the error.
	blockPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if err := os.MkdirAll(blockPath, 0o755); err != nil {
		t.Fatalf("블로킹 디렉터리 생성 실패: %v", err)
	}

	var stderrBuf strings.Builder
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	withStdin(t, `{"last_assistant_message":"x","session":{"id":"sess-failopen"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Errorf("Stop 핸들러는 classify 실패 시에도 에러를 반환하지 않아야 함(fail-open): %v", err)
		}
	})
	// The tier-promotions.jsonl must not exist (classify could not run).
	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	if _, err := os.Stat(promoPath); err == nil {
		t.Errorf("classify 실패 경로에서 tier-promotions.jsonl이 생성되어서는 안 됨")
	}
}
