---
id: SPEC-20261123-lane-continues-card
title: "A lane continues its own card after /clear or a restart: one adopt command re-registers the lane name and takes the card over at the next generation"
status: draft
revision: 1
created: 2026-11-23
---

# A lane continues its own card after /clear or a restart: one adopt command re-registers the lane name and takes the card over at the next generation

## Intent

레인이 `/clear` 를 받거나 재시작하면 호스트 세션 id 가 바뀌어요. 카드의 예약(reservation)은 옛 세션 id 에 묶여 있어서, 같은 레인의 새 세션이 자기 카드를 이어받지 못하고 임대(lease, 기본 4시간)가 끝나길 기다려요. 카드 본문과 리더 노트 두 개(5074, 5426)가 이 사례를 적었어요: 같은 레인 이름(`agent-34`)으로 다시 등록하려 하면 이전 인스턴스의 등록 임대에 막혔고, 한 카드가 `/clear` 전 세션에 묶여 16시간 동안 회수도 철회도 못 했어요.

### 이미 있는 것 (이 저장소의 팁 `caeac7689a` 에서 읽었어요)

- `card reclaim` 은 임대가 끝난 `held` 예약을 다음 세대(generation + 1)로 넘겨요. 범위·워크트리·브랜치는 그대로이고, 카드의 상태·노트·기준은 그대로예요 (`src/ledger/reservation.rs` `reclaim`, `src/card/dispatch.rs` `reclaim`). 임대가 끝난 뒤에도 워크트리에 커밋 안 한 변경이나 푸시 안 한 커밋이 있으면 `worktree_gate` 가 거절해요.
- 조기 회수(SPEC-20261120-reclaim-dead-holder, 카드 C-189 쪽)는 임대가 남아도 세 증거가 모두 있으면 받아요: 홀더의 `.live` 파일이 있고 아무도 잡지 않으며(`launcher-exited`), 레지스트리 임대가 끝났고, 워크트리가 없거나 깨끗하며 그 안에서 도는 프로세스가 없을 때예요 (`src/card/early_reclaim.rs`).
- `session-register` 는 이름 → 인스턴스 하나, 호스트 세션 id(`session` 필드)를 적어요 (`src/lane_registry.rs` `register`). 같은 이름을 다른 인스턴스가 살아 있는 임대로 쥐고 있으면 `conflict` 로 거절해요.
- `card handoff --to <세션>` 은 홀더가 직접 넘겨요. `/clear` 전에는 새 세션 id 를 모르니 레인이 스스로 쓸 수 없어요.

### 빠진 것 (같은 레인이 `/clear` 뒤 같은 워크트리에서 이어받는 경우)

이어받는 새 세션은 위의 길을 모두 지나지 못해요. 각 줄은 위 코드에서 읽은 결과예요:

1. 이름 재등록이 막혀요: `register` 는 임대가 남은 다른 인스턴스를 `conflict` 로 거절해요. 리더 노트 5074 가 같은 사례를 적었어요.
2. 조기 회수의 프로세스 증거가 막아요: 새 세션 자신이 그 워크트리에서 돌고 있어서 "워크트리 안에서 도는 프로세스 없음" 이 늘 거짓이에요 (`processes_in`).
3. 조기 회수의 `.live` 증거도 막아요: 런처(`moai cc`)가 `/clear` 뒤에도 살아서 이름의 `.live` 잠금을 쥐고 있으니 `alive` 로 답해요. 런처 없이 연 창(Orca 의 창, 그냥 `claude`)은 `.live` 파일이 아예 없어서 "한 번도 런처가 없던 세션은 사라진 게 아니다" 규칙에 걸려요.
4. 임대가 끝난 뒤에도 `worktree_gate` 가 막아요: 같은 레인의 작업 중 변경이 곧 그 워크트리의 변경인데, 이것을 "고아가 될 일" 로 세요.

### 이 SPEC 이 하는 일

