package hook

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// escalationOptions carries what only the calling handler knows.
type escalationOptions struct {
	// failed is true for a PostToolUseFailure event.
	failed bool
	// denylisted, when set, reports whether the existing destructive-command
	// denylist denies the shell command; called only under contract mode.
	denylisted func() bool
}

// observeEscalation hands one tool-call hook event to the contract-mode
// escalation detector (SPEC-AUTONOMY-ESCALATION-001). It returns nothing: the
// detector only records, so no hook decision can depend on it (REQ-AE-003).
// Under any autonomy mode other than contract it returns after resolving the
// already-loaded configuration, before reading any file (REQ-AE-001).
func observeEscalation(cfg ConfigProvider, hookName string, input *HookInput) {
	observeEscalationWith(cfg, hookName, input, escalationOptions{})
}

// observeEscalationWith is observeEscalation with handler-supplied facts. A
// PostToolUse whose tool_response reports a non-zero exit_code is also
// treated as failed.
func observeEscalationWith(cfg ConfigProvider, hookName string, input *HookInput, o escalationOptions) {
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
		Failed: o.failed || nonZeroExitCode(input.ToolResponse)}
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
	if ev.Failed {
		ev.Diagnostic = failureText(input)
	}
	if o.denylisted != nil && ev.Command != "" {
		ev.Denylisted = o.denylisted()
	}
	escalation.Observe(s, ev)
}

// failureText is the failure text of a failed call: the error field, then the
// tool_response text (stderr first).
func failureText(input *HookInput) string {
	var b strings.Builder
	b.WriteString(input.Error)
	if len(input.ToolResponse) > 0 {
		if text, _ := decodeToolResponse(input.ToolResponse); text != "" {
			b.WriteByte('\n')
			b.WriteString(text)
		}
	}
	return b.String()
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

// WithEscalationConfig gives a PostToolUse, PostToolUseFailure, or Stop
// handler the configuration the escalation detector reads, without changing
// any other behavior of the handler (the PostToolUse lint_as_instruction
// default, for one, keeps reading its own cfg). Other handlers are returned
// unchanged.
func WithEscalationConfig(h Handler, cfg ConfigProvider) Handler {
	switch p := h.(type) {
	case *postToolHandler:
		p.escalationCfg = cfg
	case *postToolUseFailureHandler:
		p.escalationCfg = cfg
	case *stopHandler:
		p.escalationCfg = cfg
	}
	return h
}
