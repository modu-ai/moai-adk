---
id: SPEC-20261123-leader-registration-lease
title: "A coordination window registers as leader with its host session id, refuses a second live one and renews each round"
status: draft
revision: 1
created: 2026-11-23
---

# A coordination window registers as leader with its host session id, refuses a second live one and renews each round

## Intent

사용자 결정(2026-11-22T23:49Z): `/moai leader` 는 상주 역할이 아니라 "등록된 조정 창"이에요. 친 창이 조정 창구로 등록되고(호스트 세션 id 포함, 중복 거절), 레인은 그 창의 세션 id 를 읽어 완료 보고를 보내요(다음 카드의 핸드셰이크). 지금은 셋 다 비어 있어요: 실제 `leader.json` 에는 `session` 필드가 없고 임대가 2026-11-18 에 끝나 있어요; `session-register` 는 `--mode leader` 가 세션 id 없이도 등록되고, 이름이 다른 두 번째 leader 항목도 막지 않아요; 갱신을 빼먹어도 아무 곳에도 보이지 않아요(등록부의 공개 읽기는 끝난 임대를 `free` 로 접어서 언제 끝났는지와 mode 가 사라져요).

이 SPEC 은 다섯 가지를 해요. (1) leader 등록은 호스트 세션 id 가 있을 때만 받고 항목에 적어요. (2) 임대가 살아 있는 leader 항목이 이미 있으면 다른 이름이든 다른 인스턴스든 거절해요. (3) 레인이 `--show` 와 `--liveness-of` 로 살아 있는 조정 창의 세션 id 를 읽게 해요. (4) `leader-round` 가 라운드마다 leader 항목의 임대를 읽어 보이고(끝난 임대는 끝난 지 몇 초인지까지), `--renew-instance` 를 주면 그 라운드에서 갱신까지 해요. (5) 갱신이 빠진 라운드는 `registration` 부분이 `expired` 로 눈에 띄어요.

등록부 임대는 생존 증명이 아니에요(런처가 있으면 `.live` 잠금이 증명이고, 설계 3.1절). 이 SPEC 은 런처 없이 열린 조정 창이 임대로만 살아 있음을 보이게 하는 길을 더해요.

## Non-goals

- 등록 항목 스키마 변경: schema 1 그대로예요. `session` 은 이미 있는 선택 필드예요.
- `/moai leader` 경로 글과 스킬·워크플로 문장: 경로를 쓰는 카드의 몫이에요. 이 SPEC 은 그 글이 부를 명령만 만들어요 (`session-register --mode leader`, `leader-round --renew-instance`).
- 완료 보고 핸드셰이크의 글과 전송(다음 카드).
- 레인 이름(`agent-N`) 등록과 `--continue`(레인 이어받기)의 변경.
- 임대 기본값 4시간, 회수 한도 3, 원장 스키마 v1, 호스트 훅.
- 실제 HOME 의 등록부와 네트워크: 모든 시험은 임시 저장소와 임시 HOME 에서 돌아요.

## Requirements

- REQ-1: When a session registers with mode leader, the `session-register` command shall write the host session id of the running session into the entry as `session`.
- REQ-2 (REQ-1): IF the running session has no host session id, THEN the `session-register` command shall refuse a leader registration naming the missing id and write nothing.
- REQ-3: When a session registers with mode leader while another entry of mode leader holds a lease that runs past now under any name, the `session-register` command shall refuse the registration with a conflict that names that entry and its lease end, never its instance, and write nothing · failure: none — the refusal is the outcome.
- REQ-4: When the instance that holds a leader entry renews it from a session whose id differs from the `session` of the entry, the `session-register` command shall refuse the renewal and leave the entry as it was · failure: none — the refusal is the outcome.
- REQ-5: When `session-register --show` reads a held entry that carries a session id, the `session-register` command shall print that id as `session` · failure: none — an entry without one prints as before.
- REQ-6: When `session-register --liveness-of` is asked for the session id of an entry of mode leader whose lease runs past now and which has no `.live` lock file, the `session-register` command shall answer `alive` with the registry lease as its evidence · failure: none — any other state keeps its existing answer.
- REQ-7: When `leader-round` answers, its shared object shall carry a `registration` part that lists every registry entry of mode leader with its name, lease end, session id when it has one and either the seconds left or the seconds since the lease ended · failure: none — REQ-9 covers an entry that cannot be read.
- REQ-8: The `registration` part shall carry the status `absent` while no leader entry exists and `ok` while every listed lease runs past now and every entry has a session id and `expired` while any listed lease has ended and `no_session` while no lease has ended and an entry lacks a session id · failure: none — the status is the outcome.
- REQ-9 (REQ-7): IF an entry of the registry cannot be read, THEN the `registration` part shall be `error` with that message alone and leave every other part of the round as it was.
- REQ-10: When `leader-round` is given `--renew-instance` with the instance of a leader entry, the `leader-round` command shall renew that entry for the default lease from the running session and report the renewal in the part before it lists the entries.
- REQ-11: While `leader-round` is not given `--renew-instance`, the `leader-round` command shall write nothing and answer `writes` false.
- REQ-12: The `registration` part shall name `session-register` and `--instance` and `leader-round --renew-instance` in its note · failure: none — the note is text.
- REQ-13: The `registration` part shall be reported without `--repo` and without `--snapshot` · failure: none — the part reads only the local registry.
- REQ-14 (REQ-10): IF the renewal is refused, THEN the `leader-round` command shall report `renewal` `refused` with the message and answer every other part of the round as it would without the flag.

