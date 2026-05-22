package cli

// @MX:NOTE: [AUTO] Session state management for workflow phases and checkpoints
// @MX:NOTE: [AUTO] State stored in .moai/state/ with blocker reports for unresolved issues
// @MX:NOTE: [AUTO] Supports dump, show-blocker subcommands for state inspection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// newStateCmd creates the root of the state command tree.
func newStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "state",
		Short:   "Manage session state",
		Long:    "Manage typed session state for workflow phases and checkpoints",
		GroupID: "tools",
	}
	cmd.AddCommand(newStateDumpCmd())
	cmd.AddCommand(newStateShowBlockerCmd())
	return cmd
}

// newStateDumpCmd creates the state dump subcommand.
// SPEC-V3R2-RT-004 AC-07, REQ-007, REQ-030, REQ-032: phase state dump + format selection + resume support.
func newStateDumpCmd() *cobra.Command {
	var format string
	var resume bool

	cmd := &cobra.Command{
		Use:   "dump <phase> <spec-id>",
		Short: "Dump current phase state",
		Long:  "Dump and display the current checkpoint state for a given phase and SPEC ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]
			specID := args[1]
			return runStateDump(phase, specID, format, resume)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "human", "출력 형식: json 또는 human")
	cmd.Flags().BoolVar(&resume, "resume", false, "stale checkpoint도 강제 로드 (--resume 모드)")

	return cmd
}

// newStateShowBlockerCmd creates the state show-blocker subcommand.
func newStateShowBlockerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-blocker",
		Short: "Show outstanding blocker",
		Long:  "Display the most recent unresolved blocker report",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowBlocker()
		},
	}
}

// runStateDump implements the state dump command.
// SPEC-V3R2-RT-004 AC-07, REQ-030, REQ-032: phase+specID based dump + format selection.
func runStateDump(phaseArg, specID, format string, resume bool) error {
	// Locate state directory
	stateDir, err := findStateDir()
	if err != nil {
		return fmt.Errorf("find state dir: %w", err)
	}

	// Create store
	store := session.NewFileSessionStore(stateDir, 3600*time.Second)

	// Parse phase
	phase := session.Phase(phaseArg)
	if !phase.Valid() {
		return fmt.Errorf("invalid phase: %s", phaseArg)
	}

	// Use HydrateWithOpts based on --resume flag
	// SPEC-V3R2-RT-004 AC-06: --resume flag is wired into HydrateWithOpts(SkipStaleCheck=true).
	opts := session.HydrateOpts{SkipStaleCheck: resume}
	state, err := store.HydrateWithOpts(phase, specID, opts)
	if err != nil {
		if err == session.ErrCheckpointStale {
			fmt.Fprintf(os.Stderr, "Warning: Checkpoint is stale. Use --resume to force load.\n")
			return err
		}
		return fmt.Errorf("hydrate state: %w", err)
	}

	if state == nil {
		fmt.Printf("No checkpoint found for phase %s, SPEC %s\n", phaseArg, specID)
		return nil
	}

	// Select output format
	switch format {
	case "json":
		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state: %w", err)
		}
		fmt.Println(string(data))
	default: // "human" or any other
		printPhaseStateHuman(state)
	}

	return nil
}

// printPhaseStateHuman prints a PhaseState in a human-readable format.
func printPhaseStateHuman(state *session.PhaseState) {
	fmt.Printf("Phase:     %s\n", state.Phase)
	fmt.Printf("SPEC ID:   %s\n", state.SPECID)
	fmt.Printf("Updated:   %s\n", state.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Provenance: source=%s origin=%s\n", state.Provenance.Source, state.Provenance.Origin)
	if state.BlockerRpt != nil {
		fmt.Printf("Blocker:   kind=%s resolved=%v\n", state.BlockerRpt.Kind, state.BlockerRpt.Resolved)
	}
	if state.Checkpoint != nil {
		data, _ := json.MarshalIndent(state.Checkpoint, "  ", "  ")
		fmt.Printf("Checkpoint:\n  %s\n", string(data))
	}
}

// runShowBlocker implements the show-blocker command.
func runShowBlocker() error {
	// Determine state directory
	stateDir, err := findStateDir()
	if err != nil {
		return fmt.Errorf("find state dir: %w", err)
	}

	// Find blocker files
	pattern := filepath.Join(stateDir, "blocker-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob blockers: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No blockers found")
		return nil
	}

	// Find the most recent unresolved blocker
	var latestBlocker *session.BlockerReport
	var latestTime time.Time

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		var blocker session.BlockerReport
		if err := json.Unmarshal(data, &blocker); err != nil {
			continue
		}

		if !blocker.Resolved && blocker.Timestamp.After(latestTime) {
			latestBlocker = &blocker
			latestTime = blocker.Timestamp
		}
	}

	if latestBlocker == nil {
		fmt.Println("No outstanding blockers found")
		return nil
	}

	// Pretty print blocker
	data, err := json.MarshalIndent(latestBlocker, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal blocker: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// findStateDir walks up the directory tree looking for .moai/state/.
func findStateDir() (string, error) {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up looking for .moai/state/
	dir := cwd
	for {
		stateDir := filepath.Join(dir, ".moai", "state")
		if info, err := os.Stat(stateDir); err == nil && info.IsDir() {
			return stateDir, nil
		}

		// Move to parent
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".moai/state/ directory not found from %s", cwd)
}
