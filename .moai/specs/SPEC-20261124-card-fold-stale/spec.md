---
id: SPEC-20261124-card-fold-stale
title: "card fold, 오래된 카드 stale 목록, 만료된 held 카드 withdraw"
status: draft
revision: 1
created: 2026-11-24
---

# card fold, 오래된 카드 stale 목록, 만료된 held 카드 withdraw

## Interview

- Q1: 카드 정리에서 리더가 손으로 한 일은 무엇이에요? -> A: (a) 같은 일을 하는 카드 24장을 하나로 합치고 나머지를 `card withdraw` 의 이유 글에 "C-m 으로 합쳤어요" 라고 적었어요. (b) 착지한 head 가 있는데 닫히지 않은 카드, lease 가 끝난 채 남은 held 카드, 아무도 쥐지 않은 review 카드, 오래 놓인 draft 카드를 `card list` 를 눈으로 훑어 찾았어요. (c) 쥔 예약의 lease 가 끝난 카드는 `card withdraw` 가 거절해서 `card reclaim` 으로 한 번 쥔 뒤에야 닫았어요. 셋 다 도구가 한 번에 해야 할 일이에요 (사용자 선택 2026-11-24, 카드 C-444).
- Q2: 새 카드 상태나 새 스키마가 필요해요? -> A: 아니요. `card` 표의 상태 값은 CHECK 로 여섯 개가 고정이고(`src/ledger/schema.rs:106-110` 의 전이 표), `card_link.kind` 는 이미 `relates` 를 받으며(`src/ledger/schema.rs:80`), 사건 종류는 자유 문자열이에요(`card.withdraw` 도 그렇게 써요). 접힌 카드는 `hold` 에 `card.fold` 사건 한 줄과 `relates` 링크 한 줄을 달아 표시해요.
- Q3: `card next` 가 이미 `already_landed` 로 말하는 것과 겹치지 않아요? -> A: `already_landed` 는 `ready` 카드만 봐요(`src/card/landed.rs:405-423`). 이 SPEC 의 `stale[]` 는 `active`/`review` 카드의 착지 증거와 닫을 명령을 새로 보태고, 같은 증거 함수(`landed::find`)를 한 번 더 부르기만 해요.
- Q4: 리더만 할 수 있어야 하나요? -> A: 예약을 끝내는 쪽(`withdraw` 와 `fold` 가 lease 끝난 held 카드를 닫는 경우)만 `--by leader` 를 요구해요. `card land` 가 같은 규칙을 써요(`src/card/land.rs:137`). 예약 없는 draft/ready 카드는 지금처럼 누구나 닫아요.

## Intent

카드 원장에 쌓인 "닫아야 하는데 닫히지 않은 카드"를 리더가 눈으로 찾고 손으로 닫고 있어요. `card_link` 는 `split_of`, `relates`, `blocks` 를 받지만 쓰는 곳은 `pr` 뿐이고(`src/ledger/workflow.rs:973`), `withdraw` 는 예약 없는 `draft`/`ready` 카드만 받고(`src/ledger/workflow.rs:830-838`), 만료된 held 예약에는 표시가 없고(`src/ledger/reservation.rs:452-457`), `draft` 는 나이를 먹어도 `next` 에 나오지 않아요. 이 SPEC 은 (1) 중복 카드를 합치는 `card fold`, (2) `card next` 의 `stale[]` 목록(닫을 명령이 붙은 네 종류), (3) lease 가 끝난 held 카드를 리더가 한 번에 withdraw 하는 길을 정해요. 새 카드 상태와 스키마 변경은 없어요.

읽는 사람: 카드를 정리하는 리더, `card next` 를 읽는 레인, 이 규칙을 문서에 옮기는 카드.

## Non-goals

- 새 카드 상태 값, 새 표, 새 링크 종류. 접힌 카드와 withdraw 한 카드는 모두 `hold` 예요.
- `stale[]` 가 카드를 스스로 닫는 일. `card next` 는 읽기 전용이고(`writes: false`) 닫을 명령만 말해요.
- 접힌 카드의 노트·기준·링크 이동이나 삭제. 노트는 그대로 읽혀요.
- `card reclaim`, `card release`, `card finish` 의 규칙 변경. `stale[]` 는 그것들을 가리킬 뿐이에요.
- 레인의 워크트리나 브랜치 삭제. 예약이 끝나도 워크트리는 리더의 `worktree-gc-plan` 몫이에요.

## Requirements

