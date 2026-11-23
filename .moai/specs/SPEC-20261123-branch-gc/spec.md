---
id: SPEC-20261123-branch-gc
title: "A read-only plan lists the local branches that are merged and unused, and an apply deletes only the ones the user confirmed, with git branch -d"
status: draft
revision: 1
created: 2026-11-23
---

# A read-only plan lists the local branches that are merged and unused, and an apply deletes only the ones the user confirmed, with git branch -d

## Intent

통합 브랜치에 이미 들어간 로컬 브랜치가 계속 쌓여요. 2026-11-23 에 이 저장소를 읽은 값은 로컬 브랜치 1842개, 그중 1313개가 `v2026.11.0` 에 병합됨이에요 (`git for-each-ref refs/heads` 와 `--merged`). 착지한 브랜치를 지우는 기능은 어디에도 없어요. `worktree-gc-plan` 은 워크트리만 다루고 브랜치는 그대로 두어요.

이 SPEC 은 두 명령을 더해요.

- `branch-gc-plan`: 읽기 전용. 병합됨, 열린 카드 예약 없음, 어떤 워크트리에도 체크아웃되지 않음, 보존 ref 아님을 모두 만족하는 로컬 브랜치만 `delete` 로 올리고, 나머지는 남기는 이유와 함께 `keep` 으로 보여요. 목록과 digest(내용 해시)를 내요.
- `branch-gc-apply`: 사용자가 고른 뒤에만 쓰는 쓰기 명령. plan 을 같은 입력으로 다시 계산해서 digest 가 같을 때만 `git branch -d`(소문자)로 지워요. 원격 브랜치는 건드리지 않아요.

브랜치 삭제는 HOLD(deletion) 라서, 이 SPEC 이 코드가 되어도 실제 저장소에서 apply 를 돌리는 일은 사용자 선택 뒤 리더의 단계예요. 이 슬라이스의 검증은 임시 저장소에서만 해요.

### 이미 있는 것

- `worktree-gc-plan`/`worktree-gc-apply` (`src/worktree_gc.rs`): 같은 모양의 plan → digest → apply. 이 SPEC 이 그 모양을 따라요.
- `src/ledger/reservation.rs`: 예약에 `branch` 열이 있고, 활성 상태는 `held`, `expired_candidate`, `held_for_review` 예요.
- 브랜치 삭제 때 `reference-transaction` 훅이 워크트리 정리 큐에 한 줄을 더해요 (SPEC-20261123-worktree-slots). apply 는 그 훅을 그대로 지나가요.

## Non-goals

- 원격 브랜치와 원격 추적 ref(`refs/remotes/**`)를 지우거나 읽는 일. fetch 도 하지 않아요.
- 스쿼시 병합이나 리베이스 병합으로 들어간 브랜치. 팁의 조상이 아니면 `not_merged` 로 남겨요.
- 워크트리 삭제. 브랜치를 지워도 워크트리는 그대로고, 워크트리 정리는 `worktree-gc-plan` 의 일이에요.
- 대문자 `-D` 나 `--force` 로 지우는 일, 보존 ref 목록을 옵션으로 바꾸는 일.
- 정리 시점을 정하는 일(주기, 훅). 이 SPEC 은 사용자가 부르는 두 명령만 더해요.

## Requirements

