package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Citation is the canonical code-citation form under this SPEC (REQ-GF-010):
// a verbatim excerpt plus the content hash of the cited REGION, with the
// file path for lookup and a line number carried as convenience data only —
// never as the sole anchor. The hash covers the region's content, not the
// whole file, so unrelated edits elsewhere in the file and line drift above
// the region leave the anchor intact (AC-GF-013/014).
//
// TreeSHA co-stamps the measured tree (REQ-GF-003/008): a citation names the
// tree it was taken from, mirroring the provenance block's anchor.
type Citation struct {
	// File is the repo-relative path of the cited file (forward slashes).
	File string `json:"file"`
	// Excerpt is the verbatim source snippet of the cited region.
	Excerpt string `json:"excerpt"`
	// RegionHash is the sha256 of the normalized region content.
	RegionHash string `json:"region_hash"`
	// Line is the 1-based line of the region start AT CITATION TIME —
	// convenience notation; resolvers must never depend on it.
	Line int `json:"line,omitempty"`
	// TreeSHA is the commit (or "dirty:<fingerprint>") of the tree the
	// citation was taken from; empty when unknown.
	TreeSHA string `json:"tree_sha,omitempty"`
}

// NewCitation builds a citation from a verbatim excerpt, hashing the region.
// An empty excerpt is rejected: a line number alone is not an anchor.
func NewCitation(file, excerpt string, line int) (Citation, error) {
	if strings.TrimSpace(excerpt) == "" {
		return Citation{}, fmt.Errorf("graph: citation excerpt is empty — a line number alone is not a valid anchor (REQ-GF-010)")
	}
	return Citation{
		File:       filepath.ToSlash(file),
		Excerpt:    excerpt,
		RegionHash: NormalizeRegionHash(excerpt),
		Line:       line,
	}, nil
}

// NormalizeRegionHash hashes the region content under the citation
// normalization rule: each line trimmed, blank lines dropped, joined with
// newlines. Whitespace-only drift anywhere outside the region's non-blank
// lines therefore does not move the hash.
func NormalizeRegionHash(excerpt string) string {
	return sha256Hex(normalizeRegion(excerpt))
}

// normalizeRegion applies the line-trim + blank-drop normalization.
func normalizeRegion(excerpt string) string {
	var kept []string
	for _, line := range strings.Split(excerpt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// sha256Hex is the shared digest helper.
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Render emits the canonical citation form: file, convenience line, region
// hash, the measured tree when known, and the excerpt itself as a quoted
// block (the excerpt IS part of the canon — it carries its own verification
// data).
func (c Citation) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s", c.File)
	if c.Line > 0 {
		fmt.Fprintf(&b, " L%d", c.Line)
	}
	fmt.Fprintf(&b, " hash=%s", c.RegionHash)
	if c.TreeSHA != "" {
		fmt.Fprintf(&b, " tree=%s", c.TreeSHA)
	}
	b.WriteString("\n")
	for _, line := range strings.Split(c.Excerpt, "\n") {
		fmt.Fprintf(&b, "> %s\n", line)
	}
	return b.String()
}

// Resolution is the outcome of resolving a citation in one tree.
type Resolution struct {
	// File is the resolved repo-relative path.
	File string `json:"file"`
	// Line is the region's physical start line IN THIS TREE.
	Line int `json:"line"`
	// Matched is true when the region content was found and hash-confirmed.
	Matched bool `json:"matched"`
	// Reason explains a non-match (region edited, file missing, …).
	Reason string `json:"reason,omitempty"`
}

// ResolveCitation locates a citation's region in the tree at root by
// CONTENT: the normalized excerpt is matched within the named file, the
// region hash confirms the match, and the resolution names the region's
// physical line in THIS tree. The same citation resolves to the same target
// in two trees that differ only by line drift above the region (REQ-GF-012);
// a citation whose region itself was edited resolves UNMATCHED with a reason
// — honest staleness, never force-resolution.
func ResolveCitation(root string, c Citation) (Resolution, error) {
	target := filepath.Join(root, filepath.FromSlash(c.File))
	data, err := os.ReadFile(target)
	if err != nil {
		return Resolution{File: c.File, Reason: fmt.Sprintf("cited file unreadable: %v", err)}, nil
	}

	want := normalizeRegion(c.Excerpt)
	if want == "" {
		return Resolution{File: c.File, Reason: "citation carries no region content"}, nil
	}
	if c.RegionHash != "" && c.RegionHash != sha256Hex(want) {
		// The citation is internally inconsistent (hash does not cover its
		// own excerpt) — report rather than silently re-hashing.
		return Resolution{File: c.File, Reason: "citation region hash does not cover its excerpt"}, nil
	}

	lines := strings.Split(string(data), "\n")
	n := len(lines)

	// normalizedLen returns how many of lines[i:j] survive normalization and
	// their joined form — the sliding window over physical lines with blanks
	// allowed between region lines (the region may span blank lines).
	// Implementation: walk every start line, accumulate non-blank normalized
	// lines until the joined string is long enough to compare.
	for start := 0; start < n; start++ {
		if strings.TrimSpace(lines[start]) == "" {
			continue // a region starts on a non-blank line
		}
		var acc []string
		for end := start; end < n; end++ {
			line := strings.TrimSpace(lines[end])
			if line == "" {
				continue
			}
			acc = append(acc, line)
			joined := strings.Join(acc, "\n")
			if joined == want {
				return Resolution{File: c.File, Line: start + 1, Matched: true}, nil
			}
			if len(joined) >= len(want) {
				break // overshot — this start cannot match
			}
		}
	}
	return Resolution{File: c.File, Reason: "cited region not found — the region content itself changed (citation is stale)"}, nil
}
