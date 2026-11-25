---
id: SPEC-20261120-reply-language-rule
title: "생성되는 CLAUDE.md 에 대화 언어로 답하라는 지시 문장을 넣고, moai update 가 Claude Code 의 language 설정을 채워요"
status: draft
revision: 2
created: 2026-11-20
branch: spec/reply-language-rule-r2-20261120
---

# 생성되는 CLAUDE.md 에 대화 언어로 답하라는 지시 문장을 넣고, moai update 가 Claude Code 의 language 설정을 채워요

## Intent

사용자 보고(2026-11-20)예요. 대화 언어가 한국어인 프로젝트에서 에이전트가 영어로 답해요. 읽어 보니 원인이 둘이에요. 하나, 생성되는 지침 파일의 관리 구간(`CLAUDE.md` 의 MoAI 사실 블록)은 "작성 언어: …; 그 밖은 영어" 설명 한 줄만 갖고 "이 언어로 답하라"는 지시 문장이 없어요(`src/agents.rs` 의 `render_agents`·`writing_line`, `src/language.rs` 의 `render_lines`). 둘, 사용자 층의 Claude Code 설정(`~/.claude/settings.json`, 또는 `CLAUDE_CONFIG_DIR` 이 가리키는 폴더)에 `language` 도 출력 스타일도 비어 있을 수 있어요. 이 SPEC 은 모든 MoAI 사용자에게 통하게 두 곳을 고쳐요: (1) `CLAUDE.md` 관리 블록이 프로필의 대화 언어(`user_chat`)가 정해졌을 때 작성 언어 줄 뒤에 지시 문장 한 줄을 갖는다. (2) `moai update` 가 관리하는 Claude Code 설정에서 `language` 가 비어 있으면 미리보기·확인·되돌리기를 거쳐 설치할 때 고른 대화 언어로 채우고, 사용자가 정한 값은 덮어쓰지 않는다. 이미 있는 `outputStyle` 처리(비어 있을 때만 채움, 같은 영수증, 되돌리기)와 같은 기계를 쓰고 새로 만들지 않아요.

2026-11-26 에 NeXT 끝(`e946ae8102`)에서 다시 읽은 사실이에요.

- 개정 1(2026-11-20)의 작업 가지 `spec/reply-language-rule-20261120` 은 NeXT 에 합쳐지지 않았어요. NeXT 에는 `language` 설정 키를 쓰는 코드도, 지시 문장도 없어요(`src/home_style.rs`, `src/ops/home_style_apply.rs`, `src/host_home.rs`, `src/update.rs` 에서 `language` 설정 처리를 찾지 못했어요). 그래서 개정 2 는 NeXT 기준에서 다시 만들고, 옛 가지는 참고 자료로만 읽어요.
- 사용자 확정 설계(2026-11-21, 카드 노트 5211): Claude Code 전용이고 팀 지침 정본은 `CLAUDE.md`, 개인은 `CLAUDE.local.md` 예요. `AGENTS.md` 는 정본이 아니고 `@AGENTS.md` 우회도 없어요. 그래서 개정 1 의 `AGENTS.md` 는 모두 `CLAUDE.md` 로 바꿔 읽어요. Codex 는 2026-11-21 에 철회돼(`SPEC-20261121-remove-codex-support`) 개정 1 의 Codex 블록 문장(REQ-8)은 없어요.
- `CLAUDE.md` 관리 블록은 사실만 담고 본문이 15줄 안이에요(`BLOCK_MAX_LINES`). 지금 본문은 최대 9줄이에요.
- `skills/moai/references/home/output-style.md` 는 "사용자의 언어로 쓴다(`language` 설정, 없으면 …)"라고 말하지만 그 설정을 MoAI 가 쓰는 곳은 없어요.
- 공식 문서를 2026-11-25T21:28Z 에 `curl` 로 직접 읽었어요(`code.claude.com/docs/en/settings-reference.md` 의 `### language`): "Claude Code passes the value verbatim to Claude as an instruction to always respond in that language, so any language name Claude can read works", "Claude Code doesn't check the value, so a misspelled name reaches Claude as written rather than producing an error", 값은 "string, any language name, such as "japanese", "spanish", or "french"", 기본은 "unset". 우선순위(`settings.md` 의 Settings precedence): "managed settings, command line, project local, shared project, user" 순이라 프로젝트 `.claude/settings.json` 의 같은 키가 사용자 값을 이겨요. 사용자 파일은 `~/.claude/settings.json` 이고, `CLAUDE_CONFIG_DIR` 은 "Claude Code then stores your settings, session history, and plugins there instead"(Windows 주석 안의 일반 문장)라고만 적혀 있고 `settings.json` 을 이름으로 말하는 문장은 못 찾았어요(unverified: `src/host_home.rs` 가 이미 그 폴더를 쓰고 있어서 이 SPEC 은 그것을 따르고, 시험으로 못 박아요).

