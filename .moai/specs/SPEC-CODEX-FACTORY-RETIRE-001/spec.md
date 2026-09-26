---
id: SPEC-CODEX-FACTORY-RETIRE-001
title: "Retire the interactive-TUI codex factory path (moai codex -k / -f)"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, factory, kanban, retirement, lane-env, handoff, launcher"
tier: L
card: t1242
related_specs: [SPEC-DUAL-HARNESS-RECOVERY-001, SPEC-FACTORY-MIXED-HOOK-001, SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001, SPEC-FACTORY-RUN-RETIRE-001, SPEC-CODEX-LOCALMD-001]
---

# SPEC-CODEX-FACTORY-RETIRE-001 — retire the interactive-TUI codex factory path

## HISTORY

- 2026-09-26 · v0.1.0 · manager-spec · Initial authoring from card t1242 (class C),
  "SPEC A" of the lead's two-SPEC design (retirement only; the headless
  `codex exec` worker rebuild is a later card, F2). Every code coordinate below was
  re-measured on this tree (`553e224f3`, branch `WT-codex-factory-retire`); the
  design's coordinates came from develop `35ab8cff3`. Divergences are recorded in
  `research.md` §R2 and summarized in §A.3.
- 2026-09-26 · v0.2.0 · manager-spec · plan-audit iter-1 FAIL (0.71) revision:
  build-tag statement (REQ-016), full-source schema guard (REQ-015), codex-hook peer
  refusal (REQ-022), M5 no-confirmation prohibition (REQ-023), posture-key rationale
  (REQ-009), develop-parent budget baseline (REQ-021), pinned absence evidence,
  scoped design guarantee plus a known-limitation exclusion, extended supersession
  list; ACs revised per `.moai/reports/plan-audit/SPEC-CODEX-FACTORY-RETIRE-001-review-1.md`.

## §0 Governing principle [HARD]

This SPEC removes a launch surface; it does not rebuild one. Every requirement
either (a) refuses the retired entry, (b) closes a leak the retired entry made
possible, (c) deletes code that only the retired entry reached, or (d) keeps a
shared surface intact. A change that belongs to none of the four is out of scope
(§E), however adjacent it looks.

## §A Background

### A.1 What exists today (measured on `553e224f3`)

- `moai codex -k` (Kanban Mode entry, `internal/cli/codex_kanban.go`, 122 lines) and
  `moai codex -f` / `-f worker[-<n>]` / `--factory-run` (Factory Mode entry,
  `internal/cli/codex_factory.go`, 147 lines) are intercepted in `runCodex`
  (`internal/cli/codex_launcher.go:689`, factory/kanban handling at `:701-744` and
  `:762-770`).
- The codex direct launch registers a factory launch-pending peer whenever the
  inherited environment carries `MOAI_KANBAN_ID` plus `MOAI_FACTORY_WORKER(S)`
  (`codex_direct_posix.go:39`, `codex_direct_windows.go:28-49`); the spawn path
  does the same and also stamps/clears the run owner (`codex_launcher.go:227-268`)
  and forwards nine `MOAI_KANBAN_*` / `MOAI_FACTORY_*` keys onto the tmux command
  line (`codex_launcher.go:296-310`).
- The direct child environment (`codexChildEnv`, `codex_launcher.go:613`) removes
  only `CLAUDE_CODE_SESSION_ID` and `MOAI_SESSION_PID`. A bare `moai codex` run
  inside a Claude factory lane therefore inherits that lane's factory identity and
  registers itself as a factory peer of the lane's run.
- The lane worktree handoff CLI of card t1082 (`factory_lane_handoff.go`,
  `_switch.go`, `_bind.go`, `_recover.go` — 747 production lines) plus the codex
  app-server thread-relocation client (`mcp_codex.go:824-906`) have **no production
  caller**: every call site of `prepareLaneHandoff`, `switchLaneHandoff*`,
  `bindLaneHandoffHeadless`, `recoverLaneHandoff` is a `_test.go` file
  (`research.md` §R3).

### A.2 Why retire before rebuilding

The interactive-TUI lane depends on a human typing `/cd` into a Codex prompt and on
an app-server relocation handshake that was never wired to a command. The rebuild
(F2) targets a headless `codex exec` worker instead. Leaving the old entry live
while F2 lands would give two codex factory doors with different semantics sharing
one `factory.db`. Retiring first leaves F2 a clean refusal to replace.

