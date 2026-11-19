---
id: SPEC-20261119-glm-launcher-settings-receipt
title: "The glm launcher reads the saved model and effort of its role (C-47, criterion 5 glm part)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261118-install-model-per-tool
---

# The glm launcher reads the saved model and effort of its role (C-47, criterion 5 glm part)

## Intent

카드 C-47 의 기준 5 는 `moai cc|glm|codex -l|-a` 가 저장된 리더/에이전트 모델과 노력(effort) 로 도구를 띄우라고 한다. 조각 8(SPEC-20261119-launcher-reads-models, PR 3232 병합)이 claude·codex·omp 를 끝냈고 glm 은 결정 D-5 에 따라 이 조각으로 미뤄졌다: 저장된 `glm` 항목의 모델이 z.ai 를 가리키는 Claude Code 환경에 어떻게 닿는지 코드만으로 정해지지 않았기 때문이다.

조사 노트(C-47 note 3452, 2026-11-19)가 그 간극을 메운다. 오늘 `moai glm -l` 은 `apply_saved_choice` 의 호스트 일치에서 빠져 저장값을 전혀 읽지 않고, `claude_argv` 기본값 `--model opus --effort medium` 을 넘긴 다음 환경의 별칭 매핑(ANTHROPIC_DEFAULT_OPUS_MODEL 등)이 opus 를 glm-5.3 로 바꾼다. Claude Code 공식 문서(model-config, 2026-11-20 읽음)는 `--model` 이 최상위 우선순위이며 모르는 전체 모델 id 를 미리 거부하지 않고 "The session sends it as you wrote it" — 쓴 그대로 API 로 보낸다고 하고, z.ai 문서는 Anthropic 호환 엔드포인트의 모델 이름에 접두사가 없음을 보인다. 저장 표의 id 는 OMP 와 공유되는 `zai/glm-5.3` 형태이므로, glm 호스트에 넘길 때 `zai/` 공급자 접두사만 떼면 z.ai 가 받는 모델 이름이 된다.

이 조각은 `apply_saved_choice` 의 호스트 일치에 glm 을 더하고, 모델을 채울 때 접두사를 떼고, 노트가 `launcher.models.glm.<역할>` 을 이름하게 한다. 환경(엔드포인트, 별칭 매핑, 1M 컨텍스트, 토큰)은 저장값으로 바뀌지 않는다.

## Non-goals

- 저장값이 없을 때 런처 기본값을 `default_choice` 표로 바꾸는 일 (사용자 결정 D-4): 저장값이 없는 glm 은 오늘처럼 `--model opus --effort medium` 과 별칭 매핑 환경으로 띄운다.
- 역할 없는 세션(`moai glm`): 사용자 결정 2026-11-08(A) 그대로 저장값을 무시한다.
- 질문·저장·되돌리기 코드(`src/update_models.rs`, `src/model_effort_questions.rs`)와 지원 표(`src/model_effort_settings.rs`): 이미 glm 을 안다.
- 환경 변수 세트 자체의 변경: 별칭 매핑과 컨텍스트 창은 Go 세대의 setGLMEnv 를 따른 사용자 결정(2026-11-10) 그대로.
- `src/main.rs` 의 도움말 문장과 `docs/`·`skills/` 문서: glm 런처가 저장값을 읽는다는 문장은 어디에도 없다(조사 노트 확인).

## Requirements

- REQ-1: When a session with a role is composed for the host `glm`, and no `--model` was given, and the user launcher settings file holds a supported saved choice for `glm` and that role, the launcher shall pass the saved model with a leading `zai/` provider prefix removed in `--model`, the way the Claude Code host takes a model. · failure: none — REQ-4 keeps today's launch when nothing is saved, and REQ-5 refuses a saved entry the support table does not accept.
- REQ-2: When a session with a role is composed for the host `glm`, and no `--effort` was given, and the file holds a supported saved choice for `glm` and that role, the launcher shall pass the saved effort in `--effort` the way the claude host does. · failure: none — the existing effort check of `claude_argv` refuses a value outside low|medium|high|xhigh|max|ultracode, and the support table offers low|high|max only.
- REQ-3: While `--model` or `--effort` is given on the command line, the launcher shall use the given value for that field and the saved choice only for the other field, the same per-field rule as the claude, codex and omp hosts. · failure: none — a given value that fails its check is refused as before.
- REQ-4: While no supported saved choice exists for `glm` and the role, or the session has no role, the launcher shall compose the glm session exactly as before this change: aliases mapped to z.ai models by the environment, the default model alias and effort of `claude_argv`, the same environment keys. · failure: none — a pinned test compares a launch with and without the file.
- REQ-5: IF a session with a role is composed for the host `glm` while the user launcher settings file holds an entry under `launcher.models` that the support table does not accept, THEN the launch shall refuse with the fixed message that names key paths only, the whole-file check the claude, codex and omp role launches already run. · failure: none — the refusal is the outcome.
- REQ-6: When the launcher passes a saved glm value, it shall add the note that names `launcher.models.glm.leader` or `launcher.models.glm.agent` and no other saved text. · failure: none — a pinned output test finds a leak.
- REQ-7: When a glm host is started, the child process shall carry the saved model in its argv and the same environment as before this change (endpoint, alias mapping, context window, the glm token under its two names). · failure: none — a started-host test records the argv and the environment of the child.

