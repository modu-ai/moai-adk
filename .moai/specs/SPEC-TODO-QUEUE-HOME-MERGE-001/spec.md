---
id: SPEC-TODO-QUEUE-HOME-MERGE-001
title: "Merge the diverged project todo queue into the canonical home SQLite store"
version: "0.4.0"
status: in-progress
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, todo, queue, merge, data-migration, dangerous-operation"
tier: M
related_specs: [SPEC-WEB-TODO-QUEUE-001, SPEC-TODO-SQLITE-001, SPEC-TODO-LANDING-STATE-001]
---

# SPEC-TODO-QUEUE-HOME-MERGE-001 — Merge the diverged project todo queue into the canonical home SQLite store

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-13 | Initial draft (card t657, plan phase, manager-spec) |
| 0.2.0 | 2026-09-13 | Plan-audit delta (D1-D8): AC coverage for REQ-TQM-003/006/018; GEARS grammar fixes (REQ-TQM-007/010/017); high-water formula includes project `last_seq`; `Names` exact-match correction; identity-UUID collision edge case; REQ-count over-budget debt accepted on coordinator authority |
| 0.3.0 | 2026-09-13 | Re-plan against the measured M3 baseline (operator decision via lead): REQ-TQM-006 rewritten as the explicit duplicate discriminator (archived-population + byte-exact content; live cards default to renumber); §6 baseline replaced with measured store figures (home mtime moving); AC-TQM-010 (live-card discriminator) and AC-TQM-011 (runtime-persistence decision record) added; M4 execution forbidden until re-audit |
| 0.4.0 | 2026-09-13 | §F0 runtime-persistence decision gate SATISFIED: operator chose option (b) — merge-scope exclusion — via the lead's question round 2026-09-13; decision record (falsifier + loss ceiling + hash-gated backup recovery path) in plan.md §F0; AC-TQM-011 precondition marked satisfied; M4 remains forbidden until re-audit |

## 1. Background

The card t657 backlog queue grew in TWO places, each unaware of the other:

| Store | Path | Cards (plan-time measurement, 2026-09-12) | Last write |
|---|---|---|---|
| Project store (LEGACY) | `<primary-checkout>/.moai/state/todo/backlog.db` | 105 | 2026-09-11 03:20 |
| Home store (CANONICAL) | `~/.moai/db/moai-adk-go-<hash>/todo/backlog.db` | 95 | in use |

Operator decision (2026-09-12): **HOME is the canonical store.**

Lead's measurements (2026-09-12):

- **32 cards exist ONLY in project**: t587 t607 t608 t614 t616 t618 t619 t620 t621 t622 t624 t626 t627 t629 t630 t632 t633 t634 t635 t636 t637 t638 t639 t640 t642 t644 t645 t646 t658 t659 t660 t661
- **22 cards exist ONLY in home**
- Of **73 shared numbers**: **53 have identical content**; **20 are DIFFERENT cards that happen to share a number**

Root cause context: the queue id allocator reissues ids it cannot see (field collisions t654/t656/t657 — the same number minted twice in different stores). This SPEC is the **workaround** (merge two diverged stores); the allocator root-cause repair is a separate concern (see Out of Scope).

## 2. What was observed in the code (plan-phase read-only exploration)

Observed on this worktree (branch WT-todo-queue-merge, base origin/develop 5e0f71175):

- **Store mechanics** (`internal/kanban/backlog_store.go`): the queue document is `backlog.json` (the compatibility name) with a sibling SQLite engine file `backlog.db` (`backlogSQLitePath`). `NewBacklogStore(path)` opens a store at an EXPLICIT path — no root-resolution side effects. `LoadPure()` is the pure read; `Mutate(func(*BacklogRecord) error)` is the sole write path: advisory lock, whole-record read-modify-write, ONE transaction, `UNIQUE(id)`, `meta.last_seq` high-water mark. `normalizeBacklogRecord` lifts `last_seq` to the max present id across items AND archived entries.
- **Card identity** (`internal/kanban/todo_identity.go`): a `todo_identities` table carries a UUIDv7 per card (`entity_kind='card'`), keyed by local id; `ensureIdentity` accepts a **preferred** UUID, so a renumbered card can keep its identity.
- **States**: `queued` / `picked` / `dropped` (`backlog_store.go`), plus a separate archive population (archived entries with their findings).
- **Existing relocation** (`internal/kanban/state_dir.go` `relocateQueueArtifacts`): copies a legacy queue into the home store **only when the target has no queue** — it REFUSES (`queueExists(to)`) rather than merging. It cannot serve this card.
- **Existing export** (`internal/cli/todo_export.go`): `moai todo export-json` writes the live queue as legacy-format `backlog.json` under the store lock — a locked point-in-time copy of ONE store, not a merge.
- **No import or merge verb exists.**
- **Queue-root resolution** (`internal/kanban/todo_root.go` + `state_dir.go`): `ResolveTodoQueueRoot(base)` → the git primary checkout; `StateDirForRoot(root)` → `~/.moai/db/<project-key>/todo` (canonical) with project-local `.moai/state/todo` as the legacy/local shape. The retired fence marker (`backlogRetiredFileName`) is the existing mechanism that refuses stale writers on a dead path.
- **Landed evidence** (`internal/kanban/prlink_landed.go`): a THREE-valued landed query (landed / not-landed / unanswerable) run over a subject stream (`--format=%s`) against a ref (`LandedRefFor`; fallback `origin/main`); surfaced per card by `moai todo pr <id>`. This is the reconciliation evidence source.

