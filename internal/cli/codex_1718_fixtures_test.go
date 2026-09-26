package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// SPEC-CODEX-PARSER-SHAPE-001 AC-CPS-011 (card t1203, GitHub #1718).
//
// The fixtures under testdata/codex-1718/ are SANITIZED REDUCTIONS of the two
// live adversarial bodies #1718 reported: they keep the shape — a greeting
// ahead of the verdict, a localized verdict label, findings as a table or as
// bold severity-word bullets — and carry none of the originating project's
// paths, names, or content. S2p is S2 with only its greeting prefix removed
// (the ctrlB analogue); N1 and N2 are negatives that mention a verdict and use
// fail/pass as ordinary words without stating one.

const codex1718Dir = "testdata/codex-1718"

// Declared finding counts: the number of findings each fixture FILE states,
// fixed here as constants independently of what the parser returns. P4, P5 and
// P11 tie the file's structure to these numbers, so a one-line prose
// reduction cannot satisfy them.
const (
	codex1718S1DeclaredFindings = 3
	codex1718S2DeclaredFindings = 1
)

// codex1718Synthesis is one expected synthesis row: a fixture file on one
// review path.
type codex1718Synthesis struct {
	file     string
	method   string
	verdict  string
	findings int
}

// codex1718Expected is the synthesis table TestCodex1718Fixtures asserts,
// row for row.
//
// History, kept because the table has changed meaning twice. The commit that
// introduced these fixtures asserted the PRE-change outputs — the values
// E-1718 recorded for the raw bodies (S1 and S2 inconclusive/0 on turn/start
// and pass/0 on review/start, S2p fail/0 on both) — which closed AC-CPS-011's
// fidelity check. Candidate (a) then widened the recognizers, and the S rows
// moved to the post-(a) expectation (AC-CPS-012): every S-fixture yields fail
// with its declared finding count on both paths. M4 (candidate (b),
// AC-CPS-004) moved the N1/N2 review/start rows a second time: the native
// request now pins its output format, so a no-signal body — which both N
// fixtures are — is downgraded from the silent pass/0 the pre-M4 table
// asserted to inconclusive/0. A widening that reads either as a stated
// verdict still fails this table.
var codex1718Expected = []codex1718Synthesis{
	{"S1.txt", codexMethodTurnStart, "fail", codex1718S1DeclaredFindings},
	{"S1.txt", codexMethodReviewStart, "fail", codex1718S1DeclaredFindings},
	{"S2.txt", codexMethodTurnStart, "fail", codex1718S2DeclaredFindings},
	{"S2.txt", codexMethodReviewStart, "fail", codex1718S2DeclaredFindings},
	{"S2p.txt", codexMethodTurnStart, "fail", codex1718S2DeclaredFindings},
	{"S2p.txt", codexMethodReviewStart, "fail", codex1718S2DeclaredFindings},
	{"N1.txt", codexMethodTurnStart, VerdictInconclusive, 0},
	{"N1.txt", codexMethodReviewStart, VerdictInconclusive, 0},
	{"N2.txt", codexMethodTurnStart, VerdictInconclusive, 0},
	{"N2.txt", codexMethodReviewStart, VerdictInconclusive, 0},
}

func readCodex1718Fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(codex1718Dir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(b)
}

// TestCodex1718Fixtures synthesizes every fixture file on both review paths,
// one SYNTH line per file per path, and each row must equal codex1718Expected.
// It was AC-CPS-011 check 1 (fidelity) at the commit that introduced the
// fixtures, and is the AC-CPS-012 verdict/count table since candidate (a).
func TestCodex1718Fixtures(t *testing.T) {
	if len(codex1718Expected) != 10 {
		t.Fatalf("expected table has %d rows, want 10 (five files × two paths)", len(codex1718Expected))
	}
	for _, want := range codex1718Expected {
		body := readCodex1718Fixture(t, want.file)
		out := synthesizeReviewOutput(body, want.method)
		t.Logf("SYNTH %s %s verdict=%s findings=%d", want.file, want.method, out.Verdict, len(out.Findings))
		if out.Verdict != want.verdict || len(out.Findings) != want.findings {
			t.Errorf("%s on %s: got %s/%d, want %s/%d", want.file, want.method,
				out.Verdict, len(out.Findings), want.verdict, want.findings)
		}
	}
}

