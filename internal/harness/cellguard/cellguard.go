// Package cellguard is the declared-site guard over profile-matrix CELL tables
// in the docs-site.
//
// # What this guard is for
//
// internal/template/profile_matrix.go decides, per agent and per profile, which
// {model, effort} pair an agent call receives. Several docs-site pages restate
// those cells as markdown tables. Nothing mechanically bound the two, so a
// change to the Go matrix left the tables asserting the old values and the hugo
// build stayed exit 0 the whole time. Both instances that produced this package
// were caught by a human comparing cells by eye:
//
//   - docs-site getting-started/faq.md described a stair the matrix does not
//     have. Seven of its twelve rows disagreed, in all four locales, and two of
//     them claimed `opus / max` — an effort the matrix uses in no cell at all.
//   - the advanced/no-haiku-3tier page was never bound to profile_matrix.go by
//     anything; `grep -rl no-haiku-3tier --include=*.go` returned no rows.
//
// # Why this guard is separate from the numeral sweep
//
// internal/web/docs_tab_contract_test.go guards a COUNT: its N1 layer sweeps for
// a numeral standing next to a noun and asserts the hit set equals an allowlist.
// That layer cannot decide anything here, because a cell is not a count — every
// cell in these tables is sanctioned, and the question is whether its VALUE
// agrees with the Go structure. The structural analogue is that file's N2 layer,
// which extracts a documented list and compares it against what the console
// actually renders. This package is that shape, applied to cells.
//
// # Decisions carried over from the guards that came before
//
//   - Rows are identified by CONTENT, never by line number. The same cell table
//     sits at line 27 in the ko page and line 36 in the other three, so a
//     line-anchored extractor would read the wrong region in three locales
//     (docs_tab_contract_test.go reached the same conclusion for its allowlist).
//   - The extracted row count is ASSERTED, not merely used. An extractor that
//     silently matches nothing agrees with everything; extracting nothing is not
//     agreement, it is a guard that read no table.
//   - Membership is compared as a SET, never as a number. rosterguard records
//     the tree state that makes this non-optional: two different populations
//     were both 12 with an intersection of 11, so a count comparison would have
//     passed vacuously on exactly the drift it existed to catch.
//   - A site that lists only part of the population declares that, and says
//     which members it omits and why. Asserting complete membership on a
//     deliberate subset would break a correct page.
package cellguard

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// TableKind names the population a cell table's rows are drawn from, and with
// it the source of truth its cells are compared against. A site declares its
// kind; a table whose first data row does not belong to that kind's population
// is not this site's table.
type TableKind string

const (
	// KindAgentMatrix marks a table whose rows are retained agent names and
	// whose cells restate template.DefaultProfileMatrix().
	KindAgentMatrix TableKind = "agent-matrix"

	// KindHarnessClass marks a table whose rows are /moai:harness purpose
	// classes and whose cells restate template.ResolveHarnessAgentModelEffort.
	// Its header carries an extra column naming the matrix row each class
	// borrows its effort from, so the profile columns are located by header
	// content rather than by position.
	KindHarnessClass TableKind = "harness-class"
)

// Profiles is the profile column order this guard asserts. It is the order the
// pages print, and the guard checks the header actually carries all three
// rather than assuming the positions.
var Profiles = []string{"high", "medium", "low"}

// Site is one registered cell-table location.
//
// A path may carry more than one site when it prints tables of different kinds;
// the kind is what separates them, because a page's two tables draw their rows
// from two different populations.
type Site struct {
	// ID is the stable identifier used in failure messages.
	ID string

	// Path is repo-root-relative, forward-slashed.
	Path string

	// Kind names the row population and the source of truth.
	Kind TableKind

	// WantRows is how many data rows this site must yield across every table
	// of its kind in the file. Asserting it is what stops an extractor that
	// reads nothing from passing.
	WantRows int

	// Complete declares that the extracted row set must EQUAL the kind's full
	// population. A site that legitimately prints a subset sets this false and
	// states SubsetReason.
	Complete bool

	// SubsetReason is required when Complete is false: which members the site
	// omits, and why that is correct rather than stale.
	SubsetReason string
}

// Cell is one {model, effort} pair as a page writes it.
type Cell struct {
	Model  string
	Effort string
}

func (c Cell) String() string { return c.Model + " / " + c.Effort }

