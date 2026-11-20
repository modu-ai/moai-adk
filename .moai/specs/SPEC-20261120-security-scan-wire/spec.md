---
id: SPEC-20261120-security-scan-wire
title: "The generated Claude Code hooks run the security-scan event after each write, and its would-notify lines have a log of their own"
status: draft
revision: 3
created: 2026-11-20
---

# The generated Claude Code hooks run the security-scan event after each write, and its would-notify lines have a log of their own

## Intent

Go 세대의 보안 3층 중 L1(방금 쓴 내용을 훑어 모델에게 알리는 층)은 SPEC-20261120-security-scan-notify(S2)로 Rust 엔진에 `security-scan` 이벤트가 생겼지만 **어디에도 연결되지 않았어요**. 그래서 Go 를 끄는 순간 이 층은 조용히 사라져요. 이 조각(S3)은 Claude Code 용 생성 훅 묶음에 연결을 더해요: `Write|Edit|MultiEdit|NotebookEdit` 뒤에 `moai hook-dispatch --host claude --event security-scan` 한 항목. 기본 모드는 `observe` 라서 연결해도 모델이 보는 것은 바뀌지 않고 로그 한 줄이 쌓일 뿐이에요(계획서 결정 2: 코퍼스로 오탐을 본 뒤에야 `enforce`).

모든 프로젝트의 생성물이 바뀌는 조각이라 가장 조심해요. 그래서 무엇이 바뀌고 무엇이 안 바뀌는지를 먼저 적어요.
- 바뀌는 것: `hooks::hook_set("claude")` 에 한 항목, 그 결과 `hooks-plan`·`hooks-apply`·`go-hooks-plan`·프로젝트 갱신이 쓰는 Claude `hooks` 객체와 그 항목 순서. 이미 MoAI 묶음이 깔린 프로젝트는 계획(`hooks-plan`)이 `update` 로 이 항목을 더하자고 제안하고, 사용자가 확인해야 써요(`claude::plan_settings` 가 소유권 영수증과 3자 비교를 하므로, 손대지 않은 이전 묶음은 `update`, 손댄 묶음은 `conflict` 로 바이트 그대로 남아요).
- 안 바뀌는 것: Codex 와 OMP 의 묶음(Codex 쓰기 연결은 S5, OMP 에는 PostToolUse 연결이 없어요), 다른 항목과 그 순서, 기본 모드, 어떤 경우에도 `deny`·`block` 을 내지 않는다는 성질.

S2 감사(노트 4492)가 이 조각 전에 정하라고 한 것 하나를 같이 해요: `would_notify` 줄이 `hook-engine.jsonl`(상한 4 MiB, 넘으면 새 줄을 못 써요)에 deny 줄과 함께 쌓이면, 연결 뒤 observe 줄이 로그를 채워 같은 로그의 deny 줄을 못 쓰게 되고 `merge_confirm enforce` 의 "로그에 못 쓰면 통과" 규칙이 발동할 수 있어요. 그래서 `would_notify` 줄은 자기 로그 파일에 써요. (감사자의 다른 권고, 경로 판정 함수를 하나로 합치는 일(N6)은 알림을 하나 놓칠 뿐이라 이 조각 범위 밖이고 후속이에요.)

## Non-goals

- Codex 연결(쓰기 가드와 PostToolUse, 계획서 S5), L2 Stop 알림(S4), ast-grep(S6), 전환 도구와 74곳(S7).
- `enforce` 로 바꾸는 일과 기본 모드 변경: 모드는 사용자가 `~/.moai-cli/hook-config.toml` 에서 켜요.
- `security-scan` 엔진의 모양·판정·알림 글 변경: S2 가 정했어요. 경로 판정을 하나로 합치기(N6), 감사의 문구·시험 지적(N1·N2·N3·N5·N9·N10)은 후속이에요.
- `hooks-status` 가 새 로그를 읽는 일: `security-scan` 은 아직 상태 요약에 들어 있지 않아요.
- 사용자가 직접 켜는 경로 규칙과 `[hooks.outside_project_writes]`: 별개예요.

## Requirements