- REQ-1: When the leader or a lane runs `moai card fold C-n --into C-m` on a card nobody holds, the tool shall move C-n to `hold`, write one `card.fold` event with the target, the reason and the by, and link C-n to C-m with one active `relates` link.
- REQ-2 (REQ-1): IF C-n is `done`, is the target itself, is already `hold`, or is held by a live lease, or IF C-m does not exist or is `hold`, THEN the tool shall refuse the fold and write nothing.
- REQ-3 (REQ-1): The tool shall keep a folded card's notes, criteria and links readable, hide it from the default `card list`, show it under `card list --all`, never offer it in `card next`, and refuse `card take` naming the card it was folded into.
- REQ-4: When `moai card next` runs, the tool shall answer `stale`, without writing to the ledger, with one row per card that needs closing, each row naming the card, its `kind` and the `command` that closes it.
- REQ-5 (REQ-4): IF the ledger is absent or git cannot be read, THEN the tool shall answer `stale` without the rows it could not prove, say why in `landing.note`, and still exit 0.
- REQ-6 (REQ-4): The tool shall give a card the first fitting kind of `landed_open` (an `active` or `review` card whose landing is proven as `already_landed` proves it), `lease_ended` (a `held` reservation whose lease ended), `review_no_holder` (a `review` card nobody holds) and `old_draft` (a `draft` last updated at least the stale-draft days ago).
- REQ-7 (REQ-6): The tool shall read the stale-draft days from the profile key `stale_draft_days`, use 14 when the key is absent, and refuse a profile that sets it to 0.
- REQ-8: When the leader runs `moai card withdraw` or `moai card fold` with `--by leader` on an `active` or `review` card whose only reservation is `held` with an ended lease, the tool shall end that reservation as `released`, move the card to `hold`, and record the former holder, generation and lease end in the event.
- REQ-9 (REQ-8): IF the lease is live, the reservation is not `held`, or `--by` is not `leader` while a reservation exists, THEN the tool shall refuse and write nothing.
- REQ-10: The tool shall leave the behavior of `moai card withdraw` for an unheld `draft` or `ready` card, and of `card take`, `card reclaim` and `card finish`, as it was.

## Acceptance

- AC-1 (REQ-1, REQ-3): on a temporary ledger, folding a `ready` card into another leaves it in `hold` with one `card.fold` event (`into`, `reason`, `by`) and one active `relates` link, `card show` still prints its notes and criteria, `card list` omits it and `card list --all` shows it, `card next` names it neither in `ready` nor in `skipped`, and `card take` refuses naming `C-m` · verify: `cargo nextest run --locked --offline --test card_fold_cli`
- AC-2 (REQ-2): folds of a `done` card, of a card into itself, into a missing card, into a `hold` card, of a `hold` card and of a card with a live lease each exit non-zero and the ledger bytes are the same before and after; folding the same card twice is refused the second time · verify: `cargo nextest run --locked --offline --test card_fold_cli`
- AC-3 (REQ-4, REQ-6): on a temporary ledger and git project with one card of each kind, `card next` answers `stale` with four rows, each with `name`, `kind` and a `command` that names the card, no card twice, a plain `ready` card absent, and the ledger bytes the same before and after · verify: `cargo nextest run --locked --offline --test card_stale_cli`
- AC-4 (REQ-5, REQ-6): with git unreadable the `landed_open` rows are missing and `landing.note` says why while the other kinds still list and the exit is 0; an `active` card whose recorded head is inside the integration branch is `landed_open` with `via` and `head` and the command `moai card finish C-n --head <head> --by leader`; one whose head is outside the branch is not listed for it; a card that is both `held` with an ended lease and landed is listed once, as `landed_open` · verify: `cargo nextest run --locked --offline --test card_stale_cli`
- AC-5 (REQ-6, REQ-7): a `draft` last updated exactly the threshold days ago is `old_draft` and one updated one second younger is not; the profile key `stale_draft_days: 3` moves the threshold; an absent key means 14; `stale_draft_days: 0` is refused when the profile is read · verify: `cargo nextest run --locked --offline --test card_stale_cli`
- AC-6 (REQ-8): `card withdraw C-n --by leader` on an `active` card whose `held` lease ended sets the reservation to `released`, the card to `hold`, writes one `card.withdraw` event that names the former holder, generation and `lease_until`, and the card's worktree folder is untouched; `card fold` does the same with a `card.fold` event · verify: `cargo nextest run --locked --offline --test card_withdraw_cli`
- AC-7 (REQ-9, REQ-10): the same call with a live lease, with `--by` other than `leader`, and on a reservation in `held_for_review` or `expired_candidate` is refused and the ledger bytes are unchanged; the existing withdraw, take, reclaim and finish tests pass as they are · verify: `cargo nextest run --locked --offline --test card_withdraw_cli --test card_take_scope_cli --test card_reclaim_cli --test card_finish_cli`

## Scope