// Structural property checks mirrored from the AC-CPS-011 ledger, so the test
// itself refuses a reduction that has the right outputs but not the shape.
var (
	codex1718P4  = regexp.MustCompile(`(?m)^\|[^|\n]*\*\*(Critical|High|Medium|Low)\*\*[^|\n]*\|`)
	codex1718P5  = regexp.MustCompile(`(?m)^- \*\*(Critical|High|Medium|Low) · \[[^]\n]+:[0-9]+\]\(<[^>\n]+>\)`)
	codex1718P11 = regexp.MustCompile(`(?m)^\|[^|\n]*\*\*(Critical|High|Medium|Low)\*\*[^|\n]*\|[^|\n]*\[[^]\n]+:[0-9]+\]\(`)
	// N7 (plan-audit iter-3): P10 is bound to the FIRST line — the verdict
	// statement — rather than satisfied by any line pairing the label with a
	// bold FAIL, and a bold PASS under the label anywhere rejects the file.
	codex1718P10Statement = regexp.MustCompile(`^[^|]*판정[^|]*\*\*FAIL\*\*`)
	codex1718P10PassNeg   = regexp.MustCompile(`판정[^|\n]*\*\*PASS\*\*`)
)

// codex1718StatesFail decides AC-CPS-011's P10 property, "the S1 prose states
// FAIL under the localized label": line 1 carries the label followed by a bold
// FAIL outside any table cell, and no line anywhere pairs the label with a
// bold PASS.
func codex1718StatesFail(body string) bool {
	first, _, _ := strings.Cut(body, "\n")
	return codex1718P10Statement.MatchString(first) && !codex1718P10PassNeg.MatchString(body)
}

func TestCodex1718Fixtures_StructuralProperties(t *testing.T) {
	s1 := readCodex1718Fixture(t, "S1.txt")
	s2 := readCodex1718Fixture(t, "S2.txt")
	s2p := readCodex1718Fixture(t, "S2p.txt")

	if got := len(codex1718P4.FindAllString(s1, -1)); got != codex1718S1DeclaredFindings {
		t.Errorf("P4: S1 has %d bold-severity table rows, want the declared %d", got, codex1718S1DeclaredFindings)
	}
	if got := len(codex1718P11.FindAllString(s1, -1)); got != codex1718S1DeclaredFindings {
		t.Errorf("P11: S1 has %d linked bold-severity rows, want the declared %d", got, codex1718S1DeclaredFindings)
	}
	if got := len(codex1718P5.FindAllString(s2, -1)); got != codex1718S2DeclaredFindings {
		t.Errorf("P5: S2 has %d linked bold-severity bullets, want the declared %d", got, codex1718S2DeclaredFindings)
	}
	if !codex1718StatesFail(s1) {
		t.Errorf("P10: S1 line 1 does not state **FAIL** under 판정, or a **PASS** is stated under it")
	}

	// P7: S2p is S2 with the greeting prefix of line 1 — up to and including
	// the first ", " — removed, and nothing else changed.
	s2First, s2Rest, _ := strings.Cut(s2, "\n")
	s2pFirst, s2pRest, _ := strings.Cut(s2p, "\n")
	_, afterGreeting, found := strings.Cut(s2First, ", ")
	if !found || s2pFirst != afterGreeting {
		t.Errorf("P7: S2p line 1 = %q, want S2 line 1 without its greeting prefix = %q", s2pFirst, afterGreeting)
	}
	if s2Rest != s2pRest {
		t.Errorf("P7: S2p differs from S2 beyond line 1")
	}
}

// TestCodex1718_P10RejectsMutants pins the N7 repair: the iter-3 mutants that
// passed the unbound P10 (M3 — the verdict statement only inside a table cell;
// M4 — 판정은 **PASS** in prose plus a separate line mentioning a bold FAIL)
// are rejected, and the faithful S1 is still accepted.
func TestCodex1718_P10RejectsMutants(t *testing.T) {
	s1 := readCodex1718Fixture(t, "S1.txt")
	_, rest, _ := strings.Cut(s1, "\n")

	m3 := "안녕하세요 팀, 이번 변경을 검토했어.\n" + rest +
		"| **Low** | [docs/guide.md:1](docs/guide.md:1) · 낮음 | 판정은 **FAIL**이야. |\n"
	m4 := "안녕하세요 팀, 리뷰어의 판정은 **PASS**야.\n" + rest +
		"판정 기준상 **FAIL** 사유는 없어.\n"

	cases := []struct {
		name string
		body string
		want bool
	}{
		{"faithful S1", s1, true},
		{"M3 verdict only in a table cell", m3, false},
		{"M4 PASS stated, FAIL mentioned elsewhere", m4, false},
	}
	for _, c := range cases {
		if got := codex1718StatesFail(c.body); got != c.want {
			t.Errorf("%s: codex1718StatesFail = %v, want %v", c.name, got, c.want)
		}
	}
}
