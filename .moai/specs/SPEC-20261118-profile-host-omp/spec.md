---
id: SPEC-20261118-profile-host-omp
title: "The project profile records OMP as a host"
status: draft
revision: 1
created: 2026-11-18
---

# The project profile records OMP as a host

## Intent

카드 C-56: `/moai project`에서 사용자가 호스트로 Claude Code, Codex CLI, OMP를 골랐는데 `project-plan`이 `omp`를 거절해서, 세션이 OMP를 결정 기록에만 적었어요
(`src/profile.rs`의 `hosts.len() > 2`와 `claude`·`codex`만 허용하는 검사). OMP는 `omp-hooks`·`omp-agents` 단계로 이미 지원하는 호스트예요(사용자 결정 2026-11-16).
프로필이 `claude`·`codex`·`omp` 셋을 받고, 그 밖의 이름은 계속 거절하고, 호스트별 단계는 고른 호스트에서만 돌게 해요. C-62(OMP를 정식 호스트로)와 결론이 같아요: OMP는 세 번째 호스트예요.

## Non-goals

- `moai omp` 실행기, OMP 쪽 새 기능(C-62), 프로젝트 스킬 폴더(`project_skills`는 `claude`·`codex`만 그대로).
- 프로필이 없거나 깨졌을 때의 호스트 단계 동작 바꾸기: 그때는 지금처럼 모든 호스트 단계를 보여 줘요.

## Requirements

- REQ-1: The profile shall accept `claude`, `codex` and `omp` as hosts, one to three of them.
- REQ-2 (REQ-1): IF the profile lists more than three hosts or a host other than those three, THEN `project-plan` shall refuse it and write nothing.
- REQ-3: When the profile is read, the `project-update-plan` host steps (`codex-hooks`, `codex-config`, `codex-skill` for `codex`; `omp-hooks`, `omp-config`, `omp-agents` for `omp`) shall run only for a host the profile lists.
- REQ-4 (REQ-3): IF the profile does not list a host, THEN that host's steps shall be `skip` with a reason naming the host and carry no printed command.
- REQ-5 (REQ-3): IF the profile is missing or broken, THEN every host step shall be planned as before.

## Acceptance

- AC-1 (REQ-1, REQ-2): 호스트 셋(`claude`, `codex`, `omp`)과 `omp` 하나는 통과하고, 넷과 모르는 이름(`zcode` 포함 셋)은 거절돼요. 경계는 3과 4예요.
  verify: `cargo test --locked --offline --lib profile::tests::three_hosts_with_omp_validate`; `cargo test --locked --offline --lib profile::tests::four_hosts_and_unknown_hosts_are_refused`
- AC-2 (REQ-1): 인터뷰 `hosts` 행이 Claude Code, Codex CLI, OMP를 다중 선택으로 내고 지금 호스트를 권장으로 앞에 둬요.
  verify: `cargo test --locked --offline --test question_bank_hosts`; `python3 -I tools/source_manifest.py check`
- AC-3 (REQ-3, REQ-4, REQ-5): `omp`를 고른 프로필은 `omp-*` 단계를 계획하고, 고르지 않은 프로필은 `omp-*`를 `skip`으로 두며, `codex`도 마찬가지예요. 깨진 프로필은 모든 단계를 그대로 계획해요.
  verify: `cargo test --locked --offline --test project_update_plan_cli host_steps_follow_the_profile_hosts`; `cargo test --locked --offline --test project_update_plan_cli a_profile_with_omp_passes_project_plan`; `cargo test --locked --offline --test project_update_plan_cli a_broken_profile_blocks_its_steps_and_not_the_host_hook_steps`
