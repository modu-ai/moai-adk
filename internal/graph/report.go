// report.go — the report cross-check edge layer (t107).
//
// A report that declares milestones carries a mandatory "## Card
// Cross-Check" section: one markdown table per report mapping each
// milestone to the queue card that delivers it (a tNN id, or an explicit
// "[new card needed]" marker when no card exists yet). This layer turns
// that table into graph edges so the reader can detect milestones with no
// live card without re-reading the report:
//
//	kind "report-milestone"  report → milestone   every table row
//	kind "milestone-card"    milestone → card     each tNN in the card column
//
// Node shapes: the report node is the repo-relative .moai/reports path;
// the milestone node is "<report-stem>#<milestone-id>" (stem-qualified so
// two reports declaring S0 never collide); the card node is the bare
// queue id (tNN — one queue per repository, see internal/kanban).
package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Edge kinds emitted by reportEdges.
const (
	// KindReportMilestone is a report→milestone edge: the report's
	// cross-check table declares the milestone.
	KindReportMilestone = "report-milestone"
	// KindMilestoneCard is a milestone→card edge: the cross-check row
	// claims the queue card delivers the milestone.
	KindMilestoneCard = "milestone-card"
)

// reportsDirRelPath is the scanned report directory (top-level *.md only —
// subdirectories such as backlog/ hold working notes, not milestone
// reports).
const reportsDirRelPath = ".moai/reports"

// cardCrossCheckHeader is the mandatory section header prefix. The exact
// canonical token is English so the distributed discipline stays
// locale-neutral; a parenthetical locale gloss may follow it
// ("## Card Cross-Check (카드 대조표)").
const cardCrossCheckHeader = "## Card Cross-Check"

// cardTokenRe matches a queue card id (tNN) inside a card-column cell.
// Word-bounded so "t100" never matches inside a longer token.
var cardTokenRe = regexp.MustCompile(`\bt([0-9]+)\b`)

// reportEdges extracts report→milestone and milestone→card edges from
// every top-level .moai/reports/*.md file carrying a Card Cross-Check
// section. Fail-open like the other layers: an absent reports directory or
// a file without the section contributes zero edges.
func reportEdges(projectRoot string) ([]Edge, error) {
	dir := filepath.Join(projectRoot, filepath.FromSlash(reportsDirRelPath))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("graph: read reports dir: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names) // deterministic emission order across filesystems

	var edges []Edge
	seen := map[string]bool{}
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue // vanished between list and read: not an artifact event
			}
			return nil, fmt.Errorf("graph: read report %s: %w", name, err)
		}
		stem := strings.TrimSuffix(name, ".md")
		rel := reportsDirRelPath + "/" + name
		for _, row := range parseCardCrossCheck(string(content)) {
			milestone := stem + "#" + row.milestone
			addReportEdge := func(kind, source, target string) {
				key := kind + "\x00" + source + "\x00" + target
				if seen[key] {
					return
				}
				seen[key] = true
				edges = append(edges, Edge{Kind: kind, Source: source, Target: target})
			}
			addReportEdge(KindReportMilestone, rel, milestone)
			for _, card := range row.cards {
				addReportEdge(KindMilestoneCard, milestone, card)
			}
		}
	}
	return edges, nil
}

// crossCheckRow is one parsed table row: the milestone id and the queue
// cards its card column claims (empty when the row marks the milestone as
// needing a new card).
type crossCheckRow struct {
	milestone string
	cards     []string
}

// parseCardCrossCheck extracts the cross-check rows from one report body.
// The section starts at a header line prefixed "## Card Cross-Check"; the
// FIRST markdown table inside the section is the cross-check table. The
// card column is located by its header cell (case-insensitive "card" or
// "카드") — never by position, and card ids are read ONLY from that cell,
// so a milestone description mentioning another card ("absorbs t21")
// cannot pollute the mapping. Rows with an empty first cell are ignored.
func parseCardCrossCheck(content string) []crossCheckRow {
	lines := strings.Split(content, "\n")

	start := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), cardCrossCheckHeader) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return nil
	}

	table := tableLines(lines[start:])
	if len(table) == 0 {
		return nil
	}

	cells := make([][]string, len(table))
	for i, line := range table {
		cells[i] = splitTableRow(line)
	}
	cardCol := cardColumnIndex(cells[0])
	if cardCol < 0 {
		return nil
	}

	var rows []crossCheckRow
	seen := map[string]bool{}
	for _, row := range cells[2:] { // [0] header, [1] separator
		if len(row) == 0 {
			continue
		}
		milestone := cleanCell(row[0])
		if milestone == "" || seen[milestone] {
			continue
		}
		seen[milestone] = true
		rowCards := rowCards(row, cardCol)
		rows = append(rows, crossCheckRow{milestone: milestone, cards: rowCards})
	}
	return rows
}

// tableLines returns the first run of consecutive table lines (leading-|
// markdown rows) in body, stopping at the first non-table line.
func tableLines(body []string) []string {
	var table []string
	for _, line := range body {
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			table = append(table, line)
		} else if len(table) > 0 {
			break
		}
	}
	return table
}

// splitTableRow splits a markdown table line into trimmed cells, dropping
// the empty cells a leading/trailing pipe produces — so cell indexes align
// whether or not the row is pipe-delimited at both ends. An interior empty
// cell (a genuinely empty milestone id) survives.
func splitTableRow(line string) []string {
	parts := strings.Split(strings.TrimSpace(line), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// cardColumnIndex finds the card column in a header row: the cell that is
// exactly "card" (case-insensitive) or "카드". -1 when absent.
func cardColumnIndex(header []string) int {
	for i, cell := range header {
		if c := strings.ToLower(cell); c == "card" || cell == "카드" {
			return i
		}
	}
	return -1
}

// cleanCell strips markdown decoration (bold/code) from a milestone cell.
func cleanCell(cell string) string {
	r := strings.NewReplacer("*", "", "`", "")
	return strings.TrimSpace(r.Replace(cell))
}

// rowCards reads the tNN tokens from a row's card cell. A cell index past
// the row's end reads as empty — a ragged table yields no cards, not a
// panic.
func rowCards(row []string, cardCol int) []string {
	if cardCol >= len(row) {
		return nil
	}
	var cards []string
	for _, m := range cardTokenRe.FindAllStringSubmatch(row[cardCol], -1) {
		cards = append(cards, "t"+m[1])
	}
	return cards
}
