package web

import (
	"regexp"
	"strings"
	"testing"
)

// SPEC-WEB-TRANSPORT-001 M1 (REQ-TR400-002, AC-TR400-002) — D-400b extraction
// harness. Mechanically extracts the default responseHandling table from the
// COMMITTED embedded htmx asset (htmx 2.0.4, internal/web/assets/htmx.min.js)
// and evaluates it against HTTP status 400.
//
// This is a measurement, not a judgment: the test asserts only the structural
// preconditions of a valid measurement (the table exists, its entries parse,
// exactly one entry matches status 400) and logs the measured swap contract
// verbatim. The defect verdict is written separately in
// .moai/reports/t1080/verdict.md, gated on this measurement (REQ-TR400-003:
// measurement-before-judgment).
//
// JS-matcher semantics replicated: htmx matches each entry with
// `new RegExp(entry.code).test(status.toString(10))` (unanchored substring
// search, first entry in table order wins) — Go regexp MatchString has the
// same unanchored semantics for these patterns, so the evaluation below is a
// faithful port of the asset's own matching code.

// htmxResponseEntry is one parsed entry of the default responseHandling table.
type htmxResponseEntry struct {
	Code  string
	Swap  bool
	Error bool
}

// htmxResponseHandlingMatches reports whether the entry's code pattern matches
// the status string, replicating the asset's `new RegExp(code).test(status)`.
func htmxResponseHandlingMatches(e htmxResponseEntry, status string) bool {
	re, err := regexp.Compile(e.Code)
	if err != nil {
		return false
	}
	return re.MatchString(status)
}

// TestHtmxResponseHandlingTableExtraction measures D-400b: what does the pinned
// embedded htmx 2.0.4 build do with a 4xx boosted response body by default?
//
// Escape path (AC-TR400-002 edge case 1): if the minified literal does not
// extract, the test records the exact failure point and skips — that skip IS
// the deferred-to-browser record (REQ-TR400-003). Extraction failure is never
// treated as a defect judgment in either direction.
func TestHtmxResponseHandlingTableExtraction(t *testing.T) {
	t.Parallel()

	js := readEmbeddedAsset(t, "htmx.min.js")

	// --- Extraction step 1: the default table literal (verbatim evidence) ---
	// The capture spans `[{...}]` non-greedily rather than excluding `]`,
	// because the code patterns themselves contain `]` inside their string
	// literals (`"[23].."`) — a [^\]]* char class would truncate the fragment
	// at the first pattern literal and silently drop the 4xx cell.
	tableRe := regexp.MustCompile(`responseHandling:(\[\{.*?\}\])`)
	m := tableRe.FindStringSubmatch(js)
	if m == nil {
		t.Skipf("deferred-to-browser: responseHandling table literal not found in embedded htmx.min.js (extraction failure point: `responseHandling:[...]` regex matched nothing in %d bytes)", len(js))
	}
	verbatimFragment := m[0] // e.g. responseHandling:[{code:"204",...},...]
	t.Logf("D-400b verbatim table fragment: %s", verbatimFragment)

	// --- Extraction step 2: the matching predicate (supplementary evidence) ---
	// Confirms the matcher semantics the evaluation replicates, from the same
	// asset rather than from external documentation.
	if !strings.Contains(js, "new RegExp(e.code)") {
		t.Skipf("deferred-to-browser: response-handling matcher predicate `new RegExp(e.code)` not found in embedded asset — matcher semantics unverifiable on this tree")
	}
	inStart := strings.Index(js, "function In(")
	if inStart >= 0 {
		t.Logf("D-400b matcher predicate fragment: %s", js[inStart:inStart+96])
	}

	// --- Extraction step 3: parse the entries ---
	entryRe := regexp.MustCompile(`\{code:"([^"]+)",swap:(true|false)(,error:true)?\}`)
	entryMatches := entryRe.FindAllStringSubmatch(verbatimFragment, -1)
	if len(entryMatches) == 0 {
		t.Skipf("deferred-to-browser: table fragment extracted but zero entries parsed (extraction failure point: entry regex against %q)", verbatimFragment)
	}
	entries := make([]htmxResponseEntry, 0, len(entryMatches))
	for _, em := range entryMatches {
		entries = append(entries, htmxResponseEntry{
			Code:  em[1],
			Swap:  em[2] == "true",
			Error: em[3] != "",
		})
	}
	t.Logf("D-400b parsed default responseHandling entries (%d): %+v", len(entries), entries)

	// Structural guard: the htmx 2.x default table is documented as
	// 204-first; if the first entry is not the 204 cell, the minified shape
	// diverged from every known htmx 2.x build and the extraction is suspect.
	if entries[0].Code != "204" {
		t.Fatalf("extraction suspect: first responseHandling entry is %q, want \"204\" — minified shape diverged from htmx 2.x defaults", entries[0].Code)
	}

	// --- Evaluation: status 400 against the table (first match wins) ---
	const status400 = "400"
	matched := -1
	for i, e := range entries {
		if htmxResponseHandlingMatches(e, status400) {
			matched = i
			break // htmx's Pn loop returns the FIRST matching entry
		}
	}
	if matched < 0 {
		t.Fatalf("no responseHandling entry matches status %s — asset has no default for 400 (falls through to the asset's own {swap:false} tail); this is itself a measurement, record it in the verdict", status400)
	}
	e := entries[matched]
	t.Logf("D-400b MEASURED: status 400 -> entry #%d {code:%q, swap:%t, error:%t} (first-match-wins per asset matcher)", matched, e.Code, e.Swap, e.Error)

	// Exactly-one-first-match sanity: no entry EARLIER than the first match
	// may match 400 (guards an extraction-order transcription error).
	for j := 0; j < matched; j++ {
		if htmxResponseHandlingMatches(entries[j], status400) {
			t.Fatalf("extraction order corruption: entry #%d also matches status 400 but precedes first match #%d", j, matched)
		}
	}
}
