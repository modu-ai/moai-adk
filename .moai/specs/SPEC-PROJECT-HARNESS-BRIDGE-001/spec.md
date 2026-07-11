---
id: SPEC-PROJECT-HARNESS-BRIDGE-001
title: "Project interview clarity-scoring + harness-spec.yaml bridge to harness generation"
version: "0.2.0"
status: in-progress
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
amendment_of: SPEC-PROJECT-HARNESS-BRIDGE-001
---

# SPEC-PROJECT-HARNESS-BRIDGE-001 — Project interview clarity-scoring + harness-spec.yaml bridge to harness generation

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-11 | 0.1.0 | Initial plan-phase draft (Tier M, 12 REQ / 14 AC). Foundation SPEC of the 3-SPEC "Project-Harness Pipeline" Epic (NO depends_on; the other two Epic SPECs depend_on this one). Two changes: (1) convert the STATIC 3-round Phase 0.3 / Phase 1.5 project interviews to clarity-scored ADAPTIVE interviews reusing the `plan/clarity-interview.md` mechanism; (2) emit a machine-readable `.moai/project/harness-spec.yaml` artifact and wire it into harness generation so `meta-harness.md` Phase 5.1 and `harness-build-entry.md` Phase 1 stop discarding / re-eliciting the interview data. Doc-only (markdown/yaml); no Go code. All Template-First. 2 open clarifications recorded in plan.md. | manager-spec |
| 2026-07-11 | 0.1.1 | Plan-phase targeted fix pass (D1-D6). D1 (factual): the Phase 1.5 (existing-project) interview lives in `project/codebase-analysis.md`, NOT `project/mode-detection.md` (which hosts Phase 0.3 only) — corrected throughout spec/plan/acceptance. D2: 2 clarifications resolved — interview.yaml schema surface = config-declared `additional_axes:` block; harness-spec.yaml re-run = OVERWRITE (matching interview.md regeneration). D3: clarity-scoring reconciled to `plan/clarity-interview.md` semantics (0-10 scale, sufficiency exit ≥ 8, abandon ≤ 3, entry floor 4 = `clarity_threshold`). D4: interview.yaml mirror achieved by wholesale template→local overwrite (pre-existing byte-drift). D5: "ambiguous" defined operationally (REQ-PHB-008); REQ-PHB-006 reworded (interview.md written by the interview phases, not doc-generation.md). Open-clarification markers removed from plan.md. | manager-spec |
| 2026-07-11 | 0.2.0 | **In-place amendment** (`completed → in-progress`) resolving a sync-auditor MUST-FIX: the extended-axes Round 4 was UNREACHABLE on both interview paths, making the SPEC's headline feature inert. Introduces the **two-stage interview** (Stage A = clarity-scored adaptive discovery, Rounds 1-3, capped by `project.max_rounds`; Stage B = mandatory extended-axes Round 4, exempt from the cap, the early-exit skip, and clarity scoring). Amends REQ-PHB-002 (early-exit now gated on required-base-field completion) and adds REQ-PHB-013 … REQ-PHB-019 + AC-PHB-015 … AC-PHB-020 (REACHABILITY ACs). See `## Amendments` below. | manager-spec |

## Amendments

### Amendment 0.2.0 — Two-stage interview (Round 4 reachability)

- **Prior completed version**: `0.1.1`
- **prior_completed_sha**: `0c9871b46b6719325427dc0126e4eb65d7b0f2d8` (sync commit; run commit `9f1f0f53b`)
- **Amendment type**: in-place (`amendment_of` is self-referential; `status: completed → in-progress`)

**Rationale.** An adversarial sync-audit found that the SPEC's headline feature — the four extended axes (verification / ui_surface / external_systems / team_sharing) flowing into `harness-spec.yaml` and pre-satisfying harness generation — is **unreachable on the default path**. Both interview hosts define a **Round 4** carrying those axes, but `.moai/config/sections/interview.yaml` sets `project.max_rounds: 3`, so Round 4 is blocked twice over: (a) the round cap hard-stops the interview at 3, and (b) the REQ-PHB-002 early-exit (clarity ≥ 8) explicitly "skips the remaining rounds". The consequence is a silent, complete failure of the feature: the four axes are never collected → `harness-spec.yaml` writes them always-empty → `harness-build-entry.md` treats them ABSENT → Discovery re-asks everything, which is precisely the information leak this SPEC exists to close. The original 14-AC matrix missed it because AC-PHB-004 greps for axis **token presence** in the interview host, never for **reachability** of the round that elicits them. Three further defects ride the same root cause: the early-exit can also strand REQUIRED base fields (F-2); the two hosts' Round topics are asymmetric and neither covers all four base fields (F-3); and `codebase-analysis.md`'s frontmatter still advertises a fixed `3-round` interview (F-4).

