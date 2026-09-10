package quality

// gate_summary.go — the execution summary a gate run builds as it goes.
//
// A run used to report only its failures. Every step that passed, skipped, or
// was turned off left no trace: runStep collapsed its success path to an empty
// string, so a run that replayed a build cache and a run that genuinely
// executed the suite were byte-identical silence at the caller's terminal.
//
// The summary is a value the run accumulates rather than a string steps append
// to. A string cannot carry a duration and cannot distinguish "configured but
// never reached" from "skipped", which is precisely the distinction a reader
// needs when a run aborts halfway.

import (
	"fmt"
	"strings"
	"time"
)

// stepOutcome names what actually became of a configured step during one run.
//
// The four values are mutually exclusive and are written as whole words so a
// reader — and a test — can tell them apart without parsing prose. A run that
// reaches a verdict never leaves a step at outcomeNotReached.
type stepOutcome string

const (
	// outcomeNotReached is every step's seeded state: configured, and nothing
	// observed about it yet. It survives to the rendered summary only when the
	// run aborted before the step's turn came.
	outcomeNotReached stepOutcome = "not reached"
	// outcomeExecuted means the command was handed to the process launcher.
	// It says nothing about the verdict — a failing step executed too.
	outcomeExecuted stepOutcome = "executed"
	// outcomeSkipped means a guard declined to run the step, for the reason the
	// record carries.
	outcomeSkipped stepOutcome = "skipped"
	// outcomeDisabled means configuration turned the step off.
	outcomeDisabled stepOutcome = "disabled"
)

// Skip reasons, one per path by which a step can go unrun. Each names its own
// observation: the config-off reason must not read as a missing tool, and the
// absent-binary reason must not read as a missing config file. Keeping them
// mutually distinct is the point — one shared "skipped" token would report
// five different findings as one.
const (
	reasonDisabledByConfig = "turned off by the gate.disabled_steps setting"
	reasonSkipTests        = "turned off by the gate.skip_tests setting"
	reasonNotReached       = "an earlier step ended the run"

	reasonOptionalBinaryAbsentFmt = "the optional tool %q was not found on PATH"
	reasonConfigFilesAbsentFmt    = "none of its config files exist in the project directory (%s)"
	reasonNoStagedMatchFmt        = "no staged file carries one of its extensions (%s)"
	reasonNoProjectSourceFmt      = "the project holds no source file with one of its extensions (%s)"
)

// summaryHeaderPrefix opens the rendered block. Callers print the gate's output
// verbatim, so the header is what tells a reader the lines below are a step
// inventory rather than a failure report.
const summaryHeaderPrefix = "quality gate steps"

// summaryDetailSep separates a row's outcome field from its detail — the
// executed command line, or the reason the step did not run. The em dash is
// chosen because neither a step label nor a command line contains one, which
// keeps the row unambiguously splittable.
const summaryDetailSep = " — "

// stepRecord is one configured step's row: what it was, what became of it, and
// what was measured while it ran.
type stepRecord struct {
	// label identifies which step the row is about. It is the key
	// gate.disabled_steps is written against.
	label string
	// outcome is what became of the step.
	outcome stepOutcome
	// command is the command line actually handed to the launcher — the step's
	// binary together with its arguments. It diverges from label whenever a
	// step is resolved to a substitute before execution, which is exactly when
	// reporting the label instead would misreport what ran.
	command string
	// duration is the wall-clock span measured around the step's own execution.
	// Zero on every outcome but outcomeExecuted.
	duration time.Duration
	// startedAt is the wall-clock instant the step's execution began. Zero on
	// every outcome but outcomeExecuted. A duration alone cannot say WHEN a
	// step ran, and the when is what makes two runs' execution windows
	// comparable — the question a serialized-run reader asks.
	startedAt time.Time
	// reason explains a non-executed outcome.
	reason string
}

// runSummary accumulates one record per configured step, in execution order.
//
// Records are seeded before the first step runs, so a step the run never
// reaches is still reportable. Nothing is written back from configuration
// afterwards: a row says only what was observed.
type runSummary struct {
	records []*stepRecord
}

// newRunSummary returns an empty summary ready to be seeded.
func newRunSummary() *runSummary {
	return &runSummary{}
}

// seed registers a configured step under its label, in execution order.
// Re-seeding a label already present is a no-op, so a toolchain that lists the
// same step twice does not produce two rows.
func (s *runSummary) seed(label string) {
	if s == nil || label == "" {
		return
	}
	if s.recordFor(label) != nil {
		return
	}
	s.records = append(s.records, &stepRecord{label: label, outcome: outcomeNotReached})
}

// recordFor returns the row for label, or nil when the label was never seeded.
func (s *runSummary) recordFor(label string) *stepRecord {
	if s == nil {
		return nil
	}
	for _, rec := range s.records {
		if rec.label == label {
			return rec
		}
	}
	return nil
}

