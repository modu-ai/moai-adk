// render.go — block-aware rendering of an agent source into one harness view.
//
// An agent source is the shipped definition file plus optional block markers:
// a claude-only block and a codex-only block, each opened and closed by a
// marker on its own line. A view keeps the lines outside every block, keeps
// the blocks of its own harness and drops the other harness's blocks; the
// marker lines themselves never survive a render. A source without markers
// renders to itself in both views.
//
// The grammar is fail-closed: an unterminated, nested, stray or mismatched
// marker is an error naming the file and the marker, so a malformed block can
// never ship half-rendered text.
package agentemit

import (
	"bytes"
	"fmt"
	"strings"
)

// View selects the harness a source is rendered for.
type View int

const (
	// ViewClaude keeps claude-only blocks and drops codex-only blocks.
	ViewClaude View = iota
	// ViewCodex keeps codex-only blocks and drops claude-only blocks.
	ViewCodex
)

// blockKinds lists the block grammars: marker text and the view that keeps the block.
var blockKinds = []struct {
	open, close string
	keep        View
}{
	{"<!-- claude-only -->", "<!-- /claude-only -->", ViewClaude},
	{"<!-- codex-only -->", "<!-- /codex-only -->", ViewCodex},
}

// blockMarker reports whether line (without surrounding space) is a block
// marker, returning the index into blockKinds and whether it opens the block.
func blockMarker(line string) (kind int, opens, ok bool) {
	for i, k := range blockKinds {
		switch line {
		case k.open:
			return i, true, true
		case k.close:
			return i, false, true
		}
	}
	return 0, false, false
}

// RenderView renders src for the given view. filename only labels errors.
func RenderView(filename string, src []byte, v View) ([]byte, error) {
	var out bytes.Buffer
	open, openedAt := -1, 0
	for n, line := range strings.SplitAfter(string(src), "\n") {
		if line == "" {
			continue
		}
		lineNo := n + 1
		kind, opens, isMarker := blockMarker(strings.TrimSpace(line))
		if !isMarker {
			if open < 0 || blockKinds[open].keep == v {
				out.WriteString(line)
			}
			continue
		}
		marker := strings.TrimSpace(line)
		switch {
		case opens && open >= 0:
			return nil, fmt.Errorf("%s:%d: nested block marker %s inside the %s block opened at line %d", filename, lineNo, marker, blockKinds[open].open, openedAt)
		case opens:
			open, openedAt = kind, lineNo
		case open < 0:
			return nil, fmt.Errorf("%s:%d: closing marker %s without an open block", filename, lineNo, marker)
		case open != kind:
			return nil, fmt.Errorf("%s:%d: closing marker %s does not match the %s block opened at line %d", filename, lineNo, marker, blockKinds[open].open, openedAt)
		default:
			open = -1
		}
	}
	if open >= 0 {
		return nil, fmt.Errorf("%s:%d: unterminated block %s (no %s before the end of the file)", filename, openedAt, blockKinds[open].open, blockKinds[open].close)
	}
	return out.Bytes(), nil
}
