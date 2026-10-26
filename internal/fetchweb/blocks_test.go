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
