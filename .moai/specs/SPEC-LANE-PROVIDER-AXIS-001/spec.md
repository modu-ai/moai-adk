---
id: SPEC-LANE-PROVIDER-AXIS-001
title: "Lane axis — the factory lane pool crossed with the provider axis"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: "internal/web (factory lane panel)"
lifecycle: spec-anchored
tags: "factory-lanes, provider-axis, web-console, regression-guard, lane-pool"
tier: M
era: V3R6
related_specs: [SPEC-MODEL-MATRIX-CORE-001, SPEC-MODEL-MATRIX-SURFACES-001, SPEC-WEB-CONSOLE-015]
---

# SPEC-LANE-PROVIDER-AXIS-001 — Lane axis

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Stood up as its own SPEC (card t1047, operator decision). Card t843 proposed a lane/role axis; that card was split into four `SPEC-MODEL-MATRIX-*` successors and the lane axis landed in none of them — `factory_lanes` / `kanban` / `launcher` return zero hits across all four successor directories, against a positive control that fires in all four. Carries the three items card t1031's absorption inventory §B ruled live: the lane pool axis, the provider spread display, and the one-provider-per-lane invariant as a regression guard. Tier M assigned on file count (6 files, inside the 5-15 band); the LOC estimate sits below the Tier S band, so the file count is what selects M, and M's artifact set is exactly the three files authored here plus `progress.md`. Status `draft`: nothing in this SPEC has landed. | manager-spec |

## Position

This SPEC is a **root**. It has no `depends_on` and shares no file with any `SPEC-MODEL-MATRIX-*` successor.

The lane axis was proposed as a crossing with a *backend multiplexing* axis that no longer exists (the gpt gateway was withdrawn by operator goal 2026-09-16; `internal/gateway` is absent entirely). The crossing partner is therefore chosen afresh here, and it is **the provider axis already carried on the lane row** (`LaneVM.Backend`) — not the profile axis and not the agent axis, both of which `SPEC-MODEL-MATRIX-CORE-001` owns.

---

## §A Context

### §A.1 The lane pool has rows but no pool

`internal/web/factory_lanes.go` builds one `LaneVM` per registered factory lane and `internal/web/screens.templ` renders them as a row list under a `Factory lanes` panel whose only aggregate is `len(k.Lanes)` — a count. Each row already carries a provider (`@backendBadge(l.Backend)` at `screens.templ:243`), so the *crossing* of lane and provider is present per row and absent in aggregate: a reader counting badges by eye is doing the aggregation the panel does not do.

### §A.2 The provider axis has two reachable values and a third vestigial constant

`internal/kanban/record.go:22-24` declares three provider constants — `BackendClaude = "claude"`, `BackendGLM = "glm"`, `BackendGPT = "gpt"`. Measured in this tree, `BackendGPT` has **zero live Go references** outside its own declaration and two doc comments (`record.go:77`, `internal/config/envkeys.go:225`), while `BackendClaude` and `BackendGLM` are referenced across many live files. The `"gpt"` string literals that do appear in `internal/cli/init.go` and `internal/cli/doctor.go` belong to the `agentWiring` / harness-selection axis, which is a different axis than the kanban provider.

So the axis a display must be sound for is **two-valued in what can actually appear**, with a vestigial third constant that a naive `switch` over the constant block would enumerate as a column. A design that only reads well at four or more providers is a defect here, and a design that presents `gpt` as an occupied bucket is asserting a state no writer can produce.

The existing `backendBadge` (`internal/web/widgets.templ:69`) already encodes the two-value assumption in its else-branch: `glm` renders `metered`, **everything else** renders `flat rate`. That is correct at two values and silently wrong at three. It is named here as observed context, not as a repair this SPEC owns.

### §A.3 The one-provider-per-lane invariant is not enforced by any check

`LaneVM.Backend` is a scalar `string` (`internal/web/factory_lanes.go:43`) assigned exactly once, from a session-joined kanban record (`internal/web/factory_lanes.go:150`). Those are the only two lines in the repository where the field appears; the field-selector assignment measured across `internal/web/` has exactly one hit, against a positive control on the sibling `CardID` field that fires twice.

The invariant therefore holds **structurally** — the type cannot hold two providers and no second writer exists — and **nothing asserts it**. The work is not to implement a constraint that already holds; it is to attach a regression guard so a later widening cannot pass silently.

---

## §B Requirements (GEARS)

### §B.1 The lane pool axis

- **REQ-LPA-001** (Ubiquitous) The factory-lane view model shall expose the lane pool as an aggregate over its registered lanes, in addition to the per-lane row list.
- **REQ-LPA-002** (Unwanted) The lane pool aggregate shall not read, import, or reference the profile axis or the agent axis.

