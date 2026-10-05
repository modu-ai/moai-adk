// budget_test.go — the MEMORY.md size measurement and index budget audit of
// SPEC-MEMORY-FOLD-BUDGET-001 (REQ-MFB-008 / REQ-MFB-009, plan.md M1).
package taxonomy

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestMeasureIndex_Counts(t *testing.T) {
	t.Parallel()
	// "a" + "한글" (two 3-byte runes) + "\n": 8 bytes, 4 code points, 1 line.
	m := MeasureIndex([]byte("a한글\n"))
	if m.Bytes != 8 {
		t.Errorf("Bytes = %d, want 8", m.Bytes)
	}
	if m.Chars != 4 {
		t.Errorf("Chars = %d, want 4 (Unicode code points, not bytes)", m.Chars)
	}
	if m.Lines != 1 {
		t.Errorf("Lines = %d, want 1", m.Lines)
	}
	if m.LoadedChars != m.Chars {
		t.Errorf("LoadedChars = %d, want %d (no frontmatter, no comment)", m.LoadedChars, m.Chars)
	}
}

// TestMeasureIndex_LinesMatchDoctorCounting pins the line measure to the
// doctor's existing count (index_lines must not move: AC-MFB-009).
func TestMeasureIndex_LinesMatchDoctorCounting(t *testing.T) {
	t.Parallel()
	cases := []struct {
		content string
		want    int
	}{
		{"a\nb\n", 2},
		{"a\nb", 2},
		{"a\nb\n\n", 2}, // trailing blank line: the doctor's TrimRight drops it
		{"", 1},         // Split("") yields one empty element, exactly as the doctor counts
	}
	for _, tc := range cases {
		if got := MeasureIndex([]byte(tc.content)).Lines; got != tc.want {
			t.Errorf("Lines(%q) = %d, want %d", tc.content, got, tc.want)
		}
	}
}

func TestMeasureIndex_LoadedChars(t *testing.T) {
	t.Parallel()
	front := "---\nname: x\ndescription: d\ntype: feedback\n---\n"
	comment := "<!-- loader ignores this -->"
	body := "body text\n" + comment + "tail\n"

	m := MeasureIndex([]byte(front + body))
	if want := utf8.RuneCountInString("body text\ntail\n"); m.LoadedChars != want {
		t.Errorf("LoadedChars = %d, want %d (frontmatter and HTML comment removed)", m.LoadedChars, want)
	}
	if want := utf8.RuneCountInString(front + body); m.Chars != want {
		t.Errorf("Chars = %d, want %d", m.Chars, want)
	}
	if want := len(front + body); m.Bytes != want {
		t.Errorf("Bytes = %d, want %d", m.Bytes, want)
	}
}

func TestMeasureIndex_LoadedChars_NoFrontmatter(t *testing.T) {
	t.Parallel()
	m := MeasureIndex([]byte("plain\n<!-- c -->\n"))
	if want := utf8.RuneCountInString("plain\n\n"); m.LoadedChars != want {
		t.Errorf("LoadedChars = %d, want %d", m.LoadedChars, want)
	}
}

func TestMeasureIndex_LoadedChars_UnclosedFrontmatter(t *testing.T) {
	t.Parallel()
	// No closing delimiter: nothing is excluded — the whole content counts.
	content := "---\nname: x\nbody\n"
	m := MeasureIndex([]byte(content))
	if m.LoadedChars != m.Chars {
		t.Errorf("LoadedChars = %d, want %d (unclosed frontmatter removes nothing)", m.LoadedChars, m.Chars)
	}
}

// TestAuditIndexBudget_ByteBoundaries walks the fixture-scale boundaries of
// AC-MFB-010: 1444 (79.98%) none, 1443 (80.04%) warn, 1155 (100%) at-cap.
func TestAuditIndexBudget_ByteBoundaries(t *testing.T) {
	t.Parallel()
	m := IndexMeasurements{Bytes: 1155, Lines: 1}

	if f := AuditIndexBudget("MEMORY.md", m, 1444, 80, 200); len(f) != 0 {
		t.Errorf("79.98%% of cap emitted %v, want none", f)
	}

	f := AuditIndexBudget("MEMORY.md", m, 1443, 80, 200)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetWarn {
		t.Errorf("80.04%% of cap = %v, want one MEMORY_INDEX_BUDGET_WARN", f)
	}

	f = AuditIndexBudget("MEMORY.md", m, 1155, 80, 200)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetAtCap {
		t.Errorf("100%% of cap = %v, want one MEMORY_INDEX_BUDGET_AT_CAP replacing the warning", f)
	}
}

func TestAuditIndexBudget_ByteOverCap(t *testing.T) {
	t.Parallel()
	m := IndexMeasurements{Bytes: 1300}
	f := AuditIndexBudget("MEMORY.md", m, 1300, 80, 200)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetAtCap {
		t.Errorf("bytes above cap = %v, want MEMORY_INDEX_BUDGET_AT_CAP only", f)
	}
}