<!-- This SPEC adds no check that blocks a write that is allowed today. -->
- REQ-1: When `branch-gc-plan` runs with a local integration ref, the `branch-gc-plan` command shall list every local branch with the action `delete` only when the branch tip is an ancestor of the integration tip, no worktree has the branch checked out, no active card reservation names the branch and the branch is not a preserved ref · failure: none — a branch that does not meet all four is listed with the action `keep`.
- REQ-2: The `branch-gc-plan` command shall treat `main`, `master`, `v2026.11.0`, `2026.11.0-rc.1`, every branch under `archive/` and the local branch the integration ref names as preserved refs.
- REQ-3: The `branch-gc-plan` command shall list every `keep` branch with the reasons `preserved`, `checked_out`, `card_reserved`, `not_merged`, `symbolic_ref` or `reservation_check_unavailable` that apply to it.
- REQ-4 (REQ-1): IF the project has a profile or a ledger that cannot be read, THEN the `branch-gc-plan` command shall keep every branch that would otherwise be deleted, with the reason `reservation_check_unavailable`.
- REQ-5: The `branch-gc-plan` command shall write no ref, no file and no ledger row and make no network call.
- REQ-6: The `branch-gc-plan` command shall answer a `plan_digest` that depends only on the integration tip and the name and tip of every `delete` branch.
- REQ-7: When `branch-gc-apply` runs with the digest of a plan and a permission reference, the `branch-gc-apply` command shall compute the plan again with the same input and run `git branch -d` for each `delete` branch only when the digests are the same and the branch tip is still the planned tip and no worktree has the branch checked out.
- REQ-8 (REQ-7): IF the permission reference is missing or empty or the recomputed digest differs from the given one, THEN the `branch-gc-apply` command shall refuse before any write and leave every branch as it was.
- REQ-9: The `branch-gc-apply` command shall delete a branch only with `git branch -d`, never with `-D`, `--force` or a remote command.
- REQ-10 (REQ-9): IF git refuses to delete a branch, THEN the `branch-gc-apply` command shall report the branch as `refused` with the git message and go on with the next branch.
- REQ-11: When `branch-gc-apply` has run `git branch -d` for a branch, the `branch-gc-apply` command shall report the branch as `deleted` only when reading the ref back shows it gone · failure: none — a ref that is still there is reported `failed`.
- REQ-12: The `moai help` listing shall show `branch-gc-plan` and `branch-gc-apply` with a one-line summary each.

## Acceptance

- AC-1 (REQ-1, REQ-3): in a temporary repository the plan lists a merged unused branch as `delete`, and lists as `keep` with its reason a branch with a commit the integration ref lacks (`not_merged`), a branch checked out in the primary checkout and in a linked worktree (`checked_out`), a branch an active card reservation names, in `held` and in review (`card_reserved`), and a symbolic ref in `refs/heads` (`symbolic_ref`); a branch whose only reservation is released is `delete` · verify: `cargo test --locked --offline --test branch_gc_cli ac1_`
- AC-2 (REQ-2): `main`, `master`, `v2026.11.0`, `2026.11.0-rc.1`, `archive/old` and the branch the integration ref names are each merged and unused in the fixture and each is `keep` with the reason `preserved` · verify: `cargo test --locked --offline --test branch_gc_cli ac2_`
- AC-3 (REQ-4): a profile that is not readable and a ledger file that is not a ledger each turn a merged unused branch into `keep` with `reservation_check_unavailable`, and a project without a profile reads as no reservations · verify: `cargo test --locked --offline --test branch_gc_cli ac3_`
- AC-4 (REQ-5, REQ-6): after the plan the refs, the worktree list, the ledger file and the project folder are the same by digest, the answer says `writes: false` and `network: disabled`, and the digest is the same on a second run and changes when a `delete` branch tip moves, a branch becomes `keep` or the integration tip moves · verify: `cargo test --locked --offline --test branch_gc_cli ac4_`
- AC-5 (REQ-7, REQ-11): apply with the plan's digest and a permission reference deletes exactly the `delete` branches, every `keep` branch and the commits stay, and the answer lists each branch as `deleted` · verify: `cargo test --locked --offline --test branch_gc_cli ac5_`
- AC-6 (REQ-8): a wrong digest, a digest of a plan made before a branch was added, a missing permission reference and an empty one each refuse and every branch is still there · verify: `cargo test --locked --offline --test branch_gc_cli ac6_`
- AC-7 (REQ-7): a branch whose tip moved after the plan, a branch a card reserved after the plan and a branch checked out after the plan each change the digest or are skipped as `refused`, and the branch stays · verify: `cargo test --locked --offline --test branch_gc_cli ac7_`
- AC-8 (REQ-9, REQ-10): a branch merged into the integration ref but not into the checked out HEAD is refused by `git branch -d`, is reported `refused` with the git message and stays, while the other branches of the same run are deleted · verify: `cargo test --locked --offline --test branch_gc_cli ac8_`
- AC-9 (REQ-9): with a remote that holds the same branches, apply leaves the remote refs, the remote-tracking refs and the remote's own branch list as they were · verify: `cargo test --locked --offline --test branch_gc_cli ac9_`
- AC-10 (REQ-12): both commands show in `moai help` with a one-line summary and their operation names are `branch.gc.plan` and `branch.gc.apply` · verify: `cargo test --locked --offline --test branch_gc_cli ac10_`

## Scope