// Row is one extracted table row: its key plus the cell under each profile.
type Row struct {
	Key   string
	Line  int
	Cells map[string]Cell
}

// tableRowRe matches a markdown table row and captures its inner span.
var tableRowRe = regexp.MustCompile(`^\s*\|(.+)\|\s*$`)

// separatorRe matches the `|---|---|` rule that follows a header row.
var separatorRe = regexp.MustCompile(`^\s*\|[\s:|-]+\|\s*$`)

// cellRe splits "opus / medium" into model and effort. The pages space the
// slash; the pattern tolerates it being absent.
var cellRe = regexp.MustCompile(`^([a-z0-9.-]+)\s*/\s*([a-z]+)$`)

// splitRow returns the trimmed cells of a markdown table row, with surrounding
// backticks and bold markers removed from each.
func splitRow(line string) ([]string, bool) {
	m := tableRowRe.FindStringSubmatch(line)
	if m == nil {
		return nil, false
	}
	parts := strings.Split(m[1], "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.Trim(strings.TrimSpace(p), "`* "))
	}
	return out, true
}

// profileColumns locates the profile columns in a header row by CONTENT. The
// ko pages annotate a column ("medium (기본)") and the harness table inserts an
// extra column before the profiles, so neither the exact text nor the position
// is dependable. It returns profile name -> column index, and false when the
// header does not carry all three.
func profileColumns(header []string) (map[string]int, bool) {
	cols := make(map[string]int, len(Profiles))
	for i, h := range header {
		lower := strings.ToLower(h)
		for _, p := range Profiles {
			if _, taken := cols[p]; taken {
				continue
			}
			if strings.Contains(lower, p) {
				cols[p] = i
				break
			}
		}
	}
	if len(cols) != len(Profiles) {
		return nil, false
	}
	return cols, true
}

// Extract reads every cell table in content whose rows belong to population,
// and returns them in file order.
//
// A table qualifies on two counts, both by content: its header carries all
// three profile columns, and its first data row's key is in population. The
// second condition is what separates two tables sharing one file without
// depending on the headings above them, which differ per locale.
func Extract(content string, population map[string]bool) ([]Row, error) {
	lines := strings.Split(content, "\n")
	var rows []Row

	for i := 0; i < len(lines); i++ {
		header, ok := splitRow(lines[i])
		if !ok || i+1 >= len(lines) || !separatorRe.MatchString(lines[i+1]) {
			continue
		}
		cols, ok := profileColumns(header)
		if !ok {
			continue
		}

		// Walk the body of this table.
		body := i + 2
		var table []Row
		for ; body < len(lines); body++ {
			cells, ok := splitRow(lines[body])
			if !ok {
				break
			}
			if len(cells) == 0 || !population[cells[0]] {
				continue
			}
			row := Row{Key: cells[0], Line: body + 1, Cells: map[string]Cell{}}
			for _, p := range Profiles {
				idx := cols[p]
				if idx >= len(cells) {
					return nil, fmt.Errorf("row %q at line %d has %d columns; profile %q sits at column %d", row.Key, row.Line, len(cells), p, idx)
				}
				m := cellRe.FindStringSubmatch(cells[idx])
				if m == nil {
					return nil, fmt.Errorf("row %q at line %d writes %q under %q, which is not a `model / effort` pair", row.Key, row.Line, cells[idx], p)
				}
				row.Cells[p] = Cell{Model: m[1], Effort: m[2]}
			}
			table = append(table, row)
		}
		rows = append(rows, table...)
		i = body - 1
	}

	seen := map[string]int{}
	for _, r := range rows {
		if prev, dup := seen[r.Key]; dup {
			return nil, fmt.Errorf("row %q appears at line %d and again at line %d; one key must own one row, or the guard asserts only the last one", r.Key, prev, r.Line)
		}
		seen[r.Key] = r.Line
	}
	return rows, nil
}

// SetDiff returns the members of want that got omits, and the members of got
// that want does not carry. Both are sorted so failure output is stable.
func SetDiff(got []Row, want []string) (missing, extra []string) {
	have := make(map[string]bool, len(got))
	for _, r := range got {
		have[r.Key] = true
	}
	wanted := make(map[string]bool, len(want))
	for _, w := range want {
		wanted[w] = true
		if !have[w] {
			missing = append(missing, w)
		}
	}
	for _, r := range got {
		if !wanted[r.Key] {
			extra = append(extra, r.Key)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}
