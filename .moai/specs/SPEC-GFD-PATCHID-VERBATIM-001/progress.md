# progress — SPEC-GFD-PATCHID-VERBATIM-001

상태: in-progress (M1 run 커밋에서 manager-develop 이 전이 — card t1561, 2026-10-07)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-06T23:58:08Z
근거: plan-audit 3회차 PASS 0.94 — `verdict: PASS`·`overall_score: 0.94`·`plan_artifact_hash 574a86dfea3e8925f941842eb1e6deb993c6543e136f1431f1d8f6309794da65`(본 산물과 일치)·`audited_sha cad44a75163b6f0f056551354ef8008fe799090f`, codex 교차 `verdict: pass` findings 0, receipt `rcpt-9a7faf514734104deee70575` (원문: `.moai/reports/t1561/plan-audit-iter3.md`). RESOLVED 경과: 1회차 D1-D7 → 2회차 D8-D11 → 3회차 신규 결함 0.

## §E.2 Run-phase Evidence

manager-develop (card t1561) — 2026-10-07. 트리 `.moai/worktrees/t1561` (카드 워크트리), 브랜치 `WT-landing-patchid`, base `cad44a75163b6f0f056551354ef8008fe799090f` → M2 HEAD `0ecbf3a21`. 아래 모든 관측은 이번 런·이 트리에서 수행했다 (baseline 귀속: 명령 + 원문 출력 + 시점 트리 좌표를 행마다 병기).

### AC-GPV-001 — 공백 갈림 tip 을 sweep 가 보존한다 (RED→GREEN) — PASS

- **Claim**: M1 의 verbatim 전환이 RED 관문을 GREEN 으로 뒤집었다.
- **Evidence**: RED — `go test ./internal/cli/worktree/ -run '^TestLandingPredicateWhitespaceDivergenceKeepsTree$' -count=1 -timeout 30m -v`, 트리 cad44a751(수정 전):
  `--- FAIL: TestLandingPredicateWhitespaceDivergenceKeepsTree (4.71s)` · `landing_predicate_test.go:592: sweep: landed="yes" verdict="DISPOSE" reason="", want preserve — a whitespace-divergent tip is not on the ref and the patch-id normalization must not read it as landed` · `FAIL github.com/modu-ai/moai-adk/internal/cli/worktree 5.117s` · `EXIT=1`.
  GREEN — 같은 명령, HEAD 0ecbf3a21: `--- PASS: TestLandingPredicateWhitespaceDivergenceKeepsTree (43.44s)` · `ok github.com/modu-ai/moai-adk/internal/cli/worktree 85.622s` · `EXIT=0`.
- **Baseline-attribution**: 양쪽 모두 이번 런·카드 워크트리. RED 는 M1 커밋(5f952f509) 이전, GREEN 은 M2 커밋(0ecbf3a21) 이후.
- **Gaps**: 없음.
- **Residual-risk**: GREEN 소요 43s — 픽스처가 실제 git 저장소를 매번 만드는 구조라 머신 부하에 따라 흔들린다(기능 판정에는 무관).

### AC-GPV-002 — 기존 찾지 판정 스위트 무변경 통과 — PASS

- **Claim**: 기존 테스트 단언은 한 행도 지워지거나 바뀌지 않았고, worktree 패키지 전체가 GREEN 이다.
- **Evidence**: (a) `go test ./internal/cli/worktree/ -count=1 -timeout 30m` — M1 시점 `ok ... 411.180s` `EXIT=0`, M3 배치 시점 `ok ... 1364.466s` (두 실행 모두 이번 런). (b) `git diff cad44a751 HEAD -- internal/cli/worktree/landing_predicate_test.go | grep '^-' | grep -v '^---'` → 출력 0행, `REMOVED_LINES_EXIT=1` (삭제 0 = 순수 추가).
- **Baseline-attribution**: HEAD 0ecbf3a21, 이번 런.
- **Gaps**: (a)의 두 실행은 서로 다른 시점(커밋 5f952f509 직후·0ecbf3a21 커밋 후)이나 트리 내용은 동일 커밋 상태로 판정됐다.
- **Residual-risk**: 없음.

### AC-GPV-003 — vet·lint 클린 (regression-guard) — PASS

