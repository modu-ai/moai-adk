---
id: SPEC-20261124-moai-project-interview
title: "/moai project 설문을 감지로 알 수 없는 것만 묻는 최소 설문으로 다시 정해요"
status: approved
revision: 4
created: 2026-11-24
approval: {by: leader, at: 2026-11-26, basis: "leader DELEGATED-APPROVAL note 9614 of the card that takes over the update.md half (answers lane note 9613, under the user's delegation of 2026-11-25 about 14:50Z, not a direct user pick), approving revision 4 at head 6ab400f830 and answering decision 1 (withdraw the hooks-conflict criterion) and decision 2 (the N7 follow-up goes to a later line, no new card); revisions 1 to 3 were approved by notes 8167, 8240 and 9400; recorded by the lane, not approved by it", digest: "sha256:98bf72be0b70fe1bc5ca6f9d49aeba26d6d27757092267ed35f368ad0585233f"}
---

# /moai project 설문을 감지로 알 수 없는 것만 묻는 최소 설문으로 다시 정해요

## Interview

사용자가 2026-11-24 `/moai project` 실행 중에 직접 정했어요(AskUserQuestion 선택). 기록은 카드 C-150 의 사용자 결정 노트 셋이고, 질문은 그 노트가 답하는 물음이에요.

- Q1: 완료 기준 질문의 선택지(`moai check affected` 통과 / `tools/local_check.sh` 통과)를 그대로 둘까요? → A: "이건 빼던지 하자. 사용자는 이거 뭔지도 모른다. 범용 오픈소스의 모든 사용자에게 맞지 않다. 모든 개발 언어를 지원해야 한다. 하드코딩 절대 금지."(노트 8151, 03:27:46Z)
- Q2: 설문 재정의 1차 — 질문 수, 운영 규칙, 방식·흐름·버전을 어떻게 할까요? → A: (1) 질문은 최소, 감지로 알 수 없는 것만 묻고 나머지는 감지값을 보여 주고 확인만. (2) 운영 규칙(보호 파일, 리뷰가 막을 실수, 묻지 않고 해도 되는 일, 완료 기준)은 안전한 범용 기본값을 적고 묻지 않음. (3) 개발 방식·git 작업 흐름·버전 방식은 감지값이 있으면 확인만, 없으면 범용 기본값(GitHub flow, SemVer, 테스트 먼저).(노트 8158, 03:43:57Z)
- Q3: 설문 재정의 2차 — 목적, 언어, 이 저장소의 이번 실행은 어떻게 할까요? → A: (4) 목적은 항상 사용자가 직접 적고 초안을 만들지 않음. (5) 언어는 대화·문서·커밋 메시지·지침 파일 네 가지를 각각 물음. (6) 이 저장소의 이번 실행은 바뀐 것만 적용(지침 파일 언어 en, 완료 기준을 범용 문장으로)하고, 계획을 보여 준 뒤 승인받아 적용.(노트 8159, 03:45:17Z)

이 SPEC 은 카드 기준 5~9(2026-11-24 에 더해진 것)를 맡아요. 기준 1·3·4(경계 조각: `project.md` 에서 옛 세대 이동 단계를 `update.md` 로 옮김, 카드·Issue 번호 스캔, 훅 충돌 선택지)는 보존된 가지 `spec/20261120-project-update-boundary` 의 몫이고 이 SPEC 이 바꾸지 않아요.

## Intent

지금 `/moai project` 는 감지로 알 수 있는 것까지 묻고, 사용자가 풀 수 없는 질문을 해요. 질문 은행 `skills/moai/references/internal/interview.md` 는 14행이고 그중 `protected`·`done`·`review_blocks`·`autonomy` 네 행은 늘 물어요. 이 저장소의 실행에서는 `done` 의 선택지가 이 저장소 개발 도구(`moai check affected`, `tools/local_check.sh`)로 만들어졌고(리더가 손으로 만든 선택지, 노트 8151), 사용자는 "뭔지도 모른다"고 했어요. 도구에 묶인 글은 Rust 가 아닌 프로젝트의 사용자에게 맞지 않아요. 또 `moai project-apply` 는 `project-plan` 이 출력한 `result` 에 `work_flow_remote`(원격 경고)가 붙어 있으면 `invalid typed JSON input` 으로 거절해요(카드 측정 2026-11-24, 계획 `sha256:aa74a169`; 읽은 곳: `src/ops/project_plan.rs:26` 가 키를 붙이고 `src/profile.rs:309` 의 `Plan` 이 `deny_unknown_fields` 라서 `src/ops/project_apply.rs` 의 입력 읽기가 거절).

바꾸는 것:

