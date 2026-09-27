package factorymsg

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// LeadPeerIdentity reports the process identity carried by a run's registered
// leader peer. It is the REQ-006 fallback source for a run row written before
// the owner column existed.
//
// The role='leader' peer is read primarily. A peer recorded under the LEGACY
// role 'lead' (a pre-rename binary, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-024)
// is read as identity evidence too — classification of the run's owner only;
// no other reader treats that peer as the leader, and the row is never
// rewritten.
//
// A launch-pending peer counts: a leader sits in that state from launch until
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
	// "leader" is the current vocabulary; "lead" is the legacy role, read as
	// identity evidence only (REQ-RNC-024).
	for _, role := range []string{"leader", "lead"} {
		var pid int
		var start string
		if err := s.db.QueryRow(`SELECT pid, process_start FROM peers WHERE role=? ORDER BY updated_at DESC LIMIT 1`, role).Scan(&pid, &start); err != nil {
			continue
		}
		if pid < 1 || strings.TrimSpace(start) == "" {
			continue
		}
		return pid, start, true
	}
	return 0, "", false
}

// LiveLegacyPeer reports whether runID's broker database holds a peer row
// recorded under the LEGACY role/slot vocabulary (lead / worker / agent and
// their numbered labels) whose owning process is still live. The first live
// legacy value found is returned; a dead legacy row is stale exactly like a
// dead new-vocabulary row and blocks nothing (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-022). Liveness reuses the run-owner classification discipline:
// pid and process-start fingerprint must both match a live process.
func LiveLegacyPeer(ctx context.Context, projectRoot, runID string) (value string, live bool, err error) {
	path, err := BrokerPath(projectRoot, runID)
	if err != nil {
		return "", false, err
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	s, err := OpenExistingWithDeadline(projectRoot, runID, 2*time.Second)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = s.Close() }()
	rows, err := s.db.QueryContext(ctx, `SELECT role, slot, pid, process_start FROM peers`)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var role, slot, start string
		var pid int
		if err := rows.Scan(&role, &slot, &pid, &start); err != nil {
			return "", false, err
		}
		isLegacyPeer := role == "lead" || kanban.IsLegacyLeadLabel(slot) ||
			kanban.IsLegacyFactoryRoleValue(role) || kanban.IsLegacyFactoryRoleValue(slot)
		if !isLegacyPeer {
			continue
		}
		if homestate.ClassifyOwnerWith(homestate.ProbeProcessIdentity, pid, start) == homestate.OwnerLive {
			if slot != "" && slot != role {
				return slot, true, nil
			}
			return role, true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", false, err
	}
	return "", false, nil
}

// LeadIdentityLookupFor adapts LeadPeerIdentity to the lookup homestate's
// reconciler takes as a parameter. The fallback crosses the package boundary
// as a function value because factorymsg imports homestate and the direction
// cannot be closed.
func LeadIdentityLookupFor(projectRoot string) homestate.LeadIdentityLookup {
	return func(runID string) (int, string, bool) { return LeadPeerIdentity(projectRoot, runID) }
}

// LeadRecordAbsentFor reports that a run has no broker database at all, so no
// lead-peer record can exist. Any stat outcome other than "does not exist"
// answers false: a broker the fallback could not read is not an absent one.
func LeadRecordAbsentFor(projectRoot string) func(runID string) bool {
	return func(runID string) bool {
		path, err := BrokerPath(projectRoot, runID)
		if err != nil {
			return false
		}
		_, err = os.Stat(path)
		return errors.Is(err, fs.ErrNotExist)
	}
}

// ReconcileOptionsFor is the option set every production retirement path
// passes: the lead-peer fallback plus both premises of the boot proof.
func ReconcileOptionsFor(projectRoot string) homestate.ReconcileOptions {
	return homestate.ReconcileOptions{
		Fallback:         LeadIdentityLookupFor(projectRoot),
		BootTime:         homestate.SystemBootTime,
		LeadRecordAbsent: LeadRecordAbsentFor(projectRoot),
	}
}
