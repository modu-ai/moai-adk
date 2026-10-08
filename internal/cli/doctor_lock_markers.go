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
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// shellSingleQuoted renders p as a POSIX single-quoted shell word — the
// only quoting under which a copied path can never execute command
// substitution or glob (gate round 23: rm "<path>" fired $(...) inside a
// crafted home path). An embedded single quote closes the word, escapes
// the quote, and reopens.
func shellSingleQuoted(p string) string {
	return "'" + strings.ReplaceAll(p, "'", `'\''`) + "'"
}

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
	irregular := ""
	for _, m := range markers {
		classify := userassets.ClassifyGuardMarker
		if m.label == "lock file" {
			// Gate round 20: the .lock file takes no flock — the guard
			// marker's flock-based absence judgment must not classify it.
			classify = userassets.ClassifyLockFile
		}
		state, pid := classify(m.path)
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
		case userassets.GuardMarkerIrregular:
			if irregular == "" {
				irregular = fmt.Sprintf("%s %s", m.label, m.path)
			}
		case userassets.GuardMarkerAbsent:
			// nothing to report for this one
		}
	}

	switch {
	case irregular != "":
		// Gate round 19: a non-regular object at a marker path (a FIFO
		// would hang any read) is surfaced, never read, and needs the
		// user's explicit attention.
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("an irregular object occupies a lock marker path: %s — moai never reads it; inspect and remove it manually once you know what created it", irregular)
		return check
	case ownerless != "":
		// REQ-LOCK-001: the marker is never auto-reclaimed — the explicit
		// confirmed removal is the resolution, gated on the user's own
		// check that no moai process is running. Gate round 23: the path
		// is printed SINGLE-QUOTED with POSIX escaping — a double-quoted
		// path executes command substitution when the user copies it
		// (gate repro: `$(touch probe)` fired inside rm "...").
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("an ownerless %s (%s) cannot be proven abandoned (no pid record) and is never auto-reclaimed — after confirming no moai process is running, remove it explicitly: rm %s", ownerless, ownerlessPath, shellSingleQuoted(ownerlessPath))
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
		// Gate round 19: the reclaim CONDITION differs by marker kind —
		// the guard marker carries no age gate, but the .lock file's
		// takeover additionally requires its age to pass the stale window
		// (DefaultStaleAfter). State the condition the code actually
		// applies.
		if strings.Contains(dead, "lock file") {
			check.Message = fmt.Sprintf("a stale %s remains from a dead owner — a later moai run takes it over once the lock's age passes the stale window (%s)", dead, userassets.DefaultStaleAfter)
		} else {
			check.Message = fmt.Sprintf("a stale %s remains from a dead owner — the next moai run reclaims it automatically", dead)
		}
		return check
	default:
		check.Status = uikit.CheckOK
		check.Message = "no user lock marker present"
		return check
	}
}
