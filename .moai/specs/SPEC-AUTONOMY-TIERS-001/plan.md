# plan.md — SPEC-AUTONOMY-TIERS-001

> Implementation plan. Order is decision-reversibility-first: the tier-persistence surface (OQ-3) leads because it determines where the selector writes and what the renderer reads; the renderer + selector follow; the kill-switch + sandbox-proof gate (OQ-1/OQ-4) close out because they are the most reversible (additive gating on top of a working renderer).

## §A. Context

This SPEC is a **redesign codification on top of a landed token**, not greenfield. The design authority is `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.1 (autonomy 3-mode — selector + USER/PROJECT split + kill-switch) + §6 risk row 2 + §6 open question 1 (G7). Per the report's critical-path diagram, AUTONOMY-TIERS is a P1 that proceeds AFTER `SPEC-STOPCHAIN-TRIM-001` completes (status: completed — the `MOAI_AUTONOMY_TIER` token + Go reader + env-key constant + 3-value enum are all landed). This SPEC consumes those surfaces and delivers the user-facing selection + behavior wiring.

The report's §3.1 framing: "native primitives as a BUNDLE (no new concept creation). moai init (CLI interactive) + moai web (console toggle) select → defaultMode to USER settings (auto/bypass are project/local-ignored, v2.1.142+), deny/ask rule set to PROJECT settings (merged across all sessions)."

The mode-token → knob table (spec.md §C) is the renderer's contract. Each tier is a coherent combination of EXISTING knobs — no new mechanism. The net-new surfaces are: (1) the selector UI in two surfaces, (2) the tier→bundle renderer calling the existing toolpolicy codegen, (3) the kill-switch gating, (4) the sandbox-proof gating on fully-autonomous.

## §B. Known Issues

- **K-1** (OQ-3): the tier-persistence surface is unresolved. STOPCHAIN-TRIM resolved the hooks' READ path to the env-key (`$MOAI_AUTONOMY_TIER` without the binary). But the SELECTOR needs a WRITABLE surface that survives across sessions. Candidates: (a) env-only — selector writes to `settings.local.json` `env` block, (b) new `workflow.yaml` key that the launcher exports to env at session start, (c) both. K-1 is the single highest-reversibility decision because it determines where M2 (selector) writes and what M3 (renderer) reads. **Recommended**: (b) `workflow.yaml` key as the persisted user selection, launcher exports to env-key at session start (shell hooks keep reading the env-key unchanged).
- **K-2** (OQ-1 / G7 — load-bearing): the fully-autonomous sandbox-proof method. The report flags G7 RED. CLI flag `--fully-autonomous` + sandbox/container attestation is the recommended option; `moai web` exposure is the rejected option (web without sandbox proof is unsafe). K-2 blocks REQ-002 + REQ-005's fully-autonomous path.
- **K-3** (OQ-4): the concrete attestation mechanism for "sandbox proof" (env marker, cgroup, CLI flag). K-3 is downstream of K-2's decision but blocks the implementation of whatever K-2 picks.
- **K-4** (OQ-2): whether the report's "classifier safety net" and "stop-goal type:agent promotion" at `automatic` imply a deny/ask delta or only a hook-side delta (already owned by STOPCHAIN-TRIM). REQ-004 asserts deny/ask identical; K-4 is the plan-auditor check that no deny/ask delta is hiding in the report's narrative.
- **K-5**: USER-scope vs PROJECT-scope settings paths. The renderer (REQ-003) writes `defaultMode` to USER scope and deny/ask to PROJECT scope. The existing `internal/config/resolver.go` / `source.go` carry the scope distinction, but the renderer MUST resolve the two paths correctly per platform (USER = `~/.claude/settings.json` or the active profile's settings; PROJECT = `<project>/.claude/settings.json`). A path-resolution bug here would write `bypassPermissions` to the PROJECT file (where Claude Code ignores it) and silently fail to grant the tier's behavior.
- **K-6**: the `disableBypassPermissionsMode` kill-switch's authoritative source. Claude Code recognizes it as a settings key; MoAI needs to read it from the managed/enterprise config layer (which may be an env-provided config path, not necessarily a project file). The read path MUST be confirmed before REQ-005's gating is implementable.
- **K-7**: template neutrality (CLAUDE.local.md §15 / §25). The distributed template (`internal/template/templates/**`) MUST NOT ship `fully-autonomous` as a default or pre-selection. The `moai init` default page in the template pre-selects `semi-auto`; the `--autonomy-tier` flag is the opt-in surface. The CI neutrality guard (`template-neutrality-check.yaml`) treats internal SPEC IDs / REQ tokens as forbidden content — the template's selector copy MUST use generic prose, not `REQ-AUTONOMY-TIERS-NNN` references.

## §C. Pre-flight (read-only reconnaissance — before M1)

1. Read `internal/cli/init.go` (paginated wizard + `--profile` / `--harness-profile` flag patterns at L81-112) — confirm where the autonomy-tier wizard page slots in and how the non-interactive `--autonomy-tier` flag mirrors `--profile`.
2. Read `internal/cli/web.go` + `internal/cli/web_port.go` — confirm the console toggle surface and whether a tier toggle is additive or requires a console architecture change.
3. Read `internal/config/toolpolicy/codegen.go` (`BuildInto` / `BuildIntoAuto` + `defaultModeOverride` at L204-270) + `internal/cli/tool_policy.go:67` (existing `--default-mode` flag) — confirm the renderer is a new CALLER of the existing codegen, not a new writer.
4. Read `internal/config/resolver.go` + `internal/config/source.go` — resolve USER-scope vs PROJECT-scope settings paths (K-5).
5. Read `internal/config/autonomy.go` + `internal/config/envkeys.go:99-108` + `internal/config/defaults.go:94-109` — confirm the token foundation the renderer consumes (do NOT modify — out of scope).
6. Grep `disableBypassPermissionsMode` across `internal/` + Claude Code settings docs — resolve K-6 (authoritative source).
7. Read `.moai/config/sections/workflow.yaml` (or the nearest equivalent) — resolve OQ-3 (does a `workflow.yaml` tier key already exist, or is it net-new?).
8. Read `internal/cli/launcher.go` (`buildEnvForLaunch` / `injectTmuxSessionEnv`) — confirm the launcher env-export path if OQ-3 picks option (b).
9. Read `.moai/specs/SPEC-STOPCHAIN-TRIM-001/spec.md` (the prerequisite) — confirm the hook-side tier behavior is downstream consumption, NOT re-scoped here.

## §D. Constraints (recap from spec.md §D — binding on the plan)

1. deny/ask rules tier-invariant (REQ-004). No tier weakens a deny.
2. Backward compat: unset / `semi-auto` = today's behavior, byte-identical renderer output (REQ-007).
3. Token foundation is owned by STOPCHAIN-TRIM — this SPEC consumes, does NOT redefine.
4. `fully-autonomous` is opt-in / local-only — template ships `semi-auto` default (REQ-006).
5. Tier M (justified in spec.md §B — 4 net-new surfaces, each a small change; multi-surface but no deep refactor).
6. No new CLI mechanism — the `--autonomy-tier` flag mirrors `--profile`; the renderer calls existing codegen.
7. Template neutrality — the template's selector copy uses generic prose; no internal SPEC IDs / REQ tokens leak (CI neutrality guard).

## §E. Self-Verification (run-phase — what manager-develop must demonstrate)

- Unit test: the 3-value closed-set validation accepts `{semi-auto, automatic, fully-autonomous}` case-insensitively, whitespace-trimmed, rejects invalid values with an error (NOT silent fallback) for the `--autonomy-tier` flag.
- Unit test: the tier→bundle renderer produces the correct `defaultMode` per tier AND byte-identical deny/ask arrays across tiers (AC-003 + AC-004).
- Unit test: the renderer writes `defaultMode` to the USER-scope path and deny/ask to the PROJECT-scope path (AC-003 path-resolution check).
- Unit test: `disableBypassPermissionsMode` engaged → `--autonomy-tier fully-autonomous` rejected, web toggle disabled, renderer refuses `bypassPermissions`, existing bypass session downgrades to `auto` + advisory (AC-005).
- Unit test: unset / `semi-auto` selection → renderer output byte-identical to today's template output (AC-007 — backward-compat regression guard).
- Integration test: an end-to-end `moai init --non-interactive --autonomy-tier <value>` produces the expected settings blocks in both USER and PROJECT scope.
- Regression test: the deny/ask arrays at `fully-autonomous` still catch main push / secrets / `rm -rf` / deploy (AC-004 cross-check against the destructive-pattern denylist).
- LSP / lint / build / test clean.

## §F. Milestones

### Milestone M1 — Tier-persistence surface (OQ-3 resolution)

Highest reversibility: the persistence surface determines where M2's selector writes and what M3's renderer reads. K-1 hazard.

**Files (expected):**
- A persistence key — either a new `workflow.yaml` key (e.g. `autonomy.tier`) read by the launcher and exported to `MOAI_AUTONOMY_TIER` at session start, OR an env-only write to `settings.local.json`'s `env` block. (OQ-3 decision.)
- `internal/config/` — a reader for the persisted tier selection (separate from `AutonomyTier()` which reads the env-key; the reader resolves persisted → effective tier).
- `internal/cli/launcher.go` (if option b) — export the persisted key to `MOAI_AUTONOMY_TIER` at session start.
- A unit test covering: persisted `automatic` → effective `automatic`; persisted unset → effective `semi-auto`; persisted `fully-autonomous` → effective `fully-autonomous` (the sandbox-proof gate is M5).

**Exit:** the persistence surface exists and round-trips; AC-007 partially exercisable from here (unset → semi-auto).

### Milestone M2 — Tier selector in `moai init` (interactive + non-interactive)

Second reversibility tier: the selector is the user-facing entry point. Depends on M1.

**Files (expected):**
- `internal/cli/init.go` — new wizard page for autonomy-tier selection (mirrors the `--profile` / `--harness-profile` pattern at L81-112); `--autonomy-tier <value>` non-interactive flag with closed-set validation (fail-loud, NOT the reader's fail-safe).
- The wizard page persists the selection via M1's surface.
- A unit test covering the wizard page selection + the `--autonomy-tier` flag validation (AC-001).

**Exit:** AC-001 green; selector persists via M1.

### Milestone M3 — Tier → permission-bundle renderer (USER vs PROJECT scope)

Third reversibility tier: the renderer is the load-bearing wiring. Depends on M1 (reads the persisted tier) + M2 (selector populates it). K-5 hazard.

**Files (expected):**
- A new renderer (likely under `internal/config/toolpolicy/` or a sibling) that calls the existing `BuildInto` / `BuildIntoAuto` codegen with the tier's `defaultMode` as the override — a NEW CALLER, not a new codegen.
- The renderer resolves USER-scope vs PROJECT-scope paths via `internal/config/resolver.go` / `source.go` (K-5) and writes `defaultMode` to USER, deny/ask to PROJECT.
- Unit tests: per-tier `defaultMode` correct; deny/ask byte-identical across tiers (AC-003 + AC-004); path-resolution correctness.

**Exit:** AC-003 + AC-004 green.

### Milestone M4 — Tier selector in `moai web` (console toggle)

Fourth reversibility tier: the web console toggle. Depends on M1 (persistence) + M3 (renderer). The toggle is additive UI on the existing console.

**Files (expected):**
- `internal/cli/web.go` + the console's toggle-control surface — a tier toggle offering 3 values.
- The toggle writes via M1's surface and re-invokes M3's renderer.
- The fully-autonomous option's enablement depends on M5 (sandbox proof) — initially ships DISABLED pending M5.

**Exit:** AC-002 green (minus the sandbox-proof gating, which lands in M5).

### Milestone M5 — Sandbox-proof gate + manager kill-switch (OQ-1 / OQ-4 + REQ-005)

Fifth reversibility tier: the gating is the most reversible (additive gating on top of a working renderer + selector). K-2 + K-3 + K-6 hazards.

**Files (expected):**
- The sandbox-proof attestation mechanism (OQ-4 decision — env marker `MOAI_SANDBOX_PROOF=<kind>` and/or a `--sandbox-proof` CLI flag).
- `internal/cli/web.go` — enable the fully-autonomous toggle ONLY when sandbox proof is present.
- The `disableBypassPermissionsMode` kill-switch read path (K-6 — managed/enterprise config layer).
- Gating in the selector (M2), the web toggle (M4), and the renderer (M3): fully-autonomous rejected when kill-switch engaged.
- Downgrade path: existing `bypassPermissions` session + kill-switch engaged → effective `auto` + advisory.
- Unit tests: AC-002 (fully-autonomous disabled without proof / under kill-switch) + AC-005 (kill-switch rejects + downgrades).

**Exit:** AC-002 + AC-005 green. G7 (OQ-1) implemented per the Implementation-Kickoff decision; the RED flag downgrades to green ONLY after the sandbox-proof gate passes AC-002 + AC-005 unit tests.

### Milestone M6 — Template neutrality pass + backward-compat regression guard

Final tier: the template defaults + the zero-delta regression guard. K-7 hazard.

**Files (expected):**
- `internal/template/templates/.claude/settings.json` (and/or the wizard default in the template's init scaffolding) — pre-select `semi-auto`; ship NO `fully-autonomous` default (REQ-006).
- Template neutrality check: the selector copy uses generic prose; no `REQ-AUTONOMY-TIERS-NNN` or internal SPEC ID leaks into the template (CI guard `template-neutrality-check.yaml`).
- Backward-compat regression test: unset / `semi-auto` → renderer output byte-identical to today's template (AC-007).

**Exit:** AC-006 + AC-007 green; CI neutrality guard green.

## §G. Anti-Patterns

- **AP-1 — Re-scoping the token.** The `MOAI_AUTONOMY_TIER` env-key, the Go reader, the enum, and the hook-side tier behavior are ALL `SPEC-STOPCHAIN-TRIM-001` property. This SPEC CONSUMES them; any edit to `internal/config/autonomy.go` / `envkeys.go` / `defaults.go` / the mode-aware hooks is OUT OF SCOPE and must be routed to STOPCHAIN-TRIM (or a follow-up).
- **AP-2 — Inventing a new permission codegen.** The tier→bundle renderer is a NEW CALLER of `internal/config/toolpolicy.BuildInto` / `BuildIntoAuto` with a `defaultModeOverride`. A parallel settings writer violates the toolpolicy SSOT (SPEC-V3R6-TOOL-POLICY-SSOT-001) and creates a drift surface.
- **AP-3 — Weakening deny/ask at higher tiers.** The §3.1 invariant is load-bearing. Any "convenience" that downgrades a deny or an ask at `automatic` / `fully-autonomous` violates REQ-004 and the security rationale (§6 risk row 2). The deny/ask arrays are IDENTICAL across tiers.
- **AP-4 — Shipping fully-autonomous as a template default.** The template ships `semi-auto`. Pre-picking `fully-autonomous` (or even `automatic`) in the distributed template violates template neutrality (CLAUDE.local.md §15 / §25) and the epic constraint that fully-autonomous is opt-in / local-only.
- **AP-5 — Silent fail-safe at the SELECTOR surface.** The `config.AutonomyTier()` reader fail-safes to `semi-auto` on invalid input (correct for a reader). The `--autonomy-tier` flag MUST fail-LOUD on invalid input (error, not fallback) — the selector is the user-facing surface; a silent fallback would hide a typo from the user.
- **AP-6 — Writing `defaultMode` to PROJECT scope.** Claude Code v2.1.142+ makes `auto` / `bypassPermissions` USER-scope-only (PROJECT/local cannot grant them). Writing `defaultMode: bypassPermissions` to the PROJECT file silently fails to grant the tier's behavior. The renderer MUST write `defaultMode` to USER scope (K-5).

## §H. Cross-References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.1, §6 risk row 2, §6 open question 1 (G7).
- Prerequisite (token foundation): `.moai/specs/SPEC-STOPCHAIN-TRIM-001/spec.md` (status: completed).
- Sibling format reference: `.moai/specs/SPEC-INFINITE-GOAL-001/spec.md` (Tier-M format, recent).
- Toolpolicy SSOT: `.moai/specs/SPEC-V3R6-TOOL-POLICY-SSOT-001/` (the codegen the renderer calls).
- Template neutrality: CLAUDE.local.md §15 + §25 + `.moai/docs/template-internal-isolation-doctrine.md`.
- Hook-side tier behavior (downstream): `.claude/hooks/moai/sync-phase-quality-gate.sh`, `internal/hook/pre_tool.go` IsGitCommit branch, `.claude/hooks/moai/handle-stop-goal.sh`.
- Claude Code settings scope: `code.claude.com/docs/en/settings` (USER vs PROJECT vs local precedence, `disableBypassPermissionsMode` kill-switch).
