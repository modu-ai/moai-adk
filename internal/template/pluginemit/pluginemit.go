// Package pluginemit generates the moai marketplace and the derived moai core
// plugin (SPEC-PLUGIN-MARKETPLACE-001).
//
// The generator reads the embedded template tree and the version SSOT
// (pkg/version.Version) and produces the Claude and Codex marketplace manifests
// at the repository root, the Claude and Codex plugin manifests under the
// plugin root, and the plugin payload (the core-tier skills, the commands
// laid out flat, and one .mcp.json). Every version it writes is the SSOT value
// with its leading "v" stripped, and the MCP entry is copied from the template
// .mcp.json, so neither can drift from its source. No component name is held in
// the generator: the payload follows the template tree.
//
// Output is deterministic: fixed field order, two-space JSON indentation, LF
// newlines, a trailing newline and no timestamp. Emit never writes files; Write
// (reached through `make plugin-emit`) is the one regeneration path, Drift the
// read-only check, and the committed tree is never edited by hand.
package pluginemit

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/modu-ai/moai-adk/internal/template"
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
	// MCPSource is the path of the template .mcp.json inside the tier view.
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
	// Modes maps every path of Files to the file mode it is committed at.
	Modes map[string]fs.FileMode
}

// Emit produces the four manifests and the payload from the template tree and
// the version. raw is the raw embed layout: catalog.yaml at its root and the
// template tree under templates/; the payload is read through the same tier view
// `moai init` deploys (the core tier, default pending OD-8). On any error it
// returns (nil, err): no partial set.
//
// @MX:NOTE: sole entry point of the plugin generator; the golden tests and `make plugin-emit-check` judge its output
func Emit(raw fs.FS, opts Options) (*Publication, error) {
	ver := strings.TrimPrefix(opts.Version, "v")
	if ver == "" {
		return nil, fmt.Errorf("pluginemit: empty version")
	}
	cat, err := template.LoadCatalog(raw)
	if err != nil {
		return nil, fmt.Errorf("pluginemit: %w", err)
	}
	view, err := template.SlimFS(raw, cat)
	if err != nil {
		return nil, fmt.Errorf("pluginemit: tier view: %w", err)
	}
	mcp, err := DeriveMCPEntry(view, opts.MCPSource)
	if err != nil {
		return nil, err
	}

	manifests := map[string]any{
		ClaudeMarketplacePath: claudeMarketplace(ver),
		CodexMarketplacePath:  codexMarketplace(),
		ClaudePluginPath:      claudePlugin(ver),
		CodexPluginPath:       codexPlugin(ver, mcp),
	}
	pub := &Publication{
		Files: make(map[string][]byte, len(manifests)),
		Modes: make(map[string]fs.FileMode, len(manifests)),
	}
	for path, m := range manifests {
		data, err := marshal(m)
		if err != nil {
			return nil, fmt.Errorf("pluginemit: marshal %s: %w", path, err)
		}
		pub.Files[path] = data
	}
	files, err := payload(view, mcp)
	if err != nil {
		return nil, err
	}
	for path, data := range files {
		pub.Files[path] = data
	}
	for path := range pub.Files {
		pub.Modes[path] = modeFor(path)
	}
	return pub, nil
}

// modeFor is the deployer's mode rule (internal/template/deployer.go): 0755 for
// a .sh file, 0644 for every other file. The embedded template tree carries no
// mode to copy (embed.FS reports 0444), so the suffix is the only rule it offers.
func modeFor(path string) fs.FileMode {
	if strings.HasSuffix(path, ".sh") {
		return 0o755
	}
	return 0o644
}
