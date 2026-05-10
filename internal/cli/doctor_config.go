package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// doctorConfigCmd provides configuration diagnostics.
// This maps to REQ-V3R2-RT-005-006 and REQ-V3R2-RT-005-007.
var doctorConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration diagnostics",
	Long:  "Inspect multi-layer configuration with provenance tracking.",
}

var configDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump merged configuration with provenance",
	Long:  "Print every merged configuration key with its provenance (source tier, file path, load timestamp).",
	RunE:  runConfigDump,
}

var configDiffCmd = &cobra.Command{
	Use:   "diff <tier-a> <tier-b>",
	Short: "Compare configuration between two tiers",
	Long:  "Show keys whose values differ between two named tiers (e.g., 'user project').",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigDiff,
}

func init() {
	doctorCmd.AddCommand(doctorConfigCmd)
	doctorConfigCmd.AddCommand(configDumpCmd)
	doctorConfigCmd.AddCommand(configDiffCmd)

	configDumpCmd.Flags().StringP("format", "f", "json", "Output format (json, yaml)")
	configDumpCmd.Flags().String("key", "", "Print only a single key (e.g., 'permission.strict_mode')")
}

// runConfigDump executes the config dump command.
// This maps to REQ-V3R2-RT-005-006, REQ-V3R2-RT-005-009, REQ-V3R2-RT-005-010.
func runConfigDump(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	key, _ := cmd.Flags().GetString("key")

	resolver := config.NewResolver()
	merged, err := resolver.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Single key mode
	if key != "" {
		// Parse key as "section.field"
		section, field, err := parseKey(key)
		if err != nil {
			return fmt.Errorf("invalid key format %q: %w", key, err)
		}

		val, ok := resolver.Key(section, field)
		if !ok {
			return fmt.Errorf("key %q not found in configuration", key)
		}

		// Print single key with provenance
		// This maps to REQ-V3R2-RT-005-032
		printKeyValue(cmd.OutOrStdout(), key, val)
		return nil
	}

	// Full dump mode
	output, err := merged.Dump(format)
	if err != nil {
		return fmt.Errorf("failed to dump configuration: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)
	return nil
}

// runConfigDiff executes the config diff command.
// This maps to REQ-V3R2-RT-005-007, REQ-V3R2-RT-005-051.
func runConfigDiff(cmd *cobra.Command, args []string) error {
	tierAName := args[0]
	tierBName := args[1]

	// Parse tier names
	tierA, err := config.ParseSource(tierAName)
	if err != nil {
		return fmt.Errorf("invalid tier name %q: %w", tierAName, err)
	}

	tierB, err := config.ParseSource(tierBName)
	if err != nil {
		return fmt.Errorf("invalid tier name %q: %w", tierBName, err)
	}

	resolver := config.NewResolver()
	_, err = resolver.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get diff between tiers (merged-view delta; no error return per T-RT005-42)
	diff := resolver.Diff(tierA, tierB)

	// Print diff
	printConfigDiff(cmd.OutOrStdout(), diff, tierA, tierB)
	return nil
}

// parseKey splits a key string into section and field parts.
func parseKey(key string) (string, string, error) {
	// Simple implementation - split on first dot
	for i, r := range key {
		if r == '.' {
			return key[:i], key[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("key must contain a dot separator (e.g., 'section.field')")
}

// printKeyValue prints a single key with its provenance to w.
func printKeyValue(w io.Writer, key string, val config.Value[any]) {
	_, _ = fmt.Fprintf(w, "Key: %s\n", key)
	_, _ = fmt.Fprintf(w, "  Value: %v\n", val.V)
	_, _ = fmt.Fprintf(w, "  Source: %s\n", val.P.Source)
	_, _ = fmt.Fprintf(w, "  Origin: %s\n", val.P.Origin)
	_, _ = fmt.Fprintf(w, "  Loaded: %s\n", val.P.Loaded.Format("2006-01-02 15:04:05"))
	if val.IsDefault() {
		_, _ = fmt.Fprintf(w, "  Default: true\n")
	}
	if len(val.P.OverriddenBy) > 0 {
		_, _ = fmt.Fprintf(w, "  Overridden by:\n")
		for _, override := range val.P.OverriddenBy {
			_, _ = fmt.Fprintf(w, "    - %s\n", override)
		}
	}
}

// printConfigDiff prints the differences between two tiers to w.
func printConfigDiff(w io.Writer, diff map[string]config.Value[any], tierA, tierB config.Source) {
	if len(diff) == 0 {
		_, _ = fmt.Fprintf(w, "No differences found between %s and %s tiers\n", tierA, tierB)
		return
	}

	_, _ = fmt.Fprintf(w, "Differences between %s and %s tiers:\n", tierA, tierB)
	_, _ = fmt.Fprintf(w, "Count: %d key(s)\n\n", len(diff))

	for key, val := range diff {
		_, _ = fmt.Fprintf(w, "Key: %s\n", key)
		if val.V != nil {
			_, _ = fmt.Fprintf(w, "  Value: %v\n", val.V)
		}
		_, _ = fmt.Fprintf(w, "  Source: %s\n", val.P.Source)
		_, _ = fmt.Fprintf(w, "  Origin: %s\n", val.P.Origin)
		_, _ = fmt.Fprintln(w)
	}
}
