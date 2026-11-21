---
id: SPEC-20261121-flow-skeleton
title: "Flow skeleton: close loop, landing, shared files, admission, signals and the measure command, each as code"
status: draft
revision: 1
created: 2026-11-21
---

# Flow skeleton: close loop, landing, shared files, admission, signals and the measure command, each as code

## Intent

Parallel work was not the mistake; the way it ran was. Only coding is parallel. Verification (two heavy slots on one machine), integration (one branch) and closing (by hand) are one line, and they were left to prose rules and a leader turn. The leader report of 2026-11-21 (ledger read at 19:38Z) measured it: 318 cards created and 82 closed in six days; landings went on (22 to 37 first-parent commits a day) while closes fell from 57 a day to 4 and 6 after the move from pull requests to local landing; 40 cards held by 19 sessions with 23 leases past their time; 146 ready cards of which 136 have no write scope; 5801 notes with a median of 13 per closed card; 148 of 293 SPEC branches not in the integration branch.

This is the master SPEC of the redesign (user order of 2026-11-21: fix the problem at the root, report first, then proceed). It writes no code. For each step it gives the requirements, the acceptance criteria, what the code at the integration tip already does (file and line), the existing card that carries it, and what is new; and it defines the measure command for the seven numbers of the report. Each step with new code gets its own SPEC and its own landing, and one writer (the holder of the close-loop card) writes the kernel steps in order.

Principles (the report): what has one right answer given the state is code, not a note or a leader turn; lanes pull inside a limit; a card proves its independence before it can be ready; finish before start; coupled areas (ledger, card, install) have one writer and an independent audit; one migration at a time, finished when the old path is gone and a scan test refuses its return.

Every file and line below was read at the integration tip `021269e673994d08a90f7448ddbedde15054c353`; a slice re-reads the lines it depends on when it starts, because the tip moves (D-11).

## Non-goals

- Code of any step: this SPEC is the plan and the measure definition only; each step with code has its own SPEC.
- A new top-level command word: every command of the kernel is an action of `moai card` (D-1).
- Train apply (re-merge and reuse of evidence), because it needs check receipts by tree first (step 5).
- Steps 5 to 7 in detail: verification cost, modes and rules into code are listed with their carrying cards and get their own SPEC when they start.
- A bulk edit of the real ledger: every test runs on a temporary home.

## Requirements

