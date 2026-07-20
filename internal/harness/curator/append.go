package curator

import (
	"errors"
	"fmt"
	"strings"
)

// AppendLearnedLocal appends a single learned entry to the
// MOAI:LEARNED-WORKFLOW-LOCAL section (BlockTypeLearnedLocal) in the file at
// path. The entry is inserted immediately before the end marker; existing
// entries' bytes are preserved unchanged (append-only contract, REQ-HEV2-012).
//
// The Tier-3 LOCAL surface is the permanent-record layer, DISTINCT from the
// digest-layer CRUD (AddBullet/UpdateBullet/DeleteBullet): this writer appends
// only — it provides no update or delete path for existing entries. Editing an
// existing LOCAL entry is a manual, human-governed operation.
//
// Deduplication (REQ-HEV2-014): when an existing entry in the LOCAL section
// already carries the same LedgerKey, the writer returns ErrDuplicateAppend and
// does NOT touch the file. A provisional entry (empty LedgerKey) bypasses the
// dedup guard — there is no key to dedup against (evidence-or-null,
// REQ-HEV2-024).
//
// Returns ErrBlockNotFound when the file does not contain a LOCAL marker block.
//
// @MX:NOTE: [AUTO] append-only contract — Tier-3 LOCAL surface appends only,
// no update/delete; dedup on ledger_key (Req-HEV2-012..014). Distinct from
// digest-layer CRUD in crud.go.
// @MX:REASON: the layer separation (permanent-record vs digest) is the
// load-bearing design decision; a future caller that adds an update/delete path
// here would violate the SPEC's Tier-3 contract.
func AppendLearnedLocal(path string, entry Bullet) error {
	if path == "" {
		return errors.New("curator: empty path")
	}
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	endIdx := findEndMarkerLine(lines, BlockTypeLearnedLocal)
	if endIdx < 0 {
		return fmt.Errorf("curator: %w (BlockTypeLearnedLocal)", ErrBlockNotFound)
	}

	// Dedup guard (REQ-HEV2-014): scan the LOCAL block body (between the start
	// and end markers) for an existing entry with the same ledger_key. The
	// scan is scoped to THIS section — a ledger_key that appears only in the
	// digest (LEARNED-WORKFLOW) block does NOT trigger dedup here, because the
	// two surfaces are distinct layers.
	if entry.LedgerKey != "" {
		startIdx := findStartMarkerLineBefore(lines, BlockTypeLearnedLocal, endIdx)
		marker := keyMarker(entry.LedgerKey)
		// startIdx+1 skips the start marker line itself; endIdx is exclusive
		// (the end marker line is not a body line).
		for _, line := range lines[startIdx+1 : endIdx] {
			if strings.Contains(line, marker) {
				return fmt.Errorf("curator: %w: %s", ErrDuplicateAppend, entry.LedgerKey)
			}
		}
	}

	// Append-only insert (REQ-HEV2-012): insert the new entry immediately
	// before the end marker without modifying any existing entry's bytes.
	newLine := bulletLine(entry)
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:endIdx]...)
	result = append(result, newLine)
	result = append(result, lines[endIdx:]...)
	return writeLines(path, result)
}

// findStartMarkerLineBefore returns the index of the start marker line for the
// given blockType, searching backwards from endIdx-1. It scopes the body scan
// to the block bounded by the nearest preceding start marker. Returns 0 when
// no start marker is found (defensive: treats the whole prefix as the body so
// the dedup scan still runs rather than silently skipping).
func findStartMarkerLineBefore(lines []string, blockType BlockType, endIdx int) int {
	spec, ok := markerRegistry[blockType]
	if !ok {
		return 0
	}
	for i := endIdx - 1; i >= 0; i-- {
		if strings.Contains(lines[i], spec.startPrefix) {
			return i
		}
	}
	return 0
}