// relabel renames a seeded row. A step whose command is resolved to a
// substitute before execution reaches executeStep under the resolved name, so
// the seeded row has to follow it or the run would report two rows for one
// step: one never reached, one unaccounted for.
func (s *runSummary) relabel(from, to string) {
	if s == nil || from == to {
		return
	}
	if rec := s.recordFor(from); rec != nil {
		rec.label = to
	}
}

// mark writes an observation onto a step's row, seeding a row first when the
// label was not part of the plan. Seeding on demand matters: a step observed
// but not listed would otherwise be reported as never reached, which is the
// same false claim in the other direction.
func (s *runSummary) mark(label string, outcome stepOutcome, reason, command string, d time.Duration) {
	s.markAt(label, outcome, reason, command, time.Time{}, d)
}

// markAt is mark with the execution-start instant; only the executed outcome
// carries one.
func (s *runSummary) markAt(label string, outcome stepOutcome, reason, command string, startedAt time.Time, d time.Duration) {
	if s == nil || label == "" {
		return
	}
	rec := s.recordFor(label)
	if rec == nil {
		s.seed(label)
		rec = s.recordFor(label)
	}
	rec.outcome = outcome
	rec.reason = reason
	rec.command = command
	rec.startedAt = startedAt
	rec.duration = d
}

// markSkipped records that a guard declined to run the step.
func (s *runSummary) markSkipped(label, reason string) {
	s.mark(label, outcomeSkipped, reason, "", 0)
}

// markDisabled records that configuration turned the step off.
func (s *runSummary) markDisabled(label, reason string) {
	s.mark(label, outcomeDisabled, reason, "", 0)
}

// markExecuted records that the command was run, with the span measured around
// it and the instant it began. Called on every exit from the execution path —
// a step that failed, timed out, or could not be launched still executed as
// far as the summary is concerned, and its command line is still what the
// gate handed over.
func (s *runSummary) markExecuted(label, command string, startedAt time.Time, d time.Duration) {
	s.markAt(label, outcomeExecuted, "", command, startedAt, d)
}

// commandLine renders a step's binary and arguments the way they were handed to
// the process launcher. This is the value REQ-GTA-004 binds to: the step's
// label is a separate field that coincides with the command line only by
// accident.
func commandLine(binary string, args ...string) string {
	if len(args) == 0 {
		return binary
	}
	return binary + " " + strings.Join(args, " ")
}

// condenseSkipReason strips the "<label>: skipped (...)" framing an axis puts
// on its own notice, leaving the observation alone. A summary row already
// carries the label and the outcome; repeating both inside the reason makes
// the row read as if the same fact were stated three times.
func condenseSkipReason(label, reason string) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(reason), label+": skipped"))
	trimmed = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "("), ")"))
	if trimmed == "" {
		return strings.TrimSpace(reason)
	}
	return trimmed
}

// render writes the summary as human-facing text. Empty when nothing was
// configured, so a run with no steps adds no noise.
func (s *runSummary) render() string {
	if s == nil || len(s.records) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s (%d configured):", summaryHeaderPrefix, len(s.records))
	for _, rec := range s.records {
		b.WriteString("\n  - ")
		b.WriteString(rec.label)
		b.WriteString(": ")
		b.WriteString(rec.outcomeField())
		if detail := rec.detail(); detail != "" {
			b.WriteString(summaryDetailSep)
			b.WriteString(detail)
		}
	}
	return b.String()
}

// startedAtFormat renders an executed step's start instant — RFC3339 with
// millisecond precision, parseable by time.Parse(time.RFC3339, …) and fine
// enough to order two runs' steps against each other.
const startedAtFormat = "2006-01-02T15:04:05.000Z07:00"

// outcomeField renders the outcome, carrying the measured duration and the
// start instant for an executed step. Exactly one outcome token appears here,
// whatever the detail beside it says.
func (r *stepRecord) outcomeField() string {
	if r.outcome == outcomeExecuted {
		return fmt.Sprintf("%s in %s at %s", outcomeExecuted, r.duration.Round(time.Millisecond), r.startedAt.Format(startedAtFormat))
	}
	return string(r.outcome)
}

// detail returns the executed command line, or the reason the step did not run.
func (r *stepRecord) detail() string {
	if r.outcome == outcomeExecuted {
		return r.command
	}
	return r.reason
}

// joinBlocks stacks two output blocks with a blank line between them, dropping
// either when it is empty. The notices a run collects and the step summary are
// separate blocks: folding the summary in as one more notice line would bury
// it, and replacing the notices with it would drop them.
func joinBlocks(first, second string) string {
	first = strings.TrimSpace(first)
	second = strings.TrimSpace(second)
	switch {
	case first == "":
		return second
	case second == "":
		return first
	default:
		return first + "\n\n" + second
	}
}