### A.3 Where the measured code contradicts the design

1. **`codex queue --thread` wake is not in this tree.** Pinned to `553e224f3` and
   scoped to code (`.moai/specs` excluded, because this SPEC's own body names the
   symbol): `git grep -c factoryQueueCodexMessage 553e224f3 -- internal cmd pkg` and
   `git grep -c 'to.Backend == "codex"' 553e224f3 -- internal` both exit 1 (no match). The symbol exists only in the uncommitted develop-worktree
   copy the lead reported. The REMOVE item reduces to the foreign-file disposition
   (REQ-CFR-019).
2. **The handoff CLI is already unreachable.** Removing it withdraws no
   operator-visible command; `moai factory handoff abandon-lane` and
   `recover-resume` are backend-neutral and stay.
3. **The "Codex lanes" wording in `AGENTS.md` sits in the §0 capability table (line
   26), not §3.** `internal/template/templates/AGENTS.md` does not exist; the mirror
   is `AGENTS.md.tmpl`, whose §8 already carries no `-f lead/agents` wording (it
   diverges from root `AGENTS.md` by design).
4. **`manager-lead` Role B carries no codex-factory wording** in any of the three
   copies (C1/C2/C3); it names only `moai cc -k` / `moai cc -f`. No agent-file edit
   is needed, so `make agents-emit` is not triggered by this SPEC.
5. **`docs-site/.../codex-dual-harness.md` carries the `-f lead/agents` launch-shape
   phrase at line 33 in all four locales**; `mcp-server.md` carries "Codex lane
   orchestrator" in the `codex_role_audit*` rows (lines 180-184) in all four locales.
6. **Tier is L, not M** (§A.4).

### A.4 Tier L justification

Measured scope: 1,016 production lines deleted across six files, about 3,000 test
lines deleted or rewritten across more than twelve test files, edits to seven
further production files, eight docs-site files, `AGENTS.md`, and a rule pair —
over 30 files in total. That exceeds both Tier M bounds (1,000 LOC, 15 files). More
than 80% of the LOC is deletion, which is why the design proposed M; the tier is
set by the file count and the cross-package blast radius (cli, factorymsg, hook,
config, docs), not by net growth.

## §B Requirements (GEARS)

### B.1 Entry refusal (Milestone M1)

- **REQ-CFR-001** — When `moai codex` receives a kanban entry token in the verb-position
  head (`-k`, `--kanban`, their `=` forms, with or without a SPEC value or a
  `--name`/`-n` pair), the launcher shall print one dedicated refusal line on stderr
  that carries the sentinel `KANBAN_MODE_UNSUPPORTED_BACKEND` and names
  `moai cc -k`, exit with code 1, and start no child process.
- **REQ-CFR-002** — When `moai codex` receives a factory entry token in the verb-position
  head (`-f`, `--factory`, their `=` forms, any role or worker-label value, or
  `--factory-run`), the launcher shall print one dedicated refusal line on stderr that
  carries the sentinel `FACTORY_MODE_UNSUPPORTED_BACKEND` and names `moai cc -f` and
  `moai glm -f`, exit with code 1, and start no child process.
- **REQ-CFR-003** — When either refusal fires, the launcher shall not have written a
  factory run row, claimed a worker-registry slot, exported a kanban or factory
  environment variable, or created a worktree.
- **REQ-CFR-004** — When an entry token appears after `--`, the launcher shall treat it
  as a codex passthrough argument and not refuse.
- **REQ-CFR-005** — The refusals shall fire regardless of the verb (`cli`, `app`,
  `status`, none) and regardless of `-w` or `--spawn` in the same head.

### B.2 Lane-environment defense (Milestone M1)

