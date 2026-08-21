---
id: SPEC-GLM-EFFORT-MAX-001
title: "GLM reasoning effort ceiling raise — collapse medium/high to max"
version: "0.1.1"
status: in-progress
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: "internal/template/glm_effort_overlay.go"
lifecycle: spec-anchored
tags: "glm, effort, reasoning-effort, overlay, collapse, session-default"
era: V3R6
tier: S
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001, SPEC-GLM-EFFORT-TUNE-001, SPEC-GLM-EFFORT-REBALANCE-001]
---

# SPEC-GLM-EFFORT-MAX-001 — GLM reasoning effort ceiling raise

## HISTORY

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-08-22 | manager-spec | Initial plan-phase draft (Tier S, card t175). Ground truth: `.moai/reports/t175/measurements.md` (committed measurement record) + direct code reads at worktree HEAD. |
| 0.1.1 | 2026-08-22 | manager-spec | Lead ratifications + audit fixes applied. D1: session default = `max` APPROVED; REQ-GER-004 (stalled draft SPEC-GLM-EFFORT-REBALANCE-001) supersession recorded (§1.3); template-mirror doc-block scope approved — both plan decision markers converted to `[RESOLVED]` headings. D2: plan file inventory reworded honestly (8 diff files). D3: profile_matrix.go comment added to REQ-GEM-005 scope. D4: AC-GEM-003 reworded (`GLMReasoningEffortHigh` deterministically unreferenced post-change). D5/D6: AC-GEM-007 rename in plan; schema_sections.go rationale-cross-reference staleness recorded (§5, not edited). |

## §1 Background — operator directive and code truth

The operator directed (card t175, order recorded in the measurements §1): raise the GLM reasoning effort so that Claude effort `medium`/`high` collapse to GLM **`max`**; `low` stays `low`. Effectively everything except `low` becomes `max`.

### §1.1 Code truth (read directly at worktree HEAD, not from summaries)

| Surface | Current behavior | Anchor |
|---|---|---|
| `CollapseClaudeEffortToGLM` | low→`low`, **medium/high→`high`**, xhigh/max→`max`, unrecognized→`max` (totality clause) | `internal/template/glm_effort_overlay.go:117-129` |
| `SessionGLMReasoningState` | hardcoded `high` — session default for sub-agents + empty-effort fallback | `internal/template/glm_effort_overlay.go:197-199` |
| Cost-policy comment | floor `high` because "a session-global value is paid by every spawn in the session" | `glm_effort_overlay.go:182-192` |
| Shim-consumption comment | marked UNVERIFIED | `glm_effort_overlay.go:194-196`, `internal/cli/glm.go:386-391,414-417` |
| Env write points | exactly two writers, both derive from the overlay (`glmReasoningEnvVars` → `SessionGLMReasoningState`; `glmReasoningEnvVarsForEffort` → `SessionGLMReasoningStateForEffort`) — no site carries its own value | `internal/cli/glm.go:392-399,418+` |
| Env clear point | key-parity only, value-agnostic | `internal/cli/glm.go:572-598` |

### §1.2 Measured ground truth (`.moai/reports/t175/measurements.md`)

- **AC-MTP-032b residual is now MEASURED**: the z.ai Anthropic-compat shim honors the Anthropic `thinking` parameter (P1/P3: thinking blocks returned, depth scales with budget) and IGNORES a top-level `reasoning_effort` field (P2: silently dropped). The live session env carries `ANTHROPIC_REASONING_EFFORT=high`, matching the code hardcode (injection verified); the session's own thinking-block responses are indirect end-to-end evidence that the env→reasoning chain is live.
- **Cost material (t127)**: trivial no-tool spawns measured ~0 subagent tokens — per-spawn fixed reasoning overhead is not the cost driver; reasoning-token increment scales on large calls, which is where the operator wants `max` depth.
- **llm.yaml `glm.effort` block is stored-only** — no code reads it at runtime (`internal/settings/schema_sections.go:192-204` documents the fields as store-only; the single runtime channel is the session-global `ANTHROPIC_REASONING_EFFORT`). Its `reasoning-max` label's channel assumption (a z.ai-native `reasoning_effort` field) is falsified by probe P2. Recorded as a documented config-code divergence; not modified here.

### §1.3 Cross-SPEC supersession (lead-ruled, 2026-08-22)

