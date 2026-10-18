//go:build !windows

// process_probe_unix.go — card t1779: the process-listing probe the codex
// lane launch's sandbox-signature check rides. `ps` is the probe the Codex
// CLI itself runs and the process-identity machinery depends on; a sandbox
// that denies process introspection surfaces here as "operation not
// permitted", which the launch recognizes BEFORE it commits side effects.

package homestate

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessListProbeError runs the external process-listing probe (ps) and
// reports the failure with the ps process's stderr SURFACED in the error
// text: in a sandbox, ps exits 1 writing "Operation not permitted" to
// stderr, and a bare exec error string ("exit status 1") would hide exactly
// the text the cli-side sandbox-signature match needs (card-review P2-1).
// The combined output judges the ENVIRONMENT, not the process.
func ProcessListProbeError(pid int) error {
	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ps probe: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
