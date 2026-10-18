package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// SPEC-LANE-JOIN-RUN-SELECT-001 (card t1853) chooser seams: the pure row
// builder and renderer (REQ-LJR-003), and the text chooser driven with
// in-memory streams (no terminal, no stdin).

// chooserFacts is the two-run fixture of AC-LJR-003: a live-owner run with no
// cards and a dead-owner run holding seven non-terminal cards. The seeded
// times and ids contain no digit 7, so a 7 in a rendered cell can only be the
// count column.
func chooserFacts() ([]homestate.RunFacts, map[int]string) {
	facts := []homestate.RunFacts{
		{
			RunOwner: homestate.RunOwner{RunID: "runa-live", LeadPID: 4242, Classification: homestate.OwnerLive, CreatedAt: "2026-10-18T02:11:00Z"},
		},
		{
			RunOwner:     homestate.RunOwner{RunID: "runb-dead", LeadPID: 5151, Classification: homestate.OwnerDead, CreatedAt: "2026-10-18T01:31:00.123456789Z"},
			LastActivity: "2026-10-18T03:00:00.5Z",
			InProgress:   7,
		},
	}
	return facts, map[int]string{4242: "lane-x9"}
}

// AC-LJR-003 (render half): the rows are typed display cells, and the table
// line carries seven cells joined by laneColSep. Each cell is asserted by
// equality, so a substring inside another cell cannot satisfy the test and a
// dropped column changes the cell count.
func TestLaneJoinChooserRendersFacts(t *testing.T) {
	facts, names := chooserFacts()
	rows := buildLaneRows(facts, names)
	if len(rows) != 2 {
		t.Fatalf("buildLaneRows returned %d rows, want 2", len(rows))
	}
	wantRows := []laneRunRow{
		{RunID: "runa-live", Liveness: laneLivenessLive, LeaderName: "lane-x9", Start: "2026-10-18 11:11:00 KST", LastActivity: "-", InProgress: 0, CanCleanup: false},
		{RunID: "runb-dead", Liveness: laneLivenessDead, LeaderName: "-", Start: "2026-10-18 10:31:00 KST", LastActivity: "2026-10-18 12:00:00 KST", InProgress: 7, CanCleanup: false},
	}
	for i, want := range wantRows {
		if rows[i] != want {
			t.Errorf("row %d = %+v, want %+v", i+1, rows[i], want)
		}
	}

	lines := strings.Split(strings.TrimRight(renderLaneTable(rows), "\n"), "\n")
	if len(lines) != len(rows) {
		t.Fatalf("renderLaneTable emitted %d lines for %d rows:\n%s", len(lines), len(rows), renderLaneTable(rows))
	}
	wantCells := [][]string{
		{"1", "runa-live", laneLivenessLive, "lane-x9", "2026-10-18 11:11:00 KST", "-", "0"},
		{"2", "runb-dead", laneLivenessDead, "-", "2026-10-18 10:31:00 KST", "2026-10-18 12:00:00 KST", "7"},
	}
	for i, line := range lines {
		cells := strings.Split(line, laneColSep)
		if len(cells) != 7 {
			t.Fatalf("line %d has %d cells, want 7: %q", i+1, len(cells), line)
		}
		for j, want := range wantCells[i] {
			if cells[j] != want {
				t.Errorf("line %d cell %d = %q, want %q", i+1, j+1, cells[j], want)
			}
		}
	}
}

// The remaining display mappings: an indeterminate owner keeps its token and
// gloss, an unknown classification is shown as stored rather than hidden, an
// empty stored time renders as the empty cell, and an unparseable one is
// shown unchanged.
func TestLaneJoinChooserRowCellEdgeCases(t *testing.T) {
	facts := []homestate.RunFacts{
		{RunOwner: homestate.RunOwner{RunID: "ra", Classification: homestate.OwnerIndeterminate}},
		{RunOwner: homestate.RunOwner{RunID: "rb", Classification: "future-class", CreatedAt: "not-a-time"}, LastActivity: "also-not-a-time"},
	}
	rows := buildLaneRows(facts, nil)
	if len(rows) != 2 {
		t.Fatalf("buildLaneRows returned %d rows, want 2", len(rows))
	}
	if rows[0].Liveness != laneLivenessIndetermin || rows[0].Start != laneEmptyCell || rows[0].LeaderName != laneEmptyCell {
		t.Errorf("indeterminate row = %+v", rows[0])
	}
	if rows[1].Liveness != "future-class" || rows[1].Start != "not-a-time" || rows[1].LastActivity != "also-not-a-time" {
		t.Errorf("unknown-classification row = %+v", rows[1])
	}
}

