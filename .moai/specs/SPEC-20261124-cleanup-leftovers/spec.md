---
id: SPEC-20261124-cleanup-leftovers
title: "Six small leftovers of the 2026-11-24 card cleanup land as separate slices of one SPEC"
status: approved
revision: 3
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8263 (2026-11-24T06:01:28Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8262 에 답함): 판 3 승인 선택지 A, head db1f0d7299; 승인 앞머리는 이 노트로 적고 최종 head 의 맨 check affected 영수증으로 READY; 재감사는 첫 감사자 몫; 범위를 열 개로 줄여 take 한 판단 좋아요; allow.tsv 재발은 새 카드 없이 이 카드의 열린 기준으로 더해요; 레인은 그 노트를 인용만 해요"
  digest: "sha256:f3aab8f9322af81ebc78c0583a7d87913e0a102228d808192a54d3ee338412fa"
---

# Six small leftovers of the 2026-11-24 card cleanup land as separate slices of one SPEC

## Intent

사용자가 2026-11-24 에 열린 카드를 정리하며 개수를 줄이라고 정했어요. 닫힌 카드 여섯 장에 작은 조각이 하나씩 남았고, 새 카드 여섯 대신 한 카드의 기준 여섯으로 묶였어요. 이 SPEC 은 그 여섯 조각 가운데 지금 할 수 있는 네 조각(기준 6, 3, 5, 2)을 정하고, 지금 할 수 없는 두 조각(기준 1, 4)이 왜 기다리는지 적어요. 조각마다 따로 커밋하고 따로 감사받을 수 있어요.

원본 카드의 마지막 노트를 기준마다 읽고 현재 통합 브랜치(`NeXT`, `546cd451a2`)와 맞춰 봤어요(Evidence).

판 2 (2026-11-24): 조각 A 에서 D 는 `NeXT` `89e8149c87` 에 착지했어요(카드 노트 8171, 감사 노트 8153). 그 뒤 카드에 기준 7 에서 10 이 더해졌고(노트 8098 과 감사 비차단, 착지 중 시험 하나의 깨짐), 판 1 의 범위에는 그 파일들이 없어요. 판 2 는 조각 E 를 더해 그 네 기준을 한 조각으로 묶어요. 조각 A 에서 D 의 요구사항, 인수 기준, 결정은 그대로예요.

