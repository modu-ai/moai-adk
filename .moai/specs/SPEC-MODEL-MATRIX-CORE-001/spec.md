---
id: SPEC-MODEL-MATRIX-CORE-001
title: "Profile matrix core — 33-cell agent-direct matrix + leaderboard verification"
version: "0.1.0"
status: in-progress
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/template + internal/cli/model.go"
lifecycle: spec-anchored
tags: "model-profile, agent-matrix, matrix-ssot, leaderboard-verification, split-successor"
tier: L
era: V3R6
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-AGENT-ARCH-V2-001]
---

# SPEC-MODEL-MATRIX-CORE-001 — Profile matrix core

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from `SPEC-MODEL-PROFILE-MATRIX-002` (card t1036) on the S0 + M1 seam, after that SPEC exceeded the Tier L REQ/AC ceilings (72/64 against 25/25). Carries S0 (leaderboard verification) and M1 (the 33-cell matrix redesign). Requirements re-numbered `REQ-MPMC-*`, criteria `AC-MPMC-*`; nothing was dropped. Tier L assigned on REQ count (18 > the Tier M ceiling of 16) and on file scope (>15 files: the Go SSOT, its tests, two config mirrors, ten agent frontmatter files across two mirrors, and the report consumer). Status is `in-progress`, not `draft`: M1 already landed under the original SPEC ID — see § Inherited run-phase state. | manager-spec |

## Position in the chain

First of four, and one of two roots. `SPEC-MODEL-MATRIX-SURFACES-001` is the other root and proceeds in parallel.

```
SPEC-MODEL-MATRIX-CORE-001 ──┬─→ SPEC-MODEL-MATRIX-CONFIG-001 ──┐
                             │                                  ├─→ SPEC-MODEL-MATRIX-DOCS-001
SPEC-MODEL-MATRIX-SURFACES-001 ───────────────────────────────── ┘
```

Nothing precedes this SPEC. `SPEC-MODEL-MATRIX-CONFIG-001` depends on it because M2 and M3 both rest on the matrix M1 lands; `SPEC-MODEL-MATRIX-DOCS-001` depends on it for both the matrix content its pages render and the S0 record every benchmark figure must trace to.

## Inherited run-phase state