func eligibilityFacts() []homestate.RunFacts {
	return []homestate.RunFacts{
		{RunOwner: homestate.RunOwner{RunID: "r1", Classification: homestate.OwnerLive}},
		{RunOwner: homestate.RunOwner{RunID: "r2", Classification: homestate.OwnerIndeterminate}},
		{RunOwner: homestate.RunOwner{RunID: "r3", Classification: homestate.OwnerDead}, InProgress: 2},
		{RunOwner: homestate.RunOwner{RunID: "r4", Classification: homestate.OwnerDead}, InProgress: 0},
	}
}

// AC-LJR-004 (eligibility half): cleanup is offered only for a dead owner
// with zero non-terminal cards, and never from a one-row table.
func TestLaneJoinChooserOptionEligibility(t *testing.T) {
	rows := buildLaneRows(eligibilityFacts(), nil)
	wantCleanup := []bool{false, false, false, true}
	for i, row := range rows {
		if row.CanCleanup != wantCleanup[i] {
			t.Errorf("row %d (%s, %s, %d cards) CanCleanup = %v, want %v", i+1, row.RunID, row.Liveness, row.InProgress, row.CanCleanup, wantCleanup[i])
		}
	}
	single := buildLaneRows(eligibilityFacts()[3:], nil)
	if len(single) != 1 || single[0].CanCleanup {
		t.Errorf("a single dead zero-card row = %+v, want CanCleanup false (a one-row table never offers cleanup)", single)
	}
}

func chooseFrom(t *testing.T, rows []laneRunRow, in io.Reader) (runChoice, string) {
	t.Helper()
	var out bytes.Buffer
	choice, err := textRunChooser{in: in, out: &out}.ChooseLaneRun(rows)
	if err != nil {
		t.Fatalf("ChooseLaneRun: %v", err)
	}
	return choice, out.String()
}

