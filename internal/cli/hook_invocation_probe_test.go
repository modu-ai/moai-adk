package cli

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestRunHookEvent_RecordsInvocationProbe(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	logPath := filepath.Join(t.TempDir(), "invocations.jsonl")
	t.Setenv(hookInvocationLogEnv, logPath)
	t.Setenv(hookHostEnv, "codex")

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{
					ToolName:  "Bash",
					SessionID: "sess-probe",
					CWD:       "/tmp/project",
				}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, input *hook.HookInput) (*hook.HookOutput, error) {
				if event != hook.EventPreToolUse {
					t.Fatalf("event = %q, want %q", event, hook.EventPreToolUse)
				}
				if input.ToolName != "Bash" {
					t.Fatalf("tool = %q, want Bash", input.ToolName)
				}
				return &hook.HookOutput{}, nil
			},
		},
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	if err := runHookEvent(cmd, hook.EventPreToolUse); err != nil {
		t.Fatalf("runHookEvent error: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read invocation log: %v", err)
	}

	var entry hookInvocationProbeEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal invocation log: %v\n%s", err, data)
	}

	if entry.Host != "codex" {
		t.Errorf("host = %q, want codex", entry.Host)
	}
	if entry.Event != hook.EventPreToolUse {
		t.Errorf("event = %q, want %q", entry.Event, hook.EventPreToolUse)
	}
	if entry.HookEventName != string(hook.EventPreToolUse) {
		t.Errorf("hook_event_name = %q, want %q", entry.HookEventName, hook.EventPreToolUse)
	}
	if entry.ToolName != "Bash" {
		t.Errorf("tool_name = %q, want Bash", entry.ToolName)
	}
	if entry.SessionID != "sess-probe" {
		t.Errorf("session_id = %q, want sess-probe", entry.SessionID)
	}
	if entry.CWD != "/tmp/project" {
		t.Errorf("cwd = %q, want /tmp/project", entry.CWD)
	}
}
