# Research — SPEC-MODEL-PROFILE-MATRIX-002

Evidence gathered during plan-phase authoring. Every claim below is attributed to the command or read that produced it. Unobserved items are named explicitly in §G.

---

## §A S0 — leaderboard verification status (BLOCKING, NOT DISCHARGED)

### §A.1 What was attempted

A single `WebFetch` probe was issued against `https://deepswe.datacurve.ai` during plan authoring, asking for the four metrics of Fable 5, Opus 4.8, and GLM 5.2 plus the leaderboard version/date.

### §A.2 What the probe returned

The probe reported leaderboard **v1.1, updated 2026-07-21**, "113 tasks", "Models (17/17)", and these rows:

| Model | Pass@1 | Avg cost | Output tokens | Agent steps |
|---|---|---|---|---|
| Claude Fable 5 | 70% ±4% | $21.63 | 119k | 88 |
| Claude Opus 4.8 | 59% ±2% | $13.22 | 135k | 120 |
| GLM 5.2 | 44% ±2% | $3.92 | 78k | 129 |

### §A.3 Why this does NOT discharge S0

The probe rows **do not match** the user-supplied readings in `spec.md` §A.2, and the mismatch is structured, not random:

- **GLM 5.2 corroborates exactly** on three of four metrics: `$3.92`, `78k`, `129` are identical to the user's reading. Its Pass@1 came back as `44% ±2%`.
- **Fable 5 and Opus 4.8 do not corroborate at all.** The probe's Fable row (`$21.63` / 70% / 88 steps) and Opus row (`$13.22` / 59% / 120 steps) are close to the figures already sitting in `README.md:63-67`, which are labelled `[max]` effort.

The most probable explanation is that the leaderboard exposes **per-effort rows per model**, and the two sources read different rows:

- The user read Fable 5 at `[low]` and Opus 4.8 at `[high]`.
- The probe surfaced the `[max]` rows for those two models (the same rows the current README already cites), while GLM 5.2's `[max]` row happens to be the row the user also read — consistent with the user's own annotation `GLM 5.2 [max]`.

Under that explanation both readings can be simultaneously correct, and neither is verified for the rows M6 actually needs.

**A secondary observation**: the probe's Fable "output tokens" is `119k` while README's Fable cell reads `170k` under the header `Tokens/solved`. These are different metrics (raw output tokens vs tokens normalised per solved task), so the discrepancy is expected and is a reminder that **column semantics must be pinned in S0, not just values**.

### §A.4 Hypothesis on the 42%-vs-45% conflict

The observed GLM 5.2 Pass@1 is `44% ±2%`, whose interval spans 42-46. Both user readings (42% and 45%) fall inside that interval. The likely cause is that the two charts the user read render the same `44% ±2%` figure differently — one plotting the lower bound or a rounded bar, the other the upper region. **This is a hypothesis, not a finding.** S0 must confirm the canonical single value and state whether the leaderboard publishes a point estimate, an interval, or both.

### §A.5 S0 exit criteria (what run-phase must produce)

A verification record, stored in `progress.md`, containing per model:

1. The **effort level** of the row read (`low` / `medium` / `high` / `max`).
2. Per-task cost, Pass@1 (point value and error bar if published), output tokens, agent steps.
3. The **column header semantics** for the token metric (raw output tokens vs tokens-per-solved).
4. The leaderboard version string and update date.
5. An explicit delta table against `spec.md` §A.2 for any metric that differs.

Until that record exists, no benchmark figure may be written to a documentation surface (REQ-MPM2-001).

**Evidence-integrity note**: the §A.2 table above is a *summarised* WebFetch response, not a verbatim page capture. It is recorded as a lead for S0, and is not itself sufficient evidence for any documentation claim.

---

## §B Current-state code facts (observed)

All paths relative to the repository root. Line numbers are drift-prone; the accompanying content tokens are the durable anchors.

### §B.1 Matrix and resolver — `internal/template/profile_matrix.go`

