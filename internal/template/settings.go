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
	StatusLine       *StatusLine            `json:"statusLine,omitempty"`
	OutputStyle      string                 `json:"outputStyle,omitempty"`
	CleanupPeriodDays int                   `json:"cleanupPeriodDays,omitempty"`
	Env              map[string]string      `json:"env,omitempty"`
	Permissions      *Permissions          `json:"permissions,omitempty"`
}

// StatusLine represents the status line configuration.
type StatusLine struct {
	Type           string `json:"type"`
	Command        string `json:"command"`
	Padding        int    `json:"padding"`
	RefreshInterval int    `json:"refreshInterval"`
}

// Permissions represents tool permissions configuration.
type Permissions struct {
	DefaultMode string   `json:"defaultMode"`
	Allow       []string `json:"allow"`
	Ask         []string `json:"ask"`
	Deny        []string `json:"deny"`
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
// Note: SessionEnd is also handled by global hook, but included here for project-level consistency.
var hookEventDefs = []hookEventDef{
	{event: "SessionStart", matcher: "", timeout: 5},
	{event: "PreCompact", matcher: "", timeout: 5},
	{event: "SessionEnd", matcher: "", timeout: 5},
	{event: "PreToolUse", matcher: "Write|Edit|Bash", timeout: 5},
	{event: "PostToolUse", matcher: "Write|Edit", timeout: 60},
	{event: "Stop", matcher: "", timeout: 5},
}

// Generate builds Settings from config and serializes to JSON.
// For project-level settings, env is managed globally (not in template).
func (g *settingsGenerator) Generate(cfg *config.Config, platform string) ([]byte, error) {
	settings := Settings{
		Hooks:            buildHooks(platform),
		StatusLine:       buildStatusLine(),
		OutputStyle:      resolveOutputStyle(cfg),
		CleanupPeriodDays: 30,
		// Env: omitted - managed globally in ~/.claude/settings.json
		Permissions:      buildPermissions(),
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
// Uses $CLAUDE_PROJECT_DIR to reference project-local hooks in .claude/hooks/moai/.
// Project-local hooks apply only to the current project and can be version controlled.
func buildHookCommand(platform, event string) string {
	// Add "handle-" prefix to match deployed hook wrapper script names
	hookScriptName := "handle-" + eventToSubcommand(event) + ".sh"

	switch platform {
	case "windows":
		// Windows: use %CLAUDE_PROJECT_DIR% for project directory
		// Note: Windows cmd.exe doesn't natively support CLAUDE_PROJECT_DIR,
		// but Claude Code sets it as an environment variable
		return `cmd.exe /c "%CLAUDE_PROJECT_DIR%\.claude\hooks\moai\` + hookScriptName + `"`
	default:
		// darwin, linux, and other unix-like platforms
		// Use $CLAUDE_PROJECT_DIR with proper quoting for paths with spaces
		return `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/` + hookScriptName + `"`
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

// buildStatusLine constructs the status line configuration.
// Uses $CLAUDE_PROJECT_DIR for consistent path resolution.
func buildStatusLine() *StatusLine {
	// Use $CLAUDE_PROJECT_DIR for absolute path resolution
	// The .moai/status_line.sh script is deployed during initialization
	command := `"$CLAUDE_PROJECT_DIR/.moai/status_line.sh"`

	return &StatusLine{
		Type:            "command",
		Command:         command,
		Padding:         0,
		RefreshInterval: 300,
	}
}

// buildPermissions constructs the default permissions for Claude Code tools.
func buildPermissions() *Permissions {
	return &Permissions{
		DefaultMode: "default",
		Allow: []string{
			"AskUserQuestion",
			"BashOutput",
			"Edit",
			"Glob",
			"Grep",
			"KillShell",
			"MultiEdit",
			"NotebookEdit",
			"Read",
			"Skill",
			"Task",
			"TaskCreate",
			"TaskGet",
			"TaskList",
			"TaskUpdate",
			"TodoWrite",
			"WebFetch",
			"WebSearch",
			"Write",
			"mcp__context7__get-library-docs",
			"mcp__context7__resolve-library-id",
			"mcp__sequential-thinking__*",
			// Git operations
			"Bash(git status:*)",
			"Bash(git log:*)",
			"Bash(git diff:*)",
			"Bash(git show:*)",
			"Bash(git blame:*)",
			"Bash(git branch:*)",
			"Bash(git remote:*)",
			"Bash(git config:*)",
			"Bash(git tag:*)",
			"Bash(git add:*)",
			"Bash(git commit:*)",
			"Bash(git push:*)",
			"Bash(git pull:*)",
			"Bash(git fetch:*)",
			"Bash(git checkout:*)",
			"Bash(git switch:*)",
			"Bash(git stash:*)",
			"Bash(git merge:*)",
			"Bash(git revert:*)",
			// GitHub CLI
			"Bash(gh issue:*)",
			"Bash(gh pr:*)",
			"Bash(gh repo view:*)",
			// File operations
			"Bash(ls:*)",
			"Bash(cat:*)",
			"Bash(less:*)",
			"Bash(head:*)",
			"Bash(tail:*)",
			"Bash(grep:*)",
			"Bash(find:*)",
			"Bash(tree:*)",
			"Bash(pwd:*)",
			"Bash(wc:*)",
			"Bash(diff:*)",
			"Bash(comm:*)",
			"Bash(sort:*)",
			"Bash(uniq:*)",
			"Bash(cut:*)",
			"Bash(awk:*)",
			"Bash(sed:*)",
			"Bash(basename:*)",
			"Bash(dirname:*)",
			"Bash(realpath:*)",
			"Bash(readlink:*)",
			"Bash(cp:*)",
			"Bash(mv:*)",
			"Bash(mkdir:*)",
			"Bash(touch:*)",
			"Bash(rsync:*)",
			"Bash(curl:*)",
			"Bash(jq:*)",
			"Bash(yq:*)",
			// Development tools
			"Bash(python:*)",
			"Bash(python3:*)",
			"Bash(node*)",
			"Bash(npm*)",
			"Bash(npx*)",
			"Bash(bun*)",
			"Bash(pnpm*)",
			"Bash(yarn*)",
			"Bash(uv*)",
			"Bash(pip*)",
			"Bash(pip3:*)",
			"Bash(black:*)",
			"Bash(ruff:*)",
			"Bash(pylint:*)",
			"Bash(flake8:*)",
			"Bash(mypy:*)",
			"Bash(eslint:*)",
			"Bash(prettier:*)",
			"Bash(pytest:*)",
			"Bash(coverage:*)",
			"Bash(make:*)",
			"Bash(ast-grep:*)",
			"Bash(sg:*)",
			"Bash(rg:*)",
			"Bash(echo:*)",
			"Bash(printf:*)",
			"Bash(test:*)",
			"Bash(true:*)",
			"Bash(false:*)",
			"Bash(which:*)",
			"Bash(type:*)",
			"Bash(man:*)",
			"Bash(pydoc:*)",
			"Bash(lsof:*)",
			"Bash(time:*)",
			"Bash(xargs:*)",
			// MoAI commands
			"Bash(moai-adk:*)",
			"Bash(moai:*)",
		},
		Ask: []string{
			"Bash(rm:*)",
			"Bash(sudo:*)",
			"Bash(chmod:*)",
			"Bash(chown:*)",
			"Read(./.env)",
			"Read(./.env.*)",
		},
		Deny: []string{
			"Read(./secrets/**)",
			"Read(~/.ssh/**)",
			"Read(~/.aws/**)",
			"Read(~/.config/gcloud/**)",
			"Write(./secrets/**)",
			"Write(~/.ssh/**)",
			"Write(~/.aws/**)",
			"Write(~/.config/gcloud/**)",
			"Edit(./secrets/**)",
			"Edit(~/.ssh/**)",
			"Edit(~/.aws/**)",
			"Edit(~/.config/gcloud/**)",
			"Grep(./secrets/**)",
			"Grep(~/.ssh/**)",
			"Grep(~/.aws/**)",
			"Grep(~/.config/gcloud/**)",
			"Glob(./secrets/**)",
			"Glob(~/.ssh/**)",
			"Glob(~/.aws/**)",
			"Glob(~/.config/gcloud/**)",
			"Bash(rm -rf /:*)",
			"Bash(rm -rf /*:*)",
			"Bash(rm -rf ~:*)",
			"Bash(rm -rf ~/*:*)",
			"Bash(rm -rf C\\:/:*)",
			"Bash(rm -rf C\\:/*:*)",
			"Bash(del /S /Q C\\:/:*)",
			"Bash(rmdir /S /Q C\\:/:*)",
			"Bash(Remove-Item -Recurse -Force C\\:/:*)",
			"Bash(Clear-Disk:*)",
			"Bash(Format-Volume:*)",
			"Bash(git push --force:*)",
			"Bash(git push -f:*)",
			"Bash(git push --force-with-lease:*)",
			"Bash(git reset --hard:*)",
			"Bash(git clean -fd:*)",
			"Bash(git clean -fdx:*)",
			"Bash(git rebase -i:*)",
			"Bash(format:*)",
			"Bash(chmod -R 777:*)",
			"Bash(chmod 777:*)",
			"Bash(dd:*)",
			"Bash(mkfs:*)",
			"Bash(fdisk:*)",
			"Bash(reboot:*)",
			"Bash(shutdown:*)",
			"Bash(init:*)",
			"Bash(systemctl:*)",
			"Bash(kill -9:*)",
			"Bash(killall:*)",
			"Bash(DROP DATABASE:*)",
			"Bash(DROP TABLE:*)",
			"Bash(TRUNCATE:*)",
			"Bash(DELETE FROM:*)",
			"Bash(mongo:*)",
			"Bash(mongosh:*)",
			"Bash(redis-cli FLUSHALL:*)",
			"Bash(redis-cli FLUSHDB:*)",
			"Bash(psql -c DROP:*)",
			"Bash(mysql -e DROP:*)",
		},
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
