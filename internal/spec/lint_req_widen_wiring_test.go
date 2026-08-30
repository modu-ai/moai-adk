package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-COVERAGE-RULE-SCOPE-001 M3 — the widened extraction wired into the live
// path, plus the severity treatment that keeps the corpus from reddening.
//
// AC-CRS-001-001 (live path), AC-CRS-001-005 (no unresolved error findings).

func writeWideSpecFixture(t *testing.T, body string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "SPEC-FIXW-001")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	fm := strings.Join([]string{
		"---",
		"id: SPEC-FIXW-001",
		`title: "fixture"`,
		`version: "0.1.0"`,
		"status: draft",
		"created: 2026-08-30",
		"updated: 2026-08-30",
		"author: t",
		"priority: P2",
		`phase: "v3.2.0 target"`,
		`module: "internal/spec"`,
		"lifecycle: spec-anchored",
		`tags: "fixture"`,
		"---",
		"",
	}, "\n")
	p := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(p, []byte(fm+body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

// TestParseSPECDoc_CollectsWidenedShapes is the live-path assertion: the
// widened shapes must reach doc.REQs, not merely a side pattern.
func TestParseSPECDoc_CollectsWidenedShapes(t *testing.T) {
	p := writeWideSpecFixture(t, strings.Join([]string{
		"## 2. Requirements",
		"",
		"- REQ-HOOK-001: The system shall do a thing.",
		"- **REQ-WF001-001**: The system shall do another thing.",
		"- REQ-VNRN-RT-001-001: The system shall do a third thing.",
		"- REQ-FIXW-001-001: The system shall do a fourth thing.",
		"",
	}, "\n"))

	doc := parseSPECDoc(p)
	if doc.ParseError != nil {
		t.Fatalf("parse error: %v", doc.ParseError)
	}
	want := []string{"REQ-HOOK-001", "REQ-WF001-001", "REQ-VNRN-RT-001-001", "REQ-FIXW-001-001"}
	if len(doc.REQs) != len(want) {
		t.Fatalf("expected %d REQs in the live path, got %d: %#v", len(want), len(doc.REQs), doc.REQs)
	}
	for i, id := range want {
		if doc.REQs[i].ID != id {
			t.Errorf("REQ[%d]: want %q, got %q", i, id, doc.REQs[i].ID)
		}
	}
}

// TestParseSPECDoc_MarksWidenedOnlyEntries checks the provenance flag that the
// severity treatment keys on. An entry the NARROW pattern already collected is
// NOT widened-only; its rules must behave exactly as before.
func TestParseSPECDoc_MarksWidenedOnlyEntries(t *testing.T) {
	p := writeWideSpecFixture(t, strings.Join([]string{
		"## 2. Requirements",
		"",
		"- REQ-FIXW-001-001: The system shall be collected by both patterns.",
		"- REQ-HOOK-001: The system shall be collected only by the widened pattern.",
		"",
	}, "\n"))

	doc := parseSPECDoc(p)
	if len(doc.REQs) != 2 {
		t.Fatalf("expected 2 REQs, got %d", len(doc.REQs))
	}
	if doc.REQs[0].Widened {
		t.Errorf("REQ-FIXW-001-001 is collected by the narrow pattern too; Widened must be false")
	}
	if !doc.REQs[1].Widened {
		t.Errorf("REQ-HOOK-001 is collected only by the widened pattern; Widened must be true")
	}
}

// TestCoverageRule_EmitsAdvisoryWarning is the severity treatment (plan.md §D
// option A). Advisory is set at the EMISSION SITE, not via eraDemotableCodes:
// that map is consulted only for SeverityError findings, so a warning can never
// reach it.
func TestCoverageRule_EmitsAdvisoryWarning(t *testing.T) {
	doc := &SPECDoc{
		Path: "SPEC-X/spec.md",
		REQs: parseREQsWide("- REQ-ABC-001: The system shall do a thing."),
	}
	findings := (&CoverageRule{}).Check(doc, nil)
	if len(findings) != 1 {
		t.Fatalf("expected 1 CoverageIncomplete finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Severity != SeverityWarning {
		t.Errorf("severity: want %q, got %q", SeverityWarning, f.Severity)
	}
	if !f.Advisory {
		t.Errorf("Advisory must be set at the emission site so --strict does not escalate it")
	}
}

// TestCoverageRule_AdvisoryWarningDoesNotGateStrict is the consequence test:
// without Advisory, --strict escalates the warning and the corpus reddens on
// the very run this SPEC exists to make survivable.
func TestCoverageRule_AdvisoryWarningDoesNotGateStrict(t *testing.T) {
	doc := &SPECDoc{
		Path: "SPEC-X/spec.md",
		REQs: parseREQsWide("- REQ-ABC-001: The system shall do a thing."),
	}
	r := &Report{Findings: (&CoverageRule{}).Check(doc, nil), Strict: true}
	if r.HasErrors() {
		t.Errorf("--strict escalated the CoverageIncomplete warning; Advisory is not being honored")
	}
}

// TestWidenedOnlyFindingsAreAdvisory covers the rules the widening ALSO reaches.
// CoverageRule is not doc.REQs' only consumer: EARSModalityRule emits
// ModalityMalformed at SeverityError and REQIDUniquenessRule emits InvalidREQID
// at SeverityError, and neither code is in eraDemotableCodes. Measured on the
// live corpus, wiring the widened pattern turns on 25 ModalityMalformed and 6
// InvalidREQID errors that no severity treatment of CoverageRule alone touches.
func TestWidenedOnlyFindingsAreAdvisory(t *testing.T) {
	t.Run("ModalityMalformed on a widened-only REQ", func(t *testing.T) {
		doc := &SPECDoc{
			Path: "SPEC-X/spec.md",
			REQs: []REQEntry{{ID: "REQ-HOOK-001", Text: "WHEN a thing happens, the system does something.", Line: 1, Widened: true}},
		}
		findings := (&EARSModalityRule{}).Check(doc, nil)
		if len(findings) != 1 || findings[0].Code != "ModalityMalformed" {
			t.Fatalf("expected 1 ModalityMalformed finding, got %#v", findings)
		}
		if findings[0].Severity != SeverityWarning || !findings[0].Advisory {
			t.Errorf("widened-only finding must be an advisory warning, got severity=%q advisory=%v",
				findings[0].Severity, findings[0].Advisory)
		}
	})

	t.Run("ModalityMalformed on a narrow REQ stays an error", func(t *testing.T) {
		doc := &SPECDoc{
			Path: "SPEC-X/spec.md",
			REQs: []REQEntry{{ID: "REQ-ABC-001-001", Text: "WHEN a thing happens, the system does something.", Line: 1, Widened: false}},
		}
		findings := (&EARSModalityRule{}).Check(doc, nil)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %#v", findings)
		}
		if findings[0].Severity != SeverityError || findings[0].Advisory {
			t.Errorf("pre-existing narrow behavior must be unchanged: want error/non-advisory, got severity=%q advisory=%v",
				findings[0].Severity, findings[0].Advisory)
		}
	})

	t.Run("InvalidREQID on a widened-only REQ", func(t *testing.T) {
		doc := &SPECDoc{
			Path: "SPEC-X/spec.md",
			REQs: []REQEntry{{ID: "REQ-256K-001", Text: "The system shall do a thing.", Line: 1, Widened: true}},
		}
		findings := (&REQIDUniquenessRule{}).Check(doc, nil)
		if len(findings) != 1 || findings[0].Code != "InvalidREQID" {
			t.Fatalf("expected 1 InvalidREQID finding, got %#v", findings)
		}
		if findings[0].Severity != SeverityWarning || !findings[0].Advisory {
			t.Errorf("widened-only finding must be an advisory warning, got severity=%q advisory=%v",
				findings[0].Severity, findings[0].Advisory)
		}
	})

	t.Run("InvalidREQID on a narrow REQ stays an error", func(t *testing.T) {
		doc := &SPECDoc{
			Path: "SPEC-X/spec.md",
			REQs: []REQEntry{{ID: "REQ-256K-001", Text: "The system shall do a thing.", Line: 1, Widened: false}},
		}
		findings := (&REQIDUniquenessRule{}).Check(doc, nil)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %#v", findings)
		}
		if findings[0].Severity != SeverityError || findings[0].Advisory {
			t.Errorf("pre-existing narrow behavior must be unchanged: want error/non-advisory, got severity=%q advisory=%v",
				findings[0].Severity, findings[0].Advisory)
		}
	})
}
