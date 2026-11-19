---
id: SPEC-20261120-question-bank-generic
title: "The interview question bank is generic: no decision model, three language rows, SemVer first with HeadVer suggested for m-flow, and no citations"
status: draft
revision: 1
created: 2026-11-20
---

# The interview question bank is generic: no decision model, three language rows, SemVer first with HeadVer suggested for m-flow, and no citations

## Intent

사용자 결정(2026-11-19, 리더 대화 (1)·(2)·(6)·(7))의 질문 은행 쪽이에요. 배포되는 스킬 글은 어떤 사용자의 프로젝트에도 맞아야 해서, 질문과 선택지가 moai-adk 자신의 개발 도구·카드 번호·Issue 번호를 말하면 안 돼요. tip(70707144ba)에서 읽은 사실이에요. `references/internal/interview.md` 의 질문 은행에는 `decision_model` 행(Jev·OpenAI)과 그 설명이 있고, 언어를 묻는 행이 없어요(언어는 `workflows/project.md` 가 `moai language-setup` 의 8항목으로 물어요). `autonomy` 행은 "SKILL.md sorts as case (2)" 라고 말하고, `work_flow` 행은 `versioning_head` 뒤에 있어서 버전 질문이 m-flow 여부를 모른 채 나와요. 세 파일(`interview.md`, `questions.md`, `versioning.md`)에는 `card C-51`, `Issue #2639`, `user decision 2026-11-16` 같은 인용이 모두 합쳐 열두 군데쯤 있어요.
이 SPEC 은 이 세 파일의 글과 그 글을 고정하는 시험만 다뤄요. `workflows/project.md`·`workflows/update.md` 의 경계와 같은 두 파일의 인용 정리는 다른 SPEC(`SPEC-20261120-project-update-boundary`)이 맡아요. 새 Rust 명령은 없어요.

## Non-goals

- `workflows/project.md`·`workflows/update.md` 바꾸기: 경계 SPEC 의 일이고, 두 파일은 다른 카드의 예약 중이에요. 그래서 project.md 의 "Decision model" 단계는 그 SPEC 이 지워요.
- OMP 줄(`questions.md` 의 호스트 표, `hosts` 행): OMP 지원을 지우는 카드가 같은 시험을 바꾸므로 거기서 지워요.
- `src/language.rs`·`moai language-setup` 의 질문 수 바꾸기: 언어 카드의 코드예요. 이 SPEC 은 은행이 묻는 세 항목을 글로 정해요.
- `src/decide/**` 와 결정 모델의 시험·SPEC 폴더 삭제: 결정 모델 삭제 카드의 일이에요. 이 SPEC 은 은행과 글만 지우고, 글을 고정하던 `tests/project_decide_model.rs` 는 주제가 사라지므로 같이 지워요.
- 모델·effort 질문, 프로필 검증 규칙, 릴리스 도구 점검 질문.

## Requirements

- REQ-1: The question bank shall hold no `decision_model` row and no text about a decision model, a provider or a key question. · failure: none — the bank ids and the file text are checked by a test, and a returning row fails it.
- REQ-2: The question bank shall ask the writing language as three rows — conversation, commit messages and documents — each filling the profile `language`. · failure: none — the row ids and their destination are checked by a test.
- REQ-3: The question bank shall state that every other language item is English and is not asked. · failure: none — the sentence is checked by a test.
- REQ-4: The `versioning` row shall offer a detected scheme first for confirmation, then SemVer, HeadVer and CalVer, with SemVer first and recommended when nothing was detected. · failure: none — the order of the options is checked by a test.
- REQ-5: When the work flow is m-flow, the `versioning` row shall label HeadVer as suggested with m-flow while SemVer stays first and recommended. · failure: none — the sentence is checked by a test.
- REQ-6: The `work_flow` row shall come before the `versioning` rows in the bank. · failure: none — the row order is checked by a test.
- REQ-7: The text of the three internal reference files shall name no development tool of this repository, no card number, no Issue number, no dated user-decision citation, and no option for a decision provider, Kiro or Cursor. · failure: none — a scan test lists the forbidden shapes and reads the three files.
- REQ-8: The `autonomy` row shall state what is asked first in its own words, without pointing at a case number of another file. · failure: none — the row text is checked by a test.

## Acceptance

- AC-1 (REQ-1): 은행 id 목록에 `decision_model` 이 없고 `interview.md` 에 `decision_model`, `Jev`, `OpenAI`, "key question" 이 없어요.
  verify: `cargo test --locked --offline --test question_bank_generic the_bank_has_no_decision_model`
