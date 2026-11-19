---
id: SPEC-20261119-leader-inbox
title: "moai card inbox lists the lane notes that still wait for the leader"
status: draft
revision: 1
created: 2026-11-19
---

# moai card inbox lists the lane notes that still wait for the leader

## Intent

리더는 한 바퀴에 카드 열두어 장을 임시 정규식으로 훑으면서 카드마다 마지막 노트 두 개만 봤어요. 2026-11-19 에 그래서 C-101 조각 B 의 tier 2 감사 요청(노트 2827)을 약 한 시간 놓쳤고(리더 노트 2874 가 "요청 노트 2827 을 늦게 읽었어요" 라고 적어요), 08:41Z 에는 감사 요청 네 건이 한꺼번에 쌓여 있었어요. 레인은 그동안 쉬었어요. 원장에는 노트가 전부 있어요 — 모자란 것은 질문(query)이에요. 이 변경은 읽기 전용 명령 `moai card inbox` 를 더해서, 카드마다 "리더의 답을 기다리는 가장 새 레인 노트"를 한 번에 나열해요.

카드 C-116 이 정하라고 한 것은 "노트 종류는 지금 자유 텍스트인데, 줄 접두어로 맞출지 새 종류를 둘지"예요. 측정 결과 노트 종류는 자유 텍스트가 아니에요: `card_note.kind` 는 `decision`·`report`·`observation`·`comment` 네 값만 허용하는 `CHECK` 예요(`src/ledger/schema.rs:67`). 새 종류는 스키마 변경이라 이 카드의 범위를 넘고, 이미 쌓인 노트 2,926개(2026-11-19 측정, `card show` 를 전 카드에 돌려 셈)에는 소급되지 않아요. 그래서 줄 접두어로 맞춰요(D-1).

## Non-goals

- 노트 종류 추가나 원장 스키마 변경, 기존 노트 고치기.
- 답을 쓰는 일(`card note`)과 배차(`card ready`·`card take`), 감사 판정 — 이 명령은 보여 주기만 해요.
- 리더가 아닌 사람의 답(사용자 결정 노트 등)을 "답"으로 세기, 그리고 한 레인의 `LANE-NOTE` 진행 줄 안의 질문을 알아내기(D-5).
- 다른 프로젝트나 GitHub 이슈 댓글 읽기.

## Requirements

- REQ-1: When the leader runs `card inbox`, the command shall list, for each card that is not done, the newest note of each request family that no later leader note answers, giving the card, the family, the note id, the author, the first line of the note and its age in seconds.
- REQ-2: The command shall recognise an audit request by the first-line prefix `감사 요청` or by a first line that starts with `LANE-DONE` and contains `audit requested`, a blocked report by the prefix `LANE-BLOCKED`, a ready line by the prefix `LANE-READY-INTEGRATION`, and a decision question by the prefix `DECISION-QUESTION`.
- REQ-3: The command shall treat a request as answered when the same card holds a note written by the leader with a larger note id.
- REQ-4: The command shall list only the newest note of a family on a card when the card holds an older note of that family.
- REQ-5: The command shall treat as the leader the author named by `--leader`, and the author `leader` when `--leader` is not given.
- REQ-6: The command shall leave out notes imported from another host and notes written by the leader.
- REQ-7: The command shall order the items by note id, oldest first.
- REQ-8: The command shall open the ledger read-only and leave the ledger, its folder and its files unchanged.
- REQ-9: While the project has no ledger, the command shall answer `ledger: "absent"` with no items and exit 0.
- REQ-10 (REQ-5): IF `--leader` is empty, longer than 80 characters or holds a control character, THEN the command shall refuse with `card inbox refused: …` and write nothing.
- REQ-11 (REQ-1): IF the ledger is blocked or cannot be read, THEN the command shall answer the ledger's own refusal and write nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-7): 한 카드에 네 요청이 하나씩 있으면(`감사 요청 tier 2 head=…`, `LANE-DONE … — audit requested`, `LANE-BLOCKED …`, `LANE-READY-INTEGRATION …`, `DECISION-QUESTION …`) 가족별로 하나씩 나오고, 접두어가 아닌 줄(`LANE-DONE` 만, `LANE-NOTE` 질문, 줄 가운데의 `LANE-BLOCKED`)은 나오지 않아요. 항목은 note id 오름차순이고 `age_secs` 는 `now` 에서 `created_at` 을 뺀 값과 같아요. `done` 카드의 요청은 나오지 않아요.
  verify: `cargo test --locked --offline --test card_inbox_cli the_inbox_lists_each_request_family_in_note_order`
