//go:build cgo

package astx

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// enrichBase returns EnrichOptions pointing at the synthetic capability-map
// fixture under testdata/enrich (implementation-path = src/auth).
func enrichBase(t *testing.T, capMap string) EnrichOptions {
	t.Helper()
	return EnrichOptions{
		ProjectRoot:       filepath.Join("testdata", "enrich"),
		CapabilityMapPath: capMap,
		MaxFilesPerPath:   2000,
		PrimaryFilesN:     5,
		PrimarySymbolsN:   10,
	}
}

// TestEnrichRows_HeaderDrivenJoin is AC-NT-011: the join resolves columns by
// header name, so a permuted column order still pairs rows correctly.
func TestEnrichRows_HeaderDrivenJoin(t *testing.T) {
	res, err := EnrichRows(enrichBase(t, "capability-map-permuted.md"))
	if err != nil {
		t.Fatalf("EnrichRows error: %v", err)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(res.Rows))
	}
	row := res.Rows[0]
	// implementation-path was in column 1 but the join should have resolved it
	// by header name and walked src/auth.
	if row.SpecID != "SPEC-AUTH-001" {
		t.Errorf("SpecID = %q, want SPEC-AUTH-001", row.SpecID)
	}
	if !row.OnDiskVerified {
		t.Errorf("OnDiskVerified = false, want true (src/auth exists under testdata/enrich)")
	}
	if row.ExtractLanguage != "go" {
		t.Errorf("ExtractLanguage = %q, want go", row.ExtractLanguage)
	}
	if row.SymbolCount == 0 {
		t.Errorf("SymbolCount = 0, want >0 (handler.go declares symbols)")
	}
}

// TestEnrichRows_MissingPathOnDiskVerifiedFalse is the on-disk-verified false
// case: a capability row whose implementation-path does not exist carries
// OnDiskVerified=false and the run does not abort.
func TestEnrichRows_MissingPathOnDiskVerifiedFalse(t *testing.T) {
	res, err := EnrichRows(enrichBase(t, "capability-map.md"))
	if err != nil {
		t.Fatalf("EnrichRows error: %v", err)
	}
	bySpec := map[string]EnrichedRow{}
	for _, r := range res.Rows {
		bySpec[r.SpecID] = r
	}
	gone, ok := bySpec["SPEC-GONE-002"]
	if !ok {
		t.Fatalf("SPEC-GONE-002 row missing; rows=%v", res.Rows)
	}
	if gone.OnDiskVerified {
		t.Errorf("SPEC-GONE-002 OnDiskVerified=true, want false (path absent)")
	}
	auth := bySpec["SPEC-AUTH-001"]
	if !auth.OnDiskVerified {
		t.Errorf("SPEC-AUTH-001 OnDiskVerified=false, want true")
	}
}

// TestEnrichRows_FileCountCeilingTruncation is AC-NT-014: a low ceiling
// truncates the walk and marks the row truncated. The src/auth fixture has
// two Go files (handler.go, token.go); a ceiling of 1 parses only the first
// and sets truncated=true.
func TestEnrichRows_FileCountCeilingTruncation(t *testing.T) {
	opts := enrichBase(t, "capability-map-permuted.md")
	opts.MaxFilesPerPath = 1
	res, err := EnrichRows(opts)
	if err != nil {
		t.Fatalf("EnrichRows error: %v", err)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(res.Rows))
	}
	row := res.Rows[0]
	if !row.Truncated {
		t.Errorf("Truncated = false, want true (MaxFilesPerPath=1 < 2 files)")
	}
	// Only one file parsed → primary_files has at most 1 entry.
	if len(row.PrimaryFiles) > 1 {
		t.Errorf("PrimaryFiles = %d entries, want <=1 (ceiling=1)", len(row.PrimaryFiles))
	}
}

// TestCurrentProvenance_NonEmpty is REQ-NT-009: provenance carries a SHA +
// captured-at sourced from git (not wall-clock). In a git worktree HEAD is
// resolvable, so both fields must be non-empty.
func TestCurrentProvenance_NonEmpty(t *testing.T) {
	p := CurrentProvenance("")
	if p.ExtractCommitSHA == "" {
		t.Errorf("ExtractCommitSHA empty, want a SHA")
	}
	if p.CapturedAt == "" {
		t.Errorf("CapturedAt empty, want an ISO-8601 committer date")
	}
}