## Non-goals

- Claude Code 의 설정 파일을 사용자 값 위에 덮어쓰는 일: 하지 않아요. 키가 이미 있으면 어떤 값이든 그대로 둬요.
- Orca 같은 도구가 따로 만든 프로필 폴더를 찾아다니는 일: 하지 않아요. `moai update` 는 `CLAUDE_CONFIG_DIR` 환경 변수가 가리키는 폴더(없으면 `~/.claude`) 하나를 다뤄요(리더 결정 N2, 카드 노트 4391). 카드 본문의 "per detected config dir" 는 이 뜻으로 읽어요. 다른 프로필은 그 환경으로 `moai update` 를 돌릴 때 같은 방식으로 채워져요.
- 런처(`moai cc`/`glm`)가 언어를 넘기게 하는 일: 두 곳이 되면 필요 없어요(D-5, NeXT 의 런처 코드에는 `language` 를 넘기는 곳이 없어요).
- 항목별 작성 언어의 목록과 질문을 바꾸는 일: 이미 있는 `writing_line` 과 설치 질문은 그대로예요. 이 SPEC 은 그 줄 뒤의 문장 한 줄만 더해요.
- 보고서·커밋·문서 언어를 지시 문장에 묶는 일: 하지 않아요. 그 항목은 `writing_line` 이 이미 말하고, 항목마다 기본값이 달라서(보고서는 영어) 한 문장에 묶으면 프로필과 어긋날 수 있어요(D-6).
- 프로젝트 `.claude/settings.json` 의 `language` 를 읽거나 바꾸는 일: 하지 않아요.
- 새 모듈 파일을 만드는 일: 하지 않아요. 문장은 `src/language.rs` 에 둬요(D-9).

## Requirements

- REQ-1: When the user layer is planned (`moai update`, `home-style-plan`) and the installer saved a conversation language MoAI has a name for, the Claude Code settings target shall plan the `language` key with the lower-case English name of that language. · failure: none
- REQ-2 (REQ-1): IF the settings file already holds a `language` key, THEN the plan shall leave it unchanged whatever its value and say so in the target's detail.
- REQ-3 (REQ-1): IF the installer saved no conversation language or MoAI has no name for it, THEN the plan shall name the reason in the target's detail and leave the `language` key out.
- REQ-4: When one apply writes the settings target, it shall write the missing keys among `outputStyle` and `language` in a single file replacement with one receipt. · failure: none
- REQ-5 (REQ-4): When a receipt is reverted, the revert shall remove only the keys that apply inserted while each still holds the value it wrote.
- REQ-6 (REQ-5): IF a key the apply inserted was changed afterwards, THEN the revert shall leave that key and report it as changed since the apply.
- REQ-7: While `CLAUDE_CONFIG_DIR` names the Claude Code config folder, the plan, the apply and the revert shall use that folder's `settings.json`. · failure: none
- REQ-8: When the profile sets `user_chat` to a language code MoAI has a name for, the generated `CLAUDE.md` managed block shall hold one imperative sentence after the writing-language line that tells the reader to reply to the user and ask questions in that language, written in the block's instruction language. · failure: none
- REQ-9 (REQ-8): IF the profile sets no `user_chat` or its code is not one MoAI has a name for, THEN the managed block shall be byte-identical to the block written before this change.
- REQ-10: When `moai update` asks the person to confirm the user layer, its question and its summary shall show each settings key the style step would write with its value. · failure: none
- REQ-11 (REQ-5): When a receipt written before this change, without a key list, is reverted, the revert shall behave as it did before this change. · failure: none

