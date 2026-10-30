package cli

import "github.com/spf13/cobra"

// newAFCmd is the `moai af` command: the card flow's verbs over the local ledger (AC-AFL-017). This is the
// skeleton the RED tests run against; the verbs arrive with the GREEN commit.
func newAFCmd() *cobra.Command {
	return &cobra.Command{Use: "af", Short: "Agentic factory card flow over the local ledger"}
}
