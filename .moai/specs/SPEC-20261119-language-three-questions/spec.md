---
id: SPEC-20261119-language-three-questions
title: "The language setup asks three questions: conversation, commit messages and documents; everything else is English"
status: draft
revision: 1
created: 2026-11-19
---

# The language setup asks three questions: conversation, commit messages and documents; everything else is English

## Intent

사용자 결정(2026-11-19, update·project 재설계). 지금 `moai language-setup` 은 언어를 **8항목**(대화·보고·Issue/PR 기록·커밋·에이전트 메시지·지침 파일·skill 지침·코드 주석)으로 묻고, 설치기가 이미 저장한 대화 언어도 다시 물어요. 코드가 실제로 읽는 항목은 셋(`user_chat`, `issue_pr_records`, `instruction_files`)뿐이고 나머지는 AGENTS.md 에 한 줄 적히기만 해요. 영어가 아닌 사용자의 AGENTS.md 가 영어면 대화 언어와 달라서 스킬이 잘 발동하지 않아요.

바꾸는 것:
1. **묻는 것은 셋**: 대화 언어(`user_chat`; AGENTS.md·REVIEW.md 도 이 언어), 커밋 메시지(`commit_messages`), 문서(`documents` — 새 항목: 보고서와 SPEC 문서). 설치기가 저장한 대화 언어는 새로 묻지 않고 **확인값**(출처를 달아)으로 보여요.
2. **나머지는 영어 고정**: 지침 파일(스킬·규칙), 에이전트 간 메시지, 코드 주석, Issue·PR 기록. AGENTS.md 관리 구간이 이를 한 문장으로 말해요.
3. **옛 프로필 그대로 읽기**: 옛 8항목 키는 계속 받아들이고, 명시돼 있으면 지금처럼 존중해요(`instruction_files` 가 AGENTS.md 언어를, `issue_pr_records` 가 GitHub 서식 언어를 정함). 없을 때의 기본이 달라져요: AGENTS.md 는 대화 언어, GitHub 서식은 영어.

## Non-goals

- The project interview text of the skill (`skills/moai/workflows/project.md`) that describes the three language questions: the change that rewrites that workflow writes it, and received the wording of this change by a note. This SPEC does not edit or test that file.
- 결정 모델 제거, 질문 은행 축소, `/moai project` 와 `/moai update` 의 경계, 설치기·셸 `moai update` 의 질문 변경 — 같은 재설계의 다른 조각이에요.
- 영어·한국어 밖의 언어로 AGENTS.md 를 쓰는 일: 그 언어의 관리 구간 문장이 없어서 지금처럼 요청 언어로 쓰고 경고해요.
- 스킬·규칙 파일의 내용 번역이나 한국어 고정 문구 정리 — 이 SPEC 은 그 파일들을 영어가 아니게 만들지 않아요.
- 코드 값 목록(`en` `ko` `ja` `zh`)과 설치기가 저장하는 `config/language.yaml` 형식.

## Requirements

