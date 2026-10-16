package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// backupRetentionReportLines renders one prune report as the lines the update
// output carries (SPEC-UPDATE-BACKUP-RETENTION-001 REQ-UBR-011, REQ-UBR-012): one
// line per removed run, one per retained excess run with its reason, a summary,
// and the unmanaged entries under .moai-backups as a count and a size. Paths are
// shown relative to projectRoot. A prune that decided nothing and found no
// unmanaged entries renders nothing.
func backupRetentionReportLines(rep backup.PruneReport, projectRoot string) []string {
	if len(rep.Items) == 0 && rep.Unmanaged.Count == 0 {
		return nil
	}
	var lines []string
	retained := 0
	for _, it := range rep.Items {
		where := retentionRunLabel(it.Dirs, projectRoot)
		switch {
		case it.Action == backup.RetentionActionRemoved:
			lines = append(lines, fmt.Sprintf("old backup run removed: %s", where))
		case it.Reason == backup.RetentionReasonProtectsVersions:
			retained++
			lines = append(lines, fmt.Sprintf("old backup run retained (%s: holds %d protected past version(s)): %s", it.Reason, it.Protects, where))
		default:
			retained++
			reason := it.Reason
			if it.Detail != "" {
				reason += ": " + it.Detail
			}
			lines = append(lines, fmt.Sprintf("old backup run retained (%s): %s", reason, where))
		}
	}
	if len(rep.Items) > 0 {
		lines = append(lines, fmt.Sprintf("backup retention: %d run(s) counted, window %d, %d older run(s) removed, %d older run(s) retained (%d distinct protected past version(s))",
			rep.Runs, rep.Window, len(rep.Items)-retained, retained, rep.Protected))
	}
	if rep.Unmanaged.Count > 0 {
		lines = append(lines, fmt.Sprintf("unmanaged entries under .moai-backups: count %d, total %d bytes, not removed", rep.Unmanaged.Count, rep.Unmanaged.Bytes))
	}
	return lines
}

// retentionRunLabel names a run by its directories, relative to projectRoot when
// they sit under it.
func retentionRunLabel(dirs []string, projectRoot string) string {
	names := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if rel, err := filepath.Rel(projectRoot, d); err == nil && !strings.HasPrefix(rel, "..") {
			d = rel
		}
		names = append(names, filepath.ToSlash(d))
	}
	return strings.Join(names, ", ")
}
