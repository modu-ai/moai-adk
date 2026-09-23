package codexwiring

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// WiringLockRelPath is the wiring lock, owned by this package and distinct
// from the `moai update` lock. A caller that holds the update lock may take
// it; the reverse order (wiring lock, then update lock) is never used, so the
// two cannot deadlock.
const WiringLockRelPath = ".moai/state/codex-wiring.lock"

// malformedLockGrace bounds how long an unreadable lock file (a writer that
// died between creating and filling it) is treated as held.
const malformedLockGrace = time.Minute

// ErrWiringLockHeld is the lock-held outcome: another live or undetermined
// owner holds the wiring lock, and no wiring file was changed.
var ErrWiringLockHeld = errors.New("codex wiring lock held by another process")

// WiringLockPayload is the JSON content of the wiring lock file.
type WiringLockPayload struct {
	PID          int    `json:"pid"`
	ProcessStart string `json:"process_start"`
}

// acquireWiringLock takes the wiring lock with O_EXCL. A lock left by a dead
// process (or a process whose identity no longer matches, i.e. a reused PID)
// is cleared and taken; a live or undetermined owner yields
// ErrWiringLockHeld.
//
// @MX:WARN: [AUTO] cross-process mutual exclusion over the wiring files
// @MX:REASON: a stale-lock clear races another clearer; O_EXCL on the retry keeps at most one winner, and an undetermined owner is treated as live so a lock is never stolen on doubt
func acquireWiringLock(projectRoot string) (release func(), err error) {
	path := filepath.Join(projectRoot, filepath.FromSlash(WiringLockRelPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create wiring lock directory: %w", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			payload, _ := json.Marshal(WiringLockPayload{PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint()})
			_, werr := f.Write(payload)
			cerr := f.Close()
			if werr != nil || cerr != nil {
				_ = os.Remove(path)
				return nil, fmt.Errorf("write wiring lock: %w", errors.Join(werr, cerr))
			}
			return func() { _ = os.Remove(path) }, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("create wiring lock: %w", err)
		}
		if !wiringLockStale(path) {
			return nil, ErrWiringLockHeld
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("clear stale wiring lock: %w", err)
		}
	}
	return nil, ErrWiringLockHeld
}

// wiringLockStale reports whether the lock file at path belongs to an owner
// that is provably gone. Anything uncertain is not stale.
func wiringLockStale(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var p WiringLockPayload
	if err := json.Unmarshal(raw, &p); err != nil || p.PID <= 0 {
		info, serr := os.Stat(path)
		return serr == nil && time.Since(info.ModTime()) > malformedLockGrace
	}
	fp, state := homestate.ProbeProcessIdentity(p.PID)
	switch state {
	case homestate.ProcessIdentityDead:
		return true
	case homestate.ProcessIdentityLive:
		return p.ProcessStart != "" && fp != p.ProcessStart
	default:
		return false
	}
}
