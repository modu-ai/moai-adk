---
id: SPEC-20261122-claude-md-layout
title: "CLAUDE.md and CLAUDE.local.md replace AGENTS.md and AGENTS.local.md in the files moai owns, with one flow rule file"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# CLAUDE.md layout: the files moai owns in a project

## Intent

moai 가 사용자 프로젝트에서 관리하는 지침 파일을 Claude Code 의 공식 파일로 바꿔요. 지금은 `AGENTS.md`(관리 블록)와 `AGENTS.local.md`(개인)를 쓰는데, Claude Code 는 `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md` 중 하나라도 있으면 `AGENTS.md` 를 읽지 않아요(공식 문서 memory.md 359-385행, 2026-11-21 에 읽은 사본). 그래서 지금은 `AGENTS.md` 를 쓰고 사용자의 `CLAUDE.md` 에 `@AGENTS.md` 한 줄을 붙이는 우회를 해요. 새 배치는 이래요.

| 자리 | 주인 | git | 내용 | 다루는 SPEC |
| --- | --- | --- | --- | --- |
| `<project>/CLAUDE.md` | 프로젝트, moai 는 블록 하나 | 추적 | 사용자 글 + `moai-workflow` 블록(사실만, 15줄 이내) | 이 SPEC |
| `<project>/CLAUDE.local.md` | 사용자 | 무시 | 개인 메모. moai 는 사용자가 고를 때만 만들어요 | 이 SPEC |
| `<project>/.claude/rules/moai/flow.md` | moai | 추적 | `work_flow` 가 고른 흐름 글 하나. `paths:` 없음 | 이 SPEC |
| 프로젝트의 AGENTS 파일 옮기기 | 사용자 선택 | | `move` / `import` / `leave` | `SPEC-20261122-claude-md-project-picks` |
| 프로젝트 레이어의 쓰기 규칙 | moai | | 영수증, 링크·UTF-8 거부, baseline, 미리보기 키, 토큰 | `SPEC-20261122-claude-md-project-layer` |
| `<config dir>/rules/moai.md` | moai | 해당 없음 | `/moai` 진입 안내 12줄 이내. `<config dir>` 는 `CLAUDE_CONFIG_DIR` 또는 `~/.claude` | `SPEC-20261122-claude-md-home-routing` |
| `.gitignore` 와 글 시험 | | | `.claude/` 통째 무시 해제, 이름 시험, 배너 고정 | `SPEC-20261122-claude-md-text-and-ignore` |

이 SPEC 은 첫 조각(조사 + SPEC)이에요. 코드, 시험, 이 저장소 자신의 `AGENTS.md` 이동은 아래 Plan 의 뒤 조각에서 해요. 요구가 12개를 넘으면 쪼개라는 규칙(SPEC 크기 예산)에 따라 다섯 SPEC 으로 나눴고, 넷은 이 SPEC 을 `parent` 로 이어요. `import` 를 고르면 `AGENTS.local.md` 는 아무 일도 없어요(Claude Code 가 원래 읽지 않던 파일이에요).

### What exists (tip `012e9470cb`)

읽기 전용 조사 에이전트가 코드와 글을 통째로 읽었고, 작성자가 `.gitignore`, `AGENTS.md`, `AGENTS.local.md`, 무시 판정을 직접 확인했어요. `a:` 는 `src/agents.rs` 예요.

