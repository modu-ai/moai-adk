// doctor_ccversion.go — the doctor session-staleness check
// (SPEC-SESSION-CC-VERSION-001 REQ-SCV-007).
//
// Per live registry session, the check reads the running Claude Code version
// (one probe — the resolver's liveness gate never probes a dead pid) against
// the installed one, and warns naming both when a session runs older. The
// check is advisory in the checkFlagSlot pattern: read-only, and it never
// returns CheckFail, so it never changes moai doctor's exit status. A version
// read that degraded to unknown is not judged — staleness cannot be read
// from unknown, and warning there would nag npm-style installs with
// unversioned binary paths. Nothing here restarts, kills, blocks, or nags
// beyond the one doctor row; the restart decision stays with the
// leader/operator at a card boundary (spec §C.3).

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/session"
)

// ccVersionStalenessCheckName is the doctor --check filter name for this item.
const ccVersionStalenessCheckName = "Session CC Version"

// doctorCCVersionEntries reads the active-session registry. Seam: the tests
// substitute a fixture set without touching the real registry.
var doctorCCVersionEntries = func() ([]session.Entry, error) {
	return session.QueryActiveWork("")
}

// checkSessionCCVersionStaleness reports, per live registry session, the
// running version against the installed one.
func checkSessionCCVersionStaleness(_ string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: ccVersionStalenessCheckName}

	entries, err := doctorCCVersionEntries()
	if err != nil {
		// Advisory: a registry read failure is diagnosed by the session
		// registry's own surfaces; here it only means staleness is not judged.
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("session registry unreadable — staleness not judged: %v", err)
		return check
	}
	if len(entries) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "no active sessions registered — nothing to compare"
		return check
	}

	lines := make([]string, 0, len(entries))
	behind := 0
	unjudged := 0
	oldestRunning, installed := "", ""
	for _, e := range entries {
		// One resolver call per entry: at most one probe per live entry
		// (REQ-SCV-007). A dead pid is never probed — the resolver's
		// liveness gate short-circuits before the platform read.
		view := session.ResolveCCVersions(e.PID)
		switch {
		case view.Running == session.UnknownCCVersion || view.Installed == session.UnknownCCVersion:
			unjudged++
			lines = append(lines, fmt.Sprintf("session %s: running %s, installed %s — staleness unknown, not judged",
				shortID(e.SessionID), view.Running, view.Installed))
		case ccVersionOlder(view.Running, view.Installed):
			behind++
			if oldestRunning == "" || ccVersionOlder(view.Running, oldestRunning) {
				oldestRunning, installed = view.Running, view.Installed
			}
			lines = append(lines, fmt.Sprintf("session %s: running %s, older than installed %s",
				shortID(e.SessionID), view.Running, view.Installed))
		default:
			lines = append(lines, fmt.Sprintf("session %s: running %s (installed %s)",
				shortID(e.SessionID), view.Running, view.Installed))
		}
	}

	if behind > 0 {
		check.Status = uikit.CheckWarn
		if behind == 1 {
			check.Message = fmt.Sprintf("a session runs Claude Code %s but the installed version is %s — a running process never picks up an updated binary; restart it at a card boundary",
				oldestRunning, installed)
		} else {
			check.Message = fmt.Sprintf("%d of %d sessions run an older Claude Code (oldest %s, installed %s) — a running process never picks up an updated binary; restart at a card boundary",
				behind, len(entries), oldestRunning, installed)
		}
		check.Detail = strings.Join(lines, "\n")
		return check
	}

	check.Status = uikit.CheckOK
	if unjudged > 0 {
		check.Message = fmt.Sprintf("no session behind; %d version read(s) unknown — staleness not judged for those", unjudged)
	} else {
		check.Message = fmt.Sprintf("%d session(s) run the installed version", len(entries))
	}
	if verbose {
		check.Detail = strings.Join(lines, "\n")
	}
	return check
}

// ccVersionOlder reports whether the running version is older than the
// installed one. The version parser only ever produces digits and dots, so
// the compare is numeric per segment — 2.1.9 is older than 2.1.288 — with a
// shorter prefix reading as older (2.1 < 2.1.1). A non-numeric segment
// (impossible from the parser, defended anyway) reads as not older:
// staleness must not be judged from an unparseable value.
func ccVersionOlder(running, installed string) bool {
	runningSegs := strings.Split(running, ".")
	installedSegs := strings.Split(installed, ".")
	for i := 0; i < len(runningSegs) && i < len(installedSegs); i++ {
		r, rerr := strconv.Atoi(runningSegs[i])
		n, nerr := strconv.Atoi(installedSegs[i])
		if rerr != nil || nerr != nil {
			return false
		}
		if r != n {
			return r < n
		}
	}
	return len(runningSegs) < len(installedSegs)
}