판 3 (2026-11-24): 조각 E 의 감사(노트 8248, head `eda95c1273`)가 FAIL 이었어요. 코드 고침은 맞고, 시험 둘과 글 둘이 모자라요. (1) `home_install_apply` 의 이동 직전 읽기를 지키는 시험이 없었어요(B1): mixed HOME 에서는 계획 단계가 캐시를 채우고 나서 이동에 닿아요. (2) 새로 읽기의 열린 파일 절반을 지키는 시험이 없었어요(B2). (3) Scope 의 조각 E 줄이 괄호 글에서 끊겨서 `moai spec show` 의 `scope.write` 에 여섯 경로가 빠졌어요(N1). (4) `audit_slot.md` 와 audit-prep SPEC 의 "첫 쓰기 전에 거부"는 `ensure` 앞의 첫 확인에만 맞아요(N3). 판 3 은 REQ-14 와 AC-10 의 글, Scope 의 줄 모양, 줄 임대로 고친 두 파일의 이름을 바로잡아요. 요구사항 REQ-15 의 뜻과 코드 고침은 그대로예요.

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
- REQ-14: The Scope and the Result of `SPEC-20261124-audit-prep` and the page `docs/design-review/appendix/operations/audit_slot.md` shall say that `audit-slot prep` refuses a slot whose own `target/` is a link or a file before it writes the holder record or anything through the link, and that the refusal is checked twice (before `ensure`, where nothing was changed, and after it, where the message names what `ensure` did) · failure: none — the Scope says "no change of behavior" and is false for that refusal.
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
- AC-10 (REQ-15): for each of the three apply paths a stub `lsof` whose first scan is empty and whose later scans show a process inside the entry keeps the entry in place (the install path with a mixed HOME whose only Go name is `state/active-sessions.json`, where the install's own plan fills the cache before the move); a later scan that keeps every working folder outside the entry and names an open file inside it keeps the entry in place too; the plan reads keep the cache; each guard mutant (an apply path back on the cached read, `open_file_paths_fresh` back on `open_file_paths`) fails its test · verify: `cargo test --locked --offline --test process_scan_snapshot_cli`
- AC-11 (REQ-16): the step test passes with a stand-in `python3` first on `PATH` that writes a cache folder into HOME unless the bytecode switch is set; the guard mutant (the switch removed) fails with the HOME guard · verify: `cargo test --locked --offline --test audit_gate_cli the_workflow_step_posts_the_status_the_gate_decides`

## Scope

- write: `.moai/specs/SPEC-20261124-cleanup-leftovers/**`, `templates/claude/output-styles/moai.md`, `moai-easy.md`, `moai-learn.md`, `tests/style_banner_scan.rs`
- write: `docs/support-matrix.md`, `docs/host-support-matrix.md`, `docs/product-naming-inventory.md`, `tests/retired_host_words.rs`, `.moai/specs/SPEC-20261122-remove-codex-shipped-text/spec.md`
- write: `src/claude.rs` (the value screens only, new lines), `tests/claude_settings_scale_keys.rs`, `.moai/specs/SPEC-20261122-project-scale-settings/spec.md`
- write: `src/worktree_gc.rs`, `src/hygiene.rs`, `tests/process_scan_snapshot_cli.rs`, `tests/process_scan_cost_cli.rs`, `tests/hygiene_apply_cli.rs`, `.moai/specs/SPEC-20261121-process-scan-snapshot/**`
- write: `skills/moai/workflows/run.md`, `CLAUDE.md`, `docs/leader-rules.md`
- write: `.moai/specs/SPEC-20261124-audit-prep/spec.md`, `docs/design-review/appendix/operations/audit_slot.md`
- write: `src/home_state.rs`, `src/home_retire.rs`, `src/home_state_retire.rs`, `src/home_install_apply.rs`, `tests/process_scan_snapshot_cli.rs`
- write: `tests/audit_gate_cli.rs`, `tests/fixtures/state_folder_literals/allow.tsv`, `skills/moai/references/source-manifest.json`
Limits of slice E (a `write:` tag takes every backticked path after it; a tag such as `write (slice E):` is no tag and takes the first path only, so the slice E lines above carry no such words): `run.md` the one sentence; `CLAUDE.md` the recurring-problems bullet only, tier 2 floor; `docs/leader-rules.md` item 4 only; the audit-prep SPEC its Scope and Result text only; `src/home_state.rs` one new `InUse` constructor in new lines; the three apply files one call each; `allow.tsv` and `source-manifest.json` one line each, by `card line-lease`.
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
| 8 | C-364 notes 8068, 8103 (R1); `docs/design-review/appendix/operations/audit_slot.md:24`; `SPEC-20261124-audit-prep` Scope | the repair made `audit-slot prep` refuse a link or file `target/` before it writes the holder record or anything through the link (checked twice, before and after `ensure`); the Scope still says "no change of behavior"; the page lists the refusal only for the lane's `target` |
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
- 판 2 (2026-11-24): 조각 A 에서 D 는 `NeXT` `89e8149c87` 에 착지했고 카드 기준 2, 3, 5, 6 은 노트 8171 을 증거로 met 로 적었어요. 조각 E 는 카드 기준 7, 8, 9, 10 이고 리더가 승인했어요(노트 8183). 위의 AC-1 에서 AC-7 의 Result 는 판 1 의 것이에요. 기준 1 은 D-5 대로 C-174 의 `21b77ec6dc` 가 `NeXT` 에 들어온 뒤 그 증거로 met 로 적어요(지금 `git merge-base --is-ancestor 21b77ec6dc NeXT` 는 exit 1).
- AC-8: 시험 통과 — 문서 셋을 고쳤어요(커밋 `a2ae88e724`): `run.md` 의 "repair card" 를 열린 후속 카드의 기준으로, `CLAUDE.md` 의 "원인에 열린 카드가 이미 있으면" 문장을 `…a HOLD or a user rule.` 뒤로, `leader-rules.md` 4번 Refill 의 긴 문장을 문장으로 나눠요. 시험이 읽는 문구는 그대로이고 `followup_per_spec_rules` 2, `recurring_problems_rule` 5, `repo_claude_md` 26 통과, `git diff --check` 0. `CLAUDE.md` 가 들어가서 tier 2 전체 감사 대상이에요. 같은 head 의 `moai check affected` 영수증은 READY 노트에 있어요.
- AC-9: met — `audit_slot.md` 에 자리 자신의 `target/` 이 링크나 파일이면 점유자 기록과 링크 안 쓰기 전에 거부한다는 줄을 더했고(코드 `real_target`, `--lane` 유무와 관계없음; 판 3 에서 확인이 두 번임을 바로잡았어요: `ensure` 앞은 아무것도 바꾸지 않고, 뒤는 `ensure` 가 한 일을 메시지에 대요), `SPEC-20261124-audit-prep` 의 Scope 와 Result 에 "동작 변화 없음"의 한 가지 예외를 적었어요(커밋 `cc3c92c477`, 글만 고침, D-6). `moai spec check SPEC-20261124-audit-prep` 오류 0 경고 0.
- AC-10: met (판 3 에서 셋째 시험과 열린 파일 시험을 더했어요) — `InUse::read_fresh()`(커밋 `d09e8bdf87`)를 `home_retire` apply, `home_state_retire` apply(첫 확인 때 한 번 읽어요), `home_install_apply` 의 이동 직전 재확인 셋에 썼어요. RED(`481e150ac8`): 새 시험 둘(`the_retire_move_recheck_runs_its_own_scan`, `the_archive_move_recheck_runs_its_own_scan`)이 이동이 일어나서 실패(exit 101); GREEN 뒤 `process_scan_snapshot_cli` 7, 이웃 `home_retire_cli` 24, `home_state_retire_cli` 21, `home_install_apply_cli` 41, `home_install_go_retire` 9, `home_state_plan` 10 통과. 가드 뮤턴트: retire 와 archive 를 옛 `read` 로 되돌리면 각 시험이 실패(exit 101). 판 2 의 첫 감사(노트 8248)가 `home_install_apply` 호출을 지키는 시험이 없다고 FAIL 로 냈고(B1), 새로 읽기의 열린 파일 절반도 지키는 시험이 없었어요(B2). 판 3 의 시험(커밋 `c1dd6db298`): `the_install_move_recheck_runs_its_own_scan` 은 mixed HOME(`.moai/moai-cli.json` 과 Go 표식이 공유 이름 `state/active-sessions.json` 하나뿐)에서 `home-install-plan` 과 `home-install-apply` 를 돌려요. 이 흐름에서는 설치의 계획이 `home_retire::plan` 으로 읽기를 채운 뒤에야 이동에 닿고, 스텁 `lsof` 의 첫 두 호출은 비고 그 뒤는 `receipts` 안의 프로세스를 보여요. head 는 `process_cwd` 로 건너뛰고 `receipts` 가 제자리에 남아요. `the_move_recheck_reads_the_open_files_again` 은 뒤 호출에서 작업 폴더는 항목 밖, 열린 파일(`cache/a.bin`)만 항목 안에 두는 스텁으로 `home-retire-apply` 를 돌려 `file_open` 으로 건너뛰는 것을 봐요. 가드 뮤턴트: `home_install_apply.rs` 의 `InUse::read_fresh()` 를 `InUse::read()` 로 되돌리면 첫 시험이 `moved`(receipts 가 옮겨짐)로 실패(exit 100); `home_state.rs` 의 `read_fresh` 가 `open_file_paths_fresh` 대신 `open_file_paths` 를 부르게 하면 둘째 시험만 실패(exit 100, 나머지 `process_scan_snapshot_cli` 8개 통과). 두 뮤턴트 모두 되돌려 `git status` 에 시험 파일만 남은 것을 읽었어요. 앞의 "시험이 못 지켜요" 글은 판 2 의 것이고 더는 맞지 않아요.
- 조각 E1 의 읽기만 한 곳: `worktree_move.rs:269` 와 `optimize.rs:876` 의 캐시 읽기는 둘 다 `plan` 함수 안이고, `worktree_move` 의 이동 직전 재확인은 이미 `busy_now`(새로 읽기, `:496`)예요. 바꾸지 않았어요.
- 위험 경로(버리는 HOME, 진짜 `lsof`): 래퍼가 한 명령의 세 번째 `lsof` 호출(계획을 다시 계산한 뒤 첫 호출) 때 항목 안에서 진짜 자식(`sleep`)을 띄워요. `home-retire-apply` 와 `home-archive-apply` 모두 `skipped:process_cwd`, 항목이 제자리에 남았고 프로브가 띄운 자식은 모두 끝났어요(`target/cleanup/probe-e1b.log`). 자식을 계획 뒤에 손으로 띄우면 apply 의 계획 재계산이 먼저 `stale or altered plan` 으로 거절해요(앞 방어선, `probe-e1.sh` 첫 실행).
- AC-11: met — RED(`728a54c505`): 빈 HOME 을 단계의 셸에 주고 `python3` 대역(변수가 없으면 `HOME/Library/Caches` 에 쓰는)을 `PATH` 맨 앞에 두자 `left in …/home: Library` 로 실패(exit 101). GREEN(`618b113c56`): 단계의 셸과 진짜 python 의 실행 파일을 묻는 도우미에 `PYTHONDONTWRITEBYTECODE=1`. `audit_gate_cli` 6/6 통과(`cargo test`), `PATH=/usr/bin:$PATH` 에 `cargo nextest`(HOME 가드 켜짐)도 6/6, `HOME-GUARD` 줄 없음. 처음 고침에서는 도우미가 가드된 HOME 에 캐시를 써서 가드가 걸렸고 같은 변수를 줘서 고쳤어요. 뮤턴트는 RED 의 상태(변수 없음)와 같아요. 워크플로 파일은 안 바꿨어요.
- 첫 영수증은 FAIL 이었어요: head `7a1f50ec6c` 의 `moai check affected` 가 시험 3906 중 4개 실패(`test=100`)를 냈고 넷 모두 이 조각의 변경이 원인이에요. 새 시험이 상태 폴더 글자를 `allow.tsv` 한도보다 많이 써서 `state_folder_literal_scan` 과 `check_affected::tests::the_repository_tests_pass_the_table_step_as_added_files` 가 실패(`process_scan_snapshot_cli.rs` 가 `.moai-cli` 6번·`.moai` 11번 대 한도 2·6)했고, `run.md` 를 고쳐서 소스 매니페스트의 sha256 이 낡아 `product_cli skill_package::source_manifest_covers_every_package_file` 과 `source_manifest_facts this_repository_manifest_is_what_the_tool_writes` 가 실패했어요. 줄 임대(`card line-lease`)로 `tests/fixtures/state_folder_literals/allow.tsv` 의 한 줄(`2 6` 을 `6 11` 로, 옛 배치를 일부러 만드는 이동 시험이라서)과 `skills/moai/references/source-manifest.json` 의 `workflows/run.md` 한 줄(새 sha256)만 고쳤어요. 이 두 파일은 위 Scope 에 적혀 있지 않아요(공유 목록 파일, 리더의 승인 노트 8183 의 "줄 임대로").
- 판 3 (2026-11-24): 조각 E 의 감사(노트 8248, head `eda95c1273`)가 FAIL 이었어요(B1, B2 가드 뮤턴트 exit 0). 고친 것: 시험 둘(위 AC-10), `state_folder_literal_scan` 의 기준선 한 줄(`tests/process_scan_snapshot_cli.rs` `.moai-cli` 6 에서 9, `.moai` 11 에서 19, 줄 임대; 새 시험의 글자 수와 같아요), Scope 의 조각 E 줄(태그가 `write (slice E):` 라서 파서가 첫 경로만 읽고 여섯 경로를 버렸어요; 지금은 `write:` 이고 `moai spec show` 의 `scope.write` 가 32개, 여섯 경로가 모두 있어요; 줄 임대로 고친 `allow.tsv`, `source-manifest.json` 도 이름을 올렸어요), REQ-14 와 `audit_slot.md`·audit-prep SPEC 의 "첫 쓰기 전" 글(N3: 첫 확인 `audit_slot.rs:572` 만 "아무것도 바꾸지 않음", 둘째 `:588` 은 `ensure` 가 한 일을 메시지에 대요). 코드(`read_fresh` 호출)는 바꾸지 않았어요. 승인 앞머리는 판 3 초안이라 지웠고, 리더 승인 뒤 같은 모양으로 적어요. audit-prep SPEC 의 글이 또 바뀌어서 그 digest 도 리더가 착지 뒤 노트로 새로 적어요(D-6).
- 돌리지 않은 것(조각 E): 독립 감사(CLAUDE.md 때문에 tier 2), Linux·Windows 실행(`reads_proc()` 은 Linux 에서 일찍 돌아가요), 전체 `cargo test`.
- 돌리지 않은 것: 조각 D 병합본의 독립 tier 2 감사, Linux 와 Windows 실행, 전체 `cargo test`, 병합 뒤 시간 전후 측정.