**SPEC-GLM-EFFORT-REBALANCE-001** (carries catalog status `in-progress`; per lead arbitration it is a **stalled draft** — v0.1.0, its v3.1.0 target already shipped without it, M1 commit 763582247 unpushed, inactive since 2026-08-15) requires in REQ-GER-004 that `SessionGLMReasoningState()` return `reasoning-high`. That requirement is a restatement of the already-landed floor behavior, not living doctrine. **Lead ruling (2026-08-22): REQ-GER-004 (SPEC-GLM-EFFORT-REBALANCE-001, stalled draft) is superseded by this SPEC** — REQ-GEM-002 raises the session default to `max`, and the run phase rewrites the cost-policy comment (`:186-192`) on the ratified grounds: the session-global env is the only reasoning channel under Branch-B; t127 measured trivial-spawn cost ≈ 0; `max` is z.ai's own omit-default. REBALANCE's retirement stays Out of Scope (§5) — its other REQs are unrelated, and the stalled draft's disposition goes to a separate lead query at batch time.

## §2 Requirements (GEARS notation)

**REQ-GEM-001** (Ubiquitous): The GLM effort collapse (`CollapseClaudeEffortToGLM`, `internal/template/glm_effort_overlay.go`) shall map Claude effort `medium` and `high` onto the GLM `max` reasoning state (`glmReasoningMax`), shall map `low` onto `low` and `xhigh`/`max` onto `max`, and shall map every unrecognized effort onto `max` (totality clause unchanged).

**REQ-GEM-002** (State-driven): **While** a GLM-backed session derives its session-global reasoning state with no effort preference set, the deriver `SessionGLMReasoningState()` shall return the `max` reasoning state (thinking enabled, `reasoning_effort: max`) — superseding the `high` floor; per the lead ruling of 2026-08-22 (plan.md §D-1, resolved), REQ-GER-004 (SPEC-GLM-EFFORT-REBALANCE-001, stalled draft) is superseded by this requirement (§1.3).

**REQ-GEM-003** (Capability gate): **Where** the `glmReasoningHigh` package variable becomes unreachable from both the collapse and the session default, the implementation shall remove that variable, and shall retain the `GLMStateHigh` and `GLMReasoningEffortHigh` constants — `GLMStateHigh` remains referenced (`GLMReasoningStateNames()` at `:136`, the settings-widget domain; `internal/settings/schema_sections.go:184`, stored-only defaults), while `GLMReasoningEffortHigh` remains as the declared wire-vocabulary token (`high` is a legal z.ai wire value and a selectable stored state), deterministically unreferenced in-repo post-change by design (every current reference dies at the removed `glmReasoningHigh` constructor `:95` or flips to `GLMReasoningEffortMax`).

**REQ-GEM-004** (Event-driven): **When** the collapse mapping changes under this SPEC, every runtime-asserting test surface that asserts `medium`/`high` → the `high` state shall be updated to assert `max` — `internal/template/glm_effort_overlay_test.go` (collapse table `:53-54`; `ResolveGLMReasoning` rows `:96,:100`; `SessionGLMReasoningState` `:150-158`; `SessionGLMReasoningStateForEffort` `:174-175,:178`; the AC-GET-003 make-or-break assertion `:141-143` re-anchored to the `low`-effort discrimination, since at effort `high` the collapse and the coding-max override now agree on `max`), `internal/cli/glm_reasoning_overlay_test.go` (`:17-27`, session-default env value), `internal/cli/glm_test.go` (`TestGLMReasoningEnvVarsForEffort` `:511-515`), and `internal/web/agentfm_glm_reasoning_test.go` (`:98`, manager-spec chip) — while the stored-only surfaces (`internal/web/glm_tier_test.go` AC-WCR-031 defaults, `internal/settings/schema_sections.go` `glmDefaultTierEffort`) shall remain unchanged.

**REQ-GEM-005** (Ubiquitous): The cost-policy comment on `SessionGLMReasoningState` (`:182-199`) shall be rewritten on the lead-ratified grounds — the session-global env is the only reasoning channel under Branch-B; t127 measured trivial-spawn cost ≈ 0 subagent tokens; `max` is z.ai's own omit-default; the UNVERIFIED shim-consumption comments (`glm_effort_overlay.go:194-196`, `internal/cli/glm.go:386-391,:414-417`) shall be updated to cite the t175 measured finding (shim honors the `thinking` parameter; top-level `reasoning_effort` ignored); the delivery-granularity comment in `internal/template/profile_matrix.go` (`:219-227`) shall state the new grouping ({medium, high, xhigh, max} converge on `max`; `low` stays `low`) in place of the dead "one state and … another" phrasing; and the collapse-documentation block in the template mirror `internal/template/templates/.moai/config/sections/llm.yaml` (`:16-23`) shall state the new mapping, also correcting its pre-existing stale "thinking off" floor line (lead-ratified 2026-08-22; plan.md §D-2, resolved).