## Acceptance

- AC-1 (REQ-1, REQ-2): a leader registration with a session id in the environment writes `session` into the entry and a registration without one is refused naming the id, with the registry folder byte for byte as it was · verify: `cargo test --locked --offline --test leader_registration_cli ac1_`
- AC-2 (REQ-3): a second leader registration under another name and one under the same name by another instance are each refused naming the first entry and its lease end without its instance; a registration after the first lease ended is accepted; an `agent` mode registration is never refused by a leader entry · verify: `cargo test --locked --offline --test leader_registration_cli ac2_`
- AC-3 (REQ-4): a renewal by the holding instance from the same session id is accepted and keeps `session`; a renewal from another session id is refused and the entry is as it was · verify: `cargo test --locked --offline --test leader_registration_cli ac3_`
- AC-4 (REQ-5, REQ-6): `--show` prints `session` of a held leader entry; `--liveness-of` with that session id answers `alive` without a `.live` file while the lease runs and `unknown` or `launcher-exited` as before once it has ended · verify: `cargo test --locked --offline --test leader_registration_cli ac4_`
- AC-5 (REQ-7, REQ-8): the `registration` part is `expired` and names the entry with the elapsed seconds for a lease that ended 38 hours earlier, `ok` for a live lease with a session id, `no_session` for a live lease without one and `absent` without a registry folder · verify: `cargo test --locked --offline --test leader_round_cli registration_`
- AC-6 (REQ-9): an unreadable entry makes the part `error` and the other shared parts answer as they did · verify: `cargo test --locked --offline --test leader_round_cli registration_`
- AC-7 (REQ-10, REQ-14): `--renew-instance` turns the 38 hour old entry into a live one in the same answer, keeps its session id, answers `writes` true and reports `renewal` `renewed`; a wrong instance reports `refused` · verify: `cargo test --locked --offline --test leader_round_cli registration_`
- AC-8 (REQ-11): a round without the flag leaves the whole clone state folder byte for byte as it was · verify: `cargo test --locked --offline --test leader_round_cli registration_`
- AC-9 (REQ-12, REQ-13): the note names the three words and the part is present while the other shared parts are `not_run` · verify: `cargo test --locked --offline --test leader_round_cli registration_`
- AC-10: the neighbour checks and the style checks pass · verify: `cargo test --locked --offline --test session_registry_cli --test launcher_liveness_cli --test leader_pulse_cli`

## Scope

- write (this slice, SPEC and the files nobody else holds): `.moai/specs/SPEC-20261123-leader-registration-lease/**`, `src/leader_round/registration.rs` (new, a child module of `src/leader_round.rs`), `src/leader_round.rs`, `src/ops/leader_round.rs`, `tests/leader_registration_cli.rs` (new), `tests/leader_round_cli.rs`, `docs/design-review/appendix/operations/leader_round.md`
- write (shared files, taken only after the card that holds them lands): `src/lane_registry.rs` (the leader rules inside `register`, `session` in `show`, one visibility word on `read_entry`), `src/ops/session_register.rs` (the `--liveness-of` leader answer), `tests/session_registry_cli.rs` (the leader tests gain a session id in their environment) and `src/help.rs` with `tests/fixtures/help/**` only if a help text changes.
- read-only: `src/session_current.rs`, `src/session.rs`.
- forbidden: the real HOME's registry, the network, host hooks, the registry entry schema, the ledger, any skill or workflow text.

## Plan

1. (this slice) The SPEC and the RED tests: `tests/leader_registration_cli.rs` AC-1 to AC-4 and `tests/leader_round_cli.rs` AC-5 to AC-9.
2. `src/leader_round/registration.rs`: the leader rules as plain functions over the entries the registry reads (the conflict scan, the session rule, the registration part, the leader liveness), so the shared files take only short call sites.
3. `src/leader_round.rs` and `src/ops/leader_round.rs`: the `registration` part and the `--renew-instance` flag; GREEN for AC-5 to AC-9.
4. After the holder of the registry files lands: the call sites in `src/lane_registry.rs` and `src/ops/session_register.rs`; GREEN for AC-1 to AC-4; the existing leader-mode tests of `tests/session_registry_cli.rs` gain a session id in their environment.
5. Targeted tests, one `check affected`, a run of the dangerous path in a throwaway (a real registry, a real second registration, a real 38 hour old entry), then READY with head, base and the receipt.

## Risks

- R-1: The leader-mode tests that exist today register without a session id in a cleared environment; REQ-2 turns them into refusals. They gain `CLAUDE_CODE_SESSION_ID` in step 4 and keep their other assertions.
- R-2: Two files this SPEC needs are held by another card for now (`src/lane_registry.rs`, `src/ops/session_register.rs`). The new rules live in `src/leader_round/registration.rs` so the shared files change by a few call lines that apply on top of that card's head.
- R-3: A `/clear` of the coordination window changes its session id and loses the instance; the new window cannot register until the lease ends or the old entry is released. This SPEC leaves that to the lane-continuation card and the release command.
- R-4: A leader entry without `.live` is alive only by its lease: a coordination window that stops renewing reads as alive until the lease ends (at most four hours). The round shows the seconds left, so the gap is visible.
- R-5: `--renew-instance` writes from a command that was read-only; it writes only when the flag is given, and the answer says so in `writes`.

## Result

Status: SPEC draft (revision 1). Nothing is implemented and no acceptance line is met. The SPEC carries no approval; the user approves it.
