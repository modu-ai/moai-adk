package cli

// SPEC-V3R2-RT-004 REQ-031, AC-13: cleanup of runs/ directory based on retention_days.
// Default behavior: dry-run (no actual deletion). Use --force flag to perform real deletion.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
)

// newCleanCmd creates the clean subcommand.
func newCleanCmd() *cobra.Command {
	var force bool
	var home bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up stale run artifacts",
		Long: `Clean up run artifacts in .moai/state/runs/ that are older than retention_days.
Default: dry-run mode (no actual deletion). Use --force to actually delete.

retention_days is read from .moai/config/sections/state.yaml.

With --home, clean the ~/.moai home directory instead of the project scope:
aged per-profile debug/ entries, releases/ binaries beyond the current
version + the 3 newest, aged root logs/, and aged backups/removed-*
directories. Only ~/.moai is touched — ~/.claude is never modified. Home
retention comes from state.home_retention_days in ~/.moai/config/sections/
state.yaml (default 30 days; explicit 0 disables).`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Status output routes through the Printer to stderr
			// (SPEC-CLI-TUX-V3-001 REQ-CTX-012/017 ratchet migration).
			p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
			if home {
				return runCleanHome(p, force)
			}
			return runClean(p, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Actually delete files (default: dry-run)")
	cmd.Flags().BoolVar(&home, "home", false, "Clean the ~/.moai home directory (allowlist-only; dry-run by default)")

	return cmd
}

// stateYAMLWrapper is the top-level key structure of state.yaml.
type stateYAMLWrapper struct {
	State struct {
		RetentionDays int `yaml:"retention_days"`
	} `yaml:"state"`
}

// runClean cleans up old runs/ directories based on retention_days.
//
// Resolution deliberately ignores CLAUDE_PROJECT_DIR
// (SPEC-CLI-STATE-DIR-BOUND-001 REQ-9): this function removes directories, and
// an inherited environment variable must not be what decides where. A worktree
// session whose CLAUDE_PROJECT_DIR still names the primary checkout would
// otherwise delete there. To clean another project, run this in it.
func runClean(p printer.Printer, force bool) error {
	// Locate state directory
	stateDir, err := findStateDirNoEnv()
	if err != nil {
		return fmt.Errorf("find state dir: %w", err)
	}
	// Announced before anything is enumerated: read commands honour
	// CLAUDE_PROJECT_DIR and this one does not, so the two can legitimately
	// resolve different projects within one session. Saying which project this
	// is after listing its files would be telling the operator too late.
	printResolvedRoot(p, stateDir)

	// Load retention_days (from state.yaml)
	retentionDays, err := loadRetentionDays(stateDir)
	if err != nil {
		return fmt.Errorf("load retention_days: %w", err)
	}

	if retentionDays <= 0 {
		p.Info("retention_days not configured or 0; nothing to clean")
		return nil
	}

	// Scan runs/ directory
	runsDir := filepath.Join(stateDir, "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			p.Info("runs/ directory not found at %s; nothing to clean", runsDir)
			return nil
		}
		return fmt.Errorf("read runs dir: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var toDelete []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			toDelete = append(toDelete, filepath.Join(runsDir, entry.Name()))
		}
	}

	if len(toDelete) == 0 {
		p.Info("No runs older than %d days found", retentionDays)
		return nil
	}

	// Dry-run or actual deletion
	for _, path := range toDelete {
		if force {
			if err := os.RemoveAll(path); err != nil {
				p.Warn("failed to remove %s: %v", path, err)
			} else {
				p.Info("Deleted: %s", path)
			}
		} else {
			p.Info("[dry-run] Would delete: %s", path)
		}
	}

	if !force {
		p.Info("%d runs eligible for deletion. Run with --force to actually delete.", len(toDelete))
	}

	return nil
}

// loadRetentionDays reads retention_days from .moai/config/sections/state.yaml.
func loadRetentionDays(stateDir string) (int, error) {
	// stateDir is .moai/state/, so navigate to .moai/config/sections/
	moaiDir := filepath.Dir(stateDir) // .moai/
	configPath := filepath.Join(moaiDir, "config", "sections", "state.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No state.yaml: retention_days = 0 (disabled)
		}
		return 0, fmt.Errorf("read state.yaml: %w", err)
	}

	var wrapper stateYAMLWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return 0, fmt.Errorf("parse state.yaml: %w", err)
	}

	return wrapper.State.RetentionDays, nil
}
