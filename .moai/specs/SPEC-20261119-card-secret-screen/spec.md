---
id: SPEC-20261119-card-secret-screen
title: "Card text screen refuses secret-named assignments on note, title, body and criterion text"
status: draft
revision: 3
created: 2026-11-19
---

# Card text screen refuses secret-named assignments on note, title, body and criterion text

## Intent

사용자가 2026-11-19 에 카드 C-122 의 선택지 3("비밀 이름이면 모두 거절")을 골랐어요(카드 결정 노트 3143, 계획 노트 3106). 지금의 노트 화면(C-68, `src/card/work.rs` 의 `blank_prose`)은 `--password` 나 `--token` 같은 옵션 이름 뒤에 등호와 평범한 값이 오면 통과시키고, `card add` 의 제목·본문·`--ac` 기준 글에는 비밀 검사가 아예 없어요(C-108 감사 N4, base 부터 있던 틈). 이 조각은 두 곳을 같은 규칙 하나로 묶어요: 이름을 하이픈(`-`)과 밑줄(`_`)로 자른 조각 가운데 `password`, `token`, `secret` 이 있거나 `api` 와 `key` 가 나란히 있으면, 이름-등호-값 꼴은 값이 무엇이든(빈 값도) 거절해요. 알려진 대가는 무해한 한도 숫자나 참·거짓 값(옵션 이름에 token 이 든 꼴)도 거절된다는 것이고, 값을 띄어 쓰면 통과해요. 사용자가 고르지 않은 두 질문은 계획의 기본값대로예요: 소문자 낱말을 하이픈으로 이은 `sk-` 덩어리는 통과를 유지하고, 기준 글·본문·제목에도 검사를 걸어요. `card criterion-add`(C-108)는 로컬 v2026.11.0 에 병합되어 기준 글을 공용 함수 `criterion_texts` 로 받으므로, 그 함수 안에 검사를 두면 `card add --ac` 와 같은 화면을 받아요.

## Non-goals

- 이미 장부에 있는 행 고치기·다시 검사하기: 새 쓰기만 검사해요(결정 3143).
- 점·카멜 이름(`db.password`, `apiKey`)과 따옴표 안의 꼴 따로 보기: 이 조각은 `[A-Za-z0-9_-]` 이름만 봐요.
- 공유 화면 `redact::redact_outgoing` 바꾸기: 그 함수는 그대로 두고 카드 쪽 화면만 고쳐요.
- `card export` 와 `--reason`·`--evidence` 의 방향 문자 검사 바꾸기.

## Requirements

- REQ-1: The card text screen shall refuse a word that holds an assignment whose name, cut at `-` and `_`, has a piece equal to `password`, `token` or `secret` (any letter case) or the two pieces `api` and `key` side by side, whatever the value is, an empty value included.
- REQ-2: The card text screen shall let an assignment through, with or without leading dashes, when no piece of its name is one of those words and its value has no credential shape.
- REQ-3: The card text screen shall refuse a word whose value has a credential shape that `redact::redact_outgoing` names, whatever the assignment name is.
- REQ-4: The card text screen shall let a run of lower-case words joined by hyphens after the `sk` prefix through, as the note screen does today.
- REQ-9: The card text screen shall read the name of an assignment as the word was written, so an `sk` run that holds a secret piece and stands before an equals sign is refused.
- REQ-10: The card text screen shall look at the text after each equals sign, colon and comma inside a word as a value of its own, so a credential shape is refused after a plain name too.
- REQ-11: The card text screen shall screen a word in time that grows in step with its length, so one word of a million bytes without a space is answered in seconds.
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
- AC-8 (REQ-5): `card criterion-add --ac` 도 같은 함수(`input::criterion_texts`)를 거쳐 비밀 이름 낱말을 거절하고 기준 글을 하나도 더하지 않아요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_covers_criterion_add`
- AC-9 (REQ-9): 에스케이 접두에 하이픈 소문자 낱말을 이은 이름의 조각 가운데 비밀 낱말이 있고 바로 뒤가 등호이면(대시 있음·없음, 비밀 낱말이 첫째·둘째·셋째 조각, 빈 값) 네 표면 모두 거절되고, 등호 없는 덩어리와 비밀 조각 없는 이름의 대입은 통과해요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_refuses_secret_names_behind_an_sk_prefix`; `cargo test --locked --offline --lib secret_screen_reads_sk_names_before_they_are_blanked`
- AC-10 (REQ-10): 클라우드 키 식별자 모양, 숫자·대문자가 섞인 에스케이 값, 점 셋 토큰 모양이 평범한 이름 뒤 등호·콜론·쉼표 뒤(괄호 안 포함)에 와도 네 표면에서 거절되고, 콜론·쉼표·등호가 든 평범한 문장 12개는 통과해요.
  verify: `cargo test --locked --offline --test card_cli secret_screen_refuses_credential_shapes_after_any_separator`; `cargo test --locked --offline --lib secret_screen_sees_values_after_a_separator`
