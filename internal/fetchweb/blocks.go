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
// reach a real endpoint by accident (REQ-FW-010). The one client the package
// does construct — NewPageClient, for the production call sites — is the
// SSRF-guarded client (ssrf.go): its dialer refuses non-public addresses and
// its redirect policy refuses non-HTTP(S) hops.
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
	s := &splitter{}
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
	// tableSeq allocates table ids in walk order. It only ever grows — it is
	// a pure allocator, never a "current table" indicator; row binding reads
	// tableStack, not this counter.
	tableSeq int
	// tableStack is the chain of tables whose walk is currently open,
	// outermost first. A row binds to the innermost open table (the stack
	// top), so a nested table inside an in-cell aside — the C3 handoff in
	// emitRowSubBlocks — allocates its own id and opens/closes around its own
	// rows without disturbing the enclosing table's binding. Reading the raw
	// counter at emission instead bound every outer row AFTER the
	// aside-carrying row to the nested table's header (R1, t1968 sync-audit
	// iter7: `cell2` carried header "InnerH" where "OuterH" was correct). A
	// snapshot/restore of the counter around the handoff would fix that row
	// but re-use the aside's ids — the next table's header would be
	// suppressed and a second aside's nested table would collide with the
	// first — so the binding is positional and the counter stays monotonic.
	tableStack []int
	// rowOf[i] = the table id block i belongs to when block i is a
	// table-row block; -1 for every other block. Keyed by CURRENT block
	// position and shifted by the same move spliceAsideRun applies to the
	// blocks slice itself, so index i always describes blocks[i]. The map
	// this slice replaced was keyed by block index too but REBUILT whole on
	// every aside splice — re-keying every row of every earlier table,
	// O(rows×splices) with an allocation per splice (P2 perf, card t1968
	// repair round); the slice only ever pays for the tail it shifts.
	rowOf []int
	// headerAt[i] = the table id whose header row sits at block position i;
	// -1 when block i is not a header row. Same position-keyed shift as
	// rowOf, so resolveHeaders reads current positions with no per-splice
	// bookkeeping of its own.
	headerAt []int
	// tableHeadered[table id] = that table's first row was a header row.
	// Keyed by the stable table id — ids are allocated once by tableSeq and
	// never shift — so emitRow's first-header check needs no position
	// bookkeeping at all.
	tableHeadered []bool
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
		// An aside's headings describe the aside, not the document around
		// it: save the path on entry and restore it on exit, so the blocks
		// AFTER the aside inherit their own ancestors' headings, not the
		// aside's (wrong provenance and wrong same-heading rescue
		// grouping). The save must be a true COPY of the stack, not the
		// slice header: a same-or-higher level heading inside the aside
		// truncates the stack and the following append reuses the freed
		// slot of the shared backing array — overwriting the saved entry —
		// so the restored path would carry the aside's heading.
		//
		// The itemDepth save/reset (F7, t1968 sync-audit iter2) scopes the
		// list-paragraph suppression per aside entry: an aside inside a list
		// item is scored separately from that item, so its paragraphs are
		// NOT "already part of the item's emitted text" and must be emitted
		// as the aside's own blocks.
		//
		// C2 (t1968 sync-audit iter6): bare text runs of a MIXED aside are
		// content in their own right. The F7b fallback was all-or-nothing —
		// it emitted the aside's collapsed text only when NO child had
		// emitted a block, so `Important<p>Details</p>Tail` kept only
		// Details and both bare runs were lost. The walk now processes each
		// child in order; the bare run preceding a block-emitting child is
		// spliced in front of that child's blocks (F7b/F7c semantics per
		// run: heading subtrees excluded, a heading-only remainder collapses
		// to nothing), and a trailing run is appended — every bare run lands
		// in document order as aside-scored content. The aside itself is a
		// run container whose trailing run is content unconditionally (the
		// F7b fallback); wrapper children recurse through the same walk
		// (leader codex-audit round 3 — see walkAsideRuns).
		saved := append([]heading(nil), s.path...)
		savedDepth := s.itemDepth
		s.itemDepth = 0
		s.walkAsideRuns(n, true)
		s.itemDepth = savedDepth
		s.path = saved
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
		if text := collapseText(n, listItemSkip); text != "" {
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
		s.tableStack = append(s.tableStack, s.tableSeq)
		s.walkChildren(n, inAside)
		s.tableStack = s.tableStack[:len(s.tableStack)-1]
	case n.Data == "tr":
		s.emitRow(n, inAside)
		s.emitRowSubBlocks(n, inAside)
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
		if inAside && s.itemDepth == 0 {
			// Leader codex-audit round 3, finding 1 (C2 class): a default-case
			// element inside an aside is a run container too — the same
			// per-run extraction the aside itself applies, so bare runs
			// nested inside wrapper children (`<aside><div>Important
			// <p>Details</p>Tail</div></aside>`) are no longer swallowed by
			// the child's own block emission. Outside asides the plain
			// descent stands: body-level bare text is not a block by design.
			//
			// The extraction is scoped to itemDepth == 0 (F1, t1968 sync-audit
			// iter10): inside a list-item body the li block already carries the
			// wrapper's inline text (collapseText folds it through
			// listItemSkip), so splicing the wrapper's bare runs around a
			// block-emitting descendant duplicates the item's text
			// (`<aside><ul><li><div>LEAD<pre>CODE</pre>TAIL</div></li></ul>
			// </aside>` emitted LEAD and TAIL twice). The plain descent lets
			// only separate-block descendants (pre/table, nested li, nested
			// aside via its own case) re-enter emission. An aside nested in an
			// item resets itemDepth on entry, so wrappers inside THAT aside
			// keep the recursive extraction — the F7 shape.
			s.walkAsideRuns(n, false)
		} else {
			s.walkChildren(n, inAside)
		}
	}
}

