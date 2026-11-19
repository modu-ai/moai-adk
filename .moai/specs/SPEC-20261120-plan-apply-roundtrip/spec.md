---
id: SPEC-20261120-plan-apply-roundtrip
title: "project-plan result applies unchanged — plan-only answer fields are ignored"
status: draft
revision: 1
created: 2026-11-20
---

# 계획 결과를 그대로 apply 에 — 계획 전용 답 필드 무시 (왕복)

## Intent

지침은 `project-plan` 의 결과를 그대로 `project-apply` 에 넘기라고 해요. 그런데 `profile::Plan` 이
`deny_unknown_fields` 여서, 계획 답이 더하는 계획 전용 필드(`n`, `work_flow_remote`) 때문에 역직렬화가
거절해요(work_flow 프로필이 있는 실행에서 발견). 이 조각은 `project-apply` 가 그 두 필드를 무시하게
해요: ops 층에서 입력을 Value 로 읽어 계획 전용 키를 걷어낸 뒤 `Plan` 으로 역직렬화합니다. 다이제스트는
선택 필드 묶음으로 계산되므로(`profile::plan_digest`) 걷어낸 키는 어차피 계산에 없고, 걷어내지 않은
다른 알 수 없는 키는 여전히 거절됩니다.

## Non-goals

- `profile.rs` 의 Plan 구조 변경 — 그 파일은 다른 카드의 예약 중(결정 노트 decision-ops-layer-fix-b).
- 계획 전용 필드의 값 검증 — 무시가 계약(값이 필요하면 계획 화면에서 이미 보였다).
- `n`·`work_flow_remote` 외의 키 허용.

## Requirements

- REQ-1: When `project-apply` reads a plan whose JSON carries the plan-only answer fields `n` or `work_flow_remote`, the command shall drop exactly those two keys before deserializing into the plan type and go on unchanged · failure: none — any other unknown key is still refused with the same message as before.
- REQ-2: When a work_flow profile plans with an integration branch that fails the branch-name check, the plan shall answer with `work_flow_remote` warnings while the unchanged plan result applies · failure: none — the warnings stay in the plan answer; apply never needs them.

## Acceptance

- AC-1 (REQ-1, REQ-2): 왕복 시험 하나 — work_flow 프로필(remote 점검이 경고를 내는 안전하지 않은 통합 가지 이름)로 `project-plan` 을 돌리고 그 결과를 한 글자도 바꾸지 않고 `project-apply` 에 넘기면 적용된다.
  verify: `cargo test --locked --offline --test profile_apply_engine roundtrip`
- AC-2: 이웃 검사.
  verify: `cargo test --locked --offline --test profile_apply_engine`; `cargo test --locked --offline --test profile_scopes_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ops/project_apply.rs`, `tests/profile_apply_engine.rs`, `.moai/specs/SPEC-20261120-plan-apply-roundtrip/spec.md`
- read-only: `src/profile.rs`(Plan·plan_digest·plan_work_flow_remote), `src/ops/project_plan.rs`(답이 키를 더하는 자리)
- forbidden: `src/profile.rs` 편집, 다른 알 수 없는 키 허용, 실제 원격 저장소, 실제 HOME

## Plan

1. RED: `tests/profile_apply_engine.rs` 에 왕복 시험 하나(work_flow 프로필, 안전하지 않은 가지 이름 — 경고는 나오고 git 은 안 돈다).
2. GREEN: `src/ops/project_apply.rs` — 입력을 Value 로 읽어 `n`·`work_flow_remote` 키를 걷어내고 `Plan` 으로 변환.
3. 이웃 검사(AC-2)와 가벼운 검사.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 무시는 ops 층에서(입력 Value 키 제거) | profile.rs 예약 회피 + 다이제스트 불변 | Plan 구조에 선택 필드로 |
| D-2 | 정확히 두 키만 걷어낸다 | 계획 답이 더하는 전부(ops/project_plan.rs 26-27) — 그 밖은 여전히 거절 | 필드 목록 확장 |

## Risks

- 계획 답이 나중에 키를 더 하면 같은 거절이 돌아온다 — 그때 이 목록에 더하는 조각(계획 답과 한 짝).
- 키 제거가 다이제스트 검증을 약화하지 않는다(다이제스트는 원래 그 키를 안 본다) — 시험이 지킨다.

## Result

(채우는 중 — RED·GREEN·검사 기록.)
