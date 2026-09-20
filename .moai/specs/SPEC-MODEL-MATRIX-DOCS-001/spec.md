---
id: SPEC-MODEL-MATRIX-DOCS-001
title: "4-locale documentation + guard realignment and full verification"
version: "0.1.0"
status: in-progress
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "docs-site + README + .claude/rules + internal/spec + internal/template"
lifecycle: spec-anchored
tags: "model-profile, documentation, i18n, haiku-guard, verification, split-successor"
tier: L
era: V3R6
depends_on: [SPEC-MODEL-MATRIX-CORE-001, SPEC-MODEL-MATRIX-CONFIG-001, SPEC-MODEL-MATRIX-SURFACES-001]
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-AGENT-ARCH-V2-001]
---

# SPEC-MODEL-MATRIX-DOCS-001 — Documentation and guard realignment

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from `SPEC-MODEL-PROFILE-MATRIX-002` (card t1036) on the M6 + M7 seam, after that SPEC exceeded the Tier L REQ/AC ceilings (72/64 against 25/25). Carries M6 (4-locale documentation) and M7 (guard realignment and full verification), plus the two cross-cutting requirements that attach to M6: the Fable-unavailability fallback and the GLM/CG pairing note. Requirements re-numbered `REQ-MPMD-*`, criteria `AC-MPMD-*`; nothing was dropped. Tier L assigned on REQ count (21 > the Tier M ceiling of 16), AC count (21 > 16), and file scope far above 15 — thirteen docs-site pages across four locales, four README copies, two byte-parity rule twins, and the lint rule with its tests. Status is `in-progress`, not `draft`: M6 already landed under the original SPEC ID — see § Inherited run-phase state. | manager-spec |

## Position in the chain

Last of four; the only successor depending on all three others.

```
SPEC-MODEL-MATRIX-CORE-001 ──┬─→ SPEC-MODEL-MATRIX-CONFIG-001 ──┐
                             │                                  ├─→ SPEC-MODEL-MATRIX-DOCS-001
SPEC-MODEL-MATRIX-SURFACES-001 ───────────────────────────────── ┘
```

- On `SPEC-MODEL-MATRIX-CORE-001`: its 33-cell matrix is what these pages render, and its S0 record is the sole source every benchmark figure must trace to.
- On `SPEC-MODEL-MATRIX-CONFIG-001`: the two-channel effort explanation describes what M3 builds, and M7 removes the two guard surfaces M2 deletes.
- On `SPEC-MODEL-MATRIX-SURFACES-001`: M7 adds guard surfaces that M5 must first clear of `haiku`.

## Inherited run-phase state

