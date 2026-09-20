# Implementation Plan — SPEC-MODEL-MATRIX-SURFACES-001

Tier M. Two milestones: M4 (init wizard question) and M5 (web console cleanup). Neither has started, and neither depends on any other successor.

This plan carries the design rationale and the observed current-state evidence inline, in §B and §C. At Tier M the artifact set is three files plus `progress.md`, so the substance that would otherwise live in `design.md` and `research.md` is here rather than dropped.

---

## §A Context

This SPEC carries the third seam of the `SPEC-MODEL-PROFILE-MATRIX-002` split: the two user-facing surfaces that still speak the pre-matrix vocabulary. Both were declared independent of every other milestone in the predecessor's plan, which is why this successor is a root with no `depends_on`.

Scope surfaces: `internal/cli/wizard/questions.go` and its translation blocks, `internal/web/agentfm.go` (model option set only), `internal/harness/v4manifest/schema.go`, the four web-console locale files, and their tests.

---

## §B Review first — the decisions most likely to change

### §B.1 The wizard value-vocabulary seam (M4)

The wizard's `model_policy` values are `high`/`medium`/`low`; the profile vocabulary is `max`/`medium`/`low`; `NormalizeToTier` bridges them (`high`→`max`).

| | Change values to max/medium/low | Keep high/medium/low |
|---|---|---|
| Normalizer hop | removed | retained |
| User-visible vocabulary | matches `llm.profile`, `moai model profile`, docs | mismatched (a user picking "High" gets `profile: max`) |
| Risk | any stored or scripted answer of `high` must still normalize | none |

**Design preference: change the values**, keeping `NormalizeToTier` as a tolerant reader for legacy `high`. The mismatch is a real user-facing confusion — the whole point of M4 is to stop the question from lying about what it does — and the normalizer already handles the backward-compatible direction.

This is carried as an open clarification (§D item 2): REQ-MPMS-003 constrains only the persisted outcome, deliberately leaving the choice to run-phase.

### §B.2 Option copy — what the descriptions must and must not say (M4)

**Must say**: which subscription tier the profile targets.
**Must not say**: that a higher profile is stronger or produces better results.

The honest framing is access-based: the Max profile uses models available on the higher subscription tiers; the Low profile restricts to models available on the entry tier. Cost and quality do **not** move monotonically with the profile name, and the option text must not imply they do.

This is copy, and should be reviewed as copy rather than as a checklist item. A description that is technically free of the forbidden claim while still *implying* the ordering fails REQ-MPMS-006.

### §B.3 Whether the `mp.*` family is actually orphaned (M5)

Its zero-consumer status is a carried-forward input from the original brief, never measured. Removing a live i18n key breaks a rendered surface silently — there is no compile error for a missing translation lookup. Re-grep for consumers before removal; do not inherit the claim.

### §B.4 Localization scope (M4 + M5)

The `ko` translation block omits `model_policy` entirely today, so the question renders English in every locale. The new entries must cover title, description, and each option's label and description, in `ko`, `ja`, and `zh`, in one change. Partial locale coverage is how parity breaks, and the failure is invisible to anyone reading only their own locale.

---

## §C Observed current state

All paths relative to the repository root. Line numbers are drift-prone; the accompanying content tokens are the durable anchors.

### §C.1 The wizard question

`internal/cli/wizard/questions.go`, question id `model_policy`, as observed during the original plan authoring:

- Title `"Select model policy"`, description `"Controls which Claude model tier is assigned to each agent. Match to your Claude plan."`
- Options: `High (Recommended)` / value `high` / "Opus for critical agents — Max $200 plan"; `Medium` / value `medium` / "Opus for key agents, sonnet for rest — Max $100 plan"; `Low` / value `low` / "Sonnet and haiku only — Plus $20 plan".
- Default `high`.

Two defects confirmed: the `Low` description references `haiku` (REQ-MPMS-002), and the option values are `high`/`medium`/`low` while the profile vocabulary is `max`/`medium`/`low`. The bridge is `template.NormalizeToTier(result.ModelPolicy)` in `internal/cli/update.go`, which maps `high`→`max`.

The `translations.go` `ko` block omits `model_policy` entirely.

**Already landed, not an M4 deliverable**: squash `31da99a7b` corrected the on-screen labels in `internal/cli/profile_setup_translations.go` across four locales, which still read `"Max - Fable 5 (low) + Opus 4.8 (high)"` while the option value was already `high`. That was a defect the matrix change surfaced; M4 adds the rewritten question.

### §C.2 The web console and v4manifest

- `internal/web/agentfm.go` `agentFMModelValues()` returns `{ModelInherit, ModelHaiku, ModelSonnet, ModelOpus, modelFable}` — **`haiku` is offered** (REQ-MPMS-007).
- `internal/harness/v4manifest/schema.go` `tierSuggestions` maps `TierLightBlue → {ModelHaiku, EffortLow}` (REQ-MPMS-008). Its neighbouring `agentTiers` table is annotated `@MX:ANCHOR` as the badge-colour SSOT with distribution `🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 = 20`, and is explicitly **not** in scope.
- `applyPerfTierEdits` in the same `agentfm.go` carries a comment block stating that frontmatter re-application is retired. Rewriting it is `SPEC-MODEL-MATRIX-CONFIG-001`'s obligation, not this SPEC's — the two SPECs touch this file for unrelated reasons and must be sequenced rather than merged.

