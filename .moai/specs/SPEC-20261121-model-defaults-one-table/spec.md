---
id: SPEC-20261121-model-defaults-one-table
title: "One table of model and effort defaults: the launcher starts a role session with the table the question screen shows"
status: draft
revision: 3
created: 2026-11-21
---

# One table of model and effort defaults: the launcher starts a role session with the table the question screen shows

## Intent

사용자 결정(2026-11-20, G2): 질문 화면이 "(default)" 로 보여 주는 도구·역할별 모델과 effort, 그리고 저장값이 없을 때 런처가 실제로 쓰는 값은 같아야 해요. 지금 NeXT 에서는 상수 표(`default_choice`)가 질문 화면의 "(default)" 를 정하지만 런처는 저장값이 없으면 `opus`/`medium` 을 채워서(`src/session.rs` 의 `DEFAULT_MODEL`·`DEFAULT_EFFORT`) 둘이 달라요. 이 SPEC 의 조각 A 는 런처가 같은 표를 읽게 해서 그 차이를 닫아요.

기본 표는 지금의 `default_choice` 4줄 그대로예요(새 값 없음). 2026-11-20 결정의 Codex·OMP 줄은 철회됐고(Codex 지원 제거, `SPEC-20261121-codex-launcher-and-models-removed`) GLM leader 는 `SPEC-20261123-glm-54-followups` D-3 이 5.4 로 바꿨어요.

| 도구 | 리더 | 에이전트 |
| --- | --- | --- |
| Claude | claude-opus-5-5 / high | claude-sonnet-5-5 / high |
| GLM | zai/glm-5.4[1m] / high | zai/glm-5.3-flash / max |

조각 B(판 3, 2026-11-26): 조각 A 가 NeXT 에 들어간(`db58fff976`) 뒤 `SPEC-20261124-launcher-retire-role-flags` 가 질문 화면을 호스트마다 모델 하나와 effort 하나를 묻는 모양으로 바꿨고(D-2, `384d9bc45c`), 저장값이 없는 평범한 `moai cc|glm` 런처는 `launcher_default(host)`(`opus`/`medium`, glm 은 `opus`/`high`)를 써요. 그래서 2026-11-20 결정의 "표로 보여 주고 행을 골라 바꾸기" 절반은 그 SPEC 의 한 호스트 한 항목 모양으로 대신됐고, 남은 틈은 하나예요: 질문 화면의 "(default)" 는 아직 `default_choice(tool, "leader")` 행(Claude 는 `claude-opus-5-5`/`high`)이라 평범한 런처가 실제로 쓰는 `medium` 과 달라요(GLM 은 `glm-5.4[1m]`/`high` 로 뜻이 같아요). 판 3 은 그 틈 하나만 닫아요: 화면의 기본값이 평범한 런처의 기본값에서 와요.

## Decisions

