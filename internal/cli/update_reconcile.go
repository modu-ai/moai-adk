package cli

// update_reconcile.go — SPEC-UPDATE-MIGRATION-001 M4 (card t1547): the cli
// wiring helpers for the preservation-based update pipeline. The default
// existing-project path replaces the wholesale clean with the two-phase
// reconciliation (classify → preserve/archive-remove → [deploy] →
// merge-or-conflict); the wholesale path survives only for the legacy
// v1→v2 fresh-install case, behind the REQ-UPM-015 guard.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// legacyFreshConfigInstall reports whether the project needs the legacy
// wholesale config fresh-install path (REQ-UPM-040): its system.yaml exists
// but cannot be parsed, so no section-level merge can represent it — the
// config model itself is incompatible, not merely stale. A missing system.yaml
// is NOT legacy (a record-less project is the normal update entry); a healthy
// one is not either. The fresh-install branch still runs under the
// REQ-UPM-015 guard and only after the Backup step's full config copy.
func legacyFreshConfigInstall(projectRoot string) bool {
	systemYAML := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)
	data, err := os.ReadFile(systemYAML)
	if err != nil {
		return false // absent (or unreadable for other reasons): not the legacy shape
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return true // unparseable: incompatible with the sectioned merge model
	}
	return false
}

// reconcileTargets returns the classification scope for a default-path run:
// the run's clean-target list (deploy-mode aware, migration-aware — the same
// list computeRunCleanTargets produces) PLUS .moai/config, the directory the
// wholesale clean used to remove inline and the pipeline now classifies
// instead (REQ-UPM-020).
func reconcileTargets(projectRoot string, cleanTargets []deploy.CleanTarget) []deploy.CleanTarget {
	targets := make([]deploy.CleanTarget, len(cleanTargets))
	copy(targets, cleanTargets)
	targets = append(targets, deploy.CleanTarget{
		DisplayPath: filepath.Join(defs.MoAIDir, defs.ConfigSubdir),
		FullPath:    filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir),
	})
	return targets
}

// reconcileExclude reports the paths another step of the update flow already
// reconciles, so the pipeline's merge phase never reprocesses them (card
// t1547 review finding 2): the config sections are merged back by
// RestoreMoaiConfigRetained from the Backup step's copy, with the
// deploy-time snapshot base.
func reconcileExclude(rel string) bool {
	return strings.HasPrefix(rel, ".moai/config/sections/")
}
