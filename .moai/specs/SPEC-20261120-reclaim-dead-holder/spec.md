---
id: SPEC-20261120-reclaim-dead-holder
title: "A reservation is reclaimed before its lease ends when its holder session is proven gone"
status: draft
revision: 2
created: 2026-11-20
---

# A reservation is reclaimed before its lease ends when its holder session is proven gone

## Intent

한 세션의 터미널이 사라졌는데도 그 세션의 예약이 임대(lease) 4시간 동안 남아서, 같은 파일을 쓸 다른 세션이 약 10시간 기다린 일이 있었어요. 지금 `card reclaim` 은 임대가 끝난 예약만 받아요. 이 SPEC 은 사용자 결정(2026-11-20, 리더 대화 선택 "허용")을 그대로 구현해요: 임대가 남아 있어도, 아래 세 증거가 모두 있을 때만 reclaim 을 받아요.

- 홀더 세션의 생존 판정이 `launcher-exited` 로 답하고(런처 생존 판정 조회, 읽기 전용), 그 세션의 `.live` 파일이 실제로 있으며, 레지스트리 임대도 이미 끝났어요. `.live` 가 없는 세션(`session-register` 만 한 Orca 창, 그냥 `claude`)은 조회가 `launcher-exited` 로 답하지만 런처가 돈 적이 없으니 증거가 아니에요(감사 FAIL B1, 리더 결정 2026-11-21).
- 예약에 적힌 워크트리가 사라졌거나, 커밋 안 한 변경과 푸시 안 한 커밋이 없고, 그 워크트리 브랜치의 stash 와 예약에 적힌 브랜치의 미푸시 커밋, 재생성 목록 밖의 ignore 파일도 없어요(그 밖의 로컬 브랜치는 세지 않아요, 리더 결정 2026-11-21). M-Flow 프로젝트(프로필의 `work_flow` 가 `m-flow` 이고 통합 브랜치가 로컬에 있을 때)에서는 리모트 추적 ref 에도 로컬 통합 브랜치에도 없는 커밋만 "푸시 안 한 커밋"으로 세요(리더 결정, 2026-11-21).
- 그 워크트리를 작업 폴더로 쓰는 프로세스가 없어요.

증거가 하나라도 없거나 읽을 수 없으면 지금처럼 거절해요. 임대 기본 4시간, reclaim 평생 3회, 원장 스키마 v1 은 그대로예요. 증거는 reclaim 사건의 사유와 내용에 남겨요.

## Non-goals

- 생존 판정 자체와 레지스트리 기록(런처 생존 카드의 일).
- 임대 기본값, reclaim 한도, 원장 스키마 변경(사용자 결정 필요).
- 자기 세션의 예약 갱신(자기 reclaim)의 동작 변경.
- 임대가 이미 끝난 예약의 reclaim 규칙 변경.
- 실행 워크플로 문서(`mode-agent`·`mode-leader`)의 문장: 그 파일은 다른 카드가 쥐고 있어요.

## Requirements

- REQ-1: The `card reclaim` command shall take over a held reservation whose lease has not ended when the liveness lookup of its holder session answers launcher-exited from a launcher lock file that exists and that no process holds and the registry lease of that session has ended and its recorded worktree is gone or clean and no process has that worktree as its working folder.
- REQ-2: The `card reclaim` command shall refuse a reservation with a live lease and name the first missing proof when any of the three proofs is missing.
- REQ-3: The worktree proof shall count a recorded path that is not absolute or cannot be read as unproven and refuse a worktree with uncommitted changes or unpushed commits or a stash entry of its branch or unpushed commits on the branch the reservation names or an ignored entry outside the list of regenerable entries.
- REQ-4: The process proof shall read the working folders of the running processes without writing anything and count a platform or a tool that cannot answer or does not answer within twenty seconds as unproven.
- REQ-5: The reclaim event of an early reclaim shall carry the proof in its reason and in its own payload field.
- REQ-6: The ledger transaction of an early reclaim shall refuse the reclaim when the current holder is not the holder the proof was made for.
- REQ-7: The early reclaim shall keep the reclaim limit of three and the lease default of four hours and leave the ledger tables and the schema version as they are.
- REQ-9: A reservation whose lease has ended shall be reclaimed as before and need no proof.
- REQ-10: The proof shall read the registry and the ledger and the worktree and write none of them.
- REQ-11: The `card reclaim` command shall refuse before any write a reason that is too long to carry the proof.
- REQ-12: The worktree check shall count every changed tracked file and every untracked file except one named in the list of tool cache files, the statusline's GitHub counts under either state folder.
- REQ-13: The takeover gate and the early proof shall count a commit as unpushed work only when it is on no remote-tracking ref and not on the integration branch that the m-flow profile of the project folder (else of the recorded worktree) names, where that branch exists as a local branch; a profile that cannot be read or is not m-flow or names no such branch in either place leaves the count to the remote-tracking refs alone.

