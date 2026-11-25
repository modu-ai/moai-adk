---
id: SPEC-20261125-land-unrelated-tracked-changes
title: "card land refuses only the tracked changes of the integration worktree that overlap the paths the landing changes"
status: draft
revision: 1
created: 2026-11-25
branch: spec/land-unrelated-tracked-changes-20261125
---

# card land refuses only the tracked changes of the integration worktree that overlap the paths the landing changes

## Interview

- none: 사용자는 쉬는 중이라 인터뷰를 하지 않았어요. 요구사항은 이 SPEC 을 낳은 카드 본문과 기준 2개, 리더 배차(2026-11-25: tier 2 라서 SPEC 초안과 결정·추천 기본값을 먼저 적고, 승인은 위임에 따라 리더가 추천안으로 기록해요. 착지 관문은 약하게 하지 않아요)가 정했어요. 이 SPEC 의 승인은 사용자 몫이고 이 파일은 승인 기록을 쓰지 않아요. 상태는 `draft` 로 두어요.

## Intent

2026-11-26 00:50~01:00 KST, 다른 저장소(mo.ai.kr)의 리더가 감사 PASS 카드를 `moai card land <card> --head 3d48041efd… --by leader --project <mo.ai.kr>` 로 착지하려다 거절당했어요: `… has 16770 tracked change(s) … commit, stash or discard them there … nothing was changed`. 착지가 건드리는 21개 파일에는 추적 변경이 0건이었고, 변경의 정체는 리더 소유가 아닌 것들이었어요(`.moai/reports` 아래 16,760개 삭제, update 산물 7개, 규칙 파일 3개). 감사 PASS 카드 4장이 review 에서 못 나갔어요.

원인은 `src/card/follow.rs` 의 `check_clean` 이에요. 통합 브랜치를 체크아웃한 워크트리에 추적 변경이 하나라도 있으면, 착지가 그 파일을 건드리는지 보지 않고 거절해요(`SPEC-20261125-land-checked-out-branch` 의 REQ-3: "every tracked change refuses first"). 그 SPEC 은 `git read-tree -m -u` 가 "착지가 건드리지 않는 경로의 로컬 변경은 그대로 둔다"는 것을 알고도 단순함을 위해 전부 거절했어요(그 SPEC 의 Risks 마지막 두 줄). 공유 체크아웃은 늘 다른 세션의 변경이 있어서 이 거절은 착지를 멈추는 쪽으로 너무 넓어요.

이 SPEC 은 거절을 착지와 겹치는 변경으로 좁혀요. 착지가 바꾸는 경로(`diff-tree from..to`)와 워크트리의 추적 변경 경로가 같은 파일이면 여전히 거절하고(관문 유지), 겹치지 않는 변경은 git 의 두 트리 병합이 하던 대로 그대로 두고 따라가요. 거절 문구에 겹침 수와 전체 수를 쓰고, 다음 할 일을 공유 체크아웃 계약(stash·reset·남의 변경 커밋 금지)과 맞춰요. 리더가 stale_base 머지 커밋을 자기 워크트리에서 만드는 방법을 흐름 문서에 적어요. 판정(PASS), 넓은 검사, compare-and-swap, `stale_base`, `tip_moved`, 보호 훅은 바꾸지 않아요.

## Non-goals

- 판정 PASS, 넓은 검사, `stale_base`, `tip_moved`, drift 거절, 보호 훅(`reference-transaction`)은 바꾸지 않아요. 착지는 정확한 head 의 독립 PASS 없이는 여전히 움직이지 않아요.
- 진행 중인 작업의 상태(merge·rebase 등), HEAD 불일치, sparse checkout, skip-worktree·assume-unchanged, 서브모듈 항목, 추적되지 않는 파일이 추가 경로에 서 있는 경우의 거절(`check_state`, `check_entry_tags`, `check_paths`)은 그대로예요.
- `card land` 가 stale_base 후보의 머지 커밋을 스스로 만들지 않아요(D-5). 다른 브랜치를 통합 브랜치에 되머지하는 리더 전용 명령은 C-532 의 일이에요.
- `card drift --restore`, `git push`, `git merge`, `git reset --hard`, `git stash` 는 쓰지 않아요.
- 진짜 저장소의 `NeXT` 는 이 카드에서 옮기지 않아요.

