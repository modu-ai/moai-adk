---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Implementation plan — card sessions start inside their worktree"
version: "0.1.0"
---

# Plan — SPEC-SESSION-DOUBLELOAD-001

Milestones are ordered by decision reversibility: the measurement that can overturn the premises comes first, the entry-form decision second, and the mechanical doctrine edits last. No time estimates; priority labels only.

## §A Context

- Source of truth for premises: `.moai/reports/t1219/verdict.md`.
- Tree: `.claude/worktrees/t1219`, branch `WT-local-doc-dedup`. All probes and edits happen here; nothing is edited in the primary checkout.
- Tier M (4 artifacts). No Go code changes.

## §B Known Issues

- Headless `claude -p` and an interactive session may attach nested instruction files at different moments (verdict § Residual-risk). The measurement must say which mode each result came from.
- The kanban dispatch rule is always-loaded and is also touched by the rules-diet work in flight (t1175). Editing it here can conflict with that branch.
- Probes that start a session inside a worktree fire this project's hooks, which write under `.moai/state/`. That is state, not a tracked-file write, but it happens in a tree another writer may own.

## §C Pre-flight

1. `git -C <tree> rev-parse --show-toplevel` → the t1219 worktree path.
2. `git -C <tree> status --short` → clean before M1 starts.
3. `claude --version` recorded into the evidence file (loader behavior is version-dependent).

## §D Constraints

- Caps (REQ-SDL-003) are written into the evidence file before the first probe: ≤ 8 probe sessions, ≤ 4 turns each (`--max-turns 4`), each wrapped in `timeout 300`, aggregate wall clock ≤ 45 minutes. A reached cap turns the unobserved item into a Gap (REQ-SDL-004).
- Every probe runs as one foreground invocation bounded by `timeout`; no background process.
- Probe outputs go to `.moai/reports/t1219/probes/`, never `/tmp`.
- Template neutrality: the template copy carries no card id, SPEC id, date, or commit hash.

## §F Milestones

### M1 — Measure the unmeasured premises (Priority High, first)

Decision this milestone can overturn: whether the double load is real on each class (instruction files, skill directories) and whether `/clear` is a usable fallback.

1. Write `.moai/reports/t1219/m1-measure.md` with the caps block, `claude --version`, and a baseline line `hard_clauses_before: <n>` from `grep -c '^\[HARD\]'` on the kanban dispatch rule.
2. Path B (headless, started in worktree): `timeout 300 claude -p "<fixed probe prompt>" --max-turns 1 --output-format json --debug-file probes/B.debug` from the worktree cwd. Record project skill dirs (`Loading skills from` line), instruction-file paths (InstructionsLoaded hook output), and prompt tokens (`usage.input_tokens + cache_creation_input_tokens + cache_read_input_tokens`).
3. Path C (headless, started in primary, moves in): same form from the primary cwd, prompt instructs `EnterWorktree` into the t1219 tree then `Read` of the worktree CLAUDE.local.md; `--max-turns 4`. Record the same three fields, with every skills-dir line and every instruction path seen.
4. Positive control: a primary-start probe with no move; its `project=[…]` must name the primary path — proves the extraction query sees the field.
5. Path A (launcher): observe `moai cc -w <probe-name>` start once. If the launcher cannot be driven headless within the caps, record Path A as a Gap and use Path B as the nearest measured proxy — say so explicitly.
6. `/clear` after move (REQ-SDL-005): attempt in an interactive session within the caps; record `clear_restores_single_load: yes | no | gap`.
7. Emit the verdict fields in machine-readable lines (see acceptance.md AC-SDL-002..005) plus the five-section report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
8. Commit M1 evidence alone (`docs(t1219): …`).

### M2 — Decide primary and fallback entry forms (Priority High)

- Apply REQ-SDL-008/009/010 to M1's recorded fields. The decision table is written into `m1-measure.md` § Decision:
  - A (or B proxy) single-load → primary = launcher start.
  - `clear_restores_single_load: yes` → fallback = `/clear` after move; `no` or `gap` → fallback = end session and relaunch.
  - Path C shows no duplication on a class → no doctrine change for that class; report it.
- If M1 contradicts the premise wholesale, stop after M2 and return a blocker to the orchestrator.

### M3 — Template edit of the kanban dispatch rule (Priority Medium)

- Edit `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`: the `wt` bullet and the "A new card starts in a new worktree" clause name the primary and fallback entry forms. Generic wording only (`<name>`, `<card-id>` placeholders).
- `make build`, then copy to `.claude/rules/moai/workflow/kanban-dispatch.md` so the two are byte-identical.

### M4 — Local lane protocol (Priority Medium)

- Edit `.claude/rules/local/gitflow-lane-protocol.md` card-worktree entry clause to the same primary/fallback. No template mirror.

### M5 — Verification (Priority Low, mechanical)

- Run acceptance.md checks; record outputs into `progress.md` §E.2 (run-phase owner).

## §G Anti-Patterns

- Editing doctrine before M1 evidence is committed.
- Reporting "no duplication" from a grep that returned zero without the positive control.
- Running an unbounded probe or `go test ./...` on the full suite.
- Putting the card id or this SPEC id into the template copy.

## §H Cross-References

- `.moai/reports/t1219/verdict.md` — measured premises.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Dispatch format, § Isolation is entered, never provisioned.
- `.claude/rules/local/gitflow-lane-protocol.md` — card-worktree entry clause.
- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern — already uses the launcher form for re-entry.

## §I Open Questions

- [NEEDS CLARIFICATION: lane lifecycle] Kanban and Factory lanes are launched by hand once and carry many cards. A launcher-primary doctrine means the operator relaunches (or `--spawn`s) a lane session per card. Is that acceptable to the operator, or must the fallback be the normal path for long-lived lanes?
- [NEEDS CLARIFICATION: rules-diet ordering] Should M3 wait for the always-loaded rules diet (t1175) to land on develop, to avoid a conflict on the same file?
- [NEEDS CLARIFICATION: probe tree] May the probes run against the t1219 worktree itself (hooks write `.moai/state/` there), or should a separate probe worktree be created through the launcher?
