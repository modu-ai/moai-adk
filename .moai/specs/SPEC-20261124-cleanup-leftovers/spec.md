---
id: SPEC-20261124-cleanup-leftovers
title: "Six small leftovers of the 2026-11-24 card cleanup land as separate slices of one SPEC"
status: draft
revision: 2
created: 2026-11-24
---

# Six small leftovers of the 2026-11-24 card cleanup land as separate slices of one SPEC

## Intent

사용자가 2026-11-24 에 열린 카드를 정리하며 개수를 줄이라고 정했어요. 닫힌 카드 여섯 장에 작은 조각이 하나씩 남았고, 새 카드 여섯 대신 한 카드의 기준 여섯으로 묶였어요. 이 SPEC 은 그 여섯 조각 가운데 지금 할 수 있는 네 조각(기준 6, 3, 5, 2)을 정하고, 지금 할 수 없는 두 조각(기준 1, 4)이 왜 기다리는지 적어요. 조각마다 따로 커밋하고 따로 감사받을 수 있어요.

원본 카드의 마지막 노트를 기준마다 읽고 현재 통합 브랜치(`NeXT`, `546cd451a2`)와 맞춰 봤어요(Evidence).

판 2 (2026-11-24): 조각 A 에서 D 는 `NeXT` `89e8149c87` 에 착지했어요(카드 노트 8171, 감사 노트 8153). 그 뒤 카드에 기준 7 에서 10 이 더해졌고(노트 8098 과 감사 비차단, 착지 중 시험 하나의 깨짐), 판 1 의 범위에는 그 파일들이 없어요. 판 2 는 조각 E 를 더해 그 네 기준을 한 조각으로 묶어요. 조각 A 에서 D 의 요구사항, 인수 기준, 결정은 그대로예요.

## Interview

- none: 사용자와 리더의 결정이 이미 기준 여섯과 순서로 적혀 있어 새로 물을 것이 없어요(배차 글 2026-11-24, 카드 C-446 본문; 사용자 결정은 카드 정리 2026-11-24).

## Non-goals

- 기준 1(하위 `.gitignore`)과 기준 4(설치·업데이트 보고)는 이 SPEC 에서 만들지 않아요. 둘 다 아래 Deferred 에 이유와 다시 시작할 조건이 있어요.
- 출력 스타일의 렌더 본문(`skills/moai/references/home/output-style.md`)과 전역 사본은 건드리지 않아요. 템플릿 세 개만이에요.
- 노트 7169 의 나머지 비차단 항목(`render.rs` 테마 주석, `docs/rust-product-path.md` 예시, `SPEC-20261123-progress-board-flow-icons` 의 superseded 표시)은 기준 문장에 없어서 하지 않아요.
- `docs/rust-product-path.md` 는 다른 카드가 쥐고 있어서 기준 3 에서도 건드리지 않아요.

## Requirements

