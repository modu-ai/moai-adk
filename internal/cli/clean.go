package cli

// SPEC-V3R2-RT-004 REQ-031, AC-13: cleanup of runs/ directory based on retention_days.
// Default behavior: dry-run (no actual deletion). Use --force flag to perform real deletion.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hygiene"
)

// newCleanCmd creates the clean subcommand.
func newCleanCmd() *cobra.Command {
	var force bool
	var home bool
	var codexSkills bool
	var reportsArchive bool
	var reportsArchiveDays int
	var auditLogs bool
	var sessionState bool
	var apply bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up stale run artifacts",
		Long: `Clean up run artifacts in .moai/state/runs/ that are older than retention_days.
Default: dry-run mode (no actual deletion). Use --force to actually delete.

retention_days is read from .moai/config/sections/state.yaml.

Exactly one scope is cleaned per invocation. --home, --codex-skills,
--reports-archive, --audit-logs and --session-state select different files
and may not be combined.

With --home, clean the ~/.moai home directory instead of the project scope.
The default is a report-only dry-run. --force removes per-profile projects/
entries older than 180 days, debug/ entries older than 30 days, and the oldest
projects/ entries needed to bring a profile under 5 GiB. It also repairs every
directory under ~/.moai to mode 0700. Profiles unused for 90 days and byte-identical
plugin trees are reported but never deleted automatically. Releases, root
logs/, and backups/removed-* retain the existing home-retention policy. This
scope touches only ~/.moai — ~/.claude is never modified.

With --codex-skills, remove ghost [[skills.config]] registrations from
~/.codex/config.toml (or $CODEX_HOME/config.toml) — entries whose declared
path is provably absent. This scope MODIFIES ~/.codex/config.toml. An entry
is kept whenever its absence cannot be proven: a relative or oddly-formed
path, an unresolvable home, a stat that did not complete, a path that
resolves, or a line range holding anything the parser did not recognise.
Under --force the file is backed up first and the backup path and sha256 are
reported.

With --reports-archive, move aging evidence directories out of
.moai/reports/ into .moai/reports/archive/<YYYY-MM>/ (move-only — nothing is
ever deleted). A candidate is a top-level entry whose name is
evidence-shaped (t<digits> or SPEC-<DOMAIN>-<NNN>), whose mtime is older
than the retention window (--reports-archive-days, default 90), and that
holds no git-tracked files. historical/, plan-audit/, worktrees/ and
archive/ are never candidates. Dry-run by default.

With --audit-logs, run the audit-log rotator (SPEC-MOAI-HYGIENE-001): it
holds every registered sink under .moai/logs/ to a size bound (default
10 MiB, keep-1). With --session-state, run the finished-session state GC:
it removes only DEAD sessions' content-datable residue past the age floor
(default 7 days), never touching LIVE or indeterminate sessions. Both
scopes are dry-run by default and mutate only with --apply on THIS
invocation — the workflow.hygiene.mode config governs the SessionStart
auto path alone and never makes a CLI invocation mutate.`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Status output routes through the Printer to stderr
			// (SPEC-CLI-TUX-V3-001 REQ-CTX-012/017 ratchet migration).
			p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
			if countScopes(home, codexSkills, reportsArchive, auditLogs, sessionState) > 1 {
				return fmt.Errorf("--home, --codex-skills, --reports-archive, --audit-logs and --session-state select different scopes; pass exactly one")
			}
			if reportsArchive {
				return runCleanReportsArchive(p, force, reportsArchiveDays)
			}
			if codexSkills {
				return runCleanCodexSkills(p, force)
			}
			if home {
				return runCleanHome(p, force)
			}
			if auditLogs {
				return runCleanHygiene(p, hygieneScopeAuditLogs, apply)
			}
			if sessionState {
				return runCleanHygiene(p, hygieneScopeSessionState, apply)
			}
			return runClean(p, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Actually delete files (default: dry-run)")
	cmd.Flags().BoolVar(&home, "home", false, "Clean the ~/.moai home directory (allowlist-only; dry-run by default)")
	cmd.Flags().BoolVar(&codexSkills, "codex-skills", false, "Remove provably-absent [[skills.config]] entries from ~/.codex/config.toml (dry-run by default)")
	cmd.Flags().BoolVar(&reportsArchive, "reports-archive", false, "Move aging evidence directories from .moai/reports/ into archive/<YYYY-MM>/ (move-only; dry-run by default)")
	cmd.Flags().IntVar(&reportsArchiveDays, "reports-archive-days", config.DefaultReportsArchiveRetentionDays, "Retention window in days for --reports-archive candidates")
	cmd.Flags().BoolVar(&auditLogs, "audit-logs", false, "Run the audit-log rotator over .moai/logs/ (dry-run by default; mutates only with --apply)")
	cmd.Flags().BoolVar(&sessionState, "session-state", false, "Run the finished-session state GC over .moai/state/ (dry-run by default; mutates only with --apply)")
	cmd.Flags().BoolVar(&apply, "apply", false, "Let this --audit-logs / --session-state invocation mutate (overrides workflow.hygiene.mode for this invocation only)")

	return cmd
}

// countScopes counts how many clean scopes are selected.
func countScopes(scopes ...bool) int {
	n := 0
	for _, s := range scopes {
		if s {
			n++
		}
	}
	return n
}

// hygieneScope selects which hygiene unit the CLI invocation runs.
type hygieneScope int

const (
	hygieneScopeAuditLogs hygieneScope = iota
	hygieneScopeSessionState
)

// runCleanHygiene runs one hygiene unit for the project scope: --audit-logs
// runs the rotator, --session-state runs the GC. The invocation mutates
// only with --apply (REQ-HYG-013): the workflow.hygiene.mode config governs
// the SessionStart auto path alone and is never consulted for the CLI's own
// mutation decision.
func runCleanHygiene(p printer.Printer, scope hygieneScope, apply bool) error {
	stateDir, err := findStateDirNoEnv()
	if err != nil {
		return fmt.Errorf("find state dir: %w", err)
	}
	moaiDir := filepath.Dir(stateDir)
	projectRoot := filepath.Dir(moaiDir)
	printResolvedRoot(p, stateDir)

	settings := hygiene.LoadSettingsFrom(projectRoot)
	if err := settings.Validate(); err != nil {
		// D30: a config-invalid value refuses the run — mutation never
		// proceeds on unvalidated floors.
		p.Warn("%v", err)
		return err
	}

	mode := hygiene.ModeReport
	if apply {
		mode = hygiene.ModeApply
	}
	if mode == hygiene.ModeReport {
		p.Info("hygiene: dry-run (pass --apply on this invocation to mutate)")
	}

	switch scope {
	case hygieneScopeAuditLogs:
		r := &hygiene.Rotator{
			LogDir:        filepath.Join(moaiDir, "logs"),
			MaxBytes:      settings.AuditLogMaxBytes,
			KeptRotations: settings.AuditLogKeptRotations,
		}
		rows, err := r.Run(mode)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Outcome == hygiene.OutcomeSummary {
				printHygieneSummary(p, row.Counts, mode)
				continue
			}
			printHygieneRow(p, row.Outcome.String(), row.Path, row.Reason, mode)
		}
	case hygieneScopeSessionState:
		g := &hygiene.GC{
			MoaiRoot:         moaiDir,
			RegistryPath:     filepath.Join(moaiDir, "state", "active-sessions.json"),
			TranscriptRoots:  hygiene.DefaultTranscriptRoots(),
			MinAge:           settings.MinAge(),
			TranscriptWindow: settings.TranscriptActivityWindow,
			HeartbeatWindow:  settings.HeartbeatStaleWindow,
		}
		rep, err := g.Run(mode)
		if err != nil {
			return err
		}
		for _, d := range rep.Decisions {
			printHygieneRow(p, d.Outcome, d.Path, d.Reason, mode)
		}
	}
	return nil
}

