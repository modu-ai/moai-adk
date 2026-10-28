package fetchweb

import (
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// blocks_test.go — SPEC-FETCH-WEB-001 M1 (AC-FW-001). Golden-case tests for
// the block splitter: block kinds, heading paths, chrome removal, and aside
// separation (REQ-FW-001), plus the token-accounting seam (REQ-FW-002).
//
// Token counts everywhere in this package are APPROXIMATE: they come from
// jev.EstimateTokens (conservative 4-bytes-per-token), not an exact
// tokenizer, and every threshold comparison downstream of them inherits that
// approximation.

// goldenParagraphsHeadings: paragraphs under nested headings carry the
// ancestor heading sequence as their path.
func TestSplitBlocks_ParagraphsCarryHeadingPaths(t *testing.T) {
	html := `<html><body>
	<h1>Install</h1>
	<p>First paragraph of the install section.</p>
	<h2>Requirements</h2>
	<p>You need Go 1.23 or later.</p>
	<h3>Notes</h3>
	<p>See the release page.</p>
	</body></html>`
	blocks := SplitBlocks(html)
	if len(blocks) != 3 {
		t.Fatalf("want 3 paragraph blocks, got %d: %+v", len(blocks), blocks)
	}
	wantPaths := [][]string{
		{"Install"},
		{"Install", "Requirements"},
		{"Install", "Requirements", "Notes"},
	}
	for i, want := range wantPaths {
		if blocks[i].Kind != KindParagraph {
			t.Errorf("block %d: kind = %q, want %q", i, blocks[i].Kind, KindParagraph)
		}
		if !reflect.DeepEqual(blocks[i].HeadingPath, want) {
			t.Errorf("block %d: path = %v, want %v", i, blocks[i].HeadingPath, want)
		}
	}
	if blocks[0].Text != "First paragraph of the install section." {
		t.Errorf("block 0 text = %q, want verbatim paragraph text", blocks[0].Text)
	}
}

// A later same-or-higher level heading replaces deeper path entries.
func TestSplitBlocks_HeadingPopOnSameLevel(t *testing.T) {
	html := `<body>
	<h1>A</h1><h2>B</h2><p>under B</p><h2>C</h2><p>under C</p><h1>D</h1><p>under D</p>
	</body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 3 {
		t.Fatalf("want 3 blocks, got %d", len(blocks))
	}
	want := [][]string{{"A", "B"}, {"A", "C"}, {"D"}}
	for i, w := range want {
		if !reflect.DeepEqual(blocks[i].HeadingPath, w) {
			t.Errorf("block %d: path = %v, want %v", i, blocks[i].HeadingPath, w)
		}
	}
}

// List items become their own blocks; a nested list's text does not leak into
// the parent item (REQ-FW-001).
func TestSplitBlocks_ListItems(t *testing.T) {
	html := `<body><h2>Steps</h2><ul>
	<li>Clone the repository</li>
	<li>Run the installer
	  <ul><li>Nested step one</li><li>Nested step two</li></ul>
	</li>
	<li>Verify the install</li>
	</ul></body>`
	blocks := SplitBlocks(html)
	// 3 top-level items + the 2 nested items, which are blocks in their own
	// right (excluded from the parent's text, not from the document).
	if len(blocks) != 5 {
		t.Fatalf("want 5 list-item blocks, got %d: %+v", len(blocks), blocks)
	}
	for i, b := range blocks {
		if b.Kind != KindListItem {
			t.Errorf("block %d: kind = %q, want %q", i, b.Kind, KindListItem)
		}
	}
	if blocks[1].Text != "Run the installer" {
		t.Errorf("nested list leaked into parent item text: %q", blocks[1].Text)
	}
	if !reflect.DeepEqual(blocks[0].HeadingPath, []string{"Steps"}) {
		t.Errorf("list item path = %v, want [Steps]", blocks[0].HeadingPath)
	}
}

// Table rows become blocks whose cells are joined; the header row is itself a
// block and body rows point at it (REQ-FW-006's accompaniment input).
func TestSplitBlocks_TableRowsCarryHeader(t *testing.T) {
	html := `<body><table>
	<thead><tr><th>Flag</th><th>Meaning</th></tr></thead>
	<tbody>
	<tr><td>--json</td><td>Emit JSON</td></tr>
	<tr><td>--quiet</td><td>Silence output</td></tr>
	</tbody>
	</table></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 3 {
		t.Fatalf("want 3 table-row blocks, got %d: %+v", len(blocks), blocks)
	}
	for i, b := range blocks {
		if b.Kind != KindTableRow {
			t.Errorf("block %d: kind = %q, want %q", i, b.Kind, KindTableRow)
		}
	}
	if blocks[0].Text != "Flag | Meaning" {
		t.Errorf("header row text = %q, want %q", blocks[0].Text, "Flag | Meaning")
	}
	if blocks[1].Text != "--json | Emit JSON" {
		t.Errorf("row text = %q, want %q", blocks[1].Text, "--json | Emit JSON")
	}
	if blocks[1].HeaderIndex != 0 || blocks[2].HeaderIndex != 0 {
		t.Errorf("body rows must point at header block 0, got %d and %d",
			blocks[1].HeaderIndex, blocks[2].HeaderIndex)
	}
}

// A table with no header row (no th cells) leaves HeaderIndex unset.
func TestSplitBlocks_TableWithoutHeader(t *testing.T) {
	html := `<body><table><tr><td>a</td></tr><tr><td>b</td></tr></table></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 2 {
		t.Fatalf("want 2 table-row blocks, got %d", len(blocks))
	}
	if blocks[0].HeaderIndex != -1 || blocks[1].HeaderIndex != -1 {
		t.Errorf("headerless table rows must carry HeaderIndex -1, got %d/%d",
			blocks[0].HeaderIndex, blocks[1].HeaderIndex)
	}
}

// Code blocks come from <pre> and keep their internal line structure.
func TestSplitBlocks_CodeBlocksPreserveNewlines(t *testing.T) {
	html := `<body><h2>Usage</h2><pre><code>moai fetch web https://example.com "what is this?"
--json
</code></pre></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 {
		t.Fatalf("want 1 code block, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Kind != KindCode {
		t.Errorf("kind = %q, want %q", blocks[0].Kind, KindCode)
	}
	want := "moai fetch web https://example.com \"what is this?\"\n--json"
	if blocks[0].Text != want {
		t.Errorf("code text = %q, want %q (newlines preserved)", blocks[0].Text, want)
	}
}

// Page chrome — navigation, footers, scripts, styles, templates — contributes
// zero blocks (REQ-FW-001).
func TestSplitBlocks_ChromeContributesZeroBlocks(t *testing.T) {
	html := `<body>
	<nav><p>Home Docs Blog</p></nav>
	<header><p>Site header text</p></header>
	<script>var x = "not content";</script>
	<style>p { color: red; }</style>
	<template><p>template text</p></template>
	<footer><p>Copyright nobody</p></footer>
	</body>`
	if blocks := SplitBlocks(html); len(blocks) != 0 {
		t.Fatalf("chrome contributed %d blocks, want 0: %+v", len(blocks), blocks)
	}
}

// aside note content is emitted as blocks flagged Aside, in document order,
// so the scorer can treat it separately (REQ-FW-001).
func TestSplitBlocks_AsideEmittedSeparately(t *testing.T) {
	html := `<body>
	<h1>Main</h1>
	<p>Main flow paragraph.</p>
	<aside><p>Side note about the main flow.</p></aside>
	<p>Another main paragraph.</p>
	</body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 3 {
		t.Fatalf("want 3 blocks, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Aside {
		t.Errorf("block 0 must be main-flow")
	}
	if !blocks[1].Aside || blocks[1].Text != "Side note about the main flow." {
		t.Errorf("block 1 must be the aside note, got %+v", blocks[1])
	}
	if blocks[2].Aside {
		t.Errorf("block 2 must be main-flow")
	}
}

// Document order is preserved across all kinds.
func TestSplitBlocks_DocumentOrder(t *testing.T) {
	html := `<body><p>one</p><ul><li>two</li></ul><pre>three
</pre><table><tr><td>four</td></tr></table></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 4 {
		t.Fatalf("want 4 blocks, got %d", len(blocks))
	}
	// The pre block's surrounding newline is trimmed; interior structure
	// would be kept.
	want := []string{"one", "two", "three", "four"}
	for i, w := range want {
		if blocks[i].Text != w {
			t.Errorf("block %d = %q, want %q (document order)", i, blocks[i].Text, w)
		}
	}
}

// Whitespace inside paragraphs collapses to single spaces; words are not
// reworded (verbatim means no paraphrase, not no normalization).
func TestSplitBlocks_ParagraphWhitespaceCollapsed(t *testing.T) {
	html := `<body><p>Hello   there,
	world</p></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Text != "Hello there, world" {
		t.Fatalf("collapsed text = %+v, want one block %q", blocks, "Hello there, world")
	}
}

// R2 (card t1968 review): a <br> separates adjacent text runs — the original
// word boundary must survive the whitespace collapse, not merge into one word.
func TestSplitBlocks_BreakKeepsWordBoundary(t *testing.T) {
	html := `<body><p>first<br>second</p></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Text != "first second" {
		t.Fatalf("br-separated text = %v, want one block %q", texts(blocks), "first second")
	}
}

// R2: two paragraphs inside one table cell keep their word boundary in the
// joined row text (a block boundary inside an extracted text run).
func TestSplitBlocks_BlockBoundaryKeepsWordBoundary(t *testing.T) {
	html := `<body><table><tr><td><p>para one</p><p>para two</p></td></tr></table></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Text != "para one para two" {
		t.Fatalf("block-boundary text = %v, want one row %q", texts(blocks), "para one para two")
	}
}

// R3 (card t1968 review): a paragraph inside a list item is already part of
// the item's own text and must not be emitted a second time; nested lists
// still recurse into their own items.
func TestSplitBlocks_ListItemParagraphNotReEmitted(t *testing.T) {
	html := `<body><ul>
<li><p>only the item text</p></li>
<li>outer<ul><li><p>inner item</p></li></ul></li>
</ul></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindListItem, "only the item text"},
		{KindListItem, "outer"},
		{KindListItem, "inner item"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text {
			t.Errorf("block %d = %s %q, want %s %q", i, blocks[i].Kind, blocks[i].Text, w.kind, w.text)
		}
	}
}

// C2 (card t1968 review round 3): a <pre> or <table> inside a list item is
// emitted as its own block kind during the descent (code block, table rows) —
// it must not ALSO fold into the item's text, which duplicated the content
// and flattened the code's preserved whitespace (same principle as R3,
// extended to the separately-emitted block kinds).
func TestSplitBlocks_ListItemPreAndTableNotFoldedIntoText(t *testing.T) {
	html := `<body><ul>
	<li>intro<pre>code()</pre></li>
	<li>row text<table><tr><td>cell</td></tr></table></li>
	</ul></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindListItem, "intro"},
		{KindCode, "code()"},
		{KindListItem, "row text"},
		{KindTableRow, "cell"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text {
			t.Errorf("block %d = %s %q, want %s %q (no folded duplicate)", i, blocks[i].Kind, blocks[i].Text, w.kind, w.text)
		}
	}
}

// F3 (t1968 sync-audit): a skipped subtree INSIDE a list item's text (the
// nested list C2 excludes) must leave a boundary separator, like the chrome
// branch does — otherwise the text runs on either side merge across the hole
// ("FirstThird"), losing the word boundary. Document order of the surviving
// runs is preserved: the item's text reads "First Third" and the nested item
// stays its own block.
func TestSplitBlocks_ListItemSkipKeepsWordBoundary(t *testing.T) {
	html := `<body><ul><li>First<ul><li>Second</li></ul>Third</li></ul></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindListItem, "First Third"},
		{KindListItem, "Second"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text {
			t.Errorf("block %d = %s %q, want %s %q (word boundary across the skipped subtree)", i, blocks[i].Kind, blocks[i].Text, w.kind, w.text)
		}
	}
}

