# Research — SPEC-MODEL-MATRIX-CONFIG-001

Evidence gathered during the plan-phase authoring of `SPEC-MODEL-PROFILE-MATRIX-002` and inherited by this successor on the M2 + M3 seam. Every claim is attributed to the command or read that produced it. Unobserved items are named explicitly in §G.

All paths relative to the repository root. Line numbers are drift-prone; the accompanying content tokens are the durable anchors.

---

## §A The 36-cell axis

```
grep -rn "model_routing_profiles|RouteModelFor" --include="*.go" --include="*.yaml"
```

Production surfaces:

- `.moai/config/sections/workflow.yaml` — the `model_routing_profiles:` block plus 3 explanatory comment lines.
- `internal/config/model_routing.go` — `RouteModelFor(specTier, phase, perfTier)` with an `@MX:ANCHOR` calling it "the spawn-time cost-routing accessor", plus 4 validator branches emitting `model_routing_profiles.*` field errors.
- `internal/config/types.go` — `ModelRoutingProfiles` field and its type, with doc comments referencing `RouteModelFor` fallback behaviour.

Test surfaces: `internal/config/model_routing_test.go` carries a full-matrix YAML fixture and several `RouteModelFor` cases.

**Confirmed**: every `RouteModelFor` call outside `internal/config/model_routing_test.go` is within `model_routing.go` itself. The `@MX:ANCHOR` claiming "spawn-time cost-routing accessor" describes an intended consumer that was never wired — this is the "zero non-test production call sites" fact, and it is what makes retirement-rather-than-migration safe.

**Post-split status**: this block is **still present**. M2 never started; `model_routing_profiles` remains in `workflow.yaml`.

---

## §B Matrix literal duplication (4 copies at authoring)

1. `internal/template/profile_matrix.go` — `defaultProfileMatrix` (Go constant, authoritative).
2. `.moai/config/sections/llm.yaml` — `llm.profiles` (flow-style, 18 cells at authoring).
3. `internal/template/templates/.moai/config/sections/llm.yaml` — `llm.profiles` (aligned style, plus ~20 lines of explanatory comment describing the six groups).
4. `internal/template/profile_matrix_test.go` — the `want` map in `TestResolveAgentModelEffort_MatrixAFidelity`.

After M2, copies 2 and 3 are removed, leaving the Go constant and the test (REQ-MPME-005).

**Additional finding**: the two `llm.yaml` copies have already drifted in a second, unrelated place. The local copy's `glm.models` block carries `opus`, `sonnet`, and `haiku` alias keys; the template copy carries only `high`, `medium`, `low`, `fable`. This drift is **out of scope** but is recorded because M2 edits both files and a naive whole-block sync would silently import the local-only alias keys — including `haiku` — into the template, which would trip the haiku-residual rule's `claude_models` surface neighbourhood. M2 must edit the `profiles:` block only.

---

## §C Two-channel effort injection — schema evidence

| Fact | Source |
|---|---|
| `Agent` tool accepts `model` with enum `sonnet\|opus\|haiku\|fable`; has no `effort` parameter | live tool schema available in the authoring session |
| `Workflow` tool `agent()` accepts `opts.model` and `opts.effort ∈ {low, medium, high, xhigh, max}`, plus `opts.agentType` to target a named subagent | live tool schema available in the authoring session |
| Workflow `agent()` effort is validated against that closed set in MoAI's own doctrine | `.claude/rules/moai/workflow/dynamic-workflows.md` |

Note the asymmetry: the Workflow effort set includes `max`, while the 33-cell matrix uses at most `xhigh`. No matrix cell can therefore produce an out-of-range `opts.effort` value.

Note also the model-enum asymmetry: the `Agent` tool enum includes `haiku` (which MoAI policy forbids) and excludes `inherit` (which the resolver returns for unmapped agents). The unmapped-agent path must therefore **omit** the `model` argument entirely rather than pass `inherit`.

---

## §D Precedent — how the predecessor's retirement was justified

`SPEC-MODEL-PROFILE-MATRIX-001` REQ-MPM-040 (quoted from the comment block at `internal/web/agentfm.go`):

