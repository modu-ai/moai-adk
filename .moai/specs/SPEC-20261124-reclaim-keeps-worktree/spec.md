---
id: SPEC-20261124-reclaim-keeps-worktree
title: "card reclaim lets a lane take over a card in the same worktree when it holds committed, unlanded commits, and still refuses uncommitted changes"
status: draft
revision: 2
created: 2026-11-24
---

# card reclaim lets a lane take over a card in the same worktree when it holds committed, unlanded commits, and still refuses uncommitted changes

## Intent

Recurrence prevention, third occurrence. `card reclaim` of a card whose lease ended refuses with "worktree dirty: … holds 0 uncommitted change(s) and N unpushed commit(s); a reclaim would orphan them" for three different cards (one with a single commit, one with five, one more whose holder session vanished after a `/clear`). The refusal comes from the takeover gate (`worktree_gate` in `src/card/dispatch.rs`), which counts a commit as lost work when it sits on no remote-tracking ref and not in the local integration branch.

That count is wrong for a takeover. `reservation::reclaim` hands the new holder the same worktree and the same branch, so a commit in that worktree is not orphaned when the new holder continues there. And lanes never push: in NeXT-Flow the READY note is a local commit, so "unpushed" is the normal state of every open card. The result is a card that can never be taken again and whose repair is blocked until somebody works around the guard.

The gate keeps one job: refuse what a takeover really loses, which is an uncommitted change. Commits are kept and the event says so.

Readers: lanes that continue a card after the holder's session ended, the leader that dispatches the repair, and the auditor of this change.

## Interview

- Q1: Does the early reclaim (a reclaim before the lease ended, on proof that the holder is gone) change too? -> A: No (D-1). A live lease makes a missed draft cost more, and the early proof also counts a stash entry, the reserved branch and ignored drafts (`early_reclaim.rs`); it keeps counting unpushed commits as it does today.
- Q2: Does the refusal for an uncommitted change get weaker? -> A: No (D-2). It stays a refusal of kind `conflict`, names the path and both counts, and writes nothing; a new test pins it with commits beside the change.
- Q3: What happens when git cannot count the commits? -> A: The takeover still refuses as unreadable (D-3); the count is also what the event reports, so an unreadable count is never read as "none".
- Q4: Does `card release` or a hand-back change? -> A: No. They end a reservation and never drop a worktree; this SPEC changes `card reclaim` only (Non-goals).

## Non-goals

- The early reclaim of a live lease: it keeps refusing unpushed commits (REQ-5).
- Removing, moving or resetting a worktree or a branch: the reclaim keeps both, and no code of this SPEC deletes anything.
- Counting stash entries or ignored drafts in the takeover gate: the takeover gate never counted them (`SPEC-20261120-reclaim-dead-holder` D-5) and does not start now.
- The three-reclaim limit of a card, the self-reclaim renewal and the generation rules.
- `card adopt`, which reads no worktree at all.

## Requirements

- REQ-1: When `card reclaim` takes over a card whose recorded worktree exists and holds no uncommitted change, the command shall reclaim it, with the same worktree and branch on the new reservation, even when that worktree holds commits that are on no remote-tracking ref and not in the local integration branch.
- REQ-2 (REQ-1): IF the recorded worktree holds an uncommitted change, THEN `card reclaim` shall refuse with the conflict kind, name the path and both counts, and write nothing, whether or not the worktree also holds unpushed commits.
- REQ-3: When `card reclaim` takes over a worktree that holds unpushed commits, the `card.reclaim` event and the command's answer shall name them, as their count and the ids of up to twenty of them, newest first.
- REQ-4 (REQ-3): IF git cannot count the unpushed commits of the recorded worktree, THEN `card reclaim` shall refuse as unreadable and write nothing.
- REQ-5: The early reclaim of a reservation whose lease has not ended shall keep refusing a worktree that holds unpushed commits, with the refusal text it has today.
- REQ-6: While the lease of the reservation has not ended, `card reclaim` shall judge the recorded worktree with the strict gate and nothing else, so that a worktree with unpushed commits is refused with `card conflict in worktree dirty: <path> holds N uncommitted change(s) and M unpushed commit(s); a reclaim would orphan them`, before the lease or any proof of a gone holder is judged.
- REQ-7: When `card reclaim` takes over a worktree that holds twenty or twenty-one unpushed commits, `kept_commits` shall list twenty ids and count every unpushed commit. · failure: none

## Acceptance

