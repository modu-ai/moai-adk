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

## Result

(status draft — 승인은 사용자 몫이라 상태는 draft 로 둬요. 아래 수치는 이 가지에서 직접 돌린 명령의 출력이고, `check affected` 영수증은 이 Result 가 든 헤드를 이름 짓는 READY 노트에 있어요.)

- AC-1: met — `the_glm_table_lists_54_and_keeps_the_53_rows`(표에 `zai/glm-5.4[1m]`·`zai/glm-5.3`·`zai/glm-5.3-flash`, 5.4 노력 low·high·max), `no_saved_value_answers_the_default_for_every_tool_and_role`(GLM leader `zai/glm-5.4[1m]`/`high`, agent `zai/glm-5.3-flash`/`max`), `model_and_effort_questions_offer_only_table_values_with_the_default_first`, `the_glm_model_question_offers_54_first_and_keeps_a_saved_53`(질문 순서 5.4·5.3·flash, 저장된 5.3 은 기본 선택지), `a_glm_53_choice_saved_before_54_is_still_supported`(저장된 `zai/glm-5.3` 은 지원 항목, 저장된 것으로 읽힘)
- AC-2: met — `a_saved_glm_54_reaches_the_argv_and_a_saved_53_still_does`(저장된 5.4 → `--model glm-5.4[1m]`, 저장된 5.3 → `--model glm-5.3`, 저장값 노트가 키 경로만 말함), `enter_saves_glm_54_for_the_leader_and_flash_for_the_agent`(모델 단계가 leader `zai/glm-5.4[1m]`·agent flash 를 저장하고 보고가 saved 로 읽음), `a_saved_glm_53_leader_stays_through_enter_everywhere`(저장된 5.3 은 미지원 목록에 안 뜨고 Enter 로 저장해도 그대로); 임시 HOME 과 `--dry-run` 만 썼고 z.ai 를 부르지 않았어요
- AC-3: met — `the_super_advisor_template_names_the_launchers_glm_model`(템플릿이 `GLM_ENV` opus 값에서 `[1m]` 뗀 `glm-5.4` 를 말하고 `glm-5.3` 이 없으며, `packs.json` 의 `super-advisor` 줄 해시가 파일의 SHA-256 `10636b55…30ed` 와 같음)
- AC-4: met — `statusline::state::tests::glm_fable_slot_names_and_sizes_its_model`(`GLM_ENV` 그대로 만든 환경에서 `Fable`·`Fable[1m]` → `glm-5.4`, Haiku → `glm-5.3-flash`, claude 가 든 fable 칸은 `Fable` 그대로, fable 칸만 있어도 1M 창, claude 만 있으면 0)

증거 (이 가지에서 직접 돌린 것):

- 먼저 실패함: RED 커밋 `e89a05b459` 의 코드(바꾸기 전 `src/`)에서 `cargo nextest run --locked --offline --no-fail-fast --lib --test model_effort_settings --test model_effort_questions --test update_model_pick --test launcher_reads_models -E <그 4개 바이너리 + 새 lib 테스트>` → exit 100, 58 run, 47 passed, 11 failed(`target/glm54/red.log`). 실패 11개는 새 테스트 7개와 기본값을 바꾼 기존 테스트 4개예요. `a_saved_glm_53_leader_stays_through_enter_everywhere` 는 5.3 유지를 지키는 테스트라 처음부터 통과해요.
- 통과: 바꾼 뒤 같은 명령 → exit 0, 58 run, 58 passed(`target/glm54/green.log`).
- 뮤턴트 6개가 모두 죽었어요(각각 하나만 바꾸고 해당 테스트를 돌린 뒤 되돌려서 트리가 깨끗한 것을 확인; `target/glm54/M1..M6.log`): M1 leader 기본값을 `zai/glm-5.3` 으로 → 6개 실패(`no_saved_value_answers…`, `enter_saves_glm_54…` 등), M2 표의 5.4 행 id 를 바꿈 → 8개 실패(`a_saved_glm_54_reaches_the_argv…`, `the_glm_table_lists_54…` 등), M3 이름 풀이의 `fable` 분기 제거 → `glm_fable_slot_names_and_sizes_its_model` 실패, M4 `GLM_ENV_SLOTS` 의 fable 칸 제거 → 같은 테스트 실패, M5 템플릿 문구를 `glm-5.3` 으로 되돌림 → 템플릿 테스트 실패, M6 `packs.json` 해시를 옛 값으로 → 같은 테스트 실패.
- `cargo fmt --check` 가 새 테스트 줄 몇 군데를 고치라고 해서 `cargo fmt` 로 그 파일들만 고쳤어요(다른 파일은 안 바뀜).

결정 (D-1~D-5 는 위 Decisions): 저장 id `zai/glm-5.4[1m]`, 5.3 행 유지, GLM agent 기본값 flash 유지, 5.4 노력 low·high·max(미확인), 상태 줄 fable 칸 추가.

범위: 카드 범위에 `templates/bundle-packs/packs.json` 이 없어서, 템플릿 문구를 바꾸면 그 파일의 해시 줄이 어긋나므로 `moai card scope`(held 카드 넓히기, 이유 기록)로 그 한 줄만 쓰려고 넓혔어요. 카드 범위에 있던 `skills/moai/references/source-manifest.json` 은 `skills/moai/` 아래 파일만 덮고 템플릿 해시를 갖지 않아 바꾸지 않았어요.

못 한 것 (not_run):

- 임시 HOME·더미 키로 손으로 띄운 `moai glm` 실행: 앞 카드에서 워크트리 격리 가드가 HOME 을 바꾸는 명령을 거부했어요. 같은 일을 하는 통합 테스트(임시 HOME, `--dry-run`)가 증거예요. 저장된 5.4 로 대역 host 를 실제로 띄우는 시험은 추가하지 않았고(저장된 5.3 이 띄운 host 로 가는 기존 시험은 그대로 통과), argv 는 `--dry-run` 이 같은 합성 경로로 보여줘요.
- 실제 `~/.moai`·`~/.claude` 에 쓰기, z.ai 호출(N5: z.ai 가 `glm-5.4[1m]` 를 받는지), Linux·Windows, 독립 감사(작성자가 아닌 세션 몫).

후속 (이 SPEC 이 만지지 않은 것):

- 5.4 가 다른 노력 값(예: medium)을 주는지 공식 문서로 확인하면 표 한 줄을 늘려요(D-4).
- 5.4 flash 가 확인되면 haiku 별칭·GLM agent 기본값을 따로 바꿔요(D-3).
