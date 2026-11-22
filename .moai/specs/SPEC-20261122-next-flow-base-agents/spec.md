---
id: SPEC-20261122-next-flow-base-agents
title: "The five base agents read whole, write only inside their scope, never start an agent, and render for Claude Code only"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-next-flow-skills-agents
---

# The base agents of NeXT-Flow

## Intent

기본 에이전트 다섯(moai-plan, moai-research, moai-impl, moai-audit, moai-docs)의 원본(`agents/*.md`)과 렌더러(`src/agent_defs.rs`)를 NeXT-Flow 와 공식 문서에 맞춰요. 이 조각은 SPEC 만 써요.

바뀌는 것: (1) 생성되는 "먼저 읽기" 줄이 "다른 내부 참조는 읽지 말 것"에서 "이 파일들을 통째로 먼저 읽고, 판단에 쓰는 다른 파일도 기대기 전에 통째로 읽는다; 검색 결과나 발췌로 판단하지 않는다"로 바뀌어요. 이 문장 때문에 조사 에이전트가 내부 참조 18개를 읽지 않은 일이 있었어요(보고서 2-3). (2) `scope` 는 쓸 수 있는 경로만 정하고 읽기는 막지 않아요. (3) 모든 기본 에이전트가 `Agent` 도구를 쓸 수 없어서, "에이전트가 에이전트를 띄우지 않는다"를 글이 아니라 호스트가 강제해요(서브에이전트도 기본으로 3층까지 띄울 수 있어요). (4) 렌더러는 Claude Code 파일만 만들어요(OMP 출력 삭제). (5) 앞머리에 `maxTurns` 와 `color` 를 더하고, moai-audit 에는 사용자 범위 쓰기 거절 훅을 달아요(사용자 범위 에이전트의 훅은 작업 공간 신뢰 없이 돌아요). (6) 감사 에이전트는 자기 규칙을 설치된 정의에서만 받고, 감사하는 작업트리의 지시 파일을 읽지 않으며, 발견 하나를 file·line·summary·failure scenario·category·CONFIRMED/PLAUSIBLE 로 적어요. (7) 빌드는 작업트리 기본 폴더로 가요(훅·규칙·스킬 보고서 물음 ⑥의 권하는 답; 사용자 선택 기록은 찾지 못해 검토에서 확인해요, D-5. `internal/build-folder` 는 부모 SPEC 이 지워요).

### What exists (tip `78c8a5a247`, 잰 시각 2026-11-22T07:2xZ)

- 원본 다섯: `agents/{audit,docs,impl,plan,research}.md`, 합쳐 18,843 바이트. 앞머리 키는 `name`, `description`, `readonly`, `isolation`, `shape`, `reads` 여섯으로 고정돼 있어요(`src/agent_defs.rs:121-130`).
- 생성 줄: `Read first, from the moai skill package: … Read no other internal reference.`(`src/agent_defs.rs:195-208`), 고정 시험은 `src/agent_defs.rs:454-469` 와 `tests/agent_defs_package.rs`.
- Claude 출력: 읽기 전용 역할에만 `disallowedTools: Write, Edit, NotebookEdit`, 쓰는 역할(moai-impl, moai-docs)에는 막는 도구가 없어요(`src/agent_defs.rs:299-318`). 원본 글은 "Do not start other agents" 문장으로만 막아요.
- OMP 출력: `render_omp`(`src/agent_defs.rs:320-369`)와 그 시험이 남아 있고, 홈 설치(`HOSTS = ["claude"]`)는 이미 OMP 를 쓰지 않아요. 프로젝트 층의 `omp_agents` 가 아직 부르므로, OMP 제거 조각이 먼저 착지해야 해요.
- 모델·노력: 시험 `rendered_agent_files_name_no_model_and_no_effort`(`tests/agent_defs_package.rs:244-283`)가 렌더 파일에 `model`·`effort` 가 없음을 고정해요. 근거는 2026-11-16 측정(`docs/agent-model-effort-inheritance.md`): 그런 파일의 자식은 부모의 모델·노력을 받았어요.
- moai-audit 본문 2단계는 PR 의 기준 참조에서 `REVIEW.md` 와 `agents/audit.md` 를 `git show` 로 읽어요(PR 시절 규칙).
- moai-impl·moai-audit 본문은 빌드를 사용자 홈의 moai 빌드 폴더(`.moai/build/`) 아래 자기 폴더로 보내라고 해요(`internal/build-folder`).

## Non-goals

- 에이전트를 설치하는 자리(설정 폴더 바로 아래 `agents/` 인지 그 아래 하위 폴더인지)와 설정 폴더 목록: 전역 설치 조각의 일. 이 SPEC 은 지금의 평평한 `agents/moai-<role>.md` 이름을 바꾸지 않아요.
- 도구 허용 목록(`tools:`)으로 바꾸기: 막는 목록을 유지해요(D-2).
- 읽기 전용 에이전트의 쓰기를 거절하는 설정 훅(H3 의 agent_type 규칙): 훅 SPEC 의 일. 이 SPEC 은 읽기 전용 목록의 원천(`readonly: true`)과 moai-audit 의 앞머리 훅만 정해요.
- `memory`, `skills`, `effort`, `model` 앞머리: 쓰지 않아요(D-1, D-3).

## Requirements

