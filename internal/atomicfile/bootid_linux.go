//go:build linux

package atomicfile

// Linux exposes the boot id as a plain file read — no syscall surface, so
// no build-tag hazard beyond this file itself.

import "os"

// currentBootID returns a boot-unique identity for this machine.
func currentBootID() string {
	raw, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return ""
	}
	// Trim the trailing newline; the uuid string is the identity.
	out := raw
	for len(out) > 0 && (out[len(out)-1] == '\n' || out[len(out)-1] == ' ') {
		out = out[:len(out)-1]
	}
	return string(out)
}
