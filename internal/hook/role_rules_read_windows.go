//go:build windows

package hook

import "os"

// readRuleFileBytes reads the rule file at path. Windows keeps the plain
// open+read: opening a FIFO-like named pipe for reading does not carry the
// POSIX open-hang semantics this guard exists for, and there is no O_NONBLOCK
// to race with — the Stat→ReadFile window has no blocking payload here. The
// TOCTOU hardening is POSIX-only by build tag (the FIFO fixture file carries
// the same //go:build !windows constraint).
func readRuleFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
