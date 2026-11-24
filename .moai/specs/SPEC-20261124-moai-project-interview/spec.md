---
id: SPEC-20261124-moai-project-interview
title: "/moai project 설문을 감지로 알 수 없는 것만 묻는 최소 설문으로 다시 정해요"
status: draft
revision: 1
created: 2026-11-24
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

## Acceptance

- AC-1 (REQ-1): the question bank lists purpose and the four languages as the asked set (the purpose as plain text first, the four languages and the confirmation in at most two calls of three), names every other row as confirmed or defaulted, and holds no row for `protected`, `done`, `review_blocks` or `autonomy` as asked · verify: `cargo test --locked --offline --test project_interview_minimal the_bank_asks_only_the_purpose_and_four_languages`
- AC-2 (REQ-2): the bank and `project.md` say the purpose is typed by the user with no draft, and name no draft-from-README option · verify: `cargo test --locked --offline --test project_interview_minimal the_purpose_is_typed_with_no_draft`
- AC-3 (REQ-3, REQ-4): `moai language-setup --questions` in a temporary project and home lists four items; with a value stored in the home it is the suggested one, and the item is still listed · verify: `cargo test --locked --offline --test language_three_questions questions_list_four_items_with_their_values`
- AC-4 (REQ-5, REQ-6): the bank and `project.md` carry one confirmation that lists the six values with their sources, the generic defaults GitHub flow, SemVer and tests first, and "a profile value is shown, not asked" · verify: `cargo test --locked --offline --test project_interview_minimal detected_values_are_confirmed_and_defaults_are_generic`
- AC-5 (REQ-7): the bank holds one default sentence for each of the four operating rules, and `project.md` says none of them is asked · verify: `cargo test --locked --offline --test project_interview_minimal operating_rules_are_never_asked_and_have_defaults`
- AC-6 (REQ-8): a profile whose `review_blocks` and `protected` decisions have source `default` yields a REVIEW.md that carries both, and a `HOLD` value yields none · verify: `cargo test --locked --offline --test agent_defs_package review_md_writes_default_sourced_decisions_but_not_hold`
- AC-7 (REQ-9, REQ-10): the scan finds no listed token in the question bank, `project.md` or `update.md` outside the named exceptions, and fails on a copy that holds `cargo test` in an option · verify: `cargo test --locked --offline --test project_interview_minimal no_question_option_or_example_names_a_language_tool`
- AC-8 (REQ-11, REQ-12): in a temporary project with a remote, the `result` of `project-plan` piped to `project-apply` applies when it holds `work_flow_remote`, and a plan with another extra key is refused as `invalid typed JSON input` · verify: `cargo test --locked --offline --test project_apply_plan_result`
- AC-9 (REQ-7): on the repository's own profile after the change, `moai project-plan .` shows only two changed keys, `language.instruction_files` set to `en` and `done` set to the default sentence, and nothing is applied before the user picks the plan · verify: `moai project-plan .` with the profile of `moai project-detect .` on stdin, read by the leader (not_run in the lane)

## Scope

- write: `.moai/specs/SPEC-20261124-moai-project-interview/**`
- write (after the approval, widened by the leader): `skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`, `skills/moai/references/internal/questions.md`, `skills/moai/references/source-manifest.json`, `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`, `tests/project_interview_minimal.rs` (new), `tests/project_apply_plan_result.rs` (new)
- write by line lease (pinned test files other cards also edit): `tests/language_three_questions.rs`, `tests/agent_defs_package.rs`, `tests/project_detect_cli.rs`, `tests/generic_flow_questions.rs`, `tests/question_bank_hosts.rs`; the plan names each pin it changes after running them
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

## Risks

- A default-sourced `review_blocks` and `protected` now reach REVIEW.md for every project that runs the interview: the user is told the sentences once, in the final table of `project.md`, and can change them later. The REVIEW.md rule "an unanswered axis is never written as a rule" stays for `HOLD`.
- Four language items ask the conversation language again for a user who set it at install. The answer is stored in the profile only; a mismatch with the home value is shown, not written to the home.
- Pinned test files are shared with other cards; a lease on them can delay the pin edits (the lane then writes `LANE-BLOCKED kind=claim`, as the card notes record for earlier slices).
- Step 6a of `project.md` lists language-server install routes (`rustup component add rust-analyzer`, `pipx install pyright`, `npm install …`); they are agent commands taken from the official plugin READMEs (core rule 25), not a question, option or example, and D-6 names them as the scan's one exception.
- The detector does not read the development method or the git work flow today; the confirmation then shows the default as a default. A project whose real flow differs corrects it in the one follow-up.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The purpose is the one open question: asked first as plain text in the conversation, typed by the user, no draft and no option from the README or manifests. `questions.md` rule 2 ("offer drafts") gets this one exception. | User decision, note 8159 (4). | Restore the drafts option. |
| D-2 | The four languages are each asked, the conversation language included; a stored value is the recommended option, not a skip. This replaces REQ-1 and REQ-5 (three items) of SPEC-20261119-language-three-questions and REQ-10's language clause with D-12 of SPEC-20261122-next-flow-skills-agents ("no question when the user scope holds it"). | User decision, note 8159 (5). | Restore the three items and the skip. |
| D-3 | The four operating rules are never asked and get these defaults, recorded with `source: default`: `protected` "secret and key files are never changed or committed"; `review_blocks` "a secret in code or logs; a change without a test; a change outside the task's scope"; `autonomy` "reads, tests and edits inside the task's scope run without asking; anything hard to undo, or that touches money, secrets or deletion, is asked first"; `done` "the project's own tests pass". | User decision, note 8158 (2); card criterion 8. | Restore the four rows as asked. |
| D-4 | REVIEW.md writes a `default`-sourced decision unless its value is `HOLD` (today `src/agents.rs:304` drops every `default` source). | Derived, not a user pick: without it D-3 changes no review guidance. Reversible in-repo; the approval of this SPEC covers it. | Put the `source != "default"` filter back. |
| D-5 | One confirmation lists the detected or profile values with their sources (stack, test and build commands, development method, hosts, git work flow, versioning); the user marks only the wrong ones. | User decision, note 8158 (1) and (3). | Ask each row again. |
| D-6 | The scan's one named exception is the language-server install table of `project.md` step 6a. | Derived: criterion 6 forbids tool names in questions, options and examples; the table is an agent command list tied to core rule 25. Options: (a) exception by name (chosen), (b) move the routes into `project-detect` output (a code change, a later card). | Take option (b) or delete the table. |
| D-7 | The `hosts` row shows the host this session runs in for confirmation and drops the Codex CLI and OMP names. | Derived from the user decision of 2026-11-21 (Codex and OMP withdrawn) and note 8158 (1). | Restore the three-host row. |
| D-8 | Card criterion 5 (ask `done` only when detection found commands) is replaced by criterion 8 (`done` is never asked); the leader withdraws criterion 5 with that reason. | The two criteria conflict; note 8158 (2) decides the later one. | Reopen criterion 5 and drop REQ-7 for `done`. |
| D-9 | `project-apply` ignores `work_flow_remote` only and still refuses every other unknown key; the plan digest covers the same bytes as before. | Card criterion 9; refusing unknown keys protects the typed input. | Accept every key (not recommended). |
| D-10 | The repository's own run plans only the two changed keys and waits for the user's plan approval before it applies. | User decision, note 8159 (6). | Re-run the full interview. |
