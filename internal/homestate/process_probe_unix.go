//go:build !windows

// process_probe_unix.go — card t1779: the process-listing probe the codex
// lane launch's sandbox-signature check rides. `ps` is the probe the Codex
// CLI itself runs and the process-identity machinery depends on; a sandbox
// that denies process introspection surfaces here as "operation not
// permitted", which the launch recognizes BEFORE it commits side effects.

package homestate

import (
	"os/exec"
	"strconv"
)

// ProcessListProbeError runs the external process-listing probe (ps) and
// reports the raw error. The output is discarded: the probe judges the
// ENVIRONMENT, not the process. A missing ps binary is an ordinary exec
// error, deliberately not distinguished here — only the caller's
// "operation not permitted" signature match decides anything.
func ProcessListProbeError(pid int) error {
	_, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
	return err
}
