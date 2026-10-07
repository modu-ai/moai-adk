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

// TestAppendProgressRecordStartHeadingSkipsFence (round-3 repair 1) — the
// scan that FINDS the §G heading tracks no fence state: a fenced example
// early in progress.md carrying a §G-heading-prefixed line gets selected
// as §G, and the record lands before the real section. After the fix the
// real heading is selected.
func TestAppendProgressRecordStartHeadingSkipsFence(t *testing.T) {
	specDir := t.TempDir()
	pre := "# progress\n\n```text\n## §G Override and Refusal Record (fenced example)\n```\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nlater section body\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	realIdx, recordIdx, eIdx := -1, -1, -1
	for i, l := range lines {
		switch {
		case l == progressSectionHeading && realIdx < 0:
			realIdx = i
		case l == "- new record" && recordIdx < 0:
			recordIdx = i
		case strings.HasPrefix(l, "## §E.2") && eIdx < 0:
			eIdx = i
		}
	}
	if realIdx < 0 || recordIdx < 0 || eIdx < 0 {
		t.Fatalf("fixture anchors missing (real %d record %d e2 %d):\n%s", realIdx, recordIdx, eIdx, raw)
	}
	if recordIdx < realIdx {
		t.Fatalf("the fenced example was selected as §G — the record landed before the real heading (record %d < real %d):\n%s", recordIdx, realIdx, raw)
	}
	if lines[eIdx-1] != "- new record" {
		t.Fatalf("the line immediately before the next real heading is %q, want the record:\n%s", lines[eIdx-1], raw)
	}
}

// TestAppendProgressRecordStartHeadingSkipsIndentedCodeBlock (round-4
// edge 5) — a fence marker indented 4+ spaces is an INDENTED CODE BLOCK in
// Markdown, not a fence: the fence-state tracker must not treat it as a
// fence open, or the real §G heading after it is swallowed by a never-
// closed "fence" and the record lands at end-of-file.
func TestAppendProgressRecordStartHeadingSkipsIndentedCodeBlock(t *testing.T) {
	specDir := t.TempDir()
	pre := "# progress\n\n    ```\n    indented code block content\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nlater section body\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	realIdx, recordIdx, eIdx := -1, -1, -1
	for i, l := range lines {
		switch {
		case l == progressSectionHeading && realIdx < 0:
			realIdx = i
		case l == "- new record" && recordIdx < 0:
			recordIdx = i
		case strings.HasPrefix(l, "## §E.2") && eIdx < 0:
			eIdx = i
		}
	}
	if realIdx < 0 || recordIdx < 0 || eIdx < 0 {
		t.Fatalf("fixture anchors missing (real %d record %d e2 %d):\n%s", realIdx, recordIdx, eIdx, raw)
	}
	if recordIdx < realIdx {
		t.Fatalf("the record landed before the real §G heading (record %d < real %d) — the indented code block swallowed it:\n%s", recordIdx, realIdx, raw)
	}
	if lines[eIdx-1] != "- new record" {
		t.Fatalf("the line immediately before the next real heading is %q, want the record:\n%s", lines[eIdx-1], raw)
	}
}