- 이름이 한 곳에 모여 있지 않아요: `AGENTS.md` 는 `a:1321,1324,1329,1783`, `AGENTS.local.md` 는 `a:1416,1419,1809`, `.gitignore` 는 `a:1440,1451,1800` 에 줄마다 적혀 있고, 구간 id 만 `a:18-21` 상수예요. 블록 15줄 상한을 검사하는 곳은 없어요(가장 짧은 블록이 14줄이고 `writing` 줄과 버전 줄이 더해져요).
- 블록 글에 "Issue scope", "Issues" 같은 Issue 시대 낱말이 남아 있어요(`a:346,367,310-313`).
- `claude_shadow` 와 선택 `claude-import` 는 "사용자의 `CLAUDE.md` 가 `AGENTS.md` 를 가린다"를 전제해요(`a:167-187`, `a:1480-1501`; `SPEC-20261119-agents-conflict-choices` REQ-11…REQ-14).
- `AGENTS.local.md` 는 커밋되는 `AGENTS.md` 로 합쳐져요(`src/agents_local_merge.rs:1-3`). 개인·무시 파일인 `CLAUDE.local.md` 와 방향이 반대예요. 충돌 한 행이 적용 전체를 막아요(`a:1674`).
- `moai update` 에는 적용 레이어 아홉 개와 9칸 `REVERT_ORDER`(`src/update.rs:1737-1747`), 고정 미리보기 키 13개(`update.rs:1978-1996`), 조건부 키 `keys`·`cleanup`·`home`·`switch`·`git_hooks`(`update.rs:1997-2016`)가 있어요. 프로젝트 폴더에 쓰는 레이어는 `hooks`(다이제스트 안)와 `git_hooks`(자기 토큰, 다이제스트 밖, 되돌리기 맨 앞)예요.
- `project-update-plan` 은 이미 선택마다 명령 하나와 `plan_digest` 하나를 찍는 `choices` 를 가져요(`src/project_update.rs:771-816`).
- 첫 실행은 `AGENTS.md` 를 쓰고 `CLAUDE.md` 는 쓰지 않아요(`src/launcher_init.rs:407-429`).
- `.gitignore` 는 `.claude/` 를 356행(`/.claude/`)과 386행(`.claude/`)에서 두 번 무시해요. `git check-ignore -v` 가 `.claude/settings.json`, `.claude/rules/x.md`, `.claude/worktrees/a` 에 386행을, `CLAUDE.local.md` 에 174행을 대요. 348-355행 주석은 `.claude/` 를 "로컬 설치본"이라고 불러요.
- 이 체크아웃의 `.claude/` 에는 `agents`, `commands`, `hooks`, `output-styles`, `rules`(비어 있음), `settings.json`, `settings.local.json`, `skills`, `workflows`, `worktrees` 가 있고 추적되는 건 없어요(Go 세대 설치 상태).
- `AGENTS.md` 를 moai 의 파일로 적은 출하 글: `skills/moai/SKILL.md:9`, `workflows/project.md:3,9,37,39,49`, `setup.md:16`, `update.md:23,24`, `references/internal/{interview.md:7,quality.md:9,versioning.md:5,review.md:8}`, 언어 가이드 10개의 3번째 줄, `workflows/audit.md:36`. `src/language.rs:93-94` 는 사용자 파일에 `AGENTS.md` 를 그려 넣고, `src/generation.rs:1407` 은 호스트가 그것을 읽는다고 알려요. 출하 글 어디에도 `.claude/rules` 나 `CLAUDE.local.md` 지시문은 없어요.
- `src/go_state.rs:182-183` 은 `CLAUDE.md`, `AGENTS.md`, `AGENTS.local.md` 를 Go 세대 백업 대상 루트 파일로 두고 `CLAUDE.local.md` 는 사용자 것이라 남긴다고 적어요.
- 출력 스타일 템플릿은 이미 `MoAI 🗿` 배너예요(커밋 `02af21e018`).

## Non-goals

- 이 조각에서 `src/**`, `tests/**`, `skills/**`, 저장소의 `.gitignore`, 저장소 자신의 `AGENTS.md`·`AGENTS.local.md` 를 바꾸는 일.
- Codex·OMP 쓰기 코드를 지우는 일(별도의 Codex 철회 작업이 맡아요). 이 SPEC 은 그 줄을 건드리지 않고, 그쪽이 먼저 들어오면 겹치는 줄이 사라질 뿐이에요.
- 흐름 이름 `NeXT-Flow` 로의 개명, 스킬·에이전트 세트 재편(각각 별도 작업). 이 SPEC 은 `flow.md` 를 놓을 자리와 규칙만 정해요.
- 배너 문구 새로 정하기(`SPEC-20261121-style-banner` 가 이미 정했어요).
- `.worktreeinclude` 를 moai 가 쓰는 일(Risks 참고).
- 사용자의 실제 HOME 에 쓰는 일. 시험은 임시 HOME 과 임시 프로젝트만 써요.

