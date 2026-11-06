//go:build unix

package cli

import (
	"os"
	"syscall"
)

// openJevCostLedger opens the decision ledger for reading, non-blocking: a
// FIFO at the path returns at once instead of blocking the read-only report
// forever, and readJevCostLedger's regular-file check then treats it as
// absent. The read-side mirror of the recorder's openLedger
// (internal/jevledger/open_unix.go) — sync audit round 1 P2 of card t1971
// (receipt rcpt-11b837f9258ea3ca20d0620e).
func openJevCostLedger(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
