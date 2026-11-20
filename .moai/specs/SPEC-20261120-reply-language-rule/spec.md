---
id: SPEC-20261120-reply-language-rule
title: "Reply-language rule in AGENTS.md and the Claude Code language setting"
status: draft
revision: 1
created: 2026-11-20
---

# Reply-language rule in AGENTS.md and the Claude Code language setting

## Intent

사용자 보고(2026-11-20)예요. 대화 언어가 한국어인 프로젝트에서 에이전트가 영어로 답해요. 읽어 보니 원인이 둘이에요. 하나, 생성되는 AGENTS.md 관리 구간은 "항목별 작성 언어" 목록만 보여 주고 "이 언어로 답하라"는 지시 문장이 없어요(`src/agents.rs` 의 `render_agents`, `src/language.rs` 의 `render_lines`). 둘, 사용자 층의 Claude Code 설정(`~/.claude/settings.json`, 또는 `CLAUDE_CONFIG_DIR` 이 가리키는 폴더)에 `language` 도 출력 스타일도 비어 있었어요. 이 SPEC 은 모든 MoAI 사용자에게 통하게 두 곳을 고쳐요: (1) AGENTS.md 관리 구간이 언어 목록 앞에 지시 문장을 갖는다. (2) `moai update` 가 관리하는 Claude Code 설정에서 `language` 가 비어 있으면 미리보기·확인·되돌리기를 거쳐 설치할 때 고른 대화 언어로 채우고, 사용자가 정한 값은 덮어쓰지 않는다. Codex 는 응답 언어 설정 키가 없어요(공식 문서 확인, 아래 Decisions), 그래서 AGENTS.md 문장이 유일한 길이고, 전역 `~/.codex/AGENTS.md` 의 MoAI 관리 블록에도 같은 한 문장을 둬요. 이미 있는 `outputStyle` 처리(비어 있을 때만 채움, 같은 영수증, 되돌리기)와 같은 기계를 쓰고 새로 만들지 않아요.

## Non-goals

- Claude Code 나 Codex 의 설정 파일을 사용자 값 위에 덮어쓰는 일: 하지 않아요. 키가 이미 있으면 어떤 값이든 그대로 둬요.
- Orca 같은 도구가 따로 만든 프로필 폴더를 찾아다니는 일: 하지 않아요. `moai update` 는 `CLAUDE_CONFIG_DIR` 환경 변수가 가리키는 폴더(없으면 `~/.claude`) 하나를 다뤄요. 다른 프로필은 그 환경으로 `moai update` 를 돌릴 때 같은 방식으로 채워져요.
- 런처(`moai cc`/`glm`/`codex`)가 언어를 넘기게 하는 일: 위 두 곳이 되면 필요 없어요(Decisions D-5, 확인 결과 런처는 `language` 를 넘기지 않아요).
- 항목별 작성 언어의 목록과 질문을 바꾸는 일: 다른 카드가 하고 있어요(세 질문 정리). 이 SPEC 은 그 목록 앞의 문장만 더해요.

## Requirements

- REQ-1: When the user layer is planned (`moai update`, `home-style-plan`) and the installer saved a conversation language MoAI has a name for, the Claude Code settings target shall plan the `language` key with the lower-case English name of that language. · failure: none
- REQ-2 (REQ-1): IF the settings file already holds a `language` key, THEN the plan shall leave it unchanged whatever its value and say so in the target's detail.
- REQ-3 (REQ-1): IF the installer saved no conversation language or MoAI has no name for it, THEN the plan shall name the reason in the target's detail and leave the `language` key out.
- REQ-4: When one apply writes the settings target, it shall write the missing keys among `outputStyle` and `language` in a single file replacement with one receipt. · failure: none
- REQ-9: When a receipt is reverted, the revert shall remove only the keys that apply inserted while each still holds the value it wrote. · failure: none
- REQ-5 (REQ-9): IF a key the apply inserted was changed afterwards, THEN the revert shall leave that key and report it as changed since the apply.
- REQ-6: While `CLAUDE_CONFIG_DIR` names the Claude Code config folder, the plan, the apply and the revert shall use that folder's `settings.json`. · failure: none
- REQ-7: When the generated AGENTS.md managed section lists the writing languages, it shall begin that part with one imperative sentence in the instruction language that names the conversation, commit and document languages and says every other text is English. · failure: none
- REQ-8: When the saved conversation language is known, the Codex managed block in the user-level AGENTS.md shall carry one sentence that tells the model to reply in that language. · failure: none

