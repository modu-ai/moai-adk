---
id: SPEC-20261121-model-step-reachable
title: "The model step of moai update stays reachable when no model setting is saved"
status: draft
revision: 1
created: 2026-11-21
---

# The model step of moai update stays reachable when no model setting is saved

## Intent

설치 흐름 점검에서 나온 틈이에요. `moai update`의 모델 단계(도구·모델·효과 질문)는 사람의 터미널에서 이번 실행이 파일을 썼을 때만 열려요(`src/update.rs`의 `wrote` 조건). 설치 스크립트가 `--confirm`으로 먼저 다 적용하고 나면 다음 `moai update`에는 적용할 것이 없어 모델 단계가 열리지 않고, 세션의 `/moai update`는 늘 `--confirm`을 쓰므로 대부분의 사람이 모델 설정을 저장할 길이 없어요. 사용자 결정(2026-11-20): 사용자 `launcher.yaml`에 저장된 모델이 없으면 터미널의 `moai update`는 적용할 것이 없어도 늘 모델 단계를 열어요.

이 SPEC은 터미널 쪽만 다뤄요. 세션의 `/moai update`가 같은 질문을 호스트 질문 도구로 묻는 쪽은 계획·적용 명령(저장 파일을 digest에 묶어 쓰는 비대화형 명령)이 먼저 있어야 해서 다음 조각이고, 이 SPEC에 넣지 않아요.

## Non-goals

- 세션 `/moai update`의 질문(스킬 글과 계획·적용 명령)은 하지 않아요.
- 질문 화면의 모양(표 + 행 고르기)과 기본값 표는 바꾸지 않아요. 다른 카드가 쓰는 중이에요.
- 사람의 터미널이 아닌 실행(파이프, `--confirm…`, `--json`)은 바꾸지 않아요. 질문 없이 값만 알려 줘요.
- 저장값이 하나라도 있는 사람의 동작은 바꾸지 않아요: 이번 실행이 파일을 썼을 때만 열려요.

## Requirements

- REQ-1: When a person's terminal runs `moai update` with nothing to apply and no model setting saved, `moai update` shall open the model step · failure: none — the step asks and a no at its save question writes nothing, so the next run asks again.
- REQ-2: When a person's terminal runs `moai update` that writes files, `moai update` shall open the model step as before · failure: none — the earlier behaviour, unchanged.
- REQ-3: While at least one model setting is saved and nothing was written, `moai update` shall leave the model step closed · failure: none.
- REQ-4: While a run is scripted or has no person's terminal, `moai update` shall report the model values without asking.
- REQ-5: When no model setting is saved, the decision to open the step shall read only the user launcher settings file and never write it · failure: none — an unreadable or damaged file counts as saved, so the step is not forced open over it.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): the decision function answers true exactly for a terminal run that is not scripted and either wrote or has no saved model; every one of the sixteen input combinations is tested. verify: `cargo test --locked --offline --test update_model_step_reachable`
- AC-2 (REQ-5): `models_saved` answers false for no file, a file without `models` and an empty `models`; true for a saved entry, a damaged `models` and an unreadable file; the file is the same bytes afterwards. verify: `cargo test --locked --offline --test update_model_step_reachable`
- AC-3 (REQ-1, REQ-4): a run on a pseudo-terminal with nothing to apply and an empty HOME state asks the tool pick; a run on a pipe still reports without asking and keeps the preview's keys. verify: `cargo test --locked --offline --test update_model_step_reachable_tty`; `cargo test --locked --offline --test update_model_pick`
- AC-4: light checks. verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/update.rs`, `tests/update_model_step_reachable.rs`, `tests/update_model_pick.rs`, `tests/update_command_cli.rs` (two terminal cases), `.moai/specs/SPEC-20261121-model-step-reachable/**`
- read-only: `src/update_models.rs`, `src/model_effort_questions.rs`, `src/update_tux.rs`, `src/main.rs`
- forbidden: keys and the real HOME, card numbers in a SPEC id, branch or commit, writes to another card's branch

## Plan

1. Test first: `tests/update_model_step_reachable.rs` holds the sixteen-combination test of `model_step_opens` and the saved-models cases of `models_saved`; both fail to compile before the functions exist (RED).
2. `src/update.rs`: add the two functions; `update_command_switch` extends a preview that has nothing to apply, on a person's terminal with no saved model, with the model step, and keeps every other preview as it was.
3. A pseudo-terminal run for the end to end proof (AC-3), as a Rust test in `tests/update_model_step_reachable_tty.rs` (macOS only, like the other terminal tests; the leader chose Rust-only tests).
4. Run the checks and write each command and exit code into Result.

## Risks

- A person who answers no at the save question is asked again at every `moai update` until a setting is saved; the decision says "always", so this stays, and a remembered refusal is a later choice for the leader.
- A preview that has nothing to apply gains a `models` key on a person's terminal only; the pipe preview keeps its keys.
- Rollback: one revert commit (two functions and one condition).

## Decisions

- 2026-11-21: this SPEC covers the terminal part only; the session part needs a plan and apply command first and goes to the leader as its own piece. Reversible, category `none`, decided by agent-17.
- 2026-11-21: two terminal cases of `tests/update_command_cli.rs` (`in_a_terminal_nothing_to_write_asks_nothing`, `in_a_terminal_an_already_applied_package_asks_nothing`) assumed that a terminal run with nothing to write asks nothing; they now save a model setting first, because with none saved the model step opens (REQ-1). Reversible, category `none`, decided by the leader.
- 2026-11-21: an unreadable or damaged settings file counts as saved, so a damaged file is reported by the step and never forced open. Reversible, category `none`, decided by agent-17.

## Result

(Filled after the implementation.)