> the former tier-profile re-application (which rewrote each shipped agent's `model:`/`effort:` and re-introduced the `[1m]`-hazard concrete-model pin) is retired — the web save now persists to `llm.yaml` only, leaving agent frontmatter at `model: inherit`.

The stated cause of retirement is **the `model:` pin**, not the `effort:` write. The revival keeps `model: inherit` untouched and writes only `effort:`, so it does not reintroduce the cited hazard. This is the load-bearing argument for the revival and is why REQ-MPME-011 is phrased as a prohibition rather than a preference.

**M3 consequence**: that comment block at `applyPerfTierEdits` becomes **false** the moment M3 lands. It must be rewritten, not merely left in place, at the same time the seam is re-wired.

---

## §E Prior-art within the repo for the M3 mitigations

| Mitigation | Existing precedent |
|---|---|
| Frontmatter-only edit | `internal/settings/agentfm/agentfm.go` `Patch` — writes frontmatter keys, leaves body bytes untouched |
| Live-files-only, no template dual-write | same `Patch`; its only caller `internal/web/agentfm.go` operates on deployed paths |
| Post-deploy reapplication ordering | `internal/cli/update.go` already calls `ApplyProfile` after the deploy step in the `--profile` flag path |
| Override-respecting exclusion | `ResolveAgentModelEffort` step 1 already short-circuits on `cfg.AgentOverrides[agent]` |

All four mitigations reuse an existing in-tree mechanism; none requires a new subsystem.

---

## §F The four M3 seams

```
grep -rn "ApplyProfile(" --include="*.go" . | grep -v _test.go
  internal/web/agentfm.go
  internal/template/profile_matrix.go   (definition)
  internal/cli/update.go                (profile flag path)
  internal/cli/update.go                (wizard path)
  internal/cli/init.go
```

Four call sites plus the definition — these are the M3 seams (REQ-MPME-016). Only the `update.go` profile-flag seam is confirmed to sit after a template deploy; the other three must be checked individually.

`internal/settings/agentfm/agentfm.go` exposes `Patch(path, model, effort string, deleteEffort bool) error` — the existing frontmatter-only editor. Its sole production caller is `internal/web/agentfm.go`. It is the natural implementation vehicle for REQ-MPME-010/011/018, **provided** it can express "leave `model:` unchanged"; if it cannot, that capability is added rather than worked around.

### Guard tests — what does and does not exist

| Guard | File | Pins effort? |
|---|---|---|
| Agent frontmatter audit | `internal/template/agent_frontmatter_audit_test.go` | not for effort values (audits retired fields, tools/skills shape) |
| Agents frontmatter CSV/mutual-exclusion | `internal/template/agents_frontmatter_test.go` | **NO** — inspected in full (190 lines); validates only `tools:` / `disallowedTools:` CSV format and mutual exclusion. Reads no `model:` or `effort:` key. |

**`agents_frontmatter_test.go` is CONFIRMED not to pin effort values.** No CI guard enforces local-vs-template agent frontmatter byte parity, so M3's deployed-only rewrite (REQ-MPME-012) will not trip an existing test — which also means nothing will catch a regression on it except the tests M3 itself writes.

### Haiku-residual coupling

The rule's four surfaces, from its own header comment:

1. Agent frontmatter/body in `.claude/agents/moai/*.md` + template mirror.
2. `claude_models` block in `llm.yaml` (glm.models exempt — exemption X2).
3. `model_routing_profiles` / `workflow_agents` / `role_profiles` in `workflow.yaml`.
4. `validRoutingModels` Go map in `internal/config/model_routing.go`.

**Surfaces 3 and 4 both target artifacts M2 deletes.** Leaving them in place after M2 means the rule scans for a block that cannot exist — harmless but misleading. Their removal is `SPEC-MODEL-MATRIX-DOCS-001`'s M7 obligation; this SPEC must hand it off explicitly rather than assume it.

---

## §F.1 Open questions for run-phase

- **[NEEDS CLARIFICATION: `profiles:` migration representability]** — REQ-MPME-007 splits on whether a user's `profiles:` customization is "representable as per-agent overrides". A group-keyed cell expands to 1-3 agent overrides mechanically, so representability is arguably always true. Should the warning branch exist at all, or should migration be unconditional with a summary notice?

One clarification owned elsewhere gates this SPEC: the `Group`-field disposition in `moai model profile --json`, owned by `SPEC-MODEL-MATRIX-CORE-001`. M3 cannot elevate that JSON to a consumed contract against an undecided shape.

The two remaining clarifications from the original SPEC travelled with their milestones: the wizard option vocabulary to `SPEC-MODEL-MATRIX-SURFACES-001`, the infographic disposition to `SPEC-MODEL-MATRIX-DOCS-001`.

---

## §G Explicitly NOT verified

No unverified input from the original SPEC's §G list is owned by M2 or M3 — all eight belong to S0, M5, M6, or M7 and travelled to the other successors.

Two items are added by this split and are unverified here:

- **Whether `Patch`'s current signature can express "leave `model:` unchanged".** `design.md` §A.2 requires the capability and states it is added if absent; which of the two applies was not measured during authoring.
- **Whether the three non-`--profile` seams sit after a template deploy.** Only the `update.go` profile-flag seam was confirmed. `init.go` is expected to have no prior deploy to race and `web/agentfm.go` no deploy at all, but neither was read.