| Symbol | Anchor | Observed |
|---|---|---|
| `defaultProfileMatrix` | `var defaultProfileMatrix = map[string]map[string]config.ModelEffort{` | 3 profiles × 6 groups = 18 cells. Carries an `@MX:ANCHOR` naming it the Matrix A SSOT. |
| Group constants | `GroupSpecAuditors = "spec_auditors"` … `GroupGit = "git"` | 6 constants in one `const` block. |
| `agentGroupMembership` | `var agentGroupMembership = map[string]string{` | 10 entries; `Explore` deliberately absent. |
| `profileMatrixAgentOrder` | `var profileMatrixAgentOrder = []string{` | 11 entries **including** `Explore` — display order already carries Explore even though the membership map does not. |
| `ProfileMatrixAgents()` | `func ProfileMatrixAgents() []string` | Defensive copy of the display order. |
| `DefaultProfileMatrix()` | `func DefaultProfileMatrix() map[string]map[string]config.ModelEffort` | Deep copy. |
| `AgentGroup(agent)` | `func AgentGroup(agent string) (string, bool)` | Returns group + membership flag. |
| `ResolveAgentModelEffort` | `func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, hasGroup bool)` | 4-step precedence: override → config `Profiles` cell → Go default cell → unknown-profile fallback to the `medium` column. Unmapped agent short-circuits to `{modelInherit, ""}, false`. |
| `ApplyProfile` | `func ApplyProfile(projectRoot, profile string) error` | Regex-replaces the `profile:` line in `llm.yaml`; inserts one under `llm:` when absent; also strips retired `plan_type:` and the `claude_models:` block via `stripRetiredLLMKeys`. Returns nil when the file is absent. |

Current 18-cell values (the `medium` column is the one the template agent frontmatter mirrors):

```
max:    spec_auditors fable/medium · develop fable/low · advisor fable/medium
        design_harness_e2e opus/high · docs sonnet/medium · git sonnet/low
medium: spec_auditors opus/high · develop opus/high · advisor fable/low
        design_harness_e2e opus/medium · docs sonnet/medium · git sonnet/low
low:    spec_auditors opus/low · develop opus/medium · advisor opus/high
        design_harness_e2e opus/low · docs sonnet/medium · git sonnet/low
```

**Delta note for M1**: the incoming 33-cell matrix changes many of these values, not only their shape. For example `max/spec_auditors` moves from `fable/medium` to `fable/low`, and `max/develop` moves from `fable/low` to `opus/xhigh` — an inversion of the develop cell's model class. M1 is therefore a value change as well as a structural change, and the fidelity test must be rewritten wholesale rather than reshaped.

### §B.2 Tests pinning the current contract — `internal/template/profile_matrix_test.go`

| Test | What it pins | M1 action |
|---|---|---|
| `TestResolveAgentModelEffort_MatrixAFidelity` | All 10 mapped agents against the `max` column, agent-keyed already | rewrite values; add `Explore` |
| `TestResolveAgentModelEffort_LowColumn` | 3 spot-checks on `low` | rewrite values |
| `TestResolveAgentModelEffort_OverridePrecedence` | Override wins; **a sibling in the same group is unaffected** | the "sibling in same group" premise dies with groups — retarget to "another agent is unaffected" |
| `TestResolveAgentModelEffort_Inherit` | `Explore` **and** `some-user-agent` both → `hasGroup=false`, `inherit` | **split** (REQ-MPM2-023) |
| `TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` | A config `profiles` cell beats the Go default | **delete** — `profiles:` is removed in M2 |
| `TestResolveAgentModelEffort_LegacyAlias` | `performance_tier: max` with no `profile` resolves the max column | keep; rewrite the expected value |
| `TestDefaultProfileMatrix_NoHaiku` | Zero haiku | keep as-is; extend per REQ-MPM2-103 |
| `TestApplyProfile_InsertsProfileWhenAbsent` | Migration insert path | unaffected |

`TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` is the only test that will fail *by design of M2* rather than by value drift — it asserts precisely the behavior REQ-MPM2-033 removes.