- REQ-1: Each of the three shipped output-style templates (`moai.md`; `moai-easy.md`; `moai-learn.md`) shall name only the progress icons of the one flow set (🗂️ 🟡 📨 🔍 🔁 🛬 ✅ beside 🚧 🙋 🔴 ❌) and none of ⬜ 🟢 ⏸️ 🔵 👀 ⛔ · failure: none — the retired icons leave the three files and a scan test fails when one returns.
- REQ-2: The update-recovery section of `docs/support-matrix.md` shall describe the Claude projection and the user skill folder that the `release_suite` fixture uses now · failure: none — the second-host sentence is false on the integration branch since the fixture moved to the Claude host alone.
- REQ-3: The pointers in `docs/host-support-matrix.md` and `docs/product-naming-inventory.md` to the deleted Codex adapter page shall say that the page was deleted and name the deleting commit · failure: none — a dangling pointer becomes a dated record.
- REQ-4: The test `the_repository_tracks_no_codex_folder` shall read the committed tree with `git ls-tree` and fail when git itself fails · failure: none — a git failure no longer passes the check as "no such entry".
- REQ-5: The refusal of an unknown `permissions` key shall print the key only through `shown_value` · failure: none — a token-shaped key split by an invisible character is no longer echoed whole.
- REQ-6: Every acceptance screen of a settings value (a `claudeMdExcludes` entry; the `model` value; the other screened values) shall judge the text as written and the text with invisible characters removed · failure: none — a token split by U+200B is refused and nothing is planned.
- REQ-7: The invisible-character set shall name each near-invisible class as put in or left out, with the reason, in the SPEC decisions of the project-scale settings SPEC · failure: none — a class left out (spaces, combining marks, look-alike letters) is written down as a remaining risk, not forgotten.
- REQ-8: The test of the invisible-character screen shall pin one split-token point for each pattern `credential_like` matches · failure: none — a pattern removed from the screen fails its own point.
- REQ-9: A refusal that removed invisible characters shall say `(invisible characters removed)` · failure: none — a stripped name no longer reads as an allowed one (display only).
- REQ-10: The process-scan snapshot slice of `SPEC-20261121-process-scan-snapshot` shall be present on the integration branch with its three deletion-safety tests: a working-directory question reads the cwd list only, a re-check right before a delete or move reads fresh, and a time-limited `lsof` is unknown and refuses · failure: none — a conflict in `src/worktree_gc.rs` is resolved by the lander of this card and audited at tier 2.
- REQ-11: The sentence of `skills/moai/workflows/run.md` that says a failed run after a push "becomes a repair card" shall say that it becomes a criterion of the open follow-up card of the SPEC the failed head names · failure: none — the sentence matches the rule `moai card criterion-add` of `CLAUDE.md` step 7.
- REQ-12: The recurring-problems rule of `CLAUDE.md` shall hold its sentence about a recurrence whose cause already has an open card after "never relaxes a gate, a HOLD or a user rule." · failure: none — "Such a card" then follows the sentence that opens the card.
- REQ-13: Item 4 of the leader's round in `docs/leader-rules.md` shall be written as one short sentence for each of its four refill cases and for the follow-up rule · failure: none — wording only, no case and no order changes.
- REQ-14: The Scope and the Result of `SPEC-20261124-audit-prep` and the page `docs/design-review/appendix/operations/audit_slot.md` shall say that `audit-slot prep` refuses a slot whose own `target/` is a link or a file before its first write · failure: none — the Scope says "no change of behavior" and is false for that refusal.
- REQ-15: When an apply path of `home_retire`, `home_state_retire` or `home_install_apply` re-reads the processes right before it moves or retires an entry, it shall read the working folders and the open files fresh and never from the command cache · failure: none — a process that started after the command's first scan blocks the move; the plan-time reads keep the cache.
- REQ-16: The test `the_workflow_step_posts_the_status_the_gate_decides` shall start the step's shell with `PYTHONDONTWRITEBYTECODE` set, so that it passes whichever `python3` is first on `PATH` · failure: none — a `python3` that writes a cache folder into HOME no longer trips the HOME guard; the workflow file is not changed.

## Acceptance