**Scope.** Doc-only (markdown + yaml), the same file-pairs as v0.1.x plus one dispatcher surface carrying the same stale fixed-round claim (F-4):

| Surface | Amendment touch |
|---------|-----------------|
| `project/mode-detection.md` (+ template mirror) | Stage A / Stage B structure; early-exit precondition; Round 4 exemption; base-field coverage |
| `project/codebase-analysis.md` (+ template mirror) | same as above, plus frontmatter `3-round` removal (F-4) |
| `project/doc-generation.md` (+ template mirror) | REQUIRED vs EXTENDED field annotation |
| `.moai/config/sections/interview.yaml` (+ template mirror) | `max_rounds` documented as Stage-A-scoped |
| `project.md` (+ template mirror) | routing-table `3-round` claim removal (F-4) — **scope delta**: this file-pair was not in the v0.1.x touch set |
| `project/meta-harness.md`, `harness-build-entry.md` (+ mirrors) | unchanged by this amendment (v0.1.x wiring stands) |

**Affected §C requirements**: REQ-PHB-002 (amended — required-base-field precondition on the early-exit); REQ-PHB-013, REQ-PHB-014, REQ-PHB-015, REQ-PHB-016, REQ-PHB-017, REQ-PHB-018, REQ-PHB-019 (new, §C.7). No existing REQ is weakened or deleted.

## §A. Context and Intent

The `/moai project` flow gathers project intent through an interview and then
generates a project-specific harness. Two structural gaps make that flow leak
information:

1. **The interview is fixed-length, not clarity-driven.** Phase 0.3 (new
   project) in `project/mode-detection.md` and Phase 1.5 (existing project) in
   `project/codebase-analysis.md` run a STATIC 3-round × 3-question interview.
   `.moai/config/sections/interview.yaml` already defines `clarity_threshold: 4`
   (the interview ENTRY floor — the clarity band at/above which the interview
   runs) and `project.max_rounds: 3`, but the project flow does NOT consume
   `clarity_threshold` at all and never adapts round count to answer clarity — it
   always runs exactly 3 rounds regardless of how clear (or unclear) the answers
   are. The **plan** flow already solves this: `plan/clarity-interview.md` scores
   answer clarity on a 0-10 scale and adapts round count — it exits early at the
   sufficiency target (clarity ≥ 8), abandons on a drop to ≤ 3, and treats 4 as
   the entry floor. This SPEC mirrors that adaptive mechanism (the SAME 0-10
   scale semantics) into both project interviews.

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
- Convert the Phase 0.3 (new-project) interview in `project/mode-detection.md`
  AND the Phase 1.5 (existing-project) interview in `project/codebase-analysis.md`
  from static 3-round to clarity-scored adaptive on a 0-10 scale: consume
  `interview.yaml` `clarity_threshold` (4) as the ENTRY floor; early-exit at the
  sufficiency target (clarity ≥ 8); abandon on a drop to ≤ 3; additional rounds
  up to `project.max_rounds` (3) — mirroring `plan/clarity-interview.md`.
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

**In scope — amendment v0.2.0 (two-stage interview)**:
- Split the interview into **Stage A** (clarity-scored adaptive discovery, Rounds
  1-3, capped by `project.max_rounds`) and **Stage B** (the mandatory extended-axes
  Round 4, which ALWAYS runs after Stage A terminates — exempt from the cap, from
  the early-exit skip, and from clarity scoring). This is what makes the four
  extended axes REACHABLE; without it the SPEC's headline feature is inert.
- Gate the Stage A early-exit on completion of the four REQUIRED base fields
  (`domain`, `goal`, `constraints`, `scope`), so a high-clarity round-1 answer can
  never strand a required field.
- Align both hosts' Stage A round topics so that, between them, all four required
  base fields are elicited (an existing-project field auto-populated from codebase
  analysis counts as answered and is NOT re-asked).
- Remove the stale fixed-`3-round` interview claim from `codebase-analysis.md`
  frontmatter and from the `project.md` routing table (both trees).

**Preserve**:
- The `/moai project` NO-SPEC scope guard (project flow never writes to
  `.moai/specs/**`).