- write: `.moai/specs/SPEC-20261124-card-fold-stale/spec.md`, `src/ledger/close.rs` (new: the fold, the widened withdraw and the stale facts), `src/card/stale.rs` (new: the `stale` rows and their commands), `tests/card_fold_cli.rs` (new), `tests/card_stale_cli.rs` (new), `tests/card_withdraw_cli.rs`
- line-lease with the card that holds the file (C-442 holds `src/ledger/workflow.rs` and `src/ops/card.rs` on 2026-11-24; the edits are small and wait for the holder's lease or a return): `src/ledger/workflow.rs` (the body of `withdraw` calls `close.rs`), `src/ledger/mod.rs` (`pub mod close;`), `src/card/mod.rs` (`pub mod stale;`), `src/card/command.rs` (the fold command, the `withdraw` answer, the list filter that hides folded cards), `src/card/work.rs` (the `stale` key of the `card next` answer), `src/ledger/reservation.rs` (the take refusal names a fold target; one query), `src/ops/card.rs` (the `Fold` subcommand, its dispatch, the `Withdraw` help text), `src/profile.rs` (the `stale_draft_days` key, as `archive_grace_days`), `src/help.rs` (one line if the help index lists subcommands)
- read-only: `src/card/landed.rs` (C-442 holds it; the proof for `active` and `review` cards comes from its `find` with more candidate ids), `src/card/land.rs`, `src/ledger/schema.rs`, `tests/card_next_skip_reason_cli.rs`
- deferred: the row of `card fold` and the `stale` key in `docs/design-review/appendix/operations/card.md`, after its holder returns the file
- forbidden: `CLAUDE.md`, the ledger file, `src/ledger/schema.rs` (no schema change)

## Plan

1. RED: `tests/card_fold_cli.rs`, `tests/card_stale_cli.rs` and the new cases of `tests/card_withdraw_cli.rs` on a temporary HOME with a git project; each new test fails today (no `fold` subcommand, no `stale` key, withdraw refuses a held card).
2. `src/ledger/close.rs`: one `closable` check shared by `fold` and `withdraw` (unheld `draft`/`ready`; `active`/`review` with no reservation or a `held` reservation whose lease ended with `--by leader`), the `fold` write (state, event, `relates` link in one transaction), and the read of the stale facts (held-ended, review without holder, drafts with their age).
3. `src/card/stale.rs`: builds the rows from those facts and from the landing proofs of `landed::find` over every `active` and `review` card (called once, together with the `ready` candidates), orders them by card, and writes each `command`.
4. Wire up: `card fold` in `src/ops/card.rs` and `src/card/command.rs`, `stale` in `src/card/work.rs`, the list filter and the take refusal for `card.fold`, the profile key.
5. Guard mutants, one per new guard: take the lease check out of the widened withdraw (a live lease must stay refused); take `card.fold` out of the list filter; make the draft threshold `>` instead of `>=`; drop the `leader` word check. Each new test must fail with exit 100.
6. Dangerous path: a throwaway run of `card withdraw --by leader` on a card whose lease ended, with a live worktree folder, showing the folder is still there.
7. `moai check affected` once at READY.

## Risks

- `card withdraw` now ends a reservation. A lane that is still working with a lapsed lease loses its card; the guard is `--by leader`, the ended lease, and the event that records the former holder. The lane can not write after that (`card check` refuses the stale generation).
- `stale[]` asks git for a landing proof of every `active` and `review` card. The lookup is the one `already_landed` uses (`COMMIT_LIMIT` bounded, git trouble turns `checked` off); the rows of that kind are missing when git could not be read, and `landing.note` says so.
- Two cards may hold `src/ledger/workflow.rs` and `src/ops/card.rs` at the same time; the logic is in new files so the shared-file edits stay a few lines.
- A fold can hide a card that was not a duplicate. The `card.fold` event and the `relates` link keep the way back: `card ready` on a `hold` card moves it to `ready` again (the transition table has `hold` → `ready`).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A folded card is `hold` with a `card.fold` event and a `relates` link | the six states are pinned by the CHECK; the free event kind is how `withdraw` marks a card | `card ready C-n` and retire the link |
| D-2 | `stale[]` is a key of `card next`, not a new command | the card text asks for it; `next` is the leader's one read | remove the key |
| D-3 | Four kinds in the order `landed_open`, `lease_ended`, `review_no_holder`, `old_draft`; one row per card | a landed card with an ended lease is closed by `finish`, not by withdraw | change the order |
| D-4 | `stale_draft_days` defaults to 14 and 0 is refused | the card says "N days (profile value)" without a number; 14 is the default this SPEC proposes, needs the leader's note | change the default |
| D-5 | Ending a reservation by withdraw or fold needs `--by leader`, an ended lease and status `held` | mirrors `card land`; `held_for_review` and `expired_candidate` belong to `recover` | drop the word check |
| D-6 | `withdraw` and `fold` also take `active`/`review` cards that nobody holds | the review-without-holder kind needs a closing command | restrict to draft and ready |
| D-7 | New logic goes to `src/ledger/close.rs` and `src/card/stale.rs` | `src/ledger/workflow.rs` and `src/ops/card.rs` are held by C-442 | move it back |
