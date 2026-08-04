---
id: SPEC-SYNC-AUDIT-FALSIFICATION-001
title: "sync-auditor falsification obligations — AC-mechanism probes, VCI §1.1 surface-3 binding, AC-class coverage minimums"
version: 0.1.0
status: completed
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P0
phase: "v3.x target"
module: ".claude/agents/moai/sync-auditor.md + internal/template/templates mirror"
lifecycle: spec-anchored
tags: "sync-auditor, falsification, ac-mechanism, verification-claim-integrity, defect-claim, ac-class-coverage, autonomy-epic, template-mirror"
tier: M
related_specs: [SPEC-INFINITE-GOAL-001, SPEC-FOURDIM-PHANTOM-001, SPEC-AUDIT-GATE-INTEGRITY-001]
---

# SPEC-SYNC-AUDIT-FALSIFICATION-001 — sync-auditor falsification obligations

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-04 | manager-spec | Initial draft — review-followup-fix epic root-cause analysis (RC1-RC5); three auditor-body improvements IMP-1 / IMP-3 / IMP-6 in scope. IMP-2 / IMP-4 / IMP-5 and RC5 explicitly out of scope. |

## §A Context — User Story

### A.1 Problem definition

A `/moai review` of 5 completed SPECs found that sync-audit FAILED to catch real defects (1 FAIL + 2 PASS-WITH-DEBT across the reviewed SPECs). The root-cause analysis identified five structural / procedural gaps (RC1-RC5), of which three are addressable inside the `sync-auditor.md` agent body:

