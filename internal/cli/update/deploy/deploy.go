// Package deploy contains the template-deploy, legacy-migration, and
// evolution-scaffold functions extracted from the root update.go as part of
// SPEC-CLI-TUX-V3-003 M3d-B1. These functions execute the destructive side of
// the update flow (stale-path cleanup, .moai/memory → .moai/state migration,
// .agency → .moai migration adapter, .moai/evolution scaffold) and share NO
// root-package-cli helpers, so the package is a leaf (cli → deploy one-way).
//
// Namespace-protection care: CleanMoaiManagedPaths is the stale-file removal
// step that runs BEFORE template deployment. It only removes MoAI-managed
// paths; user-owned namespace preservation is governed by the plan package's
// IsUserOwnedNamespace predicate at the orchestration layer, NOT here.
package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// CleanMoaiManagedPaths removes MoAI-managed directories and files before template
// deployment. This ensures stale files are cleaned up during version upgrades.
// The .moai/config/ directory is deleted entirely (backup was done by the Backup step).
// Paths that do not exist are silently skipped.
func CleanMoaiManagedPaths(projectRoot string, out io.Writer) error {
	type cleanTarget struct {
		// displayPath is shown in progress messages (e.g., ".claude/settings.json")
		displayPath string
		// fullPath is the absolute filesystem path to delete
		fullPath string
		// isGlob indicates the target uses filepath.Glob matching
		isGlob bool
	}

	targets := []cleanTarget{
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.SettingsJSON),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SettingsJSON),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.CommandsMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.CommandsMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.AgentsMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			isGlob:      true,
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.RulesMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.RulesMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.HooksMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.HooksMoaiSubdir),
		},
	}

	// Process standard targets (files and directories)
	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair in this hot path (REQ-UPR-004).
	for _, t := range targets {
		pl := tui.ProgressLine(out, fmt.Sprintf("Removing %s...", t.displayPath), nil)

		if t.isGlob {
			matches, err := filepath.Glob(t.fullPath)
			if err != nil {
				pl.Fail(fmt.Sprintf("Failed to glob %s: %v", t.displayPath, err))
				return fmt.Errorf("glob %s: %w", t.displayPath, err)
			}
			for _, match := range matches {
				if err := os.RemoveAll(match); err != nil {
					pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.displayPath, err))
					return fmt.Errorf("remove %s: %w", match, err)
				}
			}
			pl.Done(fmt.Sprintf("Removed %s", t.displayPath))
			continue
		}

		if _, err := os.Stat(t.fullPath); err != nil {
			if os.IsNotExist(err) {
				// Use Done with a "skipped" marker — semantically a successful
				// no-op (target already absent). The leading symbol shifts from
				// "-" to "✓" but the message text is preserved; this aligns
				// the visual category with the success branch.
				pl.Done(fmt.Sprintf("Skipped %s (not found)", t.displayPath))
				continue
			}
			pl.Fail(fmt.Sprintf("Failed to stat %s: %v", t.displayPath, err))
			return fmt.Errorf("stat %s: %w", t.displayPath, err)
		}

		if err := os.RemoveAll(t.fullPath); err != nil {
			pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.displayPath, err))
			return fmt.Errorf("remove %s: %w", t.displayPath, err)
		}
		pl.Done(fmt.Sprintf("Removed %s", t.displayPath))
	}

	// Clean .moai/config/ entirely - backup was already done by the Backup step.
	// For v1.x -> v2.x: old config is incompatible, fresh install needed.
	// For v2.x -> v2.x: backup includes sections/, restore will merge values back.
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	configDisplayPath := filepath.Join(defs.MoAIDir, defs.ConfigSubdir)
	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair (REQ-UPR-004).
	plConfig := tui.ProgressLine(out, fmt.Sprintf("Removing %s...", configDisplayPath), nil)

	if err := os.RemoveAll(configDir); err != nil {
		if !os.IsNotExist(err) {
			plConfig.Fail(fmt.Sprintf("Failed to remove %s: %v", configDisplayPath, err))
			return fmt.Errorf("remove %s: %w", configDisplayPath, err)
		}
	}
	plConfig.Done(fmt.Sprintf("Removed %s", configDisplayPath))

	// Migrate legacy .moai/memory/ to .moai/state/.
	// Prior to v2.x, state files (checkpoints, coverage, diagnostics) lived under
	// .moai/memory/. If the old directory still exists, migrate or remove it.
	if err := MigrateLegacyMemoryDir(projectRoot, out); err != nil {
		return err
	}

	return nil
}