// AC-LJR-004 (parser half) plus the chooser-input debt of the plan audit
// (DEBT-1): valid entries, every cancel spelling, an ineligible cleanup, and
// out-of-range or unparseable entries that print the reason and ask again.
func TestRunChooserParsesInMemoryStreams(t *testing.T) {
	rows := buildLaneRows(eligibilityFacts(), nil)
	join := func(id string) runChoice { return runChoice{Kind: choiceJoin, RunID: id} }
	cancel := runChoice{Kind: choiceCancel}

	cases := []struct {
		name        string
		input       string
		want        runChoice
		wantInvalid int // laneMsgInvalid occurrences
		wantInelig  []int
	}{
		{name: "join-row-2", input: "2\n", want: join("r2")},
		{name: "join-with-padding", input: "  3 \n", want: join("r3")},
		{name: "cleanup-eligible", input: "r4\n", want: runChoice{Kind: choiceCleanupThenJoin, RunID: "r4"}},
		{name: "cleanup-uppercase", input: "R4\n", want: runChoice{Kind: choiceCleanupThenJoin, RunID: "r4"}},
		{name: "cleanup-ineligible-asks-again", input: "r2\n4\n", want: join("r4"), wantInelig: []int{2}},
		{name: "q-cancels", input: "q\n", want: cancel},
		{name: "empty-line-cancels", input: "\n", want: cancel},
		{name: "end-of-input-cancels", input: "", want: cancel},
		{name: "unterminated-line-is-not-a-choice", input: "2", want: cancel},
		// DEBT-1: out-of-range and unparseable entries.
		{name: "out-of-range-99", input: "99\n2\n", want: join("r2"), wantInvalid: 1},
		{name: "out-of-range-r99", input: "r99\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "garbage", input: "garbage\n3\n", want: join("r3"), wantInvalid: 1},
		{name: "zero", input: "0\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "negative", input: "-1\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "r-zero", input: "r0\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "bare-r", input: "r\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "signed", input: "+2\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "overflowing-number", input: "99999999999999999999\n1\n", want: join("r1"), wantInvalid: 1},
		{name: "two-bad-entries", input: "x\n99\n4\n", want: join("r4"), wantInvalid: 2},
		{name: "invalid-then-end-of-input-cancels", input: "99\n", want: cancel, wantInvalid: 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, out := chooseFrom(t, rows, strings.NewReader(c.input))
			if got != c.want {
				t.Errorf("choice = %+v, want %+v", got, c.want)
			}
			if n := strings.Count(out, laneMsgInvalid); n != c.wantInvalid {
				t.Errorf("invalid-entry reason printed %d time(s), want %d; output:\n%s", n, c.wantInvalid, out)
			}
			for _, n := range c.wantInelig {
				if !strings.Contains(out, fmt.Sprintf(laneMsgIneligibleFmt, n)) {
					t.Errorf("output lacks the ineligible-cleanup reason for r%d:\n%s", n, out)
				}
			}
		})
	}

	// One byte per Read call yields the same answer as a whole-line read.
	got, _ := chooseFrom(t, rows, iotest.OneByteReader(strings.NewReader("r4\n")))
	if want := (runChoice{Kind: choiceCleanupThenJoin, RunID: "r4"}); got != want {
		t.Errorf("one-byte reader choice = %+v, want %+v", got, want)
	}

	// The chooser reads up to the newline and no further: the launcher hands
	// the terminal to the child session next, and over-read bytes would be
	// the child's first keystrokes.
	rest := strings.NewReader("2\nTYPEAHEAD")
	if got, _ := chooseFrom(t, rows, rest); got != join("r2") {
		t.Errorf("choice = %+v, want join r2", got)
	}
	left, _ := io.ReadAll(rest)
	if string(left) != "TYPEAHEAD" {
		t.Errorf("source reader left %q unread, want %q (the chooser over-read)", left, "TYPEAHEAD")
	}
}

// The first prompt shows the header, the column labels, one line per run and
// the legend; the one-row variant names its own header and legend and never
// advertises cleanup.
func TestRunChooserPromptText(t *testing.T) {
	rows := buildLaneRows(eligibilityFacts(), nil)
	_, out := chooseFrom(t, rows, strings.NewReader("q\n"))
	for _, want := range []string{laneMsgHeaderMany, laneMsgLegendMany, strings.Join(laneColumnLabels(), laneColSep), "r4"} {
		if !strings.Contains(out, want) {
			t.Errorf("prompt lacks %q:\n%s", want, out)
		}
	}
	for _, line := range strings.Split(strings.TrimRight(renderLaneTable(rows), "\n"), "\n") {
		if !strings.Contains(out, line) {
			t.Errorf("prompt lacks the table line %q", line)
		}
	}

	single := buildLaneRows(eligibilityFacts()[:1], nil)
	_, out = chooseFrom(t, single, strings.NewReader("q\n"))
	if !strings.Contains(out, laneMsgHeaderOne) || !strings.Contains(out, laneMsgLegendOne) {
		t.Errorf("one-row prompt lacks its own header and legend:\n%s", out)
	}
	if strings.Contains(out, laneMsgLegendMany) || strings.Contains(out, laneMsgHeaderMany) {
		t.Errorf("one-row prompt carries the many-row text:\n%s", out)
	}
}

// The catalogue is native UTF-8 source text, never an escape sequence.
func TestRunChooserCatalogueHasNoUnicodeEscapes(t *testing.T) {
	for _, s := range []string{laneMsgHeaderMany, laneMsgHeaderOne, laneMsgLegendMany, laneMsgLegendOne,
		laneMsgIneligibleFmt, laneMsgInvalid, laneMsgCleanupFailFmt, laneMsgCancelled,
		laneLivenessLive, laneLivenessDead, laneLivenessIndetermin} {
		if strings.Contains(s, `\u`) {
			t.Errorf("catalogue string %q carries an escape sequence", s)
		}
	}
}

type refusingLaneWriter struct{}

func (refusingLaneWriter) Write([]byte) (int, error) { return 0, errors.New("write refused") }

// A prompt that cannot be written is an error, not a hang or a silent cancel.
func TestRunChooserSurfacesWriteFailure(t *testing.T) {
	rows := buildLaneRows(eligibilityFacts(), nil)
	_, err := textRunChooser{in: strings.NewReader("2\n"), out: refusingLaneWriter{}}.ChooseLaneRun(rows)
	if err == nil {
		t.Fatal("ChooseLaneRun with a failing writer returned nil error")
	}
}