// walkAsideRuns processes the children of an aside-scoped container with the
// C2 per-run bare-text extraction: each bare run preceding a block-emitting
// child is spliced in front of that child's blocks, and a trailing run is
// appended after them — every run lands in document order as aside-scored
// content.
//
// Leader codex-audit round 3, finding 1: the extraction previously ran only
// over an aside's DIRECT children, so bare runs nested inside a wrapper
// child that itself contains block-emitting descendants never reached it.
// Every default-case element inside an aside now recurses through this same
// walk, so a wrapper's interior runs are spliced around the blocks its
// descendants emit, recursively, still in document order.
//
// R2 (t1968 sync-audit iter7/8 P3, debt-closure round): a HEADING child is a
// run boundary too — the run before it is emitted with the PRE-heading path
// before the heading's own walk advances the path. A heading's own text is
// structure — the heading itself emits no block, so the pre-heading run is
// appended rather than spliced: nothing after runStart has emitted a block it
// would need to precede. The heading's SUBTREE, however, can emit blocks
// (`<h3><pre>CODE</pre></h3>`); F1 (t1968 sync-audit iter11) counts that
// emission toward `emitted`, mirroring the default branch, so a WRAPPER whose
// heading carries blocks keeps its trailing run instead of dropping it. A
// heading-only boundary (empty run before it) emits nothing, so a
// heading-only aside stays empty and the leading-heading shape
// (`<aside><h3>Side</h3>note</aside>`) keeps its single post-heading
// paragraph. The boundary applies where the heading is the container's own
// child; a heading inside a no-block wrapper stays merged into the enclosing
// run per the emitTrailingAlways rule below.
//
// emitTrailingAlways distinguishes the aside itself (true: an aside of pure
// inline content is still a note, so its trailing run is emitted — the F7b
// fallback) from a wrapper (false): a wrapper whose interior produced NO
// block is pure inline content that merges into the ENCLOSING run — emitting
// it separately would split runs that were one paragraph
// (`<aside>head<div>deep body</div>tail</aside>` is ONE paragraph) — while a
// wrapper whose interior DID produce a block is mixed, and its trailing run
// is content exactly like the aside's own.
func (s *splitter) walkAsideRuns(n *html.Node, emitTrailingAlways bool) {
	runStart := n.FirstChild
	emitted := false
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		pathBefore := clonePath(s.path)
		if child.Type == html.ElementNode && headingLevels[child.Data] > 0 {
			// R2 heading boundary: emit the pre-heading run under its own
			// (pre-heading) path, then advance the run start past the
			// heading. A non-empty emission marks the container mixed, so a
			// WRAPPER's trailing run after the heading is content too.
			if text := collapseRun(runStart, child, asideFallbackSkip); text != "" {
				s.emit(Block{Kind: KindParagraph, HeadingPath: pathBefore, Text: text, Aside: true, HeaderIndex: -1}, 0)
				emitted = true
			}
			before := len(*s.blocks)
			s.walk(child, true)
			if len(*s.blocks) > before {
				// F1 (t1968 sync-audit iter11): the heading's subtree can
				// emit blocks of its own (`<h3><pre>CODE</pre></h3>` — a
				// heading is not guaranteed text-only). That emission marks
				// the container mixed exactly like a block-emitting default
				// child, so a WRAPPER's trailing run after the heading is
				// content, not a lost remainder.
				emitted = true
			}
			runStart = child.NextSibling
			continue
		}
		before := len(*s.blocks)
		s.walk(child, true) // a no-op for non-element nodes
		if len(*s.blocks) == before {
			continue // emitted nothing: the child extends the bare run
		}
		emitted = true
		if text := collapseRun(runStart, child, asideFallbackSkip); text != "" {
			s.spliceAsideRun(before, text, pathBefore)
		}
		runStart = child.NextSibling
	}
	if emitTrailingAlways || emitted {
		if text := collapseRun(runStart, nil, asideFallbackSkip); text != "" {
			s.emit(Block{Kind: KindParagraph, HeadingPath: clonePath(s.path), Text: text, Aside: true, HeaderIndex: -1}, 0)
		}
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
	if b.Kind == KindTableRow {
		s.rowOf = append(s.rowOf, table)
	} else {
		s.rowOf = append(s.rowOf, -1)
	}
	s.headerAt = append(s.headerAt, -1)
	*s.blocks = append(*s.blocks, b)
}