- AC-1 (REQ-1): in a temporary ledger, a card whose lease ended and whose worktree holds two committed, unpushed commits and no uncommitted change is reclaimed; the new reservation is `held` at the next generation with the same worktree and branch, and the worktree's head is unchanged · verify: `cargo nextest run --locked --offline --test card_reclaim_cli reclaim_keeps_a_worktree_that_holds_unpushed_commits`
- AC-2 (REQ-2): a worktree with one untracked file and no commit, and a worktree with one untracked file and two unpushed commits, are each refused with kind `conflict`, a message that holds the path and `holds 1 uncommitted change(s) and N unpushed commit(s)`, and every ledger row as it was · verify: `cargo nextest run --locked --offline --test card_reclaim_cli reclaim_still_refuses_an_uncommitted_change_with_or_without_commits`
- AC-3 (REQ-3): the `card.reclaim` event of the takeover in AC-1 holds `kept_commits` with `count` 2 and the two commit ids newest first, the command's answer holds the same object, and a worktree with twenty-five unpushed commits lists twenty ids and counts twenty-five · verify: `cargo nextest run --locked --offline --test card_reclaim_cli the_reclaim_event_names_the_kept_commits`
- AC-4 (REQ-4): a worktree whose git cannot count its commits (a repository with no commit yet) and a folder that is no checkout are each refused with `could not be read` and every ledger row as it was · verify: `cargo nextest run --locked --offline --test card_reclaim_cli reclaim_refuses_a_worktree_whose_unpushed_commits_cannot_be_counted reclaim_gates_refuse_when_git_cannot_be_read`
- AC-5 (REQ-5): a live lease with a proof of a gone holder and a worktree that holds one unpushed commit stays refused, and the refusal names `1 unpushed` · verify: `cargo nextest run --locked --offline --test card_reclaim_early_cli refused_without_the_worktree_proof one_commit_outside_the_integration_branch_still_stops_an_early_reclaim`
- AC-7 (REQ-6): a card whose lease is live and whose worktree holds one committed, unpushed commit and no uncommitted change is refused by `card reclaim` with kind `conflict` and a message that starts with `card conflict in worktree dirty` and holds `holds 0 uncommitted change(s) and 1 unpushed commit(s); a reclaim would orphan them`, with every ledger row as it was; the same card taken over after its lease ended is reclaimed · verify: `cargo nextest run --locked --offline --test card_reclaim_cli a_live_lease_keeps_the_strict_gate_and_its_refusal_text`
- AC-8 (REQ-7): worktrees with twenty and with twenty-one commits answer `count` 20 with twenty ids and `count` 21 with twenty ids, newest first · verify: `cargo nextest run --locked --offline --test card_reclaim_cli the_kept_commit_list_is_capped_at_twenty_ids`
- AC-6 (REQ-1): a commit inside the local integration branch is not counted as kept, and a commit outside it is, for a profile that names the integration branch and for one that does not · verify: `cargo nextest run --locked --offline --test card_reclaim_cli commits_inside_the_integration_branch_do_not_stop_a_takeover one_commit_outside_the_integration_branch_is_kept_by_a_takeover without_a_usable_integration_branch_the_takeover_keeps_every_unpushed_commit`

## Scope