- REQ-1: When the leader names a candidate commit for a card, the card land command shall move the integration branch to the candidate by one compare-and-swap reference update and then close the card with the proof the finish command makes · failure: none — a card that cannot be landed, a branch checked out in a worktree, a candidate that is not a commit, a SPEC that fails at the candidate, or a base that is not on the first-parent line of the tip refuses and leaves every table and reference unchanged.
- REQ-2: When the card land command runs again after an interruption between the reference move and the close, the card land command shall finish the missing half and answer as one uninterrupted run would · failure: none — a candidate already inside the tip skips the move, and a card already done answers done and writes nothing.
- REQ-3: When the candidate is the merge of a lane head onto the tip, the card land command shall print the land evidence of the candidate (its parents, whether the merge tree is clean, whether the manifest hashes match, the audit tier floor) · failure: none — evidence that cannot be computed is printed as unknown and the landing is not blocked until the typed verdict exists (REQ-9).
- REQ-4: When a path is generated at landing or is a registration line file, the scope check shall keep it out of the reasons to refuse another card's reservation · failure: none — a registration file named in a main write scope is refused with a message that names `card line-lease`.
- REQ-5: When a card is made ready, the ready command shall refuse it with every missing admission proof named at once (a write scope, a criterion, no scope that only a line lease can take) · failure: none — a card already ready is left as it is.
- REQ-6: When a card waits for another card, the waits-for rule shall make the card next command skip it with the reason waits_for and make the card take command refuse it until the other card is done or withdrawn · failure: none — a link that would close a cycle is refused at link time.
- REQ-7: When a holder takes a card, the card take command shall refuse a second active card of that holder and, when a project limit is set, a card beyond that limit · failure: none — the project limit stays unset until it is computed from the measure command.
- REQ-8: When a hook receipt arrives and less than half of the holder's lease is left, the receipt handler shall renew that lease under the lane name that holds it · failure: none — a renewal that cannot be written never blocks the tool call.
- REQ-9: When a lane submits a candidate, blocks, answers, or an auditor gives a verdict, the card command shall write one typed event row with a JSON payload instead of a note · failure: none — a payload that does not parse or a verdict without a head refuses and writes nothing.
- REQ-10: When the leader or the user asks what needs them, the card inbox command shall list exactly the unanswered blocks, the fail verdicts with no later submit, the held reservations past their lease, the landed cards that are not done, and the decision blocks owned by the user · failure: none — an empty inbox answers an empty list.
- REQ-11: When the card measure command runs, the card measure command shall print the seven numbers of the report from the ledger and git without writing anything · failure: none — a number that the ledger does not record (the leader's screen checks, before typed signals) is printed as null with the reason.
- REQ-12: When a migration is declared finished, a scan test shall refuse the return of its old path · failure: none — a migration without such a test is not finished.

## Acceptance

- AC-1 (REQ-1): a card in review with a candidate that is a merge of the tip lands in one call: the branch points at the candidate, the card is done, the reservation is released, and a refused call (branch checked out, candidate not a commit, SPEC failing at the candidate, stale base) leaves every table and reference unchanged · verify: `cargo test --locked --offline --test card_land_cli ac1_`
- AC-2 (REQ-2): a stop injected after the reference move leaves the branch moved and the card open, and the same call run again closes the card with one finish event and no second move; a second call on a done card answers done with nothing written · verify: `cargo test --locked --offline --test card_land_cli ac2_`
- AC-3 (REQ-3): the land answer carries the parents, the clean-merge flag, the manifest match and the tier floor, and evidence that cannot be computed reads unknown · verify: `cargo test --locked --offline --test card_land_cli ac3_`
- AC-4 (REQ-4): the eight help goldens are generated shared paths: two cards that both list a golden are not in overlap · verify: `cargo test --locked --offline --test card_generated_scope`
- AC-5 (REQ-4): a main write scope that names a registration file is refused with a message naming `card line-lease`, and a line lease on it is taken as before · verify: `cargo test --locked --offline --test card_scope_registration`
- AC-6 (REQ-6): a waits-for link between two cards is stored, shown, and a link that closes a cycle is refused with every table unchanged · verify: `cargo test --locked --offline --test card_link_cli`
- AC-7 (REQ-6): card next skips a waiting card with the reason waits_for, card take refuses it, and both offer it once the other card is done or withdrawn · verify: `cargo test --locked --offline --test card_dispatch_cli`
- AC-8 (REQ-5): a card without a write scope, without a criterion, or with only a line-lease path is refused at ready with all missing proofs in one answer, and cards that are already ready are untouched · verify: `cargo test --locked --offline --test card_ready_admission_cli`
- AC-9 (REQ-7): a second take by a holder with an active card refuses and names the card, and a card in review does not count · verify: `cargo test --locked --offline --test card_limit_cli ac9_`
- AC-10 (REQ-7): a project limit stored as a meta row refuses the take beyond it, and no row means no limit · verify: `cargo test --locked --offline --test card_limit_cli ac10_`
- AC-11 (REQ-8): with the lane variable exported the holder is the lane name, reservations written under a session id stay writable, and a holder that names neither is refused as before · verify: `cargo test --locked --offline --test card_lane_holder_cli`
- AC-12 (REQ-8): a receipt with under half of the lease left renews it, one with more leaves it, and a renewal failure leaves the tool call unblocked · verify: `cargo test --locked --offline --test hook_auto_renew_cli`
- AC-13 (REQ-9): a typed submit writes one event row with the head, land with no candidate reads the newest submit, and a payload that does not parse refuses · verify: `cargo test --locked --offline --test card_submit_cli`
- AC-14 (REQ-9): a typed block, answer and verdict each write one event row, and land refuses a card whose tier floor is 1 or more without a pass that names the head · verify: `cargo test --locked --offline --test card_signal_cli`
- AC-15 (REQ-10): the inbox lists the five conditions and nothing else, from typed rows, on a fixture ledger with every case and its opposite · verify: `cargo test --locked --offline --test card_inbox_cli`
- AC-16 (REQ-10): card wait returns on the first signal event after a cursor, holds no read transaction while it waits, and returns by its timeout when none comes · verify: `cargo test --locked --offline --test card_wait_cli`
- AC-17 (REQ-10, REQ-12): the leader round reads typed rows and joins them by lane name, a stranded lane is one whose lease ended and whose name is not live, and a scan test refuses a new first-line note parser · verify: `cargo test --locked --offline --test leader_round_cli`
- AC-18 (REQ-11): on a fixture ledger the measure command prints the seven numbers with their values, the null reasons, and the window, and a second run prints the same · verify: `cargo test --locked --offline --test card_measure_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-flow-skeleton/spec.md`

## Decisions

- D-1: The land command is `card land`; no top-level word is added. Every kernel command is an action of `moai card`.
- D-2: Until the typed verdict exists (step 4c), `card land` does not check audit evidence; it prints the tier floor and the leader checks the verdict note by hand.
- D-3: Only the leader lands until the typed verdict exists; after it, the holder lands with a pass that names the head.
- D-4: The limit counts active cards only; the project in-flight limit stays unset until it is computed from the measure command (REQ-7).
- D-5: The lane variable is the holder identity; reservations written under a session id stay writable.
- D-6: One kernel writer: the holder of the close-loop card writes the kernel steps in order, after its audit follow-up; no second writer in `src/card` and `src/ledger`. Independent audit of every kernel step stays.
- D-7: Typed submit comes before the read-only train (step 4a before 1c).
- D-8: A card that is superseded is a withdrawn card plus a `relates` link; no new link kind is added (a new kind needs a table rebuild, see Plan step 3).
- D-9: Acknowledge and resume lines stay notes; they are not signals.
- D-10: The measure command is `card measure`. It prints a JSON answer with the window, and for each of the seven numbers its value, its source and, when the ledger does not record it, null with the reason. The numbers are the ones of the report: M1 closed against created cards per UTC day (events `card.done`, `card.add` and `import`); M2 cards in flight (held reservations, distinct holders, leases past their time); M3 time from the first `reserve` event to `closed_at` (median, 90th percentile); M4 time from the `card.landed` event to close, and the count of landed cards that are not done; M5 notes per closed card (median, 90th percentile); M6 the leader's checks of lane screens (not recorded by the ledger today, so null until step 4f writes typed round rows); M7 reservations of shared files (rows of `reservation_scope` whose pattern is a generated shared path or a registration file). It also prints the supporting counts of the report: ready cards with no write scope and SPEC branches that are not in the integration branch (a read-only `git merge-base --is-ancestor` pass).
- D-11: Every citation in this SPEC names the tip it was read at; a slice re-reads its lines when it starts, and a line that moved is corrected in that slice's own SPEC, not by editing this one.

## Plan

Order, each step its own SPEC and landing; step 1 first, the migrations ride the train one at a time after step 1.

1. Close loop (carried by the card-finish card, landed on the tip as `f566ebf268`, and its follow-ups).
   - Exists: the landing proofs `integration_branch` (`src/card/landed.rs:121`), `tip_of` read as `refs/heads/<branch>` (`:133`), `show_at` for the SPEC at the tip (`:154`) and `contained` (`:174`); the command `finish` (`src/card/finish.rs:264`); the lease gate `finish_gate` (`src/ledger/workflow.rs:1266`) and the one write transaction `finish` (`:1310`). `card land` is one compare-and-swap reference move in front of finish.
   - New: 1a `card land` (candidate commit, optional met options, generation, by): checks with nothing written (card state, the branch not checked out in any worktree, the candidate is a commit, the SPEC passes at the candidate, the finish gate), classify against the tip (inside: skip the move; the tip on the first-parent line of the candidate: move; else refuse `stale_base`), `update-ref` compare-and-swap through the git environment scrub, then finish with the candidate as head; a stop between the move and the close is completed by the rerun, no intent row, git is the record. 1b the land evidence (parents are the tip and the head, the merge tree equals the candidate tree and is clean, the manifest matches, the tier floor is printed); it reuses `parse_head_args` (`src/bundle.rs:547`), `Git::conflict` (`:323`), `manifest_matches` (`src/gate_collect.rs:651`) and `tier_of` (`src/gate.rs:325`) and does not copy the pull-request-era forced text merge or the card name in the merge message. 1c `card train`, read only (ff_ready, already_landed, stale_base, conflict), after 4a.
2. Shared files leave the reservations (new SPEC).
   - Exists: `GENERATED_SHARED_PATHS` holds only the manifest (`src/scope.rs:37`) and `card_overlap` ignores a pair with one such side (`:55`; `tests/card_generated_scope.rs`), so the 90 manifest reservations of the report are scope rows, not blocks; the help goldens are rewritten by one variable (`tests/help_layout.rs:63`, eight cases from `:39`); `card line-lease` exists for one-line edits (`src/ops/card.rs:636`, handler `:1177`).
   - New: 2a the eight help goldens join the generated shared paths; 2b registration files (`src/lib.rs`, `src/card/mod.rs`) are refused in a main scope and the refusal names `card line-lease`.
3. Admission and the work-in-progress limit (new SPEC).
   - Exists: `card next` skips held, scope overlap, empty scope and review-with-no-holder (`SkipReason`, `src/ledger/workflow.rs:210`; `next`, `:263`) and, in the card layer, already landed (`src/card/landed.rs:385`); `card_link` allows `blocks`, `split_of`, `relates` and `pr` (`src/ledger/schema.rs:80`, unique indexes `:98`); `ready` (`src/card/dispatch.rs:86`) and `take` (`:367`); `renew` works on an ended lease as long as nobody reclaimed it (`src/ledger/reservation.rs:928`) but nothing calls it; every lane runs `hook-receipt` after each Bash call (`src/hooks.rs:82`, handler `:85`).
   - New: 3a waits-for as a `blocks` link (next skips with `waits_for`, take refuses, a cycle is refused); 3b admission at ready (all missing proofs named at once; existing ready cards untouched); 3c one active card per holder; 3d the project in-flight limit as a meta row; 3e the lane name is the holder (the launcher exports the lane variable and the holder resolver prefers it); 3f automatic renewal from the hook receipt (under half of the lease left, never blocks).
   - Constraints: a new link kind needs a table rebuild and the version gate refuses any `user_version` but the current one (`check_version`, `src/ledger/migrate.rs:38`), which is why D-8 adds none; profile structs deny unknown fields (`src/profile.rs:91`, `:176`), so a list that belongs in the profile goes in code first.
4. Typed signals, the inbox and the event wake (the leader-round card is its first slice).
   - Exists: events are rows with a free-text kind and a JSON payload (`src/ledger/schema.rs:133`, kind `:139`, payload `:140`); a note kind is pinned by a CHECK (`:68`), so signals are event rows and not a new note kind; `card inbox` reads note prefixes (`src/card/inbox.rs:50`); `card timeline` reads after a cursor (`src/ops/card.rs:874`); first-line note parsers exist five times (`src/bundle.rs:23`, `src/card/inbox.rs:50`, `src/leader_round.rs:32` (five marker words), `src/statusline/state.rs:1148` and the markers of the leader-round branch); `session_live` takes a lane name, not a session id (`src/lane_registry.rs:139`).
   - New: 4a typed submit (land without a candidate reads the newest); 4b typed block and answer; 4c typed verdict (land refuses a tier floor of 1 or more without a pass that names the head); 4d the inbox from typed rows, exactly the five conditions of REQ-10; 4e `card wait` (blocks until a signal event after a cursor, holds no read transaction; the latency of the data-version read across processes is measured first); 4f the leader round on typed rows, joining by lane name; the leader-round branch gives a session id to `session_live`, so every ended lease reads as stranded, and 4f does not copy that; a scan test refuses new marker parsers.
5. Verification cost (existing cards C-280, C-282, C-209): receipts by tree and audit by risk tier. Today `check affected` prints a summary and writes no receipt file (`src/check_affected.rs:532`, `:683`), which is why train apply waits for this step. No new SPEC here.
6. Modes (new SPEC, later): solo is the default; team size is computed from the independent ready cards and the verification capacity; a one-lane against two-lane comparison runs under the same checks.
7. Rules into code (existing card C-309 links): skill text and memory rules become checks or are removed.
8. Migrations on the train, one at a time after step 1 (cards C-199 path root, C-160 Go links, C-313 and C-291 Codex and OMP removal, C-318 old name, C-317 Python): each is finished when its old path is gone and a scan test refuses its return (REQ-12).

The direction each step has to move (the SPEC of a step records why when it does not):

| Step | Number it moves | Direction |
| --- | --- | --- |
| 1 close loop | M4 landing to close, M1 closed against created | M4 to the same call, M1 closed at least created |
| 2 shared files | M7 reservations of shared files | down to 0 |
| 3 admission and limit | M2 cards in flight, ready cards with no scope | one active card per lane, 0 ready cards with no scope |
| 4 signals | M5 notes per closed card, M6 leader screen checks | M5 to 5 or less, M6 only on an exception |
| 5 to 7 | M3 take to close | falling step by step |

The target values are proposals of the report; the first run of `card measure` sets them.

## Risks

- The kernel has one writer on purpose; a long step delays the others. The mitigation is that each step is small, lands alone and is audited independently.
- Every kernel command is a `card` action, so each adds clap lines to `src/ops/card.rs`, which other cards hold; the lines are small and go in by pair approval, one step at a time.
- A line number here can be wrong after the tip moves (D-11); a slice re-reads before it relies on one.
- M6 cannot be measured from the ledger before step 4f, so the first reports print it as null; this is stated in the answer, not guessed.
- The five first-line note parsers disagree today; until step 4 replaces them, a new marker word has to be added to all five or it is read by some of them only.

## Result

Not started: this SPEC is the plan of the redesign and writes no code. The Result is written when the SPEC is accepted and its steps have their own SPEC.