## Acceptance

- AC-1 (REQ-1): 설치할 때 한국어를 골랐고 `settings.json` 이 없거나 `language` 가 없으면 계획이 `language` 를 `korean` 으로 채우는 것으로 보이고, 적용하면 파일에 그 키가 생겨요 · verify: `cargo test --locked --offline --test home_style_language the_plan_sets_language_to_the_saved_conversation_language`
- AC-2 (REQ-2): `language` 가 이미 `japanese` 든 빈 문자열이든 그대로 두고 계획의 설명이 그 사실을 말해요 · verify: `cargo test --locked --offline --test home_style_language a_user_set_language_is_kept_and_said_so`
- AC-3 (REQ-3): 대화 언어가 저장되지 않았거나 이름을 모르는 언어(`fr`)이면 `language` 키를 계획하지 않고 이유를 말해요 · verify: `cargo test --locked --offline --test home_style_language without_a_known_conversation_language_no_key_is_planned_and_the_reason_is_named`
- AC-4 (REQ-4, REQ-5, REQ-9): 한 번의 적용이 두 키를 한 파일 교체로 쓰고, 되돌리면 그 키만 지워지고 사용자의 다른 값은 그대로이며, 적용 뒤 사용자가 바꾼 키는 지우지 않고 `changed since` 로 보고해요 · verify: `cargo test --locked --offline --test home_style_language one_apply_writes_both_keys_and_revert_removes_only_those`
- AC-5 (REQ-6): `CLAUDE_CONFIG_DIR` 로 가리킨 임시 폴더의 `settings.json` 이 채워지고 임시 HOME 의 기본 Claude 폴더는 건드리지 않아요 · verify: `cargo test --locked --offline --test home_style_language a_config_dir_profile_gets_its_own_settings`
- AC-6 (REQ-7): 지시 문장이 `en`·`ko` 두 지시 언어로 만들어지고, 대화·커밋·문서 언어 이름이 들어가며, "그 밖은 영어" 가 들어가요 · verify: `cargo test --locked --offline --lib reply_language::`
- AC-7 (REQ-8): 한국어를 골랐으면 Codex 블록에 한국어로 답하라는 한 문장이 있고, 언어를 모르면 블록 본문은 이전과 바이트가 같아요 · verify: `cargo test --locked --offline --test home_style_language the_codex_block_carries_the_reply_sentence_when_the_language_is_known`

## Scope

- write: 이 SPEC, `src/home_style.rs`, `src/reply_language.rs`(새), `tests/home_style_language.rs`(새), 필요하면 `tests/home_style_cli.rs`(기존 기대값 갱신)
- 다른 카드가 끝나야 쓸 곳: `src/agents.rs` 의 호출 한 줄(세 질문 정리가 같은 함수를 고치고 있어요), 모듈 한 줄 `src/lib.rs`(공용이라 짧게 잡고 바로 반환)
- read-only: `src/language.rs`, `src/host_home.rs`, `src/update.rs`, `src/home_routing.rs`
- forbidden: 사용자 값 덮어쓰기, 실제 HOME 과 그 아래 두 도구의 사용자 폴더 읽고 쓰기(시험은 임시 HOME 만), 카드 번호를 SPEC id·가지·커밋에 쓰는 것

## Plan

1. 조각 1(이번): SPEC, 조사 보고, PLAN 노트, 범위·순서 결정 요청.
2. 조각 2(설정 키): `home_style` 의 설정 대상이 `language` 도 채우고 한 영수증으로 되돌린다. AC-1~5.
3. 조각 3(문장): `src/reply_language.rs` 가 지시 문장을 만든다(AC-6). 호출 한 줄은 `render_agents` 를 쥔 카드가 합쳐진 뒤.
4. 조각 4(Codex 블록): 같은 한 문장을 `moai-output-style` 블록에 더한다(AC-7).

