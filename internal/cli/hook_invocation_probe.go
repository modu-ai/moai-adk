package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
)

const (
	hookInvocationLogEnv = "MOAI_HOOK_INVOCATION_LOG"
	hookHostEnv          = "MOAI_HOOK_HOST"
)

type hookInvocationProbeEntry struct {
	Timestamp     time.Time      `json:"ts"`
	Host          string         `json:"host,omitempty"`
	Event         hook.EventType `json:"event"`
	HookEventName string         `json:"hook_event_name,omitempty"`
	ToolName      string         `json:"tool_name,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
	CWD           string         `json:"cwd,omitempty"`
}

func recordHookInvocationProbe(event hook.EventType, input *hook.HookInput) error {
	logPath := os.Getenv(hookInvocationLogEnv)
	if logPath == "" {
		return nil
	}

	entry := hookInvocationProbeEntry{
		Timestamp: time.Now().UTC(),
		Host:      os.Getenv(hookHostEnv),
		Event:     event,
	}
	if input != nil {
		entry.HookEventName = input.HookEventName
		entry.ToolName = input.ToolName
		entry.SessionID = input.SessionID
		entry.CWD = input.CWD
	}

	if dir := filepath.Dir(logPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create hook invocation log dir: %w", err)
		}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open hook invocation log: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := json.NewEncoder(f).Encode(entry); err != nil {
		return fmt.Errorf("write hook invocation log: %w", err)
	}
	return nil
}
