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
- haiku: Claude Haiku (fastest, lowest cost)

Current model generation mapping:
- opus = Opus 4.8 (default effort: high across all surfaces incl. Claude Code; set xhigh explicitly for coding/agentic work)
- sonnet = Sonnet (current generation)
- haiku = Haiku (current generation)

Opus 4.8 serves the full 1M token context window by default (no beta header, no long-context premium). Fast mode (speed: "fast") is a research preview for higher output throughput.

Invalid values (NEVER use):
- glm: Not a model field value (GLM is configured via environment variables)
- high/medium/low: These are CLI policy flags, not model field values
- Pinned old versions (opus-4-0, opus-4-1, sonnet-4-5): Auto-migrated to current generation
- Full model-ID form (e.g., `claude-opus-4-8`): **official-but-intentionally-disallowed in MoAI.** Claude Code itself accepts a full model-ID string in the `model` field, but MoAI intentionally restricts agents to the four alias values above (`inherit` / `opus` / `sonnet` / `haiku`). The reason is the `[1m]` context-entitlement inheritance bug (see § Inherit-by-Default Convention): a subagent that pins a concrete full model ID — like an explicit `model: sonnet` / `model: opus` — does not inherit the parent session's `[1m]` entitlement and fails to spawn. The alias `inherit` sidesteps this. This restriction being a deliberate MoAI policy (not a stale gap) means "MoAI is outdated relative to Claude Code" is a misreading — the full-ID form is omitted on purpose.

## Inherit-by-Default Convention

[ZONE:Evolvable] [HARD] All MoAI agents SHOULD declare `model: inherit` unless explicitly assigned `haiku` for speed-critical tasks. The previous practice of declaring `model: sonnet` or `model: opus` is deprecated for new agents.

Rationale (Claude Code session inheritance bug):
- When the parent session uses an `[1m]` context variant (e.g., `claude-opus-4-7[1m]`, `claude-sonnet-4-6[1m]`) and a spawned subagent declares an explicit `model: sonnet` or `model: opus` in its frontmatter, the parent's 1M context entitlement does NOT propagate to the subagent.
- Result: subagent spawn fails with `API Error: Usage credits required for 1M context · run /usage-credits to turn them on, or /model to switch to standard context`.
- Sonnet 1M specifically requires extra usage credits on every plan (including Max), so the entitlement mismatch is unrecoverable mid-spawn.

