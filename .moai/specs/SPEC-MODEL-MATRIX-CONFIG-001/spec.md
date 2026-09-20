---
id: SPEC-MODEL-MATRIX-CONFIG-001
title: "Config retirement + effort actualization — 36-cell axis, profiles mirror, both effort channels"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/config + internal/cli + internal/web + internal/settings/agentfm"
lifecycle: spec-anchored
tags: "model-profile, effort-injection, config-retirement, migration, split-successor"
tier: L
era: V3R6
depends_on: [SPEC-MODEL-MATRIX-CORE-001]
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-MODEL-TIER-PLANTYPE-001]
---

# SPEC-MODEL-MATRIX-CONFIG-001 — Config retirement and effort actualization

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from `SPEC-MODEL-PROFILE-MATRIX-002` (card t1036) on the M2 + M3 seam, after that SPEC exceeded the Tier L REQ/AC ceilings (72/64 against 25/25). Carries M2 (retire the 36-cell axis and the `llm.yaml profiles:` mirror) and M3 (effort actualization on both channels), plus the two cross-cutting requirements that attach to them: the no-silent-drop prohibition (M2) and the dev-repo self-application no-op (M3). Requirements re-numbered `REQ-MPME-*`, criteria `AC-MPME-*`; nothing was dropped. Tier L assigned on REQ count (22 > the Tier M ceiling of 16) and on file scope (>15 files across `internal/config`, `internal/cli`, `internal/web`, `internal/settings/agentfm`, two config mirrors, and their tests). Status `draft`: nothing in this SPEC has landed. | manager-spec |

## Position in the chain

Second of four.

```
SPEC-MODEL-MATRIX-CORE-001 ──┬─→ SPEC-MODEL-MATRIX-CONFIG-001 ──┐
                             │                                  ├─→ SPEC-MODEL-MATRIX-DOCS-001
SPEC-MODEL-MATRIX-SURFACES-001 ───────────────────────────────── ┘
```

Depends on `SPEC-MODEL-MATRIX-CORE-001`: M2's "the Go constant is the sole matrix SSOT" is only meaningful once that constant is the 33-cell direct matrix, and M3's effort application reads cells from it. `SPEC-MODEL-MATRIX-SURFACES-001` is independent of this SPEC and may run in parallel. `SPEC-MODEL-MATRIX-DOCS-001` depends on this SPEC because its two-channel documentation describes what M3 builds, and its guard realignment removes surfaces M2 deletes.

**Nothing in this SPEC has landed.** Its predecessor's M1 and the downstream M6 shipped in squash `31da99a7b`; M2 and M3 were never started. `model_routing_profiles` is still present in `workflow.yaml`.

---

## §A Context

### §A.1 The 36-cell axis has no production consumer

`model_routing_profiles` in `workflow.yaml`, the `RouteModelFor(specTier, phase, perfTier)` accessor, its `ModelRoutingProfiles` type, and the `validRoutingModels` map form a 36-cell Tier×Phase routing axis. Every `RouteModelFor` call outside its own test file is within `model_routing.go` itself. The `@MX:ANCHOR` calling it "the spawn-time cost-routing accessor" describes an intended consumer that was never wired. Retiring it rather than migrating it is safe because its runtime impact is nil.

### §A.2 The `profiles:` mirror is the third and fourth copy of one literal

The matrix literal lives in four places: the Go constant, both `llm.yaml` copies, and the fidelity test's `want` map. Per-agent overrides (`llm.agent_overrides`) already provide the user-editability the mirror was justified by, at one cell of granularity instead of one group. Removing the mirror leaves the Go constant and its fidelity test.

Removal is not free: a user may have customized their `profiles:` block. Silently discarding that customization is prohibited (REQ-MPME-009).

### §A.3 Effort injection is path-dependent (SPEC-001 DECISION-001 correction)

`SPEC-MODEL-PROFILE-MATRIX-001` DECISION-001 states that effort cannot be injected per-spawn and characterizes the Workflow channel as "prompt-level". **That is inaccurate.** Verified from the live tool schemas:

| Channel | `model` | `effort` | Consequence |
|---|---|---|---|
| `Agent` tool (sub-agent delegation, the standard path) | runtime arg, `enum: sonnet\|opus\|haiku\|fable` | **no parameter exists** | frontmatter `effort:` is the effective effort → rewrite required |
| `Workflow` tool `agent()` (dynamic-workflow) | `opts.model` | `opts.effort ∈ {low, medium, high, xhigh, max}` | structured parameter, injectable directly |

