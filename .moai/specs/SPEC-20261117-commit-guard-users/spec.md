---
id: SPEC-20261117-commit-guard-users
title: issue/<n> 커밋 가드는 레인에만 — 일반 사용자는 막지 않아요
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

전수조사(2026-11-17)가 찾은 문제: MoAI hook 묶음이 모든 사용자에게 설치하는 두 가드 — `hook-dispatch --event
tool-preflight`(`src/hook_engine.rs`)와 `claim-verify . --hook`(`src/claim_verify.rs` `hook_cli`) — 가
`issue/<n>-…` 가지의 커밋에 `Moai-Claim`이 없으면 거절해요. MoAI 상태가 없는 프로젝트의 일반 사용자도 막혀요.
이 가드는 레인(에이전트) 안전장치예요. 이 SPEC은 가드를 레인이 실제로 도는 clone에서만 켜요. 레인 판별은 이미 있는
신호인 clone의 세션 등록부(`<clone common dir>/.moai-cli/sessions/`, `lane_registry::live_session_count`: 등록된
lease나 launcher가 잡은 `.live` 잠금)로 하고, 새 전역 상태는 만들지 않아요.

## Requirements

- REQ-1: When a commit without a `Moai-Claim` token runs on an `issue/<n>` branch of a clone without a live lane session, both commit guards shall allow it. · failure: none — 레인이 없으면 지킬 대상이 없어요.
- REQ-2: When such a commit runs in a clone with a live lane session and no local claim record, both commit guards shall refuse it as before. · failure: none — 거절 자체가 실패 경로예요.
- REQ-3 (REQ-1): IF the clone's session registry cannot be read, THEN both commit guards shall treat the clone as running a lane.
- REQ-4: The commit guards shall leave every branch outside `issue/` unguarded, `spec/` branches included.
- REQ-5: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-commit-guard-users/spec.md`, tracked by git.

## Acceptance

- AC-1 (REQ-1): 임시 HOME·저장소에서 `issue/12-fix-typo`의 일반 커밋을 두 가드가 모두 허용하고 실제 커밋도
  돼요; lease가 끝난 등록만 있는 clone도 같아요(lease 경계).
  verify: `cargo test --locked --offline --test commit_guard_users`
- AC-2 (REQ-2): 레인을 등록한 clone에서 토큰 없는 커밋은 두 가드가 거절하고(preflight deny, `claim-verify --hook`
  exit 2), 토큰이 있으면 preflight가 허용해요. 기존 레인 가드 시험은 레인을 등록한 fixture로 그대로 통과해요.
  verify: `cargo test --locked --offline --test commit_guard_users --test hook_dispatch_cli --test claim_verify_cli --test hook_status_cli --test hook_mode_cli --test hooks_cli --test git_env_scrub`
- AC-3 (REQ-3): `sessions` 등록부 폴더가 symlink라 읽을 수 없는 clone에서는 일반 커밋도 두 가드가 거절해요.
  verify: `cargo test --locked --offline --test commit_guard_users an_unreadable_registry_keeps_the_guard_on`
- AC-4 (REQ-4): `spec/` 가지는 레인 유무와 관계없이 두 가드가 허용해요.
  verify: `cargo test --locked --offline --test commit_guard_users spec_branches_are_not_guarded_with_or_without_a_lane`
- AC-5 (REQ-5): verify: `git ls-files .moai/specs/SPEC-20261117-commit-guard-users` → 이 파일; `moai spec check --all --project .` → exit 0

## Scope

- write: `src/claim_verify.rs`, `src/hook_engine.rs`, `tests/commit_guard_users.rs`(새 파일), `tests/common/mod.rs`,
  `tests/hook_dispatch_cli.rs`, `tests/claim_verify_cli.rs`, `tests/git_env_scrub.rs`, `tests/hook_mode_cli.rs`,
  `tests/hook_status_cli.rs`, `tests/hooks_cli.rs`, `.moai/specs/SPEC-20261117-commit-guard-users/spec.md`
- read-only: `src/lane_registry.rs`, `src/session.rs`, `src/hooks.rs`

## Plan

1. RED: `tests/commit_guard_users.rs`(일반 사용자 허용, 레인 거절, 토큰 허용, lease 끝남, `spec/`).
2. GREEN: `claim_verify::lane_clone`과 두 가드의 한 줄씩.
3. 기존 레인 시험 fixture에 레인 등록(`tests/common` `register_lane`).
4. 검사, 커밋 뒤 mutation(3개 이하), Result.

## Decisions

- D1 레인 판별 = clone의 세션 등록부에 살아 있는 세션이 있음(`live_session_count > 0`). 이유: 레인은 시작할 때
  `session-register`로 등록하고 깨어날 때마다 갱신하며, launcher는 `.live` 잠금을 쥐어요. 이 clone(`moai-adk`)에는
  claim journal이 없고 등록부만 있어서 journal로는 판별할 수 없었어요. 호스트 세션 하나하나를 레인과 연결하는
  신호(환경 변수 등)는 없어서 판별은 clone 단위예요: 레인이 도는 clone에서는 사람의 `issue/<n>` 커밋도 지금처럼 막혀요.
  되돌리기: 두 가드에서 `lane_clone` 검사 줄을 지워요.
- D2 등록부를 읽을 수 없으면 레인으로 봐요(가드 유지, 안전 쪽).
- D3 두 가드 모두에 같은 규칙을 넣었어요: `hooks.rs`가 둘 다 모든 사용자에게 설치해요.
- D4 거절 경로는 git 호출이 3번에서 5번으로 늘어요(등록부 위치를 찾는 `rev-parse` 2번). 토큰이 있는 허용 경로는
  그대로 1번이에요(`hook_dispatch_cli`의 호출 수 시험을 5로 고침).

## Result

(구현 뒤 채워요.)
