package cli

// migrate_ownership_remaining.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865), REQ-OWR-009: the remaining-file list the migration returns and the way
// the update renders it.
//
// The list has one row per file the migration leaves in place under the four
// common-asset roots, in path order. It is rendered through the update's
// end-of-run Reference rows (decision index Q3, option (a): no new artifact on
// disk): by default one line per reason with a count, because 313 rows are not
// a default print (operator ruling d-20261012T221654Z-9ced); under --verbose
// one more row per file.
//
// While a per-file report line and the grouped list would say the same thing
// twice, the update's report callback drops the per-file line unless --verbose
// is on: every such line is a remaining row and is accounted for in the grouped
// view.

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// remainingReportMarker is the tag every per-file keep line of the migration
// carries (reasonTag): a line with it describes one remaining row.
const remainingReportMarker = " [reason: "

// sortRemainingRows puts the rows in path order, so the list is the same for the
// same tree however the walk visited the roots.
func sortRemainingRows(rows []remainingRow) {
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Rel < rows[j].Rel })
}

// isRemainingRowLine reports whether a migration report line is the per-file
// description of one remaining row that the grouped view may stand in for.
// A refused removal because the run backup failed is not such a row: it is a
// failure the operator must see on the run's own output, and it is bounded by
// the files the run would have removed.
func isRemainingRowLine(line string) bool {
	if strings.Contains(line, remainingReportMarker+string(keepBackupFailed)+"]") {
		return false
	}
	return strings.HasPrefix(line, "  migration: ") && strings.Contains(line, remainingReportMarker)
}

// remainingAwareReport is the report callback the update hands the migration:
// every line goes to out, except that without --verbose a per-file line of a
// remaining row stays out (the grouped list carries it).
func remainingAwareReport(out io.Writer, verbose bool) func(format string, args ...interface{}) {
	return func(format string, args ...interface{}) {
		line := fmt.Sprintf(format, args...)
		if !verbose && isRemainingRowLine(line) {
			return
		}
		_, _ = fmt.Fprintln(out, line)
	}
}

// remainingSummaryLines renders the grouped view: one line per reason, with the
// number of files and, when every row of the group lacks the same evidence
// items, the items. Reasons are in name order.
func remainingSummaryLines(rows []remainingRow) []string {
	type group struct {
		count   int
		missing string
		uniform bool
	}
	groups := map[keepReason]*group{}
	for _, r := range rows {
		m := missingList(r.Missing)
		g, ok := groups[r.Reason]
		if !ok {
			groups[r.Reason] = &group{count: 1, missing: m, uniform: true}
			continue
		}
		g.count++
		if g.missing != m {
			g.uniform = false
		}
	}
	reasons := make([]string, 0, len(groups))
	for r := range groups {
		reasons = append(reasons, string(r))
	}
	sort.Strings(reasons)
	lines := make([]string, 0, len(reasons))
	for _, name := range reasons {
		g := groups[keepReason(name)]
		label := name
		if label == "" {
			label = "(no reason recorded)"
		}
		line := fmt.Sprintf("kept in place: %d file(s) — %s", g.count, label)
		if g.uniform && g.missing != "" {
			line += " (missing evidence: " + g.missing + ")"
		}
		lines = append(lines, line)
	}
	return lines
}

// remainingDetailLines renders one row per file, in path order, with the class,
// the reason, the missing evidence items and the counterpart and profile-side
// states.
func remainingDetailLines(rows []remainingRow) []string {
	sorted := append([]remainingRow(nil), rows...)
	sortRemainingRows(sorted)
	lines := make([]string, 0, len(sorted))
	for _, r := range sorted {
		line := fmt.Sprintf("kept: %s — %s, %s", r.Rel, r.Class, r.Reason)
		if m := missingList(r.Missing); m != "" {
			line += "; missing evidence: " + m
		}
		if r.Counterpart != "" && r.Counterpart != counterpartStatusNotApplicable {
			line += "; user-root counterpart: " + string(r.Counterpart)
		}
		if r.ProfileSide != "" && r.ProfileSide != profileSideNotApplicable {
			line += "; profile side: " + string(r.ProfileSide)
		}
		if r.CaseVariantOf != "" {
			line += "; case-variant of: " + r.CaseVariantOf
		}
		lines = append(lines, line)
	}
	return lines
}

func missingList(items []evidenceItem) string {
	names := make([]string, len(items))
	for i, m := range items {
		names[i] = string(m)
	}
	return strings.Join(names, ", ")
}

// recordRemainingList writes the remaining list into the update's Reference
// rows: the grouped lines always, and under verbose the per-file rows after
// them. In the default view a closing line says how to list the files.
func recordRemainingList(l *updateActionLedger, rows []remainingRow, verbose bool) {
	if len(rows) == 0 {
		return
	}
	for _, line := range remainingSummaryLines(rows) {
		l.referencef("%s", line)
	}
	if !verbose {
		l.referencef("run `moai update --verbose` to list each kept file with its reason")
		return
	}
	for _, line := range remainingDetailLines(rows) {
		l.referencef("%s", line)
	}
}
