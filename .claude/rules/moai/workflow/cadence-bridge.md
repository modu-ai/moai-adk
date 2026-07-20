---
description: "Sanctioned recipes composing native /loop and Cron with read-only /moai entry points for scheduled discovery work"
paths: "**/cadence-bridge.md,**/workflows/loop.md,.claude/skills/moai-workflow-loop/**,.moai/state/loop-verdict-*.json,.moai/reports/cadence/**"
---

# Cadence Bridge — Sanctioned Recipes Composing Native `/loop` and Cron with Read-Only `/moai` Entry Points

Sanctioned recipe catalog for scheduled ("cadence") discovery work. It composes Claude Code's native `/loop <interval>` scheduler (and Cron tools, e.g. CronCreate / CronList / CronDelete, where appropriate) with **strictly read-only or advisory** `/moai` entry points, so that drift, over-engineering creep, and leftover backlog items are found on a schedule without ever converting scheduled convenience into unattended write automation.

> **Loading scope**: Path-matched (loads only when a session touches a file matching the `paths:` glob above). This file was previously always-loaded; it is now scoped to reduce always-loaded context weight while preserving reachability whenever cadence work is active.
>
> **Trigger-condition list** (the rule loads when any of these files are touched): this rule file itself (`**/cadence-bridge.md` — self-referential maintenance); the `/loop` workflow file (`**/workflows/loop.md`); the loop skill tree (`.claude/skills/moai-workflow-loop/**`); loop-verdict state files (`.moai/state/loop-verdict-*.json` — read by Recipe 3); and the cadence backlog directory (`.moai/reports/cadence/**` — written by Recipes 1 and 2).
>
> **Honest-fallback clause**: `paths:` matches FILE GLOBS (a file-event mechanism) and therefore CANNOT mechanically catch a runtime `/loop` command typed into the session (a semantic event — a `/loop` invocation is not a file path). The path-match glob is a best-effort PRE-LOAD; when the orchestrator detects a `/loop` + `/moai` composition (recognizable from the `goal-directive.md` or loop-skill cross-reference), it MUST manually consult this rule regardless of path-match state. The manual-consult fallback is the GUARANTEE; the path-match pre-load is best-effort convenience.

## Why This Matters

