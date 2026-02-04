package template

import (
	"encoding/json"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Settings represents the Claude Code settings.json structure.
// Generated exclusively via json.MarshalIndent (ADR-011).
type Settings struct {
	Hooks            map[string][]HookGroup `json:"hooks,omitempty"`
	OutputStyle      string                 `json:"outputStyle,omitempty"`
	CleanupPeriodDays int                   `json:"cleanupPeriodDays,omitempty"`
	Env              map[string]string      `json:"env,omitempty"`
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
	Timeout int    `json:"timeout,omitempty"`
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

// hookEventDef defines a hook event with its configuration.
type hookEventDef struct {
	event   string // Claude Code event name (PascalCase)
	matcher string // Tool matcher pattern (empty for all)
	timeout int    // Timeout in seconds
}

// hookEvents defines the required hook events per REQ-T-023.
var hookEventDefs = []hookEventDef{
	{event: "SessionStart", matcher: "", timeout: 5},
	{event: "PreCompact", matcher: "", timeout: 5},
	{event: "SessionEnd", matcher: "", timeout: 10},
	{event: "PreToolUse", matcher: "Write|Edit|Bash", timeout: 5},
	{event: "PostToolUse", matcher: "Write|Edit", timeout: 60},
	{event: "Stop", matcher: "", timeout: 5},
}

// Generate builds Settings from config and serializes to JSON.
func (g *settingsGenerator) Generate(cfg *config.Config, platform string) ([]byte, error) {
	settings := Settings{
		Hooks:            buildHooks(platform),
		OutputStyle:      resolveOutputStyle(cfg),
		CleanupPeriodDays: 30,
		Env:              buildEnv(),
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
	hooks := make(map[string][]HookGroup, len(hookEventDefs))

	for _, def := range hookEventDefs {
		cmd := buildHookCommand(platform, def.event)
		hooks[def.event] = []HookGroup{
			{
				Matcher: def.matcher,
				Hooks: []HookEntry{
					{
						Type:    "command",
						Command: cmd,
						Timeout: def.timeout,
					},
				},
			},
		}
	}

	return hooks
}

// buildHookCommand returns the platform-appropriate hook command string.
// Uses quoted $HOME path per Claude Code official documentation.
func buildHookCommand(platform, event string) string {
	hookSubcommand := eventToSubcommand(event)

	switch platform {
	case "windows":
		// Windows: use %USERPROFILE% for home directory
		return `cmd.exe /c "%USERPROFILE%\go\bin\moai" hook ` + hookSubcommand
	default:
		// darwin, linux, and other unix-like platforms
		// Use quoted $HOME path per official Claude Code documentation
		return `"$HOME/go/bin/moai" hook ` + hookSubcommand
	}
}

// eventToSubcommand converts a PascalCase event name to a kebab-case subcommand.
// Maps to actual moai CLI hook subcommands (not Claude Code event names).
func eventToSubcommand(event string) string {
	switch event {
	case "SessionStart":
		return "session-start"
	case "PreToolUse":
		return "pre-tool"
	case "PostToolUse":
		return "post-tool"
	case "SessionEnd":
		return "session-end"
	case "Stop":
		return "stop"
	case "PreCompact":
		return "compact"
	default:
		return event
	}
}

// buildEnv constructs the default environment variables for settings.json.
func buildEnv() map[string]string {
	return map[string]string{
		"PATH": "$HOME/go/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:$HOME/.local/bin:$HOME/.cargo/bin:/opt/homebrew/bin",
	}
}

// defaultOutputStyle is the default output style for new MoAI projects.
const defaultOutputStyle = "MoAI"

// resolveOutputStyle determines the output style from configuration.
func resolveOutputStyle(cfg *config.Config) string {
	// Default to MoAI output style for all new projects.
	// Future: allow override via config (e.g., r2d2, yoda).
	return defaultOutputStyle
}

// --- MCP Configuration Generator ---

// MCPConfig represents the .mcp.json structure for Claude Code MCP servers.
type MCPConfig struct {
	Schema           string                `json:"$schema"`
	MCPServers       map[string]*MCPServer `json:"mcpServers"`
	StaggeredStartup *MCPStaggeredStartup  `json:"staggeredStartup,omitempty"`
}

// MCPServer represents a single MCP server entry.
type MCPServer struct {
	Comment string   `json:"$comment,omitempty"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// MCPStaggeredStartup configures staggered MCP server startup.
type MCPStaggeredStartup struct {
	Enabled           bool `json:"enabled"`
	DelayMs           int  `json:"delayMs"`
	ConnectionTimeout int  `json:"connectionTimeout"`
}

// MCPGenerator produces .mcp.json content from platform information.
type MCPGenerator interface {
	// GenerateMCP creates a valid JSON byte slice for .mcp.json.
	// Uses Go struct serialization only (ADR-011: no string concatenation).
	GenerateMCP(platform string) ([]byte, error)
}

// mcpGenerator is the concrete implementation of MCPGenerator.
type mcpGenerator struct{}

// NewMCPGenerator creates a new MCPGenerator.
func NewMCPGenerator() MCPGenerator {
	return &mcpGenerator{}
}

// mcpServerDef defines an MCP server package for generation.
type mcpServerDef struct {
	name    string // server name key
	comment string // human-readable comment
	pkg     string // npx package specifier
}

// defaultMCPServers lists the MCP servers included in new projects.
var defaultMCPServers = []mcpServerDef{
	{
		name:    "context7",
		comment: "Up-to-date documentation and code examples via Context7",
		pkg:     "@upstash/context7-mcp@latest",
	},
	{
		name:    "sequential-thinking",
		comment: "Step-by-step reasoning for complex problems",
		pkg:     "@modelcontextprotocol/server-sequential-thinking",
	},
}

// GenerateMCP builds MCPConfig and serializes to JSON.
func (g *mcpGenerator) GenerateMCP(platform string) ([]byte, error) {
	cfg := MCPConfig{
		Schema:     "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json",
		MCPServers: buildMCPServers(platform),
		StaggeredStartup: &MCPStaggeredStartup{
			Enabled:           true,
			DelayMs:           500,
			ConnectionTimeout: 15000,
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("mcp generate marshal: %w", err)
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("%w: generated .mcp.json is invalid", ErrInvalidJSON)
	}

	data = append(data, '\n')
	return data, nil
}

// buildMCPServers constructs the mcpServers map for all default servers.
func buildMCPServers(platform string) map[string]*MCPServer {
	servers := make(map[string]*MCPServer, len(defaultMCPServers))

	for _, def := range defaultMCPServers {
		cmd, args := buildMCPCommand(platform, def.pkg)
		servers[def.name] = &MCPServer{
			Comment: def.comment,
			Command: cmd,
			Args:    args,
		}
	}

	return servers
}

// buildMCPCommand returns the platform-appropriate command and args for an MCP server.
func buildMCPCommand(platform, pkg string) (string, []string) {
	switch platform {
	case "windows":
		return "cmd.exe", []string{"/c", "npx -y " + pkg}
	default:
		// darwin, linux: use login shell to ensure PATH includes npm/node
		return "/bin/bash", []string{"-l", "-c", "exec npx -y " + pkg}
	}
}
