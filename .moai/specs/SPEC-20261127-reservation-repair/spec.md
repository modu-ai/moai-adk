---
id: SPEC-20261127-reservation-repair
title: "A holder repairs the branch or worktree path its reservation recorded, and card take records a worktree as an absolute path"
status: draft
revision: 1
created: 2026-11-27
---

# A holder repairs the branch or worktree path its reservation recorded, and card take records a worktree as an absolute path

## Intent

예약(reservation)은 `card take` 때 받은 `--worktree` 와 `--branch` 를 그대로 적고, 그 뒤에는 고치는 명령이 없어요. 오늘(2026-11-27) 이 저장소에서 두 가지로 드러났어요.

- `--worktree` 에 상대 경로(`.claude/worktrees/leader-slot-front-r2`)를 주면 예약에 상대 경로가 그대로 저장돼요. 이 예약으로 `card submit` 을 하면 빠른 게이트의 `trailers` 단계가 `git log failed: fatal: cannot change to '.claude/worktrees/leader-slot-front-r2': No such file or directory` 로 막히고, 워크트리 안에서 부르면 `the reserved worktree .claude/worktrees/leader-slot-front-r2 is not a folder` 로 막혀요(C-561 노트 10654, 10657을 2026-11-27 에 읽었어요; 리더는 C-358 과 C-569 의 예약도 같은 모양이라고 했고 그 둘은 제가 읽지 않았어요). 같은 이유로 `card reclaim` 과 `card release` 가 쓰는 `read_worktree` 도 상대 경로를 "확인할 수 없음"으로 거절해요(`src/card/dispatch.rs:1141-1150`).
- 예약에 적힌 브랜치가 틀리면(철자, 또는 브랜치를 새로 잘라 이름이 바뀐 경우) 고치는 길이 없어요. `card --help` 의 어느 명령도 예약의 `branch` 를 바꾸지 않아요. 카드의 `card_branch` 읽기는 `cards_of_branch` 가 예약의 브랜치 이름으로 찾으므로(`src/ledger/reservation.rs:510-518`), 틀린 이름은 정리(cleanup)와 착지 확인도 틀린 이름으로 보게 해요(mo.ai.kr 보고, C-314 의 7번 항목).

이 SPEC 은 (1) `card take` 가 상대 경로를 절대 경로로 바꿔 적거나 거절하게 하고, (2) 예약을 가진 쪽이 브랜치와 워크트리 경로를 원장에 기록되는 사건 한 건으로 고치는 명령을 더해요. 예약을 바꾸는 일이라 tier 2 로 보고, 승인 전에는 구현하지 않아요.

읽는 사람: 승인하는 사용자와 리더, 구현 레인, 감사자.

## Interview

- Q1: 무엇을 원하나요? → A: C-314 카드 본문 7번 항목("no command fixes the branch recorded on a reservation — add a ledger-recorded way for the holder (or leader with reason) to correct it")과 리더가 2026-11-27 에 더한 상대 경로 사례(C-561 노트 10654·10657). 사용자의 말이 아니라 카드 본문과 리더 메시지에서 옮겼어요.
- Q2: 아래 Decisions 의 "승인 때 골라요" 항목(D-1, D-2, D-3) → A: 아직 답이 없어요. 승인 노트의 답을 따르고, 답이 없으면 구현하지 않아요.

## Non-goals

- 예약의 범위(scope), 리스(lease), 세대(generation), 보유자(holder) 는 이 명령으로 바꾸지 않아요. 범위는 `card scope`, 리스는 `card renew`, 보유자는 `card handoff` 와 `card reclaim` 이 맡아요.
- 이미 끝난(released, reclaimed) 예약과 `done` 카드의 기록은 고치지 않아요. 원장의 사건은 지우거나 고치지 않고 덧붙이기만 해요.
- 브랜치나 워크트리를 만들거나 지우거나 옮기지 않아요(git 은 읽기만 해요).
- 다른 카드의 예약을 그 보유자 대신 고치는 길은 만들지 않아요(D-1 에서 리더의 길을 고를 때만 예외예요).

## Requirements

- REQ-1: When `card take` is given a `--worktree` that is not an absolute path, it shall resolve it against the caller's current folder and record the absolute path, and refuse with the resolved path in the message when that folder does not exist or is not the top of a git checkout. · failure: none — 거절은 아무것도 적지 않아요
- REQ-2: When the holder of a live reservation runs `card repair` at its `--generation`, the command shall set the recorded `branch`, `worktree` or both to the given values and write one `card.repair` event that names the old value, the new value, `--by` and `--reason`, changing nothing else of the reservation. · failure: REQ-3
- REQ-3 (REQ-2): IF the new `--worktree` is not an absolute path to the top of a git checkout that exists, or the new `--branch` is not a short branch name git accepts, or no value is given, or the reservation is not live and held by this session, THEN `card repair` shall refuse with a message that names the cause and write nothing.
- REQ-4: When `card repair` records a new branch or worktree, `card show`, `card submit`, `card reclaim` and `cards_of_branch` shall read the new value from then on. · failure: none — 읽는 쪽은 원장의 현재 값만 읽어요
- REQ-5 (REQ-2): IF `--reason` is blank, longer than 500 characters or shaped like a secret, THEN `card repair` shall refuse and write nothing.

## Acceptance

