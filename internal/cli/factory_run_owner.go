package cli

import (
	"context"
	"strings"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// stampFactoryRunOwner rewrites a run's owner identity to the session process
// (REQ-002b). It is the restamp seam every launch door that does NOT replace
// its launching process calls once the session identity exists.
//
// Three properties are load-bearing:
//
//   - It takes an ALREADY-RESOLVED identity. The seam does not probe, poll, or
//     branch on platform; each call site passes the identity it is in a
//     position to resolve, so all platform knowledge stays where it lives.
//   - It carries NO build tag. A restamp authored inside launch_exec_windows.go
//     would sit behind //go:build windows, where a darwin host cannot compile a
//     call to it — and the pane door that needs the identical restamp is on
//     darwin.
//   - It is idempotent against the row it names, so "which doors call it" is a
//     correctness-of-coverage question, never a correctness-of-value one.
func stampFactoryRunOwner(root, runID string, pid int, processStart string) (err error) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(runID) == "" {
		return nil
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	return db.StampRunOwner(context.Background(), runID, pid, processStart)
}

// clearFactoryRunOwner removes a run's owner identity. It is the run-state half
// of the REQ-002d refusal: a door that cannot obtain a session identity refuses
// the launch, and the run must not be left carrying the launching process's
// pid — that identity is known in advance to die, and a run holding it would be
// retired while a session it never reached was meant to be alive.
func clearFactoryRunOwner(root, runID string) (err error) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(runID) == "" {
		return nil
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	return db.ClearRunOwner(context.Background(), runID)
}
