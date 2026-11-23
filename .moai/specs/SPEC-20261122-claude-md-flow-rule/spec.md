---
id: SPEC-20261122-claude-md-flow-rule
title: "The rules files in the Claude config folder are short, bind only a MoAI project and carry no role or command detail"
status: draft
revision: 5
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Flow rule: rules/moai/flow.md

## Intent

사용자 결정(2026-11-22)으로 moai 의 공통 규칙은 프로젝트가 아니라 Claude 설정 폴더 `~/.claude` 의 `rules/moai.md` 와 `rules/moai/flow.md` 두 파일에 설치돼요. 미리보기에 두 파일이 함께 나오고 기본 폴더가 미리 골라져 있으며, `moai update` 의 확인 토큰이 곧 사용자의 선택이에요(`SPEC-20261122-claude-md-home-routing` 이 폴더·상태·영수증 규칙을 정해요). 이 SPEC 은 두 파일의 내용과 한도, 충돌 판정을 정해요. `paths` 키가 없는 규칙 파일은 세션 시작에 읽혀요.

개정 4(2026-11-23, 리더 결정 04:39Z·04:46Z, 사용자 확정 P0): 공통 rules 를 처음부터 다시 설계했어요. 이 SPEC 이 최소 홈 규칙 작업 항목(MoAI 사용 조건의 짧은 활성·안전 규칙, MoAI 가 아닌 프로젝트에 영향 없음)의 요구를 흡수하고, 그 항목은 이 변경의 착지로 닫아요. 근거는 Go 판 규칙·스킬이 아니라 세 가지예요: (1) 현재 Rust MoAI-ADK 의 실제 동작: 훅이 이미 막거나 확인하는 일(`src/hook_guards.rs`: 보호 브랜치 force push 거절, `git reset --hard`·`git clean -f`·그 밖의 force push 확인)은 규칙에 다시 적지 않고, 레인·카드·착지 명령은 스킬의 `references/flows/next-flow.md` 가 필요할 때만 읽혀요. (2) 공식 문서(code.claude.com/docs/en/memory, 2026-11-23 읽음, 본문이 일부만 원문으로 읽혔어요): "Personal rules in `~/.claude/rules/` apply to every project on your machine." 와 "target under 200 lines per CLAUDE.md file. Longer files consume more context and reduce adherence." 그래서 사용자 수준 파일은 모든 프로젝트에서 읽히니 짧아야 하고 MoAI 프로젝트에만 적용된다고 먼저 말해야 해요. (3) 연구: arXiv 2602.11988 "Evaluating AGENTS.md" 초록의 "providing context files does not generally improve task success rates" 와 "increasing inference cost by over 20% on average"; TACL 2024 "Lost in the Middle" 초록의 "performance is often highest when relevant information occurs at the beginning or end of the input context". 그래서 본문은 짧게, 적용 조건을 맨 앞에 둬요. 이 읽기에서 Go 판 자료는 쓰지 않았어요.

개정 5(2026-11-23, 독립 감사 FAIL 의 수리, `SPEC-20261122-claude-md-home-routing` 개정 4): 사본은 `~/.claude` 하나이고 다른 설정 폴더는 사본 없는 줄이에요(REQ-1, REQ-8). 흐름 글의 조건 문장은 REQ-6 대로 `or when the user invokes MoAI` 까지 말해요(`flow_text_binds` 가 두 글 모두 확인).

개정 3 까지의 내용(NeXT-Flow git 규칙 전문, 작업 목록 도구 줄, Compact Instructions, 안전 줄)은 아래 결정에 따라 짧은 증거·인계 규칙으로 바뀌었어요.

## Requirements

