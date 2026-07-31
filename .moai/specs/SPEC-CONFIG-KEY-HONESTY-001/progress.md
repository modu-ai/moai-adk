# SPEC-CONFIG-KEY-HONESTY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-CONFIG-TIER-PERSIST-001]
code_baseline: d5336214e
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.72
    threshold: 0.80
    dimensions: {clarity: 0.78, completeness: 0.80, testability: 0.62, traceability: 0.72}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7, D8]
    deferred: [D9, D10, D11, D12, D13, D14, D15]
    report: .moai/reports/plan-audit/SPEC-CONFIG-KEY-HONESTY-001.md
```

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- Code baseline `d5336214e`; the worktree HEAD is a descendant on branch
  `plan/epic-update-config-audit` that changes SPEC documents only
  (`git diff --name-only d5336214e HEAD | grep -v '\.md$'` → 0 lines), so every `file:line` and
  count in these artifacts is attributable to the code baseline.
- Findings F1-F7 each re-verified against this tree while authoring; one drift recorded
  (spec.md §A.8 — shipped `workflow.yaml` worktree toggles contradict `internal/config/defaults.go`).
- **F3 re-derived path-resolved at the plan-audit revision** (D3): 287 distinct `yaml:`-tagged field
  names, **174** with zero production reads and 4 accessor-only; 161 map to a shipped key across 188
  (file, key) occurrences. This supersedes the earlier bare-field-name figures (122 / 5 / 121),
  which were produced by the very method this SPEC forbids as AP-3. The re-derivation used a
  throwaway `go/packages` selector-resolution probe over `./internal/... ./pkg/... ./cmd/...`
  (106 packages, 0 type errors); M2's guard is its durable implementation.
- The recomputation resolved the §A.3 ↔ §A.6 contradiction rather than papering over it:
  `auto_merge` is dead under path resolution and now appears in the family table, matching §A.6's
  "AutoMerge has zero reads". 43 names flip live→dead in total.
- Prose-consumer discriminator measured: dotted-path fixed-string probe yields 0-1 hits per key
  versus up to 46 for the bare leaf key.
- SPEC ID regex self-check executed: `PASS`.
- Status: `draft`. Awaiting plan-audit re-run and Implementation Kickoff Approval.

### Epic run order (dependency sequencing)

`depends_on: [SPEC-CONFIG-TIER-PERSIST-001]` records that this SPEC reads the tier-resolution
contract E3 owns — a shipped key's liveness is judged against the loader that actually resolves it,
and E3 fixes which tier that is. The edge is a **read** edge: nothing in M1-M6 writes a surface E3
also writes.

The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
`status: completed`. Every SPEC in this Epic is currently `draft`, so entering `/moai run` on this
SPEC before E3 closes raises the 3-option wait / override / abort blocker. **The dependency is
satisfied by sequencing, not by an `--ignore-deps` bypass** — the run order below is the mechanism,
and it is consistent with the orders recorded in `SPEC-UPDATE-DATA-SURVIVAL-001` §E.1 and
`SPEC-CONFIG-TIER-PERSIST-001` §E.1:

| Order | SPEC | Gate to clear before the next entry |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | reaches `status: completed` — REQ-RIL2-015/016 landed |
| 2 | `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) | reaches `status: completed` — backup coverage + failure contract landed |
| 3 | `SPEC-CONFIG-TIER-PERSIST-001` (E3) | reaches `status: completed` — tier precedence + atomic write landed |
| 4 | **`SPEC-CONFIG-KEY-HONESTY-001`** (this SPEC) | — |
| 5+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-UPDATE-CI-GUARD-001`, `SPEC-UPDATE-DOC-DRIFT-001`) | no `depends_on` edge to this SPEC |

Do **not** invoke `/moai run` on this SPEC with `--ignore-deps`. If E3 slips and starting this SPEC
early becomes necessary, the correct move is to run **M1, M2, M4, M5, and M6** — none reads E3's
tier contract — and hold **M3** open, since `quality.yaml`'s parse path is the one place where which
tier resolved the block changes the answer. That is a scope decision for the orchestrator to surface
via `AskUserQuestion`, not a flag the run-phase agent may set on its own.

One ordering constraint is internal to this SPEC and independent of the Epic order: M1 must land
before M2, because M2's guard fails on any `dead` / `unresolved` / `unbound` key absent from M1's
**P** / **R** allowlists — with no inventory, every shipped key fails at once.

### Deferred audit defects (D9-D15)

Recorded so the next iteration does not re-derive them. D9 (`qualityFileWrapper` cited at
`internal/config/types.go:1174`, whereas `parseFullQualityConfig` in package `hook` uses the
same-named type at `internal/lsp/hook/gate.go:97` — the §A.1 conclusion is unaffected, the citation
is not), D10 (§A.6's grep transcript is pre-filtered; §A.6 now says so in prose but the command is
not yet replaced with a reproducing one), D11 (NFR-CKH-002's 200-key floor against ~1020 shipped
keys), D12 (AC-CKH-016 counts tokens rather than asserting meaning), D13 (`<handoff-note-path>`
placeholder), D14 (plan.md §B2's `main-fork/` premise is false in this worktree — it exists only in
the primary checkout, so AP-4's falsification needs a synthetic fixture), D15 (`depends_on` target
in `draft` — addressed by the run order above rather than by frontmatter change).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
