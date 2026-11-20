---
id: SPEC-20261120-closing-commit-branch
title: "The closing commit of the PROJECT route never lands on the integration branch under m-flow"
status: draft
revision: 2
created: 2026-11-20
---

# The closing commit of the PROJECT route never lands on the integration branch under m-flow

## Intent

mo.ai.kr 에서 `/moai project` 를 돌린 사용자가 알려 준 일이에요(2026-11-19): 마지막 단계 8 이 "지금 가지에 커밋"을 권하는데 그 저장소에서 지금 가지는 직접 커밋이 안 되는 통합 가지(`main`)라서 단계가 건너뛰어졌어요. 모든 MoAI 사용자에게 같은 문제예요: `work_flow` 가 m-flow 이고 지금 가지가 통합 가지이면, 마무리 커밋은 통합 가지에 만들지 않고 새 가지를 만들어 거기에 커밋하거나 커밋하지 않고 두는 둘 중 하나를 골라요.

고르는 규칙(글로 쓰고 시험이 그 글을 못 박아요):

- 통합 가지의 이름은 프로필의 `work_flow.integration_branch` 예요. 적혀 있지 않으면 이 규칙은 적용하지 않고 지금의 커밋 선택(통합 가지 판단 없음)을 그대로 보여요.
- 지금 가지(`git branch --show-current`)가 통합 가지와 같고 `work_flow.flow` 가 `m-flow` 일 때만 바뀌어요. 다른 흐름이거나 다른 가지이면 지금 단계 8 이 그대로예요.
- 선택: "Commit on a new branch (recommended)" 와 "Leave them uncommitted". 새 가지의 이름은 SPEC 가지 규칙(`spec/<slug>-<yyyymmdd>`, 카드 번호 없음)이고 slug 는 이 경로가 하는 일(`project-setup` 같은 짧은 이름)이에요.
- 새 가지는 지금 커밋(HEAD)에서 `git switch -c <이름>` 으로 만들어요(작업 트리의 변경은 그대로 따라와요). 만들기가 실패하면 git 의 메시지를 보이고 커밋하지 않고 두는 쪽으로 가요.
- 통합 가지에는 어떤 경우에도 커밋하지 않아요. 취소·빈 답·무응답은 "Leave them uncommitted" 예요.

이 SPEC 은 규칙의 글과 시험이에요. 단계 8 의 나머지(경로 고르기, 비밀 이름 거르기, 푸시 안 함)는 그대로예요.

## Non-goals

- 푸시·PR 열기는 하지 않아요(단계 8 은 푸시하지 않아요).
- 통합 가지를 알아내는 새 방법을 만들지 않아요: 프로필의 `work_flow.integration_branch` 만 봐요.
- `moai spec new` 로 SPEC 을 만들지 않아요: 새 가지는 이름만 SPEC 가지 규칙을 따르고 SPEC 폴더는 사용자가 이어서 만들어요.
- github-flow·trunk-based·git-flow 의 마무리 커밋은 바꾸지 않아요.

## Requirements

- REQ-1: When the profile's `work_flow.flow` is `m-flow` and the current branch is the profile's `work_flow.integration_branch`, the closing-commit step shall offer "Commit on a new branch (recommended)" and "Leave them uncommitted" and no commit on the current branch.
- REQ-2: The closing-commit step shall name the new branch by the SPEC branch rule `spec/<slug>-<yyyymmdd>` with no card number.
- REQ-3: The closing-commit step shall create the new branch from the current commit with `git switch -c`.
- REQ-4 (REQ-1): IF creating the branch fails, THEN the closing-commit step shall show git's message and commit nothing.
- REQ-5: The closing-commit step shall keep the commit selection it has today when the flow is not m-flow, the current branch is not the integration branch or the profile names no integration branch.
- REQ-6: The closing-commit step shall treat a cancelled, empty or unanswered selection as "Leave them uncommitted".
- REQ-7: The guidance test shall fail when the commands the step names, run in a temporary repository, commit on the integration branch.

## Acceptance

- AC-1 (REQ-1, REQ-5, REQ-6): 단계 8 이 m-flow·통합 가지 조건과 두 선택지, 그 밖의 경우의 지금 선택, 취소·빈 답·무응답 규칙을 말해요.
  verify: `cargo test --locked --offline --test project_closing_branch closing_commit_on_the_integration_branch_offers_a_new_branch`
- AC-2 (REQ-2, REQ-3, REQ-4): 단계 8 이 가지 이름 규칙(`spec/<slug>-<yyyymmdd>`, 카드 번호 없음)과 `git switch -c`, 실패 때의 동작을 말해요.
  verify: `cargo test --locked --offline --test project_closing_branch the_new_branch_is_named_by_the_spec_rule_and_failure_commits_nothing`