[HARD] **M1 has landed.** It was implemented under the original SPEC ID `SPEC-MODEL-PROFILE-MATRIX-002` and shipped in squash `31da99a7b` (PR #1163): the 33-cell matrix as the Go SSOT, the `llm.yaml` and `workflow.yaml` mirrors, ten agent frontmatter files across both mirrors, the `model-policy.md` and `agent-authoring.md` rule twins, and the affected test fixtures. The verbatim landing evidence is inherited into `progress.md` §E.2.

[HARD] **S0 has NOT been discharged**, and this is the SPEC's central inherited contradiction. S0 blocks `SPEC-MODEL-MATRIX-DOCS-001` (M6), which also already landed in the same squash. M1 and M6 proceeded on the user-supplied per-effort measurements instead; S0's own verification was never performed. The blocking precondition is therefore **undischarged on already-shipped documentation**, and closing it is remediation rather than prevention. This SPEC owns S0; DOCS records the same contradiction from the consuming side.

[HARD] **Two M1 plan steps were never executed** even though M1 is recorded as landed: `plan.md` step 2 (delete the six group constants and `agentGroupMembership`) and step 3 (delete `AgentGroup`). Measured in this tree: `internal/template/profile_matrix.go` declares `agentGroupMembership` at line 209 and defines `AgentGroup` at line 464, and `internal/web/agentfm.go:491` calls `template.AgentGroup` as a live gate. The disposition — whether the landing record is wrong or the plan is stale — is owned by card **t1037** and is NOT decided here. This SPEC records the unexecuted remainder as its own and does not remove the code.

**Zero of the inherited acceptance criteria are formally verified.** `internal/` carries no `REQ-MPM2` or `AC-MPM2` markers, so requirement-to-code traceability is unestablished for the landed work.

---

## §A Context

`SPEC-MODEL-PROFILE-MATRIX-001` shipped a 3-column profile axis (`max` / `medium` / `low`) resolved through six agent groups. Two forces broke that abstraction.

### §A.1 Force 1 — the leaderboard inverted the naming semantics

The DeepSWE leaderboard v1.1 (113 tasks) reports per-model, per-effort rows. User-supplied readings for the three models of interest:

| Model [effort] | Per-task cost | Pass@1 | Output tokens | Agent steps |
|---|---|---|---|---|
| Fable 5 [low] | $3.76 | 60% | 25k | 38 |
| Opus 4.8 [high] | $4.28 | 52% | 50k | 73 |
| GLM 5.2 [max] | $3.92 | 42% *or* 45% (conflicting readings) | 78k | 129 |

**These four-metric readings are UNVERIFIED user-supplied inputs.** They are the SPEC's design input, not observed fact. S0 confirms them against the live source before any documentation number is committed. See `research.md` §A.

If the readings hold, Fable 5 at `low` effort leads on all three cost axes while scoring highest — a semantic inversion, because the profile names imply "higher = stronger and costlier" and that ordering is false. The axis is a **subscription-tier access axis**, not a performance-grade axis. Disclosure of the inversion is owned by `SPEC-MODEL-MATRIX-DOCS-001`; this SPEC owns only the verification record the disclosure rests on.

### §A.2 Force 2 — the group abstraction broke

The per-agent cells in §A.4 split two of the six groups: `spec_auditors` splits under `medium` (`manager-spec` takes `opus/high` while the two auditors take `opus/xhigh`), and `design_harness_e2e` splits three ways under `max`. Only two same-cell pairs survive across all three columns. A layer partitioning 11 agents into 6 groups — 4 of them singletons, 2 no longer holding — is dead weight: an indirection hop, a second literal, and a display column carrying no information.

### §A.3 The 33-cell matrix (authoritative — settled design input, MUST NOT be re-derived)

| agent | max | medium | low |
|---|---|---|---|
| manager-spec | fable / low | opus / high | opus / low |
| plan-auditor | fable / low | opus / xhigh | opus / low |
| sync-auditor | fable / low | opus / xhigh | opus / low |
| manager-develop | opus / xhigh | opus / high | sonnet / medium |
| super-advisor | fable / medium | fable / low | opus / medium |
| manager-design | fable / low | opus / high | sonnet / medium |
| builder-harness | opus / high | opus / high | sonnet / medium |
| e2e-tester | fable / low | opus / high | sonnet / medium |
| manager-docs | sonnet / medium | sonnet / medium | sonnet / medium |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / medium | sonnet / medium | sonnet / medium |
| *(unmapped user agents)* | inherit | inherit | inherit |

Properties asserted as acceptance criteria: 11 mapped agents × 3 profiles = 33 cells; zero `haiku`; models ⊆ `{fable, opus, sonnet}`; efforts ⊆ `{low, medium, high, xhigh}`; `inherit` never inside the matrix; three profile-invariant agents (`manager-docs`, `manager-git`, `Explore`).

The Max column's Opus assignments are a deliberate quality-first choice for high-failure-cost work, NOT a benchmark optimum. Stating that in documentation is `SPEC-MODEL-MATRIX-DOCS-001`'s obligation.

### §A.4 `Explore` changes the resolver contract

Today `Explore` is unmapped: `ResolveAgentModelEffort` returns `{inherit, ""}, hasGroup=false`, and one test pins `Explore` and `some-user-agent` to that same behaviour. Under the new matrix `Explore` is an explicit mapping (`sonnet/medium` in all three columns) while unmapped user agents still resolve to `inherit`. The test must be split.

`Explore` is an Anthropic built-in with **no agent file on disk**. Per-spawn `model` injection is the only applicable channel for it; its `medium` effort is documented intent, consumed only through the Workflow path owned by `SPEC-MODEL-MATRIX-CONFIG-001`.

---

## §B Requirements (GEARS)

### §B.1 S0 — leaderboard verification (blocking precondition)

- **REQ-MPMC-001** (State-driven) **While** the v1.1 leaderboard readings in §A.1 remain unconfirmed against the live source, the implementation shall not commit any benchmark figure to a documentation surface.
- **REQ-MPMC-002** (Ubiquitous) The verification record shall resolve the GLM 5.2 Pass@1 conflict (42% vs 45%) to a single canonical value, stating the effort level the row belongs to and the reported error bar.
- **REQ-MPMC-003** (Ubiquitous) The verification record shall pin, per model, all four metrics (per-task cost, Pass@1, output tokens, agent steps) together with the effort level of the row they were read from, and this record shall be the single source for every number written by `SPEC-MODEL-MATRIX-DOCS-001`.
- **REQ-MPMC-004** (Event-driven) **When** a confirmed metric differs from the §A.1 user-supplied reading, the implementation shall use the confirmed value and record the delta in `progress.md` rather than silently substituting it.

### §B.2 M1 — 33-cell matrix redesign

- **REQ-MPMC-005** (Ubiquitous) The profile matrix shall map `profile → agent → {model, effort}` directly, containing exactly the 33 cells transcribed in §A.3.
- **REQ-MPMC-006** (Ubiquitous) The group constants `GroupSpecAuditors`, `GroupDevelop`, `GroupAdvisor`, `GroupDesignHarnessE2E`, `GroupDocs`, `GroupGit` shall not exist after M1.
- **REQ-MPMC-007** (Ubiquitous) The `agentGroupMembership` agent→group table shall not exist after M1.
- **REQ-MPMC-008** (Ubiquitous) The `AgentGroup(agent)` accessor and every consumer of its return value shall be removed or retargeted to the direct per-agent lookup.
- **REQ-MPMC-009** (Ubiquitous) The matrix shall carry an explicit `Explore` row resolving to `sonnet / medium` in all three profile columns.
- **REQ-MPMC-010** (Event-driven) **When** the resolver is queried for an agent absent from the matrix, it shall return the `inherit` sentinel with a false membership flag, so the caller skips model injection.
- **REQ-MPMC-011** (Ubiquitous) The resolver shall preserve the existing precedence order: `llm.agent_overrides[agent]` → active-profile matrix cell → unknown-profile fallback to the `medium` column → `inherit` for unmapped agents.
- **REQ-MPMC-012** (Ubiquitous) The matrix shall contain zero cells whose model is `haiku`.
- **REQ-MPMC-013** (Ubiquitous) Every matrix cell's model shall be a member of `{fable, opus, sonnet}`.
- **REQ-MPMC-014** (Ubiquitous) Every matrix cell's effort shall be a member of `{low, medium, high, xhigh}`.
- **REQ-MPMC-015** (Unwanted) The literal `inherit` shall not appear as a model value inside the matrix itself.
- **REQ-MPMC-016** (Ubiquitous) The agents `manager-docs`, `manager-git`, and `Explore` shall resolve to an identical `{model, effort}` pair across all three profile columns.
- **REQ-MPMC-017** (Event-driven) **When** the matrix is amended, the existing resolver tests pinning the retired group vocabulary shall be amended in the same change so that no test references a removed group constant.
- **REQ-MPMC-018** (Ubiquitous) The `Explore`-and-unmapped-agent inherit test shall be split so that `Explore` asserts the explicit `sonnet / medium` mapping while an arbitrary user-agent name asserts the `inherit` fallback.

---

## §C Exclusions

### Out of Scope — configuration retirement and effort actualization

- Removing the `model_routing_profiles` block, `RouteModelFor`, or the `llm.yaml profiles:` mirror. All of M2 belongs to `SPEC-MODEL-MATRIX-CONFIG-001`.
- Writing agent frontmatter `effort:` values from the active profile, and the Workflow-path `opts.effort` lookup route. All of M3 belongs to `SPEC-MODEL-MATRIX-CONFIG-001`.

### Out of Scope — user-facing surfaces

- The `moai init` wizard's model-policy question and the web console cleanup. Both belong to `SPEC-MODEL-MATRIX-SURFACES-001`.

### Out of Scope — documentation and guards

- Every documentation surface, in every locale, including the naming-inversion disclosure and the Max-Opus rationale. They belong to `SPEC-MODEL-MATRIX-DOCS-001`, which consumes the S0 record this SPEC produces.
- The haiku-residual lint rule's surface list and the matrix property test. Both belong to `SPEC-MODEL-MATRIX-DOCS-001` (M7).

### Out of Scope — profile renaming

- Renaming `max` / `medium` / `low` to names matching the inverted ordering; adding a fourth column; reintroducing the retired `plan_type` axis. §A.1 requires disclosure, not renaming.

### Out of Scope — t1037's question

- Deciding whether the landed-M1 record is wrong or the M1 plan is stale, and removing `agentGroupMembership` / `AgentGroup` on that basis. Card t1037 owns the disposition; this SPEC records the remainder and leaves the code in place.

### Out of Scope — agent catalog

- Adding, removing, or renaming any of the 11 retained agents; changing agent `tools:`, `skills:`, or body content.

---

## §D Constraints

| # | Constraint | Source |
|---|---|---|
| C-1 | Template tree content must remain internally neutral — no SPEC IDs, REQ tokens, internal dates, or commit SHAs in `internal/template/templates/**`. | CLAUDE.local.md §25 |
| C-2 | Template-First: every template change is made in `internal/template/templates/` first, then `make build`. | CLAUDE.local.md §2 |
| C-3 | The No-Haiku policy is a HARD gate not skippable via SPEC `lint.skip`. | `internal/spec/lint_haiku_residual.go` |
| C-4 | The `Agent` tool model enum is `sonnet\|opus\|haiku\|fable`; `inherit` is not an injectable runtime value, so the unmapped-agent path must skip injection rather than pass `inherit`. | live tool schema |
| C-5 | A parallel session may be active on this shared checkout; changes are committed with explicit pathspecs. | project operating practice |

---

## §E Decisions

| # | Decision | Rationale |
|---|---|---|
| D-1 | Group system abolished; direct `profile → agent → {model, effort}` mapping. | User-confirmed. 4 of 6 groups were already singletons; the remaining 2 are split by the new cells. |
| D-2 | `Explore` becomes an explicit matrix row while unmapped user agents keep `inherit`. | The two cases were conflated by a single test; they are semantically distinct. |
| D-3 | S0 is a blocking precondition, not a parallel task. | Every documentation number depends on it; committing unverified figures would violate the no-unobserved-claim invariant. |
| D-4 | S0 stays in this SPEC rather than travelling with the M6 it blocks. | Placing S0 with M6 would make that successor 27 REQ / 26 AC — over the Tier L ceiling. The blocking relation survives as `SPEC-MODEL-MATRIX-DOCS-001`'s `depends_on` edge on this SPEC. |

---

## §F Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | S0 confirmation contradicts the §A.1 readings, invalidating the framing. | Medium | REQ-MPMC-004 — use the confirmed value, record the delta; the structural work does not depend on the numbers. |
| R-2 | S0 is being discharged *after* the documentation it blocks already shipped, so a contradiction is a correction rather than a prevention. | High | The remediation obligation is recorded in `SPEC-MODEL-MATRIX-DOCS-001`; a delta must reach the shipped pages, not only `progress.md`. |
| R-3 | `DefaultProfileMatrix()`'s Go type is unchanged while its inner-key semantics change from group key to agent name — a semantic change with no compiler signal. | Medium | Doc-comment statement plus the display-order/matrix-key agreement assertion. |
| R-4 | The unexecuted M1 remainder (`agentGroupMembership`, `AgentGroup`) leaves a live consumer at `internal/web/agentfm.go:491` against a SPEC that records M1 as landed. | Medium | Recorded here as this SPEC's own remainder; disposition owned by card t1037. |

---

## §G Traceability

| Unit | REQ range | AC range |
|---|---|---|
| S0 | REQ-MPMC-001 … 004 | AC-MPMC-001 … 003 |
| M1 | REQ-MPMC-005 … 018 | AC-MPMC-004 … 013 |

18 requirements, 13 acceptance criteria — both within the Tier L ceilings of 25 and 25. Full AC↔REQ mapping: `acceptance.md` §D.

---

## §H Cross-References

- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — the retired predecessor stub carrying the full split mapping table.
- `.moai/specs/SPEC-MODEL-MATRIX-CONFIG-001/` — successor; consumes the matrix.
- `.moai/specs/SPEC-MODEL-MATRIX-DOCS-001/` — successor; consumes both the matrix and the S0 record.
- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-001/` — the group abstraction superseded here.
