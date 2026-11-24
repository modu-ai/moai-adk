---
id: SPEC-20261124-update-flat-agent-ownership
title: "moai update 가 평평한 에이전트 파일을 우리 것으로 알아보고, 진짜 사용자 수정은 고르게 하고, conflict 가 남으면 Partial 로 보여요"
status: draft
revision: 1
created: 2026-11-24
---

# moai update 가 평평한 에이전트 파일을 우리 것으로 알아보고, 진짜 사용자 수정은 고르게 하고, conflict 가 남으면 Partial 로 보여요

## Intent

사용자가 빌드 4.2648.23 으로 `moai update` 를 돌렸을 때 agents 가 매번 `5 conflict` 였고, 적용해도 0 changed 로 Done 이 됐어요. 원인은 두 가지예요.

1. `src/agent_install.rs:328-331` 이 `~/.claude/agents/moai-<role>.md` (평평한 옛 위치)의 sha 를 이 도구의 기록(owned)과만 비교해요. 기록과 다르면 `kept_edited`, 기록이 없으면 `kept_foreign` 이고 `:343` 에서 conflict 가 돼요. 그런데 그 다섯 파일은 패키지 원본(`agents/<role>.md`)과 바이트가 같았어요. 원본은 frontmatter 가 `readonly: true` 뿐이라 렌더(`src/agent_defs.rs`)가 붙이는 `disallowedTools` 가 없고, 그 결과 읽기 전용 에이전트(plan, research, audit)가 Write/Edit 를 가진 채 돌았어요.
2. conflict 만 남으면 적용의 `writes` 가 비어 `unchanged` 가 나오고(`:497-503`), `src/update_tux.rs:2045` 의 `Some("unchanged") | None => {}` 는 줄을 찍지 않아요. 그래서 화면은 아무 말 없이 Done 으로 끝나요. 또 conflict 만 있는 계획은 아무것도 쓸 게 없다고 보아 질문 자체를 하지 않아서, 사용자가 고를 길이 없어요.

이 SPEC 은 `SPEC-20261121-base-agents-user-folder` 의 REQ-4 와 AC-2 의 conflict 부분을 대체해요 (리더 결정 2026-11-24, 노트 8447 선택 A: 옛 SPEC 은 승인 완료본이라 고치지 않고 새 SPEC 이 대체해요). 옛 SPEC 에 대체 포인터 한 줄을 넣는 일은 이 카드의 쓰기 범위 밖이라 후속 카드로 남겨요. 바이트가 같은 파일은 우리 것으로 알아보고, 진짜로 수정된 파일은 막다른 conflict 대신 사용자가 고르게 하고, conflict 가 남은 적용은 Partial 로 보여요.

읽는 쪽은 `moai update` 를 돌리는 사용자, 이 변경을 구현하는 레인, 이를 감사하는 세션이에요.

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