// F4 (t1968 sync-audit): chrome exclusion reached collapseText in N3 but not
// preText — a script/style subtree inside <pre> leaked its raw text into the
// emitted code block. The subtree is dropped with a boundary separator (the
// chrome-branch mechanism); the code text either side is unchanged.
func TestSplitBlocks_ScriptInsidePreIsChrome(t *testing.T) {
	html := `<body><pre>safe<script>hidden()</script>end</pre></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindCode {
		t.Fatalf("want one code block, got %d: %+v", len(blocks), blocks)
	}
	if got, want := blocks[0].Text, "safe end"; got != want {
		t.Errorf("code text = %q, want %q (chrome subtree dropped, word boundary kept, code text otherwise verbatim)", got, want)
	}
}

// R4 (card t1968 review): only surrounding BLANK LINES are trimmed from a
// code block; the first real code line keeps its indentation.
func TestSplitBlocks_CodeFirstLineIndentPreserved(t *testing.T) {
	html := `<body><pre>
    indented = true
    more()
</pre></body>`
	blocks := SplitBlocks(html)
	want := "    indented = true\n    more()"
	if len(blocks) != 1 || blocks[0].Text != want {
		t.Fatalf("code text = %q, want %q (first-line indent preserved)", texts(blocks), want)
	}
}

// P2-b (card t1968 codex review, repair round): a <br> inside pre/code is a
// line break in the code itself. The text-node walk used to skip the element
// silently, joining the runs on either side into one line
// ("echo firstecho second") — corrupted code semantics reaching every short
// page that passes through unfiltered. The break must extract as a newline
// boundary; all other verbatim spacing stands (D3/iter10 pre-verbatim).
func TestSplitBlocks_BreakInsidePreIsNewline(t *testing.T) {
	html := `<body><pre><code>echo first<br>echo second</code></pre></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindCode {
		t.Fatalf("want one code block, got %d: %+v", len(blocks), blocks)
	}
	if got, want := blocks[0].Text, "echo first\necho second"; got != want {
		t.Errorf("code text = %q, want %q (br as newline boundary)", got, want)
	}
	// <br/> parses to the same element — same boundary.
	html2 := `<body><pre><code>set -e<br/>cd /src</code></pre></body>`
	blocks2 := SplitBlocks(html2)
	if len(blocks2) != 1 || blocks2[0].Text != "set -e\ncd /src" {
		t.Errorf("code text = %q, want %q (br/ as newline boundary)", texts(blocks2), "set -e\ncd /src")
	}
}

