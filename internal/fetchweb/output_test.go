package fetchweb

import (
	"encoding/json"
	"strings"
	"testing"
)

// output_test.go — SPEC-FETCH-WEB-001 M3 (REQ-FW-007). The output contract:
// a first line beginning `#` carrying the status, per kept block a `##`
// heading-path line followed by verbatim text, and a --json object carrying
// the same information.

func TestRenderMarkdown_Contract(t *testing.T) {
	res := Result{
		Status: StatusFiltered,
		Blocks: []Block{
			{Kind: KindParagraph, HeadingPath: []string{"Install", "Requirements"}, Text: "You need Go 1.23 or later."},
			{Kind: KindTableRow, HeadingPath: []string{"Install"}, Text: "--json | Emit JSON"},
		},
	}
	out := RenderMarkdown(res)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 5 {
		t.Fatalf("markdown = %q", out)
	}
	if !strings.HasPrefix(lines[0], "#") || !strings.Contains(lines[0], "filtered") {
		t.Errorf("first line %q must begin # and carry the status", lines[0])
	}
	if lines[1] != "## Install > Requirements" {
		t.Errorf("block heading line = %q, want %q", lines[1], "## Install > Requirements")
	}
	if lines[2] != "You need Go 1.23 or later." {
		t.Errorf("block text = %q, want verbatim", lines[2])
	}
	if !strings.HasPrefix(lines[3], "## ") {
		t.Errorf("second block must be preceded by a ## line, got %q", lines[3])
	}
}

func TestRenderMarkdown_EmptyStatusesHaveNoBlocks(t *testing.T) {
	for _, st := range []Status{StatusEmpty, StatusNoMatch, StatusFetchError} {
		out := RenderMarkdown(Result{Status: st})
		if !strings.HasPrefix(out, "# "+string(st)) {
			t.Errorf("%s: output %q must start with the status line", st, out)
		}
		if strings.Contains(out, "## ") {
			t.Errorf("%s: no blocks may follow an empty/error status", st)
		}
	}
}

func TestRenderJSON_SameInformation(t *testing.T) {
	res := Result{
		Status:      StatusNoMatch,
		TotalTokens: 42,
	}
	raw, err := RenderJSON(res, "https://example.com/doc", "what is this?")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("--json output is not parseable: %v\n%s", err, raw)
	}
	if doc["status"] != "no_match" {
		t.Errorf("status = %v", doc["status"])
	}
	if doc["url"] != "https://example.com/doc" || doc["question"] != "what is this?" {
		t.Errorf("url/question missing: %v", doc)
	}
	if code, ok := doc["exit_code"].(float64); !ok || int(code) != StatusNoMatch.ExitCode() {
		t.Errorf("exit_code = %v, want %d", doc["exit_code"], StatusNoMatch.ExitCode())
	}

	res2 := Result{
		Status:     StatusFiltered,
		Blocks:     []Block{{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "kept text"}},
		KeptTokens: 3,
	}
	raw2, err := RenderJSON(res2, "u", "q")
	if err != nil {
		t.Fatal(err)
	}
	var doc2 struct {
		Status string `json:"status"`
		Blocks []struct {
			Kind    string   `json:"kind"`
			Heading []string `json:"heading"`
			Text    string   `json:"text"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(raw2, &doc2); err != nil {
		t.Fatal(err)
	}
	if len(doc2.Blocks) != 1 || doc2.Blocks[0].Text != "kept text" || doc2.Blocks[0].Heading[0] != "A" {
		t.Errorf("blocks not carried verbatim in JSON: %+v", doc2.Blocks)
	}
}