- The existing `interview.md` human-readable output (harness-spec.yaml is
  additive, not a replacement).
- The builder-harness specialist-generation internals (unchanged).

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

> **Amendment v0.2.0 reading note.** REQ-PHB-001 … REQ-PHB-003 govern **Stage A**
> (the clarity-scored adaptive discovery rounds). The Stage B extended-axes round
> is governed by §C.7 (REQ-PHB-013 … REQ-PHB-019) and is NOT subject to the Stage A
> clarity loop, the Stage A early-exit, or `project.max_rounds`.

### C.1 Adaptive clarity-scored interview (Stage A)

- **REQ-PHB-001** (State-driven): While the project interview's accumulated
  clarity score (0-10 scale) is below the sufficiency target (clarity ≥ 8,
  mirroring `plan/clarity-interview.md`) and above the abandon floor (≤ 3), the
  project workflow shall run one or more additional Stage A interview rounds, up
  to `project.max_rounds` (3). (`interview.yaml` `clarity_threshold` (4) is the
  interview ENTRY floor — NOT the early-exit target.)
- **REQ-PHB-002** (Compound — *amended in v0.2.0*): **While** all four REQUIRED
  base fields (`domain`, `goal`, `constraints`, `scope` — see REQ-PHB-016) have
  been answered, **when** the accumulated clarity score reaches the configured
  sufficiency target (clarity ≥ 8 on the 0-10 scale — the SAME bar as
  `plan/clarity-interview.md`) before `max_rounds` rounds have run, the project
  workflow shall terminate **Stage A** early (skip the remaining Stage A rounds),
  replacing the original always-exactly-3-rounds behavior. The early exit
  terminates Stage A ONLY — it shall not skip the mandatory Stage B round
  (REQ-PHB-013). (v0.2.0 amendment: the required-base-field precondition is NEW;
  without it a high-clarity round-1 answer could strand a REQUIRED field — e.g.
  `mode-detection.md` would skip its Scope round, `codebase-analysis.md` its
  Constraints round.)
- **REQ-PHB-003** (Ubiquitous): The clarity-scoring loop shall be applied to
  BOTH the Phase 0.3 (new-project) interview in `project/mode-detection.md` AND
  the Phase 1.5 (existing-project) interview in `project/codebase-analysis.md`,
  reusing the adaptive mechanism defined in `plan/clarity-interview.md` (the plan
  flow's clarity scoring is the reference implementation the run-phase reads and
  mirrors).

### C.2 Extended question axes

- **REQ-PHB-004** (Ubiquitous): The project interview question set — in BOTH the
  Phase 0.3 host `project/mode-detection.md` and the Phase 1.5 host
  `project/codebase-analysis.md` — shall elicit four new fields beyond the
  existing vision / technology / scope axes: verification method (test / e2e
  command), UI surface (has-UI / headless), external systems (DB / APIs /
  services), and team-sharing intent. These four axes shall be config-declared in
  an `additional_axes:` block in `.moai/config/sections/interview.yaml` (both
  trees), consistent with the existing config-as-SSOT placement of
  `clarity_threshold` / `max_rounds`.

### C.3 Machine-readable harness-spec.yaml artifact

- **REQ-PHB-005** (Ubiquitous): The project workflow shall produce a
  machine-readable artifact `.moai/project/harness-spec.yaml` carrying exactly
  the fields `domain`, `goal`, `constraints`, `scope`, `verification`,
  `external_systems`, `ui_surface`, `team_sharing`; the schema is stated in §D.
- **REQ-PHB-006** (Event-driven): When the `project/doc-generation.md` phase
  runs, it shall write `.moai/project/harness-spec.yaml`, populated from the
  interview answers recorded in `.moai/project/interview.md`. (`interview.md`
  itself is written by the interview phases — Phase 0.3 / Phase 1.5 — NOT by
  `doc-generation.md`; `doc-generation.md` READS those recorded answers to
  populate the harness-spec.yaml fields.)

### C.4 Harness generation consumes harness-spec.yaml

- **REQ-PHB-007** (Ubiquitous): `project/meta-harness.md` Phase 5.1 shall compose
  the harness-creation request from `harness-spec.yaml` PLUS `product.md` /
  `structure.md` / `tech.md` — no longer discarding the interview data.
