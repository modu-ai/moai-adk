# SPEC-GLM-EFFORT-MAX-001 — Implementation Plan (Tier S)

## §A Context

- Card t175. Worktree `.claude/worktrees/t175`, branch `WT-glm-effort-max`. Plan-phase artifacts only — no git operations in this delegation.
- SPEC: `.moai/specs/SPEC-GLM-EFFORT-MAX-001/spec.md` (6 REQ, 8 inline AC, Tier S ceilings respected).
- Ground truth: `.moai/reports/t175/measurements.md` (committed) + direct code reads listed in spec.md §1.1.
- Affected files (run-phase scope, 6 total — under the Tier S 5-file guidance by one; the 4th and 5th are test-only flips in packages already touched):
  1. `internal/template/glm_effort_overlay.go` — collapse branch (`:121-122` medium/high→`glmReasoningMax`), session default (`:197-199` → `glmReasoningMax`), dead-var removal (`:95`), comments (`:99-116` collapse doc, `:182-199` cost policy, `:194-196` UNVERIFIED).
  2. `internal/template/glm_effort_overlay_test.go` — RED flips + AC-GET-003 re-anchor.
  3. `internal/cli/glm_reasoning_overlay_test.go` — session-default env flip (`:23-25`, literal `"high"` → `template.GLMReasoningEffortMax`).
  4. `internal/cli/glm_test.go` — `TestGLMReasoningEnvVarsForEffort` rows `:511-512,:515`.
  5. `internal/web/agentfm_glm_reasoning_test.go` — manager-spec chip `:98` + comment `:77-78`.
  6. `internal/template/templates/.moai/config/sections/llm.yaml` — collapse doc block `:16-23` (scope rider §D-2) + `make build`.
- NOT touched: `internal/cli/glm.go` (both writers inherit from the overlay; only their UNVERIFIED comments change — riding REQ-GEM-005, comment-only), `internal/web/glm_tier_test.go`, `internal/settings/schema_sections.go`, any local `llm.yaml`.
- Route: repo-local Route B (PR) for all tiers — no direct `main` push (`.claude/rules/local/repo-local-pr-policy.md`).
- Cycle: TDD (`quality.yaml` `constitution.development_mode: tdd`) — RED (flip tests) → GREEN (overlay change) → doc surfaces → verify.

## §B Known Issues

- **B-1 Cross-SPEC reversal (highest-change-likelihood decision).** REQ-GEM-002 reverses SPEC-GLM-EFFORT-REBALANCE-001 REQ-GER-004 (in-progress, v3.1.0). REBALANCE's step-down already landed in the tree; our change re-lifts it on the operator's newer order. If the lead does NOT ratify, REQ-GEM-002/003 shrink to collapse-only and `glmReasoningHigh` survives via the session default — M1's RED set must then drop the two session-default assertions. Resolve FIRST (§D-1).
- **B-2 AC-GET-003 test re-anchor.** At Claude effort `high`, `ResolveGLMReasoning("builder-harness", …)` and the coding-max override now both yield `max` — the old discrimination ("not overridden") is unobservable at that input. Re-anchor the make-or-break assertion to `EffortLevelLow`: `builder-harness@low → low` (collapse) vs `manager-develop@low → max` (override). This narrows SPEC-GLM-EFFORT-TUNE-001 AC-GET-003's *witness*, not its guarantee (builder-harness is still not in the override set — `GLMCodingMaxOverrideAgents()` membership assertions are untouched).
- **B-3 Pre-existing template doc staleness.** `templates/.moai/config/sections/llm.yaml:16-23` still says "effort low → thinking off" — stale since the glm-5.3 floor rework (code moved to `low`, thinking enabled). REQ-GEM-005 fixes that line in the same hunk as the medium/high line. Mirror-parity discipline: edit template source, `make build`, mirror checks in the same commit.
- **B-4 Magic literal.** `glm_reasoning_overlay_test.go:23` asserts the raw string `"high"` (a §14 smell); the flip uses `template.GLMReasoningEffortMax`, consistent with `glm_test.go`'s constant usage.
- **B-5 Worktree-invisible config.** The local `llm.yaml` (primary checkout, gitignored) cannot be read from this worktree — all claims about its `glm.effort` block are attributed to the measurement record, not to direct observation. `llm.profiles` shadowing (REBALANCE Amendment 1) does not apply: the collapse and the hardcoded session default are not profile-derived (spec.md C-3).
- **B-6 Session env is frozen pre-change.** This live session already carries `ANTHROPIC_REASONING_EFFORT=high`; post-change verification is binary-output + fresh-session (AC-GEM-008), not this session's env.