## Requirements

- REQ-1: When the worktree that has the integration branch checked out has tracked changes, the landing shall refuse before the first write only when a tracked change is at a path the landing changes.
- REQ-2 (REQ-1): The landing shall count a tracked change (staged, unstaged or deleted) as overlapping when its path equals a path that `git diff-tree -r --no-renames <tip> <head>` lists.
- REQ-3 (REQ-1): IF a tracked change overlaps, THEN the refusal shall name the count of overlapping changes, the count of all tracked changes, the overlapping paths and `nothing was changed`.
- REQ-4 (REQ-3): The refusal shall name two next steps: the owner commits or discards the changes on a branch of their own, or the lander lands from a worktree of its own.
- REQ-5 (REQ-4): The refusal shall not tell anyone to stash, reset or commit changes that are not theirs.
- REQ-6 (REQ-1): IF every tracked change is at a path the landing leaves alone, THEN the landing shall move the branch, bring the worktree along and close the card.
- REQ-7 (REQ-6): Every tracked change at a path the landing leaves alone shall be as it was after the landing, in its content and in whether it was staged.
- REQ-8 (REQ-6): IF the index differs from the new commit at a path the landing changed after the follow step, THEN the landing shall refuse with the message of the earlier follow check.
- REQ-9 (REQ-1): The landing shall run the dry run of the two-tree merge, the state checks and the added-path check as it did before this change.
- REQ-10 (REQ-1): IF a call stopped after the branch moved and before the worktree followed, THEN the same call run again shall apply REQ-1 to the paths between the recorded old tip and the tip.
- REQ-11 (REQ-10): IF a tracked change overlaps when a call is run again, THEN the call shall refuse with the counts of REQ-3 and the way to run the same call again.
- REQ-12: The answer of `card land` shall carry the number of tracked changes the follow step left alone as `worktree_follow.kept_changes`.
- REQ-13: The flow reference and the leader rules shall say that a tracked change refuses a landing only where it overlaps a path the landing changes.
- REQ-14 (REQ-13): The leader rules shall say how the lander makes the merge of a `stale_base` head in a worktree of its own.

## Acceptance

- AC-1 (REQ-1, REQ-6, REQ-7, REQ-12): a checkout with a staged change, an unstaged change and a deleted tracked file at paths the candidate leaves alone is followed: the branch, the card and the candidate's own paths are right, and the three changes are as they were (same content, still staged where they were) · verify: `cargo test --locked --offline --test card card_land_checked_out_cli::unrelated_tracked_changes_are_kept_and_the_landing_follows`
- AC-2 (REQ-1, REQ-2, REQ-3): a staged change, an unstaged change and a deleted tracked file at a path the candidate changes each refuse beside two unrelated changes, change nothing and say `1 of its 3 tracked change(s)` and the path · verify: `cargo test --locked --offline --test card card_land_checked_out_cli::overlapping_tracked_changes_still_refuse_with_the_counts`
- AC-3 (REQ-4, REQ-5): the refusal names the owner and a worktree of the lander's own, and holds none of `stash`, `reset` or `commit them there` · verify: `cargo test --locked --offline --test card card_land_checked_out_cli::the_overlap_refusal_follows_the_shared_checkout_contract`
- AC-4 (REQ-6, REQ-7): a checkout with 1,200 unstaged deletions of tracked files the candidate leaves alone is followed, and the deletions are still deletions · verify: `cargo test --locked --offline --test card card_land_checked_out_cli::a_thousand_unrelated_deletions_follow`
- AC-5 (REQ-8, REQ-9): the existing guards still hold: a dry run git refuses stops the landing before the move, and a follow that leaves the index behind at a changed path is refused after the move · verify: `cargo test --locked --offline --test card -- card_land_checked_out_cli::a_dry_run_that_git_refuses_stops_the_landing_before_the_move card_land_checked_out_cli::a_follow_that_leaves_the_index_behind_is_refused_after_the_move`
- AC-6 (REQ-10, REQ-11): a stopped run is finished by the same call past unrelated tracked changes, and the same call refuses an overlapping change with the counts · verify: `cargo test --locked --offline --test card card_land_checked_out_cli::a_rerun_follows_past_unrelated_tracked_changes_and_refuses_overlapping_ones`
- AC-7 (REQ-9): every other landing, batch and guard test passes · verify: `cargo test --locked --offline --test card -- card_land_checked_out_cli:: card_land_cli:: card_land_batch_cli:: card_guard_cli:: card_land_scans_cli::`
- AC-8 (REQ-13, REQ-14): `docs/leader-rules.md` and the flow reference say the overlap rule, and the leader rules say how the merge of a `stale_base` head is made · verify: `manual: grep -c overlap docs/leader-rules.md skills/moai/references/flows/next-flow.md` prints one or more for each file
- AC-9: the Result holds the throwaway runs of the dangerous path (a checkout with unrelated changes, one with an overlapping change, one with 1,200 deletions, a rerun) with the command and the observed answer · verify: `manual: the Result lines of this SPEC`

