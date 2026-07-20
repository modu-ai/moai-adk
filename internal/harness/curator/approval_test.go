package curator_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// writeGatedFixture writes a CLAUDE.md under a t.TempDir()-isolated directory
// and returns its absolute path. This external test package (curator_test)
// cannot use the internal writeFixture helper, so it defines its own.
func writeGatedFixture(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

func readGated(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// TestWriteManagedBlock_RequiresApprovalToken (AC-HEV2-040, REQ-HEV2-032): the
// gated writer requires an L5 approval token to execute. A rejection leaves the
// file untouched and returns ErrApprovalRejected (no autonomous write path,
// AP-HEV2-003); an approval writes the block.
func TestWriteManagedBlock_RequiresApprovalToken(t *testing.T) {
	path := writeGatedFixture(t, "# Project\n")
	content := curator.BlockContent{Bullets: []curator.Bullet{
		{LedgerKey: "k1", Text: "distilled workflow rule"},
	}}

	// Rejected decision → no file write.
	before := readGated(t, path)
	err := curator.WriteManagedBlockGated(path, curator.BlockTypeLearnedWorkflow, content,
		curator.ApprovalDecision{Approved: false, Rationale: "not this cycle"}, nil)
	if !errors.Is(err, curator.ErrApprovalRejected) {
		t.Fatalf("rejected write should return ErrApprovalRejected, got %v", err)
	}
	if readGated(t, path) != before {
		t.Errorf("file modified despite L5 rejection (no autonomous write path)")
	}

	// Approved decision → the block is written.
	if err := curator.WriteManagedBlockGated(path, curator.BlockTypeLearnedWorkflow, content,
		curator.ApprovalDecision{Approved: true}, nil); err != nil {
		t.Fatalf("approved write should succeed, got %v", err)
	}
	if !strings.Contains(readGated(t, path), "distilled workflow rule") {
		t.Errorf("approved write did not persist the bullet")
	}
}

// TestWriteManagedBlock_RejectionRecordsLineage_NoFileWrite (AC-HEV2-041,
// REQ-HEV2-032): on rejection the writer records a LineageEntry with decision
// "rejected" carrying the rejection rationale (audit trail) via the injected
// recorder, and does NOT touch the file.
func TestWriteManagedBlock_RejectionRecordsLineage_NoFileWrite(t *testing.T) {
	path := writeGatedFixture(t, "# Project\n")
	before := readGated(t, path)
	manifest := filepath.Join(t.TempDir(), "manifest.jsonl")
	const rationale = "reviewer rejected: rule too broad"

	// The orchestrator/applier injects a recorder that appends a real
	// LineageEntry (curator cannot import harness — this external test can).
	recorder := func(outcome, rat string) error {
		return harness.WriteLineageEntry(manifest, harness.LineageEntry{
			ProposalID:     "hev2-041",
			TargetPath:     path,
			Decision:       outcome, // "rejected"
			Reason:         rat,
			LearnedSurface: "claude.md.learned-workflow",
			// BulletsChanged nil → serialized as null (evidence-or-null on a reject).
		})
	}

	err := curator.WriteManagedBlockGated(path, curator.BlockTypeLearnedWorkflow,
		curator.BlockContent{Bullets: []curator.Bullet{{LedgerKey: "k1", Text: "distilled rule"}}},
		curator.ApprovalDecision{Approved: false, Rationale: rationale}, recorder)
	if !errors.Is(err, curator.ErrApprovalRejected) {
		t.Fatalf("expected ErrApprovalRejected, got %v", err)
	}

	// File untouched.
	if readGated(t, path) != before {
		t.Errorf("file modified despite rejection")
	}

	// Lineage records exactly one "rejected" entry carrying the rationale.
	entries, lerr := harness.LoadManifest(manifest)
	if lerr != nil {
		t.Fatalf("LoadManifest: %v", lerr)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 lineage entry, got %d", len(entries))
	}
	if entries[0].Decision != "rejected" {
		t.Errorf("lineage decision = %q, want %q", entries[0].Decision, "rejected")
	}
	if entries[0].Reason != rationale {
		t.Errorf("lineage reason = %q, want %q", entries[0].Reason, rationale)
	}
}