1. 묻는 것은 감지로 알 수 없는 다섯 가지뿐이에요: 목적(사용자가 직접 적음, 초안 없음)과 언어 넷(대화, 문서, 커밋 메시지, 지침 파일 — 각각).
2. 감지나 프로필이 아는 값(스택, 테스트·빌드 명령, 개발 방식, 호스트, git 작업 흐름, 버전 방식)은 한 번에 보여 주고 확인만 받아요. 아무것도 없으면 범용 기본값(GitHub flow, SemVer, 테스트 먼저)을 기본값이라고 표시해 같은 방식으로 확인받아요.
3. 운영 규칙 넷은 묻지 않고 범용 기본값을 `source: default` 결정으로 적어요. 이 기본값이 리뷰 지침(REVIEW.md)에 실제로 들어가도록 `src/agents.rs` 의 거름 조건을 `HOLD` 기준으로 바꿔요.
4. 스킬이 사용자에게 보여 주는 질문·선택지·예시에는 언어나 도구에 묶인 이름, MoAI 내부 명령이 없어요. 저장소 스캔 테스트가 지켜요.
5. `project-apply` 가 `project-plan` 의 `result` 를 출력된 그대로 받아요.

## Non-goals

- 감지기(`project-detect`)가 개발 방식이나 git 작업 흐름을 새로 읽게 하는 일. 이 SPEC 은 감지기가 이미 내는 값(`commands`, `versioning`, `languages`, `scale`)과 프로필 값을 쓰고, 읽을 값이 없으면 범용 기본값을 보여 줘요. 감지 확대는 후속 카드로 적어요.
- 경계 조각(기준 1·3·4)과 `/moai update` 의 질문.
- HOME 의 대화 언어를 이 설문에서 바꾸는 일: 사용자 범위(실제 HOME)는 `guarded` 설정이라 이 설문의 대화 언어 답은 프로젝트 프로필 `language.user_chat` 에만 적어요.
- 이 저장소 `.moai/project.yaml` 을 레인 워크트리에서 바꾸는 일(Plan 8단계는 코드가 들어간 뒤 따로 실행해요).
- 질문 은행의 `versioning`·`versioning_format`·`versioning_head`·`work_flow` 선택지 목록과 m-flow 설명 글: 이미 범용 문장이고, 이 SPEC 은 그 행들을 "묻기"에서 "확인"으로만 바꿔요.

## Requirements

- REQ-1: When the PROJECT route interviews, the question bank shall ask only the purpose and the four languages (conversation, documents, commit messages, instruction files) and no other row. · failure: none
- REQ-2: When the route asks the purpose, it shall ask one open question in the conversation and offer no draft written from the README, the manifests or the code. · failure: none
- REQ-3: When the route asks the languages, it shall ask each of the four as its own selection with the stored or suggested value first and marked recommended, including a value the user scope already holds. · failure: none
- REQ-4: When `moai language-setup --questions` runs, the set shall list four items, `user_chat`, `documents`, `commit_messages` and `instruction_files`, each with its current, suggested and default value. · failure: none
- REQ-5: When detection or the profile gives a value for the stack, the test and build commands, the development method, the hosts, the git work flow or the versioning scheme, the route shall show all of them in one confirmation, each with its source, a value the profile already holds counting as detected, and ask a follow-up only for the ones the user marks wrong.
- REQ-6 (REQ-5): IF an item has no detected value and no profile value, THEN the route shall show the generic default marked as a default (GitHub flow, SemVer, tests first) and confirm it in the same selection.
- REQ-7: The route shall record `protected`, `review_blocks`, `autonomy` and `done` each as a `decisions` entry with `"source": "default"` and the generic default sentence of the question bank, and ask none of them. · failure: none
- REQ-8 (REQ-7): When REVIEW.md is generated, it shall carry a `review_blocks` or `protected` decision whose source is `default` and whose value is not `HOLD`. · failure: none
- REQ-9: The shipped interview texts shall word every question, option and example without a language-specific tool or a MoAI-internal command. · failure: none
- REQ-10 (REQ-9): IF a scanned file holds a token of the scan's list (`cargo`, `rustup`, `npm`, `pipx`, `go install`, `moai check`, `local_check`, `check-guard`) outside the named exceptions, THEN the scan test shall fail and name the file and the line.
- REQ-11: When `moai project-apply` reads the `result` object of `moai project-plan` exactly as printed, including `work_flow_remote`, it shall apply the plan the digest names.
- REQ-12 (REQ-11): IF the input holds an unknown key other than `work_flow_remote`, THEN `moai project-apply` shall refuse it as invalid typed JSON input.
- REQ-13: When the project workflow meets earlier-generation state, it shall name `/moai update` in one boundary sentence. · failure: none
- REQ-14: IF the boundary sentence is missing from `project.md` or the matching boundary sentence is missing from `update.md`, THEN the shipped-text test shall fail.
- REQ-15: The three project-side files and `update.md` shall carry no card number and no Issue number and no dated user decision.
- REQ-16: The interview reference shall describe the next-flow work flow as `flows/next-flow` does and name no withdrawn host. · failure: none
- REQ-17: When the interview names the integration branch, it shall give the profile value first and then the remote default branch and then the current branch. · failure: none
- REQ-18: When step 4 builds the plan, its profile field list shall include `work_flow` and `language`. · failure: none
- REQ-19: The `method` row shall use the words the profile records. · failure: none
- REQ-25: The `hosts` row shall offer only a host the project still supports. · failure: none
- REQ-20: The texts shown to the user shall carry none of the internal words the scan lists (head number, a bare version-format token, a development-leftover sentence). · failure: none
- REQ-21: The project-side texts shall not contradict each other on editing the profile later, on who runs the worktree base write, or on the receipt step. · failure: none
- REQ-22: IF a scanned file holds `pytest`, `go test` or `cargo-nextest` outside the named exception, THEN the scan test shall fail and name the file and the line.
- REQ-23: IF the named install-table cell holds a MoAI-internal token, THEN the scan test shall fail and name the file and the line.
- REQ-24: When `project-plan` prints a `work_flow_remote` warning, the project workflow shall show that warning. · failure: none
- REQ-26: The project workflow shall carry no step that moves, backs up or merges earlier-generation state. · failure: none
- REQ-27: When `/moai update` meets earlier-generation state, the update workflow shall carry the go-state move steps that the project workflow no longer carries. · failure: none
- REQ-28: When the project workflow's Completion Check offers a follow-up for earlier-generation state, it shall offer the install steps only and name `/moai update` for the rest. · failure: none
- REQ-29: When `moai language-setup --help` runs, it shall describe the four items the command asks. · failure: none
- REQ-30: The scan table entry that reads the go-state text shall name `update.md`, where that text now lives. · failure: none

