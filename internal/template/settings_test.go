package template

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk-go/internal/config"
)

func defaultTestConfig() *config.Config {
	return &config.Config{}
}

func TestSettingsGeneratorGenerate(t *testing.T) {
	gen := NewSettingsGenerator()

	t.Run("darwin_valid_json", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		// Trim trailing newline for json.Valid check
		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated settings.json is not valid JSON")
		}
	})

	t.Run("linux_valid_json", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "linux")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated settings.json is not valid JSON")
		}
	})

	t.Run("windows_valid_json", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "windows")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated settings.json is not valid JSON")
		}
	})

	t.Run("json_roundtrip", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		// Unmarshal
		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		// Re-marshal
		redata, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent error: %v", err)
		}

		// Compare (both without trailing newline)
		if string(redata) != trimmed {
			t.Errorf("roundtrip mismatch:\ngot:  %s\nwant: %s", string(redata), trimmed)
		}
	})

	t.Run("no_unexpanded_tokens", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		content := string(data)
		tokens := []string{"${", "{{", "$VAR", "$SHELL"}
		for _, tok := range tokens {
			if strings.Contains(content, tok) {
				t.Errorf("generated JSON contains unexpanded token %q", tok)
			}
		}
	})

	t.Run("nil_config", func(t *testing.T) {
		data, err := gen.Generate(nil, "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated settings.json is not valid JSON with nil config")
		}
	})
}

func TestSettingsRequiredHookEvents(t *testing.T) {
	gen := NewSettingsGenerator()

	requiredEvents := []string{
		"SessionStart",
		"PreToolUse",
		"PostToolUse",
		"SessionEnd",
		"Stop",
		"PreCompact",
	}

	platforms := []string{"darwin", "linux", "windows"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			data, err := gen.Generate(defaultTestConfig(), platform)
			if err != nil {
				t.Fatalf("Generate error: %v", err)
			}

			var settings Settings
			trimmed := strings.TrimSpace(string(data))
			if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			for _, event := range requiredEvents {
				groups, ok := settings.Hooks[event]
				if !ok {
					t.Errorf("missing hook event %q", event)
					continue
				}
				if len(groups) == 0 {
					t.Errorf("hook event %q has no groups", event)
					continue
				}
				if len(groups[0].Hooks) == 0 {
					t.Errorf("hook event %q first group has no hooks", event)
				}
			}
		})
	}
}

func TestSettingsPlatformHookCommands(t *testing.T) {
	gen := NewSettingsGenerator()

	t.Run("darwin_commands", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		cmd := settings.Hooks["SessionStart"][0].Hooks[0].Command
		if cmd != "moai hook session-start" {
			t.Errorf("darwin SessionStart command = %q, want %q", cmd, "moai hook session-start")
		}
	})

	t.Run("linux_commands", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "linux")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		cmd := settings.Hooks["SessionStart"][0].Hooks[0].Command
		if cmd != "moai hook session-start" {
			t.Errorf("linux SessionStart command = %q, want %q", cmd, "moai hook session-start")
		}
	})

	t.Run("windows_commands", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "windows")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		cmd := settings.Hooks["SessionStart"][0].Hooks[0].Command
		expected := "cmd.exe /c moai hook session-start"
		if cmd != expected {
			t.Errorf("windows SessionStart command = %q, want %q", cmd, expected)
		}
	})

	t.Run("all_hook_types_are_command", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		for event, groups := range settings.Hooks {
			for _, group := range groups {
				for _, hook := range group.Hooks {
					if hook.Type != "command" {
						t.Errorf("hook %q type = %q, want %q", event, hook.Type, "command")
					}
				}
			}
		}
	})
}

func TestEventToSubcommand(t *testing.T) {
	tests := []struct {
		event    string
		expected string
	}{
		{"SessionStart", "session-start"},
		{"PreToolUse", "pre-tool-use"},
		{"PostToolUse", "post-tool-use"},
		{"SessionEnd", "session-end"},
		{"Stop", "stop"},
		{"PreCompact", "pre-compact"},
	}

	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			got := eventToSubcommand(tt.event)
			if got != tt.expected {
				t.Errorf("eventToSubcommand(%q) = %q, want %q", tt.event, got, tt.expected)
			}
		})
	}
}
