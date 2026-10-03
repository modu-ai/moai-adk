package cli

// plugin_install_cmd.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, the `moai plugin
// install` verb (RED stub: compiles, performs nothing, is not registered).

import "github.com/spf13/cobra"

func newPluginCmd() *cobra.Command {
	return &cobra.Command{Use: "plugin"}
}

func newPluginInstallCmd() *cobra.Command {
	return &cobra.Command{Use: "install"}
}
