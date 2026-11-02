package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// okStatus matches the success status literal in either quote style.
var okStatus = regexp.MustCompile(`status:\s*['"]ok['"]`)

// TestPlanResearchFanoutGuard is a LEXICAL guard on the embedded
// plan-research-fanout.js: the repo has no JS test runner, and a Go test that
// shelled out to node would have to skip where node is absent. It proves the
// guard code is present in the shipped script, not that it behaves correctly;
// behavior was proven by a recorded node run (card t2007 verdict), which CI
// does not repeat.
func TestPlanResearchFanoutGuard(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, ".claude/workflows/plan-research-fanout.js")
	if err != nil {
		t.Fatalf("read plan-research-fanout.js: %v", err)
	}
	js := string(raw)

	// Lenses fallback: an empty or malformed args.lenses must not reach the cap.
	line, ok := between(js, "const LENSES", "\n")
	if !ok {
		t.Fatal("LENSES declaration not found")
	}
	pre, _ := between(js, "const DEFAULT_LENSES", "const LENSES")
	if guard := pre + line; !strings.Contains(guard, "Array.isArray(") || !strings.Contains(guard, "length > 0") {
		t.Error("lenses fallback does not check Array.isArray(...) and length > 0 before using args.lenses")
	}
	if !strings.Contains(line, "DEFAULT_LENSES") || !strings.Contains(line, ".slice(0, 4)") {
		t.Error("LENSES no longer falls back to DEFAULT_LENSES or lost the 4-lens cap")
	}

	// Zero surviving reports: must return insufficient_coverage, not synthesize.
	i := strings.Index(js, "validReports.length === 0")
	if i < 0 {
		t.Error("no validReports.length === 0 check: zero-valid-report path is unguarded")
	} else if end := min(len(js), i+240); !strings.Contains(js[i:end], "insufficient_coverage") {
		t.Error("zero-valid-report check does not return insufficient_coverage")
	}

	// Synthesizer result validation and explicit success status.
	// The status string alone is not enough (a header comment also names it):
	// the typeof/empty check must exist and its branch must return it.
	j := strings.Index(js, "typeof research_md !== 'string'")
	if j < 0 {
		t.Error("no typeof research_md !== 'string' check: a null/empty synthesizer result is returned as success")
	} else {
		end := min(len(js), j+240)
		branch := js[j:end]
		if !strings.Contains(branch, ".trim() === ''") || !strings.Contains(branch, "return {") || !strings.Contains(branch, "invalid_synthesis") {
			t.Error("synthesis check does not return invalid_synthesis for a non-string or blank result")
		}
		if loc := okStatus.FindStringIndex(js); loc != nil && loc[0] < j {
			t.Error("status: 'ok' is returned before the synthesis validation")
		}
	}
	if !okStatus.MatchString(js) {
		t.Error("success return carries no status: 'ok'")
	}
}
