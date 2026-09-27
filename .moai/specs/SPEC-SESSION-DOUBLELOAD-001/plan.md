---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Implementation plan — card sessions start inside their worktree"
version: "0.2.0"
---

# Plan — SPEC-SESSION-DOUBLELOAD-001

Milestones are ordered by decision reversibility. The measurement that can overturn the premises comes first, the two-tier decision second, then a gate, then the mechanical doctrine edits. No time estimates; priority labels only.

## §A Context

- Source of truth for premises: `.moai/reports/t1219/verdict.md` (Evidence 1–4). Audit record: `.moai/reports/t1219/plan-audit.md` (iter-1 FAIL 0.64).
- Tree: `.claude/worktrees/t1219`, branch `WT-local-doc-dedup`. Probes and edits happen here; nothing is edited in the primary checkout.
- Tier M (spec, plan, acceptance, progress). No Go code changes.
- Operator decisions: spec.md §F.

## §B Known Issues

- Headless `claude -p` and an interactive session may attach nested instruction files at different moments (verdict § Residual-risk). Every result names the mode it came from.
- Primary-start probes (C, control, D′) run the primary's SessionStart hooks. The primary `.claude/settings.local.json` wires the LSEL `session_drain.sh` and `backlog_check.sh` there; the worktree's local settings file does not. Handled in §D.
- The probing shell may inherit `MOAI_KANBAN*` variables from a lane environment. Handled in §D.
- `rule-load-audit.jsonl` is shared by every session that writes to the same project root. Rows are filtered by the probe's `session_id`.
- Whether InstructionsLoaded fires for a subagent, and with which identity, is unmeasured (paths D, D′).
- t1175 edits T, L and A. Doctrine edits wait for it (§F Gate G1).

## §C Pre-flight

1. `git -C <tree> rev-parse --show-toplevel` → the t1219 worktree path; `git -C <tree> status --short` → empty.
2. `claude --version` recorded into E.
3. `claude --help | grep -c -- '--setting-sources'` → `1`. This flag was observed this run; the check is repeated at M1.
4. `jq '.hooks.SessionStart' <primary>/.claude/settings.local.json` recorded, to confirm that the LSEL entries live only in the local settings layer.
5. `git -C <primary> branch --show-current` recorded as `primary_branch_before:`.
6. `grep -c '^\[HARD\]' T` recorded as `hard_clauses_before:` (observed `32` at `2fdd1f8c9`, per plan-audit baseline table).

## §D Constraints — probe isolation

- **Caps.** At most 8 probes, `--max-turns 4` at most, each under `timeout 300`, aggregate wall clock at most 45 minutes. They are written to `.moai/reports/t1219/m1-caps.md` and committed alone, before any probe runs (REQ-SDL-003). A reached cap makes the unobserved item a Gap.
- **One compound invocation per probe:**

  ```bash
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 300 claude -p "<probe prompt>" --max-turns <n> --setting-sources user,project --output-format stream-json --verbose --debug-file .moai/reports/t1219/probes/<X>.debug > .moai/reports/t1219/probes/<X>.jsonl
  ```

  Every probe command line is appended verbatim to `.moai/reports/t1219/probes/commands.txt`.
- **LSEL drain isolation.** `--setting-sources user,project` excludes the local settings layer that carries the LSEL SessionStart hooks, and every probe uses it so all paths are comparable.
  - Before and after the primary-start probes, the listing of `<primary>/.moai/state/lsel/` is recorded.
  - If it changed, the flag did not isolate the drain: E records `lsel_side_effect: declared` with both listings.
  - Otherwise E records `lsel_side_effect: excluded-by-setting-sources`.
  - Residual risk: excluding the local settings layer is assumed not to change which memory files load. The control probe compares against verdict Evidence 1.
- **Process cleanup.** Each probe's process-group id is recorded as `pgid=<n>` in E. After M1, `pgrep -g <n>` returns nothing for every recorded group. No process is killed by name.
- **Instruction-path source.**
  - Rows of `.moai/logs/rule-load-audit.jsonl`, under both the primary root and the t1219 root, filtered by the probe's `session_id`. The `session_id` is taken from the stream-json `system` init event.
  - Subagent rows are identified by whatever agent field the rows carry.
  - If no row identifies the subagent, the subagent prints its own instruction-header paths. E then marks that path `source: self-report`.
- **Skill-source extraction.** The `Loading skills from` lines of the debug file, plus any skill-listing attachment in the stream-json. The extraction command is recorded in E as `extractor_skills:`.
- **Token figure (reference only).** Take the `usage` of the first assistant message after entry: the first assistant message for A, B and D; the first one after the `EnterWorktree` tool result for C and D′. Record `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`. No decision reads it.
- **Other worktrees.** Probes write only under the t1219 root and the primary root. This is verified by session-id attribution (acceptance AC-SDL-006), not by status comparison, because other lanes are live.
- **Declared primary-side effects.** Primary-start probes register in the session registry and append to the primary `rule-load-audit.jsonl`. Both are runtime state, not tracked files.
- Probe outputs go to `.moai/reports/t1219/probes/`, never `/tmp`.

## §E Probe Matrix (7 of 8 allowed)

