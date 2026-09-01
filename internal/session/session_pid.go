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
// is the same PID `workers.json` already carries: the factory launcher stamps
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
//  1. MOAI_SESSION_PID, when it names a live process — the caller knew the
//     session PID outright and said so.
//  2. The nearest ancestor that is not a wrapper shell, when the platform can
//     report ancestry and that ancestor is live.
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
	if pid, ok := sessionPIDFromEnv(os.Getenv(config.EnvMoaiSessionPID)); ok {
		return pid, true
	}
	if pid := ancestorSessionPID(os.Getpid()); pid > 0 {
		return pid, true
	}
	return 0, false
}

// resolveSessionPID reports the PID to record for a session registered from
// this process: the owner PID when one resolves, else os.Getpid().
//
// The fallback is the pre-existing behavior, kept so a platform without
// ancestry support (Windows) or an unreadable process table degrades to what
// the registry recorded before rather than to nothing. A PID from the fallback
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