// N3 (card t1968 review round 2): script/style subtrees nested INSIDE a
// paragraph, list item, or table cell are page chrome — the walk drops them
// at block level, and the recursive text extraction must drop them too, or
// their code/style text is scored as content and emitted in quotes. The
// surrounding text runs keep their word boundary across the removed subtree.
func TestSplitBlocks_ScriptStyleInsideBlocksAreChrome(t *testing.T) {
	html := `<body>
	<p>Visible price<script>var tracker=1;</script>stays</p>
	<ul><li>item one<style>.li{color:red}</style>continues</li></ul>
	<table><tr><td>value<noscript>no js</noscript>nine</td></tr></table>
	</body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 3 {
		t.Fatalf("want 3 blocks (p, li, row), got %d: %+v", len(blocks), blocks)
	}
	want := []string{"Visible price stays", "item one continues", "value nine"}
	for i, w := range want {
		if blocks[i].Text != w {
			t.Errorf("block %d text = %q, want %q (chrome subtree dropped, word boundary kept)", i, blocks[i].Text, w)
		}
	}
}

// N4 (card t1968 review round 2): a heading inside an aside is the aside's
// own structure — it must not leak onto the main-flow blocks that FOLLOW the
// aside (neither their provenance nor their same-heading rescue grouping).
// The aside's own blocks keep the heading; the main flow after it does not.
func TestSplitBlocks_AsideHeadingDoesNotLeakPath(t *testing.T) {
	html := `<body>
	<h2>Main</h2>
	<p>before the aside</p>
	<aside><h3>Side note</h3><p>aside body</p></aside>
	<p>after the aside</p>
	</body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 3 {
		t.Fatalf("want 3 blocks (the h3 is a heading, not a block), got %d: %+v", len(blocks), blocks)
	}
	// The aside's own body legitimately carries its heading under Main.
	if !reflect.DeepEqual(blocks[1].HeadingPath, []string{"Main", "Side note"}) || !blocks[1].Aside {
		t.Errorf("aside body path = %v (Aside=%v), want [Main Side note] (Aside=true)", blocks[1].HeadingPath, blocks[1].Aside)
	}
	// The main-flow block AFTER the aside inherits only its own ancestors.
	if !reflect.DeepEqual(blocks[2].HeadingPath, []string{"Main"}) {
		t.Errorf("post-aside block path = %v, want [Main] — the aside's heading leaked into the main flow", blocks[2].HeadingPath)
	}
}

// C1 (card t1968 review round 3): the aside's path save must be a true COPY,
// not the slice header. A SAME-LEVEL heading inside the aside truncates the
// stack and the following append reuses the freed slot of the shared backing
// array — overwriting the saved entry — so the main flow after the aside
// inherited the aside's heading. (The N4 test's h3-under-h2 never truncates,
// which is why it alone could not catch this.)
func TestSplitBlocks_AsideSameLevelHeadingRestoreIsACopy(t *testing.T) {
	html := `<body>
	<h2>Main</h2>
	<aside><h2>Note</h2><p>aside body</p></aside>
	<p>body</p>
	</body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 2 {
		t.Fatalf("want 2 blocks (aside body, then main body), got %d: %+v", len(blocks), blocks)
	}
	if !reflect.DeepEqual(blocks[1].HeadingPath, []string{"Main"}) {
		t.Errorf("post-aside body path = %v, want [Main] — the aside's same-level heading overwrote the saved stack", blocks[1].HeadingPath)
	}
}

// F7 (t1968 sync-audit iter2): an <aside> inside an <li> gets the SAME
// separation as at top level — the parent item's text excludes the aside
// subtree (the listItemSkip exclusion + boundary separator), and the aside's
// own paragraphs form aside-flagged blocks (the list-paragraph suppression
// is scoped per aside entry, not inherited from the enclosing item). Before
// the repair the aside's text rode the item block ("Main Note", aside=false)
// and never reached the aside scoring stream at all.
func TestSplitBlocks_AsideInsideListItemIsSeparate(t *testing.T) {
	html := `<body><ul><li>First<aside><p>Side note</p></aside>Third</li></ul></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind  BlockKind
		text  string
		aside bool
	}{
		{KindListItem, "First Third", false},
		{KindParagraph, "Side note", true},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || blocks[i].Aside != w.aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=%v (aside subtree scored separately, word boundary kept)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text, w.aside)
		}
	}
}

// F7b (t1968 sync-audit iter3): the F7 repair excluded the aside subtree from
// the item's text, but a BARE-TEXT aside (no block-level child) emitted no
// block of its own — the walk emits only through block children — so the
// content was lost from the document entirely (auditor probe: "Side note"
// absent everywhere). The fallback: an aside whose walk yields no separate
// block still emits its text as one aside-scored block; the item's exclusion
// and word boundary stay as F7 set them.
func TestSplitBlocks_BareTextAsideInsideListItemNotLost(t *testing.T) {
	html := `<body><ul><li>Main<aside>Side note</aside>Tail</li></ul></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind  BlockKind
		text  string
		aside bool
	}{
		{KindListItem, "Main Tail", false},
		{KindParagraph, "Side note", true},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || blocks[i].Aside != w.aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=%v (bare-text aside emitted as its own block, item exclusion kept)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text, w.aside)
		}
	}
}

// F7b control, the top-level shape named in the same finding (pre-existing
// never-emitted): a bare-text aside outside any item emits too — the fallback
// is the aside walk's own, not list-specific.
func TestSplitBlocks_BareTextAsideTopLevelEmitted(t *testing.T) {
	html := `<body><p>before</p><aside>Bare note</aside><p>after</p></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind  BlockKind
		text  string
		aside bool
	}{
		{KindParagraph, "before", false},
		{KindParagraph, "Bare note", true},
		{KindParagraph, "after", false},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || blocks[i].Aside != w.aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=%v (top-level bare-text aside emitted in document order)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text, w.aside)
		}
	}
}

// F7c (t1968 sync-audit iter4): the F7b fallback's collapseText(n, nil)
// included HEADING text, so an aside whose only child is a heading — which
// emits nothing at body level and emitted nothing at the base commit, before
// the fallback existed — manufactured a scored paragraph from pure structure
// (auditor probe: `<aside><h2>Note title</h2></aside>` → paragraph
// "Note title" aside=true; on a page whose only block this is, status flips
// empty→passthrough). The fallback extraction now skips heading subtrees; a
// heading-only remainder collapses to the empty string and the fallback's
// empty guard emits no block.
func TestSplitBlocks_HeadingOnlyAsideEmitsNothing(t *testing.T) {
	html := `<body><aside><h2>Note title</h2></aside></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 0 {
		t.Fatalf("want 0 blocks, got %d: %+v (a heading is structure, not a scored paragraph)", len(blocks), blocks)
	}
}

// F7c control (same finding): the heading's text merged into the fallback
// paragraph ("Side note body") where the p-wrapped control emits only the
// body's text ("note body"). The extraction now skips heading subtrees, so
// the fallback carries only the aside's non-heading text runs. C2 resolved
// the path residual this test once pinned ([Main] for the bare remainder
// where the p-wrapped control carried [Main / Side]): the bare run is now
// emitted in position, before the aside-local path is restored, so it
// carries the same [Main / Side] path the p-wrapped control does.
func TestSplitBlocks_AsideFallbackExcludesHeadingText(t *testing.T) {
	html := `<body><h2>Main</h2><aside><h3>Side</h3>note body</aside></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindParagraph || blocks[0].Text != "note body" || !blocks[0].Aside {
		t.Fatalf("want exactly one aside paragraph %q, got %d blocks: %+v", "note body", len(blocks), blocks)
	}
	if !reflect.DeepEqual(blocks[0].HeadingPath, []string{"Main", "Side"}) {
		t.Errorf("fallback path = %v, want [Main Side] (the aside-local heading precedes the run, same as the p-wrapped control)", blocks[0].HeadingPath)
	}
}

