// Package epic implements the disk-grounded epic progress producer for
// `moai epic status <prefix>` (SPEC-EPIC-STATUS-001).
//
// The producer composes the existing `internal/spec.ListDocs` +
// `internal/spec.Audit` read path (the same pair `internal/web/board.go`
// composes) and derives an epic's milestone progress map from on-disk signals:
// SPEC dir IDs, title `(TOKEN Mx)` markers, and an optional design-report
// canonical milestone list. It is observation-only — it MUST NOT mutate any
// file (REQ-ES-002) and MUST NOT persist any epic store (REQ-ES-013).
package epic

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// Options configures an epic producer invocation. It is filled by the CLI
// flag layer and consumed by DiscoverEpic + BuildEpicStatus.
type Options struct {
	// BaseDir is the project root containing .moai/specs/. Empty defaults to "."
	// (mirroring spec.ListDocs / spec.Audit).
	BaseDir string
	// Marker overrides the inferred epic token (design.md §3 Stage 2). Empty →
	// the token is inferred as the mode across the matched set.
	Marker string
	// DesignReport bypasses auto-discovery and points at a specific report HTML
	// file. Empty → auto-discovery via the naming rule (design.md §4).
	DesignReport string
}

// EpicCandidates is the Stage-1 prefix-filter result: the matched set (whose
// SPEC IDs start with `SPEC-<prefix>-`) plus the unmatched remainder (kept for
// observability; the producer does not surface it in --json).
type EpicCandidates struct {
	Matched   []spec.DocRecord
	Unmatched []spec.DocRecord
	Prefix    string
}

// DiscoverEpic implements Stage 1 of the epic-discovery chain (design.md §3):
// it calls spec.ListDocs once and filters the returned records to those whose
// frontmatter ID (or directory-name fallback) starts with `SPEC-<prefix>-`.
//
// The prefix is matched case-sensitively (SPEC IDs are uppercase). An empty
// match set is NOT an error (AC-ES-003b); the producer returns an empty
// EpicCandidates and the caller renders the empty-epic shape.
func DiscoverEpic(prefix string, opts Options) (*EpicCandidates, error) {
	records, err := spec.ListDocs(opts.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("epic list specs: %w", err)
	}
	cand := &EpicCandidates{Prefix: prefix}
	needle := "SPEC-" + prefix + "-"
	for _, rec := range records {
		id := recordSpecID(rec)
		if strings.HasPrefix(id, needle) {
			cand.Matched = append(cand.Matched, rec)
		} else {
			cand.Unmatched = append(cand.Unmatched, rec)
		}
	}
	return cand, nil
}

// recordSpecID returns the SPEC identifier for a DocRecord, falling back to the
// directory name when the frontmatter ID is absent (mirrors boardSpecID in
// internal/web/board.go).
func recordSpecID(rec spec.DocRecord) string {
	if rec.Frontmatter.ID != "" {
		return rec.Frontmatter.ID
	}
	if rec.Path != "" {
		// rec.Path is .../SPEC-<id>/spec.md → parent dir name.
		return parentDirName(rec.Path)
	}
	return ""
}

// parentDirName returns the directory name holding path (the SPEC dir name).
func parentDirName(path string) string {
	// filepath.Base(filepath.Dir(path)) without importing filepath here to keep
	// the test surface lean; both helpers live in the stdlib.
	idx := strings.LastIndexByte(path, '/')
	if idx < 0 {
		return path
	}
	parent := path[:idx]
	if j := strings.LastIndexByte(parent, '/'); j >= 0 {
		return parent[j+1:]
	}
	return parent
}

// markerPattern is the title-regex (design.md §3 Stage 2): captures an
// uppercase TOKEN + an M-number. The first capture is the token, the second is
// the M-number digits.
var markerPattern = regexp.MustCompile(`\(([A-Z][A-Z0-9-]*)\s+M(\d+)\)`)

// ExtractMx implements Stage 2 of the epic-discovery chain. It scans each
// record's title for the `(TOKEN M<N>)` marker pattern and builds a Mx→SPEC-ID
// map. Only markers whose TOKEN matches the `token` argument are counted.
//
// Returns:
//   - mxMap:     the Mx→owning-SPEC-ID map (first marker wins; later markers
//     on the same SPEC are recorded in extras — edge case E1).
//   - untracked: SPEC IDs whose title has no matching marker (REQ-ES-004).
//   - extras:    Mx identifiers that appear as a second marker on a SPEC that
//     already contributed its first marker to mxMap (edge case E1).
func ExtractMx(records []spec.DocRecord, token string) (mxMap map[string]string, untracked, extras []string, err error) {
	mxMap = make(map[string]string)
	for _, rec := range records {
		id := recordSpecID(rec)
		title := rec.Frontmatter.Title
		matches := markerPattern.FindAllStringSubmatch(title, -1)
		// Keep only matches whose TOKEN equals the configured token.
		var matching [][]string
		for _, m := range matches {
			if len(m) >= 3 && m[1] == token {
				matching = append(matching, m)
			}
		}
		if len(matching) == 0 {
			untracked = append(untracked, id)
			continue
		}
		// First marker wins.
		first := matching[0]
		mxID := "M" + first[2]
		// Only record the first owner of each Mx; subsequent SPECs claiming the
		// same Mx fall through to extras (preserves the "first wins" invariant
		// across SPEC boundaries too, not just within one title).
		if _, exists := mxMap[mxID]; !exists {
			mxMap[mxID] = id
		} else {
			extras = append(extras, mxID)
		}
		// Additional markers on this SPEC → extras.
		for _, m := range matching[1:] {
			extras = append(extras, "M"+m[2])
		}
	}
	sort.Strings(untracked)
	sort.Strings(extras)
	return mxMap, untracked, extras, nil
}

// InferToken returns the epic token to use given the matched record set. When
// `override` is non-empty it wins outright (the --marker flag). Otherwise the
// token is the mode across all `(TOKEN Mx)` markers found in the matched
// titles, with ties broken by first-seen (design.md §3 Stage 2).
func InferToken(records []spec.DocRecord, override string) string {
	if override != "" {
		return override
	}
	counts := make(map[string]int)
	first := make(map[string]int)
	order := 0
	for _, rec := range records {
		matches := markerPattern.FindAllStringSubmatch(rec.Frontmatter.Title, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				tok := m[1]
				if _, seen := first[tok]; !seen {
					first[tok] = order
					order++
				}
				counts[tok]++
			}
		}
	}
	if len(counts) == 0 {
		return ""
	}
	best := ""
	bestCount := -1
	bestFirst := 0
	for tok, n := range counts {
		// pick highest count; on tie, the earliest first-seen wins
		if n > bestCount || (n == bestCount && first[tok] < bestFirst) {
			best = tok
			bestCount = n
			bestFirst = first[tok]
		}
	}
	return best
}
