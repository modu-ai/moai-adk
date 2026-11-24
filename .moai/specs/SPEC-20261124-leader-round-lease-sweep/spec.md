---
id: SPEC-20261124-leader-round-lease-sweep
title: "the leader round sweeps ended leases once before it builds its answer, and only lists them when the profile turns the sweep off"
status: draft
revision: 1
created: 2026-11-24
---

# the leader round sweeps ended leases once before it builds its answer, and only lists them when the profile turns the sweep off

## Interview

- Q1: 무엇이 재발했어요? -> A: `moai card sweep` (SPEC-20261124-ended-lease-sweep) 는 있지만 아무도 정해진 때에 부르지 않아요. 2026-11-24 에 리스가 끝난 카드 19장이 손으로 풀렸고, 그 카드들이 80~100시간 사이클 꼬리를 만들었어요 (카드 C-469 본문).
- Q2: 라운드는 무엇을 하는 명령이에요? -> A: `moai leader-round` (`src/ops/leader_round.rs`) 가 레인 관찰 파일과 GitHub 스냅샷을 읽고 레인마다 한 동작과 공유 부분(`round`)을 답해요. `leader_round::round` 는 `lane_next.rs` 의 테스트와 HOME-GUARD 기준표가 쥐고 있어서 서명을 바꾸지 않아요.
- Q3: sweep 안쪽을 바꾸나요? -> A: 아니요. 이 SPEC 은 `card::sweep::sweep` 를 부르기만 해요. C-117 (하트비트 회수) 이 sweep 에 더하는 경로와 겹치지 않아요.
- Q4: 프로필에서 끄면 무엇이 달라요? -> A: 새 프로필 키 `lease_sweep_in_round` (bool, 없으면 켜짐) 가 `false` 이면 라운드는 `--dry-run` 으로만 불러서 목록을 답하고 원장에 쓰지 않아요.

## Intent

리스가 끝난 지 오래된 카드가 `active` 로 남으면 일이 배차되지 않아요. `moai card sweep` 가 그 카드를 `ready` 로 돌리지만, 리더가 라운드마다 따로 불러야 해요. 이 SPEC 은 라운드가 답을 만들기 전에 sweep 을 한 번 부르게 해서 그 손 단계를 없애요.

읽는 사람: 라운드를 도는 리더, 라운드 답을 읽는 레인, 프로필 키를 문서에 옮기는 카드.

## Non-goals

- `card sweep` 안쪽의 규칙(유예, 건너뛰는 이유, 노트 내용). 그것은 SPEC-20261124-ended-lease-sweep 의 몫이고, 하트비트 경로는 C-117 의 몫이에요.
- 라운드에서 `card withdraw`, `card reclaim` 이나 다른 원장 쓰기를 부르는 일.
- 라운드 답의 다른 부분(`lanes`, `round`, `idle_lanes`)을 바꾸는 일.
- 라운드를 일정에 맞춰 돌리는 배선(`internal/loop`).

## Requirements

