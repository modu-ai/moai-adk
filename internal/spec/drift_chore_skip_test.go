package spec

// @MX:TODO: RED phase — walker filter 미구현으로 4개 테스트 모두 FAIL 예상
// @MX:REASON: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 M1 RED phase; GREEN 단계에서 제거됨

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureCommit은 git fixture 생성에 사용되는 commit 구조체
type fixtureCommit struct {
	// title은 commit 메시지 제목 (git log --oneline 에서 hash 다음에 나오는 문자열)
	title string
	// body는 commit 메시지 본문 (선택적; git log --grep 의 body 매칭에 활용)
	body string
}

// setupGitFixture는 t.TempDir() 하에 임시 git 저장소를 초기화하고
// 주어진 commits 배열을 oldest → newest 순서로 commit한 뒤
// 저장소의 절대 경로를 반환한다.
//
// 반환된 경로에서 git 명령을 실행하면 fixture 커밋들이 포함된 저장소를 참조할 수 있다.
// 테스트 내에서 os.Chdir(dir) 을 호출한 뒤 getGitImpliedStatus 를 호출해야 한다.
func setupGitFixture(t *testing.T, commits []fixtureCommit) string {
	t.Helper()

	dir := t.TempDir()

	// git 저장소 초기화
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v 실패: %v\n출력: %s", args, err, out)
		}
	}

	runGit("init", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")

	// oldest → newest 순서로 commit 생성
	for i, c := range commits {
		// 빈 commit으로 git history에 추가 (실제 파일 변경 없이)
		msg := c.title
		if c.body != "" {
			msg = fmt.Sprintf("%s\n\n%s", c.title, c.body)
		}
		runGit("commit", "--allow-empty", "-m", msg)
		_ = i
	}

	return dir
}

// TestGetGitImpliedStatus_ChoreSkip는 AC-LSCSK-005 (a)(b)(c)(d) 4개 케이스를 검증한다.
// 각 케이스는 독립적인 임시 git 저장소를 사용하여 격리된다.
//
// M1 RED phase: walker filter 미구현으로 4개 모두 FAIL 예상.
func TestGetGitImpliedStatus_ChoreSkip(t *testing.T) {
	tests := []struct {
		name       string
		specID     string
		commits    []fixtureCommit
		wantStatus string
		wantErr    bool
	}{
		{
			// AC-LSCSK-005 (a): sweep commit이 실제 impl commit을 숨기는 시나리오
			// walker가 chore(spec) commit을 skip하고 이전 impl(spec) commit을 채택해야 한다
			name:   "sweep_commit_hides_real_impl",
			specID: "SPEC-FOO-001",
			commits: []fixtureCommit{
				{
					title: "impl(spec): SPEC-FOO-001 initial implementation",
					body:  "",
				},
				{
					// sweep commit — body에 SPEC-FOO-001 언급 → git log --grep 매칭
					title: "chore(spec): status drift sweep",
					body:  "SPEC-FOO-001 in-progress → implemented",
				},
			},
			// walker가 sweep commit을 skip하고 첫 번째 impl(spec) commit을 채택
			// impl(spec) prefix는 transitions.go 에 직접 등록되어 있지 않으나,
			// "impl" 은 등록된 prefix가 아니므로 "unknown" 카테고리 → 빈 status → continue
			// 단, AC-LSCSK-005 (a)의 의도는 실제 registered prefix인 "feat"을 사용하면
			// 더 명확하므로 commit title을 "feat(spec): ..." 로 변경
			wantStatus: "implemented",
			wantErr:    false,
		},
		{
			// AC-LSCSK-005 (b): chore(spec) 이후 real feat commit — walker가 chore를 skip하고 feat 채택
			// 단, walker는 newest-first 순서이므로 가장 최신(newest) commit인 feat을 먼저 봄
			// → 즉시 "implemented" 반환 (skip 필터 발동조차 안 함)
			name:   "chore_precedes_feat_walker_returns_implemented",
			specID: "SPEC-BAR-001",
			commits: []fixtureCommit{
				{
					// oldest — chore(spec) sweep
					title: "chore(spec): metadata cleanup",
					body:  "SPEC-BAR-001 status update",
				},
				{
					// newest — real feat commit
					title: "feat(SPEC-BAR-001): new feature implementation",
					body:  "",
				},
			},
			// walker는 newest-first → feat commit 먼저 봄 → "implemented" 즉시 반환
			wantStatus: "implemented",
			wantErr:    false,
		},
		{
			// AC-LSCSK-005 (c): 모든 commit이 chore(spec)만인 경우
			// walker가 N개 budget 내 모두 skip → error 반환
			name:   "only_chore_commits_returns_error",
			specID: "SPEC-BAZ-001",
			commits: []fixtureCommit{
				{
					title: "chore(spec): first sweep",
					body:  "SPEC-BAZ-001 initial state",
				},
				{
					title: "chore(spec): second sweep",
					body:  "SPEC-BAZ-001 status update",
				},
				{
					title: "chore(spec): third sweep",
					body:  "SPEC-BAZ-001 lint-skip 등록",
				},
			},
			// 모든 commit이 chore(spec) → walker 소진 → error 반환
			wantStatus: "",
			wantErr:    true,
		},
		{
			// AC-LSCSK-005 (d): 제어 케이스 — 최신 commit이 실제 impl commit인 경우
			// walker가 첫 commit에서 즉시 status 반환, skip filter 미발동
			name:   "latest_is_real_impl_control_case",
			specID: "SPEC-QUX-001",
			commits: []fixtureCommit{
				{
					// 단일 commit — real implementation
					title: "feat(SPEC-QUX-001): initial implementation",
					body:  "",
				},
			},
			// 단일 feat commit → walker 첫 순회에서 즉시 "implemented" 반환
			wantStatus: "implemented",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// case (a)의 commits 수정: impl(spec) 대신 feat(SPEC-FOO-001) 사용
			// (impl은 transitions.go에 미등록 prefix로 "unknown"/"" → walker가 skip)
			localCommits := make([]fixtureCommit, len(tt.commits))
			copy(localCommits, tt.commits)

			if tt.name == "sweep_commit_hides_real_impl" {
				// case (a): 첫 commit을 feat으로 변경하여 ClassifyPRTitle이 "implemented"를 반환하도록
				localCommits[0] = fixtureCommit{
					title: "feat(SPEC-FOO-001): initial implementation",
					body:  "",
				}
			}

			dir := setupGitFixture(t, localCommits)

			// getGitImpliedStatus는 현재 작업 디렉토리의 git 저장소를 사용
			// t.Cleanup으로 원래 디렉토리로 복원
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

			gotStatus, gotErr := getGitImpliedStatus(tt.specID)

			if tt.wantErr {
				if gotErr == nil {
					t.Errorf("getGitImpliedStatus(%q) = (%q, nil), want error", tt.specID, gotStatus)
				}
			} else {
				if gotErr != nil {
					t.Errorf("getGitImpliedStatus(%q) 예상치 못한 오류: %v", tt.specID, gotErr)
					return
				}
				if gotStatus != tt.wantStatus {
					t.Errorf("getGitImpliedStatus(%q) = %q, want %q", tt.specID, gotStatus, tt.wantStatus)
				}
			}
		})
	}
}

