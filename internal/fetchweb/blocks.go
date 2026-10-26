// Package fetchweb is the Go port of the jevfetch pipeline (SPEC-FETCH-WEB-001):
// split a fetched HTML page into blocks, score them for relevance against a
// question in two Jev stages, and return the kept blocks verbatim.
//
// Token counts throughout this package are APPROXIMATE. They come from
// jev.EstimateTokens (a conservative bytes-per-token estimate), not an exact
// tokenizer, so every window split, threshold comparison, and savings figure
// built on them is approximate by construction (REQ-FW-002). The upstream
// reference counts with an exact tokenizer; the Go estimator is deliberately
// conservative and the two may diverge. No tokenizer dependency is added.
//
// The package never dials anything by itself: the page fetcher takes an
// injected transport and the scorer takes an Answerer seam, so no test can
// reach a real endpoint by accident (REQ-FW-010).
package fetchweb

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// BlockKind names the four block kinds the splitter produces (REQ-FW-001).
type BlockKind string

const (
	KindParagraph BlockKind = "paragraph"
	KindListItem  BlockKind = "list-item"
	KindTableRow  BlockKind = "table-row"
	KindCode      BlockKind = "code"
)

// Block is one content unit of a split page.
type Block struct {
	Kind BlockKind
	// HeadingPath is the sequence of ancestor heading texts, outermost first.
	HeadingPath []string
	// Text is the block's extracted text. Words are verbatim; runs of
	// whitespace collapse to single spaces except inside code blocks, where
	// line structure is preserved.
	Text string
	// Aside marks a block emitted from an <aside> note: it is scored
	// separately from the main flow (REQ-FW-001).
	Aside bool
	// HeaderIndex is the splitter-level index of the header-row block of the
	// table this row belongs to, or -1 when the table carries no header row.
	// The selection stage uses it to keep a table row's header adjacent to it
	// (REQ-FW-006).
	HeaderIndex int
}

// Tokens returns the block's approximate token count (jev.EstimateTokens;
// see the package note on approximation).
func (b Block) Tokens() int { return jev.EstimateTokens(b.Text) }

// DocTokens sums the blocks' approximate token counts. This is the total the
// passthrough threshold (≤ 2,000) compares against — approximate, per the
// package note.
func DocTokens(blocks []Block) int {
	total := 0
	for _, b := range blocks {
		total += b.Tokens()
	}
	return total
}

// chromeTags are the non-content furniture elements whose entire subtrees are
// dropped before splitting (REQ-FW-001: navigation, footers, scripts, styles,
// templates, and similar). `header` is the banner element, not headings —
// heading paths come from h1..h6 elements, which are content structure.
var chromeTags = map[string]bool{
	"script":   true,
	"style":    true,
	"template": true,
	"noscript": true,
	"nav":      true,
	"footer":   true,
	"header":   true,
	"form":     true,
	"iframe":   true,
	"svg":      true,
}

// headingLevels maps heading tags to their nesting level.
var headingLevels = map[string]int{
	"h1": 1, "h2": 2, "h3": 3, "h4": 4, "h5": 5, "h6": 6,
}

// SplitBlocks splits an HTML page into blocks (REQ-FW-001): paragraphs, list
// items, table rows, and code blocks, each carrying its heading path; page
// chrome removed; aside content emitted as separately-flagged blocks; blocks
// in document order.
func SplitBlocks(pageHTML string) []Block {
	if strings.TrimSpace(pageHTML) == "" {
		return nil
	}
	root, err := html.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return nil
	}
	s := &splitter{headerLocal: map[int]int{}, rowTable: map[int]int{}}
	var out []Block
	s.blocks = &out
	// html.Parse returns a DocumentNode (not an ElementNode), so the descent
	// starts from its children explicitly.
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		s.walk(c, false)
	}
	s.resolveHeaders()
	return out
}

