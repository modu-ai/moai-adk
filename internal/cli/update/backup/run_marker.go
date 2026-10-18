package backup

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// RunMarkerName is the file one update run writes at the root of every backup
// directory it creates. Its single line is the run id (NewRunID). Directories
// that carry the same run id belong to one run, so retention counts the run and
// not each directory (SPEC-UPDATE-REPAIR-RC30-001 D9, REQ-RC30-010). The marker
// is bookkeeping, not a user file: a restore never copies it into the project.
const RunMarkerName = ".moai-run"

// PriorManifestName is the file the update's Backup stage writes at the root of
// the run backup directory: the project manifest as the run loaded it, byte for
// byte (SPEC-UPDATE-OWNERSHIP-REPAIR-001). Like the run marker it is backup
// bookkeeping, not a user file: a restore never copies it into the project. The
// update code names the file through this constant, so the writer and the
// restore walks cannot drift apart.
const PriorManifestName = "manifest.prior.json"

// PriorManifestTempPrefix is the prefix of the temporary file the update code
// creates in the run backup directory while it writes manifest.prior.json (it
// reaches the final name by renaming that file). It is bookkeeping exactly like
// the final name, and a failed write may leave it behind, so a restore never
// copies it either. The writer builds the file's name from this constant.
const PriorManifestTempPrefix = PriorManifestName + ".tmp-"

// IsPriorManifestBookkeeping reports whether relPath, a path relative to a run
// backup directory, names the prior-manifest copy or its temporary file at the
// backup root. A same-named file below the root is user content and is not.
func IsPriorManifestBookkeeping(relPath string) bool {
	return relPath == PriorManifestName || strings.HasPrefix(relPath, PriorManifestTempPrefix)
}

// runIDLayout is the time layout of a run id: UTC, nanosecond precision, and
// fixed width, so lexical order equals time order.
const runIDLayout = "2006-01-02T15:04:05.000000000Z"

// NewRunID returns the run id of the update run that starts now.
func NewRunID() string {
	return time.Now().UTC().Format(runIDLayout)
}

// WriteRunMarker records runID as the run marker of the backup directory dir.
func WriteRunMarker(dir, runID string) error {
	return os.WriteFile(filepath.Join(dir, RunMarkerName), []byte(runID+"\n"), defs.FilePerm)
}

// readRunID returns the time of the run id recorded in dir's run marker. ok is
// false when dir has no readable marker or the marker's first line does not
// parse as a run id; such a directory is a run of its own. A marker that is not a
// regular file reads as no marker and never blocks the read (card t1619).
func readRunID(dir string) (time.Time, bool) {
	data, ok := readRunMarker(filepath.Join(dir, RunMarkerName))
	if !ok {
		return time.Time{}, false
	}
	line, _, _ := strings.Cut(string(data), "\n")
	id, err := time.Parse(runIDLayout, strings.TrimSpace(line))
	if err != nil {
		return time.Time{}, false
	}
	return id, true
}