## Requirements

- REQ-1: When no `AGENTS.md` exists at the project root, or `CLAUDE.md` already holds the managed block or the line `@AGENTS.md`, or the pick `import` or `move` was given, the generator shall plan the managed block `moai-workflow` inside `CLAUDE.md` at the project root.
- REQ-2 (REQ-1): IF `AGENTS.md` exists and none of those holds, THEN the generator shall plan no CLAUDE file and report that a pick is needed.
- REQ-3: The renderer shall produce at most 15 lines between the two marker lines of the managed block and state in them only the project id, the tools, the check command, the language, the versioning line, the flow name with its integration branch, the places of SPECs and cards, and the path of the flow rule file when one is planned.
- REQ-4: When `CLAUDE.md` exists without the managed block, the generator shall append the block on a new line after the last byte and keep every earlier byte.
- REQ-5 (REQ-4): IF the text between the block markers was edited after the recorded baseline, or `CLAUDE.md`, `CLAUDE.local.md` or the flow file, or a folder above one of them inside the project, is a link or is not valid UTF-8, THEN the generator shall report a conflict and write nothing to that file.
- REQ-6: When `CLAUDE.local.md` is absent, the generator shall create it only on the user's pick of a personal file. · failure: none — without the pick nothing is written, so nothing can fail
- REQ-7: Where the profile value `local_file` reads `tracked`, the generator shall treat it as `ignored` and say so in the preview.
- REQ-8: When the profile value `work_flow` names a known flow, the generator shall plan exactly one file `.claude/rules/moai/flow.md` that holds the text of that flow.
- REQ-9 (REQ-8): IF the profile value `work_flow` names no known flow or a known flow that has no shipped text, THEN the generator shall plan no flow file and name the accepted values.
- REQ-10: The renderer shall keep the flow text within 80 lines and 5 KiB and omit the `paths` frontmatter key so that the rule loads at session start.
- REQ-11: When the flow file exists, the generator shall replace it only when its bytes equal the bytes moai last wrote.
- REQ-12 (REQ-11): IF the flow file differs from the bytes moai last wrote, THEN the generator shall report a conflict and keep the file.

## Acceptance

- AC-1 (REQ-1…REQ-5): a temporary project without `AGENTS.md` gets the block in `CLAUDE.md` within 15 lines and with facts only; a project with `AGENTS.md`, no pick and no earlier block or import gets no CLAUDE file, while a project whose `CLAUDE.md` already holds the block or the import keeps getting its block refreshed; an appended block starts on a new line and keeps every earlier byte; an edited block, a `CLAUDE.md` that is a link, and one that is not UTF-8 are conflicts that write nothing.
  verify: `cargo test --locked --offline --test claude_md_layout_cli blocks_`
- AC-2 (REQ-6, REQ-7): `CLAUDE.local.md` is created only on the pick, and a profile with `tracked` plans like `ignored` with a preview note.
  verify: `cargo test --locked --offline --test claude_md_layout_cli local_`
- AC-3 (REQ-8…REQ-12): one flow file is planned from `work_flow`, an unknown flow plans none, the text has no `paths` key and stays within its limits, and a hand-edited flow file is a conflict.
  verify: `cargo test --locked --offline --test claude_md_layout_cli flow_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-claude-md-layout/**`, `.moai/specs/SPEC-20261122-claude-md-project-picks/**`, `.moai/specs/SPEC-20261122-claude-md-project-layer/**`, `.moai/specs/SPEC-20261122-claude-md-home-routing/**`, `.moai/specs/SPEC-20261122-claude-md-text-and-ignore/**`
