package factory

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestCandidateRecord pins the candidate record store (SPEC-CANDIDATE-CI-001
// REQ-CCI-005): a record keyed by (card id, pinned SHA) round-trips through
// the primary checkout's state store, an absent key reads as an explicit
// absence (not an error and not a zero verdict), and a record for a
// different pinned SHA of the same card does not collide with it.
func TestCandidateRecord(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	rec := CandidateRecord{
		CardID:            "t1478",
		PinnedSHA:         "1111111111111111111111111111111111111111",
		CandidateSHA:      "2222222222222222222222222222222222222222",
		IntegrationBranch: "develop",
		IntegrationTip:    "3333333333333333333333333333333333333333",
		CandidateBranch:   "ci/t1478",
		Verdict:           CandidateVerdictPending,
		PushedAt:          "2026-10-09T00:00:00Z",
	}
	if err := WriteCandidateRecord(root, rec); err != nil {
		t.Fatalf("WriteCandidateRecord: %v", err)
	}

	// Round-trip: the written record reads back field-identical.
	got, err := ReadCandidateRecord(root, rec.CardID, rec.PinnedSHA)
	if err != nil {
		t.Fatalf("ReadCandidateRecord: %v", err)
	}
	if got.CardID != rec.CardID || got.PinnedSHA != rec.PinnedSHA {
		t.Errorf("key: got (%s, %s), want (%s, %s)", got.CardID, got.PinnedSHA, rec.CardID, rec.PinnedSHA)
	}
	if got.CandidateSHA != rec.CandidateSHA {
		t.Errorf("CandidateSHA: got %s, want %s", got.CandidateSHA, rec.CandidateSHA)
	}
	if got.IntegrationBranch != rec.IntegrationBranch || got.IntegrationTip != rec.IntegrationTip {
		t.Errorf("integration identity: got (%s, %s), want (%s, %s)", got.IntegrationBranch, got.IntegrationTip, rec.IntegrationBranch, rec.IntegrationTip)
	}
	if got.CandidateBranch != rec.CandidateBranch {
		t.Errorf("CandidateBranch: got %s, want %s", got.CandidateBranch, rec.CandidateBranch)
	}
	if got.Verdict != CandidateVerdictPending {
		t.Errorf("Verdict: got %q, want %q", got.Verdict, CandidateVerdictPending)
	}
	if got.PushedAt != rec.PushedAt {
		t.Errorf("PushedAt: got %s, want %s", got.PushedAt, rec.PushedAt)
	}

	// Absent key: an explicit absence, distinguishable from a store error.
	if _, err := ReadCandidateRecord(root, "t9999", rec.PinnedSHA); !errors.Is(err, ErrCandidateRecordAbsent) {
		t.Errorf("absent card key: got %v, want ErrCandidateRecordAbsent", err)
	}
	if _, err := ReadCandidateRecord(root, rec.CardID, "4444444444444444444444444444444444444444"); !errors.Is(err, ErrCandidateRecordAbsent) {
		t.Errorf("absent pinned key: got %v, want ErrCandidateRecordAbsent", err)
	}

	// Same card, different pinned SHA: a distinct record, no collision.
	second := rec
	second.PinnedSHA = "5555555555555555555555555555555555555555"
	second.Verdict = CandidateVerdictGreen
	if err := WriteCandidateRecord(root, second); err != nil {
		t.Fatalf("WriteCandidateRecord (second): %v", err)
	}
	reReadFirst, err := ReadCandidateRecord(root, rec.CardID, rec.PinnedSHA)
	if err != nil {
		t.Fatalf("ReadCandidateRecord (first after second): %v", err)
	}
	if reReadFirst.Verdict != CandidateVerdictPending {
		t.Errorf("first record's verdict after second write: got %q, want %q (records are keyed by pinned SHA)", reReadFirst.Verdict, CandidateVerdictPending)
	}
	reReadSecond, err := ReadCandidateRecord(root, second.CardID, second.PinnedSHA)
	if err != nil {
		t.Fatalf("ReadCandidateRecord (second): %v", err)
	}
	if reReadSecond.Verdict != CandidateVerdictGreen {
		t.Errorf("second record's verdict: got %q, want %q", reReadSecond.Verdict, CandidateVerdictGreen)
	}

	// The store lives under the PRIMARY checkout's .moai/state, shared by
	// every linked worktree (REQ-CCI-005).
	want := filepath.Join(root, ".moai", "state", "candidate", "t1478")
	gotPath, pathErr := candidateRecordPath(root, rec.CardID, rec.PinnedSHA)
	if pathErr != nil {
		t.Fatalf("candidateRecordPath: %v", pathErr)
	}
	if filepath.Dir(gotPath) != want {
		t.Errorf("record path %s: want it under %s", gotPath, want)
	}
	if filepath.Base(gotPath) != rec.PinnedSHA+".json" {
		t.Errorf("record file name: got %s, want %s", filepath.Base(gotPath), rec.PinnedSHA+".json")
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("store dir %s: %v", want, err)
	}
}
