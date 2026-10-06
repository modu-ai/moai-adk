// Package harness — retention spawn gate (SPEC-HARNESS-DETACHED-PRUNE-001).
// REQ-DP-002: after the event append, one lock-free stamp read decides whether
// the detached prune child spawns; the once-per-interval gate is preserved,
// re-aimed from "prune attempt" to "spawn".
package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// MaybeSpawnRetentionPruner is the spawn gate between the hook record path and
// the detached prune child. It reads the prune stamp exactly once through the
// lock-free stamp read (readStampFile/stampIsFresh, the same pair the prune's
// pre-lock check uses) and, only when the stamp is stale or absent, calls spawn
// once to launch the detached child; a fresh stamp spawns nothing. The interval
// stays pruneSkipDuration (1 hour); a spawn error is returned to the caller,
// which logs it fail-open at exit 0 (REQ-DP-004).
//
// The child argv is the hidden `moai hook retention-prune` verb of the same
// binary (REQ-DP-003, spec D1/D3): `hook retention-prune --log <logPath>
// --archive <archiveDir> --days <DefaultRetentionDays>`. The archive directory
// is derived from the log path with the house convention every hook handler
// uses (<dir of log>/learning-history/archive). spawn must be non-nil; the
// production caller passes the platform real spawn (the build-tag implementation),
// tests pass recording fakes as parameters (REQ-DP-007).
//
// @MX:ANCHOR: [AUTO] MaybeSpawnRetentionPruner is the single spawn gate all four hook observe handlers reach.
// @MX:REASON: [AUTO] fan_in >= 3: recordHarnessEventWithGate wrapper called by runHarnessObserve, runHarnessObserveStop, runHarnessObserveSubagentStop, runHarnessObserveUserPromptSubmit
func MaybeSpawnRetentionPruner(logPath string, spawn func(executable string, args []string) error) error {
	now := time.Now()
	if stampIsFresh(readStampFile(logPath+pruneStateSuffix), now) {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("retention: cannot locate the moai binary for the prune child: %w", err)
	}
	archiveDir := filepath.Join(filepath.Dir(logPath), "learning-history", "archive")
	args := []string{
		"hook", "retention-prune",
		"--log", logPath,
		"--archive", archiveDir,
		"--days", strconv.Itoa(DefaultRetentionDays),
	}
	return spawn(exe, args)
}
