//go:build !darwin && !freebsd && !netbsd && !openbsd && !dragonfly && !linux

package atomicfile

// Platforms without a stdlib boot identity (windows among them): current-
// BootID returns empty, boot comparison is unavailable, and the stale-lock
// break falls back to pid liveness alone. Windows uses native process exit
// status; unverifiable processes remain alive. An incorrect break would
// discard a live writer's committed mutation, so unknown liveness never
// authorizes reclaim under the D40 invariant.

// currentBootID returns "" — no boot identity is available.
func currentBootID() string {
	return ""
}
