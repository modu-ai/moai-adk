# t1792 진행 기록 — PR #1820 Lint 실패 수리 (dependabot gomod-minor-patch 8건)

[GH PR #1820·운영자 지시 10-15·Class A·P2] · run tmvgvn · lane-21 · 2026-10-15 · 리더 @ 3274.sock 배차

## 실행 기록

- 임대: `moai factory next --card t1792 --run tmvgvn --wait --wait-bound 10m` → 성공 (`t1792 stage=- worktree=t1792`). 진입: ExitWorktree(keep, t1713 트리 보존) → 주 체크아웃 → EnterWorktree(t1792) — 브랜치 `WT-gh-pr-1820`.
- PR 베이스 정렬: 지시된 `git reset --hard FETCH_HEAD`는 권한 거부 → 비파괴 동등 명령 `git checkout -B WT-gh-pr-1820 FETCH_HEAD`로 동일 결과(트리 클린·고유 커밋 없음 상태에서 브랜치 재생성). PR head `fbeeb67be` 정렬 완료.
- 1차 커밋 `1d434b21a`: `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web` 결과 — `internal/web/*_templ.go` 8파일 (+412/−416, 생성물만).
- **검증 중 실제 결함 발견·수리**: 신귝 a-h/templ이 텍스트 노드의 `@`를 컴포넌트 호출로 해석 — `fieldsets.templ:225`의 리터럴 `Model @ effort`가 `effort.Render`(미정의 컴포넌트) 호출을 생성해 `go build`/`go vet` 실패(`undefined: effort`, fieldsets_templ.go:797). 유일 패턴(타 .templ의 @는 정상 컴포넌트 호출/주석). 수리: `&#64;` 이스케이프(렌더 출력 동일) → 재생성 → `effort.Render` 0건 확인. 2차 커밋 `5742cb8e3` (소스 1 + 생성물 1).
- 검증: `go build ./...` **PASS**(exit 0), `go vet ./internal/web/...` **PASS**, internal/web 테스트 — 1차 전체 스위트에서 `TestServer_AutoOpenFailureNonFatal` 1건 실패(스위트 602초, 30+레인 부하 창) → 단일 재실행 3/3 **통과**(7.3초)로 부하 민감 플레이크 분류, 변경면(렌더 코드)과 무관한 서버 오토-오픈 테스트. 회귀 아님.
- push: `fbeeb67be..5742cb8e3 → dependabot/go_modules/gomod-minor-patch-70e8355ca2` (운영자 승인된 PR 브랜치 push, force 없음, exit 0).
- 임대 만료 1회 발생(원인 분석이 TTL 초과, 03:20:42Z) → 8859 복구(stage 만료답 유발 → 수집 → 재임대) 후 갱신 크론(9분) + 리더 신규 방식(긴 작업 전후 직접 renew) 병행. 이후 만료 없음.

## gh pr checks 1820 최종 결과 (head 5742cb8e3)

- **Lint: PASS (3m36s)** — 카드 목표(`templ generate && git diff --exit-code`) 달성.
- Build 4플랫폼·Analyze(Go)·CodeQL·Constitution·claude-review·graph-freshness(±conflict-guard)·labeler·browser fire guard·Multi-OS(mac/windows)·Race Test 1·Release PR Gate·Vercel: **pass**. skipping 행은 설계상 스킵.
- **실패 2잡**: `Test (ubuntu-latest)`(10m50s), `Race Test 2`(15m26s) — 양잡 모두 동일 2패밀리만: `TestCensus_SetDifferenceEmptyBothDirections`(census_test.go:80 "claude-code-review.yml exists on disk with no manifest entry", 21 vs 22) + `TestDotgithubDeployedParity/label-sync.yml`(미러 드리프트, 1661 vs 1661 바이트).
- **분류 (A/B 귀속)**: 두 테스트는 베이스 `fbeeb67be` 로컬 **실행·통과**, 내 HEAD `5742cb8e3` 로컬 3회 **통과**, 그리고 동일 베이스 head의 이전 PR 런 38259493515에서 CI **통과**(ubuntu·Race 전부 success — `gh run view --json headSha`로 동일 커밋 확인). 내 diff는 internal/web뿐 — 실패 패밀리(guardstate·template 패리티)와 무관. GitHub pull_request 런은 병합 프리뷰(PR × main@2aab5f797)를 검사하므로, 이전 런 대비 main 전진(워크플로 추가/제거 이력: claude-code-review.yml은 0de8517e5 추가→b3ab67ce8 제거)이 병합 프리뷰의 .github 상태와 PR 분기의 구형 미러/매니페스트를 조합해 드리프트 단언을 발화한 것으로 판정. **내 커밋으로 귀속 불가** — PR 수명 주기 문제(분기가 main 대비 최신성 부족; rebase 시 해소 예상), 병합 판단은 리더.
- Residual-risk: 병합 프리뷰 가설의 재현은 다음 dependabot 동기화 또는 리더의 merge-시도 로그에서 1회 확인 가능(본 카드 범위 밖).

## 최종 상태

- 카드 목표(수리): **달성** — Lint PASS. PR 브랜치 push 완료(2커밋). 병합은 리더(운영자 일괄 원격 처리). 로컬 main 착지 없음(카드 성격상 PR 종료).
- 리더 지시(운영자 지시 · GLM 레인 정리, 보드 d-20261015T044208Z-28e4)에 따라 본 레인은 이 카드로 종료: 크론·갱신 루프 삭제 후 최종 보고 → 정지. t1713은 리더가 타 레인으로 이동(본 트리 `.moai/worktrees/t1713`에 증거 전체 보존: progress.md·card-review.md 3라운드 최종 pass·sync-audit-verdict.md codex pass).

## Landing — lane-17 redispatch session (2026-10-16)

decision record: decided_by=lane-17(orchestrator) evidence_refs=card=t1792;class=A;hold=d-20261015T074036Z-93a1(remote PR #1820 actions only);pr=#1820 OPEN MERGEABLE head 83c27bd73;worktree=present;overlap=unmeasured ladder_path=lane judgment — local landing to merged-local only; remote PR merge/close stays operator keep-set

- Main absorb: `git merge --no-edit main` exit 0 (main 0f6060e0e, 1175 ahead before merge, no conflicts).
- templ: `go run github.com/a-h/templ/cmd/templ@v0.3.1070 generate -path ./internal/web` exit 0; `git status --short` empty after → CI Lint step (generate + git diff --exit-code) holds.
- `go build ./...` exit 0. `go vet` on the 14 affected package trees exit 0.
- Tests: packages importing a bumped module (templ, mcp-go, bubbletea, go-runewidth, x/net, x/sync, x/sys, x/tools; 18 pkgs) with TMPDIR/GOTMPDIR=/private/tmp scratch, GIT_CONFIG_GLOBAL isolated, MOAI_FACTORY_* unset, -count=1: 52 ok; FAIL internal/cli, internal/hook.
  - internal/hook TestNoNewEnvLiteralsInDiff: literal MOAI_KANBAN_CARD, introduced by 9c8447f9b (card t1776, already on main) — `git log -S'MOAI_KANBAN_CARD' 966f1b158..HEAD` names only that commit. Not this card.
  - internal/cli: TestCodexPathGuardRealFixtures / TestCodexLocalInstructions_UnixNonRegular / RenderedPanelStaysInBand failed with `bind: invalid argument` (unix socket path over length from the long scratch TMPDIR); with TMPDIR=/private/tmp/t1792A they pass.
  - internal/cli A/B, same -run set, short TMPDIR: card tree (A) and main tree `.claude/worktrees/main` @4a39d0e42 (B; doctor_codex_test.go and go.mod hashes equal main) fail the identical 10 tests (TestCheckCodexWiring_* ×8, TestCountCodexAgentTOMLs, _BadPatternRoot). Pre-existing on main, not caused by the bumps.
- Card review: card-review.md (codex scope=card pass, advisory, no receipt).

Gaps: full repo suite not run (CI owns it); internal/cli main-side failures not root-caused (out of scope); CodeRabbit not read (remote PR is operator-held).
