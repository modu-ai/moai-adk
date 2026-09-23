package factorymsg

import (
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// LeadPeerIdentity reports the process identity carried by a run's registered
// role='lead' peer. It is the REQ-006 fallback source for a run row written
// before the owner column existed.
//
// A launch-pending peer counts: a lead sits in that state from launch until
// its SessionStart hook binds a session UUID, and its process identity is
// exactly as good a liveness source there as it is afterwards.
func LeadPeerIdentity(projectRoot, runID string) (pid int, processStart string, ok bool) {
	path, err := BrokerPath(projectRoot, runID)
	if err != nil {
		return 0, "", false
	}
	if _, err := os.Stat(path); err != nil {
		return 0, "", false
	}
	s, err := OpenExistingWithDeadline(projectRoot, runID, 2*time.Second)
	if err != nil {
		return 0, "", false
	}
	defer func() { _ = s.Close() }()
	if err := s.db.QueryRow(`SELECT pid, process_start FROM peers WHERE role='lead' ORDER BY updated_at DESC LIMIT 1`).Scan(&pid, &processStart); err != nil {
		return 0, "", false
	}
	if pid < 1 || strings.TrimSpace(processStart) == "" {
		return 0, "", false
	}
	return pid, processStart, true
}

// LeadIdentityLookupFor adapts LeadPeerIdentity to the lookup homestate's
// reconciler takes as a parameter. The fallback crosses the package boundary
// as a function value because factorymsg imports homestate and the direction
// cannot be closed.
func LeadIdentityLookupFor(projectRoot string) homestate.LeadIdentityLookup {
	return func(runID string) (int, string, bool) { return LeadPeerIdentity(projectRoot, runID) }
}