- REQ-1: The generated read-first line of every rendered agent shall tell the agent to read each listed reference whole, and every other file it relies on whole before relying on it, never by search hits or excerpts, with no sentence that forbids reading another reference.
- REQ-2: Every base agent source shall describe its scope input as the paths the agent may write and put no limit on what it reads.
- REQ-3: Every rendered base agent file shall list Agent among its disallowed tools.
- REQ-4: The agent renderer shall render Claude Code agent files only and refuse every other host name.
- REQ-5: Every rendered base agent file shall carry the maxTurns value and the color its source declares and no model or effort key.
- REQ-6: The rendered moai-audit file shall carry a PreToolUse hook entry in the hook writer's exec form that runs the read-only refusal of the tool-preflight duty for the write tools and the shell tools.
- REQ-7: The moai-audit source shall take its rules from its installed definition and read no instruction file from the worktree it audits.
- REQ-8: The moai-audit output shall list each finding with its file, line, summary, failure scenario, category and a CONFIRMED or PLAUSIBLE verdict.
- REQ-9: Every base agent's read-first list shall name only references the package ships, with the audit contract first in the list of moai-audit.
- REQ-10: The moai-impl and moai-audit sources shall send their builds to the worktree's default build folder and delete only build output inside their own worktree by relative path.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9): every rendered file opens with the new read-first line, no source or rendering holds "Read no other internal reference", every source names `scope` as write-only, and every reads entry is a shipped file.
  verify: `cargo test --locked --offline --test agent_defs_package reads_`
- AC-2 (REQ-3, REQ-4, REQ-5): the five renderings list Agent in disallowedTools, carry maxTurns and color, hold no model or effort key, and a render for any host other than claude is refused.
  verify: `cargo test --locked --offline --test agent_defs_package render_`
- AC-3 (REQ-6): the moai-audit rendering holds one PreToolUse entry in exec form with the absolute moai path given at render time, matching the write tools and the shell tools.
  verify: `cargo test --locked --offline --test agent_defs_package audit_hook_`
- AC-4 (REQ-7, REQ-8, REQ-10): the audit source names no file of the audited worktree as a rule source, its output section lists the six finding fields, and the two building sources name the worktree's default build folder and no folder under the user's home.
  verify: `cargo test --locked --offline --test agent_defs_package text_`
- AC-5: the unit tests of the renderer and the neighbouring checks.
  verify: `cargo test --locked --offline --lib agent_defs`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-next-flow-base-agents/**`
- later slices, write: `agents/*.md`, `src/agent_defs.rs`, `src/agent_install.rs` (the render call that passes the moai path), `tests/agent_defs_package.rs`, `skills/moai/references/source-manifest.json`
- read-only: `src/hook_engine.rs`, `src/hooks.rs`, `.moai/specs/SPEC-20261122-next-flow-hooks/spec.md`
- forbidden: the real HOME, the network, the integration branch, `AGENTS.md` and `AGENTS.local.md` of this repository

## Plan

1. (done) SPEC (this slice).
2. Tests first in `tests/agent_defs_package.rs` (prefixes `reads_`, `render_`, `audit_hook_`, `text_`); the older pins that name OMP, `Read no other internal reference` or the frontmatter key list follow the new rule in the same commit, each with its reason.
3. Renderer: the frontmatter keys `maxTurns` and `color` join the source keys; `disallowedTools` always holds `Agent`; `render_omp` and the OMP branch go (after the OMP removal slice lands); `render` takes the moai binary path for the audit hook.
4. Sources: the five bodies get the scope sentence, the build folder sentence, the audit rule source, the finding shape; the reads lists drop the references the parent SPEC removes and moai-audit reads the audit contract first.

## Risks

- The OMP removal slice still calls `render(_, "omp")` from the project layer; this SPEC's renderer step waits for it or the build breaks.
- The audit hook needs the absolute moai path at render time, which makes the rendering depend on where moai runs. The agent install already records the package digest; the path joins the plan so a moved binary shows as an update, never as a silent stale hook.
- A frontmatter key Claude Code does not know is ignored, not refused; a typo in `maxTurns` would be lost silently. The render test pins the exact key names.
- `maxTurns` values are proposals from the whole-docs report (plan 40, research 60, impl 120, audit 80, docs 40); a too-low value stops a long audit. The values stay in the sources so a later change is one line each.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | No `model` and no `effort` in the rendered files; the child keeps the parent's model and effort. | The measured inheritance of 2026-11-16 and the pin that holds it; GLM lanes map model names differently, so a fixed name would break them. The whole-docs report proposed fixed models; this is the open point for the reviewer. | Add both keys per role. |
| D-2 | Disallowed lists stay; no `tools` allow-lists. | An allow-list hides tools a later host adds and the web tools research needs; the default build has no Grep or Glob tool, so an allow-list from the docs example breaks search. | Switch to allow-lists. |
| D-3 | No `memory` key. | Agent memory is a TRIAL item behind the user's pick. | Add on the pick. |
| D-4 | The audit hook runs the same read-only refusal the settings hook runs. | One rule, two places: the settings hook can be missing in a config folder, the frontmatter hook of a user-scope agent runs without workspace trust. | Drop the frontmatter hook. |
| D-5 | Builds go to the worktree's default build folder. | The recommended answer to question 6 of the hooks, rules and skills report (the 2026-11-22 worktree decision against the 2026-11-17 home build folder); I found no record that the user picked it, so the reviewer confirms it before the source slice. `internal/build-folder` leaves the package either way. | Return to the home build folder. |

## Result

(Filled when the work is done. This is the SPEC-only slice.)
