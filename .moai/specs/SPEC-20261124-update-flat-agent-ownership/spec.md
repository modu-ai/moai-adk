---
id: SPEC-20261124-update-flat-agent-ownership
title: "moai update 가 평평한 에이전트 파일을 우리 것으로 알아보고, 진짜 사용자 수정은 고르게 하고, conflict 가 남으면 Partial 로 보여요"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8453 (2026-11-24T10:19:43Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8452 에 대한 답): revision 1 승인, head a34384ed05a565fb7a7e32dc81ac251913bf877a, D-6(유지를 기억하지 않음)과 D-7(conflict 만 있어도 질문)을 받아요; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요"
  digest: "sha256:8e6bbf894a92e1f88ee900a0d11c67f5c92c2b1fbccb0242aad51474e3ae9649"
---

# moai update 가 평평한 에이전트 파일을 우리 것으로 알아보고, 진짜 사용자 수정은 고르게 하고, conflict 가 남으면 Partial 로 보여요

## Intent

사용자가 빌드 4.2648.23 으로 `moai update` 를 돌렸을 때 agents 가 매번 `5 conflict` 였고, 적용해도 0 changed 로 Done 이 됐어요. 원인은 두 가지예요.

1. `src/agent_install.rs:328-331` 이 `~/.claude/agents/moai-<role>.md` (평평한 옛 위치)의 sha 를 이 도구의 기록(owned)과만 비교해요. 기록과 다르면 `kept_edited`, 기록이 없으면 `kept_foreign` 이고 `:343` 에서 conflict 가 돼요. 그런데 그 다섯 파일은 패키지 원본(`agents/<role>.md`)과 바이트가 같았어요. 원본은 frontmatter 가 `readonly: true` 뿐이라 렌더(`src/agent_defs.rs`)가 붙이는 `disallowedTools` 가 없고, 그 결과 읽기 전용 에이전트(plan, research, audit)가 Write/Edit 를 가진 채 돌았어요.
2. conflict 만 남으면 적용의 `writes` 가 비어 `unchanged` 가 나오고(`:497-503`), `src/update_tux.rs:2045` 의 `Some("unchanged") | None => {}` 는 줄을 찍지 않아요. 그래서 화면은 아무 말 없이 Done 으로 끝나요. 또 conflict 만 있는 계획은 아무것도 쓸 게 없다고 보아 질문 자체를 하지 않아서, 사용자가 고를 길이 없어요.

이 SPEC 은 `SPEC-20261121-base-agents-user-folder` 의 REQ-4 와 AC-2 의 conflict 부분을 대체해요 (리더 결정 2026-11-24, 노트 8447 선택 A: 옛 SPEC 은 승인 완료본이라 고치지 않고 새 SPEC 이 대체해요). 옛 SPEC 에 대체 포인터 한 줄을 넣는 일은 이 카드의 쓰기 범위 밖이라 후속 카드로 남겨요. 바이트가 같은 파일은 우리 것으로 알아보고, 진짜로 수정된 파일은 막다른 conflict 대신 사용자가 고르게 하고, conflict 가 남은 적용은 Partial 로 보여요.

읽는 쪽은 `moai update` 를 돌리는 사용자, 이 변경을 구현하는 레인, 이를 감사하는 세션이에요.

## Interview

- Q1: Does a new SPEC replace REQ-4 of the base-agents SPEC, or does that SPEC get a revision 2? -> A: a new SPEC (D-1); the old one is an approved, audited document with a Result, and the leader chose A on note 8447.
- Q2: Which bytes make a flat file ours? -> A: the record, the current render, or the package file of the same role in the package being applied (D-2); older version folders are left out.
- Q3: Where is the person asked, and what happens off a terminal? -> A: after the confirming yes and before the write, at a terminal only; the JSON envelope path keeps every conflict and answers `partial` (D-3, D-5).
- Q4: Where do the person's bytes go when MoAI's file is chosen? -> A: the receipt folder `before/<host>/legacy/`, so `agent_install::revert` restores them without a new backup folder (D-4).
- Q5: Is a kept file remembered, and is a conflict alone worth a question? -> A: not remembered, so the next update asks again and ends Partial (D-6); a conflict alone is asked about, or the person could never choose (D-7). Both accepted by the leader on note 8453.