### §B.3 Resolver consumers

```
grep -rn "ApplyProfile(" --include="*.go" . | grep -v _test.go
  internal/web/agentfm.go:92
  internal/template/profile_matrix.go:80   (definition)
  internal/cli/update.go:469
  internal/cli/update.go:1495
  internal/cli/init.go:608
```

Four call sites plus the definition — these are the M3 seams (REQ-MPM2-046).

The only production consumer of the resolver trio (`ProfileMatrixAgents` / `AgentGroup` / `ResolveAgentModelEffort`) is `internal/cli/model.go` `resolveModelProfileReport`, which iterates the display order and emits a per-agent report carrying `Agent`, `Group`, `Model`, `Effort`, and — under a GLM backend — `GLMModel` and `GLMReasoning`. The report is rendered as a human table or as JSON (`--json`).

**M1 impact on this consumer**: the `Group` field of `modelProfileEntry` loses its source. Either the field is dropped from the report (a JSON shape change) or it is retained as a constant `"-"`. This is a design decision recorded in `design.md` §C.

**M3 relevance**: `moai model profile --json` is the natural machine-readable route for REQ-MPM2-049 (Workflow-path `opts.effort` lookup), because it already emits per-agent `{model, effort}` under the active profile and already accounts for the GLM overlay.

### §B.4 The 36-cell axis

```
grep -rn "model_routing_profiles|RouteModelFor" --include="*.go" --include="*.yaml"
```

Production surfaces:

- `.moai/config/sections/workflow.yaml` — the `model_routing_profiles:` block plus 3 explanatory comment lines.
- `internal/config/model_routing.go` — `RouteModelFor(specTier, phase, perfTier)` with an `@MX:ANCHOR` calling it "the spawn-time cost-routing accessor", plus 4 validator branches emitting `model_routing_profiles.*` field errors.
- `internal/config/types.go` — `ModelRoutingProfiles` field and its type, with doc comments referencing `RouteModelFor` fallback behavior.

Test surfaces: `internal/config/model_routing_test.go` carries a full-matrix YAML fixture and several `RouteModelFor` cases.

**Confirmed**: every `RouteModelFor` call outside `internal/config/model_routing_test.go` is within `model_routing.go` itself. The `@MX:ANCHOR` claiming "spawn-time cost-routing accessor" describes an intended consumer that was never wired — this is the "zero non-test production call sites" fact from the SPEC brief, and it is what makes D-3 (retire rather than migrate) safe.

### §B.5 Matrix literal duplication (4 copies today)

1. `internal/template/profile_matrix.go` — `defaultProfileMatrix` (Go constant, authoritative).
2. `.moai/config/sections/llm.yaml` — `llm.profiles` (flow-style, 18 cells).
3. `internal/template/templates/.moai/config/sections/llm.yaml` — `llm.profiles` (aligned style, 18 cells, plus ~20 lines of explanatory comment describing the six groups).
4. `internal/template/profile_matrix_test.go` — the `want` map in `TestResolveAgentModelEffort_MatrixAFidelity`.

After M2 copies 2 and 3 are removed, leaving the Go constant and the test (REQ-MPM2-034).

**Additional finding not in the brief**: the two `llm.yaml` copies have already drifted in a second, unrelated place. The local copy's `glm.models` block carries `opus`, `sonnet`, and `haiku` alias keys; the template copy carries only `high`, `medium`, `low`, `fable`. This drift is **out of this SPEC's scope** but is recorded here because M2 edits both files and a naive whole-block sync would silently import the local-only alias keys — including `haiku` — into the template, which would trip the haiku-residual rule's `claude_models` surface neighbourhood. M2 must edit the `profiles:` block only.

### §B.6 Agent frontmatter — current state

Deployed tree `.claude/agents/moai/*.md` (10 files; `Explore` has no file):

