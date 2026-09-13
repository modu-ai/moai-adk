---
id: SPEC-TODO-QUEUE-HOME-MERGE-001
title: "Merge the diverged project todo queue into the canonical home SQLite store"
version: "0.2.0"
created: 2026-09-13
---

# plan.md — SPEC-TODO-QUEUE-HOME-MERGE-001

> Milestones are ordered by decision-reversibility: the tested, reversible merge-core first; every destructive step LAST and gated. No time estimates — priority labels only.

## §A Context

- Two diverged SQLite queue stores (see spec.md §1). HOME canonical (operator decision 2026-09-12). Plan/run authoring happens in worktree `.claude/worktrees/t657-queue-merge`; the REAL stores live at `<primary-checkout>/.moai/state/todo/` and `~/.moai/db/moai-adk-go-<hash>/todo/`. All data-touching steps address absolute paths explicitly (REQ-TQM-018).
- Reusable machinery (observed, not assumed): `NewBacklogStore(path)` (explicit path, no adoption), `LoadPure()` (pure read), `Mutate` (locked single-transaction whole-record write), UUIDv7 identity with preferred-UUID preservation, retired fence marker, three-valued landed query.

## §B Known Issues

- The existing one-time relocation (`relocateQueueArtifacts`) refuses when the target has a queue — it cannot merge; do not extend it.
- The id allocator reissuance root cause remains open — this merge's renumber+mapping is the workaround; a later allocator fix must consult the mapping table (`.moai/reports/t657/id-mapping.tsv`).
- `BacklogFinding.Names` is an EXACT-match predicate (`backlog_store.go:151-153`), so structured finding references need no id rewriting; the token-boundary hazard lives in FREE-TEXT note fields that embed id tokens — any rewrite there must be token-exact (e.g. `t642` must not corrupt `t6420`) and verified by a boundary-aware pass.

## §C Pre-flight (all read-only; re-run at execution time per REQ-TQM-014)

```bash
# P1. Store locations + last-write stamps
stat -f '%m %N' <primary>/.moai/state/todo/backlog.db
stat -f '%m %N' ~/.moai/db/moai-adk-go-*/todo/backlog.db
# P2. Card counts + id sets (from the merge tool's dry-run mode, below)
# P3. No live writers in the designated window (lead confirms lanes quiesced;
#     `moai session list --json` + `lsof <db paths>` empty)
```

## §D Constraints

- Destructive steps (M4, M5) NEVER run without: (1) lead-designated window, (2) operator approval recorded via the lead, (3) freshness re-check PASS.
- The merge write is ONE `Mutate` transaction on the home store; a failed transaction leaves the prior store intact (engine guarantee).
- Backup, verification, mapping table, and reconciliation evidence all land under `.moai/reports/t657/`.
- 85% coverage on new code; merge-core tested with `t.TempDir()` fixture stores only — tests never touch real queues.

## §E Self-Verification

Per-milestone: M1 unit tests (merge-core on fixtures); M2 restore rehearsal output; M3 dry-run report diffed against plan baseline; M4/M5 evidence per acceptance.md (counts, id sets, hashes — commands + verbatim output in the verdict).

## §F Milestones

### M1 (High) — Merge-core, pure and tested
- A pure merge function over two `BacklogRecord` values: dedupe identical-number/identical-content pairs; renumber colliding project-only ids from the merged high-water mark; preserve UUIDv7 identities (preferred); token-boundary rewrite of finding texts and runtime assignments; produce the merged record + in-memory mapping rows.
- Unit tests on `t.TempDir()` fixture stores covering: zero-loss, collision renumbering, identical-pair dedupe, finding-reference rewrite (incl. the `t642` vs `t6420` boundary), `last_seq` lift, dropped/archive populations preserved, and an identity-UUID collision across stores resolved by issuing a fresh identity + reconciliation row (never an error or silent reuse).
- Thin shell: `LoadPure` both stores → merge function → ONE `Mutate` on home. A `--dry-run` mode that prints the proposed actions + mapping without writing.

### M2 (High) — Backup + restore, rehearsed
- Backup procedure (copy + SHA-256 + byte count, both stores, artifacts enumerated in REQ-TQM-001) → `.moai/reports/t657/backup/`.
- Restore procedure (reverse copy + re-hash verification) defined and REHEARSED on fixture stores before M4.

### M3 (Medium) — Reconciliation evidence (read-only)
- For every queued/picked card in the union: run the landed query against the integration ref; apply REQ-TQM-010 rules; write `.moai/reports/t657/reconciliation.tsv`. No store writes in this milestone beyond the `moai todo landed` evidence recording it mandates (which itself runs inside the M4 window).

### M4 (High, DESTRUCTIVE, GATED) — Execute the merge
Gate sequence (all three recorded in the verdict before proceeding): lead-designated window + quiesced lanes → blocker-report approval returned through the lead → freshness re-check (P1/P2 re-run; any movement ⇒ re-plan).
Procedure:
1. Backup (M2 procedure) + verify hashes → record in verdict.
2. `moai todo landed <id>` evidence recording per M3 rows that resolve to landed.
3. Merge tool: `LoadPure` ×2 → merge-core → ONE `Mutate` on home.
4. Verify: post-count = 127 (or freshness-corrected figure); id-set comparison programmatic; zero-loss check vs the union of both pre-reads; mapping table committed to `.moai/reports/t657/id-mapping.tsv`.
5. On ANY verification failure: restore both stores (M2 procedure), re-verify hashes, report failure.

### M5 (Low, DESTRUCTIVE, GATED) — Retire the project store
- Only after M4 verification is observed PASS: write the retired fence marker into `.moai/state/todo/`, rename the directory to `.moai/state/todo.retired-<date>` (rename, NOT deletion), re-run `moai todo list` from the primary checkout to confirm the home queue serves, record pre/post counts + id-set comparison in the verdict.

### Rollback (defined before M4; rehearsed in M2)
- Restore both stores from the verified backups (byte-identical, hash-verified); remove the retired marker/rename if M5 had started; re-run `moai todo list` to confirm service; record the incident in the verdict.

## §G Anti-Patterns

- Never `sqlite3`-CLI into either store; never copy half of a store's artifacts.
- Never renumber a card without a mapping row; never rewrite a finding reference without token boundaries.
- Never execute M4/M5 on a stale freshness measurement.
- Never prompt the user from the lane; the approval channel is lane → blocker report → lead → operator.
- Never treat worktree-local `.moai/` data as the live queue.

## §H Cross-References

- spec.md §3 REQ-TQM-001..018; acceptance.md AC matrix.
- Code: `internal/kanban/backlog_store.go`, `internal/kanban/backlog_migrate.go`, `internal/kanban/todo_identity.go`, `internal/kanban/state_dir.go`, `internal/kanban/todo_root.go`, `internal/kanban/prlink_landed.go`, `internal/cli/todo_export.go`.
- Evidence home: `.moai/reports/t657/` (backup/, id-mapping.tsv, reconciliation.tsv).
- Related SPECs: SPEC-WEB-TODO-QUEUE-001 (queue-root resolution), SPEC-TODO-SQLITE-001 (store engine), SPEC-TODO-LANDING-STATE-001 (landed query).
