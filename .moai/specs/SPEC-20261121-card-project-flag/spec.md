---
id: SPEC-20261121-card-project-flag
title: "card CLI: --project on every reader, stand-in values pass the secret screen"
status: draft
revision: 1
created: 2026-11-21
---

# card CLI: --project on every reader, stand-in values pass the secret screen

## Intent

카드 흐름 보고서(2026-11-21)의 두 항을 이 카드의 조각으로 풀어요. (조각 2) `card list·withdraw·
next·inbox` 네 하위명령만 프로젝트를 위치 인자로 받아 `card withdraw C-1 … --project .` 같은
호출이 사용법 오류가 돼요 — 네 곳에 `--project` 를 더해 같은 방식으로 받게 해요(위치 인자는
호환을 위해 그대로, 플래그가 이기면). (조각 3a) 비밀 모양 검사가 `<redacted>`, `***`, `$NAME`,
`x` 같은 자리표시 값까지 거절하고 거절 문구가 고칠 방법을 알려 주지 않아요 — 뚜렷한 자리표시 값은
통과시키고 문구에 다시 쓰기 제안을 더해요. 위치(몇 번째 낱말) 안내는 그 호출부가 다른 카드의 살아
있는 예약에 있어서 이 카드의 다음 가지예요.

## Non-goals

- 거절이 몇 번째 낱말인지 말하기: 그 문구를 조립하는 호출부(src/card/work.rs, src/card/dispatch.rs)가
  다른 카드 예약에 있어요 — 그 카드가 반환하면 같은 카드의 다음 가지로 더해요.
- 세션 id 를 환경에서 검증 없이 읽는 일(landed·withdraw): 이 조각의 기준 밖이에요(보고서의 같은 항의
  다른 절반).
- `secret_words` 의 반환 형 변경: 호출부가 그대로라 갯수 기반을 유지해요.
- redact.rs 의 공유 검사 변경: 담당 카드의 영역이에요.

## Requirements

- REQ-1: Every card subcommand that took the project folder as a positional word shall also accept `--project <path>`, which wins over the positional word.
- REQ-2: The secret screen shall let a word through when it is a secret-named assignment whose value is an obvious stand-in (wrapped in angle brackets, all asterisks from three, a dollar name, or the literal x, xx, xxx, ... or …).
- REQ-3: The secret screen refusal shall suggest the rewording, naming a stand-in value or a different field name, beside the count it already states.

## Acceptance

- AC-1 (REQ-1): `card list`, `card withdraw`, `card next`, `card inbox` each accept `--project`
  alone in one CLI test; the positional word keeps working (the existing tests run it).
  verify: `cargo test --locked --offline --test card_inbox_cli the_reader_commands_accept_the_project_flag`
- AC-2 (REQ-2): a stand-in value passes on a note (angle-wrapped, dollar name and the literal xx
  tried), an empty value still refuses, and a real shape still refuses.
  verify: `cargo test --locked --offline --test card_secret_followups a_stand_in_value_passes_and_the_refusal_suggests_rewording`
- AC-3 (REQ-3): the refusal message names the rewording while the pinned substring stays.
  verify: 같은 시험
