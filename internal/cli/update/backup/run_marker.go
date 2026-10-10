package backup

// run_marker.go — the run marker (SPEC-UPDATE-REPAIR-RC30-001 D9, REQ-RC30-010).
// One update run writes its run id into a .moai-run marker at the root of each
// backup directory it creates, so retention counts the run and not each
// directory. The marker is bookkeeping: a restore never copies it.

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// RunMarkerName is the bookkeeping file a run writes at the root of each backup
// directory it creates.
const RunMarkerName = ".moai-run"

// runIDLayout is the layout of a run id: UTC, nanosecond precision, and fixed
// width, so the lexical order of run ids is their time order.
const runIDLayout = "2006-01-02T15:04:05.000000000Z"

// NewRunID returns the run id of the update run that starts now.
func NewRunID() string {
	return time.Now().UTC().Format(runIDLayout)
}

// WriteRunMarker records runID as the run marker of the backup directory dir.
func WriteRunMarker(dir, runID string) error {
	return os.WriteFile(filepath.Join(dir, RunMarkerName), []byte(runID+"\n"), defs.FilePerm)
}

// readRunMarker returns the run id recorded in dir's marker and its time. ok is
// false when dir has no readable marker or the marker's first line does not
// parse as a run id; such a directory is a run of its own.
func readRunMarker(dir string) (id string, when time.Time, ok bool) {
	data, err := os.ReadFile(filepath.Join(dir, RunMarkerName))
	if err != nil {
		return "", time.Time{}, false
	}
	line, _, _ := strings.Cut(string(data), "\n")
	id = strings.TrimSpace(line)
	when, err = time.Parse(runIDLayout, id)
	if err != nil {
		return "", time.Time{}, false
	}
	return id, when, true
}