## Acceptance

- AC-1 (REQ-1): 설치할 때 한국어를 골랐고 `settings.json` 이 없거나 `language` 가 없으면 계획이 `language` 를 `korean` 으로 채우는 것으로 보이고, 적용하면 파일에 그 키가 생겨요 · verify: `cargo test --locked --offline --test home_style_cli language_key_plan_sets_the_saved_conversation_language`
- AC-2 (REQ-2): `language` 가 이미 `japanese` 든 빈 문자열이든 그대로 두고 계획의 설명이 그 사실을 말해요 · verify: `cargo test --locked --offline --test home_style_cli language_key_user_value_is_kept_and_said_so`
- AC-3 (REQ-3): 대화 언어가 저장되지 않았거나 이름을 모르는 언어(`fr`)이면 `language` 키를 계획하지 않고 이유를 말해요 · verify: `cargo test --locked --offline --test home_style_cli language_key_unknown_language_names_the_reason`
- AC-4 (REQ-4, REQ-5, REQ-6, REQ-11): 한 번의 적용이 두 키를 한 파일 교체로 쓰고, 되돌리면 그 키만 지워지고 사용자의 다른 값은 그대로이며, 적용 뒤 사용자가 바꾼 키는 지우지 않고 `changed since` 로 보고하고, 키 목록이 없는 옛 영수증도 되돌려요 · verify: `cargo test --locked --offline --test home_style_cli language_key_one_apply_one_receipt_revert_removes_only_those`
- AC-5 (REQ-7): `CLAUDE_CONFIG_DIR` 로 가리킨 임시 폴더의 `settings.json` 이 채워지고 임시 HOME 의 기본 Claude 폴더는 건드리지 않아요 · verify: `cargo test --locked --offline --test home_style_cli language_key_config_dir_profile_gets_its_own_settings`
- AC-6 (REQ-8, REQ-9): 대화 언어를 정한 프로필 고정 입력(`en`·`ko` 두 지시 언어)이 만든 `CLAUDE.md` 관리 블록은 작성 언어 줄 뒤에 지시 문장을 갖고(`Reply to the user and ask questions in Korean.` / `사용자에게는 한국어로 답하고 질문하세요.`), `user_chat` 이 없거나 모르는 코드면 문장이 없고 블록이 이전과 바이트가 같아요 · verify: `cargo test --locked --offline --test language_setting reply_sentence_`
- AC-7 (REQ-8): 문장 함수가 모든 지원 언어 코드에 두 지시 언어로 문장을 만들고(한국어 이름은 모두 모음으로 끝나 조사가 `로`), 모르는 코드는 `None` 이고, 블록 본문이 `BLOCK_MAX_LINES` 안이에요 · verify: `cargo test --locked --offline --lib reply_`
- AC-8 (REQ-10): `moai update` 의 확인 질문과 요약이 `settings.json: language = korean` 처럼 설정 키와 값을 보여 주고, 저장된 언어가 없으면 `language` 를 보이지 않아요 · verify: `cargo test --locked --offline --test update_tux_golden the_update_question_names_the_settings_key_a_style_target_writes`
  verify: `cargo test --locked --offline --lib update::user_layer_tests::the_summary_names_the_language_key_a_saved_language_would_write`
