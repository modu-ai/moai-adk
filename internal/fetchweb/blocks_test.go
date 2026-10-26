package fetchweb

import (
	"reflect"
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
