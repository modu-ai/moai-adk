package merge

import (
	"io"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// MergeUserFilesAndSettleSnapshot is the restore half of an update flow: it
// merges the backed-up user files into the freshly deployed tree and then
// settles the .claude/settings.json snapshot the flow staged
// (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 REQ-USB-005, plan.md D4 ③).
//
// The settle runs whether or not there was anything to merge — a flow whose
// project had no settings.json still deployed a render that must become the
// next base (plan.md D5 case 8) — and it runs even when the merge returns an
// error: that error comes from loading the manifest or the embedded templates
// before any file is written, so the live file still holds the deployed
// render. Only a preserve path, which writes the pre-flow user file back,
// discards the staging copy. The merge error is returned unchanged; a settle
// failure only warns (REQ-USB-010).
//
// Flow order around this call: leftover judgement (backup.JudgeLeftoverSettingsSnapshot)
// before the flow's first rewrite → deploy → backup.StageDeployedSettingsSnapshot →
// this function.
func MergeUserFilesAndSettleSnapshot(projectRoot string, backups []FileBackup, out, warn io.Writer) error {
	var (
		outcome MergeOutcome
		err     error
	)
	if len(backups) > 0 {
		outcome, err = MergeUserFilesWithOutcome(projectRoot, backups, out)
	}
	backup.SettleSettingsSnapshot(projectRoot, outcome.Preserved(settingsJSONPath), warn)
	return err
}