## §C Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-glm-effort-max @ base
go test -run 'TestCollapseClaudeEffortToGLM|TestSessionGLMReasoningState|TestResolveGLMReasoning|TestGLMReasoningEnvVars' ./internal/template/ ./internal/cli/   # baseline GREEN (current behavior)
go build ./... && GOOS=windows GOARCH=amd64 go build ./...  # baseline build clean
golangci-lint run --timeout=2m ./internal/template/... ./internal/cli/... ./internal/web/... 2>&1 | tail -5  # lint baseline (NEW-vs-baseline classification)
```

## §D Constraints & Open Decisions (resolve before Implementation Kickoff Approval)

### §D-1 [NEEDS CLARIFICATION: SessionGLMReasoningState disposition — ratify max + the REQ-GER-004 reversal]

**Decision (this SPEC): raise the session default to `max` — candidate (a) of the measurements §5.**

Justification (one paragraph, for lead ratification): The operator's order is recorded as "medium/high→max, low→low — effectively everything except low is max," and under the Branch-B delivery the session-global env var is the ONLY reasoning channel every spawn pays — so if the session default stayed `high` while the collapse maps high→max, the operator's order would reach only the main-session prefs-driven path and silently NOT the default session, which is precisely where sub-agent reasoning depth is set; the original `high`-floor rationale (:186-192, "paid by every spawn") is weakened by the t127 measurement showing trivial spawns cost ~0 subagent tokens — the cost lives in large calls' reasoning tokens, which is exactly where the operator wants `max` — and `max` is additionally z.ai's own omit-default, so the session default stops fighting the backend's native default. Cost: sub-agent and empty-effort spawns reason at `max`; on trivial work the model's demand-driven depth (P1 vs P3) keeps the increment small, and the residual unquantified increment on large calls is accepted by the operator's directive. This decision reverses SPEC-GLM-EFFORT-REBALANCE-001 REQ-GER-004 (see B-1) — ratifying it means the lead also chooses REBALANCE's disposition (out of scope here).

### §D-2 [NEEDS CLARIFICATION: template-mirror llm.yaml doc-block scope]

The delegation excluded "llm.yaml (the lead owns it)" — read as the LOCAL config instance. The TEMPLATE mirror (`internal/template/templates/.moai/config/sections/llm.yaml:16-23`) is committed, distributed source whose doc block states the exact mapping REQ-GEM-001 changes; leaving it stale ships wrong docs to every user, and the adjacent "thinking off" line is already stale (B-3). **Proposal: in scope, one hunk, `make build` + mirror checks in the same commit.** If the lead reads the exclusion as covering the mirror too, drop REQ-GEM-005's mirror clause + M3's yaml edit (AC-GEM-07 narrows accordingly) and record the mirror staleness as a follow-up card.

### §D-3 Binding constraints

- PRESERVE (no edit): `internal/web/glm_tier_test.go`, `internal/settings/schema_sections.go`, `internal/template/profile_matrix.go` (comment at `:219-227` remains accurate — grouping statement unchanged), `internal/cli/glm.go` code (comments only per REQ-GEM-005), `buildTmuxClearVars`, any local `llm.yaml`.
- No magic literals (§14): use `glmReasoningMax` / `GLMReasoningEffortMax`, never a fresh `"max"` string in code.
- Affected-packages-only local tests (C-7); full-matrix verdict is CI's.
- Conventional commits `feat(SPEC-GLM-EFFORT-MAX-001): M<N> …`; Route B PR (C-6); never `--no-verify`.

## §E Self-Verification (run-phase §E evidence, attributable per VCI §2)

- **E1** AC matrix (AC-GEM-001..008) — command + verbatim output per row; AC-GEM-008 attributed to lead/operator with the observed command.
- **E2** `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` — exit 0 both.
- **E3** Coverage: `go test -cover ./internal/template/...` (affected; 85% target).
- **E4** Boundary grep n/a (no hook/harness code touched) — record the n/a with reason, do not omit silently.
- **E5** `golangci-lint run` on the three affected packages — NEW-vs-baseline split; must include the no-unused-`glmReasoningHigh` observation (AC-GEM-003).
- **E6** Commit list + push/PR state (Route B; per mission, git stays with the lead unless re-delegated).
- **E7** Blocker reports (none anticipated beyond §D riders).
- **E8** RED evidence: verbatim failing output of the flipped tests BEFORE the GREEN edit (TDD falsifiability).

## §F Milestones (priority order, no time estimates)

- **M1 — RED (decision-heavy first).** Flip the four runtime test files (template overlay test incl. session-default + AC-GET-003 low-anchor; cli session-default env; cli ForEffort rows; web agentfm chip). Capture verbatim RED output (E8). *Encodes the §D-1 decision — run only after ratification.*
- **M2 — GREEN.** `glm_effort_overlay.go`: collapse branch medium/high→`glmReasoningMax`; session default→`glmReasoningMax`; delete `glmReasoningHigh`; rewrite cost-policy + UNVERIFIED comments; glm.go UNVERIFIED comments. Affected-package tests green.
- **M3 — Doc surface.** Template-mirror doc block (per §D-2 rider) + `make build` + mirror checks, same commit.
- **M4 — Verify & evidence.** E1-E8 sweep, cross-platform build, AC matrix, `progress.md` §E.2 population (incl. AC-GEM-008 fresh-session check handoff to lead).

## §G Anti-Patterns

- Do NOT "fix" `glm_tier_test.go` or `glmDefaultTierEffort` — stored-only surface; their `high` defaults are the recorded divergence, not a defect (spec.md §5).
- Do NOT delete `GLMStateHigh`/`GLMReasoningEffortHigh` — only the `glmReasoningHigh` *value* dies; the name/wire constants stay live (REQ-GEM-003).
- Do NOT run the local full suite; do NOT push to `main`.
- Do NOT widen scope to per-agent wire channels, profile cells, or REBALANCE's artifacts.
- Do NOT claim the fresh-session env check from inside this worktree session — it needs a rebuilt binary and a new session (B-6); hand it to the lead with the exact command.

## §H Cross-References

- spec.md §1.3 (REBALANCE conflict), §4 (C-1..C-7), §5 (exclusions)
- `.claude/rules/local/repo-local-pr-policy.md` — Route B mandate
- `manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form
- `verification-claim-integrity.md` §2-3 — E-item attribution + Gaps discipline
