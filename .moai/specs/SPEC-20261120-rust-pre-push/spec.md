---
id: SPEC-20261120-rust-pre-push
title: "moai hook pre-push checks commit subjects in Rust"
status: draft
revision: 1
parent: SPEC-20261120-rust-quality-gate
created: 2026-11-20
---

# moai hook pre-push checks commit subjects in Rust

## Intent

사용자 결정(2026-11-20): `moai hook pre-push`(다른 저장소 13곳의 git pre-push 훅이 `printf '%s\n' "$SUBJECTS" | moai hook pre-push` 로 부름)도 Rust 로 옮겨요. 지금 Rust 는 `moai hook <이벤트>` 를 조용히 0 으로 삼켜서 이 검사는 아무것도 하지 않아요. 같은 호출(표준 입력에 커밋 제목을 한 줄씩)과 같은 종료 코드로 Go 빌드(`internal/cli/hook_pre_push.go`, `internal/git/convention`)가 하던 검사를 옮겨요. 다른 저장소의 훅 파일은 바꾸지 않아요.

Go 빌드를 읽고 직접 돌려 본 결과(스크래치 저장소, 빈 HOME):

- 기본값은 검사를 하지 않아요: `git_convention.validation.enforce_on_push` 가 true 이거나 `MOAI_ENFORCE_ON_PUSH` 가 `true`/`1` 일 때만 켜져요. 28개 설정 파일 중 켠 곳은 0곳이에요.
- 켜진 뒤의 종료 코드는 "작업 방식(action)" 이 정해요: `MOAI_PRE_PUSH`(skip|warn|enforce) → 없으면 `git-strategy.yaml` 의 현재 모드(`git_strategy.mode`, 기본 `team`)의 `hooks.pre_push`(기본 `warn`). 위반이 있어도 `warn` 이면 보고만 하고 0 으로 끝나요. `enforce` 일 때만 2 로 끝나요. 모드가 알 수 없는 값이면 `enforce` 예요.
- 규약 이름이 내장 3종(`conventional-commits`, `angular`, `karma`)이 아니고 `auto` 도 아니면 오류로 1 이에요.

## Non-goals

- 규약을 사용자 정의 패턴으로 늘리지 않아요(Go 빌드에도 내장 3종뿐이에요).
- 훅 파일(`.git/hooks/pre-push`)과 다른 저장소의 설정을 바꾸지 않아요.
- Go 빌드가 오류를 꾸며 그리는 상자(`ERROR` 상자)는 따라 그리지 않아요: 본문과 종료 코드가 계약이에요.

## Requirements

- REQ-1: When `moai hook pre-push` reads commit subjects on standard input and enforcement is on, the check shall report each subject that violates the active convention on stderr and exit 2 only when the pre-push action is `enforce`. · failure: none
- REQ-2: When enforcement is off (`MOAI_ENFORCE_ON_PUSH` unset and `git_convention.validation.enforce_on_push` not true) or the action is `skip`, the check shall read nothing and exit 0 silently. · failure: none
- REQ-3 (REQ-1): IF the named convention is not a built-in name and not `auto`, THEN the check shall print the load error on stderr and exit 1.
- REQ-4: When the convention is `auto`, the check shall pick the built-in convention that the most subjects of the last `auto_detection.sample_size` commits follow, and the fallback convention when the history is unreadable, empty or below `auto_detection.confidence_threshold`. · failure: none
- REQ-5: When no subject is left after trimming, the check shall print `No commit messages to validate.` on stdout and exit 0. · failure: none
- REQ-6: When every subject follows the convention, the check shall print `All <n> commit(s) follow <name> convention.` on stdout and exit 0. · failure: none
- REQ-7: When a `git-convention.yaml` value is outside what the Go loader accepts (convention not in the four names, fallback not in the three conventions, negative sample size or length, threshold outside 0..1), the check shall ignore that file and the profile, using the defaults and the action `enforce`. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-5, REQ-6, REQ-7): the exit code, stdout and stderr equal the Go build's on the recorded inputs.
  verify: `cargo test --locked --offline --test quality_gate_parity prepush`
- AC-2 (REQ-1, REQ-4, REQ-7): conventions, patterns, length, type and scope rules, auto detection and the action order.
  verify: `cargo test --locked --offline --test quality_gate_cli prepush`

## Scope

- write: `.moai/specs/SPEC-20261120-rust-pre-push/**`, `src/quality_gate/pre_push.rs`, `src/quality_gate/mod.rs`, `tests/quality_gate_cli.rs`, `tests/quality_gate_parity.rs`
- read-only: the previous generation's source (`moai-adk-go`: `internal/cli/hook_pre_push.go`, `internal/git/convention/**`, `internal/config/**`), `src/statusline/config.rs`
- forbidden: `src/main.rs`, `src/lib.rs`, `src/help.rs`, any `.git/hooks` file, any write in the real HOME

## Plan

1. RED tests (`prepush_` cases in `quality_gate_cli.rs` and the Go-recorded table in `quality_gate_parity.rs`).
2. `pre_push.rs`: config and environment, built-in conventions with hand-written matchers (the crate has no regex dependency), validation, auto detection, report text, exit codes.
3. The binary word `hook pre-push` is wired with the `gate` word once `src/main.rs` is free.

## Risks

- The built-in patterns are matched by hand, not by a regex engine; the recorded Go table and a differential test over generated subjects guard the equivalence.
- Go picks a convention at random when two score equally in `auto` (map order); this build picks the first of `conventional-commits`, `angular`, `karma`. Since `conventional-commits` accepts everything the other two accept, its score is never lower.
- Go gives up on the whole config when one value is refused and then enforces (no profile). This build does the same, silently, because Go prints nothing either; a typo in `git-convention.yaml` therefore turns a warning hook into an enforcing one.
- A `git-strategy.yaml` outside the supported YAML subset reads as absent: the defaults (`warn`) then apply, with a warning on stderr; Go would read it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Deterministic tie order in `auto` | Go's map order is random; a hook must not flip between runs | Randomise |

## Result

(written after the checks)
