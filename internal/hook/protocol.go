package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// jsonProtocol implements the Protocol interface using encoding/json.
// It communicates with Claude Code via JSON stdin/stdout as specified
// in REQ-HOOK-010 through REQ-HOOK-013.
type jsonProtocol struct{}

// NewProtocol creates a new Protocol instance for Claude Code JSON communication.
func NewProtocol() Protocol {
	return &jsonProtocol{}
}

// ReadInput reads and parses a JSON payload from the given reader.
// It accepts both Claude Code's native nested camelCase format and the legacy
// flat snake_case format. Native format (session.id, eventType, toolName) is
// normalized to flat snake_case before decoding.
// It validates required fields: session_id, cwd, and hook_event_name.
//
// Empty, blank, or whitespace-only stdin is treated as a graceful no-op success:
// a non-blocking observer hook (e.g. PreToolUse on Bash) must never fail the tool
// it observes, so an absent payload yields a default *HookInput and a nil error
// rather than ErrHookInvalidInput. This extends the existing graceful-fallback
// philosophy already applied to missing session_id / cwd / hook_event_name in
// validateInput. Non-empty malformed JSON still returns ErrHookInvalidInput.
//
// Returns ErrHookInvalidInput if non-empty JSON is malformed or otherwise unparseable.
func (p *jsonProtocol) ReadInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHookInvalidInput, err)
	}

	// Empty / blank / whitespace-only stdin: graceful no-op success. Return a
	// default *HookInput (with validateInput defaults applied) and nil error so
	// the caller exits 0 instead of surfacing a tool-execution failure.
	if len(bytes.TrimSpace(data)) == 0 {
		input := &HookInput{}
		_ = validateInput(input)
		return input, nil
	}

	normalized, err := normalizeHookInput(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHookInvalidInput, err)
	}

	var input HookInput
	if err := json.Unmarshal(normalized, &input); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHookInvalidInput, err)
	}

	if err := validateInput(&input); err != nil {
		return nil, err
	}

	return &input, nil
}

// WriteOutput serializes the HookOutput as JSON to the given writer.
// If output is nil, an empty JSON object is written.
// All JSON is produced via json.Marshal (REQ-HOOK-012: no string concatenation).
func (p *jsonProtocol) WriteOutput(w io.Writer, output *HookOutput) error {
	if output == nil {
		output = &HookOutput{}
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("write hook output: %w", err)
	}

	return nil
}

// validateInput checks that all required fields are present in the HookInput.
//
// session_id is optional for:
//   - PostToolUse/PostToolUseFailure: Claude Code bug #541 (matcher not respected)
//   - UserPromptSubmit: Claude Code 2.1.97+ sends minimal payload (issue #615)
//
// hook_event_name is inferred from the prompt field when absent:
//   - If prompt is non-empty and hook_event_name is empty → UserPromptSubmit
//
// cwd is optional for UserPromptSubmit (defaults to empty, handler falls back gracefully).
func validateInput(input *HookInput) error {
	// Infer event type from prompt field when hook_event_name is missing.
	// Claude Code 2.1.97+ may send only { "prompt": "..." } for UserPromptSubmit.
	if input.HookEventName == "" && input.Prompt != "" {
		input.HookEventName = string(EventUserPromptSubmit)
	}

	// Allow missing session_id: some events in certain Claude Code versions do not
	// send session_id. A graceful fallback is preferable to hook execution failure.
	if input.SessionID == "" {
		input.SessionID = "unknown"
	}
	// Allow missing cwd: can fall back to the $CLAUDE_PROJECT_DIR env var.
	// Hook handlers check the env var when cwd is empty.
	if input.CWD == "" {
		if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
			input.CWD = projectDir
		}
	}
	// hook_event_name is not strictly required: CLI subcommands (moai hook stop, etc.)
	// already know the event type and inject it via runHookEvent after parsing.
	// Claude Code may omit hook_event_name for events like Stop and SubagentStop.
	if input.HookEventName == "" {
		input.HookEventName = "unknown"
	}
	return nil
}