// R2 (t1968 sync-audit iter7/8 P3, debt-closure round): a heading that is a
// DIRECT child of an aside-scoped container splits the bare run at the
// heading boundary — the run BEFORE the heading carries the pre-heading path
// (it precedes the heading in the document, so the heading does not describe
// it), the run after carries the heading's path. Before the split, the
// text-only shape merged both sides into one paragraph riding the
// post-heading path ("lead tail" [Main Side]), so the pre-heading half was
// attributed a heading it comes before. The boundary mirrors what a
// block-emitting child does (C2), with one difference: a heading emits no
// block of its own, so the pre-heading run is appended, not spliced.
//
// Deliberate limit, kept from F7b: a heading nested inside a wrapper that
// emits no block does NOT split the enclosing run — that wrapper is pure
// inline content that merges into the enclosing paragraph
// (TestSplitBlocks_AsideBareOnlyWrapperMergesIntoRun); splitting inside it
// would contradict the merge rule. The split applies where the heading is
// the container's own child.
func TestSplitBlocks_AsideHeadingSplitsBareRunPaths(t *testing.T) {
	html := `<body><h2>Main</h2><aside>lead<h3>Side</h3>tail</aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		text string
		path []string
	}{
		{"lead", []string{"Main"}},
		{"tail", []string{"Main", "Side"}},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (the aside-internal heading must split the bare run)", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != KindParagraph || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true", i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, KindParagraph, w.text)
		}
		if !reflect.DeepEqual(blocks[i].HeadingPath, w.path) {
			t.Errorf("block %q path = %v, want %v (each side of the heading boundary carries its own path)", w.text, blocks[i].HeadingPath, w.path)
		}
	}
}

// R2 mixed control: the heading boundary composes with a following
// block-emitting child — three runs ("lead", "mid", "tail") land in document
// order around the paragraph block, each carrying the path of its own side of
// the boundary.
func TestSplitBlocks_AsideHeadingSplitComposesWithBlocks(t *testing.T) {
	html := `<body><h2>Main</h2><aside>lead<h3>Side</h3>mid<p>Body</p>tail</aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		text string
		path []string
	}{
		{"lead", []string{"Main"}},
		{"mid", []string{"Main", "Side"}},
		{"Body", []string{"Main", "Side"}},
		{"tail", []string{"Main", "Side"}},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %q aside=%v, want %q aside=true", i, blocks[i].Text, blocks[i].Aside, w.text)
		}
		if !reflect.DeepEqual(blocks[i].HeadingPath, w.path) {
			t.Errorf("block %q path = %v, want %v", w.text, blocks[i].HeadingPath, w.path)
		}
	}
}

// F1 (t1968 sync-audit iter11, P2): the R2 heading boundary descended into
// the heading's subtree but never counted that descent's emissions — a
// heading is not guaranteed text-only (`<h3><pre>CODE</pre></h3>` emits a
// code block). With `emitted` unset, a WRAPPER container read as pure inline
// content and silently dropped its trailing run after the heading:
// `<aside><div><h3><pre>CODE</pre></h3>TAIL</div></aside>` kept only CODE and
// lost TAIL. The heading branch now mirrors the default branch — an emission
// from the heading's subtree marks the container mixed, so the trailing run
// is content. It carries the heading's path (R2: a run after the heading is
// described by it), exactly as the aside-level split's post-heading run does.
func TestSplitBlocks_AsideWrapperHeadingWithBlockChildKeepsTrailingRun(t *testing.T) {
	html := `<body><aside><div><h3><pre>CODE</pre></h3>TAIL</div></aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
		path []string
	}{
		{KindCode, "CODE", []string{"CODE"}},
		{KindParagraph, "TAIL", []string{"CODE"}},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (a wrapper's trailing run after a block-carrying heading must not be dropped)",
			len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true", i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
		if !reflect.DeepEqual(blocks[i].HeadingPath, w.path) {
			t.Errorf("block %q path = %v, want %v (the trailing run follows the heading, so it carries the heading's path)", w.text, blocks[i].HeadingPath, w.path)
		}
	}
}

// F1 p-child variant: a p inside the heading emits the same way — the
// wrapper's trailing run after it is content, carrying the heading's path.
func TestSplitBlocks_AsideWrapperHeadingWithPChildKeepsTrailingRun(t *testing.T) {
	html := `<body><aside><div><h3><p>PBODY</p></h3>WTAIL</div></aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
		path []string
	}{
		{KindParagraph, "PBODY", []string{"PBODY"}},
		{KindParagraph, "WTAIL", []string{"PBODY"}},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (a wrapper's trailing run after a block-carrying heading must not be dropped)",
			len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true", i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
		if !reflect.DeepEqual(blocks[i].HeadingPath, w.path) {
			t.Errorf("block %q path = %v, want %v (the trailing run follows the heading, so it carries the heading's path)", w.text, blocks[i].HeadingPath, w.path)
		}
	}
}

// F1 aside-level control: at the aside itself emitTrailingAlways is already
// true, so the heading-with-block-child shape is unchanged by the fix — the
// code block emits, the heading text is structure, and there is no trailing
// run to keep.
func TestSplitBlocks_AsideLevelHeadingWithBlockChildControl(t *testing.T) {
	html := `<body><aside><h3><pre>CODE</pre></h3></aside></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindCode || blocks[0].Text != "CODE" || !blocks[0].Aside {
		t.Fatalf("want exactly one aside code block %q, got %d blocks: %v", "CODE", len(blocks), texts(blocks))
	}
	if !reflect.DeepEqual(blocks[0].HeadingPath, []string{"CODE"}) {
		t.Errorf("code block path = %v, want [CODE]", blocks[0].HeadingPath)
	}
}

// C2 (t1968 sync-audit iter6): the F7b fallback was all-or-nothing — it
// emitted the aside's collapsed text only when NO child had emitted a block,
// so a MIXED aside (`Important<p>Details</p>Tail`) kept only Details and the
// bare text runs on either side were lost from the document entirely. Bare
// runs are now extracted per run, in document order, as aside-scored content
// regardless of the child blocks beside them.
func TestSplitBlocks_MixedAsideBareTextRunsNotLost(t *testing.T) {
	html := `<body><aside>Important<p>Details</p>Tail</aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind  BlockKind
		text  string
		aside bool
	}{
		{KindParagraph, "Important", true},
		{KindParagraph, "Details", true},
		{KindParagraph, "Tail", true},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (bare runs of a mixed aside must not be lost)", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || blocks[i].Aside != w.aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=%v (bare runs in document order, aside-scored)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text, w.aside)
		}
	}
}

// C2 document-order control: the bare run BEFORE the block child precedes it
// in the split, and a run between two blocks lands between them.
func TestSplitBlocks_MixedAsideRunOrder(t *testing.T) {
	html := `<body><aside>head<ul><li>item</li></ul>tail</aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindParagraph, "head"},
		{KindListItem, "item"},
		{KindParagraph, "tail"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true (run, block, run in document order)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
	}
}

// C2 splice control: a bare run preceding a table inside the aside is
// spliced in front of the table's rows, and the row→header binding survives
// the splice — spliceAsideRun shifts rowTable keys and headerLocal positions
// recorded at or past the splice point, so resolveHeaders still binds the
// body row to its header after the shift.
func TestSplitBlocks_MixedAsideRunBeforeTableKeepsHeaderBinding(t *testing.T) {
	html := `<body><aside>intro<table><tr><th>H1</th></tr><tr><td>data</td></tr></table>tail</aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindParagraph, "intro"},
		{KindTableRow, "H1"},
		{KindTableRow, "data"},
		{KindParagraph, "tail"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true (run spliced before the table's rows)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
	}
	if blocks[2].HeaderIndex != 1 {
		t.Errorf("the body row's HeaderIndex = %d, want 1 (bound to the header row AFTER the splice)", blocks[2].HeaderIndex)
	}
}

