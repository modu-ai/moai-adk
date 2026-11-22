---
id: SPEC-20261122-claude-md-flow-rule
title: "One flow rule file in the Claude config folder carries the NeXT-Flow git rule for every project"
status: draft
revision: 3
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Flow rule: rules/moai/flow.md

## Intent

사용자 결정(2026-11-22)으로 moai 의 공통 흐름 규칙은 프로젝트가 아니라 Claude 설정 폴더(`CLAUDE_CONFIG_DIR`, 없으면 기본 폴더)의 `rules/moai/flow.md` 한 파일에 설치돼요. 미리보기에 라우팅 파일과 함께 나오고 기본 폴더가 미리 골라져 있으며, `moai update` 의 확인 토큰이 곧 사용자의 선택이에요(`SPEC-20261122-claude-md-home-routing` 이 폴더·상태·영수증 규칙을 정해요). 이 SPEC 은 그 파일의 내용과 한도, 충돌 판정을 정해요. 프로젝트의 `CLAUDE.md` 블록은 사실만 담고 흐름 한 줄만 가져요(`SPEC-20261122-claude-md-layout` REQ-3). `paths` 키가 없는 규칙 파일은 세션 시작에 읽혀요. 설정 폴더의 규칙이 압축 뒤에도 남는지와 백그라운드 세션이 푸시나 풀 리퀘스트를 하지 않는지는 이 SPEC 의 코드가 아니라 임시 설정 폴더에서 재는 측정이에요(Plan 4).

개정 2(사용자 결정 두 가지, 2026-11-22): 착지 뒤 fast-forward push 는 사용자가 원격 복사를 켰을 때만 해요(기본 꺼짐). 그리고 백그라운드 작업이 남은 채 턴을 끝낼 때의 행동 한 줄(작업 목록 도구가 있으면 목록을 최신으로, 없으면 진행 보드)이 이 파일에 들어가요. 보드 모양은 출력 스타일이 맡고, 프로젝트 `CLAUDE.md` 에는 넣지 않아요. 흐름 전문은 스킬의 흐름 참조 한 곳에 두고, 이 파일은 짧은 규칙만 가져요. 그 참조를 가리키는 한 줄은 참조 파일이 생기는 스킬 조각에서 더해요.

개정 3(리더의 글 검토, 2026-11-22): 착지 줄은 "착지자(지금은 리더 세션; 카드가 착지자를 기록)"로, 안전 줄에 삭제 규칙(자기 작업트리나 임시 폴더 안에서 cd 뒤 상대 경로로만 지우고, 절대·홈 경로 rm 금지)을 더하고, 흐름 참조를 가리키는 한 줄을 지금 이 파일에 둬요.

## Requirements

