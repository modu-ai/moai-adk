package epic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// writeSpecFixture creates a SPEC directory with a spec.md carrying the given
// frontmatter title + status. It is the minimal fixture builder for epic
// producer tests — mirrors the pattern in internal/spec/audit_test.go.
func writeSpecFixture(t *testing.T, root, id, title, status string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	content := "---\n" +
		"id: " + id + "\n" +
		"title: " + title + "\n" +
		"version: \"0.1.0\"\n" +
		"status: " + status + "\n" +
		"created: 2026-08-11\n" +
		"updated: 2026-08-11\n" +
		"author: test\n" +
		"priority: P1\n" +
		"phase: \"v3.2.0\"\n" +
		"module: internal/epic\n" +
		"lifecycle: spec-anchored\n" +
		"tags: test\n" +
		"---\n\n# " + id + "\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write spec.md %s: %v", dir, err)
	}
}

// TestDiscoverEpic_PrefixGlob verifies REQ-ES-003: the prefix matches SPEC IDs
// starting with `SPEC-<prefix>-` and excludes non-matching siblings.
func TestDiscoverEpic_PrefixGlob(t *testing.T) {
	root := t.TempDir()
	writeSpecFixture(t, root, "SPEC-NAVIGATOR-SYNC-001", "Navigator Sync (BAS M0) — foo", "completed")
	writeSpecFixture(t, root, "SPEC-NAVIGATOR-SYNC-002", "Navigator Sync (BAS M4) — bar", "in-progress")
	writeSpecFixture(t, root, "SPEC-NAVIGATOR-SYNC-003", "Navigator Sync (BAS M1) — baz", "draft")
	writeSpecFixture(t, root, "SPEC-AUTH-001", "Auth (unrelated)", "completed")

	got, err := DiscoverEpic("NAVIGATOR-SYNC", Options{BaseDir: root})
	if err != nil {
		t.Fatalf("DiscoverEpic: %v", err)
	}
	if len(got.Matched) != 3 {
		t.Fatalf("expected 3 matched, got %d", len(got.Matched))
	}
	for _, rec := range got.Matched {
		id := rec.Frontmatter.ID
		if id == "SPEC-AUTH-001" {
			t.Errorf("AUTH SPEC leaked into matched set")
		}
	}
}

// TestDiscoverEpic_EmptyMatchIsClean verifies REQ-ES-003 / AC-ES-003b: empty
// match is NOT an error.
func TestDiscoverEpic_EmptyMatchIsClean(t *testing.T) {
	root := t.TempDir()
	writeSpecFixture(t, root, "SPEC-AUTH-001", "Auth", "completed")

	got, err := DiscoverEpic("NONEXIST", Options{BaseDir: root})
	if err != nil {
		t.Fatalf("empty match should not error, got: %v", err)
	}
	if len(got.Matched) != 0 {
		t.Fatalf("expected 0 matched, got %d", len(got.Matched))
	}
}

// TestExtractMx_BASMarker verifies REQ-ES-004 / AC-ES-004: the (TOKEN Mx)
// title regex builds the Mx→SPEC map.
func TestExtractMx_BASMarker(t *testing.T) {
	records := []spec.DocRecord{
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-NAVIGATOR-SYNC-001", Title: "Navigator Sync (BAS M0) — foo"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-NAVIGATOR-SYNC-002", Title: "Navigator Sync (BAS M4) — bar"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-NAVIGATOR-SYNC-003", Title: "Navigator Sync (BAS M1) — baz"}},
	}
	mxMap, untracked, extras, err := ExtractMx(records, "BAS")
	if err != nil {
		t.Fatalf("ExtractMx: %v", err)
	}
	if mxMap["M0"] != "SPEC-NAVIGATOR-SYNC-001" {
		t.Errorf("M0 = %q, want SPEC-NAVIGATOR-SYNC-001", mxMap["M0"])
	}
	if mxMap["M4"] != "SPEC-NAVIGATOR-SYNC-002" {
		t.Errorf("M4 = %q, want SPEC-NAVIGATOR-SYNC-002", mxMap["M4"])
	}
	if mxMap["M1"] != "SPEC-NAVIGATOR-SYNC-003" {
		t.Errorf("M1 = %q, want SPEC-NAVIGATOR-SYNC-003", mxMap["M1"])
	}
	if len(untracked) != 0 {
		t.Errorf("expected 0 untracked, got %v", untracked)
	}
	if len(extras) != 0 {
		t.Errorf("expected 0 extras, got %v", extras)
	}
}

