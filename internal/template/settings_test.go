package template

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
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

	t.Run("no_unexpanded_template_tokens", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		content := string(data)
		// Note: $HOME is intentionally included in the generated settings.json
		// Only check for template-style tokens that shouldn't be present
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

	// Note: SessionEnd is excluded - handled by global moai-rank hook
	requiredEvents := []string{
		"SessionStart",
		"PreToolUse",
		"PostToolUse",
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
		expected := `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh"`
		if cmd != expected {
			t.Errorf("darwin SessionStart command = %q, want %q", cmd, expected)
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
		expected := `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh"`
		if cmd != expected {
			t.Errorf("linux SessionStart command = %q, want %q", cmd, expected)
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
		expected := `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh"`
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

	t.Run("hooks_have_timeout", func(t *testing.T) {
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
					if hook.Timeout <= 0 {
						t.Errorf("hook %q has invalid timeout %d", event, hook.Timeout)
					}
				}
			}
		}
	})

	t.Run("pretooluse_has_matcher", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		groups := settings.Hooks["PreToolUse"]
		if len(groups) == 0 {
			t.Fatal("PreToolUse has no groups")
		}
		if groups[0].Matcher != "Write|Edit|Bash" {
			t.Errorf("PreToolUse matcher = %q, want %q", groups[0].Matcher, "Write|Edit|Bash")
		}
	})

	t.Run("posttooluse_has_matcher", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		groups := settings.Hooks["PostToolUse"]
		if len(groups) == 0 {
			t.Fatal("PostToolUse has no groups")
		}
		if groups[0].Matcher != "Write|Edit" {
			t.Errorf("PostToolUse matcher = %q, want %q", groups[0].Matcher, "Write|Edit")
		}
	})
}

func TestSettingsStatusLineType(t *testing.T) {
	gen := NewSettingsGenerator()

	t.Run("statusline_has_type_command", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if settings.StatusLine == nil {
			t.Fatal("statusLine is nil")
		}
		if settings.StatusLine.Type != "command" {
			t.Errorf("statusLine.type = %q, want %q", settings.StatusLine.Type, "command")
		}
		if settings.StatusLine.Command != ".moai/status_line.sh" {
			t.Errorf("statusLine.command = %q, want %q", settings.StatusLine.Command, ".moai/status_line.sh")
		}
	})
}

func TestSettingsOutputStyleDefault(t *testing.T) {
	gen := NewSettingsGenerator()

	t.Run("default_moai_style", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		expected := "MoAI"
		if settings.OutputStyle != expected {
			t.Errorf("outputStyle = %q, want %q", settings.OutputStyle, expected)
		}
	})

	t.Run("nil_config_still_has_style", func(t *testing.T) {
		data, err := gen.Generate(nil, "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		expected := "MoAI"
		if settings.OutputStyle != expected {
			t.Errorf("nil config outputStyle = %q, want %q", settings.OutputStyle, expected)
		}
	})
}

func TestSettingsEnvDefault(t *testing.T) {
	gen := NewSettingsGenerator()

	t.Run("env_managed_globally", func(t *testing.T) {
		// Env is included in project-level settings with PATH, ENABLE_TOOL_SEARCH, and AGENT_TEAMS
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		// Env should be populated with required environment variables
		if settings.Env == nil {
			t.Fatal("env should not be nil")
		}

		// Verify required keys exist
		if _, ok := settings.Env["PATH"]; !ok {
			t.Error("env should contain PATH key")
		}
		if settings.Env["ENABLE_TOOL_SEARCH"] != "1" {
			t.Errorf("ENABLE_TOOL_SEARCH = %q, want %q", settings.Env["ENABLE_TOOL_SEARCH"], "1")
		}
		if settings.Env["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] != "1" {
			t.Errorf("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = %q, want %q", settings.Env["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"], "1")
		}
	})

	t.Run("has_cleanup_period", func(t *testing.T) {
		data, err := gen.Generate(defaultTestConfig(), "darwin")
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}

		var settings Settings
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if settings.CleanupPeriodDays != 30 {
			t.Errorf("cleanupPeriodDays = %d, want 30", settings.CleanupPeriodDays)
		}
	})
}

// --- MCP Generator Tests ---