- AC-1 (REQ-1): 임시 프로젝트에서 상대 `--worktree` 로 `card take` 하면 예약에 절대 경로가 적히고(`card show` 의 `reservation.worktree`), 없는 폴더나 체크아웃이 아닌 폴더는 해석된 경로를 대며 거절되고 아무것도 적히지 않아요 · verify: `cargo test --locked --test reservation_repair_cli take_records_a_relative_worktree_as_an_absolute_path`
- AC-2 (REQ-2, REQ-4): 보유자가 `card repair C-n --generation g --branch <새 이름> --reason <글>` 을 하면 `card.repair` 사건이 한 건 쌓이고 `card show` 가 새 브랜치를 보이고 범위·리스·세대·보유자는 그대로예요 · verify: `cargo test --locked --test reservation_repair_cli repair_changes_only_the_recorded_branch`
- AC-3 (REQ-2, REQ-4): 상대 경로로 적힌 예약을 `card repair --worktree <절대 경로>` 로 고치면 `card submit` 의 `trailers` 단계가 그 경로에서 돌아요 · verify: `cargo test --locked --test reservation_repair_cli repair_makes_a_relative_reservation_submittable`
- AC-4 (REQ-3, REQ-5): 보유자가 아닌 세션, 끝난 리스, 상대나 없는 경로, git 이 받지 않는 브랜치 이름, 값이 하나도 없는 호출, 빈 이유 각각이 거절되고 원장이 그대로예요 · verify: `cargo test --locked --test reservation_repair_cli repair_refusals_write_nothing`

## Scope

- write (구현, 승인 뒤): 이 SPEC 폴더, `tests/reservation_repair_cli.rs`, 새 모듈 `src/card/repair.rs`, `src/card/mod.rs`, `src/card/command.rs`(take 의 경로 풀이), `src/ledger/reservation.rs`(고치는 한 줄), `src/ops/card.rs`(명령 선언, 공유 파일: line-lease), `src/help.rs` 와 `tests/fixtures/help/*.txt`(공유 파일: line-lease)
- read: `src/card/dispatch.rs`, `src/card/submit.rs`, `src/ledger/schema.rs`

## Plan

1. 빨간 테스트: `tests/reservation_repair_cli.rs` 에 AC-1~AC-4 를 먼저 써서 실패를 봐요.
2. `card take` 의 `--worktree` 를 절대 경로로 풀고(REQ-1) 검사해요.
3. `card repair` 를 더해요: 보유자와 리스를 확인하고, 값 검사 뒤 한 트랜잭션에서 예약의 `worktree`/`branch` 열과 `card.repair` 사건을 써요.
4. 도움말과 도움말 고정 파일, 연산 문서(`docs/design-review/appendix/operations/card.md`)를 맞춰요.

## Decisions

| Id | Decision | Source |
|---|---|---|
| D-1 | 누가 고치나요: 승인 때 골라요. (가) 예약의 보유자만(추천: 이 SPEC 의 REQ-2 가 그렇게 써 있어요). (나) 보유자와, 이유를 적은 리더(예약이 끝났거나 보유자가 사라진 경우). (나)는 `card reclaim` 이 이미 하는 일과 겹쳐서 이번에는 넣지 않았어요. | C-314 본문 7번 항목 |
| D-2 | 상대 `--worktree` 를 어떻게 하나요: 승인 때 골라요. (가) 호출한 폴더 기준으로 절대 경로로 바꿔 적고 폴더가 체크아웃의 맨 위가 아니면 거절(추천: REQ-1). (나) 상대 경로를 항상 거절하고 절대 경로를 달라고 알려 줘요. (가)는 호출 폴더가 틀리면 엉뚱한 절대 경로가 되지만 REQ-1 이 그 폴더가 실제 체크아웃일 때만 받아요. | 리더 메시지 2026-11-27, `src/card/dispatch.rs:1141-1150` |
| D-3 | 새 명령 이름: 승인 때 골라요. `card repair`(추천) 또는 `card rebranch`. `repair` 는 브랜치와 워크트리를 한 명령에 담아요. | 이 SPEC |
| D-4 | 이미 상대 경로로 적힌 예약(C-358, C-569 등)은 이 명령이 생기면 보유자가 `--worktree` 로 한 번에 고쳐요. 일괄로 고치는 도구는 만들지 않아요. | 리더 메시지 2026-11-27 |

### Research

| Id | Fact | Read |
|---|---|---|
| R-1 | `read_worktree` 는 상대 경로를 "다른 작업 폴더 아래의 다른 폴더라서 여기에 없다는 것이 아무것도 증명하지 않는다"며 거절해요. | `src/card/dispatch.rs:1141-1150`, 2026-11-27 |
| R-2 | 예약의 `branch` 는 길이만 검사하고(`check_optional("branch", ...)`) 저장해요. | `src/ledger/reservation.rs:652-740`, 2026-11-27 |
| R-3 | `cards_of_branch` 는 예약에 적힌 짧은 브랜치 이름으로 카드를 찾아요. | `src/ledger/reservation.rs:510-518`, 2026-11-27 |
| R-4 | 상대 경로 예약으로 `card submit` 이 `git log failed: fatal: cannot change to '.claude/worktrees/leader-slot-front-r2'` 로 거절됐어요. | C-561 노트 10654, 10657, 2026-11-27 |

## Risks

- `card repair` 가 틀린 경로를 받으면 `submit` 이 엉뚱한 폴더의 git 기록을 읽어요: REQ-3 이 체크아웃의 맨 위인 절대 경로만 받아서 막아요.
- 예약을 바꾸는 명령이라 다른 세션이 쓸 수 있으면 남의 카드를 건드려요: REQ-2 가 보유자의 세션과 세대, 달리는 리스를 모두 요구해요.

## Result

초안이에요. 구현하지 않았어요. 사용자의 승인과 D-1~D-3 의 선택을 기다려요.