## Non-goals

- 평평한 파일이 아닌 곳(중첩 `agents/moai/`, 프로젝트 `.claude/agents/moai/`)의 규칙 변경이에요.
- 예전 패키지 버전 폴더(`~/.moai/versions/*/agents`)의 원본까지 비교하는 일이에요.
- JSON 봉투 경로(파이프, 호스트 도구, 레인)에 고르기 입력을 더하는 일이에요. 지금 쓰는 곳이 없어서 후속으로 남겨요.
- 사용자가 "내 것 유지"를 고른 사실을 기억해서 다음 실행에서 묻지 않는 일이에요.
- 진짜 HOME 과 사용자의 `~/.claude` 를 읽거나 쓰는 일이에요: 모든 테스트는 임시 HOME 만 써요.
- 옛 SPEC 파일을 고치는 일이에요.

## Requirements

- REQ-1: When a flat base agent file holds the bytes of its role's current render or of its role's file in the package being applied, the agent plan shall list it as ours, to be removed after its nested copy is written, whatever the record says.
- REQ-2 (REQ-1): IF a flat base agent file matches neither its record nor the current render nor the package file of its role, THEN the agent plan shall list it as a conflict and leave the file and its nested path untouched.
- REQ-3: When the plan holds a flat-file conflict and the update runs at a terminal, the update shall ask once per conflicting file whether to keep mine, use MoAI's or show the difference, after the confirming yes and before the write. · failure: none — a prompt that cannot be read counts as no answer (REQ-7)
- REQ-4 (REQ-3): When the person picks show the difference, the update shall print the lines that differ between the person's file and MoAI's render and then ask the same question again. · failure: none — a display only reads and writes nothing
- REQ-5 (REQ-3): When the person picks use MoAI's, the apply shall keep the person's bytes in the receipt, write the nested copy and remove the flat file.
- REQ-6 (REQ-5): IF the receipt of an apply that used MoAI's file is reverted, THEN the revert shall put the person's flat file back with the kept bytes and remove the nested copy that the apply wrote.
- REQ-7: When the update runs without a terminal or the person picks keep mine or gives no answer, the apply shall leave the flat file and its nested path untouched. · failure: none — nothing is written
- REQ-8 (REQ-5): IF the apply is asked to use MoAI's file for a file that the fresh plan does not list as a flat-file conflict, THEN the apply shall refuse with nothing written.
- REQ-9: When a flat-file conflict is left after the apply, the agents result shall carry the state `partial` with the files kept, also when nothing else was written. · failure: none — a result only reports
- REQ-10 (REQ-9): When the agents result is `partial`, the end screen shall show an agents row that names the number kept and end on the Partly applied line instead of Done. · failure: none — a screen only draws
- REQ-11: When the plan holds a flat-file conflict and nothing else is to be written, the update shall ask the confirming question instead of ending on the all-up-to-date screen. · failure: none — a question writes nothing until it is answered

## Acceptance

