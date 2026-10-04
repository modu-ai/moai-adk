package cli

// doctor_plugin_version.go — SPEC-PLUGIN-MARKETPLACE-001 M4 (REQ-020 to
// REQ-023): the "Plugin Version" doctor check. It compares the installed moai
// plugin version of each tool with the binary's version and only ever reports.
//
//	Claude: <CLAUDE_CONFIG_DIR | ~/.claude>/plugins/installed_plugins.json, key
//	        moai@moai-adk — one file read, no subprocess.
//	Codex:  `codex plugin list --json` under the resolved CODEX_HOME, started
//	        through the refusing runner of plugin_install.go, bounded by
//	        config.DefaultPluginVersionProbeTimeout, and only when codex
//	        resolves on PATH. The plugin cache is never read: after a
//	        registration is removed the cache directory outlives it.
//
// Every indeterminate case (not installed, a home or registry that cannot be
// read, an unknown shape, a probe that fails or times out, a development
// build) is OK or info, never warn or fail.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// pluginVersionCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const pluginVersionCheckName = "Plugin Version"

// pluginSeen is one tool's installed moai plugin version.
type pluginSeen struct {
	tool    pluginTool
	version string
	scope   string // Claude only
}

// checkPluginVersion is the registry entry point: the production wiring of
// checkPluginVersionAt. The Claude home follows CLAUDE_CONFIG_DIR, else
// ~/.claude; the Codex home comes from the existing seam, never from a path
// this check computes itself.
//
// @MX:NOTE: [AUTO] The Codex home resolves unconditionally, through
// resolveCodexHomeDir (the codexUserHomeDir seam that TestMain redirects), so
// no test reads a real Codex home; the probe runs through pluginRunner, whose
// default refuses under a test binary.
func checkPluginVersion(verbose bool) DiagnosticCheck {
	claudeHome := strings.TrimSpace(os.Getenv(config.EnvClaudeConfigDir))
	if claudeHome == "" {
		if home, err := userHomeDirFn(); err == nil {
			claudeHome = filepath.Join(home, ".claude")
		}
	}
	codexHome, _ := resolveCodexHomeDir()
	return checkPluginVersionAt(claudeHome, codexHome, pluginRunner, codexWiringLookPath, version.GetVersion(), verbose)
}

// checkPluginVersionAt is checkPluginVersion with the homes, the probe runner,
// the PATH lookup and the binary version injected, so no test mutates package
// state to exercise a comparison.
func checkPluginVersionAt(claudeHome, codexHome string, run pluginCommandRunner, lookPath func(string) (string, error), binaryVersion string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: pluginVersionCheckName, Status: uikit.CheckInfo}

	var seen []pluginSeen
	var notes []string
	unreadable := false

	if v, ok, note := readClaudePluginVersion(claudeHome); ok {
		seen = append(seen, v)
	} else if note != "" {
		notes = append(notes, note)
		unreadable = unreadable || strings.Contains(note, "unreadable")
	}
	if v, ok, note := probeCodexPluginVersion(codexHome, run, lookPath); ok {
		seen = append(seen, v)
	} else if note != "" {
		notes = append(notes, note)
		unreadable = unreadable || strings.Contains(note, "unreadable")
	}

	bin := normalizePluginVersion(binaryVersion)
	switch {
	case len(seen) == 0 && unreadable:
		check.Message = "installed moai plugin version could not be read"
	case len(seen) == 0:
		check.Message = "moai plugin not installed (run: moai plugin install)"
	case version.IsDevBuild(binaryVersion):
		check.Message = fmt.Sprintf("development build (%s); installed plugin not compared", describePluginSeen(seen))
	default:
		var mismatches []string
		for _, s := range seen {
			if normalizePluginVersion(s.version) != bin {
				mismatches = append(mismatches, fmt.Sprintf("%s plugin %s differs from binary %s (run: %s)",
					pluginSpec(s.tool).label, normalizePluginVersion(s.version), bin, pluginUpdateRemedy(s.tool)))
			}
		}
		if len(mismatches) > 0 {
			check.Status = uikit.CheckWarn
			check.Message = strings.Join(mismatches, "; ")
		} else {
			check.Status = uikit.CheckOK
			check.Message = fmt.Sprintf("moai plugin matches the binary %s (%s)", bin, describePluginSeen(seen))
		}
	}

	if verbose {
		// Not the legacy "Plugin Deployment" check, which reads a system.yaml marker.
		detail := "moai plugin installed vs binary version; unrelated to the legacy Plugin Deployment marker check"
		if len(notes) > 0 {
			detail += "; " + strings.Join(notes, "; ")
		}
		check.Detail = detail
	}
	return check
}

