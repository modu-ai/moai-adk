---
id: SPEC-20261120-shipped-workflows-commands
title: "Shipped skills and agents name moai commands only — no repository-only tools script"
status: draft
revision: 3
created: 2026-11-21
---

## Intent

moai 를 설치한 모든 사용자는 `skills/moai` 와 `agents` 를 받아요. 배포 글이 "moai-adk 저장소에서 `tools/X` 를 돌려요"라고 말하면 다른 사용자는 쓸 수 있는 것이 없어요(사용자 지시 2026-11-20: 로컬 개발 도구와 오픈소스 기능을 섞지 마라). 이 SPEC 은 두 가지를 해요. 하나는 배포 파일에 남은 `tools/` 스크립트 이름을 빼고 스캔 시험으로 다시 들어오지 못하게 막는 것이고, 다른 하나는 배포된 READY·점검 문장이 이름하는 `moai` 명령이 새 HOME 과 새 프로젝트에서 설치된 moai 만으로 답하는지 끝까지 시험하는 것이에요.

개정 3판은 통합 브랜치 `NeXT` 의 tip `64b948a105` 에서 조사를 다시 했어요. 개정 2판의 21 줄 조사표는 낡았어요. 그 뒤 `NeXT` 가 `verify.md`·`mode-agent.md`·`rule-history.md` 를 지웠고 `build-folder.md` 규칙을 `run.md` 로 옮겼고 `self-check.md` 를 Rust 전용으로 다시 썼어요. 개정 2판의 가지(`spec/shipped-workflows-commands-20261120-r2`)는 `NeXT` 에 들어오지 않았고 되살리지 않아요.

Survey (tip `64b948a105`, `git grep -nE '(^|[^A-Za-z0-9_./-])tools/[A-Za-z0-9_./-]+\.(sh|py|json)' -- skills agents`, 3 파일 10 곳):

| Shipped file | `tools/` reference | move | target |
| --- | --- | --- | --- |
| `workflows/update.md`:35 | `tools/build_folder.py remove` | (b) | 조건 문장만 남기고 저장소 도구 이름을 뺀다(미리보기가 `trim` 으로 표시한 행만) |
| `references/audit.md`:52 | 티어 2 바닥 목록의 `tools/` 항목 7개 | (b) | 한 대역어 "the repository's own check tools" 로 바꾼다(전례: 제2 호스트 폴더 대역어) |
| `references/audit.md`:52 | `tools/repo_scans.json` 이 바닥을 올리는 조건 문장 | (b) | "the repository's scan registry" 로 바꾼다 |
| `references/source-manifest.json`:33 | `tools/legacy_verdicts.py` (status 이력 문장) | (b) | 도구 이름을 뺀 이력 문장 |

`agents/impl.md` 의 `.moai/tools/`(사용자 프로젝트가 만드는 지원 스크립트 폴더)와 `workflows/project.md` 의 `golang.org/x/tools/gopls`(외부 패키지 경로)는 이 저장소의 스크립트 이름이 아니라서 조사표에 넣지 않아요.

## Non-goals

- `moai check-ready` 를 만들지 않아요. 개정 2판이 이 명령을 부르는 배포 문장은 `NeXT` 에 더는 없어요. 문장이 없는 명령은 요청 없는 기능이에요.
- `src/gate.rs` 의 `TIER2_PREFIXES` 를 바꾸지 않아요. 코드가 `tools/` 경로를 하드코딩하는 문제는 코드 질문이고 이 SPEC 의 범위가 아니에요.
- `tools/*.sh`·`tools/*.py` 자체를 고치거나 지우지 않아요.
- 개정 2판의 `docs/moai-adk-tools.md` 를 되살리지 않아요. 옮길 절차 문장이 `NeXT` 에 남아 있지 않아요.

## Requirements

- REQ-1: Where a file ships under `skills/` or `agents/`, the file shall name no `tools/<path>.sh`, `tools/<path>.py` or `tools/<path>.json` script of this repository. · failure: none — a static scan, no runtime event.
- REQ-2: While the update workflow describes build folders outside the worktrees, it shall state that a fresh preview must still mark each folder `trim` before anyone removes it, and name no repository-only tool. · failure: none — a text rule.
- REQ-3: While the audit contract lists the tier 2 floor, it shall name the repository's own check tools through one stand-in phrase that stands for every `tools/` prefix of `TIER2_PREFIXES`. · failure: none — a text rule pinned by a test.
- REQ-6 (REQ-3): IF the list lacks the stand-in phrase, repeats it or lists a `tools/` entry by name, THEN the test that compares the list with the code shall fail.
- REQ-4: When the shipped READY and check texts name a `moai` command, the installed `moai` in a temporary HOME and a temporary project shall answer that command and every `--flag` the text shows.
- REQ-5 (REQ-4): IF a command or a flag that a shipped text names is unknown to the installed `moai`, THEN the end-to-end test shall fail and name it.

