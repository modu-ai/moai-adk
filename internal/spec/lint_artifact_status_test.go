package spec_test

// lint_artifact_status_test.go — ArtifactStatusFieldForbiddenRule acceptance
// suite (SPEC-ARTIFACT-STATELESS-001 M2, REQ-AST-001-004..007).
//
// Every test below names, in its doc comment, the MUTATION that must turn it
// red. That is not decoration: the deliverable is a guard, and a guard whose
// criterion cannot fail is indistinguishable from one that is switched off —
// the failure shape card t355 landed and this SPEC's plan.md §C M2 cites by
// name. Each mutation was planted, observed red, and reverted; the verbatim
// failing output is recorded in progress.md §E.2.
//
// Fixtures live under testdata/artifactstatus/<case>/ and are linted ONE AT A
// TIME (an explicit spec.md path per call), so they deliberately share a SPEC
// id; DuplicateSPECIDRule never sees two of them in one run.
//
// The fixtures' spec.md files all carry `status: draft` in their own
// frontmatter, as the schema requires. That is the point: the rule under test
// governs the four SIBLING artifacts and must never read the file it is handed.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

const artifactStatusCode = "ArtifactStatusFieldForbidden"

// artifactStatusFindings lints one fixture case and returns only this rule's
// findings.
func artifactStatusFindings(t *testing.T, c string) []spec.Finding {
	t.Helper()
	base := filepath.Join(testdataDir, "artifactstatus", c)
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      base,
	})
	report, err := linter.Lint([]string{filepath.Join(base, "spec.md")})
	if err != nil {
		t.Fatalf("Lint(%s) returned unexpected error: %v", c, err)
	}
	return findingsForCode(report.Findings, artifactStatusCode)
}

// TestArtifactStatus_FiresOnPlanStatus decides AC-AST-001-04's positive half
// (REQ-AST-001-004): a `status:` field in plan.md frontmatter is rejected, and
// the finding points at the status LINE rather than at the file's first line.
//
// MUTATION: rename the fixture's frontmatter key from `status:` to `state:` —
// findings drop to 0.
func TestArtifactStatus_FiresOnPlanStatus(t *testing.T) {
	got := artifactStatusFindings(t, "plan-status")
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1: %+v", len(got), got)
	}
	f := got[0]
	if filepath.Base(f.File) != "plan.md" {
		t.Errorf("File = %s, want plan.md", f.File)
	}
	// The status field is the 5th line of the fixture's frontmatter block. A
	// rule reporting line 1 would be reporting the fence, not the violation.
	if f.Line != 5 {
		t.Errorf("Line = %d, want 5 (the status line, not the `---` fence)", f.Line)
	}
	if f.Severity != spec.SeverityError {
		t.Errorf("Severity = %q, want %q (plan.md §B3)", f.Severity, spec.SeverityError)
	}
	if !strings.Contains(f.Message, "status axis") {
		t.Errorf("Message does not state the axis it governs: %q", f.Message)
	}
}

// TestArtifactStatus_IgnoresBodyText decides the block-scoping half of
// REQ-AST-001-004: `status:` outside a leading frontmatter block is prose.
//
// The `no-status` fixture carries both shapes the scoping must reject: a
// plan.md whose frontmatter has every field EXCEPT status (the shape the D1
// cleanup leaves behind — a finding here would mean the rule fires on its own
// cleanup's output), and an acceptance.md that opens with a heading and
// therefore has no frontmatter block at all, yet writes `status: draft` in the
// body twice, once after a horizontal rule.
//
// MUTATION: drop the "block opens only at line 1" guard — acceptance.md's body
// `status: draft` (line 7) is read as a frontmatter field and findings rise
// from 0 to 1.
func TestArtifactStatus_IgnoresBodyText(t *testing.T) {
	if got := artifactStatusFindings(t, "no-status"); len(got) != 0 {
		t.Fatalf("findings = %d, want 0: %+v", len(got), got)
	}
}

// TestArtifactStatus_FiresOnAllFourArtifacts decides REQ-AST-001-004's set
// half: the governed set is all four artifacts, not whichever one is read
// first. design.md additionally writes `status:draft` with no space after the
// colon — the shape the counting predicate (`^status:[[:space:]]`) misses and
// this rule deliberately catches.
//
// MUTATION: remove any one name from `statelessArtifacts` — findings drop to 3.
// MUTATION: require whitespace after the colon — design.md stops firing, 3.
func TestArtifactStatus_FiresOnAllFourArtifacts(t *testing.T) {
	got := artifactStatusFindings(t, "four-artifacts")
	if len(got) != 4 {
		t.Fatalf("findings = %d, want 4: %+v", len(got), got)
	}
	want := map[string]bool{"plan.md": false, "acceptance.md": false, "design.md": false, "research.md": false}
	for _, f := range got {
		base := filepath.Base(f.File)
		if _, ok := want[base]; !ok {
			t.Errorf("finding on unexpected artifact %s", base)
			continue
		}
		want[base] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("no finding for %s — the governed set is not closed over all four", name)
		}
	}
}

// TestArtifactStatus_IgnoresSpecAndProgress decides REQ-AST-001-005 / -011: the
// two files outside the rule stay outside it. The fixture gives BOTH the
// violating shape — spec.md carries `status: draft` per the schema, progress.md
// carries one in a frontmatter block of its own.
//
// MUTATION: derive the artifact set from a directory scan of `*.md` with
// exclusions, instead of the closed `statelessArtifacts` list, and forget one
// exclusion — findings rise from 0.
func TestArtifactStatus_IgnoresSpecAndProgress(t *testing.T) {
	if got := artifactStatusFindings(t, "spec-progress-only"); len(got) != 0 {
		t.Fatalf("findings = %d, want 0: %+v", len(got), got)
	}
}

// TestArtifactStatus_SurvivesEraDemotion decides REQ-AST-001-006 behaviourally,
// where AC-AST-001-03 decides it textually. The fixture's spec.md declares no
// `era:` and has no progress.md, so the era classifier reads it as
// grandfather-era and `applyEraDemotion` runs against its findings. Because the
// code is absent from `eraDemotableCodes`, the finding must come through as a
// non-advisory error regardless.
//
// This is the criterion that turns red if someone "fixes" a red corpus by
// adding the code to `eraDemotableCodes` — the move that silently converts this
// rule into a suggestion, and that plan.md §B2 pairs with splitting M3 out.
//
// MUTATION: add "ArtifactStatusFieldForbidden" to eraDemotableCodes — severity
// becomes warning and Advisory becomes true.
func TestArtifactStatus_SurvivesEraDemotion(t *testing.T) {
	got := artifactStatusFindings(t, "grandfathered")
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1: %+v", len(got), got)
	}
	if got[0].Severity != spec.SeverityError {
		t.Errorf("Severity = %q, want %q — the code was era-demoted", got[0].Severity, spec.SeverityError)
	}
	if got[0].Advisory {
		t.Error("finding is Advisory — --strict would not escalate it; check eraDemotableCodes")
	}
}
