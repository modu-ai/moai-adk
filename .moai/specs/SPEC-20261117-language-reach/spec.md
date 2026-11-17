---
id: SPEC-20261117-language-reach
title: 언어 설정이 에이전트에 닿게 — 카드 C-87 조각 1
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

사용자 불만(2026-11-17): "/moai project 설정 때 언어랑 각종 설정을 했는데 전혀 적용이 안된거 같다. 왜 자꾸 영어로
대화를 하지?" 프로필의 항목별 언어(`language`)는 저장되지만, AGENTS.md 관리 구간은 사용자가 고르지 않은 항목을
요청 언어로 채워 넣고(`user_chat`이 일본어여도 보고는 영어), 세션 시작·압축 뒤 hook 문장에는 대화 언어가 없고,
기본 에이전트 규칙은 하위 에이전트 보고를 사용자 언어로 옮겨 전하라고 말하지 않아요. 이 SPEC은 카드 C-87의
AC1·AC3을 다루고, AC2는 기존 사용자 결정과 부딪혀서 결정을 기다려요(Decisions D4).

## Non-goals

- CLI 사람용 출력의 언어 바꾸기(카드 AC2): 사용자 결정 6878265192("CLI 사람용 출력은 영어만, 사용자 언어는 LLM
  세션 몫")와 부딪혀요. 결정이 나오면 다음 조각.
- 적용된 설정 표(카드 AC4).
- 이 저장소의 `.moai/project.yaml`(무시되는 로컬 파일) 고치기.
- `AGENTS.md`, `install/**`, `ci.yml`, `src/optimize.rs`, `tools/build_folder.py`(다른 에이전트 작업 중).

## Requirements

- REQ-1: When the managed section of AGENTS.md is rendered from a profile with a `language` key, the generator shall write one line for each item the profile sets, naming the language and its code.
- REQ-2: The generator shall write no language line for an item the profile does not set, and no language block when the profile has no `language` key.
- REQ-3 (REQ-1): IF the profile names an unknown item or an unsupported language code, THEN the plan shall refuse the profile without echoing the value.
- REQ-4: When a session starts, resumes or is compacted, the session-check hook shall add one sentence naming the conversation language from the project's `language.user_chat`, else from the language the user recorded in HOME, and telling the model to talk with the user and relay subagent reports in it.
- REQ-5 (REQ-4): IF the project profile is unreadable or names no conversation language and HOME records none, THEN the hook shall add no language sentence and still answer with exit 0.
- REQ-6: The core rules shall tell every route and agent to relay a subagent's report to the user in the user's conversation language.

## Acceptance

- AC-1 (REQ-1, REQ-2): 항목마다 그 항목 한 줄만, 고르지 않은 항목은 줄 없음, 키가 없으면 구간이 예전과 같음.
  verify: `cargo test --locked --offline --test language_setting`; `cargo test --locked --offline --lib language::`
- AC-2 (REQ-3): 모르는 항목·코드는 거절되고 값을 되풀이하지 않아요.
  verify: `cargo test --locked --offline --test language_setting unknown_items_and_codes_are_refused`
- AC-3 (REQ-4, REQ-5): 시작·재개·압축 때 대화 언어 문장이 붙고, 프로필이 없거나 깨졌고 HOME 값도 없으면 붙지 않아요.
  verify: `cargo test --locked --offline --test hook_dispatch_cli session_check_names_the_conversation_language`
- AC-4 (REQ-6): 핵심 규칙 7이 하위 에이전트 보고를 사용자 언어로 전하라고 말하고 크기 제한(4096바이트) 안이에요.
  verify: `cargo test --locked --offline --test core_rules_2876`; `python3 -I tools/source_manifest.py check`
- AC-5 (REQ-1…REQ-6): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --test agents_project_cli`;
  `cargo test --locked --offline --test gitignore_block_cli`; `cargo test --locked --offline --test repo_agents_md`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/source_manifest_test.py tests/skill_references_test.py`

## Scope

- write: `src/language.rs`, `src/agents.rs`, `src/hook_engine.rs`, `tests/language_setting.rs`,
  `tests/hook_dispatch_cli.rs`, `tests/core_rules_2876.rs`, `skills/moai/references/internal/core-rules.md`,
  `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261117-language-reach/spec.md`
- read-only: `src/ui.rs`, `src/ui_text.rs`, `src/error_next.rs`, `src/profile.rs`
- forbidden: `AGENTS.md`, `install/**`, `.github/workflows/ci.yml`, `src/optimize.rs`, `tools/build_folder.py`,
  이 저장소의 `.moai/project.yaml`, 실제 HOME

## Plan

1. 시험 먼저: 관리 구간의 항목별 줄과 빠진 항목, hook의 대화 언어 문장, RED 기록.
2. `src/language.rs`의 렌더를 고른 항목만으로, `src/hook_engine.rs`에 대화 언어 문장.
3. 핵심 규칙 7에 하위 에이전트 보고 전달 문장, skill 목록 다시 쓰기와 확인.
4. 이웃 시험, mutation, 결과.

## Risks

- 고르지 않은 항목이 이제 관리 구간에서 빠져요 → 인터뷰는 모든 항목을 쓰므로 보통 프로필에는 변화가 없고, 손으로
  일부만 적은 프로필만 줄이 줄어요(Decisions D1).
- hook이 HOME의 언어 파일을 읽어요 → 읽기만 하고, 실패는 조용히 문장 없음으로 끝나요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 관리 구간은 프로필이 정한 항목만 한 줄씩 적고, 고르지 않은 항목은 기본값으로 채우지 않아요(추측 금지); #2293의 "빠진 항목은 기본값으로 적음"을 대체 | 조정자 지시 "absent → no line (never guess)" | `render_lines_in`이 `resolve`로 다시 채우기 |
| D2 | hook의 대화 언어는 프로젝트 `user_chat`, 없으면 HOME에 사용자가 기록한 값(설치 때 선택), 둘 다 없으면 문장 없음 | 추측 금지, 사용자 기록만 | HOME 단계 빼기 |
| D3 | 대화 언어 문장은 시작·재개·압축 모두에 붙여요(압축 뒤에도 남게) | 카드 AC3 | 압축 때만 |
| D4 | 카드 AC2(CLI 출력이 `user_chat`을 따름)는 하지 않고 결정을 요청해요: 사용자 결정 6878265192(CLI 사람용 출력은 영어만)와 정면으로 부딪혀요 | core rule 1·17 | 결정 뒤 `ui::ui_lang`이 `ui_text::for_stream`을 쓰게 |
| D5 | 하위 에이전트 보고 규칙은 core-rules 7에 한 문장(모든 경로·에이전트가 읽음), `mode-agent.md`는 늘리지 않음 | 크기 제한 | 다른 파일로 옮기기 |

## Result

아직 비어 있어요(구현 뒤 채워요).