The Workflow form is `agent(prompt, {agentType: 'manager-develop', model: 'opus', effort: 'xhigh'})` — a **structured option**, not prompt-level steering.

The matrix effort is therefore consumed through **two channels**, and requirements cover both: frontmatter rewrite for the Agent-tool path, and direct `opts.effort` injection for the Workflow path, which needs a matrix-lookup route reachable from workflow scripts.

### §A.4 Frontmatter rewrite revival — why it is safe this time

`SPEC-MODEL-PROFILE-MATRIX-001` REQ-MPM-024 retired `ApplyTierProfile` because it rewrote each shipped agent's `model:` **and** `effort:`, re-introducing a concrete-model pin (the `[1m]` hazard) and producing large diffs across both the deployed and template trees. The stated cause of retirement is the `model:` pin, not the `effort:` write.

The revival is deliberately narrower; four mitigations bound it, and they must hold **simultaneously**:

1. `effort:` line only — `model:` never written by this path.
2. Deployed tree only — the template tree is immutable and keeps a fixed Medium-profile baseline.
3. Reapply **after** template deploy on `moai update`, so a profile survives an update.
4. Agents carrying an `llm.agent_overrides[<agent>].effort` are excluded, matching resolver precedence.

The precedent for (1) and (2) already exists in-tree: `internal/settings/agentfm/agentfm.go` `Patch(path, model, effort string, deleteEffort bool)` is a frontmatter-only, live-files-only editor with no template dual-write.

---

## §B Requirements (GEARS)

### §B.1 M2 — retire the 36-cell axis and the config mirror

- **REQ-MPME-001** (Ubiquitous) The `model_routing_profiles` block shall be removed from both `workflow.yaml` copies (project tree and template tree).
- **REQ-MPME-002** (Ubiquitous) The `RouteModelFor` accessor, its `ModelRoutingProfiles` type, and its config-load validators shall be removed.
- **REQ-MPME-003** (Capability gate) **Where** removal of `RouteModelFor` would orphan its tests, those tests shall be deleted rather than retargeted, since the axis has no production consumer.
- **REQ-MPME-004** (Ubiquitous) The `profiles:` block shall be removed from both `llm.yaml` copies, leaving the Go constant as the sole matrix SSOT.
- **REQ-MPME-005** (Ubiquitous) After M2 the matrix literal shall exist in exactly two places: the Go constant and its fidelity test.
- **REQ-MPME-006** (Event-driven) **When** `moai update` runs against a project whose `llm.yaml` still carries a non-empty `profiles:` block, the updater shall detect it and shall not discard the user's customization silently.
- **REQ-MPME-007** (Event-driven) **When** a detected `profiles:` customization is representable as per-agent overrides, the updater shall migrate it into `llm.agent_overrides`; **when** it is not representable, the updater shall emit a warning naming the affected profile column and cells.
- **REQ-MPME-008** (Ubiquitous) The config loader shall not fail on a legacy `llm.yaml` that still carries `profiles:` or a legacy `workflow.yaml` that still carries `model_routing_profiles`; the unknown block shall be tolerated as inert.
- **REQ-MPME-009** (Unwanted) The system shall not silently drop a user's existing `llm.yaml profiles:` customization; removal without either migration or a warning is prohibited.

### §B.2 M3 — effort actualization (both channels)

