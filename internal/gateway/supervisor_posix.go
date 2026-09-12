//go:build !windows

package gateway

import (
	"os/exec"
	"syscall"
)

// Keep terminal interrupts directed at Claude's foreground process group.
func configureSupervisorProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
