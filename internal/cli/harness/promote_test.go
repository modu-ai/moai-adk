// Package harness — `moai harness promote` tests (SPEC-HARNESS-LOOP-REPAIR-001 M2-1).
//
// A proposalgen discovery draft's designed consumer is manager-spec SPEC
// authoring, not Applier.Apply() (spec.md §A.4). The promote verb routes a
// draft to that consumer: it materialises a SPEC skeleton carrying the draft ID
// as provenance, records a durable audit record, and moves the draft out of the
// pending queue.
package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// producerDraftBody is byte-shaped after a real proposalgen-written draft: it
// carries pattern-discovery metadata (pattern_key, tier:"auto_update") and NO
// edit instruction (no target_path/field_key/new_value). It is not an apply
// input (spec.md §A.4).
func producerDraftBody(draftID string) string {
	return `{
  "confidence": 1,
  "draft_id": "` + draftID + `",
  "generated_at": "2026-07-26T03:19:55Z",
  "generator_version": "0.1.0",
  "observation_count": 42,
  "pattern_key": "user_prompt::` + draftID + `",
  "source_ts": "2026-06-17T11:59:54Z",
  "tier": "auto_update"
}
`
}

// readProvenance extracts the `provenance:` value from a spec.md frontmatter.
// Returns the raw value (trimmed) and ok=false if the field is absent.
func readProvenance(specMD string) (string, bool) {
	for _, line := range strings.Split(specMD, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "provenance:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "provenance:"))
			val = strings.Trim(val, "\"'")
			return val, true
		}
	}
	return "", false
}

// TestPromote_CreatesSPECSkeletonWithProvenance is the AC-HLR-004 reproduction.
//
// Falsification #1 (acceptance.md §D.1): remove the promotion path → no SPEC
// directory is created and RunPromote returns an error.
func TestPromote_CreatesSPECSkeletonWithProvenance(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const draftID = "PROPOSAL-20260728-aaaaaaaa"
	writeNestedDraftRaw(t, root, draftID, producerDraftBody(draftID))

	if err := RunPromote(PromoteOptions{ID: draftID, ProjectRoot: root}); err != nil {
		t.Fatalf("RunPromote: %v", err)
	}

	specMDPath := filepath.Join(root, ".moai", "specs", draftID, "spec.md")
	data, err := os.ReadFile(specMDPath)
	if err != nil {
		t.Fatalf("SPEC skeleton spec.md not created at %s: %v", specMDPath, err)
	}
	prov, ok := readProvenance(string(data))
	if !ok {
		t.Fatalf("spec.md frontmatter has no provenance field:\n%s", data)
	}
	if prov != draftID {
		t.Fatalf("provenance = %q, want %q (must equal the input draft ID exactly)", prov, draftID)
	}
}

// TestPromote_ProvenanceRoundTripsDraftID is AC-HLR-004 clause 2 — the
// provenance value is DERIVED from the input, not hard-coded. Promoting two
// different drafts must record two different values.
//
// Falsification #2 (acceptance.md §D.1): hard-coding the provenance passes a
// single-run check but fails this two-run comparison.
func TestPromote_ProvenanceRoundTripsDraftID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const draftA = "PROPOSAL-20260728-bbbbbbbb"
	const draftB = "PROPOSAL-20260728-cccccccc"
	writeNestedDraftRaw(t, root, draftA, producerDraftBody(draftA))
	writeNestedDraftRaw(t, root, draftB, producerDraftBody(draftB))

	for _, id := range []string{draftA, draftB} {
		if err := RunPromote(PromoteOptions{ID: id, ProjectRoot: root}); err != nil {
			t.Fatalf("RunPromote(%s): %v", id, err)
		}
	}

	provenances := map[string]string{}
	for _, id := range []string{draftA, draftB} {
		data, err := os.ReadFile(filepath.Join(root, ".moai", "specs", id, "spec.md"))
		if err != nil {
			t.Fatalf("read spec.md for %s: %v", id, err)
		}
		prov, ok := readProvenance(string(data))
		if !ok {
			t.Fatalf("%s spec.md missing provenance", id)
		}
		provenances[id] = prov
	}
	if provenances[draftA] != draftA || provenances[draftB] != draftB {
		t.Fatalf("provenance not derived from input: A=%q B=%q", provenances[draftA], provenances[draftB])
	}
	if provenances[draftA] == provenances[draftB] {
		t.Fatalf("two different drafts recorded the same provenance %q — value is hard-coded", provenances[draftA])
	}
}

