// Package golden provides the cumulative golden snapshot index for M1-M7 milestones.
// It validates that the testdata/ directory contains the expected number of golden
// files (>= 100 per plan.md §8.3 DOD) and that MANIFEST.txt checksums are consistent.
//
// Usage:
//
//	go test ./internal/tui/golden/... -v                  # verify
//	UPDATE_GOLDEN=1 go test ./internal/tui/golden/...     # regenerate manifest
package golden
