---
id: SPEC-20261119-card-secret-screen
title: "Card text screen refuses secret-named assignments on note, title, body and criterion text"
status: draft
revision: 1
created: 2026-11-19
---

# Card text screen refuses secret-named assignments on note, title, body and criterion text

## Intent

사용자가 2026-11-19 에 카드 C-122 의 선택지 3("비밀 이름이면 모두 거절")을 골랐어요(카드 결정 노트 3143, 계획 노트 3106). 지금의 노트 화면(C-68, `src/card/work.rs` 의 `blank_prose`)은 `--password` 나 `--token` 같은 옵션 이름 뒤에 등호와 평범한 값이 오면 통과시키고, `card add` 의 제목·본문·`--ac` 기준 글에는 비밀 검사가 아예 없어요(C-108 감사 N4, base 부터 있던 틈). 이 조각은 두 곳을 같은 규칙 하나로 묶어요: 이름을 하이픈(`-`)과 밑줄(`_`)로 자른 조각 가운데 `password`, `token`, `secret` 이 있거나 `api` 와 `key` 가 나란히 있으면, 이름-등호-값 꼴은 값이 무엇이든(빈 값도) 거절해요. 알려진 대가는 무해한 한도 숫자나 참·거짓 값(옵션 이름에 token 이 든 꼴)도 거절된다는 것이고, 값을 띄어 쓰면 통과해요. 사용자가 고르지 않은 두 질문은 계획의 기본값대로예요: 소문자 낱말을 하이픈으로 이은 `sk-` 덩어리는 통과를 유지하고, 기준 글·본문·제목에도 검사를 걸어요. `card criterion-add` 는 C-108 병합 뒤 별도 조각이에요.

## Non-goals

- `card criterion-add` 에 검사 걸기: C-108(PR 3230)이 병합된 뒤 별도 조각이에요.
- 이미 장부에 있는 행 고치기·다시 검사하기: 새 쓰기만 검사해요(결정 3143).
- 점·카멜 이름(`db.password`, `apiKey`)과 따옴표 안의 꼴 따로 보기: 이 조각은 `[A-Za-z0-9_-]` 이름만 봐요.
- 공유 화면 `redact::redact_outgoing` 바꾸기: 그 함수는 그대로 두고 카드 쪽 화면만 고쳐요.
- `card export` 와 `--reason`·`--evidence` 의 방향 문자 검사 바꾸기.

## Requirements

- REQ-1: The card text screen shall refuse a word that holds an assignment whose name, cut at `-` and `_`, has a piece equal to `password`, `token` or `secret` (any letter case) or the two pieces `api` and `key` side by side, whatever the value is, an empty value included.
- REQ-2: The card text screen shall let an assignment through, with or without leading dashes, when no piece of its name is one of those words and its value has no credential shape.
- REQ-3: The card text screen shall refuse a word whose value has a credential shape that `redact::redact_outgoing` names, whatever the assignment name is.
- REQ-4: The card text screen shall let a run of lower-case words joined by hyphens after the `sk` prefix through, as the note screen does today.
- REQ-5: Each text field that the commands `card add` or `card note` or `card criterion` store shall pass through the one function `secret_words` in `src/card/input.rs`, the fields being the title and the body and each `--ac` and the note text and `--ref` and `--evidence` and `--reason`.
- REQ-6: A refused text shall give the refusal kind `invalid_input`, name the field, count each refused word once and never repeat the text.
- REQ-7: The command `card add` shall run the screen before it writes, so a refused add leaves the ledger as it was.
- REQ-8: The commands `card show` and `card list` shall print a stored text that the screen would refuse today without an error.

## Acceptance

- AC-1 (REQ-1): 이름 조각에 password·token·secret·api-key·api_key 가 든 이름-등호-값 꼴(옵션 대시 있음·없음, 값이 평범함·빔·수·참거짓)이 노트 본문, 제목, 본문, `--ac` 네 표면에서 모두 거절돼요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_refuses_secret_named_assignments_on_every_surface`
- AC-2 (REQ-2, REQ-3, REQ-4): 비밀 낱말이 없는 이름과 평범한 값은 통과하고, 값이 자격 증명 모양(AKIA, 깃허브 토큰 접두, 숫자·대문자가 섞인 sk 접두 덩어리)이면 이름과 상관없이 거절되고, 소문자 낱말만 하이픈으로 이은 sk 접두 덩어리는 통과해요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_lets_plain_text_and_hyphenated_sk_names_through`; `cargo test --locked --offline --lib secret_screen`
- AC-3 (REQ-6, REQ-7): 거절은 `invalid_input` 이고 필드 이름(title·body·criterion 번호·note·ref)을 말하고 글을 되풀이하지 않으며, 거절된 add 는 카드·기준 글 행을 하나도 만들지 않아요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_refusals_name_the_field_and_write_nothing`
- AC-4 (REQ-6): 한 낱말은 한 번만 세요: 이름도 값도 걸리는 낱말 하나는 1, 걸리는 낱말 둘은 2.
  verify: `cargo test --locked --offline --test card_cli secret_screen_counts_each_word_once`
- AC-5 (REQ-5): 노트·ref·criterion 증거·이유가 같은 함수를 부르므로 기존 C-68 시험이 새 규칙 아래에서도 통과해요(`tests/card_work_cli.rs` 의 한 못박은 줄은 결정 3143 에 맞게 바꿔요).
  verify: `cargo test --locked --offline --test card_work_cli`
- AC-6 (REQ-8): 새 규칙이 거절할 글이 이미 든 행(원장에 직접 넣은 기준 글)을 `card show` 와 `card list` 가 오류 없이 보여 줘요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_leaves_stored_rows_alone`
- AC-7 (REQ-1, REQ-2): 기존 카드 125장의 글에서 새 규칙이 걸릴 개수를 읽기 전용으로 센 값이 Result 에 있어요.
  manual: 카드 노트 3152 의 수치(기준 글 0/268, 제목 1/125, 본문 3/113, 노트 5/3147)를 Result 에 옮겨 적고, 구현 뒤 같은 스크립트로 구현된 규칙을 다시 센 값을 한 줄 더해요.