## Acceptance

- AC-1 (REQ-1): the question bank lists purpose and the four languages as the asked set (the purpose as plain text first, the four languages and the confirmation in at most two calls of three), names every other row as confirmed or defaulted, and holds no row for `protected`, `done`, `review_blocks` or `autonomy` as asked · verify: `cargo test --locked --offline --test project_interview_minimal the_bank_asks_only_the_purpose_and_four_languages`
- AC-2 (REQ-2): the bank and `project.md` say the purpose is typed by the user with no draft, and name no draft-from-README option · verify: `cargo test --locked --offline --test project_interview_minimal the_purpose_is_typed_with_no_draft`
- AC-3 (REQ-3, REQ-4): `moai language-setup --questions` in a temporary project and home lists four items; with a value stored in the home it is the suggested one, and the item is still listed · verify: `cargo test --locked --offline --test language_three_questions questions_list_four_items_with_their_values`
- AC-4 (REQ-5, REQ-6): the bank and `project.md` carry one confirmation that lists the six values with their sources, the generic defaults GitHub flow, SemVer and tests first, and "a profile value is shown, not asked" · verify: `cargo test --locked --offline --test project_interview_minimal detected_values_are_confirmed_and_defaults_are_generic`
- AC-5 (REQ-7): the bank holds one default sentence for each of the four operating rules, and `project.md` says none of them is asked · verify: `cargo test --locked --offline --test project_interview_minimal operating_rules_are_never_asked_and_have_defaults`
- AC-6 (REQ-8): a profile whose `review_blocks` and `protected` decisions have source `default` yields a REVIEW.md that carries both, and a `HOLD` value yields none · verify: `cargo test --locked --offline --test review_section_cli review_md_writes_default_sourced_decisions_but_not_hold`
- AC-7 (REQ-9, REQ-10): the scan finds no listed token in the question bank, `project.md`, `update.md` or `questions.md` outside the named exception, and fails on a copy that holds `cargo test` in an option, a token before a full stop (`the user runs cargo.`), `tools/local_check.sh`, or a token in any cell of the install table but its command cell · verify: `cargo test --locked --offline --test project_interview_minimal no_question_option_or_example_names_a_language_tool`
- AC-8 (REQ-11, REQ-12): in a temporary project with a remote, the `result` of `project-plan` piped to `project-apply` applies when it holds `work_flow_remote`, and a plan with another extra key is refused as `invalid typed JSON input` · verify: `cargo test --locked --offline --test project_apply_plan_result`
- AC-9 (REQ-7): on the repository's own profile after the change, `moai project-plan .` shows only two changed keys, `language.instruction_files` set to `en` and `done` set to the default sentence, and nothing is applied before the user picks the plan · verify: `moai project-plan .` with the profile of `moai project-detect .` on stdin, read by the leader (not_run in the lane)

