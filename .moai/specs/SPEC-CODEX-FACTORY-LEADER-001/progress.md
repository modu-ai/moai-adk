# SPEC-CODEX-FACTORY-LEADER-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-19T06:50:56Z
- baseline: 8249bff8c3933d647b4f99e59048595f9808e83a
- SPEC ID Bash self-check: `PASS`.
- 구현·런타임·CI 성공은 아직 관측하지 않았다.

## §E.2 Run-phase Evidence

| AC | Actual Output | Status |
|---|---|---|
| AC-CFL-001/003/004/005/006/009 | 아래 focused 명령 PASS. 새 런·거부·기본값·anchor·기존 경로 및 소유자 정리 검사 | PASS (focused) |
| AC-CFL-002 | 실제 engage → 큐/팩토리 DB 재열기 → next/stage → merge-ready → status 조회 PASS. broker 파일 부재 확인. 감사 verdict와 작업 commit은 명시적 lane fixture | PASS (launcher integration) |
| AC-CFL-007/008 | 문서 담당 및 독립 감사 증거 대기 | PENDING |
| invariant: RED first | leader-red.log, leader-shell-red.log, leader-native-separator-red.log, leader-collision-red.log, leader-ulid-red.log에 실제 assertion 실패 보존 | PASS |
| invariant: inherited run preservation | outer-run 및 같은 초 timestamp 런의 backend/소유자 보존 검사 | PASS |

실행 트리: `WT-codex-factory-leader`, plan commit `c8d37362e4` + 현재 변경.

```sh
TMPDIR=/tmp go test ./internal/cli ./internal/hook -run 'TestCodexFactoryLeader|TestFactoryGuideNamesWorkerJoinInEveryLocale|TestSD_AC004|TestCodexPlainLaunch|TestCodexLaneLaunch|TestCodexFactoryEntry|TestCodexFactoryRun|TestFactoryLaneJoin|TestCCFactoryLane|TestGLMFactoryLane|TestFactoryJoinAcceptsCodexLedRun|TestFactoryLeadRefusesCodexLedRun' -count=1 -timeout 5m -coverprofile=.moai/reports/t1774/leader-coverage.out
```

```text
ok  github.com/modu-ai/moai-adk/internal/cli 23.186s coverage: 10.1% of statements
ok  github.com/modu-ai/moai-adk/internal/hook 1.064s coverage: 0.7% of statements
```

최종 ULID 전용 재검증: `TMPDIR=/tmp go test ./internal/cli -run '^TestCodexFactoryLeader' -v -count=1 -timeout 5m -coverprofile=.moai/reports/t1774/leader-new-tests-coverage.out` → `ok github.com/modu-ai/moai-adk/internal/cli 12.780s coverage: 9.6% of statements`. `leader-ulid-final.log`에 ULID 런의 실제 큐 왕복과 Claude/GLM/Codex 혼합 합류 및 교차 리더 인수 거부 PASS가 있다.

로그: `.moai/reports/t1774/leader-final-focused.log`. 패키지 전체 커버리지 85%를 주장하지 않는다.
운영자의 ULID 선택을 반영해 `github.com/oklog/ulid/v2 v2.1.2`와 `crypto/rand.Reader`를 사용한다. lowercase 26자 ULID를 저장하며 기존 cc/glm ID 생성기는 유지한다. 같은 밀리초의 엄격한 순서는 보장하지 않는다.
전체 저장소 판정은 통합 `main`의 CI run 소유이며 **PENDING**이다. 실제 모델·배포·RC 설치·최종 vet/lint/독립 감사는 아직 관측하지 않았다.

추가 AC-CFL-002 검증: `TMPDIR=/tmp go test ./internal/cli -run '^TestCodexFactoryLeaderQueueRoundTripWithoutNudge$' -v -count=1 -timeout 3m` → `ok github.com/modu-ai/moai-adk/internal/cli 4.319s`. 다시 연 큐의 `Runtime.Assignments`와 `Runtime.DispatchCurrent`, 팩토리 binding 모두 같은 run/card/lane을 가리키는지 확인했다. 로그: `.moai/reports/t1774/leader-queue-three-records.log`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-19T07:28:31Z"
run_commit_sha: pending-root-commit
run_status: focused-pass-docs-and-audit-pending
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: not-measured
l44_pre_commit_fetch: pending-root-commit
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: not-measured
cross_platform_build:
  status: not-run
total_run_phase_files: 12
m1_to_mN_commit_strategy: root-owned-explicit-path-commit
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