| Agent | model | effort |
|---|---|---|
| manager-spec | inherit | xhigh |
| plan-auditor | inherit | xhigh |
| sync-auditor | inherit | xhigh |
| manager-develop | inherit | xhigh |
| super-advisor | inherit | xhigh |
| manager-design | inherit | xhigh |
| builder-harness | inherit | high |
| e2e-tester | inherit | high |
| manager-docs | sonnet | medium |
| manager-git | sonnet | low |

**Important mismatch found**: the brief states the template baseline should hold the Medium-profile values (REQ-MPM2-043). Under the incoming 33-cell matrix the Medium column is `manager-spec opus/high`, `plan-auditor opus/xhigh`, `manager-develop opus/high`, `super-advisor fable/low`, `manager-design opus/high`, `builder-harness opus/high`, `e2e-tester opus/high`. Comparing to the table above, **six of the ten current effort values already differ from the incoming Medium column** (`manager-spec` xhigh→high, `manager-develop` xhigh→high, `super-advisor` xhigh→low, `manager-design` xhigh→high, `builder-harness` high→high *(same)*, `e2e-tester` high→high *(same)*). So M3 is not a pure no-op even at the default profile in this repository — REQ-MPM2-112's "no-op at medium" claim holds only **after** the template baseline is re-set to the new Medium column. `design.md` §E records the required ordering.

`internal/settings/agentfm/agentfm.go` exposes `Patch(path, model, effort string, deleteEffort bool) error` — the existing frontmatter-only editor. Its sole production caller is `internal/web/agentfm.go`. It is the natural implementation vehicle for REQ-MPM2-040/041/048.

### §B.7 Guard tests — what does and does not exist

| Guard | File | Scope | Pins effort? |
|---|---|---|---|
| Agent frontmatter audit | `internal/template/agent_frontmatter_audit_test.go` | template tree agents | not for effort values (audits retired fields, tools/skills shape) |
| Agents frontmatter CSV/mutual-exclusion | `internal/template/agents_frontmatter_test.go` | template tree agents | **NO** — inspected in full (190 lines); it validates only `tools:` / `disallowedTools:` CSV format and mutual exclusion. It reads no `model:` or `effort:` key. |
| Rule template mirror | `internal/template/rule_template_mirror_test.go` | `.claude/rules/**` byte parity | n/a |
| Haiku residual | `internal/spec/lint_haiku_residual.go` | 4 surfaces (see below) | n/a |

**`agents_frontmatter_test.go` is CONFIRMED not to pin effort values** — this discharges the brief's "VERIFY during planning" item. No CI guard enforces local-vs-template agent frontmatter byte parity, so M3's deployed-only rewrite (REQ-MPM2-042) will not trip an existing test.

Haiku-residual rule surfaces, from its own header comment:

1. Agent frontmatter/body in `.claude/agents/moai/*.md` + template mirror.
2. `claude_models` block in `llm.yaml` (glm.models exempt — exemption X2).
3. `model_routing_profiles` / `workflow_agents` / `role_profiles` in `workflow.yaml`.
4. `validRoutingModels` Go map in `internal/config/model_routing.go`.

Exemptions: X1 `_test.go` fixtures, X2 `glm.models.haiku`, X3 the model-policy alias closed set, X4 `internal/spec/` own source.

**M2/M7 coupling discovered**: surfaces 3 and 4 both target artifacts that M2 deletes. Leaving them in place after M2 means the rule scans for a block that cannot exist — harmless but misleading. REQ-MPM2-102 requires their removal in the same change that adds the new surfaces.

### §B.8 Web console and v4manifest

- `internal/web/agentfm.go` `agentFMModelValues()` returns `{ModelInherit, ModelHaiku, ModelSonnet, ModelOpus, modelFable}` — **`haiku` is offered** (REQ-MPM2-070).
- `internal/harness/v4manifest/schema.go` `tierSuggestions` maps `TierLightBlue → {ModelHaiku, EffortLow}` (REQ-MPM2-071). Its neighbouring `agentTiers` table is annotated `@MX:ANCHOR` as the badge-colour SSOT with distribution `🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 = 20`, and is explicitly **not** in scope.
- `applyPerfTierEdits` in `internal/web/agentfm.go` carries a comment block stating that `SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-040` retired frontmatter re-application. **M3 makes that comment false** — it must be rewritten, not merely left in place, at the same time the seam is re-wired (this is the code-side twin of the stale `agentfm.tier.desc` i18n string, REQ-MPM2-072).

