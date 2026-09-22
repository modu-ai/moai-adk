# t449 판정서 — `moai integration` 락이 호출자 cwd를 기록하는 결함 수리

- 판정: **FIXED (카드 채택·구현 완료)** — 전제는 반증되지 않았다(리드 4회 관측 = 본 트리 코드 `internal/cli/integration.go` 구 acquire 경로와 정확히 일치).
- 범위 준수: 락 직렬화 의미론·스테일 판정·PreToolUse deny 계층·플래그 표면·`integrationLockRoot`·`integrationSessionID`·`currentBranch` **무변경**. 기록되는 값과 status 표시만 수리.

## 변경 (이름 나열)

- `internal/cli/integration.go` — 신규 `resolveIntegrationTarget(explicitBranch, configuredBranch)`(해석 순서: `--branch` 플래그 → 설정된 git-flow develop 브랜치 → 호출자 폴백) + 신규 `worktreeForBranch(branch)`(`git worktree list --porcelain` 블록 파서, `filepath.FromSlash`로 Windows 안전, 못 찾으면 빈 문자열 — "잘못된 경로보다 정직한 미확정"). acquire RunE를 이 둘로 재배선. status 사람 출력이 `SessionName` 있을 때 `holder: <이름> (<id>, pid <pid>)`로 표시(레인이 자기 순번을 분간할 수 있게), 없으면 기존 형태. `--branch` 도움말 갱신.
- `internal/config/loader_integration_branch.go` — 신규 `LoadGitFlowDevelopBranch(projectRoot)`: fail-open 단일키 판독기. `mode: manual` + 활성 프로필 `workflow: git-flow` + 키 비어있지 않음 3조건을 모두 통과할 때만 값을 돌린다. 선례: `internal/config/loader_worktree_base.go:28` `LoadWorktreeBaseBranch`(같은 loadYAMLFile·빈 래퍼·실패 시 빈값 형태), git-flow 판별은 `internal/cli/hook_pre_push.go:71` `ActiveModeProfile()` 접근 관용.
- `internal/config/types.go` — 주석 전용: "NO Go consumer" 블록에서 `DevelopBranch` 면제(소비자 생김, 카드 t449).
- 테스트 신규 2파일 — `internal/cli/integration_target_test.go`(11개: porcelain 발견/미발견, 해석 순서 4케이스, e2e acquire 기록 4케이스, status 이름 표시 2케이스 — fixture 브랜치명 `fixture-integration`으로 하드코딩 부재 증명) + `internal/config/loader_integration_branch_test.go`(게이트 6케이스, github-flow·personal-profile 거절 포함). 전부 `t.TempDir()` 스크래치 리포.

## 5-섹션 증거

**Claim**: acquire는 `--branch` 부재 시 설정된 git-flow develop 브랜치와 그 브랜치가 실제 체크아웃된 워크트리를 기록하고, status는 홀더 이름을 보여준다. 호출자 cwd는 설정이 없어 호출자 트리가 실제 통합 대상인 마지막 폴백에서만 기록된다.

**Evidence** (본 런, 본 워크트리, 레인 독립 재실행):
- `go vet ./internal/cli/ ./internal/config/` → exit 0
- `go build ./...` → exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go test ./internal/cli/ -run 'Integration' -count=1` → `ok ... 18.477s`
- `go test ./internal/config/ -run 'TestLoadGitFlowDevelopBranch' -count=1` → `ok ... 0.282s`
- RED(구현 전, manager-develop 관측): 신규 테스트 6건 undefined 빌드 실패 + mid-GREEN red 1건(personal-profile이 git-flow 선언으로 게이트 통과 → `Mode == "manual"` 명시 조건으로 봉쇄) — 가드가 실제로 물리는 것을 관측함
- diff 스코프 스윕: `git diff --stat` = integration.go +types.go 주석 전용 — 금지 표면(의미론·deny·플래그) 0건

**Baseline-attribution**: 본 워크트리(`.claude/worktrees/t449`, 분기 `WT-integration-lock-record`), base = 로컬 develop `5a8449859` fast-forward 흡수(f7cabfc29→5a8449859), 2026-09-03 본 런.

**Gaps**:
- `internal/cli` 전체 스위트 미실행(600s 하한 규율 — 스코프 `-run 'Integration'`으로 대체; 전 판정은 CI 몫)
- 실제 공유 리포 위한 e2e acquire 스모크 미수행(공유 `.moai/state`에 실 락 기록 부수效应 — 스크래치 리포 e2e 테스트 4케이스로 대체)
- 상속 적색 1건 미수리(아래 Residual-risk)

**Residual-risk**:
- **상속 적색(본 카드 무관)**: `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget' -count=1` → **FAIL, base `5a8449859`에서도 재현 귀속**. 본 카드 diff 5파일은 Go 2 + 테스트 3, `.md` 0건 — 이 테스트(CLAUDE.md·rules 트리 바이트 측정) 표면과 무교집합. 리드 통지용으로 이름 밝힘.
- 워크트리 경로는 git 정규형(심볼릭 링크 resolved, 슬래시)으로 기록 — 심볼릭된 임시 디렉터에서 `/private/var` vs `/var` 모양 차이 가능. 실제 통합 트리(`.claude/worktrees/develop`)는 심볼릭 아님.
- 설정 전제: `git-strategy.yaml`에 `develop_branch`가 채워져 있어야 신 기본값이 발동 — 비어 있으면 기존(호출자 폴백) 동작. 본 리포의 설정은 이미 채워짐.
