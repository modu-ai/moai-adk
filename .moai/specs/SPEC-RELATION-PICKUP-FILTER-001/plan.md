# SPEC-RELATION-PICKUP-FILTER-001 — Implementation Plan

cycle_type: **tdd** (`quality.yaml` `development_mode: tdd`; dispatch-mandated
recommendation confirmed — the change is dispatch-filter logic over the todo
store, RED-first is cheap and the RED states are already provable from the
measured code).

## §A Context

Card t1343 (proposal P3): consume the record-only `blocks`/`depends` relations
as a self-dispatch pickup filter, and refuse circular sequencing relations at
write time. Verified basis (all in this tree, HEAD `113082295`; full quotes in
spec.md §A.2):

- Pickup selection: `autoPickTargets` — `internal/cli/todo_auto.go:142`; the
  queued arm appends every `queued` card (`:162-166`), no relation consultation.
- Relation records: `BacklogFinding{SubjectID, RelatedID, Relation, ...}`,
  `internal/kanban/backlog_store.go:194-201`; sequencing values
  `blocks`/`depends` (`:139-148`), vocabulary list (`:177-184`).
- Relate write path: `runTodoRelate` — `internal/cli/todo_relate.go:60-94`;
  no cycle check (F-3).
- Resolution path: `ArchiveCard` moves findings naming the done card into the
  archive entry — `internal/kanban/backlog_store.go:308-341` (F-4).
- Cycle-guard pattern precedent (different subsystem, do not couple):
  `ValidateGTDRelation` — `internal/kanban/gtd_relation.go:48-86` (F-5).

Merge-vs-separate (dispatch-mandated decision): **stand alone**; recorded in
spec.md §A.4 with the four-point rationale (no t1338 SPEC exists; t1338's four
enumerated pieces exclude the relation filter; S-card must not queue behind a
Tier L candidate gated on the t1240 merge; t1338 piece 1 routes fallback lanes
into `todo --auto` and inherits this filter).

## §B Known Issues (measured, this tree)

- B-I-1 — The record-only doctrine comments
  (`internal/kanban/backlog_store.go:139-148`,
  `internal/cli/todo_relate.go:15-18`) assert no dispatch path reads the
  relations. After this SPEC they are false text and MUST be rewritten in the
  same change (spec.md B.5). Not doing so ships a stale doctrine.
- B-I-2 — `RemoveFindingsNaming` (`backlog_store.go:463`) has no production
  caller. NOT touched (scope discipline); named here so the run does not
  "helpfully" wire it into the done path — `ArchiveCard` already resolves
  findings correctly.
- B-I-3 — The no-eligible-card message (`todo_auto.go:218-219`) reads "queue is
  empty or every card is untouched-by-authority". With the filter, a fully
  blocked queue reaches that line with per-card notes printed above it — the
  message itself stays, the notes carry the why (M3).

## §C Pre-flight

- SPEC ID regex pre-write check: run as Bash —
  `[[ "SPEC-RELATION-PICKUP-FILTER-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`
  → **PASS** (verbatim output cited in progress.md §E.1).
- ID uniqueness: no `SPEC-RELATION-*` or `SPEC-PICKUP-*` directory exists under
  `.moai/specs/` (measured by directory listing).
- Frontmatter validated against the canonical 12-field schema SSOT
  (`.claude/rules/moai/development/spec-frontmatter-schema.md`); `status: draft`
  set at creation; snake_case aliases not used.
- Baseline: worktree clean at plan authoring; branch `WT-relation-pickup-filter`;
  artifact set = spec.md + plan.md + acceptance.md + progress.md (Tier S set
  plus the dispatcher-mandated acceptance layer).
- Run-phase entry preconditions: absorb local `develop` before the first
  run-phase commit; all tests use `t.TempDir()` stores — **no test may touch
  the operator's live queue** (`newTodoStore()` resolves the live record; tests
  construct `kanban.NewBacklogStore(t.TempDir()/...)` instead).

## §D Constraints

- D.1 — Touch only `internal/cli/todo_auto.go`, `internal/cli/todo_relate.go`,
  `internal/kanban/backlog_store.go` (comment block B-I-1 + a record-level
  helper if placed there), their `_test.go` files, and the two doctrine comment
  blocks. No other file.
- D.2 — Tier S budget: ≤ 8 REQ (7 used), ≤ 8 AC (7 used). No schema migration,
  no new CLI verbs or flags, no `homestate` changes.
