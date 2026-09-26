package hook

import (
	"encoding/json"
	"os"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// observeEscalation hands one tool-call hook event to the contract-mode
// escalation detector (SPEC-AUTONOMY-ESCALATION-001). It returns nothing: the
// detector only records, so no hook decision can depend on it (REQ-AE-003).
// Under any autonomy mode other than contract it returns after resolving the
// already-loaded configuration, before reading any file (REQ-AE-001).
func observeEscalation(cfg ConfigProvider, hookName string, input *HookInput) {
	observeEscalationResult(cfg, hookName, input, false)
}

// observeEscalationResult is observeEscalation with the call's outcome: failed
// is true for a PostToolUseFailure event; a PostToolUse whose tool_response
// reports a non-zero exit_code is also treated as failed.
func observeEscalationResult(cfg ConfigProvider, hookName string, input *HookInput, failed bool) {
	if cfg == nil || input == nil {
		return
	}
	c := cfg.Get()
	if c == nil {
		return
	}
	s := config.ResolveAutonomy(c.Workflow)
	if !escalation.Active(s) {
		return
	}
	ev := escalation.Event{Hook: hookName, CWD: input.CWD, ToolName: input.ToolName,
		Failed: failed || nonZeroExitCode(input.ToolResponse)}
	if ev.CWD == "" {
		ev.CWD, _ = os.Getwd()
	}
	if len(input.ToolInput) > 0 {
		var in struct {
			FilePath     string `json:"file_path"`
			NotebookPath string `json:"notebook_path"`
			Command      string `json:"command"`
		}
		if json.Unmarshal(input.ToolInput, &in) == nil {
			ev.FilePath, ev.Command = in.FilePath, in.Command
			if ev.FilePath == "" {
				ev.FilePath = in.NotebookPath
			}
		}
	}
	escalation.Observe(s, ev)
}

// nonZeroExitCode reports whether a tool_response object carries a numeric
// exit_code other than zero.
func nonZeroExitCode(resp json.RawMessage) bool {
	if len(resp) == 0 {
		return false
	}
	var r struct {
		ExitCode *float64 `json:"exit_code"`
	}
	return json.Unmarshal(resp, &r) == nil && r.ExitCode != nil && *r.ExitCode != 0
}

// WithEscalationConfig gives a PostToolUse or PostToolUseFailure handler the
// configuration the escalation detector reads, without changing any other
// behavior of the handler (the PostToolUse lint_as_instruction default, for
// one, keeps reading its own cfg). Other handlers are returned unchanged.
func WithEscalationConfig(h Handler, cfg ConfigProvider) Handler {
	switch p := h.(type) {
	case *postToolHandler:
		p.escalationCfg = cfg
	case *postToolUseFailureHandler:
		p.escalationCfg = cfg
	}
	return h
}