- later slices, write: `src/agents.rs`, `src/agents_local_merge.rs`, `src/ops/agents_*.rs`, `src/update.rs`, `src/project_update.rs`, `src/generation.rs`, `src/launcher_init.rs`, `src/home_routing.rs`, `src/language.rs`, `src/claude.rs`, `skills/moai/**` (with `references/source-manifest.json`), `tests/**` that pin the names, `.gitignore`, `schemas/protocol/v1/agents-*.schema.json`
- forbidden in every slice before the last: `AGENTS.md`, `AGENTS.local.md`, `docs/agents-md-maintenance.md` of this repository, the real HOME, `~/.moai-cli`, `.claude/` of this checkout (Go-generation installed state)

## Plan

Slices in this order. Each is one SPEC-backed change with its own audit tier; the Codex removal work lands first where it touches the same lines (`home_routing.rs` HOSTS, `home_style.rs` block, `update.rs` HOSTS, `generation.rs` HOSTS).

0. Probe (no product code): an `InstructionsLoaded` hook in a `claude --worktree` session logs `file_path`, `memory_type`, `load_reason` and `cwd` to show whether the primary checkout's `CLAUDE.md` and `CLAUDE.local.md` load in a linked worktree. The result decides D-11.
1. Engine (this SPEC): file names become constants; `CLAUDE.md` is the block target; the block text is rewritten (facts, 15 lines, no Issue wording); `CLAUDE.local.md` and the flow file are new targets; the `AGENTS.local.md` merge, `claude_shadow` and `claude-import` leave the writer but their revert path stays for old receipts. Re-pin the tests that read these names (measured at the tip with a search over `tests/`: 81 files name `AGENTS.md` or `AGENTS.local.md`, 34 name `CLAUDE.md` or `CLAUDE.local.md`). The flow texts come from the four `work_flow` values (`src/profile.rs:171`); a value that is not yet a shipped text plans no flow file (REQ-9), and the block then omits the path (REQ-3). The name `NeXT-Flow` arrives with the naming work.
   Every `verify:` filter in these SPECs (`blocks_`, `local_`, `flow_` and the filters of the child SPECs) must match at least one test name; the slice counts the tests that ran, because a filter that matches nothing reports zero tests and counts as not run.
2. Project layer in `moai update` (`SPEC-20261122-claude-md-project-picks` for the picks and `SPEC-20261122-claude-md-project-layer` for receipt, link and UTF-8 refusal, baseline, token, revert order and key `project_files`). It reuses the `choices` shape of `project-update-plan`.
3. Routing (`SPEC-20261122-claude-md-home-routing`): `rules/moai.md` through the style-file pattern, the removal pick for the old block, receipt behaviour.
4. Ignore and texts (`SPEC-20261122-claude-md-text-and-ignore`): the managed ignore block, the repository `.gitignore` allow list, skill texts, language label, first launch, the text test and its exception list, the source manifest.
5. This repository's own `AGENTS.md` and `AGENTS.local.md`: split into `CLAUDE.md`, `.claude/rules/`, `CLAUDE.local.md` and `docs/`, only by the user's pick, with `tests/repo_agents_md.rs` replaced.

## Risks

