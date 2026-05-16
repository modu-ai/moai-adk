package spec

import (
	"os"
	"os/exec"
	"testing"
)

// TestGetGitImpliedStatus_HARNESS001Resolution은 walker가
// SPEC-V3R4-HARNESS-001에 대해 올바른 sync 신호를 반환하는지 검증한다.
// NAMESPACE-001 plan commit의 substring 노이즈가 아닌 실제 completed 상태를 반환해야 한다.
//
// Pre-fix (substring matching): "planned" 반환 (NAMESPACE plan commit 노이즈).
// Post-fix (word-boundary matching): "completed" 반환 (sync commit의 올바른 신호).
//
// 본 테스트는 live git history에 의존하므로 다음 환경에서는 skip된다:
//   - GitHub Actions CI (actions/checkout@v4 default fetch-depth: 1, shallow clone)
//   - SPEC-V3R4-HARNESS-001 commits이 부재한 fork/clone
// Word-boundary 헬퍼 로직은 TestGetGitImpliedStatus_SPECIDWordBoundary 5 sub-cases가 완전 검증한다.
// CI fetch-depth: 0 영구 fix는 후속 SPEC-V3R4-CI-INFRA-FIX-001에서 다룬다.
func TestGetGitImpliedStatus_HARNESS001Resolution(t *testing.T) {
	// CI 환경 자동 skip — shallow clone으로 SPEC commits 부재
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("requires full git history; CI uses actions/checkout@v4 default fetch-depth: 1 (shallow). " +
			"Word-boundary logic is fully covered by TestGetGitImpliedStatus_SPECIDWordBoundary 5 sub-cases. " +
			"Follow-up: SPEC-V3R4-CI-INFRA-FIX-001 to set fetch-depth: 0 for full-history tests.")
	}

	// Probe: target SPEC commits이 local git에 존재하는지 확인 (non-CI 환경에서도 fork/shallow 대응)
	probe := exec.Command("git", "log", "main", "--oneline", "--grep=SPEC-V3R4-HARNESS-001", "-1")
	if out, err := probe.Output(); err != nil || len(out) == 0 {
		t.Skip("SPEC-V3R4-HARNESS-001 commits not available in local git history (fork/shallow clone). " +
			"WordBoundary helper test (5 sub-cases) covers the logic.")
	}

	status, err := getGitImpliedStatus("SPEC-V3R4-HARNESS-001")
	if err != nil {
		t.Fatalf("getGitImpliedStatus returned unexpected error: %v", err)
	}
	if status != "completed" {
		t.Errorf("expected status 'completed' (genuine sync signal), got %q (likely NAMESPACE substring noise)", status)
	}
}

// TestGetGitImpliedStatus_SPECIDWordBoundary는 commitMatchesSPECID 헬퍼가
// SPEC-ID word-boundary 매칭을 정확히 수행하는지 검증한다.
//
// 5개 sub-case:
//   - C1: 정확한 매칭 (plan)
//   - C2: NAMESPACE substring 노이즈 (false-positive 차단)
//   - C3: 정확한 매칭 (sync)
//   - C4: chore-post token (SPEC- prefix 없음)
//   - C5: closeout body (SPEC- prefix 없이 HARNESS-001만 언급)
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
