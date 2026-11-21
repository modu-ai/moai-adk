---
id: SPEC-20261121-model-defaults-one-table
title: "One table of model and effort defaults for the question screen and the launcher"
status: draft
revision: 1
created: 2026-11-21
---

# One table of model and effort defaults for the question screen and the launcher

## Intent

사용자 결정(2026-11-20): 질문 화면이 "(default)"로 보여 주는 모델·역할별 값과 저장값이 없을 때 런처가 실제로 쓰는 값은 같아야 해요. 지금은 질문 화면이 `claude-opus-5-5/high` 같은 값을 보여 주는데 런처는 저장값이 없으면 `opus`/`medium`을 쓰고(Claude), Codex는 아무 값도 넘기지 않아요. 이 SPEC은 기본값을 코드의 상수 표 한 곳에 두고 두 쪽이 그 표를 읽게 해요. 질문 화면을 표 + 행 고르기로 바꾸는 일은 같은 결정의 다음 조각이고, 이 SPEC에 넣지 않아요.

기본 표(결정 그대로):

| 도구 | 리더 | 에이전트 |
| --- | --- | --- |
| Claude | claude-opus-5-5 / high | claude-sonnet-5-5 / high |
| Codex | gpt-6.2-sol / high | gpt-6.1-terra / xhigh |
| GLM (OMP도 같은 표) | zai/glm-5.3 / high | zai/glm-5.3-flash / max |

## Non-goals

- 질문 화면의 재설계(표로 보여 주기, 행을 골라 모델·효과 바꾸기)는 하지 않아요. 질문 화면 파일은 다른 카드가 쓰는 중이에요.
- 역할이 없는 세션의 `opus`/`medium` 기본값(2026-11-08 사용자 결정)은 그대로 둬요.
- `moai update`가 모델 단계를 여는 조건은 바꾸지 않아요.
- 도움말(`--model`, `--effort`) 문구는 이 SPEC의 쓰기 범위 밖이에요. 낡은 문구는 따로 보고해요.

## Requirements

- REQ-1: When the default of a tool and role is read, `model_effort_settings` shall answer from one constant table with a row for each of claude, codex, omp and glm in the leader and agent roles · failure: none — a read of a constant; an unknown tool or role answers nothing.
- REQ-2: When the default of a tool and role is read, `model_effort_settings` shall answer the model and effort of the table in the Intent, each one offered by the support table · failure: none — a read of a constant; the unit test names every row.
- REQ-3: When a session with a role starts on claude, codex, omp or glm and nothing is saved for its tool and role, the launcher shall pass the default of that table as the model and effort · failure: none — a saved entry the support table does not accept still refuses the launch as before.
- REQ-4: When a model or an effort is given on the command line, the launcher shall keep it and fill only the field that was not given from the saved choice or else the default · failure: none — no flag is dropped.
- REQ-5: While a session has no role, the launcher shall keep the model opus and the effort medium on the Claude hosts.
- REQ-6: When a saved choice for the tool and role exists, the launcher shall use it before the default · failure: none — the saved path is the one that was already tested.

## Acceptance

- AC-1 (REQ-1, REQ-2): 표가 여덟 줄이고 `default_choice` 가 줄마다 위 값을 답하며 지원 표에 있는 모델·효과임. verify: `cargo test --locked --offline --test model_effort_settings`
- AC-2 (REQ-3, REQ-4, REQ-6): 저장값이 없는 임시 HOME에서 도구 넷과 역할 둘의 `--dry-run` argv 가 표의 값을 각 호스트의 깃발로 담고, 한쪽만 준 깃발은 그 칸만 바꾸고, 저장값이 있으면 그 값이 이김. verify: `cargo test --locked --offline --test model_defaults_one_table --test launcher_reads_models`
- AC-3 (REQ-5): 역할 없는 세션의 argv 가 그대로 `opus`/`medium`. verify: `cargo test --locked --offline --test launcher_roleless`
- AC-4: 런처 시험 묶음이 표의 새 기본값으로 통과함. verify: `cargo test --locked --offline --test launcher_agent_entry --test launcher_agent_numbering --test launcher_advisor_setting --test launcher_permissions_setting --test launcher_codex_effort --test omp_launcher`
- AC-5: 가벼운 검사. verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/model_effort_settings.rs`, `src/session.rs`, `tests/model_effort_settings.rs`, `tests/model_defaults_one_table.rs`, `tests/launcher_reads_models.rs`, `tests/launcher_roleless.rs`, `tests/launcher_agent_entry.rs`, `tests/launcher_agent_numbering.rs`, `tests/launcher_advisor_setting.rs`, `tests/launcher_permissions_setting.rs`, `tests/launcher_codex_effort.rs`, `tests/omp_launcher.rs`, `.moai/specs/SPEC-20261121-model-defaults-one-table/**`
- read-only: `src/model_effort_questions.rs`, `src/update_models.rs`, `src/main.rs`
- forbidden: 키 보관소 값·실제 HOME 접근, 카드 번호를 SPEC id·브랜치·커밋에 사용, 다른 카드의 가지에 쓰기

## Plan

1. 테스트 먼저: `tests/model_defaults_one_table.rs` 에 저장값 없는 임시 HOME 에서 도구 넷×역할 둘의 argv 가 표를 따르는지 보는 시험을 두고 지금 tip 에서 실패하는지 확인(RED).
2. `model_effort_settings.rs` 에 상수 표를 두고 `default_choice` 가 읽게 함.
3. `session.rs` 의 저장값 채우기를 "저장값, 없으면 표" 로 넓힘(역할 있는 세션만).
4. 런처 시험 묶음을 새 기본값에 맞게 고침(낡은 기대를 표의 값으로).
5. AC 검사를 실행하고 Result 에 줄마다 명령과 종료 코드를 적음.

## Risks

- 역할 세션이 Claude 에서 별칭(`opus`) 대신 전체 모델 id 를 받아요. 저장값 경로가 이미 같은 id 를 넘기므로 호스트가 받는 형태는 같아요.
- Codex 역할 세션이 처음으로 `-m` 과 효과를 받아요. 사용자의 `~/.codex/config.toml` 값이 덮여요 — 이 결정의 의도예요.
- 되돌리기: 커밋 한 개 revert(상수 표와 한 함수의 폴백).

## Decisions

- 2026-11-21: 질문 화면 파일은 이 SPEC 에서 읽기만 해요 — 다른 카드가 같은 파일을 쓰는 중이에요. 되돌릴 수 있음, 분류 `none`, 결정자: agent-17.

## Result

(구현 뒤에 채워요.)