// TestPromote_DraftLeavesPendingQueue is AC-HLR-004 clause 3 — after promotion
// the draft is no longer counted as pending, and the remaining drafts are
// unchanged.
//
// Falsification #3 (acceptance.md §D.1): leaving the draft in the pending queue
// after promotion leaves the count unchanged.
func TestPromote_DraftLeavesPendingQueue(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const promoteMe = "PROPOSAL-20260728-dddddddd"
	const stays = "PROPOSAL-20260728-eeeeeeee"
	writeNestedDraftRaw(t, root, promoteMe, producerDraftBody(promoteMe))
	writeNestedDraftRaw(t, root, stays, producerDraftBody(stays))

	proposalDir := proposalgen.ProposalDir(root)
	before, err := proposalgen.ListDraftIDs(proposalDir)
	if err != nil {
		t.Fatalf("list before: %v", err)
	}
	if len(before) != 2 {
		t.Fatalf("before: %d drafts, want 2", len(before))
	}

	if err := RunPromote(PromoteOptions{ID: promoteMe, ProjectRoot: root}); err != nil {
		t.Fatalf("RunPromote: %v", err)
	}

	after, err := proposalgen.ListDraftIDs(proposalDir)
	if err != nil {
		t.Fatalf("list after: %v", err)
	}
	if len(after) != 1 {
		t.Fatalf("after promotion: %d pending drafts, want 1 (promoted draft must leave the queue)", len(after))
	}
	if after[0] != stays {
		t.Fatalf("remaining draft = %q, want %q (the non-promoted draft must be unchanged)", after[0], stays)
	}
}

// TestPromote_AppendsExactlyOneAuditRecord is AC-HLR-005 — exactly one durable
// record links draft → SPEC with a timestamp; a second promotion of a different
// draft appends exactly one further record (not zero, not two).
//
// Falsification (acceptance.md §D.2): removing the record write leaves the
// count unchanged across a promotion.
func TestPromote_AppendsExactlyOneAuditRecord(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const draftA = "PROPOSAL-20260728-11111111"
	const draftB = "PROPOSAL-20260728-22222222"
	writeNestedDraftRaw(t, root, draftA, producerDraftBody(draftA))
	writeNestedDraftRaw(t, root, draftB, producerDraftBody(draftB))

	if err := RunPromote(PromoteOptions{ID: draftA, ProjectRoot: root}); err != nil {
		t.Fatalf("RunPromote(A): %v", err)
	}
	if err := RunPromote(PromoteOptions{ID: draftB, ProjectRoot: root}); err != nil {
		t.Fatalf("RunPromote(B): %v", err)
	}

	logPath := filepath.Join(root, promoteAuditLogRel)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("audit log not created at %s: %v", logPath, err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("audit log has %d records, want exactly 2 (one per promotion)", len(lines))
	}
	for i, want := range []string{draftA, draftB} {
		if !strings.Contains(lines[i], want) {
			t.Errorf("audit record %d = %q; want it to reference draft %q", i, lines[i], want)
		}
		if !strings.Contains(lines[i], "promoted_at") {
			t.Errorf("audit record %d = %q; missing promoted_at timestamp", i, lines[i])
		}
	}
}

// TestPromote_MissingDraftReturnsError keeps the user-facing diagnostic on a
// genuinely absent draft (no silent success when there is nothing to promote).
func TestPromote_MissingDraftReturnsError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := RunPromote(PromoteOptions{ID: "PROPOSAL-does-not-exist", ProjectRoot: root})
	if err == nil {
		t.Fatal("RunPromote succeeded on a missing draft; want an error")
	}
}
