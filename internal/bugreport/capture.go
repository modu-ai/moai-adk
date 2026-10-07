package bugreport

import (
	"runtime"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// spoolAppendFn is the spool-write seam. Production is appendSpoolLine; the
// fail-open tests substitute blocking, panicking, and full stubs to prove
// Capture returns within its box without panicking.
var spoolAppendFn = appendSpoolLine

// captureFrames collects the stack at the capture site and filters it to the
// moai-internal function names. Called from a recover-site's deferred
// function, the stack still carries the panicking frames below it.
func captureFrames() []string {
	var pc [64]uintptr
	n := runtime.Callers(3, pc[:])
	if n == 0 {
		return nil
	}
	return FramesFromPC(pc[:n])
}

// Capture is the pipeline's single entry point (REQ-ANON-006/008). Every
// registered emit site — the main goroutine's deferred recover, the
// registered recover sites, the hook registry's error branches, the CLI's
// internal-error and template sites — calls exactly this function.
//
// The contract is fail-open and time-boxed:
//   - it returns nothing and never panics (the deferred recover below);
//   - it reads consent from the user-scoped store first and records NOTHING
//     when participation is off or a project file claims otherwise (the
//     reader never opens a project file);
//   - it attributes the signal in place, while the error chain is alive,
//     with errors.Is/errors.As only — no error text is ever read;
//   - user and environment verdicts are dropped without recording;
//   - a moai or ambiguous verdict appends one bounded JSONL line to the
//     user-scoped spool (D36) inside the 50 ms box — a write that does not
//     finish in time is abandoned (the bounded write still completes on its
//     own goroutine);
//   - a panic whose stack carries no moai frame stays local.
//
// Capture performs no network input or output and invokes no language
// model — the package cannot even import os/exec or net/http (REQ-ANON-025).
//
// @MX:ANCHOR: [AUTO] bugreport.Capture — the single capture entry point (main recover, twelve recover sites, hook registry, update paths)
// @MX:REASON: a second capture path would drift the consent gate, the verdict fixation, or the store location — the three properties the SPEC's privacy contract rests on (REQ-ANON-006/008)
// @MX:WARN: [AUTO] hot hook-path discipline — fail-open, time-boxed, network-free
// @MX:REASON: capture runs inside hook dispatch and main's crash path; a blocking or panicking capture would stall a hook event or eat a crash (REQ-ANON-008)
func Capture(kind Kind, err error, reason Reason, detail Detail) {
	defer func() { _ = recover() }() // fail-open: a capture bug must never take the host down

	if !kind.Valid() {
		return
	}
	if !config.ReadUserParticipation().Enabled {
		return
	}
	if detail != nil {
		if validateDetailForKind(kind, detail) != nil {
			return
		}
	}
	if !reason.Valid() && reason != "" {
		return
	}

	verdict, decided := Attribute(kind, err, reason)
	if verdict == VerdictUser || verdict == VerdictEnvironment {
		return
	}

	frames := captureFrames()
	if IsLocalOnly(kind, frames) {
		return
	}

	entry := spoolEntry{
		Kind:    kind,
		Verdict: verdict,
		Reason:  string(decided),
		Frames:  frames,
	}
	if detail != nil {
		entry.Detail = detail.Token()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = spoolAppendFn(entry)
	}()
	select {
	case <-done:
	case <-time.After(config.DefaultBugreportCaptureTimeBox):
		// Abandoned within the box; the bounded write completes on its own.
	}
}