### §C.3 The haiku-residual guard, and what it does not yet cover

The rule's four surfaces are: agent frontmatter plus template mirror; the `claude_models` block in `llm.yaml` (with `glm.models` exempt); `model_routing_profiles` / `workflow_agents` / `role_profiles` in `workflow.yaml`; and the `validRoutingModels` Go map.

Neither of the two surfaces M5 clears is on that list. Adding them is `SPEC-MODEL-MATRIX-DOCS-001`'s M7 obligation and depends on this SPEC clearing them first — so until that lands, nothing prevents `haiku` returning to either surface.

---

## §D Pre-flight

1. Resolve the wizard option-value clarification (§B.1) — it decides whether M4 changes values or only copy.
2. Re-grep the `mp.*` family for live consumers (§B.3) before treating it as orphaned.
3. `git fetch origin` and check divergence — a parallel session may be active on this shared checkout.
4. Confirm no concurrent edit to `internal/web/agentfm.go` from `SPEC-MODEL-MATRIX-CONFIG-001` (§C.2).
5. `go build ./... && go test ./internal/cli/... ./internal/web/... ./internal/harness/...` to establish a green baseline.

---

## §E Constraints carried into execution

4-locale parity as a HARD obligation, No-Haiku as a HARD non-skippable gate, backward-compatible normalization of a legacy `high` answer, the v4manifest badge table preserved untouched, and explicit pathspecs on every commit.

---

## §F Milestones

### M4 — Init wizard question

**Priority: Medium. Independent of every other successor; user-facing.**

1. Rewrite the `model_policy` question to subscription-tier framing; remove the `haiku` reference.
2. Change the option values to `max`/`medium`/`low` per §B.1, keeping `NormalizeToTier` as a tolerant reader for a legacy `high`.
3. Write option descriptions that state the target subscription tier and do **not** assert a performance ordering (§B.2).
4. Add `ko`, `ja`, and `zh` translations for the title, description, and every option label and description.
5. Test: the question renders localized text under each of the three non-English locales, and the persisted `llm.profile` is one of `max`/`medium`/`low` for every option.

Covers REQ-MPMS-001 … 006.

### M5 — Web console cleanup

**Priority: Medium. Independent of every other successor.**

1. Remove `haiku` from the agentfm model selector option set.
2. Change the v4manifest lightblue tier suggestion to `sonnet / low`. Leave `agentTiers` untouched.
3. Re-word `agentfm.tier.desc` in all four locales to describe the effort-reapplication behaviour `SPEC-MODEL-MATRIX-CONFIG-001` restores — re-word, do not delete.
4. Remove the orphaned `mp.*` key family from all four locale files — re-grep for consumers first (§B.3).
5. Verify no i18n key exists in a subset of locales after the edits.
6. Record the handoff to `SPEC-MODEL-MATRIX-DOCS-001`: both cleared surfaces now need guard coverage (§C.3).

Covers REQ-MPMS-007 … 011.

---

## §G Anti-patterns to avoid

| # | Anti-pattern | Why it bites here |
|---|---|---|
| AP-1 | Deleting `agentfm.tier.desc` instead of re-wording it | `SPEC-MODEL-MATRIX-CONFIG-001` makes the described behaviour true again; deletion loses a now-accurate string |
| AP-2 | Editing one locale and deferring the other three | 4-locale parity is a HARD obligation; deferral is how parity breaks, and the break is invisible to a reader of one locale |
| AP-3 | Removing the `mp.*` family on the inherited zero-consumer claim | A missing translation lookup has no compile error; re-grep first (§B.3) |
| AP-4 | Writing option copy that avoids the forbidden claim while still implying the ordering | REQ-MPMS-006 is about what the reader concludes, not about which words are absent |
| AP-5 | Widening M5 into the v4manifest `agentTiers` badge table | A separate SSOT with its own anchor; touching it couples two unrelated decisions |
| AP-6 | Merging this SPEC's `agentfm.go` edits with CONFIG's | Two independent concerns in one file; sequence them |
| AP-7 | Clearing `haiku` and assuming the guard will notice | Neither cleared surface is on the guard's list yet; hand the addition off explicitly (§C.3) |
| AP-8 | `git add -A` on this shared checkout | A parallel session may have staged unrelated work; use explicit pathspecs |

---

## §H Cross-References

- `spec.md` — requirements, decisions, risks
- `acceptance.md` — AC matrix and verification commands
- `.moai/specs/SPEC-MODEL-MATRIX-CONFIG-001/` — restores the behaviour `agentfm.tier.desc` must describe; shares `internal/web/agentfm.go`
- `.moai/specs/SPEC-MODEL-MATRIX-DOCS-001/` — adds the guard surfaces this SPEC clears
- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md` — retired predecessor stub
