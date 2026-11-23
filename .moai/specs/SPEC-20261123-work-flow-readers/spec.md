---
id: SPEC-20261123-work-flow-readers
title: "work_flow readers: session-check, role entry text and help example follow the profile"
status: approved
revision: 2
created: 2026-11-23
parent: SPEC-20261119-plan-remote-branch
approval:
  by: user
  at: 2026-11-23
  basis: "사용자 선택 2026-11-23T11:58:16Z 리더 세션 (카드 C-105 노트 7468, '그대로 승인'); 리더가 옮겨 적음. 개정 2 는 REQ-4·5 의 열린 물음을 레인이 리더 결정 7458 ① 안에서 닫은 것(D-3·D-4)이라 리더 확인 대기"
  digest: "sha256:4c268c73617d0818000c32ac9beac596fde04c709c6a0a8017a5cf541ccd0153"
---

# work_flow readers: session-check, role entry text and help example follow the profile

## Intent

카드 C-105 의 남은 기준 3–5 조각이에요. 프로필의 `work_flow`(흐름, 통합 브랜치)를 저장만 하고 보여 주지 않아서, 세션이 시작될 때 리더도 레인도 지금 흐름과 통합 브랜치가 무엇인지 볼 수 없어요(`.moai/project.yaml` 은 NeXT-Flow·`NeXT`). 이 SPEC 은 session-check 가 그 둘을 한 줄로 알리고, 역할 진입 글(리더 배차·레인 시작)이 프로필 값을 따르게 하고, `project-apply` 도움말 예시를 실제 입력 계약에 맞춰요. 기준 1·2 는 이미 병합됐어요(부모 SPEC).

## Interview

- Q1: 기준 3 의 "열린 PR 이 목표하는 브랜치와 불일치" 를 어떻게 하나요? → A: "통합 브랜치가 로컬에 없으면 알려요"(사용자 선택 2026-11-23T11:58:16Z, 카드 C-105 노트 7468 '그대로 승인', D-1).
- Q2: 역할 진입 글(REQ-4)은 어느 파일인가요? → A: `skills/moai/workflows/work.md` 의 take 단계예요(레인 결정 D-3, 리더 결정 7458 ① 안; 사용자 질문이 아니에요).
- Q3: `project-apply` 는 표준입력을 읽나요? → A: 읽어요. `src/ops/project_apply.rs` 가 `plan_input::<Plan>()` 으로 계획 결과를 읽어요(코드를 읽어 확인, D-4).

## Non-goals

- 기준 1(spec take 기본 base)·기준 2(project-plan 원격 검사) — 부모 SPEC 과 `SPEC-20261119-spec-take-profile-base` 에서 끝났어요.
- 열린 PR 의 목표 브랜치를 읽는 일(네트워크·`gh`) — session-check 는 로컬 읽기만 하고, NeXT-Flow 에서 PR 은 릴리스에만 있어요. 결정 D-1 참고.
- 워크트리 `baseRef: head` 설정 쓰기(카드 본문 (4)) — 별도 조각이에요.

## Requirements

- REQ-1: When the SessionStart hook runs session-check in a project whose profile names a `work_flow`, the hook shall add one sentence that names the flow and the integration branch read from the profile.
- REQ-2 (REQ-1): IF the profile is missing, unreadable or names no `work_flow`, THEN session-check shall print the text it printed before this change.
- REQ-3: When session-check names the integration branch and the project's local repository has no branch of that name, the hook shall say in the same sentence that the branch is absent locally. · failure: none — a local ref lookup that cannot run prints no branch note (REQ-2 covers the unreadable profile)
- REQ-4: The WORK route's take step shall tell the session to read the flow and the integration branch from the profile with `moai work-flow-plan --project .` and to name both in the first note it writes on the card, with no value typed into the text.
- REQ-5: The `project-apply` help examples shall end with `< plan.json`, the plan result `project-plan` prints, because the command reads a plan and not a request.
- REQ-6 (REQ-5): IF a `project-apply` help example names another input file than the plan, THEN the help pin test shall fail.

## Acceptance

- AC-1 (REQ-1): 프로필에 `work_flow`(m-flow, `NeXT`)가 있는 임시 프로젝트의 session-check 답에 흐름과 통합 브랜치가 들어 있어요. · verify: `cargo nextest run --offline --locked -E 'test(session_check_names_the_profile_work_flow)'`
- AC-2 (REQ-2): 프로필 없음·깨진 프로필·`work_flow` 없음에서 답이 이전과 바이트 같아요. · verify: `cargo nextest run --offline --locked -E 'test(session_check_without_work_flow_is_unchanged)'`
- AC-3 (REQ-3): 통합 브랜치가 로컬에 없는 임시 저장소에서 답이 "absent locally" 를 말해요. · verify: `cargo nextest run --offline --locked -E 'test(session_check_flags_a_missing_integration_branch)'`
- AC-4 (REQ-4): 작업 경로의 take 단계가 프로필에서 흐름·통합 브랜치를 읽으라고 말하고 값은 글에 박지 않아요. · verify: `cargo nextest run --offline --locked -E 'test(work_take_reads_the_profile_flow_and_integration_branch)'`
- AC-5 (REQ-5, REQ-6): `project-apply` 도움말 예시가 계획 파일을 읽고, 다른 파일을 이름하면 핀 시험이 실패해요. · verify: `cargo nextest run --offline --locked -E 'test(project_apply_examples_read_the_plan_file)'`