// printHygieneSummary renders a summary row's outcome counts.
func printHygieneSummary(p printer.Printer, counts map[string]int, mode hygiene.Mode) {
	prefix := ""
	if mode == hygiene.ModeReport {
		prefix = "[dry-run] "
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, counts[k]))
	}
	p.Info("%ssummary: %s", prefix, strings.Join(parts, ", "))
}

// printHygieneRow renders one decision line.
func printHygieneRow(p printer.Printer, outcome, path, reason string, mode hygiene.Mode) {
	prefix := ""
	if mode == hygiene.ModeReport {
		prefix = "[dry-run] "
	}
	if reason != "" {
		p.Info("%s%s: %s — %s", prefix, outcome, path, reason)
		return
	}
	p.Info("%s%s: %s", prefix, outcome, path)
}

// stateYAMLWrapper is the top-level key structure of state.yaml.
type stateYAMLWrapper struct {
	State struct {
		RetentionDays int `yaml:"retention_days"`
	} `yaml:"state"`
}

// runClean cleans up old runs/ directories based on retention_days.
//
// Resolution goes through findStateDirNoEnv, which consults no environment
// variable (SPEC-CLI-STATE-DIR-BOUND-001 REQ-9): this function removes
// directories, and an inherited variable naming some other checkout must not be
// what decides where. To clean another project, run this in it. See
// findStateDirNoEnv for which variable that is and why it is excluded here.
func runClean(p printer.Printer, force bool) error {
	// Locate state directory
	stateDir, err := findStateDirNoEnv()
	if err != nil {
		return fmt.Errorf("find state dir: %w", err)
	}
	// Announced before anything is enumerated: read commands honour the
	// project-directory environment variable and this one does not, so the two
	// can legitimately resolve different projects within one session. Saying
	// which project this is after listing its files would be too late.
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
