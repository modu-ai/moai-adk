package session

import "os"

// ProcessView is one view of the process table: the process to resolve from
// and the two queries the owner walk needs. ResolveOwnerPID runs the same walk
// over LiveProcessView(); a caller outside this package can hand the walk a
// synthetic table instead, so owner resolution is testable against an
// arbitrary tree without replacing the resolver (SPEC-HOOK-STOP-PARSE-CAP-001
// plan.md §C 6).
type ProcessView struct {
	// Self is the process the walk starts from.
	Self int
	// Info reports a process's parent PID and command name.
	Info func(pid int) (ppid int, comm string, ok bool)
	// Alive reports whether a process is alive.
	Alive func(pid int) bool
}

// LiveProcessView is the view of the real process table from this process. It
// reads the procInfo and pidIsAlive seams at call time, so ResolveOwnerPID and
// every test that swaps those seams stay on one code path.
func LiveProcessView() ProcessView {
	return ProcessView{Self: os.Getpid(), Info: procInfo, Alive: pidIsAlive}
}

// ResolveOwnerPID resolves the owning session's PID over this view, with stamp
// as the MOAI_SESSION_PID value: a live stamp that names this process or one of its
// ancestors wins, else the nearest non-wrapper ancestor, else (0, false).
func (v ProcessView) ResolveOwnerPID(stamp string) (pid int, resolved bool) {
	if pid, ok := v.sessionPIDFromStamp(stamp); ok && v.stampNamesOwnSession(pid) {
		return pid, true
	}
	if pid := v.ancestorSessionPID(v.Self); pid > 0 {
		return pid, true
	}
	return 0, false
}

func (v ProcessView) sessionPIDFromStamp(raw string) (int, bool) {
	return parseSessionPID(raw, v.Alive)
}

// stampNamesOwnSession reports whether pid plausibly names the session THIS
// process belongs to: it is this process itself, or it appears in this
// process's ancestry chain. A stamp that names neither is one this process
// inherited from an unrelated session, and honoring it is what let two
// independent sessions resolve to the same owner (card t958).
//
// Two properties are deliberate and neither is hidden:
//
// Fail-open where ancestry is UNMEASURABLE. When the view's Info reports
// nothing at all for this process — a platform that cannot read the process
// table — the stamp is accepted unchanged. That preserves today's behavior on
// those platforms rather than breaking them: the alternative is refusing every
// stamp exactly where the walk cannot answer either, which would resolve no
// owner at all. The fail-open applies only to "cannot measure", never to a
// chain that WAS measured and did not contain the stamp.
//
// KNOWN RESIDUAL: a second session that is a strict DESCENDANT of the stamping
// session still passes, because a descendant genuinely has the stamper in its
// chain and a pid alone cannot separate "I am the stamper's session" from "I
// am a different session running underneath it". The check closes the SIBLING
// case — two lanes launched from sibling shells, which is the realistic one —
// and does not close the descendant case.
func (v ProcessView) stampNamesOwnSession(pid int) bool {
	if pid == v.Self {
		return true
	}
	if _, _, ok := v.Info(v.Self); !ok {
		return true
	}
	cur := v.Self
	for depth := 0; depth < maxAncestryDepth; depth++ {
		ppid, _, ok := v.Info(cur)
		if !ok || ppid <= 1 {
			return false
		}
		if ppid == pid {
			return true
		}
		cur = ppid
	}
	return false
}

// ancestorSessionPID walks up from start and returns the first ancestor that
// is not a wrapper shell, or 0 when the ancestry cannot be resolved. PID 1 and
// below are never returned: reaching init means the walk ran past the session
// rather than finding it.
func (v ProcessView) ancestorSessionPID(start int) int {
	pid := start
	for depth := 0; depth < maxAncestryDepth; depth++ {
		ppid, _, ok := v.Info(pid)
		if !ok || ppid <= 1 {
			return 0
		}
		_, comm, ok := v.Info(ppid)
		if !ok {
			return 0
		}
		if wrapperProcessNames[comm] {
			pid = ppid
			continue
		}
		if !v.Alive(ppid) {
			return 0
		}
		return ppid
	}
	return 0
}
