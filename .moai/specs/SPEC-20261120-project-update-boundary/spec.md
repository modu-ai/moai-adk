---
id: SPEC-20261120-project-update-boundary
title: "/moai project describes the project only; /moai update owns the earlier-generation moves, the installs and the merges"
status: draft
revision: 1
created: 2026-11-20
---

# /moai project describes the project only; /moai update owns the earlier-generation moves, the installs and the merges

## Intent

사용자 결정(2026-11-19, 리더 대화 (3)·(4)·(5)·(8)·(9)·(10)·(11))의 스킬 글 쪽이에요. 보고서 `update-project-logic-2026-11-20.md` 가 읽은 사실이에요: `workflows/project.md` 가 `/moai project` 한 번에 이전 세대 상태 옮기기(단계 2: `go-settings-apply`, `go-state-apply`, 되돌리기, 백업 ignore 선택)와 worktree 옮기기와 `claude-settings` 의 `baseRef` 적용까지 맡고 있어서, 새 프로젝트도 "옮기기" 문장을 읽고, 같은 절차가 `workflows/update.md` 와 두 군데에 있어요. 사용자는 이렇게 가르기로 했어요. `/moai project` 는 프로젝트를 설명하는 일만 해요(감지·확인, 인터뷰, 프로필, AGENTS.md, 프로젝트 agent, 코드 지도, 마무리 커밋). 이전 세대 이동과 설치·병합은 `/moai update` 가 맡아요. project 는 끝에서 남은 갱신이 있으면 한 줄로 알려요. 질문은 한 번에 3개, 최대 4번이고 아는 값은 묻지 않고 확인해요.
이 SPEC 은 두 워크플로 파일의 글과 그 경계를 고정하는 시험을 다뤄요. 질문 은행 글은 `SPEC-20261120-question-bank-generic` 이 맡아요. 새 Rust 명령은 없어요. 기반은 이 SPEC 이 시작될 때의 로컬 통합 끝이고, `project.md` 를 쥔 언어 세 질문 조각과 `update.md` 를 쥔 선택지 조각이 합쳐진 뒤의 글 위에서 써요.

## Non-goals

- 질문 은행(`interview.md`·`questions.md`·`versioning.md`) 바꾸기: 질문 은행 SPEC 의 일이에요.
- 프로젝트 harness 제안 방식(웹 검색으로 스킬·agent 를 찾아 제안하기): 별도 카드의 일이에요. 이 SPEC 은 단계 6 에서 "최대 3개" 제한만 걷고 제안 방식은 한 줄 안내로 비워요.
- `src/**` 와 명령 동작 바꾸기: 글만 바꿔요. `moai project-update-plan` 의 단계 이름과 출력은 그대로예요.
- work_flow 선택지 gitlab-flow: 프로필 코드와 같이 바뀌어야 해서 별도 조각이에요(질문 은행 SPEC 결정 D-4).
- OMP 지우기: OMP 를 지우는 카드가 같은 시험을 바꿔요. 이 SPEC 이 다시 쓰는 문단에서만 OMP 글을 빼요.

## Requirements

- REQ-1: The `project.md` workflow shall hold no step that moves earlier-generation state, installs managed files or merges backed-up files. · failure: none — a test reads the file for the step names and the apply commands of those steps.
- REQ-2: The `update.md` workflow shall own the earlier-generation move steps, including the go-settings and go-state previews, applies and revert, the worktree move choice and the Claude Code worktree base setting. · failure: none — a test reads the file for each step's commands.
- REQ-3: Each of `project.md` and `update.md` shall state the same boundary sentence. · failure: none — a test requires the exact sentence in both files.
- REQ-4: The `project.md` workflow shall end with one line telling the user to run `/moai update` when `moai project-update-plan` reports steps that do not read `unchanged`. · failure: none — a test requires the line's text and its place after the closing commit step.
- REQ-5: The `project.md` workflow shall ask at most three questions per call and at most four calls, and confirm every value that reading already knows instead of asking it. · failure: none — a test requires both limits and the confirm rule.
- REQ-6: The `project.md` workflow shall offer no decision model and ask no question about the project's purpose. · failure: none — a test requires the absence of the decision-model step and of the purpose ask.
- REQ-10: The `project.md` workflow shall ask the writing language as the three items the question bank names. · failure: none — a test requires the three-item sentence.
- REQ-7: The `project.md` workflow shall not limit the number of project agents or skills it proposes. · failure: none — a test requires the absence of the old limit and a one-line pointer for the way proposals are made.
- REQ-8: The two workflow files shall name no card number, no Issue number, no dated user-decision citation and no mention of OMP in the text this SPEC rewrites. · failure: none — a scan test reads both files for those shapes.
- REQ-9: Every existing test that pins moved text shall follow the text to its new file with the same wording. · failure: none — the pinning tests run and pass, and the boundary test names each moved pin.

## Acceptance

- AC-1 (REQ-1, REQ-2): `project.md` 에 `go-settings-apply`, `go-state-apply`, `go-state-revert`, `worktree-move-apply`, `claude-settings-apply` 명령 글이 없고, `update.md` 에 각각 있어요.
  verify: `cargo test --locked --offline --test project_update_boundary project_holds_no_move_install_or_merge_step`
