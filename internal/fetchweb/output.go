package fetchweb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// output.go — SPEC-FETCH-WEB-001 M3. The output contract (REQ-FW-007): a
// first line beginning `#` carrying the status, per kept block a `##`
// heading-path line followed by the block's verbatim text; --json carries
// the same information as a machine-readable object.

// RenderMarkdown renders the result in the port's text shape. Kept blocks
// appear in document order, text verbatim — no paraphrase, no truncation.
// The empty and error statuses render the status line only. Control
// characters are escaped at this RENDER surface (escapeControlChars) — the
// verbatim-extraction guarantee lives on Block.Text, not on presentation.
func RenderMarkdown(res Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", res.Status)
	for _, blk := range res.Blocks {
		heading := strings.Join(blk.HeadingPath, " > ")
		if heading == "" {
			heading = "(no heading)"
		}
		fmt.Fprintf(&b, "## %s\n%s\n", heading, blk.Text)
	}
	return escapeControlChars(b.String())
}

// escapeControlChars returns s with every C0/C1 control character except
// \n and \t replaced by its visible \xHH form (P3, card t1968 codex
// review): RenderMarkdown feeds terminal-bound output, and a page's
// verbatim text can carry ESC/BEL (OSC 52 exfil shapes) or C1 codepoints
// that must not reach the terminal as live control bytes — only --json
// escaped before (encoding/json). The escaper is RUNE-based, never
// byte-based: C1 is U+0080–U+009F as codepoints, and their UTF-8
// encodings share bytes with multi-byte scripts — a byte-wise 0x80–0x9f
// replacement would corrupt CJK text.
func escapeControlChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r < 0x20 && r != '\n' && r != '\t') || (r >= 0x80 && r <= 0x9f) {
			fmt.Fprintf(&b, "\\x%02x", r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// jsonBlock is one kept block in the --json shape.
type jsonBlock struct {
	Kind    string   `json:"kind"`
	Heading []string `json:"heading"`
	Text    string   `json:"text"`
	Aside   bool     `json:"aside,omitempty"`
}

// jsonResult is the --json shape: the same information the markdown shape
// carries, machine-readable, including the exit-status mapping so a caller
// can branch on status without re-deriving the code.
type jsonResult struct {
	Status      string      `json:"status"`
	ExitCode    int         `json:"exit_code"`
	URL         string      `json:"url,omitempty"`
	Question    string      `json:"question,omitempty"`
	Condition   string      `json:"condition,omitempty"`
	Warnings    []string    `json:"warnings,omitempty"` // scoring-call warnings (REQ-FW-013 keep-and-warn), additive
	TotalTokens int         `json:"total_tokens"`       // approximate (package note)
	KeptTokens  int         `json:"kept_tokens"`        // approximate (package note)
	Blocks      []jsonBlock `json:"blocks"`
}

// RenderJSON renders the result as the --json object.
func RenderJSON(res Result, url, question string) ([]byte, error) {
	doc := jsonResult{
		Status:      string(res.Status),
		ExitCode:    res.Status.ExitCode(),
		URL:         url,
		Question:    question,
		Condition:   res.Condition,
		Warnings:    res.Warnings,
		TotalTokens: res.TotalTokens,
		KeptTokens:  res.KeptTokens,
	}
	if len(res.Blocks) > 0 {
		doc.Blocks = make([]jsonBlock, len(res.Blocks))
		for i, b := range res.Blocks {
			doc.Blocks[i] = jsonBlock{Kind: string(b.Kind), Heading: b.HeadingPath, Text: b.Text, Aside: b.Aside}
		}
	}
	return json.MarshalIndent(doc, "", "  ")
}
