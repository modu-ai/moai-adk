// Package cli — doctor_hook.go
// Implements "moai doctor hook" subcommand with 27-event coverage table.
// SPEC-V3R2-RT-006 REQ-050, REQ-051, AC-12.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// doctorHookOutput is the JSON output shape for "moai doctor hook --json".
type doctorHookOutput struct {
	CoverageTable []doctorHookEntry `json:"coverage_table"`
	Summary       hook.CoverageSummary `json:"summary"`
}

// doctorHookEntry is a single row in the JSON coverage table.
type doctorHookEntry struct {
	EventName          string `json:"event_name"`
	Resolution         string `json:"resolution"`
	IsActive           bool   `json:"is_active"`
	ObservabilityOptIn bool   `json:"observability_opt_in"`
	HandlerFile        string `json:"handler_file"`
}

// doctorHookCmd is the "moai doctor hook" cobra subcommand.
var doctorHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Show 27-event hook coverage table",
	Long:  "Print the 27-event hook coverage table with per-event resolution state and observability opt-in status.",
	RunE:  runDoctorHook,
}

func init() {
	doctorCmd.AddCommand(doctorHookCmd)
	doctorHookCmd.Flags().Bool("json", false, "Output as JSON")
	doctorHookCmd.Flags().String("trace", "", "Show recent log lines for the named hook event")
	doctorHookCmd.Flags().Bool("observability", false, "Filter to show only RETIRE-OBS-ONLY events")
}

// runDoctorHook implements the "moai doctor hook" command.
func runDoctorHook(cmd *cobra.Command, _ []string) error {
	jsonOutput := getBoolFlag(cmd, "json")
	traceEvent := getStringFlag(cmd, "trace")
	obsOnly := getBoolFlag(cmd, "observability")

	out := cmd.OutOrStdout()

	// Handle --trace flag: tail hook.log for the named event.
	if traceEvent != "" {
		return runDoctorHookTrace(out, traceEvent)
	}

	// Build enriched table from CoverageTable.
	entries := buildDoctorHookEntries(obsOnly)
	summary := hook.Summarize()

	if jsonOutput {
		return printDoctorHookJSON(out, entries, summary)
	}

	return printDoctorHookText(out, entries, summary)
}

// buildDoctorHookEntries converts the canonical CoverageTable to CLI output entries.
func buildDoctorHookEntries(obsOnly bool) []doctorHookEntry {
	entries := make([]doctorHookEntry, 0, len(hook.CoverageTable))
	for _, e := range hook.CoverageTable {
		if obsOnly && e.Resolution != hook.ResolutionRetireObsOnly {
			continue
		}
		entries = append(entries, doctorHookEntry{
			EventName:          e.EventName,
			Resolution:         string(e.Resolution),
			IsActive:           e.IsActive,
			ObservabilityOptIn: e.ObservabilityOptIn,
			HandlerFile:        e.HandlerFile,
		})
	}
	return entries
}

// printDoctorHookJSON writes JSON output to w.
func printDoctorHookJSON(w io.Writer, entries []doctorHookEntry, summary hook.CoverageSummary) error {
	output := doctorHookOutput{
		CoverageTable: entries,
		Summary:       summary,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	_, _ = fmt.Fprintln(w, string(data))
	return nil
}

// printDoctorHookText writes human-readable table output to w.
func printDoctorHookText(w io.Writer, entries []doctorHookEntry, summary hook.CoverageSummary) error {
	_, _ = fmt.Fprintln(w, "Hook Coverage Table — SPEC-V3R2-RT-006 §5.7")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 80))
	_, _ = fmt.Fprintf(w, "%-35s %-20s %-8s %s\n", "Event", "Resolution", "Active", "Handler")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 80))

	for _, e := range entries {
		activeStr := "✓"
		if !e.IsActive {
			if e.ObservabilityOptIn {
				activeStr = "obs"
			} else {
				activeStr = "—"
			}
		}
		_, _ = fmt.Fprintf(w, "%-35s %-20s %-8s %s\n",
			e.EventName, e.Resolution, activeStr, e.HandlerFile)
	}

	_, _ = fmt.Fprintln(w, strings.Repeat("─", 80))
	_, _ = fmt.Fprintf(w, "Summary: total=%d KEEP=%d UPGRADE=%d FIX=%d RETIRE=%d REMOVE=%d COMPOSITE=%d\n",
		summary.Total, summary.Keep, summary.Upgrade, summary.Fix,
		summary.RetireObsOnly, summary.Remove, summary.Composite)
	return nil
}

// runDoctorHookTrace tails .moai/logs/hook.log for lines matching the event name.
// Out-of-scope simplification: tail-only readout (no real-time stream).
// SPEC-V3R2-RT-006 REQ-051.
func runDoctorHookTrace(w io.Writer, eventName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	logPath := filepath.Join(cwd, ".moai", "logs", "hook.log")
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(w, "No hook.log found at %s\n", logPath)
			return nil
		}
		return fmt.Errorf("open hook.log: %w", err)
	}
	defer func() { _ = f.Close() }()

	needle := strings.ToLower(eventName)
	var matches []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), needle) {
			matches = append(matches, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan hook.log: %w", err)
	}

	if len(matches) == 0 {
		_, _ = fmt.Fprintf(w, "No log lines found for event %q in %s\n", eventName, logPath)
		return nil
	}

	// Show last N lines (most recent).
	const maxLines = 20
	start := 0
	if len(matches) > maxLines {
		start = len(matches) - maxLines
	}
	for _, line := range matches[start:] {
		_, _ = fmt.Fprintln(w, line)
	}
	return nil
}
