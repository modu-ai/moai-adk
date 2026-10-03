package cli

// doctor_plugin_version.go — SPEC-PLUGIN-MARKETPLACE-001 M4 (REQ-020 to
// REQ-023): the "Plugin Version" doctor check. RED stubs; the behavior lands
// in the GREEN commit.

import (
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// pluginVersionCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const pluginVersionCheckName = "Plugin Version"

// checkPluginVersion is the registry entry point.
func checkPluginVersion(verbose bool) DiagnosticCheck {
	return DiagnosticCheck{Name: pluginVersionCheckName, Status: uikit.CheckOK, Message: "stub"}
}

// checkPluginVersionAt is checkPluginVersion with the homes, the probe runner,
// the PATH lookup and the binary version injected.
func checkPluginVersionAt(claudeHome, codexHome string, run pluginCommandRunner, lookPath func(string) (string, error), binaryVersion string, verbose bool) DiagnosticCheck {
	return DiagnosticCheck{Name: pluginVersionCheckName, Status: uikit.CheckOK, Message: "stub"}
}