- write: `.moai/specs/SPEC-20261123-branch-gc/**`, `src/branch_gc.rs` (new), `src/ops/branch_gc_plan.rs` (new), `src/ops/branch_gc_apply.rs` (new), `src/lib.rs` (one module line), `src/help.rs` (the two command rows), `tests/branch_gc_cli.rs` (new), `docs/design-review/appendix/operations/branch_gc_plan.md` and `branch_gc_apply.md` (new).
- read-only: `src/worktree_gc.rs` (`git_out`, `parse_worktree_list` use), `src/ledger/**`, `src/project_id.rs`, `src/contract.rs`.
- forbidden: any remote command, `src/worktree_gc.rs`, the ledger schema, the real branches and worktrees of this clone (tests use temporary repositories only).

## Plan

1. (this slice) The SPEC, status draft. The user's approval and the open decision U-1 are answered by the user, never written by a lane.
2. Tests first (RED): `tests/branch_gc_cli.rs` AC-1 to AC-10 on a temporary home and repository; they fail while the commands do not exist.
3. `src/branch_gc.rs`: `plan` (one `git for-each-ref --merged` call, one `git worktree list --porcelain`, one ledger read) and `apply` (digest check, per-branch re-check, `git branch -d --`, read back).
4. `src/ops/branch_gc_plan.rs` and `src/ops/branch_gc_apply.rs`, the module line in `src/lib.rs`, the help rows, the appendix fragments.
5. Mutants: remove each guard in turn (merged check, worktree check, reservation check, preserved list, digest check, tip re-check, `-d` to `-D`) and show the test that fails.
6. Dangerous-path run in a throwaway: the real binary on a temporary repository with merged, unmerged, checked-out, reserved, preserved and symbolic branches, a remote, then plan → apply → read back.
7. Targeted tests, then one `check affected`, then READY with head, base and the receipt.

## Risks

- R-1: `git branch -d` checks "merged" against the branch's upstream, or against HEAD without one, not against the integration ref. Run from a checkout whose HEAD is behind the integration tip, git refuses some `delete` rows. That is safe (they stay, reported `refused`) and AC-8 pins it; the leader runs apply in the checkout whose HEAD is the integration tip.
- R-2: A branch can move between the re-check and the delete. `git branch -d` has no expected-old-value, so the window is the time between two git calls; the `-d` merged check still stands, and the deleted tip is in the answer so `git branch <name> <tip>` restores it.
- R-3: A branch created at the integration tip and not yet checked out or reserved looks merged and unused. Deleting it loses only the name (the commit is the tip itself). The plan lists it like any other row so the user sees it before choosing.
- R-4: Deleting a branch queues a worktree cleanup through the `reference-transaction` hook of SPEC-20261123-worktree-slots. That cleanup has its own checks; apply does not run it.
- R-5: The preserved list names this repository's branches (`v2026.11.0`, `2026.11.0-rc.1`). Another project keeps them too, which costs nothing, and its own integration branch is preserved through the integration ref.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The merged test is "the branch tip is an ancestor of the integration tip" (`git for-each-ref --merged`), one call for all branches. | The card says merged into `v2026.11.0`; one call over 1842 branches avoids 1842 processes; `worktree-gc-plan` uses the same ancestor rule. | Add a squash-merge test later as its own rule. |
| D-2 | A reservation is active when its status is `held`, `expired_candidate` or `held_for_review`; a released or reclaimed reservation does not keep a branch. | The same set `src/ledger/reservation.rs` uses for `active_worktrees`; the card says "open card reservation". | Count every reservation ever made on the branch. |
| D-3 | The plan has no recent-change window for branches. | A branch carries no unsaved work; deleting a merged branch loses only its name (R-3). | Add a window like `--recent-minutes` of the worktree plan. |
| D-4 | Apply needs a permission reference, as `worktree-gc-apply` does. | Deleting is a HOLD category; the reference records the user's pick. | Drop the argument. |

### Open decisions of the user

- U-1: whether the real apply runs on this clone, and which plan digest it is given. HOLD (deletion). Recommended: the leader runs `branch-gc-plan` on the real clone, shows the `delete` count and a sample to the user as a selection, and runs apply only after the pick.

## Result

