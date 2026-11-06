//go:build windows

package cli

import "os/exec"

// Windows has no POSIX process groups to signal: the child keeps its previous
// behaviour (close waits for the direct child and kills it after the bound).

func isolateCodexProcessGroup(*exec.Cmd) {}

func codexGroupAlive(int) bool { return false }

func signalCodexGroup(int, bool) {}