### §B.2 The provider spread display

- **REQ-LPA-003** (Ubiquitous) The factory lane panel shall display the provider spread across the registered lane pool.
- **REQ-LPA-004** (Ubiquitous) The spread shall carry a distinct bucket for lanes whose provider is unrecorded, and the buckets shall partition the pool — every registered lane shall fall in exactly one bucket.
- **REQ-LPA-005** (Unwanted) The spread shall not present a provider value that no live writer can produce as an occupied bucket.
- **REQ-LPA-006** (Ubiquitous) Every i18n key the spread introduces shall resolve in all four locales.

### §B.3 The one-provider-per-lane regression guard

- **REQ-LPA-007** (Event-detected) **When** the lane row's provider field is widened beyond a single scalar, or gains a second assignment site, the regression guard shall fail.

---

## §C Exclusions

### Out of Scope — the profile matrix and its owning SPEC

- `internal/template/profile_matrix.go` is **not touched**, and its semantics are **not redefined**. That file is owned by `SPEC-MODEL-MATRIX-CORE-001`, whose documents were updated at `159dd30df` (card t1037). Nothing in this SPEC reads it, imports it, or depends on its cell shape.
- The profile axis (`max` / `medium` / `low`) and the agent axis are excluded as crossing partners for the lane pool, because both are owned by that SPEC. REQ-LPA-002 states the exclusion as an enforceable requirement rather than leaving it to good intentions.
- The axis-vocabulary discrepancy inside `profile_matrix.go` (its header comment and its later cell-table comments name different vocabularies) is observed and **not adjudicated here** — it belongs to `SPEC-MODEL-MATRIX-CORE-001`. It is a cross-reference note only and is not a requirement of this SPEC.

### Out of Scope — the three-preset concept

- The quality / balance / economy preset triple is excluded. The discriminant is **"does it redefine `profile_matrix.go` semantics?"**, and the answer is yes: measured, the profile axis is keyed by **agent** with cell value `config.ModelEffort{Model, Effort}`, while a preset is keyed by **lane** and additionally carries the launcher. The `(model, effort)` half of a preset is therefore already computed per agent by a profile column, so defining a preset means deciding how a lane-keyed thing relates to an agent-keyed thing — which redefines the semantics of a file this SPEC may not touch.
- This exclusion is **not** "not yet measured". The measurement is done; re-measuring does not change the answer, because what would have to change is ownership, not evidence. A later card that wants presets starts by moving or widening `SPEC-MODEL-MATRIX-CORE-001`'s scope, not by re-running a grep.

### Out of Scope — three t843 items whose targets do not exist

These are **not deferred work**. There is nothing to build against, so they are excluded rather than scheduled.

- The `agents × tier` starting axis. That axis is retired — `tierProfiles` has zero live code lines, its only hit being a retirement-record comment.
- Backend multiplexing. The gpt gateway was withdrawn by operator goal 2026-09-16 and `internal/gateway` is absent entirely.
- The t837 availability slots. Card t837 was dropped for the same withdrawal.

### Out of Scope — two axes whose inputs are missing

- The `$/task` axis and the DeepSWE quality bands. `model_catalog.yaml` does not exist in either the live tree or the template tree, and `SPEC-MODEL-CATALOG-SSOT-002` is still `status: draft`. Both inputs are absent, so these are waiting items rather than design surface, and they are not revived here.

### Out of Scope — adjacent surfaces this SPEC deliberately does not touch

- `internal/cli/launcher.go`, `internal/cli/kanban.go`, `internal/cli/factory.go`, and `internal/cli/cc.go`. Measured, the CLI carries the lane concept as **parse and launch logic** — label parsing, env export, GLM resolution, a mixed-backend rejection sentinel — and renders no per-lane provider display. The spread is a console surface; adding a CLI one is a separate decision.
- `backendBadge`'s two-value else-branch (`internal/web/widgets.templ:69-82`). Named in §A.2 as observed context. Repairing it changes a widget shared by three panels (role row, lane row, monitor row) and is not this SPEC's concern.
- The lane join itself — the count-then-resolve logic, the four unresolved reasons, and the read-only property asserted by `TestFactoryLanesJoinWritesNothing`. This SPEC adds an aggregate over the rows the join already produces and changes no resolution outcome.
- `internal/kanban/record.go`. The vestigial `BackendGPT` constant is measured and reported; removing it is a separate card with its own blast radius.

---

## §D Constraints