// MigrateLegacyMemoryDir handles the .moai/memory/ → .moai/state/ migration.
// If only the old directory exists, it is renamed. If both exist, the old one
// is removed (the new directory takes precedence). If neither exists, this is
// a no-op because template deployment will create .moai/state/.
func MigrateLegacyMemoryDir(projectRoot string, out io.Writer) error {
	legacyDir := filepath.Join(projectRoot, defs.MoAIDir, "memory")
	stateDir := filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir)

	legacyDisplayPath := filepath.Join(defs.MoAIDir, "memory")

	legacyExists := false
	if _, err := os.Stat(legacyDir); err == nil {
		legacyExists = true
	}

	if !legacyExists {
		return nil
	}

	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair (REQ-UPR-004).
	plLegacy := tui.ProgressLine(out, fmt.Sprintf("Migrating %s...", legacyDisplayPath), nil)

	stateExists := false
	if _, err := os.Stat(stateDir); err == nil {
		stateExists = true
	}

	if !stateExists {
		// Rename .moai/memory/ → .moai/state/ (fast atomic move).
		if err := os.Rename(legacyDir, stateDir); err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to migrate %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("migrate %s to %s: %w", legacyDisplayPath, defs.StateSubdir, err)
		}
		plLegacy.Done(fmt.Sprintf("Migrated %s → %s", legacyDisplayPath, filepath.Join(defs.MoAIDir, defs.StateSubdir)))
	} else {
		// Both exist — state directory takes precedence; remove legacy.
		if err := os.RemoveAll(legacyDir); err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to remove %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("remove legacy %s: %w", legacyDisplayPath, err)
		}
		plLegacy.Done(fmt.Sprintf("Removed legacy %s", legacyDisplayPath))
	}

	return nil
}

// ScaffoldEvolutionDir ensures the .moai/evolution/ directory tree exists.
// Called during both init and update so that existing projects that predate
// the evolution infrastructure receive the directory structure automatically.
// Non-destructive: only creates missing files; never overwrites existing ones.
func ScaffoldEvolutionDir(projectRoot string) error {
	evolutionDir := filepath.Join(projectRoot, ".moai", "evolution")

	// Sub-directories to create.
	subdirs := []string{
		"telemetry",
		"learnings",
		"new-skills",
	}

	for _, sub := range subdirs {
		dir := filepath.Join(evolutionDir, sub)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}

		// Ensure .gitkeep exists so the directory is tracked by git.
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0o644); err != nil {
				return fmt.Errorf("create %s: %w", gitkeep, err)
			}
		}
	}

	// Create default manifest.yaml if missing.
	manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		const defaultManifest = `schema_version: 1
evolved_skills: []
new_skills: []
learnings_count: 0
last_evolution_date: ""
rate_limit:
  week_start: ""
  proposals_this_week: 0
  last_proposal_time: ""
`
		if err := os.WriteFile(manifestPath, []byte(defaultManifest), 0o644); err != nil {
			return fmt.Errorf("create %s: %w", manifestPath, err)
		}
	}

	// Create default changelog.md if missing.
	changelogPath := filepath.Join(evolutionDir, "changelog.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		const defaultChangelog = `# MoAI Evolution Changelog

All notable skill evolutions and learning graduations will be documented here.

## Format

Each entry: date, learning ID, skill affected, change summary.
`
		if err := os.WriteFile(changelogPath, []byte(defaultChangelog), 0o644); err != nil {
			return fmt.Errorf("create %s: %w", changelogPath, err)
		}
	}

	return nil
}