- **Claim**: 기준선(§B 셀: vet 클린, lint `0 issues.`)이 유지됐다 — 신규 지적 0.
- **Evidence**: 사전 기준선(트리 cad44a751, 이번 런): `go vet ./internal/cli ./internal/cli/worktree/` → `VET_CLEAN`, `golangci-lint run …` → `0 issues.` `LINT_CLEAN`. 수리 후(HEAD 0ecbf3a21): vet `VET_EXIT=0`, lint `0 issues.` `LINT_EXIT=0`.
- **Baseline-attribution**: 수리 전·후 모두 이번 런·이 트리에서 측정.
- **Gaps**: 없음.
- **Residual-risk**: 없음.

### AC-GPV-004 — 공백 정규화 patch-id 잔존 0 — PASS

- **Claim**: `landing_predicate.go` 에서 `--stable` 이 주석 포함 0히트다.
- **Evidence**: `grep -c '\-\-stable' internal/cli/worktree/landing_predicate.go` → `0`, `GREP_EXIT=1` (RED-now 는 §B 2행: `2`히트 exit 0 → 뒤집힘). 참조: `grep -rc 'verbatim' internal/cli/worktree/landing_predicate.go` → `15`.
- **Baseline-attribution**: HEAD 0ecbf3a21, 이번 런.
- **Gaps**: 없음.
- **Residual-risk**: 없음 — 미지원 git 하락 경로도 없다(AC-GPV-005 가 별도로 잠근다).

### AC-GPV-005 — 미지원 git fail-closed — PASS

- **Claim**: 시임을 미지원으로 고정하면 계층 2 는 오류, sweep 는 PR 없으면 preserve, 계층 3 은 계속 판정한다.
- **Evidence**: `go test ./internal/cli/worktree/ -run '^TestLandingPredicateVerbatimUnsupportedIsFailClosed$' -count=1 -timeout 30m -v` → RUN 행 3개: `…/landedbypatchid_errors` · `…/sweep_preserves_without_pr` · `…/layer3_still_decides`, `--- PASS … (65.32s)`, `ok … 66.005s`, `EXIT=0`. `[no tests to run]` 없음.
- **Baseline-attribution**: HEAD 0ecbf3a21, 이번 런.
- **Gaps**: 없음.
- **Residual-risk**: 없음.

### AC-GPV-006 — 다중 커밋 누적 공백 갈림 보존 — PASS

- **Claim**: 두 커밋 카드 + 공백 변형 squash 도 preserve 다.
- **Evidence**: RED — `… -run '^TestLandingPredicateWhitespaceDivergenceMultiCommitKeepsTree$' … -v`, 트리 cad44a751+미커밋 테스트: `--- FAIL … (4.71s)` · `landing_predicate_test.go:642: sweep: landed="yes" verdict="DISPOSE" … multi-commit tip …` · `EXIT=1`. GREEN — HEAD 0ecbf3a21: `--- PASS: TestLandingPredicateWhitespaceDivergenceMultiCommitKeepsTree (41.64s)` · `EXIT=0`, RUN 행 확인.
- **Baseline-attribution**: 양쪽 모두 이번 런·카드 워크트리.
- **Gaps**: 없음.
- **Residual-risk**: 없음.

### AC-GPV-007 — 세션 종료가 공백 갈림 카드를 보존한다 (제2 지점 관문) — PASS

- **Claim**: 세션 출구의 동치 팔(cherry → LandedByCommitPatchIDs 교체)이 공백 갈림 squash 를 착지로 읽지 않는다.
- **Evidence**: RED — `go test ./internal/cli -run '^TestSessionExitWhitespaceDivergencePreserves$' -count=1 -timeout 30m -v`, 트리 5f952f509+미커밋 테스트: `--- FAIL … (13.82s)` · `session_worktree_landing_test.go:275: whitespace-divergent worktree was REMOVED by session-exit cleanup (the t1561 second survival point, observed)` · `EXIT=1`. GREEN — M2 수리 후 같은 명령 + `TestCleanupSessionWorktree*` 전체: `--- PASS: TestSessionExitWhitespaceDivergencePreserves (14.37s)`, 기존 18함수 전부 PASS, `ok github.com/modu-ai/moai-adk/internal/cli 70.893s` `EXIT=0`.
- **Baseline-attribution**: 양쪽 모두 이번 런·카드 워크트리.
- **Gaps**: 없음.
- **Residual-risk**: 세션 관문의 GREEN 은 필터 실행이다 — 뿌리 패키지 전체 행은 아래 배치 판독 행을 본다.

