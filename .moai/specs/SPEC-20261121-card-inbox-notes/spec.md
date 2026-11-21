---
id: SPEC-20261121-card-inbox-notes
title: "card inbox recognises the M-Flow note families READY, MERGE-READY and CANDIDATE"
status: draft
revision: 1
created: 2026-11-21
---

# card inbox recognises the M-Flow note families READY, MERGE-READY and CANDIDATE

## Intent

카드 흐름 전수 보고서(2026-11-21)의 P13: `card inbox` 는 머리말 4종(감사 요청, LANE-DONE … audit
requested, LANE-BLOCKED, LANE-READY-INTEGRATION, DECISION-QUESTION)만 요청으로 읽어서, M-Flow 의
`READY …`, `MERGE-READY …`, `CANDIDATE …` 노트를 인식하지 못해요 — 리더가 병합 준비된 머리를
인박스에서 놓쳐요. 이 조각은 `family_of` 에 그 세 표지를 더하고(새 family 세 개, 기존 응답·최신
규칙은 그대로), 단위 시험과 CLI 시험으로 고정해요.

## Non-goals

- `--project` 일관성(보고서 P14)과 세션 id 검사: 이 조각 밖이에요 — 파일이 다른 카드들의 살아 있는
  예약에 있어서 리더 정산 뒤 같은 카드의 다음 가지에서 이어가요.
- 비밀 모양 검사의 자리표시 허용과 위치 안내: 같은 이유로 이 조각 밖이에요.
- `RETURN`, `PAIR-LINE` 표지: 카드 기준이 이름은 세 개뿐이라 이 조각도 세 개만 더해요; 새 표지는
  한 줄이면 더해지는 설계예요.
- `pending` 의 응답·최신 규칙 변경: 없어요 — 새 family 는 기존 규칙을 그대로 따라요.

## Requirements

- REQ-1: `family_of` shall recognise a first line that starts with `READY`, `MERGE-READY` or `CANDIDATE` as its own family, printed as `ready`, `merge_ready` and `candidate`.
- REQ-2: The new families shall follow the same rules as the old ones: only the newest request of a family per card is listed, and a later note by the leader answers it.
- REQ-3: The inbox unit tests shall pin the three prefixes, their case sensitivity, the not-a-request edges, and the answer and newest rules of the new families.
- REQ-4: The inbox CLI test shall list the three new families from a seeded ledger.

## Acceptance

- AC-1 (REQ-1…REQ-3): unit tests in `src/card/inbox.rs` cover the three new families and the old
  edges stay green.
  verify: `cargo test --locked --offline --lib card::inbox`
- AC-2 (REQ-4): the CLI test lists `ready`, `merge_ready` and `candidate` items from a seeded
  ledger, including one answered by a later leader note.
  verify: `cargo test --locked --offline --test card_inbox_cli`
- AC-3: 이웃 검사.
  verify: `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `python3 -I tools/source_manifest.py check` (불변이면 ok)

## Scope

- write: `src/card/inbox.rs`, `tests/card_inbox_cli.rs`,
  `.moai/specs/SPEC-20261121-card-inbox-notes/spec.md`
- read-only: `src/card/mod.rs` (변경이 필요 없으면 그대로), 리더 인박스 소비자 문서
- forbidden: 원장 스키마 변경, `pending` 서명 변경, 실제 HOME 쓰기

## Plan

1. RED: `family_of` 의 세 표지 단위 시험과 CLI 기대 행을 먼저 쓰고 실패를 읽어요(enum 항목 미구현
   컴파일 실패가 첫 RED예요).
2. GREEN: `Family` 세 항목과 `family_of` 세 갈래를 더해요.
3. 이웃 검사를 돌려요.

## Risks

- `READY` 로 시작하는 다른 뜻의 노트를 요청으로 읽어요 → 표지는 머리말 한 줄 판정이라 오답이 인박스에
  한 줄 더 보이는 것뿐이고, 리더 노트가 답하면 사라져요; 보고서 P13 이 이 표지를 요청한다고 적었어요.
- 대소문자·띄어쓰기 변주 → 시험이 lower-case 비요청과 앞 공백 허용을 고정해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 세 family 를 `Family` enum 에 추가하고 `code()` 로 `ready`·`merge_ready`·`candidate` 를 출력해요 | 기존 소비자가 `family` 문자열로 걸러 읽어요; 보고서 P13 의 이름을 그대로 써요 | 후속 카드가 표지를 데이터로 옮기면 enum 을 고쳐요 |
| D-2 | `pending` 은 고치지 않아요 | 새 family 도 같은 응답·최신 규칙이 맞아요(리더가 카드에 답하면 처리된 것) | 규칙이 달라져야 하면 그때 함수를 나눠요 |

## Result

- AC-1: met — RED 는 컴파일 실패였어요: `cargo test --locked --offline --lib card::inbox` 가
  `Family::Ready`·`Family::MergeReady`·`Family::Candidate` 변입이 없다며 5 errors, exit 101. GREEN
  뒤 같은 필터 8 passed, exit 0 — `family_is_read_from_the_start_of_the_first_line_only` 가 세
  표지·lower-case 비요청·줄 중간 표지·앞 공백·빈 줄을 묶고, 새
  `the_new_families_follow_the_answer_and_newest_rules` 가 응답(리더 노트 뒤면 숨김)·같은 family
  최신만 나열을 묶어요. 기록: 첫 두 실행은 같은 커밋의 다른 트리로 만든 빌드 캐시를 mtime 만으로
  재사용해 무의미했어요(다른 워크트리 소스로 빌드 폴더를 쓸 때는 touch 규칙이 필요 — 소스 touch 뒤
  재빌드에서 올바른 RED 를 읽었어요).
- AC-2: met — `cargo test --locked --offline --test card_inbox_cli` 4 passed, exit 0.
  `the_inbox_lists_each_request_family_in_note_order` 가 C-1 의 `READY …`·`MERGE-READY …`·
  `CANDIDATE …` 노트를 `ready`·`merge_ready`·`candidate` 행으로 나열하고 count 8 을 묶어요;
  `a_later_leader_note_answers_and_a_newer_note_supersedes` 가 리더 답 뒤의 `MERGE-READY` 는 숨고
  답 앞의 `LANE-READY-INTEGRATION` 은 나오는 걸 묶어요.
- AC-3: met — `cargo fmt --check` exit 0(적용 2곳, 재검사 0); `cargo clippy --locked --offline
  --all-targets -- -D warnings` exit 0; `python3 -I tools/source_manifest.py check` ok(이 조각은
  skills 를 고치지 않아 매니페스트 불변). 남은 것: SPEC 은 draft예요(승인은 사용자의 선택); 조각 2
  (`--project` 일관성)와 조각 3(비밀 검사)은 이 카드의 기준 1의 나머지 절반으로, 다른 카드들의
  살아 있는 예약이 정산되면 이 카드의 다음 가지에서 이어가요.
