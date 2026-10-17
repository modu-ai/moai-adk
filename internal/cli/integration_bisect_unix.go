//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
	"time"
)

// bisectGraceAfterTerm is how long a stopped `git bisect run` gets to exit after
// SIGTERM before the runtime kills it outright.
const bisectGraceAfterTerm = 10 * time.Second

// bisectRunPrepare puts `git bisect run` and the judged command in their own
// process group, so a terminal Ctrl-C reaches only the verb, and makes a
// cancelled context SIGTERM the whole group (git removes its lock files on
// SIGTERM; a SIGKILL could leave one behind) before the runtime kills it.
func bisectRunPrepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) }
	cmd.WaitDelay = bisectGraceAfterTerm
}
