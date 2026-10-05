# Research — SPEC-MODEL-MATRIX-DOCS-001

Evidence gathered during the plan-phase authoring of `SPEC-MODEL-PROFILE-MATRIX-002` and inherited by this successor on the M6 + M7 seam. Every claim is attributed to the command or read that produced it. Unobserved items are named explicitly in §G.

All paths relative to the repository root. Line numbers are drift-prone; the accompanying content tokens are the durable anchors.

---

## §A Documentation contradictions (observed)

| Surface | Observed |
|---|---|
| `README.md` (+ ko/ja/zh) | Benchmark table headed `Model [max]` with `claude-opus-4.8` 59%/$13.22, `claude-fable-5` 70%/$21.63, `claude-sonnet-5` 54%/$26.40. Columns are `Pass@1 / Per-task cost / $/solved / Tokens/solved / Steps` — note `$/solved` and `Tokens/solved` are **derived** metrics, not raw leaderboard columns. |
| `.claude/rules/moai/development/model-policy.md` | Self-contradictory. The `## Model Policy Tiers (3-tier — max/medium/low)` section asserts "under the No-Haiku policy, **all workers are Sonnet 5 fixed across all tiers**" and points at `model_routing_profiles.{max,medium,low}` as "the 3-tier config SSOT". The following section `## Per-Agent Profile Resolver` describes the modern `llm.profile` resolver and `moai model profile`. Both claims cannot hold. The stale section also names the block `SPEC-MODEL-MATRIX-CONFIG-001` deletes. |
| `docs-site/content/zh/multi-llm/model-policy.md` | Retains a Haiku column (`\| 策略 \| 计划 \| Opus \| Sonnet \| Haiku \| 适合用途 \|`) with per-agent haiku assignments — `manager-docs` haiku/haiku, `manager-git` haiku/haiku/haiku, `builder-harness` … haiku — and a prose claim that the Low policy uses "只使用 Sonnet 和 Haiku". |

`README.md` is the sharpest illustration of why S0 blocks this SPEC: the numbers currently shipped are the `[max]`-effort rows, and the framing depends on `[low]` / `[high]` rows. Replacing one set with the other without pinning the effort label would produce a table that is internally consistent but mislabelled.

**Post-landing status**: these three contradictions were the targets of M6, which **landed** in squash `31da99a7b`. Whether each is actually resolved in the shipped pages was never verified at AC level — the landing carried toolchain verification only. Re-measure rather than assume.

---

## §B The S0 dependency, and why it inverted

The leaderboard verification record is owned by `SPEC-MODEL-MATRIX-CORE-001` and was **never discharged**. Its plan-phase probe returned rows that corroborate the user reading for GLM 5.2 on three of four metrics and contradict it entirely for Fable 5 and Opus 4.8 — most likely because the two sources read different per-effort rows. That probe is a lead, not evidence, and is explicitly marked insufficient for documentation use.

M6 nonetheless shipped, using the user-supplied per-effort measurements. The consequence for this SPEC:

- Every benchmark figure now live in the README set and in `advanced/no-haiku-3tier.md` traces to an unverified reading.
- S0's eventual discharge is a **correction behind** shipped pages, not a gate ahead of them.
- The delta must therefore reach the pages, not only the predecessor's `progress.md` (REQ-MPMD-011, `design.md` §A.6).

A second figure was corrected during M6 and is recorded for completeness: the original design report stated `xhigh` occupied 7 matrix cells before the change; counting the pre-change matrix at `HEAD^` gave **6**. The docs carry 6; the original report itself is uncorrected.

---

## §C Guard surfaces — what the rule scans today

Haiku-residual rule surfaces, from its own header comment in `internal/spec/lint_haiku_residual.go`:

1. Agent frontmatter/body in `.claude/agents/moai/*.md` + template mirror.
2. `claude_models` block in `llm.yaml` (glm.models exempt — exemption X2).
3. `model_routing_profiles` / `workflow_agents` / `role_profiles` in `workflow.yaml`.
4. `validRoutingModels` Go map in `internal/config/model_routing.go`.

