---
id: SPEC-TODO-QUEUE-HOME-MERGE-001
title: "Merge the diverged project todo queue into the canonical home SQLite store"
version: "0.4.0"
created: 2026-09-13
---

# acceptance.md — SPEC-TODO-QUEUE-HOME-MERGE-001

## §D AC Matrix

| AC | Requirement | Subject | Binary-testable assertion |
|---|---|---|---|
| AC-TQM-001 | REQ-TQM-001/002 | Backup integrity | Both backups hash-equal to sources |
| AC-TQM-002 | REQ-TQM-004 | Zero-loss invariant | Every card in either store represented post-merge |
| AC-TQM-003 | REQ-TQM-005/007/008 | Mapping completeness | 1:1 between renumbered cards and mapping rows |
| AC-TQM-004 | REQ-TQM-005 | Reference rewrite | No stale old-id references in findings/assignments |
| AC-TQM-005 | REQ-TQM-010/011 | State reconciliation | Reconciliation rows carry rule + observed query answer |
| AC-TQM-006 | REQ-TQM-017 | Post-cleanup comparison | Pre/post count + id-set equality observed |
| AC-TQM-007 | REQ-TQM-015/016 | Rollback | Restore rehearsal on fixtures hash-verifies |
| AC-TQM-008 | REQ-TQM-012/013/014 | Approval + timing gates | Evidenced as preconditions in the verdict |
| AC-TQM-009 | REQ-TQM-009/§B | Mechanism + testability | Merge-core unit-tested on fixtures; write via one Mutate |
| AC-TQM-010 | REQ-TQM-006 (discriminator) | Live-card discriminator | Active project card vs matching home archived card resolves renumber, never duplicate |
| AC-TQM-011 | plan.md §F0 | Runtime-persistence decision record | Operator's option choice + loss ceiling recorded in the verdict |

## AC-TQM-001 — Backup integrity

**Given** both live stores at their absolute paths, **When** the backup procedure runs, **Then** for every copied artifact the recorded SHA-256 of the copy equals the SHA-256 of the source measured in the same run, the recorded byte count equals the source's, and the verdict carries both hash values verbatim. **And** when any copy fails verification, the merge aborts with exit non-zero, both stores byte-unchanged (source hashes re-measured equal), and nothing deleted. **And** between audit start and backup-hash verification, no artifact under either store is created, renamed, or deleted — evidenced in the verdict by a pre/post path-set + mtime listing of both store directories.

## AC-TQM-002 — Zero-loss invariant

**Given** the union U of card ids read from both stores immediately pre-merge (via `LoadPure`), **When** the merge completes, **Then** for every card in U the merged home store contains that card either under its original id or under the new id paired to it by a mapping-table row — checked per population (queued, picked, dropped, archived) and per card's text — with total = |U| (re-derived inside the M4 window; the M3-delta dry-run's union, plan-basis 780). The check is a programmatic id/text set comparison, not a count assertion alone. **And** the merge report carries one resolved-duplicate row per duplicate pair, naming both stores' ids (REQ-TQM-006) — and every duplicate row's project card originated from the project store's ARCHIVED population (AC-TQM-010): a live-population absorption is a zero-loss violation.

## AC-TQM-003 — Mapping-table completeness

**Given** the set R of project-only cards that required renumbering, **When** the merge writes `.moai/reports/t657/id-mapping.tsv`, **Then** the file holds exactly one `old_id<TAB>new_id` row per card in R (|rows| = |R|), every `new_id` exists in the merged store, every `old_id` does NOT, and the file is committed with the merge evidence.

## AC-TQM-004 — Reference rewrite correctness

**Given** findings and runtime assignments that referenced a renumbered card's old id, **When** the merge completes, **Then** no token-boundary occurrence of the old id remains in finding texts or assignment card references (a `t642` rewrite must not alter `t6420`), verified by scanning the merged record for old-id tokens of R against token boundaries.

## AC-TQM-005 — State reconciliation evidence

**Given** every queued/picked card in the union, **When** reconciliation runs, **Then** `.moai/reports/t657/reconciliation.tsv` carries one row per card: card id, pre-state, the landed-query answer observed (landed / not-landed / unanswerable), and the rule applied — and every card resolved to landed is archived WITH its `moai todo landed` evidence recorded. Cards answering unanswerable remain state-unchanged and listed.

## AC-TQM-006 — Post-cleanup count/set comparison

**Given** a PASS merge verification, **When** the project store is retired, **Then** the verdict records the actual observed commands + output for: (a) pre-retirement merged-store count + full id set, (b) post-retirement `moai todo list --json --limit 0` from the primary checkout serving the identical set, (c) the retired directory renamed intact (not deleted) with the fence marker present.

## AC-TQM-007 — Rollback rehearsed

**Given** fixture stores (t.TempDir) with known content, **When** the restore procedure runs against mutated fixtures, **Then** both fixtures are byte-identical to their backups (SHA-256 equality) and the rehearsed command sequence is recorded in the verdict — before M4 executes.

## AC-TQM-008 — Approval + timing gates (procedural preconditions)