## Scope

- write: `.moai/specs/SPEC-20261125-land-unrelated-tracked-changes/**`
- write: `src/card/follow.rs` (`check_clean`, `apply`'s final check and the lag test of `rerun`; about 120 lines), `tests/card/card_land_checked_out_cli.rs` (the rewritten AC-2 test and the new tests)
- write: `docs/leader-rules.md` (the landing paragraph), `skills/moai/references/flows/next-flow.md` (words replaced, the byte budget has no headroom) and, when that file changes, the sha256 line of `skills/moai/references/source-manifest.json` under `card line-lease`
- unchanged: `src/card/land.rs` (the call site keeps its three calls), `src/card/land/**`, `src/card/landed.rs`
- forbidden: every settings file, `CLAUDE.md`, the real repository's `NeXT`

## Plan

1. Research: measure `git read-tree -m -u` on tracked changes at paths the move leaves alone and at paths it changes (Appendix A).
2. RED: the new tests and the rewritten AC-2 test of `tests/card/card_land_checked_out_cli.rs`; each fails on the current code, which refuses every tracked change.
3. GREEN: `follow.rs` — the set of paths the move changes, the overlap check in place of `check_clean`, the final check of `apply` by those paths, the same rule in `rerun`, and the `kept_changes` key.
4. Guard mutants (tier 2): one per new guard (the overlap check, the counts, the final check by path, the rerun overlap), run in one heavy slot on the committed head and restored with git.
5. Dangerous path: the runs of AC-9 in throwaway repositories, never the real `NeXT`.
6. Docs: the overlap rule and the merge of a `stale_base` head, in the leader rules and (replacing words) the flow reference.
7. Result: each AC as met, not met or not_run with commands and exit codes; `cargo fmt --all`; one bare `moai check affected --card`.

## Risks

- `git read-tree -m -u <old> <new>` keeps a local change at a path the new commit leaves alone, but a file deleted in the worktree at a path the move changes is brought back without a word (Appendix A, `deleted_moved`: exit 0, the file reappears with the new content). The overlap check is the only guard for that case, so it covers a deleted file exactly like an edited one (REQ-2).
- A staged change at an unmoved path stays staged after the follow, so the old "index equals the candidate" check would refuse the landing after the move. REQ-8 compares the index with the candidate only at the changed paths.
- The paths the move changes can be thousands; they are read from `diff-tree` output into a set, never passed as arguments.
- The checked-out worktree is shared. A change another session makes between the plan and the follow can only be at an unmoved path (then it is kept) or at a moved path (then `read-tree` refuses for an edit, and a deletion is restored; the window is the milliseconds between `apply`'s second read and git's scan, the same window as `SPEC-20261125-land-checked-out-branch`).
- SPEC-20261125-land-sync-merge also writes `follow.rs`, the leader rules and the flow reference. This SPEC touches only `check_clean`, the final check of `apply`, the lag test of `rerun`, one paragraph of the leader rules and replaced words of the flow reference; the later merge reads the diff of each.

## Decisions

| Id | Decision | Basis | Recommended default | Undo |
| --- | --- | --- | --- | --- |
| D-1 | A tracked change overlaps when its path equals a path `diff-tree -r --no-renames <tip> <head>` lists; only overlapping changes refuse | the card's (가); a file or folder swap lists the removed file as `D`, so path equality covers it (Appendix A) | adopt | restore `check_clean` |
| D-2 | The refusal reads `<tree> has <M> tracked change(s), <N> of them at paths the landing changes (<paths>)` and gives two next steps: the owner commits or discards them on a branch of their own, or the lander lands from a worktree of its own | the card's (다) and the shared checkout contract (no stash, reset or commit of others' work) | adopt | revert the wording |
| D-3 | A staged change at an unmoved path is carried like an unstaged one and stays staged; the other choice is to refuse any staged change | `git read-tree -m -u` keeps it (Appendix A, `staged_unmoved`); a refusal would keep stopping landings for the same reason | carry | refuse staged changes again in `check_clean` |
| D-4 | `rerun` applies the same rule to the paths between the recorded old tip and the tip | a first run that now succeeds beside unrelated changes must be finishable beside them; otherwise only discarding others' work finishes it | adopt | keep `rerun` strict |
| D-5 | The merge commit of a `stale_base` head stays the lander's own step in a worktree of its own, described in the leader rules; `card land` does not make it | the card's (나) offers either; making commits inside `card land` changes the command's signing and trailer contract, a tier 2 change of its own, and SPEC-20261125-land-sync-merge already owns the merge-in command | document | a later SPEC |
| D-6 | `apply` ends with the index equal to the candidate at every changed path (read as the set of index differences, intersected with the moved set) | D-3 | adopt | restore the whole-index check and D-3 `refuse` |
| D-7 | The SPEC approval is the user's and is not recorded here | the user's decision on gate changes | n/a | n/a |