- D-1 새로 해요: 옛 조각(가지 `spec/model-defaults-one-table-20261121`)은 Codex·OMP 8줄 표를 전제로 해서 지금 팁과 충돌 9곳이 나요. 옛 가지는 보존하고 참고만 해요. 결정자: 리더(2026-11-24, 사용자 위임 2026-11-23), 되돌릴 수 있음, 분류 none.
- D-2 표는 지금 `default_choice` 4줄이에요: 새 값을 만들지 않아요. 결정자: 리더(2026-11-24).
- D-3 역할 세션은 저장값 경로와 같은 전체 id 를 호스트에 넘겨요(예: `claude-opus-5-5`, GLM 은 `zai/` 를 뗀 `glm-5.4[1m]`). 별칭(`opus`)이 아니에요. 되돌리기는 커밋 revert 하나예요. 결정자: 리더(2026-11-24).
- D-4 역할 없는 세션은 `opus`/`medium` 그대로예요(사용자 결정 2026-11-08). 결정자: 사용자.
- D-5 조각 A 를 먼저 READY·감사·착지해요. 조각 B 는 C-248 착지 뒤 그 위에서 해요. 결정자: 리더(2026-11-24).
- D-6 (판 3, 제안, 기본값 A) 질문 화면의 "(default)" 와 평범한 런처의 저장값 없는 기본값이 갈릴 때 누가 이기나요? A: 런처 값이 이겨요. 화면의 기본 effort 는 `launcher_default(host).effort` 에서 오고(Claude `medium`, GLM `high`), 기본 모델은 별칭 `opus` 가 가리키는 표의 id(`default_choice(tool, "leader").model`)예요. B: 화면 값이 이겨서 런처 기본을 `claude-opus-5-5`/`high` 로 올려요. 이유: A 는 실행 중인 동작을 바꾸지 않고(런처 기본은 사용자 결정 2026-11-08 과 2026-11-24 의 값), 화면은 읽는 쪽이라 고치기 쉬워요. B 는 `cc` 기본 effort 를 `medium` 에서 `high` 로 바꿔서 그 결정을 뒤집어요(노트 4672 의 `claude-opus-5-5`/`high` 는 2026-11-20 값이고 둘이 어긋나는 곳이 이 한 칸이에요). 되돌리기: 커밋 revert 하나.
- D-7 (판 3, 제안, 기본값 A) 역할 행(`default_choice` 의 leader·agent)은 그대로 둬요. `-a` 세션과 레거시 `session --role` 이 읽고, 화면은 더는 leader 행의 effort 를 기본으로 쓰지 않아요. 별칭 `opus` 가 leader 행의 모델 id 를 가리킨다는 전제는 호스트 쪽 매핑이라 이 저장소의 시험이 못 보여 줘요: 위험 절에 적고, 시험은 두 값(`launcher_default` 의 모델 별칭과 표의 id)이 한 함수에서 나오는지만 지켜요.

## Interview

- Q1: 옛 조각(가지 `spec/model-defaults-one-table-20261121`)을 합치나요, 새로 하나요? → A: 새로 해요(D-1). 옛 가지는 Codex·OMP 전제라 충돌 9곳이고 보존만 해요.
- Q2: 표를 4줄로 줄여도 되나요? → A: 예. 지금 `default_choice` 4줄이 표예요(D-2). 2026-11-20 결정의 Codex·OMP 줄은 철회됐고 GLM leader 는 5.4 로 바뀌었어요.
- Q3: 조각 B(질문 화면 재설계)를 언제 하나요? → A: C-248 이 착지한 뒤 그 위에서 해요. 이 카드는 B 기준으로 열린 채 남아요(D-5).
- Q4: 조각 A 를 먼저 감사·착지하나요? → A: 예(D-5).
- Q5: 역할 세션이 별칭 대신 전체 id 를 넘겨도 되나요? → A: 예(D-3). 역할 없는 세션은 `opus`/`medium` 그대로예요(D-4).
- Q6: 도움말 두 줄을 고치면 `tests/help_layout.rs` 가 읽는 골든 파일 네 개가 낡아지는데(레인 노트 8216), 어떻게 하나요? → A: 판 2 로 가요. 골든 네 파일만 Scope 의 write 에 더하고 요구·인수는 그대로예요(리더 답, 2026-11-24).

- Q7: (판 3) 조각 B 는 아직 "표로 보여 주고 행을 골라 바꾸기" 인가요? → A: 아니에요. 그 화면은 `SPEC-20261124-launcher-retire-role-flags` D-2 가 호스트 하나에 항목 하나로 바꿨고, 남은 일은 화면의 "(default)" 가 평범한 런처가 쓰는 값과 같게 하는 것 하나예요(D-6).
- Q8: (판 3) 화면의 기본이 런처와 다를 때 누구 값이 이기나요? → A: 런처 값이에요(D-6 기본값 A). 리더가 B 를 고르면 이 개정의 REQ-10·REQ-11 이 거꾸로 돼요.

## Non-goals

- 질문 화면의 재설계(표로 보여 주기, 행을 골라 모델 → effort 바꾸기): 대신된 일이에요(Q7). 판 3 은 화면의 기본값 출처만 바꾸고 질문의 개수·순서·모양은 안 바꿔요.
- 표의 값 바꾸기, 새 도구·역할 더하기, `moai update` 가 모델 단계를 여는 조건.
- 키·자격 증명, 실제 `~/.moai`·`~/.claude` 에 쓰기, 호스트 실행(`--dry-run` 만 써요).