## Acceptance

- AC-1 (REQ-1, REQ-5): with a launcher-exited answer from an unlocked launcher lock file and an ended registry lease and a clean worktree and no process in it, a reclaim of a live lease succeeds and the event reason names the proof.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli early_reclaim_succeeds`
- AC-2 (REQ-2): an alive answer and an unknown answer each leave the live lease refused, with the missing proof named and nothing written.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli refused_without_the_liveness_proof`
- AC-3 (REQ-3): a worktree with an uncommitted change, a worktree with an unpushed commit and a reservation with no recorded worktree each leave the live lease refused, and the proof's own worktree check refuses a dirty worktree when it is asked without the takeover gate before it.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli refused_without_the_worktree_proof` and `cargo test --locked --offline --test card_reclaim_early_cli the_early_proof_checks_the_worktree_by_itself`
- AC-4 (REQ-4): a process whose working folder is inside the worktree leaves the live lease refused and a folder that only shares the name prefix does not, and a tool that has not answered within its limit is killed and counts as unproven.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli process_in_the_worktree` and `cargo test --locked --offline --lib stdout_within_kills_a_command_that_overruns`
- AC-5 (REQ-5, REQ-7): the event keeps its kind and carries the proof in a payload field while the ledger tables stay the same.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli event_names_the_proof`
- AC-6 (REQ-6): the ledger refuses an early reclaim whose proof names another holder than the current one.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli reclaim_early_holder`
- AC-7 (REQ-7): a card with three reclaims refuses a fourth early reclaim and the new lease is four hours.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli early_reclaim_counts`
- AC-8 (REQ-9, REQ-10): the reclaim of an ended lease works without any registry, and the proof changes no file of the registry or the worktree.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli --test card_reclaim_cli`
- AC-9 (REQ-11): a reason of the longest allowed length is refused with the proof's length named and nothing written.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli reason_too_long`
- AC-10 (REQ-12): a worktree whose only untracked file is the listed cache counts as clean, and another untracked file, an unlisted name beside the cache and a changed tracked file each keep it dirty with the count naming only the work.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli a_listed_tool_cache` and `cargo test --locked --offline --lib only_an_untracked_listed_file_is_a_cache`
- AC-11 (REQ-1, REQ-2): a session registered without a `.live` file keeps its live reservation even with a clean pushed worktree, an unlocked `.live` file with a registry lease that still runs keeps it too, and an unlocked `.live` file with an ended registry lease lets the reclaim through.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli a_session_without_a_launcher_record_is_not_gone`
- AC-12 (REQ-3): a stash entry of the worktree's branch, unpushed commits on the branch the reservation names and an ignored entry outside the list each keep the live lease refused, while the log folder, the cargo output folder, a Python bytecode folder, a stash entry of another branch and an unrelated local branch with unpushed commits that is checked out nowhere do not.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli work_the_takeover_gate_misses` and `cargo test --locked --offline --lib only_listed_ignored_entries_are_regenerable`
- AC-13 (REQ-13): a clean worktree with four commits ahead of the remote that all sit in the local integration branch of an m-flow profile passes, whether that profile is the project folder's (the worktree's own profile has no work flow) or the worktree's own, the takeover gate and the early proof, a fifth commit outside that branch refuses and the count names one unpushed commit, the branch the reservation names passes when it is ahead only inside the integration branch and refuses with one commit outside it, and a profile that is not m-flow or names a branch missing locally keeps the refusal.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli integration_branch` and `cargo test --locked --offline --test card_reclaim_cli integration_branch`

## Scope

