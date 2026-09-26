// Package cli — doctor_mcp_provider.go
//
// `moai doctor` check: is the same MCP provider active twice, once as a local
// server and once as a claude.ai connector?
//
// Claude Code deduplicates MCP servers only by identical name. A local server
// named `foo` (project .mcp.json, global ~/.claude/.mcp.json, or user scope)
// and a claude.ai connector named `claude.ai Foo` have different names and
// different tool prefixes, so both tool sets load and every tool listing is
// paid twice in context. "MCP Scope Duplicates" cannot see this because it
// compares identical names across the two .mcp.json files only.
//
// The connector inventory lives in the Claude Code state file (.claude.json),
// an undocumented internal. The check is therefore advisory and fail-open:
// anything missing or unparseable reports OK, never FAIL.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

// mcpProviderDuplicatesCheckName is the doctor check identifier (also the
// value accepted by `moai doctor --check`).
const mcpProviderDuplicatesCheckName = "MCP Provider Duplicates"

// claudeAIConnectorPrefix is the prefix Claude Code puts on claude.ai
// connector names in the state file and in disabledMcpServers.
const claudeAIConnectorPrefix = "claude.ai "

// mcpProviderStatePath resolves the Claude Code state file. A package-level
// seam so tests point it at t.TempDir() instead of the real home.
var mcpProviderStatePath = func() string {
	home, _ := os.UserHomeDir()
	return resolveClaudeStatePath(os.Getenv, home)
}

// mcpProviderGlobalMCPPath resolves the global ~/.claude/.mcp.json (same
// source checkMCPScopeDuplicates reads). A seam for the same reason.
var mcpProviderGlobalMCPPath = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", ".mcp.json")
}

// resolveClaudeStatePath returns $CLAUDE_CONFIG_DIR/.claude.json when that
// variable is set, else <home>/.claude.json.
func resolveClaudeStatePath(getenv func(string) string, home string) string {
	if dir := strings.TrimSpace(getenv(config.EnvClaudeConfigDir)); dir != "" {
		return filepath.Join(dir, ".claude.json")
	}
	return filepath.Join(home, ".claude.json")
}

// claudeStateFile is the subset of the Claude Code state file this check reads.
type claudeStateFile struct {
	MCPServers               map[string]json.RawMessage `json:"mcpServers"`
	ClaudeAIMCPEverConnected []string                   `json:"claudeAiMcpEverConnected"`
	Projects                 map[string]claudeProjectEntry `json:"projects"`
}

// claudeProjectEntry is the per-project slice of the state file this check
// reads. DisabledMcpjsonServers lists .mcp.json servers the user rejected in
// the approval prompt — Claude Code does not load those.
type claudeProjectEntry struct {
	DisabledMCPServers     []string `json:"disabledMcpServers"`
	DisabledMcpjsonServers []string `json:"disabledMcpjsonServers"`
}

// normalizeMCPProviderKey lowercases and keeps only [a-z0-9], so `context7`,
// `Context7` and `context-7` compare equal.
func normalizeMCPProviderKey(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// @MX:NOTE: [AUTO] Matches local MCP servers against claude.ai connectors by normalized name; the state file is an undocumented Claude Code internal, so every read failure degrades to OK.
// checkMCPProviderDuplicates warns when a local MCP server and a claude.ai
// connector resolve to the same normalized name and neither side is disabled
// or approval-rejected for this project.
func checkMCPProviderDuplicates(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: mcpProviderDuplicatesCheckName, Status: uikit.CheckOK}

	data, err := os.ReadFile(mcpProviderStatePath())
	if err != nil {
		check.Message = "no Claude Code state file; not checked"
		return check
	}
	var state claudeStateFile
	if err := json.Unmarshal(data, &state); err != nil {
		check.Message = "Claude Code state file unreadable; not checked"
		return check
	}
	if len(state.ClaudeAIMCPEverConnected) == 0 {
		check.Message = "no claude.ai connectors recorded"
		return check
	}

	// Local servers with the scope label shown in the message. Project scope
	// wins the label when a name appears in several scopes.
	local := map[string]string{}
	add := func(names map[string]struct{}, scope string) {
		for name := range names {
			if _, seen := local[name]; !seen {
				local[name] = scope
			}
		}
	}
	add(parseMCPJSON(filepath.Join(projectRoot, ".mcp.json")), ".mcp.json")
	add(parseMCPJSON(mcpProviderGlobalMCPPath()), "~/.claude/.mcp.json")
	userScope := make(map[string]struct{}, len(state.MCPServers))
	for name := range state.MCPServers {
		userScope[name] = struct{}{}
	}
	add(userScope, "user scope")

	disabled := map[string]bool{}
	entry, _ := projectStateEntry(state, projectRoot)
	for _, name := range entry.DisabledMCPServers {
		disabled[name] = true
	}
	// A .mcp.json server rejected in the approval prompt is not loaded, so it
	// cannot duplicate a connector; treat rejection like a disabled server.
	for _, name := range entry.DisabledMcpjsonServers {
		disabled[name] = true
	}

	localByKey := map[string][]string{}
	for name := range local {
		key := normalizeMCPProviderKey(name)
		if key != "" {
			localByKey[key] = append(localByKey[key], name)
		}
	}

	var pairs []string
	for _, connector := range state.ClaudeAIMCPEverConnected {
		bare := strings.TrimPrefix(connector, claudeAIConnectorPrefix)
		if disabled[connector] || disabled[claudeAIConnectorPrefix+bare] {
			continue
		}
		names := localByKey[normalizeMCPProviderKey(bare)]
		sort.Strings(names)
		for _, name := range names {
			if disabled[name] {
				continue
			}
			pairs = append(pairs, fmt.Sprintf("%s (%s) + %s%s", name, local[name], claudeAIConnectorPrefix, bare))
		}
	}

	if len(pairs) == 0 {
		check.Message = fmt.Sprintf("%d local MCP server(s), %d claude.ai connector(s) — no overlap", len(local), len(state.ClaudeAIMCPEverConnected))
		return check
	}

	sort.Strings(pairs)
	check.Status = uikit.CheckWarn
	check.Message = "same MCP provider active twice: " + strings.Join(pairs, ", ")
	check.Detail = "Both sides load the same tools, so each tool listing is paid twice in context. " +
		"Disable one side in Claude Code via /mcp (recorded in disabledMcpServers), or remove the entry from .mcp.json. " +
		"Note: claudeAiMcpEverConnected records connectors ever connected, so a connector removed from the account may still be reported."
	return check
}

// projectStateEntry returns the state-file entry recorded for projectRoot.
// The state file keys projects by absolute path; the symlink-resolved form is
// tried as a fallback. A missing entry yields false.
func projectStateEntry(state claudeStateFile, projectRoot string) (claudeProjectEntry, bool) {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}
	if p, ok := state.Projects[abs]; ok {
		return p, true
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		if p, ok := state.Projects[real]; ok {
			return p, true
		}
	}
	return claudeProjectEntry{}, false
}
