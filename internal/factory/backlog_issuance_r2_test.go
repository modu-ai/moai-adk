package factory

import (
	"testing"
	"time"
)

// card t1454 card-review r2 finding 8: the in-flight overlap comparison
// reads the card's explicitly recorded expected files (Issuance.Files)
// beside the body-derived paths — the recorded attribute is the operator's
// statement, and a card whose body names no path still carries input.
func TestIssuanceOverlapReadsStoredIssuanceFiles(t *testing.T) {
	rec := &BacklogRecord{
		Items: []BacklogItem{{
			ID:       "t2",
			Text:     "work the parser without naming any path",
			State:    BacklogStatePicked,
			Issuance: &BacklogIssuance{Files: []string{"internal/cli/todo.go"}},
		}},
		Runtime: TodoRuntime{Assignments: []TodoRuntimeAssignment{{CardID: "t2", OwnerLabel: "lane-1"}}},
	}
	p := BuildIssuancePresentationFloor("fix internal/cli/todo.go soon", rec, nil, nil, nil, IssuanceDisplayFloor)
	if len(p.Overlap.Items) == 0 || p.Overlap.Items[0].Path != "internal/cli/todo.go" {
		t.Fatalf("overlap = %+v, want the stored issuance file internal/cli/todo.go among the shared paths", p.Overlap)
	}
}

// card t1454 card-review r2 finding 10: a probe past the time bound lost
// the card's probe input, so even a computed `none` was decided on partial
// input — the honest verdict is `unmeasured (time bound)`, not a measured
// no.
func TestIssuanceProbeTimeoutIsUnmeasuredNotNone(t *testing.T) {
	rec := &BacklogRecord{
		Items: []BacklogItem{{
			ID:    "t3",
			Text:  "touches internal/web/render.go only",
			State: BacklogStatePicked,
		}},
		Runtime: TodoRuntime{Assignments: []TodoRuntimeAssignment{{CardID: "t3", OwnerLabel: "lane-2"}}},
	}
	probe := func(string, string) ([]string, bool) {
		time.Sleep(3 * time.Second) // > IssuanceProbeTimeBound (2s)
		return nil, false
	}
	p := BuildIssuancePresentationFloor("touches internal/graph/graph.go", rec, nil, probe, nil, IssuanceDisplayFloor)
	if p.Overlap.None {
		t.Fatalf("overlap = none, want unmeasured (time bound) — the timed-out probe lost the card's only input")
	}
	if p.Overlap.Unmeasured != "time bound" {
		t.Fatalf("overlap unmeasured = %q, want time bound", p.Overlap.Unmeasured)
	}
}