- AC-10 (REQ-13, REQ-14): `project.md` carries the one boundary sentence that names `/moai update`, `update.md` carries the matching boundary sentence that names the project workflow, and the test pins both · verify: `cargo test --locked --offline --test project_interview_minimal boundary_sentence_is_pinned_in_both`
- AC-11 (REQ-15): the three project-side files and `update.md` hold no `C-<digits>`, no `#<3 to 5 digits>` and no `user decision <date>` · verify: `cargo test --locked --offline --test project_interview_minimal no_citation_in_project_side_files`
- AC-15 (REQ-26, REQ-27): `project.md` names none of `go-state-apply`, `go-settings-apply`, `worktree-move-plan`, `worktree-move-apply`, `go-state-revert` and `root-files`, `update.md` carries the go-state move steps, and the pins that read those steps now read `update.md` · verify: `cargo test --locked --offline --test project_interview_minimal the_move_steps_live_in_update_md` and `cargo test --locked --offline --test project_worktree_move_choice` and `cargo test --locked --offline --test update_route`
- AC-16 (REQ-28): the Completion Check of `project.md` offers the install steps only and names `/moai update` for the rest · verify: `cargo test --locked --offline --test project_preview_order` and `cargo test --locked --offline --test project_flow_closing`
- AC-17 (REQ-29): `moai language-setup --help` describes the four items · verify: `cargo test --locked --offline --test language_three_questions help_describes_the_four_items`
- AC-18 (REQ-30): the `include-str-go-state-classes-cli` entry of `tools/repo_scans.json` names `update.md` and the table test passes · verify: `cargo test --locked --offline --test repo_scans_table every_include_str_reader_of_a_repository_file_has_an_always_scan`
- AC-12 (REQ-16, REQ-17, REQ-18, REQ-19, REQ-20, REQ-21, REQ-24, REQ-25): the flow text, the integration-branch order, the step-4 field list, the `method` and `hosts` rows, the internal words and the three contradictions are pinned · verify: `cargo test --locked --offline --test project_interview_minimal flow_text_matches_next_flow` and `cargo test --locked --offline --test question_bank_hosts`
- AC-13 (REQ-22, REQ-23): the scan lists `pytest`, `go test` and `cargo-nextest` and reads the install-table cell for internal tokens · verify: `cargo test --locked --offline --test project_interview_minimal scan_tokens_cover_pytest_go_test_nextest`
- AC-14 (REQ-12): a duplicate key in a `project-apply` input is pinned as it behaves today (an altered duplicate is refused, an identical one applies) · verify: `cargo test --locked --offline --test project_apply_plan_result duplicate_key_altered_is_refused`

## Scope

- write: `.moai/specs/SPEC-20261124-moai-project-interview/**`
- write (after the approval, widened by the leader): `skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`, `skills/moai/references/internal/questions.md`, `skills/moai/references/source-manifest.json`, `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`, `tests/project_interview_minimal.rs` (new), `tests/project_apply_plan_result.rs` (new)
- write by line lease (pinned test files other cards also edit): `tests/language_three_questions.rs`, `tests/agent_defs_package.rs`, `tests/project_detect_cli.rs`, `tests/generic_flow_questions.rs`, `tests/question_bank_hosts.rs`; the plan names each pin it changes after running them
- write (revision 3, leader-approved scope of card C-150): `tests/project_apply_plan_result.rs`, and one line each in the pins that quote a removed citation: `tests/product_cli.rs`, `tests/project_applied_table.rs`, `tests/project_decide_model.rs`, `tests/project_flow_closing.rs`, `tests/update_route.rs`, `tests/next_flow_skill_scan.rs`
- revision 3 does not touch `skills/moai/workflows/update.md` (another card holds it), `src/**`, or `skills/moai/references/internal/versioning.md`
- write (revision 4, the card that takes over the update.md half): `skills/moai/workflows/update.md`, `skills/moai/workflows/project.md`, `src/ops/language_setup.rs`, the pins that read the moved steps (`tests/project_worktree_move_choice.rs`, `tests/go_state_classes_cli.rs`, `tests/update_route.rs`, `tests/project_preview_order.rs`, `tests/project_flow_closing.rs`, `tests/project_detect_cli.rs`), and by line lease `tools/repo_scans.json` (one prefix), `skills/moai/references/source-manifest.json` (the two sha256 values, through `moai check fix`), `src/help.rs` (the language-setup entry), `tests/language_three_questions.rs`
- revision 4 does not touch `src/generation.rs` or `tests/project_skills_generation.rs` (another card holds them; D-18)
- read-only: `skills/moai/workflows/update.md` (scanned; edited only if the scan finds a token), `skills/moai/references/internal/versioning.md` (scanned), `src/profile.rs`, `src/ops/project_plan.rs`, `src/project_detect.rs`
- forbidden: `CLAUDE.md`, the repository's `.moai/project.yaml` (Plan step 8 is the leader's), the real HOME, `skills/moai/SKILL.md`