func TestMCPGeneratorGenerateMCP(t *testing.T) {
	gen := NewMCPGenerator()

	t.Run("darwin_valid_json", func(t *testing.T) {
		data, err := gen.GenerateMCP("darwin")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated .mcp.json is not valid JSON")
		}
	})

	t.Run("linux_valid_json", func(t *testing.T) {
		data, err := gen.GenerateMCP("linux")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated .mcp.json is not valid JSON")
		}
	})

	t.Run("windows_valid_json", func(t *testing.T) {
		data, err := gen.GenerateMCP("windows")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		trimmed := strings.TrimSpace(string(data))
		if !json.Valid([]byte(trimmed)) {
			t.Fatal("generated .mcp.json is not valid JSON")
		}
	})

	t.Run("json_roundtrip", func(t *testing.T) {
		data, err := gen.GenerateMCP("darwin")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		var cfg MCPConfig
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		redata, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent error: %v", err)
		}

		if string(redata) != trimmed {
			t.Errorf("roundtrip mismatch:\ngot:  %s\nwant: %s", string(redata), trimmed)
		}
	})

	t.Run("no_unexpanded_tokens", func(t *testing.T) {
		data, err := gen.GenerateMCP("darwin")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		content := string(data)
		tokens := []string{"${", "{{", "$VAR", "$SHELL", "MCP_SHELL"}
		for _, tok := range tokens {
			if strings.Contains(content, tok) {
				t.Errorf("generated JSON contains unexpanded token %q", tok)
			}
		}
	})
}

func TestMCPGeneratorRequiredServers(t *testing.T) {
	gen := NewMCPGenerator()

	requiredServers := []string{"context7", "sequential-thinking"}
	platforms := []string{"darwin", "linux", "windows"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			data, err := gen.GenerateMCP(platform)
			if err != nil {
				t.Fatalf("GenerateMCP error: %v", err)
			}

			var cfg MCPConfig
			trimmed := strings.TrimSpace(string(data))
			if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			for _, name := range requiredServers {
				server, ok := cfg.MCPServers[name]
				if !ok {
					t.Errorf("missing MCP server %q", name)
					continue
				}
				if server.Command == "" {
					t.Errorf("MCP server %q has empty command", name)
				}
				if len(server.Args) == 0 {
					t.Errorf("MCP server %q has no args", name)
				}
			}
		})
	}
}

func TestMCPGeneratorPlatformCommands(t *testing.T) {
	gen := NewMCPGenerator()

	t.Run("darwin_uses_bash", func(t *testing.T) {
		data, err := gen.GenerateMCP("darwin")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		var cfg MCPConfig
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		server := cfg.MCPServers["context7"]
		if server.Command != "/bin/bash" {
			t.Errorf("darwin command = %q, want %q", server.Command, "/bin/bash")
		}
		if len(server.Args) < 3 || server.Args[0] != "-l" || server.Args[1] != "-c" {
			t.Errorf("darwin args = %v, want [-l -c ...]", server.Args)
		}
	})

	t.Run("windows_uses_cmd", func(t *testing.T) {
		data, err := gen.GenerateMCP("windows")
		if err != nil {
			t.Fatalf("GenerateMCP error: %v", err)
		}

		var cfg MCPConfig
		trimmed := strings.TrimSpace(string(data))
		if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		server := cfg.MCPServers["context7"]
		if server.Command != "cmd.exe" {
			t.Errorf("windows command = %q, want %q", server.Command, "cmd.exe")
		}
		if len(server.Args) < 2 || server.Args[0] != "/c" {
			t.Errorf("windows args = %v, want [/c ...]", server.Args)
		}
	})
}

func TestMCPGeneratorStaggeredStartup(t *testing.T) {
	gen := NewMCPGenerator()

	data, err := gen.GenerateMCP("darwin")
	if err != nil {
		t.Fatalf("GenerateMCP error: %v", err)
	}

	var cfg MCPConfig
	trimmed := strings.TrimSpace(string(data))
	if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if cfg.StaggeredStartup == nil {
		t.Fatal("staggeredStartup is nil")
	}
	if !cfg.StaggeredStartup.Enabled {
		t.Error("staggeredStartup.enabled should be true")
	}
	if cfg.StaggeredStartup.DelayMs != 500 {
		t.Errorf("staggeredStartup.delayMs = %d, want 500", cfg.StaggeredStartup.DelayMs)
	}
	if cfg.StaggeredStartup.ConnectionTimeout != 15000 {
		t.Errorf("staggeredStartup.connectionTimeout = %d, want 15000", cfg.StaggeredStartup.ConnectionTimeout)
	}
}

func TestMCPGeneratorSchema(t *testing.T) {
	gen := NewMCPGenerator()

	data, err := gen.GenerateMCP("darwin")
	if err != nil {
		t.Fatalf("GenerateMCP error: %v", err)
	}

	var cfg MCPConfig
	trimmed := strings.TrimSpace(string(data))
	if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expected := "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json"
	if cfg.Schema != expected {
		t.Errorf("schema = %q, want %q", cfg.Schema, expected)
	}
}

func TestEventToSubcommand(t *testing.T) {
	tests := []struct {
		event    string
		expected string
	}{
		{"SessionStart", "session-start"},
		{"PreToolUse", "pre-tool"},
		{"PostToolUse", "post-tool"},
		{"SessionEnd", "session-end"},
		{"Stop", "stop"},
		{"PreCompact", "compact"},
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