// splitter carries per-split state. Each table row remembers the table it
// belongs to; header positions are resolved once, after the walk, so a header
// appearing after a body row in the markup still attaches.
type splitter struct {
	blocks *[]Block
	// path is the ACTIVE heading stack. It lives on the splitter rather than
	// a walk parameter because a heading's effect must persist past the
	// heading element itself: the blocks that FOLLOW the heading (its later
	// siblings, and their descendants) carry it in their path.
	path []heading
	// tableSeq numbers tables in document order.
	tableSeq int
	// rowTable[tableRowBlockIndex] = the table a row belongs to.
	rowTable map[int]int
	// headerLocal[table id] = the splitter-level index of that table's
	// header-row block; absent when the table has none.
	headerLocal map[int]int
	// itemDepth counts the walk's position inside list-item bodies: while
	// > 0, a <p> element is part of the enclosing item's already-emitted
	// text and is not emitted again as its own block.
	itemDepth int
}

// walk visits one node. Sibling advancement belongs to the caller's loop
// (walkChildren) — a node that handled its own NextSibling would be visited
// twice, once as a child and once as a sibling.
func (s *splitter) walk(n *html.Node, inAside bool) {
	if n.Type != html.ElementNode {
		return
	}
	switch {
	case chromeTags[n.Data]:
		// dropped subtree
	case n.Data == "aside":
		s.walkChildren(n, true)
	case n.Data == "p":
		if s.itemDepth > 0 {
			// A paragraph inside a list item body is already part of that
			// item's emitted text; re-emitting it duplicates the block.
			break
		}
		if text := collapseText(n, nil); text != "" {
			s.emit(Block{Kind: KindParagraph, HeadingPath: clonePath(s.path), Text: text, Aside: inAside, HeaderIndex: -1}, 0)
		}
	case n.Data == "li":
		if text := collapseText(n, listNesting); text != "" {
			s.emit(Block{Kind: KindListItem, HeadingPath: clonePath(s.path), Text: text, Aside: inAside, HeaderIndex: -1}, 0)
		}
		// Descend anyway: a nested list's own items are blocks in their own
		// right, excluded from this item's text but not from the document.
		// The descent is marked: paragraphs inside an item body are already
		// carried by that item's text; only the nested items re-enter
		// emission.
		s.itemDepth++
		s.walkChildren(n, inAside)
		s.itemDepth--
	case n.Data == "pre":
		if text := preText(n); text != "" {
			s.emit(Block{Kind: KindCode, HeadingPath: clonePath(s.path), Text: text, Aside: inAside, HeaderIndex: -1}, 0)
		}
	case n.Data == "table":
		s.tableSeq++
		s.walkChildren(n, inAside)
	case n.Data == "tr":
		s.emitRow(n, inAside)
	case headingLevels[n.Data] > 0:
		// Headings are structure, not blocks: their text joins the path for
		// the blocks that follow. A same-or-higher level heading pops deeper
		// entries first. The push deliberately persists past this element —
		// see the path field note.
		lvl := headingLevels[n.Data]
		for len(s.path) > 0 && s.path[len(s.path)-1].level >= lvl {
			s.path = s.path[:len(s.path)-1]
		}
		s.path = append(s.path, heading{level: lvl, text: collapseText(n, nil)})
		s.walkChildren(n, inAside)
	default:
		s.walkChildren(n, inAside)
	}
}

// walkChildren descends into an element's children only.
func (s *splitter) walkChildren(n *html.Node, inAside bool) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.walk(c, inAside)
	}
}

// emit appends a block. table is 0 for non-table-row blocks.
func (s *splitter) emit(b Block, table int) {
	idx := len(*s.blocks)
	if b.Kind == KindTableRow {
		s.rowTable[idx] = table
	}
	*s.blocks = append(*s.blocks, b)
}