## Requirements

- REQ-1: The default table shall name a model and an effort offered by the support table for each of claude and glm in the leader and agent roles.
- REQ-2: The launcher shall start a claude or glm session with a role with the model and effort of the default table when nothing is saved for its tool and role.
- REQ-3: The launcher shall pass the claude default as the full model id and the glm default as the model id without the zai/ prefix.
- REQ-4: The launcher shall use the default table also when the user launcher settings file does not exist.
- REQ-5: The launcher shall use a saved choice for the tool and role before the default table.
- REQ-6: The launcher shall keep a model or an effort given on the command line and fill only the field that was not given.
- REQ-7: The launcher shall keep the model opus and the effort medium for a session without a role.
- REQ-8: The launcher shall refuse a launch whose saved model settings are not in the support table, as it does now.
- REQ-9: The help text of the model and effort options shall not say that a role session defaults to opus and medium.
- REQ-10: The model question of `moai update` shall mark as its default the table model id that the launch alias of the tool names.
- REQ-11: The effort question of `moai update` shall mark as its default the effort that a launch of the tool runs with nothing saved, when the chosen model offers that effort.
- REQ-12: The launch default and the default of the questions shall come from one function of `src/model_effort_settings.rs`.

## Acceptance

- AC-1 (REQ-1): 표가 4줄이고 `default_choice` 가 줄마다 위 값을 답하며 그 값이 지원 표에 있음; 모르는 도구·역할은 아무것도 답하지 않음 · verify: `cargo nextest run --locked --offline --test model_effort_settings`
- AC-2 (REQ-2, REQ-3, REQ-4): 저장값 없는 임시 HOME(설정 파일 자체가 없는 경우 포함)에서 도구 둘×역할 둘의 `--dry-run` argv 가 표의 모델과 effort 를 각 호스트 깃발로 담고 Claude 는 전체 id, GLM 은 `zai/` 를 뗀 id 임; 그리고 런처 기본값이 표와 같다는 시험이 표의 줄마다 있음 · verify: `cargo nextest run --locked --offline --test model_defaults_one_table`
- AC-3 (REQ-5, REQ-6, REQ-8): 저장값이 있으면 그 값이 이기고, 한쪽만 준 깃발은 그 칸만 바꾸고, 지원 표 밖의 저장값은 전과 같이 거절됨 · verify: `cargo nextest run --locked --offline --test launcher_reads_models --test model_defaults_one_table`
- AC-4 (REQ-7): 역할 없는 세션의 argv 가 그대로 `opus`/`medium` · verify: `cargo nextest run --locked --offline --test launcher_roleless`
- AC-5: 기본값이 바뀌어 낡아진 시험이 기대를 지우지 않고 표의 값으로 바뀐 채 통과함 · verify: `cargo nextest run --locked --offline --test launcher_agent_entry --test launcher_agent_numbering --test launcher_advisor_setting --test launcher_permissions_setting --test product_cli`
- AC-6 (REQ-9): `--model`·`--effort` 도움말이 역할 세션의 기본값을 `opus`/`medium` 이라고 말하지 않음 · verify: `cargo nextest run --locked --offline --test model_defaults_one_table` 의 도움말 시험과 `moai --help` 출력을 읽음
- AC-7: 가벼운 검사 · verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `moai spec check`; `moai trailer-check`
- AC-8 (REQ-10, REQ-11, REQ-12): `claude` 의 모델 질문은 `claude-opus-5-5` 를, 노력 질문은 `medium` 을, `glm` 은 `zai/glm-5.4[1m]` 과 `high` 를 기본으로 표시하고, 그 값이 `launcher_default(host).effort` 와 `default_choice(tool, "leader").model` 에서 오며, 저장된 선택과 모델이 노력을 안 줄 때의 되돌림은 전과 같음 · verify: `cargo nextest run --locked --offline --test model_effort_questions`
- AC-9 (REQ-10, REQ-11): 저장값 없는 임시 HOME 에서 `moai update` 의 모델 단계에 Enter 만 답하면 저장되는 모델·effort 가, 같은 임시 HOME 에서 저장값 없이 `cc --dry-run`·`glm --dry-run` 이 담는 effort 와 같고 모델은 그 별칭 `opus` 가 가리키는 표의 id 임 · verify: `cargo nextest run --locked --offline --test model_defaults_one_table --test update_command_cli --test update_model_pick`
- AC-10: 화면 기본이 바뀌어 낡아진 시험이 기대를 지우지 않고 새 기본으로 바뀐 채 통과함 · verify: `cargo nextest run --locked --offline --test model_effort_questions --test update_model_pick --test update_command_cli --test update_model_step_reachable --test update_model_step_reachable_tty --test model_effort_settings`

