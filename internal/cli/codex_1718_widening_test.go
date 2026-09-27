package cli

import (
	"strings"
	"testing"
)

// SPEC-CODEX-PARSER-SHAPE-001 candidate (a) on the #1718 shapes (AC-CPS-012,
// AC-CPS-006): three recognizer changes, each measured here rather than
// inherited from the numbered-list result —
//
//  1. a verdict label that follows a greeting on its line ("<greeting>,
//     **verdict: fail**");
//  2. a localized verdict label ("판정은 **FAIL**");
//  3. findings written as markdown table rows or bold severity-word bullets
//     led by a bold severity word (Critical / High / Medium / Low).

// codex1718S1Findings is S1's declared finding content: severity, message,
// file and line for every row, fixed here independently of the parser.
var codex1718S1Findings = []Finding{
	{
		Severity: "High",
		Title:    "[docs/guide.md:28](docs/guide.md:28), [docs/intro.md:3](docs/intro.md:3) · 높음 — 판매자 유형 하나에만 권한 확인 경로를 제시해. 다른 유형의 진입 경로도 함께 안내해.",
		File:     "docs/guide.md",
		Line:     28,
	},
	{
		Severity: "Medium",
		Title:    "[docs/guide.md:14](docs/guide.md:14), [docs/guide.md:53](docs/guide.md:53) · 높음 — 앞의 확인 목록과 뒤의 중단 조건이 어긋나. 간단한 요청까지 중단될 수 있으니 필수 항목만 요구해.",
		File:     "docs/guide.md",
		Line:     14,
	},
	{
		Severity: "Medium",
		Title:    "[config/checks.yaml:9](config/checks.yaml:9) · 높음 — 검증 사례에 실제 값이 없어서 일치 여부를 가릴 기준이 없어. 구체적인 테스트 값을 넣어.",
		File:     "config/checks.yaml",
		Line:     9,
	},
}

// codex1718S2Findings is S2's (and S2p's) declared finding content.
var codex1718S2Findings = []Finding{
	{
		Severity: "Medium",
		Title:    "[config/checks.yaml:38](<config/checks.yaml:38>) · 확신 높음:** `missing_setup` 사례에는 요청 목적 입력이 없는데 설정값이 비었다는 이유만으로 `blocker`를 기대해. **권고:** 요청 목적을 입력에 명시하고 간단한 요청 사례를 따로 검사해.",
		File:     "config/checks.yaml",
		Line:     38,
	},
}

func assertCodexFindingsExact(t *testing.T, label string, got, want []Finding) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: want exactly %d findings, got %d (%+v)", label, len(want), len(got), got)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Severity != w.Severity || g.Title != w.Title || g.Body != w.Title || g.File != w.File || g.Line != w.Line {
			t.Errorf("%s: finding[%d]\n got  {sev=%q file=%q line=%d title=%q body=%q}\n want {sev=%q file=%q line=%d title=%q}",
				label, i, g.Severity, g.File, g.Line, g.Title, g.Body, w.Severity, w.File, w.Line, w.Title)
		}
	}
}

func TestCodex1718Fixtures_WidenedContent(t *testing.T) {
	cases := []struct {
		file string
		want []Finding
	}{
		{"S1.txt", codex1718S1Findings},
		{"S2.txt", codex1718S2Findings},
		{"S2p.txt", codex1718S2Findings},
	}
	for _, c := range cases {
		body := readCodex1718Fixture(t, c.file)
		for _, method := range []string{codexMethodTurnStart, codexMethodReviewStart} {
			out := synthesizeReviewOutput(body, method)
			if out.Verdict != "fail" {
				t.Errorf("%s on %s: verdict = %q, want fail", c.file, method, out.Verdict)
			}
			assertCodexFindingsExact(t, c.file+" on "+method, out.Findings, c.want)
			if out.Contradiction != "" {
				t.Errorf("%s on %s: findings were recovered, so no contradiction may be reported: %q", c.file, method, out.Contradiction)
			}
		}
	}
}

