package spec

// lint_coverage_sibling_table.go — table-form REQ mappings in the sibling
// acceptance.md (card t561, GH #1696).
//
// WHAT THIS CLOSES. siblingAcceptanceCoveredREQIDs collects coverage with
// ExtractRequirementMappings, which recognises only the `maps REQ-…` form. The
// corpus mostly writes acceptance criteria as a table instead —
// `| AC-SA-001 | REQ-SA-001, 002, 007 | … |` — so a Tier M SPEC that maps every
// REQ in a table still read as uncovered. That is the #1696 false positive.
//
// WHY A SEPARATE EXTRACTOR. ExtractRequirementMappings is also called by the
// inline spec.md path (parseSingleACLine). Widening it would change inline
// coverage judgments, a surface other work is actively changing. This file is
// read only by the sibling path, so the inline path is untouched.
//
// THE THREE SCOPINGS, AND WHY ALL ARE CONSERVATIVE.
//   - Only rows that also carry an `AC-…` id are read. A REQ-first
//     traceability table (`| REQ | AC |`, `| REQ | Covered by |`) lists REQs
//     whose AC cell may be empty or say the AC is waived; 127 such rows exist in
//     the corpus. Counting them would report coverage where the author wrote
//     none.
//   - Only columns whose header names a requirement (REQ / requirement / 요구)
//     are read. A table with no such column contributes nothing, which is the
//     behavior before this file existed. Reading every column would count a REQ
//     named in a note or description column as covered, hiding a real gap.
//   - A bare numeric tail (`002`) expands to the prefix of a full REQ id only
//     when that full id precedes it, comma-separated, in the SAME cell. A tail
//     with no full id before it in its own cell is not expanded: inferring a
//     prefix from a neighbouring cell would again count something as covered
//     that the author did not write as a mapping.
//
// KNOWN LIMIT. Cells are split on `|`; a pipe inside inline code in a table cell
// splits that cell early. Such a cell can only lose a mapping, never invent one.

import (
	"regexp"
	"strings"
)

// requirementHeaderPattern decides whether a table header cell names a
// requirement column: `REQ`, `REQ ID`, `REQs`, `대응 REQ`, `Requirement(s)`,
// `요구사항`. `\b` keeps `Request` and similar words out.
var requirementHeaderPattern = regexp.MustCompile(`(?i)\breq(s|uirements?)?\b|요구`)

// tableSeparatorPattern matches a markdown table separator row such as
// `|---|:--:|`.
var tableSeparatorPattern = regexp.MustCompile(`^\|[\s:|-]*-[\s:|-]*$`)

// fullREQIDPattern matches a full REQ id whose last segment is numeric, so that
// a following bare tail has a prefix to expand from.
var fullREQIDPattern = regexp.MustCompile(`REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*`)

// numericTailPattern matches one comma-separated bare numeric tail at the start
// of the remaining cell text.
var numericTailPattern = regexp.MustCompile(`^\s*,\s*([0-9]+)\b`)

// tableRowACIDPattern matches an acceptance-criterion id. A table row is read as a
// mapping only when it names one.
var tableRowACIDPattern = regexp.MustCompile(`\bAC-[A-Z0-9]+(?:-[A-Z0-9]+)*`)

// siblingTableREQIDs returns the full REQ ids (with the `REQ-` prefix) mapped
// by requirement-headed columns of the markdown tables in text.
func siblingTableREQIDs(text string) []string {
	var ids []string
	lines := strings.Split(text, "\n")
	for i := 0; i+1 < len(lines); i++ {
		header := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(header, "|") || !tableSeparatorPattern.MatchString(strings.TrimSpace(lines[i+1])) {
			continue
		}
		var reqCols []int
		for col, cell := range splitTableRow(header) {
			if requirementHeaderPattern.MatchString(cell) {
				reqCols = append(reqCols, col)
			}
		}
		j := i + 2
		for ; j < len(lines); j++ {
			row := strings.TrimSpace(lines[j])
			if !strings.HasPrefix(row, "|") {
				break
			}
			if !tableRowACIDPattern.MatchString(row) {
				continue
			}
			cells := splitTableRow(row)
			for _, col := range reqCols {
				if col < len(cells) {
					ids = append(ids, cellREQIDs(cells[col])...)
				}
			}
		}
		i = j - 1
	}
	return ids
}

// splitTableRow returns the trimmed cells of one markdown table row.
func splitTableRow(row string) []string {
	row = strings.TrimSuffix(strings.TrimPrefix(row, "|"), "|")
	cells := strings.Split(row, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

// cellREQIDs returns the full REQ ids in one cell, expanding each bare numeric
// tail that follows a full id (directly, or through an earlier expanded tail)
// by comma within the same cell.
func cellREQIDs(cell string) []string {
	var ids []string
	for _, loc := range fullREQIDPattern.FindAllStringIndex(cell, -1) {
		full := cell[loc[0]:loc[1]]
		ids = append(ids, full)
		last := strings.LastIndex(full, "-")
		prefix := full[:last+1]
		rest := cell[loc[1]:]
		for {
			m := numericTailPattern.FindStringSubmatchIndex(rest)
			if m == nil {
				break
			}
			ids = append(ids, prefix+rest[m[2]:m[3]])
			rest = rest[m[1]:]
		}
	}
	return ids
}
