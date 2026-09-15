//go:build windows

package codexapp

import (
	"os"
	"os/exec"
)

func inheritSharedLease(cmd *exec.Cmd, lease *os.File) {}
