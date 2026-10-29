package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Issue #1948 (CLI package cleanup failure). A codex app-server starts helpers of
// its own (a plugin clone writes under CODEX_HOME) and the session's close used to
// wait for the direct child only, so a helper could still be writing when the
// caller removed the directory it wrote to. The session now runs in a process
// group of its own (pgid = the child's pid, so the group is ours and only ours),
// and close does not return while that group has members, within a fixed bound.
//
// A group of its own is also a group the terminal's Ctrl-C no longer reaches, so
// only an exit path that runs close (or shutdownCodexSessions) ends it. Every
// entry that can open a session therefore carries a signal handler that turns
// SIGINT/SIGTERM into a cancelled context (codexSignalContext), and the MCP
// server's shutdown pass reaps whatever is still tracked.

const (
	// codexGroupTermWait is how long close waits for the group to empty after
	// SIGTERM, and codexGroupKillWait how long after SIGKILL. Together they cap the
	// time close adds once the direct child is gone.
	codexGroupTermWait = time.Second
	codexGroupKillWait = time.Second

	codexGroupPoll = 10 * time.Millisecond
)

// codexGroups tracks the process groups of the sessions this process started
// and has not closed yet, so a shutdown pass can end them without a handle.
var codexGroups = struct {
	sync.Mutex
	pgids map[int]struct{}
}{pgids: map[int]struct{}{}}

func trackCodexGroup(pgid int) {
	if pgid <= 0 {
		return
	}
	codexGroups.Lock()
	codexGroups.pgids[pgid] = struct{}{}
	codexGroups.Unlock()
}

func untrackCodexGroup(pgid int) {
	codexGroups.Lock()
	delete(codexGroups.pgids, pgid)
	codexGroups.Unlock()
}

// endCodexSessionGroup ends the session's group and forgets it.
func endCodexSessionGroup(pgid int) {
	reapCodexGroup(pgid)
	untrackCodexGroup(pgid)
}

// shutdownCodexSessions ends every session group still tracked. It is the exit
// path of a process that may leave while a session is open; the groups are
// ended concurrently, so the whole pass stays within one group's bound.
func shutdownCodexSessions() {
	codexGroups.Lock()
	ids := make([]int, 0, len(codexGroups.pgids))
	for id := range codexGroups.pgids {
		ids = append(ids, id)
	}
	codexGroups.Unlock()
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			endCodexSessionGroup(id)
		}()
	}
	wg.Wait()
}

// codexSignalContext derives a context that SIGINT or SIGTERM cancels. A session
// in a group of its own gets no Ctrl-C from the terminal, so the entry that
// opened it must turn the signal into a return that runs close.
func codexSignalContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}

// reapCodexGroup ends whatever is still running in the session's process group.
// pgid is the pid of the child close started; a zero or negative value (a child
// that never started, or a platform without groups) reaps nothing. Only the
// group created for this session is signalled — never a name or a pattern.
func reapCodexGroup(pgid int) {
	if !codexGroupAlive(pgid) {
		return
	}
	signalCodexGroup(pgid, false)
	if waitCodexGroupGone(pgid, codexGroupTermWait) {
		return
	}
	signalCodexGroup(pgid, true)
	waitCodexGroupGone(pgid, codexGroupKillWait)
}

// waitCodexGroupGone polls until the group has no member or d has passed.
func waitCodexGroupGone(pgid int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for codexGroupAlive(pgid) {
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(codexGroupPoll)
	}
	return true
}
