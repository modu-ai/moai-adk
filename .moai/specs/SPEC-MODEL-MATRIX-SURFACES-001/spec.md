---
id: SPEC-MODEL-MATRIX-SURFACES-001
title: "User-facing surfaces — init wizard question + web console cleanup"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/cli/wizard + internal/web + internal/harness/v4manifest"
lifecycle: spec-anchored
tags: "model-profile, init-wizard, web-console, i18n, split-successor"
tier: M
era: V3R6
related_specs: [SPEC-MODEL-PROFILE-MATRIX-002, SPEC-WEBCONF-SIMPLIFY-001, SPEC-MODEL-MATRIX-CORE-001]
---

# SPEC-MODEL-MATRIX-SURFACES-001 — User-facing surfaces

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from `SPEC-MODEL-PROFILE-MATRIX-002` (card t1036) on the M4 + M5 seam, after that SPEC exceeded the Tier L REQ/AC ceilings (72/64 against 25/25). Carries M4 (init wizard question) and M5 (web console cleanup). Requirements re-numbered `REQ-MPMS-*`, criteria `AC-MPMS-*`; nothing was dropped. Tier M assigned on REQ count (11 ≤ 16) and AC count (10 ≤ 16), with a file scope of roughly ten files — the wizard question and its translations, the web console's model option set, the v4manifest tier table, four locale files, and their tests — inside the Tier M 5-15 band. Artifact set is the Tier M three plus `progress.md`; the design and research substance that would have gone to separate artifacts is carried in `plan.md` §B and §C. Status `draft`: nothing in this SPEC has landed. | manager-spec |

## Position in the chain

One of two roots, alongside `SPEC-MODEL-MATRIX-CORE-001`.

```
SPEC-MODEL-MATRIX-CORE-001 ──┬─→ SPEC-MODEL-MATRIX-CONFIG-001 ──┐
                             │                                  ├─→ SPEC-MODEL-MATRIX-DOCS-001
SPEC-MODEL-MATRIX-SURFACES-001 ───────────────────────────────── ┘
```

**This SPEC has no `depends_on` and may proceed in parallel with everything else.** Both of its milestones were declared independent in the predecessor's plan: M4 is independent of M1-M3 and M5 is independent of M1-M4. Neither reads the matrix, neither waits on the leaderboard record, and neither requires the effort application to exist. `SPEC-MODEL-MATRIX-DOCS-001` depends on this SPEC only because its M7 guard realignment adds surfaces that M5 first clears of `haiku`.

**Nothing in this SPEC has landed.** M4 and M5 were never started under the original SPEC ID.

One nearby change did land and is **not** an M4 deliverable: squash `31da99a7b` corrected the `moai init` wizard's on-screen model-policy **labels** (`internal/cli/profile_setup_translations.go`, four locales), which still read `"Max - Fable 5 (low) + Opus 4.8 (high)"` while the option value was already `high`. That was a defect surfaced by the matrix change, not the question rewrite M4 owns.

---

## §A Context

### §A.1 The wizard question lies about what it does

The `moai init` `model_policy` question reads as a model-class choice — its `Low` option description names `haiku`, a model MoAI policy forbids — and its option **values** are `high`/`medium`/`low` while the profile vocabulary is `max`/`medium`/`low`. A user selecting "High" gets `profile: max`. `template.NormalizeToTier` bridges the two, so the mismatch is invisible in behaviour and visible only in the user's mental model.

The profile axis is a **subscription-tier access axis**, not a performance-grade axis: under the leaderboard readings the Max profile is simultaneously cheaper and higher-scoring than the Medium profile. The question must therefore frame the choice by subscription tier, and must not assert a performance ordering the data contradicts. The full disclosure of that inversion is documentation work owned by `SPEC-MODEL-MATRIX-DOCS-001`; this SPEC's obligation is narrower — the question must not *contradict* it.

The `ko` translation block omits `model_policy` entirely, so the question renders English in every locale today.

### §A.2 The web console still offers a forbidden model