- REQ-1: When the leader round runs, the round shall call `moai card sweep` once with dry-run off before it builds the rest of its answer.
- REQ-2 (REQ-1): IF the profile key `lease_sweep_in_round` is false, THEN the round shall call the sweep with dry-run on and write nothing to the ledger.
- REQ-3 (REQ-1): When the sweep returns or lists cards, the round's answer shall name each card under `lease_sweep` with its holder, generation and lease end. · failure: none — the part is built from the sweep's own answer, and an empty answer names no card.
- REQ-4 (REQ-1): IF the project has no profile, the ledger is absent, HOME is not set or the sweep fails, THEN the round shall answer `lease_sweep` as `not_run` or `error` with the reason and still answer its other parts.
- REQ-5 (REQ-3): When a card was returned, the round's `writes` shall be true. · failure: none — `writes` is set from the same list the answer names, so it cannot disagree with it.
- REQ-6: When the profile key `lease_sweep_in_round` is absent, the round shall sweep as if it were true. · failure: none — absent and the default read the same, so a profile written before the key behaves as before.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): in a temporary ledger, a round run once returns a card whose lease ended 25 hours ago to `ready`, names it under `lease_sweep.cards` and answers `writes: true`, and a second round returns nothing · verify: `cargo nextest run --locked --offline --test leader_round_sweep_cli round_sweeps_`
- AC-2 (REQ-2, REQ-6): with `lease_sweep_in_round` false the same card stays `active` with every ledger row as it was and the answer lists it under `lease_sweep.cards` with the status `listed`, a profile that cannot be read lists only, and without the key or with `true` the card is swept · verify: `cargo nextest run --locked --offline --test leader_round_sweep_cli profile_`
- AC-3 (REQ-4): a project with no profile, a project whose ledger is absent and a HOME that is not set each leave `lease_sweep` as `not_run`, a ledger file that is not a database leaves it as `error`, each with the reason, and `lanes` and `round` are still answered · verify: `cargo nextest run --locked --offline --test leader_round_sweep_cli not_run_ error_`
- AC-4 (REQ-1): a card inside the grace period, a card in `review` and a renewed lease stay as they are and appear as skipped, so the round calls the sweep and does not widen it · verify: `cargo nextest run --locked --offline --test leader_round_sweep_cli skipped_`
- AC-5 (REQ-5): `moai check affected` passes at the head of the card · verify: `moai check affected <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261124-leader-round-lease-sweep/spec.md`
- write: `src/leader_round.rs`, `src/ops/leader_round.rs`, `tests/leader_round_sweep_cli.rs`
- write: `src/profile.rs` (shared file of the project profile: written with the line-lease command, not held in the card scope)
- write: `docs/design-review/appendix/operations/leader_round.md` (line-lease; not a shared file, the leader approves it with this SPEC)
- read-only: `src/card/sweep.rs`, `src/ledger/sweep.rs`, `src/lane_next.rs`
- forbidden: `src/card/sweep.rs`, `src/ledger/sweep.rs` (C-117 adds the heartbeat path there), the real home ledger (tests use a temporary HOME only)

## Plan

1. Tests first (RED) in `tests/leader_round_sweep_cli.rs` against a temporary HOME: AC-1 to AC-4, seen failing on the unchanged code.
2. `src/leader_round.rs`: one function `lease_sweep(project, home, enabled)` that calls `card::sweep::sweep` once and builds the `lease_sweep` part; `round()` is not touched.
3. `src/ops/leader_round.rs`: read HOME with `ledger::command::resolve_home`, call `lease_sweep` before `round()`, add the part to the answer and set `writes`.
4. `src/profile.rs`: the key `lease_sweep_in_round`, written by line-lease.
5. The long help text and the operation document name the sweep (the command no longer "reads only"); the help goldens carry only the first line of the command and do not move.
6. Guard mutants for the dry-run switch, the call itself and the `writes` flag; `moai check affected` bare once at READY.

## Risks

- A round is no longer read-only: it can end a lease and return a card. The grace (24 hours by default) and the sweep's own skips bound that, and the profile key switches it off.
- The round's `--now` does not reach the sweep, which reads the database clock as `card sweep` does; a test pins ledger times by the database, not by `--now`.
- A test that would touch the real ledger: every test opens a temporary HOME as `tests/card_sweep_cli.rs` does.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The switch is a top-level profile key `lease_sweep_in_round`, absent means on | the key sits beside `lease_sweep_hours`; `leader` settings have a required field and a project with no `leader` entry could not set it | rename or fold the key in a follow-up SPEC |
| D-2 | The call lives in `src/ops/leader_round.rs` before `round()`; `round()` keeps its signature | `lane_next.rs` and the HOME-GUARD baseline use `round()`; the reader of HOME stays the existing `resolve_home` | move the call into `round()` |
| D-3 | A failing sweep never fails the round | the round also answers lane actions the leader needs | make the failure an error |

## Result

Not yet written.
