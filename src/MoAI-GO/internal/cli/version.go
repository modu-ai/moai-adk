package cli

import (
	"encoding/json"
	"fmt"
	"github.com/anthropics/moai-adk-go/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version of MoAI-ADK currently installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.GetVersionInfo()

			if jsonOutput {
				data, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal version info: %w", err)
				}
				cmd.Println(string(data))
			} else {
				cmd.Printf("moai version %s\n", info["version"])
				cmd.Printf("commit: %s\n", info["commit"])
				cmd.Printf("built at: %s\n", info["date"])
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")

	return cmd
}
