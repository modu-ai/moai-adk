//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
)

func detachGPTSupervisor(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
