---
id: SPEC-PROGRESS-MARKER-CANON-001
title: "Canonicalize progress.md §E section markers (convention B) + era.go comment correction"
version: "0.1.0"
status: in-progress
created: 2026-06-16
updated: 2026-06-16
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/workflow + internal/spec + .claude/agents/moai"
lifecycle: spec-anchored
tags: "progress-md, era-classification, documentation-canon, marker-convention, lint"
era: V3R6
---

# SPEC-PROGRESS-MARKER-CANON-001 — progress.md §E Marker Canonicalization

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-16 | manager-spec | Initial plan-phase artifacts. Resolve §E.2 meaning conflict between `lifecycle-sync-gate.md` (convention A: §E.2 = "Sync") and `manager-develop.md`/`manager-docs.md`/`manager-spec.md` (convention B: §E.2 = "Run Evidence", §E.4 = "Sync"). User locked convention B as canonical. |

## A. Problem & Why

The MoAI SPEC lifecycle records a per-SPEC `progress.md` whose `§E.*` section markers feed the era-classification engine `internal/spec/era.go` `ClassifyEra`. The engine's marker detection (`hasProgressMarker`, era.go L161) is a plain `strings.Contains` — it tests STRING PRESENCE ONLY of the literal substrings `§E.2` and `§E.5`, never their semantic meaning. Despite that, two canonical authoring documents DISAGREE on what `§E.2` *means*, and real SPECs have drifted into three conventions:

- **Convention A** (`lifecycle-sync-gate.md` worked example, L297/L300): `§E.2 Sync-phase Audit-Ready Signal` / `§E.5 Mx-phase Audit-Ready Signal` — §E.2 means "Sync".
- **Convention B** (`manager-develop.md` L283/L284, `manager-docs.md` L136, `manager-spec.md` L274 — the agent-body ownership matrices): `§E.1 Plan` / `§E.2 Run-phase Evidence` / `§E.3 Run-phase Audit-Ready Signal` / `§E.4 Sync-phase Audit-Ready Signal` / `§E.5 Mx-phase Audit-Ready Signal` — §E.2 means "Run Evidence"; Sync lives at §E.4.
- **§F.* drift** (16 real completed SPECs, e.g. SPEC-EVIDENCE-CLAIM-INVARIANT-001): `§F.1 Plan` / `§F.2 Run` / `§F.3 Sync` / `§F.4 Mx` — these markers are NOT recognized by era.go (it greps `§E.*`), producing an H-2 V3R2-R4 INFO misclassification.