### §B.9 Wizard

`internal/cli/wizard/questions.go`, question id `model_policy`:

- Title `"Select model policy"`, description `"Controls which Claude model tier is assigned to each agent. Match to your Claude plan."`
- Options: `High (Recommended)` / `high` / "Opus for critical agents — Max $200 plan"; `Medium` / `medium` / "Opus for key agents, sonnet for rest — Max $100 plan"; `Low` / `low` / "Sonnet and haiku only — Plus $20 plan".
- Default `high`.

Two defects confirmed: the `Low` description references `haiku` (REQ-MPM2-061), and the option **values** are `high`/`medium`/`low` while the profile vocabulary is `max`/`medium`/`low`. The bridge is `template.NormalizeToTier(result.ModelPolicy)` at `internal/cli/update.go:1495`, which maps `high`→`max`. REQ-MPM2-062 constrains the persisted result, deliberately leaving the choice of "change the wizard values to max/medium/low" vs "keep high and rely on the normalizer" to `design.md` §D.

### §B.10 Documentation contradictions

| Surface | Observed |
|---|---|
| `README.md:63-67` (+ ko/ja/zh) | Benchmark table headed `Model [max]` with `claude-opus-4.8` 59%/$13.22, `claude-fable-5` 70%/$21.63, `claude-sonnet-5` 54%/$26.40. Columns are `Pass@1 / Per-task cost / $/solved / Tokens/solved / Steps` — note `$/solved` and `Tokens/solved` are **derived** metrics, not raw leaderboard columns. |
| `.claude/rules/moai/development/model-policy.md` | Self-contradictory. The `## Model Policy Tiers (3-tier — max/medium/low)` section asserts "under the No-Haiku policy, **all workers are Sonnet 5 fixed across all tiers**" and points at `model_routing_profiles.{max,medium,low}` as "the 3-tier config SSOT". The following section `## Per-Agent Profile Resolver` describes the modern `llm.profile` resolver and `moai model profile`. Both claims cannot hold. The stale section also names the block M2 deletes. |
| `docs-site/content/zh/multi-llm/model-policy.md` | Retains a Haiku column (`| 策略 | 计划 | Opus | Sonnet | Haiku | 适合用途 |`) with per-agent haiku assignments — `manager-docs` haiku/haiku, `manager-git` haiku/haiku/haiku, `builder-harness` … haiku — and a prose claim that the Low policy uses "只使用 Sonnet 和 Haiku". |

`README.md:63-67` is the sharpest illustration of why S0 blocks M6: the numbers currently shipped are the `[max]`-effort rows, and the SPEC's framing depends on `[low]` / `[high]` rows. Replacing one set with the other without pinning the effort label would produce a table that is internally consistent but mislabelled.

---

## §C Precedent — how the predecessor's retirement was justified

`SPEC-MODEL-PROFILE-MATRIX-001` REQ-MPM-040 (quoted from the comment block at `internal/web/agentfm.go`):

