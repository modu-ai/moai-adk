//go:build windows

package cli

import "os/exec"

func detachGPTSupervisor(cmd *exec.Cmd) {}