- D.3 — GEARS only; no `IF/THEN` modality (lint would flag `LegacyEARSKeyword`).
- D.4 — The live operator queue is write-forbidden for this card at every phase.
- D.5 — Ordering by decision-reversibility: the "unresolved" predicate (B.1) is
  the semantic most likely to be revisited — it leads M1 and its test names
  spell the rule; mechanical comment edits are last (M3).

## §E Self-Verification (plan-phase obligations the run inherits)

- E-1 — Every AC in `acceptance.md` carries a RED-now statement grounded in the
  measured code (§A.2) and a green-path command naming package + test pattern +
  expected `ok` output. The run's first act per milestone is the RED test.
- E-2 — Direction semantics (REQ-RPF-002) get a two-sided test: a `blocks`
  finding must block its RELATED card, a `depends` finding its SUBJECT card.
  A single-direction implementation must fail M1's RED table, not pass it.
- E-3 — The cycle guard test must cover: direct 2-cycle (refused), 3-cycle
  through an intermediate card (refused), an open chain (allowed), and the
  same-pair opposite spelling (allowed — B.3).
- E-4 — Post-change `golangci-lint run ./internal/cli/... ./internal/kanban/...`
  and `go vet` on the two packages, then `go test -timeout 30m` on them (the
  AGENTS.local.md §4 timeout derivation; no local full-suite run).

## §F Milestones (priority-ordered; no time estimates)

- **M1 (High) — the pickup filter.** RED: a store holding a queued card with a
  sequencing finding still yields it from `autoPickTargets`. GREEN: a
  record-level blocked-side predicate (kanban package, finding-existence rule
  per B.1) consulted by the queued arm of `autoPickTargets` only. REQ-RPF-001,
  -002, -003; AC-RPF-001..003.
- **M2 (High) — the cycle guard.** RED: `runTodoRelate` records a 2-cycle
  today. GREEN: waits-on reachability check over recorded sequencing findings
  (blocks/depends only), refusing before `AppendFindingOnce` with an error
  naming both endpoints; record unchanged. REQ-RPF-005; AC-RPF-005..006.
- **M3 (Medium) — labelled skips + doctrine text.** Per-skip labelled
  non-finding (card id, relation, predecessor id); exit-0 contract pinned; the
  two record-only comment blocks rewritten to name the new consumer
  (B-I-1). REQ-RPF-004, -007; AC-RPF-004.
- **M4 (Medium) — regression pins.** Dead-owner rescue arm unaffected by the
  filter; a `contains`-only card stays eligible; same-pair opposite spelling
  not refused. REQ-RPF-006; AC-RPF-007.

## §G Anti-patterns (named for the run to refuse)

- Do NOT resolve "unresolved" by scanning card states — the finding's existence
  IS the predicate (B.1). A state scan re-opens the dropped/hold semantics this
  SPEC deliberately does not touch.
- Do NOT gate the dead-owner rescue arm (B.2).
- Do NOT couple to `gtd_relations` or reuse `ValidateGTDRelation` itself — the
  data models differ (F-5); mirror the pattern, not the function.
- Do NOT add flags/verbs/config. The filter is unconditional on the `--auto`
  path, exactly like the positive-state predicate precedent (REQ-THS-012).
- Do NOT let any test read or write the operator's live queue (§C pre-flight).
- Do NOT edit `.moai/docs/` or templates — docs land at sync-phase.

## §H Cross-references

- SPEC-TODO-ANALYSIS-001 — the relate/unrelate verbs and the record-only
  posture this SPEC first consumes.
- SPEC-TODO-HOLD-STATE-001 — the positive-state-vocabulary predicate precedent
  (REQ-THS-012) and the hold state the filter must not mistake for resolution.
- SPEC-TODO-CLASSIFY-DISPATCH-001 — the lane self-pickup consumption precedent
  (t1332) and the mode-aware lease this SPEC deliberately stays beside.
- card t1309 — landed blocks/depends as record-only; its doctrine comments are
  this SPEC's B-I-1 edit targets.
- card t1240 — the named adjudication owner for factory-lease-side consumption.
- card t1338 — the recorded convergence consumer (spec.md §A.4).
- proposal P3 — `.moai/reports/autonomy-bottleneck-proposal-20260929.html`
  (primary checkout, local-only) — originating scope and constraints.
