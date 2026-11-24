---
id: SPEC-20261125-land-checked-out-branch
title: "card land moves the integration branch even when a worktree has it checked out, and brings that worktree along"
status: draft
revision: 3
created: 2026-11-25
branch: spec/20261125-land-checked-out-branch
---

# card land moves the integration branch even when a worktree has it checked out, and brings that worktree along

## Interview

- none: 사용자는 자고 있어서 인터뷰를 하지 않았어요. 요구사항은 이 SPEC 을 낳은 카드의 본문과 일곱 기준, 리더 배차(2026-11-25: git 공식 문서 조사, SPEC 초안, 초안 위에서 RED 테스트와 구현, 버리는 저장소의 위험 경로 실행, READY, 착지는 아침에 사용자 승인 뒤)가 정했어요. 이 SPEC 의 승인은 흐름 규칙("통합 브랜치는 착지 명령으로만 움직여요")을 바꾸는 일이라 사용자 몫이고 리더도 대신 승인하지 않아요. 상태는 `draft` 로 두고 승인 기록은 이 파일에 쓰지 않아요.

## Intent

2026-11-24T17:00Z 리더 라운드에서 `moai card land <card> --head 77f6cbf41d… --by leader --keep-open` (a lane slice) 이 `conflict` 로 거절됐어요: `the integration branch NeXT is checked out in <the primary checkout>: a move under it would leave that worktree behind the reference`. 주 체크아웃은 늘 `NeXT` 에 있어서(공유 체크아웃에서 브랜치를 바꾸지 않는 규칙) 이 거절은 모든 착지에서 나요.

지금은 리더가 주 체크아웃에서 손으로 `git merge --ff-only` 를 한 뒤 `card finish` 로 닫아요(`NeXT` reflog 의 `merge …: Fast-forward`, 사건의 `via: finish`). 그것은 "통합 브랜치는 착지 명령으로만 움직인다"는 규칙과 어긋나고, 호스트의 auto mode 분류기가 거절해요. 감사가 끝난 head 가 착지하지 못하고 쌓여요. 자기 개선 규칙의 "수동 리더 단계의 반복"에 해당해요.

이 SPEC 은 착지 명령이 그 일을 안전하게 하게 해요. 브랜치가 체크아웃된 워크트리를 찾고, 그 워크트리가 깨끗하고 HEAD 가 읽은 tip 과 같으면 `git update-ref` 의 compare-and-swap 한 번으로 브랜치를 옮긴 뒤 그 워크트리의 색인과 작업 트리를 새 커밋으로 맞춰요(fast-forward 와 같은 결과). 더럽거나 다른 커밋이면 아무것도 바꾸지 않고 거절하고 이유와 다음 할 일을 말해요. 판정(PASS), `stale_base`, `tip_moved`, 기준 검사는 그대로예요. 문턱을 낮추지 않아요.

현재 코드의 거절은 `src/card/land.rs` 의 `Standing::OnTip` 분기(`landed::checked_out_at` 가 비어 있지 않으면 `conflict`)에 있고, 이동은 `landed::move_branch` 의 `git update-ref -m <msg> <ref> <new> <old>` 한 번이에요. 같은 거절이 `card drift --restore`(`src/card/drift.rs`)에도 있어요.

## Non-goals

- `card drift --restore` 는 바꾸지 않아요: 그것은 브랜치를 되돌리는 일이라 워크트리를 함께 되돌리면 작업을 잃을 수 있어요. 거절을 그대로 둬요(D-4).
- 판정 PASS, `stale_base`, `tip_moved`, drift 거절, 기준 검사, 기록 줄(start row)의 순서는 바꾸지 않아요.
- `git push`, `receive.denyCurrentBranch`, `push-to-checkout` 훅, `git merge`, `git reset --hard` 는 쓰지 않아요(조사 결과의 기각 근거는 Appendix A, D-1).
- 통합 브랜치를 손으로, reset 으로, push 로 옮기는 것은 여전히 금지예요. 보호 훅(`reference-transaction`)은 그대로예요.
- 진짜 저장소의 `NeXT` 는 이 카드에서 옮기지 않아요. 착지는 아침에 사용자 승인 뒤에 해요.
- 다른 워크트리(주 체크아웃이 아닌 연결된 워크트리)가 브랜치를 체크아웃한 경우도 같은 규칙이에요. 둘 이상이면 지원하지 않고 거절해요.

## Requirements

