package cli

// update_reconcile.go — SPEC-UPDATE-MIGRATION-001 M4 (card t1547): the cli
// wiring helpers for the preservation-based update pipeline. The default
// existing-project path replaces the wholesale clean with the two-phase
// reconciliation (classify → preserve/archive-remove → [deploy] →
// merge-or-conflict); the wholesale path survives only for the legacy
// v1→v2 fresh-install case, behind the REQ-UPM-015 guard.

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
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
// list computeRunCleanTargets produces) PLUS the classification-only roots —
// .moai/config, the directory the wholesale clean used to remove inline and
// the pipeline now classifies instead (REQ-UPM-020); the common-asset roots
// as PRESERVE-ONLY targets; and the harness-neutral shared surfaces.
func reconcileTargets(projectRoot string, cleanTargets []deploy.CleanTarget) []deploy.CleanTarget {
	targets := make([]deploy.CleanTarget, 0, len(cleanTargets)+len(projectCommonAssetRels)+len(template.SharedDeployTargets())+1)
	targets = append(targets, cleanTargets...)
	// The common-asset roots re-enter the classification scope as
	// PRESERVE-ONLY targets (card t1547 repair round). The de-plugin
	// absorption's clean-side exclusion (computeRunCleanTargets) removed them
	// from the CLEAN scope — correct: the per-file user-asset migration
	// (migrateProjectCommonAssets) is their only removal — but it silently
	// dropped them from the CLASSIFICATION scope too, so user-owned files
	// under them classified nowhere and vanished from the preserved listing.
	// The PreserveOnly flag routes them to user-owned at the classifier and
	// keeps every removal walk off them.
	for _, root := range projectCommonAssetRels {
		targets = append(targets, deploy.CleanTarget{
			DisplayPath:  strings.TrimSuffix(root, "/"),
			FullPath:     filepath.Join(projectRoot, filepath.FromSlash(root)),
			PreserveOnly: true,
		})
	}
	// The harness-neutral shared surfaces (.moai/policies ← .claude/rules/moai,
	// .moai/workflows ← .claude/skills/moai/workflows — the dual/codex
	// deployers project them, gate round 4 finding 1): real deploy paths the
	// managed-clean list never covered, so a user-modified policies file was
	// overwritten without a sidecar on a force run. Full classification —
	// these ARE rewritten by the deploy, so template-owned/user-modified/
	// stale all remain honest classes here.
	for _, target := range template.SharedDeployTargets() {
		targets = append(targets, deploy.CleanTarget{
			DisplayPath: target,
			FullPath:    filepath.Join(projectRoot, filepath.FromSlash(target)),
		})
	}
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

// reconcileMergeExclude composes reconcileExclude with the mergeable-file
// set — the full merge-phase exclusion the run applies (review finding 2).
func reconcileMergeExclude() func(rel string) bool {
	return func(rel string) bool {
		if reconcileExclude(rel) {
			return true
		}
		for _, m := range mergeableFilePaths() {
			if rel == m {
				return true
			}
		}
		return false
	}
}

// reconcileManagedRoots runs the preservation reconcile over the run's
// managed-root scope (deploy-mode aware, WITHOUT any migration removal
// append — gate round 10, finding 1: the migration's wholesale removal is
// exactly its classified dropped-root list, and the managed roots reconcile
// like every default run). Shared by the default path and the migration arm;
// results land in the caller's summary/pending slots, and the legacy
// .moai/memory migration runs after the reconcile's own work (gate round 10,
// finding 4 — the wholesale clean used to carry it at the end of its walk).
func reconcileManagedRoots(projectRoot string, out io.Writer, tmplFS fs.FS, mgr manifest.Manager, deployMode template.DeployMode, summary *update.ReconciliationSummary, pending *[]update.PendingMerge) error {
	// The reconciliation is read-only until the stale archive; a
	// classification failure aborts with the tree intact (NFR-UPM-002
	// stage 1). The exclude set names the paths another step of THIS flow
	// already reconciles — the config sections restore below and the
	// mergeable-file merge — so the merge phase never reprocesses them
	// (review finding 2: reprocessing would diff the operator's pre-deploy
	// bytes against the other step's merged output and revert its delivered
	// updates).
	s, p, err := update.ReconcileManagedPaths(projectRoot, out,
		tmplFS, templateRenderCarriage(tmplFS, nil, nil), mgr.Manifest(),
		update.ReconcileOptions{
			Targets: reconcileTargets(projectRoot, computeRunCleanTargets(projectRoot, deployMode, nil)),
			Exclude: reconcileMergeExclude(),
		})
	if err != nil {
		return err
	}
	*summary = s
	*pending = p
	// The legacy .moai/memory migration (finding 4): unchanged contract, now
	// carried by the reconcile path instead of the wholesale walk's tail.
	return deploy.MigrateLegacyMemoryDir(projectRoot, out)
}
