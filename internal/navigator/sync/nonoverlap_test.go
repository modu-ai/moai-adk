package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNonOverlap_SourceGrepLSEL exercises AC-013 Lens 1 (REQ-NS-013):
// internal/navigator/sync/ source MUST NOT literally name any LSEL surface
// path (lessons-inbox, state/lsel, memory/feedback_, hns-lsel, proposals).
// Per the design.md §5 source-comment hygiene advisory, this defensive
// grep keeps the runtime non-overlap mechanically observable.
func TestNonOverlap_SourceGrepLSEL(t *testing.T) {
	pkgDir := "." // test runs in the package directory
	forbidden := []string{
		"lessons-inbox",
		"state/lsel",
		"memory/feedback_",
		"hns-lsel",
		"state/lsel/proposals",
	}
	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		// Skip the test file itself (this one names the forbidden fragments
		// as the assertion target).
		if filepath.Base(m) == "nonoverlap_test.go" {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, f := range forbidden {
			if strings.Contains(src, f) {
				t.Errorf("%s: forbidden LSEL path fragment %q appears in source", filepath.Base(m), f)
			}
		}
	}
}

// TestNonOverlap_SourceGrepForbiddenWriteSurfaces exercises AC-014 Lens 1
// (REQ-NS-014): the production source MUST NOT reference the forbidden write
// paths (.moai/specs/, capability-map.md, audit-report.{md,json},
// capability-symbols.{md,json}) as WRITE targets. READ references (the join
// engine reads these as inputs) are allowed; this test specifically scopes
// the forbidden pattern to write-shaped verbs.
//
// We approximate by ensuring no source file mentions "WriteFile" on a path
// ending in any of the forbidden surface literals — a conservative proxy
// that catches the regression where a developer adds a write to one of the
// 3 chains' outputs.
func TestNonOverlap_SourceGrepForbiddenWriteSurfaces(t *testing.T) {
	pkgDir := "."
	forbiddenSuffixes := []string{
		"capability-map.md",
		"audit-report.md",
		"audit-report.json",
		"capability-symbols.md",
		"capability-symbols.json",
	}
	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		// Skip the test file itself (this one mentions the forbidden paths
		// as the assertion target).
		if filepath.Base(m) == "nonoverlap_test.go" {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		// Only flag when a forbidden suffix appears on the SAME line as a
		// write-shaped verb. Read references (default-arg Options fields,
		// loadAuditReport, loadCapabilitySymbols) DO NOT count.
		for _, line := range strings.Split(src, "\n") {
			if !strings.Contains(line, "WriteFile") && !strings.Contains(line, "os.Rename") {
				continue
			}
			for _, sfx := range forbiddenSuffixes {
				if strings.Contains(line, sfx) {
					t.Errorf("%s: write verb targets forbidden surface %q: %s",
						filepath.Base(m), sfx, strings.TrimSpace(line))
				}
			}
		}
	}
}
