---
id: SPEC-20261121-security-turn-wire
title: "The generated Claude Code hooks run the security-turn event when a turn ends, and the docs and the capabilities answer say so"
status: draft
revision: 1
created: 2026-11-21
---

# The generated Claude Code hooks run the security-turn event when a turn ends, and the docs and the capabilities answer say so

## Intent

The turn-end scan of the previous slice (SPEC-20261121-security-scan-turn) is in the engine as the event `security-turn`, but no generated hook set calls it, so it is reachable only by a hand-run `hook-dispatch`. When the Go generation is switched off its Stop-time security notice would be gone. This slice adds the one `Stop` group to the Claude Code hook set: `moai hook-dispatch --host claude --event security-turn`. The default rule mode is `observe`, so wiring it changes nothing the model sees: one `would_notify` line per new finding set lands in `security-scan.jsonl` (decision 2 of the plan: `enforce` only after the corpus shows the false positives).

This changes the generated files of every project that runs the MoAI set, so what changes and what does not is stated first.
- Changes: `hooks::hook_set("claude")` gets one entry; so does the `hooks` object that `hooks-plan`, `hooks-apply`, `go-hooks-plan` and the project update write. A project that already holds the MoAI set sees `hooks-plan` propose an `update` that adds the group; a user-edited set stays a `conflict` with its bytes untouched (the three-way compare of `claude::plan_settings` decides, as for the security-scan group). The `hook_engine::capabilities` answer for Claude lists the event, and the two engine texts that say "not wired into generated settings in this release" say what is true now.
- Does not change: the hook sets of Codex and OMP (the Codex lane Stop hook stays the only Stop group of a Codex set), the legacy set with the old binary name, every other group and its order, the default mode, and the property that no mode ever blocks.

## Non-goals

- Codex or OMP wiring of the turn-end event: both lack the event (the engine answers a usage error for them). Codex write guards (plan slice S5), ast-grep (S6) and the conversion tool with its 74 places (S7) are later slices.
- Switching `security_turn` to `enforce` or changing a default: the user sets it in `~/.moai-cli/hook-config.toml`.
- A change of what the event reads, answers or logs: the previous slice decided that; its open follow-ups (a deletion-only tracked file counting against the 200-file limit, the wording of one Risks line) are separate.
- `docs/rust-product-path.md`: tests read it (`go_state_classes_cli`, `ops_table_gen_cli`), so a change there needs its own slice; the line that still lists only two groups is a follow-up.
- `hooks-status` reading the new log: `security-scan` is not part of the status summary yet.

## Requirements

- REQ-1: WHEN the hook set is generated for the host claude, the hook set shall hold a Stop group whose command is `<binary> hook-dispatch --host claude --event security-turn` with the hook timeout of every other group. · failure: none
- REQ-2: The Stop group shall carry no matcher key.
- REQ-3: The Stop group shall stand last in the hooks object, after every group of the other events.
- REQ-4: The hook set of the hosts codex and omp shall hold no group that runs security-turn.
- REQ-5: Every other group of every hook set shall keep its place and its content, the Codex lane Stop group included.
- REQ-6: WHEN `hooks-plan` runs on a project whose Claude hooks are the untouched set of the release before this change, the plan shall be an update that adds the group and keeps every other key of the settings file. · failure: none
- REQ-7: WHEN `hooks-plan` runs on a project whose Claude hooks already hold the group, the plan shall be unchanged. · failure: none
- REQ-8 (REQ-6, REQ-7): IF the Claude hooks of the project were edited by hand, THEN `hooks-plan` shall report a conflict and leave the bytes of the settings file untouched.
- REQ-9: The legacy set with the old binary name `moai-cli` shall hold no Stop group.
- REQ-10: The capabilities answer for claude shall list the event security-turn with the host event Stop and its default mode observe.
- REQ-11: The capabilities answers for codex and omp shall list the event security-turn as unsupported.
- REQ-12: The documents `docs/hooks-generation.md` and `docs/claude-adapter-contract.md` shall name the security-scan group and the Stop group in the Claude hook set.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-5): the Claude hooks object holds the event Stop last with exactly one group; that group has no `matcher`, the exact command and the timeout of the other groups; every other event keeps its groups, their order and their content (SessionStart, PreToolUse twice, PostToolUse receipt then security-scan, then edit-diagnostics while `[hooks.edit_diagnostics]` is enabled) · verify: `cargo test --locked --offline --test security_turn_wire_cli claude`
- AC-2 (REQ-4, REQ-5): the hooks objects of codex and omp hold no group that runs security-turn and equal what they were in their other groups; the Codex object planned with the lane loop holds one Stop group, the lane one · verify: `cargo test --locked --offline --test security_turn_wire_cli other_hosts`
- AC-3 (REQ-6, REQ-7, REQ-8): on a temporary project, a Claude settings file holding the previous release's untouched MoAI set with its ownership receipt gets the action `update` from `hooks-plan` with the Stop group listed and every other settings key kept, and after `hooks-apply` a second `hooks-plan` says unchanged; a settings file whose hooks were edited by hand stays a conflict with its bytes untouched · verify: `cargo test --locked --offline --test security_turn_wire_cli plan`
- AC-4 (REQ-9): the legacy Claude set equals the pre-P4 fixture bytes and holds no Stop group, while the current set with the same binary name would · verify: `cargo test --locked --offline --test security_turn_wire_cli legacy` and `cargo test --locked --offline --test hooks_pre_p4_fixtures`
- AC-5 (REQ-10, REQ-11): `hooks-plan --host claude` lists security-turn under the capabilities with the event Stop and the default mode observe, and the same for codex and omp lists it with `supported: false`; the security-scan entry no longer says it is not generated into settings · verify: `cargo test --locked --offline --test security_turn_wire_cli capabilities`
- AC-6 (REQ-12): both documents name the two groups and the command of the Stop group · verify: `cargo test --locked --offline --test security_turn_wire_cli docs`
- AC-7 (REQ-1…REQ-12): every test file that reads the hook set, `hooks-plan` or a count of hook groups passes after its expectation is updated for the new group and for no other reason, the hooks-pre-p4 fixtures stay unchanged, and an end-to-end run of the generated Stop command against a temporary repository answers as the previous slice's tests say · verify: the test files named in the Result, run unfiltered, and `cargo test --locked --offline --test security_turn_wire_cli end_to_end`