## Acceptance

- AC-1 (REQ-1): A scan test walks every file under `skills/` and `agents/` and finds no `tools/<path>.(sh|py|json)` that is not preceded by `/`, `.`, a letter, a digit, `_` or `-`; its allow-list is empty.
  verify: `cargo test --offline --locked --test skill_generalization_scan`
- AC-2 (REQ-2): `workflows/update.md` keeps the `trim` condition and the pinned phrases and names no repository tool.
  verify: `cargo test --offline --locked --test agent_defs_package --test update_route`
- AC-3 (REQ-3, REQ-6): The tier 2 list of `references/audit.md` equals the code floor with the stand-in, and the planted wrong lists fail.
  verify: `cargo test --offline --locked --test next_flow_skill_scan refs_audit`
- AC-4 (REQ-4, REQ-5): The end-to-end scenario derives the command list from the shipped texts, runs the read-only commands in a temporary HOME and project, runs the others with `--help`, and fails on a planted unknown command.
  verify: `cargo test --offline --locked --test e2e_temp_home_cli scenario_every_command`
- AC-5 (REQ-1, REQ-3): The source manifest matches the shipped bytes and the neighbouring tests stay green.
  verify: `cargo test --offline --locked --test product_cli skill_package:: --test source_manifest_facts --test self_audit_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-shipped-workflows-commands/`, `skills/moai/references/audit.md`, `skills/moai/workflows/update.md`, `tests/skill_generalization_scan.rs`, `tests/next_flow_skill_scan.rs`, `tests/e2e_temp_home_cli.rs`, `tests/self_audit_cli.rs`; shared files by line lease: `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`
- read-only: `src/gate.rs`, `tools/*.sh`, `tools/*.py`
- forbidden: deleting or editing a script under `tools/`, changing a shipped text's meaning beyond naming no repository tool, a shipped file naming a `tools/` script after this change lands

## Plan

1. SPEC (this file) first, then the tests that fail: the scan (`skill_generalization_scan.rs`) and the end-to-end scenario (`e2e_temp_home_cli.rs`).
2. `tests/self_audit_cli.rs`: the two floors (`mentions >= 3`, `seen >= 1`) become `== 0`, with their comments.
3. Texts: `update.md` step 7 clause, `audit.md` floor list and registry sentence, with the stand-in check in `next_flow_skill_scan.rs` (`tier2_list_problems` and its planted cases).
4. Registry and manifest under `moai card line-lease`: one new scan entry (a scan added is growth, so the path floor stays at tier 1), then the sha256 of `audit.md` and `update.md` by `shasum -a 256` and the reworded status sentence.
5. Neighbours: run the tests that read the changed files and `moai repo-scans --refs NeXT HEAD`.

## Risks

- `tests/next_flow_skill_scan.rs` is held by another card; both changes touch different functions, and the lander merges the head onto the tip once.
- The stand-in phrase absorbs any later `tools/` prefix added to `TIER2_PREFIXES`; the test keeps at least one such prefix in the code so the phrase is never empty. The same trade was made for the retired second host's folder.
- The new scan walks every file of `skills/` and `agents/`; the registry entry that picks it for `agents/` changes is a new `always` scan, observed first by running it on the tip.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | `moai check-ready` is dropped (superseded) | The shipped text that named it was deleted from `NeXT`; no remaining text needs it | Build a command when a shipped text asks for one |
| D2 | The floor list names the repository's tools through one stand-in phrase; no code change | The list is tied to `TIER2_PREFIXES` by a test; the stand-in keeps that tie and the shipped text names no repository script | Spell the entries again and widen the scan's allow-list |
| D3 | The scan matches `sh`, `py` and `json` | Three of the seven floor entries are `.json` files | Narrow the extension list |
| D4 | The scan gets its own registry entry with scope `skills/` and `agents/` | The existing group's scope names only two agent files; changing a scope is not registry growth | Add `agents/` to the existing scope under tier 2 |
| D5 | The criterion "every check the shipped VERIFY/READY steps name" is read as every `moai` command the READY and check texts name (`self-check.md`, `run.md`, `work.md`, `next-flow.md`) | The VERIFY workflow no longer exists | Reword the criterion on the card |
| D6 | The change is tier 1 by its paths (no `TIER2_PREFIXES` entry, no deletion, permission, gate or HOME change); the card says tier 2, so this SPEC stays and the leader decides whether an approval is needed | `skills/moai/references/audit.md` is not a floor path; a scan added to the registry is growth | The leader sets the tier |
| D7 | `cargo`, `git` and `shasum` are the user's own toolchain, not repository scripts, and may be named | They exist on any developer machine | — |
