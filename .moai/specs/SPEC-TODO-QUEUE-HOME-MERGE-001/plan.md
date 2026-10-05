---
id: SPEC-TODO-QUEUE-HOME-MERGE-001
title: "Merge the diverged project todo queue into the canonical home SQLite store"
version: "0.4.0"
created: 2026-09-13
---

# plan.md — SPEC-TODO-QUEUE-HOME-MERGE-001 (v0.3.0 — rewritten against the measured M3 baseline)

> Operator decision via lead: M4 live execution is FORBIDDEN until this rewrite lands and the SPEC re-audits. This plan supersedes the v0.2.0 plan's plan-time figures (105/95, 53/20/32, expected 127) with the measured M3 dry-run baseline. Priority labels only — no time estimates.

## §A Context (updated for executed work)

- **Tooling now EXISTS and ran**: the run phase built the merge core (`internal/kanban/todo_queue_merge.go` — `MergeBacklogRecords`, pure) and the procedure shell (`internal/kanban/todo_merge_procedure.go` — dry-run + the ONE-`Mutate` home write via the `mergeMutate` seam). M3 executed a dry-run; its measurements are the authoritative baseline below.
- Store mechanics unchanged: `NewBacklogStore(path)` explicit path, `LoadPure()` read, ONE `Mutate` transaction write, UUIDv7 identities, retired fence marker, three-valued landed query.
- Plan/authoring happens in worktree `.claude/worktrees/t657-queue-merge` (branch `WT-todo-queue-merge`); the REAL stores live at `<primary-checkout>/.moai/state/todo/` and `~/.moai/db/moai-adk-go-<hash>/todo/` (REQ-TQM-018 absolute paths).

## §A.1 Measured baseline (M3 dry-run, 2026-09-13 — AUTHORITATIVE; supersedes §1's plan-time figures)

| Store | Live | Archived | last_seq | mtime |
|---|---|---|---|---|
| Home (canonical) | 59 (6 queued / 24 picked / 29 done) | 349 | 697 | **MOVING — live store**; home figures are valid only inside the M4 freshness window |
| Project (legacy) | 105 (72 queued / 6 picked / 27 done) | 267 | 661 | stable since 2026-09-11 |

Observed dry-run taxonomy: 290 duplicates / 82 renumbered / 0 pure migrations; union 780; high-water 779; zero-loss 780/780 PASS; 0 identity-UUID collisions.

**These taxonomy counts are NOT a baseline**: the discriminator (§B1, REQ-TQM-006 v2) reclassifies every duplicate whose project card is in the LIVE population to renumber-migrate. Post-discriminator counts are unknown until M3-delta re-runs the dry-run. Union 780 and the pre/post freshness method carry over; every count is re-derived inside the M4 window.

## §B Known Issues (updated)

1. **Discriminator gap (operator-named hazard, blocking)** — the implemented `classify` accepts a project LIVE card (queued/picked/dropped) as a duplicate when its id matches a home card — including a home ARCHIVED card — and `mergeCardContentEqual` passes (byte-equal text, equal state, equal SpecID). A live card that is a different generation of the same task is thereby SILENTLY ABSORBED as completed work. M1-delta (§F) tightens the predicate; M4 stays forbidden until it lands and re-runs M3-delta.
2. **Runtime persistence gap** — `writeRecord` (the sole home write inside `Mutate`) does not touch `todo_runtime_runs`/`todo_runtime_assignments`; those tables are written only by the slot-lease upsert path (`todo_runtime.go`, consumers `slot_lease.go`/`factory_slots.go`). The merge core BUILDS a runtime projection in its report, but the ONE-Mutate write drops it — the project store's runtime rows do not reach home. Decision required before M4: §F0.
3. **Moving home store** — the home store is LIVE (mtime moving since measurement). Every home-side figure in §A.1 decays; only the in-window freshness re-derivation (REQ-TQM-014) is evidence.
4. The id-allocator reissuance root cause remains open — renumber + mapping table is the workaround; a later allocator fix must consult `.moai/reports/t657/id-mapping.tsv`.
5. `BacklogFinding.Names` is exact-match (`backlog_store.go:151-153`); the token-boundary hazard lives in FREE-TEXT note fields — `rewriteCardTokens` must stay token-boundary aware (`t642` must not corrupt `t6420`).

## §C Pre-flight (read-only; re-run ALL of it inside the M4 window — REQ-TQM-014)

```bash
# P1. Store locations + last-write stamps (home MUST show continued movement
#     BEFORE the window and stillness DURING it)
stat -f '%m %N' <primary>/.moai/state/todo/backlog.db
stat -f '%m %N' ~/.moai/db/moai-adk-go-*/todo/backlog.db
# P2. Full figure re-derivation via the merge tool's dry-run (counts, union,
#     high-water, taxonomy, identity collisions) — replaces §A.1 numbers
# P3. No live writers in the designated window (lead confirms lanes quiesced;
#     `moai session list --json` + `lsof` on both db paths empty)
# P4. Landed-ref re-derivation (origin/HEAD is a MOVING ref — the M3 reading
#     `LandedRefFor = origin/develop` is a dated reference, not a baseline;
#     re-run and record the fresh answer in the window):
git symbolic-ref refs/remotes/origin/HEAD
```