- write: `.moai/specs/SPEC-20261124-reclaim-keeps-worktree/spec.md`, `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `tests/card_reclaim_cli.rs`, `docs/design-review/appendix/operations/card.md` (one clause: the `kept_commits` key of the `card.reclaim` event, leader note 7893 N3)
- read-only: `CLAUDE.md`, `src/ops/card.rs` (revision 1 wrote its help sentence; card C-420 holds the file on 2026-11-24, so revision 2 leaves it as it is), `src/card/early_reclaim.rs`, `tests/card_reclaim_early_cli.rs`, `tests/ledger_reservation.rs`, `tests/prop_reservation.rs`
- forbidden: `tools/repo_scans.json`

`src/card/early_reclaim.rs` stays as it is: the early proof calls the unchanged strict gate (D-1). The public `reservation::reclaim` and `reservation::reclaim_early` keep their signatures, so no other caller changes.

## Plan

1. RED: in `tests/card_reclaim_cli.rs`, add `reclaim_keeps_a_worktree_that_holds_unpushed_commits`, `reclaim_still_refuses_an_uncommitted_change_with_or_without_commits`, `the_reclaim_event_names_the_kept_commits` and `reclaim_refuses_a_worktree_whose_unpushed_commits_cannot_be_counted`, and turn the three takeover tests that expect a refusal for unpushed commits into tests that expect a takeover and a kept count (`one_commit_outside_the_integration_branch_is_kept_by_a_takeover`, `one_commit_outside_the_project_folders_integration_branch_is_kept_by_a_takeover`, `without_a_usable_integration_branch_the_takeover_keeps_every_unpushed_commit`) and the first test's unpushed half. All fail on the current code.
2. GREEN, gate: in `src/card/dispatch.rs`, split the takeover from the early proof. `worktree_gate` keeps its text and counts for the early proof. A new takeover gate refuses an uncommitted change only, still reads the unpushed count (unreadable refuses), and returns the kept commits (count and up to twenty ids).
3. GREEN, ledger: in `src/ledger/reservation.rs`, add one public function that takes the optional early proof and the optional kept-commits object, and writes `kept_commits` into the `card.reclaim` payload; `reclaim` and `reclaim_early` call it with none.
4. The answer of `card reclaim` carries `kept_commits` when the worktree exists; the help text in `src/ops/card.rs` says that unpushed commits do not stop a takeover and that uncommitted changes do.
5. Guard mutants: remove the uncommitted refusal, count uncommitted as zero, drop the kept object from the payload, truncate the list wrongly, make an unreadable count pass, and send the early proof through the lenient gate; each must fail a named test.
6. SYNC: Result with each AC met or not met; READY note with the `check affected` receipt.
7. Revision 2, RED: add `a_live_lease_keeps_the_strict_gate_and_its_refusal_text` (AC-7) and `the_kept_commit_list_is_capped_at_twenty_ids` (AC-8) to `tests/card_reclaim_cli.rs`; AC-7 fails on the revision 1 code.
8. Revision 2, GREEN: in `takeover_gates` choose the gate by the lease state (D-6). Guard mutant: the lenient gate again for every reservation, which must fail AC-7 (exit 100). Then the `kept_commits` clause in `docs/design-review/appendix/operations/card.md`, the Result rewrite, and the READY note with a new `check affected` receipt.

### Revision 2 (repair after the audit of head 31c0cdf0d5, note 7893)

The audit found that `reclaim` runs `takeover_gates` for every reservation, whatever the lease state, and revision 1 pointed `takeover_gates` at the lenient gate. A live lease with an unpushed commit therefore stopped refusing with `card conflict in worktree dirty: …` and refused with `reservation conflict in card: the lease runs until …` instead (and not at all about the commit when the launcher is alive). REQ-5 says the text stays. The repair: `takeover_gates` reads the reservation and the ledger clock it already has open, uses the lenient takeover gate only when the reservation is `held` and `lease_ended(lease_until, now)` is true, and otherwise the strict `worktree_gate` as before, returning no `kept_commits`. A clock that cannot be read, or a reservation that is not `held`, takes the strict path. The audit's other points (N1 uncounted drafts beside commits, N2 `kept_commits` reads `HEAD`, not the reserved branch) stay as they are; N3 is the one clause of `operations/card.md`; N5 is AC-8.

## Risks

- Relaxing the takeover gate must never reach the uncommitted-change refusal: REQ-2 and AC-2 pin it with a change beside commits, and the guard mutants include its removal.
- The early proof calls the same shared helpers; splitting the gate must leave `worktree_gate` byte-for-byte for it. AC-5 and the early test file pin the refusal text; the early test file is run in the READY check.
- The kept list is read from git (`rev-list`), an extra call per takeover; the list is capped at twenty ids, so the event stays small.
- This relaxes a check, it adds no check that blocks, so rule 11 of `sdd.md` (observe first) does not apply; the Result records which refusals disappear.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only the takeover of an ended lease stops counting unpushed commits as lost; the early reclaim keeps counting them | the card asks for the ended-lease case, a live lease costs more when a draft is missed, and the early proof reads more than the takeover gate | send the early proof through the lenient gate in a later SPEC |
| D-2 | An uncommitted change keeps its refusal and message prefix (`holds N uncommitted change(s) and M unpushed commit(s)`); only the tail changes to say the uncommitted changes would be orphaned | the card forbids weakening it, and the existing tests read the prefix | restore the old tail |
| D-3 | An unreadable commit count refuses, as before | the count feeds the event; an unreadable answer is never read as "none" | none needed |
| D-4 | The event key is `kept_commits` with `count` and `ids` (up to twenty, newest first); it is present only when the recorded worktree exists and was read | the card asks that the event names the kept commits; twenty keeps the event small | rename the key in a later SPEC |
| D-5 | The ledger gets one new public function; `reclaim` and `reclaim_early` keep their signatures | other callers and tests use them as they are | fold the argument into them later |
| D-6 | The gate is chosen by the lease state read from the ledger clock (`unixepoch()` on the read-only handle `takeover_gates` already holds, the same read `early_reclaim::assess` makes): lenient only for a `held` reservation whose lease ended, strict for everything else, also when the clock cannot be read | the audit (note 7893 B1): REQ-5 keeps the live-lease refusal text; strict is the older, safe default; a lease that ends between the gate and the write only means a refusal that was stricter than needed | route every takeover through the lenient gate again |

## Result

Status: revision 1 was approved by the leader (note 7826, a leader judgment and not the user's own pick) and built; the audit of head 31c0cdf0d5 failed it (note 7893). Revision 2 is a draft; this Result is rewritten when its repair is built, and the sentence "Every other refusal … early reclaim of a live lease … is as before" below was wrong for revision 1 and is replaced then.

As built:
- `src/card/dispatch.rs`: `worktree_gate` keeps its text and counts for the early proof; both gates share one read (`read_worktree`). The new `takeover_worktree_gate` refuses an uncommitted change only (the message keeps the prefix `holds N uncommitted change(s) and M unpushed commit(s)` and ends `a reclaim would orphan the uncommitted changes`), refuses as unreadable when git cannot count or list the commits, and returns `kept_commits` (`count`, up to twenty `ids`, newest first). `reclaim` passes it to the ledger and into the answer.
- `src/ledger/reservation.rs`: the public `reclaim_keeping` writes `kept_commits` into the `card.reclaim` payload; `reclaim` and `reclaim_early` keep their signatures.
- `src/ops/card.rs`: the `card reclaim` help says that unpushed commits do not stop a takeover and that an uncommitted change does.
- `tests/card_reclaim_cli.rs`: four new tests, and the takeover tests that expected a refusal for unpushed commits now expect a takeover and a kept count. The RED commit (`9d02e57c3b`) failed 7 of 25 tests before the gate was split.

Results (each from a command run on this branch, `cargo nextest run --locked --offline --test card_reclaim_cli --test card_reclaim_early_cli`, 58 tests, 58 passed, exit 0; log `target/c428/green-run.log`):
- AC-1 met: `reclaim_keeps_a_worktree_that_holds_unpushed_commits` passes.
- AC-2 met: `reclaim_still_refuses_an_uncommitted_change_with_or_without_commits` passes, and `reclaim_gates_refuse_a_worktree_with_uncommitted_work` stays green.
- AC-3 met: `the_reclaim_event_names_the_kept_commits` passes (two commits; twenty-five commits list twenty ids and count twenty-five).
- AC-4 met: `reclaim_refuses_a_worktree_whose_unpushed_commits_cannot_be_counted` and `reclaim_gates_refuse_when_git_cannot_be_read` pass.
- AC-5 met: `refused_without_the_worktree_proof` and `one_commit_outside_the_integration_branch_still_stops_an_early_reclaim` pass; the whole early file passes.
- AC-6 met: `commits_inside_the_integration_branch_do_not_stop_a_takeover`, `one_commit_outside_the_integration_branch_is_kept_by_a_takeover` and `without_a_usable_integration_branch_the_takeover_keeps_every_unpushed_commit` pass.

Guard mutants (each applied by an edit, confirmed with a non-empty diff, run on the named test file, then restored with `git checkout -- <file>`; no mutant log holds a compile error; logs in `target/c428/`):
- the uncommitted refusal replaced by `if false`: exit 100 (4 tests fail).
- the uncommitted refusal passes when commits stand beside the change (`&& unpushed == 0`): exit 100 (2 tests fail).
- the id cap 20 changed to 19: exit 100 (2 tests fail).
- `kept_commits` not written into the event payload: exit 100 (2 tests fail).
- the strict gate of the early proof ignores unpushed commits: exit 100 on `card_reclaim_early_cli` (6 tests fail).
- an unreadable commit count read as none: the count guard alone survived (exit 0), because the listing of the ids refuses the same worktree; both guards removed together fail `reclaim_refuses_a_worktree_whose_unpushed_commits_cannot_be_counted` (exit 100). The two guards cover each other on purpose.

Refusals that disappear: a takeover of an ended lease no longer refuses a worktree that holds only unpushed commits. Every other refusal (uncommitted change, unreadable path or git, relative path, early reclaim of a live lease) is as before.

Gaps: the `check affected` receipt is in the READY note, not here.