- AC-11 (REQ-5, REQ-9, REQ-10): 노트 참조와 기준의 증거·사유도 같은 화면을 지나 비밀 이름 대입과 구분자 뒤 값 모양을 거절하고 필드 이름을 한 번 말해요.
  verify: `cargo test --locked --offline --test card_work_cli the_reference_evidence_and_reason_pass_the_secret_screen`
- AC-12 (REQ-11): 구분자(쉼표·콜론·등호)나 공백 없는 JSON 꼴이 되풀이되는 한 낱말을 200 KB, 노트 한도보다 1바이트 작게, 한도 정확히 세 크기로 넣으면 15초 안에 답하고, 낱말 끝의 자격 증명 모양은 그 긴 낱말에서도 거절돼요.
  verify: `cargo test --locked --offline --lib secret_screen_time_on_one_long_word_has_a_bound`
- AC-7 (REQ-1, REQ-2): 기존 카드 125장의 글에서 새 규칙이 걸릴 개수를 읽기 전용으로 센 값이 Result 에 있어요.
  manual: 카드 노트 3152 의 수치(기준 글 0/268, 제목 1/125, 본문 3/113, 노트 5/3147)를 Result 에 옮겨 적고, 구현 뒤 같은 스크립트로 구현된 규칙을 다시 센 값을 한 줄 더해요.

## Scope

- write: `src/card/input.rs`, `src/card/work.rs`, `tests/card_cli.rs`, `tests/card_work_cli.rs`(C-68 AC-3 시험이 옵션 token 의 등호 값 꼴을 통과로 못박은 한 곳만), `.moai/specs/SPEC-20261119-card-secret-screen/spec.md`
- read-only: `src/redact.rs`, `src/card/command.rs`, `src/card/dispatch.rs`, `src/decide/outbound.rs`
- forbidden: 실제 HOME, 네트워크, `src/redact.rs` 수정

## Plan

1. SPEC 을 쓰고 `spec check` 로 확인한 뒤 `card spec` 으로 카드에 연결해요.
2. RED: `tests/card_cli.rs` 에 표 시험 네 개와 `input.rs` 단위 시험을 먼저 쓰고, 지금 코드에서 실패하는 것을 눈으로 확인해요(입력은 조각을 이어 붙여 만들어 소스에 비밀 모양 한 줄도 남기지 않아요).
3. 검사 함수 `secret_words` 를 `src/card/input.rs` 에 두고(`blank_prose`·`without_sk_names` 를 옮겨요), `work.rs` 의 노트·ref·criterion 이 그것을 불러요. 한 낱말은 한 번만 세요(거절문의 낱말 수 표시 오류도 같이 고쳐져요).
4. `validate_add` 가 제목·본문·`--ac` 에 같은 함수를 걸어요.
5. `tests/card_work_cli.rs` 의 못박은 한 줄을 결정 3143 에 맞게 고쳐요(C-116 이 풀린 뒤 카드를 release → 전체 범위로 다시 take 해서 열었어요).
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
| D-7 | 이름 검사는 에스케이 이름을 지우기 전의 원문 낱말에 하고, 값 모양 검사만 지운 낱말에 한다 | 감사 B1(노트 3297): 지우기가 이름 조각을 없애 결정 3143 의 규칙이 구멍났어요; 등호 없는 덩어리는 여전히 이름이라 통과해요 | 이름 검사를 지운 낱말에 |
| D-8 | 값 모양은 낱말 안의 등호·콜론·쉼표 뒤마다 따로 본다(공백은 낱말 나눔이 이미 해요) | 감사 N1(노트 3297): 공유 화면은 접두 모양을 낱말 맨 앞에서만 봐요; REQ-3 은 이름과 상관없이라고 했어요 | 구분자 목록 줄이기 |

## Result

