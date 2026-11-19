---
id: SPEC-20261119-macos-on-pr
title: "macos-arm64 on PRs that touch platform-sensitive paths"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261117-supported-platforms
---

# macos-arm64 on PRs that touch platform-sensitive paths

## Intent

PR CI는 #2554(U10) 이후 linux-x86_64만 돌아요. 그래서 macOS에서만 깨지는 변경이 merge 뒤에야 tip을 붉히었어요(이력 분석
2026-11-17: merge 뒤 push 실패 68건 중 47건이 macOS-only, 9건 windows-build; 붉은 tip 25구간 15.9시간). 이 SPEC은 변경 경로가
플랫폼 민감(pty/terminal, process/signal, filesystem links/permissions, launcher, statusline, build/toolchain — 47건에서 측정)이면
merge 전에 `check` 매트릭스에 macos-arm64를 넣고, aggregate `ci`가 그 결과를 요구하게 해요. 나머지 PR은 linux-only를 유지해요.
필수 검사 이름 `ci`와 ruleset 26791615는 바꾸지 않아요(카드 C-76 leader 결정, 2026-11-18T23:21:54Z).

## Non-goals

- Windows를 PR에서 돌리기(windows-build는 push 전용 유지).
- push/workflow_dispatch 매트릭스 변화(linux + macos 그대로).
- ruleset·필수 검사·보호 설정 변경, `ci-verdict`(superseded 중립화) 로직 변경.
- macOS 전용 시험을 리눅스에서 돌리게 하기(방향이 반대예요).

## Requirements

