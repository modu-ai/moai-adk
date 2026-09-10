// codemaps_fold_judgments_test.go: Fixture-based validation of the fold and
// omission judgment verb documented in the codemaps workflow
// (`internal/template/templates/.claude/skills/moai/workflows/codemaps.md`,
// mirrored at `.claude/skills/moai/workflows/codemaps.md`).
//
// Like plan_audit_traceability_test.go, these tests EXTRACT the first ```bash
// block under the judgment-check heading and run it against inline fixtures, so
// a weakened verb turns a named cell red. The verb is never pasted here.
//
// Contract pinned here (card t566):
//   - Judgments live in fold-judgments.txt next to the generated documents, one
//     `fold <unit>` or `omission <unit>` per line; blank lines and # comments
//     are ignored.
//   - Hits are fixed-string line matches over the generated *.md documents
//     only; the judgments file is never in the scanned set.
//   - Every omission unit needs >= 1 matching line (else UNCOVERED), every fold
//     unit needs 0 (else FOLD-PROSE with the line count). Units are checked one
//     by one, never through a hit-zero census that file-granularity units never
//     appear in.
//   - An unreadable or judgment-free input, a malformed line, or an empty
//     document set is a GAP, never a silent pass. The script always exits 0.
//
// Sentinel on failure: CODEMAPS_FOLD_JUDGMENT_DRIFT — the verb in the workflow
// no longer matches the documented behavior.
package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const foldDrift = "CODEMAPS_FOLD_JUDGMENT_DRIFT"

// foldCheckHeading opens the section whose first bash block is the verb.
const foldCheckHeading = "### Fold and Omission Judgment Check"

// codemapsWorkflowDocs are the two copies of the workflow, relative to this
// package directory. The template copy is the one embedded and distributed.
var codemapsWorkflowDocs = map[string]string{
	"template": filepath.Join("..", "template", "templates", ".claude", "skills", "moai", "workflows", "codemaps.md"),
	"local":    filepath.Join("..", "..", ".claude", "skills", "moai", "workflows", "codemaps.md"),
}

// Fixture units: one package-granularity and one file-granularity unit per
// kind, since file-granularity units are the ones a package census cannot see.
const (
	foldPkg      = "src/legacy/shim"
	foldFile     = "src/web/generated_view.ext"
	omissionPkg  = "src/chain"
	omissionFile = "src/cli/mirror_heal.ext"
)

var foldJudgments = "# judgments from the last refresh\n" +
	"fold " + foldPkg + "\n" +
	"fold " + foldFile + "\n" +
	"\n" +
	"omission " + omissionPkg + "\n" +
	"omission " + omissionFile + "\n"

// greenDocs is the normal state: both omission units described, no fold unit.
func greenDocs() map[string]string {
	return map[string]string{
		"overview.md": "# Overview\n\nThe `" + omissionPkg + "` package records the delivery chain.\n",
		"modules.md":  "# Modules\n\n- `" + omissionFile + "` repairs mirrored files.\n- `src/core` holds shared types.\n",
	}
}

// runFoldVerb writes docs and (when non-nil) the judgments file into a fresh
// project's codemaps directory, then runs the verb from the project root.
func runFoldVerb(t *testing.T, docs map[string]string, judgments *string) string {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	verb := auditVerb(t, codemapsWorkflowDocs["template"], foldCheckHeading)
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, body := range docs {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if judgments != nil {
		if err := os.WriteFile(filepath.Join(dir, "fold-judgments.txt"), []byte(*judgments), 0o644); err != nil {
			t.Fatalf("write fold-judgments.txt: %v", err)
		}
	}

	cmd := exec.Command("bash", "-c", verb)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: verb execution failed: %v\noutput: %s", foldDrift, err, out)
	}
	return string(out)
}

// markerLines returns every line starting with one of the three finding markers.
func markerLines(out string) []string {
	var lines []string
	for _, marker := range []string{"UNCOVERED:", "FOLD-PROSE:", "GAP:"} {
		lines = append(lines, findingLines(out, marker)...)
	}
	return lines
}

