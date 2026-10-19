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

### 감사 수정 검증

기준: 구현 `5fa01379ba` + 현재 감사 수정. 첫 독립 감사 FAIL의 F1(native cwd anchor 우회)·F2(attached 모델 기본값 중복)를 수정했다. 새 리더의 native `-C`/`--cd`를 런 생성 전에 거부하고 `-w`를 안내하며, 모델·effort 판별은 attached 옵션·옵션 값·native `--` 경계를 구분한다. 실제 주입한 tmux primitive 실패에서 owner 잔존도 재현해 해당 오류 경로의 리더 owner를 정리했다. 도움말에 `-f, --factory`를 추가했다.

RED: `.moai/reports/t1774/audit-fixes-red.log`에 cwd 5형태 허용, `-mcustom` 기본값 중복, prompt 경계 오인, `failed spawn retained owner=36523`, 도움말 누락의 assertion 실패를 보존했다. 수정 후 해당 테스트 묶음은 `audit-fixes-focused.log`에서 `ok github.com/modu-ai/moai-adk/internal/cli 4.735s`였다.

```sh
TMPDIR=/tmp go test ./internal/cli -run 'TestFactoryEntryMatrix|TestCodexFactoryFlagRefusals|TestFactorySkillAssertionsMatchBehavior|TestCodexFactoryLeader|TestCodexPlainLaunch|TestCodexLaneLaunch|TestCodexSessionDefault' -count=1 -timeout 5m
```

```text
ok  github.com/modu-ai/moai-adk/internal/cli 52.847s
```

로그: `.moai/reports/t1774/audit-fixes-regression.log`. 수정 전 광범위 회귀에서 드러난 세 테스트 가족의 옛 계약도 이 실행에서 확인했다. 문서 검사는 거부 예시를 표에서 실제 launcher로 검사한 뒤 prose 금지 규칙을 적용하며, 잘못된 accepted 행과 금지 prose를 잡는 mutant를 유지한다. 이 결과는 수정 후 focused 검증이며 수정 전 광범위 실행을 소급해 PASS로 바꾸지 않는다. 재감사·최종 vet/lint·RC 설치는 아직 관측하지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-19T08:08:00Z"
run_commit_sha: "5fa01379ba + audit-fixes-pending-root-commit"
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

```yaml
sync_complete_at: pending-final-verification
sync_commit_sha: pending-backfill
sync_status: doc-ready-audit-and-rc-pending
b12_self_test_a: "PASS: pre-emission grep count=0"
b12_self_test_b: "PASS: tier=M; ac_source=.moai/specs/SPEC-CODEX-FACTORY-LEADER-001/acceptance.md; live=9 excluded=0 ambiguous=0; CHANGELOG references 9"
b12_self_test_c: "PASS: linked SPEC path exists (ls)"
changelog_entry_position: "[Unreleased] / Added / first entry"
frontmatter_status_transitions:
  spec: "in-progress retained; completion pending audit and RC installation"
  plan: "stateless; unchanged"
  acceptance: "stateless; unchanged"
canary_compliance_check:
  status: not-applicable
```

README와 docs-site의 ko/en/ja/zh 팩토리 안내, factory 진입 표, 로컬·배포용 factory-dispatch 규칙을 동기화했다. 새 Codex 리더 런·혼합 레인·SQLite 배차 기록·메시지 없이 진행 확인·merge-ready 경계·지원하지 않는 훅·기존 런 재개 제한을 반영했다. 배포 템플릿에는 내부 카드·SPEC 식별자나 실행 기록을 넣지 않았다.

문서 확인 기준: `WT-codex-factory-leader`, 구현 commit `5fa01379ba` 위의 문서 변경. `git diff --check`는 출력 없이 exit 0이었다. 문서에 링크한 SPEC 파일은 `ls`로 확인했다. AC counter는 stdout `9`, stderr `live=9 excluded=0 ambiguous=0`이었다.

독립 구현 감사, 최종 템플릿 빌드·회귀 검사, docs-site 빌드 및 운영자가 요청한 RC 설치 검증은 아직 완료하지 않았다. sync 완료나 최종 AC 통과를 주장하지 않으며 SPEC 상태는 `in-progress`로 유지한다.