// TestExtractMx_UntrackedAndExtras verifies AC-ES-004 + edge case E1: a SPEC
// with no marker is untracked; a SPEC with a second marker records the extra.
func TestExtractMx_UntrackedAndExtras(t *testing.T) {
	records := []spec.DocRecord{
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-NAVIGATOR-SYNC-001", Title: "Navigator Sync (BAS M0) — foo"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-NAVIGATOR-SYNC-EXTRA", Title: "Navigator Sync — no marker"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-NAVIGATOR-SYNC-002", Title: "Sync (BAS M1) foo (BAS M2) bar"}},
	}
	mxMap, untracked, extras, err := ExtractMx(records, "BAS")
	if err != nil {
		t.Fatalf("ExtractMx: %v", err)
	}
	if mxMap["M0"] != "SPEC-NAVIGATOR-SYNC-001" {
		t.Errorf("M0 = %q, want SPEC-NAVIGATOR-SYNC-001", mxMap["M0"])
	}
	// first marker wins → M1 should be set; M2 is an extra
	if mxMap["M1"] != "SPEC-NAVIGATOR-SYNC-002" {
		t.Errorf("M1 = %q, want SPEC-NAVIGATOR-SYNC-002", mxMap["M1"])
	}
	if _, has := mxMap["M2"]; has {
		t.Errorf("M2 should not be in the map (first marker wins)")
	}
	if len(extras) == 0 {
		t.Errorf("expected at least one extra (M2 from SYNC-002), got 0")
	}
	// untracked: the no-marker SPEC
	found := false
	for _, id := range untracked {
		if id == "SPEC-NAVIGATOR-SYNC-EXTRA" {
			found = true
		}
	}
	if !found {
		t.Errorf("SPEC-NAVIGATOR-SYNC-EXTRA missing from untracked: %v", untracked)
	}
}

// TestExtractMx_GenericToken verifies AC-ES-004b: the --marker flag overrides
// the default token, so a non-BAS token is recognized.
func TestExtractMx_GenericToken(t *testing.T) {
	records := []spec.DocRecord{
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-CUSTOM-001", Title: "Custom Epic (EPICX M3) — sub-feature"}},
	}
	mxMap, _, _, err := ExtractMx(records, "EPICX")
	if err != nil {
		t.Fatalf("ExtractMx: %v", err)
	}
	if mxMap["M3"] != "SPEC-CUSTOM-001" {
		t.Errorf("M3 = %q, want SPEC-CUSTOM-001", mxMap["M3"])
	}
}

// TestInferToken_Mode verifies design.md §3 Stage 2: the inferred token is
// the most-frequent token across the matched set.
func TestInferToken_Mode(t *testing.T) {
	records := []spec.DocRecord{
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-X-001", Title: "(BAS M0) foo"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-X-002", Title: "(BAS M1) foo"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-X-003", Title: "(OTHER M0) foo"}},
	}
	token := InferToken(records, "")
	if token != "BAS" {
		t.Errorf("InferToken = %q, want BAS (mode)", token)
	}
}

// TestInferToken_Override verifies the explicit-marker override wins.
func TestInferToken_Override(t *testing.T) {
	records := []spec.DocRecord{
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-X-001", Title: "(BAS M0) foo"}},
	}
	if got := InferToken(records, "OTHER"); got != "OTHER" {
		t.Errorf("InferToken override = %q, want OTHER", got)
	}
}