- REQ-1: WHEN the hook set is generated for the host claude, the hook set shall hold a PostToolUse group whose command is `<binary> hook-dispatch --host claude --event security-scan` with the hook timeout of every other group. · failure: none
- REQ-2: The matcher of that group shall equal the tool names that `hook_engine::capabilities("claude")` lists for the event security-scan, joined with a vertical bar in the same order.
- REQ-3: The hook set of the hosts codex and omp shall hold no security-scan group.
- REQ-4: The new group shall stand directly after the Bash receipt group of the event PostToolUse and before the edit-diagnostics group that is generated while that rule is enabled.
- REQ-5: Every other group of the hook set shall keep its place and its content.
- REQ-6: WHEN `hooks-plan` runs on a project whose Claude hooks are the untouched set of the release before this change, the plan shall be an update that adds the group and keeps every other key of the settings file. · failure: none
- REQ-7: WHEN `hooks-plan` runs on a project whose Claude hooks already hold the group, the plan shall be unchanged. · failure: none
- REQ-8 (REQ-6, REQ-7): IF the Claude hooks of the project were edited by hand, THEN `hooks-plan` shall report a conflict and leave the bytes of the settings file untouched.
- REQ-9: WHEN security-scan observes a finding, it shall append its would_notify line to the log `security-scan.jsonl` of the receipts folder and not to `hook-engine.jsonl`.
- REQ-10 (REQ-9): IF the log `security-scan.jsonl` has reached its size limit, THEN security-scan shall write no line, print nothing and exit 0.
- REQ-11: The lines that other events append to `hook-engine.jsonl` shall not depend on the size of `security-scan.jsonl`.
- REQ-12: The legacy set with the old binary name `moai-cli` shall hold no security-scan group.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5): the Claude hooks object holds PostToolUse groups in the order Bash receipt, security-scan, and (while `[hooks.edit_diagnostics]` is enabled) edit-diagnostics; the security-scan group has the exact command, the timeout of the other groups and a matcher equal to the capabilities tools joined with `|`; every other group is as before · verify: `cargo test --locked --offline --test security_scan_wire_cli claude`
- AC-2 (REQ-3): the hooks objects of codex and omp hold no group that runs security-scan and equal what they were before this change in their other groups · verify: `cargo test --locked --offline --test security_scan_wire_cli other_hosts`
- AC-3 (REQ-6, REQ-7, REQ-8): on a temporary project, a Claude settings file holding the previous release's untouched MoAI set and its ownership receipt gets `action` update from `hooks-plan` with the new group listed and every other settings key kept, and after `hooks-apply` a second `hooks-plan` says unchanged; a settings file whose hooks were edited by hand stays a conflict with its bytes untouched · verify: `cargo test --locked --offline --test security_scan_wire_cli plan`
- AC-4 (REQ-9, REQ-10, REQ-11): the would_notify line lands in `security-scan.jsonl` and `hook-engine.jsonl` gets none; with `security-scan.jsonl` at its limit security-scan prints nothing, exits 0 and a deny line of tool-preflight is still appended to `hook-engine.jsonl` · verify: `cargo test --locked --offline --test hook_security_scan_cli mode` and `cargo test --locked --offline --test security_scan_wire_cli log`
- AC-5 (REQ-12): the legacy Claude set equals the pre-P4 fixture bytes and holds no group that runs security-scan, while the current set with the same binary name would · verify: `cargo test --locked --offline --test security_scan_wire_cli legacy` and `cargo test --locked --offline --test hooks_pre_p4_fixtures`
- AC-6 (REQ-1…REQ-12): every test file that reads the hook set, `hooks-plan` or `PostToolUse` passes after its expectation is updated for the new group and for no other reason, and the hooks-pre-p4 fixtures stay unchanged · verify: the test files named in the Result, run unfiltered and, for the corpus tool that reads the would_notify line, `python3 -m unittest tests.security_scan_corpus_test`

## Scope

- write: `src/hooks.rs` (the handler, its command, `hook_set`, the matcher constant, the legacy set that omits the group), `src/hook_engine.rs` and `src/hook_engine/scan.rs` (the log of the would_notify line), `tools/security_scan_corpus.py` (it reads that line from the log of the event), `tests/security_scan_wire_cli.rs` (new), `tests/hook_security_scan_cli.rs`, `.moai/specs/SPEC-20261120-security-scan-wire/**`, and the neighbour tests whose expectation must follow the new group, named in the Result and added to the reservation one by one by `card scope`
- read-only: `src/hook_legacy.rs`, `src/claude.rs`, `src/go_hooks.rs`, `src/ops/go_hooks_plan.rs`, `tests/fixtures/hooks-pre-p4/**`, `docs/hooks-generation.md`
- forbidden: `skills/**`, `schemas/**`, other projects' hook files, the real HOME and ledger, the Go repository

## Plan

Tests first (RED), then code, then mutants and neighbours, then the Result with one line per AC. The worktree is `~/.moai/worktrees/moai-adk/secwire` on a branch made from the head of the S2 branch (the same card; S2 must land first).

