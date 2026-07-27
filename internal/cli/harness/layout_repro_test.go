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

// TestLoadProposalByID_ProducerSchemaMismatch is a CHARACTERIZATION test: it
// pins the currently-broken producer/consumer schema contract discovered while
// repairing the layout seam (M1). It documents a blocker for M2 (first
// apply_outcome), NOT desired behaviour.
//
// The producer (proposalgen.marshalProposalJSON) emits pattern-DISCOVERY
// metadata:
//
//	{"tier": "auto_update", "pattern_key": ..., "observation_count": ...}
//
// The consumer (harness.Proposal) expects an actionable EDIT INSTRUCTION:
//
//	{"tier": <int>, "target_path": ..., "field_key": ..., "new_value": ...}
//
// Two independent breaks follow: `tier` is a string on disk but a numeric
// harness.Tier in the struct (a hard unmarshal error), and target_path /
// field_key / new_value are absent entirely (so Applier.Apply would have
// nothing to apply even if parsing succeeded).
//
// Repairing the layout seam makes drafts VISIBLE; it does not make them
// APPLICABLE. When M2 reconciles the two schemas this test MUST fail — that
// failure is the signal to update it, not a regression.
func TestLoadProposalByID_ProducerSchemaMismatch(t *testing.T) {
	root := t.TempDir()
	const draftID = "PROPOSAL-20260727-ffffffff"
	// Byte-shaped after a real producer-written draft on disk.
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
	// The layout seam is repaired, so resolution reaches the real file...
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("nested path did not resolve to the draft on disk: %v", statErr)
	}
	// ...but the schema contract is not, so the load still fails.
	_, err = loadProposalByID(path)
	if err == nil {
		t.Fatal("producer-shaped draft loaded successfully — the M2 schema gap appears " +
			"to be fixed; update this characterization test to assert the new contract")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal string into Go struct field Proposal.tier") {
		t.Fatalf("unexpected failure mode; the known M2 blocker is the string/int tier "+
			"mismatch, got: %v", err)
	}
}