- AC-2 (REQ-2, REQ-3): 은행에 `language_conversation`, `language_commits`, `language_documents` 세 행이 있고 각 행이 프로필 `language` 로 가며, 나머지 항목은 영어이고 묻지 않는다는 문장이 있어요.
  verify: `cargo test --locked --offline --test question_bank_generic the_bank_asks_three_language_rows_and_fixes_the_rest_to_english`
- AC-3 (REQ-4, REQ-5, REQ-6): `versioning` 행의 선택지 순서가 감지된 방식 → SemVer → HeadVer → CalVer 이고, m-flow 일 때 HeadVer 를 "suggested with m-flow" 로 표기하는 문장이 있고, `work_flow` 행이 `versioning` 행보다 앞에 있어요.
  verify: `cargo test --locked --offline --test question_bank_generic versioning_options_follow_the_decided_order_and_come_after_the_work_flow`
- AC-4 (REQ-7, REQ-8): 세 파일에 금지 모양이 하나도 없고 `autonomy` 행이 다른 파일의 case 번호를 가리키지 않아요.
  verify: `cargo test --locked --offline --test question_bank_generic the_bank_files_name_no_card_issue_or_repository_tool`
- AC-5 (REQ-1…REQ-8): 이웃 검사.
  verify: `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test question_bank_hosts`; `cargo test --locked --offline --test codex_question_tool`; `cargo test --locked --offline --test core_rules_2876`; `cargo test --locked --offline --test spec_branch_name_first`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test doc_references`; `python3 -m unittest tests.source_manifest_test tests.skill_references_test`

## Scope

- write: `.moai/specs/SPEC-20261120-question-bank-generic/spec.md`, `tests/question_bank_generic.rs`, `skills/moai/references/internal/interview.md`, `skills/moai/references/internal/questions.md`, `skills/moai/references/internal/versioning.md`, `tests/agent_defs_package.rs`, `tests/update_route.rs`, `tests/project_decide_model.rs` (삭제), `skills/moai/references/source-manifest.json`
- read-only: `src/language.rs`, `src/profile.rs`, `tests/question_bank_hosts.rs`, `tests/codex_question_tool.rs`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기

## Plan

1. 시험 먼저: `tests/question_bank_generic.rs` 의 시험 넷을 쓰고, 넷이 실패하는 것(RED)을 읽어요.
2. 세 파일의 글을 고쳐 GREEN 으로 만들고, 그 글을 고정하던 기존 시험(`agent_defs_package`, `update_route`)의 인용 문구를 새 글에 맞춰요. `tests/project_decide_model.rs` 는 지워요.
3. 이웃 검사(AC-5)와 READY 전 검사 전부를 돌려요. 매니페스트 해시는 마지막 커밋에서만 써요.

## Risks

- 기존 시험이 옛 인용 문구를 고정하고 있어요(`update_route`·`agent_defs_package`) → 글을 먼저 고치고 어느 시험이 어느 문구를 읽는지 전부 돌려서 읽어요(같은 영역 시험 규칙).
- 결정 모델 삭제 카드와 같은 시험 파일(`project_decide_model.rs`, `agent_defs_package.rs`)을 건드려요 → 지우기는 같은 파일을 같은 쪽으로(삭제) 바꾸니 합쳐도 충돌이 없고, `agent_defs_package.rs` 는 서로 다른 시험을 고쳐요. 합칠 때 `git merge-tree` 로 읽어요.
- 은행에서 `decision_model` 이 사라진 동안 `workflows/project.md` 가 아직 그 행을 묻는다고 말해요 → 경계 SPEC 이 같은 단계를 지워요. 두 조각 사이는 한 릴리스 PR 안이에요.
- 인용을 빼면 규칙의 근거를 잃어요 → 근거는 `docs/` 의 결정 기록과 카드에 있고, 배포 글에는 규칙만 둬요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 질문 은행의 언어는 세 행이 묻고, 나머지 다섯 항목은 영어로 정하고 묻지 않아요 | 사용자 결정 (1): 대화·커밋·문서만 묻고 나머지는 영어; AGENTS.md 는 대화 언어를 따라요 | 행 셋을 지워요 |
| D-2 | HeadVer 는 두 번째 선택지로 두고 m-flow 일 때만 "suggested with m-flow" 로 표기해요; SemVer 가 첫째이고 권장이에요 | 사용자 결정 (7) | 표기 문장을 지워요 |
| D-3 | OMP 줄은 이 SPEC 에서 지우지 않아요 | `tests/question_bank_hosts.rs`·`tests/codex_question_tool.rs` 가 OMP 줄을 고정하고, OMP 를 지우는 카드가 그 시험을 같이 바꿔요 | 그 카드가 지워요 |

## Result

(채워요: RED, GREEN, 변이, 이웃, Not run)
