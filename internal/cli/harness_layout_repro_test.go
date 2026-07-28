// Package cli — reproduction tests for the proposal layout contract mismatch
// (SPEC-HARNESS-LOOP-REPAIR-001 M1, AC-HLR-001 / AC-HLR-002 / AC-HLR-006).
//
// The proposal producer (internal/harness/proposalgen) writes one DIRECTORY per
// draft:
//
//	.moai/harness/proposals/<DRAFT-ID>/
//	  ├── spec.md
//	  └── proposal.json
//
// Every consumer historically assumed a FLAT file named <DRAFT-ID>.json directly
// under proposals/, guarded by a `!e.IsDir()` predicate that excludes every
// generated draft by construction. The mismatch is total and silent: no error
// surfaces, the CLI simply reports an empty queue.
//
// These tests build the producer's real nested layout under t.TempDir() and
// assert the consumers can see it. Reverting the shared accessor restores the
// `0 items` / "No pending proposals" behaviour and fails them again.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeNestedDraft creates one proposal draft in the producer's nested layout.
// It deliberately mirrors proposalgen.WriteProposals rather than calling it, so
// the fixture stays independent of the producer's own test helpers.
func writeNestedDraft(t *testing.T, projectRoot, draftID string) {
	t.Helper()
	draftDir := filepath.Join(projectRoot, ".moai", "harness", "proposals", draftID)
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		t.Fatalf("draft 디렉터리 생성 실패: %v", err)
	}
	body := `{
  "draft_id": "` + draftID + `",
  "pattern_key": "test_pattern",
  "observation_count": 7,
  "confidence": 0.8,
  "tier": "observation"
}
`
	if err := os.WriteFile(filepath.Join(draftDir, "proposal.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("proposal.json 작성 실패: %v", err)
	}
	// spec.md is the human body the producer writes alongside proposal.json.
	if err := os.WriteFile(filepath.Join(draftDir, "spec.md"), []byte("# "+draftID+"\n"), 0o644); err != nil {
		t.Fatalf("spec.md 작성 실패: %v", err)
	}
}

// TestCountProposals_NestedLayout is the C1 reproduction (AC-HLR-001).
//
// Falsification: reverting countProposals to the `!e.IsDir()` predicate makes
// this return 0 against the same fixture.
func TestCountProposals_NestedLayout(t *testing.T) {
	root := t.TempDir()
	writeNestedDraft(t, root, "PROPOSAL-20260727-aaaaaaaa")
	writeNestedDraft(t, root, "PROPOSAL-20260727-bbbbbbbb")

	got := countProposals(filepath.Join(root, harnessDefaultProposalDir))
	if got != 2 {
		t.Fatalf("countProposals = %d, want 2 (nested proposals/<ID>/proposal.json layout)", got)
	}
}

// TestCountProposals_IgnoresIncompleteDraft asserts a directory without a
// proposal.json is not counted — a draft is defined by its metadata file, not
// by the mere existence of a directory.
func TestCountProposals_IgnoresIncompleteDraft(t *testing.T) {
	root := t.TempDir()
	writeNestedDraft(t, root, "PROPOSAL-20260727-complete")
	// A stray directory carrying no proposal.json is not a draft.
	stray := filepath.Join(root, harnessDefaultProposalDir, "PROPOSAL-20260727-empty")
	if err := os.MkdirAll(stray, 0o755); err != nil {
		t.Fatalf("stray 디렉터리 생성 실패: %v", err)
	}

	if got := countProposals(filepath.Join(root, harnessDefaultProposalDir)); got != 1 {
		t.Fatalf("countProposals = %d, want 1 (incomplete draft must not count)", got)
	}
}

// TestCountProposals_MissingDir preserves the graceful no-op contract: an absent
// proposals directory reports zero rather than erroring.
func TestCountProposals_MissingDir(t *testing.T) {
	root := t.TempDir()
	if got := countProposals(filepath.Join(root, harnessDefaultProposalDir)); got != 0 {
		t.Fatalf("countProposals = %d, want 0 for a missing proposals directory", got)
	}
}

// TestHarnessApply_NestedLayout is the C2 reproduction (AC-HLR-002).
//
// Falsification: reverting the apply selector to the `!e.IsDir()` predicate
// restores the "No pending proposals." branch and fails this test.
func TestHarnessApply_NestedLayout(t *testing.T) {
	root := t.TempDir()
	const draftID = "PROPOSAL-20260727-cccccccc"
	writeNestedDraft(t, root, draftID)

	cmd := newHarnessApplyCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--project-root", root})
	// --project-root is a persistent flag on the router; declare it locally so
	// the standalone factory under test can resolve the project root.
	cmd.Flags().String("project-root", root, "project root")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("harness apply 실행 실패: %v", err)
	}

	got := out.String()
	if strings.Contains(got, "No pending proposals") {
		t.Fatalf("apply reported an empty queue against a nested draft fixture:\n%s", got)
	}
	if !strings.Contains(got, draftID) {
		t.Fatalf("apply output does not carry the draft ID %q:\n%s", draftID, got)
	}
}

// TestNoFlatProposalPathDerivation is the AC-HLR-006 regression guard: no call
// site may re-derive a flat `<id>.json` proposal path. Every consumer must
// resolve through the shared accessor.
//
// Falsification: reintroducing `id+".json"` at any consumer site fails this.
func TestNoFlatProposalPathDerivation(t *testing.T) {
	// Scanned relative to internal/cli (this test's package directory).
	targets := []string{"harness.go", filepath.Join("harness", "execute.go")}
	for _, rel := range targets {
		body, err := os.ReadFile(rel)
		if err != nil {
			t.Fatalf("소스 읽기 실패 %s: %v", rel, err)
		}
		for i, line := range strings.Split(string(body), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue // prose describing the retired layout is allowed
			}
			if strings.Contains(line, `id+".json"`) || strings.Contains(line, `id + ".json"`) {
				t.Errorf("%s:%d re-derives a flat proposal path; use the shared accessor: %s",
					rel, i+1, trimmed)
			}
		}
	}
}
