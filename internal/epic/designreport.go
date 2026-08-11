package epic

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// CanonicalMilestone is one row of the design-report slice table: an Mx ID
// plus its human label (the cell text after "Mx " trimmed).
type CanonicalMilestone struct {
	ID    string
	Label string
}

// CanonicalMilestones is the parsed result of a design-report slice table. It
// is the canonical list REQ-ES-005 grounds orphan detection against.
type CanonicalMilestones struct {
	Milestones []CanonicalMilestone
	Source     string // the path parsed
}

// sliceTableRowPattern matches one `<tr><td>M<N> <label></td>` row inside the
// slice table. It is HTML-narrow (KI-1): the regex keys on the `M<N> ` prefix
// inside the first `<td>` of a `<tr>` row and captures the label text up to the
// closing `</td>` or the first `<`. Fail-open: if no rows match, the parser
// returns an empty CanonicalMilestones (not an error).
var sliceTableRowPattern = regexp.MustCompile(`<tr>\s*<td>\s*(M\d+)\s+([^<]+)</td>`)

// ParseDesignReport parses a design-report HTML file's slice table and returns
// the canonical M0..Mx milestone list. The parser is fail-open (KI-1 / edge
// case E3): a missing slice table, an unreadable file, or a reformatted report
// returns an empty CanonicalMilestones, never an error — the caller treats an
// empty list as "no canonical list" and omits orphan detection (REQ-ES-005).
func ParseDesignReport(path string) (*CanonicalMilestones, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// fail-open: treat unreadable as "no canonical list".
		return &CanonicalMilestones{Source: path}, nil
	}
	body := string(data)

	// Narrow to the slice-table section: find a `<table>` block. We don't key
	// on the Korean heading text (`슬라이스`) because that's locale-fragile; we
	// scan every `<table>` for Mx-shaped rows instead, taking the first table
	// that yields ≥1 milestone. This matches the design.md §4 contract.
	result := &CanonicalMilestones{Source: path}
	tableStart := 0
	for {
		open := strings.Index(body[tableStart:], "<table")
		if open < 0 {
			break
		}
		blkStart := tableStart + open
		close := strings.Index(body[blkStart:], "</table>")
		if close < 0 {
			break
		}
		blk := body[blkStart : blkStart+close+len("</table>")]
		rows := sliceTableRowPattern.FindAllStringSubmatch(blk, -1)
		if len(rows) > 0 {
			for _, r := range rows {
				if len(r) >= 3 {
					result.Milestones = append(result.Milestones, CanonicalMilestone{
						ID:    r[1],
						Label: strings.TrimSpace(r[2]),
					})
				}
			}
			break // first table with Mx rows wins
		}
		tableStart = blkStart + len("</table>")
	}
	return result, nil
}

// DiscoverDesignReport implements the auto-discovery rule (design.md §4): scan
// `reportsDir` for files matching `^.*-<token-lowercased>-[a-z0-9]+\.html$` and
// return the lexicographically-first match. Returns "" + nil when zero files
// match (fail-open — the caller omits orphan detection).
func DiscoverDesignReport(token, reportsDir string) (string, error) {
	if token == "" || reportsDir == "" {
		return "", nil
	}
	pattern := fmt.Sprintf(`-.*%s-[a-z0-9]+\.html$`, strings.ToLower(token))
	re := regexp.MustCompile(pattern)

	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		// fail-open: missing reports dir → no canonical list.
		return "", nil
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".html") && re.MatchString(name) {
			matches = append(matches, filepath.Join(reportsDir, name))
		}
	}
	if len(matches) == 0 {
		return "", nil
	}
	sort.Strings(matches) // lexicographic for determinism
	return matches[0], nil
}
