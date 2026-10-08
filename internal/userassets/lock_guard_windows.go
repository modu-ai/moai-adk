//go:build windows

package userassets

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// On Windows, CreateFile-based byte-range locks need a handle dance the
// stdlib does not expose cleanly; the guard degrades to a short-lived
// marker file whose removal the next acquirer retries around. The O_EXCL
// create of the lock itself is already atomic on Windows, so the guard only
// needs to bound the reclaim window, which the retry loop tolerates.
// Item 3 (fix round 3): the handle is returned to its own call (no package
// global); item 4: the marker acquisition respects the same deadline.
//
// M2 (SPEC-USERASSET-DEPLOY-GUARD-001, REQ-LOCK-001/002): the marker
// records its owner's PID — the same ownership posture the .lock file
// carries — and a marker whose owner is PROVEN dead (pid record + death
// check, lockOwnerGone) is reclaimed by the next acquirer. An ownerless
// marker (no pid record) is NEVER auto-reclaimed: it surfaces through the
// user-visible recovery path (doctor report + explicit confirmed removal),
// because a suspended process's marker has exactly this shape and age is
// no proof of death (design §3, gate-measured release-deletes-B repro).

func acquireGuard(path string, timeout time.Duration) (func(), error) {
	marker := path + ".guard"
	deadline := time.Now().Add(timeout)
	var fd *os.File
	var err error
	for {
		fd, err = os.OpenFile(marker, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			// REQ-LOCK-001: stamp the caller's PID — the .lock posture.
			// The record is the death evidence a later acquirer needs.
			_, _ = fmt.Fprintf(fd, "pid=%d acquired=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339))
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		// Held. Reclaim ONLY a death-proven marker — reclaimGuardMarker
		// re-proves the owner itself and fails closed; rename(2) makes
		// exactly one racing acquirer the winner and the losers re-run
		// the create.
		if reclaimGuardMarker(marker) {
			continue
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
	}
	release := func() {
		_ = fd.Close()
		_ = os.Remove(marker)
	}
	return release, nil
}
