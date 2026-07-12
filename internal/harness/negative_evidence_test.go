// Package harness — A7 negative-evidence registry tests (SPEC-HARNESS-EVOLVE-003 M1).
// REQ-HEV3-018..022: registry data structure, register-on-reject/rollback,
// re-proposal block, no permanent suppression.
package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAppendReadNegativeEvidence_RoundTrip verifies a write-then-read round-trip
// preserves all fields (REQ-HEV3-018 entry shape).
func TestAppendReadNegativeEvidence_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "negative-evidence.jsonl")

	ts := time.Date(2026, 7, 12, 14, 3, 0, 0, time.UTC)
	entry := NegativeEvidence{
		PatternKey:           "feature+plan+autopilot+success",
		Outcome:              NegativeEvidenceRejected,
		Timestamp:            ts,
		EvidenceCountAtEvent: 7,
		CooldownUntil:        ts.Add(NegativeEvidenceCooldown),
		NewEvidenceSinceEvent: 0,
		MachineSignalRef:     "lineage:manifest.jsonl#ln=42",
		GateOrigin:           GateOriginL3,
	}

	if err := AppendNegativeEvidence(path, entry); err != nil {
		t.Fatalf("AppendNegativeEvidence failed: %v", err)
	}

	entries, err := ReadNegativeEvidence(path)
	if err != nil {
		t.Fatalf("ReadNegativeEvidence failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	got := entries[0]
	if got.PatternKey != entry.PatternKey {
		t.Errorf("PatternKey = %q, want %q", got.PatternKey, entry.PatternKey)
	}
	if got.Outcome != entry.Outcome {
		t.Errorf("Outcome = %q, want %q", got.Outcome, entry.Outcome)
	}
	if got.CooldownUntil != entry.CooldownUntil {
		t.Errorf("CooldownUntil = %v, want %v", got.CooldownUntil, entry.CooldownUntil)
	}
	if got.GateOrigin != entry.GateOrigin {
		t.Errorf("GateOrigin = %q, want %q", got.GateOrigin, entry.GateOrigin)
	}
}

// TestReadNegativeEvidence_FileNotFound verifies Edge-2: a missing registry file
// returns an empty slice (not an error) — fresh-project first-run behavior.
func TestReadNegativeEvidence_FileNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.jsonl")

	entries, err := ReadNegativeEvidence(path)
	if err != nil {
		t.Fatalf("ReadNegativeEvidence on missing file returned error: %v (must be nil per Edge-2)", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0 for missing file", len(entries))
	}
}

// TestAppendNegativeEvidence_AppendMultiple verifies Edge-1: sequential appends
// are preserved in order (append-only jsonl).
func TestAppendNegativeEvidence_AppendMultiple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "negative-evidence.jsonl")
	ts := time.Now().UTC()

	for i := range 3 {
		entry := NegativeEvidence{
			PatternKey:    "key-" + string(rune('A'+i)),
			Outcome:       NegativeEvidenceRejected,
			Timestamp:     ts,
			CooldownUntil: ts.Add(NegativeEvidenceCooldown),
			GateOrigin:    GateOriginL3,
		}
		if err := AppendNegativeEvidence(path, entry); err != nil {
			t.Fatalf("AppendNegativeEvidence[%d] failed: %v", i, err)
		}
	}

	entries, err := ReadNegativeEvidence(path)
	if err != nil {
		t.Fatalf("ReadNegativeEvidence failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("len(entries) = %d, want 3", len(entries))
	}
}

// TestAppendNegativeEvidence_RejectsSPECID verifies REQ-HEV3-028 anti-fabrication:
// the writer rejects a pattern_key containing an internal SPEC ID.
func TestAppendNegativeEvidence_RejectsSPECID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "negative-evidence.jsonl")

	entry := NegativeEvidence{
		PatternKey: "SPEC-HARNESS-EVOLVE-003 was rejected",
		Outcome:    NegativeEvidenceRejected,
		GateOrigin: GateOriginL3,
	}

	err := AppendNegativeEvidence(path, entry)
	if err == nil {
		t.Fatal("AppendNegativeEvidence with SPEC-ID pattern_key succeeded, must be rejected (REQ-HEV3-028)")
	}

	// File must NOT be touched (byte-hash before == after)
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("registry file was created despite rejection: statErr=%v", statErr)
	}
}

