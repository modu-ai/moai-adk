package factory

import (
	"reflect"
	"strings"
	"testing"
)

// M1 RED tests (SPEC-TODO-CARD-ISSUANCE-001, AC-TCI-003 (a)-(h)). The
// issuance lookup is a read-only path SEPARATE from the duplicate
// classifier: it wants dropped and archived cards as notice targets, which
// ClassifyCardText deliberately refuses to see. Fixed inputs follow the
// acceptance criteria's named cards L1/L2/L3 and N1-N4.

func issuanceFixtureItems() ([]BacklogItem, []BacklogArchiveEntry) {
	items := []BacklogItem{
		{ID: "t900", Text: "align the CLI flag parser with the wizard defaults", State: BacklogStateQueued},
		{ID: "t901", Text: "[DROPPED — operator decision] align the CLI flag parser with the wizard defaults", State: BacklogStateDropped},
		{ID: "t902", Text: "ship the renderer refactor for the console screens", State: BacklogStatePicked},
		{ID: "t903", Text: "noise card about an unrelated database migration tool", State: BacklogStateQueued},
	}
	archived := []BacklogArchiveEntry{
		{Item: BacklogItem{ID: "t800", Text: "align the CLI flag parser with the wizard defaults for launch", State: BacklogStateQueued}},
	}
	return items, archived
}

// AC-TCI-003 (a): live, dropped AND archived cards are all notice targets.
func TestIssuanceNeighborsIncludeArchivedAndDropped(t *testing.T) {
	items, archived := issuanceFixtureItems()
	got := IssuanceNeighbors("align the CLI flag parser with the wizard defaults", items, archived)
	states := map[string]string{}
	for _, n := range got {
		states[n.ID] = n.State
	}
	if states["t900"] != "live" {
		t.Errorf("live neighbor t900 missing (states=%v)", states)
	}
	if states["t901"] != "dropped" {
		t.Errorf("dropped neighbor t901 missing (states=%v)", states)
	}
	if states["t800"] != "archived" {
		t.Errorf("archived neighbor t800 missing (states=%v)", states)
	}
}

// AC-TCI-003 (a): the dropped prefix is stripped and the reason kept.
func TestIssuanceNeighborsStripDropPrefix(t *testing.T) {
	items, archived := issuanceFixtureItems()
	got := IssuanceNeighbors("align the CLI flag parser with the wizard defaults", items, archived)
	for _, n := range got {
		if n.ID != "t901" {
			continue
		}
		if strings.HasPrefix(n.Text, "[DROPPED") {
			t.Errorf("dropped text still carries the prefix: %q", n.Text)
		}
		if !strings.Contains(n.Text, "align the CLI flag parser") {
			t.Errorf("stripped text lost the body: %q", n.Text)
		}
		if n.Reason != "operator decision" {
			t.Errorf("drop reason = %q, want %q", n.Reason, "operator decision")
		}
		return
	}
	t.Fatalf("dropped neighbor t901 not returned")
}

// AC-TCI-003 (a): the limit caps results and the display floor keeps noise
// out; a normalized-equal card is always shown regardless of the floor.
func TestIssuanceNeighborsLimitAndFloor(t *testing.T) {
	items := []BacklogItem{
		{ID: "t1", Text: "add retry budget to the exporter queue", State: BacklogStateQueued},
		{ID: "t2", Text: "add retry budget to the exporter queue with backoff", State: BacklogStateQueued},
		{ID: "t3", Text: "add retry budget handling to the exporter queue path", State: BacklogStateQueued},
		{ID: "t4", Text: "add retry budget metrics to the exporter queue worker", State: BacklogStateQueued},
		{ID: "t5", Text: "add retry budget tests to the exporter queue shim", State: BacklogStateQueued},
		{ID: "t6", Text: "totally different subject about window cleanup", State: BacklogStateQueued},
	}
	got := IssuanceNeighbors("add retry budget to the exporter queue", items, nil)
	if len(got) > IssuanceNeighborLimit {
		t.Errorf("neighbors = %d, want <= %d", len(got), IssuanceNeighborLimit)
	}
	for _, n := range got {
		if n.ID == "t6" {
			t.Errorf("below-floor noise card t6 leaked into the presentation")
		}
	}
	// An exact (normalized-equal) card is shown with Exact set even when the
	// candidate would otherwise be near the floor.
	exact := IssuanceNeighbors("ADD   RETRY BUDGET to the exporter queue", items, nil)
	sawExact := false
	for _, n := range exact {
		if n.ID == "t1" && n.Exact {
			sawExact = true
		}
	}
	if !sawExact {
		t.Errorf("normalized-equal card t1 not flagged exact (got %+v)", exact)
	}
}

