---
id: SPEC-20261123-glm-54-model
title: "The glm launcher maps the main Claude aliases to glm-5.4[1m]"
status: draft
revision: 1
created: 2026-11-23
---

# The glm launcher maps the main Claude aliases to glm-5.4[1m]

## Intent

사용자 지시(2026-11-23): "glm 5.4 가 오늘 릴리스되었다, glm-5.4[1m] 모델명으로 지금 glm-5.3 설정을 변경해라." 지금 GLM 세션의 모델은 `moai glm` 실행기가 host 프로세스에 주는 환경 `src/session.rs` 의 `GLM_ENV` 가 정해요: opus·sonnet·fable 별칭이 `glm-5.3`, haiku 별칭이 `glm-5.3-flash` 예요. 이 SPEC 은 opus·sonnet·fable 별칭을 `glm-5.4[1m]` 로 바꾸고, 실행기 안내 문구, 상태 줄의 GLM 모델 표와 이름 줄이기, 그 값을 고정하는 테스트, 현재 값을 적은 문서를 같이 맞춰요. haiku 는 사용자가 말하지 않았고 5.4 flash 도 확인되지 않아 `glm-5.3-flash` 그대로 둬요.

## Non-goals

- `src/model_effort_settings.rs` 의 `zai/glm-5.3`, `zai/glm-5.3-flash` 지원표 id: `moai update` 가 저장하는 다른 형식의 id 라 따로 확인해요.
- 키·자격 증명, 실제 HOME 에 설치하는 일, Go 도구.
- 1M 창 선언 값(`CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`)과 endpoint·timeout 값.
- 지난 SPEC 문서에 적힌 `glm-5.3` 기록.

## Requirements

- REQ-1: The glm launcher shall map the opus, sonnet and fable aliases to glm-5.4[1m].
- REQ-2: The glm launcher shall keep the haiku alias at glm-5.3-flash.
- REQ-3: The glm launcher notes shall name glm-5.4[1m] for opus, sonnet and fable and glm-5.3-flash for haiku.
- REQ-4: The status line shall show a GLM slot model glm-5.4[1m] as glm-5.4.
- REQ-5: The status line shall count a glm-5.4 slot model with a context window of 1000000 tokens.
- REQ-6: The documents that state the current glm launcher value shall name glm-5.4[1m] and not glm-5.3 for the main aliases.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `moai glm --dry-run` 의 notes 가 opus·sonnet·fable 을 `glm-5.4[1m]`, haiku 를 `glm-5.3-flash` 로 말하고, 띄운 가짜 `claude` 가 받은 환경이 같은 값을 가짐; 테스트는 바꾸기 전에 실패(RED)하고 바꾼 뒤 통과 · verify: `cargo nextest run --locked --offline --test launcher_reads_models --test launcher_glm_key_file`
- AC-2 (REQ-4, REQ-5): 상태 줄이 `glm-5.4[1m]` 를 `glm-5.4` 로 줄이고, opus 슬롯에 `glm-5.4[1m]` 만 있을 때 1M 창으로 셈 · verify: `cargo nextest run --locked --offline --lib -E 'test(statusline::)'`
- AC-3 (REQ-3, REQ-6): 실행기 안내 문구와 `docs/rust-product-path.md`, `docs/user-entry-draft.md` 가 새 값을 말하고, `glm-5.3` 을 opus·sonnet·fable 의 현재 값으로 말하는 곳이 없음 · verify: `rg -n "glm-5\.3" src/session.rs src/launcher_terminal.rs src/statusline docs/rust-product-path.md docs/user-entry-draft.md` 에 haiku 의 `glm-5.3-flash` 와 옛 값 표만 남음

## Scope

- write: `.moai/specs/SPEC-20261123-glm-54-model/**`, `src/session.rs`, `src/launcher_terminal.rs`, `src/statusline/state.rs`, `src/statusline/segments.rs`, `tests/launcher_reads_models.rs`, `tests/launcher_glm_key_file.rs`, `docs/rust-product-path.md`, `docs/user-entry-draft.md`
- read-only: `src/model_effort_settings.rs`, 다른 테스트
- forbidden: 그 밖의 경로, 실제 HOME, 키 값

## Plan

1. SPEC (이 파일).
2. RED: 실행기 테스트와 상태 줄 테스트가 `glm-5.4[1m]` 를 기대하게 바꾸고, 바뀌기 전 코드에서 실패하는 것을 기록.
3. GREEN: `GLM_ENV`, 안내 문구, 상태 줄 표와 이름 풀이, 문서.
4. 확인: 바꾼 테스트 대상, fmt, `git diff --check`, `moai trailer-check`, `check affected` 한 번, 가짜 HOME·가짜 키로 실행 미리보기.

## Risks

- Claude Code 가 `[1m]` 접미사를 1M 창 표시로 읽어요: 별칭 값에 접미사를 붙여도 1M 창 선언 환경 값은 그대로라 창 크기는 바뀌지 않아요.
- 상태 줄 이름 풀이가 환경 값을 그대로 보이면 `glm-5.4[1m]` 가 화면에 나와요: 풀이에서 접미사를 떼고 테스트로 고정해요.
