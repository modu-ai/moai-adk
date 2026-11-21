---
id: SPEC-20261120-reclaim-dead-holder
title: "A reservation is reclaimed before its lease ends when its holder session is proven gone"
status: draft
revision: 1
created: 2026-11-20
---

# A reservation is reclaimed before its lease ends when its holder session is proven gone

## Intent

한 세션의 터미널이 사라졌는데도 그 세션의 예약이 임대(lease) 4시간 동안 남아서, 같은 파일을 쓸 다른 세션이 약 10시간 기다린 일이 있었어요. 지금 `card reclaim` 은 임대가 끝난 예약만 받아요. 이 SPEC 은 사용자 결정(2026-11-20, 리더 대화 선택 "허용")을 그대로 구현해요: 임대가 남아 있어도, 아래 세 증거가 모두 있을 때만 reclaim 을 받아요.

- 홀더 세션의 생존 판정이 `launcher-exited` 로 답하고(런처 생존 판정 조회, 읽기 전용), 그 세션의 `.live` 파일이 실제로 있으며, 레지스트리 임대도 이미 끝났어요. `.live` 가 없는 세션(`session-register` 만 한 Orca 창, 그냥 `claude`)은 조회가 `launcher-exited` 로 답하지만 런처가 돈 적이 없으니 증거가 아니에요(감사 FAIL B1, 리더 결정 2026-11-21).
- 예약에 적힌 워크트리가 사라졌거나, 커밋 안 한 변경과 푸시 안 한 커밋이 없고, 그 브랜치의 stash 와 어느 워크트리도 체크아웃하지 않은 로컬 브랜치의 미푸시 커밋, 재생성 목록 밖의 ignore 파일도 없어요.
- 그 워크트리를 작업 폴더로 쓰는 프로세스가 없어요.

증거가 하나라도 없거나 읽을 수 없으면 지금처럼 거절해요. 임대 기본 4시간, reclaim 평생 3회, 원장 스키마 v1 은 그대로예요. 증거는 reclaim 사건의 사유와 내용에 남겨요.

## Non-goals

- 생존 판정 자체와 레지스트리 기록(런처 생존 카드의 일).
- 임대 기본값, reclaim 한도, 원장 스키마 변경(사용자 결정 필요).
- 자기 세션의 예약 갱신(자기 reclaim)의 동작 변경.
- 임대가 이미 끝난 예약의 reclaim 규칙 변경.
- 실행 워크플로 문서(`mode-agent`·`mode-leader`)의 문장: 그 파일은 다른 카드가 쥐고 있어요.

## Requirements

- REQ-1: The `card reclaim` command shall take over a held reservation whose lease has not ended when the liveness lookup of its holder session answers launcher-exited and its recorded worktree is gone or clean and no process has that worktree as its working folder.
- REQ-2: The `card reclaim` command shall refuse a reservation with a live lease and name the first missing proof when any of the three proofs is missing.
- REQ-3: The worktree proof shall count a recorded path that is not absolute or cannot be read as unproven and refuse a worktree with uncommitted changes or unpushed commits.
- REQ-4: The process proof shall read the working folders of the running processes without writing anything and count a platform or a tool that cannot answer or does not answer within twenty seconds as unproven.
- REQ-5: The reclaim event of an early reclaim shall carry the proof in its reason and in its own payload field.
- REQ-6: The ledger transaction of an early reclaim shall refuse the reclaim when the current holder is not the holder the proof was made for.
- REQ-7: The reclaim limit of three and the lease default of four hours shall stay as they are for an early reclaim.
- REQ-8: The early reclaim shall leave the ledger tables and the schema version as they are.
- REQ-9: A reservation whose lease has ended shall be reclaimed as before and need no proof.
- REQ-10: The proof shall read the registry and the ledger and the worktree and write none of them.
- REQ-11: The `card reclaim` command shall refuse before any write a reason that is too long to carry the proof.
- REQ-12: The worktree check shall count every changed tracked file and every untracked file except one named in the list of tool cache files, the statusline's GitHub counts under either state folder.
- REQ-13: The early proof shall count a holder session's launcher as exited only when the launcher lock file of that session exists and no process holds it and the registry lease of that session has ended.
- REQ-14: The early proof shall count a stash entry of the worktree's branch and unpushed commits on a local branch that no worktree has checked out and an ignored entry outside the list of regenerable entries as work that refuses the reclaim.

## Acceptance

- AC-1 (REQ-1, REQ-5): with a launcher-exited answer from an unlocked launcher lock file and an ended registry lease and a clean worktree and no process in it, a reclaim of a live lease succeeds and the event reason names the proof.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli early_reclaim_succeeds`
- AC-2 (REQ-2): an alive answer and an unknown answer each leave the live lease refused, with the missing proof named and nothing written.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli refused_without_the_liveness_proof`
- AC-3 (REQ-3): a worktree with an uncommitted change, a worktree with an unpushed commit and a reservation with no recorded worktree each leave the live lease refused, and the proof's own worktree check refuses a dirty worktree when it is asked without the takeover gate before it.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli refused_without_the_worktree_proof` and `cargo test --locked --offline --test card_reclaim_early_cli the_early_proof_checks_the_worktree_by_itself`
- AC-4 (REQ-4): a process whose working folder is inside the worktree leaves the live lease refused and a folder that only shares the name prefix does not, and a tool that has not answered within its limit is killed and counts as unproven.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli process_in_the_worktree` and `cargo test --locked --offline --lib stdout_within_kills_a_command_that_overruns`
- AC-5 (REQ-5, REQ-8): the event keeps its kind and carries the proof in a payload field while the ledger tables stay the same.
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
- AC-11 (REQ-13): a session registered without a `.live` file keeps its live reservation even with a clean pushed worktree, an unlocked `.live` file with a registry lease that still runs keeps it too, and an unlocked `.live` file with an ended registry lease lets the reclaim through.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli a_session_without_a_launcher_record_is_not_gone`
- AC-12 (REQ-14): a stash entry of the worktree's branch, unpushed commits on a local branch nobody has checked out and an ignored entry outside the list each keep the live lease refused, while the log folder, the cargo output folder, a Python bytecode folder and another worktree's own unpushed branch do not.
  verify: `cargo test --locked --offline --test card_reclaim_early_cli work_the_takeover_gate_misses` and `cargo test --locked --offline --lib only_listed_ignored_entries_are_regenerable`

## Scope

- write: `.moai/specs/SPEC-20261120-reclaim-dead-holder/**`, `src/card/early_reclaim.rs`, `src/card/mod.rs` (one module line), `src/card/dispatch.rs`, `src/ledger/reservation.rs` (a new function beside `reclaim`), `tests/card_reclaim_early_cli.rs`
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
| D-5 | the early path alone counts a stash entry, an unpushed local branch and an ignored entry as work; the takeover gate of an ended lease stays as it was | REQ-9 keeps the ended-lease reclaim as before, and a live lease is where a missed draft costs most | move the three checks into the shared gate |
| D-6 | a stash entry counts when its subject names the worktree's own branch (`WIP on <branch>:` or `On <branch>:`); an unpushed local branch counts only when no worktree has it checked out; an ignored entry counts unless it is the log folder, the cargo output folder or a Python bytecode folder | the stash stack and the branch list belong to the whole clone, so counting all of them would make every worktree of a busy clone unprovable; every worktree here carries the log folder and the bytecode folders | widen or narrow the list by a later decision |

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