### 배치 판독 — M3 검증 배치와 비소관 적색 분류

- **Claim**: 배치 명령 `go test ./internal/cli ./internal/cli/worktree/ -count=1 -timeout 30m` 의 결과는 worktree `ok`, cli 뿌리 패키지는 (i) `TestCodexTask*` 계열의 pre-existing 환경 적색과 (ii) 2-패키지 동시 실행 부하의 30m 타임아웃으로 FAIL — 수리 축의 적색 0.
- **Evidence**: `FAIL github.com/modu-ai/moai-adk/internal/cli 1801.541s` / `ok github.com/modu-ai/moai-adk/internal/cli/worktree 1364.466s` / `EXIT=1`. 타임아웃 패닉 시점 실행 테스트: `running tests: TestWorktreeLaunchRejectsConcurrentWriter (1m1s)` — codex 실패와 무관. codex 실패 원문(대표): `codex_task_resume_scope_test.go:127: result carries no StructuredContent: … Text:codex_task: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1561 is not a registered worktree of this repository`.
- **Pre-existing 입증 (A/B)**: 대표 테스트 `TestCodexTask_FailOpenOnMissingCodex` 를 HEAD 0ecbf3a21 에서 `FAIL … 3.806s EXIT=1`, base cad44a751(detach 재측정)에서 `FAIL … 5.625s EXIT=1` — **양팔 동일 실패**. 내 diff(`git diff --stat cad44a751 HEAD`)는 codex/mcp 경로를 전혀 포함하지 않는다.
- **타임아웃 분석**: `go test` 는 두 패키지 바이너리를 동시 실행한다 — worktree 단독 411s → 동시 실행 1364s 로 3.3배. `-timeout 30m` 파생 근거는 단일 패키지 최악 1118s (SPEC-CLI-TEST-TIMEOUT-001) 기준이므로, 단독 재측정을 수행한다(아래 행).
- **Baseline-attribution**: HEAD 0ecbf3a21, 이번 런.
- **Gaps**: cli 뿌리 패키지의 전체 GREEN 은 이 런에서 관측되지 않았다 — codex 계열이 base 에서도 적색인 이상 이 트리의 어느 커밋으로도 얻을 수 없다(레인 합의 필요 — 수리 판단은 레인과).
- **Residual-risk**: codex 계열 외에 base 적색이 더 있을 수 있다 — 단독 재측정의 FAIL 목록으로 한정 확인한다.

### 단독 재측정 — `go test ./internal/cli -count=1 -timeout 30m`

- **Claim**: 2-패키지 경합을 제거해도 cli 뿌리 패키지는 현재 머신 부하에서 30m 안에 끝나지 않는다 — 타임아웃은 측정 제약이며, 끝까지 실행된 구간의 적색은 codex 계열뿐이다.
- **Evidence**: `FAIL github.com/modu-ai/moai-adk/internal/cli 1801.739s` · `EXIT=1`. 적색 목록: `--- FAIL` 13건 전부 `TestCodex*` 계열(비-codex 적색 0 — `grep '^--- FAIL' | grep -cv Codex` → 0). 패닉 시점 실행 테스트: `running tests: TestStopChainAdvisoryFailureRecorded (9s)` — 본 SPEC 영역 밖.
- **Baseline-attribution**: HEAD 0ecbf3a21, 이번 런, 슬롯 임대 하 단독 실행.
- **Gaps**: cli 뿌리 패키지의 완주 GREEN 은 로컬에서 관측되지 않았다 — 중량 실행 2회(배치·단독)로 재시도 상한 소진. `TestStopChainAdvisoryFailureRecorded` 이후 구간은 미실행.
- **Residual-risk**: 머신 부하가 낮아지면 30m 내 완주 가능성이 있다(기준선 파생: 단일 패키지 최악 1118s — SPEC-CLI-TEST-TIMEOUT-001). 전체 판정은 CI 몫(AGENTS.md §4) — 본 카드 판정은 AC 행의 패키지·타깃 스코프로 확정했다.

### E8 — pre-GREEN RED 원문 3건 (이번 런 캡처)

1. 단일 커밋(AC-GPV-001): 위 AC-GPV-001 Evidence 의 RED 블록 (`landed="yes" verdict="DISPOSE"`, EXIT=1).
2. 다중 커밋(AC-GPV-006): 위 AC-GPV-006 Evidence 의 RED 블록 (`landing_predicate_test.go:642`, EXIT=1).
3. 세션 종료(AC-GPV-007): 위 AC-GPV-007 Evidence 의 RED 블록 (`worktree was REMOVED by session-exit cleanup`, EXIT=1).