> the former tier-profile re-application (which rewrote each shipped agent's `model:`/`effort:` and re-introduced the `[1m]`-hazard concrete-model pin) is retired — the web save now persists to `llm.yaml` only, leaving agent frontmatter at `model: inherit`.

The stated cause of retirement is **the `model:` pin**, not the `effort:` write. The revival (D-2) keeps `model: inherit` untouched and writes only `effort:`, so it does not reintroduce the cited hazard. This is the load-bearing argument for the revival and is why REQ-MPM2-041 is phrased as a prohibition rather than a preference.

---

## §D Two-channel effort injection — schema evidence

| Fact | Source |
|---|---|
| `Agent` tool accepts `model` with enum `sonnet\|opus\|haiku\|fable`; has no `effort` parameter | live tool schema available in the authoring session |
| `Workflow` tool `agent()` accepts `opts.model` and `opts.effort ∈ {low, medium, high, xhigh, max}`, plus `opts.agentType` to target a named subagent | live tool schema available in the authoring session |
| Workflow `agent()` effort is validated against that closed set in MoAI's own doctrine | `.claude/rules/moai/workflow/dynamic-workflows.md:91` |

Note the asymmetry: the Workflow effort set includes `max`, while the 33-cell matrix uses at most `xhigh`. No matrix cell can therefore produce an out-of-range `opts.effort` value.

Note also the model-enum asymmetry: the `Agent` tool enum includes `haiku` (which MoAI policy forbids) and excludes `inherit` (which the resolver returns for unmapped agents). The unmapped-agent path must therefore **omit** the `model` argument entirely rather than pass `inherit` — constraint C-7.

---

## §E Prior-art within the repo for the M3 mitigations

| Mitigation | Existing precedent |
|---|---|
| Frontmatter-only edit | `internal/settings/agentfm/agentfm.go` `Patch` — writes frontmatter keys, leaves body bytes untouched |
| Live-files-only, no template dual-write | same `Patch`; its only caller `internal/web/agentfm.go:365` operates on deployed paths |
| Post-deploy reapplication ordering | `internal/cli/update.go:469` already calls `ApplyProfile` after the deploy step in the `--profile` flag path |
| Override-respecting exclusion | `ResolveAgentModelEffort` step 1 already short-circuits on `cfg.AgentOverrides[agent]` |

All four mitigations reuse an existing in-tree mechanism; none requires a new subsystem.

---

## §F Open questions for run-phase

- **[NEEDS CLARIFICATION: `Group` field in `moai model profile --json`]** — M1 removes the group concept. Does the JSON report drop the `group` key (a breaking shape change for any consumer) or retain it as a constant `"-"`? No external consumer was found, but the `--json` flag is a documented public surface.
- **[NEEDS CLARIFICATION: wizard option value vocabulary]** — Should the wizard's option values become `max`/`medium`/`low` directly (removing the `high`→`max` normalizer hop), or stay `high`/`medium`/`low` for backward compatibility with any recorded answer? REQ-MPM2-062 constrains only the persisted outcome.
- **[NEEDS CLARIFICATION: `profiles:` migration representability]** — REQ-MPM2-036 splits on whether a user's `profiles:` customization is "representable as per-agent overrides". A group-keyed cell expands to 1-3 agent overrides mechanically, so representability is arguably always true. Should the warning branch exist at all, or should migration be unconditional with a summary notice?
- **[NEEDS CLARIFICATION: infographic disposition]** — `assets/images/readme/tokenomics-harness-{en,ko,ja,zh}.png` may embed superseded figures. §C of `spec.md` defers regeneration, but M6 must still decide whether to leave a visibly stale image beside a corrected table, or remove the image reference pending regeneration.

---

## §G Explicitly NOT verified during plan authoring

- The DeepSWE `[low]` Fable 5 row and the `[high]` Opus 4.8 row (§A.3) — the probe surfaced `[max]` rows for those models.
- The canonical GLM 5.2 Pass@1 point value (§A.4) — observed as `44% ±2%`, which matches neither user reading exactly.
- The token-column semantics of the leaderboard (§A.3) — raw output tokens vs tokens-per-solved is unresolved.
- The ja and ko copies of `multi-llm/model-policy.md` — only the zh copy was inspected for the haiku column; the brief's claim that "ja has diverged into a third shape" is carried forward as an unverified input and must be re-measured in M6.
- The `advanced/profile-matrix.md` and `advanced/no-haiku-3tier.md` line ranges cited in the brief — not opened during authoring.
- The `mp.*` i18n key family's zero-consumer status — not re-grepped; carried forward from the brief as an unverified input.
- Whether `assets/images/readme/tokenomics-harness-*.png` actually embeds benchmark numbers — not inspected.
- Cross-platform build status — not run.