- **REQ-CFR-006** — When the launcher builds the direct-launch child environment, the
  child shall carry no non-empty value for any lane launch key: `MOAI_KANBAN`,
  `MOAI_KANBAN_ID`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL`,
  `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_KANBAN_BACKEND`,
  `MOAI_KANBAN_CARD`, `MOAI_KANBAN_SETTINGS_INJECTED`, `MOAI_FACTORY_WORKER`,
  `MOAI_FACTORY_WORKERS` (hereafter "the eleven lane keys").
- **REQ-CFR-007** — When the launcher builds the spawn command, the command shall assign
  every one of the eleven lane keys an empty value and shall forward none of their
  inherited values, so neither this process nor the tmux server environment can leak
  a lane identity into the pane.
- **REQ-CFR-008** — While a codex launch runs under an environment carrying a lane
  identity, the launcher shall not register a factory launch-pending peer, stamp a
  run owner, or clear a run owner.
- **REQ-CFR-009** — The environment defense shall preserve every other inherited
  variable, the resolved `CODEX_HOME`, and the existing removal of
  `CLAUDE_CODE_SESSION_ID` and `MOAI_SESSION_PID`. The posture keys
  `MOAI_AUTONOMY_TIER` and `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` are deliberately
  preserved: neither carries a factory identity (the launch gate
  `factoryLaunchEnabled` and the hook peer registration read only identity keys), and
  an operator-set autonomy tier is an explicit choice this SPEC does not override.

### B.3 Codex-run join refusal (Milestone M1)

- **REQ-CFR-010** — When a `moai cc` or `moai glm` factory entry (worker join, or a lead
  with `--factory-run`) selects an active factory run whose recorded lead backend is
  `codex`, the entry shall refuse before any registry claim or run-row write, exit
  non-zero, and name the run id and `moai factory runs --retire <run-id>` in its
  message.
- **REQ-CFR-011** — When the selected run's lead backend is anything other than
  `codex`, the join and lead paths shall behave exactly as before this SPEC.

### B.4 Dead code and shared-core preservation (Milestones M2, M3)

- **REQ-CFR-012** — The production tree shall contain no codex kanban or codex factory
  entry code: no `codex_kanban.go`, no `codex_factory.go`, and no reference from the
  codex launch files to the factory run-owner, launch-pending, or run-selection
  helpers.
- **REQ-CFR-013** — The production tree shall contain no lane-handoff preparation,
  switch, headless-bind, or recovery CLI code and no codex thread-relocation client.
- **REQ-CFR-014** — The shared factory core shall stay intact: `factory.go`,
  `factory_launch_pending.go`, `factory_run_owner.go`, `factorymsg` (store, handoff
  API, schema), the `factory_msg_*` MCP tools, the hook factory-message and
  interactive-bind path, and `moai factory handoff abandon-lane` / `recover-resume` /
  `moai factory runs`.
- **REQ-CFR-015** — The factory state schema shall not change: the production sources
  of `internal/factorymsg` and `internal/homestate` — every `CREATE TABLE` body,
  `ALTER TABLE`, and index statement, including `runs.lead_backend` — shall be
  byte-identical before and after.
- **REQ-CFR-016** — The kept codex surfaces shall behave unchanged: bare `moai codex`,
  `cli`, `status`, `app`, `--spawn`, `-w` (including its anchor lock and the POSIX
  in-place `syscall.Exec`), `--` passthrough, `moai codex audit`, `codex_role_audit*`,
  `codex_audit`, `codex_task`, and `session_msg_*`. `syscall.Exec` stays in
  `codex_direct_posix.go` under `//go:build !windows`, with `codex_direct_windows.go`
  under `//go:build windows`; this SPEC adds no new `syscall` use and keeps that
  build-tag split.
- **REQ-CFR-017** — Tests that exercise shared code shall survive the deletion of their
  host files, running under the same names in a kept file.

### B.5 Documentation, generated config, and budget (Milestone M4)

- **REQ-CFR-018** — No user-facing surface shall advertise `moai codex -k`,
  `moai codex -f`, a `-f` lead/agents launch shape for `moai codex`, or a "Codex lane"
  role: root `AGENTS.md`, `AGENTS.md.tmpl`, the `moai-mcp-tools*.md` rule pair (local
  and template), and the docs-site `codex-dual-harness.md` and `mcp-server.md` pages in
  all four locales (ko canonical, then en/ja/zh in the same change).
- **REQ-CFR-019** — When the lead confirms the merge-window disposition, the foreign
  uncommitted files in the develop worktree shall be preserved as a patch at
  `.moai/reports/t1242/foreign-6.patch` in the **primary checkout** (not the develop
  worktree, not the card worktree) before any of them is reverted or removed.