- REQ-1: When the update plans the routing layer, the planner shall plan exactly one file `rules/moai/flow.md` inside `HOME/.claude`, without the `paths` key, within 80 lines and 5 KiB. · failure: none — nothing is written without the update's confirm token
- REQ-2 (REQ-1): IF `rules/moai/flow.md` exists and its bytes differ from the bytes moai last wrote, or the file or a folder above it inside the config folder is a link or the file is not valid UTF-8, THEN the planner shall report a conflict and keep the file.
- REQ-3: The flow text shall state the evidence and handoff rules (the smallest change that meets the acceptance criteria, the tested revision and the exact checks recorded with every skipped or failed one, missing evidence never a PASS, implementation and review kept distinct, a blocked record, no bypass of ownership, permissions or a release gate, no secrets in shared artifacts, landing and publishing only through the project's workflow, deletion by a relative path inside the own worktree or a temporary folder), one line that the project's `CLAUDE.md` wins on a conflict, and one line that points to the skill's flow reference.
- REQ-4: The flow text and the routing text shall carry no lane role, card or landing command, worktree place, rule that a hook enforces, task-list tool name or `## Compact Instructions` section, and neither the explanation of the four letters of the flow nor the retired name.
- REQ-5 (REQ-1): IF the source text of the flow rule is over its limits or holds the `paths` key, THEN the planner shall refuse the plan and name the limit.
- REQ-6: Each of the two shipped rules texts shall state, before any instruction, that it applies only in a project configured for MoAI or when the user invokes MoAI.
- REQ-7: While a session runs in a project that is not configured for MoAI, the rules that the config folder loads shall hold only the conditional text of REQ-6.
- REQ-8 (REQ-1): WHERE the `rules` folder of another config folder is a link to `HOME/.claude/rules` (also while the target is not there yet), the planner shall report each rules file as `linked` unless it is a conflict, and write nothing for it.

## Acceptance

- AC-1 (REQ-1, REQ-5): one flow file is planned in `HOME/.claude` and written only with the update's confirm token, its text has no `paths` key and stays within its limits, and a source over a limit is refused.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_plan_`
- AC-2 (REQ-2): a hand-edited file, a link file, a link folder above it and a file that is not UTF-8 are conflicts and keep their bytes.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_conflict_`
- AC-3 (REQ-3, REQ-4): the flow text holds the evidence and handoff rules, says once that the project's `CLAUDE.md` wins and points once to the flow reference; neither text carries role or command detail, hook-enforced rules, a task-list tool name or a Compact Instructions section.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_text_`
- AC-4 (REQ-6): both texts open with the condition and the routing text says what loading it does not do.
  verify: `cargo test --locked --offline --test claude_md_home_cli flow_text_binds`
- AC-5 (REQ-7): with the rules applied to a throwaway config folder, the real Claude Code loads them in a plain config folder, in a profile whose `rules` folder is a link to it and in a profile that imports the two files, loads nothing in a profile that references neither, and a project that is not configured for MoAI keeps its own files and gets nothing written.
  manual: measured with the real Claude Code against a local stand-in API, not a test of this repository; the probe scripts and their output are kept in the report folder named in the Result, and the project folder is listed before and after.
- AC-6 (REQ-8): a profile whose `rules` folder links into the default one (also with no target yet) is a reference with nothing written and no state of its own, a reference to an edited file stays a conflict, and a link that leaves the default folder is `not_linked`.
  verify: `cargo test --locked --offline --test claude_md_home_cli folders_a_`
- AC-7: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/home_routing/rules.rs`, `src/home_routing.rs` (the loop that turns another folder into a consumer row), `skills/moai/references/**` (the two texts and the source manifest), `tests/claude_md_home_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-flow-rule/**`
- read-only: `src/home_style.rs`, `src/host_home.rs`, `src/update.rs`, `src/hook_guards.rs`
- forbidden: the real HOME and the real config folder, the Codex entries of `src/home_routing.rs`, the network (the measurement talks to a stand-in server on the loopback address with a dummy key), any Go-generation rule or skill as an input

## Plan

1. The two texts from the comparison in Decisions D-7, checked by tests for limits, contents, the opening condition and the absence of role and command detail.
2. The whole-file plan of `home_routing/rules.rs` judges each file like the style file (owned hash); the source limit checks refuse an over-limit source before any plan.
3. Measure with the real `claude -p` against a local stand-in API, a dummy key and a throwaway HOME outside the session scratchpad (Claude Code refuses to load instruction files below a scratchpad folder; its debug log says so).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flow rule is planned and shown with the routing file for `~/.claude` (the other config folders are rows that write nothing), and applied only with the update's confirm token. | The user's decision of 2026-11-22: the confirm of the update is the pick. | Make it a separate pick. |
| D-2 | The lane, card and landing detail (worktree place, head-and-base submission, exactly-once landing by the lander, the protection hook, the push after a landing only with the remote copy on, remote checks never blocking) leaves the rules and stays in the skill's `references/flows/next-flow.md`, which the flow text points to. | The user-level file loads in every project; the flow text applies where a card or a landing is involved, and the skill reads that reference then (docs: user-level rules apply to every project). The skill reference already holds each of those statements. | Put the git rule back into the flow text. |
| D-3 | The safety line shrinks to the deletion rule. Force push to a protected branch is refused and `reset --hard`, `clean -f` and other force pushes are confirmed by the hooks already (`src/hook_guards.rs`), so the rules do not repeat them. | A rule that code enforces does not belong only in a rule (leader decision 2026-11-23T04:39Z); the deletion rule has no code guard yet, so one line stays until the delete guard lands and is then removed. | Restore the full safety line. |
| D-4 | The line on ending a turn while background work runs is removed. | The user's rule of 2026-11-23 retires the task-list branch; the progress board belongs to the output style the user picked, not to a rules file. | Restore a one-line board rule. |
| D-5 | The `## Compact Instructions` section is removed. | Whether a rules file's section survives a compaction was never measured; the card and reservation facts it listed are in the ledger. | Restore the section after a measurement. |
| D-6 | The flow text names one pointer line to the skill's flow reference. | The pair decision with the skill package: only one side carries the full text. | Carry the full text here. |
| D-7 | Comparison with the leader's candidate (`rules-redesign-2026-11-23/candidate/rules`): the activation condition (adopted, both texts, REQ-6), the route through the installed `moai` skill reading only the selected workflow (adopted; the host invocation `{invoke}` and the base agent names stay from the current routing text), look for an existing card or SPEC first (adopted), smallest change and evidence (adopted, with the "never a PASS" sentence), implementation and review distinct (adopted; the routing text keeps the unconditional "never audits it"), blocked record and no bypass (adopted), private data and publishing only through the configured workflow (adopted), deletion line (kept, D-3). Not adopted from the current texts: everything D-2 to D-5 removes. | Each adopted line is project-generic and short; none is enforced by a hook. | Adopt the candidate verbatim. |
| D-8 | No Go-generation rule or skill was read as an input or adopted; the one Go inventory file named in the first leader note was opened for its header lines only, before the leader's correction, and nothing was taken from it. | The user's correction of 2026-11-23 (leader note 04:46Z). | None. |

## Result

All numbers come from the `verify` commands run on the head that carries this text (throwaway HOME for every test; the real HOME and config folder were never read or written).

- AC-1 met: `claude_md_home_cli flow_plan_` runs 7 tests (one flow file per config folder, written only with the confirm token, no `paths` key, within 80 lines and 5 KiB, an over-limit source refused).
- AC-2 met: `claude_md_home_cli flow_conflict_` runs 4 tests (an edited file, a link file, a link folder above it and a non-UTF-8 file are conflicts and keep their bytes).
- AC-3 met: `claude_md_home_cli flow_text_` runs 4 tests (evidence and handoff rules, the project `CLAUDE.md` wins once, one pointer to the flow reference, no role or command detail).
- AC-4 met: `claude_md_home_cli flow_text_binds` runs 1 test (both texts open with the condition).
- AC-5 met by measurement, not by a test of this repository: see the Measured line of `SPEC-20261122-claude-md-home-routing`; the three marker sentences reach the model through the default folder, a `rules` link and a `CLAUDE.md` import, none through a profile with no reference, and the project folder is listed identical before and after.
- AC-6 met: `claude_md_home_cli folders_a_` runs 5 tests, among them the three of this requirement (a linked profile is a reference, a reference to an edited file stays a conflict, a link out of the default folder is a conflict with nothing written).
- AC-7 met: `cargo fmt --check` and `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0.
- Guard tests fail without their guard: five mutants on this revision each turned the named test red and were restored byte for byte (M7 reference off, M8 conflict relabelled, M9 condition removed, M10 "never a PASS" removed, M11 command word added); the list is in the Result of `SPEC-20261122-claude-md-home-routing`.
- Open: the language line and a "PASS needs a command run" line beyond the existing "missing evidence is never a PASS" are absent; the leader decides.
- closure_status: awaiting the independent audit.
