---
id: SPEC-20261122-next-flow-hooks
title: "the NeXT-Flow hook set: session start without git, a widened tool preflight split into a fail-open and a blocking duty, and the Issue-era hooks removed"
status: draft
revision: 1
created: 2026-11-22
---

# the NeXT-Flow hook set: session start without git, a widened tool preflight split into a fail-open and a blocking duty, and the Issue-era hooks removed

## Intent

The Claude hook set still carries the Issue era. The claim-verify guard (`moai claim-verify . --hook` before Bash) keys on `issue/<n>` lane branches that no longer exist, so it almost never fires; the hook-receipt handler (`moai hook-receipt` after Bash) writes a log whose 4 MiB cap this project reached on 2026-11-20, so nothing has been recorded since. The session start (55 to 71 ms measured) starts `git` twice to name a lane branch and a claim record nobody acts on. Meanwhile the audit of the integration-branch guard measured that `git branch -C` moves the branch with no reference transaction at all, so a Claude-side backstop is wanted: from a Claude session, the calls that move, copy, rename, delete or push the integration branch shall be refused before they run.

This SPEC reshapes the set per section 4-1 of the hooks, rules and skills report (2026-11-22): the two Issue-era entries are removed from the generated settings; the session-check duty reads files only (the git HEAD file, the integration-branch record in the common git folder, the worktree's card from the local ledger) and answers within 20 ms; the tool preflight keeps its refusals, replaces the Issue claim-token rule with a read-only ledger reservation check, refuses writes and reference moves from read-only agent types, and is split into an ordinary fail-open duty and a dangerous-call duty that the generated settings carry with `onFailure: "block"` (Claude Code v2.1.295+, hooks.md), so a missing or crashing moai blocks only those calls. The same guarded commands are added to `permissions.deny`, and the update preview names the project its hooks part touches. The user picked the split (2026-11-22, recorded on the card) revising the earlier fail-open-everything pick for the dangerous part only; no SessionEnd hook.

## Non-goals

- Removing the Codex and OMP hook adapters and the lane Stop hook: the withdrawn-host cards own that work; this SPEC changes the shared base set only as far as the Claude set requires.
- Narrowing `.gitignore` so `.claude/settings.json` becomes tracked: the layout SPEC owns it. This SPEC only measures once that a worktree session receives the hooks once the file is tracked.
- The reference-transaction guard, the landing command and the drift check in card land: landed slices. This SPEC adds the Claude-side backstop of the same refusals only.
- The pre-push receipt and the test-selection work: the remote-copy and test-selection cards.
- A SessionEnd hook and WorktreeCreate, UserPromptSubmit or Stop hooks: the report's design and the user's pick keep them out; the edit-diagnostics opt-in entry stays as it is.
- Moving every HOME state file off the retired folder: the path-root card owns the whole move. Here only the engine log stops using it.

## Requirements

- REQ-1: When the hook writer renders the Claude hook set, the hook writer shall emit no claim-verify entry and no hook-receipt entry and set onFailure to block on the dangerous-call entry alone · failure: none — the planned settings hold neither removed command and no other generated entry carries an onFailure field, and an installed settings file that still holds a removed entry plans as an update that removes it, so a missing or crashing moai blocks only the calls the dangerous-call duty checks.
- REQ-2: When the session-check duty reports the state of the worktree it runs in, the session-check duty shall read the checked-out branch from the git HEAD file and the integration branch's name from the record in the common git folder and the worktree's card from the local ledger, each read only and with no subprocess started · failure: none — a detached HEAD, an unborn branch, an absent record, an absent ledger, a folder outside git and a poisoned PATH that fails every program each answer with the fact named in fixed wording and the answer unchanged.
- REQ-3: When the session-check duty answers, the session-check duty shall keep the conversation-language line and the after-compaction advice, point at `moai card show` for the worktree's card and carry no Issue, claim or reservation wording · failure: none — no Issue number, claim record or reservation word appears in any answer.
- REQ-4: When the session-check duty runs, the session-check duty shall answer within 20 milliseconds of its own work · failure: none — an in-process test pins the median of repeated runs under the budget and one measured host run is recorded in the Result.
- REQ-5: When the ordinary tool-preflight duty checks a call, the ordinary tool-preflight duty shall refuse the same calls the tool-preflight refuses at the base of this change · failure: none — the engine's existing tool-preflight tests pass unedited and no refusal is dropped.
- REQ-6: When the dangerous-call duty checks a Bash call whose branch -c, -C, -m, -M or -D, update-ref or push names the recorded integration branch in any letter case, behind git -C or with an absolute path, or a call that recursively deletes outside the project tree or writes a guarded settings file, the dangerous-call duty shall refuse the call · failure: none — the branch refusal names the recorded spelling and the landing command, the others carry a fixed reason, and the same verbs naming another branch, an ordinary push of a lane branch and an edit of a source file pass.
- REQ-7: When the tool-preflight duty checks a call whose agent type is read-only, the tool-preflight duty shall refuse a file write and a reference move and record the agent in the deny line · failure: none — the deny line names the agent type and a short hash of the agent id, never a raw id, and a read-only agent's read call passes.
- REQ-8: When the tool-preflight duty checks a commit in a project with a local ledger, the tool-preflight duty shall refuse the commit when a path it touches is covered by the live reservation of another holder, reading the ledger only · failure: none — a commit inside this session's own reservation passes, and the issue-branch claim-token rule and its wording are gone.
- REQ-9: When the engine writes its log, the engine shall write no line under the retired home folder · failure: none — a HOME whose state root still carries the retired name logs on the project side only, and reading still reads the retired place for lines written before.
- REQ-10: When the hook writer plans the Claude settings, the hook writer shall add the guarded commands to permissions.deny as narrowing rules beside the hook set · failure: none — the rules refuse a force push, reset --hard, a forced clean, --no-verify, add -A and the integration-branch calls of the dangerous-call duty, no rule widens access, and a permissions key the user changed stays a conflict.
- REQ-11: When the update preview reports its hooks part, the update preview shall name the project folder the hooks are installed into and every entry the change adds or removes · failure: none — a preview whose hook layer changes nothing says so and still names the folder.
- REQ-12: When a session starts in a .claude/worktrees worktree of a project whose .claude/settings.json is tracked, the generated hook set shall reach that session · failure: none — one measurement of a real worktree session is recorded in the Result naming what the session saw.

## Acceptance

- AC-1 (REQ-1): a fresh plan renders the SessionStart entry and the two PreToolUse entries and neither removed command string, a settings file that still holds a removed entry plans as an update whose diff removes it, and the applied settings and the ownership receipt name no removed entry · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac1_`
- AC-2 (REQ-2, REQ-3): a worktree with a ledger, a recorded integration branch and a language recorded in HOME answers with the branch, the recorded name, the card of the branch and the language line, points at `moai card show`, and holds no Issue, claim or reservation word; the answer is byte-identical under a poisoned PATH; a detached HEAD, an unborn branch, an absent record, an absent ledger and a plain folder each name the fact · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac2_`
- AC-3 (REQ-4): the duty called in process over a prepared worktree answers with a median under 20 ms across repeated runs, and the Result records one measured host run · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac3_`
- AC-4 (REQ-5): every tool-preflight test of the engine that passes at the base of this change passes at its head unedited · verify: `cargo test --locked --offline --lib hook_engine`
- AC-5 (REQ-6): branch -c, -C, -m, -M and -D, update-ref and push naming the recorded branch — in the recorded spelling, in another letter case, behind git -C and with an absolute path — are each denied with the recorded spelling and the landing command in the reason, and each verb naming another branch passes · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac5_`
- AC-6 (REQ-6): a recursive delete outside the project tree and a Write of each guarded settings file are denied with a fixed reason, an ordinary push of a lane branch passes, and an edit of a source file passes · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac6_`
- AC-7 (REQ-1): the rendered Claude settings carry `onFailure: "block"` on the dangerous-call entry alone, every other generated entry carries no onFailure field, and the entries stay valid against the documented schema of a command hook · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac7_`
- AC-8 (REQ-7): a write tool call and a reference-move Bash call whose agent_type is a read-only agent are denied, the deny line holds the agent type and a short hash of the agent id and no raw id, and the same agent's Read call passes · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac8_`
- AC-9 (REQ-8): a commit touching a path covered by another holder's live reservation is denied with the ledger read only, a commit inside this session's own reservation passes, and no claim-token wording remains in any answer or refusal · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac9_`
- AC-10 (REQ-9): with a HOME whose state root is the retired name a deny line lands on the project side and nothing is created under the retired folder, with a marked HOME it lands under the current root, and reading sees lines from both · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac10_`
- AC-11 (REQ-10): the planned settings carry the deny rules and no allow rule, the rules and the dangerous duty's verbs come from one table, and a permissions key the user changed stays a whole-plan conflict · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac11_`
- AC-12 (REQ-11): the preview's hooks summary names the project folder and each entry the change adds or removes, and a change-free hook layer prints its action and the folder · verify: `cargo test --locked --offline --test next_flow_hooks_cli ac12_`
- AC-13 (REQ-12): one real session started in a `.claude/worktrees` worktree of a project whose settings file is tracked receives the generated hook set, measured once with the method named in the Result · verify: measured once · recorded in the Result

## Scope

- write: `.moai/specs/SPEC-20261122-next-flow-hooks/spec.md`
- write: `src/hooks.rs`, `src/hook_engine.rs`, `src/hook_guards.rs`, `src/claude.rs`, `src/update_hooks.rs`
- write: `tests/next_flow_hooks_cli.rs`
- write: the help texts and goldens the changed hook examples move (`src/help.rs`, `tests/help_layout.rs`) only as far as those examples change.

## Decisions

- D-1: The split of the preflight is two duties on one event: `tool-preflight` keeps the ordinary refusals and is generated without `onFailure`, and a new duty `dangerous-preflight` carries the dangerous refusals and is generated with `onFailure: "block"` as its own matcher group. Both run in parallel per the host docs and the most restrictive answer wins, so a deny from either blocks the call.
- D-2: The recorded integration branch is read through the one reader of the reference-guard module (the record in the common git folder, the profile value only when no record exists). A repository with neither has no integration-branch refusals: there is no name to protect.
- D-3: The read-only agent types are the base agents whose definitions disallow the write tools (moai-plan, moai-research, moai-audit today); the list lives beside the agent definitions and changes with the agent-set card, not here.
- D-4: The guarded settings files are the project's `.claude/settings.json` and `.claude/settings.local.json`, the two files the settings ownership already guards.
- D-5: The agent is recorded in the deny line as its agent type and the first 16 hex digits of the sha256 of its agent id, the same treatment a session id gets; a raw host id never enters the log.
- D-6: The ledger reservation check replaces the Issue claim-token rule: it reads the reservation rows with a live lease and refuses only when the covering reservation's holder is not the session the call names. Merges by the landing command are not `git commit` tool calls and never meet it.
- D-7: The 20 ms budget covers the duty's own work in process; the host's process start (about 7 ms measured in the report) is outside it and is covered by the one measured host run in the Result.
- D-8: The `permissions.deny` rules and the dangerous duty's verbs come from one table in the code, so the two layers cannot drift; the rules use the host's permission-rule syntax and are narrowing only.
- D-9: On a HOME whose state root still resolves to the retired name the engine log stays on the project side and never creates the retired folder; the read path keeps both places for lines written before the move. The full state move is the path-root card's.
- D-10: The session-check text keeps the after-compaction advice (instructions may already be done) but points at `moai card show` instead of the Issue-era resume comment; the language line keeps its one source in HOME.
- D-11: The ordinary checks stay fail-open when moai is missing or crashes (the user's first pick, unchanged for them); only the dangerous duty blocks on its own failure, per the user's revised pick.

## Plan

1. Tests first (RED): `tests/next_flow_hooks_cli.rs` for AC-1 to AC-3 and AC-5 to AC-12, and the engine unit tests AC-4 keeps green; the worktree measurement of AC-13 is prepared.
2. `src/hooks.rs`: the set without the removed entries, the dangerous entry with `onFailure`, the deny rules of one table beside it; `src/claude.rs`: the desired permissions.
3. `src/hook_engine.rs` and `src/hook_guards.rs`: the session-check rewrite over files, the dangerous duty, the read-only agent refusals, the ledger reservation check, the log-side rule.
4. `src/update_hooks.rs`: the summary line that names the folder and the added and removed entries; the help texts the change moves.
5. Light checks, at most five mutants after the commit, one affected run, the AC-13 measurement recorded in the Result, then the READY note with head and base.

## Risks

- A `permissions.deny` rule written too narrowly can miss a form of a guarded command (the host's Bash rule matching is best-effort). The dangerous duty is the second layer on the same table (D-8), and a test derives both from it.
- The in-process budget test measures the duty alone; the real per-call cost adds one process start. The measured-once number in the Result is the honest end-to-end figure, as the report's own measurements were.
- The read-only agent list is fixed: a user's own read-only agent type is not covered until the agent-set card lands; the refusal names the type, so the gap is visible.
- The ledger reservation check can refuse a commit by the leader inside the primary when a path is reserved by another holder; the leader lands through the landing command, whose merges are not tool commits, and a genuine conflict is exactly what the check is for.
- The worktree measurement depends on the settings file being tracked, which is the layout SPEC's change; until it lands the measurement is taken with the file copied by hand and says so in the Result.
- Removing the receipt handler stops the `hooks.jsonl` evidence the hooks-status field `fired` reads; that field reads the engine log too and keeps working, and the removed receipt log is left where it is (the archive work owns it).

## Result

Not started.
