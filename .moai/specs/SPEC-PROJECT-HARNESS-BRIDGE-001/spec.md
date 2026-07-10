---
id: SPEC-PROJECT-HARNESS-BRIDGE-001
title: "Project interview clarity-scoring + harness-spec.yaml bridge to harness generation"
version: "0.1.0"
status: draft
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows/project, internal/template/templates/.claude/skills/moai/workflows/project"
lifecycle: spec-anchored
tags: "project, interview, harness, clarity-scoring, machine-readable-spec, template-mirror"
era: V3R6
tier: M
---

# SPEC-PROJECT-HARNESS-BRIDGE-001 — Project interview clarity-scoring + harness-spec.yaml bridge to harness generation

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-11 | 0.1.0 | Initial plan-phase draft (Tier M, 12 REQ / 14 AC). Foundation SPEC of the 3-SPEC "Project-Harness Pipeline" Epic (NO depends_on; the other two Epic SPECs depend_on this one). Two changes: (1) convert the STATIC 3-round Phase 0.3 / Phase 1.5 project interviews to clarity-scored ADAPTIVE interviews reusing the `plan/clarity-interview.md` mechanism; (2) emit a machine-readable `.moai/project/harness-spec.yaml` artifact and wire it into harness generation so `meta-harness.md` Phase 5.1 and `harness-build-entry.md` Phase 1 stop discarding / re-eliciting the interview data. Doc-only (markdown/yaml); no Go code. All Template-First. 2 `[NEEDS CLARIFICATION]` markers in plan.md. | manager-spec |

## §A. Context and Intent

The `/moai project` flow gathers project intent through an interview and then
generates a project-specific harness. Two structural gaps make that flow leak
information:

1. **The interview is fixed-length, not clarity-driven.** Phase 0.3 (new
   project) and Phase 1.5 (existing project) in `project/mode-detection.md` run
   a STATIC 3-round × 3-question interview. `.moai/config/sections/interview.yaml`
   already defines `clarity_threshold: 4` and `project.max_rounds: 3`, but the
   project flow does NOT consume `clarity_threshold` — it always runs exactly 3
   rounds regardless of how clear (or unclear) the answers are. The **plan** flow
   already solves this: `plan/clarity-interview.md` scores answer clarity and
   adapts round count. This SPEC mirrors that adaptive mechanism into the project
   interview.

2. **The interview output is discarded before harness generation.** The
   interview writes `.moai/project/interview.md`, but that file is NOT passed
   into harness generation. `project/meta-harness.md` Phase 5.1 composes the
   harness-creation request from `product.md` / `structure.md` / `tech.md` only,
   then routes to `harness-build-entry.md`, whose Phase 1 (Context-First
   Discovery) re-extracts domain / goal / constraints / scope **from scratch** —
   re-asking what the interview already asked. This SPEC introduces a
   machine-readable `.moai/project/harness-spec.yaml` artifact that carries the
   interview's structured answers forward, so harness generation consumes them
   instead of re-eliciting them.

**Design premise (Anthropic-verified pattern).** The "Let Claude interview you"
pattern — an `AskUserQuestion` interview that produces a machine-readable spec
artifact which downstream steps then execute — is the target shape. This SPEC
realizes that shape for the project→harness pipeline: adaptive interview →
`harness-spec.yaml` → harness generation.

**Boundary principle.** This is the FOUNDATION of a 3-SPEC Epic. It creates the
`harness-spec.yaml` contract and the adaptive-interview behavior; the two
downstream Epic SPECs (which `depend_on` this one) build on the artifact. This
SPEC has NO `depends_on`.

## §B. Scope Summary

**In scope**:
- Convert Phase 0.3 (new-project) and Phase 1.5 (existing-project) interviews in
  `project/mode-detection.md` from static 3-round to clarity-scored adaptive
  (consume `interview.yaml` `clarity_threshold`; early-exit at threshold;
  additional rounds up to `project.max_rounds`) — mirroring
  `plan/clarity-interview.md`.