- `session-register --continue`: 같은 이름의 옛 항목이 **같은 워크트리**에 등록돼 있고 호스트 세션 id 만 다를 때, 새 세션이 그 항목을 이어받아 옛 id 를 `previous_sessions` 에 남겨요.
- `card adopt C-n --generation <g>`: 새 세션의 등록 항목이 옛 홀더의 id 를 `previous_sessions` 로 담고, 옛 id 를 이름으로 쥔 항목이 없고, 예약의 워크트리가 새 세션의 워크트리일 때, 예약을 한 거래(transaction)로 새 세션에 다음 세대로 넘기고 `card.adopt` 사건 하나를 남겨요. 워크트리의 변경·프로세스는 보지 않아요(같은 레인, 같은 트리).

옛 세션이 사실은 살아 있었다면(같은 워크트리의 같은 이름 중복 세션), 그 세션의 다음 `card check` 는 세대가 낡아서 거절돼요. 쓰기 울타리(fence)가 그대로 막아 주니 데이터는 잃지 않아요.

### 출처 (2026-11-22T22:4xZ 에 읽은 것, 팁 `caeac7689a`)

- 카드 C-268 의 본문, 기준 1, 노트 5074·5426 (`moai card show C-268`).
- `src/ledger/reservation.rs` (`reclaim`, `reclaim_early`, `reclaim_with`: 낡은 예약은 `reclaimed` 로 바뀌고 새 행은 같은 범위·워크트리·브랜치를 이어받아요; 상태 CHECK 는 `held`, `released`, `expired_candidate`, `reclaimed`, `held_for_review` 이고 `reclaim` 은 `card.reclaim` 사건 3개를 평생 한도로 세요).
- `src/card/dispatch.rs`, `src/card/early_reclaim.rs`, `src/lane_registry.rs`, `src/ops/session_register.rs`, `src/ops/card.rs`.
- `src/trailers.rs`: `Moai-Session` 값은 `<세션 id>:<워크트리 키 12자>` 이고 `same_identity` 는 둘 다 전체 형태면 값 전체를 비교해요. 그래서 `/clear` 앞뒤의 같은 레인이 한 신원으로 읽히지 않아요(Non-goals, R-4).
- 스킬 `skills/moai/workflows/mode-agent.md` 의 `Session reset`, `Waiting is switching`.

## Non-goals

- 감사의 독립성 규칙 변경. `/clear` 앞의 세션이 쓴 커밋을 `/clear` 뒤의 같은 레인이 감사해도 다른 세션으로 읽히는 문제는 그대로예요 (R-4, 후속 카드 후보).
- 임대 기본값(4시간), reclaim 한도(3), 원장 스키마 변경. 새 사건 종류 `card.adopt` 는 자유 사건 종류라서 스키마를 건드리지 않아요 (D-3).
- 다른 레인의 세션이나 이름이 다른 세션이 카드를 이어받는 길. 그것은 `card reclaim`, `card handoff`, 리더의 회수예요.
- 조기 회수(C-189)의 증거 규칙 변경.
- 워크플로 문서(`mode-agent.md`), 스킬 문장, 런처의 변경. 코드가 들어간 뒤 글 슬라이스에서 `/clear` 다음에 `session-register --continue` 와 `card adopt` 를 쓰라고 적어요 (Plan 7).
- 이름이 같은 두 세션이 한 워크트리에서 동시에 도는 상황을 막는 일. 이 SPEC 은 그 경우 옛 세션이 카드를 잃도록 두고(울타리), 막지는 않아요.

## Requirements

