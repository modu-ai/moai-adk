---
id: SPEC-DEADPKG-INVESTIGATE-001
title: "Investigate 5 flagged dead internal packages and dispose per evidence-backed verdict"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/design, internal/research, internal/runtime, internal/migrate, internal/i18n"
lifecycle: spec-anchored
tags: "dead-code, cleanup, investigation, verification, tech-debt"
tier: M
era: V3R6
related_specs: [SPEC-WF-AUDIT-GATE-001, SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001, SPEC-DEAD-CONFIG-001]
---

# SPEC-DEADPKG-INVESTIGATE-001 — Investigate 5 Flagged Dead Internal Packages

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-02 | manager-spec | Initial plan-phase authoring. Per-package investigation protocol; deletion is conditional on evidence-backed verdict, never prescribed up front. |

## §A. Context and Background

An internal audit flagged 5 packages under `internal/` as candidate "dead code" (zero
external production importers). Re-verification during plan-phase confirmed the raw
reachability counts but also surfaced that **most of the flagged packages carry evidence
of intentional retention or staging** — documented `@MX:REASON` markers, an owning SPEC
whose status is `implemented`, or a stale annotation that contradicts the grep. Blind
deletion is therefore unsafe.

This mirrors a documented failure class: the era=NONE 29-SPEC incident
(`.claude/rules/moai/core/verification-claim-integrity.md` §5), where 29 SPECs with an
absent `era:` field and `status: implemented` were nearly batch-closed as "Mx-close debt"
but were in fact grandfather-protected — the dedicated tool (`moai spec audit`)
disconfirmed the defect claim. The lesson codified there is load-bearing here:

> A defect claim is a **hypothesis** until the domain's tool confirms it.
> Absence of evidence (no importer) is not, by itself, evidence of death.

"Dead package" is a defect claim (verification-claim-integrity §1.1 surface 3). It is a
hypothesis until the run-phase investigation runs the reachability + intent + retention +
owning-SPEC checks and records the observed output. This SPEC defines that per-package
investigation protocol and requires an evidence-backed verdict **before any removal**.

### §A.1 Plan-phase reachability baseline (observed, not assumed)

The following counts were observed at plan-phase against the working tree
(module `github.com/modu-ai/moai-adk`, go 1.26.4). They are the **starting hypothesis**,
to be re-confirmed at run-phase per package:

| Package | Non-test `.go` files | External production importers (observed) | Notes |
|---------|---------------------|------------------------------------------|-------|
| `internal/design` | 19 | 0 (`grep` outside `internal/design/` empty) | DTCG design-token validation (`dtcg/` + `pipeline/`). Distinct from `internal/config/loader_design.go` (loads design.yaml — unrelated). |
| `internal/research` | 17 | 0 | A/B experiment engine (`experiment/ eval/ observe/ safety/ dashboard/`). `safety/limiter.go` `@MX:REASON` claims CLI+auto-update use it — **contradicted** by grep. |
| `internal/runtime` (top-level) | 8 | 0 production (only `budget_test.go` + 5 `cli/*_test.go`) | SPEC-WF-AUDIT-GATE-001 audit-gate. `internal/runtime/gobin` subpackage is **LIVE** (imported by `cli/update.go` + `core/project/initializer.go`) — MUST NOT be touched. |
| `internal/migrate` | 1 (`hook_cleanup.go` `CleanupUserSettings`) | 0 callers | `internal/hook/retired_events.go` documents `@MX:REASON: fan_in=3 ... consumed by ... migrate.CleanupUserSettings + future observability tooling`. Overlaps LIVE `internal/migration` (7-file framework). |
| `internal/i18n` | 2 (`errors.go`, `templates.go`) | 0 production (imported only by `internal/github/*_test.go`) | Provides `NewCommentGenerator` / `CommentData` / `Generate` (multilingual GitHub issue-close comments). |

## §B. Requirements (GEARS)

### §B.1 Verdict discipline

