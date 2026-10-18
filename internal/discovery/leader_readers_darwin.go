//go:build darwin

// Darwin candidate readers (design.md §E): `ps` for argv and the environment
// of same-uid processes (`ps eww` appends the environment to the command
// column), a direct proc_info call for the working directory (t1807 replaced
// the `lsof -d cwd` subprocess, whose cap load exceeded). Every reader declines (false) rather than guessing: a decline
// downstream is an honest NO_ACTIVE_FACTORY, never a wrong join.
package discovery

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/modu-ai/moai-adk/internal/config"
)

// readerDeadline bounds one platform reader invocation (a `ps` or `lsof`
// subprocess).
const readerDeadline = 3 * time.Second

// readerContext bounds one platform reader invocation.
func readerContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, readerDeadline)
}

// platformArgv reads the candidate's command line.
func platformArgv(ctx context.Context, pid int) ([]string, bool) {
	ctx, cancel := readerContext(ctx)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return nil, false
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) == 0 {
		return nil, false
	}
	return fields, true
}

// platformEnv reads the candidate's environment. `ps eww` renders the
// environment appended after the command in the COMMAND column, so the
// specific key is scanned out of the token stream: an argv never carries
// MOAI_KANBAN_ID, and a same-uid process is required for the environment to
// appear at all — a foreign or unreadable process simply declines here.
func platformEnv(ctx context.Context, pid int) (map[string]string, bool) {
	ctx, cancel := readerContext(ctx)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "eww", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil {
		return nil, false
	}
	env := map[string]string{}
	for _, token := range strings.Fields(string(out)) {
		if key, value, found := strings.Cut(token, "="); found && key == config.EnvFactoryRunID {
			env[key] = value
		}
	}
	if env[config.EnvFactoryRunID] == "" {
		return nil, false
	}
	return env, true
}

// proc_info(PROC_INFO_CALL_PIDINFO, pid, PROC_PIDVNODEPATHINFO, ...) fills a
// struct proc_vnodepathinfo: two vnode_info_path records (cwd, then root), each
// a 152-byte vnode_info followed by a MAXPATHLEN path. The cwd path starts at
// offset 152 (xnu bsd/sys/proc_info.h; layout stable across macOS releases).
const (
	procInfoCallPidInfo  = 2
	procPidVnodePathInfo = 9
	vnodeInfoSize        = 152
	maxPathLen           = 1024
)

// platformCwd reads the candidate's working directory with one kernel call, the
// way /proc is read on Linux — no `lsof` subprocess, whose 3 s cap was exceeded
// under load and turned a live leader into verified=0 (card t1807, t1772).
// A same-uid candidate is readable; a foreign or exited pid declines.
//
// @MX:NOTE: [AUTO] raw proc_info syscall because release builds are CGO-free
// (libproc's proc_pidinfo is a libSystem wrapper over this same call).
func platformCwd(_ context.Context, pid int) (string, bool) {
	buf := make([]byte, 2*(vnodeInfoSize+maxPathLen))
	n, _, errno := unix.Syscall6(unix.SYS_PROC_INFO, procInfoCallPidInfo, uintptr(pid), procPidVnodePathInfo, 0,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if errno != 0 || int(n) < vnodeInfoSize+maxPathLen {
		return "", false
	}
	path := buf[vnodeInfoSize : vnodeInfoSize+maxPathLen]
	if end := bytes.IndexByte(path, 0); end >= 0 {
		path = path[:end]
	}
	if len(path) == 0 {
		return "", false
	}
	return string(path), true
}