- AC-2 (REQ-3, REQ-4): 리더 노트가 요청보다 뒤(note id 큼)에 있으면 그 요청은 사라지고, 앞에 있으면 남아요(경계: 바로 앞 id·바로 뒤 id). 같은 가족의 더 새 노트가 있으면 옛 노트는 나오지 않고 새 것만 나와요. 옛 것에 답이 있고 새 것에 없으면 새 것이 나와요.
  verify: `cargo test --locked --offline --test card_inbox_cli a_later_leader_note_answers_and_a_newer_note_supersedes`
- AC-3 (REQ-5, REQ-6, REQ-10): `--leader boss` 를 주면 `boss` 가 쓴 노트만 답으로 세고 `leader` 가 쓴 노트는 답이 아니에요. 가져온(`origin = import`) 요청은 나오지 않고, 리더가 쓴 `LANE-BLOCKED` 줄은 요청으로 세지 않아요. 빈 값·81자·제어 문자는 `card inbox refused` 로 거절되고(`invalid_input`) 원장은 그대로예요.
  verify: `cargo test --locked --offline --test card_inbox_cli the_leader_name_and_imported_notes_decide_what_counts`
- AC-4 (REQ-8, REQ-9, REQ-11): 명령 전후 원장 파일과 폴더의 목록·크기·수정 시각이 같고, 원장이 없는 프로젝트는 `ledger: "absent"`, `count: 0`, 종료 0 이에요. 원장 자리에 낯선 파일이 있으면 원장의 `blocked` 거절이 그대로 나가고 아무것도 쓰지 않아요. 도움말 첫 문장은 `Reads the ledger only.` 예요.
  verify: `cargo test --locked --offline --test card_inbox_cli the_inbox_reads_only_and_an_absent_ledger_is_empty`
- AC-5 (REQ-1…REQ-11): 이웃 검사 — 접두어 판별과 답·대체 규칙의 단위 시험, 명령 도움말 목록 시험, 형식·clippy.
  verify: `cargo test --locked --offline --lib card::inbox`
  verify: `cargo test --locked --offline --test card_work_cli`
  verify: `cargo fmt --check`
  verify: `cargo clippy --offline --locked --all-targets -- -D warnings`

## Scope

- write: `src/card/inbox.rs`, `src/card/mod.rs`, `src/ops/card.rs`, `tests/card_inbox_cli.rs`, `tests/card_work_cli.rs`, `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261119-leader-inbox/spec.md`
- read-only: `src/card/command.rs`, `src/card/work.rs`, `src/ledger/schema.rs`, `src/ledger/workflow.rs` (노트·`open_read_only` 참조)
- forbidden: 실제 HOME 과 원장 쓰기, 네트워크, 키. 시험은 임시 HOME 만 써요.

조각 1(이 문서)은 `.moai/specs/SPEC-20261119-leader-inbox/**` 만 써요. 조각 2(구현)는 위 write 목록 전체가 필요한데, 카드 C-116 의 예약은 `src/card/inbox.rs`·`tests/card_inbox_cli.rs` 와 이 SPEC 폴더뿐이라 `src/card/mod.rs`·`src/ops/card.rs`·`tests/card_work_cli.rs`·운영 문서는 리더가 범위를 넓혀 주어야 해요(C-43 은 `src/card/command.rs`·`dispatch.rs`·`work.rs` 를 쥐고 있고 위 목록과 겹치지 않아요).

## Plan