[HARD] **M6 has landed.** It was implemented under the original SPEC ID `SPEC-MODEL-PROFILE-MATRIX-002` and shipped in squash `31da99a7b` (PR #1163): thirteen docs-site pages per locale (profile-matrix, no-haiku-3tier, model-policy, agent-guide, faq, cli, init, update, introduction, config-sections, tokenomics-overview, what-is-moai-adk, init-wizard) plus four README copies. The verbatim landing evidence is inherited into `progress.md` §E.2.

[HARD] **Its blocking precondition lives in a predecessor and is UNDISCHARGED.** S0 — the leaderboard verification record — is owned by `SPEC-MODEL-MATRIX-CORE-001` and was never performed. M6 proceeded on the user-supplied per-effort measurements instead. Every benchmark figure now live in these pages therefore traces to an unverified reading, not to a record.

This inverts the intended order. S0 was designed as a gate ahead of M6; it is now a correction behind it. The consequence for this SPEC is concrete and is stated as a requirement (REQ-MPMD-011): when S0 finally produces a confirmed figure that differs, the correction must reach the **shipped pages**, not merely a delta table in a predecessor's `progress.md`. No acceptance criterion can retroactively close the window during which live documentation carried unverified figures; that residual is recorded rather than closed.

**M7 has not started.** The haiku-residual rule still scans the retired surfaces, the consolidated matrix property test does not exist, and no cross-platform build was run.

**Zero of the inherited acceptance criteria are formally verified.** The landing carried toolchain-level verification (build / test / lint / docs-build / matrix cross-check), never an AC-level PASS/FAIL matrix.

---

## §A Context

### §A.1 The naming inversion must be disclosed, not silently shipped

The `max` / `medium` / `low` profile names imply an ordering "higher = stronger and costlier". Under the leaderboard readings that ordering is **false**: the Max profile is simultaneously cheaper per task and higher-scoring than the Medium profile. The axis is a **subscription-tier access axis**, not a performance-grade axis.

Renaming the profiles was rejected — it would break `llm.profile` values, the CLI flag, wizard values, config files in the field, and every documentation surface at once. Disclosure achieves the same reader outcome at a fraction of the blast radius. That disclosure is this SPEC's single most important output.

### §A.2 The Max-Opus rationale is a separate statement, easily conflated

The Max column assigns Opus to `manager-develop` and `builder-harness` deliberately, because those agents' failures are expensive to recover from. This is a quality-first choice, not a benchmark optimum — the leaderboard would favour Fable for both. Without this statement a reader who has just absorbed §A.1 reads the Max column's Opus cells as an error.

### §A.3 Documentation surfaces currently contradict each other

- `.claude/rules/moai/development/model-policy.md` is self-contradictory: one section asserts "all workers are Sonnet 5 fixed across all tiers" and points at `model_routing_profiles` as the 3-tier config SSOT, while the next section describes the modern `llm.profile` resolver. Both cannot hold — and the stale section names a block `SPEC-MODEL-MATRIX-CONFIG-001` deletes, so it will reference a nonexistent artifact.
- The Chinese copy of `multi-llm/model-policy.md` retains a Haiku column with per-agent haiku assignments.
- The four locale copies of that page have diverged in shape.

The resolution is structural rather than editorial: designate `advanced/profile-matrix.md` as authoritative (it renders the matrix, the disclosure, and the rationale) and reduce `multi-llm/model-policy.md` to a narrative page with no per-agent table of its own. Deleting that table is what makes the contradiction impossible to recur, and it is what lets the zh Haiku column be removed rather than corrected.

### §A.4 The guard must be realigned in both directions at once

The haiku-residual rule's surface list gains the two surfaces `SPEC-MODEL-MATRIX-SURFACES-001` clears, and loses the two `SPEC-MODEL-MATRIX-CONFIG-001` deletes. A guard that scans for a deleted artifact is not merely dead — it is actively misleading, because a future reader infers the artifact still exists. That is why the removals are coupled to the additions rather than deferred.

---

## §B Requirements (GEARS)

### §B.1 M6 — 4-locale documentation

- **REQ-MPMD-001** (Ubiquitous) The README benchmark table shall be replaced with S0-confirmed v1.1 figures in all four locale files.
- **REQ-MPMD-002** (Ubiquitous) The `advanced/profile-matrix.md` page shall present the 33-cell per-agent matrix in all four locales, and its agent-group table shall be removed.
- **REQ-MPMD-003** (Ubiquitous) The `advanced/no-haiku-3tier.md` benchmark table shall be replaced with S0-confirmed v1.1 figures in all four locales.
- **REQ-MPMD-004** (Unwanted) The `multi-llm/model-policy.md` page shall not assert the retired "every worker agent is pinned to Sonnet 5" policy, in any locale.
- **REQ-MPMD-005** (Unwanted) The `multi-llm/model-policy.md` Chinese copy shall not contain a Haiku column or any per-agent haiku assignment.
- **REQ-MPMD-006** (Ubiquitous) The four locale copies of `multi-llm/model-policy.md` shall share one table shape, resolving the Japanese copy's divergence.
- **REQ-MPMD-007** (Ubiquitous) The stale "all workers Sonnet fixed" tier table in `.claude/rules/moai/development/model-policy.md` shall be reconciled with the modern resolver description in the same file.
- **REQ-MPMD-008** (Event-driven) **When** `.claude/rules/moai/development/model-policy.md` is edited, its byte-identical template twin shall be edited in the same commit.
- **REQ-MPMD-009** (Ubiquitous) The model enum in `agent-authoring.md` and in `dynamic-workflows.md` shall include `fable`.
- **REQ-MPMD-010** (Ubiquitous) Every documentation reference to the retired `--model-policy` flag name shall be updated to the current `--profile` flag name.
- **REQ-MPMD-011** (Ubiquitous) The documentation shall state explicitly that the `max` / `medium` / `low` names denote **subscription-tier access**, not performance grade, and that the Max profile is both cheaper and higher-scoring than the Medium profile under the S0-confirmed data. **Where a figure already published differs from the S0-confirmed value, the published page shall be corrected**, not merely the predecessor's delta table.
- **REQ-MPMD-012** (Ubiquitous) The documentation shall state that the Max profile's Opus assignments are a deliberate quality-first choice for high-failure-cost work and not a benchmark optimum.
- **REQ-MPMD-013** (Event-driven) **When** any documentation surface in this milestone is edited, all four locale copies of that surface shall be edited in the same change.
- **REQ-MPMD-014** (Capability gate) **Where** the `fable` model is unavailable in the runtime environment, the documentation shall state the fallback the user is expected to take, since the Max profile depends on Fable for six of its eleven cells.
- **REQ-MPMD-015** (Ubiquitous) The documentation shall record that under a GLM backend the increased Fable usage collapses to `glm-5.2`, whose observed step count is the least step-efficient of the three models, and that CG-mode profile pairings warrant review on that basis.

### §B.2 M7 — guard realignment and verification

- **REQ-MPMD-016** (Ubiquitous) The haiku-residual lint rule's surface list shall include the web-console agent-frontmatter model option set.
- **REQ-MPMD-017** (Ubiquitous) The haiku-residual lint rule's surface list shall include the v4manifest tier-suggestion table.
- **REQ-MPMD-018** (Event-driven) **When** the haiku-residual rule's surfaces change, its retired-surface entries (`model_routing_profiles`, `validRoutingModels`) shall be removed in the same change so the rule does not scan for a block that no longer exists.
- **REQ-MPMD-019** (Ubiquitous) A property test shall assert the matrix invariants (33 cells, closed model set, closed effort set, no `haiku`, no in-matrix `inherit`, profile-invariant trio, display-order agreement).
- **REQ-MPMD-020** (Ubiquitous) The full Go test suite shall pass.
- **REQ-MPMD-021** (Ubiquitous) The build shall succeed for the project's release target platforms.

---

## §C Exclusions

### Out of Scope — the matrix, the config axis, and the surfaces

- Authoring or amending the 33-cell matrix, the resolver, and the leaderboard verification record. All belong to `SPEC-MODEL-MATRIX-CORE-001`. This SPEC renders the matrix and consumes the record; it produces neither.
- The `model_routing_profiles` retirement, the `profiles:` mirror removal, and both effort channels. All belong to `SPEC-MODEL-MATRIX-CONFIG-001`. REQ-MPMD-011 and the two-channel content requirements state *what* is documented; the behaviour itself is built there.
- The init wizard question and the web console cleanup. Both belong to `SPEC-MODEL-MATRIX-SURFACES-001`. M7 adds guard coverage for the surfaces that SPEC clears; it does not clear them.

### Out of Scope — infographic regeneration

- Regenerating `assets/images/readme/tokenomics-harness-{en,ko,ja,zh}.png`. The images may embed superseded benchmark figures; assessing and regenerating them is deferred to a follow-up, and M6 records the assessment result rather than performing the regeneration.

### Out of Scope — profile renaming

- Renaming the `max` / `medium` / `low` profile vocabulary to names matching the inverted ordering; adding a fourth profile column; reintroducing the retired `plan_type` axis. §A.1 requires disclosure, not renaming.

### Out of Scope — model behaviour beyond documentation

- Changing the GLM effort-collapse table or the GLM model-alias map. REQ-MPMD-015 records their consequence; it does not change them.
- Changing the v4manifest `agentTiers` badge assignment table.

### Out of Scope — the haiku alias exemptions

- The `haiku` alias in the GLM model map (`glm.models.haiku`), an exempt surface of the haiku-residual rule by design, and the rule's three other standing exemptions.

---

## §D Constraints

| # | Constraint | Source |
|---|---|---|
| C-1 | Template tree content must remain internally neutral — no SPEC IDs, REQ tokens, internal dates, or commit SHAs in `internal/template/templates/**`. | CLAUDE.local.md §25 |
| C-2 | `.claude/rules/.../model-policy.md` and its template twin are byte-parity-checked; both sides are edited in one commit. | `internal/template/rule_template_mirror_test.go` |
| C-3 | 4-locale parity is mandatory for every docs-site page and for the README set. | CLAUDE.local.md §17 |
| C-4 | The No-Haiku policy is a HARD gate not skippable via SPEC `lint.skip`. | `internal/spec/lint_haiku_residual.go` |
| C-5 | No benchmark figure may be written to a documentation surface before the S0 record exists. | predecessor REQ-MPMC-001 |
| C-6 | A parallel session may be active on this shared checkout; changes are committed with explicit pathspecs. | project operating practice |

---

## §E Decisions

| # | Decision | Rationale |
|---|---|---|
| D-1 | Profile names retained; the inverted ordering is disclosed in documentation. | User-confirmed. Renaming has a far larger blast radius than the problem it solves. |
| D-2 | `advanced/profile-matrix.md` is authoritative; `multi-llm/model-policy.md` becomes narrative with no per-agent table. | Editing two contradicting pages toward each other leaves the contradiction able to recur; deleting one table makes it structurally impossible. |
| D-3 | The stale rules-file tier section is deleted, not updated. | It names an artifact `SPEC-MODEL-MATRIX-CONFIG-001` deletes; updating it would preserve a section whose premise is gone. |
| D-4 | Guard surfaces are added and removed in the same change. | A guard scanning a deleted artifact implies to a future reader that the artifact still exists (§A.4). |
| D-5 | 4-locale execution routes through the `oss-docs` harness. | Its canonical-locale chain and same-PR 4-locale obligation are the existing mechanism for exactly this failure mode; hand-editing 4 locales × 6 surfaces is where parity breaks. |
| D-6 | Where an S0 correction contradicts an already-published figure, the published page is corrected rather than the delta being recorded only upstream. | S0 is remediation here, not prevention; a delta that stops at `progress.md` leaves the wrong number live. |

---

## §F Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | Benchmark figures are already live and trace to an unverified reading, because S0 was never discharged before M6 shipped. | High | REQ-MPMD-011's correction clause plus D-6. The window itself cannot be closed retroactively and is recorded as residual. |
| R-2 | Documentation scale: 4 locales × (README + 3 doc page families + 2 rules files). One missed locale breaks parity. | Medium | REQ-MPMD-013 + routing through the `oss-docs` harness. |
| R-3 | The Max profile leans on Fable for 6 of 11 cells; environments without Fable access break. | High | REQ-MPMD-014 — documented fallback path. |
| R-4 | Under GLM, heavier Fable usage collapses to `glm-5.2`, the least step-efficient of the three models. | Medium | REQ-MPMD-015 — record and flag CG-mode pairings for review. |
| R-5 | M7's guard additions land before `SPEC-MODEL-MATRIX-SURFACES-001` clears the surfaces, turning the guard into an immediate failure against shipped code. | Medium | The `depends_on` edge on SURFACES; verify both surfaces are clear before adding them. |
| R-6 | M7's guard removals land before `SPEC-MODEL-MATRIX-CONFIG-001` deletes the artifacts, removing coverage that is still needed. | Medium | The `depends_on` edge on CONFIG; verify both artifacts are gone before removing their surfaces. |
| R-7 | A reader absorbs the naming-inversion disclosure and then reads the Max column's Opus cells as an error. | Medium | REQ-MPMD-012 — the Max-Opus rationale is a separate, adjacent statement (§A.2). |

---

## §G Traceability

| Unit | REQ range | AC range |
|---|---|---|
| M6 (incl. the two cross-cutting documentation items) | REQ-MPMD-001 … 015 | AC-MPMD-001 … 015 |
| M7 | REQ-MPMD-016 … 021 | AC-MPMD-016 … 021 |

21 requirements, 21 acceptance criteria — both within the Tier L ceilings of 25 and 25. Full AC↔REQ mapping: `acceptance.md` §D.

---

## §H Cross-References

- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — the retired predecessor stub carrying the full split mapping table.
- `.moai/specs/SPEC-MODEL-MATRIX-CORE-001/` — predecessor; owns the matrix and the undischarged S0 record.
- `.moai/specs/SPEC-MODEL-MATRIX-CONFIG-001/` — predecessor; builds the two channels this SPEC documents, deletes the guard surfaces this SPEC removes.
- `.moai/specs/SPEC-MODEL-MATRIX-SURFACES-001/` — predecessor; clears the surfaces this SPEC adds to the guard.
- `.claude/rules/moai/development/model-policy.md` — the self-contradictory rule file reconciled here.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — one of the two model enums gaining `fable`.