- **REQ-DPI-001** (Ubiquitous): The investigation protocol **shall** assign each of the 5
  flagged packages exactly one of three verdicts before any removal —
  **(a) GENUINELY DEAD**, **(b) INTENTIONALLY RETAINED**, or **(c) RELOCATE**.
- **REQ-DPI-002** (Unwanted behavior): The run-phase agent **shall not** remove any package
  (or any of its symbols) without that package's verdict **and its supporting evidence**
  recorded in `progress.md` §E.2.
- **REQ-DPI-003** (Ubiquitous): The SPEC **shall not** prescribe deletion up front; deletion
  **shall** be conditional on the per-package verdict resolving to GENUINELY DEAD (or the
  removal-of-original half of a RELOCATE).

### §B.2 Per-package investigation evidence (four mandatory checks)

- **REQ-DPI-004** (Event-driven): **When** investigating a package, the run-phase agent
  **shall** run a **reachability grep from `cmd/moai`** (the production entrypoint) and record
  whether any non-test production path transitively reaches the package.
- **REQ-DPI-005** (Event-driven): **When** investigating a package, the run-phase agent
  **shall** run a **`git log` intent check** on the package path and record the most-recent
  commit + its SPEC provenance (staging intent, retirement intent, or none).
- **REQ-DPI-006** (Event-driven): **When** investigating a package, the run-phase agent
  **shall** run an **`@MX:REASON` scan** across the package and its referents, and **shall**
  treat any `@MX:REASON` retention claim as a hypothesis to be **confirmed or disconfirmed**
  against the reachability grep (per verification-claim-integrity §1.1 surface 3).
- **REQ-DPI-007** (Event-driven): **When** investigating a package, the run-phase agent
  **shall** check for an **owning SPEC** (e.g. `SPEC-WF-AUDIT-GATE-001` for `internal/runtime`)
  and record its `status`. **Where** an owning SPEC exists with `status ∈ {planned, in-progress,
  implemented}`, the package **shall** default to INTENTIONALLY RETAINED unless the agent
  records disconfirming evidence that the owning SPEC no longer intends the package to exist.

### §B.3 Disposition actions

- **REQ-DPI-008** (State-driven): **While** a package's verdict is INTENTIONALLY RETAINED, the
  run-phase agent **shall** add a clear, greppable retention marker (e.g. an `@MX:NOTE` or a
  package-doc comment) recording *why* it is retained and *what* would make it dead, and
  **shall not** remove the package.
- **REQ-DPI-009** (State-driven): **While** a package's verdict is RELOCATE, the run-phase agent
  **shall** move the still-needed symbol to its correct home (e.g.
  `migrate.CleanupUserSettings` → a migration step under `internal/migration`; `i18n` →
  a `internal/github` test helper) and update all call sites, then remove the emptied original
  package **in the same milestone**.
- **REQ-DPI-010** (State-driven): **While** a package's verdict is GENUINELY DEAD, the run-phase
  agent **shall** remove the package **together with its `_test.go` files in the same milestone**,
  and **shall** remove any now-orphaned `@MX:REASON` referent (e.g. the retention line in
  `internal/hook/retired_events.go`) so no dangling reference survives.

### §B.4 Safety invariants

- **REQ-DPI-011** (Unwanted behavior): The run-phase agent **shall not** modify, move, or remove
  `internal/runtime/gobin` — it is LIVE (imported by `cli/update.go` + `core/project/initializer.go`).