- AC-4 (REQ-1…REQ-5): 이웃 검사.
  verify: `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test profile_apply_engine`; `cargo test --locked --offline --lib`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/profile.rs`, `src/project_hosts.rs`, `src/project_update.rs`, `tests/project_update_plan_cli.rs`, `tests/profile_apply_engine.rs`,
  `tests/common/go_project_real_run.rs`, `tests/question_bank_hosts.rs`, `skills/moai/references/internal/interview.md`,
  `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261118-profile-host-omp/spec.md`
- read-only: `src/project_skills.rs`, `src/generation.rs`, `src/agents.rs`, `src/hooks.rs`, `src/omp_agents.rs`
- forbidden: 실제 HOME, 네트워크

## Plan

1. 시험 먼저: `profile.rs` 단위 시험(경계 3·4, 모르는 이름)과 `project_update_plan_cli`의 호스트 단계 시험.
2. `validate`: 호스트 수 상한 3, 이름은 `claude`·`codex`·`omp`.
3. 인터뷰 `hosts` 행을 세 호스트 다중 선택으로 고치고 `question_bank_hosts`·manifest 해시를 맞춤. 공용 시험 준비물의 프로필은 세 호스트를 골라 둠(고르지 않으면 호스트 단계가 skip이라 `update_end_to_end_cli`의 순서 기대가 깨져요).
4. `project_hosts::steps`가 고른 호스트 목록을 받아 고르지 않은 호스트의 단계를 `skip`으로 두고, `project_update`가 읽은 프로필의 `hosts`를 넘김.

## Risks

- 지금까지 `claude`만 고른 프로필도 `codex-*`·`omp-*` 단계가 `ready`로 보였어요 → 이제 `skip`이에요. 사용자가 보는 미리보기가 줄어드는 변화라서 PR 본문에 적어요.
- `project_skills`·`generation`은 `claude`·`codex`만 받아요 → `omp`는 거기로 넘어가지 않아요(`codex-skill`은 늘 `codex`만 요청).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 상한은 셋(알려진 호스트 수) | 이름이 세 개로 막혀 있어 넷은 중복이나 모르는 이름이에요 | 상한 없애기 |
| D-2 | 프로필을 읽을 수 없을 때는 모든 호스트 단계를 그대로 보여 줌 | 기존 시험과 동작을 지키고, 깨진 프로필에서도 호스트 단계가 계획돼요 | 모두 skip |
| D-3 | 호스트 상한은 둘이 아니라 셋 | 리더 note 2098은 C-62와 맞추라며 '두 host 상한'이라 적었지만, 카드 C-56 AC-1은 `[claude, codex, omp]` 통과를, C-62는 OMP를 정식 호스트로 정해요. 둘은 모순이라 카드 기준(셋)을 따랐고 리더에게 알렸어요 | 둘 유지(AC-1과 충돌) |

## Result

- AC-1: met — `profile::tests::three_hosts_with_omp_validate` and `four_hosts_and_unknown_hosts_are_refused` (RED 먼저: `["omp"] was refused`); 3개 통과, 넷·`zcode`·`OMP`(대소문자)·빈 목록 거절, `KNOWN_HOSTS.len()+1`로 풀면 경계 시험이 실패(감사 변이 D 잡힘).
- AC-2: met — `question_bank_hosts`(RED 먼저: `OMP not offered`): 인터뷰 `hosts` 행이 세 호스트 다중 선택이고 현재 호스트가 `(recommended)`; `python3 -I tools/source_manifest.py check` 0.
- AC-3: met — `project_update_plan_cli` 59개 중 `host_steps_follow_the_profile_hosts`(네 조합), `a_profile_with_omp_passes_project_plan`, `a_broken_profile_blocks_its_steps_and_not_the_host_hook_steps`; 게이팅을 항상 켜면 실패(변이 잡힘), `install_steps`가 `None`을 넘기거나 `omp-agents`를 게이트하지 않으면 실패(감사 변이 A·E 잡힘).
- AC-4: met — `--lib` 1420, `update_end_to_end_cli` 8, `go_*`·`project_update_hooks_cli`·`product_cli` 76, `cargo fmt --check` 0, `cargo clippy -D warnings` 0, python 시험 통과, PR #3200 `ci` 성공(run 44232193101). 감사의 비차단 N1(`omp-config`가 고른 호스트에 묶였는지 시험이 못 잡음)·N2(셋 안의 중복이 통과)는 후속으로 남김.