- AC-3 (REQ-7): 임시 저장소에서 통합 가지 위에 변경을 두고 단계가 말한 `git switch -c` 와 `git commit -- <path>` 를 돌리면 통합 가지의 HEAD 는 그대로이고 새 가지에만 커밋이 생겨요.
  verify: `cargo test --locked --offline --test project_closing_branch the_commands_leave_the_integration_branch_untouched`

## Scope

- write: `.moai/specs/SPEC-20261120-closing-commit-branch/spec.md`, `skills/moai/workflows/project.md`(단계 8), `tests/project_closing_branch.rs`(새 파일, 시험 셋), `skills/moai/references/source-manifest.json`
- read-only: `skills/moai/references/internal/interview.md`(`work_flow` 행), `skills/moai/references/internal/sdd.md`(가지 이름 규칙), `src/profile.rs`(`work_flow` 설정)
- forbidden: 푸시·PR, 새 도구 코드, 다른 흐름의 마무리 커밋

## Plan

1. SPEC 을 커밋해요(이 문서).
2. `skills/moai/workflows/project.md` 는 다른 변경이 쥐고 있어요(프로젝트·업데이트 경계 재작성): 그 변경이 파일을 돌려준 뒤에 단계 8 에 이 SPEC 의 글을 더해요. 시험은 그 파일을 쓰지 않는 새 파일 `tests/project_closing_branch.rs` 에 두어 먼저 써요. 그 전에는 리더에게 요청하는 노트에 문장과 시험을 정확히 적어 두고, 다른 변경의 단계 8 문장이 바뀌면 그 위에서 더해요.
3. RED: 시험 셋을 먼저 써서 글이 없을 때 실패하는 것을 봐요. `the_commands_leave_the_integration_branch_untouched` 는 임시 저장소(고정된 git 가드, 실제 HOME 아님)에서 `git switch -c` 와 경로 지정 `git commit` 을 돌려 통합 가지의 HEAD 가 그대로임을 확인해요.
4. GREEN: 단계 8 에 규칙 한 문단을 더해요. 스킬 파일이라 `python3 -I tools/source_manifest.py write` 로 매니페스트를 갱신해요.
5. 확인: `project_closing_branch` 와 기존 `project_flow_closing` 전체, `product_cli skill_package`, 스킬 설명을 읽는 시험, `tests.source_manifest_test`·`tests.skill_references_test`·`tests.repo_scans_test`, fmt, clippy, `moai check affected`, 변이(m-flow 조건·"never commit" 문장 지우기).

## Risks

- `project.md` 단계 8 은 이미 긴 한 줄이고 다른 변경이 같은 파일을 다시 쓰고 있어요: 그 변경이 끝난 뒤에 더하고, 같은 줄을 동시에 고치는 병합 충돌을 피해요.
- 통합 가지가 프로필에 없는 프로젝트는 이 규칙을 못 쓰고 지금처럼 동작해요: 그 경우 통합 가지에 커밋하게 될 수 있어요. 프로필에 통합 가지를 적게 하는 일은 `work_flow` 면접 행의 몫이에요.
- 새 가지 이름이 SPEC 가지 규칙을 쓰지만 SPEC 이 없어서 `moai spec new` 가 뒤따라야 해요: 가지 이름 규칙의 훅이 SPEC 없이는 `.moai/specs/` 밖 파일 쓰기를 막으니, 이 단계는 이미 써 둔 파일을 커밋만 하고 새 파일을 더 쓰지 않아요.
- `git switch -c` 는 작업 트리 변경이 새 가지와 충돌할 때 실패할 수 있어요: 같은 커밋에서 만드니 충돌은 없지만 실패하면 REQ-4 가 막아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 통합 가지는 프로필의 `work_flow.integration_branch` 로만 정한다 | 프로필에 이미 있는 값이라 새 판단 도구가 필요 없어요 | 새 판단 규칙 더하기 |
| D-2 | 새 가지는 `git switch -c <이름>` 으로 현재 커밋에서 만든다 | 작업 트리의 변경이 그대로 따라와서 이미 쓴 파일을 옮길 일이 없어요 | `git switch` 로 돌아가기 |
| D-3 | 통합 가지에는 어떤 경우에도 커밋하지 않고 취소·빈 답·무응답은 커밋하지 않고 두는 쪽이다 | 사용자가 알려 준 일: 그 가지는 직접 커밋이 안 되고, 기존 단계 8 의 무응답 규칙과 같아요 | 규칙 문장 바꾸기 |

## Result

설계 단계(2026-11-20): 글과 시험은 `project.md` 를 쥔 다른 변경이 돌려준 뒤에 더해요. 구현이 끝나면 AC 마다 `- AC-n: met — <근거>` 줄을 이 자리에 적어요.