- AC-2 (REQ-3, REQ-4): 두 파일에 같은 경계 문장이 있고 `project.md` 끝에 "남은 갱신" 한 줄이 마무리 커밋 단계 뒤에 있어요.
  verify: `cargo test --locked --offline --test project_update_boundary the_boundary_sentence_is_in_both_files_and_project_ends_with_one_line`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-10): 질문 수 한도 문장, 아는 값을 확인한다는 문장, 언어 세 항목 문장이 있고 결정 모델 단계와 purpose 질문과 "최대 3개" 제한은 없어요.
  verify: `cargo test --locked --offline --test project_update_boundary project_asks_little_and_offers_no_decision_model`
- AC-4 (REQ-8): 두 파일에 카드 번호, Issue 번호, 날짜 붙은 결정 인용, OMP 가 없어요.
  verify: `cargo test --locked --offline --test project_update_boundary the_workflows_name_no_card_issue_or_omp`
- AC-5 (REQ-9): 글을 고정하던 시험이 옮겨 간 글에서 통과해요.
  verify: `cargo test --locked --offline --test project_flow_closing --test project_preview_order --test project_worktree_move_choice --test update_route --test project_applied_table --test mode_workflow_steps --test code_atlas_reference --test agent_defs_package`; `python3 -m unittest tests.source_manifest_test tests.skill_references_test`

## Scope

- write: `.moai/specs/SPEC-20261120-project-update-boundary/spec.md`, `tests/project_update_boundary.rs`, `skills/moai/workflows/project.md`, `skills/moai/workflows/update.md`, `tests/project_flow_closing.rs`, `tests/project_preview_order.rs`, `tests/project_worktree_move_choice.rs`, `tests/update_route.rs`, `tests/project_applied_table.rs`, `tests/mode_workflow_steps.rs`, `tests/code_atlas_reference.rs`, `tests/agent_defs_package.rs`, `skills/moai/references/source-manifest.json`
- read-only: `src/ops/project_update_plan.rs`, `skills/moai/references/internal/interview.md`, `skills/moai/references/internal/questions.md`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기

## Plan

1. 기반: 로컬 통합 끝에 `project.md` 를 쥔 조각(언어 세 질문)과 `update.md` 를 쥔 조각(선택지)이 들어온 뒤의 끝에서 가지를 새로 잘라요. 기준 3 의 `project.md` 글(미리보기 순서·worktree 옮기기)이 들어온 뒤여야 해요.
2. 시험 먼저: `tests/project_update_boundary.rs` 시험 넷을 쓰고 실패(RED)를 읽어요.
3. `update.md` 에 이전 세대 이동 절차(단계 2 의 미리보기·적용·되돌리기·백업 ignore 선택·root 파일 질문), worktree 옮기기 선택, `baseRef` 적용을 옮겨요. 그 글에 이미 있는 같은 절차는 합쳐요.
4. `project.md` 에서 그 단계와 문단을 지우고, 경계 문장과 끝의 한 줄, 질문 수 한도, 확인 규칙, 언어 세 항목, 단계 6 의 한 줄 안내를 써요. 결정 모델 단계·purpose 질문·인용·OMP 글을 지워요. 닫는 커밋 단계에서 옮긴 적용 이름을 빼요.
5. 글을 고정하던 시험을 새 글로 따라가게 고치고(AC-5 목록 전부), 이웃 검사와 READY 전 검사를 돌려요. 매니페스트 해시는 마지막 커밋에서만 써요.

## Risks

- `project.md`·`update.md` 가 다른 카드의 예약 중이에요 → 그 조각들이 통합된 뒤 그 끝에서 시작해요. 먼저 시작하면 같은 문단을 두 번 고쳐 충돌해요.
- 기준 3 이 `project.md` 에 쓴 worktree 옮기기 글과 그 시험(`project_worktree_move_choice`, `project_preview_order`)이 옮겨 가야 해요 → 시험의 읽는 파일을 `update.md` 로 바꾸되 문구는 그대로 고정해 같은 뜻을 지켜요.
- 옮기면서 문장이 달라지면 감사가 읽은 동작이 바뀔 수 있어요 → 옮긴 문장은 글자 그대로 두고, 바꾼 문장은 결정 표에 적어요.
- 글만 바꾸는 SPEC 이라 명령의 단계 이름이 맞는지를 시험이 못 봐요 → `moai project-update-plan` 이 내는 단계 이름을 시험이 읽어 두 파일의 이름과 맞춰 봐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 경계는 한 문장으로 두 파일에 같게 적고 시험이 둘 다 읽어요 | 사용자 결정 (3)·(4) | 문장과 시험을 지워요 |
| D-2 | project 는 끝에서 남은 갱신이 있으면 한 줄만 알려요; 옮기기·설치·병합을 묻지도 적용하지도 않아요 | 사용자 결정 (3) | 한 줄을 지워요 |
| D-3 | 질문은 한 번에 3개, 최대 4번이고 아는 값은 확인해요 | 사용자 결정 (5); 호스트 질문 도구의 한 번 3개 한도와 맞아요 | 한도 문장을 바꿔요 |
| D-4 | 프로젝트 agent·skill 의 "최대 3개" 제한을 걷고 제안 방식은 한 줄 안내로 비워요 | 사용자 결정 (11); 제안 방식은 별도 카드 | 한 줄을 지우고 제한을 되돌려요 |
| D-5 | 날짜 붙은 인용·카드·Issue 번호·OMP 글은 이 SPEC 이 다시 쓰는 두 파일에서 지워요 | 사용자 결정 (8); 배포 글은 어떤 사용자의 프로젝트에도 맞아야 해요 | 인용을 되돌려요 |

## Result

(채워요: RED, GREEN, 변이, 이웃, Not run)