- **REQ-DPI-012** (Ubiquitous): After every milestone, the cross-platform build **shall** remain
  green (`go build ./...` and the project's cross-platform matrix), and the full test suite
  (`go test ./...`) **shall** pass.
- **REQ-DPI-013** (Ubiquitous): The disposition of `internal/migrate` **shall** preserve the
  `@MX:WARN` atomic-write safety contract in `hook_cleanup.go` (P0-4 regression guard) — whether
  the code is retained, relocated, or removed, the atomic-write behavior and its regression test
  **shall not** be silently dropped.

## §C. Per-Package Preliminary Disposition Hypotheses (to confirm in run-phase)

These are **hypotheses**, not verdicts. Each MUST be confirmed or overturned by the run-phase
evidence (REQ-DPI-004..007) and recorded in `progress.md` §E.2.

| Package | Preliminary hypothesis | Decisive run-phase question |
|---------|------------------------|------------------------------|
| `internal/design` | **Leans GENUINELY DEAD** | Did `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` (completed, "full lifecycle removal") intend the Go DTCG validator to survive the skill retirement? Do any live rules/config still consume DTCG validation? If none → GENUINELY DEAD. |
| `internal/research` | **Leans GENUINELY DEAD (stale retention claim to disconfirm)** | The `safety/limiter.go` `@MX:REASON` claims CLI+auto-update use `NewRateLimiter`; grep shows zero external importers. Disconfirm the claim across the whole repo; check for a staged A/B rollout SPEC. If no live consumer and no staged SPEC → GENUINELY DEAD. |
| `internal/runtime` (top-level) | **Leans INTENTIONALLY RETAINED (staged subsystem)** | Owning `SPEC-WF-AUDIT-GATE-001` is `implemented`. Is the audit-gate wired into a production path, or staged awaiting wiring? Disambiguate the textual `GateConfig`/`AuditGate` matches in `config/`/`lsp/`/`hook/` (likely a different gate). If staged-with-owning-SPEC → RETAINED + marker; if confirmed superseded/never-to-wire → DEAD. |
| `internal/migrate` | **Leans RELOCATE** | Is "future observability tooling" (the `@MX:REASON` in `retired_events.go`) a genuine staged plan or aspirational? If aspirational → RELOCATE `CleanupUserSettings` into `internal/migration` as a migration step (preserving atomic-write safety, REQ-DPI-013). If genuine staged plan → RETAINED + marker. |
| `internal/i18n` | **Leans RELOCATE** | Should `internal/github` production code call `NewCommentGenerator` (a wiring gap → out of scope, keep+flag) or is it a test-only helper? If test-only → RELOCATE into a `internal/github` test helper. If a genuine production wiring gap → RETAINED + flag the gap as a follow-up SPEC. |

## §D. Exclusions

This section records what is **out of scope** for this SPEC. Each excluded item is expressed as
an `### Out of Scope —` H3 sub-heading so the `OutOfScopeRule` lint is satisfied.

### Out of Scope — internal/runtime/gobin

- `internal/runtime/gobin` is LIVE (imported by `cli/update.go` + `core/project/initializer.go`).
  It is explicitly protected (REQ-DPI-011) and MUST NOT be investigated for removal, moved, or
  edited by this SPEC.

### Out of Scope — new feature wiring

- If investigation reveals that a package is dead only because a *production wiring path was
  never built* (e.g. `internal/github` should call `i18n.NewCommentGenerator` but does not, or
  the audit-gate should be invoked from `cmd/moai run` but is not), building that wiring is a
  **separate feature SPEC**. This SPEC records the gap and flags a follow-up; it does not
  implement new production integration.

### Out of Scope — internal/migration and internal/config/loader_design.go

- `internal/migration` (the LIVE 7-file migration framework) and
  `internal/config/loader_design.go` (loads `design.yaml`) are NOT flagged packages and are NOT
  investigated for removal. `internal/migration` is only a *relocation target* for a possible
  `internal/migrate` RELOCATE verdict.

### Out of Scope — audit-tool authoring

- This SPEC does not build or modify any dead-code detection tooling (e.g. a `moai spec audit`
  extension). It applies the existing manual investigation checks (grep, `git log`, `@MX` scan,
  SPEC catalog lookup) per package.

## §E. Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §5 — defect claims
  are hypotheses until the domain tool confirms them (the governing doctrine for this SPEC).
- `SPEC-WF-AUDIT-GATE-001` (`status: implemented`) — owning SPEC for `internal/runtime`.
- `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` (`status: completed`) — design-system skill retirement,
  relevant to `internal/design`.
- `SPEC-DEAD-CONFIG-001` (`status: completed`) — sibling dead-code removal precedent
  (dead `runtime.yaml` config section).
