---
id: SPEC-20261123-glm-54-followups
title: "GLM 5.4 follow-ups: the moai update GLM model table and default, the built-in template wording and the status line fable slot"
status: draft
revision: 1
created: 2026-11-23
---

# GLM 5.4 follow-ups: the moai update GLM model table and default, the built-in template wording and the status line fable slot

## Intent

사용자 지시(2026-11-23): "glm 5.4 가 오늘 릴리스되었다, glm-5.4[1m] 모델명으로 지금 glm-5.3 설정을 변경해라." 앞 SPEC(`SPEC-20261123-glm-54-model`)은 `moai glm` 실행기의 별칭 환경(`GLM_ENV`)과 상태 줄의 이름 줄이기·창 표를 바꿨고, 독립 감사(AUDIT #393)가 비차단 후속 셋을 남겼어요. 이 SPEC 이 그 셋을 닫아요.

- N1: `moai update` 가 저장하는 GLM 모델 표(`src/model_effort_settings.rs`)가 `zai/glm-5.3`·`zai/glm-5.3-flash` 만이라, 저장한 사용자는 `--model glm-5.3` 으로 떠서 `glm-5.4[1m]` 를 못 써요(저장된 모델이 별칭 환경을 건너뛰어요). 표와 GLM leader 기본값에 `glm-5.4[1m]` 를 넣어요.
- N2: 내장 템플릿 `templates/bundle-packs/.claude/agents/moai/super-advisor.md:76` 이 GLM 추론 모델을 `(glm-5.3)` 으로 말해요.
- N4: 상태 줄의 `GLM_ENV_SLOTS`(창 크기 풀이)와 `resolve_glm_model_name`(이름 풀이)에 fable 칸이 없어서, 실행기가 `ANTHROPIC_DEFAULT_FABLE_MODEL` 로 준 값을 보지 못해요.

N3(앞 SPEC 이 draft 이고 Result 가 없음)은 앞 SPEC 의 몫이라 이 SPEC 의 대상이 아니에요. N5(z.ai 가 `glm-5.4[1m]` 를 받는지)는 실제 세션 없이 확인할 수 없어요(Risks).

## Decisions

- D-1 저장 id 의 형태: `zai/glm-5.4[1m]`. 선택지: (a) `zai/glm-5.4[1m]`, (b) `zai/glm-5.4`, (c) 둘 다 표에 둠. 기본값: (a). 이유: 표의 id 는 `zai/` 제공자 접두사를 달고 있고(`src/session.rs` 의 `apply_saved_choice` 가 `strip_prefix("zai/")` 로 떼요), 떼고 나면 `glm-5.4[1m]` 가 되어 `GLM_ENV` 의 opus·sonnet·fable 별칭 값과 글자까지 같아요. 사용자도 모델 이름을 `glm-5.4[1m]` 로 말했어요. 되돌릴 수 있어요(표 한 줄). 범주 none, 결정자 이 레인(카드 C-396 범위 안의 되돌릴 수 있는 선택), 기한 없음.
- D-2 이전 값은 표에 남겨요: `zai/glm-5.3` 과 `zai/glm-5.3-flash` 는 표에 그대로 있어서 이미 저장된 값이 계속 읽히고(지원되지 않는 항목으로 막히지 않음) 실행기에는 `--model glm-5.3` 으로 가요. 모델 질문에서는 기본값 다음에 나와요.
- D-3 GLM agent 기본값은 `zai/glm-5.3-flash` / `max` 그대로: haiku 별칭이 `glm-5.3-flash` 로 남은 앞 SPEC 의 결정(5.4 flash 는 확인되지 않음)과 같아요. GLM leader 기본값만 `zai/glm-5.4[1m]` / `high` 로 바뀌어요.
- D-4 `zai/glm-5.4[1m]` 의 노력 값: 5.3 과 같은 `low`·`high`·`max` 셋. 5.4 가 다른 노력 값을 주는지 말하는 공식 문서를 이 세션에서 읽지 못했어요(unverified, z.ai 를 부르지 않음). 노력 값 표는 한 줄 바꿈으로 늘릴 수 있어요.
- D-5 상태 줄의 fable 칸: 실행기가 fable 별칭도 `glm-5.4[1m]` 로 주므로, 상태 줄의 이름 풀이(표시 이름에 `fable` 이 든 경우)와 창 크기 풀이 칸에 `ANTHROPIC_DEFAULT_FABLE_MODEL` 을 더해요. 순서는 opus, sonnet, haiku 뒤 fable 이에요.

## Non-goals

- 실행기 환경(`GLM_ENV`), 실행기 안내 문구, 상태 줄의 이름 줄이기와 창 표(앞 SPEC 에서 끝남).
- 키·자격 증명, 실제 `~/.moai`·`~/.claude` 에 쓰는 일, z.ai 호출, 실제 HOME 설치, Go 도구.
- haiku 의 5.4 flash 여부, 5.4 가 주는 노력 값의 공식 확인.
- 지난 SPEC 문서에 적힌 `glm-5.3` 기록.

## Requirements

- REQ-1: The GLM support table shall list zai/glm-5.4[1m] with the low, high and max efforts.
- REQ-2: The GLM support table shall keep zai/glm-5.3 and zai/glm-5.3-flash.
- REQ-3: The glm leader default shall be zai/glm-5.4[1m] with the high effort.
- REQ-4: The glm agent default shall stay zai/glm-5.3-flash with the max effort.
- REQ-5: The glm launcher shall pass a saved zai/glm-5.4[1m] to the host as the model glm-5.4[1m].
- REQ-6: The glm launcher shall keep passing a saved zai/glm-5.3 to the host as the model glm-5.3.
- REQ-7: The super-advisor template shall name glm-5.4 as the GLM carve-out reasoning model.
- REQ-8: The status line shall show the model of ANTHROPIC_DEFAULT_FABLE_MODEL, without the [1m] marker, for a fable slot name.
- REQ-9: The status line shall count the model of ANTHROPIC_DEFAULT_FABLE_MODEL for the context window.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 지원 표가 `zai/glm-5.4[1m]`(노력 low·high·max)·`zai/glm-5.3`·`zai/glm-5.3-flash` 를 갖고, 기본값이 GLM leader `zai/glm-5.4[1m]`/`high`, GLM agent `zai/glm-5.3-flash`/`max` 이며, `moai update` 의 GLM leader 모델 질문이 `zai/glm-5.4[1m]` 를 첫 선택지(기본)로 내놓고 이미 저장된 `zai/glm-5.3` 은 지원 항목으로 읽혀 질문의 기본값이 됨 · verify: `cargo nextest run --locked --offline --test model_effort_settings --test model_effort_questions --test update_model_pick`
- AC-2 (REQ-5, REQ-6): 저장된 `zai/glm-5.4[1m]` 가 `moai glm -l --dry-run` 의 argv 에서 `--model glm-5.4[1m]` 로 나오고, 저장된 `zai/glm-5.3` 은 `--model glm-5.3` 으로 나오며, `moai update` 의 모델 단계가 GLM leader 기본값을 `zai/glm-5.4[1m]` 로 저장함; 임시 HOME 만 쓰고 z.ai 를 부르지 않음 · verify: `cargo nextest run --locked --offline --test launcher_reads_models --test update_model_pick`
- AC-3 (REQ-7): `super-advisor.md` 의 GLM 문구가 `GLM_ENV` 의 opus 별칭 값(`[1m]` 뗀 것)을 말하고 `glm-5.3` 이 없음; 템플릿 해시 줄(`templates/bundle-packs/packs.json`)이 새 파일 해시와 같음 · verify: `cargo nextest run --locked --offline --lib -E 'test(session::)'` 와 `shasum -a 256 templates/bundle-packs/.claude/agents/moai/super-advisor.md` 가 `packs.json` 의 `super-advisor` 줄과 같음
- AC-4 (REQ-8, REQ-9): 상태 줄이 fable 칸의 `glm-5.4[1m]` 를 이름 줄에서 `glm-5.4` 로 보이고, fable 칸만 있을 때 1M 창으로 셈; Claude 이름과 다른 칸은 그대로 · verify: `cargo nextest run --locked --offline --lib -E 'test(statusline::)'`

## Scope

- write: `.moai/specs/SPEC-20261123-glm-54-followups/**`, `src/model_effort_settings.rs`, `src/session.rs`, `src/statusline/state.rs`, `templates/bundle-packs/.claude/agents/moai/super-advisor.md`, `templates/bundle-packs/packs.json`(그 템플릿의 `sha256` 한 줄만; 카드 범위를 `moai card scope` 로 넓혔어요), `tests/model_effort_settings.rs`, `tests/model_effort_questions.rs`, `tests/update_model_pick.rs`, `tests/launcher_reads_models.rs`
- read-only: `src/model_effort_questions.rs`, `src/update_models.rs`, 다른 테스트
- forbidden: 그 밖의 경로, 실제 HOME, 키 값, z.ai 호출

## Plan

1. SPEC (이 파일).
2. RED: 지원 표·기본값·모델 질문·저장된 5.4/5.3 의 실행기 argv·`moai update` 단계·템플릿 문구·상태 줄 fable 칸 테스트를 쓰고, 바뀌기 전 코드에서 실패하는 것을 기록.
3. GREEN: `src/model_effort_settings.rs` 표와 기본값, `src/statusline/state.rs` 의 fable 칸, `super-advisor.md` 문구와 `packs.json` 해시 줄.
4. 확인: 바꾼 테스트 대상(nextest 필터), fmt, `git diff --check`, `moai trailer-check`, `moai spec check`, `check affected` 한 번, 임시 HOME·더미 키의 실행 미리보기와 `moai update` 단계.

## Risks

- N5(미확인): z.ai 가 `glm-5.4[1m]` 라는 모델 이름을 실제로 받는지는 실제 세션 없이 확인할 수 없어요. 이 SPEC 은 앞 SPEC 이 정한 별칭 값과 같은 문자열을 저장 id 로 쓸 뿐이라 새 위험을 더하지 않고, 실제 호출은 하지 않아요.
- 저장된 `zai/glm-5.3` 을 모델로 고른 사용자는 별칭 환경(`glm-5.4[1m]`)을 건너뛰고 `--model glm-5.3` 으로 떠요. 이전 동작 그대로이고(AC-2), 사용자가 `moai update` 에서 5.4 를 다시 고르면 바뀌어요.
- `packs.json` 의 해시 줄이 템플릿 파일과 어긋나면 카탈로그와 파일이 달라져요: 해시를 `shasum -a 256` 으로 다시 재서 한 줄만 바꿔요.
