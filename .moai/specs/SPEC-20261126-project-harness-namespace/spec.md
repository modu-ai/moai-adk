---
id: SPEC-20261126-project-harness-namespace
title: "project generation puts its skills in moai-prefixed folders beside the moai agents folder, and moves earlier skill output into them through a preview"
status: draft
revision: 1
created: 2026-11-26
---

# project generation puts its skills in moai-prefixed folders beside the moai agents folder, and moves earlier skill output into them through a preview

## Intent

`/moai project` writes the agents and skills it generates next to the files a user makes by hand. 사용자는 사용자가 직접 만든 에이전트·스킬과 MoAI 가 설치한 것을 폴더만 보고 가려서 따로 관리할 수 있어야 한다고 했어요: 에이전트는 `.claude/agents/moai/`, 스킬은 `.claude/skills/moai-*/`.

지금 tip 에서 에이전트 쪽은 이미 끝났어요. 생성된 Claude Code 에이전트는 `.claude/agents/moai/<role>.md` 에 쓰이고, 옛 평면 경로 `.claude/agents/<role>.md` 의 기록된 파일은 계획에서 옮겨져요(`SPEC-20261121-base-agents-user-folder`). 남은 것은 스킬이에요: 프로젝트 스킬과 스킬 투영이 아직 `.claude/skills/<id>/SKILL.md` 에 쓰이고, 요청에 계속 들어 있는 스킬의 옛 경로는 계획에서 옮겨지지도 지워지지도 않아요(기록된 종류가 여전히 요청돼서 정리 단계가 건너뛰어요). Codex 와 OMP 는 철회돼서(`SPEC-20261121-remove-codex-support`) 이 SPEC 은 Claude Code 만 다뤄요.

이 SPEC 은 생성되는 스킬을 `.claude/skills/moai-<id>/SKILL.md` 로 옮기고(공식 문서가 스킬을 한 단계 깊이로만 문서화해서 묶음 폴더가 아니라 폴더 이름 접두사 `moai-` 예요, `research.md`), 이전 실행이 쓴 스킬 파일을 기존 plan·apply 미리보기로 옮겨요. 생성 기록에는 MoAI 가 쓴 파일과 그 바이트의 다이제스트가 있고, 그 기록이 "우리 파일"의 증거예요. 기록에 없는 파일은 사용자의 것이라 절대 건드리지 않아요.

## Non-goals

- 에이전트 경로: 이미 tip 에 있어요. 이 SPEC 의 시험이 새 에이전트 경로도 함께 고정해요.
- Codex·OMP 경로: 철회됐어요.
- 사용자 계층 에이전트와 스킬(`~/.claude/agents`, `~/.claude/skills`)과 프로젝트의 `moai` 스킬 복사본: 다른 계층이라 그대로예요.
- 같은 `name` 을 선언하지만 파일 이름이 다른 사용자 파일의 탐지: 하지 않아요(경로로 찾을 수 없어요).

## Requirements

- REQ-1: The `generation-plan` operation shall plan each project skill of a request as `.claude/skills/moai-<id>/SKILL.md` with the front matter `name` equal to the folder name. · failure: none — a request that cannot be planned is refused by the request check of REQ-3
- REQ-2: The `generation-plan` operation shall plan each skill projection of a request as `.claude/skills/moai-<id>/SKILL.md` with the front matter `name` equal to the folder name. · failure: none — a request that cannot be planned is refused by the request check of REQ-3
- REQ-3 (REQ-1): IF a skill id would make the folder name longer than sixty-four characters, THEN the `generation-plan` operation shall refuse the request with the name of the field and the limit and never the value.
- REQ-4: When a file that the generation record owns lies at the former path `.claude/skills/<id>/SKILL.md` of a requested skill and its bytes equal the recorded ones, the `generation-plan` operation shall plan the removal of the former path together with the creation of the new path. · failure: none — an edited former file is REQ-5
- REQ-5 (REQ-4): IF a file that the generation record owns at a former skill path was edited after it was generated, THEN the `generation-plan` operation shall plan the former path and the new path as conflict and leave the edited file where it is.
- REQ-6: The `generation-plan` operation shall plan no change to a skill folder that the generation record does not own.
- REQ-7: The `generation-apply` operation shall refuse to remove a former skill path whose bytes differ from the plan.
- REQ-8: The project route shall tell the agent that the agents and skills it generates are written inside the `moai` folder of the agents folder and the `moai-` skill folders, apart from the files the user made.
- REQ-9: The SPEC folder shall hold a research document with the official quotes, their sources and read times, which the paths rest on.

## Acceptance

