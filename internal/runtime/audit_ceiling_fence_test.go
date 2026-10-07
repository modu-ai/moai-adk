package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAppendProgressRecordSkipsHeadingInsideFence (round-2 leader-ruled
// fold, audit_ceiling.go §G next-heading scan) — a `## `-prefixed line
// inside an open fenced code block is code, not a section boundary: the
// record lands at the END of the §G block, immediately before the first
// real heading after the fence closes. Covers both fence variants (```
// and ~~~) and same-char close discrimination (a ``` line does not close a
// tilde fence).
func TestAppendProgressRecordSkipsHeadingInsideFence(t *testing.T) {
	// Backtick fence: `## example` and `~~~` inside are fence content.
	specDir := t.TempDir()
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n\n```text\n## example inside backtick fence\n~~~\n```\n\n## §E.2 Run-phase Evidence\nlater section body\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	assertRecordAtSectionEnd(t, specDir)

	// Tilde fence: a ``` line inside is content; only a tilde run closes.
	specDir2 := t.TempDir()
	pre2 := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n\n~~~markdown\n## example inside tilde fence\n```\n~~~\n\n## §E.2 Run-phase Evidence\nlater section body\n"
	if err := os.WriteFile(filepath.Join(specDir2, "progress.md"), []byte(pre2), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir2, "- new record"); err != nil {
		t.Fatal(err)
	}
	assertRecordAtSectionEnd(t, specDir2)
}

// assertRecordAtSectionEnd asserts the §G block ends with the record line
// immediately before the next real section heading.
func assertRecordAtSectionEnd(t *testing.T, specDir string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "## §E.2") {
			if lines[i-1] != "- new record" {
				t.Fatalf("the line immediately before the next real heading is %q, want the record:\n%s", lines[i-1], content)
			}
			return
		}
	}
	t.Fatalf("no §E.2 heading found:\n%s", content)
}
