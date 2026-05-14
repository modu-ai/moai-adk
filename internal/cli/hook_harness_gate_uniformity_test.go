// Package cli — isHarnessLearningEnabled 게이트 균일성 테스트 (T-C4).
// REQ-HRN-FND-009: 모든 4개 핸들러(PostToolUse/Stop/SubagentStop/UserPromptSubmit)가
// learning.enabled=false 시 동일하게 완전 no-op 동작을 보임을 검증한다.
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// handlerFunc는 훅 핸들러의 시그니처를 추상화한 타입이다.
type handlerFunc func(cmd *cobra.Command, args []string) error

// TestGateUniformity_AllHandlersNoOpWhenDisabled는 learning.enabled=false 설정 시
// 4개 핸들러 모두가 usage-log.jsonl을 생성하지 않음을 검증하는 테이블-드리븐 테스트.
// REQ-HRN-FND-009: isHarnessLearningEnabled 게이트가 모든 핸들러에 동일하게 적용됨.
func TestGateUniformity_AllHandlersNoOpWhenDisabled(t *testing.T) {
	cases := []struct {
		name    string
		handler handlerFunc
		stdin   string
	}{
		{
			name:    "PostToolUse",
			handler: runHarnessObserve,
			stdin:   `{"tool_name":"Edit","tool_input":{"path":"test.go"},"hook_event_name":"PostToolUse"}`,
		},
		{
			name:    "Stop",
			handler: runHarnessObserveStop,
			// T-A3 spec: nested session.id
			stdin: `{"last_assistant_message":"","session":{"id":"sess-gate-test"},"hook_event_name":"Stop"}`,
		},
		{
			name:    "SubagentStop",
			handler: runHarnessObserveSubagentStop,
			// T-A4 spec: camelCase agentName/agentType + nested session.id
			stdin: `{"agentName":"expert-backend","agentType":"subagent","session":{"id":"sess-gate-test"}}`,
		},
		{
			name:    "UserPromptSubmit",
			handler: runHarnessObserveUserPromptSubmit,
			stdin:   `{"prompt":"gate uniformity test prompt"}`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
			t.Chdir(dir)

			cmd := &cobra.Command{}
			withStdin(t, tc.stdin, func() {
				if err := tc.handler(cmd, nil); err != nil {
					t.Fatalf("%s 핸들러 에러 반환: %v", tc.name, err)
				}
			})

			logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
			if _, err := os.Stat(logPath); !os.IsNotExist(err) {
				t.Errorf(
					"[%s] learning.enabled=false 시 usage-log.jsonl이 존재해서는 안 됨; stat err=%v",
					tc.name, err,
				)
			}
		})
	}
}
