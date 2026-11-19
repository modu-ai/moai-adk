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

이 조각은 `src/session.rs` 의 `compose` 가 역할이 있는 세션(`-l` 리더, `-a` 에이전트)에서 저장된 선택을 읽어, 사람이 따로 `--model` 이나 `--effort` 를 주지 않은 칸에 채우게 해요. 파일은 실행할 때마다 읽으니(`user_launcher()` 가 이미 그렇게 해요) 파일을 고치고 다시 띄우면 재설치 없이 새 값이 쓰여요(기준 7).

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
3. `src/session.rs` 의 `compose` 맨 앞에서, 역할이 있고 호스트가 claude, codex, omp 이며 파일의 `launcher.models` 에 항목이 있으면 `saved_choice` 로 `r.model`, `r.effort` 가 비어 있는 칸만 채운 사본 요청으로 각 호스트의 `*_argv` 를 부른다. 지원표 밖 항목은 고정 문구로 거절한다(`unsupported_entries`).
4. 변이: 값을 채우는 줄, 역할 조건, 호스트 조건, 칸별 우선순위를 하나씩 뒤집어 시험이 빨개지는지 본다.
5. 이웃 시험, fmt, clippy, `trailer-check`, 팁과 `git merge-tree`, tier 1 독립 감사.

## Risks

- `src/session.rs` 는 여러 카드가 고치는 파일이에요(OMP 호스트 C-62 등): 이 조각은 `compose` 맨 앞에 한 블록만 더하고 `*_argv` 함수들은 건드리지 않아 병합 충돌 면을 줄여요. `card take` 가 겹치는 예약을 거절하면 그 사실을 보고해요.
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