`internal/web/agentfm.go` `agentFMModelValues()` returns `{ModelInherit, ModelHaiku, ModelSonnet, ModelOpus, modelFable}` — `haiku` is offered in the agent-frontmatter model selector. `internal/harness/v4manifest/schema.go` `tierSuggestions` maps the lightblue tier to `{ModelHaiku, EffortLow}`. Both are No-Haiku policy violations on user-facing surfaces.

Two i18n items sit beside them. `agentfm.tier.desc` describes frontmatter effort re-application as retired — a description that becomes accurate again once `SPEC-MODEL-MATRIX-CONFIG-001` restores the behaviour, so it is re-worded rather than deleted. The `mp.*` key family is orphaned; its zero-consumer status is carried forward as an unverified input and must be re-grepped rather than inherited.

---

## §B Requirements (GEARS)

### §B.1 M4 — init wizard question

- **REQ-MPMS-001** (Ubiquitous) The `moai init` model-policy question shall present subscription-tier framing rather than model-class framing.
- **REQ-MPMS-002** (Unwanted) The question's option text shall not reference `haiku`.
- **REQ-MPMS-003** (Capability gate) **Where** the wizard's option values feed the profile normalizer, the resulting persisted `llm.profile` shall be one of `max`, `medium`, `low`.
- **REQ-MPMS-004** (Ubiquitous) The question title, description, and every option label and description shall have `ko`, `ja`, and `zh` translations.
- **REQ-MPMS-005** (Event-driven) **When** the wizard renders under a non-English conversation language, the model-policy question shall not fall back to English text.
- **REQ-MPMS-006** (Ubiquitous) The option descriptions shall state the subscription tier each profile targets and shall not assert a performance ordering contradicted by the leaderboard readings.

### §B.2 M5 — web console cleanup

- **REQ-MPMS-007** (Unwanted) The web console agent-frontmatter model selector shall not offer `haiku`.
- **REQ-MPMS-008** (Ubiquitous) The v4manifest lightblue-tier suggestion shall be `sonnet / low`.
- **REQ-MPMS-009** (Ubiquitous) The `agentfm.tier.desc` i18n string shall be re-worded in all four locales to describe the effort-reapplication behavior that `SPEC-MODEL-MATRIX-CONFIG-001` restores, rather than deleted.
- **REQ-MPMS-010** (Ubiquitous) The orphaned `mp.*` i18n key family shall be removed from all four locale files.
- **REQ-MPMS-011** (Event-driven) **When** a locale file is edited in M5, all four locale files shall be edited in the same change so no key exists in a subset of locales.

---

## §C Exclusions

### Out of Scope — the matrix and the config axis

- The 33-cell matrix, the resolver, the group removal, and the leaderboard verification record. All belong to `SPEC-MODEL-MATRIX-CORE-001`.
- The `model_routing_profiles` retirement, the `llm.yaml profiles:` mirror removal, and both effort channels. All belong to `SPEC-MODEL-MATRIX-CONFIG-001`.

Note the `internal/web/agentfm.go` boundary: this SPEC edits the **model option set** (REQ-MPMS-007); `SPEC-MODEL-MATRIX-CONFIG-001` edits the **profile-save seam** and rewrites the stale `applyPerfTierEdits` comment. Two distinct concerns in one file — sequence them rather than merging them.

### Out of Scope — documentation and guards

- Every docs-site page, README locale copy, and rules-file edit, including the naming-inversion disclosure and the Max-Opus rationale. They belong to `SPEC-MODEL-MATRIX-DOCS-001`.
- Adding the web-console model option set and the v4manifest tier table to the haiku-residual lint rule's surface list. That is `SPEC-MODEL-MATRIX-DOCS-001`'s M7 obligation, and it depends on this SPEC having cleared both surfaces first.

### Out of Scope — adjacent surfaces this SPEC deliberately does not touch

