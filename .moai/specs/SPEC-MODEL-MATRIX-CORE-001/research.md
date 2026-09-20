# Research — SPEC-MODEL-MATRIX-CORE-001

Evidence gathered during the plan-phase authoring of `SPEC-MODEL-PROFILE-MATRIX-002` and inherited by this successor on the S0 + M1 seam. Every claim is attributed to the command or read that produced it. Unobserved items are named explicitly in §G.

---

## §A S0 — leaderboard verification status (BLOCKING, NOT DISCHARGED)

### §A.1 What was attempted

A single `WebFetch` probe was issued against the DeepSWE leaderboard during plan authoring, asking for the four metrics of Fable 5, Opus 4.8, and GLM 5.2 plus the leaderboard version and date.

### §A.2 What the probe returned

The probe reported leaderboard **v1.1, updated 2026-07-21**, "113 tasks", "Models (17/17)", and these rows:

| Model | Pass@1 | Avg cost | Output tokens | Agent steps |
|---|---|---|---|---|
| Claude Fable 5 | 70% ±4% | $21.63 | 119k | 88 |
| Claude Opus 4.8 | 59% ±2% | $13.22 | 135k | 120 |
| GLM 5.2 | 44% ±2% | $3.92 | 78k | 129 |

### §A.3 Why this does NOT discharge S0

The probe rows **do not match** the user-supplied readings in `spec.md` §A.1, and the mismatch is structured, not random:

- **GLM 5.2 corroborates exactly** on three of four metrics: `$3.92`, `78k`, `129` are identical to the user's reading. Its Pass@1 came back as `44% ±2%`.
- **Fable 5 and Opus 4.8 do not corroborate at all.** The probe's Fable row (`$21.63` / 70% / 88 steps) and Opus row (`$13.22` / 59% / 120 steps) are close to the figures already sitting in `README.md`, which are labelled `[max]` effort.

The most probable explanation is that the leaderboard exposes **per-effort rows per model**, and the two sources read different rows: the user read Fable 5 at `[low]` and Opus 4.8 at `[high]`, while the probe surfaced the `[max]` rows for those two models — the same rows the current README already cites — and GLM 5.2's `[max]` row happens to be the row the user also read, consistent with the user's own `GLM 5.2 [max]` annotation.

Under that explanation both readings can be simultaneously correct, and neither is verified for the rows the documentation actually needs.

**A secondary observation**: the probe's Fable "output tokens" is `119k` while README's Fable cell reads `170k` under the header `Tokens/solved`. These are different metrics (raw output tokens vs tokens normalised per solved task), so the discrepancy is expected and is a reminder that **column semantics must be pinned in S0, not just values**.

### §A.4 Hypothesis on the 42%-vs-45% conflict

The observed GLM 5.2 Pass@1 is `44% ±2%`, whose interval spans 42-46. Both user readings (42% and 45%) fall inside that interval. The likely cause is that the two charts the user read render the same `44% ±2%` figure differently. **This is a hypothesis, not a finding.** S0 must confirm the canonical single value and state whether the leaderboard publishes a point estimate, an interval, or both.

### §A.5 S0 exit criteria (what run-phase must produce)

A verification record, stored in `progress.md`, containing per model:

1. The **effort level** of the row read (`low` / `medium` / `high` / `max`).
2. Per-task cost, Pass@1 (point value and error bar if published), output tokens, agent steps.
3. The **column header semantics** for the token metric (raw output tokens vs tokens-per-solved).
4. The leaderboard version string and update date.
5. An explicit delta table against `spec.md` §A.1 for any metric that differs.

**Evidence-integrity note**: the §A.2 table above is a *summarised* WebFetch response, not a verbatim page capture. It is recorded as a lead for S0, and is not itself sufficient evidence for any documentation claim.

**Successor-specific addendum**: the documentation S0 blocks **has already shipped** under the original SPEC ID. S0's exit therefore carries a sixth obligation the original did not have — routing every non-empty delta row to `SPEC-MODEL-MATRIX-DOCS-001` as a correction against live pages.

---

## §B Current-state code facts (observed)

All paths relative to the repository root. Line numbers are drift-prone; the accompanying content tokens are the durable anchors.

### §B.1 Matrix and resolver — `internal/template/profile_matrix.go`