- write: `.moai/specs/SPEC-20261120-reclaim-dead-holder/**`, `src/card/early_reclaim.rs`, `src/card/mod.rs` (one module line), `src/card/dispatch.rs`, `src/ledger/reservation.rs` (a new function beside `reclaim`), `tests/card_reclaim_early_cli.rs`, `tests/card_reclaim_cli.rs`
- write: `src/lane_registry.rs` (a read-only function beside `liveness_of_session` that reports the launcher lock file and the registry lease of a session)
- read-only: every other file, the real HOME
- depends on: the liveness lookup `lane_registry::liveness_of_session` of the launcher-liveness SPEC; the two heads land together

## Plan

1. SPEC first (this file).
2. RED: the new integration test file and a lib test for the holder check.
3. `early_reclaim.rs`: the three proofs, each read-only, each answering proven or a named reason.
4. `reservation::reclaim_early`: the shared transaction of `reclaim` with a proven holder and without the lease check; `reclaim` calls it with none.
5. `dispatch::reclaim`: on a live lease build the proof, fold it into the reason and call the early function.
6. The unnarrowed READY set, then an independent audit (tier 2: a lease rule).
7. Repair (revision 3, leader decisions 2026-11-21): RED tests for REQ-3 and REQ-13 in both reclaim test files, then one helper in `dispatch.rs` that reads the integration branch from the project folder's profile (else the worktree's), used by the takeover gate and by the branch check of `early_reclaim.rs`, which now looks at the reservation's own branch and the worktree's HEAD with its stash and no other local branch; the `stdout_within` helper gets the same cfg as its callers.

## Risks

- A proof taken just before the write can age: the holder is proven gone, so only a new session with the same id could change it; the ledger transaction re-checks the holder and the generation.
- `lsof` or `/proc` may be missing or denied: the process proof then answers unproven and the old refusal stays.
- The liveness lookup lives in another SPEC's change; until it is in the tip this change builds on its pushed head.
- The registry is a local, unsigned folder: a forged entry could answer launcher-exited for a session that still runs. The other two proofs still stop a holder that is working, because it has uncommitted work or a process in its worktree; the liveness card owns the forging question.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the early path is automatic: `card reclaim` has no new flag | the user decision names the conditions, not an option | add a flag later |
| D-2 | a reservation with no recorded worktree is unproven | the clean-or-gone proof needs a path | allow it by a later decision |
| D-3 | the proof goes into the reason when it fits and always into a payload field | the user asked for the reason; the payload is machine-readable | drop the payload field |
| D-4 | the cache list names whole files, not patterns, and the worktree check lists untracked files one by one (`--untracked-files=all`) | a folder that git shows as one entry could hide real work next to the cache; the leader note limited the exemption to listed cache paths | empty the list |
| D-5 | the early path alone counts a stash entry, unpushed commits on the reserved branch and an ignored entry as work; the takeover gate of an ended lease stays as it was | REQ-9 keeps the ended-lease reclaim as before, and a live lease is where a missed draft costs most | move the three checks into the shared gate |
| D-6 | a stash entry counts when its subject names the worktree's own branch (`WIP on <branch>:` or `On <branch>:`); unpushed commits count only on the branch the reservation names (the worktree's HEAD is the takeover gate's) and on no other local branch; an ignored entry counts unless it is the log folder, the cargo output folder or a Python bytecode folder | the stash stack and the branch list belong to the whole clone, so counting all of them would make every worktree of a busy clone unprovable (this repository has 1607 local branches and 44 of them hold unpushed commits that no worktree has checked out, so the early reclaim could never succeed; leader decision 2026-11-21); every worktree here carries the log folder and the bytecode folders | widen or narrow the list by a later decision |
| D-7 | in an m-flow project a commit already in the local integration branch is not unpushed work, for the takeover gate and for the reserved-branch check of the early proof; the profile is read from the project folder first and from the recorded worktree only when the project folder gives no m-flow integration branch, and a profile that cannot be read in either place means "not m-flow"; the ref checks and the count run in the worktree, whose refs are the clone's | a lane worktree whose commits all landed in the local integration branch has nothing left to lose, and the leader chose this on 2026-11-21; a lane worktree has no state-root marker (it is ignored by git), so its own profile is the earlier `.moai-cli` one without a work flow, and the m-flow profile is the project folder's; an unreadable profile stays on today's safe side | drop the integration branch from the count |

## Result

Head `5e18015bbb`, built on the pushed head `436f34c93f` of the launcher-liveness repair branch (the liveness lookup it needs is not in the local integration tip yet; the two land together). The unnarrowed `moai check affected a4d4c9a164..5e18015bbb` ended `result=PASS` (diff_check, fmt, ops_rustfmt, clippy, test_lib, test_bins, test_targets, build_cli and python all exit 0), run with `CLAUDE_CODE_SESSION_ID` removed from the environment.

