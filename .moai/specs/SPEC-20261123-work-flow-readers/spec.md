---
id: SPEC-20261123-work-flow-readers
title: "work_flow readers: session-check, role entry text and help example follow the profile"
status: draft
revision: 1
created: 2026-11-23
parent: SPEC-20261119-plan-remote-branch
---

# work_flow readers: session-check, role entry text and help example follow the profile

## Intent

카드 C-105 의 남은 기준 3–5 조각이에요. 프로필의 `work_flow`(흐름, 통합 브랜치)를 저장만 하고 보여 주지 않아서, 세션이 시작될 때 리더도 레인도 지금 흐름과 통합 브랜치가 무엇인지 볼 수 없어요(`.moai/project.yaml` 은 NeXT-Flow·`NeXT`). 이 SPEC 은 session-check 가 그 둘을 한 줄로 알리고, 역할 진입 글(리더 배차·레인 시작)이 프로필 값을 따르게 하고, `project-apply` 도움말 예시를 실제 입력 계약에 맞춰요. 기준 1·2 는 이미 병합됐어요(부모 SPEC).

## Non-goals

- 기준 1(spec take 기본 base)·기준 2(project-plan 원격 검사) — 부모 SPEC 과 `SPEC-20261119-spec-take-profile-base` 에서 끝났어요.
- 열린 PR 의 목표 브랜치를 읽는 일(네트워크·`gh`) — session-check 는 로컬 읽기만 하고, NeXT-Flow 에서 PR 은 릴리스에만 있어요. 결정 D-1 참고.
- 워크트리 `baseRef: head` 설정 쓰기(카드 본문 (4)) — 별도 조각이에요.

## Requirements

- REQ-1: When the SessionStart hook runs session-check in a project whose profile names a `work_flow`, the hook shall add one sentence that names the flow and the integration branch read from the profile.
- REQ-2 (REQ-1): IF the profile is missing, unreadable or names no `work_flow`, THEN session-check shall print the text it printed before this change.
- REQ-3: When session-check names the integration branch and the project's local repository has no branch of that name, the hook shall say in the same sentence that the branch is absent locally. · failure: none — a local ref lookup that cannot run prints no branch note (REQ-2 covers the unreadable profile)
- REQ-4: When the leader dispatch line or the lane start line is built for a project whose profile names a `work_flow`, the builder shall name the flow and the integration branch read from the profile. · failure: none — a profile without `work_flow` leaves the line as it was
  [NEEDS CLARIFICATION: which text is the dispatch and start step now (the `src/lane_duty.rs` line builder, `skills/moai/workflows/work.md`, or both); the named `mode-leader.md` and `mode-agent.md` no longer exist]
- REQ-5: When `moai project-apply --help` shows its example, the example shall match the input the command actually reads. · failure: none — a help text change only
  [NEEDS CLARIFICATION: whether `project-apply` reads standard input at all; `src/help.rs:310` shows ` < request.json`, the card says stdin must be the plan result]
- REQ-6 (REQ-5): IF the example and the command's real input disagree, THEN the help-example pin test shall fail.

## Acceptance

- AC-1 (REQ-1): 프로필에 `work_flow`(m-flow, `NeXT`)가 있는 임시 프로젝트의 session-check 답에 흐름과 통합 브랜치가 들어 있어요. · verify: `cargo nextest run --offline --locked -E 'test(session_check_names_the_profile_work_flow)'`
- AC-2 (REQ-2): 프로필 없음·깨진 프로필·`work_flow` 없음에서 답이 이전과 바이트 같아요. · verify: `cargo nextest run --offline --locked -E 'test(session_check_without_work_flow_is_unchanged)'`
- AC-3 (REQ-3): 통합 브랜치가 로컬에 없는 임시 저장소에서 답이 "absent locally" 를 말해요. · verify: `cargo nextest run --offline --locked -E 'test(session_check_flags_a_missing_integration_branch)'`
- AC-4 (REQ-4): 배차·시작 줄이 프로필 값을 이름해요; 값은 프로필에서 읽고 글에 박지 않아요. · verify: 대상이 정해지면 적어요 (REQ-4 의 NEEDS CLARIFICATION)
- AC-5 (REQ-5, REQ-6): 도움말 예시가 실제 입력과 맞고 핀 시험이 통과해요. · verify: `cargo nextest run --offline --locked -E 'test(stdin_examples_match_the_commands_that_read_stdin)'`

## Scope

- write: `src/hook_engine.rs`(session-check), `src/profile.rs`(읽기 도우미가 필요할 때만), `src/help.rs`(예시), REQ-4 의 대상 파일(결정 뒤 확정), `.moai/specs/SPEC-20261123-work-flow-readers/spec.md`
- read-only: `.moai/project.yaml`, `src/card/work.rs`, `src/card/dispatch.rs`(이미 `work_flow` 를 읽는 쪽의 읽기 형태)
- forbidden: 실제 HOME·원장·키, 네트워크, `CLAUDE.md`

## Plan

1. 결정 D-1·D-2 와 SPEC 승인을 기다려요(RUN 은 NEEDS CLARIFICATION 이 열린 동안 시작하지 않아요).
2. RED: 위 시험 이름들로 임시 프로젝트 하네스(고립 HOME)에서 먼저 실패를 봐요.
3. GREEN: session-check 에 한 문장(REQ-1·3), 프로필 읽기는 실패해도 이전 답(REQ-2).
4. REQ-4·5 는 대상이 정해진 뒤 같은 순서로.
5. 이웃: `session_check` 를 고정한 기존 시험, `mode_workflow_steps`·`next_flow_skill_scan`(skills 를 고칠 때), fmt, clippy.

## Risks

- session-check 는 SessionStart 마다 돌고 답이 모델 문맥에 들어가요 — 문장을 늘리면 모든 세션 비용이 늘어요. 한 문장·고정 어휘로 제한해요.
- 프로필 읽기 실패가 훅을 깨면 세션 시작이 막혀요 — REQ-2 로 폴백을 고정해요.
- 지시 파일(skills)을 바꾸면 저장소를 읽는 시험이 깨질 수 있어요 — 바꾸기 전 `mode_workflow_steps`·`next_flow_skill_scan` 을 먼저 돌려 기준선을 잡아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기준 3 의 "열린 PR 이 목표하는 브랜치와 불일치" 를 "통합 브랜치가 로컬에 없음" 으로 바꿔요(제안, 사용자 결정 대기) | session-check 는 네트워크를 쓰지 않고 NeXT-Flow 에서 PR 은 릴리스뿐이라 옛 문구는 지킬 수 없어요 | 카드 기준 3 문구를 되돌리고 `gh` 조회 조각을 따로 열기 |
| D-2 | 카드 쓰기 범위를 현재 파일로 넓혀요(제안, 리더·사용자 결정 대기) | `mode-leader.md`·`mode-agent.md` 는 없고 session-check·도움말은 범위 밖이에요 | 범위 되돌리기 |

## Result

- 아직 구현 전이에요(draft, 승인 대기).