- Wizard questions other than `model_policy`.
- The `moai web` console's non-agentfm panels.
- The v4manifest `agentTiers` name→tier badge assignment table. Only the lightblue tier's suggested `{model, effort}` pair changes (REQ-MPMS-008); the badge distribution is explicitly preserved.
- The `haiku` alias in the GLM model map (`glm.models.haiku`), which is an exempt surface of the haiku-residual rule by design.
- The already-landed `profile_setup_translations.go` label correction, which is not an M4 deliverable.

### Out of Scope — profile renaming

- Renaming `max` / `medium` / `low`. REQ-MPMS-003 constrains the persisted value to that vocabulary; it does not authorize changing it.

---

## §D Constraints

| # | Constraint | Source |
|---|---|---|
| C-1 | Template tree content must remain internally neutral — no SPEC IDs, REQ tokens, internal dates, or commit SHAs in `internal/template/templates/**`. | CLAUDE.local.md §25 |
| C-2 | 4-locale parity is mandatory: no key may exist in a subset of locales. | CLAUDE.local.md §17 + REQ-MPMS-011 |
| C-3 | The No-Haiku policy is a HARD gate not skippable via SPEC `lint.skip`. | `internal/spec/lint_haiku_residual.go` |
| C-4 | Any stored or scripted wizard answer of `high` must continue to normalize, whatever the new option values are. | backward compatibility |
| C-5 | A parallel session may be active on this shared checkout; changes are committed with explicit pathspecs. | project operating practice |

---

## §E Decisions

| # | Decision | Rationale |
|---|---|---|
| D-1 | Profile names retained; the question discloses tier access rather than asserting a quality ordering. | Renaming has a far larger blast radius than the problem it solves; the question's job is to stop lying, not to rename the axis. |
| D-2 | `agentfm.tier.desc` is re-worded, not deleted. | `SPEC-MODEL-MATRIX-CONFIG-001` makes the described behaviour true again; deletion would lose a now-accurate string. |
| D-3 | The v4manifest `agentTiers` badge table is explicitly preserved; only the lightblue suggestion changes. | The badge distribution is a separate SSOT with its own `@MX:ANCHOR`; widening scope to it would couple two unrelated decisions. |

---

## §F Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | Changing the wizard option values breaks a stored or scripted answer of `high`. | Medium | Keep `NormalizeToTier` as a tolerant reader for legacy `high` (C-4, AC-MPMS-003). |
| R-2 | Editing one locale and deferring the other three breaks 4-locale parity. | Medium | REQ-MPMS-011 plus the key-set equality assertion (AC-MPMS-010). |
| R-3 | The `mp.*` family's zero-consumer status is inherited rather than measured; removing a live key would break a rendered surface silently. | Medium | Re-grep for consumers before removal rather than trusting the carried-forward input. |
| R-4 | `internal/web/agentfm.go` is edited by this SPEC and by `SPEC-MODEL-MATRIX-CONFIG-001` for unrelated reasons. | Low | Independent concerns in one file; sequence the two SPECs' edits. |
| R-5 | Clearing `haiku` from the two surfaces without the guard addition leaves nothing preventing its return. | Low | The guard addition is `SPEC-MODEL-MATRIX-DOCS-001`'s M7 obligation and depends on this SPEC; hand it off explicitly. |

---

## §G Traceability

| Unit | REQ range | AC range |
|---|---|---|
| M4 | REQ-MPMS-001 … 006 | AC-MPMS-001 … 005 |
| M5 | REQ-MPMS-007 … 011 | AC-MPMS-006 … 010 |

11 requirements, 10 acceptance criteria — both within the Tier M ceilings of 16 and 16. Full AC↔REQ mapping: `acceptance.md` §D.

---

## §H Cross-References

- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — the retired predecessor stub carrying the full split mapping table.
- `.moai/specs/SPEC-MODEL-MATRIX-CONFIG-001/` — restores the behaviour `agentfm.tier.desc` must describe; shares `internal/web/agentfm.go`.
- `.moai/specs/SPEC-MODEL-MATRIX-DOCS-001/` — adds the guard surfaces this SPEC clears.
- `.moai/specs/SPEC-WEBCONF-SIMPLIFY-001/` — origin of the v4manifest tier badge table whose lightblue suggestion changes here.
