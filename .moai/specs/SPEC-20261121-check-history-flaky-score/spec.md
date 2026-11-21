---
id: SPEC-20261121-check-history-flaky-score
title: "check affected records per-target history and prints a flaky score"
status: draft
revision: 1
created: 2026-11-21
---

# check affected records per-target history and prints a flaky score

## Intent

검사 부하 재설계 보고서(2026-11-21 §9.5, 외부 근거 행)의 제안: 나중에 "이 영역 변경에서 한 번도 안
깨진 대상"을 뒤로 미루려면 먼저 **대상별 결과·시간 이력**이 있어야 해요. 이 카드는 `check affected`
실행이 시험 대상 하나하나의 결과와 시간을 지역 파일(JSON Lines)에 기록하게 하고, 같은 트리에서
실패했다가 통과한 전이(플레이키)를 점수로 요약에 보여 줘요. RED 시험은 가짜 도구로 완전히 재현
가능해서 실제 cargo 없이 묶여요. 이 조각은 SPEC 과 RED 시험까지예요 — 구현은 선행 조각(정밀 선택)이
같은 소스 파일에 오는 순서를 지켜야 해서 그 뒤 같은 카드의 다음 가지로 이어가요.

## Non-goals

- `src/check_affected.rs`·`src/ops/check.rs` 의 구현: 선행 조각(정밀 선택)이 먼저 같은 파일에
  착지하는 순서를 지켜요 — 이 조각은 시험과 SPEC 만 써요.
- 이력으로 검사 대상을 줄이는 학습 선택: 이력이 쌓인 뒤의 다음 단계예요.
- 원격 저장·여러 워크트리 합치기: 지역 파일 하나(호출자가 `--flaky-history` 로 지정)만이에요.
- Python 대상의 시간 분해: 대상 하나의 통과·실패·초만 기록해요.

## Requirements

- REQ-1: `check affected --flaky-history <path>` shall append one history line per test invocation of a target, carrying the head, the target name, the exit code and the seconds it took.
- REQ-2: The flaky score of a target shall count its fail-then-pass transitions recorded against the same head.
- REQ-3: The answer shall carry the flaky targets with their scores, flakiest first, and the summary shall end with the total flaky count before the result word.
- REQ-4: The flaky score shall be computed from the history file alone, so the CLI test can seed and read it without running the real tools.

## Acceptance

- AC-1 (REQ-1, REQ-4): 가짜 도구로 두 번 실행 — 첫 실행은 한 대상이 실패(종료 3)하고 이력 파일에
  head·대상·종료 코드·초가 줄로 쌓여요; 둘째 실행(같은 트리, 전부 통과)이 그 이력을 읽어 점수를
  내요.
  verify: `cargo test --locked --offline --test check_affected_cli a_flaky_history_records_targets_and_scores_a_fail_then_pass`
- AC-2 (REQ-2, REQ-3): 둘째 실행의 결과에 flakiest-first 배열(해당 대상 점수 1)과 요약의
  ` flaky=1 ` 이 있어요.
  verify: 같은 시험
- AC-3: 이웃 검사.
  verify: `cargo test --locked --offline --test check_affected_cli`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `tests/check_affected_cli.rs`,
  `.moai/specs/SPEC-20261121-check-history-flaky-score/spec.md`
- read-only: `src/check_affected.rs`, `src/ops/check.rs` (이 조각은 읽기만 — 구현 가지가 써요)
- forbidden: 실제 cargo·python 실행(가짜 도구만), 원격 쓰기, 실제 HOME

## Plan

1. RED: `a_flaky_history_records_targets_and_scores_a_fail_then_pass` 를 먼저 쓰고 실행해 실패를
   읽어요(플래그가 아직 없어 clap 이 사용법 오류로 끝내는 것이 첫 RED).
2. GREEN(다음 가지, 선행 조각 착지 뒤): 실행 루프의 대상별 호출에서 기록하고, 이력 읽기·점수를
   순수 함수로 더하고, 요약·결과에 내보내요.
3. 이웃 검사를 돌려요.

## Risks

- 이력 파일이 커져요 → JSON Lines 에 줄 추가만 있어요; 잘라내기는 후속이에요.
- 시간이 가짜 도구에서 0 에 가까워요 → 기록은 유한한 f64 초로 하고 시험은 0 이상만 묶어요.
- 실패한 호출이 이후 실행의 FAKE_FAIL 과 섞여요 → 가짜 도구는 인수에 이름이 있을 때만 3 으로
  끝나요(기존 Tools 규칙 그대로).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이력은 호출자가 지정한 파일 하나(JSON Lines)에 줄을 더해요 | 카드 본문의 "local file"; 워크트리의 무시 폴더를 호출자가 골라요 | 후속 카드가 기본 자리와 잘라내기를 정해요 |
| D-2 | 점수는 같은 head 안의 실패→통과 전이만 셈 | 카드 본문의 "fail-then-pass on the same tree"; head 해시가 트리 대신 표준 키예요 | 트리 해시를 더로 옮기면 키를 바꿔요 |
| D-3 | 구현은 선행 조각(정밀 선택) 착지 뒤 같은 소스에서 이어가요 | 리더 배차의 순서(선행 조각이 먼저); 같은 파일 충돌을 피해요 | 순서가 바뀌면 이 가지에서 바로 구현해요 |
