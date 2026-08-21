package config

import (
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// todo_enabled.go — the single read point for workflow.todo.enabled
// (SPEC-TODO-ENABLE-FLAG-001 REQ-1).
//
// The flag suppresses the runtime todo GUIDANCE surfaces (the SessionStart
// backlog line, the statusline TODO segment, the skill's automatic routing to
// the todo workflow). It never removes the feature: the `moai todo` command
// stays registered and every verb keeps working (REQ-3), and an explicit
// `/moai todo` invocation still runs (REQ-2 D2).

// TodoEnabled reports whether the todo guidance surfaces are active.
//
// Absent key ⇒ enabled. This is the whole reason WorkflowTodoConfig.Enabled is
// a *bool: the distributed template carries no todo block, so "unset" is the
// state nearly every user is in, and it must not be confused with an explicit
// opt-out. Only a literal `enabled: false` disables.
//
// Nil-receiver safe and fail-OPEN, matching readMCPToolEnablement's posture
// (internal/cli/mcp_server.go): a caller that could not load config gets the
// guidance rather than silence, because a surface that vanished for an
// unreadable-config reason is far harder to diagnose than one that stayed.
func (c *Config) TodoEnabled() bool {
	if c == nil || c.Workflow.Todo.Enabled == nil {
		return true
	}
	return *c.Workflow.Todo.Enabled
}

// TodoEnabledForRoot is the project-root convenience form the runtime surfaces
// call. projectRoot is the project directory (the parent of .moai), NOT the
// .moai directory itself.
//
// Every failure path returns true: an empty root, a missing .moai tree, and a
// load error all resolve to enabled. Note that a malformed workflow.yaml does
// not even reach the error path here — loadWorkflowSection swallows the
// unmarshal failure and keeps the construction-time defaults for the WHOLE
// workflow section, so the key still reads as enabled (see the fourth case of
// TestTodoEnabled).
func TodoEnabledForRoot(projectRoot string) bool {
	if projectRoot == "" {
		return true
	}
	cfg, err := NewLoader().Load(filepath.Join(projectRoot, defs.MoAIDir))
	if err != nil {
		return true
	}
	return cfg.TodoEnabled()
}