// emitRow appends a table-row block. The first row of a table whose cells
// are th elements is that table's header row.
func (s *splitter) emitRow(n *html.Node, inAside bool) {
	cells := rowCells(n)
	if len(cells) == 0 {
		return
	}
	// Bind to the innermost OPEN table, not the raw counter (R1, t1968
	// sync-audit iter7 — see the tableStack field note): a nested table inside
	// an in-cell aside has already opened and closed by the time a later row
	// of the outer table emits, so the stack top is still the outer table's
	// id. The counter fallback preserves the pre-stack behavior for a stray
	// row outside any table (unreachable via html.Parse's HTML5 tree
	// construction, which drops such rows, but kept as a safe default).
	table := s.tableSeq
	if depth := len(s.tableStack); depth > 0 {
		table = s.tableStack[depth-1]
	}
	isHeader := false
	if !s.tableHasHeader(table) && rowIsHeader(n) {
		isHeader = true
		s.noteTableHeader(table)
	}
	s.emit(Block{
		Kind:        KindTableRow,
		HeadingPath: clonePath(s.path),
		Text:        strings.Join(cells, " | "),
		Aside:       inAside,
		HeaderIndex: -1,
	}, table)
	if isHeader {
		s.headerAt[len(s.headerAt)-1] = table
	}
}

// tableHasHeader reports whether table's first row was a header row. An id
// beyond the allocated range (the stray-row counter fallback below emitRow's
// binding) reads as headerless — the same miss the map lookup it replaced
// produced for an unknown id.
func (s *splitter) tableHasHeader(table int) bool {
	return table >= 0 && table < len(s.tableHeadered) && s.tableHeadered[table]
}

// noteTableHeader marks table headered, growing the per-table record to
// cover its id first.
func (s *splitter) noteTableHeader(table int) {
	for len(s.tableHeadered) <= table {
		s.tableHeadered = append(s.tableHeadered, false)
	}
	s.tableHeadered[table] = true
}

