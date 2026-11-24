---
id: SPEC-20261123-glm-fable-slot-test
title: "The GLM fable slot test tells the fable slot from the opus slot"
status: draft
revision: 1
created: 2026-11-23
---

# GLM fable 슬롯 테스트가 fable 슬롯과 opus 슬롯을 구분해요

## Intent

`glm_fable_slot_names_and_sizes_its_model` 은 런처의 `GLM_ENV` 만 써요. 거기서는 opus 슬롯과 fable 슬롯의 값이 같아서(`glm-5.4[1m]`), `resolve_glm_model_name` 의 fable 키(`ANTHROPIC_DEFAULT_FABLE_MODEL`)를 opus 키로 바꾼 뮤턴트(mutant: 지키는 코드를 뒤집은 복사본)가 218/218 테스트를 통과해 살아남았어요(C-396 감사 N2). 두 슬롯의 값이 다른 환경에서 `Fable` 이름이 fable 슬롯의 모델을 보이는지 지키는 테스트를 더해요. 코드는 바꾸지 않아요.

## Requirements

- REQ-1: When the opus slot and the fable slot of the environment hold different GLM models, the model name resolution shall show the fable slot's model for a Fable name and the opus slot's model for an Opus name.
- REQ-2 (REQ-1): IF only the opus slot holds a GLM model, THEN the model name resolution shall leave a Fable name as it is.

## Acceptance

- AC-1 (REQ-1, REQ-2): 새 테스트 `glm_fable_and_opus_slots_resolve_their_own_models` 가 통과하고, fable 키를 opus 키로 바꾼 뮤턴트에서는 실패해요 · verify: `cargo nextest run --lib -E 'test(glm_fable_and_opus_slots_resolve_their_own_models)'`
- AC-2 (REQ-1, REQ-2): 같은 변경이 `src/statusline/state.rs` 의 코드(테스트 모듈 밖)를 한 줄도 바꾸지 않아요 · verify: `git diff NeXT...HEAD --stat -- src`

## Scope

- write: `src/statusline/state.rs`(테스트 모듈만), `.moai/specs/SPEC-20261123-glm-fable-slot-test/**`
- read-only: `src/session.rs`(`GLM_ENV`)
- forbidden: `resolve_glm_model_name` 과 그 밖의 코드 변경, 새 파이썬 시험

## Plan

1. RED: 새 테스트를 쓰고, `resolve_glm_model_name` 의 fable 키를 opus 키로 바꾼 복사본에서 새 테스트가 실패하는지 봐요. 복사본은 되돌려요.
2. GREEN: 바꾸지 않은 코드에서 새 테스트와 `statusline::` 테스트가 통과하는지 봐요.
3. 뮤턴트 줄(`guard-mutant:`)과 `check affected` 영수증을 READY 노트에 실어요.
