package merge

import (
	"slices"
	"testing"
)

// TestDiffLines_IdenticalReturnsNil pins the identical-input contract: both the
// DP path and the greedy path return a nil slice (not an empty one) when the
// inputs are equal, on either side of diffLinesThreshold.
func TestDiffLines_IdenticalReturnsNil(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a, b []string
	}{
		{"nil_nil", nil, nil},
		{"empty_empty", []string{}, []string{}},
		{"three_lines", []string{"line1", "line2", "line3"}, []string{"line1", "line2", "line3"}},
		{"threshold_2000", makeLines(diffLinesThreshold, "line"), makeLines(diffLinesThreshold, "line")},
		{"over_threshold_2001", makeLines(diffLinesThreshold+1, "line"), makeLines(diffLinesThreshold+1, "line")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			edits := DiffLines(tc.a, tc.b)
			if edits != nil {
				t.Errorf("expected nil edits for identical input, got non-nil slice (len=%d)", len(edits))
			}
		})
	}
}

// TestDiffLines_IdenticalAllocatesNothing verifies that identical inputs at the
// threshold do not build the O(m*n) LCS table. The one-line-difference control
// proves the measurement can observe allocations when the DP path does run.
//
// This test must NOT call t.Parallel(): testing.AllocsPerRun reads process-wide
// allocation counters, so a concurrently running test would inflate the count.
func TestDiffLines_IdenticalAllocatesNothing(t *testing.T) {
	a := makeLines(diffLinesThreshold, "line")
	same := makeLines(diffLinesThreshold, "line")
	differ := makeLines(diffLinesThreshold, "line")
	differ[diffLinesThreshold/2] = "changed line"

	identical := testing.AllocsPerRun(5, func() {
		_ = DiffLines(a, same)
	})
	if identical != 0 {
		t.Errorf("identical %d-line inputs: expected 0 allocs per run, got %v", diffLinesThreshold, identical)
	}

	control := testing.AllocsPerRun(5, func() {
		_ = DiffLines(a, differ)
	})
	if control <= 0 {
		t.Errorf("control (%d lines, one line differs): expected allocs > 0, got %v", diffLinesThreshold, control)
	}
}

// TestDiffLines_SingleEditAtThresholdBoundary verifies that non-identical inputs
// still produce the expected single-edit script on both sides of the threshold:
// the largest input is diffLinesThreshold lines (DP path) or one more (greedy path).
func TestDiffLines_SingleEditAtThresholdBoundary(t *testing.T) {
	t.Parallel()
	const at = 1000
	for _, size := range []int{diffLinesThreshold, diffLinesThreshold + 1} {
		changeA := makeLines(size, "line")
		changeB := makeLines(size, "line")
		changeB[at] = "changed line"

		insertA := makeLines(size-1, "line")
		insertB := slices.Insert(makeLines(size-1, "line"), at, "inserted line")

		deleteA := makeLines(size, "line")
		deleteB := slices.Delete(makeLines(size, "line"), at, at+1)

		cases := []struct {
			name       string
			a, b       []string
			wantOps    []EditOp
			wantOld    int
			wantNew    int
			wantInsert string
		}{
			{"one_line_change", changeA, changeB, []EditOp{OpDelete, OpInsert}, at, at, "changed line"},
			{"one_insertion", insertA, insertB, []EditOp{OpInsert}, -1, at, "inserted line"},
			{"one_deletion", deleteA, deleteB, []EditOp{OpDelete}, at, -1, ""},
		}
		for _, tc := range cases {
			t.Run(tc.name+"_"+sizeLabel(size), func(t *testing.T) {
				t.Parallel()
				edits := DiffLines(tc.a, tc.b)
				if edits == nil {
					t.Fatalf("expected non-nil edits for non-identical input")
				}
				gotOps := make([]EditOp, 0, len(edits))
				for _, e := range edits {
					gotOps = append(gotOps, e.Op)
					switch e.Op {
					case OpDelete:
						if e.OldLine != tc.wantOld {
							t.Errorf("delete OldLine = %d, want %d", e.OldLine, tc.wantOld)
						}
					case OpInsert:
						if e.NewLine != tc.wantNew {
							t.Errorf("insert NewLine = %d, want %d", e.NewLine, tc.wantNew)
						}
						if e.NewText != tc.wantInsert {
							t.Errorf("insert NewText = %q, want %q", e.NewText, tc.wantInsert)
						}
					}
				}
				if !slices.Equal(gotOps, tc.wantOps) {
					t.Errorf("edit ops = %v, want %v", gotOps, tc.wantOps)
				}
				// The op list, index, and text assertions above determine each
				// single-edit script exactly, so no reconstruction check is needed.
				// reconstructFromGreedyEdits is deliberately not used: when a script
				// has no deletes it inserts at index 0 regardless of NewLine, so a
				// single mid-file insertion would reconstruct out of order.
			})
		}
	}
}

func sizeLabel(size int) string {
	if size > diffLinesThreshold {
		return "greedy_2001"
	}
	return "dp_2000"
}
