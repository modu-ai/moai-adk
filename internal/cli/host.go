package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenthost"
)

func newHostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "host",
		Short:   "Inspect coding-agent host compatibility",
		GroupID: "tools",
	}
	cmd.AddCommand(newHostMatrixCmd())
	return cmd
}

func newHostMatrixCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "matrix [host]",
		Short: "Show MoAI hook compatibility for Claude, Codex, and OpenCode",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var matrices []agenthost.Matrix
			if len(args) == 0 {
				matrices = agenthost.AllMatrices()
			} else {
				host, err := agenthost.ParseHost(args[0])
				if err != nil {
					return err
				}
				matrix, err := agenthost.MatrixFor(host)
				if err != nil {
					return err
				}
				matrices = []agenthost.Matrix{matrix}
			}

			if outputJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(matrices)
			}
			writeHostMatrix(cmd, matrices)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Emit JSON")
	return cmd
}

func writeHostMatrix(cmd *cobra.Command, matrices []agenthost.Matrix) {
	out := cmd.OutOrStdout()
	for i, matrix := range matrices {
		if i > 0 {
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprintf(out, "Host: %s\n", matrix.Host)
		_, _ = fmt.Fprintf(out, "Source: %s\n", matrix.Source)
		_, _ = fmt.Fprintln(out, "Event\tSupport\tHost event\tDegradation")
		for _, mapping := range matrix.Mappings {
			degradation := mapping.Degradation
			if degradation == "" {
				degradation = "-"
			}
			_, _ = fmt.Fprintf(out, "%s\t%s\t%s\t%s\n",
				mapping.Event,
				mapping.Support,
				mapping.HostEvent,
				degradation,
			)
		}
	}
}