This SPEC canonicalizes the documentation to **convention B** (the user's locked decision), corrects the misleading "§E.2 = sync" comments in `era.go` (comment-only; grep logic byte-identical), and adds a forward-looking progress.md skeleton-generation instruction to `manager-spec.md` so future SPECs emit canonical §E markers instead of ad-hoc-drifting into §F.*.

WHY this matters: the §E.2-means-sync vs §E.2-means-run-evidence conflict is a documentation-truth defect. An author reading `lifecycle-sync-gate.md` learns §E.2 = Sync; an agent reading its own ownership matrix learns §E.2 = Run Evidence. The two are mutually exclusive descriptions of the same marker. Resolving the conflict prevents future progress.md files from being authored against the wrong section map.

## B. Scope

### Out of Scope — era.go classification behavior

- The era.go grep logic (`hasSyncSection := hasProgressMarker(content, "§E.2")`, L110; `hasMxSection`, L111; H-2/H-3/H-4 branches, L116-127; `hasProgressMarker` plain `strings.Contains`, L161) MUST remain byte-identical. This SPEC corrects COMMENTS ONLY (L33, L92, and the const-block / doc-comment prose that calls §E.2 "sync"). The classification result for every existing SPEC is unchanged. No test behavior changes.

### Out of Scope — retroactive SPEC progress.md edits

- The 16 existing SPECs that use §F.1/§F.2/§F.3/§F.4 markers are `era_final: true` history per `lifecycle-sync-gate.md` § Grandfather Clause Policy. This SPEC MUST NOT retro-edit any existing SPEC's `progress.md`. The skeleton-generation instruction is forward-looking only (applies to SPECs created after this SPEC merges).

### Out of Scope — era.go variable renaming

- Renaming the misnomer variables `hasSyncSection` / `hasMxSection` to semantically accurate names is NOT in scope (it would be a behavior-preserving refactor but expands the diff beyond the three locked items). Comments clarify the misnomer; the identifiers stay.

### Out of Scope — lint rule for §E marker presence

- Adding a new spec-lint rule that enforces the canonical §E section map at lint-time is NOT in scope. This SPEC delivers documentation canonicalization + a generator instruction; mechanical lint enforcement is deferred to a potential follow-up SPEC.

### Out of Scope — heuristic table re-wording in lifecycle-sync-gate.md

- The H-1..H-6 Heuristic Detection Table (lifecycle-sync-gate.md L37-45) describes marker PRESENCE detection accurately ("§E.2 present", "§E.5 present") and is NOT semantically wrong — it never claims §E.2 *means* sync. Only the worked-example section headings (L297 `§E.2 Sync-phase Audit-Ready Signal`, L300) carry the convention-A framing and are in scope. The heuristic table rows stay as-is.

## C. Canonical §E Section Map (convention B — locked)

| Marker | Phase meaning | Owning agent (per ownership matrix) |
|--------|---------------|-------------------------------------|
| `§E.1` | Plan-phase Audit-Ready Signal | manager-spec |
| `§E.2` | Run-phase Evidence | manager-develop |
| `§E.3` | Run-phase Audit-Ready Signal | manager-develop |
| `§E.4` | Sync-phase Audit-Ready Signal | manager-docs |
| `§E.5` | Mx-phase Audit-Ready Signal | manager-docs / orchestrator |

This map already matches the agent-body ownership matrices verbatim (`manager-develop.md` §Artifacts owned, `manager-docs.md` §Artifacts owned, `manager-spec.md` §Forbidden modifications). Convention B is therefore the existing majority; this SPEC aligns the one divergent document (`lifecycle-sync-gate.md` worked example) and the era.go comments to it.

## D. Requirements (GEARS)

### D.1 Marker-notation alignment (lifecycle-sync-gate.md worked example)

- **REQ-PMC-001** (Ubiquitous): The `lifecycle-sync-gate.md` worked-example progress.md excerpt **shall** present the §E section headings per the convention-B map in §C, replacing the convention-A heading `## §E.2 Sync-phase Audit-Ready Signal` with the canonical `## §E.4 Sync-phase Audit-Ready Signal` and retaining `## §E.5 Mx-phase Audit-Ready Signal`.

- **REQ-PMC-002** (Event-driven): When the worked example asserts the auto-detection trace reaches H-4, the trace narration **shall** remain internally consistent with the edited headings — i.e. the `sync_commit_sha` field shall appear under the `§E.4` Sync heading (not `§E.2`), and `§E.2 Run-phase Evidence` shall be the run-evidence marker, while H-4 detection (which greps for literal `§E.2` AND `§E.5` presence) **shall** still hold because both literal substrings remain present.

- **REQ-PMC-003** (Ubiquitous): The worked-example edit **shall not** alter any H-1..H-6 heuristic table row, the JSON audit-output excerpt's `heuristic_matched` string, or any `era:` field semantics described elsewhere in `lifecycle-sync-gate.md`.

### D.2 era.go comment correction (behavior-preserving)

- **REQ-PMC-004** (Ubiquitous): The `era.go` source comments **shall** describe `§E.2` as the §E-section run-evidence start marker and `§E.5` as the Mx-completion marker at the THREE `//` comment sites that currently frame `§E.2` as the "sync" gate (L33 const comment, L91 ClassifyEra doc-comment, L120 inline H-3 comment), and **shall** clarify that classification is string-presence-based (the variable names `hasSyncSection`/`hasMxSection` are misnomers that really mean "the §E-section progress structure starts and reaches Mx completion"). Verbatim-stay set (MUST NOT be edited): the L122 executable `return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"` statement (it is a `return` string, not a comment — editing it would break the behavior-preservation comment-only diff gate), every `sync_commit_sha` FIELD reference, and the H-4 `§E.2 + §E.5` enumeration — these are accurate, not framing errors.

- **REQ-PMC-005** (State-driven): While `era.go` is edited, the executable grep logic — the `hasProgressMarker` body (L161), the `hasSyncSection`/`hasMxSection` assignments (L110-111), the H-2/H-3/H-4/H-5 branch conditions (L116-134), and every other non-comment statement — **shall** remain byte-identical.

- **REQ-PMC-006** (Event-detected): When `go test ./internal/spec/...` runs after the era.go comment edit, the test suite **shall** pass with the same results as before the edit (no test added, removed, or changed-outcome).

### D.3 Skeleton-generation instruction (manager-spec.md, forward-looking)

- **REQ-PMC-007** (Ubiquitous): The `manager-spec.md` agent body **shall** contain an instruction directing manager-spec to emit a canonical `progress.md` skeleton at plan-phase SPEC creation, with the five placeholder section headings per the convention-B map in §C (`§E.1 Plan-phase Audit-Ready Signal`, `§E.2 Run-phase Evidence`, `§E.3 Run-phase Audit-Ready Signal`, `§E.4 Sync-phase Audit-Ready Signal`, `§E.5 Mx-phase Audit-Ready Signal`).

- **REQ-PMC-008** (Where — capability gate): Where the manager-spec skeleton instruction prescribes placeholder content under each §E heading, the skeleton **shall** be minimal (heading + a one-line placeholder note per section; no populated evidence tables), so that the emitted `§E.2`-`§E.5` placeholder headings enable H-2-avoidance (era.go's `hasAnyProgressMarker` greps for `§E.2`/`§E.3`/`§E.4`/`§E.5` — NOT `§E.1` — per era.go L165-169, so the literal `§E.2` heading is what causes H-2 to not fire) while leaving §E.2-§E.5 *content* population to the downstream owners (manager-develop, manager-docs). The `§E.1` heading is emitted for human/audit readability, not for H-2-avoidance.

- **REQ-PMC-009** (Unwanted behavior): The manager-spec skeleton instruction **shall not** authorize manager-spec to populate `§E.2`-`§E.5` evidence content at plan-phase — those sections belong to manager-develop (§E.2/§E.3) and manager-docs (§E.4/§E.5) per the existing Forbidden-modifications matrix; the plan-phase emission is placeholder headings only.

### D.4 Template mirror + build

- **REQ-PMC-010** (Event-driven): When `manager-spec.md` is edited, the template mirror `internal/template/templates/.claude/agents/moai/manager-spec.md` **shall** receive the byte-identical edit, and `make build` **shall** regenerate `internal/template/embedded.go`.

- **REQ-PMC-011** (Ubiquitous): The `lifecycle-sync-gate.md` edit **shall** be applied to the single canonical copy only — this file has NO template mirror (it is an internal dev rule per CLAUDE.local.md §2), so no mirror edit and no `make build` is required for that file.

- **REQ-PMC-012** (Ubiquitous): The `era.go` edit **shall** be applied to the single source file only — it is Go source code with no template mirror.

## E. Constraints

- **CON-PMC-001**: Convention B is locked (user decision). This SPEC does NOT re-open the §E.2-meaning question or evaluate convention A.
- **CON-PMC-002**: ZERO new production code. The only `.go` change is era.go comment text. No new functions, no new lint rules, no new tests.
- **CON-PMC-003**: Grandfather-protected SPECs (16 using §F.*, plus all V2.x/V3R2-R4/V3R5) MUST NOT be edited.
- **CON-PMC-004**: Per the close-subject full-ID mandate, the eventual close commit MUST name the full SPEC-ID `SPEC-PROGRESS-MARKER-CANON-001` (no abbreviated prefix).
- **CON-PMC-005**: Template neutrality — `manager-spec.md` template mirror must not carry forbidden internal-content classes (this SPEC's ID is acceptable in the local `.moai/specs/` copy but the manager-spec.md body edit must use generic prose, not embed this SPEC ID).

## F. Affected Files (provisional — finalized in plan.md §F)

| File | Change | Mirror? | make build? |
|------|--------|---------|-------------|
| `.claude/rules/moai/workflow/lifecycle-sync-gate.md` | Worked-example §E heading alignment (L297/L300 + trace narration) | NO (internal dev rule) | NO |
| `internal/spec/era.go` | Comment-only correction (L33, L92, const-block/doc-comment prose) | NO (source code) | NO |
| `.claude/agents/moai/manager-spec.md` | Add progress.md skeleton-generation instruction | YES | YES |
| `internal/template/templates/.claude/agents/moai/manager-spec.md` | Byte-identical mirror edit | (is the mirror) | YES |