**REQ-GEM-006** (Ubiquitous): This SPEC shall cite `.moai/reports/t175/measurements.md` as the closure evidence for the AC-MTP-032b wire-effectiveness residual of SPEC-MODEL-TIER-PLANTYPE-001, and its run-phase verification shall demonstrate that the binary's env-write path emits `max` — unit-level via `glmReasoningEnvVars()`, session-level via a fresh GLM session observing `ANTHROPIC_REASONING_EFFORT=max` (lead/operator check, recorded in `progress.md` §E.2 with command + verbatim output per verification-claim-integrity §2).

## §3 Acceptance Criteria (inline, Given-When-Then)

**AC-GEM-001** — Given the modified collapse, when `go test -run 'TestCollapseClaudeEffortToGLM|TestSessionGLMReasoningStateForEffort' ./internal/template/` runs, then all subtests pass asserting {low→low, medium→max, high→max, xhigh→max, max→max, ""→max, "bogus"→max}, each with `ThinkingEnabled=true` and the matching `reasoning_effort` value.

**AC-GEM-002** — Given the modified session default, when `go test -run TestSessionGLMReasoningState ./internal/template/` runs, then it passes asserting `Name=max`, `ReasoningEffort=max`, `ThinkingEnabled=true`.

**AC-GEM-003** — Given the post-change tree, when `go build ./...` and `golangci-lint run ./internal/template/...` run, then both are clean — no unused-variable finding for `glmReasoningHigh` (removed); `GLMStateHigh` present and still referenced at `GLMReasoningStateNames()` (`:136`, widget domain) and `internal/settings/schema_sections.go:184` (stored-only defaults); `GLMReasoningEffortHigh` present as a declared constant (deterministically unreferenced in-repo post-change — every reference dies at the removed `glmReasoningHigh` constructor `:95` or flips to `GLMReasoningEffortMax`; exported, so not an unused finding).

**AC-GEM-004** — Given the env wire point, when `go test -run TestGLMReasoningEnvVars ./internal/cli/` runs, then the flipped test asserts `glmReasoningEnvVars()["ANTHROPIC_REASONING_EFFORT"] == "max"` (was `"high"`) and the inject↔clear parity test passes unchanged.

**AC-GEM-005** — Given the four flipped runtime test files, when `go test ./internal/template/... ./internal/cli/... ./internal/web/...` runs (affected packages only, per CLAUDE.local.md §4), then all pass.

**AC-GEM-006** — Given the stored-only surfaces, when the change lands, then `git diff --stat` (against the pre-change base) shows no entry for `internal/web/glm_tier_test.go` or `internal/settings/schema_sections.go`, and no local `llm.yaml` is modified in this worktree.

**AC-GEM-007** — Given the documentation surfaces, when grepped post-change, then no stale claim remains: the template mirror's collapse block (`internal/template/templates/.moai/config/sections/llm.yaml:16-23`) states medium/high→max and low→low (thinking enabled); the profile-matrix delivery comment (`internal/template/profile_matrix.go:219-227`) states the new convergence grouping ({medium, high, xhigh, max} → `max`; `low` → `low`); the cost-policy comment carries the ratified grounds (Branch-B sole channel; t127 trivial-spawn ≈ 0; z.ai omit-default); the shim comments cite the measured P1/P2/P3 finding instead of UNVERIFIED.