- AC-4: 이웃 검사.
  verify: `cargo test --locked --offline --test card_inbox_cli --test card_secret_followups --test card_withdraw_cli --test card_list_title_prefix`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ops/card.rs`, `src/card/input.rs`, `tests/card_inbox_cli.rs`,
  `tests/card_secret_followups.rs`, `.moai/specs/SPEC-20261121-card-project-flag/spec.md`
- read-only: `tests/card_withdraw_cli.rs`, `tests/card_list_title_prefix.rs` (이 조각은 그 파일들을
  읽고 기존 위치 인자 시험에 맡겨요), `src/card/work.rs`·`src/card/dispatch.rs`(호출부 — 다음 가지)
- forbidden: 실제 HOME 쓰기, `secret_words` 서명 변경, redact.rs 변경

## Plan

1. RED: 네 명령 `--project` 시험(card_inbox_cli.rs)과 자리표시 통과·빈 값·실모양 시험
   (card_secret_followups.rs)을 먼저 쓰고 실패를 읽어요(사용법 오류·거절이 첫 RED).
2. GREEN: ops/card.rs 의 네 하위명령에 플래그 필드와 해석을 더하고, input.rs 에 자리표시 판정과
   문구 제안을 더해요.
3. 이웃 검사를 돌려요.

## Risks

- 위치 인자와 플래그가 같은 이름이에요 → 두 필드로 두고 팔에서 이겨요; 도움말에 플래그가 이긴다고
  적어요.
- 자리표시 판정이 넓어 실모양을 통과시켜요 → 값 전체가 하나의 자리표시 모양일 때만 통과하고, 빈 값과
  실모양은 기존 규칙 그대로라 시험이 경계를 묶어요.
- 거절 문구 변경이 기존 핀을 깨요 → 핀이 묶는 부분 문자열(shape of a secret)은 유지해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 위치 인자를 없애지 않고 `--project` 를 더하고 플래그가 이겨요 | 기존 호출·문서를 깨지 않는 보고서의 "같은 방식으로 받기" | 후속 카드가 위치 인자를 공식 폐기하면 빼요 |
| D-2 | 자리표시는 값 전체가 하나의 모양일 때만 통과예요 | 부분 자리표시(`<value` 뒤 붙은 실값)는 실값으로 봐야 안전해요 | 사용자 사례가 모자라면 목록을 더해요 |
| D-3 | 거절 위치 안내는 이 조각에서 빼요 | 호출부 파일이 다른 카드 예약에 있어요; 문구만 바꿔도 핀 부분 문자열은 유지돼요 | 호출부 반환 뒤 같은 카드의 다음 가지에서 더해요 |
| D-4 | 카드 본문 예시의 `x` 는 자리표시에서 뺐어요 | 기존 감사 핀이 `"password" = x` 의 거절을 묻어요(감사 근거 있는 기존 계약); 한 글자는 검사를 시험하려고 치는 값이기도 해요 | 리더가 본문 예시를 계약으로 확정하면 목록에 더해요 |

## Result

- AC-1: met — RED 먼저: `the_reader_commands_accept_the_project_flag` 가 card_inbox_cli.rs:92 에서
  실패(`card list --project` 사용법 오류). GREEN 뒤 같은 시험 ok — `card list`(`cards` 배열)·
  `card next`(범위 있는 준비 카드를 이름 — 범위 없는 카드는 건너뛰는 기존 규칙 그대로)·
  `card withdraw`(세션 환경과 함께)·`card inbox`(차단 노트 1개만 세)가 각각 `--project` 단독으로
  돌아요. 위치 인자 호환은 기존 시험들이 그대로 통과하는 것으로 묶여요.
- AC-2: met — RED 먼저: `a_stand_in_value_passes_and_the_refusal_suggests_rewording` 실패 —
  `token=<redacted>` 가 거절됨(거절 문구 전문이 실패 출력에 남아요). GREEN 뒤 같은 시험 ok:
  `<redacted>`·`$MY_TOKEN`·`...` 통과, 빈 값(`token=`)과 실모양(16진 값)은 거절.
- AC-3: met — 같은 시험: 거절 문구가 stand-in 제안(stand-in 부분 문자열)을 가지고 감사 핀의
  `shape of a secret` 부분 문자열은 유지돼요.
- AC-4: met — 이웃: `card_withdraw_cli` 4 passed, `card_list_title_prefix` 6 passed(카멜 핀
  `name_forms…` 복구 포함), `cargo test --lib card::input` 26 passed, `cargo fmt --check` 0(적용
  1곳), `cargo clippy --locked --offline --all-targets -- -D warnings` 0, 파이썬 스캔 51 OK. 기록:
  첫 GREEN 의 두 실패를 고쳤어요 — next 기대(범위 없는 카드는 건너뛴다는 기존 규칙 반영)와 카멜 핀
  (`"password" = x` 거절 유지 — D-4 로 `x` 를 자리표시 목록에서 뺌).
- 남은 것: SPEC 은 draft예요(승인은 사용자의 선택). 거절의 위치 안내(몇 번째 낱말)는 호출부
  (src/card/work.rs·src/card/dispatch.rs)가 반환되면 같은 카드의 다음 가지에서 더해요.