- **RC1** — the sync-auditor's 4-dimension evaluation is bounded to artifact + test-exit; it has no runtime-control-flow dimension. (Partially addressed here via IMP-1's runtime falsification probe, which forces at least one runtime observation per high-blast-radius AC.)
- **RC3** — AC-mechanism-truth is NOT verified; only `exits 0`. An AC can state a mechanism ("X rejects Y") that is false, yet pass because the test exits 0 vacuously (cf. feedback_ac_stated_mechanism_can_be_false, and SPEC-INFINITE-GOAL-001's original AC-011 which passed despite the cost-cap bound being unenforced).
- **RC4** — verification-claim-integrity (VCI) is doctrine-only; no mechanical gate binds the auditor's defect/drift claims to tool output.

The remaining two root causes (RC2 cross-artifact contradiction check; RC5 cold sync-auditor mandatory on green test suite) are NOT in this SPEC's scope (see §B Out of Scope).

### A.2 Evidence anchor (2026-08-04)

| # | Defect class | Evidence |
|---|--------------|----------|
| C1 / C3 | AC states a mechanism that is false yet the test suite exits 0, so sync-audit's `Functionality` dimension records a PASS without ever exercising the mechanism | `.claude/agents/moai/sync-auditor.md` L55-66 (Per-Dimension Mechanical Verification — Functionality row runs only `go test ./...` and cross-checks against the AC matrix; no mechanism-level probe); INFINITE-GOAL-001 run history (AC-011 passed pre-falsification despite cost-cap being unenforced — `feedback_ac_stated_mechanism_can_be_false`) |
| C4 | sync-auditor defect / drift findings can be inferred from frontmatter text or grep matches without running the domain's dedicated verification tool | `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 binds "ALL actors"; the sync-auditor body L55-66 cites §1.1 surface 2 (manager-agent self-report) for Evidence cells but does NOT cite surface 3 (defect/debt/drift identification) for its `### Findings` emission |
| C3 (defense-in-depth) | A SPEC whose high-blast-radius ACs cluster in one class (e.g. all functionality, no security) can pass audit by sampling only the easy class | `.claude/agents/moai/sync-auditor.md` L33-46 (Evaluation Dimensions table — no per-class sampling minimum); L92 ("coverage: surfacing a finding that later gets filtered out is preferable to silently dropping a real bug") mentions coverage at the finding stage, not at the AC-class sampling stage |

### A.3 Fix approach

Three obligations added to the `sync-auditor.md` agent body, each mapping to a caught defect class:

- **IMP-1 (priority a)** — for at least one HIGH-BLAST-RADIUS acceptance criterion per SPEC, FALSIFY the AC's stated mechanism via a runtime probe, not merely confirm the test exits 0.
- **IMP-3 (priority c)** — bind the auditor's `### Findings` emission to VCI §1.1 surface 3 + the §3 5-section Evidence format (Claim / Evidence / Baseline / Gaps / Residual-risk). A finding without tool-output evidence is a hypothesis, not a verified finding.
- **IMP-6 (defense-in-depth)** — sample AC coverage across AC CLASSES (functionality / security / safety / craft), with the high-blast-radius AC mandatory regardless of class, so a SPEC whose high-blast-radius ACs cluster in one class cannot pass audit by sampling only the easy class.

The change is doc-only (agent definition + template mirror). No Go runtime code changes.

## §B Scope

### In scope

- Edits to `.claude/agents/moai/sync-auditor.md` (155 lines today): three new obligations added to the Evaluation Contract + Per-Dimension Mechanical Verification areas.
- Byte-identical mirror edit to `internal/template/templates/.claude/agents/moai/sync-auditor.md` (CLAUDE.local.md §2 Template-First Rule + §25 Template Content Neutrality).
- Cross-reference to `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §3 (IMP-3 binds to it; no duplication).

### Out of Scope — IMP-2 (cross-artifact claim-drift grep)

- Cross-artifact contradiction check (sync-auditor compares spec.md claims vs run.md / code) is NOT implemented here. It is a larger structural change to the auditor's reading surface (currently `.moai/specs/` only). Tracked as a follow-up; not bundled to keep this SPEC's blast radius small.

### Out of Scope — IMP-4 (security/cost/safety AC runtime-path authoring rule)

- An authoring rule requiring security / cost / safety ACs to carry a runtime-path probe is NOT implemented here. That is an authoring-side change (manager-spec body, plan.md template), not an auditor-side change. Tracked as a follow-up.

### Out of Scope — IMP-5 (sync-audit-4dim.js falsification schema + phantom-mechanism → FAIL verdict)

- The falsification schema for the `sync-audit-4dim` skill and the phantom-mechanism → FAIL verdict rule belong to the SEPARATE SPEC-FOURDIM-PHANTOM-001. Cross-referenced via `related_specs`; not implemented here. Do not duplicate its requirements in this SPEC.

### Out of Scope — RC5 (cold sync-auditor mandatory on green test suite)

- RC5's own fix (cold sync-auditor mandatory even on green — currently the binding-promotion flow SKIPS the cold sync-auditor on a green test suite, leaving only the weaker workflow-judge path) is the highest-leverage single change identified in the root-cause analysis, BUT it is an ORCHESTRATOR-side workflow policy (binding-promotion routing), NOT a `sync-auditor.md` body change. It is noted here as related context; this SPEC does NOT claim to fix RC5 in the auditor body. RC5 will be addressed by a separate orchestrator-side SPEC.

## §C Requirements (GEARS)

### REQ-SAF-001 — IMP-1: AC-mechanism falsification obligation (HIGH-BLAST-RADIUS AC)

**Where** a SPEC under sync-audit declares at least one acceptance criterion whose failure would break a safety, correctness, or security invariant (a "high-blast-radius AC"), the sync-auditor shall, for at least one such high-blast-radius AC, construct or invoke a runtime probe that FALSIFIES the AC's stated mechanism — not merely confirm the test runner exits 0.

**When** the runtime probe observes that the stated mechanism does NOT produce the outcome the AC asserts (e.g. the AC claims "mechanism M rejects input I" and the probe observes M accepting I), the sync-auditor shall record that AC as FAIL in the `### Dimension Scores` Functionality row and emit a blocking finding in `### Findings` naming the AC ID, the stated mechanism, the probe input, and the observed outcome.

The falsification obligation applies on the happy path too — **While** the cold sync-auditor IS spawned (whether on a green or red test suite), it SHALL falsify at least one high-blast-radius AC's mechanism via a runtime probe before recording a Functionality PASS.

The probe MAY be a negative probe (confirm the mechanism rejects / fails as the AC claims) where feasible; where a negative probe is not feasible, an equivalent positive probe that observes the mechanism producing the asserted outcome under the asserted input is acceptable. The probe is a runtime observation, not a re-reading of the test source.

### REQ-SAF-002 — IMP-3: Verification-Claim-Integrity §1.1 surface-3 binding for Findings

**While** emitting any entry under `### Findings` (the structured defect-list in the Output Format), the sync-auditor shall bind each finding to `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (defect / debt / drift identification claims) — a finding MUST cite the domain's dedicated verification tool output (`moai spec audit`, `go test -cover`, `golangci-lint`, etc.) as its Evidence, and MUST NOT be inferred from frontmatter text, grep matches, or file absence alone.

**When** the sync-auditor cannot obtain tool output for a finding (tool not installed, command fails, domain lacks a dedicated tool), the sync-auditor shall either (a) run the closest-equivalent mechanical check and cite its verbatim output, or (b) record the finding with an explicit `unverified-premise` marker and downgrade it to `optional` severity — never emit it as a blocking finding without tool-output evidence.

Each blocking finding SHALL be structured per the VCI §3 5-section Evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — the agent body's Findings emission guidance MUST reference this format normatively.

### REQ-SAF-003 — IMP-6: AC-class coverage minimums (defense-in-depth)

**While** sampling which acceptance criteria to verify in the Functionality dimension, the sync-auditor SHALL sample at least one AC per AC CLASS present in the SPEC's acceptance.md §D AC matrix (the canonical classes are functionality / security / safety / craft; additional classes declared in the SPEC are also covered), rather than concentrating all sampling on a single class.

**Where** the SPEC declares a high-blast-radius AC, that AC is MANDATORY in the sample regardless of its class — so a SPEC whose high-blast-radius ACs cluster in one class (e.g. all security, no functionality) cannot pass audit by sampling only the easy class. The high-blast-radius mandatory AC is sampled IN ADDITION TO (not instead of) the per-class minimum.

**When** a SPEC declares fewer AC classes than the canonical four (e.g. a purely-functional SPEC with no security ACs), the absent classes are skipped — the per-class minimum is "≥1 AC sampled per present class", not "≥1 per canonical class".

## §D Constraints

- **Doc-only**: edits to `sync-auditor.md` (agent body) + template mirror only. No Go runtime code changes in `internal/`.
- **Template mirror is byte-identical**: `.claude/agents/moai/sync-auditor.md` and `internal/template/templates/.claude/agents/moai/sync-auditor.md` MUST remain byte-identical after the edit (CLAUDE.local.md §2 [HARD] Template-First Rule). Run `diff` to verify before commit.
- **Template Content Neutrality (CLAUDE.local.md §25)**: the obligation prose in the template mirror MUST NOT carry SPEC IDs (`SPEC-SYNC-AUDIT-FALSIFICATION-001`), REQ tokens (`REQ-SAF-*`), internal dates, commit SHAs, or audit citations. The obligation is authored generically (it names the obligation, the VCI cross-reference, and the defect class — not this SPEC's identifiers). The CI guard `template-neutrality-check.yaml` is the safety net.
- **VCI cross-reference, not duplication**: IMP-3 cross-references `verification-claim-integrity.md` §1.1 surface 3 + §3; it does NOT copy those sections into the agent body. The agent body cites the SSOT.
- **Language-neutrality preserved**: any runtime-probe example in the agent body is phrased language-independently (project-language auto-detection, 4 languages as equal examples — no language is PRIMARY), matching the existing Per-Dimension Mechanical Verification table style (sync-auditor.md L55-66).

## §E Assumptions

- A "high-blast-radius AC" is identifiable from the SPEC's acceptance.md — either via an explicit marker (`high-blast-radius`, `safety-invariant`, `must-pass`, or similar) on the AC, or via the sync-auditor's judgment when the SPEC does not mark one. Where the SPEC marks none and the auditor cannot identify one by judgment, the IMP-1 obligation degrades to "falsify at least one AC's stated mechanism" (still stricter than the status quo of test-exit-only).
- The domain's dedicated verification tools exist on the host (or the auditor degrades gracefully per REQ-SAF-002 option (b)).
- IMP-5 (sync-audit-4dim falsification schema) is owned by SPEC-FOURDIM-PHANTOM-001 and is assumed to land independently; this SPEC's REQs do not depend on IMP-5 being landed first.
- RC5 (cold sync-auditor mandatory on green) is assumed to be addressed by a separate orchestrator-side SPEC; this SPEC's IMP-1 obligation binds "While the cold sync-auditor IS spawned" — i.e. it activates only when sync-auditor actually runs, and does not itself mandate the spawn.

## §F Anti-Patterns (authoring-side, do NOT replicate)

- AP-1 — authoring a falsification obligation that reads "verify the mechanism" rather than "falsify the mechanism" (asymmetric: confirmation is satisfied by a vacuous pass; falsification requires a probe that could fail).
- AP-2 — citing the VCI rule by duplicating its content into the agent body (creates a drift surface; cross-reference instead).
- AP-3 — naming a specific language's toolchain in the obligation prose (breaks 16-language template neutrality).
- AP-4 — leaving the template mirror non-byte-identical to the live agent file (breaks CLAUDE.local.md §2 Template-First Rule; the distributed users receive a stale agent).

## §G Risks

- **R-1 (procedural)**: IMP-1's runtime probe adds wall-clock latency to every sync-audit invocation. Mitigation: the obligation is scoped to "at least one high-blast-radius AC" (not all ACs), keeping the marginal cost bounded.
- **R-2 (correctness)**: IMP-1's "construct or invoke a runtime probe" is open-ended; a too-loose phrasing risks the auditor constructing a trivial probe that always passes. Mitigation: REQ-SAF-001 requires the probe to be a runtime observation of the stated mechanism (not a re-read of test source), and the negative-probe form is preferred where feasible.
- **R-3 (template-neutrality)**: the obligation prose must be generic for the template mirror. Mitigation: §D constraint + the 5-item pre-commit self-check in CLAUDE.local.md §25.3.

## §H AC summary (full GWT in acceptance.md)

- **AC-SAF-001** (REQ-SAF-001 / IMP-1): given a SPEC with a high-blast-radius AC whose stated mechanism is false yet the test suite exits 0, the sync-auditor's falsification probe FAILS the AC and records a blocking finding. Binary-testable via fixture: a fixture SPEC + fixture code where the AC mechanism is intentionally false, run sync-audit, assert `### Dimension Scores` Functionality row = FAIL and `### Findings` contains a blocking entry naming the AC.
- **AC-SAF-002** (REQ-SAF-002 / IMP-3): given a sync-audit `### Findings` entry, the entry cites verbatim tool output as Evidence (not frontmatter text / grep match alone); entries lacking tool output are downgraded to `optional` with an `unverified-premise` marker.
- **AC-SAF-003** (REQ-SAF-003 / IMP-6): given a SPEC whose acceptance.md declares ACs across classes {functionality, security} with the high-blast-radius AC in the security class, the sync-auditor's Functionality-dimension sample includes at least one security-class AC (the mandatory high-blast-radius one) AND at least one functionality-class AC — it cannot pass by sampling only functionality.
- **AC-SAF-004** (template mirror byte-identity): after the run-phase edit, `diff .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` exits 0 with no output.
- **AC-SAF-005** (template neutrality): the template mirror contains no SPEC ID (`SPEC-SYNC-AUDIT-FALSIFICATION-001`), no REQ token (`REQ-SAF-*`), no internal dates, no commit SHAs (CI guard `template-neutrality-check.yaml` PASS).

## §I Cross-References

- `.claude/agents/moai/sync-auditor.md` (L55-66 Per-Dimension Mechanical Verification, L85-94 Findings + Output Format) — primary edit surface.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §3 (5-section Evidence format) — IMP-3 cross-reference target (SSOT, not duplicated).
- `CLAUDE.local.md` §2 (Template-First Rule) + §25 (Template Internal-Content Isolation) — template-mirror constraints.
- `related_specs`: SPEC-INFINITE-GOAL-001 (AC-011 falsification precedent), SPEC-FOURDIM-PHANTOM-001 (IMP-5 owner, out-of-scope here), SPEC-AUDIT-GATE-INTEGRITY-001 (REQ-AGI-006 language-neutrality precedent).
- `feedback_ac_stated_mechanism_can_be_false` (memory topic) — IMP-1 root-cause precedent.