Status: implemented on the code slice; the SPEC front matter is not changed here (approval is recorded by the user's pick, not by a lane). `branch-gc-plan` and `branch-gc-apply` are in `src/branch_gc.rs` (the logic) and `src/ops/branch_gc_plan.rs`, `src/ops/branch_gc_apply.rs`. U-1 stays open: nothing here ran `branch-gc-apply` on this clone's branches.

- AC-1: met — `branch_gc_cli` tests `ac1_only_a_merged_unused_branch_is_a_delete_row`, `ac1_the_branch_checked_out_in_the_primary_checkout_is_kept`, `ac1_a_reservation_that_was_released_does_not_keep_its_branch`
- AC-2: met — `branch_gc_cli` tests `ac2_the_preserved_refs_are_kept_though_merged_and_unused`, `ac2_the_local_branch_the_integration_ref_names_is_preserved`; `ac2_an_integration_ref_that_looks_like_an_option_is_refused` covers the input check added beside it
- AC-3: met — `branch_gc_cli` tests `ac3_an_unreadable_profile_or_ledger_keeps_every_would_be_delete`, `ac3_a_project_without_a_profile_has_no_reservations`
- AC-4: met — `branch_gc_cli` tests `ac4_the_plan_writes_nothing_and_names_itself_read_only`, `ac4_the_digest_follows_the_integration_tip_and_the_delete_rows_only`
- AC-5: met — `branch_gc_cli` test `ac5_apply_deletes_exactly_the_delete_rows` (the symbolic ref `alias` stays; git stops listing it once its target is gone)
- AC-6: met — `branch_gc_cli` test `ac6_a_wrong_or_stale_digest_or_no_permission_refuses_and_deletes_nothing`
- AC-7: met — `branch_gc_cli` tests `ac7_a_branch_that_moved_or_was_checked_out_after_the_digest_check_is_refused` (the moved tip is a merged one, so only the re-check keeps the branch; the checkout refusal carries this command's own words) and `ac7_a_card_or_a_checkout_after_the_plan_changes_the_digest_and_refuses_the_run`
- AC-8: met — `branch_gc_cli` test `ac8_a_branch_git_refuses_to_delete_stays_and_the_rest_go_on` (HEAD on `main`; `-D` in place of `-d` fails it)
- AC-9: met — `branch_gc_cli` test `ac9_apply_leaves_the_remote_and_the_remote_tracking_refs_alone`
- AC-10: met — `branch_gc_cli` test `ac10_both_commands_are_in_the_help_listing_with_a_summary`, and `help_layout` `help_matches_its_goldens` against `tests/fixtures/help/{all,all-dumb,all-60,all-80}.txt`, which a second commit rewrote with `MOAI_HELP_GOLDEN_WRITE=1` once the card that held that folder had landed; the change is 9 added lines (the two commands in each file, three lines in the 60-column file) and no removed line.

The new tests failed 16 of 16 against the stub before the code (commit `3c93173c84`). The `check affected` receipt of the final head is on the card's READY note; an earlier run at `d7eb281cf3`, before the goldens were rewritten, had 2398 of 2399 tests passing and failed only `help_matches_its_goldens`.

Mutants (each removes one guard from `src/branch_gc.rs`, run on the committed code): merged check, checked-out check, card-reserved check, preserved check, archive prefix, integration-branch preserved, symbolic-ref check, unavailable-ledger keep, digest check, permission check, tip re-check, checkout re-check, `-d` to `-D`, option-like ref guard — 14 of 14 killed. Two (tip re-check, checkout re-check) survived the first test file because git refused the same cases itself; the test was sharpened and both are killed. Not covered by a mutant: the read-back after `git branch -d` (REQ-11 `failed`), which cannot be reached while git deletes the ref.

Dangerous path, run with the real binary in a throwaway HOME and repository (hostile branch names `feat/$(touch${IFS}PWNED)`, `feat/;touch${IFS}PWNED2`, a backtick name and a non-ASCII name; a linked worktree; a reserved card; a remote): the plan listed 5 delete rows and 6 keep rows with the right reasons; a wrong digest refused with `stale_plan` and 11 branches stayed; the plan's digest deleted the 5 branches, no `PWNED` file was created, the remote and the remote-tracking refs were byte-identical, and a second plan found 0 to delete.

A read-only `branch-gc-plan` on this clone (`--integration-ref v2026.11.0`): 1862 branches, 1226 `delete`, 636 `keep` (418 `not_merged`, 94 `checked_out`, 85 `checked_out` and `not_merged`, 36 with a card reservation, 3 `preserved`); the branch count was 1862 before and after.
