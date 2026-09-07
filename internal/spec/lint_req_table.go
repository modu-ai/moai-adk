package spec

import (
	"regexp"
	"strings"
)

// Table-form REQ collection — SPEC-SPEC-LINT-BLIND-AXES-001 axis 1 (card t518).
//
// THE DEFECT. reqLineWidePattern is anchored to a markdown list bullet
// (`^\s*[-*]\s+`), so a SPEC that states its requirements as a TABLE produced an
// empty doc.REQs. The four rules that consume doc.REQs — ModalityMalformed,
// InvalidREQID, DuplicateREQID, CoverageIncomplete — then never visited that
// document at all, and their silence was indistinguishable from a pass. Measured
// on this tree: 53 spec.md files match the list collector zero times while
// carrying at least one ID-leading table row, contributing 600 such rows.
//
// THE DISCRIMINATOR. A leading REQ-ID cell does NOT make a table a definition
// table: the corpus also uses that shape for DISPOSITION tables (what became of
// this requirement) and TRACKING tables (which AC covers it). Telling those
// apart is the design question this axis had to answer, and the operator
// answered it by naming candidate C-d: a row is a definition only when the row
// carries a marker from the narrow lexicon L1 (reqDefinitionLexiconL1 below).
//
// WHY C-d AND NOT A WIDER LEXICON. The selection criterion was REPRODUCIBILITY
// BETWEEN INDEPENDENT DERIVATIONS, not the size of the surviving set. C-d
// produced the same count (20 surviving rows inside the 53 blind SPECs) under
// two lexicons written independently of each other; the wider candidate C-e
// moved between 61 and 63 across those same two lexicons — i.e. its value
// depends on who wrote the word list.
//
// [HARD] THE MISS IS A CHOSEN COST, NOT A BUG, AND L1 IS NOT TO BE WIDENED TO
// "FIX" IT. A real definition row whose verb ending L1 does not contact is
// rejected — the measured instance is SPEC-INIT-001's
// `| REQ-N-001 | 시스템은 … 덮어쓰지 않아야 한다 |`. Adding endings to L1 would
// reduce the miss and destroy the very property C-d was chosen for. Reducing
// the miss is future work for a discriminator that does not depend on a word
// list at all (header-vocabulary reading, or a structural criterion), and that
// route requires a fresh operator decision.
//
// SEVERITY AND PROVENANCE ARE SEPARATE CONCERNS. Table-collected entries are
// demoted to advisory through the EXISTING mechanism (REQEntry.Widened, read by
// reqFindingSeverity) — no second severity axis is created. Where the entry came
// from is recorded on a SEPARATE field (REQEntry.Source) whose only consumer is
// the corpus recount; it never reaches a severity decision. See REQEntry.Source
// and spec.md §B.2.

// reqTableRowPattern matches a markdown table row whose FIRST cell is a REQ ID,
// optionally wrapped in bold markers. Verbatim from spec.md §G, and identical to
// the ROW regex in .moai/reports/t518/blind-axes-reader.py — the two must not
// drift, because the reader is what produces the corpus figures this collector
// is measured against.
var reqTableRowPattern = regexp.MustCompile(`^\s*\|\s*\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*\|`)

// reqDefinitionLexiconL1 is lexicon L1 — the narrow modality lexicon that
// discriminator C-d keys on. Matching is SUBSTRING, CASE-SENSITIVE, with no word
// boundary, applied to the WHOLE raw row line (not to any single cell). Every
// one of those four properties is copied from the reader that produced the
// corpus figures; changing any of them changes the figures.
//
// [HARD] This list is duplicated, deliberately and by decision, in three places:
// spec.md §I (the operator read it there), .moai/reports/t518/blind-axes-reader.py
// (the corpus figures come from there), and here (the shipped behaviour). Change
// one and they diverge silently. Any edit re-derives the corpus figures.
var reqDefinitionLexiconL1 = []string{
	"SHALL",
	"해야 한다",
	"해서는 안 된다",
}

// isTableDefinitionRow applies discriminator C-d to a raw table row line.
func isTableDefinitionRow(line string) bool {
	for _, marker := range reqDefinitionLexiconL1 {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

// tableRowCells splits a markdown table row into trimmed cells, dropping the
// leading and trailing pipe. It mirrors the reader's cells() helper.
func tableRowCells(line string) []string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "|")
	s = strings.TrimSuffix(s, "|")
	parts := strings.Split(s, "|")
	cells := make([]string, 0, len(parts))
	for _, c := range parts {
		cells = append(cells, strings.TrimSpace(c))
	}
	return cells
}

// tableRowBody returns the requirement statement carried by a definition row:
// the LAST non-empty cell after the ID cell.
//
// The choice is deliberate and has a named cost. A definition table is either
// two-column (`| ID | statement |`) or three-column with a modality column
// (`| ID | Ubiquitous | statement |`), and in both the statement is the last
// cell — whereas joining the trailing cells would prepend the modality word to
// the text and break the prefix that isModalityMalformed reads. The cost: a
// definition table that appends a trailing NOTE column yields the note rather
// than the statement. No such shape was observed while measuring this tree, and
// the entry is advisory in any case.
func tableRowBody(cells []string) string {
	for i := len(cells) - 1; i >= 1; i-- {
		if cells[i] != "" {
			return cells[i]
		}
	}
	return ""
}

// parseREQsTable returns one REQEntry per table row that discriminator C-d
// admits as a definition, in document order, with Line as a 1-based index into
// body.
//
// Every returned entry carries Widened = true: no table row was EVER reachable
// by the narrow pattern, so a finding against one is a newly-surfaced corpus
// fact rather than a regression an author introduced. It therefore reports
// without gating, on exactly the terms t385 established for widened list
// entries.
func parseREQsTable(body string) []REQEntry {
	var reqs []REQEntry
	for i, line := range strings.Split(body, "\n") {
		matches := reqTableRowPattern.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		if !isTableDefinitionRow(line) {
			continue
		}
		reqs = append(reqs, REQEntry{
			ID:      matches[1],
			Text:    tableRowBody(tableRowCells(line)),
			Line:    i + 1,
			Widened: true,
			Source:  REQSourceTable,
		})
	}
	return reqs
}

// mergeREQsByLine merges two document-ordered entry slices into one, ordered by
// Line. A list definition and a table definition cannot occupy the same line, so
// no de-duplication is needed; ties (impossible in practice) keep the list entry
// first, which preserves the pre-existing ordering exactly.
func mergeREQsByLine(list, table []REQEntry) []REQEntry {
	if len(table) == 0 {
		return list
	}
	merged := make([]REQEntry, 0, len(list)+len(table))
	i, j := 0, 0
	for i < len(list) && j < len(table) {
		if list[i].Line <= table[j].Line {
			merged = append(merged, list[i])
			i++
			continue
		}
		merged = append(merged, table[j])
		j++
	}
	merged = append(merged, list[i:]...)
	merged = append(merged, table[j:]...)
	return merged
}