// emitRow appends a table-row block. The first row of a table whose cells
// are th elements is that table's header row.
func (s *splitter) emitRow(n *html.Node, inAside bool) {
	cells := rowCells(n)
	if len(cells) == 0 {
		return
	}
	table := s.tableSeq
	if _, seen := s.headerLocal[table]; !seen && rowIsHeader(n) {
		s.headerLocal[table] = len(*s.blocks)
	}
	s.emit(Block{
		Kind:        KindTableRow,
		HeadingPath: clonePath(s.path),
		Text:        strings.Join(cells, " | "),
		Aside:       inAside,
		HeaderIndex: -1,
	}, table)
}

// resolveHeaders fills each table row's HeaderIndex from the recorded
// per-table header position.
func (s *splitter) resolveHeaders() {
	blocks := *s.blocks
	for i := range blocks {
		if blocks[i].Kind != KindTableRow {
			continue
		}
		table := s.rowTable[i]
		if h, ok := s.headerLocal[table]; ok && h != i {
			blocks[i].HeaderIndex = h
		}
	}
}

// heading is one entry of the active heading path.
type heading struct {
	level int
	text  string
}

func clonePath(path []heading) []string {
	out := make([]string, len(path))
	for i, h := range path {
		out[i] = h.text
	}
	return out
}

// collapseText extracts an element's text with whitespace runs collapsed to
// single spaces. skip, when non-nil, names tags whose subtrees are excluded
// (a list item's nested list must not leak into its own text). A text
// boundary element (<br>, a block container) emits a separator on each side
// so adjacent text runs across the boundary keep their word boundary; the
// final collapse reduces the separators to single spaces, never to nothing.
func collapseText(n *html.Node, skip map[string]bool) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && skip != nil && skip[node.Data] {
			return
		}
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
			return
		}
		boundary := node.Type == html.ElementNode && textBoundaryTags[node.Data]
		if boundary {
			b.WriteByte(' ')
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
		if boundary {
			b.WriteByte(' ')
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visit(c)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// textBoundaryTags are the elements whose boundaries end a run of text: the
// <br> break and the block-level containers that can nest inside an extracted
// element (a paragraph inside a list item, two paragraphs inside a cell).
// Without a separator there, directly adjacent text nodes merge into one
// word ("firstsecond"), losing the original word boundary (REQ-FW-006).
var textBoundaryTags = map[string]bool{
	"br": true, "p": true, "div": true, "li": true, "ul": true, "ol": true,
	"table": true, "thead": true, "tbody": true, "tr": true, "td": true,
	"th": true, "blockquote": true, "pre": true, "dl": true, "dt": true,
	"dd": true, "figure": true, "figcaption": true, "section": true,
	"article": true, "h1": true, "h2": true, "h3": true, "h4": true,
	"h5": true, "h6": true,
}

// listNesting excludes nested lists from a list item's own text.
var listNesting = map[string]bool{"ul": true, "ol": true}

// preText extracts a code block's text preserving line structure. Only the
// surrounding blank lines are trimmed; interior whitespace — including the
// first real code line's leading indentation — is kept verbatim.
func preText(n *html.Node) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visit(c)
	}
	return trimBlankLines(b.String())
}

// trimBlankLines removes blank lines from the start and end of s, preserving
// every interior line and the leading whitespace of the first real code line.
// A character-class Trim here would also strip that indentation, changing the
// code's meaning (REQ-FW-006 verbatim code preservation).
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// rowCells lists a table row's cell texts in order.
func rowCells(tr *html.Node) []string {
	var cells []string
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			cells = append(cells, collapseText(c, nil))
		}
	}
	return cells
}

// rowIsHeader reports whether a row is its table's header row: at least one
// th cell and no td cells (a thead row is by construction all-th).
func rowIsHeader(tr *html.Node) bool {
	hasTH, hasTD := false, false
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		switch c.Data {
		case "th":
			hasTH = true
		case "td":
			hasTD = true
		}
	}
	return hasTH && !hasTD
}