// TestValidatePatternKey_AntiFabrication verifies the pattern_key validator
// rejects internal SPEC IDs, REQ tokens, and commit SHA patterns.
func TestValidatePatternKey_AntiFabrication(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"clean structural token", "feature+plan+autopilot+success", false},
		{"SPEC ID", "SPEC-HARNESS-EVOLVE-003 was rejected", true},
		{"REQ token", "REQ-HEV3-021 failed", true},
		{"AC token", "AC-HEV3-016 blocked", true},
		{"empty", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePatternKey(tc.key)
			if tc.wantErr && err == nil {
				t.Errorf("ValidatePatternKey(%q) = nil, want error", tc.key)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ValidatePatternKey(%q) = %v, want nil", tc.key, err)
			}
		})
	}
}

// TestIsReProposalSuppressed_BlockedByCooldown verifies REQ-HEV3-021:
// entry exists with cooldown not elapsed → suppressed (ErrReProposalSuppressed shape).
func TestIsReProposalSuppressed_BlockedByCooldown(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)
	entries := []NegativeEvidence{{
		PatternKey:            "feature+plan+autopilot+success",
		Outcome:               NegativeEvidenceRejected,
		Timestamp:             ts,
		CooldownUntil:         ts.Add(48 * time.Hour),
		NewEvidenceSinceEvent: 5, // enough new evidence, but cooldown not elapsed
		GateOrigin:            GateOriginL3,
	}}

	// 10 hours after rejection — cooldown (48h) not elapsed
	check := IsReProposalSuppressed(entries, "feature+plan+autopilot+success", ts.Add(10*time.Hour))
	if !check.Suppressed {
		t.Error("Suppressed = false, want true (cooldown not elapsed)")
	}
}

// TestIsReProposalSuppressed_BlockedByNewEvidenceCount verifies REQ-HEV3-021:
// cooldown elapsed but new_evidence < N → still suppressed.
func TestIsReProposalSuppressed_BlockedByNewEvidenceCount(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)
	entries := []NegativeEvidence{{
		PatternKey:            "feature+plan+autopilot+success",
		Outcome:               NegativeEvidenceRejected,
		Timestamp:             ts,
		CooldownUntil:         ts.Add(48 * time.Hour),
		NewEvidenceSinceEvent: 2, // less than N=3
		GateOrigin:            GateOriginL3,
	}}

	// 49 hours after rejection — cooldown elapsed, but new evidence < N
	check := IsReProposalSuppressed(entries, "feature+plan+autopilot+success", ts.Add(49*time.Hour))
	if !check.Suppressed {
		t.Error("Suppressed = false, want true (new_evidence < N=3)")
	}
}

// TestIsReProposalSuppressed_LiftsAfterBothClear verifies REQ-HEV3-022:
// once cooldown elapses AND new_evidence >= N, the key re-eligibility lifts.
func TestIsReProposalSuppressed_LiftsAfterBothClear(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)
	entries := []NegativeEvidence{{
		PatternKey:            "feature+plan+autopilot+success",
		Outcome:               NegativeEvidenceRejected,
		Timestamp:             ts,
		CooldownUntil:         ts.Add(48 * time.Hour),
		NewEvidenceSinceEvent: 3, // meets N=3
		GateOrigin:            GateOriginL3,
	}}

	// 49 hours after rejection — cooldown elapsed AND new_evidence >= N
	check := IsReProposalSuppressed(entries, "feature+plan+autopilot+success", ts.Add(49*time.Hour))
	if check.Suppressed {
		t.Error("Suppressed = true, want false (cooldown elapsed + N new evidences — REQ-HEV3-022 re-eligibility)")
	}
}

