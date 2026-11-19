---
id: SPEC-20261119-plan-remote-branch
title: "project-plan checks the integration branch on the remote"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-spec-take-profile-base
---

# project-plan checks the integration branch on the remote

## Intent

카드 C-105 조각 2(기준 2). 프로필의 `work_flow.integration_branch`가 원격에 없거나(실제 사고: 로컬에만 있는 가지가 받아들여짐) PR들이 목표하는 보호 가지보다 크게 뒤처졌는데(52 커밋) 아무 검증 없이 저장돼요. 이 SPEC은 `moai project plan`이 그 가지를 원격에서 확인해 결과에 알리고, 위험한 가지 이름(-로 시작·공백·`..`)은 결코 git 인자로 가지 않게 해요.

## Non-goals

- 기준 3–5(session-check 출력·mode 문서 명시·project-apply 도움말) — 이후 조각.
- 값을 거부하는 것(인터뷰가 미리 만든 가지를 허용) — 이 조각은 경고만.
- gh/PR 목록 조회 — 원격 HEAD(기본 가지)와의 ahead/behind 비교가 `ls-remote`+`rev-list` 로 측정 가능한 범위예요.

## Requirements

- REQ-1: When `project plan` previews a profile whose `work_flow.integration_branch` is set and the project folder is a git repository with a remote, the plan answer shall carry a `work_flow_remote` object telling whether the branch exists on that remote (`git ls-remote`). The check runs in the `project-plan` entry alone — `project-update-plan` shares `plan()` but must not run it (audit B4: the preview must stay off the network).
- REQ-2 (REQ-1): IF the branch does not exist on the remote, THEN `work_flow_remote.warnings` shall say so and name the branch.
- REQ-3 (REQ-1): IF the remote names a default branch (symref) and the integration branch exists there and is behind it, THEN `work_flow_remote.warnings` shall name the behind count.
- REQ-4 (REQ-1): IF the integration branch value is not a safe branch name (leading `-`, whitespace, `..`, over 100 bytes, `HEAD`), THEN the profile validation shall refuse the plan before any git command runs for that name (audit B2: the CLI refuses outright — the in-function refusal branch stays as the second layer for internal callers only).
- REQ-5: When the profile has no `work_flow.integration_branch` or the folder is not a git repository with a remote, the plan answer shall omit `work_flow_remote` entirely (byte-identical output to before this change).
- REQ-6 (REQ-5): IF any git command for this check fails (no remote, network, timeout), THEN the plan shall warn `work_flow_remote.unavailable` with the reason and succeed as a plan.

## Acceptance

- AC-1 (REQ-1, REQ-2): 원격에 있는 가지 → `on_remote: true`·경고 없음; 원격에 없는 가지 → `on_remote: false`·경고 문구. 임시 저장소+bare 원격 하네스.
  verify: `cargo test --offline --locked --test profile_plan_cli`
- AC-2 (REQ-3): 통합 가지가 기본 가지보다 뒤처진 임시 원격 → behind 수치 경고.
  verify: `cargo test --offline --locked --test profile_plan_cli`
- AC-3 (REQ-4): `-x`·`a b`·`a..b`·100바이트 초과·`HEAD` 가지 이름 → 프로필 검증이 plan을 거절(exit≠0, "invalid or nonportable profile")하고 하네스 로깅 git 으로 어떤 git 호출도 안 난 것을 확인(감사 B2 정정: 거부는 검증 계층에서).
  verify: `cargo test --offline --locked --test profile_plan_cli`
- AC-4 (REQ-5, REQ-6, REQ-1…REQ-4): work_flow 없음/원격 없음 → `work_flow_remote` 키 부재(기존 출력 불변); git 실패 → unavailable 경고. 이웃: `cargo test --offline --locked --lib`; `cargo fmt --check`; `cargo clippy --offline --locked --all-targets -- -D warnings`
  verify: `cargo test --offline --locked --test profile_plan_cli`; `cargo test --offline --locked --lib`

## Scope

- write: `src/profile.rs`, `tests/profile_plan_cli.rs`(신규), `.moai/specs/SPEC-20261119-plan-remote-branch/spec.md`
- read-only: `src/ops/project_plan.rs`, `src/identity.rs`, `src/spec_take.rs`(git 호출 형태)
- forbidden: 실제 HOME·원장·키, 네트워크 쓰기, skills/**

## Plan

1. RED: `tests/profile_plan_cli.rs` — 임시 HOME+프로젝트 저장소+bare 원격 하네스로 AC 1–4 시험 먼저 실패 확인.
2. GREEN: `profile::plan` 끝에 `work_flow_remote` 계산 추가(안전 이름 검사→`branch_name` 재사용, ls-remote `--end-of-options`, `identity::scrub_git_env`, 실패는 unavailable).
3. 이웃: profile_plan_cli·--lib·fmt·clippy; 준비 배터리(lib 전체·home_isolation_scan·tip trailer-check).

## Risks

- plan 출력 파식을 고정하는 기존 픽스처 — `work_flow` 없는 프로필은 키 자체를 생략(skip_serializing_if)해 불변 유지(REQ-5).
- git 인자 주입 — 값은 `branch_name()` 통과 후에만 인자로; ls-remote 에 `--end-of-options`(REQ-4·배차 지시).
- 원격 기본 가지 판별 — `ls-remote --symref HEAD` 출력 파싱; 없으면 behind 검사 생략(경고 없음).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 경고만, 거절 없음 | 카드 문구 'warns or refuses' 중 인터뷰 사전 생성 허용 쪽; 거절은 관측 후 이후 조각 | refuse 로 전환 |
| D-2 | 출력 채널은 `work_flow_remote` 선택 객체 | 기존 plan JSON 불변(REQ-5) | warnings 공통 채널 |
| D-3 | behind 측정은 원격 HEAD 기준 `rev-list --left-right --count` | PR 목록 없이 측정 가능; 원격 객관 기준 | gh 조회 |

## Result

(GREEN 뒤 AC마다 한 줄)
