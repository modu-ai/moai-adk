package template

import (
	"encoding/json"
	"fmt"

	"github.com/modu-ai/moai-adk-go/internal/config"
)

// Settings represents the Claude Code settings.json structure.
// Generated exclusively via json.MarshalIndent (ADR-011).
type Settings struct {
	Hooks       map[string][]HookGroup `json:"hooks,omitempty"`
	OutputStyle string                 `json:"output_style,omitempty"`
}

// HookGroup represents a group of hooks with an optional matcher.
type HookGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []HookEntry `json:"hooks"`
}

// HookEntry represents a single hook command.
type HookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// SettingsGenerator produces settings.json content from configuration.
type SettingsGenerator interface {
	// Generate creates a valid JSON byte slice for settings.json.
	// Uses Go struct serialization only (ADR-011: no string concatenation).
	Generate(cfg *config.Config, platform string) ([]byte, error)
}

// settingsGenerator is the concrete implementation of SettingsGenerator.
type settingsGenerator struct{}

// NewSettingsGenerator creates a new SettingsGenerator.
func NewSettingsGenerator() SettingsGenerator {
	return &settingsGenerator{}
}

// hookEvents defines the required hook events per REQ-T-023.
var hookEvents = []string{
	"SessionStart",
	"PreToolUse",
	"PostToolUse",
	"SessionEnd",
	"Stop",
	"PreCompact",
}

// Generate builds Settings from config and serializes to JSON.
func (g *settingsGenerator) Generate(cfg *config.Config, platform string) ([]byte, error) {
	settings := Settings{
		Hooks:       buildHooks(platform),
		OutputStyle: resolveOutputStyle(cfg),
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("settings generate marshal: %w", err)
	}

	// Post-generation validation (ADR-011)
	if !json.Valid(data) {
		return nil, fmt.Errorf("%w: generated settings.json is invalid", ErrInvalidJSON)
	}

	// Append trailing newline
	data = append(data, '\n')

	return data, nil
}

// buildHooks constructs the hooks map for all required events.
func buildHooks(platform string) map[string][]HookGroup {
	hooks := make(map[string][]HookGroup, len(hookEvents))

	for _, event := range hookEvents {
		cmd := buildHookCommand(platform, event)
		hooks[event] = []HookGroup{
			{
				Hooks: []HookEntry{
					{
						Type:    "command",
						Command: cmd,
					},
				},
			},
		}
	}

	return hooks
}

// buildHookCommand returns the platform-appropriate hook command string.
func buildHookCommand(platform, event string) string {
	hookSubcommand := eventToSubcommand(event)

	switch platform {
	case "windows":
		return "cmd.exe /c moai hook " + hookSubcommand
	default:
		// darwin, linux, and other unix-like platforms
		return "moai hook " + hookSubcommand
	}
}

// eventToSubcommand converts a PascalCase event name to a kebab-case subcommand.
func eventToSubcommand(event string) string {
	switch event {
	case "SessionStart":
		return "session-start"
	case "PreToolUse":
		return "pre-tool-use"
	case "PostToolUse":
		return "post-tool-use"
	case "SessionEnd":
		return "session-end"
	case "Stop":
		return "stop"
	case "PreCompact":
		return "pre-compact"
	default:
		return event
	}
}

// resolveOutputStyle determines the output style from configuration.
func resolveOutputStyle(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	// Output style can be extended based on config in the future
	return ""
}