// normalizePluginVersion drops whitespace and one leading "v": the binary
// prints v3.2.0-rc.26 and the manifests carry 3.2.0-rc.26.
func normalizePluginVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func describePluginSeen(seen []pluginSeen) string {
	parts := make([]string, 0, len(seen))
	for _, s := range seen {
		p := fmt.Sprintf("%s %s", pluginSpec(s.tool).label, normalizePluginVersion(s.version))
		if s.scope != "" {
			p += " (" + s.scope + " scope)"
		}
		parts = append(parts, p)
	}
	return strings.Join(parts, ", ")
}

// pluginUpdateRemedy is the manual command that refreshes the plugin, read
// from each tool's --help and not run by this check.
func pluginUpdateRemedy(tool pluginTool) string {
	if tool == pluginToolCodex {
		return "codex plugin marketplace upgrade moai-adk, then codex plugin add " + pluginRef
	}
	return "claude plugin update " + pluginRef
}

// readClaudePluginVersion reads the installed version of moai@moai-adk from the
// Claude registry file: the user-scope entry first, else the first entry that
// carries a version. ok is false when nothing could be read; note says why when
// the cause is worth naming.
func readClaudePluginVersion(claudeHome string) (pluginSeen, bool, string) {
	if claudeHome == "" {
		return pluginSeen{}, false, "Claude home not resolved"
	}
	raw, err := os.ReadFile(filepath.Join(claudeHome, "plugins", "installed_plugins.json"))
	if err != nil {
		return pluginSeen{}, false, "no Claude plugin registry"
	}
	var reg struct {
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if err := json.Unmarshal(raw, &reg); err != nil {
		return pluginSeen{}, false, "Claude plugin registry unreadable"
	}
	entryRaw, ok := reg.Plugins[pluginRef]
	if !ok {
		return pluginSeen{}, false, "" // a registry without the plugin: not installed
	}
	var entries []struct {
		Scope   string `json:"scope"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(entryRaw, &entries); err != nil {
		return pluginSeen{}, false, "Claude plugin registry unreadable (unknown shape)"
	}
	var first *pluginSeen
	for _, e := range entries {
		if strings.TrimSpace(e.Version) == "" {
			continue
		}
		s := pluginSeen{tool: pluginToolClaude, version: e.Version, scope: e.Scope}
		if e.Scope == "user" {
			return s, true, ""
		}
		if first == nil {
			first = &s
		}
	}
	if first != nil {
		return *first, true, ""
	}
	return pluginSeen{}, false, "Claude plugin registry unreadable (no version)"
}

// probeCodexPluginVersion asks `codex plugin list --json` for the installed
// version of moai@moai-adk. It starts nothing unless codex resolves on PATH, and
// hands the child the parent's environment with CODEX_HOME set to the resolved
// home.
func probeCodexPluginVersion(codexHome string, run pluginCommandRunner, lookPath func(string) (string, error)) (pluginSeen, bool, string) {
	if run == nil || lookPath == nil {
		return pluginSeen{}, false, ""
	}
	bin, err := lookPath("codex")
	if err != nil {
		return pluginSeen{}, false, "codex not on PATH"
	}
	if codexHome == "" {
		return pluginSeen{}, false, "Codex home not resolved"
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultPluginVersionProbeTimeout)
	defer cancel()
	out, err := run.Run(ctx, bin, []string{"plugin", "list", "--json"}, envWithCodexHome(os.Environ(), codexHome))
	if err != nil {
		return pluginSeen{}, false, "codex plugin list failed or timed out (unreadable)"
	}
	var list struct {
		Installed []struct {
			PluginID string `json:"pluginId"`
			Version  string `json:"version"`
		} `json:"installed"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return pluginSeen{}, false, "codex plugin list output unreadable"
	}
	for _, p := range list.Installed {
		if p.PluginID != pluginRef {
			continue
		}
		if strings.TrimSpace(p.Version) == "" {
			return pluginSeen{}, false, "codex plugin list entry unreadable (no version)"
		}
		return pluginSeen{tool: pluginToolCodex, version: p.Version}, true, ""
	}
	return pluginSeen{}, false, ""
}

// envWithCodexHome returns env with CODEX_HOME set to home and every other
// variable untouched.
func envWithCodexHome(env []string, home string) []string {
	prefix := codexHomeEnvVar + "="
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return append(out, prefix+home)
}
