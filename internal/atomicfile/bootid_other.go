//go:build !darwin && !freebsd && !netbsd && !openbsd && !dragonfly && !linux

package atomicfile

// Platforms without a stdlib boot identity (windows among them): current-
// BootID returns empty, boot comparison is unavailable, and the stale-lock
// break falls back to pid liveness alone — which on windows reads as alive
// for every findable pid, so the break never fires there. The wedge such a
// lock leaves is operator-recoverable; an incorrect break would discard a
// live writer's committed mutation, and only the first direction is
// acceptable under the D40 invariant.

// currentBootID returns "" — no boot identity is available.
func currentBootID() string {
	return ""
}