<!-- This SPEC adds no check that blocks a write that is allowed today; it only adds a path that was refused before. -->
- REQ-1: When a session runs `session-register` with `--continue` for a lane name whose entry names another host session id and the same worktree as the running session, the `session-register` command shall replace that entry with an entry for the running session that lists the replaced session id first in `previous_sessions` and keeps at most eight ids there, newest first.
- REQ-2 (REQ-1): IF the entry names another worktree or no worktree or no session id, or the running session has no host session id or no resolvable git worktree, THEN the `session-register` command shall refuse the `--continue` call with a message that names the missing proof and leave the entry as it was.
- REQ-3: When a session registers a lane name, the `session-register` command shall record the canonical path of the git worktree that holds the running folder in the entry as `worktree` · failure: none — outside a git worktree the field is left out and REQ-2 refuses a later `--continue`.
- REQ-4: When a session runs `card adopt` for a card at its current generation and every proof of REQ-5 to REQ-7 holds, the `card adopt` command shall give the card's reservation to the running session at the next generation with the same write scope, worktree and branch and a new lease from the ledger clock and write one `card.adopt` event.
- REQ-5 (REQ-4): IF the registry entry of the running session is missing or its lease has ended or it does not list the holder session id of the reservation in `previous_sessions`, THEN the `card adopt` command shall refuse with the code `not_same_lane` and write nothing.
- REQ-6 (REQ-4): IF any registry entry still names the holder session id of the reservation, THEN the `card adopt` command shall refuse with the code `holder_registered` that names that lane name and write nothing.
- REQ-7 (REQ-4): IF the worktree the reservation records is not the canonical git worktree of the running folder or the running session's entry records another worktree or no worktree, THEN the `card adopt` command shall refuse with the code `other_worktree` and write nothing.
- REQ-8 (REQ-4): IF the card is `done` or the reservation is not `held` or the offered generation is not the card's current one, THEN the `card adopt` command shall refuse with the answer `card reclaim` gives for the same state and write nothing.
- REQ-9 (REQ-4): IF the reservation is held by another session at the write than the one the proofs were read for, THEN the `card adopt` command shall refuse with the code `holder_changed` and write nothing.
- REQ-10: When the `card adopt` command gathers its proofs, the `card adopt` command shall read the registry and the ledger and the running folder and write none of them · failure: none — a proof that cannot be read is a refusal of REQ-5 to REQ-7 with the reason named.
- REQ-11: The `card adopt` command shall leave the card's state, notes, criteria and links and the lifetime count of `card.reclaim` events as they are.
- REQ-12: When the `card adopt` command adopts a card, the `card adopt` command shall read neither the worktree for changed files, unpushed commits, stash entries or running processes nor the `.live` lock of any lane · failure: none — a thing that is not read cannot fail to be read.

## Acceptance

