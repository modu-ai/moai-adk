// session_pid.go — resolves the PID the coordination registry records.
//
// The registry is written from a hook subprocess (`moai hook session-start`
// → Registry.Register). That subprocess exits within milliseconds, so its own
// PID is dead by the time any reader probes it. Every liveness probe over the
// registry — the stale-registry caveat in the Pre-Edit Sync Check, and the
// anchor guard's own `isProcessAlive` — therefore judged every live session
// dead, which reads as "no concurrent session, safe to proceed" exactly when
// isolation is required.
//
// The registry needs the PID of the long-lived session process instead. That
// is the same PID the factory roster in `factory.db` already carries: the factory launcher stamps
// `os.Getpid()` and then `syscall.Exec`s into Claude Code, so the launcher's
// PID *becomes* the session's. A hook subprocess has no such luck — it must
// look up the ancestry it was spawned from.
package session

import (
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// maxAncestryDepth bounds the parent walk. The real chain is at most
// session → wrapper shell → moai (depth 2); the bound is generous enough for
// an extra layer of shell indirection and small enough that a cyclic or
// malformed /proc view cannot spin.
const maxAncestryDepth = 8

// wrapperProcessNames are the process names the walk steps THROUGH rather than
// stopping at: the hook wrapper is a shell script, and depending on whether the
// runtime's `sh -c` and the wrapper's own `exec` collapse, zero or more shells
// sit between the session and the moai binary. None of them is the session.
//
// Names are compared against the OS-reported command name, which several
// platforms truncate (Linux comm is 15 chars, the BSD/darwin P_comm field 16),
// so keep every entry short enough to survive truncation.
var wrapperProcessNames = map[string]bool{
	"sh": true, "bash": true, "zsh": true, "dash": true,
	"ksh": true, "fish": true, "csh": true, "tcsh": true,
	"env": true, "moai": true,
	// Windows hook runners wrap the moai binary in cmd.exe (`cmd /c`) or
	// PowerShell; proc_info_windows.go strips the ".exe" suffix, so these
	// match the same lowercase bare-name keys.
	"cmd": true, "powershell": true, "pwsh": true,
}

// procInfoFunc reports a process's parent PID and command name. It is a package
// var so the ancestry walk can be tested against a synthetic process tree
// without spawning anything — the same test seam `alive` uses in
// internal/kanban/factory_slots.go.
type procInfoFunc func(pid int) (ppid int, comm string, ok bool)

var (
	procInfo   procInfoFunc = platformProcInfo
	pidIsAlive              = isProcessAlive
	// probeLiveness is the three-valued liveness probe seam for the
	// registry's liveness-aware purge: (alive, determined), where determined
	// false means the platform measured nothing and the caller falls back to
	// its own verdict (see Registry.Purge).
	probeLiveness = probeProcessLiveness
)

// ResolveOwnerPID reports the PID of the long-lived session that owns this
// process, and whether it could be resolved at all. Resolution order:
//
//  1. MOAI_SESSION_PID, when it names a live process AND that process
//     plausibly names THIS process's own session — see stampNamesOwnSession.
//  2. The nearest ancestor that is not a wrapper shell, when the platform can
//     report ancestry and that ancestor is live.
//
// The qualifier on step 1 is card t958. An environment variable is INHERITED:
// two independent sessions launched from the same shell both carry the value
// the first one stamped, so an unconditional override collapsed their distinct
// ancestries onto one pid, and every pid-keyed ownership check downstream read
// them as one owner. (Attribution: recording the owning-session pid predates
// this, from commit 3f3465369 (card t298); commit ea939d9a7 (card t951) added
// the integration lock's releasableBy, making pid a second key on the release
// holder judgment and thereby WIDENING the exposure surface of that
// pre-existing acquire-era property — it is not a t951 regression.)
//
// There is deliberately no os.Getpid() third step. This is the seam for
// callers whose record outlives the process that writes it, and for them
// os.Getpid() is not a degraded answer but a wrong one: it names a process
// that is dead by the time any reader probes it, so every such record reads as
// abandoned the instant it is written. A caller that cannot resolve an owner
// gets (0, false) and decides for itself what an unknown owner means — see
// resolveSessionPID below for the registry's answer, and
// internal/kanban.IntegrationLock for the integration window's.
func ResolveOwnerPID() (pid int, resolved bool) {
	if pid, ok := sessionPIDFromEnv(os.Getenv(config.EnvMoaiSessionPID)); ok && stampNamesOwnSession(pid) {
		return pid, true
	}
	if pid := ancestorSessionPID(os.Getpid()); pid > 0 {
		return pid, true
	}
	return 0, false
}

// stampNamesOwnSession reports whether pid plausibly names the session THIS
// process belongs to: it is this process itself, or it appears in this
// process's ancestry chain. A stamp that names neither is one this process
// inherited from an unrelated session, and honoring it is what let two
// independent sessions resolve to the same owner (card t958).
//
// Two properties are deliberate and neither is hidden:
//
// Fail-open where ancestry is UNMEASURABLE. When the procInfo seam reports
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
func stampNamesOwnSession(pid int) bool {
	self := os.Getpid()
	if pid == self {
		return true
	}
	// Unmeasurable ancestry: accept, per the fail-open above. Measured but
	// empty is a different thing and is not this branch — procInfo reporting
	// ok=false for self is the only shape that reaches it.
	if _, _, ok := procInfo(self); !ok {
		return true
	}
	// Walk the same chain ancestorSessionPID walks, under the same depth
	// bound, and look for the stamp anywhere in it. Every ancestor counts,
	// wrapper-named or not — that is deliberately at least as permissive as
	// the resolver's own walk, so this check can only ever ignore a stamp the
	// resolver would have had no reason to trust.
	cur := self
	for depth := 0; depth < maxAncestryDepth; depth++ {
		ppid, _, ok := procInfo(cur)
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

// resolveSessionPID reports the PID to record for a session registered from
// this process: the owner PID when one resolves, else os.Getpid().
//
// The fallback is the pre-existing behavior, kept so an unreadable process
// table or a genuinely unsupported platform degrades to what the registry
// recorded before rather than to nothing — every major platform now reports
// ancestry (Linux via /proc, BSD/darwin via sysctl, Windows via Toolhelp32).
// A PID from the fallback
// may still be an ephemeral hook subprocess. That is a known residual limit,
// not a silent one: it is reached only where the ancestry is genuinely
// unavailable, and it belongs to the registry's cost profile alone — a caller
// that must not inherit it uses ResolveOwnerPID directly.
func resolveSessionPID() int {
	if pid, ok := ResolveOwnerPID(); ok {
		return pid
	}
	return os.Getpid()
}

// staleEntryDead reports whether a heartbeat-stale registry entry's session
// is safe to remove. Removable when the recorded PID is positively dead
// (ESRCH). An entry with no usable PID carries no liveness signal, and a
// platform that cannot determine liveness (the Windows probe measures
// nothing) both fall back to the heartbeat verdict that governed purges
// before the liveness probe existed — preserving the existing stale-entry
// hygiene on those paths. The asymmetry is deliberate: a wrongly kept entry
// costs one harmless "other session active" notice, while a wrongly purged
// live session hides a concurrent writer from the orchestrator's race
// checks.
func staleEntryDead(pid int) bool {
	if pid <= 0 {
		return true
	}
	alive, determined := probeLiveness(pid)
	if determined {
		return !alive
	}
	return true
}

// sessionPIDFromEnv parses an explicit PID override, accepting it only when it
// names a process that is actually alive. A stale value left over in an
// inherited environment is rejected rather than recorded.
func sessionPIDFromEnv(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	pid, err := strconv.Atoi(raw)
	if err != nil || pid <= 0 {
		return 0, false
	}
	if !pidIsAlive(pid) {
		return 0, false
	}
	return pid, true
}

// ancestorSessionPID walks up from start and returns the first ancestor that is
// not a wrapper shell, or 0 when the ancestry cannot be resolved. PID 1 and
// below are never returned: reaching init means the walk ran past the session
// rather than finding it.
func ancestorSessionPID(start int) int {
	pid := start
	for depth := 0; depth < maxAncestryDepth; depth++ {
		ppid, _, ok := procInfo(pid)
		if !ok || ppid <= 1 {
			return 0
		}
		_, comm, ok := procInfo(ppid)
		if !ok {
			return 0
		}
		if wrapperProcessNames[comm] {
			pid = ppid
			continue
		}
		if !pidIsAlive(ppid) {
			return 0
		}
		return ppid
	}
	return 0
}