- **REQ-PHB-008** (Event-driven): When `harness-build-entry.md` Phase 1
  (Context-First Discovery) runs, it shall consume `harness-spec.yaml` to
  PRE-SATISFY the domain / goal / constraints / scope fields, and shall re-ask
  ONLY fields that are absent or ambiguous in `harness-spec.yaml`. A field is
  **ambiguous** (eligible for re-ask) when its value is empty, null, a
  placeholder token (e.g. `<string>` / `TODO` / `TBD`), or multi-valued in a way
  that does not resolve to a single answer; a field carrying a single concrete
  value is PRE-SATISFIED and MUST NOT be re-asked.
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

### C.7 Two-stage interview structure (amendment v0.2.0)

The project interview is a **two-stage** procedure. Stage A is clarity-scored and
variable-length; Stage B is mandatory and fixed. The stages are governed by
different rules, and conflating them is what made the extended axes unreachable in
v0.1.x.

| Stage | Rounds | Governed by | Terminates when |
|-------|--------|-------------|-----------------|
| **A — clarity-scored adaptive discovery** | 1 … `project.max_rounds` (3) | `clarity_threshold` (4, entry floor), sufficiency exit (≥ 8), abandon floor (≤ 3), `project.max_rounds` cap | early-exit (REQ-PHB-002), abandon, OR cap reached |
| **B — mandatory extended-axes round** | Round 4 (exactly one) | REQ-PHB-013 / REQ-PHB-014 — EXEMPT from clarity scoring, from the early-exit skip, and from `project.max_rounds` | after its four axes are collected |

**Design rationale.** The four extended axes (`verification`, `ui_surface`,
`external_systems`, `team_sharing`) are **clarity-independent factual collection**,
not ambiguity resolution — there is nothing to "score" about whether a project has
a UI or which test command it runs. Subjecting them to the Stage A clarity loop
(as v0.1.x implicitly did by placing them in a Round 4 beneath a 3-round cap) is a
category error, and it is what rendered them unreachable.

- **REQ-PHB-013** (Event-driven): When Stage A terminates — by early exit
  (REQ-PHB-002), by abandon (clarity ≤ 3), OR by reaching `project.max_rounds` —
  the project workflow shall run the Stage B extended-axes round (Round 4)
  unconditionally, in BOTH the Phase 0.3 host `project/mode-detection.md` and the
  Phase 1.5 host `project/codebase-analysis.md`, before proceeding to
  documentation generation.
- **REQ-PHB-014** (Unwanted behavior): The Stage B round shall not be counted
  against `project.max_rounds`, shall not be skipped by the Stage A early exit,
  shall not be skipped by the Stage A abandon path, and shall not be gated on any
  clarity score. Both interview hosts shall state this exemption explicitly (the
  exemption is the reachability contract — a host that merely *contains* a Round 4
  heading without stating the exemption re-introduces the v0.1.x defect).
- **REQ-PHB-015** (Ubiquitous): The `project.max_rounds` (3) value in
  `.moai/config/sections/interview.yaml` shall cap ONLY the Stage A clarity-scored
  discovery rounds — NOT the interview as a whole. The config file shall document
  this scoping in-place, so a reader of `interview.yaml` alone cannot infer that
  the interview terminates after 3 rounds.
- **REQ-PHB-016** (Ubiquitous): The 8 `harness-spec.yaml` fields shall be
  partitioned into two classes with distinct collection semantics: four **REQUIRED
  base fields** (`domain`, `goal`, `constraints`, `scope`), collected in Stage A
  and gating the Stage A early exit (REQ-PHB-002); and four **EXTENDED fields**
  (`verification`, `ui_surface`, `external_systems`, `team_sharing`), collected in
  the mandatory Stage B round (REQ-PHB-013). The partition shall be stated in §D
  and in `project/doc-generation.md`.
- **REQ-PHB-017** (State-driven): While one or more REQUIRED base fields remain
  unanswered, the project workflow shall continue running Stage A rounds (up to
  `project.max_rounds`) rather than exiting early, regardless of the accumulated
  clarity score. (Stage A's cap remains a hard stop: on reaching `max_rounds` with
  a required field still unanswered, the workflow records that field as absent —
  see acceptance.md §D E6 — and proceeds to Stage B; it does NOT loop.)
- **REQ-PHB-018** (Ubiquitous): Each interview host's Stage A rounds shall, between
  them, elicit ALL FOUR required base fields (`domain`, `goal`, `constraints`,
  `scope`), and each host shall carry an explicit base-field coverage mapping
  naming which Stage A round elicits which field. For the existing-project host
  (`project/codebase-analysis.md`), a base field confidently auto-populated from
  the Phase 1 codebase analysis (e.g. `domain` inferred from the code) COUNTS as
  answered and shall NOT be re-asked (consistent with acceptance.md §D E3).
  (v0.1.x defect: `mode-detection.md` rounds are Vision / Technology / Scope — no
  constraints round; `codebase-analysis.md` rounds are Ownership-Purpose /
  Constraints / Documentation-Priority — no scope round. Neither host covers all
  four, so `doc-generation.md`'s field mapping assumed a uniform interview shape
  that does not exist.)
