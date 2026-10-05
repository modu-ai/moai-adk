//go:build !windows

package quality

import (
	"os/exec"
	"syscall"
)

// descendantTerminationNote states what the gate did about the step's
// descendants, so a timeout reason never implies more than was actually
// attempted. On Unix the whole group is signalled.
const descendantTerminationNote = "termination was signalled to the step's whole process group, so processes the step started were signalled too"

// isolateProcessGroup puts the step in a process group of its own, so a later
// group signal reaches everything the step started and nothing else. The group
// is created per step at spawn, which is what keeps the kill from reaching
// further than intended.
func isolateProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// terminateProcessGroup signals the group led by pid. With Setpgid above, the
// step's PID is its group ID. Errors are deliberately ignored: the only
// expected one is "no such process", which means the work is already done.
func terminateProcessGroup(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
