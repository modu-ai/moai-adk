package cli

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// newIntegrationRemeasureCmd — `moai integration remeasure -- <command>`
// (card t1479, REQ-MWQ-014/015/016): run the re-measure the merge gate
// requires, BEFORE joining the window queue, and store its record keyed by
// the candidate tree. The verb body is factory.RunRemeasure; the verb parses
// the command, renders the verdict, and maps the clean-check refusal to a
// non-zero exit.
func newIntegrationRemeasureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remeasure -- <command>",
		Short: "Run the re-measure and store its record keyed by the candidate tree",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.TrimSpace(strings.Join(args, " "))
			if command == "" {
				return fmt.Errorf("integration remeasure: pass the command after -- (a re-measure without a command measures nothing)")
			}
			root := integrationLockRoot()
			rec, err := factory.RunRemeasure(root, resolveProjectDir(), configuredIntegrationBranch(), command)
			if err != nil {
				return err
			}
			verdict := "valid"
			if vErr := factory.ValidateRemeasureRecord(rec); vErr != nil {
				verdict = "INVALID: " + vErr.Error()
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "re-measure recorded for tree %s (base %s): %s\n", rec.Tree[:12], rec.Base[:12], verdict)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  command: %s (exit %d)\n", rec.Command, rec.ExitCode)
			return nil
		},
	}
	return cmd
}
