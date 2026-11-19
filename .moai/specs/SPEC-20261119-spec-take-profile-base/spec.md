---
id: SPEC-20261119-spec-take-profile-base
title: "spec take reads the profile's integration branch as its default base"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261118-spec-take
---

# spec take reads the profile's integration branch as its default base

## Intent

카드 C-105 조각 1(기준 1). `.moai-cli/project.yaml`(rust 분류 프로젝트는 `.moai/project.yaml`)의 `work_flow.integration_branch`는 m-flow 통합 가지를 기록하지만 아무 명령도 읽지 않아요 — `moai spec take`는 `--base`가 없으면 지금 가지를 base로 써요(SPEC-20261118-spec-take D-1이 프로필 읽기를 되돌리기로 남겨둔 것). 이 SPEC은 take가 `--base` 없이 새 SPEC 가지를 만들 때 프로필의 통합 가지를 기본 base로 읽게 하고, 어떤 base를 왜 썼는지 결과에 알려요. 우선순위: `--base` > (가지에) 기록된 base > 프로필 > 현재 가지. 이 조각은 src·tests 만 바꿔요.

## Non-goals

- 기준 2–5(원격 존재 검증, session-check 출력, mode 문서 명시, project-apply 도움말) — 다음 조각.
- skills/** 와 source-manifest.json (배차 제외).
- 통합 가지의 원격 존재·낙후 검증(기준 2).

## Requirements

- REQ-1: When `moai spec take` starts a SPEC branch without `--base` and that branch has no recorded base, and the project profile sets `work_flow.integration_branch` to a non-empty value, the command shall start the branch from that integration branch.
- REQ-2 (REQ-1): IF the profile has no `work_flow.integration_branch` (unset, empty, or no profile can be read), THEN the command shall use the current branch as its base, as before.
- REQ-3 (REQ-1): IF the integration branch does not name a commit, THEN the command shall refuse with the invalid-base refusal and change nothing.
- REQ-4: When the base is decided, the take result shall report the base and its source — `given`, `recorded`, `profile` or `current-branch`.
- REQ-5 (REQ-4): IF no base can be decided (no `--base`, no recorded base, no profile value, and the checkout is detached or already on the SPEC branch), THEN the command shall refuse as before and change nothing.

## Acceptance

- AC-1 (REQ-1, REQ-4): 프로필이 통합 가지를 정하면 새 SPEC 가지가 거기서 시작하고 결과 JSON 이 `base` 와 `base_source: "profile"` 를 알려요.
  verify: `cargo test --offline --locked --test spec_take_cli`
- AC-2 (REQ-2, REQ-4): 프로필이 없거나 값이 비면 지금처럼 현재 가지(`base_source: "current-branch"`); `--base` 는 `given`, 재수행의 기록된 base 는 `recorded`.
  verify: `cargo test --offline --locked --test spec_take_cli`
- AC-3 (REQ-3, REQ-5): 프로필 통합 가지가 커밋이 아니면 invalid-base 거절(아무 변화 없음); base 결정 불가 상황은 기존 거절 유지.
  verify: `cargo test --offline --locked --test spec_take_cli`
- AC-4 (REQ-1…REQ-5): 이웃 검사 — spec_take_cli 전체·`cargo fmt --check`·clippy; 도웰말 무변경.
  verify: `cargo test --offline --locked --test spec_take_cli`; `cargo fmt --check`; `cargo clippy --offline --locked --all-targets -- -D warnings`

## Scope

- write: `src/spec_take.rs`, `tests/spec_take_cli.rs`, `.moai/specs/SPEC-20261119-spec-take-profile-base/spec.md`
- read-only: `src/profile.rs`, `src/state_root.rs`, `src/spec_doc.rs`
- forbidden: skills/**, source-manifest.json, 실제 HOME·원장·키, 네트워크

## Plan

1. RED: `tests/spec_take_cli.rs` 에 (a) 프로필 base 사용+보고 (b) 무프로필 현행 유지 (c) 프로필 base 비커밋 거절 (d) base_source 4값 시험 추가 → 실패 확인.
2. GREEN: `src/spec_take.rs` base 해석에 `profile::read_profile` tolerantly 읽어 `work_flow.integration_branch` 끼우고 결과 JSON 에 `base_source` 추가.
3. 이웃: spec_take_cli 전체·fmt·clippy. help 무변경 확인.

## Risks

- 프로필 읽기가 take를 거르면 안 돼요 → 읽기 실패는 None (REQ-2 경로), take 본래 오류와 섞지 않아요.
- state 폴더는 프로젝트 분류에 따라 `.moai`/`.moai-cli` → `profile::read_profile` 의 경로 규칙을 그대로 써요(직접 경로 계산 안 함).
- 기존 시험의 `base == "main"` 단언은 무프로필 경로라 그대로 통과해야 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 우선순위 `--base` > recorded(재수행) > 프로필 > 현재 가지 | 기존 재수행 의미 유지; 배차가 '프로필을 기본'으로 지정 | 프로필 > recorded |
| D-2 | 보고는 결과 JSON 의 `base_source` 한 필드(값 4개) | criterion '어떤 base 를 왜' — 값 자체가 이유 | 별도 reason 문장 |
| D-3 | 프로필 읽기 실패·빈 값은 현재 가지로 폴백(거절 아님) | 프로필 없는 프로젝트의 현행 동작 보존 | 거절로 변경 |

## Result

- AC-1: met — `take_without_base_reads_the_profile_integration_branch`: 프로필 통합 가지에서 시작·기록, `base_source: "profile"` (RED 먼저: 구현 전 4 실패 — base_source 부재·프로필 무시).
- AC-2: met — 무프로필 `current-branch`(기존 시험 보강), `given`·`recorded` 보고, 빈 integration_branch 폴백, 재수행 recorded > 나중 프로필 값(D-1) 시험.
- AC-3: met — 프로필 base 비커밋 → invalid-base 거절·상태 무변화; detached 무base 기존 거절 유지(분리 시험).
- AC-4: met — `cargo test --offline --locked --test spec_take_cli` 14 통과(TEST_EXIT=0), `cargo fmt --check`(FMT_EXIT=0), clippy -D warnings(CLIPPY_EXIT=0), 도움말 무변경; 변이 3/3 KILLED(프로필 분기 제거·우선순위 뒤집기·base_source 미보고, target/c105/mut.log — M2는 픽스처 결함(프로필을 가지에만 커밋)을 고친 뒤 올바른 단언에서 사살 확인).