The runtime ships a native `/loop` interval scheduler and Cron tools; MoAI ships read-only/advisory entry points (`/moai gate`, `/moai review --lean`) and a persisted ceiling-exit backlog (the loop's verdict-file mechanism — see `.claude/skills/moai/workflows/loop.md` § Ceiling-Exit Verdict Contract). Nothing previously composed them. `.claude/rules/moai/workflow/goal-directive.md` documents the native-`/loop`-vs-`/moai loop` distinction and stops there — every discovery pathway in the harness was user-initiated (one-shot subcommands) or PR-gated (CI watch activates only after a PR exists). Work that nobody asks about — drift, over-engineering creep, ceiling-exit leftovers — was found by nothing. This catalog is that bridge: doctrine only, no scheduler mechanics, no daemon, no new config keys.

## Catalog-Level HARD Invariant (binds every recipe — present and future)

[HARD] A scheduled invocation SHALL NOT write, commit, or push. Scheduled runs are restricted to read-only or advisory entry points, or — at most — Level-1 fixes (the fix.md Level 1 "Immediate: no approval required" class: formatting/import-sorting working-tree edits) left **uncommitted and unpushed**. A scheduled invocation SHALL NOT enter run-phase.

**The single governing sentence, binding all recipes present and future:** scheduled runs never commit, never push, never enter run-phase; Level-1 uncommitted working-tree edits are the sole permitted exception.

The Implementation Kickoff Approval (the plan→run HUMAN GATE; see `.claude/rules/moai/workflow/orchestration-mode-selection.md` header) is **human-only and cadence-unsatisfiable** — no scheduled invocation, however framed, can substitute for or bypass it. A cadence discovery is input to a human decision; it is never itself a decision.

This invariant is stated once, here, at the catalog level — it is NOT restated per-recipe, so that adding a new recipe in the future does not accidentally omit it.

## Eligibility Table — Which `/moai` Entry Points Are Cadence-Safe, and Why

| Entry point | Class | Read-only/advisory rationale |
|---|---|---|
| `/moai gate` | validation-only | Lint + format + type-check + test, <30s; a lightweight pre-commit quality gate that applies no fixes and modifies no files. |
| `/moai review --lean` | advisory-only | "Read-only and advisory: applies no fixes, modifies no files, renders no PASS/FAIL verdict" — an over-engineering-only lean scan. |
| `.moai/state/loop-verdict-*.json` read | prose-reader | Reads a persisted ceiling-exit / one-shot-residue file left by the loop workflow; no tool invocation and no mutation — a plain file read. |
| Level-1 fix.md class (formatting, import-sorting) | uncommitted-edit-only | The catalog invariant's sole write-adjacent exception: a working-tree edit is permitted, but it is never committed and never pushed by a scheduled run. |
| Harness discovery prompt (Recipe 4) | discovery-only | A self-contained read-only analysis prompt scoped to a built harness's domain; it finds drift and gaps, persists findings to the queue surface per the Discovery-to-Queue Contract, and applies no fixes and modifies no files beyond that queue record. |

**Not cadence-eligible** — these MUST NOT appear in any recipe or be added to this table: `/moai run`, `/moai sync`, `/moai loop`, and `/moai fix` beyond the Level-1-no-commit carve-out above, or any other subcommand that mutates git state. Scheduling any of these is out of scope for this catalog (see spec-level Out of Scope; a would-be recipe substituting one of these in is explicitly unsanctioned).

## Recipe Catalog

### Recipe 1 — Drift Watcher

```
/loop 30m /moai gate
```

Interval guidance: 30 minutes — frequent enough to catch drift shortly after it lands, cheap enough (`/moai gate` is a <30s validation pass) to run every half hour without noticeable cost. Read-only rationale: `/moai gate` runs lint + format + type-check + test in parallel and applies no fixes — a pure validation pass.

### Recipe 2 — Lean Review

```
nightly: /moai review --lean
```

Interval guidance: nightly (once per day, off-peak) — the lean scan is a heavier read than the drift watcher and is best run when it will not compete with active work. Read-only rationale: `--lean` "short-circuits the comprehensive 4-perspective analysis" and runs only the over-engineering scan; it is read-only and advisory, applying no fixes, modifying no files, and rendering no PASS/FAIL verdict.

### Recipe 3 — Backlog Re-Discovery

```
periodic read: .moai/state/loop-verdict-*.json
```

Interval guidance: periodic (e.g. daily or weekly, less time-sensitive than the other two) — this recipe re-surfaces ceiling-exit or one-shot-residue leftovers that the loop workflow already persisted; there is no new work to detect between reads, only accumulated leftovers to resurface. Read-only rationale: the recipe is a **prose reader** of an already-persisted JSON file (schema: `.claude/skills/moai/workflows/loop.md` § Remaining-Issue Persistence — `exit_kind`, `remaining_issues[]`, etc.); it never invokes a tool and never mutates the file it reads.

### Recipe 4 — Scheduled Harness Discovery

Recipe class covering scheduled discovery runs for any built harness. The scheduled payload is a **self-contained discovery prompt** — it names the harness and its domain scope and embeds its execution constraints inline, so a scheduled turn carries those constraints even when this rule file is not loaded. The payload deliberately does NOT invoke the harness entry command: the entry command dispatches write-capable specialists, and a scheduled run must never reach that dispatch path. Payload template (substitute the placeholders; loop and cron forms below):

```
Discovery run for the <harness-name> harness (domain: <harness-domain-scope>).
Perform read-only analysis only: scan this harness's domain scope for drift,
gaps, and stale artifacts. Queue persistence: record any findings to the
active TaskList when a session ledger is live, otherwise to
.moai/reports/cadence/<date>.md, and surface them at the next interactive
session. Beyond that queue record: no writes, no commits, no pushes, and no
run-phase entry.
```

Loop form (session-scoped — the armed loop dies with the session and must be re-armed per session): `/loop <interval> <discovery prompt above>`. Cron form (persistent across sessions): register the same discovery prompt via CronCreate with the chosen interval; CronList inspects and CronDelete unregisters it. Interval guidance: match the harness's drift tempo — `30m` for fast-moving domains, `nightly` for typical documentation or review harnesses, weekly for slow-moving ones. Read-only rationale: the payload performs analysis only and persists findings per the existing Discovery-to-Queue Contract (the section below — the queue mechanism is reused verbatim; no new queue surface is introduced). Where Cron tools are unavailable in the runtime version, this recipe degrades to the native `/loop` form only, consistent with the catalog's existing fallback clause.

## Discovery-to-Queue Contract

**When** a cadence run FINDS work — gate failures, lean-review findings, or unresolved verdict leftovers — the cadence run SHALL persist the discovery to a queue surface: the active TaskList when a session ledger is live, otherwise a backlog record at `.moai/reports/cadence/<date>.md` (an orchestrator-written local artifact, no Go loader — the reports namespace is the analyze-what-exists home for this kind of record). The discovery SHALL surface to the user at the next interactive session.

The cadence run SHALL NOT auto-execute any remediation. This is the contract's second half and is just as binding as the persistence half: a cadence run may find and queue work, but it never acts on what it finds beyond the Level-1-no-commit carve-out above.

## When to Schedule vs Event-Driven

Not every discovery pathway belongs on a cadence. A recipe here is appropriate when work needs to be found **without** a triggering event — drift that accumulates silently, over-engineering that creeps in unnoticed, leftovers nobody remembers to re-check. Do NOT cron what should be event-driven: CI failures already have a dedicated event path (`scripts/ci-watch/run.sh` → the CI auto-fix loop; see `.claude/rules/moai/workflow/ci-watch-protocol.md` and `.claude/rules/moai/workflow/ci-autofix-protocol.md`) that activates on the event itself (a failing check), not on a wall-clock interval. Scheduling a cadence recipe to re-poll something that already has an event-driven trigger duplicates effort and adds latency instead of removing it.

## Fallback and Edge Cases

- **Concurrent session holds the checkout**: the read-only recipes (drift watcher, lean review, backlog re-discovery) are race-safe by construction — they do not write. The backlog-record write (when a session ledger is not live) follows the standard multi-session pre-spawn check discipline, or degrades to report-only output if a race is detected.
- **`.moai/state/loop-verdict-*.json` absent**: recipe 3 completes as a no-op with an informational note, not an error — absence means no ceiling exits have occurred yet, or the producing mechanism has not landed.
- **A pasted recipe substitutes a write-capable subcommand** (e.g. `/loop 30m /moai fix`): this is explicitly unsanctioned per the catalog invariant and the eligibility table above; reviewers and orchestrators reject it on sight. Mechanical blocking of such a substitution is out of scope for this doctrine-only bridge.
- **Cron tools unavailable in the runtime version**: recipes degrade to native `/loop` only; no functionality is lost for Recipes 1-3, since none require Cron specifically. Recipe 4's cron form degrades to its session-scoped `/loop` form — cross-session persistence is lost until a later session re-arms the loop.

## Cross-References

- `.claude/rules/moai/workflow/goal-directive.md` § Comparing Autonomous-Continuation Approaches — the native-`/loop`-vs-`/moai loop` distinctness note this bridge composes around, without merging.
- `.claude/skills/moai/workflows/gate.md`, `.claude/skills/moai/workflows/review.md` (`--lean` mode), `.claude/skills/moai/workflows/fix.md` (Level 1 classification) — the cited read-only/advisory/uncommitted-edit-only entry points.
- `.claude/skills/moai/workflows/loop.md` § Ceiling-Exit Verdict Contract / § Remaining-Issue Persistence — the verdict-file schema recipe 3 reads (consumed here, not defined here).
- `.claude/rules/moai/workflow/ci-watch-protocol.md`, `.claude/rules/moai/workflow/ci-autofix-protocol.md` — the event-driven CI path this catalog deliberately does not duplicate.
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (header) — the Implementation Kickoff Approval invariant this bridge cites and never weakens.
- `.claude/rules/moai/core/askuser-protocol.md` — the AskUserQuestion channel monopoly, unaffected by any cadence recipe (a cadence discovery is input to a human decision, never a decision itself).

---

Version: 1.0.0
Status: Active — doctrine-only bridge; no Go scheduler, no daemon, no new config keys
Classification: Evolvable orchestration guidance — applies whenever a scheduled `/loop` or Cron invocation composes with a `/moai` entry point