- AC-1 (REQ-1, REQ-2): a plan for one agent, one project skill and one skill projection names exactly `.claude/agents/moai/<role>.md`, `.claude/skills/moai-<id>/SKILL.md` for the skill and for the projection, writes the front matter `name` equal to the role id and to the folder name, plans nothing at the former skill paths, and a second plan of the same request is `unchanged` · verify: `cargo test --locked --offline --test project_harness_namespace fresh_`
- AC-2 (REQ-3): an id that makes the folder name sixty-four characters is planned and one that makes it sixty-five is refused with the field and the limit · verify: `cargo test --locked --offline --test project_harness_namespace folder_name_`
- AC-3 (REQ-4, REQ-5, REQ-6, REQ-7): in temporary projects, an earlier run's owned and unedited skill moves through plan and apply with its body kept (the front matter name follows the folder) and the next plan is `unchanged`; an edited one stays, both paths are conflict and apply refuses; a skill folder of the user's at a former path is never changed and never planned; a former path edited after the plan is not removed · verify: `cargo test --locked --offline --test project_harness_namespace moved_`
- AC-4 (REQ-9): `research.md` of this folder holds the official quotes and at least four read times, and the test reads them · verify: `cargo test --locked --offline --test project_harness_namespace research_`
- AC-5 (REQ-8): step 6 of `skills/moai/workflows/project.md` names the namespace folders, and the readers of that text still pass · verify: `cargo test --locked --offline --test project_harness_namespace text_`

## Scope

- write: `.moai/specs/SPEC-20261126-project-harness-namespace/`, `src/generation.rs`, `tests/project_harness_namespace.rs`(새)
- write (the tests that pin the former skill paths, listed in the Result when found): `tests/project_skills_generation.rs` and the readers of `generation-plan` that name a generated skill path
- write: `skills/moai/workflows/project.md`(step 6 의 두 문장만), `skills/moai/references/source-manifest.json`(한 줄 자리 예약으로만)
- read-only: `skills/moai/references/internal/sdd.md`
- forbidden: Codex·OMP 경로, 카드 번호를 SPEC id·가지·커밋에 쓰는 것, 사용자 계층 폴더 쓰기

## Plan

1. SPEC 과 `research.md`(공식 문서 인용과 읽은 시각), 승인 요청(옮김 계획이 파일을 지우므로 tier 2).
2. 시험 먼저: `tests/project_harness_namespace.rs`(빨강: 스킬 경로가 옛 것), 그다음 `src/generation.rs`(경로 함수, 이름, 길이 한도, 옮김 계획, 기록 경로 검사가 옛 경로와 새 경로를 모두 받기).
3. 옛 스킬 경로를 고정한 시험을 따라가게 고치고, 단계 6 글, 소스 목록, 영역 전체 실행, 가드 변이, READY.

## Risks

- 옮김 계획은 파일을 지워요(tier 2). 지우는 쪽은 기록이 증거인 파일만이고, 바이트가 기록과 같을 때만이에요. 편집된 파일은 남고 충돌이에요. 지우기 경로의 시험은 임시 프로젝트에서 실제 계획·적용을 돌려요.
- 스킬 `name` 이 폴더 이름이 되면 사용자의 같은 id 스킬과 이름이 겹치지 않아요. 에이전트는 그대로 `name: prj-<name>` 이에요(`prj-` 접두사가 이미 사용자 계층 `moai-*` 에이전트와 가려요).
- 폴더 이름 `moai-` 가 붙으면 id 는 최대 59자예요. 플랫폼 문서의 스킬 이름 한도(64자)에 맞춰 요청 검사가 막아요.
- `skills/moai/workflows/project.md` 는 다른 카드가 쥐었던 파일이에요. 쓰기 전에 `moai card check` 가 허용하는지 읽고, 두 문장만 바꿔요.
- 옛 경로를 고정한 시험이 많아요. READY 영수증(`check affected --card`)이 선택하는 시험을 모두 돌려 읽어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 스킬 이름공간은 묶음 폴더가 아니라 폴더 이름 접두사 `moai-` 예요 | 공식 스킬 문서가 `.claude/skills/<skill-name>/SKILL.md` 한 단계만 문서화하고 묶음 폴더에는 침묵해요(`research.md`) | 접두사 함수 하나 |
| D-2 | 스킬 front matter `name` 은 폴더 이름이에요 | 사용자의 같은 id 스킬과 이름이 겹치지 않게 | `name` 만 id 로 |
| D-3 | 옮김은 에이전트 평면 경로 옮김과 같은 모양이에요: 기록된 파일이 기준과 같으면 옛 경로 삭제와 새 경로 생성, 편집된 파일은 충돌 | tip 의 기존 에이전트 옮김 코드(`generation.rs` 평면 역할 블록) | 새 블록 제거 |
| D-4 | 기록 경로 검사(`kind_names_path`)는 한 종류의 옛 경로와 새 경로를 모두 받아요 | 에이전트의 `named_paths` 가 이미 두 경로를 받아요 | 목록 한 줄 |
| D-5 | 요청에 아직 있는 종류의 기록된 옛 경로 항목은 옮김 블록이 맡고, 요청에 없는 종류는 기존 "더 이상 요청되지 않음" 블록이 맡아요 | 두 블록이 같은 파일을 두 번 계획하지 않게 | 블록 합치기 |
| D-6 | Codex·OMP 는 다루지 않아요 | `SPEC-20261121-remove-codex-support` | — |

## Result

(아직 구현 전이에요.)
