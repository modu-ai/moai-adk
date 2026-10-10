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
// parse as a run id; such a directory is a run of its own.
func readRunID(dir string) (time.Time, bool) {
	data, err := os.ReadFile(filepath.Join(dir, RunMarkerName))
	if err != nil {
		return time.Time{}, false
	}
	line, _, _ := strings.Cut(string(data), "\n")
	id, err := time.Parse(runIDLayout, strings.TrimSpace(line))
	if err != nil {
		return time.Time{}, false
	}
	return id, true
}
