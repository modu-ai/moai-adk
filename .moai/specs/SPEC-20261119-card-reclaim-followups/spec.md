---
id: SPEC-20261119-card-reclaim-followups
title: "card reclaim follow-ups — help wording, one lease-end test, secret check on --reason, old-generation CLI tests"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-card-reclaim
---

# card reclaim follow-ups — help wording, one lease-end test, secret check on --reason, old-generation CLI tests

## Intent

`moai card reclaim`(SPEC-20261119-card-reclaim, PR #3217)의 독립 감사가 막지 않는 후속 항목 다섯 개를 한 조각으로 정리해요: 도움말·동작 문서가
`review` 카드는 `card reopen` 을 먼저 해야 돌려줄 수 있다고 말하게 하고(N-2), 거절 종류가 `conflict` 하나가 아니라 `conflict`·`not_found`·`invalid_input`
이라고 말하게 하고(N-3), lease 끝 경계 시험이 시계의 같은 초에 기대지 않게 하고(N-5), `--reason` 에 `card note` 와 같은 비밀값 모양 검사를 하고(N-6),
옛 세대의 `pr`·`done`·`reopen`·`review` 가 명령 줄에서 거절됨을 시험으로 묶어요(N-8). 동작이 바뀌는 곳은 `--reason` 검사 하나예요.

## Non-goals

- 사용자의 "언제든" 회수(N-4), 회수 때 범위 겹침 재검사(N-7), 옛 보유자 확인과 세 번 한도.
- `card criterion` 의 근거·사유 글에 비밀값 검사를 넣기(같은 수준의 빈틈이지만 이 조각 밖이에요).

## Requirements

- REQ-1: The help of `card reclaim` and its operations document shall say that a card in `review` is reopened with `card reopen` before it is given back with `card release`.
- REQ-2: The help of `card reclaim` and its operations document shall name the three refusal kinds, `conflict`, `not_found` and `invalid_input`, and what each one covers.
- REQ-3: Where the ledger decides that a reservation lease has ended (the fence, `release`, `reclaim`), the decision shall be one function, `lease_ended`, with the edge `lease_until <= now`.
- REQ-4: When `--reason` has a word with the shape of a secret (the filter `card note` uses, `redact::redact_outgoing`), `card reclaim` shall refuse it as `invalid_input` before the ledger is opened, without repeating the word · failure: none — the refusal is this requirement's behavior; the check runs before anything is read or written, so nothing can be half done.
- REQ-5: While a newer generation of the card exists, `card pr`, `card done`, `card reopen` and `card review` shall refuse the reclaimed generation.

## Acceptance

- AC-1 (REQ-1, REQ-2): `card reclaim --help` 가 `card reopen` 먼저와 세 거절 종류를 말하고, 동작 문서의 `card reclaim` 줄이 같은 말을 해요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_help_names_reopen_and_the_refusal_kinds`; `manual: docs/design-review/appendix/operations/card.md 의 card reclaim 줄과 `reclaim` 항목을 읽어 도움말과 같은 말인지 본다`
- AC-2 (REQ-3): `lease_ended` 가 N-1·N·N+1 초에서 참·참·거짓이고, 울타리·`release`·`reclaim` 이 모두 그 함수를 써요; 같은 초 시험이 시계의 초가 바뀌어도 같은 답을 내요.
  verify: `cargo test --locked --offline --lib lease_ended`; `cargo test --locked --offline --test ledger_reservation reclaim`
- AC-3 (REQ-4): 비밀값 모양의 `--reason` 은 `invalid_input` 이고 원장 모든 행이 그대로이며 오류 글에 그 단어가 없어요; 평범한 이유는 받아들여요.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_refuses_a_secret_shaped_reason`
- AC-4 (REQ-5): 회수한 뒤 옛 세대의 `pr`·`done`·`reopen`·`review` 가 `is stale` 의 `conflict` 이고 카드 상태·예약·링크가 그대로예요.
  verify: `cargo test --locked --offline --test card_reclaim_cli the_old_generation_cannot_move_or_close_the_card`
- AC-5 (REQ-1…REQ-5): 이웃 검사.
  verify: `cargo test --locked --offline --test card_reclaim_cli`; `cargo test --locked --offline --test ledger_reservation`; `cargo test --locked --offline --test card_dispatch_cli`;
  `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --test ac_command_lint`;
  `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ledger/reservation.rs`, `src/card/dispatch.rs`, `src/ops/card.rs`, `tests/card_reclaim_cli.rs`, `tests/ledger_reservation.rs`,
  `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261119-card-reclaim-followups/spec.md`
- read-only: `src/redact.rs`, `src/card/work.rs`, `src/ledger/workflow.rs`, `tests/common/**`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소, 스키마 바꾸기

## Plan

1. 시험 먼저: 도움말 시험, 비밀값 모양 이유 시험, 옛 세대 CLI 시험, `lease_ended` 단위 시험.
2. `lease_ended` 를 `reservation.rs` 에 두고 울타리(`evaluate`)·`release`·`reclaim` 이 쓰게 해요; `dispatch::reclaim_reason` 에 `redact_outgoing` 검사.
3. `src/ops/card.rs` 도움말과 동작 문서를 고쳐요.

## Risks

- 비밀값 모양 검사는 "조심스러운 모양 거름망"이지 탐지기가 아니에요(`redact.rs` 설명): 거름망을 지나는 비밀은 이유 글에 남을 수 있어요(`card note` 와 같은 수준).
- `lease_ended` 로 바꾸는 세 곳은 같은 부등호(`<=`)를 이미 쓰고 있어서 동작은 같아야 해요; 시험이 세 곳 모두를 묶어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 같은 초 경계는 시계에 기대는 통합 시험 대신 순수 함수 `lease_ended` 의 N-1·N·N+1 단위 시험으로 묶음 | 감사 N-5, 리더 지시: 시계에 기대지 않게 | 통합 시험만 두기 |
| D-2 | `--reason` 검사는 `card note` 와 같은 `redact_outgoing`, 글은 되풀이하지 않고 단어 수만 말함 | 감사 N-6, 리더 지시: `card note` 와 같은 redact | `criterion` 처럼 검사 없음 |
| D-3 | 근거·사유 글의 비밀값 검사를 `criterion` 에 넣는 일은 하지 않음 | 이 조각의 범위는 `reclaim`; 다른 명령의 입력 규칙은 별도 SPEC | `criterion` 에도 넣기 |

## Result

(RUN 과 VERIFY 뒤에 AC 마다 한 줄 씁니다.)
