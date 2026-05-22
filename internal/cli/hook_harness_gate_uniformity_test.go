// Package cli — isHarnessLearningEnabled gate uniformity tests (T-C4).
// REQ-HRN-FND-009: verifies that all four handlers (PostToolUse / Stop / SubagentStop / UserPromptSubmit)
// behave identically as complete no-ops when learning.enabled=false.
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// handlerFunc abstracts the signature of a hook handler.
type handlerFunc func(cmd *cobra.Command, args []string) error

// TestGateUniformity_AllHandlersNoOpWhenDisabled is a table-driven test verifying that,
// when learning.enabled=false, none of the four handlers create usage-log.jsonl.
// REQ-HRN-FND-009: the isHarnessLearningEnabled gate applies uniformly to every handler.
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