| Symbol | Anchor | Observed |
|---|---|---|
| `defaultProfileMatrix` | `var defaultProfileMatrix = map[string]map[string]config.ModelEffort{` | Was 3 profiles × 6 groups = 18 cells at authoring. Carries an `@MX:ANCHOR` naming it the Matrix A SSOT. |
| Group constants | `GroupSpecAuditors = "spec_auditors"` … `GroupGit = "git"` | 6 constants in one `const` block. |
| `agentGroupMembership` | `var agentGroupMembership = map[string]string{` | 10 entries; `Explore` deliberately absent. |
| `profileMatrixAgentOrder` | `var profileMatrixAgentOrder = []string{` | 11 entries **including** `Explore` — display order already carries Explore even though the membership map does not. |
| `ProfileMatrixAgents()` | `func ProfileMatrixAgents() []string` | Defensive copy of the display order. |
| `DefaultProfileMatrix()` | `func DefaultProfileMatrix() map[string]map[string]config.ModelEffort` | Deep copy. |
| `AgentGroup(agent)` | `func AgentGroup(agent string) (string, bool)` | Returns group + membership flag. |
| `ResolveAgentModelEffort` | `func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, hasGroup bool)` | 4-step precedence: override → config `Profiles` cell → Go default cell → unknown-profile fallback to the `medium` column. Unmapped agent short-circuits to `{modelInherit, ""}, false`. |

Pre-M1 18-cell values (the `medium` column is the one the template agent frontmatter mirrors):

```
max:    spec_auditors fable/medium · develop fable/low · advisor fable/medium
        design_harness_e2e opus/high · docs sonnet/medium · git sonnet/low
medium: spec_auditors opus/high · develop opus/high · advisor fable/low
        design_harness_e2e opus/medium · docs sonnet/medium · git sonnet/low
low:    spec_auditors opus/low · develop opus/medium · advisor opus/high
        design_harness_e2e opus/low · docs sonnet/medium · git sonnet/low
```

**Delta note**: the incoming 33-cell matrix changes many of these values, not only their shape. `max/spec_auditors` moves from `fable/medium` to `fable/low`, and `max/develop` moves from `fable/low` to `opus/xhigh` — an inversion of the develop cell's model class. M1 is therefore a value change as well as a structural change, and the fidelity test must be rewritten wholesale rather than reshaped.

### §B.1a Post-landing measurement (successor-specific, this tree)

M1 is recorded as landed in squash `31da99a7b`, but two of its plan steps are observably unexecuted. Measured in this worktree:

```
$ grep -n 'agentGroupMembership' internal/template/profile_matrix.go
204:// agentGroupMembership is the agent-name → group SSOT (REQ-MPM-011). Agents with
209:var agentGroupMembership = map[string]string{
465:	g, ok := agentGroupMembership[agent]

$ grep -n '^func AgentGroup' internal/template/profile_matrix.go
464:func AgentGroup(agent string) (string, bool) {

$ grep -n 'AgentGroup' internal/web/agentfm.go
491:		if _, ok := template.AgentGroup(a.Name); !ok {
```

`AgentGroup` is therefore not merely undeleted — it is a **live gate** in the web console. A `main`-lineage reading of the same file gives different line numbers (152 / 402); the numbers above are this tree's, on the develop lineage.

Card **t1037** owns the disposition: is the landing record wrong, or is the M1 plan stale? This SPEC does not decide it and does not remove the code.

### §B.2 Tests pinning the current contract — `internal/template/profile_matrix_test.go`

| Test | What it pins | M1 action |
|---|---|---|
| `TestResolveAgentModelEffort_MatrixAFidelity` | All 10 mapped agents against the `max` column, agent-keyed already | rewrite values; add `Explore` |
| `TestResolveAgentModelEffort_LowColumn` | 3 spot-checks on `low` | rewrite values |
| `TestResolveAgentModelEffort_OverridePrecedence` | Override wins; **a sibling in the same group is unaffected** | the "sibling in same group" premise dies with groups — retarget to "another agent is unaffected" |
| `TestResolveAgentModelEffort_Inherit` | `Explore` **and** `some-user-agent` both → `hasGroup=false`, `inherit` | **split** (REQ-MPMC-018) |
| `TestResolveAgentModelEffort_LegacyAlias` | `performance_tier: max` with no `profile` resolves the max column | keep; rewrite the expected value |
| `TestDefaultProfileMatrix_NoHaiku` | Zero haiku | keep as-is |
| `TestApplyProfile_InsertsProfileWhenAbsent` | Migration insert path | unaffected |

`TestResolveAgentModelEffort_ConfigProfilesOverrideDefault` also pins a behaviour that dies — but by `SPEC-MODEL-MATRIX-CONFIG-001`'s M2, not by M1, so it is listed in that successor's research rather than here.

### §B.3 Resolver consumers

