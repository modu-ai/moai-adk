// anchor_lock.go — SPEC-WORKTREE-REAPER-001 M2: the shared, lock-aware anchor
// decision consumed by all three sweeps that would otherwise remove a worktree
// out from under a live session.
//
// The git worktree lock is the AUTHORITATIVE anchor source (REQ-WR-006).
// Claude Code writes it at EnterWorktree time — in the same act that anchors
// the session — and releases it at ExitWorktree, so it cannot drift from the
// thing it describes. The session registry is a SUPPLEMENTARY source
// (REQ-WR-010): its `cwd` is corrected only when the CwdChanged hook calls
// RelocateSession, and that correction is measurably not running for most
// lanes. The two are UNIONED (REQ-WR-009); neither replaces the other.
//
// The decision is fail-CLOSED (REQ-WR-008) while the surrounding sweep is
// fail-open: failing open on the sweep costs a preserved worktree, failing
// open on this guard costs a live session's shell.
package session

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LockInfo is the git worktree lock state for one worktree, as carried by
// `git worktree list --porcelain`.
//
// Reason holds the STORED lock reason, not the porcelain line: git renders
// `locked <reason>` and `locked ` is git's own prefix. A reason-less lock
// renders as a bare `locked` line and yields Locked=true, Reason=""
// (design.md §B.3).
type LockInfo struct {
	Locked bool
	Reason string
}

// AnchorSource names which source produced an anchor verdict (REQ-WR-011).
type AnchorSource string

const (
	// AnchorSourceNone means no source claimed an anchor.
	AnchorSourceNone AnchorSource = "none"
	// AnchorSourceLock means the git worktree lock claimed the anchor.
	AnchorSourceLock AnchorSource = "lock"
	// AnchorSourceRegistry means the session registry claimed the anchor.
	AnchorSourceRegistry AnchorSource = "registry"
)

// AnchorVerdict is the union decision plus the source that produced it.
type AnchorVerdict struct {
	// Anchored is true when a live session is (or may be) anchored in the tree.
	Anchored bool
	// Source names which input produced the verdict.
	Source AnchorSource
	// Detail is a short human-readable qualifier for the preserve notice.
	Detail string
}

// sessionProcessLiveness is the two-valued liveness probe seam (design.md
// §B.5). It is a package variable so tests can express the undetermined case,
// which no real platform reliably produces on demand.
var sessionProcessLiveness = probeProcessLiveness

// ParseWorktreeLocks parses `git worktree list --porcelain` into a
// worktree-path → LockInfo map. Only locked entries are present in the result;
// an absent key means the lock source has NO OPINION about that worktree
// (REQ-WR-020) — it does NOT mean "not anchored".
func ParseWorktreeLocks(porcelain string) map[string]LockInfo {
	locks := make(map[string]LockInfo)
	path := ""
	for _, line := range strings.Split(porcelain, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			path = ""
		case strings.HasPrefix(line, "worktree "):
			path = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		case line == "locked":
			if path != "" {
				locks[path] = LockInfo{Locked: true}
			}
		case strings.HasPrefix(line, "locked "):
			if path != "" {
				locks[path] = LockInfo{Locked: true, Reason: strings.TrimPrefix(line, "locked ")}
			}
		}
	}
	return locks
}

// LockRefusesRemoval reports whether non-forced `git worktree remove` is KNOWN
// to refuse this worktree (REQ-WR-021 pre-detection set). git refuses a locked
// tree regardless of whether the locking process is still alive, and the
// condition never clears on its own — so attempting the removal would emit an
// error-shaped notice on every sweep, forever, for a correctly-behaving tree.
//
// The pre-detection set is deliberately NON-EXHAUSTIVE: a populated submodule
// also refuses with a clean porcelain (design.md §B.6a, EC-12). Everything
// outside this set falls through to the fail-open path, where git refuses and
// the sweep emits a cause-naming notice.
func LockRefusesRemoval(lock LockInfo) bool { return lock.Locked }

// AnchorDecision is the union of the lock source and the registry source
// (REQ-WR-009/010). The lock is consulted first because it is authoritative;
// the registry is consulted whenever the lock source does not claim an anchor,
// including the no-opinion case where the tree carries no lock at all.
func AnchorDecision(treePath string, lock LockInfo, now time.Time) AnchorVerdict {
	if anchored, detail := lockAnchorVerdict(lock); anchored {
		return AnchorVerdict{Anchored: true, Source: AnchorSourceLock, Detail: detail}
	}
	if entries := LiveAnchoredSessions(treePath, now); len(entries) > 0 {
		return AnchorVerdict{
			Anchored: true,
			Source:   AnchorSourceRegistry,
			Detail:   fmt.Sprintf("%d live registry entr%s", len(entries), plural(len(entries))),
		}
	}
	return AnchorVerdict{Anchored: false, Source: AnchorSourceNone}
}

// lockAnchorVerdict implements the design.md §B.4 fail-closed table for the
// lock source alone. "Confirmed dead" is the ONLY negative this source may
// assert, and only when the probe positively established death.
//
// The returned bool is the anchor verdict; a false verdict covers both "the
// lock has no opinion" (no lock line) and "the lock positively saw a dead
// process" — in both cases the registry still gets its say in AnchorDecision.
func lockAnchorVerdict(lock LockInfo) (bool, string) {
	if !lock.Locked {
		return false, "" // no opinion — the registry decides (REQ-WR-020)
	}
	pid, ok := parseLockPID(lock.Reason)
	if !ok {
		// No pid token, a non-integer pid, or an empty reason: the lock is
		// present but unreadable, so it is treated as anchored (REQ-WR-008).
		return true, "locked, no readable pid in the lock reason"
	}
	alive, determined := sessionProcessLiveness(pid)
	switch {
	case !determined:
		return true, fmt.Sprintf("locked by pid %d, liveness undetermined", pid)
	case alive:
		return true, fmt.Sprintf("locked by live pid %d", pid)
	default:
		return false, fmt.Sprintf("locked by pid %d, confirmed dead", pid)
	}
}

// parseLockPID extracts the integer following a `pid ` token in a stored lock
// reason (e.g. `claude session t207 (pid 36912 start Sun Aug 23 ...)`). It is
// deliberately narrow: anything it does not recognise is reported as
// unreadable, never as "unlocked".
func parseLockPID(reason string) (int, bool) {
	idx := strings.Index(reason, "pid ")
	if idx < 0 {
		return 0, false
	}
	rest := reason[idx+len("pid "):]
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	pid, err := strconv.Atoi(rest[:end])
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