// C3 (t1968 sync-audit iter6): an <aside> inside a <td> folded into the row
// text (the walk never descends into cells, so the aside case never ran and
// cellSkip did not exclude it) — the aside's content rode the row block and
// never reached the aside scoring stream. The row's text extraction now
// excludes aside subtrees, and the row scan emits them as separate
// aside-scored blocks, the same exclusion+separate-emission pattern F8
// applies to pre inside a cell.
func TestSplitBlocks_AsideInsideCellIsSeparateBlock(t *testing.T) {
	html := `<body><table><tr><td>Main<aside><p>Note</p></aside></td></tr></table></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind  BlockKind
		text  string
		aside bool
	}{
		{KindTableRow, "Main", false},
		{KindParagraph, "Note", true},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (cell text excludes the aside; the aside emits its own block)", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || blocks[i].Aside != w.aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=%v",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text, w.aside)
		}
	}
}

// R1 (t1968 sync-audit iter7): the C3 repair hands an in-cell aside to
// s.walk, and a nested <table> inside that aside advances the shared table
// counter mid-outer-table — every outer row emitted AFTER the aside-carrying
// row then bound its HeaderIndex to the INNER table's header row (auditor
// probe: the outer `cell2` row carried header "InnerH" instead of "OuterH";
// correct at the pre-C3 base). Table-id issuance is now decoupled from the
// raw counter — a row binds to the innermost OPEN table — so the nested table
// keeps its own blocks and header binding while the outer table's later rows
// keep theirs.
func TestSplitBlocks_NestedTableInCellAsideKeepsOuterHeaderBinding(t *testing.T) {
	html := `<body><table>
	<tr><th>OuterH</th></tr>
	<tr><td>cell1<aside><table><tr><th>InnerH</th></tr><tr><td>inner1</td></tr></table></aside></td></tr>
	<tr><td>cell2</td></tr>
	</table></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind        BlockKind
		text        string
		aside       bool
		headerIndex int
	}{
		{KindTableRow, "OuterH", false, -1},
		{KindTableRow, "cell1", false, 0},
		{KindTableRow, "InnerH", true, -1},
		{KindTableRow, "inner1", true, 2},
		{KindTableRow, "cell2", false, 0},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || blocks[i].Aside != w.aside || blocks[i].HeaderIndex != w.headerIndex {
			t.Errorf("block %d = %s %q aside=%v header=%d, want %s %q aside=%v header=%d",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, blocks[i].HeaderIndex, w.kind, w.text, w.aside, w.headerIndex)
		}
	}
	if header := blocks[4].HeaderIndex; blocks[header].Text != "OuterH" {
		t.Errorf("cell2's HeaderIndex %d points at %q, want the OuterH header row", header, blocks[header].Text)
	}
}

// R1 collateral (t1968 sync-audit iter8 repair): the table AFTER an outer
// table that contains an aside-nested table must keep its OWN header. A
// snapshot/restore of the counter around the aside handoff would restore the
// counter to the outer table's id, so this next table would re-use the nested
// table's id — its header row is then suppressed (headerLocal already
// carries the id) and its body rows bind the INNER header. The stack binding
// keeps the counter monotonic, so ids never collide.
func TestSplitBlocks_TableAfterAsideNestedTableKeepsOwnHeader(t *testing.T) {
	html := `<body><table>
	<tr><th>OuterH</th></tr>
	<tr><td>cell1<aside><table><tr><th>InnerH</th></tr><tr><td>inner1</td></tr></table></aside></td></tr>
	<tr><td>cell2</td></tr>
	</table>
	<table>
	<tr><th>SecondH</th></tr>
	<tr><td>second1</td></tr>
	</table></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		text        string
		headerIndex int
	}{
		{"OuterH", -1}, {"cell1", 0}, {"InnerH", -1}, {"inner1", 2},
		{"cell2", 0}, {"SecondH", -1}, {"second1", 5},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Text != w.text || blocks[i].HeaderIndex != w.headerIndex {
			t.Errorf("block %d = %q header=%d, want %q header=%d (each table's rows bind their own header)",
				i, blocks[i].Text, blocks[i].HeaderIndex, w.text, w.headerIndex)
		}
	}
}

// R1 collateral (t1968 sync-audit iter8 repair): TWO aside-nested tables in
// one outer table must hold distinct ids. A counter snapshot/restore would
// rewind to the outer id before each aside handoff, so the second nested
// table would re-use the first one's id — its header suppressed, its rows
// bound to the FIRST nested table's header (InA). With the stack binding both
// nested tables allocate fresh ids and every row binds its own header.
func TestSplitBlocks_TwoCellAsidesWithTablesDistinctBindings(t *testing.T) {
	html := `<body><table>
	<tr><th>OuterH</th></tr>
	<tr><td>a1<aside><table><tr><th>InA</th></tr><tr><td>ia</td></tr></table></aside></td></tr>
	<tr><td>a2<aside><table><tr><th>InB</th></tr><tr><td>ib</td></tr></table></aside></td></tr>
	<tr><td>cell2</td></tr>
	</table></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		text        string
		aside       bool
		headerIndex int
	}{
		{"OuterH", false, -1},
		{"a1", false, 0},
		{"InA", true, -1},
		{"ia", true, 2},
		{"a2", false, 0},
		{"InB", true, -1},
		{"ib", true, 5},
		{"cell2", false, 0},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Text != w.text || blocks[i].Aside != w.aside || blocks[i].HeaderIndex != w.headerIndex {
			t.Errorf("block %d = %q aside=%v header=%d, want %q aside=%v header=%d",
				i, blocks[i].Text, blocks[i].Aside, blocks[i].HeaderIndex, w.text, w.aside, w.headerIndex)
		}
	}
}

