package hook

// SPEC-HOOK-MATCHER-POWERSHELL-001 (card t1224) — REQ-HMP-016 wire format.
//
// The payloads below reproduce the key set and value shapes Claude Code
// 2.1.283 delivered to a matcher-"PowerShell" hook in the LIVE arm-B capture
// (.moai/reports/t1224/live/B/hook.log and post.log; hashes in progress.md).
// Only cwd, transcript_path and the ids are neutralized. They run through the
// real protocol decoder, so a runtime shape change fails here rather than
// leaving the PowerShell guards and evidence path silently reading nothing.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

const hmpLivePreToolPayload = `{"cwd":"/tmp/probe","effort":{"level":"medium"},"hook_event_name":"PreToolUse","permission_mode":"bypassPermissions","prompt_id":"p-1","session_id":"s-live","tool_input":{"command":"%s","description":"probe"},"tool_name":"PowerShell","tool_use_id":"toolu_1","transcript_path":"/tmp/probe/t.jsonl"}`

const hmpLivePostToolPayload = `{"cwd":"/tmp/probe","duration_ms":812,"effort":{"level":"medium"},"hook_event_name":"PostToolUse","permission_mode":"bypassPermissions","prompt_id":"p-1","session_id":"s-live","tool_input":{"command":"%s","description":"probe"},"tool_name":"PowerShell","tool_response":{"stdout":"%s","stderr":"","interrupted":false,"isImage":false},"tool_use_id":"toolu_1","transcript_path":"/tmp/probe/t.jsonl"}`

func hmpDecodeLive(t *testing.T, payload string) *HookInput {
	t.Helper()
	in, err := NewProtocol().ReadInput(strings.NewReader(payload))
	if err != nil {
		t.Fatalf("decode observed PowerShell payload: %v", err)
	}
	return in
}

// TestHMPLiveWireFormat pins that the observed PowerShell payload shape feeds
// the fields the guards and the evidence writer read.
func TestHMPLiveWireFormat(t *testing.T) {
	hmpIsolateHome(t)

	pre := hmpDecodeLive(t, strings.Replace(hmpLivePreToolPayload, "%s", "terraform destroy", 1))
	if pre.ToolName != "PowerShell" || !IsShellTool(pre.ToolName) {
		t.Fatalf("decoded tool_name %q is not a shell tool", pre.ToolName)
	}
	if got := extractBranchStateCommand(pre.ToolInput); got != "terraform destroy" {
		t.Fatalf("command read from observed tool_input = %q", got)
	}
	pre.CWD = t.TempDir()
	h := hmpHandler(hmpCfg(false, false, config.SlotLeaseConfig{}), pre.CWD)
	out, err := h.Handle(context.Background(), pre)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if decisionOf(out) != DecisionDeny {
		t.Errorf("observed-shape PowerShell `terraform destroy` decided %q, want deny", decisionOf(out))
	}

	post := hmpDecodeLive(t, strings.Replace(strings.Replace(hmpLivePostToolPayload, "%s", "go test ./x/...", 1), "%s", `ok  \tgithub.com/x/y\t0.42s\n`, 1))
	rec, ok := buildEvidenceRecord(post)
	if !ok || !rec.IsTestPass || rec.IsTestFail {
		t.Errorf("observed-shape PowerShell go test record ok=%v pass=%v fail=%v, want a pass", ok, rec.IsTestPass, rec.IsTestFail)
	}
}
