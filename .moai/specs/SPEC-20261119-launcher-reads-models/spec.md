---
id: SPEC-20261119-launcher-reads-models
title: "The launchers read the saved model and effort of each tool and role (C-47 slice 8)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261118-install-model-per-tool
---

# The launchers read the saved model and effort of each tool and role (C-47 slice 8)

## Intent

카드 C-47 의 기준 5 와 7 이 열려 있어요. `moai update` 의 모델 단계가 `~/.moai-cli/config/launcher.yaml` 의 `launcher.models.<도구>.<역할>` 에 모델과 노력(effort) 값을 저장하지만(조각 3, PR 3212), 런처(`moai cc|codex|omp -l|-a`)는 그 값을 읽지 않고 고정 값(Claude 는 `--model opus --effort medium`)이나 아무 값도 안 넘기는 규칙(Codex 와 OMP)을 써요. 그래서 사람이 질문에 답해 저장한 선택이 세션에 닿지 않아요.

이 조각은 `src/session.rs` 의 `compose` 가 역할이 있는 세션(`-l` 리더, `-a` 에이전트)에서 저장된 선택을 읽어, 사람이 따로 `--model` 이나 `--effort` 를 주지 않은 칸에 채우게 해요. 값을 채우는 일은 `apply_saved_choice` 한 함수이고, 실제로 호스트를 띄우는 `launch()` 도 같은 채워진 요청을 써야 OMP 의 저장된 z.ai 모델이 키를 받으므로 `compose` 를 `compose_resolved` 로 나누고 `launch()` 도 그 함수를 불러요(감사 N2). 파일은 실행할 때마다 읽으니(`user_launcher()` 가 이미 그렇게 해요) 파일을 고치고 다시 띄우면 재설치 없이 새 값이 쓰여요(기준 7).