## 3. Requirements (GEARS)

> Notation: GEARS (current). Subject is generalized per the GEARS grammar.

### 3.1 Backup-first (HARD)

**REQ-TQM-001** — Before any store mutation, the merge procedure shall copy both stores' artifacts (each store's `backlog.db`, its `backlog.json` sibling where present, and any `.migrated` quarantine) byte-for-byte into a backup directory under `.moai/reports/t657/backup/`, and shall verify each copy by byte count and SHA-256 equality with its source, recording both hashes in the verdict.

**REQ-TQM-002** — When any backup copy fails verification, the merge procedure shall abort with neither store modified and nothing deleted.

**REQ-TQM-003** — The merge procedure shall not delete, rename, or write the retired fence marker on any store artifact before the backup exists AND its integrity is verified (REQ-TQM-001). Read-only observation (`LoadPure`, mtime/size stats) is permitted before backup.

### 3.2 Zero-loss merge

**REQ-TQM-004** — The merged home store shall contain every card present in either source store — identified by its original id, or by an old→new row in the mapping table — across all populations (queued, picked, dropped, archived), with its findings.

**REQ-TQM-005** — When a project-only card's id collides with a DIFFERENT home card, the merge shall reissue the project-only card's id from the merged high-water mark (`max(last_seq_home, last_seq_project, max id home, max id project)`), preserve the card's UUIDv7 identity, rewrite references to the old id in finding texts and runtime assignments, and append one `old → new` row to the mapping-table file.

**REQ-TQM-006** — When a number is shared, the merge shall classify the pair as a resolved duplicate ONLY when the discriminator holds: the project card is in the project store's ARCHIVED population (completed work), AND the two cards are content-equal — byte-identical text, equal state, and equal `SpecID` (both present and equal, or both absent); metadata (AddedAt, landing evidence) is outside the comparison. Any project card in the LIVE population (queued, picked, or dropped) whose id matches a home card — live or archived — shall default to renumber-migrate per REQ-TQM-005, never to duplicate absorption, however similar the content: a matching number is not evidence of the same completed work. The merge report shall carry one resolved-duplicate row per duplicate, naming both stores' ids.

**REQ-TQM-007** — When a number is shared and the content differs (the 20 plan-time pairs), the home card shall keep the number, and the project variant shall be inserted under a newly issued id per REQ-TQM-005; the mapping table is the only traceability bridge for commit messages and `.moai/reports/<card-id>/` paths that reference the old number.

**REQ-TQM-008** — The mapping-table file shall live at `.moai/reports/t657/id-mapping.tsv`, one `old_id<TAB>new_id` row per renumbered card, committed alongside the merge evidence so the table survives the card's closure.

**REQ-TQM-009** — The merge shall execute the home-store write through the store's locked `Mutate` path as ONE transaction (whole-record write), and shall read both sources through `LoadPure`; the procedure shall not hand-edit SQLite bytes and shall not issue direct SQL mutations outside the engine.

### 3.3 State reconciliation

**REQ-TQM-010** — The merge shall reconcile card states to truthful values using the landed-attribution query (`moai todo pr` / `prlink_landed.go`) against the project's integration ref as the evidence source, under these rules:

1. A card in `queued`/`picked` whose landed query answers **landed** → record landing evidence (`moai todo landed <id>`) and archive the card (`done`).
2. A card whose landed query answers **unanswerable** → leave the state unchanged and list it in the reconciliation report as unresolved (operator follow-up; not a merge failure).
3. A `dropped` card stays `dropped`.
4. When a number-shared pair has divergent states, the home card's state shall win, and the project variant's state difference shall be recorded in the reconciliation report.

