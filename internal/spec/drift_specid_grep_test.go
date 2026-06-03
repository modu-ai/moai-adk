package spec

import (
	"os/exec"
	"testing"
)

// TestGetGitImpliedStatus_HARNESS001Resolution verifies the walker adopts the
// GENUINE SPEC-V3R4-HARNESS-001 sync commit, NOT substring noise from the
// NAMESPACE-001 plan commit (the word-boundary protection assertion).
//
// Word-boundary semantics (the LSGF-001 protection this test guards): the walker
// adopts the genuine `sync(SPEC-V3R4-HARNESS-001): ... draft → implemented` commit
// (e8e38b17b) and NOT the `SPEC-V3R4-HARNESS-NAMESPACE-001` plan-commit substring
// noise (which would yield "planned").
//
//   - Pre-LSGF-001 (substring matching): returned "planned" (NAMESPACE noise).
//   - Post-LSGF-001 (word-boundary matching): adopts the genuine HARNESS-001 sync commit.
//
// SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 M2 (mechanism ②): the genuine sync commit
// subject literally reads "status transition draft → implemented", so the corrected
// 4-phase model resolves it to "implemented" (sync-phase = implemented; "completed"
// is reachable ONLY via a close-infix). The legacy expectation was "completed"
// under the stale {sync→completed} rule; the corrected rule yields "implemented".
// The word-boundary protection itself is UNCHANGED — that logic is fully verified
// by TestGetGitImpliedStatus_SPECIDWordBoundary's 5 sub-cases (still green). The
// binding assertion here is "NOT planned" (no NAMESPACE noise); "implemented" is
// the corrected genuine-signal value.
//
// Because this test depends on live git history, it is skipped in:
//   - fork/shallow clones where SPEC-V3R4-HARNESS-001 commits are absent.
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
	// @MX:NOTE: [AUTO] CI shallow-clone skip guard permanently removed.
	// @MX:REASON: After SPEC-V3R4-CI-INFRA-FIX-001 W3, all 5 ci.yml checkout steps
	// use fetch-depth: 0; SPEC commits are available in CI environments too.
	// Permanently resolves the GITHUB_ACTIONS env workaround from LSGF-001 PR #948.

	// Probe: verify target SPEC commits exist in local git (covers fork/shallow-clone user environments).
	probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
	if out, err := probe.Output(); err != nil || len(out) == 0 {
		t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone). " +
			"WordBoundary helper test (5 sub-cases) covers the logic.")
	}

	status, err := getGitImpliedStatus("SPEC-V3R4-HARNESS-001")
	if err != nil {
		t.Fatalf("getGitImpliedStatus returned unexpected error: %v", err)
	}
	// Binding word-boundary assertion: must NOT be "planned" (NAMESPACE substring noise).
	if status == "planned" {
		t.Errorf("got 'planned' — NAMESPACE-001 plan-commit substring noise was adopted (LSGF-001 word-boundary regression)")
	}
	// Corrected genuine-signal value (mechanism ②): sync-phase = implemented.
	if status != "implemented" {
		t.Errorf("expected status 'implemented' (genuine sync signal; sync-phase = implemented in 4-phase model), got %q", status)
	}
}

// TestGetGitImpliedStatus_SPECIDWordBoundary verifies that the commitMatchesSPECID
// helper performs precise SPEC-ID word-boundary matching.
//
// Five sub-cases:
//   - C1: exact match (plan)
//   - C2: NAMESPACE substring noise (false-positive blocked)
//   - C3: exact match (sync)
//   - C4: chore-post token (no SPEC- prefix)
//   - C5: closeout body (mentions HARNESS-001 only, without the SPEC- prefix)
func TestGetGitImpliedStatus_SPECIDWordBoundary(t *testing.T) {
	tests := []struct {
		name        string
		commitTitle string
		specID      string
		want        bool
	}{
		{
			name:        "C1 exact match (plan)",
			commitTitle: "plan(spec): SPEC-V3R4-HARNESS-001 — initial",
			specID:      "SPEC-V3R4-HARNESS-001",
			want:        true,
		},
		{
			name:        "C2 substring noise (NAMESPACE)",
			commitTitle: "plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — supersedes 001",
			specID:      "SPEC-V3R4-HARNESS-001",
			want:        false,
		},
		{
			name:        "C3 exact match (sync)",
			commitTitle: "sync(SPEC-V3R4-HARNESS-001): status transition",
			specID:      "SPEC-V3R4-HARNESS-001",
			want:        true,
		},
		{
			name:        "C4 chore-post token (no SPEC- prefix)",
			commitTitle: "chore(post-V3R4-HARNESS-001): cleanup",
			specID:      "SPEC-V3R4-HARNESS-001",
			want:        false,
		},
		{
			name:        "C5 closeout body (HARNESS-001 without SPEC- prefix)",
			commitTitle: "sync(specs): closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed)",
			specID:      "SPEC-V3R4-HARNESS-001",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commitMatchesSPECID(tt.commitTitle, tt.specID)
			if got != tt.want {
				t.Errorf("commitMatchesSPECID(%q, %q) = %v, want %v",
					tt.commitTitle, tt.specID, got, tt.want)
			}
		})
	}
}