- **REQ-PHB-019** (Unwanted behavior): No project-flow workflow surface — including
  the interview hosts' YAML frontmatter `description:` and the `project.md`
  routing table, in BOTH the local and template trees — shall describe the project
  interview as a fixed-length `3-round` interview, because the interview is
  variable-length (Stage A) plus one mandatory round (Stage B).

## §D. harness-spec.yaml Schema (SSOT)

The machine-readable harness-generation input artifact. Written to
`.moai/project/harness-spec.yaml` by `project/doc-generation.md`; consumed by
`project/meta-harness.md` Phase 5.1 and `harness-build-entry.md` Phase 1.

```yaml
# .moai/project/harness-spec.yaml — machine-readable harness generation input

# --- REQUIRED base fields (Stage A: clarity-scored discovery, Rounds 1-3) ---
# These four gate the Stage A early exit (REQ-PHB-002) — Stage A does not exit
# early until all four are answered.
domain: <string>              # REQUIRED — primary problem domain (e.g. "cli-tooling", "web-api")
goal: <string>                # REQUIRED — one-line project goal / success condition
constraints: [<string>, ...]  # REQUIRED — hard constraints (performance, security, compatibility)
scope: <string>               # REQUIRED — in-scope / out-of-scope boundary summary

# --- EXTENDED fields (Stage B: mandatory Round 4, REQ-PHB-013) ---
# Collected in the mandatory extended-axes round, which ALWAYS runs after Stage A
# terminates — exempt from `project.max_rounds`, from the Stage A early-exit skip,
# and from clarity scoring (REQ-PHB-014). These are clarity-independent factual
# collection, not ambiguity resolution.
verification: <string>        # EXTENDED — test / e2e command or verification method
external_systems: [<string>, ...]  # EXTENDED — DB / APIs / services the project integrates
ui_surface: <enum>            # EXTENDED — has-ui | headless
team_sharing: <enum>          # EXTENDED — solo | team-shared
```

**Field classes (REQ-PHB-016).** The 8 fields are partitioned by *how they are
collected*, not by whether they may be empty:

| Class | Fields | Collected in | Early-exit gate? |
|-------|--------|--------------|------------------|
| REQUIRED (base) | `domain`, `goal`, `constraints`, `scope` | Stage A (Rounds 1-3, clarity-scored, `max_rounds`-capped) | YES — Stage A cannot exit early while any is unanswered (REQ-PHB-002 / REQ-PHB-017) |
| EXTENDED | `verification`, `ui_surface`, `external_systems`, `team_sharing` | **Stage B — the mandatory Round 4** (REQ-PHB-013), which always runs | N/A — Stage B is exempt from the cap, the early exit, and clarity scoring (REQ-PHB-014) |

Field population: `doc-generation.md` maps the prose answers recorded in
`.moai/project/interview.md` onto these 8 fields — the vision / goal answer →
`goal`, the domain / problem answer → `domain`, the scope answer → `scope`, the
constraints answer → `constraints` (all four from Stage A, per each host's
base-field coverage mapping — REQ-PHB-018), and the four extended interview axes →
`verification` / `external_systems` / `ui_surface` / `team_sharing` respectively
(all four from the mandatory Stage B round).

The full machine-verifiable AC matrix (AC-PHB-001 … AC-PHB-020) lives in
`acceptance.md` (SSOT). Every REQ above maps to at least one AC; preservation
REQs (REQ-PHB-010) map to a NO-WRITE / absence assertion; the v0.2.0 amendment REQs
(REQ-PHB-013 … REQ-PHB-019) map to the REACHABILITY ACs (AC-PHB-015 … AC-PHB-020),
which assert that Round 4 is *reachable*, not merely that its axis tokens are
*present* — the exact failure mode the original 14-AC matrix missed.

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

- `.claude/skills/moai/workflows/project/mode-detection.md` — Phase 0.3
  (new-project) interview host (adaptive conversion target).
- `.claude/skills/moai/workflows/project/codebase-analysis.md` — Phase 1.5
  (existing-project) interview host (adaptive conversion target).
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
