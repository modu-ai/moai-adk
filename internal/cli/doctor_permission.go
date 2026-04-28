package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/permission"
)

var permissionCmd = &cobra.Command{
	Use:     "permission",
	Short:   "Diagnose permission resolution",
	Long:    "Diagnose permission stack resolution with trace output and rule provenance.",
	RunE:    runDoctorPermission,
}

func init() {
	doctorCmd.AddCommand(permissionCmd)

	permissionCmd.Flags().String("tool", "", "Tool name to resolve (e.g., Bash, Read, Write)")
	permissionCmd.Flags().String("input", "", "Tool input arguments")
	permissionCmd.Flags().Bool("trace", false, "Show full resolution trace as JSON")
	permissionCmd.Flags().Bool("dry-run", false, "Show resolution without executing")
}

// runDoctorPermission executes the permission diagnostic workflow.
//
// Usage:
//   moai doctor permission --tool Bash --input "go test ./..." --trace
//   moai doctor permission --tool Write --input "/tmp/test.txt" --dry-run
func runDoctorPermission(cmd *cobra.Command, _ []string) error {
	tool := getStringFlag(cmd, "tool")
	input := getStringFlag(cmd, "input")
	showTrace := getBoolFlag(cmd, "trace")
	dryRun := getBoolFlag(cmd, "dry-run")

	if tool == "" {
		return fmt.Errorf("--tool is required")
	}

	out := cmd.OutOrStdout()

	// Create resolver with default context
	resolver := permission.NewPermissionResolver()
	ctx := permission.ResolveContext{
		Mode:            permission.ModeDefault,
		IsFork:          false,
		ParentAvailable: false,
		ForkDepth:       0,
		IsInteractive:   true,
		StrictMode:      false,
		RulesByTier:     make(map[config.Source][]permission.PermissionRule),
	}

	// Resolve the tool invocation
	result, err := resolver.Resolve(tool, []byte(input), ctx)
	if err != nil {
		return fmt.Errorf("resolve error: %w", err)
	}

	// Print resolution result
	_, _ = fmt.Fprintf(out, "Permission Resolution Result:\n\n")
	_, _ = fmt.Fprintf(out, "%s\n", result.String())

	// Print trace if requested
	if showTrace {
		traceJSON, err := result.ExportTrace()
		if err != nil {
			return fmt.Errorf("export trace: %w", err)
		}

		_, _ = fmt.Fprintln(out, "\n--- Resolution Trace (JSON) ---")
		var prettyJSON map[string]any
		if err := json.Unmarshal([]byte(traceJSON), &prettyJSON); err == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
			_, _ = fmt.Fprintf(out, "  %s\n", string(formatted))
		} else {
			_, _ = fmt.Fprintf(out, "  %s\n", traceJSON)
		}
	}

	// Print dry-run notice
	if dryRun {
		_, _ = fmt.Fprintln(out, "\n[Dry run: tool not executed]")
	}

	return nil
}
