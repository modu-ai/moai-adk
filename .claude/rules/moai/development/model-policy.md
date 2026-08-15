---
paths: "**/.claude/agents/**"
---

# Model Policy

Rules for agent model field values and multi-model architecture.

## Valid Model Field Values

Agent definition `model` field accepts only these values:
- inherit: Uses parent session's model (default)
- opus: Claude Opus (highest capability)
- sonnet: Claude Sonnet (balanced)
- fable: Claude Fable (current generation; added to the model enum per CC v2.1.196 model-priority update)
- haiku: Claude Haiku (fastest, lowest cost)

Current model generation mapping:
- opus = Opus 5 (`claude-opus-5`) on the Anthropic API — the default Opus model as of CC 2.1.219, native 1M context, fast mode available (default effort: high across all surfaces incl. Claude Code; set xhigh explicitly for coding/agentic work; Opus 5 carries a previously-set effort level across sessions — no hold)
- sonnet = Sonnet 5 on the Anthropic API (current generation; native 1M window, no `[1m]` suffix, no usage credits — CC 2.1.197). Behind an LLM gateway or with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K. See § Sonnet 5 Native-1M Re-scope (CC 2.1.198). Note that `CLAUDE_CODE_DISABLE_1M_CONTEXT` is not sonnet-specific: per the CC 2.1.223 changelog it holds **every** Claude model with a native 1M window — Opus 5 included — to 200K via auto-compaction, and a startup warning appears when auto-compaction is not holding the session to 200K. The official env-vars page still describes the flag in Sonnet-5 terms (upstream doc lag as of CC 2.1.225); the changelog is the current source.
- fable = Fable (current generation; added to the model enum per CC v2.1.196 model-priority update)
- haiku = Haiku (current generation; retired from MoAI agent routing per the No-Haiku policy — value remains valid for documentation/example YAML)

Opus 5 and Opus 4.8 serve the full 1M token context window by default (no beta header, no long-context premium). Fast mode applies to Opus 5 and Opus 4.8 as of CC 2.1.219 (Opus 4.7 was removed from fast mode). Explore (Anthropic built-in) inherits the session model per CC v2.1.198 — no separate deployment or model pin needed.

Invalid values (NEVER use):
- glm: Not a model field value (GLM is configured via environment variables)
- high/medium/low: These are CLI policy flags, not model field values
- Pinned old versions (opus-4-0, opus-4-1, sonnet-4-5): Auto-migrated to current generation
- Full model-ID form (e.g., `claude-opus-5`): **official-but-intentionally-disallowed in MoAI.** Claude Code itself accepts a full model-ID string in the `model` field, but MoAI intentionally restricts agents to the five alias values above (`inherit` / `opus` / `sonnet` / `fable` / `haiku`). The reason is the `[1m]` context-entitlement inheritance bug (see § Inherit-by-Default Convention): a subagent that pins a concrete full model ID — like an explicit `model: sonnet` / `model: opus` — does not inherit the parent session's `[1m]` entitlement and fails to spawn. The alias `inherit` sidesteps this. This restriction being a deliberate MoAI policy (not a stale gap) means "MoAI is outdated relative to Claude Code" is a misreading — the full-ID form is omitted on purpose.

## Inherit-by-Default Convention

[ZONE:Evolvable] [HARD] All MoAI agents SHOULD declare `model: inherit` unless explicitly assigned `haiku` for speed-critical tasks. The previous practice of declaring `model: sonnet` or `model: opus` is deprecated for new agents.

Rationale (Claude Code session inheritance bug):
- When the parent session uses an `[1m]` context variant (e.g., `claude-opus-5[1m]`, `claude-sonnet-5[1m]`) and a spawned subagent declares an explicit `model: sonnet` or `model: opus` in its frontmatter, the parent's 1M context entitlement does NOT propagate to the subagent.
- Result: subagent spawn fails with `API Error: Usage credits required for 1M context · run /usage-credits to turn them on, or /model to switch to standard context`.
- Sonnet **4.6** 1M specifically requires extra usage credits on every plan (including Max), so the entitlement mismatch is unrecoverable mid-spawn. (Re-scoped CC 2.1.198: on the Anthropic API `sonnet` now resolves to **Sonnet 5** with a native 1M window that needs NO usage credits and exposes no `[1m]` suffix, so this credit-mismatch failure binds Sonnet 4.6 1M and gateway-selected Sonnet paths — NOT Sonnet 5 on the Anthropic API. See § Sonnet 5 Native-1M Re-scope below.)