func TestCodemapsFoldJudgments(t *testing.T) {
	t.Parallel()

	t.Run("green_normal_state", func(t *testing.T) {
		t.Parallel()
		// fold-judgments.txt sits in the scanned directory and names every
		// unit; zero FOLD-PROSE lines here also proves it does not self-match.
		out := runFoldVerb(t, greenDocs(), strPtr(foldJudgments))
		if lines := markerLines(out); len(lines) != 0 {
			t.Errorf("%s: normal state produced markers %v; output: %s", foldDrift, lines, out)
		}
		if !hasFinding(out, "COLLECTED: fold=2 omission=2") {
			t.Errorf("%s: expected COLLECTED: fold=2 omission=2; got: %s", foldDrift, out)
		}
	})

	t.Run("red_fold_prose", func(t *testing.T) {
		t.Parallel()
		docs := greenDocs()
		docs["modules.md"] += "- `" + foldFile + "` renders the generated view.\n"
		out := runFoldVerb(t, docs, strPtr(foldJudgments))
		want := []string{"FOLD-PROSE: " + foldFile + " 1"}
		if lines := markerLines(out); strings.Join(lines, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s: expected exactly %v; got %v; output: %s", foldDrift, want, lines, out)
		}
	})

	t.Run("red_uncovered_omission", func(t *testing.T) {
		t.Parallel()
		docs := greenDocs()
		docs["overview.md"] = "# Overview\n\nNo chain description here.\n"
		out := runFoldVerb(t, docs, strPtr(foldJudgments))
		want := []string{"UNCOVERED: " + omissionPkg}
		if lines := markerLines(out); strings.Join(lines, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s: expected exactly %v; got %v; output: %s", foldDrift, want, lines, out)
		}
	})

	t.Run("gap_missing_judgments", func(t *testing.T) {
		t.Parallel()
		out := runFoldVerb(t, greenDocs(), nil)
		lines := markerLines(out)
		if len(lines) != 1 || !strings.HasPrefix(lines[0], "GAP:") {
			t.Errorf("%s: a missing judgments file must yield exactly one GAP line; got %v", foldDrift, lines)
		}
		if hasFinding(out, "COLLECTED:") {
			t.Errorf("%s: a missing judgments file produced a measurement; got: %s", foldDrift, out)
		}
	})

	t.Run("gap_empty_judgments", func(t *testing.T) {
		t.Parallel()
		out := runFoldVerb(t, greenDocs(), strPtr("# only a comment\n\n"))
		lines := markerLines(out)
		if len(lines) != 1 || !strings.HasPrefix(lines[0], "GAP:") {
			t.Errorf("%s: zero judgments must yield exactly one GAP line; got %v", foldDrift, lines)
		}
		if !hasFinding(out, "COLLECTED: fold=0 omission=0") {
			t.Errorf("%s: expected COLLECTED: fold=0 omission=0; got: %s", foldDrift, out)
		}
	})

	t.Run("gap_malformed_line", func(t *testing.T) {
		t.Parallel()
		out := runFoldVerb(t, greenDocs(), strPtr(foldJudgments+"omit src/typo\n"))
		want := []string{"GAP: malformed judgment line 7: omit src/typo"}
		if lines := markerLines(out); strings.Join(lines, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s: expected exactly %v; got %v; output: %s", foldDrift, want, lines, out)
		}
		if !hasFinding(out, "COLLECTED: fold=2 omission=2") {
			t.Errorf("%s: valid judgments beside a malformed line were not checked; got: %s", foldDrift, out)
		}
	})

	t.Run("gap_no_documents", func(t *testing.T) {
		t.Parallel()
		// With no documents every fold unit reads zero hits; that must not pass.
		out := runFoldVerb(t, map[string]string{}, strPtr(foldJudgments))
		lines := markerLines(out)
		if len(lines) != 1 || !strings.HasPrefix(lines[0], "GAP:") {
			t.Errorf("%s: an empty document set must yield exactly one GAP line; got %v", foldDrift, lines)
		}
	})
}

// TestCodemapsFoldJudgments_VerbIdenticalAcrossCopies: the local mirror carries
// the same verb as the distributed template copy.
func TestCodemapsFoldJudgments_VerbIdenticalAcrossCopies(t *testing.T) {
	t.Parallel()

	if auditVerb(t, codemapsWorkflowDocs["template"], foldCheckHeading) != auditVerb(t, codemapsWorkflowDocs["local"], foldCheckHeading) {
		t.Errorf("%s: the verb differs between the template and local copies", foldDrift)
	}
}