- REQ-1: When the update plans the routing layer for a config folder, the planner shall plan exactly one file `rules/moai/flow.md` inside that folder, with the NeXT-Flow text, without the `paths` key, within 80 lines and 5 KiB. · failure: none — nothing is written without the update's confirm token
- REQ-2 (REQ-1): IF `rules/moai/flow.md` exists and its bytes differ from the bytes moai last wrote, or the file or a folder above it inside the config folder is a link or the file is not valid UTF-8, THEN the planner shall report a conflict and keep the file.
- REQ-3: The flow text shall carry the NeXT-Flow git rule (worktrees inside the project's `.claude/worktrees`, the lane submits head and base only, exactly one landing by the lander the card records, no push or pull request by a background session, the fast-forward push after a landing only when the user has turned the remote copy on), one line on ending a turn while background work still runs (the task list when the TaskCreate tool is offered, otherwise the progress board), a `## Compact Instructions` section, one line that the project's `CLAUDE.md` wins on a conflict, one line that points to the skill's flow reference, and one safety line (no force push, `reset --hard`, `git clean -f`, `--no-verify` or `git add -A`, and deletion only inside the own worktree or a temporary folder by a relative path).
- REQ-4: The flow text shall contain neither an explanation of the four letters of the flow nor the retired name, because the one shipped text that explains the flow is the skill text of the flow SPEC.
- REQ-5 (REQ-1): IF the source text of the flow rule is over its limits or holds the `paths` key, THEN the planner shall refuse the plan and name the limit.

## Acceptance

- AC-1 (REQ-1, REQ-5): one flow file is planned in each config folder the routing layer is planned for and written only with the update's confirm token, its text has no `paths` key and stays within its limits, and a source over a limit is refused.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_plan_`
- AC-2 (REQ-2): a hand-edited file, a link file, a link folder above it and a file that is not UTF-8 are conflicts and keep their bytes.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_conflict_`
- AC-3 (REQ-3, REQ-4): the text holds the four contents, exactly one push line that names the remote copy as its condition and exactly one background-work line, and neither the explanation of the four letters nor the retired name.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_text_`
- AC-4: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/home_routing/rules.rs`, `skills/moai/references/**` (the flow text and the source manifest), `tests/claude_md_home_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-flow-rule/**`
- read-only: `src/home_routing.rs`, `src/home_style.rs`, `src/host_home.rs`, `src/update.rs`
- forbidden: the real HOME and the real config folder, the Codex entries of `src/home_routing.rs`, the network

## Plan

1. The text of the flow rule from the content list of the redesign report (the git rule, Compact Instructions, the project wins, the safety line), checked by tests for limits and contents; the leader reads the wording before it lands.
2. The whole-file plan of `home_routing/rules.rs` judges the file like the style file (owned hash).
3. The source limit checks refuse an over-limit source before any plan.
4. Measure, with a local stand-in API and a temporary HOME and config folder: whether the flow rule and its Compact Instructions section reach a summarizer request and the first request after a compaction, and whether a background session keeps the flow text beside its built-in git instructions. The result goes into the report folder; a run against the real API is the user's choice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The safety line sits in the flow file. | The leader's note for the flow rule lists it there. | Put it in the routing file. |
| D-2 | The flow rule is planned and shown with the routing file for each config folder, the default folder preselected, and applied only with the update's confirm token. | The user's decision of 2026-11-22: the confirm of the update is the pick; the flow rule is not an extra pick. | Make it a separate pick. |
| D-3 | The commit-only text (no push, no pull request, submit head and base, land with the landing command) is the git rule of the flow text. | The leader's answer on the wording; the text is shown to the leader before it lands. | Describe landing in the skill text only. |
| D-4 | The push line says the lander pushes only when the user has turned the remote copy on, off by default. | User decision relayed in the leader's landing note of the skill package SPECs (2026-11-22). | Push after every landing. |
| D-5 | The background-work behaviour line sits in this file; the board's look stays in the output style; no project `CLAUDE.md` carries it. | User decision of 2026-11-22 recorded on this card: the rule file holds whatever output style the user picked. | Put the line in the output style only. |
| D-6 | This file holds the short rules; the full flow text lives in the skill's flow reference, and this file carries the one pointer line to it now (revision 3, leader review; the reference itself comes with the skill slice). | Pair decision with the skill package card, asked by the leader: only one side carries the full text, and the rule file loads in every session. | Carry the full text here. |

## Result

- AC-1 met: `cargo test --locked --offline --test claude_md_home_cli flow_plan_` runs 7 tests (one file per config folder, written only with the update's confirm token, no `paths` key, limits, an over-limit source refused).
- AC-2 met: `... flow_conflict_` runs 4 tests (edited file, link file, link folder above it, a file that is not UTF-8; each keeps its bytes).
- AC-3 met: `... flow_text_` runs 2 tests (the four contents are present, the single push line names the remote copy and the default off, the single background-work line names TaskCreate, the task list and the progress board; neither the four letters nor the retired name appear).
- AC-4 met: `moai check affected` on the slice reports fmt, ops rustfmt and clippy `--all-targets -D warnings` at exit 0.
- Open, not a code item: the measurement of Plan 4 (whether the rule and its Compact Instructions section survive a compaction, and a background session's git behaviour) has not been run; it needs the temporary config folder and the local stand-in API and is the user's call for a real-API run.
- Two lines of the flow text come from the SPEC and the leader's wording and were not checked against code: the protection hook line and the remote-check line. The push line follows the user's decision on the remote copy; the code that pushes is another card's. The leader reads the wording before it lands.
- closure_status: awaiting the independent audit.
