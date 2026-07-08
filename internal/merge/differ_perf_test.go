package merge

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

// reconstructFromGreedyEdits rebuilds b from a + greedy edit script.
// Works for diffLinesGreedy output (contiguous delete/insert ranges).
func reconstructFromGreedyEdits(a []string, edits []Edit) []string {
	var delStart, delEnd int = -1, -1
	var insTexts []string
	for _, e := range edits {
		if e.Op == OpDelete {
			if delStart == -1 {
				delStart = e.OldLine
			}
			delEnd = e.OldLine + 1
		} else if e.Op == OpInsert {
			insTexts = append(insTexts, e.NewText)
		}
	}
	if delStart == -1 && len(insTexts) == 0 {
		return a // no edits
	}
	if delStart == -1 {
		delStart = 0
		delEnd = 0
	}
	result := make([]string, 0, len(a)-delEnd+delStart+len(insTexts))
	result = append(result, a[:delStart]...)
	result = append(result, insTexts...)
	result = append(result, a[delEnd:]...)
	return result
}

// TestDiffLinesGreedyProperty (AC-PERF-005b) verifies that applying the greedy
// fallback's edit script to `a` produces `b` exactly. Tests multiple input shapes.
func TestDiffLinesGreedyProperty(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a, b []string
	}{
		{"identical_large", makeLines(5000, "line"), makeLines(5000, "line")},
		{"completely_different", makeLines(5000, "old"), makeLines(5000, "new")},
		{"insert_at_start", makeLines(2500, "x"), append([]string{"extra"}, makeLines(2500, "x")...)},
		{"delete_at_end", append(makeLines(2500, "x"), "extra"), makeLines(2500, "x")},
		{"common_prefix_suffix", makeGreedyMid(5000), makeGreedyMid2(5000)},
		{"empty_a", []string{}, makeLines(3000, "b")},
		{"empty_b", makeLines(3000, "a"), []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			edits := DiffLines(tc.a, tc.b)
			// Verify threshold was exceeded (greedy path activated)
			if len(tc.a) <= diffLinesThreshold && len(tc.b) <= diffLinesThreshold {
				// Small inputs use DP — verify with existing test semantics
				return
			}
			result := reconstructFromGreedyEdits(tc.a, edits)
			if !slices.Equal(result, tc.b) {
				t.Errorf("greedy edits do not reconstruct b\n  result len=%d, b len=%d", len(result), len(tc.b))
			}
		})
	}
}

// TestDiffLinesThresholdPreservation (AC-PERF-005c) verifies that inputs at or
// below the threshold produce the same results as before the guard was added.
func TestDiffLinesThresholdPreservation(t *testing.T) {
	t.Parallel()
	// Small inputs should use DP path and produce identical results
	a := []string{"a", "b", "c", "d", "e"}
	b := []string{"a", "x", "c", "y", "e"}
	edits := DiffLines(a, b)
	// Verify there are exactly 2 deletes and 2 inserts (minimal edit script)
	var deletes, inserts int
	for _, e := range edits {
		if e.Op == OpDelete {
			deletes++
		} else if e.Op == OpInsert {
			inserts++
		}
	}
	if deletes != 2 || inserts != 2 {
		t.Errorf("expected 2 deletes + 2 inserts for small input, got %d deletes + %d inserts", deletes, inserts)
	}
}

// BenchmarkDiffLines5000 (AC-PERF-005a) measures allocated bytes for 5,000-line
// inputs. The greedy path should allocate ~10x less than the full DP path.
func BenchmarkDiffLines5000(b *testing.B) {
	aLines := makeLines(5000, "original")
	blines := makeLines(5000, "modified")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DiffLines(aLines, blines)
	}
}

func makeLines(n int, prefix string) []string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("%s line %d", prefix, i)
	}
	return lines
}

func makeGreedyMid(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		if i > 100 && i < n-100 {
			lines[i] = fmt.Sprintf("shared %d", i)
		} else {
			lines[i] = fmt.Sprintf("a %d", i)
		}
	}
	return lines
}

func makeGreedyMid2(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		if i > 100 && i < n-100 {
			lines[i] = fmt.Sprintf("shared %d", i)
		} else {
			lines[i] = fmt.Sprintf("b %d", i)
		}
	}
	return lines
}

// RandomInput generates a pseudo-random large input pair for property testing.
func randomLargeInput(seed int64, n int) []string {
	r := rand.New(rand.NewSource(seed))
	lines := make([]string, n)
	for i := range lines {
		lines[i] = strings.Repeat("x", r.Intn(20)+1) + fmt.Sprintf("%d", i)
	}
	return lines
}