## Scope

- write: `src/hook_engine.rs`(session-check), `src/help.rs`(예시와 핀 시험), `skills/moai/workflows/work.md`(REQ-4, D-3), `tests/mode_workflow_steps.rs`(REQ-4 핀 시험), `.moai/specs/SPEC-20261123-work-flow-readers/spec.md`
- read-only: `.moai/project.yaml`, `src/card/work.rs`, `src/card/dispatch.rs`(이미 `work_flow` 를 읽는 쪽의 읽기 형태)
- forbidden: 실제 HOME·원장·키, 네트워크, `CLAUDE.md`

## Plan

1. 결정 D-1·D-2 와 SPEC 승인은 받았고, REQ-4·5 의 열린 물음은 D-3·D-4 로 닫았어요.
2. RED: 위 시험 이름들로 임시 프로젝트 하네스(고립 HOME)에서 먼저 실패를 봐요.
3. GREEN: session-check 에 한 문장(REQ-1·3), 프로필 읽기는 실패해도 이전 답(REQ-2).
4. REQ-4(work.md 한 문장과 그 핀 시험)·REQ-5(도움말 예시와 핀 시험)도 같은 순서로.
5. 이웃: `session_check` 를 고정한 기존 시험, `mode_workflow_steps`·`next_flow_skill_scan`(skills 를 고칠 때), fmt, clippy.

## Risks

- session-check 는 SessionStart 마다 돌고 답이 모델 문맥에 들어가요 — 문장을 늘리면 모든 세션 비용이 늘어요. 한 문장·고정 어휘로 제한해요.
- 프로필 읽기 실패가 훅을 깨면 세션 시작이 막혀요 — REQ-2 로 폴백을 고정해요.
- 지시 파일(skills)을 바꾸면 저장소를 읽는 시험이 깨질 수 있어요 — 바꾸기 전 `mode_workflow_steps`·`next_flow_skill_scan` 을 먼저 돌려 기준선을 잡아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기준 3 의 "열린 PR 이 목표하는 브랜치와 불일치" 를 "통합 브랜치가 로컬에 없음" 으로 바꿔요(사용자 결정 2026-11-23, 카드 노트 '사용자 선택') | session-check 는 네트워크를 쓰지 않고 NeXT-Flow 에서 PR 은 릴리스뿐이라 옛 문구는 지킬 수 없어요 | 카드 기준 3 문구를 되돌리고 `gh` 조회 조각을 따로 열기 |
| D-2 | 카드 쓰기 범위를 현재 파일로 넓혀요(리더 결정 2026-11-23 ①, 7448) | `mode-leader.md`·`mode-agent.md` 는 없고 session-check·도움말은 범위 밖이었어요 | 범위 되돌리기 |
| D-3 | REQ-4 의 대상은 `skills/moai/workflows/work.md` 의 take 단계예요(레인 결정, 리더 결정 ① 안) | `src/lane_duty.rs` 의 줄은 옛 Issue·PR 어휘의 골든 줄이고 프로젝트 경로 없이 `leader_round` 에서도 불리며, 값을 글에 박지 않으려면 take 단계가 `moai work-flow-plan` 으로 읽게 하는 편이 한 곳이에요. 핀 시험은 `mode_workflow_steps` 에 둬서 범위에 그 파일 하나를 보태요 | `work.md` 문장을 지우고 `lane_duty::line` 에 프로젝트 인자를 더하는 조각을 따로 열기 |
| D-4 | REQ-5 의 사실은 `project-apply` 가 `plan_input::<Plan>` 으로 표준입력의 계획 결과를 읽는다는 것이고, 핀 시험은 `help.rs` 안에 둬요(레인 결정) | `src/main.rs` 의 `stdin_examples_match…` 는 파일 이름을 보지 않고 `main.rs` 는 범위 밖이에요. 같은 어긋남이 다른 `*-apply` 예시에도 있지만 이 조각은 `project-apply` 만 고쳐요 | 도움말 예시를 `request.json` 으로 되돌리기 |

## Result

- 아직 구현 전이에요(draft, 승인 대기).
