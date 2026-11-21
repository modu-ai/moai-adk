---
id: SPEC-20261121-plan-shows-dropped-manifests
title: "The profile plan names the recorded manifests that applying it would drop"
status: draft
revision: 1
created: 2026-11-21
---

# The profile plan names the recorded manifests that applying it would drop

## Intent

모노레포에서 프로필을 다시 적용하면 `detected.manifests` 가 55 개에서 1 개로 줄었어요(탐색이 부분적이었어요; 탐색은 SPEC-20261121-detect-walk-breadth 가 고쳐요). 그런데 미리보기는 `detected` 키가 "바뀐다"고만 말하고 어떤 manifest 가 사라지는지는 말하지 않아서, 사용자는 줄어드는 것을 보지 못한 채 적용했어요. 이 SPEC 은 탐색이 맞든 틀리든 같은 일이 조용히 일어나지 않게 해요: 계획이 `detected` 키를 적용하면서 기록돼 있던 manifest 가 새 목록에 없으면, 그 이름들을 계획과 업데이트 미리보기에 보여 줘요. 적용은 사용자가 그 계획(의 digest)을 확인해야만 되고 digest 가 이 목록을 덮으니, 보여 준 것이 곧 사용자의 결정이에요.

## Non-goals

- 기록된 목록을 몰래 합치거나 되돌리지 않아요(보여 주기만 해요).
- 탐색(`project-detect`)을 바꾸지 않아요(SPEC-20261121-detect-walk-breadth).
- 사용자가 손으로 고쳐서 충돌로 남은 `detected` 키는 덮어쓰지 않으니 이 목록도 비어 있어요.
- 다른 키(`tools`, `languages` 등)가 줄어드는 것은 이 SPEC 의 대상이 아니에요.

## Requirements

- REQ-1: When the profile plan applies the `detected` key over a recorded profile, the plan shall carry `dropped_manifests`, the sorted `detected.manifests` entries of the recorded profile that the new `detected.manifests` leaves out, and shall leave the field out when none is left out.
- REQ-2: The plan digest shall cover `dropped_manifests` when it is not empty and shall be computed as before when it is empty.
- REQ-3: When the update plan's profile step has dropped manifests, the step `reason` shall say how many would drop and the step `preview` shall carry the list.
- REQ-4 (REQ-2): IF a plan's `dropped_manifests` is not what the current files give, THEN `project-apply` shall refuse the plan as stale or altered and write nothing.

## Acceptance

- AC-1 (REQ-1): 기록된 `detected.manifests` 가 `apps/a/package.json`, `apps/b/go.mod`, `package.json` 일 때 새 목록이 `package.json` 뿐이면 `project-plan` 이 `dropped_manifests` 로 앞의 둘을 정렬해 보여요.
  verify: `cargo test --locked --offline --test project_update_drops_cli the_project_plan_names_the_recorded_manifests_a_smaller_detected_list_drops`
- AC-2 (REQ-3): 같은 상황에서 `project-update-plan` 의 `profile` 단계가 `ready` 이고 `reason` 에 "2 recorded manifests would drop" 이 있고 `preview.dropped_manifests` 가 그 둘이에요.
  verify: `cargo test --locked --offline --test project_update_drops_cli the_update_plan_profile_step_says_how_many_manifests_would_drop`
- AC-3 (REQ-1): 목록이 그대로이거나 늘기만 하면 `dropped_manifests` 가 계획에 없고 `reason` 은 전과 같아요.
  verify: `cargo test --locked --offline --test project_update_drops_cli a_list_that_only_grows_or_stays_drops_nothing`
- AC-4 (REQ-2, REQ-4): `dropped_manifests` 를 지운 계획은 `project-apply` 가 "stale or altered plan" 으로 거절하고 프로필 파일이 그대로예요.
  verify: `cargo test --locked --offline --test project_update_drops_cli a_plan_with_the_dropped_list_removed_is_refused_at_apply`
- AC-5 (REQ-1): 사용자가 손으로 고쳐서 충돌인 `detected` 키는 `dropped_manifests` 가 없어요.
  verify: `cargo test --locked --offline --test project_update_drops_cli a_hand_edited_detected_key_drops_nothing`
- AC-6: 이웃 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test profile_apply_engine --test profile_plan_cli --test project_update_plan_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-plan-shows-dropped-manifests/spec.md`, `src/profile.rs`(`Plan` 의 새 필드, `plan` 과 `plan_digest`), `src/project_update.rs`(`profile_step` 의 `reason` 과 `preview`), `tests/project_update_drops_cli.rs`(새 파일)
- read-only: `src/project_detect.rs`, `src/ops/project_plan.rs`(계획을 그대로 내보내서 바꿀 곳이 없어요)
- forbidden: 탐색 코드, 다른 키의 비교, 기록된 목록을 몰래 바꾸는 일

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: `tests/project_update_drops_cli.rs` 에 AC-1~5 시험을 먼저 써서 실패하는 것을 봐요(파일을 쥔 카드가 풀 때까지는 여기까지예요).
3. GREEN: `profile::plan` 이 `detected` 가 적용될 때(`delta.changed` 나 `delta.removed` 에 있을 때) 줄어드는 항목을 구해 `Plan.dropped_manifests` 에 넣고 digest 에 더해요(비어 있으면 digest 를 전과 같게 해요). `profile_step` 이 `reason` 과 `preview` 에 보여요.
4. 확인: 새 시험, 이웃 시험, fmt, clippy, 변이, `moai check affected`.

## Risks

- `Plan` 에 필드가 하나 더해져요. 이 구조체는 `deny_unknown_fields` 라서 옛 계획 JSON 은 그대로 읽히고(필드가 없으면 빈 목록이에요), 새 계획을 옛 바이너리가 읽으면 거절돼요: 계획은 같은 실행에서 만들고 쓰는 단기 값이라 문제가 없어요.
- digest 계산을 건드려요: 목록이 비어 있을 때 바이트를 바꾸지 않아서 기존 digest 와 이를 박아 둔 시험은 그대로예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 보여 주고 digest 로 덮는다(별도 확인 질문을 더하지 않는다) | 적용은 이미 계획 digest 를 확인해야 해요. 목록이 digest 안에 있으면 보지 않고 적용할 수 없어요 | 필드 빼기 |
| D-2 | 기록된 목록을 대신 합치지 않는다 | 사라진 파일이 정말 지워진 것일 수 있어요. 사용자가 보고 정해요 | 해당 없음 |

## Result

(구현 뒤에 채워요.)