Upstream tracking (Anthropic claude-code repository):
- [Issue #45847](https://github.com/anthropics/claude-code/issues/45847): skill with `model: sonnet` frontmatter fails from Opus 4.6/4.7 [1m] parent
- [Issue #51060](https://github.com/anthropics/claude-code/issues/51060): subagent with `model: opus` ignores parent's Extra Usage flag
- [Issue #36670](https://github.com/anthropics/claude-code/issues/36670): Team teammates do not inherit the `[1m]` context variant from leader

Workaround pattern (`model: inherit`):
- The subagent fully inherits the parent's model + context entitlement, eliminating the mismatch.
- Reference implementation: `.claude/agents/moai/plan-auditor.md` has used `model: inherit`.
- 9 of the 10 MoAI-custom retained agents under `.claude/agents/moai/` declare `model: inherit`; only `manager-git` still pins `model: sonnet`, which matches its profile-matrix row exactly (sonnet / low in every column). `manager-docs` returned to `model: inherit` when the matrix moved it onto Opus in the high and medium columns — a static pin can no longer express a row whose model varies by column. The No-Haiku policy retired the former `model: haiku` exception — cost reduction is achieved via effort tiering, not model-class substitution.

Exceptions (do NOT migrate to inherit):
- Documentation/example YAML inside skill bodies (`.claude/skills/moai-foundation-cc/reference/**/*.md`) — these mirror official Claude Code documentation and MUST show all valid values (`sonnet`, `opus`, `haiku`, `inherit`) for educational purposes.

## Baseline-Refill Breaker (team sonnet — second failure mode; Sonnet 5 / Opus 5-resolved)

[ZONE:Evolvable] The `[1m]` entitlement bug in § Inherit-by-Default Convention is the *spawn-time* failure mode (a frontmatter `model: sonnet` pin → spawn fails with a 1M credit error). A **distinct second failure mode** historically affected team-mode teammates spawned via per-spawn `model: "sonnet"` override:

| failure mode | trigger | symptom | mitigation |
|--------------|---------|---------|------------|
| `[1m]` credit-fail | frontmatter `model: sonnet` pin | spawn fails immediately (`Usage credits required for 1M`) | use `model: inherit` |
| baseline-refill breaker | per-spawn `model: "sonnet"` in team mode on a **200K-variant** model | spawn succeeds, but the 200K window saturates under the heavy baseline → autocompact → rapid-refill → runtime circuit breaker → zero output | historical only — see resolution below |

**Resolution (Sonnet 5 / Opus 5 era):** the breaker required a teammate to fall back to a **200K context variant** after the `[1m]` suffix was stripped on teammate spawn (Anthropic issues #36670 / #34421; the suffix-strip mechanism is still OPEN upstream). The fallback target — a 200K context variant — **no longer exists in the current default lineup**: Sonnet 5 ships a single 1M-token context window (1M is both default and maximum; no smaller variant — per platform.claude.com Sonnet 5 model docs), and Opus 5 likewise serves the full 1M window by default. With no 200K variant to fall back to, a teammate spawned as `sonnet` / `opus` operates at 1M regardless of suffix stripping, and the rapid-refill → circuit-breaker → zero-output cascade cannot trigger. The mechanism (#36670) is technically still open but its observable impact on the current default lineup is zero.

The breaker therefore remains documented only as a **historical hazard for legacy 200K-variant models** (Sonnet 4.x, Opus 4.6, and Haiku 4.5 which is still 200K): on those models a teammate can still fall back to 200K. For the current default lineup the operational mitigation (single `manager-develop` + Milestone split over team mode) is no longer forced by the breaker — though team mode is additionally disabled by default per the Phase 4 re-design (`.claude/rules/moai/workflow/orchestration-mode-selection.md`), in favor of subagent fanout (Mode 4) for multi-domain research/review and sequential sub-agent (Mode 5) for coding.

## `[1m]` Constraint Re-Verification (CC 2.1.178; Sonnet 5 / Opus 5 practical-impact update)

The `[1m]` entitlement-inheritance constraint was re-verified against CC 2.1.178. **Verdict: STILL-ACTIVE mechanism, but ZERO PRACTICAL IMPACT on the current default lineup (Sonnet 5 / Opus 5).** Per-agent `model:` pins remain forbidden regardless (the re-verification records per-agent pinning as out-of-scope).

Evidence fetched via the GitHub issue API + the canonical CC CHANGELOG:

- Issue #45847 (skill with `model:` fails from `[1m]` parent): **closed**, labeled `duplicate` — no explicit "fixed" resolution.
- Issue #51060 (subagent `model: opus` spawn fails): **closed**, labeled `bug, area:model, area:agents, stale` — no CHANGELOG entry fixes the spawn-time entitlement-inheritance root cause.
- Issue #36670 (Team teammates don't inherit `[1m]` from leader): **OPEN** — the Team-mode path is confirmed unfixed at CC 2.1.178.
- CC 2.1.172 fixes ("1M context stuck session", "doubled `[1m]` suffix") address the *symptom* and *suffix normalization*, NOT the *spawn-time entitlement mismatch*. CC 2.1.173/2.1.174 are Fable-5-suffix and background-env-inheritance fixes — orthogonal.

**Why STILL-ACTIVE mechanism ≠ practical impact (Sonnet 5 / Opus 5 era):** the #36670 mechanism strips the `[1m]` suffix on teammate spawn, which historically forced a fallback to a 200K variant. Sonnet 5 and Opus 5 have **no 200K variant** — their context window is 1M as both default and maximum (per platform.claude.com Sonnet 5 model docs; Opus 5 serves the full 1M window by default). The stripped-suffix teammate therefore still resolves to 1M; there is no smaller variant to degrade into. The mechanism is unfixed upstream, but on the current default lineup it has nothing to break. The `model: inherit` convention is retained as defense-in-depth and for legacy-200K-variant models (Haiku 4.5 still ships 200K).

A follow-up SPEC (conditional) MAY re-enable per-agent pinning only when #36670 is closed-with-fixed AND a CHANGELOG confirms Team `[1m]` inheritance for explicit `model:` teammates — though for Sonnet 5 / Opus 5 the practical case for that re-enablement has dissolved.

### Sonnet 5 Native-1M Re-scope (CC 2.1.198)

Re-verified against the CC 2.1.197 Sonnet 5 release + `code.claude.com/docs/en/model-config`. **Verdict: the sonnet-side premise of the `[1m]` doctrine changed; the doctrine is re-scoped, NOT deleted.**

- On the Anthropic API, `sonnet` resolves to **Sonnet 5** with a **native 1M context window — there is no 200K variant, no `[1m]` suffix to select, and no usage credits required on any plan** (sessions auto-compact ~967K). A frontmatter `model: sonnet` pin therefore no longer trips the sonnet-side `[1m]` credit-mismatch on the Anthropic API — the credit entitlement it used to require does not exist for Sonnet 5.
- The `[1m]` doctrine still binds two surviving paths: (1) the **opus`[1m]`** path (Opus variants still expose `[1m]` selection), and (2) the **gateway / older-model** path — behind an LLM gateway (`ANTHROPIC_BASE_URL` non-Anthropic) or with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K and `sonnet[1m]` selects the 1M window for Sonnet 5. The flag's blast radius widened at CC 2.1.223 — it now holds every native-1M Claude model (Opus 5 included) to 200K, so setting it changes the Opus-side budget too, not only the sonnet-side one.
- **`model: inherit` remains the MoAI default regardless.** Sonnet 5's native 1M removes the sonnet-side spawn-time hazard on the Anthropic API, but per-agent `model:` pins stay deprecated (§ Inherit-by-Default Convention) — the opus`[1m]` + Team-mode paths still make `inherit` the safe choice.
- **Gap (NOT re-verified this run)**: issue #36670 (Team teammates don't inherit `[1m]` from leader) was **NOT re-checked** against CC 2.1.198 — its state is unknown; do NOT claim it closed. The Team-mode `[1m]` non-inheritance premise in § Baseline-Refill Breaker is unchanged pending a fresh issue-state observation.

## Default-Model Cost Lever (Default = sonnet, no allowlist enforcement)

[ZONE:Evolvable] [HARD] The `[1m]`-safe cost lever is the **Default model** set at the settings level, NOT per-agent `model:` pins. The deployed `settings.json` template (`.claude/settings.json.tmpl`) sets ONLY:

```json
"model": "sonnet"
```

The template deliberately does **NOT** set `availableModels` or `enforceAvailableModels`. A closed `availableModels` allowlist combined with `enforceAvailableModels: true` hides any model not in the list from the `/model` picker (CC v2.1.172 behavior), which caused two problems:

1. **New-model lockout** — every new Claude model (for example a new `fable` generation, or any future tier) was invisible in `/model` until an operator manually appended it to the allowlist. This recurred on every model release.
2. **GLM allowlist maintenance** — enforcement forced every GLM swap target (`glm-5.2` and the other GLM tiers) to be enumerated in the allowlist, or the swap was declined (see GLM-mode reconciliation below).

Dropping `enforceAvailableModels` resolves both at once: all Claude models (current and future) auto-appear in the picker with no maintenance, and the GLM swap is admitted without an allowlist. Only the Default-model cost lever is retained — `"model": "sonnet"` alone still routes the busy-agent cost through Sonnet by default.

Why this is `[1m]`-safe: the lever operates on the **Default** model resolution at the settings level, not on per-agent explicit pins, so it does not trigger the spawn-time entitlement-inheritance failure (#45847/#51060/#36670). The cost-routing thesis (route the busy-agent cost through Sonnet, not Opus) flows through the Default; deep-reasoning exceptions use per-spawn `Agent(model: "opus")` only for the 5-10% of tasks where Opus wins (architecture, complex perf) — and even those inherit the parent `[1m]` entitlement because they are spawned without a frontmatter `model:` pin (the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug).

### GLM-mode reconciliation

[ZONE:Evolvable] [HARD] With `enforceAvailableModels` unset, GLM mode needs no allowlist reconciliation. When GLM mode is active (`moai glm` whole-session, or the GLM teammate panes of `moai cg`), the GLM activation sets `ANTHROPIC_DEFAULT_OPUS_MODEL` to the configured GLM high model (default `glm-5.2`), surfaced in the model UI as the Opus-slot alias. The CC 2.1.176 redirect-blocking semantics — which decline an `ANTHROPIC_DEFAULT_*_MODEL` redirect to a model NOT in `availableModels` — apply ONLY when `enforceAvailableModels` is `true`. Because the template no longer sets that flag, the GLM swap is never checked against an allowlist and is admitted directly; the session runs on the configured GLM model instead of silently falling back to Sonnet.

This supersedes the earlier approach of enumerating the GLM model ids in `availableModels` (the `[1m]`-variant + raw-GLM-id expansion). That expansion existed only to satisfy `enforceAvailableModels: true`; removing the enforcement flag removes the need for the expansion entirely. The Default model stays `sonnet` — a non-GLM (`moai cc` / plain Claude) session still resolves its Default to Sonnet; the only change is that no model is hidden and no swap is declined.

Scope note: this is a **static template change** in `.claude/settings.json.tmpl` (removal of the `availableModels` + `enforceAvailableModels` keys). It touches no Go runtime code (`glm.go` / `launcher.go` / `settings.go` unchanged) and writes nothing to `settings.local.json` — so the solo `moai glm` "settings.local.json clean" design (no GLM env leak to subsequent plain-`claude` invocations) is preserved.

## Model Policy Tiers (3-tier — high/medium/low)

Model policy is set via `moai init --model-policy <tier>`. The 3-tier system is `high` / `medium` / `low`. The top tier was formerly named `max`; that name is still READ as an alias for `high` so pre-rename configs keep resolving, but it is never written back. The rename unified three vocabularies that previously disagreed — `llm.profile`, the legacy `llm.performance_tier`, and the `ModelPolicy` CLI tokens are now all `{high, medium, low}`, so the former `high -> max` projection is an identity and no migration pass is required. Under the No-Haiku policy (SPEC-AGENT-ARCH-V2-001 §D) haiku is absent from every tier; the tier governs (a) where Fable and Opus are deployed and (b) how aggressively effort is lowered. The `model_routing_profiles.{high,medium,low}` matrices in `workflow.yaml` are the Tier×Phase config SSOT, and `llm.profiles.{high,medium,low}` is the per-agent SSOT.

| Tier | Philosophy | Top-model deployment | Effort baseline |
|------|------------|-----------------|------------------------|
| `high` | Quality first — Opus throughout; `max` reserved for the two rows where the marginal point is decision-critical | Opus on every row except the two single-shot procedural rows (manager-git, Explore); Fable 0 | `max` on manager-develop + super-advisor; `high` on manager-spec/plan-auditor/sync-auditor/design/harness; `medium` on e2e; `low` on docs; procedural `low` |
| `medium` (default) | Phase-weighted — the phase that produces code carries the ceiling; the phases that frame and record it step down | Opus on every reasoning / coding / authoring agent; Fable 0 | `high` on manager-spec/plan-auditor/manager-develop/sync-auditor/super-advisor; `medium` on design/harness; e2e `low`; docs `low`; procedural `low` |
| `low` | Cost minimum — Opus `low` still outscores Sonnet at any effort AND costs less per task | Opus retained on every agentic row; Sonnet on docs / e2e / git / Explore | Agentic `low`; super-advisor `medium`; procedural `low` |

The per-agent cells are the `llm.profiles` matrix (11 retained agents × 3 columns = 33 cells; Go SSOT `template.DefaultProfileMatrix`). Every agent row is monotone across `high >= medium >= low`. The matrix encodes two model rules, both measured on a long-horizon coding-agent benchmark that reports score, cost per task, output tokens, and agent steps at every effort level:

1. **Opus dominates Sonnet at every effort.** Opus at `low` scores higher AND costs less per task than Sonnet at any level, because Sonnet spends a multiple of the agent steps and output tokens to finish the same long-horizon task. Completion efficiency — not unit token price — drives cost, so Opus is the model for every multi-turn agentic row.
2. **Sonnet is retained only for single-shot, input-dominated, non-agentic rows** (`Explore` search, `manager-git` mechanics, plus the docs / e2e rows in the `low` column), where that multi-step completion failure does not apply and the lower input price does.

`xhigh` is absent from the matrix on purpose: on Opus it scores the same as `high` at materially higher cost, so it is strictly dominated. `max` is reserved for the two rows whose invocation is rare AND whose marginal quality is decision-critical (`manager-develop`, `super-advisor` — high column only).

## Per-Agent Profile Resolver (model injection source)

The per-agent model+effort **profile** (config `llm.profile` ∈ {high, medium, low}; the superseded `max` is read as `high`) is the runtime-arg **model** injection source the orchestrator reads at spawn time. Query it with the read-only accessor:

```bash
moai model profile          # human table
moai model profile --json   # machine-readable
```

The resolver maps the active profile + optional `llm.agent_overrides` to each retained agent's `{model, effort}` by agent NAME. Lookup is per-agent, not per-group: per-agent cells split two of the former groups, so the group layer no longer carries routing information and survives only as a display classification. Precedence: `agent_overrides[agent]` → active profile cell (config `llm.profiles`) → Go-default cell → `inherit`. Only agents outside the retained catalog (any user-added agent) resolve to `inherit`; `Explore` is an explicit matrix row. A pre-rename `llm.profiles` mirror still keyed by GROUP name simply misses on lookup and falls through to the Go default — a stale mirror degrades, it does not break.

The resolved **model** is the value the orchestrator injects as a per-spawn `Agent(model: <alias>)` runtime arg — `[1m]`-safe and distinct from the frontmatter `model:` field (see § Inherit-by-Default Convention), so a profile change never re-introduces the concrete-frontmatter-`model:` spawn-failure risk. Agent `.md` frontmatter stays at `model: inherit`; no init / update / web save mutates it.

The resolved **effort** reaches an agent through a **path-dependent channel**, and the channel determines which value is actually effective:

| Channel | Orchestration path | `model` | `effort` | Effective effort source |
|---------|--------------------|---------|----------|-------------------------|
| `Agent` tool | sub-agent delegation (the standard path) | runtime arg (`sonnet\|opus\|haiku\|fable`) | **no parameter exists** | the agent file's frontmatter `effort:` |
| `Workflow` `agent()` | dynamic-workflow | `opts.model` | `opts.effort` (structured, `{low,medium,high,xhigh,max}`) | the injected `opts.effort` |

Because the `Agent` tool has no `effort` parameter, the shipped agent frontmatter `effort:` is the load-bearing value on the standard path — it is not merely documentation. The template tree pins each agent's frontmatter to the **medium** column so a deployment's baseline matches the default profile; a non-medium profile is delivered by the frontmatter-rewrite path, never by a per-spawn arg. The GLM effort overlay (`CollapseClaudeEffortToGLM` + the `manager-develop` coding-max override) consumes the same value under a GLM backend.

**`Explore` is the one exception**: it is an Anthropic built-in with **no agent file on disk**, so neither channel above can carry its effort — there is no frontmatter to pin and no `effort` parameter to pass. Its effort is therefore **stated at call time** in the spawn prompt alongside the search-breadth qualifier, and is never persisted. The default is `low` (a read-only breadth-first sweep, the vendor-documented `low` use case for subagents); raise it to `medium` in the same prompt when the request asks for a `very thorough` sweep, so breadth and effort move together. The matrix row for `Explore` records `sonnet / low` as the call-time default, not as an injected value.

## Harness-Agent Model Policy (model-uniform, effort-differentiated)

Generated harness specialists (`/moai:harness`, `.claude/agents/harness/`) sit OUTSIDE the retained catalog, so `ResolveAgentModelEffort` returns `inherit` for them and no per-spawn injection applies — their frontmatter is the only channel. They are therefore assigned at generation time, and they are **model-uniform on purpose**: every generated harness agent is pinned to `opus`, and the only axis that differentiates them is `effort`. The rationale is that a harness specialist is a persistent, user-owned worker whose distinction is reasoning DEPTH rather than model tier.

Pinning is safe on the current lineup: Fable 5, Opus 5, and Sonnet 5 all carry a 1M context window, so an explicit pin no longer loses the 1M entitlement that the inherit-by-default convention (§ Inherit-by-Default Convention) exists to preserve. Only Haiku 4.5 remains 200K, and haiku is retired from MoAI routing.

The effort comes from `llm.harness_agents[<profile>][<purpose class>].effort`, whose 7 purpose classes share the `workflow_agents` taxonomy vocabulary in `workflow.yaml`. Each class inherits the effort of one profile-matrix row:

| purpose class | derives effort from | typical harness role |
|---------------|---------------------|----------------------|
| `read-only-extract` | `Explore` | search / reconnaissance worker |
| `mechanical-transform` | `manager-git` | translation, format conversion, release mechanics |
| `synthesize` | `manager-docs` | content authoring |
| `research` | `plan-auditor` | upstream tracking, investigation |
| `verify-judge` | `sync-auditor` | quality / audit specialist |
| `implement` | `manager-develop` | code, hook, CI specialist |
| `design-architecture` | `manager-design` | structure / navigation curation |

An unrecognized class falls back to `implement`. Because harness agents are **user-owned** (`moai update` never modifies them), assignment happens only at generation time: a later `llm.profile` change does NOT retroactively rewrite an existing harness agent's frontmatter. Re-alignment is an explicit user action (`/moai:harness` re-run or a manual edit).

## CG Mode

CG Mode (Claude + GLM) uses environment variable overrides, not model field changes:
- Leader session: Uses Claude models (no GLM env)
- Teammate sessions: Inherit GLM env from tmux session
- Activation: `moai cg` (requires tmux)

## Effort Levels

Claude models support five effort levels that control reasoning depth. The set is `low` → `medium` → `high` → `xhigh` → `max`; availability is per-model — `max` is offered on Fable 5, Opus 5, Sonnet 5 and the 4.6+ Opus/Sonnet generations, while `xhigh` is the newer level and is offered on Fable 5, Opus 5, Opus 4.8/4.7 and Sonnet 5. Haiku 4.5 supports neither. Opus 5 calibration:

- max: unconstrained token spend; reserve for tasks that justify it
- xhigh: the recommended starting point for coding and agentic work
- high: the default on Opus 5 / Opus 4.8 across all surfaces (API and Claude Code); minimum for intelligence-sensitive work
- medium: balanced cost step-down for work that can trade off intelligence
- low: short, scoped, latency-sensitive tasks — and the documented level for read-only subagents

On Opus 5, `low` and `medium` are stronger than on earlier Opus models, so they are the primary control for token cost and response time wherever quality holds. Effort settings carried over from an earlier model generation should be re-swept rather than reused. Note that on Opus 5, thinking cannot be disabled at `xhigh` or `max` (a request setting `thinking: disabled` at those levels is rejected); MoAI never disables thinking, so this constraint does not bind any MoAI path.

The per-agent effort defaults across the 3-tier profile system (high/medium/low) are the `llm.profiles` matrix (33 cells; Go SSOT `template.DefaultProfileMatrix`). The `model_routing_profiles` cells in `workflow.yaml` carry the `{model, effort}` pair for each Tier-Phase × perfTier combination. Two matrix cells use `max` effort — `manager-develop` and `super-advisor`, both in the `high` column — and no cell uses `xhigh`, which on Opus scores the same as `high` at materially higher cost, leaving `max` as the only level above `high`.

The `medium` column is phase-weighted rather than cost-minimal: run (`manager-develop`) and review (`sync-auditor`) take `high`, plan (`manager-spec`, `plan-auditor`) steps down to `high`, and sync (`manager-docs`) steps down to `low`. The column departs from the `medium` knee of the Opus cost/score curve in both directions, so it is not the benchmark-derived cost floor and must not be re-derived back onto that anchor. Under a GLM backend these per-agent cells are not what reaches the wire: z.ai's 3-state reasoning control collapses `{medium, high}` onto one state and `{xhigh, max}` onto another, but the delivery channel is session-global — the launcher injects one `ANTHROPIC_REASONING_EFFORT` per session, derived from `SessionGLMReasoningState()`, never a per-agent value. The cells record per-agent intent and take effect on Claude-backed sessions, where agent frontmatter carries the effort.

Note: `ultrathink` is a Claude Code one-turn keyword that requests deeper reasoning for that prompt; MoAI standardizes it to `effort: xhigh` (the coding/agentic level above) for that turn.

## Rules

- Agent `model` field must be one of: inherit, opus, sonnet, fable, haiku
- [ZONE:Evolvable] [HARD] New agent definitions SHOULD use `model: inherit` (default); explicit `sonnet`/`opus` are deprecated due to Claude Code Issue #45847/#51060 (see Inherit-by-Default Convention)
- `model: haiku` is retired from MoAI agent routing per the No-Haiku policy (SPEC-AGENT-ARCH-V2-001 §D); the HaikuResidualRule lint enforces 0 haiku references in agent frontmatter, claude_models, model_routing_profiles, workflow_agents, and the retired Agent Teams `role_profiles` surface (historical configs). Former haiku slots use `sonnet` with `effort: low`.
- GLM is configured via env vars in settings.json, never via model field
- Model policy tier (high/medium/low) is a CLI concern, not an agent definition concern
- CG Mode uses tmux session-level env isolation for model routing
- Old model versions are auto-migrated: do not pin to specific version IDs
