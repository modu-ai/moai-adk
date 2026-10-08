// doctor_lock_markers.go — the `moai doctor` user-lock marker row
// (SPEC-USERASSET-DEPLOY-GUARD-001 M2, REQ-LOCK-001's visible-recovery
// path). A lock marker whose owner is PROVEN dead is reclaimed by the next
// acquisition automatically — nothing for the user to do. A marker with NO
// pid record is NEVER auto-reclaimed (design §3: a suspended process's
// marker has exactly that shape, and age is no proof of death) — this row
// is its user-visible resolution: name the file, state the contract, and
// give the explicit confirmed-removal procedure. Read-only, report-only,
// doctor house style.
package cli

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

func checkUserLockMarkers(homeDir string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "User Lock"}

	type marker struct {
		label string
		path  string
	}
	markers := []marker{
		{"guard marker", userassets.GuardMarkerPath(homeDir)},
		{"lock file", userassets.LockPath(homeDir)},
	}

	ownerless := ""
	ownerlessPath := ""
	alive := ""
	dead := ""
	for _, m := range markers {
		state, pid := userassets.ClassifyGuardMarker(m.path)
		switch state {
		case userassets.GuardMarkerOwnerless:
			if ownerless == "" {
				ownerless = m.label
				ownerlessPath = m.path
			}
		case userassets.GuardMarkerOwnerAlive:
			if alive == "" {
				alive = fmt.Sprintf("%s %s (pid %d)", m.label, m.path, pid)
			}
		case userassets.GuardMarkerOwnerDead:
			if dead == "" {
				dead = fmt.Sprintf("%s %s (dead pid %d)", m.label, m.path, pid)
			}
		case userassets.GuardMarkerAbsent:
			// nothing to report for this one
		}
	}

	switch {
	case ownerless != "":
		// REQ-LOCK-001: the marker is never auto-reclaimed — the explicit
		// confirmed removal is the resolution, gated on the user's own
		// check that no moai process is running.
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("an ownerless %s (%s) cannot be proven abandoned (no pid record) and is never auto-reclaimed — after confirming no moai process is running, remove it explicitly: rm \"%s\"", ownerless, ownerlessPath, ownerlessPath)
		if verbose {
			check.Detail = "the marker predates PID ownership records (or was created by a foreign tool); the guard refuses and waits rather than guessing. Removing it while a process holds the lock opens a second-writer window."
		}
		return check
	case alive != "":
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("the user lock is held by a running process: %s — wait for it to finish or remove the stale marker explicitly once you have confirmed the process is gone", alive)
		return check
	case dead != "":
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("a stale %s remains from a dead owner — the next moai run reclaims it automatically", dead)
		return check
	default:
		check.Status = uikit.CheckOK
		check.Message = "no user lock marker present"
		return check
	}
}
