//go:build windows

package hook

import (
	"io"
	"os"
)

// readRuleFileBytes reads the rule file at path. Windows keeps the plain
// open+read: opening a FIFO-like named pipe for reading does not carry the
// POSIX open-hang semantics this guard exists for, and there is no O_NONBLOCK
// to race with — the Stat→ReadFile window has no blocking payload here. The
// TOCTOU hardening is POSIX-only by build tag (the FIFO fixture file carries
// the same //go:build !windows constraint).
func readRuleFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// readRuleFileBytesBounded is readRuleFileBytes with the read itself
// bounded: at most max bytes leave the file into memory (card t1776 audit
// round-6 — the cap must apply to the read, not after it).
func readRuleFileBytesBounded(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(io.LimitReader(f, max))
}
