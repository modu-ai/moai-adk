// Package harness — reproduction test for the C3 proposal-path resolution site
// (SPEC-HARNESS-LOOP-REPAIR-001 M1, AC-HLR-003).
//
// resolveProposalPath historically derived a flat `<proposals>/<id>.json` path
// while the producer writes `<proposals>/<id>/proposal.json`. Every
// `moai harness execute --id <ID>` therefore failed with "proposal not found",
// regardless of how many drafts existed on disk.
package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeNestedDraftJSON creates a draft in the producer's nested layout and
// returns the path of the proposal.json it wrote. The body is shaped to the
// CONSUMER schema (harness.Proposal) so path-resolution tests exercise only the
// layout contract. See TestLoadProposalByID_ProducerSchemaMismatch for the
// separate producer-schema gap.
func writeNestedDraftJSON(t *testing.T, projectRoot, draftID string) string {
	t.Helper()
	path := writeNestedDraftRaw(t, projectRoot, draftID,
		`{"id":"`+draftID+`","pattern_key":"test_pattern","tier":1,"observation_count":7}`+"\n")
	return path
}

// writeNestedDraftRaw writes an arbitrary proposal.json body in the producer's
// nested layout, so a test can pin the exact on-disk shape it means to assert.
func writeNestedDraftRaw(t *testing.T, projectRoot, draftID, body string) string {
	t.Helper()
	draftDir := filepath.Join(projectRoot, ".moai", "harness", "proposals", draftID)
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		t.Fatalf("draft 디렉터리 생성 실패: %v", err)
	}
	path := filepath.Join(draftDir, "proposal.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("proposal.json 작성 실패: %v", err)
	}
	return path
}

// TestResolveProposalPath_NestedLayout is the C3 reproduction (AC-HLR-003).
//
// Falsification: restoring the `id+".json"` derivation points at a file that
// does not exist and fails this test.
func TestResolveProposalPath_NestedLayout(t *testing.T) {
	root := t.TempDir()
	const draftID = "PROPOSAL-20260727-dddddddd"
	want := writeNestedDraftJSON(t, root, draftID)

	got, err := resolveProposalPath(root, draftID)
	if err != nil {
		t.Fatalf("resolveProposalPath 오류: %v", err)
	}
	if got != want {
		t.Fatalf("resolveProposalPath = %q, want %q", got, want)
	}
	if _, statErr := os.Stat(got); statErr != nil {
		t.Fatalf("resolved path does not exist on disk: %v", statErr)
	}
}

// TestResolveProposalPath_RejectsTraversal preserves the existing path-traversal
// guard — the accessor move must not weaken input validation.
func TestResolveProposalPath_RejectsTraversal(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"", "../escape", `sub/dir`, `..`, `a\b`} {
		if _, err := resolveProposalPath(root, id); err == nil {
			t.Errorf("resolveProposalPath(%q) accepted an unsafe ID; want error", id)
		}
	}
}

// TestLoadProposalByID_NestedLayout asserts the execute path loads a draft end
// to end through the resolved nested path.
func TestLoadProposalByID_NestedLayout(t *testing.T) {
	root := t.TempDir()
	const draftID = "PROPOSAL-20260727-eeeeeeee"
	writeNestedDraftJSON(t, root, draftID)

	path, err := resolveProposalPath(root, draftID)
	if err != nil {
		t.Fatalf("resolveProposalPath 오류: %v", err)
	}
	prop, err := loadProposalByID(path)
	if err != nil {
		t.Fatalf("loadProposalByID 오류: %v", err)
	}
	if prop.PatternKey != "test_pattern" {
		t.Fatalf("loaded PatternKey = %q, want %q", prop.PatternKey, "test_pattern")
	}
}

// TestLoadProposalByID_MissingDraft keeps the user-facing "proposal not found"
// diagnostic on a genuinely absent draft.
func TestLoadProposalByID_MissingDraft(t *testing.T) {
	root := t.TempDir()
	path, err := resolveProposalPath(root, "PROPOSAL-does-not-exist")
	if err != nil {
		t.Fatalf("resolveProposalPath 오류: %v", err)
	}
	_, err = loadProposalByID(path)
	if err == nil {
		t.Fatal("loadProposalByID succeeded on a missing draft; want an error")
	}
	if !strings.Contains(err.Error(), "proposal not found") {
		t.Fatalf("error = %q, want it to mention \"proposal not found\"", err.Error())
	}
	// The nested layout puts every draft's metadata at .../<ID>/proposal.json,
	// so a filename-only diagnostic would name "proposal.json" for every draft.
	// The message must carry the draft ID to be actionable.
	if !strings.Contains(err.Error(), "PROPOSAL-does-not-exist") {
		t.Fatalf("error = %q, want it to name the draft ID", err.Error())
	}
}