1. RED: `tests/security_scan_wire_cli.rs` with the fixture model of `tests/hooks_cli.rs` (a temporary project and HOME). Cases: `claude` (AC-1), `other_hosts` (AC-2), `plan` (AC-3: the previous release's set is built in the test from the current `hooks_object` with the new group removed, and its ownership receipt by running `hooks-apply` against an older-set fixture if one can be had, else by the same writer the plan reads), `log` (AC-4's second command). `tests/hook_security_scan_cli.rs` changes where its `mode` tests read the line.
2. GREEN: `Handler::SecurityScan`, `CLAUDE_SECURITY_SCAN_MATCHER`, its arm in `command_with`, the `hook_set` insertion after the Bash receipt group (Claude only); `SECURITY_SCAN_LOG` and the append in the would_notify writer of `src/hook_engine.rs`.
3. Neighbours: run the test files that name `hook_set(`, `hooks-plan` or `PostToolUse` (19 files at this head: `binary_state_cli`, `cli_name_moai`, `edit_diagnostics_cli`, `go_hooks_migrate`, `hook_dispatch_cli`, `hook_legacy_cli`, `hook_mode_cli`, `hook_security_scan_cli`, `hook_status_cli`, `hooks_cli`, `hooks_conflict_choices_cli`, `hooks_pre_p4_fixtures`, `host_home_cli`, `lane_stop_hook_cli`, `mode_workflow_steps`, `no_moai_on_path`, `project_update_hooks_cli`, `status_line_cli`, `update_hooks_cli`); update only the expectation that the new group changes, and widen the reservation by `card scope` file by file.
4. Mutants (at most 5): the group dropped, the matcher narrowed by one tool, the group placed before the receipt group, the group added for codex, the would_notify line written to the old log.