- AC-1: met — `cargo test --locked --offline --test card_cli secret_screen_refuses_secret_named_assignments_on_every_surface` 통과: 이름 8종(password·token·secret·api_key·api-key·max-token·skip-password·DB_PASSWORD) × 값 4종(평범·빔·수·참거짓) × 대시 2종 × 표면 4곳(노트·제목·본문·`--ac`) 256 호출이 모두 `invalid_input`, 카드·기준 글·노트 행 수 불변.
- AC-2: met — `--test card_cli secret_screen_lets_plain_text_and_hyphenated_sk_names_through` 와 `cargo test --locked --offline --lib secret_screen`(단위 5개) 통과: 비밀 낱말 없는 이름·평범한 값·띄어 쓴 값·소문자 하이픈 sk 이름 통과, AKIA·깃허브 토큰 접두·숫자 섞인 sk 거절.
- AC-3: met — `--test card_cli secret_screen_refusals_name_the_field_and_write_nothing` 통과: 문구가 "the title has 1 word"·"the body has 1 word"·"criterion 2 has 1 word"·"the note text has 1 word"·"the note ref has 1 word" 로 필드를 말하고 값(`zz9top`)을 되풀이하지 않으며, 거절된 add 는 카드·기준 글·이벤트 행을 만들지 않음.
- AC-4: met — `--test card_cli secret_screen_counts_each_word_once` 통과: 이름도 값도 걸리는 낱말 하나 = 1, 걸리는 낱말 둘 = 2(이전 화면은 하나를 2로 셌어요).
- AC-5: met — `cargo test --locked --offline --test card_work_cli` 25/25 통과(`--test card_cli` 19, `card_reclaim_cli` 9, `card_dispatch_cli` 25, `--lib card::` 77 도 통과). 못박은 한 줄은 `tests/card_work_cli.rs` 의 C-68 AC-3 시험: 통과하던 옵션 token 의 등호 꼴을 옵션 mode 의 등호 꼴로 바꾸고, token 꼴이 거절되는 단언을 더했어요.
- AC-6: met — `--test card_cli secret_screen_leaves_stored_rows_alone` 통과: 원장에 직접 넣은 제목·기준 글(새 규칙이 거절할 글)을 `card show`·`card list` 가 오류 없이 보여 줘요(이 시험은 구현 전에도 통과했어요 — REQ-8 은 새 쓰기만 검사해서 처음부터 참).
- AC-7: met — 읽기 전용으로 센 값. 구현 전(노트 3152, 정규식 근사): 기준 글 0/268, 제목 1/125, 본문 3/113, 노트 5/3147. 구현 뒤(이 가지 `moai-cli` 를 임시 HOME·임시 장부에서 `card add`·`card note` 로 돌려서 센 값, `scratchpad c122/count_after.py`, 2026-11-19): 기준 글 0/268, 제목 1/125(C-122), 본문 3/113(C-3·C-68·C-122), 노트 7/3205(C-3·C-22·C-32; 노트 수는 그 사이 늘었어요). 실제 장부에는 쓰지 않았고 기존 행은 그대로예요.
- 변이: 5개를 하나씩 적용해 새 시험이 잡는지 봤어요(모두 사살, 복원 뒤 `git status` 깨끗): (1) `secret` 조각을 목록에서 뺌 → `secret_screen_name_pieces_decide` 실패, (2) `api`+`key` 쌍 규칙을 무력화 → 같은 시험 실패, (3) 빈 값 허용 → `secret_screen_refuses_secret_named_assignments_on_every_surface` 실패, (4) 낱말 수를 최대 1로 → 단위 2개와 `secret_screen_counts_each_word_once` 실패, (5) 소문자 하이픈 sk 이름 예외 끄기 → `secret_screen_passes_plain_assignments_and_names` 와 `secret_screen_lets_plain_text_and_hyphenated_sk_names_through` 실패.
- AC-8: met — 로컬 v2026.11.0(C-108 포함)을 이 가지에 병합한 뒤 `cargo test --locked --offline --test card_cli secret_screen_covers_criterion_add` 통과: 둘째 기준 글에 token 이름의 등호 값 꼴이 든 `card criterion-add` 가 "criterion 2 has 1 word" 로 거절되고 기준 글 행은 0개, 평범한 글은 `added: [1]`. 병합 충돌은 `src/card/input.rs` 한 곳(공용 `criterion_texts` 안에 검사를 둠). 병합 뒤 `--lib` 1495·`card_cli` 20·`card_work_cli` 25·`card_criterion_add_cli` 7·`card_reclaim_cli` 16·`card_dispatch_cli` 25 통과.
- 따로 남긴 것(범위 밖): `card reclaim --reason`(`src/card/dispatch.rs:389-391`)은 아직 예전 공유 화면을 직접 불러요 — 이 조각의 화면과 다른 규칙이에요(후속 후보).
- 수리(감사 FAIL, 노트 3297, 수리 head 는 노트 참조):
  - AC-9: met — 수리 전 RED(두 시험이 거절을 기대하는 곳에서 실패)를 보고 `input.rs` 를 고친 뒤 `--test card_cli secret_screen_refuses_secret_names_behind_an_sk_prefix` 와 `--lib secret_screen_reads_sk_names_before_they_are_blanked` 통과.
  - AC-10: met — `--test card_cli secret_screen_refuses_credential_shapes_after_any_separator` 와 `--lib secret_screen_sees_values_after_a_separator` 통과; 평범한 문장 12개(콜론·쉼표·등호 포함)가 네 표면에서 계속 통과해요.
  - AC-11: met — `--test card_work_cli the_reference_evidence_and_reason_pass_the_secret_screen` 통과(노트 참조·증거·사유 각 6개 입력, 필드 이름 한 번).
  - 수리 변이 4개(모두 사살, 복원 뒤 `git status` 깨끗): (1) 이름 검사를 지운 낱말에 되돌림 → `secret_screen_reads_sk_names_before_they_are_blanked` 실패, (2) 구분자 뒤 검사를 끔 → `secret_screen_sees_values_after_a_separator` 실패, (3) 쉼표를 구분자에서 뺌 → 같은 시험 실패, (4) 콜론을 뺌 → 같은 시험 실패.
  - N2 변이 3개(노트 참조·증거·사유 검사를 하나씩 끔): 모두 `--test card_work_cli the_reference_evidence_and_reason_pass_the_secret_screen` 실패로 사살, 복원 뒤 `git status` 깨끗. 시험 입력에 예전 화면이 통과시키던 꼴(비밀 낱말 플래그에 수 값, 토큰 조각이 든 한도 플래그에 수 값)을 넣었어요.
  - N4: 비밀 이름 꼴 글자가 `tests/card_work_cli.rs` 의 시험 한 줄과 `src/card/input.rs` 의 문서 주석에 있던 것을 조각 이어 붙이기·풀어 쓰기로 바꿨어요(`git diff v2026.11.0 HEAD` 의 더한 줄에서 비밀 이름 등호 꼴 검색 0건).
  - 기존 카드 글 재측정(수리한 head 의 `moai-cli`, 임시 HOME·임시 장부, `count_after.py`, 2026-11-19): 기준 글 0/287, 제목 1/130, 본문 3/118, 노트 7/3301 — 수리 전(0/268, 1/125, 3/113, 7/3205)과 같은 카드의 같은 글만 걸려요(걸리는 카드: 3, 22, 32, 68, 122).