- AC-9 (REQ-1…REQ-11): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test language_three_questions`
  verify: `cargo test --locked --offline --test home_style_cli`
  verify: `cargo test --locked --offline --test update_tux_golden`

## Scope

- write: `.moai/specs/SPEC-20261120-reply-language-rule/spec.md`, `src/language.rs`, `src/agents.rs`, `src/home_style.rs`, `src/update.rs`, `src/update_tux.rs`, `tests/language_setting.rs`, `tests/home_style_cli.rs`, `tests/update_tux_golden.rs`
- read-only: `src/host_home.rs`, `src/home_root.rs`, `src/ops/home_style_apply.rs`, `src/ops/home_style_plan.rs`, `tests/language_three_questions.rs`
- 사용자 층 스타일 설명 파일(skills 의 moai 참조 중 output-style 문서)은 해시가 시험에 고정돼 있어서 바꾸지 않아요.
- forbidden: 사용자 값 덮어쓰기, 실제 HOME 과 그 아래 Claude 설정 폴더 읽고 쓰기(시험은 임시 HOME 만), 공용 파일(`src/lib.rs`, `tools/repo_scans.json` 등)을 범위에 넣는 것(필요하면 `card line-lease`), 카드 번호를 SPEC id·가지·커밋에 쓰는 것
- 다른 카드가 쥔 곳: `src/update.rs`, `src/update_tux.rs` 는 다른 카드 둘이 진행 중이에요. 이 SPEC 의 화면 조각(Plan 4)은 그 둘이 끝난 뒤 새 기준에서 해요.

## Plan

1. 테스트 먼저, 조각마다 빨간 시험을 먼저 봐요.
2. 조각 1(지시 문장, 기준 1): `src/language.rs` 에 `reply_sentence(conversation, ui)` 를 더하고(`render_lines`·`english_name` 옆, 비공개 `code_name` 사용), `src/agents.rs` 에 `reply_line(p, ui)` 를 `writing_line` 옆에 더해 `render_agents` 본문이 `{writing}{reply}{versioning}{flow}` 가 되게 한다. AC-6·AC-7.
3. 조각 2(설정 키, 기준 2): `src/home_style.rs` 의 `setting_plan`·`insert_style`·`revert_setting` 이 `language` 도 같은 대상·같은 영수증으로 채우고 되돌린다. `TargetPlan` 과 영수증에 키 목록을 `serde(default, skip_serializing_if)` 로 더해 옛 계획 digest 와 영수증이 그대로이게 한다. 기존 가드(스타일 파일이 우리 것이 아닐 때, 공유 폴더)는 `outputStyle` 에만 적용하고 `language` 는 설정 파일을 읽을 수 없거나 객체가 아닐 때만 멈춘다. AC-1~AC-5.
4. 조각 3(화면, 기준 2의 확인): `src/update.rs` 의 `layer_summary` 와 `src/update_tux.rs` 의 스타일 행·변경 목록이 설정 키와 값을 보인다. 확인은 기존 사용자 층 확인 하나이고 새 토큰은 없다. AC-8.
5. 조각 4: Result 에 AC 마다 `- AC-n: met — 근거` 한 줄을 쓴다.
6. 위험 경로(HOME 이 임시 폴더인 실제 `moai update`), 가드 뮤턴트 한 줄씩, `check affected --card` 한 번.

## Risks

- **전역 설정 변경이에요(tier 2).** 파일을 쓰는 일은 기존 `outputStyle` 과 같은 길(미리보기, `update` 의 확인, 받은 영수증으로 되돌리기)만 타고, 키가 이미 있으면 쓰지 않아요. `language` 값은 Claude Code 가 검증 없이 모델에 넘기는 언어 이름이라 틀린 이름을 쓰면 엉뚱한 언어로 답해요: 이름은 MoAI 가 아는 코드 표(`english_name`)에서만 가져와요.
- 계획이 두 키를 한 대상에 묶어서 `TargetPlan` 과 영수증이 커져요. 옛 영수증을 읽는 되돌리기는 `outputStyle` 만 지우는 옛 동작을 지켜야 해요(AC-4 의 시험).
- 대화 언어의 출처가 둘이에요: `CLAUDE.md` 문장은 프로필의 `user_chat`, `language` 키는 HOME 에 저장한 설치 선택이에요. 둘이 다르면 문장과 설정이 서로 다른 언어를 말해요(프로필이 프로젝트 쪽이라 프로젝트가 더 가까워요).
- 프로젝트의 `.claude/settings.json` 에 `language` 가 있으면 사용자 층 값을 이겨요(공식 문서의 우선순위, Intent). 이 SPEC 은 사용자 층만 채우고 프로젝트 설정은 건드리지 않아요. 그 경우 `CLAUDE.md` 의 지시 문장(REQ-8)이 둘째 안전망이에요.
- 화면 조각이 쥔 파일이 다른 카드와 겹쳐서 착지 때 충돌할 수 있어요. 그래서 조각 3 은 그 카드들이 끝난 뒤 새 기준에서 해요.
- `src/update.rs`·`src/update_tux.rs` 를 건드리면 `check affected` 가 넓은 스캔 묶음을 골라요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | Claude Code 의 키는 `language`(문자열, 언어 이름), 값은 `language::english_name` 의 소문자(`korean`) | 공식 설정 참조(`code.claude.com/docs/en/settings-reference.md`, 2026-11-25T21:28:36Z 읽음, sha256 `f87b28f014d323f1d99655f4dd8f329693e8295eec0f1cecf8af0bcc534a5249`; 우선순위는 `settings.md`, 같은 시각, sha256 `7587a3251756cd151948718b70fff20fd26e3a4b545858d172d0941d57ab8f10`; 해시는 `curl -sL` 이 돌려준 원본 바이트의 것이고 인용한 텍스트와 같은 요청이 아닐 수 있어요) | 키 한 줄 삭제 |
| D-2 | `outputStyle` 과 `language` 는 한 대상·한 파일 교체·한 영수증으로 써요 | 같은 파일에 대상 둘을 쓰면 둘째 쓰기가 "계획 뒤 바뀜" 으로 거절돼요(`apply_for` 의 `expected_sha256` 비교) | 대상을 둘로 가르고 쓰기를 직렬화 |
| D-3 | 키가 이미 있으면 값이 무엇이든 건드리지 않아요. `outputStyle` 이 다른 값이거나 스타일 파일이 우리 것이 아니어도 `language` 가 비어 있으면 `language` 는 채워요 | 카드의 "사용자 값은 덮어쓰지 않음"; 두 키는 서로 다른 사용자 선택이에요 | 둘을 함께 충돌로 묶기 |
| D-4 | `moai update` 는 `CLAUDE_CONFIG_DIR`(없으면 `~/.claude`) 한 폴더를 다뤄요; 다른 프로필을 찾아다니지 않아요 | 리더 결정 N2(노트 4391); 도구별 프로필 위치는 일반 사용자 기준이 아니에요 | 후속 카드로 프로필 목록 |
| D-5 | 런처는 `language` 를 넘기지 않고 그대로 두어요 | NeXT 의 런처 코드에서 `language` 를 넘기는 곳을 찾지 못했어요(`src/launcher_init.rs` 의 `language: "en"` 한 곳은 시험용 기본 요청) | 후속 |
| D-6 | 지시 문장은 대화 언어만 말해요(`Reply to the user and ask questions in <language>.` / `사용자에게는 <언어>로 답하고 질문하세요.`) | 보고서·커밋·문서 언어는 `writing_line` 이 이미 말하고, 항목 기본값이 달라 한 문장에 묶으면 프로필과 어긋날 수 있어요(보고서는 영어) | 개정 1 의 세 언어 문장 |
| D-7 | `user_chat` 이 `en` 이어도 문장을 써요 | 명시한 선택은 결정적으로 보이는 것이 낫고, 테스트가 두 지시 언어를 모두 고정해요 | `en` 이면 생략 |
| D-8 | 확인은 기존 사용자 층 확인 하나가 설정 키와 값을 보여요. 새 확인 토큰은 두지 않아요 | 감사 B1 이 요구한 것은 사용자가 무엇에 동의하는지 보이는 것이고, 토큰을 더하면 `update.rs` 변경이 커져요(카드 문구 "own confirm" 은 이 사용자 층 확인을 가리킨다고 읽어요) | 키마다 별도 토큰 |
| D-9 | 문장은 `src/language.rs` 안에 두고 새 모듈 파일을 만들지 않아요 | 함수 하나(약 20줄)이고 비공개 `code_name` 이 필요해요. 새 파일은 공용 `src/lib.rs` 줄 임대가 필요해요 | `src/reply_language.rs` |
| D-10 | Claude Code 전용이에요. Codex 블록과 `AGENTS.md` 는 다루지 않아요 | 사용자 확정(2026-11-21)과 Codex 철회(2026-11-21) | 개정 1 의 REQ-8 |

## Result

개정 2 를 쓴 단계예요. 구현은 아직 시작하지 않았어요.