- REQ-1: When `card land` moves the integration branch and exactly one worktree has that branch checked out, the command shall bring that worktree's index and working tree to the candidate commit after the branch moves.
- REQ-2 (REQ-1): The command shall move the branch only by `git update-ref` with the tip it read as the old value.
- REQ-3 (REQ-1): IF the checked-out worktree has a tracked change (a staged change, an unstaged change or a deleted tracked file), THEN the command shall refuse before the first write and name the paths.
- REQ-4 (REQ-1): IF the HEAD of the checked-out worktree is not the tip the command read, or a merge, a rebase, a cherry-pick, a revert, a bisect or an index lock is in progress there, THEN the command shall refuse before the first write.
- REQ-5 (REQ-1): IF a file of the checked-out worktree that no commit tracks, ignored or not, stands at a path the candidate adds, or where the candidate needs a folder, THEN the command shall refuse before the first write and name the paths.
- REQ-6 (REQ-1): IF more than one worktree has the branch checked out, or the checked-out worktree uses a sparse checkout or has a skip-worktree or assume-unchanged entry, or the move changes a submodule entry, THEN the command shall refuse before the first write and name the cause.
- REQ-7 (REQ-1): Every refusal of the checked-out worktree shall carry the reason, the next thing to do and the words `nothing was changed`.
- REQ-8 (REQ-1): IF a call stopped after the branch moved and before the worktree followed, THEN the same call run again shall finish the worktree step and close the card without moving the branch a second time.
- REQ-9 (REQ-1): The answer of `card land` shall carry a `worktree_follow` key with the worktree path, the status (`followed`, `finished` or `none`), and the commits it went from and to.
- REQ-10: The landing shall answer every refusal that comes before the worktree check as it did before this change.
- REQ-11: The flow reference and the leader rules shall describe a landing with the branch checked out as the command's own step, with the refusals of REQ-3 to REQ-6.
- REQ-12: The Appendix shall record the git documentation read for this change, each source with its read time, and an adopt or reject decision with its reason for every option the card names.
- REQ-13 (REQ-8): IF the index of the checked-out worktree equals neither the old tip's tree nor the new one's when a call is run again, THEN the command shall leave that worktree alone, close the card and say so in `worktree_follow` with the way to bring that worktree along.
- REQ-14 (REQ-5): IF a file that no commit tracks, ignored or not, stands at a path the candidate adds when the follow is about to write, THEN the command shall leave that file untouched and refuse after the move with the reason and the way to run the same call again.
- REQ-15 (REQ-5): IF a tracked path stands where the candidate adds one and the same move deletes that path (a tracked folder the candidate turns into a file, a file renamed by case only), THEN the command shall follow the worktree without refusing for that path.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9): a clean checkout with the integration branch checked out is moved by `card land`: the branch, the index and the working tree equal the candidate, an untracked file that does not overlap stays, the card closes, and `worktree_follow` names the path and both commits · verify: `cargo test --locked --offline --test card_land_checked_out_cli a_clean_checkout_follows_the_branch_and_the_card_closes`
- AC-2 (REQ-3, REQ-7): a staged change, an unstaged change and a deleted tracked file each refuse, change nothing (the branch, the ledger, the index and the working tree) and say the reason and the next step · verify: `cargo test --locked --offline --test card_land_checked_out_cli a_tracked_change_refuses_and_changes_nothing`
- AC-3 (REQ-4, REQ-7): a worktree whose HEAD is another commit than the tip read, and a worktree in a merge, a rebase, a cherry-pick, a revert, a bisect or with an index lock, each refuse and change nothing · verify: `cargo test --locked --offline --test card_land_checked_out_cli a_moved_head_or_a_pending_operation_refuses_and_changes_nothing`
- AC-4 (REQ-5, REQ-7): an untracked file and an ignored file at a path the candidate adds each refuse and change nothing, while an untracked file elsewhere does not stop the landing · verify: `cargo test --locked --offline --test card_land_checked_out_cli an_untracked_file_at_an_added_path_refuses_and_other_untracked_files_stay`
- AC-5 (REQ-6, REQ-7): two worktrees on the branch, a sparse checkout, a skip-worktree entry and a submodule change each refuse and change nothing · verify: `cargo test --locked --offline --test card_land_checked_out_cli an_unsupported_checkout_refuses_and_changes_nothing`
- AC-6 (REQ-8): a run that fails at the worktree step after the branch moved leaves the card open and `worktree_follow` unfinished; the same call run again moves the branch no second time, finishes the worktree and closes the card · verify: `cargo test --locked --offline --test card_land_checked_out_cli a_stopped_run_is_finished_by_the_same_call_and_moves_once`
- AC-7 (REQ-8): a run whose process is killed with SIGKILL between the branch move and the worktree step is finished the same way · verify: `cargo test --locked --offline --test card_land_checked_out_cli a_killed_run_is_finished_by_the_same_call`
- AC-8 (REQ-10): every other land, batch and guard test passes unchanged; the one test that held the old refusal of a clean checkout is rewritten to hold a refusal that stays (a dirty checkout) and is named in the Result · verify: `cargo test --locked --offline --test card_land_cli --test card_land_batch_cli --test card_guard_cli --test card_land_scans_cli`
- AC-9 (REQ-11): `docs/leader-rules.md` and the flow reference name the follow step, its four refusals and the rerun · verify: `manual: grep -c worktree_follow docs/leader-rules.md skills/moai/references/flows/next-flow.md` prints one or more for each file
- AC-10 (REQ-12): Appendix A holds the sources with read times and an adopt or reject line for each option of the card · verify: `moai spec check SPEC-20261125-land-checked-out-branch --project .`
- AC-12 (REQ-5, REQ-8, REQ-13): the guards that only one rule fails are held by their own tests: a tracked file the candidate replaces by a folder follows, a dry run that git refuses stops the landing before the move, a `read-tree` that exits 0 and leaves the index behind is refused after the move, a run killed after the follow closes on the rerun without following again, a rerun refuses what a first run would refuse, and a rerun leaves an index that matches neither commit alone · verify: `cargo test --locked --offline --test card_land_checked_out_cli`
- AC-13 (REQ-14): an ignored file that a git wrapper creates right after the branch move, at a path the candidate adds, is never overwritten: the call refuses after the move and says so, the card stays open, and the same call run again after the file is gone finishes the worktree · verify: `cargo test --locked --offline --test card_land_checked_out_cli an_ignored_file_that_appears_after_the_plan_is_never_overwritten`
- AC-14 (REQ-15): a tracked folder the candidate turns into a file follows, a folder that also holds a file no commit tracks refuses, and a case-only rename follows · verify: `cargo test --locked --offline --test card_land_checked_out_cli a_tracked_folder_the_candidate_turns_into_a_file_follows a_folder_with_a_stranger_in_it_is_not_turned_into_a_file a_case_only_rename_follows`
- AC-11: the Result holds the throwaway runs of the dangerous path: a clean checkout, a dirty checkout, an untracked file over a path of the new commit, another worktree on the branch, and a rerun of a run that died in the middle, each with the command and the observed answer · verify: `manual: the Result lines of this SPEC`