- AC-1 (REQ-1): on a temporary HOME a flat file equal to the render of its role, and a flat file equal to the package file of its role while the record names other bytes or none, are both planned `legacy: remove` with no conflict, the apply moves them, and a second apply reports `unchanged` · verify: `cargo nextest run --locked --offline --test agents_flat_ownership a_flat_file_equal_to_the_render_or_the_package_file_is_ours`
- AC-2 (REQ-2): a flat file that differs from its record, the render and the package file (one with a record, one with none) stays a conflict, and an apply with no choice leaves it and writes no nested copy · verify: `cargo nextest run --locked --offline --test agents_flat_ownership a_flat_file_that_matches_nothing_stays_a_conflict`
- AC-3 (REQ-3): with scripted answers the prompt asks once per conflicting file and offers keep mine, use MoAI's and show the difference, and a file that is no conflict is never asked about · verify: `cargo nextest run --locked --offline --test agents_flat_ownership the_choice_asks_once_per_conflicting_file`
- AC-4 (REQ-4): the answer for the difference prints the differing lines of both sides and asks the same file again, and an answer after it decides the file · verify: `cargo nextest run --locked --offline --test agents_flat_ownership the_difference_is_shown_and_the_question_comes_again`
- AC-5 (REQ-5): the apply with a use-MoAI's choice leaves the nested copy with the render bytes, no flat file, the person's bytes under `before/claude/legacy/` of the receipt, and a record that names the nested bytes · verify: `cargo nextest run --locked --offline --test agents_flat_ownership use_moais_keeps_the_bytes_and_moves_the_file`
- AC-6 (REQ-6): `agent_install::revert` of that receipt puts the person's flat file back with its bytes and removes the nested copy · verify: `cargo nextest run --locked --offline --test agents_flat_ownership the_revert_restores_the_persons_file`
- AC-7 (REQ-7): an apply with no choice, with keep mine, with an empty answer and with end of input each leaves both paths byte for byte as they were · verify: `cargo nextest run --locked --offline --test agents_flat_ownership keep_and_no_answer_leave_both_paths_untouched`
- AC-8 (REQ-8): a choice naming a file that is no flat-file conflict (an unchanged one, one that is ours, an unknown name) is refused and nothing is written · verify: `cargo nextest run --locked --offline --test agents_flat_ownership a_choice_for_a_file_that_is_no_conflict_is_refused`
- AC-9 (REQ-9): an apply that leaves a conflict answers `partial` with the kept names, with other writes and with none, and an apply with no conflict still answers `applied` or `unchanged` with no `kept` key · verify: `cargo nextest run --locked --offline --test agents_flat_ownership a_left_conflict_makes_the_agents_result_partial`
- AC-10 (REQ-10): the end screen of a `partial` agents result has an agents row with the kept count and ends on Partly applied, in English and Korean, and the screen of an all-applied run is unchanged · verify: `cargo nextest run --locked --offline --test update_tux_golden the_applied_screen_shows_a_kept_agent_file_as_partial`
- AC-11 (REQ-11): the question predicate of the agents plan is true for a plan whose only item is a flat-file conflict and false for a plan with only `unchanged` files · verify: `cargo nextest run --locked --offline --test agents_flat_ownership a_conflict_alone_is_worth_a_question`
- AC-12 (REQ-1, REQ-2, REQ-5, REQ-7): the tests of `SPEC-20261121-base-agents-user-folder` still pass, with the edited-or-foreign case now kept only by the choice rule · verify: `cargo nextest run --locked --offline --test base_agents_nested --test agents_home_cli`
- AC-13 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11): the whole library and the binaries build and pass and the SPEC checks clean · verify: `moai check affected` and `moai spec check SPEC-20261124-update-flat-agent-ownership`

## Scope