Research notes (read 2026-11-20 on the S2 head `eb2c17a655`): `src/hooks.rs` (`HOOK_SET`, `hook_set`, `hooks_object_with`, `edit_diagnostics_timeout` as the precedent of a conditional PostToolUse group, `command_with`, `append_receipt_line`, `RECEIPT_LOG_LIMIT`); `src/hook_legacy.rs` (recognition by exact equality with the generator's own output; a rewrite goes through the writers that own the files); the S2 audit note 4492 (N6, N11).

## Risks

- Every Write, Edit, MultiEdit and NotebookEdit now starts one more hook process: the S2 latency report (added median 10 ms or less on small inputs, 50 ms or less on 1 MiB) is the cost; observe mode keeps the model's view unchanged.
- A project that already runs the MoAI set sees `hooks-plan` propose an update; nothing is written without the user's confirmation, and a hand-edited set stays a conflict.
- Test files that count hook groups will fail until their expectation follows; each is updated for the new group only and named in the Result.
- Two groups of PostToolUse match the same tools (edit-diagnostics and security-scan): the host runs matching groups side by side; neither reads the other's output.
- A project that runs the security-scan group on a machine whose `moai` is the Go build gets exit 1 from it: the existing binary gate (`hook_legacy::gate`) refuses to write a set to such a PATH, which is why the generated binary is `moai` and gated.
- A project that holds the previous release's set without the ownership record (a fresh clone that commits `.claude/settings.json` but not `.moai-cli/`) gets `conflict` on `hooks` from `hooks-plan`, with no replacement choice offered, and `project update` words it as a key the user set; every untouched previous-release install reads `drifted` in `hooks-status` until the update is applied. This follows from the three-way design that every change of the generated set meets; a follow-up decides the wording and a way out.

## Result

All six criteria are met, on the head named in the READY note (a branch made from the head of the first slice, `eb2c17a655`, so the first slice must land first). Commits on top of the first slice: SPEC, RED tests (`Moai-Phase: red`), then the wiring and the neighbour expectations.

- AC-1: met — `cargo test --locked --offline --test security_scan_wire_cli claude`: exit 0, 2 passed. The Claude hooks object holds PostToolUse groups Bash receipt then security-scan, the group has the command `moai hook-dispatch --host claude --event security-scan`, the hook timeout and the matcher `Write|Edit|MultiEdit|NotebookEdit` (equal to the capabilities answer joined with `|`); the other events keep their group counts; run through the binary with `[hooks.edit_diagnostics]` enabled the order is receipt, security-scan, edit-diagnostics.
- AC-2: met — filter `other_hosts`: exit 0, 1 passed. The hooks objects of codex and omp hold no command that names security-scan and keep one PreToolUse and one PostToolUse group.
- AC-3: met — filter `plan`: exit 0, 2 passed. A settings file holding the set without the new group and its ownership receipt is planned as `update` that lists the group, keeps the unrelated key and, after `hooks-apply`, gives `unchanged`; the same file with one hand-edited command is a `conflict` and its bytes stay.
- AC-4: met — `--test hook_security_scan_cli mode`: exit 0, 4 passed (the would_notify line is read from `security-scan.jsonl` and `hook-engine.jsonl` holds none) and `--test security_scan_wire_cli log`: exit 0, 2 passed (one test: an observed finding leaves `hook-engine.jsonl` without a line, one byte below the limit `security-scan.jsonl` still takes a line and at the limit it does not; the other: with `security-scan.jsonl` at exactly the limit the event prints nothing, exits 0 and leaves its size unchanged, and a deny of tool-preflight still reaches `hook-engine.jsonl`).
- AC-5: met — filter `legacy`: exit 0, 1 passed (the legacy Claude set has no security-scan group, the current one has two PostToolUse groups) and `--test hooks_pre_p4_fixtures`: exit 0, 6 passed (the legacy generator still equals the pre-P4 fixture bytes; the fixtures are unchanged).
- AC-6: met — the neighbour test files, each run unfiltered, exit 0: binary_state_cli 11, cli_name_moai 5, edit_diagnostics_cli 14 (1 ignored as before), go_hooks_migrate 17, hook_dispatch_cli 56, hook_legacy_cli 10, hook_mode_cli 5, hook_security_scan_cli 16, hook_status_cli 10, hooks_cli 62, hooks_conflict_choices_cli 24, hooks_pre_p4_fixtures 6, host_home_cli 12, lane_stop_hook_cli 5, mode_workflow_steps 43, no_moai_on_path 2, project_update_hooks_cli 23, security_scan_wire_cli 7, status_line_cli 16, update_hooks_cli 11, and `--lib hook_engine` 17 and `--lib hooks::` 16. Four of them needed an expectation change for the new group and nothing else: hooks_cli (generated hooks 4 to 5), no_moai_on_path (plan entries 4 to 5), hook_legacy_cli (the commands a fresh install writes, 4 to 5, in a field of its own because a rewrite still replaces 4 entries), hook_security_scan_cli (Claude lists the event once, Codex still none). The corpus tool test `python3 -m unittest tests.security_scan_corpus_test`: exit 0, 3 tests OK (the tool reads the would_notify line from `security-scan.jsonl`; before the repair it reported no finding for any file).
- Found while running the neighbours: the legacy set was the current generator with the old binary name, so the new group first leaked into it; hooks_pre_p4_fixtures and hook_legacy_cli (`drifted` instead of `legacy_binary`) failed on it. The code was fixed (`hook_set_of` leaves the group out for the old name) and REQ-12 and AC-5 were added; no test was loosened.
- Audit round 1 (head `0516fffc3d`): B1, `tools/security_scan_corpus.py` still read the would_notify line from `hook-engine.jsonl`, so the corpus tool reported no finding and its python test failed (1 failure, 1 error); the check set that selects tests by changed file did not pick that test and no grep of the log name had been run. Repaired: the tool reads `security-scan.jsonl`; the test passes (3 OK) with a binary built from the repaired sources. N1, the second half of the log test also passed on the old code: a test was added that fails when the line reaches the engine log and that pins the limit edge at limit-1 and limit (verified: mutant M5 now fails it); the TDD lock was lifted with `tdd reopen` and the reason. Opinion taken: `hook_set_of` gives the group only to `HOOK_BINARY` and fails closed for any other name. Follow-ups, not in this change: N2 documents that still list PostToolUse as Bash only (`docs/hooks-generation.md:34`, `docs/claude-adapter-contract.md:36`, `docs/rust-product-path.md:46` and `:1006`, the last one without the new log); N3 the no-record conflict and the `project update` wording (`src/project_update.rs` near 1020 to 1040, read not run); N4 `drifted` until the update is applied.
- Mutants (python3 of a script that edits the sources, runs the targeted tests and restores them; 5 of 5 killed, none a compile error, `restored: True`, `git status` empty afterwards): the group dropped (5 tests fail), the matcher narrowed by one tool (1), the group placed before the receipt group (2), codex given the Claude branch (1), the would_notify line written to the old log (1).
- Not run: `project update` end to end for a previous-release set, with and without the ownership record; a real Claude Code session (that the host starts the new group and shows no output in observe mode rests on the S2 report and the host documentation read for S2); the set of the previous release was built in the test from the current generator without the group, not captured from an installed release; the hook latency was not measured again (the engine is unchanged, only a group was added; the S2 figures stand); the Windows and Linux builds.