- **REQ-MPME-010** (Ubiquitous) The system shall provide an agent-effort application function that writes each deployed MoAI agent's frontmatter `effort:` value from the active profile's matrix cell.
- **REQ-MPME-011** (Ubiquitous) The effort application function shall rewrite the `effort:` line only and shall not write the `model:` line.
- **REQ-MPME-012** (Unwanted) The effort application function shall not modify any file under `internal/template/templates/.claude/agents/`.
- **REQ-MPME-013** (Ubiquitous) The template tree's agent frontmatter shall carry the Medium-profile effort values as a fixed baseline independent of any project's active profile.
- **REQ-MPME-014** (Event-driven) **When** `moai update` deploys templates, the effort application shall run after deployment completes, so a user's non-default profile survives the update.
- **REQ-MPME-015** (Capability gate) **Where** an agent has an `llm.agent_overrides[<agent>].effort` entry, the effort application function shall exclude that agent from rewrite.
- **REQ-MPME-016** (Ubiquitous) The effort application shall be invoked at the four existing profile-application seams: `moai init`, the `moai update` profile flag path, the `moai update` wizard path, and the web-console profile save path.
- **REQ-MPME-017** (Event-driven) **When** the effort application encounters an agent name with no corresponding file on disk, it shall skip that agent without error.
- **REQ-MPME-018** (Unwanted) The effort application shall not modify agent body bytes; only the frontmatter `effort:` value shall change.
- **REQ-MPME-019** (Ubiquitous) The system shall expose a machine-readable route by which a dynamic-workflow script can obtain an agent's resolved `{model, effort}` under the active profile, for injection as the `Workflow` tool's `opts.model` / `opts.effort`.
- **REQ-MPME-020** (Ubiquitous) The documentation shall state that effort reaches the agent through two distinct channels — frontmatter for the `Agent` tool path (which has no `effort` parameter) and `opts.effort` for the `Workflow` tool path — and shall record that `SPEC-MODEL-PROFILE-MATRIX-001` DECISION-001's "effort cannot be injected per-spawn / Workflow is prompt-level" wording is superseded.
- **REQ-MPME-021** (Ubiquitous) The documentation shall state that `Explore` has no agent file, so its matrix effort is documented intent consumed only through the Workflow path.
- **REQ-MPME-022** (Event-driven) **When** the effort application runs inside this repository itself, it shall be a no-op at the default `medium` profile, because the deployed agent frontmatter values already equal the template baseline.

---

## §C Exclusions

### Out of Scope — the matrix itself

- Authoring, transcribing, or amending the 33-cell matrix, the resolver precedence order, the `Explore` row, or the group removal. All of M1 belongs to `SPEC-MODEL-MATRIX-CORE-001`; this SPEC consumes the matrix it ships.
- The leaderboard verification record (S0). It belongs to `SPEC-MODEL-MATRIX-CORE-001`.

### Out of Scope — user-facing surfaces

- The `moai init` wizard's model-policy question, its option vocabulary, and its translations. They belong to `SPEC-MODEL-MATRIX-SURFACES-001`.
- The web console's agent-frontmatter model selector, the v4manifest tier suggestion, and the `agentfm.tier.desc` / `mp.*` i18n keys. They belong to `SPEC-MODEL-MATRIX-SURFACES-001`. Note the boundary: this SPEC edits `internal/web/agentfm.go` at the **profile-save seam** (REQ-MPME-016) and rewrites its stale `applyPerfTierEdits` comment; SURFACES edits the **model option set** in the same file. Two distinct concerns in one file — sequence them rather than merging them.

### Out of Scope — documentation surfaces and guards

- Every docs-site page, README locale copy, and rules-file edit, including the naming-inversion disclosure. They belong to `SPEC-MODEL-MATRIX-DOCS-001`. REQ-MPME-020 and REQ-MPME-021 state what must be documented about the two channels; DOCS owns where and in which locales it lands.
- The haiku-residual lint rule's surface list. Removing the `model_routing_profiles` and `validRoutingModels` surfaces this SPEC deletes belongs to `SPEC-MODEL-MATRIX-DOCS-001` (M7) — a deliberate coupling recorded in §F R-4.

### Out of Scope — model behavior and routing beyond the matrix

- Changing the GLM effort-collapse table (`low`→thinking-off, `medium`/`high`→reasoning-high, `xhigh`/`max`→reasoning-max) or the `manager-develop` coding-max singleton override.
- Changing the GLM model-alias map (`fable`→`glm-5.2`).
- Live z.ai wire-effectiveness validation of the GLM effort overlay — inherited as pending from `SPEC-MODEL-PROFILE-MATRIX-001` and unchanged here.
- Orchestrator spawn-time routing policy (which agent handles which task).

### Out of Scope — the `glm.models` drift

- Reconciling the pre-existing drift between the two `llm.yaml` copies' `glm.models` blocks (the local copy carries `opus`/`sonnet`/`haiku` alias keys the template lacks). M2 edits the `profiles:` block only; a whole-block sync is prohibited (K-5).

---

## §D Constraints