1. 이 SPEC 을 쓰고 `spec check` 로 GEARS 오류 0 을 확인한 뒤 가지 `spec/20261119-leader-inbox` 에 push 해요(PR 없음, M-Flow).
2. 시험 먼저(RED): `tests/card_inbox_cli.rs` 가 임시 HOME 의 원장에 카드와 노트를 `card add`·`card note --by` 로 만들고 네 AC 를 단언해요. 명령이 없는 상태에서 먼저 실패시켜요.
3. `src/card/inbox.rs`: 읽기 전용 연결로 `done` 이 아닌 카드의 `origin = 'local'` 노트를 note id 순으로 읽어 가족을 판별하고(`first_line` 접두어), 카드·가족마다 가장 새 노트를 고른 뒤, 같은 카드에 더 큰 note id 의 리더 노트가 있으면 뺀 목록을 `age_secs = now - created_at`(원장 시계 `unixepoch()`)과 함께 돌려줘요.
4. `src/ops/card.rs` 에 `inbox [PROJECT] [--home DIR] [--leader NAME]` 동작을 연결하고, 운영 문서 표와 도움말 목록 시험에 한 줄씩 더해요.
5. 변이(mutation): 답 판정을 `>` 에서 `>=` 로, 가족별 최신 선택을 전체 최신으로, `origin` 걸러내기를 지우기 — 각각 한 번 돌려 시험이 실패하는지 봐요.
6. 실제 원장에 한 번 읽기 전용으로 돌려 하루치 요청(C-101·C-33·C-47)이 나오는지 확인해요.

## Risks

- 리더 판별은 노트의 `author` 문자열(`--by leader`)이에요. 레인이 `--by leader` 로 노트를 쓰면 답으로 세어져요 — 노트는 원래 자기 신고라서 이 명령은 알림용이고 권한을 주지 않아요(D-2).
- "뒤에 리더 노트가 있으면 답"은 거칠어요. 리더가 요청과 무관한 노트를 뒤에 쓰면 요청이 사라져요(거짓 음성). 대신 요청 노트를 정확히 가리키는 약속(`note <id>` 인용)은 리더 노트 310개 중 일부만 따르고 있어서 필수로 걸 수 없어요(D-3).
- 2026-11-19 에 늦게 답한 C-33 노트 2856 은 `LANE-NOTE` 로 시작한 진행 줄 안의 질문이라 이 목록에 나오지 않아요(D-5). 그 줄 모양까지 잡으려면 레인이 질문을 `DECISION-QUESTION` 으로 시작하게 하거나 접두어를 더 넣어야 해요 — 후속.
- 접두어는 한국어 `감사 요청` 을 포함해요. 레인이 다른 표기(`AUDIT-REQUEST`)를 쓰면 안 보여요 — 실제 쓰인 표기를 측정한 목록으로 시작했고 목록 하나(`FAMILIES`)만 고치면 돼요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 노트 종류 대신 첫 줄 접두어로 맞춰요 | `card_note.kind` 는 4값 `CHECK`(`src/ledger/schema.rs:67`), 새 종류는 스키마 변경이고 노트 2,926개에 소급 안 돼요 | 스키마에 종류를 더하고 이 명령이 종류를 읽게 해요 |
| D-2 | 리더는 `author` 가 `leader` 인 노트(`--leader` 로 바꿀 수 있음) | 2026-11-19 측정: 로컬 노트 872개 중 리더 310개의 `author` 가 모두 `leader`, 레인은 세션 id 나 `agent-N` | 세션 id 로 판별 |
| D-3 | 같은 카드에서 더 큰 note id 의 리더 노트가 있으면 답 | 요청 노트를 인용하는 약속은 필수가 아니고(리더 노트 일부만 `note <id> 답` 표기) 시간 순서는 항상 있어요 | 인용이 있는 답만 세요(`--strict`) |
| D-4 | 가족 넷: 감사 요청, `LANE-BLOCKED`, `LANE-READY-INTEGRATION`, `DECISION-QUESTION` | 카드 C-116 의 세 가족 + 측정에서 본 `DECISION-QUESTION` 접두어 | `FAMILIES` 목록 수정 |
| D-5 | `LANE-NOTE` 진행 줄 속 질문은 이번 조각에서 잡지 않아요 | 진행 줄이 질문 줄보다 훨씬 많아서 접두어로 걸면 잡음이 커요 | `LANE-NOTE` 를 가족에 더함 |
| D-6 | `done` 카드와 가져온 노트는 뺍니다 | 닫힌 카드의 옛 `LANE-BLOCKED` 와 GitHub 댓글 2,054개가 잡음이에요 | 상태·출처 걸러내기를 옵션으로 |

## Result
