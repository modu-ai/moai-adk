package hook

import (
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
// Returns ErrHookInvalidInput if the JSON is malformed or required fields are missing.
func (p *jsonProtocol) ReadInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHookInvalidInput, err)
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

	// session_id 누락 허용: Claude Code 버전에 따라 일부 이벤트에서 session_id가
	// 전송되지 않는 경우가 있음. 훅 실행 실패보다 graceful fallback이 바람직함.
	if input.SessionID == "" {
		input.SessionID = "unknown"
	}
	// cwd 누락 허용: $CLAUDE_PROJECT_DIR 환경변수로 폴백 가능.
	// 훅 핸들러에서 cwd가 빈 경우 환경변수를 확인하도록 구현됨.
	if input.CWD == "" {
		if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
			input.CWD = projectDir
		}
	}
	if input.HookEventName == "" {
		return fmt.Errorf("%w: missing required field hook_event_name", ErrHookInvalidInput)
	}
	return nil
}