- Claude Code stops reading `AGENTS.md` as soon as moai writes a `CLAUDE.md`. `move` and `import` keep the user's text reachable; `leave` writes no CLAUDE file for that reason.
- Orca profiles have their own config folder without `rules/`, and their `CLAUDE.md` imports `~/.claude/CLAUDE.md`. A file `~/.claude/rules/moai.md` would not load there, and removing the old block from `~/.claude/CLAUDE.md` would silently drop routing for those sessions. That is why removal is a separate pick (D-8). Whether Claude Code loads `rules/` under a profile folder is unverified.
- Whether a linked worktree session also loads the primary checkout's `CLAUDE.md` and `CLAUDE.local.md` is unverified (the docs are silent; the local measurement of 2026-11-13 was for `AGENTS.md`). A wrong guess either loads the file twice or leaves a worktree without it.
- Narrowing the repository `.gitignore` shows the installed Go-generation `.claude/` content in `git status`. The slice must use an allow list and read `settings.json` first (it may carry a per-machine `PATH`).
- Two earlier draft SPECs (`agents-conflict-choices`, `agents-choices-step`) and their tests describe `claude-import`. Old receipts that hold such rows must still revert.
- Many tests pin names. A new always-present preview key breaks `update_user_layer_cli`; the layer key is conditional for that reason.
- `SPEC-20261122-worktree-place` also edits the managed block template in `src/agents.rs`; whichever lands second rebases a text change.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The block lives in the root `CLAUDE.md`, not in `.claude/CLAUDE.md`. | Both load at launch (memory.md L55-62); the project-root file survives compaction (context-window.md L64). | Change the target constant. |
| D-2 | The pick order is `import` (recommended), `move`, `leave`; the user chooses. | `import` changes least and keeps one source of truth; nothing is copied twice. | Reorder the choices. |
| D-3 | Originals are never deleted or changed; `move` copies. | The rule that nothing of the user's is deleted; the card criterion says originals are kept. | None needed. |
| D-4 | A moai-generated `AGENTS.local.md` (the first-creation text) is not copied. | `is_generated_local` already recognises it (`a:396`); copying moai's own template would be noise. | Copy it too. |
| D-5 | The old managed block inside `AGENTS.md` is not copied and not touched; the preview says it is no longer updated. | Originals stay as they are; a fresh block is rendered in `CLAUDE.md`. | Strip it on a later pick. |
| D-6 | The flow rule has no `paths` key. | Unscoped rules load at launch and survive compaction (context-window.md L64, L79). | Add `paths` for a scoped rule. |
| D-7 | `rules/moai.md` is a whole file owned by moai, judged by the last-written hash like the style file. | The file is moai's alone; markers add nothing. | Use a marker block. |
| D-8 | Removing the old routing block is a separate pick and never the default. | Orca profiles import `~/.claude/CLAUDE.md` and have no `rules/` folder. | Make removal the default once the Orca path is verified. |
| D-9 | The repository `.gitignore` is narrowed with an allow list: ignore `.claude/*`, re-include `settings.json` and `rules/` after reading what they hold; `.claude/worktrees/` and `CLAUDE.local.md` stay ignored. | A blanket `.claude/` rule blocks tracked settings and rules; the checkout holds untracked Go-generation state. | Restore the two old lines. |
| D-10 | The project layer is a project layer of `moai update` with its own token outside the plan digest and a key shown only when there is work. | Matches `git_hooks`; keeps the 13 fixed keys. | Move it inside the digest. |
| D-11 | `.worktreeinclude` is not written until slice 0 answers whether the primary `CLAUDE.local.md` loads in a worktree. | Docs are silent; a copy could load twice. | Write the line after the probe. |
| D-12 | This SPEC family supersedes in part `SPEC-20261119-agents-conflict-choices` (REQ-11…REQ-14, REQ-16, REQ-17) and `SPEC-20261119-agents-choices-step` (REQ-5, REQ-8, REQ-9, REQ-11); `claude-import` stays only as a revert path. | The shadow model no longer holds when `CLAUDE.md` is moai's file. | Keep both SPECs and the pick. |
| D-13 | `local_file` value `tracked` is read as `ignored` with a preview note. | A personal file is not shared. | Keep `tracked` for `CLAUDE.local.md`. |
| D-14 | The banner stays as `SPEC-20261121-style-banner` decided. | Implemented and pinned. | None. |
| D-15 | The requirements are split over five SPECs with this one as parent. | Twelve requirements per SPEC is the size budget; each slice gets its own SPEC and audit. | Merge them. |

## Result

(Filled when the work is done. This is the research-and-SPEC slice only; no acceptance criterion is met yet.)
