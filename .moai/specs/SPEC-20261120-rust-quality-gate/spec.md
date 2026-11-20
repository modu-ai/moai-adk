---
id: SPEC-20261120-rust-quality-gate
title: "moai gate and moai hook pre-push run the same commit checks in Rust"
status: draft
revision: 1
created: 2026-11-20
---

# moai gate and moai hook pre-push run the same commit checks in Rust

## Intent

사용자 결정(2026-11-19): `moai` 는 Rust 만 실행해요. 지금 Go 빌드의 `moai gate` 를 13개 저장소의 git pre-commit 훅이 `MOAI_PRECOMMIT=1 moai gate` 로 불러요. Rust 에는 `gate` 가 없어서(`moai-cli gate` → 종료 코드 2) Go 위임이 사라지면 그 저장소의 모든 커밋이 막혀요. 이 SPEC 은 Go `gate`(`internal/cli/gate.go`, `internal/hook/quality`)가 하는 검사를 읽기 전용으로 읽고 Rust 에 같은 검사·같은 종료 코드로 옮겨요. 검사를 건너뛰는 빈 gate 는 만들지 않아요: 건너뛰는 경우(설정에서 끔, 도구 없음 등)는 Go 와 같은 이유로, 요약에 이름을 적어서만 건너뛰어요.

사용자 결정 추가(2026-11-20): 같은 13개 저장소의 git pre-push 훅이 `moai hook pre-push` 로 커밋 제목 규약 검사를 불러요. Rust 는 `moai hook <이벤트>` 를 조용히 0 으로 삼켜서(`src/main.rs` 의 호환 처리) 지금 이 검사는 아무것도 하지 않아요. 그 검사는 자식 SPEC `SPEC-20261120-rust-pre-push` 가 같은 호출(표준 입력에 제목을 한 줄씩)과 같은 종료 코드로 옮겨요(요구사항 수 상한 때문에 나눴어요). 훅 파일은 저장소마다 사용자 몫이라 건드리지 않아요.

Go 의 gate 는 크기가 커서(핵심 약 4,000줄) 조각으로 나눠요.

- **조각 3a (이 SPEC, 1차 개정)**: 명령 본체 — 프로젝트 폴더 결정, `gate.yaml` 읽기, pre-commit 표식에 따른 건너뛰기, 16개 언어 표와 감지(루트와 4단계 아래), 단계 실행(vet·lint·test, 시간 제한, 프로세스 그룹 종료, 깃 환경 정리), 요약, 전체 시간 상한, 실행 잠금, 종료 코드.
- **조각 3b**: 노드 패키지 관리자·스크립트 판별(`resolveNode*`), 타입체크 축(`typecheck`), 이 둘이 쓰는 JSONC 읽기.
- **조각 3c**: ast-grep 규칙 검사(`astgrep_gate.go`), 그래프 신선도 검사(`runGraphFreshnessStep`).
- **pre-push (조각 3d, 자식 SPEC)**: `moai hook pre-push` — `SPEC-20261120-rust-pre-push` 가 맡아요.
- **연결**: `src/main.rs` 에 `gate` 를 clap 해석 전에 처리하는 한 줄(`hook` 과 같은 방식)과 `mod` 선언. `main.rs` 는 다른 일이 쥐고 있어서, 풀리는 대로 하는 작은 조각이에요.

3a 가 아직 하지 않는 축(노드 프로젝트의 eslint·biome·oxlint·typecheck·npm test 전부 — 어떤 스크립트·패키지 관리자를 고를지가 3b 의 일이라서 틀린 명령으로 커밋을 막느니 이름을 적고 건너뛰어요 — 그리고 ast-grep, 그래프 신선도)은 요약에 `not run — not ported yet` 로 이름을 적어요 — 말없이 지나가지 않아요.

## Non-goals

- 다른 저장소의 훅·`.git/hooks` 를 바꾸지 않아요(저장소마다 사용자 몫).
- Go `gate` 의 플래그·새 동작을 더하지 않아요: 같은 입력에 같은 검사와 종료 코드가 목표예요.
- Windows 의 실행 잠금과 프로세스 그룹: 유닉스만 구현하고 Windows 는 잠금 없이 실행하며 그 사실을 말해요(Go 도 Windows 에서 같은 한계를 문구로 적어요).
- 도움말 목록·골든 파일(`src/help.rs` 와 `tests/fixtures/help`)은 건드리지 않아요: 이 명령은 git 훅이 부르는 숨은 명령이에요.

## Requirements