| # | Constraint | Source |
|---|---|---|
| C-1 | The lane panel is read-only. Nothing added here may write to the factory registry, the session registry, or a kanban record. | `internal/web/factory_lanes.go` header; `TestFactoryLanesJoinWritesNothing` |
| C-2 | 4-locale parity is mandatory: no i18n key may exist in a subset of the four locale blocks of `internal/web/assets/i18n.js`. | project i18n practice + REQ-LPA-006 |
| C-3 | `*_templ.go` files are generated by `templ generate` and are never hand-edited. | templ toolchain |
| C-4 | An unresolvable lane is **present** with a marker, never dropped. The pool aggregate must count it, because a lane missing from the aggregate is indistinguishable from a lane that never ran. | `internal/web/factory_lanes.go` `LaneVM` doc comment |
| C-5 | A parallel session may be active on this shared checkout; changes are committed with explicit pathspecs. | project operating practice |

---

## §E Decisions

| # | Decision | Rationale |
|---|---|---|
| D-1 | The crossing partner for the lane pool is the **provider** axis, not profile and not agent. | The provider is already on the lane row, so the crossing needs no new data source and no new owner. Profile and agent are owned elsewhere (§C). |
| D-2 | The invariant ships as a **regression guard**, not as an enforcement check. | It already holds structurally (§A.3). Implementing enforcement would duplicate a property the type has, and a duplicated check is a second thing to keep true. |
| D-3 | The unrecorded-provider case is a **named bucket**, not folded into a provider. | An unresolved lane carries an empty `Backend`. Folding it into `claude` would render a confident wrong count, which is exactly what the lane section's join logic already refuses to do for rows. |
| D-4 | The spread is designed for two reachable values; the vestigial third constant is excluded from the bucket set by measurement of live writers, not by copying the constant block. | Enumerating `kanban.Backend*` would present `gpt` as a column no writer can fill. |
| D-5 | The display lands on the web console only. | Measured: the CLI renders no per-lane provider (§C). Adding one is a separate decision, not an implication of this one. |

---

## §F Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | The regression guard is written so that both of its mutations are caught by the compiler instead of by the guard, making it a check that cannot fail for the reason it exists. | High | AC-LPA-007 requires the mutation to be the **coherent** widening (field and assignment changed together); an incoherent half-edit breaks the build and proves nothing about the guard. |
| R-2 | The bucket set is derived by enumerating `kanban.Backend*`, reintroducing `gpt` as a column. | Medium | REQ-LPA-005 plus AC-LPA-005, which asserts the rendered bucket set against reachable values and asserts `gpt` absent. |
| R-3 | The aggregate drops unresolved lanes, so the buckets no longer sum to the pool. | Medium | REQ-LPA-004 states the partition property; AC-LPA-004 asserts the sum equals the lane count. |
| R-4 | New i18n keys land in one locale block and the other three are deferred. | Medium | C-2 plus AC-LPA-006, following the existing `TestSPECIntroducedKeysResolveInEveryLocale` pattern in `internal/web/factory_lane_section_test.go`. |
| R-5 | Scope creeps into `profile_matrix.go` through a "while we are here" reading of the pool axis. | Medium | REQ-LPA-002 makes the exclusion a mechanically checked requirement (AC-LPA-002), not a prose boundary. The preceding card's verdict recorded that an unstated scope boundary is what lets the next person cross it. |

---

## §G Traceability

| Unit | REQ range | AC range |
|---|---|---|
| M1 — lane pool axis | REQ-LPA-001, 002 | AC-LPA-001, 002 |
| M2 — provider spread display | REQ-LPA-003 … 006 | AC-LPA-003 … 006 |
| M3 — regression guard | REQ-LPA-007 | AC-LPA-007, 008 |

7 requirements, 8 acceptance criteria — both within the Tier M ceilings of 16 and 16. Full AC↔REQ mapping: `acceptance.md` §D.

---

## §H Cross-References

- `.moai/reports/t1047/premeasure.md` — the pre-authoring re-measurement in this tree: absence of an owning successor, the invariant-enforcement measurement, the preset/profile semantic comparison.
- `.moai/specs/SPEC-MODEL-MATRIX-CORE-001/` — owns `internal/template/profile_matrix.go`, the profile axis, and the unadjudicated axis-vocabulary discrepancy inside that file.
- `.moai/specs/SPEC-MODEL-MATRIX-SURFACES-001/` — the successor whose `§ Out of Scope` names "the `moai web` console's non-agentfm panels", which is exactly the panel this SPEC works on. The two do not overlap.
- `.moai/specs/SPEC-MODEL-CATALOG-SSOT-002/` — the `draft` SPEC whose absent `model_catalog.yaml` is one of the two missing inputs in §C.