// TestLoadProposalByID_DiscoveryDraftDiagnostic is the AC-HLR-014 reproduction
// (SPEC-HARNESS-LOOP-REPAIR-001 M2-3).
//
// A proposalgen discovery draft carries pattern-discovery metadata
// (tier:"auto_update", pattern_key, observation_count) and NO edit instruction
// (target_path/field_key/new_value). It is not an apply input (spec.md §A.4).
// loadProposalByID MUST reject it with an honest diagnostic naming the reason,
// NOT the raw "cannot unmarshal string into Go struct field Proposal.tier"
// symptom that the string/numeric tier split produces.
//
// Falsification (acceptance.md §D.3): reverting the de-wiring restores the raw
// unmarshal diagnostic.
func TestLoadProposalByID_DiscoveryDraftDiagnostic(t *testing.T) {
	root := t.TempDir()
	const draftID = "PROPOSAL-20260728-gggggggg"
	// Byte-shaped after a real proposalgen-written draft: string tier, no edit fields.
	writeNestedDraftRaw(t, root, draftID, `{
  "confidence": 1,
  "draft_id": "`+draftID+`",
  "generated_at": "2026-07-26T03:19:55Z",
  "generator_version": "0.1.0",
  "observation_count": 196,
  "pattern_key": "user_prompt::",
  "source_ts": "2026-06-17T11:59:54Z",
  "tier": "auto_update"
}
`)

	path, err := resolveProposalPath(root, draftID)
	if err != nil {
		t.Fatalf("resolveProposalPath 오류: %v", err)
	}
	_, err = loadProposalByID(path)
	if err == nil {
		t.Fatal("loadProposalByID succeeded on a discovery draft; want a diagnostic error")
	}
	msg := err.Error()
	// MUST NOT surface the raw unmarshal symptom.
	if strings.Contains(msg, "cannot unmarshal string into Go struct field") {
		t.Fatalf("error is the raw unmarshal symptom, not the honest diagnostic:\n%s", msg)
	}
	// MUST name the discovery-draft reason and point at the promotion path.
	for _, want := range []string{"discovery draft", "target_path", "promote"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing substring %q", msg, want)
		}
	}
}

// RETIRED (SPEC-HARNESS-LOOP-REPAIR-001 M2-4, AC-HLR-016 clause 2):
// TestLoadProposalByID_ProducerSchemaMismatch previously pinned the
// producer/consumer tier mismatch (producer writes "tier":"auto_update" string;
// harness.Proposal.Tier is numeric) as a CHARACTERIZATION of a known-broken
// state. M2-3 de-wires the execute→draft path: loadProposalByID now detects a
// producer-shaped discovery draft and returns an honest diagnostic instead of
// the raw unmarshal error. The mismatch therefore no longer has a live
// consumer — REQ-HLR-012 DISSOLVES the tier split rather than repairing it (no
// harness.Tier JSON codec is added; the only consumer was this execute path).
// The new contract is pinned by TestLoadProposalByID_DiscoveryDraftDiagnostic
// above. The retired test is reproduced below as a comment for traceability:
// it asserted err.Error() contained "cannot unmarshal string into Go struct
// field Proposal.tier", which is exactly the raw symptom M2-3 replaces.
//
// (The test function body is intentionally removed; do not restore it without
// also reverting M2-3 — the two are coupled.)

// TestLoadProposalByID_ProducerSchemaMismatch_REMOVED is a sentinel that keeps
// the retirement grep-discoverable without re-asserting the defunct contract.
func TestLoadProposalByID_ProducerSchemaMismatch_RetiredByM2(t *testing.T) {
	t.Parallel()
	// This test intentionally asserts nothing about the old mismatch. It exists
	// so `grep ProducerSchemaMismatch` still resolves to a documented retirement
	// rather than a silent deletion (AC-HLR-016 clause 2 rationale requirement).
}