- Extend the interview question set to elicit four new fields: verification
  method, UI surface, external systems, team-sharing intent.
- Emit `.moai/project/harness-spec.yaml` (8-field machine-readable artifact)
  from `project/doc-generation.md`, alongside `interview.md`.
- Wire `project/meta-harness.md` Phase 5.1 to compose from `harness-spec.yaml`
  PLUS `product/structure/tech.md`.
- Wire `harness-build-entry.md` Phase 1 to consume `harness-spec.yaml` and
  pre-satisfy already-answered fields (re-ask only absent / ambiguous fields).
- Mirror all edits into `internal/template/templates/...` (Template-First),
  `make build`, keep the neutrality CI guard green.

**Preserve**:
- The `/moai project` NO-SPEC scope guard (project flow never writes to
  `.moai/specs/**`).
- The existing `interview.md` human-readable output (harness-spec.yaml is
  additive, not a replacement).
- The builder-harness specialist-generation internals (unchanged).

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 Adaptive clarity-scored interview

- **REQ-PHB-001** (State-driven): While the project interview's accumulated
  clarity score is below the `interview.yaml` `clarity_threshold` (4), the
  project workflow shall run one or more additional interview rounds, up to
  `project.max_rounds` (3).
- **REQ-PHB-002** (Event-driven): When the accumulated clarity score reaches
  `clarity_threshold` before `max_rounds` rounds have run, the project workflow
  shall terminate the interview early (skip the remaining rounds), replacing the
  current always-exactly-3-rounds behavior.
- **REQ-PHB-003** (Ubiquitous): The clarity-scoring loop shall be applied to
  BOTH Phase 0.3 (new-project) and Phase 1.5 (existing-project) interviews in
  `project/mode-detection.md`, reusing the adaptive mechanism defined in
  `plan/clarity-interview.md` (the plan flow's clarity scoring is the reference
  implementation the run-phase reads and mirrors).

### C.2 Extended question axes

- **REQ-PHB-004** (Ubiquitous): The project interview question set shall elicit
  four new fields beyond the existing vision / technology / scope axes —
  verification method (test / e2e command), UI surface (has-UI / headless),
  external systems (DB / APIs / services), and team-sharing intent.

### C.3 Machine-readable harness-spec.yaml artifact

- **REQ-PHB-005** (Ubiquitous): The project workflow shall produce a
  machine-readable artifact `.moai/project/harness-spec.yaml` carrying exactly
  the fields `domain`, `goal`, `constraints`, `scope`, `verification`,
  `external_systems`, `ui_surface`, `team_sharing`; the schema is stated in §D.
- **REQ-PHB-006** (Event-driven): When the `project/doc-generation.md` phase
  runs, it shall write `harness-spec.yaml` alongside `.moai/project/interview.md`,
  populated from the interview answers.

### C.4 Harness generation consumes harness-spec.yaml

- **REQ-PHB-007** (Ubiquitous): `project/meta-harness.md` Phase 5.1 shall compose
  the harness-creation request from `harness-spec.yaml` PLUS `product.md` /
  `structure.md` / `tech.md` — no longer discarding the interview data.
- **REQ-PHB-008** (Event-driven): When `harness-build-entry.md` Phase 1
  (Context-First Discovery) runs, it shall consume `harness-spec.yaml` to
  PRE-SATISFY the domain / goal / constraints / scope fields, and shall re-ask
  ONLY fields that are absent or ambiguous in `harness-spec.yaml`.
- **REQ-PHB-009** (Unwanted behavior): The `harness-build-entry.md` Phase 1
  Discovery shall not re-ask a question for any field already answered in
  `harness-spec.yaml` (no duplicate interview of an already-elicited field).

### C.5 NO-SPEC scope guard (preservation invariant)

- **REQ-PHB-010** (Unwanted behavior): The project workflow — including the new
  adaptive interview and the `harness-spec.yaml` write — shall not write to
  `.moai/specs/**`; the `harness-spec.yaml` artifact shall live under
  `.moai/project/`.

### C.6 Template-First mirror invariant