저장된 값은 사람이 직접 고른 선택이에요. Codex 와 OMP 는 지금도 "사람이 요청하지 않으면 런처가 모델 기본값을 고르지 않는다"(리더 결정 6905681774, #3096 S7b)인데, 저장된 선택은 그 요청에 해당하므로 그 규칙을 깨지 않아요. 저장된 값이 없으면 오늘과 똑같이 움직여요.

## Non-goals

- `glm` 호스트(z.ai 를 가리키는 Claude Code): 저장된 `glm` 항목의 모델이 환경 변수 매핑(`ANTHROPIC_DEFAULT_*_MODEL`)에 어떻게 닿는지 코드만으로 정해지지 않아요. 이 조각은 `glm` 을 그대로 두고, 다음 작은 조각에서 조사 노트 뒤에 다뤄요(결정 D-5).
- 저장값이 없을 때 런처 기본값을 `default_choice` 표로 바꾸는 일: 모든 사람의 Claude 에이전트가 `opus/medium` 에서 `claude-sonnet-5-5/high` 로 바뀌므로 하지 않아요(결정 D-4).
- 역할 없는 세션(`moai cc` 만): 사용자 결정 2026-11-08(A) 그대로.
- 질문, 저장, 되돌리기 코드(`src/update_models.rs`, `src/model_effort_questions.rs`, `src/model_effort_settings.rs`)와 `src/main.rs`, 도움말.
- 카드 C-47 의 다른 남은 일(영수증 폴더 링크 N2, 상태 폴더 링크 결정).

## Requirements

- REQ-1: When a session with a role is composed for the host `claude`, `codex` or `omp`, and no `--model` was given, and the user launcher settings file holds a supported saved choice for that host and role, the launcher shall pass the saved model in the way that host takes a model. · failure: none — REQ-8 covers a file that cannot be read, and the id check that refuses a given model refuses a saved one the same way.
- REQ-2: When a session with a role is composed for the host `claude`, `codex` or `omp`, and no `--effort` was given, and the user launcher settings file holds a supported saved choice for that host and role, the launcher shall pass the saved effort in the way that host takes an effort. · failure: none — REQ-8 covers a file that cannot be read, and an effort the host does not take is left out with the host's existing note.
- REQ-3: When `--model` or `--effort` is given, the launcher shall use the given value for that field and the saved value only for the other field. · failure: none — a given value that fails its check is refused as before.
- REQ-4: While no supported saved choice exists for the host and role, the launcher shall compose the session as it did before this change. · failure: none — the existing launcher tests are the check.
- REQ-5: While a session has no role, the launcher shall ignore the saved choices. · failure: none — a pinned test finds a difference.
- REQ-6: While the host is `glm`, the launcher shall ignore the saved choices. · failure: none — a pinned test finds a difference.
- REQ-7: When the user launcher settings file is rewritten between two launches, the launcher shall use the values of the file as it is at the second launch without another step. · failure: none — REQ-8 covers a rewritten file that cannot be read.
- REQ-8: IF a session with a role is composed for the host `claude`, `codex` or `omp` while the user launcher settings file holds a saved entry under `launcher.models` that the support table does not accept, THEN the launcher shall refuse with a fixed message that names at most key paths and no file text. · failure: none — the refusal is the outcome. (`unsupported_entries` of `model_effort_settings` already says the launcher refuses such a file until the entries are replaced; a file that cannot be read or does not parse is already refused for every launch by the existing check that `tests/launcher_permissions_setting.rs` pins)
- REQ-9: When the launcher passes a saved value, it shall add a note that says which key path the value came from and no other saved text. · failure: none — a pinned output test finds a leak.

## Acceptance

- AC-1 (REQ-1, REQ-2): with a saved choice in a synthetic HOME, `moai cc|codex|omp -l|-a --dry-run` prints the argv with the saved model and effort in each host's own flag (claude `--model` and `--effort`, codex `-m` and `-c model_reasoning_effort="…"`, omp `--model` and `--thinking`), for every host × role of the three.
  verify: `cargo test --locked --offline --test launcher_reads_models saved_values_reach_the_argv_of_each_host_and_role`
  verify: `cargo test --locked --offline --test launcher_reads_models a_saved_zai_model_reaches_an_omp_host_that_is_started`
- AC-2 (REQ-3): `--model` alone replaces only the model and `--effort` alone replaces only the effort, the other field still from the file.
  verify: `cargo test --locked --offline --test launcher_reads_models explicit_flags_win_per_field`
- AC-3 (REQ-4, REQ-5, REQ-6): with no saved choice the three hosts compose as before (claude `--model opus --effort medium` for a role, codex and omp without a model or effort flag); a roleless session and a `glm` session with saved entries compose byte for byte like the same session with no file.
  verify: `cargo test --locked --offline --test launcher_reads_models without_a_saved_choice_nothing_changes`
  verify: `cargo test --locked --offline --test launcher_reads_models a_roleless_session_and_glm_ignore_saved_choices`
- AC-4 (REQ-7): launch, rewrite the file with other values, launch again: the second argv carries the new values.
  verify: `cargo test --locked --offline --test launcher_reads_models a_rewritten_file_is_used_by_the_next_launch`
- AC-5 (REQ-8): a file with a saved entry outside the support table (an unknown model, an effort the model does not offer, or the bad entry of another tool) refuses a role launch on claude, codex and omp with exit 1 and nothing on stdout; the message holds at most key paths and no text of the file; a roleless launch and a `glm` launch are not refused by it.
  verify: `cargo test --locked --offline --test launcher_reads_models an_unsupported_saved_entry_refuses_a_role_launch`
- AC-6 (REQ-9): the note names the key path (for example `launcher.models.claude.leader`) and the output holds no other saved text.
  verify: `cargo test --locked --offline --test launcher_reads_models a_saved_value_adds_a_note_with_its_key_path`
- AC-7 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the neighbour launcher checks and the style checks pass.
  verify: `cargo test --locked --offline --test launcher_permissions_setting`
  verify: `cargo test --locked --offline --test launcher_advisor_setting`
  verify: `cargo test --locked --offline --test launcher_agent_numbering`
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-launcher-reads-models/spec.md`, `src/session.rs`, `tests/launcher_reads_models.rs`
- read-only: `src/model_effort_settings.rs`, `src/model_effort_questions.rs`, `src/update_models.rs`, `src/main.rs`, `tests/launcher_permissions_setting.rs`
- forbidden: `src/main.rs`, 실제 HOME 과 원장, 네트워크, 키, `skills/` 아래 파일(문서 문장은 다음 조각)

## Plan

1. SPEC 을 쓰고 `spec check` 오류 0 을 확인한다.
2. RED: `tests/launcher_reads_models.rs` 를 쓰고 저장값 시험이 지금 코드로 실패하는 것을 본다(없는 줄은 argv 에 없어요).
3. `src/session.rs` 에 `apply_saved_choice` 를 더하고 `compose` 와 `launch()` 가 둘 다 그 결과(`compose_resolved`)를 쓰게 한다. 그 함수는 역할이 있고 호스트가 claude, codex, omp 이며 파일의 `launcher.models` 에 항목이 있으면 `saved_choice` 로 `r.model`, `r.effort` 가 비어 있는 칸만 채운 사본 요청으로 각 호스트의 `*_argv` 를 부른다. 지원표 밖 항목은 고정 문구로 거절한다(`unsupported_entries`).
4. 변이: 값을 채우는 줄, 역할 조건, 호스트 조건, 칸별 우선순위를 하나씩 뒤집어 시험이 빨개지는지 본다.
5. 이웃 시험, fmt, clippy, `trailer-check`, 팁과 `git merge-tree`, tier 1 독립 감사.

## Risks

- `src/session.rs` 는 여러 카드가 고치는 파일이에요(OMP 호스트 C-62 등): 이 조각은 함수 둘(`apply_saved_choice`, `saved_note`)을 더하고 `compose`, `launch()` 의 첫 줄만 바꾸며 `*_argv` 함수들은 건드리지 않아 병합 충돌 면을 줄여요. `card take` 가 겹치는 예약을 거절하면 그 사실을 보고해요.
- 손으로 고친 항목 하나가 지원표 밖이면 claude, codex, omp 의 역할 세션이 전부 거절돼요(다른 도구의 항목이어도): `unsupported_entries` 주석이 이미 정한 방식이지만, 사람은 거절 문구의 키 경로를 보고 그 항목을 고치거나 `moai update` 로 바꿔야 해요. 구조가 깨진 파일은 `moai update` 도 거절하므로 손으로 고쳐야 해요(감사 3번).
- 저장 모델이 `claude-fable-5-2` 이고 `launcher.advisor` 가 `opus` 나 `sonnet` 이면 `--advisor` 랭킹 검사가 새로 거절해요(`--model` 을 직접 줘도 같은 규칙이에요). 거절 문구가 저장값 때문이라고 말하지 않고 이 조각에는 그 시험이 없어요: 다음 조각 후보예요(감사 N3).
- 저장값 노트는 `--dry-run` 과 `--json` 출력에만 나오고 터미널 시작 화면의 경고 줄(`launcher_terminal.rs` 의 `WARNINGS`)에는 안 나와요: 사람이 직접 저장한 값이라 비용이 몰래 바뀌는 건 아니지만 시작 화면에 한 줄 보이게 하는 일은 다음 조각 후보예요(감사 N4).
- Claude 호스트의 저장 모델 id 는 별칭(`opus`) 대신 전체 id(`claude-opus-5-5`)로 나가요: 지원표(`model_effort_settings`)가 그 id 를 공식 문서 근거로 갖고 있어요(카드 기준 2 가 met). 호스트가 id 를 거절하면 그 줄은 고쳐야 하는 사실이 돼요 — 실제 호스트는 실행하지 않았고 argv 만 확인해요.
- 파일이 깨졌을 때 역할 있는 세션이 막히는 점: `launcher.role_access` 가 이미 같은 방식(고정 문구 거절)이라 그 방식에 맞췄어요(결정 D-6).

## Decisions

| id | decision | why | reversible |
| --- | --- | --- | --- |
| D-1 | 저장값은 역할이 있는 세션(`-l`, `-a`)에서만 쓴다 | 저장은 도구와 역할 쌍이고, 역할 없는 세션은 별도 기본값 규칙(2026-11-08 A)이 있어요 | 조건 하나 |
| D-2 | 칸마다 순서는 명시 플래그, 저장값, 오늘의 기본값이다 | 사람이 그때 준 값이 가장 강해야 하고 저장값은 기본값보다 강한 선택이에요 | 순서 |
| D-3 | 저장된 모델 id 는 저장된 그대로 넘긴다 | 지원표가 정한 id 를 런처가 바꾸면 저장과 실행이 달라져요 | 한 줄 |
| D-4 | 저장값이 없으면 오늘의 런처 기본값을 그대로 둔다(표로 바꾸지 않는다) | 표로 바꾸면 모든 사람의 Claude 에이전트가 바뀌어요 | 기본값 줄 |
| D-5 | glm 은 이 조각에서 건드리지 않는다 | 저장된 glm 모델이 환경 변수 매핑에 닿는 방식이 코드로 정해지지 않았어요 | 다음 조각 |
| D-6 | 지원표 밖 저장 항목이 있으면 역할 있는 세션(claude, codex, omp)을 고정 문구로 거절한다(읽지 못하거나 파싱 못 하는 파일은 기존 검사가 이미 모든 실행에서 거절해요) | `unsupported_entries` 의 주석이 이미 "런처가 그런 파일을 거절한다" 고 적어 두었어요. 파일 내용은 인용하지 않고 키 경로만 말해요 | 거절을 무시로 |

(D-4, D-5 는 오늘 사람이 받는 것을 바꾸는 쪽이라 리더가 다르게 정하면 REQ 와 시험이 같이 바뀌어요: 카드 노트 3150.)

## Result

- 독립 감사: 첫 head 0d7ab66e28 은 새 문맥의 읽기 전용 `moai-audit` agent(이 변경을 쓰지 않음; `Moai-Session` 트레일러는 레인 세션의 id 라서 감사 agent 의 세션 id 와 같아요)가 PASS 를 권고했고 막는 발견은 없었어요. 막지 않는 발견 N1~N6: N1(`compose` 의 설명 주석이 `saved_note` 위에 붙음)은 사실로 확인해 고쳤고(커밋 00b68bc3de), N2(SPEC 이 `compose_resolved` 분리와 `launch()` 수정을 말하지 않음)는 SPEC Intro·Plan·Risks 를 고쳤고(커밋 c15bcf369f), N5(AC-5 시험이 키 경로를 정확히 보지 않음)는 시험이 세 파일의 정확한 키 경로(`launcher.models.claude.leader`, `launcher.models.omp.agent`, `launcher.models.codex.agent`)를 단언하게 고쳤어요(같은 커밋; 변이 10 이 이 단언을 확인해요). N3(저장된 `claude-fable-5-2` 와 `launcher.advisor` 의 랭킹 충돌 문구)와 N4(터미널 시작 화면에 저장값 노트가 안 보임)는 Risks 에 다음 조각 후보로 적었어요. N6(원격 팁이 움직임)은 겹치는 파일이 없다고 감사가 확인했고, 나는 로컬 `v2026.11.0` 과의 합친 트리도 아래처럼 시험했어요.
- AC-1: met — 2026-11-19 이 worktree 에서 `cargo test --offline --locked --test launcher_reads_models` 는 구현 전(RED 커밋 e990ee9270) 2 passed 5 failed(저장값을 읽는 시험 5개가 실패, 기존 동작을 못 박는 시험 2개 통과), 구현 뒤 8 passed 예요. `saved_values_reach_the_argv_of_each_host_and_role` 는 claude(`--model`·`--effort`), codex(`-m`·`-c model_reasoning_effort="…"`), omp(`--model`·`--thinking`) × 리더·에이전트 6가지의 argv 를 읽고, `a_saved_zai_model_reaches_an_omp_host_that_is_started` 는 스텁 `omp` 를 실제로 시작해 인자(`--model zai/glm-5.3-flash`, `--thinking max`)와 `ZAI_API_KEY` 값이 닿는 것을 읽어요(`launch()` 가 채운 요청을 쓰는 이유).
- AC-2: met — `explicit_flags_win_per_field` 통과: `--model` 만 주면 노력은 파일에서, `--effort` 만 주면 모델은 파일에서, 둘 다 주면 파일 값도 `launcher.models` 노트도 없어요.
- AC-3: met — `without_a_saved_choice_nothing_changes`(파일 없음: claude `--model opus --effort medium`, codex·omp 는 모델·노력 플래그 없음; 다른 도구만 저장된 파일에서도 같음)와 `a_roleless_session_and_glm_ignore_saved_choices`(역할 없는 `cc`·`codex`·`omp` 와 `glm`·`glm -l`·`glm -a` 가 저장값이 있는 홈과 없는 홈에서 같은 프로젝트로 argv·env 키·노트가 같아요) 둘 다 구현 전에도 통과였고 구현 뒤에도 통과예요.
- AC-4: met — `a_rewritten_file_is_used_by_the_next_launch` 통과: 파일을 다른 값으로 고쳐 다시 띄우면 새 값(claude-sonnet-5-5 / medium), 파일을 지우면 오늘의 기본값(opus / medium).
- AC-5: met — `an_unsupported_saved_entry_refuses_a_role_launch` 통과: 모르는 모델, 모델이 제공하지 않는 노력(glm 모델의 `medium`), 다른 도구의 나쁜 항목 세 파일이 `cc -l`·`codex -a`·`omp -l` 를 exit 1, stdout 비어 있음, stderr 에 `launcher.models` 키 경로만 있고 파일 값(`SECRET`)은 없게 거절하고, 같은 파일에서 역할 없는 `cc` 와 `glm -l` 은 성공해요.
- AC-6: met — `a_saved_value_adds_a_note_with_its_key_path` 통과: 여섯 경우 모두 노트에 `launcher.models.<도구>.<역할>` 이 있고 출력에 파일의 다른 값(`SENTINEL-SAVED-TEXT`)은 없어요.
- AC-7: met — 이웃 시험 통과(`cargo test --offline --locked`): `launcher_permissions_setting` 6, `launcher_advisor_setting` 7, `launcher_agent_numbering` 25, `launcher_codex_effort` 5, `launcher_roleless` 4, `omp_launcher` 17, `launcher_agent_entry` 7, `launcher_terminal_output` 11, `launcher_glm_key_file` 8, `launcher_skill_check` 4, `launcher_auto_init` 18, `--lib session` 27. `cargo fmt --check` exit 0, `git diff --check origin/v2026.11.0...HEAD` exit 0.
- 변이: `src/session.rs` 에서 한 군데씩 뒤집는 변이 9개(glm 을 호스트에 넣음, codex 를 뺌, 저장 모델이 `--model` 을 이김, 저장 노력이 `--effort` 를 이김, 지원표 밖 항목을 거절하지 않음, 노트를 지움, `launch()` 가 채우지 않은 요청을 씀, 저장 모델을 다른 값으로 바꿈, 역할 없는 세션이 리더 선택을 읽음)가 모두 `launcher_reads_models` 에서 1개 이상 실패시켰고 실패한 시험 이름이 변이마다 달랐어요(원복은 `cmp` 로 확인, 실행 뒤 `git status --short` 0줄). 감사 N5 수리 뒤 변이 10(거절 문구가 키 경로 대신 고정 문자열 `launcher.models` 를 찍음)도 `an_unsupported_saved_entry_refuses_a_role_launch` 만 실패시켰어요.
- 합친 트리: `git merge-tree --write-tree <로컬 v2026.11.0 f01de44dfa> 0d7ab66e28` exit 0 이고 그 트리(`a7981d61fb`)를 스크래치에 풀어 `cargo test --offline --locked --test` 로 돌린 `launcher_reads_models` 8, `omp_launcher` 17(로컬 `v2026.11.0` 에 병합된 OMP 호스트 카드 C-62 포함), `launcher_permissions_setting` 6, `launcher_agent_numbering` 25 가 통과예요. 원격 팁 747738af63 과도 exit 0.
- 안 한 것: 실제 Claude Code·Codex·OMP 호스트는 시작하지 않았어요(argv 와 스텁 `omp` 만 확인). 호스트가 전체 id(`claude-opus-5-5` 등)를 받아들이는지는 지원표의 공식 문서 근거(카드 기준 2)에 기대요. `glm` 호스트는 이 조각 밖이에요(D-5). 스킬 문서(`update.md` 등)의 문장은 이 조각에서 바꾸지 않았어요. 전체 로컬 `cargo test`, Windows, 파이썬 전체 실행은 돌리지 않았고, 리더 공지(2026-11-19)의 여섯 점검은 READY 노트에 종료 코드로 적어요. `status: completed`(승인)는 이 레인의 몫이 아니에요.