- AC-1 (REQ-1): none of the three templates holds a retired icon and each names the flow icons it uses · verify: `cargo test --locked --offline --test style_banner_scan the_three_style_templates_use_the_one_icon_set`
- AC-2 (REQ-2, REQ-3): the three documents carry the corrected sentences and rows of commit `9706de1802` and the cherry-pick applies with no conflict on the tip · verify: `git diff --check`
- AC-3 (REQ-4): the check fails on a failing git and passes on an absent `.codex` entry; the old `rev-parse` form fails the same test when put back (guard mutant) · verify: `cargo test --locked --offline --test retired_host_words the_repository_tracks_no_codex_folder`
- AC-4 (REQ-5, REQ-6, REQ-8, REQ-9): the four tests of the C4 patch pass and the lib test of the unknown key passes · verify: `cargo test --locked --offline --test claude_settings_scale_keys`
- AC-5 (REQ-7): the decisions D-6 to D-8 of `SPEC-20261122-project-scale-settings` list each near-invisible class as in or out · verify: `moai spec check SPEC-20261122-project-scale-settings`
- AC-6 (REQ-10): the snapshot slice passes its own tests on the tip: the cwd-only scan, the fresh re-check of a delete and a move, and the time-limit refusal, and the two earlier survivors MB and MH stay killed · verify: `cargo test --locked --offline --test process_scan_snapshot_cli`
- AC-7 (REQ-10): the hygiene re-check test and the cost test of the two neighbours pass · verify: `cargo test --locked --offline --test hygiene_apply_cli`
- AC-8 (REQ-11, REQ-12, REQ-13): the three files carry the corrected sentences, no rule changed its meaning, and the repository scans that read them pass · verify: `git diff --check` and the `moai check affected <base>..<head>` receipt
- AC-9 (REQ-14): the page and the SPEC name the refusal of a link or file `target/`, and the SPEC still checks · verify: `moai spec check SPEC-20261124-audit-prep`
- AC-10 (REQ-15): for each of the three apply paths a stub `lsof` whose first scan is empty and whose later scans show a process inside the entry keeps the entry in place; the plan reads keep the cache; the guard mutant (an apply path back on the cached read) fails its test · verify: `cargo test --locked --offline --test process_scan_snapshot_cli`
- AC-11 (REQ-16): the step test passes with a stand-in `python3` first on `PATH` that writes a cache folder into HOME unless the bytecode switch is set; the guard mutant (the switch removed) fails with the HOME guard · verify: `cargo test --locked --offline --test audit_gate_cli the_workflow_step_posts_the_status_the_gate_decides`

## Scope

- write: `.moai/specs/SPEC-20261124-cleanup-leftovers/**`, `templates/claude/output-styles/moai.md`, `moai-easy.md`, `moai-learn.md`, `tests/style_banner_scan.rs`
- write: `docs/support-matrix.md`, `docs/host-support-matrix.md`, `docs/product-naming-inventory.md`, `tests/retired_host_words.rs`, `.moai/specs/SPEC-20261122-remove-codex-shipped-text/spec.md`
- write: `src/claude.rs` (the value screens only, new lines), `tests/claude_settings_scale_keys.rs`, `.moai/specs/SPEC-20261122-project-scale-settings/spec.md`
- write: `src/worktree_gc.rs`, `src/hygiene.rs`, `tests/process_scan_snapshot_cli.rs`, `tests/process_scan_cost_cli.rs`, `tests/hygiene_apply_cli.rs`, `.moai/specs/SPEC-20261121-process-scan-snapshot/**`
- write (slice E, revision 2): `skills/moai/workflows/run.md` (the one sentence), `CLAUDE.md` (the recurring-problems bullet only; tier 2 floor), `docs/leader-rules.md` (item 4 only)
- write (slice E): `.moai/specs/SPEC-20261124-audit-prep/spec.md` (its Scope and Result text only), `docs/design-review/appendix/operations/audit_slot.md`
- write (slice E): `src/home_state.rs` (one new `InUse` constructor, new lines), `src/home_retire.rs`, `src/home_state_retire.rs`, `src/home_install_apply.rs` (one call each), `tests/process_scan_snapshot_cli.rs`
- write (slice E): `tests/audit_gate_cli.rs`
- read-only: everything else; `docs/rust-product-path.md` stays with its holder
- forbidden: push, pull request, merge of a moved tip, the real HOME, the user's global files, card numbers in SPEC ids, branches and commits

## Plan

Order of slices, one commit group each, so each can be audited and landed alone:

1. Slice A, REQ-1 (card criterion 6): RED `the_three_style_templates_use_the_one_icon_set` fails on the retired icons of the three files; GREEN rewrites the three icon sentences with the flow set; one test only, no source change.
2. Slice B, REQ-2 to REQ-4 (card criterion 3): `git cherry-pick 9706de1802` onto this branch (a dry merge applied with no conflict); re-run `retired_host_words`; run the guard mutant (put the old `rev-parse` form back, the test must fail). The old commit's `Moai-Session` trailer is the earlier lane's; the commit gets this session's trailers when the check demands them.
3. Slice C, REQ-5 to REQ-9 (card criterion 5): the written but never compiled patch `c334c4.patch` of the earlier lane applies on the tip (files: `src/claude.rs`, `tests/claude_settings_scale_keys.rs`, the settings SPEC); RED on the base copy, GREEN, mutants N1 raw key, N2 raw only, U+2800 left out, N5 note removed. Risk named by that lane: the added "(invisible characters removed)" text may change wording that other tests read, so the neighbours of `claude_settings_cli` run too.
4. Slice D, REQ-10 (card criterion 2): merge the branch head `7059238965` of `spec/20261121-process-scan-snapshot` (audit PASS head `e803e25056` for the test delta; the `src` rounds passed earlier) onto the tip. The merge conflicts in `src/worktree_gc.rs` (dry run `git merge-tree`: one conflict, `hygiene_apply_cli` and `process_scan_cost_cli` merge cleanly). The conflict is resolved once and the result is a tier 2 change (it guards deletions), so it goes to the independent audit with the guard-mutant lines MB and MH re-run on the merged tree.

5. Slice E, REQ-11 to REQ-16 (card criteria 7, 8, 9, 10), four small groups, one commit group each; the order puts the code groups first so the tier 2 file comes last:
   - E1 (criterion 9, REQ-15): RED first. The pin test goes into `tests/process_scan_snapshot_cli.rs` when the home fixtures can be shared there; otherwise into one new test file, and the new-test-file precheck of `SPEC-20261124-new-test-file-precheck` runs first. GREEN adds `InUse::read_fresh()` in `src/home_state.rs` (the same three scans as `InUse::read`, the working folders and open files through `process_cwds_fresh` and `open_file_paths_fresh`) and switches the three apply-time calls (`src/home_retire.rs` apply, `src/home_state_retire.rs` apply, `src/home_install_apply.rs` apply). The plan-time calls stay on the cache. Guard mutant: one apply call back on `InUse::read`.
   - E2 (criterion 10, REQ-16): RED is the step test with a stand-in `python3` first on `PATH` that writes a cache folder into HOME unless `PYTHONDONTWRITEBYTECODE` is set (it fails today on any machine, not only with the Xcode python); GREEN gives the step's shell that variable. Guard mutant: the variable removed.
   - E3 (criterion 8, REQ-14): text only: the two sentences in the `SPEC-20261124-audit-prep` Scope and Result and the page `audit_slot.md`; read the refusal in `src/audit_slot.rs` first so the sentence says what the code does.
   - E4 (criterion 7, REQ-11 to REQ-13): text only: three files; `CLAUDE.md` is on the tier 2 floor list, so the whole slice gets the full audit. `CLAUDE.md` changes only through this card's scope, as its own rule says.

### Deferred (not part of slices A to E)

- Criterion 1, C-239 AC2 (nested `.gitignore` in `newly_visible`): still open on 2026-11-24. The earlier block ended: `ce8e7e7513` is an ancestor of `NeXT` (read 2026-11-24, exit 0). The work itself is C-174 criterion 3 (card ledger: met at head `21b77ec6dc`, which is not an ancestor of `NeXT`); on `NeXT` `newly_visible` (`src/agents.rs:685`) still asks git for the ignored files of `tools/` and `evals/` and has no nested-rule test (`git grep a_nested_gitignore_under_the_state_folder_keeps_a_file_out_of_the_visible_list` finds nothing). The work is done once, in C-174, so this card does not copy it (D-1). See D-5 for what this criterion does next.
- Criterion 4, C-331 AC2 (improvement reports at install and update): `not_run`, after RC1 by the leader's dispatch.

## Risks