## Scope

- write: `.moai/specs/SPEC-20261125-land-checked-out-branch/**`
- write: `src/card/follow.rs` (new: the plan, the preconditions, the follow step and the finish of a stopped run), `src/card/mod.rs` (the `mod follow;` line), `src/card/landed.rs` (read helpers the follow step shares)
- write, held by the batch-landing card until it gives them back, so written only under `moai card line-lease` or after that card releases them: `src/card/land.rs` (the call site, about 25 lines)
- write: `tests/card_land_checked_out_cli.rs` (new), `tests/card_land_cli.rs` (the one test of D-6)
- write: `docs/leader-rules.md`; `skills/moai/references/flows/next-flow.md` (held by the batch-landing card; its byte budget has no headroom, so the new wording replaces words) and, when that file changes, the sha256 line of `skills/moai/references/source-manifest.json` by the project's own command
- unchanged: `src/card/land/**` (held by the batch-landing card; the batch calls `land` for every head, so each head follows the worktree on its own), `src/card/land/check.rs` (the registry-tier card), `src/card/drift.rs` (D-4)
- forbidden: every settings file, `CLAUDE.md`, the real repository's `NeXT`

## Plan

1. Research: git's own documentation and the source of `update_worktree()` in `receive-pack.c`, with a measurement of `git read-tree -m -u` in a throwaway repository (Appendix A); the adopt and reject lines of the card's options.
2. RED: `tests/card_land_checked_out_cli.rs` with the eight tests above; each fails on the current code because the command refuses a clean checkout. Rewrite the one old test of D-6.
3. GREEN: `src/card/follow.rs` — `plan` (the worktree and every precondition, read only, with the dry run of the two-tree merge), `apply` (after the move) and `finish` (a stopped run, from the start row's tip); then the call site in `land.rs` under a line lease.
4. Guard mutants: one per refusal and per rule (REQ-3 to REQ-8), run in one heavy slot on the committed head, each restored with git.
5. Dangerous path: the five throwaway runs of AC-11 in temporary repositories, never the real `NeXT`.
6. Docs: the leader rules and the flow reference say the new behaviour in their own words.
7. Result: each AC as met, not met or not_run, with the commands and exit codes; the bare `moai check affected` of the final head.

## Risks

- The ledger start row, the reference guard intent and the branch move stay as they are; the new step comes after the move, so a failure there leaves the branch moved. Mitigation: REQ-8 — the same call finds the head inside the tip and finishes the worktree from the tip the start row recorded; a worktree whose index is neither at that old tree nor at the new one is left alone, the card closes with `status: none` and a note that tells its owner how to bring it along (REQ-13; the branch moved and the landing is done, so no later call could repair it, and the leftover is the owner's own staged work mixed with the lag: `git diff --cached <old tip>` lists exactly the owner's changes, then `git reset --hard` brings the checkout to the new tip).
- `git read-tree` overwrites an ignored file at an added path without a word and has no switch that refuses it (Appendix B). The plan reads the added paths before the first write and `apply` reads them again right before `read-tree`, because the ledger rows, the scans and the branch move sit between the two. A file created in the few milliseconds between that second read and git's own scan is not caught; a git wrapper that writes the file right before `read-tree` still overwrites it. The risk is accepted: the window is not the minutes a landing takes, and no documented lock closes it.
- Seen by reading the code, not observed: a card kept open with `--keep-open` and called again with the same head cannot tell a user who staged the whole old tree by hand from a lagging index; and a concurrent call that lost the compare-and-swap closes without a follow. `MOAI_GIT_BIN` is read by release builds too; this change did not add that.
- The primary checkout is shared by every lane and the leader. A file another session left changed there refuses the landing, which is the safe side; the answer names the paths.
- `git read-tree -m -u` keeps a local change to a file the new commit does not touch. The precondition of REQ-3 refuses every tracked change first, so that case never reaches it.
- The follow step writes inside a worktree the command does not own. Mitigation: nothing is written before every precondition and a dry run of the same two-tree merge pass; the write is one `git read-tree -m -u <old tree> <new tree>`, git's own, and no file is removed or overwritten by hand.
- The flow reference has no byte headroom (`FLOW_REF_TIP_BYTES`); the new wording replaces words and is measured.
- The batch-landing card holds `land.rs`, `land/**` and the flow reference now; the work that needs them waits for its release or a line lease. Nothing here reads or writes the real repository's `NeXT`.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The branch moves by `git update-ref` first and the worktree follows with `git read-tree -m -u <old tree> <new tree>` | the compare-and-swap decides the truth; a stopped run leaves a state a rerun can finish (Appendix A, questions 1 to 4) | remove the follow step; the refusal returns |
| D-2 | Before the move a dry run (`git read-tree -m -u -n`) of the same two-tree merge must pass in the worktree | refuses an overlap of untracked files before any write (Appendix A, question 3) | none needed |
| D-3 | An untracked file that does not stand on a path the candidate adds does not stop the landing; a file git ignores at such a path does stop it | the primary checkout always holds some (`.moai/state/`); `updateInstead` ignores untracked files the same way; measured case f: `git read-tree -m -u` exits 0 and overwrites an ignored file at an added path without a word, so the command checks those paths itself | tighten to refuse any untracked file |
| D-4 | `card drift --restore` keeps its refusal | it moves the branch backward; following it would delete work | a later SPEC |
| D-5 | A stopped run is recognised from the tip the ledger start row of this card and head recorded: the index equals that old commit's tree and the working tree equals the index | a staged change by a person cannot be told from a lagging index without the old tip | none needed |
| D-6 | The test `ac6_a_checked_out_branch_and_a_head_that_is_no_commit_refuse_before_the_move` of `tests/card_land_cli.rs` is rewritten: its checked-out case becomes a dirty checkout, which still refuses | the old case (a clean checkout) is the behaviour this SPEC changes | revert the commit |
| D-7 | The SPEC approval is the user's and is not recorded here | the user's decision on flow-rule changes | n/a |

### Appendix A — git documentation read

Research by the `moai-research` agent on 2026-11-24 (UTC) between 17:26:05Z and about 17:31Z, git 2.54.0 (Apple Git-157) on this machine. `RAW` means the page text was read with `curl`, `WF` means `WebFetch`, whose quotes are close but not byte-verified; the per-page time after 17:26:49Z was not clocked. The agent's own throwaway runs were refused by the worktree guard, so the measurements below were made by this lane with `target/c483/probe.sh` (throwaway repositories under the session scratch folder, run 2026-11-24T17:31:19Z, output `target/c483/probe1.out`).

Sources: `git-update-ref` (RAW, 2.53.0), `git-read-tree` (RAW), `receive-pack.c` at tag v2.54.0 (RAW, read 17:26:14Z), `git-merge` (WF, 2.54.0), `githooks` push-to-checkout (WF, 2.54.0), `git-worktree` (WF), `git-checkout` and `git-switch` (WF), `git-status` (WF), `gitrepository-layout` (WF), `git-update-index` (WF), `git-rev-parse` (WF), `git-push` and `git-config` (WF, the `receive.denyCurrentBranch` text was cut short), and `unpack-trees.c`, `read-tree.c`, `wt-status.c` at v2.54.0 (WF). Every URL is under `https://git-scm.com/docs/` or `https://raw.githubusercontent.com/git/git/v2.54.0/`.

| Option of the card | Decision | Reason and source |
| --- | --- | --- |
| `git update-ref <ref> <new> <old>` on a checked-out branch | ADOPT as the one move | `git-update-ref` (RAW): it "stores the <new-oid> in the <ref> … after verifying that the current value of the <ref> matches <old-oid>"; it says nothing of the index or working tree, so they stay at the old commit (measured: case `crash`, `git status` shows the reverse diff `M a.txt`, `D ign.txt`, `D new.txt`) and a second `update-ref` with the old value fails (`cannot lock ref 'refs/heads/B': is at … but expected …`, exit 128) |
| `receive.denyCurrentBranch=updateInstead` | REJECT as the mechanism, ADOPT its preconditions | it works only through a push to the repository; `receive-pack.c` v2.54.0 `update_worktree()` runs `git update-index -q --ignore-submodules --refresh`, `git diff-files --quiet --ignore-submodules --`, `git diff-index --quiet --cached --ignore-submodules HEAD`, then `git read-tree -u -m <new>` (one-tree form), and it moves the worktree first and the reference second, so a lost compare-and-swap leaves the worktree ahead of a reference someone else moved |
| `push-to-checkout` hook | REJECT | same push path; the hook doc (WF): "when both the working tree and the index match the current commit, they are updated to match the newly pushed tip"; it only names `git read-tree -u -m HEAD "$1"` as the example body |
| `git merge --ff-only` inside the worktree | REJECT | a second writer of the integration reference besides `landed::move_branch`; whether it compares against the value it read first was not verified (research: "not verified"); it runs the worktree's `post-merge` hook and words its refusals itself |
| `git reset --hard`, `git read-tree --reset` | REJECT | `git-read-tree` (RAW): with `-u` "updates leading to loss of working tree changes or untracked files or directories will not abort the operation" |
| `git read-tree -m -u <old> <new>` after the reference moved | ADOPT | the two-tree form is `git switch`'s own carry-forward rule; measured exit codes in the table below; idempotent when the worktree is already at `<new>` (case `rerun`, exit 0 twice) |
| `git read-tree -m -u -n` before the move | ADOPT as the dry run | `git-read-tree` (RAW): "-n … Check if the command would error out, without updating the index or the files in the working tree for real"; measured: the dry run's exit code and message equalled the real run's in all ten cases |

Measured with `read-tree -m -u <C0> <C1>` where C1 changes `a.txt`, adds `new.txt` and a force-added ignored `ign.txt`, and leaves `b.txt` alone (`probe1.out`):

| Case | Exit | Effect |
| --- | --- | --- |
| clean | 0 | index and files follow |
| unstaged edit of a path C1 changes | 128, `Entry 'a.txt' not uptodate. Cannot merge.` | nothing changed |
| unstaged edit of a path C1 leaves alone | 0 | the local edit is kept: the command refuses every tracked change itself first (REQ-3) |
| staged edit of a path C1 leaves alone | 0 | the staged edit is kept: same |
| staged edit of a path C1 changes | 128, `Entry 'a.txt' would be overwritten by merge. Cannot merge.` | nothing changed |
| untracked file at an added path | 128, `Untracked working tree file 'new.txt' would be overwritten by merge.` | nothing changed |
| untracked file elsewhere | 0 | stays |
| ignored file at an added path | 0 | **overwritten silently** (`ignored-local` became `i1`): the command checks added paths itself (REQ-5, D-3) |
| deleted tracked file C1 leaves alone / changes | 0 / 0 | stays deleted / is restored: refused earlier by REQ-3 |

Stopped run (case `crash`, `rerun`): after `update-ref` alone `git diff-index --cached --quiet <new>` exits 1 and `--quiet <old>` exits 0, `git diff-files --quiet` exits 0; `git read-tree -m -u <old> <new>` then leaves `git status` empty, and a second call also exits 0. A run is therefore recognised by the index tree equal to the old commit's tree with a clean working tree, and finished by the same read-tree (D-1, D-5).

Detection (from `wt-status.c` and the repository layout, WF; `MERGE_HEAD` resolved per worktree and observed: `.git/MERGE_HEAD` exists during a `merge --no-commit`): `MERGE_HEAD`, `CHERRY_PICK_HEAD`, `REVERT_HEAD`, `REBASE_HEAD`, the folders `rebase-merge`, `rebase-apply`, `sequencer` and the file `BISECT_LOG` under the worktree's git dir (`git rev-parse --absolute-git-dir` run in that worktree), and `index.lock`. Several worktrees on one branch exist only through `git worktree add --force` or `checkout --ignore-other-worktrees` (WF); `git worktree list --porcelain` prints one `branch refs/heads/<name>` line per holder (observed). Sparse checkout (`core.sparseCheckout`), skip-worktree and assume-unchanged entries (`git ls-files -v`) and mode 160000 entries are refused without a measurement.

Measured while implementing (git 2.54.0, throwaway repositories): `git ls-files -v` tags `S` for a skip-worktree entry (paths a sparse checkout excludes carry it too), `h` for assume-unchanged and `H` for an ordinary entry, so any tag other than `H` refuses; a local file where the candidate needs a folder passes both the dry run and the real run (exit 0) and leaves the index and the files out of step, so the command checks the ancestors of every added path itself, leaving out a path the same diff deletes; `git diff-files` calls a file rewritten with the same bytes changed (stat data differs) while `git diff` compares content, so the rerun asks `git diff`.

Not verified: every WF quote beyond the cut-off text; the sequencer state path; that `git merge --ff-only` compare-and-swaps; Apple Git-157 against upstream 2.54.0 on these paths; behaviour on Linux or Windows.

### Appendix B — an ignored file at a path the head adds (audit repair, revision 3)

Question: can `git read-tree` be made to refuse an ignored file that stands at a path the new commit adds? Settled by a research agent that read the sources between 2026-11-24T19:45Z and 19:46:34Z (the manual pages with a fetch tool, then the raw `Documentation/*.adoc` and source files at git tag v2.54.0 to check every quote); this lane did not read them itself.

- `git-read-tree` manual (https://git-scm.com/docs/git-read-tree): it never says what happens to an ignored file in the way. `-u`: "After a successful merge, update the files in the work tree with the result of the merge." `--reset`: "When used with `-u`, updates leading to loss of working tree changes or untracked files or directories will not abort the operation."
- `builtin/read-tree.c` at v2.54.0: lines 149-152 define `--exclude-per-directory` with the help text "allow explicitly ignored files to be overwritten" (not in the manual; it only checks its argument), and lines 229-230 read `if (opts.update && !opts.reset) opts.preserve_ignored = 0;`. No read-tree option sets `preserve_ignored`.
- `unpack-trees.c` at v2.54.0: lines 1919-1922 build the exclude rules `if (!o->preserve_ignored)`, and lines 2434-2440 answer `/* ce->name is explicitly excluded, so it is Ok to overwrite it. */ return 0;`. That is why the file is overwritten without a word, which Appendix A case f measured.
- `git-checkout` manual (https://git-scm.com/docs/git-checkout): "`--overwrite-ignore` / `--no-overwrite-ignore`: Silently overwrite ignored files when switching branches. This is the default behavior. Use `--no-overwrite-ignore` to abort the operation when the new branch contains ignored files." `builtin/checkout.c` sets `preserve_ignored` at lines 705 and 773; `git switch` accepts the option in code (lines 1744-1746) without documenting it; `git restore` does not have it.
- Lock: `builtin/read-tree.c` line 191 holds the index lock; neither the manual nor the code read mentions a lock on work-tree files.

Decision. Adopt: read the added paths again right before `read-tree` (`follow::apply`). Reject: `git checkout --no-overwrite-ignore` or `git switch` (they move HEAD or a branch, and the branch is already moved by the one compare-and-swap, a second mover would break the exactly-once rule), `--exclude-per-directory` (it permits the overwrite), `read-tree --reset -u` (it never aborts), and a lock (none documented). What stays: a file created in the milliseconds between that second read and git's own scan is overwritten (Risks).

Not verified: git newer than v2.54.0; that `git checkout --no-overwrite-ignore` aborts as the manual says (not run).

## Result

`status: draft`; the approval is the user's and is not recorded here. Implemented in `src/card/follow.rs`, called from `src/card/land.rs`; the checks of the READY head are in the card's REPAIRED note. Revision 3 repairs the first audit (data loss by an ignored file in the window, a tracked path the same move deletes counted as an obstacle, the help text, a personal path, the unmerged-index wording).

Acceptance (each line is a result of this run, the commands quoted; the new test file is run as `cargo nextest run --locked --offline --test card_land_checked_out_cli` → `17 tests run: 17 passed`, target/c483/show2.log):

- AC-1 met: `a_clean_checkout_follows_the_branch_and_the_card_closes` passes in that run.
- AC-2 met: `a_tracked_change_refuses_and_changes_nothing` (a staged change, an unstaged change, a deleted tracked file) passes in that run.
- AC-3 met: `a_moved_head_or_a_pending_operation_refuses_and_changes_nothing` (nine marks and a HEAD that is not the tip) passes in that run.
- AC-4 met: `an_untracked_file_at_an_added_path_refuses_and_other_untracked_files_stay` (untracked, ignored, a file where a folder is needed) passes in that run.
- AC-5 met: `an_unsupported_checkout_refuses_and_changes_nothing` (two worktrees, a sparse checkout, a skip-worktree entry, an assume-unchanged entry, a submodule entry, an unmerged entry) passes in that run.
- AC-6 met: `a_stopped_run_is_finished_by_the_same_call_and_moves_once` passes in that run.
- AC-7 met: `a_killed_run_is_finished_by_the_same_call` passes in that run.
- AC-8 met: `cargo nextest run --locked --offline --test card_land_cli --test card_land_batch_cli --test card_guard_cli --test card_land_scans_cli` → `108 tests run: 108 passed` (target/c483/ac8.log). The rewritten test is `ac6_a_checked_out_branch_and_a_head_that_is_no_commit_refuse_before_the_move`: its checked-out case now holds a tracked change, which still refuses.
- AC-9 met: `grep -c worktree_follow docs/leader-rules.md skills/moai/references/flows/next-flow.md` → `1` and `1`; the flow reference stays at the 8051 bytes `tests/next_flow_skill_scan.rs` pins (the sentence is paid by shortening other words; the first receipt of this change failed on that limit), the six tests that read the file pass (`158 tests run: 158 passed`, target/c483/green8.log, the file unchanged since), `python3 -I tools/source_manifest.py check` answers `ok: true`; `card land --help` names the follow, its refusals and `worktree_follow` (audit B3 repaired; `help_layout` passes).
- AC-10 met: Appendix A and B as written; `moai spec check SPEC-20261125-land-checked-out-branch --project .` → `errors: 0, warnings: 1` (`size_requirements`: 15 requirements, kept because each added requirement keeps one `shall`).
- AC-11 met for the runs below; no run touched a real repository.
- AC-12 met: the six guard tests of AC-12 pass in that run.
- AC-13 met: `an_ignored_file_that_appears_after_the_plan_is_never_overwritten` passes in that run.
- AC-14 met: `a_tracked_folder_the_candidate_turns_into_a_file_follows`, `a_folder_with_a_stranger_in_it_is_not_turned_into_a_file` and `a_case_only_rename_follows` pass in that run (the last one on this Mac, whose file system ignores case; on one that does not it is a plain delete and add).

Dangerous path (AC-11): each run is the real `moai card land` process on a throwaway home and a throwaway repository whose integration branch `integ` is checked out in the project folder, started by `C483_SHOW=1 cargo nextest run --locked --offline --test card_land_checked_out_cli --no-capture` (target/c483/show.log). Paths below stand for the temporary project folder.

1. Clean checkout (`a_clean_checkout_follows_the_branch_and_the_card_closes`): answer `{"worktree":"<project>","status":"followed","from":"065e9225…","to":"474249ee…","note":null}`; branch, index and files equal the head, an untracked `scratch.txt` stayed.
2. Dirty checkout (`a_tracked_change_refuses_and_changes_nothing`): `<project> has 1 tracked change(s): b.txt; commit, stash or discard them there (the landing never touches work it does not own), then run the call again; nothing was changed`; the same for an unstaged `a.txt` and a deleted `b.txt`; the branch, the ledger rows and the checkout bytes were equal before and after.
3. Untracked file over a path of the new commit (`an_untracked_file_at_an_added_path_refuses_and_other_untracked_files_stay`): `<project> holds 1 file(s) that no commit tracks at a path the new commit adds: notes/x.md; move or delete them (or add them to the commit), then run the call again; nothing was changed`; the same for an ignored `notes/x.md`, and for a file `notes` where the head needs a folder (`… notes (a file where the new commit needs a folder, for notes/x.md) …`); once removed the landing answered `followed` and an untracked `elsewhere.txt` stayed.
4. Another worktree on the branch (`an_unsupported_checkout_refuses_and_changes_nothing`): `the integration branch integ is checked out in more than one worktree (<project>, <second>); leave it checked out in one of them (switch the others to another branch) and run the call again; nothing was changed`; a sparse checkout, a skip-worktree entry (`S: a.txt`), an assume-unchanged entry (`h: a.txt`) and a submodule entry (`vendor/sub`) refused the same way.
5. Rerun of a run that died in the middle (`a_killed_run_is_finished_by_the_same_call`, a git wrapper sent `kill -9` to the lander right after `update-ref`): the rerun answered `{"status":"finished","from":"39da4775…","to":"fbb27492…"}`, moved nothing (`moved: false`, one `card.land_started` row, one reflog entry of the landing) and closed the card; `a_stopped_run_is_finished_by_the_same_call_and_moves_once` (a wrapper made `read-tree` fail) ended the same way, and a worktree whose index matched neither commit was left alone with the note `the index of the worktree is neither the old tip's tree nor the new one's, so it was left alone`.

6. An ignored file that appeared after the plan (`an_ignored_file_that_appears_after_the_plan_is_never_overwritten`; a git wrapper wrote `notes/x.md`, ignored by `.git/info/exclude`, right after `update-ref`): the call answered `<project> holds 1 file(s) that no commit tracks at a path the new commit adds: notes/x.md; move or delete them (or add them to the commit), then run the call again; the integration branch integ already moved to 9a18956f… and the card is still open: run the same call again after the cause is repaired`; the file still held `precious`, the card stayed in review and the worktree lagged; the same call with the file still there refused the same way and the checkout bytes were equal before and after; once the file was moved away the same call answered `finished`, `notes/x.md` held the lane's text, one reflog entry of the landing, the card `done`.
7. A tracked folder the head turns into a file followed (`a_tracked_folder_the_candidate_turns_into_a_file_follows`, answer `followed`, `d` a file); the same folder holding an untracked `d/mine.txt` refused: `<project> holds 1 file(s) that no commit tracks at a path the new commit adds: d/ (a folder that holds files no commit tracks: d/mine.txt); move or delete them (or add them to the commit), then run the call again; nothing was changed`; a case-only rename `b.txt` to `B.txt` followed (`a_case_only_rename_follows`), `git ls-files` listed `B.txt` only.
8. A rerun whose index matched neither commit (`a_rerun_leaves_a_worktree_alone_whose_index_matches_neither_commit`): the card closed and the answer was `{"status":"none", … "note":"the index of the worktree is neither the old tip's tree nor the new one's, so it was left alone and shows the move backwards; `git diff --cached <old tip>` there lists exactly the changes staged by hand: save them (for example `git diff --cached <old tip> > keep.patch`), run `git reset --hard` there to bring it to <new tip>, and apply the patch again"}`; the worktree bytes were equal before and after.

Guard mutants (`python3 -I target/c483/mutants.py follow` and `land`, one heavy slot per batch, exact-once edits restored with `git checkout --`, tree clean after each batch, run on the code head `3f76b85876372b0f36848a7bd389b6dba040e475`; outputs target/c483/mut-follow3.out and target/c483/mut-land2.out): 23 of 23 killed (M1 to M13 and M15 to M22 in `follow.rs`, M30 and M31 in `land.rs`). Earlier runs: the first (head a6da5f0c0c) killed 18 of 18 but did not cover the audit's findings; the repair added M17 to M22 and made M4, M7, M8 and M9 follow the rewritten code; one run left M14 (the rerun's own read of the added paths) alive because `apply` reads the same paths right before the write and nothing a caller sees differs, so that duplicate read was removed with its mutant. M20 (the exemption of a file the same move deletes under another spelling) and M21 are killed by the case-only test, which needs a file system that ignores case; on Linux they would survive.

Not run: a landing on the real repository or the real `NeXT` branch (the user's approval and a morning landing come first); Linux and Windows; Apple Git against upstream git on these paths (git 2.54.0 only); a git wrapper that writes the ignored file right before `read-tree` (the accepted window, Risks); a worktree on a network file system; a repository with `core.fsmonitor` or a submodule that has content. Cosmetic: an unmerged path is listed once per stage (`a.txt, a.txt, a.txt`).
