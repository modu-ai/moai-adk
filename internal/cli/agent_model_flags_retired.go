package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deprecatedAgentModelFlagUsage is the help text of every retired per-agent
// model flag (init --profile/--model-policy/--high/--medium-alias/--low,
// update --profile).
const deprecatedAgentModelFlagUsage = "Deprecated, no effect: subagents inherit the main session's model and effort (set the main-session policy with `moai profile setup`)"

// warnDeprecatedAgentModelFlags prints one deprecation warning to the
// command's stderr for each named flag the user set (a non-empty string or a
// true bool). The flags are accepted with any value so scripts that pass them
// keep working; they write nothing (SPEC-AGENT-MODEL-INHERIT-001 D10/D13).
func warnDeprecatedAgentModelFlags(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			continue
		}
		set := f.Value.String()
		if set == "" || set == "false" {
			continue
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"Warning: --%s is deprecated and has no effect: subagents now inherit the main session's model and effort. "+
				"Set the main-session model policy with `moai profile setup`.\n", name)
	}
}
