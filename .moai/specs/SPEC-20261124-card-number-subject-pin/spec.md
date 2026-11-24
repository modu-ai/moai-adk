---
id: SPEC-20261124-card-number-subject-pin
title: "trailer-check pins the card-number rule on commit subjects and on quoted ledger output"
status: draft
revision: 1
created: 2026-11-24
branch: spec/20261125-card-number-subject-pin
parent: SPEC-20261120-commit-message-card-numbers
---

# trailer-check pins the card-number rule on commit subjects and on quoted ledger output

## Intent

`moai trailer-check` 는 이미 커밋 제목과 본문의 카드 번호 꼴을 거절해요(`--card-numbers-since`, 상위 SPEC 으로 착지). 그런데 제목 쪽은 실패 방향 테스트가 없어서 제목을 읽는 줄이 빠져도 모든 테스트가 통과해요(감사 비차단 발견). 이 조각은 제목 쪽 실패와 통과를 테스트로 고정하고, 실제 커밋 기록에서 허용되는 모양(원장 출력 인용 등)을 잘못 막는지 재서 결과를 남겨요. `--message-file` 쪽 제목은 기존 테스트가 이미 막고 있어서(별도 함수) 이 조각이 다루지 않아요. commit-msg 훅의 설치는 이 조각 밖이에요.

## Requirements

- REQ-1: While the card-number rule is active for a range, trailer-check shall report a failing `card_number` finding on line 1 for a card-number shape in the subject of a commit after the recorded start.
- REQ-2: While the card-number rule is active for a range, trailer-check shall leave a subject that carries only a quoted ledger refusal, a SPEC id, a version, a hexadecimal SHA or an `AC-<n>` neighbour without a `card_number` finding.
- REQ-3: While the Result holds the measurement, it shall give the command, the range and the counts of the rule run over the real commit history, and the verdict on each distinct allowed shape the run flagged.

## Acceptance

- AC-1 (REQ-1, REQ-2): fixture commits whose subject carries a card number fail with a `card_number` finding on line 1, and fixture subjects with the allowed shapes pass · verify: `cargo test --locked --offline --test trailer_check_cli -- card_numbers_in_a_subject_fail_after_a_recorded_start allowed_shapes_in_a_subject_pass`
- AC-2 (REQ-3): the Result holds the history measurement and the spec check passes · verify: `moai spec check SPEC-20261124-card-number-subject-pin --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-card-number-subject-pin/**`
- write: `tests/trailer_check_cli.rs`
- read-only: everything else, `src/trailers.rs` included; a false block found by the measurement waits for a revision of this SPEC

## Plan

1. 실제 기록(2026-11-19 이후 통합 브랜치)에 규칙을 돌려 분류하고 잘못 막는 모양이 있는지 읽는다.
2. 제목 쪽 실패·통과와 message-file 제목 테스트를 `tests/trailer_check_cli.rs` 에 더한다.
3. 제목 읽는 줄과 인용 면제를 하나씩 없애는 뮤턴트로 새 테스트가 죽는지 본다.
4. 영수증과 READY 노트(뮤턴트 출력 파일 경로 포함).