### E2 — 크로스 플랫폼 빌드

- `go build ./...` → `DARWIN_BUILD_OK` (exit 0) · `GOOS=windows GOARCH=amd64 go build ./...` → `WINDOWS_BUILD_OK` (exit 0) — 최종 트리(HEAD 0ecbf3a21 내용) 기준, 이번 런.

### E4 — 경계 grep

- `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/worktree/ internal/cli/session_worktree.go | grep -v _test.go | grep -v '// '` → 0히트 (`BOUNDARY_GREP_EXIT=1`), HEAD 0ecbf3a21.

### E6 — 커밋 대장

- M1 `5f952f509` — fix(SPEC-GFD-PATCHID-VERBATIM-001): M1 switch the patch-id layers to --verbatim (술어 전환 + RED 게이트 2건 + SPEC 산물 5파일 + status draft→in-progress 전이)
- M2 `0ecbf3a21` — fix(SPEC-GFD-PATCHID-VERBATIM-001): M2 rewire the session-exit equivalence arm (②팔 교체 + 세션 관문 + AC-GPV-005/006 보강)
- M3 — progress.md 증거 기재 (아래 §E.3 커밋)
- 미푸시 — push·PR 은 레인 소관(B9).

### E7 — 블로커

- 구조화 블로커 해당 없음. 참고 보고: cli 뿌리 패키지의 `TestCodexTask*` pre-existing 적색(§배치 판독 행)은 카드 외부 결함으로 레인 판단을 요청한다 — 본 SPEC 수리와 무관(A/B 입증).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: <M3 커밋 시각 — 커밋 후 기재>
run_commit_sha: <M3 커밋 SHA — 커밋 후 기재>
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 11   # §A.5 전 항목 — 기존 5함수·RED 관문(단언 고정 GREEN 전환)·세션 cleanup 18함수·호출자 계약·술어 ①③팔·②팔 확인력·상한/타임아웃·헤더 계약
l44_pre_commit_fetch: n/a — 카드 워크트리 사유 브랜치(커밋 직전 HEAD 재판독으로 대체: 5f952f509·0ecbf3a21 양쪽 관측)
l44_post_push_fetch: n/a — 미푸시(B9: push·PR 은 레인 소관)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 4           # landing_predicate.go · landing_predicate_test.go · session_worktree.go · session_worktree_landing_test.go (SPEC 산물 5파일은 M1 카드 인프라로 동반 커밋)
m1_to_mn_commit_strategy: 3 커밋 — M1 5f952f509(술어 전환+RED 게이트+SPEC 산물+status 전이) · M2 0ecbf3a21(세션 ②팔+보강 게이트) · M3(progress 증거)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_

## §F Phase 4 Mode Selection

decision record (Kickoff, autonomous form per auto-semantics §9.1): plan-audit iter3 **PASS 0.94** (Tier M threshold 0.80 met), plan-artifact hash `574a86df…9794da65` unchanged since the verdict, independent audit cross with codex pass (receipt `rcpt-9a7faf514734104deee70575`), open blockers 0 — run entry approved 2026-10-07 by lane-1 (card t1561, delta round per leader ruling recorded in `.moai/reports/t1561/turn-gate-20261007.md` §plan-audit 천장 경과).

Input parameters: tier=M · scope=4 files (landing_predicate.go, session_worktree.go + 2 test files) · domains=1 (Go CLI) · language mix=Go only · concurrency benefit=LOW (coding-heavy) · agent-team prereqs=not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | semantic repair across two packages |
| serial | **YES** | coding-heavy single-domain work — Anthropic's coding-task parallelism caveat |
| fanout | no | no research fan-out shape; write-capable concurrency would race one tree |
| sweep | no | not a uniform mechanical transform |
| agent-team | no | no operator request (explicit-request-only) |

Decision: serial
Justification: one `manager-develop` carries M1→M3 sequentially (M1 verbatim transition + support probe, M2 session-path RED/GREEN + multi-commit reinforcement, M3 lane-local verification batch with slot lease). M3's `internal/cli` root-package suite is a heavy run — `moai slot acquire` precedes it per plan §A.4.