- **REQ-PHB-011** (Ubiquitous): Every edit shall be made in
  `internal/template/templates/...` FIRST, then mirrored byte-identically to the
  local `.claude/` copy, then compiled via `make build`; template neutrality (no
  internal SPEC IDs, internal dates, or commit SHAs in the template tree) shall
  be preserved.
- **REQ-PHB-012** (Event-driven): When `moai init` deploys a fresh project from
  the post-change binary, the deployed tree shall carry the adaptive
  clarity-scored interview, the `harness-spec.yaml` write path, and the extended
  interview axes (mirror parity with the local tree).

## §D. harness-spec.yaml Schema (SSOT)

The machine-readable harness-generation input artifact. Written to
`.moai/project/harness-spec.yaml` by `project/doc-generation.md`; consumed by
`project/meta-harness.md` Phase 5.1 and `harness-build-entry.md` Phase 1.

```yaml
# .moai/project/harness-spec.yaml — machine-readable harness generation input
domain: <string>              # primary problem domain (e.g. "cli-tooling", "web-api")
goal: <string>                # one-line project goal / success condition
constraints: [<string>, ...]  # hard constraints (performance, security, compatibility)
scope: <string>               # in-scope / out-of-scope boundary summary
verification: <string>        # test / e2e command or verification method
external_systems: [<string>, ...]  # DB / APIs / services the project integrates
ui_surface: <enum>            # has-ui | headless
team_sharing: <enum>          # solo | team-shared
```

The full machine-verifiable AC matrix (AC-PHB-001 … AC-PHB-014) lives in
`acceptance.md` (SSOT). Every REQ above maps to at least one AC; preservation
REQs (REQ-PHB-010) map to a NO-WRITE / absence assertion.

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — Downstream Epic SPECs

- The two dependent Epic SPECs (which `depend_on` this one) that consume the
  `harness-spec.yaml` contract for their own features are separate deliverables.
  This SPEC creates the artifact + adaptive-interview behavior; it does NOT wire
  the downstream consumers beyond `meta-harness.md` / `harness-build-entry.md`.

### Out of Scope — builder-harness specialist internals

- The `builder-harness` agent's specialist-generation logic (how it authors the
  actual harness skills / agents from the composed request) is unchanged. This
  SPEC changes only WHAT input the request carries (harness-spec.yaml), not HOW
  specialists are generated.

### Out of Scope — Go code changes

- This SPEC is doc-only (markdown / yaml under `.claude/skills/...` and
  `.moai/config/sections/interview.yaml` and their template mirrors). No
  `internal/` / `pkg/` / `cmd/` Go source is modified. No `harness-spec.yaml`
  Go parser is added here (consumers read it as prose-context via the workflow
  skills, not via a typed loader).

### Out of Scope — interview.md format redesign

- The existing human-readable `.moai/project/interview.md` output format is
  preserved as-is; `harness-spec.yaml` is an additive machine-readable sibling,
  not a replacement or reformat of `interview.md`.

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site
  4-locale updates for the new project-interview behavior are a follow-up
  sync / docs concern.

## §F. Cross-References

- `.claude/skills/moai/workflows/project/mode-detection.md` — Phase 0.3 / 1.5
  interview host (adaptive conversion target).
- `.claude/skills/moai/workflows/project/doc-generation.md` — harness-spec.yaml
  write site.
- `.claude/skills/moai/workflows/project/meta-harness.md` — Phase 5.1 compose
  site.
- `.claude/skills/moai/workflows/harness-build-entry.md` — Phase 1 Context-First
  Discovery consume site.
- `.claude/skills/moai/workflows/plan/clarity-interview.md` — the reference
  adaptive clarity-scoring mechanism this SPEC mirrors into the project flow.
- `.moai/config/sections/interview.yaml` — `clarity_threshold: 4` +
  `project.max_rounds: 3` (already present; project flow will start consuming
  `clarity_threshold`).
- `plan.md` / `acceptance.md` — implementation plan + AC matrix (SSOT).