## Acceptance

- AC-1 (REQ-1, REQ-2): with the saved choices of the test fixture in a synthetic HOME, `moai glm -l --dry-run` prints `--model glm-5.3-flash --effort max` and `moai glm -a --dry-run` prints `--model glm-5.3 --effort low` (the `zai/` prefix of the saved id removed), with the environment keys unchanged.
  verify: `cargo test --locked --offline --test launcher_reads_models saved_values_reach_the_argv_of_each_host_and_role`
- AC-2 (REQ-3): on `glm -l` an explicit `--model opus` keeps that alias in the argv and only the effort comes from the file; an explicit `--effort` keeps that level and only the model comes from the file.
  verify: `cargo test --locked --offline --test launcher_reads_models explicit_flags_win_per_field`
- AC-3 (REQ-4): a roleless `glm` session and a `glm` role session with no saved `glm` entry compose byte for byte like the same session in a HOME without the file; with no file at all the role session keeps `--model opus --effort medium`.
  verify: `cargo test --locked --offline --test launcher_reads_models a_roleless_session_and_glm_ignore_saved_choices`
  verify: `cargo test --locked --offline --test launcher_reads_models without_a_saved_choice_nothing_changes`
- AC-4 (REQ-5): a file with a saved entry outside the support table refuses `glm -l` with exit 1 and nothing on stdout, the message holding key paths only, while a roleless `glm` launch still starts.
  verify: `cargo test --locked --offline --test launcher_reads_models an_unsupported_saved_entry_refuses_a_role_launch`
- AC-5 (REQ-6): the note of a saved glm value names `launcher.models.glm.leader` or `launcher.models.glm.agent`, and the output holds no other saved text.
  verify: `cargo test --locked --offline --test launcher_reads_models a_saved_value_adds_a_note_with_its_key_path`
- AC-6 (REQ-7): a started glm host (stub `claude` on PATH) receives the saved model in its argv and the alias mapping and the glm token in its environment.
  verify: `cargo test --locked --offline --test launcher_reads_models a_saved_glm_model_reaches_the_started_host`
- AC-7 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): the neighbour launcher checks and the style checks pass.
  verify: `cargo test --locked --offline --test launcher_permissions_setting`
  verify: `cargo test --locked --offline --test launcher_advisor_setting`
  verify: `cargo test --locked --offline --test launcher_agent_numbering`
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-glm-launcher-settings-receipt/spec.md`, `src/session.rs`, `tests/launcher_reads_models.rs`
- read-only: `src/model_effort_settings.rs`, `src/main.rs`, `tests/launcher_permissions_setting.rs`
- forbidden: `src/main.rs`, 실제 HOME 과 원장, 네트워크, 키, 지원 표와 질문 코드

## Plan

1. SPEC 을 쓰고 검사 오류 0 을 확인한다.
2. RED: `tests/launcher_reads_models.rs` 의 glm 줄(기준표 1, 2, 5, 6)과 새 시작 경로 시험(7)이 지금 코드로 실패하는 것을 본다.
3. GREEN: `src/session.rs` 의 `apply_saved_choice` 호스트 일치에 glm 을 더하고 모델 칸을 채울 때 `zai/` 접두사를 떼며, `saved_note` 에 glm 팔을 추가한다.
4. 변이: 호스트 조건, 접두사 제거, 칸별 우선, 노트 팔을 하나씩 뒤집어 시험이 빨개지는지 본다.
5. 이웃 시험, fmt, clippy, `trailer-check`, 팁과 `git merge-tree`, tier 1 독립 감사.

## Risks

- `zai/` 접두사 제거는 표(id 는 OMP 와 공유)와 호스트(z.ai 는 접두사 없는 이름) 사이의 번역층이다: 한 줄 접두사 제거로 유지하고 근거를 조사 노트(3452)에 둔다. 표가 나중에 접두사 없는 glm id 를 받으면 이 줄이 그대로 통과된다.
- 지원 밖 항목이 파일 어디에 있어도 glm 역할 실행이 거절된다(다른 도구의 항목이어도): claude·codex·omp 와 같은 전체 파일 검사이며, 거절 문구의 키 경로를 보고 고치거나 `moai update` 로 바꾼다.
- 저장 노력 `max` 는 Claude Code 가 `--effort` 플래그로만 받는 값이다: 런처는 플래그로 넘기므로 해당 없음(조사 노트의 문서 근거).