| # | Constraint | Source |
|---|---|---|
| C-1 | Template tree content must remain internally neutral — no SPEC IDs, REQ tokens, internal dates, or commit SHAs in `internal/template/templates/**`. | CLAUDE.local.md §25 |
| C-2 | Template-First: every template change is made in `internal/template/templates/` first, then `make build`. | CLAUDE.local.md §2 |
| C-3 | The No-Haiku policy is a HARD gate not skippable via SPEC `lint.skip`. | `internal/spec/lint_haiku_residual.go` |
| C-4 | `settings.local.json` and machine-specific values are never written by this work. | CLAUDE.local.md §2 |
| C-5 | A parallel session may be active on this shared checkout; changes are committed with explicit pathspecs. | project operating practice |
| C-6 | A config-load hard failure on a legacy block would break existing projects on upgrade; tolerance is mandatory, not optional. | REQ-MPME-008 |

---

## §E Decisions

| # | Decision | Rationale |
|---|---|---|
| D-1 | 36-cell Tier×Phase axis retired entirely rather than migrated. | User-confirmed. Zero non-test production call sites, so runtime impact is nil. |
| D-2 | `llm.yaml profiles:` removed; the Go constant is the single SSOT. | User-confirmed. Drops matrix literal duplication from 4 copies to 2. |
| D-3 | Frontmatter `effort:` rewrite revived, narrowed to one line, deployed tree only, post-deploy, override-excluded. | User-confirmed. Without it the matrix effort is inert on the `Agent` tool path, which has no `effort` parameter. |
| D-4 | Legacy `profiles:` / `model_routing_profiles` blocks are tolerated as inert rather than rejected on load. | A hard load failure would break existing projects on upgrade. |
| D-5 | Channel B is exposed, not built — `moai model profile --json` already computes `{model, effort}` per agent under the active profile, including the GLM overlay. | Building a second lookup would be a third copy of resolver logic. |

---

## §F Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | Removing `llm.yaml profiles:` silently drops an existing user override. | High | REQ-MPME-006/007/009 — detect, migrate to `agent_overrides`, or warn. |
| R-2 | Frontmatter rewrite reintroduces large diffs, the failure mode that retired the predecessor. | Medium | REQ-MPME-011/012/018 — one line per file, deployed tree only. |
| R-3 | A whole-block sync of the two `llm.yaml` copies imports the local-only `glm.models` alias keys — including `haiku` — into the template. | Medium | M2 edits the `profiles:` block only (K-5). |
| R-4 | The haiku-residual rule keeps scanning `model_routing_profiles` and `validRoutingModels` after M2 deletes them — a guard pointing at a nonexistent artifact implies to a future reader that the artifact still exists. | Medium | The removal is `SPEC-MODEL-MATRIX-DOCS-001`'s M7 obligation; this SPEC must hand it off explicitly rather than assume it. |
| R-5 | This repository's own `.claude/agents/moai/` becomes a rewrite target. | Low | REQ-MPME-022 — no-op at the default `medium` profile; divergence only under a non-default profile, which is a maintainer choice. |
| R-6 | REQ-MPME-022's no-op holds only after the template baseline is re-set by the predecessor's M1 step 7. Treating it as already true produces visible churn that looks like a bug. | Medium | Verify the predecessor's step 7 landed before asserting the no-op (K-1). |
| R-7 | `internal/web/agentfm.go` is edited by both this SPEC (profile-save seam, comment) and `SPEC-MODEL-MATRIX-SURFACES-001` (model option set). | Low | Independent concerns in one file; sequence the two SPECs' edits rather than merging them. |

---

## §G Traceability

| Unit | REQ range | AC range |
|---|---|---|
| M2 (incl. the no-silent-drop cross-cutting item) | REQ-MPME-001 … 009 | AC-MPME-001 … 008 |
| M3 (incl. the dev-repo no-op cross-cutting item) | REQ-MPME-010 … 022 | AC-MPME-009 … 020 |

22 requirements, 20 acceptance criteria — both within the Tier L ceilings of 25 and 25. Full AC↔REQ mapping: `acceptance.md` §D.

---

## §H Cross-References

- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — the retired predecessor stub carrying the full split mapping table.
- `.moai/specs/SPEC-MODEL-MATRIX-CORE-001/` — predecessor; ships the matrix this SPEC consumes.
- `.moai/specs/SPEC-MODEL-MATRIX-DOCS-001/` — successor; documents the two channels and removes the guard surfaces M2 deletes.
- `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/` — origin of the GLM effort overlay retained unchanged.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — Workflow `agent()` effort closed set corroborating §A.3.