## Risks

- **전역 설정 변경이에요(tier 2).** 파일을 쓰는 일은 기존 `outputStyle` 과 같은 길(미리보기, `update` 의 별도 확인, 받은 영수증으로 되돌리기)만 타고, 키가 이미 있으면 쓰지 않아요. `language` 값은 Claude Code 가 검증 없이 그대로 모델에 넘기는 언어 이름이라 틀린 이름을 쓰면 엉뚱한 언어로 답해요: 이름은 MoAI 가 아는 코드 표에서만 가져와요.
- 계획이 두 키를 한 대상에 묶어서 기존 계획의 모양(`TargetPlan`)과 영수증(`inserted` 텍스트)이 커져요. 옛 영수증을 읽는 되돌리기는 `outputStyle` 만 지우는 옛 동작을 그대로 지켜야 해요(시험).
- Codex 전역 `AGENTS.md` 의 문장은 프로젝트 `AGENTS.md` 가 뒤에 와서 덮을 수 있어요(공식 문서: 루트에서 아래로 이어 붙이고 가까운 파일이 뒤). 그래서 프로젝트 쪽 지시 문장(REQ-7)도 필요해요.
- 세 질문을 정리하는 카드가 `render_agents` 를 고치고 있어서 호출 줄을 넣는 시점이 늦어져요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Claude Code 의 키는 `language`(문자열, 언어 이름), 값은 `language::english_name` 의 소문자(`korean`) | 공식 설정 참조: "string, any language name, such as `japanese`, `spanish`, or `french`; Claude Code doesn't validate it"; 모델에 그대로 전달된다고 적혀 있어요(code.claude.com/docs/en/settings-reference.md, 2026-11-20 04:36Z 읽음) | 키 한 줄 삭제 |
| D-2 | `outputStyle` 과 `language` 는 한 대상·한 파일 교체·한 영수증으로 써요 | 같은 파일에 대상 둘을 쓰면 둘째 쓰기가 "계획 뒤 바뀜" 으로 거절돼요(`apply_for` 의 `expected_sha256` 비교) | 대상을 둘로 가르고 쓰기를 직렬화 |
| D-3 | 키가 이미 있으면 값이 무엇이든 건드리지 않아요. `outputStyle` 이 다른 값이어도 `language` 가 비어 있으면 `language` 는 채워요 | 카드의 "사용자 값은 덮어쓰지 않음"; 두 키는 서로 다른 사용자 선택이에요 | 둘을 함께 충돌로 묶기 |
| D-4 | `moai update` 는 `CLAUDE_CONFIG_DIR`(없으면 `~/.claude`) 한 폴더를 다뤄요; 다른 프로필을 찾아다니지 않아요 | `host_home::resolve` 가 이미 그 폴더를 구해요; 도구별 프로필 위치는 일반 사용자 기준이 아니에요 | 후속 카드로 프로필 목록 |
| D-5 | 런처는 `language` 를 넘기지 않고 그대로 두어요 | 런처 코드에서 `language` 를 넘기는 곳을 찾지 못했어요(`src/launcher*.rs` 에서 `language: "en"` 한 곳은 시험용 기본 요청) | 후속 |
| D-6 | Codex 에는 응답 언어 설정 키가 없어서 AGENTS.md 문장만 써요 | 공식 설정 참조(learn.chatgpt.com/docs/config-file/config-reference, 2026-11-20 04:36Z 읽음)에서 `language`·`locale`·`respond in` 항목을 찾지 못했어요(요약 모델을 거친 읽기라 꼬리 부분은 확정하지 못했어요). `developer_instructions` 는 지시를 주입하지만 언어 용도로는 문서에 없어요 | 키가 생기면 그 키로 |
| D-7 | 지시 문장은 지시 언어(`en`·`ko`)로 쓰고 항목 목록 앞에 둬요 | 카드 본문의 제안; 목록만 있고 문장이 없는 것이 원인이었어요 | 목록만 두기 |

## Result

(아직 구현 전이에요.)
