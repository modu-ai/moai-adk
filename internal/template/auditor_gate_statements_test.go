package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M6 (AC-AEC-005, REQ-AEC-014; card t2017): the
// plan-auditor and sync-auditor definitions state the real gate behavior.
//
// Three sentences were wrong: that an unmet required gate returns verdict fail on
// the single-backend path, and that a non-empty gate_unmet makes the overall verdict
// fail on the convergence path. The code keeps an unmet gate at inconclusive unless a
// backend reported a defect (or the gate configuration cannot be read). The same
// statements have to hold in every copy: the two agent sources, the template Claude
// and Codex files rendered from them, and the four local definitions the emitter does
// not render. The test reads all ten files directly, so a fix that touches only the
// sources, only the .claude copies, or only the template copies stays red.

var auditorGateFiles = func() []string {
	var out []string
	for _, name := range []string{"plan-auditor", "sync-auditor"} {
		out = append(out,
			"internal/template/agentemit/source/"+name+".md",
			"internal/template/templates/.claude/agents/moai/"+name+".md",
			"internal/template/templates/.codex/agents/moai/"+name+".toml",
			".claude/agents/moai/"+name+".md",
			".codex/agents/moai/"+name+".toml",
		)
	}
	return out
}()

// The three wrong sentences (acceptance evidence EV-10, EV-11, EV-20).
var auditorWrongSentences = []string{
	"returns `verdict: fail` with a non-empty",
	"gate_unmet` makes the overall verdict `fail`",
	"the overall verdict is `fail`",
}

// The positive statements, each compared with line breaks folded to single spaces.
var auditorPositiveStatements = []struct{ name, text string }{
	{"an unmet gate with no reported defect is inconclusive and exported as INCONCLUSIVE",
		"the result is `inconclusive` with a non-empty `gate_unmet`, yields no PASS, and the verdict file you export says `INCONCLUSIVE`"},
	{"a reported fail stays fail", "When a required backend reported `fail`, the result stays `fail`"},
	{"the single-backend path keeps an unmet gate inconclusive",
		"keeps the verdict `inconclusive` with a non-empty `gate_unmet`"},
	{"the unreadable-configuration fail case is kept", "returns `fail` only when the gate configuration cannot be read"},
	{"the verdict machine line lists INCONCLUSIVE", "verdict: <PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE>"},
	{"the last-line template lists INCONCLUSIVE", "AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE> spec="},
	{"a FAIL file states blocking_count", "`blocking_count: <n>`"},
	{"a FAIL file names each blocking defect", "one `blocking_defect: <id>` line per BLOCKING defect"},
}

func foldedAuditorFile(t *testing.T, rel string) (raw, folded string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b), strings.Join(strings.Fields(string(b)), " ")
}

func TestAEC005_AuditorDefinitions_ContainNoneOfTheWrongSentences(t *testing.T) {
	for _, rel := range auditorGateFiles {
		raw, _ := foldedAuditorFile(t, rel)
		for _, wrong := range auditorWrongSentences {
			if n := strings.Count(raw, wrong); n != 0 {
				t.Errorf("%s still carries the wrong sentence %q (%d times)", rel, wrong, n)
			}
		}
	}
}

func TestAEC005_AuditorDefinitions_StateTheRealBehavior(t *testing.T) {
	for _, rel := range auditorGateFiles {
		_, folded := foldedAuditorFile(t, rel)
		for _, p := range auditorPositiveStatements {
			if !strings.Contains(folded, p.text) {
				t.Errorf("%s lacks the statement: %s (%q)", rel, p.name, p.text)
			}
		}
	}
}

// The Codex files carry the text inside a TOML literal string; a triple-quote
// sequence anywhere inside it would end the string early. Each file has exactly its
// opening and closing delimiters of the two literal strings it had before.
func TestAEC005_CodexDefinitions_KeepTheirLiteralStringDelimiters(t *testing.T) {
	for _, rel := range auditorGateFiles {
		if !strings.HasSuffix(rel, ".toml") {
			continue
		}
		raw, _ := foldedAuditorFile(t, rel)
		if n := strings.Count(raw, "'''"); n != 4 {
			t.Errorf("%s carries %d triple-quote sequences, want the 4 it had (two literal strings)", rel, n)
		}
	}
}

// Control, passes today: the unreadable-configuration fail wording is not the wrong
// sentence, so the guard cannot be turned red by keeping it.
func TestAEC005_ControlsThePatternsAreNarrow(t *testing.T) {
	kept := "it returns `fail` only when the gate configuration cannot be read"
	for _, wrong := range auditorWrongSentences {
		if strings.Contains(kept, wrong) {
			t.Errorf("the kept sentence matches the wrong-sentence pattern %q", wrong)
		}
	}
}
