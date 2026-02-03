package hook

import (
	"encoding/json"
	"fmt"
	"io"
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
// It validates required fields: session_id, cwd, and hook_event_name.
// Returns ErrHookInvalidInput if the JSON is malformed or required fields are missing.
func (p *jsonProtocol) ReadInput(r io.Reader) (*HookInput, error) {
	var input HookInput

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&input); err != nil {
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
func validateInput(input *HookInput) error {
	if input.SessionID == "" {
		return fmt.Errorf("%w: missing required field session_id", ErrHookInvalidInput)
	}
	if input.CWD == "" {
		return fmt.Errorf("%w: missing required field cwd", ErrHookInvalidInput)
	}
	if input.HookEventName == "" {
		return fmt.Errorf("%w: missing required field hook_event_name", ErrHookInvalidInput)
	}
	return nil
}
