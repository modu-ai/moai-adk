// Package pluginemit generates the moai marketplace and the manifests of the
// derived moai core plugin (SPEC-PLUGIN-MARKETPLACE-001).
//
// The generator reads the embedded template tree and the version SSOT
// (pkg/version.Version) and writes four manifests: the Claude and Codex
// marketplace manifests at the repository root, and the Claude and Codex
// plugin manifests under the plugin root. Every version it writes is the
// SSOT value with its leading "v" stripped, and the Codex manifest's MCP entry
// is copied from the template .mcp.json, so neither can drift from its source.
//
// Output is deterministic: fixed field order, two-space JSON indentation, LF
// newlines, a trailing newline and no timestamp. The package never writes
// files itself; the golden test writes them behind PLUGIN_EMIT_UPDATE, and the
// committed files are never edited by hand.
package pluginemit

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// EnvUpdate is the environment switch that flips the golden tests into
// regeneration mode (the maintainer path); unset or empty, they compare.
const EnvUpdate = "PLUGIN_EMIT_UPDATE"

// Plugin and marketplace identity, and the committed locations of the four
// manifests (default pending OD-4: a committed generated tree under
// plugins/moai, entry source a relative path, no ref pin).
const (
	// PluginName is the plugin identifier. It is also the MCP server key and
	// the one generator literal allowed to equal a template component name.
	PluginName = "moai"
	// MarketplaceName names both marketplaces.
	MarketplaceName = "moai-adk"
	// PluginRoot is the repository-root-relative directory of the plugin.
	PluginRoot = "plugins/" + PluginName

	// ClaudeMarketplacePath and CodexMarketplacePath are the marketplace
	// manifests; ClaudePluginPath and CodexPluginPath the plugin manifests.
	ClaudeMarketplacePath = ".claude-plugin/marketplace.json"
	CodexMarketplacePath  = ".agents/plugins/marketplace.json"
	ClaudePluginPath      = PluginRoot + "/.claude-plugin/plugin.json"
	CodexPluginPath       = PluginRoot + "/.codex-plugin/plugin.json"
)

// Options selects the inputs of one emission.
type Options struct {
	// Version is the version SSOT value, with or without a leading "v".
	Version string
	// MCPSource is the fs-relative path of the template .mcp.json.
	MCPSource string
}

// DefaultOptions returns the options for the real template tree. The version
// is read from the SSOT at call time, not cached, so a test that sets
// version.Version sees its value.
func DefaultOptions() Options {
	return Options{
		Version:   version.Version,
		MCPSource: ".mcp.json",
	}
}

// Publication is the deterministic output of one emission.
type Publication struct {
	// Files maps each repository-root-relative path (forward slashes) to its
	// bytes.
	Files map[string][]byte
}

// Emit produces the four manifests from the template tree and the version.
// On any error it returns (nil, err): no partial set.
//
// @MX:NOTE: sole entry point of the plugin generator; the golden tests and `make plugin-emit-check` judge its output
func Emit(fsys fs.FS, opts Options) (*Publication, error) {
	ver := strings.TrimPrefix(opts.Version, "v")
	if ver == "" {
		return nil, fmt.Errorf("pluginemit: empty version")
	}
	mcp, err := DeriveMCPEntry(fsys, opts.MCPSource)
	if err != nil {
		return nil, err
	}

	manifests := map[string]any{
		ClaudeMarketplacePath: claudeMarketplace(ver),
		CodexMarketplacePath:  codexMarketplace(),
		ClaudePluginPath:      claudePlugin(ver),
		CodexPluginPath:       codexPlugin(ver, mcp),
	}
	pub := &Publication{Files: make(map[string][]byte, len(manifests))}
	for path, m := range manifests {
		data, err := marshal(m)
		if err != nil {
			return nil, fmt.Errorf("pluginemit: marshal %s: %w", path, err)
		}
		pub.Files[path] = data
	}
	return pub, nil
}