Upstream tracking (Anthropic claude-code repository):
- [Issue #45847](https://github.com/anthropics/claude-code/issues/45847): skill with `model: sonnet` frontmatter fails from Opus 4.6/4.7 [1m] parent
- [Issue #51060](https://github.com/anthropics/claude-code/issues/51060): subagent with `model: opus` ignores parent's Extra Usage flag
- [Issue #36670](https://github.com/anthropics/claude-code/issues/36670): Team teammates do not inherit the `[1m]` context variant from leader

Workaround pattern (`model: inherit`):
- The subagent fully inherits the parent's model + context entitlement, eliminating the mismatch.
- Reference implementation: `.claude/agents/moai/plan-auditor.md` has used `model: inherit`.
- All package agents under `.claude/agents/moai/` (7 retained agents per FLAT v.2.x baseline) declare `model: inherit`, except `manager-docs` and `manager-git` which use `model: haiku`.

Exceptions (do NOT migrate to inherit):
- `model: haiku` agents (`manager-docs`, `manager-git`) — Haiku has no `[1m]` variant, so the bug does NOT apply. Speed-critical agents should remain on `haiku` for cost and latency.
- Documentation/example YAML inside skill bodies (`.claude/skills/moai-foundation-cc/reference/**/*.md`) — these mirror official Claude Code documentation and MUST show all valid values (`sonnet`, `opus`, `haiku`, `inherit`) for educational purposes.

## Baseline-Refill Breaker (team sonnet — second failure mode)

[ZONE:Evolvable] The `[1m]` entitlement bug in § Inherit-by-Default Convention is the *spawn-time* failure mode (a frontmatter `model: sonnet` pin → spawn fails with a 1M credit error). A **distinct second failure mode** affects team-mode teammates spawned via per-spawn `model: "sonnet"` override (`.claude/skills/moai/team/run.md`, `.moai/config/sections/workflow.yaml` role_profiles):

| failure mode | trigger | symptom | mitigation |
|--------------|---------|---------|------------|
| `[1m]` credit-fail | frontmatter `model: sonnet` pin | spawn fails immediately (`Usage credits required for 1M`) | use `model: inherit` |
| baseline-refill breaker | per-spawn `model: "sonnet"` in team mode | spawn succeeds, but the 200K window saturates under the heavy baseline → autocompact → rapid-refill → runtime circuit breaker → zero output | avoid large SPECs in team mode (see below) |

The breaker mode is NOT reliably fixed by switching to `model: inherit`, because Team teammates do not inherit the leader's `[1m]` entitlement (Anthropic issue #36670, OPEN) — the teammate falls back to 200K and the breaker can recur. The robust mitigation is operational: for **large SPECs**, prefer a single `manager-develop` (`model: inherit`, 1M window) + Round split (`.claude/rules/moai/development/sprint-round-naming.md`) over team mode; reserve team mode for **small SPECs** where the 200K window has headroom. Cross-reference: `.claude/rules/moai/workflow/team-protocol.md` § Role Matrix Constraints; `.claude/skills/moai/team/run.md` § Baseline-Refill Breaker Hazard.

## `[1m]` Constraint Re-Verification (CC 2.1.178)

The `[1m]` entitlement-inheritance constraint was re-verified against CC 2.1.178. **Verdict: STILL-ACTIVE (conservative).** Per-agent `model:` pins remain forbidden regardless of this verdict (the re-verification records per-agent pinning as out-of-scope).

Evidence fetched via the GitHub issue API + the canonical CC CHANGELOG:

- Issue #45847 (skill with `model:` fails from `[1m]` parent): **closed**, labeled `duplicate` — no explicit "fixed" resolution.
- Issue #51060 (subagent `model: opus` spawn fails): **closed**, labeled `bug, area:model, area:agents, stale` — no CHANGELOG entry fixes the spawn-time entitlement-inheritance root cause.
- Issue #36670 (Team teammates don't inherit `[1m]` from leader): **OPEN** — the Team-mode path is confirmed unfixed at CC 2.1.178.
- CC 2.1.172 fixes ("1M context stuck session", "doubled `[1m]` suffix") address the *symptom* and *suffix normalization*, NOT the *spawn-time entitlement mismatch*. CC 2.1.173/2.1.174 are Fable-5-suffix and background-env-inheritance fixes — orthogonal.

Because the Team-mode path (#36670) is open and no CHANGELOG resolves the single-spawn root cause, the constraint is treated as still-active. A follow-up SPEC (conditional) MAY re-enable per-agent pinning only when #36670 is closed-with-fixed AND a CHANGELOG confirms Team `[1m]` inheritance for explicit `model:` teammates.

## Default-Model Cost Lever (CC 2.1.175)

[ZONE:Evolvable] [HARD] The `[1m]`-safe cost lever is the **Default-model** routed via `availableModels` + `enforceAvailableModels`, NOT per-agent `model:` pins. The deployed `settings.json` template (`.claude/settings.json.tmpl`) sets:

```json
"model": "sonnet",
"availableModels": ["sonnet", "opus", "haiku"],
"enforceAvailableModels": true
```

Semantics (CC 2.1.175 CHANGELOG verbatim): _"Added `enforceAvailableModels` managed setting — when enabled, the `availableModels` allowlist also constrains the Default model (a Default that would resolve to a disallowed model now falls back to the first allowed model)"_. CC 2.1.176 further tightens enforcement: alias model picks can no longer be redirected to a blocked model via `ANTHROPIC_DEFAULT_*_MODEL` env vars.

Why this is `[1m]`-safe: the lever operates on the **Default** model resolution at the settings level, not on per-agent explicit pins, so it does not trigger the spawn-time entitlement-inheritance failure (#45847/#51060/#36670). The cost-routing thesis (route the busy-agent cost through Sonnet, not Opus) flows through the Default; deep-reasoning exceptions use per-spawn `Agent(model: "opus")` only for the 5-10% of tasks where Opus wins (architecture, complex perf) — and even those inherit the parent `[1m]` entitlement because they are spawned without a frontmatter `model:` pin (the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug).

### GLM-mode reconciliation

[ZONE:Evolvable] [HARD] The `enforceAvailableModels: true` cost lever above interacts with GLM mode. When GLM mode is active (`moai glm` whole-session, or the GLM teammate panes of `moai cg`), the GLM activation sets `ANTHROPIC_DEFAULT_OPUS_MODEL` to the configured GLM high model (default `glm-5.2`), surfaced in the model UI as the Opus-slot alias. CC 2.1.176 redirect-blocking semantics mean that an `ANTHROPIC_DEFAULT_*_MODEL` redirect to a model NOT in `availableModels` is blocked, and the active model silently falls back to the first allowed model (Sonnet). Before reconciliation, the allowlist was `["sonnet", "opus", "haiku"]` — which does NOT contain the Opus-slot `[1m]` alias variant, so every GLM session fell back to Sonnet and the GLM cost-optimization purpose was defeated.

The reconciliation EXPANDS the allowlist in two stages. Stage 1 added the `[1m]` canonical alias variants (`opus[1m]`, `sonnet[1m]`); this admitted the Claude 1M-context aliases but was INSUFFICIENT for GLM mode, because Claude Code (v2.1.186) forwards the resolved swap target — the raw GLM model id (e.g. `glm-5.2`) — and checks THAT against `availableModels`. The raw GLM ids were still absent, so the swap was still declined (`availableModels allowlist; declining the swap`) and the session fell back to Sonnet. Stage 2 therefore adds the raw GLM model ids directly:

```json
"availableModels": ["sonnet", "opus", "haiku", "opus[1m]", "sonnet[1m]", "glm-5.2", "glm-5.1", "glm-4.7", "glm-4.6", "glm-4.5", "glm-4.5-air"]
```

The expansion is **allowlist-only**: the Default model stays `sonnet` and `enforceAvailableModels` stays `true`, both byte-unchanged. A non-GLM (Claude `moai cc` / plain Claude) session still resolves its Default to Sonnet and still has enforcement active — the only behavioral change is that the GLM swap targets (`glm-5.2` and the other GLM tiers) are now admitted instead of declined.

The raw GLM model ids are added directly to the settings `availableModels` allowlist. They are NOT added to `modelCanonical` (`internal/web/validate.go`: `["opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"]`) — that set governs only the web/profile model picker and does NOT cross-check the settings `availableModels` array, so the GLM ids live in the settings allowlist without surfacing in the web picker.

Scope note: this reconciliation is a **static template allowlist expansion** in `.claude/settings.json.tmpl`. It touches no Go runtime code (`glm.go` / `launcher.go` / `settings.go` unchanged) and writes nothing to `settings.local.json` — so the solo `moai glm` "settings.local.json clean" design (no GLM env leak to subsequent plain-`claude` invocations) is preserved.

## Model Policy Tiers

Model policy is set via `moai init --model-policy <tier>`. The tier columns reference **role profiles** (the `workflow.yaml` role_profile / domain-whitelist taxonomy), not static agent files. Under the 8-agent catalog consolidation, the retained MoAI-custom agents (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`) all default to `model: inherit`; the tier governs which role_profile / domain scope is routed to Opus vs Sonnet vs Haiku when the orchestrator spawns `Agent(general-purpose)`.

| Tier | Description | Opus (deep reasoning) | Sonnet (implementation) | Haiku (mechanical / read-only) |
|------|-------------|------------------------|--------------------------|--------------------------------|
| high | Maximum quality | spec-planning, architecture, security-review | backend, frontend, ddd, tdd implementation | docs, git, read-only-investigation |
| medium | Balanced (default) | spec-planning, architecture, security-review | backend, frontend, ddd, tdd implementation | docs, git, read-only-investigation |
| low | Cost optimized | None | spec-planning, architecture, security-review | All other role profiles |

## CG Mode

CG Mode (Claude + GLM) uses environment variable overrides, not model field changes:
- Leader session: Uses Claude models (no GLM env)
- Teammate sessions: Inherit GLM env from tmux session
- Activation: `moai cg` (requires tmux)

## Effort Levels

Claude models support effort levels that control reasoning depth (Opus 4.8 calibration):
- xhigh: best setting for coding and agentic use cases
- high: default on Opus 4.8 across all surfaces; minimum for intelligence-sensitive work
- medium: cost-sensitive work that can trade off intelligence
- low: short, scoped, latency-sensitive tasks

Note: `ultrathink` is a Claude Code one-turn keyword that requests deeper reasoning for that prompt; MoAI standardizes it to `effort: xhigh` (the coding/agentic level above) for that turn.

## Rules

- Agent `model` field must be one of: inherit, opus, sonnet, haiku
- [ZONE:Evolvable] [HARD] New agent definitions SHOULD use `model: inherit` (default); explicit `sonnet`/`opus` are deprecated due to Claude Code Issue #45847/#51060 (see Inherit-by-Default Convention)
- `model: haiku` remains valid for speed-critical agents (immune to the [1m] entitlement bug)
- GLM is configured via env vars in settings.json, never via model field
- Model policy tier is a CLI concern, not an agent definition concern
- CG Mode uses tmux session-level env isolation for model routing
- Old model versions are auto-migrated: do not pin to specific version IDs