## Plan

1. RED: write `tests/project_interview_minimal.rs` and `tests/project_apply_plan_result.rs` and the changed pins first; each fails against today's text and code (the apply test fails with `invalid typed JSON input`).
2. GREEN, code: `project-apply` takes `work_flow_remote` and refuses every other unknown key; `language-setup` lists four items; REVIEW.md writes `default`-sourced decisions that are not `HOLD`.
3. GREEN, text: rewrite the question bank (asked set, confirmed set, defaults table), `project.md` step 3 and its language bullet, and `questions.md` rule 2 (the purpose is the one open question); keep `source-manifest.json` in step.
4. Run the pinning tests the change touches (`agent_defs_package`, `language_three_questions`, `project_detect_cli`, `generic_flow_questions`, `question_bank_hosts`, `model_effort_questions`) and the scan; record each exit code.
5. Guard mutants, one line per new guard test: the scan with its token list emptied, the bank with `done` asked again, the apply with the extra-key refusal removed, REVIEW.md with the `default` filter restored; each must fail a test.
6. Dangerous path: not applicable (no delete, reset, process start or permission rule); the apply change is exercised in a throwaway project with a bare-repository remote.
7. `moai check affected` once at READY, unwrapped, on the head.
8. After the landing, the leader runs the repository's own profile update (AC-9): plan shown, approval, apply; this step is outside the lane worktree.

Revision 4 (the update.md half):