- AC-1: met — `early_reclaim_succeeds` in `tests/card_reclaim_early_cli.rs`
- AC-2: met — `refused_without_the_liveness_proof`
- AC-3: met — `refused_without_the_worktree_proof`
- AC-4: met — `process_in_the_worktree`
- AC-5: met — `event_names_the_proof`
- AC-6: met — `reclaim_early_holder`
- AC-7: met — `early_reclaim_counts`
- AC-8: met — `card_reclaim_early_cli` and `card_reclaim_cli` both pass in the affected run
- AC-9: met — `reason_too_long`
- AC-10: met — `a_listed_tool_cache_is_not_work` and the lib test `only_an_untracked_listed_file_is_a_cache`
- AC-11: met — `a_session_without_a_launcher_record_is_not_gone`
- AC-12: met — `work_the_takeover_gate_misses_keeps_the_early_reclaim_refused`, `an_unrelated_local_branch_with_unpushed_commits_does_not_stop_an_early_reclaim` and the lib test `only_listed_ignored_entries_are_regenerable`
- AC-13: met — `commits_inside_the_integration_branch_do_not_stop_an_early_reclaim` and `commits_inside_the_integration_branch_do_not_stop_a_takeover` (four commits ahead of the remote, all inside the local integration branch), `one_commit_outside_the_integration_branch_still_stops_an_early_reclaim` and `one_commit_outside_the_integration_branch_still_stops_a_takeover` (four plus one refuse naming one), `the_project_folders_integration_branch_counts_when_the_worktree_profile_has_no_work_flow` in both files (the profile of the project folder, the worktree profile without a work flow), `the_reserved_branch_ahead_only_inside_the_integration_branch_is_not_work`, and `without_a_usable_integration_branch_an_early_reclaim_counts_every_unpushed_commit` with its takeover gate twin (six cases each: no profile, another work flow, another work flow that names a local branch, m-flow without a branch, m-flow with a branch this clone lacks, an unreadable profile)

Repair of the tier-2 audit FAIL (a running session registered without a launcher passed as exited), code head `77ca0149b3`: `cargo test --locked --offline --no-fail-fast --test card_reclaim_early_cli --test card_reclaim_cli` ended exit 0 (14 and 16 passed), `cargo test --locked --offline --lib early_reclaim` ended exit 0 (5 passed), run with `CLAUDE_CODE_SESSION_ID` removed. Three mutants of the early proof (the `.live` file check removed, the registry lease check removed, the extra work check not called) each turned `card_reclaim_early_cli` red; the file was restored and the tree was clean afterwards. The proof's own worktree check is now asked directly by `the_early_proof_checks_the_worktree_by_itself` (AC-3), and the `lsof` limit by `stdout_within_kills_a_command_that_overruns` (AC-4). Not done here: the card number in the body of the launcher-liveness branch's commit (that branch is another change's).

Repair of the second tier-2 audit (the early proof walked every local branch of the clone, so it could not succeed in a clone with many branches, and a commit already in the local integration branch counted as unpushed), code heads `b880c9e4e2` and `da38b6216b`: `cargo test --locked --offline --no-fail-fast --test card_reclaim_early_cli --test card_reclaim_cli` ended exit 0 (21 and 21 passed) and `cargo test --locked --offline --lib early_reclaim` ended exit 0 (5 passed); the delta `moai check affected 2519484227e7..b880c9e4e2a6` ended `result=PASS` (targets 63; diff_check, fmt, clippy, test_lib, test_bins, test_targets, build_cli and python all exit 0). The proof now counts the stash of the worktree's own branch, the reservation's own branch and the worktree HEAD, and a commit counts as unpushed only when it is on no remote-tracking ref and not in the local integration branch of an m-flow profile, read from the project folder first because a lane worktree has no state-root marker. Mutants: the project folder profile not read, the integration branch not excluded, the reservation's own branch not checked and the worktree fallback removed each turned the two reclaim test files red, and the audit's work flow check removed turned both red after the github-flow case was added (before it, that mutant survived); each time the file was restored byte for byte and the tree was clean. The cfg alignment of the `lsof` helper was read back in this lane, not compiled for Windows.