// P2 (card t1968 codex review, repair round): a <table> nested DIRECTLY in a
// <td> is neither emitted nor excluded. emitRowSubBlocks.scan had branches
// for pre and aside but none for table, so the nested table's rows never
// became table blocks and its flattened text rode inside the parent cell's
// text. The nested table must emit through the same table machinery the
// body-level walk uses (its own table id via tableStack, header rows binding
// to ITS OWN header), and the parent cell's text must exclude the nested
// subtree (cellSkip) so the parent row reads only its own content.
func TestSplitBlocks_NestedTableInCellEmitsOwnBlocks(t *testing.T) {
	html := `<body><table>
	<tr><th>OuterH</th></tr>
	<tr><td>lead <table><tr><th>InnerH</th></tr><tr><td>inner1</td></tr></table> tail</td></tr>
	<tr><td>cell2</td></tr>
	</table></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		text        string
		headerIndex int
	}{
		{"OuterH", -1},
		{"lead tail", 0},
		{"InnerH", -1},
		{"inner1", 2},
		{"cell2", 0},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Text != w.text || blocks[i].HeaderIndex != w.headerIndex {
			t.Errorf("block %d = %q header=%d, want %q header=%d (inner rows bind the inner header; the parent cell excludes the inner table's text)",
				i, blocks[i].Text, blocks[i].HeaderIndex, w.text, w.headerIndex)
		}
	}
}

// F8 (t1968 sync-audit iter2): a <pre> inside a <td> keeps its formatting —
// the code block is emitted verbatim as its own block (the C2 pattern for
// pre inside a list item) and the row's text EXCLUDES the pre subtree,
// keeping only the cell's own text runs across the boundary separator.
// Before the repair the row text flattened the code ("first second") and no
// code block was emitted — the walk never descends into cells, so the pre
// was otherwise unreachable.
func TestSplitBlocks_PreInsideCellIsVerbatimCodeBlock(t *testing.T) {
	html := "<body><table><tr><td>Usage:<pre>  first\n    second</pre></td><td>after</td></tr></table></body>"
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindTableRow, "Usage: | after"},
		{KindCode, "  first\n    second"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text {
			t.Errorf("block %d = %s %q, want %s %q (code verbatim, row text excludes it)",
				i, blocks[i].Kind, blocks[i].Text, w.kind, w.text)
		}
	}
}

// F8b (t1968 sync-audit iter3): the F8 row-level code scan descended into
// chrome subtrees — unlike the body-level walk, rowCells, and collapseText —
// so a pre inside nav/template/footer inside a cell re-emitted the chrome's
// content as a code block (auditor probe: `<td>Visible<nav><pre>CHROME_ONLY
// </pre></nav></td>` emitted code "CHROME_ONLY"; the body-level control emits
// nothing). The scan now skips chromeTags subtrees like every other path:
// only non-chrome pre content is emitted.
func TestSplitBlocks_PreInsideChromeInCellIsNotEmitted(t *testing.T) {
	html := `<body><table><tr><td>Visible<nav><pre>CHROME_ONLY</pre></nav></td></tr></table></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindTableRow || blocks[0].Text != "Visible" {
		t.Fatalf("want exactly one table-row block %q, got %d blocks: %+v", "Visible", len(blocks), blocks)
	}
}

// Token accounting rides jev.EstimateTokens (REQ-FW-002): a block's token
// count is the conservative byte/4 estimate, and document totals sum it.
func TestBlockTokens_UseEstimator(t *testing.T) {
	b := Block{Kind: KindParagraph, Text: "0123456789"} // 10 bytes -> 3 tokens (ceil 10/4)
	if got := b.Tokens(); got != jev.EstimateTokens("0123456789") {
		t.Errorf("Tokens() = %d, want jev.EstimateTokens value %d", got, jev.EstimateTokens("0123456789"))
	}
	if jev.EstimateTokens("0123456789") != 3 {
		t.Errorf("estimator premise changed: 10 bytes -> %d tokens, want 3", jev.EstimateTokens("0123456789"))
	}
}

// DocTokens is the whole-document token count the passthrough threshold
// compares against (approximate — see package note).
func TestDocTokens_SumsBlocks(t *testing.T) {
	blocks := []Block{
		{Kind: KindParagraph, Text: "aaaaaaaa"}, // 8 bytes -> 2
		{Kind: KindParagraph, Text: "bbbb"},     // 4 bytes -> 1
	}
	if got := DocTokens(blocks); got != 3 {
		t.Errorf("DocTokens = %d, want 3", got)
	}
}

// Leader codex-audit round 3, finding 1 (C2 class): the per-run extraction
// ran only over an aside's DIRECT children, so a wrapper child that itself
// contains block-emitting descendants swallowed the bare runs inside it —
// `<aside><div>Important<p>Details</p>Tail</div></aside>` emitted only
// Details; Important and Tail were lost from the document entirely. The
// extraction is recursive: a wrapper's interior runs are spliced around the
// blocks its descendants emit, in document order, aside-scored.
func TestSplitBlocks_MixedAsideBareRunsInsideWrapperNotLost(t *testing.T) {
	html := `<body><aside><div>Important<p>Details</p>Tail</div></aside></body>`
	blocks := SplitBlocks(html)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindParagraph, "Important"},
		{KindParagraph, "Details"},
		{KindParagraph, "Tail"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (bare runs nested inside a wrapper child must not be lost)", len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true (wrapper interior runs in document order, aside-scored)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
	}
}

// F1 regression (t1968 sync-audit iter10, probe P7): the round-3 recursive
// run extraction must not reach inside a list-item body. The li block already
// carries the wrapper's inline text (listItemSkip folds it in while excluding
// the separately-emitted pre), so re-splicing the wrapper's bare runs around
// the pre duplicates the item's text — `<aside><ul><li><div>LEAD<pre>CODE
// </pre>TAIL</div></li></ul></aside>` emitted [LEAD TAIL, LEAD, CODE, TAIL].
// Inside an item body only separate-block descendants (pre/table, nested li,
// nested aside via its own case) re-enter emission.
func TestSplitBlocks_WrapperInsideListItemNotDuplicated(t *testing.T) {
	page := `<body><aside><ul><li><div>LEAD<pre>CODE</pre>TAIL</div></li></ul></aside></body>`
	blocks := SplitBlocks(page)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindListItem, "LEAD TAIL"},
		{KindCode, "CODE"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (a wrapper inside a list-item body must not re-extract the runs the li block already carries)",
			len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true (li text carries the runs; the pre is its own code block)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
	}
}

// F1 regression, nested-list variant (t1968 sync-audit iter10, probe P9): the
// inner li carries its wrapper's inline text ("A3 C3"); the wrapper's pre is
// the only descendant that re-enters emission. The duplication showed as
// [TOP, A3 C3, A3, B3, C3].
func TestSplitBlocks_NestedListInnerWrapperNotDuplicated(t *testing.T) {
	page := `<body><aside><ul><li>TOP<ul><li><div>A3<pre>B3</pre>C3</div></li></ul></li></ul></aside></body>`
	blocks := SplitBlocks(page)
	want := []struct {
		kind BlockKind
		text string
	}{
		{KindListItem, "TOP"},
		{KindListItem, "A3 C3"},
		{KindCode, "B3"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("want %d blocks, got %d: %v (an inner-list wrapper's runs belong to the inner li's text, not separate paragraphs)",
			len(want), len(blocks), texts(blocks))
	}
	for i, w := range want {
		if blocks[i].Kind != w.kind || blocks[i].Text != w.text || !blocks[i].Aside {
			t.Errorf("block %d = %s %q aside=%v, want %s %q aside=true (nested items in order; the pre is its own code block)",
				i, blocks[i].Kind, blocks[i].Text, blocks[i].Aside, w.kind, w.text)
		}
	}
}

// Class-close sweep (leader codex-audit round 3): a corpus over the wrapper
// space {aside, section, article, li, td, div, blockquote, figure} × child
// shapes {bare-only, block-only, mixed, multi-mixed, nested-wrapper-blocks,
// heading-with-block-child}, all inside an aside. The invariant: NO bare text run anywhere in the
// subtree is dropped when any block is emitted from the same subtree — every
// planted word appears, in document order, among the emitted blocks' texts.
// (`td` outside a table has its tag dropped by the HTML5 parser — its
// children are hoisted to the aside — so that row exercises the direct
// C2 path; every other wrapper keeps its element and exercises the
// recursive extraction.)
func TestSplitBlocks_AsideWrapperCorpusNoBareRunLost(t *testing.T) {
	wrappers := []string{"aside", "section", "article", "li", "td", "div", "blockquote", "figure"}
	shapes := []struct {
		name  string
		inner func(tag string) string
		words func(tag string) []string
	}{
		{"bare-only",
			func(tag string) string { return "BARE_" + tag },
			func(tag string) []string { return []string{"BARE_" + tag} }},
		{"block-only",
			func(tag string) string { return "<p>P1_" + tag + "</p><p>P2_" + tag + "</p>" },
			func(tag string) []string { return []string{"P1_" + tag, "P2_" + tag} }},
		{"mixed",
			func(tag string) string { return "B1_" + tag + "<p>P1_" + tag + "</p>B2_" + tag },
			func(tag string) []string { return []string{"B1_" + tag, "P1_" + tag, "B2_" + tag} }},
		{"multi-mixed",
			func(tag string) string {
				return "B1_" + tag + "<p>P1_" + tag + "</p>B2_" + tag + "<p>P2_" + tag + "</p>B3_" + tag
			},
			func(tag string) []string {
				return []string{"B1_" + tag, "P1_" + tag, "B2_" + tag, "P2_" + tag, "B3_" + tag}
			}},
		{"nested-wrapper-blocks",
			func(tag string) string {
				return "<div>N1_" + tag + "<p>NP1_" + tag + "</p>N2_" + tag + "</div>B1_" + tag
			},
			func(tag string) []string {
				return []string{"N1_" + tag, "NP1_" + tag, "N2_" + tag, "B1_" + tag}
			}},
		// F1 axis (t1968 sync-audit iter11): a heading whose own subtree
		// carries a block-emitting child, first in the wrapper with a
		// trailing run after it. The auditor named this axis's absence as
		// the blind spot that let the trailing-run drop through: every other
		// shape reaches `emitted` through the default branch, which counts
		// its child's emissions — only the heading branch could emit from a
		// descent without marking the container mixed. The p variant keeps
		// all eight wrappers on one uniform assertion: a p inside an item
		// body is suppressed and folds into the li text in document order
		// (the item corpus's documented fold), where a pre would emit after
		// the li text and invert emission order by design — the pre-in-
		// heading exact shape is pinned by
		// TestSplitBlocks_AsideWrapperHeadingWithBlockChildKeepsTrailingRun.
		// (`aside` and the hoisted `td` keep the trailing run through
		// emitTrailingAlways, and `li` folds both words into its item text —
		// controls; the remaining five wrappers need the fix.)
		{"heading-with-block-child",
			func(tag string) string { return "<h3><p>P_" + tag + "</p></h3>T_" + tag },
			func(tag string) []string { return []string{"P_" + tag, "T_" + tag} }},
	}
	for _, w := range wrappers {
		for _, sh := range shapes {
			tag := w + "-" + sh.name
			page := "<body><h2>C</h2><aside><" + w + ">" + sh.inner(tag) + "</" + w + "></aside></body>"
			t.Run(w+"/"+sh.name, func(t *testing.T) {
				blocks := SplitBlocks(page)
				var got []string
				for _, b := range blocks {
					if !b.Aside {
						t.Errorf("block %q is not aside-scored (every block from inside the aside carries Aside=true)", b.Text)
					}
					got = append(got, b.Text)
				}
				if len(got) == 0 {
					t.Fatalf("extractor emitted no block: planted content lost wholesale from %s", page)
				}
				joined := strings.Fields(strings.Join(got, " "))
				want := sh.words(tag)
				matched := 0
				for _, word := range joined {
					if matched < len(want) && word == want[matched] {
						matched++
					}
				}
				if matched != len(want) {
					t.Errorf("%s/%s: bare text runs dropped: planted %v, matched %d/%d in document order among %v",
						w, sh.name, want, matched, len(want), got)
				}
			})
		}
	}
}

// Class-close sweep, itemDepth axis (t1968 sync-audit iter10 F1): the round-3
// recursive extraction must not reach inside list-item bodies — the li block
// already carries a wrapper's inline text (collapseText folds it through
// listItemSkip's block exclusions), so re-extracting bare runs around a
// block-emitting descendant duplicates the item's text. The corpus: the same
// wrapper set as the round-3 sweep, each planted INSIDE a list-item body,
// over child shapes {wrapper-with-blocks, wrapper-with-mixed, nested}. The
// invariant inverts the round-3 one: no-loss AND no-duplication — every
// planted word (item text and shape content alike) appears EXACTLY ONCE
// among the emitted blocks' texts. Emission order is not asserted here: the
// li text block lands before its descendants' separate blocks by design (the
// li case emits then descends), so order is pinned by the P7/P9 exact-output
// regressions above instead.
// (`td` is omitted: the parser drops the tag outside a table and its children
// fuse into the item text as directly adjacent text nodes — no element
// boundary, so no separator fires and the planted words merge into single
// tokens, unassertable at word granularity; the td-in-cell path is covered by
// TestSplitBlocks_AsideCellCorpusNoBareRunLost below. `li` inside a
// div-inside-li is restructured by the parser into a sibling item, which
// still exercises the fixed path whenever the shape nests a wrapper inside
// the restructured item's body. `aside` hits its own case, which scopes
// itemDepth per entry — the F7/P8 shape, a control.)
func TestSplitBlocks_AsideWrapperInListItemCorpusNoDuplication(t *testing.T) {
	wrappers := []string{"aside", "section", "article", "li", "div", "blockquote", "figure"}
	shapes := []struct {
		name  string
		inner func(tag string) string
		words func(tag string) []string
	}{
		{"wrapper-with-blocks",
			func(tag string) string { return "<pre>P1_" + tag + "</pre><pre>P2_" + tag + "</pre>" },
			func(tag string) []string {
				return []string{"HEAD_" + tag, "P1_" + tag, "P2_" + tag, "TAIL_" + tag}
			}},
		{"wrapper-with-mixed",
			func(tag string) string { return "B1_" + tag + "<pre>P1_" + tag + "</pre>B2_" + tag },
			func(tag string) []string {
				return []string{"HEAD_" + tag, "B1_" + tag, "P1_" + tag, "B2_" + tag, "TAIL_" + tag}
			}},
		{"nested",
			func(tag string) string {
				return "B1_" + tag + "<div>N1_" + tag + "<pre>P1_" + tag + "</pre>N2_" + tag + "</div>B2_" + tag
			},
			func(tag string) []string {
				return []string{"HEAD_" + tag, "B1_" + tag, "N1_" + tag, "P1_" + tag, "N2_" + tag, "B2_" + tag, "TAIL_" + tag}
			}},
	}
	for _, w := range wrappers {
		for _, sh := range shapes {
			tag := w + "-" + sh.name
			page := "<body><aside><ul><li>HEAD_" + tag + "<" + w + ">" + sh.inner(tag) + "</" + w + ">TAIL_" + tag + "</li></ul></aside></body>"
			t.Run(w+"/"+sh.name, func(t *testing.T) {
				blocks := SplitBlocks(page)
				counts := map[string]int{}
				for _, b := range blocks {
					if !b.Aside {
						t.Errorf("block %q is not aside-scored (every block from inside the aside carries Aside=true)", b.Text)
					}
					for _, word := range strings.Fields(b.Text) {
						counts[word]++
					}
				}
				if len(blocks) == 0 {
					t.Fatalf("extractor emitted no block: planted content lost wholesale from %s", page)
				}
				for _, word := range sh.words(tag) {
					switch counts[word] {
					case 0:
						t.Errorf("%s/%s: planted word %q lost (0 occurrences among %v)", w, sh.name, word, texts(blocks))
					case 1:
					default:
						t.Errorf("%s/%s: planted word %q duplicated (%d occurrences among %v — the li block already carries it)",
							w, sh.name, word, counts[word], texts(blocks))
					}
				}
			})
		}
	}
}

// A real td needs a table. Cells are never walked as containers — the cell's
// text (bare runs and folded p text alike) rides the row block — so the
// no-loss invariant holds through the row text, with the pre emitted after
// the row as its own verbatim code block.
func TestSplitBlocks_AsideCellCorpusNoBareRunLost(t *testing.T) {
	html := `<body><aside><table><tr><td>MIXTD1<p>PTD1</p>MIXTD2<pre>  code</pre></td></tr></table></aside></body>`
	blocks := SplitBlocks(html)
	var got []string
	for _, b := range blocks {
		if !b.Aside {
			t.Errorf("block %q is not aside-scored", b.Text)
		}
		got = append(got, b.Text)
	}
	want := []string{"MIXTD1", "PTD1", "MIXTD2", "code"}
	joined := strings.Fields(strings.Join(got, " "))
	matched := 0
	for _, word := range joined {
		if matched < len(want) && word == want[matched] {
			matched++
		}
	}
	if matched != len(want) {
		t.Errorf("cell bare runs dropped: planted %v, matched %d/%d among %v", want, matched, len(want), got)
	}
}

// D3 (t1968 §5 debt ledger — untested territory, debt-closure round):
// characterization corpus for CDATA sections and frameset documents. These
// are OBSERVED-behavior pins (probe-verified against x/net/html's HTML5
// tree construction), not new behavior: the policies below document what
// the splitter does with the shapes so a parser upgrade that changes them
// surfaces here. Proxy-env interactions remain out of scope by design.
func TestSplitBlocks_UntestedTerritoryCorpusCDATAAndFrameset(t *testing.T) {
	t.Run("cdata-in-html-content-is-bogus-comment", func(t *testing.T) {
		// HTML5 tree construction treats <![CDATA[...]]> outside foreign
		// content as a BOGUS COMMENT ending at the first '>' — the section's
		// interior is not character data, and markup-looking text after that
		// '>' remains as ordinary text. Policy: the comment node is dropped
		// (structure, not content); the tail text stays, joined across the
		// hole.
		blocks := SplitBlocks(`<body><p>before <![CDATA[raw <b>bold</b> cdata]]> after</p></body>`)
		if len(blocks) != 1 || blocks[0].Kind != KindParagraph || blocks[0].Text != "before bold cdata]]> after" {
			t.Fatalf("want one paragraph %q, got %v", "before bold cdata]]> after", texts(blocks))
		}
	})
	t.Run("cdata-only-cell-yields-empty-row", func(t *testing.T) {
		// A cell whose entire content is one CDATA bogus comment (no '>' in
		// its interior) collapses to empty cell text: the row block still
		// emits, with empty text. Policy: comment-only content is dropped,
		// not substituted.
		blocks := SplitBlocks(`<body><table><tr><th>H</th></tr><tr><td><![CDATA[a | b]]></td></tr></table></body>`)
		if len(blocks) != 2 || blocks[0].Kind != KindTableRow || blocks[0].Text != "H" ||
			blocks[1].Kind != KindTableRow || blocks[1].Text != "" {
			t.Fatalf("want rows [H ] with the CDATA-only cell empty, got %v", texts(blocks))
		}
	})
	t.Run("cdata-in-pre-dropped-without-separator", func(t *testing.T) {
		// preText drops the comment node and keeps the surrounding code text
		// with its original spacing (verbatim preservation; the hole is not
		// given a separator, so the two adjacent spaces survive).
		blocks := SplitBlocks(`<body><pre>code <![CDATA[x < y]]> tail</pre></body>`)
		if len(blocks) != 1 || blocks[0].Kind != KindCode || blocks[0].Text != "code  tail" {
			t.Fatalf("want code %q, got %v", "code  tail", texts(blocks))
		}
	})
	t.Run("cdata-in-foreign-content-is-text", func(t *testing.T) {
		// In foreign content (math — svg is chrome and dropped whole) a
		// CDATA section IS character data per HTML5: the splitter emits it
		// as part of the enclosing paragraph's text, markup characters
		// intact. This is the "content emitted as text where appropriate"
		// arm of the policy.
		blocks := SplitBlocks(`<body><p>a <math><![CDATA[x<y]]></math> b</p></body>`)
		if len(blocks) != 1 || blocks[0].Kind != KindParagraph || blocks[0].Text != "a x<y b" {
			t.Fatalf("want paragraph %q, got %v", "a x<y b", texts(blocks))
		}
	})
	t.Run("cdata-in-aside-dropped-run-keeps-boundary", func(t *testing.T) {
		// The aside fallback's run extraction drops the bogus comment; the
		// surrounding bare text keeps its word boundary.
		blocks := SplitBlocks(`<body><aside>note <![CDATA[raw cdata]]> tail</aside></body>`)
		if len(blocks) != 1 || blocks[0].Kind != KindParagraph || !blocks[0].Aside || blocks[0].Text != "note tail" {
			t.Fatalf("want one aside paragraph %q, got %v", "note tail", texts(blocks))
		}
	})
	t.Run("frameset-document-no-crash-no-blocks", func(t *testing.T) {
		// A frameset document has no body: the walk sees frameset/frame/
		// noframes, none of which are block containers, and noframes is a
		// RAW TEXT element — its interior stays one unparsed text node.
		// Policy: nothing is emitted (no crash, no phantom blocks); the
		// frames' TARGET documents are separate fetches, out of a
		// single-page split's scope by design.
		blocks := SplitBlocks(`<html><head><title>FS</title></head><frameset cols="50%,50%"><frame src="a.html"><frame src="b.html"><noframes><p>fallback content</p></noframes></frameset></html>`)
		if len(blocks) != 0 {
			t.Fatalf("want 0 blocks from a frameset document, got %v", texts(blocks))
		}
	})
	t.Run("frameset-noframes-plain-text-also-silent", func(t *testing.T) {
		// Plain text (no markup) inside noframes is still raw text under a
		// non-content container: the same policy — silent, no crash.
		blocks := SplitBlocks(`<html><frameset rows="*"><frame src="x.html"><noframes>plain fallback text</noframes></frameset></html>`)
		if len(blocks) != 0 {
			t.Fatalf("want 0 blocks, got %v", texts(blocks))
		}
	})
}

// Recursion guards: the recursive extraction must not SPLIT runs that were
// one paragraph. A wrapper with no block-emitting descendant is pure inline
// content and merges into the enclosing run (F7b semantics — the audit
// probe `<aside>...<div>deep body</div>...` stays one paragraph), and an
// inline element (span) never starts a run of its own.
func TestSplitBlocks_AsideBareOnlyWrapperMergesIntoRun(t *testing.T) {
	html := `<body><aside>head<div>deep body</div>tail</aside></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindParagraph || blocks[0].Text != "head deep body tail" || !blocks[0].Aside {
		t.Fatalf("want exactly one aside paragraph %q, got %d blocks: %v (a wrapper with no interior block merges into the enclosing run)",
			"head deep body tail", len(blocks), texts(blocks))
	}
}

func TestSplitBlocks_AsideInlineWrapperRunNotSplit(t *testing.T) {
	html := `<body><aside>Hello <span>world</span> again</aside></body>`
	blocks := SplitBlocks(html)
	if len(blocks) != 1 || blocks[0].Kind != KindParagraph || blocks[0].Text != "Hello world again" || !blocks[0].Aside {
		t.Fatalf("want exactly one aside paragraph %q, got %d blocks: %v (an inline wrapper never splits the run around it)",
			"Hello world again", len(blocks), texts(blocks))
	}
}