| # | Path | Start cwd | Prompt / action | Turns |
|---|---|---|---|---|
| 1 | B | t1219 worktree | "reply ok" | 1 |
| 2 | control | primary | "reply ok" (no move) | 1 |
| 3 | C | primary | `EnterWorktree` into t1219, then `Read` the worktree CLAUDE.local.md | ≤4 |
| 4 | D | t1219 worktree | spawn one read-only subagent that prints the paths of its instruction headers | ≤4 |
| 5 | D′ | primary | `EnterWorktree` into t1219, then the same subagent spawn as #4 | ≤4 |
| 6 | A | t1219 via `moai cc -w t1219` | one interactive turn, operator-assisted, bounded by `timeout` | 1 |
| 7 | /clear | primary → `EnterWorktree` → `/clear` | operator performs the `/clear` once in an interactive session; the loaded set is then read by the same extractors | ≤2 |

- The one spare probe is reserved for a single retry of a probe that timed out.
- If #6 cannot be run within the caps: `skills_paths_A: gap`, `instruction_paths_A: gap`, and `path_a_proxy: B` is declared.
- If #7 cannot be run: `clear_restores_single_load: gap` and `clear_method: none`.

## §F Milestones

### M1 — Measure (Priority High, runs now)

Decision this milestone can overturn: whether each class (instruction files, skills) is duplicated on each path, and whether `/clear` is a usable lane standard.

1. Pre-flight (§C). Write and commit `m1-caps.md` alone. The commit message contains t1219.
2. Run probes #1–#7 per §D and §E.
3. Positive control (REQ-SDL-007): run `extractor_skills` and the instruction extractor on an input known to hold two sources. Use probe #3's output if it holds two. Otherwise use a fixture `probes/fixture-two-sources.txt`, built from a real debug line plus a second path. The control must print `count=2`, and E names the file.
4. Write `.moai/reports/t1219/m1-measure.md` (E) containing:
   - the machine-readable lines listed in acceptance.md;
   - the five-section report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
5. Commit E and `probes/` (`docs(t1219): …`).

### M2 — Decide the two tiers (Priority High, runs now)

Apply REQ-SDL-008/009/010 to E and write § Decision into E:

| Evidence | Decision field |
|---|---|
| A (or B as declared proxy) shows only worktree paths for both classes | `decision_primary: launcher` |
| `clear_restores_single_load: yes` | `decision_lane_standard: move-clear-resend` |
| `clear_restores_single_load: no` or `gap` | `decision_lane_standard: relaunch` |
| `premise_instructions_C: single` and `premise_skills_C: single` | `decision_primary: none`, `decision_lane_standard: none`; M3/M4 do not run; blocker to the orchestrator |

Paths D and D′ are recorded. If D shows primary paths (a subagent does not follow its parent's worktree start), E states that the two tiers do not close the subagent case. This goes to the orchestrator as a finding; it is not a doctrine edit here.

### Gate G1 — t1175 on develop (REQ-SDL-016)

M3 and M4 do not start until both hold:
1. `git log --format=%H -1 --grep='t1175' develop` is non-empty.
2. That commit is an ancestor of HEAD after absorbing local develop.

After absorbing, re-measure `hard_clauses_before` and the anchors of acceptance AC-SDL-014 on the absorbed tree.

### M3 — Template edits, then mirrors (Priority Medium, after G1)

- T (template kanban dispatch rule): state the two tiers in the `wt` bullet, the "A new card starts in a new worktree" clause, and the "The `/clear` handoff between phases" clause.
  - Where the lane standard applies, the new-card clause spells the sequence `ExitWorktree` → `EnterWorktree(<card-id>)` → `/clear`, and the handoff clause orders a card change as move → `/clear` → re-send pointer. This moves the lane's existing `/clear`; it adds none.
  - Where relaunch applies, both clauses say to end the lane session and relaunch it with the launcher.
- DT (template detail companion): the dispatch example `wt: EnterWorktree(t0)` becomes the primary entry form.
- AT (`AGENTS.md.tmpl`): the "Start a new card in a new worktree" paragraph names launcher start at session start, and the lane standard as the in-session alternative. One or two sentences; this file is always-loaded and budget-limited.
- Generic placeholders only (`<name>`, `<card-id>`).
- `make build`, then copy T→L and DT→DL, and apply the same paragraph to A.

### M4 — Local lane protocol (Priority Medium, after G1)

G §1: the entry bullet names launcher start for a fresh card session and the lane standard (or relaunch) for a long-lived lane. It drops `EnterWorktree(<card-id>)` as an equal alternative. No template mirror.

### M5 — Verification (Priority Low, mechanical)

Run acceptance.md checks and record their outputs in `progress.md` §E.2 (run-phase owner).

## §G Anti-Patterns

- Editing doctrine before E is committed, or before G1.
- Reporting "no duplication" from a zero count without the two-source positive control.
- Running a probe without the `unset … &&` prefix in the same invocation, or with a trailing `kill` instead of `timeout`.
- Killing processes by name.
- Putting the card id or this SPEC id into template lines.
- Running `go test ./...` on the full suite.

## §H Cross-References and Resolved Questions

- `.moai/reports/t1219/verdict.md` — measured premises (Evidence 1–4).
- `.moai/reports/t1219/plan-audit.md` — iter-1 defects D1–D20.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Dispatch format, § The `/clear` handoff between phases, § Isolation is provisioned by MoAI, then entered through a launcher.
- `.claude/rules/local/gitflow-lane-protocol.md` §1 — card-worktree entry.
- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern — already uses the launcher form for new-session entry.

Resolved by operator decision (2026-09-27), formerly open:
- Lane lifecycle → two tiers, with the lane standard conditional on M1 (spec.md §F).
- Ordering against t1175 → M1/M2 now; M3/M4 behind Gate G1.
- Probe tree → the t1219 worktree, under the §D isolation conditions.
- `/clear` measurement method → operator-performed once (probe #7). If that is not possible within the caps, the result is a Gap and lanes relaunch.
