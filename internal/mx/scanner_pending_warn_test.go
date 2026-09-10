package mx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scanPendingWarnFixture writes body verbatim (no header is prepended, so the
// line numbers in body are the line numbers the scanner sees), scans it, and
// logs every tag and warning so a -v run records the observed state.
func scanPendingWarnFixture(t *testing.T, body string) (string, []Tag, []string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.go")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	s := NewScanner()
	tags, err := s.ScanFile(path)
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	warnings := s.GetWarnings()
	for _, tag := range tags {
		t.Logf("tag line=%d kind=%s spec=%q reason=%q", tag.Line, tag.Kind, tag.SpecRef, tag.Reason)
	}
	t.Logf("warnings=%d %v", len(warnings), warnings)
	return path, tags, warnings
}

// missingReasonAt counts MissingReasonForWarn warnings naming path:line.
func missingReasonAt(warnings []string, path string, line int) int {
	loc := fmt.Sprintf("%s:%d ", path, line)
	n := 0
	for _, w := range warnings {
		if strings.Contains(w, "MissingReasonForWarn") && strings.Contains(w, loc) {
			n++
		}
	}
	return n
}

func countWarnings(warnings []string, fragment string) int {
	n := 0
	for _, w := range warnings {
		if strings.Contains(w, fragment) {
			n++
		}
	}
	return n
}

// C1: a reason-less WARN followed by a non-WARN tag must not vanish silently.
// The unpaired WARN is reported and — like the three other unpaired-WARN exits
// (window expiry, too-late REASON, end of file) — is not added to the tags.
func TestPendingWarn_C1_ReasonlessWarnThenNote_EmitsMissingReason(t *testing.T) {
	path, tags, warnings := scanPendingWarnFixture(t, "package fixture\n// @MX:WARN: missing reason\n// @MX:NOTE: another tag\n")

	if got := missingReasonAt(warnings, path, 2); got != 1 {
		t.Errorf("MissingReasonForWarn for WARN at line 2: want 1, got %d (%v)", got, warnings)
	}
	if len(tags) != 1 || tags[0].Kind != MXNote {
		t.Errorf("tags: want exactly the NOTE, got %+v", tags)
	}
}

// C2: a reason-less WARN followed by another WARN — the first WARN is reported,
// and the second (also reason-less) is reported at end of file.
func TestPendingWarn_C2_ReasonlessWarnThenWarn_FirstWarnGetsMissingReason(t *testing.T) {
	path, tags, warnings := scanPendingWarnFixture(t, "package fixture\n// @MX:WARN: first\n// @MX:WARN: second\n")

	if got := missingReasonAt(warnings, path, 2); got != 1 {
		t.Errorf("MissingReasonForWarn for first WARN at line 2: want 1, got %d (%v)", got, warnings)
	}
	if got := missingReasonAt(warnings, path, 3); got != 1 {
		t.Errorf("MissingReasonForWarn for second WARN at line 3: want 1, got %d (%v)", got, warnings)
	}
	if len(tags) != 0 {
		t.Errorf("tags: want none (both WARNs unpaired), got %+v", tags)
	}
}

// C3: the normal pairing — WARN then REASON — yields one WARN with its Reason
// and no MissingReasonForWarn.
func TestPendingWarn_C3_WarnThenReason_NoMissingReason(t *testing.T) {
	_, tags, warnings := scanPendingWarnFixture(t, "package fixture\n// @MX:WARN: danger\n// @MX:REASON: shared state\n")

	if len(tags) != 1 || tags[0].Kind != MXWarn || tags[0].Reason != "shared state" {
		t.Errorf("tags: want one WARN with reason %q, got %+v", "shared state", tags)
	}
	if got := countWarnings(warnings, "MissingReasonForWarn"); got != 0 {
		t.Errorf("MissingReasonForWarn: want 0, got %d (%v)", got, warnings)
	}
}

// C4: an @MX:SPEC between a WARN and its REASON belongs to the WARN, not to the
// tag that precedes the WARN.
func TestPendingWarn_C4_SpecBetweenWarnAndReason_AttachesToWarn(t *testing.T) {
	_, tags, warnings := scanPendingWarnFixture(t,
		"package fixture\n// @MX:NOTE: prior tag\n// @MX:WARN: danger\n// @MX:SPEC: SPEC-WARN-001\n// @MX:REASON: shared state\n")
	assertSpecOnWarnOnly(t, tags, warnings)
}

// C5: the reverse order (REASON before SPEC) already attaches correctly and
// must reach the same end state as C4.
func TestPendingWarn_C5_SpecAfterReason_AttachesToWarn(t *testing.T) {
	_, tags, warnings := scanPendingWarnFixture(t,
		"package fixture\n// @MX:NOTE: prior tag\n// @MX:WARN: danger\n// @MX:REASON: shared state\n// @MX:SPEC: SPEC-WARN-001\n")
	assertSpecOnWarnOnly(t, tags, warnings)
}

func assertSpecOnWarnOnly(t *testing.T, tags []Tag, warnings []string) {
	t.Helper()
	if len(tags) != 2 {
		t.Fatalf("tags: want 2, got %d (%+v)", len(tags), tags)
	}
	note, warn := tags[0], tags[1]
	if note.Kind != MXNote || note.SpecRef != "" {
		t.Errorf("NOTE: want kind=NOTE spec=\"\", got kind=%s spec=%q", note.Kind, note.SpecRef)
	}
	if warn.Kind != MXWarn || warn.SpecRef != "SPEC-WARN-001" || warn.Reason != "shared state" {
		t.Errorf("WARN: want kind=WARN spec=SPEC-WARN-001 reason=%q, got kind=%s spec=%q reason=%q",
			"shared state", warn.Kind, warn.SpecRef, warn.Reason)
	}
	for _, frag := range []string{"MissingReasonForWarn", "DanglingSpecRef"} {
		if got := countWarnings(warnings, frag); got != 0 {
			t.Errorf("%s: want 0, got %d (%v)", frag, got, warnings)
		}
	}
}

// The three pre-existing unpaired-WARN exits. Each one reports the WARN and
// does not add it to the tags; the superseded-by-a-new-tag exit (C1, C2)
// follows the same rule.

// E1: window expiry on a non-@MX line (line 6 > WARN line 2 + 3).
func TestPendingWarnExit_E1_WindowExpiryOnPlainLine_DiagnosticOnly(t *testing.T) {
	path, tags, warnings := scanPendingWarnFixture(t,
		"package fixture\n// @MX:WARN: danger\nfunc a() {}\nfunc b() {}\nfunc c() {}\nfunc d() {}\n")
	assertUnpairedWarnDiagnosticOnly(t, path, 2, tags, warnings)
}

// E2: a REASON that arrives too late (line 6 > WARN line 2 + 3). The lines in
// between are @MX sub-lines, so the plain-line expiry arm (E1) never runs and
// the too-late REASON arm is the one that reports.
func TestPendingWarnExit_E2_TooLateReason_DiagnosticOnly(t *testing.T) {
	path, tags, warnings := scanPendingWarnFixture(t,
		"package fixture\n// @MX:WARN: danger\n// @MX:PRIORITY: P1\n// @MX:PRIORITY: P1\n// @MX:PRIORITY: P1\n// @MX:REASON: too late\n")
	assertUnpairedWarnDiagnosticOnly(t, path, 2, tags, warnings)
}

// E3: end of file with the WARN still pending.
func TestPendingWarnExit_E3_EndOfFile_DiagnosticOnly(t *testing.T) {
	path, tags, warnings := scanPendingWarnFixture(t, "package fixture\n// @MX:WARN: danger\n")
	assertUnpairedWarnDiagnosticOnly(t, path, 2, tags, warnings)
}

func assertUnpairedWarnDiagnosticOnly(t *testing.T, path string, line int, tags []Tag, warnings []string) {
	t.Helper()
	if got := missingReasonAt(warnings, path, line); got != 1 {
		t.Errorf("MissingReasonForWarn for WARN at line %d: want 1, got %d (%v)", line, got, warnings)
	}
	for _, tag := range tags {
		if tag.Kind == MXWarn {
			t.Errorf("unpaired WARN must not be appended to tags, got %+v", tag)
		}
	}
}

// A pending WARN with no tag before it is a valid owner for @MX:SPEC: the
// sub-line attaches to the WARN and is not reported as dangling. (Before the
// ownership fix this shape emitted DanglingSpecRef because only tags already
// appended were considered owners.)
func TestPendingWarn_SpecOnPendingWarnWithoutPriorTag_NotDangling(t *testing.T) {
	_, tags, warnings := scanPendingWarnFixture(t,
		"package fixture\n// @MX:WARN: danger\n// @MX:SPEC: SPEC-WARN-001\n// @MX:REASON: shared state\n")

	if len(tags) != 1 || tags[0].Kind != MXWarn || tags[0].SpecRef != "SPEC-WARN-001" || tags[0].Reason != "shared state" {
		t.Errorf("tags: want one WARN spec=SPEC-WARN-001 reason=%q, got %+v", "shared state", tags)
	}
	if got := countWarnings(warnings, "DanglingSpecRef"); got != 0 {
		t.Errorf("DanglingSpecRef: want 0, got %d (%v)", got, warnings)
	}
}