- write: `.moai/specs/SPEC-20261124-update-flat-agent-ownership/spec.md`, `src/agent_install.rs`, `tests/agents_flat_ownership.rs`, `tests/base_agents_nested.rs` (only if a line must follow), `tests/update_tux_golden.rs`, `tests/fixtures/update_tux/`
- line lease (held by another card, one `moai card line-lease` edit each): `src/update.rs` (the question predicate at `agents_write_anything` and the confirmed agents part that calls `apply_for`), `src/update_tux.rs` (the `part` closure of the applied screen and the words of the kept row)
- read-only: `src/agent_defs.rs`, `src/progress.rs`, `src/ui_text.rs`, `src/update_cleanup.rs`, `.moai/specs/SPEC-20261121-base-agents-user-folder/spec.md`
- forbidden: the real HOME and every folder under it (the user's agents folder and both state folders), `CLAUDE.md`, `tools/repo_scans.json`, the old SPEC file, any card number in a file name, SPEC id, branch or commit message

## Plan

1. RED: `tests/agents_flat_ownership.rs` with the tests of AC-1 to AC-9 and AC-11 on a temporary HOME, and the screen test of AC-10 with its fixture; each fails on the current code for the right reason (a file equal to the render is a conflict; there is no choice; the result is `unchanged`).
2. `src/agent_install.rs`: read the package file bytes of each role from the agent source folder next to the render; the flat file is ours when its sha equals the record, the render sha or the package file sha (REQ-1); the plan shape and its JSON stay as they are, so fixtures that name no conflict keep their bytes. A new `apply_with_choices` takes the file names to use MoAI's file for and `apply_for` calls it with none, so no caller changes. A choice for a file that the fresh plan does not list as `kept_edited` or `kept_foreign` is refused before any write (REQ-8). The use-MoAI's path reuses the existing legacy removal and its `before/<host>/legacy/` receipt copy, so the revert already restores it (REQ-5, REQ-6).
3. `src/agent_install.rs`: `choose_flat_conflicts(plan, input, output, lang)` asks per conflicting file with injected reader and writer (REQ-3, REQ-4, REQ-7), and `asks_anything(plan)` is true when any file is a flat-file conflict or `writes_anything` (REQ-11). The result gets `kept` and the state `partial` when a conflict is left (REQ-9); the key is left out when empty.
4. `src/update.rs` by line lease: `agents_write_anything` calls `asks_anything`; the confirmed agents part asks `choose_flat_conflicts` on stdin and stderr only when `update_tux::human()` and `can_ask()` hold and passes the answer to `apply_with_choices`. `src/update_tux.rs` by line lease: a `partial` arm of the `part` closure draws the agents row with the kept count and marks the run partial (REQ-10).
5. Guard mutants, run once each and restored: the render equality removed, the package-file equality removed, the refusal of an unplanned choice removed, the `kept` state removed, the keep path writing the nested copy, the revert not restoring the flat file.
6. SYNC: the Result with each AC met, not met or not_run; then the READY note with the `moai check affected` receipt. A run of the dangerous path in a throwaway HOME (a flat file equal to the render, an edited one, a linked agents folder) goes into the READY, not only green tests.

## Risks

- A flat file that the person edited must never be removed or overwritten without the person's choice. The only path that removes it is the use-MoAI's choice, which keeps the bytes in the receipt first; REQ-8 and AC-8 pin that nothing else can ask for it.
- Equality with the package file or the render is decided by digest on the bytes only; a symbolic link or a folder at a flat path is already refused by `current_file` and `host_folders_plain`, and the tests keep a link case.
- A kept file makes every later `moai update` ask again and end Partial, because the choice is not remembered (D-6). That is noisy but safe; remembering a kept choice is a follow-up if people complain.
- The question is read from the terminal in the human screen path only; the JSON envelope path never asks, so a host or lane run keeps every conflict and reports `partial`.
- `src/update.rs` and `src/update_tux.rs` are held by another card; if a line lease is refused the work on those two files waits and the refusal is written on the card.
- The wiring in `src/update.rs` is not reached by a test that runs the real installed package; AC-11 pins the predicate and AC-5 the apply, and the wiring is read by the audit and run in a throwaway HOME for the READY.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A new SPEC replaces REQ-4 and the conflict half of AC-2 of the base-agents SPEC | 리더 결정 2026-11-24, 노트 8447 선택 A (사용자 위임 2026-11-23T13:57Z 안); 옛 SPEC 은 승인 완료본이라 고치지 않아요 | 옛 SPEC 을 revision 2 로 올리는 새 SPEC 으로 바꿔요 |
| D-2 | Only the render and the package file of the package being applied count as ours | 카드 기준 1번의 범위; 예전 버전 폴더를 더 읽는 일은 규칙 20 에 따라 뺐어요 | 예전 버전 폴더 원본을 같은 비교에 더해요 |
| D-3 | The question comes after the confirming yes and before the write, at a terminal only | 질문 한 번이면 화면이 단순하고, 쓰기 전에 고르게 해요; 봉투 경로는 아무것도 묻지 않아요 | 확인 앞으로 옮겨요 |
| D-4 | The person's bytes are kept in the receipt folder `before/<host>/legacy/`, no new backup folder | 이미 있는 영수증과 `agent_install::revert` 가 읽는 자리라서 새 경로와 새 되돌리기가 필요 없어요 | `~/.moai/backups/` 아래 새 폴더로 바꿔요 |
| D-5 | The JSON envelope path has no choice input; it keeps every conflict and answers `partial` | 지금 쓰는 호스트나 레인이 없어요 (규칙 20) | 봉투에 고르기 입력을 더하는 후속 SPEC |
| D-6 | A kept file is not remembered: the next update asks again and ends Partial | 기억하는 기록은 새 상태 파일과 규칙이 필요해요; 카드 기준 3번이 conflict 가 남으면 Partial 이라고 정했어요 | 유지 선택을 기록하는 후속 SPEC |
| D-7 | A conflict alone is worth a question (REQ-11) | 질문이 없으면 사용자가 고를 길이 없어서 카드 기준 2번이 참이 될 수 없어요 | `asks_anything` 을 `writes_anything` 으로 되돌려요 |

## Result

Measured 2026-11-24 (UTC) on macOS in the worktree `.claude/worktrees/update-flat-agents`, base `NeXT` `974267e03d` (the checks did not follow the tip). Commits: draft `a34384ed05`, approval `bbdfd28286` (leader note 8453; the lane recorded no approval of its own), RED `df681517cf`, library code `c3296b92ff`, three added guard tests `64b1061643`, `78b39de496`, `5ea79c70df`, screen row `90f810b744`, question wiring `910893c34d`, rustfmt `5cf587e851`, scan table `46ba8dbc58`. `src/update.rs` and `src/update_tux.rs` were edited under a line lease after C-160 gave them back. The receipt is `moai check affected 974267e03d..HEAD` on head `46ba8dbc586b`: `targets=175 scans=listed:151 diff_check=0 fmt=0 clippy=0 test=0 result=PASS`, 3972 tests passed, log folder `target/check-affected-46ba8dbc586b/`. Two earlier runs failed and were repaired: `cargo fmt` on `5ea79c70df`, and three scan-table tests on `5cf587e851`.

- AC-1 met: `a_flat_file_equal_to_the_render_or_the_package_file_is_ours` fails on the RED head and passes on `c3296b92ff`; both kinds are planned `legacy: remove`, the apply moves them and a second apply is `unchanged`.
- AC-2 met: `a_flat_file_that_matches_nothing_stays_a_conflict` passes, with and without a record; no nested copy is written.
- AC-3 met in part: `the_choice_asks_once_per_conflicting_file` passes on the library function with scripted input. The wiring in `src/update.rs` (a terminal run asks, any other run keeps every file) has no test that starts the real `moai update` against a package, so the live prompt is `not_run`.
- AC-4 met: `the_difference_is_shown_and_the_question_comes_again`, and `the_difference_shows_control_characters_as_text` for the escape handling.
- AC-5 met: `use_moais_keeps_the_bytes_and_moves_the_file`.
- AC-6 met: `the_revert_restores_the_persons_file` and `the_revert_of_a_choice_that_wrote_no_nested_copy_keeps_the_file_the_persons`; after a revert the file is a conflict again, not ours.
- AC-7 met: `keep_and_no_answer_leave_both_paths_untouched`.
- AC-8 met: `a_choice_for_a_file_that_is_no_conflict_is_refused`; `a_choice_never_overwrites_an_edited_nested_copy` and `a_link_at_the_flat_path_is_never_followed_or_removed` pin the two refusals around it.
- AC-9 met: `a_left_conflict_makes_the_agents_result_partial`.
- AC-10 met: `the_applied_screen_shows_a_kept_agent_file_as_partial` in English and Korean; the unchanged applied fixture still ends Done.
- AC-11 met: `a_conflict_alone_is_worth_a_question`.
- AC-12 met: `base_agents_nested` and `agents_home_cli` pass inside the 76-test run and inside the affected check.
- AC-13 met: the receipt above; `moai spec check SPEC-20261124-update-flat-agent-ownership` is run once more before READY.

Guard mutants (each by one exact replacement, confirmed to change the file, run in one heavy slot, restored from a byte copy; logs `target/c459/mut-*.log`): render equality removed, package-file equality removed, the unplanned-choice refusal removed, the partial state removed, a kept file written as chosen, the revert skipping the replace action, the revert recording the person's bytes as ours, a conflict alone not asked, an edited nested copy overwritten, the difference printing equal lines, the difference printing raw bytes, and the chosen flat file not removed each fail their named test with exit 100. The revert mutant first survived (exit 0) because the nested revert hid it; the added test kills it.

Not run: Windows and Linux (Mac fixtures only), a live `moai update` against a real package, and the JSON envelope path with a choice (D-5, no input there).

### 후속 (2026-11-27)

NeXT tip `8933f238cd` 위의 브랜치 `spec/update-flat-agent-followup-20261127` 에서 위 `not_run` 과 감사 후속 세 건을 닫았어요. 요구사항과 승인 본문은 건드리지 않았어요. 커밋은 RED `550d0200a5`, 코드 `7ac223c720`, 실제 update 테스트와 스캔 표 `e7774f2c7c` 예요. 모든 실행은 macOS, debug 빌드로 heavy 슬롯 안에서 했고 임시 HOME 만 썼어요.

- AC-3 (실제 연결): 위에서 `not_run` 이던 실제 `moai update` 연결을 `tests/update_flat_agents_tty.rs` 가 닫아요. 임시 HOME 에 이 저장소의 `skills/moai` 와 `agents` 를 바이너리 옆 버전 폴더로 복사하고, 터미널(의사 터미널) 실행은 묻고 사람의 답을 적용해요. 터미널이 아닌 실행은 묻지 않고 모든 conflict 를 지키며 `partial` 로 답해요. 테스트 네 개가 통과했어요.
- 막힌 호스트 이름 붙이기: 에이전트 폴더가 링크라서 호스트가 막히면 적용 결과가 `blocked` 배열(호스트와 이유)과 `reason` 을 내고, 다른 쓰기가 있었으면 `partial`, 없었으면 `blocked` 로 답해요. 링크를 통해서는 아무것도 쓰지 않아요(`an_apply_names_a_blocked_host_and_writes_nothing_through_the_link`, 링크를 건 실제 update 테스트 `a_blocked_agents_folder_ends_the_update_partly_applied`). 막힘이 없는 적용에는 `blocked` 키가 없어요. 끝 화면은 `Some(other)` 갈래로 `Partly applied` 를 보여요(`the_applied_screen_shows_a_blocked_agents_host_as_partial`, 영어와 한국어).
- 차이 줄 수: 차이는 80줄에서 멈추고 나머지 개수를 세어요(`the_difference_stops_at_eighty_lines_and_counts_the_rest`, 80줄과 81줄 경계).
- 외부 소유 파일: 기록과 render 와 패키지 파일 어느 것과도 다른 평평한 파일은 묻고 답으로 지키거나 옮겨요(`a_foreign_flat_file_is_asked_about_and_kept_or_moved_by_the_choice`).

가드 뮤턴트는 한 번에 하나씩 `Edit` 로 걸고 이름 붙인 테스트를 슬롯에서 돌린 뒤 `git restore` 로 되돌렸어요. 모두 테스트가 실패했어요(exit 100): (a) 막힌 호스트 이름 붙이기를 뺌, (b) 막힘이 있어도 상태를 그대로 둠, (c) 외부 소유 파일을 conflict 로 보지 않음, (d) 터미널에서도 묻지 않음(`a_terminal_run_asks_about_a_conflict_alone_and_use_moais_moves_the_file`), (e) 확인 질문 조건을 쓰기만 보게 되돌림(터미널 테스트 세 개 실패), (f) 끝 화면의 `Some(other)` 갈래에서 `bad` 를 뺌(골든 테스트).

범위 어긋남: 이 SPEC 의 Scope 는 `tools/repo_scans.json` 을 forbidden 으로 적었지만, 새 테스트 파일이 스캔 표(`reads-skills` 의 `test_targets` 와 `rules`)에 줄을 요구해서 그 파일을 줄 임대(`card line-lease`)로 두 줄만 고쳤어요. 승인 본문은 그대로 두었고, 이 어긋남은 리더가 READY 에서 판단해요.