func TestAuditIndexBudget_LineBoundaries(t *testing.T) {
	t.Parallel()
	m := IndexMeasurements{Bytes: 1, Lines: 17}

	if f := AuditIndexBudget("MEMORY.md", m, 25000, 80, 22); len(f) != 0 {
		t.Errorf("17 lines of a 22-line cap (77.3%%) emitted %v, want none", f)
	}

	f := AuditIndexBudget("MEMORY.md", m, 25000, 80, 21)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetWarn || !strings.Contains(f[0].Detail, "17") {
		t.Errorf("17 lines of a 21-line cap = %v, want one lines-axis MEMORY_INDEX_BUDGET_WARN naming the value", f)
	}

	// 17 lines of a 17-line cap: at the cap, not above it — the lines-axis
	// warning fires and MEMORY_INDEX_OVERFLOW does not (that is AuditIndex's,
	// not the budget audit's).
	f = AuditIndexBudget("MEMORY.md", m, 25000, 80, 17)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetWarn {
		t.Errorf("17 lines of a 17-line cap = %v, want the lines-axis warning only", f)
	}

	// 17 lines of a 16-line cap: above the cap — the budget audit emits no
	// lines finding (MEMORY_INDEX_OVERFLOW owns that case).
	f = AuditIndexBudget("MEMORY.md", m, 25000, 80, 16)
	if len(f) != 0 {
		t.Errorf("17 lines above a 16-line cap = %v, want nothing from the budget audit", f)
	}
}

// TestAuditIndexBudget_DefaultsFollowConfig ties the flagless invocation to
// the config constants by reference, never to literals.
func TestAuditIndexBudget_DefaultsFollowConfig(t *testing.T) {
	t.Parallel()
	capValue := config.DefaultMemoryIndexByteCap
	warnPct := config.DefaultMemoryIndexWarnPercent

	atCap := IndexMeasurements{Bytes: capValue}
	f := AuditIndexBudget("MEMORY.md", atCap, 0, 0, 0)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetAtCap {
		t.Errorf("bytes at DefaultMemoryIndexByteCap = %v, want AT_CAP", f)
	}

	warnAt := capValue * warnPct / 100
	f = AuditIndexBudget("MEMORY.md", IndexMeasurements{Bytes: warnAt}, 0, 0, 0)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetWarn {
		t.Errorf("bytes at %d%% of DefaultMemoryIndexByteCap = %v, want WARN", warnPct, f)
	}

	f = AuditIndexBudget("MEMORY.md", IndexMeasurements{Bytes: warnAt - 1}, 0, 0, 0)
	if len(f) != 0 {
		t.Errorf("bytes one below the warn point = %v, want none", f)
	}
}

func TestAuditIndexBudget_InvalidWarnPercentFallsBack(t *testing.T) {
	t.Parallel()
	// warnPercent 150 can never fire on its own axis (90% of cap < 150%);
	// a falling-back 80 does fire at 90% of the cap.
	m := IndexMeasurements{Bytes: 90}
	f := AuditIndexBudget("MEMORY.md", m, 100, 150, 200)
	if len(f) != 1 || f[0].Code != WarnIndexBudgetWarn {
		t.Errorf("warnPercent 150 = %v, want one WARN from the config fallback (80)", f)
	}
}

// TestAuditIndexBudget_FindingText: every budget finding names the axis, the
// value, the cap, the percentage and the basis sentence.
func TestAuditIndexBudget_FindingText(t *testing.T) {
	t.Parallel()
	f := AuditIndexBudget("MEMORY.md", IndexMeasurements{Bytes: 1155, Lines: 17}, 1443, 80, 21)
	if len(f) != 2 {
		t.Fatalf("findings = %v, want the byte warn and the lines warn", f)
	}
	byteFinding, lineFinding := f[0], f[1]

	for _, finding := range f {
		if finding.Path != "MEMORY.md" {
			t.Errorf("Path = %q, want MEMORY.md", finding.Path)
		}
		for _, part := range []string{"raw bytes", "conservative proxy", "unconfirmed", "80"} {
			if !strings.Contains(finding.Detail, part) {
				t.Errorf("detail %q does not name %q", finding.Detail, part)
			}
		}
	}
	for _, part := range []string{"1155", "1443"} {
		if !strings.Contains(byteFinding.Detail, part) {
			t.Errorf("byte finding %q does not name %q", byteFinding.Detail, part)
		}
	}
	for _, part := range []string{"17", "21"} {
		if !strings.Contains(lineFinding.Detail, part) {
			t.Errorf("line finding %q does not name %q", lineFinding.Detail, part)
		}
	}
}