## Scope

- write: `src/hooks.rs` (the handler, its command, `hook_set_with`, the legacy set that omits the group), `src/hook_engine.rs` (the capabilities entries and the two "not generated" texts), `docs/hooks-generation.md`, `docs/claude-adapter-contract.md`, `tests/security_turn_wire_cli.rs` (new), the neighbour tests whose expectation must follow the new group (named in the Result and added to the reservation one by one by `card scope`), `.moai/specs/SPEC-20261121-security-turn-wire/**`
- read-only: `src/hook_legacy.rs`, `src/claude.rs`, `src/go_hooks.rs`, `src/ops/go_hooks_plan.rs`, `src/project_update.rs`, `tests/fixtures/hooks-pre-p4/**`, `docs/rust-product-path.md`
- forbidden: `skills/**`, `schemas/**`, other projects' hook files, the real HOME and ledger, the Go repository

## Plan

Tests first (RED), then code, then neighbours and a grep of every reader of what changes, then mutants, then the Result with one line per AC. The worktree is `~/.moai/worktrees/moai-adk/turnwire` on a branch made from the head of the turn-scan branch (`3ef9856743`; that slice, the security-scan wiring slice and the security-scan slice must land first, in that order).

1. RED: `tests/security_turn_wire_cli.rs` with the fixture model of `tests/security_scan_wire_cli.rs` (temporary project and HOME, the real binary). Cases `claude`, `other_hosts`, `plan`, `legacy`, `capabilities`, `docs`, `end_to_end`.
2. GREEN: `Handler::SecurityTurn`, its arm in `command_with`, the Stop entry pushed last in `hook_set_with` for the current binary on claude, the capabilities entries (claude, codex, omp) and the two texts; the two documents.
3. Neighbours: a grep with a path argument of `tests/`, `tools/`, `docs/` and `.github/` for every reader of the hook set, of `hooks-plan` output, of a group count, of the string `Stop` next to the generated set and of the capabilities keys; then run each test file it finds unfiltered (the previous two slices each missed one reader: a tool, and a repo-wide scan).
4. Mutants (at most 5): the Stop group dropped, the group given a matcher, the group placed before PostToolUse, the group added for codex, the group added to the legacy set.

Research notes (read 2026-11-21 on the turn-scan head `3ef9856743`): `src/hooks.rs` (`HOOK_SET`, `hook_set_with` with its single bool for the current binary, `hooks_object_with`: events keep the insertion order because `serde_json` has `preserve_order`, so a pushed Stop entry is last; the edit-diagnostics group is pushed into the existing PostToolUse array and so does not move it; `command_with`; the doc comments that say `Stop` is not in the set), `src/hook_engine.rs` (`capabilities`, `security_scan_capabilities` with its text "not wired into generated settings in this release"), `src/go_hooks.rs:219` and `src/project_update.rs:1102` (the two other readers of `hooks_object("claude")`: both take the whole object, so the new group flows through them), the previous wiring SPEC (`SPEC-20261120-security-scan-wire`, its Result and the audit that found a tool reading the old log), the previous slice's audit (worst case of one Stop run: `rev-parse`, then `diff` and `ls-files` of 3 s each at most plus the scan, measured at about 1 s at the text budget, which is under the 10 s timeout of every group).

## Risks

- Every turn now ends with one more hook process. The engine bounds its own work (three git calls of 3 s at most each, 200 tracked and 200 untracked files, 8 MiB of text; the previous slice's audit measured about 1 s at the text budget); the host's 10 s timeout is the generated one, so the worst case of the engine stays under it, and a host that cancels the hook loses one log line, nothing else.
- A project that already runs the MoAI set sees `hooks-plan` propose an update; nothing is written without the user's confirmation, and a hand-edited set stays a conflict. A project that holds the previous release's set without the ownership record gets `conflict` with no replacement choice offered, as the previous wiring slice documented; the follow-up that decides the wording and a way out covers this group too.
- Whether the host shows a `systemMessage` at turn end is not verified in a real session (the Stop sections of the hooks page do not mention it). Observe, the default, does not depend on it; until a real session shows it, `enforce` may have no visible effect while the log line still records it. This slice does not claim otherwise in any document.
- A Claude project that already has its own `Stop` hook (a user's) is a hand-edited `hooks` object: `hooks-plan` reports a conflict and offers no merge; the MoAI writer owns the whole `hooks` key. This is the same rule as for every other event.
- Test files that count hook groups or events will fail until their expectation follows; each is updated for the new group only and named in the Result.
- The Codex and OMP sets stay without the event, so the turn-end scan covers Claude Code only; the capabilities answer says so.

## Result

(not yet written)