- REQ-1: When `moai language-setup --questions` runs, the setup step shall list exactly three items, `user_chat`, `commit_messages` and `documents`, each with its current value, its suggested value and its default. · failure: none
- REQ-2: When the installer saved a conversation language and the profile has no `user_chat`, the setup step shall suggest that language for `user_chat` and name its source as the installer. · failure: none
- REQ-3 (REQ-2): IF the saved conversation language is missing, unreadable or a language the setup has no name for, THEN the setup step shall suggest the next source in order (the profile, the previous generation's backup, English) and name that source.
- REQ-4: When answers are given, the setup step shall return a profile whose `language` names the three items and keeps every other `language` key the current profile already had, without adding a key for an item it does not ask. · failure: none
- REQ-5: When the setup step runs on a terminal, it shall ask one prompt for each of the three items and no other. · failure: none
- REQ-6: When the profile sets `language`, the AGENTS.md managed section shall list the `documents` item as the language of reports and SPEC documents when the profile sets it, followed by one sentence that says instruction files, skills and rules, messages between agents, code comments, and issue and PR records are written in English unless a line above says otherwise. · failure: none
- REQ-7: When the profile's `user_chat` is `en` or `ko` and the profile has no `instruction_files`, the AGENTS.md managed section and the REVIEW.md managed section shall be written in that language. · failure: none
- REQ-8 (REQ-7): IF the profile's `user_chat` is another code, THEN the plan shall write both sections in the request language and add a warning that names the code.
- REQ-9: While the profile sets `instruction_files` to `en` or `ko`, the managed sections shall be written in that language. · failure: none
- REQ-10: When a profile carries the old eight `language` items, the plan and the apply shall run without a conflict, leaving the profile's own values in the profile. · failure: none
- REQ-11: When the profile sets `language` without `issue_pr_records`, the GitHub forms shall be written in English; a profile that sets one decides the form language, and a profile with no `language` keeps the Korean forms it always had. · failure: none
- REQ-13 (REQ-9): IF the profile's `instruction_files` is a code other than `en` or `ko`, THEN the plan shall write the managed sections in the `user_chat` language when that is `en` or `ko` (else the request language) and add a warning that names `instruction_files` and its code.

## Acceptance

- AC-1 (REQ-1): `--questions` lists the ids `user_chat`, `commit_messages`, `documents` in that order, each with `current`, `suggested` and `default`.
  verify: `cargo test --locked --offline --test language_three_questions questions_list_exactly_three_items_with_their_values`
- AC-2 (REQ-2, REQ-3): the installer's saved language is the suggested `user_chat` with source `installer`; a missing or unsupported value falls to the next source, named.
  verify: `cargo test --locked --offline --test language_three_questions the_installer_value_is_the_confirmed_conversation_default`
  verify: `cargo test --locked --offline --test language_three_questions a_bad_installer_value_falls_to_the_next_source`
- AC-3 (REQ-4): answers give a profile with the three items and an untouched old key; a fresh profile gets no key for the English-only items.
  verify: `cargo test --locked --offline --test language_three_questions answers_name_the_three_items_and_keep_old_keys`
- AC-4 (REQ-5): the terminal step reads three answer lines.
  verify: `cargo test --locked --offline --test language_three_questions the_terminal_step_asks_three_prompts`
- AC-5 (REQ-6): the managed section has the English-rest sentence and the `documents` line.
  verify: `cargo test --locked --offline --test language_three_questions the_managed_section_names_documents_and_the_english_rest`
- AC-6 (REQ-7, REQ-9): AGENTS.md and REVIEW.md follow `user_chat` for `en` and `ko`; an explicit `instruction_files` wins.
  verify: `cargo test --locked --offline --test language_three_questions agents_and_review_follow_the_conversation_language`
  verify: `cargo test --locked --offline --test language_three_questions an_explicit_instruction_files_still_wins`
- AC-7 (REQ-8): another conversation language writes in the request language with a warning naming the code.
  verify: `cargo test --locked --offline --test language_three_questions another_conversation_language_falls_back_with_a_warning`
- AC-8 (REQ-10): a profile with the old eight items plans and applies without a conflict and keeps its values.
  verify: `cargo test --locked --offline --test language_three_questions a_profile_with_the_old_eight_items_plans_and_applies_without_conflict`
- AC-12 (REQ-13): an unsupported `instruction_files` code hands over to `user_chat` (else the request language) and warns.
  verify: `cargo test --locked --offline --test language_three_questions an_unsupported_instruction_files_code_hands_over_to_the_conversation_language`
- AC-9 (REQ-11): with `language` set, form language is English without `issue_pr_records` and follows it when set; with no `language` it stays Korean.
  verify: `cargo test --locked --offline --test language_three_questions the_form_language_is_english_unless_issue_pr_records_is_set`
- AC-11: neighbours (targeted; the full run is cited from the PR's CI run).
  verify: `cargo test --locked --offline --test language_three_questions --test language_setting --test skeleton_language --test go_settings_defaults --test language_set_cli`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261119-language-three-questions/**`, `src/language.rs`, `src/agents.rs`, `src/ops/language_setup.rs`, `tests/language_three_questions.rs`, `tests/language_setting.rs`, `tests/skeleton_language.rs`, `tests/go_settings_defaults.rs`, `skills/moai/references/source-manifest.json`
- read-only: `src/github_templates.rs`, `src/go_settings.rs`, `src/profile.rs`, `src/hook_engine.rs`, the installer's `install/install.sh`
- forbidden: any write in the real HOME, a change to the code values or to the installer's saved file format, a new card, Issue or pull request number in the text this change adds

## Plan

1. SPEC, then RED tests (`tests/language_three_questions.rs`).
2. `src/language.rs`: asked items and legacy items, sources, answers, terminal step, rendered lines; `src/agents.rs`: the managed-section language.
3. Update the older tests that name the eight items; checks, mutants, a read-only audit by another agent.

## Risks

- A project whose profile sets only `user_chat: ko` gets an AGENTS.md in Korean (it was English) and English GitHub forms (they were Korean): both are the decided defaults, shown in the plan as a changed managed section.
- An older binary reading a profile that carries `documents` refuses it as an unknown item; the profile is read by the binary that wrote it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `documents` is a new item; an old `reports` value is a suggestion for it, not a copy | The user's decision names reports and SPEC documents as one language | Set `documents` in the profile |
| D-2 | Old keys stay valid and keep their meaning when set | Old profiles must read as before | none needed |
| D-3 | The English-only items get no key in a new profile | Fixed English needs no storage; one sentence says it | Add the key by hand |

## Result

`moai language-setup` 이 이제 세 가지(대화 `user_chat`, 커밋 메시지 `commit_messages`, 문서 `documents`)만 묻고, 나머지는 영어 한 문장으로 말해요. AGENTS.md·REVIEW.md 는 `instruction_files` 가 없으면 대화 언어(en·ko)로 써요. 아래 종료 코드는 모두 0 이에요(릴리즈 팁 `6a8fa9cacf` 를 병합한 가지).

- 새 시험 12개(`--test language_three_questions`): 처음 12개였고, 프로젝트 워크플로 글 시험은 요구에서 빠져 지웠어요(Non-goals). 독립 감사(FAIL, 차단 2건) 뒤 한 시험을 더했고(REQ-13, `instruction_files` 가 지원 밖 코드일 때), `language` 없는 프로필의 한국어 서식 입력을 AC-9 에 더했어요.
- 감사 뒤 고친 것: B1 — Go 백업에만 있는 옛 항목(`code_comments`)이 새 프로필로 새지 않는다는 단언을 `go_settings_defaults` 에 더했어요(그 필터를 `true` 로 바꾼 변이를 이제 잡아요). B2 — REQ-11 을 `language` 가 있는 프로필로 좁히고, `language` 가 없으면 예전 한국어 서식이 그대로라는 것을 시험으로 고정했어요(기본값을 바꾸면 이미 쓰는 프로젝트의 서식이 바뀌어서 하지 않았어요). N1 — REQ-13 과 시험. N3~N6 은 문체·표현 판단이라 그대로 뒀어요.
- 옛 시험 9개가 새 동작을 말하도록 고쳤어요: lib `language::tests` 4개, `language_setting` 3개, `go_settings_defaults` 2개. `--lib` 1604 통과, `--bins` 56 통과, `language_*`·`agents_*`·`go_settings_*`·`skeleton_language`·`project_*`·`cli_project` 시험 대상 전부 통과.
- `tools/repo_scans.json` 의 검사 39개(cargo 31 + python 8) 전부 통과. 처음 한 번 실패한 것은 매니페스트의 `workflows/project.md` 해시(병합 팁의 값으로 되돌려 해결)와 SPEC 요구 줄 형식(REQ-12 를 Non-goals 로 옮겨 해결)이었어요. 다시 돌려 통과(`product_cli skill_package::` 18, `codex_adapter`·`home_routing_cli`·`exe_link_cli` 37, `source_manifest_test`, `spec_cli the_repository_specs_pass_check`).
- 이 Mac 에서 도는 macOS 시험 파일 10개와 `clippy --all-targets -D warnings`, `cargo fmt --check`, `rustfmt --edition 2021 --check`, `git diff --check` 모두 통과. `moai spec check` 오류 0 경고 0.
- 변이 시험 5개 모두 잡혔어요: `documents` 가 옛 `reports` 값을 잇지 않음, 터미널 단계가 모든 항목을 물음, 답 없는 질문 항목이 값을 못 받음, AGENTS.md 가 대화 언어를 무시함, `user_chat` 코드 경고가 사라짐.

아직 안 한 것: `skills/moai/workflows/project.md` 의 언어 문장(다른 변경이 쓰고, 문장은 노트로 넘겼어요). `moai language-setup` 의 한 줄 도움말("Choose the writing language item by item")은 도움말 골든 6개와 `product_cli` 가 같이 걸려 있어 이 변경에서 못 고쳤어요(`src/help.rs`·골든 파일 후속). 영어 긴 도움말 본문(`src/ops/language_setup.rs`)만 고쳤어요. 영어 한 줄 도움말과 한국어 도움말(`src/help.rs`)은 옛 설명 그대로예요.
