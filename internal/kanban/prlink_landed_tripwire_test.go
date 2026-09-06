// prlink_landed_tripwire_test.go — SPEC-TODO-LANDING-ATTRIBUTION-001
// AC-TLA-006 (M3 item 1, gated): the tripwire asserts the query the landed
// check runs, and it obtains that query by calling the exported builder, NOT
// by re-constructing it — a test that rebuilds the argv itself proves nothing
// about the code.
//
// The tripwire replaces the historical -E engine-flag tripwire
// (SPEC-KANBAN-QUEUE-PR-SYNC-001): the silent failure this file exists
// against now lives in the query SHAPE. A %B regression would feed commit
// bodies to the positional matcher as if they were subjects — byte-identical
// to a working predicate on most fixtures and wrong on exactly the
// body-mention population this SPEC was written for. A --grep regression
// would reintroduce the occurrence test REQ-TLA-001 forbids.
package kanban

import (
	"slices"
	"testing"
)

func TestLandedSubjectArgs_Tripwire(t *testing.T) {
	// The builder's own output is what is asserted — never a transcription.
	args, err := LandedSubjectArgs("origin/develop")
	if err != nil {
		t.Fatalf("argv: %v", err)
	}

	// Shape: a subject stream, on the resolved ref.
	if !slices.Contains(args, LandedSubjectFormatFlag) {
		t.Errorf("argv %v does not carry %s — the query is no longer a subject stream", args, LandedSubjectFormatFlag)
	}
	if !slices.Contains(args, "origin/develop") {
		t.Errorf("argv %v does not name the resolved ref", args)
	}

	// Forbidden shapes, each a named regression:
	if slices.Contains(args, "--format=%B") {
		t.Errorf("argv %v carries --format=%%B — commit bodies would be fed to the positional matcher as subjects", args)
	}
	if slices.Contains(args, "--grep") {
		t.Errorf("argv %v carries --grep — an occurrence test, which REQ-TLA-001 forbids", args)
	}
	if slices.Contains(args, "-E") || slices.Contains(args, "--extended-regexp") {
		t.Errorf("argv %v carries the POSIX ERE engine — \\b is unsupported there and the empty result is silent", args)
	}
	if slices.Contains(args, "--perl-regexp") {
		t.Errorf("argv %v carries the retired engine flag; the matcher is Go-side now", args)
	}
}