- Slice D touches the deletion guards (`src/worktree_gc.rs`, `src/hygiene.rs`); a wrong conflict resolution can let a busy folder be deleted. The dangerous path runs in a throwaway: a held-open folder is refused by plan, by re-check and by a time-limited `lsof`.
- Slice C is a patch that was never compiled; an expected compile error or a changed refusal wording is repaired inside slice C, not outside its scope.
- Slice E E1 changes what a home move waits on: a re-check that reads fresh costs one more process scan per apply (a few milliseconds, audit note N1 of 8153) and can only refuse more, never move more. A fresh read that cannot be made (`None`) keeps the entry in place, as the existing `Unknown` answer does.
- Slice E E4 edits `CLAUDE.md`, which every agent loads; a wrong move of one sentence changes a rule's meaning. The slice moves one sentence whole and rewrites no rule, and the full audit reads the before and after side by side.
- Slice E E3 edits an approved SPEC (`SPEC-20261124-audit-prep`, revision 2): its digest changes. D-6 asks the leader how that SPEC records it.
- `src/worktree_move.rs:269` and `src/optimize.rs:876` also read the cached `process_cwds()`; the card names four home files and neither of these. Slice E reads both and reports whether they re-check right before a move, as a note; it changes neither.
- The scope above is wide because one card holds six slices. `src/claude.rs`, `src/worktree_gc.rs` and `src/hygiene.rs` are shared files: a holder of another card that needs them gets a line-lease; this card writes only the lines named in the slice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Criterion 1 waits for C-174; the lane does not copy C-174 code to work around it. | `git grep newly_visible NeXT` finds none; scope discipline. | card open |
| D-2 | Criterion 4 stays `not_run` until after RC1. | the dispatch of the leader, 2026-11-24. | card open |
| D-3 | Slice B cherry-picks the earlier commit as written; its docs text is not edited again. | audit 6935 B1 to B4 are still true on the tip (`docs/support-matrix.md:111`). | revert the commit |
| D-4 | Slice C takes the earlier patch's decisions D-6 to D-8 of the project-scale settings SPEC as the proposal for approval (D-6 judge both texts, D-7 per-class list, D-8 display note). | note 7032 of the source card. | card open |
| D-5 | Criterion 1 waits for the landing of C-174 criterion 3 and is then marked met with that landing as evidence (default). The other option is to withdraw it with the reason "same work as C-174 criterion 3, tracked there". | note 8133 of this card (the work is done once, in C-174); on 2026-11-24 `21b77ec6dc` is not on `NeXT`, so the criterion cannot be met today. | the card keeps the criterion; a withdrawal is one `--withdraw` call |
| D-6 | The approved `SPEC-20261124-audit-prep` is edited in its Scope and Result text only (criterion 8, requirement and acceptance lines untouched). Whether that needs a revision of that SPEC is the leader's pick; the default is no new revision, with the commit message naming the text-only edit and the SPEC's digest line updated by the leader's approval note. | criterion 8 names those two places; a revision for one sentence would reopen an approved SPEC. | revert the commit |
| D-7 | The pin of criterion 10 sets `PYTHONDONTWRITEBYTECODE` on the shell the test starts, and does not change `.github/workflows/audit-gate.yml`. | the workflow runs on a CI runner with no HOME guard; the failure is the test's own HOME guard; a workflow file is outside this card's scope. | one line in the test |
| D-8 | Criterion 9 adds one new constructor `InUse::read_fresh()` and does not add a flag to `InUse::read()`. Plan-time calls keep the cache. | `hygiene` already uses the fresh readers for its re-check (audit PASS 8153, MH); a second constructor keeps each call's intent readable and the mutant a one-word change. | the three calls go back to `read()` |

### Evidence

Read on 2026-11-24 from the card ledger and the repository at `546cd451a2`:

| Criterion | Source read | Finding |
| --- | --- | --- |
| 6 | C-372 notes 7134, 7169 | three files still hold ⬜ 🟢 ⏸️ 🔵 (`moai.md:17`, `moai-easy.md:15`, `moai-learn.md:15`) |
| 3 | C-292 notes 6935, 7000, 7008 | commit `9706de1802` exists locally, is not in `HEAD`, and cherry-picks without conflict |
| 1 | C-239 notes 4638, 4833, 4932; C-174 notes 4597, 4638 | note 4833 is the slice 1 merge note; the open part is N5 of audit 4638, and its code is not on `NeXT` |
| 5 | C-334 notes 6840, 6842, 7032 | C3 (`609a497d1c`) is in `NeXT`; patch `c334c4.patch` (21031 bytes) exists in its worktree |
| 2 | C-275 notes 5717, 5733, 6743, 7062 | slice 2 head `7059238965` is not in `NeXT`; a dry merge has one conflict in `src/worktree_gc.rs` |
| 4 | C-331 | not read; `not_run` by the dispatch |
| 7 | card notes 8098, 8153; `skills/moai/workflows/run.md:60`, `CLAUDE.md:22`, `docs/leader-rules.md:25` at `89e8149c87` | the sentence added by commit `f5b6d5b2f4` sits before "Such a card" in `CLAUDE.md:22`; run.md:60 still says "becomes a repair card"; leader-rules.md item 4 is one long line |
| 8 | C-364 notes 8068, 8103 (R1); `docs/design-review/appendix/operations/audit_slot.md:24`; `SPEC-20261124-audit-prep` Scope | the repair made `audit-slot prep` refuse a link or file `target/` before its first write; the Scope still says "no change of behavior"; the page lists the refusal only for the lane's `target` |
| 9 | AUDIT 8153 N1; `src/home_state.rs:447`; call sites `home_retire.rs:546`, `home_state_retire.rs:393`, `home_install_apply.rs:729` (apply), `home_retire.rs:259`, `home_state_retire.rs:256`, `home_state.rs:559` (plan) | `InUse::read` uses `process_cwds()` and `open_file_paths()`, both behind the command cache (`src/worktree_gc.rs:527`, `:536`); the fresh readers `process_cwds_fresh` and `open_file_paths_fresh` exist (`:553`, `:558`) |
| 10 | landing note 8171; `tests/audit_gate_cli.rs:326-343` | the shell child gets `PATH`, `GH_TOKEN`, `PR_*`, `TMPDIR` and no bytecode switch; `/usr/bin/python3` is 3.9.6 on this machine and the pyenv shim is first on the normal `PATH` |

## Result

