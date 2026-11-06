package cli

// update_progress.go — card t2053: a progress label for the `moai update`
// steps that used to run with no output at all (the binary update check, the
// pre-migration housekeeping, the old-backup prune). They reuse
// tui.ProgressLine, the in-place line every other long update step already
// uses, so non-TTY output stays one plain line per event with no escape
// sequences.

import (
	"io"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// runUpdateProgressStep prints label as an in-flight line, runs fn, then
// replaces the line with done.
func runUpdateProgressStep(out io.Writer, label, done string, fn func()) {
	pl := tui.ProgressLine(out, label, nil)
	fn()
	pl.Done(done)
}

// pruneBackupsWithProgress is backup.PruneBackups behind a progress label: the
// prune hashes retained file versions to decide what a run is the last copy of
// and can take a while on a project with many backups.
func pruneBackupsWithProgress(out io.Writer, projectRoot string) backup.PruneReport {
	var rep backup.PruneReport
	runUpdateProgressStep(out, "Pruning old backups...", "Old backups checked", func() {
		rep = backup.PruneBackups(projectRoot, 5, config.LoadBackupPastVersionsKeep(projectRoot))
	})
	return rep
}