- REQ-1: When a `pull_request` event's changed paths include a platform-sensitive path, the `check` job's matrix shall include both `linux-x86_64` and `macos-arm64`.
- REQ-2 (REQ-1): IF a `pull_request` event's changed paths include no platform-sensitive path, THEN the `check` job's matrix shall include `linux-x86_64` only (unchanged #2554 U10).
- REQ-3: When the event is `push` or `workflow_dispatch`, the `check` job's matrix shall include both targets regardless of the changed paths.
- REQ-4 (REQ-3): IF the event is `push` or `workflow_dispatch` and the changed paths include no platform-sensitive path, THEN the `check` job's matrix shall still include both targets.
- REQ-5: When `tools/ci_scope.py` reads the changed paths, it shall decide platform sensitivity and report `macos_needed` with `macos_reasons` in the same JSON it already emits.
- REQ-6 (REQ-5): IF the classifier input is empty, unparsable or internally erroneous, THEN it shall answer `macos_needed: true` (fail safe to the wider matrix), leaving the docs-only decision's fail-safe unchanged.
- REQ-7: When the run's `check` matrix is decided, the aggregate job shall keep the name `ci` and fail unless every target of that matrix succeeded.
- REQ-8 (REQ-7): IF any target of the `check` matrix fails or is skipped while the run is not superseded, THEN the aggregate `ci` check shall report failure under the unchanged name.
- REQ-9: When the workflow starts, a preceding `targets` job shall compute the `check` matrix by running the classifier on the event's merged diff and emit it as a job output.
- REQ-10 (REQ-9): IF the `targets` job cannot compute a matrix (checkout, diff or classifier failure), THEN it shall fail so the `ci` check is red, never silently linux-only.

## Acceptance

- AC-1 (REQ-1, REQ-5): 플랫폼 민감 경로를 담은 파일 목록은 `macos_needed: true`와 범주 이유를 돌려줘요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.ci_scope_test -k Platform`
- AC-2 (REQ-2, REQ-5): 민감하지 않은 목록(docs-only, 원장/스킬 문서 등)은 `macos_needed: false`를 돌려줘요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.ci_scope_test -k Platform`
- AC-3 (REQ-6): 빈 입력·쓰레기 입력·내부 오류는 `macos_needed: true`로 안전하게 실패해요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.ci_scope_test -k Platform`
- AC-4 (REQ-1, REQ-2): 규칙표는 역대 실패한 merge들의 파일 목록으로 단위 시험돼요 — 2026-11-19에 읽은 v2026.11.0 push 실패 69 run의 병합 델타
  파일 목록 각각이 판정되고, `false`로 나온 run은 시험에 run id로 이름 나열돼요(여섯 범주 밖 — 일부는 macOS-only였을 수 있고, 그 간극은 Result에 기록).
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.ci_scope_test.CiScopePlatformTest.test_historical_failed_merge_file_lists`
- AC-5 (REQ-3, REQ-4, REQ-7, REQ-8, REQ-9, REQ-10): ci.yml 계약 — `targets` job이 매트릭스를 출력하고 `check`가 `fromJSON(needs.targets.outputs.matrix)`를
  소비하며, aggregate job 이름은 `ci` 그대로, push/dispatch 분기와 `ci-verdict` 조건은 변하지 않아요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.ci_scope_test.CiWorkflowMatrixContractTest`; `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'github_assets_test.py'`

## Scope

- write: `.github/workflows/ci.yml`, `tools/ci_scope.py`, `tests/ci_scope_test.py`, `tests/github_assets_test.py`, `.moai/specs/SPEC-20261119-macos-on-pr/spec.md`
- read-only: `tools/local_check.sh`, `tools/ci_superseded.py`, `docs/pr-contract.md`
- forbidden: ruleset 26791615, 필수 검사 이름/구성, 실제 HOME·원장·키, 네트워크 쓰기

## Plan

1. 증거: 실패 push 69 run의 head에서 병합 델타 파일 목록을 뽑고(스크립트, `target/c76/`), 여섯 범주를 실제 경로 규칙으로 대응.
2. RED: `tests/ci_scope_test.py`에 Platform 분류·계약 시험 추가 → 실패 확인.
3. GREEN: `tools/ci_scope.py`에 규칙표·`macos_needed` 추가; `ci.yml`에 `targets` job과 매트릭스 소비 변경.
4. 이웃 검사: `tests/ci_scope_test.py` 전체, `tests/github_assets_test.py`, `python3 -m unittest discover`, `cargo fmt --check`, 워크플로 YAML 파싱.

## Risks

- 매트릭스를 `fromJSON(needs.targets.outputs.matrix)`로 바꾸면 잘못된 JSON이 check 를 건너뛰어요 → `targets` 실패/빈 출력은 안전 쪽(전체 매트릭스)으로
  빠지고, 잡히면 `ci`가 붉어요(REQ-9, REQ-10). YAML 계약 시험이 소비 형태를 못박아요.
- `targets` job이 PR에서 merge diff 없이는 판정 못 함 → `github.event.pull_request.base.sha` 기반 `git diff --name-only --no-renames`(기존 Scope 단계와 같은 형태).
- 분류기 과다 판정(macOS가 자주 돎)은 비용만 있고 안전해요; 과소 판정만 위험해요 → 규칙표는 범주 단위로 넓게, 예외 목록은 시험에 이름 올림.
- macOS runner 대기(12–21분 관측)가 민감 PR의 merge를 늦춰요 → push 매트릭스와 같은 큐 규칙(job concurrency)이라 완화 수단 없음; 감사에 관측치 기록.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 판정은 `ci_scope.py` 한 곳에서(docs-only과 같은 입력, 같은 JSON에 키 추가) | 완료 조건 2; 두 분류기가 갈라지지 않음 | 별도 도구로 분리 |
| D-2 | 매트릭스 조립은 새 `targets` job; `check`는 소비만 | 매트릭스는 job 정의 시점에 필요해 판정을 앞세워야 함 | 워크플로 수준 삼항 복원 |
| D-3 | 근거 단위는 여섯 측정 범주를 경로 규칙(접두사/정확히)으로 옮긴 표 | 완료 조건 1의 범주 목록 | 파일별 allowlist |
| D-4 | 역대 검증은 실패 push 69 run의 파일 목록 전체(47 macOS-only 포함) | run별 job 구분은 API 69호출이 필요해 비례에 안 맞음; 전체 집합은 47을 덮는 상위 집합 | 47건만 골라 시험 |

## Result

- AC-1: met — `tests.ci_scope_test` 34 시험 통과; 여섯 범주 대표 경로 22개가 `macos_needed: true` + 범주 이유를 돌려줘요(RED 먼저: 구현 전 오류 8·실패 3).
- AC-2: met — 비민감 목록(docs·skills·원장 로직) `false`; 접두사 누출 경계(`src/statusline_notes.rs`, `adapters/claude/statusline.py`)도 `false`.
- AC-3: met — 빈 입력·제어문자 경로·내부 오류가 모두 `true`; 돌연변이 M1·M2·M3로 안전실패 분기가 시험에 묶여 있음을 확인(6/6 KILLED, `target/c76/mut.log`).
- AC-4: met — 역대 실패 69 run의 파일 목록 전수 판정: 29 true / 40 false, false는 run id로 시험에 명명(`HISTORICAL_FALSE_RUN_IDS`). 간극(감사 N-1 실측, 2026-11-19): 이 69 run 중 macos-arm64-only는 28 run이고, 이 표가 그중 macOS를 미리 돌려 줬을 run은 9뿐이에요(원인 아닌 run 단위 집계). 표 넓히기는 후속 카드.
- AC-5: met — YAML 계약 시험·YAML 파싱(PyYAML, job 그래프 확인) 통과; `tests/github_assets_test.py`의 `rc_ci_matrix` 4시험은 새 구조(`targets` job의 `full=`/`linux=` 목록)로 고쳐 39 시험 전부 통과(리더 인계 뒤, 카드 note 2383·2386; `python3 -m unittest tests.github_assets_test tests.ci_scope_test` → Ran 73 tests, OK).