- AC-1: met — `style_banner_scan the_three_style_templates_use_the_one_icon_set` 통과. RED(`8964e785fa`): `moai.md still carries the old icon ⬜`, exit 101, 단언 실패. GREEN(`2cd06d691b`)이 세 템플릿의 아이콘 문장을 흐름 한 벌로 바꿔요; 시험의 순서 검사는 `moai-easy.md` 가 한 줄에 배너 `✅` 를 앞서 가져서 각 단계를 앞 단계 뒤에서 찾게 고쳤어요(`6a3b61d723` 은 rustfmt). 뮤턴트: 옛 `⬜` 를 `moai-learn.md` 에 되돌림 → exit 101, `moai.md` 의 `🛬` 단계 삭제 → exit 101.
- AC-2: met — `9706de1802` 를 `c58a746719` 로 cherry-pick(충돌 없음, `moai trailer-check 2cd06d691b..HEAD` failing 0); `docs/support-matrix.md` 의 "두 번째 host" 글이 0 줄이에요; `git diff --check` 0.
- AC-3: met — 옛 `rev-parse` 판정은 git 이 실패해도 통과했어요(`GIT_DIR` 을 없는 곳으로 둔 실행에서 `assert!(out.success)` 를 뺀 뮤턴트가 exit 0). 일반 환경에서는 git 이 실패하지 않아 그 단언을 지키는 시험이 없어서 exit 0 가드였어요(`GIT_DIR` 실험은 시험이 아니에요). 그래서 판정을 `codex_entry_verdict` 로 빼고 시험 둘을 더했어요. RED(`e0a16d3eb0`): `a_failed_git_run_fails_the_codex_check` 단언 실패 exit 101; GREEN(`6a113dfe20`): 9 통과; 뮤턴트(`if !success` 를 `if false && !success`) → exit 101.
- AC-4: met — `claude_settings_scale_keys` 13 통과(RED `691442615b` 에서 다섯 시험이 단언으로 실패, GREEN `e7b667ea1d`), `claude::tests::an_unknown_permissions_key_is_shown_without_a_split_token` 통과, 이웃 `claude_settings_cli` 18·`project_scale_plan_lib` 11·`project_scale_detect_lib` 6 통과. 뮤턴트 넷 모두 exit 101: N1 권한 키를 날것으로 찍기(2 시험), N2 날것 글만 판정(6 시험), N3 U+2800 을 목록에서 뺌(1 시험), N5 지웠다는 표시 끔(1 시험).
- AC-5: met — `moai spec check SPEC-20261122-project-scale-settings` 오류 0 경고 0; D-6 에서 D-8 이 갈래마다 넣음과 뺌을 이유와 함께 적어요.
- AC-6: met — 병합 `a236de79ec` 의 `process_scan_snapshot_cli` 5 통과(cwd 만 묻기, 주입 표, 멈춘 lsof, 삭제 직전 재확인, 시간 한도 거절). 가드 뮤턴트 MB(시간 한도 kill 분기가 `return None` 대신 `break`) → 세 시험 실패 exit 101.
- AC-7: met — `hygiene_apply_cli` 11, `process_scan_cost_cli` 2, `hygiene_cli` 3, `process_table_default` 5, `home_state_plan` 10, `worktree_gc` 14, `worktree_gc_apply` 15, `worktree_gc_hygiene` 9, lib `worktree_gc` 22 통과. 가드 뮤턴트 MH(hygiene 재확인이 `process_cwds_fresh` 대신 캐시를 읽음) → `the_hygiene_move_recheck_runs_its_own_scan` 실패 exit 101; MR(`real` 을 파일로 읽음) → 시험 여섯 실패 exit 101.
- 조각 D 의 병합 판단: 충돌은 `src/worktree_gc.rs` 한 곳이었어요. 통합 브랜치의 `injected_empty_table`(`empty` 만)은 조각 2 의 `injected_table`(`empty` 와 파일)이 대신하고, 시험 도우미가 진짜 표를 쓰라고 넣는 값 `real` 은 주입 없음으로 읽어요(안 읽으면 `home_state_plan` 시험 셋이 `process_check_unavailable` 로 깨져요, 합치기 전에 세 시험이 그렇게 실패했어요). 테스트 환경 변수는 `MOAI_TEST_*` 로 읽고, 새 시험 파일의 상태 폴더 글자 수를 `allow.tsv` 에 올렸어요(`state_folder_literal_scan` 8 통과).
- 위험 경로(버리는 폴더, 진짜 `lsof`): 쉬는 폴더 → 막힘 없음; 안에서 일하는 진짜 자식 프로세스 → `in_use`·`file_open_in_use`; 안의 파일을 연 자식 → `file_open_in_use`; 몇 줄 출력하고 멈추는 가짜 `lsof` 와 500 ms 한도 → `process_check_unavailable`·`open_file_check_unavailable`; 자식이 끝난 뒤 → 막힘 없음; 남은 프로세스 0(`target/cleanup/probe.log`).
- 기준 1(C-239 하위 `.gitignore`)은 C-174 착지를 기다려서 `blocked`, 기준 4(C-331)는 RC1 뒤라 `not_run` 이에요(Deferred).
- 판 2 (2026-11-24, 초안): 조각 A 에서 D 는 `NeXT` `89e8149c87` 에 착지했고 카드 기준 2, 3, 5, 6 은 노트 8171 을 증거로 met 로 적었어요. 조각 E 는 카드 기준 7, 8, 9, 10 이에요. 조각 E 는 아직 시작하지 않았어요(승인 전). 기준 1 은 D-5 의 선택을 기다려요. 위의 AC-1 에서 AC-7 의 Result 는 판 1 의 것이고 조각 E 의 AC-8 에서 AC-11 은 `not_run` 이에요.
- 돌리지 않은 것: 조각 D 병합본의 독립 tier 2 감사, Linux 와 Windows 실행, 전체 `cargo test`, 병합 뒤 시간 전후 측정.