## Scope

- write: `.moai/specs/SPEC-20261121-model-defaults-one-table/**`, `src/model_effort_settings.rs`, `src/session.rs`, `tests/model_effort_settings.rs`, `tests/model_defaults_one_table.rs`, `tests/launcher_reads_models.rs`, `tests/launcher_roleless.rs`, `tests/launcher_agent_entry.rs`, `tests/launcher_agent_numbering.rs`, `tests/launcher_advisor_setting.rs`, `tests/launcher_permissions_setting.rs`, `tests/product_cli.rs`, 그리고 `src/main.rs` 의 `--model`·`--effort` 도움말 두 줄(공유 파일이라 줄 임대 `moai card line-lease` 로만), 그 두 줄이 바뀌어 낡아지는 도움말 골든 파일 `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`(그 두 줄의 문구만 바꿔요), 그리고 판 3 이 더하는 `src/model_effort_questions.rs`, `tests/model_effort_questions.rs`, `tests/update_model_pick.rs`, `tests/update_command_cli.rs`
- read-only: `src/update_models.rs`
- forbidden: 그 밖의 경로, 실제 HOME·키 보관소, 카드 번호를 SPEC id·브랜치·커밋에 쓰기, 다른 카드의 가지에 쓰기

## Plan

1. SPEC (이 파일), 승인은 리더가 읽고 기록해요.
2. RED: `tests/model_defaults_one_table.rs` 에 저장값 없는 임시 HOME 에서 도구 둘×역할 둘의 argv 가 표를 따르는지, 설정 파일이 없을 때도 같은지, 도움말 문구가 낡지 않았는지 쓰고 지금 팁에서 실패하는 것을 기록해요.
3. GREEN: `src/session.rs` 의 `apply_saved_choice` 를 "저장값, 없으면 표" 로 넓혀요(역할 세션만, 설정 파일이 없을 때 포함, 저장값 거절 규칙은 그대로). 표는 `model_effort_settings.rs` 에서 읽어요.
4. 낡은 기대(`without_a_saved_choice_nothing_changes`, 런처 시험 묶음, `product_cli` 의 런처 시험)를 지우지 않고 표의 값으로 바꿔요.
5. `src/main.rs` 도움말 두 줄을 줄 임대로 고쳐요.
6. 가드 뮤턴트: 새 시험마다 코드 한 곳을 빼서 실패하는지 보고 READY 에 `guard-mutant:` 줄로 적어요.
7. AC 검사를 돌리고 Result 에 줄마다 명령과 종료 코드를 적어요. `check affected` 는 READY 때 한 번, 감싸지 않고 돌려요.
8. (판 3, 승인 뒤) RED: `tests/model_effort_questions.rs` 에 화면 기본이 `launcher_default` 에서 온다는 시험(`claude` `medium`, `glm` `high`)과 `tests/update_command_cli.rs` 에 Enter 로 저장한 값이 저장값 없는 `cc --dry-run` 의 effort 와 같다는 시험을 쓰고 지금 팁에서 실패하는 것을 기록해요.
9. (판 3) GREEN: `src/model_effort_settings.rs` 에 화면과 런처가 함께 읽는 함수 하나를 두고(`launcher_default` 의 effort 와 leader 행의 모델 id), `src/model_effort_questions.rs` 의 `table_default` 가 그 함수를 읽게 해요. `update_model_pick`·`update_command_cli` 의 낡은 기대는 지우지 않고 새 기본으로 바꿔요. 가드 뮤턴트는 `table_default` 가 다시 leader 행의 effort 를 읽게 하는 한 곳이에요.