### Appendix A — measurement

git 2.x of this machine (`git --version` read in the session), throwaway repositories under `target/c540/runs/` (the script `target/c540/probe.sh`, output `target/c540/probe.out`, run 2026-11-25T16:28Z, the repository's own `git` with no global or system config). `read-tree -m -u <C0> <C1>`, where C1 changes `a.txt` and adds `new.txt` and leaves `b.txt` and `d/x` alone. The dry run `-n` and the real run gave the same exit code in every case.

| Case | Exit | Effect |
| --- | --- | --- |
| unstaged edit of `b.txt` (unmoved) | 0 | edit kept, still unstaged |
| staged edit of `b.txt` (unmoved) | 0 | edit kept, still staged |
| `b.txt` deleted in the worktree (unmoved) | 0 | stays deleted |
| `b.txt` removed with `git rm` (unmoved) | 0 | stays removed and staged |
| `d/` deleted in the worktree (unmoved) | 0 | stays deleted |
| unstaged edit of `a.txt` (moved) | 128, `Entry 'a.txt' not uptodate. Cannot merge.` | nothing changed |
| staged edit of `a.txt` (moved) | 128, `Entry 'a.txt' would be overwritten by merge. Cannot merge.` | nothing changed |
| `a.txt` deleted in the worktree (moved) | 0 | **the file is brought back with the new content** (`a1`), the deletion is lost: git does not refuse, so the command's own overlap check is the guard |

Sources already read for the earlier SPEC (`SPEC-20261125-land-checked-out-branch`, Appendix A, read 2026-11-24): `git-read-tree` (the two-tree form is `git switch`'s carry-forward), `unpack-trees.c` v2.54.0. This SPEC adds no new source; the table above is the new measurement.

## Result

`status: draft`; the approval is the user's and was recorded by the leader's delegated approval note on the card (digest `sha256:947c4562553732159a4fcb6a8d7a3ad4953fa6595aaafc6300c98e9d019039c8`, revision 1); it is not recorded in this file. Implemented in `src/card/follow.rs` (the call sites in `src/card/land.rs` are unchanged: `answer` reads `kept_changes` itself). The checks of the READY head are in the card's READY note.

Acceptance (each line is a result of this run on this machine, the commands quoted):

- AC-1 met: `unrelated_tracked_changes_are_kept_and_the_landing_follows` passes (`cargo test --locked --offline --test card -- card_land_checked_out_cli::` → `22 passed; 0 failed`, 2026-11-25).
- AC-2 met: `overlapping_tracked_changes_still_refuse_with_the_counts` passes in that run. Observed refusal, three cases (staged, unstaged, deleted `a.txt`), each: `<tmp>/demo has overlapping tracked changes: 1 of its 3 tracked change(s) are at paths the landing changes (a.txt); …; nothing was changed`.
- AC-3 met: `the_overlap_refusal_follows_the_shared_checkout_contract` passes in that run. Observed: `… the landing never touches work it does not own: the owner of those changes commits or discards them on a branch of their own, or you land from a worktree of your own that has the integration branch checked out, then run the call again; nothing was changed`.
- AC-4 met: `a_thousand_unrelated_deletions_follow` passes in that run (1,200 tracked files deleted, `kept_changes` 1200, the deletions still deletions).
- AC-5 met: `a_dry_run_that_git_refuses_stops_the_landing_before_the_move` and `a_follow_that_leaves_the_index_behind_is_refused_after_the_move` pass in that run.
- AC-6 met: `a_rerun_follows_past_unrelated_tracked_changes_and_refuses_overlapping_ones` passes in that run (finished with `kept_changes` 2; refused with `1 of its 2 tracked change(s)`, then finished).
- AC-7 met: `cargo test --locked --offline --test card -- card_land_checked_out_cli:: card_land_cli:: card_land_batch_cli:: card_guard_cli:: card_land_scans_cli::` → `137 passed; 0 failed`. Two existing tests were rewritten because they held the old refusal of every tracked change: `a_tracked_change_refuses_and_changes_nothing` (replaced by AC-2's test) and the edit case of `a_rerun_refuses_what_a_first_run_would_refuse` (an edit at an unmoved path is no obstacle any more); `ac6_a_checked_out_branch_and_a_head_that_is_no_commit_refuse_before_the_move` of `tests/card/card_land_cli.rs` now lands a candidate that changes the edited `a.txt`.
- AC-8 met: `grep -c overlap docs/leader-rules.md skills/moai/references/flows/next-flow.md` → `2` and `1`; the flow reference is 8111 bytes (the pinned limit is 8121; the sentence is paid by shortening other words) and the sha256 line of `source-manifest.json` follows.
- AC-9 met: the dangerous path ran in throwaway repositories with a temporary HOME and the built binary (the card CLI tests): staged, unstaged and deleted changes at unmoved paths followed byte for byte (status and file contents equal before and after, the staged change still staged); the same three at a changed path refused with the answer of AC-2, changed nothing and left the card open; 1,200 deletions followed; a rerun finished past unrelated changes and refused an overlapping one. Observed in `target/c540/danger2.log`.

Guard mutants (one heavy slot, the committed head, restored with git; `target/c540/mutants.out`): `overlap_refusal_removed` exit 101 (3 tests fail), `deleted_file_not_counted` exit 101 (4), `rerun_overlap_skipped` exit 101 (1), `final_check_dropped` exit 101 (1), `final_check_everywhere` exit 101 (1), `kept_changes_zero` exit 101 (3).

Differences from the approved text, none of them weakening a gate:

- The refusal reads `<tree> has overlapping tracked changes: N of its M tracked change(s) are at paths the landing changes (<paths>)`, the wording of AC-2, instead of D-2's template; both carry the two counts and the paths.
- The way out names a worktree of the lander's own that has the integration branch checked out, because the landing follows only the worktree that holds the branch; the leader rules say that a merge worktree alone does not lift the refusal.
- REQ-10 on a rerun applies to edits and deletions in the working tree; an unrelated staged change on a rerun still makes the index match neither commit, and the rerun leaves the worktree alone with its note as before (the safe side; no test changed for it).
- The write scope grew by `tests/card/card_land_cli.rs` (card scope event, reason recorded), found by AC-7.
- The docs step was done in this session, not by `moai-docs`.