// TestCodex1718_PartialDriftFailsExactCount is AC-CPS-012's partial-drift case:
// one finding of S1 or S2 rewritten into a shape the widened recognizers do
// NOT accept makes the exact count fail rather than pass on the survivors.
func TestCodex1718_PartialDriftFailsExactCount(t *testing.T) {
	s1 := readCodex1718Fixture(t, "S1.txt")
	s1Drift := strings.Replace(s1, "| **Medium** | [docs/guide.md:14]", "| Medium (중간) | [docs/guide.md:14]", 1)
	if s1Drift == s1 {
		t.Fatal("drift rewrite of S1 did not apply")
	}
	s2 := readCodex1718Fixture(t, "S2.txt")
	s2Drift := strings.Replace(s2, "- **Medium · [config/checks.yaml:38]", "- Medium · [config/checks.yaml:38]", 1)
	if s2Drift == s2 {
		t.Fatal("drift rewrite of S2 did not apply")
	}

	cases := []struct {
		name     string
		body     string
		declared int
		survived int
	}{
		{"S1 one row drifted", s1Drift, codex1718S1DeclaredFindings, codex1718S1DeclaredFindings - 1},
		{"S2 one bullet drifted", s2Drift, codex1718S2DeclaredFindings, codex1718S2DeclaredFindings - 1},
	}
	for _, c := range cases {
		got := len(synthesizeReviewOutput(c.body, codexMethodTurnStart).Findings)
		if got == c.declared {
			t.Errorf("%s: exact count %d still matched after drift — the drift is invisible", c.name, got)
		}
		if got != c.survived {
			t.Errorf("%s: got %d findings, want exactly the %d survivors", c.name, got, c.survived)
		}
	}
}

// TestCodexWidenedVerdictRecognizers pins recognizer changes 1 and 2 and the
// narrowness contract they must keep: prose ABOUT a verdict is not a verdict.
func TestCodexWidenedVerdictRecognizers(t *testing.T) {
	positives := []struct {
		name string
		body string
		want string
	}{
		{"greeting then bold english label", "Hello team, **verdict: fail**. Two issues block this.", "fail"},
		{"greeting then plain english label", "Hi all, verdict: pass — nothing blocks the merge.", "pass"},
		{"localized label after greeting, bold word", "안녕하세요 팀, 리뷰어의 판정은 **FAIL**이야.", "fail"},
		{"localized label at line head with colon", "판정: pass", "pass"},
		{"localized label at line head, bold word", "판정은 **INCONCLUSIVE**", VerdictInconclusive},
	}
	for _, c := range positives {
		signals := codexVerdictSignalsOf(c.body)
		if got := adoptConservativeVerdict(signals); got != c.want {
			t.Errorf("%s: adopted verdict = %q (signals %+v), want %q\nbody: %s", c.name, got, signals, c.want, c.body)
		}
	}

	// Prose ABOUT a verdict is not a verdict: no signal at all may be read.
	negatives := []struct {
		name string
		body string
	}{
		{"english mention after greeting", "Hello team, the verdict on caching is still open and the old tests pass anyway."},
		{"english mention at line head", "Verdict on the retry path is hard to call; the flaky test may fail again."},
		{"localized mention without statement", "판정 기준상 **FAIL** 사유는 없어."},
		{"localized word unemphasized, no colon", "판정은 fail 여부를 가리기 어렵다."},
		{"localized statement inside a table cell", "| **Low** | [a.go:1](a.go:1) | 판정은 **FAIL**이야. |"},
	}
	for _, c := range negatives {
		if signals := codexVerdictSignalsOf(c.body); len(signals) != 0 {
			t.Errorf("%s: read %d verdict signal(s) %+v from prose that states no verdict\nbody: %s", c.name, len(signals), signals, c.body)
		}
	}
}

// TestCodexWidenedFindingRecognizers pins recognizer change 3 and its
// narrowness: a table header, a table without a bold severity cell, and a
// bold word that merely starts with a severity word are not findings.
func TestCodexWidenedFindingRecognizers(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"table rows with bold severity", "| Sev | Where | What |\n|---|---|---|\n| **High** | a.go:1 | x |\n| **Low** | b.go:2 | y |", 2},
		{"bold severity bullets", "- **Critical · a.go:1 · x\n- **Low:** b.go:2 y", 2},
		{"NEG table without bold severity", "| check | status |\n| --- | --- |\n| secrets | Blocking |", 0},
		{"NEG bold word starting with a severity word", "- **High-level summary** of the change", 0},
		{"NEG bold severity inside prose", "The risk is **High** overall.", 0},
	}
	for _, c := range cases {
		if got := len(codexFindingsOf(c.body)); got != c.want {
			t.Errorf("%s: %d findings, want %d", c.name, got, c.want)
		}
	}
}
