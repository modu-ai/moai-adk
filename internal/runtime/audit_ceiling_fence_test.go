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

// TestAppendProgressRecordGateInputsSingleSectionG (round-4 edge 7b) —
// both gate inputs leave EXACTLY ONE §G section after the append:
//
//	gate input A — a line whose backtick fence info string contains a
//	backtick is INLINE CODE, not a fence open (CommonMark); a phantom open
//	here swallowed the real §G and duplicated the section at end-of-file.
//	gate input B — a closed list-item fence (opener and closer indented to
//	the item's content column) opens and closes cleanly; the real §G after
//	it is found.
func TestAppendProgressRecordGateInputsSingleSectionG(t *testing.T) {
	cases := []struct {
		name string
		pre  string
	}{
		{
			name: "backtick_info_string_with_backtick_is_inline_code",
			pre:  "# progress\n\n```json with `quotes` inside\nnot actually fenced\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nbody\n",
		},
		{
			// gate round-37 edge 8: inline triple-backtick code
			// (```example```) is not an unclosed fence opener — the same
			// info-string rule, whose RED face is shared with input A
			// (both misread as a phantom open at the pre-fix tree).
			name: "inline_triple_backtick_code_is_not_a_fence_open",
			pre:  "# progress\n\n```example``` inline code line\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nbody\n",
		},
		{
			name: "closed_list_item_fence",
			pre:  "# progress\n\n- item\n  ```text\n  fenced\n  ```\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nbody\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			specDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(tc.pre), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := appendProgressRecord(specDir, "- new record"); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
			if err != nil {
				t.Fatal(err)
			}
			content := string(raw)
			if got := strings.Count(content, progressSectionHeading); got != 1 {
				t.Fatalf("§G heading appears %d times, want 1 (no duplicate section):\n%s", got, content)
			}
			lines := strings.Split(content, "\n")
			for i, l := range lines {
				if strings.HasPrefix(l, "## §E.2") {
					if lines[i-1] != "- new record" {
						t.Fatalf("the record did not land at the real §G block's end (line before §E.2 is %q):\n%s", lines[i-1], content)
					}
					return
				}
			}
			t.Fatalf("no §E.2 heading found:\n%s", content)
		})
	}
}

// TestAppendProgressRecordStartHeadingSkipsListFence (consolidated item 4,
// gate round-38) — a fenced block INSIDE a list item ("- ```text" … "  ```")
// is one fence: the item-content fence opener must be recognized and the
// indented closer must close it, or the closer is misread as an opener, a
// phantom swallows the real §G, and a duplicate section appears at
// end-of-file.
func TestAppendProgressRecordStartHeadingSkipsListFence(t *testing.T) {
	specDir := t.TempDir()
	pre := "# progress\n\n- ```text\n  fenced item content\n  ```\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nbody\n"
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
	content := string(raw)
	if got := strings.Count(content, progressSectionHeading); got != 1 {
		t.Fatalf("§G heading appears %d times, want 1 (the list-item fence swallowed the real one):\n%s", got, content)
	}
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

// TestAppendProgressRecordClosesOrderedListFence (gate round-46 item 6)
// — the closer of a `10. ```text` opener sits at the ITEM's content
// column (marker width 3 + one space = indent 4): the closer's indent is
// judged RELATIVE to the opener's content column, or the never-closed
// fence swallows the real §G and duplicates the section at end-of-file
// (Goldmark renders the block correctly — match that).
func TestAppendProgressRecordClosesOrderedListFence(t *testing.T) {
	specDir := t.TempDir()
	pre := "# progress\n\n10. ```text\n    fenced item content\n    ```\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nbody\n"
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
	content := string(raw)
	if got := strings.Count(content, progressSectionHeading); got != 1 {
		t.Fatalf("§G heading appears %d times, want 1 (the ordered-list fence swallowed the real one):\n%s", got, content)
	}
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
