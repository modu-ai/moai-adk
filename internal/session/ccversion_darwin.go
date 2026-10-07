//go:build darwin

// ccversion_darwin.go — the running-version platform probe: the pid's text
// (executable) mapping via lsof (SPEC-SESSION-CC-VERSION-001 REQ-SCV-001).
// The parse of the returned text is shared (ccversion.go), where the anchor
// duty lives: `-d txt` also lists mapped frameworks and dylibs, and only the
// mapping line naming the claude binary itself may satisfy the read.
package session

import (
	"context"
	"os/exec"
	"strconv"
	"time"
)

// ccversionLsofTimeout bounds the lsof exec: a wedged process table must not
// hang a `session list --cc-version` or a doctor run past it.
const ccversionLsofTimeout = 5 * time.Second

// platformReadProcessMapping reads the process's text mapping. The output is
// lsof's raw text (one line per mapping; the executable mapping's path field
// names the binary) — the shared parse anchors on the claude-binary line.
// Failure — lsof absent, the pid dead, an unreadable mapping, the bound
// exceeded — is ok=false: the caller renders unknown, never an error.
func platformReadProcessMapping(pid int) (string, bool) {
	if pid <= 0 {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), ccversionLsofTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "lsof", "-a", "-d", "txt", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}
