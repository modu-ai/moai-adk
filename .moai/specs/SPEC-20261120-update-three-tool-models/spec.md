---
id: SPEC-20261120-update-three-tool-models
title: "The update model questions offer claude, gpt and glm with the researched defaults"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# The update model questions offer claude, gpt and glm with the researched defaults

## Intent

셸 `moai update` 의 모델 단계가 도구 선택에서 Claude Code·Codex·OMP 세 개를 묻고 OMP 를 고르면 glm 이 함께 저장돼요. 사용자 결정(2026-11-19, 재설계 부모 카드 노트)은 세 도구만 묻는다고 정했어요: claude, gpt(codex), glm. 기본값도 정해졌고, 정확한 id 와 노력 값은 공식 문서 조사(카드 리서치 노트, 2026-11-20)로 확정했어요 — 문서에 gpt-6.2-terra 는 없어 유일한 terra id gpt-6.1-terra 를 쓰고, 가장 새로운 sol 은 gpt-6.2-sol 이에요.

바뀌는 것: 도구 선택 세 개(표시 Claude·GPT (Codex)·GLM, 내부 키는 codex 그대로), 기본값 표(클로드 리더 노력 medium→high, GPT 리더 모델 gpt-6.1-sol→gpt-6.2-sol), 각 선택이 자기 도구만 쓰기. 저장은 기존 launcher.yaml receipt·models-revert 그대로, 표의 omp 행과 저장된 omp 항목은 그대로 유효해요(moai omp 런처 호환). 배포 문구·선택지에 개발 도구 이름·카드 번호·Issue 번호를 쓰지 않아요.

## Non-goals

- 런처의 저장값 읽기(이미 세 런처가 읽어요)와 저장 형식: 안 바꿔요.
- omp 행 제거: 저장된 omp 항목 호환을 위해 표에 남아요.
- 질문 문구의 재설계 전반(보고서의 다른 단계): 이 조각은 모델 단계의 도구 세트·기본값·표시만 다뤄요.
- Out of scope: 실제 HOME, 네트워크, 카드 번호·Issue 번호의 출력 사용.

## Requirements

- REQ-1: When the model stage asks the tool pick on a terminal, the offered tools shall be exactly three — Claude, GPT (Codex) and GLM — with each pick writing the choice of that tool alone. · failure: none — the pick question is the outcome.
- REQ-2: When no saved choice exists for a tool and role, the offered default shall be the researched table: claude leader claude-opus-5-5 with high, claude agent claude-sonnet-5-5 with high, gpt leader gpt-6.2-sol with high, gpt agent gpt-6.1-terra with xhigh, glm leader zai/glm-5.3 with high, glm agent zai/glm-5.3-flash with max. · failure: none — the preview shows the default before anything is written.
- REQ-3: While a saved file holds entries, they shall reload as before, omp entries included, with the support table still accepting them. · failure: none — an entry outside the table is refused as before.
- REQ-4: The effort list a model offers shall contain only the values the official documents list for it (the research note's mapping: claude low·medium·high·xhigh·max, gpt low·medium·high·xhigh·max, glm low·high·max). · failure: none — the lists are the table's rows.

## Acceptance

- AC-1 (REQ-1): the pick question offers the three tools with those labels and no fourth; each pick writes only its own tool's questions.
  verify: `cargo test --locked --offline --test model_effort_questions`
  verify: `cargo test --locked --offline --test update_model_pick`
- AC-2 (REQ-2): with an empty HOME the preview and the report name the researched defaults for every tool and role.
  verify: `cargo test --locked --offline --test update_model_pick`
  verify: `cargo test --locked --offline --test model_effort_settings`
- AC-3 (REQ-3): a saved file with claude, codex, glm and omp entries reloads and the omp entry stays supported.
  verify: `cargo test --locked --offline --test model_effort_settings`
- AC-4 (REQ-4): every model's effort options in the questions match the mapping rows.
  verify: `cargo test --locked --offline --test model_effort_questions`
- AC-5: the neighbour update and launcher checks and the style checks pass.
  verify: `cargo test --locked --offline --test launcher_reads_models`
  verify: `cargo test --locked --offline --test launcher_settings_receipt`
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261120-update-three-tool-models/spec.md`, `src/model_effort_questions.rs`, `src/model_effort_settings.rs`, `src/update_models.rs`, `tests/model_effort_questions.rs`, `tests/model_effort_settings.rs`, `tests/update_model_pick.rs`
- read-only: `src/session.rs`, `src/ops/update.rs`
- forbidden: 실제 HOME, 저장 형식과 receipt·revert 경로, 카드 번호·Issue 번호의 출력 사용

## Plan

1. RED: 세 시험 파일의 기대값을 새 도구 세트·표시·기본값으로 바꿔 지금 코드로 깨지는 것을 봐요.
2. GREEN: TOOL_CHOICES 와 표시를 세 개로, 쓰기 매핑을 1:1 로, default_choice 두 칸(클로드 리더 노력, GPT 리더 모델)을 고쳐요.
3. 이웃 시험·여섯 가벼운 검사·변이(기본값 한 칸씩, 도구 세트, 쓰기 매핑)·감사·READY.

## Risks

- 저장된 gpt-6.1-sol 리더 항목은 그대로 유효해요(표에 남아요) — 기본값만 바뀌어요. 새 기본 gpt-6.2-sol 은 문서와 로컬 카탈로그 둘 다 확인했어요.
- 표시 문구 "GPT (Codex)" 가 도구 이름 정책과 어긋나지 않은지: 사용자 결정문의 표기 그대로예요.