## Risks

- 역할 세션이 Claude 에서 별칭(`opus`) 대신 전체 id 를 받아요. 저장값 경로가 이미 같은 id 를 넘기므로 호스트가 받는 형태는 같아요. 되돌리기는 커밋 revert 하나예요.
- 설정 파일이 없는 사용자의 역할 세션이 `opus`/`medium` 에서 `claude-opus-5-5`/`high` 로 바뀌어요(결정의 의도). 역할 없는 세션은 그대로예요.
- 낡은 기대를 못 박은 시험이 여럿이라 한 시험만 고치면 다른 묶음이 빨개져요: 표의 값으로 한꺼번에 바꾸고 묶음 전체를 돌려요.
- 조각 B 가 같은 표를 질문 화면에서 다시 읽어요. 조각 A 는 표의 모양(`default_choice` 시그니처)을 바꾸지 않아서 B 와 부딪치지 않아요.
- (판 3) 별칭 `opus` 가 `claude-opus-5-5` 를 가리킨다는 전제는 호스트(Claude Code)의 매핑이고 이 저장소가 시험하지 못해요. 틀리면 화면의 모델 기본이 런처와 다른 모델을 뜻할 수 있어요(effort 는 시험이 지켜요). 호스트가 별칭을 더 새 모델로 옮기면 표의 leader 행을 같이 올려요.
- (판 3) 판 3 뒤 `moai update` 에서 Claude 에 Enter 만 누르면 `claude-opus-5-5`/`medium` 이 저장돼요(전에는 `high`). 저장된 값은 평범한 런처가 이미 하던 것과 같아요. 되돌리기는 커밋 revert 하나예요.

## Result

(status approved — 이 Result 는 이 가지에서 직접 돌린 명령의 출력이에요. `check affected` 영수증과 가벼운 검사(AC-7)는 이 Result 가 든 헤드를 이름 짓는 READY 노트에 있어요.)

