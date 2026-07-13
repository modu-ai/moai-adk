package cli

// Deferred binary self-update notice for moai init
// (SPEC-CLI-TUX-V3-002 M2b, REQ-TUX2-001..004).
//
// The former runInit flow executed a BLOCKING network self-update check
// (runBinaryUpdateStep followed by a process re-exec) before the first
// wizard interaction. This file replaces it with a deferred, check-only,
// non-blocking notice:
//
//   - The check starts only AFTER wizard completion (interactive) and after
//     the first phase output (non-interactive) — zero network before the
//     first wizard interaction (REQ-TUX2-001/004).
//   - The check is CHECK-ONLY: it never installs and never re-execs the
//     process — wizard answers must survive (REQ-TUX2-002). When a newer
//     binary exists, a stderr notice with the `moai update` follow-up hint
//     is emitted instead (explicit trade-off: templates deploy from the
//     current binary; the notice tells the user how to refresh).
//   - shouldSkipBinaryUpdate semantics (templates-only flag / env guard /
//     dev-build) are behavior-preserving: any skip condition also skips the
//     deferred check (REQ-TUX2-003).
//   - Init exit never blocks on network I/O beyond a bounded grace window
//     (REQ-TUX2-004): a check still in flight at flush time is abandoned.

import (
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// deferredUpdateResult carries the outcome of the deferred check.
type deferredUpdateResult struct {
	Available      bool
	LatestVersion  string
	CurrentVersion string
	Err            error
}

// deferredUpdateEnabled reports whether the deferred binary-update check
// should run. Injectable seam (REQ-TUX2-003 characterization); the default
// preserves shouldSkipBinaryUpdate's three skip semantics verbatim.
var deferredUpdateEnabled = func(cmd *cobra.Command) bool {
	return !shouldSkipBinaryUpdate(cmd)
}

// deferredUpdateCheck performs the network version check. Injectable seam
// (REQ-TUX2-001 network-order contract; tests assert call ordering).
var deferredUpdateCheck = defaultDeferredUpdateCheck

// deferredNoticeGrace bounds how long init exit waits for an in-flight
// check before abandoning it (REQ-TUX2-004 non-blocking property).
var deferredNoticeGrace = 1 * time.Second

// isInteractiveStdin reports whether stdin is a terminal (wizard gate).
// Injectable seam so the network-order contract is testable without a TTY.
var isInteractiveStdin = func() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

// runWizardFn runs the init wizard. Injectable seam for the network-order
// contract test (AC-TUX2-001); the default dispatches on the mode flags.
var runWizardFn = func(rootFlag, locale string, standardMode, advancedMode bool) (*wizard.WizardResult, error) {
	if standardMode {
		return wizard.RunWithDefaultsModes(rootFlag, locale, standardMode, advancedMode)
	}
	return wizard.RunWithDefaults(rootFlag, locale)
}

// startDeferredUpdateNotice launches the non-blocking deferred binary-update
// check and returns a flush func the caller invokes at init exit. When any
// shouldSkipBinaryUpdate condition holds, no goroutine starts and the flush
// is a no-op (REQ-TUX2-003).
func startDeferredUpdateNotice(cmd *cobra.Command) func(p printer.Printer) {
	if !deferredUpdateEnabled(cmd) {
		return func(printer.Printer) {}
	}

	// Capture the seam value before spawning so the goroutine never reads
	// the package var concurrently with a reassignment (race-safe).
	check := deferredUpdateCheck
	ch := make(chan *deferredUpdateResult, 1)
	go func() { ch <- check(cmd) }()

	return func(p printer.Printer) {
		select {
		case res := <-ch:
			if res == nil || res.Err != nil || !res.Available {
				// Non-fatal by design: a failed or up-to-date check never
				// affects the init result (acceptance.md §A scenario 1).
				return
			}
			p.Info("A newer moai binary is available: %s (current: %s). Run 'moai update' to upgrade and refresh templates.",
				res.LatestVersion, res.CurrentVersion)
		case <-time.After(deferredNoticeGrace):
			// Check still in flight — init exit must not block on network
			// I/O (REQ-TUX2-004). The abandoned goroutine drains into the
			// buffered channel and is garbage collected.
		}
	}
}

// defaultDeferredUpdateCheck mirrors runBinaryUpdateStep's checker wiring but
// is strictly CHECK-ONLY: no download, no install, no re-exec (REQ-TUX2-002).
func defaultDeferredUpdateCheck(_ *cobra.Command) *deferredUpdateResult {
	if deps != nil {
		if err := deps.EnsureUpdate(); err != nil {
			return &deferredUpdateResult{Err: err}
		}
	}
	if deps == nil || deps.UpdateChecker == nil {
		return &deferredUpdateResult{}
	}

	current := version.GetVersion()
	available, info, err := deps.UpdateChecker.IsUpdateAvailable(current)
	if err != nil {
		return &deferredUpdateResult{Err: err}
	}
	res := &deferredUpdateResult{Available: available, CurrentVersion: current}
	if info != nil {
		res.LatestVersion = info.Version
	}
	return res
}
