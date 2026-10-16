package update

import (
	"strings"
	"testing"
)

// The table cuts a cell wider than its column from the right, so a long relative
// path loses its file name: the part a user needs to recognize the row.
func TestPreviewTableKeepsFileNameVisible(t *testing.T) {
	const long = ".claude/skills/moai-workflow-testing/references/long-reference-name-for-width-check.md"
	model := newPreviewModel([]FilePreviewInput{{RelPath: long, Exists: true}}, nil, PreviewOptions{Interactive: true, Width: 80, Height: 24})
	view := model.table.View()
	if !strings.Contains(view, "long-reference-name-for-width-check.md") {
		t.Fatalf("table hid the file name of %q:\n%s", long, view)
	}
}

// The non-interactive fallback never pauses for a decision: renderFallbackToStdout
// returns confirmed=true unconditionally, so its conflict heading must not promise
// a gate that this path never applies.
func TestRenderFallbackConflictHeadingMakesNoGatePromise(t *testing.T) {
	out := renderFallback([]FileClassification{{RelPath: ".moai/config/config.yaml", Class: ClassConflict}}, false)
	if strings.Contains(out, "resolve before proceeding") {
		t.Fatalf("conflict heading promises a gate the non-interactive path never applies:\n%s", out)
	}
}

// elidePath keeps the file name whole and shortens the leading directories, so
// the table cell stays exactly the column width and still names the row.
func TestElidePathKeepsFileName(t *testing.T) {
	short := ".moai/config/config.yaml"
	if got := elidePath(short, 52); got != short {
		t.Fatalf("short path changed: %q", got)
	}
	long := ".claude/skills/moai-workflow-testing/references/long-reference-name-for-width-check.md"
	got := elidePath(long, 52)
	if n := len([]rune(got)); n != 52 {
		t.Fatalf("elided width = %d, want 52: %q", n, got)
	}
	if !strings.HasSuffix(got, "/long-reference-name-for-width-check.md") {
		t.Fatalf("file name not kept whole: %q", got)
	}
	if !strings.HasPrefix(got, ".claude/") {
		t.Fatalf("leading directory lost: %q", got)
	}
	// A name too long to fit keeps its end, which carries the extension.
	name := "averyveryveryveryveryveryveryveryveryveryveryveryverylongname.md"
	if got := elidePath(name, 20); len([]rune(got)) != 20 || !strings.HasSuffix(got, ".md") {
		t.Fatalf("bare long name = %q", got)
	}
}
