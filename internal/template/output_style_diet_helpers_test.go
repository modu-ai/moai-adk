// Helpers for the output-style prefix-diet guards (SPEC-PREFIX-DIET-001, card t1450).
//
// The unit extractor and the binding-token counter defined here are the normative
// implementation of spec.md section B. The independent python3 one-liner of AC-PFD-004 counts
// the same tokens a second way so the two implementations check each other.
package template

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf16"
)

// dietStyleFiles lists the three deployed output styles in a fixed order.
var dietStyleFiles = []string{"moai.md", "moai-easy.md", "moai-learn.md"}

// dietStyleTemplateDir is the template source directory, relative to the package directory.
const dietStyleTemplateDir = "templates/.claude/output-styles/moai"

// Binding tokens, in the order of the counter returned by dietTokenCounts.
var dietTokenNames = [4]string{"[HARD]", "MUST NOT", "MUST", "shall "}

var (
	dietHeadingRe = regexp.MustCompile(`^#{1,6}\s`)
	dietTableRe   = regexp.MustCompile(`^\s*\|`)
	dietListRe    = regexp.MustCompile(`^\s*([-*+]|\d+[.)])\s`)
	dietFenceRe   = regexp.MustCompile("^\\s*```")
)

// dietUTF16Len returns the length of s in UTF-16 code units (the JavaScript string length).
func dietUTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// dietSHA256 returns the hex SHA-256 of s.
func dietSHA256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// dietTokenCounts counts the binding tokens of spec.md section B without overlap: MUST NOT is
// counted first, and MUST only where it is not the start of MUST NOT.
func dietTokenCounts(s string) [4]int {
	mustNot := strings.Count(s, "MUST NOT")
	return [4]int{
		strings.Count(s, "[HARD]"),
		mustNot,
		strings.Count(s, "MUST") - mustNot,
		strings.Count(s, "shall "),
	}
}

func dietHasToken(c [4]int) bool {
	return c[0]+c[1]+c[2]+c[3] > 0
}

// dietSplitFrontmatter splits a style file into its frontmatter (through the closing --- line and
// its newline) and its body (everything after).
func dietSplitFrontmatter(text string) (head, body string, err error) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return "", "", fmt.Errorf("no frontmatter")
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return strings.Join(lines[:i+1], "\n") + "\n", strings.Join(lines[i+1:], "\n"), nil
		}
	}
	return "", "", fmt.Errorf("frontmatter not closed")
}

// dietExtractUnits segments a style body into units (spec.md section B). A unit is a run of lines;
// trailing blank lines belong to the preceding unit and leading blanks to the first unit, so the
// unit texts concatenated reproduce the body byte for byte. Code fences are opaque, a heading or a
// table row always opens a unit, and a list or a fence that follows a unit directly or after one
// blank line belongs to that unit.
func dietExtractUnits(body string) []string {
	lines := strings.Split(body, "\n")
	var units [][]string
	var cur []string
	started := false
	curKind := ""
	inFence := false
	gap := 0
	var pending []string
	for _, ln := range lines {
		blank := strings.TrimSpace(ln) == ""
		if inFence {
			cur = append(cur, ln)
			if dietFenceRe.MatchString(ln) {
				inFence = false
			}
			continue
		}
		if blank {
			gap++
			pending = append(pending, ln)
			continue
		}
		isHead := dietHeadingRe.MatchString(ln)
		isTable := dietTableRe.MatchString(ln)
		isList := dietListRe.MatchString(ln)
		isFence := dietFenceRe.MatchString(ln)
		var isNew bool
		switch {
		case !started:
			isNew = true
		case isHead || isTable:
			isNew = true
		case curKind == "table":
			isNew = true
		case gap == 0:
			isNew = false
		case gap == 1 && (isList || isFence):
			isNew = false
		default:
			isNew = true
		}
		if isNew {
			if started {
				cur = append(cur, pending...)
				units = append(units, cur)
				cur = nil
			} else {
				cur = append([]string(nil), pending...)
				started = true
			}
			pending = nil
			switch {
			case isTable:
				curKind = "table"
			case isHead:
				curKind = "heading"
			default:
				curKind = "text"
			}
		} else {
			cur = append(cur, pending...)
			pending = nil
		}
		cur = append(cur, ln)
		if isFence {
			inFence = true
		}
		gap = 0
	}
	if started {
		cur = append(cur, pending...)
		units = append(units, cur)
	}
	out := make([]string, len(units))
	for i, u := range units {
		out[i] = strings.Join(u, "\n")
		if i < len(units)-1 {
			out[i] += "\n"
		}
	}
	return out
}

