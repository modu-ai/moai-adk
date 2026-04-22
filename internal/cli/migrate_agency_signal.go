//go:build !windows

// migrate_agency_signal.go: SIGINT/SIGTERM checkpoint handling for agency migration.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-013
package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// installSignalHandler sets up SIGINT/SIGTERM handling for the migration.
// When a signal is received, the checkpoint is flushed and the migration is aborted.
// Returns a cancel function that should be deferred to clean up the signal handler.
//
// @MX:WARN: [AUTO] goroutine launched for signal handling; must be cancelled on exit
// @MX:REASON: [AUTO] signal.Notify channel requires explicit stop; goroutine leak risk
func installSignalHandler(checkpoint *migrationCheckpoint, cpPath string, onSignal func(os.Signal)) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		select {
		case sig := <-sigCh:
			checkpoint.Timestamp = time.Now()
			if err := writeCheckpoint(cpPath, checkpoint); err != nil {
				fmt.Fprintf(os.Stderr, "warn: could not write checkpoint: %v\n", err)
			}
			if onSignal != nil {
				onSignal(sig)
			}
		case <-done:
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