- AC-1: met — `tests/model_effort_settings.rs` 가 `default_choice` 4줄과 지원 표를 이미 pin 하고 있어요(`no_saved_value_answers_the_default_for_every_tool_and_role`, `every_default_is_inside_the_support_table`, `the_glm_table_lists_54_and_keeps_the_53_rows`). 아래 묶음 실행(green2)에서 통과했고, 표 값과 그 시험은 바꾸지 않았어요.
- AC-2: met — `tests/model_defaults_one_table.rs` 의 `without_a_settings_file_a_role_session_starts_with_the_table_default`(설정 파일 없음, 도구 둘×역할 둘의 argv 가 표의 모델·effort 이고 Claude 는 전체 id, GLM 은 `zai/` 를 뗀 id, 기본값은 저장값 노트를 달지 않음)와 `a_settings_file_without_models_gives_the_table_default`(설정 파일은 있으나 `models` 없음).
- AC-3: met — `a_flag_given_keeps_that_field_and_the_table_fills_the_other`(준 깃발은 남고 다른 칸만 표), `a_saved_choice_wins_over_the_table_and_the_other_role_keeps_the_table`(저장값 > 표, 저장 안 한 역할은 표); 지원 표 밖 저장값 거절은 기존 `an_unsupported_saved_entry_refuses_a_role_launch` 가 그대로 통과.
- AC-4: met — `a_session_without_a_role_keeps_opus_and_medium`(새)와 `launcher_roleless` 의 roleless 시험들. 레거시 `session` 하위 명령은 런처 역할 세션이 아니어서 `opus`/`medium` 그대로예요(`role_paths_are_unchanged`).
- AC-5: met — 낡은 기대 16개를 지우지 않고 표의 값으로 바꿨어요(`launcher_advisor_setting`, `launcher_agent_entry`, `launcher_agent_numbering`, `launcher_permissions_setting`, `launcher_reads_models`, `launcher_roleless`, `product_cli`). 한 가지 바뀐 행동: 에이전트 기본 모델이 sonnet 이라 advisor `sonnet` 이 더는 낮은 등급이 아니어서, 낮은 등급 거절 시험은 `cc -a --model opus` 로 물어요.
- AC-6: met — `the_model_and_effort_help_names_the_default_table`(낡은 문구가 없고 `default table`·`role` 이 있음)와 `MOAI_HELP_GOLDEN_WRITE=1 cargo test --test help_layout` 로 다시 쓴 골든 네 파일(`all`·`all-60`·`all-80`·`all-dumb`)이 그 두 줄만 바뀐 것을 diff 로 읽었어요.
- AC-7: met — 가벼운 검사 다섯 개를 이 Result 가 든 헤드에서 돌렸어요. `git diff --check`, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings` 는 `moai check affected` 의 `diff_check`·`fmt`·`clippy` 단계 exit 0 이고, `moai spec check` 는 errors 0·warnings 0, `moai trailer-check` 는 실패 0 이에요. 이 줄을 담은 커밋의 해시는 그 커밋 뒤에야 알 수 있어서 헤드 해시와 단계별 exit·로그 폴더는 READY 노트에 있어요. 첫 독립 감사(FAIL, 노트 8304)가 이 줄이 없다는 것(B1)과 `tests/launcher_reads_models.rs` 의 낡은 주석(N1)을 지적했고, 둘 다 이 수리에서 고쳤어요.

증거 (이 가지에서 직접 돌린 것):

- 먼저 실패함: RED 커밋 `9ae2a56761` 에서 `cargo nextest run --locked --offline --no-fail-fast --test model_defaults_one_table` → exit 100, 6 run, 1 passed(`a_session_without_a_role_keeps_opus_and_medium`, 처음부터 통과하는 보호 시험), 5 failed(`target/model-defaults/red.log`). 실패 이유는 `opus` 가 나오고 표의 `claude-opus-5-5` 가 아님, 그리고 낡은 도움말 문구예요.
- 통과: 코드 변경 뒤 런처 시험 묶음(`--lib` 와 새 시험, `model_effort_settings`·`model_effort_questions`, `launcher_*` 열 개, `codex_launcher_removed`, `omp_launcher_removed`, `product_cli`, `update_model_pick`) → 2128 run. 낡은 기대를 고친 뒤 2127 passed, 1 failed(`role_paths_are_unchanged` 의 레거시 `session` 기대를 제가 잘못 바꾼 것이라 되돌림)(`target/model-defaults/green2.log`); 되돌린 뒤 `--test model_defaults_one_table --test launcher_roleless` → 9 run, 9 passed, exit 0(`green3.log`).
- 뮤턴트 7개가 모두 죽었어요(`target/model-defaults/mutants.sh`, 결과 `mutants.log`, 매번 적용 여부와 복원을 확인): M1 표 폴백 제거, M2 설정 파일이 있을 때만 표, M3 GLM `zai/` 접두사를 떼지 않음, M4 준 모델 깃발을 덮어씀, M5 기본값에도 저장값 노트, M6 표가 저장값을 이김, M7 역할 없는 세션에도 표 → 모두 `--test model_defaults_one_table` exit 100.
- 버리는 HOME 의 `--dry-run`(`target/model-defaults/dry.sh`, 출력 `dry.out`, 설정 파일 없음, 더미 토큰, 호스트 미시작): `cc -l` → `claude-opus-5-5`/`high`, `cc -a` → `claude-sonnet-5-5`/`high`, `glm -l` → `glm-5.4[1m]`/`high`, `glm -a` → `glm-5.3-flash`/`max`, `cc` → `opus`/`medium`; 모두 exit 0.

결정 (D-1~D-5 는 위 Decisions): 새로 함, 표는 지금 4줄, 전체 id, 조각 A 먼저, 도움말 골든은 판 2 로 범위에 더함.

못 한 것 (not_run):

- 조각 B(질문 화면 재설계)는 C-248 착지 뒤라 이 개정에 없어요. 이 카드는 B 기준으로 열린 채 남아요.
- 실제 `~/.moai`·`~/.claude` 에 쓰기, 실제 호스트 시작, Linux·Windows, 독립 감사(작성자가 아닌 세션 몫).