// AC-TCI-003 (b): same-component open cards only — shared depth-2 key, open
// states, and cards with no paths never appear.
func TestIssuanceSameComponentOpenCards(t *testing.T) {
	items := []BacklogItem{
		{ID: "t10", Text: "harden the add path in internal/cli/todo.go against reentry", State: BacklogStateQueued},
		{ID: "t11", Text: "split the todo verb registration out of internal/cli/todo.go", State: BacklogStatePicked},
		{ID: "t12", Text: "[DROPPED — done elsewhere] old work touching internal/cli/todo.go", State: BacklogStateDropped},
		{ID: "t13", Text: "rewrite the web console screens without any path", State: BacklogStateQueued},
	}
	got := IssuanceSameComponent("make internal/cli/todo.go issuance aware", items)
	ids := map[string]bool{}
	for _, c := range got {
		ids[c.ID] = true
		if c.Key != "internal/cli" {
			t.Errorf("component key = %q, want internal/cli", c.Key)
		}
	}
	if !ids["t10"] || !ids["t11"] {
		t.Errorf("open same-component cards missing (got %v)", ids)
	}
	if ids["t12"] {
		t.Errorf("dropped card t12 must not be a component target (open only)")
	}
	if ids["t13"] {
		t.Errorf("path-less card t13 must not match a component key")
	}
}

// AC-TCI-003 (c): completed SPECs only — draft excluded — with the heuristic
// marker carried on the match.
func TestIssuanceCompletedSpecCoverage(t *testing.T) {
	specs := []IssuanceCompletedSpec{
		{ID: "SPEC-TODO-ANALYSIS-001", Status: "completed", Module: "internal/factory", Title: "mechanical text analyser behind todo add", Tags: "analysis, classification"},
		{ID: "SPEC-WEB-CONSOLE-001", Status: "draft", Module: "internal/web", Title: "mechanical text analyser console", Tags: "web"},
	}
	got := IssuanceCompletedSpecCoverage("teach the mechanical text analyser behind todo add a new measure", specs)
	if len(got) != 1 {
		t.Fatalf("matches = %d, want 1 (draft excluded)", len(got))
	}
	if got[0].ID != "SPEC-TODO-ANALYSIS-001" {
		t.Errorf("matched %q, want the completed SPEC", got[0].ID)
	}
	if got[0].Status != "completed" {
		t.Errorf("status = %q, want completed", got[0].Status)
	}
}

// AC-TCI-003 (d): a candidate with no expected files is `unmeasured`, never
// `none`; the same holds when no in-flight card carries input.
func TestIssuanceInFlightOverlapUnmeasured(t *testing.T) {
	got := IssuanceInFlightOverlap(nil, []IssuanceInFlightCard{
		{ID: "t20", Lane: "lane-1", Files: []string{"internal/cli/todo.go"}},
	})
	if got.Unmeasured == "" || got.None {
		t.Errorf("no-candidate-files: want unmeasured, got %+v", got)
	}
	got = IssuanceInFlightOverlap([]string{"internal/cli/todo.go"}, []IssuanceInFlightCard{
		{ID: "t20", Lane: "lane-1", Files: nil},
	})
	if got.Unmeasured == "" || got.None {
		t.Errorf("no-in-flight-input: want unmeasured, got %+v", got)
	}
}

// AC-TCI-003 (f): the positive case — one item per shared path, each naming
// the in-flight card id, lane and path, measure=file-overlap; the input-less
// in-flight card L3 is excluded from both comparison and report.
func TestIssuanceInFlightOverlapReportsSharedPath(t *testing.T) {
	candidate := []string{"internal/cli/todo.go", "internal/web/todo_view.go"}
	inFlight := []IssuanceInFlightCard{
		{ID: "t30", Lane: "lane-1", Files: []string{"internal/cli/todo.go"}},
		{ID: "t31", Lane: "lane-2", Files: []string{"internal/web/todo_view.go"}},
		{ID: "t32", Lane: "lane-3", Files: nil},
	}
	got := IssuanceInFlightOverlap(candidate, inFlight)
	if got.Unmeasured != "" || got.None {
		t.Fatalf("positive case mislabelled: %+v", got)
	}
	want := []IssuanceOverlapItem{
		{CardID: "t30", Lane: "lane-1", Path: "internal/cli/todo.go"},
		{CardID: "t31", Lane: "lane-2", Path: "internal/web/todo_view.go"},
	}
	if !reflect.DeepEqual(got.Items, want) {
		t.Errorf("items = %+v, want %+v", got.Items, want)
	}
	for _, it := range got.Items {
		for _, other := range inFlight {
			if other.ID == it.CardID && len(other.Files) == 0 {
				t.Errorf("input-less card %s leaked into the report", it.CardID)
			}
		}
	}
}

// AC-TCI-003 (g)(h): a measured no is `none`, not `unmeasured`; and the path
// comparison is FULL normalized paths — a same-basename different-directory
// path and a string-prefix sibling produce no overlap item.
func TestIssuanceInFlightOverlapNoneIsMeasured(t *testing.T) {
	// (g): candidate shares nothing with the input-carrying in-flight card.
	got := IssuanceInFlightOverlap([]string{"internal/graph/graph.go"}, []IssuanceInFlightCard{
		{ID: "t30", Lane: "lane-1", Files: []string{"internal/cli/todo.go"}},
	})
	if !got.None || got.Unmeasured != "" || len(got.Items) != 0 {
		t.Errorf("measured none: want None with no items, got %+v", got)
	}
	// (h): basename collision and prefix sibling are not overlaps.
	got = IssuanceInFlightOverlap([]string{"internal/web/todo.go", "internal/cli/todo.go.orig"}, []IssuanceInFlightCard{
		{ID: "t30", Lane: "lane-1", Files: []string{"internal/cli/todo.go"}},
	})
	if !got.None || len(got.Items) != 0 {
		t.Errorf("full-path equality: want none with no items, got %+v", got)
	}
}
