package hook

// Card t2001 — Monitor tool calls under a Bash-matcher hook (wire format).
//
// Hypothesis under test: Claude Code 2.1.314+ runs Bash-matcher hooks for
// Monitor commands, and a Monitor call arrives with tool_name "Monitor",
// which IsShellTool does not name, so the shell guards would be skipped.
//
// Measured on Claude Code 2.1.316 in a fresh `claude -p` process (evidence:
// .moai/reports/t2001/verdict.md, Claim 1, raw hook rows included): a hook registered on matcher
// "Bash" ONLY received the model's Monitor tool call (the transcript tool_use
// is named Monitor) with tool_name "Bash" and the command under
// tool_input.command. Positive control: the Bash call is recorded; negative
// controls: the Read and ToolSearch calls in the same run are not. The
// hypothesis is refuted for the shipped matcher group "Write|Edit|Bash".
//
// The payload below reproduces that observed shape (cwd, transcript_path and
// ids neutralized). It runs through the real protocol decoder and the real
// PreToolUse handler, so a runtime change that stops delivering Monitor
// commands to Bash guards under the Bash name, or a guard change that stops
// reading the command from this shape, fails here.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

const monitorUnderBashMatcherPayload = `{"cwd":"/tmp/probe","hook_event_name":"PreToolUse","permission_mode":"default","session_id":"s-monitor","tool_input":{"command":"%s","description":"probe"},"tool_name":"Bash","tool_use_id":"toolu_1","transcript_path":"/tmp/probe/t.jsonl"}`

// TestMonitorCommandUnderBashMatcherIsGuarded pins that the observed payload
// for a Monitor call delivered to a Bash-matcher hook is a shell call the
// guards read and decide on.
func TestMonitorCommandUnderBashMatcherIsGuarded(t *testing.T) {
	hmpIsolateHome(t)

	in, err := NewProtocol().ReadInput(strings.NewReader(strings.Replace(monitorUnderBashMatcherPayload, "%s", "terraform destroy", 1)))
	if err != nil {
		t.Fatalf("decode observed Monitor-under-Bash payload: %v", err)
	}
	if !IsShellTool(in.ToolName) {
		t.Fatalf("decoded tool_name %q is not a shell tool", in.ToolName)
	}
	if got := extractBranchStateCommand(in.ToolInput); got != "terraform destroy" {
		t.Fatalf("command read from observed tool_input = %q", got)
	}
	in.CWD = t.TempDir()
	h := hmpHandler(hmpCfg(false, false, config.SlotLeaseConfig{}), in.CWD)
	out, err := h.Handle(context.Background(), in)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if decisionOf(out) != DecisionDeny {
		t.Errorf("observed-shape Monitor command `terraform destroy` decided %q, want deny", decisionOf(out))
	}
}
