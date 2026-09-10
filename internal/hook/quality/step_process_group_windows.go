//go:build windows

package quality

import "os/exec"

// descendantTerminationNote states what the gate did about the step's
// descendants. Windows has no process-group primitive of the kind used on
// Unix, and a job object — which would end the tree properly — is a second
// mechanism with its own failure modes and is deliberately out of this scope.
// So the reason says what is true here: the wait is bounded, and descendants
// may still be running.
const descendantTerminationNote = "this platform provides no process-group primitive, so descendants of the step may have survived"

// isolateProcessGroup is a no-op on Windows. stepWaitGrace still bounds the
// return; only the group signal is unavailable.
func isolateProcessGroup(cmd *exec.Cmd) {}

// terminateProcessGroup is a no-op on Windows, for the same reason.
func terminateProcessGroup(pid int) {}
