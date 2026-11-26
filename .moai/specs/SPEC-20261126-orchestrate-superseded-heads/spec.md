---
id: SPEC-20261126-orchestrate-superseded-heads
title: "`moai card orchestrate plan` 이 낡은 PASS 머리를 착지 대기열에서 빼 따로 보여줘요"
status: draft
revision: 1
created: 2026-11-26
parent: SPEC-20261124-orchestrate-plan
---

# 낡은 PASS 머리는 착지 대기열 밖에 두기

## Intent

`moai card orchestrate plan` 의 `land_queue` 는 PASS 판정을 받고 아직 착지하지 않은 머리를 모두 싣고, 카드의 지금 머리가 아니면 `current: false` 만 붙여요. 그 줄에도 `moai card land` 명령이 그대로 붙어 있어서, 대기열을 위에서부터 읽으면 낡은 머리를 착지하라는 줄이 섞여 있어요. 2026-11-26 에 설치본 `moai card orchestrate plan` 을 실제 원장에 한 번 돌려 `jq` 로 센 값은 대기열 31 건 가운데 `current: false` 8 건이에요.

앞 조각의 감사(AUDIT #348, 노트 8160)는 두 가지를 비차단 지적으로 남겼어요. S2: AUDIT 노트 첫 줄의 대소문자 검사를 고정하는 테스트가 없어서, 그 검사를 지운 뮤턴트가 살아남았어요. S4: 낡은 PASS 머리가 대기열에 남아요. 카드 기준 3 이 이 두 가지를 닫으라고 해요. 이 SPEC 은 그 기준만 다뤄요.

## Interview

- Q1: 이번 조각은 어디까지인가요? -> A: 카드 기준 3 만이에요. 기준 1·2 의 `run` 은 다음 조각에서 따로 승인을 요청하고, 레인 시작은 별도 카드로 나눴어요(리더 결정 노트 10183).
- Q2: 낡은 머리를 지우나요, 따로 보여 주나요? -> A: 따로 보여 줘요. 착지하는 쪽이 만든 머지 머리처럼 넘김 노트 없이 판정만 난 머리도 `current: false` 로 읽혀서, 지우면 착지할 머리가 말없이 사라질 수 있어요(D-1).

## Non-goals

- `plan` 이 쓰기, 보내기, 시작, 착지를 하는 일. 계획은 계속 읽기만 해요.
- 지금 머리를 정하는 규칙(가장 새 넘김 노트의 머리, 없으면 가장 새 판정의 머리)을 바꾸는 일.
- 판정을 읽는 방식과 `moai card land` 의 검사를 바꾸는 일.
- 스킬과 워크플로 글.

## Requirements

- REQ-1: When `moai card orchestrate plan` runs, the command shall print a `superseded` list beside `land_queue` · failure: none — a project without a ledger prints both lists empty.
- REQ-2: When a head qualifies for the landing queue (judged `pass`, not landed, reviewer outside the card's author sessions) and differs from the card's current head, the command shall list it in `superseded` and leave it out of `land_queue` · failure: none — a head that does not qualify is listed in neither.
- REQ-3 (REQ-2): The command shall give each `superseded` entry the card, the head, the reviewer, the source, the tier, the event or note number, `current: false` and a key, and no `command` · failure: none — only a `land_queue` entry carries a `card land` line.
- REQ-4: IF the first line of a note begins with any other spelling of `AUDIT ` than that exact one, THEN the command shall read the note as no verdict · failure: none — the head stays unjudged and is listed for audit.
- REQ-5: When `moai leader brief` runs, the command shall build `queue` from `land_queue` only and report the number of `superseded` entries as `counts.superseded` · failure: none — a project without a ledger reads 0.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): on a temporary ledger a card with a `pass` head that a newer hand-back note replaced is listed in `superseded` without a `command`, the current `pass` head is in `land_queue`, a replaced head judged `fail` or judged by an author session is in neither list, and a ledger with nothing replaced prints `superseded: []` · verify: `cargo test --locked --offline --test card orchestrate_cli::superseded_`
- AC-2 (REQ-4): a first-line note reading `audit … verdict: PASS` leaves the head listed for audit and out of both lists · verify: `cargo test --locked --offline --test card orchestrate_cli::judged_a_lowercase`
- AC-3 (REQ-5): `brief` on a ledger with one current and one replaced `pass` head lists one queue item and `counts.superseded` 1 · verify: `cargo test --locked --offline --test leader_brief_cli superseded_`
- AC-4: the neighbour checks pass: the help fixtures, the card tests and the style checks · verify: `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings` and `moai check affected <base>..<head>` at the READY head

## Scope

- write: `.moai/specs/SPEC-20261126-orchestrate-superseded-heads/**`, `src/card/orchestrate.rs`, `tests/card/orchestrate_cli.rs`, `src/leader_brief.rs`, `tests/leader_brief_cli.rs`, `docs/design-review/appendix/operations/leader.md`
- shared files, by line lease only: `src/ops/card.rs` (help text), `docs/design-review/appendix/operations/card.md` (the row of the command)
- read-only: `src/card/landed.rs`, `src/gate.rs`, `src/ledger/**`
- forbidden: a real `claude` or `orca` process, any model call, any network, the real HOME, any ledger write by the plan

## Plan

1. 빨간 테스트 먼저: `superseded_` 와 `judged_a_lowercase`, 그다음 brief 의 `superseded_`.
2. `src/card/orchestrate.rs` 에서 대기열 항목을 지금 머리와 아닌 머리로 나눠요. `src/leader_brief.rs` 는 `old` 갈래를 지우고 개수를 더해요.
3. 도움말 줄과 문서 행, 가드 뮤턴트, `check affected` 한 번, READY.

## Risks

- 착지하는 쪽이 만든 머지 머리는 넘김 노트가 없어서 `superseded` 에 실려요. 지워지지 않으니 보이지만, `card land` 줄은 붙지 않아요. 착지 자체는 `moai card land` 의 검사가 정해요.
- 넘김 노트가 없는 카드는 가장 새 판정의 머리를 지금 머리로 보는데, 판정 이벤트가 노트보다 앞서 읽혀요. 낡은 이벤트와 새 노트가 함께 있으면 새 머리가 `superseded` 로 갈 수 있어요. 이 규칙은 이번에 바꾸지 않아요.

## Decisions

| ID | Decision | Evidence | Alternative |
| --- | --- | --- | --- |
| D-1 | 낡은 머리는 지우지 않고 `superseded` 에 따로 실어요. | 카드 기준 3 은 두 방법을 다 허용해요. 머지 머리는 넘김 노트 없이 판정될 수 있어요. | 대기열에서 지우기만 해요. |
| D-2 | `superseded` 항목에는 `command` 를 싣지 않아요. | 낡은 머리를 착지하라는 줄이 이 카드가 고치려는 것이에요. | 명령을 그대로 두고 표시만 해요. |
| D-3 | `leader brief` 는 대기열에서 낡은 항목을 빼고 개수만 보여 줘요. | brief 는 plan 의 답을 줄여 만들고 4096 바이트 상한이 있어요(SPEC-20261124-leader-brief REQ-1, D-2). | brief 의 대기열에 `old` 표시로 계속 실어요. |

## Result

2026-11-26 에 레인이 잰 값이에요. 구현 커밋은 `c9202b96e6`, 문구 커밋은 `455cc3605d` 와 그 뒤의 문서 커밋이에요.

- 빨간 테스트(`a37e9ae36c`): `--test card orchestrate_cli::` 는 33 통과·8 실패, `--test leader_brief_cli` 는 4 통과·5 실패였어요. 대소문자 테스트는 처음부터 통과해요(지금 동작을 고정하는 테스트예요).
- AC-1 met (REQ-1, REQ-2, REQ-3): 적힌 확인 명령이 종료 코드 0, 5 통과.
- AC-2 met (REQ-4): 적힌 확인 명령이 종료 코드 0, 1 통과. 코드는 바꾸지 않았고, 첫 줄 검사를 지운 뮤턴트가 이 테스트에서 죽어요(앞 감사의 S2).
- AC-3 met (REQ-5): 적힌 확인 명령이 종료 코드 0, 2 통과.
- AC-4 not_run here: 이웃 검사는 READY 노트의 `check affected` 영수증이에요.
- 가드 뮤턴트 4개(나누기 제거, 낡은 항목에 `command` 남기기, `AUDIT ` 검사 제거, 개수 0 고정)가 모두 종료 코드 101 이에요.
- 남은 글: `src/card/mod.rs` 의 모듈 설명과 `moai leader brief` 도움말은 `superseded` 를 따로 말하지 않아요. 틀린 문장은 아니고 이번 범위 밖이에요.