- AC-1 (REQ-1, REQ-3): after a simulated `/clear` (the same lane name, the same worktree, another host session id in the environment) `session-register --continue` answers `continued`, the entry holds the new session id, the worktree and the old id first in `previous_sessions`; a third call from a third session id lists both old ids newest first; a ninth id drops the oldest · verify: `cargo test --locked --offline --test session_registry_cli ac1_`
- AC-2 (REQ-2): the same call without `--continue` still ends in `conflict`; a `--continue` call for an entry of another worktree, for an entry without a worktree, from a session id-less environment and from a folder outside git each refuse naming the missing proof and the entry file is byte for byte as it was · verify: `cargo test --locked --offline --test session_registry_cli ac2_`
- AC-3 (REQ-4, REQ-11): after the registration of AC-1, `card adopt` moves the reservation to the new session at generation + 1 with the same scope, worktree and branch and a four hour lease; the card's state (also `review`), notes, criteria and links are as before; a `card check` by the old session at the old generation refuses as stale · verify: `cargo test --locked --offline --test card_adopt_cli ac3_`
- AC-4 (REQ-5): a session whose entry does not list the holder (another lane's entry, a fresh registration without `--continue`, an entry with an ended lease, no entry) is refused with `not_same_lane` and the ledger is as it was · verify: `cargo test --locked --offline --test card_adopt_cli ac4_`
- AC-5 (REQ-6): an entry that still names the holder session id, under the same or another lane name, is refused with `holder_registered` naming the lane · verify: `cargo test --locked --offline --test card_adopt_cli ac5_`
- AC-6 (REQ-7): a reservation recorded for another worktree, a symbolic link to the right worktree that resolves to another one and an entry that records no worktree each refuse with `other_worktree` · verify: `cargo test --locked --offline --test card_adopt_cli ac6_`
- AC-7 (REQ-8, REQ-11): a `done` card, a reservation that is released, a lower and a higher generation each refuse with the answer `card reclaim` gives for the same state, and a card that already carries three `card.reclaim` events is adopted · verify: `cargo test --locked --offline --test card_adopt_cli ac7_`
- AC-8 (REQ-9): a reservation that changed holder between the proof and the write is refused with `holder_changed` and the ledger shows no `card.adopt` event · verify: `cargo test --locked --offline --lib ac8_`
- AC-9 (REQ-10): a refused and a successful call leave the registry files and the running folder unchanged except for the ledger rows the success writes, compared by file digests before and after · verify: `cargo test --locked --offline --test card_adopt_cli ac9_`
- AC-10 (REQ-12): an adoption with an uncommitted file, an unpushed commit, a stash entry, a process inside the worktree and a held `.live` lock all present succeeds · verify: `cargo test --locked --offline --test card_adopt_cli ac10_`
- AC-11 (REQ-4): the `card.adopt` event holds the old and the new holder, both generations, the lane name, the worktree and the `previous_sessions` proof; one event per adoption and none for a refusal · verify: `cargo test --locked --offline --test card_adopt_cli ac11_`
- AC-12 (REQ-11): the same card adopted four times in turn by four sessions of one lane keeps the reclaim count at zero and a later `card reclaim` of a dead holder still works up to its own limit of three · verify: `cargo test --locked --offline --test card_adopt_cli ac12_`
- AC-13 (REQ-11): every test of `card_reclaim_cli`, `card_reclaim_early_cli` and `session_registry_cli` that passed before this SPEC keeps passing · verify: `cargo test --locked --offline --test card_reclaim_cli --test card_reclaim_early_cli --test session_registry_cli`

## Scope

- write (this slice): `.moai/specs/SPEC-20261123-lane-continues-card/**`
- write (code slice, after approval):
  - `src/card/adopt.rs` (new), `src/card/mod.rs` (one module line);
  - `src/ledger/reservation.rs` (the adopt transaction beside `reclaim_with`);
  - `src/lane_registry.rs` (the `worktree` and `previous_sessions` fields, `register` with `--continue`, a read function that tells which entry names a session id);
  - `src/ops/session_register.rs` (the `--continue` flag), `src/ops/card.rs` (the `adopt` clap variant and its dispatch line);
  - `tests/card_adopt_cli.rs` (new), `tests/session_registry_cli.rs`, `docs/design-review/appendix/operations/card.md` (the operation line).
- read-only: `src/card/dispatch.rs` (error wording, `card_id`, `open_for_write`), `src/card/early_reclaim.rs`, `src/trailers.rs`, `src/session_current.rs`, `src/ledger/schema.rs`.
- The card's own write scope is empty and the reservation taken for this slice holds the SPEC folder only; `src/card/command.rs` is reserved for another card. The code slice needs the files above added by the leader (`card scope`); none of them is `src/card/command.rs`.
- forbidden: `src/card/early_reclaim.rs` (C-189's proofs stay as they are), the ledger schema (`src/ledger/schema.rs`, `src/ledger/migrate.rs`), any skill, workflow or agent text, the launcher (`src/session.rs`).

## Plan

1. (this slice) The SPEC, status draft; the user's approval and the open decisions U-1 and U-2 are answered by the user, never written by a lane.
2. After approval, tests first (RED): `tests/session_registry_cli.rs` AC-1, AC-2; `tests/card_adopt_cli.rs` AC-3 to AC-7, AC-9 to AC-12, on a temporary home and a temporary repository with a second worktree; the lib test AC-8 for the holder change inside the transaction.
3. `src/lane_registry.rs`: the entry fields and the `--continue` branch of `register`; `src/ops/session_register.rs`: the flag; the answer word `continued`.
4. `src/ledger/reservation.rs`: the adopt transaction (the reclaim steps without the lease rule and the reclaim count, one `card.adopt` event).
5. `src/card/adopt.rs` and the `card adopt` command: the proofs in the order of REQ-5, REQ-6, REQ-7, each read-only.
6. Dangerous-path run in a throwaway for the READY: a real second worktree, a real second session id in the environment, a real `/clear`-shaped sequence of the commands, and the stale refusal of the old id.
7. Text slice, after the code lands: `mode-agent.md` (`Session reset`, `Waiting is switching`: after a `/clear` run `session-register --continue` and `card adopt`) and the registry help text; a separate card, the audit rule text of tier 2 applies to the code slice.
8. Targeted tests, then one `check affected`, then READY with head, base and the receipt.

## Risks

- R-1: Two live sessions with one lane name in one worktree (a person opens a second window by mistake) let the later `--continue` take the name and the card; the first session loses its writes at the next `card check` (stale generation, the fence of `card check`) and its next `session-register` renewal fails. No data is lost, and nothing can stop it from outside because both sessions look the same to the registry. A person who sees the refusal in the first session decides which one to keep.
- R-2: `previous_sessions` and `worktree` are plain text in an unsigned local file, like the rest of the registry; anyone who can write the clone's registry can forge a continuation. The registry is under the clone's common git dir and not written from the repository. The generation fence and the card's lease stay as the check on every write.
- R-3: A path that is a symbolic link to the right worktree must not pass `other_worktree`; the proof compares canonical paths of both, and AC-6 covers the link.
- R-4: Audit independence reads the full `Moai-Session` value, so the same lane before and after a `/clear` counts as two sessions in `src/audit_gate.rs` and in the landing rules. An adopted card does not make that worse, but it does not fix it either. `previous_sessions` is the record a later card can read for it.
- R-5: Adoption carries no cap. A loop that clears and adopts without end writes one event each time; the lease is renewed to four hours each time. The `card.adopt` events make that visible in the ledger; a cap is added when it is seen.
- R-6: The old session id stays in the entry only up to eight ids. A card held by a session older than the eighth `/clear` is not adopted and goes the `card reclaim` way (lease end or the early proof).
- R-7: A restart that does not change the worktree but registers under a new lane name finds no entry to continue; the new name is a new lane (`not_same_lane`), by design.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The continuation is explicit: `session-register --continue` and `card adopt` are two commands the lane runs, not a hidden rule of `register` or `take`. | The registry's conflict refusal protects a lane name from a second live session; an explicit flag shows in the transcript and in the entry that the lane chose to continue. | Make `register` continue by itself when the worktree matches. |
| D-2 | The proof of "same lane" is the registry's own record: the running session's entry lists the holder in `previous_sessions`, and no entry names the holder any more. | The card says "registered lane name equals the old holder's lane name, … the old session id is no longer registered live"; the registry is the one place that maps session ids to lane names, and a replaced entry would otherwise lose the old id. | Store the lane name on the reservation (schema change, a later SPEC). |
| D-3 | The adoption is its own event kind `card.adopt` and the old reservation becomes `reclaimed`; no new status and no schema change. | `reservation.status` has a CHECK list that a new word would need a migration for; events have a free `kind` (as `card.verdict` and `card.land_started` use). A separate event kind keeps the reclaim count of three clean. | A `adopted` status with a migration. |
| D-4 | The worktree, the processes and the `.live` lock are not read. | The same lane in the same tree is the one who has the work; the card text says "no dirty-worktree check (same lane, same tree)". The fence stops a mistaken adoption from writing. | Add the worktree gate back as a warning in the answer. |
| D-5 | The eight-id cap on `previous_sessions`. | A file the registry reads is limited to 4096 bytes (`MAX_ENTRY_BYTES`); eight ids of 36 characters stay far under it. | Raise the cap with the byte bound. |
| D-6 | `card adopt` also adopts a card in `review`; the card's state is not changed. | The card in `review` is the case where the new session can neither release nor take the card (`Waiting is switching`). | Refuse `review` and send it through `card reopen`. |

### Open decisions of the user

- U-1: whether `--continue` is allowed while the old entry's lease is live and `<name>.live` is held by a running launcher (the normal `/clear` case under `moai cc`). This SPEC says yes, because the lock cannot tell a `/clear` from a second session (R-1). The other option: refuse while the lock is held and tell the person to run `session-register --release` before the `/clear`; a lane cleared by the leader without a release then waits for the lease. Recommended: yes. Reversible by a later change.
- U-2: whether a lane may adopt a card whose `Moai-Claim` or card reservation was reclaimed by the leader in the meantime. This SPEC says no: the reservation at the offered generation must be the current one, and a reclaimed card has a new generation (REQ-8). No decision is needed unless the user wants the opposite.

## Result

Status: SPEC draft (revision 1). Nothing is implemented and no acceptance line is met. The base is tip `caeac7689a`. The SPEC carries no approval; U-1 and U-2 wait for the user.