- REQ-1: When `moai gate` runs with `MOAI_PRECOMMIT=1` and the project's `gate.pre_commit.enabled` is not true, the gate shall skip every project-wide step, print one notice naming the `gate.pre_commit.enabled` key on stderr and exit 0. · failure: none
- REQ-2: When `moai gate` runs, the gate shall read `<project>/.moai/config/sections/gate.yaml` over the built-in defaults, with the project directory taken from `CLAUDE_PROJECT_DIR` when it is set and from the working directory otherwise. · failure: none
- REQ-3 (REQ-2): IF `gate.yaml` cannot be read as the YAML subset the reader supports, THEN the gate shall warn on stderr and use the built-in defaults.
- REQ-4: When the gate is enabled, it shall detect every language whose marker file sits at the project root or at most four levels below it, skipping dependency and build folders and not descending below a matched module, and run each language's vet, lint and test steps in the directory that holds the marker. · failure: none
- REQ-5: When a step's optional tool is missing from PATH, none of its config files exists, no staged file carries its extensions, or the project holds no source file it can check, the gate shall skip the step and record the reason. · failure: none
- REQ-6: When a step exits non-zero, the gate shall stop, print `quality gate failed: <step>` and the step's output on stderr and exit 1, except that a pytest exit 5 and a dotnet restore failure pass as Go passes them. · failure: none
- REQ-7: When a step outlives its time limit, the gate shall end the step's whole process group, print the timeout message naming the step and its limit on stderr and exit 1. · failure: none
- REQ-8: When the run ends, the gate shall print the step summary on stderr: one line for every configured step with its outcome (`executed`, `skipped`, `disabled` or `not reached`), its reason or its command, and one line for each axis this build does not run yet naming it as `not run — not ported yet`. · failure: none
- REQ-9: While another gate run holds the project's lock file, the gate shall wait up to `gate.timeouts.lock_wait` seconds and then run unserialized and say so, never letting the lock decide the exit code. · failure: none
- REQ-10: When `gate.enabled` is false, the gate shall print nothing and exit 0 before any step runs. · failure: none
- REQ-11: When a step runs, the gate shall run it without a shell, in its module's root, with the git repository-location variables removed from its environment. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-10): the marker with the default config prints the notice and exits 0; `enabled: false` prints nothing and exits 0.
  verify: `cargo test --locked --offline --test quality_gate_cli the_precommit_marker_skips_unless_opted_in`
  verify: `cargo test --locked --offline --test quality_gate_cli a_disabled_gate_passes_silently`
- AC-2 (REQ-2, REQ-3): project dir order and config reading; defaults on an unreadable file with a warning.
  verify: `cargo test --locked --offline --test quality_gate_config`
- AC-3 (REQ-4): root, nested, skipped-folder, depth-limit and glob-marker detection.
  verify: `cargo test --locked --offline --test quality_gate_steps detection`
- AC-4 (REQ-5, REQ-6, REQ-7, REQ-11): skip reasons, failure message and exit code, pytest exit 5, dotnet restore, timeout with group kill, no shell, git variables removed.
  verify: `cargo test --locked --offline --test quality_gate_steps steps`
- AC-5 (REQ-8, REQ-9): summary lines, the `not ported yet` axes, lock wait and unserialized run.
  verify: `cargo test --locked --offline --test quality_gate_cli summary`
  verify: `cargo test --locked --offline --test quality_gate_cli lock`
- AC-6 (REQ-4, REQ-6): the same fixture projects give the same exit code and the same step names and reasons as the Go build, recorded as a table.
  verify: `cargo test --locked --offline --test quality_gate_parity`
- AC-7: neighbours.
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261120-rust-quality-gate/**`, `src/quality_gate.rs`, `src/quality_gate/**`, `tests/quality_gate_cli.rs`, `tests/quality_gate_config.rs`, `tests/quality_gate_steps.rs`, `tests/quality_gate_parity.rs`
- read-only: the previous generation's source (`moai-adk-go`: `internal/cli/gate*.go`, `internal/hook/quality/**`, `internal/config/**`), `src/statusline/config.rs`, `src/edit_diagnostics.rs`
- forbidden: `src/main.rs`, `src/lib.rs`, `src/help.rs` and the help goldens (other work holds them; the wiring is a later step), any `.git/hooks` file, any write in the real HOME

## Plan

1. SPEC, then RED tests (the five test files; modules included by path so no wiring file is needed).
2. `src/quality_gate/`: `config`, `table` (16 languages), `detect`, `step` (runner), `summary`, `lock`, `mod` (`run`).
3. Parity table: build fixture projects with fake toolchains on PATH, run `moai.old gate` once, record exit code and the summary shape, assert the Rust gate matches.
4. Checks, mutants, audit; the wiring step when `main.rs` is free.

## Risks

- A `gate.yaml` outside the supported YAML subset (flow collections, sequences) is read as unreadable and gets the defaults plus a warning — Go would read it. The shipped file is block style.
- Until the wiring step the code is compiled by the tests only; a build of `moai-cli` has no `gate` word yet.
- The parity table runs the Go build once on a developer machine; CI carries the recorded table, not the Go build.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Lock file stays `<project>/.moai/state/gate-run.lock` | A Go run and a Rust run on one tree then serialize | Move the path |
| D-2 | Axes not ported yet are named in the summary, not silent | "No empty gate": the user's decision of 2026-11-19 | Remove the line when the axis lands |
| D-3 | The command is hidden from help | It is called by git hooks, not typed; keeps the help goldens untouched | Add it to the help groups |

## Result

(written after the checks)
