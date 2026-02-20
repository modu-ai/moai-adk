package merge

import (
	"fmt"
	"strings"
)

// EditOp represents a single edit operation in a diff.
type EditOp int

const (
	// OpEqual means the line is unchanged.
	OpEqual EditOp = iota
	// OpInsert means a line was added.
	OpInsert
	// OpDelete means a line was removed.
	OpDelete
)

// Edit represents a single line-level edit operation.
type Edit struct {
	// Op is the edit operation type.
	Op EditOp
	// OldLine is the 0-based index in the original (a) slice. -1 for inserts.
	OldLine int
	// NewLine is the 0-based index in the modified (b) slice. -1 for deletes.
	NewLine int
	// NewText holds the text for insert operations.
	NewText string
}

// DiffLines computes the edit script between two slices of lines using a
// simple LCS-based diff algorithm. It returns a minimal list of insert and
// delete operations needed to transform a into b.
func DiffLines(a, b []string) []Edit {
	// Compute the longest common subsequence table.
	m := len(a)
	n := len(b)

	// dp[i][j] = length of LCS of a[:i] and b[:j]
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to produce the edit script.
	var edits []Edit
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			// Equal - no edit needed, just move diagonally.
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			// Insert from b.
			edits = append(edits, Edit{
				Op:      OpInsert,
				OldLine: -1,
				NewLine: j - 1,
				NewText: b[j-1],
			})
			j--
		} else if i > 0 {
			// Delete from a.
			edits = append(edits, Edit{
				Op:      OpDelete,
				OldLine: i - 1,
				NewLine: -1,
			})
			i--
		}
	}

	// Reverse to get forward order.
	for left, right := 0, len(edits)-1; left < right; left, right = left+1, right-1 {
		edits[left], edits[right] = edits[right], edits[left]
	}

	return edits
}

// UnifiedDiff produces a unified diff string comparing base and current content.
// Returns an empty string if the files are identical.
func UnifiedDiff(filename string, base, current []byte) string {
	aLines := splitLines(string(base))
	bLines := splitLines(string(current))

	edits := DiffLines(aLines, bLines)
	if len(edits) == 0 {
		return ""
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "--- a/%s\n", filename)
	fmt.Fprintf(&sb, "+++ b/%s\n", filename)

	// Build a full annotated sequence for generating hunks.
	type annotatedLine struct {
		op   EditOp
		text string
		aIdx int // 0-based index in a, -1 if insert
		bIdx int // 0-based index in b, -1 if delete
	}

	// Reconstruct the full annotated sequence from edits.
	var annotated []annotatedLine
	ei := 0 // index into edits
	ai := 0 // current position in a
	bi := 0 // current position in b

	for ai < len(aLines) || bi < len(bLines) {
		if ei < len(edits) {
			e := edits[ei]
			switch e.Op {
			case OpDelete:
				if e.OldLine == ai {
					annotated = append(annotated, annotatedLine{op: OpDelete, text: aLines[ai], aIdx: ai, bIdx: -1})
					ai++
					ei++
					continue
				}
			case OpInsert:
				if e.NewLine == bi {
					annotated = append(annotated, annotatedLine{op: OpInsert, text: bLines[bi], aIdx: -1, bIdx: bi})
					bi++
					ei++
					continue
				}
			}
		}

		// Equal line.
		if ai < len(aLines) && bi < len(bLines) && aLines[ai] == bLines[bi] {
			annotated = append(annotated, annotatedLine{op: OpEqual, text: aLines[ai], aIdx: ai, bIdx: bi})
			ai++
			bi++
		} else if ai < len(aLines) {
			// Shouldn't happen with correct diff, but handle gracefully.
			annotated = append(annotated, annotatedLine{op: OpDelete, text: aLines[ai], aIdx: ai, bIdx: -1})
			ai++
		} else if bi < len(bLines) {
			annotated = append(annotated, annotatedLine{op: OpInsert, text: bLines[bi], aIdx: -1, bIdx: bi})
			bi++
		}
	}

	// Generate hunks with context (3 lines).
	const contextLines = 3

	// Find ranges of changed lines and their surrounding context.
	type hunkRange struct {
		start, end int // indices into annotated
	}

	var hunks []hunkRange
	i := 0
	for i < len(annotated) {
		if annotated[i].op != OpEqual {
			start := max(i-contextLines, 0)
			// Find the end of this change group.
			end := i
			for end < len(annotated) {
				if annotated[end].op != OpEqual {
					end++
				} else {
					// Check if next change is within context distance.
					nextChange := -1
					for k := end; k < len(annotated) && k <= end+2*contextLines; k++ {
						if annotated[k].op != OpEqual {
							nextChange = k
							break
						}
					}
					if nextChange >= 0 {
						end = nextChange + 1
					} else {
						break
					}
				}
			}
			endCtx := min(end+contextLines, len(annotated))
			hunks = append(hunks, hunkRange{start: start, end: endCtx})
			i = end
		} else {
			i++
		}
	}

	// Merge overlapping hunks.
	var merged []hunkRange
	for _, h := range hunks {
		if len(merged) > 0 && h.start <= merged[len(merged)-1].end {
			merged[len(merged)-1].end = h.end
		} else {
			merged = append(merged, h)
		}
	}

	for _, h := range merged {
		// Compute hunk header line numbers.
		aStart := 0
		aCount := 0
		bStart := 0
		bCount := 0
		first := true

		for idx := h.start; idx < h.end; idx++ {
			al := annotated[idx]
			switch al.op {
			case OpEqual:
				if first {
					aStart = al.aIdx + 1
					bStart = al.bIdx + 1
					first = false
				}
				aCount++
				bCount++
			case OpDelete:
				if first {
					aStart = al.aIdx + 1
					bStart = al.bIdx + 1
					if bStart < 1 {
						// Compute from context.
						bStart = aStart
					}
					first = false
				}
				aCount++
			case OpInsert:
				if first {
					aStart = max(al.aIdx+1, 1)
					bStart = al.bIdx + 1
					first = false
				}
				bCount++
			}
		}

		if first {
			// No lines in hunk (shouldn't happen).
			continue
		}

		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", aStart, aCount, bStart, bCount)

		for idx := h.start; idx < h.end; idx++ {
			al := annotated[idx]
			switch al.op {
			case OpEqual:
				sb.WriteString(" " + al.text + "\n")
			case OpDelete:
				sb.WriteString("-" + al.text + "\n")
			case OpInsert:
				sb.WriteString("+" + al.text + "\n")
			}
		}
	}

	return sb.String()
}

// splitLines splits a string into lines, removing trailing empty lines
// caused by a final newline.
func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}
