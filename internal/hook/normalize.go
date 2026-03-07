package hook

import "encoding/json"

// camelToSnakeMap maps Claude Code's native camelCase top-level field names to
// the flat snake_case format expected by HookInput. Session fields are handled
// separately via flattenSession.
var camelToSnakeMap = map[string]string{
	"eventType":           "hook_event_name",
	"toolName":            "tool_name",
	"toolInput":           "tool_input",
	"toolOutput":          "tool_output",
	"toolResponse":        "tool_response",
	"toolUseId":           "tool_use_id",
	"transcriptPath":      "transcript_path",
	"permissionMode":      "permission_mode",
	"agentId":             "agent_id",
	"agentTranscriptPath": "agent_transcript_path",
	"agentType":           "agent_type",
	"agentName":           "agent_name",
	"stopHookActive":      "stop_hook_active",
	"customInstructions":  "custom_instructions",
	"isInterrupt":         "is_interrupt",
	"notificationType":    "notification_type",
	"teamName":            "team_name",
	"teammateName":        "teammate_name",
	"taskId":              "task_id",
	"taskSubject":         "task_subject",
	"taskDescription":     "task_description",
	"worktreePath":        "worktree_path",
	"worktreeBranch":      "worktree_branch",
}

// normalizeHookInput converts Claude Code's native nested camelCase JSON format
// to the flat snake_case format expected by HookInput.
//
// Claude Code 2.1.x sends hook input in a nested camelCase format:
//
//	{
//	  "eventType": "PreToolUse",
//	  "toolName": "Bash",
//	  "toolInput": { "command": "echo test" },
//	  "session": { "id": "sess-123", "cwd": "/path", "projectDir": "/path" }
//	}
//
// moai-adk's internal representation uses flat snake_case:
//
//	{
//	  "hook_event_name": "PreToolUse",
//	  "tool_name": "Bash",
//	  "tool_input": { "command": "echo test" },
//	  "session_id": "sess-123",
//	  "cwd": "/path",
//	  "project_dir": "/path"
//	}
//
// This function detects the format from the presence of discriminating fields
// and normalizes native format to flat snake_case. Inputs already in flat
// snake_case are returned unchanged (backward compatible).
func normalizeHookInput(data []byte) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Already in flat snake_case format — pass through unchanged.
	if _, ok := raw["session_id"]; ok {
		return data, nil
	}
	if _, ok := raw["hook_event_name"]; ok {
		return data, nil
	}

	// Check for native Claude Code format indicators.
	_, hasSession := raw["session"]
	_, hasEventType := raw["eventType"]
	if !hasSession && !hasEventType {
		// Ambiguous format; return as-is and let validation report what is missing.
		return data, nil
	}

	// Build normalized output, starting with a copy of all existing fields.
	normalized := make(map[string]json.RawMessage, len(raw))
	for k, v := range raw {
		normalized[k] = v
	}

	// Rename top-level camelCase keys to snake_case.
	for camel, snake := range camelToSnakeMap {
		if v, ok := raw[camel]; ok {
			normalized[snake] = v
			delete(normalized, camel)
		}
	}

	// Flatten the nested session object.
	// session.id → session_id
	// session.cwd → cwd
	// session.projectDir → project_dir
	if sessionRaw, ok := raw["session"]; ok {
		var session map[string]json.RawMessage
		if err := json.Unmarshal(sessionRaw, &session); err == nil {
			if id, ok := session["id"]; ok {
				normalized["session_id"] = id
			}
			if cwd, ok := session["cwd"]; ok {
				normalized["cwd"] = cwd
			}
			if projectDir, ok := session["projectDir"]; ok {
				normalized["project_dir"] = projectDir
			}
		}
		delete(normalized, "session")
	}

	return json.Marshal(normalized)
}