**AC-GEM-008** — Given a fresh GLM session launched from a binary built after this change (run-phase, lead/operator), when the session env is inspected, then `ANTHROPIC_REASONING_EFFORT=max` — recorded verbatim in `progress.md` §E.2 (command + observed output; residual risk noted: this live session's own env stays `high` because it was launched pre-change).

## §4 Constraints

- **C-1 code-only delivery change**: both env writers derive from the overlay (`internal/cli/glm.go:392-399,:418+`); no delivery-site value change is needed or allowed. The clear-list (`buildTmuxClearVars`) and `launcher.go:1189` filter are value-agnostic.
- **C-2 no Template-First mirror for the Go file**: `glm_effort_overlay.go` is Go source at `internal/template/` root, NOT under `internal/template/templates/` — the Template-First rule does not apply to it. The only template-surface touch is the doc block in `templates/.moai/config/sections/llm.yaml` (REQ-GEM-005; lead-ratified 2026-08-22, plan.md §D-2 resolved), which requires `make build` when applied.
- **C-3 config-shadow immunity**: `llm.profiles` shadowing (SPEC-GLM-EFFORT-REBALANCE-001 Amendment 1) affects profile *cells*, not the collapse mapping or the hardcoded session default — REQ-GEM-001/002 are shadow-immune.
- **C-4 llm.yaml divergence is recorded, not fixed**: the local `.moai/config/sections/llm.yaml` is gitignored primary-checkout runtime state (invisible from this worktree); its `glm.effort` stored block is lead-owned and Out of Scope.
- **C-5 z.ai semantics preserved**: `max` is z.ai's omit-default and its coding-task recommendation; `low` remains the wire-real floor under glm-5.3 (thinking cannot be disabled); the totality clause (never under-reason on unrecognized effort) is untouched.
- **C-6 repo-local Route B**: this repository requires PR-merge for ALL tiers (`.claude/rules/local/repo-local-pr-policy.md`); no direct push to `main`.
- **C-7 affected-packages-only local testing** (CLAUDE.local.md §4): no local full suite; CI owns the full-matrix verdict.

## §5 Out of Scope

### Out of Scope — llm.yaml glm.effort stored block and settings schema

- The stored-only `llm.glm.effort.*` tier fields (defaults asserted by `internal/web/glm_tier_test.go` AC-WCR-031; produced by `internal/settings/schema_sections.go` `glmDefaultTierEffort`) are NOT modified. They are stored-only (no runtime reader), the local file is lead-owned, and the config-code divergence is recorded in §1.2 rather than reconciled.
- Known post-change staleness, recorded not fixed: `internal/settings/schema_sections.go:175-180` grounds the stored tier-effort defaults on the same rationale as `SessionGLMReasoningState` (a rationale cross-reference). Once the overlay's grounds change (REQ-GEM-005), that cross-reference describes a superseded rationale. The file is a preserved stored-only surface (AC-GEM-006); the stale cross-reference is recorded here rather than edited.

### Out of Scope — SPEC-GLM-EFFORT-REBALANCE-001 retirement

- This SPEC records the REQ-level supersession of REQ-GER-004 (stalled draft; §1.3) but does NOT edit REBALANCE's frontmatter (`partially_superseded_by`), body, or progress, and does not retire the draft: its other REQs (plan/sync profile-row efforts) are unrelated to this change, and the stalled draft's disposition goes to a separate lead query at batch time.

### Out of Scope — per-spawn reasoning channel

- The delivery-granularity limitation (one session-global `ANTHROPIC_REASONING_EFFORT`, no per-agent wire value) stands. A per-spawn channel or a request-body `reasoning_effort` path is a separate design problem (unchanged from SPEC-GLM-EFFORT-TUNE-001 §E).

### Out of Scope — profile-matrix cells and agent efforts

- No `tierProfiles` cell, no `llm.profiles` entry, and no agent-frontmatter `effort:` value is changed. The Claude-side effort that feeds the collapse is settled input owned by other SPECs.

### Out of Scope — max-vs-high reasoning-token quantification

- Quantifying the reasoning-token increment of `max` vs `high` under high-load tasks is not attempted (measured gap, measurements §6). The P1/P3 depth-scales-with-budget structure plus the t127 trivial-spawn measurement are the decision material; further quantification is a follow-up if the lead requests it.

## §6 Cross-References

- `.moai/reports/t175/measurements.md` — committed measurement record (ground truth for §1.2)
- SPEC-MODEL-TIER-PLANTYPE-001 — parent overlay SPEC; AC-MTP-032b residual closed by REQ-GEM-006
- SPEC-GLM-EFFORT-TUNE-001 — override-set tune (completed); its AC-GET-003 assertion is re-anchored by REQ-GEM-004
- SPEC-GLM-EFFORT-REBALANCE-001 — in-progress sibling whose REQ-GER-004 this SPEC reverses (§1.3)
- `verification-claim-integrity.md` §2 — attribution contract binding AC-GEM-008's evidence record
- CLAUDE.local.md §2 (Template-First scope), §4 (affected-packages testing), §14 (no magic literals — the `"high"` literal in `glm_reasoning_overlay_test.go:23` flips to the `template` constant, not a new literal)