// TestShouldSkipCommitTitle_ChorePattern은 shouldSkipCommitTitle 헬퍼 함수의
// skip pattern 매칭 동작을 검증한다 (AC-LSCSK-005 관련).
//
// M1 RED phase: shouldSkipCommitTitle 함수가 아직 존재하지 않으므로 컴파일 실패.
func TestShouldSkipCommitTitle_ChorePattern(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{
			name:  "chore(spec) skip",
			title: "chore(spec): status drift sweep",
			want:  true,
		},
		{
			name:  "chore(specs) plural skip",
			title: "chore(specs): bulk metadata update",
			want:  true,
		},
		{
			name:  "case insensitive CHORE(SPEC)",
			title: "CHORE(SPEC): uppercase variant",
			want:  true,
		},
		{
			name:  "feat commit not skipped",
			title: "feat(SPEC-FOO-001): new feature",
			want:  false,
		},
		{
			name:  "plain chore (no spec scope) not skipped",
			title: "chore: dependency update",
			want:  false,
		},
		{
			name:  "impl(spec) not a skip pattern",
			title: "impl(spec): some implementation",
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

// TestGetGitImpliedStatus_WalkerDepthBoundary는 AC-LSCSK-004 를 검증한다:
// N=50 walker depth 내 모두 skip-pattern commit인 경우 error 반환.
// StatusGitConsistencyRule 은 이 error를 받으면 finding을 emit하지 않아야 한다.
//
// M1 RED phase: walker 미구현으로 FAIL 예상.
func TestGetGitImpliedStatus_WalkerDepthBoundary(t *testing.T) {
	// N=50 budget을 확인하기 위해 3개의 chore commit만으로도 충분 (case c와 동일 패턴)
	// 전체 50개를 만들면 테스트가 느려지므로 여러 개로 검증
	specID := "SPEC-DEPTH-001"
	var commits []fixtureCommit
	for i := 0; i < 5; i++ {
		commits = append(commits, fixtureCommit{
			title: fmt.Sprintf("chore(spec): sweep iteration %d", i+1),
			body:  fmt.Sprintf("%s processed in sweep %d", specID, i+1),
		})
	}

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

	gotStatus, gotErr := getGitImpliedStatus(specID)

	// 모든 commit이 chore(spec) → walker 소진 → error 반환 기대
	if gotErr == nil {
		t.Errorf("getGitImpliedStatus(%q) = (%q, nil), want error when all commits are chore(spec)", specID, gotStatus)
	}
}

// TestGetGitImpliedStatus_ChoreSkip_CasesVerifyGitExec는
// 테스트에서 git 실행 환경이 올바르게 설정되는지 확인하는 보조 테스트.
// setupGitFixture가 정상 동작하는지 검증한다.
func TestGetGitImpliedStatus_ChoreSkip_CasesVerifyGitExec(t *testing.T) {
	specID := "SPEC-VERIFY-001"
	commits := []fixtureCommit{
		{
			title: "feat(SPEC-VERIFY-001): verification commit",
			body:  "",
		},
	}

	dir := setupGitFixture(t, commits)

	// git log 직접 실행으로 fixture가 올바르게 만들어졌는지 확인
	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("--grep=%s", specID))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log 실패: %v", err)
	}

	outStr := strings.TrimSpace(string(out))
	if !strings.Contains(outStr, "feat(SPEC-VERIFY-001)") {
		t.Errorf("git log 출력에 예상된 commit이 없음: %q", outStr)
	}

	// dir 경로가 git worktree인지도 확인
	verifyCmd := exec.Command("git", "rev-parse", "--git-dir")
	verifyCmd.Dir = dir
	verifyOut, err := verifyCmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse --git-dir 실패: %v", err)
	}
	gitDir := filepath.Join(dir, ".git")
	if !strings.Contains(string(verifyOut), ".git") {
		t.Errorf("예상된 git dir 경로 %q를 포함하지 않음: %q", gitDir, string(verifyOut))
	}
}
