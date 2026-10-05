package hook

// REQ-AE-006 wiring: a failed Bash call reaches the detector through the
// PostToolUseFailure handler (and a PostToolUse whose tool_response carries a
// non-zero exit_code), so an invariant command that fails trips
// invariant-violation (command); the handler's own output is unchanged.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

func escalationProvider(t *testing.T, w *escalationtest.Worktree) ConfigProvider {
	t.Helper()
	cfg, err := config.NewLoader().Load(w.Path(".moai"))
	if err != nil {
		t.Fatal(err)
	}
	return staticConfigProvider{cfg: cfg}
}

func TestInvariantFailureReachesDetectorFromHooks(t *testing.T) {
	for _, via := range []string{"post-tool-use-failure", "post-tool-use-exit-code"} {
		t.Run(via, func(t *testing.T) {
			w := escalationHookFixture(t, "")
			w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
			pre, post := escalationHandlers(t, w)
			if _, err := pre.Handle(context.Background(), escalationInput(w, "Write",
				map[string]string{"file_path": w.Path("internal/fixture/a.go"), "content": "x\n"})); err != nil {
				t.Fatal(err)
			}
			in := escalationInput(w, "Bash", map[string]string{"command": "go test ./..."})
			var out *HookOutput
			var err error
			if via == "post-tool-use-failure" {
				failure := WithEscalationConfig(NewPostToolUseFailureHandler(), escalationProvider(t, w))
				in.Error = "Exit code 1"
				want, _ := NewPostToolUseFailureHandler().Handle(context.Background(), in)
				out, err = failure.Handle(context.Background(), in)
				wb, _ := json.Marshal(want)
				ob, _ := json.Marshal(out)
				if string(wb) != string(ob) {
					t.Errorf("failure handler output changed:\n got %s\nwant %s", ob, wb)
				}
			} else {
				in.ToolResponse = json.RawMessage(`{"stdout":"","stderr":"FAIL","exit_code":1}`)
				out, err = post.Handle(context.Background(), in)
			}
			if err != nil || out == nil {
				t.Fatalf("Handle: %v", err)
			}
			var n int
			for _, name := range escalationRecordNames(t, w) {
				if strings.HasPrefix(name, escalation.ClassInvariantViolation+"-") {
					n++
				}
			}
			if n != 1 {
				t.Errorf("invariant-violation records = %d (%v), want 1", n, escalationRecordNames(t, w))
			}
		})
	}
}
