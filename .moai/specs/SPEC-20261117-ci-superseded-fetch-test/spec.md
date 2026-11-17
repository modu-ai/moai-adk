---
id: SPEC-20261117-ci-superseded-fetch-test
title: "ci-verdict가 작업을 읽는 경로의 시험과 ci.yml 머리 주석 바로잡기"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-ci-superseded-neutral
---

## Intent

C-83 감사(PASS)의 비차단 지적 두 가지를 고쳐요. N1: `tools/ci_superseded.py`의 `fetch_jobs`(CI에서 `gh api`로 작업과
annotation을 읽는 부분)에 시험이 없어요. N2: `ci.yml` 머리 주석이 "통합 규칙 세트는 `moai/audit`만 요구"라고 적었지만, 규칙
세트 26791615는 `moai/audit`와 `ci`를 둘 다 요구해요(`gh api repos/modu-ai/moai-adk/rules/branches/v2026.11.0`,
2026-11-17 10:11Z 읽음: `required_status_checks` = `moai/audit`, `ci`(integration_id 15368)).

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `When` 요구가 없어 실패 경로 요구도 없어요.

- REQ-1: The job read of `tools/ci_superseded.py` shall ask only for the jobs of the run's own attempt, up to 100 at once.
- REQ-2: The job read shall fetch check-run annotations only for cancelled target jobs.
- REQ-3: The script shall exit 2 and write no state when a read fails.
- REQ-4: The header comment of `.github/workflows/ci.yml` shall name both checks the integration ruleset requires.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 가짜 `gh`가 받은 호출이 정확히 시도 범위 작업 목록 한 번과 취소된 대상 하나의 annotation 한
  번이고, annotation 읽기가 실패하면 exit 2에 state가 없어요.
  verify: `python3 -m unittest tests/ci_superseded_test.py`
- AC-2 (REQ-4): 주석이 두 검사와 읽은 명령을 적고, 워크플로 시험과 actionlint가 통과해요.
  verify: `grep -c 'requires both `moai/audit` and `ci`' .github/workflows/ci.yml` → `1`; `python3 -m unittest tests/github_assets_test.py`;
  `actionlint .github/workflows/ci.yml`

## Scope

- write: `tests/ci_superseded_test.py`, `.github/workflows/ci.yml`(머리 주석만),
  `.moai/specs/SPEC-20261117-ci-superseded-fetch-test/spec.md`
- read-only: `tools/ci_superseded.py`
- forbidden: `.github/workflows/audit-gate.yml`, 규칙 세트 설정(읽기만), 네트워크 쓰기

## Plan

1. 규칙 세트를 읽어(읽기만) 주석을 고쳐요.
2. 가짜 `gh`(PATH 앞에 둔 실행 파일, 호출을 기록하고 정해 둔 답을 돌려줌)로 `fetch_jobs`를 거치는 시험 2개.
3. 검사, 커밋, mutation(모든 시도를 읽는 경로, 모든 대상의 annotation 읽기).

## Decisions

- D1 이 가지는 #3168(C-83)이 아직 병합되지 않아 그 head `ac4fecd2b7` 위에 쌓았어요(리더 지시).
- D2 시험은 스크립트를 고치지 않고 동작만 고정해요. 그래서 RED 대신 mutation으로 시험이 잡는지 보여요.

## Result

(검사 뒤에 채워요.)