// TestIsReProposalSuppressed_NoEntry verifies no entry → not suppressed.
func TestIsReProposalSuppressed_NoEntry(t *testing.T) {
	t.Parallel()

	entries := []NegativeEvidence{}
	check := IsReProposalSuppressed(entries, "feature+plan+autopilot+success", time.Now())
	if check.Suppressed {
		t.Error("Suppressed = true, want false (no entry)")
	}
}

// TestIsReProposalSuppressed_CooldownBoundary verifies Edge-3:
// at exactly cooldown_until, the cooldown is elapsed (now >= cooldown_until, inclusive).
func TestIsReProposalSuppressed_CooldownBoundary(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)
	cooldown := ts.Add(48 * time.Hour)
	entries := []NegativeEvidence{{
		PatternKey:            "feature+plan+autopilot+success",
		Outcome:               NegativeEvidenceRejected,
		Timestamp:             ts,
		CooldownUntil:         cooldown,
		NewEvidenceSinceEvent: 3, // meets N — so only cooldown is the gate at the boundary
		GateOrigin:            GateOriginL3,
	}}

	// now == cooldown_until exactly → elapsed (inclusive boundary, Edge-3)
	check := IsReProposalSuppressed(entries, "feature+plan+autopilot+success", cooldown)
	if check.Suppressed {
		t.Error("Suppressed = true at now == cooldown_until, want false (Edge-3 inclusive boundary)")
	}
}

// TestIsReProposalSuppressed_LatestEntryWins verifies that when multiple entries
// exist for the same key, the LATEST entry (by timestamp) is the one consulted.
func TestIsReProposalSuppressed_LatestEntryWins(t *testing.T) {
	t.Parallel()

	ts1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC) // later
	key := "feature+plan+autopilot+success"
	entries := []NegativeEvidence{
		{PatternKey: key, Outcome: NegativeEvidenceRejected, Timestamp: ts1, CooldownUntil: ts1.Add(48 * time.Hour), NewEvidenceSinceEvent: 10, GateOrigin: GateOriginL3},
		{PatternKey: key, Outcome: NegativeEvidenceRolledBack, Timestamp: ts2, CooldownUntil: ts2.Add(48 * time.Hour), NewEvidenceSinceEvent: 0, GateOrigin: GateOriginRollback},
	}

	// Consult at ts2 + 10h — the latest entry (ts2) has cooldown not elapsed
	check := IsReProposalSuppressed(entries, key, ts2.Add(10*time.Hour))
	if !check.Suppressed {
		t.Error("Suppressed = false, want true (latest entry ts2 cooldown not elapsed)")
	}
	if check.LatestEntry == nil || check.LatestEntry.Outcome != NegativeEvidenceRolledBack {
		t.Errorf("LatestEntry outcome = %v, want rolled-back (the ts2 entry)", check.LatestEntry)
	}
}

// TestRegisterVetoAsNegativeEvidence verifies REQ-HEV3-014: a Canary veto
// auto-registers the vetoed pattern key in the A7 registry with outcome "rolled-back".
func TestRegisterVetoAsNegativeEvidence(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "negative-evidence.jsonl")
	ts := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)

	err := RegisterVetoAsNegativeEvidence(path, "feature+plan+autopilot+success", 7, ts)
	if err != nil {
		t.Fatalf("RegisterVetoAsNegativeEvidence failed: %v", err)
	}

	entries, err := ReadNegativeEvidence(path)
	if err != nil {
		t.Fatalf("ReadNegativeEvidence failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Outcome != NegativeEvidenceRolledBack {
		t.Errorf("Outcome = %q, want %q", e.Outcome, NegativeEvidenceRolledBack)
	}
	if e.GateOrigin != GateOriginRollback {
		t.Errorf("GateOrigin = %q, want %q", e.GateOrigin, GateOriginRollback)
	}
	if e.EvidenceCountAtEvent != 7 {
		t.Errorf("EvidenceCountAtEvent = %d, want 7", e.EvidenceCountAtEvent)
	}
}