**REQ-TQM-011** — The merge shall write the reconciliation evidence (per-card rule applied + query answer observed) to `.moai/reports/t657/reconciliation.tsv` before the destructive step executes.

### 3.4 Approval and timing gates (procedural — evidenced in the verdict, not machine-testable code)

**REQ-TQM-012** — The merge procedure shall execute the destructive step (the home-store merge write, and separately the project-store retirement) ONLY inside a lead-designated window in which no concurrent queue writer is active.

**REQ-TQM-013** — Before the destructive step, the lane shall return a blocker report requesting operator approval routed through the LEAD; the lane shall never prompt the user directly and shall not execute the destructive step without the lead-recorded approval.

**REQ-TQM-014** — Immediately before the destructive step, the procedure shall re-verify freshness: both stores' last-write timestamps and card counts re-measured and compared against the plan-time figures; **when any figure has moved since plan, the procedure shall re-plan (re-diff the two stores) instead of executing a stale merge**.

### 3.5 Reversibility

**REQ-TQM-015** — A restore procedure shall be defined and rehearsed on fixture stores BEFORE the merge executes (it may never be exercised on real data).

**REQ-TQM-016** — When post-merge verification fails, the merge procedure shall restore both stores from the verified backups (byte-identical restore, re-verified by hash) before reporting failure.

### 3.6 Cleanup and verification

**REQ-TQM-017** — **Where** the merge is verified by actual observation (pre/post card counts and full id sets compared programmatically, and REQ-TQM-004's zero-loss invariant confirmed against the live merged store), the procedure shall retire the project store (`.moai/state/todo/`) by writing the existing retired fence marker and preserving the directory contents (rename, not deletion), so the store remains inspectable.

**REQ-TQM-018** — The merge tooling shall address real data by explicit absolute path (`<primary-checkout>/.moai/state/todo/` and the home store under `~/.moai/db/`); worktree-local `.moai/` artifacts are authoring artifacts and are never the merge subject.

## 4. Constraints

- **Dangerous data operation**: P1; every destructive step is last, gated, and reversible.
- The merge runs against LIVE data that lanes are actively reading; the single-writer window is a hard precondition (REQ-TQM-012), not a nicety.
- The merge tool is a one-off: no new permanent user-facing CLI verb; reuse the `internal/kanban` store API.
- macOS + Linux paths; `~` expansion resolved via the tool's own home resolution, never hand-concatenated.
- Zero concurrent writers during execution: lanes must be quiesced by the lead (freshness re-check is the mechanical probe, `lsof`/mtime + count stability the evidence).

## 5. Out of Scope

### Out of Scope — id-allocator root cause
- The queue id allocator reissuing ids it cannot see (t654/t656/t657 collisions) is NOT repaired here; this SPEC only works around it via renumbering + the mapping table. Root-cause repair is a separate card.

### Out of Scope — card t684 / branch WT-auto-done-on-land
- The pending t684 integration is neither touched nor depended on; this SPEC's work must merge cleanly regardless of t684's fate.

### Out of Scope — general-purpose import/export verbs
- No new permanent `moai todo import` / `merge` CLI surface; the one-off merge tool lives outside the user-facing verb set. `moai todo export-json` already covers the downgrade direction.

### Out of Scope — queue-root resolution changes
- `ResolveTodoQueueRoot` / `StateDirForRoot` / the adoption logic are unchanged; the merge addresses stores by explicit path.

### Out of Scope — queue content decisions
- Which unresolved reconciliation cards the operator later drops or re-prioritizes is operator work, recorded post-merge; the merge guarantees truthfulness of states, not prioritization.

## 6. Measured baseline (M3 dry-run, 2026-09-13 — supersedes the plan-time §1 figures; still subject to REQ-TQM-014 in-window re-derivation)

| Store | Live | Archived | last_seq | mtime |
|---|---|---|---|---|
| Home (canonical) | 59 (6 queued / 24 picked / 29 done) | 349 | 697 | **MOVING — live store**; home figures valid only inside the M4 freshness window |
| Project (legacy) | 105 (72 queued / 6 picked / 27 done) | 267 | 661 | stable since 2026-09-11 |

- Observed dry-run taxonomy (pre-discriminator): 290 duplicates / 82 renumbered / 0 pure migrations; union 780; high-water 779; zero-loss 780/780 PASS; 0 identity-UUID collisions.
- The 290/82 figures are NOT a baseline: REQ-TQM-006's discriminator reclassifies every duplicate whose project card is in the LIVE population to renumber-migrate. Post-discriminator counts are unknown until the M3-delta dry-run re-runs (plan.md §F M3-delta).