// dietFirstLine returns the first non-blank line of a unit text.
func dietFirstLine(u string) string {
	for _, ln := range strings.Split(u, "\n") {
		if strings.TrimSpace(ln) != "" {
			return ln
		}
	}
	return ""
}

// dietSectionRange locates the section that opens at the unit whose first line equals title and
// runs up to (not including) the next heading of the same or a higher level, or the next --- unit.
func dietSectionRange(units []string, title string, level int) (first, last int, ok bool) {
	first = -1
	for i, u := range units {
		if dietFirstLine(u) == title {
			first = i
			break
		}
	}
	if first < 0 {
		return 0, 0, false
	}
	last = first
	for j := first + 1; j < len(units); j++ {
		fl := dietFirstLine(units[j])
		if strings.TrimSpace(fl) == "---" {
			break
		}
		if m := dietHeadingRe.FindString(fl); m != "" {
			n := len(strings.TrimRight(m, " \t"))
			if n <= level {
				break
			}
		}
		last = j
	}
	return first, last, true
}

// Fixture shapes -----------------------------------------------------------------------------

type dietRow struct {
	ID         string `json:"id"`
	File       string `json:"file"`
	Index      int    `json:"index"`
	Kind       string `json:"kind"`
	Source     string `json:"source"`
	BeforeText string `json:"before_text"`
	AfterText  string `json:"after_text"`
	Treatment  string `json:"treatment"`
	Survivor   string `json:"survivor"`
	Note       string `json:"note"`
}

type dietFrozenRef struct {
	Title      string `json:"title"`
	FirstIndex int    `json:"first_index"`
	LastIndex  int    `json:"last_index"`
}

type dietFileMeta struct {
	AnchorBodySHA256        string          `json:"anchor_body_sha256"`
	AnchorFrontmatterSHA256 string          `json:"anchor_frontmatter_sha256"`
	AnchorUnits             int             `json:"anchor_units"`
	AnchorFileUTF16         int             `json:"anchor_file_utf16"`
	AnchorTokens            [4]int          `json:"anchor_tokens"`
	Frozen                  []dietFrozenRef `json:"frozen"`
}

type dietLedger struct {
	Anchor string                   `json:"anchor"`
	Files  map[string]*dietFileMeta `json:"files"`
	Rows   []dietRow                `json:"rows"`
}

type dietFrozenSection struct {
	File       string `json:"file"`
	Title      string `json:"title"`
	Level      int    `json:"level"`
	FirstIndex int    `json:"first_index"`
	LastIndex  int    `json:"last_index"`
	SHA256     string `json:"sha256"`
	Text       string `json:"text"`
}

type dietFrozenDoc struct {
	Anchor   string              `json:"anchor"`
	Sections []dietFrozenSection `json:"sections"`
}

type dietLocalTable struct {
	File string   `json:"file"`
	ID   string   `json:"id"`
	Rows []string `json:"rows"`
}

type dietLocalDoc struct {
	Anchor string           `json:"anchor"`
	Tables []dietLocalTable `json:"tables"`
}

func dietLoadJSON(t *testing.T, name string, v any) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
}

// dietReadDeployed returns the whole text of each deployed (template) output-style file.
func dietReadDeployed(t *testing.T) map[string]string {
	t.Helper()
	out := make(map[string]string, len(dietStyleFiles))
	for _, f := range dietStyleFiles {
		raw, err := os.ReadFile(filepath.Join(dietStyleTemplateDir, f))
		if err != nil {
			t.Fatalf("read deployed %s: %v", f, err)
		}
		out[f] = string(raw)
	}
	return out
}

// dietClone deep-copies a ledger through JSON so a mutation never leaks into the shared fixture.
func dietClone(t *testing.T, led *dietLedger) *dietLedger {
	t.Helper()
	raw, err := json.Marshal(led)
	if err != nil {
		t.Fatalf("clone marshal: %v", err)
	}
	out := &dietLedger{}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("clone unmarshal: %v", err)
	}
	return out
}
