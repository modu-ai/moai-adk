//go:build windows

package cli

import "os/exec"

// afGateProcessGroup is a no-op on native Windows: the gate runs with sh, which a native Windows host does not
// have (see the help text of `moai af land`), and the default cancellation kills only the direct child.
func afGateProcessGroup(*exec.Cmd) {}