Exemptions: X1 `_test.go` fixtures, X2 `glm.models.haiku`, X3 the model-policy alias closed set, X4 `internal/spec/` own source.

**Surfaces 3 and 4 both target artifacts `SPEC-MODEL-MATRIX-CONFIG-001` deletes.** Leaving them in place after that deletion means the rule scans for a block that cannot exist — harmless but misleading. REQ-MPMD-018 requires their removal in the same change that adds the new surfaces.

**Neither surface M7 adds is on the list today.** The web-console agentfm model option set and the v4manifest tier-suggestion table are both cleared by `SPEC-MODEL-MATRIX-SURFACES-001`; until M7 lands, nothing prevents `haiku` returning to either.

The two surfaces to be added, as observed:

- `internal/web/agentfm.go` `agentFMModelValues()` returns `{ModelInherit, ModelHaiku, ModelSonnet, ModelOpus, modelFable}`.
- `internal/harness/v4manifest/schema.go` `tierSuggestions` maps `TierLightBlue → {ModelHaiku, EffortLow}`.

---

## §D Byte-parity and mirror obligations

`internal/template/rule_template_mirror_test.go` enforces byte parity between `.claude/rules/**` and its template twin. `.claude/rules/moai/development/model-policy.md` is therefore edited on both sides in one commit (C-2); an edit to one side alone fails `TestRuleTemplateMirror`.

This is the mechanical guard behind REQ-MPMD-008, and it is the only documentation requirement in this SPEC with a mechanical rather than manual verification.

---

## §E The matrix property test

REQ-MPMD-019 consolidates the seven structural invariants into one property test. The invariants are owned by `SPEC-MODEL-MATRIX-CORE-001`'s design; this SPEC owns only the test that asserts them together:

cell count (33), model closed set, effort closed set, no haiku, no in-matrix inherit, profile-invariant trio, display-order agreement.

`TestDefaultProfileMatrix_NoHaiku` already exists and covers one of the seven. Whether the property test subsumes or sits beside it is an implementation choice for M7.

---

## §F Open questions for run-phase

- **[NEEDS CLARIFICATION: infographic disposition]** — `assets/images/readme/tokenomics-harness-{en,ko,ja,zh}.png` may embed superseded figures. Regeneration is out of scope, but M6 must still decide whether to leave a visibly stale image beside a corrected table, or remove the image reference pending regeneration. **Post-landing note**: M6 shipped without this decision being recorded, so the images are in whatever state they were — which is itself unverified (§G).

The three other clarifications carried by the original SPEC travelled with their milestones: the `Group` JSON field to `SPEC-MODEL-MATRIX-CORE-001`, the `profiles:` migration representability to `SPEC-MODEL-MATRIX-CONFIG-001`, and the wizard option vocabulary to `SPEC-MODEL-MATRIX-SURFACES-001`.

---

## §G Explicitly NOT verified

Inherited unverified inputs whose owning unit is M6 or M7:

- **The ja and ko copies of `multi-llm/model-policy.md`.** Only the zh copy was inspected for the haiku column; the claim that ja has diverged into a third shape is carried forward as an unverified input and must be re-measured.
- **The `advanced/profile-matrix.md` and `advanced/no-haiku-3tier.md` line ranges** cited in the original brief — not opened during authoring.
- **Whether `assets/images/readme/tokenomics-harness-*.png` actually embeds benchmark numbers** — not inspected.
- **Cross-platform build status** — not run.

All four remain unverified after M6's landing: the milestone shipped without re-measuring any of them, so they are inherited into this successor exactly as they were, rather than discharged.

Added by this split:

- **Whether M6's landed pages actually resolve the three contradictions in §A.** The landing record names the thirteen pages and four READMEs it touched, but no AC-level verification was performed on any of them. A page being edited is not evidence that the edit achieved what the requirement asked.

The remaining unverified inputs from the original SPEC belong to other successors: the three S0 leaderboard items to `SPEC-MODEL-MATRIX-CORE-001`, the `mp.*` i18n orphan status to `SPEC-MODEL-MATRIX-SURFACES-001`.