**Moving-home caveat (front and center)**: because home is a live store, P2's output is stale the moment a lane writes. The freshness check is therefore a TWO-READING bracket: re-run P1 immediately BEFORE the destructive write and again immediately AFTER; both readings must show the same mtime/count pair as the P2 baseline. Any movement ⇒ ABORT, re-dry-run, re-baseline (re-plan), re-gate. The bracket readings are verdict evidence.

## §D Constraints

- Destructive steps (M4, M5) NEVER run without: (1) lead-designated window, (2) operator approval recorded via the lead, (3) freshness bracket PASS, (4) the §F0 runtime decision — recorded 2026-09-13, option (b).
- The merge write is ONE `Mutate` transaction on the home store; a failed transaction leaves the prior store intact (engine guarantee).
- Backup, verification, mapping table, reconciliation evidence under `.moai/reports/t657/`.
- 85% coverage on new/changed code; merge-core tests use `t.TempDir()` fixture stores only — tests never touch real queues.

## §E Self-Verification

M1-delta: unit tests (discriminator fixtures). M2: restore rehearsal output (already rehearsed; re-run if restore path changed). M3-delta: fresh dry-run report = the new baseline, diffed against §A.1 with reclassification deltas explained row-by-row. M4/M5: evidence per acceptance.md — counts, id sets, hashes, bracket readings; commands + verbatim output in the verdict.

## §F0 Runtime persistence — decision record (RESOLVED 2026-09-13)

> **Decision: option (b) — merge-scope exclusion of runtime persistence.**
> Decider: the operator. Channel: the lead's question round, 2026-09-13.
> Falsifier retained: (a) becomes justified only if the operator names a live consumer of project-side runtime history — none is known.
> Loss ceiling: the project store's HISTORICAL `todo_runtime_runs` / `todo_runtime_assignments` rows (past factory-run audit trail) are not migrated. Active leases heal on their next slot-lease write (the machinery upserts run and assignment rows itself); dead leases have no live consumer.
> Recovery path: the project-store backup (hash-gated, never deleted) preserves the audit trail permanently; the merge report's runtime projection documents what would have migrated.

**What is at stake**: the project store holds `todo_runtime_runs` + `todo_runtime_assignments` rows (factory/slot-lease history: run registrations, per-card lease owners, reported states). The merge's ONE-`Mutate` home write does not persist them — they are dropped. The merge core already builds the would-be projection (new runs join; assignments join under rewritten card ids, home winning collisions).