// resolveHeaders fills each table row's HeaderIndex from the recorded
// per-table header position. The position-keyed headerAt record is inverted
// into a table→position map in one pass here — positions stayed current
// through every splice, so no deferred shifting is replayed.
func (s *splitter) resolveHeaders() {
	blocks := *s.blocks
	headerByTable := make(map[int]int, len(s.headerAt))
	for i, table := range s.headerAt {
		if table >= 0 {
			headerByTable[table] = i
		}
	}
	for i := range blocks {
		if blocks[i].Kind != KindTableRow {
			continue
		}
		if h, ok := headerByTable[s.rowOf[i]]; ok && h != i {
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
// (a list item's nested list must not leak into its own text). Chrome
// subtrees (chromeTags — a script inside a paragraph, a style inside a table
// cell) are ALWAYS excluded, with a separator so the text runs on either
// side keep their word boundary across the removed subtree. A text boundary
// element (<br>, a block container) emits a separator on each side so
// adjacent text runs across the boundary keep their word boundary; the final
// collapse reduces the separators to single spaces, never to nothing.
func collapseText(n *html.Node, skip map[string]bool) string {
	return collapseRun(n.FirstChild, nil, skip)
}

// collapseRun collapses the text of the sibling run [first, end) — end nil
// means to the last sibling — with the same visitor collapseText applies
// over one node's children. C2's per-run extraction walks an aside's
// children one at a time and needs the collapsed text of an explicit
// SIBLING RANGE (the bare run between two block-emitting children), which a
// single node's children cannot express.
func collapseRun(first, end *html.Node, skip map[string]bool) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && chromeTags[node.Data] {
			// Chrome nested inside an extracted element is dropped with its
			// subtree; without the separator the surrounding runs would
			// merge into one word across the hole.
			b.WriteByte(' ')
			return
		}
		if node.Type == html.ElementNode && skip != nil && skip[node.Data] {
			// The skipped subtree (a nested list, pre, or table inside a list
			// item) is emitted as its own block during the descent; the hole
			// it leaves here needs the same boundary separator the chrome
			// branch writes, or the text runs on either side merge across it
			// ("FirstThird" — F3, t1968 sync-audit).
			b.WriteByte(' ')
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
	for c := first; c != nil && c != end; c = c.NextSibling {
		visit(c)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// spliceAsideRun inserts one bare-run paragraph at index at, in front of the
// blocks a block-emitting child of an aside just produced (C2, t1968
// sync-audit iter6), and shifts the position-keyed table bookkeeping
// (rowOf, headerAt) with the same move the blocks slice itself takes, so
// index i keeps describing blocks[i] and resolveHeaders still binds rows to
// their headers after the shift. The shift costs only the tail it moves —
// the same tail the blocks copy already moves — instead of re-keying every
// row of every earlier table (P2 perf, card t1968 repair round: an aside of
// M crossings after an N-row table paid O(N×M) map reinsertions and an
// N-entry allocation per splice here).
func (s *splitter) spliceAsideRun(at int, text string, path []string) {
	b := Block{Kind: KindParagraph, HeadingPath: path, Text: text, Aside: true, HeaderIndex: -1}
	blocks := *s.blocks
	blocks = append(blocks, Block{})
	copy(blocks[at+1:], blocks[at:])
	blocks[at] = b
	*s.blocks = blocks
	s.rowOf = append(s.rowOf, -1)
	copy(s.rowOf[at+1:], s.rowOf[at:])
	s.rowOf[at] = -1
	s.headerAt = append(s.headerAt, -1)
	copy(s.headerAt[at+1:], s.headerAt[at:])
	s.headerAt[at] = -1
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

// listItemSkip excludes from a list item's own text the children that are
// separately emitted as their own blocks during the descent: nested lists
// (their items are blocks in their own right), pre/table (emitted as code
// and table-row blocks), and aside (F7, t1968 sync-audit iter2 — the aside's
// subtree is scored as separately-flagged aside blocks, so an aside inside a
// list item gets the same separation as at top level instead of riding the
// item's text). Folding them into the item text would duplicate the content,
// flatten a code block's preserved whitespace, or contaminate the item's
// verbatim text with aside content.
var listItemSkip = map[string]bool{"ul": true, "ol": true, "pre": true, "table": true, "aside": true}

// cellSkip excludes from a table cell's row text the subtrees emitted as
// their own blocks (F8 + C3 + P2, t1968 sync-audit / card t1968 codex
// review): a pre inside a cell is a code block whose interior whitespace IS
// the code's formatting — the row's whitespace collapse would destroy it; an
// aside inside a cell is aside-scored content that must reach the aside
// stream, not fold into the row's text; and a table nested directly in a
// cell is emitted as its own table-row blocks, so the parent row reads only
// its own content around the boundary separator (the same exclusion
// listItemSkip applies to nested tables).
var cellSkip = map[string]bool{"pre": true, "aside": true, "table": true}

// asideFallbackSkip excludes heading subtrees from the F7b aside fallback's
// text extraction (F7c, t1968 sync-audit iter4): a heading's text is
// structure — it joined the aside-local path during the walk, exactly as a
// body-level heading does — not paragraph content. Without the skip, a
// heading-only aside manufactured a scored paragraph from pure structure
// (`<aside><h2>Note title</h2></aside>` → paragraph "Note title"), and a
// heading+body aside merged the heading's words into the body text
// ("Side note body" where the p-wrapped control emits "note body"). The
// skipped heading leaves a boundary separator, so a heading-only remainder
// collapses to the empty string and the fallback's guard emits no block.
var asideFallbackSkip = map[string]bool{
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// preText extracts a code block's text preserving line structure. Only the
// surrounding blank lines are trimmed; interior whitespace — including the
// first real code line's leading indentation — is kept verbatim. Chrome
// subtrees (script/style — F4, t1968 sync-audit) are dropped with a boundary
// separator, the same exclusion collapseText applies; the code text either
// side of the hole is otherwise untouched. A <br> is a line break in the
// code itself (P2, card t1968 codex review): it emits a newline boundary so
// the text runs on either side keep their lines; every other byte of code
// spacing stands verbatim.
func preText(n *html.Node) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && chromeTags[node.Data] {
			b.WriteByte(' ')
			return
		}
		if node.Type == html.ElementNode && node.Data == "br" {
			b.WriteByte('\n')
			return
		}
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

// emitRowSubBlocks emits the separately-scored blocks nested inside a table
// row (F8 + C3, t1968 sync-audit). The walk never descends into cells
// (emitRow reads them as text), so a pre or an aside inside a td was
// unreachable — the pre's only trace was the flattened, whitespace-collapsed
// row text (F8), and the aside's content rode the row's text without ever
// reaching the aside scoring stream (C3). The row's extraction excludes both
// subtrees (cellSkip), and this scan emits each one after its row in
// document order: the pre verbatim as its own code block, the aside handed
// to the aside walk itself (whose paragraphs and bare runs, C2, become
// aside-flagged blocks) — the same exclusion+separate-emission pattern the
// list-item skip applies.
func (s *splitter) emitRowSubBlocks(tr *html.Node, inAside bool) {
	var scan func(*html.Node)
	scan = func(node *html.Node) {
		if node.Type == html.ElementNode && chromeTags[node.Data] {
			// Chrome subtree (F8b, t1968 sync-audit iter3): a pre inside
			// nav/template/footer inside a cell is chrome like everywhere
			// else — the body-level walk drops the whole subtree, and this
			// scan must not re-emit its content as a code block.
			return
		}
		if node.Type == html.ElementNode && node.Data == "pre" {
			if text := preText(node); text != "" {
				s.emit(Block{Kind: KindCode, HeadingPath: clonePath(s.path), Text: text, Aside: inAside, HeaderIndex: -1}, 0)
			}
			return // a pre nested in a pre is carried by the outer block's text
		}
		if node.Type == html.ElementNode && node.Data == "aside" {
			// C3: the aside subtree is aside-scored content — hand it to the
			// aside walk, whose own emission flags every block it produces.
			// R1 (t1968 sync-audit iter7): a nested <table> inside this aside
			// advances tableSeq mid-outer-table, but no counter restore is
			// needed HERE — rows bind to tableStack's top, and the nested
			// table's open/close is balanced inside this walk, so the outer
			// table's later rows rebind to their own id (see tableStack).
			s.walk(node, true)
			return // the walk covers the subtree; scanning on would duplicate
		}
		if node.Type == html.ElementNode && node.Data == "table" {
			// P2 (card t1968 codex review): a <table> nested directly in a
			// cell emits through the same table machinery the body-level walk
			// uses — the walk's table case opens its own id on tableStack, so
			// its rows bind to THEIR table and its header row (headerAt)
			// binds to it, exactly like the aside-nested table of R1 iter7.
			// Scanning on would descend into the rows without ever emitting
			// them (no tr branch here) while emitRow had already folded their
			// flattened text into the parent cell.
			s.walk(node, inAside)
			return // the walk covers the subtree; scanning on would duplicate
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			scan(c)
		}
	}
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		scan(c)
	}
}

// rowCells lists a table row's cell texts in order.
func rowCells(tr *html.Node) []string {
	var cells []string
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			cells = append(cells, collapseText(c, cellSkip))
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