The only production consumer of the resolver trio (`ProfileMatrixAgents` / `AgentGroup` / `ResolveAgentModelEffort`) is `internal/cli/model.go` `resolveModelProfileReport`, which iterates the display order and emits a per-agent report carrying `Agent`, `Group`, `Model`, `Effort`, and — under a GLM backend — `GLMModel` and `GLMReasoning`. The report is rendered as a human table or as JSON (`--json`).

**M1 impact on this consumer**: the `Group` field of `modelProfileEntry` loses its source. Either the field is dropped from the report (a JSON shape change) or it is retained as a constant `"-"`. Design decision: `design.md` §B.2.

`internal/web/agentfm.go:491` is a second `AgentGroup` consumer, measured above in §B.1a and not present in the original research.

### §B.4 Agent frontmatter — current state

Deployed tree `.claude/agents/moai/*.md` (10 files; `Explore` has no file), as observed at authoring:

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

**Important mismatch found**: the template baseline should hold the Medium-profile values. Under the incoming 33-cell matrix the Medium column is `manager-spec opus/high`, `plan-auditor opus/xhigh`, `manager-develop opus/high`, `super-advisor fable/low`, `manager-design opus/high`, `builder-harness opus/high`, `e2e-tester opus/high`. Comparing to the table above, several current effort values differ from the incoming Medium column. The downstream "no-op at medium" claim — owned by `SPEC-MODEL-MATRIX-CONFIG-001` — holds only **after** the template baseline is re-set by M1 step 7.

### §B.5 Matrix literal duplication

At authoring the literal existed in four places: the Go constant, both `llm.yaml` copies, and the fidelity test's `want` map. Reducing that to two is `SPEC-MODEL-MATRIX-CONFIG-001`'s M2 obligation; this SPEC's step 7 only re-sets the template agent frontmatter, not the config mirrors.

---

## §C Two-channel effort injection — schema evidence (inherited context)

| Fact | Source |
|---|---|
| `Agent` tool accepts `model` with enum `sonnet\|opus\|haiku\|fable`; has no `effort` parameter | live tool schema available in the authoring session |
| `Workflow` tool `agent()` accepts `opts.model` and `opts.effort ∈ {low, medium, high, xhigh, max}`, plus `opts.agentType` to target a named subagent | live tool schema available in the authoring session |

Note the model-enum asymmetry: the `Agent` tool enum includes `haiku` (which MoAI policy forbids) and excludes `inherit` (which the resolver returns for unmapped agents). The unmapped-agent path must therefore **omit** the `model` argument entirely rather than pass `inherit` — constraint C-4.

The consequences for the effort channels belong to `SPEC-MODEL-MATRIX-CONFIG-001`; recorded here because the resolver contract this SPEC ships is what both channels read.

---

## §F Open questions for run-phase

- **[NEEDS CLARIFICATION: `Group` field in `moai model profile --json`]** — M1 removes the group concept. Does the JSON report drop the `group` key (a breaking shape change for any consumer) or retain it as a constant `"-"`? No external consumer was found, but the `--json` flag is a documented public surface, and `SPEC-MODEL-MATRIX-CONFIG-001` elevates this JSON to a consumed contract.

The three other clarifications carried by the original SPEC travelled with their milestones: the wizard option vocabulary to `SPEC-MODEL-MATRIX-SURFACES-001`, the `profiles:` migration representability to `SPEC-MODEL-MATRIX-CONFIG-001`, and the infographic disposition to `SPEC-MODEL-MATRIX-DOCS-001`.

---

## §G Explicitly NOT verified

Inherited unverified inputs whose owning unit is S0 or M1:

- The DeepSWE `[low]` Fable 5 row and the `[high]` Opus 4.8 row (§A.3) — the probe surfaced `[max]` rows for those models.
- The canonical GLM 5.2 Pass@1 point value (§A.4) — observed as `44% ±2%`, which matches neither user reading exactly.
- The token-column semantics of the leaderboard (§A.3) — raw output tokens vs tokens-per-solved is unresolved.

Successor-specific addition:

- **Whether the landed matrix actually equals `spec.md` §A.3.** `progress.md` §E.2 records the landed distribution as Opus 25 / Sonnet 8 / Fable 0 / `xhigh` 0 / `max` in 2 cells, while §A.3 carries six Fable cells, two `xhigh` cells, and no `max` effort. The two descriptions are inconsistent and the inconsistency was not reconciled before the split. Re-measure before treating M1 as discharged.

The remaining unverified inputs from the original SPEC belong to other successors: the ja/ko `multi-llm/model-policy.md` shape, the `advanced/*` line ranges, the infographic contents, and cross-platform build status to `SPEC-MODEL-MATRIX-DOCS-001`; the `mp.*` i18n orphan status to `SPEC-MODEL-MATRIX-SURFACES-001`.