**Option (a) — extend the engine to persist runtime on the merge write.**
- Shape: NOT a `writeRecord` extension (that function is the `@MX:ANCHOR` sole write path for EVERY card edit — extending it makes every `todo add` carry runtime-write logic); the minimal shape is a sibling method reusing the existing upsert statements (`INSERT ... ON CONFLICT DO UPDATE`, already present in `todo_runtime.go`), invoked by the procedure after the Mutate inside its own short transaction.
- Blast radius: one new method + procedure call; `writeRecord` untouched; `copyRuntime` stays migration-only (its contract forbids live-DB use). Tests: schema-freeze test untouched; new tests mirror the existing migration roundtrip's runtime comparison.
- Risk: a SECOND transaction after the cards' Mutate weakens atomicity — cards can land while runtime fails (degradation is reportable, not silent: the procedure records the failure and continues to verification with a marked gap). The union-vs-clobber problem is real: home's own runtime rows may change between the merge read and this write, so the write must UNION (the report's projection already does), never replace.

**Option (b) — exclude runtime persistence from merge scope (accept the loss with a named ceiling).**
- What is lost: only the project store's HISTORICAL run/assignment rows (past factory runs, audit trail). Active-state healing: the slot-lease machinery upserts BOTH run rows and assignment rows on lease activity (`ON CONFLICT DO UPDATE`), so any lease that is still live re-registers itself on its next write — projections are recomputed by their producer. Dead leases (project store stable since 2026-09-11; its lanes are long gone) have no live consumer.
- Recovery: the project-store backup (never deleted) preserves the audit trail permanently and readably (SQLite); the merge report's runtime projection documents what WOULD have migrated.
- Risk: an operator auditing historical factory runs loses the project-side half of the story unless they consult the backup.

**Recommendation (presented to the operator): (b).** Reasons: (i) the projection's producer heals live state by construction — what (a) would add over (b) is exclusively dead history; (ii) (a)'s second transaction buys non-atomic complexity on a one-off procedure; (iii) the backup makes (b)'s loss reversible-by-hand forever, which is the reversibility posture this SPEC already commits to.

**The operator adopted the recommendation: option (b)** — see the decision record atop this section. The choice is recorded per AC-TQM-011.

## §F Milestones (destructive steps LAST and gated)

### M1-delta (High) — Discriminator fix, pure and tested
- Tighten `classify`: a project card is classifiable as duplicate ONLY when it comes from the project record's ARCHIVED population AND `mergeCardContentEqual` holds against its home counterpart. Every project card from the LIVE population (`Items` — queued, picked, dropped) whose id exists in home (live or archived) is renumbered above the merged high-water, with mapping row, reference rewrite, and identity resolution — regardless of content similarity. Content equality stays BYTE-EXACT (no normalization): anything not byte-identical renumbers. Fail-safe default: ambiguity resolves to renumber, never absorption.
- Fixture tests: (i) a PICKED project card whose id+text match a home ARCHIVED card resolves renumber with a mapping row and its content present under the new id (the operator's named scenario); same for queued and dropped; (ii) the duplicate population after merge contains ONLY project-archived-origin cards; (iii) existing zero-loss/renumber/dedupe tests still pass.

### M2 (High) — Backup + restore (rehearsed; re-run unchanged)
- Backup (copy + SHA-256 + byte count, both stores, artifacts per REQ-TQM-001) → `.moai/reports/t657/backup/`; restore procedure rehearsed on fixtures. No changes from v0.2.0; re-run against the real stores happens inside M4's window, first.

### M3-delta (Medium) — Re-baseline dry-run (read-only)
- Re-run the dry-run against the LIVE stores after M1-delta lands. Output = the NEW authoritative baseline: counts, union, high-water, post-discriminator taxonomy, identity-collision count, and the reclassification ledger (which of the 290 duplicates became renumbers). Diffed against §A.1 and recorded; the diff is verdict evidence. This milestone replaces the stale 290/82 figures and is the LAST read before M4.

### M4 (High, DESTRUCTIVE, GATED) — Execute the merge
Gate sequence — ALL recorded in the verdict before proceeding: lead-designated window + quiesced lanes → blocker-report approval returned through the lead → §F0 runtime decision (satisfied 2026-09-13, option b) → freshness bracket P1+P2 PASS (both bracket readings still).
Procedure:
1. Backup (M2) + verify hashes → verdict.
2. `moai todo landed <id>` evidence recording per M3 reconciliation rows resolving to landed.
3. Merge: `LoadPure` ×2 → merge core → ONE `Mutate` on home (runtime handling per §F0: option (b) — none; the report's runtime projection is preserved as evidence).
4. Post-write freshness bracket reading (must equal pre-write).
5. Verify: post-count = re-derived union; id-set comparison programmatic; zero-loss vs both pre-reads (duplicates restricted to project-archived origins per AC-TQM-010); mapping table committed to `.moai/reports/t657/id-mapping.tsv`.
6. On ANY verification failure: restore both stores (M2 procedure), re-verify hashes, report failure.

### M5 (Low, DESTRUCTIVE, GATED) — Retire the project store
- Only after M4 verification observed PASS: write the retired fence marker into `.moai/state/todo/`, rename the directory to `.moai/state/todo.retired-<date>` (rename, NOT deletion), re-run `moai todo list --json --limit 0` from the primary checkout to confirm the home queue serves the identical set, record pre/post counts + id-set comparison in the verdict.

### Rollback (rehearsed in M2; unchanged)
- Restore both stores from the verified backups (byte-identical, hash-verified); remove the retired marker/rename if M5 had started; re-run `moai todo list` to confirm service; record the incident in the verdict.

## §G Anti-Patterns

- Never classify a live-population project card as a duplicate, however similar its content.
- Never re-derive counts outside the window and execute on them; never execute on a one-sided freshness reading.
- Never `sqlite3`-CLI into either store; never copy half a store's artifacts; never use `copyRuntime` on a live database.
- Never renumber without a mapping row; never rewrite a free-text reference without token boundaries.
- Never prompt the user from the lane; approval travels lane → blocker report → lead → operator.
- Never treat worktree-local `.moai/` data as the live queue.

## §H Cross-References

- spec.md §3 REQ-TQM-001..018 (REQ-TQM-006 v2 = the discriminator); acceptance.md AC-TQM-001..011.
- Code: `internal/kanban/todo_queue_merge.go` (merge core), `todo_merge_procedure.go` (procedure + `mergeMutate` seam), `backlog_store.go`, `backlog_migrate.go` (`writeRecord` — runtime gap), `todo_runtime.go` (upsert path), `slot_lease.go` (runtime consumer), `todo_identity.go`, `state_dir.go`, `prlink_landed.go`.
- Evidence home: `.moai/reports/t657/` (backup/, id-mapping.tsv, reconciliation.tsv, runtime projection).
- Related SPECs: SPEC-WEB-TODO-QUEUE-001, SPEC-TODO-SQLITE-001, SPEC-TODO-LANDING-STATE-001.
