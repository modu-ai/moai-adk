# Plan — SPEC-TODO-ANALYZER-CONFORMANCE-001

## §A Context

Card t1311 (Class C, Tier S): two documentation/behavior conformance items from the 2026-09-29
todo-logic review §P5. The code is already conformant — `internal/kanban/backlog_store.go:141-158`
carries the third `jev` source constant and `internal/cli/todo_jev_finding.go` is its measured
emission site; the queue documentation (`.claude/skills/moai/workflows/gtd.md` + template mirror)
is not. Second item: the `spec_id` pick-time promise (gtd.md:252-253, both surfaces) contradicts
measured practice (9/9 empty; t472 axis D 53/53 persistent). The SPEC DECIDES option (b) —
abolish the promise, keep the column record-only (rationale: spec.md §A.3).

## §B Known Issues

- **B.1 Two-surface editing hazard.** The live skill doc and its template mirror carry the same
  sentences at the same line numbers today (both :252, both :258-259 — measured). A single-sided
  edit is the recurring failure shape (`todo_skill_doc_test.go` exists precisely because it
  happened before). Both surfaces are edited in the same milestone, never sequentially across
  milestones.
- **B.2 Mirror neutrality ceiling.** The natural wording "per SPEC-JEV-CONSUMERS-001" is
  FORBIDDEN in the mirror (the neutrality regex flags `SPEC-…-001`). Mirror wording references
  only user-visible surfaces (`workflow.jev.enabled`, `moai gtd relate`, `moai todo add`).
- **B.3 Conflation hazard.** The doc's existing `machine-only` explanation is written around the
  mechanical/agent pair. The `jev` clause must be additive: a `jev` finding neither claims the
  `machine-only` mark nor clears it, and the wording must not read `jev` as a subtype of `agent`.
- **B.4 Report path is primary-checkout-local.** `.moai/reports/todo-logic-review-20260929.md`
  exists in the primary checkout only (reports are local-only); queue figures are cited from it,
  not re-measured.

## §C Pre-flight

- Worktree `.moai/worktrees/t1311`, branch `WT-jev-enum-backfill`, HEAD `68e37864a` — clean at
  session start (measured). Isolated worktree, so the shared-checkout pre-edit probe is exempt;
  re-read HEAD immediately before the run-phase commit per the standing staleness rule.
- Baseline to cite in §E: `go test ./internal/cli/ -run 'TestTodoSkillDocument'` and
  `go test ./internal/kanban/ -run 'SchemaFreeze|JevFinding'` pass at plan HEAD (run before M1
  touches anything; record verbatim output).

## §D Constraints

- Documentation-only plus one new `_test.go`; zero production Go changes.
- Template-First: mirror edit + `make build` before verification; `agents-emit` not applicable
  (no agent file touched).
- No time estimates; milestones ordered by decision-reversibility — the decision-bearing wording
  (spec_id posture) lands first.
- Doc register: the gtd.md doc is English prose in the existing voice; match the surrounding
  bullet style (em-dash clauses, no emoji).

## §E Self-Verification

Run-phase §E evidence lands in progress.md §E.2 (owned by manager-develop). Plan-phase
verification obligations handed over:

1. AC-TAC-001/002 grep binary checks on both surfaces (verbatim output into §E.2).
2. AC-TAC-003 neutrality scan via the existing test.
3. AC-TAC-004 zero-match grep (exit 1 expected = pass).
4. AC-TAC-005 seeded-regression guard proof (scratch-copy revert → fail → restore → pass).
5. AC-TAC-006 affected-package tests + `git diff --name-only` scope proof.

## §F Milestones

### M1 (Priority High) — Abolish the spec_id pick-time promise (the decision surface)
Rewrite the `spec_id` bullet at gtd.md:252-253 on BOTH surfaces: record-only optional annotation;
empty-on-picked is normal; the `state` field is what distinguishes queued from picked; the
association's authoritative carriers are the dispatch fields, the commit-message card id, and the
evidence path. `next --spec` stays a supported record verb. Flip AC-TAC-004.
Files: `.claude/skills/moai/workflows/gtd.md`, `internal/template/templates/.claude/skills/moai/workflows/gtd.md`.

### M2 (Priority High) — Document the jev source value (both surfaces)
Extend the `findings` field description at gtd.md:258-262 on BOTH surfaces: enumerate
`mechanical` / `agent` / `jev`; add the jev clause per REQ-TAC-002 (admission-only, gate
`workflow.jev.enabled`, near-duplicate relation, model confidence rendered `p=`, record-only,
never a queue-mutation input, neither claims nor clears `machine-only`). Keep the mirror wording
neutral (§B.2). Flip AC-TAC-001/002/003.
Files: same two surfaces.

### M3 (Priority Medium) — Parity guard + re-measurement
Add a `todo_skill_doc_test.go`-patterned guard in `internal/cli` asserting (a) both surfaces
enumerate `jev` in the findings description, (b) neither carries "filled in when the item is
picked", (c) mirror neutrality scan still clean. Then run the §E batch, `make build` before the
test batch, and record verbatim outputs in progress.md §E.2. Flip AC-TAC-005/006.
Files: `internal/cli/<new>_test.go`.

## §G Anti-Patterns

- Editing only one of the two gtd.md surfaces (the failure B.1 names).
- Writing `jev` wording that implies a review happened (`agent`-conflation, REQ-JEVN-003 hazard).
- "Improving" the analyzer, dedup, or schema while in the file (scope discipline; §E Exclusions).
- Naming SPEC ids, REQ tokens, dates, or SHAs inside the mirror.

## §H Cross-References

- `.moai/reports/todo-logic-review-20260929.md` §P5 — origin analysis (primary-checkout-local).
- `internal/kanban/backlog_store.go:141-158` — the three source constants and their rationale.
- `internal/cli/todo_jev_finding.go` — the emission site (admission-only, gate, record-only).
- `internal/cli/todo_skill_doc_test.go` — the both-surface + neutrality test pattern M3 follows.
- SPEC-JEV-CONSUMERS-001 (REQ-JEVN-002/003/006, REQ-JEVC-011/012) — the jev consumer contract the
  doc wording must not weaken.
- SPEC-TODO-HOLD-STATE-001 — sibling P3 item; its two-surface doc discipline is the precedent.
