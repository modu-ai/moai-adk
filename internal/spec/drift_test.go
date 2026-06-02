package spec

import (
	"os"
	"testing"
)

// runGetGitImpliedStatusInFixture는 setupGitFixture로 임시 repo를 만들고
// 그 안에서 getGitImpliedStatus를 실행하는 헬퍼다. fixtureCommit /
// setupGitFixture는 drift_chore_skip_test.go에 정의되어 있다 (같은 패키지).
//
// commits는 oldest→newest 순서로 전달한다 (setupGitFixture 규약).
func runGetGitImpliedStatusInFixture(t *testing.T, specID string, commits []fixtureCommit) (string, error) {
	t.Helper()

	dir := setupGitFixture(t, commits)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("현재 디렉토리 확인 실패: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("원래 디렉토리 복원 실패 (무시): %v", err)
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("임시 디렉토리로 이동 실패: %v", err)
	}

	return getGitImpliedStatus(specID)
}

// TestShouldSkipCommitTitle_BackfillChore는 AC-DCA-003 (narrow backfill-skip) 검증.
// SPEC-ID-scoped backfill chore는 skip되어야 하고, 기존 chore(spec):/chore(specs):
// metadata-sweep skip은 그대로 유지되어야 한다 (AC-LSCSK-003 보존, REQ-DCA-006).
// close-infix를 가진 commit은 skip되면 안 된다 (D5 guard, REQ-DCA-005).
func TestShouldSkipCommitTitle_BackfillChore(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{
			// 신규: SPEC-ID-scoped backfill chore는 skip (walker가 진짜 close commit으로 진행)
			name:  "SPEC-ID-scoped backfill chore는 skip",
			title: "chore(SPEC-EXAMPLE-001): backfill §E.2/§E.5 commit SHA",
			want:  true,
		},
		{
			name:  "SPEC-ID-scoped L60 mx_commit_sha backfill도 skip",
			title: "chore(SPEC-V3R6-CI-FLAKY-STABILIZE-002): L60 mx_commit_sha backfill — 41868a664",
			want:  true,
		},
		{
			// D5 guard: backfill + close-infix 결합 subject는 skip하면 안 된다
			// (close-infix가 이겨서 completed로 분류되어야 함, REQ-DCA-005 / AP-2)
			name:  "backfill + 4-phase close 결합 subject는 skip 금지 (close-infix가 이김)",
			title: "chore(SPEC-EXAMPLE-001): backfill §E.2 + Mx-phase audit-ready signal + 4-phase close",
			want:  false,
		},
		{
			// AC-LSCSK-003 보존: metadata-sweep chore(spec):는 여전히 skip
			name:  "chore(spec) metadata-sweep은 여전히 skip (AC-LSCSK-003)",
			title: "chore(spec): status drift sweep",
			want:  true,
		},
		{
			name:  "chore(specs) plural metadata-sweep도 여전히 skip",
			title: "chore(specs): bulk metadata update",
			want:  true,
		},
		{
			// SPEC-ID-scoped이지만 backfill marker가 없는 일반 chore는 skip 금지
			// (진짜 partial-work chore — generic chore 규칙으로 처리)
			name:  "backfill marker 없는 SPEC-ID-scoped chore는 skip 금지",
			title: "chore(SPEC-EXAMPLE-001): template internal-content leak cleanup pass 2",
			want:  false,
		},
		{
			// close commit 자체는 skip 금지 (close-infix가 completed로 분류)
			name:  "정규 close commit은 skip 금지",
			title: "chore(SPEC-EXAMPLE-001): Mx-phase audit-ready signal + 4-phase close",
			want:  false,
		},
		{
			name:  "feat commit은 skip 금지",
			title: "feat(SPEC-EXAMPLE-001): new feature",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipCommitTitle(tt.title)
			if got != tt.want {
				t.Errorf("shouldSkipCommitTitle(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}

// TestGetGitImpliedStatus_CloseInfixWinsBeforeSyncDocs는 AC-DCA-004 검증
// (REQ-DCA-004). newest-first walker는 (backfill skip →) 4-phase close commit에서
// completed를 반환해야 하고, 더 오래된 sync `docs` commit (in-progress 규칙)까지
// 내려가면 안 된다.
func TestGetGitImpliedStatus_CloseInfixWinsBeforeSyncDocs(t *testing.T) {
	specID := "SPEC-X-001"
	// oldest → newest 순서
	commits := []fixtureCommit{
		{title: "feat(SPEC-X-001): M1 initial implementation"},
		{title: "docs(SPEC-X-001): sync-phase artifacts"},
		{title: "chore(SPEC-X-001): Mx-phase audit-ready signal + 4-phase close"},
		// newest: backfill chore (skip되어야 함 → 바로 아래 close commit이 적중)
		{title: "chore(SPEC-X-001): backfill §E.2/§E.5 commit SHA"},
	}

	gotStatus, gotErr := runGetGitImpliedStatusInFixture(t, specID, commits)
	if gotErr != nil {
		t.Fatalf("getGitImpliedStatus(%q) 예상치 못한 오류: %v", specID, gotErr)
	}
	if gotStatus != "completed" {
		t.Errorf("getGitImpliedStatus(%q) = %q, want %q (close-infix가 sync docs보다 먼저 이겨야 함)",
			specID, gotStatus, "completed")
	}
}

// TestGetGitImpliedStatus_CloseInfixDirect는 AC-DCA-004 변형:
// newest commit이 backfill 없이 바로 close-infix인 경우 (예: ANTHROPIC-AUDIT-TIER3-001)
// 즉시 completed를 반환해야 한다.
func TestGetGitImpliedStatus_CloseInfixDirect(t *testing.T) {
	specID := "SPEC-X-002"
	commits := []fixtureCommit{
		{title: "feat(SPEC-X-002): M1 implementation"},
		{title: "docs(SPEC-X-002): sync-phase artifacts"},
		// newest: close commit 직접 (backfill 없음)
		{title: "chore(SPEC-X-002): Mx-phase audit-ready signal + 4-phase close"},
	}

	gotStatus, gotErr := runGetGitImpliedStatusInFixture(t, specID, commits)
	if gotErr != nil {
		t.Fatalf("getGitImpliedStatus(%q) 예상치 못한 오류: %v", specID, gotErr)
	}
	if gotStatus != "completed" {
		t.Errorf("getGitImpliedStatus(%q) = %q, want %q", specID, gotStatus, "completed")
	}
}

// TestGetGitImpliedStatus_BackfillNoRegression은 AC-DCA-008 / D4 검증
// (REQ-DCA-010, REQ-DCA-005). close-infix가 전혀 없는 fixture에서 backfill chore를
// skip하면 그 아래의 sync `docs` commit이 노출되어 implemented가 되어야 하고,
// completed를 만들어내면 안 된다 (genuine incomplete-close 보호, AP-2 anti-goal).
func TestGetGitImpliedStatus_BackfillNoRegression(t *testing.T) {
	specID := "SPEC-Y-001"
	commits := []fixtureCommit{
		{title: "feat(SPEC-Y-001): M1 implementation"},
		{title: "docs(SPEC-Y-001): sync-phase artifacts"},
		// newest: backfill chore (skip) → 위 docs commit 노출 → implemented
		// (close-infix가 어디에도 없으므로 completed를 추론하면 안 됨)
		{title: "chore(SPEC-Y-001): backfill §E.2 sync_commit_sha — abc1234"},
	}

	gotStatus, gotErr := runGetGitImpliedStatusInFixture(t, specID, commits)
	if gotErr != nil {
		t.Fatalf("getGitImpliedStatus(%q) 예상치 못한 오류: %v", specID, gotErr)
	}
	if gotStatus != "implemented" {
		t.Errorf("getGitImpliedStatus(%q) = %q, want %q (close-infix 부재 시 backfill-skip이 completed를 발명하면 안 됨)",
			specID, gotStatus, "implemented")
	}
}

// TestGetGitImpliedStatus_CombinedBackfillCloseSubject는 D5 검증
// (REQ-DCA-005 / AP-2). 단일 commit에 backfill과 4-phase close가 결합된 경우
// close-infix가 backfill-skip을 이겨서 completed로 분류되어야 한다.
func TestGetGitImpliedStatus_CombinedBackfillCloseSubject(t *testing.T) {
	specID := "SPEC-Z-001"
	commits := []fixtureCommit{
		{title: "feat(SPEC-Z-001): M1 implementation"},
		{title: "docs(SPEC-Z-001): sync-phase artifacts"},
		// newest: backfill + close-infix 결합 → close-infix가 이김 → completed
		{title: "chore(SPEC-Z-001): backfill §E.2 + Mx-phase audit-ready signal + 4-phase close"},
	}

	gotStatus, gotErr := runGetGitImpliedStatusInFixture(t, specID, commits)
	if gotErr != nil {
		t.Fatalf("getGitImpliedStatus(%q) 예상치 못한 오류: %v", specID, gotErr)
	}
	if gotStatus != "completed" {
		t.Errorf("getGitImpliedStatus(%q) = %q, want %q (결합 subject에서 close-infix가 backfill-skip을 이겨야 함)",
			specID, gotStatus, "completed")
	}
}