// TestMarshalCapabilitySymbolsJSON_StableSchema is AC-NT-010: the JSON
// envelope carries the four top-level fields and each row the design §5.2
// fields. It also asserts valid JSON structure (parseable by encoding/json).
func TestMarshalCapabilitySymbolsJSON_StableSchema(t *testing.T) {
	p := Provenance{ExtractCommitSHA: "abc123", CapturedAt: "2026-01-01T00:00:00+00:00"}
	rows := []EnrichedRow{
		{
			SpecID: "SPEC-X-001", Title: "X", ImplementationPath: "internal/x",
			OnDiskVerified: true, ExtractLanguage: "go",
			PrimaryFiles:   []string{"internal/x/a.go"},
			PrimarySymbols: []Symbol{{Name: "Foo", Kind: "type", File: "internal/x/a.go", Line: 1}},
			SymbolCount:    1, Supported: true,
		},
	}
	b := MarshalCapabilitySymbolsJSON(p, ".moai/project/navigator/capability-map.md", rows)
	// Must be valid JSON.
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("JSON output invalid: %v\n%s", err, b)
	}
	for _, k := range []string{"extracted_at", "extract_commit", "source_capability_map", "rows"} {
		if _, ok := doc[k]; !ok {
			t.Errorf("top-level field %q missing from JSON", k)
		}
	}
	rowList, _ := doc["rows"].([]any)
	if len(rowList) != 1 {
		t.Fatalf("rows len = %d, want 1", len(rowList))
	}
	r0, _ := rowList[0].(map[string]any)
	for _, k := range []string{"spec_id", "title", "implementation_path", "on_disk_verified",
		"extract_language", "primary_files", "primary_symbols", "symbol_count",
		"truncated", "supported"} {
		if _, ok := r0[k]; !ok {
			t.Errorf("row field %q missing from JSON", k)
		}
	}
}

// TestRenderMarkdown_ContainsHeader is AC-NT-010 dual-output: the markdown
// carries the provenance header and a table row per EnrichedRow.
func TestRenderMarkdown_ContainsHeader(t *testing.T) {
	p := Provenance{ExtractCommitSHA: "abc123", CapturedAt: "2026-01-01T00:00:00+00:00"}
	rows := []EnrichedRow{{SpecID: "SPEC-X-001", Title: "X", OnDiskVerified: true, SymbolCount: 2}}
	b := string(RenderMarkdown(p, "cap.md", rows))
	for _, want := range []string{"Capability Symbols (AST-derived)", "abc123", "cap.md", "SPEC-X-001", "✓"} {
		if !strings.Contains(b, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
}

// TestEnrichRows_Idempotent is AC-NT-012: two consecutive runs on the same
// HEAD produce byte-identical JSON output (no wall-clock timestamp).
func TestEnrichRows_Idempotent(t *testing.T) {
	opts := enrichBase(t, "capability-map.md")
	r1, _ := EnrichRows(opts)
	r2, _ := EnrichRows(opts)
	b1 := MarshalCapabilitySymbolsJSON(r1.Provenance, "cap.md", r1.Rows)
	b2 := MarshalCapabilitySymbolsJSON(r2.Provenance, "cap.md", r2.Rows)
	if !bytes.Equal(b1, b2) {
		t.Errorf("idempotence violated: two runs produced different JSON")
	}
}

// TestParseCapabilityMap_SeparatorAndMissingTable covers the header-driven
// parser directly: a separator row is skipped, and a file with no table
// yields zero rows.
func TestParseCapabilityMap_SeparatorAndMissingTable(t *testing.T) {
	rows := parseCapabilityMap([]byte("| a | b |\n|---|---|\n| 1 | 2 |\n"))
	if len(rows) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(rows))
	}
	if rows[0]["a"] != "1" || rows[0]["b"] != "2" {
		t.Errorf("parsed row = %v, want a=1 b=2", rows[0])
	}
	none := parseCapabilityMap([]byte("# just prose\n\nno table here\n"))
	if len(none) != 0 {
		t.Errorf("expected 0 rows for table-less input, got %d", len(none))
	}
}