Not machine-testable code; each gate is EVIDENCED in the verdict as follows:
- **Window**: the lead-designated execution time + the quiescence probe output (`moai session list --json` empty for this scope; no live writer on the db files).
- **Approval**: the blocker report the lane returned, and the lead-relayed operator approval text + timestamp.
- **Freshness**: the REQ-TQM-014 re-measurement (mtimes + counts) with its comparison against the plan baseline, run within the window immediately before the destructive step; any movement → re-plan, and the re-plan is itself recorded.

## AC-TQM-009 — Mechanism and testability

**Given** the merge tooling, **When** its unit suite runs, **Then** the merge-core is exercised purely on fixture stores with ≥85% coverage on new code, and the real-data write path issues exactly ONE `Mutate` transaction on the home store (asserted in tests; no direct SQL mutation, no byte-level edits). **And** the tool resolves both stores by absolute path and refuses a queue resolved from the current working tree (REQ-TQM-018).

## AC-TQM-010 — Live-card discriminator (no silent absorption)

**Given** a project card in the LIVE population (queued, picked, or dropped) whose id matches a home card — including a home ARCHIVED card — with byte-identical text, equal state, and equal `SpecID`, **When** the merge classifies it, **Then** it resolves to renumber-migrate with a mapping row and its content is present in the merged store under the new id; it is never classified duplicate and never dropped. **And** across the whole merge, every entry in the duplicate population originated from the project record's ARCHIVED population, verified by a population check over the merge report against the pre-merge project record. Unit-tested on fixtures (M1-delta), including the operator's named scenario: a PICKED project card vs a matching home archived card.

## AC-TQM-011 — Runtime-persistence decision record

**Given** the two runtime-persistence options of plan.md §F0 — (a) persist `todo_runtime_runs`/`todo_runtime_assignments` to home via a sibling upsert-path write, or (b) exclude runtime persistence from merge scope — **When** the operator decides via the lead, **Then** the verdict records the chosen option and the comparison's decisive factors. **And** where (b) is chosen, the verdict states the explicit loss ceiling: project-store HISTORICAL runtime rows are not migrated; active leases heal on their next slot-lease write (the machinery upserts run and assignment rows itself); the project-store backup remains the permanent recovery path for the audit trail; and the merge report's runtime projection is preserved as evidence. The operator's decision is a precondition for M4 entry.

> **Precondition status: SATISFIED — decision recorded 2026-09-13, option (b) (plan.md §F0 decision record; decider = operator, channel = the lead's question round 2026-09-13).** The Given/When/Then above remains the machine-checkable form: the M4 verdict must still carry the recorded option, the loss ceiling, and the recovery-path statement.

## §D.1 Severity — AC-TQM-001/002/010 are MUST-PASS (data loss or live-work absorption = card failure); AC-TQM-003..007 and AC-TQM-009 MUST-PASS; AC-TQM-008 and AC-TQM-011 are gates whose absence of evidence blocks M4 entry.

## §D.2 Quality gates (TRUST 5)

- Tested: merge-core unit coverage ≥85% on fixtures; go test on affected packages.
- Readable/Unified: project Go conventions; golangci-lint clean.
- Secured: no path traversal from mapping rows; explicit absolute paths only; no SQL interpolation.
- Trackable: evidence under `.moai/reports/t657/`; conventional commits citing t657.

## §D.3 Indirect verification

Zero-loss (AC-TQM-002) is additionally cross-checked by text search: every card text present pre-merge is found in the merged store (catches an id-mapping bug that a pure id comparison would miss).

## §D.4 Closure gates

M4 closes only when AC-TQM-001..005 are observed PASS; M5 closes only when AC-TQM-006 is observed PASS. AC-TQM-008 evidence must exist BEFORE M4 entry.

## §D.5 Edge cases

- A project-only card whose id collides AND whose content differs from the home card (the 20 pairs) — renumber, never overwrite.
- Renumbering must lift `last_seq` to max across both stores including archived ids (prevents reissuance-by-merge — the very defect class this card works around; the merge must not itself mint a colliding id).
- Empty/absent artifacts (no `.migrated` quarantine, missing `backlog.json` sibling) — backup enumerates what exists, records absent items as absent.
- Store lock contention during the merge write — the window guarantees no writers; lock acquisition failure aborts, never retries blind.
- An identity-UUID collision across stores — the merge issues a fresh UUIDv7 identity for the affected card and records the re-issue in `.moai/reports/t657/reconciliation.tsv` (tested in M1: fixture stores carrying the same card UUID must merge with a fresh identity plus a reconciliation row, never an error or a silent reuse).

## §D.6 Forward-looking checks

- Post-merge, the id allocator still cannot see across stores — until the root cause is repaired, new ids issue only from the merged high-water mark (this merge lifts it; a later allocator card must consult the mapping table).

## §D.7 Definition of Done

All ACs observed PASS with verbatim evidence in the verdict; mapping table + reconciliation + backup hashes committed under `.moai/reports/t657/`; both stores truthful; project store retired with fence marker; zero card loss.