- 수리 2(재감사 FAIL, 노트 3357, B2 제곱 시간):
  - 원인과 고침: 낱말 안 구분자마다 뒤 전체를 공유 화면에 다시 넘겼어요. 구분자를 공백으로 바꿔 한 번만 넘기도록 고쳤어요(`src/card/input.rs` `word_is_refused`). 구분자를 품은 접두 모양은 `TOKEN_PREFIXES`·`redact_outgoing` 에 없지만, 결과가 모두 같은 것은 아니에요(재감사 N8, 노트 3396): 접두 모양의 길이를 다음 구분자 전까지만 세게 되어, 짧은 접두 조각이 쉼표·콜론을 넘어 이어지는 두 꼴이 거절에서 통과로 바뀌었어요. 진짜 키는 세 구분자를 품지 않아서 놓치는 키 모양은 없어요. 이름 검사·값 검사 다른 갈래는 그대로예요.
  - 측정(수리 전 → 후, debug 빌드, 20 KB 한 낱말, 같은 시험 탐침): 쉼표 5.33초 → 0.026초, 콜론 5.72초 → 0.016초, 등호 4.61초 → 0.022초, JSON 꼴 3.23초 → 0.011초, 하이픈·`sk-` 는 수 ms 로 처음부터 빨랐어요.
  - AC-12: met — `--lib secret_screen_time_on_one_long_word_has_a_bound` 통과(네 꼴 곱하기 200 KB·한도−1·한도, 15초 상한, 끝 모양 확인). 시험은 스레드에서 화면을 돌려 제곱 시간이 되돌아오면 15초 뒤 바로 실패해요.
  - 변이 2개(모두 사살, 복원 뒤 `git status` 깨끗): (1) 구분자마다 뒤 전체를 넘기는 옛 꼴로 되돌림 → 시간 상한 시험이 15.01초에 실패, (2) 변환에서 쉼표를 뺌 → `secret_screen_sees_values_after_a_separator` 실패.
- 돌리지 않은 것: 전체 `cargo test`, CI, 실제 HOME, 점·카멜 이름.