9. RED: add `boundary_sentence_is_pinned_in_both` and `the_move_steps_live_in_update_md` to `tests/project_interview_minimal.rs`, widen `no_citation_in_project_side_files` to `update.md`, add `help_describes_the_four_items`; each fails on this tree. Repoint the pins named in the Scope to `update.md` in the same commit so the old text and the new text are never both unpinned.
10. GREEN, text: move the go-state move steps (project step 2, the snapshot preamble, the backup waits, step 8's apply names) into `update.md` as one paragraph, leave a one-line stub in `project.md` that keeps the step numbers the tests pin, trim the Completion Check to the install steps, strip every citation from `update.md`; refresh `source-manifest.json` with `moai check fix <base>..<head>` on a committed head.
11. GREEN, code and table: `src/ops/language_setup.rs` and `src/help.rs` say four items; the `include-str-go-state-classes-cli` prefix becomes `update.md` (a change, not a growth, so tier 2).
12. Guard mutants, one line per new guard test: delete the `update.md` boundary sentence, restore one citation in `update.md`, restore "three questions" in the help, put the scan prefix back to `project.md`; each must fail a test.
13. Dangerous path: not applicable (text, scan table and help text; no delete, reset, process start or permission rule).

## Risks

- A default-sourced `review_blocks` and `protected` now reach REVIEW.md for every project that runs the interview: the user is told the sentences once, in the final table of `project.md`, and can change them later. The REVIEW.md rule "an unanswered axis is never written as a rule" stays for `HOLD`.
- Four language items ask the conversation language again for a user who set it at install. The answer is stored in the profile only; a mismatch with the home value is shown, not written to the home.
- Pinned test files are shared with other cards; a lease on them can delay the pin edits (the lane then writes `LANE-BLOCKED kind=claim`, as the card notes record for earlier slices).
- Step 6a of `project.md` lists language-server install routes in a table (`rustup component add rust-analyzer`, `pipx install pyright`, `npm install …`); they are agent commands taken from the official plugin READMEs (core rule 25), not a question, option or example, and D-6 names them as the scan's one exception.
- The detector does not read the development method or the git work flow today; the confirmation then shows the default as a default. A project whose real flow differs corrects it in the one follow-up.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The purpose is the one open question: asked first as plain text in the conversation, typed by the user, no draft and no option from the README or manifests. `questions.md` rule 2 ("offer drafts") gets this one exception. | User decision, note 8159 (4). | Restore the drafts option. |
| D-2 | The four languages are each asked, the conversation language included; a stored value is the recommended option, not a skip. This replaces REQ-1 and REQ-5 (three items) of SPEC-20261119-language-three-questions and REQ-10's language clause with D-12 of SPEC-20261122-next-flow-skills-agents ("no question when the user scope holds it"). | User decision, note 8159 (5). | Restore the three items and the skip. |
| D-3 | The four operating rules are never asked and get these defaults, recorded with `source: default`: `protected` "secret and key files are never changed or committed"; `review_blocks` "a secret in code or logs; a change without a test; a change outside the task's scope"; `autonomy` "reads, tests and edits inside the task's scope run without asking; anything hard to undo, or that touches money, secrets or deletion, is asked first"; `done` "the project's own tests pass". | User decision, note 8158 (2); card criterion 8. | Restore the four rows as asked. |
| D-4 | REVIEW.md writes a `default`-sourced decision unless its value is `HOLD` (today `src/agents.rs:304` drops every `default` source). | Derived, not a user pick: without it D-3 changes no review guidance. Reversible in-repo; the approval of this SPEC covers it. | Put the `source != "default"` filter back. |
| D-5 | One confirmation lists the detected or profile values with their sources (stack, test and build commands, development method, hosts, git work flow, versioning); the user marks only the wrong ones. | User decision, note 8158 (1) and (3). | Ask each row again. |
| D-6 | The scan's one named exception is the command cell (the last cell of each row) of the language-server install table of `project.md` step 6a, the table whose header row is `\| Language server \| Install route \|`; every other cell and every other line, the rest of that step included, is scanned (revision 2 narrows it from the whole step-6a line, audit #150 N2). | Derived: criterion 6 forbids tool names in questions, options and examples; the table is an agent command list tied to core rule 25. Options: (a) exception by name (chosen), (b) move the routes into `project-detect` output (a code change, a later card). | Take option (b) or delete the table. |
| D-7 | The `hosts` row shows the host this session runs in for confirmation and drops the Codex CLI and OMP names. | Derived from the user decision of 2026-11-21 (Codex and OMP withdrawn) and note 8158 (1). | Restore the three-host row. |
| D-8 | Card criterion 5 (ask `done` only when detection found commands) is replaced by criterion 8 (`done` is never asked); the leader withdraws criterion 5 with that reason. | The two criteria conflict; note 8158 (2) decides the later one. | Reopen criterion 5 and drop REQ-7 for `done`. |
| D-9 | `project-apply` ignores `work_flow_remote` only and still refuses every other unknown key; the plan digest covers the same bytes as before. | Card criterion 9; refusing unknown keys protects the typed input. | Accept every key (not recommended). |
| D-10 | The repository's own run plans only the two changed keys and waits for the user's plan approval before it applies. | User decision, note 8159 (6). | Re-run the full interview. |
| D-11 | Revision 3 supersedes the revision-2 Non-goal lines that keep the versioning and `work_flow` option lists and the flow description unchanged, for REQ-16 to REQ-20 only. | The survey of 2026-11-25 (items B4, B5, B8, B9) found those lines wrong against `flows/next-flow` and `moai work-flow-plan`. Derived, reversible in-repo. | Restore the revision-2 lines. |
| D-12 | Revision 3 takes over the project half of criteria 1 and 3 of the draft `SPEC-20261120-project-update-boundary`: the boundary sentence and the citation scan. The removal of the step-2 body (the earlier-generation move steps) waits until `update.md` carries that guidance; card criterion 1 stays open until then. | `update.md` is held by another card and does not yet carry the worktree-move, "I closed other sessions" and root-file picks. Deleting first would drop them. | Remove the boundary sentence and its test. |
| D-13 | Splitting the combined line of `## Result` into one line per AC is not a revision: `## Result` is outside the approval digest. | `moai spec show` hashes requirements, acceptance and `forbidden:` only. | Re-join the lines. |
| D-14 | `moai language-setup --help` (N6), default-sourced rules in project skills (N7) and the `KNOWN_HOSTS` cleanup need `src/**` and are follow-ups of card C-150, not part of revision 3. | Outside the write scope. | none |
| D-16 | Revision 4 supersedes the wait in D-12: the earlier-generation move steps move from `project.md` to `update.md` now that the card that held `update.md` has landed. Card criterion 1 closes with this revision. | The leader moved criteria 1, 3, 4 and 10 of the earlier card here (its notes 9448 and 9450). Revisions 1 to 3 were approved by notes 8167, 8240 and 9400. | Put the step-2 body back into `project.md`. |
| D-17 | Card criterion 3 (the hooks conflict choices text for `update.md`) is withdrawn with a reason: the draft SPEC `SPEC-20261120-hooks-conflict-choices` covers only the withdrawn Codex and OMP hooks and none of its code is on the tip; for Claude the `hooks` conflict is reported with nothing written, so a choices text would promise what the product does not do. | Leader decision, approval note 9614 (delegated, not a direct user pick). | Reopen the criterion and write a SPEC with code. |
| D-18 | The N7 follow-up (default-sourced review rules in the generated project skills) needs `src/generation.rs` and `tests/project_skills_generation.rs`, which another active card holds; it waits for that card and is recorded as one line of `docs/later.md` by the leader, with no new card now. Card criterion 4 closes on N3, N5 and N6: N3 and N5 are already met on this tree (`scan_tokens_cover_pytest_go_test_nextest`, `duplicate_key_altered_is_refused`); N6 is REQ-29. | Scope held by another card; a lane never widens into a held path. Leader decision, approval note 9614. | Do N7 in a card after that card lands. |
| D-19 | The Completion Check keeps the install steps and names `/moai update` for backup and merge steps; the three move-apply operation names leave `project.md`. | Plan recommendation; the boundary sentence already sends earlier-generation state to `/moai update` (REQ-13). | Keep the names in `project.md`. |
| D-15 | The final table keeps its state words (`applied`, `default`, `HOLD`, `not applied: <reason>`) and gets one plain sentence that explains them; the worktree base sentence keeps the clause that the user adds `baseRef` themselves for a `conflict`. | Other SPEC pins the words; changing them is a leader decision, not a lane one (plan Q2, Q4). | Rename the words and their pins in a later revision. |

## Result

Revision 2 repairs audit #150 (head `548822bf0a`, verdict FAIL). Repair head `d5450b118b` on the same branch, base `e9c6b93400` (the NeXT tip the first READY named). Checks of 2026-11-24 on a Mac under load from other lanes; the same-head `moai check affected` receipt is in the READY note, not here.

- B1 (blocking, REQ-10): the scan in `tests/project_interview_minimal.rs` ended a token only at a space, a quote or a closing mark, so `the user runs cargo.` and `tools/local_check.sh` passed. The language-tool tokens (`cargo`, `rustup`, `npm`, `pipx`, `go install`) now also end at a full stop that closes a sentence (`Cargo.toml` stays unflagged); the internal tokens (`moai check`, `local_check`, `check-guard`) end at any boundary. The guard test plants each missed shape.
- N1: `questions.md` is in the scan list (`SCANNED`), pinned by `the_scan_reads_the_question_rules_too`.
- N2: the exception is the command cell of the install table of step 6a (D-6, revision 2). The commands moved out of the prose into the table; a token in another cell, in a row of another table or in the prose of the same step is scanned. `the_named_exception_is_one_install_table` pins that the header stands once and in `project.md` only.
- N11: the source label the user sees is `detected` (`interview.md` confirmation text, `project.md` step 3), and the bank rows say "detection lists" instead of naming `project-detect`. Commands the agent runs (`moai project-detect`) keep their name.
- N4: `apply_still_refuses_another_unknown_key` removes `work_flow_remote` from its input first, so the refusal can only come from the other key.
- N8, N9: AC-6 verify names `review_section_cli` (where the test lives); this Result is written.
- Checks, exit codes (repair head `d5450b118b`, unmutated): `cargo test --locked --offline --test project_interview_minimal --test project_apply_plan_result` under `moai heavy-slot-acquire`, exit 0: `project_apply_plan_result` 3 passed, `project_interview_minimal` 8 passed; `python3 -I tools/source_manifest.py check` ok. `moai spec check SPEC-20261124-moai-project-interview --project .` on the revision-2 text reports one error, `approval_digest` (the approval block still names the digest of revision 1; the lane does not record an approval). Logs: `target/c150-repair/`.
- Guard mutants (one at a time, file restored and `git diff --quiet` confirmed): (m1) `Some('.') => false` exit 101, `the_tool_scan_flags_a_planted_tool_and_spares_only_the_named_exception` failed; (m2) the internal-token boundary disabled exit 101, same test; (m3) the whole row of the exception table exempt exit 101, same test; (m4) `questions.md` out of `SCANNED` exit 101, `the_scan_reads_the_question_rules_too` failed.
- AC-1: met — `the_bank_asks_only_the_purpose_and_four_languages`, one of the four tests of `project_interview_minimal`, passed on the repair head.
- AC-2: met — `the_purpose_is_typed_with_no_draft`, one of the four tests of `project_interview_minimal`, passed on the repair head.
- AC-4: met — `detected_values_are_confirmed_and_defaults_are_generic`, one of the four tests of `project_interview_minimal`, passed on the repair head.
- AC-5: met — `operating_rules_are_never_asked_and_have_defaults`, one of the four tests of `project_interview_minimal`, passed on the repair head.
- AC-6: met — `review_section_cli review_md_writes_default_sourced_decisions_but_not_hold` was passed at the first head by the auditor (18/18); this repair does not touch it.
- AC-7: met — `no_question_option_or_example_names_a_language_tool` passed over the four files with the new rules; the guard test plants the missed shapes.
- AC-8: met — `project_apply_plan_result` 3 passed.
- AC-3: unchanged by this repair (`language_three_questions`); the first READY's receipt covers it.
- AC-9: not_run — the leader's step after landing.
- Gaps, kept as follow-ups from the audit: N3 (the token list does not cover `pytest` or `go test`; REQ-10 fixes the list), N5 (a duplicate key in a `project-apply` input is merged by the JSON reader; the plan digest catches a changed content, no pinning test), N6 (`moai language-setup --help` still says three questions; `src/ops/language_setup.rs` is outside the scope), N7 (project skills do not carry default-sourced rules; REQ-8 covers REVIEW.md only). Card criterion 2 says three language items while D-2 and REQ-4 say four; the leader reads four as the requirement (audit note).

Revision 3 (project text repair), base 71d819ce8a, head b3cbbc9184. Evidence below is the verification run of the implementation step, `target/zoe/run2.log` in the worktree (`moai heavy-slot-acquire` around `cargo test --locked --offline`, 25 test targets, exit 101 because of three failing targets). `skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md` and the two new test files were last written before that run; `tests/project_harness_label.rs` and `skills/moai/references/source-manifest.json` were changed after it, by `b3cbbc9184`. The red run at `dd291c323d` is recorded only in that commit message (nine of the new tests fail on purpose): no log of it is readable in `target/zoe/` (`run1.log` stops at `product_cli`, `run2.log` is the run after the rewrite), so it is not_run here. No run on `b3cbbc9184` itself is readable; the receipt of the READY note covers it.

- AC-10: not met (partly) — `boundary_sentence_is_pinned` passed in `run2.log`, so the one boundary sentence naming `/moai update` exists and is pinned. The step-2 body stays in `project.md` until `update.md` carries that guidance (D-12), so card criterion 1 stays open.
- AC-11: met — `no_citation_in_project_side_files` passed in `run2.log`.
- AC-12: met — in `run2.log` `project_interview_minimal` passed `flow_text_matches_next_flow`, `integration_branch_order_matches_work_flow_plan`, `step_four_lists_all_profile_fields`, `rows_use_profile_words`, `no_internal_word_in_user_shown_text` and `contradictions_are_gone`, and `question_bank_hosts` passed 1 of 1.
- AC-13: met — `scan_tokens_cover_pytest_go_test_nextest` and `internal_tokens_scanned_in_exempt_cell` passed in `run2.log`.
- AC-14: met — `duplicate_key_altered_is_refused` and `duplicate_key_identical_applies` passed in `run2.log` (`project_apply_plan_result` 5 passed). They pin today's behaviour (characterization); no refusal code was added and `src/**` is untouched.
- Receipt runs of `moai check affected --card` (logs under `target/check-affected-*/` in the worktree): the first run at `e082ce1dc6` stopped at clippy (`unnecessary_map_or` in the sentence-end scan of `project_interview_minimal`, fixed in `066118be9c`); the second run at `066118be9c` passed clippy and failed three tests, `check_affected` `the_repository_tests_pass_the_table_step_as_added_files` and `the_repository_tests_pass_the_scan_step_as_added_files` and `repo_scans_table` `every_file_reader_has_an_always_scan_or_an_exemption`, because the new tests spelled a state folder name and opened the SPEC, `src/ops/work_flow_plan.rs` and a tool script at run time without a `reads-` scan naming them. `7a4b7fc625` drops the SPEC Result line test (`spec check` and `bundle check` read those lines), the source anchor and the tool path literal; no registry file changed. After it `project_interview_minimal` (18 passed) and `repo_scans_table` (33 passed) passed in one slot run. The READY note carries the receipt of the final head.
- Other results of the implementation run `run2.log` (before those fixes): `project_interview_minimal` 18 passed 1 failed (the dropped Result line test); `product_cli` `source_manifest_covers_every_package_file` (stale `interview.md` hash) and `project_harness_label` `workflow_text_names_the_step_where_it_lists_it` failed, which `b3cbbc9184` fixed. The other targets of the run (`agent_defs_package`, `generic_flow_questions`, `next_flow_skill_scan`, `project_applied_table`, `project_apply_plan_result`, `project_decide_model`, `project_detect_cli`, `project_flow_closing`, `update_route` and the rest) passed. `language_three_questions` and `review_section_cli` were not part of that run (not_run).
- Gaps, revision 3:
  - Card criterion 1 (the step-2 body of `project.md` and the pin that `update.md` carries it) stays open until the owner card of `update.md` lands.
  - Card criterion 3 stays open for `update.md`: it still holds an Issue number and a card number, and the owner card of `update.md` rewrites it.
  - Card criterion 4 (the hooks conflict text in `update.md`) stays open.
  - Card criterion 10 N6 (`moai language-setup --help`, `src/ops/language_setup.rs`) and N7 (default-sourced rules in project skills, `src/generation.rs`) are follow-ups that need `src/**` (D-14), as is the `KNOWN_HOSTS` cleanup.
  - The Rust successor of the `skill_references` test was not found; the old Python test was not run.
  - The final-table state words and the `baseRef` clause are unchanged (D-15).