- **REQ-CFR-020** — The codex wiring generator shall not alter the `env_vars`
  allowlist of the generated `[mcp_servers.moai]` table, so existing projects do not
  start reporting `.codex/config.toml` drift.
- **REQ-CFR-021** — The always-loaded instruction surface shall not grow: the change
  shall leave `TestAlwaysLoadedTokenBudget` passing on the merge tree with a measured
  surface no larger than that of the merge commit's develop parent, both measured in
  the same run.

### B.6 Codex hook peer refusal and merge-window gate (Milestones M1, M5)

- **REQ-CFR-022** — When a MoAI hook runs under `--harness codex`, the hook shall not
  register, bind, or rotate a factory peer, whatever lane keys its environment carries;
  the same hook without `--harness codex` shall keep registering exactly as before.
- **REQ-CFR-023** — The run lane shall not modify, revert, move, or remove any path in
  the develop worktree before the lead's confirmation of the M5 disposition is recorded
  in `progress.md` §E.2.

## §C Decisions recorded here (not deferred)

- **D1 — `env_vars` stays.** Changing `mcpServerEnvVarsValue` makes every existing
  project's `[mcp_servers.moai]` table non-canonical; the writer never repairs a
  user-owned table, so doctor would report drift forever. With REQ-CFR-006/007 the
  keys arrive empty in a codex child and forwarding them is inert. F2's headless
  worker is expected to need factory attribution through the same allowlist.
- **D2 — Refusals reuse the existing sentinels** (`KANBAN_MODE_UNSUPPORTED_BACKEND`,
  `FACTORY_MODE_UNSUPPORTED_BACKEND`) so the codex refusal is greppable the same way
  as the `moai cg` refusal.
- **D3 — The hook interactive-bind path and the handoff Store API stay.** They sit on
  the Claude hook's per-prompt path and the tables stay; with no producer left they
  are inert. Removing them is a separate behavior change to a Claude path.
- **D4 — Fixture values:** a test fixture that uses `"codex"` only as an opaque backend
  string keeps it; a test whose path crosses REQ-CFR-010 or whose subject is a removed
  codex path switches to `claude`/`glm` or is removed with its subject.

## §D Acceptance criteria

Enumerated in `acceptance.md` (AC-CFR-001 … AC-CFR-025).

## §E Exclusions

### Out of Scope — the rebuild

- The headless `codex exec` factory worker (`moai codex -f agent` rebuild, card F2).
- A queue `assign` field or any dispatch-model change.

### Out of Scope — adjacent migrations and queue acts

- `AGENTS.local.md` migration (SPEC B).
- Dropping cards t1075, t1145, t1193, t1172 — operator queue acts, not code.
- Removing the handoff Store API, the hook interactive-bind path, or any factory.db
  table (D3).

### Out of Scope — pre-existing drift noticed during research

- `AGENTS.md` line 26 states `moai codex -w` "never creates" a tree, while
  `resolveOrCreateCodexWorktreeDir` creates one. This SPEC only removes the word
  "lanes" from that row; the create/no-create wording is reported, not fixed.
- Completed SPECs that describe `moai codex -k`/`-f` or the codex `-f` local-instruction
  shapes (SPEC-FACTORY-MIXED-HOOK-001, SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001,
  SPEC-DUAL-HARNESS-RECOVERY-001, SPEC-CODEX-LOCALMD-001, SPEC-FACTORY-RUN-RETIRE-001)
  are not rewritten; their supersession marking and the codemaps regeneration
  (`.moai/project/codemaps/modules.md:71` describes the `moai codex -f` entry) belong
  to the sync phase.

### Out of Scope — codex processes not launched through `moai codex` (known limitation)

- A `codex` binary started directly in a Claude lane terminal (not through the
  `moai codex` launcher) inherits the lane environment. REQ-CFR-006/007 do not reach it.
  REQ-CFR-022 closes its hook path (`--harness codex` hooks register no peer), but its
  `moai mcp-server` still receives the lane keys through `env_vars` (D1). Whether the
  `factory_msg_*` tools then attribute to the lane is unmeasured; they require a
  registered peer matching the caller's owner process, which this SPEC leaves absent.
  Follow-up card suggested: measure and, if needed, refuse lane attribution in the MCP
  server for a codex-owned caller.
