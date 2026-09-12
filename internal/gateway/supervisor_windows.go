//go:build windows

package gateway

import (
	"os/exec"
	"syscall"
)

// The retained launcher directly supervises the child and its private stop pipe.
func configureSupervisorProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