## Scope

- write: `src/card/input.rs`, `src/card/work.rs`, `tests/card_cli.rs`, `.moai/specs/SPEC-20261119-card-secret-screen/spec.md`
- write (후속, 범위가 열린 뒤): `tests/card_work_cli.rs` 의 못박은 한 줄(C-68 AC-3 시험이 옵션 token 의 등호 값 꼴을 통과로 못박은 곳)
- read-only: `src/redact.rs`, `src/card/command.rs`, `src/decide/outbound.rs`, `tests/card_work_cli.rs`
- forbidden: 실제 HOME, 네트워크, `src/redact.rs` 수정, `card criterion-add`

## Plan

1. SPEC 을 쓰고 `spec check` 로 확인한 뒤 `card spec` 으로 카드에 연결해요.
2. RED: `tests/card_cli.rs` 에 표 시험 네 개와 `input.rs` 단위 시험을 먼저 쓰고, 지금 코드에서 실패하는 것을 눈으로 확인해요(입력은 조각을 이어 붙여 만들어 소스에 비밀 모양 한 줄도 남기지 않아요).
3. 검사 함수 `secret_words` 를 `src/card/input.rs` 에 두고(`blank_prose`·`without_sk_names` 를 옮겨요), `work.rs` 의 노트·ref·criterion 이 그것을 불러요. 한 낱말은 한 번만 세요(거절문의 낱말 수 표시 오류도 같이 고쳐져요).
4. `validate_add` 가 제목·본문·`--ac` 에 같은 함수를 걸어요.
5. `tests/card_work_cli.rs` 의 못박은 한 줄을 결정 3143 에 맞게 고쳐요(범위가 열린 뒤).
6. 시험을 돌려요: `--test card_cli`, `--test card_work_cli`, `--lib secret_screen`, fmt, clippy `--all-targets`.

## Risks

- 기준 글에 새 규칙이 거는 오탐: 기존 기준 글 268개에서는 0개예요(노트 3152). 앞으로 쓰는 글이 옵션 token 의 등호 숫자 꼴을 쓰면 거절돼요(알려진 대가; 값을 띄어 쓰면 통과).
- `blank_prose` 를 옮기면서 노트의 기존 통과·거절 줄이 달라질 수 있어요: C-68 시험(`card_work_cli`)과 이 SPEC 의 표 시험이 같은 입력을 지켜요.
- 이 저장소의 설치된 `moai`(2026-11-15 빌드)는 옛 노트 화면이라 sk 접두 이름이나 옵션 등호 꼴을 적은 카드 노트도 거절해요: 노트는 풀어 쓰고, 시험은 이 가지에서 빌드한 `moai-cli` 로 돌려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이름을 `-`·`_` 로 자른 조각이 비밀 낱말과 같을 때만 비밀 이름(부분 문자열은 아님) | 결정 3143 의 말("조각에 비밀 낱말"); `tokens`·`passwordless` 같은 말을 막지 않아요 | 조각 비교를 부분 문자열로 | 
| D-2 | `api` 와 `key` 가 나란한 두 조각을 한 낱말(`api_key`, `api-key`)로 본다 | 계획 노트 3106 의 다섯 낱말 | 쌍 규칙 빼기 |
| D-3 | 값이 빈 꼴도 거절한다 | 이름이 비밀이면 값과 상관없다는 결정; base 도 거절했어요 | 빈 값 예외 |
| D-4 | 검사 함수는 `src/card/input.rs` 에 두고 `work.rs` 가 부른다 | 입력 규칙은 `input.rs` 에 모여 있어요; `work.rs` 는 이미 `input` 을 써요; 쓰기 범위 안 | 함수를 `work.rs` 로 옮기기 |
| D-5 | 거절은 낱말 단위로 센다 | 거절문의 "값 하나가 2" 표시 오류 수리(note 3106 5절) | 글 단위로 세기 |
| D-6 | 이미 있는 행은 다시 검사하지 않는다 | 결정 3143("기존 행은 고치지 않음") | 별도 마이그레이션 카드 |

## Result

(구현 뒤에 채워요.)
