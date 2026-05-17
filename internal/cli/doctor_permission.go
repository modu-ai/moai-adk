package cli

import (
	"encoding/json"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/permission"
	"github.com/spf13/cobra"
)

var permissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "Diagnose permission resolution",
	Long:  "Diagnose permission stack resolution with trace output and rule provenance.",
	RunE:  runDoctorPermission,
}

func init() {
	doctorCmd.AddCommand(permissionCmd)

	permissionCmd.Flags().String("tool", "", "Tool name to resolve (e.g., Bash, Read, Write)")
	permissionCmd.Flags().String("input", "", "Tool input arguments")
	permissionCmd.Flags().Bool("trace", false, "Show full resolution trace as JSON")
	permissionCmd.Flags().Bool("dry-run", false, "Show resolution without executing")

	// T-RT002-28: 추가 플래그 — --all-tiers, --mode, --fork, --format.
	permissionCmd.Flags().Bool("all-tiers", false, "Show all 8 tiers in the resolution output")
	permissionCmd.Flags().String("mode", "default", "Permission mode to simulate (default|acceptEdits|bypassPermissions|plan|bubble)")
	permissionCmd.Flags().Bool("fork", false, "Simulate fork agent context (IsFork=true)")
	permissionCmd.Flags().String("format", "human", "Output format: human|json")
}

// runDoctorPermission executes the permission diagnostic workflow.
//
// Usage:
//
//	moai doctor permission --tool Bash --input "go test ./..." --trace
//	moai doctor permission --tool Write --input "/tmp/test.txt" --dry-run
//	moai doctor permission --tool Bash --input "go build" --all-tiers --format json
//	moai doctor permission --tool Write --input "/tmp/x" --mode plan
//	moai doctor permission --tool Write --input "/tmp/x" --fork
func runDoctorPermission(cmd *cobra.Command, _ []string) error {
	tool := getStringFlag(cmd, "tool")
	input := getStringFlag(cmd, "input")
	showTrace := getBoolFlag(cmd, "trace")
	dryRun := getBoolFlag(cmd, "dry-run")
	allTiers := getBoolFlag(cmd, "all-tiers")
	modeStr := getStringFlag(cmd, "mode")
	isFork := getBoolFlag(cmd, "fork")
	format := getStringFlag(cmd, "format")

	if tool == "" {
		return fmt.Errorf("--tool is required")
	}

	out := cmd.OutOrStdout()

	// 모드 파싱.
	mode, err := permission.ParsePermissionMode(modeStr)
	if err != nil {
		mode = permission.ModeDefault
	}

	// Create resolver with context from flags
	resolver := permission.NewPermissionResolver()
	ctx := permission.ResolveContext{
		Mode:            mode,
		IsFork:          isFork,
		ParentAvailable: !isFork, // fork 시 parent available 기본 true.
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

	// --format json 출력.
	if format == "json" {
		return printJSONResult(out, result, tool, input)
	}

	// Human-readable 출력.
	_, _ = fmt.Fprintf(out, "Permission Resolution Result:\n\n")
	_, _ = fmt.Fprintf(out, "%s\n", result.String())

	// --all-tiers: 모든 tier 출력.
	if allTiers {
		_, _ = fmt.Fprintln(out, "\n--- All Tiers Inspected ---")
		for i, try := range result.Trace.Tries {
			_, _ = fmt.Fprintf(out, "  [%d] tier=%s matched=%v", i+1, try.Tier, try.Matched)
			if try.Rule != nil {
				_, _ = fmt.Fprintf(out, " rule=%s action=%s origin=%s",
					try.Rule.Pattern, try.Rule.Action, try.Rule.Origin)
			}
			_, _ = fmt.Fprintln(out)
		}
	}

	// --trace: JSON trace 출력.
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

		// hook tier sentinel 을 "tier": "hook" 으로 stringify.
		// T-RT002-28: result.ExportTrace() 의 hook tier sentinel (config.Source(999)) 처리.
		_, _ = fmt.Fprintln(out, "  (hook tier displayed as \"hook\" in trace output)")
	}

	// --dry-run notice.
	if dryRun {
		_, _ = fmt.Fprintln(out, "\n[Dry run: tool not executed]")
	}

	return nil
}

// printJSONResult JSON 형식으로 resolution 결과를 출력한다.
// T-RT002-28: --format json 지원.
func printJSONResult(out interface{ Write(p []byte) (n int, err error) }, result *permission.ResolveResult, tool, input string) error {
	type jsonOutput struct {
		Tool       string `json:"tool"`
		Input      string `json:"input"`
		Decision   string `json:"decision"`
		ResolvedBy string `json:"resolved_by"`
		Origin     string `json:"origin"`
		SystemMsg  string `json:"system_message,omitempty"`
	}

	output := jsonOutput{
		Tool:       tool,
		Input:      input,
		Decision:   string(result.Decision),
		ResolvedBy: result.ResolvedBy.String(),
		Origin:     result.Origin,
		SystemMsg:  result.SystemMessage,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	_, err = fmt.Fprintf(out, "%s\n", string(data))
	return err
}
