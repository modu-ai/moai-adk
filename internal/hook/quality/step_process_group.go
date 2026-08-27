package quality

import "time"

// A step's deadline signals the direct child and nothing else. When a
// descendant outlives that child holding the step's inherited stdout and
// stderr, the pipes never reach EOF and cmd.Wait blocks past the very deadline
// that was supposed to bound it — so the budget the gate reports and the moment
// the gate returns become two different things.
//
// Two mechanisms are needed, and neither substitutes for the other:
//
//  1. stepWaitGrace, applied as exec.Cmd.WaitDelay, bounds how long Wait may
//     wait on those pipes. It unblocks the gate.
//  2. isolateProcessGroup plus terminateProcessGroup actually end the
//     descendant. Without this, mechanism 1 returns promptly and leaves the
//     orphan running forever.
//
// The platform half lives in step_process_group_unix.go and
// step_process_group_windows.go, following the build-tagged pair at
// internal/spec/lock_{unix,windows}.go: no naked syscall in this shared body.

// stepWaitGrace bounds the wait after a step's process has been asked to stop,
// or has exited leaving its output pipes held. It is a fixed safety bound
// rather than a configured knob: a configured value set wrong reintroduces the
// unbounded wait this exists to prevent.
const stepWaitGrace = 2 * time.Second
