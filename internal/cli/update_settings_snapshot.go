package cli

import (
	"io"

	updatemerge "github.com/modu-ai/moai-adk/internal/cli/update/merge"
	"github.com/modu-ai/moai-adk/internal/core/project"
)

// update_settings_snapshot.go holds the internal/cli seams of the
// .claude/settings.json base snapshot (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001,
// plan.md D8). The lifecycle itself lives in internal/cli/update/backup and
// internal/cli/update/merge; the flows here only call it at fixed points:
//
//	① backup.JudgeLeftoverSettingsSnapshot — runUpdate before the deny-rule strip,
//	   runInit before executor.Execute
//	② backup.StageDeployedSettingsSnapshot — right after each deploy succeeds
//	③ mergeUserFilesSettlingSnapshot / backup.SettleSettingsSnapshot — flow end

// preMergeSettingsSnapshotHook runs immediately before the post-deploy merge of
// the update flows. Production leaves it a no-op; a test replaces it to observe
// the canonical and staging copies at the one moment inside a flow where both
// must hold their pre-merge values (AC-USB-005). Tests that replace it must not
// call t.Parallel().
var preMergeSettingsSnapshotHook = func(projectRoot string) {}

// applyAutonomyTierBundleFn is project.ApplyAutonomyTierBundle behind a
// package-level variable so a test can stand in a deterministic rewrite of the
// project settings.json and prove the init staging copy was taken before it
// (AC-USB-007 `init`, plan.md D8 N-06). Tests that replace it must not call
// t.Parallel().
var applyAutonomyTierBundleFn = project.ApplyAutonomyTierBundle

// mergeUserFilesSettlingSnapshot is the restore step shared by the template-sync
// and clean-reinstall flows: the observation hook, the post-deploy merge, and
// the promotion decision for the staged settings.json render. It must be called
// even when there are no backups, so a flow whose project had no settings.json
// still promotes the render it deployed (plan.md D5 case 8).
func mergeUserFilesSettlingSnapshot(projectRoot string, backups []updatemerge.FileBackup, out, warn io.Writer) error {
	preMergeSettingsSnapshotHook(projectRoot)
	return updatemerge.MergeUserFilesAndSettleSnapshot(projectRoot, backups, out, warn)
}
